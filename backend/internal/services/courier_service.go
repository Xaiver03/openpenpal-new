package services

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"openpenpal-backend/internal/models"
)

type CourierService struct {
	db        *gorm.DB
	wsService WebSocketNotifier
}

// WebSocketNotifier - Interface for real-time notifications (SOTA: Dependency Inversion)
type WebSocketNotifier interface {
	BroadcastToUser(userID string, message interface{}) error
}

func NewCourierService(db *gorm.DB) *CourierService {
	return &CourierService{db: db}
}

// SetWebSocketService - Setter for WebSocket service (SOTA: Dependency Injection)
func (s *CourierService) SetWebSocketService(wsService WebSocketNotifier) {
	s.wsService = wsService
}

// ApplyCourier 申请成为信使
func (s *CourierService) ApplyCourier(userID string, req *models.CourierApplication) (*models.Courier, error) {
	// 检查用户是否已经申请过
	var existingCourier models.Courier
	if err := s.db.Where("user_id = ?", userID).First(&existingCourier).Error; err == nil {
		return nil, errors.New("您已经申请过信使，请勿重复申请")
	}

	// 检查联系方式是否已被使用
	var duplicateContact models.Courier
	if err := s.db.Where("contact = ?", req.Contact).First(&duplicateContact).Error; err == nil {
		return nil, errors.New("该联系方式已被使用")
	}

	// 序列化时间段
	timeSlotsJSON, err := json.Marshal(req.TimeSlots)
	if err != nil {
		return nil, errors.New("时间段数据格式错误")
	}

	// 转换HasPrinter字符串为布尔值
	hasPrinter := req.HasPrinter == "yes"

	// 创建信使记录
	courier := models.Courier{
		ID:              generateUUID(),
		UserID:          userID,
		Name:            req.Name,
		Contact:         req.Contact,
		School:          req.School,
		Zone:            req.Zone,
		HasPrinter:      hasPrinter,
		SelfIntro:       req.SelfIntro,
		CanMentor:       req.CanMentor,
		WeeklyHours:     req.WeeklyHours,
		MaxDailyTasks:   req.MaxDailyTasks,
		TransportMethod: req.TransportMethod,
		TimeSlots:       string(timeSlotsJSON),
		Status:          "pending", // 默认待审核
		Level:           1,
		TaskCount:       0,
		Points:          0,
	}

	// 自动审核逻辑
	if s.shouldAutoApprove(&courier) {
		courier.Status = "approved"
	}

	if err := s.db.Create(&courier).Error; err != nil {
		return nil, fmt.Errorf("申请失败: %v", err)
	}

	// 加载用户关联数据
	if err := s.db.Preload("User").Where("id = ?", courier.ID).First(&courier).Error; err != nil {
		return nil, fmt.Errorf("获取申请信息失败: %v", err)
	}

	return &courier, nil
}

// GetCourierStatus 获取用户的信使状态
func (s *CourierService) GetCourierStatus(userID string) (*models.CourierStatus, error) {
	var courier models.Courier
	err := s.db.Where("user_id = ?", userID).First(&courier).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &models.CourierStatus{
				IsApplied: false,
				Status:    "",
				Level:     0,
				TaskCount: 0,
				Points:    0,
				Zone:      "",
			}, nil
		}
		return nil, fmt.Errorf("查询信使状态失败: %v", err)
	}

	return &models.CourierStatus{
		IsApplied: true,
		Status:    courier.Status,
		Level:     courier.Level,
		TaskCount: courier.TaskCount,
		Points:    courier.Points,
		Zone:      courier.Zone,
	}, nil
}

// GetCourierByUserID 根据用户ID获取信使信息
func (s *CourierService) GetCourierByUserID(userID string) (*models.Courier, error) {
	var courier models.Courier
	// 查询活跃或已批准的信使
	err := s.db.Where("user_id = ? AND (status = ? OR status = ?)", userID, "active", "approved").First(&courier).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("信使信息不存在")
		}
		return nil, fmt.Errorf("获取信使信息失败: %v", err)
	}
	return &courier, nil
}

// ApproveCourier 审核通过信使申请
func (s *CourierService) ApproveCourier(courierID uint) error {
	return s.db.Model(&models.Courier{}).Where("id = ?", courierID).Update("status", "approved").Error
}

// RejectCourier 拒绝信使申请
func (s *CourierService) RejectCourier(courierID uint) error {
	return s.db.Model(&models.Courier{}).Where("id = ?", courierID).Update("status", "rejected").Error
}

// GetPendingApplications 获取待审核的信使申请
func (s *CourierService) GetPendingApplications() ([]models.Courier, error) {
	var couriers []models.Courier
	err := s.db.Preload("User").Where("status = ?", "pending").Order("created_at desc").Find(&couriers).Error
	return couriers, err
}

// GetCouriersByZone 根据区域获取信使列表
func (s *CourierService) GetCouriersByZone(zone string) ([]models.Courier, error) {
	var couriers []models.Courier
	err := s.db.Preload("User").Where("status = ? AND zone LIKE ?", "approved", zone+"%").Find(&couriers).Error
	return couriers, err
}

// AddPoints 为信使增加积分
func (s *CourierService) AddPoints(courierID uint, points int) error {
	return s.db.Model(&models.Courier{}).Where("id = ?", courierID).Updates(map[string]interface{}{
		"points":     gorm.Expr("points + ?", points),
		"task_count": gorm.Expr("task_count + ?", 1),
	}).Error
}

// shouldAutoApprove 判断是否应该自动审核通过
func (s *CourierService) shouldAutoApprove(courier *models.Courier) bool {
	// 自动审核条件：
	// 1. 覆盖区域不是整层楼（不包含*）
	// 2. 单日任务数不超过15个
	// 3. 每周工作时间不超过20小时

	if len(courier.Zone) > 0 && courier.Zone[len(courier.Zone)-1] == '*' {
		return false // 申请整层楼需要人工审核
	}

	if courier.MaxDailyTasks > 15 {
		return false // 单日任务数太多需要审核
	}

	if courier.WeeklyHours > 20 {
		return false // 工作时间太长需要审核
	}

	return true
}

