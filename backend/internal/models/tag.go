package models

import (
	"time"
)

// TagType 标签类型
type TagType string

const (
	TagTypeUser      TagType = "user"      // 用户创建的标签
	TagTypeSystem    TagType = "system"    // 系统预定义标签
	TagTypeAI        TagType = "ai"        // AI生成的标签
	TagTypeCategory  TagType = "category"  // 分类标签
	TagTypeTrending  TagType = "trending"  // 热门标签
)

// TagStatus 标签状态
type TagStatus string

const (
	TagStatusActive    TagStatus = "active"    // 活跃
	TagStatusInactive  TagStatus = "inactive"  // 非活跃
	TagStatusBanned    TagStatus = "banned"    // 被禁用
	TagStatusPending   TagStatus = "pending"   // 待审核
)

// 使用已在moderation.go中定义的ContentType常量
// ContentTypeLetter   = "letter"   // 信件
// ContentTypeMuseum   = "museum"   // 博物馆条目  
// ContentTypeComment  = "comment"  // 评论
// ContentTypeProfile  = "profile"  // 用户资料
// ContentTypeEnvelope = "envelope" // 信封设计

// Tag 标签模型
type Tag struct {
	ID          string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Name        string    `json:"name" gorm:"type:varchar(100);not null;index"`
	DisplayName string    `json:"display_name" gorm:"type:varchar(100)"`               // 显示名称（支持中文等）
	Description string    `json:"description" gorm:"type:text"`                       // 标签描述
	Type        TagType   `json:"type" gorm:"type:varchar(20);not null;default:'user'"` // 标签类型
	Status      TagStatus `json:"status" gorm:"type:varchar(20);not null;default:'active'"` // 标签状态
	Color       string    `json:"color" gorm:"type:varchar(7)"`                       // 标签颜色（hex）
	Icon        string    `json:"icon" gorm:"type:varchar(50)"`                       // 标签图标
	
	// 分类相关
	CategoryID   *string `json:"category_id" gorm:"type:varchar(36);index"`          // 父分类ID
	Category     *Tag    `json:"category,omitempty" gorm:"foreignKey:CategoryID"`    // 父分类
	SubTags      []Tag   `json:"sub_tags,omitempty" gorm:"foreignKey:CategoryID"`    // 子标签
	
	// 统计信息
	UsageCount   int64   `json:"usage_count" gorm:"default:0;index"`                 // 使用次数
	TrendingScore float64 `json:"trending_score" gorm:"default:0"`                   // 热度分数
	
	// 元数据
	CreatedBy   string    `json:"created_by" gorm:"type:varchar(36)"`                // 创建者ID
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	
	// 关联数据（不存储在数据库中）
	ContentCount int64 `json:"content_count" gorm:"-"` // 关联的内容数量
	IsFollowed   bool  `json:"is_followed" gorm:"-"`   // 当前用户是否关注该标签
}

// ContentTag 内容标签关联模型
type ContentTag struct {
	ID          string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	ContentType string    `json:"content_type" gorm:"type:varchar(20);not null;index"` // 内容类型
	ContentID   string    `json:"content_id" gorm:"type:varchar(36);not null;index"`   // 内容ID
	TagID       string    `json:"tag_id" gorm:"type:varchar(36);not null;index"`       // 标签ID
	
	// 关联信息
	Tag         Tag       `json:"tag" gorm:"foreignKey:TagID"`                         // 标签详情
	
	// 标记信息
	Source      string    `json:"source" gorm:"type:varchar(20);default:'user'"`      // 标记来源：user, ai, system
	Confidence  float64   `json:"confidence" gorm:"default:1.0"`                      // 置信度（AI标记时使用）
	CreatedBy   string    `json:"created_by" gorm:"type:varchar(36)"`                 // 标记者ID
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	
	// 复合索引
	// 在数据库中创建: content_type + content_id + tag_id 的唯一索引
}

