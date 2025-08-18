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
		if _, err := scheduler.CreateTask(&task, "system"); err != nil {
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
	log.Printf("[SchedulerTasks] Pushing daily writing inspiration")
	
	if st.aiService == nil || st.notificationSvc == nil {
		return fmt.Errorf("required services not initialized")
	}
	
	// Query users with inspiration notifications enabled
	users, err := st.getUsersWithInspirationEnabled(ctx)
	if err != nil {
		return fmt.Errorf("failed to get users with inspiration enabled: %w", err)
	}
	
	if len(users) == 0 {
		log.Printf("[SchedulerTasks] No users with inspiration notifications enabled")
		return nil
	}
	
	log.Printf("[SchedulerTasks] Found %d users with inspiration notifications enabled", len(users))
	
	// Generate and send inspiration for each user
	successCount := 0
	for _, user := range users {
		if err := st.sendDailyInspiration(ctx, &user); err != nil {
			log.Printf("[SchedulerTasks] Failed to send inspiration to user %s: %v", user.ID, err)
			continue
		}
		successCount++
	}
	
	log.Printf("[SchedulerTasks] Successfully sent daily inspiration to %d users", successCount)
	return nil
}

// getUsersWithInspirationEnabled queries users who have inspiration notifications enabled
func (st *SchedulerTasks) getUsersWithInspirationEnabled(ctx context.Context) ([]models.User, error) {
	var users []models.User
	
	// Query users with active status and inspiration notifications enabled
	// This assumes there's a notification preferences system
	// For now, get all active users and assume they want inspiration
	// In production, this should check actual notification preferences
	err := st.letterService.GetDB().Where("is_active = ?", true).Find(&users).Error
	
	return users, err
}

// sendDailyInspiration generates and sends daily inspiration to a user
func (st *SchedulerTasks) sendDailyInspiration(ctx context.Context, user *models.User) error {
	// Generate simple daily inspiration (simplified implementation)
	inspirationContent := fmt.Sprintf("今日写作主题：记录生活中的美好瞬间\n\n亲爱的%s，今天不如写一封信，分享你今天发现的小美好吧！", user.Username)
	
	// Send notification with the inspiration
	if st.notificationSvc != nil {
		notificationData := map[string]interface{}{
			"title":   "每日写作灵感",
			"content": inspirationContent,
			"type":    "writing_inspiration",
		}
		
		if err := st.notificationSvc.NotifyUser(user.ID, "writing_inspiration", notificationData); err != nil {
			log.Printf("[SchedulerTasks] Failed to send daily inspiration to user %s: %v", user.ID, err)
		} else {
			log.Printf("[SchedulerTasks] Sent daily inspiration to user %s", user.ID)
		}
	}
	
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
	
	if st.letterService == nil {
		return fmt.Errorf("letter service not initialized")
	}
	
	// Calculate cutoff date
	cutoffDate := time.Now().AddDate(0, 0, -int(daysOld))
	
	// Find unbound letters older than cutoff date
	unboundLetters, err := st.findUnboundLetters(ctx, cutoffDate)
	if err != nil {
		return fmt.Errorf("failed to find unbound letters: %w", err)
	}
	
	if len(unboundLetters) == 0 {
		log.Printf("[SchedulerTasks] No unbound letters found for cleanup")
		return nil
	}
	
	log.Printf("[SchedulerTasks] Found %d unbound letters for cleanup", len(unboundLetters))
	
	// Process each letter for cleanup
	processedCount := 0
	for _, letter := range unboundLetters {
		if err := st.processLetterCleanup(ctx, &letter, targetStatus); err != nil {
			log.Printf("[SchedulerTasks] Failed to cleanup letter %s: %v", letter.ID, err)
			continue
		}
		processedCount++
	}
	
	log.Printf("[SchedulerTasks] Successfully cleaned up %d letters", processedCount)
	return nil
}

