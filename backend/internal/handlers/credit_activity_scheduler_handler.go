package handlers

import (
	"fmt"
	"openpenpal-backend/internal/models"
	"openpenpal-backend/internal/services"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"shared/pkg/response"
)

// CreditActivitySchedulerHandler 积分活动调度器处理器
type CreditActivitySchedulerHandler struct {
	scheduler *services.CreditActivityScheduler
}

// NewCreditActivitySchedulerHandler 创建调度器处理器实例
func NewCreditActivitySchedulerHandler(scheduler *services.CreditActivityScheduler) *CreditActivitySchedulerHandler {
	return &CreditActivitySchedulerHandler{
		scheduler: scheduler,
	}
}

// ===================== 调度器控制 API =====================

// StartScheduler 启动调度器（管理员）
func (h *CreditActivitySchedulerHandler) StartScheduler(c *gin.Context) {
	resp := response.NewGinResponse()

	// 验证管理员权限
	user := c.MustGet("user").(*models.User)
	if user.Role != "admin" && user.Role != "super_admin" {
		resp.Unauthorized(c, "只有管理员可以控制调度器")
		return
	}

	if err := h.scheduler.Start(); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	resp.SuccessWithMessage(c, "调度器启动成功", nil)
}

// StopScheduler 停止调度器（管理员）
func (h *CreditActivitySchedulerHandler) StopScheduler(c *gin.Context) {
	resp := response.NewGinResponse()

	// 验证管理员权限
	user := c.MustGet("user").(*models.User)
	if user.Role != "admin" && user.Role != "super_admin" {
		resp.Unauthorized(c, "只有管理员可以控制调度器")
		return
	}

	if err := h.scheduler.Stop(); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	resp.SuccessWithMessage(c, "调度器停止成功", nil)
}

// GetSchedulerStatus 获取调度器状态
func (h *CreditActivitySchedulerHandler) GetSchedulerStatus(c *gin.Context) {
	resp := response.NewGinResponse()

	// 验证管理员权限
	user := c.MustGet("user").(*models.User)
	if user.Role != "admin" && user.Role != "super_admin" {
		resp.Unauthorized(c, "只有管理员可以查看调度器状态")
		return
	}

	status := h.scheduler.GetSchedulerStatus()
	resp.Success(c, status)
}

// ===================== 任务调度管理 API =====================

