package services

import (
	"courier-service/internal/models"
	"courier-service/internal/utils"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// HierarchyService 层级管理服务
type HierarchyService struct {
	db        *gorm.DB
	wsManager *utils.WebSocketManager
}

// NewHierarchyService 创建层级管理服务
func NewHierarchyService(db *gorm.DB, wsManager *utils.WebSocketManager) *HierarchyService {
	return &HierarchyService{
		db:        db,
		wsManager: wsManager,
	}
}

// GetDB 获取数据库连接
func (s *HierarchyService) GetDB() *gorm.DB {
	return s.db
}

// CreateSubordinate 创建下级信使
func (s *HierarchyService) CreateSubordinate(managerID string, req *models.CreateSubordinateRequest) (*models.Courier, error) {
	// 获取管理者信息
	var manager models.Courier
	if err := s.db.First(&manager, managerID).Error; err != nil {
		return nil, fmt.Errorf("管理者不存在: %w", err)
	}

	// 检查是否有权限创建下级
	if !manager.CanCreateSubordinate() {
		return nil, errors.New("当前等级无权创建下级信使")
	}

	// 验证层级关系
	if !manager.CanManageSubordinate(req.Level) {
		return nil, errors.New("只能创建比自己等级低一级的信使")
	}

	// 检查用户是否已经是信使
	var existingCourier models.Courier
	if err := s.db.Where("user_id = ?", req.UserID).First(&existingCourier).Error; err == nil {
		return nil, errors.New("该用户已经是信使")
	}

	// 验证区域编码格式
	if !s.validateZoneCode(req.ZoneCode, req.ZoneType) {
		return nil, errors.New("区域编码格式不正确")
	}

	// 创建新的信使
	subordinate := &models.Courier{
		UserID:      req.UserID,
		Zone:        fmt.Sprintf("%s-%s", req.ZoneType, req.ZoneCode),
		Phone:       req.Phone,
		IDCard:      req.IDCard,
		Status:      models.CourierStatusApproved, // 上级创建的直接审核通过
		Level:       req.Level,
		Experience:  req.Experience,
		Note:        req.Note,
		ParentID:    &managerID,
		ZoneCode:    req.ZoneCode,
		ZoneType:    req.ZoneType,
		Points:      0,
		CreatedByID: &managerID,
		ApprovedAt:  &time.Time{},
	}

	if err := s.db.Create(subordinate).Error; err != nil {
		return nil, fmt.Errorf("创建下级信使失败: %w", err)
	}

	// 发送通知
	s.notifySubordinateCreated(manager, subordinate)

	return subordinate, nil
}

// GetSubordinates 获取下级信使列表
func (s *HierarchyService) GetSubordinates(managerID string) ([]models.Courier, error) {
	var subordinates []models.Courier
	err := s.db.Where("parent_id = ?", managerID).
		Order("level ASC, points DESC").
		Find(&subordinates).Error

	if err != nil {
		return nil, fmt.Errorf("获取下级信使失败: %w", err)
	}

	return subordinates, nil
}

// AssignZone 分配管理区域
func (s *HierarchyService) AssignZone(managerID string, courierID string, req *models.AssignZoneRequest) error {
	// 验证管理权限
	if !s.canManageCourier(managerID, courierID) {
		return errors.New("无权限管理该信使")
	}

	// 获取信使信息
	var courier models.Courier
	if err := s.db.First(&courier, courierID).Error; err != nil {
		return fmt.Errorf("信使不存在: %w", err)
	}

	// 验证区域编码
	if !s.validateZoneCode(req.ZoneCode, req.ZoneType) {
		return errors.New("区域编码格式不正确")
	}

	// 检查区域是否已被其他信使占用
	var existingCourier models.Courier
	if err := s.db.Where("zone_code = ? AND id != ?", req.ZoneCode, courierID).First(&existingCourier).Error; err == nil {
		return errors.New("该区域已被其他信使管理")
	}

	// 更新区域分配
	updates := map[string]interface{}{
		"zone_code": req.ZoneCode,
		"zone_type": req.ZoneType,
		"zone":      fmt.Sprintf("%s-%s", req.ZoneType, req.ZoneCode),
	}

	if err := s.db.Model(&courier).Updates(updates).Error; err != nil {
		return fmt.Errorf("分配区域失败: %w", err)
	}

	// 发送通知
	s.notifyZoneAssigned(courierID, req.ZoneCode, req.ZoneType)

	return nil
}

// TransferSubordinate 转移下级信使归属
func (s *HierarchyService) TransferSubordinate(managerID string, courierID string, req *models.TransferSubordinateRequest) error {
	// 验证当前管理权限
	if !s.canManageCourier(managerID, courierID) {
		return errors.New("无权限转移该信使")
	}

	// 获取目标管理者信息
	var newManager models.Courier
	if err := s.db.First(&newManager, req.NewParentID).Error; err != nil {
		return fmt.Errorf("目标管理者不存在: %w", err)
	}

	// 获取被转移的信使信息
	var courier models.Courier
	if err := s.db.First(&courier, courierID).Error; err != nil {
		return fmt.Errorf("信使不存在: %w", err)
	}

	// 检查目标管理者是否有权限管理该等级
	if !newManager.CanManageSubordinate(courier.Level) {
		return errors.New("目标管理者无权限管理该等级信使")
	}

	// 执行转移
	if err := s.db.Model(&courier).Update("parent_id", req.NewParentID).Error; err != nil {
		return fmt.Errorf("转移失败: %w", err)
	}

	// 记录操作日志
	s.logHierarchyOperation(managerID, courierID, "transfer", req.Reason)

	// 发送通知
	s.notifySubordinateTransferred(managerID, req.NewParentID, courierID, req.Reason)

	return nil
}

// GetHierarchy 获取层级结构
func (s *HierarchyService) GetHierarchy(courierID string) (*models.CourierHierarchyResponse, error) {
	var courier models.Courier
	if err := s.db.Preload("Parent").Preload("Subordinates").First(&courier, courierID).Error; err != nil {
		return nil, fmt.Errorf("信使不存在: %w", err)
	}

	// 构建权限列表
	permissions := s.getPermissions(courier.Level)
	canManage := s.getManageableOperations(courier.Level)

	return &models.CourierHierarchyResponse{
		Courier:      courier,
		Parent:       courier.Parent,
		Subordinates: courier.Subordinates,
		Level:        courier.Level,
		CanManage:    canManage,
		Permissions:  permissions,
	}, nil
}

// 私有方法

// canManageCourier 检查是否可以管理指定信使
func (s *HierarchyService) canManageCourier(managerID string, courierID string) bool {
	var courier models.Courier
	if err := s.db.First(&courier, courierID).Error; err != nil {
		return false
	}

	return courier.IsSubordinateOf(managerID)
}

// validateZoneCode 验证区域编码格式
func (s *HierarchyService) validateZoneCode(zoneCode string, zoneType string) bool {
	// 简单的格式验证，实际项目中可能需要更复杂的规则
	if len(zoneCode) < 3 {
		return false
	}

	switch zoneType {
	case models.ZoneTypeBuilding:
		// 楼栋编码格式: SCHOOL001-BUILDING001
		return len(zoneCode) >= 10
	case models.ZoneTypeArea:
		// 片区编码格式: SCHOOL001-AREA001
		return len(zoneCode) >= 10
	case models.ZoneTypeSchool:
		// 学校编码格式: SCHOOL001
		return len(zoneCode) >= 6
	case models.ZoneTypeCity:
		// 城市编码格式: CITY001
		return len(zoneCode) >= 4
	default:
		return false
	}
}

// getPermissions 获取指定等级的权限列表
func (s *HierarchyService) getPermissions(level int) []string {
	basePermissions := []string{"scan_code", "view_tasks", "update_status"}

	switch level {
	case models.CourierLevelOne:
		return basePermissions
	case models.CourierLevelTwo:
		return append(basePermissions, "manage_level_one", "assign_tasks")
	case models.CourierLevelThree:
		return append(basePermissions, "manage_level_two", "manage_level_one", "assign_zones", "view_statistics")
	case models.CourierLevelFour:
		return append(basePermissions, "manage_all_levels", "system_administration", "create_subordinates")
	default:
		return basePermissions
	}
}

// getManageableOperations 获取可管理的操作列表
func (s *HierarchyService) getManageableOperations(level int) []string {
	operations := []string{}

	if level >= models.CourierLevelTwo {
		operations = append(operations, "create_subordinate", "assign_task")
	}
	if level >= models.CourierLevelThree {
		operations = append(operations, "assign_zone", "view_subordinate_stats")
	}
	if level >= models.CourierLevelFour {
		operations = append(operations, "transfer_subordinate", "system_config")
	}

	return operations
}

// 通知方法

// notifySubordinateCreated 通知下级信使创建成功
func (s *HierarchyService) notifySubordinateCreated(manager models.Courier, subordinate *models.Courier) {
	event := utils.WebSocketEvent{
		Type: "SUBORDINATE_CREATED",
		Data: map[string]interface{}{
			"manager_id":     manager.ID,
			"subordinate_id": subordinate.ID,
			"level":          subordinate.Level,
			"zone_code":      subordinate.ZoneCode,
		},
		Timestamp: time.Now(),
	}

	// 通知管理者
	s.wsManager.BroadcastToUser(manager.UserID, event)
	// 通知新信使
	s.wsManager.BroadcastToUser(subordinate.UserID, event)
}

// notifyZoneAssigned 通知区域分配成功
func (s *HierarchyService) notifyZoneAssigned(courierID string, zoneCode string, zoneType string) {
	event := utils.WebSocketEvent{
		Type: "ZONE_ASSIGNED",
		Data: map[string]interface{}{
			"courier_id": courierID,
			"zone_code":  zoneCode,
			"zone_type":  zoneType,
		},
		Timestamp: time.Now(),
	}

	var courier models.Courier
	if s.db.First(&courier, courierID).Error == nil {
		s.wsManager.BroadcastToUser(courier.UserID, event)
	}
}

// notifySubordinateTransferred 通知信使转移成功
func (s *HierarchyService) notifySubordinateTransferred(oldManagerID string, newManagerID string, courierID string, reason string) {
	event := utils.WebSocketEvent{
		Type: "SUBORDINATE_TRANSFERRED",
		Data: map[string]interface{}{
			"old_manager_id": oldManagerID,
			"new_manager_id": newManagerID,
			"courier_id":     courierID,
			"reason":         reason,
		},
		Timestamp: time.Now(),
	}

	// 通知所有相关人员
	var oldManager, newManager, courier models.Courier
	if s.db.First(&oldManager, oldManagerID).Error == nil {
		s.wsManager.BroadcastToUser(oldManager.UserID, event)
	}
	if s.db.First(&newManager, newManagerID).Error == nil {
		s.wsManager.BroadcastToUser(newManager.UserID, event)
	}
	if s.db.First(&courier, courierID).Error == nil {
		s.wsManager.BroadcastToUser(courier.UserID, event)
	}
}

// logHierarchyOperation 记录层级操作日志
func (s *HierarchyService) logHierarchyOperation(operatorID string, targetID string, operation string, reason string) {
	// 这里可以实现操作日志记录
	// 简化实现，实际项目中应该有专门的日志表
	fmt.Printf("[HIERARCHY_LOG] Operator: %s, Target: %s, Operation: %s, Reason: %s, Time: %s\n",
		operatorID, targetID, operation, reason, time.Now().Format(time.RFC3339))
}
