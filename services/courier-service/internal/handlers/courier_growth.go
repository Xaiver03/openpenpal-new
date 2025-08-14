package handlers

import (
	"courier-service/internal/models"
	"courier-service/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CourierGrowthHandler 信使成长路径与激励系统处理器
type CourierGrowthHandler struct {
	growthService *services.CourierGrowthService
}

// NewCourierGrowthHandler 创建成长路径处理器
func NewCourierGrowthHandler(growthService *services.CourierGrowthService) *CourierGrowthHandler {
	return &CourierGrowthHandler{
		growthService: growthService,
	}
}

// GetGrowthPath 获取成长路径配置
func (h *CourierGrowthHandler) GetGrowthPath(c *gin.Context) {
	courierID := c.GetString("user_id")

	growthPath, err := h.growthService.GetGrowthPath(courierID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.CodeInternalError,
			"Failed to get growth path",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(growthPath))
}

// GetGrowthProgress 获取成长进度
func (h *CourierGrowthHandler) GetGrowthProgress(c *gin.Context) {
	courierID := c.GetString("user_id")

	progress, err := h.growthService.GetGrowthProgress(courierID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.CodeInternalError,
			"Failed to get growth progress",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(progress))
}

// CheckUpgradeRequirements 检查晋升条件
func (h *CourierGrowthHandler) CheckUpgradeRequirements(c *gin.Context) {
	courierID := c.GetString("user_id")

	var request struct {
		TargetLevel models.CourierLevel `json:"target_level" binding:"required,min=2,max=4"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.CodeParamError,
			"Invalid request parameters",
			err.Error(),
		))
		return
	}

	requirements, err := h.growthService.CheckUpgradeRequirements(courierID, request.TargetLevel)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.CodeInternalError,
			"Failed to check upgrade requirements",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"courier_id":      courierID,
		"target_level":    request.TargetLevel,
		"requirements":    requirements.Requirements,
		"can_upgrade":     requirements.CanUpgrade,
		"completion_rate": requirements.CompletionRate,
	}))
}

// GetAvailableIncentives 获取可领取的激励奖励
func (h *CourierGrowthHandler) GetAvailableIncentives(c *gin.Context) {
	courierID := c.GetString("user_id")

	incentives, err := h.growthService.GetAvailableIncentives(courierID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.CodeInternalError,
			"Failed to get available incentives",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"courier_id": courierID,
		"incentives": incentives,
	}))
}

// ClaimIncentive 领取激励奖励
func (h *CourierGrowthHandler) ClaimIncentive(c *gin.Context) {
	courierID := c.GetString("user_id")
	incentiveType := c.Param("type")

	var request struct {
		Reference string `json:"reference"`        // 关联的任务或活动ID
		Amount    *int   `json:"amount,omitempty"` // 指定金额（可选）
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.CodeParamError,
			"Invalid request parameters",
			err.Error(),
		))
		return
	}

	result, err := h.growthService.ClaimIncentive(courierID, models.IncentiveType(incentiveType), request.Reference, request.Amount)
	if err != nil {
		c.JSON(http.StatusConflict, models.ErrorResponse(
			models.CodeConflict,
			"Failed to claim incentive",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(result))
}

// GetRanking 获取区域排行榜数据
func (h *CourierGrowthHandler) GetRanking(c *gin.Context) {
	var query struct {
		ZoneType  string `form:"zone_type"`
		ZoneID    string `form:"zone_id"`
		TimeRange string `form:"time_range"` // daily, weekly, monthly
		Limit     int    `form:"limit"`
		RankingBy string `form:"ranking_by"` // points, tasks, completion_rate
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
	if query.Limit > 100 {
		query.Limit = 100
	}
	if query.TimeRange == "" {
		query.TimeRange = "monthly"
	}
	if query.RankingBy == "" {
		query.RankingBy = "points"
	}

	ranking, err := h.growthService.GetRanking(query.ZoneType, query.ZoneID, query.TimeRange, query.RankingBy, query.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.CodeInternalError,
			"Failed to get ranking",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"zone_type":   query.ZoneType,
		"zone_id":     query.ZoneID,
		"time_range":  query.TimeRange,
		"ranking_by":  query.RankingBy,
		"ranking":     ranking,
		"total_count": len(ranking),
	}))
}

// UpdateTaskStatistics 记录任务完成统计
func (h *CourierGrowthHandler) UpdateTaskStatistics(c *gin.Context) {
	courierID := c.GetString("user_id")

	var request struct {
		TaskID         string   `json:"task_id" binding:"required"`
		Action         string   `json:"action" binding:"required,oneof=accepted completed failed"`
		DeliveryTime   *int     `json:"delivery_time,omitempty"`   // 投递时间(分钟)
		Distance       *float64 `json:"distance,omitempty"`        // 投递距离(km)
		Rating         *float64 `json:"rating,omitempty"`          // 评分
		EarningsAmount *float64 `json:"earnings_amount,omitempty"` // 收入金额
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.CodeParamError,
			"Invalid request parameters",
			err.Error(),
		))
		return
	}

	err := h.growthService.UpdateTaskStatistics(courierID, request.TaskID, request.Action, map[string]interface{}{
		"delivery_time":   request.DeliveryTime,
		"distance":        request.Distance,
		"rating":          request.Rating,
		"earnings_amount": request.EarningsAmount,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.CodeInternalError,
			"Failed to update task statistics",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"courier_id": courierID,
		"task_id":    request.TaskID,
		"action":     request.Action,
		"message":    "Task statistics updated successfully",
	}))
}

// GetEarnedBadges 获取已获得徽章
func (h *CourierGrowthHandler) GetEarnedBadges(c *gin.Context) {
	courierID := c.GetString("user_id")

	badges, err := h.growthService.GetEarnedBadges(courierID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.CodeInternalError,
			"Failed to get earned badges",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"courier_id": courierID,
		"badges":     badges,
		"total":      len(badges),
	}))
}

// AwardBadge 颁发徽章 (系统自动或管理员手动)
func (h *CourierGrowthHandler) AwardBadge(c *gin.Context) {
	// 检查权限，只有高级信使或管理员可以手动颁发徽章
	courierLevel, exists := c.Get("courier_level")
	if !exists || courierLevel.(models.CourierLevel) < models.LevelThree {
		c.JSON(http.StatusForbidden, models.ErrorResponse(
			models.CodeUnauthorized,
			"Insufficient permission to award badges",
			nil,
		))
		return
	}

	awardedBy := c.GetString("user_id")

	var request struct {
		CourierID string `json:"courier_id" binding:"required"`
		BadgeCode string `json:"badge_code" binding:"required"`
		Reason    string `json:"reason" binding:"required"`
		Reference string `json:"reference"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.CodeParamError,
			"Invalid request parameters",
			err.Error(),
		))
		return
	}

	err := h.growthService.AwardBadge(request.CourierID, request.BadgeCode, request.Reason, request.Reference, awardedBy)
	if err != nil {
		c.JSON(http.StatusConflict, models.ErrorResponse(
			models.CodeConflict,
			"Failed to award badge",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"courier_id": request.CourierID,
		"badge_code": request.BadgeCode,
		"awarded_by": awardedBy,
		"reason":     request.Reason,
		"message":    "Badge awarded successfully",
	}))
}