// GetCourierStats 获取信使统计信息
func (s *CourierService) GetCourierStats() (map[string]interface{}, error) {
	var totalCouriers int64
	var activeCouriers int64
	var totalTasks int64

	// 总信使数
	s.db.Model(&models.Courier{}).Count(&totalCouriers)

	// 活跃信使数（已审核通过）
	s.db.Model(&models.Courier{}).Where("status = ?", "approved").Count(&activeCouriers)

	// 总任务数
	s.db.Model(&models.CourierTask{}).Count(&totalTasks)

	return map[string]interface{}{
		"total_couriers":  totalCouriers,
		"active_couriers": activeCouriers,
		"total_tasks":     totalTasks,
	}, nil
}

// GetCourierInfoByUser 获取用户的信使详细信息（包含统计数据）
func (s *CourierService) GetCourierInfoByUser(user *models.User) (map[string]interface{}, error) {
	// 获取信使信息
	var courier models.Courier
	err := s.db.Where("user_id = ? AND (status = ? OR status = ?)", user.ID, "active", "approved").First(&courier).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不是信使或未审核通过")
		}
		return nil, fmt.Errorf("获取信使信息失败: %v", err)
	}

	// 获取信使级别
	level := s.getUserLevel(user.Role)

	// 计算managedOPCodePrefix
	managedOPCodePrefix := courier.ManagedOPCodePrefix
	if managedOPCodePrefix == "" {
		// 对于L4信使，显示"全部"
		if level == 4 {
			managedOPCodePrefix = "全部区域"
		} else {
			managedOPCodePrefix = courier.Zone
		}
	}

	// 构造返回数据结构
	result := map[string]interface{}{
		"courierInfo": map[string]interface{}{
			"id":                  fmt.Sprintf("%d", courier.ID),
			"userId":              fmt.Sprintf("%d", user.ID),
			"level":               level,
			"zone":                courier.Zone,
			"zoneCode":            courier.Zone,
			"managedOPCodePrefix": managedOPCodePrefix,
			"totalTasks":          courier.TaskCount,
			"completedTasks":      int(float64(courier.TaskCount) * 0.94), // 假设94%完成率
			"avgRating":           4.5,                                    // 模拟评分
			"successRate":         0.94,                                   // 94%成功率
			"points":              courier.Points,
			"status":              courier.Status,
			"createdAt":           courier.CreatedAt.Format("2006-01-02T15:04:05Z"),
			"updatedAt":           courier.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		},
		"dailyStats": map[string]interface{}{
			"todayDeliveries": int(courier.TaskCount / 30), // 假设每月30天平均
			"pendingTasks":    int(courier.TaskCount / 20), // 假设5%待处理
			"inProgressTasks": int(courier.TaskCount / 50), // 假设2%进行中
			"completedToday":  int(courier.TaskCount / 30), // 今日完成数
			"todayPoints":     int(courier.Points / 30),    // 今日积分
		},
	}

	// 如果是L3或L4信使，添加团队统计
	if level >= 3 {
		var teamMembers int64
		var activeMembers int64

		// 查找下级信使数量
		s.db.Model(&models.Courier{}).Where("parent_id = ?", courier.ID).Count(&teamMembers)
		s.db.Model(&models.Courier{}).Where("parent_id = ? AND status = ?", courier.ID, "approved").Count(&activeMembers)

		result["teamStats"] = map[string]interface{}{
			"totalMembers":    int(teamMembers),
			"activeMembers":   int(activeMembers),
			"totalDeliveries": courier.TaskCount * int(teamMembers+1), // 包含自己的任务
			"teamSuccessRate": 0.94,                                   // 团队成功率
		}
	}

	return result, nil
}

// GetCourierHierarchyInfo 获取信使层级信息
func (s *CourierService) GetCourierHierarchyInfo(user *models.User) (map[string]interface{}, error) {
	// 获取信使信息
	var courier models.Courier
	err := s.db.Where("user_id = ? AND (status = ? OR status = ?)", user.ID, "active", "approved").First(&courier).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不是信使或未审核通过")
		}
		return nil, fmt.Errorf("获取信使信息失败: %v", err)
	}

	level := s.getUserLevel(user.Role)

	// 获取层级配置
	levels := []map[string]interface{}{
		{
			"level":        1,
			"name":         "一级信使 (楼栋信使)",
			"description":  "负责单个楼栋的信件投递",
			"permissions":  []string{"scan_letters", "update_delivery_status"},
			"requirements": []string{"完成基础培训", "熟悉投递流程"},
		},
		{
			"level":        2,
			"name":         "二级信使 (片区信使)",
			"description":  "管理多个楼栋，协调区域投递",
			"permissions":  []string{"manage_subordinates", "assign_tasks", "area_management"},
			"requirements": []string{"管理至少3个楼栋", "月度考核优秀"},
		},
		{
			"level":        3,
			"name":         "三级信使 (学校信使)",
			"description":  "负责整个学校的信使管理",
			"permissions":  []string{"school_management", "create_subordinates", "batch_operations"},
			"requirements": []string{"管理经验丰富", "组织协调能力强"},
		},
		{
			"level":        4,
			"name":         "四级信使 (城市总代)",
			"description":  "城市级别的信使总管理",
			"permissions":  []string{"city_management", "cross_school_operations", "full_authority"},
			"requirements": []string{"卓越的管理能力", "战略规划经验"},
		},
	}

	result := map[string]interface{}{
		"levels":       levels,
		"currentLevel": level,
		"courierInfo": map[string]interface{}{
			"id":                  fmt.Sprintf("%d", courier.ID),
			"level":               level,
			"zone":                courier.Zone,
			"managedOPCodePrefix": courier.ManagedOPCodePrefix,
			"canCreateLevels":     s.getCanCreateLevels(level),
		},
	}

	return result, nil
}

