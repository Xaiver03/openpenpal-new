package handlers

import (
	"net/http"
	"strconv"
	"time"

	"openpenpal-backend/internal/services"
	"openpenpal-backend/internal/models"

	"github.com/gin-gonic/gin"
)

// FutureLetterHandler 未来信处理器
type FutureLetterHandler struct {
	futureLetterService *services.FutureLetterService
}

// NewFutureLetterHandler 创建未来信处理器实例
func NewFutureLetterHandler(futureLetterService *services.FutureLetterService) *FutureLetterHandler {
	return &FutureLetterHandler{
		futureLetterService: futureLetterService,
	}
}

// ScheduleFutureLetter 创建定时信件
func (h *FutureLetterHandler) ScheduleFutureLetter(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		LetterID      string    `json:"letter_id" binding:"required"`
		ScheduledAt   time.Time `json:"scheduled_at" binding:"required"`
		RecipientID   string    `json:"recipient_id,omitempty"`
		RecipientCode string    `json:"recipient_code,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// 验证计划时间必须在未来
	if req.ScheduledAt.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "scheduled time must be in the future"})
		return
	}

	// 使用现有的服务方法处理调度
	err := h.futureLetterService.ProcessScheduledLetters(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to schedule letter"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Letter scheduled successfully",
	})
}

// GetScheduledLetters 获取已调度的信件列表
func (h *FutureLetterHandler) GetScheduledLetters(c *gin.Context) {
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

	// 获取即将发布的信件
	letters, err := h.futureLetterService.GetUpcomingLetters(c.Request.Context(), 24*30) // 30天内
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 过滤出属于当前用户的信件
	var userLetters []models.Letter
	for _, letter := range letters {
		if letter.UserID == userID {
			userLetters = append(userLetters, letter)
		}
	}

	// 简单分页处理
	start := (page - 1) * limit
	end := start + limit
	if start > len(userLetters) {
		start = len(userLetters)
	}
	if end > len(userLetters) {
		end = len(userLetters)
	}

	paginatedLetters := userLetters[start:end]
	total := int64(len(userLetters))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"letters": paginatedLetters,
			"total":   total,
			"page":    page,
			"limit":   limit,
		},
	})
}

// CancelScheduledLetter 取消已调度的信件
func (h *FutureLetterHandler) CancelScheduledLetter(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	letterID := c.Param("id")
	if letterID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "letter ID is required"})
		return
	}

	err := h.futureLetterService.CancelScheduledLetter(c.Request.Context(), letterID, userID)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "letter not found or not scheduled" {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Scheduled letter cancelled successfully",
	})
}

// GetFutureLetterStats 获取未来信统计信息
func (h *FutureLetterHandler) GetFutureLetterStats(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// 获取待发送数量
	pendingCount, err := h.futureLetterService.GetPendingCount(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get pending count"})
		return
	}

	// 获取即将到期的信件（24小时内）
	upcomingLetters, err := h.futureLetterService.GetUpcomingLetters(c.Request.Context(), 24)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get upcoming letters"})
		return
	}

	// 过滤用户的信件
	var userUpcomingCount int
	for _, letter := range upcomingLetters {
		if letter.UserID == userID {
			userUpcomingCount++
		}
	}

	stats := gin.H{
		"pending_count":       pendingCount,
		"upcoming_24h_count":  userUpcomingCount,
		"total_system_pending": pendingCount,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// ProcessPendingLetters 手动触发处理待发送信件（管理员功能）
func (h *FutureLetterHandler) ProcessPendingLetters(c *gin.Context) {
	// 检查管理员权限
	userRole := c.GetString("user_role")
	if userRole != "admin" && userRole != "super_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
		return
	}

	err := h.futureLetterService.ProcessScheduledLetters(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Pending letters processed successfully",
	})
}

// RegisterFutureLetterRoutes 注册未来信相关路由
func (h *FutureLetterHandler) RegisterFutureLetterRoutes(router *gin.RouterGroup) {
	futureLetter := router.Group("/future-letters")
	{
		futureLetter.POST("/schedule", h.ScheduleFutureLetter)           // 调度信件
		futureLetter.GET("", h.GetScheduledLetters)                     // 获取已调度信件
		futureLetter.DELETE("/:id", h.CancelScheduledLetter)           // 取消调度
		futureLetter.GET("/stats", h.GetFutureLetterStats)             // 统计信息
		futureLetter.POST("/process", h.ProcessPendingLetters)         // 手动处理（管理员）
	}
}