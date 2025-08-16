// Package servicemesh provides intelligent load balancing capabilities
package servicemesh

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"sync"
	"time"
)

// IntelligentLoadBalancer provides AI-driven load balancing
type IntelligentLoadBalancer struct {
	// Load balancing strategies
	strategies map[string]LoadBalanceStrategy
	
	// Performance tracking
	performanceTracker *PerformanceTracker
	
	// AI predictor for load optimization
	loadPredictor *LoadPredictor
	
	// Configuration
	config *LoadBalancerConfig
	
	// Mutex for thread safety
	mu sync.RWMutex
}

// LoadBalanceStrategy defines load balancing algorithms
type LoadBalanceStrategy interface {
	SelectInstance(services []*ServiceInstance, request *LoadBalanceRequest) (*ServiceInstance, error)
	Name() string
}

// LoadBalancerConfig holds load balancer configuration
type LoadBalancerConfig struct {
	DefaultStrategy    string        `json:"default_strategy"`
	HealthCheckWeight  float64       `json:"health_check_weight"`
	LatencyWeight      float64       `json:"latency_weight"`
	CPUWeight          float64       `json:"cpu_weight"`
	MemoryWeight       float64       `json:"memory_weight"`
	ErrorRateWeight    float64       `json:"error_rate_weight"`
	RequestCountWeight float64       `json:"request_count_weight"`
	PredictionWindow   time.Duration `json:"prediction_window"`
	LearningRate       float64       `json:"learning_rate"`
}

// PerformanceTracker tracks service performance metrics
type PerformanceTracker struct {
	mu      sync.RWMutex
	metrics map[string]*ServicePerformanceHistory
}

// ServicePerformanceHistory holds historical performance data
type ServicePerformanceHistory struct {
	ServiceID        string                    `json:"service_id"`
	ResponseTimes    []ResponseTimeMetric     `json:"response_times"`
	ErrorRates       []ErrorRateMetric        `json:"error_rates"`
	ThroughputRates  []ThroughputMetric       `json:"throughput_rates"`
	ResourceUsage    []ResourceUsageMetric    `json:"resource_usage"`
	SuccessCount     int64                    `json:"success_count"`
	FailureCount     int64                    `json:"failure_count"`
	LastUpdated      time.Time                `json:"last_updated"`
}

// ResponseTimeMetric tracks response time over time
type ResponseTimeMetric struct {
	Timestamp    time.Time `json:"timestamp"`
	ResponseTime float64   `json:"response_time"`
	RequestType  string    `json:"request_type"`
}

// ErrorRateMetric tracks error rate over time
type ErrorRateMetric struct {
	Timestamp time.Time `json:"timestamp"`
	ErrorRate float64   `json:"error_rate"`
	ErrorType string    `json:"error_type"`
}

// ThroughputMetric tracks throughput over time
type ThroughputMetric struct {
	Timestamp           time.Time `json:"timestamp"`
	RequestsPerSecond   float64   `json:"requests_per_second"`
	ConcurrentRequests  int       `json:"concurrent_requests"`
}

// ResourceUsageMetric tracks resource usage over time
type ResourceUsageMetric struct {
	Timestamp   time.Time `json:"timestamp"`
	CPUUsage    float64   `json:"cpu_usage"`
	MemoryUsage float64   `json:"memory_usage"`
	DiskUsage   float64   `json:"disk_usage"`
	NetworkIO   float64   `json:"network_io"`
}

// LoadPredictor uses AI algorithms to predict optimal load distribution
type LoadPredictor struct {
	// Neural network weights for load prediction
	weights map[string][]float64
	
	// Historical load patterns
	loadPatterns map[string]*LoadPattern
	
	// Learning parameters
	learningRate float64
	
	mu sync.RWMutex
}

// LoadPattern represents historical load patterns for prediction
type LoadPattern struct {
	ServiceID       string    `json:"service_id"`
	HourlyPatterns  [24]float64 `json:"hourly_patterns"`
	DailyPatterns   [7]float64  `json:"daily_patterns"`
	SeasonalFactors []float64   `json:"seasonal_factors"`
	LastUpdated     time.Time   `json:"last_updated"`
}

