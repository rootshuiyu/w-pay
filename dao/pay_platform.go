package dao

import (
	"wpay/model"

	"gorm.io/gorm"
)

type PayPlatformDAO struct{}

func NewPayPlatformDAO() *PayPlatformDAO { return &PayPlatformDAO{} }

func (d *PayPlatformDAO) Create(p *model.PayPlatform) error {
	q := DB
	if p.PlatformCode == "" {
		q = q.Omit("platform_code")
	}
	return q.Create(p).Error
}

func (d *PayPlatformDAO) Update(p *model.PayPlatform) error {
	if p.PlatformCode == "" {
		return DB.Model(p).Omit("platform_code").Updates(p).Error
	}
	return DB.Save(p).Error
}

func (d *PayPlatformDAO) Delete(id uint64) error {
	return DB.Delete(&model.PayPlatform{}, id).Error
}

func (d *PayPlatformDAO) FindByID(id uint64) (*model.PayPlatform, error) {
	var p model.PayPlatform
	err := DB.First(&p, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &p, err
}

func (d *PayPlatformDAO) FindByAppKey(appKey string) (*model.PayPlatform, error) {
	if appKey == "" {
		return nil, nil
	}
	var p model.PayPlatform
	err := DB.Where("app_key = ?", appKey).First(&p).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &p, err
}

func (d *PayPlatformDAO) FindByCode(code string) (*model.PayPlatform, error) {
	if code == "" {
		return nil, nil
	}
	var p model.PayPlatform
	err := DB.Where("platform_code = ?", code).First(&p).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &p, err
}

func (d *PayPlatformDAO) List(keyword string, status *int8, page, pageSize int) ([]model.PayPlatform, int64, error) {
	var list []model.PayPlatform
	var total int64
	q := DB.Model(&model.PayPlatform{})
	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("platform_name LIKE ? OR platform_code LIKE ? OR app_key LIKE ?", like, like, like)
	}
	if status != nil {
		q = q.Where("status = ?", *status)
	}
	q.Count(&total)
	err := q.Offset((page - 1) * pageSize).Limit(pageSize).Order("id DESC").Find(&list).Error
	return list, total, err
}

func (d *PayPlatformDAO) ListActive() ([]model.PayPlatform, error) {
	var list []model.PayPlatform
	err := DB.Where("status = ?", model.PayPlatformStatusNormal).Order("id ASC").Find(&list).Error
	return list, err
}

func (d *PayPlatformDAO) CountChannels(platformID uint64) (int64, error) {
	var n int64
	err := DB.Model(&model.PayChannel{}).Where("platform_id = ? AND status = 1", platformID).Count(&n).Error
	return n, err
}

func (d *PayPlatformDAO) ListChannels(platformID uint64) ([]model.PayChannel, error) {
	var list []model.PayChannel
	err := DB.Where("platform_id = ?", platformID).Order("pay_type ASC, id ASC").Find(&list).Error
	return list, err
}

// SetChannels 重置平台绑定的商户码（未在列表中的解绑，列表内的绑定到该平台）
func (d *PayPlatformDAO) SetChannels(platformID uint64, channelIDs []uint64) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.PayChannel{}).Where("platform_id = ?", platformID).
			Update("platform_id", 0).Error; err != nil {
			return err
		}
		if len(channelIDs) == 0 {
			return nil
		}
		return tx.Model(&model.PayChannel{}).Where("id IN ?", channelIDs).
			Update("platform_id", platformID).Error
	})
}

func (d *PayPlatformDAO) AppendChannel(platformID, channelID uint64) error {
	return DB.Model(&model.PayChannel{}).Where("id = ?", channelID).Update("platform_id", platformID).Error
}

func (d *PayPlatformDAO) UnbindChannel(channelID uint64) error {
	return DB.Model(&model.PayChannel{}).Where("id = ?", channelID).Update("platform_id", 0).Error
}

func (d *PayPlatformDAO) ListByIDs(ids []uint64) ([]model.PayPlatform, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var list []model.PayPlatform
	err := DB.Where("id IN ?", ids).Find(&list).Error
	return list, err
}

func (d *PayPlatformDAO) ListAllChannelsBrief(unassignedOnly bool) ([]model.PayChannel, error) {
	var list []model.PayChannel
	q := DB.Select("id", "store_id", "pay_type", "status", "pool_enabled", "platform_id", "mch_no", "daily_limit_fen", "daily_used_fen", "single_limit_fen").
		Where("status = 1 AND pool_enabled = 1")
	if unassignedOnly {
		q = q.Where("platform_id = 0")
	}
	err := q.Order("id ASC").Find(&list).Error
	return list, err
}
