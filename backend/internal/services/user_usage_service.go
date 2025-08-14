package services

import (
	"fmt"
	"openpenpal-backend/internal/config"
	"openpenpal-backend/internal/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserUsageService 用户使用量服务
type UserUsageService struct {
	db     *gorm.DB
	config *config.Config
}

// UserDailyUsage 用户每日使用量记录
type UserDailyUsage struct {
	ID                 string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	UserID             string    `json:"user_id" gorm:"type:varchar(36);not null;index"`
	Date               time.Time `json:"date" gorm:"type:date;not null;index"`
	InspirationsUsed   int       `json:"inspirations_used" gorm:"default:0"`
	AIRepliesGenerated int       `json:"ai_replies_generated" gorm:"default:0"`
	PenpalMatches      int       `json:"penpal_matches" gorm:"default:0"`
	LettersCurated     int       `json:"letters_curated" gorm:"default:0"`
	CreatedAt          time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt          time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// UserUsageLimits 用户使用限制配置
type UserUsageLimits struct {
	DailyInspirations int `json:"daily_inspirations"`
	DailyAIReplies    int `json:"daily_ai_replies"`
	DailyMatches      int `json:"daily_matches"`
	DailyCurations    int `json:"daily_curations"`
}

// DefaultUsageLimits 默认使用限制（符合PRD要求）
var DefaultUsageLimits = UserUsageLimits{
	DailyInspirations: 2,  // 每日最多2条灵感，符合PRD要求
	DailyAIReplies:    5,  // 每日最多5条AI回信
	DailyMatches:      3,  // 每日最多3次匹配
	DailyCurations:    10, // 每日最多10次策展
}

// NewUserUsageService 创建用户使用量服务实例
func NewUserUsageService(db *gorm.DB, config *config.Config) *UserUsageService {
	service := &UserUsageService{
		db:     db,
		config: config,
	}

	// 自动迁移数据表
	db.AutoMigrate(&UserDailyUsage{})

	return service
}

// GetUserDailyUsage 获取用户今日使用量
func (s *UserUsageService) GetUserDailyUsage(userID string) (*UserDailyUsage, error) {
	today := time.Now().Truncate(24 * time.Hour)

	var usage UserDailyUsage
	err := s.db.Where("user_id = ? AND date = ?", userID, today).First(&usage).Error

	if err != nil && err == gorm.ErrRecordNotFound {
		// 如果没有记录，创建新的
		usage = UserDailyUsage{
			ID:                 uuid.New().String(),
			UserID:             userID,
			Date:               today,
			InspirationsUsed:   0,
			AIRepliesGenerated: 0,
			PenpalMatches:      0,
			LettersCurated:     0,
		}
		err = s.db.Create(&usage).Error
	}

	return &usage, err
}

// CanUseInspiration 检查用户是否可以使用灵感功能
func (s *UserUsageService) CanUseInspiration(userID string) (bool, error) {
	usage, err := s.GetUserDailyUsage(userID)
	if err != nil {
		return false, err
	}

	return usage.InspirationsUsed < DefaultUsageLimits.DailyInspirations, nil
}

// UseInspiration 记录用户使用了一次灵感
func (s *UserUsageService) UseInspiration(userID string) error {
	usage, err := s.GetUserDailyUsage(userID)
	if err != nil {
		return err
	}

	if usage.InspirationsUsed >= DefaultUsageLimits.DailyInspirations {
		return fmt.Errorf("daily inspiration limit exceeded")
	}

	usage.InspirationsUsed++
	return s.db.Save(usage).Error
}

// CanUseAIReply 检查用户是否可以使用AI回信功能
func (s *UserUsageService) CanUseAIReply(userID string) (bool, error) {
	usage, err := s.GetUserDailyUsage(userID)
	if err != nil {
		return false, err
	}

	return usage.AIRepliesGenerated < DefaultUsageLimits.DailyAIReplies, nil
}

// UseAIReply 记录用户使用了一次AI回信
func (s *UserUsageService) UseAIReply(userID string) error {
	usage, err := s.GetUserDailyUsage(userID)
	if err != nil {
		return err
	}

	if usage.AIRepliesGenerated >= DefaultUsageLimits.DailyAIReplies {
		return fmt.Errorf("daily AI reply limit exceeded")
	}

	usage.AIRepliesGenerated++
	return s.db.Save(usage).Error
}

// CanUsePenpalMatch 检查用户是否可以使用笔友匹配功能
func (s *UserUsageService) CanUsePenpalMatch(userID string) (bool, error) {
	usage, err := s.GetUserDailyUsage(userID)
	if err != nil {
		return false, err
	}

	return usage.PenpalMatches < DefaultUsageLimits.DailyMatches, nil
}

// UsePenpalMatch 记录用户使用了一次笔友匹配
func (s *UserUsageService) UsePenpalMatch(userID string) error {
	usage, err := s.GetUserDailyUsage(userID)
	if err != nil {
		return err
	}

	if usage.PenpalMatches >= DefaultUsageLimits.DailyMatches {
		return fmt.Errorf("daily penpal match limit exceeded")
	}

	usage.PenpalMatches++
	return s.db.Save(usage).Error
}

// CanUseCuration 检查用户是否可以使用策展功能
func (s *UserUsageService) CanUseCuration(userID string) (bool, error) {
	usage, err := s.GetUserDailyUsage(userID)
	if err != nil {
		return false, err
	}

	return usage.LettersCurated < DefaultUsageLimits.DailyCurations, nil
}

// UseCuration 记录用户使用了一次策展
func (s *UserUsageService) UseCuration(userID string) error {
	usage, err := s.GetUserDailyUsage(userID)
	if err != nil {
		return err
	}

	if usage.LettersCurated >= DefaultUsageLimits.DailyCurations {
		return fmt.Errorf("daily curation limit exceeded")
	}

	usage.LettersCurated++
	return s.db.Save(usage).Error
}

// GetUserUsageStats 获取用户使用统计
func (s *UserUsageService) GetUserUsageStats(userID string) (*models.AIUsageStats, error) {
	usage, err := s.GetUserDailyUsage(userID)
	if err != nil {
		return nil, err
	}

	return &models.AIUsageStats{
		UserID:        userID,
		RequestCount:  usage.AIRepliesGenerated + usage.InspirationsUsed + usage.PenpalMatches + usage.LettersCurated,
		LastRequestAt: usage.UpdatedAt,
		DailyLimit:    DefaultUsageLimits.DailyAIReplies,
		MonthlyLimit:  DefaultUsageLimits.DailyAIReplies * 30,
		CreatedAt:     usage.CreatedAt,
		UpdatedAt:     usage.UpdatedAt,
	}, nil
}

// GetUserUsageHistory 获取用户使用历史（最近7天）
func (s *UserUsageService) GetUserUsageHistory(userID string) ([]UserDailyUsage, error) {
	var history []UserDailyUsage

	sevenDaysAgo := time.Now().AddDate(0, 0, -7).Truncate(24 * time.Hour)

	err := s.db.Where("user_id = ? AND date >= ?", userID, sevenDaysAgo).
		Order("date DESC").
		Find(&history).Error

	return history, err
}

// CleanupOldRecords 清理旧记录（保留30天）
func (s *UserUsageService) CleanupOldRecords() error {
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30).Truncate(24 * time.Hour)

	return s.db.Where("date < ?", thirtyDaysAgo).Delete(&UserDailyUsage{}).Error
}

// max 辅助函数
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
