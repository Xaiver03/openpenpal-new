// Package servicemesh provides real-time health monitoring capabilities
package servicemesh

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// RealTimeHealthMonitor provides comprehensive health monitoring for services
type RealTimeHealthMonitor struct {
	// Health check configurations per service
	healthChecks map[string]*HealthCheckConfig
	
	// Current health states per service
	healthStates map[string]*HealthState
	
	// Health check strategies
	strategies map[string]HealthCheckStrategy
	
	// Health event subscribers
	subscribers map[string][]HealthEventSubscriber
	
	// Metrics collector
	metricsCollector *HealthMetricsCollector
	
	// Alert manager
	alertManager *HealthAlertManager
	
	// Configuration
	config *HealthMonitorConfig
	
	// HTTP client for health checks
	httpClient *http.Client
	
	// Mutex for thread safety
	mu sync.RWMutex
	
	// Running state
	running bool
}

// HealthCheckConfig defines health check configuration for a service
type HealthCheckConfig struct {
	ServiceID       string            `json:"service_id"`
	Enabled         bool              `json:"enabled"`
	Strategy        string            `json:"strategy"`
	Interval        time.Duration     `json:"interval"`
	Timeout         time.Duration     `json:"timeout"`
	Retries         int               `json:"retries"`
	RetryDelay      time.Duration     `json:"retry_delay"`
	
	// HTTP health check configuration
	HTTPConfig      *HTTPHealthConfig `json:"http_config,omitempty"`
	
	// TCP health check configuration
	TCPConfig       *TCPHealthConfig  `json:"tcp_config,omitempty"`
	
	// Custom health check configuration
	CustomConfig    *CustomHealthConfig `json:"custom_config,omitempty"`
	
	// Thresholds
	HealthyThreshold   int `json:"healthy_threshold"`
	UnhealthyThreshold int `json:"unhealthy_threshold"`
	
	// Advanced settings
	GracePeriod        time.Duration     `json:"grace_period"`
	SuccessBeforeHealthy int            `json:"success_before_healthy"`
	FailuresBeforeUnhealthy int         `json:"failures_before_unhealthy"`
	
	// Metadata
	Tags               map[string]string `json:"tags"`
	LastUpdated        time.Time         `json:"last_updated"`
}

// HTTPHealthConfig defines HTTP-based health check configuration
type HTTPHealthConfig struct {
	URL                string            `json:"url"`
	Method             string            `json:"method"`
	Headers            map[string]string `json:"headers"`
	Body               string            `json:"body"`
	ExpectedStatusCode int               `json:"expected_status_code"`
	ExpectedBody       string            `json:"expected_body"`
	ValidateSSL        bool              `json:"validate_ssl"`
}

// TCPHealthConfig defines TCP-based health check configuration
type TCPHealthConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

// CustomHealthConfig defines custom health check configuration
type CustomHealthConfig struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env"`
}

// HealthState represents the current health state of a service
type HealthState struct {
	ServiceID           string                 `json:"service_id"`
	Status              HealthStatus           `json:"status"`
	LastCheckTime       time.Time              `json:"last_check_time"`
	LastStatusChange    time.Time              `json:"last_status_change"`
	ConsecutiveSuccesses int                   `json:"consecutive_successes"`
	ConsecutiveFailures  int                   `json:"consecutive_failures"`
	TotalChecks         int64                  `json:"total_checks"`
	SuccessfulChecks    int64                  `json:"successful_checks"`
	FailedChecks        int64                  `json:"failed_checks"`
	
	// Recent check results
	RecentResults       []HealthCheckResult    `json:"recent_results"`
	
	// Performance metrics
	AverageResponseTime float64                `json:"average_response_time"`
	P95ResponseTime     float64                `json:"p95_response_time"`
	P99ResponseTime     float64                `json:"p99_response_time"`
	UpTimePercentage    float64                `json:"uptime_percentage"`
	
	// Detailed status information
	StatusDetails       map[string]interface{} `json:"status_details"`
	LastError           string                 `json:"last_error"`
	
	// Health score (0-1)
	HealthScore         float64                `json:"health_score"`
}

// HealthCheckResult represents the result of a health check
type HealthCheckResult struct {
	Timestamp    time.Time              `json:"timestamp"`
	Success      bool                   `json:"success"`
	ResponseTime float64                `json:"response_time"`
	StatusCode   int                    `json:"status_code"`
	Message      string                 `json:"message"`
	Error        string                 `json:"error"`
	Details      map[string]interface{} `json:"details"`
}

// HealthMonitorConfig holds global health monitor configuration
type HealthMonitorConfig struct {
	DefaultInterval                time.Duration `json:"default_interval"`
	DefaultTimeout                 time.Duration `json:"default_timeout"`
	DefaultRetries                 int           `json:"default_retries"`
	DefaultRetryDelay              time.Duration `json:"default_retry_delay"`
	DefaultHealthyThreshold        int           `json:"default_healthy_threshold"`
	DefaultUnhealthyThreshold      int           `json:"default_unhealthy_threshold"`
	DefaultSuccessBeforeHealthy    int           `json:"default_success_before_healthy"`
	DefaultFailuresBeforeUnhealthy int           `json:"default_failures_before_unhealthy"`
	MaxConcurrentChecks            int           `json:"max_concurrent_checks"`
	MetricsRetentionPeriod         time.Duration `json:"metrics_retention_period"`
	AlertCooldownPeriod            time.Duration `json:"alert_cooldown_period"`
	EnableMetricsCollection        bool          `json:"enable_metrics_collection"`
	EnableAlerting                 bool          `json:"enable_alerting"`
}

