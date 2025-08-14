package services

import (
	"courier-service/internal/models"
	"courier-service/internal/utils"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// PostalManagementService 编号分配权限控制服务
type PostalManagementService struct {
	db        *gorm.DB
	redis     *redis.Client
	wsManager *utils.WebSocketManager
}

// NewPostalManagementService 创建编号管理服务
func NewPostalManagementService(db *gorm.DB, redis *redis.Client, wsManager *utils.WebSocketManager) *PostalManagementService {
	return &PostalManagementService{
		db:        db,
		redis:     redis,
		wsManager: wsManager,
	}
}

// GetPendingApplications 获取权限范围内待审核编号申请
func (s *PostalManagementService) GetPendingApplications(courierID, schoolID, areaID, status string, limit, offset int) ([]models.PostalCodeApplication, int64, error) {
	// 获取信使权限范围
	scope, err := s.GetPermissionScope(courierID)
	if err != nil {
		return nil, 0, err
	}

	// 构建查询条件
	query := s.db.Model(&models.PostalCodeApplication{})

	// 应用权限过滤
	var schoolConditions []string
	for _, school := range scope.Schools {
		if schoolID != "" && school.SchoolID != schoolID {
			continue
		}

		for _, area := range school.Areas {
			if areaID != "" && area.AreaID != areaID {
				continue
			}

			condition := fmt.Sprintf("(school_id = '%s' AND area_id = '%s')", school.SchoolID, area.AreaID)
			schoolConditions = append(schoolConditions, condition)
		}
	}

	if len(schoolConditions) == 0 {
		return []models.PostalCodeApplication{}, 0, nil
	}

	query = query.Where(strings.Join(schoolConditions, " OR "))

	// 状态过滤
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 获取总数
	var total int64
	query.Count(&total)

	// 获取分页数据
	var applications []models.PostalCodeApplication
	err = query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&applications).Error
	if err != nil {
		return nil, 0, fmt.Errorf("查询申请失败: %w", err)
	}

	return applications, total, nil
}

// CanReviewApplication 检查是否可以审核申请
func (s *PostalManagementService) CanReviewApplication(reviewerID, applicationID string) (bool, error) {
	// 获取申请信息
	var application models.PostalCodeApplication
	if err := s.db.Where("id = ?", applicationID).First(&application).Error; err != nil {
		return false, fmt.Errorf("申请不存在: %w", err)
	}

	// 获取审核者权限范围
	scope, err := s.GetPermissionScope(reviewerID)
	if err != nil {
		return false, err
	}

	// 检查是否有权限审核该申请
	for _, school := range scope.Schools {
		if school.SchoolID == application.SchoolID {
			for _, area := range school.Areas {
				if area.AreaID == application.AreaID {
					return s.hasPermissionForAction(scope.Level, "approve", area.AreaID), nil
				}
			}
		}
	}

	return false, nil
}

// ReviewApplication 审核编号申请
func (s *PostalManagementService) ReviewApplication(applicationID, action, assignedCode, comment, reviewerID string) (map[string]interface{}, error) {
	// 获取申请信息
	var application models.PostalCodeApplication
	if err := s.db.Where("id = ?", applicationID).First(&application).Error; err != nil {
		return nil, fmt.Errorf("申请不存在: %w", err)
	}

	// 检查状态转换是否有效
	var newStatus models.PostalCodeStatus
	switch action {
	case "approve":
		newStatus = models.PostalCodeStatusApproved
		if assignedCode == "" {
			assignedCode = s.generateNextCode(application.SchoolID, application.AreaID)
		}
		application.AssignedCode = assignedCode
	case "reject":
		newStatus = models.PostalCodeStatusRejected
	default:
		return nil, errors.New("无效的审核动作")
	}

	if !application.Status.CanTransitionTo(newStatus) {
		return nil, errors.New("无效的状态转换")
	}

	// 更新申请状态
	now := time.Now()
	application.Status = newStatus
	application.ReviewerID = &reviewerID
	application.ReviewedAt = &now
	application.ReviewComment = comment

	if err := s.db.Save(&application).Error; err != nil {
		return nil, fmt.Errorf("更新申请状态失败: %w", err)
	}

	// 如果批准且分配了编号，创建编号分配记录
	if action == "approve" && assignedCode != "" {
		assignment := &models.PostalCodeAssignment{
			UserID:     application.UserID,
			PostalCode: assignedCode,
			SchoolID:   application.SchoolID,
			AreaID:     application.AreaID,
			AssignedBy: reviewerID,
			AssignedAt: time.Now(),
			IsActive:   true,
		}

		if err := s.db.Create(assignment).Error; err != nil {
			return nil, fmt.Errorf("创建编号分配记录失败: %w", err)
		}

		// 更新申请状态为已分配
		application.Status = models.PostalCodeStatusAssigned
		s.db.Save(&application)
	}

	// 发送通知
	s.notifyApplicationReviewed(&application, action, reviewerID)

	return map[string]interface{}{
		"application_id": applicationID,
		"action":         action,
		"status":         application.Status,
		"assigned_code":  assignedCode,
		"reviewed_by":    reviewerID,
		"reviewed_at":    application.ReviewedAt,
		"message":        fmt.Sprintf("申请已%s", getActionName(action)),
	}, nil
}

