package common

import "fmt"

// Redis Key 设计（交付文档规范）
const (
	// 门店渠道配置缓存 Key: store:channel:{store_id}:{pay_type}
	KeyStoreChannelPrefix = "store:channel:"
	KeyAdminTokenPrefix   = "wpay:admin:token:"
	KeyCallbackIdempotent = "wpay:callback:idempotent:"
	KeyRateLimitPrefix    = "wpay:ratelimit:"
	KeyChannelRotatePrefix = "channel:rotate:"
)

func ChannelRotateKey(payType int8, platformID, storeID uint64) string {
	return fmt.Sprintf("%s%d:%d:%d", KeyChannelRotatePrefix, payType, platformID, storeID)
}

// StoreChannelKey 单门店单渠道缓存 Key，后台修改后主动 Delete 实现热更新
func StoreChannelKey(storeID uint64, payType int8) string {
	return fmt.Sprintf("%s%d:%d", KeyStoreChannelPrefix, storeID, payType)
}

func AdminTokenKey(adminID uint64) string {
	return fmt.Sprintf("%s%d", KeyAdminTokenPrefix, adminID)
}

func CallbackIdempotentKey(channel, transactionID string) string {
	return fmt.Sprintf("%s%s:%s", KeyCallbackIdempotent, channel, transactionID)
}

func RateLimitKey(ip string) string {
	return fmt.Sprintf("%s%s", KeyRateLimitPrefix, ip)
}
