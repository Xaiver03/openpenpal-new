package loadbalancer

import (
	"api-gateway/internal/discovery"
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"go.uber.org/zap"
)

// LoadBalancerConfig holds configuration for the load balancer
type LoadBalancerConfig struct {
	DefaultAlgorithm   string            `json:"default_algorithm"`
	ServiceAlgorithms  map[string]string `json:"service_algorithms"`
	HealthCheckEnabled bool              `json:"health_check_enabled"`
	MetricsEnabled     bool              `json:"metrics_enabled"`
	
	// Circuit breaker integration
	CircuitBreakerEnabled bool `json:"circuit_breaker_enabled"`
	
	// Session affinity
	SessionAffinityEnabled bool          `json:"session_affinity_enabled"`
	SessionAffinityTTL     time.Duration `json:"session_affinity_ttl"`
	
	// Performance monitoring
	PerformanceMonitoringEnabled bool          `json:"performance_monitoring_enabled"`
	PerformanceMonitoringWindow  time.Duration `json:"performance_monitoring_window"`
	
	// Gradual recovery settings
	GradualRecoveryEnabled   bool    `json:"gradual_recovery_enabled"`
	RecoveryThreshold        float64 `json:"recovery_threshold"`
	RecoveryStepSize         float64 `json:"recovery_step_size"`
	RecoveryCheckInterval    time.Duration `json:"recovery_check_interval"`
}

// LoadBalancerManager manages load balancing across services
type LoadBalancerManager struct {
	config           *LoadBalancerConfig
	serviceDiscovery *discovery.ServiceDiscovery
	logger           *zap.Logger
	
	// Service-specific load balancers
	serviceBalancers map[string]LoadBalancingAlgorithm
	serviceInstances map[string][]*ServiceInstance
	
	// Session affinity
	sessionStore *SessionAffinityStore
	
	// Circuit breaker integration
	circuitBreakerManager CircuitBreakerIntegration
	
	// Metrics collection
	metricsCollector MetricsCollector
	
	// Gradual recovery
	recoveryManager *GradualRecoveryManager
	
	mutex sync.RWMutex
	ctx   context.Context
	cancel context.CancelFunc
}

// CircuitBreakerIntegration defines interface for circuit breaker integration
type CircuitBreakerIntegration interface {
	IsServiceAvailable(serviceName string) bool
	MarkServiceFailure(serviceName string, instance *ServiceInstance)
	MarkServiceSuccess(serviceName string, instance *ServiceInstance)
}

// MetricsCollector defines interface for metrics collection
type MetricsCollector interface {
	RecordLoadBalancerMetrics(serviceName, algorithm string, instance *ServiceInstance, responseTime time.Duration, success bool)
	RecordServiceInstanceMetrics(serviceName string, instance *ServiceInstance)
}

// SessionAffinityStore manages session-to-instance mappings
type SessionAffinityStore struct {
	sessions map[string]*SessionMapping
	mutex    sync.RWMutex
	ttl      time.Duration
}

type SessionMapping struct {
	InstanceHost string    `json:"instance_host"`
	CreatedAt    time.Time `json:"created_at"`
	LastUsed     time.Time `json:"last_used"`
}

// GradualRecoveryManager manages gradual recovery of unhealthy instances
type GradualRecoveryManager struct {
	recoveringInstances map[string]*RecoveryState
	mutex               sync.RWMutex
	config              *LoadBalancerConfig
	logger              *zap.Logger
}

type RecoveryState struct {
	Instance          *ServiceInstance `json:"instance"`
	RecoveryWeight    float64          `json:"recovery_weight"`
	SuccessfulChecks  int              `json:"successful_checks"`
	LastCheckTime     time.Time        `json:"last_check_time"`
	RecoveryStartTime time.Time        `json:"recovery_start_time"`
}

