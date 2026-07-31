package dao

import (
	"wpay/model"

	"gorm.io/gorm"
)

type PayChannelDAO struct{}

func NewPayChannelDAO() *PayChannelDAO { return &PayChannelDAO{} }

func (d *PayChannelDAO) Create(ch *model.PayChannel) error {
	return DB.Create(ch).Error
}

func (d *PayChannelDAO) Update(ch *model.PayChannel) error {
	return DB.Save(ch).Error
}

func (d *PayChannelDAO) Delete(id uint64) error {
	return DB.Delete(&model.PayChannel{}, id).Error
}

func (d *PayChannelDAO) FindByID(id uint64) (*model.PayChannel, error) {
	var ch model.PayChannel
	err := DB.First(&ch, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &ch, err
}

func (d *PayChannelDAO) FindByStoreAndType(storeID uint64, payType model.PayType) (*model.PayChannel, error) {
	var ch model.PayChannel
	err := DB.Where("store_id = ? AND pay_type = ? AND status = 1", storeID, payType).
		Order("id ASC").First(&ch).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &ch, err
}

func (d *PayChannelDAO) ListByStore(storeID uint64) ([]model.PayChannel, error) {
	var list []model.PayChannel
	err := DB.Where("store_id = ?", storeID).Find(&list).Error
	return list, err
}

func (d *PayChannelDAO) List(page, pageSize int, storeID *uint64) ([]model.PayChannel, int64, error) {
	var list []model.PayChannel
	var total int64
	q := DB.Model(&model.PayChannel{})
	if storeID != nil {
		q = q.Where("store_id = ?", *storeID)
	}
	q.Count(&total)
	err := q.Offset((page - 1) * pageSize).Limit(pageSize).Order("id DESC").Find(&list).Error
	return list, total, err
}

// ListAllEnabled 获取所有启用的渠道（预热用）
func (d *PayChannelDAO) ListAllEnabled() ([]model.PayChannel, error) {
	var list []model.PayChannel
	err := DB.Where("status = 1 AND deleted_at IS NULL").Order("id ASC").Find(&list).Error
	return list, err
}

// CountEnabledByStoreAndType 统计门店某支付方式下启用渠道数
func (d *PayChannelDAO) CountEnabledByStoreAndType(storeID uint64, payType model.PayType) (int64, error) {
	var n int64
	err := DB.Model(&model.PayChannel{}).
		Where("store_id = ? AND pay_type = ? AND status = 1", storeID, payType).
		Count(&n).Error
	return n, err
}

// FindByMchNo 通过商户号和支付类型查找渠道
func (d *PayChannelDAO) FindByMchNo(mchNo string, payType model.PayType) (*model.PayChannel, error) {
	var ch model.PayChannel
	err := DB.Where("mch_no = ? AND pay_type = ?", mchNo, payType).
		Order("id ASC").First(&ch).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &ch, err
}
