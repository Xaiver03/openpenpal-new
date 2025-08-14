package handlers

import (
	"net/http"
	"openpenpal-backend/internal/services"
	"openpenpal-backend/internal/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ContentSecurityHandler 内容安全处理器
type ContentSecurityHandler struct {
	securityService *services.ContentSecurityService
}

// NewContentSecurityHandler 创建内容安全处理器
func NewContentSecurityHandler(securityService *services.ContentSecurityService) *ContentSecurityHandler {
	return &ContentSecurityHandler{
		securityService: securityService,
	}
}

// CheckContent 内容安全检查
// @Summary 内容安全检查
// @Description 检查文本内容的安全性
// @Tags ContentSecurity
// @Accept json
// @Produce json
// @Param request body object{content=string,content_type=string} true "检查请求"
// @Success 200 {object} services.SecurityCheckResult
// @Router /api/v1/security/check [post]
func (h *ContentSecurityHandler) CheckContent(c *gin.Context) {
	var req struct {
		Content     string `json:"content" binding:"required"`
		ContentType string `json:"content_type"`
		ContentID   string `json:"content_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	// 从JWT中获取用户ID
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

	// 设置默认值
	if req.ContentType == "" {
		req.ContentType = "text"
	}

	// 执行安全检查
	result, err := h.securityService.CheckContent(c.Request.Context(), userIDStr, req.ContentType, req.ContentID, req.Content)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to check content security", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Content security check completed", result)
}

// GetUserViolations 获取用户违规记录
// @Summary 获取用户违规记录
// @Description 获取当前用户的内容违规历史
// @Tags ContentSecurity
// @Produce json
// @Param limit query int false "限制数量"
// @Success 200 {array} services.ContentViolationRecord
// @Router /api/v1/security/violations [get]
func (h *ContentSecurityHandler) GetUserViolations(c *gin.Context) {
	// 从JWT中获取用户ID
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

	// 获取limit参数
	limit := 20 // 默认20条
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	// 获取违规记录
	violations, err := h.securityService.GetUserViolationHistory(userIDStr, limit)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get violation history", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Violation history retrieved successfully", violations)
}

// GetPendingReviews 获取待审核的违规内容（管理员功能）
// @Summary 获取待审核的违规内容
// @Description 管理员获取待审核的内容违规记录
// @Tags ContentSecurity
// @Produce json
// @Param limit query int false "限制数量"
// @Success 200 {array} services.ContentViolationRecord
// @Router /api/v1/admin/security/pending [get]
func (h *ContentSecurityHandler) GetPendingReviews(c *gin.Context) {
	// 检查管理员权限
	// 这里应该有权限检查逻辑

	// 获取limit参数
	limit := 50 // 默认50条
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	// 获取待审核记录
	reviews, err := h.securityService.GetPendingReviews(limit)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to get pending reviews", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Pending reviews retrieved successfully", reviews)
}

// ReviewViolation 审核违规内容（管理员功能）
// @Summary 审核违规内容
// @Description 管理员审核内容违规记录
// @Tags ContentSecurity
// @Accept json
// @Produce json
// @Param violation_id path string true "违规记录ID"
// @Param request body object{action=string} true "审核请求"
// @Success 200 {object} object{message=string}
// @Router /api/v1/admin/security/review/{violation_id} [put]
func (h *ContentSecurityHandler) ReviewViolation(c *gin.Context) {
	violationID := c.Param("violation_id")
	if violationID == "" {
		utils.BadRequestResponse(c, "Violation ID is required", nil)
		return
	}

	var req struct {
		Action string `json:"action" binding:"required"` // approved, blocked, dismissed
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request format", err)
		return
	}

	// 验证action值
	validActions := map[string]bool{
		"approved":  true,
		"blocked":   true,
		"dismissed": true,
	}
	if !validActions[req.Action] {
		utils.BadRequestResponse(c, "Invalid action. Must be one of: approved, blocked, dismissed", nil)
		return
	}

	// 从JWT中获取审核员ID
	reviewerID, exists := c.Get("user_id")
	if !exists {
		utils.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	reviewerIDStr, ok := reviewerID.(string)
	if !ok {
		utils.InternalServerErrorResponse(c, "Invalid user ID format", nil)
		return
	}

	// 执行审核
	err := h.securityService.ReviewViolation(violationID, reviewerIDStr, req.Action)
	if err != nil {
		utils.InternalServerErrorResponse(c, "Failed to review violation", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Violation reviewed successfully", gin.H{
		"violation_id": violationID,
		"action":       req.Action,
		"reviewed_by":  reviewerIDStr,
	})
}
