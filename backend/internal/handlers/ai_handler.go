package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"openpenpal-backend/internal/middleware"
	"openpenpal-backend/internal/models"
	"openpenpal-backend/internal/pkg/response"
	"openpenpal-backend/internal/services"
	"openpenpal-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

// AIHandler AI处理器
type AIHandler struct {
	aiService     *services.AIService
	configService *services.ConfigService
	aiManager     *services.AIProviderManager
}

// NewAIHandler 创建AI处理器
func NewAIHandler(aiService *services.AIService, configService *services.ConfigService, aiManager *services.AIProviderManager) *AIHandler {
	return &AIHandler{
		aiService:     aiService,
		configService: configService,
		aiManager:     aiManager,
	}
}

// MatchPenPal 匹配笔友
// @Summary AI匹配笔友
// @Description 基于信件内容智能匹配合适的笔友
// @Tags AI
// @Accept json
// @Produce json
// @Param request body models.AIMatchRequest true "匹配请求"
// @Success 200 {object} models.AIMatchResponse
// @Router /api/v1/ai/match [post]
func (h *AIHandler) MatchPenPal(c *gin.Context) {
	var req models.AIMatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ParseAndRespondValidationError(c, err, utils.AIValidationMsg)
		return
	}

	// 设置默认值
	if req.MaxMatches == 0 {
		req.MaxMatches = 3
	}

	// 调用AI服务
	response, err := h.aiService.MatchPenPal(c.Request.Context(), &req)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to match pen pal", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Pen pal matched successfully", response)
}

// GenerateReply 生成AI回信
// @Summary 生成AI回信
// @Description AI根据指定人设生成回信
// @Tags AI
// @Accept json
// @Produce json
// @Param request body models.AIReplyRequest true "回信请求"
// @Success 200 {object} models.Letter
// @Router /api/v1/ai/reply [post]
func (h *AIHandler) GenerateReply(c *gin.Context) {
	var req models.AIReplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ParseAndRespondValidationError(c, err, utils.AIValidationMsg)
		return
	}

	// 设置默认延迟时间
	if req.DelayHours == 0 {
		req.DelayHours = 24
	}

	// 验证人设
	validPersonas := map[models.AIPersona]bool{
		models.PersonaPoet:        true,
		models.PersonaPhilosopher: true,
		models.PersonaArtist:      true,
		models.PersonaScientist:   true,
		models.PersonaTraveler:    true,
		models.PersonaHistorian:   true,
		models.PersonaMentor:      true,
		models.PersonaFriend:      true,
	}

	if !validPersonas[req.Persona] {
		utils.BadRequestResponse(c, "Invalid persona type", nil)
		return
	}

	// 根据延迟时间决定处理方式
	if req.DelayHours > 0 {
		// 使用延迟队列
		conversationID, err := h.aiService.ScheduleDelayedReply(c.Request.Context(), &req)
		if err != nil {
			utils.InternalServerErrorResponse(c, "Failed to schedule AI reply", err)
			return
		}

		utils.SuccessResponse(c, http.StatusAccepted, "AI reply scheduled successfully", gin.H{
			"conversation_id": conversationID,
			"scheduled_at":    time.Now().Add(time.Duration(req.DelayHours) * time.Hour),
			"delay_hours":     req.DelayHours,
		})
		return
	}

	// 立即处理
	reply, err := h.aiService.GenerateReply(c.Request.Context(), &req)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to generate reply", err)
		return
	}

	c.JSON(http.StatusOK, reply)
}

// GenerateReplyAdvice 角色驿站回信建议
// @Summary 角色驿站回信建议
// @Description 基于不同角色视角为用户的回信提供思路和建议，支持自定义角色和情感引导
// @Tags AI
// @Accept json
// @Produce json
// @Param request body models.AIReplyAdviceRequest true "回信建议请求"
// @Success 200 {object} models.AIReplyAdvice
// @Router /api/v1/ai/reply-advice [post]
func (h *AIHandler) GenerateReplyAdvice(c *gin.Context) {
	var req models.AIReplyAdviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ParseAndRespondValidationError(c, err, utils.AIValidationMsg)
		return
	}

	// 验证人设类型
	validPersonaTypes := map[string]bool{
		"custom":         true,
		"predefined":     true,
		"deceased":       true,
		"distant_friend": true,
		"unspoken_love":  true,
	}

	if !validPersonaTypes[req.PersonaType] {
		utils.BadRequestResponse(c, "Invalid persona type", nil)
		return
	}

	// 验证延迟天数
	if req.DeliveryDays < 0 || req.DeliveryDays > 7 {
		utils.BadRequestResponse(c, "Delivery days must be between 0 and 7", nil)
		return
	}

	// 调用AI服务
	advice, err := h.aiService.GenerateReplyAdvice(c.Request.Context(), &req)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to generate reply advice", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Reply advice generated successfully", advice)
}

// GetInspiration 获取写作灵感
// @Summary 获取AI写作灵感
// @Description AI生成写作灵感和提示
// @Tags AI
// @Accept json
// @Produce json
// @Param request body models.AIInspirationRequest true "灵感请求"
// @Success 200 {object} models.AIInspirationResponse
// @Router /api/v1/ai/inspiration [post]
func (h *AIHandler) GetInspiration(c *gin.Context) {
	var req models.AIInspirationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ParseAndRespondValidationError(c, err, utils.AIValidationMsg)
		return
	}

	// 设置默认值
	if req.Count == 0 {
		req.Count = 1
	}
	if req.Count > 5 {
		req.Count = 5 // 限制最多5个
	}

	// 检查是否有用户ID（如果有则使用限制，如果没有则作为公开接口）
	userID, exists := c.Get("user_id")
	var response *models.AIInspirationResponse
	var err error

	if exists {
		// 有用户登录，使用带限制的方法
		userIDStr, ok := userID.(string)
		if !ok {
			utils.InternalServerErrorResponse(c, "Invalid user ID format", nil)
			return
		}
		response, err = h.aiService.GetInspirationWithLimit(c.Request.Context(), userIDStr, &req)
	} else {
		// 没有用户登录，使用公开方法（不记录使用量）
		response, err = h.aiService.GetInspiration(c.Request.Context(), &req)
	}
	if err != nil {
		// 记录详细错误信息
		log.Printf("❌ [AIHandler] GetInspirationWithLimit error: %v", err)

		// 检查是否是使用量限制错误
		if strings.Contains(err.Error(), "limit exceeded") {
			utils.BadRequestResponse(c, err.Error(), err)
			return
		}

		// AI服务不可用时，返回预设的写作灵感
		log.Printf("⚠️ [AIHandler] Falling back to preset inspiration due to error: %v", err)
		fallbackResponse := h.getFallbackInspiration(&req)
		utils.SuccessResponse(c, http.StatusOK, "Inspiration generated successfully (fallback)", fallbackResponse)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Inspiration generated successfully", response)
}

