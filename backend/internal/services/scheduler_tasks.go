package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"openpenpal-backend/internal/models"
	"time"
)

// SchedulerTasks contains all automated task implementations
type SchedulerTasks struct {
	futureLetterSvc  *FutureLetterService
	letterService    *LetterService
	aiService        *AIService
	notificationSvc  *NotificationService
	envelopeService  *EnvelopeService
	courierService   *CourierService
}

// NewSchedulerTasks creates a new scheduler tasks instance
func NewSchedulerTasks(
	futureLetterSvc *FutureLetterService,
	letterService *LetterService,
	aiService *AIService,
	notificationSvc *NotificationService,
	envelopeService *EnvelopeService,
	courierService *CourierService,
) *SchedulerTasks {
	return &SchedulerTasks{
		futureLetterSvc:  futureLetterSvc,
		letterService:    letterService,
		aiService:        aiService,
		notificationSvc:  notificationSvc,
		envelopeService:  envelopeService,
		courierService:   courierService,
	}
}

// RegisterDefaultTasks registers all FSD-required automated tasks
func (st *SchedulerTasks) RegisterDefaultTasks(scheduler *SchedulerService) error {
	tasks := []models.CreateTaskRequest{
		{
			Name:           "Future Letter Auto-unlock",
			Description:    "Automatically unlock scheduled future letters every 10 minutes",
			TaskType:       models.TaskType("future_letter_unlock"),
			Priority:       models.TaskPriorityHigh,
			CronExpression: "*/10 * * * *", // Every 10 minutes
			Payload: map[string]interface{}{
				"handler": "ProcessFutureLetters",
			},
			MaxRetries:  3,
			TimeoutSecs: 300,
		},
		{
			Name:           "Daily Writing Inspiration Push",
			Description:    "Send daily writing inspiration to users at 8 PM",
			TaskType:       models.TaskType("writing_inspiration_push"),
			Priority:       models.TaskPriorityNormal,
			CronExpression: "0 20 * * *", // Daily at 20:00
			Payload: map[string]interface{}{
				"handler": "PushDailyInspiration",
			},
			MaxRetries:  3,
			TimeoutSecs: 600,
		},
		{
			Name:           "Letter Status Cleanup",
			Description:    "Clean up unbound letters older than 7 days",
			TaskType:       models.TaskType("letter_cleanup"),
			Priority:       models.TaskPriorityNormal,
			CronExpression: "0 3 * * *", // Daily at 03:00
			Payload: map[string]interface{}{
				"handler":     "CleanupUnboundLetters",
				"days_old":    7,
				"target_status": "draft_cleanup",
			},
			MaxRetries:  3,
			TimeoutSecs: 1800,
		},
		{
			Name:           "Courier Timeout Reminder",
			Description:    "Check for courier tasks that haven't been delivered after 48 hours",
			TaskType:       models.TaskType("courier_timeout_check"),
			Priority:       models.TaskPriorityHigh,
			CronExpression: "0 * * * *", // Every hour
			Payload: map[string]interface{}{
				"handler":       "CheckCourierTimeouts",
				"timeout_hours": 48,
			},
			MaxRetries:  3,
			TimeoutSecs: 600,
		},
		{
			Name:           "AI Penpal Scheduled Replies",
			Description:    "Process scheduled AI penpal replies",
			TaskType:       models.TaskType("ai_penpal_reply"),
			Priority:       models.TaskPriorityNormal,
			CronExpression: "0 * * * *", // Every hour
			Payload: map[string]interface{}{
				"handler": "ProcessAIPenpalReplies",
			},
			MaxRetries:  3,
			TimeoutSecs: 900,
		},
	}

	for _, task := range tasks {
		if err := scheduler.CreateTask(context.Background(), &task); err != nil {
			log.Printf("Failed to register task %s: %v", task.Name, err)
			continue
		}
		log.Printf("Successfully registered task: %s", task.Name)
	}

	return nil
}

