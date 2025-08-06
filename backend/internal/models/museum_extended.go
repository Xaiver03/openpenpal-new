package models

import (
	"time"
)

// MuseumTag 博物馆标签
type MuseumTag struct {
	ID         string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Name       string    `json:"name" gorm:"type:varchar(50);not null;uniqueIndex"`
	Category   string    `json:"category" gorm:"type:varchar(50);default:'general'"`
	UsageCount int       `json:"usage_count" gorm:"default:0"`
	CreatedAt  time.Time `json:"created_at"`
}

func (MuseumTag) TableName() string {
	return "museum_tags"
}

// MuseumInteraction 博物馆条目互动记录
type MuseumInteraction struct {
	ID        string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	EntryID   string    `json:"entry_id" gorm:"type:varchar(36);not null;index"`
	UserID    string    `json:"user_id" gorm:"type:varchar(36);not null;index"`
	Type      string    `json:"type" gorm:"type:varchar(20);not null"` // view, like, bookmark, share
	CreatedAt time.Time `json:"created_at"`
}

func (MuseumInteraction) TableName() string {
	return "museum_interactions"
}

// MuseumReaction 博物馆条目反应
type MuseumReaction struct {
	ID           string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	EntryID      string    `json:"entry_id" gorm:"type:varchar(36);not null;index"`
	UserID       string    `json:"user_id" gorm:"type:varchar(36);not null;index"`
	ReactionType string    `json:"reaction_type" gorm:"type:varchar(20);not null"` // like, love, inspiring, touching
	Comment      string    `json:"comment" gorm:"type:text"`
	CreatedAt    time.Time `json:"created_at"`
}

func (MuseumReaction) TableName() string {
	return "museum_reactions"
}

// MuseumExhibition and MuseumExhibitionEntry are already defined in museum.go

// MuseumSubmission 博物馆提交记录
type MuseumSubmission struct {
	ID                string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	LetterID          string    `json:"letter_id" gorm:"type:varchar(36);not null"`
	SubmittedBy       string    `json:"submitted_by" gorm:"type:varchar(36);not null"`
	DisplayPreference string    `json:"display_preference" gorm:"type:varchar(20);default:'anonymous'"` // anonymous, penName, realName
	PenName           *string   `json:"pen_name" gorm:"type:varchar(100)"`
	SubmissionReason  string    `json:"submission_reason" gorm:"type:text"`
	CuratorNotes      *string   `json:"curator_notes" gorm:"type:text"`
	Status            string    `json:"status" gorm:"type:varchar(20);default:'pending'"` // pending, approved, rejected, withdrawn
	SubmittedAt       time.Time `json:"submitted_at"`
	ReviewedAt        *time.Time `json:"reviewed_at"`
	ReviewedBy        *string   `json:"reviewed_by" gorm:"type:varchar(36)"`
	
	// Relations
	Letter *Letter `json:"letter,omitempty" gorm:"foreignKey:LetterID"`
}

func (MuseumSubmission) TableName() string {
	return "museum_submissions"
}

// MuseumAnalytics 博物馆分析数据
type MuseumAnalytics struct {
	TotalEntries      int64          `json:"total_entries"`
	TotalViews        int64          `json:"total_views"`
	TotalLikes        int64          `json:"total_likes"`
	TotalShares       int64          `json:"total_shares"`
	PopularCategories []CategoryStat `json:"popular_categories"`
	PopularTags       []TagStat      `json:"popular_tags"`
	DailyStats        []DailyStat    `json:"daily_stats"`
	GeneratedAt       time.Time      `json:"generated_at"`
}

// CategoryStat 分类统计
type CategoryStat struct {
	Category string `json:"category"`
	Count    int    `json:"count"`
}

// TagStat 标签统计
type TagStat struct {
	Tag   string `json:"tag"`
	Count int    `json:"count"`
}

// DailyStat 每日统计
type DailyStat struct {
	Date        string `json:"date"`
	Views       int    `json:"views"`
	Likes       int    `json:"likes"`
	Submissions int    `json:"submissions"`
}

// 更新MuseumItem模型以包含新字段
type MuseumItemExtended struct {
	MuseumItem
	Letter           *Letter           `json:"letter,omitempty" gorm:"foreignKey:SourceID"`
	Submission       *MuseumSubmission `json:"submission,omitempty" gorm:"foreignKey:SubmissionID"`
	SubmissionID     *string           `json:"submission_id" gorm:"type:varchar(36)"`
	DisplayTitle     string            `json:"display_title" gorm:"type:varchar(200)"`
	AuthorDisplayType string           `json:"author_display_type" gorm:"type:varchar(20);default:'anonymous'"`
	AuthorDisplayName *string          `json:"author_display_name" gorm:"type:varchar(100)"`
	CuratorType      string            `json:"curator_type" gorm:"type:varchar(20);default:'system'"` // system, user, admin
	CuratorID        string            `json:"curator_id" gorm:"type:varchar(36)"`
	Categories       []string          `json:"categories" gorm:"-"`
	Tags             []string          `json:"tags_array" gorm:"-"`
	ModerationStatus string            `json:"moderation_status" gorm:"type:varchar(20);default:'pending'"`
	BookmarkCount    int               `json:"bookmark_count" gorm:"default:0"`
	AIMetadata       map[string]interface{} `json:"ai_metadata" gorm:"-"`
	PublishedAt      *time.Time        `json:"published_at"`
	FeaturedAt       *time.Time        `json:"featured_at"`
	ModeratedBy      *string           `json:"moderated_by" gorm:"type:varchar(36)"`
	ModeratedAt      *time.Time        `json:"moderated_at"`
	ModerationNotes  *string           `json:"moderation_notes" gorm:"type:text"`
}