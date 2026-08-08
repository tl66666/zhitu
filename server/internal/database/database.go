// Package database 负责数据库连接初始化与自动迁移
package database

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/glebarez/sqlite"
	"github.com/zhitu/server/internal/config"
	"github.com/zhitu/server/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Init 根据 config 初始化 SQLite 数据库连接
// 使用 glebarez/sqlite —— 纯 Go 实现，无需 CGO
func Init(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	// 确保数据目录存在
	if dir := filepath.Dir(cfg.Path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("create db dir: %w", err)
		}
	}

	db, err := gorm.Open(sqlite.Open(cfg.Path), &gorm.Config{
		Logger:                                   logger.Default.LogMode(parseLogLevel(cfg.LogLevel)),
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	// 启用 WAL 模式提升并发性能
	if sqlDB, err := db.DB(); err == nil {
		_, _ = sqlDB.Exec("PRAGMA journal_mode=WAL;")
		_, _ = sqlDB.Exec("PRAGMA foreign_keys=ON;")
	}

	// 自动迁移
	if err := autoMigrate(db); err != nil {
		return nil, fmt.Errorf("auto migrate: %w", err)
	}

	return db, nil
}

// autoMigrate 自动建表/补字段
// 新增业务模块时在此处追加迁移
func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.BrowserWorkspace{},
		&models.BrowserState{},
		// 用户档案
		&models.UserProfile{},
		&models.UserEducation{},
		&models.UserWork{},
		&models.UserProject{},
		&models.UserSkill{},
		&models.UserHonor{},
		&models.UserPractice{},
		// 简历
		&models.Resume{},
		&models.ResumeVersion{},
		&models.ResumeAIOperation{},
		// 模拟面试
		&models.Interview{},
		&models.InterviewMessage{},
		&models.InterviewScore{},
		&models.InterviewReport{},
		// 投递看板
		&models.Delivery{},
		&models.DeliveryRound{},
		&models.DeliveryFeedback{},
	)
}

// parseLogLevel 将配置字符串映射为 gorm 日志级别
func parseLogLevel(s string) logger.LogLevel {
	switch s {
	case "silent":
		return logger.Silent
	case "error":
		return logger.Error
	case "warn":
		return logger.Warn
	case "info":
		return logger.Info
	default:
		return logger.Warn
	}
}
