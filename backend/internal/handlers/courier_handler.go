package handlers

import (
	"strconv"

	"openpenpal-backend/internal/middleware"
	"openpenpal-backend/internal/models"
	"openpenpal-backend/internal/services"

	"github.com/gin-gonic/gin"
	"openpenpal-backend/internal/pkg/response"
)

type CourierHandler struct {
	courierService *services.CourierService
}

func NewCourierHandler(courierService *services.CourierService) *CourierHandler {
	return &CourierHandler{
		courierService: courierService,
	}
}

// ApplyCourier 申请成为信使
// @Summary 申请成为信使
// @Description 用户申请成为信使
// @Tags courier
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param application body models.CourierApplication true "申请信息"
// @Success 200 {object} models.Courier
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/courier/apply [post]
func (h *CourierHandler) ApplyCourier(c *gin.Context) {
	resp := response.NewGinResponse()

	var req models.CourierApplication
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "参数验证失败: "+err.Error())
		return
	}

	// 从JWT中获取用户ID
	userID, exists := middleware.GetUserID(c)
	if !exists {
		resp.Unauthorized(c, "用户认证失败")
		return
	}

	courier, err := h.courierService.ApplyCourier(userID, &req)
	if err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	resp.SuccessWithMessage(c, "申请提交成功", courier)
}

// GetCourierStatus 获取信使状态
// @Summary 获取用户信使状态
// @Description 获取当前用户的信使申请状态和信息
// @Tags courier
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} models.CourierStatus
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/courier/status [get]
func (h *CourierHandler) GetCourierStatus(c *gin.Context) {
	resp := response.NewGinResponse()

	userID, exists := middleware.GetUserID(c)
	if !exists {
		resp.Unauthorized(c, "用户认证失败")
		return
	}

	status, err := h.courierService.GetCourierStatus(userID)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.Success(c, status)
}

// GetCourierProfile 获取信使详细信息
// @Summary 获取信使详细信息
// @Description 获取当前用户的信使详细信息
// @Tags courier
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} models.Courier
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/courier/profile [get]
func (h *CourierHandler) GetCourierProfile(c *gin.Context) {
	resp := response.NewGinResponse()

	userID, exists := middleware.GetUserID(c)
	if !exists {
		resp.Unauthorized(c, "用户认证失败")
		return
	}

	courier, err := h.courierService.GetCourierByUserID(userID)
	if err != nil {
		if err.Error() == "信使信息不存在" {
			resp.NotFound(c, err.Error())
		} else {
			resp.InternalServerError(c, err.Error())
		}
		return
	}

	resp.Success(c, courier)
}

// GetCourierStats 获取信使统计信息（公开接口）
// @Summary 获取信使统计信息
// @Description 获取平台信使相关的统计数据
// @Tags courier
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/courier/stats [get]
func (h *CourierHandler) GetCourierStats(c *gin.Context) {
	resp := response.NewGinResponse()

	stats, err := h.courierService.GetCourierStats()
	if err != nil {
		resp.InternalServerError(c, "获取统计信息失败")
		return
	}

	resp.Success(c, stats)
}

// --- 管理员接口 ---

// GetPendingApplications 获取待审核申请（管理员）
// @Summary 获取待审核申请
// @Description 管理员获取待审核的信使申请列表
// @Tags admin
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {array} models.Courier
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/admin/courier/applications [get]
func (h *CourierHandler) GetPendingApplications(c *gin.Context) {
	resp := response.NewGinResponse()

	applications, err := h.courierService.GetPendingApplications()
	if err != nil {
		resp.InternalServerError(c, "获取申请列表失败")
		return
	}

	resp.Success(c, applications)
}

// ApproveCourierApplication 审核通过信使申请（管理员）
// @Summary 审核通过信使申请
// @Description 管理员审核通过信使申请
// @Tags admin
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path int true "信使ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/admin/courier/{id}/approve [post]
func (h *CourierHandler) ApproveCourierApplication(c *gin.Context) {
	resp := response.NewGinResponse()

	courierIDStr := c.Param("id")
	courierID, err := strconv.ParseUint(courierIDStr, 10, 32)
	if err != nil {
		resp.BadRequest(c, "无效的信使ID")
		return
	}

	if err := h.courierService.ApproveCourier(uint(courierID)); err != nil {
		resp.InternalServerError(c, "审核操作失败")
		return
	}

	resp.OK(c, "申请已审核通过")
}

