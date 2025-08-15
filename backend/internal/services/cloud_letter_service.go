package services

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"openpenpal-backend/internal/config"
	"openpenpal-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CloudLetterService 云中锦书服务 - 支持向真实世界角色写信
type CloudLetterService struct {
	db              *gorm.DB
	config          *config.Config
	letterSvc       *LetterService
	aiSvc           *AIService
	courierSvc      *CourierService
	notificationSvc *NotificationService
}

// Type aliases for convenience
type PersonaRelationship = models.PersonaRelationship
type CloudPersona = models.CloudPersona
type CloudLetter = models.CloudLetter
type CloudLetterStatus = models.CloudLetterStatus

// Re-export constants for convenience
const (
	RelationshipDeceased      = models.RelationshipDeceased
	RelationshipDistantFriend = models.RelationshipDistantFriend
	RelationshipUnspokenLove  = models.RelationshipUnspokenLove
	RelationshipCustom        = models.RelationshipCustom

	CloudLetterStatusDraft          = models.CloudLetterStatusDraft
	CloudLetterStatusAIEnhanced     = models.CloudLetterStatusAIEnhanced
	CloudLetterStatusUnderReview    = models.CloudLetterStatusUnderReview
	CloudLetterStatusRevisionNeeded = models.CloudLetterStatusRevisionNeeded
	CloudLetterStatusApproved       = models.CloudLetterStatusApproved
	CloudLetterStatusDelivered      = models.CloudLetterStatusDelivered
	CloudLetterStatusReplied        = models.CloudLetterStatusReplied
)

// PersonaCreateRequest 创建人物请求
type PersonaCreateRequest struct {
	Name            string              `json:"name" binding:"required,min=1,max=100"`
	Relationship    PersonaRelationship `json:"relationship" binding:"required,oneof=deceased distant_friend unspoken_love custom"`
	Description     string              `json:"description" binding:"max=500"`
	BackgroundStory string              `json:"background_story" binding:"max=1000"`
	Personality     string              `json:"personality" binding:"max=500"`
	Memories        string              `json:"memories" binding:"max=1000"`
}

// CloudLetterCreateRequest 创建云信件请求
type CloudLetterCreateRequest struct {
	PersonaID     string    `json:"persona_id" binding:"required,uuid"`
	Content       string    `json:"content" binding:"required,min=10,max=5000"`
	DeliveryDate  *time.Time `json:"delivery_date,omitempty"`
	EmotionalTone string    `json:"emotional_tone,omitempty"`
}

// NewCloudLetterService 创建云中锦书服务
func NewCloudLetterService(db *gorm.DB, config *config.Config) *CloudLetterService {
	return &CloudLetterService{
		db:     db,
		config: config,
	}
}

// SetLetterService 设置信件服务
func (s *CloudLetterService) SetLetterService(letterSvc *LetterService) {
	s.letterSvc = letterSvc
}

// SetAIService 设置AI服务
func (s *CloudLetterService) SetAIService(aiSvc *AIService) {
	s.aiSvc = aiSvc
}

// SetCourierService 设置信使服务
func (s *CloudLetterService) SetCourierService(courierSvc *CourierService) {
	s.courierSvc = courierSvc
}

// SetNotificationService 设置通知服务
func (s *CloudLetterService) SetNotificationService(notificationSvc *NotificationService) {
	s.notificationSvc = notificationSvc
}

// CreatePersona 创建自定义人物角色
func (s *CloudLetterService) CreatePersona(ctx context.Context, userID string, req *PersonaCreateRequest) (*CloudPersona, error) {
	log.Printf("🎭 [CloudLetter] Creating persona: %s (type: %s)", req.Name, req.Relationship)

	persona := &CloudPersona{
		ID:              uuid.New().String(),
		UserID:          userID,
		Name:            req.Name,
		Relationship:    req.Relationship,
		Description:     req.Description,
		BackgroundStory: req.BackgroundStory,
		Personality:     req.Personality,
		Memories:        req.Memories,
		IsActive:        true,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := s.db.Create(persona).Error; err != nil {
		return nil, fmt.Errorf("failed to create persona: %w", err)
	}

	log.Printf("✅ [CloudLetter] Successfully created persona: %s (ID: %s)", persona.Name, persona.ID)
	return persona, nil
}

// GetUserPersonas 获取用户的所有人物角色
func (s *CloudLetterService) GetUserPersonas(ctx context.Context, userID string) ([]*CloudPersona, error) {
	var personas []*CloudPersona
	
	err := s.db.Where("user_id = ? AND is_active = ?", userID, true).
		Order("created_at DESC").
		Find(&personas).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get personas: %w", err)
	}

	return personas, nil
}

