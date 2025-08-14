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

// CourierGrowthService 信使成长路径与激励系统服务
type CourierGrowthService struct {
	db        *gorm.DB
	redis     *redis.Client
	wsManager *utils.WebSocketManager
}

// NewCourierGrowthService 创建成长路径服务
func NewCourierGrowthService(db *gorm.DB, redis *redis.Client, wsManager *utils.WebSocketManager) *CourierGrowthService {
	return &CourierGrowthService{
		db:        db,
		redis:     redis,
		wsManager: wsManager,
	}
}

// GetGrowthPath 获取成长路径配置
func (s *CourierGrowthService) GetGrowthPath(courierID string) (map[string]interface{}, error) {
	// 获取当前信使等级
	var courier models.Courier
	if err := s.db.Where("user_id = ?", courierID).First(&courier).Error; err != nil {
		return nil, fmt.Errorf("courier not found: %w", err)
	}

	currentLevel := models.CourierLevel(courier.Level)

	// 构建成长路径数据
	growthPath := map[string]interface{}{
		"courier_id":    courierID,
		"current_level": currentLevel,
		"current_name":  currentLevel.GetLevelName(),
		"paths":         make([]map[string]interface{}, 0),
	}

	// 添加所有可能的升级路径
	for level := models.LevelOne; level <= models.LevelFour; level++ {
		if level <= currentLevel {
			continue // 跳过已达到的等级
		}

		requirements := models.DefaultGrowthRequirements[level]
		pathInfo := map[string]interface{}{
			"target_level": level,
			"target_name":  level.GetLevelName(),
			"requirements": requirements,
			"zone_type":    models.DefaultZoneMapping[level],
			"permissions":  models.DefaultPermissionMatrix[level],
		}

		// 检查是否可以升级到这个等级
		if level == currentLevel+1 {
			progress, err := s.calculateProgressForLevel(courierID, level)
			if err == nil {
				pathInfo["can_upgrade"] = progress.CanUpgrade
				pathInfo["completion_rate"] = progress.CompletionRate
				pathInfo["detailed_requirements"] = progress.Requirements
			}
		}

		growthPath["paths"] = append(growthPath["paths"].([]map[string]interface{}), pathInfo)
	}

	return growthPath, nil
}

// GetGrowthProgress 获取成长进度
func (s *CourierGrowthService) GetGrowthProgress(courierID string) (*models.CourierGrowthProgress, error) {
	// 获取当前信使信息
	var courier models.Courier
	if err := s.db.Where("user_id = ?", courierID).First(&courier).Error; err != nil {
		return nil, fmt.Errorf("courier not found: %w", err)
	}

	currentLevel := models.CourierLevel(courier.Level)

	// 获取积分信息
	points, err := s.getOrCreatePoints(courierID)
	if err != nil {
		return nil, fmt.Errorf("failed to get points: %w", err)
	}

	// 获取徽章数量
	var badgeCount int64
	s.db.Model(&models.CourierBadgeEarned{}).Where("courier_id = ?", courierID).Count(&badgeCount)

	progress := &models.CourierGrowthProgress{
		CourierID:       courierID,
		CurrentLevel:    currentLevel,
		TotalPoints:     points.Total,
		AvailablePoints: points.Available,
		BadgesEarned:    int(badgeCount),
		LastUpdated:     time.Now(),
	}

	// 如果还不是最高等级，检查下一等级升级条件
	if currentLevel < models.LevelFour {
		nextLevel := currentLevel + 1
		progress.NextLevel = &nextLevel

		// 计算升级进度
		levelProgress, err := s.calculateProgressForLevel(courierID, nextLevel)
		if err == nil {
			progress.CanUpgrade = levelProgress.CanUpgrade
			progress.Requirements = levelProgress.Requirements
			progress.CompletionRate = levelProgress.CompletionRate
		}
	}

	return progress, nil
}

// CheckUpgradeRequirements 检查晋升条件
func (s *CourierGrowthService) CheckUpgradeRequirements(courierID string, targetLevel models.CourierLevel) (*models.CourierGrowthProgress, error) {
	// 获取当前等级
	var courier models.Courier
	if err := s.db.Where("user_id = ?", courierID).First(&courier).Error; err != nil {
		return nil, fmt.Errorf("courier not found: %w", err)
	}

	currentLevel := models.CourierLevel(courier.Level)
	if targetLevel != currentLevel+1 {
		return nil, errors.New("只能申请下一等级")
	}

	return s.calculateProgressForLevel(courierID, targetLevel)
}

