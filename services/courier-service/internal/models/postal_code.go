package models

import (
	"time"
)

// PostalCodeApplication 编号申请模型
type PostalCodeApplication struct {
	ID            uint             `json:"id" gorm:"primaryKey"`
	UserID        string           `json:"user_id" gorm:"not null;index"`
	SchoolID      string           `json:"school_id" gorm:"not null"`
	SchoolName    string           `json:"school_name" gorm:"not null"`
	AreaID        string           `json:"area_id" gorm:"not null"`
	AreaName      string           `json:"area_name" gorm:"not null"`
	RequestedCode string           `json:"requested_code"`
	AssignedCode  string           `json:"assigned_code"`
	Status        PostalCodeStatus `json:"status" gorm:"default:pending"`
	ApplicantInfo string           `json:"applicant_info" gorm:"type:json"` // 申请人信息JSON
	Reason        string           `json:"reason"`                          // 申请理由
	Evidence      string           `json:"evidence" gorm:"type:json"`       // 证明材料JSON
	ReviewerID    *string          `json:"reviewer_id,omitempty"`
	ReviewedAt    *time.Time       `json:"reviewed_at,omitempty"`
	ReviewComment string           `json:"review_comment"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
}

// PostalCodeAssignment 编号分配记录模型
type PostalCodeAssignment struct {
	ID            uint       `json:"id" gorm:"primaryKey"`
	UserID        string     `json:"user_id" gorm:"not null;index"`
	PostalCode    string     `json:"postal_code" gorm:"not null;unique"`
	SchoolID      string     `json:"school_id" gorm:"not null"`
	AreaID        string     `json:"area_id" gorm:"not null"`
	BuildingID    *string    `json:"building_id,omitempty"`
	RoomNumber    *string    `json:"room_number,omitempty"`
	IsActive      bool       `json:"is_active" gorm:"default:true"`
	AssignedBy    string     `json:"assigned_by" gorm:"not null"` // 分配者ID
	AssignedAt    time.Time  `json:"assigned_at"`
	DeactivatedBy *string    `json:"deactivated_by,omitempty"`
	DeactivatedAt *time.Time `json:"deactivated_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// PostalCodeRule 编号规则模型
type PostalCodeRule struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	SchoolID   string    `json:"school_id" gorm:"not null;unique"`
	SchoolName string    `json:"school_name" gorm:"not null"`
	Prefix     string    `json:"prefix" gorm:"not null"`      // 学校前缀
	AreaRules  string    `json:"area_rules" gorm:"type:json"` // 片区规则JSON
	TotalCodes int       `json:"total_codes" gorm:"default:0"`
	UsedCodes  int       `json:"used_codes" gorm:"default:0"`
	IsActive   bool      `json:"is_active" gorm:"default:true"`
	CreatedBy  string    `json:"created_by"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// PostalCodeZone 编号管理区域模型
type PostalCodeZone struct {
	ID           uint            `json:"id" gorm:"primaryKey"`
	ZoneID       string          `json:"zone_id" gorm:"not null;unique"`
	ZoneName     string          `json:"zone_name" gorm:"not null"`
	ZoneType     CourierZoneType `json:"zone_type" gorm:"not null"`
	SchoolID     string          `json:"school_id" gorm:"not null"`
	ParentZone   *string         `json:"parent_zone,omitempty"`
	ManagerID    *string         `json:"manager_id,omitempty"`        // 负责管理的信使ID
	ManagerLevel *CourierLevel   `json:"manager_level,omitempty"`     // 管理者等级
	CodeRange    string          `json:"code_range" gorm:"type:json"` // 编号范围JSON
	IsActive     bool            `json:"is_active" gorm:"default:true"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// PostalCodeStatus 编号申请状态
type PostalCodeStatus string

const (
	PostalCodeStatusPending   PostalCodeStatus = "pending"   // 待审核
	PostalCodeStatusApproved  PostalCodeStatus = "approved"  // 已批准
	PostalCodeStatusRejected  PostalCodeStatus = "rejected"  // 已拒绝
	PostalCodeStatusAssigned  PostalCodeStatus = "assigned"  // 已分配
	PostalCodeStatusCanceled PostalCodeStatus = "canceled" // 已取消
)

// PostalCodeApplicationRequest 编号申请请求
type PostalCodeApplicationRequest struct {
	SchoolID      string                 `json:"school_id" binding:"required"`
	AreaID        string                 `json:"area_id" binding:"required"`
	RequestedCode string                 `json:"requested_code"`
	Reason        string                 `json:"reason" binding:"required"`
	ApplicantInfo map[string]interface{} `json:"applicant_info" binding:"required"`
	Evidence      map[string]interface{} `json:"evidence"`
}

// PostalCodeReviewRequest 编号审核请求
type PostalCodeReviewRequest struct {
	Action        string `json:"action" binding:"required,oneof=approve reject"`
	AssignedCode  string `json:"assigned_code"`
	ReviewComment string `json:"review_comment"`
}

// PostalCodeBatchAssignRequest 批量分配编号请求
type PostalCodeBatchAssignRequest struct {
	SchoolID    string                     `json:"school_id" binding:"required"`
	AreaID      string                     `json:"area_id" binding:"required"`
	Assignments []PostalCodeAssignmentItem `json:"assignments" binding:"required"`
}

// PostalCodeAssignmentItem 编号分配项
type PostalCodeAssignmentItem struct {
	UserID     string `json:"user_id" binding:"required"`
	PostalCode string `json:"postal_code" binding:"required"`
	BuildingID string `json:"building_id"`
	RoomNumber string `json:"room_number"`
}

// PostalCodeStatistics 编号统计信息
type PostalCodeStatistics struct {
	SchoolID        string  `json:"school_id"`
	SchoolName      string  `json:"school_name"`
	TotalCodes      int     `json:"total_codes"`
	AssignedCodes   int     `json:"assigned_codes"`
	UnassignedCodes int     `json:"unassigned_codes"`
	PendingApps     int     `json:"pending_applications"`
	UtilizationRate float64 `json:"utilization_rate"`
}

// PostalCodePermissionScope 编号权限范围
type PostalCodePermissionScope struct {
	CourierID  string       `json:"courier_id"`
	Level      CourierLevel `json:"level"`
	CanManage  []string     `json:"can_manage"`  // 可以管理的区域ID列表
	CanAssign  []string     `json:"can_assign"`  // 可以分配编号的区域ID列表
	CanApprove []string     `json:"can_approve"` // 可以审核申请的区域ID列表
	Schools    []SchoolInfo `json:"schools"`     // 可管理的学校列表
}

// SchoolInfo 学校信息
type SchoolInfo struct {
	SchoolID   string     `json:"school_id"`
	SchoolName string     `json:"school_name"`
	Areas      []AreaInfo `json:"areas"`
}

// AreaInfo 片区信息
type AreaInfo struct {
	AreaID    string         `json:"area_id"`
	AreaName  string         `json:"area_name"`
	Buildings []BuildingInfo `json:"buildings"`
	ManagerID string         `json:"manager_id"`
	CodeRange string         `json:"code_range"`
}

// BuildingInfo 楼栋信息
type BuildingInfo struct {
	BuildingID   string `json:"building_id"`
	BuildingName string `json:"building_name"`
	ManagerID    string `json:"manager_id"`
	CodeRange    string `json:"code_range"`
}

// GetStatusName 获取状态名称
func (s PostalCodeStatus) GetName() string {
	switch s {
	case PostalCodeStatusPending:
		return "待审核"
	case PostalCodeStatusApproved:
		return "已批准"
	case PostalCodeStatusRejected:
		return "已拒绝"
	case PostalCodeStatusAssigned:
		return "已分配"
	case PostalCodeStatusCanceled:
		return "已取消"
	default:
		return "未知状态"
	}
}

// CanTransitionTo 检查状态是否可以转换
func (s PostalCodeStatus) CanTransitionTo(target PostalCodeStatus) bool {
	validTransitions := map[PostalCodeStatus][]PostalCodeStatus{
		PostalCodeStatusPending: {
			PostalCodeStatusApproved,
			PostalCodeStatusRejected,
			PostalCodeStatusCanceled,
		},
		PostalCodeStatusApproved: {
			PostalCodeStatusAssigned,
			PostalCodeStatusCanceled,
		},
		PostalCodeStatusRejected: {
			PostalCodeStatusPending, // 可以重新申请
		},
		PostalCodeStatusAssigned: {
			// 已分配的编号一般不能改变状态，除非特殊情况
		},
		PostalCodeStatusCanceled: {
			PostalCodeStatusPending, // 可以重新申请
		},
	}

	allowed, exists := validTransitions[s]
	if !exists {
		return false
	}

	for _, allowedStatus := range allowed {
		if allowedStatus == target {
			return true
		}
	}
	return false
}

// DefaultSchoolRules 默认学校编号规则配置
var DefaultSchoolRules = map[string]PostalCodeRule{
	"beijing_university": {
		SchoolID:   "beijing_university",
		SchoolName: "北京大学",
		Prefix:     "PKU",
		AreaRules: `{
			"areas": [
				{"id": "area_01", "name": "燕园片区", "range": "001-199"},
				{"id": "area_02", "name": "万柳片区", "range": "200-299"},
				{"id": "area_03", "name": "圆明园片区", "range": "300-399"}
			]
		}`,
		TotalCodes: 999,
		IsActive:   true,
	},
	"tsinghua_university": {
		SchoolID:   "tsinghua_university",
		SchoolName: "清华大学",
		Prefix:     "THU",
		AreaRules: `{
			"areas": [
				{"id": "area_01", "name": "主校区", "range": "001-399"},
				{"id": "area_02", "name": "紫荆片区", "range": "400-599"},
				{"id": "area_03", "name": "东区", "range": "600-799"}
			]
		}`,
		TotalCodes: 999,
		IsActive:   true,
	},
}

// PostalCodePermissionMatrix 编号权限矩阵（基于信使等级）
var PostalCodePermissionMatrix = map[CourierLevel][]string{
	LevelOne: {
		"view_own_building", // 查看本楼栋编号分配
	},
	LevelTwo: {
		"view_own_building",
		"view_area",        // 查看片区编号分配
		"approve_building", // 审核楼栋级编号申请
		"assign_building",  // 分配楼栋级编号
	},
	LevelThree: {
		"view_own_building",
		"view_area",
		"view_campus", // 查看全校编号分配
		"approve_building",
		"approve_area", // 审核片区级编号申请
		"assign_building",
		"assign_area",  // 分配片区级编号
		"manage_rules", // 管理编号规则
	},
	LevelFour: {
		"view_own_building",
		"view_area",
		"view_campus",
		"view_city", // 查看全域编号分配
		"approve_building",
		"approve_area",
		"approve_campus", // 审核校级编号申请
		"assign_building",
		"assign_area",
		"assign_campus", // 分配校级编号
		"manage_rules",
		"manage_schools", // 管理学校编号规则
		"batch_assign",   // 批量分配编号
	},
}
