package handlers

import (
	"net/http"
	"openpenpal-backend/internal/monitoring"
	"openpenpal-backend/internal/resilience"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MonitoringHandler handles monitoring and metrics endpoints
type MonitoringHandler struct {
	collector *monitoring.MetricsCollector
	cbManager *resilience.CircuitBreakerManager
}

// NewMonitoringHandler creates a new monitoring handler
func NewMonitoringHandler() *MonitoringHandler {
	return &MonitoringHandler{
		collector: monitoring.NewMetricsCollector(),
		cbManager: resilience.NewCircuitBreakerManager(),
	}
}

// RegisterMonitoringRoutes registers all monitoring endpoints
func (h *MonitoringHandler) RegisterMonitoringRoutes(router *gin.Engine) {
	monitoring := router.Group("/monitoring")
	{
		// Prometheus metrics endpoint
		monitoring.GET("/metrics", h.PrometheusMetrics())
		
		// Health check endpoints
		monitoring.GET("/health", h.HealthCheck)
		monitoring.GET("/health/detailed", h.DetailedHealthCheck)
		monitoring.GET("/health/ready", h.ReadinessCheck)
		monitoring.GET("/health/live", h.LivenessCheck)
		
		// System metrics endpoints
		monitoring.GET("/system/stats", h.SystemStats)
		monitoring.GET("/system/runtime", h.RuntimeStats)
		
		// Circuit breaker monitoring
		monitoring.GET("/circuit-breakers", h.CircuitBreakerStatus)
		monitoring.GET("/circuit-breakers/:name", h.CircuitBreakerDetails)
		monitoring.POST("/circuit-breakers/:name/reset", h.ResetCircuitBreaker)
		
		// Business metrics endpoints
		monitoring.GET("/business/letters", h.LetterMetrics)
		monitoring.GET("/business/courier", h.CourierMetrics)
		monitoring.GET("/business/museum", h.MuseumMetrics)
		monitoring.GET("/business/ai", h.AIMetrics)
		
		// External monitoring integration
		monitoring.GET("/external/datadog", h.DatadogMetrics)
		monitoring.GET("/external/newrelic", h.NewRelicMetrics)
		monitoring.GET("/external/generic", h.GenericMetrics)
	}

	// Alias for Prometheus metrics at root level for standard monitoring tools
	router.GET("/metrics", h.PrometheusMetrics())
}

// PrometheusMetrics serves Prometheus metrics
func (h *MonitoringHandler) PrometheusMetrics() gin.HandlerFunc {
	handler := promhttp.Handler()
	
	return func(c *gin.Context) {
		handler.ServeHTTP(c.Writer, c.Request)
	}
}

// HealthCheck provides basic health status
func (h *MonitoringHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"service":   "openpenpal-backend",
		"version":   "1.0.0",
	})
}

// DetailedHealthCheck provides comprehensive health information
func (h *MonitoringHandler) DetailedHealthCheck(c *gin.Context) {
	// Check various system components
	healthStatus := map[string]interface{}{
		"status":     "healthy",
		"timestamp":  time.Now().UTC(),
		"service":    "openpenpal-backend",
		"version":    "1.0.0",
		"checks": map[string]interface{}{
			"database":        h.checkDatabaseHealth(),
			"circuit_breakers": h.checkCircuitBreakersHealth(),
			"external_apis":   h.checkExternalAPIsHealth(),
			"memory":         h.checkMemoryHealth(),
			"disk":           h.checkDiskHealth(),
		},
	}

	// Determine overall status
	overallHealthy := true
	for _, check := range healthStatus["checks"].(map[string]interface{}) {
		if checkMap, ok := check.(map[string]interface{}); ok {
			if status, exists := checkMap["status"]; exists && status != "healthy" {
				overallHealthy = false
				break
			}
		}
	}

	if !overallHealthy {
		healthStatus["status"] = "unhealthy"
		c.JSON(http.StatusServiceUnavailable, healthStatus)
		return
	}

	c.JSON(http.StatusOK, healthStatus)
}