// HealthCheckStrategy defines an interface for health check strategies
type HealthCheckStrategy interface {
	Check(ctx context.Context, config *HealthCheckConfig, state *HealthState) (*HealthCheckResult, error)
	Name() string
}

// HealthEventSubscriber defines an interface for health event subscribers
type HealthEventSubscriber interface {
	OnHealthChange(serviceID string, oldStatus, newStatus HealthStatus, details map[string]interface{})
}

// HealthMetricsCollector collects and aggregates health metrics
type HealthMetricsCollector struct {
	mu      sync.RWMutex
	metrics map[string]*HealthMetrics
}

// HealthMetrics holds aggregated health metrics for a service
type HealthMetrics struct {
	ServiceID              string    `json:"service_id"`
	TotalChecks            int64     `json:"total_checks"`
	SuccessfulChecks       int64     `json:"successful_checks"`
	FailedChecks           int64     `json:"failed_checks"`
	AverageResponseTime    float64   `json:"average_response_time"`
	P95ResponseTime        float64   `json:"p95_response_time"`
	P99ResponseTime        float64   `json:"p99_response_time"`
	UptimePercentage       float64   `json:"uptime_percentage"`
	DowntimeMinutes        float64   `json:"downtime_minutes"`
	MeanTimeBetweenFailures float64  `json:"mean_time_between_failures"`
	MeanTimeToRecover      float64   `json:"mean_time_to_recover"`
	LastUpdated            time.Time `json:"last_updated"`
}

// HealthAlertManager manages health-related alerts
type HealthAlertManager struct {
	mu                sync.RWMutex
	alertRules        map[string]*AlertRule
	activeAlerts      map[string]*ActiveAlert
	alertHistory      []AlertEvent
	lastAlertTime     map[string]time.Time
	cooldownPeriod    time.Duration
}

// AlertRule defines when to trigger alerts
type AlertRule struct {
	ID                  string        `json:"id"`
	ServiceID           string        `json:"service_id"`
	Condition           string        `json:"condition"`
	Threshold           float64       `json:"threshold"`
	Duration            time.Duration `json:"duration"`
	Severity            string        `json:"severity"`
	Message             string        `json:"message"`
	Enabled             bool          `json:"enabled"`
	LastTriggered       time.Time     `json:"last_triggered"`
}

// ActiveAlert represents an active alert
type ActiveAlert struct {
	ID          string                 `json:"id"`
	RuleID      string                 `json:"rule_id"`
	ServiceID   string                 `json:"service_id"`
	Severity    string                 `json:"severity"`
	Message     string                 `json:"message"`
	StartTime   time.Time              `json:"start_time"`
	LastUpdate  time.Time              `json:"last_update"`
	Details     map[string]interface{} `json:"details"`
}

// AlertEvent represents a historical alert event
type AlertEvent struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"` // "triggered", "resolved"
	RuleID    string                 `json:"rule_id"`
	ServiceID string                 `json:"service_id"`
	Severity  string                 `json:"severity"`
	Message   string                 `json:"message"`
	Timestamp time.Time              `json:"timestamp"`
	Details   map[string]interface{} `json:"details"`
}

// NewRealTimeHealthMonitor creates a new real-time health monitor
func NewRealTimeHealthMonitor() *RealTimeHealthMonitor {
	config := &HealthMonitorConfig{
		DefaultInterval:                30 * time.Second,
		DefaultTimeout:                 5 * time.Second,
		DefaultRetries:                 3,
		DefaultRetryDelay:              1 * time.Second,
		DefaultHealthyThreshold:        2,
		DefaultUnhealthyThreshold:      3,
		DefaultSuccessBeforeHealthy:    2,
		DefaultFailuresBeforeUnhealthy: 3,
		MaxConcurrentChecks:            50,
		MetricsRetentionPeriod:         24 * time.Hour,
		AlertCooldownPeriod:            5 * time.Minute,
		EnableMetricsCollection:        true,
		EnableAlerting:                 true,
	}

	monitor := &RealTimeHealthMonitor{
		healthChecks:     make(map[string]*HealthCheckConfig),
		healthStates:     make(map[string]*HealthState),
		strategies:       make(map[string]HealthCheckStrategy),
		subscribers:      make(map[string][]HealthEventSubscriber),
		metricsCollector: NewHealthMetricsCollector(),
		alertManager:     NewHealthAlertManager(config.AlertCooldownPeriod),
		config:           config,
		httpClient: &http.Client{
			Timeout: config.DefaultTimeout,
		},
	}

	// Register built-in strategies
	monitor.registerStrategies()

	return monitor
}

