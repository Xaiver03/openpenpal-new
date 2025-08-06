package handlers

import (
	"courier-service/internal/models"
	"courier-service/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// PostalManagementHandler 编号分配权限控制处理器
type PostalManagementHandler struct {
	postalService *services.PostalManagementService
}

// NewPostalManagementHandler 创建编号管理处理器
func NewPostalManagementHandler(postalService *services.PostalManagementService) *PostalManagementHandler {
	return &PostalManagementHandler{
		postalService: postalService,
	}
}

// GetPendingApplications 获取权限范围内待审核编号申请
func (h *PostalManagementHandler) GetPendingApplications(c *gin.Context) {
	courierID := c.GetString("user_id")

	var query struct {
		SchoolID string `form:"school_id"`
		AreaID   string `form:"area_id"`
		Status   string `form:"status"`
		Limit    int    `form:"limit"`
		Offset   int    `form:"offset"`
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
	if query.Status == "" {
		query.Status = "pending"
	}

	applications, total, err := h.postalService.GetPendingApplications(courierID, query.SchoolID, query.AreaID, query.Status, query.Limit, query.Offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.CodeInternalError,
			"Failed to get pending applications",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"applications": applications,
		"total":        total,
		"limit":        query.Limit,
		"offset":       query.Offset,
		"status":       query.Status,
	}))
}

// ApproveApplication 审核编号申请
func (h *PostalManagementHandler) ApproveApplication(c *gin.Context) {
	applicationID := c.Param("application_id")
	reviewerID := c.GetString("user_id")

	var request models.PostalCodeReviewRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.CodeParamError,
			"Invalid request parameters",
			err.Error(),
		))
		return
	}

	// 检查审核权限
	canReview, err := h.postalService.CanReviewApplication(reviewerID, applicationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.CodeInternalError,
			"Failed to check review permission",
			err.Error(),
		))
		return
	}

	if !canReview {
		c.JSON(http.StatusForbidden, models.ErrorResponse(
			models.CodeUnauthorized,
			"Insufficient permission to review this application",
			nil,
		))
		return
	}

	result, err := h.postalService.ReviewApplication(applicationID, request.Action, request.AssignedCode, request.ReviewComment, reviewerID)
	if err != nil {
		c.JSON(http.StatusConflict, models.ErrorResponse(
			models.CodeConflict,
			"Failed to review application",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(result))
}

// RejectApplication 拒绝编号申请
func (h *PostalManagementHandler) RejectApplication(c *gin.Context) {
	applicationID := c.Param("application_id")
	reviewerID := c.GetString("user_id")

	var request struct {
		Reason string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.CodeParamError,
			"Invalid request parameters",
			err.Error(),
		))
		return
	}

	// 检查审核权限
	canReview, err := h.postalService.CanReviewApplication(reviewerID, applicationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.CodeInternalError,
			"Failed to check review permission",
			err.Error(),
		))
		return
	}

	if !canReview {
		c.JSON(http.StatusForbidden, models.ErrorResponse(
			models.CodeUnauthorized,
			"Insufficient permission to review this application",
			nil,
		))
		return
	}

	result, err := h.postalService.ReviewApplication(applicationID, "reject", "", request.Reason, reviewerID)
	if err != nil {
		c.JSON(http.StatusConflict, models.ErrorResponse(
			models.CodeConflict,
			"Failed to reject application",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(result))
}

// GetAssignedCodes 获取权限范围内已分配编号
func (h *PostalManagementHandler) GetAssignedCodes(c *gin.Context) {
	courierID := c.GetString("user_id")

	var query struct {
		SchoolID   string `form:"school_id"`
		AreaID     string `form:"area_id"`
		BuildingID string `form:"building_id"`
		UserID     string `form:"user_id"`
		IsActive   *bool  `form:"is_active"`
		Limit      int    `form:"limit"`
		Offset     int    `form:"offset"`
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
		query.Limit = 50
	}
	if query.Limit > 200 {
		query.Limit = 200
	}

	assignments, total, err := h.postalService.GetAssignedCodes(courierID, map[string]interface{}{
		"school_id":   query.SchoolID,
		"area_id":     query.AreaID,
		"building_id": query.BuildingID,
		"user_id":     query.UserID,
		"is_active":   query.IsActive,
		"limit":       query.Limit,
		"offset":      query.Offset,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.CodeInternalError,
			"Failed to get assigned codes",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"assignments": assignments,
		"total":       total,
		"limit":       query.Limit,
		"offset":      query.Offset,
	}))
}

