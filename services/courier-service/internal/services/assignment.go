package services

import (
	"courier-service/internal/models"
	"courier-service/internal/utils"
	"fmt"
	"sort"
	"time"

	"gorm.io/gorm"
)

// AssignmentService 任务分配服务
type AssignmentService struct {
	db              *gorm.DB
	locationService *LocationService
	wsManager       *utils.WebSocketManager
}

// NewAssignmentService 创建任务分配服务实例
func NewAssignmentService(db *gorm.DB, locationService *LocationService, wsManager *utils.WebSocketManager) *AssignmentService {
	return &AssignmentService{
		db:              db,
		locationService: locationService,
		wsManager:       wsManager,
	}
}

// CourierScore 信使评分结构
type CourierScore struct {
	Courier      models.Courier
	Score        float64
	Distance     float64
	CurrentTasks int
}

// AutoAssignTask 自动分配任务给合适的信使
func (s *AssignmentService) AutoAssignTask(task *models.Task) (*models.Courier, error) {
	// 1. 解析取件位置坐标
	pickupLat, pickupLng, err := s.locationService.ParseLocation(task.PickupLocation)
	if err != nil {
		return nil, err
	}

	// 2. 根据4级层级结构查找合适的信使
	nearbyCouries := s.findCouriersByHierarchy(task, pickupLat, pickupLng, 10.0) // 10km范围
	if len(nearbyCouries) == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	// 3. 计算信使评分并排序（考虑层级优先级）
	courierScores := s.calculateCourierScoresWithHierarchy(nearbyCouries, task, pickupLat, pickupLng)
	if len(courierScores) == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	// 4. 选择最优信使
	bestCourier := courierScores[0].Courier

	// 5. 检查层级权限和区域管辖
	if !s.validateAssignmentPermission(&bestCourier, task) {
		return nil, fmt.Errorf("信使无权限处理该区域任务")
	}

	// 6. 更新任务分配信息
	err = s.assignTaskToCourier(task, &bestCourier)
	if err != nil {
		return nil, err
	}

	// 7. 发送通知
	s.notifyTaskAssignment(task, &bestCourier)

	return &bestCourier, nil
}

// findCouriersByHierarchy 根据4级层级结构查找合适的信使
func (s *AssignmentService) findCouriersByHierarchy(task *models.Task, lat, lng, radiusKm float64) []models.Courier {
	// 确定任务的区域编码
	taskZoneCode := s.extractZoneCodeFromLocation(task.PickupLocation)
	
	var couriers []models.Courier

	// 按层级优先级查找信使：楼栋级 > 片区级 > 学校级 > 城市级
	// 1. 优先找楼栋级信使（最精确匹配）
	s.db.Where("status = ? AND zone_type = ? AND zone_code = ? AND rating >= ?", 
		models.CourierStatusApproved, models.ZoneTypeBuilding, taskZoneCode, 3.0).Find(&couriers)

	// 2. 如果没有楼栋级，找片区级
	if len(couriers) == 0 {
		areaZoneCode := s.extractParentZoneCode(taskZoneCode, models.ZoneTypeArea)
		s.db.Where("status = ? AND zone_type = ? AND zone_code = ? AND rating >= ?", 
			models.CourierStatusApproved, models.ZoneTypeArea, areaZoneCode, 3.0).Find(&couriers)
	}

	// 3. 如果没有片区级，找学校级
	if len(couriers) == 0 {
		schoolZoneCode := s.extractParentZoneCode(taskZoneCode, models.ZoneTypeSchool)
		s.db.Where("status = ? AND zone_type = ? AND zone_code = ? AND rating >= ?", 
			models.CourierStatusApproved, models.ZoneTypeSchool, schoolZoneCode, 3.0).Find(&couriers)
	}

	// 4. 最后找城市级（兜底）
	if len(couriers) == 0 {
		cityZoneCode := s.extractParentZoneCode(taskZoneCode, models.ZoneTypeCity)
		s.db.Where("status = ? AND zone_type = ? AND zone_code = ? AND rating >= ?", 
			models.CourierStatusApproved, models.ZoneTypeCity, cityZoneCode, 3.0).Find(&couriers)
	}

	return couriers
}

// findNearbyActiveCouriers 查找附近的活跃信使（原有方法保持兼容）
func (s *AssignmentService) findNearbyActiveCouriers(lat, lng, radiusKm float64) []models.Courier {
	var couriers []models.Courier

	// 查找所有活跃的信使
	s.db.Where("status = ? AND rating >= ?", models.CourierStatusApproved, 3.0).Find(&couriers)

	var nearbyCouriers []models.Courier
	for _, courier := range couriers {
		// 这里可以根据信使的位置信息过滤
		// 暂时返回所有活跃信使
		nearbyCouriers = append(nearbyCouriers, courier)
	}

	return nearbyCouriers
}

