package models

import (
	"time"
)

// UserCredit 用户积分模型
type UserCredit struct {
	ID        string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	UserID    string    `json:"user_id" gorm:"not null;uniqueIndex"`
	Total     int       `json:"total" gorm:"default:0"`     // 总积分
	Available int       `json:"available" gorm:"default:0"` // 可用积分
	Used      int       `json:"used" gorm:"default:0"`      // 已使用积分
	Earned    int       `json:"earned" gorm:"default:0"`    // 累计获得积分
	Level     int       `json:"level" gorm:"default:1"`     // 用户等级
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 设置表名
func (UserCredit) TableName() string {
	return "user_credits"
}

// CreditTransaction 积分交易记录
type CreditTransaction struct {
	ID          string     `json:"id" gorm:"primaryKey;type:varchar(36)"`
	UserID      string     `json:"user_id" gorm:"not null;index"`
	Type        string     `json:"type" gorm:"not null"`        // earn, spend, refund
	Amount      int        `json:"amount" gorm:"not null"`      // 积分数量
	Description string     `json:"description" gorm:"not null"` // 积分说明
	Reference   string     `json:"reference"`                   // 关联ID（信件ID、任务ID等）
	
	// Phase 4.1: 积分有效期机制
	ExpiresAt   *time.Time `json:"expires_at" gorm:"index"`     // 过期时间
	ExpiredAt   *time.Time `json:"expired_at"`                  // 实际过期时间
	IsExpired   bool       `json:"is_expired" gorm:"default:false;index"` // 是否已过期
	
	CreatedAt   time.Time  `json:"created_at"`
}

// TableName 设置表名
func (CreditTransaction) TableName() string {
	return "credit_transactions"
}

// CreditRule 积分规则配置（可选，用于动态配置积分规则）
type CreditRule struct {
	ID          string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Action      string    `json:"action" gorm:"not null;uniqueIndex"` // 操作类型：letter_created, letter_delivered等
	Points      int       `json:"points" gorm:"not null"`             // 积分数量
	Description string    `json:"description" gorm:"not null"`        // 规则描述
	IsActive    bool      `json:"is_active" gorm:"default:true"`      // 是否启用
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName 设置表名
func (CreditRule) TableName() string {
	return "credit_rules"
}

// UserLevel 用户等级配置
type UserLevel struct {
	ID          string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Level       int       `json:"level" gorm:"not null;uniqueIndex"` // 等级
	Name        string    `json:"name" gorm:"not null"`              // 等级名称
	RequiredExp int       `json:"required_exp" gorm:"not null"`      // 所需经验值
	Description string    `json:"description"`                       // 等级描述
	Benefits    string    `json:"benefits" gorm:"type:text"`         // 等级福利（JSON格式）
	IconURL     string    `json:"icon_url"`                          // 等级图标
	IsActive    bool      `json:"is_active" gorm:"default:true"`     // 是否启用
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName 设置表名
func (UserLevel) TableName() string {
	return "user_levels"
}

// Phase 4.1: 积分有效期机制相关模型

// CreditExpirationRule 积分过期规则配置
type CreditExpirationRule struct {
	ID             string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	RuleName       string    `json:"rule_name" gorm:"not null"`                   // 规则名称
	CreditType     string    `json:"credit_type" gorm:"not null;index"`           // 积分类型 (earn类型的具体分类)
	ExpirationDays int       `json:"expiration_days" gorm:"not null"`             // 过期天数
	NotifyDays     int       `json:"notify_days" gorm:"default:7"`                // 过期前提醒天数
	IsActive       bool      `json:"is_active" gorm:"default:true"`               // 是否启用
	Priority       int       `json:"priority" gorm:"default:0"`                   // 优先级
	Description    string    `json:"description"`                                 // 规则描述
	CreatedBy      string    `json:"created_by" gorm:"type:varchar(36)"`          // 创建人
	UpdatedBy      string    `json:"updated_by" gorm:"type:varchar(36)"`          // 更新人
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// TableName 设置表名
func (CreditExpirationRule) TableName() string {
	return "credit_expiration_rules"
}

// CreditExpirationBatch 积分过期批次记录
type CreditExpirationBatch struct {
	ID                string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	BatchDate         time.Time `json:"batch_date" gorm:"not null;index"`                // 批次日期
	TotalCredits      int       `json:"total_credits" gorm:"not null"`                   // 过期积分总数
	TotalUsers        int       `json:"total_users" gorm:"not null"`                     // 受影响用户数
	TotalTransactions int       `json:"total_transactions" gorm:"not null"`              // 过期交易数
	Status            string    `json:"status" gorm:"default:'pending'"`                 // pending, processing, completed, failed
	StartedAt         *time.Time `json:"started_at"`                                      // 开始处理时间
	CompletedAt       *time.Time `json:"completed_at"`                                    // 完成时间
	ErrorMessage      string    `json:"error_message" gorm:"type:text"`                  // 错误信息
	ProcessedBy       string    `json:"processed_by" gorm:"type:varchar(36)"`            // 处理人（系统或管理员）
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// TableName 设置表名
func (CreditExpirationBatch) TableName() string {
	return "credit_expiration_batches"
}

// CreditExpirationLog 积分过期日志
type CreditExpirationLog struct {
	ID               string     `json:"id" gorm:"primaryKey;type:varchar(36)"`
	BatchID          string     `json:"batch_id" gorm:"not null;index"`              // 批次ID
	UserID           string     `json:"user_id" gorm:"not null;index"`               // 用户ID
	TransactionID    string     `json:"transaction_id" gorm:"not null;index"`        // 原积分交易ID
	ExpiredCredits   int        `json:"expired_credits" gorm:"not null"`             // 过期积分数
	OriginalAmount   int        `json:"original_amount" gorm:"not null"`             // 原始积分数
	ExpirationReason string     `json:"expiration_reason" gorm:"not null"`           // 过期原因
	NotificationSent bool       `json:"notification_sent" gorm:"default:false"`      // 是否已发送通知
	CreatedAt        time.Time  `json:"created_at"`
}

// TableName 设置表名
func (CreditExpirationLog) TableName() string {
	return "credit_expiration_logs"
}

// CreditExpirationNotification 积分过期通知记录
type CreditExpirationNotification struct {
	ID                string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	UserID            string    `json:"user_id" gorm:"not null;index"`               // 用户ID
	NotificationType  string    `json:"notification_type" gorm:"not null"`           // warning, expired
	CreditsToExpire   int       `json:"credits_to_expire" gorm:"not null"`           // 即将过期或已过期的积分数
	ExpirationDate    time.Time `json:"expiration_date" gorm:"not null"`             // 过期日期
	NotificationSent  bool      `json:"notification_sent" gorm:"default:false"`      // 是否已发送
	NotificationTime  *time.Time `json:"notification_time"`                           // 发送时间
	NotificationError string    `json:"notification_error" gorm:"type:text"`         // 通知错误信息
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// TableName 设置表名
func (CreditExpirationNotification) TableName() string {
	return "credit_expiration_notifications"
}
