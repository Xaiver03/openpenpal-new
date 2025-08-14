package handlers

import (
	"courier-service/internal/models"
	"courier-service/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// HierarchyHandler 层级管理处理器
type HierarchyHandler struct {
	hierarchyService *services.HierarchyService
}

// NewHierarchyHandler 创建层级管理处理器
func NewHierarchyHandler(hierarchyService *services.HierarchyService) *HierarchyHandler {
	return &HierarchyHandler{
		hierarchyService: hierarchyService,
	}
}

// CreateSubordinate 创建下级信使
// POST /api/courier/hierarchy/subordinates
func (h *HierarchyHandler) CreateSubordinate(c *gin.Context) {
	// 获取当前管理者ID (从JWT中获取)
	managerUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
		return
	}

	// 通过UserID查找信使ID
	var manager models.Courier
	if err := h.hierarchyService.GetDB().Where("user_id = ?", managerUserID).First(&manager).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "信使身份未找到"})
		return
	}

	var req models.CreateSubordinateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	subordinate, err := h.hierarchyService.CreateSubordinate(manager.ID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Code:    0,
		Message: "下级信使创建成功",
		Data:    subordinate,
	})
}

// GetSubordinates 获取下级信使列表
// GET /api/courier/hierarchy/subordinates
func (h *HierarchyHandler) GetSubordinates(c *gin.Context) {
	managerUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
		return
	}

	var manager models.Courier
	if err := h.hierarchyService.GetDB().Where("user_id = ?", managerUserID).First(&manager).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "信使身份未找到"})
		return
	}

	subordinates, err := h.hierarchyService.GetSubordinates(manager.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Code:    0,
		Message: "获取成功",
		Data:    subordinates,
	})
}

// AssignZone 分配管理区域
// PUT /api/courier/hierarchy/subordinates/:id/zone
func (h *HierarchyHandler) AssignZone(c *gin.Context) {
	managerUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
		return
	}

	courierIDStr := c.Param("id")
	if courierIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "信使ID不能为空"})
		return
	}

	var manager models.Courier
	if err := h.hierarchyService.GetDB().Where("user_id = ?", managerUserID).First(&manager).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "信使身份未找到"})
		return
	}

	var req models.AssignZoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	if err := h.hierarchyService.AssignZone(manager.ID, courierIDStr, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Code:    0,
		Message: "区域分配成功",
		Data:    nil,
	})
}

// TransferSubordinate 转移下级信使归属
// PUT /api/courier/hierarchy/subordinates/:id/transfer
func (h *HierarchyHandler) TransferSubordinate(c *gin.Context) {
	managerUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
		return
	}

	courierIDStr := c.Param("id")
	if courierIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "信使ID不能为空"})
		return
	}

	var manager models.Courier
	if err := h.hierarchyService.GetDB().Where("user_id = ?", managerUserID).First(&manager).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "信使身份未找到"})
		return
	}

	var req models.TransferSubordinateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	if err := h.hierarchyService.TransferSubordinate(manager.ID, courierIDStr, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Code:    0,
		Message: "转移成功",
		Data:    nil,
	})
}

// GetHierarchy 获取层级结构
// GET /api/courier/hierarchy
func (h *HierarchyHandler) GetHierarchy(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
		return
	}

	var courier models.Courier
	if err := h.hierarchyService.GetDB().Where("user_id = ?", userID).First(&courier).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "信使身份未找到"})
		return
	}

	hierarchy, err := h.hierarchyService.GetHierarchy(courier.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Code:    0,
		Message: "获取成功",
		Data:    hierarchy,
	})
}

// GetSubordinateDetail 获取下级信使详情
// GET /api/courier/hierarchy/subordinates/:id
func (h *HierarchyHandler) GetSubordinateDetail(c *gin.Context) {
	managerUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
		return
	}

	courierIDStr := c.Param("id")
	if courierIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "信使ID不能为空"})
		return
	}

	var manager models.Courier
	if err := h.hierarchyService.GetDB().Where("user_id = ?", managerUserID).First(&manager).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "信使身份未找到"})
		return
	}

	// 检查权限
	var subordinate models.Courier
	if err := h.hierarchyService.GetDB().Where("id = ?", courierIDStr).First(&subordinate).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "信使不存在"})
		return
	}

	if !subordinate.IsSubordinateOf(manager.ID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权限查看该信使信息"})
		return
	}

	// 获取详细层级信息
	hierarchy, err := h.hierarchyService.GetHierarchy(courierIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Code:    0,
		Message: "获取成功",
		Data:    hierarchy,
	})
}

// RegisterHierarchyRoutes 注册层级管理路由
func RegisterHierarchyRoutes(router *gin.RouterGroup, hierarchyService *services.HierarchyService) {
	handler := NewHierarchyHandler(hierarchyService)

	hierarchy := router.Group("/hierarchy")
	{
		// 层级信息
		hierarchy.GET("", handler.GetHierarchy)

		// 下级信使管理
		subordinates := hierarchy.Group("/subordinates")
		{
			subordinates.POST("", handler.CreateSubordinate)
			subordinates.GET("", handler.GetSubordinates)
			subordinates.GET("/:id", handler.GetSubordinateDetail)
			subordinates.PUT("/:id/zone", handler.AssignZone)
			subordinates.PUT("/:id/transfer", handler.TransferSubordinate)
		}
	}
}
