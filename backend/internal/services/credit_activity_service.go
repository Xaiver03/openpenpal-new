package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"openpenpal-backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/datatypes"
)

// CreditActivityService 积分活动服务
type CreditActivityService struct {
	db                     *gorm.DB
	creditService          *CreditService
	creditLimiterService   *CreditLimiterService
}

// NewCreditActivityService 创建积分活动服务实例
func NewCreditActivityService(db *gorm.DB, creditService *CreditService, creditLimiterService *CreditLimiterService) *CreditActivityService {
	return &CreditActivityService{
		db:                   db,
		creditService:        creditService,
		creditLimiterService: creditLimiterService,
	}
}

// ==================== 活动管理 ====================

// CreateActivity 创建积分活动
func (s *CreditActivityService) CreateActivity(activity *models.CreditActivity) error {
	// 验证活动数据
	if err := s.validateActivity(activity); err != nil {
		return fmt.Errorf("活动验证失败: %w", err)
	}

	// 设置默认值
	if activity.ID == uuid.Nil {
		activity.ID = uuid.New()
	}
	if activity.Status == "" {
		activity.Status = models.CreditActivityStatusDraft
	}

	// 开始事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 创建活动
	if err := tx.Create(activity).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("创建活动失败: %w", err)
	}

	// 如果活动有规则，创建规则
	if activity.TriggerRules != nil {
		rule := &models.CreditActivityRule{
			ActivityID:    activity.ID,
			RuleType:      string(activity.TriggerType),
			RuleName:      fmt.Sprintf("%s_trigger_rule", activity.Name),
			RuleCondition: activity.TriggerRules,
			RuleAction:    activity.RewardRules,
			Priority:      activity.Priority,
			IsActive:      true,
		}
		if err := tx.Create(rule).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("创建活动规则失败: %w", err)
		}
	}

	// 如果活动需要调度，创建调度记录
	if activity.Status == models.CreditActivityStatusPending {
		schedule := &models.CreditActivitySchedule{
			ActivityID:    activity.ID,
			ScheduledTime: activity.StartTime,
			Status:        "pending",
		}
		if err := tx.Create(schedule).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("创建活动调度失败: %w", err)
		}
	}

	// 创建统计记录
	stats := &models.CreditActivityStatistics{
		ActivityID: activity.ID,
	}
	if err := tx.Create(stats).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("创建活动统计失败: %w", err)
	}

	// 记录日志
	s.logActivityAction(tx, activity.ID, "", "create", map[string]interface{}{
		"activity_name": activity.Name,
		"created_by":    activity.CreatedBy,
	})

	return tx.Commit().Error
}

// GetActivityByID 获取活动详情
func (s *CreditActivityService) GetActivityByID(activityID uuid.UUID) (*models.CreditActivity, error) {
	var activity models.CreditActivity
	err := s.db.Where("id = ?", activityID).First(&activity).Error
	if err != nil {
		return nil, err
	}
	return &activity, nil
}

// GetActivities 获取活动列表
func (s *CreditActivityService) GetActivities(params map[string]interface{}) ([]models.CreditActivity, int64, error) {
	var activities []models.CreditActivity
	var total int64

	query := s.db.Model(&models.CreditActivity{})

	// 应用筛选条件
	if status, ok := params["status"].(models.CreditActivityStatus); ok {
		query = query.Where("status = ?", status)
	}
	if activityType, ok := params["activity_type"].(models.CreditActivityType); ok {
		query = query.Where("activity_type = ?", activityType)
	}
	if targetType, ok := params["target_type"].(models.CreditActivityTargetType); ok {
		query = query.Where("target_type = ?", targetType)
	}
	if keyword, ok := params["keyword"].(string); ok && keyword != "" {
		query = query.Where("name LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if activeOnly, ok := params["active_only"].(bool); ok && activeOnly {
		now := time.Now()
		query = query.Where("status = ? AND start_time <= ? AND end_time >= ?", 
			models.CreditActivityStatusActive, now, now)
	}

	// 计数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 排序
	if sortBy, ok := params["sort_by"].(string); ok {
		switch sortBy {
		case "priority_desc":
			query = query.Order("priority DESC, created_at DESC")
		case "start_time_asc":
			query = query.Order("start_time ASC")
		case "start_time_desc":
			query = query.Order("start_time DESC")
		case "created_at_desc":
			query = query.Order("created_at DESC")
		default:
			query = query.Order("priority DESC, created_at DESC")
		}
	} else {
		query = query.Order("priority DESC, created_at DESC")
	}

	// 分页
	page := 1
	limit := 20
	if p, ok := params["page"].(int); ok && p > 0 {
		page = p
	}
	if l, ok := params["limit"].(int); ok && l > 0 && l <= 100 {
		limit = l
	}
	offset := (page - 1) * limit
	query = query.Offset(offset).Limit(limit)

	// 查询
	if err := query.Find(&activities).Error; err != nil {
		return nil, 0, err
	}

	return activities, total, nil
}

// UpdateActivity 更新活动
func (s *CreditActivityService) UpdateActivity(activityID uuid.UUID, updates map[string]interface{}) error {
	// 检查活动是否存在
	var activity models.CreditActivity
	if err := s.db.Where("id = ?", activityID).First(&activity).Error; err != nil {
		return fmt.Errorf("活动不存在: %w", err)
	}

	// 验证更新数据
	if err := s.validateActivityUpdates(&activity, updates); err != nil {
		return fmt.Errorf("更新验证失败: %w", err)
	}

	// 记录更新人
	if updatedBy, ok := updates["updated_by"].(string); ok {
		updates["updated_by"] = updatedBy
	}
	updates["updated_at"] = time.Now()

	// 开始事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新活动
	if err := tx.Model(&activity).Updates(updates).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("更新活动失败: %w", err)
	}

	// 记录日志
	s.logActivityAction(tx, activityID, "", "update", updates)

	return tx.Commit().Error
}

