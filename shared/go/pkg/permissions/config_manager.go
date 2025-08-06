/**
 * 权限配置管理器 - SOTA动态权限配置
 */

package permissions

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// ConfigManager 权限配置管理器
type ConfigManager struct {
	customRoleConfigs         map[UserRole]*RolePermissionConfig
	customCourierLevelConfigs map[CourierLevel]*CourierLevelPermissionConfig
	defaultRoleConfigs        map[UserRole]*RolePermissionConfig
	defaultCourierLevelConfigs map[CourierLevel]*CourierLevelPermissionConfig
	mutex                     sync.RWMutex
}

// NewConfigManager 创建新的配置管理器
func NewConfigManager() *ConfigManager {
	cm := &ConfigManager{
		customRoleConfigs:         make(map[UserRole]*RolePermissionConfig),
		customCourierLevelConfigs: make(map[CourierLevel]*CourierLevelPermissionConfig),
		mutex:                     sync.RWMutex{},
	}
	
	cm.initializeDefaultConfigs()
	return cm
}

// initializeDefaultConfigs 初始化默认权限配置
func (cm *ConfigManager) initializeDefaultConfigs() {
	// 默认角色权限配置
	cm.defaultRoleConfigs = map[UserRole]*RolePermissionConfig{
		RoleUser: {
			RoleID: RoleUser,
			Permissions: []string{
				"READ_LETTER", "WRITE_LETTER", "MANAGE_PROFILE", 
				"VIEW_PLAZA", "PARTICIPATE_PLAZA",
			},
			IsCustom: false,
		},
		RoleCourier: {
			RoleID: RoleCourier,
			Permissions: []string{
				"READ_LETTER", "WRITE_LETTER", "MANAGE_PROFILE",
				"VIEW_PLAZA", "PARTICIPATE_PLAZA",
				"COURIER_SCAN_CODE", "COURIER_DELIVER_LETTER",
				"COURIER_VIEW_TASKS", "COURIER_UPDATE_STATUS", "COURIER_VIEW_POINTS",
			},
			IsCustom: false,
		},
		RoleSeniorCourier: {
			RoleID: RoleSeniorCourier,
			Permissions: []string{
				"READ_LETTER", "WRITE_LETTER", "MANAGE_PROFILE",
				"VIEW_PLAZA", "PARTICIPATE_PLAZA",
				"COURIER_SCAN_CODE", "COURIER_DELIVER_LETTER",
				"COURIER_VIEW_TASKS", "COURIER_UPDATE_STATUS", "COURIER_VIEW_POINTS",
				"MANAGE_SUBORDINATES", "ASSIGN_TASKS", "VIEW_REGION_STATS",
			},
			IsCustom: false,
		},
		RoleCourierCoordinator: {
			RoleID: RoleCourierCoordinator,
			Permissions: []string{
				"READ_LETTER", "WRITE_LETTER", "MANAGE_PROFILE",
				"VIEW_PLAZA", "PARTICIPATE_PLAZA",
				"COURIER_SCAN_CODE", "COURIER_DELIVER_LETTER",
				"COURIER_VIEW_TASKS", "COURIER_UPDATE_STATUS", "COURIER_VIEW_POINTS",
				"MANAGE_SUBORDINATES", "ASSIGN_TASKS", "VIEW_REGION_STATS",
				"MANAGE_POSTAL_CODES", "APPROVE_COURIER_APPLICATIONS",
			},
			IsCustom: false,
		},
		RoleSchoolAdmin: {
			RoleID: RoleSchoolAdmin,
			Permissions: []string{
				"READ_LETTER", "WRITE_LETTER", "MANAGE_PROFILE",
				"VIEW_PLAZA", "PARTICIPATE_PLAZA",
				"MANAGE_USERS", "MANAGE_LETTERS", "MANAGE_COURIERS",
				"MANAGE_SCHOOLS", "VIEW_ANALYTICS",
			},
			IsCustom: false,
		},
		RolePlatformAdmin: {
			RoleID: RolePlatformAdmin,
			Permissions: []string{
				"READ_LETTER", "WRITE_LETTER", "MANAGE_PROFILE",
				"VIEW_PLAZA", "PARTICIPATE_PLAZA",
				"MANAGE_USERS", "MANAGE_LETTERS", "MANAGE_COURIERS",
				"MANAGE_SCHOOLS", "VIEW_ANALYTICS", "AUDIT_LOGS",
				"MANAGE_SYSTEM_SETTINGS",
			},
			IsCustom: false,
		},
		RoleAdmin: {
			RoleID: RoleAdmin,
			Permissions: []string{
				"READ_LETTER", "WRITE_LETTER", "MANAGE_PROFILE",
				"VIEW_PLAZA", "PARTICIPATE_PLAZA",
				"MANAGE_USERS", "MANAGE_LETTERS", "MANAGE_COURIERS",
				"MANAGE_SCHOOLS", "VIEW_ANALYTICS", "AUDIT_LOGS",
				"MANAGE_SYSTEM_SETTINGS", "MANAGE_PERMISSIONS", "API_ADMIN",
			},
			IsCustom: false,
		},
		RoleSuperAdmin: {
			RoleID: RoleSuperAdmin,
			Permissions: []string{
				// 超级管理员拥有所有权限
			},
			IsCustom: false,
		},
	}

	// 默认信使等级权限配置
	cm.defaultCourierLevelConfigs = map[CourierLevel]*CourierLevelPermissionConfig{
		CourierLevel1: {
			Level: CourierLevel1,
			Permissions: []string{
				"COURIER_SCAN_CODE", "COURIER_DELIVER_LETTER",
				"COURIER_VIEW_TASKS", "COURIER_UPDATE_STATUS",
			},
			IsCustom: false,
		},
		CourierLevel2: {
			Level: CourierLevel2,
			Permissions: []string{
				"COURIER_SCAN_CODE", "COURIER_DELIVER_LETTER",
				"COURIER_VIEW_TASKS", "COURIER_UPDATE_STATUS", "COURIER_VIEW_POINTS",
				"MANAGE_SUBORDINATES", "VIEW_REGION_STATS",
			},
			IsCustom: false,
		},
		CourierLevel3: {
			Level: CourierLevel3,
			Permissions: []string{
				"COURIER_SCAN_CODE", "COURIER_DELIVER_LETTER",
				"COURIER_VIEW_TASKS", "COURIER_UPDATE_STATUS", "COURIER_VIEW_POINTS",
				"MANAGE_SUBORDINATES", "ASSIGN_TASKS", "VIEW_REGION_STATS",
				"MANAGE_POSTAL_CODES",
			},
			IsCustom: false,
		},
		CourierLevel4: {
			Level: CourierLevel4,
			Permissions: []string{
				"COURIER_SCAN_CODE", "COURIER_DELIVER_LETTER",
				"COURIER_VIEW_TASKS", "COURIER_UPDATE_STATUS", "COURIER_VIEW_POINTS",
				"MANAGE_SUBORDINATES", "ASSIGN_TASKS", "VIEW_REGION_STATS",
				"MANAGE_POSTAL_CODES", "APPROVE_COURIER_APPLICATIONS",
			},
			IsCustom: false,
		},
	}
}

