package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"openpenpal-backend/internal/config"
	"openpenpal-backend/internal/models"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ModerationService 内容审核服务
type ModerationService struct {
	db        *gorm.DB
	config    *config.Config
	aiService *AIService
}

// NewModerationService 创建审核服务实例
func NewModerationService(db *gorm.DB, config *config.Config, aiService *AIService) *ModerationService {
	return &ModerationService{
		db:        db,
		config:    config,
		aiService: aiService,
	}
}

// ModerateContent 审核内容
func (s *ModerationService) ModerateContent(ctx context.Context, req *models.ModerationRequest) (*models.ModerationResponse, error) {
	// 创建审核记录
	record := &models.ModerationRecord{
		ID:            uuid.New().String(),
		ContentType:   req.ContentType,
		ContentID:     req.ContentID,
		UserID:        req.UserID,
		Content:       req.Content,
		ImageURLs:     strings.Join(req.ImageURLs, ","), // Convert slice to string
		Status:        models.ModerationPending,
		AutoModerated: true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// 1. 敏感词检查
	wordCheckResult := s.checkSensitiveWords(req.Content)
	if wordCheckResult.Level == models.LevelBlock {
		record.Status = models.ModerationRejected
		record.Level = wordCheckResult.Level
		record.Score = 1.0
		reasonsJSON, _ := json.Marshal(wordCheckResult.Reasons)
		categoriesJSON, _ := json.Marshal(wordCheckResult.Categories)
		record.Reasons = string(reasonsJSON)
		record.Categories = string(categoriesJSON)
		s.db.Create(record)

		var reasons []string
		var categories []string
		json.Unmarshal([]byte(record.Reasons), &reasons)
		json.Unmarshal([]byte(record.Categories), &categories)

		return &models.ModerationResponse{
			ID:         record.ID,
			Status:     record.Status,
			Level:      record.Level,
			Score:      record.Score,
			Reasons:    reasons,
			Categories: categories,
			NeedReview: false,
		}, nil
	}

	// 2. 规则检查
	ruleCheckResult := s.checkRules(req)
	if ruleCheckResult.Action == "block" {
		record.Status = models.ModerationRejected
		record.Level = models.LevelHigh
		record.Score = 0.9
		var existingReasons []string
		if record.Reasons != "" {
			json.Unmarshal([]byte(record.Reasons), &existingReasons)
		}
		allReasons := append(existingReasons, ruleCheckResult.Reasons...)
		reasonsJSON, _ := json.Marshal(allReasons)
		record.Reasons = string(reasonsJSON)
		s.db.Create(record)

		var reasons []string
		var categories []string
		json.Unmarshal([]byte(record.Reasons), &reasons)
		json.Unmarshal([]byte(record.Categories), &categories)

		return &models.ModerationResponse{
			ID:         record.ID,
			Status:     record.Status,
			Level:      record.Level,
			Score:      record.Score,
			Reasons:    reasons,
			Categories: categories,
			NeedReview: false,
		}, nil
	}

	// 3. AI审核（如果配置了AI服务）
	if s.config.AIProvider != "" && s.aiService != nil {
		aiResult, err := s.moderateWithAI(ctx, req)
		if err == nil {
			record.AIProvider = s.config.AIProvider
			record.AIResponse = aiResult.RawResponse
			record.Score = aiResult.Score
			categoriesJSON, _ := json.Marshal(aiResult.Categories)
			record.Categories = string(categoriesJSON)

			// 根据AI评分决定状态
			if aiResult.Score > 0.8 {
				record.Status = models.ModerationRejected
				record.Level = models.LevelHigh
			} else if aiResult.Score > 0.5 {
				record.Status = models.ModerationReview
				record.Level = models.LevelMedium
			} else {
				record.Status = models.ModerationApproved
				record.Level = models.LevelLow
			}
		}
	}

	// 4. 综合判断
	if record.Status == models.ModerationPending {
		// 如果没有AI审核，根据敏感词和规则检查结果判断
		if len(wordCheckResult.Reasons) > 0 || ruleCheckResult.Action == "review" {
			record.Status = models.ModerationReview
			record.Level = models.LevelMedium
			record.Score = 0.5
		} else {
			record.Status = models.ModerationApproved
			record.Level = models.LevelLow
			record.Score = 0.1
		}
	}

	// 保存审核记录
	if err := s.db.Create(record).Error; err != nil {
		return nil, fmt.Errorf("failed to save moderation record: %w", err)
	}

	// 如果需要人工审核，加入审核队列
	if record.Status == models.ModerationReview {
		queue := &models.ModerationQueue{
			ID:        uuid.New().String(),
			RecordID:  record.ID,
			Priority:  s.calculatePriority(record),
			Status:    "pending",
			CreatedAt: time.Now(),
		}
		s.db.Create(queue)
	}

	// Convert string fields back to slices for response
	var reasons []string
	var categories []string

	if record.Reasons != "" {
		json.Unmarshal([]byte(record.Reasons), &reasons)
	}
	if record.Categories != "" {
		json.Unmarshal([]byte(record.Categories), &categories)
	}

	return &models.ModerationResponse{
		ID:         record.ID,
		Status:     record.Status,
		Level:      record.Level,
		Score:      record.Score,
		Reasons:    reasons,
		Categories: categories,
		NeedReview: record.Status == models.ModerationReview,
	}, nil
}

// ReviewContent 人工审核内容
func (s *ModerationService) ReviewContent(ctx context.Context, req *models.ReviewRequest, reviewerID string) error {
	// 获取审核记录
	var record models.ModerationRecord
	if err := s.db.Where("id = ?", req.RecordID).First(&record).Error; err != nil {
		return fmt.Errorf("moderation record not found: %w", err)
	}

	// 更新审核状态
	now := time.Now()
	record.Status = req.Status
	record.ReviewerID = &reviewerID
	record.ReviewNote = req.ReviewNote
	record.ReviewedAt = &now
	record.AutoModerated = false
	record.UpdatedAt = now

	if err := s.db.Save(&record).Error; err != nil {
		return fmt.Errorf("failed to update moderation record: %w", err)
	}

	// 更新审核队列状态
	s.db.Model(&models.ModerationQueue{}).
		Where("record_id = ?", record.ID).
		Update("status", "completed")

	// 根据审核结果更新原始内容状态
	if err := s.updateContentStatus(&record); err != nil {
		return fmt.Errorf("failed to update content status: %w", err)
	}

	return nil
}

// GetModerationQueue 获取待审核队列
func (s *ModerationService) GetModerationQueue(limit int) ([]models.ModerationQueue, error) {
	var queue []models.ModerationQueue
	err := s.db.Preload("Record").
		Where("status = ?", "pending").
		Order("priority DESC, created_at ASC").
		Limit(limit).
		Find(&queue).Error
	return queue, err
}

// GetModerationStats 获取审核统计
func (s *ModerationService) GetModerationStats(startDate, endDate time.Time) ([]models.ModerationStats, error) {
	var stats []models.ModerationStats
	err := s.db.Where("date BETWEEN ? AND ?", startDate, endDate).
		Order("date DESC").
		Find(&stats).Error
	return stats, err
}

// 辅助方法

// checkSensitiveWords 检查敏感词
func (s *ModerationService) checkSensitiveWords(content string) struct {
	Level      models.ModerationLevel
	Reasons    []string
	Categories []string
} {
	result := struct {
		Level      models.ModerationLevel
		Reasons    []string
		Categories []string
	}{
		Level:      models.LevelLow,
		Reasons:    []string{},
		Categories: []string{},
	}

	// 获取激活的敏感词
	var words []models.SensitiveWord
	s.db.Where("is_active = ?", true).Find(&words)

	contentLower := strings.ToLower(content)
	for _, word := range words {
		if strings.Contains(contentLower, strings.ToLower(word.Word)) {
			result.Reasons = append(result.Reasons, fmt.Sprintf("包含敏感词: %s", word.Word))
			if word.Category != "" && !contains(result.Categories, word.Category) {
				result.Categories = append(result.Categories, word.Category)
			}

			// 更新风险等级
			if word.Level == models.LevelBlock {
				result.Level = models.LevelBlock
			} else if word.Level == models.LevelHigh && result.Level != models.LevelBlock {
				result.Level = models.LevelHigh
			} else if word.Level == models.LevelMedium && result.Level == models.LevelLow {
				result.Level = models.LevelMedium
			}
		}
	}

	return result
}

// checkRules 检查审核规则
func (s *ModerationService) checkRules(req *models.ModerationRequest) struct {
	Action  string
	Reasons []string
} {
	result := struct {
		Action  string
		Reasons []string
	}{
		Action:  "pass",
		Reasons: []string{},
	}

	// 获取适用的规则
	var rules []models.ModerationRule
	s.db.Where("content_type = ? AND is_active = ?", req.ContentType, true).
		Order("priority DESC").
		Find(&rules)

	for _, rule := range rules {
		matched := false

		switch rule.RuleType {
		case "keyword":
			if strings.Contains(strings.ToLower(req.Content), strings.ToLower(rule.Pattern)) {
				matched = true
			}
		case "regex":
			if regex, err := regexp.Compile(rule.Pattern); err == nil {
				if regex.MatchString(req.Content) {
					matched = true
				}
			}
		}

		if matched {
			result.Reasons = append(result.Reasons, rule.Name)
			if rule.Action == "block" {
				result.Action = "block"
				break
			} else if rule.Action == "review" && result.Action != "block" {
				result.Action = "review"
			}
		}
	}

	return result
}

// moderateWithAI 使用AI进行内容审核
func (s *ModerationService) moderateWithAI(ctx context.Context, req *models.ModerationRequest) (*AIModerateResult, error) {
	// 获取AI配置
	aiConfig, err := s.aiService.GetActiveProvider()
	if err != nil {
		return nil, err
	}

	// 构建审核提示词
	prompt := fmt.Sprintf(`
请审核以下内容是否包含违规信息：

内容类型：%s
内容：%s

请分析内容是否包含以下违规类型：
1. 政治敏感
2. 色情低俗
3. 暴力恐怖
4. 违法犯罪
5. 辱骂攻击
6. 广告营销
7. 个人隐私
8. 其他违规

返回JSON格式：
{
  "score": 0.0-1.0的风险分数,
  "categories": ["违规类型1", "违规类型2"],
  "reasons": ["具体原因1", "具体原因2"],
  "suggestion": "pass/review/block"
}
`, req.ContentType, req.Content)

	// 调用AI API
	response, err := s.aiService.callAIAPI(ctx, aiConfig, prompt, models.TaskTypeModerate)
	if err != nil {
		return nil, err
	}

	// 解析响应
	var result AIModerateResult
	result.RawResponse = response

	// 尝试解析JSON响应
	var jsonResult struct {
		Score      float64  `json:"score"`
		Categories []string `json:"categories"`
		Reasons    []string `json:"reasons"`
		Suggestion string   `json:"suggestion"`
	}

	if err := json.Unmarshal([]byte(response), &jsonResult); err == nil {
		result.Score = jsonResult.Score
		result.Categories = jsonResult.Categories
		result.Reasons = jsonResult.Reasons
		result.Suggestion = jsonResult.Suggestion
	} else {
		// 如果解析失败，使用默认值
		result.Score = 0.3
		result.Categories = []string{"未分类"}
		result.Reasons = []string{"AI审核结果解析失败，建议人工复审"}
		result.Suggestion = "review"
	}

	return &result, nil
}

// calculatePriority 计算审核优先级
func (s *ModerationService) calculatePriority(record *models.ModerationRecord) int {
	priority := 0

	// 根据风险等级
	switch record.Level {
	case models.LevelHigh:
		priority += 100
	case models.LevelMedium:
		priority += 50
	case models.LevelLow:
		priority += 10
	}

	// 根据内容类型
	switch record.ContentType {
	case models.ContentTypeLetter:
		priority += 20
	case models.ContentTypeMuseum:
		priority += 15
	case models.ContentTypeProfile:
		priority += 10
	}

	// 根据分数
	priority += int(record.Score * 100)

	return priority
}

// updateContentStatus 更新原始内容的状态
func (s *ModerationService) updateContentStatus(record *models.ModerationRecord) error {
	switch record.ContentType {
	case models.ContentTypeLetter:
		// 更新信件状态
		if record.Status == models.ModerationRejected {
			return s.db.Model(&models.Letter{}).
				Where("id = ?", record.ContentID).
				Update("status", models.StatusDraft).Error
		}
	case models.ContentTypeMuseum:
		// 更新博物馆内容状态
		if record.Status == models.ModerationApproved {
			return s.db.Model(&models.MuseumItem{}).
				Where("id = ?", record.ContentID).
				Update("status", models.MuseumItemApproved).Error
		} else if record.Status == models.ModerationRejected {
			return s.db.Model(&models.MuseumItem{}).
				Where("id = ?", record.ContentID).
				Update("status", models.MuseumItemRejected).Error
		}
	}
	return nil
}

// AddSensitiveWord 添加敏感词
func (s *ModerationService) AddSensitiveWord(req *models.SensitiveWordRequest, createdBy string) error {
	word := &models.SensitiveWord{
		ID:        uuid.New().String(),
		Word:      req.Word,
		Category:  req.Category,
		Level:     req.Level,
		IsActive:  true,
		CreatedBy: createdBy,
		CreatedAt: time.Now(),
	}
	return s.db.Create(word).Error
}

// UpdateSensitiveWord 更新敏感词
func (s *ModerationService) UpdateSensitiveWord(id string, req *models.SensitiveWordRequest) error {
	return s.db.Model(&models.SensitiveWord{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"word":     req.Word,
			"category": req.Category,
			"level":    req.Level,
		}).Error
}

// DeleteSensitiveWord 删除敏感词
func (s *ModerationService) DeleteSensitiveWord(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.SensitiveWord{}).Error
}

