package services

import (
	"api-gateway/internal/models"
	"fmt"
	"math"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// MetricsService 性能监控服务
type MetricsService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewMetricsService 创建性能监控服务
func NewMetricsService(db *gorm.DB, logger *zap.Logger) *MetricsService {
	return &MetricsService{
		db:     db,
		logger: logger,
	}
}

// SavePerformanceMetric 保存性能指标
func (s *MetricsService) SavePerformanceMetric(userID string, req *models.PerformanceRequest) error {
	metric := &models.PerformanceMetric{
		SessionID:      req.SessionID,
		UserID:         userID,
		PageURL:        req.PageURL,
		LCP:            req.LCP,
		FID:            req.FID,
		CLS:            req.CLS,
		TTFB:           req.TTFB,
		LoadTime:       req.LoadTime,
		DOMReady:       req.DOMReady,
		FirstPaint:     req.FirstPaint,
		JSHeapSize:     req.JSHeapSize,
		ConnectionType: req.ConnectionType,
		DownloadSpeed:  req.DownloadSpeed,
		DeviceType:     req.DeviceType,
		ScreenSize:     req.ScreenSize,
		Timestamp:      time.Now(),
	}

	if err := s.db.Create(metric).Error; err != nil {
		s.logger.Error("Failed to save performance metric", zap.Error(err))
		return err
	}

	// 检查是否需要创建告警
	s.checkAndCreateAlerts(metric)

	s.logger.Info("Performance metric saved successfully",
		zap.String("session_id", req.SessionID),
		zap.String("page_url", req.PageURL),
		zap.Float64("lcp", req.LCP),
	)

	return nil
}

// GetDashboardMetrics 获取仪表板指标
func (s *MetricsService) GetDashboardMetrics(timeRange string) (*models.DashboardMetrics, error) {
	var duration time.Duration

	switch timeRange {
	case "1h":
		duration = time.Hour
	case "24h":
		duration = 24 * time.Hour
	case "7d":
		duration = 7 * 24 * time.Hour
	case "30d":
		duration = 30 * 24 * time.Hour
	default:
		duration = 24 * time.Hour
		timeRange = "24h"
	}

	since := time.Now().Add(-duration)

	// 获取平均性能指标
	var avgMetrics struct {
		AvgLCP  float64
		AvgFID  float64
		AvgCLS  float64
		AvgTTFB float64
	}

	err := s.db.Model(&models.PerformanceMetric{}).
		Where("timestamp >= ?", since).
		Select("AVG(lcp) as avg_lcp, AVG(fid) as avg_fid, AVG(cls) as avg_cls, AVG(ttfb) as avg_ttfb").
		Scan(&avgMetrics).Error

	if err != nil {
		s.logger.Error("Failed to get average metrics", zap.Error(err))
		return nil, err
	}

	// 获取设备分布
	deviceBreakdown, err := s.getDeviceBreakdown(since)
	if err != nil {
		s.logger.Error("Failed to get device breakdown", zap.Error(err))
		return nil, err
	}

	// 获取最慢和最快页面
	topSlowPages, err := s.getTopSlowPages(since, 5)
	if err != nil {
		s.logger.Error("Failed to get slow pages", zap.Error(err))
		return nil, err
	}

	topFastPages, err := s.getTopFastPages(since, 5)
	if err != nil {
		s.logger.Error("Failed to get fast pages", zap.Error(err))
		return nil, err
	}

	// 获取告警统计
	errorCount, criticalAlerts, err := s.getAlertStats(since)
	if err != nil {
		s.logger.Error("Failed to get alert stats", zap.Error(err))
		return nil, err
	}

	// 获取趋势数据
	trendData, err := s.getTrendData(since, timeRange)
	if err != nil {
		s.logger.Error("Failed to get trend data", zap.Error(err))
		return nil, err
	}

	// 计算性能评分
	performanceScore := s.calculatePerformanceScore(avgMetrics.AvgLCP, avgMetrics.AvgFID, avgMetrics.AvgCLS, avgMetrics.AvgTTFB)

	dashboard := &models.DashboardMetrics{
		TimeRange:        timeRange,
		AvgLCP:           avgMetrics.AvgLCP,
		AvgFID:           avgMetrics.AvgFID,
		AvgCLS:           avgMetrics.AvgCLS,
		AvgTTFB:          avgMetrics.AvgTTFB,
		PerformanceScore: performanceScore,
		TopSlowPages:     topSlowPages,
		TopFastPages:     topFastPages,
		DeviceBreakdown:  deviceBreakdown,
		ErrorCount:       errorCount,
		CriticalAlerts:   criticalAlerts,
		TrendData:        trendData,
		LastUpdated:      time.Now(),
	}

	return dashboard, nil
}

// GetActiveAlerts 获取活跃告警
func (s *MetricsService) GetActiveAlerts(limit int) ([]models.PerformanceAlert, error) {
	var alerts []models.PerformanceAlert

	query := s.db.Where("status = ?", "active").
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&alerts).Error
	if err != nil {
		s.logger.Error("Failed to get active alerts", zap.Error(err))
		return nil, err
	}

	return alerts, nil
}