// ScheduleActivity 安排活动执行（管理员）
func (h *CreditActivitySchedulerHandler) ScheduleActivity(c *gin.Context) {
	resp := response.NewGinResponse()

	// 验证管理员权限
	user := c.MustGet("user").(*models.User)
	if user.Role != "admin" && user.Role != "super_admin" {
		resp.Unauthorized(c, "只有管理员可以安排活动执行")
		return
	}

	var req struct {
		ActivityID       string                 `json:"activity_id" binding:"required"`
		ScheduledTime    string                 `json:"scheduled_time" binding:"required"`
		ExecutionDetails map[string]interface{} `json:"execution_details"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	// 解析活动ID
	activityID, err := uuid.Parse(req.ActivityID)
	if err != nil {
		resp.BadRequest(c, "无效的活动ID")
		return
	}

	// 解析调度时间
	scheduledTime, err := time.Parse(time.RFC3339, req.ScheduledTime)
	if err != nil {
		resp.BadRequest(c, "无效的调度时间格式，请使用 RFC3339 格式")
		return
	}

	// 检查调度时间不能是过去时间
	if scheduledTime.Before(time.Now()) {
		resp.BadRequest(c, "调度时间不能是过去时间")
		return
	}

	schedule, err := h.scheduler.ScheduleActivity(activityID, scheduledTime, req.ExecutionDetails)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.CreatedWithMessage(c, "活动调度安排成功", schedule)
}

// GetScheduledTasks 获取调度任务列表（管理员）
func (h *CreditActivitySchedulerHandler) GetScheduledTasks(c *gin.Context) {
	resp := response.NewGinResponse()

	// 验证管理员权限
	user := c.MustGet("user").(*models.User)
	if user.Role != "admin" && user.Role != "super_admin" {
		resp.Unauthorized(c, "只有管理员可以查看调度任务")
		return
	}

	// 解析查询参数
	status := c.Query("status")
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	tasks, err := h.scheduler.GetScheduledTasks(status, limit)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.Success(c, gin.H{
		"tasks": tasks,
		"count": len(tasks),
	})
}

// CancelScheduledTask 取消调度任务（管理员）
func (h *CreditActivitySchedulerHandler) CancelScheduledTask(c *gin.Context) {
	resp := response.NewGinResponse()

	// 验证管理员权限
	user := c.MustGet("user").(*models.User)
	if user.Role != "admin" && user.Role != "super_admin" {
		resp.Unauthorized(c, "只有管理员可以取消调度任务")
		return
	}

	scheduleIDStr := c.Param("id")
	scheduleID, err := uuid.Parse(scheduleIDStr)
	if err != nil {
		resp.BadRequest(c, "无效的任务ID")
		return
	}

	if err := h.scheduler.CancelScheduledTask(scheduleID); err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.SuccessWithMessage(c, "调度任务取消成功", nil)
}

// ===================== 批量操作 API =====================

// ScheduleRecurringActivities 安排重复活动（管理员）
func (h *CreditActivitySchedulerHandler) ScheduleRecurringActivities(c *gin.Context) {
	resp := response.NewGinResponse()

	// 验证管理员权限
	user := c.MustGet("user").(*models.User)
	if user.Role != "admin" && user.Role != "super_admin" {
		resp.Unauthorized(c, "只有管理员可以安排重复活动")
		return
	}

	var req struct {
		ActivityIDs      []string               `json:"activity_ids" binding:"required"`
		StartTime        string                 `json:"start_time" binding:"required"`
		RepeatPattern    string                 `json:"repeat_pattern" binding:"required"` // daily, weekly, monthly
		RepeatInterval   int                    `json:"repeat_interval" binding:"required"`
		EndTime          string                 `json:"end_time"`
		ExecutionDetails map[string]interface{} `json:"execution_details"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	// 解析开始时间
	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		resp.BadRequest(c, "无效的开始时间格式")
		return
	}

	// 解析结束时间（可选）
	if req.EndTime != "" {
		_, err := time.Parse(time.RFC3339, req.EndTime)
		if err != nil {
			resp.BadRequest(c, "无效的结束时间格式")
			return
		}
		// 注意: 未来可以在活动模型中使用endTime来设置重复结束日期
	}

	// 验证重复模式
	validPatterns := map[string]bool{
		"daily":   true,
		"weekly":  true,
		"monthly": true,
		"yearly":  true,
	}
	if !validPatterns[req.RepeatPattern] {
		resp.BadRequest(c, "无效的重复模式")
		return
	}

	if req.RepeatInterval <= 0 {
		resp.BadRequest(c, "重复间隔必须大于0")
		return
	}

	successCount := 0
	var schedules []*models.CreditActivitySchedule

	for _, activityIDStr := range req.ActivityIDs {
		activityID, err := uuid.Parse(activityIDStr)
		if err != nil {
			continue
		}

		// 为每个活动创建调度
		schedule, err := h.scheduler.ScheduleActivity(activityID, startTime, req.ExecutionDetails)
		if err != nil {
			continue
		}

		schedules = append(schedules, schedule)
		successCount++
	}

	resp.Success(c, gin.H{
		"message":        fmt.Sprintf("成功安排 %d 个活动的重复执行", successCount),
		"scheduled_count": successCount,
		"total_count":    len(req.ActivityIDs),
		"schedules":      schedules,
	})
}

// ProcessImmediateExecution 立即执行活动（管理员）
func (h *CreditActivitySchedulerHandler) ProcessImmediateExecution(c *gin.Context) {
	resp := response.NewGinResponse()

	// 验证管理员权限
	user := c.MustGet("user").(*models.User)
	if user.Role != "admin" && user.Role != "super_admin" {
		resp.Unauthorized(c, "只有管理员可以立即执行活动")
		return
	}

	var req struct {
		ActivityID       string                 `json:"activity_id" binding:"required"`
		ExecutionDetails map[string]interface{} `json:"execution_details"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	// 解析活动ID
	activityID, err := uuid.Parse(req.ActivityID)
	if err != nil {
		resp.BadRequest(c, "无效的活动ID")
		return
	}

	// 立即安排执行（调度时间为当前时间）
	schedule, err := h.scheduler.ScheduleActivity(activityID, time.Now(), req.ExecutionDetails)
	if err != nil {
		resp.InternalServerError(c, err.Error())
		return
	}

	resp.CreatedWithMessage(c, "活动已安排立即执行", schedule)
}

// ===================== 调度统计 API =====================

// GetSchedulingStatistics 获取调度统计（管理员）
func (h *CreditActivitySchedulerHandler) GetSchedulingStatistics(c *gin.Context) {
	resp := response.NewGinResponse()

	// 验证管理员权限
	user := c.MustGet("user").(*models.User)
	if user.Role != "admin" && user.Role != "super_admin" {
		resp.Unauthorized(c, "只有管理员可以查看调度统计")
		return
	}

	status := h.scheduler.GetSchedulerStatus()
	
	// 添加额外的统计信息
	now := time.Now()
	today := now.Truncate(24 * time.Hour)
	thisWeek := today.AddDate(0, 0, -int(today.Weekday()))

	// 这里可以添加更多统计查询，比如：
	// - 今日执行的任务数
	// - 本周执行的任务数
	// - 平均执行时间
	// - 失败率等

	resp.Success(c, gin.H{
		"scheduler_status": status,
		"statistics": gin.H{
			"current_time": now,
			"today":        today,
			"this_week":    thisWeek,
		},
	})
}