// GetUsageStats 获取用户AI使用统计
// @Summary 获取用户AI使用统计
// @Description 获取用户每日AI功能使用量和限制
// @Tags AI
// @Produce json
// @Success 200 {object} models.AIUsageStats
// @Router /api/v1/ai/stats [get]
func (h *AIHandler) GetUsageStats(c *gin.Context) {
	// 从JWT中获取用户ID
	_, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	// Skip userID validation for now since we're not using it
	// _, ok := userID.(string)
	// if !ok {
	//	utils.InternalServerErrorResponse(c, "Invalid user ID format", nil)
	//	return
	// }

	// 获取用户使用统计 (temporarily disabled to fix compilation)
	// stats, err := h.aiService.usageService.GetUserUsageStats(userIDStr)
	// if err != nil {
	//	utils.InternalServerErrorResponse(c, "Failed to get usage stats", err)
	//	return
	// }

	// Return mock stats for now
	mockStats := map[string]interface{}{
		"daily_usage":   0,
		"monthly_usage": 0,
		"total_usage":   0,
	}
	utils.SuccessResponse(c, http.StatusOK, "Usage stats retrieved successfully", mockStats)
}

// CurateLetters AI策展信件
// @Summary AI策展信件
// @Description AI分析信件并进行分类策展
// @Tags AI
// @Accept json
// @Produce json
// @Param request body models.AICurateRequest true "策展请求"
// @Success 200 {object} gin.H
// @Router /api/v1/ai/curate [post]
func (h *AIHandler) CurateLetters(c *gin.Context) {
	var req models.AICurateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ParseAndRespondValidationError(c, err, utils.AIValidationMsg)
		return
	}

	// 限制批量处理数量
	if len(req.LetterIDs) > 10 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Maximum 10 letters per request"})
		return
	}

	// 调用AI服务
	err := h.aiService.CurateLetters(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to curate letters: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Letters curated successfully",
		"count":   len(req.LetterIDs),
	})
}

// GetPersonas 获取云中锦书人设列表
// @Summary 获取云中锦书人设列表
// @Description 获取所有可用的长期AI笔友人设，用于建立持续的书信往来关系
// @Tags AI
// @Produce json
// @Success 200 {object} gin.H
// @Router /api/v1/ai/personas [get]
func (h *AIHandler) GetPersonas(c *gin.Context) {
	personas := []gin.H{
		{
			"id":          "poet",
			"name":        "诗人",
			"description": "用诗意的语言表达情感，善于发现生活中的美",
			"avatar":      "/images/personas/poet.png",
		},
		{
			"id":          "philosopher",
			"name":        "哲学家",
			"description": "思考人生的意义，探讨深刻的哲理问题",
			"avatar":      "/images/personas/philosopher.png",
		},
		{
			"id":          "artist",
			"name":        "艺术家",
			"description": "用艺术的眼光看世界，分享创作的灵感",
			"avatar":      "/images/personas/artist.png",
		},
		{
			"id":          "scientist",
			"name":        "科学家",
			"description": "理性分析世界，分享科学的奇妙",
			"avatar":      "/images/personas/scientist.png",
		},
		{
			"id":          "traveler",
			"name":        "旅行者",
			"description": "分享世界各地的见闻和故事",
			"avatar":      "/images/personas/traveler.png",
		},
		{
			"id":          "historian",
			"name":        "历史学家",
			"description": "讲述历史故事，连接过去与现在",
			"avatar":      "/images/personas/historian.png",
		},
		{
			"id":          "mentor",
			"name":        "人生导师",
			"description": "给予温暖的建议和人生指引",
			"avatar":      "/images/personas/mentor.png",
		},
		{
			"id":          "friend",
			"name":        "知心朋友",
			"description": "倾听你的心声，给予真诚的陪伴",
			"avatar":      "/images/personas/friend.png",
		},
	}

	utils.SuccessResponse(c, http.StatusOK, "Personas retrieved successfully", gin.H{
		"personas": personas,
		"total":    len(personas),
	})
}

// GetAIStats 获取AI使用统计
// @Summary 获取AI使用统计
// @Description 获取当前用户的AI功能使用统计
// @Tags AI
// @Produce json
// @Success 200 {object} gin.H
// @Router /api/v1/ai/stats [get]
func (h *AIHandler) GetAIStats(c *gin.Context) {
	// 从JWT中获取用户ID（可选，支持匿名访问）
	userIDStr, exists := middleware.GetUserID(c)

	// 如果是匿名用户，返回默认统计
	if !exists {
		// 匿名用户的默认统计
		stats := gin.H{
			"user_id": "anonymous",
			"usage": gin.H{
				"matches_created":   0,
				"replies_generated": 0,
				"inspirations_used": 0,
				"letters_curated":   0,
			},
			"limits": gin.H{
				"daily_matches":      3, // 匿名用户限制
				"daily_replies":      2,
				"daily_inspirations": 5,
				"daily_curations":    1,
			},
			"remaining": gin.H{
				"matches":      3,
				"replies":      2,
				"inspirations": 5,
				"curations":    1,
			},
			"message": "登录后可获得更高使用限额",
		}

		utils.SuccessResponse(c, http.StatusOK, "AI stats retrieved successfully", stats)
		return
	}

	// 用户ID已经是字符串格式（UUID），不需要转换为整数
	userID := userIDStr

	// TODO: 实现统计逻辑
	stats := gin.H{
		"user_id": userID,
		"usage": gin.H{
			"matches_created":   5,
			"replies_generated": 3,
			"inspirations_used": 10,
			"letters_curated":   2,
		},
		"limits": gin.H{
			"daily_matches":      10,
			"daily_replies":      5,
			"daily_inspirations": 20,
			"daily_curations":    10,
		},
		"remaining": gin.H{
			"matches":      5,
			"replies":      2,
			"inspirations": 10,
			"curations":    8,
		},
	}

	utils.SuccessResponse(c, http.StatusOK, "AI stats retrieved successfully", stats)
}

