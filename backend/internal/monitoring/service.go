package monitoring

import (
	"context"
	"fmt"
	"net/http"
	"openpenpal-backend/internal/resilience"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MonitoringService provides comprehensive monitoring and metrics collection
type MonitoringService struct {
	collector           *MetricsCollector
	cbIntegration      *CircuitBreakerMetricsIntegration
	server             *http.Server
	config             *MonitoringConfig
	ctx                context.Context
	cancel             context.CancelFunc
	wg                 sync.WaitGroup
	isRunning          bool
	mutex              sync.RWMutex
}

// MonitoringConfig holds configuration for the monitoring service
type MonitoringConfig struct {
	// Metrics collection settings
	MetricsCollectionInterval time.Duration
	MetricsPort               string
	MetricsPath               string
	
	// Circuit breaker monitoring settings
	CircuitBreakerMonitoringInterval time.Duration
	AlertThresholds                  AlertThresholds
	
	// External integrations
	ExternalIntegrations ExternalIntegrations
	
	// System monitoring settings
	SystemMetricsEnabled bool
	BusinessMetricsEnabled bool
}

// AlertThresholds defines thresholds for various alerts
type AlertThresholds struct {
	HighFailureRate           float64
	ConsecutiveFailuresLimit  uint32
	CircuitOpenAlertEnabled   bool
	MemoryUsageThreshold      float64
	ResponseTimeThreshold     time.Duration
}

// ExternalIntegrations defines external monitoring integrations
type ExternalIntegrations struct {
	DatadogEnabled    bool
	DatadogAPIKey     string
	NewRelicEnabled   bool
	NewRelicLicenseKey string
	WebhookURL        string
	SlackWebhook      string
}

// NewMonitoringService creates a new monitoring service
func NewMonitoringService(config *MonitoringConfig) *MonitoringService {
	if config == nil {
		config = DefaultMonitoringConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())
	
	collector := NewMetricsCollector()
	cbManager := resilience.DefaultCircuitBreakerManager
	cbIntegration := NewCircuitBreakerMetricsIntegration(collector, cbManager)

	return &MonitoringService{
		collector:     collector,
		cbIntegration: cbIntegration,
		config:        config,
		ctx:           ctx,
		cancel:        cancel,
	}
}

// Start starts the monitoring service
func (ms *MonitoringService) Start() error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	if ms.isRunning {
		return fmt.Errorf("monitoring service is already running")
	}

	// Start circuit breaker metrics integration
	ms.cbIntegration.Start(ms.config.CircuitBreakerMonitoringInterval)

	// Start system metrics collection
	if ms.config.SystemMetricsEnabled {
		ms.startSystemMetricsCollection()
	}

	// Start metrics server
	if err := ms.startMetricsServer(); err != nil {
		return fmt.Errorf("failed to start metrics server: %w", err)
	}

	ms.isRunning = true
	return nil
}

// Stop stops the monitoring service
func (ms *MonitoringService) Stop() error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	if !ms.isRunning {
		return nil
	}

	// Stop circuit breaker integration
	ms.cbIntegration.Stop()

	// Stop metrics server
	if ms.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := ms.server.Shutdown(ctx); err != nil {
			return fmt.Errorf("failed to shutdown metrics server: %w", err)
		}
	}

	// Cancel context and wait for goroutines
	ms.cancel()
	ms.wg.Wait()

	ms.isRunning = false
	return nil
}

// startMetricsServer starts the Prometheus metrics HTTP server
func (ms *MonitoringService) startMetricsServer() error {
	router := gin.New()
	router.Use(gin.Recovery())

	// Add Prometheus metrics endpoint
	router.GET(ms.config.MetricsPath, func(c *gin.Context) {
		promhttp.Handler().ServeHTTP(c.Writer, c.Request)
	})

	// Add health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"service": "monitoring",
			"timestamp": time.Now().UTC(),
		})
	})

	ms.server = &http.Server{
		Addr:    ":" + ms.config.MetricsPort,
		Handler: router,
	}

	ms.wg.Add(1)
	go func() {
		defer ms.wg.Done()
		if err := ms.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Metrics server error: %v\n", err)
		}
	}()

	return nil
}

// startSystemMetricsCollection starts collecting system-level metrics
func (ms *MonitoringService) startSystemMetricsCollection() {
	ms.wg.Add(1)
	go func() {
		defer ms.wg.Done()
		ticker := time.NewTicker(ms.config.MetricsCollectionInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				ms.collectSystemMetrics()
			case <-ms.ctx.Done():
				return
			}
		}
	}()
}

// collectSystemMetrics collects system-level metrics
func (ms *MonitoringService) collectSystemMetrics() {
	// This would collect various system metrics like:
	// - Memory usage
	// - CPU usage
	// - Disk usage
	// - Network statistics
	// - Database connection pool stats
	// - Queue sizes
	// - Active user counts

	// Example implementation:
	ms.collectMemoryMetrics()
	ms.collectDatabaseMetrics()
	ms.collectBusinessMetrics()
}

