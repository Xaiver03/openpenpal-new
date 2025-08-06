package main

import (
	"fmt"
	"log"
	"strconv"

	"openpenpal-backend/internal/config"
	"openpenpal-backend/internal/models"
	"shared/pkg/database"

	"github.com/joho/godotenv"
)

func main() {
	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// 创建数据库配置，不自动迁移
	dbConfig := &database.Config{
		Type:     database.DatabaseType(cfg.DatabaseType),
		Database: cfg.DatabaseURL,
		Host:     cfg.DBHost,
		Username: cfg.DBUser,
		Password: cfg.DBPassword,
		SSLMode:  cfg.DBSSLMode,
	}
	
	// 处理端口号
	if cfg.DBPort != "" {
		if port, err := strconv.Atoi(cfg.DBPort); err == nil {
			dbConfig.Port = port
		}
	}

	// 处理不同数据库类型
	if cfg.DatabaseType == "postgres" || cfg.DatabaseType == "postgresql" {
		dbConfig.Type = database.PostgreSQL
		if cfg.DatabaseURL == "./openpenpal.db" || cfg.DatabaseURL == "" {
			dbConfig.Database = cfg.DatabaseName
		}
	}

	// 使用管理器连接数据库
	manager := database.GetDefaultManager()
	if err := manager.AddConfig("default", dbConfig); err != nil {
		log.Fatal("Failed to add database config:", err)
	}
	
	db, err := manager.Connect("default")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("✅ Connected to database!")

	// 逐个迁移表
	fmt.Println("\nStarting manual migration...")

	// 1. 先迁移基础表（没有外键依赖的）
	basicTables := []struct {
		name  string
		model interface{}
	}{
		{"User", &models.User{}},
		{"Letter", &models.Letter{}},
		{"Courier", &models.Courier{}},
	}

	for _, table := range basicTables {
		fmt.Printf("Migrating %s... ", table.name)
		if err := db.AutoMigrate(table.model); err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			// 继续尝试其他表
			continue
		}
		fmt.Println("✅")
	}

	// 2. 迁移有外键依赖的表
	dependentTables := []struct {
		name  string
		model interface{}
	}{
		{"UserProfile", &models.UserProfile{}},
		{"LetterCode", &models.LetterCode{}},
		{"StatusLog", &models.StatusLog{}},
		{"LetterPhoto", &models.LetterPhoto{}},
		{"CourierTask", &models.CourierTask{}},
	}

	for _, table := range dependentTables {
		fmt.Printf("Migrating %s... ", table.name)
		if err := db.AutoMigrate(table.model); err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			continue
		}
		fmt.Println("✅")
	}

	fmt.Println("\n✨ Migration process completed!")

	// 检查表
	var tables []string
	rows, err := db.Raw("SELECT tablename FROM pg_tables WHERE schemaname = 'public'").Rows()
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var table string
			rows.Scan(&table)
			tables = append(tables, table)
		}
	}

	fmt.Printf("\nCreated tables (%d):\n", len(tables))
	for _, table := range tables {
		fmt.Printf("  - %s\n", table)
	}
}