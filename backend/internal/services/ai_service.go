package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"openpenpal-backend/internal/config"
	"openpenpal-backend/internal/models"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AIService AI服务
type AIService struct {
	db              *gorm.DB
	config          *config.Config
	client          *http.Client
	usageService    *UserUsageService
	securityService *ContentSecurityService
	creditTaskSvc   *CreditTaskService // 积分任务服务
}

// NewAIService 创建AI服务实例
func NewAIService(db *gorm.DB, config *config.Config) *AIService {
	service := &AIService{
		db:     db,
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		usageService: NewUserUsageService(db, config),
	}

	// 延迟初始化安全服务（避免循环依赖）
	service.securityService = NewContentSecurityService(db, config, service)

	return service
}

// SetCreditTaskService 设置积分任务服务（避免循环依赖）
func (s *AIService) SetCreditTaskService(creditTaskSvc *CreditTaskService) {
	s.creditTaskSvc = creditTaskSvc
}

// GetActiveProvider 获取当前激活的AI提供商配置
func (s *AIService) GetActiveProvider() (*models.AIConfig, error) {
	var config models.AIConfig
	err := s.db.Where("is_active = ? AND used_quota < daily_quota", true).
		Order("priority DESC").
		First(&config).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 如果没有配置，创建默认配置
			return s.createDefaultConfig()
		}
		return nil, err
	}

	// 检查是否需要重置配额
	if time.Now().After(config.QuotaResetAt) {
		config.UsedQuota = 0
		config.QuotaResetAt = time.Now().Add(24 * time.Hour)
		s.db.Save(&config)
	}

	return &config, nil
}

