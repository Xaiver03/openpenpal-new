package models

import (
	"time"
)

// AnalyticsMetricType 分析指标类型
type AnalyticsMetricType string

const (
	MetricTypeUser        AnalyticsMetricType = "user"        // 用户指标
	MetricTypeLetter      AnalyticsMetricType = "letter"      // 信件指标
	MetricTypeCourier     AnalyticsMetricType = "courier"     // 信使指标
	MetricTypeSystem      AnalyticsMetricType = "system"      // 系统指标
	MetricTypeEngagement  AnalyticsMetricType = "engagement"  // 用户参与度
	MetricTypePerformance AnalyticsMetricType = "performance" // 性能指标
)

// AnalyticsGranularity 数据粒度
type AnalyticsGranularity string

const (
	GranularityHourly  AnalyticsGranularity = "hourly"  // 小时级
	GranularityDaily   AnalyticsGranularity = "daily"   // 日级
	GranularityWeekly  AnalyticsGranularity = "weekly"  // 周级
	GranularityMonthly AnalyticsGranularity = "monthly" // 月级
	GranularityYearly  AnalyticsGranularity = "yearly"  // 年级
)

// AnalyticsMetric 分析指标基础模型
type AnalyticsMetric struct {
	ID          string               `json:"id" gorm:"primaryKey;type:varchar(36)"`
	MetricType  AnalyticsMetricType  `json:"metricType" gorm:"column:metric_type;type:varchar(20);not null;index"`
	MetricName  string               `json:"metricName" gorm:"column:metric_name;type:varchar(100);not null"`
	Value       float64              `json:"value" gorm:"not null"`
	Unit        string               `json:"unit" gorm:"type:varchar(20)"`       // count, percentage, bytes, ms等
	Dimension   string               `json:"dimension" gorm:"type:varchar(100)"` // 维度信息，如用户ID、学校代码等
	Granularity AnalyticsGranularity `json:"granularity" gorm:"type:varchar(20);not null"`
	Timestamp   time.Time            `json:"timestamp" gorm:"not null;index"`
	Metadata    string               `json:"metadata" gorm:"type:text"` // JSON格式的额外数据
	CreatedAt   time.Time            `json:"createdAt" gorm:"column:created_at"`
}

func (AnalyticsMetric) TableName() string {
	return "analytics_metrics"
}

// UserAnalytics 用户分析数据
type UserAnalytics struct {
	ID              string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	UserID          string    `json:"userId" gorm:"column:user_id;type:varchar(36);not null;index"`
	Date            time.Time `json:"date" gorm:"type:date;not null;index"`
	LettersSent     int       `json:"lettersSent" gorm:"column:letters_sent;default:0"`
	LettersReceived int       `json:"lettersReceived" gorm:"column:letters_received;default:0"`
	LettersRead     int       `json:"lettersRead" gorm:"column:letters_read;default:0"`
	LoginCount      int       `json:"loginCount" gorm:"column:login_count;default:0"`
	SessionDuration int       `json:"sessionDuration" gorm:"column:session_duration;default:0"` // 秒
	CourierTasks    int       `json:"courierTasks" gorm:"column:courier_tasks;default:0"`
	MuseumVisits    int       `json:"museumVisits" gorm:"column:museum_visits;default:0"`
	EngagementScore float64   `json:"engagementScore" gorm:"column:engagement_score;default:0"`
	RetentionDays   int       `json:"retentionDays" gorm:"column:retention_days;default:0"` // 连续活跃天数
	CreatedAt       time.Time `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt       time.Time `json:"updatedAt" gorm:"column:updated_at"`
}

func (UserAnalytics) TableName() string {
	return "user_analytics"
}

// SystemAnalytics 系统分析数据
type SystemAnalytics struct {
	ID                    string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Date                  time.Time `json:"date" gorm:"type:date;not null;uniqueIndex"`
	ActiveUsers           int       `json:"activeUsers" gorm:"column:active_users;default:0"`
	NewUsers              int       `json:"newUsers" gorm:"column:new_users;default:0"`
	TotalUsers            int       `json:"totalUsers" gorm:"column:total_users;default:0"`
	LettersCreated        int       `json:"lettersCreated" gorm:"column:letters_created;default:0"`
	LettersDelivered      int       `json:"lettersDelivered" gorm:"column:letters_delivered;default:0"`
	CourierTasksCompleted int       `json:"courierTasksCompleted" gorm:"column:courier_tasks_completed;default:0"`
	MuseumItemsAdded      int       `json:"museumItemsAdded" gorm:"column:museum_items_added;default:0"`
	AvgResponseTime       float64   `json:"avgResponseTime" gorm:"column:avg_response_time;default:0"` // 毫秒
	ErrorRate             float64   `json:"errorRate" gorm:"column:error_rate;default:0"`              // 百分比
	ServerUptime          float64   `json:"serverUptime" gorm:"column:server_uptime;default:0"`        // 百分比
	CreatedAt             time.Time `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt             time.Time `json:"updatedAt" gorm:"column:updated_at"`
}

func (SystemAnalytics) TableName() string {
	return "system_analytics"
}

