package proxy

import (
	"api-gateway/internal/discovery"
	"api-gateway/internal/loadbalancer"
	"api-gateway/internal/models"
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// EnhancedProxyManager extends the basic proxy manager with advanced load balancing
type EnhancedProxyManager struct {
	*Manager // Embed the original manager
	loadBalancerManager  *loadbalancer.LoadBalancerManager
	advancedProxyManager *loadbalancer.AdvancedProxyManager
	logger               *zap.Logger
	config               *EnhancedProxyConfig
}

// EnhancedProxyConfig holds configuration for enhanced proxy features
type EnhancedProxyConfig struct {
	// Load balancing
	LoadBalancingEnabled bool   `json:"load_balancing_enabled"`
	DefaultAlgorithm     string `json:"default_algorithm"`
	
	// Circuit breaker integration
	CircuitBreakerEnabled     bool   `json:"circuit_breaker_enabled"`
	CircuitBreakerBackendURL  string `json:"circuit_breaker_backend_url"`
	
	// Metrics integration
	MetricsEnabled    bool   `json:"metrics_enabled"`
	MetricsBackendURL string `json:"metrics_backend_url"`
	
	// Retry configuration
	RetryEnabled     bool          `json:"retry_enabled"`
	MaxRetries       int           `json:"max_retries"`
	RetryDelay       time.Duration `json:"retry_delay"`
	RetryBackoff     float64       `json:"retry_backoff"`
	
	// Timeout configuration
	RequestTimeout   time.Duration `json:"request_timeout"`
	ResponseTimeout  time.Duration `json:"response_timeout"`
	
	// Rate limiting integration
	RateLimitingEnabled bool `json:"rate_limiting_enabled"`
	
	// Request tracing
	TracingEnabled bool `json:"tracing_enabled"`
}

// NewEnhancedProxyManager creates a new enhanced proxy manager
func NewEnhancedProxyManager(
	serviceDiscovery *discovery.ServiceDiscovery,
	loadBalancerManager *loadbalancer.LoadBalancerManager,
	logger *zap.Logger,
	config *EnhancedProxyConfig,
) *EnhancedProxyManager {
	// Create the basic manager
	basicManager := NewManager(serviceDiscovery, logger)
	
	// Create advanced proxy manager
	advancedProxyManager := loadbalancer.NewAdvancedProxyManager(loadBalancerManager, logger)
	
	return &EnhancedProxyManager{
		Manager:              basicManager,
		loadBalancerManager:  loadBalancerManager,
		advancedProxyManager: advancedProxyManager,
		logger:               logger,
		config:               config,
	}
}

// EnhancedProxyHandler creates an enhanced proxy handler with advanced load balancing
func (epm *EnhancedProxyManager) EnhancedProxyHandler(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		
		// Extract request information for load balancing
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()
		headers := epm.extractHeaders(c)
		
		epm.logger.Debug("Processing request with enhanced proxy",
			zap.String("service", serviceName),
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method),
			zap.String("client_ip", clientIP),
		)
		
		// Check rate limiting if enabled
		if epm.config.RateLimitingEnabled {
			if limited := epm.checkRateLimit(c, serviceName); limited {
				return
			}
		}
		
		// Select instance using advanced load balancing
		var instance *loadbalancer.ServiceInstance
		var err error
		
		if epm.config.LoadBalancingEnabled {
			instance, err = epm.advancedProxyManager.SelectInstanceForRequest(serviceName, clientIP, userAgent, headers)
			if err != nil {
				epm.handleServiceUnavailable(c, serviceName, err)
				return
			}
		} else {
			// Fallback to basic service discovery
			basicInstance, err := epm.serviceDiscovery.GetHealthyInstance(serviceName)
			if err != nil {
				epm.handleServiceUnavailable(c, serviceName, err)
				return
			}
			instance = loadbalancer.WrapServiceInstance(basicInstance)
		}
		
		// Setup request context and tracing
		ctx := epm.setupRequestContext(c, serviceName, instance)
		c.Request = c.Request.WithContext(ctx)
		
		// Execute request with retry logic
		success := epm.executeRequestWithRetry(c, serviceName, instance, startTime)
		
		// Update statistics
		duration := time.Since(startTime)
		statusCode := c.Writer.Status()
		
		if epm.config.LoadBalancingEnabled {
			epm.advancedProxyManager.UpdateRequestStats(serviceName, instance, duration, statusCode)
		}
		
		// Log request completion
		epm.logRequestCompletion(c, serviceName, instance, duration, success)
	}
}

