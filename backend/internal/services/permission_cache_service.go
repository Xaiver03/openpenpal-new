package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"openpenpal-backend/internal/config"
	"openpenpal-backend/internal/logger"
)

// PermissionCacheService 权限缓存服务
type PermissionCacheService struct {
	redis  *redis.Client
	config *config.Config
	logger *logger.SmartLogger
	
	// 缓存过期时间设置（秒）
	UserPermissionTTL  int // 用户权限缓存：5分钟
	RolePermissionTTL  int // 角色权限缓存：10分钟
	MenuCacheTTL       int // 菜单缓存：30分钟
	SessionTTL         int // 会话缓存：24小时
}

// NewPermissionCacheService 创建权限缓存服务
func NewPermissionCacheService(redis *redis.Client, cfg *config.Config, logger *logger.SmartLogger) *PermissionCacheService {
	return &PermissionCacheService{
		redis:  redis,
		config: cfg,
		logger: logger,
		
		// 安全的缓存过期时间配置
		UserPermissionTTL:  300,  // 5分钟 - 权限变更快速生效
		RolePermissionTTL:  600,  // 10分钟 - 角色变更及时生效
		MenuCacheTTL:       1800, // 30分钟 - 菜单结构相对稳定
		SessionTTL:         86400, // 24小时 - 会话保持
	}
}

// CachedPermission 缓存的权限信息
type CachedPermission struct {
	UserID      string            `json:"user_id"`
	Permissions []string          `json:"permissions"`
	Roles       []string          `json:"roles"`
	Metadata    map[string]string `json:"metadata"`
	CachedAt    time.Time         `json:"cached_at"`
	ExpiresAt   time.Time         `json:"expires_at"`
}

// CachedMenu 缓存的菜单信息
type CachedMenu struct {
	MenuID    string           `json:"menu_id"`
	ParentID  string           `json:"parent_id"`
	Name      string           `json:"name"`
	Path      string           `json:"path"`
	Icon      string           `json:"icon"`
	Component string           `json:"component"`
	Children  []*CachedMenu    `json:"children,omitempty"`
	Meta      map[string]interface{} `json:"meta"`
}

// SetUserPermissions 设置用户权限缓存
func (s *PermissionCacheService) SetUserPermissions(ctx context.Context, userID string, permissions []string, roles []string) error {
	key := s.getUserPermissionKey(userID)
	
	cached := &CachedPermission{
		UserID:      userID,
		Permissions: permissions,
		Roles:       roles,
		Metadata:    make(map[string]string),
		CachedAt:    time.Now(),
		ExpiresAt:   time.Now().Add(time.Duration(s.UserPermissionTTL) * time.Second),
	}
	
	data, err := json.Marshal(cached)
	if err != nil {
		s.logger.Error("Failed to marshal user permissions: %v", err)
		return fmt.Errorf("failed to marshal permissions: %w", err)
	}
	
	err = s.redis.Set(ctx, key, data, time.Duration(s.UserPermissionTTL)*time.Second).Err()
	if err != nil {
		s.logger.Error("Failed to cache user permissions for %s: %v", userID, err)
		return fmt.Errorf("failed to cache permissions: %w", err)
	}
	
	s.logger.Debug("Cached permissions for user %s (TTL: %ds)", userID, s.UserPermissionTTL)
	return nil
}

// GetUserPermissions 获取用户权限缓存
func (s *PermissionCacheService) GetUserPermissions(ctx context.Context, userID string) (*CachedPermission, error) {
	key := s.getUserPermissionKey(userID)
	
	data, err := s.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		s.logger.Debug("No cached permissions found for user %s", userID)
		return nil, nil
	}
	if err != nil {
		s.logger.Error("Failed to get cached permissions for %s: %v", userID, err)
		return nil, fmt.Errorf("failed to get cached permissions: %w", err)
	}
	
	var cached CachedPermission
	if err := json.Unmarshal([]byte(data), &cached); err != nil {
		s.logger.Error("Failed to unmarshal cached permissions for %s: %v", userID, err)
		return nil, fmt.Errorf("failed to unmarshal permissions: %w", err)
	}
	
	// 检查是否过期
	if time.Now().After(cached.ExpiresAt) {
		s.logger.Debug("Cached permissions expired for user %s", userID)
		s.InvalidateUserPermissions(ctx, userID)
		return nil, nil
	}
	
	s.logger.Debug("Retrieved cached permissions for user %s", userID)
	return &cached, nil
}

// InvalidateUserPermissions 清除用户权限缓存
func (s *PermissionCacheService) InvalidateUserPermissions(ctx context.Context, userID string) error {
	patterns := []string{
		s.getUserPermissionKey(userID),
		s.getUserMenuKey(userID, "*"),
		s.getUserSessionKey(userID, "*"),
	}
	
	for _, pattern := range patterns {
		if err := s.deletePattern(ctx, pattern); err != nil {
			s.logger.Error("Failed to invalidate cache pattern %s: %v", pattern, err)
		}
	}
	
	s.logger.Info("Invalidated all cache for user %s", userID)
	return nil
}