// NewLoadBalancerManager creates a new load balancer manager
func NewLoadBalancerManager(config *LoadBalancerConfig, serviceDiscovery *discovery.ServiceDiscovery, logger *zap.Logger) *LoadBalancerManager {
	ctx, cancel := context.WithCancel(context.Background())
	
	lbm := &LoadBalancerManager{
		config:           config,
		serviceDiscovery: serviceDiscovery,
		logger:           logger,
		serviceBalancers: make(map[string]LoadBalancingAlgorithm),
		serviceInstances: make(map[string][]*ServiceInstance),
		ctx:              ctx,
		cancel:           cancel,
	}
	
	// Initialize session affinity store
	if config.SessionAffinityEnabled {
		lbm.sessionStore = &SessionAffinityStore{
			sessions: make(map[string]*SessionMapping),
			ttl:      config.SessionAffinityTTL,
		}
	}
	
	// Initialize gradual recovery manager
	if config.GradualRecoveryEnabled {
		lbm.recoveryManager = &GradualRecoveryManager{
			recoveringInstances: make(map[string]*RecoveryState),
			config:              config,
			logger:              logger,
		}
	}
	
	// Initialize service balancers
	lbm.initializeServiceBalancers()
	
	return lbm
}

// SetCircuitBreakerIntegration sets the circuit breaker integration
func (lbm *LoadBalancerManager) SetCircuitBreakerIntegration(cbi CircuitBreakerIntegration) {
	lbm.circuitBreakerManager = cbi
}

// SetMetricsCollector sets the metrics collector
func (lbm *LoadBalancerManager) SetMetricsCollector(mc MetricsCollector) {
	lbm.metricsCollector = mc
}

// Start starts the load balancer manager
func (lbm *LoadBalancerManager) Start() error {
	lbm.logger.Info("Starting load balancer manager")
	
	// Start session affinity cleanup if enabled
	if lbm.config.SessionAffinityEnabled {
		go lbm.sessionAffinityCleanup()
	}
	
	// Start gradual recovery monitoring if enabled
	if lbm.config.GradualRecoveryEnabled {
		go lbm.gradualRecoveryMonitoring()
	}
	
	// Start performance monitoring if enabled
	if lbm.config.PerformanceMonitoringEnabled {
		go lbm.performanceMonitoring()
	}
	
	return nil
}

// Stop stops the load balancer manager
func (lbm *LoadBalancerManager) Stop() error {
	lbm.logger.Info("Stopping load balancer manager")
	lbm.cancel()
	return nil
}

// SelectInstance selects an instance for the given service using load balancing
func (lbm *LoadBalancerManager) SelectInstance(serviceName, sessionID string) (*ServiceInstance, error) {
	// Check session affinity first
	if lbm.config.SessionAffinityEnabled && sessionID != "" {
		if instance := lbm.getSessionAffinityInstance(serviceName, sessionID); instance != nil {
			return instance, nil
		}
	}
	
	// Get available instances
	instances, err := lbm.getAvailableInstances(serviceName)
	if err != nil {
		return nil, fmt.Errorf("failed to get available instances: %w", err)
	}
	
	if len(instances) == 0 {
		return nil, fmt.Errorf("no available instances for service %s", serviceName)
	}
	
	// Get the appropriate load balancer
	balancer := lbm.getServiceBalancer(serviceName)
	
	// Select instance using load balancing algorithm
	selectedInstance := balancer.Select(instances)
	if selectedInstance == nil {
		return nil, fmt.Errorf("load balancer failed to select instance for service %s", serviceName)
	}
	
	// Store session affinity if enabled
	if lbm.config.SessionAffinityEnabled && sessionID != "" {
		lbm.setSessionAffinity(sessionID, selectedInstance)
	}
	
	// Increment connection count
	selectedInstance.IncrementConnections()
	
	return selectedInstance, nil
}

// UpdateInstanceStats updates instance statistics after a request
func (lbm *LoadBalancerManager) UpdateInstanceStats(serviceName string, instance *ServiceInstance, responseTime time.Duration, success bool) {
	// Decrement connection count
	instance.DecrementConnections()
	
	// Update instance statistics
	balancer := lbm.getServiceBalancer(serviceName)
	balancer.UpdateInstanceStats(instance, responseTime, success)
	
	// Update circuit breaker if enabled
	if lbm.config.CircuitBreakerEnabled && lbm.circuitBreakerManager != nil {
		if success {
			lbm.circuitBreakerManager.MarkServiceSuccess(serviceName, instance)
		} else {
			lbm.circuitBreakerManager.MarkServiceFailure(serviceName, instance)
		}
	}
	
	// Update gradual recovery if enabled
	if lbm.config.GradualRecoveryEnabled {
		lbm.updateRecoveryStats(instance, success)
	}
	
	// Collect metrics if enabled
	if lbm.config.MetricsEnabled && lbm.metricsCollector != nil {
		algorithm := balancer.GetAlgorithmType()
		lbm.metricsCollector.RecordLoadBalancerMetrics(serviceName, algorithm, instance, responseTime, success)
	}
}

