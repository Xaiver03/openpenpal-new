package models

import (
	"time"
)

// Courier 信使模型
type Courier struct {
	ID          string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	UserID      string    `gorm:"not null;unique;type:varchar(36)" json:"user_id"`
	Name        string    `gorm:"type:varchar(100);not null" json:"name"`
	Contact     string    `gorm:"type:varchar(100);not null" json:"contact"`
	School      string    `gorm:"type:varchar(100);not null" json:"school"`
	Zone        string    `gorm:"not null;type:varchar(50)" json:"zone"`              // 服务区域
	Phone       string    `json:"phone"`             // 联系电话
	IDCard      string    `json:"id_card"`           // 身份证号
	Status      string    `gorm:"default:'pending';type:varchar(20)" json:"status"`     // pending, approved, suspended, rejected
	Level       int       `gorm:"default:1" json:"level"`            // 信使等级 (1-4级)
	Rating      float64   `gorm:"default:5.0" json:"rating"`         // 评分
	Experience  string    `json:"experience"`                        // 工作经验描述
	Note        string    `json:"note"`                              // 管理员备注
	ApprovedAt  *time.Time `json:"approved_at,omitempty"`            // 审核通过时间
	
	// 4级层级系统新增字段
	ParentID    *string   `gorm:"index;type:varchar(36)" json:"parent_id,omitempty"`  // 上级信使ID
	ZoneCode    string    `gorm:"index" json:"zone_code"`            // 负责区域编码
	ZoneType    string    `json:"zone_type"`                         // city/school/zone/building
	Points      int       `gorm:"default:0" json:"points"`           // 积分
	CreatedByID *string   `gorm:"index;type:varchar(36)" json:"created_by_id,omitempty"` // 创建者ID(上级管理员)
	
	// 层级关系
	Parent       *Courier   `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Subordinates []Courier  `gorm:"foreignKey:ParentID" json:"subordinates,omitempty"`
	CreatedBy    *Courier   `gorm:"foreignKey:CreatedByID" json:"created_by,omitempty"`
	
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CourierApplication 信使申请表单
type CourierApplication struct {
	Zone       string `json:"zone" binding:"required"`
	Phone      string `json:"phone" binding:"required"`
	IDCard     string `json:"id_card" binding:"required"`
	Experience string `json:"experience"`
	Level      int    `json:"level,omitempty"`    // 申请等级（默认为1）
	ZoneCode   string `json:"zone_code,omitempty"` // 申请管理的区域编码
}

// CreateSubordinateRequest 创建下级信使请求
type CreateSubordinateRequest struct {
	UserID     string `json:"user_id" binding:"required"`
	Level      int    `json:"level" binding:"required"`
	ZoneCode   string `json:"zone_code" binding:"required"`
	ZoneType   string `json:"zone_type" binding:"required"`
	Phone      string `json:"phone" binding:"required"`
	IDCard     string `json:"id_card" binding:"required"`
	Experience string `json:"experience"`
	Note       string `json:"note"`
}

// AssignZoneRequest 分配区域请求
type AssignZoneRequest struct {
	ZoneCode   string `json:"zone_code" binding:"required"`
	ZoneType   string `json:"zone_type" binding:"required"`
}

// TransferSubordinateRequest 转移下级信使请求
type TransferSubordinateRequest struct {
	NewParentID string `json:"new_parent_id" binding:"required"`
	Reason      string `json:"reason"`
}

// CourierHierarchyResponse 层级结构响应
type CourierHierarchyResponse struct {
	Courier      Courier   `json:"courier"`
	Parent       *Courier  `json:"parent,omitempty"`
	Subordinates []Courier `json:"subordinates,omitempty"`
	Level        int       `json:"level"`
	CanManage    []string  `json:"can_manage"`    // 可管理的操作列表
	Permissions  []string  `json:"permissions"`   // 权限列表
}

// CourierStats 信使统计信息
type CourierStats struct {
	TotalTasks      int     `json:"total_tasks"`
	CompletedTasks  int     `json:"completed_tasks"`
	SuccessRate     float64 `json:"success_rate"`
	AverageRating   float64 `json:"average_rating"`
	TotalEarnings   float64 `json:"total_earnings"`
	ThisMonthTasks  int     `json:"this_month_tasks"`
}

// CourierRanking 信使排行榜
type CourierRanking struct {
	ID            string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	CourierID     string    `gorm:"not null;index;type:varchar(36)" json:"courier_id"`
	Courier       Courier   `gorm:"foreignKey:CourierID;references:ID" json:"courier"`
	SchoolRank    int       `json:"school_rank"`
	ZoneRank      int       `json:"zone_rank"`
	NationalRank  int       `json:"national_rank"`
	Points        int       `json:"points"`
	TotalTasks    int       `json:"total_tasks"`
	SuccessRate   float64   `json:"success_rate"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// CourierLeaderboardRequest 排行榜请求
type CourierLeaderboardRequest struct {
	Type   string `json:"type" binding:"required"` // school/zone/national
	Limit  int    `json:"limit"`                   // 默认10
	Offset int    `json:"offset"`                  // 默认0
	ZoneCode string `json:"zone_code,omitempty"`   // 区域筛选
}

// CourierLeaderboardResponse 排行榜响应
type CourierLeaderboardResponse struct {
	Rankings []CourierRanking `json:"rankings"`
	Total    int              `json:"total"`
	Page     int              `json:"page"`
	Limit    int              `json:"limit"`
}

// CourierPointsHistory 积分历史记录
type CourierPointsHistory struct {
	ID          string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	CourierID   string    `gorm:"not null;index;type:varchar(36)" json:"courier_id"`
	TaskID      *string   `gorm:"index;type:varchar(36)" json:"task_id,omitempty"`
	Points      int       `json:"points"`       // 变动积分（可为负）
	Type        string    `json:"type"`         // task_completion/bonus/penalty/exchange
	Description string    `json:"description"`  // 积分变动说明
	CreatedAt   time.Time `json:"created_at"`
}

// 信使状态常量
const (
	CourierStatusPending   = "pending"
	CourierStatusApproved  = "approved" 
	CourierStatusSuspended = "suspended"
	CourierStatusRejected  = "rejected"
)

// 信使等级常量 (4级层级系统)
const (
	CourierLevelOne   = 1 // 一级信使：楼栋级
	CourierLevelTwo   = 2 // 二级信使：片区级  
	CourierLevelThree = 3 // 三级信使：学校级
	CourierLevelFour  = 4 // 四级信使：城市级
)

// 区域类型常量
const (
	ZoneTypeBuilding = "building" // 楼栋
	ZoneTypeArea     = "area"     // 片区
	ZoneTypeSchool   = "school"   // 学校
	ZoneTypeCity     = "city"     // 城市
)

// 层级权限映射
var LevelZoneTypeMapping = map[int]string{
	CourierLevelOne:   ZoneTypeBuilding,
	CourierLevelTwo:   ZoneTypeArea,
	CourierLevelThree: ZoneTypeSchool,
	CourierLevelFour:  ZoneTypeCity,
}

// IsActive 检查信使是否可以接受任务
func (c *Courier) IsActive() bool {
	return c.Status == CourierStatusApproved
}

// CanAcceptTask 检查信使是否可以接受新任务
func (c *Courier) CanAcceptTask() bool {
	return c.IsActive() && c.Rating >= 3.0
}

// GetLevelName 获取等级名称
func (c *Courier) GetLevelName() string {
	switch c.Level {
	case CourierLevelOne:
		return "一级信使"
	case CourierLevelTwo:
		return "二级信使"
	case CourierLevelThree:
		return "三级信使"
	case CourierLevelFour:
		return "四级信使"
	default:
		return "未知等级"
	}
}

// GetZoneTypeName 获取区域类型名称
func (c *Courier) GetZoneTypeName() string {
	switch c.ZoneType {
	case ZoneTypeBuilding:
		return "楼栋"
	case ZoneTypeArea:
		return "片区"
	case ZoneTypeSchool:
		return "学校"
	case ZoneTypeCity:
		return "城市"
	default:
		return "未知区域"
	}
}

// CanManageSubordinate 检查是否可以管理下级信使
func (c *Courier) CanManageSubordinate(targetLevel int) bool {
	if !c.IsActive() {
		return false
	}
	
	// 只能管理比自己等级低的信使
	return c.Level > targetLevel && c.Level-targetLevel == 1
}

// CanCreateSubordinate 检查是否可以创建下级信使
func (c *Courier) CanCreateSubordinate() bool {
	// 二级以上信使可以创建下级
	return c.IsActive() && c.Level >= CourierLevelTwo
}

// IsSubordinateOf 检查是否为指定信使的下级
func (c *Courier) IsSubordinateOf(managerID string) bool {
	return c.ParentID != nil && *c.ParentID == managerID
}

// HasSubordinates 检查是否有下级信使
func (c *Courier) HasSubordinates() bool {
	return len(c.Subordinates) > 0
}

// 层级任务分配相关结构体

// HierarchicalTaskAssignmentRequest 层级任务分配请求
type HierarchicalTaskAssignmentRequest struct {
	TaskID           string `json:"task_id" binding:"required"`
	AssignmentType   string `json:"assignment_type" binding:"required"` // direct, cascade, auto_hierarchy
	TargetCourierID  *string `json:"target_courier_id,omitempty"`
	Priority         int    `json:"priority,omitempty"`
	Notes            string `json:"notes,omitempty"`
}

// BatchHierarchicalAssignmentRequest 批量层级分配请求
type BatchHierarchicalAssignmentRequest struct {
	AssignmentType    string                    `json:"assignment_type" binding:"required"`
	TaskAssignments   []TaskAssignmentItem      `json:"task_assignments" binding:"required"`
	Notes             string                    `json:"notes,omitempty"`
}

// TaskAssignmentItem 任务分配项
type TaskAssignmentItem struct {
	TaskID          string  `json:"task_id" binding:"required"`
	TargetCourierID *string `json:"target_courier_id,omitempty"`
	Priority        int     `json:"priority,omitempty"`
	Notes           string  `json:"notes,omitempty"`
}

// TaskReassignmentRequest 任务重新分配请求
type TaskReassignmentRequest struct {
	TaskID        string `json:"task_id" binding:"required"`
	NewCourierID  string `json:"new_courier_id" binding:"required"`
	Reason        string `json:"reason" binding:"required"`
}

// TaskAssignmentResult 任务分配结果
type TaskAssignmentResult struct {
	TaskID            string `json:"task_id"`
	Success           bool   `json:"success"`
	AssignedCourierID *string `json:"assigned_courier_id,omitempty"`
	Error             string `json:"error,omitempty"`
}

// BatchAssignmentResponse 批量分配响应
type BatchAssignmentResponse struct {
	Results      []TaskAssignmentResult `json:"results"`
	SuccessCount int                   `json:"success_count"`
	TotalCount   int                   `json:"total_count"`
}

// TaskAssignmentHistory 任务分配历史
type TaskAssignmentHistory struct {
	ID                  string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	TaskID              string    `gorm:"index" json:"task_id"`
	AssignedCourierID   string    `gorm:"index;type:varchar(36)" json:"assigned_courier_id"`
	AssignedBy          string    `gorm:"index;type:varchar(36)" json:"assigned_by"`
	AssignmentType      string    `json:"assignment_type"` // direct, cascade, auto_hierarchy, reassignment
	PreviousCourierID   *string   `json:"previous_courier_id,omitempty"`
	ReassignmentReason  string    `json:"reassignment_reason,omitempty"`
	CreatedAt           time.Time `json:"created_at"`

	// 关联 - 注释掉避免循环依赖
	// Task            Task    `gorm:"foreignKey:TaskID;references:TaskID" json:"task,omitempty"`
	AssignedCourier Courier `gorm:"foreignKey:AssignedCourierID" json:"assigned_courier,omitempty"`
	Assigner        Courier `gorm:"foreignKey:AssignedBy" json:"assigner,omitempty"`
}

// AssignmentHistoryResponse 分配历史响应
type AssignmentHistoryResponse struct {
	Assignments []TaskAssignmentHistory `json:"assignments"`
	Total       int                     `json:"total"`
	Page        int                     `json:"page"`
	Limit       int                     `json:"limit"`
}