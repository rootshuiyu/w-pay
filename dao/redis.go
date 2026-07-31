package dao

import (
	"context"
	"encoding/json"
	"math/rand"
	"time"

	"wpay/common"
	"wpay/config"
	"wpay/model"

	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client

func InitRedis(cfg *config.RedisConfig) error {
	RDB = redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return RDB.Ping(ctx).Err()
}

// ChannelCacheItem 单门店单渠道缓存（Key: store:channel:{store_id}:{pay_type}）
type ChannelCacheItem struct {
	Store   model.Store    `json:"store"`
	Channel model.PayChannel `json:"channel"`
}

// SetChannelCache 写入渠道缓存（带 TTL + 随机 jitter 防雪崩）
func SetChannelCache(storeID uint64, payType model.PayType, item *ChannelCacheItem) error {
	ctx := context.Background()
	bytes, err := json.Marshal(item)
	if err != nil {
		return err
	}
	ttl := channelCacheTTL()
	return RDB.Set(ctx, common.StoreChannelKey(storeID, int8(payType)), bytes, ttl).Err()
}

func channelCacheTTL() time.Duration {
	hours := 24
	jitterHours := 4
	if config.Global != nil {
		if config.Global.Cache.ChannelTTLHours > 0 {
			hours = config.Global.Cache.ChannelTTLHours
		}
		if config.Global.Cache.ChannelTTLJitterHours > 0 {
			jitterHours = config.Global.Cache.ChannelTTLJitterHours
		}
	}
	base := time.Duration(hours) * time.Hour
	if jitterHours > 0 {
		base += time.Duration(rand.Intn(jitterHours)) * time.Hour
	}
	return base
}

// GetChannelCache 读取渠道缓存，miss 返回 nil
func GetChannelCache(storeID uint64, payType model.PayType) (*ChannelCacheItem, error) {
	ctx := context.Background()
	val, err := RDB.Get(ctx, common.StoreChannelKey(storeID, int8(payType))).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var item ChannelCacheItem
	if err := json.Unmarshal(val, &item); err != nil {
		return nil, err
	}
	return &item, nil
}

// DeleteChannelCache 渠道热更新核心：后台修改后 Delete 对应 Key，下次下单自动回源 DB 并回填
func DeleteChannelCache(storeID uint64, payType model.PayType) error {
	ctx := context.Background()
	return RDB.Del(ctx, common.StoreChannelKey(storeID, int8(payType))).Err()
}

// DeleteAllChannelCacheForStore 门店停用/删除时清除该门店全部渠道缓存
func DeleteAllChannelCacheForStore(storeID uint64) error {
	ctx := context.Background()
	keys := []string{
		common.StoreChannelKey(storeID, int8(model.PayTypeWechat)),
		common.StoreChannelKey(storeID, int8(model.PayTypeAlipay)),
	}
	return RDB.Del(ctx, keys...).Err()
}

// LoadAndCacheChannel 缓存 miss 时从 DB 加载并写入 Redis
func LoadAndCacheChannel(storeID uint64, payType model.PayType) (*ChannelCacheItem, error) {
	var store model.Store
	if err := DB.First(&store, storeID).Error; err != nil {
		return nil, err
	}
	chDAO := NewPayChannelDAO()
	ch, err := chDAO.FindByStoreAndType(storeID, payType)
	if err != nil || ch == nil {
		return nil, err
	}
	item := &ChannelCacheItem{Store: store, Channel: *ch}
	_ = SetChannelCache(storeID, payType, item)
	return item, nil
}

// WarmupAllStoreCache 系统启动：仅预热「单渠道门店」到 Redis，多码门店走轮询池不缓存
func WarmupAllStoreCache() error {
	chDAO := NewPayChannelDAO()
	channels, err := chDAO.ListAllEnabled()
	if err != nil {
		return err
	}
	count := 0
	for _, ch := range channels {
		n, err := chDAO.CountEnabledByStoreAndType(ch.StoreID, ch.PayType)
		if err != nil || n != 1 {
			continue
		}
		var store model.Store
		if err := DB.First(&store, ch.StoreID).Error; err != nil {
			continue
		}
		if store.Status != model.StoreStatusNormal {
			continue
		}
		item := &ChannelCacheItem{Store: store, Channel: ch}
		if err := SetChannelCache(ch.StoreID, ch.PayType, item); err != nil {
			common.Log.Warn("warmup cache store=%d pay_type=%d failed: %v", ch.StoreID, ch.PayType, err)
			continue
		}
		count++
	}
	common.Log.Info("channel cache warmup done, keys=%d", count)
	return nil
}

func SetAdminToken(adminID uint64, token string, expire time.Duration) error {
	ctx := context.Background()
	return RDB.Set(ctx, common.AdminTokenKey(adminID), token, expire).Err()
}

func GetAdminToken(adminID uint64) (string, error) {
	ctx := context.Background()
	return RDB.Get(ctx, common.AdminTokenKey(adminID)).Result()
}

func DeleteAdminToken(adminID uint64) error {
	ctx := context.Background()
	return RDB.Del(ctx, common.AdminTokenKey(adminID)).Err()
}

func SetCallbackIdempotent(channel, transactionID string, expire time.Duration) (bool, error) {
	ctx := context.Background()
	key := common.CallbackIdempotentKey(channel, transactionID)
	return RDB.SetNX(ctx, key, "1", expire).Result()
}

func IncrRateLimit(ip string, window time.Duration) (int64, error) {
	ctx := context.Background()
	key := common.RateLimitKey(ip)
	pipe := RDB.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, window)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}
	return incr.Val(), nil
}

func DeleteExpiredSensitiveLogs(before time.Time) (int64, error) {
	result := DB.Unscoped().Where("created_at < ?", before).Delete(&model.SensitiveLog{})
	return result.RowsAffected, result.Error
}