// calculateCourierScoresWithHierarchy 计算信使评分（考虑层级优先级）
func (s *AssignmentService) calculateCourierScoresWithHierarchy(couriers []models.Courier, task *models.Task, taskLat, taskLng float64) []CourierScore {
	var scores []CourierScore

	for _, courier := range couriers {
		score := s.calculateIndividualScoreWithHierarchy(&courier, task, taskLat, taskLng)
		scores = append(scores, score)
	}

	// 按评分排序（降序），层级优先级更高
	sort.Slice(scores, func(i, j int) bool {
		// 首先按层级优先级排序
		levelPriorityI := s.getLevelPriority(scores[i].Courier.ZoneType)
		levelPriorityJ := s.getLevelPriority(scores[j].Courier.ZoneType)
		if levelPriorityI != levelPriorityJ {
			return levelPriorityI > levelPriorityJ
		}
		// 相同层级再按评分排序
		return scores[i].Score > scores[j].Score
	})

	return scores
}

// calculateCourierScores 计算信使评分（原有方法保持兼容）
func (s *AssignmentService) calculateCourierScores(couriers []models.Courier, taskLat, taskLng float64) []CourierScore {
	var scores []CourierScore

	for _, courier := range couriers {
		score := s.calculateIndividualScore(&courier, taskLat, taskLng)
		scores = append(scores, score)
	}

	// 按评分排序（降序）
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].Score > scores[j].Score
	})

	return scores
}

// calculateIndividualScoreWithHierarchy 计算单个信使的评分（考虑层级）
func (s *AssignmentService) calculateIndividualScoreWithHierarchy(courier *models.Courier, task *models.Task, taskLat, taskLng float64) CourierScore {
	score := CourierScore{
		Courier: *courier,
	}

	// 基础评分（信使评分 * 20）
	baseScore := courier.Rating * 20.0

	// 层级优先级加分（楼栋级最高，城市级最低）
	hierarchyBonus := s.getHierarchyBonus(courier.ZoneType)
	
	// 区域匹配度加分
	zoneMatchScore := s.calculateZoneMatchScore(courier, task)

	// 距离评分（越近越好，最大50分）
	courierLat, courierLng, _ := s.locationService.ParseLocation(courier.Zone)
	distance := s.locationService.CalculateDistance(courierLat, courierLng, taskLat, taskLng)
	score.Distance = distance

	distanceScore := 50.0
	if distance > 0 {
		distanceScore = 50.0 / (1.0 + distance*distance)
	}

	// 工作负载评分（当前任务越少越好，最大30分）
	currentTasks := s.getCurrentTaskCount(courier.UserID)
	score.CurrentTasks = currentTasks

	workloadScore := 30.0
	if currentTasks > 0 {
		workloadScore = 30.0 / (1.0 + float64(currentTasks)*0.5)
	}

	// 经验值加分（最大20分）
	experienceScore := 10.0 // 基础经验分数
	if len(courier.Experience) > 50 {
		experienceScore = 20.0 // 详细经验描述获得更高分
	}

	// 积分加分（最大15分）
	pointScore := float64(courier.Points) * 0.01
	if pointScore > 15.0 {
		pointScore = 15.0
	}

	// 最终评分
	score.Score = baseScore + hierarchyBonus + zoneMatchScore + distanceScore + workloadScore + experienceScore + pointScore

	return score
}

// calculateIndividualScore 计算单个信使的评分（原有方法保持兼容）
func (s *AssignmentService) calculateIndividualScore(courier *models.Courier, taskLat, taskLng float64) CourierScore {
	score := CourierScore{
		Courier: *courier,
	}

	// 基础评分（信使评分 * 20）
	baseScore := courier.Rating * 20.0

	// 距离评分（越近越好，最大50分）
	// 假设信使在其服务区域的中心位置
	courierLat, courierLng, _ := s.locationService.ParseLocation(courier.Zone)
	distance := s.locationService.CalculateDistance(courierLat, courierLng, taskLat, taskLng)
	score.Distance = distance

	distanceScore := 50.0
	if distance > 0 {
		distanceScore = 50.0 / (1.0 + distance*distance) // 距离越远分数越低
	}

	// 工作负载评分（当前任务越少越好，最大30分）
	currentTasks := s.getCurrentTaskCount(courier.UserID)
	score.CurrentTasks = currentTasks

	workloadScore := 30.0
	if currentTasks > 0 {
		workloadScore = 30.0 / (1.0 + float64(currentTasks)*0.5)
	}

	// 最终评分
	score.Score = baseScore + distanceScore + workloadScore

	return score
}

