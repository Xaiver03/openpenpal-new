// Package servicemesh provides adaptive circuit breaker functionality
package servicemesh

import (
	"context"
	"fmt"
	"log"
	"math"
	"sync"
	"time"
)

// AdaptiveCircuitBreaker provides intelligent circuit breaking with self-adaptation
type AdaptiveCircuitBreaker struct {
	// Circuit configurations per service
	circuits map[string]*CircuitBreakerConfig
	
	// Circuit states per service
	states map[string]*CircuitBreakerState
	
	// AI predictor for circuit thresholds
	thresholdPredictor *ThresholdPredictor
	
	// Performance analyzer
	performanceAnalyzer *PerformanceAnalyzer
	
	// Global configuration
	globalConfig *GlobalCircuitConfig
	
	// Mutex for thread safety
	mu sync.RWMutex
	
	// Running state
	running bool
}

// CircuitBreakerConfig holds configuration for a specific service circuit
type CircuitBreakerConfig struct {
	ServiceID               string        `json:"service_id"`
	FailureThreshold        int           `json:"failure_threshold"`
	SuccessThreshold        int           `json:"success_threshold"`
	TimeoutThreshold        time.Duration `json:"timeout_threshold"`
	SlowCallThreshold       time.Duration `json:"slow_call_threshold"`
	SlowCallRateThreshold   float64       `json:"slow_call_rate_threshold"`
	FailureRateThreshold    float64       `json:"failure_rate_threshold"`
	MinimumRequestThreshold int           `json:"minimum_request_threshold"`
	SlidingWindowSize       int           `json:"sliding_window_size"`
	WaitDurationInOpenState time.Duration `json:"wait_duration_in_open_state"`
	MaxWaitDurationInHalfOpenState time.Duration `json:"max_wait_duration_in_half_open_state"`
	
	// Adaptive configuration
	AdaptiveMode            bool    `json:"adaptive_mode"`
	LearningRate            float64 `json:"learning_rate"`
	AdaptationInterval      time.Duration `json:"adaptation_interval"`
	
	// Last updated
	LastUpdated time.Time `json:"last_updated"`
}

// CircuitBreakerState tracks the current state of a circuit
type CircuitBreakerState struct {
	ServiceID       string        `json:"service_id"`
	State           CircuitState  `json:"state"`
	FailureCount    int           `json:"failure_count"`
	SuccessCount    int           `json:"success_count"`
	SlowCallCount   int           `json:"slow_call_count"`
	TotalCalls      int           `json:"total_calls"`
	LastFailureTime time.Time     `json:"last_failure_time"`
	LastSuccessTime time.Time     `json:"last_success_time"`
	StateChangedAt  time.Time     `json:"state_changed_at"`
	
	// Sliding window for metrics
	CallWindow      []CallResult  `json:"call_window"`
	WindowStartTime time.Time     `json:"window_start_time"`
	
	// Performance metrics
	AverageResponseTime float64 `json:"average_response_time"`
	P95ResponseTime     float64 `json:"p95_response_time"`
	P99ResponseTime     float64 `json:"p99_response_time"`
	
	// Health score (0-1)
	HealthScore float64 `json:"health_score"`
}

// CallResult represents the result of a service call
type CallResult struct {
	Timestamp    time.Time `json:"timestamp"`
	Success      bool      `json:"success"`
	ResponseTime float64   `json:"response_time"`
	ErrorType    string    `json:"error_type"`
	StatusCode   int       `json:"status_code"`
}

// GlobalCircuitConfig holds global circuit breaker configuration
type GlobalCircuitConfig struct {
	DefaultFailureThreshold        int           `json:"default_failure_threshold"`
	DefaultSuccessThreshold        int           `json:"default_success_threshold"`
	DefaultTimeoutThreshold        time.Duration `json:"default_timeout_threshold"`
	DefaultSlowCallThreshold       time.Duration `json:"default_slow_call_threshold"`
	DefaultFailureRateThreshold    float64       `json:"default_failure_rate_threshold"`
	DefaultSlowCallRateThreshold   float64       `json:"default_slow_call_rate_threshold"`
	DefaultSlidingWindowSize       int           `json:"default_sliding_window_size"`
	DefaultWaitDurationInOpenState time.Duration `json:"default_wait_duration_in_open_state"`
	MaxCircuits                    int           `json:"max_circuits"`
	EnableAdaptiveMode             bool          `json:"enable_adaptive_mode"`
	GlobalLearningRate             float64       `json:"global_learning_rate"`
	HealthCheckInterval            time.Duration `json:"health_check_interval"`
	MetricsRetentionPeriod         time.Duration `json:"metrics_retention_period"`
}

// ThresholdPredictor uses AI to predict optimal circuit breaker thresholds
type ThresholdPredictor struct {
	// Neural network weights for threshold prediction
	weights map[string]*ThresholdWeights
	
	// Historical performance data
	historicalData map[string]*HistoricalPerformance
	
	// Learning parameters
	learningRate float64
	
	mu sync.RWMutex
}

