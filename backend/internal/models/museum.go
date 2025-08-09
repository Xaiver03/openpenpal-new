package models

import (
	"strings"
	"time"
)

// MuseumItemStatus 博物馆物品状态
type MuseumItemStatus string

const (
	MuseumItemPending  MuseumItemStatus = "pending"
	MuseumItemApproved MuseumItemStatus = "approved"
	MuseumItemRejected MuseumItemStatus = "rejected"
)

// MuseumSourceType 博物馆物品来源类型
type MuseumSourceType string

const (
	SourceTypeLetter MuseumSourceType = "letter"
	SourceTypePhoto  MuseumSourceType = "photo"
	SourceTypeAudio  MuseumSourceType = "audio"
)

// MuseumItem 博物馆物品模型 (对应Prisma的museum_items表)
type MuseumItem struct {
	ID          string           `json:"id" gorm:"primaryKey;type:varchar(36)"`
	SourceType  MuseumSourceType `json:"sourceType" gorm:"column:source_type;type:varchar(20);not null"`
	SourceID    string           `json:"sourceId" gorm:"column:source_id;type:varchar(36);not null"`
	Title       string           `json:"title" gorm:"type:varchar(200)"`
	Description string           `json:"description" gorm:"type:text"`
	Tags        string           `json:"tags" gorm:"type:text"`
	Status      MuseumItemStatus `json:"status" gorm:"type:varchar(20);default:'pending'"`
	SubmittedBy string           `json:"submittedBy" gorm:"column:submitted_by;type:varchar(36)"`
	ApprovedBy  *string          `json:"approvedBy" gorm:"column:approved_by;type:varchar(36)"`
	ApprovedAt  *time.Time       `json:"approvedAt" gorm:"column:approved_at"`
	ViewCount   int              `json:"viewCount" gorm:"column:view_count;default:0"`
	LikeCount   int              `json:"likeCount" gorm:"column:like_count;default:0"`
	ShareCount  int              `json:"shareCount" gorm:"column:share_count;default:0"`
	CommentCount int             `json:"commentCount" gorm:"column:comment_count;default:0"`
	FeaturedAt   *time.Time      `json:"featuredAt" gorm:"column:featured_at"`
	
	// OP Code System Integration - 地理位置标记
	OriginOPCode string          `json:"origin_op_code,omitempty" gorm:"type:varchar(6);index"` // 来源地理位置OP Code
	CreatedAt   time.Time        `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt   time.Time        `json:"updatedAt" gorm:"column:updated_at"`
	
	// GORM Relations - Fixed database relation mappings
	Letter      *Letter          `json:"letter,omitempty" gorm:"foreignKey:SourceID;references:ID"`
	SubmittedByUser *User        `json:"submitted_by_user,omitempty" gorm:"foreignKey:SubmittedBy;references:ID"`
	ApprovedByUser  *User        `json:"approved_by_user,omitempty" gorm:"foreignKey:ApprovedBy;references:ID"`
}

func (MuseumItem) TableName() string {
	return "museum_items"
}

// MuseumCollection 博物馆收藏模型 (对应Prisma的museum_collections表)
type MuseumCollection struct {
	ID          string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Name        string    `json:"name" gorm:"type:varchar(200);not null"`
	Description string    `json:"description" gorm:"type:text"`
	CreatedBy   string    `json:"createdBy" gorm:"column:created_by;type:varchar(36);not null"`
	IsPublic    bool      `json:"isPublic" gorm:"column:is_public;default:true"`
	ItemCount   int       `json:"itemCount" gorm:"column:item_count;default:0"`
	CreatedAt   time.Time `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt   time.Time `json:"updatedAt" gorm:"column:updated_at"`
}

func (MuseumCollection) TableName() string {
	return "museum_collections"
}

// MuseumExhibitionEntry 博物馆展览条目关联表
type MuseumExhibitionEntry struct {
	ID           string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	CollectionID string    `json:"collectionId" gorm:"column:collection_id;type:varchar(36);not null"`
	ItemID       string    `json:"itemId" gorm:"column:item_id;type:varchar(36);not null"`
	DisplayOrder int       `json:"displayOrder" gorm:"column:display_order;default:0"`
	CreatedAt    time.Time `json:"createdAt" gorm:"column:created_at"`
}

func (MuseumExhibitionEntry) TableName() string {
	return "museum_exhibition_entries"
}

