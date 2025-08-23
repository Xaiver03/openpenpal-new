package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"openpenpal-backend/internal/models"
)

// CreditLimiterService 积分限制服务
type CreditLimiterService struct {
	db              *gorm.DB
	redis           *redis.Client
	ctx             context.Context
	enhancedDetector *EnhancedFraudDetector // Phase 1.3: 增强防作弊检测器
}

// NewCreditLimiterService 创建积分限制服务
func NewCreditLimiterService(db *gorm.DB, redis *redis.Client) *CreditLimiterService {
	return &CreditLimiterService{
		db:               db,
		redis:            redis,
		ctx:              context.Background(),
		enhancedDetector: NewEnhancedFraudDetector(db), // Phase 1.3: 初始化增强检测器
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

// DetectAnomalous 检测异常行为 (Phase 1.3: 增强版本)
func (s *CreditLimiterService) DetectAnomalous(userID string, actionType string, metadata map[string]string) (*models.FraudAlert, error) {
	// 使用增强检测器进行深度分析
	if s.enhancedDetector != nil {
		// 获取积分数据（从metadata或默认值）
		points := 10 // 默认积分
		if pointsStr, exists := metadata["points"]; exists {
			if p, err := strconv.Atoi(pointsStr); err == nil {
				points = p
			}
		}

		result, err := s.enhancedDetector.DetectAdvancedFraud(userID, actionType, points, metadata)
		if err != nil {
			log.Printf("Enhanced fraud detection failed: %v", err)
			// 降级到基础检测
			return s.detectAnomalousBasic(userID, actionType, metadata)
		}

		// 如果检测到高风险，返回最严重的告警
		if result.IsAnomalous && len(result.Alerts) > 0 {
			// 更新用户风险分数
			if err := s.UpdateRiskScore(userID, result.RiskScore*0.1); err != nil {
				log.Printf("Failed to update risk score: %v", err)
			}

			// 返回最高优先级的告警
			highestAlert := result.Alerts[0]
			for _, alert := range result.Alerts {
				if alert.Severity == models.SeverityHigh {
					highestAlert = alert
					break
				}
			}

			// 记录检测结果到数据库
			go s.logDetectionResult(userID, actionType, result)

			return highestAlert, nil
		}
	}

	// 降级到基础检测
	return s.detectAnomalousBasic(userID, actionType, metadata)
}

// detectAnomalousBasic 基础异常检测（原有逻辑）
func (s *CreditLimiterService) detectAnomalousBasic(userID string, actionType string, metadata map[string]string) (*models.FraudAlert, error) {
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

// Phase 1.3: 增强防作弊检测相关方法

// GetDetectionResult 获取增强检测结果
func (s *CreditLimiterService) GetDetectionResult(userID string, actionType string, points int, metadata map[string]string) (*FraudDetectionResult, error) {
	if s.enhancedDetector == nil {
		return nil, fmt.Errorf("enhanced detector not available")
	}
	
	return s.enhancedDetector.DetectAdvancedFraud(userID, actionType, points, metadata)
}

// logDetectionResult 记录检测结果
func (s *CreditLimiterService) logDetectionResult(userID string, actionType string, result *FraudDetectionResult) {
	// 将检测结果序列化存储
	evidenceJSON, _ := json.Marshal(result.Evidence)
	patternsJSON, _ := json.Marshal(result.DetectedPatterns)
	recommendationsJSON, _ := json.Marshal(result.Recommendations)

	// 创建检测日志记录
	detectionLog := models.FraudDetectionLog{
		ID:              uuid.New().String(),
		UserID:          userID,
		ActionType:      actionType,
		RiskScore:       result.RiskScore,
		IsAnomalous:     result.IsAnomalous,
		DetectedPatterns: string(patternsJSON),
		Evidence:        string(evidenceJSON),
		Recommendations: string(recommendationsJSON),
		AlertCount:      len(result.Alerts),
		CreatedAt:       time.Now(),
	}

	if err := s.db.Create(&detectionLog).Error; err != nil {
		log.Printf("Failed to log detection result: %v", err)
	}

	// 如果风险分数很高，同时更新用户风险等级
	if result.RiskScore >= 0.8 {
		log.Printf("High risk user detected: %s (score: %.2f)", userID, result.RiskScore)
		
		// 考虑自动封禁高风险用户
		if result.RiskScore >= 0.9 {
			reason := fmt.Sprintf("自动检测到高风险行为，风险分数: %.2f", result.RiskScore)
			duration := time.Hour * 24 // 24小时封禁
			if err := s.BlockUser(userID, reason, duration); err != nil {
				log.Printf("Failed to auto-block high risk user: %v", err)
			} else {
				log.Printf("Auto-blocked user %s for 24 hours due to high risk score", userID)
			}
		}
	}
}

// GetUserRiskAnalysis 获取用户风险分析报告
func (s *CreditLimiterService) GetUserRiskAnalysis(userID string) (*UserRiskAnalysis, error) {
	analysis := &UserRiskAnalysis{
		UserID:    userID,
		Timestamp: time.Now(),
	}

	// 获取用户基础风险信息
	riskScore, err := s.GetRiskScore(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get risk score: %w", err)
	}
	analysis.CurrentRiskScore = riskScore

	// 获取最近的检测日志
	var recentLogs []models.FraudDetectionLog
	since := time.Now().Add(-time.Hour * 24 * 7) // 最近7天
	err = s.db.Where("user_id = ? AND created_at > ?", userID, since).
		Order("created_at DESC").
		Limit(10).
		Find(&recentLogs).Error
	
	if err != nil {
		log.Printf("Failed to get recent logs: %v", err)
	} else {
		analysis.RecentDetections = len(recentLogs)
		
		// 分析检测趋势
		anomalousCount := 0
		totalRiskScore := 0.0
		for _, logEntry := range recentLogs {
			if logEntry.IsAnomalous {
				anomalousCount++
			}
			totalRiskScore += logEntry.RiskScore
		}
		
		if len(recentLogs) > 0 {
			analysis.AnomalousRate = float64(anomalousCount) / float64(len(recentLogs))
			analysis.AvgRiskScore = totalRiskScore / float64(len(recentLogs))
		}
	}

	// 获取用户行为统计
	var actionStats struct {
		TotalActions int64
		TotalPoints  int
		UniqueIPs    int64
		UniqueDevices int64
	}

	since = time.Now().Add(-time.Hour * 24 * 30) // 最近30天
	s.db.Model(&models.UserCreditAction{}).
		Where("user_id = ? AND created_at > ?", userID, since).
		Select("COUNT(*) as total_actions, COALESCE(SUM(points), 0) as total_points").
		Scan(&actionStats)

	s.db.Model(&models.UserCreditAction{}).
		Where("user_id = ? AND created_at > ? AND ip_address != ''", userID, since).
		Count(&actionStats.UniqueIPs)

	s.db.Model(&models.UserCreditAction{}).
		Where("user_id = ? AND created_at > ? AND device_id != ''", userID, since).
		Count(&actionStats.UniqueDevices)

	analysis.MonthlyActions = actionStats.TotalActions
	analysis.MonthlyPoints = actionStats.TotalPoints
	analysis.UniqueIPs = actionStats.UniqueIPs
	analysis.UniqueDevices = actionStats.UniqueDevices

	// 计算风险等级
	analysis.RiskLevel = s.calculateRiskLevel(riskScore)

	// 生成建议
	analysis.Recommendations = s.generateRiskRecommendations(analysis)

	return analysis, nil
}

// UserRiskAnalysis 用户风险分析结果
type UserRiskAnalysis struct {
	UserID            string              `json:"user_id"`
	Timestamp         time.Time           `json:"timestamp"`
	CurrentRiskScore  float64             `json:"current_risk_score"`
	RiskLevel         models.RiskLevel    `json:"risk_level"`
	RecentDetections  int                 `json:"recent_detections"`
	AnomalousRate     float64             `json:"anomalous_rate"`
	AvgRiskScore      float64             `json:"avg_risk_score"`
	MonthlyActions    int64               `json:"monthly_actions"`
	MonthlyPoints     int                 `json:"monthly_points"`
	UniqueIPs         int64               `json:"unique_ips"`
	UniqueDevices     int64               `json:"unique_devices"`
	Recommendations   []string            `json:"recommendations"`
}

// generateRiskRecommendations 生成风险管理建议
func (s *CreditLimiterService) generateRiskRecommendations(analysis *UserRiskAnalysis) []string {
	recommendations := []string{}

	if analysis.CurrentRiskScore >= 0.8 {
		recommendations = append(recommendations, "用户风险极高，建议立即人工审核")
		recommendations = append(recommendations, "暂停所有积分获取权限")
	} else if analysis.CurrentRiskScore >= 0.6 {
		recommendations = append(recommendations, "用户风险较高，建议加强监控")
		recommendations = append(recommendations, "要求额外验证步骤")
	} else if analysis.CurrentRiskScore >= 0.4 {
		recommendations = append(recommendations, "用户风险中等，建议定期检查")
	}

	if analysis.AnomalousRate > 0.5 {
		recommendations = append(recommendations, "异常行为频率过高，需要详细调查")
	}

	if analysis.UniqueIPs > 10 {
		recommendations = append(recommendations, "IP地址变化频繁，可能使用代理或VPN")
	}

	if analysis.UniqueDevices > 5 {
		recommendations = append(recommendations, "设备变化频繁，需要验证账户安全性")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "用户行为正常，继续常规监控")
	}

	return recommendations
}

// DB 提供数据库访问（用于handlers）
func (s *CreditLimiterService) DB() *gorm.DB {
	return s.db
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