package handlers

import (
	"openpenpal-backend/internal/models"
	"openpenpal-backend/internal/services"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"openpenpal-backend/internal/pkg/response"
)

// CreditExpirationHandler 积分过期处理器
type CreditExpirationHandler struct {
	expirationService *services.CreditExpirationService
	creditService     *services.CreditService
}

// NewCreditExpirationHandler 创建积分过期处理器实例
func NewCreditExpirationHandler(expirationService *services.CreditExpirationService, creditService *services.CreditService) *CreditExpirationHandler {
	return &CreditExpirationHandler{
		expirationService: expirationService,
		creditService:     creditService,
	}
}

// ===================== 用户相关API =====================

// GetUserExpiringCredits 获取用户即将过期的积分
func (h *CreditExpirationHandler) GetUserExpiringCredits(c *gin.Context) {
	resp := response.NewGinResponse()
	user := c.MustGet("user").(*models.User)

	// 解析查询参数
	daysStr := c.DefaultQuery("days", "30") // 默认查询30天内过期的积分
	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 {
		resp.BadRequest(c, "无效的天数参数")
		return
	}

	// 限制最大查询天数
	if days > 365 {
		days = 365
	}

	transactions, totalCredits, err := h.expirationService.GetUserExpiringCredits(user.ID, days)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.Success(c, gin.H{
		"transactions":   transactions,
		"total_credits":  totalCredits,
		"days":          days,
		"count":         len(transactions),
	})
}

// GetUserExpirationHistory 获取用户积分过期历史
func (h *CreditExpirationHandler) GetUserExpirationHistory(c *gin.Context) {
	resp := response.NewGinResponse()
	user := c.MustGet("user").(*models.User)

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// 查询过期记录
	var logs []models.CreditExpirationLog
	var total int64

	// Note: For better architecture, we should add a method to the service instead of accessing db directly
	// This is a temporary implementation - should be refactored to use service methods

	// Get database instance from credit service
	db := h.creditService.GetDB()
	
	err := db.Model(&models.CreditExpirationLog{}).
		Where("user_id = ?", user.ID).
		Count(&total).Error
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	err = db.Where("user_id = ?", user.ID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&logs).Error
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.Success(c, gin.H{
		"logs":      logs,
		"total":     total,
		"page":      page,
		"page_size": limit,
		"has_next":  int64(page*limit) < total,
	})
}

// ===================== 管理员API =====================

// GetExpirationRules 获取过期规则列表（管理员）
func (h *CreditExpirationHandler) GetExpirationRules(c *gin.Context) {
	resp := response.NewGinResponse()

	// 验证管理员权限
	user := c.MustGet("user").(*models.User)
	if user.Role != "admin" && user.Role != "super_admin" {
		resp.Unauthorized(c, "只有管理员可以查看过期规则")
		return
	}

	rules, err := h.expirationService.GetExpirationRules()
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.Success(c, gin.H{
		"rules": rules,
		"count": len(rules),
	})
}

// CreateExpirationRule 创建过期规则（管理员）
func (h *CreditExpirationHandler) CreateExpirationRule(c *gin.Context) {
	resp := response.NewGinResponse()

	// 验证管理员权限
	user := c.MustGet("user").(*models.User)
	if user.Role != "admin" && user.Role != "super_admin" {
		resp.Unauthorized(c, "只有管理员可以创建过期规则")
		return
	}

	var rule models.CreditExpirationRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	// 验证必填字段
	if rule.RuleName == "" {
		resp.BadRequest(c, "规则名称不能为空")
		return
	}

	if rule.CreditType == "" {
		resp.BadRequest(c, "积分类型不能为空")
		return
	}

	if rule.ExpirationDays <= 0 {
		resp.BadRequest(c, "过期天数必须大于0")
		return
	}

	// 设置创建信息
	rule.CreatedBy = user.ID
	rule.UpdatedBy = user.ID

	if err := h.expirationService.CreateExpirationRule(&rule); err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.CreatedWithMessage(c, "过期规则创建成功", rule)
}

