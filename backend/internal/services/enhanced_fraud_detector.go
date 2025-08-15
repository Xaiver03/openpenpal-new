package services

import (
	"fmt"
	"log"
	"math"
	"sort"
	"time"

	"openpenpal-backend/internal/models"

	"gorm.io/gorm"
)

// EnhancedFraudDetector Phase 1.3: 增强的防作弊检测系统
type EnhancedFraudDetector struct {
	db *gorm.DB
}

// NewEnhancedFraudDetector 创建增强防作弊检测器
func NewEnhancedFraudDetector(db *gorm.DB) *EnhancedFraudDetector {
	return &EnhancedFraudDetector{db: db}
}

// FraudDetectionResult 检测结果
type FraudDetectionResult struct {
	IsAnomalous      bool                   `json:"is_anomalous"`
	RiskScore        float64                `json:"risk_score"`        // 0-1之间的风险分数
	DetectedPatterns []string               `json:"detected_patterns"` // 检测到的异常模式
	Alerts           []*models.FraudAlert   `json:"alerts"`            // 具体的告警信息
	Recommendations  []string               `json:"recommendations"`   // 推荐的处理建议
	Evidence         map[string]interface{} `json:"evidence"`          // 证据数据
}

// DetectAdvancedFraud Phase 1.3: 高级防作弊检测
func (fd *EnhancedFraudDetector) DetectAdvancedFraud(userID string, actionType string, points int, metadata map[string]string) (*FraudDetectionResult, error) {
	result := &FraudDetectionResult{
		IsAnomalous:      false,
		RiskScore:        0.0,
		DetectedPatterns: []string{},
		Alerts:           []*models.FraudAlert{},
		Recommendations:  []string{},
		Evidence:         make(map[string]interface{}),
	}

	// 1. 行为频率模式检测
	if pattern, alert := fd.detectFrequencyPattern(userID, actionType, result.Evidence); pattern != "" {
		result.DetectedPatterns = append(result.DetectedPatterns, pattern)
		if alert != nil {
			result.Alerts = append(result.Alerts, alert)
		}
		result.RiskScore += 0.25
	}

	// 2. 时间序列异常检测
	if pattern, alert := fd.detectTimeSeriesAnomaly(userID, actionType, result.Evidence); pattern != "" {
		result.DetectedPatterns = append(result.DetectedPatterns, pattern)
		if alert != nil {
			result.Alerts = append(result.Alerts, alert)
		}
		result.RiskScore += 0.20
	}

	// 3. 地理位置异常检测
	if pattern, alert := fd.detectGeographicAnomaly(userID, metadata, result.Evidence); pattern != "" {
		result.DetectedPatterns = append(result.DetectedPatterns, pattern)
		if alert != nil {
			result.Alerts = append(result.Alerts, alert)
		}
		result.RiskScore += 0.15
	}

	// 4. 行为链路异常检测
	if pattern, alert := fd.detectBehaviorChainAnomaly(userID, actionType, result.Evidence); pattern != "" {
		result.DetectedPatterns = append(result.DetectedPatterns, pattern)
		if alert != nil {
			result.Alerts = append(result.Alerts, alert)
		}
		result.RiskScore += 0.30
	}

	// 5. 积分获取模式检测
	if pattern, alert := fd.detectPointsPattern(userID, points, result.Evidence); pattern != "" {
		result.DetectedPatterns = append(result.DetectedPatterns, pattern)
		if alert != nil {
			result.Alerts = append(result.Alerts, alert)
		}
		result.RiskScore += 0.20
	}

	// 6. 设备指纹异常检测
	if pattern, alert := fd.detectDeviceFingerprintAnomaly(userID, metadata, result.Evidence); pattern != "" {
		result.DetectedPatterns = append(result.DetectedPatterns, pattern)
		if alert != nil {
			result.Alerts = append(result.Alerts, alert)
		}
		result.RiskScore += 0.10
	}

	// 限制风险分数在0-1之间
	if result.RiskScore > 1.0 {
		result.RiskScore = 1.0
	}

	// 判断是否异常
	result.IsAnomalous = result.RiskScore > 0.6 || len(result.Alerts) > 0

	// 生成处理建议
	result.Recommendations = fd.generateRecommendations(result)

	return result, nil
}

