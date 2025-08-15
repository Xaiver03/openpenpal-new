package services

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"openpenpal-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CreditTaskService 积分任务服务 - 模块化积分奖励系统
type CreditTaskService struct {
	db           *gorm.DB
	creditSvc    *CreditService
	limiterSvc   *CreditLimiterService
	workerPool   chan struct{} // 工作池，控制并发执行数量
}

// NewCreditTaskService 创建积分任务服务
func NewCreditTaskService(db *gorm.DB, creditSvc *CreditService, limiterSvc *CreditLimiterService) *CreditTaskService {
	service := &CreditTaskService{
		db:         db,
		creditSvc:  creditSvc,
		limiterSvc: limiterSvc,
		workerPool: make(chan struct{}, 10), // 最多10个并发任务
	}
	
	// 启动任务处理器
	go service.taskProcessor()
	
	return service
}

// CreateTask 创建积分任务 (with limit checking)
func (s *CreditTaskService) CreateTask(taskType models.CreditTaskType, userID string, points int, description, reference string) (*models.CreditTask, error) {
	return s.CreateTaskWithMetadata(taskType, userID, points, description, reference, nil)
}

// CreateTaskWithMetadata 创建积分任务（带元数据）
func (s *CreditTaskService) CreateTaskWithMetadata(taskType models.CreditTaskType, userID string, points int, description, reference string, metadata map[string]string) (*models.CreditTask, error) {
	// 1. 检查用户是否被封禁
	if s.limiterSvc != nil {
		if blocked, err := s.limiterSvc.IsUserBlocked(userID); err != nil {
			log.Printf("Failed to check user block status: %v", err)
		} else if blocked {
			return nil, fmt.Errorf("user is blocked from earning credits")
		}
	}

	// 2. 检查积分限制
	actionType := string(taskType)
	if s.limiterSvc != nil {
		limitStatus, err := s.limiterSvc.CheckLimit(userID, actionType, points)
		if err != nil {
			log.Printf("Failed to check limit: %v", err)
		} else if limitStatus.IsLimited {
			return nil, fmt.Errorf("credit limit exceeded: %s (current: %d/%d)", 
				limitStatus.Period, limitStatus.CurrentCount, limitStatus.MaxCount)
		}
	}

	// 3. 检查作弊风险
	if s.limiterSvc != nil && metadata != nil {
		if alert, err := s.limiterSvc.DetectAnomalous(userID, actionType, metadata); err != nil {
			log.Printf("Failed to detect fraud: %v", err)
		} else if alert != nil && alert.Severity == models.SeverityHigh {
			// 高风险行为，暂时拒绝
			log.Printf("High risk behavior detected for user %s: %s", userID, alert.Description)
			return nil, fmt.Errorf("suspicious activity detected, please try again later")
		}
	}

	// 4. 创建任务
	task := &models.CreditTask{
		ID:          uuid.New().String(),
		TaskType:    taskType,
		UserID:      userID,
		Status:      models.TaskStatusPending,
		Points:      points,
		Description: description,
		Reference:   reference,
		Priority:    s.getTaskPriority(taskType),
		MaxAttempts: 3,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	// 检查任务规则和限制
	if err := s.validateTaskConstraints(task); err != nil {
		return nil, fmt.Errorf("task validation failed: %w", err)
	}
	
	if err := s.db.Create(task).Error; err != nil {
		return nil, fmt.Errorf("failed to create credit task: %w", err)
	}
	
	log.Printf("Created credit task: %s for user %s (%s, %d points)", 
		task.ID, userID, taskType, points)
	
	return task, nil
}

// ExecuteTask 执行积分任务
func (s *CreditTaskService) ExecuteTask(taskID string) error {
	var task models.CreditTask
	if err := s.db.Where("id = ?", taskID).First(&task).Error; err != nil {
		return fmt.Errorf("task not found: %w", err)
	}
	
	if !task.ShouldExecuteNow() {
		return fmt.Errorf("task cannot be executed now: %s", task.Status)
	}
	
	// 更新状态为执行中
	task.Status = models.TaskStatusExecuting
	task.Attempts++
	now := time.Now()
	task.ExecutedAt = &now
	task.UpdatedAt = now
	
	if err := s.db.Save(&task).Error; err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}
	
	// 执行积分奖励
	err := s.executeReward(&task)
	
	if err != nil {
		// 标记失败
		task.Status = models.TaskStatusFailed
		task.ErrorMessage = err.Error()
		failedAt := time.Now()
		task.FailedAt = &failedAt
	} else {
		// 标记完成
		task.Status = models.TaskStatusCompleted
		completedAt := time.Now()
		task.CompletedAt = &completedAt
	}
	
	task.UpdatedAt = time.Now()
	if saveErr := s.db.Save(&task).Error; saveErr != nil {
		log.Printf("Failed to save task completion status: %v", saveErr)
	}
	
	return err
}

