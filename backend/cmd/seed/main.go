package main

import (
	"fmt"
	"log"

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

	// 检查是否已有数据
	var userCount int64
	if err := db.Model(&models.User{}).Count(&userCount).Error; err != nil {
		log.Fatal("Failed to count users:", err)
	}

	if userCount > 0 {
		fmt.Printf("Database already has %d users, skipping seed...\n", userCount)
		return
	}

	// 创建管理员用户
	fmt.Println("\nCreating admin user...")
	adminUser := models.User{
		ID:           "admin-user",
		Username:     "admin",
		Email:        "admin@openpenpal.com",
		PasswordHash: "$2a$10$dwSXE/fBcbAJVy0jMZHYI.vFjjUZFYRMPpeAzcgmHd.XqwfqgOrEW", // admin123
		Nickname:     "系统管理员",
		Role:         models.RoleSuperAdmin,
		SchoolCode:   "SYSTEM",
		IsActive:     true,
	}

	if err := db.Create(&adminUser).Error; err != nil {
		log.Fatal("Failed to create admin user:", err)
	}
	fmt.Println("✅ Admin user created!")

	// 创建测试用户
	fmt.Println("\nCreating test users...")
	testUsers := []models.User{
		{
			ID:           "test-user-alice",
			Username:     "alice",
			Email:        "alice@example.com",
			PasswordHash: "$2a$10$N9qo8uLOickgx2ZMRZoMye.Lf/rSDRfYHYxX1dpIjJJNzTHFN1UTO", // secret
			Nickname:     "Alice",
			Role:         models.RoleUser,
			SchoolCode:   "BJDX01",
			IsActive:     true,
		},
		{
			ID:           "test-user-bob",
			Username:     "bob",
			Email:        "bob@example.com",
			PasswordHash: "$2a$10$N9qo8uLOickgx2ZMRZoMye.Lf/rSDRfYHYxX1dpIjJJNzTHFN1UTO", // secret
			Nickname:     "Bob",
			Role:         models.RoleUser,
			SchoolCode:   "BJDX01",
			IsActive:     true,
		},
	}

	for _, user := range testUsers {
		if err := db.Create(&user).Error; err != nil {
			log.Printf("Failed to create user %s: %v", user.Username, err)
			continue
		}
		fmt.Printf("✅ Created user: %s\n", user.Username)
	}

	// 创建测试信件
	fmt.Println("\nCreating test letters...")
	testLetters := []models.Letter{
		{
			ID:      "test-letter-1",
			UserID:  "test-user-alice",
			Title:   "Welcome to OpenPenPal",
			Content: "This is a test letter to demonstrate the system.",
			Style:   models.StyleClassic,
			Status:  models.StatusDraft,
		},
		{
			ID:      "test-letter-2",
			UserID:  "test-user-bob",
			Title:   "Hello from Bob",
			Content: "Testing the letter system with PostgreSQL!",
			Style:   models.StyleModern,
			Status:  models.StatusGenerated,
		},
	}

	for _, letter := range testLetters {
		if err := db.Create(&letter).Error; err != nil {
			log.Printf("Failed to create letter %s: %v", letter.Title, err)
			continue
		}
		fmt.Printf("✅ Created letter: %s\n", letter.Title)
	}

	fmt.Println("\n✨ Seed data created successfully!")

	// 显示统计
	db.Model(&models.User{}).Count(&userCount)
	var letterCount int64
	db.Model(&models.Letter{}).Count(&letterCount)

	fmt.Printf("\nDatabase now contains:\n")
	fmt.Printf("  - %d users\n", userCount)
	fmt.Printf("  - %d letters\n", letterCount)
}
