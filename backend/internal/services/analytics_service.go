package services

import (
	"encoding/json"
	"fmt"
	"math"
	"openpenpal-backend/internal/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AnalyticsService 数据分析服务
type AnalyticsService struct {
	db *gorm.DB
}

// NewAnalyticsService 创建数据分析服务实例
func NewAnalyticsService(db *gorm.DB) *AnalyticsService {
	return &AnalyticsService{db: db}
}

// RecordMetric 记录分析指标
func (s *AnalyticsService) RecordMetric(metricType models.AnalyticsMetricType, name string, value float64,
	unit string, dimension string, granularity models.AnalyticsGranularity, metadata map[string]interface{}) error {

	metadataJSON := "{}"
	if metadata != nil {
		if data, err := json.Marshal(metadata); err == nil {
			metadataJSON = string(data)
		}
	}

	metric := &models.AnalyticsMetric{
		ID:          uuid.New().String(),
		MetricType:  metricType,
		MetricName:  name,
		Value:       value,
		Unit:        unit,
		Dimension:   dimension,
		Granularity: granularity,
		Timestamp:   time.Now(),
		Metadata:    metadataJSON,
		CreatedAt:   time.Now(),
	}

	return s.db.Create(metric).Error
}

// GetMetrics 获取分析指标
func (s *AnalyticsService) GetMetrics(query *models.AnalyticsQuery) ([]models.AnalyticsMetric, error) {
	var metrics []models.AnalyticsMetric

	db := s.db.Model(&models.AnalyticsMetric{})

	// 应用过滤条件
	if query.MetricType != "" {
		db = db.Where("metric_type = ?", query.MetricType)
	}
	if query.Granularity != "" {
		db = db.Where("granularity = ?", query.Granularity)
	}
	if !query.StartDate.IsZero() {
		db = db.Where("timestamp >= ?", query.StartDate)
	}
	if !query.EndDate.IsZero() {
		db = db.Where("timestamp <= ?", query.EndDate)
	}
	if query.Dimension != "" {
		db = db.Where("dimension = ?", query.Dimension)
	}

	// 排序和限制
	db = db.Order("timestamp DESC")
	if query.Limit > 0 {
		db = db.Limit(query.Limit)
	}

	err := db.Find(&metrics).Error
	return metrics, err
}

// GetMetricSummary 获取指标摘要
func (s *AnalyticsService) GetMetricSummary(metricType models.AnalyticsMetricType,
	metricName string, startDate, endDate time.Time) (*models.MetricSummary, error) {

	var result struct {
		Total   float64 `json:"total"`
		Average float64 `json:"average"`
		Maximum float64 `json:"maximum"`
		Minimum float64 `json:"minimum"`
		Count   int     `json:"count"`
		Unit    string  `json:"unit"`
	}

	err := s.db.Model(&models.AnalyticsMetric{}).
		Select(`
			SUM(value) as total,
			AVG(value) as average,
			MAX(value) as maximum,
			MIN(value) as minimum,
			COUNT(*) as count,
			MAX(unit) as unit
		`).
		Where("metric_type = ? AND metric_name = ? AND timestamp BETWEEN ? AND ?",
			metricType, metricName, startDate, endDate).
		Scan(&result).Error

	if err != nil {
		return nil, err
	}

	return &models.MetricSummary{
		MetricName: metricName,
		Total:      result.Total,
		Average:    result.Average,
		Maximum:    result.Maximum,
		Minimum:    result.Minimum,
		Count:      result.Count,
		Unit:       result.Unit,
	}, nil
}

// UpdateUserAnalytics 更新用户分析数据
func (s *AnalyticsService) UpdateUserAnalytics(userID string, date time.Time) error {
	// 获取或创建用户分析记录
	var analytics models.UserAnalytics
	err := s.db.Where("user_id = ? AND date = ?", userID, date.Format("2006-01-02")).
		First(&analytics).Error

	if err == gorm.ErrRecordNotFound {
		analytics = models.UserAnalytics{
			ID:        uuid.New().String(),
			UserID:    userID,
			Date:      date,
			CreatedAt: time.Now(),
		}
	}

	// 计算当天的统计数据
	analytics = s.calculateUserDailyStats(userID, date, analytics)
	analytics.UpdatedAt = time.Now()

	// 保存或更新
	return s.db.Save(&analytics).Error
}

// calculateUserDailyStats 计算用户当天统计数据
func (s *AnalyticsService) calculateUserDailyStats(userID string, date time.Time, analytics models.UserAnalytics) models.UserAnalytics {
	dateStr := date.Format("2006-01-02")

	// 计算发送信件数
	var sentCount int64
	s.db.Model(&models.Letter{}).
		Where("user_id = ? AND DATE(created_at) = ?", userID, dateStr).
		Count(&sentCount)
	analytics.LettersSent = int(sentCount)

	// 计算接收信件数（通过信件编码）
	var receivedCount int64
	s.db.Table("letters l").
		Joins("JOIN letter_codes lc ON l.id = lc.letter_id").
		Where("l.user_id != ? AND lc.recipient_id = ? AND DATE(l.created_at) = ?",
			userID, userID, dateStr).
		Count(&receivedCount)
	analytics.LettersReceived = int(receivedCount)

	// 计算已读信件数
	var readCount int64
	s.db.Model(&models.Letter{}).
		Where("user_id = ? AND status = ? AND DATE(updated_at) = ?",
			userID, models.StatusRead, dateStr).
		Count(&readCount)
	analytics.LettersRead = int(readCount)

	// 计算信使任务数
	var taskCount int64
	s.db.Model(&models.CourierTask{}).
		Where("courier_id = ? AND DATE(created_at) = ?", userID, dateStr).
		Count(&taskCount)
	analytics.CourierTasks = int(taskCount)

	// 计算参与度分数（简化算法）
	engagementScore := float64(analytics.LettersSent*3 + analytics.LettersReceived*2 +
		analytics.LettersRead*1 + analytics.CourierTasks*2)
	analytics.EngagementScore = engagementScore

	// 计算连续活跃天数
	analytics.RetentionDays = s.calculateRetentionDays(userID, date)

	return analytics
}

// calculateRetentionDays 计算连续活跃天数
func (s *AnalyticsService) calculateRetentionDays(userID string, currentDate time.Time) int {
	days := 0
	checkDate := currentDate

	for days < 365 { // 最多检查365天
		dateStr := checkDate.Format("2006-01-02")

		// 检查当天是否有活动
		var activityCount int64
		s.db.Model(&models.Letter{}).
			Where("user_id = ? AND DATE(created_at) = ?", userID, dateStr).
			Count(&activityCount)

		if activityCount == 0 {
			// 检查其他活动
			s.db.Model(&models.CourierTask{}).
				Where("courier_id = ? AND DATE(created_at) = ?", userID, dateStr).
				Count(&activityCount)
		}

		if activityCount == 0 {
			break
		}

		days++
		checkDate = checkDate.AddDate(0, 0, -1)
	}

	return days
}

// UpdateSystemAnalytics 更新系统分析数据
func (s *AnalyticsService) UpdateSystemAnalytics(date time.Time) error {
	var analytics models.SystemAnalytics
	err := s.db.Where("date = ?", date.Format("2006-01-02")).
		First(&analytics).Error

	if err == gorm.ErrRecordNotFound {
		analytics = models.SystemAnalytics{
			ID:        uuid.New().String(),
			Date:      date,
			CreatedAt: time.Now(),
		}
	}

	// 计算系统统计数据
	analytics = s.calculateSystemDailyStats(date, analytics)
	analytics.UpdatedAt = time.Now()

	return s.db.Save(&analytics).Error
}

// calculateSystemDailyStats 计算系统当天统计数据
func (s *AnalyticsService) calculateSystemDailyStats(date time.Time, analytics models.SystemAnalytics) models.SystemAnalytics {
	dateStr := date.Format("2006-01-02")

	// 活跃用户数（当天有任何活动的用户）
	var activeUsersCount int64
	s.db.Model(&models.UserAnalytics{}).
		Where("date = ? AND (letters_sent > 0 OR letters_received > 0 OR courier_tasks > 0)", dateStr).
		Count(&activeUsersCount)
	analytics.ActiveUsers = int(activeUsersCount)

	// 新用户数
	var newUsersCount int64
	s.db.Model(&models.User{}).
		Where("DATE(created_at) = ?", dateStr).
		Count(&newUsersCount)
	analytics.NewUsers = int(newUsersCount)

	// 总用户数
	var totalUsersCount int64
	s.db.Model(&models.User{}).
		Where("created_at <= ?", date.AddDate(0, 0, 1)).
		Count(&totalUsersCount)
	analytics.TotalUsers = int(totalUsersCount)

	// 创建的信件数
	var lettersCreatedCount int64
	s.db.Model(&models.Letter{}).
		Where("DATE(created_at) = ?", dateStr).
		Count(&lettersCreatedCount)
	analytics.LettersCreated = int(lettersCreatedCount)

	// 已送达的信件数
	var lettersDeliveredCount int64
	s.db.Model(&models.Letter{}).
		Where("status = ? AND DATE(updated_at) = ?", models.StatusDelivered, dateStr).
		Count(&lettersDeliveredCount)
	analytics.LettersDelivered = int(lettersDeliveredCount)

	// 完成的信使任务数
	var courierTasksCount int64
	s.db.Model(&models.CourierTask{}).
		Where("status = 'completed' AND DATE(updated_at) = ?", dateStr).
		Count(&courierTasksCount)
	analytics.CourierTasksCompleted = int(courierTasksCount)

	// 添加的博物馆物品数
	var museumItemsCount int64
	s.db.Model(&models.MuseumItem{}).
		Where("DATE(created_at) = ?", dateStr).
		Count(&museumItemsCount)
	analytics.MuseumItemsAdded = int(museumItemsCount)

	// 计算平均响应时间和错误率（从性能指标中）
	var avgResponse float64
	var errorRate float64

	s.db.Model(&models.PerformanceMetric{}).
		Where("DATE(timestamp) = ?", dateStr).
		Select("AVG(response_time)").
		Scan(&avgResponse)
	analytics.AvgResponseTime = avgResponse

	var totalRequests, errorRequests int64
	s.db.Model(&models.PerformanceMetric{}).
		Where("DATE(timestamp) = ?", dateStr).
		Count(&totalRequests)

	s.db.Model(&models.PerformanceMetric{}).
		Where("DATE(timestamp) = ? AND status_code >= 400", dateStr).
		Count(&errorRequests)

	if totalRequests > 0 {
		errorRate = float64(errorRequests) / float64(totalRequests) * 100
	}
	analytics.ErrorRate = errorRate

	// 假设服务器正常运行时间为99%（实际应该从监控系统获取）
	analytics.ServerUptime = 99.0

	return analytics
}

// RecordPerformanceMetric 记录性能指标
func (s *AnalyticsService) RecordPerformanceMetric(endpoint, method string,
	responseTime float64, statusCode int, userAgent, ipAddress string, userID *string) error {

	metric := &models.PerformanceMetric{
		ID:           uuid.New().String(),
		Endpoint:     endpoint,
		Method:       method,
		ResponseTime: responseTime,
		StatusCode:   statusCode,
		UserAgent:    userAgent,
		IPAddress:    ipAddress,
		UserID:       userID,
		Timestamp:    time.Now(),
		CreatedAt:    time.Now(),
	}

	return s.db.Create(metric).Error
}

// GenerateReport 生成分析报告
func (s *AnalyticsService) GenerateReport(req *models.GenerateReportRequest, generatedBy string) (*models.AnalyticsReport, error) {
	report := &models.AnalyticsReport{
		ID:          uuid.New().String(),
		ReportType:  req.ReportType,
		Title:       req.Title,
		Description: req.Description,
		Granularity: req.Granularity,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		Status:      "generating",
		GeneratedBy: generatedBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 保存报告记录
	if err := s.db.Create(report).Error; err != nil {
		return nil, err
	}

	// 异步生成报告数据
	go s.generateReportData(report)

	return report, nil
}

// generateReportData 生成报告数据（异步）
func (s *AnalyticsService) generateReportData(report *models.AnalyticsReport) {
	var reportData map[string]interface{}

	switch report.ReportType {
	case models.MetricTypeUser:
		reportData = s.generateUserReport(report.StartDate, report.EndDate, report.Granularity)
	case models.MetricTypeLetter:
		reportData = s.generateLetterReport(report.StartDate, report.EndDate, report.Granularity)
	case models.MetricTypeCourier:
		reportData = s.generateCourierReport(report.StartDate, report.EndDate, report.Granularity)
	case models.MetricTypeSystem:
		reportData = s.generateSystemReport(report.StartDate, report.EndDate, report.Granularity)
	default:
		reportData = map[string]interface{}{"error": "Unknown report type"}
	}

	// 将数据序列化为JSON
	dataJSON, err := json.Marshal(reportData)
	if err != nil {
		report.Status = "failed"
		report.Data = fmt.Sprintf(`{"error": "%s"}`, err.Error())
	} else {
		report.Status = "completed"
		report.Data = string(dataJSON)
	}

	report.UpdatedAt = time.Now()
	s.db.Save(report)
}

// generateUserReport 生成用户报告数据
func (s *AnalyticsService) generateUserReport(startDate, endDate time.Time, granularity models.AnalyticsGranularity) map[string]interface{} {
	var analytics []models.UserAnalytics
	s.db.Where("date BETWEEN ? AND ?", startDate, endDate).
		Find(&analytics)

	data := map[string]interface{}{
		"summary": map[string]interface{}{
			"totalUsers":           len(analytics),
			"avgEngagementScore":   s.calculateAvgEngagement(analytics),
			"totalLettersSent":     s.sumLettersSent(analytics),
			"totalLettersReceived": s.sumLettersReceived(analytics),
		},
		"timeline":   s.buildUserTimeline(analytics, granularity),
		"engagement": s.buildEngagementData(analytics),
	}

	return data
}

// generateLetterReport 生成信件报告数据
func (s *AnalyticsService) generateLetterReport(startDate, endDate time.Time, granularity models.AnalyticsGranularity) map[string]interface{} {
	var totalLetters, deliveredLetters int64

	s.db.Model(&models.Letter{}).
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Count(&totalLetters)

	s.db.Model(&models.Letter{}).
		Where("status = ? AND updated_at BETWEEN ? AND ?", models.StatusDelivered, startDate, endDate).
		Count(&deliveredLetters)

	deliveryRate := float64(0)
	if totalLetters > 0 {
		deliveryRate = float64(deliveredLetters) / float64(totalLetters) * 100
	}

	return map[string]interface{}{
		"summary": map[string]interface{}{
			"totalLetters":     totalLetters,
			"deliveredLetters": deliveredLetters,
			"deliveryRate":     deliveryRate,
		},
		"statusDistribution": s.getLetterStatusDistribution(startDate, endDate),
		"styleDistribution":  s.getLetterStyleDistribution(startDate, endDate),
	}
}

// generateCourierReport 生成信使报告数据
func (s *AnalyticsService) generateCourierReport(startDate, endDate time.Time, granularity models.AnalyticsGranularity) map[string]interface{} {
	var totalTasks, completedTasks int64

	s.db.Model(&models.CourierTask{}).
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Count(&totalTasks)

	s.db.Model(&models.CourierTask{}).
		Where("status = 'completed' AND updated_at BETWEEN ? AND ?", startDate, endDate).
		Count(&completedTasks)

	completionRate := float64(0)
	if totalTasks > 0 {
		completionRate = float64(completedTasks) / float64(totalTasks) * 100
	}

	return map[string]interface{}{
		"summary": map[string]interface{}{
			"totalTasks":     totalTasks,
			"completedTasks": completedTasks,
			"completionRate": completionRate,
		},
		"courierPerformance": s.getCourierPerformanceData(startDate, endDate),
	}
}

// generateSystemReport 生成系统报告数据
func (s *AnalyticsService) generateSystemReport(startDate, endDate time.Time, granularity models.AnalyticsGranularity) map[string]interface{} {
	var analytics []models.SystemAnalytics
	s.db.Where("date BETWEEN ? AND ?", startDate, endDate).
		Order("date ASC").
		Find(&analytics)

	return map[string]interface{}{
		"summary":     s.calculateSystemSummary(analytics),
		"timeline":    s.buildSystemTimeline(analytics, granularity),
		"performance": s.getSystemPerformanceData(startDate, endDate),
	}
}

// GetDashboardData 获取仪表板数据
func (s *AnalyticsService) GetDashboardData() (*models.DashboardData, error) {
	data := &models.DashboardData{}
	today := time.Now()

	// 概览数据
	var totalUsersCount int64
	s.db.Model(&models.User{}).Count(&totalUsersCount)
	data.Overview.TotalUsers = int(totalUsersCount)

	var activeUsersCount int64
	s.db.Model(&models.UserAnalytics{}).
		Where("date = ?", today.Format("2006-01-02")).
		Where("letters_sent > 0 OR letters_received > 0 OR courier_tasks > 0").
		Count(&activeUsersCount)
	data.Overview.ActiveUsers = int(activeUsersCount)

	var totalLettersCount int64
	s.db.Model(&models.Letter{}).Count(&totalLettersCount)
	data.Overview.TotalLetters = int(totalLettersCount)

	var lettersTodayCount int64
	s.db.Model(&models.Letter{}).
		Where("DATE(created_at) = ?", today.Format("2006-01-02")).
		Count(&lettersTodayCount)
	data.Overview.LettersToday = int(lettersTodayCount)

	// 参与度和送达率
	data.Overview.EngagementRate = s.calculateOverallEngagementRate()
	data.Overview.DeliveryRate = s.calculateOverallDeliveryRate()

	// 趋势数据
	data.TrendData.UserGrowth = s.getUserGrowthTrend(30)
	data.TrendData.LetterActivity = s.getLetterActivityTrend(30)
	data.TrendData.CourierActivity = s.getCourierActivityTrend(30)
	data.TrendData.SystemHealth = s.getSystemHealthTrend(30)

	// 最近活动
	data.RecentActivity = s.getRecentActivity(20)

	// 警报
	data.Alerts = s.getSystemAlerts()

	return data, nil
}

// 辅助方法实现...

func (s *AnalyticsService) calculateAvgEngagement(analytics []models.UserAnalytics) float64 {
	if len(analytics) == 0 {
		return 0
	}

	total := 0.0
	for _, a := range analytics {
		total += a.EngagementScore
	}
	return total / float64(len(analytics))
}

func (s *AnalyticsService) sumLettersSent(analytics []models.UserAnalytics) int {
	total := 0
	for _, a := range analytics {
		total += a.LettersSent
	}
	return total
}

func (s *AnalyticsService) sumLettersReceived(analytics []models.UserAnalytics) int {
	total := 0
	for _, a := range analytics {
		total += a.LettersReceived
	}
	return total
}

func (s *AnalyticsService) buildUserTimeline(analytics []models.UserAnalytics, granularity models.AnalyticsGranularity) []models.DataPoint {
	// 简化实现，实际应根据granularity聚合数据
	points := make([]models.DataPoint, 0, len(analytics))
	for _, a := range analytics {
		points = append(points, models.DataPoint{
			Timestamp: a.Date,
			Value:     a.EngagementScore,
			Label:     a.Date.Format("2006-01-02"),
		})
	}
	return points
}

func (s *AnalyticsService) buildEngagementData(analytics []models.UserAnalytics) map[string]interface{} {
	return map[string]interface{}{
		"averageScore": s.calculateAvgEngagement(analytics),
		"distribution": "TODO: Implement engagement distribution",
	}
}

func (s *AnalyticsService) getLetterStatusDistribution(startDate, endDate time.Time) map[string]int {
	var results []struct {
		Status string
		Count  int
	}

	s.db.Model(&models.Letter{}).
		Select("status, count(*) as count").
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Group("status").
		Scan(&results)

	distribution := make(map[string]int)
	for _, r := range results {
		distribution[r.Status] = r.Count
	}

	return distribution
}

func (s *AnalyticsService) getLetterStyleDistribution(startDate, endDate time.Time) map[string]int {
	var results []struct {
		Style string
		Count int
	}

	s.db.Model(&models.Letter{}).
		Select("style, count(*) as count").
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Group("style").
		Scan(&results)

	distribution := make(map[string]int)
	for _, r := range results {
		distribution[r.Style] = r.Count
	}

	return distribution
}

func (s *AnalyticsService) getCourierPerformanceData(startDate, endDate time.Time) []map[string]interface{} {
	return []map[string]interface{}{
		{"courier_id": "example", "tasks_completed": 10, "avg_completion_time": 120},
	}
}

func (s *AnalyticsService) calculateSystemSummary(analytics []models.SystemAnalytics) map[string]interface{} {
	if len(analytics) == 0 {
		return map[string]interface{}{}
	}

	latest := analytics[len(analytics)-1]
	return map[string]interface{}{
		"totalUsers":      latest.TotalUsers,
		"activeUsers":     latest.ActiveUsers,
		"avgResponseTime": latest.AvgResponseTime,
		"errorRate":       latest.ErrorRate,
		"serverUptime":    latest.ServerUptime,
	}
}

func (s *AnalyticsService) buildSystemTimeline(analytics []models.SystemAnalytics, granularity models.AnalyticsGranularity) []models.DataPoint {
	points := make([]models.DataPoint, 0, len(analytics))
	for _, a := range analytics {
		points = append(points, models.DataPoint{
			Timestamp: a.Date,
			Value:     float64(a.ActiveUsers),
			Label:     a.Date.Format("2006-01-02"),
		})
	}
	return points
}

func (s *AnalyticsService) getSystemPerformanceData(startDate, endDate time.Time) map[string]interface{} {
	var avgResponseTime float64
	s.db.Model(&models.PerformanceMetric{}).
		Where("timestamp BETWEEN ? AND ?", startDate, endDate).
		Select("AVG(response_time)").
		Scan(&avgResponseTime)

	return map[string]interface{}{
		"avgResponseTime": avgResponseTime,
	}
}

func (s *AnalyticsService) calculateOverallEngagementRate() float64 {
	// 计算过去7天的平均参与度
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -7)

	var analytics []models.UserAnalytics
	s.db.Where("date BETWEEN ? AND ?", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")).
		Find(&analytics)

	if len(analytics) == 0 {
		return 0.0
	}

	// 计算总活跃用户数和总用户数
	var activeUsers int
	var totalPossibleUsers int64
	
	// 获取7天前的总用户数
	s.db.Model(&models.User{}).
		Where("created_at <= ?", startDate).
		Count(&totalPossibleUsers)
	
	if totalPossibleUsers == 0 {
		return 0.0
	}

	// 统计有活动的用户
	activeUserMap := make(map[string]bool)
	for _, a := range analytics {
		if a.LettersSent > 0 || a.LettersReceived > 0 || a.CourierTasks > 0 {
			activeUserMap[a.UserID] = true
		}
	}
	activeUsers = len(activeUserMap)

	// 计算参与度百分比
	engagementRate := float64(activeUsers) / float64(totalPossibleUsers) * 100
	
	// 保留一位小数
	return math.Round(engagementRate*10) / 10
}

func (s *AnalyticsService) calculateOverallDeliveryRate() float64 {
	var total, delivered int64
	s.db.Model(&models.Letter{}).Count(&total)
	s.db.Model(&models.Letter{}).Where("status = ?", models.StatusDelivered).Count(&delivered)

	if total == 0 {
		return 0
	}

	return float64(delivered) / float64(total) * 100
}

func (s *AnalyticsService) getUserGrowthTrend(days int) []models.DataPoint {
	points := make([]models.DataPoint, 0, days)
	for i := days; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i)
		var count int64
		s.db.Model(&models.User{}).Where("DATE(created_at) = ?", date.Format("2006-01-02")).Count(&count)

		points = append(points, models.DataPoint{
			Timestamp: date,
			Value:     float64(count),
			Label:     date.Format("01-02"),
		})
	}
	return points
}

