package handlers

import (
	"log"
	"net/http"
	"openpenpal-backend/internal/services"
	"openpenpal-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

// CloudLetterHandler 云中锦书处理器
type CloudLetterHandler struct {
	cloudLetterSvc *services.CloudLetterService
}

// NewCloudLetterHandler 创建云中锦书处理器
func NewCloudLetterHandler(cloudLetterSvc *services.CloudLetterService) *CloudLetterHandler {
	return &CloudLetterHandler{
		cloudLetterSvc: cloudLetterSvc,
	}
}

// CreatePersona 创建自定义人物角色
// @Summary 创建自定义人物角色
// @Description 用户创建一个真实世界的人物角色，用于云中锦书功能
// @Tags CloudLetter
// @Accept json
// @Produce json
// @Param request body services.PersonaCreateRequest true "创建人物请求"
// @Success 201 {object} services.CloudPersona
// @Router /api/v1/cloud-letters/personas [post]
func (h *CloudLetterHandler) CreatePersona(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		utils.InternalServerErrorResponse(c, "Invalid user ID format", nil)
		return
	}

	var req services.PersonaCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ParseAndRespondValidationError(c, err, utils.LetterValidationMsg)
		return
	}

	persona, err := h.cloudLetterSvc.CreatePersona(c.Request.Context(), userIDStr, &req)
	if err != nil {
		log.Printf("❌ [CloudLetterHandler] Failed to create persona: %v", err)
		utils.InternalServerErrorResponse(c, "Failed to create persona", err)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Persona created successfully", persona)
}

// GetPersonas 获取用户的人物角色列表
// @Summary 获取用户的人物角色列表
// @Description 获取当前用户创建的所有人物角色
// @Tags CloudLetter
// @Produce json
// @Success 200 {array} services.CloudPersona
// @Router /api/v1/cloud-letters/personas [get]
func (h *CloudLetterHandler) GetPersonas(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		utils.InternalServerErrorResponse(c, "Invalid user ID format", nil)
		return
	}

	personas, err := h.cloudLetterSvc.GetUserPersonas(c.Request.Context(), userIDStr)
	if err != nil {
		log.Printf("❌ [CloudLetterHandler] Failed to get personas: %v", err)
		utils.InternalServerErrorResponse(c, "Failed to get personas", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Personas retrieved successfully", gin.H{
		"personas": personas,
		"count":    len(personas),
	})
}

// UpdatePersona 更新人物角色
// @Summary 更新人物角色
// @Description 更新用户创建的人物角色信息
// @Tags CloudLetter
// @Accept json
// @Produce json
// @Param persona_id path string true "人物角色ID"
// @Param request body gin.H true "更新内容"
// @Success 200 {object} gin.H
// @Router /api/v1/cloud-letters/personas/{persona_id} [put]
func (h *CloudLetterHandler) UpdatePersona(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		utils.InternalServerErrorResponse(c, "Invalid user ID format", nil)
		return
	}

	personaID := c.Param("persona_id")
	if personaID == "" {
		utils.BadRequestResponse(c, "Persona ID is required", nil)
		return
	}

	var updates gin.H
	if err := c.ShouldBindJSON(&updates); err != nil {
		utils.ParseAndRespondValidationError(c, err, utils.LetterValidationMsg)
		return
	}

	// 转换为map[string]interface{}
	updateMap := make(map[string]interface{})
	for k, v := range updates {
		updateMap[k] = v
	}

	err := h.cloudLetterSvc.UpdatePersona(c.Request.Context(), userIDStr, personaID, updateMap)
	if err != nil {
		log.Printf("❌ [CloudLetterHandler] Failed to update persona: %v", err)
		utils.InternalServerErrorResponse(c, "Failed to update persona", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Persona updated successfully", nil)
}

// CreateCloudLetter 创建云信件
// @Summary 创建云信件
// @Description 用户向指定的人物角色写信
// @Tags CloudLetter
// @Accept json
// @Produce json
// @Param request body services.CloudLetterCreateRequest true "创建云信件请求"
// @Success 201 {object} services.CloudLetter
// @Router /api/v1/cloud-letters [post]
func (h *CloudLetterHandler) CreateCloudLetter(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		utils.InternalServerErrorResponse(c, "Invalid user ID format", nil)
		return
	}

	var req services.CloudLetterCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ParseAndRespondValidationError(c, err, utils.LetterValidationMsg)
		return
	}

	cloudLetter, err := h.cloudLetterSvc.CreateCloudLetter(c.Request.Context(), userIDStr, &req)
	if err != nil {
		log.Printf("❌ [CloudLetterHandler] Failed to create cloud letter: %v", err)
		utils.InternalServerErrorResponse(c, "Failed to create cloud letter", err)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Cloud letter created successfully", cloudLetter)
}

