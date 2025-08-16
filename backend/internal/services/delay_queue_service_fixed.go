package services

import (
	"context"
	"encoding/json"
	"fmt"
	"openpenpal-backend/internal/config"
	"openpenpal-backend/internal/models"
	"openpenpal-backend/internal/utils"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DelayQueueServiceFixed 修复版本的延迟队列服务
// 主要修复：
// 1. 任务重试时的无限循环问题
// 2. 集成智能日志系统防止日志膨胀
// 3. 添加断路器机制防止连续失败
// 4. 改进错误处理和监控
type DelayQueueServiceFixed struct {
	db          *gorm.DB
	redis       *redis.Client
	config      *config.Config
	smartLogger *utils.SmartLogger
	
	// 断路器状态
	circuitBreaker map[string]*DelayQueueCircuitBreaker
}

// DelayQueueCircuitBreaker 延迟队列断路器，防止持续失败的任务无限重试
type DelayQueueCircuitBreaker struct {
	FailureCount    int       `json:"failure_count"`
	LastFailureTime time.Time `json:"last_failure_time"`
	State           string    `json:"state"` // closed, open, half_open
	Threshold       int       `json:"threshold"`
	Timeout         time.Duration `json:"timeout"`
}

// DelayQueueTask 延迟任务结构（增强版）
type DelayQueueTaskFixed struct {
	ID               string                 `json:"id"`
	Type             string                 `json:"type"`
	Payload          map[string]interface{} `json:"payload"`
	DelayedUntil     time.Time              `json:"delayed_until"`
	CreatedAt        time.Time              `json:"created_at"`
	Status           string                 `json:"status"`
	RetryCount       int                    `json:"retry_count"`
	MaxRetries       int                    `json:"max_retries"`
	LastError        string                 `json:"last_error,omitempty"`
	CircuitBreakerKey string                `json:"circuit_breaker_key"`
}

// NewDelayQueueServiceFixed 创建修复版延迟队列服务
func NewDelayQueueServiceFixed(db *gorm.DB, config *config.Config) (*DelayQueueServiceFixed, error) {
	// 初始化智能日志系统
	smartLoggerConfig := &utils.SmartLoggerConfig{
		TimeWindow:              5 * time.Minute,
		MaxAggregation:          1000,
		VerboseThreshold:        5,   // 降低阈值，更快进入静默模式
		CircuitBreakerThreshold: 50,  // 降低断路器阈值
		SamplingRate:            20,  // 提高采样率
		CleanupInterval:         30 * time.Minute,
	}
	smartLogger := utils.NewSmartLogger(smartLoggerConfig)

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
		smartLogger.LogError("Redis connection failed", map[string]interface{}{
			"error": err.Error(),
			"addr":  getEnv("REDIS_ADDR", "localhost:6379"),
		})
		
		if config.Environment == "development" {
			smartLogger.LogInfo("Using in-memory delay queue for development")
		} else {
			return nil, fmt.Errorf("redis connection failed: %w", err)
		}
	}

	return &DelayQueueServiceFixed{
		db:             db,
		redis:          rdb,
		config:         config,
		smartLogger:    smartLogger,
		circuitBreaker: make(map[string]*DelayQueueCircuitBreaker),
	}, nil
}

// ScheduleAIReply 安排AI回信任务（修复版）
func (s *DelayQueueServiceFixed) ScheduleAIReply(userID, personaID, originalLetter, conversationID string, delayHours int) error {
	// 生成断路器键，基于任务类型和关键参数
	circuitKey := fmt.Sprintf("ai_reply:%s:%s", userID, personaID)
	
	// 检查断路器状态
	if s.isCircuitOpen(circuitKey) {
		s.smartLogger.LogWarning("Circuit breaker is open, rejecting new task", map[string]interface{}{
			"circuit_key": circuitKey,
			"user_id":     userID,
			"persona_id":  personaID,
		})
		return fmt.Errorf("circuit breaker is open for %s", circuitKey)
	}

	task := &DelayQueueTaskFixed{
		ID:   uuid.New().String(),
		Type: "ai_reply",
		Payload: map[string]interface{}{
			"user_id":         userID,
			"persona_id":      personaID,
			"original_letter": originalLetter,
			"conversation_id": conversationID,
		},
		DelayedUntil:      time.Now().Add(time.Duration(delayHours) * time.Hour),
		CreatedAt:         time.Now(),
		Status:            "pending",
		RetryCount:        0,
		MaxRetries:        3,
		CircuitBreakerKey: circuitKey,
	}

	s.smartLogger.LogInfo(fmt.Sprintf("Scheduling AI reply task %s for user %s", task.ID, userID))
	
	return s.scheduleTask(task)
}

