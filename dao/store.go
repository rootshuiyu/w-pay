package dao

import (
	"wpay/model"

	"gorm.io/gorm"
)

type StoreDAO struct{}

func NewStoreDAO() *StoreDAO { return &StoreDAO{} }

func (d *StoreDAO) Create(store *model.Store) error {
	if store.TaxSubject != "" {
		store.SubjectInfo = store.TaxSubject
	}
	q := DB
	if store.StoreCode == "" {
		q = q.Omit("store_code")
	}
	return q.Create(store).Error
}

func (d *StoreDAO) Update(store *model.Store) error {
	if store.TaxSubject != "" {
		store.SubjectInfo = store.TaxSubject
	}
	return DB.Save(store).Error
}

func (d *StoreDAO) Delete(id uint64) error {
	return DB.Delete(&model.Store{}, id).Error
}

func (d *StoreDAO) FindByID(id uint64) (*model.Store, error) {
	var store model.Store
	err := DB.First(&store, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if store.TaxSubject == "" {
		store.TaxSubject = store.SubjectInfo
	}
	return &store, err
}

func (d *StoreDAO) FindByCode(code string) (*model.Store, error) {
	if code == "" {
		return nil, nil
	}
	var store model.Store
	err := DB.Where("store_code = ?", code).First(&store).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &store, err
}

func (d *StoreDAO) List(keyword string, status *int8, hideSystem bool, page, pageSize int) ([]model.Store, int64, error) {
	var list []model.Store
	var total int64
	q := DB.Model(&model.Store{})
	if hideSystem {
		q = q.Where("store_code IS DISTINCT FROM ? AND remark IS DISTINCT FROM ?", "__POOL__", "代收平台自动创建")
	}
	if keyword != "" {
		q = q.Where("store_name LIKE ? OR store_code LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if status != nil {
		q = q.Where("status = ?", *status)
	}
	q.Count(&total)
	err := q.Offset((page - 1) * pageSize).Limit(pageSize).Order("id DESC").Find(&list).Error
	for i := range list {
		if list[i].TaxSubject == "" {
			list[i].TaxSubject = list[i].SubjectInfo
		}
	}
	return list, total, err
}

func (d *StoreDAO) ListByIDs(ids []uint64) ([]model.Store, error) {
	var list []model.Store
	if len(ids) == 0 {
		return list, nil
	}
	err := DB.Where("id IN ?", ids).Find(&list).Error
	return list, err
}
