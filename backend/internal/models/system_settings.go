package models

import (
	"time"
)

// SystemSettings 系统配置模型 - 对应 system_settings 表
type SystemSettings struct {
	ID        string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Key       string    `json:"key" gorm:"uniqueIndex;type:varchar(100);not null"`
	Value     string    `json:"value" gorm:"type:text"`
	Category  string    `json:"category" gorm:"type:varchar(50);index"`
	DataType  string    `json:"data_type" gorm:"type:varchar(20);default:'string'"` // string, number, boolean, json
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (SystemSettings) TableName() string {
	return "system_settings"
}

// SystemConfig 完整的系统配置结构体 - 匹配前端期望的格式
type SystemConfig struct {
	// 基本设置
	SiteName        string `json:"site_name"`
	SiteDescription string `json:"site_description"`
	SiteLogo        string `json:"site_logo"`
	MaintenanceMode bool   `json:"maintenance_mode"`

	// 邮件设置
	SMTPHost         string `json:"smtp_host"`
	SMTPPort         int    `json:"smtp_port"`
	SMTPUsername     string `json:"smtp_username"`
	SMTPPassword     string `json:"smtp_password"`
	SMTPEncryption   string `json:"smtp_encryption"` // tls, ssl, none
	EmailFromName    string `json:"email_from_name"`
	EmailFromAddress string `json:"email_from_address"`

	// 信件设置
	MaxLetterLength      int      `json:"max_letter_length"`
	AllowedFileTypes     []string `json:"allowed_file_types"`
	MaxFileSize          int      `json:"max_file_size"` // MB
	LetterReviewRequired bool     `json:"letter_review_required"`
	AutoDeliveryEnabled  bool     `json:"auto_delivery_enabled"`

	// 用户设置
	UserRegistrationEnabled   bool `json:"user_registration_enabled"`
	EmailVerificationRequired bool `json:"email_verification_required"`
	MaxUsersPerSchool         int  `json:"max_users_per_school"`
	UserInactiveDays          int  `json:"user_inactive_days"`

	// 信使设置
	CourierApplicationEnabled bool `json:"courier_application_enabled"`
	CourierAutoApproval       bool `json:"courier_auto_approval"`
	MaxDeliveryDistance       int  `json:"max_delivery_distance"` // km
	CourierRatingRequired     bool `json:"courier_rating_required"`

	// 安全设置
	PasswordMinLength      int  `json:"password_min_length"`
	PasswordRequireSymbols bool `json:"password_require_symbols"`
	PasswordRequireNumbers bool `json:"password_require_numbers"`
	SessionTimeout         int  `json:"session_timeout"` // seconds
	MaxLoginAttempts       int  `json:"max_login_attempts"`
	JWTExpiryHours         int  `json:"jwt_expiry_hours"`     // JWT过期时间（小时）
	RefreshTokenDays       int  `json:"refresh_token_days"`   // 刷新Token过期时间（天）
	EnableTokenRefresh     bool `json:"enable_token_refresh"` // 是否启用Token自动刷新

	// 通知设置
	EmailNotifications bool `json:"email_notifications"`
	SMSNotifications   bool `json:"sms_notifications"`
	PushNotifications  bool `json:"push_notifications"`
	AdminNotifications bool `json:"admin_notifications"`
}

// DefaultSystemConfig 返回默认系统配置
func DefaultSystemConfig() *SystemConfig {
	return &SystemConfig{
		// 基本设置
		SiteName:        "OpenPenPal",
		SiteDescription: "温暖的校园信件投递平台",
		SiteLogo:        "",
		MaintenanceMode: false,

		// 邮件设置
		SMTPHost:         "smtp.example.com",
		SMTPPort:         587,
		SMTPUsername:     "",
		SMTPPassword:     "",
		SMTPEncryption:   "tls",
		EmailFromName:    "OpenPenPal",
		EmailFromAddress: "noreply@openpenpal.com",

		// 信件设置
		MaxLetterLength:      5000,
		AllowedFileTypes:     []string{"jpg", "png", "pdf"},
		MaxFileSize:          10,
		LetterReviewRequired: false,
		AutoDeliveryEnabled:  true,

		// 用户设置
		UserRegistrationEnabled:   true,
		EmailVerificationRequired: true,
		MaxUsersPerSchool:         10000,
		UserInactiveDays:          90,

		// 信使设置
		CourierApplicationEnabled: true,
		CourierAutoApproval:       false,
		MaxDeliveryDistance:       10,
		CourierRatingRequired:     true,

		// 安全设置
		PasswordMinLength:      6,
		PasswordRequireSymbols: false,
		PasswordRequireNumbers: true,
		SessionTimeout:         3600,
		MaxLoginAttempts:       5,
		JWTExpiryHours:         24,   // 24小时
		RefreshTokenDays:       7,    // 7天
		EnableTokenRefresh:     true, // 启用自动刷新

		// 通知设置
		EmailNotifications: true,
		SMSNotifications:   false,
		PushNotifications:  true,
		AdminNotifications: true,
	}
}

// SystemConfigCategory 配置分类常量
const (
	CategoryGeneral      = "general"
	CategoryEmail        = "email"
	CategoryLetter       = "letter"
	CategoryUser         = "user"
	CategoryCourier      = "courier"
	CategorySecurity     = "security"
	CategoryNotification = "notification"
)

// SystemConfigKeys 系统配置键常量
const (
	// 基本设置
	KeySiteName        = "site_name"
	KeySiteDescription = "site_description"
	KeySiteLogo        = "site_logo"
	KeyMaintenanceMode = "maintenance_mode"

	// 邮件设置
	KeySMTPHost         = "smtp_host"
	KeySMTPPort         = "smtp_port"
	KeySMTPUsername     = "smtp_username"
	KeySMTPPassword     = "smtp_password"
	KeySMTPEncryption   = "smtp_encryption"
	KeyEmailFromName    = "email_from_name"
	KeyEmailFromAddress = "email_from_address"

	// 信件设置
	KeyMaxLetterLength      = "max_letter_length"
	KeyAllowedFileTypes     = "allowed_file_types"
	KeyMaxFileSize          = "max_file_size"
	KeyLetterReviewRequired = "letter_review_required"
	KeyAutoDeliveryEnabled  = "auto_delivery_enabled"

	// 用户设置
	KeyUserRegistrationEnabled   = "user_registration_enabled"
	KeyEmailVerificationRequired = "email_verification_required"
	KeyMaxUsersPerSchool         = "max_users_per_school"
	KeyUserInactiveDays          = "user_inactive_days"

	// 信使设置
	KeyCourierApplicationEnabled = "courier_application_enabled"
	KeyCourierAutoApproval       = "courier_auto_approval"
	KeyMaxDeliveryDistance       = "max_delivery_distance"
	KeyCourierRatingRequired     = "courier_rating_required"

	// 安全设置
	KeyPasswordMinLength      = "password_min_length"
	KeyPasswordRequireSymbols = "password_require_symbols"
	KeyPasswordRequireNumbers = "password_require_numbers"
	KeySessionTimeout         = "session_timeout"
	KeyMaxLoginAttempts       = "max_login_attempts"

	// 通知设置
	KeyEmailNotifications = "email_notifications"
	KeySMSNotifications   = "sms_notifications"
	KeyPushNotifications  = "push_notifications"
	KeyAdminNotifications = "admin_notifications"
)