// GetAssignedCodes 获取权限范围内已分配编号
func (s *PostalManagementService) GetAssignedCodes(courierID string, filters map[string]interface{}) ([]models.PostalCodeAssignment, int64, error) {
	// 获取权限范围
	scope, err := s.GetPermissionScope(courierID)
	if err != nil {
		return nil, 0, err
	}

	// 构建查询条件
	query := s.db.Model(&models.PostalCodeAssignment{})

	// 应用权限过滤
	var conditions []string
	for _, school := range scope.Schools {
		schoolID := school.SchoolID
		if filterSchoolID, ok := filters["school_id"].(string); ok && filterSchoolID != "" && filterSchoolID != schoolID {
			continue
		}

		for _, area := range school.Areas {
			areaID := area.AreaID
			if filterAreaID, ok := filters["area_id"].(string); ok && filterAreaID != "" && filterAreaID != areaID {
				continue
			}

			condition := fmt.Sprintf("(school_id = '%s' AND area_id = '%s')", schoolID, areaID)
			conditions = append(conditions, condition)
		}
	}

	if len(conditions) == 0 {
		return []models.PostalCodeAssignment{}, 0, nil
	}

	query = query.Where(strings.Join(conditions, " OR "))

	// 应用其他过滤条件
	if buildingID, ok := filters["building_id"].(string); ok && buildingID != "" {
		query = query.Where("building_id = ?", buildingID)
	}

	if userID, ok := filters["user_id"].(string); ok && userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	if isActive, ok := filters["is_active"].(*bool); ok && isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	// 获取总数
	var total int64
	query.Count(&total)

	// 分页查询
	limit := filters["limit"].(int)
	offset := filters["offset"].(int)

	var assignments []models.PostalCodeAssignment
	err = query.Limit(limit).Offset(offset).Order("assigned_at DESC").Find(&assignments).Error
	if err != nil {
		return nil, 0, fmt.Errorf("查询分配记录失败: %w", err)
	}

	return assignments, total, nil
}

// CanBatchAssign 检查是否可以批量分配编号
func (s *PostalManagementService) CanBatchAssign(courierID, schoolID, areaID string) (bool, error) {
	scope, err := s.GetPermissionScope(courierID)
	if err != nil {
		return false, err
	}

	// 检查是否有批量分配权限（通常需要三级以上信使）
	if scope.Level < models.LevelThree {
		return false, nil
	}

	// 检查对特定区域的权限
	for _, school := range scope.Schools {
		if school.SchoolID == schoolID {
			for _, area := range school.Areas {
				if area.AreaID == areaID {
					return s.hasPermissionForAction(scope.Level, "batch_assign", areaID), nil
				}
			}
		}
	}

	return false, nil
}

