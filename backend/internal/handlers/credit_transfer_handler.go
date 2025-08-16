package handlers

import (
	"strconv"
	"time"

	"openpenpal-backend/internal/models"
	"openpenpal-backend/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"openpenpal-backend/internal/pkg/response"
)

// Phase 4.2: 积分转赠处理器

// CreditTransferHandler 积分转赠处理器
type CreditTransferHandler struct {
	transferService *services.CreditTransferService
	creditService   *services.CreditService
}

// NewCreditTransferHandler 创建积分转赠处理器实例
func NewCreditTransferHandler(transferService *services.CreditTransferService, creditService *services.CreditService) *CreditTransferHandler {
	return &CreditTransferHandler{
		transferService: transferService,
		creditService:   creditService,
	}
}

// ===================== 用户相关API =====================

// CreateTransfer 创建积分转赠
func (h *CreditTransferHandler) CreateTransfer(c *gin.Context) {
	resp := response.NewGinResponse()
	user := c.MustGet("user").(*models.User)

	var request models.CreateCreditTransferRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	// 验证用户ID格式
	if _, err := uuid.Parse(request.ToUserID); err != nil {
		resp.BadRequest(c, "无效的用户ID格式")
		return
	}

	transfer, err := h.transferService.CreateTransfer(user.ID, &request)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	// 构建响应
	transferResponse := &models.CreditTransferResponse{
		Transfer:     *transfer,
		FromUsername: user.Username,
		CanCancel:    transfer.CanBeCanceled(),
		CanProcess:   transfer.CanBeProcessed(),
		IsExpired:    transfer.IsExpired(),
	}

	resp.CreatedWithMessage(c, "积分转赠创建成功", transferResponse)
}

// ProcessTransfer 处理积分转赠
func (h *CreditTransferHandler) ProcessTransfer(c *gin.Context) {
	resp := response.NewGinResponse()
	user := c.MustGet("user").(*models.User)

	transferID := c.Param("id")
	if _, err := uuid.Parse(transferID); err != nil {
		resp.BadRequest(c, "无效的转赠ID")
		return
	}

	var request models.ProcessCreditTransferRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	if err := h.transferService.ProcessTransfer(transferID, user.ID, &request); err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	message := "积分转赠已接受"
	if request.Action == "reject" {
		message = "积分转赠已拒绝"
	}

	resp.SuccessWithMessage(c, message, gin.H{
		"transfer_id": transferID,
		"action":      request.Action,
	})
}

// CancelTransfer 取消积分转赠
func (h *CreditTransferHandler) CancelTransfer(c *gin.Context) {
	resp := response.NewGinResponse()
	user := c.MustGet("user").(*models.User)

	transferID := c.Param("id")
	if _, err := uuid.Parse(transferID); err != nil {
		resp.BadRequest(c, "无效的转赠ID")
		return
	}

	if err := h.transferService.CancelTransfer(transferID, user.ID); err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.SuccessWithMessage(c, "积分转赠已取消", gin.H{
		"transfer_id": transferID,
	})
}

// GetUserTransfers 获取用户转赠记录
func (h *CreditTransferHandler) GetUserTransfers(c *gin.Context) {
	resp := response.NewGinResponse()
	user := c.MustGet("user").(*models.User)

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	statusStr := c.Query("status")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	var status models.CreditTransferStatus
	if statusStr != "" {
		status = models.CreditTransferStatus(statusStr)
	}

	transfers, total, err := h.transferService.GetUserTransfers(user.ID, page, limit, status)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	// 构建响应数据
	transferResponses := make([]models.CreditTransferResponse, len(transfers))
	for i, transfer := range transfers {
		transferResponses[i] = models.CreditTransferResponse{
			Transfer:     transfer,
			FromUsername: transfer.FromUser.Username,
			ToUsername:   transfer.ToUser.Username,
			CanCancel:    transfer.CanBeCanceled(),
			CanProcess:   transfer.CanBeProcessed(),
			IsExpired:    transfer.IsExpired(),
		}
	}

	resp.Success(c, gin.H{
		"transfers":  transferResponses,
		"total":      total,
		"page":       page,
		"page_size":  limit,
		"has_next":   int64(page*limit) < total,
	})
}

