package middleware

import (
	"log"
	"net/http"
	"strings"

	"openpenpal-backend/internal/config"
	"openpenpal-backend/internal/models"
	"openpenpal-backend/pkg/auth"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// WebSocketAuthMiddleware WebSocket专用认证中间件
// 支持从查询参数或Authorization header获取token
func WebSocketAuthMiddleware(config *config.Config, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 打印调试信息
		log.Printf("WebSocket Auth Middleware - Path: %s, Query: %s", c.Request.URL.Path, c.Request.URL.RawQuery)
		
		var token string

		// 1. 优先从查询参数获取token（WebSocket客户端常用方式）
		token = c.Query("token")
		if token != "" {
			log.Printf("WebSocket Auth - Token from query: %s...", token[:min(20, len(token))])
		}

		// 2. 如果查询参数没有，尝试从Authorization header获取
		if token == "" {
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && parts[0] == "Bearer" {
					token = parts[1]
					log.Printf("WebSocket Auth - Token from header: %s...", token[:min(20, len(token))])
				}
			}
		}

		// 3. 验证token
		if token == "" {
			log.Printf("WebSocket Auth Failed - No token provided")
			// 对于WebSocket请求，不要返回JSON响应
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 使用ValidateJWT验证token
		claims, err := auth.ValidateJWT(token, config.JWTSecret)
		if err != nil {
			log.Printf("WebSocket Auth Failed - JWT validation error: %v", err)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 4. 从数据库加载完整用户信息
		var user models.User
		if err := db.Where("id = ?", claims.UserID).First(&user).Error; err != nil {
			log.Printf("WebSocket Auth Failed - User not found: %s", claims.UserID)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 5. 检查用户是否激活
		if !user.IsActive {
			log.Printf("WebSocket Auth Failed - User deactivated: %s", user.Username)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 6. 将用户信息添加到上下文
		c.Set("user_id", claims.UserID)
		c.Set("user_role", claims.Role)
		c.Set("user", &user) // 添加完整用户对象
		c.Set("school_code", user.SchoolCode) // 从用户对象获取school_code
		
		// 7. 应用角色兼容层（SOTA解决方案：统一认证处理）
		// 确保WebSocket连接也享受角色映射的好处
		frontendRole := models.GetFrontendRole(user.Role)
		c.Set("frontend_role", frontendRole)
		
		// 如果是信使角色，设置信使级别信息
		if courierInfo := models.GetCourierLevelInfo(user.Role); courierInfo != nil {
			c.Set("courier_level", courierInfo.Level)
			c.Set("courier_info", courierInfo)
		}
		
		log.Printf("WebSocket Auth Success - User: %s, Role: %s", user.Username, user.Role)
		c.Next()
	}
}

// min 返回两个数中较小的一个
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}