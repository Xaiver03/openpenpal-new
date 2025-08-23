package config

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"os"

	"openpenpal-backend/internal/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// SecureSeedManager 安全种子数据管理器
type SecureSeedManager struct {
	db         *gorm.DB
	bcryptCost int
}

// NewSecureSeedManager 创建安全种子数据管理器
func NewSecureSeedManager(db *gorm.DB, bcryptCost int) *SecureSeedManager {
	return &SecureSeedManager{
		db:         db,
		bcryptCost: bcryptCost,
	}
}

// generateSecurePassword 生成安全随机密码
func (s *SecureSeedManager) generateSecurePassword(length int) (string, error) {
	// 生成随机字节
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// 使用base64编码生成可读密码
	password := base64.URLEncoding.EncodeToString(bytes)[:length]
	return password, nil
}

// hashPassword 安全哈希密码
func (s *SecureSeedManager) hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), s.bcryptCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hash), nil
}

// UserCredential 用户凭据结构
type UserCredential struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
	Email    string `json:"email"`
}

// SecureSeedData 安全初始化数据（动态生成密码）
func (s *SecureSeedManager) SecureSeedData() error {
	// 检查是否已有数据
	var userCount int64
	if err := s.db.Model(&models.User{}).Count(&userCount).Error; err != nil {
		return err
	}

	if userCount > 0 {
		log.Println("Database already seeded, skipping secure seed...")
		return nil
	}

	log.Println("🔐 Starting SECURE seed data generation...")

	// 从环境变量获取初始管理员密码
	adminPassword := os.Getenv("INITIAL_ADMIN_PASSWORD")
	if adminPassword == "" {
		// 检查环境：开发环境使用固定密码，生产环境生成随机密码
		env := os.Getenv("ENVIRONMENT")
		if env == "development" || env == "" {
			adminPassword = "admin123" // 开发环境默认密码
			log.Printf("⚠️  Using default development admin password: %s", adminPassword)
		} else {
			// 生产环境生成安全密码
			var err error
			adminPassword, err = s.generateSecurePassword(16)
			if err != nil {
				return fmt.Errorf("failed to generate admin password: %w", err)
			}
			log.Printf("🔐 Generated secure admin password: %s", adminPassword)
			log.Printf("⚠️  IMPORTANT: Save this password! Set INITIAL_ADMIN_PASSWORD env var to override.")
		}
	}

	// 从环境变量获取测试用户密码
	testPassword := os.Getenv("TEST_USER_PASSWORD")
	if testPassword == "" {
		testPassword = "secret123" // 默认测试密码，满足8位字符要求
		log.Printf("⚠️  Using default test password: %s", testPassword)
	}

	// 用户凭据配置
	userConfigs := []struct {
		ID       string
		Username string
		Email    string
		Password string
		Nickname string
		Role     models.UserRole
		School   string
	}{
		// 超级管理员 - 在开发环境中使用固定密码，生产环境使用强密码
		{"admin-super", "admin", "admin@openpenpal.com", adminPassword, "系统管理员", models.RoleSuperAdmin, "SYSTEM"},

		// 四级信使系统 - 使用环境变量或生成的密码
		{"courier-l1", "courier_level1", "courier1@openpenpal.com", testPassword, "一级信使", models.RoleCourierLevel1, "PKU001"},
		{"courier-l2", "courier_level2", "courier2@openpenpal.com", testPassword, "二级信使", models.RoleCourierLevel2, "PKU001"},
		{"courier-l3", "courier_level3", "courier3@openpenpal.com", testPassword, "三级信使", models.RoleCourierLevel3, "PKU001"},
		{"courier-l4", "courier_level4", "courier4@openpenpal.com", testPassword, "四级信使", models.RoleCourierLevel4, "PKU001"},

		// 测试用户
		{"user-alice", "alice", "alice@openpenpal.com", testPassword, "Alice", models.RoleUser, "PKU001"},
		{"user-bob", "bob", "bob@openpenpal.com", testPassword, "Bob", models.RoleUser, "PKU001"},
	}

	var createdCredentials []UserCredential

	for _, config := range userConfigs {
		// 动态哈希每个密码
		hashedPassword, err := s.hashPassword(config.Password)
		if err != nil {
			return fmt.Errorf("failed to hash password for %s: %w", config.Username, err)
		}

		user := models.User{
			ID:           config.ID,
			Username:     config.Username,
			Email:        config.Email,
			PasswordHash: hashedPassword,
			Nickname:     config.Nickname,
			Role:         config.Role,
			SchoolCode:   config.School,
			IsActive:     true,
		}

		if err := s.db.Create(&user).Error; err != nil {
			return fmt.Errorf("failed to create user %s: %w", config.Username, err)
		}

		// 记录凭据（仅用于开发环境输出）
		createdCredentials = append(createdCredentials, UserCredential{
			Username: config.Username,
			Password: config.Password,
			Role:     string(config.Role),
			Email:    config.Email,
		})

		log.Printf("✅ Created user: %s (role: %s)", config.Username, config.Role)
	}

	// 仅在开发环境输出凭据信息
	env := os.Getenv("ENVIRONMENT")
	if env == "development" || env == "" {
		log.Println("\n🔐 === DEVELOPMENT CREDENTIALS ===")
		for _, cred := range createdCredentials {
			log.Printf("Username: %-20s Password: %-20s Role: %s", cred.Username, cred.Password, cred.Role)
		}
		log.Println("🔐 ================================\n")
	}

	log.Printf("✅ Secure seed completed: %d users created with dynamic password hashing", len(createdCredentials))
	return nil
}

// RegenerateUserPassword 重新生成用户密码（管理员功能）
func (s *SecureSeedManager) RegenerateUserPassword(username string, newPassword string) error {
	if newPassword == "" {
		var err error
		newPassword, err = s.generateSecurePassword(12)
		if err != nil {
			return fmt.Errorf("failed to generate new password: %w", err)
		}
	}

	hashedPassword, err := s.hashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	result := s.db.Model(&models.User{}).Where("username = ?", username).Update("password_hash", hashedPassword)
	if result.Error != nil {
		return fmt.Errorf("failed to update password for %s: %w", username, result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("user %s not found", username)
	}

	log.Printf("🔐 Password updated for user: %s (new password: %s)", username, newPassword)
	return nil
}

// ValidatePasswordStrength 验证密码强度
func (s *SecureSeedManager) ValidatePasswordStrength(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	// 可以添加更多密码强度检查
	// 例如：包含大小写字母、数字、特殊字符等

	return nil
}

// GetBCryptCost 获取当前bcrypt成本
func (s *SecureSeedManager) GetBCryptCost() int {
	return s.bcryptCost
}

// SetBCryptCost 设置bcrypt成本（需要重启生效）
func (s *SecureSeedManager) SetBCryptCost(cost int) error {
	if cost < 10 || cost > 15 {
		return fmt.Errorf("bcrypt cost must be between 10 and 15, got %d", cost)
	}
	s.bcryptCost = cost
	return nil
}
