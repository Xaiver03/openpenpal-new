package config

import (
	"courier-service/internal/models"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Config struct {
	Port         string
	DatabaseURL  string
	RedisURL     string
	JWTSecret    string
	Environment  string
	WebSocketURL string
}

func Load() *Config {
	// 加载 .env 文件
	godotenv.Load()

	return &Config{
		Port:         getEnv("PORT", "8002"),
		DatabaseURL:  getEnv("DATABASE_URL", "postgres://rocalight:password@localhost:5432/openpenpal?sslmode=disable"),
		RedisURL:     getEnv("REDIS_URL", "redis://localhost:6379/0"),
		JWTSecret:    getEnv("JWT_SECRET", "your-super-secret-jwt-key"),
		Environment:  getEnv("ENVIRONMENT", "development"),
		WebSocketURL: getEnv("WEBSOCKET_URL", "ws://localhost:8080/ws"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func InitDatabase(databaseURL string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		// 跳过默认事务，避免视图约束问题
		SkipDefaultTransaction: true,
		// 不在迁移时创建约束
		DisableAutomaticPing: false,
	})
	if err != nil {
		return nil, err
	}

	// 自动迁移数据库表
	err = autoMigrate(db)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func autoMigrate(db *gorm.DB) error {
	// 检查couriers表是否已存在
	if db.Migrator().HasTable("couriers") {
		// 表已存在，只添加缺失的列，不修改现有列类型
		// 这避免了视图依赖的列类型修改错误
		type Column struct {
			Name string
			Type string
		}

		// 需要检查的新列
		newColumns := []string{"zone_code", "zone_type", "parent_id", "created_by_id", "points"}

		for _, col := range newColumns {
			if !db.Migrator().HasColumn(&models.Courier{}, col) {
				// 使用原生SQL添加列，避免GORM尝试修改其他列
				var addColumnSQL string
				switch col {
				case "zone_code", "zone_type":
					addColumnSQL = `ALTER TABLE couriers ADD COLUMN IF NOT EXISTS ` + col + ` VARCHAR(255)`
				case "parent_id", "created_by_id":
					addColumnSQL = `ALTER TABLE couriers ADD COLUMN IF NOT EXISTS ` + col + ` VARCHAR(36)`
				case "points":
					addColumnSQL = `ALTER TABLE couriers ADD COLUMN IF NOT EXISTS ` + col + ` INTEGER DEFAULT 0`
				}

				if addColumnSQL != "" {
					if err := db.Exec(addColumnSQL).Error; err != nil {
						// 忽略列已存在的错误
						if !strings.Contains(err.Error(), "already exists") {
							return err
						}
					}
				}
			}
		}
	} else {
		// 表不存在，创建新表
		err := db.AutoMigrate(&models.Courier{})
		if err != nil {
			return err
		}
	}

	// 迁移其他模型
	err := db.AutoMigrate(
		&models.CourierLevelModel{},
		&models.CourierPermissionModel{},
		&models.CourierBadge{},
		&models.CourierPoints{},
		&models.SignalCode{},
		&models.SignalCodeBatch{},
		&models.SignalCodeRule{},
		&models.PostalCodeRule{},
		&models.PostalCodeZone{},
		&models.CourierZone{},
		&models.LevelUpgradeRequest{},
		&models.CourierGrowthPath{},
		&models.CourierIncentive{},
		&models.CourierStatistics{},
		&models.CourierPointsTransaction{},
		&models.CourierBadgeEarned{},
		&models.PostalCodeApplication{},
		&models.PostalCodeAssignment{},
		&models.SignalCodeUsageLog{},
		&models.CourierRanking{},
		&models.CourierPointsHistory{},
		&models.TaskAssignmentHistory{},
		&models.Task{},
		&models.ScanRecord{},
	)
	if err != nil {
		return err
	}

	return nil
}
