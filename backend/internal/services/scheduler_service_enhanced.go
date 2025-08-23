package services

import (
	"context"
	"fmt"
	"log"
	"openpenpal-backend/internal/models"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

// EnhancedSchedulerService provides distributed scheduling with locking
type EnhancedSchedulerService struct {
	baseScheduler  *SchedulerService
	lockManager    *DistributedLockManager
	schedulerTasks *SchedulerTasks
	db             *gorm.DB
}

// NewEnhancedSchedulerService creates a new enhanced scheduler service
func NewEnhancedSchedulerService(
	db *gorm.DB,
	redisClient *redis.Client,
	futureLetterSvc *FutureLetterService,
	letterService *LetterService,
	aiService *AIService,
	notificationSvc *NotificationService,
	envelopeService *EnvelopeService,
	courierService *CourierService,
) *EnhancedSchedulerService {
	// Create base scheduler service
	baseScheduler := NewSchedulerService(db)
	
	// Create lock manager
	lockManager := NewDistributedLockManager(redisClient, "scheduler:")
	
	// Create scheduler tasks
	schedulerTasks := NewSchedulerTasks(
		futureLetterSvc,
		letterService,
		aiService,
		notificationSvc,
		envelopeService,
		courierService,
	)
	
	enhanced := &EnhancedSchedulerService{
		baseScheduler:  baseScheduler,
		lockManager:    lockManager,
		schedulerTasks: schedulerTasks,
		db:            db,
	}
	
	return enhanced
}

// executeTaskWithLock executes a task with distributed locking
func (es *EnhancedSchedulerService) executeTaskWithLock(task *models.ScheduledTask) {
	ctx := context.Background()
	execution := &models.TaskExecution{
		ID:        generateExecutionID(),
		TaskID:    task.ID,
		Status:    models.SchedulerTaskStatusRunning,
		StartedAt: time.Now(),
		WorkerID:  es.getWorkerID(),
		ProcessPID: os.Getpid(),
	}

	// Try to acquire distributed lock for this task
	lockKey := fmt.Sprintf("task:%s", task.ID)
	lockExpiration := time.Duration(task.TimeoutSecs) * time.Second
	if lockExpiration < 30*time.Second {
		lockExpiration = 30 * time.Second
	}

	err := es.lockManager.RunWithLockExtension(
		ctx,
		lockKey,
		lockExpiration,
		10*time.Second, // Extend every 10 seconds
		func(lockCtx context.Context) error {
			// Save execution record
			if err := es.db.Create(execution).Error; err != nil {
				log.Printf("Failed to create execution record: %v", err)
				return err
			}

			// Update task status
			es.db.Model(task).Updates(map[string]interface{}{
				"status":      models.SchedulerTaskStatusRunning,
				"last_run_at": time.Now(),
			})

			// Execute the actual task
			result, taskErr := es.schedulerTasks.ExecuteTask(lockCtx, task)
			
			// Update execution record
			execution.EndedAt = &[]time.Time{time.Now()}[0]
			execution.Duration = int(time.Since(execution.StartedAt).Milliseconds())
			
			if taskErr != nil {
				execution.Status = models.SchedulerTaskStatusFailed
				execution.Error = taskErr.Error()
				es.db.Model(task).Updates(map[string]interface{}{
					"last_status":   models.SchedulerTaskStatusFailed,
					"last_error":    taskErr.Error(),
					"failure_count": gorm.Expr("failure_count + 1"),
				})
			} else {
				execution.Status = models.SchedulerTaskStatusCompleted
				execution.Result = result.Result
				es.db.Model(task).Updates(map[string]interface{}{
					"last_status":  models.SchedulerTaskStatusCompleted,
					"last_result":  result.Result,
					"run_count":    gorm.Expr("run_count + 1"),
				})
			}
			
			// Save execution updates
			es.db.Save(execution)
			
			// Calculate next run time
			es.updateNextRunTime(task)
			
			return taskErr
		},
	)

	if err != nil {
		log.Printf("Failed to execute task %s with lock: %v", task.ID, err)
		
		// If we couldn't acquire the lock, it means another instance is handling it
		if err.Error() == "lock not acquired" {
			log.Printf("Task %s is being handled by another instance", task.ID)
			return
		}
		
		// For other errors, record the failure
		execution.Status = models.SchedulerTaskStatusFailed
		execution.Error = fmt.Sprintf("Lock error: %v", err)
		execution.EndedAt = &[]time.Time{time.Now()}[0]
		es.db.Save(execution)
	}
}

// RegisterFSDTasks registers all FSD-required tasks
func (es *EnhancedSchedulerService) RegisterFSDTasks() error {
	return es.schedulerTasks.RegisterDefaultTasks(es.baseScheduler)
}

// GetLockStatus returns the status of distributed locks
func (es *EnhancedSchedulerService) GetLockStatus(ctx context.Context) (map[string]interface{}, error) {
	// Get all active tasks
	var tasks []models.ScheduledTask
	if err := es.db.Where("is_active = ?", true).Find(&tasks).Error; err != nil {
		return nil, err
	}
	
	lockStatus := make(map[string]interface{})
	for _, task := range tasks {
		lockKey := fmt.Sprintf("task:%s", task.ID)
		lock := es.lockManager.NewLock(lockKey, 1*time.Second)
		
		lockStatus[task.Name] = map[string]interface{}{
			"task_id": task.ID,
			"locked":  lock.IsHeld(ctx),
			"key":     lockKey,
		}
	}
	
	return lockStatus, nil
}

// ForceReleaseLock forcefully releases a lock (admin operation)
func (es *EnhancedSchedulerService) ForceReleaseLock(ctx context.Context, taskID string) error {
	lockKey := fmt.Sprintf("task:%s", taskID)
	
	// Force delete the key regardless of value (bypassing lock instance)
	err := es.lockManager.client.Del(ctx, es.lockManager.prefix+lockKey).Err()
	if err != nil {
		return fmt.Errorf("failed to force release lock: %w", err)
	}
	
	log.Printf("Force released lock for task: %s", taskID)
	return nil
}

// Helper methods

// getWorkerID returns the worker ID from base scheduler
func (es *EnhancedSchedulerService) getWorkerID() string {
	return "enhanced-scheduler-" + fmt.Sprintf("%d", os.Getpid())
}

// updateNextRunTime calculates and updates next run time for a task
func (es *EnhancedSchedulerService) updateNextRunTime(task *models.ScheduledTask) {
	if task.CronExpression != "" {
		if nextTime, err := es.getNextRunTime(task.CronExpression); err == nil {
			es.db.Model(&models.ScheduledTask{}).Where("id = ?", task.ID).Update("next_run_at", nextTime)
		}
	}
}

// getNextRunTime calculates next run time from cron expression
func (es *EnhancedSchedulerService) getNextRunTime(cronExpr string) (time.Time, error) {
	schedule, err := cron.ParseStandard(cronExpr)
	if err != nil {
		return time.Time{}, err
	}
	return schedule.Next(time.Now()), nil
}

// generateExecutionID generates a unique execution ID
func generateExecutionID() string {
	return fmt.Sprintf("exec-%d-%d", time.Now().UnixNano(), os.Getpid())
}