// BatchAssignCodes 批量分配编号
func (s *PostalManagementService) BatchAssignCodes(assignerID, schoolID, areaID string, assignments []models.PostalCodeAssignmentItem) (map[string]interface{}, error) {
	var results []map[string]interface{}
	var successCount, failureCount int

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, item := range assignments {
		// 检查编号是否已被使用
		var existingAssignment models.PostalCodeAssignment
		if err := tx.Where("postal_code = ? AND is_active = ?", item.PostalCode, true).First(&existingAssignment).Error; err == nil {
			results = append(results, map[string]interface{}{
				"user_id":     item.UserID,
				"postal_code": item.PostalCode,
				"status":      "failed",
				"reason":      "编号已被使用",
			})
			failureCount++
			continue
		}

		// 创建分配记录
		assignment := &models.PostalCodeAssignment{
			UserID:     item.UserID,
			PostalCode: item.PostalCode,
			SchoolID:   schoolID,
			AreaID:     areaID,
			BuildingID: &item.BuildingID,
			RoomNumber: &item.RoomNumber,
			AssignedBy: assignerID,
			AssignedAt: time.Now(),
			IsActive:   true,
		}

		if err := tx.Create(assignment).Error; err != nil {
			results = append(results, map[string]interface{}{
				"user_id":     item.UserID,
				"postal_code": item.PostalCode,
				"status":      "failed",
				"reason":      err.Error(),
			})
			failureCount++
			continue
		}

		results = append(results, map[string]interface{}{
			"user_id":     item.UserID,
			"postal_code": item.PostalCode,
			"status":      "success",
			"assigned_at": assignment.AssignedAt,
		})
		successCount++
	}

	if failureCount > 0 && successCount == 0 {
		tx.Rollback()
		return nil, errors.New("批量分配全部失败")
	}

	tx.Commit()

	// 发送通知
	s.notifyBatchAssignment(assignerID, schoolID, areaID, successCount, failureCount)

	return map[string]interface{}{
		"total_count":   len(assignments),
		"success_count": successCount,
		"failure_count": failureCount,
		"results":       results,
		"assigned_by":   assignerID,
		"school_id":     schoolID,
		"area_id":       areaID,
	}, nil
}

// CanAssignCode 检查是否可以分配编号
func (s *PostalManagementService) CanAssignCode(courierID, schoolID, areaID, buildingID string) (bool, error) {
	scope, err := s.GetPermissionScope(courierID)
	if err != nil {
		return false, err
	}

	// 检查对特定区域的分配权限
	for _, school := range scope.Schools {
		if school.SchoolID == schoolID {
			for _, area := range school.Areas {
				if area.AreaID == areaID {
					// 根据建筑ID确定需要的权限级别
					requiredAction := "assign_area"
					if buildingID != "" {
						requiredAction = "assign_building"
					}
					return s.hasPermissionForAction(scope.Level, requiredAction, areaID), nil
				}
			}
		}
	}

	return false, nil
}

// AssignSingleCode 分配单个编号
func (s *PostalManagementService) AssignSingleCode(assignerID, userID, postalCode, schoolID, areaID, buildingID, roomNumber string) (*models.PostalCodeAssignment, error) {
	// 检查编号是否已被使用
	var existingAssignment models.PostalCodeAssignment
	if err := s.db.Where("postal_code = ? AND is_active = ?", postalCode, true).First(&existingAssignment).Error; err == nil {
		return nil, errors.New("编号已被使用")
	}

	// 检查用户是否已有编号
	if err := s.db.Where("user_id = ? AND is_active = ?", userID, true).First(&existingAssignment).Error; err == nil {
		return nil, errors.New("用户已有活跃编号")
	}

	// 创建分配记录
	assignment := &models.PostalCodeAssignment{
		UserID:     userID,
		PostalCode: postalCode,
		SchoolID:   schoolID,
		AreaID:     areaID,
		AssignedBy: assignerID,
		AssignedAt: time.Now(),
		IsActive:   true,
	}

	if buildingID != "" {
		assignment.BuildingID = &buildingID
	}
	if roomNumber != "" {
		assignment.RoomNumber = &roomNumber
	}

	if err := s.db.Create(assignment).Error; err != nil {
		return nil, fmt.Errorf("创建分配记录失败: %w", err)
	}

	// 发送通知
	s.notifyCodeAssigned(assignment)

	return assignment, nil
}

