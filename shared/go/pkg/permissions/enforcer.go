/**
 * 权限执行器 - SOTA动态权限生效机制
 */

package permissions

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Enforcer 权限执行器
type Enforcer struct {
	configManager   *ConfigManager
	permissionCache map[string]*cacheEntry
	cacheTimeout    time.Duration
	mutex           sync.RWMutex
	listeners       []PermissionChangeListener
}

// cacheEntry 缓存条目
type cacheEntry struct {
	permissions []string
	timestamp   time.Time
}

// PermissionChangeListener 权限变更监听器
type PermissionChangeListener func(event PermissionChangeEvent)

// NewEnforcer 创建新的权限执行器
func NewEnforcer(configManager *ConfigManager) *Enforcer {
	return &Enforcer{
		configManager:   configManager,
		permissionCache: make(map[string]*cacheEntry),
		cacheTimeout:    5 * time.Minute, // 5分钟缓存
		mutex:           sync.RWMutex{},
		listeners:       make([]PermissionChangeListener, 0),
	}
}

// CheckPermission 检查用户权限 - 支持缓存和实时刷新
func (e *Enforcer) CheckPermission(ctx context.Context, user User, permission string, options ...CheckOption) (bool, error) {
	// 超级管理员拥有所有权限
	if user.Role == RoleSuperAdmin {
		return true, nil
	}

	opts := &checkOptions{}
	for _, opt := range options {
		opt(opts)
	}

	cacheKey := e.generateCacheKey(user)

	// 检查缓存（除非强制刷新）
	if !opts.forceRefresh {
		if cached := e.getCachedPermissions(cacheKey); cached != nil {
			return e.containsPermission(cached, permission), nil
		}
	}

	// 获取最新权限
	permissions := e.refreshUserPermissions(user)
	return e.containsPermission(permissions, permission), nil
}

// BatchCheckPermissions 批量检查权限
func (e *Enforcer) BatchCheckPermissions(ctx context.Context, user User, permissions []string, options ...CheckOption) (map[string]bool, error) {
	result := make(map[string]bool)

	// 超级管理员拥有所有权限
	if user.Role == RoleSuperAdmin {
		for _, p := range permissions {
			result[p] = true
		}
		return result, nil
	}

	opts := &checkOptions{}
	for _, opt := range options {
		opt(opts)
	}

	userPermissions := e.getUserPermissions(user, opts.forceRefresh)

	for _, permission := range permissions {
		result[permission] = e.containsPermission(userPermissions, permission)
	}

	return result, nil
}

// EnforcePermission 权限守卫 - 如果没有权限则返回错误
func (e *Enforcer) EnforcePermission(ctx context.Context, user User, permission string, options ...CheckOption) error {
	hasPermission, err := e.CheckPermission(ctx, user, permission, options...)
	if err != nil {
		return err
	}

	if !hasPermission {
		return &PermissionError{
			Message:    fmt.Sprintf("Permission denied: %s", permission),
			Permission: permission,
			UserRole:   user.Role,
			Code:       "PERMISSION_DENIED",
		}
	}

	return nil
}

// getUserPermissions 获取用户权限（带缓存）
func (e *Enforcer) getUserPermissions(user User, forceRefresh bool) []string {
	cacheKey := e.generateCacheKey(user)

	if !forceRefresh {
		if cached := e.getCachedPermissions(cacheKey); cached != nil {
			return cached
		}
	}

	return e.refreshUserPermissions(user)
}

// generateCacheKey 生成缓存键
func (e *Enforcer) generateCacheKey(user User) string {
	if user.CourierInfo != nil {
		return fmt.Sprintf("%s_%d", user.Role, user.CourierInfo.Level)
	}
	return fmt.Sprintf("%s_no_level", user.Role)
}

// getCachedPermissions 获取缓存的权限
func (e *Enforcer) getCachedPermissions(cacheKey string) []string {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	entry, exists := e.permissionCache[cacheKey]
	if !exists {
		return nil
	}

	// 检查缓存是否过期
	if time.Since(entry.timestamp) > e.cacheTimeout {
		delete(e.permissionCache, cacheKey)
		return nil
	}

	return entry.permissions
}

