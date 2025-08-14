package services

import (
	"courier-service/internal/models"
	"courier-service/internal/utils"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// HierarchicalAssignmentService 层级化任务分配服务
type HierarchicalAssignmentService struct {
	db                *gorm.DB
	assignmentService *AssignmentService
	hierarchyService  *HierarchyService
	wsManager         *utils.WebSocketManager
}

// NewHierarchicalAssignmentService 创建层级化任务分配服务
func NewHierarchicalAssignmentService(
	db *gorm.DB,
	assignmentService *AssignmentService,
	hierarchyService *HierarchyService,
	wsManager *utils.WebSocketManager,
) *HierarchicalAssignmentService {
	return &HierarchicalAssignmentService{
		db:                db,
		assignmentService: assignmentService,
		hierarchyService:  hierarchyService,
		wsManager:         wsManager,
	}
}

// AssignTaskByHierarchy 根据层级结构分配任务
func (s *HierarchicalAssignmentService) AssignTaskByHierarchy(managerID string, req *models.HierarchicalTaskAssignmentRequest) (*models.Task, error) {
	// 获取管理者信息
	var manager models.Courier
	if err := s.db.First(&manager, managerID).Error; err != nil {
		return nil, fmt.Errorf("管理者不存在: %w", err)
	}

	// 获取任务信息
	var task models.Task
	if err := s.db.Where("task_id = ?", req.TaskID).First(&task).Error; err != nil {
		return nil, fmt.Errorf("任务不存在: %w", err)
	}

	// 检查任务状态
	if task.Status != models.TaskStatusAvailable {
		return nil, errors.New("任务不可分配")
	}

	// 根据分配类型执行不同的分配策略
	switch req.AssignmentType {
	case "direct":
		return s.directAssignment(&manager, &task, req)
	case "cascade":
		return s.cascadeAssignment(&manager, &task, req)
	case "auto_hierarchy":
		return s.autoHierarchyAssignment(&manager, &task, req)
	default:
		return nil, errors.New("不支持的分配类型")
	}
}

// directAssignment 直接分配给指定信使
func (s *HierarchicalAssignmentService) directAssignment(manager *models.Courier, task *models.Task, req *models.HierarchicalTaskAssignmentRequest) (*models.Task, error) {
	// 获取目标信使信息
	var targetCourier models.Courier
	if err := s.db.First(&targetCourier, req.TargetCourierID).Error; err != nil {
		return nil, fmt.Errorf("目标信使不存在: %w", err)
	}

	// 检查管理权限
	if !s.canAssignToSubordinate(manager, &targetCourier) {
		return nil, errors.New("无权限分配任务给该信使")
	}

	// 检查信使是否有处理该区域任务的权限
	if !s.canHandleTaskInZone(&targetCourier, task) {
		return nil, errors.New("目标信使无权限处理该区域的任务")
	}

	// 执行分配
	err := s.assignTaskToCourier(task, &targetCourier, manager.ID)
	if err != nil {
		return nil, fmt.Errorf("分配任务失败: %w", err)
	}

	return task, nil
}

// cascadeAssignment 级联分配（向下级分配）
func (s *HierarchicalAssignmentService) cascadeAssignment(manager *models.Courier, task *models.Task, _ *models.HierarchicalTaskAssignmentRequest) (*models.Task, error) {
	// 获取所有下级信使
	subordinates, err := s.hierarchyService.GetSubordinates(manager.ID)
	if err != nil {
		return nil, fmt.Errorf("获取下级信使失败: %w", err)
	}

	if len(subordinates) == 0 {
		return nil, errors.New("没有可分配的下级信使")
	}

	// 筛选能够处理该区域任务的信使
	var validCouriers []models.Courier
	for _, subordinate := range subordinates {
		if s.canHandleTaskInZone(&subordinate, task) {
			validCouriers = append(validCouriers, subordinate)
		}
	}

	if len(validCouriers) == 0 {
		return nil, errors.New("没有可处理该区域任务的下级信使")
	}

	// 选择最优信使（基于评分）
	bestCourier := s.selectBestCourierForTask(validCouriers, task)

	// 执行分配
	err = s.assignTaskToCourier(task, bestCourier, manager.ID)
	if err != nil {
		return nil, fmt.Errorf("分配任务失败: %w", err)
	}

	return task, nil
}

// autoHierarchyAssignment 自动层级分配
func (s *HierarchicalAssignmentService) autoHierarchyAssignment(manager *models.Courier, task *models.Task, req *models.HierarchicalTaskAssignmentRequest) (*models.Task, error) {
	// 使用现有的自动分配服务，但限制在管理范围内
	taskZoneCode := s.assignmentService.extractZoneCodeFromLocation(task.PickupLocation)

	// 查找管理范围内的最优信使
	bestCourier, err := s.findBestCourierInHierarchy(manager, taskZoneCode, task)
	if err != nil {
		return nil, fmt.Errorf("找不到合适的信使: %w", err)
	}

	// 执行分配
	err = s.assignTaskToCourier(task, bestCourier, manager.ID)
	if err != nil {
		return nil, fmt.Errorf("分配任务失败: %w", err)
	}

	return task, nil
}

// BatchAssignByHierarchy 批量层级分配任务
func (s *HierarchicalAssignmentService) BatchAssignByHierarchy(managerID string, req *models.BatchHierarchicalAssignmentRequest) (*models.BatchAssignmentResponse, error) {
	var results []models.TaskAssignmentResult
	successCount := 0

	for _, taskAssignment := range req.TaskAssignments {
		singleReq := &models.HierarchicalTaskAssignmentRequest{
			TaskID:          taskAssignment.TaskID,
			AssignmentType:  req.AssignmentType,
			TargetCourierID: taskAssignment.TargetCourierID,
			Priority:        taskAssignment.Priority,
			Notes:           taskAssignment.Notes,
		}

		task, err := s.AssignTaskByHierarchy(managerID, singleReq)
		result := models.TaskAssignmentResult{
			TaskID: taskAssignment.TaskID,
		}

		if err != nil {
			result.Success = false
			result.Error = err.Error()
		} else {
			result.Success = true
			result.AssignedCourierID = task.CourierID
			successCount++
		}

		results = append(results, result)
	}

	return &models.BatchAssignmentResponse{
		Results:      results,
		SuccessCount: successCount,
		TotalCount:   len(req.TaskAssignments),
	}, nil
}

// GetAssignmentHistory 获取分配历史
func (s *HierarchicalAssignmentService) GetAssignmentHistory(managerID string, limit int, offset int) (*models.AssignmentHistoryResponse, error) {
	var assignments []models.TaskAssignmentHistory
	var total int64

	query := s.db.Model(&models.TaskAssignmentHistory{}).Where("assigned_by = ?", managerID)

	// 获取总数
	query.Count(&total)

	// 获取分页数据
	err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Preload("Task").
		Preload("AssignedCourier").
		Find(&assignments).Error

	if err != nil {
		return nil, fmt.Errorf("获取分配历史失败: %w", err)
	}

	return &models.AssignmentHistoryResponse{
		Assignments: assignments,
		Total:       int(total),
		Page:        offset/limit + 1,
		Limit:       limit,
	}, nil
}

// ReassignTask 重新分配任务
func (s *HierarchicalAssignmentService) ReassignTask(managerID string, req *models.TaskReassignmentRequest) (*models.Task, error) {
	// 获取管理者信息
	var manager models.Courier
	if err := s.db.First(&manager, managerID).Error; err != nil {
		return nil, fmt.Errorf("管理者不存在: %w", err)
	}

	// 获取任务信息
	var task models.Task
	if err := s.db.Where("task_id = ?", req.TaskID).First(&task).Error; err != nil {
		return nil, fmt.Errorf("任务不存在: %w", err)
	}

	// 获取原信使信息
	var oldCourier models.Courier
	if task.CourierID != nil {
		s.db.Where("user_id = ?", *task.CourierID).First(&oldCourier)
	}

	// 检查管理权限
	if task.CourierID != nil && !s.canAssignToSubordinate(&manager, &oldCourier) {
		return nil, errors.New("无权限重新分配该任务")
	}

	// 获取新的目标信使
	var newCourier models.Courier
	if err := s.db.First(&newCourier, req.NewCourierID).Error; err != nil {
		return nil, fmt.Errorf("新的目标信使不存在: %w", err)
	}

	// 检查新信使权限
	if !s.canAssignToSubordinate(&manager, &newCourier) {
		return nil, errors.New("无权限分配任务给新信使")
	}

	// 执行重新分配
	err := s.reassignTaskToCourier(&task, &newCourier, manager.ID, req.Reason)
	if err != nil {
		return nil, fmt.Errorf("重新分配失败: %w", err)
	}

	return &task, nil
}

// 私有方法

// canAssignToSubordinate 检查是否可以分配任务给指定信使
func (s *HierarchicalAssignmentService) canAssignToSubordinate(manager *models.Courier, courier *models.Courier) bool {
	// 检查是否是直接下属
	if courier.ParentID != nil && *courier.ParentID == manager.ID {
		return true
	}

	// 检查是否在管理范围内（同等级或更低等级）
	if manager.Level >= courier.Level {
		return s.isInManagementScope(manager, courier)
	}

	return false
}

// canHandleTaskInZone 检查信使是否能处理指定区域的任务
func (s *HierarchicalAssignmentService) canHandleTaskInZone(courier *models.Courier, task *models.Task) bool {
	taskZoneCode := s.assignmentService.extractZoneCodeFromLocation(task.PickupLocation)
	return s.assignmentService.validateAssignmentPermission(courier, task) ||
		courier.ZoneCode == taskZoneCode ||
		s.assignmentService.isParentZone(courier.ZoneCode, courier.ZoneType, taskZoneCode)
}

// selectBestCourierForTask 为任务选择最佳信使
func (s *HierarchicalAssignmentService) selectBestCourierForTask(couriers []models.Courier, task *models.Task) *models.Courier {
	if len(couriers) == 0 {
		return nil
	}

	// 使用现有的评分机制
	pickupLat, pickupLng, _ := s.assignmentService.locationService.ParseLocation(task.PickupLocation)
	scores := s.assignmentService.calculateCourierScoresWithHierarchy(couriers, task, pickupLat, pickupLng)

	if len(scores) > 0 {
		return &scores[0].Courier
	}

	return &couriers[0]
}

// findBestCourierInHierarchy 在层级结构中找到最佳信使
func (s *HierarchicalAssignmentService) findBestCourierInHierarchy(manager *models.Courier, taskZoneCode string, task *models.Task) (*models.Courier, error) {
	// 获取管理范围内的所有信使
	var managedCouriers []models.Courier

	// 查找所有下属信使
	s.db.Where("parent_id = ? OR (level <= ? AND zone_code LIKE ?)",
		manager.ID, manager.Level, taskZoneCode+"%").
		Find(&managedCouriers)

	// 筛选能够处理该任务的信使
	var validCouriers []models.Courier
	for _, courier := range managedCouriers {
		if s.canHandleTaskInZone(&courier, task) {
			validCouriers = append(validCouriers, courier)
		}
	}

	if len(validCouriers) == 0 {
		return nil, errors.New("没有找到合适的信使")
	}

	return s.selectBestCourierForTask(validCouriers, task), nil
}

// isInManagementScope 检查是否在管理范围内
func (s *HierarchicalAssignmentService) isInManagementScope(manager *models.Courier, courier *models.Courier) bool {
	// 根据层级和区域范围判断
	switch manager.ZoneType {
	case models.ZoneTypeCity:
		// 城市级可以管理城市内所有信使
		return s.assignmentService.getSchoolToCityMapping(
			s.assignmentService.extractParentZoneCode(
				s.assignmentService.extractParentZoneCode(courier.ZoneCode, models.ZoneTypeArea),
				models.ZoneTypeSchool,
			),
		) == manager.ZoneCode
	case models.ZoneTypeSchool:
		// 学校级可以管理同学校的信使
		return len(courier.ZoneCode) > len(manager.ZoneCode) &&
			courier.ZoneCode[:len(manager.ZoneCode)] == manager.ZoneCode
	case models.ZoneTypeArea:
		// 片区级可以管理同片区的信使
		return len(courier.ZoneCode) > len(manager.ZoneCode) &&
			courier.ZoneCode[:len(manager.ZoneCode)] == manager.ZoneCode
	default:
		return false
	}
}

// assignTaskToCourier 分配任务给信使
func (s *HierarchicalAssignmentService) assignTaskToCourier(task *models.Task, courier *models.Courier, assignedBy string) error {
	now := time.Now()
	deadline := now.Add(4 * time.Hour)

	// 开始事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新任务信息
	err := tx.Model(task).Updates(map[string]interface{}{
		"courier_id":  &courier.UserID,
		"status":      models.TaskStatusAccepted,
		"accepted_at": &now,
		"deadline":    &deadline,
	}).Error

	if err != nil {
		tx.Rollback()
		return err
	}

	// 记录分配历史
	history := &models.TaskAssignmentHistory{
		TaskID:            task.TaskID,
		AssignedCourierID: courier.ID,
		AssignedBy:        assignedBy,
		AssignmentType:    "hierarchical",
		CreatedAt:         now,
	}

	if err := tx.Create(history).Error; err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()

	// 发送通知
	s.notifyTaskAssigned(task, courier, assignedBy)

	return nil
}

// reassignTaskToCourier 重新分配任务
func (s *HierarchicalAssignmentService) reassignTaskToCourier(task *models.Task, newCourier *models.Courier, assignedBy string, reason string) error {
	now := time.Now()
	oldCourierID := task.CourierID

	// 开始事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新任务信息
	err := tx.Model(task).Updates(map[string]interface{}{
		"courier_id":  &newCourier.UserID,
		"status":      models.TaskStatusAccepted,
		"accepted_at": &now,
	}).Error

	if err != nil {
		tx.Rollback()
		return err
	}

	// 记录重新分配历史
	history := &models.TaskAssignmentHistory{
		TaskID:             task.TaskID,
		AssignedCourierID:  newCourier.ID,
		AssignedBy:         assignedBy,
		AssignmentType:     "reassignment",
		PreviousCourierID:  oldCourierID,
		ReassignmentReason: reason,
		CreatedAt:          now,
	}

	if err := tx.Create(history).Error; err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()

	// 发送通知
	s.notifyTaskReassigned(task, newCourier, oldCourierID, reason)

	return nil
}

// 通知方法

// notifyTaskAssigned 通知任务已分配
func (s *HierarchicalAssignmentService) notifyTaskAssigned(task *models.Task, courier *models.Courier, assignedBy string) {
	event := utils.WebSocketEvent{
		Type: "HIERARCHICAL_TASK_ASSIGNED",
		Data: map[string]interface{}{
			"task_id":     task.TaskID,
			"courier_id":  courier.UserID,
			"assigned_by": assignedBy,
			"level":       courier.Level,
			"zone_type":   courier.ZoneType,
			"deadline":    task.Deadline,
		},
		Timestamp: time.Now(),
	}

	// 通知信使
	s.wsManager.BroadcastToUser(courier.UserID, event)

	// 通知分配者
	var assigner models.Courier
	if s.db.First(&assigner, assignedBy).Error == nil {
		s.wsManager.BroadcastToUser(assigner.UserID, event)
	}
}

// notifyTaskReassigned 通知任务已重新分配
func (s *HierarchicalAssignmentService) notifyTaskReassigned(task *models.Task, newCourier *models.Courier, oldCourierID *string, reason string) {
	event := utils.WebSocketEvent{
		Type: "TASK_REASSIGNED",
		Data: map[string]interface{}{
			"task_id":           task.TaskID,
			"new_courier_id":    newCourier.UserID,
			"old_courier_id":    oldCourierID,
			"reason":            reason,
			"reassignment_time": time.Now(),
		},
		Timestamp: time.Now(),
	}

	// 通知新信使
	s.wsManager.BroadcastToUser(newCourier.UserID, event)

	// 通知原信使
	if oldCourierID != nil {
		s.wsManager.BroadcastToUser(*oldCourierID, event)
	}
}

// GetDB 获取数据库连接
func (s *HierarchicalAssignmentService) GetDB() *gorm.DB {
	return s.db
}