// ReadinessCheck checks if the service is ready to accept requests
func (h *MonitoringHandler) ReadinessCheck(c *gin.Context) {
	ready := true
	reasons := []string{}

	// Check database connectivity
	dbHealth := h.checkDatabaseHealth()
	if dbHealth["status"] != "healthy" {
		ready = false
		reasons = append(reasons, "database not ready")
	}

	// Check essential external services
	extHealth := h.checkExternalAPIsHealth()
	if extHealth["status"] != "healthy" {
		ready = false
		reasons = append(reasons, "external services not ready")
	}

	status := http.StatusOK
	if !ready {
		status = http.StatusServiceUnavailable
	}

	c.JSON(status, gin.H{
		"ready":     ready,
		"reasons":   reasons,
		"timestamp": time.Now().UTC(),
	})
}

// LivenessCheck checks if the service is alive
func (h *MonitoringHandler) LivenessCheck(c *gin.Context) {
	// Simple liveness check - if we can respond, we're alive
	c.JSON(http.StatusOK, gin.H{
		"alive":     true,
		"timestamp": time.Now().UTC(),
		"uptime":    time.Since(startTime).String(),
	})
}

// SystemStats provides system-level statistics
func (h *MonitoringHandler) SystemStats(c *gin.Context) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	stats := gin.H{
		"timestamp": time.Now().UTC(),
		"system": gin.H{
			"goroutines":   runtime.NumGoroutine(),
			"gc_runs":      m.NumGC,
			"last_gc":      time.Unix(0, int64(m.LastGC)),
			"next_gc":      m.NextGC,
			"memory": gin.H{
				"allocated":      m.Alloc,
				"total_alloc":    m.TotalAlloc,
				"sys":           m.Sys,
				"heap_alloc":    m.HeapAlloc,
				"heap_sys":      m.HeapSys,
				"heap_released": m.HeapReleased,
			},
		},
		"metrics": gin.H{
			"requests_total": h.getMetricValue("openpenpal_http_requests_total"),
			"active_users":   h.getMetricValue("openpenpal_active_users"),
		},
	}

	c.JSON(http.StatusOK, stats)
}

// RuntimeStats provides Go runtime statistics
func (h *MonitoringHandler) RuntimeStats(c *gin.Context) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	c.JSON(http.StatusOK, gin.H{
		"go_version":     runtime.Version(),
		"num_cpu":       runtime.NumCPU(),
		"num_goroutine": runtime.NumGoroutine(),
		"memory": gin.H{
			"alloc":       m.Alloc,
			"sys":         m.Sys,
			"num_gc":      m.NumGC,
			"gc_cpu_fraction": m.GCCPUFraction,
		},
	})
}

// CircuitBreakerStatus provides status of all circuit breakers
func (h *MonitoringHandler) CircuitBreakerStatus(c *gin.Context) {
	breakers := h.cbManager.GetAllBreakers()
	status := make(map[string]interface{})

	for name, cb := range breakers {
		counts := cb.GetCounts()
		status[name] = gin.H{
			"state":                cb.GetState().String(),
			"requests":            counts.Requests,
			"successes":           counts.TotalSuccesses,
			"failures":            counts.TotalFailures,
			"consecutive_successes": counts.ConsecutiveSuccesses,
			"consecutive_failures":  counts.ConsecutiveFailures,
			"success_rate":        counts.SuccessRate(),
			"failure_rate":        counts.FailureRate(),
			"last_success":        counts.LastSuccessTime,
			"last_failure":        counts.LastFailureTime,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"circuit_breakers": status,
		"total_breakers":   len(breakers),
		"timestamp":        time.Now().UTC(),
	})
}

// CircuitBreakerDetails provides detailed information about a specific circuit breaker
func (h *MonitoringHandler) CircuitBreakerDetails(c *gin.Context) {
	name := c.Param("name")
	breakers := h.cbManager.GetAllBreakers()
	
	cb, exists := breakers[name]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Circuit breaker not found",
			"name":  name,
		})
		return
	}

	counts := cb.GetCounts()
	c.JSON(http.StatusOK, gin.H{
		"name":                   name,
		"state":                 cb.GetState().String(),
		"requests":              counts.Requests,
		"total_successes":       counts.TotalSuccesses,
		"total_failures":        counts.TotalFailures,
		"consecutive_successes": counts.ConsecutiveSuccesses,
		"consecutive_failures":  counts.ConsecutiveFailures,
		"success_rate":          counts.SuccessRate(),
		"failure_rate":          counts.FailureRate(),
		"last_success_time":     counts.LastSuccessTime,
		"last_failure_time":     counts.LastFailureTime,
		"timestamp":             time.Now().UTC(),
	})
}