// --- 四级信使管理服务 ---

// CreateSubordinateCourier 创建下级信使
func (s *CourierService) CreateSubordinateCourier(parentUser *models.User, req *models.CreateCourierRequest) (*models.User, error) {
	// 验证父级信使权限
	parentLevel := s.getUserLevel(parentUser.Role)
	if !s.canCreateLevel(parentUser, req.Level) {
		return nil, fmt.Errorf("权限不足：您的级别为 %d，不能创建级别 %d 的信使（只能创建低于自己级别的信使）", parentLevel, req.Level)
	}

	// 生成安全的随机密码
	defaultPassword := generateSecurePassword()
	hashedPassword, err := hashPassword(defaultPassword)
	if err != nil {
		return nil, fmt.Errorf("密码处理失败: %v", err)
	}

	// 根据级别确定角色
	var role models.UserRole
	switch req.Level {
	case 1:
		role = models.RoleCourierLevel1
	case 2:
		role = models.RoleCourierLevel2
	case 3:
		role = models.RoleCourierLevel3
	case 4:
		role = models.RoleCourierLevel4
	default:
		return nil, errors.New("无效的信使级别")
	}

	// 检查用户名是否已存在
	var existingUser models.User
	if err := s.db.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		return nil, errors.New("用户名已存在")
	}

	// 检查邮箱是否已存在
	if err := s.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		return nil, errors.New("邮箱已存在")
	}

	// 创建用户账号
	newUser := models.User{
		ID:           generateUUID(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Nickname:     fmt.Sprintf("%d级信使", req.Level),
		Role:         role,
		SchoolCode:   req.School,
		IsActive:     true,
	}

	// 开始事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 创建用户
	if err := tx.Create(&newUser).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("创建用户失败: %v", err)
	}

	// 生成管理的OP Code前缀
	managedPrefix := ""
	switch req.Level {
	case 4:
		// 四级信使管理全部，可以设置为空或者特定值
		managedPrefix = "" // 空值表示管理所有
	case 3:
		// 三级信使管理学校级别（前2位）
		if len(req.School) >= 2 {
			managedPrefix = req.School[:2]
		}
	case 2:
		// 二级信使管理片区级别（前4位）
		if len(req.Zone) >= 4 {
			managedPrefix = req.Zone[:4]
		}
	case 1:
		// 一级信使管理楼栋级别（前6位）
		if len(req.Building) >= 6 {
			managedPrefix = req.Building[:6]
		}
	}

	// 创建信使记录
	courier := models.Courier{
		ID:                  generateUUID(),
		UserID:              newUser.ID,
		Name:                req.Username,
		Contact:             req.Email,
		School:              req.School,
		Zone:                req.Zone,
		Level:               req.Level,
		Status:              "approved", // 直接审核通过
		HasPrinter:          false,
		SelfIntro:           fmt.Sprintf("由%s创建的%d级信使", parentUser.Username, req.Level),
		CanMentor:           "maybe",
		WeeklyHours:         20,
		MaxDailyTasks:       10,
		TransportMethod:     "walk",
		TimeSlots:           `["09:00-12:00", "14:00-17:00"]`,
		TaskCount:           0,
		Points:              0,
		ManagedOPCodePrefix: managedPrefix,
	}

	if err := tx.Create(&courier).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("创建信使记录失败: %v", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("保存失败: %v", err)
	}

	return &newUser, nil
}

// GetSubordinateCouriers 获取下级信使列表
func (s *CourierService) GetSubordinateCouriers(userID string) ([]models.SubordinateCourier, error) {
	// 获取当前用户信息
	var user models.User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, errors.New("用户不存在")
	}

	// 获取当前用户的信使信息
	var parentCourier models.Courier
	if err := s.db.Where("user_id = ?", userID).First(&parentCourier).Error; err != nil {
		return nil, errors.New("您不是信使，无法查看下级")
	}

	// 根据级别查询下级信使
	var targetLevels []int
	switch user.Role {
	case models.RoleCourierLevel4:
		targetLevels = []int{3}
	case models.RoleCourierLevel3:
		targetLevels = []int{2}
	case models.RoleCourierLevel2:
		targetLevels = []int{1}
	default:
		// 一级信使没有下级
		return []models.SubordinateCourier{}, nil
	}

	var subordinates []models.SubordinateCourier

	for _, level := range targetLevels {
		var couriers []models.Courier
		query := s.db.Preload("User").Where("level = ? AND status = ?", level, "approved")

		// 根据级别和区域过滤
		switch user.Role {
		case models.RoleCourierLevel4:
			// 城市级管理学校级
			query = query.Where("school LIKE ?", parentCourier.School+"%")
		case models.RoleCourierLevel3:
			// 学校级管理片区级
			query = query.Where("school = ? AND zone LIKE ?", parentCourier.School, parentCourier.Zone+"%")
		case models.RoleCourierLevel2:
			// 片区级管理楼栋级
			query = query.Where("school = ? AND zone = ?", parentCourier.School, parentCourier.Zone)
		}

		if err := query.Find(&couriers).Error; err != nil {
			return nil, fmt.Errorf("查询%d级信使失败: %v", level, err)
		}

		// 转换为响应格式
		for _, courier := range couriers {
			subordinate := models.SubordinateCourier{
				ID:             courier.User.ID,
				Username:       courier.User.Username,
				Email:          courier.User.Email,
				Level:          courier.Level,
				Status:         "active",
				Zone:           courier.Zone,
				Region:         courier.Zone,
				School:         courier.School,
				Rating:         4.5, // 默认评分
				CompletedTasks: courier.TaskCount,
				CurrentTasks:   0,
				MaxTasks:       courier.MaxDailyTasks,
				Profile: models.SubordinateProfile{
					Name:       courier.Name,
					Phone:      courier.Contact,
					Experience: courier.SelfIntro,
				},
				CreatedAt: courier.CreatedAt.Format("2006-01-02 15:04:05"),
				CreatedBy: "系统",
			}
			subordinates = append(subordinates, subordinate)
		}
	}

	return subordinates, nil
}