// executeReward 执行具体的积分奖励
func (s *CreditTaskService) executeReward(task *models.CreditTask) error {
	// 1. 使用新的限制系统检查（如果可用）
	if s.limiterSvc != nil {
		limitStatus, err := s.limiterSvc.CheckLimit(task.UserID, string(task.TaskType), task.Points)
		if err != nil {
			log.Printf("Failed to check limit in executeReward: %v", err)
		} else if limitStatus.IsLimited {
			return fmt.Errorf("credit limit exceeded during execution: %s", limitStatus.Period)
		}
	} else {
		// 2. 回退到旧的每日限制检查
		canExecute, err := s.creditSvc.CheckDailyLimit(task.UserID, string(task.TaskType))
		if err != nil {
			return fmt.Errorf("failed to check daily limit: %w", err)
		}
		if !canExecute {
			return fmt.Errorf("daily limit exceeded for task type: %s", task.TaskType)
		}
	}
	
	// 根据任务类型执行相应的奖励方法
	switch task.TaskType {
	case models.TaskTypeLetterCreated:
		return s.creditSvc.RewardLetterCreated(task.UserID, task.Reference)
	case models.TaskTypeLetterGenerated:
		return s.creditSvc.RewardLetterGenerated(task.UserID, task.Reference)
	case models.TaskTypeLetterDelivered:
		return s.creditSvc.RewardLetterDelivered(task.UserID, task.Reference)
	case models.TaskTypeLetterRead:
		return s.creditSvc.RewardLetterRead(task.UserID, task.Reference)
	case models.TaskTypeReceiveLetter:
		return s.creditSvc.RewardReceiveLetter(task.UserID, task.Reference)
	case models.TaskTypePublicLetterLike:
		return s.creditSvc.RewardPublicLetterLike(task.UserID, task.Reference)
	case models.TaskTypeWritingChallenge:
		return s.creditSvc.RewardWritingChallenge(task.UserID, task.Reference)
	case models.TaskTypeAIInteraction:
		return s.creditSvc.RewardAIInteraction(task.UserID, task.Reference)
	case models.TaskTypeCourierFirstTask:
		return s.creditSvc.RewardCourierFirstTask(task.UserID, task.Reference)
	case models.TaskTypeCourierDelivery:
		return s.creditSvc.RewardCourierDelivery(task.UserID, task.Reference)
	case models.TaskTypeMuseumSubmit:
		return s.creditSvc.RewardMuseumSubmit(task.UserID, task.Reference)
	case models.TaskTypeMuseumApproved:
		return s.creditSvc.RewardMuseumApproved(task.UserID, task.Reference)
	case models.TaskTypeMuseumLiked:
		return s.creditSvc.RewardMuseumLiked(task.UserID, task.Reference)
	case models.TaskTypeOPCodeApproval:
		return s.creditSvc.RewardOPCodeApproval(task.UserID, task.Reference)
	case models.TaskTypeCommunityBadge:
		return s.creditSvc.RewardCommunityBadge(task.UserID, task.Reference)
	case models.TaskTypeAdminReward:
		return s.creditSvc.RewardAdminCustom(task.UserID, task.Points, task.Description, task.Reference)
	default:
		return fmt.Errorf("unknown task type: %s", task.TaskType)
	}
	
	// 3. 记录用户行为到限制系统（在奖励执行成功后）
	if s.limiterSvc != nil {
		metadata := map[string]string{
			"reference": task.Reference,
			"task_id":   task.ID,
		}
		
		if err := s.limiterSvc.RecordAction(task.UserID, string(task.TaskType), task.Points, metadata); err != nil {
			log.Printf("Failed to record action to limiter: %v", err)
			// 不返回错误，因为积分已经发放成功
		}
	}
	
	return nil
}

