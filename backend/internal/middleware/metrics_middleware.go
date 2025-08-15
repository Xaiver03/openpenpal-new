package middleware

import (
	"bytes"
	"io"
	"openpenpal-backend/internal/monitoring"
	"openpenpal-backend/internal/resilience"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// MetricsMiddleware provides comprehensive metrics collection for HTTP requests
type MetricsMiddleware struct {
	collector     *monitoring.MetricsCollector
	serviceName   string
	excludePaths  map[string]bool
}

// NewMetricsMiddleware creates a new metrics middleware
func NewMetricsMiddleware(serviceName string) *MetricsMiddleware {
	return &MetricsMiddleware{
		collector:   monitoring.DefaultMetricsCollector,
		serviceName: serviceName,
		excludePaths: map[string]bool{
			"/health":           true,
			"/metrics":          true,
			"/swagger":          true,
			"/favicon.ico":      true,
			"/ping":            true,
			"/docs/health":     true,
		},
	}
}

// SetMetricsCollector allows setting a custom metrics collector
func (mm *MetricsMiddleware) SetMetricsCollector(collector *monitoring.MetricsCollector) {
	mm.collector = collector
}

// RequestMetrics creates middleware for tracking HTTP request metrics
func (mm *MetricsMiddleware) RequestMetrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip metrics collection for excluded paths
		if mm.excludePaths[c.Request.URL.Path] {
			c.Next()
			return
		}

		start := time.Now()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		// Normalize path for better grouping
		normalizedPath := mm.normalizePath(path)

		// Increment in-flight requests
		mm.collector.IncrementHTTPRequestsInFlight(mm.serviceName)

		// Create a response writer wrapper to capture response size
		rww := &responseWriterWrapper{
			ResponseWriter: c.Writer,
			body:          &bytes.Buffer{},
		}
		c.Writer = rww

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Decrement in-flight requests
		mm.collector.DecrementHTTPRequestsInFlight(mm.serviceName)

		// Get status code
		statusCode := strconv.Itoa(c.Writer.Status())

		// Record metrics
		mm.collector.IncrementHTTPRequests(
			c.Request.Method,
			normalizedPath,
			statusCode,
			mm.serviceName,
		)

		mm.collector.ObserveHTTPRequestDuration(
			c.Request.Method,
			normalizedPath,
			mm.serviceName,
			duration,
		)

		// Record response size if available
		if rww.body != nil {
			mm.collector.ObserveHTTPResponseSize(
				c.Request.Method,
				normalizedPath,
				mm.serviceName,
				float64(rww.body.Len()),
			)
		}

		// Record error rates for alerting
		if c.Writer.Status() >= 400 {
			mm.recordErrorMetrics(c, normalizedPath, statusCode)
		}
	}
}

// CircuitBreakerMetrics creates middleware for tracking circuit breaker metrics
func (mm *MetricsMiddleware) CircuitBreakerMetrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if circuit breaker is present in context
		if cb, exists := c.Get("circuit_breaker"); exists {
			if circuitBreaker, ok := cb.(*resilience.CircuitBreaker); ok {
				serviceName := mm.getServiceNameFromContext(c)
				
				// Get initial state
				initialState := circuitBreaker.GetState()
				mm.collector.SetCircuitBreakerState(serviceName, int(initialState))

				// Set up monitoring for state changes during request
				defer func() {
					finalState := circuitBreaker.GetState()
					if finalState != initialState {
						mm.collector.SetCircuitBreakerState(serviceName, int(finalState))
						
						// Log state transition
						mm.recordCircuitBreakerTransition(serviceName, initialState, finalState)
					}
					
					// Record request outcome
					if c.Writer.Status() >= 500 {
						mm.collector.IncrementCircuitBreakerFailures(serviceName, "server_error")
					} else if c.Writer.Status() >= 400 {
						mm.collector.IncrementCircuitBreakerFailures(serviceName, "client_error")
					}
					
					// Record request result
					result := "success"
					if c.Writer.Status() >= 400 {
						result = "failure"
					}
					mm.collector.IncrementCircuitBreakerRequests(serviceName, result)
				}()
			}
		}

		c.Next()
	}
}

