package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"openpenpal-backend/internal/config"
	"openpenpal-backend/internal/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// 管理员密码重置工具
// 用于在紧急情况下重置用户密码

func main() {
	var (
		username  = flag.String("user", "", "Username to reset password for")
		password  = flag.String("password", "", "New password (if empty, will generate random)")
		listUsers = flag.Bool("list", false, "List all users")
		dbReset   = flag.Bool("reset-db", false, "Reset all users with secure passwords")
	)
	flag.Parse()

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// 连接数据库
	db, err := config.SetupDatabaseDirect(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	if *listUsers {
		listAllUsers(db)
		return
	}

	if *dbReset {
		resetAllUsersSecurely(db, cfg.BCryptCost)
		return
	}

	if *username == "" {
		fmt.Println("Usage:")
		fmt.Println("  Reset specific user password:")
		fmt.Println("    go run cmd/admin/reset_passwords.go -user=admin -password=newpassword")
		fmt.Println("  Generate random password:")
		fmt.Println("    go run cmd/admin/reset_passwords.go -user=admin")
		fmt.Println("  List all users:")
		fmt.Println("    go run cmd/admin/reset_passwords.go -list")
		fmt.Println("  Reset all users with secure passwords:")
		fmt.Println("    go run cmd/admin/reset_passwords.go -reset-db")
		return
	}

	// 重置单个用户密码
	if err := resetUserPassword(db, *username, *password, cfg.BCryptCost); err != nil {
		log.Fatal("Failed to reset password:", err)
	}
}

func listAllUsers(db *gorm.DB) {
	var users []models.User
	if err := db.Select("id, username, email, role, is_active, created_at").Find(&users).Error; err != nil {
		log.Fatal("Failed to list users:", err)
	}

	fmt.Printf("%-20s %-20s %-30s %-15s %-8s %s\n", "ID", "Username", "Email", "Role", "Active", "Created")
	fmt.Println(strings.Repeat("-", 120))

	for _, user := range users {
		fmt.Printf("%-20s %-20s %-30s %-15s %-8t %s\n",
			user.ID, user.Username, user.Email, user.Role, user.IsActive, user.CreatedAt.Format("2006-01-02"))
	}
}

func resetUserPassword(db *gorm.DB, username, newPassword string, bcryptCost int) error {
	// 如果没有提供密码，生成随机密码
	if newPassword == "" {
		seedManager := config.NewSecureSeedManager(db, bcryptCost)
		if err := seedManager.RegenerateUserPassword(username, ""); err != nil {
			return err
		}
		return nil
	}

	// 验证密码强度
	if len(newPassword) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	// 哈希新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcryptCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// 更新数据库
	result := db.Model(&models.User{}).Where("username = ?", username).Update("password_hash", string(hashedPassword))
	if result.Error != nil {
		return fmt.Errorf("failed to update password: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("user %s not found", username)
	}

	fmt.Printf("✅ Password updated for user: %s\n", username)
	fmt.Printf("🔐 New password: %s\n", newPassword)
	fmt.Printf("⚠️  Please ensure the user changes this password on first login\n")

	return nil
}

func resetAllUsersSecurely(db *gorm.DB, bcryptCost int) {
	fmt.Println("🔐 Resetting ALL users with secure passwords...")
	fmt.Print("⚠️  This will reset passwords for ALL users. Continue? (y/N): ")

	var confirm string
	fmt.Scanln(&confirm)

	if confirm != "y" && confirm != "Y" {
		fmt.Println("Operation cancelled.")
		return
	}

	// 删除所有现有用户
	if err := db.Exec("DELETE FROM users").Error; err != nil {
		log.Fatal("Failed to clear users table:", err)
	}

	// 使用安全种子管理器重新创建
	seedManager := config.NewSecureSeedManager(db, bcryptCost)
	if err := seedManager.SecureSeedData(); err != nil {
		log.Fatal("Failed to create secure seed data:", err)
	}

	fmt.Println("✅ All users reset with secure passwords!")
}