// ThresholdWeights holds neural network weights for threshold prediction
type ThresholdWeights struct {
	FailureRateWeight     float64 `json:"failure_rate_weight"`
	ResponseTimeWeight    float64 `json:"response_time_weight"`
	ThroughputWeight      float64 `json:"throughput_weight"`
	ErrorVarianceWeight   float64 `json:"error_variance_weight"`
	TimePatternWeight     float64 `json:"time_pattern_weight"`
	LoadPatternWeight     float64 `json:"load_pattern_weight"`
	SeasonalWeight        float64 `json:"seasonal_weight"`
	BiasWeight            float64 `json:"bias_weight"`
}

// HistoricalPerformance stores historical performance data for AI analysis
type HistoricalPerformance struct {
	ServiceID           string                `json:"service_id"`
	DailyPatterns       [24]PerformanceMetric `json:"daily_patterns"`
	WeeklyPatterns      [7]PerformanceMetric  `json:"weekly_patterns"`
	FailurePatterns     []FailurePattern      `json:"failure_patterns"`
	RecoveryPatterns    []RecoveryPattern     `json:"recovery_patterns"`
	LoadCorrelations    []LoadCorrelation     `json:"load_correlations"`
	LastUpdated         time.Time             `json:"last_updated"`
}

// PerformanceMetric represents aggregated performance metrics
type PerformanceMetric struct {
	AverageResponseTime float64 `json:"average_response_time"`
	FailureRate         float64 `json:"failure_rate"`
	Throughput          float64 `json:"throughput"`
	ErrorVariance       float64 `json:"error_variance"`
	SampleCount         int     `json:"sample_count"`
}

// FailurePattern represents patterns in service failures
type FailurePattern struct {
	Timestamp       time.Time `json:"timestamp"`
	FailureType     string    `json:"failure_type"`
	DurationMinutes int       `json:"duration_minutes"`
	RecoveryTime    int       `json:"recovery_time"`
	LoadAtFailure   float64   `json:"load_at_failure"`
	PrecedingEvents []string  `json:"preceding_events"`
}

// RecoveryPattern represents patterns in service recovery
type RecoveryPattern struct {
	Timestamp        time.Time `json:"timestamp"`
	RecoveryTrigger  string    `json:"recovery_trigger"`
	RecoveryDuration int       `json:"recovery_duration"`
	StabilityPeriod  int       `json:"stability_period"`
	LoadAtRecovery   float64   `json:"load_at_recovery"`
}

// LoadCorrelation represents correlation between load and performance
type LoadCorrelation struct {
	LoadLevel       float64 `json:"load_level"`
	ResponseTime    float64 `json:"response_time"`
	FailureRate     float64 `json:"failure_rate"`
	CorrelationTime time.Time `json:"correlation_time"`
}

// PerformanceAnalyzer analyzes service performance for circuit decisions
type PerformanceAnalyzer struct {
	// Statistical models
	anomalyDetector *AnomalyDetector
	trendAnalyzer   *TrendAnalyzer
	
	mu sync.RWMutex
}

// AnomalyDetector is defined in anomaly.go

// BaselineMetrics holds baseline performance metrics
type BaselineMetrics struct {
	ServiceID               string  `json:"service_id"`
	BaselineResponseTime    float64 `json:"baseline_response_time"`
	BaselineFailureRate     float64 `json:"baseline_failure_rate"`
	BaselineThroughput      float64 `json:"baseline_throughput"`
	ResponseTimeStdDev      float64 `json:"response_time_std_dev"`
	FailureRateStdDev       float64 `json:"failure_rate_std_dev"`
	ThroughputStdDev        float64 `json:"throughput_std_dev"`
	LastCalibrated          time.Time `json:"last_calibrated"`
}

// TrendAnalyzer analyzes performance trends
type TrendAnalyzer struct {
	trendData map[string]*TrendData
}

// TrendData holds trend analysis data
type TrendData struct {
	ServiceID           string    `json:"service_id"`
	ResponseTimeTrend   float64   `json:"response_time_trend"`
	FailureRateTrend    float64   `json:"failure_rate_trend"`
	ThroughputTrend     float64   `json:"throughput_trend"`
	TrendConfidence     float64   `json:"trend_confidence"`
	TrendStartTime      time.Time `json:"trend_start_time"`
}

// NewAdaptiveCircuitBreaker creates a new adaptive circuit breaker
func NewAdaptiveCircuitBreaker() *AdaptiveCircuitBreaker {
	globalConfig := &GlobalCircuitConfig{
		DefaultFailureThreshold:        10,
		DefaultSuccessThreshold:        5,
		DefaultTimeoutThreshold:        5 * time.Second,
		DefaultSlowCallThreshold:       1 * time.Second,
		DefaultFailureRateThreshold:    0.5,
		DefaultSlowCallRateThreshold:   0.3,
		DefaultSlidingWindowSize:       100,
		DefaultWaitDurationInOpenState: 30 * time.Second,
		MaxCircuits:                    1000,
		EnableAdaptiveMode:             true,
		GlobalLearningRate:             0.01,
		HealthCheckInterval:            10 * time.Second,
		MetricsRetentionPeriod:         24 * time.Hour,
	}

	return &AdaptiveCircuitBreaker{
		circuits:            make(map[string]*CircuitBreakerConfig),
		states:              make(map[string]*CircuitBreakerState),
		thresholdPredictor:  NewThresholdPredictor(globalConfig.GlobalLearningRate),
		performanceAnalyzer: NewPerformanceAnalyzer(),
		globalConfig:        globalConfig,
	}
}

