package config

import (
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"openpenpal-backend/internal/models"
)

// SetupDatabaseDirect 直接使用 GORM 初始化数据库连接（绕过 shared 包的问题）
func SetupDatabaseDirect(config *Config) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	switch config.DatabaseType {
	case "postgres":
		log.Println("Connecting to PostgreSQL...")
		// 构建 PostgreSQL 连接字符串
		var dsn string
		if config.DatabaseURL != "" && config.DatabaseURL != "./openpenpal.db" && strings.HasPrefix(config.DatabaseURL, "postgres") {
			// 如果提供了完整的 PostgreSQL URL，直接使用
			dsn = config.DatabaseURL
		} else {
			// 从单独的参数构建连接字符串
			dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
				config.DBHost,
				config.DBPort,
				config.DBUser,
				config.DBPassword,
				config.DatabaseName,
				config.DBSSLMode,
			)
		}
		log.Printf("PostgreSQL DSN: host=%s port=%s user=%s dbname=%s sslmode=%s",
			config.DBHost, config.DBPort, config.DBUser, config.DatabaseName, config.DBSSLMode)

		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
		}
	case "sqlite":
		log.Println("Connecting to SQLite...")
		db, err = gorm.Open(sqlite.Open(config.DatabaseURL), &gorm.Config{})
		if err != nil {
			return nil, fmt.Errorf("failed to connect to SQLite: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported database type: %s", config.DatabaseType)
	}

	// 自动迁移表结构
	if err := autoMigrateDirect(db); err != nil {
		return nil, fmt.Errorf("failed to auto migrate: %w", err)
	}

	log.Println("Database connected and migrated successfully")

	// Run extended migrations for new features
	log.Println("Starting extended migrations from SetupDatabaseDirect...")
	if err := MigrateExtendedModels(db); err != nil {
		log.Printf("Extended migration error: %v", err)
		return nil, fmt.Errorf("failed to run extended migrations: %w", err)
	}
	log.Println("Extended migrations completed successfully")

	// Set JSON defaults for PostgreSQL compatibility
	if config.DatabaseType == "postgres" {
		if err := SetJSONDefaults(db); err != nil {
			log.Printf("Warning: Failed to set JSON defaults: %v", err)
		}

		// 创建性能优化索引 - SOTA实现
		log.Println("Starting performance optimization...")
		if err := CreateOptimizedIndexes(db); err != nil {
			log.Printf("Warning: Failed to create optimized indexes: %v", err)
		}

		// 创建性能视图
		if err := CreatePerformanceViews(db); err != nil {
			log.Printf("Warning: Failed to create performance views: %v", err)
		}

		log.Println("Performance optimization completed")
	}

	return db, nil
}

// autoMigrateDirect 自动迁移数据库表结构
func autoMigrateDirect(db *gorm.DB) error {
	log.Println("Starting main auto migration from autoMigrateDirect...")
	// 基础模型
	modelsToMigrate := []interface{}{
		// 用户相关
		&models.User{},
		&models.UserProfile{},

		// 信件相关
		&models.Letter{},
		&models.LetterCode{},
		&models.StatusLog{},
		&models.LetterPhoto{},

		// 信使相关
		&models.Courier{},
		&models.CourierTask{},
		&models.LevelUpgradeRequest{},

		// 信封相关
		&models.Envelope{},
		&models.EnvelopeDesign{},
		&models.EnvelopeVote{},
		&models.EnvelopeOrder{},

		// AI相关模型
		&models.AIMatch{},
		&models.AIReply{},
		&models.AIInspiration{},
		&models.AICuration{},
		&models.AIConfig{},
		&models.AIUsageLog{},

		// 积分系统相关模型
		&models.UserCredit{},
		&models.CreditTransaction{},
		&models.CreditRule{},
		&models.UserLevel{},

		// Museum相关模型
		&models.MuseumItem{},
		&models.MuseumCollection{},
		&models.MuseumExhibitionEntry{},
		&models.MuseumEntry{},
		&models.MuseumExhibition{},

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
	}

	// Use simple AutoMigrate - it should handle existing tables
	err := db.AutoMigrate(modelsToMigrate...)
	if err != nil {
		// If it's just table already exists errors, we can ignore them
		if strings.Contains(err.Error(), "already exists") {
			log.Printf("Some tables already exist during migration, this is normal: %v", err)
		} else {
			log.Printf("Migration error: %v", err)
			return err
		}
	}
	log.Println("Main auto migration completed successfully in autoMigrateDirect")
	return nil
}

// ParsePostgreSQLURL 解析 PostgreSQL URL
func ParsePostgreSQLURL(dbURL string) (host string, port int, user string, password string, dbname string, sslmode string, err error) {
	u, err := url.Parse(dbURL)
	if err != nil {
		return "", 0, "", "", "", "", err
	}

	host = u.Hostname()
	if host == "" {
		host = "localhost"
	}

	portStr := u.Port()
	if portStr == "" {
		port = 5432
	} else {
		port, err = strconv.Atoi(portStr)
		if err != nil {
			return "", 0, "", "", "", "", err
		}
	}

	user = u.User.Username()
	password, _ = u.User.Password()

	// Remove leading slash from path
	dbname = strings.TrimPrefix(u.Path, "/")

	// Parse query parameters
	params := u.Query()
	sslmode = params.Get("sslmode")
	if sslmode == "" {
		sslmode = "require"
	}

	return
}
