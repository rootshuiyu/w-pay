package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	"wpay/common"
	"wpay/dao"
	"wpay/model"
)

type PayPlatformService struct {
	platformDAO *dao.PayPlatformDAO
	channelDAO  *dao.PayChannelDAO
	storeSvc    *StoreService
	channelSvc  *PayChannelService
}

func NewPayPlatformService() *PayPlatformService {
	return &PayPlatformService{
		platformDAO: dao.NewPayPlatformDAO(),
		channelDAO:  dao.NewPayChannelDAO(),
		storeSvc:    NewStoreService(),
		channelSvc:  NewPayChannelService(),
	}
}

type PlatformCreateRequest struct {
	PlatformCode string `json:"platform_code"`
	PlatformName string `json:"platform_name" binding:"required"`
	AppKey       string `json:"app_key"`
	AllowedIPs   string `json:"allowed_ips"`
	Remark       string `json:"remark"`
}

type PlatformEditRequest struct {
	ID           common.FlexUint64 `json:"id" binding:"required"`
	PlatformCode string            `json:"platform_code"`
	PlatformName string            `json:"platform_name"`
	AllowedIPs   *string           `json:"allowed_ips"`
	Remark       string            `json:"remark"`
}

type PlatformStatusRequest struct {
	ID     common.FlexUint64 `json:"id" binding:"required"`
	Status int8              `json:"status"`
}

type PlatformDeleteRequest struct {
	ID common.FlexUint64 `json:"id"`
}

type PlatformSetChannelsRequest struct {
	PlatformID common.FlexUint64   `json:"platform_id" binding:"required"`
	ChannelIDs []common.FlexUint64 `json:"channel_ids"`
}

func GenerateAppKey() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "pk_" + hex.EncodeToString(b), nil
}

func GeneratePlatformCode() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "pc_" + hex.EncodeToString(b), nil
}

func (s *PayPlatformService) Create(req PlatformCreateRequest) (*model.PayPlatform, error) {
	for _, field := range []string{req.PlatformCode, req.PlatformName, req.AppKey, req.AllowedIPs, req.Remark} {
		if err := common.SanitizeString(field); err != nil {
			return nil, err
		}
	}
	if err := common.ValidateLength(req.PlatformName, 1, 128, "平台名称"); err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.AllowedIPs) == "" {
		return nil, common.ErrInvalidInput("请填写对接 IP 白名单")
	}
	platformCode := strings.TrimSpace(req.PlatformCode)
	if platformCode == "" {
		var err error
		platformCode, err = GeneratePlatformCode()
		if err != nil {
			return nil, err
		}
	} else {
		exist, err := s.platformDAO.FindByCode(platformCode)
		if err != nil {
			return nil, err
		}
		if exist != nil {
			return nil, common.ErrInvalidInput("平台编号已存在")
		}
	}
	appKey := strings.TrimSpace(req.AppKey)
	if appKey == "" {
		var err error
		appKey, err = GenerateAppKey()
		if err != nil {
			return nil, err
		}
	}
	exist, err := s.platformDAO.FindByAppKey(appKey)
	if err != nil {
		return nil, err
	}
	if exist != nil {
		return nil, common.ErrInvalidInput("app_key 已存在")
	}

	id, err := common.GenerateID()
	if err != nil {
		return nil, err
	}
	p := &model.PayPlatform{
		BaseModel:    model.BaseModel{ID: id},
		PlatformCode: platformCode,
		PlatformName: req.PlatformName,
		AppKey:       appKey,
		AllowedIPs:   strings.TrimSpace(req.AllowedIPs),
		Status:       model.PayPlatformStatusNormal,
		Remark:       req.Remark,
	}
	if err := s.platformDAO.Create(p); err != nil {
		return nil, common.WrapDAOError(err)
	}
	return p, nil
}