// executeRequestWithRetry executes a request with retry logic
func (epm *EnhancedProxyManager) executeRequestWithRetry(c *gin.Context, serviceName string, instance *loadbalancer.ServiceInstance, startTime time.Time) bool {
	var lastErr error
	maxRetries := 1
	
	if epm.config.RetryEnabled {
		maxRetries = epm.config.MaxRetries + 1
	}
	
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			// Apply retry delay with backoff
			delay := time.Duration(float64(epm.config.RetryDelay) * epm.calculateBackoff(attempt))
			time.Sleep(delay)
			
			epm.logger.Debug("Retrying request",
				zap.String("service", serviceName),
				zap.String("instance", instance.Host),
				zap.Int("attempt", attempt+1),
				zap.Duration("delay", delay),
			)
			
			// Try to get a different instance for retry
			if epm.config.LoadBalancingEnabled {
				newInstance, err := epm.advancedProxyManager.SelectInstanceForRequest(
					serviceName,
					c.ClientIP(),
					c.Request.UserAgent(),
					epm.extractHeaders(c),
				)
				if err == nil {
					instance = newInstance
				}
			}
		}
		
		// Create proxy for this instance
		proxy := epm.createEnhancedProxy(serviceName, instance)
		
		// Setup response capture
		responseCapture := &responseCapture{ResponseWriter: c.Writer}
		c.Writer = responseCapture
		
		// Execute proxy request
		proxy.ServeHTTP(c.Writer, c.Request)
		
		// Check if request was successful
		if responseCapture.statusCode < 500 {
			return responseCapture.statusCode < 400
		}
		
		lastErr = fmt.Errorf("server error: %d", responseCapture.statusCode)
		
		// Mark instance as having an error for load balancing
		if epm.config.LoadBalancingEnabled {
			duration := time.Since(startTime)
			epm.advancedProxyManager.UpdateRequestStats(serviceName, instance, duration, responseCapture.statusCode)
		}
	}
	
	// All retries failed
	epm.logger.Error("All retry attempts failed",
		zap.String("service", serviceName),
		zap.String("instance", instance.Host),
		zap.Int("attempts", maxRetries),
		zap.Error(lastErr),
	)
	
	return false
}

// createEnhancedProxy creates an enhanced reverse proxy for an instance
func (epm *EnhancedProxyManager) createEnhancedProxy(serviceName string, instance *loadbalancer.ServiceInstance) *httputil.ReverseProxy {
	target, _ := url.Parse(instance.Host)
	proxy := httputil.NewSingleHostReverseProxy(target)
	
	// Enhanced Director function
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		epm.enhancedModifyRequest(req, serviceName, instance)
	}
	
	// Enhanced Error Handler
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		epm.enhancedHandleProxyError(w, r, serviceName, instance, err)
	}
	
	// Enhanced Response Modifier
	proxy.ModifyResponse = func(resp *http.Response) error {
		return epm.enhancedModifyResponse(resp, serviceName, instance)
	}
	
	return proxy
}

