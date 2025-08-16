package handlers

import (
	"openpenpal-backend/internal/models"
	"openpenpal-backend/internal/services"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"openpenpal-backend/internal/pkg/response"
)

// CreditActivityHandler 积分活动处理器
type CreditActivityHandler struct {
	creditActivityService *services.CreditActivityService
	creditService         *services.CreditService
}

// NewCreditActivityHandler 创建积分活动处理器实例
func NewCreditActivityHandler(creditActivityService *services.CreditActivityService, creditService *services.CreditService) *CreditActivityHandler {
	return &CreditActivityHandler{
		creditActivityService: creditActivityService,
		creditService:         creditService,
	}
}

// ===================== 活动管理 API =====================

// GetActivities 获取活动列表
func (h *CreditActivityHandler) GetActivities(c *gin.Context) {
	resp := response.NewGinResponse()
	params := make(map[string]interface{})

	// 解析查询参数
	if page, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil && page > 0 {
		params["page"] = page
	}
	if limit, err := strconv.Atoi(c.DefaultQuery("limit", "20")); err == nil && limit > 0 && limit <= 100 {
		params["limit"] = limit
	}
	if status := c.Query("status"); status != "" {
		params["status"] = models.CreditActivityStatus(status)
	}
	if activityType := c.Query("activity_type"); activityType != "" {
		params["activity_type"] = models.CreditActivityType(activityType)
	}
	if targetType := c.Query("target_type"); targetType != "" {
		params["target_type"] = models.CreditActivityTargetType(targetType)
	}
	if keyword := c.Query("keyword"); keyword != "" {
		params["keyword"] = keyword
	}
	if sortBy := c.Query("sort_by"); sortBy != "" {
		params["sort_by"] = sortBy
	}
	if activeOnly := c.Query("active_only"); activeOnly == "true" {
		params["active_only"] = true
	}

	activities, total, err := h.creditActivityService.GetActivities(params)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	page := 1
	limit := 20
	if p, ok := params["page"]; ok {
		page = p.(int)
	}
	if l, ok := params["limit"]; ok {
		limit = l.(int)
	}
	hasNext := int64(page*limit) < total

	resp.Success(c, gin.H{
		"items":     activities,
		"total":     total,
		"page":      page,
		"page_size": limit,
		"has_next":  hasNext,
	})
}

// GetActivity 获取活动详情
func (h *CreditActivityHandler) GetActivity(c *gin.Context) {
	resp := response.NewGinResponse()

	activityID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		resp.BadRequest(c, "无效的活动ID")
		return
	}

	activity, err := h.creditActivityService.GetActivityByID(activityID)
	if err != nil {
		resp.NotFound(c, "活动不存在")
		return
	}

	// 获取活动统计
	stats, err := h.creditActivityService.GetActivityStatistics(activityID)
	if err != nil {
		stats = nil
	}

	resp.Success(c, gin.H{
		"activity":   activity,
		"statistics": stats,
	})
}

// CreateActivity 创建活动（管理员）
func (h *CreditActivityHandler) CreateActivity(c *gin.Context) {
	resp := response.NewGinResponse()

	// 验证管理员权限
	user := c.MustGet("user").(*models.User)
	if user.Role != "admin" && user.Role != "super_admin" {
		resp.Unauthorized(c, "只有管理员可以创建活动")
		return
	}

	var activity models.CreditActivity
	if err := c.ShouldBindJSON(&activity); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	// 设置创建人
	activity.CreatedBy = user.ID
	activity.UpdatedBy = user.ID

	if err := h.creditActivityService.CreateActivity(&activity); err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.CreatedWithMessage(c, "活动创建成功", activity)
}