// createDefaultConfig 创建默认AI配置
func (s *AIService) createDefaultConfig() (*models.AIConfig, error) {
	var config *models.AIConfig

	// 根据配置的Provider创建相应的默认配置
	switch s.config.AIProvider {
	case "moonshot":
		config = &models.AIConfig{
			ID:           uuid.New().String(),
			Provider:     models.ProviderMoonshot,
			APIKey:       s.config.MoonshotAPIKey,
			APIEndpoint:  "https://api.moonshot.cn/v1/chat/completions",
			Model:        "moonshot-v1-8k",
			Temperature:  0.9, // 提高温度以增加创造性和多样性
			MaxTokens:    1000,
			IsActive:     true,
			Priority:     100,
			DailyQuota:   10000,
			UsedQuota:    0,
			QuotaResetAt: time.Now().Add(24 * time.Hour),
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
	case "siliconflow":
		config = &models.AIConfig{
			ID:           uuid.New().String(),
			Provider:     models.ProviderSiliconFlow,
			APIKey:       s.config.SiliconFlowAPIKey,
			APIEndpoint:  "https://api.siliconflow.cn/v1/chat/completions",
			Model:        "Qwen/Qwen2.5-7B-Instruct",
			Temperature:  0.7,
			MaxTokens:    1000,
			IsActive:     true,
			Priority:     100,
			DailyQuota:   10000,
			UsedQuota:    0,
			QuotaResetAt: time.Now().Add(24 * time.Hour),
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
	case "claude":
		config = &models.AIConfig{
			ID:           uuid.New().String(),
			Provider:     models.ProviderClaude,
			APIKey:       s.config.ClaudeAPIKey,
			APIEndpoint:  "https://api.anthropic.com/v1/messages",
			Model:        "claude-3-sonnet-20240229",
			Temperature:  0.7,
			MaxTokens:    1000,
			IsActive:     true,
			Priority:     100,
			DailyQuota:   10000,
			UsedQuota:    0,
			QuotaResetAt: time.Now().Add(24 * time.Hour),
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
	default: // openai
		config = &models.AIConfig{
			ID:           uuid.New().String(),
			Provider:     models.ProviderOpenAI,
			APIKey:       s.config.OpenAIAPIKey,
			APIEndpoint:  "https://api.openai.com/v1/chat/completions",
			Model:        "gpt-3.5-turbo",
			Temperature:  0.7,
			MaxTokens:    1000,
			IsActive:     true,
			Priority:     100,
			DailyQuota:   10000,
			UsedQuota:    0,
			QuotaResetAt: time.Now().Add(24 * time.Hour),
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
	}

	if err := s.db.Create(config).Error; err != nil {
		return nil, err
	}

	return config, nil
}

// MatchPenPal AI匹配笔友
func (s *AIService) MatchPenPal(ctx context.Context, req *models.AIMatchRequest) (*models.AIMatchResponse, error) {
	// 获取信件信息
	var letter models.Letter
	if err := s.db.Preload("User").First(&letter, "id = ?", req.LetterID).Error; err != nil {
		return nil, fmt.Errorf("letter not found: %w", err)
	}

	// 检查用户每日使用量
	canUse, err := s.usageService.CanUsePenpalMatch(letter.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check usage limit: %w", err)
	}

	if !canUse {
		return nil, fmt.Errorf("daily penpal match limit exceeded (max %d per day)", DefaultUsageLimits.DailyMatches)
	}

	// 获取AI配置
	aiConfig, err := s.GetActiveProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to get AI provider: %w", err)
	}

	// 构建匹配提示词
	prompt := s.buildMatchPrompt(letter)

	// 调用AI API
	aiResponse, err := s.callAIAPI(ctx, aiConfig, prompt, models.TaskTypeMatch)
	if err != nil {
		return nil, fmt.Errorf("AI API call failed: %w", err)
	}

	// 解析AI响应并查找匹配用户
	matches, err := s.parseMatchResponse(aiResponse, letter.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	// 保存匹配记录
	for _, match := range matches.Matches {
		matchRecord := &models.AIMatch{
			ID:            uuid.New().String(),
			LetterID:      req.LetterID,
			MatchedUserID: match.UserID,
			MatchScore:    match.Score,
			MatchReason:   match.Reason,
			Status:        "pending",
			Provider:      aiConfig.Provider,
			CreatedAt:     time.Now(),
		}
		s.db.Create(matchRecord)
	}

	// 记录使用日志
	s.logAIUsage(letter.UserID, models.TaskTypeMatch, req.LetterID, aiConfig, 200, 300, "success", "")

	// 记录使用量
	if err := s.usageService.UsePenpalMatch(letter.UserID); err != nil {
		log.Printf("Failed to record penpal match usage for user %s: %v", letter.UserID, err)
	}

	// 触发AI互动积分奖励 - FSD规格
	if s.creditTaskSvc != nil {
		go func() {
			if err := s.creditTaskSvc.TriggerAIInteractionReward(letter.UserID, req.LetterID); err != nil {
				log.Printf("Failed to trigger AI interaction reward for user %s: %v", letter.UserID, err)
			}
		}()
	}

	return matches, nil
}

// ScheduleDelayedReply 安排延迟AI回信
func (s *AIService) ScheduleDelayedReply(ctx context.Context, req *models.AIReplyRequest) (string, error) {
	// 获取原信件
	var originalLetter models.Letter
	if err := s.db.Preload("User").First(&originalLetter, "id = ?", req.LetterID).Error; err != nil {
		return "", fmt.Errorf("letter not found: %w", err)
	}

	// 检查用户每日使用量
	canUse, err := s.usageService.CanUseAIReply(originalLetter.UserID)
	if err != nil {
		return "", fmt.Errorf("failed to check usage limit: %w", err)
	}

	if !canUse {
		return "", fmt.Errorf("daily AI reply limit exceeded (max %d per day)", DefaultUsageLimits.DailyAIReplies)
	}

	// 创建延迟队列服务
	delayQueueService, err := NewDelayQueueService(s.db, s.config)
	if err != nil {
		// 如果延迟队列不可用，降级到立即处理
		log.Printf("Delay queue unavailable, processing immediately: %v", err)
		reply, err := s.GenerateReply(ctx, req)
		if err != nil {
			return "", err
		}
		return reply.ID, nil
	}

	// 创建对话ID（用于追踪延迟任务）
	conversationID := uuid.New().String()

	// 安排延迟任务
	err = delayQueueService.ScheduleAIReply(
		originalLetter.UserID,
		string(req.Persona),
		originalLetter.Content,
		conversationID,
		req.DelayHours,
	)
	if err != nil {
		return "", fmt.Errorf("failed to schedule delayed reply: %w", err)
	}

	// 创建AI回信记录（状态为scheduled）
	aiReply := &models.AIReply{
		ID:               conversationID,
		OriginalLetterID: req.LetterID,
		ReplyLetterID:    "", // 稍后会更新
		Persona:          req.Persona,
		Provider:         models.ProviderDefault,
		DelayHours:       req.DelayHours,
		ScheduledAt:      time.Now().Add(time.Duration(req.DelayHours) * time.Hour),
		CreatedAt:        time.Now(),
	}
	s.db.Create(aiReply)

	// 记录使用量
	if err := s.usageService.UseAIReply(originalLetter.UserID); err != nil {
		log.Printf("Failed to record AI reply usage for user %s: %v", originalLetter.UserID, err)
	}

	// 触发AI互动积分奖励 - FSD规格
	if s.creditTaskSvc != nil {
		go func() {
			if err := s.creditTaskSvc.TriggerAIInteractionReward(originalLetter.UserID, conversationID); err != nil {
				log.Printf("Failed to trigger AI interaction reward for user %s: %v", originalLetter.UserID, err)
			}
		}()
	}

	return conversationID, nil
}

// GenerateReply 生成AI回信
func (s *AIService) GenerateReply(ctx context.Context, req *models.AIReplyRequest) (*models.Letter, error) {
	// 获取原信件
	var originalLetter models.Letter
	if err := s.db.Preload("User").First(&originalLetter, "id = ?", req.LetterID).Error; err != nil {
		return nil, fmt.Errorf("letter not found: %w", err)
	}

	// 对原信件内容进行安全检查
	if s.securityService != nil {
		securityResult, err := s.securityService.CheckContent(ctx, originalLetter.UserID, "letter", originalLetter.ID, originalLetter.Content)
		if err != nil {
			log.Printf("Security check failed for original letter: %v", err)
		} else if !securityResult.IsSafe && securityResult.RiskLevel == "critical" {
			return nil, fmt.Errorf("original letter contains unsafe content, reply generation blocked")
		}
	}

	// 获取AI配置
	aiConfig, err := s.GetActiveProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to get AI provider: %w", err)
	}

	// 构建回信提示词
	prompt := s.buildReplyPrompt(originalLetter, req.Persona)

	// 调用AI API
	aiResponse, err := s.callAIAPI(ctx, aiConfig, prompt, models.TaskTypeReply)
	if err != nil {
		return nil, fmt.Errorf("AI API call failed: %w", err)
	}

	// 创建回信
	replyContent := s.extractContentFromAIResponse(aiResponse)

	// 对生成的回信内容进行安全检查
	if s.securityService != nil {
		securityResult, err := s.securityService.CheckContent(ctx, "ai_"+string(req.Persona), "ai_reply", "", replyContent)
		if err != nil {
			log.Printf("Security check failed for AI reply: %v", err)
		} else if !securityResult.IsSafe {
			// 使用过滤后的内容
			replyContent = securityResult.FilteredContent
			log.Printf("AI reply content filtered due to security concerns")
		}
	}

	reply := &models.Letter{
		ID:        uuid.New().String(),
		UserID:    "ai_" + string(req.Persona), // AI用户ID
		Title:     fmt.Sprintf("Re: %s", originalLetter.Title),
		Content:   replyContent,
		Style:     originalLetter.Style,
		Status:    models.StatusGenerated,
		ReplyTo:   originalLetter.ID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 保存回信
	if err := s.db.Create(reply).Error; err != nil {
		return nil, fmt.Errorf("failed to save reply: %w", err)
	}

	// 创建AI回信记录
	aiReply := &models.AIReply{
		ID:               uuid.New().String(),
		OriginalLetterID: req.LetterID,
		ReplyLetterID:    reply.ID,
		Persona:          req.Persona,
		Provider:         aiConfig.Provider,
		DelayHours:       req.DelayHours,
		ScheduledAt:      time.Now().Add(time.Duration(req.DelayHours) * time.Hour),
		CreatedAt:        time.Now(),
	}
	s.db.Create(aiReply)

	// 记录使用日志
	s.logAIUsage(originalLetter.UserID, models.TaskTypeReply, reply.ID, aiConfig, 300, 400, "success", "")

	// 触发AI互动积分奖励 - FSD规格
	if s.creditTaskSvc != nil {
		go func() {
			if err := s.creditTaskSvc.TriggerAIInteractionReward(originalLetter.UserID, aiReply.ID); err != nil {
				log.Printf("Failed to trigger AI interaction reward for user %s: %v", originalLetter.UserID, err)
			}
		}()
	}

	return reply, nil
}

// GetInspirationWithLimit 获取写作灵感（带使用量限制和安全检查）
func (s *AIService) GetInspirationWithLimit(ctx context.Context, userID string, req *models.AIInspirationRequest) (*models.AIInspirationResponse, error) {
	// 检查用户每日使用量
	canUse, err := s.usageService.CanUseInspiration(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check usage limit: %w", err)
	}

	if !canUse {
		return nil, fmt.Errorf("daily inspiration limit exceeded (max %d per day)", DefaultUsageLimits.DailyInspirations)
	}

	// 对请求内容进行安全检查
	if req.Theme != "" {
		securityResult, err := s.securityService.CheckContent(ctx, userID, "inspiration_request", "", req.Theme)
		if err != nil {
			log.Printf("Security check failed for inspiration request: %v", err)
		} else if !securityResult.IsSafe {
			return nil, fmt.Errorf("content security check failed: %s", strings.Join(securityResult.ViolationType, ", "))
		}
	}

	// 调用原始方法获取灵感
	response, err := s.GetInspiration(ctx, req)
	if err != nil {
		return nil, err
	}

	// 对生成的灵感内容进行安全检查和过滤
	if s.securityService != nil {
		for i, inspiration := range response.Inspirations {
			securityResult, err := s.securityService.CheckContent(ctx, userID, "inspiration", inspiration.ID, inspiration.Prompt)
			if err != nil {
				log.Printf("Security check failed for generated inspiration: %v", err)
			} else if !securityResult.IsSafe {
				// 使用过滤后的内容
				response.Inspirations[i].Prompt = securityResult.FilteredContent
			}
		}
	}

	// 记录使用量
	if err := s.usageService.UseInspiration(userID); err != nil {
		log.Printf("Failed to record inspiration usage for user %s: %v", userID, err)
		// 不返回错误，因为灵感已经生成了
	}

	// 触发AI互动积分奖励 - FSD规格
	if s.creditTaskSvc != nil {
		go func() {
			// 使用灵感ID作为会话引用
			sessionID := ""
			if len(response.Inspirations) > 0 {
				sessionID = response.Inspirations[0].ID
			}
			if err := s.creditTaskSvc.TriggerAIInteractionReward(userID, sessionID); err != nil {
				log.Printf("Failed to trigger AI interaction reward for user %s: %v", userID, err)
			}
		}()
	}

	return response, nil
}

// GetInspiration 获取写作灵感
func (s *AIService) GetInspiration(ctx context.Context, req *models.AIInspirationRequest) (*models.AIInspirationResponse, error) {
	log.Printf("🎯 [GetInspiration] Starting inspiration generation...")

	// 在开发环境中，如果没有API密钥，使用本地生成的灵感
	if s.config.Environment == "development" && s.config.MoonshotAPIKey == "" {
		log.Printf("⚠️ [GetInspiration] No AI API key in development, using local inspirations")
		return s.generateLocalInspirations(req), nil
	}

	// 获取AI配置
	aiConfig, err := s.GetActiveProvider()
	if err != nil {
		log.Printf("❌ [GetInspiration] Failed to get AI provider: %v", err)
		// 在开发环境降级到本地生成
		if s.config.Environment == "development" {
			return s.generateLocalInspirations(req), nil
		}
		return nil, fmt.Errorf("failed to get AI provider: %w", err)
	}

	log.Printf("🔧 [GetInspiration] Using provider: %s, Model: %s", aiConfig.Provider, aiConfig.Model)

	// 构建灵感提示词
	prompt := s.buildInspirationPrompt(req)
	log.Printf("📝 [GetInspiration] Generated prompt: %d characters", len(prompt))

	// 调用AI API
	log.Printf("🚀 [GetInspiration] Calling AI API...")
	aiResponse, err := s.callAIAPI(ctx, aiConfig, prompt, models.TaskTypeInspiration)
	if err != nil {
		log.Printf("❌ [GetInspiration] AI API call failed: %v", err)
		// 在开发环境降级到本地生成
		if s.config.Environment == "development" {
			return s.generateLocalInspirations(req), nil
		}
		return nil, fmt.Errorf("AI API call failed: %w", err)
	}

	log.Printf("✅ [GetInspiration] AI API response received: %d characters", len(aiResponse))

	// 解析灵感响应
	inspirations, err := s.parseInspirationResponse(aiResponse)
	if err != nil {
		log.Printf("❌ [GetInspiration] Failed to parse AI response: %v", err)
		// 在开发环境降级到本地生成
		if s.config.Environment == "development" {
			return s.generateLocalInspirations(req), nil
		}
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	// 保存灵感记录
	log.Printf("💾 [GetInspiration] Saving %d inspirations to database...", len(inspirations.Inspirations))
	for i, insp := range inspirations.Inspirations {
		inspiration := &models.AIInspiration{
			ID:        uuid.New().String(),
			Theme:     insp.Theme,
			Prompt:    insp.Prompt,
			Style:     insp.Style,
			Tags:      fmt.Sprintf("[%s]", strings.Join(insp.Tags, ",")), // Convert slice to string
			Provider:  aiConfig.Provider,
			CreatedAt: time.Now(),
			IsActive:  true,
		}
		if err := s.db.Create(inspiration).Error; err != nil {
			log.Printf("⚠️ [GetInspiration] Failed to save inspiration %d: %v", i, err)
		}
		inspirations.Inspirations[i].ID = inspiration.ID
	}

	// 记录使用日志
	s.logAIUsage("system", models.TaskTypeInspiration, "", aiConfig, 100, 200, "success", "")

	log.Printf("✅ [GetInspiration] Successfully generated %d inspirations", len(inspirations.Inspirations))

	return inspirations, nil
}

// GenerateReplyAdvice 生成回信角度建议
func (s *AIService) GenerateReplyAdvice(ctx context.Context, req *models.AIReplyAdviceRequest) (*models.AIReplyAdvice, error) {
	// 获取原信件
	var originalLetter models.Letter
	if err := s.db.Preload("User").First(&originalLetter, "id = ?", req.LetterID).Error; err != nil {
		return nil, fmt.Errorf("letter not found: %w", err)
	}

	// 获取AI配置
	aiConfig, err := s.GetActiveProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to get AI provider: %w", err)
	}

	// 构建回信建议提示词
	prompt := s.buildReplyAdvicePrompt(originalLetter, req)

	// 调用AI API
	aiResponse, err := s.callAIAPI(ctx, aiConfig, prompt, models.TaskTypeReply)
	if err != nil {
		return nil, fmt.Errorf("AI API call failed: %w", err)
	}

	// 解析AI响应
	advice, err := s.parseReplyAdviceResponse(aiResponse, originalLetter, req)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	// 设置调度时间
	if req.DeliveryDays > 0 {
		scheduledTime := time.Now().Add(time.Duration(req.DeliveryDays) * 24 * time.Hour)
		advice.ScheduledFor = &scheduledTime
	}

	// 保存回信建议记录
	if err := s.db.Create(advice).Error; err != nil {
		return nil, fmt.Errorf("failed to save reply advice: %w", err)
	}

	// 记录使用日志
	s.logAIUsage(originalLetter.UserID, models.TaskTypeReply, advice.ID, aiConfig, 300, 500, "success", "")

	return advice, nil
}

// CurateLetters AI策展信件
func (s *AIService) CurateLetters(ctx context.Context, req *models.AICurateRequest) error {
	// 获取AI配置
	aiConfig, err := s.GetActiveProvider()
	if err != nil {
		return fmt.Errorf("failed to get AI provider: %w", err)
	}

	// 获取信件内容
	var letters []models.Letter
	if err := s.db.Where("id IN ?", req.LetterIDs).Find(&letters).Error; err != nil {
		return fmt.Errorf("failed to get letters: %w", err)
	}

	// 逐个处理信件
	for _, letter := range letters {
		// 构建策展提示词
		prompt := s.buildCuratePrompt(letter)

		// 调用AI API
		aiResponse, err := s.callAIAPI(ctx, aiConfig, prompt, models.TaskTypeCurate)
		if err != nil {
			continue // 跳过失败的
		}

		// 解析策展信息
		curation, err := s.parseCurationResponse(aiResponse, letter.ID)
		if err != nil {
			continue
		}

		// 保存策展记录
		if req.ExhibitionID != "" {
			curation.ExhibitionID = &req.ExhibitionID
		}

		if req.AutoApprove {
			now := time.Now()
			curation.ApprovedAt = &now
		}

		s.db.Create(curation)
	}

	return nil
}

// 辅助方法

// callAIAPI 调用AI API
func (s *AIService) callAIAPI(ctx context.Context, config *models.AIConfig, prompt string, taskType models.AITaskType) (string, error) {
	// 根据不同的Provider调用不同的API
	switch config.Provider {
	case models.ProviderOpenAI:
		return s.callOpenAI(ctx, config, prompt)
	case models.ProviderClaude:
		return s.callClaude(ctx, config, prompt)
	case models.ProviderSiliconFlow:
		return s.callSiliconFlow(ctx, config, prompt)
	case models.ProviderMoonshot:
		return s.callMoonshot(ctx, config, prompt)
	default:
		return s.callMoonshot(ctx, config, prompt) // 默认使用Moonshot
	}
}

// callOpenAI 调用OpenAI API
func (s *AIService) callOpenAI(ctx context.Context, config *models.AIConfig, prompt string) (string, error) {
	// 构建请求体
	requestBody := map[string]interface{}{
		"model": config.Model,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "你是OpenPenPal的AI助手，帮助用户进行笔友匹配、生成回信、提供写作灵感和策展信件。请用温暖、友好的语气回应。",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": config.Temperature,
		"max_tokens":  config.MaxTokens,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "POST", config.APIEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.APIKey))

	// 发送请求
	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("AI API error: %s", string(body))
	}

	// 解析响应
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	// 提取内容
	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", errors.New("invalid AI response format")
	}

	choice := choices[0].(map[string]interface{})
	message := choice["message"].(map[string]interface{})
	content := message["content"].(string)

	// 更新配额使用
	config.UsedQuota += int(result["usage"].(map[string]interface{})["total_tokens"].(float64))
	s.db.Save(config)

	return content, nil
}

// callClaude 调用Claude API (示例实现)
func (s *AIService) callClaude(ctx context.Context, config *models.AIConfig, prompt string) (string, error) {
	// Claude API实现类似，这里简化处理
	return s.callOpenAI(ctx, config, prompt)
}

// callSiliconFlow 调用SiliconFlow API
func (s *AIService) callSiliconFlow(ctx context.Context, config *models.AIConfig, prompt string) (string, error) {
	log.Printf("🤖 [SiliconFlow] Starting API call...")
	log.Printf("🤖 [SiliconFlow] API Endpoint: %s", config.APIEndpoint)
	log.Printf("🤖 [SiliconFlow] Model: %s", config.Model)

	// 构建请求体 - SiliconFlow API与OpenAI兼容
	requestBody := map[string]interface{}{
		"model": config.Model,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "你是OpenPenPal的AI助手，帮助用户进行笔友匹配、生成回信、提供写作灵感和策展信件。请用温暖、友好的语气回应。",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": config.Temperature,
		"max_tokens":  config.MaxTokens,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		log.Printf("❌ [SiliconFlow] Failed to marshal request body: %v", err)
		return "", err
	}

	log.Printf("🤖 [SiliconFlow] Request body size: %d bytes", len(jsonData))

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "POST", config.APIEndpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("❌ [SiliconFlow] Failed to create request: %v", err)
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.APIKey))

	// 安全日志：不记录任何API密钥信息
	if len(config.APIKey) > 0 {
		log.Printf("🔑 [SiliconFlow] API Key configured")
	} else {
		log.Printf("⚠️ [SiliconFlow] No API Key configured")
	}

	// 发送请求
	log.Printf("🚀 [SiliconFlow] Sending request to %s", config.APIEndpoint)
	resp, err := s.client.Do(req)
	if err != nil {
		log.Printf("❌ [SiliconFlow] Request failed: %v", err)
		return "", fmt.Errorf("SiliconFlow API request failed: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("❌ [SiliconFlow] Failed to read response body: %v", err)
		return "", err
	}

	log.Printf("📥 [SiliconFlow] Response status: %d", resp.StatusCode)
	log.Printf("📥 [SiliconFlow] Response body size: %d bytes", len(body))

	if resp.StatusCode != http.StatusOK {
		log.Printf("❌ [SiliconFlow] API error response: %s", string(body))
		return "", fmt.Errorf("SiliconFlow AI API error (status %d): %s", resp.StatusCode, string(body))
	}

	// 解析响应 - SiliconFlow API返回格式与OpenAI兼容
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("❌ [SiliconFlow] Failed to parse response JSON: %v", err)
		log.Printf("❌ [SiliconFlow] Raw response: %s", string(body))
		return "", fmt.Errorf("failed to parse SiliconFlow response: %w", err)
	}

	// 提取内容
	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		log.Printf("❌ [SiliconFlow] Invalid response format - no choices found")
		log.Printf("❌ [SiliconFlow] Response structure: %+v", result)
		return "", errors.New("invalid SiliconFlow AI response format: no choices found")
	}

	choice, ok := choices[0].(map[string]interface{})
	if !ok {
		log.Printf("❌ [SiliconFlow] Invalid choice format")
		return "", errors.New("invalid SiliconFlow AI response format: invalid choice format")
	}

	message, ok := choice["message"].(map[string]interface{})
	if !ok {
		log.Printf("❌ [SiliconFlow] Invalid message format")
		return "", errors.New("invalid SiliconFlow AI response format: invalid message format")
	}

	content, ok := message["content"].(string)
	if !ok {
		log.Printf("❌ [SiliconFlow] Invalid content format")
		return "", errors.New("invalid SiliconFlow AI response format: invalid content format")
	}

	log.Printf("✅ [SiliconFlow] Successfully extracted content: %d characters", len(content))

	// 更新配额使用
	if usage, ok := result["usage"].(map[string]interface{}); ok {
		if totalTokens, ok := usage["total_tokens"].(float64); ok {
			config.UsedQuota += int(totalTokens)
			s.db.Save(config)
			log.Printf("📊 [SiliconFlow] Token usage: %.0f tokens", totalTokens)
		}
	}

	return content, nil
}

