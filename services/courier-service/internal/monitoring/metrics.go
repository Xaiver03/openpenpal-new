package monitoring

import (
	"courier-service/internal/errors"
	"courier-service/internal/logging"
	"sync"
	"time"
)

// MetricType 指标类型
type MetricType string

const (
	CounterType   MetricType = "counter"
	GaugeType     MetricType = "gauge"
	HistogramType MetricType = "histogram"
	SummaryType   MetricType = "summary"
)

// Metric 指标接口
type Metric interface {
	Name() string
	Type() MetricType
	Labels() map[string]string
	Value() interface{}
	Timestamp() time.Time
}

// Counter 计数器指标
type Counter struct {
	name      string
	value     int64
	labels    map[string]string
	mutex     sync.RWMutex
	timestamp time.Time
}

// NewCounter 创建计数器
func NewCounter(name string, labels map[string]string) *Counter {
	return &Counter{
		name:      name,
		value:     0,
		labels:    labels,
		timestamp: time.Now(),
	}
}

func (c *Counter) Name() string              { return c.name }
func (c *Counter) Type() MetricType          { return CounterType }
func (c *Counter) Labels() map[string]string { return c.labels }
func (c *Counter) Timestamp() time.Time      { return c.timestamp }

func (c *Counter) Value() interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.value
}

func (c *Counter) Inc() {
	c.Add(1)
}

func (c *Counter) Add(value int64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.value += value
	c.timestamp = time.Now()
}

// Gauge 仪表指标
type Gauge struct {
	name      string
	value     float64
	labels    map[string]string
	mutex     sync.RWMutex
	timestamp time.Time
}

// NewGauge 创建仪表
func NewGauge(name string, labels map[string]string) *Gauge {
	return &Gauge{
		name:      name,
		value:     0,
		labels:    labels,
		timestamp: time.Now(),
	}
}

func (g *Gauge) Name() string              { return g.name }
func (g *Gauge) Type() MetricType          { return GaugeType }
func (g *Gauge) Labels() map[string]string { return g.labels }
func (g *Gauge) Timestamp() time.Time      { return g.timestamp }

func (g *Gauge) Value() interface{} {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	return g.value
}

func (g *Gauge) Set(value float64) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.value = value
	g.timestamp = time.Now()
}

func (g *Gauge) Inc() {
	g.Add(1)
}

func (g *Gauge) Dec() {
	g.Add(-1)
}

func (g *Gauge) Add(value float64) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.value += value
	g.timestamp = time.Now()
}

// Histogram 直方图指标
type Histogram struct {
	name      string
	buckets   []float64
	counts    []int64
	sum       float64
	total     int64
	labels    map[string]string
	mutex     sync.RWMutex
	timestamp time.Time
}

// NewHistogram 创建直方图
func NewHistogram(name string, buckets []float64, labels map[string]string) *Histogram {
	return &Histogram{
		name:      name,
		buckets:   buckets,
		counts:    make([]int64, len(buckets)+1),
		labels:    labels,
		timestamp: time.Now(),
	}
}

func (h *Histogram) Name() string              { return h.name }
func (h *Histogram) Type() MetricType          { return HistogramType }
func (h *Histogram) Labels() map[string]string { return h.labels }
func (h *Histogram) Timestamp() time.Time      { return h.timestamp }

func (h *Histogram) Value() interface{} {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	return map[string]interface{}{
		"buckets": h.buckets,
		"counts":  h.counts,
		"sum":     h.sum,
		"total":   h.total,
	}
}

func (h *Histogram) Observe(value float64) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.sum += value
	h.total++
	h.timestamp = time.Now()

	// 找到对应的桶
	for i, bucket := range h.buckets {
		if value <= bucket {
			h.counts[i]++
			break
		}
	}
	// 如果超过所有桶，记录在最后一个桶
	if value > h.buckets[len(h.buckets)-1] {
		h.counts[len(h.buckets)]++
	}
}

// MetricsRegistry 指标注册表
type MetricsRegistry struct {
	metrics map[string]Metric
	mutex   sync.RWMutex
	logger  logging.Logger
}

// NewMetricsRegistry 创建指标注册表
func NewMetricsRegistry(logger logging.Logger) *MetricsRegistry {
	if logger == nil {
		logger = logging.GetDefaultLogger()
	}

	return &MetricsRegistry{
		metrics: make(map[string]Metric),
		logger:  logger,
	}
}

// Register 注册指标
func (mr *MetricsRegistry) Register(metric Metric) {
	mr.mutex.Lock()
	defer mr.mutex.Unlock()

	key := mr.generateKey(metric.Name(), metric.Labels())
	mr.metrics[key] = metric
}

