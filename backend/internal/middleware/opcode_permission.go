package middleware

import (
	"net/http"
	"strings"

	"openpenpal-backend/internal/models"

	"github.com/gin-gonic/gin"
)

// OPCodePermission OP Code权限类型
type OPCodePermission int

const (
	OPCodePermissionView OPCodePermission = iota
	OPCodePermissionEdit
	OPCodePermissionCreate
	OPCodePermissionDelete
	OPCodePermissionBatch
)

// CourierInfo 信使信息（与前端保持一致）
type CourierInfo struct {
	ID                   string `json:"id"`
	Level                int    `json:"level"`                      // 1-4级
	ManagedOPCodePrefix  string `json:"managedOPCodePrefix"`        // 管理的OP Code前缀
	ZoneCode             string `json:"zoneCode"`                   // 区域代码（兼容字段）
}

// OPCodePermissions 权限对象
type OPCodePermissions struct {
	CanView   bool `json:"canView"`
	CanEdit   bool `json:"canEdit"`
	CanCreate bool `json:"canCreate"`
	CanDelete bool `json:"canDelete"`
	CanBatch  bool `json:"canBatch"`
}

// ValidateOPCodeAccess 验证信使是否有权限访问指定的OP Code
// 与前端的validateOPCodeAccess函数保持完全一致
func ValidateOPCodeAccess(courier CourierInfo, targetOPCode string) OPCodePermissions {
	defaultPermissions := OPCodePermissions{
		CanView:   false,
		CanEdit:   false,
		CanCreate: false,
		CanDelete: false,
		CanBatch:  false,
	}

	if courier.ID == "" || targetOPCode == "" {
		return defaultPermissions
	}

	level := courier.Level
	managedPrefix := courier.ManagedOPCodePrefix
	if managedPrefix == "" {
		managedPrefix = courier.ZoneCode
	}

	// 规范化OP Code
	opCode := strings.ToUpper(targetOPCode)

	switch level {
	case 4: // L4信使（城市总监）- 全局权限
		return OPCodePermissions{
			CanView:   true,
			CanEdit:   true,
			CanCreate: true,
			CanDelete: true,
			CanBatch:  true,
		}

	case 3: // L3信使（学校协调员）- 同校权限 (AA**)
		schoolPrefix := ""
		if len(managedPrefix) >= 2 {
			schoolPrefix = managedPrefix[:2]
		}
		if strings.HasPrefix(opCode, schoolPrefix) {
			return OPCodePermissions{
				CanView:   true,
				CanEdit:   true,
				CanCreate: true,
				CanDelete: true,
				CanBatch:  true,
			}
		}

	case 2: // L2信使（片区管理员）- 同区域权限 (AABB**)
		areaPrefix := ""
		if len(managedPrefix) >= 4 {
			areaPrefix = managedPrefix[:4]
		}
		if strings.HasPrefix(opCode, areaPrefix) {
			return OPCodePermissions{
				CanView:   true,
				CanEdit:   true,
				CanCreate: true,
				CanDelete: true,
				CanBatch:  true,
			}
		}

	case 1: // L1信使（楼栋投递员）- 精确投递权限
		buildingPrefix := ""
		if len(managedPrefix) >= 4 {
			buildingPrefix = managedPrefix[:4]
		}
		if strings.HasPrefix(opCode, buildingPrefix) {
			return OPCodePermissions{
				CanView:   true,
				CanEdit:   true,  // L1可以编辑单个宿舍邮编
				CanCreate: false, // L1不能创建新的投递点
				CanDelete: false, // L1不能删除投递点
				CanBatch:  false, // L1不能批量操作
			}
		}
	}

	return defaultPermissions
}

// ValidateAreaAccess 验证信使是否可以管理指定区域的数据
// 与前端的validateAreaAccess函数保持完全一致
func ValidateAreaAccess(courier CourierInfo, targetArea string, operationType string) bool {
	if courier.ID == "" || targetArea == "" {
		return false
	}

	level := courier.Level
	managedPrefix := courier.ManagedOPCodePrefix
	if managedPrefix == "" {
		managedPrefix = courier.ZoneCode
	}

	switch level {
	case 4: // L4信使 - 全局权限
		return true

	case 3: // L3信使 - 同校权限
		schoolCode := ""
		if len(managedPrefix) >= 2 {
			schoolCode = managedPrefix[:2]
		}
		return strings.HasPrefix(targetArea, schoolCode)

	case 2: // L2信使 - 同片区权限
		areaCode := ""
		if len(managedPrefix) >= 4 {
			areaCode = managedPrefix[:4]
		}
		return strings.HasPrefix(targetArea, areaCode)

	case 1: // L1信使 - 只能查看，不能编辑区域数据
		if operationType == "view" {
			buildingCode := ""
			if len(managedPrefix) >= 4 {
				buildingCode = managedPrefix[:4]
			}
			return strings.HasPrefix(targetArea, buildingCode)
		}
		return false

	default:
		return false
	}
}

// GetCourierInfoFromUser 从用户模型提取信使信息
func GetCourierInfoFromUser(user *models.User) CourierInfo {
	courier := CourierInfo{
		ID:                  user.ID,
		ManagedOPCodePrefix: user.OPCode, // 使用用户的OP Code作为管理前缀
		ZoneCode:            user.SchoolCode,
	}

	// 从角色中提取信使级别
	switch user.Role {
	case models.RoleCourierLevel1:
		courier.Level = 1
	case models.RoleCourierLevel2:
		courier.Level = 2
	case models.RoleCourierLevel3:
		courier.Level = 3
	case models.RoleCourierLevel4:
		courier.Level = 4
	default:
		courier.Level = 0 // 非信使角色
	}

	return courier
}

