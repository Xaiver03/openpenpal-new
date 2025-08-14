package handlers

import (
	"fmt"
	"net/http"
	"time"

	"openpenpal-backend/internal/models"
	"openpenpal-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// BarcodeHandler 条码系统处理器 - PRD规格实现
type BarcodeHandler struct {
	letterService    *services.LetterService
	opcodeService    *services.OPCodeService
	scanEventService *services.ScanEventService
}

// NewBarcodeHandler 创建条码处理器
func NewBarcodeHandler(letterService *services.LetterService, opcodeService *services.OPCodeService, scanEventService *services.ScanEventService) *BarcodeHandler {
	return &BarcodeHandler{
		letterService:    letterService,
		opcodeService:    opcodeService,
		scanEventService: scanEventService,
	}
}

// BindBarcodeRequest 绑定条码请求
type BindBarcodeRequest struct {
	RecipientOPCode string `json:"recipient_op_code" binding:"required,len=6"`
	EnvelopeID      string `json:"envelope_id,omitempty"`
}

// UpdateBarcodeStatusRequest 更新条码状态请求
type UpdateBarcodeStatusRequest struct {
	Status        models.BarcodeStatus `json:"status" binding:"required"`
	CurrentOPCode string               `json:"current_op_code,omitempty"`
	ScannedBy     string               `json:"scanned_by,omitempty"`
	Note          string               `json:"note,omitempty"`
}

// CreateBarcode 创建条码 - PRD规格: POST /api/barcodes
func (h *BarcodeHandler) CreateBarcode(c *gin.Context) {
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"code":    4001,
			"message": "用户未认证",
		})
		return
	}
	user := userInterface.(*models.User)

	var req struct {
		LetterID string `json:"letter_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"code":    4001,
			"message": "请求参数无效",
			"error":   err.Error(),
		})
		return
	}

	// 验证信件所有权
	var letter models.Letter
	if err := h.letterService.GetDB().Where("id = ? AND user_id = ?", req.LetterID, user.ID).First(&letter).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"code":    4004,
			"message": "信件不存在或无权限",
		})
		return
	}

	// 生成条码
	letterCode, err := h.letterService.GenerateCode(req.LetterID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"code":    5001,
			"message": "生成条码失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"code":    200,
		"message": "条码创建成功",
		"data": gin.H{
			"id":             letterCode.ID,
			"code":           letterCode.Code,
			"status":         letterCode.Status,
			"recipient_code": letterCode.RecipientCode,
			"qr_code_url":    letterCode.QRCodeURL,
			"created_at":     letterCode.CreatedAt,
		},
	})
}

// BindBarcode 绑定条码 - PRD规格: PATCH /api/barcodes/:id/bind
func (h *BarcodeHandler) BindBarcode(c *gin.Context) {
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"code":    4001,
			"message": "用户未认证",
		})
		return
	}
	user := userInterface.(*models.User)

	barcodeID := c.Param("id")
	if barcodeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"code":    4001,
			"message": "条码ID不能为空",
		})
		return
	}

	var req BindBarcodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"code":    4001,
			"message": "请求参数无效",
			"error":   err.Error(),
		})
		return
	}

	// 验证OP Code
	if err := models.ValidateOPCode(req.RecipientOPCode); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"code":    4001,
			"message": "OP Code格式不正确",
			"error":   err.Error(),
		})
		return
	}

	// 获取条码记录
	var letterCode models.LetterCode
	if err := h.letterService.GetDB().Where("id = ?", barcodeID).First(&letterCode).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"code":    4004,
			"message": "条码不存在",
		})
		return
	}

	// 验证条码状态是否可以绑定
	if !letterCode.CanBeBound() {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"code":           4001,
			"message":        "条码状态不允许绑定",
			"current_status": letterCode.Status,
		})
		return
	}

	// 验证用户权限（通过信件所有权）
	var letter models.Letter
	if err := h.letterService.GetDB().Where("id = ? AND user_id = ?", letterCode.LetterID, user.ID).First(&letter).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"code":    4003,
			"message": "无权限操作此条码",
		})
		return
	}

	// 执行绑定
	now := time.Now()
	letterCode.Status = models.BarcodeStatusBound
	letterCode.RecipientCode = req.RecipientOPCode
	letterCode.EnvelopeID = req.EnvelopeID
	letterCode.BoundAt = &now
	letterCode.UpdatedAt = now

	if err := h.letterService.GetDB().Save(&letterCode).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"code":    5001,
			"message": "绑定失败",
			"error":   err.Error(),
		})
		return
	}

	// 记录扫描事件
	h.recordScanEvent(c, letterCode.ID, user.ID, models.ScanEventTypeBind, req.RecipientOPCode, models.BarcodeStatusUnactivated, models.BarcodeStatusBound, "条码绑定成功")

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"code":    200,
		"message": "绑定成功",
		"data": gin.H{
			"id":             letterCode.ID,
			"status":         letterCode.Status,
			"recipient_code": letterCode.RecipientCode,
			"bound_at":       letterCode.BoundAt,
		},
	})
}

// UpdateBarcodeStatus 更新条码状态 - PRD规格: PATCH /api/barcodes/:id/status
func (h *BarcodeHandler) UpdateBarcodeStatus(c *gin.Context) {
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"code":    4001,
			"message": "用户未认证",
		})
		return
	}
	user := userInterface.(*models.User)

	barcodeID := c.Param("id")
	if barcodeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"code":    4001,
			"message": "条码ID不能为空",
		})
		return
	}

	var req UpdateBarcodeStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"code":    4001,
			"message": "请求参数无效",
			"error":   err.Error(),
		})
		return
	}

	// 获取条码记录
	var letterCode models.LetterCode
	if err := h.letterService.GetDB().Where("id = ?", barcodeID).First(&letterCode).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"code":    4004,
			"message": "条码不存在",
		})
		return
	}

	// 验证状态转换是否有效
	if !letterCode.IsValidTransition(req.Status) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"code":           4001,
			"message":        "无效的状态转换",
			"current_status": letterCode.Status,
			"target_status":  req.Status,
		})
		return
	}

	// 验证用户权限（信使权限检查）
	if user.Role != models.RoleCourierLevel1 && user.Role != models.RoleCourierLevel2 &&
		user.Role != models.RoleCourierLevel3 && user.Role != models.RoleCourierLevel4 &&
		user.Role != models.RolePlatformAdmin && user.Role != models.RoleSuperAdmin {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"code":    4003,
			"message": "需要信使权限",
		})
		return
	}

	// 更新状态
	oldStatus := letterCode.Status
	now := time.Now()
	letterCode.Status = req.Status
	letterCode.LastScannedBy = user.ID
	letterCode.LastScannedAt = &now
	letterCode.ScanCount++
	letterCode.UpdatedAt = now

	if req.Status == models.BarcodeStatusDelivered {
		letterCode.DeliveredAt = &now
	}

	if err := h.letterService.GetDB().Save(&letterCode).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"code":    5001,
			"message": "状态更新失败",
			"error":   err.Error(),
		})
		return
	}

	// 记录扫描事件
	scanType := models.ScanEventTypeTransit
	if req.Status == models.BarcodeStatusInTransit && oldStatus == models.BarcodeStatusBound {
		scanType = models.ScanEventTypePickup
	} else if req.Status == models.BarcodeStatusDelivered {
		scanType = models.ScanEventTypeDelivery
	}

	h.recordScanEvent(c, letterCode.ID, user.ID, scanType, req.CurrentOPCode, oldStatus, req.Status, req.Note)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"code":    200,
		"message": "状态更新成功",
		"data": gin.H{
			"id":              letterCode.ID,
			"status":          letterCode.Status,
			"last_scanned_by": letterCode.LastScannedBy,
			"last_scanned_at": letterCode.LastScannedAt,
			"scan_count":      letterCode.ScanCount,
		},
	})
}

// GetBarcodeStatus 获取条码状态 - PRD规格: GET /api/barcodes/:id/status
func (h *BarcodeHandler) GetBarcodeStatus(c *gin.Context) {
	barcodeID := c.Param("id")
	if barcodeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"code":    4001,
			"message": "条码ID不能为空",
		})
		return
	}

	var letterCode models.LetterCode
	if err := h.letterService.GetDB().Where("id = ?", barcodeID).First(&letterCode).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"code":    4004,
			"message": "条码不存在",
		})
		return
	}

	// 获取扫描历史
	scanHistory, _ := h.getScanHistory(letterCode.ID)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"id":              letterCode.ID,
			"code":            letterCode.Code,
			"status":          letterCode.Status,
			"status_display":  letterCode.GetStatusDisplayName(),
			"recipient_code":  letterCode.RecipientCode,
			"envelope_id":     letterCode.EnvelopeID,
			"bound_at":        letterCode.BoundAt,
			"delivered_at":    letterCode.DeliveredAt,
			"last_scanned_by": letterCode.LastScannedBy,
			"last_scanned_at": letterCode.LastScannedAt,
			"scan_count":      letterCode.ScanCount,
			"scan_history":    scanHistory,
			"created_at":      letterCode.CreatedAt,
			"updated_at":      letterCode.UpdatedAt,
		},
	})
}

// ValidateBarcodeOperation 验证条码操作权限 - PRD规格: POST /api/barcodes/:id/validate
func (h *BarcodeHandler) ValidateBarcodeOperation(c *gin.Context) {
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"code":    4001,
			"message": "用户未认证",
		})
		return
	}
	user := userInterface.(*models.User)

	barcodeID := c.Param("id")
	if barcodeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"code":    4001,
			"message": "条码ID不能为空",
		})
		return
	}

	var req struct {
		Operation    string `json:"operation" binding:"required"` // bind, scan, deliver
		TargetOPCode string `json:"target_op_code,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"code":    4001,
			"message": "请求参数无效",
			"error":   err.Error(),
		})
		return
	}

	var letterCode models.LetterCode
	if err := h.letterService.GetDB().Where("id = ?", barcodeID).First(&letterCode).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"code":    4004,
			"message": "条码不存在",
		})
		return
	}

	// 验证操作权限
	canOperate := false
	reason := ""

	switch req.Operation {
	case "bind":
		canOperate = letterCode.CanBeBound()
		if !canOperate {
			reason = "条码状态不允许绑定"
		}
	case "scan":
		canOperate = letterCode.IsActive()
		if !canOperate {
			reason = "条码已失效"
		}
		// 验证信使OP Code权限
		if canOperate && req.TargetOPCode != "" && h.opcodeService != nil {
			hasAccess, err := h.opcodeService.ValidateCourierAccess(user.ID, req.TargetOPCode)
			if err != nil || !hasAccess {
				canOperate = false
				reason = "无权限访问目标OP Code区域"
			}
		}
	case "deliver":
		canOperate = letterCode.Status == models.BarcodeStatusInTransit
		if !canOperate {
			reason = "条码状态不允许送达"
		}
	default:
		canOperate = false
		reason = "不支持的操作类型"
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"code":    200,
		"message": "验证完成",
		"data": gin.H{
			"can_operate":    canOperate,
			"reason":         reason,
			"current_status": letterCode.Status,
			"user_role":      user.Role,
		},
	})
}

