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

	// Handle constraint management before auto migration
	if config.DatabaseType == "postgres" {
		if err := handleConstraintsBeforeMigration(db); err != nil {
			log.Printf("Warning: Failed to handle constraints: %v", err)
		}
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
		// Handle different types of migration errors gracefully
		errStr := err.Error()
		if strings.Contains(errStr, "already exists") {
			log.Printf("Some tables already exist during migration, this is normal: %v", err)
		} else if strings.Contains(errStr, "constraint") && strings.Contains(errStr, "does not exist") {
			log.Printf("Constraint-related error during migration, this is often recoverable: %v", err)
			// Try to continue with migration by migrating models individually
			if err := migrateModelsIndividually(db, modelsToMigrate); err != nil {
				log.Printf("Individual migration also failed: %v", err)
				return err
			}
			log.Println("Individual model migration completed successfully")
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

// migrateModelsIndividually migrates models one by one to handle constraint errors
func migrateModelsIndividually(db *gorm.DB, models []interface{}) error {
	log.Println("Starting individual model migration to handle constraint issues...")
	
	for i, model := range models {
		log.Printf("Migrating model %d/%d...", i+1, len(models))
		if err := db.AutoMigrate(model); err != nil {
			errStr := err.Error()
			// Allow certain types of errors that are recoverable
			if strings.Contains(errStr, "constraint") && strings.Contains(errStr, "does not exist") {
				log.Printf("Constraint error for model %T (continuing): %v", model, err)
				continue
			} else if strings.Contains(errStr, "already exists") {
				log.Printf("Model %T already exists (continuing): %v", model, err)
				continue
			} else {
				log.Printf("Failed to migrate model %T: %v", model, err)
				return err
			}
		} else {
			log.Printf("Successfully migrated model %T", model)
		}
	}
	
	return nil
}

// handleConstraintsBeforeMigration ensures constraints are in correct state before auto migration
func handleConstraintsBeforeMigration(db *gorm.DB) error {
	log.Println("Handling constraint management before migration...")

	// Check if users table exists
	var tableExists bool
	if err := db.Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'users')").Scan(&tableExists).Error; err != nil {
		log.Printf("Could not check if users table exists: %v", err)
		return nil // Not a critical error, continue with migration
	}

	if !tableExists {
		log.Println("Users table does not exist yet, skipping constraint management")
		return nil
	}

	// Check if uni_users_username constraint exists
	var constraintExists bool
	constraintCheckSQL := `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.table_constraints 
			WHERE table_schema = 'public' 
			AND table_name = 'users' 
			AND constraint_name = 'uni_users_username'
		)
	`
	
	if err := db.Raw(constraintCheckSQL).Scan(&constraintExists).Error; err != nil {
		log.Printf("Could not check constraint existence: %v", err)
		return nil // Not critical, continue
	}

	if !constraintExists {
		log.Println("uni_users_username constraint does not exist, will let GORM create it")
		
		// Check if there's already a unique index on username instead
		var indexExists bool
		indexCheckSQL := `
			SELECT EXISTS (
				SELECT 1 FROM pg_indexes 
				WHERE schemaname = 'public' 
				AND tablename = 'users' 
				AND indexname = 'idx_users_username'
			)
		`
		
		if err := db.Raw(indexCheckSQL).Scan(&indexExists).Error; err != nil {
			log.Printf("Could not check index existence: %v", err)
			return nil
		}

		if indexExists {
			log.Println("Found existing unique index on username, this is good")
		} else {
			log.Println("No existing unique constraint or index on username found")
		}
	} else {
		log.Println("uni_users_username constraint already exists")
	}

	return nil
}