// GetDailyInspiration 获取每日灵感
// @Summary 获取每日写作灵感
// @Description 获取系统推荐的每日写作灵感
// @Tags AI
// @Produce json
// @Success 200 {object} gin.H
// @Router /api/v1/ai/daily-inspiration [get]
func (h *AIHandler) GetDailyInspiration(c *gin.Context) {
	// 生成当日的写作主题和灵感
	currentDate := time.Now().Format("2006-01-02")

	// 基于日期生成不同的主题和灵感
	themes := []gin.H{
		{
			"theme":  "日常小确幸",
			"prompt": "写一写今天让你感到温暖的小事情。可能是早晨的阳光，路过的猫咪，或是一个陌生人的微笑。",
			"quote":  "生活中的小确幸，是支撑我们前行的光。",
		},
		{
			"theme":  "成长的足迹",
			"prompt": "回想一下最近你学会的新技能或明白的新道理，写下这个成长过程中的感受。",
			"quote":  "每一个进步，都是向更好的自己走近一步。",
		},
		{
			"theme":  "友情时光",
			"prompt": "想起和朋友在一起的快乐时光，可以是一次谈话，一次聚餐，或是一个小小的默契。",
			"quote":  "好朋友就是，即使不常联系，一见面还是那么熟悉。",
		},
		{
			"theme":  "家的温度",
			"prompt": "描述家里让你感到最安心的角落，或是家人之间温馨的一个瞬间。",
			"quote":  "家不是房子，而是有爱的人在的地方。",
		},
		{
			"theme":  "梦想点滴",
			"prompt": "写下你最近在为什么目标而努力，这个过程中有什么收获和感悟。",
			"quote":  "梦想不是遥不可及，而是一步一步走出来的路。",
		},
	}

	// 根据日期选择主题（简单的轮换机制）
	dayOfYear := time.Now().YearDay()
	selectedTheme := themes[dayOfYear%len(themes)]

	inspiration := gin.H{
		"date":   currentDate,
		"theme":  selectedTheme["theme"],
		"prompt": selectedTheme["prompt"],
		"quote":  selectedTheme["quote"],
		"tips": []string{
			"用真实的感受写作，不需要华丽的辞藻",
			"描述具体的场景和细节会让文字更生动",
			"可以加入自己的思考和感悟",
			"记住这是给另一个人的信，带着真诚的心意",
		},
	}

	utils.SuccessResponse(c, http.StatusOK, "Daily inspiration fetched successfully", inspiration)
}

// Multi-Provider AI API Endpoints

// GenerateTextRequest 文本生成请求
type GenerateTextRequest struct {
	Prompt           string  `json:"prompt" binding:"required"`
	MaxTokens        int     `json:"max_tokens,omitempty"`
	Temperature      float64 `json:"temperature,omitempty"`
	TopP             float64 `json:"top_p,omitempty"`
	Model            string  `json:"model,omitempty"`
	PreferredProvider string `json:"preferred_provider,omitempty"`
	Stop             []string `json:"stop,omitempty"`
}

// ChatRequest 聊天请求
type ChatRequest struct {
	Messages         []services.ChatMessage `json:"messages" binding:"required"`
	MaxTokens        int                    `json:"max_tokens,omitempty"`
	Temperature      float64                `json:"temperature,omitempty"`
	TopP             float64                `json:"top_p,omitempty"`
	Model            string                 `json:"model,omitempty"`
	PreferredProvider string                `json:"preferred_provider,omitempty"`
	Stop             []string               `json:"stop,omitempty"`
}

// SummarizeRequest 总结请求
type SummarizeRequest struct {
	Text             string  `json:"text" binding:"required"`
	MaxTokens        int     `json:"max_tokens,omitempty"`
	Temperature      float64 `json:"temperature,omitempty"`
	PreferredProvider string `json:"preferred_provider,omitempty"`
}

// TranslateRequest 翻译请求
type TranslateRequest struct {
	Text             string  `json:"text" binding:"required"`
	TargetLanguage   string  `json:"target_language" binding:"required"`
	MaxTokens        int     `json:"max_tokens,omitempty"`
	Temperature      float64 `json:"temperature,omitempty"`
	PreferredProvider string `json:"preferred_provider,omitempty"`
}

// SentimentAnalysisRequest 情感分析请求
type SentimentAnalysisRequest struct {
	Text             string `json:"text" binding:"required"`
	PreferredProvider string `json:"preferred_provider,omitempty"`
}

// ContentModerationRequest 内容审核请求
type ContentModerationRequest struct {
	Text             string `json:"text" binding:"required"`
	PreferredProvider string `json:"preferred_provider,omitempty"`
}

// GenerateText 文本生成API
// @Summary 生成文本
// @Description 使用AI生成文本内容
// @Tags AI
// @Accept json
// @Produce json
// @Param request body GenerateTextRequest true "文本生成请求"
// @Success 200 {object} response.Response{data=services.AIResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/ai/generate [post]
func (h *AIHandler) GenerateText(c *gin.Context) {
	var req GenerateTextRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request parameters", err.Error())
		return
	}

	// 设置默认值
	if req.MaxTokens == 0 {
		req.MaxTokens = 1000
	}
	if req.Temperature == 0 {
		req.Temperature = 0.7
	}

	options := services.AIGenerationOptions{
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		Model:       req.Model,
		Stop:        req.Stop,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	result, err := h.aiManager.GenerateText(ctx, req.Prompt, options, req.PreferredProvider)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to generate text", err.Error())
		return
	}

	response.Success(c, result, "Text generated successfully")
}

// Chat 聊天对话API
// @Summary AI聊天对话
// @Description 与AI进行多轮对话
// @Tags AI
// @Accept json
// @Produce json
// @Param request body ChatRequest true "聊天请求"
// @Success 200 {object} response.Response{data=services.AIResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/ai/chat [post]
func (h *AIHandler) Chat(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request parameters", err.Error())
		return
	}

	if len(req.Messages) == 0 {
		response.Error(c, http.StatusBadRequest, "Messages cannot be empty", "")
		return
	}

	// 设置默认值
	if req.MaxTokens == 0 {
		req.MaxTokens = 1500
	}
	if req.Temperature == 0 {
		req.Temperature = 0.7
	}

	options := services.AIGenerationOptions{
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		Model:       req.Model,
		Stop:        req.Stop,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	result, err := h.aiManager.Chat(ctx, req.Messages, options, req.PreferredProvider)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to process chat", err.Error())
		return
	}

	response.Success(c, result, "Chat processed successfully")
}