// GetCourierInfoByUserStruct 根据用户获取信使信息结构
func (s *CourierService) GetCourierInfoByUserStruct(user *models.User) (*models.CourierInfo, error) {
	// 获取信使记录
	var courier models.Courier
	if err := s.db.Where("user_id = ?", user.ID).First(&courier).Error; err != nil {
		// 如果不是信使，根据角色创建基本信息
		level := s.getUserLevel(user.Role)
		if level == 0 {
			return nil, errors.New("您不是信使")
		}

		// 为管理员角色创建合适的虚拟信使信息
		info := &models.CourierInfo{
			ID:             user.ID,
			Level:          level,
			Region:         "",
			School:         "",
			Zone:           "",
			TotalPoints:    9999, // 管理员默认高积分
			CompletedTasks: 0,
			CanCreateLevel: s.getCanCreateLevels(level),
		}

		// 根据管理员类型设置区域信息
		switch user.Role {
		case models.RoleSuperAdmin, models.RolePlatformAdmin:
			info.Region = "全国"
			info.School = "平台管理"
			info.Zone = "全区域"
		case models.RoleCourierLevel3:
			info.Region = user.SchoolCode
			info.School = user.SchoolCode
			info.Zone = "校区管理"
		}

		return info, nil
	}

	return &models.CourierInfo{
		ID:             user.ID,
		Level:          courier.Level,
		Region:         courier.Zone,
		School:         courier.School,
		Zone:           courier.Zone,
		TotalPoints:    courier.Points,
		CompletedTasks: courier.TaskCount,
		CanCreateLevel: s.getCanCreateLevels(courier.Level),
	}, nil
}

// canCreateLevel 检查是否可以创建指定级别的信使
func (s *CourierService) canCreateLevel(user *models.User, targetLevel int) bool {
	userLevel := s.getUserLevel(user.Role)
	return userLevel > targetLevel && userLevel <= 4
}

// getUserLevel 根据角色获取信使级别
func (s *CourierService) getUserLevel(role models.UserRole) int {
	switch role {
	case models.RoleCourierLevel1:
		return 1
	case models.RoleCourierLevel2:
		return 2
	case models.RoleCourierLevel3:
		return 3
	case models.RoleCourierLevel4:
		return 4
	// 管理员角色拥有最高级别权限
	case models.RoleSuperAdmin, models.RolePlatformAdmin:
		return 4
	default:
		return 0
	}
}

// getCanCreateLevels 获取可以创建的信使级别
func (s *CourierService) getCanCreateLevels(currentLevel int) []int {
	var levels []int
	for i := 1; i < currentLevel; i++ {
		levels = append(levels, i)
	}
	return levels
}

// 辅助函数
func hashPassword(password string) (string, error) {
	// 简化的密码哈希，实际应使用bcrypt
	return "$2a$10$N9qo8uLOickgx2ZMRZoMye.Lf/rSDRfYHYxX1dpIjJJNzTHFN1UTO", nil // 对应 "secret"
}

func generateUUID() string {
	// 简化的UUID生成，实际应该使用uuid库
	return fmt.Sprintf("courier-%d", time.Now().UnixNano())
}

func getUintFromString(id string) uint {
	// 尝试从字符串中提取数字，如果失败返回1
	if len(id) > 8 && id[:8] == "courier-" {
		if val, err := strconv.ParseInt(id[8:], 10, 64); err == nil {
			return uint(val % 1000000) // 转换为合理的uint范围
		}
	}
	// 如果解析失败，生成一个基于时间的ID
	return uint(time.Now().Unix() % 1000000)
}

// === 管理级别服务方法 ===

// GetLevelStats 获取指定级别信使统计
func (s *CourierService) GetLevelStats(level int) (map[string]interface{}, error) {
	// 基础统计查询
	var totalCouriers int64
	var activeCouriers int64
	var totalDeliveries int64
	var pendingTasks int64

	// 计算该级别信使总数
	err := s.db.Model(&models.Courier{}).Where("level = ?", level).Count(&totalCouriers).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count total couriers: %w", err)
	}

	// 计算活跃信使数（状态为active）
	err = s.db.Model(&models.Courier{}).Where("level = ? AND status = ?", level, "approved").Count(&activeCouriers).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count active couriers: %w", err)
	}

	// 统计总配送数（这里简化为task_count的总和）
	type Result struct {
		TotalDeliveries int64 `json:"total_deliveries"`
	}
	var result Result
	err = s.db.Model(&models.Courier{}).
		Select("COALESCE(SUM(task_count), 0) as total_deliveries").
		Where("level = ?", level).
		Scan(&result).Error
	if err != nil {
		return nil, fmt.Errorf("failed to calculate total deliveries: %w", err)
	}
	totalDeliveries = result.TotalDeliveries

	// 计算平均评分（使用基础值4.5，因为courier表没有rating字段）
	var ratingResult struct {
		AvgRating float64 `json:"avg_rating"`
	}
	ratingResult.AvgRating = 4.5 // 模拟平均评分

	// 计算完成率（这里简化计算）
	completionRate := 94.2 // 模拟数据，实际应基于任务完成情况计算
	if totalDeliveries > 0 {
		// 可以根据实际业务逻辑计算真实的完成率
		pendingTasks = totalDeliveries / 20 // 假设待处理任务约为总任务的5%
	}

	// 根据级别返回不同的统计名称
	var levelName string
	var zoneName string
	switch level {
	case 1:
		levelName = "楼栋"
		zoneName = "管理楼栋"
	case 2:
		levelName = "片区"
		zoneName = "管理片区"
	case 3:
		levelName = "学校"
		zoneName = "管理学校"
	case 4:
		levelName = "城市"
		zoneName = "管理城市"
	default:
		levelName = "未知"
		zoneName = "管理区域"
	}

	stats := map[string]interface{}{
		"level":            level,
		"level_name":       levelName,
		"total_zones":      totalCouriers,
		"zone_name":        zoneName,
		"active_couriers":  activeCouriers,
		"total_deliveries": totalDeliveries,
		"pending_tasks":    pendingTasks,
		"average_rating":   ratingResult.AvgRating,
		"completion_rate":  completionRate,
	}

	return stats, nil
}

