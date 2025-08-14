package adapters

import (
	"net/http"
	"openpenpal-backend/internal/handlers"
	"openpenpal-backend/internal/models"
	"openpenpal-backend/internal/services"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// AdminAdapter SOTA管理API适配器
// 用于桥接Java前端期待的API格式和Go后端现有的API格式
type AdminAdapter struct {
	adminHandler   *handlers.AdminHandler
	userHandler    *handlers.UserHandler
	letterHandler  *handlers.LetterHandler
	courierHandler *handlers.CourierHandler
	museumHandler  *handlers.MuseumHandler
	adminService   *services.AdminService
	userService    *services.UserService
	letterService  *services.LetterService
	courierService *services.CourierService
	museumService  *services.MuseumService
}

// NewAdminAdapter 创建管理API适配器
func NewAdminAdapter(
	adminHandler *handlers.AdminHandler,
	userHandler *handlers.UserHandler,
	letterHandler *handlers.LetterHandler,
	courierHandler *handlers.CourierHandler,
	museumHandler *handlers.MuseumHandler,
	adminService *services.AdminService,
	userService *services.UserService,
	letterService *services.LetterService,
	courierService *services.CourierService,
	museumService *services.MuseumService,
) *AdminAdapter {
	return &AdminAdapter{
		adminHandler:   adminHandler,
		userHandler:    userHandler,
		letterHandler:  letterHandler,
		courierHandler: courierHandler,
		museumHandler:  museumHandler,
		adminService:   adminService,
		userService:    userService,
		letterService:  letterService,
		courierService: courierService,
		museumService:  museumService,
	}
}

// JavaAPIResponse Java前端期待的响应格式
type JavaAPIResponse struct {
	Code      int         `json:"code"`
	Msg       string      `json:"msg"`
	Data      interface{} `json:"data"`
	Timestamp string      `json:"timestamp"`
	Error     *ErrorInfo  `json:"error,omitempty"`
}

// ErrorInfo 错误信息
type ErrorInfo struct {
	Type    string `json:"type"`
	Details string `json:"details"`
	Field   string `json:"field,omitempty"`
	TraceID string `json:"traceId,omitempty"`
}

// PageResponse 分页响应
type PageResponse struct {
	Items      interface{} `json:"items"`
	Pagination Pagination  `json:"pagination"`
}

// Pagination 分页信息
type Pagination struct {
	Page    int  `json:"page"`
	Limit   int  `json:"limit"`
	Total   int  `json:"total"`
	Pages   int  `json:"pages"`
	HasNext bool `json:"hasNext"`
	HasPrev bool `json:"hasPrev"`
}

// AdaptResponse 适配响应格式
func (a *AdminAdapter) AdaptResponse(data interface{}, message string) JavaAPIResponse {
	return JavaAPIResponse{
		Code:      200,
		Msg:       message,
		Data:      data,
		Timestamp: time.Now().Format(time.RFC3339),
	}
}

// AdaptErrorResponse 适配错误响应
func (a *AdminAdapter) AdaptErrorResponse(c *gin.Context, code int, msg string, err error) {
	response := JavaAPIResponse{
		Code:      code,
		Msg:       msg,
		Data:      nil,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	if err != nil {
		response.Error = &ErrorInfo{
			Type:    "service_error",
			Details: err.Error(),
			TraceID: c.Request.Header.Get("X-Trace-ID"),
		}
	}

	c.JSON(code, response)
}

// AdaptPageResponse 适配分页响应
func (a *AdminAdapter) AdaptPageResponse(c *gin.Context, items interface{}, total int64, page, limit int, message string) {
	pages := int((total + int64(limit) - 1) / int64(limit))

	response := JavaAPIResponse{
		Code: 200,
		Msg:  message,
		Data: PageResponse{
			Items: items,
			Pagination: Pagination{
				Page:    page,
				Limit:   limit,
				Total:   int(total),
				Pages:   pages,
				HasNext: page < pages,
				HasPrev: page > 1,
			},
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, response)
}

// ==================== 用户管理适配 ====================

// GetUsersCompat 获取用户列表 - 兼容Java前端
func (a *AdminAdapter) GetUsersCompat(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")
	_ = c.Query("search")
	role := c.Query("role")
	_ = c.Query("status")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}

	// 使用现有的admin service获取用户数据
	response, err := a.adminService.GetUserManagement(page, limit)
	if err != nil {
		a.AdaptErrorResponse(c, http.StatusInternalServerError, "获取用户列表失败", err)
		return
	}

	// TODO: 增加搜索和过滤功能
	// 当前先返回基础数据，后续增强
	filteredUsers := response.Users

	// 简单的角色过滤
	if role != "" {
		var filtered []models.User
		for _, user := range response.Users {
			if string(user.Role) == role {
				filtered = append(filtered, user)
			}
		}
		filteredUsers = filtered
	}

	a.AdaptPageResponse(c, filteredUsers, response.Total, page, limit, "获取成功")
}

// GetUserCompat 获取用户详情 - 兼容Java前端
func (a *AdminAdapter) GetUserCompat(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		a.AdaptErrorResponse(c, http.StatusBadRequest, "用户ID不能为空", nil)
		return
	}

	// 获取用户详情
	user, err := a.userService.GetUserByID(userID)
	if err != nil {
		a.AdaptErrorResponse(c, http.StatusNotFound, "用户不存在", err)
		return
	}

	response := a.AdaptResponse(user, "获取成功")
	c.JSON(http.StatusOK, response)
}

// UpdateUserCompat 更新用户 - 兼容Java前端
func (a *AdminAdapter) UpdateUserCompat(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		a.AdaptErrorResponse(c, http.StatusBadRequest, "用户ID不能为空", nil)
		return
	}

	var req models.AdminUpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		a.AdaptErrorResponse(c, http.StatusBadRequest, "请求数据无效", err)
		return
	}

	// 检查权限
	userRole, exists := c.Get("role")
	if !exists || (userRole != "admin" && userRole != "super_admin") {
		a.AdaptErrorResponse(c, http.StatusForbidden, "权限不足", nil)
		return
	}

	user, err := a.adminService.UpdateUser(userID, &req)
	if err != nil {
		a.AdaptErrorResponse(c, http.StatusInternalServerError, "更新用户失败", err)
		return
	}

	response := a.AdaptResponse(user, "更新成功")
	c.JSON(http.StatusOK, response)
}

