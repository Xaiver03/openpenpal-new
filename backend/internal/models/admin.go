package models

import "time"

// AdminDashboardStats 管理后台统计数据
type AdminDashboardStats struct {
	// 用户统计
	TotalUsers     int64 `json:"total_users"`
	NewUsersToday  int64 `json:"new_users_today"`
	ActiveCouriers int64 `json:"active_couriers"`

	// 信件统计
	TotalLetters             int64            `json:"total_letters"`
	LettersToday             int64            `json:"letters_today"`
	LetterStatusDistribution map[string]int64 `json:"letter_status_distribution"`

	// 博物馆统计
	MuseumExhibits int64 `json:"museum_exhibits"`

	// 信封统计
	EnvelopeOrders int64 `json:"envelope_orders"`

	// 通知统计
	TotalNotifications int64 `json:"total_notifications"`

	// 系统健康状态
	SystemHealth *SystemHealth `json:"system_health"`
}

// SystemHealth 系统健康状态
type SystemHealth struct {
	DatabaseStatus string    `json:"database_status"`
	ServiceStatus  string    `json:"service_status"`
	LastUpdated    time.Time `json:"last_updated"`
}

// AdminActivity 管理员活动记录
type AdminActivity struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"` // user_registered, letter_created, etc.
	Description string    `json:"description"`
	UserID      string    `json:"user_id"`
	CreatedAt   time.Time `json:"created_at"`
}

// UserManagementResponse 用户管理响应
type UserManagementResponse struct {
	Users []User `json:"users"`
	Total int64  `json:"total"`
	Page  int    `json:"page"`
	Limit int    `json:"limit"`
}

// AdminSystemSettings 管理后台系统设置 (简化版本)
type AdminSystemSettings struct {
	SiteName             string    `json:"site_name"`
	SiteDescription      string    `json:"site_description"`
	RegistrationOpen     bool      `json:"registration_open"`
	MaintenanceMode      bool      `json:"maintenance_mode"`
	MaxLettersPerDay     int       `json:"max_letters_per_day"`
	MaxEnvelopesPerOrder int       `json:"max_envelopes_per_order"`
	EmailEnabled         bool      `json:"email_enabled"`
	SMSEnabled           bool      `json:"sms_enabled"`
	LastUpdated          time.Time `json:"last_updated"`
}

// AdminDashboardRequest 管理后台请求参数
type AdminDashboardRequest struct {
	TimeRange string `json:"time_range" form:"time_range"` // today, week, month, year
	StartDate string `json:"start_date" form:"start_date"`
	EndDate   string `json:"end_date" form:"end_date"`
}

// ChartData 图表数据
type ChartData struct {
	Labels   []string  `json:"labels"`
	Datasets []Dataset `json:"datasets"`
}

// Dataset 数据集
type Dataset struct {
	Label           string    `json:"label"`
	Data            []float64 `json:"data"`
	BackgroundColor string    `json:"backgroundColor"`
	BorderColor     string    `json:"borderColor"`
}

// AnalyticsData 分析数据
type AnalyticsData struct {
	UserGrowth   *ChartData `json:"user_growth"`
	LetterTrends *ChartData `json:"letter_trends"`
	CourierStats *ChartData `json:"courier_stats"`
}
