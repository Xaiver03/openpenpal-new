package services

import (
	"context"
	"courier-service/internal/models"
	"courier-service/internal/utils"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// TaskService 任务服务
type TaskService struct {
	db        *gorm.DB
	redis     *redis.Client
	wsManager *utils.WebSocketManager
}

// NewTaskService 创建任务服务实例
func NewTaskService(db *gorm.DB, redis *redis.Client, wsManager *utils.WebSocketManager) *TaskService {
	return &TaskService{
		db:        db,
		redis:     redis,
		wsManager: wsManager,
	}
}

// CreateTask 创建任务
func (s *TaskService) CreateTask(letterID, pickupLocation, deliveryLocation string, queueService *QueueService) (*models.Task, error) {
	task := &models.Task{
		TaskID:           utils.GenerateTaskID(),
		LetterID:         letterID,
		PickupLocation:   pickupLocation,
		DeliveryLocation: deliveryLocation,
		Status:           models.TaskStatusAvailable,
		Priority:         models.TaskPriorityNormal,
		Reward:           5.0, // 默认奖励
	}

	if err := s.db.Create(task).Error; err != nil {
		return nil, err
	}

	// 推送任务到队列
	if queueService != nil {
		queueService.PushTaskToQueue(task)
	}

	return task, nil
}

// GetAvailableTasks 获取可用任务列表
func (s *TaskService) GetAvailableTasks(query *models.TaskQuery) ([]models.Task, int64, error) {
	var tasks []models.Task
	var total int64

	db := s.db.Model(&models.Task{})

	// 应用筛选条件
	if query.Zone != "" {
		// 这里可以根据区域筛选，需要地理位置匹配
		// 暂时简单处理
	}

	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	} else {
		db = db.Where("status = ?", models.TaskStatusAvailable)
	}

	if query.Priority != "" {
		db = db.Where("priority = ?", query.Priority)
	}

	if query.CourierID != "" {
		db = db.Where("courier_id = ?", query.CourierID)
	}

	// 获取总数
	db.Count(&total)

	// 分页查询
	err := db.Order("priority DESC, created_at ASC").
		Limit(query.Limit).
		Offset(query.Offset).
		Find(&tasks).Error

	return tasks, total, err
}

// AcceptTask 接受任务
func (s *TaskService) AcceptTask(taskID, courierID string, request *models.TaskAcceptRequest) (*models.Task, error) {
	var task models.Task
	if err := s.db.Where("task_id = ?", taskID).First(&task).Error; err != nil {
		return nil, err
	}

	// 检查任务状态
	if !task.CanTransitionTo(models.TaskStatusAccepted) {
		return nil, gorm.ErrInvalidTransaction
	}

	// 检查信使是否可以接受任务
	var courier models.Courier
	if err := s.db.Where("user_id = ?", courierID).First(&courier).Error; err != nil {
		return nil, err
	}

	if !courier.CanAcceptTask() {
		return nil, gorm.ErrInvalidValue
	}

	// 更新任务状态
	now := time.Now()
	deadline := now.Add(4 * time.Hour) // 默认4小时截止

	updates := map[string]interface{}{
		"courier_id":     &courierID,
		"status":         models.TaskStatusAccepted,
		"estimated_time": request.EstimatedTime,
		"accepted_at":    &now,
		"deadline":       &deadline,
	}

	if err := s.db.Model(&task).Where("task_id = ?", taskID).Updates(updates).Error; err != nil {
		return nil, err
	}

	// 重新获取更新后的任务
	s.db.Where("task_id = ?", taskID).First(&task)

	// 发送WebSocket通知
	s.wsManager.SendTaskUpdate(taskID, models.TaskStatusAccepted, courierID)
	s.wsManager.SendTaskAssignment(taskID, courierID, task)

	return &task, nil
}

