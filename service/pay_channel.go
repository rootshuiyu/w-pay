package service

import (
	"strconv"

	"wpay/common"
	"wpay/dao"
	"wpay/model"
)

type PayChannelService struct {
	channelDAO *dao.PayChannelDAO
	storeDAO   *dao.StoreDAO
}

func NewPayChannelService() *PayChannelService {
	return &PayChannelService{
		channelDAO: dao.NewPayChannelDAO(),
		storeDAO:   dao.NewStoreDAO(),
	}
}

type ChannelCreateRequest struct {
	StoreID        common.FlexUint64 `json:"store_id" binding:"required"`
	PayType        model.PayType     `json:"pay_type" binding:"required"`
	PoolEnabled    *int8             `json:"pool_enabled"`
	DailyLimitFen  int64             `json:"daily_limit_fen"`
	SingleLimitFen int64             `json:"single_limit_fen"`
	MchNo          string            `json:"mch_no" binding:"required"`
	MchKey     string        `json:"mch_key"`
	AppID      string        `json:"app_id"`
	SerialNo   string        `json:"serial_no"`
	PrivateKey string        `json:"private_key"`
	PublicKey  string        `json:"public_key"`
	NotifyURL  string        `json:"notify_url"`
	CertFile   string        `json:"cert_file"`
	Remark     string        `json:"remark"`
}

type ChannelUpdateRequest struct {
	Status         *int8  `json:"status"`
	PoolEnabled    *int8  `json:"pool_enabled"`
	DailyLimitFen  *int64 `json:"daily_limit_fen"`
	SingleLimitFen *int64 `json:"single_limit_fen"`
	MchNo          string `json:"mch_no"`
	MchKey     string `json:"mch_key"`
	AppID      string `json:"app_id"`
	SerialNo   string `json:"serial_no"`
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
	NotifyURL  string `json:"notify_url"`
	CertFile   string `json:"cert_file"`
	Remark     string `json:"remark"`
}

type ChannelEditRequest struct {
	ID             common.FlexUint64 `json:"id" binding:"required"`
	PoolEnabled    *int8             `json:"pool_enabled"`
	DailyLimitFen  *int64            `json:"daily_limit_fen"`
	SingleLimitFen *int64            `json:"single_limit_fen"`
	MchNo          string            `json:"mch_no"`
	MchKey     string `json:"mch_key"`
	AppID      string `json:"app_id"`
	SerialNo   string `json:"serial_no"`
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
	NotifyURL  string `json:"notify_url"`
	CertFile   string `json:"cert_file"`
	Remark     string `json:"remark"`
}

type ChannelStatusRequest struct {
	ID     common.FlexUint64 `json:"id" binding:"required"`
	Status int8              `json:"status"` // 1启用 0关停
}

// PoolAddRequest 公共码池入库（无需选门店）
type PoolAddRequest struct {
	PayType        model.PayType `json:"pay_type" binding:"required"`
	MchNo          string        `json:"mch_no" binding:"required"`
	PoolEnabled    *int8         `json:"pool_enabled"`
	DailyLimitFen  int64         `json:"daily_limit_fen"`
	SingleLimitFen int64         `json:"single_limit_fen"`
	MchKey         string        `json:"mch_key"`
	AppID          string        `json:"app_id"`
	SerialNo       string        `json:"serial_no"`
	PrivateKey     string        `json:"private_key"`
	PublicKey      string        `json:"public_key"`
	Remark         string        `json:"remark"`
}

// PoolAdd 录入商户码到公共码池
func (s *PayChannelService) PoolAdd(req PoolAddRequest) (*model.PayChannel, error) {
	storeSvc := NewStoreService()
	st, err := storeSvc.EnsurePoolStore()
	if err != nil {
		return nil, err
	}
	poolEnabled := int8(1)
	if req.PoolEnabled != nil {
		poolEnabled = *req.PoolEnabled
	} else {
		// 前端传空时，默认启用参与轮询（与界面开关一致）
		poolEnabled = 1
	}
	return s.Create(ChannelCreateRequest{
		StoreID:        common.FlexUint64(st.ID),
		PayType:        req.PayType,
		PoolEnabled:    &poolEnabled,
		DailyLimitFen:  req.DailyLimitFen,
		SingleLimitFen: req.SingleLimitFen,
		MchNo:          req.MchNo,
		MchKey:         req.MchKey,
		AppID:          req.AppID,
		SerialNo:       req.SerialNo,
		PrivateKey:     req.PrivateKey,
		PublicKey:      req.PublicKey,
		Remark:         req.Remark,
	})
}