// callMoonshot 调用Moonshot API
func (s *AIService) callMoonshot(ctx context.Context, config *models.AIConfig, prompt string) (string, error) {
	// Use the fixed implementation
	return s.callMoonshotFixed(ctx, config, prompt)
}

// generateLocalInspirations 生成本地灵感（开发环境备用方案）
func (s *AIService) generateLocalInspirations(req *models.AIInspirationRequest) *models.AIInspirationResponse {
	log.Printf("🎯 [generateLocalInspirations] Generating local inspirations...")

	// 预定义的主题灵感库
	inspirationPool := []struct {
		Theme  string
		Prompt string
		Style  string
		Tags   []string
	}{
		{
			Theme:  "日常感悟",
			Prompt: "写一写今天早晨醒来时的第一个念头，以及它如何影响了你一天的心情。可以是对新一天的期待，也可以是昨夜梦境的延续。",
			Style:  "温暖治愈",
			Tags:   []string{"日常", "感悟", "晨光"},
		},
		{
			Theme:  "友情岁月",
			Prompt: "回忆一个和朋友共同经历的特别时刻，可能是一次深夜谈心，一场说走就走的旅行，或是默默陪伴的温暖。",
			Style:  "深情怀念",
			Tags:   []string{"友情", "回忆", "陪伴"},
		},
		{
			Theme:  "成长印记",
			Prompt: "写下最近让你感到自己成长的一个瞬间，可能是勇敢说出了内心的想法，或是学会了放下某些执念。",
			Style:  "励志向上",
			Tags:   []string{"成长", "勇气", "改变"},
		},
		{
			Theme:  "城市漫步",
			Prompt: "描述你所在城市的一个角落，可能是清晨无人的街道，黄昏时分的公园，或是深夜还在营业的小店。",
			Style:  "诗意浪漫",
			Tags:   []string{"城市", "风景", "生活"},
		},
		{
			Theme:  "美食记忆",
			Prompt: "写一道让你印象深刻的食物，不只是味道，更是与它相关的人和故事。可能是妈妈的拿手菜，或是旅途中的意外发现。",
			Style:  "温情细腻",
			Tags:   []string{"美食", "记忆", "温暖"},
		},
		{
			Theme:  "季节转换",
			Prompt: "感受当下季节的变化，写下窗外的风景、空气的味道，以及季节更替带给你的心情变化。",
			Style:  "自然清新",
			Tags:   []string{"季节", "自然", "变化"},
		},
		{
			Theme:  "夜晚思绪",
			Prompt: "在安静的夜晚，写下此刻涌上心头的想法。可能是对过往的释怀，对未来的憧憬，或是对当下的珍惜。",
			Style:  "深沉内省",
			Tags:   []string{"夜晚", "思考", "内心"},
		},
		{
			Theme:  "童年回响",
			Prompt: "回想一个童年的场景，可能是夏天的蝉鸣，雨后的泥土味，或是放学路上的小发现。",
			Style:  "怀旧温馨",
			Tags:   []string{"童年", "回忆", "纯真"},
		},
		{
			Theme:  "爱的表达",
			Prompt: "写一写生活中那些不起眼的爱的表达，可能是一杯递到手中的热茶，一个不经意的拥抱，或是默默的守护。",
			Style:  "温柔深情",
			Tags:   []string{"爱", "细节", "感动"},
		},
		{
			Theme:  "独处时光",
			Prompt: "描述你享受独处的一个时刻，在这个时刻里你做了什么，想了什么，感受到了什么。",
			Style:  "宁静致远",
			Tags:   []string{"独处", "自我", "平静"},
		},
		{
			Theme:  "小确幸",
			Prompt: "记录今天遇到的一个小确幸，可能是意外听到的一首好歌，收到的一条温暖消息，或是看到的一个微笑。",
			Style:  "轻松愉快",
			Tags:   []string{"幸福", "日常", "美好"},
		},
		{
			Theme:  "旅途见闻",
			Prompt: "写下一次旅行中最难忘的片段，不一定是著名景点，可能是路边的风景，陌生人的善意，或是内心的触动。",
			Style:  "游记散文",
			Tags:   []string{"旅行", "见闻", "感悟"},
		},
	}

	// 根据请求参数筛选灵感
	var selectedInspirations []struct {
		Theme  string
		Prompt string
		Style  string
		Tags   []string
	}

	// 如果指定了主题，优先选择匹配的主题
	if req.Theme != "" {
		for _, insp := range inspirationPool {
			if strings.Contains(insp.Theme, req.Theme) || strings.Contains(insp.Prompt, req.Theme) {
				selectedInspirations = append(selectedInspirations, insp)
			}
		}
	}

	// 如果没有找到匹配的，或者没有指定主题，随机选择
	if len(selectedInspirations) == 0 {
		// 使用时间戳作为随机种子，确保每次调用有不同结果
		timestamp := time.Now().UnixNano()
		startIdx := int(timestamp % int64(len(inspirationPool)))

		count := req.Count
		if count == 0 || count > 5 {
			count = 1
		}

		for i := 0; i < count && i < len(inspirationPool); i++ {
			idx := (startIdx + i) % len(inspirationPool)
			selectedInspirations = append(selectedInspirations, inspirationPool[idx])
		}
	}

	// 构建响应
	response := &models.AIInspirationResponse{
		Inspirations: make([]struct {
			ID     string   `json:"id"`
			Theme  string   `json:"theme"`
			Prompt string   `json:"prompt"`
			Style  string   `json:"style"`
			Tags   []string `json:"tags"`
		}, 0),
	}

	// 限制返回数量
	maxCount := req.Count
	if maxCount == 0 || maxCount > 5 {
		maxCount = 1
	}

	for i := 0; i < len(selectedInspirations) && i < maxCount; i++ {
		insp := selectedInspirations[i]
		response.Inspirations = append(response.Inspirations, struct {
			ID     string   `json:"id"`
			Theme  string   `json:"theme"`
			Prompt string   `json:"prompt"`
			Style  string   `json:"style"`
			Tags   []string `json:"tags"`
		}{
			ID:     uuid.New().String(),
			Theme:  insp.Theme,
			Prompt: insp.Prompt,
			Style:  insp.Style,
			Tags:   insp.Tags,
		})

		// 保存到数据库
		inspiration := &models.AIInspiration{
			ID:        response.Inspirations[i].ID,
			Theme:     insp.Theme,
			Prompt:    insp.Prompt,
			Style:     insp.Style,
			Tags:      fmt.Sprintf("[%s]", strings.Join(insp.Tags, ",")),
			Provider:  models.ProviderLocal,
			CreatedAt: time.Now(),
			IsActive:  true,
		}
		if err := s.db.Create(inspiration).Error; err != nil {
			log.Printf("⚠️ [generateLocalInspirations] Failed to save inspiration: %v", err)
		}
	}

	log.Printf("✅ [generateLocalInspirations] Generated %d local inspirations", len(response.Inspirations))
	return response
}