// UpdateTaskStatus 更新任务状态（通过扫码）
func (s *TaskService) UpdateTaskStatus(letterCode, courierID string, scanRequest *models.ScanRequest) (*models.ScanResponse, error) {
	// 根据信件编号查找任务
	var task models.Task
	if err := s.db.Where("letter_id = ?", letterCode).First(&task).Error; err != nil {
		return nil, err
	}

	// 验证信使权限
	if task.CourierID == nil || *task.CourierID != courierID {
		return nil, gorm.ErrInvalidValue
	}

	// 获取目标状态
	targetStatus, exists := models.ActionToStatus[scanRequest.Action]
	if !exists {
		return nil, gorm.ErrInvalidValue
	}

	// 检查状态转换是否有效
	if !task.CanTransitionTo(targetStatus) {
		return nil, gorm.ErrInvalidTransaction
	}

	oldStatus := task.Status

	// 更新任务状态
	updates := map[string]interface{}{
		"status": targetStatus,
	}

	if targetStatus == models.TaskStatusDelivered {
		now := time.Now()
		updates["completed_at"] = &now
	}

	if err := s.db.Model(&task).Where("letter_id = ?", letterCode).Updates(updates).Error; err != nil {
		return nil, err
	}

	// 创建扫码记录
	scanRecord := &models.ScanRecord{
		TaskID:    task.TaskID,
		CourierID: courierID,
		LetterID:  letterCode,
		Action:    scanRequest.Action,
		Location:  scanRequest.Location,
		Latitude:  scanRequest.Latitude,
		Longitude: scanRequest.Longitude,
		Note:      scanRequest.Note,
		PhotoURL:  scanRequest.PhotoURL,
		Timestamp: time.Now(),
	}

	if err := s.db.Create(scanRecord).Error; err != nil {
		log.Printf("Failed to create scan record: %v", err)
	}

	// 发送WebSocket通知
	s.wsManager.SendTaskUpdate(task.TaskID, targetStatus, courierID)

	response := &models.ScanResponse{
		LetterID:  letterCode,
		OldStatus: oldStatus,
		NewStatus: targetStatus,
		ScanTime:  scanRecord.Timestamp,
		Location:  scanRequest.Location,
	}

	return response, nil
}

// GetTaskByID 根据任务ID获取任务
func (s *TaskService) GetTaskByID(taskID string) (*models.Task, error) {
	var task models.Task
	err := s.db.Where("task_id = ?", taskID).First(&task).Error
	return &task, err
}

// pushTaskToQueue 将任务推送到Redis队列
func (s *TaskService) pushTaskToQueue(task *models.Task) {
	ctx := context.Background()
	taskJSON, err := json.Marshal(task)
	if err != nil {
		log.Printf("Failed to marshal task: %v", err)
		return
	}

	queueName := "tasks:normal"
	if task.Priority == models.TaskPriorityUrgent {
		queueName = "tasks:urgent"
	} else if task.Priority == models.TaskPriorityExpress {
		queueName = "tasks:express"
	}

	err = s.redis.LPush(ctx, queueName, taskJSON).Err()
	if err != nil {
		log.Printf("Failed to push task to queue: %v", err)
	}
}

// ConsumeTaskQueue 消费任务队列
func (s *TaskService) ConsumeTaskQueue() {
	ctx := context.Background()
	
	for {
		// 按优先级处理队列：express > urgent > normal
		result, err := s.redis.BRPop(ctx, 1*time.Second, 
			"tasks:express", "tasks:urgent", "tasks:normal").Result()
		
		if err != nil {
			if err != redis.Nil {
				log.Printf("Redis BRPop error: %v", err)
			}
			continue
		}

		var task models.Task
		if err := json.Unmarshal([]byte(result[1]), &task); err != nil {
			log.Printf("Failed to unmarshal task: %v", err)
			continue
		}

		// 处理任务（这里可以添加自动分配逻辑）
		s.processTask(&task)
	}
}

// processTask 处理任务（自动分配等）- 增强OP Code支持
func (s *TaskService) processTask(task *models.Task) {
	// 这里可以实现自动任务分配逻辑
	log.Printf("Processing task: %s for letter: %s", task.TaskID, task.LetterID)
	
	// FSD条码系统增强处理
	// 1. 根据OP Code查找合适的区域信使
	// 2. 根据信使级别进行任务分配
	// 3. 发送推送通知给适格的信使
	// 4. 设置适当的任务优先级和奖励
	
	// TODO: 实现基于OP Code的智能任务分配算法
	// - 解析取件地和送达地的OP Code
	// - 查找具有对应权限的信使
	// - 考虑地理距离和信使负载
	// - 自动分配或推送通知
}

// ========================= FSD条码系统支持方法 =========================

