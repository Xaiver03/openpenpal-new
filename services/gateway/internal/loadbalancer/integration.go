package loadbalancer

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"
)

// OpenPenPalCircuitBreakerIntegration integrates with OpenPenPal backend circuit breaker
type OpenPenPalCircuitBreakerIntegration struct {
	backendURL string
	client     *http.Client
	logger     *zap.Logger
}

// NewOpenPenPalCircuitBreakerIntegration creates a new circuit breaker integration
func NewOpenPenPalCircuitBreakerIntegration(backendURL string, logger *zap.Logger) *OpenPenPalCircuitBreakerIntegration {
	return &OpenPenPalCircuitBreakerIntegration{
		backendURL: backendURL,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		logger: logger,
	}
}

// IsServiceAvailable checks if service is available via circuit breaker
func (opci *OpenPenPalCircuitBreakerIntegration) IsServiceAvailable(serviceName string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Check circuit breaker status from OpenPenPal backend
	url := fmt.Sprintf("%s/monitoring/circuit-breakers/%s", opci.backendURL, serviceName)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		opci.logger.Error("Failed to create circuit breaker check request",
			zap.String("service", serviceName),
			zap.Error(err),
		)
		return true // Default to available on error
	}

	resp, err := opci.client.Do(req)
	if err != nil {
		opci.logger.Error("Failed to check circuit breaker status",
			zap.String("service", serviceName),
			zap.Error(err),
		)
		return true // Default to available on error
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return true // Default to available if service is down
	}

	var cbStatus struct {
		State string `json:"state"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&cbStatus); err != nil {
		opci.logger.Error("Failed to decode circuit breaker status",
			zap.String("service", serviceName),
			zap.Error(err),
		)
		return true
	}

	// Service is available if circuit breaker is not open
	return cbStatus.State != "OPEN"
}

// MarkServiceFailure marks a service failure in the circuit breaker
func (opci *OpenPenPalCircuitBreakerIntegration) MarkServiceFailure(serviceName string, instance *ServiceInstance) {
	go opci.sendCircuitBreakerEvent(serviceName, instance, false)
}

// MarkServiceSuccess marks a service success in the circuit breaker
func (opci *OpenPenPalCircuitBreakerIntegration) MarkServiceSuccess(serviceName string, instance *ServiceInstance) {
	go opci.sendCircuitBreakerEvent(serviceName, instance, true)
}

// sendCircuitBreakerEvent sends an event to the circuit breaker
func (opci *OpenPenPalCircuitBreakerIntegration) sendCircuitBreakerEvent(serviceName string, instance *ServiceInstance, success bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	eventType := "failure"
	if success {
		eventType = "success"
	}

	// This would integrate with the OpenPenPal backend circuit breaker system
	// For now, we'll log the event
	opci.logger.Debug("Circuit breaker event",
		zap.String("service", serviceName),
		zap.String("instance", instance.Host),
		zap.String("event", eventType),
	)
}

// OpenPenPalMetricsCollector integrates with OpenPenPal backend metrics
type OpenPenPalMetricsCollector struct {
	backendURL string
	client     *http.Client
	logger     *zap.Logger
}

// NewOpenPenPalMetricsCollector creates a new metrics collector integration
func NewOpenPenPalMetricsCollector(backendURL string, logger *zap.Logger) *OpenPenPalMetricsCollector {
	return &OpenPenPalMetricsCollector{
		backendURL: backendURL,
		client: &http.Client{
			Timeout: 3 * time.Second,
		},
		logger: logger,
	}
}

// RecordLoadBalancerMetrics records load balancer metrics
func (opmc *OpenPenPalMetricsCollector) RecordLoadBalancerMetrics(serviceName, algorithm string, instance *ServiceInstance, responseTime time.Duration, success bool) {
	go opmc.sendMetric(&Metric{
		Name:      "gateway_load_balancer_request",
		Type:      "counter",
		Value:     1,
		Timestamp: time.Now(),
		Tags: map[string]string{
			"service":      serviceName,
			"algorithm":    algorithm,
			"instance":     instance.Host,
			"success":      strconv.FormatBool(success),
		},
	})

	go opmc.sendMetric(&Metric{
		Name:      "gateway_load_balancer_response_time",
		Type:      "histogram",
		Value:     float64(responseTime.Milliseconds()),
		Timestamp: time.Now(),
		Tags: map[string]string{
			"service":   serviceName,
			"algorithm": algorithm,
			"instance":  instance.Host,
		},
	})
}

// RecordServiceInstanceMetrics records service instance metrics
func (opmc *OpenPenPalMetricsCollector) RecordServiceInstanceMetrics(serviceName string, instance *ServiceInstance) {
	metrics := []*Metric{
		{
			Name:      "gateway_instance_active_connections",
			Type:      "gauge",
			Value:     float64(instance.GetActiveConnections()),
			Timestamp: time.Now(),
			Tags: map[string]string{
				"service":  serviceName,
				"instance": instance.Host,
			},
		},
		{
			Name:      "gateway_instance_total_requests",
			Type:      "counter",
			Value:     float64(instance.TotalRequests),
			Timestamp: time.Now(),
			Tags: map[string]string{
				"service":  serviceName,
				"instance": instance.Host,
			},
		},
		{
			Name:      "gateway_instance_success_rate",
			Type:      "gauge",
			Value:     instance.GetSuccessRate() * 100,
			Timestamp: time.Now(),
			Tags: map[string]string{
				"service":  serviceName,
				"instance": instance.Host,
			},
		},
		{
			Name:      "gateway_instance_average_response_time",
			Type:      "gauge",
			Value:     float64(instance.AverageResponse.Milliseconds()),
			Timestamp: time.Now(),
			Tags: map[string]string{
				"service":  serviceName,
				"instance": instance.Host,
			},
		},
		{
			Name:      "gateway_instance_health_score",
			Type:      "gauge",
			Value:     instance.Score,
			Timestamp: time.Now(),
			Tags: map[string]string{
				"service":  serviceName,
				"instance": instance.Host,
			},
		},
	}

	for _, metric := range metrics {
		go opmc.sendMetric(metric)
	}
}

// Metric represents a metric to be sent
type Metric struct {
	Name      string            `json:"name"`
	Type      string            `json:"type"`
	Value     float64           `json:"value"`
	Timestamp time.Time         `json:"timestamp"`
	Tags      map[string]string `json:"tags"`
}

// sendMetric sends a metric to the OpenPenPal backend
func (opmc *OpenPenPalMetricsCollector) sendMetric(metric *Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Convert metric to format expected by OpenPenPal backend
	metricData := map[string]interface{}{
		"name":      metric.Name,
		"type":      metric.Type,
		"value":     metric.Value,
		"timestamp": metric.Timestamp.Unix(),
		"tags":      metric.Tags,
	}

	jsonData, err := json.Marshal(metricData)
	if err != nil {
		opmc.logger.Error("Failed to marshal metric", zap.Error(err))
		return
	}

	// In a real implementation, this would send to the OpenPenPal backend metrics endpoint
	// For now, we'll log it
	opmc.logger.Debug("Sending metric to OpenPenPal backend",
		zap.String("metric", string(jsonData)),
	)
}

// AdvancedProxyManager extends the basic proxy manager with load balancing
type AdvancedProxyManager struct {
	loadBalancerManager *LoadBalancerManager
	logger              *zap.Logger
}

// NewAdvancedProxyManager creates a new advanced proxy manager
func NewAdvancedProxyManager(lbm *LoadBalancerManager, logger *zap.Logger) *AdvancedProxyManager {
	return &AdvancedProxyManager{
		loadBalancerManager: lbm,
		logger:              logger,
	}
}

// SelectInstanceForRequest selects the best instance for a request using advanced load balancing
func (apm *AdvancedProxyManager) SelectInstanceForRequest(serviceName, clientIP, userAgent string, headers map[string]string) (*ServiceInstance, error) {
	// Extract session ID from headers if available
	sessionID := apm.extractSessionID(headers)
	
	// Add request fingerprinting for consistent hashing
	requestFingerprint := apm.generateRequestFingerprint(clientIP, userAgent, sessionID)
	
	// Select instance using load balancer
	instance, err := apm.loadBalancerManager.SelectInstance(serviceName, requestFingerprint)
	if err != nil {
		apm.logger.Error("Failed to select instance",
			zap.String("service", serviceName),
			zap.String("client_ip", clientIP),
			zap.Error(err),
		)
		return nil, err
	}
	
	apm.logger.Debug("Selected instance for request",
		zap.String("service", serviceName),
		zap.String("instance", instance.Host),
		zap.String("client_ip", clientIP),
		zap.String("session_id", sessionID),
	)
	
	return instance, nil
}

// UpdateRequestStats updates statistics after a request is completed
func (apm *AdvancedProxyManager) UpdateRequestStats(serviceName string, instance *ServiceInstance, responseTime time.Duration, statusCode int) {
	success := statusCode < 400
	
	apm.loadBalancerManager.UpdateInstanceStats(serviceName, instance, responseTime, success)
	
	apm.logger.Debug("Updated request stats",
		zap.String("service", serviceName),
		zap.String("instance", instance.Host),
		zap.Duration("response_time", responseTime),
		zap.Int("status_code", statusCode),
		zap.Bool("success", success),
	)
}

// extractSessionID extracts session ID from request headers
func (apm *AdvancedProxyManager) extractSessionID(headers map[string]string) string {
	// Try common session header names
	sessionHeaders := []string{
		"X-Session-ID",
		"Session-ID",
		"Authorization", // For JWT tokens
		"Cookie",        // For session cookies
	}
	
	for _, header := range sessionHeaders {
		if value, exists := headers[header]; exists && value != "" {
			// For Authorization header, extract session info from JWT
			if header == "Authorization" && len(value) > 7 && value[:7] == "Bearer " {
				return apm.extractJWTSessionID(value[7:])
			}
			// For Cookie header, extract session cookie
			if header == "Cookie" {
				return apm.extractSessionCookie(value)
			}
			return value
		}
	}
	
	return ""
}

// generateRequestFingerprint generates a fingerprint for request routing
func (apm *AdvancedProxyManager) generateRequestFingerprint(clientIP, userAgent, sessionID string) string {
	if sessionID != "" {
		return sessionID
	}
	
	// Create fingerprint from client IP and user agent
	return fmt.Sprintf("%s-%s", clientIP, userAgent)
}

// extractJWTSessionID extracts session ID from JWT token
func (apm *AdvancedProxyManager) extractJWTSessionID(token string) string {
	// In a real implementation, this would decode the JWT and extract user/session info
	// For now, return a hash of the token
	return fmt.Sprintf("jwt-%x", token[:min(len(token), 8)])
}

// extractSessionCookie extracts session ID from cookie header
func (apm *AdvancedProxyManager) extractSessionCookie(cookieHeader string) string {
	// Parse cookies and find session cookie
	// This is a simplified implementation
	if len(cookieHeader) > 20 {
		return fmt.Sprintf("cookie-%x", cookieHeader[:min(len(cookieHeader), 8)])
	}
	return ""
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// HealthAwareRoutingConfig configures health-aware routing
type HealthAwareRoutingConfig struct {
	HealthThreshold     float64       `json:"health_threshold"`
	UnhealthyPenalty    float64       `json:"unhealthy_penalty"`
	RecoveryGracePeriod time.Duration `json:"recovery_grace_period"`
	HealthCheckInterval time.Duration `json:"health_check_interval"`
}

// PerformanceRoutingConfig configures performance-based routing
type PerformanceRoutingConfig struct {
	ResponseTimeWeight     float64       `json:"response_time_weight"`
	ThroughputWeight      float64       `json:"throughput_weight"`
	ErrorRateWeight       float64       `json:"error_rate_weight"`
	PerformanceWindow     time.Duration `json:"performance_window"`
	AdaptationRate        float64       `json:"adaptation_rate"`
}

// GeographicRoutingConfig configures geographic routing (future enhancement)
type GeographicRoutingConfig struct {
	Enabled           bool                       `json:"enabled"`
	RegionPreferences map[string][]string       `json:"region_preferences"`
	LatencyThresholds map[string]time.Duration  `json:"latency_thresholds"`
}

// ComprehensiveLoadBalancerConfig extends the basic config with advanced features
type ComprehensiveLoadBalancerConfig struct {
	*LoadBalancerConfig
	HealthAwareRouting   *HealthAwareRoutingConfig   `json:"health_aware_routing"`
	PerformanceRouting   *PerformanceRoutingConfig   `json:"performance_routing"`
	GeographicRouting    *GeographicRoutingConfig    `json:"geographic_routing"`
}

// CreateComprehensiveConfig creates a comprehensive load balancer configuration
func CreateComprehensiveConfig() *ComprehensiveLoadBalancerConfig {
	return &ComprehensiveLoadBalancerConfig{
		LoadBalancerConfig: DefaultLoadBalancerConfig(),
		HealthAwareRouting: &HealthAwareRoutingConfig{
			HealthThreshold:     0.8,
			UnhealthyPenalty:    0.5,
			RecoveryGracePeriod: 2 * time.Minute,
			HealthCheckInterval: 30 * time.Second,
		},
		PerformanceRouting: &PerformanceRoutingConfig{
			ResponseTimeWeight: 0.4,
			ThroughputWeight:   0.3,
			ErrorRateWeight:    0.3,
			PerformanceWindow:  5 * time.Minute,
			AdaptationRate:     0.1,
		},
		GeographicRouting: &GeographicRoutingConfig{
			Enabled:           false, // Disabled by default
			RegionPreferences: make(map[string][]string),
			LatencyThresholds: make(map[string]time.Duration),
		},
	}
}