// GetTransferStats 获取用户转赠统计
func (h *CreditTransferHandler) GetTransferStats(c *gin.Context) {
	resp := response.NewGinResponse()
	user := c.MustGet("user").(*models.User)

	stats, err := h.transferService.GetTransferStats(user.ID)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.Success(c, stats)
}

// GetTransfer 获取单个转赠详情
func (h *CreditTransferHandler) GetTransfer(c *gin.Context) {
	resp := response.NewGinResponse()
	user := c.MustGet("user").(*models.User)

	transferID := c.Param("id")
	if _, err := uuid.Parse(transferID); err != nil {
		resp.BadRequest(c, "无效的转赠ID")
		return
	}

	var transfer models.CreditTransfer
	err := h.transferService.GetDB().Where("id = ? AND (from_user_id = ? OR to_user_id = ?)", 
		transferID, user.ID, user.ID).
		Preload("FromUser").Preload("ToUser").
		First(&transfer).Error

	if err != nil {
		resp.NotFound(c, "转赠记录未找到")
		return
	}

	transferResponse := models.CreditTransferResponse{
		Transfer:     transfer,
		FromUsername: transfer.FromUser.Username,
		ToUsername:   transfer.ToUser.Username,
		CanCancel:    transfer.CanBeCanceled() && transfer.FromUserID == user.ID,
		CanProcess:   transfer.CanBeProcessed() && transfer.ToUserID == user.ID,
		IsExpired:    transfer.IsExpired(),
	}

	resp.Success(c, transferResponse)
}

// BatchTransfer 批量转赠积分
func (h *CreditTransferHandler) BatchTransfer(c *gin.Context) {
	resp := response.NewGinResponse()
	user := c.MustGet("user").(*models.User)

	var request models.BatchTransferRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	// 验证用户ID格式
	for _, userID := range request.ToUserIDs {
		if _, err := uuid.Parse(userID); err != nil {
			resp.BadRequest(c, "无效的用户ID格式: "+userID)
			return
		}
	}

	batchResponse, err := h.transferService.BatchTransfer(user.ID, &request)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.SuccessWithMessage(c, "批量转赠处理完成", batchResponse)
}

// GetPendingTransfers 获取待处理的转赠
func (h *CreditTransferHandler) GetPendingTransfers(c *gin.Context) {
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

	transfers, _, err := h.transferService.GetUserTransfers(user.ID, page, limit, models.TransferStatusPending)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	// 只返回转入给当前用户的待处理转赠
	pendingTransfers := make([]models.CreditTransferResponse, 0)
	for _, transfer := range transfers {
		if transfer.ToUserID == user.ID {
			pendingTransfers = append(pendingTransfers, models.CreditTransferResponse{
				Transfer:     transfer,
				FromUsername: transfer.FromUser.Username,
				ToUsername:   transfer.ToUser.Username,
				CanCancel:    false,
				CanProcess:   transfer.CanBeProcessed(),
				IsExpired:    transfer.IsExpired(),
			})
		}
	}

	resp.Success(c, gin.H{
		"transfers": pendingTransfers,
		"total":     len(pendingTransfers),
		"page":      page,
		"page_size": limit,
	})
}

// ValidateTransfer 验证转赠可行性
func (h *CreditTransferHandler) ValidateTransfer(c *gin.Context) {
	resp := response.NewGinResponse()
	user := c.MustGet("user").(*models.User)

	var request models.CreateCreditTransferRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	// 获取用户积分信息
	userCredit, err := h.creditService.GetUserCreditInfo(user.ID)
	if err != nil {
		resp.InternalServerError(c, "获取积分信息失败")
		return
	}

	// 模拟创建转赠（不实际创建）来验证可行性
	validation := gin.H{
		"valid": true,
		"user_available_credits": userCredit.Available,
		"estimated_fee": 0,
		"total_required": request.Amount,
		"remaining_after_transfer": userCredit.Available - request.Amount,
	}

	// 这里可以添加更多验证逻辑，如获取费率等
	if userCredit.Available < request.Amount {
		validation["valid"] = false
		validation["error"] = "积分余额不足"
	}

	resp.Success(c, validation)
}