// collectMemoryMetrics collects memory-related metrics
func (ms *MonitoringService) collectMemoryMetrics() {
	// This would collect Go runtime memory metrics
	// and set them using the metrics collector
}

// collectDatabaseMetrics collects database-related metrics
func (ms *MonitoringService) collectDatabaseMetrics() {
	// This would collect database connection pool metrics
	// and query performance metrics
}

// collectBusinessMetrics collects business-specific metrics
func (ms *MonitoringService) collectBusinessMetrics() {
	if !ms.config.BusinessMetricsEnabled {
		return
	}

	// This would collect business metrics like:
	// - Active user counts
	// - Letter creation rates
	// - Courier task completion rates
	// - Museum entry submission rates
	// - AI interaction success rates
}

// GetCollector returns the metrics collector instance
func (ms *MonitoringService) GetCollector() *MetricsCollector {
	return ms.collector
}

// GetCircuitBreakerIntegration returns the circuit breaker integration instance
func (ms *MonitoringService) GetCircuitBreakerIntegration() *CircuitBreakerMetricsIntegration {
	return ms.cbIntegration
}

// GetStatus returns the current status of the monitoring service
func (ms *MonitoringService) GetStatus() map[string]interface{} {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	status := map[string]interface{}{
		"running":    ms.isRunning,
		"timestamp":  time.Now().UTC(),
		"config": map[string]interface{}{
			"metrics_port":                        ms.config.MetricsPort,
			"metrics_path":                       ms.config.MetricsPath,
			"metrics_collection_interval":        ms.config.MetricsCollectionInterval.String(),
			"circuit_breaker_monitoring_interval": ms.config.CircuitBreakerMonitoringInterval.String(),
			"system_metrics_enabled":             ms.config.SystemMetricsEnabled,
			"business_metrics_enabled":           ms.config.BusinessMetricsEnabled,
		},
	}

	if ms.isRunning {
		status["circuit_breaker_integration"] = ms.cbIntegration.GetIntegrationStatus()
		status["server_address"] = ms.server.Addr
	}

	return status
}

// RegisterWithGin registers monitoring middleware with a Gin router
func (ms *MonitoringService) RegisterWithGin(router *gin.Engine, serviceName string) {
	// Note: Metrics middleware will be applied separately to avoid import cycle
	// This function exists for future extensibility
}

// SendExternalMetrics sends metrics to external monitoring systems
func (ms *MonitoringService) SendExternalMetrics() error {
	if !ms.config.ExternalIntegrations.DatadogEnabled && 
	   !ms.config.ExternalIntegrations.NewRelicEnabled {
		return nil // No external integrations enabled
	}

	// Collect all metrics
	metrics := ms.collectAllMetricsForExport()

	// Send to enabled external systems
	if ms.config.ExternalIntegrations.DatadogEnabled {
		if err := ms.sendToDatadog(metrics); err != nil {
			return fmt.Errorf("failed to send metrics to Datadog: %w", err)
		}
	}

	if ms.config.ExternalIntegrations.NewRelicEnabled {
		if err := ms.sendToNewRelic(metrics); err != nil {
			return fmt.Errorf("failed to send metrics to New Relic: %w", err)
		}
	}

	return nil
}

// collectAllMetricsForExport collects all metrics for external export
func (ms *MonitoringService) collectAllMetricsForExport() map[string]float64 {
	// This would collect all current metric values
	// from the Prometheus registry for external export
	return map[string]float64{
		"http_requests_total":      1000,
		"http_request_duration":    0.150,
		"circuit_breaker_failures": 10,
		"active_users":            75,
		// ... more metrics
	}
}

// sendToDatadog sends metrics to Datadog
func (ms *MonitoringService) sendToDatadog(metrics map[string]float64) error {
	// Implementation for sending metrics to Datadog API
	// This would use the Datadog Go client or HTTP API
	return nil
}

// sendToNewRelic sends metrics to New Relic
func (ms *MonitoringService) sendToNewRelic(metrics map[string]float64) error {
	// Implementation for sending metrics to New Relic API
	// This would use the New Relic Go client or HTTP API
	return nil
}

// DefaultMonitoringConfig returns a default monitoring configuration
func DefaultMonitoringConfig() *MonitoringConfig {
	return &MonitoringConfig{
		MetricsCollectionInterval:            30 * time.Second,
		MetricsPort:                         "9090",
		MetricsPath:                         "/metrics",
		CircuitBreakerMonitoringInterval:    10 * time.Second,
		SystemMetricsEnabled:                true,
		BusinessMetricsEnabled:              true,
		AlertThresholds: AlertThresholds{
			HighFailureRate:          80.0,
			ConsecutiveFailuresLimit: 10,
			CircuitOpenAlertEnabled:  true,
			MemoryUsageThreshold:     85.0,
			ResponseTimeThreshold:    5 * time.Second,
		},
		ExternalIntegrations: ExternalIntegrations{
			DatadogEnabled:  false,
			NewRelicEnabled: false,
		},
	}
}

// Global monitoring service instance
var DefaultMonitoringService = NewMonitoringService(DefaultMonitoringConfig())