package models

import (
	"time"
	"gorm.io/gorm"
)

// FollowStatus 关注状态枚举
type FollowStatus string

const (
	FollowStatusActive  FollowStatus = "active"  // 正常关注
	FollowStatusBlocked FollowStatus = "blocked" // 被屏蔽
	FollowStatusMuted   FollowStatus = "muted"   // 静音
)

// UserRelationship 用户关系模型
type UserRelationship struct {
	ID                  string       `json:"id" gorm:"primaryKey;type:varchar(36)"`
	FollowerID          string       `json:"follower_id" gorm:"type:varchar(36);index"`
	FollowingID         string       `json:"following_id" gorm:"type:varchar(36);index"`
	Status              FollowStatus `json:"status" gorm:"type:varchar(20);default:'active'"`
	NotificationEnabled bool         `json:"notification_enabled" gorm:"default:true"`
	CreatedAt           time.Time    `json:"created_at"`
	UpdatedAt           time.Time    `json:"updated_at"`

	// 关联
	Follower  User `json:"follower,omitempty" gorm:"foreignKey:FollowerID"`
	Following User `json:"following,omitempty" gorm:"foreignKey:FollowingID"`
}

// FollowUser 关注用户响应结构（匹配前端类型）
type FollowUser struct {
	ID                    string    `json:"id"`
	Username              string    `json:"username"`
	Nickname              string    `json:"nickname"`
	Avatar                string    `json:"avatar"`
	Bio                   string    `json:"bio"`
	School                string    `json:"school"`
	OPCode                string    `json:"op_code"`
	WritingLevel          int       `json:"writing_level"`
	CourierLevel          int       `json:"courier_level"`
	FollowersCount        int       `json:"followers_count"`
	FollowingCount        int       `json:"following_count"`
	LettersCount          int       `json:"letters_count"`
	IsFollowing           bool      `json:"is_following,omitempty"`
	IsFollower            bool      `json:"is_follower,omitempty"`
	FollowStatus          string    `json:"follow_status,omitempty"`
	FollowedAt            string    `json:"followed_at,omitempty"`
	MutualFollowersCount  int       `json:"mutual_followers_count,omitempty"`
}