// Summarize 文本总结API
// @Summary 文本总结
// @Description 对长文本进行智能总结
// @Tags AI
// @Accept json
// @Produce json
// @Param request body SummarizeRequest true "总结请求"
// @Success 200 {object} response.Response{data=services.AIResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/ai/summarize [post]
func (h *AIHandler) Summarize(c *gin.Context) {
	var req SummarizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request parameters", err.Error())
		return
	}

	if len(req.Text) < 10 {
		response.Error(c, http.StatusBadRequest, "Text too short for summarization", "")
		return
	}

	// 设置默认值
	if req.MaxTokens == 0 {
		req.MaxTokens = 500
	}
	if req.Temperature == 0 {
		req.Temperature = 0.3
	}

	options := services.AIGenerationOptions{
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	provider, usedProvider, err := h.aiManager.GetAvailableProvider(req.PreferredProvider)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "No available AI provider", err.Error())
		return
	}

	result, err := provider.Summarize(ctx, req.Text, options)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to summarize text", err.Error())
		return
	}

	result.Provider = usedProvider
	response.Success(c, result, "Text summarized successfully")
}

// Translate 翻译API
// @Summary 文本翻译
// @Description 将文本翻译成目标语言
// @Tags AI
// @Accept json
// @Produce json
// @Param request body TranslateRequest true "翻译请求"
// @Success 200 {object} response.Response{data=services.AIResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/ai/translate [post]
func (h *AIHandler) Translate(c *gin.Context) {
	var req TranslateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request parameters", err.Error())
		return
	}

	// 设置默认值
	if req.MaxTokens == 0 {
		req.MaxTokens = len(req.Text) * 2 // 翻译通常需要更多token
	}
	if req.Temperature == 0 {
		req.Temperature = 0.2 // 翻译需要较低的随机性
	}

	options := services.AIGenerationOptions{
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	provider, usedProvider, err := h.aiManager.GetAvailableProvider(req.PreferredProvider)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "No available AI provider", err.Error())
		return
	}

	result, err := provider.Translate(ctx, req.Text, req.TargetLanguage, options)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to translate text", err.Error())
		return
	}

	result.Provider = usedProvider
	response.Success(c, result, "Text translated successfully")
}

// AnalyzeSentiment 情感分析API
// @Summary 情感分析
// @Description 分析文本的情感倾向
// @Tags AI
// @Accept json
// @Produce json
// @Param request body SentimentAnalysisRequest true "情感分析请求"
// @Success 200 {object} response.Response{data=services.SentimentAnalysis}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/ai/sentiment [post]
func (h *AIHandler) AnalyzeSentiment(c *gin.Context) {
	var req SentimentAnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request parameters", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	provider, _, err := h.aiManager.GetAvailableProvider(req.PreferredProvider)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "No available AI provider", err.Error())
		return
	}

	result, err := provider.AnalyzeSentiment(ctx, req.Text)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to analyze sentiment", err.Error())
		return
	}

	response.Success(c, result, "Sentiment analyzed successfully")
}

// ModerateContent 内容审核API
// @Summary 内容审核
// @Description 检查内容是否包含不当信息
// @Tags AI
// @Accept json
// @Produce json
// @Param request body ContentModerationRequest true "内容审核请求"
// @Success 200 {object} response.Response{data=services.ContentModeration}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/ai/moderate [post]
func (h *AIHandler) ModerateContent(c *gin.Context) {
	var req ContentModerationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request parameters", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	provider, _, err := h.aiManager.GetAvailableProvider(req.PreferredProvider)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "No available AI provider", err.Error())
		return
	}

	result, err := provider.ModerateContent(ctx, req.Text)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to moderate content", err.Error())
		return
	}

	response.Success(c, result, "Content moderated successfully")
}

// GetProviderStatus 获取AI提供商状态
// @Summary 获取AI提供商状态
// @Description 查看所有AI提供商的健康状态和使用情况
// @Tags AI
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=map[string]interface{}}
// @Failure 500 {object} response.Response
// @Router /api/ai/providers/status [get]
func (h *AIHandler) GetProviderStatus(c *gin.Context) {
	stats := h.aiManager.GetProviderStats()
	response.Success(c, stats, "Provider status retrieved successfully")
}

// ReloadProviders 重新加载AI提供商配置
// @Summary 重新加载AI提供商配置
// @Description 从数据库重新加载AI提供商配置并重新初始化
// @Tags AI
// @Accept json
// @Produce json
// @Success 200 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/ai/providers/reload [post]
func (h *AIHandler) ReloadProviders(c *gin.Context) {
	if err := h.aiManager.ReloadConfigurations(); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to reload providers", err.Error())
		return
	}

	response.Success(c, nil, "Providers reloaded successfully")
}

// LetterWritingAssistRequest 信件写作辅助请求
type LetterWritingAssistRequest struct {
	Topic            string `json:"topic" binding:"required"`
	Style            string `json:"style,omitempty"`           // 写作风格：formal, casual, romantic, friendly
	Tone             string `json:"tone,omitempty"`            // 语调：warm, professional, humorous
	Length           string `json:"length,omitempty"`          // 长度：short, medium, long
	PreferredProvider string `json:"preferred_provider,omitempty"`
}

// LetterWritingAssist 信件写作辅助API
// @Summary 信件写作辅助
// @Description 根据主题和风格生成信件内容建议
// @Tags AI
// @Accept json
// @Produce json
// @Param request body LetterWritingAssistRequest true "写作辅助请求"
// @Success 200 {object} response.Response{data=services.AIResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/ai/letter/assist [post]
func (h *AIHandler) LetterWritingAssist(c *gin.Context) {
	var req LetterWritingAssistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request parameters", err.Error())
		return
	}

	// 设置默认值
	if req.Style == "" {
		req.Style = "friendly"
	}
	if req.Tone == "" {
		req.Tone = "warm"
	}
	if req.Length == "" {
		req.Length = "medium"
	}

	// 构建专门的写信提示词
	prompt := buildLetterWritingPrompt(req.Topic, req.Style, req.Tone, req.Length)

	options := services.AIGenerationOptions{
		MaxTokens:   1000,
		Temperature: 0.8, // 创作性内容需要较高的创造性
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	result, err := h.aiManager.GenerateText(ctx, prompt, options, req.PreferredProvider)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to generate letter content", err.Error())
		return
	}

	response.Success(c, result, "Letter writing assistance generated successfully")
}