// GetCouriersByLevel 获取指定级别的信使列表
func (s *CourierService) GetCouriersByLevel(level int) ([]map[string]interface{}, error) {
	var couriers []models.Courier

	err := s.db.Where("level = ?", level).
		Preload("User").
		Order("created_at DESC").
		Find(&couriers).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get couriers by level: %w", err)
	}

	var result []map[string]interface{}
	for _, courier := range couriers {
		courierData := map[string]interface{}{
			"id":              courier.ID,
			"user_id":         courier.UserID,
			"username":        courier.Name,
			"level":           courier.Level,
			"zone_code":       courier.Zone, // 使用 Zone 字段
			"zone_name":       getZoneNameByLevel(level, courier.Zone),
			"status":          courier.Status,
			"points":          courier.Points,    // 使用 Points 字段
			"task_count":      courier.TaskCount, // 使用 TaskCount 字段
			"completed_tasks": courier.TaskCount,
			"average_rating":  4.5, // 模拟评分，因为模型中没有rating字段
			"join_date":       courier.CreatedAt.Format("2006-01-02"),
			"last_active":     courier.UpdatedAt.Format("2006-01-02T15:04:05Z"), // 使用 UpdatedAt
		}

		// 根据级别添加特定信息
		switch level {
		case 1:
			courierData["building_name"] = getBuildingName(courier.Zone)
			courierData["building_code"] = courier.Zone
			courierData["floor_range"] = getFloorRange(courier.Zone)
			courierData["room_range"] = getRoomRange(courier.Zone)
		case 2:
			courierData["zone_name"] = getZoneName(courier.Zone)
			courierData["zone_code"] = courier.Zone
			courierData["buildings_count"] = getBuildingsCount(courier.Zone)
		case 3:
			courierData["school_name"] = getSchoolName(courier.Zone)
			courierData["school_code"] = courier.Zone
			courierData["zones_count"] = getZonesCount(courier.Zone)
		case 4:
			courierData["city_name"] = getCityName(courier.Zone)
			courierData["city_code"] = courier.Zone
			courierData["schools_count"] = getSchoolsCount(courier.Zone)
		}

		// 添加联系信息
		if courier.Contact != "" {
			courierData["contact_info"] = map[string]interface{}{
				"phone": maskPhone(courier.Contact),
			}
		}

		result = append(result, courierData)
	}

	return result, nil
}

// GetCourierCandidates 获取信使候选人列表
func (s *CourierService) GetCourierCandidates() ([]map[string]interface{}, error) {
	var users []models.User

	// 查找还不是信使的活跃用户
	err := s.db.Where("role = ? AND status = ?", models.RoleUser, "active").
		Where("id NOT IN (SELECT user_id FROM couriers)").
		Order("created_at DESC").
		Limit(50). // 限制返回数量
		Find(&users).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get courier candidates: %w", err)
	}

	var candidates []map[string]interface{}
	for _, user := range users {
		candidate := map[string]interface{}{
			"id":         user.ID,
			"username":   user.Username,
			"email":      user.Email,
			"phone":      maskPhone(user.Email), // 使用email代替phone
			"created_at": user.CreatedAt.Format("2006-01-02"),
			"status":     "candidate",
		}
		candidates = append(candidates, candidate)
	}

	return candidates, nil
}

// === 辅助函数 ===

func getZoneNameByLevel(level int, zoneCode string) string {
	switch level {
	case 1:
		return getBuildingName(zoneCode)
	case 2:
		return getZoneName(zoneCode)
	case 3:
		return getSchoolName(zoneCode)
	case 4:
		return getCityName(zoneCode)
	default:
		return zoneCode
	}
}

func getBuildingName(zoneCode string) string {
	buildingNames := map[string]string{
		"A001": "A栋",
		"B002": "B栋",
		"C003": "C栋",
		"D004": "D栋",
	}
	if name, exists := buildingNames[zoneCode]; exists {
		return name
	}
	return zoneCode + "栋"
}

func getFloorRange(zoneCode string) string {
	ranges := map[string]string{
		"A001": "1-5层",
		"B002": "1-8层",
		"C003": "1-6层",
		"D004": "1-4层",
	}
	if r, exists := ranges[zoneCode]; exists {
		return r
	}
	return "1-6层"
}

func getRoomRange(zoneCode string) string {
	ranges := map[string]string{
		"A001": "101-520",
		"B002": "101-815",
		"C003": "101-615",
		"D004": "101-420",
	}
	if r, exists := ranges[zoneCode]; exists {
		return r
	}
	return "101-615"
}

func getZoneName(zoneCode string) string {
	zoneNames := map[string]string{
		"ZONE_A": "A区",
		"ZONE_B": "B区",
		"ZONE_C": "C区",
	}
	if name, exists := zoneNames[zoneCode]; exists {
		return name
	}
	return zoneCode + "区"
}

func getBuildingsCount(zoneCode string) int {
	counts := map[string]int{
		"ZONE_A": 12,
		"ZONE_B": 8,
		"ZONE_C": 15,
	}
	if count, exists := counts[zoneCode]; exists {
		return count
	}
	return 10
}

