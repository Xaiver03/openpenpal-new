package handlers

import (
	"net/http"
	"openpenpal-backend/internal/middleware"
	"openpenpal-backend/internal/models"
	"openpenpal-backend/internal/services"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// SchedulerHandler 任务调度处理器
type SchedulerHandler struct {
	schedulerService *services.SchedulerService
}

// NewSchedulerHandler 创建任务调度处理器实例
func NewSchedulerHandler(schedulerService *services.SchedulerService) *SchedulerHandler {
	return &SchedulerHandler{
		schedulerService: schedulerService,
	}
}

// CreateTask 创建定时任务
// @Summary 创建定时任务
// @Description 创建新的定时任务
// @Tags scheduler
// @Accept json
// @Produce json
// @Param task body models.CreateTaskRequest true "任务创建请求"
// @Success 201 {object} models.ScheduledTask "创建成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/scheduler/tasks [post]
func (h *SchedulerHandler) CreateTask(c *gin.Context) {
	var req models.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request parameters",
			"details": err.Error(),
		})
		return
	}

	// 从JWT中获取用户ID
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	task, err := h.schedulerService.CreateTask(&req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create task",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, task)
}

// GetTasks 获取任务列表
// @Summary 获取任务列表
// @Description 根据查询条件获取任务列表
// @Tags scheduler
// @Accept json
// @Produce json
// @Param task_type query string false "任务类型"
// @Param priority query string false "优先级"
// @Param status query string false "状态"
// @Param is_active query bool false "是否激活"
// @Param start_date query string false "开始日期"
// @Param end_date query string false "结束日期"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param sort_by query string false "排序字段"
// @Param sort_order query string false "排序方向"
// @Success 200 {object} map[string]interface{} "任务列表"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/scheduler/tasks [get]
func (h *SchedulerHandler) GetTasks(c *gin.Context) {
	query := &models.TaskQuery{
		Page:     1,
		PageSize: 20,
	}

	// 解析查询参数
	if taskType := c.Query("task_type"); taskType != "" {
		query.TaskType = models.TaskType(taskType)
	}
	if priority := c.Query("priority"); priority != "" {
		query.Priority = models.TaskPriority(priority)
	}
	if status := c.Query("status"); status != "" {
		query.Status = models.TaskStatus(status)
	}
	if isActiveStr := c.Query("is_active"); isActiveStr != "" {
		if isActive, err := strconv.ParseBool(isActiveStr); err == nil {
			query.IsActive = &isActive
		}
	}
	if startDate := c.Query("start_date"); startDate != "" {
		if parsed, err := time.Parse("2006-01-02", startDate); err == nil {
			query.StartDate = &parsed
		}
	}
	if endDate := c.Query("end_date"); endDate != "" {
		if parsed, err := time.Parse("2006-01-02", endDate); err == nil {
			query.EndDate = &parsed
		}
	}
	if page := c.Query("page"); page != "" {
		if parsed, err := strconv.Atoi(page); err == nil && parsed > 0 {
			query.Page = parsed
		}
	}
	if pageSize := c.Query("page_size"); pageSize != "" {
		if parsed, err := strconv.Atoi(pageSize); err == nil && parsed > 0 && parsed <= 100 {
			query.PageSize = parsed
		}
	}
	query.SortBy = c.Query("sort_by")
	query.SortOrder = c.Query("sort_order")

	tasks, total, err := h.schedulerService.GetTasks(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get tasks",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks": tasks,
		"pagination": gin.H{
			"page":        query.Page,
			"page_size":   query.PageSize,
			"total":       total,
			"total_pages": (total + int64(query.PageSize) - 1) / int64(query.PageSize),
		},
	})
}

// GetTask 获取单个任务详情
// @Summary 获取任务详情
// @Description 根据ID获取任务详情
// @Tags scheduler
// @Accept json
// @Produce json
// @Param id path string true "任务ID"
// @Success 200 {object} models.ScheduledTask "任务详情"
// @Failure 404 {object} map[string]interface{} "任务不存在"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/scheduler/tasks/{id} [get]
func (h *SchedulerHandler) GetTask(c *gin.Context) {
	taskID := c.Param("id")

	task, err := h.schedulerService.GetTaskByID(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Task not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, task)
}

// UpdateTaskStatus 更新任务状态
// @Summary 更新任务状态
// @Description 更新指定任务的状态
// @Tags scheduler
// @Accept json
// @Produce json
// @Param id path string true "任务ID"
// @Param request body map[string]string true "状态更新请求"
// @Success 200 {object} map[string]interface{} "更新成功"
// @Failure 400 {object} map[string]interface{} "请求参数错误"
// @Failure 404 {object} map[string]interface{} "任务不存在"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/scheduler/tasks/{id}/status [put]
func (h *SchedulerHandler) UpdateTaskStatus(c *gin.Context) {
	taskID := c.Param("id")

	var req struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request parameters",
			"details": err.Error(),
		})
		return
	}

	err := h.schedulerService.UpdateTaskStatus(taskID, models.TaskStatus(req.Status))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update task status",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Task status updated successfully",
	})
}

// EnableTask 启用任务
// @Summary 启用任务
// @Description 启用指定的任务
// @Tags scheduler
// @Accept json
// @Produce json
// @Param id path string true "任务ID"
// @Success 200 {object} map[string]interface{} "启用成功"
// @Failure 404 {object} map[string]interface{} "任务不存在"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/scheduler/tasks/{id}/enable [post]
func (h *SchedulerHandler) EnableTask(c *gin.Context) {
	taskID := c.Param("id")

	err := h.schedulerService.EnableTask(taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to enable task",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Task enabled successfully",
	})
}