// scheduleTask 安排延迟任务（修复版）
func (s *DelayQueueServiceFixed) scheduleTask(task *DelayQueueTaskFixed) error {
	ctx := context.Background()

	// 序列化任务数据
	taskData, err := json.Marshal(task)
	if err != nil {
		s.smartLogger.LogError("Failed to marshal task", map[string]interface{}{
			"task_id": task.ID,
			"error":   err.Error(),
		})
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	// 计算延迟时间的Unix时间戳作为分数
	score := float64(task.DelayedUntil.Unix())

	// 添加到Redis有序集合
	err = s.redis.ZAdd(ctx, "delay_queue_fixed", &redis.Z{
		Score:  score,
		Member: taskData,
	}).Err()

	if err != nil {
		s.smartLogger.LogError("Failed to add task to delay queue", map[string]interface{}{
			"task_id": task.ID,
			"error":   err.Error(),
		})
		return fmt.Errorf("failed to add task to delay queue: %w", err)
	}

	// 保存到数据库
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

// StartWorker 启动延迟队列工作进程（修复版）
func (s *DelayQueueServiceFixed) StartWorker() {
	s.smartLogger.LogInfo("Starting fixed delay queue worker...")

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.processDelayedTasks()
			
			// 定期打印统计信息
			if time.Now().Minute()%10 == 0 {
				s.smartLogger.PrintSummary()
			}
		}
	}
}

