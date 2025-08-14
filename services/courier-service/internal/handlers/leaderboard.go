package handlers

import (
	"courier-service/internal/models"
	"courier-service/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// LeaderboardHandler 排行榜处理器
type LeaderboardHandler struct {
	leaderboardService *services.LeaderboardService
}

// NewLeaderboardHandler 创建排行榜处理器
func NewLeaderboardHandler(leaderboardService *services.LeaderboardService) *LeaderboardHandler {
	return &LeaderboardHandler{
		leaderboardService: leaderboardService,
	}
}

// GetSchoolLeaderboard 获取学校排行榜
// GET /api/courier/leaderboard/school
func (h *LeaderboardHandler) GetSchoolLeaderboard(c *gin.Context) {
	req := &models.CourierLeaderboardRequest{
		Type:   "school",
		Limit:  10,
		Offset: 0,
	}

	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			req.Limit = l
		}
	}

	if offset := c.Query("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil {
			req.Offset = o
		}
	}

	if zoneCode := c.Query("zone_code"); zoneCode != "" {
		req.ZoneCode = zoneCode
	}

	response, err := h.leaderboardService.GetSchoolLeaderboard(req)
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

// GetZoneLeaderboard 获取片区排行榜
// GET /api/courier/leaderboard/zone
func (h *LeaderboardHandler) GetZoneLeaderboard(c *gin.Context) {
	req := &models.CourierLeaderboardRequest{
		Type:   "zone",
		Limit:  10,
		Offset: 0,
	}

	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			req.Limit = l
		}
	}

	if offset := c.Query("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil {
			req.Offset = o
		}
	}

	if zoneCode := c.Query("zone_code"); zoneCode != "" {
		req.ZoneCode = zoneCode
	}

	response, err := h.leaderboardService.GetZoneLeaderboard(req)
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

// GetNationalLeaderboard 获取全国排行榜
// GET /api/courier/leaderboard/national
func (h *LeaderboardHandler) GetNationalLeaderboard(c *gin.Context) {
	req := &models.CourierLeaderboardRequest{
		Type:   "national",
		Limit:  10,
		Offset: 0,
	}

	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			req.Limit = l
		}
	}

	if offset := c.Query("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil {
			req.Offset = o
		}
	}

	response, err := h.leaderboardService.GetNationalLeaderboard(req)
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

// GetPointsHistory 获取积分历史
// GET /api/courier/points-history
func (h *LeaderboardHandler) GetPointsHistory(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
		return
	}

	var courier models.Courier
	if err := h.leaderboardService.GetDB().Where("user_id = ?", userID).First(&courier).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "信使身份未找到"})
		return
	}

	limit := 20
	offset := 0

	if l := c.Query("limit"); l != "" {
		if parsedLimit, err := strconv.Atoi(l); err == nil {
			limit = parsedLimit
		}
	}

	if o := c.Query("offset"); o != "" {
		if parsedOffset, err := strconv.Atoi(o); err == nil {
			offset = parsedOffset
		}
	}

	history, total, err := h.leaderboardService.GetPointsHistory(courier.ID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Code:    0,
		Message: "获取成功",
		Data: gin.H{
			"history": history,
			"total":   total,
			"page":    offset/limit + 1,
			"limit":   limit,
		},
	})
}

// GetMyRank 获取我的排名
// GET /api/courier/my-rank
func (h *LeaderboardHandler) GetMyRank(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
		return
	}

	var courier models.Courier
	if err := h.leaderboardService.GetDB().Where("user_id = ?", userID).First(&courier).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "信使身份未找到"})
		return
	}

	ranking, err := h.leaderboardService.GetCourierRank(courier.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Code:    0,
		Message: "获取成功",
		Data:    ranking,
	})
}

// AddPointsManual 手动添加积分 (管理员功能)
// POST /api/courier/admin/points
func (h *LeaderboardHandler) AddPointsManual(c *gin.Context) {
	var req struct {
		CourierID   string `json:"courier_id" binding:"required"`
		Points      int    `json:"points" binding:"required"`
		Type        string `json:"type" binding:"required"`
		Description string `json:"description" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	err := h.leaderboardService.AddPoints(req.CourierID, req.Points, req.Type, req.Description, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Code:    0,
		Message: "积分添加成功",
		Data:    nil,
	})
}

// UpdateRankings 更新排行榜 (管理员功能)
// POST /api/courier/admin/update-rankings
func (h *LeaderboardHandler) UpdateRankings(c *gin.Context) {
	err := h.leaderboardService.UpdateRankings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Code:    0,
		Message: "排行榜更新成功",
		Data:    nil,
	})
}

// GetLeaderboardStats 获取排行榜统计信息
// GET /api/courier/leaderboard/stats
func (h *LeaderboardHandler) GetLeaderboardStats(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
		return
	}

	var courier models.Courier
	if err := h.leaderboardService.GetDB().Where("user_id = ?", userID).First(&courier).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "信使身份未找到"})
		return
	}

	// 获取个人排名
	ranking, err := h.leaderboardService.GetCourierRank(courier.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 获取积分历史统计
	var totalPoints int
	h.leaderboardService.GetDB().Model(&models.CourierPointsHistory{}).
		Where("courier_id = ?", courier.ID).
		Select("SUM(points)").
		Scan(&totalPoints)

	var thisMonthPoints int
	h.leaderboardService.GetDB().Model(&models.CourierPointsHistory{}).
		Where("courier_id = ? AND created_at >= DATE_TRUNC('month', CURRENT_DATE)", courier.ID).
		Select("SUM(points)").
		Scan(&thisMonthPoints)

	stats := gin.H{
		"current_points":    courier.Points,
		"total_earned":      totalPoints,
		"this_month_earned": thisMonthPoints,
		"school_rank":       ranking.SchoolRank,
		"zone_rank":         ranking.ZoneRank,
		"national_rank":     ranking.NationalRank,
		"total_tasks":       ranking.TotalTasks,
		"success_rate":      ranking.SuccessRate,
		"level":             courier.Level,
		"level_name":        courier.GetLevelName(),
		"zone_type":         courier.ZoneType,
		"zone_name":         courier.GetZoneTypeName(),
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Code:    0,
		Message: "获取成功",
		Data:    stats,
	})
}

// RegisterLeaderboardRoutes 注册排行榜路由
func RegisterLeaderboardRoutes(router *gin.RouterGroup, leaderboardService *services.LeaderboardService) {
	handler := NewLeaderboardHandler(leaderboardService)

	// 排行榜路由组
	leaderboard := router.Group("/leaderboard")
	{
		leaderboard.GET("/school", handler.GetSchoolLeaderboard)
		leaderboard.GET("/zone", handler.GetZoneLeaderboard)
		leaderboard.GET("/national", handler.GetNationalLeaderboard)
		leaderboard.GET("/stats", handler.GetLeaderboardStats)
	}

	// 积分相关路由
	router.GET("/points-history", handler.GetPointsHistory)
	router.GET("/my-rank", handler.GetMyRank)

	// 管理员路由
	admin := router.Group("/admin")
	{
		admin.POST("/points", handler.AddPointsManual)
		admin.POST("/update-rankings", handler.UpdateRankings)
	}
}
