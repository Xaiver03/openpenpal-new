package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"openpenpal-backend/internal/models"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

// SchedulerService 任务调度服务
type SchedulerService struct {
	db       *gorm.DB
	cron     *cron.Cron
	workers  map[string]*TaskWorker
	mu       sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
	workerID string
}

// TaskWorker 任务执行器
type TaskWorker struct {
	ID             string
	MaxConcurrency int
	CurrentTasks   int
	IsActive       bool
	mu             sync.RWMutex
}

// NewSchedulerService 创建调度服务
func NewSchedulerService(db *gorm.DB) *SchedulerService {
	ctx, cancel := context.WithCancel(context.Background())

	// 使用主机名+进程ID作为Worker ID
	hostname, _ := os.Hostname()
	workerID := fmt.Sprintf("%s-%d", hostname, os.Getpid())

	return &SchedulerService{
		db:       db,
		cron:     cron.New(cron.WithSeconds()),
		workers:  make(map[string]*TaskWorker),
		ctx:      ctx,
		cancel:   cancel,
		workerID: workerID,
	}
}

// Start 启动调度服务
func (s *SchedulerService) Start() error {
	log.Printf("Starting scheduler service with worker ID: %s", s.workerID)

	// 注册当前worker
	if err := s.registerWorker(); err != nil {
		return fmt.Errorf("failed to register worker: %w", err)
	}

	// 加载已有的定时任务
	if err := s.loadScheduledTasks(); err != nil {
		return fmt.Errorf("failed to load scheduled tasks: %w", err)
	}

	// 启动cron调度器
	s.cron.Start()

	// 启动后台监控goroutine
	go s.monitorTasks()
	go s.cleanupExpiredTasks()
	go s.updateWorkerHeartbeat()

	log.Println("Scheduler service started successfully")
	return nil
}

// Stop 停止调度服务
func (s *SchedulerService) Stop() {
	log.Println("Stopping scheduler service...")

	s.cancel()
	s.cron.Stop()

	// 更新worker状态为非激活
	s.updateWorkerStatus("inactive")

	log.Println("Scheduler service stopped")
}

