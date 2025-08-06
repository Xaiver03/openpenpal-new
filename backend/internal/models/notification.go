package models

import (
	"time"
)

// NotificationStatus 通知状态
type NotificationStatus string

const (
	NotificationPending   NotificationStatus = "pending"   // 待发送
	NotificationSent      NotificationStatus = "sent"      // 已发送
	NotificationFailed    NotificationStatus = "failed"    // 发送失败
	NotificationRead      NotificationStatus = "read"      // 已读
	NotificationCancelled NotificationStatus = "cancelled" // 已取消
)

// NotificationType 通知类型
type NotificationType string

const (
	NotificationLetter     NotificationType = "letter"     // 信件通知
	NotificationCourier    NotificationType = "courier"    // 信使通知
	NotificationSystem     NotificationType = "system"     // 系统通知
	NotificationAccount    NotificationType = "account"    // 账户通知
	NotificationMuseum     NotificationType = "museum"     // 博物馆通知
	NotificationPromotion  NotificationType = "promotion"  // 推广通知
	NotificationModeration NotificationType = "moderation" // 审核通知
)

// NotificationChannel 通知渠道
type NotificationChannel string

const (
	ChannelWebSocket NotificationChannel = "websocket" // WebSocket实时通知
	ChannelEmail     NotificationChannel = "email"     // 邮件通知
	ChannelSMS       NotificationChannel = "sms"       // 短信通知
	ChannelPush      NotificationChannel = "push"      // 推送通知
)

// NotificationPriority 通知优先级
type NotificationPriority string

const (
	PriorityLow      NotificationPriority = "low"      // 低优先级
	PriorityNormal   NotificationPriority = "normal"   // 普通优先级
	PriorityHigh     NotificationPriority = "high"     // 高优先级
	PriorityCritical NotificationPriority = "critical" // 紧急优先级
)

// Notification 基础通知模型
type Notification struct {
	ID           string               `json:"id" gorm:"primaryKey;type:varchar(36)"`
	UserID       string               `json:"userId" gorm:"column:user_id;type:varchar(36);not null;index"`
	Type         NotificationType     `json:"type" gorm:"type:varchar(20);not null"`
	Channel      NotificationChannel  `json:"channel" gorm:"type:varchar(20);not null"`
	Priority     NotificationPriority `json:"priority" gorm:"type:varchar(20);default:'normal'"`
	Title        string               `json:"title" gorm:"type:varchar(200);not null"`
	Content      string               `json:"content" gorm:"type:text;not null"`
	Data         string               `json:"data" gorm:"type:text"` // JSON格式的额外数据
	Status       NotificationStatus   `json:"status" gorm:"type:varchar(20);default:'pending'"`
	ScheduledAt  *time.Time           `json:"scheduledAt" gorm:"column:scheduled_at"` // 定时发送
	SentAt       *time.Time           `json:"sentAt" gorm:"column:sent_at"`
	ReadAt       *time.Time           `json:"readAt" gorm:"column:read_at"`
	RetryCount   int                  `json:"retryCount" gorm:"column:retry_count;default:0"`
	MaxRetries   int                  `json:"maxRetries" gorm:"column:max_retries;default:3"`
	ErrorMessage string               `json:"errorMessage" gorm:"column:error_message;type:text"`
	CreatedAt    time.Time            `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt    time.Time            `json:"updatedAt" gorm:"column:updated_at"`
}

func (Notification) TableName() string {
	return "notifications"
}

// EmailTemplate 邮件模板
type EmailTemplate struct {
	ID           string               `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Name         string               `json:"name" gorm:"type:varchar(100);uniqueIndex;not null"`
	Type         NotificationType     `json:"type" gorm:"type:varchar(20);not null"`
	Subject      string               `json:"subject" gorm:"type:varchar(200);not null"`
	HTMLContent  string               `json:"htmlContent" gorm:"column:html_content;type:text;not null"`
	PlainContent string               `json:"plainContent" gorm:"column:plain_content;type:text"`
	Variables    string               `json:"variables" gorm:"type:text"` // JSON格式的模板变量说明
	IsActive     bool                 `json:"isActive" gorm:"column:is_active;default:true"`
	Priority     NotificationPriority `json:"priority" gorm:"type:varchar(20);default:'normal'"`
	CreatedBy    string               `json:"createdBy" gorm:"column:created_by;type:varchar(36)"`
	CreatedAt    time.Time            `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt    time.Time            `json:"updatedAt" gorm:"column:updated_at"`
}

func (EmailTemplate) TableName() string {
	return "email_templates"
}

// EmailLog 邮件发送日志
type EmailLog struct {
	ID             string             `json:"id" gorm:"primaryKey;type:varchar(36)"`
	NotificationID string             `json:"notificationId" gorm:"column:notification_id;type:varchar(36);index"`
	UserID         string             `json:"userId" gorm:"column:user_id;type:varchar(36);index"`
	ToEmail        string             `json:"toEmail" gorm:"column:to_email;type:varchar(255);not null"`
	FromEmail      string             `json:"fromEmail" gorm:"column:from_email;type:varchar(255)"`
	Subject        string             `json:"subject" gorm:"type:varchar(500);not null"`
	TemplateID     *string            `json:"templateId" gorm:"column:template_id;type:varchar(36)"`
	Provider       string             `json:"provider" gorm:"type:varchar(50)"` // smtp, sendgrid, etc
	Status         NotificationStatus `json:"status" gorm:"type:varchar(20);default:'pending'"`
	SentAt         *time.Time         `json:"sentAt" gorm:"column:sent_at"`
	DeliveredAt    *time.Time         `json:"deliveredAt" gorm:"column:delivered_at"`
	OpenedAt       *time.Time         `json:"openedAt" gorm:"column:opened_at"`
	ClickedAt      *time.Time         `json:"clickedAt" gorm:"column:clicked_at"`
	BouncedAt      *time.Time         `json:"bouncedAt" gorm:"column:bounced_at"`
	ErrorMessage   string             `json:"errorMessage" gorm:"column:error_message;type:text"`
	RetryCount     int                `json:"retryCount" gorm:"column:retry_count;default:0"`
	CreatedAt      time.Time          `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt      time.Time          `json:"updatedAt" gorm:"column:updated_at"`
}

