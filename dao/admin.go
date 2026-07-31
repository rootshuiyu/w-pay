package dao

import (
	"wpay/model"

	"gorm.io/gorm"
)

// AdminDAO 管理员数据访问
type AdminDAO struct{}

func NewAdminDAO() *AdminDAO { return &AdminDAO{} }

func (d *AdminDAO) FindByUsername(username string) (*model.Admin, error) {
	var admin model.Admin
	err := DB.Where("username = ? AND status = 1", username).First(&admin).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &admin, err
}

func (d *AdminDAO) FindByID(id uint64) (*model.Admin, error) {
	var admin model.Admin
	err := DB.First(&admin, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &admin, err
}

func (d *AdminDAO) Create(admin *model.Admin) error {
	return DB.Create(admin).Error
}

func (d *AdminDAO) Update(admin *model.Admin) error {
	return DB.Save(admin).Error
}

func (d *AdminDAO) List(page, pageSize int) ([]model.Admin, int64, error) {
	var list []model.Admin
	var total int64
	q := DB.Model(&model.Admin{})
	q.Count(&total)
	err := q.Offset((page - 1) * pageSize).Limit(pageSize).Order("id DESC").Find(&list).Error
	return list, total, err
}