func (s *AnalyticsService) getLetterActivityTrend(days int) []models.DataPoint {
	points := make([]models.DataPoint, 0, days)
	for i := days; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i)
		var count int64
		s.db.Model(&models.Letter{}).Where("DATE(created_at) = ?", date.Format("2006-01-02")).Count(&count)

		points = append(points, models.DataPoint{
			Timestamp: date,
			Value:     float64(count),
			Label:     date.Format("01-02"),
		})
	}
	return points
}

func (s *AnalyticsService) getCourierActivityTrend(days int) []models.DataPoint {
	points := make([]models.DataPoint, 0, days)
	for i := days; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i)
		var count int64
		s.db.Model(&models.CourierTask{}).Where("DATE(created_at) = ?", date.Format("2006-01-02")).Count(&count)

		points = append(points, models.DataPoint{
			Timestamp: date,
			Value:     float64(count),
			Label:     date.Format("01-02"),
		})
	}
	return points
}

func (s *AnalyticsService) getSystemHealthTrend(days int) []models.DataPoint {
	points := make([]models.DataPoint, 0, days)
	for i := days; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i)

		points = append(points, models.DataPoint{
			Timestamp: date,
			Value:     99.5, // 假设系统健康度
			Label:     date.Format("01-02"),
		})
	}
	return points
}

func (s *AnalyticsService) getRecentActivity(limit int) []models.ActivityItem {
	activities := make([]models.ActivityItem, 0, limit)

	// 获取最近的信件活动
	var letters []models.Letter
	s.db.Preload("User").Order("created_at DESC").Limit(limit / 2).Find(&letters)

	for _, letter := range letters {
		activities = append(activities, models.ActivityItem{
			Type:        "letter_created",
			Description: fmt.Sprintf("Created letter: %s", letter.Title),
			UserID:      letter.UserID,
			Timestamp:   letter.CreatedAt,
			Metadata: map[string]interface{}{
				"letter_id": letter.ID,
				"style":     letter.Style,
			},
		})
	}

	return activities
}

func (s *AnalyticsService) getSystemAlerts() []models.AlertItem {
	alerts := []models.AlertItem{}

	// 检查错误率
	errorRate := s.calculateOverallEngagementRate()
	if errorRate > 5.0 {
		alerts = append(alerts, models.AlertItem{
			Level:     "warning",
			Message:   fmt.Sprintf("High error rate: %.1f%%", errorRate),
			Type:      "performance",
			Timestamp: time.Now(),
		})
	}

	return alerts
}
