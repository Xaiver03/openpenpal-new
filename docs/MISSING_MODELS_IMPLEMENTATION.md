# 缺失模型实现建议

根据对 `migration_fix.go` 的分析，以下模型在系统中被引用但尚未实现。本文档提供了实现建议。

## 1. DriftBottle (漂流瓶)

### 业务背景
漂流瓶是产品文档中提到的重要功能，允许用户将信件"漂流"给陌生人。

### 建议实现
```go
// internal/models/drift_bottle.go
package models

import (
    "time"
    "gorm.io/gorm"
)

type DriftBottleStatus string

const (
    DriftBottleStatusFloating  DriftBottleStatus = "floating"   // 漂流中
    DriftBottleStatusCollected DriftBottleStatus = "collected"  // 已捞取
    DriftBottleStatusExpired   DriftBottleStatus = "expired"    // 已过期
)

type DriftBottle struct {
    ID            string            `json:"id" gorm:"primaryKey;type:varchar(36)"`
    LetterID      string            `json:"letter_id" gorm:"type:varchar(36);not null;uniqueIndex"`
    SenderID      string            `json:"sender_id" gorm:"type:varchar(36);not null;index"`
    CollectorID   string            `json:"collector_id,omitempty" gorm:"type:varchar(36);index"`
    Status        DriftBottleStatus `json:"status" gorm:"type:varchar(20);not null;default:'floating'"`
    Theme         string            `json:"theme" gorm:"type:varchar(50)"` // 主题标签
    Region        string            `json:"region" gorm:"type:varchar(50)"` // 漂流区域
    CollectedAt   *time.Time        `json:"collected_at,omitempty"`
    ExpiresAt     time.Time         `json:"expires_at" gorm:"not null"`
    CreatedAt     time.Time         `json:"created_at"`
    UpdatedAt     time.Time         `json:"updated_at"`
    DeletedAt     gorm.DeletedAt    `json:"-" gorm:"index"`
    
    // 关联
    Letter    *Letter `json:"letter,omitempty" gorm:"foreignKey:LetterID"`
    Sender    *User   `json:"sender,omitempty" gorm:"foreignKey:SenderID"`
    Collector *User   `json:"collector,omitempty" gorm:"foreignKey:CollectorID"`
}

func (DriftBottle) TableName() string {
    return "drift_bottles"
}
```

## 2. FutureLetter (未来信)

### 业务背景
未来信允许用户写信给未来的自己或他人，在指定日期送达。

### 建议实现
```go
// internal/models/future_letter.go
package models

import (
    "time"
    "gorm.io/gorm"
)

type FutureLetterStatus string

const (
    FutureLetterStatusScheduled FutureLetterStatus = "scheduled" // 已计划
    FutureLetterStatusSent      FutureLetterStatus = "sent"      // 已发送
    FutureLetterStatusCancelled FutureLetterStatus = "cancelled" // 已取消
)

type FutureLetter struct {
    ID               string             `json:"id" gorm:"primaryKey;type:varchar(36)"`
    LetterID         string             `json:"letter_id" gorm:"type:varchar(36);not null;uniqueIndex"`
    SenderID         string             `json:"sender_id" gorm:"type:varchar(36);not null;index"`
    RecipientID      string             `json:"recipient_id,omitempty" gorm:"type:varchar(36);index"`
    RecipientOPCode  string             `json:"recipient_op_code,omitempty" gorm:"type:varchar(6);index"`
    Status           FutureLetterStatus `json:"status" gorm:"type:varchar(20);not null;default:'scheduled'"`
    ScheduledDate    time.Time          `json:"scheduled_date" gorm:"not null;index"`
    DeliveryMethod   string             `json:"delivery_method" gorm:"type:varchar(20);default:'system'"` // system/courier
    ReminderEnabled  bool               `json:"reminder_enabled" gorm:"default:true"`
    ReminderDays     int                `json:"reminder_days" gorm:"default:7"` // 提前提醒天数
    LastReminderSent *time.Time         `json:"last_reminder_sent,omitempty"`
    SentAt           *time.Time         `json:"sent_at,omitempty"`
    CreatedAt        time.Time          `json:"created_at"`
    UpdatedAt        time.Time          `json:"updated_at"`
    DeletedAt        gorm.DeletedAt     `json:"-" gorm:"index"`
    
    // 关联
    Letter    *Letter `json:"letter,omitempty" gorm:"foreignKey:LetterID"`
    Sender    *User   `json:"sender,omitempty" gorm:"foreignKey:SenderID"`
    Recipient *User   `json:"recipient,omitempty" gorm:"foreignKey:RecipientID"`
}

func (FutureLetter) TableName() string {
    return "future_letters"
}
```