// GORM兼容性视图模型 (匹配预期的GORM表名)
type MuseumEntry struct {
	ID                string                 `json:"id" gorm:"primaryKey;type:varchar(36)"`
	LetterID          string                 `json:"letterId" gorm:"column:letter_id;type:varchar(36)"`
	SubmissionID      *string                `json:"submissionId" gorm:"column:submission_id;type:varchar(36)"`
	DisplayTitle      string                 `json:"displayTitle" gorm:"column:display_title;type:varchar(200)"`
	AuthorDisplayType string                 `json:"authorDisplayType" gorm:"column:author_display_type;type:varchar(20)"`
	AuthorDisplayName *string                `json:"authorDisplayName" gorm:"column:author_display_name;type:varchar(100)"`
	CuratorType       string                 `json:"curatorType" gorm:"column:curator_type;type:varchar(20)"`
	CuratorID         string                 `json:"curatorId" gorm:"column:curator_id;type:varchar(36)"`
	Categories        []string               `json:"categories" gorm:"type:text[]"`
	Tags              []string               `json:"tags" gorm:"type:text[]"`
	Status            MuseumItemStatus       `json:"status" gorm:"type:varchar(20)"`
	ModerationStatus  MuseumItemStatus       `json:"moderationStatus" gorm:"column:moderation_status;type:varchar(20)"`
	ViewCount         int                    `json:"viewCount" gorm:"column:view_count;default:0"`
	LikeCount         int                    `json:"likeCount" gorm:"column:like_count;default:0"`
	BookmarkCount     int                    `json:"bookmarkCount" gorm:"column:bookmark_count;default:0"`
	ShareCount        int                    `json:"shareCount" gorm:"column:share_count;default:0"`
	AIMetadata        string                 `json:"aiMetadata" gorm:"column:ai_metadata;type:text"` // Store as JSON string for SQLite compatibility
	SubmittedAt       time.Time              `json:"submittedAt" gorm:"column:submitted_at"`
	ApprovedAt        *time.Time             `json:"approvedAt" gorm:"column:approved_at"`
	FeaturedAt        *time.Time             `json:"featuredAt" gorm:"column:featured_at"`
	WithdrawnAt       *time.Time             `json:"withdrawnAt" gorm:"column:withdrawn_at"`
	CreatedAt         time.Time              `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt         time.Time              `json:"updatedAt" gorm:"column:updated_at"`
	DeletedAt         *time.Time             `json:"deletedAt" gorm:"column:deleted_at"`
}

func (MuseumEntry) TableName() string {
	return "museum_entries"
}

type MuseumExhibition struct {
	ID             string     `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Title          string     `json:"title" gorm:"type:varchar(200);not null"`
	Description    string     `json:"description" gorm:"type:text"`
	ThemeKeywords  string     `json:"themeKeywords" gorm:"column:theme_keywords;type:text"` // Store as comma-separated string for PostgreSQL compatibility
	Status         string     `json:"status" gorm:"type:varchar(20);default:'draft'"`
	CreatorID      string     `json:"creatorId" gorm:"column:creator_id;type:varchar(36)"`
	StartDate      time.Time  `json:"startDate" gorm:"column:start_date"`
	EndDate        *time.Time `json:"endDate" gorm:"column:end_date"`
	MaxEntries     int        `json:"maxEntries" gorm:"column:max_entries;default:50"`
	CurrentEntries int        `json:"currentEntries" gorm:"column:current_entries;default:0"`
	ViewCount      int        `json:"viewCount" gorm:"column:view_count;default:0"`
	IsPublic       bool       `json:"isPublic" gorm:"column:is_public;default:true"`
	IsFeatured     bool       `json:"isFeatured" gorm:"column:is_featured;default:false"`
	DisplayOrder   int        `json:"displayOrder" gorm:"column:display_order;default:0"`
	CreatedAt      time.Time  `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt      time.Time  `json:"updatedAt" gorm:"column:updated_at"`
	DeletedAt      *time.Time `json:"deletedAt" gorm:"column:deleted_at"`
}

func (MuseumExhibition) TableName() string {
	return "museum_exhibitions"
}

// GetThemeKeywordsSlice returns theme keywords as a slice
func (me *MuseumExhibition) GetThemeKeywordsSlice() []string {
	if me.ThemeKeywords == "" {
		return []string{}
	}
	keywords := strings.Split(me.ThemeKeywords, ",")
	result := make([]string, 0, len(keywords))
	for _, keyword := range keywords {
		if trimmed := strings.TrimSpace(keyword); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// SetThemeKeywordsSlice sets theme keywords from a slice
func (me *MuseumExhibition) SetThemeKeywordsSlice(keywords []string) {
	if len(keywords) == 0 {
		me.ThemeKeywords = ""
		return
	}
	me.ThemeKeywords = strings.Join(keywords, ",")
}