// taskProcessor 任务处理器 - 后台处理积分任务
func (s *CreditTaskService) taskProcessor() {
	ticker := time.NewTicker(30 * time.Second) // 每30秒检查一次
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			s.processPendingTasks()
		case s.workerPool <- struct{}{}: // 获取工作许可
			go func() {
				defer func() { <-s.workerPool }() // 释放工作许可
				s.processNextTask()
			}()
		}
	}
}

// processPendingTasks 处理待执行的任务
func (s *CreditTaskService) processPendingTasks() {
	var tasks []models.CreditTask
	err := s.db.Where("status IN (?, ?) AND (scheduled_at IS NULL OR scheduled_at <= ?)", 
		models.TaskStatusPending, models.TaskStatusScheduled, time.Now()).
		Order("priority DESC, created_at ASC").
		Limit(20).
		Find(&tasks).Error
	
	if err != nil {
		log.Printf("Failed to query pending tasks: %v", err)
		return
	}
	
	for _, task := range tasks {
		select {
		case s.workerPool <- struct{}{}:
			go func(t models.CreditTask) {
				defer func() { <-s.workerPool }()
				if err := s.ExecuteTask(t.ID); err != nil {
					log.Printf("Failed to execute task %s: %v", t.ID, err)
				}
			}(task)
		default:
			// 工作池满了，跳过这次处理
			break
		}
	}
}

// processNextTask 处理下一个任务
func (s *CreditTaskService) processNextTask() {
	var task models.CreditTask
	err := s.db.Where("status IN (?, ?) AND (scheduled_at IS NULL OR scheduled_at <= ?)", 
		models.TaskStatusPending, models.TaskStatusScheduled, time.Now()).
		Order("priority DESC, created_at ASC").
		First(&task).Error
	
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			log.Printf("Failed to query next task: %v", err)
		}
		return
	}
	
	if err := s.ExecuteTask(task.ID); err != nil {
		log.Printf("Failed to execute task %s: %v", task.ID, err)
	}
}

// validateTaskConstraints 验证任务约束
func (s *CreditTaskService) validateTaskConstraints(task *models.CreditTask) error {
	// 检查任务规则
	var rule models.CreditTaskRule
	err := s.db.Where("task_type = ? AND is_active = true", task.TaskType).First(&rule).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to query task rule: %w", err)
	}
	
	if err == nil {
		// 应用规则配置
		if rule.DailyLimit > 0 {
			task.DailyLimit = rule.DailyLimit
		}
		if rule.WeeklyLimit > 0 {
			task.WeeklyLimit = rule.WeeklyLimit
		}
		task.Constraints = rule.Constraints
		
		// 检查每日限制
		if rule.DailyLimit > 0 {
			canExecute, err := s.creditSvc.CheckDailyLimit(task.UserID, string(task.TaskType))
			if err != nil {
				return fmt.Errorf("failed to check daily limit: %w", err)
			}
			if !canExecute {
				return fmt.Errorf("daily limit exceeded for task type: %s", task.TaskType)
			}
		}
	}
	
	return nil
}