// BatchAssignCodes 批量分配编号
func (h *PostalManagementHandler) BatchAssignCodes(c *gin.Context) {
	assignerID := c.GetString("user_id")

	var request models.PostalCodeBatchAssignRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.CodeParamError,
			"Invalid request parameters",
			err.Error(),
		))
		return
	}

	// 检查批量分配权限
	canBatchAssign, err := h.postalService.CanBatchAssign(assignerID, request.SchoolID, request.AreaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.CodeInternalError,
			"Failed to check batch assign permission",
			err.Error(),
		))
		return
	}

	if !canBatchAssign {
		c.JSON(http.StatusForbidden, models.ErrorResponse(
			models.CodeUnauthorized,
			"Insufficient permission for batch assignment",
			nil,
		))
		return
	}

	result, err := h.postalService.BatchAssignCodes(assignerID, request.SchoolID, request.AreaID, request.Assignments)
	if err != nil {
		c.JSON(http.StatusConflict, models.ErrorResponse(
			models.CodeConflict,
			"Failed to batch assign codes",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(result))
}

// AssignSingleCode 分配单个编号
func (h *PostalManagementHandler) AssignSingleCode(c *gin.Context) {
	assignerID := c.GetString("user_id")

	var request struct {
		UserID     string `json:"user_id" binding:"required"`
		PostalCode string `json:"postal_code" binding:"required"`
		SchoolID   string `json:"school_id" binding:"required"`
		AreaID     string `json:"area_id" binding:"required"`
		BuildingID string `json:"building_id"`
		RoomNumber string `json:"room_number"`
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
	canAssign, err := h.postalService.CanAssignCode(assignerID, request.SchoolID, request.AreaID, request.BuildingID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.CodeInternalError,
			"Failed to check assign permission",
			err.Error(),
		))
		return
	}

	if !canAssign {
		c.JSON(http.StatusForbidden, models.ErrorResponse(
			models.CodeUnauthorized,
			"Insufficient permission to assign code in this area",
			nil,
		))
		return
	}

	assignment, err := h.postalService.AssignSingleCode(assignerID, request.UserID, request.PostalCode, request.SchoolID, request.AreaID, request.BuildingID, request.RoomNumber)
	if err != nil {
		c.JSON(http.StatusConflict, models.ErrorResponse(
			models.CodeConflict,
			"Failed to assign postal code",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusCreated, models.SuccessResponse(assignment))
}

// DeactivateCode 停用编号
func (h *PostalManagementHandler) DeactivateCode(c *gin.Context) {
	assignmentID := c.Param("assignment_id")
	operatorID := c.GetString("user_id")

	var request struct {
		Reason string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.CodeParamError,
			"Invalid request parameters",
			err.Error(),
		))
		return
	}

	// 检查停用权限
	canDeactivate, err := h.postalService.CanDeactivateCode(operatorID, assignmentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.CodeInternalError,
			"Failed to check deactivate permission",
			err.Error(),
		))
		return
	}

	if !canDeactivate {
		c.JSON(http.StatusForbidden, models.ErrorResponse(
			models.CodeUnauthorized,
			"Insufficient permission to deactivate this code",
			nil,
		))
		return
	}

	err = h.postalService.DeactivateCode(assignmentID, operatorID, request.Reason)
	if err != nil {
		c.JSON(http.StatusConflict, models.ErrorResponse(
			models.CodeConflict,
			"Failed to deactivate postal code",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"assignment_id": assignmentID,
		"deactivated_by": operatorID,
		"reason":        request.Reason,
		"message":       "Postal code deactivated successfully",
	}))
}

// GetPermissionScope 获取当前信使的编号管理权限范围
func (h *PostalManagementHandler) GetPermissionScope(c *gin.Context) {
	courierID := c.GetString("user_id")

	scope, err := h.postalService.GetPermissionScope(courierID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.CodeInternalError,
			"Failed to get permission scope",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(scope))
}

// GetStatistics 获取编号分配统计信息
func (h *PostalManagementHandler) GetStatistics(c *gin.Context) {
	courierID := c.GetString("user_id")

	var query struct {
		SchoolID  string `form:"school_id"`
		AreaID    string `form:"area_id"`
		TimeRange string `form:"time_range"` // daily, weekly, monthly
	}

	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.CodeParamError,
			"Invalid query parameters",
			err.Error(),
		))
		return
	}

	if query.TimeRange == "" {
		query.TimeRange = "monthly"
	}

	statistics, err := h.postalService.GetStatistics(courierID, query.SchoolID, query.AreaID, query.TimeRange)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.CodeInternalError,
			"Failed to get statistics",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(statistics))
}

