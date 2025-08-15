package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"openpenpal/internal/models"
)

// CreditLimiterService 积分限制服务
type CreditLimiterService struct {
	db    *gorm.DB
	redis *redis.Client
	ctx   context.Context
}

// NewCreditLimiterService 创建积分限制服务
func NewCreditLimiterService(db *gorm.DB, redis *redis.Client) *CreditLimiterService {
	return &CreditLimiterService{
		db:    db,
		redis: redis,
		ctx:   context.Background(),
	}
}

// RateLimiter 限流器接口
type RateLimiter interface {
	CheckLimit(userID string, actionType string, points int) (*models.LimitStatus, error)
	RecordAction(userID string, actionType string, points int, metadata map[string]string) error
	GetLimitStatus(userID string, actionType string) (*models.LimitStatus, error)
	ResetUserLimits(userID string) error
}

// AntiFraudEngine 防作弊引擎接口
type AntiFraudEngine interface {
	DetectAnomalous(userID string, actionType string, metadata map[string]string) (*models.FraudAlert, error)
	GetRiskScore(userID string) (float64, error)
	UpdateRiskScore(userID string, increment float64) error
	BlockUser(userID string, reason string, duration time.Duration) error
	IsUserBlocked(userID string) (bool, error)
}

// CheckLimit 检查用户行为是否超过限制
func (s *CreditLimiterService) CheckLimit(userID string, actionType string, points int) (*models.LimitStatus, error) {
	// 1. 获取适用的限制规则
	rules, err := s.getApplicableRules(actionType)
	if err != nil {
		return nil, fmt.Errorf("failed to get rules: %w", err)
	}

	// 2. 检查每个规则
	for _, rule := range rules {
		status, err := s.checkRuleLimit(userID, rule, points)
		if err != nil {
			return nil, fmt.Errorf("failed to check rule %s: %w", rule.ID, err)
		}

		// 如果任何一个规则被违反，返回限制状态
		if status.IsLimited {
			return status, nil
		}
	}

	// 3. 如果没有违反任何规则，返回正常状态
	return s.createNormalStatus(actionType), nil
}

// RecordAction 记录用户行为
func (s *CreditLimiterService) RecordAction(userID string, actionType string, points int, metadata map[string]string) error {
	// 1. 创建行为记录
	action := &models.UserCreditAction{
		ID:         uuid.New().String(),
		UserID:     userID,
		ActionType: actionType,
		Points:     points,
		CreatedAt:  time.Now(),
	}

	// 2. 从metadata中提取信息
	if metadata != nil {
		action.IPAddress = metadata["ip_address"]
		action.DeviceID = metadata["device_id"]
		action.UserAgent = metadata["user_agent"]
		action.Reference = metadata["reference"]
	}

	// 3. 保存到数据库
	if err := s.db.Create(action).Error; err != nil {
		return fmt.Errorf("failed to record action: %w", err)
	}

	// 4. 更新Redis计数器
	if err := s.updateRedisCounters(userID, actionType, points); err != nil {
		log.Printf("Failed to update redis counters: %v", err)
		// 不返回错误，因为数据库已记录
	}

	return nil
}

// GetLimitStatus 获取用户当前限制状态
func (s *CreditLimiterService) GetLimitStatus(userID string, actionType string) (*models.LimitStatus, error) {
	rules, err := s.getApplicableRules(actionType)
	if err != nil {
		return nil, fmt.Errorf("failed to get rules: %w", err)
	}

	// 返回最严格的限制状态
	var strictestStatus *models.LimitStatus
	for _, rule := range rules {
		status, err := s.checkRuleLimit(userID, rule, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to check rule %s: %w", rule.ID, err)
		}

		if strictestStatus == nil || status.IsLimited {
			strictestStatus = status
		}
	}

	if strictestStatus == nil {
		return s.createNormalStatus(actionType), nil
	}

	return strictestStatus, nil
}

