package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"openpenpal-backend/internal/config"
	"openpenpal-backend/internal/models"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UnifiedAIService 统一的AI服务，整合了所有AI功能
// 集成了配置管理、SOTA增强功能和原有的AI服务能力
type UnifiedAIService struct {
	*EnhancedAIService          // 继承SOTA增强功能
	configService      *ConfigService // 配置服务
	templateCache      map[string][]models.AIContentTemplate
	cacheLastUpdated   time.Time
	cacheTTL          time.Duration
}

// NewUnifiedAIService 创建统一AI服务实例
func NewUnifiedAIService(db *gorm.DB, config *config.Config) *UnifiedAIService {
	// 创建配置服务
	configService := NewConfigService(db)
	
	// 创建增强AI服务
	enhancedService := NewEnhancedAIService(db, config)
	
	// 创建统一服务
	unifiedService := &UnifiedAIService{
		EnhancedAIService: enhancedService,
		configService:     configService,
		templateCache:     make(map[string][]models.AIContentTemplate),
		cacheTTL:         5 * time.Minute,
	}

	log.Println("✅ [UnifiedAIService] 统一AI服务初始化完成")
	return unifiedService
}

// GetInspiration 获取写作灵感（使用配置化模板）
func (s *UnifiedAIService) GetInspiration(ctx context.Context, req *models.AIInspirationRequest) (*models.AIInspirationResponse, error) {
	log.Printf("🎯 [UnifiedAIService] 获取写作灵感，主题: %s, 数量: %d", req.Theme, req.Count)

	// 记录指标
	s.metrics.IncrementRequest()
	startTime := time.Now()

	defer func() {
		duration := time.Since(startTime)
		s.metrics.RecordResponseTime(duration)
	}()

	// 尝试从配置化模板获取灵感
	inspirations, err := s.getInspirationsFromConfig(req)
	if err != nil {
		log.Printf("⚠️ [UnifiedAIService] 从配置获取灵感失败，使用fallback: %v", err)
		s.metrics.IncrementFallback()
		return s.getFallbackInspiration(req), nil
	}

	// 如果配置化模板数量不足，补充AI生成的内容
	if len(inspirations) < req.Count {
		log.Printf("🤖 [UnifiedAIService] 配置模板不足，使用AI生成补充内容")
		aiInspirations, err := s.generateAIInspirations(ctx, req)
		if err == nil && len(aiInspirations) > 0 {
			inspirations = append(inspirations, aiInspirations...)
		}
	}

	// 确保数量不超过请求
	if len(inspirations) > req.Count {
		inspirations = inspirations[:req.Count]
	}

	s.metrics.IncrementSuccess()
	
	response := &models.AIInspirationResponse{
		Inspirations: inspirations,
		Metadata: map[string]interface{}{
			"source":           "unified_config",
			"template_count":   len(inspirations),
			"generation_time":  time.Since(startTime).Milliseconds(),
			"cache_used":      s.isCacheValid(),
		},
	}

	log.Printf("✅ [UnifiedAIService] 成功返回 %d 条灵感", len(inspirations))
	return response, nil
}