// calculateProgressForLevel 计算特定等级的升级进度
func (s *CourierGrowthService) calculateProgressForLevel(courierID string, targetLevel models.CourierLevel) (*models.CourierGrowthProgress, error) {
	requirements := models.DefaultGrowthRequirements[targetLevel]
	if requirements == nil {
		return nil, errors.New("无效的目标等级")
	}

	var completedCount int
	var detailedRequirements []models.GrowthRequirement

	for _, req := range requirements {
		detailedReq := req
		completed, current := s.checkSingleRequirement(courierID, req)
		detailedReq.Completed = completed
		detailedReq.Current = current

		if completed {
			completedCount++
		}

		detailedRequirements = append(detailedRequirements, detailedReq)
	}

	completionRate := float64(completedCount) / float64(len(requirements)) * 100
	canUpgrade := completedCount == len(requirements)

	return &models.CourierGrowthProgress{
		CourierID:      courierID,
		CurrentLevel:   models.CourierLevel(0), // 将在调用处设置
		CanUpgrade:     canUpgrade,
		Requirements:   detailedRequirements,
		CompletionRate: completionRate,
		LastUpdated:    time.Now(),
	}, nil
}

// checkSingleRequirement 检查单个要求
func (s *CourierGrowthService) checkSingleRequirement(courierID string, req models.GrowthRequirement) (bool, interface{}) {
	switch req.Type {
	case "delivery_count":
		// 检查累计投递数量
		var count int64
		s.db.Model(&models.Task{}).
			Where("courier_id = ? AND status = ?", courierID, "delivered").
			Count(&count)
		target := int64(req.Target.(float64))
		return count >= target, count

	case "consecutive_days":
		// 检查连续投递天数 (简化实现)
		days := int(req.Target.(float64))
		since := time.Now().AddDate(0, 0, -days)
		var count int64
		s.db.Model(&models.Task{}).
			Where("courier_id = ? AND status = ? AND completed_at > ?", courierID, "delivered", since).
			Count(&count)
		return count >= int64(days), count

	case "manage_couriers":
		// 检查管理的信使数量
		var zoneCount int64
		s.db.Model(&models.CourierZone{}).
			Where("courier_id = ? AND is_active = ?", courierID, true).
			Count(&zoneCount)
		target := int64(req.Target.(float64))
		return zoneCount >= target, zoneCount

	case "completion_rate":
		// 检查完成率
		oneMonthAgo := time.Now().AddDate(0, -1, 0)
		var totalTasks, completedTasks int64

		s.db.Model(&models.Task{}).
			Where("courier_id = ? AND created_at > ?", courierID, oneMonthAgo).
			Count(&totalTasks)

		s.db.Model(&models.Task{}).
			Where("courier_id = ? AND status = ? AND created_at > ?", courierID, "delivered", oneMonthAgo).
			Count(&completedTasks)

		if totalTasks == 0 {
			return false, 0.0
		}

		rate := float64(completedTasks) / float64(totalTasks) * 100
		target := req.Target.(float64)
		return rate >= target, rate

	case "service_duration":
		// 检查服务时长
		var courier models.Courier
		if err := s.db.Where("user_id = ?", courierID).First(&courier).Error; err != nil {
			return false, 0
		}

		if courier.ApprovedAt == nil {
			return false, 0
		}

		months := int(time.Since(*courier.ApprovedAt).Hours() / 24 / 30)
		target := int(req.Target.(float64))
		return months >= target, months

	case "school_recommendation":
		// 检查校级推荐 (简化检查)
		var courier models.Courier
		if err := s.db.Where("user_id = ?", courierID).First(&courier).Error; err != nil {
			return false, false
		}
		return courier.Note == "校级推荐", courier.Note == "校级推荐"

	case "platform_approval":
		// 检查平台备案 (简化检查)
		// 这里可以检查特定的标记或记录
		return true, true // 简化实现

	default:
		return false, nil
	}
}

