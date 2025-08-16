package models

import (
	"time"
)

// CreditLimitRule 积分限制规则
type CreditLimitRule struct {
	ID          string              `json:"id" gorm:"primaryKey;type:varchar(36)"`
	ActionType  string              `json:"action_type" gorm:"not null;index"` // 行为类型 (letter_created, letter_replied, etc.)
	LimitType   CreditLimitType     `json:"limit_type" gorm:"not null"`        // 限制类型
	LimitPeriod CreditLimitPeriod   `json:"limit_period" gorm:"not null"`      // 限制周期
	MaxCount    int                 `json:"max_count" gorm:"not null"`         // 最大次数
	MaxPoints   int                 `json:"max_points"`                        // 最大积分 (可选)
	Enabled     bool                `json:"enabled" gorm:"default:true"`       // 是否启用
	Priority    int                 `json:"priority" gorm:"default:100"`       // 优先级 (数字越小优先级越高)
	Description string              `json:"description"`                       // 规则描述
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
}

// CreditLimitType 限制类型
type CreditLimitType string

const (
	LimitTypeCount      CreditLimitType = "count"      // 按次数限制
	LimitTypePoints     CreditLimitType = "points"     // 按积分限制
	LimitTypeCombined   CreditLimitType = "combined"   // 按次数和积分同时限制
)

// CreditLimitPeriod 限制周期
type CreditLimitPeriod string

const (
	LimitPeriodDaily   CreditLimitPeriod = "daily"   // 每日
	LimitPeriodWeekly  CreditLimitPeriod = "weekly"  // 每周
	LimitPeriodMonthly CreditLimitPeriod = "monthly" // 每月
)

// UserCreditAction 用户积分行为记录
type UserCreditAction struct {
	ID         string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	UserID     string    `json:"user_id" gorm:"not null;index"`
	ActionType string    `json:"action_type" gorm:"not null;index"`
	Points     int       `json:"points" gorm:"not null"`
	IPAddress  string    `json:"ip_address" gorm:"type:varchar(45)"`
	DeviceID   string    `json:"device_id" gorm:"type:varchar(100)"`
	UserAgent  string    `json:"user_agent"`
	Reference  string    `json:"reference" gorm:"index"` // 关联的业务ID
	CreatedAt  time.Time `json:"created_at"`
}