// validateOPCodePermission 验证OP Code权限
func (s *TaskService) validateOPCodePermission(recipientOPCode, operatorOPCode string, scannerLevel int) bool {
	if len(recipientOPCode) != 6 || len(operatorOPCode) != 6 {
		return false
	}
	
	// 提取学校代码（前2位）和区域代码（中2位）
	recipientSchool := recipientOPCode[:2]
	recipientArea := recipientOPCode[2:4]
	operatorSchool := operatorOPCode[:2]
	operatorArea := operatorOPCode[2:4]
	
	// 根据扫码员级别验证权限
	switch scannerLevel {
	case 4: // 城市总代 - 可以操作任意区域
		return true
	case 3: // 校级信使 - 只能操作同学校
		return recipientSchool == operatorSchool
	case 2: // 片区信使 - 只能操作同学校同区域
		return recipientSchool == operatorSchool && recipientArea == operatorArea
	case 1: // 楼栋信使 - 只能操作精确匹配的前4位
		return recipientOPCode[:4] == operatorOPCode[:4]
	default:
		return false
	}
}

// getNextAction 获取下一步操作建议
func (s *TaskService) getNextAction(currentStatus string) string {
	nextActions := map[string]string{
		models.TaskStatusAccepted:  "请前往取件地点收取信件",
		models.TaskStatusCollected: "请将信件运送至目的地",
		models.TaskStatusInTransit: "请将信件送达收件人",
		models.TaskStatusDelivered: "任务已完成，感谢您的服务！",
		models.TaskStatusFailed:    "请联系客服处理失败情况",
	}
	
	if action, exists := nextActions[currentStatus]; exists {
		return action
	}
	return "请继续按照流程操作"
}

// calculateEstimatedDelivery 计算预计送达时间
func (s *TaskService) calculateEstimatedDelivery(currentStatus string, taskCreatedAt time.Time) *time.Time {
	now := time.Now()
	
	switch currentStatus {
	case models.TaskStatusAccepted:
		// 已接取，预计4小时内完成
		estimated := now.Add(4 * time.Hour)
		return &estimated
	case models.TaskStatusCollected:
		// 已收取，预计2小时内送达
		estimated := now.Add(2 * time.Hour)
		return &estimated
	case models.TaskStatusInTransit:
		// 投递中，预计1小时内送达
		estimated := now.Add(1 * time.Hour)
		return &estimated
	case models.TaskStatusDelivered:
		// 已送达，返回实际送达时间
		return &now
	default:
		return nil
	}
}

// getCourierPermissions 获取信使权限
func (s *TaskService) getCourierPermissions(courierID string, scannerLevel int) map[string]bool {
	permissions := map[string]bool{
		"can_scan":           true,
		"can_collect":        scannerLevel >= 1,
		"can_deliver":        scannerLevel >= 1,
		"can_cross_building": scannerLevel >= 2,
		"can_cross_area":     scannerLevel >= 3,
		"can_cross_school":   scannerLevel >= 4,
		"can_manage_tasks":   scannerLevel >= 3,
		"can_assign_tasks":   scannerLevel >= 4,
	}
	
	return permissions
}

// ScanBarcode 扫码接口 - 兼容旧版本同时支持FSD增强功能
func (s *TaskService) ScanBarcode(barcodeCode, courierID string, scanRequest *models.ScanRequest) (*models.ScanResponse, error) {
	// 如果请求中没有条码编号，使用参数中的
	if scanRequest.BarcodeCode == "" {
		scanRequest.BarcodeCode = barcodeCode
	}
	
	return s.UpdateTaskStatus(barcodeCode, courierID, scanRequest)
}

// GetScanHistory 获取扫码历史
func (s *TaskService) GetScanHistory(letterID string, limit int) ([]models.ScanRecord, error) {
	var records []models.ScanRecord
	
	query := s.db.Where("letter_id = ? OR barcode_code = ?", letterID, letterID).
		Order("timestamp DESC")
		
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	err := query.Find(&records).Error
	return records, err
}

// ValidateBarcodeAccess 验证条码访问权限
func (s *TaskService) ValidateBarcodeAccess(barcodeCode, courierID string, action string) error {
	// 查找任务
	var task models.Task
	if err := s.db.Where("letter_id = ?", barcodeCode).First(&task).Error; err != nil {
		return fmt.Errorf("任务不存在: %w", err)
	}
	
	// 检查是否已分配给当前信使
	if task.CourierID == nil || *task.CourierID != courierID {
		return fmt.Errorf("任务未分配给当前信使")
	}
	
	// 检查状态转换是否合法
	targetStatus, exists := models.ActionToStatus[action]
	if !exists {
		return fmt.Errorf("无效的操作: %s", action)
	}
	
	if !task.CanTransitionTo(targetStatus) {
		return fmt.Errorf("无效的状态转换: %s -> %s", task.Status, targetStatus)
	}
	
	return nil
}