// SetUserMenus 设置用户菜单缓存
func (s *PermissionCacheService) SetUserMenus(ctx context.Context, userID string, bizType string, menus []*CachedMenu) error {
	key := s.getUserMenuKey(userID, bizType)
	
	data, err := json.Marshal(menus)
	if err != nil {
		s.logger.Error("Failed to marshal user menus: %v", err)
		return fmt.Errorf("failed to marshal menus: %w", err)
	}
	
	err = s.redis.Set(ctx, key, data, time.Duration(s.MenuCacheTTL)*time.Second).Err()
	if err != nil {
		s.logger.Error("Failed to cache user menus for %s: %v", userID, err)
		return fmt.Errorf("failed to cache menus: %w", err)
	}
	
	s.logger.Debug("Cached menus for user %s bizType %s (TTL: %ds)", userID, bizType, s.MenuCacheTTL)
	return nil
}

// GetUserMenus 获取用户菜单缓存
func (s *PermissionCacheService) GetUserMenus(ctx context.Context, userID string, bizType string) ([]*CachedMenu, error) {
	key := s.getUserMenuKey(userID, bizType)
	
	data, err := s.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		s.logger.Debug("No cached menus found for user %s bizType %s", userID, bizType)
		return nil, nil
	}
	if err != nil {
		s.logger.Error("Failed to get cached menus for %s: %v", userID, err)
		return nil, fmt.Errorf("failed to get cached menus: %w", err)
	}
	
	var menus []*CachedMenu
	if err := json.Unmarshal([]byte(data), &menus); err != nil {
		s.logger.Error("Failed to unmarshal cached menus for %s: %v", userID, err)
		return nil, fmt.Errorf("failed to unmarshal menus: %w", err)
	}
	
	s.logger.Debug("Retrieved cached menus for user %s bizType %s", userID, bizType)
	return menus, nil
}

// InvalidateRolePermissions 清除角色相关的权限缓存
func (s *PermissionCacheService) InvalidateRolePermissions(ctx context.Context, roleID string) error {
	// 获取拥有该角色的所有用户
	userIDs, err := s.getUsersByRole(ctx, roleID)
	if err != nil {
		s.logger.Error("Failed to get users by role %s: %v", roleID, err)
		return err
	}
	
	// 清除每个用户的权限缓存
	for _, userID := range userIDs {
		if err := s.InvalidateUserPermissions(ctx, userID); err != nil {
			s.logger.Error("Failed to invalidate permissions for user %s: %v", userID, err)
		}
	}
	
	s.logger.Info("Invalidated permissions for role %s affecting %d users", roleID, len(userIDs))
	return nil
}

// InvalidateMenuPermissions 清除菜单相关的权限缓存
func (s *PermissionCacheService) InvalidateMenuPermissions(ctx context.Context, menuID string) error {
	// 清除所有菜单缓存
	pattern := "openpenpal:menu:*"
	if err := s.deletePattern(ctx, pattern); err != nil {
		s.logger.Error("Failed to invalidate menu cache: %v", err)
		return err
	}
	
	s.logger.Info("Invalidated all menu cache due to menu %s change", menuID)
	return nil
}

// SetSessionPermissions 设置会话权限缓存
func (s *PermissionCacheService) SetSessionPermissions(ctx context.Context, sessionID string, userID string, permissions map[string]interface{}) error {
	key := s.getSessionKey(sessionID)
	
	sessionData := map[string]interface{}{
		"user_id":     userID,
		"permissions": permissions,
		"created_at":  time.Now(),
		"expires_at":  time.Now().Add(time.Duration(s.SessionTTL) * time.Second),
	}
	
	data, err := json.Marshal(sessionData)
	if err != nil {
		s.logger.Error("Failed to marshal session data: %v", err)
		return fmt.Errorf("failed to marshal session: %w", err)
	}
	
	err = s.redis.Set(ctx, key, data, time.Duration(s.SessionTTL)*time.Second).Err()
	if err != nil {
		s.logger.Error("Failed to cache session %s: %v", sessionID, err)
		return fmt.Errorf("failed to cache session: %w", err)
	}
	
	s.logger.Debug("Cached session %s for user %s (TTL: %ds)", sessionID, userID, s.SessionTTL)
	return nil
}