// RejectCourierApplication 拒绝信使申请（管理员）
// @Summary 拒绝信使申请
// @Description 管理员拒绝信使申请
// @Tags admin
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path int true "信使ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/admin/courier/{id}/reject [post]
func (h *CourierHandler) RejectCourierApplication(c *gin.Context) {
	resp := response.NewGinResponse()

	courierIDStr := c.Param("id")
	courierID, err := strconv.ParseUint(courierIDStr, 10, 32)
	if err != nil {
		resp.BadRequest(c, "无效的信使ID")
		return
	}

	if err := h.courierService.RejectCourier(uint(courierID)); err != nil {
		resp.InternalServerError(c, "审核操作失败")
		return
	}

	resp.OK(c, "申请已被拒绝")
}

// --- 四级信使管理API ---

// CreateCourier 创建下级信使
// @Summary 创建下级信使
// @Description 高级信使创建下级信使
// @Tags courier
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param courier body models.CreateCourierRequest true "信使信息"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/courier/create [post]
func (h *CourierHandler) CreateCourier(c *gin.Context) {
	resp := response.NewGinResponse()

	var req models.CreateCourierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "参数验证失败: "+err.Error())
		return
	}

	// 从JWT中获取用户信息
	user, exists := c.Get("user")
	if !exists {
		resp.Unauthorized(c, "用户认证失败")
		return
	}

	courier, err := h.courierService.CreateSubordinateCourier(user.(*models.User), &req)
	if err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	resp.SuccessWithMessage(c, "创建信使成功", courier)
}

// GetSubordinates 获取下级信使列表
// @Summary 获取下级信使列表
// @Description 获取当前信使的所有下级信使列表
// @Tags courier
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/courier/subordinates [get]
func (h *CourierHandler) GetSubordinates(c *gin.Context) {
	resp := response.NewGinResponse()

	userID, exists := middleware.GetUserID(c)
	if !exists {
		resp.Unauthorized(c, "用户认证失败")
		return
	}

	subordinates, err := h.courierService.GetSubordinateCouriers(userID)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.Success(c, gin.H{"couriers": subordinates})
}

// GetCourierInfo 获取当前信使信息
// @Summary 获取当前信使信息
// @Description 获取当前登录信使的详细信息
// @Tags courier
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/courier/me [get]
func (h *CourierHandler) GetCourierInfo(c *gin.Context) {
	resp := response.NewGinResponse()

	user, exists := c.Get("user")
	if !exists {
		resp.Unauthorized(c, "用户认证失败")
		return
	}

	userModel := user.(*models.User)
	courierInfo, err := h.courierService.GetCourierInfoByUser(userModel)
	if err != nil {
		// 返回更详细的错误信息
		resp.Success(c, gin.H{
			"is_courier": false,
			"user_role":  userModel.Role,
			"message":    err.Error(),
		})
		return
	}

	// 返回完整的信息，包括用户角色
	resp.Success(c, gin.H{
		"courier_info": courierInfo,
		"is_courier":   true,
		"user_role":    userModel.Role,
	})
}

// GetMyStats 获取当前信使的详细统计信息
// @Summary 获取当前信使的详细统计信息
// @Description 获取当前登录信使的详细统计数据，包括任务统计、团队数据等
// @Tags courier
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/courier/stats [get]
func (h *CourierHandler) GetMyStats(c *gin.Context) {
	resp := response.NewGinResponse()

	user, exists := c.Get("user")
	if !exists {
		resp.Unauthorized(c, "用户认证失败")
		return
	}

	userModel := user.(*models.User)
	stats, err := h.courierService.GetCourierInfoByUser(userModel)
	if err != nil {
		resp.InternalServerError(c, "获取统计信息失败: "+err.Error())
		return
	}

	resp.Success(c, stats)
}

// === 管理级别API (按信使等级) ===