// DeleteActivity 删除活动（软删除）
func (s *CreditActivityService) DeleteActivity(activityID uuid.UUID) error {
	// 检查活动是否存在
	var activity models.CreditActivity
	if err := s.db.Where("id = ?", activityID).First(&activity).Error; err != nil {
		return fmt.Errorf("活动不存在: %w", err)
	}

	// 检查活动是否可以删除
	if activity.Status == models.CreditActivityStatusActive {
		return errors.New("不能删除进行中的活动")
	}

	// 开始事务
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 软删除活动
	if err := tx.Delete(&activity).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("删除活动失败: %w", err)
	}

	// 记录日志
	s.logActivityAction(tx, activityID, "", "delete", nil)

	return tx.Commit().Error
}

// ==================== 活动状态管理 ====================

// StartActivity 启动活动
func (s *CreditActivityService) StartActivity(activityID uuid.UUID) error {
	var activity models.CreditActivity
	if err := s.db.Where("id = ?", activityID).First(&activity).Error; err != nil {
		return fmt.Errorf("活动不存在: %w", err)
	}

	// 验证活动状态
	if activity.Status != models.CreditActivityStatusPending {
		return fmt.Errorf("只有待开始的活动才能启动")
	}

	// 检查时间
	now := time.Now()
	if now.Before(activity.StartTime) {
		return fmt.Errorf("活动尚未到开始时间")
	}

	// 更新状态
	updates := map[string]interface{}{
		"status":     models.CreditActivityStatusActive,
		"updated_at": now,
	}

	if err := s.db.Model(&activity).Updates(updates).Error; err != nil {
		return fmt.Errorf("更新活动状态失败: %w", err)
	}

	// 记录日志
	s.logActivityAction(nil, activityID, "", "start", nil)

	return nil
}

// PauseActivity 暂停活动
func (s *CreditActivityService) PauseActivity(activityID uuid.UUID, reason string) error {
	var activity models.CreditActivity
	if err := s.db.Where("id = ?", activityID).First(&activity).Error; err != nil {
		return fmt.Errorf("活动不存在: %w", err)
	}

	// 验证活动状态
	if activity.Status != models.CreditActivityStatusActive {
		return fmt.Errorf("只有进行中的活动才能暂停")
	}

	// 更新状态
	updates := map[string]interface{}{
		"status":     models.CreditActivityStatusPaused,
		"updated_at": time.Now(),
	}

	if err := s.db.Model(&activity).Updates(updates).Error; err != nil {
		return fmt.Errorf("更新活动状态失败: %w", err)
	}

	// 记录日志
	s.logActivityAction(nil, activityID, "", "pause", map[string]interface{}{
		"reason": reason,
	})

	return nil
}

// ResumeActivity 恢复活动
func (s *CreditActivityService) ResumeActivity(activityID uuid.UUID) error {
	var activity models.CreditActivity
	if err := s.db.Where("id = ?", activityID).First(&activity).Error; err != nil {
		return fmt.Errorf("活动不存在: %w", err)
	}

	// 验证活动状态
	if activity.Status != models.CreditActivityStatusPaused {
		return fmt.Errorf("只有暂停的活动才能恢复")
	}

	// 检查活动是否已过期
	if time.Now().After(activity.EndTime) {
		return fmt.Errorf("活动已过期，无法恢复")
	}

	// 更新状态
	updates := map[string]interface{}{
		"status":     models.CreditActivityStatusActive,
		"updated_at": time.Now(),
	}

	if err := s.db.Model(&activity).Updates(updates).Error; err != nil {
		return fmt.Errorf("更新活动状态失败: %w", err)
	}

	// 记录日志
	s.logActivityAction(nil, activityID, "", "resume", nil)

	return nil
}

