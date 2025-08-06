/**
 * OpenPenPal SOTA权限系统 - 统一导出
 * 
 * 这是一个企业级的动态权限管理系统，支持：
 * - 29个细粒度权限模块
 * - 8种用户角色 + 4个信使等级
 * - 动态权限配置和实时生效
 * - 权限依赖检查和冲突管理
 * - 智能缓存和性能优化
 * - 完整的审计追踪
 */

package permissions

import "context"

// ================================
// 主要类型导出
// ================================

// 主要类型直接导入 - 避免重复定义

// ================================
// 常量导出
// ================================

const (
	// 用户角色
	UserRoleUser               = RoleUser
	UserRoleCourier           = RoleCourier
	UserRoleSeniorCourier     = RoleSeniorCourier
	UserRoleCourierCoordinator = RoleCourierCoordinator
	UserRoleSchoolAdmin       = RoleSchoolAdmin
	UserRolePlatformAdmin     = RolePlatformAdmin
	UserRoleAdmin             = RoleAdmin
	UserRoleSuperAdmin        = RoleSuperAdmin

	// 信使等级
	CourierLevelOne   = CourierLevel1
	CourierLevelTwo   = CourierLevel2
	CourierLevelThree = CourierLevel3
	CourierLevelFour  = CourierLevel4

	// 权限分类
	PermissionCategoryBasic      = CategoryBasic
	PermissionCategoryCourier    = CategoryCourier
	PermissionCategoryManagement = CategoryManagement
	PermissionCategoryAdmin      = CategoryAdmin
	PermissionCategorySystem     = CategorySystem

	// 风险等级
	RiskLevelLow      = RiskLow
	RiskLevelMedium   = RiskMedium
	RiskLevelHigh     = RiskHigh
	RiskLevelCritical = RiskCritical
)

// ================================
// 核心服务接口
// ================================

// PermissionServiceInterface 权限服务接口
type PermissionServiceInterface interface {
	// 基础权限检查
	HasPermission(user User, permission string) bool
	HasRolePermission(role UserRole, permission string) bool
	HasCourierLevelPermission(level CourierLevel, permission string) bool
	HasAnyPermission(user User, permissions []string) bool
	HasAllPermissions(user User, permissions []string) bool

	// 权限获取
	GetUserPermissions(user User) []string
	GetRolePermissions(role UserRole) []string
	GetCourierLevelPermissions(level CourierLevel) []string

	// 特殊权限检查
	CanAccessAdmin(user User) bool
	CanRoleAccessAdmin(role UserRole) bool
	IsCourier(user User) bool

	// 权限模块
	GetPermissionModule(permissionID string) *PermissionModule
	GetAllPermissionModules() map[string]*PermissionModule
	GetUserPermissionDetails(user User) []struct {
		ID      string            `json:"id"`
		Module  *PermissionModule `json:"module"`
		Granted bool              `json:"granted"`
	}

	// 配置管理
	UpdateRolePermissions(roleID UserRole, permissions []string, modifiedBy string) error
	UpdateCourierLevelPermissions(level CourierLevel, permissions []string, modifiedBy string) error
	ResetRolePermissions(roleID UserRole)
	ResetCourierLevelPermissions(level CourierLevel)

	// 配置查询
	GetRolePermissionConfig(roleID UserRole) *RolePermissionConfig
	GetCourierLevelPermissionConfig(level CourierLevel) *CourierLevelPermissionConfig
	GetAllCustomConfigs() (map[UserRole]*RolePermissionConfig, map[CourierLevel]*CourierLevelPermissionConfig)

	// 导入导出
	ExportConfigs() (string, error)
	ImportConfigs(configData string, overwrite bool) error

	// 权限检查和分析
	CheckPermission(user User, permission string) *PermissionCheckResult
	AnalyzeUserPermissions(user User) *PermissionAnalysis
	RefreshPermissions() error
}

// PermissionEnforcerInterface 权限执行器接口
type PermissionEnforcerInterface interface {
	// 动态权限检查
	CheckPermission(ctx context.Context, user User, permission string, options ...CheckOption) (bool, error)
	BatchCheckPermissions(ctx context.Context, user User, permissions []string, options ...CheckOption) (map[string]bool, error)
	EnforcePermission(ctx context.Context, user User, permission string, options ...CheckOption) error

	// 缓存管理
	ClearCache(user *User)
	HandlePermissionChange(event PermissionChangeEvent)

	// 监听器管理
	AddPermissionChangeListener(listener PermissionChangeListener)
	RemovePermissionChangeListener(listener PermissionChangeListener)

	// 权限分析
	AnalyzeUserPermissions(user User) *PermissionAnalysis
}

// ================================
// 工厂函数
// ================================

// NewPermissionService 创建权限服务实例
func NewPermissionService() PermissionServiceInterface {
	return NewService()
}

