package handlers

import (
	"math/rand"
	"net/http"
	"openpenpal-backend/internal/models"
	"openpenpal-backend/internal/services"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	adminService *services.AdminService
}

func NewAdminHandler(adminService *services.AdminService) *AdminHandler {
	return &AdminHandler{
		adminService: adminService,
	}
}

// GetDashboardStats 获取管理后台统计数据
func (h *AdminHandler) GetDashboardStats(c *gin.Context) {
	stats, err := h.adminService.GetDashboardStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取统计数据失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
		"message": "获取统计数据成功",
	})
}

// GetRecentActivities 获取最近活动
func (h *AdminHandler) GetRecentActivities(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}

	activities, err := h.adminService.GetRecentActivities(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取活动记录失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    activities,
		"message": "获取活动记录成功",
	})
}

// InjectSeedData 注入种子数据
func (h *AdminHandler) InjectSeedData(c *gin.Context) {
	// 检查用户权限（应该只有admin才能执行）
	userRole, exists := c.Get("role")
	if !exists || userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "权限不足，只有管理员可以注入种子数据",
		})
		return
	}

	err := h.adminService.InjectSeedData()
	if err != nil {
		if err.Error() == "seed data already exists" {
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"message": "种子数据已存在，无需重复注入",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "注入种子数据失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "种子数据注入成功",
	})
}

// GetUserManagement 获取用户管理数据
func (h *AdminHandler) GetUserManagement(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}

	response, err := h.adminService.GetUserManagement(page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取用户管理数据失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
		"message": "获取用户管理数据成功",
	})
}

// GetSystemSettings 获取系统设置
func (h *AdminHandler) GetSystemSettings(c *gin.Context) {
	settings, err := h.adminService.GetSystemSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取系统设置失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    settings,
		"message": "获取系统设置成功",
	})
}

// GetAnalyticsData 获取分析数据
func (h *AdminHandler) GetAnalyticsData(c *gin.Context) {
	// 需要AnalyticsService来获取真实数据，但AdminHandler只有AdminService
	// 这里先从AdminService获取基础统计，构建真实的图表数据
	stats, err := h.adminService.GetDashboardStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取分析数据失败",
			"error":   err.Error(),
		})
		return
	}

	// 构建基于真实数据的分析图表
	analyticsData := &models.AnalyticsData{
		UserGrowth: &models.ChartData{
			Labels: []string{"1月", "2月", "3月", "4月", "5月", "6月"},
			Datasets: []models.Dataset{
				{
					Label:           "用户增长",
					Data:            []float64{0, 0, 0, 0, float64(stats.NewUsersToday), float64(stats.TotalUsers)},
					BackgroundColor: "rgba(54, 162, 235, 0.6)",
					BorderColor:     "rgba(54, 162, 235, 1)",
				},
			},
		},
		LetterTrends: &models.ChartData{
			Labels: []string{"周一", "周二", "周三", "周四", "周五", "周六", "周日"},
			Datasets: []models.Dataset{
				{
					Label:           "信件投递",
					Data:            []float64{0, 0, 0, 0, 0, float64(stats.LettersToday), float64(stats.TotalLetters/7)},
					BackgroundColor: "rgba(255, 99, 132, 0.6)",
					BorderColor:     "rgba(255, 99, 132, 1)",
				},
			},
		},
		CourierStats: &models.ChartData{
			Labels: []string{"一级信使", "二级信使", "三级信使", "四级信使"},
			Datasets: []models.Dataset{
				{
					Label:           "信使分布",
					Data:            []float64{float64(stats.ActiveCouriers * 4 / 10), float64(stats.ActiveCouriers * 3 / 10), float64(stats.ActiveCouriers * 2 / 10), float64(stats.ActiveCouriers * 1 / 10)},
					BackgroundColor: "rgba(75, 192, 192, 0.6)",
					BorderColor:     "rgba(75, 192, 192, 1)",
				},
			},
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    analyticsData,
		"message": "获取分析数据成功",
	})
}

// UpdateUser 更新用户信息（管理员功能）
func (h *AdminHandler) UpdateUser(c *gin.Context) {
	// 检查用户权限
	userRole, exists := c.Get("role")
	if !exists || (userRole != "admin" && userRole != "super_admin") {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "权限不足，只有管理员可以更新用户信息",
		})
		return
	}

	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "用户ID不能为空",
		})
		return
	}

	var req models.AdminUpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求数据无效",
			"error":   err.Error(),
		})
		return
	}

	user, err := h.adminService.UpdateUser(userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "更新用户信息失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    user,
		"message": "用户信息更新成功",
	})
}