// CompleteActivity 结束活动
func (s *CreditActivityService) CompleteActivity(activityID uuid.UUID) error {
	var activity models.CreditActivity
	if err := s.db.Where("id = ?", activityID).First(&activity).Error; err != nil {
		return fmt.Errorf("活动不存在: %w", err)
	}

	// 更新状态
	updates := map[string]interface{}{
		"status":     models.CreditActivityStatusCompleted,
		"updated_at": time.Now(),
	}

	if err := s.db.Model(&activity).Updates(updates).Error; err != nil {
		return fmt.Errorf("更新活动状态失败: %w", err)
	}

	// 更新统计数据
	if err := s.updateActivityStatistics(activityID); err != nil {
		log.Printf("更新活动统计失败: %v", err)
	}

	// 记录日志
	s.logActivityAction(nil, activityID, "", "complete", nil)

	return nil
}

// ==================== 活动参与 ====================

// ParticipateActivity 参与活动
func (s *CreditActivityService) ParticipateActivity(userID string, activityID uuid.UUID, triggerData map[string]interface{}) (*models.CreditActivityParticipation, error) {
	// 获取活动信息
	activity, err := s.GetActivityByID(activityID)
	if err != nil {
		return nil, fmt.Errorf("获取活动信息失败: %w", err)
	}

	// 验证活动状态
	if !activity.IsActive() {
		return nil, errors.New("活动不在进行中")
	}

	// 检查用户是否符合目标条件
	if !s.checkTargetRules(userID, activity) {
		return nil, errors.New("用户不符合活动参与条件")
	}

	// 检查用户是否已参与
	var existingParticipation models.CreditActivityParticipation
	err = s.db.Where("activity_id = ? AND user_id = ?", activityID, userID).First(&existingParticipation).Error
	if err == nil {
		// 已参与，检查是否可以多次参与
		if activity.MaxRewardsPerUser > 0 && existingParticipation.Progress >= activity.MaxRewardsPerUser {
			return nil, errors.New("已达到最大参与次数")
		}
		// 更新进度
		return s.updateParticipation(&existingParticipation, activity, triggerData)
	}

	// 检查活动参与人数限制
	if activity.MaxParticipants > 0 {
		var participantCount int64
		s.db.Model(&models.CreditActivityParticipation{}).Where("activity_id = ?", activityID).Count(&participantCount)
		if int(participantCount) >= activity.MaxParticipants {
			return nil, errors.New("活动参与人数已满")
		}
	}

	// 将 triggerData 转换为 JSON
	progressDetailsJSON, err := json.Marshal(triggerData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal trigger data: %w", err)
	}

	// 创建参与记录
	participation := &models.CreditActivityParticipation{
		ActivityID:      activityID,
		UserID:          userID,
		ParticipatedAt:  time.Now(),
		Progress:        0,
		ProgressDetails: datatypes.JSON(progressDetailsJSON),
	}

	// 检查触发条件
	if s.checkTriggerRules(activity, triggerData) {
		// 计算奖励
		reward := s.calculateReward(activity, participation)
		if reward > 0 {
			// 检查预算
			if activity.Budget > 0 && activity.ConsumedBudget+reward > activity.Budget {
				return nil, errors.New("活动预算不足")
			}

			// 发放奖励
			if err := s.awardCredits(userID, activityID, reward, activity); err != nil {
				return nil, fmt.Errorf("发放奖励失败: %w", err)
			}

			participation.RewardCredits = reward
			participation.CompletedAt = &[]time.Time{time.Now()}[0]
			participation.Progress = 100
		}
	}

	// 保存参与记录
	if err := s.db.Create(participation).Error; err != nil {
		return nil, fmt.Errorf("创建参与记录失败: %w", err)
	}

	// 记录日志
	s.logActivityAction(nil, activityID, userID, "participate", triggerData)

	return participation, nil
}

