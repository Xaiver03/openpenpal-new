package services

import (
	"fmt"
	"log"
	"sync"
	"time"

	"openpenpal-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CreditActivityScheduler 积分活动调度器
type CreditActivityScheduler struct {
	db                      *gorm.DB
	creditActivityService   *CreditActivityService
	ticker                  *time.Ticker
	stopChan               chan bool
	running                bool
	mu                     sync.RWMutex
	
	// 配置
	interval               time.Duration
	maxConcurrentTasks     int
	retryMaxAttempts       int
	retryBackoffMultiplier float64
}

// NewCreditActivityScheduler 创建调度器实例
func NewCreditActivityScheduler(db *gorm.DB, creditActivityService *CreditActivityService) *CreditActivityScheduler {
	return &CreditActivityScheduler{
		db:                      db,
		creditActivityService:   creditActivityService,
		stopChan:               make(chan bool),
		interval:               30 * time.Second,  // 每30秒检查一次
		maxConcurrentTasks:     5,                // 最多并发执行5个任务
		retryMaxAttempts:       3,                // 最大重试次数
		retryBackoffMultiplier: 2.0,              // 重试回退倍数
	}
}

// Start 启动调度器
func (s *CreditActivityScheduler) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("scheduler is already running")
	}

	log.Println("Starting credit activity scheduler...")
	s.ticker = time.NewTicker(s.interval)
	s.running = true

	go s.run()
	log.Printf("Credit activity scheduler started with interval: %v", s.interval)
	return nil
}

// Stop 停止调度器
func (s *CreditActivityScheduler) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return fmt.Errorf("scheduler is not running")
	}

	log.Println("Stopping credit activity scheduler...")
	s.ticker.Stop()
	s.stopChan <- true
	s.running = false
	
	log.Println("Credit activity scheduler stopped")
	return nil
}

// IsRunning 检查调度器是否运行中
func (s *CreditActivityScheduler) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// run 主调度循环
func (s *CreditActivityScheduler) run() {
	for {
		select {
		case <-s.ticker.C:
			s.processScheduledTasks()
		case <-s.stopChan:
			return
		}
	}
}

// processScheduledTasks 处理定时任务
func (s *CreditActivityScheduler) processScheduledTasks() {
	// 获取待处理的调度任务
	tasks, err := s.getPendingScheduledTasks()
	if err != nil {
		log.Printf("Error getting pending tasks: %v", err)
		return
	}

	if len(tasks) == 0 {
		return
	}

	log.Printf("Processing %d scheduled tasks", len(tasks))

	// 使用工作池限制并发数
	taskChan := make(chan *models.CreditActivitySchedule, len(tasks))
	resultChan := make(chan error, len(tasks))

	// 启动工作者
	for i := 0; i < minInt(s.maxConcurrentTasks, len(tasks)); i++ {
		go s.taskWorker(taskChan, resultChan)
	}

	// 发送任务到工作池
	for _, task := range tasks {
		taskChan <- task
	}
	close(taskChan)

	// 等待所有任务完成
	for i := 0; i < len(tasks); i++ {
		if err := <-resultChan; err != nil {
			log.Printf("Task execution error: %v", err)
		}
	}
}

// taskWorker 任务工作者
func (s *CreditActivityScheduler) taskWorker(taskChan <-chan *models.CreditActivitySchedule, resultChan chan<- error) {
	for task := range taskChan {
		err := s.executeTask(task)
		resultChan <- err
	}
}

// getPendingScheduledTasks 获取待处理的定时任务
func (s *CreditActivityScheduler) getPendingScheduledTasks() ([]*models.CreditActivitySchedule, error) {
	var tasks []*models.CreditActivitySchedule
	
	now := time.Now()
	
	err := s.db.Where("status = ? AND scheduled_time <= ?", "pending", now).
		Or("status = ? AND next_retry_time IS NOT NULL AND next_retry_time <= ?", "failed", now).
		Order("scheduled_time ASC").
		Find(&tasks).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to get pending tasks: %w", err)
	}

	return tasks, nil
}

