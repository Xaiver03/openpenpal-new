package handlers

import (
	"courier-service/internal/middleware"
	"courier-service/internal/models"
	"courier-service/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CourierHandler 信使处理器
type CourierHandler struct {
	courierService *services.CourierService
}

// NewCourierHandler 创建信使处理器
func NewCourierHandler(courierService *services.CourierService) *CourierHandler {
	return &CourierHandler{
		courierService: courierService,
	}
}

// ApplyCourier 申请成为信使
func (h *CourierHandler) ApplyCourier(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(
			models.CodeUnauthorized,
			"User ID not found",
			nil,
		))
		return
	}

	var application models.CourierApplication
	if err := c.ShouldBindJSON(&application); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.CodeParamError,
			"Invalid request parameters",
			err.Error(),
		))
		return
	}

	courier, err := h.courierService.ApplyCourier(userID, &application)
	if err != nil {
		c.JSON(http.StatusConflict, models.ErrorResponse(
			models.CodeConflict,
			"Application failed",
			err.Error(),
		))
		return
	}

	response := map[string]interface{}{
		"application_id": "CA" + courier.ID,
		"status":         courier.Status,
		"submitted_at":   courier.CreatedAt,
	}

	c.JSON(http.StatusOK, models.SuccessResponse(response))
}

// GetCourierInfo 获取信使信息
func (h *CourierHandler) GetCourierInfo(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(
			models.CodeUnauthorized,
			"User ID not found",
			nil,
		))
		return
	}

	courier, err := h.courierService.GetCourierByUserID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(
			models.CodeNotFound,
			"Courier not found",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(courier))
}

// GetCourierStats 获取信使统计信息
func (h *CourierHandler) GetCourierStats(c *gin.Context) {
	courierID := c.Param("courier_id")
	if courierID == "" {
		courierID = middleware.GetUserID(c)
	}

	if courierID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.CodeParamError,
			"Courier ID is required",
			nil,
		))
		return
	}

	stats, err := h.courierService.GetCourierStats(courierID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.CodeInternalError,
			"Failed to get courier stats",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(stats))
}

// ApproveCourier 审核通过信使申请（管理员功能）
func (h *CourierHandler) ApproveCourier(c *gin.Context) {
	role := middleware.GetUserRole(c)
	if role != "admin" && role != "super_admin" {
		c.JSON(http.StatusForbidden, models.ErrorResponse(
			models.CodeUnauthorized,
			"Admin permission required",
			nil,
		))
		return
	}

	courierIDStr := c.Param("courier_id")
	if courierIDStr == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.CodeParamError,
			"Courier ID is required",
			nil,
		))
		return
	}

	var request struct {
		Note string `json:"note"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.CodeParamError,
			"Invalid request parameters",
			err.Error(),
		))
		return
	}

	err := h.courierService.ApproveCourier(courierIDStr, request.Note)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.CodeInternalError,
			"Failed to approve courier",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"courier_id": courierIDStr,
		"status":     "approved",
		"note":       request.Note,
	}))
}

// RejectCourier 拒绝信使申请（管理员功能）
func (h *CourierHandler) RejectCourier(c *gin.Context) {
	role := middleware.GetUserRole(c)
	if role != "admin" && role != "super_admin" {
		c.JSON(http.StatusForbidden, models.ErrorResponse(
			models.CodeUnauthorized,
			"Admin permission required",
			nil,
		))
		return
	}

	courierIDStr := c.Param("courier_id")
	if courierIDStr == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.CodeParamError,
			"Courier ID is required",
			nil,
		))
		return
	}

	var request struct {
		Note string `json:"note"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.CodeParamError,
			"Invalid request parameters",
			err.Error(),
		))
		return
	}

	err := h.courierService.RejectCourier(courierIDStr, request.Note)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.CodeInternalError,
			"Failed to reject courier",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"courier_id": courierIDStr,
		"status":     "rejected",
		"note":       request.Note,
	}))
}

// RegisterCourierRoutes 注册信使相关路由
func RegisterCourierRoutes(router *gin.RouterGroup, courierService *services.CourierService) {
	handler := NewCourierHandler(courierService)

	router.POST("/apply", handler.ApplyCourier)
	router.GET("/info", handler.GetCourierInfo)
	router.GET("/stats/:courier_id", handler.GetCourierStats)
	router.GET("/stats", handler.GetCourierStats)

	// 管理员路由
	admin := router.Group("/admin")
	{
		admin.PUT("/approve/:courier_id", handler.ApproveCourier)
		admin.PUT("/reject/:courier_id", handler.RejectCourier)
	}
}