// GetUserParticipations 获取用户参与记录
func (s *CreditActivityService) GetUserParticipations(userID string, params map[string]interface{}) ([]models.CreditActivityParticipation, int64, error) {
	var participations []models.CreditActivityParticipation
	var total int64

	query := s.db.Model(&models.CreditActivityParticipation{}).Where("user_id = ?", userID)

	// 预加载活动信息
	query = query.Preload("Activity")

	// 应用筛选条件
	if activityID, ok := params["activity_id"].(uuid.UUID); ok {
		query = query.Where("activity_id = ?", activityID)
	}
	if completed, ok := params["completed"].(bool); ok {
		if completed {
			query = query.Where("completed_at IS NOT NULL")
		} else {
			query = query.Where("completed_at IS NULL")
		}
	}

	// 计数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 排序
	query = query.Order("participated_at DESC")

	// 分页
	page := 1
	limit := 20
	if p, ok := params["page"].(int); ok && p > 0 {
		page = p
	}
	if l, ok := params["limit"].(int); ok && l > 0 && l <= 100 {
		limit = l
	}
	offset := (page - 1) * limit
	query = query.Offset(offset).Limit(limit)

	// 查询
	if err := query.Find(&participations).Error; err != nil {
		return nil, 0, err
	}

	return participations, total, nil
}

// ==================== 活动规则引擎 ====================

// EvaluateActivityTrigger 评估活动触发
func (s *CreditActivityService) EvaluateActivityTrigger(userID string, triggerType models.CreditActivityTriggerType, triggerData map[string]interface{}) error {
	// 获取所有匹配触发类型的活动
	var activities []models.CreditActivity
	now := time.Now()
	err := s.db.Where("status = ? AND trigger_type = ? AND start_time <= ? AND end_time >= ?",
		models.CreditActivityStatusActive, triggerType, now, now).Find(&activities).Error
	if err != nil {
		return err
	}

	// 评估每个活动
	for _, activity := range activities {
		// 检查用户是否符合目标条件
		if !s.checkTargetRules(userID, &activity) {
			continue
		}

		// 尝试参与活动
		if _, err := s.ParticipateActivity(userID, activity.ID, triggerData); err != nil {
			log.Printf("用户 %s 参与活动 %s 失败: %v", userID, activity.Name, err)
		}
	}

	return nil
}

// ProcessScheduledActivities 处理定时活动
func (s *CreditActivityService) ProcessScheduledActivities() error {
	// 获取待处理的调度
	var schedules []models.CreditActivitySchedule
	now := time.Now()
	err := s.db.Where("status = ? AND scheduled_time <= ?", "pending", now).
		Preload("Activity").Find(&schedules).Error
	if err != nil {
		return err
	}

	// 处理每个调度
	for _, schedule := range schedules {
		if err := s.processSchedule(&schedule); err != nil {
			log.Printf("处理活动调度 %s 失败: %v", schedule.ID, err)
			// 更新调度状态为失败
			s.db.Model(&schedule).Updates(map[string]interface{}{
				"status":        "failed",
				"error_message": err.Error(),
				"updated_at":    time.Now(),
			})
		}
	}

	return nil
}

// ==================== 活动统计 ====================

// GetActivityStatistics 获取活动统计
func (s *CreditActivityService) GetActivityStatistics(activityID uuid.UUID) (*models.CreditActivityStatistics, error) {
	var stats models.CreditActivityStatistics
	err := s.db.Where("activity_id = ?", activityID).First(&stats).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 创建新的统计记录
			stats = models.CreditActivityStatistics{
				ActivityID: activityID,
			}
			if err := s.db.Create(&stats).Error; err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	// 如果统计数据过期，更新统计
	if time.Since(stats.LastCalculatedAt) > time.Hour {
		if err := s.updateActivityStatistics(activityID); err != nil {
			log.Printf("更新活动统计失败: %v", err)
		}
		// 重新获取更新后的统计
		s.db.Where("activity_id = ?", activityID).First(&stats)
	}

	return &stats, nil
}

// GetAllActivitiesStatistics 获取所有活动统计概览
func (s *CreditActivityService) GetAllActivitiesStatistics() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 活动总数
	var totalActivities int64
	s.db.Model(&models.CreditActivity{}).Count(&totalActivities)
	stats["total_activities"] = totalActivities

	// 进行中的活动
	var activeActivities int64
	now := time.Now()
	s.db.Model(&models.CreditActivity{}).
		Where("status = ? AND start_time <= ? AND end_time >= ?", 
			models.CreditActivityStatusActive, now, now).
		Count(&activeActivities)
	stats["active_activities"] = activeActivities

	// 总参与人次
	var totalParticipations int64
	s.db.Model(&models.CreditActivityParticipation{}).Count(&totalParticipations)
	stats["total_participations"] = totalParticipations

	// 总发放积分
	var totalCreditsAwarded int64
	s.db.Model(&models.CreditActivityParticipation{}).
		Select("COALESCE(SUM(reward_credits), 0)").
		Row().Scan(&totalCreditsAwarded)
	stats["total_credits_awarded"] = totalCreditsAwarded

	// 活动类型分布
	var typeDistribution []struct {
		ActivityType string `json:"activity_type"`
		Count        int64  `json:"count"`
	}
	s.db.Model(&models.CreditActivity{}).
		Select("activity_type, COUNT(*) as count").
		Group("activity_type").
		Scan(&typeDistribution)
	stats["type_distribution"] = typeDistribution

	// 热门活动
	var popularActivities []struct {
		ActivityID    uuid.UUID `json:"activity_id"`
		ActivityName  string    `json:"activity_name"`
		Participants  int64     `json:"participants"`
		TotalCredits  int64     `json:"total_credits"`
	}
	s.db.Table("credit_activity_participations").
		Select("activity_id, a.name as activity_name, COUNT(DISTINCT user_id) as participants, COALESCE(SUM(reward_credits), 0) as total_credits").
		Joins("JOIN credit_activities a ON a.id = credit_activity_participations.activity_id").
		Where("a.deleted_at IS NULL").
		Group("activity_id, a.name").
		Order("participants DESC").
		Limit(10).
		Scan(&popularActivities)
	stats["popular_activities"] = popularActivities

	return stats, nil
}