// initializeServiceBalancers initializes load balancers for each service
func (lbm *LoadBalancerManager) initializeServiceBalancers() {
	lbm.mutex.Lock()
	defer lbm.mutex.Unlock()
	
	// Get all services from service discovery
	for serviceName := range lbm.config.ServiceAlgorithms {
		algorithm := lbm.config.ServiceAlgorithms[serviceName]
		if algorithm == "" {
			algorithm = lbm.config.DefaultAlgorithm
		}
		
		lbm.serviceBalancers[serviceName] = CreateAlgorithm(algorithm)
		lbm.logger.Info("Initialized load balancer",
			zap.String("service", serviceName),
			zap.String("algorithm", algorithm),
		)
	}
}

// getAvailableInstances gets available instances for a service
func (lbm *LoadBalancerManager) getAvailableInstances(serviceName string) ([]*ServiceInstance, error) {
	// Get instances from service discovery
	discoveryInstances := lbm.serviceDiscovery.GetServiceInstances(serviceName)
	if len(discoveryInstances) == 0 {
		return nil, fmt.Errorf("no instances found for service %s", serviceName)
	}
	
	// Convert to enhanced service instances
	var instances []*ServiceInstance
	for _, discoveryInstance := range discoveryInstances {
		// Only include healthy instances
		if !discoveryInstance.Healthy {
			continue
		}
		
		// Check circuit breaker if enabled
		if lbm.config.CircuitBreakerEnabled && lbm.circuitBreakerManager != nil {
			if !lbm.circuitBreakerManager.IsServiceAvailable(serviceName) {
				continue
			}
		}
		
		// Wrap or get existing enhanced instance
		instance := lbm.getOrCreateEnhancedInstance(discoveryInstance)
		instances = append(instances, instance)
	}
	
	// Add recovering instances if gradual recovery is enabled
	if lbm.config.GradualRecoveryEnabled {
		recoveringInstances := lbm.getRecoveringInstances(serviceName)
		instances = append(instances, recoveringInstances...)
	}
	
	return instances, nil
}

// getOrCreateEnhancedInstance gets or creates an enhanced service instance
func (lbm *LoadBalancerManager) getOrCreateEnhancedInstance(discoveryInstance *discovery.ServiceInstance) *ServiceInstance {
	lbm.mutex.Lock()
	defer lbm.mutex.Unlock()
	
	serviceInstances := lbm.serviceInstances[discoveryInstance.Name]
	
	// Look for existing instance
	for _, instance := range serviceInstances {
		if instance.Host == discoveryInstance.Host {
			// Update the underlying discovery instance
			instance.ServiceInstance = discoveryInstance
			return instance
		}
	}
	
	// Create new enhanced instance
	newInstance := WrapServiceInstance(discoveryInstance)
	
	// Add to service instances
	if lbm.serviceInstances[discoveryInstance.Name] == nil {
		lbm.serviceInstances[discoveryInstance.Name] = make([]*ServiceInstance, 0)
	}
	lbm.serviceInstances[discoveryInstance.Name] = append(lbm.serviceInstances[discoveryInstance.Name], newInstance)
	
	return newInstance
}

// getServiceBalancer gets the load balancer for a service
func (lbm *LoadBalancerManager) getServiceBalancer(serviceName string) LoadBalancingAlgorithm {
	lbm.mutex.RLock()
	defer lbm.mutex.RUnlock()
	
	if balancer, exists := lbm.serviceBalancers[serviceName]; exists {
		return balancer
	}
	
	// Create default balancer if not exists
	algorithm := lbm.config.DefaultAlgorithm
	balancer := CreateAlgorithm(algorithm)
	
	lbm.mutex.RUnlock()
	lbm.mutex.Lock()
	lbm.serviceBalancers[serviceName] = balancer
	lbm.mutex.Unlock()
	lbm.mutex.RLock()
	
	return balancer
}