// NewIntelligentLoadBalancer creates a new intelligent load balancer
func NewIntelligentLoadBalancer() *IntelligentLoadBalancer {
	config := &LoadBalancerConfig{
		DefaultStrategy:    "ai_weighted",
		HealthCheckWeight:  0.3,
		LatencyWeight:      0.25,
		CPUWeight:          0.15,
		MemoryWeight:       0.1,
		ErrorRateWeight:    0.2,
		RequestCountWeight: 0.1,
		PredictionWindow:   5 * time.Minute,
		LearningRate:       0.01,
	}

	lb := &IntelligentLoadBalancer{
		strategies:         make(map[string]LoadBalanceStrategy),
		performanceTracker: NewPerformanceTracker(),
		loadPredictor:      NewLoadPredictor(config.LearningRate),
		config:             config,
	}

	// Register built-in strategies
	lb.registerStrategies()

	return lb
}

// NewPerformanceTracker creates a new performance tracker
func NewPerformanceTracker() *PerformanceTracker {
	return &PerformanceTracker{
		metrics: make(map[string]*ServicePerformanceHistory),
	}
}

// NewLoadPredictor creates a new load predictor
func NewLoadPredictor(learningRate float64) *LoadPredictor {
	return &LoadPredictor{
		weights:      make(map[string][]float64),
		loadPatterns: make(map[string]*LoadPattern),
		learningRate: learningRate,
	}
}

// registerStrategies registers all available load balancing strategies
func (lb *IntelligentLoadBalancer) registerStrategies() {
	lb.strategies["round_robin"] = &RoundRobinStrategy{}
	lb.strategies["least_connections"] = &LeastConnectionsStrategy{}
	lb.strategies["weighted_round_robin"] = &WeightedRoundRobinStrategy{}
	lb.strategies["least_response_time"] = &LeastResponseTimeStrategy{tracker: lb.performanceTracker}
	lb.strategies["ai_weighted"] = &AIWeightedStrategy{
		tracker:   lb.performanceTracker,
		predictor: lb.loadPredictor,
		config:    lb.config,
	}
	lb.strategies["adaptive"] = &AdaptiveStrategy{
		tracker:    lb.performanceTracker,
		predictor:  lb.loadPredictor,
		config:     lb.config,
		strategies: lb.strategies,
	}
}

// SelectInstance selects the best service instance for a request
func (lb *IntelligentLoadBalancer) SelectInstance(services []*ServiceInstance, request *LoadBalanceRequest) (*ServiceInstance, error) {
	if len(services) == 0 {
		return nil, fmt.Errorf("no services available")
	}

	// Determine strategy based on request or use default
	strategyName := lb.config.DefaultStrategy
	if request != nil && request.Headers != nil {
		if customStrategy, exists := request.Headers["X-LB-Strategy"]; exists {
			strategyName = customStrategy
		}
	}

	strategy, exists := lb.strategies[strategyName]
	if !exists {
		log.Printf("⚠️  Unknown load balancing strategy: %s, using default", strategyName)
		strategy = lb.strategies[lb.config.DefaultStrategy]
	}

	// Select instance using the chosen strategy
	instance, err := strategy.SelectInstance(services, request)
	if err != nil {
		return nil, fmt.Errorf("failed to select instance using %s strategy: %w", strategyName, err)
	}

	// Record the selection for learning
	lb.recordSelection(instance, request)

	return instance, nil
}

// recordSelection records a load balancing decision for learning
func (lb *IntelligentLoadBalancer) recordSelection(instance *ServiceInstance, request *LoadBalanceRequest) {
	go func() {
		// Update performance tracking
		lb.performanceTracker.RecordRequest(instance.ID, request)
		
		// Update load predictor
		lb.loadPredictor.RecordSelection(instance.ID, request)
	}()
}

// RecordResponse records the response metrics for learning
func (lb *IntelligentLoadBalancer) RecordResponse(serviceID string, responseTime float64, success bool) {
	lb.performanceTracker.RecordResponse(serviceID, responseTime, success)
	lb.loadPredictor.UpdatePerformance(serviceID, responseTime, success)
}