// UpdateActivity 更新活动（管理员）
func (h *CreditActivityHandler) UpdateActivity(c *gin.Context) {
	resp := response.NewGinResponse()

	// 验证管理员权限
	user := c.MustGet("user").(*models.User)
	if user.Role != "admin" && user.Role != "super_admin" {
		resp.Unauthorized(c, "只有管理员可以更新活动")
		return
	}

	activityID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		resp.BadRequest(c, "无效的活动ID")
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	// 设置更新人
	updates["updated_by"] = user.ID

	if err := h.creditActivityService.UpdateActivity(activityID, updates); err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.SuccessWithMessage(c, "活动更新成功", nil)
}

// DeleteActivity 删除活动（管理员）
func (h *CreditActivityHandler) DeleteActivity(c *gin.Context) {
	resp := response.NewGinResponse()

	// 验证管理员权限
	user := c.MustGet("user").(*models.User)
	if user.Role != "admin" && user.Role != "super_admin" {
		resp.Unauthorized(c, "只有管理员可以删除活动")
		return
	}

	activityID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		resp.BadRequest(c, "无效的活动ID")
		return
	}

	if err := h.creditActivityService.DeleteActivity(activityID); err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.SuccessWithMessage(c, "活动删除成功", nil)
}

// ===================== 活动状态管理 API =====================

// StartActivity 启动活动（管理员）
func (h *CreditActivityHandler) StartActivity(c *gin.Context) {
	resp := response.NewGinResponse()

	// 验证管理员权限
	user := c.MustGet("user").(*models.User)
	if user.Role != "admin" && user.Role != "super_admin" {
		resp.Unauthorized(c, "只有管理员可以启动活动")
		return
	}

	activityID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		resp.BadRequest(c, "无效的活动ID")
		return
	}

	if err := h.creditActivityService.StartActivity(activityID); err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.SuccessWithMessage(c, "活动启动成功", nil)
}

// PauseActivity 暂停活动（管理员）
func (h *CreditActivityHandler) PauseActivity(c *gin.Context) {
	resp := response.NewGinResponse()

	// 验证管理员权限
	user := c.MustGet("user").(*models.User)
	if user.Role != "admin" && user.Role != "super_admin" {
		resp.Unauthorized(c, "只有管理员可以暂停活动")
		return
	}

	activityID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		resp.BadRequest(c, "无效的活动ID")
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	if err := h.creditActivityService.PauseActivity(activityID, req.Reason); err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.SuccessWithMessage(c, "活动暂停成功", nil)
}

// ResumeActivity 恢复活动（管理员）
func (h *CreditActivityHandler) ResumeActivity(c *gin.Context) {
	resp := response.NewGinResponse()

	// 验证管理员权限
	user := c.MustGet("user").(*models.User)
	if user.Role != "admin" && user.Role != "super_admin" {
		resp.Unauthorized(c, "只有管理员可以恢复活动")
		return
	}

	activityID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		resp.BadRequest(c, "无效的活动ID")
		return
	}

	if err := h.creditActivityService.ResumeActivity(activityID); err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.SuccessWithMessage(c, "活动恢复成功", nil)
}

// CompleteActivity 结束活动（管理员）
func (h *CreditActivityHandler) CompleteActivity(c *gin.Context) {
	resp := response.NewGinResponse()

	// 验证管理员权限
	user := c.MustGet("user").(*models.User)
	if user.Role != "admin" && user.Role != "super_admin" {
		resp.Unauthorized(c, "只有管理员可以结束活动")
		return
	}

	activityID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		resp.BadRequest(c, "无效的活动ID")
		return
	}

	if err := h.creditActivityService.CompleteActivity(activityID); err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.SuccessWithMessage(c, "活动结束成功", nil)
}

// ===================== 用户参与 API =====================

// ParticipateActivity 参与活动
func (h *CreditActivityHandler) ParticipateActivity(c *gin.Context) {
	resp := response.NewGinResponse()
	user := c.MustGet("user").(*models.User)

	activityID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		resp.BadRequest(c, "无效的活动ID")
		return
	}

	var triggerData map[string]interface{}
	if err := c.ShouldBindJSON(&triggerData); err != nil {
		triggerData = make(map[string]interface{})
	}

	participation, err := h.creditActivityService.ParticipateActivity(user.ID, activityID, triggerData)
	if err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	resp.SuccessWithMessage(c, "参与活动成功", participation)
}