// Create 绑定门店独立支付渠道，成功后 Delete Redis 缓存 Key 实现热更新
func (s *PayChannelService) Create(req ChannelCreateRequest) (*model.PayChannel, error) {
	storeID := req.StoreID.Uint64()
	store, err := s.storeDAO.FindByID(storeID)
	if err != nil {
		return nil, err
	}
	if store == nil {
		return nil, common.ErrInvalidInput("门店不存在")
	}
	if !req.PayType.Valid() {
		return nil, common.ErrInvalidInput("pay_type 无效，1=微信 2=支付宝")
	}
	
	// 检查商户号是否已存在（同一支付类型下）
	exist, err := s.channelDAO.FindByMchNo(req.MchNo, req.PayType)
	if err != nil {
		return nil, err
	}
	if exist != nil {
		return nil, common.ErrInvalidInput("该商户号已存在，请勿重复录入")
	}
	
	poolEnabled := int8(1)
	if req.PoolEnabled != nil {
		poolEnabled = *req.PoolEnabled
	}
	ch := &model.PayChannel{
		StoreID:        storeID,
		PayType:        req.PayType,
		Status:         1,
		PoolEnabled:    poolEnabled,
		DailyLimitFen:  req.DailyLimitFen,
		SingleLimitFen: req.SingleLimitFen,
		RotateWeight:   1,
		MchNo:          req.MchNo,
		MchKey:     req.MchKey,
		AppID:      req.AppID,
		SerialNo:   req.SerialNo,
		PrivateKey: req.PrivateKey,
		PublicKey:  req.PublicKey,
		NotifyURL:  req.NotifyURL,
		CertFile:   req.CertFile,
		Remark:     req.Remark,
	}
	if err := s.channelDAO.Create(ch); err != nil {
		return nil, err
	}
	// 渠道热更新：Delete 对应 store:channel:{store_id}:{pay_type}，下次下单自动回源
	_ = dao.DeleteChannelCache(storeID, req.PayType)
	return ch, nil
}

func (s *PayChannelService) Update(id uint64, req ChannelUpdateRequest) (*model.PayChannel, error) {
	ch, err := s.channelDAO.FindByID(id)
	if err != nil {
		return nil, err
	}
	if ch == nil {
		return nil, common.ErrInvalidInput("渠道配置不存在")
	}

	patch := dao.MchCredentialPatch{
		MchNo: req.MchNo, MchKey: req.MchKey, SerialNo: req.SerialNo,
		PrivateKey: req.PrivateKey, PublicKey: req.PublicKey,
	}
	// 更换商户核心参数前归档旧密钥，旧订单回调 7 天内仍可用旧密钥验签
	if dao.CredentialChanged(ch, patch) {
		if err := dao.NewPayChannelHistoryDAO().SaveSnapshot(ch); err != nil {
			common.Log.Warn("save channel history failed: %v", err)
		}
	}

	if req.Status != nil {
		ch.Status = *req.Status
	}
	if req.PoolEnabled != nil {
		ch.PoolEnabled = *req.PoolEnabled
	}
	if req.DailyLimitFen != nil {
		ch.DailyLimitFen = *req.DailyLimitFen
	}
	if req.SingleLimitFen != nil {
		ch.SingleLimitFen = *req.SingleLimitFen
	}
	if req.MchNo != "" {
		ch.MchNo = req.MchNo
	}
	if req.MchKey != "" {
		ch.MchKey = req.MchKey
	}
	if req.AppID != "" {
		ch.AppID = req.AppID
	}
	if req.SerialNo != "" {
		ch.SerialNo = req.SerialNo
	}
	if req.PrivateKey != "" {
		ch.PrivateKey = req.PrivateKey
	}
	if req.PublicKey != "" {
		ch.PublicKey = req.PublicKey
	}
	if req.NotifyURL != "" {
		ch.NotifyURL = req.NotifyURL
	}
	if req.CertFile != "" {
		ch.CertFile = req.CertFile
	}
	if req.Remark != "" {
		ch.Remark = req.Remark
	}
	if err := s.channelDAO.Update(ch); err != nil {
		return nil, err
	}
	_ = dao.DeleteChannelCache(ch.StoreID, ch.PayType)
	return ch, nil
}

func (s *PayChannelService) ListByStore(storeID uint64) ([]model.PayChannel, error) {
	return s.channelDAO.ListByStore(storeID)
}

func MaskChannel(ch *model.PayChannel) map[string]interface{} {
	return map[string]interface{}{
		"id":               strconv.FormatUint(ch.ID, 10),
		"store_id":         strconv.FormatUint(ch.StoreID, 10),
		"pay_type":         ch.PayType,
		"status":           ch.Status,
		"pool_enabled":     ch.PoolEnabled,
		"platform_id":      strconv.FormatUint(ch.PlatformID, 10),
		"daily_limit_fen":  ch.DailyLimitFen,
		"single_limit_fen": ch.SingleLimitFen,
		"daily_used_fen":   ch.DailyUsedFen,
		"daily_reset_date": ch.DailyResetDate,
		"mch_no":           ch.MchNo,
		"app_id":      ch.AppID,
		"serial_no":   ch.SerialNo,
		"notify_url":  ch.NotifyURL,
		"cert_file":   ch.CertFile,
		"remark":      ch.Remark,
		"mch_key":     common.MaskPrivateKey(ch.MchKey),
		"private_key": common.MaskPrivateKey(ch.PrivateKey),
		"public_key":  common.MaskPrivateKey(ch.PublicKey),
		"created_at":  ch.CreatedAt,
		"updated_at":  ch.UpdatedAt,
	}
}
