package config

import (
	"fmt"
	"log"
	"strconv"

	"openpenpal-backend/internal/models"
	"shared/pkg/database"

	"gorm.io/gorm"
)

// SetupDatabase 初始化数据库连接 - 使用统一数据库管理器
func SetupDatabase(config *Config) (*gorm.DB, error) {
	// 创建统一数据库配置
	dbConfig := &database.Config{
		Type:     database.DatabaseType(config.DatabaseType),
		Database: config.DatabaseURL,
		Host:     config.DBHost,
		Username: config.DBUser,
		Password: config.DBPassword,
		SSLMode:  config.DBSSLMode,
	}
	
	// 处理端口号
	if config.DBPort != "" {
		if port, err := strconv.Atoi(config.DBPort); err == nil {
			dbConfig.Port = port
		}
	}

	// 处理不同数据库类型
	if config.DatabaseType == "postgres" || config.DatabaseType == "postgresql" {
		dbConfig.Type = database.PostgreSQL
		// 如果没有完整的URL，使用DatabaseName
		if config.DatabaseURL == "./openpenpal.db" || config.DatabaseURL == "" {
			dbConfig.Database = config.DatabaseName
		}
	} else if config.DatabaseType == "sqlite" {
		dbConfig.Type = database.SQLite
		dbConfig.Database = config.DatabaseURL
	} else {
		return nil, fmt.Errorf("unsupported database type: %s", config.DatabaseType)
	}

	// 使用统一数据库管理器连接
	db, err := database.InitDefaultDatabase(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 自动迁移表结构
	if err := autoMigrate(db); err != nil {
		return nil, fmt.Errorf("failed to auto migrate: %w", err)
	}

	log.Println("Database connected via unified manager and migrated successfully")
	
	// Run extended migrations for new features
	log.Println("Starting extended migrations...")
	if err := MigrateExtendedModels(db); err != nil {
		log.Printf("Extended migration error: %v", err)
		return nil, fmt.Errorf("failed to run extended migrations: %w", err)
	}
	log.Println("Extended migrations completed successfully")
	
	return db, nil
}

// autoMigrate 自动迁移数据库表结构
func autoMigrate(db *gorm.DB) error {
	log.Println("Starting main auto migration...")
	err := db.AutoMigrate(
		&models.User{},
		&models.UserProfile{},
		&models.Letter{},
		&models.LetterCode{},
		&models.StatusLog{},
		&models.LetterPhoto{},
		&models.LetterLike{},
		&models.LetterShare{},
		// Note: LetterTemplate moved to extended migration to handle null values
		&models.LetterThread{}, 
		&models.LetterReply{},
		&models.Courier{},
		&models.CourierTask{},
		// AI相关模型
		&models.AIMatch{},
		&models.AIReply{},
		&models.AIReplyAdvice{},
		&models.AIInspiration{},
		&models.AICuration{},
		&models.AIConfig{},
		&models.AIUsageLog{},
		// 审核相关模型
		&models.ModerationRecord{},
		&models.SensitiveWord{},
		&models.ModerationRule{},
		&models.ModerationQueue{},
		&models.ModerationStats{},
		// 通知相关模型
		&models.Notification{},
		&models.EmailTemplate{},
		&models.EmailLog{},
		&models.NotificationPreference{},
		&models.NotificationBatch{},
		// 博物馆相关模型
		&models.MuseumItem{},
		&models.MuseumCollection{},
		&models.MuseumExhibitionEntry{},
		&models.MuseumEntry{},
		&models.MuseumExhibition{},
		// 信封相关模型
		&models.EnvelopeDesign{},
		&models.Envelope{},
		&models.EnvelopeVote{},
		&models.EnvelopeOrder{},
		// 商店相关模型
		&models.Product{},
		&models.Cart{},
		&models.CartItem{},
		&models.Order{},
		&models.OrderItem{},
		&models.ProductReview{},
		&models.ProductFavorite{},
		// 数据分析相关模型
		&models.AnalyticsMetric{},
		&models.UserAnalytics{},
		&models.SystemAnalytics{},
		&models.PerformanceMetric{},
		&models.AnalyticsReport{},
		// 任务调度相关模型
		&models.ScheduledTask{},
		&models.TaskExecution{},
		&models.TaskTemplate{},
		&models.TaskWorker{},
		// 存储相关模型
		&models.StorageFile{},
		&models.StorageConfig{},
		&models.StorageOperation{},
		// 积分系统相关模型
		&models.UserCredit{},
		&models.CreditTransaction{},
		&models.CreditRule{},
		&models.UserLevel{},
		// 系统配置相关模型
		&models.SystemSettings{},
	)
	if err != nil {
		log.Printf("Main migration error: %v", err)
		return err
	}
	log.Println("Main auto migration completed successfully")
	return nil
}

// SeedData 安全初始化测试数据 - 重构版本
func SeedData(db *gorm.DB) error {
	log.Println("🔐 Using SECURE seed data system...")
	
	// 使用安全种子管理器
	bcryptCost := 12 // 生产级别的bcrypt成本
	seedManager := NewSecureSeedManager(db, bcryptCost)
	
	// 执行安全种子数据生成
	if err := seedManager.SecureSeedData(); err != nil {
		return fmt.Errorf("secure seed failed: %w", err)
	}
	
	return nil
}

// LegacySeedData 旧版本的硬编码种子数据（已弃用 - 仅保留作为参考）
// ⚠️ 警告：此函数包含硬编码密码哈希，在生产环境中不安全！
func LegacySeedData(db *gorm.DB) error {
	log.Println("⚠️ WARNING: Using LEGACY seed data with hardcoded hashes!")
	log.Println("⚠️ This is INSECURE for production use!")
	
	// 检查是否已有数据
	var userCount int64
	if err := db.Model(&models.User{}).Count(&userCount).Error; err != nil {
		return err
	}

	if userCount > 0 {
		log.Println("Database already seeded, skipping...")
		return nil
	}

	// ⚠️ 以下是不安全的硬编码哈希 - 已弃用
	testUsers := []models.User{
		// 普通用户
		{
			ID:           "test-user-1",
			Username:     "alice",
			Email:        "alice@example.com",
			PasswordHash: "$2a$10$N9qo8uLOickgx2ZMRZoMye.Lf/rSDRfYHYxX1dpIjJJNzTHFN1UTO", // secret
			Nickname:     "Alice",
			Role:         models.RoleUser,
			SchoolCode:   "BJDX01",
			IsActive:     true,
		},
		{
			ID:           "test-user-2",
			Username:     "bob",
			Email:        "bob@example.com",
			PasswordHash: "$2a$10$N9qo8uLOickgx2ZMRZoMye.Lf/rSDRfYHYxX1dpIjJJNzTHFN1UTO", // secret
			Nickname:     "Bob",
			Role:         models.RoleUser,
			SchoolCode:   "BJDX01",
			IsActive:     true,
		},
		// 信使
		{
			ID:           "test-courier-1",
			Username:     "courier1",
			Email:        "courier1@example.com",
			PasswordHash: "$2a$10$N9qo8uLOickgx2ZMRZoMye.Lf/rSDRfYHYxX1dpIjJJNzTHFN1UTO", // secret
			Nickname:     "普通信使",
			Role:         models.RoleCourier,
			SchoolCode:   "BJDX01",
			IsActive:     true,
		},
		// 高级信使
		{
			ID:           "test-senior-courier",
			Username:     "senior_courier",
			Email:        "senior@example.com",
			PasswordHash: "$2a$10$N9qo8uLOickgx2ZMRZoMye.Lf/rSDRfYHYxX1dpIjJJNzTHFN1UTO", // secret
			Nickname:     "高级信使",
			Role:         models.RoleSeniorCourier,
			SchoolCode:   "BJDX01",
			IsActive:     true,
		},
		// 信使协调员
		{
			ID:           "test-coordinator",
			Username:     "coordinator",
			Email:        "coordinator@example.com",
			PasswordHash: "$2a$10$N9qo8uLOickgx2ZMRZoMye.Lf/rSDRfYHYxX1dpIjJJNzTHFN1UTO", // secret
			Nickname:     "信使协调员",
			Role:         models.RoleCourierCoordinator,
			SchoolCode:   "BJDX01",
			IsActive:     true,
		},
		// 学校管理员
		{
			ID:           "test-school-admin",
			Username:     "school_admin",
			Email:        "school_admin@example.com",
			PasswordHash: "$2a$10$N9qo8uLOickgx2ZMRZoMye.Lf/rSDRfYHYxX1dpIjJJNzTHFN1UTO", // secret
			Nickname:     "北大管理员",
			Role:         models.RoleSchoolAdmin,
			SchoolCode:   "BJDX01",
			IsActive:     true,
		},
		// 平台管理员
		{
			ID:           "test-platform-admin",
			Username:     "platform_admin",
			Email:        "platform_admin@example.com",
			PasswordHash: "$2a$10$N9qo8uLOickgx2ZMRZoMye.Lf/rSDRfYHYxX1dpIjJJNzTHFN1UTO", // secret
			Nickname:     "平台管理员",
			Role:         models.RolePlatformAdmin,
			SchoolCode:   "SYSTEM",
			IsActive:     true,
		},
		// 超级管理员
		{
			ID:           "test-super-admin",
			Username:     "super_admin",
			Email:        "super_admin@example.com",
			PasswordHash: "$2a$10$N9qo8uLOickgx2ZMRZoMye.Lf/rSDRfYHYxX1dpIjJJNzTHFN1UTO", // secret
			Nickname:     "超级管理员",
			Role:         models.RoleSuperAdmin,
			SchoolCode:   "SYSTEM",
			IsActive:     true,
		},
		// 四级信使系统测试账号 - 使用正确的密码哈希
		{
			ID:           "courier-level1",
			Username:     "courier_level1",
			Email:        "courier1@openpenpal.com",
			PasswordHash: "$2a$10$N9qo8uLOickgx2ZMRZoMye.Lf/rSDRfYHYxX1dpIjJJNzTHFN1UTO", // secret
			Nickname:     "一级信使",
			Role:         models.RoleCourierLevel1,
			SchoolCode:   "PKU001",
			IsActive:     true,
		},
		{
			ID:           "courier-level2",
			Username:     "courier_level2",
			Email:        "courier2@openpenpal.com",
			PasswordHash: "$2a$10$N9qo8uLOickgx2ZMRZoMye.Lf/rSDRfYHYxX1dpIjJJNzTHFN1UTO", // secret
			Nickname:     "二级信使",
			Role:         models.RoleCourierLevel2,
			SchoolCode:   "PKU001",
			IsActive:     true,
		},
		{
			ID:           "courier-level3",
			Username:     "courier_level3",
			Email:        "courier3@openpenpal.com",
			PasswordHash: "$2a$10$N9qo8uLOickgx2ZMRZoMye.Lf/rSDRfYHYxX1dpIjJJNzTHFN1UTO", // secret
			Nickname:     "三级信使",
			Role:         models.RoleCourierLevel3,
			SchoolCode:   "PKU001",
			IsActive:     true,
		},
		{
			ID:           "courier-level4",
			Username:     "courier_level4",
			Email:        "courier4@openpenpal.com",
			PasswordHash: "$2a$10$N9qo8uLOickgx2ZMRZoMye.Lf/rSDRfYHYxX1dpIjJJNzTHFN1UTO", // secret
			Nickname:     "四级信使",
			Role:         models.RoleCourierLevel4,
			SchoolCode:   "PKU001",
			IsActive:     true,
		},
		{
			ID:           "courier-building-1",
			Username:     "courier_building",
			Email:        "courier_building@penpal.com",
			PasswordHash: "$2a$10$Cm0hFv7kUKfUc5Q6booKiehnQsHSFF7.4LYuqWVkgFqCYda3qqGCS", // courier001
			Nickname:     "楼层信使",
			Role:         models.RoleCourier,
			SchoolCode:   "PKU001",
			IsActive:     true,
		},
		{
			ID:           "courier-area-1",
			Username:     "courier_area",
			Email:        "courier_area@penpal.com",
			PasswordHash: "$2a$10$b75vhT53SdpdtJRcf4WzrOOpLAaBRgZ9Ix.AEfrH/UngIxoxscQNm", // courier002
			Nickname:     "片区信使",
			Role:         models.RoleSeniorCourier,
			SchoolCode:   "PKU001",
			IsActive:     true,
		},
		{
			ID:           "courier-school-1",
			Username:     "courier_school",
			Email:        "courier_school@penpal.com",
			PasswordHash: "$2a$10$ClnxSMuPM6YdlWXuswYE1OjWm06yR48cdGEqp0/YP/h9OI/u2gwvm", // courier003
			Nickname:     "学校信使",
			Role:         models.RoleCourierCoordinator,
			SchoolCode:   "PKU001",
			IsActive:     true,
		},
		{
			ID:           "courier-city-1",
			Username:     "courier_city",
			Email:        "courier_city@penpal.com",
			PasswordHash: "$2a$10$9V.Mbl5QqL0.tZWaJ0nTrulHIXPgeyWaex.lKrvG.r5HqDaldbd6S", // courier004
			Nickname:     "城市信使",
			Role:         models.RolePlatformAdmin,
			SchoolCode:   "SYSTEM",
			IsActive:     true,
		},
		{
			ID:           "test-admin",
			Username:     "admin",
			Email:        "admin@penpal.com",
			PasswordHash: "$2a$10$dwSXE/fBcbAJVy0jMZHYI.vFjjUZFYRMPpeAzcgmHd.XqwfqgOrEW", // admin123
			Nickname:     "系统管理员",
			Role:         models.RoleSuperAdmin,
			SchoolCode:   "SYSTEM",
			IsActive:     true,
		},
	}

	for _, user := range testUsers {
		if err := db.Create(&user).Error; err != nil {
			return fmt.Errorf("failed to seed user %s: %w", user.Username, err)
		}
	}

	// 创建测试信件
	testLetters := []models.Letter{
		{
			ID:      "test-letter-1",
			UserID:  "test-user-1",
			Title:   "给朋友的第一封信",
			Content: "亲爱的朋友，\n\n这是我通过OpenPenPal发送的第一封信。希望你能收到这份温暖的问候。\n\n你的朋友\nAlice",
			Style:   models.StyleClassic,
			Status:  models.StatusDraft,
		},
		{
			ID:      "test-letter-2",
			UserID:  "test-user-2",
			Title:   "感谢信",
			Content: "谢谢你上次的帮助，我真的很感激。这个项目让我们能够重新体验手写信的魅力。\n\nBob",
			Style:   models.StyleModern,
			Status:  models.StatusGenerated,
		},
	}

	for _, letter := range testLetters {
		if err := db.Create(&letter).Error; err != nil {
			return fmt.Errorf("failed to seed letter %s: %w", letter.Title, err)
		}
	}

	log.Println("Test data seeded successfully")
	return nil
}