// findUnboundLetters finds letters that are unbound and older than cutoff date
func (st *SchedulerTasks) findUnboundLetters(ctx context.Context, cutoffDate time.Time) ([]models.Letter, error) {
	var unboundLetters []models.Letter
	
	// Find letters that are:
	// - In draft status
	// - Created before cutoff date
	// - Don't have associated barcode (letter_codes table)
	// - Haven't been marked for cleanup already
	err := st.letterService.GetDB().Where(
		"status = ? AND created_at < ? AND id NOT IN (SELECT letter_id FROM letter_codes WHERE letter_id IS NOT NULL) AND status != ?",
		"draft", cutoffDate, "draft_cleanup",
	).Find(&unboundLetters).Error
	
	return unboundLetters, err
}

// processLetterCleanup processes a single letter for cleanup
func (st *SchedulerTasks) processLetterCleanup(ctx context.Context, letter *models.Letter, targetStatus string) error {
	// Update letter status to cleanup status
	updateData := map[string]interface{}{
		"status":           targetStatus,
		"cleanup_scheduled_at": time.Now(),
	}
	
	if err := st.letterService.GetDB().Model(letter).Updates(updateData).Error; err != nil {
		return fmt.Errorf("failed to update letter status: %w", err)
	}
	
	// Send notification to user about cleanup
	if st.notificationSvc != nil && letter.AuthorID != "" {
		notificationContent := fmt.Sprintf("您的草稿信件《%s》因超过7天未绑定条码，已被移入清理区。如需保留，请及时处理。", letter.Title)
		
		notificationData := map[string]interface{}{
			"title":     "信件清理提醒",
			"content":   notificationContent,
			"type":      "letter_cleanup",
			"letter_id": letter.ID,
		}
		
		if err := st.notificationSvc.NotifyUser(letter.AuthorID, "letter_cleanup", notificationData); err != nil {
			log.Printf("[SchedulerTasks] Failed to send cleanup notification to user %s: %v", letter.AuthorID, err)
		}
	}
	
	return nil
}

func (st *SchedulerTasks) checkCourierTimeouts(ctx context.Context, payload map[string]interface{}) error {
	timeoutHours, _ := payload["timeout_hours"].(float64)
	if timeoutHours == 0 {
		timeoutHours = 48
	}

	log.Printf("[SchedulerTasks] Checking courier tasks with timeout > %.0f hours", timeoutHours)
	
	if st.courierService == nil {
		return fmt.Errorf("courier service not initialized")
	}
	
	// Calculate timeout cutoff time
	timeoutDuration := time.Duration(timeoutHours) * time.Hour
	cutoffTime := time.Now().Add(-timeoutDuration)
	
	// Find overdue courier tasks
	overdueTasks, err := st.findOverdueCourierTasks(ctx, cutoffTime)
	if err != nil {
		return fmt.Errorf("failed to find overdue courier tasks: %w", err)
	}
	
	if len(overdueTasks) == 0 {
		log.Printf("[SchedulerTasks] No overdue courier tasks found")
		return nil
	}
	
	log.Printf("[SchedulerTasks] Found %d overdue courier tasks", len(overdueTasks))
	
	// Process each overdue task
	processedCount := 0
	for _, task := range overdueTasks {
		if err := st.processCourierTimeout(ctx, &task); err != nil {
			log.Printf("[SchedulerTasks] Failed to process timeout for task %s: %v", task.ID, err)
			continue
		}
		processedCount++
	}
	
	log.Printf("[SchedulerTasks] Successfully processed %d courier timeout notifications", processedCount)
	return nil
}

// findOverdueCourierTasks finds courier tasks that are overdue (simplified implementation)
func (st *SchedulerTasks) findOverdueCourierTasks(ctx context.Context, cutoffTime time.Time) ([]models.CourierTask, error) {
	// Simplified implementation - find tasks that are overdue
	// TODO: Implement when CourierService.FindOverdueTasks method is available
	var overdueTasks []models.CourierTask
	
	// For now, return empty slice - will implement when courier service methods are available
	log.Printf("[SchedulerTasks] CourierService.FindOverdueTasks not yet implemented, returning empty list")
	return overdueTasks, nil
}