// GetAvailableIncentives 获取可领取的激励奖励
func (s *CourierGrowthService) GetAvailableIncentives(courierID string) ([]map[string]interface{}, error) {
	var incentives []map[string]interface{}

	// 检查各种激励类型
	for _, incentive := range models.DefaultIncentives {
		available, amount, reason := s.checkIncentiveAvailability(courierID, incentive)
		if available {
			incentives = append(incentives, map[string]interface{}{
				"type":        incentive.Type,
				"name":        incentive.Name,
				"description": incentive.Description,
				"amount":      amount,
				"reason":      reason,
			})
		}
	}

	// 检查自动徽章奖励
	badges := s.checkAutoAwardBadges(courierID)
	for _, badge := range badges {
		incentives = append(incentives, map[string]interface{}{
			"type":        "badge",
			"name":        badge.Name,
			"description": badge.Description,
			"badge_code":  badge.Code,
			"points":      badge.Points,
		})
	}

	return incentives, nil
}

// checkIncentiveAvailability 检查激励可用性
func (s *CourierGrowthService) checkIncentiveAvailability(courierID string, incentive models.CourierIncentive) (bool, float64, string) {
	switch incentive.Type {
	case models.IncentiveTypeSubsidy:
		// 检查未领取的投递补贴
		var unclaimedTasks int64
		s.db.Model(&models.Task{}).
			Where("courier_id = ? AND status = ? AND subsidy_claimed = ?", courierID, "delivered", false).
			Count(&unclaimedTasks)

		if unclaimedTasks > 0 {
			return true, float64(unclaimedTasks) * incentive.Value, fmt.Sprintf("%d个未领取投递补贴", unclaimedTasks)
		}

	case models.IncentiveTypeCommission:
		// 检查月度返佣
		return s.checkMonthlyCommission(courierID, incentive)

	case models.IncentiveTypeBonus:
		// 检查新手奖励等特殊奖金
		return s.checkSpecialBonus(courierID, incentive)
	}

	return false, 0, ""
}

// checkMonthlyCommission 检查月度返佣
func (s *CourierGrowthService) checkMonthlyCommission(courierID string, incentive models.CourierIncentive) (bool, float64, string) {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	var monthlyTasks int64
	var totalEarnings float64

	s.db.Model(&models.Task{}).
		Where("courier_id = ? AND status = ? AND completed_at >= ?", courierID, "delivered", startOfMonth).
		Count(&monthlyTasks)

	// 这里需要根据实际数据结构计算总收入
	// 简化实现，假设每单基础收入
	totalEarnings = float64(monthlyTasks) * 5.0 // 假设每单5元基础收入

	// 检查是否满足最小任务数
	var conditions map[string]interface{}
	json.Unmarshal([]byte(incentive.Conditions), &conditions)

	minTasks := int64(conditions["min_tasks"].(float64))
	rate := conditions["rate"].(float64)

	if monthlyTasks >= minTasks {
		commission := totalEarnings * rate
		return true, commission, fmt.Sprintf("本月完成%d单，返佣%.1f%%", monthlyTasks, rate*100)
	}

	return false, 0, ""
}

// checkSpecialBonus 检查特殊奖金
func (s *CourierGrowthService) checkSpecialBonus(courierID string, incentive models.CourierIncentive) (bool, float64, string) {
	var courier models.Courier
	if err := s.db.Where("user_id = ?", courierID).First(&courier).Error; err != nil {
		return false, 0, ""
	}

	// 检查是否是新手（注册不到30天）
	if courier.ApprovedAt != nil {
		daysSinceApproval := int(time.Since(*courier.ApprovedAt).Hours() / 24)
		if daysSinceApproval <= 30 {
			// 检查是否已经领取过新手奖励
			// 这里需要添加领取记录的检查逻辑
			return true, incentive.Value, "新手专享奖励"
		}
	}

	return false, 0, ""
}

// checkAutoAwardBadges 检查自动颁发的徽章
func (s *CourierGrowthService) checkAutoAwardBadges(courierID string) []models.CourierBadge {
	var availableBadges []models.CourierBadge

	for _, defaultBadge := range models.DefaultBadges {
		// 检查是否已经获得该徽章
		var earnedBadge models.CourierBadgeEarned
		if err := s.db.Where("courier_id = ? AND badge_id = ?", courierID, defaultBadge.ID).First(&earnedBadge).Error; err == nil {
			continue // 已经获得了这个徽章
		}

		// 检查是否满足获得条件
		if s.checkBadgeConditions(courierID, defaultBadge) {
			availableBadges = append(availableBadges, defaultBadge)
		}
	}

	return availableBadges
}