// CreateCloudLetter 创建云信件
func (s *CloudLetterService) CreateCloudLetter(ctx context.Context, userID string, req *CloudLetterCreateRequest) (*CloudLetter, error) {
	log.Printf("✉️ [CloudLetter] Creating cloud letter for persona: %s", req.PersonaID)

	// 验证persona是否存在且属于该用户
	var persona CloudPersona
	if err := s.db.Where("id = ? AND user_id = ? AND is_active = ?", req.PersonaID, userID, true).First(&persona).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("persona not found or not accessible")
		}
		return nil, fmt.Errorf("failed to validate persona: %w", err)
	}

	// 创建云信件
	cloudLetter := &CloudLetter{
		ID:              uuid.New().String(),
		UserID:          userID,
		PersonaID:       req.PersonaID,
		OriginalContent: req.Content,
		Status:          CloudLetterStatusDraft,
		DeliveryDate:    req.DeliveryDate,
		EmotionalTone:   req.EmotionalTone,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := s.db.Create(cloudLetter).Error; err != nil {
		return nil, fmt.Errorf("failed to create cloud letter: %w", err)
	}

	// 启动AI增强流程
	go s.enhanceLetterWithAI(context.Background(), cloudLetter, &persona)

	log.Printf("✅ [CloudLetter] Successfully created cloud letter: %s", cloudLetter.ID)
	return cloudLetter, nil
}

// enhanceLetterWithAI 使用AI增强信件内容
func (s *CloudLetterService) enhanceLetterWithAI(ctx context.Context, letter *CloudLetter, persona *CloudPersona) {
	log.Printf("🤖 [CloudLetter] Starting AI enhancement for letter: %s", letter.ID)

	if s.aiSvc == nil {
		log.Printf("⚠️ [CloudLetter] AI service not available, skipping enhancement")
		return
	}

	// 构建增强提示词
	enhancePrompt := s.buildEnhancementPrompt(letter, persona)

	// 调用AI服务增强内容
	if aiResponse, err := s.callAIForEnhancement(ctx, enhancePrompt); err == nil {
		// 更新增强后的内容
		s.db.Model(letter).Updates(map[string]interface{}{
			"ai_enhanced_draft": aiResponse,
			"status":           CloudLetterStatusAIEnhanced,
			"updated_at":       time.Now(),
		})

		log.Printf("✅ [CloudLetter] AI enhancement completed for letter: %s", letter.ID)

		// 自动提交审核（如果配置允许）
		if s.shouldAutoSubmitForReview(persona.Relationship) {
			go s.submitForReview(context.Background(), letter.ID)
		}
	} else {
		log.Printf("❌ [CloudLetter] AI enhancement failed: %v", err)
	}
}

// buildEnhancementPrompt 构建AI增强提示词
func (s *CloudLetterService) buildEnhancementPrompt(letter *CloudLetter, persona *CloudPersona) string {
	var prompt strings.Builder

	prompt.WriteString("请作为专业的情感文字编辑，帮助用户改善这封写给特殊人物的信件。\n\n")
	
	// 人物背景信息
	prompt.WriteString(fmt.Sprintf("收信人信息：\n"))
	prompt.WriteString(fmt.Sprintf("- 姓名：%s\n", persona.Name))
	prompt.WriteString(fmt.Sprintf("- 关系：%s\n", s.getRelationshipDescription(persona.Relationship)))
	
	if persona.Description != "" {
		prompt.WriteString(fmt.Sprintf("- 描述：%s\n", persona.Description))
	}
	
	if persona.Personality != "" {
		prompt.WriteString(fmt.Sprintf("- 性格：%s\n", persona.Personality))
	}
	
	if persona.Memories != "" {
		prompt.WriteString(fmt.Sprintf("- 共同回忆：%s\n", persona.Memories))
	}

	// 原始信件内容
	prompt.WriteString(fmt.Sprintf("\n原始信件内容：\n%s\n\n", letter.OriginalContent))

	// 增强指导
	prompt.WriteString("请根据以下要求改善信件：\n")
	prompt.WriteString("1. 保持原作者的真实情感和意图\n")
	prompt.WriteString("2. 根据收信人关系调整语调和措辞\n")
	prompt.WriteString("3. 增加情感深度和文字美感\n")
	prompt.WriteString("4. 确保内容适合这种特殊关系\n")
	
	// 根据关系类型添加特定指导
	switch persona.Relationship {
	case RelationshipDeceased:
		prompt.WriteString("5. 体现对已故亲人的思念和爱意\n")
		prompt.WriteString("6. 表达感恩和美好回忆\n")
	case RelationshipDistantFriend:
		prompt.WriteString("5. 表达久别重逢的喜悦和友谊的珍贵\n")
		prompt.WriteString("6. 回忆共同的美好时光\n")
	case RelationshipUnspokenLove:
		prompt.WriteString("5. 保持含蓄而深情的表达\n")
		prompt.WriteString("6. 避免过于直接的表白，保持美感\n")
	}

	prompt.WriteString("\n请返回改善后的信件内容，要求文字优美、情感真挚、符合中文表达习惯。")

	return prompt.String()
}