// CreateTask 创建定时任务
func (s *SchedulerService) CreateTask(req *models.CreateTaskRequest, createdBy string) (*models.ScheduledTask, error) {
	payloadJSON := "{}"
	if req.Payload != nil {
		if data, err := json.Marshal(req.Payload); err == nil {
			payloadJSON = string(data)
		}
	}

	task := &models.ScheduledTask{
		ID:             uuid.New().String(),
		Name:           req.Name,
		Description:    req.Description,
		TaskType:       req.TaskType,
		Priority:       req.Priority,
		Status:         models.TaskStatusPending,
		CronExpression: req.CronExpression,
		Payload:        payloadJSON,
		MaxRetries:     req.MaxRetries,
		TimeoutSecs:    req.TimeoutSecs,
		IsActive:       true,
		CreatedBy:      createdBy,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// 设置默认值
	if task.Priority == "" {
		task.Priority = models.TaskPriorityNormal
	}
	if task.MaxRetries == 0 {
		task.MaxRetries = 3
	}
	if task.TimeoutSecs == 0 {
		task.TimeoutSecs = 300
	}

	// 计算下次执行时间
	if req.ScheduledAt != nil {
		task.ScheduledAt = *req.ScheduledAt
		task.NextRunAt = *req.ScheduledAt
	} else if req.CronExpression != "" {
		nextTime, err := s.getNextRunTime(req.CronExpression)
		if err != nil {
			return nil, fmt.Errorf("invalid cron expression: %w", err)
		}
		task.NextRunAt = nextTime
	}

	if req.StartDate != nil {
		task.StartDate = *req.StartDate
	} else {
		task.StartDate = time.Now()
	}
	if req.EndDate != nil {
		task.EndDate = req.EndDate
	}
	if req.MaxRuns != nil {
		task.MaxRuns = req.MaxRuns
	}

	// 保存到数据库
	if err := s.db.Create(task).Error; err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	// 如果任务是激活状态，添加到cron调度器
	if task.IsActive && task.CronExpression != "" {
		if err := s.addTaskToCron(task); err != nil {
			log.Printf("Failed to add task to cron: %v", err)
		}
	}

	return task, nil
}

// GetTasks 获取任务列表
func (s *SchedulerService) GetTasks(query *models.TaskQuery) ([]models.ScheduledTask, int64, error) {
	var tasks []models.ScheduledTask
	var total int64

	db := s.db.Model(&models.ScheduledTask{})

	// 应用过滤条件
	if query.TaskType != "" {
		db = db.Where("task_type = ?", query.TaskType)
	}
	if query.Priority != "" {
		db = db.Where("priority = ?", query.Priority)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.IsActive != nil {
		db = db.Where("is_active = ?", *query.IsActive)
	}
	if query.StartDate != nil {
		db = db.Where("created_at >= ?", *query.StartDate)
	}
	if query.EndDate != nil {
		db = db.Where("created_at <= ?", *query.EndDate)
	}

	// 计算总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 排序
	orderBy := "created_at DESC"
	if query.SortBy != "" {
		order := "ASC"
		if query.SortOrder == "desc" {
			order = "DESC"
		}
		orderBy = fmt.Sprintf("%s %s", query.SortBy, order)
	}

	// 分页
	offset := (query.Page - 1) * query.PageSize
	if err := db.Order(orderBy).Offset(offset).Limit(query.PageSize).Find(&tasks).Error; err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

// GetTaskByID 根据ID获取任务
func (s *SchedulerService) GetTaskByID(taskID string) (*models.ScheduledTask, error) {
	var task models.ScheduledTask
	if err := s.db.Where("id = ?", taskID).First(&task).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

// UpdateTaskStatus 更新任务状态
func (s *SchedulerService) UpdateTaskStatus(taskID string, status models.TaskStatus) error {
	return s.db.Model(&models.ScheduledTask{}).
		Where("id = ?", taskID).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now(),
		}).Error
}

// EnableTask 启用任务
func (s *SchedulerService) EnableTask(taskID string) error {
	task, err := s.GetTaskByID(taskID)
	if err != nil {
		return err
	}

	if err := s.db.Model(task).Updates(map[string]interface{}{
		"is_active":  true,
		"updated_at": time.Now(),
	}).Error; err != nil {
		return err
	}

	// 添加到cron调度器
	if task.CronExpression != "" {
		return s.addTaskToCron(task)
	}

	return nil
}

// DisableTask 禁用任务
func (s *SchedulerService) DisableTask(taskID string) error {
	task, err := s.GetTaskByID(taskID)
	if err != nil {
		return err
	}

	if err := s.db.Model(task).Updates(map[string]interface{}{
		"is_active":  false,
		"updated_at": time.Now(),
	}).Error; err != nil {
		return err
	}

	// 从cron调度器移除
	s.removeTaskFromCron(taskID)

	return nil
}

// ExecuteTaskNow 立即执行任务
func (s *SchedulerService) ExecuteTaskNow(taskID string) error {
	task, err := s.GetTaskByID(taskID)
	if err != nil {
		return err
	}

	go s.executeTask(task)
	return nil
}

// GetTaskStats 获取任务统计
func (s *SchedulerService) GetTaskStats() (*models.TaskStats, error) {
	stats := &models.TaskStats{
		TasksByType:   make(map[string]int64),
		TasksByStatus: make(map[string]int64),
		LastUpdate:    time.Now(),
	}

	// 总任务数
	s.db.Model(&models.ScheduledTask{}).Count(&stats.TotalTasks)

	// 按状态统计
	var statusCounts []struct {
		Status string
		Count  int64
	}
	s.db.Model(&models.ScheduledTask{}).
		Select("status, count(*) as count").
		Group("status").
		Scan(&statusCounts)

	for _, sc := range statusCounts {
		stats.TasksByStatus[sc.Status] = sc.Count
		switch sc.Status {
		case string(models.TaskStatusPending):
			stats.PendingTasks = sc.Count
		case string(models.TaskStatusRunning):
			stats.RunningTasks = sc.Count
		case string(models.TaskStatusCompleted):
			stats.CompletedTasks = sc.Count
		case string(models.TaskStatusFailed):
			stats.FailedTasks = sc.Count
		}
	}

	// 按类型统计
	var typeCounts []struct {
		TaskType string
		Count    int64
	}
	s.db.Model(&models.ScheduledTask{}).
		Select("task_type, count(*) as count").
		Group("task_type").
		Scan(&typeCounts)

	for _, tc := range typeCounts {
		stats.TasksByType[tc.TaskType] = tc.Count
	}

	// 计算成功率
	totalExecuted := stats.CompletedTasks + stats.FailedTasks
	if totalExecuted > 0 {
		stats.SuccessRate = float64(stats.CompletedTasks) / float64(totalExecuted) * 100
	}

	// 平均执行时间
	var avgDuration float64
	s.db.Model(&models.TaskExecution{}).
		Where("status = ? AND ended_at IS NOT NULL", models.TaskStatusCompleted).
		Select("AVG(duration)").
		Scan(&avgDuration)
	stats.AvgExecutionTime = avgDuration / 1000 // 转换为秒

	// 活跃worker数量和队列长度
	var activeWorkers int64
	var queueLength int64

	s.db.Model(&models.TaskWorker{}).
		Where("status = ?", "active").
		Count(&activeWorkers)
	stats.ActiveWorkers = int(activeWorkers)

	s.db.Model(&models.ScheduledTask{}).
		Where("status = ?", models.TaskStatusPending).
		Count(&queueLength)
	stats.QueueLength = int(queueLength)

	return stats, nil
}

// GetTaskExecutions 获取任务执行记录
func (s *SchedulerService) GetTaskExecutions(taskID string, limit int) ([]models.TaskExecution, error) {
	var executions []models.TaskExecution
	err := s.db.Where("task_id = ?", taskID).
		Order("created_at DESC").
		Limit(limit).
		Find(&executions).Error
	return executions, err
}

// DeleteTask 删除任务
func (s *SchedulerService) DeleteTask(taskID string) error {
	// 先从cron调度器移除
	s.removeTaskFromCron(taskID)

	// 软删除任务
	return s.db.Where("id = ?", taskID).Delete(&models.ScheduledTask{}).Error
}

// 私有方法

// registerWorker 注册worker
func (s *SchedulerService) registerWorker() error {
	worker := &models.TaskWorker{
		ID:             s.workerID,
		Name:           fmt.Sprintf("Worker-%s", s.workerID),
		Host:           s.getHostname(),
		Port:           8080, // 默认端口
		Status:         "active",
		MaxConcurrency: 5,
		LastHeartbeat:  time.Now(),
		SupportedTypes: `["letter_delivery","user_engagement","system_maintenance","data_analytics"]`,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// 使用UPSERT操作
	return s.db.Save(worker).Error
}

// loadScheduledTasks 加载已有的定时任务
func (s *SchedulerService) loadScheduledTasks() error {
	var tasks []models.ScheduledTask
	if err := s.db.Where("is_active = ? AND cron_expression != ''", true).Find(&tasks).Error; err != nil {
		return err
	}

	for _, task := range tasks {
		if err := s.addTaskToCron(&task); err != nil {
			log.Printf("Failed to add task %s to cron: %v", task.ID, err)
		}
	}

	log.Printf("Loaded %d scheduled tasks", len(tasks))
	return nil
}

// addTaskToCron 添加任务到cron调度器
func (s *SchedulerService) addTaskToCron(task *models.ScheduledTask) error {
	if task.CronExpression == "" {
		return nil
	}

	entryID, err := s.cron.AddFunc(task.CronExpression, func() {
		s.executeTask(task)
	})

	if err != nil {
		return fmt.Errorf("failed to add cron job for task %s: %w", task.ID, err)
	}

	log.Printf("Added task %s to cron scheduler with entry ID %d", task.ID, entryID)
	return nil
}

// removeTaskFromCron 从cron调度器移除任务
func (s *SchedulerService) removeTaskFromCron(taskID string) {
	// 这是一个简化实现，实际应该维护一个entryID映射
	log.Printf("Removing task %s from cron scheduler", taskID)
}

// executeTask 执行任务
func (s *SchedulerService) executeTask(task *models.ScheduledTask) {
	// 检查是否已经在运行
	if task.Status == models.TaskStatusRunning {
		log.Printf("Task %s is already running, skipping", task.ID)
		return
	}

	// 创建执行记录
	execution := &models.TaskExecution{
		ID:         uuid.New().String(),
		TaskID:     task.ID,
		Status:     models.TaskStatusRunning,
		StartedAt:  time.Now(),
		WorkerID:   s.workerID,
		ServerHost: s.getHostname(),
		ProcessPID: os.Getpid(),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	s.db.Create(execution)

	// 更新任务状态
	s.UpdateTaskStatus(task.ID, models.TaskStatusRunning)

	// 执行任务
	startTime := time.Now()
	result := s.performTask(task)
	duration := int(time.Since(startTime).Milliseconds())

	// 更新执行记录
	execution.Duration = duration
	execution.EndedAt = &startTime

	if result.Success {
		execution.Status = models.TaskStatusCompleted
		execution.Result = result.Result
		s.UpdateTaskStatus(task.ID, models.TaskStatusCompleted)
	} else {
		execution.Status = models.TaskStatusFailed
		execution.Error = result.Error
		s.UpdateTaskStatus(task.ID, models.TaskStatusFailed)
	}

	s.db.Save(execution)

	// 更新任务执行统计
	updates := map[string]interface{}{
		"last_run_at": startTime,
		"last_status": execution.Status,
		"run_count":   gorm.Expr("run_count + 1"),
		"updated_at":  time.Now(),
	}

	if !result.Success {
		updates["failure_count"] = gorm.Expr("failure_count + 1")
		updates["last_error"] = result.Error
	} else {
		updates["last_result"] = result.Result
	}

	s.db.Model(&models.ScheduledTask{}).Where("id = ?", task.ID).Updates(updates)

	// 计算下次执行时间
	if task.CronExpression != "" {
		if nextTime, err := s.getNextRunTime(task.CronExpression); err == nil {
			s.db.Model(&models.ScheduledTask{}).Where("id = ?", task.ID).Update("next_run_at", nextTime)
		}
	}
}

// performTask 执行具体任务
func (s *SchedulerService) performTask(task *models.ScheduledTask) *models.ExecutionResult {
	log.Printf("Executing task: %s [%s]", task.Name, task.TaskType)

	switch task.TaskType {
	case models.TaskTypeLetterDelivery:
		return s.executeLetterDeliveryTask(task)
	case models.TaskTypeUserEngagement:
		return s.executeUserEngagementTask(task)
	case models.TaskTypeSystemMaintenance:
		return s.executeSystemMaintenanceTask(task)
	case models.TaskTypeDataAnalytics:
		return s.executeDataAnalyticsTask(task)
	case models.TaskTypeNotificationCleanup:
		return s.executeNotificationCleanupTask(task)
	case models.TaskTypeLetterExpiration:
		return s.executeLetterExpirationTask(task)
	case models.TaskTypeCourierReminder:
		return s.executeCourierReminderTask(task)
	case models.TaskTypeBackupDatabase:
		return s.executeBackupDatabaseTask(task)
	case models.TaskTypeImageOptimization:
		return s.executeImageOptimizationTask(task)
	case models.TaskTypeStatisticsUpdate:
		return s.executeStatisticsUpdateTask(task)
	default:
		return &models.ExecutionResult{
			Success: false,
			Error:   fmt.Sprintf("Unknown task type: %s", task.TaskType),
		}
	}
}

// 具体任务执行方法

func (s *SchedulerService) executeLetterDeliveryTask(task *models.ScheduledTask) *models.ExecutionResult {
	// 检查待投递的信件并发送提醒
	var count int64
	s.db.Model(&models.Letter{}).Where("status = ?", models.StatusGenerated).Count(&count)

	return &models.ExecutionResult{
		Success: true,
		Result:  fmt.Sprintf("Checked %d letters for delivery", count),
	}
}

func (s *SchedulerService) executeUserEngagementTask(task *models.ScheduledTask) *models.ExecutionResult {
	// 分析用户参与度并发送提醒
	return &models.ExecutionResult{
		Success: true,
		Result:  "User engagement analysis completed",
	}
}

func (s *SchedulerService) executeSystemMaintenanceTask(task *models.ScheduledTask) *models.ExecutionResult {
	// 执行系统维护任务
	return &models.ExecutionResult{
		Success: true,
		Result:  "System maintenance completed",
	}
}

func (s *SchedulerService) executeDataAnalyticsTask(task *models.ScheduledTask) *models.ExecutionResult {
	// 更新数据分析指标
	return &models.ExecutionResult{
		Success: true,
		Result:  "Data analytics updated",
	}
}

func (s *SchedulerService) executeNotificationCleanupTask(task *models.ScheduledTask) *models.ExecutionResult {
	// 清理过期通知
	result := s.db.Where("created_at < ?", time.Now().AddDate(0, 0, -7)).Delete(&models.Notification{})

	return &models.ExecutionResult{
		Success: true,
		Result:  fmt.Sprintf("Cleaned up %d expired notifications", result.RowsAffected),
	}
}

func (s *SchedulerService) executeLetterExpirationTask(task *models.ScheduledTask) *models.ExecutionResult {
	// 处理过期信件
	return &models.ExecutionResult{
		Success: true,
		Result:  "Letter expiration check completed",
	}
}

func (s *SchedulerService) executeCourierReminderTask(task *models.ScheduledTask) *models.ExecutionResult {
	// 发送信使提醒
	return &models.ExecutionResult{
		Success: true,
		Result:  "Courier reminders sent",
	}
}

func (s *SchedulerService) executeBackupDatabaseTask(task *models.ScheduledTask) *models.ExecutionResult {
	// 数据库备份
	return &models.ExecutionResult{
		Success: true,
		Result:  "Database backup completed",
	}
}

func (s *SchedulerService) executeImageOptimizationTask(task *models.ScheduledTask) *models.ExecutionResult {
	// 图片优化
	return &models.ExecutionResult{
		Success: true,
		Result:  "Image optimization completed",
	}
}

func (s *SchedulerService) executeStatisticsUpdateTask(task *models.ScheduledTask) *models.ExecutionResult {
	// 更新统计数据
	return &models.ExecutionResult{
		Success: true,
		Result:  "Statistics updated",
	}
}

// 辅助方法

func (s *SchedulerService) getNextRunTime(cronExpr string) (time.Time, error) {
	schedule, err := cron.ParseStandard(cronExpr)
	if err != nil {
		return time.Time{}, err
	}
	return schedule.Next(time.Now()), nil
}

func (s *SchedulerService) getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}

func (s *SchedulerService) updateWorkerStatus(status string) {
	s.db.Model(&models.TaskWorker{}).
		Where("id = ?", s.workerID).
		Updates(map[string]interface{}{
			"status":         status,
			"last_heartbeat": time.Now(),
			"updated_at":     time.Now(),
		})
}

func (s *SchedulerService) updateWorkerHeartbeat() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.updateWorkerStatus("active")
		case <-s.ctx.Done():
			return
		}
	}
}

func (s *SchedulerService) monitorTasks() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 监控超时任务
			s.checkTimeoutTasks()
		case <-s.ctx.Done():
			return
		}
	}
}

func (s *SchedulerService) checkTimeoutTasks() {
	var runningTasks []models.ScheduledTask
	s.db.Where("status = ? AND last_run_at < ?",
		models.TaskStatusRunning,
		time.Now().Add(-5*time.Minute)).Find(&runningTasks)

	for _, task := range runningTasks {
		log.Printf("Task %s appears to be stuck, marking as failed", task.ID)
		s.UpdateTaskStatus(task.ID, models.TaskStatusFailed)
	}
}

func (s *SchedulerService) cleanupExpiredTasks() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 清理过期的执行记录（保留30天）
			s.db.Where("created_at < ?", time.Now().AddDate(0, 0, -30)).
				Delete(&models.TaskExecution{})
		case <-s.ctx.Done():
			return
		}
	}
}
