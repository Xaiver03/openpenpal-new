package middleware

import (
	"courier-service/internal/models"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CourierAuthMiddleware 信使权限验证中间件
type CourierAuthMiddleware struct {
	db *gorm.DB
}

// NewCourierAuthMiddleware 创建信使权限中间件
func NewCourierAuthMiddleware(db *gorm.DB) *CourierAuthMiddleware {
	return &CourierAuthMiddleware{db: db}
}

// RequirePermission 需要特定权限的中间件
func (cam *CourierAuthMiddleware) RequirePermission(permission models.CourierPermission) gin.HandlerFunc {
	return func(c *gin.Context) {
		courierID := GetUserID(c)
		if courierID == "" {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse(
				models.CodeUnauthorized,
				"Courier ID not found",
				nil,
			))
			c.Abort()
			return
		}

		// 获取信使信息
		var courier models.Courier
		if err := cam.db.Where("user_id = ?", courierID).First(&courier).Error; err != nil {
			c.JSON(http.StatusForbidden, models.ErrorResponse(
				models.CodeUnauthorized,
				"Courier not found or not approved",
				err.Error(),
			))
			c.Abort()
			return
		}

		// 检查信使状态
		if courier.Status != "approved" {
			c.JSON(http.StatusForbidden, models.ErrorResponse(
				models.CodeUnauthorized,
				"Courier not approved",
				nil,
			))
			c.Abort()
			return
		}

		// 检查权限
		hasPermission, reason := cam.checkPermission(courierID, permission, c)
		if !hasPermission {
			c.JSON(http.StatusForbidden, models.ErrorResponse(
				models.CodeUnauthorized,
				"Insufficient permission",
				reason,
			))
			c.Abort()
			return
		}

		// 将信使信息添加到上下文
		c.Set("courier_info", courier)
		c.Next()
	}
}

// RequireLevel 需要特定等级的中间件
func (cam *CourierAuthMiddleware) RequireLevel(minLevel models.CourierLevel) gin.HandlerFunc {
	return func(c *gin.Context) {
		courierID := GetUserID(c)
		if courierID == "" {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse(
				models.CodeUnauthorized,
				"Courier ID not found",
				nil,
			))
			c.Abort()
			return
		}

		// 获取信使信息
		var courier models.Courier
		if err := cam.db.Where("user_id = ?", courierID).First(&courier).Error; err != nil {
			c.JSON(http.StatusForbidden, models.ErrorResponse(
				models.CodeUnauthorized,
				"Courier not found",
				err.Error(),
			))
			c.Abort()
			return
		}

		// 检查等级 (假设在courier表中有level字段，需要数据库结构更新)
		courierLevel := models.CourierLevel(courier.Level) // 需要在Courier模型中添加Level字段
		if courierLevel < minLevel {
			c.JSON(http.StatusForbidden, models.ErrorResponse(
				models.CodeUnauthorized,
				"Insufficient courier level",
				map[string]interface{}{
					"required_level": minLevel,
					"current_level":  courierLevel,
				},
			))
			c.Abort()
			return
		}

		c.Set("courier_info", courier)
		c.Set("courier_level", courierLevel)
		c.Next()
	}
}

// RequireZoneAccess 需要特定区域访问权限的中间件
func (cam *CourierAuthMiddleware) RequireZoneAccess(zoneType models.CourierZoneType, zoneID string) gin.HandlerFunc {
	return func(c *gin.Context) {
		courierID := GetUserID(c)
		if courierID == "" {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse(
				models.CodeUnauthorized,
				"Courier ID not found",
				nil,
			))
			c.Abort()
			return
		}

		// 检查区域访问权限
		hasAccess, reason := cam.checkZoneAccess(courierID, zoneType, zoneID)
		if !hasAccess {
			c.JSON(http.StatusForbidden, models.ErrorResponse(
				models.CodeUnauthorized,
				"No access to the specified zone",
				reason,
			))
			c.Abort()
			return
		}

		c.Next()
	}
}

// checkPermission 检查信使是否有指定权限
func (cam *CourierAuthMiddleware) checkPermission(courierID string, permission models.CourierPermission, c *gin.Context) (bool, string) {
	// 获取信使等级
	var courier models.Courier
	if err := cam.db.Where("user_id = ?", courierID).First(&courier).Error; err != nil {
		return false, "Courier not found"
	}

	courierLevel := models.CourierLevel(courier.Level)

	// 检查权限矩阵
	allowedPermissions, exists := models.DefaultPermissionMatrix[courierLevel]
	if !exists {
		return false, "Invalid courier level"
	}

	// 检查是否有该权限
	for _, allowedPermission := range allowedPermissions {
		if allowedPermission == permission {
			// 还需要检查区域权限
			return cam.checkContextualPermission(courierID, courierLevel, permission, c)
		}
	}

	return false, "Permission not granted for this level"
}