// GetStrategy returns a load balancing strategy by name
func (lb *IntelligentLoadBalancer) GetStrategy(name string) (LoadBalanceStrategy, bool) {
	lb.mu.RLock()
	defer lb.mu.RUnlock()
	
	strategy, exists := lb.strategies[name]
	return strategy, exists
}

// RegisterStrategy registers a custom load balancing strategy
func (lb *IntelligentLoadBalancer) RegisterStrategy(name string, strategy LoadBalanceStrategy) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	
	lb.strategies[name] = strategy
	log.Printf("📊 Registered load balancing strategy: %s", name)
}

// GetPerformanceStats returns performance statistics for all services
func (lb *IntelligentLoadBalancer) GetPerformanceStats() map[string]*ServicePerformanceHistory {
	return lb.performanceTracker.GetAllMetrics()
}

// === Load Balancing Strategies ===

// RoundRobinStrategy implements round-robin load balancing
type RoundRobinStrategy struct {
	mu      sync.Mutex
	counter int
}

func (rr *RoundRobinStrategy) Name() string { return "round_robin" }

func (rr *RoundRobinStrategy) SelectInstance(services []*ServiceInstance, request *LoadBalanceRequest) (*ServiceInstance, error) {
	rr.mu.Lock()
	defer rr.mu.Unlock()
	
	index := rr.counter % len(services)
	rr.counter++
	
	return services[index], nil
}

// LeastConnectionsStrategy selects the service with the least connections
type LeastConnectionsStrategy struct{}

func (lc *LeastConnectionsStrategy) Name() string { return "least_connections" }

func (lc *LeastConnectionsStrategy) SelectInstance(services []*ServiceInstance, request *LoadBalanceRequest) (*ServiceInstance, error) {
	var bestService *ServiceInstance
	minConnections := float64(math.MaxFloat64)
	
	for _, service := range services {
		connections := service.Metrics.RequestsPerSecond // Use RPS as a proxy for connections
		if connections < minConnections {
			minConnections = connections
			bestService = service
		}
	}
	
	if bestService == nil {
		return services[0], nil // Fallback to first service
	}
	
	return bestService, nil
}

// WeightedRoundRobinStrategy implements weighted round-robin based on capacity
type WeightedRoundRobinStrategy struct{}

func (wrr *WeightedRoundRobinStrategy) Name() string { return "weighted_round_robin" }

func (wrr *WeightedRoundRobinStrategy) SelectInstance(services []*ServiceInstance, request *LoadBalanceRequest) (*ServiceInstance, error) {
	// Calculate weights based on inverse of current load
	totalWeight := 0.0
	weights := make([]float64, len(services))
	
	for i, service := range services {
		// Weight inversely proportional to CPU usage
		weight := 100.0 - service.Metrics.CPUUsage
		if weight <= 0 {
			weight = 1.0
		}
		weights[i] = weight
		totalWeight += weight
	}
	
	// Select based on weighted random selection
	random := rand.Float64() * totalWeight
	cumulative := 0.0
	
	for i, weight := range weights {
		cumulative += weight
		if random <= cumulative {
			return services[i], nil
		}
	}
	
	return services[len(services)-1], nil // Fallback to last service
}

// LeastResponseTimeStrategy selects the service with the lowest response time
type LeastResponseTimeStrategy struct {
	tracker *PerformanceTracker
}

func (lrt *LeastResponseTimeStrategy) Name() string { return "least_response_time" }

func (lrt *LeastResponseTimeStrategy) SelectInstance(services []*ServiceInstance, request *LoadBalanceRequest) (*ServiceInstance, error) {
	var bestService *ServiceInstance
	minResponseTime := float64(math.MaxFloat64)
	
	for _, service := range services {
		avgResponseTime := lrt.tracker.GetAverageResponseTime(service.ID)
		if avgResponseTime < minResponseTime {
			minResponseTime = avgResponseTime
			bestService = service
		}
	}
	
	if bestService == nil {
		return services[0], nil // Fallback to first service
	}
	
	return bestService, nil
}