// GenerateReply 生成AI回信（使用配置化人设）
func (s *UnifiedAIService) GenerateReply(ctx context.Context, req *models.AIReplyRequest) (*models.Letter, error) {
	log.Printf("👤 [UnifiedAIService] 生成AI回信，人设: %s", req.Persona)

	// 获取人设配置
	personaConfig, err := s.configService.GetPersonaConfig(string(req.Persona))
	if err != nil {
		log.Printf("⚠️ [UnifiedAIService] 获取人设配置失败，使用默认: %v", err)
		personaConfig = s.getDefaultPersonaConfig(req.Persona)
	}

	// 获取系统提示词
	systemPrompt, err := s.configService.GetSystemPrompt("reply")
	if err != nil {
		log.Printf("⚠️ [UnifiedAIService] 获取系统提示词失败，使用默认: %v", err)
		systemPrompt = s.getDefaultSystemPrompt("reply")
	}

	// 构建AI提示词
	prompt := s.buildReplyPrompt(req, personaConfig, systemPrompt)

	// 调用AI生成回信
	content, err := s.callAIProvider(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("AI回信生成失败: %w", err)
	}

	// 构建回信Letter对象
	reply := &models.Letter{
		ID:        uuid.New().String(),
		Title:     s.generateReplyTitle(req.OriginalLetter.Title),
		Content:   content,
		Type:      models.LetterTypeReply,
		Status:    models.LetterStatusDraft,
		Metadata: map[string]interface{}{
			"ai_generated":    true,
			"persona":         string(req.Persona),
			"original_letter": req.OriginalLetter.ID,
			"generation_time": time.Now(),
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	log.Printf("✅ [UnifiedAIService] AI回信生成成功，长度: %d 字符", len(content))
	return reply, nil
}

// MatchPenPal 笔友匹配（使用配置化匹配算法，支持用户可控延迟）
func (s *UnifiedAIService) MatchPenPal(ctx context.Context, req *models.AIMatchRequest) (*models.AIMatchResponse, error) {
	log.Printf("💌 [UnifiedAIService] 执行笔友匹配，信件ID: %s, 延迟选项: %s", req.LetterID, req.DelayOption)

	// 计算用户选择的延迟时间
	delayMinutes := s.calculateUserDelay(req.DelayOption)
	
	// 获取匹配算法配置
	matchConfig, err := s.configService.GetConfig("matching", "algorithm")
	if err != nil {
		log.Printf("⚠️ [UnifiedAIService] 获取匹配算法配置失败，使用默认: %v", err)
	}

	// 如果有延迟要求，使用延迟队列
	if delayMinutes > 0 {
		log.Printf("🕐 [UnifiedAIService] 延迟 %d 分钟后执行匹配", delayMinutes)
		
		// 创建延迟任务
		task := &models.DelayQueueRecord{
			ID:           uuid.New().String(),
			TaskType:     "ai_match",
			Payload:      s.marshalMatchRequest(req),
			DelayedUntil: time.Now().Add(time.Duration(delayMinutes) * time.Minute),
			Status:       "pending",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		
		// 调度延迟任务
		if err := s.db.Create(task).Error; err != nil {
			log.Printf("❌ [UnifiedAIService] 创建延迟匹配任务失败: %v", err)
			// 降级：立即执行匹配
			return s.performImmediateMatch(ctx, req, matchConfig)
		}
		
		// 返回处理中状态
		return &models.AIMatchResponse{
			Status:  "processing",
			Message: fmt.Sprintf("正在为您寻找最合适的笔友，预计 %d 分钟后完成匹配...", delayMinutes),
			Metadata: map[string]interface{}{
				"delay_minutes": delayMinutes,
				"task_id":       task.ID,
			},
		}, nil
	}

	// 立即执行匹配
	return s.performImmediateMatch(ctx, req, matchConfig)
}

// 计算用户选择的延迟时间
func (s *UnifiedAIService) calculateUserDelay(delayOption string) int {
	switch delayOption {
	case "quick":
		// 1-10分钟随机延迟
		return rand.Intn(10) + 1
	case "normal":
		// 10-30分钟随机延迟
		return rand.Intn(21) + 10
	case "slow":
		// 30-60分钟随机延迟
		return rand.Intn(31) + 30
	default:
		// 默认无延迟（向后兼容）
		return 0
	}
}

// 执行立即匹配
func (s *UnifiedAIService) performImmediateMatch(ctx context.Context, req *models.AIMatchRequest, matchConfig *AIConfigData) (*models.AIMatchResponse, error) {
	// 调用原有的匹配逻辑（继承自EnhancedAIService）
	response, err := s.EnhancedAIService.MatchPenPal(ctx, req)
	if err != nil {
		return nil, err
	}

	// 使用配置增强匹配结果
	if matchConfig != nil {
		response = s.enhanceMatchResult(response, matchConfig)
	}

	return response, nil
}

// 序列化匹配请求
func (s *UnifiedAIService) marshalMatchRequest(req *models.AIMatchRequest) string {
	data, _ := json.Marshal(req)
	return string(data)
}

// GetPersonaList 获取可用人设列表（从配置）
func (s *UnifiedAIService) GetPersonaList() ([]models.AIPersonaInfo, error) {
	log.Println("👥 [UnifiedAIService] 获取人设列表")

	personas := []models.AIPersona{
		models.PersonaPoet, models.PersonaPhilosopher, models.PersonaArtist,
		models.PersonaScientist, models.PersonaTraveler, models.PersonaHistorian,
		models.PersonaMentor, models.PersonaFriend,
	}

	result := make([]models.AIPersonaInfo, 0, len(personas))

	for _, persona := range personas {
		personaConfig, err := s.configService.GetPersonaConfig(string(persona))
		if err != nil {
			// 如果配置不存在，使用默认配置
			personaConfig = s.getDefaultPersonaConfig(persona)
		}

		info := models.AIPersonaInfo{
			ID:          string(persona),
			Name:        personaConfig.Name,
			Description: personaConfig.Description,
			Style:       personaConfig.Style,
			Available:   true,
		}

		result = append(result, info)
	}

	log.Printf("✅ [UnifiedAIService] 返回 %d 个可用人设", len(result))
	return result, nil
}

// 从配置获取灵感内容
func (s *UnifiedAIService) getInspirationsFromConfig(req *models.AIInspirationRequest) ([]struct {
	ID     string   `json:"id"`
	Theme  string   `json:"theme"`
	Prompt string   `json:"prompt"`
	Style  string   `json:"style"`
	Tags   []string `json:"tags"`
}, error) {
	// 检查缓存
	if s.isCacheValid() {
		if cached, exists := s.templateCache["inspiration"]; exists {
			return s.selectTemplatesFromCache(cached, req), nil
		}
	}

	// 从配置服务获取模板
	templates, err := s.configService.GetTemplates("inspiration")
	if err != nil {
		return nil, err
	}

	// 更新缓存
	s.templateCache["inspiration"] = templates
	s.cacheLastUpdated = time.Now()

	// 根据请求筛选模板
	return s.selectTemplatesFromCache(templates, req), nil
}

// 从缓存中选择合适的模板
func (s *UnifiedAIService) selectTemplatesFromCache(templates []models.AIContentTemplate, req *models.AIInspirationRequest) []struct {
	ID     string   `json:"id"`
	Theme  string   `json:"theme"`
	Prompt string   `json:"prompt"`
	Style  string   `json:"style"`
	Tags   []string `json:"tags"`
} {
	var selected []models.AIContentTemplate

	// 如果指定了主题，优先选择匹配的模板
	if req.Theme != "" {
		for _, template := range templates {
			if strings.Contains(template.Category, req.Theme) ||
				strings.Contains(template.Title, req.Theme) ||
				strings.Contains(template.Content, req.Theme) {
				selected = append(selected, template)
			}
		}
	}

	// 如果没有匹配的模板，随机选择
	if len(selected) == 0 {
		selected = templates
	}

	// 随机打乱顺序
	rand.Shuffle(len(selected), func(i, j int) {
		selected[i], selected[j] = selected[j], selected[i]
	})

	// 限制数量
	count := req.Count
	if count == 0 || count > 5 {
		count = 3
	}
	if len(selected) > count {
		selected = selected[:count]
	}

	// 转换为响应格式
	result := make([]struct {
		ID     string   `json:"id"`
		Theme  string   `json:"theme"`
		Prompt string   `json:"prompt"`
		Style  string   `json:"style"`
		Tags   []string `json:"tags"`
	}, len(selected))

	for i, template := range selected {
		// 更新使用统计
		go s.configService.UpdateTemplateUsage(template.ID)

		result[i] = struct {
			ID     string   `json:"id"`
			Theme  string   `json:"theme"`
			Prompt string   `json:"prompt"`
			Style  string   `json:"style"`
			Tags   []string `json:"tags"`
		}{
			ID:     template.ID,
			Theme:  template.Category,
			Prompt: template.Content,
			Style:  s.getTemplateStyle(template),
			Tags:   template.Tags,
		}
	}

	return result
}

// 生成AI补充灵感
func (s *UnifiedAIService) generateAIInspirations(ctx context.Context, req *models.AIInspirationRequest) ([]struct {
	ID     string   `json:"id"`
	Theme  string   `json:"theme"`
	Prompt string   `json:"prompt"`
	Style  string   `json:"style"`
	Tags   []string `json:"tags"`
}, error) {
	// 获取AI生成提示词配置
	systemPrompt, err := s.configService.GetSystemPrompt("inspiration")
	if err != nil {
		systemPrompt = s.getDefaultSystemPrompt("inspiration")
	}

	// 构建AI生成提示
	prompt := fmt.Sprintf("%s\n\n请为主题'%s'生成一条写作灵感。要求：温暖人文、激发创作、具体可操作。", 
		systemPrompt.Prompt, req.Theme)

	// 调用AI生成
	content, err := s.callAIProvider(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// 解析生成的内容
	inspiration := struct {
		ID     string   `json:"id"`
		Theme  string   `json:"theme"`
		Prompt string   `json:"prompt"`
		Style  string   `json:"style"`
		Tags   []string `json:"tags"`
	}{
		ID:     uuid.New().String(),
		Theme:  req.Theme,
		Prompt: content,
		Style:  "AI生成",
		Tags:   []string{"AI生成", req.Theme},
	}

	return []struct {
		ID     string   `json:"id"`
		Theme  string   `json:"theme"`
		Prompt string   `json:"prompt"`
		Style  string   `json:"style"`
		Tags   []string `json:"tags"`
	}{inspiration}, nil
}

// 构建回信提示词
func (s *UnifiedAIService) buildReplyPrompt(req *models.AIReplyRequest, personaConfig *PersonaConfig, systemPrompt *SystemPromptConfig) string {
	return fmt.Sprintf(`%s

人设设定：%s
写作风格：%s

原始信件：
标题：%s
内容：%s

请以这个人设的身份，写一封温暖而真诚的回信。要求：
1. 体现人设特色和风格
2. 针对原信件内容进行回应
3. 保持温暖人文的语气
4. 长度控制在300-500字

回信内容：`,
		systemPrompt.Prompt,
		personaConfig.Description,
		personaConfig.Style,
		req.OriginalLetter.Title,
		req.OriginalLetter.Content)
}

// 生成回信标题
func (s *UnifiedAIService) generateReplyTitle(originalTitle string) string {
	if strings.HasPrefix(originalTitle, "Re: ") {
		return originalTitle
	}
	return fmt.Sprintf("Re: %s", originalTitle)
}

// 增强匹配结果
func (s *UnifiedAIService) enhanceMatchResult(response *models.AIMatchResponse, matchConfig *AIConfigData) *models.AIMatchResponse {
	// 这里可以根据配置调整匹配算法的权重、评分等
	// 暂时返回原始结果
	return response
}

// 获取默认人设配置
func (s *UnifiedAIService) getDefaultPersonaConfig(persona models.AIPersona) *PersonaConfig {
	defaultConfigs := map[models.AIPersona]*PersonaConfig{
		models.PersonaFriend: {
			Name:        "知心朋友",
			Description: "温暖贴心的好朋友，总是愿意倾听和陪伴",
			Prompt:      "我是你的知心朋友，总是愿意倾听你的心声。我会用最真诚和温暖的态度与你分享生活中的点点滴滴。",
			Style:       "温暖亲切",
		},
		models.PersonaMentor: {
			Name:        "人生导师",
			Description: "温和睿智的人生指导者，乐于分享人生智慧和经验",
			Prompt:      "我是你的人生导师，拥有丰富的人生阅历。我会用温和而智慧的方式为你答疑解惑，分享人生的智慧。",
			Style:       "温和智慧",
		},
	}

	if config, exists := defaultConfigs[persona]; exists {
		return config
	}

	// 默认朋友人设
	return defaultConfigs[models.PersonaFriend]
}

// 获取默认系统提示词
func (s *UnifiedAIService) getDefaultSystemPrompt(promptType string) *SystemPromptConfig {
	return &SystemPromptConfig{
		Prompt:        "你是OpenPenPal的AI助手，请用温暖友好的语气回应用户。",
		Temperature:   0.9,
		MaxTokens:     1000,
		ContextWindow: 4000,
		Guidelines:    []string{"保持温暖友好", "避免AI腔调"},
	}
}

// 获取模板样式
func (s *UnifiedAIService) getTemplateStyle(template models.AIContentTemplate) string {
	if style, exists := template.Metadata["style"]; exists {
		if styleStr, ok := style.(string); ok {
			return styleStr
		}
	}
	return template.Category
}

// 检查缓存是否有效
func (s *UnifiedAIService) isCacheValid() bool {
	return time.Since(s.cacheLastUpdated) < s.cacheTTL
}

// 调用AI提供商（继承自原有实现）
func (s *UnifiedAIService) callAIProvider(ctx context.Context, prompt string) (string, error) {
	// 获取活跃的AI配置
	config, err := s.GetActiveProvider()
	if err != nil {
		return "", fmt.Errorf("获取AI配置失败: %w", err)
	}

	// 调用相应的AI提供商
	return s.callAIAPI(ctx, config, prompt, models.TaskTypeReply)
}

// getFallbackInspiration 保留原有的fallback机制
func (s *UnifiedAIService) getFallbackInspiration(req *models.AIInspirationRequest) *models.AIInspirationResponse {
	// 调用原有的generateLocalInspirations方法
	return s.AIService.generateLocalInspirations(req)
}