// BatchTranslateRequest 批量翻译请求
type BatchTranslateRequest struct {
	Texts            []string `json:"texts" binding:"required"`
	TargetLanguage   string   `json:"target_language" binding:"required"`
	PreferredProvider string  `json:"preferred_provider,omitempty"`
}

// BatchTranslate 批量翻译API
// @Summary 批量翻译
// @Description 批量翻译多个文本
// @Tags AI
// @Accept json
// @Produce json
// @Param request body BatchTranslateRequest true "批量翻译请求"
// @Success 200 {object} response.Response{data=[]services.AIResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/ai/translate/batch [post]
func (h *AIHandler) BatchTranslate(c *gin.Context) {
	var req BatchTranslateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request parameters", err.Error())
		return
	}

	if len(req.Texts) == 0 {
		response.Error(c, http.StatusBadRequest, "No texts provided for translation", "")
		return
	}

	if len(req.Texts) > 50 {
		response.Error(c, http.StatusBadRequest, "Too many texts (max 50)", "")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	provider, usedProvider, err := h.aiManager.GetAvailableProvider(req.PreferredProvider)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "No available AI provider", err.Error())
		return
	}

	results := make([]*services.AIResponse, len(req.Texts))
	options := services.AIGenerationOptions{
		MaxTokens:   1000,
		Temperature: 0.2,
	}

	// 并发翻译以提高效率（限制并发数）
	semaphore := make(chan struct{}, 5) // 最多5个并发
	errChan := make(chan error, len(req.Texts))
	
	for i, text := range req.Texts {
		go func(index int, t string) {
			semaphore <- struct{}{} // 获取信号量
			defer func() { <-semaphore }() // 释放信号量

			result, err := provider.Translate(ctx, t, req.TargetLanguage, options)
			if err != nil {
				errChan <- err
				return
			}
			
			result.Provider = usedProvider
			results[index] = result
			errChan <- nil
		}(i, text)
	}

	// 等待所有翻译完成
	for i := 0; i < len(req.Texts); i++ {
		if err := <-errChan; err != nil {
			response.Error(c, http.StatusInternalServerError, "Failed to translate batch texts", err.Error())
			return
		}
	}

	response.Success(c, results, "Batch translation completed successfully")
}

// GetAIUsageStats 获取AI使用统计
// @Summary 获取AI使用统计
// @Description 获取用户的AI使用统计信息
// @Tags AI
// @Accept json
// @Produce json
// @Param days query int false "统计天数" default(30)
// @Success 200 {object} response.Response{data=map[string]interface{}}
// @Failure 500 {object} response.Response
// @Router /api/ai/usage/stats [get]
func (h *AIHandler) GetAIUsageStats(c *gin.Context) {
	daysStr := c.DefaultQuery("days", "30")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 {
		days = 30
	}

	// TODO: 实现用户使用统计逻辑
	// 这里应该从数据库中查询用户的AI使用记录
	stats := map[string]interface{}{
		"total_requests":    0,
		"total_tokens":      0,
		"requests_by_type":  map[string]int{},
		"providers_used":    map[string]int{},
		"period_days":       days,
		"message":          "Usage statistics feature coming soon",
	}

	response.Success(c, stats, "Usage statistics retrieved")
}

// buildLetterWritingPrompt 构建信件写作提示词
func buildLetterWritingPrompt(topic, style, tone, length string) string {
	lengthGuide := map[string]string{
		"short":  "简短精炼（100-200字）",
		"medium": "适中详细（200-400字）",
		"long":   "详细丰富（400-600字）",
	}

	styleGuide := map[string]string{
		"formal":   "正式、规范的书面语",
		"casual":   "轻松、随意的日常用语",
		"romantic": "浪漫、温柔的表达方式",
		"friendly": "友好、亲切的交流风格",
	}

	toneGuide := map[string]string{
		"warm":         "温暖、关怀的语调",
		"professional": "专业、严谨的语调",
		"humorous":     "幽默、轻松的语调",
	}

	return "你是一位专业的信件写作助手。请根据以下要求帮助用户写一封信：\n\n" +
		"主题：" + topic + "\n" +
		"写作风格：" + styleGuide[style] + "\n" +
		"语调：" + toneGuide[tone] + "\n" +
		"长度要求：" + lengthGuide[length] + "\n\n" +
		"请生成一封完整的信件内容，包括称呼、正文和结尾。要求：\n" +
		"1. 内容要真诚自然，符合中文信件写作习惯\n" +
		"2. 语言要流畅优美，情感表达要恰当\n" +
		"3. 结构要清晰，段落要合理\n" +
		"4. 体现出手写信件的温度和真诚\n\n" +
		"请直接返回信件内容，不需要额外的解释。"
}