// DetectAnomalous 检测异常行为
func (s *CreditLimiterService) DetectAnomalous(userID string, actionType string, metadata map[string]string) (*models.FraudAlert, error) {
	// 1. 检查频率异常
	if alert := s.checkFrequencyAnomaly(userID, actionType); alert != nil {
		return alert, nil
	}

	// 2. 检查IP异常
	if alert := s.checkIPAnomaly(userID, metadata["ip_address"]); alert != nil {
		return alert, nil
	}

	// 3. 检查设备异常
	if alert := s.checkDeviceAnomaly(userID, metadata["device_id"]); alert != nil {
		return alert, nil
	}

	// 4. 检查积分异常
	if alert := s.checkPointsAnomaly(userID, actionType); alert != nil {
		return alert, nil
	}

	return nil, nil
}

// GetRiskScore 获取用户风险分数
func (s *CreditLimiterService) GetRiskScore(userID string) (float64, error) {
	var riskUser models.CreditRiskUser
	err := s.db.Where("user_id = ?", userID).First(&riskUser).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 新用户，默认低风险
			return 0.1, nil
		}
		return 0, fmt.Errorf("failed to get risk user: %w", err)
	}

	return riskUser.CalculateRiskScore(), nil
}

// UpdateRiskScore 更新用户风险分数
func (s *CreditLimiterService) UpdateRiskScore(userID string, increment float64) error {
	var riskUser models.CreditRiskUser
	err := s.db.Where("user_id = ?", userID).First(&riskUser).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 创建新的风险用户记录
			riskUser = models.CreditRiskUser{
				UserID:    userID,
				RiskScore: increment,
				RiskLevel: models.RiskLevelLow,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
		} else {
			return fmt.Errorf("failed to get risk user: %w", err)
		}
	} else {
		riskUser.RiskScore += increment
		riskUser.UpdatedAt = time.Now()
	}

	// 确保分数在合理范围内
	if riskUser.RiskScore < 0 {
		riskUser.RiskScore = 0
	}
	if riskUser.RiskScore > 1 {
		riskUser.RiskScore = 1
	}

	// 根据分数更新风险等级
	riskUser.RiskLevel = s.calculateRiskLevel(riskUser.RiskScore)

	return s.db.Save(&riskUser).Error
}

// BlockUser 封禁用户
func (s *CreditLimiterService) BlockUser(userID string, reason string, duration time.Duration) error {
	now := time.Now()
	blockedUntil := now.Add(duration)

	var riskUser models.CreditRiskUser
	err := s.db.Where("user_id = ?", userID).First(&riskUser).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			riskUser = models.CreditRiskUser{
				UserID:    userID,
				CreatedAt: now,
			}
		} else {
			return fmt.Errorf("failed to get risk user: %w", err)
		}
	}

	riskUser.RiskLevel = models.RiskLevelBlocked
	riskUser.RiskScore = 1.0
	riskUser.BlockedUntil = &blockedUntil
	riskUser.Reason = reason
	riskUser.UpdatedAt = now

	return s.db.Save(&riskUser).Error
}

// IsUserBlocked 检查用户是否被封禁
func (s *CreditLimiterService) IsUserBlocked(userID string) (bool, error) {
	var riskUser models.CreditRiskUser
	err := s.db.Where("user_id = ?", userID).First(&riskUser).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, fmt.Errorf("failed to get risk user: %w", err)
	}

	return riskUser.IsBlocked(), nil
}

// Private methods

// getApplicableRules 获取适用的规则
func (s *CreditLimiterService) getApplicableRules(actionType string) ([]*models.CreditLimitRule, error) {
	var rules []*models.CreditLimitRule
	err := s.db.Where("action_type = ? AND enabled = ?", actionType, true).
		Order("priority ASC").
		Find(&rules).Error
	if err != nil {
		return nil, err
	}

	return rules, nil
}