// NewHealthMetricsCollector creates a new health metrics collector
func NewHealthMetricsCollector() *HealthMetricsCollector {
	return &HealthMetricsCollector{
		metrics: make(map[string]*HealthMetrics),
	}
}

// NewHealthAlertManager creates a new health alert manager
func NewHealthAlertManager(cooldownPeriod time.Duration) *HealthAlertManager {
	return &HealthAlertManager{
		alertRules:     make(map[string]*AlertRule),
		activeAlerts:   make(map[string]*ActiveAlert),
		alertHistory:   make([]AlertEvent, 0),
		lastAlertTime:  make(map[string]time.Time),
		cooldownPeriod: cooldownPeriod,
	}
}

// registerStrategies registers built-in health check strategies
func (rhm *RealTimeHealthMonitor) registerStrategies() {
	rhm.strategies["http"] = &HTTPHealthCheckStrategy{client: rhm.httpClient}
	rhm.strategies["tcp"] = &TCPHealthCheckStrategy{}
	rhm.strategies["custom"] = &CustomHealthCheckStrategy{}
}

// Start starts the real-time health monitor
func (rhm *RealTimeHealthMonitor) Start(ctx context.Context) error {
	rhm.mu.Lock()
	defer rhm.mu.Unlock()

	if rhm.running {
		return fmt.Errorf("health monitor is already running")
	}

	log.Println("🏥 Starting Real-Time Health Monitor")

	// Start health check loops for each registered service
	for serviceID := range rhm.healthChecks {
		go rhm.healthCheckLoop(ctx, serviceID)
	}

	// Start metrics collection loop
	if rhm.config.EnableMetricsCollection {
		go rhm.metricsCollectionLoop(ctx)
	}

	// Start alert management loop
	if rhm.config.EnableAlerting {
		go rhm.alertManagementLoop(ctx)
	}

	rhm.running = true
	log.Println("✅ Real-Time Health Monitor started")

	return nil
}

// RegisterService registers a service for health monitoring
func (rhm *RealTimeHealthMonitor) RegisterService(serviceID string, config *HealthCheckConfig) error {
	rhm.mu.Lock()
	defer rhm.mu.Unlock()

	// Set defaults if not provided
	if config.Interval == 0 {
		config.Interval = rhm.config.DefaultInterval
	}
	if config.Timeout == 0 {
		config.Timeout = rhm.config.DefaultTimeout
	}
	if config.Retries == 0 {
		config.Retries = rhm.config.DefaultRetries
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = rhm.config.DefaultRetryDelay
	}
	if config.HealthyThreshold == 0 {
		config.HealthyThreshold = rhm.config.DefaultHealthyThreshold
	}
	if config.UnhealthyThreshold == 0 {
		config.UnhealthyThreshold = rhm.config.DefaultUnhealthyThreshold
	}
	if config.SuccessBeforeHealthy == 0 {
		config.SuccessBeforeHealthy = rhm.config.DefaultSuccessBeforeHealthy
	}
	if config.FailuresBeforeUnhealthy == 0 {
		config.FailuresBeforeUnhealthy = rhm.config.DefaultFailuresBeforeUnhealthy
	}

	config.ServiceID = serviceID
	config.LastUpdated = time.Now()
	config.Enabled = true

	// Create initial health state
	state := &HealthState{
		ServiceID:        serviceID,
		Status:           HealthStatusUnknown,
		LastCheckTime:    time.Time{},
		LastStatusChange: time.Now(),
		RecentResults:    make([]HealthCheckResult, 0),
		StatusDetails:    make(map[string]interface{}),
		HealthScore:      0.5, // Neutral score initially
	}

	rhm.healthChecks[serviceID] = config
	rhm.healthStates[serviceID] = state

	// Initialize metrics
	rhm.metricsCollector.InitializeService(serviceID)

	// Start health check loop if monitor is running
	if rhm.running {
		go rhm.healthCheckLoop(context.Background(), serviceID)
	}

	log.Printf("🏥 Registered health monitoring for service: %s", serviceID)

	return nil
}

// DeregisterService removes a service from health monitoring
func (rhm *RealTimeHealthMonitor) DeregisterService(serviceID string) error {
	rhm.mu.Lock()
	defer rhm.mu.Unlock()

	delete(rhm.healthChecks, serviceID)
	delete(rhm.healthStates, serviceID)

	log.Printf("🏥 Deregistered health monitoring for service: %s", serviceID)

	return nil
}

// GetHealthStatus returns the current health status of a service
func (rhm *RealTimeHealthMonitor) GetHealthStatus(serviceID string) (*HealthState, error) {
	rhm.mu.RLock()
	defer rhm.mu.RUnlock()

	state, exists := rhm.healthStates[serviceID]
	if !exists {
		return nil, fmt.Errorf("service not found: %s", serviceID)
	}

	// Return a copy to avoid race conditions
	stateCopy := *state
	return &stateCopy, nil
}

