package main

import (
	"fmt"
	"log"

	"openpenpal-backend/internal/config"
	"openpenpal-backend/internal/logger"
	"openpenpal-backend/internal/services"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	fmt.Println("🔐 Starting sensitive data encryption migration...")

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志
	smartLogger := logger.NewSmartLogger("encrypt_migration", logger.INFO)

	// 连接数据库
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 创建加密服务
	encryptionService, err := services.NewEncryptionService(cfg, smartLogger)
	if err != nil {
		log.Fatalf("Failed to create encryption service: %v", err)
	}

	// 创建迁移服务
	migrationService := services.NewEncryptionMigrationService(db, encryptionService, smartLogger)

	// 执行迁移
	if err := migrationService.MigrateUserProfiles(); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	fmt.Println("✅ Sensitive data encryption migration completed successfully!")
}