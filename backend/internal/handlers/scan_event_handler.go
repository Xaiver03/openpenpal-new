package handlers

import (
	"net/http"
	"strconv"

	"openpenpal-backend/internal/models"
	"openpenpal-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// ScanEventHandler 扫描事件处理器
type ScanEventHandler struct {
	scanEventService *services.ScanEventService
}

// NewScanEventHandler 创建扫描事件处理器
func NewScanEventHandler(scanEventService *services.ScanEventService) *ScanEventHandler {
	return &ScanEventHandler{
		scanEventService: scanEventService,
	}
}

// GetScanHistory 获取扫描历史
func (h *ScanEventHandler) GetScanHistory(c *gin.Context) {
	var query models.ScanEventQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"code":    4001,
			"message": "查询参数无效",
			"error":   err.Error(),
		})
		return
	}

	// 设置默认值
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 || query.PageSize > 100 {
		query.PageSize = 20
	}

	events, total, err := h.scanEventService.GetScanHistory(&query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"code":    5001,
			"message": "获取扫描历史失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"events": events,
			"pagination": gin.H{
				"page":       query.Page,
				"page_size":  query.PageSize,
				"total":      total,
				"total_page": (total + int64(query.PageSize) - 1) / int64(query.PageSize),
			},
		},
	})
}

// GetScanEventByID 根据ID获取扫描事件
func (h *ScanEventHandler) GetScanEventByID(c *gin.Context) {
	eventID := c.Param("id")
	if eventID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"code":    4001,
			"message": "扫描事件ID不能为空",
		})
		return
	}

	event, err := h.scanEventService.GetScanEventByID(eventID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"code":    4004,
			"message": "扫描事件不存在",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"code":    200,
		"message": "获取成功",
		"data":    event,
	})
}

// GetBarcodeTimeline 获取条码时间线
func (h *ScanEventHandler) GetBarcodeTimeline(c *gin.Context) {
	barcodeID := c.Param("barcode_id")
	if barcodeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"code":    4001,
			"message": "条码ID不能为空",
		})
		return
	}

	events, err := h.scanEventService.GetBarcodeTimeline(barcodeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"code":    5001,
			"message": "获取条码时间线失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"barcode_id": barcodeID,
			"timeline":   events,
		},
	})
}

// GetScanEventSummary 获取扫描事件统计摘要
func (h *ScanEventHandler) GetScanEventSummary(c *gin.Context) {
	var query models.ScanEventQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"code":    4001,
			"message": "查询参数无效",
			"error":   err.Error(),
		})
		return
	}

	summary, err := h.scanEventService.GetScanEventSummary(&query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"code":    5001,
			"message": "获取统计摘要失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"code":    200,
		"message": "获取成功",
		"data":    summary,
	})
}

// GetUserScanActivity 获取用户扫描活动
func (h *ScanEventHandler) GetUserScanActivity(c *gin.Context) {
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

	// 获取查询参数
	daysParam := c.DefaultQuery("days", "30")
	days, err := strconv.Atoi(daysParam)
	if err != nil || days <= 0 || days > 365 {
		days = 30
	}

	events, err := h.scanEventService.GetUserScanActivity(user.ID, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"code":    5001,
			"message": "获取用户扫描活动失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"code":    200,
		"message": "获取成功",
		"data": gin.H{
			"user_id": user.ID,
			"days":    days,
			"events":  events,
		},
	})
}

// GetLocationScanStats 获取位置扫描统计
func (h *ScanEventHandler) GetLocationScanStats(c *gin.Context) {
	opCode := c.Param("op_code")
	if opCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"code":    4001,
			"message": "OP Code不能为空",
		})
		return
	}

	// 验证OP Code格式
	if err := models.ValidateOPCode(opCode); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"code":    4001,
			"message": "OP Code格式不正确",
			"error":   err.Error(),
		})
		return
	}

	// 获取查询参数
	daysParam := c.DefaultQuery("days", "7")
	days, err := strconv.Atoi(daysParam)
	if err != nil || days <= 0 || days > 365 {
		days = 7
	}

	stats, err := h.scanEventService.GetLocationScanStats(opCode, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"code":    5001,
			"message": "获取位置扫描统计失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"code":    200,
		"message": "获取成功",
		"data":    stats,
	})
}

// CreateScanEvent 手动创建扫描事件（管理员功能）
func (h *ScanEventHandler) CreateScanEvent(c *gin.Context) {
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

	// 检查权限
	if user.Role != models.RolePlatformAdmin && user.Role != models.RoleSuperAdmin {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"code":    4003,
			"message": "需要管理员权限",
		})
		return
	}

	var req models.ScanEventCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"code":    4001,
			"message": "请求参数无效",
			"error":   err.Error(),
		})
		return
	}

	// 获取请求信息
	userAgent := c.GetHeader("User-Agent")
	ipAddress := c.ClientIP()

	event, err := h.scanEventService.CreateScanEvent(&req, user.ID, userAgent, ipAddress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"code":    5001,
			"message": "创建扫描事件失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"code":    200,
		"message": "创建成功",
		"data":    event,
	})
}

// CleanupOldScanEvents 清理旧的扫描事件（管理员功能）
func (h *ScanEventHandler) CleanupOldScanEvents(c *gin.Context) {
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

	// 检查权限
	if user.Role != models.RolePlatformAdmin && user.Role != models.RoleSuperAdmin {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"code":    4003,
			"message": "需要管理员权限",
		})
		return
	}

	var req struct {
		Days int `json:"days" binding:"required,min=1,max=365"`
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

	deletedCount, err := h.scanEventService.CleanupOldScanEvents(req.Days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"code":    5001,
			"message": "清理失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"code":    200,
		"message": "清理完成",
		"data": gin.H{
			"deleted_count": deletedCount,
			"days":          req.Days,
		},
	})
}