// OPCodePermissionMiddleware OP Code权限检查中间件
func OPCodePermissionMiddleware(requiredPermission OPCodePermission) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户信息
		userInterface, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "用户未认证",
				"message": "请先登录",
			})
			c.Abort()
			return
		}

		user, ok := userInterface.(*models.User)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "用户信息格式错误",
			})
			c.Abort()
			return
		}

		// 提取信使信息
		courier := GetCourierInfoFromUser(user)

		// 如果不是信使角色，检查是否是管理员
		if courier.Level == 0 {
			// 管理员角色可以访问
			if user.Role == models.RolePlatformAdmin || user.Role == models.RoleSuperAdmin {
				c.Set("courier_info", courier)
				c.Next()
				return
			}

			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "权限不足",
				"message": "需要信使权限才能访问此功能",
			})
			c.Abort()
			return
		}

		// 从请求中获取目标OP Code（可以从路径参数、查询参数或请求体中获取）
		targetOPCode := extractOPCodeFromRequest(c)

		// 验证OP Code权限
		permissions := ValidateOPCodeAccess(courier, targetOPCode)

		// 检查所需权限
		hasPermission := false
		switch requiredPermission {
		case OPCodePermissionView:
			hasPermission = permissions.CanView
		case OPCodePermissionEdit:
			hasPermission = permissions.CanEdit
		case OPCodePermissionCreate:
			hasPermission = permissions.CanCreate
		case OPCodePermissionDelete:
			hasPermission = permissions.CanDelete
		case OPCodePermissionBatch:
			hasPermission = permissions.CanBatch
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "权限不足",
				"message": "您没有执行此操作的权限",
				"details": gin.H{
					"required_permission": getPermissionName(requiredPermission),
					"courier_level":       courier.Level,
					"target_opcode":       targetOPCode,
					"managed_prefix":      courier.ManagedOPCodePrefix,
				},
			})
			c.Abort()
			return
		}

		// 将信使信息和权限添加到上下文
		c.Set("courier_info", courier)
		c.Set("opcode_permissions", permissions)
		c.Next()
	}
}

// extractOPCodeFromRequest 从请求中提取OP Code
func extractOPCodeFromRequest(c *gin.Context) string {
	// 尝试从路径参数中获取
	if opCode := c.Param("opcode"); opCode != "" {
		return opCode
	}
	if opCode := c.Param("op_code"); opCode != "" {
		return opCode
	}

	// 尝试从查询参数中获取
	if opCode := c.Query("opcode"); opCode != "" {
		return opCode
	}
	if opCode := c.Query("op_code"); opCode != "" {
		return opCode
	}

	// 尝试从请求体中获取（如果是JSON请求）
	if c.ContentType() == "application/json" {
		var body map[string]interface{}
		if err := c.ShouldBindJSON(&body); err == nil {
			if opCode, exists := body["op_code"]; exists {
				if opCodeStr, ok := opCode.(string); ok {
					return opCodeStr
				}
			}
			if opCode, exists := body["opcode"]; exists {
				if opCodeStr, ok := opCode.(string); ok {
					return opCodeStr
				}
			}
		}
	}

	// 如果找不到OP Code，返回空字符串
	return ""
}

// getPermissionName 获取权限名称
func getPermissionName(permission OPCodePermission) string {
	names := map[OPCodePermission]string{
		OPCodePermissionView:   "查看",
		OPCodePermissionEdit:   "编辑",
		OPCodePermissionCreate: "创建",
		OPCodePermissionDelete: "删除",
		OPCodePermissionBatch:  "批量操作",
	}
	if name, exists := names[permission]; exists {
		return name
	}
	return "未知权限"
}

// CourierLevelMiddleware 信使级别检查中间件
func CourierLevelMiddleware(minLevel int) gin.HandlerFunc {
	return func(c *gin.Context) {
		userInterface, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "用户未认证",
			})
			c.Abort()
			return
		}

		user, ok := userInterface.(*models.User)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "用户信息格式错误",
			})
			c.Abort()
			return
		}

		courier := GetCourierInfoFromUser(user)

		// 管理员可以访问所有级别
		if user.Role == models.RolePlatformAdmin || user.Role == models.RoleSuperAdmin {
			c.Set("courier_info", courier)
			c.Next()
			return
		}

		// 检查信使级别
		if courier.Level < minLevel {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "权限不足",
				"message": "需要更高级别的信使权限",
				"details": gin.H{
					"required_level": minLevel,
					"current_level":  courier.Level,
				},
			})
			c.Abort()
			return
		}

		c.Set("courier_info", courier)
		c.Next()
	}
}

// GetCourierInfo 从上下文中获取信使信息
func GetCourierInfo(c *gin.Context) (CourierInfo, bool) {
	courierInterface, exists := c.Get("courier_info")
	if !exists {
		return CourierInfo{}, false
	}
	courier, ok := courierInterface.(CourierInfo)
	return courier, ok
}

// GetOPCodePermissions 从上下文中获取OP Code权限
func GetOPCodePermissions(c *gin.Context) (OPCodePermissions, bool) {
	permissionsInterface, exists := c.Get("opcode_permissions")
	if !exists {
		return OPCodePermissions{}, false
	}
	permissions, ok := permissionsInterface.(OPCodePermissions)
	return permissions, ok
}