// Session Affinity methods

func (lbm *LoadBalancerManager) getSessionAffinityInstance(serviceName, sessionID string) *ServiceInstance {
	if lbm.sessionStore == nil {
		return nil
	}
	
	lbm.sessionStore.mutex.RLock()
	defer lbm.sessionStore.mutex.RUnlock()
	
	mapping, exists := lbm.sessionStore.sessions[sessionID]
	if !exists {
		return nil
	}
	
	// Check if session has expired
	if time.Since(mapping.LastUsed) > lbm.sessionStore.ttl {
		return nil
	}
	
	// Find the instance
	serviceInstances := lbm.serviceInstances[serviceName]
	for _, instance := range serviceInstances {
		if instance.Host == mapping.InstanceHost && instance.Healthy {
			mapping.LastUsed = time.Now()
			return instance
		}
	}
	
	return nil
}

func (lbm *LoadBalancerManager) setSessionAffinity(sessionID string, instance *ServiceInstance) {
	if lbm.sessionStore == nil {
		return
	}
	
	lbm.sessionStore.mutex.Lock()
	defer lbm.sessionStore.mutex.Unlock()
	
	lbm.sessionStore.sessions[sessionID] = &SessionMapping{
		InstanceHost: instance.Host,
		CreatedAt:    time.Now(),
		LastUsed:     time.Now(),
	}
}

func (lbm *LoadBalancerManager) sessionAffinityCleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			lbm.cleanupExpiredSessions()
		case <-lbm.ctx.Done():
			return
		}
	}
}

func (lbm *LoadBalancerManager) cleanupExpiredSessions() {
	if lbm.sessionStore == nil {
		return
	}
	
	lbm.sessionStore.mutex.Lock()
	defer lbm.sessionStore.mutex.Unlock()
	
	now := time.Now()
	for sessionID, mapping := range lbm.sessionStore.sessions {
		if now.Sub(mapping.LastUsed) > lbm.sessionStore.ttl {
			delete(lbm.sessionStore.sessions, sessionID)
		}
	}
}

// Gradual Recovery methods

func (lbm *LoadBalancerManager) getRecoveringInstances(serviceName string) []*ServiceInstance {
	if lbm.recoveryManager == nil {
		return nil
	}
	
	lbm.recoveryManager.mutex.RLock()
	defer lbm.recoveryManager.mutex.RUnlock()
	
	var instances []*ServiceInstance
	for _, recoveryState := range lbm.recoveryManager.recoveringInstances {
		if recoveryState.Instance.Name == serviceName {
			instances = append(instances, recoveryState.Instance)
		}
	}
	
	return instances
}

func (lbm *LoadBalancerManager) updateRecoveryStats(instance *ServiceInstance, success bool) {
	if lbm.recoveryManager == nil {
		return
	}
	
	lbm.recoveryManager.mutex.Lock()
	defer lbm.recoveryManager.mutex.Unlock()
	
	key := fmt.Sprintf("%s-%s", instance.Name, instance.Host)
	if recoveryState, exists := lbm.recoveryManager.recoveringInstances[key]; exists {
		recoveryState.LastCheckTime = time.Now()
		
		if success {
			recoveryState.SuccessfulChecks++
			
			// Gradually increase weight
			if recoveryState.RecoveryWeight < 1.0 {
				recoveryState.RecoveryWeight = math.Min(1.0, recoveryState.RecoveryWeight+lbm.config.RecoveryStepSize)
				instance.Weight = int(float64(instance.Weight) * recoveryState.RecoveryWeight)
			}
			
			// Remove from recovery if fully recovered
			if recoveryState.RecoveryWeight >= 1.0 && recoveryState.SuccessfulChecks >= 5 {
				delete(lbm.recoveryManager.recoveringInstances, key)
				lbm.logger.Info("Instance fully recovered",
					zap.String("service", instance.Name),
					zap.String("host", instance.Host),
				)
			}
		} else {
			recoveryState.SuccessfulChecks = 0
			recoveryState.RecoveryWeight = math.Max(0.1, recoveryState.RecoveryWeight*0.8)
			instance.Weight = int(float64(instance.Weight) * recoveryState.RecoveryWeight)
		}
	}
}