// PerformanceMetric 性能指标
type PerformanceMetric struct {
	ID           string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Endpoint     string    `json:"endpoint" gorm:"type:varchar(200);not null;index"`
	Method       string    `json:"method" gorm:"type:varchar(10);not null"`
	ResponseTime float64   `json:"responseTime" gorm:"column:response_time;not null"` // 毫秒
	StatusCode   int       `json:"statusCode" gorm:"column:status_code;not null"`
	UserAgent    string    `json:"userAgent" gorm:"column:user_agent;type:varchar(500)"`
	IPAddress    string    `json:"ipAddress" gorm:"column:ip_address;type:varchar(45)"`
	UserID       *string   `json:"userId" gorm:"column:user_id;type:varchar(36);index"`
	Timestamp    time.Time `json:"timestamp" gorm:"not null;index"`
	CreatedAt    time.Time `json:"createdAt" gorm:"column:created_at"`
}

func (PerformanceMetric) TableName() string {
	return "performance_metrics"
}

// AnalyticsReport 分析报告
type AnalyticsReport struct {
	ID          string               `json:"id" gorm:"primaryKey;type:varchar(36)"`
	ReportType  AnalyticsMetricType  `json:"reportType" gorm:"column:report_type;type:varchar(20);not null"`
	Title       string               `json:"title" gorm:"type:varchar(200);not null"`
	Description string               `json:"description" gorm:"type:text"`
	Granularity AnalyticsGranularity `json:"granularity" gorm:"type:varchar(20);not null"`
	StartDate   time.Time            `json:"startDate" gorm:"column:start_date;not null"`
	EndDate     time.Time            `json:"endDate" gorm:"column:end_date;not null"`
	Data        string               `json:"data" gorm:"type:text;not null"`                      // JSON格式的报告数据
	Status      string               `json:"status" gorm:"type:varchar(20);default:'generating'"` // generating, completed, failed
	GeneratedBy string               `json:"generatedBy" gorm:"column:generated_by;type:varchar(36)"`
	CreatedAt   time.Time            `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt   time.Time            `json:"updatedAt" gorm:"column:updated_at"`
}

func (AnalyticsReport) TableName() string {
	return "analytics_reports"
}

// Request/Response DTOs

// GenerateReportRequest 生成报告请求
type GenerateReportRequest struct {
	ReportType  AnalyticsMetricType    `json:"reportType" binding:"required"`
	Title       string                 `json:"title" binding:"required"`
	Description string                 `json:"description"`
	Granularity AnalyticsGranularity   `json:"granularity" binding:"required"`
	StartDate   time.Time              `json:"startDate" binding:"required"`
	EndDate     time.Time              `json:"endDate" binding:"required"`
	Filters     map[string]interface{} `json:"filters"` // 过滤条件
}

// AnalyticsQuery 分析查询请求
type AnalyticsQuery struct {
	MetricType  AnalyticsMetricType  `json:"metricType" form:"metricType"`
	Granularity AnalyticsGranularity `json:"granularity" form:"granularity"`
	StartDate   time.Time            `json:"startDate" form:"startDate"`
	EndDate     time.Time            `json:"endDate" form:"endDate"`
	Dimension   string               `json:"dimension" form:"dimension"`
	GroupBy     []string             `json:"groupBy" form:"groupBy"`
	Limit       int                  `json:"limit" form:"limit,default=100"`
}

// MetricSummary 指标摘要
type MetricSummary struct {
	MetricName string  `json:"metricName"`
	Total      float64 `json:"total"`
	Average    float64 `json:"average"`
	Maximum    float64 `json:"maximum"`
	Minimum    float64 `json:"minimum"`
	Count      int     `json:"count"`
	Unit       string  `json:"unit"`
}

// DashboardData 仪表板数据
type DashboardData struct {
	Overview struct {
		TotalUsers     int     `json:"totalUsers"`
		ActiveUsers    int     `json:"activeUsers"`
		TotalLetters   int     `json:"totalLetters"`
		LettersToday   int     `json:"lettersToday"`
		EngagementRate float64 `json:"engagementRate"`
		DeliveryRate   float64 `json:"deliveryRate"`
	} `json:"overview"`

	TrendData struct {
		UserGrowth      []DataPoint `json:"userGrowth"`
		LetterActivity  []DataPoint `json:"letterActivity"`
		CourierActivity []DataPoint `json:"courierActivity"`
		SystemHealth    []DataPoint `json:"systemHealth"`
	} `json:"trendData"`

	RecentActivity []ActivityItem `json:"recentActivity"`
	Alerts         []AlertItem    `json:"alerts"`
}

// DataPoint 数据点
type DataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
	Label     string    `json:"label"`
}

// ActivityItem 活动项目
type ActivityItem struct {
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	UserID      string                 `json:"userId"`
	Timestamp   time.Time              `json:"timestamp"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// AlertItem 警报项目
type AlertItem struct {
	Level        string    `json:"level"` // info, warning, error, critical
	Message      string    `json:"message"`
	Type         string    `json:"type"`
	Timestamp    time.Time `json:"timestamp"`
	Acknowledged bool      `json:"acknowledged"`
}