// GetSensitiveWords 获取敏感词列表
func (s *ModerationService) GetSensitiveWords(category string, level models.ModerationLevel) ([]models.SensitiveWord, error) {
	query := s.db.Where("is_active = ?", true)

	if category != "" {
		query = query.Where("category = ?", category)
	}

	if level != "" {
		query = query.Where("level = ?", level)
	}

	var words []models.SensitiveWord
	err := query.Order("created_at DESC").Find(&words).Error
	return words, err
}

// AddModerationRule 添加审核规则
func (s *ModerationService) AddModerationRule(req *models.ModerationRuleRequest, createdBy string) error {
	rule := &models.ModerationRule{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		ContentType: req.ContentType,
		RuleType:    req.RuleType,
		Pattern:     req.Pattern,
		Action:      req.Action,
		Priority:    req.Priority,
		IsActive:    true,
		CreatedBy:   createdBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 验证正则表达式
	if rule.RuleType == "regex" {
		if _, err := regexp.Compile(rule.Pattern); err != nil {
			return errors.New("invalid regex pattern")
		}
	}

	return s.db.Create(rule).Error
}

// UpdateModerationRule 更新审核规则
func (s *ModerationService) UpdateModerationRule(id string, req *models.ModerationRuleRequest) error {
	// 验证正则表达式
	if req.RuleType == "regex" {
		if _, err := regexp.Compile(req.Pattern); err != nil {
			return errors.New("invalid regex pattern")
		}
	}

	return s.db.Model(&models.ModerationRule{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"name":         req.Name,
			"description":  req.Description,
			"content_type": req.ContentType,
			"rule_type":    req.RuleType,
			"pattern":      req.Pattern,
			"action":       req.Action,
			"priority":     req.Priority,
			"updated_at":   time.Now(),
		}).Error
}

// DeleteModerationRule 删除审核规则
func (s *ModerationService) DeleteModerationRule(id string) error {
	return s.db.Where("id = ?", id).Delete(&models.ModerationRule{}).Error
}

// GetModerationRules 获取审核规则列表
func (s *ModerationService) GetModerationRules(contentType models.ContentType) ([]models.ModerationRule, error) {
	query := s.db.Where("is_active = ?", true)

	if contentType != "" {
		query = query.Where("content_type = ?", contentType)
	}

	var rules []models.ModerationRule
	err := query.Order("priority DESC, created_at DESC").Find(&rules).Error
	return rules, err
}

// 辅助结构体

// AIModerateResult AI审核结果
type AIModerateResult struct {
	Score       float64
	Categories  []string
	Reasons     []string
	Suggestion  string
	RawResponse string
}

// 工具函数

// contains 检查字符串数组是否包含某个元素
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