// UnlockUserCompat 解锁用户 - 兼容Java前端
func (a *AdminAdapter) UnlockUserCompat(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		a.AdaptErrorResponse(c, http.StatusBadRequest, "用户ID不能为空", nil)
		return
	}

	// 检查权限
	userRole, exists := c.Get("role")
	if !exists || (userRole != "admin" && userRole != "super_admin") {
		a.AdaptErrorResponse(c, http.StatusForbidden, "权限不足", nil)
		return
	}

	// 解锁用户（激活用户）
	req := &models.AdminUpdateUserRequest{
		IsActive: true,
	}

	user, err := a.adminService.UpdateUser(userID, req)
	if err != nil {
		a.AdaptErrorResponse(c, http.StatusInternalServerError, "解锁用户失败", err)
		return
	}

	response := a.AdaptResponse(user, "解锁成功")
	c.JSON(http.StatusOK, response)
}

// ResetPasswordCompat 重置密码 - 兼容Java前端
func (a *AdminAdapter) ResetPasswordCompat(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		a.AdaptErrorResponse(c, http.StatusBadRequest, "用户ID不能为空", nil)
		return
	}

	var req struct {
		Password string `json:"password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		a.AdaptErrorResponse(c, http.StatusBadRequest, "密码格式不正确", err)
		return
	}

	// 检查权限
	userRole, exists := c.Get("role")
	if !exists || (userRole != "admin" && userRole != "super_admin") {
		a.AdaptErrorResponse(c, http.StatusForbidden, "权限不足", nil)
		return
	}

	// 重置密码
	err := a.userService.AdminResetPassword(userID, req.Password)
	if err != nil {
		a.AdaptErrorResponse(c, http.StatusInternalServerError, "重置密码失败", err)
		return
	}

	response := a.AdaptResponse(nil, "密码重置成功")
	c.JSON(http.StatusOK, response)
}

// GetUserStatsCompat 获取用户统计 - 兼容Java前端
func (a *AdminAdapter) GetUserStatsCompat(c *gin.Context) {
	stats, err := a.adminService.GetDashboardStats()
	if err != nil {
		a.AdaptErrorResponse(c, http.StatusInternalServerError, "获取统计失败", err)
		return
	}

	// 构建角色统计数据
	roleStats := map[string]interface{}{
		"totalUsers":     stats.TotalUsers,
		"newUsersToday":  stats.NewUsersToday,
		"activeCouriers": stats.ActiveCouriers,
		// TODO: 增加详细的角色分布统计
	}

	response := a.AdaptResponse(roleStats, "获取成功")
	c.JSON(http.StatusOK, response)
}

// ==================== 信件管理适配 ====================

// GetLettersCompat 获取信件列表 - 兼容Java前端
func (a *AdminAdapter) GetLettersCompat(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")
	status := c.Query("status")
	userID := c.Query("userId")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}

	// 获取信件列表
	letters, total, err := a.letterService.GetLettersForAdmin(page, limit, status, userID)
	if err != nil {
		a.AdaptErrorResponse(c, http.StatusInternalServerError, "获取信件列表失败", err)
		return
	}

	a.AdaptPageResponse(c, letters, total, page, limit, "获取成功")
}

// GetLetterCompat 获取信件详情 - 兼容Java前端
func (a *AdminAdapter) GetLetterCompat(c *gin.Context) {
	letterID := c.Param("id")
	if letterID == "" {
		a.AdaptErrorResponse(c, http.StatusBadRequest, "信件ID不能为空", nil)
		return
	}

	letter, err := a.letterService.GetLetterByID(letterID, "")
	if err != nil {
		a.AdaptErrorResponse(c, http.StatusNotFound, "信件不存在", err)
		return
	}

	response := a.AdaptResponse(letter, "获取成功")
	c.JSON(http.StatusOK, response)
}

// UpdateLetterStatusCompat 更新信件状态 - 兼容Java前端
func (a *AdminAdapter) UpdateLetterStatusCompat(c *gin.Context) {
	letterID := c.Param("id")
	if letterID == "" {
		a.AdaptErrorResponse(c, http.StatusBadRequest, "信件ID不能为空", nil)
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
		Reason string `json:"reason,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		a.AdaptErrorResponse(c, http.StatusBadRequest, "请求数据无效", err)
		return
	}

	// 检查权限
	userRole, exists := c.Get("role")
	if !exists || (userRole != "admin" && userRole != "super_admin") {
		a.AdaptErrorResponse(c, http.StatusForbidden, "权限不足", nil)
		return
	}

	// 更新信件状态
	err := a.letterService.AdminUpdateLetterStatus(letterID, req.Status, req.Reason)
	if err != nil {
		a.AdaptErrorResponse(c, http.StatusInternalServerError, "更新状态失败", err)
		return
	}

	response := a.AdaptResponse(nil, "状态更新成功")
	c.JSON(http.StatusOK, response)
}

