package services

import (
	"fmt"
	"openpenpal-backend/internal/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PrivacyService 隐私设置服务
type PrivacyService struct {
	db *gorm.DB
}

// NewPrivacyService 创建隐私服务实例
func NewPrivacyService(db *gorm.DB) *PrivacyService {
	return &PrivacyService{
		db: db,
	}
}

// GetPrivacySettings 获取用户隐私设置
func (s *PrivacyService) GetPrivacySettings(userID string) (*models.PrivacySettings, error) {
	var settings models.PrivacySettings

	err := s.db.Where("user_id = ?", userID).First(&settings).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 如果不存在，创建默认设置
			return s.CreateDefaultSettings(userID)
		}
		return nil, fmt.Errorf("failed to get privacy settings: %w", err)
	}

	return &settings, nil
}

// CreateDefaultSettings 创建默认隐私设置
func (s *PrivacyService) CreateDefaultSettings(userID string) (*models.PrivacySettings, error) {
	settings := models.GetDefaultPrivacySettings(userID)
	settings.ID = uuid.New().String()
	settings.CreatedAt = time.Now()
	settings.UpdatedAt = time.Now()

	err := s.db.Create(settings).Error
	if err != nil {
		return nil, fmt.Errorf("failed to create default privacy settings: %w", err)
	}

	return settings, nil
}

// UpdatePrivacySettings 更新隐私设置
func (s *PrivacyService) UpdatePrivacySettings(userID string, updates *models.UpdatePrivacySettingsRequest) (*models.PrivacySettings, error) {
	// 获取现有设置
	settings, err := s.GetPrivacySettings(userID)
	if err != nil {
		return nil, err
	}

	// 应用更新
	if updates.ProfileVisibility != nil {
		if updates.ProfileVisibility.Bio != "" {
			settings.ProfileVisibility.Bio = updates.ProfileVisibility.Bio
		}
		if updates.ProfileVisibility.SchoolInfo != "" {
			settings.ProfileVisibility.SchoolInfo = updates.ProfileVisibility.SchoolInfo
		}
		if updates.ProfileVisibility.ContactInfo != "" {
			settings.ProfileVisibility.ContactInfo = updates.ProfileVisibility.ContactInfo
		}
		if updates.ProfileVisibility.ActivityFeed != "" {
			settings.ProfileVisibility.ActivityFeed = updates.ProfileVisibility.ActivityFeed
		}
		if updates.ProfileVisibility.FollowLists != "" {
			settings.ProfileVisibility.FollowLists = updates.ProfileVisibility.FollowLists
		}
		if updates.ProfileVisibility.Statistics != "" {
			settings.ProfileVisibility.Statistics = updates.ProfileVisibility.Statistics
		}
		if updates.ProfileVisibility.LastActive != "" {
			settings.ProfileVisibility.LastActive = updates.ProfileVisibility.LastActive
		}
	}

	if updates.SocialPrivacy != nil {
		settings.SocialPrivacy = *updates.SocialPrivacy
	}

	if updates.NotificationPrivacy != nil {
		settings.NotificationPrivacy = *updates.NotificationPrivacy
	}

	if updates.BlockingSettings != nil {
		settings.BlockingSettings = *updates.BlockingSettings
	}

	settings.UpdatedAt = time.Now()

	// 保存更新
	err = s.db.Save(settings).Error
	if err != nil {
		return nil, fmt.Errorf("failed to update privacy settings: %w", err)
	}

	return settings, nil
}

// ResetPrivacySettings 重置隐私设置为默认值
func (s *PrivacyService) ResetPrivacySettings(userID string) (*models.PrivacySettings, error) {
	// 删除现有设置
	err := s.db.Where("user_id = ?", userID).Delete(&models.PrivacySettings{}).Error
	if err != nil {
		return nil, fmt.Errorf("failed to delete existing privacy settings: %w", err)
	}

	// 创建新的默认设置
	return s.CreateDefaultSettings(userID)
}

