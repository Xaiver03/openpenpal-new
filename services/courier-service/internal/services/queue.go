package services

import (
	"context"
	"courier-service/internal/models"
	"courier-service/internal/utils"
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// QueueService Redis队列服务
type QueueService struct {
	redis             *redis.Client
	db                *gorm.DB
	wsManager         *utils.WebSocketManager
	assignmentService *AssignmentService
}

// NewQueueService 创建队列服务实例
func NewQueueService(redis *redis.Client, db *gorm.DB, wsManager *utils.WebSocketManager, assignmentService *AssignmentService) *QueueService {
	return &QueueService{
		redis:             redis,
		db:                db,
		wsManager:         wsManager,
		assignmentService: assignmentService,
	}
}

// TaskQueueEvent 任务队列事件
type TaskQueueEvent struct {
	Type      string                 `json:"type"`
	TaskID    string                 `json:"task_id"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
	Retry     int                    `json:"retry"`
}

// QueueNames Redis队列名称常量
const (
	QueueTaskExpress    = "tasks:express"
	QueueTaskUrgent     = "tasks:urgent"
	QueueTaskNormal     = "tasks:normal"
	QueueTaskAssignment = "tasks:assignment"
	QueueNotification   = "notifications"
	QueueTaskRetry      = "tasks:retry"
)

// PushTaskToQueue 将任务推送到Redis队列
func (s *QueueService) PushTaskToQueue(task *models.Task) error {
	ctx := context.Background()

	event := TaskQueueEvent{
		Type:      "NEW_TASK",
		TaskID:    task.TaskID,
		Data: map[string]interface{}{
			"task_id":            task.TaskID,
			"letter_id":          task.LetterID,
			"pickup_location":    task.PickupLocation,
			"delivery_location":  task.DeliveryLocation,
			"priority":           task.Priority,
			"reward":             task.Reward,
		},
		Timestamp: time.Now(),
		Retry:     0,
	}

	eventJSON, err := json.Marshal(event)
	if err != nil {
		return err
	}

	// 根据优先级选择队列
	queueName := QueueTaskNormal
	switch task.Priority {
	case models.TaskPriorityExpress:
		queueName = QueueTaskExpress
	case models.TaskPriorityUrgent:
		queueName = QueueTaskUrgent
	}

	// 推送到队列
	err = s.redis.LPush(ctx, queueName, eventJSON).Err()
	if err != nil {
		log.Printf("Failed to push task to queue: %v", err)
		return err
	}

	// 设置队列统计
	s.updateQueueStats(queueName, "pushed")

	log.Printf("Task %s pushed to queue %s", task.TaskID, queueName)
	return nil
}

// PushAssignmentTask 推送任务分配事件
func (s *QueueService) PushAssignmentTask(taskID string) error {
	ctx := context.Background()

	event := TaskQueueEvent{
		Type:   "AUTO_ASSIGN",
		TaskID: taskID,
		Data: map[string]interface{}{
			"task_id": taskID,
		},
		Timestamp: time.Now(),
		Retry:     0,
	}

	eventJSON, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return s.redis.LPush(ctx, QueueTaskAssignment, eventJSON).Err()
}

// PushNotification 推送通知事件
func (s *QueueService) PushNotification(notification utils.WebSocketEvent) error {
	ctx := context.Background()

	notificationJSON, err := json.Marshal(notification)
	if err != nil {
		return err
	}

	return s.redis.LPush(ctx, QueueNotification, notificationJSON).Err()
}

// ConsumeTaskQueues 消费任务队列
func (s *QueueService) ConsumeTaskQueues() {
	ctx := context.Background()
	log.Println("Starting task queue consumer...")

	for {
		// 按优先级处理队列：express > urgent > normal
		result, err := s.redis.BRPop(ctx, 1*time.Second,
			QueueTaskExpress, QueueTaskUrgent, QueueTaskNormal).Result()

		if err != nil {
			if err != redis.Nil {
				log.Printf("Redis BRPop error: %v", err)
			}
			continue
		}

		queueName := result[0]
		eventData := result[1]

		var event TaskQueueEvent
		if err := json.Unmarshal([]byte(eventData), &event); err != nil {
			log.Printf("Failed to unmarshal task event: %v", err)
			continue
		}

		// 处理任务事件
		s.processTaskEvent(&event, queueName)
	}
}

// ConsumeAssignmentQueue 消费任务分配队列
func (s *QueueService) ConsumeAssignmentQueue() {
	ctx := context.Background()
	log.Println("Starting assignment queue consumer...")

	for {
		result, err := s.redis.BRPop(ctx, 2*time.Second, QueueTaskAssignment).Result()
		if err != nil {
			if err != redis.Nil {
				log.Printf("Assignment queue error: %v", err)
			}
			continue
		}

		eventData := result[1]
		var event TaskQueueEvent
		if err := json.Unmarshal([]byte(eventData), &event); err != nil {
			log.Printf("Failed to unmarshal assignment event: %v", err)
			continue
		}

		// 处理自动分配
		s.processAssignmentEvent(&event)
	}
}

// ConsumeNotificationQueue 消费通知队列
func (s *QueueService) ConsumeNotificationQueue() {
	ctx := context.Background()
	log.Println("Starting notification queue consumer...")

	for {
		result, err := s.redis.BRPop(ctx, 1*time.Second, QueueNotification).Result()
		if err != nil {
			if err != redis.Nil {
				log.Printf("Notification queue error: %v", err)
			}
			continue
		}

		notificationData := result[1]
		var notification utils.WebSocketEvent
		if err := json.Unmarshal([]byte(notificationData), &notification); err != nil {
			log.Printf("Failed to unmarshal notification: %v", err)
			continue
		}

		// 发送WebSocket通知
		s.processNotification(&notification)
	}
}

// processTaskEvent 处理任务事件
func (s *QueueService) processTaskEvent(event *TaskQueueEvent, queueName string) {
	log.Printf("Processing task event: %s from queue: %s", event.Type, queueName)

	switch event.Type {
	case "NEW_TASK":
		s.handleNewTask(event)
	case "TASK_UPDATE":
		s.handleTaskUpdate(event)
	case "TASK_TIMEOUT":
		s.handleTaskTimeout(event)
	default:
		log.Printf("Unknown task event type: %s", event.Type)
	}

	// 更新队列统计
	s.updateQueueStats(queueName, "processed")
}

// processAssignmentEvent 处理任务分配事件
func (s *QueueService) processAssignmentEvent(event *TaskQueueEvent) {
	log.Printf("Processing assignment event: %s for task: %s", event.Type, event.TaskID)

	if event.Type == "AUTO_ASSIGN" {
		// 获取任务信息
		var task models.Task
		if err := s.db.Where("task_id = ?", event.TaskID).First(&task).Error; err != nil {
			log.Printf("Task not found: %s", event.TaskID)
			return
		}

		// 尝试自动分配
		courier, err := s.assignmentService.AutoAssignTask(&task)
		if err != nil {
			log.Printf("Auto assignment failed for task %s: %v", event.TaskID, err)
			
			// 重试机制
			if event.Retry < 3 {
				s.retryAssignment(event)
			}
			return
		}

		log.Printf("Task %s auto-assigned to courier %s", event.TaskID, courier.UserID)
	}
}

// processNotification 处理通知事件
func (s *QueueService) processNotification(notification *utils.WebSocketEvent) {
	log.Printf("Processing notification: %s", notification.Type)

	// 根据通知类型发送到不同的用户群体
	switch notification.Type {
	case "COURIER_TASK_UPDATE", "NEW_TASK_ASSIGNMENT":
		if data, ok := notification.Data.(map[string]interface{}); ok {
			if userID, ok := data["courier_id"].(string); ok {
				s.wsManager.BroadcastToUser(userID, *notification)
			}
		}
	case "TASK_AUTO_ASSIGNED", "NEW_COURIER_APPLICATION":
		s.wsManager.BroadcastToAdmins(*notification)
	default:
		s.wsManager.BroadcastToAll(*notification)
	}
}

// handleNewTask 处理新任务
func (s *QueueService) handleNewTask(event *TaskQueueEvent) {
	// 推送到自动分配队列
	s.PushAssignmentTask(event.TaskID)

	// 广播新任务通知
	notification := utils.WebSocketEvent{
		Type: "NEW_TASK_AVAILABLE",
		Data: event.Data,
		Timestamp: time.Now(),
	}
	s.wsManager.BroadcastToAll(notification)
}

// handleTaskUpdate 处理任务更新
func (s *QueueService) handleTaskUpdate(event *TaskQueueEvent) {
	// 广播任务状态更新
	notification := utils.WebSocketEvent{
		Type: "TASK_STATUS_UPDATE",
		Data: event.Data,
		Timestamp: time.Now(),
	}
	
	if courierID, ok := event.Data["courier_id"].(string); ok {
		s.wsManager.BroadcastToUser(courierID, notification)
	}
	s.wsManager.BroadcastToAdmins(notification)
}

// handleTaskTimeout 处理任务超时
func (s *QueueService) handleTaskTimeout(event *TaskQueueEvent) {
	log.Printf("Handling task timeout: %s", event.TaskID)
	
	// 重新分配任务
	s.assignmentService.ReassignFailedTasks()
}

// retryAssignment 重试任务分配
func (s *QueueService) retryAssignment(event *TaskQueueEvent) {
	ctx := context.Background()
	
	// 增加重试次数
	event.Retry++
	event.Timestamp = time.Now()

	eventJSON, err := json.Marshal(event)
	if err != nil {
		log.Printf("Failed to marshal retry event: %v", err)
		return
	}

	// 延迟重试（2的指数次方分钟）
	delay := time.Duration(1<<event.Retry) * time.Minute
	
	// 推送到重试队列
	s.redis.ZAdd(ctx, QueueTaskRetry, redis.Z{
		Score:  float64(time.Now().Add(delay).Unix()),
		Member: eventJSON,
	})

	log.Printf("Task %s scheduled for retry %d in %v", event.TaskID, event.Retry, delay)
}

// ProcessRetryQueue 处理重试队列
func (s *QueueService) ProcessRetryQueue() {
	ctx := context.Background()
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			now := time.Now().Unix()
			
			// 获取到期的重试任务
			result, err := s.redis.ZRangeByScore(ctx, QueueTaskRetry, &redis.ZRangeBy{
				Min: "0",
				Max: strconv.FormatInt(now, 10),
			}).Result()

			if err != nil {
				log.Printf("Error fetching retry tasks: %v", err)
				continue
			}

			for _, eventData := range result {
				var event TaskQueueEvent
				if err := json.Unmarshal([]byte(eventData), &event); err != nil {
					continue
				}

				// 重新推送到分配队列
				s.PushAssignmentTask(event.TaskID)

				// 从重试队列中移除
				s.redis.ZRem(ctx, QueueTaskRetry, eventData)
			}
		}
	}
}

// updateQueueStats 更新队列统计
func (s *QueueService) updateQueueStats(queueName, action string) {
	ctx := context.Background()
	key := "stats:queue:" + queueName + ":" + action
	s.redis.Incr(ctx, key)
	s.redis.Expire(ctx, key, 24*time.Hour) // 24小时过期
}

// GetQueueStats 获取队列统计信息
func (s *QueueService) GetQueueStats() map[string]interface{} {
	ctx := context.Background()
	stats := make(map[string]interface{})

	queues := []string{QueueTaskExpress, QueueTaskUrgent, QueueTaskNormal, QueueTaskAssignment, QueueNotification}
	
	for _, queue := range queues {
		length, _ := s.redis.LLen(ctx, queue).Result()
		stats[queue] = map[string]interface{}{
			"length": length,
		}
	}

	// 重试队列长度
	retryLength, _ := s.redis.ZCard(ctx, QueueTaskRetry).Result()
	stats[QueueTaskRetry] = map[string]interface{}{
		"length": retryLength,
	}

	return stats
}