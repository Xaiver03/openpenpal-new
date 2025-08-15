package models

import (
	"time"
)

// ModerationStatus 审核状态
type ModerationStatus string

const (
	ModerationPending  ModerationStatus = "pending"  // 待审核
	ModerationApproved ModerationStatus = "approved" // 已通过
	ModerationRejected ModerationStatus = "rejected" // 已拒绝
	ModerationReview   ModerationStatus = "review"   // 需人工复审
)

// ContentType 内容类型
type ContentType string

const (
	ContentTypeLetter   ContentType = "letter"   // 信件内容
	ContentTypeProfile  ContentType = "profile"  // 用户资料
	ContentTypePhoto    ContentType = "photo"    // 照片
	ContentTypeMuseum   ContentType = "museum"   // 博物馆内容
	ContentTypeEnvelope ContentType = "envelope" // 信封设计
	ContentTypeComment  ContentType = "comment"  // 评论内容
)

// ModerationLevel 审核等级
type ModerationLevel string

const (
	LevelLow    ModerationLevel = "low"    // 低风险
	LevelMedium ModerationLevel = "medium" // 中风险
	LevelHigh   ModerationLevel = "high"   // 高风险
	LevelBlock  ModerationLevel = "block"  // 需要屏蔽
)

// ModerationRecord 审核记录
type ModerationRecord struct {
	ID            string           `json:"id" gorm:"primaryKey;type:varchar(36)"`
	ContentType   ContentType      `json:"content_type" gorm:"type:varchar(20);not null;index"`
	ContentID     string           `json:"content_id" gorm:"type:varchar(36);not null;index"`
	UserID        string           `json:"user_id" gorm:"type:varchar(36);not null;index"`
	Content       string           `json:"content" gorm:"type:text"`
	ImageURLs     string           `json:"image_urls" gorm:"type:text"`
	Status        ModerationStatus `json:"status" gorm:"type:varchar(20);not null;default:'pending'"`
	Level         ModerationLevel  `json:"level" gorm:"type:varchar(20)"`
	Score         float64          `json:"score" gorm:"default:0"` // 风险分数 0-1
	Reasons       string           `json:"reasons" gorm:"type:text"`
	Categories    string           `json:"categories" gorm:"type:text"`
	AIProvider    string           `json:"ai_provider" gorm:"type:varchar(20)"`
	AIResponse    string           `json:"ai_response" gorm:"type:text"`
	ReviewerID    *string          `json:"reviewer_id" gorm:"type:varchar(36)"`
	ReviewNote    string           `json:"review_note" gorm:"type:text"`
	ReviewedAt    *time.Time       `json:"reviewed_at"`
	AutoModerated bool             `json:"auto_moderated" gorm:"default:true"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
}

// SensitiveWord 敏感词库
type SensitiveWord struct {
	ID        string          `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Word      string          `json:"word" gorm:"type:varchar(100);not null;uniqueIndex"`
	Category  string          `json:"category" gorm:"type:varchar(50)"` // 政治、色情、暴力、广告等
	Level     ModerationLevel `json:"level" gorm:"type:varchar(20);not null"`
	IsActive  bool            `json:"is_active" gorm:"default:true"`
	Reason    string          `json:"reason" gorm:"type:text"`         // 添加原因说明
	CreatedBy string          `json:"created_by" gorm:"type:varchar(36)"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"` // 添加更新时间
}