// ==================== 活动模板 ====================

// CreateActivityTemplate 创建活动模板
func (s *CreditActivityService) CreateActivityTemplate(template *models.CreditActivityTemplate) error {
	if template.ID == uuid.Nil {
		template.ID = uuid.New()
	}
	return s.db.Create(template).Error
}

// GetActivityTemplates 获取活动模板列表
func (s *CreditActivityService) GetActivityTemplates(category string, isPublic bool) ([]models.CreditActivityTemplate, error) {
	var templates []models.CreditActivityTemplate
	query := s.db.Model(&models.CreditActivityTemplate{})

	if category != "" {
		query = query.Where("category = ?", category)
	}
	if isPublic {
		query = query.Where("is_public = ?", true)
	}

	query = query.Order("usage_count DESC, created_at DESC")

	err := query.Find(&templates).Error
	return templates, err
}

// CreateActivityFromTemplate 从模板创建活动
func (s *CreditActivityService) CreateActivityFromTemplate(templateID uuid.UUID, customData map[string]interface{}) (*models.CreditActivity, error) {
	// 获取模板
	var template models.CreditActivityTemplate
	if err := s.db.Where("id = ?", templateID).First(&template).Error; err != nil {
		return nil, fmt.Errorf("模板不存在: %w", err)
	}

	// 解析模板数据
	var templateData map[string]interface{}
	if err := json.Unmarshal(template.TemplateData, &templateData); err != nil {
		return nil, fmt.Errorf("解析模板数据失败: %w", err)
	}

	// 合并自定义数据
	for k, v := range customData {
		templateData[k] = v
	}

	// 创建活动
	activity := &models.CreditActivity{
		Name:             templateData["name"].(string),
		Description:      templateData["description"].(string),
		ActivityType:     models.CreditActivityType(templateData["activity_type"].(string)),
		TargetType:       models.CreditActivityTargetType(templateData["target_type"].(string)),
		TriggerType:      models.CreditActivityTriggerType(templateData["trigger_type"].(string)),
		RewardCredits:    int(templateData["reward_credits"].(float64)),
		Status:           models.CreditActivityStatusDraft,
	}

	// 设置其他字段
	if targetRules, ok := templateData["target_rules"]; ok {
		activity.TargetRules, _ = json.Marshal(targetRules)
	}
	if triggerRules, ok := templateData["trigger_rules"]; ok {
		activity.TriggerRules, _ = json.Marshal(triggerRules)
	}
	if rewardRules, ok := templateData["reward_rules"]; ok {
		activity.RewardRules, _ = json.Marshal(rewardRules)
	}

	// 创建活动
	if err := s.CreateActivity(activity); err != nil {
		return nil, err
	}

	// 更新模板使用次数
	s.db.Model(&template).UpdateColumn("usage_count", gorm.Expr("usage_count + ?", 1))

	return activity, nil
}

// ==================== 私有辅助方法 ====================

// validateActivity 验证活动数据
func (s *CreditActivityService) validateActivity(activity *models.CreditActivity) error {
	if activity.Name == "" {
		return errors.New("活动名称不能为空")
	}
	if activity.RewardCredits < 0 {
		return errors.New("奖励积分不能为负数")
	}
	if activity.StartTime.After(activity.EndTime) {
		return errors.New("开始时间不能晚于结束时间")
	}
	if activity.MaxRewardsPerUser < 0 {
		return errors.New("每用户最大奖励次数不能为负数")
	}
	if activity.Budget < 0 {
		return errors.New("活动预算不能为负数")
	}
	return nil
}