// CanDeactivateCode 检查是否可以停用编号
func (s *PostalManagementService) CanDeactivateCode(operatorID, assignmentID string) (bool, error) {
	// 获取分配记录
	var assignment models.PostalCodeAssignment
	if err := s.db.Where("id = ?", assignmentID).First(&assignment).Error; err != nil {
		return false, fmt.Errorf("分配记录不存在: %w", err)
	}

	// 获取操作者权限
	scope, err := s.GetPermissionScope(operatorID)
	if err != nil {
		return false, err
	}

	// 检查是否有权限停用该区域的编号
	for _, school := range scope.Schools {
		if school.SchoolID == assignment.SchoolID {
			for _, area := range school.Areas {
				if area.AreaID == assignment.AreaID {
					return s.hasPermissionForAction(scope.Level, "assign", area.AreaID), nil
				}
			}
		}
	}

	return false, nil
}

// DeactivateCode 停用编号
func (s *PostalManagementService) DeactivateCode(assignmentID, operatorID, reason string) error {
	// 获取分配记录
	var assignment models.PostalCodeAssignment
	if err := s.db.Where("id = ?", assignmentID).First(&assignment).Error; err != nil {
		return fmt.Errorf("分配记录不存在: %w", err)
	}

	if !assignment.IsActive {
		return errors.New("编号已被停用")
	}

	// 更新状态
	now := time.Now()
	assignment.IsActive = false
	assignment.DeactivatedBy = &operatorID
	assignment.DeactivatedAt = &now

	if err := s.db.Save(&assignment).Error; err != nil {
		return fmt.Errorf("更新分配记录失败: %w", err)
	}

	// 发送通知
	s.notifyCodeDeactivated(&assignment, operatorID, reason)

	return nil
}

// GetPermissionScope 获取编号管理权限范围
func (s *PostalManagementService) GetPermissionScope(courierID string) (*models.PostalCodePermissionScope, error) {
	// 获取信使信息
	var courier models.Courier
	if err := s.db.Where("user_id = ?", courierID).First(&courier).Error; err != nil {
		return nil, fmt.Errorf("信使不存在: %w", err)
	}

	level := models.CourierLevel(courier.Level)

	// 获取管理的区域
	var zones []models.CourierZone
	s.db.Where("courier_id = ? AND is_active = ?", courierID, true).Find(&zones)

	// 构建权限范围
	scope := &models.PostalCodePermissionScope{
		CourierID:  courierID,
		Level:      level,
		CanManage:  make([]string, 0),
		CanAssign:  make([]string, 0),
		CanApprove: make([]string, 0),
		Schools:    make([]models.SchoolInfo, 0),
	}

	// 根据等级和管理区域构建权限范围
	schoolMap := make(map[string]*models.SchoolInfo)

	for _, zone := range zones {
		// 根据区域类型和等级确定权限
		if s.hasPermissionForAction(level, "manage", zone.ZoneID) {
			scope.CanManage = append(scope.CanManage, zone.ZoneID)
		}
		if s.hasPermissionForAction(level, "assign", zone.ZoneID) {
			scope.CanAssign = append(scope.CanAssign, zone.ZoneID)
		}
		if s.hasPermissionForAction(level, "approve", zone.ZoneID) {
			scope.CanApprove = append(scope.CanApprove, zone.ZoneID)
		}

		// 根据实际业务逻辑构建学校和区域信息
		// 这里需要根据实际的数据结构实现
		schoolInfo := s.buildSchoolInfoFromZone(zone)
		if schoolInfo != nil {
			if existing, exists := schoolMap[schoolInfo.SchoolID]; exists {
				// 合并区域信息
				existing.Areas = append(existing.Areas, schoolInfo.Areas...)
			} else {
				schoolMap[schoolInfo.SchoolID] = schoolInfo
			}
		}
	}

	// 转换为切片
	for _, school := range schoolMap {
		scope.Schools = append(scope.Schools, *school)
	}

	return scope, nil
}

// buildSchoolInfoFromZone 根据区域构建学校信息 (简化实现)
func (s *PostalManagementService) buildSchoolInfoFromZone(zone models.CourierZone) *models.SchoolInfo {
	// 这里需要根据实际的业务逻辑实现
	// 简化实现，假设区域ID包含学校信息
	parts := strings.Split(zone.ZoneID, "_")
	if len(parts) < 2 {
		return nil
	}

	schoolID := parts[0]
	areaID := zone.ZoneID

	return &models.SchoolInfo{
		SchoolID:   schoolID,
		SchoolName: zone.ZoneName, // 简化实现
		Areas: []models.AreaInfo{
			{
				AreaID:    areaID,
				AreaName:  zone.ZoneName,
				ManagerID: courierIDFromZone(zone),
				CodeRange: "001-999", // 简化实现
			},
		},
	}
}

