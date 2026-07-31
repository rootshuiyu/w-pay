package middleware

import (
	"net"
	"strings"

	"wpay/common"

	"github.com/gin-gonic/gin"
)

// IPWhitelist 按 IP/CIDR 限制访问来源。
// 合规：后台管理接口限内网访问，支付回调仅放行微信/支付宝官方 IP。
// entries 为空表示不限制（dev 环境默认），生产通过 config/prod.yaml 或环境变量注入。
// 平台对接白名单配置了但为空时，按 fail-closed 拒绝访问，必须显式配置可信IP。
func IPWhitelist(scope string, entries []string, allowEmpty bool) gin.HandlerFunc {
	if len(entries) == 0 {
		if allowEmpty {
			common.Log.Info("%s ip whitelist empty, allow all", scope)
			return func(c *gin.Context) { c.Next() }
		}
		common.Log.Error("%s ip whitelist configured but empty, all requests will be rejected", scope)
		return func(c *gin.Context) {
			common.Log.Warn("%s ip rejected: empty whitelist %s %s", scope, c.ClientIP(), c.Request.URL.Path)
			common.Forbidden(c, "来源 IP 不在白名单内")
			c.Abort()
		}
	}

	ips, nets := parseIPEntries(scope, entries)
	if len(ips) == 0 && len(nets) == 0 {
		// 配置了白名单但无一条可用：按 fail-closed 全部拒绝，避免误以为已生效
		common.Log.Error("%s ip whitelist configured but no valid entry, all requests will be rejected", scope)
		return func(c *gin.Context) {
			common.Log.Warn("%s ip rejected: no valid entries %s %s", scope, c.ClientIP(), c.Request.URL.Path)
			common.Forbidden(c, "来源 IP 不在白名单内")
			c.Abort()
		}
	}

	common.Log.Info("%s ip whitelist enabled: %d ip, %d cidr", scope, len(ips), len(nets))

	return func(c *gin.Context) {
		ip := net.ParseIP(c.ClientIP())
		if ip != nil && ipAllowed(ip, ips, nets) {
			c.Next()
			return
		}
		common.Log.Warn("%s ip rejected: %s %s", scope, c.ClientIP(), c.Request.URL.Path)
		common.Forbidden(c, "来源 IP 不在白名单内")
		c.Abort()
	}
}

func ipAllowed(ip net.IP, ips []net.IP, nets []*net.IPNet) bool {
	for _, allowed := range ips {
		if allowed.Equal(ip) {
			return true
		}
	}
	for _, n := range nets {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

// ParseIPList 解析逗号/空格分隔的 IP、CIDR 列表
func ParseIPList(raw string) []string {
	fields := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ';' || r == ' ' || r == '\n' || r == '\t'
	})
	entries := make([]string, 0, len(fields))
	for _, f := range fields {
		if f != "" {
			entries = append(entries, f)
		}
	}
	return entries
}

func parseIPEntries(scope string, entries []string) ([]net.IP, []*net.IPNet) {
	var ips []net.IP
	var nets []*net.IPNet
	for _, e := range entries {
		if strings.Contains(e, "/") {
			_, n, err := net.ParseCIDR(e)
			if err != nil {
				common.Log.Warn("%s ip whitelist: invalid cidr %q, ignored", scope, e)
				continue
			}
			nets = append(nets, n)
			continue
		}
		ip := net.ParseIP(e)
		if ip == nil {
			common.Log.Warn("%s ip whitelist: invalid ip %q, ignored", scope, e)
			continue
		}
		ips = append(ips, ip)
	}
	return ips, nets
}
