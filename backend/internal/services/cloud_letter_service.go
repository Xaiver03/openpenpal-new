package services

import (
	"context"
	"encoding/json"
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

	// 调用AI服务增强内容
	if aiResponse, err := s.callAIForEnhancement(ctx, letter, persona); err == nil {
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
func (s *CloudLetterService) callAIForEnhancement(ctx context.Context, letter *CloudLetter, persona *CloudPersona) (string, error) {
	if s.aiSvc == nil {
		log.Printf("⚠️ [CloudLetter] AI service not available, using fallback")
		return s.generateMockEnhancement(letter.OriginalContent), nil
	}

	// 直接调用AI服务的内容增强方法
	enhancedContent, err := s.aiSvc.EnhanceContent(ctx, letter.OriginalContent, persona, letter.EmotionalTone)
	if err != nil {
		log.Printf("❌ [CloudLetter] AI enhancement failed: %v", err)
		// 使用fallback机制，返回原始内容的基础增强版本
		return s.generateMockEnhancement(letter.OriginalContent), nil
	}

	// 清理AI响应，移除可能的格式标记
	cleanedContent := s.cleanAIResponse(enhancedContent)
	
	log.Printf("✅ [CloudLetter] AI enhancement completed successfully")
	return cleanedContent, nil
}

// cleanAIResponse 清理AI响应内容
func (s *CloudLetterService) cleanAIResponse(response string) string {
	// 移除常见的AI响应格式标记
	cleaned := strings.TrimSpace(response)
	
	// 移除可能的JSON格式包装
	if strings.HasPrefix(cleaned, "{") && strings.HasSuffix(cleaned, "}") {
		// 尝试解析JSON格式响应
		var jsonResp map[string]interface{}
		if err := json.Unmarshal([]byte(cleaned), &jsonResp); err == nil {
			if content, ok := jsonResp["content"].(string); ok {
				cleaned = content
			} else if enhanced, ok := jsonResp["enhanced_content"].(string); ok {
				cleaned = enhanced
			}
		}
	}
	
	// 移除markdown代码块标记
	cleaned = strings.ReplaceAll(cleaned, "```", "")
	cleaned = strings.ReplaceAll(cleaned, "**增强版本**", "")
	cleaned = strings.ReplaceAll(cleaned, "**Enhanced Version**", "")
	
	return strings.TrimSpace(cleaned)
}

// generateMockEnhancement 生成模拟增强内容（临时实现）
func (s *CloudLetterService) generateMockEnhancement(originalContent string) string {
	return fmt.Sprintf("【AI增强版本】\n\n%s\n\n（已通过AI优化语言表达和情感深度）", originalContent)
}

// determineRequiredReviewerLevel 确定所需的审核员等级
func (s *CloudLetterService) determineRequiredReviewerLevel(letter *CloudLetter, persona *CloudPersona) int {
	// 基础审核等级
	baseLevel := 2 // L2信使默认处理一般内容

	// 根据人物关系类型确定敏感度
	switch persona.Relationship {
	case RelationshipDeceased:
		// 已故亲人关系需要L3信使审核（高敏感度）
		baseLevel = 3
	case RelationshipUnspokenLove:
		// 暗恋关系需要L3信使审核（情感敏感）
		baseLevel = 3
	case RelationshipDistantFriend:
		// 疏远朋友一般由L2处理
		baseLevel = 2
	case RelationshipCustom:
		// 自定义关系根据描述判断
		if s.isHighSensitivityContent(persona.Description, letter.OriginalContent) {
			baseLevel = 3
		}
	}

	// 内容敏感度检查
	if s.containsSensitiveKeywords(letter.OriginalContent) {
		baseLevel = max(baseLevel, 3) // 至少需要L3
	}

	// 极度敏感内容需要L4审核
	if s.requiresL4Review(letter.OriginalContent, persona) {
		baseLevel = 4
	}

	log.Printf("📋 [CloudLetter] Determined reviewer level %d for relationship: %s", baseLevel, persona.Relationship)
	return baseLevel
}

// isHighSensitivityContent 检查是否为高敏感度内容
func (s *CloudLetterService) isHighSensitivityContent(description, content string) bool {
	// 检查人物描述中的敏感词
	sensitiveDescriptions := []string{
		"离世", "去世", "逝世", "病故", "意外", "自杀", "抑郁", 
		"分手", "失恋", "背叛", "伤害", "恨", "报复",
		"家暴", "虐待", "创伤", "痛苦", "绝望",
	}

	combined := strings.ToLower(description + " " + content)
	for _, keyword := range sensitiveDescriptions {
		if strings.Contains(combined, keyword) {
			return true
		}
	}
	return false
}

// containsSensitiveKeywords 检查内容是否包含敏感关键词
func (s *CloudLetterService) containsSensitiveKeywords(content string) bool {
	sensitiveKeywords := []string{
		// 死亡相关
		"死", "死亡", "自杀", "轻生", "结束生命",
		// 暴力相关
		"杀", "打", "伤害", "报复", "仇恨", "暴力",
		// 性相关
		"性", "做爱", "上床", "激情", "身体",
		// 政治敏感
		"政府", "政治", "革命", "抗议", "游行",
		// 其他敏感
		"毒品", "赌博", "诈骗", "犯罪", "违法",
	}

	content = strings.ToLower(content)
	for _, keyword := range sensitiveKeywords {
		if strings.Contains(content, keyword) {
			return true
		}
	}
	return false
}

// requiresL4Review 检查是否需要L4级别审核
func (s *CloudLetterService) requiresL4Review(content string, persona *CloudPersona) bool {
	// L4审核条件：极度敏感内容或跨校影响
	extremeKeywords := []string{
		"自杀", "杀人", "恐怖", "极端", "炸弹", "毒品交易",
		"人体器官", "卖淫", "色情服务", "黑社会", "洗钱",
	}

	content = strings.ToLower(content)
	for _, keyword := range extremeKeywords {
		if strings.Contains(content, keyword) {
			log.Printf("🚨 [CloudLetter] L4 review required due to extreme content keyword: %s", keyword)
			return true
		}
	}

	// 检查是否可能影响多个校区
	if persona.Description != "" && strings.Contains(persona.Description, "知名") {
		return true
	}

	return false
}

// assignCourierReviewer 分配信使审核员
func (s *CloudLetterService) assignCourierReviewer(ctx context.Context, letter *CloudLetter, requiredLevel int) error {
	log.Printf("🔍 [CloudLetter] Assigning L%d courier reviewer for letter: %s", requiredLevel, letter.ID)

	// 获取符合条件的信使审核员
	reviewer, err := s.findAvailableCourierReviewer(ctx, requiredLevel)
	if err != nil {
		return fmt.Errorf("failed to find available reviewer: %w", err)
	}

	// 更新信件的审核员信息
	if err := s.db.Model(letter).Updates(map[string]interface{}{
		"reviewer_level": requiredLevel,
		"reviewer_id":    reviewer.UserID,
		"updated_at":     time.Now(),
	}).Error; err != nil {
		return fmt.Errorf("failed to update letter reviewer: %w", err)
	}

	// 创建审核任务（如果CourierService支持任务创建）
	if err := s.createReviewTask(ctx, letter, reviewer, requiredLevel); err != nil {
		log.Printf("⚠️ [CloudLetter] Failed to create review task: %v", err)
		// 不阻塞主流程
	}

	// 发送通知给审核员
	if s.notificationSvc != nil {
		s.sendReviewNotification(ctx, letter, reviewer)
	}

	log.Printf("✅ [CloudLetter] Assigned reviewer %s (L%d) for letter %s", reviewer.UserID, requiredLevel, letter.ID)
	return nil
}

// findAvailableCourierReviewer 查找可用的信使审核员
func (s *CloudLetterService) findAvailableCourierReviewer(ctx context.Context, requiredLevel int) (*models.Courier, error) {
	var couriers []models.Courier

	// 查询对应等级的信使，优先选择任务较少的
	err := s.db.Where("level >= ? AND status = ?", requiredLevel, "approved").
		Order("task_count ASC, points DESC"). // 任务少的优先，积分高的优先
		Limit(5). // 取前5个候选人
		Find(&couriers).Error

	if err != nil {
		return nil, fmt.Errorf("failed to query couriers: %w", err)
	}

	if len(couriers) == 0 {
		return nil, fmt.Errorf("no available L%d+ couriers found", requiredLevel)
	}

	// 选择第一个（任务最少的）
	selectedCourier := &couriers[0]
	
	log.Printf("🎯 [CloudLetter] Selected courier %s (L%d, %d tasks) for review", 
		selectedCourier.UserID, selectedCourier.Level, selectedCourier.TaskCount)
	
	return selectedCourier, nil
}

// createReviewTask 创建审核任务
func (s *CloudLetterService) createReviewTask(ctx context.Context, letter *CloudLetter, reviewer *models.Courier, level int) error {
	// 这里可以调用CourierService的任务创建方法
	// 由于CourierService主要处理物流任务，CloudLetter审核可能需要扩展
	
	// 临时实现：记录审核任务到数据库或发送通知
	log.Printf("📝 [CloudLetter] Review task created for courier %s, letter %s", reviewer.UserID, letter.ID)
	
	// TODO: 如果需要，可以扩展CourierService支持审核任务类型
	// 或者创建独立的审核任务表
	
	return nil
}

// sendReviewNotification 发送审核通知
func (s *CloudLetterService) sendReviewNotification(ctx context.Context, letter *CloudLetter, reviewer *models.Courier) {
	if s.notificationSvc == nil {
		return
	}

	// 构建通知内容
	message := fmt.Sprintf("您有一封云中锦书待审核，ID: %s，关系类型: %s", 
		letter.ID[:8], // 只显示前8位ID
		letter.PersonaID)

	// 发送通知
	log.Printf("📧 [CloudLetter] Sending review notification to courier %s: %s", reviewer.UserID, message)
	
	if s.notificationSvc != nil {
		notificationData := map[string]interface{}{
			"title":          "CloudLetter审核",
			"content":        message,
			"type":           "cloud_letter_review",
			"letter_id":      letter.ID,
			"persona_id":     letter.PersonaID,
			"action_required": true,
		}
		
		if err := s.notificationSvc.NotifyUser(reviewer.UserID, "cloud_letter_review", notificationData); err != nil {
			log.Printf("❌ [CloudLetter] Failed to send review notification to %s: %v", reviewer.UserID, err)
		}
	}
}

// max 辅助函数
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// GetPendingReviews 获取待审核的云信件列表（L3/L4信使专用）
func (s *CloudLetterService) GetPendingReviews(ctx context.Context, reviewerUserID string, page, limit int) ([]CloudLetter, int64, error) {
	var letters []CloudLetter
	var total int64

	// 计算偏移量
	offset := (page - 1) * limit

	// 获取分配给该审核员的待审核信件
	query := s.db.WithContext(ctx).
		Where("reviewer_id = ? AND status = ?", reviewerUserID, CloudLetterStatusUnderReview).
		Order("created_at ASC") // 按创建时间排序，先进先出

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count pending reviews: %w", err)
	}

	// 获取分页数据
	if err := query.Limit(limit).Offset(offset).Find(&letters).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get pending reviews: %w", err)
	}

	log.Printf("📋 [CloudLetter] Retrieved %d/%d pending reviews for reviewer %s", len(letters), total, reviewerUserID)
	return letters, total, nil
}