// NewThresholdPredictor creates a new threshold predictor
func NewThresholdPredictor(learningRate float64) *ThresholdPredictor {
	return &ThresholdPredictor{
		weights:        make(map[string]*ThresholdWeights),
		historicalData: make(map[string]*HistoricalPerformance),
		learningRate:   learningRate,
	}
}

// NewPerformanceAnalyzer creates a new performance analyzer
func NewPerformanceAnalyzer() *PerformanceAnalyzer {
	return &PerformanceAnalyzer{
		anomalyDetector: NewAnomalyDetector(2.0), // 2 standard deviations
		trendAnalyzer: &TrendAnalyzer{
			trendData: make(map[string]*TrendData),
		},
	}
}

// Start starts the adaptive circuit breaker
func (acb *AdaptiveCircuitBreaker) Start(ctx context.Context) error {
	acb.mu.Lock()
	defer acb.mu.Unlock()

	if acb.running {
		return fmt.Errorf("adaptive circuit breaker is already running")
	}

	log.Println("🔌 Starting Adaptive Circuit Breaker")

	// Start background monitoring
	go acb.monitorCircuits(ctx)
	
	// Start adaptive threshold adjustment
	if acb.globalConfig.EnableAdaptiveMode {
		go acb.adaptiveThresholdAdjustment(ctx)
	}
	
	// Start performance analysis
	go acb.performanceAnalysisLoop(ctx)

	acb.running = true
	log.Println("✅ Adaptive Circuit Breaker started")

	return nil
}

// RegisterService registers a service with the circuit breaker
func (acb *AdaptiveCircuitBreaker) RegisterService(serviceID string) error {
	acb.mu.Lock()
	defer acb.mu.Unlock()

	// Create default configuration
	config := &CircuitBreakerConfig{
		ServiceID:                      serviceID,
		FailureThreshold:               acb.globalConfig.DefaultFailureThreshold,
		SuccessThreshold:               acb.globalConfig.DefaultSuccessThreshold,
		TimeoutThreshold:               acb.globalConfig.DefaultTimeoutThreshold,
		SlowCallThreshold:              acb.globalConfig.DefaultSlowCallThreshold,
		FailureRateThreshold:           acb.globalConfig.DefaultFailureRateThreshold,
		SlowCallRateThreshold:          acb.globalConfig.DefaultSlowCallRateThreshold,
		MinimumRequestThreshold:        10,
		SlidingWindowSize:              acb.globalConfig.DefaultSlidingWindowSize,
		WaitDurationInOpenState:        acb.globalConfig.DefaultWaitDurationInOpenState,
		MaxWaitDurationInHalfOpenState: 60 * time.Second,
		AdaptiveMode:                   acb.globalConfig.EnableAdaptiveMode,
		LearningRate:                   acb.globalConfig.GlobalLearningRate,
		AdaptationInterval:             5 * time.Minute,
		LastUpdated:                    time.Now(),
	}

	// Create initial state
	state := &CircuitBreakerState{
		ServiceID:           serviceID,
		State:               CircuitStateClosed,
		CallWindow:          make([]CallResult, 0, config.SlidingWindowSize),
		WindowStartTime:     time.Now(),
		StateChangedAt:      time.Now(),
		HealthScore:         1.0,
	}

	acb.circuits[serviceID] = config
	acb.states[serviceID] = state

	// Initialize AI components
	acb.thresholdPredictor.InitializeService(serviceID)
	acb.performanceAnalyzer.InitializeService(serviceID)

	log.Printf("🔌 Registered circuit breaker for service: %s", serviceID)

	return nil
}

// RecordCall records a service call result
func (acb *AdaptiveCircuitBreaker) RecordCall(serviceID string, responseTime float64, success bool, errorType string, statusCode int) error {
	acb.mu.Lock()
	defer acb.mu.Unlock()

	state, exists := acb.states[serviceID]
	if !exists {
		return fmt.Errorf("service not registered: %s", serviceID)
	}

	config := acb.circuits[serviceID]
	
	// Create call result
	callResult := CallResult{
		Timestamp:    time.Now(),
		Success:      success,
		ResponseTime: responseTime,
		ErrorType:    errorType,
		StatusCode:   statusCode,
	}

	// Add to sliding window
	state.CallWindow = append(state.CallWindow, callResult)
	
	// Maintain window size
	if len(state.CallWindow) > config.SlidingWindowSize {
		state.CallWindow = state.CallWindow[1:]
	}

	// Update counters
	state.TotalCalls++
	if success {
		state.SuccessCount++
		state.LastSuccessTime = time.Now()
	} else {
		state.FailureCount++
		state.LastFailureTime = time.Now()
	}

	// Check for slow calls
	if responseTime > config.SlowCallThreshold.Seconds()*1000 {
		state.SlowCallCount++
	}

	// Update performance metrics
	acb.updatePerformanceMetrics(state)

	// Update health score
	acb.updateHealthScore(state, config)

	// Check state transitions
	acb.checkStateTransition(serviceID, state, config)

	// Record for AI learning
	acb.thresholdPredictor.RecordCall(serviceID, callResult)
	acb.performanceAnalyzer.RecordCall(serviceID, callResult)

	return nil
}

