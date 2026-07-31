package dao

import (
	"fmt"
	"time"

	"wpay/config"
	"wpay/model"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB(cfg *config.DatabaseConfig) error {
	logLevel := logger.Info
	if !config.IsProd() {
		logLevel = logger.Warn
	}
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger:                 logger.Default.LogMode(logLevel),
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return fmt.Errorf("connect postgres: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Hour)
	DB = db
	return nil
}

func AutoMigrate() error {
	if err := DB.AutoMigrate(
		&model.Admin{},
		&model.Store{},
		&model.PayChannel{},
		&model.PayChannelHistory{},
		&model.Order{},
		&model.SensitiveLog{},
		&model.IPWhitelistPolicy{},
		&model.IPWhitelistEntry{},
		&model.PayPlatform{},
	); err != nil {
		return err
	}
	// 允许多商户码：移除旧版 store_id+pay_type 唯一索引
	_ = DB.Exec("DROP INDEX IF EXISTS idx_store_pay_type").Error
	return nil
}
