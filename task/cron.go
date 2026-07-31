package task

import (
	"time"

	"wpay/common"
	"wpay/config"
	"wpay/dao"
	"wpay/service"

	"github.com/robfig/cron/v3"
)

// StartCron 启动定时任务
func StartCron() *cron.Cron {
	c := cron.New()

	orderSvc := service.NewOrderService()
	timeoutMin := config.Global.Order.TimeoutMinutes

	// 每5分钟关闭超时未支付订单
	_, _ = c.AddFunc("*/5 * * * *", func() {
		count, err := orderSvc.CloseTimeoutOrders(timeoutMin)
		if err != nil {
			common.Log.Error("close timeout orders failed: %v", err)
			return
		}
		if count > 0 {
			common.Log.Info("closed %d timeout orders", count)
		}
	})

	// 每天凌晨清理过期敏感日志（留存6个月）
	_, _ = c.AddFunc("0 3 * * *", func() {
		months := config.Global.Log.SensitiveRetentionMonths
		if months <= 0 {
			months = 6
		}
		before := time.Now().AddDate(0, -months, 0)
		count, err := dao.DeleteExpiredSensitiveLogs(before)
		if err != nil {
			common.Log.Error("clean sensitive logs failed: %v", err)
			return
		}
		common.Log.Info("cleaned %d expired sensitive logs before %s", count, before.Format("2006-01-02"))
	})

	// 每小时刷新门店缓存（兜底）
	_, _ = c.AddFunc("0 * * * *", func() {
		if err := dao.WarmupAllStoreCache(); err != nil {
			common.Log.Warn("warmup store cache failed: %v", err)
		}
	})

	// 每天凌晨清理过期渠道历史密钥（7天留存）
	_, _ = c.AddFunc("0 4 * * *", func() {
		count, err := dao.NewPayChannelHistoryDAO().DeleteExpired(time.Now())
		if err != nil {
			common.Log.Error("clean channel history failed: %v", err)
			return
		}
		if count > 0 {
			common.Log.Info("cleaned %d expired channel history records", count)
		}
	})

	// 每天 0:05 重置商户码日额度
	_, _ = c.AddFunc("5 0 * * *", func() {
		count, err := dao.NewPayChannelDAO().ResetAllDailyUsed()
		if err != nil {
			common.Log.Error("reset channel daily quota failed: %v", err)
			return
		}
		if count > 0 {
			common.Log.Info("reset daily quota for %d channels", count)
		}
	})

	c.Start()
	common.Log.Info("cron tasks started")
	return c
}