// checkBadgeConditions 检查徽章获得条件
func (s *CourierGrowthService) checkBadgeConditions(courierID string, badge models.CourierBadge) bool {
	var conditions map[string]interface{}
	if err := json.Unmarshal([]byte(badge.Conditions), &conditions); err != nil {
		return false
	}

	conditionType := conditions["type"].(string)

	switch conditionType {
	case "rating":
		// 检查评分条件
		targetRating := conditions["value"].(float64)
		_ = conditions["duration"].(string) // 暂时忽略duration

		var avgRating float64
		// 根据duration计算时间范围内的平均评分
		// 这里需要根据实际的评分数据结构实现

		return avgRating >= targetRating

	case "ranking":
		// 检查排名条件
		_ = conditions["value"].(float64) // 暂时忽略topPercent
		_ = conditions["period"].(string) // 暂时忽略period

		// 计算在指定期间内的排名百分比
		// 这里需要实现排名计算逻辑

		return false // 简化实现

	case "delivery_time":
		// 检查投递时间条件
		maxTime := conditions["value"].(float64)

		var avgTime float64
		// 计算平均投递时间
		// 这里需要根据实际的统计数据实现

		return avgTime <= maxTime

	case "distance":
		// 检查投递距离条件
		_ = conditions["value"].(float64) // 暂时忽略targetDistance
		_ = conditions["period"].(string) // 暂时忽略period

		// 计算指定期间内的总投递距离
		// 这里需要根据实际的统计数据实现

		return false // 简化实现

	case "service_duration":
		// 检查服务时长条件
		targetMonths := conditions["value"].(float64)

		var courier models.Courier
		if err := s.db.Where("user_id = ?", courierID).First(&courier).Error; err != nil {
			return false
		}

		if courier.ApprovedAt == nil {
			return false
		}

		months := time.Since(*courier.ApprovedAt).Hours() / 24 / 30
		return months >= targetMonths
	}

	return false
}

// ClaimIncentive 领取激励奖励
func (s *CourierGrowthService) ClaimIncentive(courierID string, incentiveType models.IncentiveType, reference string, amount *int) (map[string]interface{}, error) {
	switch incentiveType {
	case models.IncentiveTypeSubsidy:
		return s.claimSubsidy(courierID, reference)

	case models.IncentiveTypePoints:
		return s.claimPoints(courierID, reference, amount)

	case models.IncentiveTypeCommission:
		return s.claimCommission(courierID, reference)

	case models.IncentiveTypeBadge:
		return s.claimBadge(courierID, reference)

	default:
		return nil, errors.New("不支持的激励类型")
	}
}

// claimSubsidy 领取投递补贴
func (s *CourierGrowthService) claimSubsidy(courierID, taskID string) (map[string]interface{}, error) {
	// 检查任务是否存在且未领取补贴
	var task models.Task
	if err := s.db.Where("courier_id = ? AND task_id = ? AND status = ?", courierID, taskID, "delivered").First(&task).Error; err != nil {
		return nil, errors.New("任务不存在或未完成")
	}

	// 这里需要检查是否已经领取过补贴
	// 假设任务表中有subsidy_claimed字段

	// 更新任务状态为已领取补贴
	// 这里需要根据实际的数据结构实现

	// 记录积分交易
	points := 10 // 基础积分奖励
	s.addPoints(courierID, points, "投递补贴", taskID)

	return map[string]interface{}{
		"type":    "subsidy",
		"amount":  2.0, // 基础补贴金额
		"points":  points,
		"task_id": taskID,
		"message": "投递补贴领取成功",
	}, nil
}

// claimPoints 领取积分奖励
func (s *CourierGrowthService) claimPoints(courierID, reference string, amount *int) (map[string]interface{}, error) {
	pointsAmount := 10 // 默认积分
	if amount != nil {
		pointsAmount = *amount
	}

	if err := s.addPoints(courierID, pointsAmount, "任务完成奖励", reference); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"type":      "points",
		"amount":    pointsAmount,
		"reference": reference,
		"message":   "积分奖励领取成功",
	}, nil
}