// AIWeightedStrategy uses AI to calculate optimal weights
type AIWeightedStrategy struct {
	tracker   *PerformanceTracker
	predictor *LoadPredictor
	config    *LoadBalancerConfig
}

func (ai *AIWeightedStrategy) Name() string { return "ai_weighted" }

func (ai *AIWeightedStrategy) SelectInstance(services []*ServiceInstance, request *LoadBalanceRequest) (*ServiceInstance, error) {
	scores := make([]float64, len(services))
	
	for i, service := range services {
		score := ai.calculateAIScore(service, request)
		scores[i] = score
	}
	
	// Select the service with the highest score
	maxScore := float64(-1)
	var bestIndex int
	
	for i, score := range scores {
		if score > maxScore {
			maxScore = score
			bestIndex = i
		}
	}
	
	return services[bestIndex], nil
}

// calculateAIScore calculates an AI-driven score for service selection
func (ai *AIWeightedStrategy) calculateAIScore(service *ServiceInstance, request *LoadBalanceRequest) float64 {
	score := 0.0
	
	// Health score
	healthScore := ai.getHealthScore(service.Health)
	score += healthScore * ai.config.HealthCheckWeight
	
	// Latency score (lower is better)
	latencyScore := 1.0 / (1.0 + service.Metrics.AverageLatency/100.0)
	score += latencyScore * ai.config.LatencyWeight
	
	// CPU utilization score (lower is better)
	cpuScore := 1.0 - (service.Metrics.CPUUsage / 100.0)
	score += cpuScore * ai.config.CPUWeight
	
	// Memory utilization score (lower is better)
	memoryScore := 1.0 - (service.Metrics.MemoryUsage / 100.0)
	score += memoryScore * ai.config.MemoryWeight
	
	// Error rate score (lower is better)
	errorScore := 1.0 - service.Metrics.ErrorRate
	score += errorScore * ai.config.ErrorRateWeight
	
	// Request count score (balanced load)
	avgRPS := ai.getAverageRPS()
	if avgRPS > 0 {
		requestScore := 1.0 - math.Abs(service.Metrics.RequestsPerSecond-avgRPS)/avgRPS
		score += requestScore * ai.config.RequestCountWeight
	}
	
	// Apply AI prediction boost
	prediction := ai.predictor.PredictLoad(service.ID, request)
	score += prediction * 0.1 // 10% weight for prediction
	
	return score
}

// getHealthScore converts health status to numeric score
func (ai *AIWeightedStrategy) getHealthScore(health HealthStatus) float64 {
	switch health {
	case HealthStatusHealthy:
		return 1.0
	case HealthStatusUnhealthy:
		return 0.5
	case HealthStatusCritical:
		return 0.1
	default:
		return 0.0
	}
}

// getAverageRPS calculates the average requests per second across all services
func (ai *AIWeightedStrategy) getAverageRPS() float64 {
	// This would typically calculate from all tracked services
	return 100.0 // Placeholder
}

// AdaptiveStrategy dynamically selects the best strategy based on conditions
type AdaptiveStrategy struct {
	tracker    *PerformanceTracker
	predictor  *LoadPredictor
	config     *LoadBalancerConfig
	strategies map[string]LoadBalanceStrategy
	mu         sync.RWMutex
}

func (as *AdaptiveStrategy) Name() string { return "adaptive" }

func (as *AdaptiveStrategy) SelectInstance(services []*ServiceInstance, request *LoadBalanceRequest) (*ServiceInstance, error) {
	// Analyze current conditions and select the best strategy
	bestStrategy := as.selectBestStrategy(services, request)
	
	return bestStrategy.SelectInstance(services, request)
}

// selectBestStrategy dynamically selects the best strategy
func (as *AdaptiveStrategy) selectBestStrategy(services []*ServiceInstance, request *LoadBalanceRequest) LoadBalanceStrategy {
	// Analyze system conditions
	avgCPU := as.getAverageCPU(services)
	avgLatency := as.getAverageLatency(services)
	loadVariance := as.getLoadVariance(services)
	
	// Select strategy based on conditions
	if avgCPU > 80.0 {
		// High CPU - use least connections
		return as.strategies["least_connections"]
	} else if avgLatency > 1000.0 {
		// High latency - use least response time
		return as.strategies["least_response_time"]
	} else if loadVariance > 0.5 {
		// High load variance - use AI weighted
		return as.strategies["ai_weighted"]
	} else {
		// Normal conditions - use weighted round robin
		return as.strategies["weighted_round_robin"]
	}
}