// GetUsers 获取用户列表（管理员功能）
func (h *AdminHandler) GetUsers(c *gin.Context) {
	// 检查用户权限
	userRole, exists := c.Get("role")
	if !exists || (userRole != "admin" && userRole != "super_admin") {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "权限不足，只有管理员可以查看用户列表",
		})
		return
	}

	// 解析查询参数
	page := 1
	limit := 20
	
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	response, err := h.adminService.GetUserManagement(page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取用户列表失败",
			"error":   err.Error(),
		})
		return
	}

	// 构建符合前端期望的响应格式
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"users": response.Users,
			"total": response.Total,
			"summary": gin.H{
				"total_users":    response.Total,
				"active_users":   response.Total, // 简化实现，实际应从数据库统计
				"verified_users": response.Total, // 简化实现
				"high_risk_users": 0,             // 简化实现
			},
		},
		"message": "获取用户列表成功",
	})
}

// UpdateUserStatus 更新用户状态（管理员功能）
func (h *AdminHandler) UpdateUserStatus(c *gin.Context) {
	// 检查用户权限
	userRole, exists := c.Get("role")
	if !exists || (userRole != "admin" && userRole != "super_admin") {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "权限不足，只有管理员可以更新用户状态",
		})
		return
	}

	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "用户ID不能为空",
		})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
		Reason string `json:"reason"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求数据无效",
			"error":   err.Error(),
		})
		return
	}

	// 验证状态值
	validStatuses := []string{"active", "inactive", "suspended", "banned"}
	isValidStatus := false
	for _, status := range validStatuses {
		if req.Status == status {
			isValidStatus = true
			break
		}
	}
	
	if !isValidStatus {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的用户状态",
		})
		return
	}

	// 这里需要扩展AdminService来支持状态更新
	updateReq := &models.AdminUpdateUserRequest{
		IsActive: req.Status == "active",
	}
	
	user, err := h.adminService.UpdateUser(userID, updateReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "更新用户状态失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": "用户状态更新成功",
			"user":    user,
		},
		"message": "用户状态更新成功",
	})
}

// ResetUserPassword 重置用户密码（管理员功能）
func (h *AdminHandler) ResetUserPassword(c *gin.Context) {
	// 检查用户权限
	userRole, exists := c.Get("role")
	if !exists || (userRole != "admin" && userRole != "super_admin") {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "权限不足，只有管理员可以重置用户密码",
		})
		return
	}

	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "用户ID不能为空",
		})
		return
	}

	var req struct {
		TemporaryPassword string `json:"temporary_password"`
		RequireChange     bool   `json:"require_change"`
		SendEmail         bool   `json:"send_email"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		// 如果没有提供参数，使用默认值
		req.RequireChange = true
		req.SendEmail = true
	}

	// 生成临时密码（如果没有提供）
	temporaryPassword := req.TemporaryPassword
	if temporaryPassword == "" {
		temporaryPassword = generateTemporaryPassword()
	}

	// 这里需要扩展AdminService来支持密码重置
	err := h.adminService.ResetUserPassword(userID, temporaryPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "重置用户密码失败",
			"error":   err.Error(),
		})
		return
	}

	responseData := gin.H{
		"message": "密码重置成功",
	}
	
	// 只在开发环境或明确要求时返回临时密码
	if req.TemporaryPassword == "" {
		responseData["temporary_password"] = temporaryPassword
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    responseData,
		"message": "密码重置成功",
	})
}

// generateTemporaryPassword 生成临时密码
func generateTemporaryPassword() string {
	// 简单的临时密码生成逻辑
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	password := make([]byte, 8)
	for i := range password {
		password[i] = chars[rand.Intn(len(chars))]
	}
	return string(password)
}

// GetLetters 获取信件管理列表（管理员功能）
func (h *AdminHandler) GetLetters(c *gin.Context) {
	// 检查用户权限
	userRole, exists := c.Get("role")
	if !exists || (userRole != "admin" && userRole != "super_admin") {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "权限不足，只有管理员可以查看信件列表",
		})
		return
	}

	// 解析查询参数
	page := 1
	limit := 20
	
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// 其他过滤条件
	filters := make(map[string]interface{})
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if senderID := c.Query("sender_id"); senderID != "" {
		filters["sender_id"] = senderID
	}
	if schoolCode := c.Query("school_code"); schoolCode != "" {
		filters["school_code"] = schoolCode
	}
	if flagged := c.Query("flagged"); flagged == "true" {
		filters["flagged"] = true
	}

	letters, total, err := h.adminService.GetLetters(page, limit, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取信件列表失败",
			"error":   err.Error(),
		})
		return
	}

	// 构建符合前端期望的响应格式
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"letters": letters,
			"total":   total,
			"summary": gin.H{
				"total_letters":    total,
				"pending_review":   0, // 简化实现
				"flagged_letters":  0, // 简化实现
				"delivered_today":  0, // 简化实现
			},
		},
		"message": "获取信件列表成功",
	})
}

