package models

import (
	"gorm.io/gorm"
	"time"
)

// TaskType 任务类型
type TaskType string

const (
	TaskTypeLetterDelivery      TaskType = "letter_delivery"      // 信件投递提醒
	TaskTypeUserEngagement      TaskType = "user_engagement"      // 用户参与度检查
	TaskTypeSystemMaintenance   TaskType = "system_maintenance"   // 系统维护
	TaskTypeDataAnalytics       TaskType = "data_analytics"       // 数据分析
	TaskTypeNotificationCleanup TaskType = "notification_cleanup" // 通知清理
	TaskTypeLetterExpiration    TaskType = "letter_expiration"    // 信件过期处理
	TaskTypeCourierReminder     TaskType = "courier_reminder"     // 信使提醒
	TaskTypeBackupDatabase      TaskType = "backup_database"      // 数据库备份
	TaskTypeImageOptimization   TaskType = "image_optimization"   // 图片优化
	TaskTypeStatisticsUpdate    TaskType = "statistics_update"    // 统计数据更新
)

// SchedulerTaskStatus 定时任务状态
type SchedulerTaskStatus string

const (
	SchedulerTaskStatusPending   SchedulerTaskStatus = "pending"   // 待执行
	SchedulerTaskStatusRunning   SchedulerTaskStatus = "running"   // 执行中
	SchedulerTaskStatusCompleted SchedulerTaskStatus = "completed" // 已完成
	SchedulerTaskStatusFailed    SchedulerTaskStatus = "failed"    // 执行失败
	SchedulerTaskStatusCanceled  SchedulerTaskStatus = "canceled"  // 已取消
	SchedulerTaskStatusSkipped   SchedulerTaskStatus = "skipped"   // 已跳过
)

// TaskPriority 任务优先级
type TaskPriority string

const (
	TaskPriorityLow    TaskPriority = "low"    // 低优先级
	TaskPriorityNormal TaskPriority = "normal" // 普通优先级
	TaskPriorityHigh   TaskPriority = "high"   // 高优先级
	TaskPriorityUrgent TaskPriority = "urgent" // 紧急
)

