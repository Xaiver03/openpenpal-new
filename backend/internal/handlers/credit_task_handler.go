package handlers

import (
	"strconv"
	"strings"

	"openpenpal-backend/internal/middleware"
	"openpenpal-backend/internal/models"
	"openpenpal-backend/internal/services"

	"github.com/gin-gonic/gin"
	"shared/pkg/response"
)

type CreditTaskHandler struct {
	taskService   *services.CreditTaskService
	creditService *services.CreditService
}

func NewCreditTaskHandler(taskService *services.CreditTaskService, creditService *services.CreditService) *CreditTaskHandler {
	return &CreditTaskHandler{
		taskService:   taskService,
		creditService: creditService,
	}
}

// GetUserTasks 获取用户积分任务列表
func (h *CreditTaskHandler) GetUserTasks(c *gin.Context) {
	resp := response.NewGinResponse()

	userID, exists := middleware.GetUserID(c)
	if !exists {
		resp.Unauthorized(c, "User not authenticated")
		return
	}

	// 解析参数
	status := models.CreditTaskStatus(c.Query("status"))
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

	tasks, total, err := h.taskService.GetUserTasks(userID, status, limit, offset)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.Success(c, gin.H{
		"tasks": tasks,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// CreateTask 创建积分任务 (管理员功能)
func (h *CreditTaskHandler) CreateTask(c *gin.Context) {
	resp := response.NewGinResponse()

	var req struct {
		TaskType    models.CreditTaskType `json:"task_type" binding:"required"`
		UserID      string                `json:"user_id" binding:"required"`
		Points      int                   `json:"points" binding:"required,min=1"`
		Description string                `json:"description" binding:"required"`
		Reference   string                `json:"reference"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	task, err := h.taskService.CreateTask(req.TaskType, req.UserID, req.Points, req.Description, req.Reference)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.Success(c, task)
}

// ExecuteTask 手动执行积分任务 (管理员功能)
func (h *CreditTaskHandler) ExecuteTask(c *gin.Context) {
	resp := response.NewGinResponse()

	taskID := c.Param("task_id")
	if taskID == "" {
		resp.BadRequest(c, "Task ID is required")
		return
	}

	err := h.taskService.ExecuteTask(taskID)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.OK(c, "Task executed successfully")
}

// GetTaskStatistics 获取积分任务统计 (管理员功能)
func (h *CreditTaskHandler) GetTaskStatistics(c *gin.Context) {
	resp := response.NewGinResponse()

	timeRange := c.Query("range")
	if timeRange == "" {
		timeRange = "today"
	}

	stats, err := h.taskService.GetTaskStatistics(timeRange)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.Success(c, gin.H{
		"statistics": stats,
		"time_range": timeRange,
	})
}

// CreateBatchTasks 批量创建积分任务 (管理员功能)
func (h *CreditTaskHandler) CreateBatchTasks(c *gin.Context) {
	resp := response.NewGinResponse()

	var req struct {
		BatchName   string                `json:"batch_name" binding:"required"`
		TaskType    models.CreditTaskType `json:"task_type" binding:"required"`
		UserIDs     []string              `json:"user_ids" binding:"required,min=1"`
		Points      int                   `json:"points" binding:"required,min=1"`
		Description string                `json:"description" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	batch, err := h.taskService.CreateBatchTasks(req.BatchName, req.TaskType, req.UserIDs, req.Points, req.Description)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.Success(c, batch)
}

// =============== 快速触发接口 ===============

// TriggerLetterCreatedReward 触发写信奖励
func (h *CreditTaskHandler) TriggerLetterCreatedReward(c *gin.Context) {
	resp := response.NewGinResponse()

	userID, exists := middleware.GetUserID(c)
	if !exists {
		resp.Unauthorized(c, "User not authenticated")
		return
	}

	letterID := c.Param("letter_id")
	if letterID == "" {
		resp.BadRequest(c, "Letter ID is required")
		return
	}

	err := h.taskService.TriggerLetterCreatedReward(userID, letterID)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.OK(c, "Letter creation reward triggered")
}

// TriggerPublicLetterLikeReward 触发公开信点赞奖励
func (h *CreditTaskHandler) TriggerPublicLetterLikeReward(c *gin.Context) {
	resp := response.NewGinResponse()

	var req struct {
		LetterID   string `json:"letter_id" binding:"required"`
		LikedByID  string `json:"liked_by_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	// 这里需要获取信件作者的用户ID
	// 实际实现中应该查询信件表获取作者ID
	err := h.taskService.TriggerPublicLetterLikeReward(req.LikedByID, req.LetterID)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.OK(c, "Public letter like reward triggered")
}

// TriggerAIInteractionReward 触发AI互动奖励
func (h *CreditTaskHandler) TriggerAIInteractionReward(c *gin.Context) {
	resp := response.NewGinResponse()

	userID, exists := middleware.GetUserID(c)
	if !exists {
		resp.Unauthorized(c, "User not authenticated")
		return
	}

	var req struct {
		SessionID string `json:"session_id" binding:"required"`
		Rating    int    `json:"rating" binding:"required,min=1,max=5"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	err := h.taskService.TriggerAIInteractionReward(userID, req.SessionID)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.OK(c, "AI interaction reward triggered")
}

// TriggerCourierDeliveryReward 触发信使送达奖励
func (h *CreditTaskHandler) TriggerCourierDeliveryReward(c *gin.Context) {
	resp := response.NewGinResponse()

	userID, exists := middleware.GetUserID(c)
	if !exists {
		resp.Unauthorized(c, "User not authenticated")
		return
	}

	taskID := c.Param("task_id")
	if taskID == "" {
		resp.BadRequest(c, "Task ID is required")
		return
	}

	err := h.taskService.TriggerCourierDeliveryReward(userID, taskID)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.OK(c, "Courier delivery reward triggered")
}

// GetCreditTaskRules 获取积分任务规则配置
func (h *CreditTaskHandler) GetCreditTaskRules(c *gin.Context) {
	resp := response.NewGinResponse()

	rules := gin.H{
		"task_types": gin.H{
			"letter_created":    gin.H{"points": services.PointsLetterCreated, "description": "成功写信并绑定条码"},
			"letter_generated":  gin.H{"points": services.PointsLetterGenerated, "description": "生成信件编号"},
			"letter_delivered":  gin.H{"points": services.PointsLetterDelivered, "description": "信件送达"},
			"letter_read":       gin.H{"points": services.PointsLetterRead, "description": "信件被阅读"},
			"receive_letter":    gin.H{"points": services.PointsReceiveLetter, "description": "收到信件"},
			"public_like":       gin.H{"points": services.PointsPublicLetterLike, "description": "公开信被点赞"},
			"writing_challenge": gin.H{"points": services.PointsWritingChallenge, "description": "参与写作挑战"},
			"ai_interaction":    gin.H{"points": services.PointsAIInteraction, "description": "AI互动评价"},
			"courier_first":     gin.H{"points": services.PointsCourierFirstTask, "description": "信使首次任务"},
			"courier_delivery":  gin.H{"points": services.PointsCourierDelivery, "description": "信使送达信件"},
			"museum_submit":     gin.H{"points": services.PointsMuseumSubmit, "description": "博物馆提交"},
			"museum_approved":   gin.H{"points": services.PointsMuseumApproved, "description": "博物馆审核通过"},
			"museum_liked":      gin.H{"points": services.PointsMuseumLiked, "description": "博物馆点赞"},
			"opcode_approval":   gin.H{"points": services.PointsOPCodeApproval, "description": "OP Code审核成功"},
			"community_badge":   gin.H{"points": services.PointsCommunityBadge, "description": "社区贡献徽章"},
		},
		"daily_limits": gin.H{
			"letter_created":    3,  // 每日上限3封
			"receive_letter":    5,  // 每日上限5封
			"public_like":       20, // 每封信上限20赞
			"writing_challenge": 1,  // 每周限一次
			"ai_interaction":    3,  // 每日限3次
		},
		"task_statuses": []string{
			string(models.TaskStatusPending),
			string(models.TaskStatusScheduled),
			string(models.TaskStatusExecuting),
			string(models.TaskStatusCompleted),
			string(models.TaskStatusFailed),
			string(models.TaskStatusCancelled),
			string(models.TaskStatusSkipped),
		},
	}

	resp.Success(c, rules)
}

// GetUserCreditSummary 获取用户积分摘要（包含任务统计）
func (h *CreditTaskHandler) GetUserCreditSummary(c *gin.Context) {
	resp := response.NewGinResponse()

	userID, exists := middleware.GetUserID(c)
	if !exists {
		resp.Unauthorized(c, "User not authenticated")
		return
	}

	// 获取用户积分信息
	credit, err := h.creditService.GetUserCreditInfo(userID)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	// 获取用户任务统计
	pendingTasks, _, err := h.taskService.GetUserTasks(userID, models.TaskStatusPending, 100, 0)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	completedTasks, _, err := h.taskService.GetUserTasks(userID, models.TaskStatusCompleted, 100, 0)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	// 获取每日积分统计
	dailyStats, err := h.creditService.GetDailyStats(userID)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	// 按任务类型统计积分来源
	taskTypeStats := make(map[string]int)
	for _, task := range completedTasks {
		taskTypeKey := strings.ReplaceAll(string(task.TaskType), "_", " ")
		taskTypeStats[taskTypeKey] += task.Points
	}

	summary := gin.H{
		"credit_info": credit,
		"daily_stats": dailyStats,
		"task_summary": gin.H{
			"pending_count":   len(pendingTasks),
			"completed_count": len(completedTasks),
			"type_breakdown":  taskTypeStats,
		},
		"recent_pending_tasks":   limitSlice(pendingTasks, 5),
		"recent_completed_tasks": limitSlice(completedTasks, 10),
	}

	resp.Success(c, summary)
}

// limitSlice 限制切片长度的辅助函数
func limitSlice(tasks []models.CreditTask, limit int) []models.CreditTask {
	if len(tasks) <= limit {
		return tasks
	}
	return tasks[:limit]
}