// enhancedModifyRequest modifies the request with enhanced features
func (epm *EnhancedProxyManager) enhancedModifyRequest(req *http.Request, serviceName string, instance *loadbalancer.ServiceInstance) {
	// Apply original modifications
	epm.modifyRequest(req, serviceName)
	
	// Add load balancing information
	req.Header.Set("X-LB-Algorithm", epm.loadBalancerManager.GetLoadBalancerStats()[serviceName].(map[string]interface{})["algorithm"].(string))
	req.Header.Set("X-LB-Instance-Score", fmt.Sprintf("%.2f", instance.Score))
	req.Header.Set("X-LB-Instance-Connections", strconv.FormatInt(instance.GetActiveConnections(), 10))
	
	// Add circuit breaker information
	if epm.config.CircuitBreakerEnabled {
		req.Header.Set("X-CB-Enabled", "true")
	}
	
	// Add performance metrics
	req.Header.Set("X-Instance-Response-Time", instance.AverageResponse.String())
	req.Header.Set("X-Instance-Success-Rate", fmt.Sprintf("%.2f", instance.GetSuccessRate()*100))
	
	// Add request timing for timeout handling
	if epm.config.RequestTimeout > 0 {
		ctx, cancel := context.WithTimeout(req.Context(), epm.config.RequestTimeout)
		req = req.WithContext(ctx)
		
		// Store cancel function for cleanup
		req.Header.Set("X-Request-Timeout", epm.config.RequestTimeout.String())
		
		// Note: In production, you'd need to manage the cancel function properly
		_ = cancel
	}
}

// enhancedModifyResponse modifies the response with enhanced features
func (epm *EnhancedProxyManager) enhancedModifyResponse(resp *http.Response, serviceName string, instance *loadbalancer.ServiceInstance) error {
	// Apply original modifications
	err := epm.modifyResponse(resp, serviceName)
	if err != nil {
		return err
	}
	
	// Add enhanced headers
	resp.Header.Set("X-LB-Instance", instance.Host)
	resp.Header.Set("X-LB-Instance-Score", fmt.Sprintf("%.2f", instance.Score))
	resp.Header.Set("X-Response-Time", instance.LastResponseTime.String())
	
	// Add performance information
	resp.Header.Set("X-Instance-Load", strconv.FormatInt(instance.GetActiveConnections(), 10))
	resp.Header.Set("X-Instance-Success-Rate", fmt.Sprintf("%.2f%%", instance.GetSuccessRate()*100))
	
	// Add circuit breaker status
	if epm.config.CircuitBreakerEnabled {
		resp.Header.Set("X-CB-Status", "enabled")
	}
	
	return nil
}

// enhancedHandleProxyError handles proxy errors with enhanced features
func (epm *EnhancedProxyManager) enhancedHandleProxyError(w http.ResponseWriter, r *http.Request, serviceName string, instance *loadbalancer.ServiceInstance, err error) {
	epm.logger.Error("Enhanced proxy error",
		zap.String("service", serviceName),
		zap.String("instance", instance.Host),
		zap.String("path", r.URL.Path),
		zap.Error(err),
	)
	
	// Mark instance as unhealthy in service discovery
	epm.serviceDiscovery.MarkUnhealthy(serviceName, instance.Host)
	
	// Update load balancer statistics
	if epm.config.LoadBalancingEnabled {
		epm.advancedProxyManager.UpdateRequestStats(serviceName, instance, 0, 502)
	}
	
	// Enhanced error response
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Error-Source", "gateway")
	w.Header().Set("X-Failed-Instance", instance.Host)
	w.WriteHeader(http.StatusBadGateway)
	
	errorResponse := models.ErrorResponse{
		Code:      http.StatusBadGateway,
		Message:   "Service temporarily unavailable",
		Details:   "Instance failed: " + instance.Host,
		Timestamp: time.Now(),
		Path:      r.URL.Path,
	}
	
	if respBytes, err := errorResponse.ToJSON(); err == nil {
		w.Write(respBytes)
	}
}

// Helper methods

// extractHeaders extracts relevant headers for load balancing
func (epm *EnhancedProxyManager) extractHeaders(c *gin.Context) map[string]string {
	headers := make(map[string]string)
	
	relevantHeaders := []string{
		"Authorization",
		"X-Session-ID",
		"Session-ID",
		"Cookie",
		"X-User-ID",
		"X-Tenant-ID",
	}
	
	for _, header := range relevantHeaders {
		if value := c.GetHeader(header); value != "" {
			headers[header] = value
		}
	}
	
	return headers
}

