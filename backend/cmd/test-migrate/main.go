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

	// 连接数据库
	fmt.Println("Connecting to database...")
	db, err := config.SetupDatabase(cfg)
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}

	fmt.Println("✅ Connected successfully!")

	// 逐个迁移表，找出问题
	tables := []struct {
		name  string
		model interface{}
	}{
		{"User", &models.User{}},
		{"UserProfile", &models.UserProfile{}},
		{"Letter", &models.Letter{}},
		{"LetterCode", &models.LetterCode{}},
		{"StatusLog", &models.StatusLog{}},
		{"LetterPhoto", &models.LetterPhoto{}},
		{"Courier", &models.Courier{}},
		{"CourierTask", &models.CourierTask{}},
	}

	for _, table := range tables {
		fmt.Printf("Migrating %s... ", table.name)
		if err := db.AutoMigrate(table.model); err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✅")
	}

	fmt.Println("\n✨ All migrations completed successfully!")
}