// getTaskPriority 获取任务优先级
func (s *CreditTaskService) getTaskPriority(taskType models.CreditTaskType) int {
	priorities := map[models.CreditTaskType]int{
		// 高优先级：时效性强的任务
		models.TaskTypeCourierDelivery:  8,
		models.TaskTypeLetterDelivered: 7,
		models.TaskTypeLetterRead:      6,
		
		// 中优先级：用户行为奖励
		models.TaskTypeLetterCreated:    5,
		models.TaskTypeReceiveLetter:    5,
		models.TaskTypePublicLetterLike: 4,
		models.TaskTypeAIInteraction:    4,
		
		// 低优先级：系统奖励
		models.TaskTypeMuseumSubmit:     3,
		models.TaskTypeMuseumApproved:   3,
		models.TaskTypeWritingChallenge: 2,
		models.TaskTypeOPCodeApproval:   2,
		models.TaskTypeCommunityBadge:   1,
		models.TaskTypeAdminReward:      1,
	}
	
	if priority, exists := priorities[taskType]; exists {
		return priority
	}
	return 0 // 默认优先级
}

// GetTaskStatistics 获取任务统计信息
func (s *CreditTaskService) GetTaskStatistics(timeRange string) ([]models.CreditTaskStatistics, error) {
	var stats []models.CreditTaskStatistics
	
	var timeCondition string
	switch timeRange {
	case "today":
		timeCondition = "DATE(created_at) = CURRENT_DATE"
	case "week":
		timeCondition = "created_at >= DATE_TRUNC('week', CURRENT_DATE)"
	case "month":
		timeCondition = "created_at >= DATE_TRUNC('month', CURRENT_DATE)"
	default:
		timeCondition = "1=1" // 全部时间
	}
	
	query := `
		SELECT 
			task_type,
			COUNT(*) as total_tasks,
			COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_tasks,
			COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed_tasks,
			COALESCE(SUM(CASE WHEN status = 'completed' THEN points ELSE 0 END), 0) as total_points,
			ROUND(
				COUNT(CASE WHEN status = 'completed' THEN 1 END)::FLOAT / 
				NULLIF(COUNT(*), 0) * 100, 2
			) as success_rate,
			ROUND(
				AVG(CASE WHEN status = 'completed' THEN points ELSE NULL END), 2
			) as avg_points
		FROM credit_tasks 
		WHERE ` + timeCondition + `
		GROUP BY task_type
		ORDER BY total_tasks DESC
	`
	
	if err := s.db.Raw(query).Scan(&stats).Error; err != nil {
		return nil, fmt.Errorf("failed to get task statistics: %w", err)
	}
	
	return stats, nil
}

// RetryFailedTasks 重试失败的任务
func (s *CreditTaskService) RetryFailedTasks(maxAge time.Duration) (int, error) {
	cutoffTime := time.Now().Add(-maxAge)
	
	var tasks []models.CreditTask
	err := s.db.Where("status = ? AND can_retry = true AND failed_at > ? AND attempts < max_attempts", 
		models.TaskStatusFailed, cutoffTime).Find(&tasks).Error
	
	if err != nil {
		return 0, fmt.Errorf("failed to query failed tasks: %w", err)
	}
	
	retriedCount := 0
	for _, task := range tasks {
		if task.CanRetry() {
			task.Status = models.TaskStatusPending
			task.ErrorMessage = ""
			task.FailedAt = nil
			task.UpdatedAt = time.Now()
			
			if err := s.db.Save(&task).Error; err != nil {
				log.Printf("Failed to reset task %s for retry: %v", task.ID, err)
				continue
			}
			retriedCount++
		}
	}
	
	return retriedCount, nil
}

