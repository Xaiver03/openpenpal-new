package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// CreditActivityType 积分活动类型
type CreditActivityType string

const (
	CreditActivityTypeDaily       CreditActivityType = "daily"         // 每日活动
	CreditActivityTypeWeekly      CreditActivityType = "weekly"        // 每周活动
	CreditActivityTypeMonthly     CreditActivityType = "monthly"       // 每月活动
	CreditActivityTypeSeasonal    CreditActivityType = "seasonal"      // 季节活动
	CreditActivityTypeSpecial     CreditActivityType = "special"       // 特殊活动
	CreditActivityTypeFirstTime   CreditActivityType = "first_time"    // 首次活动
	CreditActivityTypeCumulative  CreditActivityType = "cumulative"    // 累计活动
	CreditActivityTypeTimeLimited CreditActivityType = "time_limited"  // 限时活动
)

// CreditActivityStatus 积分活动状态
type CreditActivityStatus string

const (
	CreditActivityStatusDraft     CreditActivityStatus = "draft"       // 草稿
	CreditActivityStatusPending   CreditActivityStatus = "pending"     // 待开始
	CreditActivityStatusActive    CreditActivityStatus = "active"      // 进行中
	CreditActivityStatusPaused    CreditActivityStatus = "paused"      // 已暂停
	CreditActivityStatusCompleted CreditActivityStatus = "completed"   // 已结束
	CreditActivityStatusCancelled CreditActivityStatus = "cancelled"   // 已取消
)

// CreditActivityTargetType 活动目标类型
type CreditActivityTargetType string

const (
	CreditActivityTargetAll       CreditActivityTargetType = "all"        // 所有用户
	CreditActivityTargetNewUsers  CreditActivityTargetType = "new_users"  // 新用户
	CreditActivityTargetLevel     CreditActivityTargetType = "level"      // 指定等级
	CreditActivityTargetSchool    CreditActivityTargetType = "school"     // 指定学校
	CreditActivityTargetCustom    CreditActivityTargetType = "custom"     // 自定义
)

// CreditActivityTriggerType 活动触发类型
type CreditActivityTriggerType string

const (
	CreditActivityTriggerLogin         CreditActivityTriggerType = "login"          // 登录
	CreditActivityTriggerLetter        CreditActivityTriggerType = "letter"         // 写信
	CreditActivityTriggerReceive       CreditActivityTriggerType = "receive"        // 收信
	CreditActivityTriggerMuseum        CreditActivityTriggerType = "museum"         // 博物馆投稿
	CreditActivityTriggerRedemption    CreditActivityTriggerType = "redemption"     // 商城兑换
	CreditActivityTriggerInvite        CreditActivityTriggerType = "invite"         // 邀请好友
	CreditActivityTriggerConsecutive   CreditActivityTriggerType = "consecutive"    // 连续行为
	CreditActivityTriggerCumulative    CreditActivityTriggerType = "cumulative"     // 累计行为
	CreditActivityTriggerScheduled     CreditActivityTriggerType = "scheduled"      // 定时触发
)

