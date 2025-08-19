package monitoring

import (
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	once               sync.Once
	metricsCollectorInstance *MetricsCollector
)

// MetricsCollector manages all Prometheus metrics for the OpenPenPal platform
type MetricsCollector struct {
	// HTTP Request metrics
	httpRequestsTotal     *prometheus.CounterVec
	httpRequestDuration   *prometheus.HistogramVec
	httpRequestsInFlight  *prometheus.GaugeVec
	httpResponseSize      *prometheus.HistogramVec

	// Business logic metrics
	lettersCreated        *prometheus.CounterVec
	lettersDelivered      *prometheus.CounterVec
	courierTasks         *prometheus.CounterVec
	museumEntries        *prometheus.CounterVec
	aiInteractions       *prometheus.CounterVec

	// Circuit breaker metrics
	circuitBreakerState   *prometheus.GaugeVec
	circuitBreakerRequests *prometheus.CounterVec
	circuitBreakerFailures *prometheus.CounterVec

	// Database metrics
	dbConnections         *prometheus.GaugeVec
	dbQueryDuration       *prometheus.HistogramVec
	dbTransactions        *prometheus.CounterVec

	// System metrics
	activeUsers           *prometheus.GaugeVec
	queueSize            *prometheus.GaugeVec
	creditTransactions   *prometheus.CounterVec

	// External service metrics
	externalAPIRequests   *prometheus.CounterVec
	externalAPILatency    *prometheus.HistogramVec

	mutex sync.RWMutex
}

// NewMetricsCollector creates a new metrics collector with all required metrics
// Uses singleton pattern to prevent duplicate registration
func NewMetricsCollector() *MetricsCollector {
	once.Do(func() {
		metricsCollectorInstance = &MetricsCollector{}
		metricsCollectorInstance.initializeMetrics()
	})
	return metricsCollectorInstance
}

// initializeMetrics sets up all Prometheus metrics
func (mc *MetricsCollector) initializeMetrics() {
	// HTTP Request metrics
	mc.httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "openpenpal_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status_code", "service"},
	)

	mc.httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "openpenpal_http_request_duration_seconds",
			Help:    "HTTP request latency distributions",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint", "service"},
	)

	mc.httpRequestsInFlight = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "openpenpal_http_requests_in_flight",
			Help: "Current number of HTTP requests being processed",
		},
		[]string{"service"},
	)

	mc.httpResponseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "openpenpal_http_response_size_bytes",
			Help:    "HTTP response size distributions",
			Buckets: []float64{100, 1000, 10000, 100000, 1000000},
		},
		[]string{"method", "endpoint", "service"},
	)

	// Business logic metrics
	mc.lettersCreated = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "openpenpal_letters_created_total",
			Help: "Total number of letters created",
		},
		[]string{"user_role", "letter_type", "visibility"},
	)

	mc.lettersDelivered = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "openpenpal_letters_delivered_total",
			Help: "Total number of letters delivered",
		},
		[]string{"courier_level", "delivery_method"},
	)

	mc.courierTasks = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "openpenpal_courier_tasks_total",
			Help: "Total number of courier tasks",
		},
		[]string{"task_type", "status", "courier_level"},
	)

	mc.museumEntries = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "openpenpal_museum_entries_total",
			Help: "Total number of museum entries",
		},
		[]string{"entry_type", "status"},
	)

	mc.aiInteractions = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "openpenpal_ai_interactions_total",
			Help: "Total number of AI interactions",
		},
		[]string{"interaction_type", "success"},
	)

	// Circuit breaker metrics
	mc.circuitBreakerState = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "openpenpal_circuit_breaker_state",
			Help: "Current circuit breaker state (0=closed, 1=open, 2=half-open)",
		},
		[]string{"service_name"},
	)

	mc.circuitBreakerRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "openpenpal_circuit_breaker_requests_total",
			Help: "Total number of requests through circuit breaker",
		},
		[]string{"service_name", "result"},
	)

	mc.circuitBreakerFailures = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "openpenpal_circuit_breaker_failures_total",
			Help: "Total number of circuit breaker failures",
		},
		[]string{"service_name", "failure_type"},
	)

	// Database metrics
	mc.dbConnections = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "openpenpal_db_connections",
			Help: "Current number of database connections",
		},
		[]string{"state"}, // open, in_use, idle
	)

	mc.dbQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "openpenpal_db_query_duration_seconds",
			Help:    "Database query duration distributions",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0, 5.0},
		},
		[]string{"operation", "table"},
	)

	mc.dbTransactions = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "openpenpal_db_transactions_total",
			Help: "Total number of database transactions",
		},
		[]string{"operation", "status"},
	)

	// System metrics
	mc.activeUsers = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "openpenpal_active_users",
			Help: "Current number of active users",
		},
		[]string{"time_window"}, // 5m, 15m, 1h, 24h
	)

	mc.queueSize = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "openpenpal_queue_size",
			Help: "Current queue sizes",
		},
		[]string{"queue_name"},
	)

	mc.creditTransactions = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "openpenpal_credit_transactions_total",
			Help: "Total number of credit transactions",
		},
		[]string{"transaction_type", "user_role"},
	)

	// External service metrics
	mc.externalAPIRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "openpenpal_external_api_requests_total",
			Help: "Total number of external API requests",
		},
		[]string{"api_name", "status_code"},
	)

	mc.externalAPILatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "openpenpal_external_api_latency_seconds",
			Help:    "External API request latency distributions",
			Buckets: []float64{0.1, 0.5, 1.0, 2.0, 5.0, 10.0, 30.0},
		},
		[]string{"api_name"},
	)
}