func (lbm *LoadBalancerManager) gradualRecoveryMonitoring() {
	ticker := time.NewTicker(lbm.config.RecoveryCheckInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			lbm.checkForRecoveringInstances()
		case <-lbm.ctx.Done():
			return
		}
	}
}

func (lbm *LoadBalancerManager) checkForRecoveringInstances() {
	// This would check for instances that are becoming healthy again
	// and add them to gradual recovery process
}

// Performance monitoring

func (lbm *LoadBalancerManager) performanceMonitoring() {
	ticker := time.NewTicker(lbm.config.PerformanceMonitoringWindow)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			lbm.collectPerformanceMetrics()
		case <-lbm.ctx.Done():
			return
		}
	}
}

func (lbm *LoadBalancerManager) collectPerformanceMetrics() {
	lbm.mutex.RLock()
	defer lbm.mutex.RUnlock()
	
	for serviceName, instances := range lbm.serviceInstances {
		for _, instance := range instances {
			if lbm.metricsCollector != nil {
				lbm.metricsCollector.RecordServiceInstanceMetrics(serviceName, instance)
			}
		}
	}
}

// GetLoadBalancerStats returns statistics about the load balancer
func (lbm *LoadBalancerManager) GetLoadBalancerStats() map[string]interface{} {
	lbm.mutex.RLock()
	defer lbm.mutex.RUnlock()
	
	stats := make(map[string]interface{})
	
	for serviceName, instances := range lbm.serviceInstances {
		serviceStats := make(map[string]interface{})
		
		var totalRequests, successRequests, failedRequests int64
		var totalActiveConnections int64
		var averageResponseTime time.Duration
		healthyInstances := 0
		
		for _, instance := range instances {
			if instance.Healthy {
				healthyInstances++
			}
			totalRequests += instance.TotalRequests
			successRequests += instance.SuccessRequests
			failedRequests += instance.FailedRequests
			totalActiveConnections += instance.GetActiveConnections()
			averageResponseTime += instance.AverageResponse
		}
		
		if len(instances) > 0 {
			averageResponseTime = averageResponseTime / time.Duration(len(instances))
		}
		
		algorithm := "unknown"
		if balancer, exists := lbm.serviceBalancers[serviceName]; exists {
			algorithm = balancer.GetAlgorithmType()
		}
		
		serviceStats["algorithm"] = algorithm
		serviceStats["total_instances"] = len(instances)
		serviceStats["healthy_instances"] = healthyInstances
		serviceStats["total_requests"] = totalRequests
		serviceStats["success_requests"] = successRequests
		serviceStats["failed_requests"] = failedRequests
		serviceStats["active_connections"] = totalActiveConnections
		serviceStats["average_response_time"] = averageResponseTime.String()
		
		if totalRequests > 0 {
			serviceStats["success_rate"] = float64(successRequests) / float64(totalRequests) * 100
		}
		
		stats[serviceName] = serviceStats
	}
	
	return stats
}

// DefaultLoadBalancerConfig returns a default configuration
func DefaultLoadBalancerConfig() *LoadBalancerConfig {
	return &LoadBalancerConfig{
		DefaultAlgorithm:    "adaptive",
		ServiceAlgorithms:   make(map[string]string),
		HealthCheckEnabled:  true,
		MetricsEnabled:      true,
		CircuitBreakerEnabled: true,
		SessionAffinityEnabled: true,
		SessionAffinityTTL:     30 * time.Minute,
		PerformanceMonitoringEnabled: true,
		PerformanceMonitoringWindow:  1 * time.Minute,
		GradualRecoveryEnabled:   true,
		RecoveryThreshold:        0.8,
		RecoveryStepSize:         0.1,
		RecoveryCheckInterval:    30 * time.Second,
	}
}