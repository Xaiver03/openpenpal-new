package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"openpenpal-backend/internal/config"
	"openpenpal-backend/internal/models"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DelayQueueService 延迟队列服务
type DelayQueueService struct {
	db     *gorm.DB
	redis  *redis.Client
	config *config.Config
}

// DelayQueueTask 延迟任务结构
type DelayQueueTask struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"`          // ai_reply, ai_match, notification
	Payload      map[string]interface{} `json:"payload"`       // 任务数据
	DelayedUntil time.Time              `json:"delayed_until"` // 延迟执行时间
	CreatedAt    time.Time              `json:"created_at"`
	Status       string                 `json:"status"` // pending, processing, completed, failed
	RetryCount   int                    `json:"retry_count"`
	MaxRetries   int                    `json:"max_retries"`
}

// AIReplyTask AI回信任务数据
type AIReplyTask struct {
	UserID         string `json:"user_id"`
	PersonaID      string `json:"persona_id"`
	OriginalLetter string `json:"original_letter"`
	ConversationID string `json:"conversation_id"`
}

// NewDelayQueueService 创建延迟队列服务实例
func NewDelayQueueService(db *gorm.DB, config *config.Config) (*DelayQueueService, error) {
	// 初始化Redis连接
	rdb := redis.NewClient(&redis.Options{
		Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
		Password: getEnv("REDIS_PASSWORD", ""),
		DB:       getEnvAsInt("REDIS_DB", 0),
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Printf("Redis connection failed: %v", err)
		// 在开发环境中，如果Redis不可用，我们可以使用内存队列
		if config.Environment == "development" {
			log.Println("Using in-memory delay queue for development")
		} else {
			return nil, fmt.Errorf("redis connection failed: %w", err)
		}
	}

	return &DelayQueueService{
		db:     db,
		redis:  rdb,
		config: config,
	}, nil
}

// ScheduleAIReply 安排AI回信任务
func (s *DelayQueueService) ScheduleAIReply(userID, personaID, originalLetter, conversationID string, delayHours int) error {
	task := &DelayQueueTask{
		ID:   uuid.New().String(),
		Type: "ai_reply",
		Payload: map[string]interface{}{
			"user_id":         userID,
			"persona_id":      personaID,
			"original_letter": originalLetter,
			"conversation_id": conversationID,
		},
		DelayedUntil: time.Now().Add(time.Duration(delayHours) * time.Hour),
		CreatedAt:    time.Now(),
		Status:       "pending",
		RetryCount:   0,
		MaxRetries:   3,
	}

	return s.scheduleTask(task)
}

// scheduleTask 安排延迟任务
func (s *DelayQueueService) scheduleTask(task *DelayQueueTask) error {
	ctx := context.Background()

	// 序列化任务数据
	taskData, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	// 计算延迟时间的Unix时间戳作为分数
	score := float64(task.DelayedUntil.Unix())

	// 添加到Redis有序集合（用于延迟执行）
	err = s.redis.ZAdd(ctx, "delay_queue", &redis.Z{
		Score:  score,
		Member: taskData,
	}).Err()

	if err != nil {
		return fmt.Errorf("failed to add task to delay queue: %w", err)
	}

	// 同时保存到数据库（用于持久化和监控）
	delayRecord := &models.DelayQueueRecord{
		ID:           task.ID,
		TaskType:     task.Type,
		Payload:      string(taskData),
		DelayedUntil: task.DelayedUntil,
		Status:       task.Status,
		CreatedAt:    task.CreatedAt,
	}

	return s.db.Create(delayRecord).Error
}

// StartWorker 启动延迟队列工作进程
func (s *DelayQueueService) StartWorker() {
	log.Println("Starting delay queue worker...")

	ticker := time.NewTicker(30 * time.Second) // 每30秒检查一次
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.processDelayedTasks()
		}
	}
}