// GetRolePermissions 获取角色权限
func (cm *ConfigManager) GetRolePermissions(role UserRole) []string {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	// 优先使用自定义配置
	if customConfig, exists := cm.customRoleConfigs[role]; exists && customConfig != nil {
		return append([]string(nil), customConfig.Permissions...)
	}

	// 使用默认配置
	if defaultConfig, exists := cm.defaultRoleConfigs[role]; exists && defaultConfig != nil {
		return append([]string(nil), defaultConfig.Permissions...)
	}

	return []string{}
}

// GetCourierLevelPermissions 获取信使等级权限
func (cm *ConfigManager) GetCourierLevelPermissions(level CourierLevel) []string {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	// 优先使用自定义配置
	if customConfig, exists := cm.customCourierLevelConfigs[level]; exists && customConfig != nil {
		return append([]string(nil), customConfig.Permissions...)
	}

	// 使用默认配置
	if defaultConfig, exists := cm.defaultCourierLevelConfigs[level]; exists && defaultConfig != nil {
		return append([]string(nil), defaultConfig.Permissions...)
	}

	return []string{}
}

// GetUserPermissions 获取用户的所有权限
func (cm *ConfigManager) GetUserPermissions(user User) []string {
	rolePermissions := cm.GetRolePermissions(user.Role)
	
	// 如果用户是信使，还需要加上等级权限
	if user.CourierInfo != nil {
		courierPermissions := cm.GetCourierLevelPermissions(user.CourierInfo.Level)
		
		// 合并权限并去重
		permissionSet := make(map[string]bool)
		for _, p := range rolePermissions {
			permissionSet[p] = true
		}
		for _, p := range courierPermissions {
			permissionSet[p] = true
		}
		
		result := make([]string, 0, len(permissionSet))
		for p := range permissionSet {
			result = append(result, p)
		}
		return result
	}
	
	return rolePermissions
}