// checkContextualPermission 检查上下文相关权限
func (cam *CourierAuthMiddleware) checkContextualPermission(courierID string, level models.CourierLevel, permission models.CourierPermission, c *gin.Context) (bool, string) {
	// 从请求路径或参数中获取区域信息
	zoneInfo := cam.extractZoneInfoFromRequest(c)

	if zoneInfo == nil {
		// 如果没有区域信息，按等级给予基础权限
		return true, ""
	}

	// 检查信使是否有权限访问该区域
	hasAccess, reason := cam.checkZoneAccess(courierID, zoneInfo.ZoneType, zoneInfo.ZoneID)
	if !hasAccess {
		return false, reason
	}

	// 检查权限是否适用于该区域类型
	expectedZoneType := models.DefaultZoneMapping[level]
	if !cam.isZoneTypeCompatible(expectedZoneType, zoneInfo.ZoneType) {
		return false, "Zone type not compatible with courier level"
	}

	return true, ""
}

// checkZoneAccess 检查区域访问权限
func (cam *CourierAuthMiddleware) checkZoneAccess(courierID string, zoneType models.CourierZoneType, zoneID string) (bool, string) {
	// 查询信使的管理区域
	var zones []models.CourierZone
	if err := cam.db.Where("courier_id = ? AND is_active = ?", courierID, true).Find(&zones).Error; err != nil {
		return false, "Failed to query courier zones"
	}

	// 检查直接区域权限
	for _, zone := range zones {
		if zone.ZoneType == zoneType && zone.ZoneID == zoneID {
			return true, ""
		}
	}

	// 检查层级权限 (上级区域包含下级区域)
	for _, zone := range zones {
		if cam.isParentZone(zone.ZoneType, zone.ZoneID, zoneType, zoneID) {
			return true, ""
		}
	}

	return false, "No access to the specified zone"
}

// extractZoneInfoFromRequest 从请求中提取区域信息
func (cam *CourierAuthMiddleware) extractZoneInfoFromRequest(c *gin.Context) *struct {
	ZoneType models.CourierZoneType
	ZoneID   string
} {
	// 从路径参数中提取
	if zoneType := c.Param("zone_type"); zoneType != "" {
		if zoneID := c.Param("zone_id"); zoneID != "" {
			return &struct {
				ZoneType models.CourierZoneType
				ZoneID   string
			}{
				ZoneType: models.CourierZoneType(zoneType),
				ZoneID:   zoneID,
			}
		}
	}

	// 从查询参数中提取
	if zoneType := c.Query("zone_type"); zoneType != "" {
		if zoneID := c.Query("zone_id"); zoneID != "" {
			return &struct {
				ZoneType models.CourierZoneType
				ZoneID   string
			}{
				ZoneType: models.CourierZoneType(zoneType),
				ZoneID:   zoneID,
			}
		}
	}

	// 从请求体中提取 (对于POST/PUT请求)
	var requestBody map[string]interface{}
	if c.Request.Method == "POST" || c.Request.Method == "PUT" {
		if err := c.ShouldBindJSON(&requestBody); err == nil {
			if zoneType, ok := requestBody["zone_type"].(string); ok {
				if zoneID, ok := requestBody["zone_id"].(string); ok {
					return &struct {
						ZoneType models.CourierZoneType
						ZoneID   string
					}{
						ZoneType: models.CourierZoneType(zoneType),
						ZoneID:   zoneID,
					}
				}
			}
		}
	}

	return nil
}

// isZoneTypeCompatible 检查区域类型是否兼容
func (cam *CourierAuthMiddleware) isZoneTypeCompatible(courierZoneType, requestZoneType models.CourierZoneType) bool {
	// 等级映射关系
	zoneHierarchy := map[models.CourierZoneType]int{
		models.ZoneBuilding: 1,
		models.ZoneArea:     2,
		models.ZoneCampus:   3,
		models.ZoneCity:     4,
	}

	courierLevel := zoneHierarchy[courierZoneType]
	requestLevel := zoneHierarchy[requestZoneType]

	// 高级信使可以管理低级区域，但不能反过来
	return courierLevel >= requestLevel
}

// isParentZone 检查是否为父级区域
func (cam *CourierAuthMiddleware) isParentZone(parentZoneType models.CourierZoneType, parentZoneID string, childZoneType models.CourierZoneType, childZoneID string) bool {
	// 这里需要实现区域层级关系的具体逻辑
	// 例如：城市包含校区，校区包含片区，片区包含楼栋

	zoneHierarchy := map[models.CourierZoneType]int{
		models.ZoneBuilding: 1,
		models.ZoneArea:     2,
		models.ZoneCampus:   3,
		models.ZoneCity:     4,
	}

	parentLevel := zoneHierarchy[parentZoneType]
	childLevel := zoneHierarchy[childZoneType]

	// 只有上级区域才能包含下级区域
	if parentLevel <= childLevel {
		return false
	}

	// 这里应该查询实际的区域层级关系数据库
	// 暂时使用简单的字符串包含关系作为示例
	return strings.HasPrefix(childZoneID, parentZoneID)
}

// GetCourierInfo 获取当前请求的信使信息
func GetCourierInfo(c *gin.Context) (*models.Courier, bool) {
	if courier, exists := c.Get("courier_info"); exists {
		if courierModel, ok := courier.(models.Courier); ok {
			return &courierModel, true
		}
	}
	return nil, false
}

// GetCourierLevel 获取当前请求的信使等级
func GetCourierLevel(c *gin.Context) (models.CourierLevel, bool) {
	if level, exists := c.Get("courier_level"); exists {
		if courierLevel, ok := level.(models.CourierLevel); ok {
			return courierLevel, true
		}
	}
	return 0, false
}