// buildMatchPrompt 构建匹配提示词
func (s *AIService) buildMatchPrompt(letter models.Letter) string {
	return fmt.Sprintf(`
基于以下信件内容，为写信人推荐3个最合适的笔友：

信件标题：%s
信件内容：%s
写信人：%s

请分析信件的情感、主题和风格，推荐具有相似兴趣或互补特质的用户。
返回JSON格式：
{
  "matches": [
    {
      "match_type": "similar_interest|complementary",
      "score": 0.95,
      "reason": "推荐理由",
      "tags": ["共同标签1", "共同标签2"]
    }
  ]
}
`, letter.Title, letter.Content, letter.User.Username)
}

// buildReplyPrompt 构建回信提示词
func (s *AIService) buildReplyPrompt(letter models.Letter, persona models.AIPersona) string {
	personaDesc := s.getPersonaDescription(persona)
	return fmt.Sprintf(`
你是一个%s，请以这个身份给下面的信件写一封回信：

原信标题：%s
原信内容：%s

请保持%s的语气和风格，写一封温暖、真诚的回信。
回信应该：
1. 对原信内容有所回应
2. 分享你的见解或经历
3. 提出一些问题促进对话
4. 字数在200-500字之间

请直接返回回信内容，不要包含其他说明。
`, personaDesc, letter.Title, letter.Content, personaDesc)
}