// TagCategory 标签分类模型
type TagCategory struct {
	ID          string `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Name        string `json:"name" gorm:"type:varchar(100);not null;uniqueIndex"`
	DisplayName string `json:"display_name" gorm:"type:varchar(100)"`
	Description string `json:"description" gorm:"type:text"`
	Color       string `json:"color" gorm:"type:varchar(7)"`
	Icon        string `json:"icon" gorm:"type:varchar(50)"`
	SortOrder   int    `json:"sort_order" gorm:"default:0"`
	IsActive    bool   `json:"is_active" gorm:"default:true"`
	
	// 关联数据
	Tags        []Tag     `json:"tags,omitempty" gorm:"foreignKey:CategoryID"`
	TagCount    int64     `json:"tag_count" gorm:"-"`
	
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// UserTagFollow 用户标签关注模型
type UserTagFollow struct {
	ID        string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	UserID    string    `json:"user_id" gorm:"type:varchar(36);not null;index"`
	TagID     string    `json:"tag_id" gorm:"type:varchar(36);not null;index"`
	
	// 关联信息
	Tag       Tag       `json:"tag" gorm:"foreignKey:TagID"`
	
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	
	// 复合唯一索引: user_id + tag_id
}

// TagTrend 标签趋势模型
type TagTrend struct {
	ID           string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	TagID        string    `json:"tag_id" gorm:"type:varchar(36);not null;index"`
	Date         time.Time `json:"date" gorm:"type:date;not null;index"`
	UsageCount   int64     `json:"usage_count" gorm:"default:0"`
	TrendScore   float64   `json:"trend_score" gorm:"default:0"`
	Rank         int       `json:"rank" gorm:"default:0"`
	
	// 关联信息
	Tag          Tag       `json:"tag" gorm:"foreignKey:TagID"`
	
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	
	// 复合唯一索引: tag_id + date
}

// =============== Request/Response DTOs ===============

// TagRequest 标签创建/更新请求
type TagRequest struct {
	Name        string  `json:"name" binding:"required,min=1,max=100"`
	DisplayName string  `json:"display_name" binding:"max=100"`
	Description string  `json:"description" binding:"max=500"`
	Color       string  `json:"color" binding:"omitempty,len=7"`
	Icon        string  `json:"icon" binding:"max=50"`
	CategoryID  *string `json:"category_id"`
}

// TagSearchRequest 标签搜索请求
type TagSearchRequest struct {
	Query       string   `json:"query"`
	Type        TagType  `json:"type"`
	CategoryID  *string  `json:"category_id"`
	Status      TagStatus `json:"status"`
	Page        int      `json:"page" binding:"min=1"`
	Limit       int      `json:"limit" binding:"min=1,max=100"`
	SortBy      string   `json:"sort_by"` // name, usage_count, trending_score, created_at
	SortOrder   string   `json:"sort_order"` // asc, desc
}

// ContentTagRequest 内容标签关联请求
type ContentTagRequest struct {
	ContentType string   `json:"content_type" binding:"required"`
	ContentID   string   `json:"content_id" binding:"required"`
	TagIDs      []string `json:"tag_ids" binding:"required"`
}

// TagSuggestionRequest 标签建议请求
type TagSuggestionRequest struct {
	ContentType string `json:"content_type" binding:"required"`
	ContentID   string `json:"content_id"`
	Content     string `json:"content"`
	Limit       int    `json:"limit" binding:"min=1,max=20"`
}

// TagCategoryRequest 标签分类请求
type TagCategoryRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=100"`
	DisplayName string `json:"display_name" binding:"max=100"`
	Description string `json:"description" binding:"max=500"`
	Color       string `json:"color" binding:"omitempty,len=7"`
	Icon        string `json:"icon" binding:"max=50"`
	SortOrder   int    `json:"sort_order"`
}

// =============== Response DTOs ===============

// TagResponse 标签响应
type TagResponse struct {
	Tag
	ContentCount int64 `json:"content_count"`
	IsFollowed   bool  `json:"is_followed"`
}

// TagListResponse 标签列表响应
type TagListResponse struct {
	Tags       []TagResponse `json:"tags"`
	Total      int64         `json:"total"`
	Page       int           `json:"page"`
	Limit      int           `json:"limit"`
	TotalPages int           `json:"total_pages"`
}

// TagStatsResponse 标签统计响应
type TagStatsResponse struct {
	TotalTags       int64   `json:"total_tags"`
	UserTags        int64   `json:"user_tags"`
	SystemTags      int64   `json:"system_tags"`
	AITags          int64   `json:"ai_tags"`
	TotalUsage      int64   `json:"total_usage"`
	AvgUsagePerTag  float64 `json:"avg_usage_per_tag"`
	TrendingTags    []Tag   `json:"trending_tags"`
	PopularTags     []Tag   `json:"popular_tags"`
}

// TagSuggestionResponse 标签建议响应
type TagSuggestionResponse struct {
	SuggestedTags   []Tag   `json:"suggested_tags"`
	AIGeneratedTags []Tag   `json:"ai_generated_tags"`
	RelatedTags     []Tag   `json:"related_tags"`
	PopularTags     []Tag   `json:"popular_tags"`
}

// ContentTagsResponse 内容标签响应
type ContentTagsResponse struct {
	ContentType string `json:"content_type"`
	ContentID   string `json:"content_id"`
	Tags        []Tag  `json:"tags"`
	TagCount    int    `json:"tag_count"`
}

// TagTrendResponse 标签趋势响应
type TagTrendResponse struct {
	TagID       string              `json:"tag_id"`
	TagName     string              `json:"tag_name"`
	TrendData   []TagTrendDataPoint `json:"trend_data"`
	CurrentRank int                 `json:"current_rank"`
	TrendChange string              `json:"trend_change"` // up, down, stable
}

// TagTrendDataPoint 标签趋势数据点
type TagTrendDataPoint struct {
	Date       string  `json:"date"`
	Usage      int64   `json:"usage"`
	TrendScore float64 `json:"trend_score"`
	Rank       int     `json:"rank"`
}