// CreditActivity 积分活动模型
type CreditActivity struct {
	ID               uuid.UUID                  `gorm:"type:uuid;primary_key" json:"id"`
	Name             string                     `gorm:"type:varchar(200);not null" json:"name"`
	Description      string                     `gorm:"type:text" json:"description"`
	ActivityType     CreditActivityType         `gorm:"type:varchar(50);not null" json:"activity_type"`
	Status           CreditActivityStatus       `gorm:"type:varchar(20);default:'draft'" json:"status"`
	TargetType       CreditActivityTargetType   `gorm:"type:varchar(50);not null" json:"target_type"`
	TargetRules      datatypes.JSON             `gorm:"type:jsonb" json:"target_rules"`             // 目标规则（如等级范围、学校列表等）
	TriggerType      CreditActivityTriggerType  `gorm:"type:varchar(50);not null" json:"trigger_type"`
	TriggerRules     datatypes.JSON             `gorm:"type:jsonb" json:"trigger_rules"`             // 触发规则（如连续天数、累计次数等）
	RewardCredits    int                        `gorm:"type:int;not null" json:"reward_credits"`     // 奖励积分数
	RewardRules      datatypes.JSON             `gorm:"type:jsonb" json:"reward_rules"`              // 奖励规则（如递增、翻倍等）
	MaxParticipants  int                        `gorm:"type:int;default:0" json:"max_participants"`  // 最大参与人数，0表示不限
	MaxRewardsPerUser int                       `gorm:"type:int;default:1" json:"max_rewards_per_user"` // 每用户最大奖励次数
	Budget           int                        `gorm:"type:int;default:0" json:"budget"`            // 活动预算（总积分）
	ConsumedBudget   int                        `gorm:"type:int;default:0" json:"consumed_budget"`   // 已消耗预算
	Priority         int                        `gorm:"type:int;default:0" json:"priority"`          // 优先级
	StartTime        time.Time                  `json:"start_time"`
	EndTime          time.Time                  `json:"end_time"`
	RepeatPattern    string                     `gorm:"type:varchar(100)" json:"repeat_pattern"`     // 重复模式（如 daily, weekly 等）
	RepeatInterval   int                        `gorm:"type:int;default:0" json:"repeat_interval"`   // 重复间隔
	RepeatEndDate    *time.Time                 `json:"repeat_end_date,omitempty"`                   // 重复结束日期
	RequireVerification bool                    `gorm:"default:false" json:"require_verification"`    // 是否需要验证
	VerificationRules datatypes.JSON            `gorm:"type:jsonb" json:"verification_rules"`         // 验证规则
	DisplaySettings  datatypes.JSON             `gorm:"type:jsonb" json:"display_settings"`          // 显示设置（图标、颜色等）
	CreatedBy        string                     `gorm:"type:varchar(36)" json:"created_by"`
	UpdatedBy        string                     `gorm:"type:varchar(36)" json:"updated_by"`
	CreatedAt        time.Time                  `json:"created_at"`
	UpdatedAt        time.Time                  `json:"updated_at"`
	DeletedAt        *time.Time                 `gorm:"index" json:"deleted_at,omitempty"`
}

func (a *CreditActivity) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

// IsActive 检查活动是否进行中
func (a *CreditActivity) IsActive() bool {
	now := time.Now()
	return a.Status == CreditActivityStatusActive &&
		now.After(a.StartTime) &&
		now.Before(a.EndTime)
}

