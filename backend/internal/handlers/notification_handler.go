package handlers

import (
	"net/http"
	"openpenpal-backend/internal/middleware"
	"openpenpal-backend/internal/models"
	"openpenpal-backend/internal/services"
	"strconv"

	"github.com/gin-gonic/gin"
)

// NotificationHandler 通知处理器
type NotificationHandler struct {
	notificationService *services.NotificationService
}

// NewNotificationHandler 创建通知处理器实例
func NewNotificationHandler(notificationService *services.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
	}
}

// SendNotification 发送通知
// @Summary 发送通知
// @Description 发送单个或批量通知给用户
// @Tags notifications
// @Accept json
// @Produce json
// @Param request body models.SendNotificationRequest true "发送通知请求"
// @Success 200 {object} map[string]interface{} "成功响应"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/notifications/send [post]
func (h *NotificationHandler) SendNotification(c *gin.Context) {
	var req models.SendNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request parameters",
			"details": err.Error(),
		})
		return
	}

	// 验证必要字段
	if len(req.UserIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "userIds is required",
		})
		return
	}

	// 设置默认值
	if req.Priority == "" {
		req.Priority = models.PriorityNormal
	}

	// 根据渠道发送通知
	switch req.Channel {
	case models.ChannelEmail:
		err := h.notificationService.SendEmailNotification(&req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to send email notification",
				"details": err.Error(),
			})
			return
		}
	default:
		// 通用通知发送
		for _, userID := range req.UserIDs {
			data := req.Data
			if data == nil {
				data = make(map[string]interface{})
			}
			data["title"] = req.Title
			data["content"] = req.Content

			err := h.notificationService.NotifyUser(userID, string(req.Type), data)
			if err != nil {
				// 记录错误但继续处理其他用户
				continue
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Notification sent successfully",
		"count":   len(req.UserIDs),
	})
}

// GetUserNotifications 获取用户通知列表
// @Summary 获取用户通知列表
// @Description 分页获取用户的通知列表
// @Tags notifications
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(20)
// @Success 200 {object} models.NotificationListResponse "通知列表"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/notifications [get]
func (h *NotificationHandler) GetUserNotifications(c *gin.Context) {
	// 从JWT中获取用户ID
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	// 参数验证
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 获取通知列表
	response, err := h.notificationService.GetUserNotifications(userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get notifications",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// MarkNotificationAsRead 标记通知为已读
// @Summary 标记通知为已读
// @Description 标记单个通知为已读状态
// @Tags notifications
// @Accept json
// @Produce json
// @Param id path string true "通知ID"
// @Success 200 {object} map[string]interface{} "成功响应"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/notifications/{id}/read [post]
func (h *NotificationHandler) MarkNotificationAsRead(c *gin.Context) {
	notificationID := c.Param("id")
	if notificationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Notification ID is required",
		})
		return
	}

	// 从JWT中获取用户ID
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// 标记为已读
	err := h.notificationService.MarkAsRead(notificationID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to mark notification as read",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Notification marked as read",
	})
}

// MarkAllNotificationsAsRead 标记所有通知为已读
// @Summary 标记所有通知为已读
// @Description 标记用户的所有未读通知为已读状态
// @Tags notifications
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "成功响应"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/notifications/read-all [post]
func (h *NotificationHandler) MarkAllNotificationsAsRead(c *gin.Context) {
	// 从JWT中获取用户ID
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// 标记所有为已读
	err := h.notificationService.MarkAllAsRead(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to mark all notifications as read",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "All notifications marked as read",
	})
}

// GetUserPreferences 获取用户通知偏好
// @Summary 获取用户通知偏好
// @Description 获取用户的通知偏好设置
// @Tags notifications
// @Accept json
// @Produce json
// @Success 200 {object} models.NotificationPreference "用户通知偏好"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/notifications/preferences [get]
func (h *NotificationHandler) GetUserPreferences(c *gin.Context) {
	// 从JWT中获取用户ID
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// 获取用户偏好
	prefs, err := h.notificationService.GetUserPreferences(userID)
	if err != nil {
		// 如果没有设置偏好，返回默认值
		c.JSON(http.StatusOK, gin.H{
			"emailEnabled": true,
			"smsEnabled":   false,
			"pushEnabled":  true,
			"frequency":    "realtime",
			"language":     "zh-CN",
			"timezone":     "Asia/Shanghai",
		})
		return
	}

	c.JSON(http.StatusOK, prefs)
}

// UpdateUserPreferences 更新用户通知偏好
// @Summary 更新用户通知偏好
// @Description 更新用户的通知偏好设置
// @Tags notifications
// @Accept json
// @Produce json
// @Param preferences body models.NotificationPreference true "通知偏好设置"
// @Success 200 {object} map[string]interface{} "成功响应"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/notifications/preferences [put]
func (h *NotificationHandler) UpdateUserPreferences(c *gin.Context) {
	// 从JWT中获取用户ID
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	var prefs models.NotificationPreference
	if err := c.ShouldBindJSON(&prefs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request parameters",
			"details": err.Error(),
		})
		return
	}

	// 更新偏好
	err := h.notificationService.UpdateUserPreferences(userID, &prefs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update preferences",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Preferences updated successfully",
	})
}

// TestEmailNotification 测试邮件通知
// @Summary 测试邮件通知
// @Description 发送测试邮件通知（仅限开发环境）
// @Tags notifications
// @Accept json
// @Produce json
// @Param request body map[string]string true "测试邮件请求"
// @Success 200 {object} map[string]interface{} "成功响应"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/notifications/test-email [post]
func (h *NotificationHandler) TestEmailNotification(c *gin.Context) {
	// 从JWT中获取用户ID
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	var req struct {
		Subject string `json:"subject"`
		Content string `json:"content"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request parameters",
			"details": err.Error(),
		})
		return
	}

	// 设置默认值
	if req.Subject == "" {
		req.Subject = "OpenPenPal 测试邮件"
	}
	if req.Content == "" {
		req.Content = "这是一封来自 OpenPenPal 系统的测试邮件，用于验证邮件通知功能是否正常工作。"
	}

	// 发送测试邮件
	emailReq := &models.SendNotificationRequest{
		UserIDs:  []string{userID},
		Type:     models.NotificationSystem,
		Channel:  models.ChannelEmail,
		Priority: models.PriorityNormal,
		Title:    req.Subject,
		Content:  req.Content,
		Data: map[string]interface{}{
			"test": true,
		},
	}

	err := h.notificationService.SendEmailNotification(emailReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to send test email",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Test email sent successfully",
	})
}
