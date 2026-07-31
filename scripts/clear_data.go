//go:build ignore

// 清空业务测试数据：go run scripts/clear_data.go
package main

import (
	"context"
	"fmt"
	"os"

	"wpay/config"
	"wpay/dao"
	"wpay/model"
)

func main() {
	if os.Getenv("APP_ENV") == "" {
		os.Setenv("APP_ENV", "dev")
	}
	cfg, err := config.Load(os.Getenv("APP_ENV"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}
	if err := dao.InitDB(&cfg.Database); err != nil {
		fmt.Fprintf(os.Stderr, "connect db: %v\n", err)
		os.Exit(1)
	}
	if err := dao.InitRedis(&cfg.Redis); err != nil {
		fmt.Println("warn: redis unavailable, skip cache flush")
	} else if err := dao.RDB.FlushDB(context.Background()).Err(); err != nil {
		fmt.Fprintf(os.Stderr, "redis flush: %v\n", err)
	} else {
		fmt.Println("redis: flushed")
	}

	tables := []interface{}{
		&model.Order{},
		&model.PayChannelHistory{},
		&model.PayChannel{},
		&model.PayPlatform{},
		&model.SensitiveLog{},
		&model.Store{},
	}
	for _, m := range tables {
		if err := dao.DB.Unscoped().Where("1 = 1").Delete(m).Error; err != nil {
			fmt.Fprintf(os.Stderr, "delete failed: %v\n", err)
			os.Exit(1)
		}
	}
	fmt.Println("cleared: orders, pay_channel_history, pay_channel, pay_platform, sensitive_logs, store")
	fmt.Println("kept: admin accounts, ip whitelist")
}