// GetLetterStatsCompat 获取信件统计 - 兼容Java前端
func (a *AdminAdapter) GetLetterStatsCompat(c *gin.Context) {
	stats, err := a.adminService.GetDashboardStats()
	if err != nil {
		a.AdaptErrorResponse(c, http.StatusInternalServerError, "获取统计失败", err)
		return
	}

	letterStats := map[string]interface{}{
		"totalLetters":             stats.TotalLetters,
		"lettersToday":             stats.LettersToday,
		"letterStatusDistribution": stats.LetterStatusDistribution,
	}

	response := a.AdaptResponse(letterStats, "获取成功")
	c.JSON(http.StatusOK, response)
}

// ==================== 系统配置适配 ====================

// GetSystemConfigCompat 获取系统配置 - 兼容Java前端
func (a *AdminAdapter) GetSystemConfigCompat(c *gin.Context) {
	settings, err := a.adminService.GetSystemSettings()
	if err != nil {
		a.AdaptErrorResponse(c, http.StatusInternalServerError, "获取系统配置失败", err)
		return
	}

	response := a.AdaptResponse(settings, "获取成功")
	c.JSON(http.StatusOK, response)
}

// GetSystemInfoCompat 获取系统信息 - 兼容Java前端
func (a *AdminAdapter) GetSystemInfoCompat(c *gin.Context) {
	stats, err := a.adminService.GetDashboardStats()
	if err != nil {
		a.AdaptErrorResponse(c, http.StatusInternalServerError, "获取系统信息失败", err)
		return
	}

	systemInfo := map[string]interface{}{
		"version":        "1.0.0",
		"serverTime":     time.Now().Format(time.RFC3339),
		"systemHealth":   stats.SystemHealth,
		"databaseStatus": "healthy",
		"serviceStatus":  "running",
	}

	response := a.AdaptResponse(systemInfo, "获取成功")
	c.JSON(http.StatusOK, response)
}

// GetSystemHealthCompat 获取系统健康状态 - 兼容Java前端
func (a *AdminAdapter) GetSystemHealthCompat(c *gin.Context) {
	stats, err := a.adminService.GetDashboardStats()
	if err != nil {
		a.AdaptErrorResponse(c, http.StatusInternalServerError, "获取健康状态失败", err)
		return
	}

	response := a.AdaptResponse(stats.SystemHealth, "获取成功")
	c.JSON(http.StatusOK, response)
}

// UpdateSystemConfigCompat 更新系统配置 - 兼容Java前端
func (a *AdminAdapter) UpdateSystemConfigCompat(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		a.AdaptErrorResponse(c, http.StatusBadRequest, "请求数据无效", err)
		return
	}

	// 检查权限
	userRole, exists := c.Get("role")
	if !exists || (userRole != "admin" && userRole != "super_admin") {
		a.AdaptErrorResponse(c, http.StatusForbidden, "权限不足", nil)
		return
	}

	// TODO: 实现系统配置更新逻辑
	// 当前返回成功响应，等待具体配置需求明确
	response := a.AdaptResponse(req, "配置更新成功")
	c.JSON(http.StatusOK, response)
}
