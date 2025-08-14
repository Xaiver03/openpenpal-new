package models

import (
	"time"
)

// PerformanceMetric 前端性能指标
type PerformanceMetric struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	SessionID string `json:"session_id" gorm:"not null;index"`
	UserID    string `json:"user_id" gorm:"index"`
	PageURL   string `json:"page_url" gorm:"not null"`

	// Core Web Vitals
	LCP  float64 `json:"lcp"`  // Largest Contentful Paint
	FID  float64 `json:"fid"`  // First Input Delay
	CLS  float64 `json:"cls"`  // Cumulative Layout Shift
	TTFB float64 `json:"ttfb"` // Time to First Byte

	// 额外性能指标
	LoadTime   float64 `json:"load_time"`    // 页面加载时间
	DOMReady   float64 `json:"dom_ready"`    // DOM加载完成时间
	FirstPaint float64 `json:"first_paint"`  // 首次绘制时间
	JSHeapSize int64   `json:"js_heap_size"` // JS堆内存使用

	// 网络信息
	ConnectionType string  `json:"connection_type"` // 网络类型
	DownloadSpeed  float64 `json:"download_speed"`  // 下载速度

	// 设备信息
	UserAgent  string `json:"user_agent"`
	DeviceType string `json:"device_type"` // mobile/desktop/tablet
	ScreenSize string `json:"screen_size"` // 1920x1080

	Timestamp time.Time `json:"timestamp" gorm:"autoCreateTime"`
	CreatedAt time.Time `json:"created_at"`
}

// PerformanceAlert 性能告警
type PerformanceAlert struct {
	ID         uint       `json:"id" gorm:"primaryKey"`
	MetricType string     `json:"metric_type" gorm:"not null"` // lcp/fid/cls/ttfb
	Threshold  float64    `json:"threshold"`                   // 告警阈值
	Value      float64    `json:"value"`                       // 实际值
	PageURL    string     `json:"page_url"`
	SessionID  string     `json:"session_id"`
	UserID     string     `json:"user_id"`
	Severity   string     `json:"severity"`                     // low/medium/high/critical
	Status     string     `json:"status" gorm:"default:active"` // active/resolved
	Message    string     `json:"message"`
	CreatedAt  time.Time  `json:"created_at"`
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`
}

// DashboardMetrics 仪表板聚合指标
type DashboardMetrics struct {
	TimeRange string `json:"time_range"` // 1h/24h/7d/30d

	// Core Web Vitals 平均值
	AvgLCP  float64 `json:"avg_lcp"`
	AvgFID  float64 `json:"avg_fid"`
	AvgCLS  float64 `json:"avg_cls"`
	AvgTTFB float64 `json:"avg_ttfb"`

	// 性能评分
	PerformanceScore int `json:"performance_score"` // 0-100

	// 页面分析
	TopSlowPages []PageMetrics `json:"top_slow_pages"`
	TopFastPages []PageMetrics `json:"top_fast_pages"`

	// 设备分析
	DeviceBreakdown map[string]int `json:"device_breakdown"`

	// 错误分析
	ErrorCount     int `json:"error_count"`
	CriticalAlerts int `json:"critical_alerts"`

	// 趋势数据
	TrendData []TrendPoint `json:"trend_data"`

	LastUpdated time.Time `json:"last_updated"`
}

// PageMetrics 页面性能指标
type PageMetrics struct {
	URL         string  `json:"url"`
	AvgLoadTime float64 `json:"avg_load_time"`
	AvgLCP      float64 `json:"avg_lcp"`
	VisitCount  int     `json:"visit_count"`
}

// TrendPoint 趋势数据点
type TrendPoint struct {
	Timestamp time.Time `json:"timestamp"`
	LCP       float64   `json:"lcp"`
	FID       float64   `json:"fid"`
	CLS       float64   `json:"cls"`
	TTFB      float64   `json:"ttfb"`
}

// PerformanceRequest 性能数据提交请求
type PerformanceRequest struct {
	SessionID      string  `json:"session_id" binding:"required"`
	PageURL        string  `json:"page_url" binding:"required"`
	LCP            float64 `json:"lcp"`
	FID            float64 `json:"fid"`
	CLS            float64 `json:"cls"`
	TTFB           float64 `json:"ttfb"`
	LoadTime       float64 `json:"load_time"`
	DOMReady       float64 `json:"dom_ready"`
	FirstPaint     float64 `json:"first_paint"`
	JSHeapSize     int64   `json:"js_heap_size"`
	ConnectionType string  `json:"connection_type"`
	DownloadSpeed  float64 `json:"download_speed"`
	DeviceType     string  `json:"device_type"`
	ScreenSize     string  `json:"screen_size"`
}

// AlertRequest 告警请求
type AlertRequest struct {
	MetricType string  `json:"metric_type" binding:"required"`
	Threshold  float64 `json:"threshold" binding:"required"`
	Value      float64 `json:"value" binding:"required"`
	PageURL    string  `json:"page_url" binding:"required"`
	SessionID  string  `json:"session_id"`
	Severity   string  `json:"severity"`
	Message    string  `json:"message"`
}