// GetCloudLetter 获取云信件详情
// @Summary 获取云信件详情
// @Description 获取指定云信件的详细信息，包括关联的人物角色
// @Tags CloudLetter
// @Produce json
// @Param letter_id path string true "信件ID"
// @Success 200 {object} gin.H
// @Router /api/v1/cloud-letters/{letter_id} [get]
func (h *CloudLetterHandler) GetCloudLetter(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		utils.InternalServerErrorResponse(c, "Invalid user ID format", nil)
		return
	}

	letterID := c.Param("letter_id")
	if letterID == "" {
		utils.BadRequestResponse(c, "Letter ID is required", nil)
		return
	}

	letter, persona, err := h.cloudLetterSvc.GetCloudLetter(c.Request.Context(), userIDStr, letterID)
	if err != nil {
		log.Printf("❌ [CloudLetterHandler] Failed to get cloud letter: %v", err)
		utils.InternalServerErrorResponse(c, "Failed to get cloud letter", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Cloud letter retrieved successfully", gin.H{
		"letter":  letter,
		"persona": persona,
	})
}

// GetCloudLetters 获取用户的云信件列表
// @Summary 获取用户的云信件列表
// @Description 获取当前用户的所有云信件，可按状态筛选
// @Tags CloudLetter
// @Produce json
// @Param status query string false "信件状态筛选"
// @Success 200 {array} services.CloudLetter
// @Router /api/v1/cloud-letters [get]
func (h *CloudLetterHandler) GetCloudLetters(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		utils.InternalServerErrorResponse(c, "Invalid user ID format", nil)
		return
	}

	status := c.Query("status")
	var statusFilter services.CloudLetterStatus
	if status != "" {
		statusFilter = services.CloudLetterStatus(status)
	}

	letters, err := h.cloudLetterSvc.GetUserCloudLetters(c.Request.Context(), userIDStr, statusFilter)
	if err != nil {
		log.Printf("❌ [CloudLetterHandler] Failed to get cloud letters: %v", err)
		utils.InternalServerErrorResponse(c, "Failed to get cloud letters", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Cloud letters retrieved successfully", gin.H{
		"letters": letters,
		"count":   len(letters),
	})
}

// GetPersonaTypes 获取支持的人物关系类型
// @Summary 获取支持的人物关系类型
// @Description 获取云中锦书支持的所有人物关系类型及其描述
// @Tags CloudLetter
// @Produce json
// @Success 200 {array} gin.H
// @Router /api/v1/cloud-letters/persona-types [get]
func (h *CloudLetterHandler) GetPersonaTypes(c *gin.Context) {
	personaTypes := []gin.H{
		{
			"type":        "deceased",
			"name":        "已故亲友",
			"description": "向已经离世的亲人或朋友写信，表达思念和感情",
			"icon":        "💐",
			"emotional_tone": "深情怀念",
		},
		{
			"type":        "distant_friend",
			"name":        "疏远朋友",
			"description": "向多年未见或失去联系的朋友写信，重温友谊",
			"icon":        "🤝",
			"emotional_tone": "温暖友谊",
		},
		{
			"type":        "unspoken_love",
			"name":        "未说出口的爱",
			"description": "向暗恋或未曾表白的人写信，表达内心情感",
			"icon":        "💌",
			"emotional_tone": "含蓄深情",
		},
		{
			"type":        "custom",
			"name":        "自定义关系",
			"description": "创建特殊的人物关系，如老师、导师、偶像等",
			"icon":        "✨",
			"emotional_tone": "个性化",
		},
	}

	utils.SuccessResponse(c, http.StatusOK, "Persona types retrieved successfully", gin.H{
		"types": personaTypes,
		"count": len(personaTypes),
		"feature_description": "云中锦书 - 向真实世界中的特殊人物写信，让AI帮助你表达内心深处的情感",
	})
}

// GetLetterStatusOptions 获取信件状态选项
// @Summary 获取信件状态选项
// @Description 获取云信件的所有可能状态及其描述
// @Tags CloudLetter
// @Produce json
// @Success 200 {array} gin.H
// @Router /api/v1/cloud-letters/status-options [get]
func (h *CloudLetterHandler) GetLetterStatusOptions(c *gin.Context) {
	statusOptions := []gin.H{
		{
			"status":      "draft",
			"name":        "草稿",
			"description": "信件正在编写中",
			"color":       "#64748B",
		},
		{
			"status":      "ai_enhanced",
			"name":        "AI增强完成",
			"description": "AI已完成内容优化和增强",
			"color":       "#3B82F6",
		},
		{
			"status":      "under_review",
			"name":        "审核中",
			"description": "信件正在等待高级信使审核",
			"color":       "#F59E0B",
		},
		{
			"status":      "revision_needed",
			"name":        "需要修改",
			"description": "审核员建议修改后重新提交",
			"color":       "#EF4444",
		},
		{
			"status":      "approved",
			"name":        "已批准",
			"description": "信件已通过审核，准备投递",
			"color":       "#10B981",
		},
		{
			"status":      "delivered",
			"name":        "已投递",
			"description": "信件已成功投递",
			"color":       "#8B5CF6",
		},
		{
			"status":      "replied",
			"name":        "已回信",
			"description": "已收到AI生成的回信",
			"color":       "#EC4899",
		},
	}

	utils.SuccessResponse(c, http.StatusOK, "Letter status options retrieved successfully", gin.H{
		"statuses": statusOptions,
		"count":    len(statusOptions),
	})
}