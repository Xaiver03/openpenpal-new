package models

import (
	"time"
)

// CreditTask 积分任务模型 - 模块化积分奖励系统
type CreditTask struct {
	ID          string          `json:"id" gorm:"primaryKey;type:varchar(36)"`
	TaskType    CreditTaskType  `json:"task_type" gorm:"not null;index"`        // 任务类型
	UserID      string          `json:"user_id" gorm:"not null;index"`          // 用户ID
	Status      CreditTaskStatus `json:"status" gorm:"default:'pending'"`       // 任务状态
	Points      int             `json:"points" gorm:"not null"`                 // 奖励积分
	Description string          `json:"description" gorm:"not null"`            // 任务描述
	Reference   string          `json:"reference"`                              // 关联对象ID
	Metadata    string          `json:"metadata" gorm:"type:jsonb"`             // 扩展数据
	
	// 执行控制
	Priority     int        `json:"priority" gorm:"default:0"`              // 优先级 (0-10, 10最高)
	MaxAttempts  int        `json:"max_attempts" gorm:"default:3"`          // 最大重试次数
	Attempts     int        `json:"attempts" gorm:"default:0"`              // 已尝试次数
	ScheduledAt  *time.Time `json:"scheduled_at,omitempty"`                 // 计划执行时间
	ExecutedAt   *time.Time `json:"executed_at,omitempty"`                  // 实际执行时间
	CompletedAt  *time.Time `json:"completed_at,omitempty"`                 // 完成时间
	FailedAt     *time.Time `json:"failed_at,omitempty"`                    // 失败时间
	ErrorMessage string     `json:"error_message"`                          // 错误信息
	
	// 限制条件
	DailyLimit   int        `json:"daily_limit" gorm:"default:0"`           // 每日限制 (0=无限制)
	WeeklyLimit  int        `json:"weekly_limit" gorm:"default:0"`          // 每周限制
	Constraints  string     `json:"constraints" gorm:"type:jsonb"`          // 约束条件JSON
	
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreditTaskType 积分任务类型
type CreditTaskType string

const (
	// 信件相关任务
	TaskTypeLetterCreated    CreditTaskType = "letter_created"    // 创建信件
	TaskTypeLetterGenerated  CreditTaskType = "letter_generated"  // 生成编号
	TaskTypeLetterDelivered  CreditTaskType = "letter_delivered"  // 信件送达
	TaskTypeLetterRead       CreditTaskType = "letter_read"       // 信件被阅读
	TaskTypeReceiveLetter    CreditTaskType = "receive_letter"    // 收到信件
	TaskTypePublicLetterLike CreditTaskType = "public_like"       // 公开信被点赞
	
	// 写作与挑战任务
	TaskTypeWritingChallenge CreditTaskType = "writing_challenge" // 写作挑战
	TaskTypeAIInteraction    CreditTaskType = "ai_interaction"    // AI互动
	
	// 信使相关任务
	TaskTypeCourierFirstTask CreditTaskType = "courier_first"     // 信使首次任务
	TaskTypeCourierDelivery  CreditTaskType = "courier_delivery"  // 信使送达
	
	// 博物馆相关任务
	TaskTypeMuseumSubmit     CreditTaskType = "museum_submit"     // 博物馆提交
	TaskTypeMuseumApproved   CreditTaskType = "museum_approved"   // 博物馆审核通过
	TaskTypeMuseumLiked      CreditTaskType = "museum_liked"      // 博物馆点赞
	
	// 系统管理任务
	TaskTypeOPCodeApproval   CreditTaskType = "opcode_approval"   // OP Code审核
	TaskTypeCommunityBadge   CreditTaskType = "community_badge"   // 社区徽章
	TaskTypeAdminReward      CreditTaskType = "admin_reward"      // 管理员奖励
)

// CreditTaskStatus 积分任务状态
type CreditTaskStatus string

const (
	TaskStatusPending   CreditTaskStatus = "pending"   // 等待执行
	TaskStatusScheduled CreditTaskStatus = "scheduled" // 已计划
	TaskStatusExecuting CreditTaskStatus = "executing" // 执行中
	TaskStatusCompleted CreditTaskStatus = "completed" // 已完成
	TaskStatusFailed    CreditTaskStatus = "failed"    // 执行失败
	TaskStatusCancelled CreditTaskStatus = "cancelled" // 已取消
	TaskStatusSkipped   CreditTaskStatus = "skipped"   // 已跳过(达到限制)
)

// TableName 设置表名
func (CreditTask) TableName() string {
	return "credit_tasks"
}

// CreditTaskQueue 积分任务队列模型
type CreditTaskQueue struct {
	ID          string          `json:"id" gorm:"primaryKey;type:varchar(36)"`
	QueueName   string          `json:"queue_name" gorm:"not null;index"`       // 队列名称
	TaskType    CreditTaskType  `json:"task_type" gorm:"not null"`              // 任务类型
	Priority    int             `json:"priority" gorm:"default:0;index"`        // 优先级
	Payload     string          `json:"payload" gorm:"type:jsonb"`              // 任务载荷
	Status      CreditTaskStatus `json:"status" gorm:"default:'pending';index"` // 状态
	RetryCount  int             `json:"retry_count" gorm:"default:0"`           // 重试次数
	MaxRetries  int             `json:"max_retries" gorm:"default:3"`           // 最大重试
	ProcessedAt *time.Time      `json:"processed_at,omitempty"`                 // 处理时间
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// TableName 设置表名
func (CreditTaskQueue) TableName() string {
	return "credit_task_queues"
}

// CreditTaskRule 积分任务规则配置
type CreditTaskRule struct {
	ID          string         `json:"id" gorm:"primaryKey;type:varchar(36)"`
	TaskType    CreditTaskType `json:"task_type" gorm:"not null;uniqueIndex"`  // 任务类型
	Points      int            `json:"points" gorm:"not null"`                 // 基础积分
	DailyLimit  int            `json:"daily_limit" gorm:"default:0"`           // 每日限制
	WeeklyLimit int            `json:"weekly_limit" gorm:"default:0"`          // 每周限制
	IsActive    bool           `json:"is_active" gorm:"default:true"`          // 是否启用
	AutoExecute bool           `json:"auto_execute" gorm:"default:true"`       // 是否自动执行
	Description string         `json:"description"`                            // 规则描述
	Constraints string         `json:"constraints" gorm:"type:jsonb"`          // 约束条件
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// TableName 设置表名
func (CreditTaskRule) TableName() string {
	return "credit_task_rules"
}

// CreditTaskBatch 批量积分任务处理
type CreditTaskBatch struct {
	ID          string             `json:"id" gorm:"primaryKey;type:varchar(36)"`
	BatchName   string             `json:"batch_name" gorm:"not null"`             // 批次名称
	TaskType    CreditTaskType     `json:"task_type" gorm:"not null"`              // 任务类型
	Status      CreditTaskStatus   `json:"status" gorm:"default:'pending'"`        // 批次状态
	TotalTasks  int                `json:"total_tasks" gorm:"default:0"`           // 总任务数
	CompletedTasks int             `json:"completed_tasks" gorm:"default:0"`       // 已完成数
	FailedTasks int                `json:"failed_tasks" gorm:"default:0"`          // 失败数
	TotalPoints int                `json:"total_points" gorm:"default:0"`          // 总积分
	StartedAt   *time.Time         `json:"started_at,omitempty"`                   // 开始时间
	CompletedAt *time.Time         `json:"completed_at,omitempty"`                 // 完成时间
	CreatedBy   string             `json:"created_by" gorm:"not null"`             // 创建者
	Metadata    string             `json:"metadata" gorm:"type:jsonb"`             // 批次元数据
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

// TableName 设置表名
func (CreditTaskBatch) TableName() string {
	return "credit_task_batches"
}

// CreditTaskStatistics 积分任务统计
type CreditTaskStatistics struct {
	TaskType       CreditTaskType `json:"task_type"`
	TotalTasks     int64          `json:"total_tasks"`
	CompletedTasks int64          `json:"completed_tasks"`
	FailedTasks    int64          `json:"failed_tasks"`
	TotalPoints    int64          `json:"total_points"`
	SuccessRate    float64        `json:"success_rate"`
	AvgPoints      float64        `json:"avg_points"`
}

// TaskMetadata 任务元数据结构
type TaskMetadata struct {
	SourceID     string                 `json:"source_id,omitempty"`     // 来源对象ID
	SourceType   string                 `json:"source_type,omitempty"`   // 来源类型
	TriggerEvent string                 `json:"trigger_event,omitempty"` // 触发事件
	UserLevel    int                    `json:"user_level,omitempty"`    // 用户等级
	Extra        map[string]interface{} `json:"extra,omitempty"`         // 扩展字段
}

// TaskConstraints 任务约束条件
type TaskConstraints struct {
	MinUserLevel     int                    `json:"min_user_level,omitempty"`     // 最低用户等级
	MaxUserLevel     int                    `json:"max_user_level,omitempty"`     // 最高用户等级
	RequiredRoles    []string               `json:"required_roles,omitempty"`     // 必需角色
	ExcludedRoles    []string               `json:"excluded_roles,omitempty"`     // 排除角色
	TimeWindow       *TimeWindow            `json:"time_window,omitempty"`        // 时间窗口
	Prerequisites    []string               `json:"prerequisites,omitempty"`      // 前置条件
	Custom           map[string]interface{} `json:"custom,omitempty"`             // 自定义约束
}

// TimeWindow 时间窗口
type TimeWindow struct {
	StartTime string `json:"start_time"` // HH:MM 格式
	EndTime   string `json:"end_time"`   // HH:MM 格式
	Days      []int  `json:"days"`       // 周几 (1-7, 1=周一)
}

// IsValidTransition 检查状态转换是否有效
func (ct *CreditTask) IsValidTransition(newStatus CreditTaskStatus) bool {
	validTransitions := map[CreditTaskStatus][]CreditTaskStatus{
		TaskStatusPending:   {TaskStatusScheduled, TaskStatusExecuting, TaskStatusCancelled},
		TaskStatusScheduled: {TaskStatusExecuting, TaskStatusCancelled},
		TaskStatusExecuting: {TaskStatusCompleted, TaskStatusFailed},
		TaskStatusFailed:    {TaskStatusExecuting, TaskStatusCancelled}, // 允许重试
		TaskStatusCompleted: {},                                         // 完成状态不可变更
		TaskStatusCancelled: {},                                         // 取消状态不可变更
		TaskStatusSkipped:   {},                                         // 跳过状态不可变更
	}
	
	allowed, exists := validTransitions[ct.Status]
	if !exists {
		return false
	}
	
	for _, allowedStatus := range allowed {
		if allowedStatus == newStatus {
			return true
		}
	}
	return false
}

// CanRetry 检查是否可以重试
func (ct *CreditTask) CanRetry() bool {
	return ct.Status == TaskStatusFailed && ct.Attempts < ct.MaxAttempts
}

// ShouldExecuteNow 检查是否应该立即执行
func (ct *CreditTask) ShouldExecuteNow() bool {
	if ct.Status != TaskStatusPending && ct.Status != TaskStatusScheduled {
		return false
	}
	
	if ct.ScheduledAt != nil && time.Now().Before(*ct.ScheduledAt) {
		return false
	}
	
	return true
}