// NewPermissionEnforcer 创建权限执行器实例
func NewPermissionEnforcer(configManager *ConfigManager) PermissionEnforcerInterface {
	return NewEnforcer(configManager)
}

// ================================
// 全局默认实例
// ================================

var (
	// DefaultService 默认权限服务实例
	DefaultService PermissionServiceInterface = GetDefaultService()
)

// ================================
// 便捷函数
// ================================

// QuickCheck 快速权限检查
func QuickCheck(user User, permission string) bool {
	return DefaultService.HasPermission(user, permission)
}

// QuickCanAccessAdmin 快速管理权限检查
func QuickCanAccessAdmin(user User) bool {
	return DefaultService.CanAccessAdmin(user)
}

// QuickAnalyze 快速权限分析
func QuickAnalyze(user User) *PermissionAnalysis {
	return DefaultService.AnalyzeUserPermissions(user)
}

// ================================
// 权限常量定义
// ================================

const (
	// 基础权限
	PermissionReadLetter         = "READ_LETTER"
	PermissionWriteLetter        = "WRITE_LETTER"
	PermissionManageProfile      = "MANAGE_PROFILE"
	PermissionViewPlaza          = "VIEW_PLAZA"
	PermissionParticipatePlaza   = "PARTICIPATE_PLAZA"

	// 信使权限
	PermissionCourierScanCode      = "COURIER_SCAN_CODE"
	PermissionCourierDeliverLetter = "COURIER_DELIVER_LETTER"
	PermissionCourierViewTasks     = "COURIER_VIEW_TASKS"
	PermissionCourierUpdateStatus  = "COURIER_UPDATE_STATUS"
	PermissionCourierViewPoints    = "COURIER_VIEW_POINTS"

	// 管理权限
	PermissionManageSubordinates         = "MANAGE_SUBORDINATES"
	PermissionAssignTasks                = "ASSIGN_TASKS"
	PermissionViewRegionStats            = "VIEW_REGION_STATS"
	PermissionManagePostalCodes          = "MANAGE_POSTAL_CODES"
	PermissionApproveCourierApplications = "APPROVE_COURIER_APPLICATIONS"

	// 管理员权限
	PermissionManageUsers    = "MANAGE_USERS"
	PermissionManageLetters  = "MANAGE_LETTERS"
	PermissionManageCouriers = "MANAGE_COURIERS"
	PermissionManageSchools  = "MANAGE_SCHOOLS"
	PermissionViewAnalytics  = "VIEW_ANALYTICS"
	PermissionAuditLogs      = "AUDIT_LOGS"

	// 系统权限
	PermissionManageSystemSettings = "MANAGE_SYSTEM_SETTINGS"
	PermissionManagePermissions    = "MANAGE_PERMISSIONS"
	PermissionSystemAdmin          = "SYSTEM_ADMIN"
	PermissionDatabaseAccess       = "DATABASE_ACCESS"
	PermissionAPIAdmin             = "API_ADMIN"
)

// ================================
// 版本信息
// ================================

const (
	// Version 权限系统版本
	Version = "1.0.0"
	// BuildTime 构建时间
	BuildTime = "2025-07-25"
	// Description 系统描述
	Description = "OpenPenPal SOTA Dynamic Permission System"
)

// GetVersion 获取版本信息
func GetVersion() map[string]string {
	return map[string]string{
		"version":     Version,
		"build_time":  BuildTime,
		"description": Description,
	}
}

// ================================
// 使用示例
// ================================

/*
使用示例：

1. 基础权限检查：
```go
import "github.com/openpenpal/shared/go/pkg/permissions"

user := permissions.User{
    Role: permissions.UserRoleCourier,
    CourierInfo: &permissions.CourierInfo{
        Level: permissions.CourierLevelTwo,
    },
}

// 检查用户是否可以扫描信件
if permissions.QuickCheck(user, permissions.PermissionCourierScanCode) {
    // 用户有权限
}

// 检查用户是否可以访问管理后台
if permissions.QuickCanAccessAdmin(user) {
    // 用户可以访问管理后台
}
```

2. 高级权限检查：
```go
service := permissions.NewPermissionService()

// 批量检查权限
requiredPermissions := []string{
    permissions.PermissionCourierScanCode,
    permissions.PermissionCourierDeliverLetter,
}

hasAll := service.HasAllPermissions(user, requiredPermissions)
hasAny := service.HasAnyPermission(user, requiredPermissions)
```

3. 动态权限配置：
```go
// 更新角色权限
err := service.UpdateRolePermissions(
    permissions.UserRoleCourier,
    []string{permissions.PermissionCourierScanCode, permissions.PermissionManageSubordinates},
    "admin_user",
)

// 导出配置
configJSON, err := service.ExportConfigs()
```

4. 权限分析：
```go
analysis := permissions.QuickAnalyze(user)
fmt.Printf("权限覆盖率: %.2f%%\n", analysis.Coverage)
fmt.Printf("风险等级: %s\n", analysis.RiskLevel)
```
*/