// processCourierTimeout processes a single courier timeout
func (st *SchedulerTasks) processCourierTimeout(ctx context.Context, task *models.CourierTask) error {
	// Send timeout notification to courier
	if st.notificationSvc != nil && task.CourierID != "" {
		notificationContent := fmt.Sprintf("您的配送任务 %s 已超过48小时未完成，请尽快处理", task.ID)
		
		notificationData := map[string]interface{}{
			"title":      "配送任务超时提醒",
			"content":    notificationContent,
			"type":       "courier_timeout",
			"task_id":    task.ID,
			"task_type":  "delivery",
			"urgency":    "high",
		}
		
		if err := st.notificationSvc.NotifyUser(task.CourierID, "courier_timeout", notificationData); err != nil {
			log.Printf("[SchedulerTasks] Failed to send timeout notification to courier %s: %v", task.CourierID, err)
		} else {
			log.Printf("[SchedulerTasks] Sent timeout notification to courier %s for task %s", task.CourierID, task.ID)
		}
	}
	
	// Update task with timeout notification timestamp (simplified)
	// TODO: Implement when CourierService.UpdateTaskTimeoutNotification method is available
	log.Printf("[SchedulerTasks] Would update timeout notification for task %s", task.ID)
	
	// If timeout count exceeds threshold, consider reassignment (simplified check)
	// TODO: Implement timeout counting when TimeoutCount field is available
	shouldReassign := true // Simplified logic for now
	if shouldReassign {
		if err := st.considerTaskReassignment(ctx, task); err != nil {
			log.Printf("Failed to reassign overdue task %s: %v", task.ID, err)
		}
	}
	
	return nil
}

// considerTaskReassignment considers reassigning an overdue task to another courier
func (st *SchedulerTasks) considerTaskReassignment(ctx context.Context, task *models.CourierTask) error {
	log.Printf("[SchedulerTasks] Considering reassignment for overdue task %s", task.ID)
	
	// Find alternative couriers in the same area (simplified)
	// TODO: Implement when CourierService.FindAvailableCouriersInArea method is available
	var availableCouriers []string // Simplified representation
	log.Printf("[SchedulerTasks] Would find alternative couriers for task %s", task.ID)
	
	// For now, assume no alternative couriers to avoid complex logic
	if len(availableCouriers) == 0 {
		log.Printf("[SchedulerTasks] No alternative couriers available for task %s", task.ID)
		return nil
	}
	
	// Notify admin about potential reassignment
	// In production, this could automatically reassign or create reassignment request
	if st.notificationSvc != nil {
		notificationContent := fmt.Sprintf("任务 %s 多次超时，建议重新分配给其他信使。原信使：%s，可用信使：%d", task.ID, task.CourierID, len(availableCouriers))
		
		// Find admin users to notify (simplified - notify platform admins)
		var admins []models.User
		if err := st.letterService.GetDB().Where("role IN ?", []string{"platform_admin", "super_admin"}).Find(&admins).Error; err == nil {
			for _, admin := range admins {
				notificationData := map[string]interface{}{
					"title":              "任务重新分配建议",
					"content":            notificationContent,
					"type":               "task_reassignment",
					"task_id":            task.ID,
					"original_courier":   task.CourierID,
					"available_couriers": len(availableCouriers),
					"urgency":           "high",
				}
				
				if err := st.notificationSvc.NotifyUser(admin.ID, "task_reassignment", notificationData); err != nil {
					log.Printf("[SchedulerTasks] Failed to send reassignment notification to admin %s: %v", admin.ID, err)
				}
			}
		}
	}
	
	return nil
}

func (st *SchedulerTasks) processAIPenpalReplies(ctx context.Context) error {
	log.Printf("[SchedulerTasks] Processing scheduled AI penpal replies")
	
	if st.aiService == nil {
		return fmt.Errorf("AI service not initialized")
	}
	
	// Process delayed AI replies from the delay queue
	processedCount, err := st.aiService.ProcessDelayedReplies(ctx)
	if err != nil {
		log.Printf("[SchedulerTasks] Error processing delayed AI replies: %v", err)
		return err
	}
	
	log.Printf("[SchedulerTasks] Successfully processed %d delayed AI replies", processedCount)
	return nil
}