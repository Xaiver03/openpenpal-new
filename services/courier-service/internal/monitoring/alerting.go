package monitoring

import (
	"context"
	"courier-service/internal/errors"
	"courier-service/internal/logging"
	"fmt"
	"sync"
	"time"
)

// AlertLevel 告警级别
type AlertLevel string

const (
	AlertLevelInfo     AlertLevel = "INFO"
	AlertLevelWarning  AlertLevel = "WARNING"
	AlertLevelCritical AlertLevel = "CRITICAL"
	AlertLevelFatal    AlertLevel = "FATAL"
)

// AlertRule 告警规则
type AlertRule struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Level       AlertLevel        `json:"level"`
	Condition   func() bool       `json:"-"`
	Threshold   float64           `json:"threshold"`
	Duration    time.Duration     `json:"duration"`
	Cooldown    time.Duration     `json:"cooldown"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	Enabled     bool              `json:"enabled"`
}

// Alert 告警
type Alert struct {
	ID          string            `json:"id"`
	Rule        *AlertRule        `json:"rule"`
	Level       AlertLevel        `json:"level"`
	Status      AlertStatus       `json:"status"`
	Message     string            `json:"message"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	StartTime   time.Time         `json:"start_time"`
	EndTime     *time.Time        `json:"end_time,omitempty"`
	Duration    time.Duration     `json:"duration"`
	Count       int               `json:"count"`
	LastUpdate  time.Time         `json:"last_update"`
}

// AlertStatus 告警状态
type AlertStatus string

const (
	AlertStatusPending    AlertStatus = "PENDING"
	AlertStatusFiring     AlertStatus = "FIRING"
	AlertStatusResolved   AlertStatus = "RESOLVED"
	AlertStatusSuppressed AlertStatus = "SUPPRESSED"
)

// AlertHandler 告警处理器接口
type AlertHandler interface {
	Handle(ctx context.Context, alert *Alert) error
	CanHandle(alert *Alert) bool
	Priority() int
}

// LogAlertHandler 日志告警处理器
type LogAlertHandler struct {
	logger logging.Logger
}

// NewLogAlertHandler 创建日志告警处理器
func NewLogAlertHandler(logger logging.Logger) *LogAlertHandler {
	return &LogAlertHandler{logger: logger}
}

func (h *LogAlertHandler) Handle(_ context.Context, alert *Alert) error {
	switch alert.Level {
	case AlertLevelInfo:
		h.logger.Info("Alert fired", "alert", alert)
	case AlertLevelWarning:
		h.logger.Warn("Alert fired", "alert", alert)
	case AlertLevelCritical, AlertLevelFatal:
		h.logger.Error("Alert fired", "alert", alert)
	}
	return nil
}

func (h *LogAlertHandler) CanHandle(_ *Alert) bool {
	return true
}

func (h *LogAlertHandler) Priority() int {
	return 1000 // 低优先级
}

// WebhookAlertHandler Webhook告警处理器
type WebhookAlertHandler struct {
	URL    string
	client interface{} // HTTP客户端接口
}

func (h *WebhookAlertHandler) Handle(_ context.Context, _ *Alert) error {
	// 这里实现Webhook调用逻辑
	// 发送HTTP POST请求到指定URL
	return nil
}

func (h *WebhookAlertHandler) CanHandle(alert *Alert) bool {
	return h.URL != ""
}

func (h *WebhookAlertHandler) Priority() int {
	return 500 // 中等优先级
}

// AlertManager 告警管理器
type AlertManager struct {
	rules        map[string]*AlertRule
	activeAlerts map[string]*Alert
	handlers     []AlertHandler
	registry     *MetricsRegistry
	logger       logging.Logger
	mutex        sync.RWMutex
	ticker       *time.Ticker
	ctx          context.Context
	cancel       context.CancelFunc
}

// NewAlertManager 创建告警管理器
func NewAlertManager(registry *MetricsRegistry, logger logging.Logger) *AlertManager {
	if logger == nil {
		logger = logging.GetDefaultLogger()
	}

	ctx, cancel := context.WithCancel(context.Background())

	am := &AlertManager{
		rules:        make(map[string]*AlertRule),
		activeAlerts: make(map[string]*Alert),
		handlers:     make([]AlertHandler, 0),
		registry:     registry,
		logger:       logger,
		ctx:          ctx,
		cancel:       cancel,
	}

	// 注册默认处理器
	am.RegisterHandler(NewLogAlertHandler(logger))

	return am
}

// RegisterRule 注册告警规则
func (am *AlertManager) RegisterRule(rule *AlertRule) {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	am.rules[rule.Name] = rule
	am.logger.Info("Alert rule registered", "rule_name", rule.Name, "level", rule.Level)
}

