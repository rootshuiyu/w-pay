package model

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel 每张表统一字段：创建时间、更新时间、逻辑删除
type BaseModel struct {
	ID        uint64         `gorm:"primaryKey" json:"id,string"`
	CreatedAt time.Time      `gorm:"index" json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// AdminRole 管理员角色
type AdminRole string

const (
	AdminRoleSuperAdmin AdminRole = "super_admin"
	AdminRoleFinance    AdminRole = "finance"
	AdminRoleOperator   AdminRole = "operator"
)

// Admin 管理员账号表 admin
type Admin struct {
	BaseModel
	Username     string    `gorm:"size:64;uniqueIndex;not null" json:"username"`
	PasswordHash string    `gorm:"size:128;not null" json:"-"`
	Role         AdminRole `gorm:"size:32;not null;index" json:"role"`
	Phone        string    `gorm:"size:20" json:"phone"`
	Status       int8      `gorm:"default:1;index" json:"status"` // 1正常 0禁用
	RealName     string    `gorm:"size:64" json:"real_name"`
}

func (Admin) TableName() string { return "admin" }

// StoreStatus 门店状态
const (
	StoreStatusNormal   int8 = 1 // 正常
	StoreStatusDisabled int8 = 0 // 停用
)

// Store 门店信息表 store（无数量上限，store_id 使用雪花ID）
// 合规：支持无上限自定义新增门店，不限制数量
type Store struct {
	BaseModel
	StoreCode   string `gorm:"size:64;uniqueIndex" json:"store_code"`       // 可选业务编号
	StoreName   string `gorm:"size:128;not null;index" json:"store_name"`
	Address     string `gorm:"size:512" json:"address"`
	TaxSubject  string `gorm:"size:256" json:"tax_subject"` // 个体户主体全称
	SubjectInfo string `gorm:"size:256" json:"-"`           // 兼容旧字段，写入时同步 TaxSubject
	Status      int8   `gorm:"default:1;index" json:"status"`
	Remark      string `gorm:"size:512" json:"remark"`
}

func (Store) TableName() string { return "store" }

// PayType 支付渠道类型：1微信 2支付宝
type PayType int8

const (
	PayTypeWechat PayType = 1
	PayTypeAlipay PayType = 2
)

func (p PayType) String() string {
	switch p {
	case PayTypeWechat:
		return "wechat"
	case PayTypeAlipay:
		return "alipay"
	default:
		return "unknown"
	}
}

func PayTypeFromString(s string) PayType {
	switch s {
	case "wechat", "1":
		return PayTypeWechat
	case "alipay", "2":
		return PayTypeAlipay
	default:
		return 0
	}
}

func (p PayType) Valid() bool {
	return p == PayTypeWechat || p == PayTypeAlipay
}

// PayChannel 门店支付渠道配置表 pay_channel
// 支持同一门店同类型多商户码；pool_enabled=1 参与代收轮询池
type PayChannel struct {
	BaseModel
	StoreID        uint64  `gorm:"index:idx_channel_store;not null" json:"store_id,string"`
	PayType        PayType `gorm:"index:idx_channel_store_pay;not null" json:"pay_type"` // 1微信 2支付宝
	Status         int8    `gorm:"default:1;index" json:"status"`                          // 1启用 0关停
	PoolEnabled    int8    `gorm:"default:0;index" json:"pool_enabled"`                    // 1参与轮询代收
	DailyLimitFen  int64   `gorm:"default:0" json:"daily_limit_fen"`                       // 日收款上限(分)，0不限
	SingleLimitFen int64   `gorm:"default:0" json:"single_limit_fen"`                      // 单笔上限(分)，0不限
	DailyUsedFen   int64   `gorm:"default:0" json:"daily_used_fen"`                        // 今日已收(分)
	DailyResetDate string  `gorm:"size:10" json:"daily_reset_date"`                        // YYYY-MM-DD
	RotateWeight   int     `gorm:"default:1" json:"rotate_weight"`                         // 轮询权重
	PlatformID     uint64  `gorm:"index:idx_channel_platform;default:0" json:"platform_id,string"` // 所属代收平台，0=未分配
	MchNo          string  `gorm:"size:64" json:"mch_no"`
	MchKey     string  `gorm:"size:256" json:"-"`                                        // 商户密钥/APIv3Key
	AppID      string  `gorm:"size:64" json:"app_id"`
	SerialNo   string  `gorm:"size:64" json:"serial_no"` // 微信证书序列号
	PrivateKey string  `gorm:"type:text" json:"-"`
	PublicKey  string  `gorm:"type:text" json:"-"`
	NotifyURL  string  `gorm:"size:512" json:"notify_url"`
	CertFile   string  `gorm:"size:512" json:"cert_file"` // 证书路径
	Remark     string  `gorm:"size:256" json:"remark"`
}

func (PayChannel) TableName() string { return "pay_channel" }

// PayPlatform 代收平台（码池组）：对接方通过 app_key 或 IP 识别，绑定独立商户码集合
type PayPlatform struct {
	BaseModel
	PlatformCode string `gorm:"size:64;uniqueIndex" json:"platform_code"`
	PlatformName string `gorm:"size:128;not null;index" json:"platform_name"`
	AppKey       string `gorm:"size:64;uniqueIndex;not null" json:"app_key"`
	AllowedIPs   string `gorm:"size:1024" json:"allowed_ips"` // 逗号分隔 IP/CIDR，可选；配合 app_key 或 IP 识别
	Status       int8   `gorm:"default:1;index" json:"status"`  // 1启用 0停用
	Remark       string `gorm:"size:512" json:"remark"`
}

func (PayPlatform) TableName() string { return "pay_platform" }

const PayPlatformStatusNormal int8 = 1

// Enabled 渠道是否启用
func (c *PayChannel) Enabled() bool { return c.Status == 1 }

// CanAccept 校验单笔与日额度（amount 单位：分）
func (c *PayChannel) CanAccept(amount int64) bool {
	if c.SingleLimitFen > 0 && amount > c.SingleLimitFen {
		return false
	}
	if c.DailyLimitFen > 0 && c.DailyUsedFen+amount > c.DailyLimitFen {
		return false
	}
	return true
}

// OrderStatus 订单状态
type OrderStatus int8

const (
	OrderStatusPending OrderStatus = 0 // 待支付
	OrderStatusPaid    OrderStatus = 1 // 已支付
	OrderStatusClosed  OrderStatus = 2 // 已关闭
	OrderStatusRefund  OrderStatus = 3 // 退款
)

// Order 订单主表 orders（预留分表：store_id + created_at）
type Order struct {
	BaseModel
	OrderNo       string      `gorm:"size:32;uniqueIndex;not null" json:"order_id"`
	PlatformID    uint64      `gorm:"index;default:0" json:"platform_id,string"` // 代收平台
	StoreID       uint64      `gorm:"index:idx_store_created;not null" json:"store_id,string"`
	ChannelID     uint64      `gorm:"index" json:"channel_id,string"` // 下单时绑定的 pay_channel.id，回调优先用历史密钥验签
	PayType       PayType     `gorm:"not null;index" json:"pay_type"`
	TotalAmount   int64       `gorm:"not null" json:"total_amount"`
	PayAmount     int64       `gorm:"default:0" json:"pay_amount"`
	OrderStatus   OrderStatus `gorm:"default:0;index" json:"order_status"`
	DeviceSN      string      `gorm:"size:64;index" json:"device_sn"`
	Subject       string      `gorm:"size:256" json:"subject"`
	NotifyData    string      `gorm:"type:text" json:"notify_data,omitempty"`
	PayTime       *time.Time  `json:"pay_time"`
	TransactionID string      `gorm:"size:64;index" json:"transaction_id"`
	PayScene      string      `gorm:"size:16;default:h5" json:"pay_scene"`
	QRCodeURL     string      `gorm:"size:512" json:"qr_code_url"`
}

func (Order) TableName() string { return "orders" }

// PayChannelHistory 渠道历史密钥（更换商户时归档旧配置，7天内可用于旧订单回调验签）
// 合规：仅服务端验签使用，不对外暴露；定期清理过期记录
type PayChannelHistory struct {
	BaseModel
	ChannelID  uint64    `gorm:"index:idx_channel_expires;not null" json:"channel_id"`
	StoreID    uint64    `gorm:"index" json:"store_id"`
	PayType    PayType   `gorm:"not null" json:"pay_type"`
	MchNo      string    `gorm:"size:64" json:"mch_no"`
	MchKey     string    `gorm:"size:256" json:"-"`
	AppID      string    `gorm:"size:64" json:"app_id"`
	SerialNo   string    `gorm:"size:64" json:"serial_no"`
	PrivateKey string    `gorm:"type:text" json:"-"`
	PublicKey  string    `gorm:"type:text" json:"-"`
	ExpiresAt  time.Time `gorm:"index:idx_channel_expires" json:"expires_at"`
}

func (PayChannelHistory) TableName() string { return "pay_channel_history" }

// ToPayChannel 转为验签用的渠道快照
func (h *PayChannelHistory) ToPayChannel() PayChannel {
	return PayChannel{
		StoreID:    h.StoreID,
		PayType:    h.PayType,
		MchNo:      h.MchNo,
		MchKey:     h.MchKey,
		AppID:      h.AppID,
		SerialNo:   h.SerialNo,
		PrivateKey: h.PrivateKey,
		PublicKey:  h.PublicKey,
	}
}

// SensitiveLog 敏感操作日志（6个月自动清理）
type SensitiveLog struct {
	BaseModel
	Action  string `gorm:"size:64;index" json:"action"`
	Content string `gorm:"type:text" json:"content"`
	AdminID uint64 `gorm:"index" json:"admin_id"`
	IP      string `gorm:"size:64" json:"ip"`
}

func (SensitiveLog) TableName() string { return "sensitive_logs" }

// IPWhitelistScope 白名单作用域
type IPWhitelistScope string

const (
	IPWhitelistScopeAdmin        IPWhitelistScope = "admin"
	IPWhitelistScopeCallback     IPWhitelistScope = "callback"
	IPWhitelistScopePay          IPWhitelistScope = "pay" // 收银端 /api/pay/*
	IPWhitelistScopeTrustedProxy IPWhitelistScope = "trusted_proxy"
)

func (s IPWhitelistScope) Valid() bool {
	switch s {
	case IPWhitelistScopeAdmin, IPWhitelistScopeCallback, IPWhitelistScopePay, IPWhitelistScopeTrustedProxy:
		return true
	default:
		return false
	}
}

// IPWhitelistPolicy 各作用域是否启用限制（1启用 0不限制）
type IPWhitelistPolicy struct {
	Scope     string    `gorm:"size:32;primaryKey" json:"scope"`
	Enabled   int8      `gorm:"default:0;index" json:"enabled"`
	Remark    string    `gorm:"size:256" json:"remark"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (IPWhitelistPolicy) TableName() string { return "ip_whitelist_policy" }

// IPWhitelistEntry IP/CIDR 白名单条目
type IPWhitelistEntry struct {
	BaseModel
	Scope  string `gorm:"size:32;index:idx_scope_cidr,unique;not null" json:"scope"`
	CIDR   string `gorm:"size:64;index:idx_scope_cidr,unique;not null" json:"cidr"`
	Remark string `gorm:"size:256" json:"remark"`
	Status int8   `gorm:"default:1;index" json:"status"` // 1启用 0停用
}

func (IPWhitelistEntry) TableName() string { return "ip_whitelist_entry" }