// checkRuleLimit 检查规则限制
func (s *CreditLimiterService) checkRuleLimit(userID string, rule *models.CreditLimitRule, additionalPoints int) (*models.LimitStatus, error) {
	now := time.Now()
	periodStart := rule.LimitPeriod.GetPeriodStart(now)
	periodEnd := rule.LimitPeriod.GetPeriodEnd(now)

	// 从数据库查询用户在此周期内的行为
	var count int64
	var totalPoints int
	
	query := s.db.Model(&models.UserCreditAction{}).
		Where("user_id = ? AND action_type = ? AND created_at >= ? AND created_at < ?",
			userID, rule.ActionType, periodStart, periodEnd)
	
	if err := query.Count(&count).Error; err != nil {
		return nil, fmt.Errorf("failed to count actions: %w", err)
	}

	// 计算总积分
	type Result struct {
		TotalPoints int
	}
	var result Result
	err := query.Select("COALESCE(SUM(points), 0) as total_points").Scan(&result).Error
	if err != nil {
		return nil, fmt.Errorf("failed to sum points: %w", err)
	}
	totalPoints = result.TotalPoints

	// 检查是否超过限制（包括即将添加的积分）
	isLimited := rule.IsLimitExceeded(int(count), totalPoints+additionalPoints)

	status := &models.LimitStatus{
		ActionType:    rule.ActionType,
		Period:        string(rule.LimitPeriod),
		CurrentCount:  int(count),
		MaxCount:      rule.MaxCount,
		CurrentPoints: totalPoints,
		MaxPoints:     rule.MaxPoints,
		IsLimited:     isLimited,
		ResetAt:       periodEnd,
	}

	return status, nil
}

// createNormalStatus 创建正常状态
func (s *CreditLimiterService) createNormalStatus(actionType string) *models.LimitStatus {
	return &models.LimitStatus{
		ActionType:    actionType,
		Period:        "daily",
		CurrentCount:  0,
		MaxCount:      1000, // 默认值
		CurrentPoints: 0,
		MaxPoints:     1000, // 默认值
		IsLimited:     false,
		ResetAt:       time.Now().AddDate(0, 0, 1),
	}
}

// updateRedisCounters 更新Redis计数器
func (s *CreditLimiterService) updateRedisCounters(userID string, actionType string, points int) error {
	now := time.Now()
	
	// 更新日计数器
	dailyKey := fmt.Sprintf("credit_limit:%s:%s:daily:%s", userID, actionType, now.Format("2006-01-02"))
	dailyTTL := time.Until(now.AddDate(0, 0, 1).Truncate(24 * time.Hour))
	
	pipe := s.redis.Pipeline()
	pipe.IncrBy(s.ctx, dailyKey+":count", 1)
	pipe.IncrBy(s.ctx, dailyKey+":points", int64(points))
	pipe.Expire(s.ctx, dailyKey+":count", dailyTTL)
	pipe.Expire(s.ctx, dailyKey+":points", dailyTTL)
	
	// 更新周计数器
	weekKey := fmt.Sprintf("credit_limit:%s:%s:weekly:%s", userID, actionType, now.Format("2006-01"))
	pipe.IncrBy(s.ctx, weekKey+":count", 1)
	pipe.IncrBy(s.ctx, weekKey+":points", int64(points))
	pipe.Expire(s.ctx, weekKey+":count", 7*24*time.Hour)
	pipe.Expire(s.ctx, weekKey+":points", 7*24*time.Hour)
	
	_, err := pipe.Exec(s.ctx)
	return err
}

// calculateRiskLevel 计算风险等级
func (s *CreditLimiterService) calculateRiskLevel(score float64) models.CreditRiskLevel {
	if score >= models.GetRiskThreshold(models.RiskLevelBlocked) {
		return models.RiskLevelBlocked
	} else if score >= models.GetRiskThreshold(models.RiskLevelHigh) {
		return models.RiskLevelHigh
	} else if score >= models.GetRiskThreshold(models.RiskLevelMedium) {
		return models.RiskLevelMedium
	}
	return models.RiskLevelLow
}

// Anomaly detection methods