// setupRequestContext sets up request context with tracing and timeout
func (epm *EnhancedProxyManager) setupRequestContext(c *gin.Context, serviceName string, instance *loadbalancer.ServiceInstance) context.Context {
	ctx := c.Request.Context()
	
	// Add service and instance information to context
	ctx = context.WithValue(ctx, "service_name", serviceName)
	ctx = context.WithValue(ctx, "instance_host", instance.Host)
	ctx = context.WithValue(ctx, "instance_score", instance.Score)
	
	// Add tracing information if enabled
	if epm.config.TracingEnabled {
		traceID := c.GetHeader("X-Trace-ID")
		if traceID == "" {
			traceID = epm.generateTraceID()
			c.Header("X-Trace-ID", traceID)
		}
		ctx = context.WithValue(ctx, "trace_id", traceID)
	}
	
	return ctx
}

// checkRateLimit checks rate limiting for the request
func (epm *EnhancedProxyManager) checkRateLimit(c *gin.Context, serviceName string) bool {
	// This would integrate with a rate limiting service
	// For now, we'll just log and allow all requests
	epm.logger.Debug("Rate limit check",
		zap.String("service", serviceName),
		zap.String("client_ip", c.ClientIP()),
	)
	return false
}

// handleServiceUnavailable handles service unavailable errors
func (epm *EnhancedProxyManager) handleServiceUnavailable(c *gin.Context, serviceName string, err error) {
	epm.logger.Error("Service unavailable",
		zap.String("service", serviceName),
		zap.Error(err),
	)
	
	c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
		Code:      http.StatusServiceUnavailable,
		Message:   "Service temporarily unavailable",
		Details:   err.Error(),
		Timestamp: time.Now(),
		Path:      c.Request.URL.Path,
	})
}

// calculateBackoff calculates backoff delay for retries
func (epm *EnhancedProxyManager) calculateBackoff(attempt int) float64 {
	if epm.config.RetryBackoff <= 1.0 {
		return 1.0
	}
	
	backoff := 1.0
	for i := 0; i < attempt; i++ {
		backoff *= epm.config.RetryBackoff
	}
	
	return backoff
}

// logRequestCompletion logs request completion with enhanced information
func (epm *EnhancedProxyManager) logRequestCompletion(c *gin.Context, serviceName string, instance *loadbalancer.ServiceInstance, duration time.Duration, success bool) {
	epm.logger.Info("Enhanced proxy request completed",
		zap.String("service", serviceName),
		zap.String("instance", instance.Host),
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.Int("status", c.Writer.Status()),
		zap.Duration("duration", duration),
		zap.String("client_ip", c.ClientIP()),
		zap.Bool("success", success),
		zap.Float64("instance_score", instance.Score),
		zap.Int64("instance_connections", instance.GetActiveConnections()),
		zap.Float64("instance_success_rate", instance.GetSuccessRate()),
		zap.String("trace_id", c.GetHeader("X-Trace-ID")),
	)
}

// responseCapture captures response information for retry logic
type responseCapture struct {
	gin.ResponseWriter
	statusCode int
	written    bool
}

func (rc *responseCapture) WriteHeader(code int) {
	if !rc.written {
		rc.statusCode = code
		rc.written = true
		rc.ResponseWriter.WriteHeader(code)
	}
}

func (rc *responseCapture) Write(data []byte) (int, error) {
	if !rc.written {
		rc.statusCode = http.StatusOK
		rc.written = true
	}
	return rc.ResponseWriter.Write(data)
}

// DefaultEnhancedProxyConfig returns a default enhanced proxy configuration
func DefaultEnhancedProxyConfig() *EnhancedProxyConfig {
	return &EnhancedProxyConfig{
		LoadBalancingEnabled:      true,
		DefaultAlgorithm:          "adaptive",
		CircuitBreakerEnabled:     true,
		CircuitBreakerBackendURL:  "http://localhost:8080",
		MetricsEnabled:            true,
		MetricsBackendURL:         "http://localhost:8080",
		RetryEnabled:              true,
		MaxRetries:                2,
		RetryDelay:                100 * time.Millisecond,
		RetryBackoff:              2.0,
		RequestTimeout:            30 * time.Second,
		ResponseTimeout:           30 * time.Second,
		RateLimitingEnabled:       false,
		TracingEnabled:            true,
	}
}