// IsCallAllowed checks if a call is allowed based on circuit state
func (acb *AdaptiveCircuitBreaker) IsCallAllowed(serviceID string) (bool, error) {
	acb.mu.RLock()
	defer acb.mu.RUnlock()

	state, exists := acb.states[serviceID]
	if !exists {
		return true, fmt.Errorf("service not registered: %s", serviceID)
	}

	config := acb.circuits[serviceID]

	switch state.State {
	case CircuitStateClosed:
		return true, nil
		
	case CircuitStateOpen:
		// Check if wait duration has passed
		if time.Since(state.StateChangedAt) >= config.WaitDurationInOpenState {
			// Transition to half-open
			state.State = CircuitStateHalfOpen
			state.StateChangedAt = time.Now()
			log.Printf("🔌 Circuit %s transitioned to HALF_OPEN", serviceID)
			return true, nil
		}
		return false, nil
		
	case CircuitStateHalfOpen:
		// Allow limited calls in half-open state
		recentSuccesses := acb.countRecentSuccesses(state, 1*time.Minute)
		if recentSuccesses < config.SuccessThreshold/2 {
			return true, nil
		}
		return false, nil
		
	default:
		return false, fmt.Errorf("unknown circuit state: %s", state.State)
	}
}

// GetCircuitState returns the current state of a circuit
func (acb *AdaptiveCircuitBreaker) GetCircuitState(serviceID string) (CircuitState, error) {
	acb.mu.RLock()
	defer acb.mu.RUnlock()

	state, exists := acb.states[serviceID]
	if !exists {
		return "", fmt.Errorf("service not registered: %s", serviceID)
	}

	return state.State, nil
}

// GetCircuitMetrics returns detailed metrics for a circuit
func (acb *AdaptiveCircuitBreaker) GetCircuitMetrics(serviceID string) (*CircuitMetrics, error) {
	acb.mu.RLock()
	defer acb.mu.RUnlock()

	state, exists := acb.states[serviceID]
	if !exists {
		return nil, fmt.Errorf("service not registered: %s", serviceID)
	}

	config := acb.circuits[serviceID]

	// Calculate current metrics
	failureRate := 0.0
	slowCallRate := 0.0
	if state.TotalCalls > 0 {
		failureRate = float64(state.FailureCount) / float64(state.TotalCalls)
		slowCallRate = float64(state.SlowCallCount) / float64(state.TotalCalls)
	}

	return &CircuitMetrics{
		ServiceID:           serviceID,
		State:               state.State,
		FailureRate:         failureRate,
		SlowCallRate:        slowCallRate,
		TotalCalls:          state.TotalCalls,
		FailureCount:        state.FailureCount,
		SuccessCount:        state.SuccessCount,
		SlowCallCount:       state.SlowCallCount,
		AverageResponseTime: state.AverageResponseTime,
		P95ResponseTime:     state.P95ResponseTime,
		P99ResponseTime:     state.P99ResponseTime,
		HealthScore:         state.HealthScore,
		LastFailureTime:     state.LastFailureTime,
		LastSuccessTime:     state.LastSuccessTime,
		StateChangedAt:      state.StateChangedAt,
		Configuration:       *config,
	}, nil
}

// CircuitMetrics represents circuit breaker metrics
type CircuitMetrics struct {
	ServiceID           string                `json:"service_id"`
	State               CircuitState          `json:"state"`
	FailureRate         float64               `json:"failure_rate"`
	SlowCallRate        float64               `json:"slow_call_rate"`
	TotalCalls          int                   `json:"total_calls"`
	FailureCount        int                   `json:"failure_count"`
	SuccessCount        int                   `json:"success_count"`
	SlowCallCount       int                   `json:"slow_call_count"`
	AverageResponseTime float64               `json:"average_response_time"`
	P95ResponseTime     float64               `json:"p95_response_time"`
	P99ResponseTime     float64               `json:"p99_response_time"`
	HealthScore         float64               `json:"health_score"`
	LastFailureTime     time.Time             `json:"last_failure_time"`
	LastSuccessTime     time.Time             `json:"last_success_time"`
	StateChangedAt      time.Time             `json:"state_changed_at"`
	Configuration       CircuitBreakerConfig  `json:"configuration"`
}

// UpdateMetrics updates circuit metrics from external source
func (acb *AdaptiveCircuitBreaker) UpdateMetrics(serviceID string, metrics *ServiceMetrics) {
	acb.mu.Lock()
	defer acb.mu.Unlock()

	state, exists := acb.states[serviceID]
	if !exists {
		return
	}

	// Update performance metrics from external source
	state.AverageResponseTime = metrics.AverageLatency
	
	// Update health score based on external metrics
	healthScore := 1.0
	healthScore *= (100.0 - metrics.CPUUsage) / 100.0    // CPU factor
	healthScore *= (100.0 - metrics.MemoryUsage) / 100.0 // Memory factor
	healthScore *= (1.0 - metrics.ErrorRate)             // Error rate factor
	
	state.HealthScore = healthScore
}

