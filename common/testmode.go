package common

import "os"

// TestMode E2E 测试模式：跳过真实微信/支付宝 SDK 调用，返回 mock 付款码
func TestMode() bool {
	return os.Getenv("WPAY_TEST_MODE") == "1"
}