// claimCommission 领取返佣
func (s *CourierGrowthService) claimCommission(_ string, reference string) (map[string]interface{}, error) {
	// 计算本月返佣金额
	// 这里需要根据实际的业务逻辑实现

	commissionAmount := 25.0 // 示例金额

	return map[string]interface{}{
		"type":      "commission",
		"amount":    commissionAmount,
		"reference": reference,
		"message":   "月度返佣领取成功",
	}, nil
}

// claimBadge 领取徽章
func (s *CourierGrowthService) claimBadge(courierID, badgeCode string) (map[string]interface{}, error) {
	// 查找徽章
	var badge models.CourierBadge
	if err := s.db.Where("code = ? AND is_active = ?", badgeCode, true).First(&badge).Error; err != nil {
		return nil, errors.New("徽章不存在")
	}

	// 检查是否已经获得
	var earnedBadge models.CourierBadgeEarned
	if err := s.db.Where("courier_id = ? AND badge_id = ?", courierID, badge.ID).First(&earnedBadge).Error; err == nil {
		return nil, errors.New("已经获得该徽章")
	}

	// 颁发徽章
	if err := s.AwardBadge(courierID, badgeCode, "自动获得", "", "system"); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"type":       "badge",
		"badge_code": badgeCode,
		"badge_name": badge.Name,
		"points":     badge.Points,
		"message":    "徽章获得成功",
	}, nil
}

// GetRanking 获取排行榜数据
func (s *CourierGrowthService) GetRanking(zoneType, zoneID, timeRange, _ string, limit int) ([]models.CourierRankingInfo, error) {
	// 这里需要根据实际的数据结构和业务逻辑实现排行榜查询
	// 简化实现，返回示例数据

	var ranking []models.CourierRankingInfo

	// 构建基础查询
	query := s.db.Model(&models.Courier{}).
		Select("user_id as courier_id, level, rating").
		Where("status = ?", "approved")

	// 根据区域类型过滤
	if zoneType != "" && zoneID != "" {
		// TODO: 实现区域过滤逻辑
		// 需要根据zoneType和zoneID对排行榜数据进行过滤
	}

	// 根据时间范围过滤
	switch timeRange {
	case "daily":
		query = query.Where("created_at >= ?", time.Now().AddDate(0, 0, -1))
	case "weekly":
		query = query.Where("created_at >= ?", time.Now().AddDate(0, 0, -7))
	case "monthly":
		query = query.Where("created_at >= ?", time.Now().AddDate(0, -1, 0))
	}

	// 执行查询
	var couriers []models.Courier
	if err := query.Limit(limit).Find(&couriers).Error; err != nil {
		return nil, err
	}

	// 构建排行榜数据
	for i, courier := range couriers {
		rankingInfo := models.CourierRankingInfo{
			CourierID: courier.UserID,
			Level:     courier.Level,
			Rank:      i + 1,
			// 这里需要从其他表获取更多统计数据
		}
		ranking = append(ranking, rankingInfo)
	}

	return ranking, nil
}

// UpdateTaskStatistics 更新任务统计
func (s *CourierGrowthService) UpdateTaskStatistics(courierID, taskID, action string, data map[string]interface{}) error {
	today := time.Now().Truncate(24 * time.Hour)

	// 获取或创建今日统计记录
	var stats models.CourierStatistics
	if err := s.db.Where("courier_id = ? AND date = ?", courierID, today).First(&stats).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			stats = models.CourierStatistics{
				CourierID: courierID,
				Date:      today,
			}
			s.db.Create(&stats)
		} else {
			return err
		}
	}

	// 根据动作更新统计
	switch action {
	case "accepted":
		stats.TasksAccepted++

	case "completed":
		stats.TasksCompleted++

		// 更新相关数据
		if deliveryTime, ok := data["delivery_time"].(*int); ok && deliveryTime != nil {
			stats.TotalDeliveryTime += *deliveryTime
		}

		if distance, ok := data["distance"].(*float64); ok && distance != nil {
			stats.DistanceTraveled += *distance
		}

		if earnings, ok := data["earnings_amount"].(*float64); ok && earnings != nil {
			stats.EarningsAmount += *earnings
		}

		// 重新计算完成率
		if stats.TasksAccepted > 0 {
			stats.CompletionRate = float64(stats.TasksCompleted) / float64(stats.TasksAccepted) * 100
		}

		// 自动奖励积分
		s.addPoints(courierID, 10, "任务完成", taskID)

	case "failed":
		stats.TasksFailed++

		// 重新计算完成率
		if stats.TasksAccepted > 0 {
			stats.CompletionRate = float64(stats.TasksCompleted) / float64(stats.TasksAccepted) * 100
		}
	}

	// 保存更新
	if err := s.db.Save(&stats).Error; err != nil {
		return err
	}

	// 检查是否可以自动颁发徽章
	go s.checkAndAwardBadges(courierID)

	return nil
}

