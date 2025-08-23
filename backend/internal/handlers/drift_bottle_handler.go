package handlers

import (
	"net/http"
	"strconv"
	"time"

	"openpenpal-backend/internal/services"
	"openpenpal-backend/internal/models"

	"github.com/gin-gonic/gin"
)

// DriftBottleHandler 漂流瓶处理器
type DriftBottleHandler struct {
	driftBottleService *services.DriftBottleService
}

// NewDriftBottleHandler 创建漂流瓶处理器实例
func NewDriftBottleHandler(driftBottleService *services.DriftBottleService) *DriftBottleHandler {
	return &DriftBottleHandler{
		driftBottleService: driftBottleService,
	}
}

// CreateDriftBottle 创建漂流瓶
func (h *DriftBottleHandler) CreateDriftBottle(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req services.DriftBottleCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	bottle, err := h.driftBottleService.CreateDriftBottle(c.Request.Context(), userID, &req)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "letter not found" || err.Error() == "unauthorized: letter does not belong to user" {
			status = http.StatusBadRequest
		} else if err.Error() == "letter is already a drift bottle" {
			status = http.StatusConflict
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    bottle,
	})
}

// CollectDriftBottle 捞取漂流瓶
func (h *DriftBottleHandler) CollectDriftBottle(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	bottle, err := h.driftBottleService.CollectDriftBottle(c.Request.Context(), userID)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "no drift bottles available" {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    bottle,
	})
}

// GetMyDriftBottles 获取我的漂流瓶列表
func (h *DriftBottleHandler) GetMyDriftBottles(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	bottles, total, err := h.driftBottleService.GetMyDriftBottles(c.Request.Context(), userID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"bottles": bottles,
			"total":   total,
			"page":    page,
			"limit":   limit,
		},
	})
}

// GetCollectedBottles 获取我捞取的漂流瓶
func (h *DriftBottleHandler) GetCollectedBottles(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	bottles, total, err := h.driftBottleService.GetCollectedBottles(c.Request.Context(), userID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"bottles": bottles,
			"total":   total,
			"page":    page,
			"limit":   limit,
		},
	})
}

// GetFloatingBottles 获取漂流中的瓶子（用于展示）
func (h *DriftBottleHandler) GetFloatingBottles(c *gin.Context) {
	// 解析参数
	region := c.Query("region")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit < 1 || limit > 50 {
		limit = 20
	}

	bottles, err := h.driftBottleService.GetFloatingBottles(c.Request.Context(), region, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    bottles,
	})
}

// GetDriftBottleStats 获取漂流瓶统计信息
func (h *DriftBottleHandler) GetDriftBottleStats(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// 统计我发送的漂流瓶
	var sentCount int64
	var collectedCount int64
	var floatingCount int64

	// 使用原始查询来获取统计数据
	if err := h.driftBottleService.GetDB().Model(&models.DriftBottle{}).
		Where("sender_id = ?", userID).Count(&sentCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get sent count"})
		return
	}

	// 统计我捞取的漂流瓶
	if err := h.driftBottleService.GetDB().Model(&models.DriftBottle{}).
		Where("collector_id = ?", userID).Count(&collectedCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get collected count"})
		return
	}

	// 统计我的漂流瓶中仍在漂流的
	if err := h.driftBottleService.GetDB().Model(&models.DriftBottle{}).
		Where("sender_id = ? AND status = ?", userID, models.DriftBottleStatusFloating).
		Where("expires_at > ?", time.Now()).
		Count(&floatingCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get floating count"})
		return
	}

	stats := gin.H{
		"sent_count":      sentCount,
		"collected_count": collectedCount,
		"floating_count":  floatingCount,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// RegisterDriftBottleRoutes 注册漂流瓶相关路由
func (h *DriftBottleHandler) RegisterDriftBottleRoutes(router *gin.RouterGroup) {
	driftBottle := router.Group("/drift-bottles")
	{
		driftBottle.POST("", h.CreateDriftBottle)              // 创建漂流瓶
		driftBottle.POST("/collect", h.CollectDriftBottle)     // 捞取漂流瓶
		driftBottle.GET("/my", h.GetMyDriftBottles)            // 我的漂流瓶
		driftBottle.GET("/collected", h.GetCollectedBottles)   // 我捞取的漂流瓶
		driftBottle.GET("/floating", h.GetFloatingBottles)     // 漂流中的瓶子
		driftBottle.GET("/stats", h.GetDriftBottleStats)       // 统计信息
	}
}