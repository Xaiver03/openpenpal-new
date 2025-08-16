package services

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"log"
	"openpenpal-backend/internal/config"
	"openpenpal-backend/internal/models"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/microcosm-cc/bluemonday"
	"gorm.io/gorm"
)

// ContentSecurityService 内容安全服务 - 增强XSS防护版本
type ContentSecurityService struct {
	db                *gorm.DB
	config            *config.Config
	aiService         *AIService
	htmlSanitizer     *bluemonday.Policy // 标准HTML清理器
	strictSanitizer   *bluemonday.Policy // 严格HTML清理器
	xssPatterns       []*regexp.Regexp   // XSS攻击模式
	maxContentLength  int                // 最大内容长度
	sensitiveWords    map[string]bool    // 敏感词映射表
}

// SecurityCheckResult 安全检查结果 - 增强XSS检测版本
type SecurityCheckResult struct {
	IsSafe              bool                   `json:"is_safe"`
	RiskLevel           string                 `json:"risk_level"`           // low, medium, high, critical
	ViolationType       []string               `json:"violation_type"`       // spam, sensitive_info, inappropriate, harmful, xss_attempt
	Confidence          float64                `json:"confidence"`           // 0.0-1.0
	FilteredContent     string                 `json:"filtered_content"`     // 过滤后的内容
	SanitizedContent    string                 `json:"sanitized_content"`    // HTML清理后的内容
	Suggestions         []string               `json:"suggestions"`          // 改进建议
	Details             map[string]interface{} `json:"details"`              // 详细信息
	XSSDetected         bool                   `json:"xss_detected"`         // 是否检测到XSS
	HTMLCleaned         bool                   `json:"html_cleaned"`         // 是否进行了HTML清理
	RequiresModeration  bool                   `json:"requires_moderation"`  // 是否需要人工审核
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

// NewContentSecurityService 创建内容安全服务实例 - 增强XSS防护版本
func NewContentSecurityService(db *gorm.DB, config *config.Config, aiService *AIService) *ContentSecurityService {
	service := &ContentSecurityService{
		db:               db,
		config:           config,
		aiService:        aiService,
		maxContentLength: 10000, // 10KB 最大内容长度
	}

	// 初始化HTML清理器
	service.initializeHTMLSanitizers()
	
	// 初始化XSS检测模式
	service.initializeXSSPatterns()

	// 自动迁移数据表
	db.AutoMigrate(&ContentViolationRecord{})

	return service
}

// initializeHTMLSanitizers 初始化HTML清理器
func (s *ContentSecurityService) initializeHTMLSanitizers() {
	// 标准清理器 - 允许安全的HTML标签用于富文本内容
	s.htmlSanitizer = bluemonday.UGCPolicy()
	
	// 允许的安全标签和属性
	s.htmlSanitizer.AllowElements("b", "i", "em", "strong", "u", "br", "p", "div", "span", "h1", "h2", "h3", "h4", "h5", "h6")
	s.htmlSanitizer.AllowAttrs("class").OnElements("span", "div", "p")
	s.htmlSanitizer.AllowURLSchemes("http", "https")
	
	// 移除所有脚本相关的属性和标签
	s.htmlSanitizer.AllowNoAttrs().OnElements("script", "style", "link", "meta")
	
	// 严格清理器 - 仅允许纯文本，用于评论等
	s.strictSanitizer = bluemonday.StrictPolicy()
}

// initializeXSSPatterns 初始化XSS检测模式
func (s *ContentSecurityService) initializeXSSPatterns() {
	patterns := []string{
		// JavaScript注入模式
		`(?i)<script[^>]*>.*?</script>`,
		`(?i)javascript\s*:`,
		`(?i)on\w+\s*=`,
		`(?i)eval\s*\(`,
		`(?i)expression\s*\(`,
		`(?i)document\.cookie`,
		`(?i)window\.location`,
		`(?i)alert\s*\(`,
		`(?i)confirm\s*\(`,
		`(?i)prompt\s*\(`,
		
		// HTML注入模式
		`(?i)<iframe[^>]*>`,
		`(?i)<object[^>]*>`,
		`(?i)<embed[^>]*>`,
		`(?i)<applet[^>]*>`,
		`(?i)<form[^>]*>`,
		`(?i)<meta[^>]*>`,
		`(?i)<link[^>]*>`,
		`(?i)<style[^>]*>`,
		
		// 数据URI和危险协议
		`(?i)data\s*:\s*text/html`,
		`(?i)data\s*:\s*application/javascript`,
		`(?i)vbscript\s*:`,
		`(?i)about\s*:`,
		`(?i)file\s*:`,
		
		// 编码绕过模式
		`(?i)&#x[0-9a-f]+;`,
		`(?i)&#[0-9]+;`,
		`(?i)%[0-9a-f]{2}`,
		
		// CSS注入模式
		`(?i)expression\s*\(`,
		`(?i)@import`,
		`(?i)behavior\s*:`,
	}
	
	s.xssPatterns = make([]*regexp.Regexp, 0, len(patterns))
	for _, pattern := range patterns {
		if regex, err := regexp.Compile(pattern); err == nil {
			s.xssPatterns = append(s.xssPatterns, regex)
		} else {
			log.Printf("Failed to compile XSS pattern: %s, error: %v", pattern, err)
		}
	}
}

// CheckContent 检查内容安全性 - 增强XSS防护版本
func (s *ContentSecurityService) CheckContent(ctx context.Context, userID, contentType, contentID, content string) (*SecurityCheckResult, error) {
	result := &SecurityCheckResult{
		IsSafe:             true,
		RiskLevel:          "low",
		ViolationType:      []string{},
		Confidence:         0.0,
		FilteredContent:    content,
		SanitizedContent:   content,
		Suggestions:        []string{},
		Details:            make(map[string]interface{}),
		XSSDetected:        false,
		HTMLCleaned:        false,
		RequiresModeration: false,
	}

	// 1. 内容长度检查
	if len(content) > s.maxContentLength {
		result.ViolationType = append(result.ViolationType, "content_too_long")
		result.Suggestions = append(result.Suggestions, fmt.Sprintf("内容长度超过限制 (%d > %d)", len(content), s.maxContentLength))
		result.Confidence += 0.3
	}

	// 2. XSS检测（优先级最高）
	s.checkXSSAttempts(content, result)

	// 3. HTML清理和安全化
	s.sanitizeContent(content, contentType, result)

	// 4. 基础规则检查
	s.checkBasicRules(content, result)

	// 5. 敏感词检查
	s.checkSensitiveWords(content, result)

	// 6. 敏感信息检查
	s.checkSensitiveInfo(content, result)

	// 7. AI安全检查（如果可用）
	if s.aiService != nil {
		err := s.checkWithAI(ctx, content, result)
		if err != nil {
			log.Printf("AI security check failed: %v", err)
		}
	}

	// 8. 计算最终风险等级
	s.calculateFinalRiskLevel(result)

	// 9. 记录违规内容
	if !result.IsSafe || result.RequiresModeration {
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
		result.Confidence = maxFloat(result.Confidence, aiResult.Confidence)
		result.Suggestions = append(result.Suggestions, aiResult.Suggestions...)
		result.Details["ai_analysis"] = aiResult
	}

	return nil
}

// checkXSSAttempts XSS攻击检测
func (s *ContentSecurityService) checkXSSAttempts(content string, result *SecurityCheckResult) {
	xssFound := false
	detectedPatterns := []string{}

	for _, pattern := range s.xssPatterns {
		if matches := pattern.FindAllString(content, -1); len(matches) > 0 {
			xssFound = true
			for _, match := range matches {
				detectedPatterns = append(detectedPatterns, match)
			}
		}
	}

	if xssFound {
		result.XSSDetected = true
		result.ViolationType = append(result.ViolationType, "xss_attempt")
		result.Confidence += 0.8 // XSS攻击是高风险
		result.Suggestions = append(result.Suggestions, "检测到可能的XSS攻击尝试，内容已被清理")
		result.Details["xss_patterns"] = detectedPatterns
		
		log.Printf("XSS attempt detected for content: %s, patterns: %v", content[:min(len(content), 100)], detectedPatterns)
	}
}

// sanitizeContent HTML清理和内容安全化
func (s *ContentSecurityService) sanitizeContent(content, contentType string, result *SecurityCheckResult) {
	originalContent := content
	
	// 根据内容类型选择合适的清理策略
	var cleanedContent string
	if contentType == "comment" || contentType == "reply" {
		// 评论使用严格模式，仅允许纯文本
		cleanedContent = s.strictSanitizer.Sanitize(content)
	} else {
		// 信件等内容使用标准模式，允许安全的HTML
		cleanedContent = s.htmlSanitizer.Sanitize(content)
	}
	
	// 额外的安全清理
	cleanedContent = s.additionalSecurityCleaning(cleanedContent)
	
	// 更新结果
	result.SanitizedContent = cleanedContent
	result.HTMLCleaned = originalContent != cleanedContent
	
	if result.HTMLCleaned {
		result.Details["html_changes"] = map[string]string{
			"original": originalContent,
			"cleaned":  cleanedContent,
		}
		log.Printf("Content sanitized: %d chars removed", len(originalContent)-len(cleanedContent))
	}
}

// additionalSecurityCleaning 额外的安全清理
func (s *ContentSecurityService) additionalSecurityCleaning(content string) string {
	// 1. HTML实体编码潜在危险字符
	content = html.EscapeString(content)
	
	// 2. 移除控制字符和零宽度字符
	content = regexp.MustCompile(`[\x00-\x08\x0B\x0C\x0E-\x1F\x7F\uFEFF\u200B-\u200D\u2060]`).ReplaceAllString(content, "")
	
	// 3. 标准化空白字符
	content = regexp.MustCompile(`\s+`).ReplaceAllString(content, " ")
	content = regexp.MustCompile(`\n{3,}`).ReplaceAllString(content, "\n\n")
	
	// 4. 移除可疑的Unicode字符（可能用于绕过过滤）
	content = regexp.MustCompile(`[\u2028\u2029]`).ReplaceAllString(content, " ")
	
	// 5. 清理多余的标点符号
	content = regexp.MustCompile(`[!@#$%^&*()]{4,}`).ReplaceAllString(content, "***")
	
	return strings.TrimSpace(content)
}

// calculateFinalRiskLevel 计算最终风险等级 - 增强版本
func (s *ContentSecurityService) calculateFinalRiskLevel(result *SecurityCheckResult) {
	// XSS攻击直接判定为critical
	if result.XSSDetected {
		result.RiskLevel = "critical"
		result.IsSafe = false
		result.RequiresModeration = true
		return
	}

	if result.Confidence >= 0.8 {
		result.RiskLevel = "critical"
		result.IsSafe = false
		result.RequiresModeration = true
	} else if result.Confidence >= 0.6 {
		result.RiskLevel = "high"
		result.IsSafe = false
		result.RequiresModeration = true
	} else if result.Confidence >= 0.4 {
		result.RiskLevel = "medium"
		result.IsSafe = len(result.ViolationType) <= 1 // 中等风险，少量违规可能安全
		result.RequiresModeration = result.Confidence >= 0.5
	} else {
		result.RiskLevel = "low"
		result.IsSafe = true
		result.RequiresModeration = false
	}

	// 特定违规类型直接判定为不安全
	criticalViolations := []string{"personal_info", "harmful", "hate_speech", "xss_attempt"}
	for _, violation := range result.ViolationType {
		for _, critical := range criticalViolations {
			if violation == critical {
				result.IsSafe = false
				result.RequiresModeration = true
				if violation == "xss_attempt" || violation == "harmful" {
					result.RiskLevel = "critical"
				} else if result.RiskLevel == "low" || result.RiskLevel == "medium" {
					result.RiskLevel = "high"
				}
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

// maxFloat 辅助函数
func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// min 辅助函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// =============== 新增：专门用于评论系统的安全方法 ===============

// ValidateCommentContent 专门用于评论内容验证
func (s *ContentSecurityService) ValidateCommentContent(content, userID string) (*SecurityCheckResult, error) {
	if strings.TrimSpace(content) == "" {
		return nil, fmt.Errorf("评论内容不能为空")
	}
	
	// 评论内容使用最严格的安全检查
	return s.CheckContent(context.Background(), userID, "comment", "", content)
}

// ValidateLetterContent 专门用于信件内容验证
func (s *ContentSecurityService) ValidateLetterContent(content, userID string) (*SecurityCheckResult, error) {
	if strings.TrimSpace(content) == "" {
		return nil, fmt.Errorf("信件内容不能为空")
	}
	
	// 信件内容允许更多格式，但仍需安全检查
	return s.CheckContent(context.Background(), userID, "letter", "", content)
}

// GetSecurityStatistics 获取安全统计信息
func (s *ContentSecurityService) GetSecurityStatistics(days int) (map[string]interface{}, error) {
	if days <= 0 {
		days = 7 // 默认7天
	}
	
	since := time.Now().AddDate(0, 0, -days)
	
	stats := make(map[string]interface{})
	
	// 总违规事件数
	var totalViolations int64
	if err := s.db.Model(&ContentViolationRecord{}).
		Where("created_at >= ?", since).
		Count(&totalViolations).Error; err != nil {
		return nil, fmt.Errorf("获取总违规数失败: %w", err)
	}
	stats["total_violations"] = totalViolations
	
	// XSS攻击尝试数
	var xssAttempts int64
	if err := s.db.Model(&ContentViolationRecord{}).
		Where("created_at >= ? AND violation_type LIKE ?", since, "%xss_attempt%").
		Count(&xssAttempts).Error; err != nil {
		return nil, fmt.Errorf("获取XSS攻击数失败: %w", err)
	}
	stats["xss_attempts"] = xssAttempts
	
	// 被阻止的内容数
	var blockedContent int64
	if err := s.db.Model(&ContentViolationRecord{}).
		Where("created_at >= ? AND action = ?", since, "blocked").
		Count(&blockedContent).Error; err != nil {
		return nil, fmt.Errorf("获取被阻止内容数失败: %w", err)
	}
	stats["blocked_content"] = blockedContent
	
	// 需要审核的内容数
	var pendingReview int64
	if err := s.db.Model(&ContentViolationRecord{}).
		Where("review_status = ?", "pending").
		Count(&pendingReview).Error; err != nil {
		return nil, fmt.Errorf("获取待审核内容数失败: %w", err)
	}
	stats["pending_review"] = pendingReview
	
	// 违规类型分布
	var violationTypes []struct {
		ViolationType string `json:"violation_type"`
		Count         int64  `json:"count"`
	}
	if err := s.db.Model(&ContentViolationRecord{}).
		Select("violation_type, COUNT(*) as count").
		Where("created_at >= ?", since).
		Group("violation_type").
		Order("count DESC").
		Limit(10).
		Find(&violationTypes).Error; err != nil {
		return nil, fmt.Errorf("获取违规类型分布失败: %w", err)
	}
	stats["violation_types"] = violationTypes
	
	// 风险等级分布
	var riskLevels []struct {
		RiskLevel string `json:"risk_level"`
		Count     int64  `json:"count"`
	}
	if err := s.db.Model(&ContentViolationRecord{}).
		Select("risk_level, COUNT(*) as count").
		Where("created_at >= ?", since).
		Group("risk_level").
		Find(&riskLevels).Error; err != nil {
		return nil, fmt.Errorf("获取风险等级分布失败: %w", err)
	}
	stats["risk_levels"] = riskLevels
	
	stats["period_days"] = days
	stats["generated_at"] = time.Now()
	
	return stats, nil
}

// loadSensitiveWords 从数据库加载敏感词到内存
func (s *ContentSecurityService) loadSensitiveWords() error {
	var words []models.SensitiveWord
	if err := s.db.Where("is_active = ?", true).Find(&words).Error; err != nil {
		log.Printf("Failed to load sensitive words: %v", err)
		return err
	}
	
	for _, word := range words {
		s.sensitiveWords[strings.ToLower(word.Word)] = true
	}
	
	log.Printf("Loaded %d sensitive words into memory", len(words))
	return nil
}

// RefreshSensitiveWords 刷新敏感词库
func (s *ContentSecurityService) RefreshSensitiveWords() error {
	s.sensitiveWords = make(map[string]bool)
	return s.loadSensitiveWords()
}

// =============== 敏感词管理API方法 ===============

// GetSensitiveWords 获取敏感词列表
func (s *ContentSecurityService) GetSensitiveWords(page, limit int, category, isActive string) ([]models.SensitiveWord, int64, error) {
	var words []models.SensitiveWord
	var total int64
	
	query := s.db.Model(&models.SensitiveWord{})
	
	// 过滤条件
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if isActive != "" {
		if isActive == "true" {
			query = query.Where("is_active = ?", true)
		} else if isActive == "false" {
			query = query.Where("is_active = ?", false)
		}
	}
	
	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("获取敏感词总数失败: %w", err)
	}
	
	// 分页查询
	offset := (page - 1) * limit
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&words).Error; err != nil {
		return nil, 0, fmt.Errorf("获取敏感词列表失败: %w", err)
	}
	
	return words, total, nil
}

// UpdateSensitiveWord 更新敏感词
func (s *ContentSecurityService) UpdateSensitiveWord(id, word, category, level string) error {
	updates := map[string]interface{}{
		"word":       strings.ToLower(word),
		"category":   category,
		"level":      level,
		"updated_at": time.Now(),
	}
	
	if err := s.db.Model(&models.SensitiveWord{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return fmt.Errorf("更新敏感词失败: %w", err)
	}
	
	// 刷新内存中的敏感词库
	s.RefreshSensitiveWords()
	
	return nil
}

// DeleteSensitiveWord 删除敏感词（软删除）
func (s *ContentSecurityService) DeleteSensitiveWord(id string) error {
	// 获取要删除的敏感词
	var word models.SensitiveWord
	if err := s.db.Where("id = ?", id).First(&word).Error; err != nil {
		return fmt.Errorf("敏感词不存在: %w", err)
	}
	
	// 软删除：设置为非活跃状态
	if err := s.db.Model(&models.SensitiveWord{}).Where("id = ?", id).Update("is_active", false).Error; err != nil {
		return fmt.Errorf("删除敏感词失败: %w", err)
	}
	
	// 从内存中移除
	delete(s.sensitiveWords, word.Word)
	
	return nil
}

// GetSensitiveWordStats 获取敏感词统计信息
func (s *ContentSecurityService) GetSensitiveWordStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// 总敏感词数
	var totalWords int64
	if err := s.db.Model(&models.SensitiveWord{}).Count(&totalWords).Error; err != nil {
		return nil, fmt.Errorf("获取总敏感词数失败: %w", err)
	}
	stats["total_words"] = totalWords
	
	// 活跃敏感词数
	var activeWords int64
	if err := s.db.Model(&models.SensitiveWord{}).Where("is_active = ?", true).Count(&activeWords).Error; err != nil {
		return nil, fmt.Errorf("获取活跃敏感词数失败: %w", err)
	}
	stats["active_words"] = activeWords
	
	// 按分类统计
	var categoryStats []struct {
		Category string `json:"category"`
		Count    int64  `json:"count"`
	}
	if err := s.db.Model(&models.SensitiveWord{}).
		Select("category, COUNT(*) as count").
		Where("is_active = ?", true).
		Group("category").
		Find(&categoryStats).Error; err != nil {
		return nil, fmt.Errorf("获取分类统计失败: %w", err)
	}
	stats["categories"] = categoryStats
	
	// 按级别统计
	var levelStats []struct {
		Level string `json:"level"`
		Count int64  `json:"count"`
	}
	if err := s.db.Model(&models.SensitiveWord{}).
		Select("level, COUNT(*) as count").
		Where("is_active = ?", true).
		Group("level").
		Find(&levelStats).Error; err != nil {
		return nil, fmt.Errorf("获取级别统计失败: %w", err)
	}
	stats["levels"] = levelStats
	
	// 最近添加的敏感词（前10个）
	var recentWords []models.SensitiveWord
	if err := s.db.Where("is_active = ?", true).
		Order("created_at DESC").
		Limit(10).
		Find(&recentWords).Error; err != nil {
		return nil, fmt.Errorf("获取最近添加的敏感词失败: %w", err)
	}
	stats["recent_words"] = recentWords
	
	// 内存中加载的敏感词数
	stats["loaded_in_memory"] = len(s.sensitiveWords)
	
	stats["generated_at"] = time.Now()
	
	return stats, nil
}

// AddSensitiveWord 添加敏感词
func (s *ContentSecurityService) AddSensitiveWord(word, category, level string) error {
	sensitiveWord := &models.SensitiveWord{
		ID:        uuid.New().String(),
		Word:      strings.ToLower(word),
		Category:  category,
		Level:     models.ModerationLevel(level),
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	if err := s.db.Create(sensitiveWord).Error; err != nil {
		return fmt.Errorf("添加敏感词失败: %w", err)
	}
	
	// 更新内存中的敏感词库
	s.sensitiveWords[strings.ToLower(word)] = true
	
	return nil
}