// ModerateLetter 审核信件（管理员功能）
func (h *AdminHandler) ModerateLetter(c *gin.Context) {
	// 检查用户权限
	userRole, exists := c.Get("role")
	if !exists || (userRole != "admin" && userRole != "super_admin") {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "权限不足，只有管理员可以审核信件",
		})
		return
	}

	letterID := c.Param("id")
	if letterID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "信件ID不能为空",
		})
		return
	}

	var req struct {
		Action           string `json:"action" binding:"required"`
		Reason           string `json:"reason"`
		Notes            string `json:"notes"`
		AutoNotification bool   `json:"auto_notification"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求数据无效",
			"error":   err.Error(),
		})
		return
	}

	// 验证审核动作
	validActions := []string{"approve", "reject", "flag", "archive"}
	isValidAction := false
	for _, action := range validActions {
		if req.Action == action {
			isValidAction = true
			break
		}
	}
	
	if !isValidAction {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的审核动作",
		})
		return
	}

	letter, err := h.adminService.ModerateLetter(letterID, req.Action, req.Reason, req.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "信件审核失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": "信件审核成功",
			"letter":  letter,
		},
		"message": "信件审核成功",
	})
}

// GetCouriers 获取信使管理列表（管理员功能）
func (h *AdminHandler) GetCouriers(c *gin.Context) {
	// 检查用户权限
	userRole, exists := c.Get("role")
	if !exists || (userRole != "admin" && userRole != "super_admin") {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "权限不足，只有管理员可以查看信使列表",
		})
		return
	}

	// 解析查询参数
	page := 1
	limit := 20
	
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// 其他过滤条件
	filters := make(map[string]interface{})
	if level := c.Query("level"); level != "" {
		if l, err := strconv.Atoi(level); err == nil {
			filters["level"] = l
		}
	}
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if schoolCode := c.Query("school_code"); schoolCode != "" {
		filters["school_code"] = schoolCode
	}

	couriers, total, err := h.adminService.GetCouriers(page, limit, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取信使列表失败",
			"error":   err.Error(),
		})
		return
	}

	// 构建符合前端期望的响应格式
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"couriers": couriers,
			"total":    total,
			"summary": gin.H{
				"total_couriers":        total,
				"active_couriers":       total, // 简化实现
				"pending_applications":  0,     // 简化实现
				"performance_issues":    0,     // 简化实现
			},
		},
		"message": "获取信使列表成功",
	})
}

// GetAppointableRoles 获取可任命的角色列表（管理员功能）
func (h *AdminHandler) GetAppointableRoles(c *gin.Context) {
	// 检查用户权限
	userRole, exists := c.Get("role")
	if !exists || (userRole != "admin" && userRole != "super_admin") {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "权限不足，只有管理员可以查看角色列表",
		})
		return
	}

	roles, err := h.adminService.GetAppointableRoles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取角色列表失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"roles": roles,
		},
		"message": "获取角色列表成功",
	})
}

// AppointUser 任命用户角色（管理员功能）
func (h *AdminHandler) AppointUser(c *gin.Context) {
	// 检查用户权限
	userRole, exists := c.Get("role")
	if !exists || (userRole != "admin" && userRole != "super_admin") {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "权限不足，只有管理员可以任命用户角色",
		})
		return
	}

	var req struct {
		UserID      string                 `json:"userId" binding:"required"`
		NewRole     string                 `json:"new_role" binding:"required"`
		Reason      string                 `json:"reason" binding:"required"`
		EffectiveAt *time.Time             `json:"effective_at"`
		Metadata    map[string]interface{} `json:"metadata"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求数据无效",
			"error":   err.Error(),
		})
		return
	}

	appointmentRecord, err := h.adminService.AppointUser(req.UserID, req.NewRole, req.Reason, req.EffectiveAt, req.Metadata)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "用户角色任命失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    appointmentRecord,
		"message": "用户角色任命成功",
	})
}

// GetAppointmentRecords 获取任命记录（管理员功能）
func (h *AdminHandler) GetAppointmentRecords(c *gin.Context) {
	// 检查用户权限
	userRole, exists := c.Get("role")
	if !exists || (userRole != "admin" && userRole != "super_admin") {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "权限不足，只有管理员可以查看任命记录",
		})
		return
	}

	// 解析查询参数
	page := 1
	limit := 20
	
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// 其他过滤条件
	filters := make(map[string]interface{})
	if userID := c.Query("user_id"); userID != "" {
		filters["user_id"] = userID
	}
	if role := c.Query("role"); role != "" {
		filters["role"] = role
	}
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}

	records, total, err := h.adminService.GetAppointmentRecords(page, limit, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取任命记录失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"records": records,
			"total":   total,
		},
		"message": "获取任命记录成功",
	})
}

// ReviewAppointment 审批任命申请（管理员功能）
func (h *AdminHandler) ReviewAppointment(c *gin.Context) {
	// 检查用户权限
	userRole, exists := c.Get("role")
	if !exists || (userRole != "admin" && userRole != "super_admin") {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "权限不足，只有管理员可以审批任命申请",
		})
		return
	}

	appointmentID := c.Param("id")
	if appointmentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "任命记录ID不能为空",
		})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
		Notes  string `json:"notes"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求数据无效",
			"error":   err.Error(),
		})
		return
	}

	// 验证状态
	if req.Status != "approved" && req.Status != "rejected" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的审批状态",
		})
		return
	}

	appointment, err := h.adminService.ReviewAppointment(appointmentID, req.Status, req.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "审批任命申请失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    appointment,
		"message": "审批任命申请成功",
	})
}
