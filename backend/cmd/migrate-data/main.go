package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"openpenpal-backend/internal/config"
	"openpenpal-backend/internal/models"

	"gorm.io/driver/postgres"
	// TODO: Keep SQLite import for migration tool when needed
	// "gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TODO: Re-enable migration tool when needed
// This tool migrates data from SQLite to PostgreSQL
func main() {
	fmt.Println("=== SQLite to PostgreSQL Migration Tool ===")
	fmt.Println("This tool is currently disabled as we are using PostgreSQL-only setup.")
	fmt.Println("To enable this tool, uncomment the SQLite driver import and implementation.")
	return

	/*
	if len(os.Args) < 2 {
		fmt.Println("使用方法: go run main.go [sqlite-file]")
		fmt.Println("示例: go run main.go ../openpenpal.db")
		os.Exit(1)
	}

	sqliteFile := os.Args[1]

	// 检查 SQLite 文件是否存在
	if _, err := os.Stat(sqliteFile); os.IsNotExist(err) {
		log.Fatal("SQLite 文件不存在:", sqliteFile)
	}

	fmt.Println("=== SQLite 到 PostgreSQL 数据迁移 ===")
	fmt.Printf("源数据库: %s\n", sqliteFile)

	// 连接 SQLite
	sqliteDB, err := gorm.Open(sqlite.Open(sqliteFile), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatal("连接 SQLite 失败:", err)
	}

	// 加载配置并连接 PostgreSQL
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("加载配置失败:", err)
	}

	if cfg.DatabaseType != "postgres" {
		log.Fatal("请设置 DATABASE_TYPE=postgres")
	}

	// 构建 PostgreSQL DSN
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DatabaseName, cfg.DBPort, cfg.DBSSLMode)

	postgresDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatal("连接 PostgreSQL 失败:", err)
	}

	fmt.Println("✅ 数据库连接成功")

	// 开始迁移
	fmt.Println("\n开始迁移数据...")

	// 1. 迁移用户
	fmt.Print("迁移用户数据... ")
	if err := migrateTable(sqliteDB, postgresDB, &models.User{}, "users"); err != nil {
		log.Fatal("迁移用户失败:", err)
	}

	// 2. 迁移用户档案
	fmt.Print("迁移用户档案... ")
	if err := migrateTable(sqliteDB, postgresDB, &models.UserProfile{}, "user_profiles"); err != nil {
		log.Printf("迁移用户档案失败: %v", err)
	}

	// 3. 迁移信件
	fmt.Print("迁移信件数据... ")
	if err := migrateTable(sqliteDB, postgresDB, &models.Letter{}, "letters"); err != nil {
		log.Fatal("迁移信件失败:", err)
	}

	// 4. 迁移信件编码
	fmt.Print("迁移信件编码... ")
	if err := migrateTable(sqliteDB, postgresDB, &models.LetterCode{}, "letter_codes"); err != nil {
		log.Fatal("迁移信件编码失败:", err)
	}

	// 5. 迁移信使
	fmt.Print("迁移信使数据... ")
	if err := migrateTable(sqliteDB, postgresDB, &models.Courier{}, "couriers"); err != nil {
		log.Printf("迁移信使失败: %v", err)
	}

	// 6. 迁移信使任务
	fmt.Print("迁移信使任务... ")
	if err := migrateTable(sqliteDB, postgresDB, &models.CourierTask{}, "courier_tasks"); err != nil {
		log.Printf("迁移信使任务失败: %v", err)
	}

	// 7. 迁移用户积分
	fmt.Print("迁移用户积分... ")
	if err := migrateTable(sqliteDB, postgresDB, &models.UserCredit{}, "user_credits"); err != nil {
		log.Printf("迁移用户积分失败: %v", err)
	}

	// 8. 迁移博物馆条目
	fmt.Print("迁移博物馆条目... ")
	if err := migrateTable(sqliteDB, postgresDB, &models.MuseumEntry{}, "museum_entries"); err != nil {
		log.Printf("迁移博物馆条目失败: %v", err)
	}

	// 9. 迁移博物馆二维码 - 注释掉因为模型不存在
	// fmt.Print("迁移博物馆二维码... ")
	// if err := migrateTable(sqliteDB, postgresDB, &models.MuseumQRCode{}, "museum_qr_codes"); err != nil {
	// 	log.Printf("迁移博物馆二维码失败: %v", err)
	// }

	fmt.Println("\n✨ 数据迁移完成！")

	// 显示统计信息
	showStatistics(postgresDB)
}

func migrateTable(source, dest *gorm.DB, model interface{}, tableName string) error {
	// 检查源表是否存在
	var count int64
	source.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", tableName).Scan(&count)
	if count == 0 {
		fmt.Printf("跳过（表不存在）\n")
		return nil
	}

	// 获取记录数
	var totalCount int64
	if err := source.Table(tableName).Count(&totalCount).Error; err != nil {
		return err
	}

	if totalCount == 0 {
		fmt.Printf("跳过（0 条记录）\n")
		return nil
	}

	// 批量迁移
	batchSize := 100
	for offset := 0; offset < int(totalCount); offset += batchSize {
		var records []map[string]interface{}

		if err := source.Table(tableName).
			Offset(offset).
			Limit(batchSize).
			Find(&records).Error; err != nil {
			return err
		}

		// 处理时间字段
		for _, record := range records {
			processTimeFields(record)
		}

		// 插入到目标数据库
		if len(records) > 0 {
			if err := dest.Table(tableName).Create(&records).Error; err != nil {
				// 尝试单条插入以找出问题记录
				for _, record := range records {
					if err := dest.Table(tableName).Create(&record).Error; err != nil {
						log.Printf("插入记录失败: %v, 记录: %v", err, record["id"])
					}
				}
			}
		}
	}

	fmt.Printf("✓ %d 条记录\n", totalCount)
	return nil
}

func processTimeFields(record map[string]interface{}) {
	timeFields := []string{"created_at", "updated_at", "deleted_at", "last_login_at", "published_at"}

	for _, field := range timeFields {
		if val, ok := record[field]; ok && val != nil {
			switch v := val.(type) {
			case string:
				if v != "" && v != "0000-00-00 00:00:00" {
					if t, err := time.Parse("2006-01-02 15:04:05", v); err == nil {
						record[field] = t
					} else if t, err := time.Parse(time.RFC3339, v); err == nil {
						record[field] = t
					}
				} else {
					record[field] = nil
				}
			case time.Time:
				if v.IsZero() {
					record[field] = nil
				}
			}
		}
	}
}

func showStatistics(db *gorm.DB) {
	fmt.Println("\n=== 迁移统计 ===")

	tables := map[string]string{
		"users":          "用户",
		"letters":        "信件",
		"letter_codes":   "信件编码",
		"couriers":       "信使",
		"courier_tasks":  "信使任务",
		"museum_entries": "博物馆条目",
	}

	for table, name := range tables {
		var count int64
		db.Table(table).Count(&count)
		if count > 0 {
			fmt.Printf("%s: %d 条\n", name, count)
		}
	}
	*/
}