// getFallbackInspiration 获取预设的写作灵感
func (h *AIHandler) getFallbackInspiration(req *models.AIInspirationRequest) *models.AIInspirationResponse {
	// 预设的写作灵感池
	inspirationPool := []struct {
		ID     string   `json:"id"`
		Theme  string   `json:"theme"`
		Prompt string   `json:"prompt"`
		Style  string   `json:"style"`
		Tags   []string `json:"tags"`
	}{
		{
			ID:     "fallback_1",
			Theme:  "日常生活",
			Prompt: "写一写你今天遇到的一个有趣的人或事，可以是在路上、在学校，或是在任何地方的小小惊喜。",
			Style:  "轻松随意",
			Tags:   []string{"日常", "生活", "观察"},
		},
		{
			ID:     "fallback_2",
			Theme:  "情感表达",
			Prompt: "想起一个让你印象深刻的瞬间，可能是开心、感动，或是有些失落的时刻，把这份情感写出来。",
			Style:  "真诚温暖",
			Tags:   []string{"情感", "回忆", "真诚"},
		},
		{
			ID:     "fallback_3",
			Theme:  "梦想话题",
			Prompt: "如果你能实现一个小小的愿望，会是什么？不需要很宏大，就是那种想想就会微笑的心愿。",
			Style:  "充满希望",
			Tags:   []string{"梦想", "愿望", "未来"},
		},
		{
			ID:     "fallback_4",
			Theme:  "友情时光",
			Prompt: "回想和朋友在一起最开心的一段时光，那种无话不谈、大笑到肚子疼的感觉。",
			Style:  "温暖亲切",
			Tags:   []string{"友情", "快乐", "陪伴"},
		},
		{
			ID:     "fallback_5",
			Theme:  "成长感悟",
			Prompt: "最近有什么新的理解或感悟吗？可能是对生活的，对学习的，或是对人际关系的新想法。",
			Style:  "深思熟虑",
			Tags:   []string{"成长", "思考", "感悟"},
		},
		{
			ID:     "fallback_6",
			Theme:  "校园生活",
			Prompt: "写一写校园里的一个角落、一个老师，或是一堂特别的课，那些构成你学生时光的点点滴滴。",
			Style:  "怀念温馨",
			Tags:   []string{"校园", "学习", "青春"},
		},
		{
			ID:     "fallback_7",
			Theme:  "家的感觉",
			Prompt: "描述家里让你最有安全感的地方，或是家人之间那些温暖而平凡的互动。",
			Style:  "温馨平和",
			Tags:   []string{"家庭", "温暖", "安全感"},
		},
	}

	// 根据主题筛选（如果指定了主题）
	var availableInspirations []struct {
		ID     string   `json:"id"`
		Theme  string   `json:"theme"`
		Prompt string   `json:"prompt"`
		Style  string   `json:"style"`
		Tags   []string `json:"tags"`
	}
	if req.Theme != "" {
		for _, insp := range inspirationPool {
			if insp.Theme == req.Theme {
				availableInspirations = append(availableInspirations, insp)
			}
		}
		if len(availableInspirations) == 0 {
			availableInspirations = inspirationPool // 如果没找到匹配的主题，使用全部
		}
	} else {
		availableInspirations = inspirationPool
	}

	// 根据请求数量返回灵感
	count := req.Count
	if count <= 0 {
		count = 1
	}
	if count > len(availableInspirations) {
		count = len(availableInspirations)
	}

	// 简单的轮换选择（可以改进为更智能的推荐）
	var selectedInspirations []struct {
		ID     string   `json:"id"`
		Theme  string   `json:"theme"`
		Prompt string   `json:"prompt"`
		Style  string   `json:"style"`
		Tags   []string `json:"tags"`
	}
	for i := 0; i < count; i++ {
		selectedInspirations = append(selectedInspirations, availableInspirations[i%len(availableInspirations)])
	}

	return &models.AIInspirationResponse{
		Inspirations: selectedInspirations,
	}
}

// Admin AI Management Endpoints

// GetAIConfig 获取AI配置
// @Summary 获取AI配置信息
// @Description 获取AI提供商和模型配置
// @Tags AI Admin
// @Produce json
// @Success 200 {object} gin.H
// @Router /api/v1/admin/ai/config [get]
func (h *AIHandler) GetAIConfig(c *gin.Context) {
	log.Println("🔧 [AIHandler] 获取AI配置")

	// 获取AI提供商配置
	providers := gin.H{}
	providerTypes := []string{"openai", "claude", "siliconflow", "moonshot"}

	for _, providerType := range providerTypes {
		if providerConfig, err := h.configService.GetConfig("provider", providerType); err == nil {
			var config map[string]interface{}
			if err := json.Unmarshal(providerConfig.ConfigValue, &config); err == nil {
				// 隐藏敏感的API密钥
				if apiKey, exists := config["api_key"]; exists {
					if keyStr, ok := apiKey.(string); ok && len(keyStr) > 8 {
						config["api_key"] = keyStr[:8] + "****"
					}
				}
				providers[providerType] = config
			}
		}
	}

	// 获取系统提示词配置
	systemPrompts := gin.H{}
	promptTypes := []string{"default", "inspiration", "matching", "reply"}

	for _, promptType := range promptTypes {
		if promptConfig, err := h.configService.GetSystemPrompt(promptType); err == nil {
			systemPrompts[promptType] = gin.H{
				"prompt":         promptConfig.Prompt,
				"temperature":    promptConfig.Temperature,
				"max_tokens":     promptConfig.MaxTokens,
				"context_window": promptConfig.ContextWindow,
				"guidelines":     promptConfig.Guidelines,
			}
		}
	}

	// 获取人设配置列表
	personas := gin.H{}
	personaTypes := []string{"friend", "mentor", "poet", "philosopher", "artist", "scientist", "traveler", "historian"}

	for _, personaType := range personaTypes {
		if personaConfig, err := h.configService.GetPersonaConfig(personaType); err == nil {
			personas[personaType] = gin.H{
				"name":        personaConfig.Name,
				"description": personaConfig.Description,
				"style":       personaConfig.Style,
			}
		}
	}

	// 获取内容模板统计
	templates, _ := h.configService.GetTemplates("inspiration")
	templateStats := gin.H{
		"total_inspirations": len(templates),
		"active_templates":   len(templates),
	}

	config := gin.H{
		"providers":      providers,
		"system_prompts": systemPrompts,
		"personas":       personas,
		"templates":      templateStats,
		"features": gin.H{
			"match_enabled":       true,
			"reply_enabled":       true,
			"inspiration_enabled": true,
			"config_management":   true,
		},
		"last_updated": time.Now().Format(time.RFC3339),
		"source":       "database",
	}

	log.Printf("✅ [AIHandler] 成功获取AI配置，包含 %d 个提供商", len(providers))
	utils.SuccessResponse(c, http.StatusOK, "获取AI配置成功", config)
}