func (s *PayPlatformService) Update(id uint64, req PlatformEditRequest) (*model.PayPlatform, error) {
	p, err := s.platformDAO.FindByID(id)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, common.ErrInvalidInput("平台不存在")
	}
	for _, field := range []string{req.PlatformCode, req.PlatformName, req.Remark} {
		if err := common.SanitizeString(field); err != nil {
			return nil, err
		}
	}
	if req.AllowedIPs != nil {
		if err := common.SanitizeString(*req.AllowedIPs); err != nil {
			return nil, err
		}
	}
	if req.PlatformName != "" {
		p.PlatformName = req.PlatformName
	}
	if req.PlatformCode != "" && req.PlatformCode != p.PlatformCode {
		exist, err := s.platformDAO.FindByCode(req.PlatformCode)
		if err != nil {
			return nil, err
		}
		if exist != nil && exist.ID != id {
			return nil, common.ErrInvalidInput("平台编号已存在")
		}
		p.PlatformCode = req.PlatformCode
	}
	if req.AllowedIPs != nil {
		ips := strings.TrimSpace(*req.AllowedIPs)
		if ips == "" {
			return nil, common.ErrInvalidInput("对接 IP 白名单不能为空")
		}
		p.AllowedIPs = ips
	}
	if req.Remark != "" {
		p.Remark = req.Remark
	}
	if err := s.platformDAO.Update(p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *PayPlatformService) UpdateStatus(id uint64, status int8) (*model.PayPlatform, error) {
	p, err := s.platformDAO.FindByID(id)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, common.ErrInvalidInput("平台不存在")
	}
	if status == model.PayPlatformStatusNormal && strings.TrimSpace(p.AllowedIPs) == "" {
		return nil, common.ErrInvalidInput("启用前请先配置对接 IP 白名单")
	}
	p.Status = status
	if err := s.platformDAO.Update(p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *PayPlatformService) Delete(id uint64) error {
	p, err := s.platformDAO.FindByID(id)
	if err != nil {
		return err
	}
	if p == nil {
		return common.ErrInvalidInput("平台不存在")
	}
	if err := s.platformDAO.SetChannels(id, nil); err != nil {
		return err
	}
	return s.platformDAO.Delete(id)
}

func (s *PayPlatformService) List(keyword string, status *int8, page, pageSize int) ([]map[string]interface{}, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	list, total, err := s.platformDAO.List(keyword, status, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	out := make([]map[string]interface{}, 0, len(list))
	for i := range list {
		cnt, _ := s.platformDAO.CountChannels(list[i].ID)
		out = append(out, map[string]interface{}{
			"id":            list[i].ID,
			"platform_code": list[i].PlatformCode,
			"platform_name": list[i].PlatformName,
			"app_key":       list[i].AppKey,
			"allowed_ips":   list[i].AllowedIPs,
			"status":        list[i].Status,
			"remark":        list[i].Remark,
			"channel_count": cnt,
			"created_at":    list[i].CreatedAt,
			"updated_at":    list[i].UpdatedAt,
		})
	}
	return out, total, nil
}

func (s *PayPlatformService) ListChannels(platformID uint64) ([]map[string]interface{}, error) {
	list, err := s.platformDAO.ListChannels(platformID)
	if err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, 0, len(list))
	for i := range list {
		out = append(out, MaskChannel(&list[i]))
	}
	return out, nil
}

func (s *PayPlatformService) ListAvailableChannels(unassignedOnly bool) ([]map[string]interface{}, error) {
	list, err := s.platformDAO.ListAllChannelsBrief(unassignedOnly)
	if err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, 0, len(list))
	for i := range list {
		out = append(out, MaskChannel(&list[i]))
	}
	return out, nil
}

type PlatformBindChannelsRequest struct {
	PlatformID common.FlexUint64   `json:"platform_id" binding:"required"`
	ChannelIDs []common.FlexUint64 `json:"channel_ids" binding:"required,min=1"`
}

type PlatformUnbindChannelRequest struct {
	PlatformID common.FlexUint64 `json:"platform_id" binding:"required"`
	ChannelID  common.FlexUint64 `json:"channel_id" binding:"required"`
}