// detectFrequencyPattern 检测频率模式异常
func (fd *EnhancedFraudDetector) detectFrequencyPattern(userID string, actionType string, evidence map[string]interface{}) (string, *models.FraudAlert) {
	// 获取用户最近的行为数据
	var actions []models.UserCreditAction
	since := time.Now().Add(-time.Hour) // 最近1小时
	
	err := fd.db.Where("user_id = ? AND action_type = ? AND created_at > ?", userID, actionType, since).
		Order("created_at ASC").
		Find(&actions).Error
	
	if err != nil || len(actions) < 3 {
		return "", nil
	}

	// 计算时间间隔的统计信息
	intervals := make([]time.Duration, len(actions)-1)
	for i := 1; i < len(actions); i++ {
		intervals[i-1] = actions[i].CreatedAt.Sub(actions[i-1].CreatedAt)
	}

	// 计算间隔的标准差
	avgInterval := fd.calculateAverageInterval(intervals)
	stdDev := fd.calculateStandardDeviation(intervals, avgInterval)

	evidence["frequency_analysis"] = map[string]interface{}{
		"action_count":    len(actions),
		"avg_interval":    avgInterval.Seconds(),
		"std_deviation":   stdDev.Seconds(),
		"time_window":     "1hour",
	}

	// 如果间隔太规律（标准差很小）且频率很高，可能是机器人行为
	if stdDev < time.Second*5 && avgInterval < time.Minute*2 && len(actions) > 10 {
		alert := &models.FraudAlert{
			UserID:      userID,
			AlertType:   models.AlertTypeFrequency,
			Severity:    models.SeverityHigh,
			Description: fmt.Sprintf("检测到机器人式规律行为：%d次操作，平均间隔%.1f秒", len(actions), avgInterval.Seconds()),
			Evidence: map[string]interface{}{
				"pattern_type":  "robot_behavior",
				"regularity":    stdDev.Seconds(),
				"frequency":     len(actions),
			},
			CreatedAt: time.Now(),
		}
		return "robot_behavior_pattern", alert
	}

	// 如果频率突然增高
	if len(actions) > 20 {
		return "high_frequency_burst", nil
	}

	return "", nil
}

// detectTimeSeriesAnomaly 时间序列异常检测
func (fd *EnhancedFraudDetector) detectTimeSeriesAnomaly(userID string, actionType string, evidence map[string]interface{}) (string, *models.FraudAlert) {
	// 获取用户过去7天的每日行为统计
	var dailyStats []struct {
		Date  time.Time
		Count int64
		Points int
	}

	for i := 6; i >= 0; i-- {
		day := time.Now().AddDate(0, 0, -i).Truncate(24 * time.Hour)
		dayEnd := day.Add(24 * time.Hour)

		var stat struct {
			Count  int64
			Points int
		}

		fd.db.Model(&models.UserCreditAction{}).
			Where("user_id = ? AND action_type = ? AND created_at >= ? AND created_at < ?", 
				userID, actionType, day, dayEnd).
			Select("COUNT(*) as count, COALESCE(SUM(points), 0) as points").
			Scan(&stat)

		dailyStats = append(dailyStats, struct {
			Date   time.Time
			Count  int64
			Points int
		}{Date: day, Count: stat.Count, Points: stat.Points})
	}

	evidence["time_series_analysis"] = dailyStats

	// 检测异常峰值
	todayCount := dailyStats[len(dailyStats)-1].Count
	avgCount := fd.calculateAverageCount(dailyStats[:len(dailyStats)-1])

	if float64(todayCount) > avgCount*3 && todayCount > 5 {
		alert := &models.FraudAlert{
			UserID:      userID,
			AlertType:   models.AlertTypeFrequency,
			Severity:    models.SeverityMedium,
			Description: fmt.Sprintf("今日行为频率异常：%d次，历史平均%.1f次", todayCount, avgCount),
			Evidence: map[string]interface{}{
				"today_count":   todayCount,
				"avg_count":     avgCount,
				"spike_ratio":   float64(todayCount) / avgCount,
			},
			CreatedAt: time.Now(),
		}
		return "frequency_spike", alert
	}

	return "", nil
}

// detectGeographicAnomaly 地理位置异常检测
func (fd *EnhancedFraudDetector) detectGeographicAnomaly(userID string, metadata map[string]string, evidence map[string]interface{}) (string, *models.FraudAlert) {
	ip := metadata["ip_address"]
	if ip == "" {
		return "", nil
	}

	// 获取用户最近的IP使用历史
	var recentIPs []string
	since := time.Now().Add(-time.Hour * 24) // 最近24小时

	fd.db.Model(&models.UserCreditAction{}).
		Where("user_id = ? AND created_at > ? AND ip_address != ''", userID, since).
		Distinct("ip_address").
		Pluck("ip_address", &recentIPs)

	evidence["geographic_analysis"] = map[string]interface{}{
		"current_ip":    ip,
		"recent_ips":    recentIPs,
		"ip_count":      len(recentIPs),
		"time_window":   "24hours",
	}

	// 检测IP跳跃异常
	if len(recentIPs) > 8 {
		alert := &models.FraudAlert{
			UserID:      userID,
			AlertType:   models.AlertTypeIP,
			Severity:    models.SeverityHigh,
			Description: fmt.Sprintf("24小时内使用了%d个不同IP地址，可能存在代理或VPN", len(recentIPs)),
			Evidence: map[string]interface{}{
				"ip_diversity": len(recentIPs),
				"pattern_type": "ip_hopping",
			},
			CreatedAt: time.Now(),
		}
		return "ip_hopping_pattern", alert
	}

	return "", nil
}