// GetAllHealthStates returns health states for all monitored services
func (rhm *RealTimeHealthMonitor) GetAllHealthStates() map[string]*HealthState {
	rhm.mu.RLock()
	defer rhm.mu.RUnlock()

	result := make(map[string]*HealthState)
	for serviceID, state := range rhm.healthStates {
		stateCopy := *state
		result[serviceID] = &stateCopy
	}

	return result
}

// SubscribeToHealthEvents subscribes to health change events for a service
func (rhm *RealTimeHealthMonitor) SubscribeToHealthEvents(serviceID string, subscriber HealthEventSubscriber) {
	rhm.mu.Lock()
	defer rhm.mu.Unlock()

	if rhm.subscribers[serviceID] == nil {
		rhm.subscribers[serviceID] = make([]HealthEventSubscriber, 0)
	}

	rhm.subscribers[serviceID] = append(rhm.subscribers[serviceID], subscriber)
}

// healthCheckLoop runs the health check loop for a specific service
func (rhm *RealTimeHealthMonitor) healthCheckLoop(ctx context.Context, serviceID string) {
	rhm.mu.RLock()
	config, configExists := rhm.healthChecks[serviceID]
	rhm.mu.RUnlock()

	if !configExists || !config.Enabled {
		return
	}

	ticker := time.NewTicker(config.Interval)
	defer ticker.Stop()

	// Perform initial health check
	rhm.performHealthCheck(ctx, serviceID)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			rhm.performHealthCheck(ctx, serviceID)
		}
	}
}

// performHealthCheck performs a health check for a service
func (rhm *RealTimeHealthMonitor) performHealthCheck(ctx context.Context, serviceID string) {
	rhm.mu.RLock()
	config, configExists := rhm.healthChecks[serviceID]
	state, stateExists := rhm.healthStates[serviceID]
	rhm.mu.RUnlock()

	if !configExists || !stateExists || !config.Enabled {
		return
	}

	strategy, strategyExists := rhm.strategies[config.Strategy]
	if !strategyExists {
		log.Printf("⚠️  Unknown health check strategy: %s for service %s", config.Strategy, serviceID)
		return
	}

	// Perform health check with retries
	var result *HealthCheckResult
	var err error

	for attempt := 0; attempt <= config.Retries; attempt++ {
		checkCtx, cancel := context.WithTimeout(ctx, config.Timeout)
		result, err = strategy.Check(checkCtx, config, state)
		cancel()

		if err == nil && result.Success {
			break
		}

		if attempt < config.Retries {
			time.Sleep(config.RetryDelay)
		}
	}

	if err != nil {
		result = &HealthCheckResult{
			Timestamp:    time.Now(),
			Success:      false,
			ResponseTime: config.Timeout.Seconds() * 1000, // Timeout as response time
			Message:      "Health check failed",
			Error:        err.Error(),
		}
	}

	// Update health state
	rhm.updateHealthState(serviceID, result)

	// Collect metrics
	if rhm.config.EnableMetricsCollection {
		rhm.metricsCollector.RecordHealthCheck(serviceID, result)
	}

	// Check alert conditions
	if rhm.config.EnableAlerting {
		rhm.alertManager.CheckAlertConditions(serviceID, state, result)
	}
}

// updateHealthState updates the health state based on check result
func (rhm *RealTimeHealthMonitor) updateHealthState(serviceID string, result *HealthCheckResult) {
	rhm.mu.Lock()
	defer rhm.mu.Unlock()

	state, exists := rhm.healthStates[serviceID]
	if !exists {
		return
	}

	config := rhm.healthChecks[serviceID]
	oldStatus := state.Status

	// Update check counters
	state.TotalChecks++
	state.LastCheckTime = result.Timestamp

	if result.Success {
		state.SuccessfulChecks++
		state.ConsecutiveSuccesses++
		state.ConsecutiveFailures = 0
	} else {
		state.FailedChecks++
		state.ConsecutiveFailures++
		state.ConsecutiveSuccesses = 0
		state.LastError = result.Error
	}

	// Add to recent results
	state.RecentResults = append(state.RecentResults, *result)
	
	// Keep only recent results (last 100)
	if len(state.RecentResults) > 100 {
		state.RecentResults = state.RecentResults[1:]
	}

	// Update performance metrics
	rhm.updatePerformanceMetrics(state)

	// Determine new status
	newStatus := rhm.determineHealthStatus(state, config, result)
	
	if newStatus != oldStatus {
		state.Status = newStatus
		state.LastStatusChange = time.Now()
		
		// Notify subscribers
		rhm.notifyHealthChange(serviceID, oldStatus, newStatus, state.StatusDetails)
		
		log.Printf("🏥 Service %s health changed: %s -> %s", serviceID, oldStatus, newStatus)
	}

	state.Status = newStatus

	// Update health score
	rhm.updateHealthScore(state)
}

