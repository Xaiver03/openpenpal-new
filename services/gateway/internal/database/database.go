package database

import (
	"api-gateway/internal/models"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"go.uber.org/zap"
)

// InitDB 初始化数据库连接
func InitDB(databaseURL string, zapLogger *zap.Logger) (*gorm.DB, error) {
	// 配置GORM日志器
	gormLogger := logger.New(
		NewGormLoggerAdapter(zapLogger),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	// 连接数据库
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 获取底层的sql.DB
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// 设置连接池
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	zapLogger.Info("Database connection established successfully")

	// 自动迁移
	if err := autoMigrate(db, zapLogger); err != nil {
		return nil, fmt.Errorf("failed to auto migrate: %w", err)
	}

	return db, nil
}

// autoMigrate 自动迁移数据库表
func autoMigrate(db *gorm.DB, logger *zap.Logger) error {
	logger.Info("Starting database auto migration...")

	// 迁移所有模型
	err := db.AutoMigrate(
		&models.PerformanceMetric{},
		&models.PerformanceAlert{},
	)

	if err != nil {
		logger.Error("Failed to auto migrate database", zap.Error(err))
		return err
	}

	logger.Info("Database auto migration completed successfully")
	return nil
}

// GormLoggerAdapter GORM日志适配器
type GormLoggerAdapter struct {
	logger *zap.Logger
}

// NewGormLoggerAdapter 创建GORM日志适配器
func NewGormLoggerAdapter(logger *zap.Logger) *GormLoggerAdapter {
	return &GormLoggerAdapter{logger: logger}
}

// Printf 实现gorm/logger接口
func (l *GormLoggerAdapter) Printf(template string, args ...interface{}) {
	l.logger.Info(fmt.Sprintf(template, args...))
}

// CloseDB 关闭数据库连接
func CloseDB(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}