package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"openpenpal-backend/internal/models"
	"openpenpal-backend/internal/services"
)

// CreditLimitHandler 积分限制处理器
type CreditLimitHandler struct {
	limiterService *services.CreditLimiterService
}

// NewCreditLimitHandler 创建积分限制处理器
func NewCreditLimitHandler(limiterService *services.CreditLimiterService) *CreditLimitHandler {
	return &CreditLimitHandler{
		limiterService: limiterService,
	}
}

// GetLimitRules 获取限制规则列表
func (h *CreditLimitHandler) GetLimitRules(c *gin.Context) {
	var rules []models.CreditLimitRule
	
	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit
	
	// 筛选参数
	actionType := c.Query("action_type")
	enabled := c.Query("enabled")
	
	query := h.limiterService.DB().Model(&models.CreditLimitRule{})
	
	if actionType != "" {
		query = query.Where("action_type = ?", actionType)
	}
	if enabled != "" {
		query = query.Where("enabled = ?", enabled == "true")
	}
	
	// 获取总数
	var total int64
	query.Count(&total)
	
	// 获取列表
	err := query.Order("priority ASC, created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&rules).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get rules"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"rules": rules,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// CreateLimitRule 创建限制规则
func (h *CreditLimitHandler) CreateLimitRule(c *gin.Context) {
	// 检查管理员权限
	if !h.isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
		return
	}
	
	var req struct {
		ActionType  string                    `json:"action_type" binding:"required"`
		LimitType   models.CreditLimitType    `json:"limit_type" binding:"required"`
		LimitPeriod models.CreditLimitPeriod  `json:"limit_period" binding:"required"`
		MaxCount    int                       `json:"max_count" binding:"required"`
		MaxPoints   int                       `json:"max_points"`
		Enabled     bool                      `json:"enabled"`
		Priority    int                       `json:"priority"`
		Description string                    `json:"description"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	rule := &models.CreditLimitRule{
		ID:          uuid.New().String(),
		ActionType:  req.ActionType,
		LimitType:   req.LimitType,
		LimitPeriod: req.LimitPeriod,
		MaxCount:    req.MaxCount,
		MaxPoints:   req.MaxPoints,
		Enabled:     req.Enabled,
		Priority:    req.Priority,
		Description: req.Description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	if err := h.limiterService.DB().Create(rule).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create rule"})
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{"rule": rule})
}

// UpdateLimitRule 更新限制规则
func (h *CreditLimitHandler) UpdateLimitRule(c *gin.Context) {
	// 检查管理员权限
	if !h.isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
		return
	}
	
	ruleID := c.Param("id")
	
	var req struct {
		LimitType   models.CreditLimitType    `json:"limit_type"`
		LimitPeriod models.CreditLimitPeriod  `json:"limit_period"`
		MaxCount    int                       `json:"max_count"`
		MaxPoints   int                       `json:"max_points"`
		Enabled     *bool                     `json:"enabled"`
		Priority    int                       `json:"priority"`
		Description string                    `json:"description"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	var rule models.CreditLimitRule
	if err := h.limiterService.DB().Where("id = ?", ruleID).First(&rule).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Rule not found"})
		return
	}
	
	// 更新字段
	if req.LimitType != "" {
		rule.LimitType = req.LimitType
	}
	if req.LimitPeriod != "" {
		rule.LimitPeriod = req.LimitPeriod
	}
	if req.MaxCount > 0 {
		rule.MaxCount = req.MaxCount
	}
	if req.MaxPoints > 0 {
		rule.MaxPoints = req.MaxPoints
	}
	if req.Enabled != nil {
		rule.Enabled = *req.Enabled
	}
	if req.Priority > 0 {
		rule.Priority = req.Priority
	}
	if req.Description != "" {
		rule.Description = req.Description
	}
	rule.UpdatedAt = time.Now()
	
	if err := h.limiterService.DB().Save(&rule).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update rule"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"rule": rule})
}

// DeleteLimitRule 删除限制规则
func (h *CreditLimitHandler) DeleteLimitRule(c *gin.Context) {
	// 检查管理员权限
	if !h.isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
		return
	}
	
	ruleID := c.Param("id")
	
	if err := h.limiterService.DB().Where("id = ?", ruleID).Delete(&models.CreditLimitRule{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete rule"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Rule deleted successfully"})
}

