package models

import (
	"time"
)

// AuditLog 审计日志模型
type AuditLog struct {
	ID         string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	UserID     string    `gorm:"type:varchar(36);index" json:"user_id"`
	Action     string    `gorm:"type:varchar(50);index" json:"action"`       // 审计事件类型
	Resource   string    `gorm:"type:varchar(50);index" json:"resource"`     // 资源类型
	ResourceID string    `gorm:"type:varchar(36);index" json:"resource_id"`  // 资源ID
	Details    string    `gorm:"type:text" json:"details"`                   // JSON格式的详细信息
	IP         string    `gorm:"type:varchar(45)" json:"ip"`                 // IPv4/IPv6地址
	UserAgent  string    `gorm:"type:text" json:"user_agent"`               // 用户代理
	Result     string    `gorm:"type:varchar(20)" json:"result"`            // success/failure
	Error      string    `gorm:"type:text" json:"error,omitempty"`          // 错误信息
	Duration   float64   `gorm:"type:decimal(10,3)" json:"duration"`        // 操作耗时(秒)
	CreatedAt  time.Time `gorm:"index" json:"created_at"`
}

// TableName 指定表名
func (AuditLog) TableName() string {
	return "audit_logs"
}