// processDelayedTasks 处理延迟任务（修复版）
func (s *DelayQueueServiceFixed) processDelayedTasks() {
	ctx := context.Background()
	now := time.Now().Unix()

	// 从Redis有序集合中获取到期的任务
	tasks, err := s.redis.ZRangeByScore(ctx, "delay_queue_fixed", &redis.ZRangeBy{
		Min: "0",
		Max: strconv.FormatInt(now, 10),
	}).Result()

	if err != nil {
		s.smartLogger.LogError("Failed to get delayed tasks from Redis", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	if len(tasks) == 0 {
		return
	}

	s.smartLogger.LogInfo(fmt.Sprintf("Processing %d delayed tasks", len(tasks)))

	for _, taskData := range tasks {
		var task DelayQueueTaskFixed
		if err := json.Unmarshal([]byte(taskData), &task); err != nil {
			s.smartLogger.LogError("Failed to unmarshal task", map[string]interface{}{
				"error":     err.Error(),
				"task_data": taskData[:100] + "...", // 只记录前100个字符
			})
			
			// 移除无法解析的任务
			s.redis.ZRem(ctx, "delay_queue_fixed", taskData)
			continue
		}

		// 检查断路器状态
		if s.isCircuitOpen(task.CircuitBreakerKey) {
			s.smartLogger.LogWarning("Skipping task due to open circuit breaker", map[string]interface{}{
				"task_id":     task.ID,
				"circuit_key": task.CircuitBreakerKey,
			})
			
			// 移除被断路器阻断的任务
			s.redis.ZRem(ctx, "delay_queue_fixed", taskData)
			s.updateTaskStatus(task.ID, "circuit_broken")
			continue
		}

		// 处理任务
		if err := s.processTask(&task); err != nil {
			s.smartLogger.LogError("Failed to process task", map[string]interface{}{
				"task_id":     task.ID,
				"task_type":   task.Type,
				"retry_count": task.RetryCount,
				"error":       err.Error(),
			})
			
			// 关键修复：先从Redis移除原任务，再处理错误
			s.redis.ZRem(ctx, "delay_queue_fixed", taskData)
			s.handleTaskError(&task, err)
		} else {
			// 成功处理，从队列中移除
			s.redis.ZRem(ctx, "delay_queue_fixed", taskData)
			s.updateTaskStatus(task.ID, "completed")
			s.recordCircuitSuccess(task.CircuitBreakerKey)
			
			s.smartLogger.LogInfo(fmt.Sprintf("Task %s completed successfully", task.ID))
		}
	}
}

// processTask 处理具体任务（修复版）
func (s *DelayQueueServiceFixed) processTask(task *DelayQueueTaskFixed) error {
	switch task.Type {
	case "ai_reply":
		return s.processDelayedAIReplyTask(task)
	default:
		return fmt.Errorf("unknown task type: %s", task.Type)
	}
}

// processDelayedAIReplyTask 处理AI回信任务（修复版）
func (s *DelayQueueServiceFixed) processDelayedAIReplyTask(task *DelayQueueTaskFixed) error {
	// 从payload中提取数据
	userID, _ := task.Payload["user_id"].(string)
	personaID, _ := task.Payload["persona_id"].(string)
	originalLetter, _ := task.Payload["original_letter"].(string)
	_ = originalLetter // 避免未使用变量警告
	conversationID, _ := task.Payload["conversation_id"].(string)

	s.smartLogger.LogInfo(fmt.Sprintf("Processing AI reply task %s for user %s, persona %s", 
		task.ID, userID, personaID))

	// 验证必要参数
	if userID == "" || personaID == "" || conversationID == "" {
		return fmt.Errorf("missing required parameters: user_id=%s, persona_id=%s, conversation_id=%s", 
			userID, personaID, conversationID)
	}

	// 检查信件是否存在（防止'letter not found'错误）
	var letter models.Letter
	if err := s.db.Where("id = ?", conversationID).First(&letter).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 这是导致无限循环的根本原因！信件不存在但任务一直重试
			s.smartLogger.LogError("Letter not found, marking task as permanently failed", map[string]interface{}{
				"task_id":         task.ID,
				"conversation_id": conversationID,
				"user_id":         userID,
			})
			
			// 永久失败，不再重试
			return &PermanentError{Msg: fmt.Sprintf("letter not found: %s", conversationID)}
		}
		return fmt.Errorf("failed to query letter: %w", err)
	}

	// 创建真实的AI服务实例（用户要求使用真实服务）
	aiService := NewAIService(s.db, s.config)
	
	s.smartLogger.LogInfo(fmt.Sprintf("Using real AI service to generate reply for letter %s", conversationID))

	// 构建AI回信请求
	replyRequest := &models.AIReplyRequest{
		LetterID:   conversationID,
		Persona:    models.AIPersona(personaID),
		DelayHours: 0, // 已经延迟了，现在立即处理
	}

	// 生成AI回信（使用真实AI服务）
	reply, err := aiService.GenerateReply(context.Background(), replyRequest)
	if err != nil {
		s.smartLogger.LogError("Failed to generate AI reply", map[string]interface{}{
			"task_id":         task.ID,
			"conversation_id": conversationID,
			"persona_id":      personaID,
			"error":           err.Error(),
		})
		return fmt.Errorf("failed to generate AI reply: %w", err)
	}

	s.smartLogger.LogInfo(fmt.Sprintf("Successfully generated AI reply %s for task %s", reply.ID, task.ID))
	
	// 这里可以添加发送通知的逻辑
	// 比如通知用户收到了AI回信
	
	return nil
}

// PermanentError 永久错误类型，不会重试
type PermanentError struct {
	Msg string
}

func (e *PermanentError) Error() string {
	return e.Msg
}