// HTTP Metrics methods
func (mc *MetricsCollector) IncrementHTTPRequests(method, endpoint, statusCode, service string) {
	mc.httpRequestsTotal.WithLabelValues(method, endpoint, statusCode, service).Inc()
}

func (mc *MetricsCollector) ObserveHTTPRequestDuration(method, endpoint, service string, duration time.Duration) {
	mc.httpRequestDuration.WithLabelValues(method, endpoint, service).Observe(duration.Seconds())
}

func (mc *MetricsCollector) IncrementHTTPRequestsInFlight(service string) {
	mc.httpRequestsInFlight.WithLabelValues(service).Inc()
}

func (mc *MetricsCollector) DecrementHTTPRequestsInFlight(service string) {
	mc.httpRequestsInFlight.WithLabelValues(service).Dec()
}

func (mc *MetricsCollector) ObserveHTTPResponseSize(method, endpoint, service string, size float64) {
	mc.httpResponseSize.WithLabelValues(method, endpoint, service).Observe(size)
}

// Business logic metrics methods
func (mc *MetricsCollector) IncrementLettersCreated(userRole, letterType, visibility string) {
	mc.lettersCreated.WithLabelValues(userRole, letterType, visibility).Inc()
}

func (mc *MetricsCollector) IncrementLettersDelivered(courierLevel, deliveryMethod string) {
	mc.lettersDelivered.WithLabelValues(courierLevel, deliveryMethod).Inc()
}

func (mc *MetricsCollector) IncrementCourierTasks(taskType, status, courierLevel string) {
	mc.courierTasks.WithLabelValues(taskType, status, courierLevel).Inc()
}

func (mc *MetricsCollector) IncrementMuseumEntries(entryType, status string) {
	mc.museumEntries.WithLabelValues(entryType, status).Inc()
}

func (mc *MetricsCollector) IncrementAIInteractions(interactionType string, success bool) {
	successStr := strconv.FormatBool(success)
	mc.aiInteractions.WithLabelValues(interactionType, successStr).Inc()
}

// Circuit breaker metrics methods
func (mc *MetricsCollector) SetCircuitBreakerState(serviceName string, state int) {
	mc.circuitBreakerState.WithLabelValues(serviceName).Set(float64(state))
}

func (mc *MetricsCollector) IncrementCircuitBreakerRequests(serviceName, result string) {
	mc.circuitBreakerRequests.WithLabelValues(serviceName, result).Inc()
}

func (mc *MetricsCollector) IncrementCircuitBreakerFailures(serviceName, failureType string) {
	mc.circuitBreakerFailures.WithLabelValues(serviceName, failureType).Inc()
}

// Database metrics methods
func (mc *MetricsCollector) SetDBConnections(state string, count float64) {
	mc.dbConnections.WithLabelValues(state).Set(count)
}

func (mc *MetricsCollector) ObserveDBQueryDuration(operation, table string, duration time.Duration) {
	mc.dbQueryDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
}

func (mc *MetricsCollector) IncrementDBTransactions(operation, status string) {
	mc.dbTransactions.WithLabelValues(operation, status).Inc()
}

// System metrics methods
func (mc *MetricsCollector) SetActiveUsers(timeWindow string, count float64) {
	mc.activeUsers.WithLabelValues(timeWindow).Set(count)
}

func (mc *MetricsCollector) SetQueueSize(queueName string, size float64) {
	mc.queueSize.WithLabelValues(queueName).Set(size)
}

func (mc *MetricsCollector) IncrementCreditTransactions(transactionType, userRole string) {
	mc.creditTransactions.WithLabelValues(transactionType, userRole).Inc()
}

// External service metrics methods
func (mc *MetricsCollector) IncrementExternalAPIRequests(apiName, statusCode string) {
	mc.externalAPIRequests.WithLabelValues(apiName, statusCode).Inc()
}

func (mc *MetricsCollector) ObserveExternalAPILatency(apiName string, duration time.Duration) {
	mc.externalAPILatency.WithLabelValues(apiName).Observe(duration.Seconds())
}

// GetMetricsRegistry returns the Prometheus registry for external use
func (mc *MetricsCollector) GetMetricsRegistry() *prometheus.Registry {
	return prometheus.DefaultRegisterer.(*prometheus.Registry)
}

// Default metrics collector instance
// Commented out to avoid duplicate registration
// var DefaultMetricsCollector = NewMetricsCollector()