// recordScanEvent 记录扫描事件（内部方法）
func (h *BarcodeHandler) recordScanEvent(c *gin.Context, barcodeID, scannedBy string, scanType models.ScanEventType, opCode string, oldStatus, newStatus models.BarcodeStatus, note string) {
	if h.scanEventService == nil {
		return // 如果扫描事件服务未初始化，直接返回
	}

	// 获取请求信息
	userAgent := c.GetHeader("User-Agent")
	ipAddress := c.ClientIP()

	// 创建扫描事件
	req := &models.ScanEventCreateRequest{
		BarcodeID: barcodeID,
		ScanType:  scanType,
		OPCode:    opCode,
		OldStatus: oldStatus,
		NewStatus: newStatus,
		Note:      note,
	}

	// 记录扫描事件
	_, err := h.scanEventService.CreateScanEvent(req, scannedBy, userAgent, ipAddress)
	if err != nil {
		// 记录失败不影响主流程，只记录日志
		fmt.Printf("Warning: Failed to record scan event: %v\n", err)
	}
}

// getScanHistory 获取扫描历史（内部方法）
func (h *BarcodeHandler) getScanHistory(barcodeID string) ([]models.ScanEvent, error) {
	if h.scanEventService == nil {
		return []models.ScanEvent{}, nil
	}

	// 获取条码时间线
	events, err := h.scanEventService.GetBarcodeTimeline(barcodeID)
	if err != nil {
		return []models.ScanEvent{}, err
	}

	return events, nil
}