// getCurrentTaskCount 获取信使当前的任务数量
func (s *AssignmentService) getCurrentTaskCount(courierID string) int {
	var count int64
	s.db.Model(&models.Task{}).Where("courier_id = ? AND status IN ?", courierID, 
		[]string{models.TaskStatusAccepted, models.TaskStatusCollected, models.TaskStatusInTransit}).Count(&count)
	return int(count)
}

// assignTaskToCourier 将任务分配给信使
func (s *AssignmentService) assignTaskToCourier(task *models.Task, courier *models.Courier) error {
	now := time.Now()
	deadline := now.Add(4 * time.Hour) // 4小时截止

	updates := map[string]interface{}{
		"courier_id":  &courier.UserID,
		"status":      models.TaskStatusAccepted,
		"accepted_at": &now,
		"deadline":    &deadline,
	}

	return s.db.Model(task).Where("task_id = ?", task.TaskID).Updates(updates).Error
}

// notifyTaskAssignment 通知任务分配
func (s *AssignmentService) notifyTaskAssignment(task *models.Task, courier *models.Courier) {
	// 通知信使
	s.wsManager.SendTaskAssignment(task.TaskID, courier.UserID, map[string]interface{}{
		"task_id":            task.TaskID,
		"letter_id":          task.LetterID,
		"pickup_location":    task.PickupLocation,
		"delivery_location":  task.DeliveryLocation,
		"priority":           task.Priority,
		"reward":             task.Reward,
		"estimated_distance": task.EstimatedDistance,
		"deadline":           task.Deadline,
	})

	// 通知管理员
	s.wsManager.BroadcastToAdmins(utils.WebSocketEvent{
		Type: "TASK_AUTO_ASSIGNED",
		Data: map[string]interface{}{
			"task_id":    task.TaskID,
			"courier_id": courier.UserID,
			"zone":       courier.Zone,
		},
		Timestamp: time.Now(),
	})
}

// FindOptimalCouriersForTask 为任务找到最优的信使候选
func (s *AssignmentService) FindOptimalCouriersForTask(taskID string, limit int) ([]CourierScore, error) {
	// 获取任务信息
	var task models.Task
	if err := s.db.Where("task_id = ?", taskID).First(&task).Error; err != nil {
		return nil, err
	}

	// 解析位置
	pickupLat, pickupLng, err := s.locationService.ParseLocation(task.PickupLocation)
	if err != nil {
		return nil, err
	}

	// 查找附近信使
	nearbyCouries := s.findNearbyActiveCouriers(pickupLat, pickupLng, 15.0) // 15km范围

	// 计算评分
	courierScores := s.calculateCourierScores(nearbyCouries, pickupLat, pickupLng)

	// 限制返回数量
	if len(courierScores) > limit {
		courierScores = courierScores[:limit]
	}

	return courierScores, nil
}

// BatchAssignTasks 批量自动分配任务
func (s *AssignmentService) BatchAssignTasks(maxTasks int) (int, error) {
	// 获取待分配的任务
	var tasks []models.Task
	err := s.db.Where("status = ? AND courier_id IS NULL", models.TaskStatusAvailable).
		Order("priority DESC, created_at ASC").
		Limit(maxTasks).
		Find(&tasks).Error

	if err != nil {
		return 0, err
	}

	assignedCount := 0
	for _, task := range tasks {
		courier, err := s.AutoAssignTask(&task)
		if err == nil && courier != nil {
			assignedCount++
		}
	}

	return assignedCount, nil
}

// ReassignFailedTasks 重新分配失败的任务
func (s *AssignmentService) ReassignFailedTasks() error {
	// 查找超时未完成的任务
	deadline := time.Now().Add(-6 * time.Hour) // 6小时前的任务
	var timeoutTasks []models.Task

	err := s.db.Where("status IN ? AND deadline < ?", 
		[]string{models.TaskStatusAccepted, models.TaskStatusCollected}, deadline).
		Find(&timeoutTasks).Error

	if err != nil {
		return err
	}

	// 重置任务状态并重新分配
	for _, task := range timeoutTasks {
		// 重置任务
		s.db.Model(&task).Updates(map[string]interface{}{
			"courier_id":  nil,
			"status":      models.TaskStatusAvailable,
			"accepted_at": nil,
			"deadline":    nil,
		})

		// 尝试重新分配
		s.AutoAssignTask(&task)
	}

	return nil
}

// 新增的层级相关方法

// extractZoneCodeFromLocation 从位置信息提取区域编码
func (s *AssignmentService) extractZoneCodeFromLocation(location string) string {
	// 这里应该根据实际的位置格式来解析
	// 示例：假设位置格式为 "PKU001-AREA001-BUILDING001"
	// 这里简化处理，返回楼栋级编码
	if len(location) >= 20 {
		return location // 假设已经是标准格式
	}
	// 默认返回一个示例编码
	return "PKU001-AREA001-BUILDING001"
}