// HasPermission 检查用户是否拥有特定权限
func (cm *ConfigManager) HasPermission(user User, permission string) bool {
	permissions := cm.GetUserPermissions(user)
	for _, p := range permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// HasAnyPermission 检查用户是否拥有任一权限
func (cm *ConfigManager) HasAnyPermission(user User, permissions []string) bool {
	userPermissions := cm.GetUserPermissions(user)
	userPermSet := make(map[string]bool)
	for _, p := range userPermissions {
		userPermSet[p] = true
	}
	
	for _, p := range permissions {
		if userPermSet[p] {
			return true
		}
	}
	return false
}

// HasAllPermissions 检查用户是否拥有所有权限
func (cm *ConfigManager) HasAllPermissions(user User, permissions []string) bool {
	userPermissions := cm.GetUserPermissions(user)
	userPermSet := make(map[string]bool)
	for _, p := range userPermissions {
		userPermSet[p] = true
	}
	
	for _, p := range permissions {
		if !userPermSet[p] {
			return false
		}
	}
	return true
}

// UpdateRolePermissions 更新角色权限
func (cm *ConfigManager) UpdateRolePermissions(role UserRole, permissions []string, modifiedBy string) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// 验证权限依赖关系
	if err := ValidatePermissionDependencies(permissions); err != nil {
		return err
	}

	cm.customRoleConfigs[role] = &RolePermissionConfig{
		RoleID:      role,
		Permissions: append([]string(nil), permissions...),
		ModifiedBy:  modifiedBy,
		ModifiedAt:  time.Now(),
		IsCustom:    true,
	}

	return nil
}

// UpdateCourierLevelPermissions 更新信使等级权限
func (cm *ConfigManager) UpdateCourierLevelPermissions(level CourierLevel, permissions []string, modifiedBy string) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// 验证权限依赖关系
	if err := ValidatePermissionDependencies(permissions); err != nil {
		return err
	}

	cm.customCourierLevelConfigs[level] = &CourierLevelPermissionConfig{
		Level:       level,
		Permissions: append([]string(nil), permissions...),
		ModifiedBy:  modifiedBy,
		ModifiedAt:  time.Now(),
		IsCustom:    true,
	}

	return nil
}

// ResetRolePermissions 重置角色权限为默认值
func (cm *ConfigManager) ResetRolePermissions(role UserRole) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	delete(cm.customRoleConfigs, role)
}

// ResetCourierLevelPermissions 重置信使等级权限为默认值
func (cm *ConfigManager) ResetCourierLevelPermissions(level CourierLevel) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	delete(cm.customCourierLevelConfigs, level)
}

