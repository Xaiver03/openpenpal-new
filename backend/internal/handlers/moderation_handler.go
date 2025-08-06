package handlers

import (
	"fmt"
	"net/http"
	"openpenpal-backend/internal/middleware"
	"openpenpal-backend/internal/models"
	"openpenpal-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// ModerationHandler 审核处理器
type ModerationHandler struct {
	moderationService *services.ModerationService
}

// NewModerationHandler 创建审核处理器
func NewModerationHandler(moderationService *services.ModerationService) *ModerationHandler {
	return &ModerationHandler{
		moderationService: moderationService,
	}
}

// ModerateContent 审核内容
// @Summary 审核内容
// @Description 对内容进行自动审核
// @Tags Moderation
// @Accept json
// @Produce json
// @Param request body models.ModerationRequest true "审核请求"
// @Success 200 {object} models.ModerationResponse
// @Router /api/v1/moderation/check [post]
func (h *ModerationHandler) ModerateContent(c *gin.Context) {
	var req models.ModerationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// 如果没有提供用户ID，从JWT中获取
	if req.UserID == "" {
		userID, exists := middleware.GetUserID(c)
		if exists {
			req.UserID = userID
		}
	}

	// 调用审核服务
	response, err := h.moderationService.ModerateContent(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to moderate content: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// ReviewContent 人工审核内容
// @Summary 人工审核内容
// @Description 管理员对内容进行人工审核
// @Tags Moderation
// @Accept json
// @Produce json
// @Param request body models.ReviewRequest true "审核请求"
// @Success 200 {object} gin.H
// @Router /api/v1/admin/moderation/review [post]
func (h *ModerationHandler) ReviewContent(c *gin.Context) {
	var req models.ReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// 获取审核员ID
	reviewerID, _ := middleware.GetUserID(c)

	// 调用审核服务
	err := h.moderationService.ReviewContent(c.Request.Context(), &req, reviewerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to review content: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Content reviewed successfully",
		"status":  req.Status,
	})
}

// GetModerationQueue 获取待审核队列
// @Summary 获取待审核队列
// @Description 获取需要人工审核的内容队列
// @Tags Moderation
// @Produce json
// @Param limit query int false "限制数量" default(20)
// @Success 200 {object} gin.H
// @Router /api/v1/admin/moderation/queue [get]
func (h *ModerationHandler) GetModerationQueue(c *gin.Context) {
	limit := 20
	if l := c.Query("limit"); l != "" {
		if parsedLimit, err := parseInt(l); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	queue, err := h.moderationService.GetModerationQueue(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get moderation queue: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"queue": queue,
		"total": len(queue),
	})
}

// GetModerationStats 获取审核统计
// @Summary 获取审核统计
// @Description 获取指定时间范围内的审核统计数据
// @Tags Moderation
// @Produce json
// @Param start_date query string false "开始日期" format(date)
// @Param end_date query string false "结束日期" format(date)
// @Success 200 {object} gin.H
// @Router /api/v1/admin/moderation/stats [get]
func (h *ModerationHandler) GetModerationStats(c *gin.Context) {
	// TODO: 解析日期参数并调用服务
	c.JSON(http.StatusOK, gin.H{
		"stats":   []models.ModerationStats{},
		"message": "Statistics feature coming soon",
	})
}

// AddSensitiveWord 添加敏感词
// @Summary 添加敏感词
// @Description 添加新的敏感词到词库
// @Tags Moderation
// @Accept json
// @Produce json
// @Param request body models.SensitiveWordRequest true "敏感词请求"
// @Success 200 {object} gin.H
// @Router /api/v1/admin/moderation/sensitive-words [post]
func (h *ModerationHandler) AddSensitiveWord(c *gin.Context) {
	var req models.SensitiveWordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// 获取创建者ID
	createdBy, _ := middleware.GetUserID(c)

	// 调用服务
	err := h.moderationService.AddSensitiveWord(&req, createdBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add sensitive word: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Sensitive word added successfully",
		"word":    req.Word,
	})
}

// UpdateSensitiveWord 更新敏感词
// @Summary 更新敏感词
// @Description 更新已有的敏感词
// @Tags Moderation
// @Accept json
// @Produce json
// @Param id path string true "敏感词ID"
// @Param request body models.SensitiveWordRequest true "敏感词请求"
// @Success 200 {object} gin.H
// @Router /api/v1/admin/moderation/sensitive-words/:id [put]
func (h *ModerationHandler) UpdateSensitiveWord(c *gin.Context) {
	id := c.Param("id")

	var req models.SensitiveWordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// 调用服务
	err := h.moderationService.UpdateSensitiveWord(id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update sensitive word: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Sensitive word updated successfully",
		"id":      id,
	})
}