// checkStateTransition checks and performs state transitions
func (acb *AdaptiveCircuitBreaker) checkStateTransition(serviceID string, state *CircuitBreakerState, config *CircuitBreakerConfig) {
	currentState := state.State
	
	switch currentState {
	case CircuitStateClosed:
		if acb.shouldOpenCircuit(state, config) {
			state.State = CircuitStateOpen
			state.StateChangedAt = time.Now()
			log.Printf("🔌 Circuit %s opened due to failures", serviceID)
		}
		
	case CircuitStateHalfOpen:
		if acb.shouldCloseCircuit(state, config) {
			state.State = CircuitStateClosed
			state.StateChangedAt = time.Now()
			state.FailureCount = 0
			state.SuccessCount = 0
			state.SlowCallCount = 0
			log.Printf("🔌 Circuit %s closed after recovery", serviceID)
		} else if acb.shouldOpenCircuit(state, config) {
			state.State = CircuitStateOpen
			state.StateChangedAt = time.Now()
			log.Printf("🔌 Circuit %s reopened due to continued failures", serviceID)
		}
	}
}

// shouldOpenCircuit determines if circuit should be opened
func (acb *AdaptiveCircuitBreaker) shouldOpenCircuit(state *CircuitBreakerState, config *CircuitBreakerConfig) bool {
	// Need minimum number of requests
	if state.TotalCalls < config.MinimumRequestThreshold {
		return false
	}

	// Check failure rate threshold
	failureRate := float64(state.FailureCount) / float64(state.TotalCalls)
	if failureRate >= config.FailureRateThreshold {
		return true
	}

	// Check slow call rate threshold
	slowCallRate := float64(state.SlowCallCount) / float64(state.TotalCalls)
	if slowCallRate >= config.SlowCallRateThreshold {
		return true
	}

	// Check health score
	if state.HealthScore < 0.3 {
		return true
	}

	return false
}

// shouldCloseCircuit determines if circuit should be closed
func (acb *AdaptiveCircuitBreaker) shouldCloseCircuit(state *CircuitBreakerState, config *CircuitBreakerConfig) bool {
	recentSuccesses := acb.countRecentSuccesses(state, config.MaxWaitDurationInHalfOpenState)
	return recentSuccesses >= config.SuccessThreshold
}

// countRecentSuccesses counts successful calls in the given time window
func (acb *AdaptiveCircuitBreaker) countRecentSuccesses(state *CircuitBreakerState, window time.Duration) int {
	cutoff := time.Now().Add(-window)
	count := 0
	
	for _, call := range state.CallWindow {
		if call.Timestamp.After(cutoff) && call.Success {
			count++
		}
	}
	
	return count
}

// updatePerformanceMetrics updates various performance metrics
func (acb *AdaptiveCircuitBreaker) updatePerformanceMetrics(state *CircuitBreakerState) {
	if len(state.CallWindow) == 0 {
		return
	}

	// Calculate response time metrics
	responseTimes := make([]float64, 0, len(state.CallWindow))
	for _, call := range state.CallWindow {
		responseTimes = append(responseTimes, call.ResponseTime)
	}

	// Calculate average
	sum := 0.0
	for _, rt := range responseTimes {
		sum += rt
	}
	state.AverageResponseTime = sum / float64(len(responseTimes))

	// Calculate percentiles
	if len(responseTimes) > 0 {
		sorted := make([]float64, len(responseTimes))
		copy(sorted, responseTimes)
		
		// Simple sort for percentile calculation
		for i := 0; i < len(sorted); i++ {
			for j := i + 1; j < len(sorted); j++ {
				if sorted[i] > sorted[j] {
					sorted[i], sorted[j] = sorted[j], sorted[i]
				}
			}
		}

		p95Index := int(float64(len(sorted)) * 0.95)
		p99Index := int(float64(len(sorted)) * 0.99)
		
		if p95Index < len(sorted) {
			state.P95ResponseTime = sorted[p95Index]
		}
		if p99Index < len(sorted) {
			state.P99ResponseTime = sorted[p99Index]
		}
	}
}

// updateHealthScore calculates and updates the health score
func (acb *AdaptiveCircuitBreaker) updateHealthScore(state *CircuitBreakerState, config *CircuitBreakerConfig) {
	if state.TotalCalls == 0 {
		state.HealthScore = 1.0
		return
	}

	// Calculate component scores
	failureRate := float64(state.FailureCount) / float64(state.TotalCalls)
	slowCallRate := float64(state.SlowCallCount) / float64(state.TotalCalls)
	
	// Health score components
	failureScore := 1.0 - failureRate
	speedScore := 1.0 - slowCallRate
	
	// Recent performance weight
	recentWeight := 0.7
	if len(state.CallWindow) > 0 {
		recentCalls := 0
		recentFailures := 0
		cutoff := time.Now().Add(-5 * time.Minute)
		
		for _, call := range state.CallWindow {
			if call.Timestamp.After(cutoff) {
				recentCalls++
				if !call.Success {
					recentFailures++
				}
			}
		}
		
		if recentCalls > 0 {
			recentFailureRate := float64(recentFailures) / float64(recentCalls)
			recentScore := 1.0 - recentFailureRate
			failureScore = recentWeight*recentScore + (1.0-recentWeight)*failureScore
		}
	}

	// Combine scores
	state.HealthScore = 0.6*failureScore + 0.4*speedScore
	
	// Clamp to [0, 1]
	if state.HealthScore < 0 {
		state.HealthScore = 0
	}
	if state.HealthScore > 1 {
		state.HealthScore = 1
	}
}

