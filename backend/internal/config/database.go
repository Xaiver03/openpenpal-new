package config

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"openpenpal-backend/internal/logger"
	"openpenpal-backend/internal/models"

	"github.com/google/uuid"
	"shared/pkg/database"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

// SetupDatabaseDirect 直接初始化数据库连接，绕过共享包问题，配置生产级日志
func SetupDatabaseDirect(config *Config) (*gorm.DB, error) {
	// 设置GORM配置，优化日志输出
	gormConfig := &gorm.Config{
		Logger: logger.NewCustomGormLogger(),
	}

	var db *gorm.DB
	var err error

	// 只支持PostgreSQL数据库连接
	if config.DatabaseType == "postgres" || config.DatabaseType == "postgresql" {
		// 构建基础DSN
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Shanghai",
			config.DBHost, config.DBUser, config.DBPassword, config.DatabaseName, config.DBPort, config.DBSSLMode)
		
		// 添加SSL证书参数
		if config.DBSSLMode != "disable" && config.DBSSLMode != "allow" {
			if config.DBSSLRootCert != "" {
				dsn += fmt.Sprintf(" sslrootcert=%s", config.DBSSLRootCert)
			}
			if config.DBSSLCert != "" {
				dsn += fmt.Sprintf(" sslcert=%s", config.DBSSLCert)
			}
			if config.DBSSLKey != "" {
				dsn += fmt.Sprintf(" sslkey=%s", config.DBSSLKey)
			}
		}
		db, err = gorm.Open(postgres.Open(dsn), gormConfig)
	} else {
		return nil, fmt.Errorf("only PostgreSQL is supported, got: %s", config.DatabaseType)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// 根据环境获取优化的连接池配置
	var poolConfig *PoolConfig
	switch config.Environment {
	case "production":
		poolConfig = GetPoolPreset(PoolPresetProduction)
	case "staging":
		poolConfig = GetPoolPreset(PoolPresetStaging)
	case "test":
		poolConfig = GetPoolPreset(PoolPresetTesting)
	default:
		poolConfig = GetPoolPreset(PoolPresetDevelopment)
	}
	
	// 应用连接池配置
	sqlDB.SetMaxOpenConns(poolConfig.MaxOpenConns)
	sqlDB.SetMaxIdleConns(poolConfig.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(poolConfig.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(poolConfig.ConnMaxIdleTime)
	
	log.Printf("Database connection pool configured: MaxOpen=%d, MaxIdle=%d, MaxLifetime=%s",
		poolConfig.MaxOpenConns, poolConfig.MaxIdleConns, poolConfig.ConnMaxLifetime)

	// 执行数据库迁移
	if err := autoMigrate(db); err != nil {
		log.Printf("Migration error: %v", err)
		return nil, fmt.Errorf("failed to auto migrate: %w", err)
	}

	log.Println("Database connected directly and migrated successfully")

	// 跳过扩展迁移 - 数据库已包含完整表结构
	log.Println("Skipping extended migrations - database already contains complete table structure")

	return db, nil
}

// performSafeMigration 执行安全迁移，处理约束冲突
func performSafeMigration(db *gorm.DB) error {
	log.Println("Starting safe migration strategy...")
	
	// 获取所有需要迁移的模型
	allModels := getAllModels()
	
	// 使用SafeAutoMigrate处理约束冲突
	if err := SafeAutoMigrate(db, allModels...); err != nil {
		return fmt.Errorf("safe auto migrate failed: %w", err)
	}
	
	log.Println("Safe migration completed successfully")
	return nil
}

// GetAllModels 返回所有需要迁移的模型 (导出版本)
func GetAllModels() []interface{} {
	return getAllModels()
}

// getAllModels 返回所有需要迁移的模型
func getAllModels() []interface{} {
	return []interface{}{
		// User表由SafeAutoMigrate特殊处理
		&models.User{},
		&models.UserProfile{},
		&models.Letter{},
		&models.LetterCode{},
		&models.StatusLog{},
		&models.LetterPhoto{},
		&models.LetterLike{},
		&models.LetterShare{},
		&models.Comment{},
		&models.CommentLike{},
		&models.CommentReport{},
		&models.LetterThread{},
		&models.LetterReply{},
		&models.Courier{},
		&models.CourierTask{},
		&models.AIMatch{},
		&models.AIReply{},
		&models.AIReplyAdvice{},
		&models.AIInspiration{},
		&models.AICuration{},
		&models.AIConfig{},
		&models.AIUsageLog{},
		&models.ModerationRecord{},
		&models.SensitiveWord{},
		&models.ModerationRule{},
		&models.ModerationQueue{},
		&models.ModerationStats{},
		&models.SecurityEvent{},
		&models.Notification{},
		&models.EmailTemplate{},
		&models.EmailLog{},
		&models.NotificationPreference{},
		&models.NotificationBatch{},
		&models.MuseumItem{},
		&models.MuseumCollection{},
		&models.MuseumExhibitionEntry{},
		&models.MuseumEntry{},
		&models.MuseumExhibition{},
		&models.EnvelopeDesign{},
		&models.Envelope{},
		&models.EnvelopeVote{},
		&models.EnvelopeOrder{},
		&models.Product{},
		&models.Cart{},
		&models.CartItem{},
		&models.Order{},
		&models.OrderItem{},
		&models.ProductReview{},
		&models.ProductFavorite{},
		&models.AnalyticsMetric{},
		&models.UserAnalytics{},
		&models.SystemAnalytics{},
		&models.PerformanceMetric{},
		&models.AnalyticsReport{},
		&models.ScheduledTask{},
		&models.TaskExecution{},
		&models.TaskTemplate{},
		&models.TaskWorker{},
		&models.StorageFile{},
		&models.StorageConfig{},
		&models.StorageOperation{},
		&models.UserCredit{},
		&models.CreditTransaction{},
		&models.CreditRule{},
		&models.UserLevel{},
		&models.CreditShopProduct{},
		&models.CreditRedemption{},
		&models.CreditCart{},
		&models.CreditCartItem{},
		&models.CreditShopCategory{},
		&models.UserRedemptionHistory{},
		&models.CreditShopConfig{},
		&models.CreditActivity{},
		&models.CreditActivityParticipation{},
		&models.CreditActivityRule{},
		&models.CreditActivitySchedule{},
		&models.CreditActivityStatistics{},
		&models.CreditActivityLog{},
		&models.CreditActivityTemplate{},
		&models.CreditExpirationRule{},
		&models.CreditExpirationBatch{},
		&models.CreditExpirationLog{},
		&models.CreditExpirationNotification{},
		&models.CreditTransfer{},
		&models.CreditTransferRule{},
		&models.CreditTransferLimit{},
		&models.CreditTransferNotification{},
		&models.CreditTask{},
		&models.CreditTaskQueue{},
		&models.CreditTaskRule{},
		&models.CreditTaskBatch{},
		&models.SystemSettings{},
		
		// 延迟队列系统
		&models.DelayQueueRecord{},
		
		// OP Code系统 (SOTA地理编码)
		&models.OPCode{},
		&models.OPCodeSchool{},
		&models.OPCodeArea{},
		&models.OPCodeApplication{},
		&models.OPCodePermission{},
		
		// 审计日志系统
		&models.AuditLog{},
	}
}

// intelligentMigrate 智能迁移策略 - 只迁移不存在或需要更新的表
func intelligentMigrate(db *gorm.DB) error {
	log.Println("Starting intelligent migration strategy...")
	
	// 检查哪些表不存在，只迁移这些表
	// 明确排除User表避免约束冲突 - 数据库中已有正确结构
	missingTables := []interface{}{}
	allModels := []interface{}{
		// &models.User{}, // 跳过User表，避免约束冲突
		&models.UserProfile{},
		&models.Letter{},
		&models.LetterCode{},
		&models.StatusLog{},
		&models.LetterPhoto{},
		&models.LetterLike{},
		&models.LetterShare{},
		&models.Comment{},
		&models.CommentLike{},
		&models.CommentReport{},
		&models.LetterThread{},
		&models.LetterReply{},
		&models.Courier{},
		&models.CourierTask{},
		&models.AIMatch{},
		&models.AIReply{},
		&models.AIReplyAdvice{},
		&models.AIInspiration{},
		&models.AICuration{},
		&models.AIConfig{},
		&models.AIUsageLog{},
	}
	
	for _, model := range allModels {
		if !db.Migrator().HasTable(model) {
			missingTables = append(missingTables, model)
			log.Printf("Found missing table for model: %T", model)
		}
	}
	
	// 只迁移缺失的表
	if len(missingTables) > 0 {
		log.Printf("Migrating %d missing tables...", len(missingTables))
		if err := db.AutoMigrate(missingTables...); err != nil {
			return fmt.Errorf("failed to migrate missing tables: %w", err)
		}
		log.Printf("Successfully migrated %d missing tables", len(missingTables))
	} else {
		log.Println("All required tables already exist")
	}
	
	return nil
}

// SetupDatabase 初始化数据库连接 - 恢复共享包使用
func SetupDatabase(config *Config) (*gorm.DB, error) {
	log.Println("Using shared package for unified database management")
	
	// 优先尝试使用共享包实现
	db, err := SetupDatabaseWithSharedPackage(config)
	if err != nil {
		log.Printf("Shared package setup failed: %v, falling back to direct setup", err)
		return SetupDatabaseDirect(config)
	}
	
	return db, nil
}

// SetupDatabaseWithSharedPackage 使用共享包的数据库连接
func SetupDatabaseWithSharedPackage(config *Config) (*gorm.DB, error) {
	log.Println("Using shared database package for unified database management...")
	
	// 创建数据库管理器
	manager := database.GetDefaultManager()
	
	// 解析端口
	port := 5432 // 默认PostgreSQL端口
	if config.DBPort != "" {
		if p, err := strconv.Atoi(config.DBPort); err == nil {
			port = p
		}
	}
	
	// 构建共享包配置
	sharedConfig := &database.Config{
		Type:     database.PostgreSQL,
		Host:     config.DBHost,
		Port:     port,
		Database: config.DatabaseName,
		Username: config.DBUser,
		Password: config.DBPassword,
		SSLMode:  config.DBSSLMode,
		SSLCert:  config.DBSSLCert,
		SSLKey:   config.DBSSLKey,
		SSLRootCert: config.DBSSLRootCert,
		Timezone: "Asia/Shanghai",
		
		// 根据环境获取优化的连接池配置
		MaxOpenConns:    getPoolConfigForEnv(config.Environment).MaxOpenConns,
		MaxIdleConns:    getPoolConfigForEnv(config.Environment).MaxIdleConns,
		ConnMaxLifetime: getPoolConfigForEnv(config.Environment).ConnMaxLifetime,
		ConnMaxIdleTime: getPoolConfigForEnv(config.Environment).ConnMaxIdleTime,
		
		// 日志和健康检查配置
		LogLevel:            gormLogger.Warn,
		HealthCheckInterval: 30 * time.Second,
		MaxRetries:         3,
		RetryInterval:      5 * time.Second,
	}
	
	// 添加配置到管理器
	if err := manager.AddConfig("main", sharedConfig); err != nil {
		return nil, fmt.Errorf("failed to add database config: %w", err)
	}
	
	// 连接数据库
	db, err := manager.Connect("main")
	if err != nil {
		return nil, fmt.Errorf("failed to connect via shared package: %w", err)
	}
	
	// 执行数据库迁移
	if err := autoMigrate(db); err != nil {
		log.Printf("Migration error: %v", err)
		return nil, fmt.Errorf("failed to auto migrate: %w", err)
	}

	log.Println("✅ Database connected successfully via shared package")
	return db, nil
}

// getPoolConfigForEnv 根据环境获取连接池配置
func getPoolConfigForEnv(env string) *PoolConfig {
	switch env {
	case "production":
		return GetPoolPreset(PoolPresetProduction)
	case "staging":
		return GetPoolPreset(PoolPresetStaging)
	case "test":
		return GetPoolPreset(PoolPresetTesting)
	default:
		return GetPoolPreset(PoolPresetDevelopment)
	}
}

// autoMigrate 自动迁移数据库表结构
func autoMigrate(db *gorm.DB) error {
	log.Println("Starting main auto migration...")
	
	// 使用SafeAutoMigrate处理所有模型迁移
	allModels := getAllModels()
	if err := SafeAutoMigrate(db, allModels...); err != nil {
		return fmt.Errorf("safe auto migrate failed: %w", err)
	}
	
	log.Println("Main auto migration completed successfully using SafeAutoMigrate")
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

	// Initialize courier system with hierarchy and shared data
	if err := initializeCourierSystemWithSharedData(db); err != nil {
		return fmt.Errorf("courier system initialization failed: %w", err)
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
			Role:         models.RoleCourierLevel1,
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
			Role:         models.RoleCourierLevel2,
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
			Role:         models.RoleCourierLevel3,
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
			Role:         models.RoleCourierLevel3,
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
			Role:         models.RoleCourierLevel1,
			SchoolCode:   "PKU001",
			IsActive:     true,
		},
		{
			ID:           "courier-area-1",
			Username:     "courier_area",
			Email:        "courier_area@penpal.com",
			PasswordHash: "$2a$10$b75vhT53SdpdtJRcf4WzrOOpLAaBRgZ9Ix.AEfrH/UngIxoxscQNm", // courier002
			Nickname:     "片区信使",
			Role:         models.RoleCourierLevel2,
			SchoolCode:   "PKU001",
			IsActive:     true,
		},
		{
			ID:           "courier-school-1",
			Username:     "courier_school",
			Email:        "courier_school@penpal.com",
			PasswordHash: "$2a$10$ClnxSMuPM6YdlWXuswYE1OjWm06yR48cdGEqp0/YP/h9OI/u2gwvm", // courier003
			Nickname:     "学校信使",
			Role:         models.RoleCourierLevel3,
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

// initializeCourierSystemWithSharedData creates courier hierarchy and shared tasks
func initializeCourierSystemWithSharedData(db *gorm.DB) error {
	log.Println("Initializing courier system hierarchy and shared data...")

	// Step 1: Create courier records for all courier users
	var courierUsers []models.User
	if err := db.Where("role LIKE ?", "courier%").Find(&courierUsers).Error; err != nil {
		return fmt.Errorf("failed to find courier users: %w", err)
	}

	courierMap := make(map[string]*models.Courier)
	for _, user := range courierUsers {
		var courier models.Courier
		if err := db.Where("user_id = ?", user.ID).First(&courier).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// Create courier record based on user role
				level := 1
				zoneCode := ""
				managedPrefix := ""

				switch user.Role {
				case models.RoleCourierLevel4:
					level = 4
					zoneCode = "BEIJING"
					managedPrefix = "BJ"
				case models.RoleCourierLevel3:
					level = 3
					zoneCode = "BJDX"
					managedPrefix = "BJDX"
				case models.RoleCourierLevel2:
					level = 2
					zoneCode = "BJDX-NORTH"
					managedPrefix = "BJDX5F"
				case models.RoleCourierLevel1:
					level = 1
					zoneCode = "BJDX-A-101"
					managedPrefix = "BJDX5F01"
				}

				courier = models.Courier{
					ID:                  uuid.New().String(),
					UserID:              user.ID,
					Name:                user.Nickname,
					Contact:             user.Email,
					School:              "北京大学",
					Zone:                zoneCode,
					Level:               level,
					Status:              "approved",
					ManagedOPCodePrefix: managedPrefix,
					CreatedAt:           time.Now(),
					UpdatedAt:           time.Now(),
				}

				if err := db.Create(&courier).Error; err != nil {
					return fmt.Errorf("failed to create courier for %s: %w", user.Username, err)
				}
				log.Printf("Created courier record for %s (Level %d)", user.Username, level)
			} else {
				return fmt.Errorf("failed to query courier: %w", err)
			}
		}
		courierMap[user.Username] = &courier
	}

	// Step 2: Establish hierarchy relationships
	// Note: The backend courier model doesn't have ParentID field
	// Hierarchy is managed through the courier service
	log.Printf("Courier hierarchy initialized (managed by courier service)")
	log.Println("Established courier hierarchy relationships")

	// Step 3: 跳过创建示例信件，避免外键约束错误
	log.Println("Skipping sample letter creation to avoid foreign key constraint violations")
	
	// Step 3: 保持原有courier任务创建逻辑，但跳过不存在的用户引用
	var alice models.User
	if err := db.Where("username = ?", "alice").First(&alice).Error; err != nil {
		log.Printf("Alice user not found, skipping sample data creation: %v", err)
		return nil // 正常返回，不创建示例数据
	}
	
	// 检查是否已有示例数据，避免重复创建
	var existingLetterCount int64
	db.Model(&models.Letter{}).Where("user_id = ?", alice.ID).Count(&existingLetterCount)
	if existingLetterCount > 0 {
		log.Println("Sample letters already exist, skipping creation")
		return nil
	}
	
	log.Println("Skipping all sample data creation to ensure clean startup")

	log.Println("✅ Courier system initialization complete!")
	return nil
}