// DeleteSensitiveWord 删除敏感词
// @Summary 删除敏感词
// @Description 从词库中删除敏感词
// @Tags Moderation
// @Produce json
// @Param id path string true "敏感词ID"
// @Success 200 {object} gin.H
// @Router /api/v1/admin/moderation/sensitive-words/:id [delete]
func (h *ModerationHandler) DeleteSensitiveWord(c *gin.Context) {
	id := c.Param("id")

	// 调用服务
	err := h.moderationService.DeleteSensitiveWord(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete sensitive word: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Sensitive word deleted successfully",
		"id":      id,
	})
}

// GetSensitiveWords 获取敏感词列表
// @Summary 获取敏感词列表
// @Description 获取敏感词库中的词汇列表
// @Tags Moderation
// @Produce json
// @Param category query string false "分类"
// @Param level query string false "等级"
// @Success 200 {object} gin.H
// @Router /api/v1/admin/moderation/sensitive-words [get]
func (h *ModerationHandler) GetSensitiveWords(c *gin.Context) {
	category := c.Query("category")
	level := models.ModerationLevel(c.Query("level"))

	words, err := h.moderationService.GetSensitiveWords(category, level)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get sensitive words: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"words": words,
		"total": len(words),
	})
}

// AddModerationRule 添加审核规则
// @Summary 添加审核规则
// @Description 添加新的内容审核规则
// @Tags Moderation
// @Accept json
// @Produce json
// @Param request body models.ModerationRuleRequest true "规则请求"
// @Success 200 {object} gin.H
// @Router /api/v1/admin/moderation/rules [post]
func (h *ModerationHandler) AddModerationRule(c *gin.Context) {
	var req models.ModerationRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// 获取创建者ID
	createdBy, _ := middleware.GetUserID(c)

	// 调用服务
	err := h.moderationService.AddModerationRule(&req, createdBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add moderation rule: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Moderation rule added successfully",
		"rule":    req.Name,
	})
}

// UpdateModerationRule 更新审核规则
// @Summary 更新审核规则
// @Description 更新已有的审核规则
// @Tags Moderation
// @Accept json
// @Produce json
// @Param id path string true "规则ID"
// @Param request body models.ModerationRuleRequest true "规则请求"
// @Success 200 {object} gin.H
// @Router /api/v1/admin/moderation/rules/:id [put]
func (h *ModerationHandler) UpdateModerationRule(c *gin.Context) {
	id := c.Param("id")

	var req models.ModerationRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// 调用服务
	err := h.moderationService.UpdateModerationRule(id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update moderation rule: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Moderation rule updated successfully",
		"id":      id,
	})
}

// DeleteModerationRule 删除审核规则
// @Summary 删除审核规则
// @Description 删除审核规则
// @Tags Moderation
// @Produce json
// @Param id path string true "规则ID"
// @Success 200 {object} gin.H
// @Router /api/v1/admin/moderation/rules/:id [delete]
func (h *ModerationHandler) DeleteModerationRule(c *gin.Context) {
	id := c.Param("id")

	// 调用服务
	err := h.moderationService.DeleteModerationRule(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete moderation rule: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Moderation rule deleted successfully",
		"id":      id,
	})
}

// GetModerationRules 获取审核规则列表
// @Summary 获取审核规则列表
// @Description 获取内容审核规则列表
// @Tags Moderation
// @Produce json
// @Param content_type query string false "内容类型"
// @Success 200 {object} gin.H
// @Router /api/v1/admin/moderation/rules [get]
func (h *ModerationHandler) GetModerationRules(c *gin.Context) {
	contentType := models.ContentType(c.Query("content_type"))

	rules, err := h.moderationService.GetModerationRules(contentType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get moderation rules: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"rules": rules,
		"total": len(rules),
	})
}

// 辅助函数

// parseInt 解析整数
func parseInt(s string) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}