// GetUserParticipations 获取用户参与记录
func (h *CreditActivityHandler) GetUserParticipations(c *gin.Context) {
	resp := response.NewGinResponse()
	user := c.MustGet("user").(*models.User)

	params := make(map[string]interface{})

	// 解析查询参数
	if page, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil && page > 0 {
		params["page"] = page
	}
	if limit, err := strconv.Atoi(c.DefaultQuery("limit", "20")); err == nil && limit > 0 && limit <= 100 {
		params["limit"] = limit
	}
	if activityID := c.Query("activity_id"); activityID != "" {
		if id, err := uuid.Parse(activityID); err == nil {
			params["activity_id"] = id
		}
	}
	if completed := c.Query("completed"); completed == "true" {
		params["completed"] = true
	} else if completed == "false" {
		params["completed"] = false
	}

	participations, total, err := h.creditActivityService.GetUserParticipations(user.ID, params)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	page := 1
	limit := 20
	if p, ok := params["page"]; ok {
		page = p.(int)
	}
	if l, ok := params["limit"]; ok {
		limit = l.(int)
	}
	hasNext := int64(page*limit) < total

	resp.Success(c, gin.H{
		"items":     participations,
		"total":     total,
		"page":      page,
		"page_size": limit,
		"has_next":  hasNext,
	})
}

// GetActiveActivities 获取进行中的活动
func (h *CreditActivityHandler) GetActiveActivities(c *gin.Context) {
	resp := response.NewGinResponse()

	params := map[string]interface{}{
		"status":      models.CreditActivityStatusActive,
		"active_only": true,
		"page":        1,
		"limit":       20,
	}

	// 解析查询参数
	if page, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil && page > 0 {
		params["page"] = page
	}
	if limit, err := strconv.Atoi(c.DefaultQuery("limit", "20")); err == nil && limit > 0 && limit <= 100 {
		params["limit"] = limit
	}
	if activityType := c.Query("activity_type"); activityType != "" {
		params["activity_type"] = models.CreditActivityType(activityType)
	}

	activities, total, err := h.creditActivityService.GetActivities(params)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	// 如果用户已登录，获取用户的参与状态
	if userInterface, exists := c.Get("user"); exists {
		if user, ok := userInterface.(*models.User); ok {
			// 为每个活动添加用户参与状态
			for i := range activities {
				participation, _, _ := h.creditActivityService.GetUserParticipations(user.ID, map[string]interface{}{
					"activity_id": activities[i].ID,
					"limit":       1,
				})
				if len(participation) > 0 {
					// 将参与信息添加到活动中（这里简化处理，实际可能需要在模型中添加字段）
					// activities[i].UserParticipation = participation[0]
				}
			}
		}
	}

	page := params["page"].(int)
	limit := params["limit"].(int)
	hasNext := int64(page*limit) < total

	resp.Success(c, gin.H{
		"items":     activities,
		"total":     total,
		"page":      page,
		"page_size": limit,
		"has_next":  hasNext,
	})
}

// ===================== 统计 API =====================

// GetActivityStatistics 获取活动统计（管理员）
func (h *CreditActivityHandler) GetActivityStatistics(c *gin.Context) {
	resp := response.NewGinResponse()

	// 验证管理员权限
	user := c.MustGet("user").(*models.User)
	if user.Role != "admin" && user.Role != "super_admin" {
		resp.Unauthorized(c, "只有管理员可以查看活动统计")
		return
	}

	activityID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		resp.BadRequest(c, "无效的活动ID")
		return
	}

	stats, err := h.creditActivityService.GetActivityStatistics(activityID)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.Success(c, stats)
}

