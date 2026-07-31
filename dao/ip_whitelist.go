package dao

import (
	"wpay/model"

	"gorm.io/gorm"
)

type IPWhitelistDAO struct{}

func NewIPWhitelistDAO() *IPWhitelistDAO { return &IPWhitelistDAO{} }

func (d *IPWhitelistDAO) EnsurePolicies() error {
	scopes := []string{
		string(model.IPWhitelistScopeAdmin),
		string(model.IPWhitelistScopeCallback),
		string(model.IPWhitelistScopePay),
		string(model.IPWhitelistScopeTrustedProxy),
	}
	for _, scope := range scopes {
		var count int64
		if err := DB.Model(&model.IPWhitelistPolicy{}).Where("scope = ?", scope).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			if err := DB.Create(&model.IPWhitelistPolicy{Scope: scope, Enabled: 0}).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *IPWhitelistDAO) ListPolicies() ([]model.IPWhitelistPolicy, error) {
	var list []model.IPWhitelistPolicy
	err := DB.Order("scope ASC").Find(&list).Error
	return list, err
}

func (d *IPWhitelistDAO) UpdatePolicy(scope string, enabled int8, remark string) error {
	return DB.Model(&model.IPWhitelistPolicy{}).Where("scope = ?", scope).
		Updates(map[string]interface{}{"enabled": enabled, "remark": remark}).Error
}

func (d *IPWhitelistDAO) CreateEntry(entry *model.IPWhitelistEntry) error {
	return DB.Create(entry).Error
}

func (d *IPWhitelistDAO) UpdateEntry(entry *model.IPWhitelistEntry) error {
	return DB.Save(entry).Error
}

func (d *IPWhitelistDAO) FindEntryByID(id uint64) (*model.IPWhitelistEntry, error) {
	var entry model.IPWhitelistEntry
	err := DB.First(&entry, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &entry, err
}

func (d *IPWhitelistDAO) FindEntry(scope, cidr string) (*model.IPWhitelistEntry, error) {
	var entry model.IPWhitelistEntry
	err := DB.Where("scope = ? AND cidr = ?", scope, cidr).First(&entry).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &entry, err
}

func (d *IPWhitelistDAO) DeleteEntry(id uint64) error {
	return DB.Delete(&model.IPWhitelistEntry{}, id).Error
}

func (d *IPWhitelistDAO) CountEntries(scope string) (int64, error) {
	var count int64
	q := DB.Model(&model.IPWhitelistEntry{})
	if scope != "" {
		q = q.Where("scope = ?", scope)
	}
	err := q.Count(&count).Error
	return count, err
}

func (d *IPWhitelistDAO) ListEntries(scope string, status *int8) ([]model.IPWhitelistEntry, error) {
	var list []model.IPWhitelistEntry
	q := DB.Model(&model.IPWhitelistEntry{})
	if scope != "" {
		q = q.Where("scope = ?", scope)
	}
	if status != nil {
		q = q.Where("status = ?", *status)
	}
	err := q.Order("id ASC").Find(&list).Error
	return list, err
}

func (d *IPWhitelistDAO) ListActiveCIDRs(scope string) ([]string, error) {
	var list []model.IPWhitelistEntry
	err := DB.Where("scope = ? AND status = 1", scope).Order("id ASC").Find(&list).Error
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(list))
	for _, e := range list {
		out = append(out, e.CIDR)
	}
	return out, nil
}
