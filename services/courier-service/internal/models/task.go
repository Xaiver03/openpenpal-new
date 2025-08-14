package models

import (
	"time"
)

// Task 任务模型 - FSD增强：添加OP Code支持
type Task struct {
	ID                string     `gorm:"primaryKey;type:varchar(36)" json:"id"`
	TaskID            string     `gorm:"unique;not null" json:"task_id"`
	LetterID          string     `gorm:"not null" json:"letter_id"`
	CourierID         *string    `json:"courier_id,omitempty"`
	PickupLocation    string     `gorm:"not null" json:"pickup_location"`
	DeliveryLocation  string     `gorm:"not null" json:"delivery_location"`
	PickupLat         float64    `json:"pickup_lat"`   // 取件地点纬度
	PickupLng         float64    `json:"pickup_lng"`   // 取件地点经度
	DeliveryLat       float64    `json:"delivery_lat"` // 送达地点纬度
	DeliveryLng       float64    `json:"delivery_lng"` // 送达地点经度
	Status            string     `gorm:"default:available" json:"status"`
	Priority          string     `gorm:"default:normal" json:"priority"` // normal, urgent, express
	Reward            float64    `gorm:"default:5.0" json:"reward"`
	EstimatedDistance string     `json:"estimated_distance"`
	EstimatedTime     string     `json:"estimated_time"` // 预计完成时间
	ContactInfo       string     `json:"contact_info"`   // 联系方式
	SpecialNote       string     `json:"special_note"`   // 特殊说明
	AcceptedAt        *time.Time `json:"accepted_at,omitempty"`
	CompletedAt       *time.Time `json:"completed_at,omitempty"`
	Deadline          *time.Time `json:"deadline,omitempty"`

	// FSD增强字段 - OP Code支持
	PickupOPCode   string `json:"pickup_op_code,omitempty" gorm:"type:varchar(6);index"`   // 取件OP Code
	DeliveryOPCode string `json:"delivery_op_code,omitempty" gorm:"type:varchar(6);index"` // 送达OP Code
	CurrentOPCode  string `json:"current_op_code,omitempty" gorm:"type:varchar(6)"`        // 当前位置OP Code

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TaskAcceptRequest 接受任务请求
type TaskAcceptRequest struct {
	EstimatedTime string `json:"estimated_time" binding:"required"`
	Note          string `json:"note"`
}

// TaskQuery 任务查询参数
type TaskQuery struct {
	Zone      string `form:"zone"`
	Status    string `form:"status"`
	Priority  string `form:"priority"`
	Limit     int    `form:"limit,default=10"`
	Offset    int    `form:"offset,default=0"`
	CourierID string `form:"courier_id"`
}

// 任务状态常量
const (
	TaskStatusAvailable = "available"  // 可接取
	TaskStatusAccepted  = "accepted"   // 已接取
	TaskStatusCollected = "collected"  // 已收取
	TaskStatusInTransit = "in_transit" // 投递中
	TaskStatusDelivered = "delivered"  // 已投递
	TaskStatusFailed    = "failed"     // 投递失败
	TaskStatusCanceled = "canceled"  // 已取消
)

// 任务优先级常量
const (
	TaskPriorityNormal  = "normal"
	TaskPriorityUrgent  = "urgent"
	TaskPriorityExpress = "express"
)

// 状态流转映射
var ValidTaskTransitions = map[string][]string{
	TaskStatusAvailable: {TaskStatusAccepted, TaskStatusCanceled},
	TaskStatusAccepted:  {TaskStatusCollected, TaskStatusCanceled},
	TaskStatusCollected: {TaskStatusInTransit},
	TaskStatusInTransit: {TaskStatusDelivered, TaskStatusFailed},
}

// CanTransitionTo 检查是否可以转换到目标状态
func (t *Task) CanTransitionTo(targetStatus string) bool {
	allowedStatuses, exists := ValidTaskTransitions[t.Status]
	if !exists {
		return false
	}

	for _, status := range allowedStatuses {
		if status == targetStatus {
			return true
		}
	}
	return false
}

// IsAssigned 检查任务是否已分配
func (t *Task) IsAssigned() bool {
	return t.CourierID != nil && *t.CourierID != ""
}

// IsCompleted 检查任务是否已完成
func (t *Task) IsCompleted() bool {
	return t.Status == TaskStatusDelivered
}

// IsActive 检查任务是否处于活跃状态
func (t *Task) IsActive() bool {
	return t.Status != TaskStatusDelivered &&
		t.Status != TaskStatusFailed &&
		t.Status != TaskStatusCanceled
}