func (EmailLog) TableName() string {
	return "email_logs"
}

// NotificationPreference 用户通知偏好设置
type NotificationPreference struct {
	ID           string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	UserID       string    `json:"userId" gorm:"column:user_id;type:varchar(36);uniqueIndex;not null"`
	EmailEnabled bool      `json:"emailEnabled" gorm:"column:email_enabled;default:true"`
	SMSEnabled   bool      `json:"smsEnabled" gorm:"column:sms_enabled;default:false"`
	PushEnabled  bool      `json:"pushEnabled" gorm:"column:push_enabled;default:true"`
	Types        string    `json:"types" gorm:"type:text"`                                // JSON格式，指定哪些类型的通知启用
	QuietHours   string    `json:"quietHours" gorm:"column:quiet_hours;type:varchar(50)"` // 例: "22:00-08:00"
	Frequency    string    `json:"frequency" gorm:"type:varchar(20);default:'realtime'"`  // realtime, daily, weekly
	Language     string    `json:"language" gorm:"type:varchar(10);default:'zh-CN'"`
	Timezone     string    `json:"timezone" gorm:"type:varchar(50);default:'Asia/Shanghai'"`
	CreatedAt    time.Time `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt    time.Time `json:"updatedAt" gorm:"column:updated_at"`
}

func (NotificationPreference) TableName() string {
	return "notification_preferences"
}

// NotificationBatch 批量通知任务
type NotificationBatch struct {
	ID               string              `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Name             string              `json:"name" gorm:"type:varchar(200);not null"`
	Type             NotificationType    `json:"type" gorm:"type:varchar(20);not null"`
	Channel          NotificationChannel `json:"channel" gorm:"type:varchar(20);not null"`
	TemplateID       *string             `json:"templateId" gorm:"column:template_id;type:varchar(36)"`
	TargetUsers      string              `json:"targetUsers" gorm:"column:target_users;type:text"`           // JSON格式的用户ID列表
	FilterConditions string              `json:"filterConditions" gorm:"column:filter_conditions;type:text"` // JSON格式的筛选条件
	TotalCount       int                 `json:"totalCount" gorm:"column:total_count;default:0"`
	SentCount        int                 `json:"sentCount" gorm:"column:sent_count;default:0"`
	FailedCount      int                 `json:"failedCount" gorm:"column:failed_count;default:0"`
	Status           string              `json:"status" gorm:"type:varchar(20);default:'preparing'"` // preparing, sending, completed, failed
	ScheduledAt      *time.Time          `json:"scheduledAt" gorm:"column:scheduled_at"`
	StartedAt        *time.Time          `json:"startedAt" gorm:"column:started_at"`
	CompletedAt      *time.Time          `json:"completedAt" gorm:"column:completed_at"`
	CreatedBy        string              `json:"createdBy" gorm:"column:created_by;type:varchar(36);not null"`
	CreatedAt        time.Time           `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt        time.Time           `json:"updatedAt" gorm:"column:updated_at"`
}

func (NotificationBatch) TableName() string {
	return "notification_batches"
}

// Request/Response DTOs

// SendNotificationRequest 发送通知请求
type SendNotificationRequest struct {
	UserIDs    []string               `json:"userIds" binding:"required"`
	Type       NotificationType       `json:"type" binding:"required"`
	Channel    NotificationChannel    `json:"channel" binding:"required"`
	Priority   NotificationPriority   `json:"priority"`
	Title      string                 `json:"title" binding:"required"`
	Content    string                 `json:"content" binding:"required"`
	Data       map[string]interface{} `json:"data"`
	ScheduleAt *time.Time             `json:"scheduleAt"` // 定时发送
	TemplateID *string                `json:"templateId"` // 使用模板
}

// NotificationListResponse 通知列表响应
type NotificationListResponse struct {
	Notifications []Notification `json:"notifications"`
	Total         int64          `json:"total"`
	Page          int            `json:"page"`
	PageSize      int            `json:"pageSize"`
	UnreadCount   int64          `json:"unreadCount"`
}

// EmailConfig 邮件配置
type EmailConfig struct {
	SMTPHost     string `json:"smtpHost"`
	SMTPPort     int    `json:"smtpPort"`
	SMTPUsername string `json:"smtpUsername"`
	SMTPPassword string `json:"-"` // 不返回密码
	FromEmail    string `json:"fromEmail"`
	FromName     string `json:"fromName"`
	Provider     string `json:"provider"` // smtp, sendgrid, mailgun等
	APIKey       string `json:"-"`        // 第三方服务API密钥
	IsActive     bool   `json:"isActive"`
}
