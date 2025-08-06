package middleware

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"openpenpal-backend/internal/config"
	"openpenpal-backend/internal/models"
	"openpenpal-backend/pkg/auth"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// UserCache 用户信息缓存
type UserCache struct {
	users map[string]*CachedUser
	mu    sync.RWMutex
	ttl   time.Duration
}

type CachedUser struct {
	User      *models.User
	ExpiresAt time.Time
}

var (
	userCache *UserCache
	once      sync.Once
)

func init() {
	once.Do(func() {
		userCache = &UserCache{
			users: make(map[string]*CachedUser),
			ttl:   5 * time.Minute, // 5分钟缓存TTL
		}
		
		// 启动清理协程
		go userCache.cleanupExpired()
	})
}

func (uc *UserCache) Get(userID string) (*models.User, bool) {
	uc.mu.RLock()
	cached, exists := uc.users[userID]
	uc.mu.RUnlock()
	
	if !exists || time.Now().After(cached.ExpiresAt) {
		return nil, false
	}
	
	return cached.User, true
}

func (uc *UserCache) Set(userID string, user *models.User) {
	uc.mu.Lock()
	uc.users[userID] = &CachedUser{
		User:      user,
		ExpiresAt: time.Now().Add(uc.ttl),
	}
	uc.mu.Unlock()
}

func (uc *UserCache) Delete(userID string) {
	uc.mu.Lock()
	delete(uc.users, userID)
	uc.mu.Unlock()
}

func (uc *UserCache) cleanupExpired() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		uc.mu.Lock()
		now := time.Now()
		for userID, cached := range uc.users {
			if now.After(cached.ExpiresAt) {
				delete(uc.users, userID)
			}
		}
		uc.mu.Unlock()
	}
}

// OptimizedAuthMiddleware 优化的JWT认证中间件
func OptimizedAuthMiddleware(config *config.Config, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "缺少认证令牌",
				"message": "请在请求头中提供Authorization令牌",
			})
			c.Abort()
			return
		}

		token, err := auth.ExtractTokenFromHeader(authHeader)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "令牌格式无效",
				"message": "Authorization头格式必须为 'Bearer <token>'",
			})
			c.Abort()
			return
		}

		claims, err := auth.ValidateJWT(token, config.JWTSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "令牌无效或已过期",
				"message": "请重新登录获取有效令牌",
			})
			c.Abort()
			return
		}

		// 首先尝试从缓存获取用户信息
		user, cached := userCache.Get(claims.UserID)
		if !cached {
			// 缓存未命中，从数据库查询
			var dbUser models.User
			
			// 使用上下文超时控制数据库查询
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			
			if err := db.WithContext(ctx).Where("id = ?", claims.UserID).First(&dbUser).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					c.JSON(http.StatusUnauthorized, gin.H{
						"success": false,
						"error":   "用户不存在",
						"message": "该用户账号可能已被删除",
					})
				} else {
					c.JSON(http.StatusInternalServerError, gin.H{
						"success": false,
						"error":   "数据库查询失败",
						"message": "请稍后重试",
					})
				}
				c.Abort()
				return
			}
			
			// 更新缓存
			user = &dbUser
			userCache.Set(claims.UserID, user)
		}

		// 检查用户是否激活
		if !user.IsActive {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "账号已被禁用",
				"message": "您的账号已被管理员禁用，如有疑问请联系客服",
			})
			c.Abort()
			return
		}

		// 将用户信息添加到上下文
		c.Set("user_id", claims.UserID)
		c.Set("user_role", claims.Role)
		c.Set("user", user)
		
		// 添加性能指标
		c.Header("X-Cache-Hit", fmt.Sprintf("%t", cached))
		
		c.Next()
	}
}

// CacheMiddleware 通用缓存中间件
func CacheMiddleware() gin.HandlerFunc {
	type CacheEntry struct {
		Data      interface{}
		ExpiresAt time.Time
	}
	
	cache := make(map[string]*CacheEntry)
	var cacheMu sync.RWMutex
	
	return func(c *gin.Context) {
		// 只对GET请求进行缓存
		if c.Request.Method != "GET" {
			c.Next()
			return
		}
		
		cacheKey := c.Request.URL.Path + "?" + c.Request.URL.RawQuery
		
		cacheMu.RLock()
		entry, exists := cache[cacheKey]
		cacheMu.RUnlock()
		
		if exists && time.Now().Before(entry.ExpiresAt) {
			c.Header("X-Cache", "HIT")
			c.JSON(http.StatusOK, entry.Data)
			c.Abort()
			return
		}
		
		c.Next()
		
		// 缓存响应（仅对成功响应）
		if c.Writer.Status() == http.StatusOK {
			// 这里需要根据实际需要实现响应缓存逻辑
			c.Header("X-Cache", "MISS")
		}
	}
}

// HealthCheckMiddleware 健康检查中间件
func HealthCheckMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/health" || c.Request.URL.Path == "/ping" {
			c.JSON(http.StatusOK, gin.H{
				"status":    "ok",
				"timestamp": time.Now().Unix(),
				"service":   "openpenpal-backend",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequestIDMiddleware 已在 request_id.go 中定义，此处移除重复定义

// 移除重复的函数定义，这些函数已在其他文件中定义

// InvalidateCacheForUser 清除特定用户的缓存
func InvalidateCacheForUser(userID string) {
	userCache.Delete(userID)
}

// GetCacheStats 获取缓存统计信息
func GetCacheStats() map[string]interface{} {
	userCache.mu.RLock()
	defer userCache.mu.RUnlock()
	
	return map[string]interface{}{
		"cached_users": len(userCache.users),
		"ttl_minutes":  userCache.ttl.Minutes(),
	}
}