// validateActivityUpdates 验证活动更新数据
func (s *CreditActivityService) validateActivityUpdates(activity *models.CreditActivity, updates map[string]interface{}) error {
	// 检查是否允许更新状态
	if newStatus, ok := updates["status"].(models.CreditActivityStatus); ok {
		if !s.isValidStatusTransition(activity.Status, newStatus) {
			return fmt.Errorf("不允许从 %s 状态转换到 %s 状态", activity.Status, newStatus)
		}
	}

	// 检查时间更新
	if startTime, ok := updates["start_time"].(time.Time); ok {
		if endTime, ok := updates["end_time"].(time.Time); ok {
			if startTime.After(endTime) {
				return errors.New("开始时间不能晚于结束时间")
			}
		} else if startTime.After(activity.EndTime) {
			return errors.New("开始时间不能晚于结束时间")
		}
	}

	return nil
}

// isValidStatusTransition 检查状态转换是否有效
func (s *CreditActivityService) isValidStatusTransition(from, to models.CreditActivityStatus) bool {
	validTransitions := map[models.CreditActivityStatus][]models.CreditActivityStatus{
		models.CreditActivityStatusDraft:     {models.CreditActivityStatusPending, models.CreditActivityStatusCancelled},
		models.CreditActivityStatusPending:   {models.CreditActivityStatusActive, models.CreditActivityStatusCancelled},
		models.CreditActivityStatusActive:    {models.CreditActivityStatusPaused, models.CreditActivityStatusCompleted},
		models.CreditActivityStatusPaused:    {models.CreditActivityStatusActive, models.CreditActivityStatusCompleted, models.CreditActivityStatusCancelled},
		models.CreditActivityStatusCompleted: {},
		models.CreditActivityStatusCancelled: {},
	}

	allowedStatuses, ok := validTransitions[from]
	if !ok {
		return false
	}

	for _, status := range allowedStatuses {
		if status == to {
			return true
		}
	}

	return false
}

// checkTargetRules 检查用户是否符合目标规则
func (s *CreditActivityService) checkTargetRules(userID string, activity *models.CreditActivity) bool {
	// 如果是所有用户，直接返回true
	if activity.TargetType == models.CreditActivityTargetAll {
		return true
	}

	// 获取用户信息
	var user models.User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return false
	}

	// 根据目标类型检查
	switch activity.TargetType {
	case models.CreditActivityTargetNewUsers:
		// 检查是否为新用户（例如：注册时间在30天内）
		return time.Since(user.CreatedAt) <= 30*24*time.Hour
		
	case models.CreditActivityTargetLevel:
		// 从目标规则中获取等级要求
		if activity.TargetRules != nil {
			var targetRules map[string]interface{}
			if err := json.Unmarshal(activity.TargetRules, &targetRules); err == nil {
				if minLevel, ok := targetRules["min_level"].(float64); ok {
					// 获取用户积分信息
					userCredit, err := s.creditService.GetUserCredit(userID)
					if err != nil {
						return false
					}
					return userCredit.Level >= int(minLevel)
				}
			}
		}
		
	case models.CreditActivityTargetSchool:
		// 从目标规则中获取学校列表
		if activity.TargetRules != nil {
			var targetRules map[string]interface{}
			if err := json.Unmarshal(activity.TargetRules, &targetRules); err == nil {
				if schools, ok := targetRules["schools"].([]interface{}); ok {
					for _, school := range schools {
						if schoolCode, ok := school.(string); ok && schoolCode == user.SchoolCode {
							return true
						}
					}
				}
			}
		}
		
	case models.CreditActivityTargetCustom:
		// 自定义规则，需要更复杂的逻辑
		// 这里可以扩展更多自定义规则
		return true
	}

	return false
}

// checkTriggerRules 检查触发规则
func (s *CreditActivityService) checkTriggerRules(activity *models.CreditActivity, triggerData map[string]interface{}) bool {
	// 根据触发类型检查
	switch activity.TriggerType {
	case models.CreditActivityTriggerLogin:
		// 登录触发，直接返回true
		return true
		
	case models.CreditActivityTriggerLetter:
		// 写信触发，检查信件数据
		if letterID, ok := triggerData["letter_id"].(string); ok && letterID != "" {
			return true
		}
		
	case models.CreditActivityTriggerConsecutive:
		// 连续行为触发，检查连续天数
		if activity.TriggerRules != nil {
			var triggerRules map[string]interface{}
			if err := json.Unmarshal(activity.TriggerRules, &triggerRules); err == nil {
				if requiredDays, ok := triggerRules["consecutive_days"].(float64); ok {
					if consecutiveDays, ok := triggerData["consecutive_days"].(float64); ok {
						return consecutiveDays >= requiredDays
					}
				}
			}
		}
		
	case models.CreditActivityTriggerCumulative:
		// 累计行为触发，检查累计次数
		if activity.TriggerRules != nil {
			var triggerRules map[string]interface{}
			if err := json.Unmarshal(activity.TriggerRules, &triggerRules); err == nil {
				if requiredCount, ok := triggerRules["cumulative_count"].(float64); ok {
					if cumulativeCount, ok := triggerData["cumulative_count"].(float64); ok {
						return cumulativeCount >= requiredCount
					}
				}
			}
		}
	}

	return false
}

