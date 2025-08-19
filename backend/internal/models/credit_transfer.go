package models

import (
	"time"

	"gorm.io/datatypes"
)

// Phase 4.2: 积分转赠系统数据模型

// CreditTransferStatus 积分转赠状态
type CreditTransferStatus string

const (
	TransferStatusPending   CreditTransferStatus = "pending"   // 待处理
	TransferStatusProcessed CreditTransferStatus = "processed" // 已处理
	TransferStatusCanceled  CreditTransferStatus = "canceled"  // 已取消
	TransferStatusExpired   CreditTransferStatus = "expired"   // 已过期
	TransferStatusRejected  CreditTransferStatus = "rejected"  // 被拒绝
)

// CreditTransferType 积分转赠类型
type CreditTransferType string

const (
	TransferTypeDirect CreditTransferType = "direct" // 直接转赠
	TransferTypeGift   CreditTransferType = "gift"   // 礼物转赠
	TransferTypeReward CreditTransferType = "reward" // 奖励转赠
)

// CreditTransfer 积分转赠记录
type CreditTransfer struct {
	ID               string               `json:"id" gorm:"primaryKey;type:varchar(36)"`
	FromUserID       string               `json:"from_user_id" gorm:"type:varchar(36);not null;index"`              // 转出用户ID
	ToUserID         string               `json:"to_user_id" gorm:"type:varchar(36);not null;index"`                // 转入用户ID
	Amount           int                  `json:"amount" gorm:"not null"`                                            // 转赠积分数量
	TransferType     CreditTransferType   `json:"transfer_type" gorm:"type:varchar(20);not null"` // 转赠类型
	Status           CreditTransferStatus `json:"status" gorm:"type:varchar(20);default:'pending';index"` // 转赠状态
	Message          string               `json:"message" gorm:"type:text"`                                          // 转赠留言
	ProcessedAt      *time.Time           `json:"processed_at" gorm:"index"`                                         // 处理时间
	ExpiresAt        time.Time            `json:"expires_at" gorm:"not null;index"`                                  // 过期时间
	Fee              int                  `json:"fee" gorm:"default:0"`                                              // 转赠手续费
	Reference        string               `json:"reference" gorm:"type:varchar(100)"`                               // 关联引用
	Metadata         datatypes.JSON       `json:"metadata" gorm:"type:json"`                                         // 额外元数据
	CreatedAt        time.Time            `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time            `json:"updated_at" gorm:"autoUpdateTime"`

	// 关联关系
	FromUser User `json:"from_user" gorm:"foreignKey:FromUserID;references:ID"`
	ToUser   User `json:"to_user" gorm:"foreignKey:ToUserID;references:ID"`
}

// TableName 设置表名
func (CreditTransfer) TableName() string {
	return "credit_transfers"
}

// IsExpired 检查转赠是否已过期
func (ct *CreditTransfer) IsExpired() bool {
	return time.Now().After(ct.ExpiresAt)
}

// CanBeCanceled 检查转赠是否可以取消
func (ct *CreditTransfer) CanBeCanceled() bool {
	return ct.Status == TransferStatusPending && !ct.IsExpired()
}

// CanBeProcessed 检查转赠是否可以处理
func (ct *CreditTransfer) CanBeProcessed() bool {
	return ct.Status == TransferStatusPending && !ct.IsExpired()
}

// CreditTransferRule 积分转赠规则
type CreditTransferRule struct {
	ID                    string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	RuleName              string    `json:"rule_name" gorm:"type:varchar(100);not null"`                        // 规则名称
	MinAmount             int       `json:"min_amount" gorm:"default:1"`                                        // 最小转赠数量
	MaxAmount             int       `json:"max_amount" gorm:"default:1000"`                                     // 最大转赠数量
	DailyLimit            int       `json:"daily_limit" gorm:"default:500"`                                     // 每日转赠限制
	MonthlyLimit          int       `json:"monthly_limit" gorm:"default:5000"`                                  // 每月转赠限制
	FeeRate               float64   `json:"fee_rate" gorm:"type:decimal(5,4);default:0"`                        // 手续费率 (0-1)
	MinFee                int       `json:"min_fee" gorm:"default:0"`                                           // 最小手续费
	MaxFee                int       `json:"max_fee" gorm:"default:100"`                                         // 最大手续费
	ExpirationHours       int       `json:"expiration_hours" gorm:"default:72"`                                 // 转赠过期小时数
	RequireConfirmation   bool      `json:"require_confirmation" gorm:"default:true"`                           // 是否需要确认
	AllowSelfTransfer     bool      `json:"allow_self_transfer" gorm:"default:false"`                           // 是否允许自转
	RestrictedUserLevels  datatypes.JSON `json:"restricted_user_levels" gorm:"type:json"`                         // 受限用户等级
	AllowedTransferTypes  datatypes.JSON `json:"allowed_transfer_types" gorm:"type:json"`                         // 允许的转赠类型
	IsActive              bool      `json:"is_active" gorm:"default:true"`                                      // 是否启用
	Priority              int       `json:"priority" gorm:"default:0"`                                          // 优先级
	ApplicableUserRoles   datatypes.JSON `json:"applicable_user_roles" gorm:"type:json"`                          // 适用用户角色
	Description           string    `json:"description" gorm:"type:text"`                                       // 规则描述
	CreatedBy             string    `json:"created_by" gorm:"type:varchar(36)"`                                 // 创建人
	UpdatedBy             string    `json:"updated_by" gorm:"type:varchar(36)"`                                 // 更新人
	CreatedAt             time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt             time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName 设置表名
func (CreditTransferRule) TableName() string {
	return "credit_transfer_rules"
}

// CalculateFee 计算转赠手续费
func (ctr *CreditTransferRule) CalculateFee(amount int) int {
	fee := int(float64(amount) * ctr.FeeRate)
	if fee < ctr.MinFee {
		fee = ctr.MinFee
	}
	if fee > ctr.MaxFee {
		fee = ctr.MaxFee
	}
	return fee
}

// CreditTransferLimit 积分转赠限制记录
type CreditTransferLimit struct {
	ID           string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	UserID       string    `json:"user_id" gorm:"type:varchar(36);not null;index"`       // 用户ID
	Date         time.Time `json:"date" gorm:"type:date;not null;index"`                 // 日期
	DailyUsed    int       `json:"daily_used" gorm:"default:0"`                          // 当日已使用
	MonthlyUsed  int       `json:"monthly_used" gorm:"default:0"`                        // 当月已使用
	DailyCount   int       `json:"daily_count" gorm:"default:0"`                         // 当日转赠次数
	MonthlyCount int       `json:"monthly_count" gorm:"default:0"`                       // 当月转赠次数
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// 关联关系
	User User `json:"user" gorm:"foreignKey:UserID;references:ID"`
}

// TableName 设置表名
func (CreditTransferLimit) TableName() string {
	return "credit_transfer_limits"
}

// CreditTransferNotification 积分转赠通知
type CreditTransferNotification struct {
	ID           string                     `json:"id" gorm:"primaryKey;type:varchar(36)"`
	TransferID   string                     `json:"transfer_id" gorm:"type:varchar(36);not null;index"`                                                      // 转赠ID
	UserID       string                     `json:"user_id" gorm:"type:varchar(36);not null;index"`                                                          // 接收用户ID
	NotificationType string                 `json:"notification_type" gorm:"type:varchar(50);not null"` // 通知类型
	Title        string                     `json:"title" gorm:"type:varchar(200);not null"`                                                                 // 通知标题
	Content      string                     `json:"content" gorm:"type:text;not null"`                                                                       // 通知内容
	IsRead       bool                       `json:"is_read" gorm:"default:false"`                                                                            // 是否已读
	ReadAt       *time.Time                 `json:"read_at"`                                                                                                  // 阅读时间
	CreatedAt    time.Time                  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time                  `json:"updated_at" gorm:"autoUpdateTime"`

	// 关联关系
	Transfer CreditTransfer `json:"transfer" gorm:"foreignKey:TransferID;references:ID"`
	User     User           `json:"user" gorm:"foreignKey:UserID;references:ID"`
}

// TableName 设置表名
func (CreditTransferNotification) TableName() string {
	return "credit_transfer_notifications"
}

// 请求和响应模型

// CreateCreditTransferRequest 创建积分转赠请求
type CreateCreditTransferRequest struct {
	ToUserID     string             `json:"to_user_id" binding:"required"`        // 转入用户ID
	Amount       int                `json:"amount" binding:"required,min=1"`      // 转赠数量
	TransferType CreditTransferType `json:"transfer_type" binding:"required"`     // 转赠类型
	Message      string             `json:"message"`                              // 转赠留言
	Reference    string             `json:"reference"`                            // 关联引用
}

// ProcessCreditTransferRequest 处理积分转赠请求
type ProcessCreditTransferRequest struct {
	Action string `json:"action" binding:"required,oneof=accept reject"` // 处理动作：accept/reject
	Reason string `json:"reason"`                                         // 处理原因（拒绝时）
}

// CreditTransferResponse 积分转赠响应
type CreditTransferResponse struct {
	Transfer     CreditTransfer `json:"transfer"`
	FromUsername string         `json:"from_username"`
	ToUsername   string         `json:"to_username"`
	CanCancel    bool           `json:"can_cancel"`
	CanProcess   bool           `json:"can_process"`
	IsExpired    bool           `json:"is_expired"`
}

// CreditTransferStatsResponse 积分转赠统计响应
type CreditTransferStatsResponse struct {
	TotalTransfers    int64   `json:"total_transfers"`     // 总转赠数
	TotalAmount       int     `json:"total_amount"`        // 总转赠积分
	TotalFees         int     `json:"total_fees"`          // 总手续费
	PendingTransfers  int64   `json:"pending_transfers"`   // 待处理转赠
	ProcessedTransfers int64  `json:"processed_transfers"` // 已处理转赠
	CanceledTransfers int64   `json:"canceled_transfers"`  // 已取消转赠
	ExpiredTransfers  int64   `json:"expired_transfers"`   // 已过期转赠
	AverageAmount     float64 `json:"average_amount"`      // 平均转赠数量
	DailyUsed         int     `json:"daily_used"`          // 当日已使用
	DailyLimit        int     `json:"daily_limit"`         // 当日限制
	MonthlyUsed       int     `json:"monthly_used"`        // 当月已使用
	MonthlyLimit      int     `json:"monthly_limit"`       // 当月限制
}

// UserTransferSummary 用户转赠摘要
type UserTransferSummary struct {
	UserID           string `json:"user_id"`
	Username         string `json:"username"`
	SentCount        int64  `json:"sent_count"`        // 发送次数
	ReceivedCount    int64  `json:"received_count"`    // 接收次数
	SentAmount       int    `json:"sent_amount"`       // 发送总额
	ReceivedAmount   int    `json:"received_amount"`   // 接收总额
	PaidFees         int    `json:"paid_fees"`         // 支付手续费
	LastTransferTime *time.Time `json:"last_transfer_time"` // 最后转赠时间
}

// CreditTransferSearchRequest 积分转赠搜索请求
type CreditTransferSearchRequest struct {
	UserID       string               `json:"user_id"`       // 用户ID过滤
	Status       CreditTransferStatus `json:"status"`        // 状态过滤
	TransferType CreditTransferType   `json:"transfer_type"` // 类型过滤
	StartDate    *time.Time           `json:"start_date"`    // 开始日期
	EndDate      *time.Time           `json:"end_date"`      // 结束日期
	MinAmount    *int                 `json:"min_amount"`    // 最小金额
	MaxAmount    *int                 `json:"max_amount"`    // 最大金额
	Page         int                  `json:"page"`          // 页码
	Limit        int                  `json:"limit"`         // 每页数量
}

// BatchTransferRequest 批量转赠请求
type BatchTransferRequest struct {
	ToUserIDs    []string           `json:"to_user_ids" binding:"required,min=1"`    // 转入用户ID列表
	Amount       int                `json:"amount" binding:"required,min=1"`         // 每人转赠数量
	TransferType CreditTransferType `json:"transfer_type" binding:"required"`        // 转赠类型
	Message      string             `json:"message"`                                 // 转赠留言
	Reference    string             `json:"reference"`                               // 关联引用
}

// BatchTransferResponse 批量转赠响应
type BatchTransferResponse struct {
	SuccessCount int                `json:"success_count"`   // 成功数量
	FailureCount int                `json:"failure_count"`   // 失败数量
	Transfers    []CreditTransfer   `json:"transfers"`       // 转赠记录列表
	Errors       []string           `json:"errors"`          // 错误信息列表
}