// GetAllActivitiesStatistics 获取所有活动统计概览（管理员）
func (h *CreditActivityHandler) GetAllActivitiesStatistics(c *gin.Context) {
	resp := response.NewGinResponse()

	// 验证管理员权限
	user := c.MustGet("user").(*models.User)
	if user.Role != "admin" && user.Role != "super_admin" {
		resp.Unauthorized(c, "只有管理员可以查看统计概览")
		return
	}

	stats, err := h.creditActivityService.GetAllActivitiesStatistics()
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.Success(c, stats)
}

// ===================== 活动模板 API =====================

// GetActivityTemplates 获取活动模板
func (h *CreditActivityHandler) GetActivityTemplates(c *gin.Context) {
	resp := response.NewGinResponse()

	category := c.Query("category")
	isPublic := c.Query("is_public") != "false"

	templates, err := h.creditActivityService.GetActivityTemplates(category, isPublic)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.Success(c, templates)
}

// CreateActivityTemplate 创建活动模板（管理员）
func (h *CreditActivityHandler) CreateActivityTemplate(c *gin.Context) {
	resp := response.NewGinResponse()

	// 验证管理员权限
	user := c.MustGet("user").(*models.User)
	if user.Role != "admin" && user.Role != "super_admin" {
		resp.Unauthorized(c, "只有管理员可以创建活动模板")
		return
	}

	var template models.CreditActivityTemplate
	if err := c.ShouldBindJSON(&template); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	template.CreatedBy = user.ID

	if err := h.creditActivityService.CreateActivityTemplate(&template); err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.CreatedWithMessage(c, "活动模板创建成功", template)
}

// CreateActivityFromTemplate 从模板创建活动（管理员）
func (h *CreditActivityHandler) CreateActivityFromTemplate(c *gin.Context) {
	resp := response.NewGinResponse()

	// 验证管理员权限
	user := c.MustGet("user").(*models.User)
	if user.Role != "admin" && user.Role != "super_admin" {
		resp.Unauthorized(c, "只有管理员可以从模板创建活动")
		return
	}

	templateID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		resp.BadRequest(c, "无效的模板ID")
		return
	}

	var customData map[string]interface{}
	if err := c.ShouldBindJSON(&customData); err != nil {
		customData = make(map[string]interface{})
	}

	// 添加创建人信息
	customData["created_by"] = user.ID
	customData["updated_by"] = user.ID

	activity, err := h.creditActivityService.CreateActivityFromTemplate(templateID, customData)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.CreatedWithMessage(c, "从模板创建活动成功", activity)
}

// ===================== 活动触发 API =====================

// TriggerActivity 触发活动事件（内部使用）
func (h *CreditActivityHandler) TriggerActivity(c *gin.Context) {
	resp := response.NewGinResponse()
	user := c.MustGet("user").(*models.User)

	var req struct {
		TriggerType models.CreditActivityTriggerType `json:"trigger_type" binding:"required"`
		TriggerData map[string]interface{}           `json:"trigger_data"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	// 评估活动触发
	err := h.creditActivityService.EvaluateActivityTrigger(user.ID, req.TriggerType, req.TriggerData)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.SuccessWithMessage(c, "活动触发评估完成", nil)
}

// ProcessScheduledActivities 处理定时活动（系统调用）
func (h *CreditActivityHandler) ProcessScheduledActivities(c *gin.Context) {
	resp := response.NewGinResponse()

	// 验证是否为系统调用或管理员
	if authHeader := c.GetHeader("X-System-Token"); authHeader == "" {
		user, exists := c.Get("user")
		if !exists {
			resp.Unauthorized(c, "需要系统权限")
			return
		}
		if u, ok := user.(*models.User); !ok || (u.Role != "admin" && u.Role != "super_admin") {
			resp.Unauthorized(c, "需要管理员权限")
			return
		}
	}

	err := h.creditActivityService.ProcessScheduledActivities()
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.SuccessWithMessage(c, "定时活动处理完成", gin.H{
		"processed_at": time.Now(),
	})
}