// monitorCircuits monitors all circuits and performs maintenance
func (acb *AdaptiveCircuitBreaker) monitorCircuits(ctx context.Context) {
	ticker := time.NewTicker(acb.globalConfig.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			acb.performMaintenance()
		}
	}
}

// performMaintenance performs periodic maintenance tasks
func (acb *AdaptiveCircuitBreaker) performMaintenance() {
	acb.mu.Lock()
	defer acb.mu.Unlock()

	now := time.Now()
	retentionCutoff := now.Add(-acb.globalConfig.MetricsRetentionPeriod)

	for serviceID, state := range acb.states {
		// Clean old metrics
		acb.cleanOldMetrics(state, retentionCutoff)
		
		// Update performance baselines
		acb.performanceAnalyzer.UpdateBaseline(serviceID, state.CallWindow)
		
		// Log circuit status
		if state.State != CircuitStateClosed {
			log.Printf("🔌 Circuit %s is %s (Health: %.2f)", serviceID, state.State, state.HealthScore)
		}
	}
}

// cleanOldMetrics removes old metrics beyond retention period
func (acb *AdaptiveCircuitBreaker) cleanOldMetrics(state *CircuitBreakerState, cutoff time.Time) {
	filtered := make([]CallResult, 0, len(state.CallWindow))
	
	for _, call := range state.CallWindow {
		if call.Timestamp.After(cutoff) {
			filtered = append(filtered, call)
		}
	}
	
	state.CallWindow = filtered
}

// adaptiveThresholdAdjustment continuously adjusts thresholds based on AI predictions
func (acb *AdaptiveCircuitBreaker) adaptiveThresholdAdjustment(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			acb.adjustThresholds()
		}
	}
}

// adjustThresholds uses AI to adjust circuit breaker thresholds
func (acb *AdaptiveCircuitBreaker) adjustThresholds() {
	acb.mu.Lock()
	defer acb.mu.Unlock()

	for serviceID, config := range acb.circuits {
		if !config.AdaptiveMode {
			continue
		}

		// Get AI-predicted optimal thresholds
		predictions := acb.thresholdPredictor.PredictOptimalThresholds(serviceID)
		
		// Gradually adjust thresholds
		learningRate := config.LearningRate
		
		config.FailureRateThreshold += learningRate * (predictions.FailureRateThreshold - config.FailureRateThreshold)
		config.SlowCallRateThreshold += learningRate * (predictions.SlowCallRateThreshold - config.SlowCallRateThreshold)
		
		// Clamp thresholds to reasonable bounds
		config.FailureRateThreshold = math.Max(0.1, math.Min(0.9, config.FailureRateThreshold))
		config.SlowCallRateThreshold = math.Max(0.1, math.Min(0.8, config.SlowCallRateThreshold))
		
		config.LastUpdated = time.Now()
	}
}

// performanceAnalysisLoop continuously analyzes performance patterns
func (acb *AdaptiveCircuitBreaker) performanceAnalysisLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			acb.analyzePerformancePatterns()
		}
	}
}

// analyzePerformancePatterns analyzes performance patterns for anomalies and trends
func (acb *AdaptiveCircuitBreaker) analyzePerformancePatterns() {
	acb.mu.RLock()
	states := make(map[string]*CircuitBreakerState)
	for k, v := range acb.states {
		states[k] = v
	}
	acb.mu.RUnlock()

	for serviceID, state := range states {
		// Detect anomalies
		anomalies := acb.performanceAnalyzer.DetectAnomalies(serviceID, state.CallWindow)
		if len(anomalies) > 0 {
			log.Printf("🚨 Detected %d performance anomalies in service %s", len(anomalies), serviceID)
		}

		// Analyze trends
		trends := acb.performanceAnalyzer.AnalyzeTrends(serviceID, state.CallWindow)
		if trends.TrendConfidence > 0.8 {
			if trends.ResponseTimeTrend > 0.1 {
				log.Printf("📈 Service %s showing increasing response time trend", serviceID)
			}
			if trends.FailureRateTrend > 0.05 {
				log.Printf("📈 Service %s showing increasing failure rate trend", serviceID)
			}
		}
	}
}

// === ThresholdPredictor Methods ===

// InitializeService initializes AI components for a service
func (tp *ThresholdPredictor) InitializeService(serviceID string) {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	// Initialize weights with default values
	tp.weights[serviceID] = &ThresholdWeights{
		FailureRateWeight:   0.3,
		ResponseTimeWeight:  0.25,
		ThroughputWeight:    0.2,
		ErrorVarianceWeight: 0.15,
		TimePatternWeight:   0.05,
		LoadPatternWeight:   0.03,
		SeasonalWeight:      0.02,
		BiasWeight:          0.0,
	}

	// Initialize historical data
	tp.historicalData[serviceID] = &HistoricalPerformance{
		ServiceID:   serviceID,
		LastUpdated: time.Now(),
	}
}