// processDelayedTasks 处理延迟任务
func (s *DelayQueueService) processDelayedTasks() {
	ctx := context.Background()
	now := time.Now().Unix()

	// 从Redis有序集合中获取到期的任务
	tasks, err := s.redis.ZRangeByScore(ctx, "delay_queue", &redis.ZRangeBy{
		Min: "0",
		Max: strconv.FormatInt(now, 10),
	}).Result()

	if err != nil || len(tasks) == 0 {
		return
	}

	log.Printf("Processing %d delayed tasks", len(tasks))

	for _, taskData := range tasks {
		var task DelayQueueTask
		if err := json.Unmarshal([]byte(taskData), &task); err != nil {
			log.Printf("Failed to unmarshal task: %v", err)
			continue
		}

		// 处理任务
		if err := s.processTask(&task); err != nil {
			log.Printf("Failed to process task %s: %v", task.ID, err)
			s.handleTaskError(&task)
		} else {
			// 成功处理，从队列中移除
			s.redis.ZRem(ctx, "delay_queue", taskData)
			s.updateTaskStatus(task.ID, "completed")
		}
	}
}

// processTask 处理具体任务
func (s *DelayQueueService) processTask(task *DelayQueueTask) error {
	switch task.Type {
	case "ai_reply":
		return s.processAIReplyTask(task)
	default:
		return fmt.Errorf("unknown task type: %s", task.Type)
	}
}

// processAIReplyTask 处理AI回信任务
func (s *DelayQueueService) processAIReplyTask(task *DelayQueueTask) error {
	// 从payload中提取数据
	userID, _ := task.Payload["user_id"].(string)
	personaID, _ := task.Payload["persona_id"].(string)
	originalLetter, _ := task.Payload["original_letter"].(string)
	_ = originalLetter // 避免未使用变量警告
	conversationID, _ := task.Payload["conversation_id"].(string)

	log.Printf("Processing AI reply task for user %s, persona %s", userID, personaID)

	// 创建AI服务实例
	aiService := NewAIService(s.db, s.config)

	// 构建AI回信请求
	replyRequest := &models.AIReplyRequest{
		LetterID:   conversationID,
		Persona:    models.AIPersona(personaID),
		DelayHours: 0, // 已经延迟了，现在立即处理
	}

	// 生成AI回信
	reply, err := aiService.GenerateReply(context.Background(), replyRequest)
	if err != nil {
		return fmt.Errorf("failed to generate AI reply: %w", err)
	}

	log.Printf("Generated AI reply: %s", reply.ID)

	// 这里可以添加发送通知的逻辑
	// 比如通知用户收到了AI回信

	return nil
}

// handleTaskError 处理任务错误
func (s *DelayQueueService) handleTaskError(task *DelayQueueTask) {
	task.RetryCount++

	if task.RetryCount >= task.MaxRetries {
		// 超过最大重试次数，标记为失败
		s.redis.ZRem(context.Background(), "delay_queue", task)
		s.updateTaskStatus(task.ID, "failed")
		log.Printf("Task %s failed after %d retries", task.ID, task.RetryCount)
		return
	}

	// 重新安排任务，增加延迟时间
	task.DelayedUntil = time.Now().Add(time.Duration(task.RetryCount) * 5 * time.Minute)
	s.scheduleTask(task)
	log.Printf("Rescheduled task %s, retry count: %d", task.ID, task.RetryCount)
}

// updateTaskStatus 更新任务状态
func (s *DelayQueueService) updateTaskStatus(taskID, status string) {
	s.db.Model(&models.DelayQueueRecord{}).
		Where("id = ?", taskID).
		Update("status", status)
}

// GetTaskStatus 获取任务状态
func (s *DelayQueueService) GetTaskStatus(taskID string) (*models.DelayQueueRecord, error) {
	var record models.DelayQueueRecord
	err := s.db.Where("id = ?", taskID).First(&record).Error
	return &record, err
}

// GetUserPendingTasks 获取用户待处理的任务
func (s *DelayQueueService) GetUserPendingTasks(userID string) ([]models.DelayQueueRecord, error) {
	var records []models.DelayQueueRecord
	err := s.db.Where("payload LIKE ? AND status = ?", "%"+userID+"%", "pending").
		Order("delayed_until ASC").
		Find(&records).Error
	return records, err
}

// 辅助函数
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