// CreditActivityParticipation 活动参与记录
type CreditActivityParticipation struct {
	ID                  uuid.UUID      `gorm:"type:uuid;primary_key" json:"id"`
	ActivityID          uuid.UUID      `gorm:"type:uuid;not null;index" json:"activity_id"`
	Activity            *CreditActivity `gorm:"foreignKey:ActivityID" json:"activity,omitempty"`
	UserID              string         `gorm:"type:varchar(36);not null;index" json:"user_id"`
	User                *User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	ParticipatedAt      time.Time      `json:"participated_at"`
	CompletedAt         *time.Time     `json:"completed_at,omitempty"`
	Progress            int            `gorm:"type:int;default:0" json:"progress"`             // 进度（百分比或计数）
	ProgressDetails     datatypes.JSON `gorm:"type:jsonb" json:"progress_details"`             // 详细进度信息
	RewardCredits       int            `gorm:"type:int;default:0" json:"reward_credits"`       // 获得的积分
	RewardDetails       datatypes.JSON `gorm:"type:jsonb" json:"reward_details"`               // 奖励详情
	VerificationStatus  string         `gorm:"type:varchar(20)" json:"verification_status"`    // 验证状态
	VerificationDetails datatypes.JSON `gorm:"type:jsonb" json:"verification_details"`         // 验证详情
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`

	// 添加唯一索引，确保用户在同一活动中只有一条记录
	// 在数据库中通过 UNIQUE INDEX 实现
}

func (p *CreditActivityParticipation) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	if p.ParticipatedAt.IsZero() {
		p.ParticipatedAt = time.Now()
	}
	return nil
}

// CreditActivityRule 活动规则模型（用于规则引擎）
type CreditActivityRule struct {
	ID              uuid.UUID      `gorm:"type:uuid;primary_key" json:"id"`
	ActivityID      uuid.UUID      `gorm:"type:uuid;not null;index" json:"activity_id"`
	Activity        *CreditActivity `gorm:"foreignKey:ActivityID" json:"activity,omitempty"`
	RuleType        string         `gorm:"type:varchar(50);not null" json:"rule_type"`      // 规则类型
	RuleName        string         `gorm:"type:varchar(100);not null" json:"rule_name"`
	RuleCondition   datatypes.JSON `gorm:"type:jsonb;not null" json:"rule_condition"`       // 规则条件
	RuleAction      datatypes.JSON `gorm:"type:jsonb;not null" json:"rule_action"`          // 规则动作
	Priority        int            `gorm:"type:int;default:0" json:"priority"`
	IsActive        bool           `gorm:"default:true" json:"is_active"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

func (r *CreditActivityRule) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

// CreditActivitySchedule 活动调度记录
type CreditActivitySchedule struct {
	ID                uuid.UUID       `gorm:"type:uuid;primary_key" json:"id"`
	ActivityID        uuid.UUID       `gorm:"type:uuid;not null;index" json:"activity_id"`
	Activity          *CreditActivity `gorm:"foreignKey:ActivityID" json:"activity,omitempty"`
	ScheduledTime     time.Time       `json:"scheduled_time"`
	ExecutedTime      *time.Time      `json:"executed_time,omitempty"`
	Status            string          `gorm:"type:varchar(20);default:'pending'" json:"status"` // pending, executing, completed, failed
	ExecutionDetails  datatypes.JSON  `gorm:"type:jsonb" json:"execution_details"`
	ErrorMessage      string          `gorm:"type:text" json:"error_message,omitempty"`
	RetryCount        int             `gorm:"type:int;default:0" json:"retry_count"`
	NextRetryTime     *time.Time      `json:"next_retry_time,omitempty"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

func (s *CreditActivitySchedule) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// CreditActivityStatistics 活动统计
type CreditActivityStatistics struct {
	ID                   uuid.UUID       `gorm:"type:uuid;primary_key" json:"id"`
	ActivityID           uuid.UUID       `gorm:"type:uuid;not null;uniqueIndex" json:"activity_id"`
	Activity             *CreditActivity `gorm:"foreignKey:ActivityID" json:"activity,omitempty"`
	TotalParticipants    int             `gorm:"type:int;default:0" json:"total_participants"`
	CompletedParticipants int            `gorm:"type:int;default:0" json:"completed_participants"`
	TotalCreditsAwarded  int             `gorm:"type:int;default:0" json:"total_credits_awarded"`
	AverageCompletion    float64         `gorm:"type:decimal(5,2);default:0" json:"average_completion"`
	PopularityScore      float64         `gorm:"type:decimal(10,2);default:0" json:"popularity_score"`
	ParticipantsByLevel  datatypes.JSON  `gorm:"type:jsonb" json:"participants_by_level"`
	ParticipantsBySchool datatypes.JSON  `gorm:"type:jsonb" json:"participants_by_school"`
	DailyStatistics      datatypes.JSON  `gorm:"type:jsonb" json:"daily_statistics"`
	LastCalculatedAt     time.Time       `json:"last_calculated_at"`
	CreatedAt            time.Time       `json:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at"`
}

func (s *CreditActivityStatistics) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	if s.LastCalculatedAt.IsZero() {
		s.LastCalculatedAt = time.Now()
	}
	return nil
}

// CreditActivityLog 活动日志
type CreditActivityLog struct {
	ID           uuid.UUID      `gorm:"type:uuid;primary_key" json:"id"`
	ActivityID   uuid.UUID      `gorm:"type:uuid;not null;index" json:"activity_id"`
	UserID       string         `gorm:"type:varchar(36);index" json:"user_id"`
	Action       string         `gorm:"type:varchar(50);not null" json:"action"`       // 动作类型
	Details      datatypes.JSON `gorm:"type:jsonb" json:"details"`                     // 详细信息
	IPAddress    string         `gorm:"type:varchar(45)" json:"ip_address"`
	UserAgent    string         `gorm:"type:text" json:"user_agent"`
	CreatedAt    time.Time      `json:"created_at"`
}

func (l *CreditActivityLog) BeforeCreate(tx *gorm.DB) error {
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}
	return nil
}

// CreditActivityTemplate 活动模板
type CreditActivityTemplate struct {
	ID               uuid.UUID      `gorm:"type:uuid;primary_key" json:"id"`
	Name             string         `gorm:"type:varchar(200);not null" json:"name"`
	Description      string         `gorm:"type:text" json:"description"`
	Category         string         `gorm:"type:varchar(50)" json:"category"`
	TemplateData     datatypes.JSON `gorm:"type:jsonb;not null" json:"template_data"`       // 模板数据
	PreviewImageURL  string         `gorm:"type:text" json:"preview_image_url"`
	IsPublic         bool           `gorm:"default:true" json:"is_public"`
	UsageCount       int            `gorm:"type:int;default:0" json:"usage_count"`
	CreatedBy        string         `gorm:"type:varchar(36)" json:"created_by"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

func (t *CreditActivityTemplate) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}