// RecordCall records a call for AI learning
func (tp *ThresholdPredictor) RecordCall(serviceID string, call CallResult) {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	historical, exists := tp.historicalData[serviceID]
	if !exists {
		return
	}

	// Update hourly patterns
	hour := call.Timestamp.Hour()
	metric := &historical.DailyPatterns[hour]
	metric.SampleCount++
	
	if call.Success {
		metric.AverageResponseTime += tp.learningRate * (call.ResponseTime - metric.AverageResponseTime)
	} else {
		metric.FailureRate += tp.learningRate * (1.0 - metric.FailureRate)
	}

	// Update weekly patterns
	weekday := int(call.Timestamp.Weekday())
	weeklyMetric := &historical.WeeklyPatterns[weekday]
	weeklyMetric.SampleCount++
	
	if call.Success {
		weeklyMetric.AverageResponseTime += tp.learningRate * (call.ResponseTime - weeklyMetric.AverageResponseTime)
	} else {
		weeklyMetric.FailureRate += tp.learningRate * (1.0 - weeklyMetric.FailureRate)
	}

	historical.LastUpdated = time.Now()
}

// PredictOptimalThresholds predicts optimal thresholds using AI
func (tp *ThresholdPredictor) PredictOptimalThresholds(serviceID string) *ThresholdPrediction {
	tp.mu.RLock()
	defer tp.mu.RUnlock()

	weights, weightExists := tp.weights[serviceID]
	historical, histExists := tp.historicalData[serviceID]
	
	if !weightExists || !histExists {
		// Return default thresholds
		return &ThresholdPrediction{
			FailureRateThreshold:  0.5,
			SlowCallRateThreshold: 0.3,
			Confidence:            0.0,
		}
	}

	// Calculate current patterns
	hour := time.Now().Hour()
	weekday := int(time.Now().Weekday())
	
	hourlyMetric := historical.DailyPatterns[hour]
	weeklyMetric := historical.WeeklyPatterns[weekday]

	// Neural network prediction (simplified)
	features := []float64{
		hourlyMetric.FailureRate,
		hourlyMetric.AverageResponseTime / 1000.0, // Normalize
		weeklyMetric.FailureRate,
		weeklyMetric.AverageResponseTime / 1000.0,
		float64(hour) / 24.0,
		float64(weekday) / 7.0,
	}

	// Apply weights
	failureThreshold := 0.5 // Base threshold
	slowCallThreshold := 0.3 // Base threshold
	
	for i, feature := range features {
		if i < len(features)-2 {
			failureThreshold += weights.FailureRateWeight * feature
			slowCallThreshold += weights.ResponseTimeWeight * feature
		}
	}

	// Clamp predictions
	failureThreshold = math.Max(0.1, math.Min(0.9, failureThreshold))
	slowCallThreshold = math.Max(0.1, math.Min(0.8, slowCallThreshold))

	// Calculate confidence based on sample size
	confidence := math.Min(1.0, float64(hourlyMetric.SampleCount)/100.0)

	return &ThresholdPrediction{
		FailureRateThreshold:  failureThreshold,
		SlowCallRateThreshold: slowCallThreshold,
		Confidence:            confidence,
	}
}

// ThresholdPrediction represents AI-predicted thresholds
type ThresholdPrediction struct {
	FailureRateThreshold  float64 `json:"failure_rate_threshold"`
	SlowCallRateThreshold float64 `json:"slow_call_rate_threshold"`
	Confidence            float64 `json:"confidence"`
}

// === PerformanceAnalyzer Methods ===

// InitializeService initializes performance analysis for a service
func (pa *PerformanceAnalyzer) InitializeService(serviceID string) {
	pa.mu.Lock()
	defer pa.mu.Unlock()

	// Initialize baseline metrics with new AnomalyDetector
	// The baseline will be automatically created when first metrics are provided

	// Initialize trend data
	pa.trendAnalyzer.trendData[serviceID] = &TrendData{
		ServiceID:       serviceID,
		TrendStartTime:  time.Now(),
		TrendConfidence: 0.0,
	}
}

