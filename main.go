package main

import (
	"fmt"
	"os"

	"wpay/common"
	"wpay/config"
	"wpay/dao"
	"wpay/router"
	"wpay/service"
	"wpay/task"
)

func main() {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "dev"
	}

	cfg, err := config.Load(env)
	if err != nil {
		panic(fmt.Sprintf("load config failed: %v", err))
	}

	common.InitLogger(cfg.Log.Level)

	if err := common.InitSnowflake(1); err != nil {
		panic(fmt.Sprintf("init snowflake failed: %v", err))
	}

	if err := dao.InitDB(&cfg.Database); err != nil {
		panic(fmt.Sprintf("init postgres failed: %v", err))
	}

	if err := dao.InitRedis(&cfg.Redis); err != nil {
		panic(fmt.Sprintf("init redis failed: %v", err))
	}

	if env == "dev" {
		if err := dao.AutoMigrate(); err != nil {
			common.Log.Warn("auto migrate: %v", err)
		}
	}
	
	// Initialize default admin for both dev and prod environments
	if err := service.InitDefaultAdmin(); err != nil {
		common.Log.Warn("init default admin: %v", err)
	}

	if err := service.InitIPWhitelist(); err != nil {
		common.Log.Warn("init ip whitelist: %v", err)
	}

	if err := dao.WarmupAllStoreCache(); err != nil {
		common.Log.Warn("warmup cache: %v", err)
	}

	cron := task.StartCron()
	defer cron.Stop()

	r := router.Setup(cfg.Server.Mode)
	addr := fmt.Sprintf(":%d", cfg.Server.Port)

	common.Log.Info("%s server starting on %s env=%s", common.ServiceDisplayName, addr, env)

	if cfg.Server.TLSEnabled {
		if err := r.RunTLS(addr, cfg.Server.CertFile, cfg.Server.KeyFile); err != nil {
			panic(err)
		}
	} else {
		if err := r.Run(addr); err != nil {
			panic(err)
		}
	}
}
