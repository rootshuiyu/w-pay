package common

import (
	"regexp"
	"strings"
)

// 合规铁律：支付敏感信息绝不落地持久化，日志全部脱敏

var (
	openIDPattern  = regexp.MustCompile(`(openid[=:]\s*)([a-zA-Z0-9_-]{6,})`)
	authCodePattern = regexp.MustCompile(`(\d{16,24})`)
)

// MaskOpenID 脱敏 openid，仅保留前4后4
func MaskOpenID(openid string) string {
	if openid == "" {
		return ""
	}
	if len(openid) <= 8 {
		return "****"
	}
	return openid[:4] + "****" + openid[len(openid)-4:]
}

// MaskAuthCode 脱敏付款授权码，绝不持久化
func MaskAuthCode(code string) string {
	if code == "" {
		return ""
	}
	if len(code) <= 6 {
		return "******"
	}
	return code[:3] + "****" + code[len(code)-3:]
}

// MaskPrivateKey 脱敏私钥/证书内容
func MaskPrivateKey(key string) string {
	if key == "" {
		return ""
	}
	return "[PRIVATE_KEY_REDACTED]"
}

// MaskBankCard 脱敏银行卡号
func MaskBankCard(card string) string {
	if card == "" {
		return ""
	}
	if len(card) <= 8 {
		return "****"
	}
	return card[:4] + "****" + card[len(card)-4:]
}

// DesensitizeLog 对日志字符串做敏感信息脱敏
func DesensitizeLog(content string) string {
	if content == "" {
		return content
	}
	content = openIDPattern.ReplaceAllString(content, "${1}****")
	content = authCodePattern.ReplaceAllStringFunc(content, func(s string) string {
		if len(s) >= 16 {
			return MaskAuthCode(s)
		}
		return s
	})
	// 私钥片段
	if strings.Contains(content, "BEGIN") && strings.Contains(content, "PRIVATE KEY") {
		return "[SENSITIVE_CONTENT_REDACTED]"
	}
	return content
}