// GetFirstLevelStats 获取一级信使管理统计
// @Summary 获取一级信使管理统计
// @Description 获取楼栋级信使管理统计信息
// @Tags courier-management
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/courier/management/level-1/stats [get]
func (h *CourierHandler) GetFirstLevelStats(c *gin.Context) {
	resp := response.NewGinResponse()

	stats, err := h.courierService.GetLevelStats(1)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.Success(c, stats)
}

// GetFirstLevelCouriers 获取一级信使列表
// @Summary 获取一级信使列表
// @Description 获取楼栋级信使列表
// @Tags courier-management
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/courier/management/level-1/couriers [get]
func (h *CourierHandler) GetFirstLevelCouriers(c *gin.Context) {
	resp := response.NewGinResponse()

	couriers, err := h.courierService.GetCouriersByLevel(1)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.Success(c, gin.H{"couriers": couriers})
}

// GetSecondLevelStats 获取二级信使管理统计
func (h *CourierHandler) GetSecondLevelStats(c *gin.Context) {
	resp := response.NewGinResponse()

	stats, err := h.courierService.GetLevelStats(2)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.Success(c, stats)
}

// GetSecondLevelCouriers 获取二级信使列表
func (h *CourierHandler) GetSecondLevelCouriers(c *gin.Context) {
	resp := response.NewGinResponse()

	couriers, err := h.courierService.GetCouriersByLevel(2)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.Success(c, gin.H{"couriers": couriers})
}

// GetThirdLevelStats 获取三级信使管理统计
func (h *CourierHandler) GetThirdLevelStats(c *gin.Context) {
	resp := response.NewGinResponse()

	stats, err := h.courierService.GetLevelStats(3)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.Success(c, stats)
}

// GetThirdLevelCouriers 获取三级信使列表
func (h *CourierHandler) GetThirdLevelCouriers(c *gin.Context) {
	resp := response.NewGinResponse()

	couriers, err := h.courierService.GetCouriersByLevel(3)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.Success(c, gin.H{"couriers": couriers})
}

// GetFourthLevelStats 获取四级信使管理统计
func (h *CourierHandler) GetFourthLevelStats(c *gin.Context) {
	resp := response.NewGinResponse()

	stats, err := h.courierService.GetLevelStats(4)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.Success(c, stats)
}

// GetFourthLevelCouriers 获取四级信使列表
func (h *CourierHandler) GetFourthLevelCouriers(c *gin.Context) {
	resp := response.NewGinResponse()

	couriers, err := h.courierService.GetCouriersByLevel(4)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.Success(c, gin.H{"couriers": couriers})
}

// GetCourierCandidates 获取信使候选人列表
// @Summary 获取信使候选人列表
// @Description 获取可以成为信使的用户列表
// @Tags courier-management
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/courier/candidates [get]
func (h *CourierHandler) GetCourierCandidates(c *gin.Context) {
	resp := response.NewGinResponse()

	candidates, err := h.courierService.GetCourierCandidates()
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.Success(c, gin.H{"candidates": candidates})
}

// GetCourierTasks 获取信使任务列表
// @Summary 获取信使任务列表
// @Description 获取当前信使的任务列表，支持状态和优先级筛选
// @Tags courier
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param status query string false "任务状态 (pending,collected,in_transit,delivered,failed)"
// @Param priority query string false "优先级 (normal,urgent)"
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(10)
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/courier/tasks [get]
func (h *CourierHandler) GetCourierTasks(c *gin.Context) {
	resp := response.NewGinResponse()

	// 获取当前用户ID
	userID, exists := middleware.GetUserID(c)
	if !exists {
		resp.Unauthorized(c, "用户认证失败")
		return
	}

	// 获取查询参数
	status := c.Query("status")
	priority := c.Query("priority")
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "10")

	// 转换分页参数
	pageInt, err := strconv.Atoi(page)
	if err != nil || pageInt < 1 {
		pageInt = 1
	}

	limitInt, err := strconv.Atoi(limit)
	if err != nil || limitInt < 1 || limitInt > 100 {
		limitInt = 10
	}

	// 获取任务列表
	tasks, total, err := h.courierService.GetCourierTasks(userID, status, priority, pageInt, limitInt)
	if err != nil {
		resp.InternalServerError(c, "获取任务列表失败: "+err.Error())
		return
	}

	// 返回数据
	resp.Success(c, gin.H{
		"tasks": tasks,
		"pagination": gin.H{
			"page":  pageInt,
			"limit": limitInt,
			"total": total,
			"pages": (total + int64(limitInt) - 1) / int64(limitInt),
		},
	})
}

