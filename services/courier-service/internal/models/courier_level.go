package models

import (
	"time"
)

// CourierLevel 信使等级枚举
type CourierLevel int

const (
	LevelOne   CourierLevel = 1 // 一级信使：楼栋级
	LevelTwo   CourierLevel = 2 // 二级信使：片区级
	LevelThree CourierLevel = 3 // 三级信使：校区级
	LevelFour  CourierLevel = 4 // 四级信使：城市级
)

// CourierPermission 信使权限类型枚举
type CourierPermission string

const (
	PermissionScan          CourierPermission = "scan"           // 扫码登记权限
	PermissionStatusChange  CourierPermission = "status_change"  // 状态变更权限
	PermissionHandover      CourierPermission = "handover"       // 向上转交权限
	PermissionPackage       CourierPermission = "package"        // 打包分拣权限
	PermissionDistribute    CourierPermission = "distribute"     // 信封分发权限
	PermissionReceiveHandover CourierPermission = "receive_handover" // 接收转交权限
	PermissionFeedback      CourierPermission = "feedback"       // 用户反馈处理权限
	PermissionPerformance   CourierPermission = "performance"    // 绩效查看权限
)

// CourierZoneType 管理区域类型
type CourierZoneType string

const (
	ZoneBuilding CourierZoneType = "building" // 楼栋
	ZoneArea     CourierZoneType = "area"     // 片区
	ZoneCampus   CourierZoneType = "campus"   // 校区
	ZoneCity     CourierZoneType = "city"     // 城市
)

// CourierLevel 信使等级模型
type CourierLevelModel struct {
	ID          uint            `json:"id" gorm:"primaryKey"`
	Level       CourierLevel    `json:"level" gorm:"not null;unique"`
	Name        string          `json:"name" gorm:"not null"`
	Description string          `json:"description"`
	ZoneType    CourierZoneType `json:"zone_type" gorm:"not null"`
	Permissions []string        `json:"permissions" gorm:"type:json"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// CourierPermissionModel 权限配置模型
type CourierPermissionModel struct {
	ID          uint              `json:"id" gorm:"primaryKey"`
	Permission  CourierPermission `json:"permission" gorm:"not null;unique"`
	Name        string            `json:"name" gorm:"not null"`
	Description string            `json:"description"`
	Levels      []int             `json:"levels" gorm:"type:json"` // 允许的等级列表
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// CourierZone 信使管理区域模型
type CourierZone struct {
	ID         uint            `json:"id" gorm:"primaryKey"`
	CourierID  string          `json:"courier_id" gorm:"not null;index"`
	ZoneType   CourierZoneType `json:"zone_type" gorm:"not null"`
	ZoneID     string          `json:"zone_id" gorm:"not null"`
	ZoneName   string          `json:"zone_name" gorm:"not null"`
	ParentZone *string         `json:"parent_zone,omitempty"`
	IsActive   bool            `json:"is_active" gorm:"default:true"`
	AssignedAt time.Time       `json:"assigned_at"`
	AssignedBy string          `json:"assigned_by"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

// LevelUpgradeRequest 等级升级申请模型
type LevelUpgradeRequest struct {
	ID            uint         `json:"id" gorm:"primaryKey"`
	CourierID     string       `json:"courier_id" gorm:"not null;index"`
	CurrentLevel  CourierLevel `json:"current_level" gorm:"not null"`
	RequestLevel  CourierLevel `json:"request_level" gorm:"not null"`
	Reason        string       `json:"reason"`
	Evidence      string       `json:"evidence"` // JSON格式的证据数据
	Status        string       `json:"status" gorm:"default:pending"` // pending,approved,rejected
	ReviewedBy    *string      `json:"reviewed_by,omitempty"`
	ReviewedAt    *time.Time   `json:"reviewed_at,omitempty"`
	ReviewComment string       `json:"review_comment"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
}

// CourierPermissionCheck 权限检查结果
type CourierPermissionCheck struct {
	CourierID    string              `json:"courier_id"`
	Level        CourierLevel        `json:"level"`
	Permission   CourierPermission   `json:"permission"`
	ZoneType     CourierZoneType     `json:"zone_type"`
	ZoneID       string              `json:"zone_id"`
	HasPermission bool               `json:"has_permission"`
	Reason       string              `json:"reason,omitempty"`
}

// GetLevelName 获取等级名称
func (l CourierLevel) GetLevelName() string {
	switch l {
	case LevelOne:
		return "一级信使"
	case LevelTwo:
		return "二级信使"
	case LevelThree:
		return "三级信使"
	case LevelFour:
		return "四级信使"
	default:
		return "未知等级"
	}
}

// GetZoneTypeName 获取区域类型名称
func (z CourierZoneType) GetZoneTypeName() string {
	switch z {
	case ZoneBuilding:
		return "楼栋"
	case ZoneArea:
		return "片区"
	case ZoneCampus:
		return "校区"
	case ZoneCity:
		return "城市"
	default:
		return "未知区域"
	}
}

// GetPermissionName 获取权限名称
func (p CourierPermission) GetPermissionName() string {
	switch p {
	case PermissionScan:
		return "扫码登记"
	case PermissionStatusChange:
		return "状态变更"
	case PermissionHandover:
		return "向上转交"
	case PermissionPackage:
		return "打包分拣"
	case PermissionDistribute:
		return "信封分发"
	case PermissionReceiveHandover:
		return "接收转交"
	case PermissionFeedback:
		return "用户反馈处理"
	case PermissionPerformance:
		return "绩效查看"
	default:
		return "未知权限"
	}
}

// DefaultPermissionMatrix 默认权限矩阵配置 (基于PRD权限矩阵)
var DefaultPermissionMatrix = map[CourierLevel][]CourierPermission{
	LevelOne: {
		PermissionScan,        // 本楼栋扫码登记
		PermissionStatusChange, // 状态变更
		PermissionHandover,    // 向上转交
	},
	LevelTwo: {
		PermissionScan,            // 片区扫码登记
		PermissionStatusChange,    // 状态变更
		PermissionPackage,         // 打包分拣
		PermissionDistribute,      // 信封分发
		PermissionReceiveHandover, // 接收一级转交
		PermissionHandover,        // 向上转交
	},
	LevelThree: {
		PermissionScan,            // 全校扫码登记
		PermissionStatusChange,    // 状态变更
		PermissionPackage,         // 打包分拣
		PermissionDistribute,      // 信封分发
		PermissionReceiveHandover, // 接收下级转交
		PermissionFeedback,        // 用户反馈处理
		PermissionPerformance,     // 校级绩效查看
		PermissionHandover,        // 向上转交
	},
	LevelFour: {
		PermissionScan,            // 全域扫码登记
		PermissionStatusChange,    // 状态变更
		PermissionPackage,         // 打包分拣
		PermissionDistribute,      // 信封分发
		PermissionReceiveHandover, // 接收下级转交
		PermissionFeedback,        // 用户反馈处理
		PermissionPerformance,     // 全域绩效查看
		// 注意：四级信使不可向上转交 (最高等级)
	},
}

// DefaultZoneMapping 默认区域对应关系
var DefaultZoneMapping = map[CourierLevel]CourierZoneType{
	LevelOne:   ZoneBuilding,
	LevelTwo:   ZoneArea,
	LevelThree: ZoneCampus,
	LevelFour:  ZoneCity,
}