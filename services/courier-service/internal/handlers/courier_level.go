package handlers

import (
	"courier-service/internal/models"
	"courier-service/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CourierLevelHandler 信使等级管理处理器
type CourierLevelHandler struct {
	courierService    *services.CourierService
	levelService      *services.CourierLevelService
}

// NewCourierLevelHandler 创建信使等级处理器
func NewCourierLevelHandler(courierService *services.CourierService, levelService *services.CourierLevelService) *CourierLevelHandler {
	return &CourierLevelHandler{
		courierService: courierService,
		levelService:   levelService,
	}
}

// CheckLevel 验证信使等级和权限范围
func (h *CourierLevelHandler) CheckLevel(c *gin.Context) {
	courierID := c.Param("courier_id")
	if courierID == "" {
		courierID = c.GetString("user_id") // 从JWT中获取
	}

	// 获取信使等级信息
	levelInfo, err := h.levelService.GetCourierLevelInfo(courierID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(
			models.CodeNotFound,
			"Courier not found",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(levelInfo))
}

// GetUpgradeRequests 获取等级升级申请列表
func (h *CourierLevelHandler) GetUpgradeRequests(c *gin.Context) {
	// 只有管理员或高级信使可以查看升级申请
	courierLevel, exists := c.Get("courier_level")
	if !exists || courierLevel.(models.CourierLevel) < models.LevelThree {
		c.JSON(http.StatusForbidden, models.ErrorResponse(
			models.CodeUnauthorized,
			"Insufficient permission to view upgrade requests",
			nil,
		))
		return
	}

	var query struct {
		Status string `form:"status"`
		Limit  int    `form:"limit"`
		Offset int    `form:"offset"`
	}

	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.CodeParamError,
			"Invalid query parameters",
			err.Error(),
		))
		return
	}

	// 设置默认值
	if query.Limit <= 0 {
		query.Limit = 20
	}
	if query.Limit > 100 {
		query.Limit = 100
	}

	requests, total, err := h.levelService.GetUpgradeRequests(query.Status, query.Limit, query.Offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.CodeInternalError,
			"Failed to get upgrade requests",
			err.Error(),
		))
		return
	}

	response := map[string]interface{}{
		"requests": requests,
		"total":    total,
		"limit":    query.Limit,
		"offset":   query.Offset,
	}

	c.JSON(http.StatusOK, models.SuccessResponse(response))
}

// ProcessUpgradeRequest 处理等级升级申请
func (h *CourierLevelHandler) ProcessUpgradeRequest(c *gin.Context) {
	requestID := c.Param("request_id")
	reviewerID := c.GetString("user_id")

	var request struct {
		Action  string `json:"action" binding:"required,oneof=approve reject"`
		Comment string `json:"comment"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.CodeParamError,
			"Invalid request parameters",
			err.Error(),
		))
		return
	}

	// 检查权限
	courierLevel, exists := c.Get("courier_level")
	if !exists || courierLevel.(models.CourierLevel) < models.LevelThree {
		c.JSON(http.StatusForbidden, models.ErrorResponse(
			models.CodeUnauthorized,
			"Insufficient permission to process upgrade requests",
			nil,
		))
		return
	}

	err := h.levelService.ProcessUpgradeRequest(requestID, request.Action, request.Comment, reviewerID)
	if err != nil {
		c.JSON(http.StatusConflict, models.ErrorResponse(
			models.CodeConflict,
			"Failed to process upgrade request",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"request_id": requestID,
		"action":     request.Action,
		"processed_by": reviewerID,
		"message":    "Upgrade request processed successfully",
	}))
}

// SubmitUpgradeRequest 提交等级升级申请
func (h *CourierLevelHandler) SubmitUpgradeRequest(c *gin.Context) {
	courierID := c.GetString("user_id")

	var request struct {
		RequestLevel models.CourierLevel `json:"request_level" binding:"required,min=2,max=4"`
		Reason       string              `json:"reason" binding:"required"`
		Evidence     map[string]interface{} `json:"evidence"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.CodeParamError,
			"Invalid request parameters",
			err.Error(),
		))
		return
	}

	// 检查升级条件
	canUpgrade, reason := h.levelService.CheckUpgradeEligibility(courierID, request.RequestLevel)
	if !canUpgrade {
		c.JSON(http.StatusConflict, models.ErrorResponse(
			models.CodeConflict,
			"Upgrade requirements not met",
			reason,
		))
		return
	}

	upgradeRequest, err := h.levelService.SubmitUpgradeRequest(courierID, request.RequestLevel, request.Reason, request.Evidence)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.CodeInternalError,
			"Failed to submit upgrade request",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusCreated, models.SuccessResponse(upgradeRequest))
}

// GetZoneManagement 获取管理区域信息
func (h *CourierLevelHandler) GetZoneManagement(c *gin.Context) {
	courierID := c.GetString("user_id")

	zones, err := h.levelService.GetCourierZones(courierID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.CodeInternalError,
			"Failed to get zone information",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"courier_id": courierID,
		"zones":      zones,
	}))
}