// ===================== 管理员API =====================

// GetAllTransfers 获取所有转赠记录（管理员）
func (h *CreditTransferHandler) GetAllTransfers(c *gin.Context) {
	resp := response.NewGinResponse()

	// 验证管理员权限
	user := c.MustGet("user").(*models.User)
	if user.Role != "admin" && user.Role != models.RolePlatformAdmin && user.Role != models.RoleSuperAdmin {
		resp.Unauthorized(c, "需要管理员权限")
		return
	}

	// 解析查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	statusStr := c.Query("status")
	userID := c.Query("user_id")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 50
	}

	offset := (page - 1) * limit

	// 构建查询
	query := h.transferService.GetDB().Model(&models.CreditTransfer{})

	if statusStr != "" {
		query = query.Where("status = ?", statusStr)
	}

	if userID != "" {
		query = query.Where("from_user_id = ? OR to_user_id = ?", userID, userID)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	// 获取数据
	var transfers []models.CreditTransfer
	err := query.Preload("FromUser").Preload("ToUser").
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&transfers).Error

	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	// 构建响应
	transferResponses := make([]models.CreditTransferResponse, len(transfers))
	for i, transfer := range transfers {
		transferResponses[i] = models.CreditTransferResponse{
			Transfer:     transfer,
			FromUsername: transfer.FromUser.Username,
			ToUsername:   transfer.ToUser.Username,
			CanCancel:    transfer.CanBeCanceled(),
			CanProcess:   transfer.CanBeProcessed(),
			IsExpired:    transfer.IsExpired(),
		}
	}

	resp.Success(c, gin.H{
		"transfers": transferResponses,
		"total":     total,
		"page":      page,
		"page_size": limit,
		"has_next":  int64(page*limit) < total,
	})
}

// GetTransferStatistics 获取转赠统计信息（管理员）
func (h *CreditTransferHandler) GetTransferStatistics(c *gin.Context) {
	resp := response.NewGinResponse()

	// 验证管理员权限
	user := c.MustGet("user").(*models.User)
	if user.Role != "admin" && user.Role != models.RolePlatformAdmin && user.Role != models.RoleSuperAdmin {
		resp.Unauthorized(c, "需要管理员权限")
		return
	}

	// 获取统计数据
	stats := make(map[string]interface{})

	// 总体统计
	var totalStats struct {
		TotalTransfers    int64   `json:"total_transfers"`
		TotalAmount       int64   `json:"total_amount"`
		TotalFees         int64   `json:"total_fees"`
		PendingTransfers  int64   `json:"pending_transfers"`
		ProcessedTransfers int64  `json:"processed_transfers"`
		CanceledTransfers int64   `json:"canceled_transfers"`
		ExpiredTransfers  int64   `json:"expired_transfers"`
		RejectedTransfers int64   `json:"rejected_transfers"`
	}

	err := h.transferService.GetDB().Model(&models.CreditTransfer{}).
		Select(`
			COUNT(*) as total_transfers,
			COALESCE(SUM(amount), 0) as total_amount,
			COALESCE(SUM(fee), 0) as total_fees,
			COALESCE(SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END), 0) as pending_transfers,
			COALESCE(SUM(CASE WHEN status = 'processed' THEN 1 ELSE 0 END), 0) as processed_transfers,
			COALESCE(SUM(CASE WHEN status = 'canceled' THEN 1 ELSE 0 END), 0) as canceled_transfers,
			COALESCE(SUM(CASE WHEN status = 'expired' THEN 1 ELSE 0 END), 0) as expired_transfers,
			COALESCE(SUM(CASE WHEN status = 'rejected' THEN 1 ELSE 0 END), 0) as rejected_transfers
		`).
		Scan(&totalStats).Error

	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	stats["overview"] = totalStats

	// 今日统计
	today := time.Now().Truncate(24 * time.Hour)
	var todayStats struct {
		TodayTransfers int64 `json:"today_transfers"`
		TodayAmount    int64 `json:"today_amount"`
		TodayFees      int64 `json:"today_fees"`
	}

	err = h.transferService.GetDB().Model(&models.CreditTransfer{}).
		Select(`
			COUNT(*) as today_transfers,
			COALESCE(SUM(amount), 0) as today_amount,
			COALESCE(SUM(fee), 0) as today_fees
		`).
		Where("created_at >= ?", today).
		Scan(&todayStats).Error

	if err == nil {
		stats["today"] = todayStats
	}

	// 按类型统计
	var typeStats []struct {
		TransferType string `json:"transfer_type"`
		Count        int64  `json:"count"`
		TotalAmount  int64  `json:"total_amount"`
	}

	err = h.transferService.GetDB().Model(&models.CreditTransfer{}).
		Select("transfer_type, COUNT(*) as count, COALESCE(SUM(amount), 0) as total_amount").
		Group("transfer_type").
		Scan(&typeStats).Error

	if err == nil {
		stats["by_type"] = typeStats
	}

	resp.Success(c, stats)
}