// calculateReward 计算奖励
func (s *CreditActivityService) calculateReward(activity *models.CreditActivity, participation *models.CreditActivityParticipation) int {
	baseReward := activity.RewardCredits

	// 如果有奖励规则，应用规则
	if activity.RewardRules != nil {
		var rewardRules map[string]interface{}
		if err := json.Unmarshal(activity.RewardRules, &rewardRules); err == nil {
			// 递增奖励
			if multiplier, ok := rewardRules["multiplier"].(float64); ok {
				baseReward = int(float64(baseReward) * multiplier)
			}
			
			// 基于进度的奖励
			if progressReward, ok := rewardRules["progress_based"].(bool); ok && progressReward {
				if participation.Progress > 0 {
					baseReward = baseReward * participation.Progress / 100
				}
			}
		}
	}

	return baseReward
}

// awardCredits 发放积分奖励
func (s *CreditActivityService) awardCredits(userID string, activityID uuid.UUID, credits int, activity *models.CreditActivity) error {
	// 使用积分服务发放积分
	err := s.creditService.AddCredits(userID, credits, 
		fmt.Sprintf("参与活动: %s", activity.Name), 
		fmt.Sprintf("activity_%s", activityID.String()))
	if err != nil {
		return err
	}

	// 更新活动消耗预算
	s.db.Model(activity).UpdateColumn("consumed_budget", gorm.Expr("consumed_budget + ?", credits))

	return nil
}

// updateParticipation 更新参与记录
func (s *CreditActivityService) updateParticipation(participation *models.CreditActivityParticipation, activity *models.CreditActivity, triggerData map[string]interface{}) (*models.CreditActivityParticipation, error) {
	// 更新进度
	participation.Progress++
	
	// 合并进度详情
	var progressDetails map[string]interface{}
	if participation.ProgressDetails != nil {
		json.Unmarshal(participation.ProgressDetails, &progressDetails)
	} else {
		progressDetails = make(map[string]interface{})
	}
	for k, v := range triggerData {
		progressDetails[k] = v
	}
	participation.ProgressDetails, _ = json.Marshal(progressDetails)

	// 检查是否完成
	if s.checkTriggerRules(activity, progressDetails) {
		reward := s.calculateReward(activity, participation)
		if reward > 0 {
			// 发放奖励
			if err := s.awardCredits(participation.UserID, activity.ID, reward, activity); err != nil {
				return nil, err
			}
			
			participation.RewardCredits += reward
			participation.CompletedAt = &[]time.Time{time.Now()}[0]
		}
	}

	// 更新记录
	if err := s.db.Save(participation).Error; err != nil {
		return nil, err
	}

	return participation, nil
}

// processSchedule 处理调度
func (s *CreditActivityService) processSchedule(schedule *models.CreditActivitySchedule) error {
	// 更新调度状态
	s.db.Model(schedule).Update("status", "executing")

	// 根据活动类型处理
	if schedule.Activity.TriggerType == models.CreditActivityTriggerScheduled {
		// 定时触发的活动，自动为符合条件的用户发放奖励
		if err := s.processScheduledRewards(schedule.Activity); err != nil {
			return err
		}
	}

	// 更新调度状态为完成
	now := time.Now()
	s.db.Model(schedule).Updates(map[string]interface{}{
		"status":       "completed",
		"executed_time": now,
		"updated_at":    now,
	})

	// 如果是重复活动，创建下一次调度
	if schedule.Activity.RepeatPattern != "" {
		nextSchedule := s.calculateNextSchedule(schedule.Activity)
		if nextSchedule != nil && (schedule.Activity.RepeatEndDate == nil || nextSchedule.Before(*schedule.Activity.RepeatEndDate)) {
			newSchedule := &models.CreditActivitySchedule{
				ActivityID:    schedule.ActivityID,
				ScheduledTime: *nextSchedule,
				Status:        "pending",
			}
			s.db.Create(newSchedule)
		}
	}

	return nil
}

