package handlers

import (
	"courier-service/internal/middleware"
	"courier-service/internal/models"
	"courier-service/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ScanHandler 扫码处理器
type ScanHandler struct {
	taskService     *services.TaskService
	locationService *services.LocationService
}

// NewScanHandler 创建扫码处理器
func NewScanHandler(taskService *services.TaskService, locationService *services.LocationService) *ScanHandler {
	return &ScanHandler{
		taskService:     taskService,
		locationService: locationService,
	}
}

// ScanLetterCode 扫描信件二维码更新状态
func (h *ScanHandler) ScanLetterCode(c *gin.Context) {
	letterCode := c.Param("letter_code")
	courierID := middleware.GetUserID(c)

	if courierID == "" {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(
			models.CodeUnauthorized,
			"User ID not found",
			nil,
		))
		return
	}

	var request models.ScanRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.CodeParamError,
			"Invalid request parameters",
			err.Error(),
		))
		return
	}

	// 验证扫码动作
	if !isValidScanAction(request.Action) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.CodeParamError,
			"Invalid scan action",
			"Action must be one of: collected, in_transit, delivered, failed",
		))
		return
	}

	response, err := h.taskService.UpdateTaskStatus(letterCode, courierID, &request)
	if err != nil {
		c.JSON(http.StatusConflict, models.ErrorResponse(
			models.CodeConflict,
			"Failed to update task status",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(response))
}

// GetScanHistory 获取扫码历史记录
func (h *ScanHandler) GetScanHistory(c *gin.Context) {
	letterCode := c.Param("letter_code")
	courierID := middleware.GetUserID(c)

	// 这里可以添加获取扫码历史的逻辑
	// 暂时返回空响应
	response := map[string]interface{}{
		"letter_code": letterCode,
		"courier_id":  courierID,
		"scan_records": []interface{}{},
	}

	c.JSON(http.StatusOK, models.SuccessResponse(response))
}

// isValidScanAction 验证扫码动作是否有效
func isValidScanAction(action string) bool {
	validActions := []string{
		models.ScanActionCollected,
		models.ScanActionInTransit,
		models.ScanActionDelivered,
		models.ScanActionFailed,
	}

	for _, validAction := range validActions {
		if action == validAction {
			return true
		}
	}
	return false
}

// RegisterScanRoutes 注册扫码相关路由
func RegisterScanRoutes(router *gin.RouterGroup, taskService *services.TaskService, locationService *services.LocationService) {
	handler := NewScanHandler(taskService, locationService)

	router.POST("/scan/:letter_code", handler.ScanLetterCode)
	router.GET("/scan/:letter_code/history", handler.GetScanHistory)
}