// UpdateExpirationRule 更新过期规则（管理员）
func (h *CreditExpirationHandler) UpdateExpirationRule(c *gin.Context) {
	resp := response.NewGinResponse()

	// 验证管理员权限
	user := c.MustGet("user").(*models.User)
	if user.Role != "admin" && user.Role != "super_admin" {
		resp.Unauthorized(c, "只有管理员可以更新过期规则")
		return
	}

	ruleID := c.Param("id")
	if _, err := uuid.Parse(ruleID); err != nil {
		resp.BadRequest(c, "无效的规则ID")
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	// 设置更新人
	updates["updated_by"] = user.ID

	if err := h.expirationService.UpdateExpirationRule(ruleID, updates); err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.SuccessWithMessage(c, "过期规则更新成功", nil)
}

// DeleteExpirationRule 删除过期规则（管理员）
func (h *CreditExpirationHandler) DeleteExpirationRule(c *gin.Context) {
	resp := response.NewGinResponse()

	// 验证管理员权限
	user := c.MustGet("user").(*models.User)
	if user.Role != "admin" && user.Role != "super_admin" {
		resp.Unauthorized(c, "只有管理员可以删除过期规则")
		return
	}

	ruleID := c.Param("id")
	if _, err := uuid.Parse(ruleID); err != nil {
		resp.BadRequest(c, "无效的规则ID")
		return
	}

	if err := h.expirationService.DeleteExpirationRule(ruleID); err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.SuccessWithMessage(c, "过期规则删除成功", nil)
}

// ProcessExpiredCredits 手动处理过期积分（管理员）
func (h *CreditExpirationHandler) ProcessExpiredCredits(c *gin.Context) {
	resp := response.NewGinResponse()

	// 验证管理员权限
	user := c.MustGet("user").(*models.User)
	if user.Role != "admin" && user.Role != "super_admin" {
		resp.Unauthorized(c, "只有管理员可以手动处理过期积分")
		return
	}

	// 异步处理过期积分
	go func() {
		if err := h.expirationService.ProcessExpiredCredits(); err != nil {
			// 这里可以记录错误日志或发送通知给管理员
			// log.Printf("Failed to process expired credits: %v", err)
		}
	}()

	resp.SuccessWithMessage(c, "过期积分处理任务已启动", gin.H{
		"message":    "处理任务已在后台启动",
		"started_at": time.Now(),
	})
}

// SendExpirationWarnings 手动发送过期警告（管理员）
func (h *CreditExpirationHandler) SendExpirationWarnings(c *gin.Context) {
	resp := response.NewGinResponse()

	// 验证管理员权限
	user := c.MustGet("user").(*models.User)
	if user.Role != "admin" && user.Role != "super_admin" {
		resp.Unauthorized(c, "只有管理员可以发送过期警告")
		return
	}

	// 异步发送警告
	go func() {
		if err := h.expirationService.SendExpirationWarnings(); err != nil {
			// 这里可以记录错误日志
			// log.Printf("Failed to send expiration warnings: %v", err)
		}
	}()

	resp.SuccessWithMessage(c, "过期警告发送任务已启动", gin.H{
		"message":    "警告发送任务已在后台启动",
		"started_at": time.Now(),
	})
}

// GetExpirationStatistics 获取过期统计信息（管理员）
func (h *CreditExpirationHandler) GetExpirationStatistics(c *gin.Context) {
	resp := response.NewGinResponse()

	// 验证管理员权限
	user := c.MustGet("user").(*models.User)
	if user.Role != "admin" && user.Role != "super_admin" {
		resp.Unauthorized(c, "只有管理员可以查看过期统计")
		return
	}

	stats, err := h.expirationService.GetExpirationStatistics()
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.Success(c, stats)
}

// GetExpirationBatches 获取过期批次列表（管理员）
func (h *CreditExpirationHandler) GetExpirationBatches(c *gin.Context) {
	resp := response.NewGinResponse()

	// 验证管理员权限
	user := c.MustGet("user").(*models.User)
	if user.Role != "admin" && user.Role != "super_admin" {
		resp.Unauthorized(c, "只有管理员可以查看过期批次")
		return
	}

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// 查询批次记录
	var batches []models.CreditExpirationBatch
	var total int64

	// Get database instance from credit service
	db := h.creditService.GetDB()
	
	err := db.Model(&models.CreditExpirationBatch{}).Count(&total).Error
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	err = db.Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&batches).Error
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.Success(c, gin.H{
		"batches":   batches,
		"total":     total,
		"page":      page,
		"page_size": limit,
		"has_next":  int64(page*limit) < total,
	})
}

// GetExpirationLogs 获取过期日志（管理员）
func (h *CreditExpirationHandler) GetExpirationLogs(c *gin.Context) {
	resp := response.NewGinResponse()

	// 验证管理员权限
	user := c.MustGet("user").(*models.User)
	if user.Role != "admin" && user.Role != "super_admin" {
		resp.Unauthorized(c, "只有管理员可以查看过期日志")
		return
	}

	// 解析查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	userID := c.Query("user_id")
	batchID := c.Query("batch_id")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 50
	}

	offset := (page - 1) * limit

	// Get database instance from credit service
	db := h.creditService.GetDB()
	
	// 构建查询
	query := db.Model(&models.CreditExpirationLog{})
	
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	
	if batchID != "" {
		query = query.Where("batch_id = ?", batchID)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	// 获取记录
	var logs []models.CreditExpirationLog
	err := query.Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&logs).Error
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.Success(c, gin.H{
		"logs":      logs,
		"total":     total,
		"page":      page,
		"page_size": limit,
		"has_next":  int64(page*limit) < total,
		"filters": gin.H{
			"user_id":  userID,
			"batch_id": batchID,
		},
	})
}

// GetExpirationNotifications 获取过期通知记录（管理员）
func (h *CreditExpirationHandler) GetExpirationNotifications(c *gin.Context) {
	resp := response.NewGinResponse()

	// 验证管理员权限
	user := c.MustGet("user").(*models.User)
	if user.Role != "admin" && user.Role != "super_admin" {
		resp.Unauthorized(c, "只有管理员可以查看通知记录")
		return
	}

	// 解析查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	notificationType := c.Query("type") // warning, expired
	sent := c.Query("sent")             // true, false

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 50
	}

	offset := (page - 1) * limit

	// Get database instance from credit service
	db := h.creditService.GetDB()
	
	// 构建查询
	query := db.Model(&models.CreditExpirationNotification{})
	
	if notificationType != "" {
		query = query.Where("notification_type = ?", notificationType)
	}
	
	if sent != "" {
		sentBool := sent == "true"
		query = query.Where("notification_sent = ?", sentBool)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	// 获取记录
	var notifications []models.CreditExpirationNotification
	err := query.Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&notifications).Error
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.Success(c, gin.H{
		"notifications": notifications,
		"total":         total,
		"page":          page,
		"page_size":     limit,
		"has_next":      int64(page*limit) < total,
		"filters": gin.H{
			"type": notificationType,
			"sent": sent,
		},
	})
}