// processScheduledRewards 处理定时奖励
func (s *CreditActivityService) processScheduledRewards(activity *models.CreditActivity) error {
	// 获取符合条件的用户
	var users []models.User
	query := s.db.Model(&models.User{})

	// 应用目标规则
	switch activity.TargetType {
	case models.CreditActivityTargetAll:
		// 所有活跃用户
		query = query.Where("is_active = ?", true)
	case models.CreditActivityTargetNewUsers:
		// 新用户
		query = query.Where("created_at >= ?", time.Now().AddDate(0, 0, -30))
	case models.CreditActivityTargetSchool:
		// 特定学校
		if activity.TargetRules != nil {
			var targetRules map[string]interface{}
			if err := json.Unmarshal(activity.TargetRules, &targetRules); err == nil {
				if schools, ok := targetRules["schools"].([]interface{}); ok {
					schoolCodes := make([]string, 0)
					for _, school := range schools {
						if code, ok := school.(string); ok {
							schoolCodes = append(schoolCodes, code)
						}
					}
					query = query.Where("school_code IN ?", schoolCodes)
				}
			}
		}
	}

	// 获取用户列表
	if err := query.Find(&users).Error; err != nil {
		return err
	}

	// 为每个用户创建参与记录
	for _, user := range users {
		_, err := s.ParticipateActivity(user.ID, activity.ID, map[string]interface{}{
			"trigger_type": "scheduled",
			"trigger_time": time.Now(),
		})
		if err != nil {
			log.Printf("处理用户 %s 的定时奖励失败: %v", user.ID, err)
		}
	}

	return nil
}

// calculateNextSchedule 计算下一次调度时间
func (s *CreditActivityService) calculateNextSchedule(activity *models.CreditActivity) *time.Time {
	now := time.Now()
	var next time.Time

	switch activity.RepeatPattern {
	case "daily":
		next = now.AddDate(0, 0, 1)
	case "weekly":
		next = now.AddDate(0, 0, 7)
	case "monthly":
		next = now.AddDate(0, 1, 0)
	default:
		if activity.RepeatInterval > 0 {
			next = now.Add(time.Duration(activity.RepeatInterval) * time.Hour)
		} else {
			return nil
		}
	}

	// 调整到活动的具体时间
	next = time.Date(next.Year(), next.Month(), next.Day(), 
		activity.StartTime.Hour(), activity.StartTime.Minute(), 0, 0, next.Location())

	return &next
}

// updateActivityStatistics 更新活动统计
func (s *CreditActivityService) updateActivityStatistics(activityID uuid.UUID) error {
	var stats models.CreditActivityStatistics
	
	// 获取或创建统计记录
	err := s.db.Where("activity_id = ?", activityID).First(&stats).Error
	if err == gorm.ErrRecordNotFound {
		stats = models.CreditActivityStatistics{
			ActivityID: activityID,
		}
	} else if err != nil {
		return err
	}

	// 更新统计数据
	var result struct {
		TotalParticipants     int64
		CompletedParticipants int64
		TotalCreditsAwarded   int64
		AverageCompletion     float64
	}

	s.db.Model(&models.CreditActivityParticipation{}).
		Where("activity_id = ?", activityID).
		Select("COUNT(DISTINCT user_id) as total_participants, " +
			"COUNT(DISTINCT CASE WHEN completed_at IS NOT NULL THEN user_id END) as completed_participants, " +
			"COALESCE(SUM(reward_credits), 0) as total_credits_awarded, " +
			"COALESCE(AVG(progress), 0) as average_completion").
		Row().Scan(&result.TotalParticipants, &result.CompletedParticipants, 
			&result.TotalCreditsAwarded, &result.AverageCompletion)

	stats.TotalParticipants = int(result.TotalParticipants)
	stats.CompletedParticipants = int(result.CompletedParticipants)
	stats.TotalCreditsAwarded = int(result.TotalCreditsAwarded)
	stats.AverageCompletion = result.AverageCompletion

	// 计算人气评分
	stats.PopularityScore = float64(stats.TotalParticipants) * 0.7 + 
		float64(stats.CompletedParticipants) * 0.3

	// 更新按等级和学校的参与统计
	// TODO: 实现更详细的统计

	stats.LastCalculatedAt = time.Now()

	// 保存统计
	return s.db.Save(&stats).Error
}

// logActivityAction 记录活动日志
func (s *CreditActivityService) logActivityAction(tx *gorm.DB, activityID uuid.UUID, userID string, action string, details interface{}) {
	activityLog := &models.CreditActivityLog{
		ActivityID: activityID,
		UserID:     userID,
		Action:     action,
		CreatedAt:  time.Now(),
	}

	if details != nil {
		activityLog.Details, _ = json.Marshal(details)
	}

	db := s.db
	if tx != nil {
		db = tx
	}

	if err := db.Create(activityLog).Error; err != nil {
		log.Printf("记录活动日志失败: %v", err)
	}
}