// DetectAnomalies detects performance anomalies
func (pa *PerformanceAnalyzer) DetectAnomalies(serviceID string, calls []CallResult) []string {
	if len(calls) == 0 {
		return nil
	}
	
	// Get baseline from the new AnomalyDetector
	anomalyBaseline := pa.anomalyDetector.GetBaseline(serviceID)
	if anomalyBaseline == nil {
		return nil
	}
	
	// Convert to legacy baseline format for compatibility
	baseline := &BaselineMetrics{
		ServiceID:            serviceID,
		BaselineResponseTime: anomalyBaseline.AverageResponseTime,
		BaselineFailureRate:  anomalyBaseline.AverageErrorRate,
		ResponseTimeStdDev:   anomalyBaseline.AverageResponseTime * 0.1, // 10% deviation
		FailureRateStdDev:    anomalyBaseline.AverageErrorRate * 0.1,
		LastCalibrated:       anomalyBaseline.LastUpdated,
	}

	var anomalies []string
	
	// Analyze recent calls
	recentCalls := calls
	if len(calls) > 50 {
		recentCalls = calls[len(calls)-50:] // Last 50 calls
	}

	// Calculate current metrics
	totalTime := 0.0
	failureCount := 0
	
	for _, call := range recentCalls {
		totalTime += call.ResponseTime
		if !call.Success {
			failureCount++
		}
	}

	avgResponseTime := totalTime / float64(len(recentCalls))
	currentFailureRate := float64(failureCount) / float64(len(recentCalls))

	// Check for anomalies
	responseTimeThreshold := baseline.BaselineResponseTime + pa.anomalyDetector.sensitivity*baseline.ResponseTimeStdDev
	if avgResponseTime > responseTimeThreshold {
		anomalies = append(anomalies, fmt.Sprintf("High response time: %.2f ms (baseline: %.2f ms)", avgResponseTime, baseline.BaselineResponseTime))
	}

	failureRateThreshold := baseline.BaselineFailureRate + pa.anomalyDetector.sensitivity*baseline.FailureRateStdDev
	if currentFailureRate > failureRateThreshold {
		anomalies = append(anomalies, fmt.Sprintf("High failure rate: %.2f%% (baseline: %.2f%%)", currentFailureRate*100, baseline.BaselineFailureRate*100))
	}

	return anomalies
}

// AnalyzeTrends analyzes performance trends
func (pa *PerformanceAnalyzer) AnalyzeTrends(serviceID string, calls []CallResult) *TrendData {
	pa.mu.Lock()
	defer pa.mu.Unlock()

	trend, exists := pa.trendAnalyzer.trendData[serviceID]
	if !exists {
		return &TrendData{}
	}

	if len(calls) < 20 {
		return trend
	}

	// Simple linear trend analysis
	recentCalls := calls
	if len(calls) > 100 {
		recentCalls = calls[len(calls)-100:] // Last 100 calls
	}

	// Calculate trend for response time
	responseTrend := pa.calculateLinearTrend(recentCalls, func(call CallResult) float64 {
		return call.ResponseTime
	})

	// Calculate trend for failure rate (simplified)
	failureTrend := pa.calculateFailureRateTrend(recentCalls)

	trend.ResponseTimeTrend = responseTrend
	trend.FailureRateTrend = failureTrend
	trend.TrendConfidence = math.Min(1.0, float64(len(recentCalls))/100.0)

	return trend
}

// calculateLinearTrend calculates linear trend for a metric
func (pa *PerformanceAnalyzer) calculateLinearTrend(calls []CallResult, valueFunc func(CallResult) float64) float64 {
	if len(calls) < 2 {
		return 0.0
	}

	n := float64(len(calls))
	sumX := 0.0
	sumY := 0.0
	sumXY := 0.0
	sumXX := 0.0

	for i, call := range calls {
		x := float64(i)
		y := valueFunc(call)
		
		sumX += x
		sumY += y
		sumXY += x * y
		sumXX += x * x
	}

	// Calculate slope (trend)
	slope := (n*sumXY - sumX*sumY) / (n*sumXX - sumX*sumX)
	
	return slope / n // Normalize by sample size
}

// calculateFailureRateTrend calculates trend in failure rate
func (pa *PerformanceAnalyzer) calculateFailureRateTrend(calls []CallResult) float64 {
	if len(calls) < 10 {
		return 0.0
	}

	// Split into two halves and compare failure rates
	midPoint := len(calls) / 2
	firstHalf := calls[:midPoint]
	secondHalf := calls[midPoint:]

	firstFailures := 0
	for _, call := range firstHalf {
		if !call.Success {
			firstFailures++
		}
	}

	secondFailures := 0
	for _, call := range secondHalf {
		if !call.Success {
			secondFailures++
		}
	}

	firstRate := float64(firstFailures) / float64(len(firstHalf))
	secondRate := float64(secondFailures) / float64(len(secondHalf))

	return secondRate - firstRate
}

// UpdateBaseline updates baseline metrics
func (pa *PerformanceAnalyzer) UpdateBaseline(serviceID string, calls []CallResult) {
	if len(calls) == 0 {
		return
	}

	// Calculate metrics from calls
	totalResponseTime := 0.0
	failureCount := 0
	for _, call := range calls {
		totalResponseTime += call.ResponseTime
		if !call.Success {
			failureCount++
		}
	}

	avgResponseTime := totalResponseTime / float64(len(calls))
	errorRate := float64(failureCount) / float64(len(calls))

	// Create ServiceMetrics and update baseline
	metrics := &ServiceMetrics{
		AverageLatency:    avgResponseTime,
		ErrorRate:         errorRate,
		RequestsPerSecond: float64(len(calls)), // Approximate
		CPUUsage:          50.0,  // Default values for now
		MemoryUsage:       40.0,
	}

	// Update baseline using the new AnomalyDetector
	pa.anomalyDetector.UpdateBaseline(serviceID, metrics)
	
	// Simplified baseline update - legacy calculation removed for now
	log.Printf("📊 Updated baseline for %s using AnomalyDetector", serviceID)
}

// RecordCall records a call for performance analysis
func (pa *PerformanceAnalyzer) RecordCall(serviceID string, call CallResult) {
	// This method can be used for real-time analysis
	// Currently, analysis is done in batch mode
}