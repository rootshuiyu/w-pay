package middleware

import (
	"net"
	"sync"

	"wpay/common"

	"github.com/gin-gonic/gin"
)

// ScopeSnapshot 某作用域的白名单快照（线程安全只读）
type ScopeSnapshot struct {
	Enforced bool
	Entries  []string
	ips      []net.IP
	nets     []*net.IPNet
}

var (
	registryMu   sync.RWMutex
	registry     = map[string]*ScopeSnapshot{}
	ginEngineRef *gin.Engine
)

// BindEngine 供白名单热更新时同步可信反向代理
func BindEngine(e *gin.Engine) {
	ginEngineRef = e
}

// ReloadWhitelistRegistry 从策略与条目重建内存快照并应用可信代理
func ReloadWhitelistRegistry(policies map[string]bool, entriesByScope map[string][]string) {
	registryMu.Lock()
	defer registryMu.Unlock()

	registry = make(map[string]*ScopeSnapshot, len(entriesByScope))
	for scope, entries := range entriesByScope {
		enforced := policies[scope]
		snap := &ScopeSnapshot{
			Enforced: enforced,
			Entries:  append([]string(nil), entries...),
		}
		if enforced && len(entries) > 0 {
			snap.ips, snap.nets = parseIPEntries(scope, entries)
		}
		registry[scope] = snap
	}

	if ginEngineRef != nil {
		proxies := entriesByScope["trusted_proxy"]
		if policies["trusted_proxy"] && len(proxies) > 0 {
			_ = ginEngineRef.SetTrustedProxies(proxies)
		} else {
			_ = ginEngineRef.SetTrustedProxies(nil)
		}
	}
}

func getScopeSnapshot(scope string) *ScopeSnapshot {
	registryMu.RLock()
	defer registryMu.RUnlock()
	return registry[scope]
}

// ReapplyTrustedProxies 路由初始化后应用已加载的可信代理配置
func ReapplyTrustedProxies() {
	registryMu.RLock()
	defer registryMu.RUnlock()
	if ginEngineRef == nil {
		return
	}
	snap := registry["trusted_proxy"]
	if snap != nil && snap.Enforced && len(snap.Entries) > 0 {
		_ = ginEngineRef.SetTrustedProxies(snap.Entries)
		return
	}
	_ = ginEngineRef.SetTrustedProxies(nil)
}

// DynamicIPWhitelist 从内存快照读取规则，支持后台热更新
func DynamicIPWhitelist(scope string) gin.HandlerFunc {
	return func(c *gin.Context) {
		snap := getScopeSnapshot(scope)
		if snap == nil || !snap.Enforced {
			c.Next()
			return
		}
		if len(snap.Entries) == 0 || (len(snap.ips) == 0 && len(snap.nets) == 0) {
			commonForbidden(c)
			return
		}
		ip := net.ParseIP(c.ClientIP())
		if ip != nil && ipAllowed(ip, snap.ips, snap.nets) {
			c.Next()
			return
		}
		common.Log.Warn("%s ip rejected: %s %s", scope, c.ClientIP(), c.Request.URL.Path)
		common.Forbidden(c, "来源 IP 不在白名单内")
		c.Abort()
	}
}

func commonForbidden(c *gin.Context) {
	common.Forbidden(c, "来源 IP 不在白名单内")
	c.Abort()
}
