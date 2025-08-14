package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"openpenpal-backend/internal/config"
	"openpenpal-backend/internal/models"
	"openpenpal-backend/pkg/auth"
	"openpenpal-backend/pkg/cache"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AuthMiddleware JWT认证中间件 - 优化版本，支持缓存和黑名单
func AuthMiddleware(config *config.Config, db *gorm.DB) gin.HandlerFunc {
	userCache := cache.GetUserCache()
	blacklist := cache.GetTokenBlacklist()

	return func(c *gin.Context) {
		// 性能监控
		start := time.Now()

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

		// 检查Token是否在黑名单中
		if claims.RegisteredClaims.ID != "" && blacklist.IsBlacklisted(claims.RegisteredClaims.ID) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "令牌已被注销",
				"message": "该令牌已被主动注销，请重新登录",
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
			// 清除该用户的缓存
			userCache.Delete(claims.UserID)

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
		c.Set("token_jti", claims.RegisteredClaims.ID) // 保存JWT ID用于注销

		// 添加性能监控头
		duration := time.Since(start)
		c.Header("X-Auth-Time", fmt.Sprintf("%v", duration))
		c.Header("X-Cache-Hit", fmt.Sprintf("%t", cached))

		c.Next()
	}
}

// OptionalAuthMiddleware 可选认证中间件
func OptionalAuthMiddleware(config *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			token, err := auth.ExtractTokenFromHeader(authHeader)
			if err == nil {
				claims, err := auth.ValidateJWT(token, config.JWTSecret)
				if err == nil {
					c.Set("user_id", claims.UserID)
					c.Set("user_role", claims.Role)
				}
			}
		}
		c.Next()
	}
}

// CORSMiddleware CORS中间件
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// 允许的域名列表
		allowedOrigins := []string{
			"http://localhost:3000",
			"http://localhost:3001",
			"https://openpenpal.example.com",
		}

		allowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				allowed = true
				break
			}
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, x-client-version, x-csrf-token")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// LoggerMiddleware 日志中间件
func LoggerMiddleware() gin.HandlerFunc {
	return gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{"/health", "/ping"},
	})
}

// RecoveryMiddleware 错误恢复中间件
func RecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
		})
	})
}

// GetUserID 从上下文中获取用户ID
func GetUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", false
	}
	return userID.(string), true
}

// GetUserRole 从上下文中获取用户角色
func GetUserRole(c *gin.Context) (string, bool) {
	userRole, exists := c.Get("user_role")
	if !exists {
		return "", false
	}
	return userRole.(string), true
}

// RequireAuth 检查是否已认证
func RequireAuth(c *gin.Context) bool {
	_, exists := c.Get("user_id")
	return exists
}