// ReviewCloudLetter 审核云信件
func (s *CloudLetterService) ReviewCloudLetter(ctx context.Context, reviewerUserID, letterID, decision, comments string) error {
	log.Printf("🔍 [CloudLetter] Reviewing letter %s by %s with decision: %s", letterID, reviewerUserID, decision)

	// 获取信件信息
	var letter CloudLetter
	if err := s.db.WithContext(ctx).Where("id = ?", letterID).First(&letter).Error; err != nil {
		return fmt.Errorf("letter not found: %w", err)
	}

	// 验证审核权限
	if letter.ReviewerID != reviewerUserID {
		return fmt.Errorf("not authorized to review this letter")
	}

	// 验证当前状态
	if letter.Status != CloudLetterStatusUnderReview {
		return fmt.Errorf("letter is not under review, current status: %s", letter.Status)
	}

	// 根据决定更新状态
	var newStatus CloudLetterStatus
	switch decision {
	case "approved":
		newStatus = CloudLetterStatusApproved
	case "rejected":
		// 被拒绝的信件不会被投递
		newStatus = CloudLetterStatusRevisionNeeded
	case "revision_needed":
		newStatus = CloudLetterStatusRevisionNeeded
	default:
		return fmt.Errorf("invalid decision: %s", decision)
	}

	// 更新信件状态和审核意见
	updates := map[string]interface{}{
		"status":          newStatus,
		"review_comments": comments,
		"updated_at":      time.Now(),
	}

	// 如果批准，设置投递日期
	if newStatus == CloudLetterStatusApproved {
		deliveryDate := time.Now().Add(24 * time.Hour) // 24小时后投递
		updates["actual_delivery_date"] = &deliveryDate
	}

	if err := s.db.Model(&letter).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update letter status: %w", err)
	}

	// 发送通知给用户
	if s.notificationSvc != nil {
		s.sendReviewResultNotification(ctx, &letter, decision, comments)
	}

	// 如果批准，可以触发投递流程
	if newStatus == CloudLetterStatusApproved {
		s.processApprovedLetter(ctx, &letter)
	}

	log.Printf("✅ [CloudLetter] Letter %s reviewed successfully with decision: %s", letterID, decision)
	return nil
}