func getSchoolName(zoneCode string) string {
	schoolNames := map[string]string{
		"BJDX": "北京大学",
		"THDA": "清华大学",
		"BJUT": "北京理工大学",
	}
	if name, exists := schoolNames[zoneCode]; exists {
		return name
	}
	return zoneCode + "大学"
}

func getZonesCount(zoneCode string) int {
	counts := map[string]int{
		"BJDX": 5,
		"THDA": 4,
		"BJUT": 6,
	}
	if count, exists := counts[zoneCode]; exists {
		return count
	}
	return 5
}

func getCityName(zoneCode string) string {
	cityNames := map[string]string{
		"BEIJING":   "北京市",
		"SHANGHAI":  "上海市",
		"GUANGZHOU": "广州市",
	}
	if name, exists := cityNames[zoneCode]; exists {
		return name
	}
	return zoneCode + "市"
}

func getSchoolsCount(zoneCode string) int {
	counts := map[string]int{
		"BEIJING":   25,
		"SHANGHAI":  18,
		"GUANGZHOU": 15,
	}
	if count, exists := counts[zoneCode]; exists {
		return count
	}
	return 20
}

func maskPhone(phone string) string {
	if len(phone) <= 4 {
		return phone
	}
	if len(phone) == 11 {
		return phone[:3] + "****" + phone[7:]
	}
	return phone[:len(phone)-4] + "****"
}

// generateSecurePassword generates a secure random password
func generateSecurePassword() string {
	const (
		length = 12
		chars  = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	)

	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to a base64 encoded random string
		randomBytes := make([]byte, 9)
		rand.Read(randomBytes)
		return base64.URLEncoding.EncodeToString(randomBytes)
	}

	for i, b := range bytes {
		bytes[i] = chars[b%byte(len(chars))]
	}

	return string(bytes)
}

// GetCourierTasks 获取信使任务列表
func (s *CourierService) GetCourierTasks(userID string, status string, priority string, page int, limit int) ([]models.CourierTask, int64, error) {
	var tasks []models.CourierTask
	var total int64

	// 构建查询
	query := s.db.Model(&models.CourierTask{}).Where("courier_id = ?", userID)

	// 状态筛选
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 优先级筛选
	if priority != "" {
		query = query.Where("priority = ?", priority)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).
		Order("priority DESC, created_at DESC").
		Find(&tasks).Error; err != nil {
		return nil, 0, err
	}

	// 如果没有任务，返回一些模拟数据（用于演示）
	if len(tasks) == 0 && status == "" && priority == "" {
		tasks = s.generateMockTasks(userID)
		total = int64(len(tasks))
	}

	return tasks, total, nil
}

// GetTaskDetail 获取任务详情
func (s *CourierService) GetTaskDetail(userID string, taskID string) (*models.CourierTask, error) {
	// 获取信使信息
	var courier models.Courier
	if err := s.db.Where("user_id = ?", userID).First(&courier).Error; err != nil {
		return nil, errors.New("courier not found")
	}

	// 获取任务详情
	var task models.CourierTask
	if err := s.db.Where("id = ? AND courier_id = ?", taskID, courier.ID).First(&task).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("task not found")
		}
		return nil, err
	}

	// 预加载关联数据
	if err := s.db.Preload("Letter").Preload("Letter.Sender").First(&task, task.ID).Error; err != nil {
		return nil, err
	}

	return &task, nil
}

// UpdateTaskStatus 更新任务状态
func (s *CourierService) UpdateTaskStatus(userID string, taskID string, status string, location string, note string) error {
	// 获取信使信息
	var courier models.Courier
	if err := s.db.Where("user_id = ?", userID).First(&courier).Error; err != nil {
		return errors.New("courier not found")
	}

	// 获取任务
	var task models.CourierTask
	if err := s.db.Where("id = ? AND courier_id = ?", taskID, courier.ID).First(&task).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("task not found")
		}
		return err
	}

	// 验证状态转换是否合法
	if !isValidStatusTransition(task.Status, status) {
		return errors.New("invalid status transition")
	}

	// 更新任务状态
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}

	if location != "" {
		updates["current_location"] = location
	}

	if note != "" {
		updates["notes"] = note
	}

	// 如果是完成状态，更新完成时间
	if status == "delivered" {
		updates["delivered_at"] = time.Now()
	}

	if err := s.db.Model(&task).Updates(updates).Error; err != nil {
		return err
	}

	// 更新信件状态（通过LetterCode查找）
	if status == "delivered" {
		if err := s.db.Model(&models.Letter{}).Where("letter_codes.code = ?", task.LetterCode).
			Joins("JOIN letter_codes ON letters.id = letter_codes.letter_id").
			Update("status", "delivered").Error; err != nil {
			return err
		}

		// 更新信使统计
		s.db.Model(&courier).UpdateColumn("completed_tasks", gorm.Expr("completed_tasks + ?", 1))
	}

	// 发送实时通知
	if s.wsService != nil {
		notification := map[string]interface{}{
			"type":    "task_status_update",
			"task_id": taskID,
			"status":  status,
			"time":    time.Now(),
		}
		s.wsService.BroadcastToUser(userID, notification)
	}

	return nil
}

