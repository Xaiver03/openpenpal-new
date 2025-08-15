package models

import (
	"time"

	"github.com/google/uuid"
)

// PersonaRelationship 人物关系类型
type PersonaRelationship string

const (
	RelationshipDeceased     PersonaRelationship = "deceased"      // 已故亲人
	RelationshipDistantFriend PersonaRelationship = "distant_friend" // 疏远朋友
	RelationshipUnspokenLove PersonaRelationship = "unspoken_love"  // 暗恋对象
	RelationshipCustom       PersonaRelationship = "custom"        // 自定义关系
)

// CloudPersona 云中锦书人物模型
type CloudPersona struct {
	ID           string              `json:"id" gorm:"primaryKey;type:varchar(36)"`
	UserID       string              `json:"user_id" gorm:"type:varchar(36);not null;index"`
	Name         string              `json:"name" gorm:"type:varchar(100);not null"`
	Relationship PersonaRelationship `json:"relationship" gorm:"type:varchar(50);not null"`
	Description  string              `json:"description" gorm:"type:text"`
	BackgroundStory string           `json:"background_story" gorm:"type:text"` // 背景故事
	Personality  string              `json:"personality" gorm:"type:text"`      // 性格特征
	Memories     string              `json:"memories" gorm:"type:text"`         // 共同回忆
	LastInteraction time.Time        `json:"last_interaction"`                   // 最后交流时间
	IsActive     bool                `json:"is_active" gorm:"default:true"`
	CreatedAt    time.Time           `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time           `json:"updated_at" gorm:"autoUpdateTime"`
}

// CloudLetterStatus 云信件状态
type CloudLetterStatus string

const (
	CloudLetterStatusDraft         CloudLetterStatus = "draft"          // 草稿
	CloudLetterStatusAIEnhanced    CloudLetterStatus = "ai_enhanced"    // AI增强完成
	CloudLetterStatusUnderReview   CloudLetterStatus = "under_review"   // 审核中
	CloudLetterStatusRevisionNeeded CloudLetterStatus = "revision_needed" // 需要修改
	CloudLetterStatusApproved      CloudLetterStatus = "approved"       // 已批准
	CloudLetterStatusDelivered     CloudLetterStatus = "delivered"      // 已投递
	CloudLetterStatusReplied       CloudLetterStatus = "replied"        // 已回信
)

// CloudLetter 云信件模型
type CloudLetter struct {
	ID              string             `json:"id" gorm:"primaryKey;type:varchar(36)"`
	UserID          string             `json:"user_id" gorm:"type:varchar(36);not null;index"`
	PersonaID       string             `json:"persona_id" gorm:"type:varchar(36);not null;index"`
	OriginalContent string             `json:"original_content" gorm:"type:text;not null"` // 用户原始内容
	AIEnhancedDraft string             `json:"ai_enhanced_draft" gorm:"type:text"`         // AI增强草稿
	FinalContent    string             `json:"final_content" gorm:"type:text"`             // 最终内容
	AIReply         string             `json:"ai_reply" gorm:"type:text"`                  // AI回信
	Status          CloudLetterStatus  `json:"status" gorm:"type:varchar(20);default:'draft'"`
	ReviewerLevel   int                `json:"reviewer_level" gorm:"default:0"`         // 审核员等级
	ReviewerID      string             `json:"reviewer_id" gorm:"type:varchar(36)"`     // 审核员ID
	ReviewComments  string             `json:"review_comments" gorm:"type:text"`        // 审核意见
	DeliveryDate    *time.Time         `json:"delivery_date"`                          // 预定投递时间
	ActualDeliveryDate *time.Time      `json:"actual_delivery_date"`                   // 实际投递时间
	EmotionalTone   string             `json:"emotional_tone" gorm:"type:varchar(50)"` // 情感色调
	CreatedAt       time.Time          `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time          `json:"updated_at" gorm:"autoUpdateTime"`
}

// BeforeCreate GORM hook to generate UUID
func (c *CloudPersona) BeforeCreate() error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}

// BeforeCreate GORM hook to generate UUID
func (c *CloudLetter) BeforeCreate() error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}