/**
 * 权限系统类型定义 - SOTA权限管理核心类型
 */

package permissions

import (
	"time"
)

// UserRole 用户角色枚举
type UserRole string

const (
	RoleUser               UserRole = "user"
	RoleCourier           UserRole = "courier"
	RoleSeniorCourier     UserRole = "senior_courier"
	RoleCourierCoordinator UserRole = "courier_coordinator"
	RoleSchoolAdmin       UserRole = "school_admin"
	RolePlatformAdmin     UserRole = "platform_admin"
	RoleAdmin             UserRole = "admin"
	RoleSuperAdmin        UserRole = "super_admin"
)

// CourierLevel 信使等级
type CourierLevel int

const (
	CourierLevel1 CourierLevel = 1
	CourierLevel2 CourierLevel = 2
	CourierLevel3 CourierLevel = 3
	CourierLevel4 CourierLevel = 4
)

// PermissionCategory 权限分类
type PermissionCategory string

const (
	CategoryBasic      PermissionCategory = "basic"
	CategoryCourier    PermissionCategory = "courier"
	CategoryManagement PermissionCategory = "management"
	CategoryAdmin      PermissionCategory = "admin"
	CategorySystem     PermissionCategory = "system"
)

// RiskLevel 风险等级
type RiskLevel string

const (
	RiskLow      RiskLevel = "low"
	RiskMedium   RiskLevel = "medium"
	RiskHigh     RiskLevel = "high"
	RiskCritical RiskLevel = "critical"
)

// PermissionModule 权限模块定义
type PermissionModule struct {
	ID           string             `json:"id"`
	Name         string             `json:"name"`
	Description  string             `json:"description"`
	Category     PermissionCategory `json:"category"`
	RiskLevel    RiskLevel          `json:"riskLevel"`
	Dependencies []string           `json:"dependencies,omitempty"`
	Conflicts    []string           `json:"conflicts,omitempty"`
	IsSystemCore bool               `json:"isSystemCore,omitempty"`
}

// User 用户信息
type User struct {
	Role        UserRole     `json:"role"`
	CourierInfo *CourierInfo `json:"courierInfo,omitempty"`
}

// CourierInfo 信使信息
type CourierInfo struct {
	Level CourierLevel `json:"level"`
}

// RolePermissionConfig 角色权限配置
type RolePermissionConfig struct {
	RoleID      UserRole  `json:"roleId"`
	Permissions []string  `json:"permissions"`
	ModifiedBy  string    `json:"modifiedBy"`
	ModifiedAt  time.Time `json:"modifiedAt"`
	IsCustom    bool      `json:"isCustom"`
}

// CourierLevelPermissionConfig 信使等级权限配置
type CourierLevelPermissionConfig struct {
	Level       CourierLevel `json:"level"`
	Permissions []string     `json:"permissions"`
	ModifiedBy  string       `json:"modifiedBy"`
	ModifiedAt  time.Time    `json:"modifiedAt"`
	IsCustom    bool         `json:"isCustom"`
}

// PermissionCheckResult 权限检查结果
type PermissionCheckResult struct {
	Granted     bool     `json:"granted"`
	Permission  string   `json:"permission"`
	User        User     `json:"user"`
	RiskLevel   RiskLevel `json:"riskLevel,omitempty"`
	Reason      string   `json:"reason,omitempty"`
}

// PermissionChangeEvent 权限变更事件
type PermissionChangeEvent struct {
	Type       string    `json:"type"`
	Target     string    `json:"target"`
	TargetType string    `json:"targetType"`
	ModifiedBy string    `json:"modifiedBy"`
	Timestamp  time.Time `json:"timestamp"`
	Changes    *struct {
		Added   []string `json:"added"`
		Removed []string `json:"removed"`
	} `json:"changes,omitempty"`
}

// PermissionError 权限错误
type PermissionError struct {
	Message    string   `json:"message"`
	Permission string   `json:"permission"`
	UserRole   UserRole `json:"userRole"`
	Code       string   `json:"code"`
}

func (e *PermissionError) Error() string {
	return e.Message
}

// PermissionAnalysis 权限分析结果
type PermissionAnalysis struct {
	TotalPermissions        int                               `json:"totalPermissions"`
	GrantedPermissions      int                               `json:"grantedPermissions"`
	Coverage                float64                           `json:"coverage"`
	MissingPermissions      []string                          `json:"missingPermissions"`
	PermissionsByCategory   map[PermissionCategory][]string   `json:"permissionsByCategory"`
	RiskLevel              RiskLevel                         `json:"riskLevel"`
}