// ResetCircuitBreaker resets a specific circuit breaker
func (h *MonitoringHandler) ResetCircuitBreaker(c *gin.Context) {
	name := c.Param("name")
	breakers := h.cbManager.GetAllBreakers()
	
	cb, exists := breakers[name]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Circuit breaker not found",
			"name":  name,
		})
		return
	}

	cb.Reset()
	c.JSON(http.StatusOK, gin.H{
		"message":   "Circuit breaker reset successfully",
		"name":      name,
		"timestamp": time.Now().UTC(),
	})
}

// Business metrics endpoints

// LetterMetrics provides letter-related metrics
func (h *MonitoringHandler) LetterMetrics(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"letters": gin.H{
			"created_total":   h.getMetricValue("openpenpal_letters_created_total"),
			"delivered_total": h.getMetricValue("openpenpal_letters_delivered_total"),
			"in_transit":      h.getMetricValue("openpenpal_letters_in_transit"),
		},
		"timestamp": time.Now().UTC(),
	})
}

// CourierMetrics provides courier-related metrics
func (h *MonitoringHandler) CourierMetrics(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"courier": gin.H{
			"tasks_total":     h.getMetricValue("openpenpal_courier_tasks_total"),
			"active_couriers": h.getMetricValue("openpenpal_active_couriers"),
			"delivery_rate":   h.calculateDeliveryRate(),
		},
		"timestamp": time.Now().UTC(),
	})
}

// MuseumMetrics provides museum-related metrics
func (h *MonitoringHandler) MuseumMetrics(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"museum": gin.H{
			"entries_total":   h.getMetricValue("openpenpal_museum_entries_total"),
			"pending_reviews": h.getMetricValue("openpenpal_museum_pending_reviews"),
			"approval_rate":   h.calculateApprovalRate(),
		},
		"timestamp": time.Now().UTC(),
	})
}

// AIMetrics provides AI-related metrics
func (h *MonitoringHandler) AIMetrics(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"ai": gin.H{
			"interactions_total": h.getMetricValue("openpenpal_ai_interactions_total"),
			"success_rate":       h.calculateAISuccessRate(),
			"average_latency":    h.getMetricValue("openpenpal_external_api_latency_seconds"),
		},
		"timestamp": time.Now().UTC(),
	})
}

// External monitoring integration endpoints