// courierIDFromZone 从区域获取信使ID (简化实现)
func courierIDFromZone(zone models.CourierZone) string {
	return zone.CourierID
}

// hasPermissionForAction 检查是否有特定动作的权限
func (s *PostalManagementService) hasPermissionForAction(level models.CourierLevel, action, _ string) bool {
	permissions, exists := models.PostalCodePermissionMatrix[level]
	if !exists {
		return false
	}

	for _, permission := range permissions {
		if strings.Contains(permission, action) {
			return true
		}
	}

	return false
}

// generateNextCode 生成下一个可用编号
func (s *PostalManagementService) generateNextCode(schoolID, areaID string) string {
	// 获取学校规则
	var rule models.PostalCodeRule
	if err := s.db.Where("school_id = ?", schoolID).First(&rule).Error; err != nil {
		return fmt.Sprintf("%s%s%03d", strings.ToUpper(schoolID[:3]), areaID, 1)
	}

	// 解析区域规则
	var areaRules map[string]interface{}
	json.Unmarshal([]byte(rule.AreaRules), &areaRules)

	areas, ok := areaRules["areas"].([]interface{})
	if !ok {
		return fmt.Sprintf("%s%s%03d", rule.Prefix, areaID, 1)
	}

	// 查找对应区域的编号范围
	for _, area := range areas {
		areaMap := area.(map[string]interface{})
		if areaMap["id"].(string) == areaID {
			codeRange := areaMap["range"].(string)
			parts := strings.Split(codeRange, "-")
			if len(parts) == 2 {
				start, _ := strconv.Atoi(parts[0])
				end, _ := strconv.Atoi(parts[1])

				// 查找已使用的最大编号
				var maxUsed int
				var assignments []models.PostalCodeAssignment
				s.db.Where("school_id = ? AND area_id = ? AND is_active = ?", schoolID, areaID, true).Find(&assignments)

				for _, assignment := range assignments {
					// 提取编号中的数字部分
					code := assignment.PostalCode
					if len(code) >= 3 {
						numPart := code[len(code)-3:]
						if num, err := strconv.Atoi(numPart); err == nil && num > maxUsed {
							maxUsed = num
						}
					}
				}

				nextNum := maxUsed + 1
				if nextNum >= start && nextNum <= end {
					return fmt.Sprintf("%s%s%03d", rule.Prefix, areaID, nextNum)
				}
			}
		}
	}

	return fmt.Sprintf("%s%s%03d", rule.Prefix, areaID, 1)
}

// GetStatistics 获取统计信息
func (s *PostalManagementService) GetStatistics(courierID, schoolID, areaID, timeRange string) (map[string]interface{}, error) {
	scope, err := s.GetPermissionScope(courierID)
	if err != nil {
		return nil, err
	}

	var statistics []models.PostalCodeStatistics

	for _, school := range scope.Schools {
		if schoolID != "" && school.SchoolID != schoolID {
			continue
		}

		for _, area := range school.Areas {
			if areaID != "" && area.AreaID != areaID {
				continue
			}

			// 计算统计数据
			var totalAssigned, activeAssigned int64
			s.db.Model(&models.PostalCodeAssignment{}).
				Where("school_id = ? AND area_id = ?", school.SchoolID, area.AreaID).
				Count(&totalAssigned)

			s.db.Model(&models.PostalCodeAssignment{}).
				Where("school_id = ? AND area_id = ? AND is_active = ?", school.SchoolID, area.AreaID, true).
				Count(&activeAssigned)

			var pendingApps int64
			s.db.Model(&models.PostalCodeApplication{}).
				Where("school_id = ? AND area_id = ? AND status = ?", school.SchoolID, area.AreaID, "pending").
				Count(&pendingApps)

			// 假设总编号数为999
			totalCodes := 999
			utilizationRate := float64(activeAssigned) / float64(totalCodes) * 100

			stat := models.PostalCodeStatistics{
				SchoolID:        school.SchoolID,
				SchoolName:      school.SchoolName,
				TotalCodes:      totalCodes,
				AssignedCodes:   int(activeAssigned),
				UnassignedCodes: totalCodes - int(activeAssigned),
				PendingApps:     int(pendingApps),
				UtilizationRate: utilizationRate,
			}

			statistics = append(statistics, stat)
		}
	}

	return map[string]interface{}{
		"time_range":  timeRange,
		"statistics":  statistics,
		"total_count": len(statistics),
	}, nil
}

