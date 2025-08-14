package models

import (
	"time"
)

// PrivacyLevel 隐私级别
type PrivacyLevel string

const (
	PrivacyPublic  PrivacyLevel = "public"  // 公开
	PrivacySchool  PrivacyLevel = "school"  // 同校可见
	PrivacyFriends PrivacyLevel = "friends" // 好友可见
	PrivacyPrivate PrivacyLevel = "private" // 仅自己
)

// ProfileVisibility 个人资料可见性设置
type ProfileVisibility struct {
	Bio          PrivacyLevel `json:"bio" gorm:"type:varchar(20);default:'school'"`
	SchoolInfo   PrivacyLevel `json:"school_info" gorm:"type:varchar(20);default:'public'"`
	ContactInfo  PrivacyLevel `json:"contact_info" gorm:"type:varchar(20);default:'friends'"`
	ActivityFeed PrivacyLevel `json:"activity_feed" gorm:"type:varchar(20);default:'school'"`
	FollowLists  PrivacyLevel `json:"follow_lists" gorm:"type:varchar(20);default:'friends'"`
	Statistics   PrivacyLevel `json:"statistics" gorm:"type:varchar(20);default:'public'"`
	LastActive   PrivacyLevel `json:"last_active" gorm:"type:varchar(20);default:'school'"`
}

// SocialPrivacy 社交隐私设置
type SocialPrivacy struct {
	AllowFollowRequests   bool `json:"allow_follow_requests" gorm:"default:true"`
	AllowComments         bool `json:"allow_comments" gorm:"default:true"`
	AllowDirectMessages   bool `json:"allow_direct_messages" gorm:"default:true"`
	ShowInDiscovery       bool `json:"show_in_discovery" gorm:"default:true"`
	ShowInSuggestions     bool `json:"show_in_suggestions" gorm:"default:true"`
	AllowSchoolSearch     bool `json:"allow_school_search" gorm:"default:true"`
}

// NotificationPrivacy 通知隐私设置
type NotificationPrivacy struct {
	NewFollowers       bool `json:"new_followers" gorm:"default:true"`
	FollowRequests     bool `json:"follow_requests" gorm:"default:true"`
	Comments           bool `json:"comments" gorm:"default:true"`
	Mentions           bool `json:"mentions" gorm:"default:true"`
	DirectMessages     bool `json:"direct_messages" gorm:"default:true"`
	SystemUpdates      bool `json:"system_updates" gorm:"default:true"`
	EmailNotifications bool `json:"email_notifications" gorm:"default:false"`
}

// BlockingSettings 屏蔽设置
type BlockingSettings struct {
	BlockedUsers           []string `json:"blocked_users" gorm:"type:json"`
	MutedUsers            []string `json:"muted_users" gorm:"type:json"`
	BlockedKeywords       []string `json:"blocked_keywords" gorm:"type:json"`
	AutoBlockNewAccounts  bool     `json:"auto_block_new_accounts" gorm:"default:false"`
	BlockNonSchoolUsers   bool     `json:"block_non_school_users" gorm:"default:false"`
}

