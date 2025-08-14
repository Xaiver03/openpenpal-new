package models

import (
	"gorm.io/gorm"
	"time"
)

// Courier 信使模型
type Courier struct {
	ID      string `json:"id" gorm:"primaryKey;type:varchar(36)"`
	UserID  string `json:"user_id" gorm:"type:varchar(36);not null;index"`
	User    User   `json:"user" gorm:"foreignKey:UserID;references:ID"`
	Name    string `json:"name" gorm:"size:100;not null"`
	Contact string `json:"contact" gorm:"size:100;not null"`
	School  string `json:"school" gorm:"size:100;not null"`
	Zone    string `json:"zone" gorm:"size:50;not null"` // 覆盖区域编码（兼容旧系统）

	// OP Code权限管理
	ManagedOPCodePrefix string `json:"managed_op_code_prefix" gorm:"size:6;index"` // 管理的OP Code前缀，如: PK5F**表示管理PK5F开头的所有地址

	HasPrinter      bool   `json:"has_printer" gorm:"default:false"`
	SelfIntro       string `json:"self_intro" gorm:"type:text"`
	CanMentor       string `json:"can_mentor" gorm:"size:20;default:'no'"` // yes/maybe/no
	WeeklyHours     int    `json:"weekly_hours" gorm:"default:5"`
	MaxDailyTasks   int    `json:"max_daily_tasks" gorm:"default:10"`
	TransportMethod string `json:"transport_method" gorm:"size:20"`         // walk/bike/ebike
	TimeSlots       string `json:"time_slots" gorm:"type:json"`             // JSON array of time slots
	Status          string `json:"status" gorm:"size:20;default:'pending'"` // pending/approved/rejected
	Level           int    `json:"level" gorm:"default:1"`
	TaskCount       int    `json:"task_count" gorm:"default:0"`
	Points          int    `json:"points" gorm:"default:0"`

	// 层级管理字段（与数据库对齐）
	ZoneCode    string  `json:"zone_code" gorm:"type:text"`
	ZoneType    string  `json:"zone_type" gorm:"type:text"`
	ParentID    *string `json:"parent_id" gorm:"type:varchar(36);index"`
	CreatedByID *string `json:"created_by_id" gorm:"type:varchar(36)"`
	Phone       string  `json:"phone" gorm:"type:text"`
	IDCard      string  `json:"id_card" gorm:"type:text"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// 关联关系
	Parent    *Courier  `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Children  []Courier `json:"children,omitempty" gorm:"foreignKey:ParentID"`
	CreatedBy *User     `json:"created_by,omitempty" gorm:"foreignKey:CreatedByID"`
}