// AssignZone 分配信使管理区域
func (h *CourierLevelHandler) AssignZone(c *gin.Context) {
	assignerID := c.GetString("user_id")

	var request struct {
		CourierID string                  `json:"courier_id" binding:"required"`
		ZoneType  models.CourierZoneType  `json:"zone_type" binding:"required"`
		ZoneID    string                  `json:"zone_id" binding:"required"`
		ZoneName  string                  `json:"zone_name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.CodeParamError,
			"Invalid request parameters",
			err.Error(),
		))
		return
	}

	// 检查分配权限
	canAssign, reason := h.levelService.CanAssignZone(assignerID, request.CourierID, request.ZoneType)
	if !canAssign {
		c.JSON(http.StatusForbidden, models.ErrorResponse(
			models.CodeUnauthorized,
			"Insufficient permission to assign zone",
			reason,
		))
		return
	}

	err := h.levelService.AssignZone(request.CourierID, request.ZoneType, request.ZoneID, request.ZoneName, assignerID)
	if err != nil {
		c.JSON(http.StatusConflict, models.ErrorResponse(
			models.CodeConflict,
			"Failed to assign zone",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"courier_id": request.CourierID,
		"zone_type":  request.ZoneType,
		"zone_id":    request.ZoneID,
		"zone_name":  request.ZoneName,
		"assigned_by": assignerID,
		"message":    "Zone assigned successfully",
	}))
}

// GetPerformanceScope 获取权限范围内的绩效数据
func (h *CourierLevelHandler) GetPerformanceScope(c *gin.Context) {
	courierID := c.GetString("user_id")

	// 获取查询参数
	var query struct {
		TimeRange string `form:"time_range"` // 1d, 7d, 30d
		ZoneType  string `form:"zone_type"`
		ZoneID    string `form:"zone_id"`
	}

	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.CodeParamError,
			"Invalid query parameters",
			err.Error(),
		))
		return
	}

	// 设置默认时间范围
	if query.TimeRange == "" {
		query.TimeRange = "7d"
	}

	// 检查查看权限
	canView, zones := h.levelService.GetViewablePerformanceScope(courierID)
	if !canView {
		c.JSON(http.StatusForbidden, models.ErrorResponse(
			models.CodeUnauthorized,
			"No permission to view performance data",
			nil,
		))
		return
	}

	// 获取绩效数据
	performanceData, err := h.levelService.GetPerformanceData(courierID, zones, query.TimeRange, query.ZoneType, query.ZoneID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.CodeInternalError,
			"Failed to get performance data",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(performanceData))
}

// GetLevelsConfig 获取所有等级配置
func (h *CourierLevelHandler) GetLevelsConfig(c *gin.Context) {
	levelsConfig := map[string]interface{}{
		"levels": []map[string]interface{}{
			{
				"level":       1,
				"name":        "一级信使",
				"zone_type":   "building",
				"permissions": models.DefaultPermissionMatrix[models.LevelOne],
				"description": "负责本楼栋信件扫码登记、状态变更和向上转交",
			},
			{
				"level":       2,
				"name":        "二级信使",
				"zone_type":   "area",
				"permissions": models.DefaultPermissionMatrix[models.LevelTwo],
				"description": "负责片区管理、打包分拣、信封分发和接收一级转交",
			},
			{
				"level":       3,
				"name":        "三级信使",
				"zone_type":   "campus",
				"permissions": models.DefaultPermissionMatrix[models.LevelThree],
				"description": "负责全校权限、用户反馈处理和校级绩效查看",
			},
			{
				"level":       4,
				"name":        "四级信使",
				"zone_type":   "city",
				"permissions": models.DefaultPermissionMatrix[models.LevelFour],
				"description": "负责全域权限、多校管理，不可向上转交",
			},
		},
		"permissions": []map[string]interface{}{
			{"code": "scan", "name": "扫码登记", "description": "扫描信件二维码进行状态更新"},
			{"code": "status_change", "name": "状态变更", "description": "更改信件投递状态"},
			{"code": "handover", "name": "向上转交", "description": "将任务转交给上级信使"},
			{"code": "package", "name": "打包分拣", "description": "对信件进行打包和分拣"},
			{"code": "distribute", "name": "信封分发", "description": "分发信件到下级信使"},
			{"code": "receive_handover", "name": "接收转交", "description": "接收下级信使的转交任务"},
			{"code": "feedback", "name": "用户反馈处理", "description": "处理用户投诉和反馈"},
			{"code": "performance", "name": "绩效查看", "description": "查看绩效和统计数据"},
		},
		"zone_types": []map[string]interface{}{
			{"code": "building", "name": "楼栋", "description": "单个楼栋或宿舍楼"},
			{"code": "area", "name": "片区", "description": "多个楼栋组成的片区"},
			{"code": "campus", "name": "校区", "description": "整个校园区域"},
			{"code": "city", "name": "城市", "description": "整个城市范围"},
		},
	}

	c.JSON(http.StatusOK, models.SuccessResponse(levelsConfig))
}

// RegisterCourierLevelRoutes 注册信使等级相关路由
func RegisterCourierLevelRoutes(router *gin.RouterGroup, courierService *services.CourierService, levelService *services.CourierLevelService) {
	handler := NewCourierLevelHandler(courierService, levelService)

	// 公开接口
	router.GET("/levels/config", handler.GetLevelsConfig)

	// 需要认证的接口
	authenticated := router.Group("")
	// authenticated.Use(middleware.JWTAuth()) // 假设已经在上级应用了

	// 个人等级信息
	authenticated.GET("/level/check", handler.CheckLevel)
	authenticated.GET("/level/check/:courier_id", handler.CheckLevel)
	authenticated.GET("/zone/management", handler.GetZoneManagement)
	authenticated.GET("/performance/scope", handler.GetPerformanceScope)

	// 升级申请
	authenticated.POST("/level/upgrade", handler.SubmitUpgradeRequest)

	// 需要高级权限的接口
	privileged := authenticated.Group("")
	// privileged.Use(middleware.RequireLevel(models.LevelThree)) // 需要三级以上权限

	// 管理功能
	privileged.GET("/level/upgrade-requests", handler.GetUpgradeRequests)
	privileged.PUT("/level/upgrade/:request_id", handler.ProcessUpgradeRequest)
	privileged.POST("/zone/assign", handler.AssignZone)
}