// BindChannels 从公共码池追加绑定商户码（不影响已绑定码）
func (s *PayPlatformService) BindChannels(platformID uint64, channelIDs []uint64) error {
	p, err := s.platformDAO.FindByID(platformID)
	if err != nil || p == nil {
		return common.ErrInvalidInput("平台不存在")
	}
	for _, cid := range channelIDs {
		ch, err := s.channelDAO.FindByID(cid)
		if err != nil {
			return err
		}
		if ch == nil {
			return common.ErrInvalidInput("商户码不存在")
		}
		if ch.PlatformID != 0 && ch.PlatformID != platformID {
			return common.ErrInvalidInput("商户码已绑定其他平台：" + ch.MchNo)
		}
		if err := s.platformDAO.AppendChannel(platformID, cid); err != nil {
			return err
		}
		_ = dao.DeleteChannelCache(ch.StoreID, ch.PayType)
	}
	return nil
}

// UnbindChannel 解绑单个商户码回公共码池
func (s *PayPlatformService) UnbindChannel(platformID, channelID uint64) error {
	ch, err := s.channelDAO.FindByID(channelID)
	if err != nil {
		return err
	}
	if ch == nil {
		return common.ErrInvalidInput("商户码不存在")
	}
	if ch.PlatformID != platformID {
		return common.ErrInvalidInput("该商户码不属于此平台")
	}
	if err := s.platformDAO.UnbindChannel(channelID); err != nil {
		return err
	}
	_ = dao.DeleteChannelCache(ch.StoreID, ch.PayType)
	return nil
}

func (s *PayPlatformService) SetChannels(platformID uint64, channelIDs []uint64) error {
	p, err := s.platformDAO.FindByID(platformID)
	if err != nil {
		return err
	}
	if p == nil {
		return common.ErrInvalidInput("平台不存在")
	}
	for _, cid := range channelIDs {
		ch, err := s.channelDAO.FindByID(cid)
		if err != nil {
			return err
		}
		if ch == nil {
			return common.ErrInvalidInput("商户码不存在")
		}
	}
	if err := s.platformDAO.SetChannels(platformID, channelIDs); err != nil {
		return err
	}
	for _, cid := range channelIDs {
		ch, _ := s.channelDAO.FindByID(cid)
		if ch != nil {
			_ = dao.DeleteChannelCache(ch.StoreID, ch.PayType)
		}
	}
	return nil
}

func (s *PayPlatformService) RegenerateAppKey(id uint64) (*model.PayPlatform, error) {
	p, err := s.platformDAO.FindByID(id)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, common.ErrInvalidInput("平台不存在")
	}
	key, err := GenerateAppKey()
	if err != nil {
		return nil, err
	}
	p.AppKey = key
	if err := s.platformDAO.Update(p); err != nil {
		return nil, err
	}
	return p, nil
}

// Resolve 识别对接平台：app_key 优先，否则 IP 匹配 allowed_ips；均未命中返回 nil（走未分配码池）。
// 平台必须配置非空 allowed_ips，空白名单一律拒绝。
func (s *PayPlatformService) Resolve(appKey, clientIP string) (*model.PayPlatform, error) {
	appKey = strings.TrimSpace(appKey)
	if appKey != "" {
		p, err := s.platformDAO.FindByAppKey(appKey)
		if err != nil {
			return nil, err
		}
		if p == nil {
			return nil, common.ErrInvalidInput("无效的 app_key")
		}
		if p.Status != model.PayPlatformStatusNormal {
			return nil, common.ErrInvalidInput("代收平台已停用")
		}
		if strings.TrimSpace(p.AllowedIPs) == "" {
			return nil, common.ErrInvalidInput("平台未配置对接 IP 白名单")
		}
		if !common.IPInAllowList(clientIP, p.AllowedIPs) {
			return nil, common.ErrInvalidInput("来源 IP 不在平台白名单内")
		}
		return p, nil
	}

	list, err := s.platformDAO.ListActive()
	if err != nil {
		return nil, err
	}
	var matched *model.PayPlatform
	for i := range list {
		p := &list[i]
		if strings.TrimSpace(p.AllowedIPs) == "" {
			continue
		}
		if !common.IPInAllowList(clientIP, p.AllowedIPs) {
			continue
		}
		if matched != nil {
			return nil, common.ErrInvalidInput("来源 IP 匹配多个平台，请使用 app_key")
		}
		matched = p
	}
	return matched, nil
}