// CreateAlert 创建性能告警
func (s *MetricsService) CreateAlert(userID string, req *models.AlertRequest) error {
	alert := &models.PerformanceAlert{
		MetricType: req.MetricType,
		Threshold:  req.Threshold,
		Value:      req.Value,
		PageURL:    req.PageURL,
		SessionID:  req.SessionID,
		UserID:     userID,
		Severity:   req.Severity,
		Message:    req.Message,
		Status:     "active",
		CreatedAt:  time.Now(),
	}

	if err := s.db.Create(alert).Error; err != nil {
		s.logger.Error("Failed to create alert", zap.Error(err))
		return err
	}

	s.logger.Warn("Performance alert created",
		zap.String("metric_type", req.MetricType),
		zap.Float64("threshold", req.Threshold),
		zap.Float64("value", req.Value),
		zap.String("severity", req.Severity),
	)

	return nil
}

// checkAndCreateAlerts 检查并创建告警
func (s *MetricsService) checkAndCreateAlerts(metric *models.PerformanceMetric) {
	alerts := []struct {
		metricType string
		value      float64
		threshold  float64
		severity   string
	}{
		// Core Web Vitals 阈值基于Google标准
		{"lcp", metric.LCP, 2500, "medium"},  // LCP > 2.5s
		{"lcp", metric.LCP, 4000, "high"},    // LCP > 4s
		{"fid", metric.FID, 100, "medium"},   // FID > 100ms
		{"fid", metric.FID, 300, "high"},     // FID > 300ms
		{"cls", metric.CLS, 0.1, "medium"},   // CLS > 0.1
		{"cls", metric.CLS, 0.25, "high"},    // CLS > 0.25
		{"ttfb", metric.TTFB, 800, "medium"}, // TTFB > 800ms
		{"ttfb", metric.TTFB, 1800, "high"},  // TTFB > 1.8s
	}

	for _, alert := range alerts {
		if alert.value > alert.threshold {
			alertReq := &models.AlertRequest{
				MetricType: alert.metricType,
				Threshold:  alert.threshold,
				Value:      alert.value,
				PageURL:    metric.PageURL,
				SessionID:  metric.SessionID,
				Severity:   alert.severity,
				Message:    fmt.Sprintf("%s exceeded threshold: %.2f > %.2f", alert.metricType, alert.value, alert.threshold),
			}

			s.CreateAlert(metric.UserID, alertReq)
		}
	}
}