// BusinessMetrics creates middleware for tracking business-specific metrics
func (mm *MetricsMiddleware) BusinessMetrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Only track business metrics for successful requests
		if c.Writer.Status() < 400 {
			mm.trackBusinessMetrics(c)
		}
	}
}

// DatabaseMetrics creates middleware for tracking database operation metrics
func (mm *MetricsMiddleware) DatabaseMetrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set up database metrics collection in context
		c.Set("db_metrics_collector", mm.collector)
		
		c.Next()
	}
}

// ExternalAPIMetrics creates middleware for tracking external API call metrics
func (mm *MetricsMiddleware) ExternalAPIMetrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set up external API metrics collection in context
		c.Set("external_api_metrics_collector", mm.collector)
		
		c.Next()
	}
}

// Helper methods

// normalizePath normalizes URL paths for better metric grouping
func (mm *MetricsMiddleware) normalizePath(path string) string {
	// Replace path parameters with placeholders
	normalizedPath := path
	
	// Common patterns to normalize
	patterns := map[string]string{
		"/api/v1/letters/[0-9a-f-]+":     "/api/v1/letters/:id",
		"/api/v1/users/[0-9a-f-]+":      "/api/v1/users/:id",
		"/api/v1/courier/tasks/[0-9a-f-]+": "/api/v1/courier/tasks/:id",
		"/api/v1/museum/entries/[0-9a-f-]+": "/api/v1/museum/entries/:id",
		"/api/v1/admin/[^/]+/[0-9a-f-]+": "/api/v1/admin/:resource/:id",
	}
	
	for pattern, replacement := range patterns {
		// Simple pattern matching - in production, use regex
		if strings.Contains(normalizedPath, "letters/") && 
		   !strings.Contains(normalizedPath, "letters/barcode") {
			normalizedPath = strings.Replace(normalizedPath, 
				normalizedPath[strings.LastIndex(normalizedPath, "/")+1:],
				":id", 1)
		}
	}
	
	return normalizedPath
}

// getServiceNameFromContext extracts service name from context or returns default
func (mm *MetricsMiddleware) getServiceNameFromContext(c *gin.Context) string {
	if serviceName, exists := c.Get("service_name"); exists {
		if name, ok := serviceName.(string); ok {
			return name
		}
	}
	return mm.serviceName
}

// recordErrorMetrics records specific error metrics for alerting
func (mm *MetricsMiddleware) recordErrorMetrics(c *gin.Context, path, statusCode string) {
	// You can add specific error tracking here
	// For example, track rate limiting, authentication failures, etc.
	
	if c.Writer.Status() == 429 {
		// Rate limiting error
		mm.collector.IncrementCircuitBreakerFailures(mm.serviceName, "rate_limit")
	} else if c.Writer.Status() == 401 || c.Writer.Status() == 403 {
		// Authentication/authorization errors
		mm.collector.IncrementCircuitBreakerFailures(mm.serviceName, "auth_error")
	} else if c.Writer.Status() >= 500 {
		// Server errors
		mm.collector.IncrementCircuitBreakerFailures(mm.serviceName, "server_error")
	}
}

// recordCircuitBreakerTransition logs circuit breaker state transitions
func (mm *MetricsMiddleware) recordCircuitBreakerTransition(serviceName string, from, to resilience.CircuitBreakerState) {
	// Log state transitions for alerting
	switch to {
	case resilience.StateOpen:
		mm.collector.IncrementCircuitBreakerFailures(serviceName, "circuit_opened")
	case resilience.StateClosed:
		if from == resilience.StateOpen || from == resilience.StateHalfOpen {
			mm.collector.IncrementCircuitBreakerRequests(serviceName, "circuit_closed")
		}
	case resilience.StateHalfOpen:
		mm.collector.IncrementCircuitBreakerRequests(serviceName, "circuit_half_open")
	}
}

