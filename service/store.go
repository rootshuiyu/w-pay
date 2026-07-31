package service

import (
	"wpay/common"
	"wpay/dao"
	"wpay/model"
)

type StoreService struct {
	storeDAO *dao.StoreDAO
}

func NewStoreService() *StoreService {
	return &StoreService{storeDAO: dao.NewStoreDAO()}
}

type StoreCreateRequest struct {
	StoreCode  string `json:"store_code"`
	StoreName  string `json:"store_name" binding:"required"`
	Address    string `json:"address"`
	TaxSubject string `json:"tax_subject"`
	Remark     string `json:"remark"`
}

type StoreUpdateRequest struct {
	StoreName  string `json:"store_name"`
	Address    string `json:"address"`
	TaxSubject string `json:"tax_subject"`
	Status     *int8  `json:"status"`
	Remark     string `json:"remark"`
}

type StoreEditRequest struct {
	ID         common.FlexUint64 `json:"id" binding:"required"`
	StoreName  string            `json:"store_name"`
	Address    string            `json:"address"`
	TaxSubject string            `json:"tax_subject"`
	Remark     string            `json:"remark"`
}

type StoreStatusRequest struct {
	ID     common.FlexUint64 `json:"id" binding:"required"`
	Status int8              `json:"status"` // 1正常 0停用
}

type StoreDeleteRequest struct {
	ID common.FlexUint64 `json:"id"`
}

const PoolStoreCode = "__POOL__"

// EnsurePoolStore 获取或创建公共码池门店（录入商户码统一归属）
func (s *StoreService) EnsurePoolStore() (*model.Store, error) {
	exist, err := s.storeDAO.FindByCode(PoolStoreCode)
	if err != nil {
		return nil, err
	}
	if exist != nil {
		return exist, nil
	}
	id, err := common.GenerateID()
	if err != nil {
		return nil, err
	}
	st := &model.Store{
		BaseModel:  model.BaseModel{ID: id},
		StoreCode:  PoolStoreCode,
		StoreName:  "公共码池",
		TaxSubject: "系统公共码池",
		Remark:     "商户码入库默认主体，无需手动选择门店",
		Status:     1,
	}
	if err := s.storeDAO.Create(st); err != nil {
		return nil, common.WrapDAOError(err)
	}
	return st, nil
}
func (s *StoreService) Create(req StoreCreateRequest) (*model.Store, error) {
	for _, field := range []string{req.StoreCode, req.StoreName, req.Address, req.TaxSubject, req.Remark} {
		if err := common.SanitizeString(field); err != nil {
			return nil, err
		}
	}
	if err := common.ValidateLength(req.StoreName, 1, 128, "门店名称"); err != nil {
		return nil, err
	}
	if req.StoreCode != "" {
		exist, err := s.storeDAO.FindByCode(req.StoreCode)
		if err != nil {
			return nil, err
		}
		if exist != nil {
			return nil, common.ErrInvalidInput("门店编号已存在")
		}
	}

	storeID, err := common.GenerateID()
	if err != nil {
		return nil, err
	}

	store := &model.Store{
		BaseModel:  model.BaseModel{ID: storeID},
		StoreCode:  req.StoreCode,
		StoreName:  req.StoreName,
		Address:    req.Address,
		TaxSubject: req.TaxSubject,
		Status:     model.StoreStatusNormal,
		Remark:     req.Remark,
	}
	if err := s.storeDAO.Create(store); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *StoreService) Update(id uint64, req StoreUpdateRequest) (*model.Store, error) {
	store, err := s.storeDAO.FindByID(id)
	if err != nil {
		return nil, err
	}
	if store == nil {
		return nil, common.ErrInvalidInput("门店不存在")
	}
	for _, field := range []string{req.StoreName, req.Address, req.TaxSubject, req.Remark} {
		if err := common.SanitizeString(field); err != nil {
			return nil, err
		}
	}
	if req.StoreName != "" {
		store.StoreName = req.StoreName
	}
	if req.Address != "" {
		store.Address = req.Address
	}
	if req.TaxSubject != "" {
		store.TaxSubject = req.TaxSubject
	}
	if req.Status != nil {
		store.Status = *req.Status
	}
	if req.Remark != "" {
		store.Remark = req.Remark
	}
	if err := s.storeDAO.Update(store); err != nil {
		return nil, err
	}
	// 门店状态变更：清除该门店全部渠道缓存，下次下单回源 DB
	_ = dao.DeleteAllChannelCacheForStore(store.ID)
	return store, nil
}

func (s *StoreService) Delete(id uint64) error {
	store, err := s.storeDAO.FindByID(id)
	if err != nil {
		return err
	}
	if store == nil {
		return common.ErrInvalidInput("门店不存在")
	}
	if err := s.storeDAO.Delete(id); err != nil {
		return err
	}
	_ = dao.DeleteAllChannelCacheForStore(id)
	return nil
}

func (s *StoreService) Get(id uint64) (*model.Store, error) {
	return s.storeDAO.FindByID(id)
}

func (s *StoreService) List(keyword string, status *int8, hideSystem bool, page, pageSize int) ([]model.Store, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return s.storeDAO.List(keyword, status, hideSystem, page, pageSize)
}