// GetUserLimitStatus 获取用户限制状态
func (h *CreditLimitHandler) GetUserLimitStatus(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}
	
	actionType := c.Query("action_type")
	if actionType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "action_type is required"})
		return
	}
	
	status, err := h.limiterService.GetLimitStatus(userID, actionType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get limit status"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"status": status})
}

// GetUserRiskInfo 获取用户风险信息
func (h *CreditLimitHandler) GetUserRiskInfo(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未登录"})
		return
	}
	
	// 获取风险分数
	riskScore, err := h.limiterService.GetRiskScore(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get risk score"})
		return
	}
	
	// 检查是否被封禁
	isBlocked, err := h.limiterService.IsUserBlocked(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check block status"})
		return
	}
	
	// 获取风险用户详情
	var riskUser models.CreditRiskUser
	h.limiterService.DB().Where("user_id = ?", userID).First(&riskUser)
	
	c.JSON(http.StatusOK, gin.H{
		"risk_score": riskScore,
		"is_blocked": isBlocked,
		"risk_level": riskUser.RiskLevel,
		"blocked_until": riskUser.BlockedUntil,
		"reason": riskUser.Reason,
	})
}

// GetRiskUsers 获取风险用户列表（管理员）
func (h *CreditLimitHandler) GetRiskUsers(c *gin.Context) {
	// 检查管理员权限
	if !h.isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
		return
	}
	
	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit
	
	// 筛选参数
	riskLevel := c.Query("risk_level")
	
	query := h.limiterService.DB().Model(&models.CreditRiskUser{})
	
	if riskLevel != "" {
		query = query.Where("risk_level = ?", riskLevel)
	}
	
	// 获取总数
	var total int64
	query.Count(&total)
	
	// 获取列表
	var riskUsers []models.CreditRiskUser
	err := query.Order("risk_score DESC, updated_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&riskUsers).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get risk users"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"users": riskUsers,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// BlockUser 封禁用户（管理员）
func (h *CreditLimitHandler) BlockUser(c *gin.Context) {
	// 检查管理员权限
	if !h.isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
		return
	}
	
	var req struct {
		UserID   string `json:"user_id" binding:"required"`
		Reason   string `json:"reason" binding:"required"`
		Duration int    `json:"duration"` // 小时数，0表示永久
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	var duration time.Duration
	if req.Duration > 0 {
		duration = time.Hour * time.Duration(req.Duration)
	} else {
		duration = time.Hour * 24 * 365 * 10 // 10年，相当于永久
	}
	
	err := h.limiterService.BlockUser(req.UserID, req.Reason, duration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to block user"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "User blocked successfully"})
}

// UnblockUser 解封用户（管理员）
func (h *CreditLimitHandler) UnblockUser(c *gin.Context) {
	// 检查管理员权限
	if !h.isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
		return
	}
	
	userID := c.Param("user_id")
	
	var riskUser models.CreditRiskUser
	if err := h.limiterService.DB().Where("user_id = ?", userID).First(&riskUser).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	
	riskUser.RiskLevel = models.RiskLevelLow
	riskUser.RiskScore = 0.1
	riskUser.BlockedUntil = nil
	riskUser.Reason = ""
	riskUser.UpdatedAt = time.Now()
	
	if err := h.limiterService.DB().Save(&riskUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unblock user"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "User unblocked successfully"})
}

// GetUserActions 获取用户行为记录（管理员）
func (h *CreditLimitHandler) GetUserActions(c *gin.Context) {
	// 检查管理员权限
	if !h.isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
		return
	}
	
	userID := c.Param("user_id")
	
	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset := (page - 1) * limit
	
	// 时间筛选
	since := c.Query("since")
	actionType := c.Query("action_type")
	
	query := h.limiterService.DB().Model(&models.UserCreditAction{}).
		Where("user_id = ?", userID)
	
	if since != "" {
		if sinceTime, err := time.Parse("2006-01-02", since); err == nil {
			query = query.Where("created_at >= ?", sinceTime)
		}
	}
	
	if actionType != "" {
		query = query.Where("action_type = ?", actionType)
	}
	
	// 获取总数
	var total int64
	query.Count(&total)
	
	// 获取列表
	var actions []models.UserCreditAction
	err := query.Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&actions).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user actions"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"actions": actions,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}

// Private methods

// isAdmin 检查是否是管理员
func (h *CreditLimitHandler) isAdmin(c *gin.Context) bool {
	userRole := c.GetString("user_role")
	return userRole == "super_admin" || userRole == "admin"
}