// CheckPrivacy 检查用户是否可以访问目标用户的特定内容/行为
func (s *PrivacyService) CheckPrivacy(viewerID, targetUserID, action string) (*models.PrivacyCheckResult, error) {
	// 如果是同一个用户，允许所有访问
	if viewerID == targetUserID {
		return &models.PrivacyCheckResult{
			CanViewProfile:   true,
			CanViewBio:       true,
			CanViewSchool:    true,
			CanViewContact:   true,
			CanViewActivity:  true,
			CanViewFollowers: true,
			CanViewFollowing: true,
			CanViewStats:     true,
			CanFollow:        false, // 不能关注自己
			CanComment:       true,
			CanMessage:       true,
		}, nil
	}

	// 获取目标用户的隐私设置
	targetSettings, err := s.GetPrivacySettings(targetUserID)
	if err != nil {
		return nil, err
	}

	// 检查是否被屏蔽
	if targetSettings.IsBlocked(viewerID) {
		return &models.PrivacyCheckResult{
			CanViewProfile:   false,
			CanViewBio:       false,
			CanViewSchool:    false,
			CanViewContact:   false,
			CanViewActivity:  false,
			CanViewFollowers: false,
			CanViewFollowing: false,
			CanViewStats:     false,
			CanFollow:        false,
			CanComment:       false,
			CanMessage:       false,
			Reason:           "You are blocked by this user",
		}, nil
	}

	// 获取用户关系信息
	relationship, err := s.getUserRelationship(viewerID, targetUserID)
	if err != nil {
		return nil, err
	}

	// 检查各种权限
	result := &models.PrivacyCheckResult{}

	result.CanViewProfile = true // 基本资料总是可见
	result.CanViewBio = s.canAccessByPrivacyLevel(targetSettings.ProfileVisibility.Bio, viewerID, targetUserID, relationship)
	result.CanViewSchool = s.canAccessByPrivacyLevel(targetSettings.ProfileVisibility.SchoolInfo, viewerID, targetUserID, relationship)
	result.CanViewContact = s.canAccessByPrivacyLevel(targetSettings.ProfileVisibility.ContactInfo, viewerID, targetUserID, relationship)
	result.CanViewActivity = s.canAccessByPrivacyLevel(targetSettings.ProfileVisibility.ActivityFeed, viewerID, targetUserID, relationship)
	result.CanViewFollowers = s.canAccessByPrivacyLevel(targetSettings.ProfileVisibility.FollowLists, viewerID, targetUserID, relationship)
	result.CanViewFollowing = s.canAccessByPrivacyLevel(targetSettings.ProfileVisibility.FollowLists, viewerID, targetUserID, relationship)
	result.CanViewStats = s.canAccessByPrivacyLevel(targetSettings.ProfileVisibility.Statistics, viewerID, targetUserID, relationship)

	// 社交权限检查
	result.CanFollow = targetSettings.SocialPrivacy.AllowFollowRequests
	result.CanComment = targetSettings.SocialPrivacy.AllowComments
	result.CanMessage = targetSettings.SocialPrivacy.AllowDirectMessages

	return result, nil
}

// BatchCheckPrivacy 批量检查隐私权限
func (s *PrivacyService) BatchCheckPrivacy(viewerID, targetUserID string, actions []string) (map[string]*models.PrivacyCheckResult, error) {
	results := make(map[string]*models.PrivacyCheckResult)

	for _, action := range actions {
		result, err := s.CheckPrivacy(viewerID, targetUserID, action)
		if err != nil {
			return nil, err
		}
		results[action] = result
	}

	return results, nil
}

// BlockUser 屏蔽用户
func (s *PrivacyService) BlockUser(userID, targetUserID string) error {
	settings, err := s.GetPrivacySettings(userID)
	if err != nil {
		return err
	}

	settings.AddBlockedUser(targetUserID)
	settings.UpdatedAt = time.Now()

	return s.db.Save(settings).Error
}

// UnblockUser 取消屏蔽用户
func (s *PrivacyService) UnblockUser(userID, targetUserID string) error {
	settings, err := s.GetPrivacySettings(userID)
	if err != nil {
		return err
	}

	settings.RemoveBlockedUser(targetUserID)
	settings.UpdatedAt = time.Now()

	return s.db.Save(settings).Error
}

