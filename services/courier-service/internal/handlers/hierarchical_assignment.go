package handlers

import (
	"courier-service/internal/models"
	"courier-service/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// HierarchicalAssignmentHandler 层级分配处理器
type HierarchicalAssignmentHandler struct {
	service *services.HierarchicalAssignmentService
}

// NewHierarchicalAssignmentHandler 创建层级分配处理器
func NewHierarchicalAssignmentHandler(service *services.HierarchicalAssignmentService) *HierarchicalAssignmentHandler {
	return &HierarchicalAssignmentHandler{
		service: service,
	}
}

// AssignTaskByHierarchy 根据层级分配任务
// POST /api/courier/hierarchy/assign-task
func (h *HierarchicalAssignmentHandler) AssignTaskByHierarchy(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
		return
	}

	// 获取信使ID
	var courier models.Courier
	if err := h.service.GetDB().Where("user_id = ?", userID).First(&courier).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "信使身份未找到"})
		return
	}

	var req models.HierarchicalTaskAssignmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	task, err := h.service.AssignTaskByHierarchy(courier.ID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Code:    0,
		Message: "任务分配成功",
		Data:    task,
	})
}

// BatchAssignByHierarchy 批量层级分配任务
// POST /api/courier/hierarchy/batch-assign
func (h *HierarchicalAssignmentHandler) BatchAssignByHierarchy(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
		return
	}

	// 获取信使ID
	var courier models.Courier
	if err := h.service.GetDB().Where("user_id = ?", userID).First(&courier).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "信使身份未找到"})
		return
	}

	var req models.BatchHierarchicalAssignmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	response, err := h.service.BatchAssignByHierarchy(courier.ID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Code:    0,
		Message: "批量分配完成",
		Data:    response,
	})
}

// GetAssignmentHistory 获取分配历史
// GET /api/courier/hierarchy/assignment-history
func (h *HierarchicalAssignmentHandler) GetAssignmentHistory(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
		return
	}

	// 获取信使ID
	var courier models.Courier
	if err := h.service.GetDB().Where("user_id = ?", userID).First(&courier).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "信使身份未找到"})
		return
	}

	limit := 20
	offset := 0

	if l := c.Query("limit"); l != "" {
		if parsedLimit, err := strconv.Atoi(l); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	if o := c.Query("offset"); o != "" {
		if parsedOffset, err := strconv.Atoi(o); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	response, err := h.service.GetAssignmentHistory(courier.ID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Code:    0,
		Message: "获取成功",
		Data:    response,
	})
}

// ReassignTask 重新分配任务
// POST /api/courier/hierarchy/reassign-task
func (h *HierarchicalAssignmentHandler) ReassignTask(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
		return
	}

	// 获取信使ID
	var courier models.Courier
	if err := h.service.GetDB().Where("user_id = ?", userID).First(&courier).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "信使身份未找到"})
		return
	}

	var req models.TaskReassignmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	task, err := h.service.ReassignTask(courier.ID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Code:    0,
		Message: "重新分配成功",
		Data:    task,
	})
}

// GetAssignableSubordinates 获取可分配的下级信使列表
// GET /api/courier/hierarchy/assignable-subordinates
func (h *HierarchicalAssignmentHandler) GetAssignableSubordinates(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
		return
	}

	// 获取信使ID
	var courier models.Courier
	if err := h.service.GetDB().Where("user_id = ?", userID).First(&courier).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "信使身份未找到"})
		return
	}

	taskZone := c.Query("task_zone")
	
	// 获取所有下级信使
	var subordinates []models.Courier
	query := h.service.GetDB().Where("parent_id = ? OR (level <= ? AND status = ?)", 
		courier.ID, courier.Level, models.CourierStatusApproved)

	// 如果指定了任务区域，进一步筛选
	if taskZone != "" {
		query = query.Where("zone_code LIKE ?", taskZone+"%")
	}

	if err := query.Order("level DESC, points DESC").Find(&subordinates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取下级信使失败"})
		return
	}

	// 构建响应数据
	var response []gin.H
	for _, sub := range subordinates {
		response = append(response, gin.H{
			"id":         sub.ID,
			"user_id":    sub.UserID,
			"zone":       sub.Zone,
			"zone_type":  sub.ZoneType,
			"zone_code":  sub.ZoneCode,
			"level":      sub.Level,
			"rating":     sub.Rating,
			"points":     sub.Points,
			"experience": sub.Experience,
			"status":     sub.Status,
		})
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Code:    0,
		Message: "获取成功",
		Data:    response,
	})
}