// SearchCodes 搜索编号
func (s *PostalManagementService) SearchCodes(courierID string, filters map[string]interface{}) ([]models.PostalCodeAssignment, error) {
	scope, err := s.GetPermissionScope(courierID)
	if err != nil {
		return nil, err
	}

	query := s.db.Model(&models.PostalCodeAssignment{})

	// 应用权限过滤
	var conditions []string
	for _, school := range scope.Schools {
		for _, area := range school.Areas {
			condition := fmt.Sprintf("(school_id = '%s' AND area_id = '%s')", school.SchoolID, area.AreaID)
			conditions = append(conditions, condition)
		}
	}

	if len(conditions) == 0 {
		return []models.PostalCodeAssignment{}, nil
	}

	query = query.Where(strings.Join(conditions, " OR "))

	// 应用搜索过滤器
	if code, ok := filters["code"].(string); ok && code != "" {
		query = query.Where("postal_code LIKE ?", "%"+code+"%")
	}

	if userID, ok := filters["user_id"].(string); ok && userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	if schoolID, ok := filters["school_id"].(string); ok && schoolID != "" {
		query = query.Where("school_id = ?", schoolID)
	}

	if areaID, ok := filters["area_id"].(string); ok && areaID != "" {
		query = query.Where("area_id = ?", areaID)
	}

	if isActive, ok := filters["is_active"].(*bool); ok && isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	limit := filters["limit"].(int)
	var results []models.PostalCodeAssignment
	err = query.Limit(limit).Order("assigned_at DESC").Find(&results).Error
	if err != nil {
		return nil, fmt.Errorf("搜索编号失败: %w", err)
	}

	return results, nil
}

// ValidateCodeRange 验证编号范围
func (s *PostalManagementService) ValidateCodeRange(schoolID, areaID string, codeRange []string) (map[string]interface{}, error) {
	var validCodes []string
	var invalidCodes []string
	var conflicts []map[string]interface{}

	for _, code := range codeRange {
		// 检查编号格式
		if !s.isValidCodeFormat(code, schoolID, areaID) {
			invalidCodes = append(invalidCodes, code)
			continue
		}

		// 检查是否已被使用
		var assignment models.PostalCodeAssignment
		if err := s.db.Where("postal_code = ? AND is_active = ?", code, true).First(&assignment).Error; err == nil {
			conflicts = append(conflicts, map[string]interface{}{
				"code":        code,
				"used_by":     assignment.UserID,
				"assigned_at": assignment.AssignedAt,
			})
			continue
		}

		validCodes = append(validCodes, code)
	}

	return map[string]interface{}{
		"valid_codes":    validCodes,
		"invalid_codes":  invalidCodes,
		"conflicts":      conflicts,
		"total_checked":  len(codeRange),
		"valid_count":    len(validCodes),
		"invalid_count":  len(invalidCodes),
		"conflict_count": len(conflicts),
	}, nil
}

// isValidCodeFormat 检查编号格式是否有效
func (s *PostalManagementService) isValidCodeFormat(code, schoolID, _ string) bool {
	// 获取学校规则
	var rule models.PostalCodeRule
	if err := s.db.Where("school_id = ?", schoolID).First(&rule).Error; err != nil {
		return false
	}

	// 检查前缀
	if !strings.HasPrefix(code, rule.Prefix) {
		return false
	}

	// 检查长度和格式
	if len(code) < len(rule.Prefix)+3 {
		return false
	}

	// 检查数字部分
	numPart := code[len(rule.Prefix):]
	if _, err := strconv.Atoi(numPart); err != nil {
		return false
	}

	return true
}