// PrivacySettings 用户隐私设置
type PrivacySettings struct {
	ID                   string               `json:"id" gorm:"primaryKey;type:varchar(36)"`
	UserID               string               `json:"user_id" gorm:"type:varchar(36);uniqueIndex;not null"`
	ProfileVisibility    ProfileVisibility    `json:"profile_visibility" gorm:"embedded;embeddedPrefix:profile_"`
	SocialPrivacy        SocialPrivacy        `json:"social_privacy" gorm:"embedded;embeddedPrefix:social_"`
	NotificationPrivacy  NotificationPrivacy  `json:"notification_privacy" gorm:"embedded;embeddedPrefix:notification_"`
	BlockingSettings     BlockingSettings     `json:"blocking_settings" gorm:"embedded;embeddedPrefix:blocking_"`
	CreatedAt            time.Time            `json:"created_at"`
	UpdatedAt            time.Time            `json:"updated_at"`
	
	// Associations
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// PrivacyCheckResult 隐私检查结果
type PrivacyCheckResult struct {
	CanViewProfile   bool   `json:"can_view_profile"`
	CanViewBio       bool   `json:"can_view_bio"`
	CanViewSchool    bool   `json:"can_view_school"`
	CanViewContact   bool   `json:"can_view_contact"`
	CanViewActivity  bool   `json:"can_view_activity"`
	CanViewFollowers bool   `json:"can_view_followers"`
	CanViewFollowing bool   `json:"can_view_following"`
	CanViewStats     bool   `json:"can_view_stats"`
	CanFollow        bool   `json:"can_follow"`
	CanComment       bool   `json:"can_comment"`
	CanMessage       bool   `json:"can_message"`
	Reason           string `json:"reason,omitempty"`
}

// Request/Response DTOs

// UpdatePrivacySettingsRequest 更新隐私设置请求
type UpdatePrivacySettingsRequest struct {
	ProfileVisibility   *ProfileVisibility   `json:"profile_visibility,omitempty"`
	SocialPrivacy       *SocialPrivacy       `json:"social_privacy,omitempty"`
	NotificationPrivacy *NotificationPrivacy `json:"notification_privacy,omitempty"`
	BlockingSettings    *BlockingSettings    `json:"blocking_settings,omitempty"`
}

// BlockUserRequest 屏蔽用户请求
type BlockUserRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

// MuteUserRequest 静音用户请求  
type MuteUserRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

// AddBlockedKeywordRequest 添加屏蔽关键词请求
type AddBlockedKeywordRequest struct {
	Keyword string `json:"keyword" binding:"required"`
}

// BatchPrivacyCheckRequest 批量隐私检查请求
type BatchPrivacyCheckRequest struct {
	Actions []string `json:"actions" binding:"required"`
}

// Helper methods for PrivacySettings

// TableName returns the table name for GORM
func (PrivacySettings) TableName() string {
	return "privacy_settings"
}

// IsBlocked 检查用户是否被屏蔽
func (ps *PrivacySettings) IsBlocked(userID string) bool {
	for _, blockedID := range ps.BlockingSettings.BlockedUsers {
		if blockedID == userID {
			return true
		}
	}
	return false
}

// IsMuted 检查用户是否被静音
func (ps *PrivacySettings) IsMuted(userID string) bool {
	for _, mutedID := range ps.BlockingSettings.MutedUsers {
		if mutedID == userID {
			return true
		}
	}
	return false
}

// HasBlockedKeyword 检查内容是否包含屏蔽关键词
func (ps *PrivacySettings) HasBlockedKeyword(content string) bool {
	for _, keyword := range ps.BlockingSettings.BlockedKeywords {
		if len(keyword) > 0 && contains(content, keyword) {
			return true
		}
	}
	return false
}

// AddBlockedUser 添加屏蔽用户
func (ps *PrivacySettings) AddBlockedUser(userID string) {
	// 避免重复添加
	if !ps.IsBlocked(userID) {
		ps.BlockingSettings.BlockedUsers = append(ps.BlockingSettings.BlockedUsers, userID)
	}
}

// RemoveBlockedUser 移除屏蔽用户
func (ps *PrivacySettings) RemoveBlockedUser(userID string) {
	filtered := make([]string, 0, len(ps.BlockingSettings.BlockedUsers))
	for _, blockedID := range ps.BlockingSettings.BlockedUsers {
		if blockedID != userID {
			filtered = append(filtered, blockedID)
		}
	}
	ps.BlockingSettings.BlockedUsers = filtered
}

// AddMutedUser 添加静音用户
func (ps *PrivacySettings) AddMutedUser(userID string) {
	// 避免重复添加
	if !ps.IsMuted(userID) {
		ps.BlockingSettings.MutedUsers = append(ps.BlockingSettings.MutedUsers, userID)
	}
}

// RemoveMutedUser 移除静音用户
func (ps *PrivacySettings) RemoveMutedUser(userID string) {
	filtered := make([]string, 0, len(ps.BlockingSettings.MutedUsers))
	for _, mutedID := range ps.BlockingSettings.MutedUsers {
		if mutedID != userID {
			filtered = append(filtered, mutedID)
		}
	}
	ps.BlockingSettings.MutedUsers = filtered
}

// AddBlockedKeyword 添加屏蔽关键词
func (ps *PrivacySettings) AddBlockedKeyword(keyword string) {
	// 避免重复添加
	for _, existingKeyword := range ps.BlockingSettings.BlockedKeywords {
		if existingKeyword == keyword {
			return
		}
	}
	ps.BlockingSettings.BlockedKeywords = append(ps.BlockingSettings.BlockedKeywords, keyword)
}

// RemoveBlockedKeyword 移除屏蔽关键词
func (ps *PrivacySettings) RemoveBlockedKeyword(keyword string) {
	filtered := make([]string, 0, len(ps.BlockingSettings.BlockedKeywords))
	for _, existingKeyword := range ps.BlockingSettings.BlockedKeywords {
		if existingKeyword != keyword {
			filtered = append(filtered, existingKeyword)
		}
	}
	ps.BlockingSettings.BlockedKeywords = filtered
}

// GetDefaultPrivacySettings 获取默认隐私设置
func GetDefaultPrivacySettings(userID string) *PrivacySettings {
	return &PrivacySettings{
		UserID: userID,
		ProfileVisibility: ProfileVisibility{
			Bio:          PrivacySchool,
			SchoolInfo:   PrivacyPublic,
			ContactInfo:  PrivacyFriends,
			ActivityFeed: PrivacySchool,
			FollowLists:  PrivacyFriends,
			Statistics:   PrivacyPublic,
			LastActive:   PrivacySchool,
		},
		SocialPrivacy: SocialPrivacy{
			AllowFollowRequests:   true,
			AllowComments:         true,
			AllowDirectMessages:   true,
			ShowInDiscovery:       true,
			ShowInSuggestions:     true,
			AllowSchoolSearch:     true,
		},
		NotificationPrivacy: NotificationPrivacy{
			NewFollowers:       true,
			FollowRequests:     true,
			Comments:           true,
			Mentions:           true,
			DirectMessages:     true,
			SystemUpdates:      true,
			EmailNotifications: false,
		},
		BlockingSettings: BlockingSettings{
			BlockedUsers:          make([]string, 0),
			MutedUsers:           make([]string, 0),
			BlockedKeywords:      make([]string, 0),
			AutoBlockNewAccounts: false,
			BlockNonSchoolUsers:  false,
		},
	}
}

// Helper function to check if string contains substring (case-insensitive)
func contains(s, substr string) bool {
	// Simple case-insensitive contains check
	// In production, you might want to use a more sophisticated text matching
	s = toLower(s)
	substr = toLower(substr)
	return len(s) >= len(substr) && indexOf(s, substr) >= 0
}

// Simple toLowerCase implementation
func toLower(s string) string {
	result := make([]rune, 0, len(s))
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			result = append(result, r+32)
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

// Simple indexOf implementation
func indexOf(s, substr string) int {
	if len(substr) == 0 {
		return 0
	}
	if len(substr) > len(s) {
		return -1
	}
	
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}