/**
 * 权限服务 - SOTA权限检查和管理接口
 */

package permissions

import (
	"sync"
)

// Service 权限服务
type Service struct {
	configManager *ConfigManager
	enforcer      *Enforcer
	mutex         sync.RWMutex
}

// NewService 创建新的权限服务
func NewService() *Service {
	configManager := NewConfigManager()
	return &Service{
		configManager: configManager,
		enforcer:      NewEnforcer(configManager),
		mutex:         sync.RWMutex{},
	}
}

// HasPermission 检查用户是否拥有特定权限
func (s *Service) HasPermission(user User, permission string) bool {
	// 超级管理员拥有所有权限
	if user.Role == RoleSuperAdmin {
		return true
	}

	return s.configManager.HasPermission(user, permission)
}

// HasRolePermission 检查角色是否拥有特定权限（静态检查）
func (s *Service) HasRolePermission(role UserRole, permission string) bool {
	if role == RoleSuperAdmin {
		return true
	}

	user := User{Role: role}
	return s.configManager.HasPermission(user, permission)
}

// HasCourierLevelPermission 检查信使等级是否拥有特定权限
func (s *Service) HasCourierLevelPermission(level CourierLevel, permission string) bool {
	permissions := s.configManager.GetCourierLevelPermissions(level)
	for _, p := range permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// GetUserPermissions 获取用户所有权限
func (s *Service) GetUserPermissions(user User) []string {
	return s.configManager.GetUserPermissions(user)
}

// GetRolePermissions 获取角色所有权限
func (s *Service) GetRolePermissions(role UserRole) []string {
	return s.configManager.GetRolePermissions(role)
}

// GetCourierLevelPermissions 获取信使等级所有权限
func (s *Service) GetCourierLevelPermissions(level CourierLevel) []string {
	return s.configManager.GetCourierLevelPermissions(level)
}

// HasAnyPermission 检查用户是否拥有任一权限
func (s *Service) HasAnyPermission(user User, permissions []string) bool {
	if user.Role == RoleSuperAdmin {
		return true
	}

	return s.configManager.HasAnyPermission(user, permissions)
}

// HasAllPermissions 检查用户是否拥有所有权限
func (s *Service) HasAllPermissions(user User, permissions []string) bool {
	if user.Role == RoleSuperAdmin {
		return true
	}

	return s.configManager.HasAllPermissions(user, permissions)
}

// CanAccessAdmin 检查用户是否可以访问管理后台
func (s *Service) CanAccessAdmin(user User) bool {
	// 管理员角色直接有权限
	adminRoles := []UserRole{RoleSchoolAdmin, RolePlatformAdmin, RoleAdmin, RoleSuperAdmin}
	for _, role := range adminRoles {
		if user.Role == role {
			return true
		}
	}

	// 高级信使角色有权限
	seniorCourierRoles := []UserRole{RoleSeniorCourier, RoleCourierCoordinator}
	for _, role := range seniorCourierRoles {
		if user.Role == role {
			return true
		}
	}

	// 高等级信使有权限
	if user.CourierInfo != nil && user.CourierInfo.Level > 1 {
		return true
	}

	// 基于权限模块检查
	adminPermissions := []string{
		"MANAGE_USERS", "MANAGE_LETTERS", "MANAGE_COURIERS",
		"MANAGE_SCHOOLS", "MANAGE_SYSTEM_SETTINGS", "VIEW_ANALYTICS",
	}

	return s.HasAnyPermission(user, adminPermissions)
}

// CanRoleAccessAdmin 检查角色是否可以访问管理后台（静态检查）
func (s *Service) CanRoleAccessAdmin(role UserRole) bool {
	adminRoles := []UserRole{RoleSchoolAdmin, RolePlatformAdmin, RoleAdmin, RoleSuperAdmin}
	seniorCourierRoles := []UserRole{RoleSeniorCourier, RoleCourierCoordinator}

	for _, adminRole := range adminRoles {
		if role == adminRole {
			return true
		}
	}

	for _, seniorRole := range seniorCourierRoles {
		if role == seniorRole {
			return true
		}
	}

	return false
}

// IsCourier 检查用户是否为信使
func (s *Service) IsCourier(user User) bool {
	courierRoles := []UserRole{RoleCourier, RoleSeniorCourier, RoleCourierCoordinator}
	for _, role := range courierRoles {
		if user.Role == role {
			return true
		}
	}

	return user.CourierInfo != nil
}

// GetPermissionModule 获取权限模块信息
func (s *Service) GetPermissionModule(permissionID string) *PermissionModule {
	return GetPermissionModule(permissionID)
}

// GetAllPermissionModules 获取所有权限模块
func (s *Service) GetAllPermissionModules() map[string]*PermissionModule {
	return GetPermissionModules()
}

// GetUserPermissionDetails 获取用户权限详情（包含权限模块信息）
func (s *Service) GetUserPermissionDetails(user User) []struct {
	ID      string            `json:"id"`
	Module  *PermissionModule `json:"module"`
	Granted bool              `json:"granted"`
} {
	permissions := s.GetUserPermissions(user)
	modules := GetPermissionModules()

	var result []struct {
		ID      string            `json:"id"`
		Module  *PermissionModule `json:"module"`
		Granted bool              `json:"granted"`
	}

	for _, permissionID := range permissions {
		if module, exists := modules[permissionID]; exists {
			result = append(result, struct {
				ID      string            `json:"id"`
				Module  *PermissionModule `json:"module"`
				Granted bool              `json:"granted"`
			}{
				ID:      permissionID,
				Module:  module,
				Granted: true,
			})
		}
	}

	return result
}

// RefreshPermissions 刷新权限配置
func (s *Service) RefreshPermissions() error {
	return s.configManager.Refresh()
}

// ================================
// 权限配置管理（管理员专用）
// ================================

// UpdateRolePermissions 更新角色权限配置
func (s *Service) UpdateRolePermissions(roleID UserRole, permissions []string, modifiedBy string) error {
	return s.configManager.UpdateRolePermissions(roleID, permissions, modifiedBy)
}

// UpdateCourierLevelPermissions 更新信使等级权限配置
func (s *Service) UpdateCourierLevelPermissions(level CourierLevel, permissions []string, modifiedBy string) error {
	return s.configManager.UpdateCourierLevelPermissions(level, permissions, modifiedBy)
}

// ResetRolePermissions 重置角色权限为默认配置
func (s *Service) ResetRolePermissions(roleID UserRole) {
	s.configManager.ResetRolePermissions(roleID)
}

// ResetCourierLevelPermissions 重置信使等级权限为默认配置
func (s *Service) ResetCourierLevelPermissions(level CourierLevel) {
	s.configManager.ResetCourierLevelPermissions(level)
}

// GetRolePermissionConfig 获取角色权限配置信息
func (s *Service) GetRolePermissionConfig(roleID UserRole) *RolePermissionConfig {
	return s.configManager.GetRolePermissionConfig(roleID)
}

// GetCourierLevelPermissionConfig 获取信使等级权限配置信息
func (s *Service) GetCourierLevelPermissionConfig(level CourierLevel) *CourierLevelPermissionConfig {
	return s.configManager.GetCourierLevelPermissionConfig(level)
}

// GetAllCustomConfigs 获取所有自定义配置
func (s *Service) GetAllCustomConfigs() (map[UserRole]*RolePermissionConfig, map[CourierLevel]*CourierLevelPermissionConfig) {
	return s.configManager.GetAllCustomConfigs()
}

// ExportConfigs 导出权限配置
func (s *Service) ExportConfigs() (string, error) {
	return s.configManager.ExportConfigs()
}

// ImportConfigs 导入权限配置
func (s *Service) ImportConfigs(configData string, overwrite bool) error {
	return s.configManager.ImportConfigs(configData, overwrite)
}

// CheckPermission 权限检查（带结果详情）
func (s *Service) CheckPermission(user User, permission string) *PermissionCheckResult {
	module := GetPermissionModule(permission)
	riskLevel := RiskLow
	if module != nil {
		riskLevel = module.RiskLevel
	}

	granted := s.HasPermission(user, permission)
	reason := ""
	if !granted {
		reason = "Permission denied"
	}

	return &PermissionCheckResult{
		Granted:    granted,
		Permission: permission,
		User:       user,
		RiskLevel:  riskLevel,
		Reason:     reason,
	}
}

// AnalyzeUserPermissions 分析用户权限覆盖率
func (s *Service) AnalyzeUserPermissions(user User) *PermissionAnalysis {
	userPermissions := s.GetUserPermissions(user)
	allModules := GetPermissionModules()
	allPermissions := make([]string, 0, len(allModules))
	for id := range allModules {
		allPermissions = append(allPermissions, id)
	}

	coverage := float64(len(userPermissions)) / float64(len(allPermissions)) * 100

	missingPermissions := make([]string, 0)
	userPermSet := make(map[string]bool)
	for _, p := range userPermissions {
		userPermSet[p] = true
	}
	for _, p := range allPermissions {
		if !userPermSet[p] {
			missingPermissions = append(missingPermissions, p)
		}
	}

	permissionsByCategory := make(map[PermissionCategory][]string)
	for _, permission := range userPermissions {
		module := allModules[permission]
		if module != nil {
			permissionsByCategory[module.Category] = append(permissionsByCategory[module.Category], permission)
		}
	}

	riskLevel := CalculatePermissionRiskLevel(userPermissions)

	return &PermissionAnalysis{
		TotalPermissions:      len(allPermissions),
		GrantedPermissions:    len(userPermissions),
		Coverage:              coverage,
		MissingPermissions:    missingPermissions,
		PermissionsByCategory: permissionsByCategory,
		RiskLevel:            riskLevel,
	}
}

// ================================
// 全局权限服务实例
// ================================

var (
	defaultService *Service
	serviceOnce    sync.Once
)

// GetDefaultService 获取默认权限服务实例
func GetDefaultService() *Service {
	serviceOnce.Do(func() {
		defaultService = NewService()
	})
	return defaultService
}

// ================================
// 便捷函数
// ================================

// HasPermission 便捷权限检查函数
func HasPermission(user User, permission string) bool {
	return GetDefaultService().HasPermission(user, permission)
}

// CanAccessAdmin 便捷管理权限检查函数
func CanAccessAdmin(user User) bool {
	return GetDefaultService().CanAccessAdmin(user)
}

// CheckPermission 便捷权限检查函数（带详情）
func CheckPermission(user User, permission string) *PermissionCheckResult {
	return GetDefaultService().CheckPermission(user, permission)
}