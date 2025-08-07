package services

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"openpenpal-backend/internal/config"
	"openpenpal-backend/internal/models"
	"openpenpal-backend/pkg/auth"
	"openpenpal-backend/pkg/utils"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	db     *gorm.DB
	config *config.Config
}

func NewUserService(db *gorm.DB, config *config.Config) *UserService {
	return &UserService{
		db:     db,
		config: config,
	}
}

// GetDB returns the database instance
func (s *UserService) GetDB() *gorm.DB {
	return s.db
}

// Register 用户注册
func (s *UserService) Register(req *models.RegisterRequest) (*models.User, error) {
	// 检查用户名和邮箱是否已存在
	var existingUser models.User
	if err := s.db.Where("username = ? OR email = ?", req.Username, req.Email).First(&existingUser).Error; err == nil {
		return nil, fmt.Errorf("username or email already exists")
	}

	// 验证学校代码
	if !utils.IsValidSchoolCode(req.SchoolCode) {
		return nil, fmt.Errorf("invalid school code")
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), s.config.BCryptCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 创建用户
	user := &models.User{
		ID:           uuid.New().String(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Nickname:     req.Nickname,
		Role:         models.RoleUser,
		SchoolCode:   req.SchoolCode,
		IsActive:     true,
	}

	if err := s.db.Create(user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// 清除密码哈希，不返回给客户端
	user.PasswordHash = ""
	return user, nil
}

// getSystemConfig 获取系统配置
func (s *UserService) getSystemConfig() (*models.SystemConfig, error) {
	var config models.SystemConfig
	// 尝试从数据库获取配置（系统配置表只有一条记录）
	err := s.db.First(&config).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 如果没有配置，创建默认配置
			config = *models.DefaultSystemConfig()
			if err := s.db.Create(&config).Error; err != nil {
				return nil, err
			}
			return &config, nil
		}
		return nil, err
	}
	return &config, nil
}

// GetSystemConfig 公开方法获取系统配置
func (s *UserService) GetSystemConfig() (*models.SystemConfig, error) {
	return s.getSystemConfig()
}

// UpdateLastActivity 更新用户最后活动时间
func (s *UserService) UpdateLastActivity(userID string, lastActivity *time.Time) error {
	return s.db.Model(&models.User{}).Where("id = ?", userID).Update("last_login_at", lastActivity).Error
}

// Login 用户登录
func (s *UserService) Login(req *models.LoginRequest) (*models.LoginResponse, error) {
	// 查找用户
	var user models.User
	if err := s.db.Where("username = ? OR email = ?", req.Username, req.Username).First(&user).Error; err != nil {
		return nil, fmt.Errorf("invalid username or password")
	}

	// 检查用户是否激活
	if !user.IsActive {
		return nil, fmt.Errorf("user account is disabled")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("invalid username or password")
	}

	// 从系统配置获取JWT过期时间
	systemConfig, err := s.getSystemConfig()
	if err != nil {
		// 如果获取配置失败，使用默认值
		systemConfig = models.DefaultSystemConfig()
	}
	
	// 生成JWT令牌
	expiryHours := systemConfig.JWTExpiryHours
	if expiryHours <= 0 {
		expiryHours = 24 // 默认24小时
	}
	expiresAt := time.Now().Add(time.Duration(expiryHours) * time.Hour)
	token, err := auth.GenerateJWT(user.ID, user.Role, s.config.JWTSecret, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// 更新最后登录时间
	now := time.Now()
	s.db.Model(&user).Update("last_login_at", &now)

	// 清除密码哈希
	user.PasswordHash = ""

	return &models.LoginResponse{
		Token:     token,
		User:      &user,
		ExpiresAt: expiresAt,
	}, nil
}

// GetUserByID 通过ID获取用户
func (s *UserService) GetUserByID(userID string) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, "id = ?", userID).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	user.PasswordHash = ""
	return &user, nil
}

// UpdateProfile 更新用户档案
func (s *UserService) UpdateProfile(userID string, req *models.UpdateProfileRequest) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, "id = ?", userID).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// 更新字段
	updates := make(map[string]interface{})
	if req.Nickname != "" {
		updates["nickname"] = req.Nickname
	}
	if req.Avatar != "" {
		updates["avatar"] = req.Avatar
	}

	if len(updates) > 0 {
		if err := s.db.Model(&user).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("failed to update profile: %w", err)
		}
	}

	// 更新或创建用户档案
	var profile models.UserProfile
	if err := s.db.First(&profile, "user_id = ?", userID).Error; err != nil {
		// 创建新档案
		profile = models.UserProfile{
			UserID:  userID,
			Bio:     req.Bio,
			Address: req.Address,
		}
		if err := s.db.Create(&profile).Error; err != nil {
			return nil, fmt.Errorf("failed to create profile: %w", err)
		}
	} else {
		// 更新现有档案
		profileUpdates := make(map[string]interface{})
		if req.Bio != "" {
			profileUpdates["bio"] = req.Bio
		}
		if req.Address != "" {
			profileUpdates["address"] = req.Address
		}
		if len(profileUpdates) > 0 {
			if err := s.db.Model(&profile).Updates(profileUpdates).Error; err != nil {
				return nil, fmt.Errorf("failed to update profile: %w", err)
			}
		}
	}

	user.PasswordHash = ""
	return &user, nil
}

