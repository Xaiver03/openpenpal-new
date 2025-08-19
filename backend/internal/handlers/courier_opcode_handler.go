package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"openpenpal-backend/internal/models"
	"openpenpal-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// CourierOPCodeHandler 信使OP Code处理器
type CourierOPCodeHandler struct {
	opcodeService  *services.OPCodeService
	courierService *services.CourierService
}

// NewCourierOPCodeHandler 创建信使OP Code处理器
func NewCourierOPCodeHandler(opcodeService *services.OPCodeService, courierService *services.CourierService) *CourierOPCodeHandler {
	return &CourierOPCodeHandler{
		opcodeService:  opcodeService,
		courierService: courierService,
	}
}

// GetApplications 获取待审核申请列表（根据信使级别过滤）
func (h *CourierOPCodeHandler) GetApplications(c *gin.Context) {
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"code":    4001,
			"message": "用户未认证",
		})
		return
	}
	user := userInterface.(*models.User)

	// 检查信使身份
	courierLevel := h.getCourierLevel(models.UserRole(user.Role))
	if courierLevel == 0 {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"code":    4003,
			"message": "需要信使权限",
		})
		return
	}

	// 获取信使信息
	courier, err := h.courierService.GetCourierByUserID(user.ID)
	if err != nil || courier == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"code":    4004,
			"message": "信使信息不存在",
		})
		return
	}

	// 获取所有申请
	applications, err := h.opcodeService.GetPendingApplications()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"code":    5001,
			"message": "获取申请列表失败",
			"error":   err.Error(),
		})
		return
	}

	// 根据信使级别和管理范围过滤申请
	filteredApps := h.filterApplicationsByLevel(applications, courier, courierLevel)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"code":    200,
		"message": "获取成功",
		"data":    filteredApps,
	})
}

// ReviewApplication 审核申请（信使专用）
func (h *CourierOPCodeHandler) ReviewApplication(c *gin.Context) {
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"code":    4001,
			"message": "用户未认证",
		})
		return
	}
	user := userInterface.(*models.User)

	// 检查信使级别
	courierLevel := h.getCourierLevel(models.UserRole(user.Role))
	if courierLevel < 2 {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"code":    4003,
			"message": "需要二级或以上信使权限",
		})
		return
	}

	applicationID := c.Param("application_id")
	var req struct {
		Status    string `json:"status" binding:"required,oneof=approved rejected"`
		PointCode string `json:"point_code"`
		Reason    string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"code":    4001,
			"message": "请求参数无效",
			"error":   err.Error(),
		})
		return
	}

	// 获取申请信息
	application, err := h.opcodeService.GetApplicationByID(applicationID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"code":    4004,
			"message": "申请不存在",
		})
		return
	}

	// 获取信使信息
	courier, err := h.courierService.GetCourierByUserID(user.ID)
	if err != nil || courier == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"code":    4004,
			"message": "信使信息不存在",
		})
		return
	}

	// 验证权限：检查申请的编码是否在信使管理范围内
	if !h.canManageOPCode(application, courier, courierLevel) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"code":    4003,
			"message": "该申请不在您的管理范围内",
		})
		return
	}

	// 处理审核
	if req.Status == "approved" {
		if req.PointCode == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"code":    4001,
				"message": "通过申请需要指定后两位编码",
			})
			return
		}

		// 分配OP Code
		err = h.opcodeService.AssignOPCode(user.ID, applicationID, req.PointCode)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"code":    5001,
				"message": "分配编码失败",
				"error":   err.Error(),
			})
			return
		}
	} else {
		// 拒绝申请
		err = h.opcodeService.RejectApplication(applicationID, user.ID, req.Reason)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"code":    5001,
				"message": "拒绝申请失败",
				"error":   err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"code":    200,
		"message": "审核成功",
	})
}