// SearchCodes 搜索编号
func (h *PostalManagementHandler) SearchCodes(c *gin.Context) {
	courierID := c.GetString("user_id")

	var query struct {
		Code      string `form:"code"`
		UserID    string `form:"user_id"`
		SchoolID  string `form:"school_id"`
		AreaID    string `form:"area_id"`
		IsActive  *bool  `form:"is_active"`
		Limit     int    `form:"limit"`
	}

	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.CodeParamError,
			"Invalid query parameters",
			err.Error(),
		))
		return
	}

	if query.Code == "" && query.UserID == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.CodeParamError,
			"Either code or user_id is required",
			nil,
		))
		return
	}

	if query.Limit <= 0 {
		query.Limit = 20
	}
	if query.Limit > 100 {
		query.Limit = 100
	}

	results, err := h.postalService.SearchCodes(courierID, map[string]interface{}{
		"code":      query.Code,
		"user_id":   query.UserID,
		"school_id": query.SchoolID,
		"area_id":   query.AreaID,
		"is_active": query.IsActive,
		"limit":     query.Limit,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.CodeInternalError,
			"Failed to search codes",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"results": results,
		"total":   len(results),
	}))
}

// ValidateCodeRange 验证编号范围
func (h *PostalManagementHandler) ValidateCodeRange(c *gin.Context) {
	var request struct {
		SchoolID  string   `json:"school_id" binding:"required"`
		AreaID    string   `json:"area_id" binding:"required"`
		CodeRange []string `json:"code_range" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.CodeParamError,
			"Invalid request parameters",
			err.Error(),
		))
		return
	}

	validation, err := h.postalService.ValidateCodeRange(request.SchoolID, request.AreaID, request.CodeRange)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.CodeInternalError,
			"Failed to validate code range",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(validation))
}

// GetApplicationHistory 获取申请历史
func (h *PostalManagementHandler) GetApplicationHistory(c *gin.Context) {
	courierID := c.GetString("user_id")

	var query struct {
		UserID    string `form:"user_id"`
		SchoolID  string `form:"school_id"`
		Status    string `form:"status"`
		StartDate string `form:"start_date"`
		EndDate   string `form:"end_date"`
		Limit     int    `form:"limit"`
		Offset    int    `form:"offset"`
	}

	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.CodeParamError,
			"Invalid query parameters",
			err.Error(),
		))
		return
	}

	if query.Limit <= 0 {
		query.Limit = 20
	}
	if query.Limit > 100 {
		query.Limit = 100
	}

	history, total, err := h.postalService.GetApplicationHistory(courierID, map[string]interface{}{
		"user_id":    query.UserID,
		"school_id":  query.SchoolID,
		"status":     query.Status,
		"start_date": query.StartDate,
		"end_date":   query.EndDate,
		"limit":      query.Limit,
		"offset":     query.Offset,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.CodeInternalError,
			"Failed to get application history",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"history": history,
		"total":   total,
		"limit":   query.Limit,
		"offset":  query.Offset,
	}))
}

// RegisterPostalManagementRoutes 注册编号分配权限控制相关路由
func RegisterPostalManagementRoutes(router *gin.RouterGroup, postalService *services.PostalManagementService) {
	handler := NewPostalManagementHandler(postalService)

	// 需要认证的接口
	authenticated := router.Group("/postal")
	// authenticated.Use(middleware.JWTAuth()) // 假设已经在上级应用了

	// 申请管理
	authenticated.GET("/pending", handler.GetPendingApplications)
	authenticated.PUT("/approve/:application_id", handler.ApproveApplication)
	authenticated.PUT("/reject/:application_id", handler.RejectApplication)
	authenticated.GET("/history", handler.GetApplicationHistory)

	// 编号管理
	authenticated.GET("/assigned", handler.GetAssignedCodes)
	authenticated.POST("/assign", handler.AssignSingleCode)
	authenticated.POST("/batch-assign", handler.BatchAssignCodes)
	authenticated.PUT("/deactivate/:assignment_id", handler.DeactivateCode)

	// 查询和统计
	authenticated.GET("/search", handler.SearchCodes)
	authenticated.GET("/statistics", handler.GetStatistics)
	authenticated.GET("/permission-scope", handler.GetPermissionScope)

	// 工具接口
	authenticated.POST("/validate-range", handler.ValidateCodeRange)
}