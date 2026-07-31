package common

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

var (
	sqlInjectionPattern = regexp.MustCompile(`(?i)(union\s+select|insert\s+into|delete\s+from|drop\s+table|update\s+\w+\s+set|--|;|\/\*)`)
	xssPattern          = regexp.MustCompile(`(?i)(<script|javascript:|onerror=|onload=|<iframe)`)
)

// SanitizeString 入参统一校验：防 SQL 注入、XSS
func SanitizeString(input string) error {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil
	}
	if sqlInjectionPattern.MatchString(input) {
		return ErrInvalidInput("检测到非法字符")
	}
	if xssPattern.MatchString(input) {
		return ErrInvalidInput("检测到非法脚本内容")
	}
	return nil
}

// ValidateLength 校验字符串长度
func ValidateLength(input string, min, max int, fieldName string) error {
	length := utf8.RuneCountInString(input)
	if length < min || length > max {
		return ErrInvalidInput(fieldName + "长度不合法")
	}
	return nil
}

// BizError 业务错误
type BizError struct {
	Msg string
}

func (e *BizError) Error() string {
	return e.Msg
}

func ErrInvalidInput(msg string) error {
	return &BizError{Msg: msg}
}

func IsBizError(err error) (*BizError, bool) {
	if err == nil {
		return nil, false
	}
	be, ok := err.(*BizError)
	return be, ok
}

// WrapDAOError 将数据库错误转为可读业务提示
func WrapDAOError(err error) error {
	if err == nil {
		return nil
	}
	if be, ok := IsBizError(err); ok {
		return be
	}
	msg := err.Error()
	if strings.Contains(msg, "duplicate key") || strings.Contains(msg, "23505") || strings.Contains(msg, "UNIQUE") {
		return ErrInvalidInput("记录已存在，请检查是否重复")
	}
	return ErrInvalidInput("保存失败：" + msg)
}