// CreateOPCode 创建OP Code（信使专用）
func (h *CourierOPCodeHandler) CreateOPCode(c *gin.Context) {
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"code":    4001,
			"message": "用户未认证",
		})
		return
	}
	user := userInterface.(*models.User)

	// 检查信使级别
	courierLevel := h.getCourierLevel(models.UserRole(user.Role))
	if courierLevel < 2 {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"code":    4003,
			"message": "需要二级或以上信使权限",
		})
		return
	}

	var req struct {
		Code        string `json:"code" binding:"required,len=6"`
		Type        string `json:"type" binding:"required"`
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		IsPublic    bool   `json:"is_public"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"code":    4001,
			"message": "请求参数无效",
			"error":   err.Error(),
		})
		return
	}

	// 获取信使信息
	courier, err := h.courierService.GetCourierByUserID(user.ID)
	if err != nil || courier == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"code":    4004,
			"message": "信使信息不存在",
		})
		return
	}

	// 验证创建权限
	if !h.canCreateOPCode(req.Code, courier, courierLevel) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"code":    4003,
			"message": "您没有创建该编码的权限",
		})
		return
	}

	// 创建OP Code
	opCode := &models.SignalCode{
		Code:        req.Code,
		Type:        req.Type,
		Description: fmt.Sprintf("%s - %s", req.Name, req.Description),
		Status:      "active",
		IsPublic:    req.IsPublic,
		OwnerID:     &user.ID,
	}

	if err := h.opcodeService.CreateOPCode(opCode); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"code":    5001,
			"message": "创建编码失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"code":    200,
		"message": "创建成功",
		"data":    opCode,
	})
}

// GetManagedOPCodes 获取管理的OP Code列表
func (h *CourierOPCodeHandler) GetManagedOPCodes(c *gin.Context) {
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"code":    4001,
			"message": "用户未认证",
		})
		return
	}
	user := userInterface.(*models.User)

	// 获取信使信息
	courier, err := h.courierService.GetCourierByUserID(user.ID)
	if err != nil || courier == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"code":    4004,
			"message": "信使信息不存在",
		})
		return
	}

	// 获取管理范围内的所有OP Code
	prefix := courier.ManagedOPCodePrefix
	if prefix == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"code":    200,
			"message": "获取成功",
			"data":    []interface{}{},
		})
		return
	}

	codes, err := h.opcodeService.GetOPCodesByPrefix(prefix)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"code":    5001,
			"message": "获取编码列表失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"code":    200,
		"message": "获取成功",
		"data":    codes,
	})
}

// 辅助方法：获取信使级别
func (h *CourierOPCodeHandler) getCourierLevel(role string) int {
	switch role {
	case models.RoleCourierLevel1:
		return 1
	case models.RoleCourierLevel2:
		return 2
	case models.RoleCourierLevel3:
		return 3
	case models.RoleCourierLevel4:
		return 4
	default:
		return 0
	}
}

// 辅助方法：根据级别过滤申请
func (h *CourierOPCodeHandler) filterApplicationsByLevel(
	applications []*models.OPCodeApplication,
	courier *models.Courier,
	level int,
) []*models.OPCodeApplication {
	if courier.ManagedOPCodePrefix == "" {
		return []*models.OPCodeApplication{}
	}

	var filtered []*models.OPCodeApplication
	for _, app := range applications {
		// 构建申请的编码前缀
		appPrefix := app.SchoolCode + app.AreaCode

		// 根据级别判断是否可以管理
		switch level {
		case 2: // 二级信使管理具体投递点（完整4位前缀匹配）
			if strings.HasPrefix(appPrefix, courier.ManagedOPCodePrefix[:4]) {
				filtered = append(filtered, app)
			}
		case 3: // 三级信使管理学校内的片区（前2位匹配）
			if strings.HasPrefix(appPrefix, courier.ManagedOPCodePrefix[:2]) {
				filtered = append(filtered, app)
			}
		case 4: // 四级信使管理所有学校
			filtered = append(filtered, app)
		}
	}

	return filtered
}

// 辅助方法：检查是否可以管理该申请
func (h *CourierOPCodeHandler) canManageOPCode(
	application *models.OPCodeApplication,
	courier *models.Courier,
	level int,
) bool {
	if courier.ManagedOPCodePrefix == "" {
		return false
	}

	appPrefix := application.SchoolCode + application.AreaCode

	switch level {
	case 2: // 二级信使
		return strings.HasPrefix(appPrefix, courier.ManagedOPCodePrefix[:4])
	case 3: // 三级信使
		return strings.HasPrefix(appPrefix, courier.ManagedOPCodePrefix[:2])
	case 4: // 四级信使
		return true
	default:
		return false
	}
}

// 辅助方法：检查是否可以创建该编码
func (h *CourierOPCodeHandler) canCreateOPCode(
	code string,
	courier *models.Courier,
	level int,
) bool {
	if len(code) != 6 || courier.ManagedOPCodePrefix == "" {
		return false
	}

	switch level {
	case 2: // 二级信使只能创建其管理的4位前缀下的编码
		return strings.HasPrefix(code, courier.ManagedOPCodePrefix[:4])
	case 3: // 三级信使可以创建其管理的2位前缀下的编码
		return strings.HasPrefix(code, courier.ManagedOPCodePrefix[:2])
	case 4: // 四级信使可以创建任何编码
		return true
	default:
		return false
	}
}

// UpdateOPCode 更新OP Code信息
func (h *CourierOPCodeHandler) UpdateOPCode(c *gin.Context) {
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"code":    4001,
			"message": "用户未认证",
		})
		return
	}
	user := userInterface.(*models.User)

	opcodeID := c.Param("id")
	
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		IsPublic    bool   `json:"is_public"`
		IsActive    bool   `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"code":    4001,
			"message": "请求参数无效",
			"error":   err.Error(),
		})
		return
	}

	// 获取OP Code信息
	opCode, err := h.opcodeService.GetOPCodeByID(opcodeID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"code":    4004,
			"message": "编码不存在",
		})
		return
	}

	// 获取信使信息
	courier, err := h.courierService.GetCourierByUserID(user.ID)
	if err != nil || courier == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"code":    4004,
			"message": "信使信息不存在",
		})
		return
	}

	// 验证权限
	courierLevel := h.getCourierLevel(models.UserRole(user.Role))
	if !h.canCreateOPCode(opCode.Code, courier, courierLevel) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"code":    4003,
			"message": "您没有修改该编码的权限",
		})
		return
	}

	// 更新信息
	opCode.Description = fmt.Sprintf("%s - %s", req.Name, req.Description)
	opCode.IsPublic = req.IsPublic
	if req.IsActive {
		opCode.Status = "active"
	} else {
		opCode.Status = "inactive"
	}

	if err := h.opcodeService.UpdateOPCode(opCode); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"code":    5001,
			"message": "更新失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"code":    200,
		"message": "更新成功",
		"data":    opCode,
	})
}

// DeleteOPCode 删除OP Code
func (h *CourierOPCodeHandler) DeleteOPCode(c *gin.Context) {
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"code":    4001,
			"message": "用户未认证",
		})
		return
	}
	user := userInterface.(*models.User)

	// 检查信使级别
	courierLevel := h.getCourierLevel(models.UserRole(user.Role))
	if courierLevel < 3 {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"code":    4003,
			"message": "需要三级或以上信使权限才能删除编码",
		})
		return
	}

	opcodeID := c.Param("id")

	// 获取OP Code信息
	opCode, err := h.opcodeService.GetOPCodeByID(opcodeID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"code":    4004,
			"message": "编码不存在",
		})
		return
	}

	// 获取信使信息
	courier, err := h.courierService.GetCourierByUserID(user.ID)
	if err != nil || courier == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"code":    4004,
			"message": "信使信息不存在",
		})
		return
	}

	// 验证权限
	if !h.canCreateOPCode(opCode.Code, courier, courierLevel) {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"code":    4003,
			"message": "您没有删除该编码的权限",
		})
		return
	}

	// 删除编码
	if err := h.opcodeService.DeleteOPCode(opcodeID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"code":    5001,
			"message": "删除失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"code":    200,
		"message": "删除成功",
	})
}