// GetEarnedBadges 获取已获得徽章
func (s *CourierGrowthService) GetEarnedBadges(courierID string) ([]models.CourierBadgeEarned, error) {
	var badges []models.CourierBadgeEarned
	if err := s.db.Preload("Badge").Where("courier_id = ?", courierID).Order("earned_at DESC").Find(&badges).Error; err != nil {
		return nil, err
	}
	return badges, nil
}

// AwardBadge 颁发徽章
func (s *CourierGrowthService) AwardBadge(courierID, badgeCode, reason, reference, _ string) error {
	// 查找徽章
	var badge models.CourierBadge
	if err := s.db.Where("code = ? AND is_active = ?", badgeCode, true).First(&badge).Error; err != nil {
		return errors.New("徽章不存在")
	}

	// 检查是否已经获得
	var earnedBadge models.CourierBadgeEarned
	if err := s.db.Where("courier_id = ? AND badge_id = ?", courierID, badge.ID).First(&earnedBadge).Error; err == nil {
		return errors.New("已经获得该徽章")
	}

	// 创建获得记录
	earnedBadge = models.CourierBadgeEarned{
		CourierID: courierID,
		BadgeID:   badge.ID,
		EarnedAt:  time.Now(),
		Reason:    reason,
		Reference: reference,
	}

	if err := s.db.Create(&earnedBadge).Error; err != nil {
		return err
	}

	// 奖励积分
	if badge.Points > 0 {
		s.addPoints(courierID, badge.Points, "徽章奖励: "+badge.Name, badgeCode)
	}

	// 发送通知
	s.notifyBadgeAwarded(courierID, badge, reason)

	return nil
}

// GetPointsBalance 获取积分余额
func (s *CourierGrowthService) GetPointsBalance(courierID string) (*models.CourierPoints, error) {
	return s.getOrCreatePoints(courierID)
}

// GetPointsHistory 获取积分交易历史
func (s *CourierGrowthService) GetPointsHistory(courierID, transactionType string, limit, offset int) ([]models.CourierPointsTransaction, int64, error) {
	var transactions []models.CourierPointsTransaction
	var total int64

	query := s.db.Model(&models.CourierPointsTransaction{}).Where("courier_id = ?", courierID)

	if transactionType != "" {
		query = query.Where("type = ?", transactionType)
	}

	// 获取总数
	query.Count(&total)

	// 获取分页数据
	if err := query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&transactions).Error; err != nil {
		return nil, 0, err
	}

	return transactions, total, nil
}

// GetAllBadges 获取所有徽章
func (s *CourierGrowthService) GetAllBadges() ([]models.CourierBadge, error) {
	var badges []models.CourierBadge
	if err := s.db.Where("is_active = ?", true).Order("rarity DESC, points DESC").Find(&badges).Error; err != nil {
		return nil, err
	}
	return badges, nil
}