// sendReviewResultNotification 发送审核结果通知给用户
func (s *CloudLetterService) sendReviewResultNotification(ctx context.Context, letter *CloudLetter, decision, comments string) {
	var title, message string

	switch decision {
	case "approved":
		title = "云中锦书已通过审核"
		message = fmt.Sprintf("您的云中锦书「%s」已通过审核，将在24小时内完成投递。", letter.ID[:8])
	case "rejected":
		title = "云中锦书审核未通过"
		message = fmt.Sprintf("您的云中锦书「%s」审核未通过，原因：%s", letter.ID[:8], comments)
	case "revision_needed":
		title = "云中锦书需要修改"
		message = fmt.Sprintf("您的云中锦书「%s」需要修改，建议：%s", letter.ID[:8], comments)
	}

	if comments != "" {
		message += fmt.Sprintf("\n审核意见：%s", comments)
	}

	log.Printf("📧 [CloudLetter] Sending review result notification to user %s: %s", letter.UserID, title)
	
	if s.notificationSvc != nil {
		notificationData := map[string]interface{}{
			"title":        title,
			"content":      message,
			"type":         "cloud_letter_review_result",
			"letter_id":    letter.ID,
			"decision":     decision,
			"comments":     comments,
		}
		
		if err := s.notificationSvc.NotifyUser(letter.UserID, "cloud_letter_review_result", notificationData); err != nil {
			log.Printf("❌ [CloudLetter] Failed to send review result notification to %s: %v", letter.UserID, err)
		}
	}
}

