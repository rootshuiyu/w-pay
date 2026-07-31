package common

import (
	"net"
	"strings"
)

// IPInAllowList 判断 clientIP 是否命中逗号/空格分隔的 IP/CIDR 列表。
// 列表为空视为未配置白名单，一律拒绝（代收平台对接必须显式配置 IP）。
func IPInAllowList(clientIP, allowList string) bool {
	raw := strings.TrimSpace(allowList)
	if raw == "" {
		return false
	}
	ip := net.ParseIP(strings.TrimSpace(clientIP))
	if ip == nil {
		return false
	}
	fields := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ';' || r == ' ' || r == '\n' || r == '\t'
	})
	for _, e := range fields {
		e = strings.TrimSpace(e)
		if e == "" {
			continue
		}
		if strings.Contains(e, "/") {
			_, n, err := net.ParseCIDR(e)
			if err == nil && n.Contains(ip) {
				return true
			}
			continue
		}
		if allowed := net.ParseIP(e); allowed != nil && allowed.Equal(ip) {
			return true
		}
	}
	return false
}