// extractParentZoneCode 提取父级区域编码
func (s *AssignmentService) extractParentZoneCode(childZoneCode string, parentType string) string {
	switch parentType {
	case models.ZoneTypeArea:
		// 从楼栋级提取片区级： "PKU001-AREA001-BUILDING001" -> "PKU001-AREA001"
		if len(childZoneCode) > 12 {
			return childZoneCode[:12]
		}
	case models.ZoneTypeSchool:
		// 从片区级提取学校级： "PKU001-AREA001" -> "PKU001"
		if len(childZoneCode) > 6 {
			return childZoneCode[:6]
		}
	case models.ZoneTypeCity:
		// 从学校级提取城市级： "PKU001" -> "BJ001"(假设北京地区）
		// 这里需要一个学校到城市的映射关系
		return s.getSchoolToCityMapping(childZoneCode)
	}
	return childZoneCode
}

// getSchoolToCityMapping 获取学校到城市的映射
func (s *AssignmentService) getSchoolToCityMapping(schoolCode string) string {
	// 这里应该有一个映射表，暂时用简单逻辑
	if len(schoolCode) >= 3 {
		prefix := schoolCode[:3]
		switch prefix {
		case "PKU", "TSI": // 北京地区学校
			return "BJ001"
		case "FUD", "XMU": // 福建地区学校
			return "FJ001"
		default:
			return "CN001" // 默认城市
		}
	}
	return "CN001"
}

// getLevelPriority 获取层级优先级
func (s *AssignmentService) getLevelPriority(zoneType string) int {
	switch zoneType {
	case models.ZoneTypeBuilding:
		return 4 // 最高优先级
	case models.ZoneTypeArea:
		return 3
	case models.ZoneTypeSchool:
		return 2
	case models.ZoneTypeCity:
		return 1 // 最低优先级
	default:
		return 0
	}
}

// getHierarchyBonus 获取层级加分
func (s *AssignmentService) getHierarchyBonus(zoneType string) float64 {
	switch zoneType {
	case models.ZoneTypeBuilding:
		return 40.0 // 楼栋级最高加分
	case models.ZoneTypeArea:
		return 30.0
	case models.ZoneTypeSchool:
		return 20.0
	case models.ZoneTypeCity:
		return 10.0
	default:
		return 0.0
	}
}

// calculateZoneMatchScore 计算区域匹配度加分
func (s *AssignmentService) calculateZoneMatchScore(courier *models.Courier, task *models.Task) float64 {
	taskZoneCode := s.extractZoneCodeFromLocation(task.PickupLocation)
	
	// 如果是精确匹配
	if courier.ZoneCode == taskZoneCode {
		return 50.0
	}
	
	// 如果是上级管辖区域
	if s.isParentZone(courier.ZoneCode, courier.ZoneType, taskZoneCode) {
		return 30.0
	}
	
	// 如果是同级但不同区域
	if courier.ZoneType == models.ZoneTypeBuilding && 
		s.extractParentZoneCode(courier.ZoneCode, models.ZoneTypeArea) == 
		s.extractParentZoneCode(taskZoneCode, models.ZoneTypeArea) {
		return 20.0
	}
	
	return 0.0
}

// isParentZone 检查是否为父级管辖区域
func (s *AssignmentService) isParentZone(courierZoneCode, courierZoneType, taskZoneCode string) bool {
	switch courierZoneType {
	case models.ZoneTypeArea:
		// 片区级可以管辖同片区的所有楼栋
		return len(taskZoneCode) > len(courierZoneCode) && 
			taskZoneCode[:len(courierZoneCode)] == courierZoneCode
	case models.ZoneTypeSchool:
		// 学校级可以管辖同学校的所有片区和楼栋
		return len(taskZoneCode) > len(courierZoneCode) && 
			taskZoneCode[:len(courierZoneCode)] == courierZoneCode
	case models.ZoneTypeCity:
		// 城市级可以管辖城市内所有任务
		schoolCode := s.extractParentZoneCode(
			s.extractParentZoneCode(taskZoneCode, models.ZoneTypeArea),
			models.ZoneTypeSchool,
		)
		return s.getSchoolToCityMapping(schoolCode) == courierZoneCode
	default:
		return false
	}
}

// validateAssignmentPermission 验证分配权限
func (s *AssignmentService) validateAssignmentPermission(courier *models.Courier, task *models.Task) bool {
	taskZoneCode := s.extractZoneCodeFromLocation(task.PickupLocation)
	
	// 检查信使是否有权限处理该区域的任务
	return courier.ZoneCode == taskZoneCode || 
		s.isParentZone(courier.ZoneCode, courier.ZoneType, taskZoneCode)
}