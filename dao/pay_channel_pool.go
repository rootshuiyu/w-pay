package dao

import (
	"time"

	"wpay/model"

	"gorm.io/gorm"
)

func (d *PayChannelDAO) ListByStoreAndType(storeID uint64, payType model.PayType) ([]model.PayChannel, error) {
	var list []model.PayChannel
	err := DB.Where("store_id = ? AND pay_type = ? AND status = 1", storeID, payType).
		Order("id ASC").Find(&list).Error
	return list, err
}

// ListPoolCandidates 轮询池候选：启用 + 参与轮询 + 指定支付方式 + 平台
func (d *PayChannelDAO) ListPoolCandidates(payType model.PayType, platformID, storeID uint64) ([]model.PayChannel, error) {
	q := DB.Where("status = 1 AND pool_enabled = 1 AND pay_type = ?", payType)
	if platformID > 0 {
		q = q.Where("platform_id = ?", platformID)
	} else {
		q = q.Where("platform_id = 0")
	}
	if storeID > 0 {
		q = q.Where("store_id = ?", storeID)
	}
	var list []model.PayChannel
	err := q.Order("id ASC").Find(&list).Error
	return list, err
}

// ListPlatformSingleChannels 平台单码（不参与轮询但已绑定平台）
func (d *PayChannelDAO) ListPlatformSingleChannels(payType model.PayType, platformID, storeID uint64) ([]model.PayChannel, error) {
	if platformID == 0 {
		return nil, nil
	}
	q := DB.Where("status = 1 AND pool_enabled = 0 AND pay_type = ? AND platform_id = ?", payType, platformID)
	if storeID > 0 {
		q = q.Where("store_id = ?", storeID)
	}
	var list []model.PayChannel
	err := q.Order("id ASC").Find(&list).Error
	return list, err
}

func (d *PayChannelDAO) ResetDailyUsedIfNeeded(ch *model.PayChannel) error {
	today := time.Now().Format("2006-01-02")
	if ch.DailyResetDate == today {
		return nil
	}
	return DB.Model(ch).Updates(map[string]interface{}{
		"daily_used_fen":   0,
		"daily_reset_date": today,
	}).Error
}

func (d *PayChannelDAO) AddDailyUsed(id uint64, amount int64) error {
	return DB.Model(&model.PayChannel{}).Where("id = ?", id).
		UpdateColumn("daily_used_fen", gorm.Expr("daily_used_fen + ?", amount)).Error
}

func (d *PayChannelDAO) ResetAllDailyUsed() (int64, error) {
	today := time.Now().Format("2006-01-02")
	result := DB.Model(&model.PayChannel{}).Where("daily_reset_date <> ? OR daily_reset_date IS NULL OR daily_reset_date = ''", today).
		Updates(map[string]interface{}{"daily_used_fen": 0, "daily_reset_date": today})
	return result.RowsAffected, result.Error
}

func (d *PayChannelDAO) ListPoolStats(payType *model.PayType, platformID *uint64) ([]model.PayChannel, error) {
	q := DB.Where("pool_enabled = 1 AND status = 1")
	if payType != nil {
		q = q.Where("pay_type = ?", *payType)
	}
	if platformID != nil {
		if *platformID > 0 {
			q = q.Where("platform_id = ?", *platformID)
		} else {
			q = q.Where("platform_id = 0")
		}
	}
	var list []model.PayChannel
	err := q.Order("pay_type ASC, id ASC").Find(&list).Error
	return list, err
}