// detectBehaviorChainAnomaly 行为链路异常检测
func (fd *EnhancedFraudDetector) detectBehaviorChainAnomaly(userID string, actionType string, evidence map[string]interface{}) (string, *models.FraudAlert) {
	// 获取用户最近的行为序列
	var recentActions []models.UserCreditAction
	since := time.Now().Add(-time.Hour * 2) // 最近2小时

	err := fd.db.Where("user_id = ? AND created_at > ?", userID, since).
		Order("created_at ASC").
		Find(&recentActions).Error

	if err != nil || len(recentActions) < 5 {
		return "", nil
	}

	// 分析行为序列模式
	actionTypeSequence := make([]string, len(recentActions))
	for i, action := range recentActions {
		actionTypeSequence[i] = action.ActionType
	}

	evidence["behavior_chain_analysis"] = map[string]interface{}{
		"action_sequence": actionTypeSequence,
		"sequence_length": len(actionTypeSequence),
		"time_span":       "2hours",
	}

	// 检测重复模式
	patternLength := fd.detectRepeatingPattern(actionTypeSequence)
	if patternLength > 0 && patternLength < 5 {
		alert := &models.FraudAlert{
			UserID:      userID,
			AlertType:   models.AlertTypePattern,
			Severity:    models.SeverityMedium,
			Description: fmt.Sprintf("检测到长度为%d的重复行为模式", patternLength),
			Evidence: map[string]interface{}{
				"pattern_length": patternLength,
				"pattern_type":   "repeating_sequence",
			},
			CreatedAt: time.Now(),
		}
		return "repeating_behavior_pattern", alert
	}

	return "", nil
}

// detectPointsPattern 积分获取模式检测
func (fd *EnhancedFraudDetector) detectPointsPattern(userID string, points int, evidence map[string]interface{}) (string, *models.FraudAlert) {
	// 获取用户今日的积分记录
	today := time.Now().Truncate(24 * time.Hour)
	
	var pointsRecords []int
	fd.db.Model(&models.UserCreditAction{}).
		Where("user_id = ? AND created_at >= ?", userID, today).
		Pluck("points", &pointsRecords)

	if len(pointsRecords) == 0 {
		return "", nil
	}

	// 计算积分获取的统计信息
	totalPoints := 0
	for _, p := range pointsRecords {
		totalPoints += p
	}

	evidence["points_analysis"] = map[string]interface{}{
		"today_total":    totalPoints,
		"action_count":   len(pointsRecords),
		"avg_per_action": float64(totalPoints) / float64(len(pointsRecords)),
		"max_single":     fd.maxInt(pointsRecords),
		"min_single":     fd.minInt(pointsRecords),
	}

	// 检测积分获取异常
	if totalPoints > 1000 {
		severity := models.SeverityMedium
		if totalPoints > 2000 {
			severity = models.SeverityHigh
		}

		alert := &models.FraudAlert{
			UserID:      userID,
			AlertType:   models.AlertTypePoints,
			Severity:    severity,
			Description: fmt.Sprintf("今日积分获取异常：总计%d积分，%d次操作", totalPoints, len(pointsRecords)),
			Evidence: map[string]interface{}{
				"daily_total": totalPoints,
				"threshold":   1000,
			},
			CreatedAt: time.Now(),
		}
		return "excessive_points_pattern", alert
	}

	return "", nil
}