// getRelationshipDescription 获取关系描述
func (s *CloudLetterService) getRelationshipDescription(relationship PersonaRelationship) string {
	descriptions := map[PersonaRelationship]string{
		RelationshipDeceased:      "已故亲人/朋友",
		RelationshipDistantFriend: "多年未见的朋友",
		RelationshipUnspokenLove:  "未曾表白的心上人",
		RelationshipCustom:        "自定义关系",
	}
	
	if desc, exists := descriptions[relationship]; exists {
		return desc
	}
	return string(relationship)
}

// callAIForEnhancement 调用AI服务进行内容增强
func (s *CloudLetterService) callAIForEnhancement(ctx context.Context, prompt string) (string, error) {
	// 这里应该调用AI服务的文本增强功能
	// 暂时返回模拟增强内容
	return s.generateMockEnhancement(prompt), nil
}

// generateMockEnhancement 生成模拟增强内容（临时实现）
func (s *CloudLetterService) generateMockEnhancement(originalContent string) string {
	return fmt.Sprintf("【AI增强版本】\n\n%s\n\n（已通过AI优化语言表达和情感深度）", originalContent)
}

// shouldAutoSubmitForReview 判断是否应该自动提交审核
func (s *CloudLetterService) shouldAutoSubmitForReview(relationship PersonaRelationship) bool {
	// 敏感关系需要审核
	sensitiveRelationships := []PersonaRelationship{
		RelationshipDeceased,
		RelationshipUnspokenLove,
	}
	
	for _, sensitive := range sensitiveRelationships {
		if relationship == sensitive {
			return true
		}
	}
	
	return false
}

// submitForReview 提交信件审核
func (s *CloudLetterService) submitForReview(ctx context.Context, letterID string) {
	log.Printf("📝 [CloudLetter] Submitting letter for review: %s", letterID)

	// 更新状态为审核中
	s.db.Model(&CloudLetter{}).Where("id = ?", letterID).Updates(map[string]interface{}{
		"status":     CloudLetterStatusUnderReview,
		"updated_at": time.Now(),
	})

	// TODO: 实现自动分配给L3/L4信使审核的逻辑
	// 这里需要与CourierService集成

	log.Printf("✅ [CloudLetter] Letter submitted for review: %s", letterID)
}

// GetCloudLetter 获取云信件详情
func (s *CloudLetterService) GetCloudLetter(ctx context.Context, userID, letterID string) (*CloudLetter, *CloudPersona, error) {
	var letter CloudLetter
	var persona CloudPersona

	// 获取信件信息
	if err := s.db.Where("id = ? AND user_id = ?", letterID, userID).First(&letter).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil, fmt.Errorf("cloud letter not found")
		}
		return nil, nil, fmt.Errorf("failed to get cloud letter: %w", err)
	}

	// 获取关联的人物信息
	if err := s.db.Where("id = ?", letter.PersonaID).First(&persona).Error; err != nil {
		return &letter, nil, fmt.Errorf("failed to get persona: %w", err)
	}

	return &letter, &persona, nil
}

// GetUserCloudLetters 获取用户的云信件列表
func (s *CloudLetterService) GetUserCloudLetters(ctx context.Context, userID string, status CloudLetterStatus) ([]*CloudLetter, error) {
	var letters []*CloudLetter
	
	query := s.db.Where("user_id = ?", userID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	
	err := query.Order("created_at DESC").Find(&letters).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get cloud letters: %w", err)
	}

	return letters, nil
}

// UpdatePersona 更新人物角色
func (s *CloudLetterService) UpdatePersona(ctx context.Context, userID, personaID string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	
	result := s.db.Model(&CloudPersona{}).
		Where("id = ? AND user_id = ?", personaID, userID).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to update persona: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("persona not found or not accessible")
	}

	return nil
}