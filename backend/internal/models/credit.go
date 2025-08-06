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
	ID          string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	UserID      string    `json:"user_id" gorm:"not null;index"`
	Type        string    `json:"type" gorm:"not null"`        // earn, spend, refund
	Amount      int       `json:"amount" gorm:"not null"`      // 积分数量
	Description string    `json:"description" gorm:"not null"` // 积分说明
	Reference   string    `json:"reference"`                   // 关联ID（信件ID、任务ID等）
	CreatedAt   time.Time `json:"created_at"`
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