// executeTask 执行单个任务
func (s *CreditActivityScheduler) executeTask(schedule *models.CreditActivitySchedule) error {
	// 更新任务状态为执行中
	now := time.Now()
	updateData := map[string]interface{}{
		"status":        "executing",
		"executed_time": now,
	}
	
	if err := s.db.Model(schedule).Updates(updateData).Error; err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	// 执行具体任务
	var taskError error
	switch schedule.ExecutionDetails {
	case nil:
		// 默认执行 ProcessScheduledActivities
		taskError = s.creditActivityService.ProcessScheduledActivities()
	default:
		// 处理其他类型的任务（可扩展）
		taskError = s.executeCustomTask(schedule)
	}

	// 更新任务结果
	if taskError != nil {
		return s.handleTaskFailure(schedule, taskError)
	} else {
		return s.handleTaskSuccess(schedule)
	}
}

// executeCustomTask 执行自定义任务
func (s *CreditActivityScheduler) executeCustomTask(schedule *models.CreditActivitySchedule) error {
	// 这里可以根据 ExecutionDetails 中的任务类型执行不同的任务
	// 目前先执行默认的 ProcessScheduledActivities
	return s.creditActivityService.ProcessScheduledActivities()
}

// handleTaskSuccess 处理任务成功
func (s *CreditActivityScheduler) handleTaskSuccess(schedule *models.CreditActivitySchedule) error {
	updateData := map[string]interface{}{
		"status":        "completed",
		"error_message": "",
		"retry_count":   0,
	}

	err := s.db.Model(schedule).Updates(updateData).Error
	if err != nil {
		return fmt.Errorf("failed to update successful task: %w", err)
	}

	log.Printf("Task %s completed successfully", schedule.ID)
	
	// 如果是重复任务，创建下一次执行
	if err := s.scheduleNextExecution(schedule); err != nil {
		log.Printf("Failed to schedule next execution: %v", err)
	}

	return nil
}

// handleTaskFailure 处理任务失败
func (s *CreditActivityScheduler) handleTaskFailure(schedule *models.CreditActivitySchedule, taskError error) error {
	schedule.RetryCount++
	
	updateData := map[string]interface{}{
		"error_message": taskError.Error(),
		"retry_count":   schedule.RetryCount,
	}

	// 检查是否超过最大重试次数
	if schedule.RetryCount >= s.retryMaxAttempts {
		updateData["status"] = "failed"
		updateData["next_retry_time"] = nil
	} else {
		// 计算下次重试时间（指数回退）
		backoffDuration := time.Duration(float64(time.Minute) * 
			pow(s.retryBackoffMultiplier, float64(schedule.RetryCount-1)))
		nextRetry := time.Now().Add(backoffDuration)
		
		updateData["status"] = "failed"
		updateData["next_retry_time"] = nextRetry
		
		log.Printf("Task %s failed, retry %d/%d scheduled for %v", 
			schedule.ID, schedule.RetryCount, s.retryMaxAttempts, nextRetry)
	}

	err := s.db.Model(schedule).Updates(updateData).Error
	if err != nil {
		return fmt.Errorf("failed to update failed task: %w", err)
	}

	return taskError
}