// getAverageCPU calculates average CPU usage across services
func (as *AdaptiveStrategy) getAverageCPU(services []*ServiceInstance) float64 {
	total := 0.0
	for _, service := range services {
		total += service.Metrics.CPUUsage
	}
	return total / float64(len(services))
}

// getAverageLatency calculates average latency across services
func (as *AdaptiveStrategy) getAverageLatency(services []*ServiceInstance) float64 {
	total := 0.0
	for _, service := range services {
		total += service.Metrics.AverageLatency
	}
	return total / float64(len(services))
}

// getLoadVariance calculates load variance across services
func (as *AdaptiveStrategy) getLoadVariance(services []*ServiceInstance) float64 {
	if len(services) <= 1 {
		return 0.0
	}
	
	loads := make([]float64, len(services))
	sum := 0.0
	
	for i, service := range services {
		load := service.Metrics.RequestsPerSecond
		loads[i] = load
		sum += load
	}
	
	mean := sum / float64(len(services))
	variance := 0.0
	
	for _, load := range loads {
		variance += math.Pow(load-mean, 2)
	}
	
	return variance / float64(len(services))
}

// === Performance Tracker Methods ===

// RecordRequest records a new request
func (pt *PerformanceTracker) RecordRequest(serviceID string, request *LoadBalanceRequest) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	
	if _, exists := pt.metrics[serviceID]; !exists {
		pt.metrics[serviceID] = &ServicePerformanceHistory{
			ServiceID:   serviceID,
			LastUpdated: time.Now(),
		}
	}
	
	// Record throughput
	pt.metrics[serviceID].ThroughputRates = append(pt.metrics[serviceID].ThroughputRates, ThroughputMetric{
		Timestamp:         time.Now(),
		RequestsPerSecond: 1.0, // Will be aggregated later
	})
}

// RecordResponse records a response
func (pt *PerformanceTracker) RecordResponse(serviceID string, responseTime float64, success bool) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	
	if _, exists := pt.metrics[serviceID]; !exists {
		pt.metrics[serviceID] = &ServicePerformanceHistory{
			ServiceID:   serviceID,
			LastUpdated: time.Now(),
		}
	}
	
	history := pt.metrics[serviceID]
	
	// Record response time
	history.ResponseTimes = append(history.ResponseTimes, ResponseTimeMetric{
		Timestamp:    time.Now(),
		ResponseTime: responseTime,
	})
	
	// Update success/failure counts
	if success {
		history.SuccessCount++
	} else {
		history.FailureCount++
	}
	
	// Calculate and record error rate
	totalRequests := history.SuccessCount + history.FailureCount
	errorRate := 0.0
	if totalRequests > 0 {
		errorRate = float64(history.FailureCount) / float64(totalRequests)
	}
	
	history.ErrorRates = append(history.ErrorRates, ErrorRateMetric{
		Timestamp: time.Now(),
		ErrorRate: errorRate,
	})
	
	history.LastUpdated = time.Now()
	
	// Keep only recent metrics (last 1000 entries)
	pt.trimMetrics(history)
}

// GetAverageResponseTime returns the average response time for a service
func (pt *PerformanceTracker) GetAverageResponseTime(serviceID string) float64 {
	pt.mu.RLock()
	defer pt.mu.RUnlock()
	
	history, exists := pt.metrics[serviceID]
	if !exists || len(history.ResponseTimes) == 0 {
		return 0.0
	}
	
	total := 0.0
	count := 0
	cutoff := time.Now().Add(-5 * time.Minute) // Last 5 minutes
	
	for _, metric := range history.ResponseTimes {
		if metric.Timestamp.After(cutoff) {
			total += metric.ResponseTime
			count++
		}
	}
	
	if count == 0 {
		return 0.0
	}
	
	return total / float64(count)
}