// GetPointsBalance 获取积分余额
func (h *CourierGrowthHandler) GetPointsBalance(c *gin.Context) {
	courierID := c.GetString("user_id")

	pointsBalance, err := h.growthService.GetPointsBalance(courierID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.CodeInternalError,
			"Failed to get points balance",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(pointsBalance))
}

// GetPointsHistory 获取积分交易历史
func (h *CourierGrowthHandler) GetPointsHistory(c *gin.Context) {
	courierID := c.GetString("user_id")

	var query struct {
		Type   string `form:"type"` // earn, spend, refund, expire
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

	transactions, total, err := h.growthService.GetPointsHistory(courierID, query.Type, query.Limit, query.Offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.CodeInternalError,
			"Failed to get points history",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"courier_id":   courierID,
		"transactions": transactions,
		"total":        total,
		"limit":        query.Limit,
		"offset":       query.Offset,
	}))
}

// GetAllBadges 获取所有可获得的徽章
func (h *CourierGrowthHandler) GetAllBadges(c *gin.Context) {
	badges, err := h.growthService.GetAllBadges()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.CodeInternalError,
			"Failed to get badges",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(map[string]interface{}{
		"badges": badges,
		"total":  len(badges),
	}))
}

// GetPerformanceStatistics 获取个人任务统计
func (h *CourierGrowthHandler) GetPerformanceStatistics(c *gin.Context) {
	courierID := c.GetString("user_id")

	var query struct {
		TimeRange string `form:"time_range"` // daily, weekly, monthly, yearly
		StartDate string `form:"start_date"` // YYYY-MM-DD
		EndDate   string `form:"end_date"`   // YYYY-MM-DD
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

	statistics, err := h.growthService.GetPerformanceStatistics(courierID, query.TimeRange, query.StartDate, query.EndDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.CodeInternalError,
			"Failed to get performance statistics",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(statistics))
}

// RegisterCourierGrowthRoutes 注册信使成长路径相关路由
func RegisterCourierGrowthRoutes(router *gin.RouterGroup, growthService *services.CourierGrowthService) {
	handler := NewCourierGrowthHandler(growthService)

	// 公开接口
	router.GET("/growth/badges", handler.GetAllBadges)

	// 需要认证的接口
	authenticated := router.Group("")
	// authenticated.Use(middleware.JWTAuth()) // 假设已经在上级应用了

	// 成长路径相关
	authenticated.GET("/growth/path", handler.GetGrowthPath)
	authenticated.GET("/growth/progress", handler.GetGrowthProgress)
	authenticated.POST("/growth/check-requirements", handler.CheckUpgradeRequirements)

	// 激励系统相关
	authenticated.GET("/incentives/available", handler.GetAvailableIncentives)
	authenticated.POST("/incentives/claim/:type", handler.ClaimIncentive)

	// 统计和排行榜
	authenticated.GET("/statistics/ranking", handler.GetRanking)
	authenticated.PUT("/statistics/task-complete", handler.UpdateTaskStatistics)
	authenticated.GET("/statistics/performance", handler.GetPerformanceStatistics)

	// 徽章系统
	authenticated.GET("/badges/earned", handler.GetEarnedBadges)

	// 积分系统
	authenticated.GET("/points/balance", handler.GetPointsBalance)
	authenticated.GET("/points/history", handler.GetPointsHistory)

	// 需要高级权限的接口
	privileged := authenticated.Group("")
	// privileged.Use(middleware.RequireLevel(models.LevelThree)) // 需要三级以上权限

	// 管理功能
	privileged.POST("/badges/award", handler.AwardBadge)
}