// GetTaskDetail 获取任务详情
func (h *CourierHandler) GetTaskDetail(c *gin.Context) {
	resp := response.NewGinResponse()

	// 从JWT中获取用户ID
	userID, exists := middleware.GetUserID(c)
	if !exists {
		resp.Unauthorized(c, "用户认证失败")
		return
	}

	taskID := c.Param("id")
	if taskID == "" {
		resp.BadRequest(c, "任务ID不能为空")
		return
	}

	task, err := h.courierService.GetTaskDetail(userID, taskID)
	if err != nil {
		if err.Error() == "task not found" {
			resp.NotFound(c, "任务不存在")
			return
		}
		resp.InternalServerError(c, "获取任务详情失败")
		return
	}

	resp.Success(c, task)
}

// UpdateTaskStatus 更新任务状态
func (h *CourierHandler) UpdateTaskStatus(c *gin.Context) {
	resp := response.NewGinResponse()

	// 从JWT中获取用户ID
	userID, exists := middleware.GetUserID(c)
	if !exists {
		resp.Unauthorized(c, "用户认证失败")
		return
	}

	taskID := c.Param("id")
	if taskID == "" {
		resp.BadRequest(c, "任务ID不能为空")
		return
	}

	var req struct {
		Status   string `json:"status" binding:"required,oneof=accepted collected in_transit delivered failed"`
		Location string `json:"location"`
		Note     string `json:"note"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "参数验证失败: "+err.Error())
		return
	}

	err := h.courierService.UpdateTaskStatus(userID, taskID, req.Status, req.Location, req.Note)
	if err != nil {
		if err.Error() == "task not found" {
			resp.NotFound(c, "任务不存在")
			return
		}
		if err.Error() == "permission denied" {
			resp.Forbidden(c, "无权操作此任务")
			return
		}
		resp.InternalServerError(c, "更新任务状态失败")
		return
	}

	resp.Success(c, gin.H{
		"message": "任务状态更新成功",
		"status":  req.Status,
	})
}

// ScanCode 扫码处理（新的统一扫码接口）
func (h *CourierHandler) ScanCode(c *gin.Context) {
	resp := response.NewGinResponse()

	// 从JWT中获取用户ID
	userID, exists := middleware.GetUserID(c)
	if !exists {
		resp.Unauthorized(c, "用户认证失败")
		return
	}

	var req struct {
		Code      string  `json:"code" binding:"required"`
		Action    string  `json:"action" binding:"required,oneof=pickup deliver"`
		Location  string  `json:"location"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		Note      string  `json:"note"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, "参数验证失败: "+err.Error())
		return
	}

	result, err := h.courierService.ProcessScan(userID, req.Code, req.Action, req.Location, req.Latitude, req.Longitude, req.Note)
	if err != nil {
		if err.Error() == "invalid code" {
			resp.BadRequest(c, "无效的扫码内容")
			return
		}
		if err.Error() == "permission denied" {
			resp.Forbidden(c, "无权处理此信件")
			return
		}
		resp.InternalServerError(c, "扫码处理失败")
		return
	}

	resp.Success(c, result)
}

// GetSubordinatesV2 获取下级信使列表 (备用方法，防止路由冲突)

// GetHierarchyInfo 获取层级信息
// @Summary 获取信使层级信息
// @Description 获取当前信使的层级信息和权限
// @Tags courier
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/courier/hierarchy [get]
func (h *CourierHandler) GetHierarchyInfo(c *gin.Context) {
	resp := response.NewGinResponse()

	user, exists := c.Get("user")
	if !exists {
		resp.Unauthorized(c, "用户认证失败")
		return
	}

	userModel := user.(*models.User)
	hierarchyInfo, err := h.courierService.GetCourierHierarchyInfo(userModel)
	if err != nil {
		resp.InternalServerError(c, "获取层级信息失败: "+err.Error())
		return
	}

	resp.Success(c, hierarchyInfo)
}