// setCachedPermissions 设置权限缓存
func (e *Enforcer) setCachedPermissions(cacheKey string, permissions []string) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.permissionCache[cacheKey] = &cacheEntry{
		permissions: append([]string(nil), permissions...),
		timestamp:   time.Now(),
	}
}

// refreshUserPermissions 刷新用户权限缓存
func (e *Enforcer) refreshUserPermissions(user User) []string {
	permissions := e.configManager.GetUserPermissions(user)
	cacheKey := e.generateCacheKey(user)
	e.setCachedPermissions(cacheKey, permissions)
	return permissions
}

// containsPermission 检查权限列表是否包含指定权限
func (e *Enforcer) containsPermission(permissions []string, permission string) bool {
	for _, p := range permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// ClearCache 清除权限缓存
func (e *Enforcer) ClearCache(user *User) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	if user != nil {
		cacheKey := e.generateCacheKey(*user)
		delete(e.permissionCache, cacheKey)
	} else {
		e.permissionCache = make(map[string]*cacheEntry)
	}
}

// HandlePermissionChange 处理权限变更事件
func (e *Enforcer) HandlePermissionChange(event PermissionChangeEvent) {
	// 根据变更类型清除相关缓存
	switch event.TargetType {
	case "role":
		// 清除特定角色的缓存
		e.mutex.Lock()
		for key := range e.permissionCache {
			if len(key) > len(event.Target) && key[:len(event.Target)] == event.Target {
				delete(e.permissionCache, key)
			}
		}
		e.mutex.Unlock()

	case "courier-level":
		// 清除特定信使等级的缓存
		e.mutex.Lock()
		for key := range e.permissionCache {
			if len(key) > len(event.Target) && key[len(key)-len(event.Target):] == event.Target {
				delete(e.permissionCache, key)
			}
		}
		e.mutex.Unlock()

	case "system":
		// 系统级变更，清除所有缓存
		e.ClearCache(nil)
	}

	// 通知所有监听器
	for _, listener := range e.listeners {
		go listener(event)
	}
}

// AddPermissionChangeListener 添加权限变更监听器
func (e *Enforcer) AddPermissionChangeListener(listener PermissionChangeListener) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	e.listeners = append(e.listeners, listener)
}

// RemovePermissionChangeListener 移除权限变更监听器
func (e *Enforcer) RemovePermissionChangeListener(targetListener PermissionChangeListener) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	// 注意：在Go中比较函数比较复杂，这里提供一个简化的实现
	// 实际应用中可能需要使用其他方式来标识和移除特定的监听器
	newListeners := make([]PermissionChangeListener, 0, len(e.listeners))
	for _, listener := range e.listeners {
		// 这里的比较可能不会按预期工作，需要根据实际需求调整
		if fmt.Sprintf("%p", listener) != fmt.Sprintf("%p", targetListener) {
			newListeners = append(newListeners, listener)
		}
	}
	e.listeners = newListeners
}

// AnalyzeUserPermissions 分析用户权限覆盖率
func (e *Enforcer) AnalyzeUserPermissions(user User) *PermissionAnalysis {
	userPermissions := e.getUserPermissions(user, false)
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
// 选项模式
// ================================

// CheckOption 权限检查选项
type CheckOption func(*checkOptions)

type checkOptions struct {
	forceRefresh bool
}

// WithForceRefresh 强制刷新权限缓存
func WithForceRefresh(force bool) CheckOption {
	return func(opts *checkOptions) {
		opts.forceRefresh = force
	}
}

// ================================
// 权限装饰器
// ================================

// RequirePermission 权限装饰器函数类型
type RequirePermission func(permission string) func(next func(ctx context.Context, user User) error) func(ctx context.Context, user User) error

// NewRequirePermission 创建权限装饰器
func (e *Enforcer) NewRequirePermission() RequirePermission {
	return func(permission string) func(next func(ctx context.Context, user User) error) func(ctx context.Context, user User) error {
		return func(next func(ctx context.Context, user User) error) func(ctx context.Context, user User) error {
			return func(ctx context.Context, user User) error {
				if err := e.EnforcePermission(ctx, user, permission); err != nil {
					return err
				}
				return next(ctx, user)
			}
		}
	}
}