// ProcessExpiredTransfers 手动处理过期转赠（管理员）
func (h *CreditTransferHandler) ProcessExpiredTransfers(c *gin.Context) {
	resp := response.NewGinResponse()

	// 验证管理员权限
	user := c.MustGet("user").(*models.User)
	if user.Role != "admin" && user.Role != models.RolePlatformAdmin && user.Role != models.RoleSuperAdmin {
		resp.Unauthorized(c, "需要管理员权限")
		return
	}

	// 异步处理过期转赠
	go func() {
		if err := h.transferService.ProcessExpiredTransfers(); err != nil {
			// 这里可以记录错误日志或发送通知给管理员
		}
	}()

	resp.SuccessWithMessage(c, "过期转赠处理任务已启动", gin.H{
		"message":    "处理任务已在后台启动",
		"started_at": time.Now(),
	})
}

// AdminCancelTransfer 管理员取消转赠
func (h *CreditTransferHandler) AdminCancelTransfer(c *gin.Context) {
	resp := response.NewGinResponse()

	// 验证管理员权限
	user := c.MustGet("user").(*models.User)
	if user.Role != "admin" && user.Role != models.RolePlatformAdmin && user.Role != models.RoleSuperAdmin {
		resp.Unauthorized(c, "需要管理员权限")
		return
	}

	transferID := c.Param("id")
	if _, err := uuid.Parse(transferID); err != nil {
		resp.BadRequest(c, "无效的转赠ID")
		return
	}

	var request struct {
		Reason string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	// 获取转赠记录
	var transfer models.CreditTransfer
	if err := h.transferService.GetDB().Where("id = ?", transferID).First(&transfer).Error; err != nil {
		resp.NotFound(c, "转赠记录未找到")
		return
	}

	// 管理员可以取消任何待处理的转赠
	if transfer.Status != models.TransferStatusPending {
		resp.BadRequest(c, "只能取消待处理状态的转赠")
		return
	}

	// 使用转出用户的身份取消转赠
	if err := h.transferService.CancelTransfer(transferID, transfer.FromUserID); err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.SuccessWithMessage(c, "转赠已被管理员取消", gin.H{
		"transfer_id": transferID,
		"reason":      request.Reason,
		"admin_id":    user.ID,
	})
}

// 辅助方法

// GetDB 获取数据库实例（用于其他服务访问）
func (h *CreditTransferHandler) GetDB() *gorm.DB {
	return h.transferService.GetDB()
}