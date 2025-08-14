package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"openpenpal-backend/internal/config"
	"openpenpal-backend/internal/models"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ContentSecurityService 内容安全服务
type ContentSecurityService struct {
	db        *gorm.DB
	config    *config.Config
	aiService *AIService
}

// SecurityCheckResult 安全检查结果
type SecurityCheckResult struct {
	IsSafe          bool                   `json:"is_safe"`
	RiskLevel       string                 `json:"risk_level"`       // low, medium, high, critical
	ViolationType   []string               `json:"violation_type"`   // spam, sensitive_info, inappropriate, harmful
	Confidence      float64                `json:"confidence"`       // 0.0-1.0
	FilteredContent string                 `json:"filtered_content"` // 过滤后的内容
	Suggestions     []string               `json:"suggestions"`      // 改进建议
	Details         map[string]interface{} `json:"details"`          // 详细信息
}

// ContentViolationRecord 内容违规记录
type ContentViolationRecord struct {
	ID            string     `json:"id" gorm:"primaryKey;type:varchar(36)"`
	UserID        string     `json:"user_id" gorm:"type:varchar(36);not null;index"`
	ContentType   string     `json:"content_type" gorm:"type:varchar(50);not null"` // letter, reply, inspiration
	ContentID     string     `json:"content_id" gorm:"type:varchar(36);index"`
	OriginalText  string     `json:"original_text" gorm:"type:text"`
	ViolationType string     `json:"violation_type" gorm:"type:varchar(100)"`
	RiskLevel     string     `json:"risk_level" gorm:"type:varchar(20)"`
	Action        string     `json:"action" gorm:"type:varchar(50)"`                          // blocked, filtered, flagged, approved
	ReviewStatus  string     `json:"review_status" gorm:"type:varchar(20);default:'pending'"` // pending, reviewed, dismissed
	CreatedAt     time.Time  `json:"created_at" gorm:"autoCreateTime"`
	ReviewedAt    *time.Time `json:"reviewed_at"`
	ReviewedBy    string     `json:"reviewed_by" gorm:"type:varchar(36)"`
}

// 敏感词库（基础版本）
var sensitiveWords = []string{
	// 个人信息相关
	"电话", "手机", "微信", "QQ", "邮箱", "地址", "身份证",
	"phone", "wechat", "email", "address", "id",

	// 商业广告相关
	"广告", "推广", "代理", "兼职", "赚钱", "投资", "理财",
	"advertisement", "promotion", "part-time", "investment",

	// 不当内容
	"暴力", "仇恨", "歧视", "政治", "宗教极端",
	"violence", "hate", "discrimination",
}