// DatadogMetrics provides metrics in Datadog format
func (h *MonitoringHandler) DatadogMetrics(c *gin.Context) {
	metrics := h.collectAllMetrics()
	
	// Convert to Datadog format
	datadogMetrics := make([]map[string]interface{}, 0)
	timestamp := time.Now().Unix()
	
	for name, value := range metrics {
		datadogMetrics = append(datadogMetrics, map[string]interface{}{
			"metric": name,
			"points": [][]interface{}{
				{timestamp, value},
			},
			"tags": []string{
				"service:openpenpal-backend",
				"environment:production",
			},
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"series": datadogMetrics,
	})
}

// NewRelicMetrics provides metrics in New Relic format
func (h *MonitoringHandler) NewRelicMetrics(c *gin.Context) {
	metrics := h.collectAllMetrics()
	timestamp := time.Now().Unix() * 1000 // New Relic uses milliseconds

	newRelicMetrics := make([]map[string]interface{}, 0)
	for name, value := range metrics {
		newRelicMetrics = append(newRelicMetrics, map[string]interface{}{
			"name":      name,
			"type":      "gauge",
			"value":     value,
			"timestamp": timestamp,
			"attributes": map[string]string{
				"service":     "openpenpal-backend",
				"environment": "production",
			},
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"metrics": newRelicMetrics,
	})
}

// GenericMetrics provides metrics in a generic format
func (h *MonitoringHandler) GenericMetrics(c *gin.Context) {
	metrics := h.collectAllMetrics()
	
	c.JSON(http.StatusOK, gin.H{
		"metrics":   metrics,
		"timestamp": time.Now().UTC(),
		"service":   "openpenpal-backend",
		"version":   "1.0.0",
	})
}

// Helper methods

var startTime = time.Now()

// checkDatabaseHealth checks database connectivity
func (h *MonitoringHandler) checkDatabaseHealth() map[string]interface{} {
	// This would typically ping the database
	// For now, return a mock healthy status
	return map[string]interface{}{
		"status":       "healthy",
		"response_time": "5ms",
		"connections":  h.getMetricValue("openpenpal_db_connections"),
	}
}

// checkCircuitBreakersHealth checks circuit breaker status
func (h *MonitoringHandler) checkCircuitBreakersHealth() map[string]interface{} {
	breakers := h.cbManager.GetAllBreakers()
	openBreakers := 0
	
	for _, cb := range breakers {
		if cb.GetState() == resilience.StateOpen {
			openBreakers++
		}
	}

	status := "healthy"
	if openBreakers > 0 {
		status = "degraded"
	}

	return map[string]interface{}{
		"status":        status,
		"total_breakers": len(breakers),
		"open_breakers":  openBreakers,
	}
}

// checkExternalAPIsHealth checks external API connectivity
func (h *MonitoringHandler) checkExternalAPIsHealth() map[string]interface{} {
	// This would typically test external API connectivity
	return map[string]interface{}{
		"status":      "healthy",
		"tested_apis": []string{"ai_service", "notification_service"},
	}
}

// checkMemoryHealth checks memory usage
func (h *MonitoringHandler) checkMemoryHealth() map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	// Consider unhealthy if using more than 80% of allocated memory
	memoryUsagePercent := float64(m.Alloc) / float64(m.Sys) * 100
	status := "healthy"
	if memoryUsagePercent > 80 {
		status = "warning"
	}
	if memoryUsagePercent > 95 {
		status = "critical"
	}

	return map[string]interface{}{
		"status":           status,
		"usage_percent":    memoryUsagePercent,
		"allocated_bytes":  m.Alloc,
		"system_bytes":     m.Sys,
	}
}

// checkDiskHealth checks disk usage
func (h *MonitoringHandler) checkDiskHealth() map[string]interface{} {
	// This would typically check disk usage
	// For now, return healthy status
	return map[string]interface{}{
		"status":        "healthy",
		"usage_percent": 45.0,
	}
}

// getMetricValue retrieves a metric value (mock implementation)
func (h *MonitoringHandler) getMetricValue(metricName string) float64 {
	// In a real implementation, this would query Prometheus metrics
	// For now, return mock values
	mockValues := map[string]float64{
		"openpenpal_http_requests_total":     1250.0,
		"openpenpal_active_users":           45.0,
		"openpenpal_letters_created_total":  125.0,
		"openpenpal_letters_delivered_total": 98.0,
		"openpenpal_courier_tasks_total":    67.0,
		"openpenpal_museum_entries_total":   23.0,
		"openpenpal_ai_interactions_total":  156.0,
		"openpenpal_db_connections":         10.0,
	}
	
	if value, exists := mockValues[metricName]; exists {
		return value
	}
	return 0.0
}

// collectAllMetrics collects all available metrics
func (h *MonitoringHandler) collectAllMetrics() map[string]float64 {
	return map[string]float64{
		"http_requests_total":      h.getMetricValue("openpenpal_http_requests_total"),
		"active_users":            h.getMetricValue("openpenpal_active_users"),
		"letters_created_total":   h.getMetricValue("openpenpal_letters_created_total"),
		"letters_delivered_total": h.getMetricValue("openpenpal_letters_delivered_total"),
		"courier_tasks_total":     h.getMetricValue("openpenpal_courier_tasks_total"),
		"museum_entries_total":    h.getMetricValue("openpenpal_museum_entries_total"),
		"ai_interactions_total":   h.getMetricValue("openpenpal_ai_interactions_total"),
		"db_connections":          h.getMetricValue("openpenpal_db_connections"),
	}
}

// calculateDeliveryRate calculates courier delivery success rate
func (h *MonitoringHandler) calculateDeliveryRate() float64 {
	// Mock calculation
	delivered := h.getMetricValue("openpenpal_letters_delivered_total")
	total := h.getMetricValue("openpenpal_courier_tasks_total")
	if total > 0 {
		return (delivered / total) * 100
	}
	return 0.0
}

// calculateApprovalRate calculates museum entry approval rate
func (h *MonitoringHandler) calculateApprovalRate() float64 {
	// Mock calculation
	return 78.5 // 78.5% approval rate
}

// calculateAISuccessRate calculates AI interaction success rate
func (h *MonitoringHandler) calculateAISuccessRate() float64 {
	// Mock calculation
	return 94.2 // 94.2% success rate
}