// RegisterHandler 注册告警处理器
func (am *AlertManager) RegisterHandler(handler AlertHandler) {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	am.handlers = append(am.handlers, handler)

	// 按优先级排序（优先级高的在前）
	for i := len(am.handlers) - 1; i > 0; i-- {
		if am.handlers[i].Priority() < am.handlers[i-1].Priority() {
			am.handlers[i], am.handlers[i-1] = am.handlers[i-1], am.handlers[i]
		}
	}
}

// Start 启动告警管理器
func (am *AlertManager) Start(interval time.Duration) {
	am.mutex.Lock()
	if am.ticker != nil {
		am.mutex.Unlock()
		return // 已经启动
	}

	am.ticker = time.NewTicker(interval)
	am.mutex.Unlock()

	go am.run()
	am.logger.Info("Alert manager started", "check_interval", interval)
}

// Stop 停止告警管理器
func (am *AlertManager) Stop() {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	if am.ticker != nil {
		am.ticker.Stop()
		am.ticker = nil
	}

	am.cancel()
	am.logger.Info("Alert manager stopped")
}

// run 运行告警检查循环
func (am *AlertManager) run() {
	for {
		select {
		case <-am.ctx.Done():
			return
		case <-am.ticker.C:
			am.checkRules()
		}
	}
}

// checkRules 检查告警规则
func (am *AlertManager) checkRules() {
	am.mutex.RLock()
	rules := make(map[string]*AlertRule)
	for k, v := range am.rules {
		if v.Enabled {
			rules[k] = v
		}
	}
	am.mutex.RUnlock()

	for _, rule := range rules {
		am.evaluateRule(rule)
	}
}

// evaluateRule 评估告警规则
func (am *AlertManager) evaluateRule(rule *AlertRule) {
	if !rule.Condition() {
		// 条件不满足，检查是否需要解决告警
		am.resolveAlert(rule.Name)
		return
	}

	am.mutex.Lock()
	defer am.mutex.Unlock()

	alertID := rule.Name
	now := time.Now()

	if alert, exists := am.activeAlerts[alertID]; exists {
		// 告警已存在，更新信息
		alert.Count++
		alert.LastUpdate = now
		alert.Duration = now.Sub(alert.StartTime)

		// 如果状态是PENDING且超过了持续时间，转为FIRING
		if alert.Status == AlertStatusPending && alert.Duration >= rule.Duration {
			alert.Status = AlertStatusFiring
			am.fireAlert(alert)
		}
	} else {
		// 创建新告警
		alert := &Alert{
			ID:          alertID,
			Rule:        rule,
			Level:       rule.Level,
			Status:      AlertStatusPending,
			Message:     am.generateAlertMessage(rule),
			Labels:      rule.Labels,
			Annotations: rule.Annotations,
			StartTime:   now,
			Count:       1,
			LastUpdate:  now,
		}

		am.activeAlerts[alertID] = alert

		// 如果持续时间为0，直接触发
		if rule.Duration == 0 {
			alert.Status = AlertStatusFiring
			am.fireAlert(alert)
		}
	}
}

// resolveAlert 解决告警
func (am *AlertManager) resolveAlert(ruleNme string) {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	if alert, exists := am.activeAlerts[ruleNme]; exists {
		now := time.Now()
		alert.Status = AlertStatusResolved
		alert.EndTime = &now
		alert.Duration = now.Sub(alert.StartTime)
		alert.LastUpdate = now

		am.logger.Info("Alert resolved",
			"alert_id", alert.ID,
			"duration", alert.Duration,
			"count", alert.Count,
		)

		// 通知处理器
		go am.notifyHandlers(alert)

		// 从活跃告警中移除
		delete(am.activeAlerts, ruleNme)
	}
}

// fireAlert 触发告警
func (am *AlertManager) fireAlert(alert *Alert) {
	am.logger.Warn("Alert fired",
		"alert_id", alert.ID,
		"level", alert.Level,
		"message", alert.Message,
		"count", alert.Count,
	)

	// 异步通知处理器
	go am.notifyHandlers(alert)
}

// notifyHandlers 通知处理器
func (am *AlertManager) notifyHandlers(alert *Alert) {
	for _, handler := range am.handlers {
		if handler.CanHandle(alert) {
			if err := handler.Handle(am.ctx, alert); err != nil {
				am.logger.Error("Alert handler failed",
					"alert_id", alert.ID,
					"handler", fmt.Sprintf("%T", handler),
					"error", err,
				)
			}
		}
	}
}

// generateAlertMessage 生成告警消息
func (am *AlertManager) generateAlertMessage(rule *AlertRule) string {
	return fmt.Sprintf("[%s] %s", rule.Level, rule.Description)
}

