package config

import (
	"openpenpal-backend/internal/models"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// SetupTestDB 设置测试数据库
func SetupTestDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // 测试时关闭日志
	})
	if err != nil {
		return nil, err
	}
	
	// 自动迁移所有模型
	err = db.AutoMigrate(
		&models.User{},
		&models.Letter{},
		&models.LetterCode{},
		&models.StatusLog{}, // 添加StatusLog模型
		&models.Courier{},
		&models.CourierTask{},
		&models.Notification{},
		&models.NotificationPreference{},
		&models.MuseumItem{},
		&models.MuseumExhibition{},
		&models.Envelope{},
		&models.EnvelopeDesign{},
		&models.EnvelopeOrder{},
		&models.LetterTemplate{},
		&models.AIUsageLog{},
		&models.Product{},
		&models.Order{},
		&models.UserCredit{},
		&models.CreditTransaction{},
		&models.SystemSettings{},
		&models.AdminActivity{},
		&models.ModerationRecord{},
		&models.ModerationRule{},
	)
	if err != nil {
		return nil, err
	}
	
	return db, nil
}

// GetTestConfig 获取测试配置
func GetTestConfig() *Config {
	return &Config{
		// Server
		Port:        "8080",
		Host:        "localhost",
		Environment: "test",

		// Database
		DatabaseType: "sqlite",
		DatabaseURL:  ":memory:",
		DatabaseName: "test",
		
		// App
		AppName:    "OpenPenPal Test",
		AppVersion: "1.0.0",
		BaseURL:    "http://localhost:8080",

		// Security
		JWTSecret:  "test-secret-key-for-testing-only",
		BCryptCost: 10, // 测试时降低成本加快速度

		// Frontend
		FrontendURL: "http://localhost:3000",

		// QR Code
		QRCodeStorePath: "/tmp/test-qrcodes",

		// AI
		MoonshotAPIKey: "test-moonshot-key",
		AIProvider:     "moonshot",
		
		// Email/SMTP (测试时禁用)
		SMTPHost:         "",
		SMTPPort:         0,
		SMTPUsername:     "",
		SMTPPassword:     "",
		EmailFromAddress: "test@openpenpal.com",
		EmailFromName:    "OpenPenPal Test",
	}
}

// CreateTestUser 创建测试用户
func CreateTestUser(db *gorm.DB, username string, role models.UserRole) *models.User {
	user := &models.User{
		ID:           "test-" + username,
		Username:     username,
		Email:        username + "@test.com",
		PasswordHash: "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy", // password: "password123"
		Nickname:     "Test " + username,
		Role:         role,
		SchoolCode:   "BJDX",
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	
	db.Create(user)
	return user
}

// CreateTestLetter 创建测试信件
func CreateTestLetter(db *gorm.DB, userID string) *models.Letter {
	letter := &models.Letter{
		ID:              "test-letter-" + userID,
		UserID:          userID,
		Title:           "Test Letter",
		Content:         "This is a test letter content.",
		Style:           models.StyleClassic,
		Status:          models.StatusDraft,
		Visibility:      models.VisibilityPrivate,
		RecipientOPCode: "PK5F01",
		SenderOPCode:    "PK3D12",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	
	db.Create(letter)
	return letter
}

// CreateTestCourier 创建测试信使
func CreateTestCourier(db *gorm.DB, userID string, level int) *models.Courier {
	courier := &models.Courier{
		ID:                  "courier-" + userID,
		UserID:              userID,
		Name:                "Test Courier",
		Contact:             "13800138000",
		School:              "北京大学",
		Zone:                "BJDX",
		HasPrinter:          true,
		Level:               level,
		Status:              "approved",
		Points:              0,
		TaskCount:           0,
		ManagedOPCodePrefix: "PK",
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}
	
	db.Create(courier)
	return courier
}