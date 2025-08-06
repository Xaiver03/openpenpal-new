package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/smtp"
	"openpenpal-backend/internal/config"
	"openpenpal-backend/internal/models"
	"openpenpal-backend/internal/websocket"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// NotificationService 通知服务
type NotificationService struct {
	db        *gorm.DB
	config    *config.Config
	wsService *websocket.WebSocketService
}

// NewNotificationService 创建通知服务实例
func NewNotificationService(db *gorm.DB, config *config.Config) *NotificationService {
	return &NotificationService{
		db:     db,
		config: config,
	}
}

// SetWebSocketService 设置WebSocket服务（避免循环依赖）
func (s *NotificationService) SetWebSocketService(wsService *websocket.WebSocketService) {
	s.wsService = wsService
}

// NotifyUser 发送通知给用户
func (s *NotificationService) NotifyUser(userID string, notificationType string, data map[string]interface{}) error {
	// 获取用户通知偏好
	prefs, err := s.GetUserPreferences(userID)
	if err != nil {
		// 如果获取偏好失败，使用默认设置
		prefs = s.getDefaultPreferences(userID)
	}

	// 获取用户信息
	var user models.User
	if err := s.db.First(&user, "id = ?", userID).Error; err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// 构建通知内容
	title, content := s.buildNotificationContent(notificationType, data)

	// 根据偏好和类型选择发送渠道
	channels := s.determineChannels(prefs, notificationType)

	// 多渠道发送
	for _, channel := range channels {
		notification := &models.Notification{
			ID:        uuid.New().String(),
			UserID:    userID,
			Type:      models.NotificationType(notificationType),
			Channel:   channel,
			Priority:  s.determinePriority(notificationType),
			Title:     title,
			Content:   content,
			Data:      s.mapToJSON(data),
			Status:    models.NotificationPending,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// 保存通知记录
		if err := s.db.Create(notification).Error; err != nil {
			continue // 保存失败继续下一个渠道
		}

		// 根据渠道发送
		switch channel {
		case models.ChannelEmail:
			go s.sendEmailNotification(notification, &user)
		case models.ChannelWebSocket:
			go s.sendWebSocketNotification(notification, &user)
		default:
			// 标记为已处理（其他渠道暂不实现）
			notification.Status = models.NotificationSent
			s.db.Save(notification)
		}
	}

	return nil
}

// SendEmailNotification 发送邮件通知
func (s *NotificationService) SendEmailNotification(req *models.SendNotificationRequest) error {
	for _, userID := range req.UserIDs {
		// 获取用户信息
		var user models.User
		if err := s.db.First(&user, "id = ?", userID).Error; err != nil {
			continue // 用户不存在，跳过
		}

		// 检查用户邮件偏好
		prefs, _ := s.GetUserPreferences(userID)
		if prefs != nil && !prefs.EmailEnabled {
			continue // 用户禁用了邮件通知
		}

		// 创建通知记录
		notification := &models.Notification{
			ID:          uuid.New().String(),
			UserID:      userID,
			Type:        req.Type,
			Channel:     req.Channel,
			Priority:    req.Priority,
			Title:       req.Title,
			Content:     req.Content,
			Data:        s.mapToJSON(req.Data),
			Status:      models.NotificationPending,
			ScheduledAt: req.ScheduleAt,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if err := s.db.Create(notification).Error; err != nil {
			continue
		}

		// 立即发送或定时发送
		if req.ScheduleAt == nil || req.ScheduleAt.Before(time.Now()) {
			go s.sendEmailNotification(notification, &user)
		} else {
			// 定时任务（简化实现，实际可用定时任务系统）
			go func(n *models.Notification, u *models.User) {
				time.Sleep(time.Until(*req.ScheduleAt))
				s.sendEmailNotification(n, u)
			}(notification, &user)
		}
	}

	return nil
}

// sendEmailNotification 发送邮件通知（内部方法）
func (s *NotificationService) sendEmailNotification(notification *models.Notification, user *models.User) {
	// 创建邮件日志
	emailLog := &models.EmailLog{
		ID:             uuid.New().String(),
		NotificationID: notification.ID,
		UserID:         notification.UserID,
		ToEmail:        user.Email,
		FromEmail:      s.config.EmailFromAddress,
		Subject:        notification.Title,
		Provider:       "smtp",
		Status:         models.NotificationPending,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	s.db.Create(emailLog)

	// 发送邮件
	err := s.sendSMTPEmail(user.Email, notification.Title, notification.Content, user.Username)

	now := time.Now()
	if err != nil {
		// 发送失败
		notification.Status = models.NotificationFailed
		notification.ErrorMessage = err.Error()
		notification.RetryCount++

		emailLog.Status = models.NotificationFailed
		emailLog.ErrorMessage = err.Error()
		emailLog.RetryCount++
	} else {
		// 发送成功
		notification.Status = models.NotificationSent
		notification.SentAt = &now

		emailLog.Status = models.NotificationSent
		emailLog.SentAt = &now
	}

	notification.UpdatedAt = now
	emailLog.UpdatedAt = now

	s.db.Save(notification)
	s.db.Save(emailLog)
}

// sendWebSocketNotification 发送WebSocket通知（内部方法）
func (s *NotificationService) sendWebSocketNotification(notification *models.Notification, user *models.User) {
	now := time.Now()

	// 如果WebSocket服务可用，发送实时通知
	if s.wsService != nil {
		notificationData := &websocket.NotificationData{
			NotificationID: notification.ID,
			Title:          notification.Title,
			Content:        notification.Content,
			Type:           s.mapNotificationTypeToWebSocket(string(notification.Type)),
			Priority:       s.mapNotificationPriorityToWebSocket(notification.Priority),
			ActionURL:      "", // 可以根据通知类型生成相应的操作URL
			CreatedAt:      notification.CreatedAt,
			ExpiresAt:      nil, // 可以根据需要设置过期时间
		}

		s.wsService.BroadcastNotification(user.ID, notificationData)

		// 标记为已发送
		notification.Status = models.NotificationSent
		notification.SentAt = &now
	} else {
		// WebSocket服务不可用，标记为失败
		notification.Status = models.NotificationFailed
		notification.ErrorMessage = "WebSocket service not available"
	}

	notification.UpdatedAt = now
	s.db.Save(notification)
}

// sendSMTPEmail 通过SMTP发送邮件
func (s *NotificationService) sendSMTPEmail(toEmail, subject, content, userName string) error {
	if s.config.SMTPHost == "" {
		return errors.New("SMTP not configured")
	}

	// 构建邮件内容
	emailBody := s.buildEmailHTML(subject, content, userName)

	// SMTP配置
	auth := smtp.PlainAuth("", s.config.SMTPUsername, s.config.SMTPPassword, s.config.SMTPHost)

	// 构建邮件头
	headers := make(map[string]string)
	headers["From"] = fmt.Sprintf("%s <%s>", s.config.EmailFromName, s.config.EmailFromAddress)
	headers["To"] = toEmail
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	// 构建完整邮件
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + emailBody

	// 发送邮件
	addr := fmt.Sprintf("%s:%d", s.config.SMTPHost, s.config.SMTPPort)
	return smtp.SendMail(addr, auth, s.config.EmailFromAddress, []string{toEmail}, []byte(message))
}

// buildEmailHTML 构建HTML邮件内容
func (s *NotificationService) buildEmailHTML(subject, content, userName string) string {
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Subject}}</title>
    <style>
        body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background-color: #f4f4f4; }
        .container { max-width: 600px; margin: 0 auto; background-color: white; padding: 20px; border-radius: 10px; box-shadow: 0 0 10px rgba(0,0,0,0.1); }
        .header { text-align: center; padding: 20px 0; border-bottom: 2px solid #007bff; margin-bottom: 20px; }
        .logo { font-size: 24px; font-weight: bold; color: #007bff; }
        .content { padding: 20px 0; }
        .footer { text-align: center; padding: 20px 0; border-top: 1px solid #eee; margin-top: 20px; color: #666; font-size: 14px; }
        .button { display: inline-block; padding: 12px 24px; background-color: #007bff; color: white; text-decoration: none; border-radius: 5px; margin: 10px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div class="logo">✉️ OpenPenPal</div>
        </div>
        <div class="content">
            <h2>Hi {{.UserName}},</h2>
            <p>{{.Content}}</p>
        </div>
        <div class="footer">
            <p>此邮件由 OpenPenPal 系统自动发送，请勿回复。</p>
            <p>如有疑问，请联系我们的客服团队。</p>
            <p>&copy; 2024 OpenPenPal. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`
	t, _ := template.New("email").Parse(tmpl)
	var buf bytes.Buffer
	t.Execute(&buf, map[string]interface{}{
		"Subject":  subject,
		"Content":  content,
		"UserName": userName,
	})
	return buf.String()
}

// GetUserNotifications 获取用户通知列表
func (s *NotificationService) GetUserNotifications(userID string, page, pageSize int) (*models.NotificationListResponse, error) {
	var notifications []models.Notification
	var total int64
	var unreadCount int64

	// 计算偏移量
	offset := (page - 1) * pageSize

	// 获取总数
	s.db.Model(&models.Notification{}).Where("user_id = ?", userID).Count(&total)

	// 获取未读数
	s.db.Model(&models.Notification{}).Where("user_id = ? AND status != ?", userID, models.NotificationRead).Count(&unreadCount)

	// 获取通知列表
	err := s.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&notifications).Error

	if err != nil {
		return nil, err
	}

	return &models.NotificationListResponse{
		Notifications: notifications,
		Total:         total,
		Page:          page,
		PageSize:      pageSize,
		UnreadCount:   unreadCount,
	}, nil
}

// MarkAsRead 标记通知为已读
func (s *NotificationService) MarkAsRead(notificationID, userID string) error {
	now := time.Now()
	return s.db.Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Updates(map[string]interface{}{
			"status":     models.NotificationRead,
			"read_at":    &now,
			"updated_at": now,
		}).Error
}

// MarkAllAsRead 标记所有通知为已读
func (s *NotificationService) MarkAllAsRead(userID string) error {
	now := time.Now()
	return s.db.Model(&models.Notification{}).
		Where("user_id = ? AND status != ?", userID, models.NotificationRead).
		Updates(map[string]interface{}{
			"status":     models.NotificationRead,
			"read_at":    &now,
			"updated_at": now,
		}).Error
}

// GetUserPreferences 获取用户通知偏好
func (s *NotificationService) GetUserPreferences(userID string) (*models.NotificationPreference, error) {
	var prefs models.NotificationPreference
	err := s.db.Where("user_id = ?", userID).First(&prefs).Error
	if err != nil {
		return nil, err
	}
	return &prefs, nil
}

// UpdateUserPreferences 更新用户通知偏好
func (s *NotificationService) UpdateUserPreferences(userID string, prefs *models.NotificationPreference) error {
	prefs.UserID = userID
	prefs.UpdatedAt = time.Now()

	// 检查是否已存在
	var existing models.NotificationPreference
	err := s.db.Where("user_id = ?", userID).First(&existing).Error

	if err == nil {
		// 更新现有记录
		prefs.ID = existing.ID
		prefs.CreatedAt = existing.CreatedAt
		return s.db.Save(prefs).Error
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		// 创建新记录
		prefs.ID = uuid.New().String()
		prefs.CreatedAt = time.Now()
		return s.db.Create(prefs).Error
	}

	return err
}

// 辅助方法

// getDefaultPreferences 获取默认通知偏好
func (s *NotificationService) getDefaultPreferences(userID string) *models.NotificationPreference {
	return &models.NotificationPreference{
		ID:           uuid.New().String(),
		UserID:       userID,
		EmailEnabled: true,
		SMSEnabled:   false,
		PushEnabled:  true,
		Frequency:    "realtime",
		Language:     "zh-CN",
		Timezone:     "Asia/Shanghai",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

// buildNotificationContent 构建通知内容
func (s *NotificationService) buildNotificationContent(notificationType string, data map[string]interface{}) (string, string) {
	switch notificationType {
	case "letter_received":
		return "您有新的信件", "您收到了一封新的手写信件，请及时查看。"
	case "letter_delivered":
		return "信件已送达", "您的信件已成功送达收件人。"
	case "courier_assigned":
		return "信使已接单", "您的信件已被信使接单，正在配送中。"
	case "museum_approved":
		return "博物馆展品通过审核", "您提交的博物馆展品已通过审核，现已公开展示。"
	case "letter_reply_received":
		return "您有新的回信", "您的信件收到了回信，快去看看吧！"
	case "delivery_task_created":
		return "新配送任务", "系统为您创建了新的配送任务。"
	case "system_maintenance":
		return "系统维护通知", "系统将于指定时间进行维护，期间可能影响服务使用。"
	default:
		return "系统通知", "您有一条新的系统通知。"
	}
}

// determineChannels 根据偏好确定发送渠道
func (s *NotificationService) determineChannels(prefs *models.NotificationPreference, notificationType string) []models.NotificationChannel {
	var channels []models.NotificationChannel

	if prefs == nil {
		// 默认渠道
		return []models.NotificationChannel{models.ChannelWebSocket, models.ChannelEmail}
	}

	// 根据用户偏好选择渠道
	if prefs.EmailEnabled {
		channels = append(channels, models.ChannelEmail)
	}
	if prefs.PushEnabled {
		channels = append(channels, models.ChannelWebSocket)
	}
	if prefs.SMSEnabled {
		channels = append(channels, models.ChannelSMS)
	}

	if len(channels) == 0 {
		// 至少保证WebSocket通知
		channels = []models.NotificationChannel{models.ChannelWebSocket}
	}

	return channels
}

// determinePriority 根据通知类型确定优先级
func (s *NotificationService) determinePriority(notificationType string) models.NotificationPriority {
	switch notificationType {
	case "system_maintenance", "account_security":
		return models.PriorityCritical
	case "letter_received", "courier_assigned":
		return models.PriorityHigh
	case "letter_delivered", "museum_approved":
		return models.PriorityNormal
	default:
		return models.PriorityLow
	}
}

// mapToJSON 将map转换为JSON字符串
func (s *NotificationService) mapToJSON(data map[string]interface{}) string {
	if data == nil {
		return "{}"
	}
	jsonData, _ := json.Marshal(data)
	return string(jsonData)
}

// mapNotificationTypeToWebSocket 将通知类型映射到WebSocket格式
func (s *NotificationService) mapNotificationTypeToWebSocket(notificationType string) string {
	switch notificationType {
	case "letter_received", "letter_reply_received":
		return "success"
	case "letter_delivered":
		return "info"
	case "system_maintenance":
		return "warning"
	default:
		return "info"
	}
}

// mapNotificationPriorityToWebSocket 将通知优先级映射到WebSocket格式
func (s *NotificationService) mapNotificationPriorityToWebSocket(priority models.NotificationPriority) string {
	switch priority {
	case models.PriorityCritical:
		return "high"
	case models.PriorityHigh:
		return "high"
	case models.PriorityNormal:
		return "normal"
	case models.PriorityLow:
		return "low"
	default:
		return "normal"
	}
}