// GetAllMetrics returns all performance metrics
func (pt *PerformanceTracker) GetAllMetrics() map[string]*ServicePerformanceHistory {
	pt.mu.RLock()
	defer pt.mu.RUnlock()
	
	// Return a copy to avoid race conditions
	result := make(map[string]*ServicePerformanceHistory)
	for k, v := range pt.metrics {
		result[k] = v
	}
	
	return result
}

// trimMetrics keeps only the most recent metrics
func (pt *PerformanceTracker) trimMetrics(history *ServicePerformanceHistory) {
	maxEntries := 1000
	
	if len(history.ResponseTimes) > maxEntries {
		history.ResponseTimes = history.ResponseTimes[len(history.ResponseTimes)-maxEntries:]
	}
	
	if len(history.ErrorRates) > maxEntries {
		history.ErrorRates = history.ErrorRates[len(history.ErrorRates)-maxEntries:]
	}
	
	if len(history.ThroughputRates) > maxEntries {
		history.ThroughputRates = history.ThroughputRates[len(history.ThroughputRates)-maxEntries:]
	}
}

// === Load Predictor Methods ===

// RecordSelection records a load balancing selection
func (lp *LoadPredictor) RecordSelection(serviceID string, request *LoadBalanceRequest) {
	lp.mu.Lock()
	defer lp.mu.Unlock()
	
	// Initialize pattern if not exists
	if _, exists := lp.loadPatterns[serviceID]; !exists {
		lp.loadPatterns[serviceID] = &LoadPattern{
			ServiceID:   serviceID,
			LastUpdated: time.Now(),
		}
	}
	
	// Update hourly pattern
	hour := time.Now().Hour()
	lp.loadPatterns[serviceID].HourlyPatterns[hour] += 1.0
	
	// Update daily pattern
	weekday := int(time.Now().Weekday())
	lp.loadPatterns[serviceID].DailyPatterns[weekday] += 1.0
}

// UpdatePerformance updates performance data for prediction model
func (lp *LoadPredictor) UpdatePerformance(serviceID string, responseTime float64, success bool) {
	lp.mu.Lock()
	defer lp.mu.Unlock()
	
	// Initialize weights if not exists
	if _, exists := lp.weights[serviceID]; !exists {
		lp.weights[serviceID] = make([]float64, 4) // [latency, success_rate, hour, day]
	}
	
	// Simple gradient descent update
	weights := lp.weights[serviceID]
	
	// Update latency weight
	weights[0] += lp.learningRate * (1.0/responseTime - weights[0])
	
	// Update success rate weight
	successValue := 0.0
	if success {
		successValue = 1.0
	}
	weights[1] += lp.learningRate * (successValue - weights[1])
	
	// Update temporal weights
	hour := float64(time.Now().Hour()) / 24.0
	day := float64(time.Now().Weekday()) / 7.0
	weights[2] += lp.learningRate * (hour - weights[2])
	weights[3] += lp.learningRate * (day - weights[3])
}

// PredictLoad predicts the optimal load for a service
func (lp *LoadPredictor) PredictLoad(serviceID string, request *LoadBalanceRequest) float64 {
	lp.mu.RLock()
	defer lp.mu.RUnlock()
	
	weights, exists := lp.weights[serviceID]
	if !exists {
		return 0.5 // Neutral prediction
	}
	
	pattern, patternExists := lp.loadPatterns[serviceID]
	if !patternExists {
		return 0.5 // Neutral prediction
	}
	
	// Simple prediction based on time patterns and weights
	hour := time.Now().Hour()
	day := int(time.Now().Weekday())
	
	hourlyScore := pattern.HourlyPatterns[hour] / 100.0 // Normalize
	dailyScore := pattern.DailyPatterns[day] / 100.0    // Normalize
	
	// Weighted combination
	prediction := weights[0]*0.3 + weights[1]*0.3 + hourlyScore*weights[2]*0.2 + dailyScore*weights[3]*0.2
	
	// Clamp to [0, 1]
	if prediction < 0 {
		prediction = 0
	}
	if prediction > 1 {
		prediction = 1
	}
	
	return prediction
}