// ExecuteTask executes a scheduled task based on its handler
func (st *SchedulerTasks) ExecuteTask(ctx context.Context, task *models.ScheduledTask) (*models.ExecutionResult, error) {
	startTime := time.Now()
	result := &models.ExecutionResult{
		Success:  false,
		Duration: 0,
		Metadata: make(map[string]interface{}),
	}

	// Parse payload to get handler
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(task.Payload), &payload); err != nil {
		result.Error = fmt.Sprintf("Failed to parse payload: %v", err)
		return result, err
	}

	handler, ok := payload["handler"].(string)
	if !ok {
		result.Error = "Handler not specified in payload"
		return result, fmt.Errorf("handler not specified")
	}

	// Execute based on handler
	var err error
	switch handler {
	case "ProcessFutureLetters":
		err = st.processFutureLetters(ctx)
	case "PushDailyInspiration":
		err = st.pushDailyInspiration(ctx)
	case "CleanupUnboundLetters":
		err = st.cleanupUnboundLetters(ctx, payload)
	case "CheckCourierTimeouts":
		err = st.checkCourierTimeouts(ctx, payload)
	case "ProcessAIPenpalReplies":
		err = st.processAIPenpalReplies(ctx)
	default:
		err = fmt.Errorf("unknown handler: %s", handler)
	}

	// Calculate duration
	result.Duration = int(time.Since(startTime).Milliseconds())

	if err != nil {
		result.Error = err.Error()
		return result, err
	}

	result.Success = true
	result.Result = fmt.Sprintf("Task completed successfully in %dms", result.Duration)
	return result, nil
}

// Task implementations

func (st *SchedulerTasks) processFutureLetters(ctx context.Context) error {
	if st.futureLetterSvc == nil {
		return fmt.Errorf("future letter service not initialized")
	}
	return st.futureLetterSvc.ProcessScheduledLetters(ctx)
}

func (st *SchedulerTasks) pushDailyInspiration(ctx context.Context) error {
	// Get all active users who have inspiration notifications enabled
	// This would be implemented based on your user preferences
	log.Printf("[SchedulerTasks] Pushing daily writing inspiration")
	
	// Implementation would:
	// 1. Query users with inspiration notifications enabled
	// 2. Generate personalized inspiration for each user
	// 3. Send notifications via preferred channel (email/push/in-app)
	
	// Placeholder for now
	return nil
}

func (st *SchedulerTasks) cleanupUnboundLetters(ctx context.Context, payload map[string]interface{}) error {
	daysOld, _ := payload["days_old"].(float64)
	if daysOld == 0 {
		daysOld = 7
	}

	targetStatus, _ := payload["target_status"].(string)
	if targetStatus == "" {
		targetStatus = "draft_cleanup"
	}

	log.Printf("[SchedulerTasks] Cleaning up letters older than %.0f days", daysOld)
	
	// Implementation would:
	// 1. Find letters in draft status without barcodes older than N days
	// 2. Move them to cleanup status
	// 3. Notify users about pending cleanup
	
	return nil
}

func (st *SchedulerTasks) checkCourierTimeouts(ctx context.Context, payload map[string]interface{}) error {
	timeoutHours, _ := payload["timeout_hours"].(float64)
	if timeoutHours == 0 {
		timeoutHours = 48
	}

	log.Printf("[SchedulerTasks] Checking courier tasks with timeout > %.0f hours", timeoutHours)
	
	// Implementation would:
	// 1. Find courier tasks accepted but not delivered after timeout
	// 2. Send reminder notifications to couriers
	// 3. Optionally reassign tasks if configured
	
	return nil
}

func (st *SchedulerTasks) processAIPenpalReplies(ctx context.Context) error {
	log.Printf("[SchedulerTasks] Processing scheduled AI penpal replies")
	
	// This is already implemented in DelayQueueService
	// We just need to ensure it's being called regularly
	
	return nil
}