// CreateBatchTasks 批量创建积分任务
func (s *CreditTaskService) CreateBatchTasks(batchName string, taskType models.CreditTaskType, userIDs []string, points int, description string) (*models.CreditTaskBatch, error) {
	batch := &models.CreditTaskBatch{
		ID:         uuid.New().String(),
		BatchName:  batchName,
		TaskType:   taskType,
		Status:     models.TaskStatusPending,
		TotalTasks: len(userIDs),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	
	tx := s.db.Begin()
	
	// 创建批次记录
	if err := tx.Create(batch).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create batch: %w", err)
	}
	
	// 创建任务
	for _, userID := range userIDs {
		task := &models.CreditTask{
			ID:          uuid.New().String(),
			TaskType:    taskType,
			UserID:      userID,
			Status:      models.TaskStatusPending,
			Points:      points,
			Description: fmt.Sprintf("[批次: %s] %s", batchName, description),
			Reference:   batch.ID,
			Priority:    s.getTaskPriority(taskType),
			MaxAttempts: 3,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		
		if err := tx.Create(task).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to create task for user %s: %w", userID, err)
		}
	}
	
	tx.Commit()
	
	log.Printf("Created batch %s with %d tasks", batchName, len(userIDs))
	return batch, nil
}

// GetUserTasks 获取用户的积分任务列表
func (s *CreditTaskService) GetUserTasks(userID string, status models.CreditTaskStatus, limit, offset int) ([]models.CreditTask, int64, error) {
	var tasks []models.CreditTask
	var total int64
	
	query := s.db.Model(&models.CreditTask{}).Where("user_id = ?", userID)
	
	if status != "" {
		query = query.Where("status = ?", status)
	}
	
	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count user tasks: %w", err)
	}
	
	// 获取分页数据
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&tasks).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get user tasks: %w", err)
	}
	
	return tasks, total, nil
}

// =============== 便利方法：快速创建常见任务 ===============

// TriggerLetterCreatedReward 触发写信奖励
func (s *CreditTaskService) TriggerLetterCreatedReward(userID, letterID string) error {
	_, err := s.CreateTask(models.TaskTypeLetterCreated, userID, PointsLetterCreated, "成功写信并绑定条码", letterID)
	return err
}

// TriggerCourierDeliveryReward 触发信使送达奖励
func (s *CreditTaskService) TriggerCourierDeliveryReward(userID, taskID string) error {
	_, err := s.CreateTask(models.TaskTypeCourierDelivery, userID, PointsCourierDelivery, "信使成功送达一封信", taskID)
	return err
}

// TriggerPublicLetterLikeReward 触发公开信点赞奖励
func (s *CreditTaskService) TriggerPublicLetterLikeReward(userID, letterID string) error {
	_, err := s.CreateTask(models.TaskTypePublicLetterLike, userID, PointsPublicLetterLike, "公开信被点赞", letterID)
	return err
}

// TriggerAIInteractionReward 触发AI互动奖励
func (s *CreditTaskService) TriggerAIInteractionReward(userID, sessionID string) error {
	_, err := s.CreateTask(models.TaskTypeAIInteraction, userID, PointsAIInteraction, "使用AI笔友并留下评价", sessionID)
	return err
}

// TriggerMuseumSubmitReward 触发博物馆提交奖励
func (s *CreditTaskService) TriggerMuseumSubmitReward(userID, itemID string) error {
	_, err := s.CreateTask(models.TaskTypeMuseumSubmit, userID, PointsMuseumSubmit, "提交作品到博物馆", itemID)
	return err
}

// TriggerMuseumApprovedReward 触发博物馆审核通过奖励
func (s *CreditTaskService) TriggerMuseumApprovedReward(userID, itemID string) error {
	_, err := s.CreateTask(models.TaskTypeMuseumApproved, userID, PointsMuseumApproved, "博物馆作品审核通过", itemID)
	return err
}

// TriggerMuseumLikedReward 触发博物馆作品被点赞奖励
func (s *CreditTaskService) TriggerMuseumLikedReward(userID, itemID string) error {
	_, err := s.CreateTask(models.TaskTypeMuseumLiked, userID, PointsMuseumLiked, "博物馆作品获得点赞", itemID)
	return err
}