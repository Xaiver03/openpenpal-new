package services

import (
	"context"
	"fmt"
	"log"
	"openpenpal-backend/internal/models"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

// EnhancedSchedulerService extends SchedulerService with distributed locking
type EnhancedSchedulerService struct {
	*SchedulerService
	lockManager   *DistributedLockManager
	schedulerTasks *SchedulerTasks
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
		SchedulerService: baseScheduler,
		lockManager:     lockManager,
		schedulerTasks:  schedulerTasks,
	}
	
	// Override the executeTask method to use distributed locking
	baseScheduler.executeTask = enhanced.executeTaskWithLock
	
	return enhanced
}

// executeTaskWithLock executes a task with distributed locking
func (es *EnhancedSchedulerService) executeTaskWithLock(task *models.ScheduledTask) {
	ctx := context.Background()
	execution := &models.TaskExecution{
		ID:        generateID(),
		TaskID:    task.ID,
		Status:    models.TaskStatusRunning,
		StartedAt: time.Now(),
		WorkerID:  es.workerID,
		ProcessPID: es.getProcessPID(),
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
				"status":      models.TaskStatusRunning,
				"last_run_at": time.Now(),
			})

			// Execute the actual task
			result, taskErr := es.schedulerTasks.ExecuteTask(lockCtx, task)
			
			// Update execution record
			execution.EndedAt = &[]time.Time{time.Now()}[0]
			execution.Duration = int(time.Since(execution.StartedAt).Milliseconds())
			
			if taskErr != nil {
				execution.Status = models.TaskStatusFailed
				execution.Error = taskErr.Error()
				es.db.Model(task).Updates(map[string]interface{}{
					"last_status":   models.TaskStatusFailed,
					"last_error":    taskErr.Error(),
					"failure_count": gorm.Expr("failure_count + 1"),
				})
			} else {
				execution.Status = models.TaskStatusCompleted
				execution.Result = result.Result
				es.db.Model(task).Updates(map[string]interface{}{
					"last_status":  models.TaskStatusCompleted,
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
		execution.Status = models.TaskStatusFailed
		execution.Error = fmt.Sprintf("Lock error: %v", err)
		execution.EndedAt = &[]time.Time{time.Now()}[0]
		es.db.Save(execution)
	}
}

// RegisterFSDTasks registers all FSD-required tasks
func (es *EnhancedSchedulerService) RegisterFSDTasks() error {
	return es.schedulerTasks.RegisterDefaultTasks(es.SchedulerService)
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
	
	// Create a lock instance just to release it
	lock := es.lockManager.NewLock(lockKey, 1*time.Second)
	
	// Force delete the key regardless of value
	err := es.lockManager.client.Del(ctx, es.lockManager.prefix+lockKey).Err()
	if err != nil {
		return fmt.Errorf("failed to force release lock: %w", err)
	}
	
	log.Printf("Force released lock for task: %s", taskID)
	return nil
}