// buildInspirationPrompt 构建灵感提示词
func (s *AIService) buildInspirationPrompt(req *models.AIInspirationRequest) string {
	theme := req.Theme
	if theme == "" {
		theme = "随机主题"
	}

	style := req.Style
	if style == "" {
		style = "温暖友好"
	}

	count := req.Count
	if count == 0 {
		count = 1
	}

	// 添加随机性和时间元素以确保每次生成不同的内容
	timestamp := time.Now().Unix()
	randomSeed := uuid.New().String()[:8]

	return fmt.Sprintf(`
请生成%d个独特的写信灵感提示（时间戳：%d，种子：%s）：

主题：%s
风格：%s
标签：%s

要求：
1. 每个灵感必须独一无二，避免重复
2. 提供具体的写作切入点
3. 激发情感共鸣
4. 适合手写信的形式
5. 50-100字的描述
6. 创意新颖，避免老套

请确保每次生成的内容都不同，充分发挥创造力。

返回JSON格式：
{
  "inspirations": [
    {
      "theme": "主题",
      "prompt": "写作提示",
      "style": "风格",
      "tags": ["标签1", "标签2"]
    }
  ]
}
`, count, timestamp, randomSeed, theme, style, strings.Join(req.Tags, ", "))
}

// buildReplyAdvicePrompt 构建回信建议提示词
func (s *AIService) buildReplyAdvicePrompt(letter models.Letter, req *models.AIReplyAdviceRequest) string {

	return fmt.Sprintf(`
你是OpenPenPal的AI助手，现在需要以"%s"的身份为以下信件提供回信角度建议：

人设信息：
- 身份：%s
- 关系：%s
- 详细描述：%s

原信信息：
- 标题：%s
- 内容：%s
- 作者：%s

请提供温暖而有深度的回信角度建议，包括：
1. 3-5个不同的回信角度/观点
2. 情感基调（温暖、关怀、理解等）
3. 建议的话题方向
4. 适合的写作风格
5. 回信的关键要点

要求：
- 体现人设的特点和情感温度
- 考虑关系的特殊性（如已故亲人的怀念、老友的思念等）
- 提供具体而可操作的建议
- 保持真诚和温暖的语调

返回JSON格式：
{
  "perspectives": ["角度1", "角度2", "角度3"],
  "emotional_tone": "情感基调",
  "suggested_topics": "建议话题",
  "writing_style": "写作风格",
  "key_points": "关键要点"
}
`, req.PersonaName, req.PersonaName, req.Relationship, req.PersonaDesc,
		letter.Title, letter.Content, letter.User.Username)
}