func (s *PayPlatformService) PoolStats(platformID uint64, payType *model.PayType) ([]model.PayChannel, error) {
	pid := platformID
	return NewChannelPoolService().PoolStatsForPlatform(payType, &pid)
}

type PlatformQuickSetupChannel struct {
	PayType        model.PayType `json:"pay_type"`
	MchNo          string        `json:"mch_no" binding:"required"`
	PoolEnabled    *int8         `json:"pool_enabled"`
	DailyLimitFen  int64         `json:"daily_limit_fen"`
	SingleLimitFen int64         `json:"single_limit_fen"`
	MchKey         string        `json:"mch_key"`
	AppID          string        `json:"app_id"`
	SerialNo       string        `json:"serial_no"`
	PrivateKey     string        `json:"private_key"`
	PublicKey      string        `json:"public_key"`
}

type PlatformQuickSetupRequest struct {
	PlatformName string                      `json:"platform_name" binding:"required"`
	AllowedIPs   string                      `json:"allowed_ips"`
	Remark       string                      `json:"remark"`
	Channels     []PlatformQuickSetupChannel `json:"channels" binding:"required,min=1"`
}

type PlatformAddChannelRequest struct {
	PlatformID     common.FlexUint64 `json:"platform_id" binding:"required"`
	PayType        model.PayType     `json:"pay_type"`
	MchNo          string            `json:"mch_no" binding:"required"`
	PoolEnabled    *int8             `json:"pool_enabled"`
	DailyLimitFen  int64             `json:"daily_limit_fen"`
	SingleLimitFen int64             `json:"single_limit_fen"`
	MchKey         string            `json:"mch_key"`
	AppID          string            `json:"app_id"`
	SerialNo       string            `json:"serial_no"`
	PrivateKey     string            `json:"private_key"`
	PublicKey      string            `json:"public_key"`
	Remark         string            `json:"remark"`
}

// QuickSetup 一键接入：创建平台 + 门店 + 商户码并绑定
func (s *PayPlatformService) QuickSetup(req PlatformQuickSetupRequest) (map[string]interface{}, error) {
	if len(req.Channels) == 0 {
		return nil, common.ErrInvalidInput("请至少添加一个商户码")
	}
	seen := make(map[string]bool)
	for _, spec := range req.Channels {
		mch := strings.TrimSpace(spec.MchNo)
		if mch == "" {
			return nil, common.ErrInvalidInput("商户号不能为空")
		}
		key := fmt.Sprintf("%d:%s", spec.PayType, mch)
		if seen[key] {
			return nil, common.ErrInvalidInput("同一平台内商户号不能重复：" + mch)
		}
		seen[key] = true
	}

	p, err := s.Create(PlatformCreateRequest{
		PlatformName: req.PlatformName,
		AllowedIPs:   req.AllowedIPs,
		Remark:       req.Remark,
	})
	if err != nil {
		return nil, err
	}

	var storeID uint64
	var createdChIDs []uint64
	defer func() {
		if err == nil {
			return
		}
		for _, id := range createdChIDs {
			_ = s.channelDAO.Delete(id)
		}
		if storeID > 0 {
			_ = s.storeSvc.Delete(storeID)
		}
		_ = s.platformDAO.Delete(p.ID)
	}()

	store, err := s.storeSvc.Create(StoreCreateRequest{
		StoreName:  req.PlatformName + "·收款主体",
		TaxSubject: req.PlatformName,
		Remark:     "代收平台自动创建",
	})
	if err != nil {
		return nil, err
	}
	storeID = store.ID

	masked := make([]map[string]interface{}, 0, len(req.Channels))
	for _, spec := range req.Channels {
		var ch *model.PayChannel
		ch, err = s.createPlatformChannel(p.ID, store.ID, spec)
		if err != nil {
			err = common.WrapDAOError(err)
			return nil, err
		}
		createdChIDs = append(createdChIDs, ch.ID)
		masked = append(masked, MaskChannel(ch))
	}
	err = nil
	return map[string]interface{}{
		"platform": p,
		"store_id": store.ID,
		"channels": masked,
		"app_key":  p.AppKey,
	}, nil
}