// determineHealthStatus determines the health status based on recent results
func (rhm *RealTimeHealthMonitor) determineHealthStatus(state *HealthState, config *HealthCheckConfig, result *HealthCheckResult) HealthStatus {
	// During grace period, maintain unknown status
	if time.Since(state.LastStatusChange) < config.GracePeriod {
		if state.Status == HealthStatusUnknown {
			return HealthStatusUnknown
		}
	}

	// Determine status based on consecutive results
	if result.Success {
		if state.Status == HealthStatusUnhealthy || state.Status == HealthStatusCritical {
			// Need enough consecutive successes to become healthy
			if state.ConsecutiveSuccesses >= config.SuccessBeforeHealthy {
				return HealthStatusHealthy
			}
		} else {
			// Already healthy or unknown
			if state.ConsecutiveSuccesses >= config.HealthyThreshold {
				return HealthStatusHealthy
			}
		}
	} else {
		// Failed check
		if state.ConsecutiveFailures >= config.FailuresBeforeUnhealthy {
			// Determine severity based on failure patterns
			if rhm.isCriticalFailure(state, result) {
				return HealthStatusCritical
			}
			return HealthStatusUnhealthy
		}
	}

	// No status change
	return state.Status
}

// isCriticalFailure determines if a failure is critical
func (rhm *RealTimeHealthMonitor) isCriticalFailure(state *HealthState, result *HealthCheckResult) bool {
	// Consider it critical if:
	// 1. Very high response time (> 10 seconds)
	// 2. Many consecutive failures (> 10)
	// 3. Specific error types
	
	if result.ResponseTime > 10000 { // 10 seconds
		return true
	}
	
	if state.ConsecutiveFailures > 10 {
		return true
	}
	
	// Check for specific critical errors
	criticalErrors := []string{"connection refused", "timeout", "dns resolution failed"}
	for _, criticalError := range criticalErrors {
		if result.Error != "" && len(result.Error) > 0 {
			// Simple substring check
			errorLower := result.Error
			for _, char := range criticalError {
				found := false
				for _, errChar := range errorLower {
					if char == errChar {
						found = true
						break
					}
				}
				if !found {
					break
				}
			}
		}
	}
	
	return false
}

// updatePerformanceMetrics updates performance metrics for a service
func (rhm *RealTimeHealthMonitor) updatePerformanceMetrics(state *HealthState) {
	if len(state.RecentResults) == 0 {
		return
	}

	// Calculate average response time from recent results
	totalResponseTime := 0.0
	successCount := 0
	
	for _, result := range state.RecentResults {
		if result.Success {
			totalResponseTime += result.ResponseTime
			successCount++
		}
	}
	
	if successCount > 0 {
		state.AverageResponseTime = totalResponseTime / float64(successCount)
	}

	// Calculate uptime percentage
	if state.TotalChecks > 0 {
		state.UpTimePercentage = float64(state.SuccessfulChecks) / float64(state.TotalChecks) * 100.0
	}

	// Calculate percentiles (simplified)
	var responseTimes []float64
	for _, result := range state.RecentResults {
		if result.Success {
			responseTimes = append(responseTimes, result.ResponseTime)
		}
	}

	if len(responseTimes) > 0 {
		// Simple sort for percentile calculation
		for i := 0; i < len(responseTimes); i++ {
			for j := i + 1; j < len(responseTimes); j++ {
				if responseTimes[i] > responseTimes[j] {
					responseTimes[i], responseTimes[j] = responseTimes[j], responseTimes[i]
				}
			}
		}

		p95Index := int(float64(len(responseTimes)) * 0.95)
		p99Index := int(float64(len(responseTimes)) * 0.99)
		
		if p95Index < len(responseTimes) {
			state.P95ResponseTime = responseTimes[p95Index]
		}
		if p99Index < len(responseTimes) {
			state.P99ResponseTime = responseTimes[p99Index]
		}
	}
}

// updateHealthScore calculates and updates the health score
func (rhm *RealTimeHealthMonitor) updateHealthScore(state *HealthState) {
	score := 0.0
	
	// Base score on status
	switch state.Status {
	case HealthStatusHealthy:
		score = 1.0
	case HealthStatusUnhealthy:
		score = 0.3
	case HealthStatusCritical:
		score = 0.1
	case HealthStatusUnknown:
		score = 0.5
	}
	
	// Adjust based on recent performance
	if len(state.RecentResults) > 0 {
		recentSuccessRate := 0.0
		recentCount := 0
		cutoff := time.Now().Add(-5 * time.Minute)
		
		for _, result := range state.RecentResults {
			if result.Timestamp.After(cutoff) {
				recentCount++
				if result.Success {
					recentSuccessRate += 1.0
				}
			}
		}
		
		if recentCount > 0 {
			recentSuccessRate /= float64(recentCount)
			// Weighted combination of status score and recent performance
			score = 0.7*score + 0.3*recentSuccessRate
		}
	}
	
	// Adjust based on response time
	if state.AverageResponseTime > 0 {
		responseTimeFactor := 1.0
		if state.AverageResponseTime > 1000 { // > 1 second
			responseTimeFactor = 0.8
		}
		if state.AverageResponseTime > 5000 { // > 5 seconds
			responseTimeFactor = 0.5
		}
		score *= responseTimeFactor
	}
	
	state.HealthScore = score
}