// buildCuratePrompt 构建策展提示词
func (s *AIService) buildCuratePrompt(letter models.Letter) string {
	return fmt.Sprintf(`
请为以下信件进行博物馆策展分析：

信件标题：%s
信件内容：%s

请分析：
1. 适合的展览类别
2. 关键标签（3-5个）
3. 一句话总结（20字以内）
4. 精彩片段（1-3个，每个不超过50字）

返回JSON格式：
{
  "category": "类别",
  "tags": ["标签1", "标签2"],
  "summary": "总结",
  "highlights": ["片段1", "片段2"],
  "score": 0.85
}
`, letter.Title, letter.Content)
}

// 解析辅助方法

// parseMatchResponse 解析匹配响应
func (s *AIService) parseMatchResponse(aiResponse string, excludeUserID string) (*models.AIMatchResponse, error) {
	// 先尝试解析AI返回的JSON
	var aiResult struct {
		Matches []struct {
			MatchType string   `json:"match_type"`
			Score     float64  `json:"score"`
			Reason    string   `json:"reason"`
			Tags      []string `json:"tags"`
		} `json:"matches"`
	}

	if err := json.Unmarshal([]byte(aiResponse), &aiResult); err != nil {
		// 如果解析失败，创建默认响应
		return &models.AIMatchResponse{
			Matches: []struct {
				UserID     string   `json:"user_id"`
				Username   string   `json:"username"`
				Score      float64  `json:"score"`
				Reason     string   `json:"reason"`
				CommonTags []string `json:"common_tags"`
			}{},
		}, nil
	}

	// 基于AI推荐查找实际用户
	response := &models.AIMatchResponse{}

	for _, match := range aiResult.Matches {
		// 查找符合条件的用户（这里简化处理，实际应该根据标签和类型查找）
		var users []models.User
		query := s.db.Where("id != ?", excludeUserID).Limit(1)

		// 可以根据match_type和tags进一步筛选
		if err := query.Find(&users).Error; err != nil {
			continue
		}

		for _, user := range users {
			response.Matches = append(response.Matches, struct {
				UserID     string   `json:"user_id"`
				Username   string   `json:"username"`
				Score      float64  `json:"score"`
				Reason     string   `json:"reason"`
				CommonTags []string `json:"common_tags"`
			}{
				UserID:     user.ID,
				Username:   user.Username,
				Score:      match.Score,
				Reason:     match.Reason,
				CommonTags: match.Tags,
			})
		}
	}

	return response, nil
}

// parseInspirationResponse 解析灵感响应
func (s *AIService) parseInspirationResponse(aiResponse string) (*models.AIInspirationResponse, error) {
	log.Printf("🎨 [AI Inspiration] Parsing response of length: %d", len(aiResponse))

	// 处理可能包含Markdown代码块的响应
	cleanResponse := aiResponse
	if strings.Contains(aiResponse, "```json") {
		// 提取JSON内容，处理可能的换行符
		start := strings.Index(aiResponse, "```json")
		if start != -1 {
			// 跳过 ```json 和可能的换行符
			start += 7
			// 找到 ```json 后的第一个 { 或 [
			for start < len(aiResponse) && (aiResponse[start] == '\n' || aiResponse[start] == '\r' || aiResponse[start] == ' ') {
				start++
			}

			end := strings.LastIndex(aiResponse, "```")
			if end > start {
				cleanResponse = strings.TrimSpace(aiResponse[start:end])
				log.Printf("🧹 [AI Inspiration] Cleaned markdown wrapper, new length: %d", len(cleanResponse))
			}
		}
	} else if strings.Contains(aiResponse, "```") {
		// 处理只有```的情况
		start := strings.Index(aiResponse, "```") + 3
		// 跳过可能的换行符
		for start < len(aiResponse) && (aiResponse[start] == '\n' || aiResponse[start] == '\r' || aiResponse[start] == ' ') {
			start++
		}

		end := strings.LastIndex(aiResponse, "```")
		if end > start {
			cleanResponse = strings.TrimSpace(aiResponse[start:end])
		}
	}

	var response models.AIInspirationResponse
	if err := json.Unmarshal([]byte(cleanResponse), &response); err != nil {
		log.Printf("❌ [AI Inspiration] Failed to parse JSON response: %v", err)
		log.Printf("❌ [AI Inspiration] Raw AI response: %s", aiResponse)

		// 如果解析失败，返回默认灵感
		log.Printf("⚠️ [AI Inspiration] Using fallback inspiration due to parse error")
		return &models.AIInspirationResponse{
			Inspirations: []struct {
				ID     string   `json:"id"`
				Theme  string   `json:"theme"`
				Prompt string   `json:"prompt"`
				Style  string   `json:"style"`
				Tags   []string `json:"tags"`
			}{
				{
					Theme:  "日常感悟",
					Prompt: "写一写今天遇到的一件小事，以及它给你带来的感受",
					Style:  "温暖",
					Tags:   []string{"日常", "感悟"},
				},
			},
		}, nil
	}

	log.Printf("✅ [AI Inspiration] Successfully parsed %d inspirations", len(response.Inspirations))
	return &response, nil
}