// GetRolePermissionConfig 获取角色权限配置详情
func (cm *ConfigManager) GetRolePermissionConfig(role UserRole) *RolePermissionConfig {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	if customConfig, exists := cm.customRoleConfigs[role]; exists {
		return customConfig
	}

	if defaultConfig, exists := cm.defaultRoleConfigs[role]; exists {
		return defaultConfig
	}

	return nil
}

// GetCourierLevelPermissionConfig 获取信使等级权限配置详情
func (cm *ConfigManager) GetCourierLevelPermissionConfig(level CourierLevel) *CourierLevelPermissionConfig {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	if customConfig, exists := cm.customCourierLevelConfigs[level]; exists {
		return customConfig
	}

	if defaultConfig, exists := cm.defaultCourierLevelConfigs[level]; exists {
		return defaultConfig
	}

	return nil
}

// GetAllCustomConfigs 获取所有自定义配置
func (cm *ConfigManager) GetAllCustomConfigs() (map[UserRole]*RolePermissionConfig, map[CourierLevel]*CourierLevelPermissionConfig) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	roleConfigs := make(map[UserRole]*RolePermissionConfig)
	for k, v := range cm.customRoleConfigs {
		roleConfigs[k] = v
	}

	courierConfigs := make(map[CourierLevel]*CourierLevelPermissionConfig)
	for k, v := range cm.customCourierLevelConfigs {
		courierConfigs[k] = v
	}

	return roleConfigs, courierConfigs
}

// ExportConfigs 导出权限配置
func (cm *ConfigManager) ExportConfigs() (string, error) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	roleConfigs, courierConfigs := cm.GetAllCustomConfigs()

	exportData := struct {
		Version             string                                               `json:"version"`
		Timestamp           time.Time                                            `json:"timestamp"`
		RoleConfigs         map[UserRole]*RolePermissionConfig                   `json:"roleConfigs"`
		CourierLevelConfigs map[CourierLevel]*CourierLevelPermissionConfig       `json:"courierLevelConfigs"`
	}{
		Version:             "1.0",
		Timestamp:           time.Now(),
		RoleConfigs:         roleConfigs,
		CourierLevelConfigs: courierConfigs,
	}

	data, err := json.MarshalIndent(exportData, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to export configs: %w", err)
	}

	return string(data), nil
}

// ImportConfigs 导入权限配置
func (cm *ConfigManager) ImportConfigs(configData string, overwrite bool) error {
	var importData struct {
		Version             string                                               `json:"version"`
		Timestamp           time.Time                                            `json:"timestamp"`
		RoleConfigs         map[UserRole]*RolePermissionConfig                   `json:"roleConfigs"`
		CourierLevelConfigs map[CourierLevel]*CourierLevelPermissionConfig       `json:"courierLevelConfigs"`
	}

	if err := json.Unmarshal([]byte(configData), &importData); err != nil {
		return fmt.Errorf("failed to parse config data: %w", err)
	}

	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if overwrite {
		cm.customRoleConfigs = make(map[UserRole]*RolePermissionConfig)
		cm.customCourierLevelConfigs = make(map[CourierLevel]*CourierLevelPermissionConfig)
	}

	// 导入角色配置
	for role, config := range importData.RoleConfigs {
		if err := ValidatePermissionDependencies(config.Permissions); err != nil {
			return fmt.Errorf("invalid permissions for role %s: %w", role, err)
		}
		cm.customRoleConfigs[role] = config
	}

	// 导入信使等级配置
	for level, config := range importData.CourierLevelConfigs {
		if err := ValidatePermissionDependencies(config.Permissions); err != nil {
			return fmt.Errorf("invalid permissions for courier level %d: %w", level, err)
		}
		cm.customCourierLevelConfigs[level] = config
	}

	return nil
}

// Refresh 刷新配置（可用于从外部存储重新加载）
func (cm *ConfigManager) Refresh() error {
	// 在实际应用中，这里可以从数据库或其他存储加载配置
	// 目前只是重新初始化默认配置
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	
	cm.initializeDefaultConfigs()
	return nil
}