// MuteUser 静音用户
func (s *PrivacyService) MuteUser(userID, targetUserID string) error {
	settings, err := s.GetPrivacySettings(userID)
	if err != nil {
		return err
	}

	settings.AddMutedUser(targetUserID)
	settings.UpdatedAt = time.Now()

	return s.db.Save(settings).Error
}

// UnmuteUser 取消静音用户
func (s *PrivacyService) UnmuteUser(userID, targetUserID string) error {
	settings, err := s.GetPrivacySettings(userID)
	if err != nil {
		return err
	}

	settings.RemoveMutedUser(targetUserID)
	settings.UpdatedAt = time.Now()

	return s.db.Save(settings).Error
}

// AddBlockedKeyword 添加屏蔽关键词
func (s *PrivacyService) AddBlockedKeyword(userID, keyword string) error {
	settings, err := s.GetPrivacySettings(userID)
	if err != nil {
		return err
	}

	settings.AddBlockedKeyword(keyword)
	settings.UpdatedAt = time.Now()

	return s.db.Save(settings).Error
}

// RemoveBlockedKeyword 移除屏蔽关键词
func (s *PrivacyService) RemoveBlockedKeyword(userID, keyword string) error {
	settings, err := s.GetPrivacySettings(userID)
	if err != nil {
		return err
	}

	settings.RemoveBlockedKeyword(keyword)
	settings.UpdatedAt = time.Now()

	return s.db.Save(settings).Error
}

// GetBlockedUsers 获取屏蔽用户列表
func (s *PrivacyService) GetBlockedUsers(userID string) ([]string, error) {
	settings, err := s.GetPrivacySettings(userID)
	if err != nil {
		return nil, err
	}

	return settings.BlockingSettings.BlockedUsers, nil
}

// GetMutedUsers 获取静音用户列表
func (s *PrivacyService) GetMutedUsers(userID string) ([]string, error) {
	settings, err := s.GetPrivacySettings(userID)
	if err != nil {
		return nil, err
	}

	return settings.BlockingSettings.MutedUsers, nil
}

// GetBlockedKeywords 获取屏蔽关键词列表
func (s *PrivacyService) GetBlockedKeywords(userID string) ([]string, error) {
	settings, err := s.GetPrivacySettings(userID)
	if err != nil {
		return nil, err
	}

	return settings.BlockingSettings.BlockedKeywords, nil
}

// Helper methods

// getUserRelationship 获取用户关系信息
func (s *PrivacyService) getUserRelationship(viewerID, targetUserID string) (*userRelationship, error) {
	relationship := &userRelationship{
		IsSameSchool: false,
		IsFollowing:  false,
		IsFollower:   false,
		IsMutual:     false,
	}

	// 检查是否同校
	var viewerUser, targetUser models.User
	err := s.db.Select("school_code").Where("id = ?", viewerID).First(&viewerUser).Error
	if err != nil {
		return relationship, err
	}

	err = s.db.Select("school_code").Where("id = ?", targetUserID).First(&targetUser).Error
	if err != nil {
		return relationship, err
	}

	relationship.IsSameSchool = viewerUser.SchoolCode == targetUser.SchoolCode

	// 检查关注关系（如果有关注系统的话）
	// TODO: 与关注系统集成

	return relationship, nil
}

// canAccessByPrivacyLevel 根据隐私级别检查是否可以访问
func (s *PrivacyService) canAccessByPrivacyLevel(level models.PrivacyLevel, viewerID, targetUserID string, relationship *userRelationship) bool {
	switch level {
	case models.PrivacyPublic:
		return true
	case models.PrivacySchool:
		return relationship.IsSameSchool
	case models.PrivacyFriends:
		return relationship.IsFollowing && relationship.IsFollower // 互关
	case models.PrivacyPrivate:
		return viewerID == targetUserID
	default:
		return false
	}
}

// userRelationship 用户关系信息
type userRelationship struct {
	IsSameSchool bool
	IsFollowing  bool
	IsFollower   bool
	IsMutual     bool
}