// getDeviceBreakdown 获取设备分布
func (s *MetricsService) getDeviceBreakdown(since time.Time) (map[string]int, error) {
	var results []struct {
		DeviceType string
		Count      int
	}

	err := s.db.Model(&models.PerformanceMetric{}).
		Where("timestamp >= ?", since).
		Select("device_type, COUNT(*) as count").
		Group("device_type").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	breakdown := make(map[string]int)
	for _, result := range results {
		breakdown[result.DeviceType] = result.Count
	}

	return breakdown, nil
}

// getTopSlowPages 获取最慢页面
func (s *MetricsService) getTopSlowPages(since time.Time, limit int) ([]models.PageMetrics, error) {
	var results []models.PageMetrics

	err := s.db.Model(&models.PerformanceMetric{}).
		Where("timestamp >= ?", since).
		Select("page_url as url, AVG(load_time) as avg_load_time, AVG(lcp) as avg_lcp, COUNT(*) as visit_count").
		Group("page_url").
		Order("avg_load_time DESC").
		Limit(limit).
		Scan(&results).Error

	return results, err
}

// getTopFastPages 获取最快页面
func (s *MetricsService) getTopFastPages(since time.Time, limit int) ([]models.PageMetrics, error) {
	var results []models.PageMetrics

	err := s.db.Model(&models.PerformanceMetric{}).
		Where("timestamp >= ?", since).
		Select("page_url as url, AVG(load_time) as avg_load_time, AVG(lcp) as avg_lcp, COUNT(*) as visit_count").
		Group("page_url").
		Order("avg_load_time ASC").
		Limit(limit).
		Scan(&results).Error

	return results, err
}

// getAlertStats 获取告警统计
func (s *MetricsService) getAlertStats(since time.Time) (int, int, error) {
	var errorCount, criticalAlerts int64

	// 获取总告警数
	err := s.db.Model(&models.PerformanceAlert{}).
		Where("created_at >= ?", since).
		Count(&errorCount).Error

	if err != nil {
		return 0, 0, err
	}

	// 获取严重告警数
	err = s.db.Model(&models.PerformanceAlert{}).
		Where("created_at >= ? AND severity IN (?)", since, []string{"high", "critical"}).
		Count(&criticalAlerts).Error

	return int(errorCount), int(criticalAlerts), err
}

// getTrendData 获取趋势数据
func (s *MetricsService) getTrendData(since time.Time, timeRange string) ([]models.TrendPoint, error) {
	var interval string

	switch timeRange {
	case "1h":
		interval = "5 minute"
	case "24h":
		interval = "1 hour"
	case "7d":
		interval = "6 hour"
	case "30d":
		interval = "1 day"
	default:
		interval = "1 hour"
	}

	var results []models.TrendPoint

	query := fmt.Sprintf(`
		SELECT 
			date_trunc('%s', timestamp) as timestamp,
			AVG(lcp) as lcp,
			AVG(fid) as fid,
			AVG(cls) as cls,
			AVG(ttfb) as ttfb
		FROM performance_metrics 
		WHERE timestamp >= $1 
		GROUP BY date_trunc('%s', timestamp)
		ORDER BY timestamp ASC
	`, interval, interval)

	err := s.db.Raw(query, since).Scan(&results).Error

	return results, err
}

// calculatePerformanceScore 计算性能评分
func (s *MetricsService) calculatePerformanceScore(lcp, fid, cls, ttfb float64) int {
	// 基于Google Core Web Vitals标准计算评分
	score := 100.0

	// LCP评分 (权重: 30%)
	if lcp > 4000 {
		score -= 30
	} else if lcp > 2500 {
		score -= 15
	}

	// FID评分 (权重: 30%)
	if fid > 300 {
		score -= 30
	} else if fid > 100 {
		score -= 15
	}

	// CLS评分 (权重: 30%)
	if cls > 0.25 {
		score -= 30
	} else if cls > 0.1 {
		score -= 15
	}

	// TTFB评分 (权重: 10%)
	if ttfb > 1800 {
		score -= 10
	} else if ttfb > 800 {
		score -= 5
	}

	return int(math.Max(0, score))
}