// ChangePassword 修改密码
func (s *UserService) ChangePassword(userID string, req *models.ChangePasswordRequest) error {
	var user models.User
	if err := s.db.First(&user, "id = ?", userID).Error; err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword)); err != nil {
		return fmt.Errorf("old password is incorrect")
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), s.config.BCryptCost)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// 更新密码
	if err := s.db.Model(&user).Update("password_hash", string(hashedPassword)).Error; err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// GetUserStats 获取用户统计
func (s *UserService) GetUserStats(userID string) (*models.UserStats, error) {
	stats := &models.UserStats{}

	// 发送的信件数量
	if err := s.db.Model(&models.Letter{}).Where("user_id = ? AND status != ?",
		userID, models.StatusDraft).Count(&stats.LettersSent).Error; err != nil {
		return nil, fmt.Errorf("failed to count sent letters: %w", err)
	}

	// 收到的信件数量
	if err := s.db.Model(&models.Letter{}).Where("reply_to = ?",
		userID).Count(&stats.LettersReceived).Error; err != nil {
		return nil, fmt.Errorf("failed to count received letters: %w", err)
	}

	// 草稿数量
	if err := s.db.Model(&models.Letter{}).Where("user_id = ? AND status = ?",
		userID, models.StatusDraft).Count(&stats.DraftsCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count drafts: %w", err)
	}

	// 已送达数量
	if err := s.db.Model(&models.Letter{}).Where("user_id = ? AND status IN ?",
		userID, []models.LetterStatus{models.StatusDelivered, models.StatusRead}).
		Count(&stats.DeliveredCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count delivered letters: %w", err)
	}

	return stats, nil
}

// DeactivateUser 停用用户
func (s *UserService) DeactivateUser(userID string) error {
	if err := s.db.Model(&models.User{}).Where("id = ?", userID).
		Update("is_active", false).Error; err != nil {
		return fmt.Errorf("failed to deactivate user: %w", err)
	}
	return nil
}

// ReactivateUser 重新激活用户
func (s *UserService) ReactivateUser(userID string) error {
	if err := s.db.Model(&models.User{}).Where("id = ?", userID).
		Update("is_active", true).Error; err != nil {
		return fmt.Errorf("failed to reactivate user: %w", err)
	}
	return nil
}

// SaveAvatar 保存用户头像
func (s *UserService) SaveAvatar(userID string, file multipart.File, filename string) (string, error) {
	// 创建上传目录
	uploadDir := filepath.Join("uploads", "avatars")
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %w", err)
	}

	// 获取旧头像 URL
	var user models.User
	if err := s.db.Select("avatar").Where("id = ?", userID).First(&user).Error; err != nil {
		return "", fmt.Errorf("failed to get user: %w", err)
	}

	// 如果存在旧头像，删除旧文件
	if user.Avatar != "" {
		oldPath := filepath.Join(".", user.Avatar)
		_ = os.Remove(oldPath) // 忽略删除错误
	}

	// 保存新文件
	filePath := filepath.Join(uploadDir, filename)
	dstFile, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, file); err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	// 更新数据库
	avatarURL := "/" + filepath.ToSlash(filePath)
	if err := s.db.Model(&models.User{}).Where("id = ?", userID).
		Update("avatar", avatarURL).Error; err != nil {
		// 如果数据库更新失败，删除已上传的文件
		_ = os.Remove(filePath)
		return "", fmt.Errorf("failed to update avatar: %w", err)
	}

	return avatarURL, nil
}

// RemoveAvatar 移除用户头像
func (s *UserService) RemoveAvatar(userID string) error {
	// 获取当前头像 URL
	var user models.User
	if err := s.db.Select("avatar").Where("id = ?", userID).First(&user).Error; err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// 删除文件
	if user.Avatar != "" {
		filePath := filepath.Join(".", user.Avatar)
		_ = os.Remove(filePath) // 忽略删除错误
	}

	// 清空数据库中的头像字段
	if err := s.db.Model(&models.User{}).Where("id = ?", userID).
		Update("avatar", "").Error; err != nil {
		return fmt.Errorf("failed to update avatar: %w", err)
	}

	return nil
}
