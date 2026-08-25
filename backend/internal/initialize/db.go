// Package initialize 应用初始化：日志、数据库、Redis、依赖容器、定时任务。
package initialize

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	gormlogger "gorm.io/gorm/logger"
	"gorm.io/gorm"

	"golang-learn/internal/config"
	"golang-learn/internal/model"
)

// InitDB 初始化 MySQL 连接并自动迁移表结构。
// 迁移策略：AutoMigrate 按模型生成/更新表（教学项目简化方案；
// 生产项目建议使用版本化的 SQL migration 脚本）。
func InitDB(cfg *config.MySQLConfig) (*gorm.DB, error) {
	// GORM 日志级别映射
	var logLevel gormlogger.LogLevel
	switch cfg.LogLevel {
	case "silent":
		logLevel = gormlogger.Silent
	case "error":
		logLevel = gormlogger.Error
	case "warn":
		logLevel = gormlogger.Warn
	default:
		logLevel = gormlogger.Info
	}

	gormCfg := &gorm.Config{
		Logger: gormlogger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			gormlogger.Config{
				SlowThreshold:             200 * time.Millisecond, // 慢查询阈值
				LogLevel:                  logLevel,
				IgnoreRecordNotFoundError: true, // 忽略 ErrRecordNotFound 噪音
			},
		),
	}

	// 连接数据库（带重试：Docker 环境下 MySQL 可能比应用先启动完成）
	var db *gorm.DB
	var err error
	for i := 0; i < 5; i++ {
		db, err = gorm.Open(mysql.Open(cfg.DSN()), gormCfg)
		if err == nil {
			break
		}
		log.Printf("[InitDB] 连接 MySQL 失败（第 %d 次）：%v，3 秒后重试", i+1, err)
		time.Sleep(3 * time.Second)
	}
	if err != nil {
		return nil, fmt.Errorf("连接 MySQL 失败: %w", err)
	}

	// 连接池配置
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Hour) // 连接最大存活时间，避免使用被 MySQL 关闭的连接

	// 自动迁移表结构（顺序无依赖，一次迁移全部模型）
	if err := db.AutoMigrate(
		&model.User{},
		&model.Friend{},
		&model.Account{},
		&model.Group{},
		&model.Tag{},
		&model.Interaction{},
		&model.Reminder{},
		&model.ChatMessage{},
	); err != nil {
		return nil, fmt.Errorf("自动迁移失败: %w", err)
	}

	return db, nil
}