// GetAssignmentStats 获取分配统计信息
// GET /api/courier/hierarchy/assignment-stats
func (h *HierarchicalAssignmentHandler) GetAssignmentStats(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
		return
	}

	// 获取信使ID
	var courier models.Courier
	if err := h.service.GetDB().Where("user_id = ?", userID).First(&courier).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "信使身份未找到"})
		return
	}

	// 统计分配数据
	var totalAssigned int64
	var todayAssigned int64
	var thisWeekAssigned int64

	// 总分配数
	h.service.GetDB().Model(&models.TaskAssignmentHistory{}).
		Where("assigned_by = ?", courier.ID).
		Count(&totalAssigned)

	// 今日分配数
	h.service.GetDB().Model(&models.TaskAssignmentHistory{}).
		Where("assigned_by = ? AND DATE(created_at) = CURRENT_DATE", courier.ID).
		Count(&todayAssigned)

	// 本周分配数
	h.service.GetDB().Model(&models.TaskAssignmentHistory{}).
		Where("assigned_by = ? AND created_at >= DATE_TRUNC('week', CURRENT_DATE)", courier.ID).
		Count(&thisWeekAssigned)

	// 按分配类型统计
	var typeStats []map[string]interface{}
	h.service.GetDB().Model(&models.TaskAssignmentHistory{}).
		Select("assignment_type, COUNT(*) as count").
		Where("assigned_by = ?", courier.ID).
		Group("assignment_type").
		Scan(&typeStats)

	// 下级信使数量
	var subordinateCount int64
	h.service.GetDB().Model(&models.Courier{}).
		Where("parent_id = ?", courier.ID).
		Count(&subordinateCount)

	stats := gin.H{
		"total_assigned":      totalAssigned,
		"today_assigned":      todayAssigned,
		"this_week_assigned":  thisWeekAssigned,
		"assignment_by_type":  typeStats,
		"subordinate_count":   subordinateCount,
		"management_level":    courier.Level,
		"zone_type":          courier.ZoneType,
		"can_assign":         courier.Level >= models.CourierLevelTwo,
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Code:    0,
		Message: "获取成功",
		Data:    stats,
	})
}

// GetPendingAssignments 获取待处理的分配请求
// GET /api/courier/hierarchy/pending-assignments
func (h *HierarchicalAssignmentHandler) GetPendingAssignments(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
		return
	}

	// 获取信使ID
	var courier models.Courier
	if err := h.service.GetDB().Where("user_id = ?", userID).First(&courier).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "信使身份未找到"})
		return
	}

	// 获取管理范围内的待分配任务
	var tasks []models.Task
	query := h.service.GetDB().Where("status = ? AND courier_id IS NULL", models.TaskStatusAvailable)

	// 根据管理级别筛选任务范围
	switch courier.ZoneType {
	case models.ZoneTypeBuilding:
		query = query.Where("pickup_location LIKE ?", "%"+courier.ZoneCode+"%")
	case models.ZoneTypeArea:
		query = query.Where("pickup_location LIKE ?", "%"+courier.ZoneCode+"%")
	case models.ZoneTypeSchool:
		query = query.Where("pickup_location LIKE ?", "%"+courier.ZoneCode+"%")
	case models.ZoneTypeCity:
		// 城市级可以看到所有任务，但这里可能需要根据实际需求调整
	}

	if err := query.Order("priority DESC, created_at ASC").
		Limit(50).
		Find(&tasks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取待分配任务失败"})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Code:    0,
		Message: "获取成功",
		Data: gin.H{
			"tasks": tasks,
			"count": len(tasks),
		},
	})
}

// RegisterHierarchicalAssignmentRoutes 注册层级分配路由
func RegisterHierarchicalAssignmentRoutes(router *gin.RouterGroup, service *services.HierarchicalAssignmentService) {
	handler := NewHierarchicalAssignmentHandler(service)

	hierarchy := router.Group("/hierarchy")
	{
		// 任务分配
		hierarchy.POST("/assign-task", handler.AssignTaskByHierarchy)
		hierarchy.POST("/batch-assign", handler.BatchAssignByHierarchy)
		hierarchy.POST("/reassign-task", handler.ReassignTask)

		// 查询接口
		hierarchy.GET("/assignment-history", handler.GetAssignmentHistory)
		hierarchy.GET("/assignable-subordinates", handler.GetAssignableSubordinates)
		hierarchy.GET("/assignment-stats", handler.GetAssignmentStats)
		hierarchy.GET("/pending-assignments", handler.GetPendingAssignments)
	}
}