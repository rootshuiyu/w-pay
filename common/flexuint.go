package common

import (
	"fmt"
	"strconv"
	"strings"
)

// FlexUint64 兼容 JSON 数字与字符串两种输入的 uint64。
// 雪花 ID 超出 JS Number 安全整数范围（2^53），前端必须以字符串传递；
// 收银机等旧客户端仍可能传数字，此类型两者都接受，输出统一为字符串。
type FlexUint64 uint64

func (f *FlexUint64) UnmarshalJSON(b []byte) error {
	s := strings.Trim(strings.TrimSpace(string(b)), `"`)
	if s == "" || s == "null" {
		*f = 0
		return nil
	}
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid uint64: %s", s)
	}
	*f = FlexUint64(v)
	return nil
}

func (f FlexUint64) MarshalJSON() ([]byte, error) {
	return []byte(`"` + strconv.FormatUint(uint64(f), 10) + `"`), nil
}

func (f FlexUint64) Uint64() uint64 { return uint64(f) }