// GetSessionPermissions 获取会话权限缓存
func (s *PermissionCacheService) GetSessionPermissions(ctx context.Context, sessionID string) (map[string]interface{}, error) {
	key := s.getSessionKey(sessionID)
	
	data, err := s.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		s.logger.Debug("No cached session found: %s", sessionID)
		return nil, nil
	}
	if err != nil {
		s.logger.Error("Failed to get cached session %s: %v", sessionID, err)
		return nil, fmt.Errorf("failed to get cached session: %w", err)
	}
	
	var sessionData map[string]interface{}
	if err := json.Unmarshal([]byte(data), &sessionData); err != nil {
		s.logger.Error("Failed to unmarshal session data %s: %v", sessionID, err)
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}
	
	s.logger.Debug("Retrieved cached session %s", sessionID)
	return sessionData, nil
}

// InvalidateSession 清除会话缓存
func (s *PermissionCacheService) InvalidateSession(ctx context.Context, sessionID string) error {
	key := s.getSessionKey(sessionID)
	
	err := s.redis.Del(ctx, key).Err()
	if err != nil {
		s.logger.Error("Failed to invalidate session %s: %v", sessionID, err)
		return fmt.Errorf("failed to invalidate session: %w", err)
	}
	
	s.logger.Debug("Invalidated session %s", sessionID)
	return nil
}

// GetCacheStats 获取缓存统计信息
func (s *PermissionCacheService) GetCacheStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// 统计各类缓存数量
	patterns := map[string]string{
		"user_permissions": "openpenpal:user:*:permissions",
		"user_menus":       "openpenpal:user:*:menu:*",
		"sessions":         "openpenpal:session:*",
		"roles":            "openpenpal:role:*",
	}
	
	for category, pattern := range patterns {
		keys, err := s.redis.Keys(ctx, pattern).Result()
		if err != nil {
			s.logger.Error("Failed to get keys for pattern %s: %v", pattern, err)
			continue
		}
		stats[category+"_count"] = len(keys)
	}
	
	// 添加配置信息
	stats["ttl_config"] = map[string]int{
		"user_permission_ttl": s.UserPermissionTTL,
		"role_permission_ttl": s.RolePermissionTTL,
		"menu_cache_ttl":      s.MenuCacheTTL,
		"session_ttl":         s.SessionTTL,
	}
	
	return stats, nil
}

// Helper methods for cache keys
func (s *PermissionCacheService) getUserPermissionKey(userID string) string {
	return fmt.Sprintf("openpenpal:user:%s:permissions", userID)
}

func (s *PermissionCacheService) getUserMenuKey(userID, bizType string) string {
	return fmt.Sprintf("openpenpal:user:%s:menu:%s", userID, bizType)
}

func (s *PermissionCacheService) getUserSessionKey(userID, sessionID string) string {
	return fmt.Sprintf("openpenpal:user:%s:session:%s", userID, sessionID)
}

func (s *PermissionCacheService) getSessionKey(sessionID string) string {
	return fmt.Sprintf("openpenpal:session:%s", sessionID)
}

func (s *PermissionCacheService) getRoleKey(roleID string) string {
	return fmt.Sprintf("openpenpal:role:%s", roleID)
}

// deletePattern 删除匹配模式的缓存键
func (s *PermissionCacheService) deletePattern(ctx context.Context, pattern string) error {
	keys, err := s.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}
	
	if len(keys) > 0 {
		return s.redis.Del(ctx, keys...).Err()
	}
	
	return nil
}

// getUsersByRole 获取拥有指定角色的用户ID列表（这里需要查询数据库）
func (s *PermissionCacheService) getUsersByRole(ctx context.Context, roleID string) ([]string, error) {
	// 这里应该查询数据库获取用户列表
	// 为了简化，返回空列表，实际实现时需要注入数据库服务
	s.logger.Warn("getUsersByRole not implemented, should query database")
	return []string{}, nil
}

// UpdateCacheConfig 更新缓存配置
func (s *PermissionCacheService) UpdateCacheConfig(config map[string]int) {
	if ttl, exists := config["user_permission_ttl"]; exists && ttl > 0 && ttl <= 3600 {
		s.UserPermissionTTL = ttl
		s.logger.Info("Updated UserPermissionTTL to %d seconds", ttl)
	}
	
	if ttl, exists := config["role_permission_ttl"]; exists && ttl > 0 && ttl <= 7200 {
		s.RolePermissionTTL = ttl
		s.logger.Info("Updated RolePermissionTTL to %d seconds", ttl)
	}
	
	if ttl, exists := config["menu_cache_ttl"]; exists && ttl > 0 && ttl <= 7200 {
		s.MenuCacheTTL = ttl
		s.logger.Info("Updated MenuCacheTTL to %d seconds", ttl)
	}
	
	if ttl, exists := config["session_ttl"]; exists && ttl > 0 && ttl <= 604800 {
		s.SessionTTL = ttl
		s.logger.Info("Updated SessionTTL to %d seconds", ttl)
	}
}