// parseReplyAdviceResponse 解析回信建议响应
func (s *AIService) parseReplyAdviceResponse(aiResponse string, letter models.Letter, req *models.AIReplyAdviceRequest) (*models.AIReplyAdvice, error) {
	var result struct {
		Perspectives    []string `json:"perspectives"`
		EmotionalTone   string   `json:"emotional_tone"`
		SuggestedTopics string   `json:"suggested_topics"`
		WritingStyle    string   `json:"writing_style"`
		KeyPoints       string   `json:"key_points"`
	}

	if err := json.Unmarshal([]byte(aiResponse), &result); err != nil {
		// 提供默认建议
		result.Perspectives = []string{
			"从回忆与情感的角度回应",
			"从经验分享的角度给予建议",
			"从理解与共鸣的角度表达支持",
		}
		result.EmotionalTone = "温暖关怀"
		result.SuggestedTopics = "情感共鸣、人生感悟、美好回忆"
		result.WritingStyle = "温暖亲切、真诚朴实"
		result.KeyPoints = "表达理解和关怀，分享相关经历，给予温暖的建议"
	}

	// 创建回信建议记录
	advice := &models.AIReplyAdvice{
		ID:              uuid.New().String(),
		LetterID:        req.LetterID,
		UserID:          letter.UserID,
		PersonaType:     req.PersonaType,
		PersonaName:     req.PersonaName,
		PersonaDesc:     req.PersonaDesc,
		Perspectives:    fmt.Sprintf("[\"%s\"]", strings.Join(result.Perspectives, "\", \"")),
		EmotionalTone:   result.EmotionalTone,
		SuggestedTopics: result.SuggestedTopics,
		WritingStyle:    result.WritingStyle,
		KeyPoints:       result.KeyPoints,
		DeliveryDelay:   req.DeliveryDays,
		Provider:        models.ProviderSiliconFlow, // 使用当前配置的Provider
		CreatedAt:       time.Now(),
	}

	return advice, nil
}

// parseCurationResponse 解析策展响应
func (s *AIService) parseCurationResponse(aiResponse string, letterID string) (*models.AICuration, error) {
	var result struct {
		Category   string   `json:"category"`
		Tags       []string `json:"tags"`
		Summary    string   `json:"summary"`
		Highlights []string `json:"highlights"`
		Score      float64  `json:"score"`
	}

	if err := json.Unmarshal([]byte(aiResponse), &result); err != nil {
		// 默认策展信息
		result.Category = "其他"
		result.Tags = []string{"待分类"}
		result.Summary = "一封有趣的信件"
		result.Highlights = []string{}
		result.Score = 0.5
	}

	return &models.AICuration{
		ID:         uuid.New().String(),
		LetterID:   letterID,
		Category:   result.Category,
		Tags:       fmt.Sprintf("[%s]", strings.Join(result.Tags, ",")), // Convert slice to string
		Summary:    result.Summary,
		Highlights: fmt.Sprintf("[%s]", strings.Join(result.Highlights, ",")), // Convert slice to string
		Score:      result.Score,
		Provider:   models.ProviderOpenAI,
		CreatedAt:  time.Now(),
	}, nil
}

// extractContentFromAIResponse 从AI响应中提取内容
func (s *AIService) extractContentFromAIResponse(aiResponse string) string {
	// 直接返回AI生成的内容
	return strings.TrimSpace(aiResponse)
}

// getPersonaContext 获取人设上下文信息
func (s *AIService) getPersonaContext(personaType, personaName, personaDesc, relationship string) string {
	var context strings.Builder

	context.WriteString(fmt.Sprintf("身份：%s\n", personaName))

	if relationship != "" {
		context.WriteString(fmt.Sprintf("关系：%s\n", relationship))
	}

	if personaDesc != "" {
		context.WriteString(fmt.Sprintf("人设描述：%s\n", personaDesc))
	}

	// 根据人设类型添加特定的上下文
	switch personaType {
	case "deceased":
		context.WriteString("特别提醒：作为已故的亲人，回信应该带有深深的思念和爱意，体现对在世亲人的关怀和指引。\n")
	case "distant_friend":
		context.WriteString("特别提醒：作为多年未见的好友，回信应该表达久别重逢的喜悦和对友谊的珍惜。\n")
	case "unspoken_love":
		context.WriteString("特别提醒：作为未曾表白的爱人，回信应该含蓄而深情，表达内心的情感但保持一定的距离感。\n")
	case "custom":
		context.WriteString("特别提醒：根据自定义的人设特点，回信应该充分体现角色的独特性格和背景。\n")
	}

	return context.String()
}

// getPersonaDescription 获取人设描述
func (s *AIService) getPersonaDescription(persona models.AIPersona) string {
	descriptions := map[models.AIPersona]string{
		models.PersonaPoet:        "诗人",
		models.PersonaPhilosopher: "哲学家",
		models.PersonaArtist:      "艺术家",
		models.PersonaScientist:   "科学家",
		models.PersonaTraveler:    "旅行者",
		models.PersonaHistorian:   "历史学家",
		models.PersonaMentor:      "人生导师",
		models.PersonaFriend:      "知心朋友",
	}

	if desc, ok := descriptions[persona]; ok {
		return desc
	}
	return "朋友"
}

// logAIUsage 记录AI使用日志
func (s *AIService) logAIUsage(userID string, taskType models.AITaskType, taskID string,
	config *models.AIConfig, inputTokens, outputTokens int, status, errorMsg string) {

	log := &models.AIUsageLog{
		ID:           uuid.New().String(),
		UserID:       userID,
		TaskType:     taskType,
		TaskID:       taskID,
		Provider:     config.Provider,
		Model:        config.Model,
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
		TotalTokens:  inputTokens + outputTokens,
		Status:       status,
		ErrorMessage: errorMsg,
		CreatedAt:    time.Now(),
	}

	s.db.Create(log)
}

// GetAIUsageStats 获取用户AI使用统计
func (s *AIService) GetAIUsageStats(userID string) (map[string]interface{}, error) {
	// 如果是匿名用户
	if userID == "" || userID == "anonymous" {
		return map[string]interface{}{
			"user_id": "anonymous",
			"usage": map[string]int{
				"matches_created":   0,
				"replies_generated": 0,
				"inspirations_used": 0,
				"letters_curated":   0,
			},
			"limits": map[string]int{
				"daily_matches":      3,
				"daily_replies":      2,
				"daily_inspirations": 5,
				"daily_curations":    1,
			},
			"remaining": map[string]int{
				"matches":      3,
				"replies":      2,
				"inspirations": 5,
				"curations":    1,
			},
		}, nil
	}

	// TODO: 获取实际的用户使用统计
	// 临时返回模拟数据
	return map[string]interface{}{
		"user_id": userID,
		"usage": map[string]int{
			"matches_created":   5,
			"replies_generated": 3,
			"inspirations_used": 10,
			"letters_curated":   2,
		},
		"limits": map[string]int{
			"daily_matches":      10,
			"daily_replies":      5,
			"daily_inspirations": 20,
			"daily_curations":    5,
		},
		"remaining": map[string]int{
			"matches":      5,
			"replies":      2,
			"inspirations": 10,
			"curations":    3,
		},
	}, nil
}