// CourierTask 信使任务模型
type CourierTask struct {
	ID              string `json:"id" gorm:"primaryKey;type:varchar(36)"`
	CourierID       string `json:"courier_id" gorm:"type:varchar(36);not null;index"`
	LetterCode      string `json:"letterCode" gorm:"type:varchar(50);not null;index"`
	Title           string `json:"title" gorm:"type:varchar(200);not null"`
	SenderName      string `json:"senderName" gorm:"type:varchar(100);not null"`
	SenderPhone     string `json:"senderPhone,omitempty" gorm:"type:varchar(20)"`
	RecipientHint   string `json:"recipientHint" gorm:"type:varchar(200)"`
	TargetLocation  string `json:"targetLocation" gorm:"type:varchar(200);not null"`
	CurrentLocation string `json:"currentLocation,omitempty" gorm:"type:varchar(200)"`

	// OP Code System Integration - 地理编码支持
	PickupOPCode   string     `json:"pickup_op_code,omitempty" gorm:"type:varchar(6);index"`   // 取件OP Code
	DeliveryOPCode string     `json:"delivery_op_code,omitempty" gorm:"type:varchar(6);index"` // 送件OP Code
	CurrentOPCode  string     `json:"current_op_code,omitempty" gorm:"type:varchar(6);index"`  // 当前位置OP Code
	Priority       string     `json:"priority" gorm:"type:varchar(20);default:'normal'"`       // normal, urgent
	Status         string     `json:"status" gorm:"type:varchar(20);default:'pending'"`        // pending, collected, in_transit, delivered, failed
	EstimatedTime  int        `json:"estimatedTime" gorm:"type:int;default:30"`                // 预计时间（分钟）
	Distance       float64    `json:"distance" gorm:"type:decimal(10,2)"`                      // 距离（公里）
	CreatedAt      time.Time  `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt      time.Time  `json:"updatedAt" gorm:"autoUpdateTime"`
	Deadline       time.Time  `json:"deadline,omitempty"`
	CompletedAt    *time.Time `json:"completedAt,omitempty"`
	Instructions   string     `json:"instructions,omitempty" gorm:"type:text"`
	Reward         int        `json:"reward" gorm:"type:int;default:10"` // 积分奖励
	FailureReason  string     `json:"failureReason,omitempty" gorm:"type:text"`
	// 关联
	Courier *User       `json:"courier,omitempty" gorm:"foreignKey:CourierID;references:ID"`
	Letter  *LetterCode `json:"letter,omitempty" gorm:"foreignKey:LetterCode;references:Code"`
}

// CourierApplication 信使申请请求
type CourierApplication struct {
	Name            string   `json:"name" binding:"required,min=2,max=50"`
	Contact         string   `json:"contact" binding:"required,min=5,max=50"`
	School          string   `json:"school" binding:"required,min=2,max=100"`
	Zone            string   `json:"zone" binding:"required,min=2,max=50"`
	HasPrinter      string   `json:"hasPrinter" binding:"required,oneof=yes no"`
	SelfIntro       string   `json:"selfIntro"`
	CanMentor       string   `json:"canMentor" binding:"required,oneof=yes maybe no"`
	WeeklyHours     int      `json:"weeklyHours" binding:"min=1,max=40"`
	MaxDailyTasks   int      `json:"maxDailyTasks" binding:"min=1,max=50"`
	TransportMethod string   `json:"transportMethod" binding:"required,oneof=walk bike ebike"`
	TimeSlots       []string `json:"timeSlots" binding:"required,min=1"`
}

// CourierStatus 信使状态响应
type CourierStatus struct {
	IsApplied bool   `json:"is_applied"`
	Status    string `json:"status"`
	Level     int    `json:"level"`
	TaskCount int    `json:"task_count"`
	Points    int    `json:"points"`
	Zone      string `json:"zone"`
}

// TableName 指定表名
func (Courier) TableName() string {
	return "couriers"
}

func (CourierTask) TableName() string {
	return "courier_tasks"
}

// LevelUpgradeRequest 信使升级请求模型（与数据库对齐）
type LevelUpgradeRequest struct {
	ID            int64      `json:"id" gorm:"primaryKey;autoIncrement"`
	CourierID     string     `json:"courier_id" gorm:"type:text;not null;index"`
	CurrentLevel  int64      `json:"current_level" gorm:"not null"`
	RequestLevel  int64      `json:"request_level" gorm:"not null"`
	Reason        string     `json:"reason" gorm:"type:text"`
	Evidence      string     `json:"evidence" gorm:"type:text"`
	Status        string     `json:"status" gorm:"type:text;default:'pending'"`
	ReviewedBy    string     `json:"reviewed_by" gorm:"type:text"`
	ReviewedAt    *time.Time `json:"reviewed_at"`
	ReviewComment string     `json:"review_comment" gorm:"type:text"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`

	// 关联关系
	Courier  *Courier `json:"courier,omitempty" gorm:"foreignKey:CourierID"`
	Reviewer *User    `json:"reviewer,omitempty" gorm:"foreignKey:ReviewedBy"`
}

// TableName 指定表名
func (LevelUpgradeRequest) TableName() string {
	return "level_upgrade_requests"
}

// --- 四级信使系统相关模型 ---

// CreateCourierRequest 创建信使请求
type CreateCourierRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Level    int    `json:"level" binding:"required,min=1,max=4"`
	Region   string `json:"region"`
	School   string `json:"school"`
	Zone     string `json:"zone"`
	Building string `json:"building"`
}

// CourierInfo 信使信息响应
type CourierInfo struct {
	ID             string `json:"id"`
	Level          int    `json:"level"`
	Region         string `json:"region"`
	School         string `json:"school"`
	Zone           string `json:"zone"`
	Building       string `json:"building,omitempty"`
	TotalPoints    int    `json:"total_points"`
	CompletedTasks int    `json:"completed_tasks"`
	ParentID       string `json:"parent_id,omitempty"`
	CanCreateLevel []int  `json:"can_create_level"`
}

// SubordinateCourier 下级信使信息
type SubordinateCourier struct {
	ID             string             `json:"id"`
	Username       string             `json:"username"`
	Email          string             `json:"email"`
	Level          int                `json:"level"`
	Status         string             `json:"status"`
	Zone           string             `json:"zone"`
	Region         string             `json:"region"`
	School         string             `json:"school"`
	Building       string             `json:"building,omitempty"`
	Rating         float64            `json:"rating"`
	CompletedTasks int                `json:"completed_tasks"`
	CurrentTasks   int                `json:"current_tasks"`
	MaxTasks       int                `json:"max_tasks"`
	ParentID       string             `json:"parent_id,omitempty"`
	Profile        SubordinateProfile `json:"profile"`
	CreatedAt      string             `json:"created_at"`
	CreatedBy      string             `json:"created_by"`
}

// CourierTaskStatus 信使任务状态常量
const (
	CourierTaskStatusPending   = "pending"    // 待接取
	CourierTaskStatusCollected = "collected"  // 已取件
	CourierTaskStatusInTransit = "in_transit" // 配送中
	CourierTaskStatusDelivered = "delivered"  // 已送达
	CourierTaskStatusFailed    = "failed"     // 配送失败
)

// CourierTaskPriority 信使任务优先级常量
const (
	CourierTaskPriorityNormal = "normal" // 普通
	CourierTaskPriorityUrgent = "urgent" // 紧急
)

// SubordinateProfile 下级信使档案
type SubordinateProfile struct {
	Name       string `json:"name"`
	Phone      string `json:"phone"`
	Experience string `json:"experience"`
	Avatar     string `json:"avatar,omitempty"`
}
