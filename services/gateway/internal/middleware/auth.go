package middleware

import (
	"github.com/openpenpal/shared/go/pkg/middleware"
	"github.com/openpenpal/shared/go/pkg/permissions"

	"github.com/gin-gonic/gin"
)

// JWTAuth JWT认证中间件 - 使用统一认证中间件
func JWTAuth(jwtSecret string) gin.HandlerFunc {
	// 创建JWT配置
	config := &middleware.JWTConfig{
		SigningKey: []byte(jwtSecret),
		SkipperFunc: func(c *gin.Context) bool {
			skipPaths := []string{
				"/health",
				"/metrics",
				"/api/v1/auth/register",
				"/api/v1/auth/login",
				"/api/v1/auth/refresh",
				"/ping",
			}

			for _, path := range skipPaths {
				if c.Request.URL.Path == path {
					return true
				}
			}
			return false
		},
	}

	// 使用统一认证中间件
	return middleware.JWTMiddleware(config)
}

// AdminAuth 管理员权限中间件 - 使用统一权限检查
func AdminAuth() gin.HandlerFunc {
	return middleware.RequireRole(permissions.UserRoleAdmin)
}

// CourierAuth 信使权限中间件 - 使用统一权限检查
func CourierAuth() gin.HandlerFunc {
	return middleware.RequirePermission(permissions.PermissionCourierScanCode)
}

// GetUserID 从上下文获取用户ID - 使用JWT Claims
func GetUserID(c *gin.Context) string {
	claims, exists := c.Get("jwt_claims")
	if !exists {
		return ""
	}

	jwtClaims, ok := claims.(*middleware.JWTClaims)
	if !ok {
		return ""
	}

	return jwtClaims.UserID
}

// GetUserRole 从上下文获取用户角色 - 使用JWT Claims
func GetUserRole(c *gin.Context) string {
	claims, exists := c.Get("jwt_claims")
	if !exists {
		return ""
	}

	jwtClaims, ok := claims.(*middleware.JWTClaims)
	if !ok {
		return ""
	}

	return string(jwtClaims.Role)
}

// GetUser 从上下文获取完整用户信息 - 基于JWT Claims构建
func GetUser(c *gin.Context) *permissions.User {
	claims, exists := c.Get("jwt_claims")
	if !exists {
		return nil
	}

	jwtClaims, ok := claims.(*middleware.JWTClaims)
	if !ok {
		return nil
	}

	return &permissions.User{
		Role:        jwtClaims.Role,
		CourierInfo: jwtClaims.CourierInfo,
	}
}