// detectDeviceFingerprintAnomaly 设备指纹异常检测
func (fd *EnhancedFraudDetector) detectDeviceFingerprintAnomaly(userID string, metadata map[string]string, evidence map[string]interface{}) (string, *models.FraudAlert) {
	deviceID := metadata["device_id"]
	userAgent := metadata["user_agent"]

	if deviceID == "" && userAgent == "" {
		return "", nil
	}

	// 获取用户最近的设备使用记录
	since := time.Now().Add(-time.Hour * 12) // 最近12小时

	var deviceInfo []struct {
		DeviceID  string
		UserAgent string
		Count     int64
	}

	fd.db.Model(&models.UserCreditAction{}).
		Where("user_id = ? AND created_at > ?", userID, since).
		Group("device_id, user_agent").
		Select("device_id, user_agent, COUNT(*) as count").
		Scan(&deviceInfo)

	evidence["device_analysis"] = map[string]interface{}{
		"current_device":  deviceID,
		"current_agent":   userAgent,
		"device_count":    len(deviceInfo),
		"device_history":  deviceInfo,
	}

	// 检测设备切换异常
	if len(deviceInfo) > 5 {
		alert := &models.FraudAlert{
			UserID:      userID,
			AlertType:   models.AlertTypeDevice,
			Severity:    models.SeverityMedium,
			Description: fmt.Sprintf("12小时内使用了%d个不同设备或浏览器", len(deviceInfo)),
			Evidence: map[string]interface{}{
				"device_diversity": len(deviceInfo),
				"pattern_type":     "device_switching",
			},
			CreatedAt: time.Now(),
		}
		return "device_switching_pattern", alert
	}

	return "", nil
}

// generateRecommendations 生成处理建议
func (fd *EnhancedFraudDetector) generateRecommendations(result *FraudDetectionResult) []string {
	recommendations := []string{}

	if result.RiskScore >= 0.8 {
		recommendations = append(recommendations, "建议立即暂停用户积分获取权限")
		recommendations = append(recommendations, "需要人工审核用户行为")
	} else if result.RiskScore >= 0.6 {
		recommendations = append(recommendations, "建议增加验证码验证")
		recommendations = append(recommendations, "临时降低积分获取速度")
	} else if result.RiskScore >= 0.4 {
		recommendations = append(recommendations, "建议加强监控")
		recommendations = append(recommendations, "记录详细行为日志")
	}

	for _, pattern := range result.DetectedPatterns {
		switch pattern {
		case "robot_behavior_pattern":
			recommendations = append(recommendations, "疑似机器人行为，建议要求完成图形验证码")
		case "ip_hopping_pattern":
			recommendations = append(recommendations, "IP地址异常，建议验证账户安全性")
		case "excessive_points_pattern":
			recommendations = append(recommendations, "积分获取异常，建议检查任务完成的真实性")
		}
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "行为正常，继续监控")
	}

	return recommendations
}

// Helper methods

func (fd *EnhancedFraudDetector) calculateAverageInterval(intervals []time.Duration) time.Duration {
	if len(intervals) == 0 {
		return 0
	}
	
	total := time.Duration(0)
	for _, interval := range intervals {
		total += interval
	}
	return total / time.Duration(len(intervals))
}

func (fd *EnhancedFraudDetector) calculateStandardDeviation(intervals []time.Duration, avg time.Duration) time.Duration {
	if len(intervals) == 0 {
		return 0
	}

	sum := 0.0
	for _, interval := range intervals {
		diff := float64(interval - avg)
		sum += diff * diff
	}
	
	variance := sum / float64(len(intervals))
	return time.Duration(math.Sqrt(variance))
}

func (fd *EnhancedFraudDetector) calculateAverageCount(stats []struct {
	Date   time.Time
	Count  int64
	Points int
}) float64 {
	if len(stats) == 0 {
		return 0
	}

	total := int64(0)
	for _, stat := range stats {
		total += stat.Count
	}
	return float64(total) / float64(len(stats))
}

func (fd *EnhancedFraudDetector) detectRepeatingPattern(sequence []string) int {
	if len(sequence) < 6 {
		return 0
	}

	// 检测长度2-5的重复模式
	for patternLen := 2; patternLen <= 5; patternLen++ {
		if len(sequence) < patternLen*3 {
			continue
		}

		pattern := sequence[:patternLen]
		repeats := 1

		for i := patternLen; i+patternLen <= len(sequence); i += patternLen {
			nextPattern := sequence[i : i+patternLen]
			if fd.sliceEqual(pattern, nextPattern) {
				repeats++
			} else {
				break
			}
		}

		if repeats >= 3 {
			return patternLen
		}
	}

	return 0
}

func (fd *EnhancedFraudDetector) sliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func (fd *EnhancedFraudDetector) maxInt(slice []int) int {
	if len(slice) == 0 {
		return 0
	}
	max := slice[0]
	for _, v := range slice[1:] {
		if v > max {
			max = v
		}
	}
	return max
}

func (fd *EnhancedFraudDetector) minInt(slice []int) int {
	if len(slice) == 0 {
		return 0
	}
	min := slice[0]
	for _, v := range slice[1:] {
		if v < min {
			min = v
		}
	}
	return min
}