// GetPerformanceStatistics 获取绩效统计
func (s *CourierGrowthService) GetPerformanceStatistics(courierID, timeRange, startDate, endDate string) (map[string]interface{}, error) {
	// 计算时间范围
	var since, until time.Time
	now := time.Now()

	if startDate != "" && endDate != "" {
		// 使用指定的日期范围
		var err error
		since, err = time.Parse("2006-01-02", startDate)
		if err != nil {
			return nil, errors.New("无效的开始日期格式")
		}
		until, err = time.Parse("2006-01-02", endDate)
		if err != nil {
			return nil, errors.New("无效的结束日期格式")
		}
	} else {
		// 使用时间范围
		switch timeRange {
		case "daily":
			since = now.AddDate(0, 0, -1)
		case "weekly":
			since = now.AddDate(0, 0, -7)
		case "monthly":
			since = now.AddDate(0, -1, 0)
		case "yearly":
			since = now.AddDate(-1, 0, 0)
		default:
			since = now.AddDate(0, -1, 0) // 默认一个月
		}
		until = now
	}

	// 查询统计数据
	var stats []models.CourierStatistics
	if err := s.db.Where("courier_id = ? AND date >= ? AND date <= ?", courierID, since, until).Find(&stats).Error; err != nil {
		return nil, err
	}

	// 计算汇总数据
	var totalAccepted, totalCompleted, totalFailed int
	var totalDeliveryTime int
	var totalDistance, totalEarnings float64
	var totalPoints int

	for _, stat := range stats {
		totalAccepted += stat.TasksAccepted
		totalCompleted += stat.TasksCompleted
		totalFailed += stat.TasksFailed
		totalDeliveryTime += stat.TotalDeliveryTime
		totalDistance += stat.DistanceTraveled
		totalEarnings += stat.EarningsAmount
		totalPoints += stat.PointsEarned
	}

	// 计算平均值
	var completionRate float64
	if totalAccepted > 0 {
		completionRate = float64(totalCompleted) / float64(totalAccepted) * 100
	}

	var avgDeliveryTime float64
	if totalCompleted > 0 {
		avgDeliveryTime = float64(totalDeliveryTime) / float64(totalCompleted)
	}

	return map[string]interface{}{
		"time_range":          timeRange,
		"start_date":          since.Format("2006-01-02"),
		"end_date":            until.Format("2006-01-02"),
		"total_accepted":      totalAccepted,
		"total_completed":     totalCompleted,
		"total_failed":        totalFailed,
		"completion_rate":     completionRate,
		"total_delivery_time": totalDeliveryTime,
		"avg_delivery_time":   avgDeliveryTime,
		"total_distance":      totalDistance,
		"total_earnings":      totalEarnings,
		"total_points":        totalPoints,
		"daily_stats":         stats,
	}, nil
}

// 辅助方法

// getOrCreatePoints 获取或创建积分记录
func (s *CourierGrowthService) getOrCreatePoints(courierID string) (*models.CourierPoints, error) {
	var points models.CourierPoints
	if err := s.db.Where("courier_id = ?", courierID).First(&points).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			points = models.CourierPoints{
				CourierID: courierID,
				Total:     0,
				Available: 0,
				Used:      0,
				Earned:    0,
			}
			if err := s.db.Create(&points).Error; err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	return &points, nil
}

// addPoints 增加积分
func (s *CourierGrowthService) addPoints(courierID string, amount int, description, reference string) error {
	// 获取积分记录
	points, err := s.getOrCreatePoints(courierID)
	if err != nil {
		return err
	}

	// 更新积分
	points.Total += amount
	points.Available += amount
	points.Earned += amount

	if err := s.db.Save(points).Error; err != nil {
		return err
	}

	// 创建交易记录
	transaction := models.CourierPointsTransaction{
		CourierID:   courierID,
		Type:        models.PointsEarn,
		Amount:      amount,
		Description: description,
		Reference:   reference,
	}

	return s.db.Create(&transaction).Error
}

// checkAndAwardBadges 检查并自动颁发徽章
func (s *CourierGrowthService) checkAndAwardBadges(courierID string) {
	badges := s.checkAutoAwardBadges(courierID)
	for _, badge := range badges {
		s.AwardBadge(courierID, badge.Code, "自动获得", "", "system")
	}
}

// 通知相关方法
func (s *CourierGrowthService) notifyBadgeAwarded(courierID string, badge models.CourierBadge, reason string) {
	event := utils.WebSocketEvent{
		Type: "BADGE_AWARDED",
		Data: map[string]interface{}{
			"courier_id":   courierID,
			"badge_code":   badge.Code,
			"badge_name":   badge.Name,
			"badge_rarity": badge.Rarity,
			"points":       badge.Points,
			"reason":       reason,
		},
		Timestamp: time.Now(),
	}
	s.wsManager.BroadcastToUser(courierID, event)
}
