package services

import (
	"courier-service/internal/models"
	"courier-service/internal/utils"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// CourierLevelService 信使等级管理服务
type CourierLevelService struct {
	db        *gorm.DB
	redis     *redis.Client
	wsManager *utils.WebSocketManager
}

// NewCourierLevelService 创建信使等级服务
func NewCourierLevelService(db *gorm.DB, redis *redis.Client, wsManager *utils.WebSocketManager) *CourierLevelService {
	return &CourierLevelService{
		db:        db,
		redis:     redis,
		wsManager: wsManager,
	}
}

// CourierLevelInfo 信使等级信息结构
type CourierLevelInfo struct {
	CourierID   string                     `json:"courier_id"`
	Level       models.CourierLevel        `json:"level"`
	LevelName   string                     `json:"level_name"`
	ZoneType    models.CourierZoneType     `json:"zone_type"`
	Permissions []models.CourierPermission `json:"permissions"`
	Zones       []models.CourierZone       `json:"zones"`
	CanUpgrade  bool                       `json:"can_upgrade"`
	NextLevel   *models.CourierLevel       `json:"next_level,omitempty"`
}

// GetCourierLevelInfo 获取信使等级信息
func (s *CourierLevelService) GetCourierLevelInfo(courierID string) (*CourierLevelInfo, error) {
	// 获取信使基本信息
	var courier models.Courier
	if err := s.db.Where("user_id = ?", courierID).First(&courier).Error; err != nil {
		return nil, fmt.Errorf("courier not found: %w", err)
	}

	level := models.CourierLevel(courier.Level)

	// 获取管理区域
	var zones []models.CourierZone
	s.db.Where("courier_id = ? AND is_active = ?", courierID, true).Find(&zones)

	// 检查是否可以升级
	canUpgrade := false
	var nextLevel *models.CourierLevel
	if level < models.LevelFour {
		next := level + 1
		nextLevel = &next
		canUpgrade, _ = s.CheckUpgradeEligibility(courierID, next)
	}

	return &CourierLevelInfo{
		CourierID:   courierID,
		Level:       level,
		LevelName:   level.GetLevelName(),
		ZoneType:    models.DefaultZoneMapping[level],
		Permissions: models.DefaultPermissionMatrix[level],
		Zones:       zones,
		CanUpgrade:  canUpgrade,
		NextLevel:   nextLevel,
	}, nil
}

// CheckUpgradeEligibility 检查升级资格
func (s *CourierLevelService) CheckUpgradeEligibility(courierID string, targetLevel models.CourierLevel) (bool, string) {
	// 获取当前信使信息
	var courier models.Courier
	if err := s.db.Where("user_id = ?", courierID).First(&courier).Error; err != nil {
		return false, "信使不存在"
	}

	currentLevel := models.CourierLevel(courier.Level)

	// 检查等级顺序
	if targetLevel != currentLevel+1 {
		return false, "只能申请下一等级"
	}

	// 根据PRD成长分级机制检查升级条件
	switch targetLevel {
	case models.LevelTwo:
		return s.checkLevelTwoRequirements(courierID)
	case models.LevelThree:
		return s.checkLevelThreeRequirements(courierID)
	case models.LevelFour:
		return s.checkLevelFourRequirements(courierID)
	default:
		return false, "无效的目标等级"
	}
}

// checkLevelTwoRequirements 检查二级信使升级条件
func (s *CourierLevelService) checkLevelTwoRequirements(courierID string) (bool, string) {
	// 1级→2级: 累计投递10封信 + 连续7天投递

	// 检查累计投递数量
	var totalDeliveries int64
	if err := s.db.Model(&models.Task{}).
		Where("courier_id = ? AND status = ?", courierID, "delivered").
		Count(&totalDeliveries).Error; err != nil {
		return false, "无法查询投递记录"
	}

	if totalDeliveries < 10 {
		return false, fmt.Sprintf("累计投递数量不足，需要10封，当前%d封", totalDeliveries)
	}

	// 检查连续7天投递 (简化实现，检查最近7天是否有投递记录)
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	var recentDeliveries int64
	if err := s.db.Model(&models.Task{}).
		Where("courier_id = ? AND status = ? AND completed_at > ?", courierID, "delivered", sevenDaysAgo).
		Count(&recentDeliveries).Error; err != nil {
		return false, "无法查询最近投递记录"
	}

	if recentDeliveries < 7 {
		return false, "最近7天投递量不足"
	}

	return true, ""
}

// checkLevelThreeRequirements 检查三级信使升级条件
func (s *CourierLevelService) checkLevelThreeRequirements(courierID string) (bool, string) {
	// 2级→3级: 管理≥3位1级信使 + 月完成率>80%

	// 检查管理的一级信使数量 (这里简化为检查分配的区域数量)
	var managedZones int64
	if err := s.db.Model(&models.CourierZone{}).
		Where("courier_id = ? AND zone_type = ? AND is_active = ?", courierID, models.ZoneArea, true).
		Count(&managedZones).Error; err != nil {
		return false, "无法查询管理区域"
	}

	if managedZones < 3 {
		return false, fmt.Sprintf("管理区域不足，需要3个，当前%d个", managedZones)
	}

	// 检查月完成率
	oneMonthAgo := time.Now().AddDate(0, -1, 0)

	var totalTasks, completedTasks int64
	s.db.Model(&models.Task{}).
		Where("courier_id = ? AND created_at > ?", courierID, oneMonthAgo).
		Count(&totalTasks)

	s.db.Model(&models.Task{}).
		Where("courier_id = ? AND status = ? AND created_at > ?", courierID, "delivered", oneMonthAgo).
		Count(&completedTasks)

	if totalTasks == 0 {
		return false, "本月无任务记录"
	}

	completionRate := float64(completedTasks) / float64(totalTasks) * 100
	if completionRate < 80 {
		return false, fmt.Sprintf("月完成率不足，需要80%%，当前%.1f%%", completionRate)
	}

	return true, ""
}

// checkLevelFourRequirements 检查四级信使升级条件
func (s *CourierLevelService) checkLevelFourRequirements(courierID string) (bool, string) {
	// 3级→4级: 校级推荐 + 3个月服务时长

	// 检查服务时长
	var courier models.Courier
	if err := s.db.Where("user_id = ?", courierID).First(&courier).Error; err != nil {
		return false, "信使不存在"
	}

	if courier.ApprovedAt == nil {
		return false, "信使审核时间不明确"
	}

	threeMonthsAgo := time.Now().AddDate(0, -3, 0)
	if courier.ApprovedAt.After(threeMonthsAgo) {
		return false, "服务时长不足3个月"
	}

	// 检查是否有校级推荐 (这里简化为检查是否有特殊标记)
	// 实际实现中可能需要额外的推荐表
	if courier.Note != "校级推荐" { // 简化检查
		return false, "需要校级推荐"
	}

	return true, ""
}

// SubmitUpgradeRequest 提交升级申请
func (s *CourierLevelService) SubmitUpgradeRequest(courierID string, targetLevel models.CourierLevel, reason string, evidence map[string]interface{}) (*models.LevelUpgradeRequest, error) {
	// 检查是否已有待处理的申请
	var existingRequest models.LevelUpgradeRequest
	if err := s.db.Where("courier_id = ? AND status = ?", courierID, "pending").First(&existingRequest).Error; err == nil {
		return nil, errors.New("已有待处理的升级申请")
	}

	// 获取当前等级
	var courier models.Courier
	if err := s.db.Where("user_id = ?", courierID).First(&courier).Error; err != nil {
		return nil, fmt.Errorf("信使不存在: %w", err)
	}

	evidenceJSON, _ := json.Marshal(evidence)

	upgradeRequest := &models.LevelUpgradeRequest{
		CourierID:    courierID,
		CurrentLevel: models.CourierLevel(courier.Level),
		RequestLevel: targetLevel,
		Reason:       reason,
		Evidence:     string(evidenceJSON),
		Status:       "pending",
	}

	if err := s.db.Create(upgradeRequest).Error; err != nil {
		return nil, fmt.Errorf("创建升级申请失败: %w", err)
	}

	// 发送通知给管理员
	s.notifyUpgradeRequest(upgradeRequest)

	return upgradeRequest, nil
}

// GetUpgradeRequests 获取升级申请列表
func (s *CourierLevelService) GetUpgradeRequests(status string, limit, offset int) ([]models.LevelUpgradeRequest, int64, error) {
	var requests []models.LevelUpgradeRequest
	var total int64

	query := s.db.Model(&models.LevelUpgradeRequest{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 获取总数
	query.Count(&total)

	// 获取分页数据
	if err := query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&requests).Error; err != nil {
		return nil, 0, fmt.Errorf("查询升级申请失败: %w", err)
	}

	return requests, total, nil
}

// ProcessUpgradeRequest 处理升级申请
func (s *CourierLevelService) ProcessUpgradeRequest(requestID, action, comment, reviewerID string) error {
	var request models.LevelUpgradeRequest
	if err := s.db.Where("id = ?", requestID).First(&request).Error; err != nil {
		return fmt.Errorf("升级申请不存在: %w", err)
	}

	if request.Status != "pending" {
		return errors.New("申请已被处理")
	}

	now := time.Now()
	request.Status = action
	request.ReviewedBy = &reviewerID
	request.ReviewedAt = &now
	request.ReviewComment = comment

	if err := s.db.Save(&request).Error; err != nil {
		return fmt.Errorf("更新申请状态失败: %w", err)
	}

	// 如果批准，更新信使等级
	if action == "approved" {
		if err := s.upgradeCourier(request.CourierID, request.RequestLevel); err != nil {
			return fmt.Errorf("升级信使等级失败: %w", err)
		}
	}

	// 发送通知
	s.notifyUpgradeResult(&request)

	return nil
}

// upgradeCourier 升级信使等级
func (s *CourierLevelService) upgradeCourier(courierID string, newLevel models.CourierLevel) error {
	return s.db.Model(&models.Courier{}).
		Where("user_id = ?", courierID).
		Update("level", int(newLevel)).Error
}

// GetCourierZones 获取信使管理区域
func (s *CourierLevelService) GetCourierZones(courierID string) ([]models.CourierZone, error) {
	var zones []models.CourierZone
	if err := s.db.Where("courier_id = ? AND is_active = ?", courierID, true).Find(&zones).Error; err != nil {
		return nil, fmt.Errorf("查询管理区域失败: %w", err)
	}
	return zones, nil
}

// CanAssignZone 检查是否可以分配区域
func (s *CourierLevelService) CanAssignZone(assignerID, targetCourierID string, zoneType models.CourierZoneType) (bool, string) {
	// 获取分配者等级
	var assigner models.Courier
	if err := s.db.Where("user_id = ?", assignerID).First(&assigner).Error; err != nil {
		return false, "分配者不存在"
	}

	assignerLevel := models.CourierLevel(assigner.Level)

	// 获取目标信使等级
	var target models.Courier
	if err := s.db.Where("user_id = ?", targetCourierID).First(&target).Error; err != nil {
		return false, "目标信使不存在"
	}

	targetLevel := models.CourierLevel(target.Level)

	// 检查权限：只能给同级或下级分配区域
	if assignerLevel < targetLevel {
		return false, "无权限给更高等级信使分配区域"
	}

	// 检查区域类型权限
	assignerZoneType := models.DefaultZoneMapping[assignerLevel]
	targetZoneType := models.DefaultZoneMapping[targetLevel]

	// 分配者的区域类型必须高于或等于目标区域类型
	zoneHierarchy := map[models.CourierZoneType]int{
		models.ZoneBuilding: 1,
		models.ZoneArea:     2,
		models.ZoneCampus:   3,
		models.ZoneCity:     4,
	}

	if zoneHierarchy[assignerZoneType] < zoneHierarchy[zoneType] {
		return false, "分配者无权限管理该类型区域"
	}

	if zoneHierarchy[targetZoneType] < zoneHierarchy[zoneType] {
		return false, "目标信使等级不足以管理该类型区域"
	}

	return true, ""
}

// AssignZone 分配管理区域
func (s *CourierLevelService) AssignZone(courierID string, zoneType models.CourierZoneType, zoneID, zoneName, assignerID string) error {
	// 检查是否已分配该区域
	var existingZone models.CourierZone
	if err := s.db.Where("courier_id = ? AND zone_type = ? AND zone_id = ? AND is_active = ?",
		courierID, zoneType, zoneID, true).First(&existingZone).Error; err == nil {
		return errors.New("该区域已分配给此信使")
	}

	zone := &models.CourierZone{
		CourierID:  courierID,
		ZoneType:   zoneType,
		ZoneID:     zoneID,
		ZoneName:   zoneName,
		IsActive:   true,
		AssignedAt: time.Now(),
		AssignedBy: assignerID,
	}

	if err := s.db.Create(zone).Error; err != nil {
		return fmt.Errorf("分配区域失败: %w", err)
	}

	// 发送通知
	s.notifyZoneAssignment(zone)

	return nil
}

// GetViewablePerformanceScope 获取可查看的绩效范围
func (s *CourierLevelService) GetViewablePerformanceScope(courierID string) (bool, []models.CourierZone) {
	// 获取信使等级
	var courier models.Courier
	if err := s.db.Where("user_id = ?", courierID).First(&courier).Error; err != nil {
		return false, nil
	}

	level := models.CourierLevel(courier.Level)

	// 检查是否有绩效查看权限
	permissions := models.DefaultPermissionMatrix[level]
	hasPermission := false
	for _, perm := range permissions {
		if perm == models.PermissionPerformance {
			hasPermission = true
			break
		}
	}

	if !hasPermission {
		return false, nil
	}

	// 获取管理区域
	var zones []models.CourierZone
	s.db.Where("courier_id = ? AND is_active = ?", courierID, true).Find(&zones)

	return true, zones
}

// GetPerformanceData 获取绩效数据
func (s *CourierLevelService) GetPerformanceData(_ string, zones []models.CourierZone, timeRange, zoneType, zoneID string) (map[string]interface{}, error) {
	// 解析时间范围
	var since time.Time
	switch timeRange {
	case "1d":
		since = time.Now().AddDate(0, 0, -1)
	case "7d":
		since = time.Now().AddDate(0, 0, -7)
	case "30d":
		since = time.Now().AddDate(0, 0, -30)
	default:
		since = time.Now().AddDate(0, 0, -7)
	}

	// 构建查询条件
	query := s.db.Model(&models.Task{}).Where("created_at > ?", since)

	// 如果指定了区域，过滤区域
	if zoneType != "" && zoneID != "" {
		// 这里需要根据实际的数据结构来过滤
		// 简化实现，假设任务表中有zone相关字段
	}

	// 统计数据
	var totalTasks, completedTasks, failedTasks int64
	query.Count(&totalTasks)
	query.Where("status = ?", "delivered").Count(&completedTasks)
	query.Where("status = ?", "failed").Count(&failedTasks)

	completionRate := float64(0)
	if totalTasks > 0 {
		completionRate = float64(completedTasks) / float64(totalTasks) * 100
	}

	return map[string]interface{}{
		"time_range":      timeRange,
		"total_tasks":     totalTasks,
		"completed_tasks": completedTasks,
		"failed_tasks":    failedTasks,
		"completion_rate": completionRate,
		"zones":           zones,
	}, nil
}

// 通知相关方法
func (s *CourierLevelService) notifyUpgradeRequest(request *models.LevelUpgradeRequest) {
	event := utils.WebSocketEvent{
		Type: "COURIER_UPGRADE_REQUEST",
		Data: map[string]interface{}{
			"request_id":    request.ID,
			"courier_id":    request.CourierID,
			"current_level": request.CurrentLevel,
			"request_level": request.RequestLevel,
		},
		Timestamp: time.Now(),
	}
	s.wsManager.BroadcastToAdmins(event)
}

func (s *CourierLevelService) notifyUpgradeResult(request *models.LevelUpgradeRequest) {
	event := utils.WebSocketEvent{
		Type: "COURIER_UPGRADE_RESULT",
		Data: map[string]interface{}{
			"request_id": request.ID,
			"courier_id": request.CourierID,
			"status":     request.Status,
			"new_level":  request.RequestLevel,
		},
		Timestamp: time.Now(),
	}
	s.wsManager.BroadcastToUser(request.CourierID, event)
}

func (s *CourierLevelService) notifyZoneAssignment(zone *models.CourierZone) {
	event := utils.WebSocketEvent{
		Type: "ZONE_ASSIGNMENT",
		Data: map[string]interface{}{
			"courier_id": zone.CourierID,
			"zone_type":  zone.ZoneType,
			"zone_id":    zone.ZoneID,
			"zone_name":  zone.ZoneName,
		},
		Timestamp: time.Now(),
	}
	s.wsManager.BroadcastToUser(zone.CourierID, event)
}