// trackBusinessMetrics tracks business-specific metrics based on the endpoint
func (mm *MetricsMiddleware) trackBusinessMetrics(c *gin.Context) {
	path := c.FullPath()
	method := c.Request.Method
	
	// Extract user role from context if available
	userRole := "anonymous"
	if user, exists := c.Get("user"); exists {
		// Assuming user has a Role field
		if userMap, ok := user.(map[string]interface{}); ok {
			if role, exists := userMap["role"]; exists {
				if roleStr, ok := role.(string); ok {
					userRole = roleStr
				}
			}
		}
	}

	// Track business metrics based on endpoints
	switch {
	case method == "POST" && strings.Contains(path, "/letters"):
		// Letter creation
		visibility := c.PostForm("visibility")
		if visibility == "" {
			visibility = "private" // default
		}
		mm.collector.IncrementLettersCreated(userRole, "standard", visibility)
		
	case method == "POST" && strings.Contains(path, "/courier/apply"):
		// Courier application
		mm.collector.IncrementCourierTasks("application", "submitted", userRole)
		
	case method == "POST" && strings.Contains(path, "/museum/submit"):
		// Museum submission
		mm.collector.IncrementMuseumEntries("letter", "submitted")
		
	case method == "POST" && strings.Contains(path, "/ai/"):
		// AI interaction
		success := c.Writer.Status() < 400
		interactionType := mm.extractAIInteractionType(path)
		mm.collector.IncrementAIInteractions(interactionType, success)
		
	case method == "POST" && strings.Contains(path, "/credit"):
		// Credit transaction
		transactionType := "earn"
		if strings.Contains(path, "spend") {
			transactionType = "spend"
		}
		mm.collector.IncrementCreditTransactions(transactionType, userRole)
	}
}

// extractAIInteractionType extracts AI interaction type from path
func (mm *MetricsMiddleware) extractAIInteractionType(path string) string {
	if strings.Contains(path, "/inspiration") {
		return "inspiration"
	} else if strings.Contains(path, "/match") {
		return "match"
	} else if strings.Contains(path, "/reply") {
		return "reply"
	}
	return "unknown"
}

// responseWriterWrapper wraps gin.ResponseWriter to capture response body size
type responseWriterWrapper struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (rww *responseWriterWrapper) Write(b []byte) (int, error) {
	// Write to both the original writer and our buffer
	n, err := rww.ResponseWriter.Write(b)
	if rww.body != nil {
		rww.body.Write(b[:n])
	}
	return n, err
}

func (rww *responseWriterWrapper) WriteString(s string) (int, error) {
	// Write to both the original writer and our buffer
	n, err := rww.ResponseWriter.WriteString(s)
	if rww.body != nil {
		rww.body.WriteString(s[:n])
	}
	return n, err
}

// Helper function for tracking external API calls
func TrackExternalAPICall(apiName string, statusCode int, duration time.Duration) {
	monitoring.DefaultMetricsCollector.IncrementExternalAPIRequests(apiName, strconv.Itoa(statusCode))
	monitoring.DefaultMetricsCollector.ObserveExternalAPILatency(apiName, duration)
}

// Helper function for tracking database operations
func TrackDatabaseOperation(operation, table string, duration time.Duration, success bool) {
	monitoring.DefaultMetricsCollector.ObserveDBQueryDuration(operation, table, duration)
	
	status := "success"
	if !success {
		status = "failure"
	}
	monitoring.DefaultMetricsCollector.IncrementDBTransactions(operation, status)
}

// Default middleware instances
var (
	DefaultMetricsMiddleware = NewMetricsMiddleware("openpenpal-backend")
)