// GetApplicationHistory 获取申请历史
func (s *PostalManagementService) GetApplicationHistory(courierID string, filters map[string]interface{}) ([]models.PostalCodeApplication, int64, error) {
	scope, err := s.GetPermissionScope(courierID)
	if err != nil {
		return nil, 0, err
	}

	query := s.db.Model(&models.PostalCodeApplication{})

	// 应用权限过滤
	var conditions []string
	for _, school := range scope.Schools {
		for _, area := range school.Areas {
			condition := fmt.Sprintf("(school_id = '%s' AND area_id = '%s')", school.SchoolID, area.AreaID)
			conditions = append(conditions, condition)
		}
	}

	if len(conditions) == 0 {
		return []models.PostalCodeApplication{}, 0, nil
	}

	query = query.Where(strings.Join(conditions, " OR "))

	// 应用过滤器
	if userID, ok := filters["user_id"].(string); ok && userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	if schoolID, ok := filters["school_id"].(string); ok && schoolID != "" {
		query = query.Where("school_id = ?", schoolID)
	}

	if status, ok := filters["status"].(string); ok && status != "" {
		query = query.Where("status = ?", status)
	}

	if startDate, ok := filters["start_date"].(string); ok && startDate != "" {
		if date, err := time.Parse("2006-01-02", startDate); err == nil {
			query = query.Where("created_at >= ?", date)
		}
	}

	if endDate, ok := filters["end_date"].(string); ok && endDate != "" {
		if date, err := time.Parse("2006-01-02", endDate); err == nil {
			query = query.Where("created_at <= ?", date.Add(24*time.Hour))
		}
	}

	// 获取总数
	var total int64
	query.Count(&total)

	// 分页查询
	limit := filters["limit"].(int)
	offset := filters["offset"].(int)

	var applications []models.PostalCodeApplication
	err = query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&applications).Error
	if err != nil {
		return nil, 0, fmt.Errorf("查询申请历史失败: %w", err)
	}

	return applications, total, nil
}

// 通知相关方法

func (s *PostalManagementService) notifyApplicationReviewed(application *models.PostalCodeApplication, action, reviewerID string) {
	event := utils.WebSocketEvent{
		Type: "POSTAL_APPLICATION_REVIEWED",
		Data: map[string]interface{}{
			"application_id": application.ID,
			"user_id":        application.UserID,
			"action":         action,
			"status":         application.Status,
			"assigned_code":  application.AssignedCode,
			"reviewed_by":    reviewerID,
			"reviewed_at":    application.ReviewedAt,
		},
		Timestamp: time.Now(),
	}
	s.wsManager.BroadcastToUser(application.UserID, event)
}

func (s *PostalManagementService) notifyCodeAssigned(assignment *models.PostalCodeAssignment) {
	event := utils.WebSocketEvent{
		Type: "POSTAL_CODE_ASSIGNED",
		Data: map[string]interface{}{
			"user_id":     assignment.UserID,
			"postal_code": assignment.PostalCode,
			"school_id":   assignment.SchoolID,
			"area_id":     assignment.AreaID,
			"assigned_by": assignment.AssignedBy,
			"assigned_at": assignment.AssignedAt,
		},
		Timestamp: time.Now(),
	}
	s.wsManager.BroadcastToUser(assignment.UserID, event)
}

func (s *PostalManagementService) notifyCodeDeactivated(assignment *models.PostalCodeAssignment, operatorID, reason string) {
	event := utils.WebSocketEvent{
		Type: "POSTAL_CODE_DEACTIVATED",
		Data: map[string]interface{}{
			"user_id":        assignment.UserID,
			"postal_code":    assignment.PostalCode,
			"deactivated_by": operatorID,
			"deactivated_at": assignment.DeactivatedAt,
			"reason":         reason,
		},
		Timestamp: time.Now(),
	}
	s.wsManager.BroadcastToUser(assignment.UserID, event)
}

func (s *PostalManagementService) notifyBatchAssignment(assignerID, schoolID, areaID string, successCount, failureCount int) {
	event := utils.WebSocketEvent{
		Type: "POSTAL_BATCH_ASSIGNMENT",
		Data: map[string]interface{}{
			"assigned_by":   assignerID,
			"school_id":     schoolID,
			"area_id":       areaID,
			"success_count": successCount,
			"failure_count": failureCount,
		},
		Timestamp: time.Now(),
	}
	s.wsManager.BroadcastToAdmins(event)
}

// 辅助函数
func getActionName(action string) string {
	switch action {
	case "approve":
		return "批准"
	case "reject":
		return "拒绝"
	default:
		return "处理"
	}
}