// notifyHealthChange notifies subscribers of health status changes
func (rhm *RealTimeHealthMonitor) notifyHealthChange(serviceID string, oldStatus, newStatus HealthStatus, details map[string]interface{}) {
	subscribers, exists := rhm.subscribers[serviceID]
	if !exists {
		return
	}

	for _, subscriber := range subscribers {
		go func(sub HealthEventSubscriber) {
			sub.OnHealthChange(serviceID, oldStatus, newStatus, details)
		}(subscriber)
	}
}

// metricsCollectionLoop periodically collects and aggregates metrics
func (rhm *RealTimeHealthMonitor) metricsCollectionLoop(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			rhm.collectMetrics()
		}
	}
}

// collectMetrics collects metrics from all monitored services
func (rhm *RealTimeHealthMonitor) collectMetrics() {
	rhm.mu.RLock()
	states := make(map[string]*HealthState)
	for k, v := range rhm.healthStates {
		stateCopy := *v
		states[k] = &stateCopy
	}
	rhm.mu.RUnlock()

	for serviceID, state := range states {
		rhm.metricsCollector.UpdateMetrics(serviceID, state)
	}
}

// alertManagementLoop manages health alerts
func (rhm *RealTimeHealthMonitor) alertManagementLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			rhm.alertManager.ProcessAlerts()
		}
	}
}

// GetHealthMetrics returns health metrics for a service
func (rhm *RealTimeHealthMonitor) GetHealthMetrics(serviceID string) (*HealthMetrics, error) {
	return rhm.metricsCollector.GetMetrics(serviceID)
}

// GetAllHealthMetrics returns health metrics for all services
func (rhm *RealTimeHealthMonitor) GetAllHealthMetrics() map[string]*HealthMetrics {
	return rhm.metricsCollector.GetAllMetrics()
}

// === Health Check Strategies ===

// HTTPHealthCheckStrategy implements HTTP-based health checks
type HTTPHealthCheckStrategy struct {
	client *http.Client
}

func (h *HTTPHealthCheckStrategy) Name() string { return "http" }