// 敏感信息正则表达式
var sensitivePatterns = []*regexp.Regexp{
	regexp.MustCompile(`\d{11}`),                                         // 手机号
	regexp.MustCompile(`\d{3}-\d{4}-\d{4}`),                              // 电话号码
	regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`), // 邮箱
	regexp.MustCompile(`\d{15}|\d{18}`),                                  // 身份证号
	// 可以添加更多模式
}

// NewContentSecurityService 创建内容安全服务实例
func NewContentSecurityService(db *gorm.DB, config *config.Config, aiService *AIService) *ContentSecurityService {
	service := &ContentSecurityService{
		db:        db,
		config:    config,
		aiService: aiService,
	}

	// 自动迁移数据表
	db.AutoMigrate(&ContentViolationRecord{})

	return service
}

// CheckContent 检查内容安全性
func (s *ContentSecurityService) CheckContent(ctx context.Context, userID, contentType, contentID, content string) (*SecurityCheckResult, error) {
	result := &SecurityCheckResult{
		IsSafe:          true,
		RiskLevel:       "low",
		ViolationType:   []string{},
		Confidence:      0.0,
		FilteredContent: content,
		Suggestions:     []string{},
		Details:         make(map[string]interface{}),
	}

	// 1. 基础规则检查
	s.checkBasicRules(content, result)

	// 2. 敏感词检查
	s.checkSensitiveWords(content, result)

	// 3. 敏感信息检查
	s.checkSensitiveInfo(content, result)

	// 4. AI安全检查（如果可用）
	if s.aiService != nil {
		err := s.checkWithAI(ctx, content, result)
		if err != nil {
			log.Printf("AI security check failed: %v", err)
		}
	}

	// 5. 计算最终风险等级
	s.calculateFinalRiskLevel(result)

	// 6. 记录违规内容
	if !result.IsSafe {
		s.recordViolation(userID, contentType, contentID, content, result)
	}

	return result, nil
}

// checkBasicRules 基础规则检查
func (s *ContentSecurityService) checkBasicRules(content string, result *SecurityCheckResult) {
	content = strings.ToLower(content)

	// 检查内容长度
	if len(content) < 10 {
		result.ViolationType = append(result.ViolationType, "too_short")
		result.Suggestions = append(result.Suggestions, "内容过短，建议增加更多表达")
	}

	if len(content) > 5000 {
		result.ViolationType = append(result.ViolationType, "too_long")
		result.Suggestions = append(result.Suggestions, "内容过长，建议精简表达")
	}

	// 检查重复字符
	if s.hasExcessiveRepetition(content) {
		result.ViolationType = append(result.ViolationType, "excessive_repetition")
		result.Suggestions = append(result.Suggestions, "避免过多重复字符或词语")
		result.Confidence += 0.3
	}

	// 检查特殊字符
	if s.hasExcessiveSpecialChars(content) {
		result.ViolationType = append(result.ViolationType, "excessive_special_chars")
		result.Suggestions = append(result.Suggestions, "减少特殊字符的使用")
		result.Confidence += 0.2
	}
}

// checkSensitiveWords 敏感词检查
func (s *ContentSecurityService) checkSensitiveWords(content string, result *SecurityCheckResult) {
	contentLower := strings.ToLower(content)
	foundWords := []string{}

	for _, word := range sensitiveWords {
		if strings.Contains(contentLower, word) {
			foundWords = append(foundWords, word)
		}
	}

	if len(foundWords) > 0 {
		result.ViolationType = append(result.ViolationType, "sensitive_words")
		result.Confidence += float64(len(foundWords)) * 0.1
		result.Details["sensitive_words"] = foundWords
		result.Suggestions = append(result.Suggestions, "请避免使用敏感词汇，保持内容的纯洁性")

		// 过滤敏感词（替换为星号）
		filteredContent := content
		for _, word := range foundWords {
			replacement := strings.Repeat("*", len(word))
			filteredContent = strings.ReplaceAll(filteredContent, word, replacement)
		}
		result.FilteredContent = filteredContent
	}
}

// checkSensitiveInfo 敏感信息检查
func (s *ContentSecurityService) checkSensitiveInfo(content string, result *SecurityCheckResult) {
	foundPatterns := []string{}

	for _, pattern := range sensitivePatterns {
		if pattern.MatchString(content) {
			foundPatterns = append(foundPatterns, pattern.String())
		}
	}

	if len(foundPatterns) > 0 {
		result.ViolationType = append(result.ViolationType, "personal_info")
		result.Confidence += 0.8 // 个人信息泄露是高风险
		result.Details["sensitive_patterns"] = foundPatterns
		result.Suggestions = append(result.Suggestions, "检测到可能的个人信息，请避免在信件中透露")

		// 过滤个人信息
		filteredContent := content
		for _, pattern := range sensitivePatterns {
			filteredContent = pattern.ReplaceAllString(filteredContent, "[已过滤]")
		}
		result.FilteredContent = filteredContent
	}
}

// checkWithAI AI安全检查
func (s *ContentSecurityService) checkWithAI(ctx context.Context, content string, result *SecurityCheckResult) error {
	// 构建AI安全检查提示词
	prompt := fmt.Sprintf(`
请分析以下文本内容的安全性，从以下维度进行评估：
1. 是否包含不当内容（暴力、仇恨、歧视等）
2. 是否包含个人隐私信息
3. 是否包含商业广告或垃圾信息
4. 整体内容是否适合在笔友平台分享

内容：%s

请返回JSON格式的分析结果：
{
  "is_safe": true/false,
  "risk_factors": ["factor1", "factor2"],
  "confidence": 0.0-1.0,
  "suggestions": ["suggestion1", "suggestion2"]
}
`, content)

	// 调用AI服务
	aiConfig, err := s.aiService.GetActiveProvider()
	if err != nil {
		return err
	}

	aiResponse, err := s.aiService.callAIAPI(ctx, aiConfig, prompt, models.TaskTypeModerate)
	if err != nil {
		return err
	}

	// 解析AI响应
	var aiResult struct {
		IsSafe      bool     `json:"is_safe"`
		RiskFactors []string `json:"risk_factors"`
		Confidence  float64  `json:"confidence"`
		Suggestions []string `json:"suggestions"`
	}

	if err := json.Unmarshal([]byte(aiResponse), &aiResult); err != nil {
		// AI响应解析失败，记录日志但不影响主流程
		log.Printf("Failed to parse AI security check response: %v", err)
		return nil
	}

	// 合并AI检查结果
	if !aiResult.IsSafe {
		result.ViolationType = append(result.ViolationType, aiResult.RiskFactors...)
		result.Confidence = max(result.Confidence, aiResult.Confidence)
		result.Suggestions = append(result.Suggestions, aiResult.Suggestions...)
		result.Details["ai_analysis"] = aiResult
	}

	return nil
}

// calculateFinalRiskLevel 计算最终风险等级
func (s *ContentSecurityService) calculateFinalRiskLevel(result *SecurityCheckResult) {
	if result.Confidence >= 0.8 {
		result.RiskLevel = "critical"
		result.IsSafe = false
	} else if result.Confidence >= 0.6 {
		result.RiskLevel = "high"
		result.IsSafe = false
	} else if result.Confidence >= 0.4 {
		result.RiskLevel = "medium"
		result.IsSafe = len(result.ViolationType) == 0 // 中等风险但无明确违规类型时可能安全
	} else {
		result.RiskLevel = "low"
		result.IsSafe = true
	}

	// 特定违规类型直接判定为不安全
	criticalViolations := []string{"personal_info", "harmful", "hate_speech"}
	for _, violation := range result.ViolationType {
		for _, critical := range criticalViolations {
			if violation == critical {
				result.IsSafe = false
				result.RiskLevel = "critical"
				return
			}
		}
	}
}

// recordViolation 记录违规内容
func (s *ContentSecurityService) recordViolation(userID, contentType, contentID, content string, result *SecurityCheckResult) {
	violationTypes := strings.Join(result.ViolationType, ",")

	action := "flagged"
	if result.RiskLevel == "critical" {
		action = "blocked"
	} else if result.RiskLevel == "high" {
		action = "filtered"
	}

	record := &ContentViolationRecord{
		ID:            uuid.New().String(),
		UserID:        userID,
		ContentType:   contentType,
		ContentID:     contentID,
		OriginalText:  content,
		ViolationType: violationTypes,
		RiskLevel:     result.RiskLevel,
		Action:        action,
		ReviewStatus:  "pending",
		CreatedAt:     time.Now(),
	}

	if err := s.db.Create(record).Error; err != nil {
		log.Printf("Failed to record content violation: %v", err)
	}
}

// GetUserViolationHistory 获取用户违规历史
func (s *ContentSecurityService) GetUserViolationHistory(userID string, limit int) ([]ContentViolationRecord, error) {
	var records []ContentViolationRecord

	query := s.db.Where("user_id = ?", userID).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&records).Error
	return records, err
}

// GetPendingReviews 获取待审核的违规内容
func (s *ContentSecurityService) GetPendingReviews(limit int) ([]ContentViolationRecord, error) {
	var records []ContentViolationRecord

	query := s.db.Where("review_status = ?", "pending").
		Order("created_at ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&records).Error
	return records, err
}

// ReviewViolation 审核违规内容
func (s *ContentSecurityService) ReviewViolation(violationID, reviewerID, action string) error {
	now := time.Now()
	return s.db.Model(&ContentViolationRecord{}).
		Where("id = ?", violationID).
		Updates(map[string]interface{}{
			"review_status": "reviewed",
			"action":        action,
			"reviewed_at":   &now,
			"reviewed_by":   reviewerID,
		}).Error
}

// 辅助方法

// hasExcessiveRepetition 检查是否有过度重复
func (s *ContentSecurityService) hasExcessiveRepetition(content string) bool {
	// 检查连续重复字符
	for i := 0; i < len(content)-2; i++ {
		if content[i] == content[i+1] && content[i+1] == content[i+2] {
			return true
		}
	}

	// 检查重复词语模式
	words := strings.Fields(content)
	if len(words) < 3 {
		return false
	}

	for i := 0; i < len(words)-2; i++ {
		if words[i] == words[i+1] && words[i+1] == words[i+2] {
			return true
		}
	}

	return false
}

// hasExcessiveSpecialChars 检查是否有过多特殊字符
func (s *ContentSecurityService) hasExcessiveSpecialChars(content string) bool {
	specialCharCount := 0
	for _, char := range content {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == ' ' || char == '\n' || char == '\t' ||
			(char >= 0x4e00 && char <= 0x9fff)) { // 中文字符范围
			specialCharCount++
		}
	}

	// 如果特殊字符超过总字符的30%，认为过多
	return float64(specialCharCount)/float64(len(content)) > 0.3
}

// max 辅助函数
func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