// ModerationRule 审核规则
type ModerationRule struct {
	ID          string      `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Name        string      `json:"name" gorm:"type:varchar(100);not null"`
	Description string      `json:"description" gorm:"type:text"`
	ContentType ContentType `json:"content_type" gorm:"type:varchar(20);not null"`
	RuleType    string      `json:"rule_type" gorm:"type:varchar(50)"` // keyword, regex, ai
	Pattern     string      `json:"pattern" gorm:"type:text"`
	Action      string      `json:"action" gorm:"type:varchar(20)"` // block, review, pass
	Priority    int         `json:"priority" gorm:"default:0"`
	IsActive    bool        `json:"is_active" gorm:"default:true"`
	CreatedBy   string      `json:"created_by" gorm:"type:varchar(36)"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// ModerationQueue 审核队列
type ModerationQueue struct {
	ID         string           `json:"id" gorm:"primaryKey;type:varchar(36)"`
	RecordID   string           `json:"record_id" gorm:"type:varchar(36);not null;index"`
	Record     ModerationRecord `json:"record" gorm:"foreignKey:RecordID"`
	Priority   int              `json:"priority" gorm:"default:0"` // 优先级，越大越优先
	AssignedTo *string          `json:"assigned_to" gorm:"type:varchar(36)"`
	AssignedAt *time.Time       `json:"assigned_at"`
	Status     string           `json:"status" gorm:"type:varchar(20);default:'pending'"` // pending, processing, completed
	CreatedAt  time.Time        `json:"created_at"`
}

// ModerationStats 审核统计
type ModerationStats struct {
	ID                string      `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Date              time.Time   `json:"date" gorm:"type:date;not null;index"`
	ContentType       ContentType `json:"content_type" gorm:"type:varchar(20);not null"`
	TotalCount        int64       `json:"total_count" gorm:"default:0"`
	ApprovedCount     int64       `json:"approved_count" gorm:"default:0"`
	RejectedCount     int64       `json:"rejected_count" gorm:"default:0"`
	ReviewCount       int64       `json:"review_count" gorm:"default:0"`
	AutoModerateCount int64       `json:"auto_moderate_count" gorm:"default:0"`
	AvgProcessTime    float64     `json:"avg_process_time" gorm:"default:0"` // 平均处理时间（秒）
	CreatedAt         time.Time   `json:"created_at"`
}

// Request/Response DTOs

// ModerationRequest 审核请求
type ModerationRequest struct {
	ContentType ContentType `json:"content_type" binding:"required"`
	ContentID   string      `json:"content_id" binding:"required"`
	Content     string      `json:"content"`
	ImageURLs   []string    `json:"image_urls"`
	UserID      string      `json:"user_id"`
}

// ModerationResponse 审核响应
type ModerationResponse struct {
	ID         string           `json:"id"`
	Status     ModerationStatus `json:"status"`
	Level      ModerationLevel  `json:"level"`
	Score      float64          `json:"score"`
	Reasons    []string         `json:"reasons"`
	Categories []string         `json:"categories"`
	NeedReview bool             `json:"need_review"`
}

// ReviewRequest 人工审核请求
type ReviewRequest struct {
	RecordID   string           `json:"record_id" binding:"required"`
	Status     ModerationStatus `json:"status" binding:"required"`
	ReviewNote string           `json:"review_note"`
}

// SensitiveWordRequest 敏感词请求
type SensitiveWordRequest struct {
	Word     string          `json:"word" binding:"required"`
	Category string          `json:"category"`
	Level    ModerationLevel `json:"level" binding:"required"`
}

// ModerationRuleRequest 审核规则请求
type ModerationRuleRequest struct {
	Name        string      `json:"name" binding:"required"`
	Description string      `json:"description"`
	ContentType ContentType `json:"content_type" binding:"required"`
	RuleType    string      `json:"rule_type" binding:"required"`
	Pattern     string      `json:"pattern" binding:"required"`
	Action      string      `json:"action" binding:"required"`
	Priority    int         `json:"priority"`
}

// =============== 新增：安全事件和统计模型 ===============

// SecurityEvent 安全事件记录 - 用于XSS攻击等安全事件记录
type SecurityEvent struct {
	ID          string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	UserID      string    `json:"user_id" gorm:"type:varchar(36);not null;index"`
	EventType   string    `json:"event_type" gorm:"type:varchar(50);not null"` // content_security_violation, xss_attempt, etc.
	ContentType string    `json:"content_type" gorm:"type:varchar(20)"`         // comment, letter, etc.
	Content     string    `json:"content" gorm:"type:text"`                     // 触发事件的内容
	RiskScore   int       `json:"risk_score" gorm:"default:0"`                  // 风险分数 0-100
	Violations  int       `json:"violations" gorm:"default:0"`                  // 违规数量
	IPAddress   string    `json:"ip_address" gorm:"type:varchar(45)"`           // IP地址
	UserAgent   string    `json:"user_agent" gorm:"type:varchar(500)"`          // 用户代理
	Details     string    `json:"details" gorm:"type:text"`                     // JSON格式的详细信息
	Handled     bool      `json:"handled" gorm:"default:false"`                 // 是否已处理
	HandledBy   *string   `json:"handled_by" gorm:"type:varchar(36)"`           // 处理人ID
	HandledAt   *time.Time `json:"handled_at"`                                  // 处理时间
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// SecurityStats 安全统计信息
type SecurityStats struct {
	TotalEvents       int64 `json:"total_events"`
	HighRiskEvents    int64 `json:"high_risk_events"`
	BlockedContent    int64 `json:"blocked_content"`
	AverageRiskScore  int   `json:"average_risk_score"`
	XSSAttempts       int64 `json:"xss_attempts"`
	PendingReview     int64 `json:"pending_review"`
	AutoBlocked       int64 `json:"auto_blocked"`
	ManualReviews     int64 `json:"manual_reviews"`
}