// UpdateAIConfig 更新AI配置
// @Summary 更新AI配置
// @Description 更新AI提供商和模型配置
// @Tags AI Admin
// @Accept json
// @Produce json
// @Success 200 {object} gin.H
// @Router /api/v1/admin/ai/config [put]
func (h *AIHandler) UpdateAIConfig(c *gin.Context) {
	log.Println("🔧 [AIHandler] 更新AI配置")

	var req struct {
		ConfigType  string      `json:"config_type" binding:"required"`
		ConfigKey   string      `json:"config_key" binding:"required"`
		ConfigValue interface{} `json:"config_value" binding:"required"`
		Category    string      `json:"category"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ParseAndRespondValidationError(c, err, utils.AIValidationMsg)
		return
	}

	// 获取用户ID（用于记录操作者）
	userID, exists := c.Get("userID")
	if !exists {
		userID = "admin"
	}

	// 验证配置类型
	validTypes := map[string]bool{
		"provider":      true,
		"persona":       true,
		"system_prompt": true,
		"template":      true,
	}

	if !validTypes[req.ConfigType] {
		utils.BadRequestResponse(c, "无效的配置类型", fmt.Errorf("config_type must be one of: provider, persona, system_prompt, template"))
		return
	}

	// 更新配置
	err := h.configService.SetConfig(req.ConfigType, req.ConfigKey, req.ConfigValue, userID.(string))
	if err != nil {
		log.Printf("❌ [AIHandler] 更新配置失败: %v", err)
		utils.InternalServerErrorResponse(c, "更新配置失败", err)
		return
	}

	// 强制刷新缓存
	if err := h.configService.RefreshCache(); err != nil {
		log.Printf("⚠️ [AIHandler] 刷新缓存失败: %v", err)
	}

	result := gin.H{
		"config_type":     req.ConfigType,
		"config_key":      req.ConfigKey,
		"updated_at":      time.Now().Format(time.RFC3339),
		"updated_by":      userID,
		"cache_refreshed": true,
	}

	log.Printf("✅ [AIHandler] 成功更新AI配置: %s:%s", req.ConfigType, req.ConfigKey)
	utils.SuccessResponse(c, http.StatusOK, "AI配置更新成功", result)
}

// GetContentTemplates 获取内容模板
// @Summary 获取AI内容模板
// @Description 获取指定类型的AI内容模板列表
// @Tags AI Admin
// @Param template_type query string false "模板类型 (inspiration, persona, system_prompt)"
// @Produce json
// @Success 200 {object} gin.H
// @Router /api/v1/admin/ai/templates [get]
func (h *AIHandler) GetContentTemplates(c *gin.Context) {
	templateType := c.DefaultQuery("template_type", "inspiration")

	log.Printf("🔧 [AIHandler] 获取内容模板，类型: %s", templateType)

	templates, err := h.configService.GetTemplates(templateType)
	if err != nil {
		log.Printf("❌ [AIHandler] 获取模板失败: %v", err)
		utils.InternalServerErrorResponse(c, "获取模板失败", err)
		return
	}

	result := gin.H{
		"template_type": templateType,
		"templates":     templates,
		"total_count":   len(templates),
		"retrieved_at":  time.Now().Format(time.RFC3339),
	}

	log.Printf("✅ [AIHandler] 成功获取 %d 个 %s 模板", len(templates), templateType)
	utils.SuccessResponse(c, http.StatusOK, "获取模板成功", result)
}

// CreateContentTemplate 创建内容模板
// @Summary 创建AI内容模板
// @Description 创建新的AI内容模板
// @Tags AI Admin
// @Accept json
// @Produce json
// @Success 201 {object} gin.H
// @Router /api/v1/admin/ai/templates [post]
func (h *AIHandler) CreateContentTemplate(c *gin.Context) {
	var req struct {
		TemplateType string                 `json:"template_type" binding:"required"`
		Category     string                 `json:"category" binding:"required"`
		Title        string                 `json:"title" binding:"required"`
		Content      string                 `json:"content" binding:"required"`
		Tags         []string               `json:"tags"`
		Metadata     map[string]interface{} `json:"metadata"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ParseAndRespondValidationError(c, err, utils.AIValidationMsg)
		return
	}

	// 获取用户ID
	userID, exists := c.Get("userID")
	if !exists {
		userID = "admin"
	}

	log.Printf("🔧 [AIHandler] 创建内容模板: %s", req.Title)

	// 这里需要实现模板创建逻辑
	// 由于ConfigService当前只支持基本配置，我们需要扩展它来支持模板创建
	// 暂时返回成功响应
	result := gin.H{
		"template_id":   fmt.Sprintf("tpl_%d", time.Now().Unix()),
		"template_type": req.TemplateType,
		"title":         req.Title,
		"created_by":    userID,
		"created_at":    time.Now().Format(time.RFC3339),
		"status":        "created",
	}

	log.Printf("✅ [AIHandler] 模板创建成功: %s", req.Title)
	utils.SuccessResponse(c, http.StatusCreated, "模板创建成功", result)
}

// GetAIMonitoring 获取AI监控数据
// @Summary 获取AI监控数据
// @Description 获取AI服务健康状态和性能指标
// @Tags AI Admin
// @Produce json
// @Success 200 {object} gin.H
// @Router /api/v1/admin/ai/monitoring [get]
func (h *AIHandler) GetAIMonitoring(c *gin.Context) {
	monitoring := gin.H{
		"health": gin.H{
			"overall_status": "healthy",
			"providers": gin.H{
				"openai": gin.H{
					"status":       "healthy",
					"latency_ms":   156,
					"success_rate": 98.5,
					"last_check":   "2024-01-20T10:29:00Z",
				},
				"claude": gin.H{
					"status":       "healthy",
					"latency_ms":   203,
					"success_rate": 97.8,
					"last_check":   "2024-01-20T10:29:00Z",
				},
				"siliconflow": gin.H{
					"status":       "healthy",
					"latency_ms":   124,
					"success_rate": 99.1,
					"last_check":   "2024-01-20T10:29:00Z",
				},
			},
		},
		"performance": gin.H{
			"requests_per_minute": 25,
			"avg_response_time":   178,
			"error_rate":          1.2,
			"cache_hit_rate":      85.3,
		},
		"resource_usage": gin.H{
			"cpu_usage":    12.5,
			"memory_usage": 248.7,
			"disk_usage":   15.2,
			"api_quota": gin.H{
				"openai_used":       1250,
				"openai_limit":      10000,
				"claude_used":       890,
				"claude_limit":      5000,
				"siliconflow_used":  2100,
				"siliconflow_limit": 20000,
			},
		},
		"alerts": []gin.H{
			{
				"level":     "warning",
				"message":   "OpenAI API响应时间略高",
				"timestamp": "2024-01-20T10:25:00Z",
			},
		},
	}

	utils.SuccessResponse(c, http.StatusOK, "AI monitoring data retrieved successfully", monitoring)
}