// FollowStats 关注统计
type FollowStats struct {
	UserID              string `json:"user_id" gorm:"primaryKey;type:varchar(36)"`
	FollowersCount      int    `json:"followers_count" gorm:"default:0"`
	FollowingCount      int    `json:"following_count" gorm:"default:0"`
	MutualFollowsCount  int    `json:"mutual_follows_count" gorm:"default:0"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// FollowActivity 关注活动记录
type FollowActivity struct {
	ID        string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	ActorID   string    `json:"actor_id" gorm:"type:varchar(36);index"`
	TargetID  string    `json:"target_id" gorm:"type:varchar(36);index"`
	Type      string    `json:"type" gorm:"type:varchar(50)"`
	CreatedAt time.Time `json:"created_at"`

	// 关联
	Actor  User `json:"actor,omitempty" gorm:"foreignKey:ActorID"`
	Target User `json:"target,omitempty" gorm:"foreignKey:TargetID"`
}

// UserSuggestion 用户推荐
type UserSuggestion struct {
	ID              string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	UserID          string    `json:"user_id" gorm:"type:varchar(36);index"`
	SuggestedUserID string    `json:"suggested_user_id" gorm:"type:varchar(36);index"`
	Reason          string    `json:"reason" gorm:"type:varchar(100)"`
	Score           float64   `json:"score" gorm:"type:decimal(5,4)"`
	CreatedAt       time.Time `json:"created_at"`
	
	// 关联
	User          User `json:"user,omitempty" gorm:"foreignKey:UserID"`
	SuggestedUser User `json:"suggested_user,omitempty" gorm:"foreignKey:SuggestedUserID"`
}

// Request/Response 结构

// FollowActionRequest 关注操作请求
type FollowActionRequest struct {
	UserID              string `json:"user_id" binding:"required"`
	NotificationEnabled bool   `json:"notification_enabled"`
}

// FollowActionResponse 关注操作响应
type FollowActionResponse struct {
	Success        bool   `json:"success"`
	IsFollowing    bool   `json:"is_following"`
	FollowerCount  int    `json:"follower_count"`
	FollowingCount int    `json:"following_count"`
	FollowedAt     string `json:"followed_at,omitempty"`
	Message        string `json:"message,omitempty"`
}

// FollowListRequest 关注列表请求
type FollowListRequest struct {
	Page         int      `json:"page" form:"page"`
	Limit        int      `json:"limit" form:"limit"`
	SortBy       string   `json:"sort_by" form:"sort_by"`
	Order        string   `json:"order" form:"order"`
	Search       string   `json:"search" form:"search"`
	SchoolFilter string   `json:"school_filter" form:"school_filter"`
	StatusFilter []string `json:"status_filter" form:"status_filter"`
}

// FollowListResponse 关注列表响应
type FollowListResponse struct {
	Users []FollowUser `json:"users"`
	Pagination struct {
		Page  int `json:"page"`
		Limit int `json:"limit"`
		Total int `json:"total"`
		Pages int `json:"pages"`
	} `json:"pagination"`
}

// UserSearchRequest 用户搜索请求
type UserSearchRequest struct {
	Query        string `json:"query" form:"query"`
	SchoolCode   string `json:"school_code" form:"school_code"`
	Role         string `json:"role" form:"role"`
	MinFollowers int    `json:"min_followers" form:"min_followers"`
	MaxFollowers int    `json:"max_followers" form:"max_followers"`
	ActiveSince  string `json:"active_since" form:"active_since"`
	SortBy       string `json:"sort_by" form:"sort_by"`
	Order        string `json:"order" form:"order"`
	Limit        int    `json:"limit" form:"limit"`
	Offset       int    `json:"offset" form:"offset"`
}

// UserSearchResponse 用户搜索响应
type UserSearchResponse struct {
	Users           []FollowUser `json:"users"`
	Total           int          `json:"total"`
	Query           string       `json:"query"`
	Suggestions     []string     `json:"suggestions,omitempty"`
	FiltersApplied  interface{}  `json:"filters_applied"`
}

// UserSuggestionsRequest 用户推荐请求
type UserSuggestionsRequest struct {
	Limit             int     `json:"limit" form:"limit"`
	BasedOn           string  `json:"based_on" form:"based_on"`
	ExcludeFollowed   bool    `json:"exclude_followed" form:"exclude_followed"`
	MinActivityScore  float64 `json:"min_activity_score" form:"min_activity_score"`
}

// UserSuggestionsResponse 用户推荐响应
type UserSuggestionsResponse struct {
	Suggestions         []FollowSuggestionItem `json:"suggestions"`
	AlgorithmUsed       string                 `json:"algorithm_used"`
	RefreshAvailableAt  string                 `json:"refresh_available_at"`
}

// FollowSuggestionItem 推荐项
type FollowSuggestionItem struct {
	User             FollowUser   `json:"user"`
	Reason           string       `json:"reason"`
	ConfidenceScore  float64      `json:"confidence_score"`
	MutualFollowers  []FollowUser `json:"mutual_followers,omitempty"`
	CommonInterests  []string     `json:"common_interests,omitempty"`
}

// FollowStatsResponse 关注统计响应
type FollowStatsResponse struct {
	FollowersCount       int          `json:"followers_count"`
	FollowingCount       int          `json:"following_count"`
	MutualFollowersCount int          `json:"mutual_followers_count"`
	RecentFollowers      []FollowUser `json:"recent_followers"`
	PopularFollowing     []FollowUser `json:"popular_following"`
}

// TableName 指定表名
func (UserRelationship) TableName() string {
	return "user_relationships"
}

func (FollowStats) TableName() string {
	return "user_follow_stats"
}

func (FollowActivity) TableName() string {
	return "follow_activities"
}

func (UserSuggestion) TableName() string {
	return "user_suggestions"
}

// BeforeCreate GORM钩子 - 生成UUID
func (ur *UserRelationship) BeforeCreate(tx *gorm.DB) error {
	if ur.ID == "" {
		ur.ID = generateUUID()
	}
	return nil
}

func (fa *FollowActivity) BeforeCreate(tx *gorm.DB) error {
	if fa.ID == "" {
		fa.ID = generateUUID()
	}
	return nil
}

func (us *UserSuggestion) BeforeCreate(tx *gorm.DB) error {
	if us.ID == "" {
		us.ID = generateUUID()
	}
	return nil
}

// IsActive 检查关注关系是否为活跃状态
func (ur *UserRelationship) IsActive() bool {
	return ur.Status == FollowStatusActive
}

// CanReceiveNotifications 检查是否可以接收通知
func (ur *UserRelationship) CanReceiveNotifications() bool {
	return ur.IsActive() && ur.NotificationEnabled
}

// ToFollowUser 将User转换为FollowUser
func (u *User) ToFollowUser(stats *FollowStats, profile *UserProfileExtended) *FollowUser {
	followUser := &FollowUser{
		ID:           u.ID,
		Username:     u.Username,
		Nickname:     u.Nickname,
		Avatar:       u.Avatar,
		WritingLevel: 1, // Default
		CourierLevel: 0, // Default
	}

	// 添加统计信息
	if stats != nil {
		followUser.FollowersCount = stats.FollowersCount
		followUser.FollowingCount = stats.FollowingCount
	}

	// 添加档案信息
	if profile != nil {
		followUser.Bio = profile.Bio
		followUser.School = profile.School
		followUser.OPCode = profile.OPCode
		followUser.WritingLevel = profile.WritingLevel
		followUser.CourierLevel = profile.CourierLevel
	}

	return followUser
}

// Note: generateUUID and generateRandomString functions are already defined in other model files