func (s *PayPlatformService) AddChannel(req PlatformAddChannelRequest) (map[string]interface{}, error) {
	platformID := req.PlatformID.Uint64()
	p, err := s.platformDAO.FindByID(platformID)
	if err != nil || p == nil {
		return nil, common.ErrInvalidInput("平台不存在")
	}
	storeID, err := s.resolvePlatformStoreID(platformID, p.PlatformName)
	if err != nil {
		return nil, err
	}
	spec := PlatformQuickSetupChannel{
		PayType:        req.PayType,
		MchNo:          req.MchNo,
		PoolEnabled:    req.PoolEnabled,
		DailyLimitFen:  req.DailyLimitFen,
		SingleLimitFen: req.SingleLimitFen,
		MchKey:         req.MchKey,
		AppID:          req.AppID,
		SerialNo:       req.SerialNo,
		PrivateKey:     req.PrivateKey,
		PublicKey:      req.PublicKey,
	}
	if !spec.PayType.Valid() {
		spec.PayType = model.PayTypeWechat
	}
	ch, err := s.createPlatformChannel(platformID, storeID, spec)
	if err != nil {
		return nil, err
	}
	if req.Remark != "" {
		ch.Remark = req.Remark
		_ = s.channelDAO.Update(ch)
	}
	return map[string]interface{}{"channel": MaskChannel(ch)}, nil
}

func (s *PayPlatformService) resolvePlatformStoreID(platformID uint64, platformName string) (uint64, error) {
	list, err := s.platformDAO.ListChannels(platformID)
	if err != nil {
		return 0, err
	}
	if len(list) > 0 {
		return list[0].StoreID, nil
	}
	store, err := s.storeSvc.Create(StoreCreateRequest{
		StoreName:  platformName + "·收款主体",
		TaxSubject: platformName,
		Remark:     "代收平台自动创建",
	})
	if err != nil {
		return 0, err
	}
	return store.ID, nil
}

func (s *PayPlatformService) createPlatformChannel(platformID, storeID uint64, spec PlatformQuickSetupChannel) (*model.PayChannel, error) {
	poolEnabled := int8(1)
	if spec.PoolEnabled != nil {
		poolEnabled = *spec.PoolEnabled
	}
	if !spec.PayType.Valid() {
		spec.PayType = model.PayTypeWechat
	}
	ch, err := s.channelSvc.Create(ChannelCreateRequest{
		StoreID:        common.FlexUint64(storeID),
		PayType:        spec.PayType,
		PoolEnabled:    &poolEnabled,
		DailyLimitFen:  spec.DailyLimitFen,
		SingleLimitFen: spec.SingleLimitFen,
		MchNo:          spec.MchNo,
		MchKey:         spec.MchKey,
		AppID:          spec.AppID,
		SerialNo:       spec.SerialNo,
		PrivateKey:     spec.PrivateKey,
		PublicKey:      spec.PublicKey,
	})
	if err != nil {
		return nil, err
	}
	if err := s.platformDAO.AppendChannel(platformID, ch.ID); err != nil {
		return nil, err
	}
	ch.PlatformID = platformID
	return ch, nil
}

func (s *PayPlatformService) GetDetail(platformID uint64) (map[string]interface{}, error) {
	p, err := s.platformDAO.FindByID(platformID)
	if err != nil || p == nil {
		return nil, common.ErrInvalidInput("平台不存在")
	}
	channels, err := s.ListChannels(platformID)
	if err != nil {
		return nil, err
	}
	cnt, _ := s.platformDAO.CountChannels(platformID)
	return map[string]interface{}{
		"id":            p.ID,
		"platform_name": p.PlatformName,
		"platform_code": p.PlatformCode,
		"app_key":       p.AppKey,
		"allowed_ips":   p.AllowedIPs,
		"status":        p.Status,
		"remark":        p.Remark,
		"channel_count": cnt,
		"channels":      channels,
		"created_at":    p.CreatedAt,
	}, nil
}