// scheduleNextExecution 安排下次执行（用于重复任务）
func (s *CreditActivityScheduler) scheduleNextExecution(schedule *models.CreditActivitySchedule) error {
	// 获取关联的活动
	var activity models.CreditActivity
	if err := s.db.First(&activity, schedule.ActivityID).Error; err != nil {
		return fmt.Errorf("failed to get activity: %w", err)
	}

	// 如果活动不需要重复执行，直接返回
	if activity.RepeatPattern == "" || activity.RepeatInterval <= 0 {
		return nil
	}

	// 计算下次执行时间
	nextTime, err := s.calculateNextExecutionTime(schedule.ScheduledTime, activity.RepeatPattern, activity.RepeatInterval)
	if err != nil {
		return fmt.Errorf("failed to calculate next execution time: %w", err)
	}

	// 检查是否超过重复结束时间
	if activity.RepeatEndDate != nil && nextTime.After(*activity.RepeatEndDate) {
		log.Printf("Activity %s repeat pattern ended", activity.ID)
		return nil
	}

	// 创建新的调度记录
	nextSchedule := &models.CreditActivitySchedule{
		ID:               uuid.New(),
		ActivityID:       activity.ID,
		ScheduledTime:    nextTime,
		Status:          "pending",
		ExecutionDetails: schedule.ExecutionDetails,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := s.db.Create(nextSchedule).Error; err != nil {
		return fmt.Errorf("failed to create next schedule: %w", err)
	}

	log.Printf("Scheduled next execution for activity %s at %v", activity.ID, nextTime)
	return nil
}

// calculateNextExecutionTime 计算下次执行时间
func (s *CreditActivityScheduler) calculateNextExecutionTime(lastTime time.Time, pattern string, interval int) (time.Time, error) {
	switch pattern {
	case "daily":
		return lastTime.AddDate(0, 0, interval), nil
	case "weekly":
		return lastTime.AddDate(0, 0, 7*interval), nil
	case "monthly":
		return lastTime.AddDate(0, interval, 0), nil
	case "yearly":
		return lastTime.AddDate(interval, 0, 0), nil
	case "hourly":
		return lastTime.Add(time.Hour * time.Duration(interval)), nil
	case "minutely":
		return lastTime.Add(time.Minute * time.Duration(interval)), nil
	default:
		return time.Time{}, fmt.Errorf("unsupported repeat pattern: %s", pattern)
	}
}

// ScheduleActivity 安排活动执行
func (s *CreditActivityScheduler) ScheduleActivity(activityID uuid.UUID, scheduledTime time.Time, executionDetails map[string]interface{}) (*models.CreditActivitySchedule, error) {
	schedule := &models.CreditActivitySchedule{
		ID:            uuid.New(),
		ActivityID:    activityID,
		ScheduledTime: scheduledTime,
		Status:        "pending",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// 如果有执行详情，设置为JSON
	if executionDetails != nil {
		// 这里可以将 executionDetails 转换为 datatypes.JSON
		// 暂时留空，根据需要实现
	}

	if err := s.db.Create(schedule).Error; err != nil {
		return nil, fmt.Errorf("failed to create schedule: %w", err)
	}

	log.Printf("Activity %s scheduled for execution at %v", activityID, scheduledTime)
	return schedule, nil
}

// GetScheduledTasks 获取调度任务列表
func (s *CreditActivityScheduler) GetScheduledTasks(status string, limit int) ([]*models.CreditActivitySchedule, error) {
	var tasks []*models.CreditActivitySchedule
	
	query := s.db.Preload("Activity").Order("scheduled_time ASC")
	
	if status != "" {
		query = query.Where("status = ?", status)
	}
	
	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&tasks).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get scheduled tasks: %w", err)
	}

	return tasks, nil
}

// CancelScheduledTask 取消调度任务
func (s *CreditActivityScheduler) CancelScheduledTask(scheduleID uuid.UUID) error {
	updateData := map[string]interface{}{
		"status":     "cancelled",
		"updated_at": time.Now(),
	}

	err := s.db.Model(&models.CreditActivitySchedule{}).
		Where("id = ? AND status IN (?)", scheduleID, []string{"pending", "failed"}).
		Updates(updateData).Error

	if err != nil {
		return fmt.Errorf("failed to cancel scheduled task: %w", err)
	}

	log.Printf("Scheduled task %s cancelled", scheduleID)
	return nil
}

// GetSchedulerStatus 获取调度器状态
func (s *CreditActivityScheduler) GetSchedulerStatus() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 获取任务统计
	var stats struct {
		Pending   int64
		Executing int64
		Completed int64
		Failed    int64
		Cancelled int64
	}

	s.db.Model(&models.CreditActivitySchedule{}).Where("status = ?", "pending").Count(&stats.Pending)
	s.db.Model(&models.CreditActivitySchedule{}).Where("status = ?", "executing").Count(&stats.Executing)
	s.db.Model(&models.CreditActivitySchedule{}).Where("status = ?", "completed").Count(&stats.Completed)
	s.db.Model(&models.CreditActivitySchedule{}).Where("status = ?", "failed").Count(&stats.Failed)
	s.db.Model(&models.CreditActivitySchedule{}).Where("status = ?", "cancelled").Count(&stats.Cancelled)

	return map[string]interface{}{
		"running":                s.running,
		"interval_seconds":       int(s.interval.Seconds()),
		"max_concurrent_tasks":   s.maxConcurrentTasks,
		"retry_max_attempts":     s.retryMaxAttempts,
		"task_statistics": map[string]int64{
			"pending":   stats.Pending,
			"executing": stats.Executing,
			"completed": stats.Completed,
			"failed":    stats.Failed,
			"cancelled": stats.Cancelled,
		},
	}
}

// 辅助函数
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func pow(base, exp float64) float64 {
	result := 1.0
	for exp > 0 {
		result *= base
		exp--
	}
	return result
}