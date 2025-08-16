package handlers

import (
	"strconv"

	"openpenpal-backend/internal/middleware"
	"openpenpal-backend/internal/pkg/response"
	"openpenpal-backend/internal/services"

	"github.com/gin-gonic/gin"
)

type CreditHandler struct {
	creditService *services.CreditService
}

func NewCreditHandler(creditService *services.CreditService) *CreditHandler {
	return &CreditHandler{
		creditService: creditService,
	}
}

// GetUserCredit 获取用户积分信息
func (h *CreditHandler) GetUserCredit(c *gin.Context) {
	resp := response.NewGinResponse()

	userID, exists := middleware.GetUserID(c)
	if !exists {
		resp.Unauthorized(c, "User not authenticated")
		return
	}

	credit, err := h.creditService.GetUserCreditInfo(userID)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.Success(c, credit)
}

// GetCreditHistory 获取用户积分历史
func (h *CreditHandler) GetCreditHistory(c *gin.Context) {
	resp := response.NewGinResponse()

	userID, exists := middleware.GetUserID(c)
	if !exists {
		resp.Unauthorized(c, "User not authenticated")
		return
	}

	// 解析分页参数
	page := 1
	limit := 20

	if pageStr := c.Query("page"); pageStr != "" {
		if parsed, err := strconv.Atoi(pageStr); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	offset := (page - 1) * limit

	transactions, total, err := h.creditService.GetCreditHistory(userID, limit, offset)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.Success(c, gin.H{
		"transactions": transactions,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// GetLeaderboard 获取积分排行榜
func (h *CreditHandler) GetLeaderboard(c *gin.Context) {
	resp := response.NewGinResponse()

	// 解析限制参数
	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	leaderboard, err := h.creditService.GetLeaderboard(limit)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.Success(c, gin.H{
		"leaderboard": leaderboard,
		"count":       len(leaderboard),
	})
}

// GetCreditRules 获取积分规则
func (h *CreditHandler) GetCreditRules(c *gin.Context) {
	resp := response.NewGinResponse()

	rules := gin.H{
		"letter_rewards": gin.H{
			"letter_created":   services.PointsLetterCreated,
			"letter_generated": services.PointsLetterGenerated,
			"letter_delivered": services.PointsLetterDelivered,
			"letter_read":      services.PointsLetterRead,
			"receive_letter":   services.PointsReceiveLetter,
		},
		"envelope_rewards": gin.H{
			"envelope_purchase": services.PointsEnvelopePurchase,
			"envelope_binding":  services.PointsEnvelopeBinding,
		},
		"museum_rewards": gin.H{
			"museum_submit":   services.PointsMuseumSubmit,
			"museum_approved": services.PointsMuseumApproved,
			"museum_liked":    services.PointsMuseumLiked,
		},
		"level_requirements": services.PointsLevelUp,
	}

	resp.Success(c, rules)
}

// GetUserLevel 获取用户等级信息
func (h *CreditHandler) GetUserLevel(c *gin.Context) {
	resp := response.NewGinResponse()

	userID, exists := middleware.GetUserID(c)
	if !exists {
		resp.Unauthorized(c, "User not authenticated")
		return
	}

	credit, err := h.creditService.GetUserCreditInfo(userID)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	// 计算到下一级的进度
	nextLevel := credit.Level + 1
	var nextLevelPoints int
	var progress float64

	if nextLevel < len(services.PointsLevelUp) {
		nextLevelPoints = services.PointsLevelUp[nextLevel]
		currentLevelPoints := services.PointsLevelUp[credit.Level]
		progress = float64(credit.Total-currentLevelPoints) / float64(nextLevelPoints-currentLevelPoints) * 100
	} else {
		// 已经达到最高等级
		nextLevelPoints = credit.Total
		progress = 100
	}

	levelInfo := gin.H{
		"current_level":     credit.Level,
		"current_points":    credit.Total,
		"next_level":        nextLevel,
		"next_level_points": nextLevelPoints,
		"progress_percent":  progress,
		"available_points":  credit.Available,
		"used_points":       credit.Used,
		"earned_points":     credit.Earned,
	}

	resp.Success(c, levelInfo)
}

// AdminAddPoints 管理员添加积分 (仅管理员可用)
func (h *CreditHandler) AdminAddPoints(c *gin.Context) {
	resp := response.NewGinResponse()

	var req struct {
		UserID      string `json:"user_id" binding:"required"`
		Points      int    `json:"points" binding:"required,min=1"`
		Description string `json:"description" binding:"required"`
		Reference   string `json:"reference"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	err := h.creditService.AddPoints(req.UserID, req.Points, req.Description, req.Reference)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.OK(c, "Points added successfully")
}

// AdminSpendPoints 管理员扣除积分 (仅管理员可用)
func (h *CreditHandler) AdminSpendPoints(c *gin.Context) {
	resp := response.NewGinResponse()

	var req struct {
		UserID      string `json:"user_id" binding:"required"`
		Points      int    `json:"points" binding:"required,min=1"`
		Description string `json:"description" binding:"required"`
		Reference   string `json:"reference"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	err := h.creditService.SpendPoints(req.UserID, req.Points, req.Description, req.Reference)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.OK(c, "Points deducted successfully")
}

// AdminGetUserCredit 管理员获取指定用户积分信息
func (h *CreditHandler) AdminGetUserCredit(c *gin.Context) {
	resp := response.NewGinResponse()

	userID := c.Param("user_id")
	if userID == "" {
		resp.BadRequest(c, "User ID is required")
		return
	}

	credit, err := h.creditService.GetUserCreditInfo(userID)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	// 获取该用户的积分历史（最近10条）
	transactions, _, err := h.creditService.GetCreditHistory(userID, 10, 0)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.Success(c, gin.H{
		"credit":              credit,
		"recent_transactions": transactions,
	})
}

// GetCreditStats 获取积分统计信息
func (h *CreditHandler) GetCreditStats(c *gin.Context) {
	resp := response.NewGinResponse()

	userID, exists := middleware.GetUserID(c)
	if !exists {
		resp.Unauthorized(c, "User not authenticated")
		return
	}

	credit, err := h.creditService.GetUserCreditInfo(userID)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	// 获取排行榜排名
	leaderboard, err := h.creditService.GetLeaderboard(1000) // 获取足够多的数据来确定排名
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	rank := 0
	for i, entry := range leaderboard {
		if entry.UserID == userID {
			rank = i + 1
			break
		}
	}

	stats := gin.H{
		"total_points":     credit.Total,
		"available_points": credit.Available,
		"used_points":      credit.Used,
		"earned_points":    credit.Earned,
		"level":            credit.Level,
		"rank":             rank,
		"total_users":      len(leaderboard),
	}

	resp.Success(c, stats)
}