// DisableTask 禁用任务
// @Summary 禁用任务
// @Description 禁用指定的任务
// @Tags scheduler
// @Accept json
// @Produce json
// @Param id path string true "任务ID"
// @Success 200 {object} map[string]interface{} "禁用成功"
// @Failure 404 {object} map[string]interface{} "任务不存在"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/scheduler/tasks/{id}/disable [post]
func (h *SchedulerHandler) DisableTask(c *gin.Context) {
	taskID := c.Param("id")

	err := h.schedulerService.DisableTask(taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to disable task",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Task disabled successfully",
	})
}

// ExecuteTaskNow 立即执行任务
// @Summary 立即执行任务
// @Description 立即执行指定的任务
// @Tags scheduler
// @Accept json
// @Produce json
// @Param id path string true "任务ID"
// @Success 200 {object} map[string]interface{} "执行成功"
// @Failure 404 {object} map[string]interface{} "任务不存在"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/scheduler/tasks/{id}/execute [post]
func (h *SchedulerHandler) ExecuteTaskNow(c *gin.Context) {
	taskID := c.Param("id")

	err := h.schedulerService.ExecuteTaskNow(taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to execute task",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Task execution started",
	})
}

// DeleteTask 删除任务
// @Summary 删除任务
// @Description 删除指定的任务
// @Tags scheduler
// @Accept json
// @Produce json
// @Param id path string true "任务ID"
// @Success 200 {object} map[string]interface{} "删除成功"
// @Failure 404 {object} map[string]interface{} "任务不存在"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/scheduler/tasks/{id} [delete]
func (h *SchedulerHandler) DeleteTask(c *gin.Context) {
	taskID := c.Param("id")

	err := h.schedulerService.DeleteTask(taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete task",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Task deleted successfully",
	})
}

// GetTaskStats 获取任务统计
// @Summary 获取任务统计
// @Description 获取任务调度系统的统计信息
// @Tags scheduler
// @Accept json
// @Produce json
// @Success 200 {object} models.TaskStats "统计信息"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/scheduler/stats [get]
func (h *SchedulerHandler) GetTaskStats(c *gin.Context) {
	stats, err := h.schedulerService.GetTaskStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get task stats",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetTaskExecutions 获取任务执行记录
// @Summary 获取任务执行记录
// @Description 获取指定任务的执行记录
// @Tags scheduler
// @Accept json
// @Produce json
// @Param id path string true "任务ID"
// @Param limit query int false "限制数量" default(50)
// @Success 200 {object} []models.TaskExecution "执行记录列表"
// @Failure 404 {object} map[string]interface{} "任务不存在"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/scheduler/tasks/{id}/executions [get]
func (h *SchedulerHandler) GetTaskExecutions(c *gin.Context) {
	taskID := c.Param("id")
	limit := 50

	if limitStr := c.Query("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 && parsed <= 1000 {
			limit = parsed
		}
	}

	executions, err := h.schedulerService.GetTaskExecutions(taskID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get task executions",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"executions": executions,
		"count":      len(executions),
	})
}

// CreateDefaultTasks 创建默认任务
// @Summary 创建默认任务
// @Description 创建系统默认的定时任务
// @Tags scheduler
// @Accept json
// @Produce json
// @Success 201 {object} map[string]interface{} "创建成功"
// @Failure 500 {object} map[string]interface{} "服务器内部错误"
// @Router /api/v1/scheduler/tasks/defaults [post]
func (h *SchedulerHandler) CreateDefaultTasks(c *gin.Context) {
	// 从JWT中获取用户ID
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// 默认任务列表
	defaultTasks := []models.CreateTaskRequest{
		{
			Name:           "信件投递提醒",
			Description:    "每小时检查待投递的信件并发送提醒",
			TaskType:       models.TaskTypeLetterDelivery,
			Priority:       models.TaskPriorityNormal,
			CronExpression: "0 0 * * * *", // 每小时执行
			MaxRetries:     3,
			TimeoutSecs:    300,
		},
		{
			Name:           "用户参与度分析",
			Description:    "每日分析用户参与度并发送报告",
			TaskType:       models.TaskTypeUserEngagement,
			Priority:       models.TaskPriorityNormal,
			CronExpression: "0 0 8 * * *", // 每天8点执行
			MaxRetries:     3,
			TimeoutSecs:    600,
		},
		{
			Name:           "通知清理",
			Description:    "每周清理过期通知",
			TaskType:       models.TaskTypeNotificationCleanup,
			Priority:       models.TaskPriorityLow,
			CronExpression: "0 0 2 * * 0", // 每周日凌晨2点执行
			MaxRetries:     2,
			TimeoutSecs:    900,
		},
		{
			Name:           "数据分析更新",
			Description:    "每日更新系统数据分析",
			TaskType:       models.TaskTypeDataAnalytics,
			Priority:       models.TaskPriorityHigh,
			CronExpression: "0 0 6 * * *", // 每天6点执行
			MaxRetries:     3,
			TimeoutSecs:    1800,
		},
		{
			Name:           "统计数据更新",
			Description:    "每15分钟更新统计数据",
			TaskType:       models.TaskTypeStatisticsUpdate,
			Priority:       models.TaskPriorityNormal,
			CronExpression: "0 */15 * * * *", // 每15分钟执行
			MaxRetries:     2,
			TimeoutSecs:    180,
		},
	}

	var createdTasks []string
	createdBy := userID

	for _, taskReq := range defaultTasks {
		task, err := h.schedulerService.CreateTask(&taskReq, createdBy)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to create default tasks",
				"details": err.Error(),
			})
			return
		}
		createdTasks = append(createdTasks, task.ID)
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":       "Default tasks created successfully",
		"created_tasks": createdTasks,
		"count":         len(createdTasks),
	})
}