// EnhanceContent 增强文本内容 - 专门为CloudLetter等场景设计
func (s *AIService) EnhanceContent(ctx context.Context, content string, persona *CloudPersona, emotionalTone string) (string, error) {
	log.Printf("🤖 [AIService] Starting content enhancement")

	// 获取AI配置
	aiConfig, err := s.GetActiveProvider()
	if err != nil {
		return "", fmt.Errorf("failed to get AI provider: %w", err)
	}

	// 构建增强提示词
	prompt := s.buildContentEnhancementPrompt(content, persona, emotionalTone)

	// 调用AI API
	enhancedContent, err := s.callAIAPI(ctx, aiConfig, prompt, models.TaskTypeCurate)
	if err != nil {
		return "", fmt.Errorf("AI API call failed: %w", err)
	}

	// 记录使用日志
	if persona != nil {
		s.logAIUsage(persona.UserID, models.TaskTypeCurate, "", aiConfig, 200, 300, "success", "")
	}

	log.Printf("✅ [AIService] Content enhancement completed")
	return enhancedContent, nil
}

// buildContentEnhancementPrompt 构建内容增强提示词
func (s *AIService) buildContentEnhancementPrompt(content string, persona *CloudPersona, emotionalTone string) string {
	var prompt strings.Builder

	prompt.WriteString("作为专业的情感文字编辑，请帮助改善这封信件的表达方式。\n\n")

	// 如果有人物角色信息，添加上下文
	if persona != nil {
		prompt.WriteString("收信人信息：\n")
		prompt.WriteString(fmt.Sprintf("- 姓名：%s\n", persona.Name))
		prompt.WriteString(fmt.Sprintf("- 关系：%s\n", persona.Relationship))
		
		if persona.Description != "" {
			prompt.WriteString(fmt.Sprintf("- 描述：%s\n", persona.Description))
		}
		
		if persona.Personality != "" {
			prompt.WriteString(fmt.Sprintf("- 性格：%s\n", persona.Personality))
		}
		
		if persona.Memories != "" {
			prompt.WriteString(fmt.Sprintf("- 共同回忆：%s\n", persona.Memories))
		}
		prompt.WriteString("\n")
	}

	// 情感色调指导
	if emotionalTone != "" {
		prompt.WriteString(fmt.Sprintf("期望情感色调：%s\n\n", emotionalTone))
	}

	// 原始内容
	prompt.WriteString(fmt.Sprintf("原始信件内容：\n%s\n\n", content))

	// 增强指导原则
	prompt.WriteString("请根据以下要求改善信件：\n")
	prompt.WriteString("1. 保持原作者的真实情感和意图，不改变核心表达\n")
	prompt.WriteString("2. 优化语言表达，增加文字的美感和感染力\n")
	prompt.WriteString("3. 根据收信人关系调整语调和措辞的亲密程度\n")
	prompt.WriteString("4. 增强情感深度，让表达更加真挚动人\n")
	prompt.WriteString("5. 确保内容适合这种特殊关系，避免不合适的表达\n")
	prompt.WriteString("6. 保持中文表达习惯，语句通顺自然\n\n")

	// 根据关系类型添加特定指导
	if persona != nil {
		switch persona.Relationship {
		case RelationshipDeceased:
			prompt.WriteString("7. 体现对已故亲人的深切思念和爱意\n")
			prompt.WriteString("8. 表达感恩之情和美好回忆\n")
		case RelationshipDistantFriend:
			prompt.WriteString("7. 表达久别重逢的喜悦和友谊的珍贵\n")
			prompt.WriteString("8. 适当回忆共同的美好时光\n")
		case RelationshipUnspokenLove:
			prompt.WriteString("7. 保持含蓄而深情的表达方式\n")
			prompt.WriteString("8. 避免过于直接的表白，保持美感和意境\n")
		}
	}

	prompt.WriteString("\n请直接返回改善后的信件内容，不需要额外说明或格式标记。")

	return prompt.String()
}

// ProcessDelayedReplies processes delayed AI replies from the delay queue
func (s *AIService) ProcessDelayedReplies(ctx context.Context) (int, error) {
	log.Printf("[AIService] Processing delayed AI replies")
	
	// Query the delay queue database for ready-to-process AI reply tasks
	var delayRecords []models.DelayQueueRecord
	err := s.db.Where("task_type = ? AND status = ? AND execute_at <= ?", 
		"ai_reply", "pending", time.Now()).Find(&delayRecords).Error
	if err != nil {
		return 0, fmt.Errorf("failed to query delayed AI reply tasks: %w", err)
	}
	
	processedCount := 0
	for _, record := range delayRecords {
		// Parse task payload
		var aiReplyTask AIReplyTask
		if err := json.Unmarshal([]byte(record.Payload), &aiReplyTask); err != nil {
			log.Printf("Failed to unmarshal AI reply task %s: %v", record.ID, err)
			continue
		}
		
		// Process the AI reply
		if err := s.processDelayedAIReply(ctx, &aiReplyTask); err != nil {
			log.Printf("Failed to process delayed AI reply %s: %v", record.ID, err)
			// Mark as failed and increment retry count
			record.Status = "failed"
			record.RetryCount++
			s.db.Save(&record)
			continue
		}
		
		// Mark as completed
		record.Status = "completed"
		record.CompletedAt = time.Now()
		if err := s.db.Save(&record).Error; err != nil {
			log.Printf("Failed to update delay record status: %v", err)
		}
		
		processedCount++
	}
	
	log.Printf("[AIService] Successfully processed %d delayed AI replies", processedCount)
	return processedCount, nil
}

// processDelayedAIReply processes a single delayed AI reply task
func (s *AIService) processDelayedAIReply(ctx context.Context, task *AIReplyTask) error {
	// Create AI reply request
	aiReq := &models.AIReplyRequest{
		UserID:         task.UserID,
		PersonaID:      task.PersonaID,
		OriginalLetter: task.OriginalLetter,
	}
	
	// Generate the AI reply
	replyLetter, err := s.GenerateReply(ctx, aiReq)
	if err != nil {
		return fmt.Errorf("failed to generate AI reply: %w", err)
	}
	
	// The reply letter is already saved in GenerateReply method
	log.Printf("Successfully generated delayed AI reply for user %s", task.UserID)
	return nil
}

// AIReplyTask represents data structure for delayed AI reply tasks
type AIReplyTask struct {
	UserID         string `json:"user_id"`
	PersonaID      string `json:"persona_id"`
	OriginalLetter string `json:"original_letter"`
	ConversationID string `json:"conversation_id"`
}