func (h *HTTPHealthCheckStrategy) Check(ctx context.Context, config *HealthCheckConfig, state *HealthState) (*HealthCheckResult, error) {
	if config.HTTPConfig == nil {
		return nil, fmt.Errorf("HTTP health check configuration is missing")
	}

	startTime := time.Now()
	
	req, err := http.NewRequestWithContext(ctx, config.HTTPConfig.Method, config.HTTPConfig.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Add headers
	for key, value := range config.HTTPConfig.Headers {
		req.Header.Set(key, value)
	}

	resp, err := h.client.Do(req)
	responseTime := time.Since(startTime).Seconds() * 1000 // Convert to milliseconds

	result := &HealthCheckResult{
		Timestamp:    startTime,
		ResponseTime: responseTime,
		Details:      make(map[string]interface{}),
	}

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		result.Message = "HTTP request failed"
		return result, nil
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode
	result.Details["status_code"] = resp.StatusCode
	result.Details["headers"] = resp.Header

	// Check status code
	expectedStatusCode := config.HTTPConfig.ExpectedStatusCode
	if expectedStatusCode == 0 {
		expectedStatusCode = 200 // Default to 200 OK
	}

	if resp.StatusCode == expectedStatusCode {
		result.Success = true
		result.Message = "HTTP health check passed"
	} else {
		result.Success = false
		result.Message = fmt.Sprintf("Expected status %d, got %d", expectedStatusCode, resp.StatusCode)
	}

	// TODO: Add body validation if ExpectedBody is configured

	return result, nil
}

// TCPHealthCheckStrategy implements TCP-based health checks
type TCPHealthCheckStrategy struct{}

func (t *TCPHealthCheckStrategy) Name() string { return "tcp" }

func (t *TCPHealthCheckStrategy) Check(ctx context.Context, config *HealthCheckConfig, state *HealthState) (*HealthCheckResult, error) {
	if config.TCPConfig == nil {
		return nil, fmt.Errorf("TCP health check configuration is missing")
	}

	startTime := time.Now()
	address := fmt.Sprintf("%s:%d", config.TCPConfig.Host, config.TCPConfig.Port)
	
	// TODO: Implement actual TCP connection check
	// For now, return a placeholder result
	
	responseTime := time.Since(startTime).Seconds() * 1000

	result := &HealthCheckResult{
		Timestamp:    startTime,
		Success:      true, // Placeholder
		ResponseTime: responseTime,
		Message:      fmt.Sprintf("TCP connection to %s successful", address),
		Details:      map[string]interface{}{"address": address},
	}

	return result, nil
}

// CustomHealthCheckStrategy implements custom command-based health checks
type CustomHealthCheckStrategy struct{}

func (c *CustomHealthCheckStrategy) Name() string { return "custom" }

func (c *CustomHealthCheckStrategy) Check(ctx context.Context, config *HealthCheckConfig, state *HealthState) (*HealthCheckResult, error) {
	if config.CustomConfig == nil {
		return nil, fmt.Errorf("custom health check configuration is missing")
	}

	startTime := time.Now()
	
	// TODO: Implement actual command execution
	// For now, return a placeholder result
	
	responseTime := time.Since(startTime).Seconds() * 1000

	result := &HealthCheckResult{
		Timestamp:    startTime,
		Success:      true, // Placeholder
		ResponseTime: responseTime,
		Message:      "Custom health check passed",
		Details:      map[string]interface{}{"command": config.CustomConfig.Command},
	}

	return result, nil
}

// === HealthMetricsCollector Methods ===

// InitializeService initializes metrics for a service
func (hmc *HealthMetricsCollector) InitializeService(serviceID string) {
	hmc.mu.Lock()
	defer hmc.mu.Unlock()

	hmc.metrics[serviceID] = &HealthMetrics{
		ServiceID:       serviceID,
		UptimePercentage: 100.0,
		LastUpdated:     time.Now(),
	}
}

// RecordHealthCheck records a health check result
func (hmc *HealthMetricsCollector) RecordHealthCheck(serviceID string, result *HealthCheckResult) {
	hmc.mu.Lock()
	defer hmc.mu.Unlock()

	metrics, exists := hmc.metrics[serviceID]
	if !exists {
		return
	}

	metrics.TotalChecks++
	if result.Success {
		metrics.SuccessfulChecks++
	} else {
		metrics.FailedChecks++
	}

	// Update response time metrics
	if result.Success {
		alpha := 0.1 // Exponential smoothing factor
		if metrics.AverageResponseTime == 0 {
			metrics.AverageResponseTime = result.ResponseTime
		} else {
			metrics.AverageResponseTime = alpha*result.ResponseTime + (1-alpha)*metrics.AverageResponseTime
		}
	}

	// Update uptime percentage
	if metrics.TotalChecks > 0 {
		metrics.UptimePercentage = float64(metrics.SuccessfulChecks) / float64(metrics.TotalChecks) * 100.0
	}

	metrics.LastUpdated = time.Now()
}

// UpdateMetrics updates comprehensive metrics for a service
func (hmc *HealthMetricsCollector) UpdateMetrics(serviceID string, state *HealthState) {
	hmc.mu.Lock()
	defer hmc.mu.Unlock()

	metrics, exists := hmc.metrics[serviceID]
	if !exists {
		return
	}

	// Update from state
	metrics.TotalChecks = state.TotalChecks
	metrics.SuccessfulChecks = state.SuccessfulChecks
	metrics.FailedChecks = state.FailedChecks
	metrics.AverageResponseTime = state.AverageResponseTime
	metrics.P95ResponseTime = state.P95ResponseTime
	metrics.P99ResponseTime = state.P99ResponseTime
	metrics.UptimePercentage = state.UpTimePercentage

	// Calculate additional metrics
	if len(state.RecentResults) > 1 {
		hmc.calculateAdvancedMetrics(metrics, state)
	}

	metrics.LastUpdated = time.Now()
}

// calculateAdvancedMetrics calculates advanced metrics like MTBF and MTTR
func (hmc *HealthMetricsCollector) calculateAdvancedMetrics(metrics *HealthMetrics, state *HealthState) {
	// Calculate Mean Time Between Failures (MTBF)
	failureEvents := 0
	var lastFailureTime time.Time
	totalTimeBetweenFailures := 0.0

	for i, result := range state.RecentResults {
		if !result.Success {
			if !lastFailureTime.IsZero() && i > 0 {
				timeBetween := result.Timestamp.Sub(lastFailureTime).Minutes()
				totalTimeBetweenFailures += timeBetween
				failureEvents++
			}
			lastFailureTime = result.Timestamp
		}
	}

	if failureEvents > 0 {
		metrics.MeanTimeBetweenFailures = totalTimeBetweenFailures / float64(failureEvents)
	}

	// Calculate Mean Time To Recovery (MTTR)
	recoveryEvents := 0
	totalRecoveryTime := 0.0
	var failureStartTime time.Time
	inFailureState := false

	for _, result := range state.RecentResults {
		if !result.Success && !inFailureState {
			failureStartTime = result.Timestamp
			inFailureState = true
		} else if result.Success && inFailureState {
			recoveryTime := result.Timestamp.Sub(failureStartTime).Minutes()
			totalRecoveryTime += recoveryTime
			recoveryEvents++
			inFailureState = false
		}
	}

	if recoveryEvents > 0 {
		metrics.MeanTimeToRecover = totalRecoveryTime / float64(recoveryEvents)
	}
}

// GetMetrics returns metrics for a specific service
func (hmc *HealthMetricsCollector) GetMetrics(serviceID string) (*HealthMetrics, error) {
	hmc.mu.RLock()
	defer hmc.mu.RUnlock()

	metrics, exists := hmc.metrics[serviceID]
	if !exists {
		return nil, fmt.Errorf("metrics not found for service: %s", serviceID)
	}

	// Return a copy
	metricsCopy := *metrics
	return &metricsCopy, nil
}

// GetAllMetrics returns metrics for all services
func (hmc *HealthMetricsCollector) GetAllMetrics() map[string]*HealthMetrics {
	hmc.mu.RLock()
	defer hmc.mu.RUnlock()

	result := make(map[string]*HealthMetrics)
	for serviceID, metrics := range hmc.metrics {
		metricsCopy := *metrics
		result[serviceID] = &metricsCopy
	}

	return result
}

// === HealthAlertManager Methods ===

// CheckAlertConditions checks if any alert conditions are met
func (ham *HealthAlertManager) CheckAlertConditions(serviceID string, state *HealthState, result *HealthCheckResult) {
	ham.mu.RLock()
	rules := make([]*AlertRule, 0)
	for _, rule := range ham.alertRules {
		if rule.ServiceID == serviceID && rule.Enabled {
			ruleCopy := *rule
			rules = append(rules, &ruleCopy)
		}
	}
	ham.mu.RUnlock()

	for _, rule := range rules {
		if ham.evaluateAlertCondition(rule, state, result) {
			ham.triggerAlert(rule, state, result)
		}
	}
}

// evaluateAlertCondition evaluates if an alert condition is met
func (ham *HealthAlertManager) evaluateAlertCondition(rule *AlertRule, state *HealthState, result *HealthCheckResult) bool {
	switch rule.Condition {
	case "failure_rate_above":
		if state.TotalChecks > 0 {
			failureRate := float64(state.FailedChecks) / float64(state.TotalChecks)
			return failureRate > rule.Threshold
		}
	case "response_time_above":
		return result.ResponseTime > rule.Threshold
	case "consecutive_failures":
		return float64(state.ConsecutiveFailures) >= rule.Threshold
	case "health_score_below":
		return state.HealthScore < rule.Threshold
	case "status_unhealthy":
		return state.Status == HealthStatusUnhealthy || state.Status == HealthStatusCritical
	}

	return false
}

// triggerAlert triggers an alert
func (ham *HealthAlertManager) triggerAlert(rule *AlertRule, state *HealthState, result *HealthCheckResult) {
	ham.mu.Lock()
	defer ham.mu.Unlock()

	// Check cooldown period
	if lastAlert, exists := ham.lastAlertTime[rule.ID]; exists {
		if time.Since(lastAlert) < ham.cooldownPeriod {
			return
		}
	}

	alert := &ActiveAlert{
		ID:        fmt.Sprintf("alert_%d", time.Now().UnixNano()),
		RuleID:    rule.ID,
		ServiceID: rule.ServiceID,
		Severity:  rule.Severity,
		Message:   rule.Message,
		StartTime: time.Now(),
		LastUpdate: time.Now(),
		Details: map[string]interface{}{
			"health_status":        state.Status,
			"consecutive_failures": state.ConsecutiveFailures,
			"health_score":         state.HealthScore,
			"last_error":           state.LastError,
		},
	}

	ham.activeAlerts[alert.ID] = alert
	ham.lastAlertTime[rule.ID] = time.Now()
	rule.LastTriggered = time.Now()

	// Record alert event
	event := AlertEvent{
		ID:        alert.ID,
		Type:      "triggered",
		RuleID:    rule.ID,
		ServiceID: rule.ServiceID,
		Severity:  rule.Severity,
		Message:   rule.Message,
		Timestamp: time.Now(),
		Details:   alert.Details,
	}

	ham.alertHistory = append(ham.alertHistory, event)

	log.Printf("🚨 Alert triggered: %s - %s (Service: %s)", alert.Severity, alert.Message, alert.ServiceID)
}

// ProcessAlerts processes active alerts and resolves them if conditions are no longer met
func (ham *HealthAlertManager) ProcessAlerts() {
	ham.mu.Lock()
	defer ham.mu.Unlock()

	// This would typically check if alert conditions are still met
	// and resolve alerts that are no longer applicable
	
	// For now, we'll just log active alerts
	if len(ham.activeAlerts) > 0 {
		log.Printf("🚨 Active alerts: %d", len(ham.activeAlerts))
	}
}

// AddAlertRule adds a new alert rule
func (ham *HealthAlertManager) AddAlertRule(rule *AlertRule) {
	ham.mu.Lock()
	defer ham.mu.Unlock()

	ham.alertRules[rule.ID] = rule
	log.Printf("🚨 Added alert rule: %s for service %s", rule.ID, rule.ServiceID)
}

// GetActiveAlerts returns all active alerts
func (ham *HealthAlertManager) GetActiveAlerts() map[string]*ActiveAlert {
	ham.mu.RLock()
	defer ham.mu.RUnlock()

	result := make(map[string]*ActiveAlert)
	for k, v := range ham.activeAlerts {
		alertCopy := *v
		result[k] = &alertCopy
	}

	return result
}

// GetAlertHistory returns recent alert history
func (ham *HealthAlertManager) GetAlertHistory(limit int) []AlertEvent {
	ham.mu.RLock()
	defer ham.mu.RUnlock()

	if limit <= 0 || limit > len(ham.alertHistory) {
		limit = len(ham.alertHistory)
	}

	result := make([]AlertEvent, limit)
	startIndex := len(ham.alertHistory) - limit
	
	for i := 0; i < limit; i++ {
		result[i] = ham.alertHistory[startIndex+i]
	}

	return result
}