## 3. CourierPromotion (信使晋升记录)

### 业务背景
记录信使在4级体系中的晋升历史。

### 建议实现
```go
// internal/models/courier_promotion.go
package models

import (
    "time"
    "gorm.io/gorm"
)

type CourierPromotion struct {
    ID                string         `json:"id" gorm:"primaryKey;type:varchar(36)"`
    CourierID         string         `json:"courier_id" gorm:"type:varchar(36);not null;index"`
    FromLevel         int            `json:"from_level" gorm:"not null"`
    ToLevel           int            `json:"to_level" gorm:"not null"`
    PromotionType     string         `json:"promotion_type" gorm:"type:varchar(20)"` // auto/manual/special
    Reason            string         `json:"reason" gorm:"type:text"`
    PerformanceScore  float64        `json:"performance_score"`
    TaskCount         int            `json:"task_count"`
    SuccessRate       float64        `json:"success_rate"`
    ApprovedBy        string         `json:"approved_by,omitempty" gorm:"type:varchar(36)"`
    ApprovedAt        *time.Time     `json:"approved_at,omitempty"`
    EffectiveDate     time.Time      `json:"effective_date" gorm:"not null"`
    CreatedAt         time.Time      `json:"created_at"`
    UpdatedAt         time.Time      `json:"updated_at"`
    DeletedAt         gorm.DeletedAt `json:"-" gorm:"index"`
    
    // 关联
    Courier    *Courier `json:"courier,omitempty" gorm:"foreignKey:CourierID"`
    ApprovedBy *User    `json:"approved_by_user,omitempty" gorm:"foreignKey:ApprovedBy"`
}

func (CourierPromotion) TableName() string {
    return "courier_promotions"
}
```

## 4. CourierStats (信使统计)

### 业务背景
实时统计信使的工作表现数据。

### 建议实现
```go
// internal/models/courier_stats.go
package models

import (
    "time"
    "gorm.io/gorm"
)

type CourierStats struct {
    ID                   string         `json:"id" gorm:"primaryKey;type:varchar(36)"`
    CourierID            string         `json:"courier_id" gorm:"type:varchar(36);not null;uniqueIndex"`
    TotalTasks           int            `json:"total_tasks" gorm:"default:0"`
    CompletedTasks       int            `json:"completed_tasks" gorm:"default:0"`
    FailedTasks          int            `json:"failed_tasks" gorm:"default:0"`
    InProgressTasks      int            `json:"in_progress_tasks" gorm:"default:0"`
    AverageDeliveryTime  float64        `json:"average_delivery_time"` // 小时
    SuccessRate          float64        `json:"success_rate"`
    CustomerRating       float64        `json:"customer_rating" gorm:"default:0"`
    TotalRatings         int            `json:"total_ratings" gorm:"default:0"`
    MonthlyTasks         int            `json:"monthly_tasks" gorm:"default:0"`
    WeeklyTasks          int            `json:"weekly_tasks" gorm:"default:0"`
    LastTaskAt           *time.Time     `json:"last_task_at,omitempty"`
    LastUpdatedAt        time.Time      `json:"last_updated_at"`
    CreatedAt            time.Time      `json:"created_at"`
    UpdatedAt            time.Time      `json:"updated_at"`
    DeletedAt            gorm.DeletedAt `json:"-" gorm:"index"`
    
    // 关联
    Courier *Courier `json:"courier,omitempty" gorm:"foreignKey:CourierID"`
}

func (CourierStats) TableName() string {
    return "courier_stats"
}
```

## 5. AIConversation (AI对话记录)

### 业务背景
记录用户与AI助手的对话历史，用于改进AI服务。

