package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"openpenpal-backend/internal/models"
)

// PermissionMiddleware 权限检查中间件
func PermissionMiddleware(requiredPermission models.Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户信息（应该在AuthMiddleware中设置）
		userInterface, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
			c.Abort()
			return
		}

		user, ok := userInterface.(*models.User)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "用户信息格式错误"})
			c.Abort()
			return
		}

		// 检查权限
		if !user.HasPermission(requiredPermission) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":               "权限不足",
				"required_permission": string(requiredPermission),
				"user_role":           string(user.Role),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RoleMiddleware 角色检查中间件（原有的，现在支持更多角色）
func RoleMiddleware(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userInterface, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
			c.Abort()
			return
		}

		user, ok := userInterface.(*models.User)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "用户信息格式错误"})
			c.Abort()
			return
		}

		// 转换为新的角色类型
		requiredUserRole := models.UserRole(requiredRole)

		// 检查角色权限
		if !user.HasRole(requiredUserRole) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":         "角色权限不足",
				"required_role": requiredRole,
				"user_role":     string(user.Role),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// SchoolAdminMiddleware 学校管理员权限检查
func SchoolAdminMiddleware() gin.HandlerFunc {
	return RoleMiddleware("school_admin")
}

// PlatformAdminMiddleware 平台管理员权限检查
func PlatformAdminMiddleware() gin.HandlerFunc {
	return RoleMiddleware("platform_admin")
}

// SuperAdminMiddleware 超级管理员权限检查
func SuperAdminMiddleware() gin.HandlerFunc {
	return RoleMiddleware("super_admin")
}

// CourierMiddleware 信使权限检查
func CourierMiddleware() gin.HandlerFunc {
	return RoleMiddleware("courier")
}

// SameSchoolMiddleware 同校权限检查
func SameSchoolMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userInterface, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
			c.Abort()
			return
		}

		user, ok := userInterface.(*models.User)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "用户信息格式错误"})
			c.Abort()
			return
		}

		// 超级管理员和平台管理员可以跨校管理
		if user.HasRole(models.RolePlatformAdmin) {
			c.Next()
			return
		}

		// 学校管理员只能管理同校用户
		// 这里需要根据具体的业务逻辑来实现
		// 例如，从请求参数中获取目标用户的学校信息

		c.Next()
	}
}

// CheckOPCodePermission 检查OP Code权限的辅助函数
func CheckOPCodePermission(user *models.User, targetOPCode string, requiredPermission string) bool {
	// 提取信使信息
	courier := GetCourierInfoFromUser(user)
	
	// 管理员拥有所有权限
	if user.Role == models.RolePlatformAdmin || user.Role == models.RoleSuperAdmin {
		return true
	}
	
	// 非信使角色无OP Code权限
	if courier.Level == 0 {
		return false
	}
	
	// 验证OP Code权限
	permissions := ValidateOPCodeAccess(courier, targetOPCode)
	
	switch requiredPermission {
	case "view":
		return permissions.CanView
	case "edit":
		return permissions.CanEdit
	case "create":
		return permissions.CanCreate
	case "delete":
		return permissions.CanDelete
	case "batch":
		return permissions.CanBatch
	default:
		return false
	}
}