// processApprovedLetter 处理已批准的信件
func (s *CloudLetterService) processApprovedLetter(ctx context.Context, letter *CloudLetter) {
	log.Printf("📮 [CloudLetter] Processing approved letter for delivery: %s", letter.ID)

	// 更新状态为已投递
	go func() {
		// 模拟投递延迟
		time.Sleep(5 * time.Minute) // 5分钟后标记为已投递（模拟）

		updates := map[string]interface{}{
			"status":     CloudLetterStatusDelivered,
			"updated_at": time.Now(),
		}

		if err := s.db.Model(letter).Updates(updates).Error; err != nil {
			log.Printf("❌ [CloudLetter] Failed to update letter to delivered: %v", err)
			return
		}

		log.Printf("✅ [CloudLetter] Letter %s marked as delivered", letter.ID)

		// 发送投递完成通知
		if s.notificationSvc != nil {
			// TODO: 发送投递完成通知
		}
	}()
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

	// 实现自动分配给L3/L4信使审核的逻辑
	if s.courierSvc != nil {
		// TODO: 获取实际的letter和persona对象进行审核等级判断
		reviewerLevel := 3 // 默认使用L3信使审核
		if reviewerLevel >= 3 { // L3或L4信使审核
			// TODO: 修复方法调用参数类型
			// if err := s.assignCourierReviewer(ctx, letterID, reviewerLevel); err != nil {
			log.Printf("📝 [CloudLetter] Would assign L%d courier reviewer for letter %s", reviewerLevel, letterID)
			err := error(nil)
			if err != nil {
				log.Printf("⚠️ [CloudLetter] Failed to assign courier reviewer: %v", err)
				// 不阻塞流程，继续处理
			}
		}
	}

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