// GetCounter 获取或创建计数器
func (mr *MetricsRegistry) GetCounter(name string, labels map[string]string) *Counter {
	key := mr.generateKey(name, labels)

	mr.mutex.RLock()
	if metric, exists := mr.metrics[key]; exists {
		mr.mutex.RUnlock()
		if counter, ok := metric.(*Counter); ok {
			return counter
		}
	}
	mr.mutex.RUnlock()

	counter := NewCounter(name, labels)
	mr.Register(counter)
	return counter
}

// GetGauge 获取或创建仪表
func (mr *MetricsRegistry) GetGauge(name string, labels map[string]string) *Gauge {
	key := mr.generateKey(name, labels)

	mr.mutex.RLock()
	if metric, exists := mr.metrics[key]; exists {
		mr.mutex.RUnlock()
		if gauge, ok := metric.(*Gauge); ok {
			return gauge
		}
	}
	mr.mutex.RUnlock()

	gauge := NewGauge(name, labels)
	mr.Register(gauge)
	return gauge
}

// GetHistogram 获取或创建直方图
func (mr *MetricsRegistry) GetHistogram(name string, buckets []float64, labels map[string]string) *Histogram {
	key := mr.generateKey(name, labels)

	mr.mutex.RLock()
	if metric, exists := mr.metrics[key]; exists {
		mr.mutex.RUnlock()
		if histogram, ok := metric.(*Histogram); ok {
			return histogram
		}
	}
	mr.mutex.RUnlock()

	histogram := NewHistogram(name, buckets, labels)
	mr.Register(histogram)
	return histogram
}

// GetAllMetrics 获取所有指标
func (mr *MetricsRegistry) GetAllMetrics() map[string]Metric {
	mr.mutex.RLock()
	defer mr.mutex.RUnlock()

	result := make(map[string]Metric)
	for k, v := range mr.metrics {
		result[k] = v
	}

	return result
}

// generateKey 生成指标键
func (mr *MetricsRegistry) generateKey(name string, labels map[string]string) string {
	key := name
	for k, v := range labels {
		key += ":" + k + "=" + v
	}
	return key
}

// ErrorMetrics 错误指标收集器
type ErrorMetrics struct {
	registry     *MetricsRegistry
	errorCounter *Counter
	errorByCode  map[errors.ErrorCode]*Counter
	errorByType  map[errors.ErrorType]*Counter
	responseTime *Histogram
	logger       logging.Logger
	mutex        sync.RWMutex
}

// NewErrorMetrics 创建错误指标收集器
func NewErrorMetrics(registry *MetricsRegistry, logger logging.Logger) *ErrorMetrics {
	if logger == nil {
		logger = logging.GetDefaultLogger()
	}

	return &ErrorMetrics{
		registry:     registry,
		errorCounter: registry.GetCounter("courier_service_errors_total", nil),
		errorByCode:  make(map[errors.ErrorCode]*Counter),
		errorByType:  make(map[errors.ErrorType]*Counter),
		responseTime: registry.GetHistogram(
			"courier_service_response_time_seconds",
			[]float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
			nil,
		),
		logger: logger,
	}
}

// RecordError 记录错误
func (em *ErrorMetrics) RecordError(err error, labels map[string]string) {
	em.errorCounter.Inc()

	var courierErr *errors.CourierServiceError
	if errors.As(err, &courierErr) {
		// 按错误代码统计
		em.mutex.Lock()
		if counter, exists := em.errorByCode[courierErr.Code]; exists {
			counter.Inc()
		} else {
			counterLabels := map[string]string{"error_code": string(courierErr.Code)}
			for k, v := range labels {
				counterLabels[k] = v
			}
			counter = em.registry.GetCounter("courier_service_errors_by_code", counterLabels)
			em.errorByCode[courierErr.Code] = counter
			counter.Inc()
		}

		// 按错误类型统计
		if counter, exists := em.errorByType[courierErr.Type]; exists {
			counter.Inc()
		} else {
			counterLabels := map[string]string{"error_type": string(courierErr.Type)}
			for k, v := range labels {
				counterLabels[k] = v
			}
			counter = em.registry.GetCounter("courier_service_errors_by_type", counterLabels)
			em.errorByType[courierErr.Type] = counter
			counter.Inc()
		}
		em.mutex.Unlock()

		em.logger.Debug("Error metric recorded",
			"error_code", courierErr.Code,
			"error_type", courierErr.Type,
			"retryable", courierErr.Retryable,
		)
	}
}

// RecordResponseTime 记录响应时间
func (em *ErrorMetrics) RecordResponseTime(duration time.Duration) {
	em.responseTime.Observe(duration.Seconds())
}

