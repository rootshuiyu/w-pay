package dao

import (
	"time"

	"wpay/model"
)

const ChannelHistoryRetentionDays = 7

type PayChannelHistoryDAO struct{}

func NewPayChannelHistoryDAO() *PayChannelHistoryDAO { return &PayChannelHistoryDAO{} }

// SaveSnapshot 更换商户前归档旧密钥，供旧订单回调验签（保留 7 天）
func (d *PayChannelHistoryDAO) SaveSnapshot(ch *model.PayChannel) error {
	h := &model.PayChannelHistory{
		ChannelID:  ch.ID,
		StoreID:    ch.StoreID,
		PayType:    ch.PayType,
		MchNo:      ch.MchNo,
		MchKey:     ch.MchKey,
		AppID:      ch.AppID,
		SerialNo:   ch.SerialNo,
		PrivateKey: ch.PrivateKey,
		PublicKey:  ch.PublicKey,
		ExpiresAt:  time.Now().Add(ChannelHistoryRetentionDays * 24 * time.Hour),
	}
	return DB.Create(h).Error
}

// ListValidByStoreAndType 获取未过期的历史密钥（微信回调逐条尝试验签）
func (d *PayChannelHistoryDAO) ListValidByStoreAndType(storeID uint64, payType model.PayType) ([]model.PayChannelHistory, error) {
	var list []model.PayChannelHistory
	err := DB.Where("store_id = ? AND pay_type = ? AND expires_at > ?", storeID, payType, time.Now()).
		Order("created_at DESC").Find(&list).Error
	return list, err
}

// ListValidByChannelID 按渠道 ID 获取未过期历史（支付宝回调优先匹配下单渠道）
func (d *PayChannelHistoryDAO) ListValidByChannelID(channelID uint64) ([]model.PayChannelHistory, error) {
	var list []model.PayChannelHistory
	err := DB.Where("channel_id = ? AND expires_at > ?", channelID, time.Now()).
		Order("created_at DESC").Find(&list).Error
	return list, err
}

// DeleteExpired 清理过期历史密钥
func (d *PayChannelHistoryDAO) DeleteExpired(before time.Time) (int64, error) {
	result := DB.Unscoped().Where("expires_at < ?", before).Delete(&model.PayChannelHistory{})
	return result.RowsAffected, result.Error
}

// CredentialChanged 判断商户核心参数是否变更
func CredentialChanged(ch *model.PayChannel, req MchCredentialPatch) bool {
	if req.MchNo != "" && req.MchNo != ch.MchNo {
		return true
	}
	if req.MchKey != "" && req.MchKey != ch.MchKey {
		return true
	}
	if req.SerialNo != "" && req.SerialNo != ch.SerialNo {
		return true
	}
	if req.PrivateKey != "" && req.PrivateKey != ch.PrivateKey {
		return true
	}
	if req.PublicKey != "" && req.PublicKey != ch.PublicKey {
		return true
	}
	return false
}

// MchCredentialPatch 商户凭证变更字段
type MchCredentialPatch struct {
	MchNo      string
	MchKey     string
	SerialNo   string
	PrivateKey string
	PublicKey  string
}
