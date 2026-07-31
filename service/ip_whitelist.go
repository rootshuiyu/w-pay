package service

import (
	"net"
	"strings"

	"wpay/common"
	"wpay/common/middleware"
	"wpay/config"
	"wpay/dao"
	"wpay/model"
)

type IPWhitelistService struct {
	dao *dao.IPWhitelistDAO
}

func NewIPWhitelistService() *IPWhitelistService {
	return &IPWhitelistService{dao: dao.NewIPWhitelistDAO()}
}

type IPWhitelistPolicyRequest struct {
	Scope   string `json:"scope" binding:"required"`
	Enabled int8   `json:"enabled"`
	Remark  string `json:"remark"`
}

type IPWhitelistAddRequest struct {
	Scope  string `json:"scope" binding:"required"`
	CIDR   string `json:"cidr" binding:"required"`
	Remark string `json:"remark"`
}

type IPWhitelistEditRequest struct {
	ID     common.FlexUint64 `json:"id" binding:"required"`
	Remark string            `json:"remark"`
}

type IPWhitelistStatusRequest struct {
	ID     common.FlexUint64 `json:"id" binding:"required"`
	Status int8              `json:"status"`
}

type IPWhitelistDeleteRequest struct {
	ID common.FlexUint64 `json:"id"`
}

type IPWhitelistOverview struct {
	Policies []model.IPWhitelistPolicy `json:"policies"`
	Entries  []model.IPWhitelistEntry  `json:"entries"`
}

// InitIPWhitelist 启动时初始化策略、从配置种子导入，并加载到内存
func InitIPWhitelist() error {
	svc := NewIPWhitelistService()
	if err := svc.dao.EnsurePolicies(); err != nil {
		return err
	}
	if err := svc.seedFromConfigIfEmpty(); err != nil {
		return err
	}
	return svc.ReloadRegistry()
}

func (s *IPWhitelistService) seedFromConfigIfEmpty() error {
	if config.Global == nil {
		return nil
	}
	seeds := map[string]string{
		string(model.IPWhitelistScopeAdmin):        config.Global.Security.AdminIPWhitelist,
		string(model.IPWhitelistScopeCallback):     config.Global.Security.CallbackIPWhitelist,
		string(model.IPWhitelistScopePay):          config.Global.Security.PayIPWhitelist,
		string(model.IPWhitelistScopeTrustedProxy): config.Global.Security.TrustedProxies,
	}
	for scope, raw := range seeds {
		count, err := s.dao.CountEntries(scope)
		if err != nil {
			return err
		}
		if count > 0 || strings.TrimSpace(raw) == "" {
			continue
		}
		for _, cidr := range middleware.ParseIPList(raw) {
			if err := validateCIDR(cidr); err != nil {
				continue
			}
			entry := &model.IPWhitelistEntry{
				Scope:  scope,
				CIDR:   cidr,
				Remark: "从环境配置导入",
				Status: 1,
			}
			if err := s.dao.CreateEntry(entry); err != nil {
				return err
			}
		}
		if n, _ := s.dao.CountEntries(scope); n > 0 {
			enabled := int8(0)
			if scope == string(model.IPWhitelistScopeAdmin) && config.IsProd() {
				enabled = 1
			}
			_ = s.dao.UpdatePolicy(scope, enabled, "环境配置种子导入")
		}
	}
	return nil
}

func (s *IPWhitelistService) ReloadRegistry() error {
	policies, err := s.dao.ListPolicies()
	if err != nil {
		return err
	}
	policyMap := make(map[string]bool, len(policies))
	for _, p := range policies {
		policyMap[p.Scope] = p.Enabled == 1
	}

	scopes := []string{
		string(model.IPWhitelistScopeAdmin),
		string(model.IPWhitelistScopeCallback),
		string(model.IPWhitelistScopePay),
		string(model.IPWhitelistScopeTrustedProxy),
	}
	entriesByScope := make(map[string][]string, len(scopes))
	for _, scope := range scopes {
		cidrs, err := s.dao.ListActiveCIDRs(scope)
		if err != nil {
			return err
		}
		entriesByScope[scope] = cidrs
	}
	middleware.ReloadWhitelistRegistry(policyMap, entriesByScope)
	return nil
}