### 建议实现
```go
// internal/models/ai_conversation.go
package models

import (
    "time"
    "gorm.io/gorm"
    "gorm.io/datatypes"
)

type AIConversation struct {
    ID              string         `json:"id" gorm:"primaryKey;type:varchar(36)"`
    UserID          string         `json:"user_id" gorm:"type:varchar(36);not null;index"`
    SessionID       string         `json:"session_id" gorm:"type:varchar(36);index"`
    ConversationType string        `json:"conversation_type" gorm:"type:varchar(50)"` // inspiration/reply_advice/match
    Context         datatypes.JSON `json:"context" gorm:"type:jsonb"`
    Messages        datatypes.JSON `json:"messages" gorm:"type:jsonb"` // 对话消息数组
    TotalTokens     int            `json:"total_tokens" gorm:"default:0"`
    ModelUsed       string         `json:"model_used" gorm:"type:varchar(50)"`
    UserSatisfaction int           `json:"user_satisfaction,omitempty"` // 1-5评分
    UserFeedback    string         `json:"user_feedback,omitempty" gorm:"type:text"`
    StartedAt       time.Time      `json:"started_at"`
    EndedAt         *time.Time     `json:"ended_at,omitempty"`
    CreatedAt       time.Time      `json:"created_at"`
    UpdatedAt       time.Time      `json:"updated_at"`
    DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`
    
    // 关联
    User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (AIConversation) TableName() string {
    return "ai_conversations"
}
```

## 6. SchedulerLock (调度器锁)

### 业务背景
防止分布式环境中的任务重复执行。

### 建议实现
```go
// internal/models/scheduler_lock.go
package models

import (
    "time"
)

type SchedulerLock struct {
    LockName      string    `json:"lock_name" gorm:"primaryKey;type:varchar(255)"`
    LockedBy      string    `json:"locked_by" gorm:"type:varchar(255);not null"`
    LockedAt      time.Time `json:"locked_at" gorm:"not null"`
    LockUntil     time.Time `json:"lock_until" gorm:"not null;index"`
    LastHeartbeat time.Time `json:"last_heartbeat"`
}

func (SchedulerLock) TableName() string {
    return "scheduler_locks"
}
```

## 7. SchedulerLog (调度器日志)

### 业务背景
记录定时任务的执行历史。

### 建议实现
```go
// internal/models/scheduler_log.go
package models

import (
    "time"
    "gorm.io/gorm"
    "gorm.io/datatypes"
)

type SchedulerLog struct {
    ID           string         `json:"id" gorm:"primaryKey;type:varchar(36)"`
    TaskID       string         `json:"task_id" gorm:"type:varchar(36);not null;index"`
    TaskName     string         `json:"task_name" gorm:"type:varchar(255);not null;index"`
    ExecutionID  string         `json:"execution_id" gorm:"type:varchar(36);index"`
    Status       string         `json:"status" gorm:"type:varchar(20);not null"` // started/success/failed
    StartedAt    time.Time      `json:"started_at" gorm:"not null"`
    EndedAt      *time.Time     `json:"ended_at,omitempty"`
    Duration     int64          `json:"duration,omitempty"` // 毫秒
    Output       string         `json:"output,omitempty" gorm:"type:text"`
    ErrorMessage string         `json:"error_message,omitempty" gorm:"type:text"`
    Metadata     datatypes.JSON `json:"metadata,omitempty" gorm:"type:jsonb"`
    CreatedAt    time.Time      `json:"created_at"`
    DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

func (SchedulerLog) TableName() string {
    return "scheduler_logs"
}
```

## 实现建议

1. **优先级**：
   - 高：DriftBottle, FutureLetter（核心业务功能）
   - 中：CourierPromotion, CourierStats（信使系统增强）
   - 低：AIConversation, SchedulerLock, SchedulerLog（系统优化）

2. **实施步骤**：
   - 创建模型文件
   - 添加到 migration_fix.go
   - 创建对应的 service 层
   - 添加 API handlers
   - 编写测试用例
   - 更新文档

3. **注意事项**：
   - 所有ID字段使用UUID (varchar(36))
   - 包含软删除支持 (DeletedAt)
   - 添加必要的索引以优化查询性能
   - 使用JSONB类型存储复杂数据结构
   - 遵循现有的命名规范和代码风格

4. **数据库迁移**：
   - 使用GORM的AutoMigrate功能
   - 考虑向后兼容性
   - 为生产环境准备迁移脚本