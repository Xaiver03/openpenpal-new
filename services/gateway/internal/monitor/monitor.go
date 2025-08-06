package monitor

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// HTTP请求计数器
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gateway_http_requests_total",
			Help: "Total number of HTTP requests processed by the gateway",
		},
		[]string{"method", "path", "status"},
	)

	// HTTP请求耗时直方图
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "gateway_http_request_duration_seconds",
			Help: "Duration of HTTP requests processed by the gateway",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 2, 5, 10},
		},
		[]string{"method", "path", "status"},
	)

	// 服务健康状态指标
	serviceHealthGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gateway_service_health",
			Help: "Health status of backend services (1=healthy, 0=unhealthy)",
		},
		[]string{"service", "instance"},
	)

	// 活跃连接数
	activeConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "gateway_active_connections",
			Help: "Number of active connections to the gateway",
		},
	)

	// 代理请求计数器
	proxyRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gateway_proxy_requests_total",
			Help: "Total number of proxy requests by service",
		},
		[]string{"service", "status"},
	)

	// 限流触发计数器
	rateLimitTriggered = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gateway_rate_limit_triggered_total",
			Help: "Total number of requests that triggered rate limiting",
		},
		[]string{"client_type"},
	)

	// 服务发现事件计数器
	serviceDiscoveryEvents = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gateway_service_discovery_events_total",
			Help: "Total number of service discovery events",
		},
		[]string{"service", "event_type"},
	)
)

// InitMetrics 初始化监控指标
func InitMetrics() {
	// 注册指标
	prometheus.MustRegister(
		httpRequestsTotal,
		httpRequestDuration,
		serviceHealthGauge,
		activeConnections,
		proxyRequestsTotal,
		rateLimitTriggered,
		serviceDiscoveryEvents,
	)
}

// InitLogger 初始化日志器
func InitLogger(level string) *zap.Logger {
	// 解析日志级别
	var zapLevel zapcore.Level
	switch level {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	// 配置日志器
	config := zap.Config{
		Level:             zap.NewAtomicLevelAt(zapLevel),
		Development:       false,
		DisableCaller:     false,
		DisableStacktrace: false,
		Sampling:          nil,
		Encoding:          "json",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "timestamp",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "message",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	logger, err := config.Build()
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}

	return logger
}

// MetricsHandler 返回Prometheus指标处理器
func MetricsHandler() http.Handler {
	return promhttp.Handler()
}

// RecordHTTPRequest 记录HTTP请求指标
func RecordHTTPRequest(method, path string, status int, duration time.Duration) {
	statusStr := http.StatusText(status)
	if statusStr == "" {
		statusStr = "unknown"
	}

	httpRequestsTotal.WithLabelValues(method, path, statusStr).Inc()
	httpRequestDuration.WithLabelValues(method, path, statusStr).Observe(duration.Seconds())
}

// RecordProxyRequest 记录代理请求指标
func RecordProxyRequest(service string, status int) {
	statusStr := "success"
	if status >= 400 {
		statusStr = "error"
	}
	
	proxyRequestsTotal.WithLabelValues(service, statusStr).Inc()
}

// UpdateServiceHealth 更新服务健康状态
func UpdateServiceHealth(service, instance string, healthy bool) {
	value := 0.0
	if healthy {
		value = 1.0
	}
	
	serviceHealthGauge.WithLabelValues(service, instance).Set(value)
}

// RecordRateLimitEvent 记录限流事件
func RecordRateLimitEvent(clientType string) {
	rateLimitTriggered.WithLabelValues(clientType).Inc()
}

// RecordServiceDiscoveryEvent 记录服务发现事件
func RecordServiceDiscoveryEvent(service, eventType string) {
	serviceDiscoveryEvents.WithLabelValues(service, eventType).Inc()
}

// UpdateActiveConnections 更新活跃连接数
func UpdateActiveConnections(count float64) {
	activeConnections.Set(count)
}

// GetMetricsSummary 获取指标摘要
func GetMetricsSummary() map[string]interface{} {
	summary := make(map[string]interface{})

	// 这里可以收集当前的指标值
	// 需要使用prometheus的Gatherer接口来获取当前值
	
	return summary
}

// LogLevel 日志级别
type LogLevel string

const (
	DebugLevel LogLevel = "debug"
	InfoLevel  LogLevel = "info"
	WarnLevel  LogLevel = "warn"
	ErrorLevel LogLevel = "error"
)

// HealthChecker 健康检查器接口
type HealthChecker interface {
	CheckHealth() (bool, error)
}

// ServiceMonitor 服务监控器
type ServiceMonitor struct {
	logger      *zap.Logger
	healthChecks map[string]HealthChecker
}

// NewServiceMonitor 创建服务监控器
func NewServiceMonitor(logger *zap.Logger) *ServiceMonitor {
	return &ServiceMonitor{
		logger:       logger,
		healthChecks: make(map[string]HealthChecker),
	}
}

// RegisterHealthCheck 注册健康检查
func (sm *ServiceMonitor) RegisterHealthCheck(name string, checker HealthChecker) {
	sm.healthChecks[name] = checker
}

// CheckAllServices 检查所有服务
func (sm *ServiceMonitor) CheckAllServices() map[string]bool {
	results := make(map[string]bool)
	
	for name, checker := range sm.healthChecks {
		healthy, err := checker.CheckHealth()
		if err != nil {
			sm.logger.Error("Health check failed",
				zap.String("service", name),
				zap.Error(err),
			)
			healthy = false
		}
		
		results[name] = healthy
		UpdateServiceHealth(name, "default", healthy)
	}
	
	return results
}