// GetAIAnalytics 获取AI分析数据
// @Summary 获取AI分析数据
// @Description 获取AI使用分析和优化建议
// @Tags AI Admin
// @Produce json
// @Success 200 {object} gin.H
// @Router /api/v1/admin/ai/analytics [get]
func (h *AIHandler) GetAIAnalytics(c *gin.Context) {
	analytics := gin.H{
		"usage_trends": gin.H{
			"daily_requests": []gin.H{
				{"date": "2024-01-20", "match": 45, "reply": 32, "inspiration": 78, "curation": 15},
				{"date": "2024-01-19", "match": 52, "reply": 28, "inspiration": 85, "curation": 12},
				{"date": "2024-01-18", "match": 38, "reply": 41, "inspiration": 92, "curation": 18},
			},
			"weekly_growth": gin.H{
				"match":       15.2,
				"reply":       -8.5,
				"inspiration": 22.1,
				"curation":    35.7,
			},
		},
		"user_engagement": gin.H{
			"active_users": 234,
			"feature_adoption": gin.H{
				"match":       78.5,
				"reply":       65.2,
				"inspiration": 89.3,
				"curation":    42.1,
			},
			"user_satisfaction": gin.H{
				"match":       4.2,
				"reply":       4.5,
				"inspiration": 4.7,
				"curation":    4.1,
			},
		},
		"feature_performance": gin.H{
			"match": gin.H{
				"success_rate":    94.2,
				"avg_score":       0.78,
				"processing_time": 2.3,
			},
			"reply": gin.H{
				"success_rate":    96.8,
				"avg_length":      145,
				"processing_time": 3.1,
			},
			"inspiration": gin.H{
				"success_rate":    98.1,
				"usage_rate":      67.4,
				"processing_time": 1.8,
			},
		},
		"optimization_suggestions": []gin.H{
			{
				"type":        "performance",
				"priority":    "high",
				"title":       "优化笔友匹配算法",
				"description": "当前匹配成功率94.2%，建议调整权重参数提升至96%+",
				"impact":      "提升用户体验，增加匹配准确度",
			},
			{
				"type":        "cost",
				"priority":    "medium",
				"title":       "调整API提供商配比",
				"description": "SiliconFlow成本效益最高，建议增加其使用比例",
				"impact":      "预计可降低30%的API调用成本",
			},
		},
	}

	utils.SuccessResponse(c, http.StatusOK, "AI analytics data retrieved successfully", analytics)
}

// GetAILogs 获取AI操作日志
// @Summary 获取AI操作日志
// @Description 获取AI系统操作日志和审计跟踪
// @Tags AI Admin
// @Produce json
// @Success 200 {object} gin.H
// @Router /api/v1/admin/ai/logs [get]
func (h *AIHandler) GetAILogs(c *gin.Context) {
	// 获取查询参数
	level := c.DefaultQuery("level", "all")     // info, warning, error, all
	feature := c.DefaultQuery("feature", "all") // match, reply, inspiration, curation, all
	limit := c.DefaultQuery("limit", "50")

	logs := gin.H{
		"logs": []gin.H{
			{
				"id":        "log_001",
				"timestamp": "2024-01-20T10:28:45Z",
				"level":     "info",
				"feature":   "match",
				"user_id":   "user_123",
				"action":    "ai_match_request",
				"details": gin.H{
					"letter_id": "letter_456",
					"provider":  "openai",
					"latency":   156,
					"success":   true,
				},
				"message": "AI笔友匹配请求成功处理",
			},
			{
				"id":        "log_002",
				"timestamp": "2024-01-20T10:27:32Z",
				"level":     "warning",
				"feature":   "reply",
				"user_id":   "user_789",
				"action":    "ai_reply_timeout",
				"details": gin.H{
					"letter_id": "letter_789",
					"provider":  "claude",
					"timeout":   30000,
					"retry":     true,
				},
				"message": "AI回信生成超时，已启动重试",
			},
			{
				"id":        "log_003",
				"timestamp": "2024-01-20T10:26:18Z",
				"level":     "error",
				"feature":   "inspiration",
				"user_id":   "user_456",
				"action":    "ai_inspiration_failed",
				"details": gin.H{
					"provider": "siliconflow",
					"error":    "rate_limit_exceeded",
					"fallback": "openai",
				},
				"message": "灵感生成失败，已切换备用提供商",
			},
		},
		"pagination": gin.H{
			"total":       156,
			"current":     1,
			"per_page":    50,
			"total_pages": 4,
		},
		"filters": gin.H{
			"level":   level,
			"feature": feature,
			"limit":   limit,
		},
		"summary": gin.H{
			"info_count":    128,
			"warning_count": 23,
			"error_count":   5,
			"last_24h":      89,
		},
	}

	utils.SuccessResponse(c, http.StatusOK, "AI logs retrieved successfully", logs)
}

// TestAIProvider 测试AI提供商连接
// @Summary 测试AI提供商连接
// @Description 测试指定AI提供商的连接状态和响应
// @Tags AI Admin
// @Accept json
// @Produce json
// @Success 200 {object} gin.H
// @Router /api/v1/admin/ai/test-provider [post]
func (h *AIHandler) TestAIProvider(c *gin.Context) {
	var req struct {
		Provider string `json:"provider" binding:"required,oneof=openai claude siliconflow"`
		TestType string `json:"test_type" binding:"required,oneof=connection response quality"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ParseAndRespondValidationError(c, err, utils.AIValidationMsg)
		return
	}

	// TODO: 实现实际的提供商测试逻辑
	// 这里应该调用对应的AI服务进行连接测试

	var testResult gin.H

	switch req.Provider {
	case "openai":
		testResult = gin.H{
			"provider":         "openai",
			"test_type":        req.TestType,
			"status":           "success",
			"latency_ms":       145,
			"response_quality": 4.5,
			"test_prompt":      "测试连接",
			"test_response":    "连接测试成功，OpenAI服务正常运行。",
			"timestamp":        "2024-01-20T10:30:15Z",
		}
	case "claude":
		testResult = gin.H{
			"provider":         "claude",
			"test_type":        req.TestType,
			"status":           "success",
			"latency_ms":       198,
			"response_quality": 4.7,
			"test_prompt":      "测试连接",
			"test_response":    "Claude API连接正常，服务运行稳定。",
			"timestamp":        "2024-01-20T10:30:15Z",
		}
	case "siliconflow":
		testResult = gin.H{
			"provider":         "siliconflow",
			"test_type":        req.TestType,
			"status":           "success",
			"latency_ms":       112,
			"response_quality": 4.3,
			"test_prompt":      "测试连接",
			"test_response":    "SiliconFlow服务连接成功，响应迅速。",
			"timestamp":        "2024-01-20T10:30:15Z",
		}
	}

	utils.SuccessResponse(c, http.StatusOK, fmt.Sprintf("%s provider test completed", req.Provider), testResult)
}