func (s *IPWhitelistService) Overview(scope string) (*IPWhitelistOverview, error) {
	if scope != "" && !model.IPWhitelistScope(scope).Valid() {
		return nil, common.ErrInvalidInput("无效的作用域")
	}
	policies, err := s.dao.ListPolicies()
	if err != nil {
		return nil, err
	}
	entries, err := s.dao.ListEntries(scope, nil)
	if err != nil {
		return nil, err
	}
	return &IPWhitelistOverview{Policies: policies, Entries: entries}, nil
}

func (s *IPWhitelistService) UpdatePolicy(req IPWhitelistPolicyRequest) error {
	if !model.IPWhitelistScope(req.Scope).Valid() {
		return common.ErrInvalidInput("无效的作用域")
	}
	if req.Enabled != 0 && req.Enabled != 1 {
		return common.ErrInvalidInput("enabled 仅支持 0 或 1")
	}
	if err := common.SanitizeString(req.Remark); err != nil {
		return err
	}
	if err := s.dao.UpdatePolicy(req.Scope, req.Enabled, req.Remark); err != nil {
		return err
	}
	return s.ReloadRegistry()
}

func (s *IPWhitelistService) Add(req IPWhitelistAddRequest) (*model.IPWhitelistEntry, error) {
	if !model.IPWhitelistScope(req.Scope).Valid() {
		return nil, common.ErrInvalidInput("无效的作用域")
	}
	cidr := strings.TrimSpace(req.CIDR)
	if err := validateCIDR(cidr); err != nil {
		return nil, err
	}
	if err := common.SanitizeString(req.Remark); err != nil {
		return nil, err
	}
	if exist, _ := s.dao.FindEntry(req.Scope, cidr); exist != nil {
		return nil, common.ErrInvalidInput("该 IP/CIDR 已存在")
	}
	entry := &model.IPWhitelistEntry{
		Scope:  req.Scope,
		CIDR:   cidr,
		Remark: req.Remark,
		Status: 1,
	}
	if err := s.dao.CreateEntry(entry); err != nil {
		return nil, err
	}
	if err := s.ReloadRegistry(); err != nil {
		return nil, err
	}
	return entry, nil
}

func (s *IPWhitelistService) Edit(req IPWhitelistEditRequest) (*model.IPWhitelistEntry, error) {
	entry, err := s.dao.FindEntryByID(req.ID.Uint64())
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, common.ErrInvalidInput("条目不存在")
	}
	if err := common.SanitizeString(req.Remark); err != nil {
		return nil, err
	}
	entry.Remark = req.Remark
	if err := s.dao.UpdateEntry(entry); err != nil {
		return nil, err
	}
	return entry, nil
}

func (s *IPWhitelistService) UpdateStatus(req IPWhitelistStatusRequest) (*model.IPWhitelistEntry, error) {
	entry, err := s.dao.FindEntryByID(req.ID.Uint64())
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, common.ErrInvalidInput("条目不存在")
	}
	if req.Status != 0 && req.Status != 1 {
		return nil, common.ErrInvalidInput("status 仅支持 0 或 1")
	}
	entry.Status = req.Status
	if err := s.dao.UpdateEntry(entry); err != nil {
		return nil, err
	}
	if err := s.ReloadRegistry(); err != nil {
		return nil, err
	}
	return entry, nil
}

func (s *IPWhitelistService) Delete(id uint64) error {
	entry, err := s.dao.FindEntryByID(id)
	if err != nil {
		return err
	}
	if entry == nil {
		return common.ErrInvalidInput("条目不存在")
	}
	if err := s.dao.DeleteEntry(id); err != nil {
		return err
	}
	return s.ReloadRegistry()
}

func validateCIDR(raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return common.ErrInvalidInput("IP/CIDR 不能为空")
	}
	if strings.Contains(raw, "/") {
		if _, _, err := net.ParseCIDR(raw); err != nil {
			return common.ErrInvalidInput("无效的 CIDR 格式")
		}
		return nil
	}
	if ip := net.ParseIP(raw); ip == nil {
		return common.ErrInvalidInput("无效的 IP 格式")
	}
	return nil
}
