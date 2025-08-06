package main

import (
	"fmt"
	"log"
	"os"

	"openpenpal-backend/internal/config"
	"openpenpal-backend/internal/models"

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

	fmt.Println("=== 数据库连接测试 ===")
	fmt.Printf("数据库类型: %s\n", cfg.DatabaseType)
	
	if cfg.DatabaseType == "postgres" {
		fmt.Printf("数据库主机: %s\n", cfg.DBHost)
		fmt.Printf("数据库端口: %s\n", cfg.DBPort)
		fmt.Printf("数据库名称: %s\n", cfg.DatabaseName)
		fmt.Printf("数据库用户: %s\n", cfg.DBUser)
	} else {
		fmt.Printf("数据库文件: %s\n", cfg.DatabaseURL)
	}

	// 连接数据库
	fmt.Println("\n正在连接数据库...")
	db, err := config.SetupDatabase(cfg)
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}

	fmt.Println("✅ 数据库连接成功！")

	// 测试查询
	var count int64
	if err := db.Model(&models.User{}).Count(&count).Error; err != nil {
		log.Fatal("查询失败:", err)
	}

	fmt.Printf("\n当前用户数: %d\n", count)

	// 测试创建表
	fmt.Println("\n检查数据库表...")
	
	// 获取数据库连接
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("获取数据库连接失败:", err)
	}
	defer sqlDB.Close()

	// 测试连接
	if err := sqlDB.Ping(); err != nil {
		log.Fatal("Ping 失败:", err)
	}

	fmt.Println("✅ 数据库连接正常！")
	
	// 显示表信息
	var tables []string
	if cfg.DatabaseType == "postgres" {
		// PostgreSQL 查询表
		rows, err := db.Raw("SELECT tablename FROM pg_tables WHERE schemaname = 'public'").Rows()
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var table string
				rows.Scan(&table)
				tables = append(tables, table)
			}
		}
	} else {
		// SQLite 查询表
		rows, err := db.Raw("SELECT name FROM sqlite_master WHERE type='table'").Rows()
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var table string
				rows.Scan(&table)
				tables = append(tables, table)
			}
		}
	}

	fmt.Printf("\n数据库表 (%d 个):\n", len(tables))
	for _, table := range tables {
		fmt.Printf("  - %s\n", table)
	}

	fmt.Println("\n✨ 数据库测试完成！")
	os.Exit(0)
}