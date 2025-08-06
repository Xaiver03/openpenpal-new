package handlers

import (
	"courier-service/internal/middleware"
	"courier-service/internal/models"
	"courier-service/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// TaskHandler 任务处理器
type TaskHandler struct {
	taskService  *services.TaskService
	queueService *services.QueueService
}

// NewTaskHandler 创建任务处理器
func NewTaskHandler(taskService *services.TaskService, queueService *services.QueueService) *TaskHandler {
	return &TaskHandler{
		taskService:  taskService,
		queueService: queueService,
	}
}

// GetTasks 获取任务列表
func (h *TaskHandler) GetTasks(c *gin.Context) {
	var query models.TaskQuery
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
		query.Limit = 10
	}
	if query.Limit > 100 {
		query.Limit = 100
	}

	tasks, total, err := h.taskService.GetAvailableTasks(&query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.CodeInternalError,
			"Failed to get tasks",
			err.Error(),
		))
		return
	}

	response := map[string]interface{}{
		"tasks": tasks,
		"total": total,
		"limit": query.Limit,
		"offset": query.Offset,
	}

	c.JSON(http.StatusOK, models.SuccessResponse(response))
}

// AcceptTask 接受任务
func (h *TaskHandler) AcceptTask(c *gin.Context) {
	taskID := c.Param("task_id")
	courierID := middleware.GetUserID(c)

	if courierID == "" {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(
			models.CodeUnauthorized,
			"User ID not found",
			nil,
		))
		return
	}

	var request models.TaskAcceptRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.CodeParamError,
			"Invalid request parameters",
			err.Error(),
		))
		return
	}

	task, err := h.taskService.AcceptTask(taskID, courierID, &request)
	if err != nil {
		c.JSON(http.StatusConflict, models.ErrorResponse(
			models.CodeConflict,
			"Failed to accept task",
			err.Error(),
		))
		return
	}

	response := map[string]interface{}{
		"task_id":     task.TaskID,
		"courier_id":  courierID,
		"accepted_at": task.AcceptedAt,
		"deadline":    task.Deadline,
	}

	c.JSON(http.StatusOK, models.SuccessResponse(response))
}

// GetTaskDetail 获取任务详情
func (h *TaskHandler) GetTaskDetail(c *gin.Context) {
	taskID := c.Param("task_id")

	task, err := h.taskService.GetTaskByID(taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(
			models.CodeNotFound,
			"Task not found",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(task))
}

// CreateTask 创建任务（管理员功能）
func (h *TaskHandler) CreateTask(c *gin.Context) {
	role := middleware.GetUserRole(c)
	if role != "admin" && role != "super_admin" {
		c.JSON(http.StatusForbidden, models.ErrorResponse(
			models.CodeUnauthorized,
			"Admin permission required",
			nil,
		))
		return
	}

	var request struct {
		LetterID         string `json:"letter_id" binding:"required"`
		PickupLocation   string `json:"pickup_location" binding:"required"`
		DeliveryLocation string `json:"delivery_location" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			models.CodeParamError,
			"Invalid request parameters",
			err.Error(),
		))
		return
	}

	task, err := h.taskService.CreateTask(
		request.LetterID,
		request.PickupLocation,
		request.DeliveryLocation,
		h.queueService,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			models.CodeInternalError,
			"Failed to create task",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusCreated, models.SuccessResponse(task))
}

// RegisterTaskRoutes 注册任务相关路由
func RegisterTaskRoutes(router *gin.RouterGroup, taskService *services.TaskService, queueService *services.QueueService) {
	handler := NewTaskHandler(taskService, queueService)

	router.GET("/tasks", handler.GetTasks)
	router.GET("/tasks/:task_id", handler.GetTaskDetail)
	router.PUT("/tasks/:task_id/accept", handler.AcceptTask)

	// 管理员路由
	admin := router.Group("/admin")
	{
		admin.POST("/tasks", handler.CreateTask)
	}
}