// GetActiveAlerts 获取活跃告警
func (am *AlertManager) GetActiveAlerts() map[string]*Alert {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	result := make(map[string]*Alert)
	for k, v := range am.activeAlerts {
		result[k] = v
	}

	return result
}

// GetAlertHistory 获取告警历史（这里简化实现）
func (am *AlertManager) GetAlertHistory(_ int) []*Alert {
	// 在实际实现中，应该从持久化存储中获取历史告警
	return []*Alert{}
}

// 预定义告警规则

// CreateErrorRateRule 创建错误率告警规则
func CreateErrorRateRule(registry *MetricsRegistry) *AlertRule {
	return &AlertRule{
		Name:        "high_error_rate",
		Description: "Error rate is too high",
		Level:       AlertLevelWarning,
		Condition: func() bool {
			// 获取错误计数指标
			errorCounter := registry.GetCounter("courier_service_errors_total", nil)
			totalRequests := registry.GetCounter("courier_service_requests_total", nil)

			errorCount := errorCounter.Value().(int64)
			requestCount := totalRequests.Value().(int64)

			if requestCount > 100 { // 最小请求数阈值
				errorRate := float64(errorCount) / float64(requestCount)
				return errorRate > 0.05 // 5%错误率阈值
			}

			return false
		},
		Duration: 5 * time.Minute,
		Cooldown: 10 * time.Minute,
		Enabled:  true,
	}
}

// CreateResponseTimeRule 创建响应时间告警规则
func CreateResponseTimeRule(registry *MetricsRegistry) *AlertRule {
	return &AlertRule{
		Name:        "high_response_time",
		Description: "Response time is too high",
		Level:       AlertLevelWarning,
		Condition: func() bool {
			histogram := registry.GetHistogram(
				"courier_service_response_time_seconds",
				[]float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
				nil,
			)

			value := histogram.Value().(map[string]interface{})
			total := value["total"].(int64)
			sum := value["sum"].(float64)

			if total > 0 {
				avgResponseTime := sum / float64(total)
				return avgResponseTime > 2.0 // 2秒阈值
			}

			return false
		},
		Duration: 3 * time.Minute,
		Cooldown: 5 * time.Minute,
		Enabled:  true,
	}
}

// CreateDatabaseConnectionRule 创建数据库连接告警规则
func CreateDatabaseConnectionRule(registry *MetricsRegistry) *AlertRule {
	return &AlertRule{
		Name:        "database_connection_failure",
		Description: "Database connection failures detected",
		Level:       AlertLevelCritical,
		Condition: func() bool {
			// 检查数据库连接错误
			dbErrorCounter := registry.GetCounter("courier_service_errors_by_code",
				map[string]string{"error_code": string(errors.CodeDatabaseError)})

			// 检查最近是否有数据库错误
			return dbErrorCounter.Value().(int64) > 0
		},
		Duration: 1 * time.Minute,
		Cooldown: 15 * time.Minute,
		Enabled:  true,
	}
}

// CreateCircuitBreakerRule 创建熔断器告警规则
func CreateCircuitBreakerRule() *AlertRule {
	return &AlertRule{
		Name:        "circuit_breaker_open",
		Description: "Circuit breaker is open",
		Level:       AlertLevelCritical,
		Condition: func() bool {
			// 检查熔断器状态
			// 这里需要实际的熔断器状态检查逻辑
			// 暂时返回false
			return false
		},
		Duration: 0, // 立即触发
		Cooldown: 5 * time.Minute,
		Enabled:  true,
	}
}

// 全局告警管理器
var globalAlertManager *AlertManager
var alertManagerOnce sync.Once

// InitGlobalAlertManager 初始化全局告警管理器
func InitGlobalAlertManager(registry *MetricsRegistry, logger logging.Logger) {
	alertManagerOnce.Do(func() {
		globalAlertManager = NewAlertManager(registry, logger)

		// 注册预定义规则
		globalAlertManager.RegisterRule(CreateErrorRateRule(registry))
		globalAlertManager.RegisterRule(CreateResponseTimeRule(registry))
		globalAlertManager.RegisterRule(CreateDatabaseConnectionRule(registry))
		globalAlertManager.RegisterRule(CreateCircuitBreakerRule())

		// 启动告警管理器
		globalAlertManager.Start(30 * time.Second)
	})
}

// GetGlobalAlertManager 获取全局告警管理器
func GetGlobalAlertManager() *AlertManager {
	if globalAlertManager == nil {
		InitGlobalAlertManager(GetGlobalRegistry(), nil)
	}
	return globalAlertManager
}