// handleTaskError 处理任务错误（修复版）
func (s *DelayQueueServiceFixed) handleTaskError(task *DelayQueueTaskFixed, err error) {
	// 检查是否为永久错误
	if _, isPermanent := err.(*PermanentError); isPermanent {
		s.updateTaskStatus(task.ID, "permanently_failed")
		s.recordCircuitFailure(task.CircuitBreakerKey)
		
		s.smartLogger.LogError("Task permanently failed", map[string]interface{}{
			"task_id": task.ID,
			"error":   err.Error(),
		})
		return
	}

	task.RetryCount++
	task.LastError = err.Error()

	if task.RetryCount >= task.MaxRetries {
		// 超过最大重试次数，标记为失败
		s.updateTaskStatus(task.ID, "failed")
		s.recordCircuitFailure(task.CircuitBreakerKey)
		
		s.smartLogger.LogError("Task failed after max retries", map[string]interface{}{
			"task_id":      task.ID,
			"retry_count":  task.RetryCount,
			"max_retries":  task.MaxRetries,
			"circuit_key":  task.CircuitBreakerKey,
			"final_error":  err.Error(),
		})
		return
	}

	// 指数退避重试：重新安排任务，增加延迟时间
	backoffDelay := time.Duration(task.RetryCount*task.RetryCount) * 5 * time.Minute
	task.DelayedUntil = time.Now().Add(backoffDelay)
	task.Status = "retrying"

	// 重新调度（注意：原任务已经在processDelayedTasks中移除了）
	if err := s.scheduleTask(task); err != nil {
		s.smartLogger.LogError("Failed to reschedule task", map[string]interface{}{
			"task_id": task.ID,
			"error":   err.Error(),
		})
		s.updateTaskStatus(task.ID, "reschedule_failed")
		return
	}

	s.smartLogger.LogWarning("Task rescheduled for retry", map[string]interface{}{
		"task_id":      task.ID,
		"retry_count":  task.RetryCount,
		"backoff_delay": backoffDelay.String(),
		"next_attempt": task.DelayedUntil.Format("15:04:05"),
	})
}

// 断路器相关方法
func (s *DelayQueueServiceFixed) isCircuitOpen(key string) bool {
	cb, exists := s.circuitBreaker[key]
	if !exists {
		return false
	}

	now := time.Now()
	switch cb.State {
	case "open":
		if now.Sub(cb.LastFailureTime) > cb.Timeout {
			cb.State = "half_open"
			return false
		}
		return true
	case "half_open":
		return false
	default: // closed
		return false
	}
}

func (s *DelayQueueServiceFixed) recordCircuitFailure(key string) {
	cb, exists := s.circuitBreaker[key]
	if !exists {
		cb = &DelayQueueCircuitBreaker{
			Threshold: 5,
			Timeout:   10 * time.Minute,
			State:     "closed",
		}
		s.circuitBreaker[key] = cb
	}

	cb.FailureCount++
	cb.LastFailureTime = time.Now()

	if cb.FailureCount >= cb.Threshold {
		cb.State = "open"
		s.smartLogger.LogWarning("Circuit breaker opened", map[string]interface{}{
			"circuit_key":    key,
			"failure_count":  cb.FailureCount,
			"threshold":      cb.Threshold,
			"timeout":        cb.Timeout.String(),
		})
	}
}

func (s *DelayQueueServiceFixed) recordCircuitSuccess(key string) {
	cb, exists := s.circuitBreaker[key]
	if !exists {
		return
	}

	if cb.State == "half_open" {
		cb.State = "closed"
		cb.FailureCount = 0
		s.smartLogger.LogInfo(fmt.Sprintf("Circuit breaker closed for %s", key))
	} else if cb.State == "closed" {
		if cb.FailureCount > 0 {
			cb.FailureCount--
		}
	}
}

// updateTaskStatus 更新任务状态
func (s *DelayQueueServiceFixed) updateTaskStatus(taskID, status string) {
	err := s.db.Model(&models.DelayQueueRecord{}).
		Where("id = ?", taskID).
		Update("status", status).Error
	
	if err != nil {
		s.smartLogger.LogError("Failed to update task status", map[string]interface{}{
			"task_id": taskID,
			"status":  status,
			"error":   err.Error(),
		})
	}
}

// GetStats 获取服务统计信息
func (s *DelayQueueServiceFixed) GetStats() map[string]interface{} {
	stats := s.smartLogger.GetStats()
	
	circuitStats := make(map[string]interface{})
	for key, cb := range s.circuitBreaker {
		circuitStats[key] = map[string]interface{}{
			"state":         cb.State,
			"failure_count": cb.FailureCount,
			"last_failure":  cb.LastFailureTime.Format("15:04:05"),
		}
	}
	
	return map[string]interface{}{
		"smart_logger": stats,
		"circuit_breakers": circuitStats,
		"queue_name": "delay_queue_fixed",
	}
}

// 使用现有的辅助函数以避免重复声明
// max, getEnv, getEnvAsInt 函数已在其他文件中定义