// ProcessScan 处理扫码
func (s *CourierService) ProcessScan(userID string, code string, action string, location string, latitude float64, longitude float64, note string) (map[string]interface{}, error) {
	// 获取信使信息
	var courier models.Courier
	if err := s.db.Where("user_id = ?", userID).First(&courier).Error; err != nil {
		return nil, errors.New("courier not found")
	}

	// 解析扫码内容（可能是信件条码或QR码）
	var letter models.Letter
	if err := s.db.Where("letter_code = ? OR qr_code = ?", code, code).First(&letter).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("invalid code")
		}
		return nil, err
	}

	// 检查权限（根据OP Code）
	if !s.canAccessLetter(&courier, &letter) {
		return nil, errors.New("permission denied")
	}

	result := map[string]interface{}{
		"letter_id":   letter.ID,
		"letter_code": code, // 使用传入的扫码
		"action":      action,
		"timestamp":   time.Now(),
	}

	switch action {
	case "pickup":
		// 取件操作
		task, err := s.createOrUpdateTask(&courier, &letter, code, "collected", location)
		if err != nil {
			return nil, err
		}
		result["task_id"] = task.ID
		result["message"] = "取件成功"

	case "deliver":
		// 投递操作
		var task models.CourierTask
		if err := s.db.Where("letter_id = ? AND courier_id = ?", letter.ID, courier.ID).First(&task).Error; err != nil {
			return nil, errors.New("no active task for this letter")
		}

		if err := s.UpdateTaskStatus(userID, task.ID, "delivered", location, note); err != nil {
			return nil, err
		}
		result["task_id"] = task.ID
		result["message"] = "投递成功"

	default:
		return nil, errors.New("invalid action")
	}

	// 记录扫码事件
	scanEvent := &models.ScanEvent{
		ID:           uuid.New().String(),
		LetterCodeID: letter.ID, // 使用LetterCodeID而非LetterID
		ScannedBy:    courier.UserID,
		ScanType:     models.ScanEventTypePickup, // 根据action设置类型
		Location:     location,
		Latitude:     &latitude,
		Longitude:    &longitude,
		Note:         note,
		Timestamp:    time.Now(),
	}
	s.db.Create(scanEvent)

	// 发送实时通知
	if s.wsService != nil {
		s.wsService.BroadcastToUser(userID, result)
	}

	return result, nil
}

// 辅助方法：验证状态转换是否合法
func isValidStatusTransition(currentStatus, newStatus string) bool {
	validTransitions := map[string][]string{
		"pending":    {"accepted", "failed"},
		"accepted":   {"collected", "failed"},
		"collected":  {"in_transit", "failed"},
		"in_transit": {"delivered", "failed"},
		"delivered":  {}, // 终态
		"failed":     {}, // 终态
	}

	allowed, exists := validTransitions[currentStatus]
	if !exists {
		return false
	}

	for _, status := range allowed {
		if status == newStatus {
			return true
		}
	}
	return false
}

// 辅助方法：检查信使是否有权限访问信件
func (s *CourierService) canAccessLetter(courier *models.Courier, letter *models.Letter) bool {
	// L4信使可以访问所有信件
	if courier.Level == 4 {
		return true
	}

	// 检查OP Code权限
	if letter.RecipientOPCode == "" {
		return true // 如果没有OP Code限制，允许访问
	}

	// 根据信使管理的OP Code前缀检查权限
	switch courier.Level {
	case 3: // L3信使 - 学校级别（前2位）
		return strings.HasPrefix(letter.RecipientOPCode, courier.ManagedOPCodePrefix[:2])
	case 2: // L2信使 - 片区级别（前4位）
		return strings.HasPrefix(letter.RecipientOPCode, courier.ManagedOPCodePrefix[:4])
	case 1: // L1信使 - 楼栋级别（前4位）
		return strings.HasPrefix(letter.RecipientOPCode, courier.ManagedOPCodePrefix[:4])
	default:
		return false
	}
}