func (s *CreditLimiterService) checkFrequencyAnomaly(userID string, actionType string) *models.FraudAlert {
	// 检查短时间内的高频操作
	since := time.Now().Add(-time.Minute * 5) // 5分钟内
	
	var count int64
	s.db.Model(&models.UserCreditAction{}).
		Where("user_id = ? AND action_type = ? AND created_at > ?", userID, actionType, since).
		Count(&count)
	
	// 如果5分钟内操作超过10次，认为异常
	if count > 10 {
		return &models.FraudAlert{
			UserID:      userID,
			AlertType:   models.AlertTypeFrequency,
			Severity:    models.SeverityHigh,
			Description: fmt.Sprintf("用户在5分钟内进行了%d次%s操作", count, actionType),
			Evidence: map[string]interface{}{
				"action_type": actionType,
				"count":       count,
				"time_window": "5min",
			},
			CreatedAt: time.Now(),
		}
	}
	
	return nil
}

func (s *CreditLimiterService) checkIPAnomaly(userID string, ipAddress string) *models.FraudAlert {
	if ipAddress == "" {
		return nil
	}
	
	// 检查用户是否在短时间内使用了多个IP
	since := time.Now().Add(-time.Hour) // 1小时内
	
	var ips []string
	s.db.Model(&models.UserCreditAction{}).
		Where("user_id = ? AND created_at > ? AND ip_address != ''", userID, since).
		Distinct("ip_address").
		Pluck("ip_address", &ips)
	
	// 如果1小时内使用了超过5个不同IP，认为异常
	if len(ips) > 5 {
		return &models.FraudAlert{
			UserID:      userID,
			AlertType:   models.AlertTypeIP,
			Severity:    models.SeverityMedium,
			Description: fmt.Sprintf("用户在1小时内使用了%d个不同IP地址", len(ips)),
			Evidence: map[string]interface{}{
				"ip_count":    len(ips),
				"time_window": "1hour",
				"ips":         ips,
			},
			CreatedAt: time.Now(),
		}
	}
	
	return nil
}

func (s *CreditLimiterService) checkDeviceAnomaly(userID string, deviceID string) *models.FraudAlert {
	if deviceID == "" {
		return nil
	}
	
	// 检查用户是否在短时间内使用了多个设备
	since := time.Now().Add(-time.Hour * 6) // 6小时内
	
	var devices []string
	s.db.Model(&models.UserCreditAction{}).
		Where("user_id = ? AND created_at > ? AND device_id != ''", userID, since).
		Distinct("device_id").
		Pluck("device_id", &devices)
	
	// 如果6小时内使用了超过3个不同设备，认为异常
	if len(devices) > 3 {
		return &models.FraudAlert{
			UserID:      userID,
			AlertType:   models.AlertTypeDevice,
			Severity:    models.SeverityMedium,
			Description: fmt.Sprintf("用户在6小时内使用了%d个不同设备", len(devices)),
			Evidence: map[string]interface{}{
				"device_count": len(devices),
				"time_window":  "6hours",
				"devices":      devices,
			},
			CreatedAt: time.Now(),
		}
	}
	
	return nil
}

func (s *CreditLimiterService) checkPointsAnomaly(userID string, actionType string) *models.FraudAlert {
	// 检查用户今日获得的积分是否异常
	today := time.Now().Truncate(24 * time.Hour)
	
	type Result struct {
		TotalPoints int
		Count       int64
	}
	var result Result
	
	s.db.Model(&models.UserCreditAction{}).
		Where("user_id = ? AND created_at >= ?", userID, today).
		Select("COALESCE(SUM(points), 0) as total_points, COUNT(*) as count").
		Scan(&result)
	
	// 如果今日获得积分超过500或操作次数超过50，认为异常
	if result.TotalPoints > 500 || result.Count > 50 {
		severity := models.SeverityMedium
		if result.TotalPoints > 1000 || result.Count > 100 {
			severity = models.SeverityHigh
		}
		
		return &models.FraudAlert{
			UserID:      userID,
			AlertType:   models.AlertTypePoints,
			Severity:    severity,
			Description: fmt.Sprintf("用户今日获得了%d积分，进行了%d次操作", result.TotalPoints, result.Count),
			Evidence: map[string]interface{}{
				"total_points": result.TotalPoints,
				"action_count": result.Count,
				"time_window":  "today",
			},
			CreatedAt: time.Now(),
		}
	}
	
	return nil
}