// ScheduledTask 定时任务模型
type ScheduledTask struct {
	ID          string       `json:"id" gorm:"primaryKey"`
	Name        string       `json:"name" gorm:"size:100;not null"`
	Description string       `json:"description" gorm:"type:text"`
	TaskType    TaskType     `json:"task_type" gorm:"size:50;not null"`
	Priority    TaskPriority `json:"priority" gorm:"size:20;default:'normal'"`
	Status      SchedulerTaskStatus   `json:"status" gorm:"size:20;default:'pending'"`

	// 调度配置
	CronExpression string    `json:"cron_expression" gorm:"size:100"` // Cron表达式
	ScheduledAt    time.Time `json:"scheduled_at"`                    // 计划执行时间
	NextRunAt      time.Time `json:"next_run_at"`                     // 下次执行时间

	// 执行记录
	LastRunAt    *time.Time `json:"last_run_at"`   // 上次执行时间
	LastStatus   SchedulerTaskStatus `json:"last_status"`   // 上次执行状态
	RunCount     int        `json:"run_count"`     // 执行次数
	FailureCount int        `json:"failure_count"` // 失败次数

	// 任务配置
	Payload     string `json:"payload" gorm:"type:json"`        // 任务参数（JSON格式）
	MaxRetries  int    `json:"max_retries" gorm:"default:3"`    // 最大重试次数
	TimeoutSecs int    `json:"timeout_secs" gorm:"default:300"` // 超时时间（秒）

	// 执行限制
	IsActive  bool       `json:"is_active" gorm:"default:true"` // 是否激活
	StartDate time.Time  `json:"start_date"`                    // 开始日期
	EndDate   *time.Time `json:"end_date"`                      // 结束日期
	MaxRuns   *int       `json:"max_runs"`                      // 最大执行次数

	// 执行结果
	LastResult string `json:"last_result" gorm:"type:text"` // 上次执行结果
	LastError  string `json:"last_error" gorm:"type:text"`  // 上次执行错误

	// 审计字段
	CreatedBy string         `json:"created_by" gorm:"size:50"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// TaskExecution 任务执行记录
type TaskExecution struct {
	ID     string        `json:"id" gorm:"primaryKey"`
	TaskID string        `json:"task_id" gorm:"not null;index"`
	Task   ScheduledTask `json:"task" gorm:"foreignKey:TaskID"`

	Status    SchedulerTaskStatus `json:"status" gorm:"size:20;default:'pending'"`
	StartedAt time.Time  `json:"started_at"`
	EndedAt   *time.Time `json:"ended_at"`
	Duration  int        `json:"duration"` // 执行时长（毫秒）

	// 执行结果
	Result     string `json:"result" gorm:"type:text"`      // 执行结果
	Error      string `json:"error" gorm:"type:text"`       // 错误信息
	Output     string `json:"output" gorm:"type:text"`      // 输出日志
	RetryCount int    `json:"retry_count" gorm:"default:0"` // 重试次数

	// 执行环境
	WorkerID   string `json:"worker_id" gorm:"size:50"`    // 执行器ID
	ServerHost string `json:"server_host" gorm:"size:100"` // 服务器主机
	ProcessPID int    `json:"process_pid"`                 // 进程ID

	// 资源使用
	MemoryUsage int64   `json:"memory_usage"` // 内存使用量（字节）
	CPUUsage    float64 `json:"cpu_usage"`    // CPU使用率

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TaskTemplate 任务模板
type TaskTemplate struct {
	ID          string       `json:"id" gorm:"primaryKey"`
	Name        string       `json:"name" gorm:"size:100;not null"`
	Description string       `json:"description" gorm:"type:text"`
	TaskType    TaskType     `json:"task_type" gorm:"size:50;not null"`
	Priority    TaskPriority `json:"priority" gorm:"size:20;default:'normal'"`

	// 默认配置
	DefaultCron    string `json:"default_cron" gorm:"size:100"`
	DefaultPayload string `json:"default_payload" gorm:"type:json"`
	DefaultTimeout int    `json:"default_timeout" gorm:"default:300"`
	DefaultRetries int    `json:"default_retries" gorm:"default:3"`

	// 模板设置
	IsEnabled bool   `json:"is_enabled" gorm:"default:true"`
	Category  string `json:"category" gorm:"size:50"`
	Tags      string `json:"tags" gorm:"size:200"`

	CreatedBy string    `json:"created_by" gorm:"size:50"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TaskWorker 任务执行器
type TaskWorker struct {
	ID     string `json:"id" gorm:"primaryKey"`
	Name   string `json:"name" gorm:"size:100;not null"`
	Host   string `json:"host" gorm:"size:100;not null"`
	Port   int    `json:"port"`
	Status string `json:"status" gorm:"size:20;default:'active'"` // active/inactive/busy/error

	// 工作负载
	MaxConcurrency int `json:"max_concurrency" gorm:"default:5"` // 最大并发数
	CurrentTasks   int `json:"current_tasks" gorm:"default:0"`   // 当前任务数
	CompletedTasks int `json:"completed_tasks" gorm:"default:0"` // 已完成任务数
	FailedTasks    int `json:"failed_tasks" gorm:"default:0"`    // 失败任务数

	// 健康状态
	LastHeartbeat time.Time `json:"last_heartbeat"`
	LastError     string    `json:"last_error" gorm:"type:text"`

	// 支持的任务类型
	SupportedTypes string `json:"supported_types" gorm:"size:500"` // JSON数组

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// 请求和响应模型

// CreateTaskRequest 创建任务请求
type CreateTaskRequest struct {
	Name           string                 `json:"name" binding:"required,min=1,max=100"`
	Description    string                 `json:"description"`
	TaskType       TaskType               `json:"task_type" binding:"required"`
	Priority       TaskPriority           `json:"priority"`
	CronExpression string                 `json:"cron_expression"`
	ScheduledAt    *time.Time             `json:"scheduled_at"`
	Payload        map[string]interface{} `json:"payload"`
	MaxRetries     int                    `json:"max_retries"`
	TimeoutSecs    int                    `json:"timeout_secs"`
	StartDate      *time.Time             `json:"start_date"`
	EndDate        *time.Time             `json:"end_date"`
	MaxRuns        *int                   `json:"max_runs"`
}

// TaskQuery 任务查询参数
type TaskQuery struct {
	TaskType  TaskType     `json:"task_type" form:"task_type"`
	Priority  TaskPriority `json:"priority" form:"priority"`
	Status    SchedulerTaskStatus   `json:"status" form:"status"`
	IsActive  *bool        `json:"is_active" form:"is_active"`
	StartDate *time.Time   `json:"start_date" form:"start_date"`
	EndDate   *time.Time   `json:"end_date" form:"end_date"`
	Page      int          `json:"page" form:"page" binding:"min=1"`
	PageSize  int          `json:"page_size" form:"page_size" binding:"min=1,max=100"`
	SortBy    string       `json:"sort_by" form:"sort_by"`
	SortOrder string       `json:"sort_order" form:"sort_order"`
}

// TaskStats 任务统计
type TaskStats struct {
	TotalTasks       int64            `json:"total_tasks"`
	PendingTasks     int64            `json:"pending_tasks"`
	RunningTasks     int64            `json:"running_tasks"`
	CompletedTasks   int64            `json:"completed_tasks"`
	FailedTasks      int64            `json:"failed_tasks"`
	TasksByType      map[string]int64 `json:"tasks_by_type"`
	TasksByStatus    map[string]int64 `json:"tasks_by_status"`
	AvgExecutionTime float64          `json:"avg_execution_time"`
	SuccessRate      float64          `json:"success_rate"`
	ActiveWorkers    int              `json:"active_workers"`
	QueueLength      int              `json:"queue_length"`
	LastUpdate       time.Time        `json:"last_update"`
}

// ExecutionResult 执行结果
type ExecutionResult struct {
	Success  bool                   `json:"success"`
	Result   string                 `json:"result"`
	Error    string                 `json:"error,omitempty"`
	Duration int                    `json:"duration"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// TableName 指定表名
func (ScheduledTask) TableName() string {
	return "scheduled_tasks"
}

func (TaskExecution) TableName() string {
	return "task_executions"
}

func (TaskTemplate) TableName() string {
	return "task_templates"
}

func (TaskWorker) TableName() string {
	return "task_workers"
}