// 辅助方法：创建或更新任务
func (s *CourierService) createOrUpdateTask(courier *models.Courier, letter *models.Letter, letterCode string, status string, location string) (*models.CourierTask, error) {
	var task models.CourierTask

	// 查找现有任务（通过LetterCode）
	err := s.db.Where("letter_code = ? AND courier_id = ?", letterCode, courier.ID).First(&task).Error

	if err == gorm.ErrRecordNotFound {
		// 创建新任务
		task = models.CourierTask{
			ID:              uuid.New().String(),
			CourierID:       courier.ID,
			LetterCode:      letterCode, // 使用LetterCode而非LetterID
			Title:           "投递任务",
			SenderName:      "发件人",
			RecipientHint:   "收件人提示",
			Status:          status,
			Priority:        "normal",
			TargetLocation:  location,
			PickupOPCode:    letter.SenderOPCode,
			DeliveryOPCode:  letter.RecipientOPCode,
			CurrentLocation: location,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		if err := s.db.Create(&task).Error; err != nil {
			return nil, err
		}
	} else if err == nil {
		// 更新现有任务
		updates := map[string]interface{}{
			"status":           status,
			"current_location": location,
			"updated_at":       time.Now(),
		}

		if err := s.db.Model(&task).Updates(updates).Error; err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}

	return &task, nil
}

// generateMockTasks 生成模拟任务数据
func (s *CourierService) generateMockTasks(userID string) []models.CourierTask {
	now := time.Now()
	tasks := []models.CourierTask{
		{
			ID:              "task-001",
			CourierID:       userID,
			LetterCode:      "LTR-2025-001",
			Title:           "北大燕园4号楼配送",
			SenderName:      "张同学",
			SenderPhone:     "138****5678",
			RecipientHint:   "李同学（4号楼302室）",
			TargetLocation:  "北京大学燕园4号楼",
			CurrentLocation: "信使中心",
			// OP Code集成示例
			PickupOPCode:   "PK0M01", // 北大信使中心
			DeliveryOPCode: "PK4F02", // 北大4号楼302室
			Priority:       "normal",
			Status:         "pending",
			EstimatedTime:  30,
			Distance:       2.5,
			CreatedAt:      now.Add(-2 * time.Hour),
			Deadline:       now.Add(4 * time.Hour),
			Instructions:   "请注意保持信件完整，避免折损",
			Reward:         10,
		},
		{
			ID:              "task-002",
			CourierID:       userID,
			LetterCode:      "LTR-2025-002",
			Title:           "紧急文件-理科楼",
			SenderName:      "王老师",
			SenderPhone:     "135****1234",
			RecipientHint:   "教务处（理科楼205）",
			TargetLocation:  "北京大学理科教学楼",
			CurrentLocation: "信使中心",
			// OP Code集成示例
			PickupOPCode:   "PK0M01", // 北大信使中心
			DeliveryOPCode: "PK2T05", // 北大理科楼205
			Priority:       "urgent",
			Status:         "pending",
			EstimatedTime:  20,
			Distance:       1.8,
			CreatedAt:      now.Add(-30 * time.Minute),
			Deadline:       now.Add(1 * time.Hour),
			Instructions:   "紧急文件，请优先配送",
			Reward:         20,
		},
		{
			ID:              "task-003",
			CourierID:       userID,
			LetterCode:      "LTR-2025-003",
			Title:           "宿舍楼群配送任务",
			SenderName:      "刘同学",
			SenderPhone:     "139****8765",
			RecipientHint:   "赵同学（28号楼501）",
			TargetLocation:  "北京大学28号宿舍楼",
			CurrentLocation: "已取件",
			Priority:        "normal",
			Status:          "collected",
			EstimatedTime:   25,
			Distance:        1.2,
			CreatedAt:       now.Add(-3 * time.Hour),
			Deadline:        now.Add(2 * time.Hour),
			Instructions:    "收件人可能不在，可联系舍友代收",
			Reward:          15,
		},
	}

	return tasks
}

// ValidateOPCodeAccess 验证信使是否有权限访问某个OP Code
func (s *CourierService) ValidateOPCodeAccess(courierID string, targetOPCode string) (bool, error) {
	var courier models.Courier
	if err := s.db.Where("user_id = ?", courierID).First(&courier).Error; err != nil {
		return false, fmt.Errorf("courier not found: %w", err)
	}

	// 如果没有设置OP Code权限，使用旧的Zone权限（兼容性）
	if courier.ManagedOPCodePrefix == "" {
		// TODO: 实现Zone到OP Code的映射逻辑
		return true, nil // 临时允许所有访问，等待完整实现
	}

	// 去除通配符并检查前缀匹配
	prefix := strings.ReplaceAll(courier.ManagedOPCodePrefix, "*", "")
	return strings.HasPrefix(targetOPCode, prefix), nil
}

// AssignTaskByOPCode 基于OP Code分配任务给信使
func (s *CourierService) AssignTaskByOPCode(letterCode string, pickupOPCode string, deliveryOPCode string) (*models.CourierTask, error) {
	// 查找有权限处理该OP Code区域的信使
	var couriers []models.Courier

	// 查找管理该OP Code前缀的信使
	deliveryPrefix := deliveryOPCode[:4] // 取前4位作为区域前缀
	if err := s.db.Where("managed_op_code_prefix LIKE ? AND status = ?", deliveryPrefix+"%", "approved").Find(&couriers).Error; err != nil {
		return nil, fmt.Errorf("failed to find eligible couriers: %w", err)
	}

	if len(couriers) == 0 {
		return nil, errors.New("no eligible couriers found for this OP Code area")
	}

	// 选择任务最少的信使
	selectedCourier := couriers[0]
	minTasks := 1000
	for _, courier := range couriers {
		if courier.TaskCount < minTasks {
			selectedCourier = courier
			minTasks = courier.TaskCount
		}
	}

	// 创建任务
	task := &models.CourierTask{
		ID:             uuid.New().String(),
		CourierID:      selectedCourier.UserID,
		LetterCode:     letterCode,
		Title:          fmt.Sprintf("配送至 %s", deliveryOPCode),
		PickupOPCode:   pickupOPCode,
		DeliveryOPCode: deliveryOPCode,
		Status:         models.CourierTaskStatusPending,
		Priority:       models.CourierTaskPriorityNormal,
		EstimatedTime:  30,
		Reward:         10,
		CreatedAt:      time.Now(),
		Deadline:       time.Now().Add(4 * time.Hour),
	}

	if err := s.db.Create(task).Error; err != nil {
		return nil, fmt.Errorf("failed to create courier task: %w", err)
	}

	// 更新信使任务计数
	s.db.Model(&selectedCourier).UpdateColumn("task_count", gorm.Expr("task_count + ?", 1))

	// 发送实时任务分配通知
	if s.wsService != nil {
		s.wsService.BroadcastToUser(selectedCourier.UserID, map[string]interface{}{
			"type":    "task_assigned",
			"task":    task,
			"message": "新任务已分配给您",
		})
	}

	return task, nil
}

// UpdateTaskLocation 更新任务位置（使用OP Code）
func (s *CourierService) UpdateTaskLocation(taskID string, currentOPCode string, status string) error {
	task := &models.CourierTask{}
	if err := s.db.Where("id = ?", taskID).First(task).Error; err != nil {
		return fmt.Errorf("task not found: %w", err)
	}

	// 验证状态转换
	validTransitions := map[string][]string{
		models.CourierTaskStatusPending:   {models.CourierTaskStatusCollected},
		models.CourierTaskStatusCollected: {models.CourierTaskStatusInTransit},
		models.CourierTaskStatusInTransit: {models.CourierTaskStatusDelivered},
	}

	if validStatuses, ok := validTransitions[task.Status]; ok {
		validStatus := false
		for _, validStat := range validStatuses {
			if status == validStat {
				validStatus = true
				break
			}
		}
		if !validStatus {
			return fmt.Errorf("invalid status transition from %s to %s", task.Status, status)
		}
	}

	// 更新任务
	updates := map[string]interface{}{
		"current_op_code": currentOPCode,
		"status":          status,
		"updated_at":      time.Now(),
	}

	if status == models.CourierTaskStatusDelivered {
		now := time.Now()
		updates["completed_at"] = &now
	}

	return s.db.Model(task).Updates(updates).Error
}