// BusinessMetrics 业务指标收集器
type BusinessMetrics struct {
	registry             *MetricsRegistry
	taskAssignments      *Counter
	taskCompletions      *Counter
	courierRegistrations *Counter
	activeConnections    *Gauge
	queueSize            *Gauge
}

// NewBusinessMetrics 创建业务指标收集器
func NewBusinessMetrics(registry *MetricsRegistry) *BusinessMetrics {
	return &BusinessMetrics{
		registry:             registry,
		taskAssignments:      registry.GetCounter("courier_service_task_assignments_total", nil),
		taskCompletions:      registry.GetCounter("courier_service_task_completions_total", nil),
		courierRegistrations: registry.GetCounter("courier_service_courier_registrations_total", nil),
		activeConnections:    registry.GetGauge("courier_service_active_connections", nil),
		queueSize:            registry.GetGauge("courier_service_queue_size", nil),
	}
}

// RecordTaskAssignment 记录任务分配
func (bm *BusinessMetrics) RecordTaskAssignment(assignmentType string) {
	labels := map[string]string{"assignment_type": assignmentType}
	counter := bm.registry.GetCounter("courier_service_task_assignments_by_type", labels)
	counter.Inc()
	bm.taskAssignments.Inc()
}

// RecordTaskCompletion 记录任务完成
func (bm *BusinessMetrics) RecordTaskCompletion(status string) {
	labels := map[string]string{"status": status}
	counter := bm.registry.GetCounter("courier_service_task_completions_by_status", labels)
	counter.Inc()
	bm.taskCompletions.Inc()
}

// RecordCourierRegistration 记录信使注册
func (bm *BusinessMetrics) RecordCourierRegistration(level int) {
	labels := map[string]string{"level": string(rune('0' + level))}
	counter := bm.registry.GetCounter("courier_service_courier_registrations_by_level", labels)
	counter.Inc()
	bm.courierRegistrations.Inc()
}

// SetActiveConnections 设置活跃连接数
func (bm *BusinessMetrics) SetActiveConnections(count int) {
	bm.activeConnections.Set(float64(count))
}

// SetQueueSize 设置队列大小
func (bm *BusinessMetrics) SetQueueSize(size int) {
	bm.queueSize.Set(float64(size))
}

// 全局指标管理
var (
	globalRegistry        *MetricsRegistry
	globalErrorMetrics    *ErrorMetrics
	globalBusinessMetrics *BusinessMetrics
	globalOnce            sync.Once
)

// InitGlobalMetrics 初始化全局指标
func InitGlobalMetrics(logger logging.Logger) {
	globalOnce.Do(func() {
		globalRegistry = NewMetricsRegistry(logger)
		globalErrorMetrics = NewErrorMetrics(globalRegistry, logger)
		globalBusinessMetrics = NewBusinessMetrics(globalRegistry)
	})
}

// GetGlobalRegistry 获取全局注册表
func GetGlobalRegistry() *MetricsRegistry {
	if globalRegistry == nil {
		InitGlobalMetrics(nil)
	}
	return globalRegistry
}

// GetGlobalErrorMetrics 获取全局错误指标
func GetGlobalErrorMetrics() *ErrorMetrics {
	if globalErrorMetrics == nil {
		InitGlobalMetrics(nil)
	}
	return globalErrorMetrics
}

// GetGlobalBusinessMetrics 获取全局业务指标
func GetGlobalBusinessMetrics() *BusinessMetrics {
	if globalBusinessMetrics == nil {
		InitGlobalMetrics(nil)
	}
	return globalBusinessMetrics
}

// 便捷函数

// RecordError 记录错误（全局）
func RecordError(err error, labels map[string]string) {
	GetGlobalErrorMetrics().RecordError(err, labels)
}

// RecordResponseTime 记录响应时间（全局）
func RecordResponseTime(duration time.Duration) {
	GetGlobalErrorMetrics().RecordResponseTime(duration)
}

// RecordTaskAssignment 记录任务分配（全局）
func RecordTaskAssignment(assignmentType string) {
	GetGlobalBusinessMetrics().RecordTaskAssignment(assignmentType)
}

// RecordTaskCompletion 记录任务完成（全局）
func RecordTaskCompletion(status string) {
	GetGlobalBusinessMetrics().RecordTaskCompletion(status)
}

// RecordCourierRegistration 记录信使注册（全局）
func RecordCourierRegistration(level int) {
	GetGlobalBusinessMetrics().RecordCourierRegistration(level)
}

// SetActiveConnections 设置活跃连接数（全局）
func SetActiveConnections(count int) {
	GetGlobalBusinessMetrics().SetActiveConnections(count)
}

// SetQueueSize 设置队列大小（全局）
func SetQueueSize(size int) {
	GetGlobalBusinessMetrics().SetQueueSize(size)
}