// CreditRiskUser 风险用户
type CreditRiskUser struct {
	UserID      string           `json:"user_id" gorm:"primaryKey;type:varchar(36)"`
	RiskScore   float64          `json:"risk_score" gorm:"type:decimal(5,2);default:0"`
	RiskLevel   CreditRiskLevel  `json:"risk_level" gorm:"default:'low'"`
	BlockedUntil *time.Time      `json:"blocked_until"`
	Reason      string           `json:"reason"`
	Notes       string           `json:"notes"`
	LastAlertAt *time.Time       `json:"last_alert_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
	CreatedAt   time.Time        `json:"created_at"`
}

// CreditRiskLevel 风险等级
type CreditRiskLevel string

// RiskLevel 风险等级别名（兼容性）
type RiskLevel = CreditRiskLevel

const (
	RiskLevelLow     CreditRiskLevel = "low"     // 低风险
	RiskLevelMedium  CreditRiskLevel = "medium"  // 中风险
	RiskLevelHigh    CreditRiskLevel = "high"    // 高风险
	RiskLevelBlocked CreditRiskLevel = "blocked" // 已封禁
)

// LimitStatus 限制状态
type LimitStatus struct {
	ActionType    string `json:"action_type"`
	Period        string `json:"period"`        // daily/weekly/monthly
	CurrentCount  int    `json:"current_count"` // 当前次数
	MaxCount      int    `json:"max_count"`     // 最大次数
	CurrentPoints int    `json:"current_points"` // 当前积分
	MaxPoints     int    `json:"max_points"`     // 最大积分
	Allowed       bool   `json:"allowed"`       // 是否允许操作（与IsLimited相反）
	IsLimited     bool   `json:"is_limited"`    // 是否已达限制
	Reason        string `json:"reason"`        // 限制原因
	ResetAt       time.Time `json:"reset_at"`   // 重置时间
}

// FraudAlert 作弊警报
type FraudAlert struct {
	UserID      string             `json:"user_id"`
	AlertType   FraudAlertType     `json:"alert_type"`
	Severity    AlertSeverity      `json:"severity"`
	Description string             `json:"description"`
	Evidence    map[string]interface{} `json:"evidence"`
	CreatedAt   time.Time          `json:"created_at"`
}

// FraudDetectionLog Phase 1.3: 防作弊检测日志
type FraudDetectionLog struct {
	ID              string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	UserID          string    `json:"user_id" gorm:"not null;index"`
	ActionType      string    `json:"action_type" gorm:"not null"`
	RiskScore       float64   `json:"risk_score" gorm:"not null"`
	IsAnomalous     bool      `json:"is_anomalous" gorm:"not null"`
	DetectedPatterns string   `json:"detected_patterns" gorm:"type:text"` // JSON array of patterns
	Evidence        string    `json:"evidence" gorm:"type:text"`          // JSON object
	Recommendations string    `json:"recommendations" gorm:"type:text"`   // JSON array
	AlertCount      int       `json:"alert_count" gorm:"default:0"`
	CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// FraudAlertType 作弊警报类型
type FraudAlertType string

const (
	AlertTypeFrequency     FraudAlertType = "frequency"      // 频率异常
	AlertTypePattern       FraudAlertType = "pattern"        // 行为模式异常
	AlertTypeIP            FraudAlertType = "ip"             // IP异常
	AlertTypeDevice        FraudAlertType = "device"         // 设备异常
	AlertTypePoints        FraudAlertType = "points"         // 积分异常
)

// AlertSeverity 警报严重程度
type AlertSeverity string

const (
	SeverityLow      AlertSeverity = "low"      // 低
	SeverityMedium   AlertSeverity = "medium"   // 中
	SeverityHigh     AlertSeverity = "high"     // 高
	SeverityCritical AlertSeverity = "critical" // 严重
)

// TableName 指定表名
func (CreditLimitRule) TableName() string {
	return "credit_limit_rules"
}

func (UserCreditAction) TableName() string {
	return "user_credit_actions"
}

func (CreditRiskUser) TableName() string {
	return "credit_risk_users"
}

func (FraudDetectionLog) TableName() string {
	return "fraud_detection_logs"
}

// GetPeriodStart 获取周期开始时间
func (p CreditLimitPeriod) GetPeriodStart(now time.Time) time.Time {
	switch p {
	case LimitPeriodDaily:
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	case LimitPeriodWeekly:
		// 以周一为开始
		weekday := int(now.Weekday())
		if weekday == 0 { // 周日
			weekday = 7
		}
		daysBack := weekday - 1
		startDate := now.AddDate(0, 0, -daysBack)
		return time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, now.Location())
	case LimitPeriodMonthly:
		return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	default:
		return now
	}
}

// GetPeriodEnd 获取周期结束时间
func (p CreditLimitPeriod) GetPeriodEnd(now time.Time) time.Time {
	start := p.GetPeriodStart(now)
	switch p {
	case LimitPeriodDaily:
		return start.AddDate(0, 0, 1)
	case LimitPeriodWeekly:
		return start.AddDate(0, 0, 7)
	case LimitPeriodMonthly:
		return start.AddDate(0, 1, 0)
	default:
		return now
	}
}

// IsLimitExceeded 检查是否超过限制
func (r *CreditLimitRule) IsLimitExceeded(currentCount, currentPoints int) bool {
	if !r.Enabled {
		return false
	}

	switch r.LimitType {
	case LimitTypeCount:
		return currentCount >= r.MaxCount
	case LimitTypePoints:
		return currentPoints >= r.MaxPoints
	case LimitTypeCombined:
		return currentCount >= r.MaxCount || currentPoints >= r.MaxPoints
	default:
		return false
	}
}

// CalculateRiskScore 计算风险分数
func (u *CreditRiskUser) CalculateRiskScore() float64 {
	// 基础分数
	score := u.RiskScore
	
	// 根据风险等级调整
	switch u.RiskLevel {
	case RiskLevelLow:
		score = score * 0.8
	case RiskLevelMedium:
		score = score * 1.2
	case RiskLevelHigh:
		score = score * 1.5
	case RiskLevelBlocked:
		score = 1.0
	}
	
	// 确保分数在0-1范围内
	if score > 1.0 {
		score = 1.0
	}
	if score < 0.0 {
		score = 0.0
	}
	
	return score
}

// IsBlocked 检查用户是否被封禁
func (u *CreditRiskUser) IsBlocked() bool {
	if u.RiskLevel == RiskLevelBlocked {
		return true
	}
	
	if u.BlockedUntil != nil && time.Now().Before(*u.BlockedUntil) {
		return true
	}
	
	return false
}

// GetRiskThreshold 获取风险阈值
func GetRiskThreshold(level CreditRiskLevel) float64 {
	switch level {
	case RiskLevelLow:
		return 0.3
	case RiskLevelMedium:
		return 0.6
	case RiskLevelHigh:
		return 0.8
	case RiskLevelBlocked:
		return 1.0
	default:
		return 0.0
	}
}