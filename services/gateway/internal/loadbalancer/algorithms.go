package loadbalancer

import (
	"api-gateway/internal/discovery"
	"crypto/md5"
	"fmt"
	"hash/fnv"
	"math"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// LoadBalancingAlgorithm defines the interface for load balancing algorithms
type LoadBalancingAlgorithm interface {
	Select(instances []*ServiceInstance) *ServiceInstance
	UpdateInstanceStats(instance *ServiceInstance, responseTime time.Duration, success bool)
	GetAlgorithmType() string
}

// ServiceInstance extends the discovery.ServiceInstance with additional metrics
type ServiceInstance struct {
	*discovery.ServiceInstance
	ActiveConnections int64         `json:"active_connections"`
	TotalRequests     int64         `json:"total_requests"`
	SuccessRequests   int64         `json:"success_requests"`
	FailedRequests    int64         `json:"failed_requests"`
	AverageResponse   time.Duration `json:"average_response"`
	LastResponseTime  time.Duration `json:"last_response_time"`
	LastUsed          time.Time     `json:"last_used"`
	CpuUsage          float64       `json:"cpu_usage"`
	MemoryUsage       float64       `json:"memory_usage"`
	Score             float64       `json:"score"` // Overall health score
	mutex             sync.RWMutex
}

// WrapServiceInstance wraps a discovery.ServiceInstance
func WrapServiceInstance(instance *discovery.ServiceInstance) *ServiceInstance {
	return &ServiceInstance{
		ServiceInstance:   instance,
		ActiveConnections: 0,
		TotalRequests:     0,
		SuccessRequests:   0,
		FailedRequests:    0,
		AverageResponse:   0,
		LastResponseTime:  0,
		LastUsed:          time.Now(),
		CpuUsage:          0,
		MemoryUsage:       0,
		Score:             1.0,
	}
}

// IncrementConnections increments active connections
func (si *ServiceInstance) IncrementConnections() {
	atomic.AddInt64(&si.ActiveConnections, 1)
}

// DecrementConnections decrements active connections
func (si *ServiceInstance) DecrementConnections() {
	atomic.AddInt64(&si.ActiveConnections, -1)
}

// UpdateResponseTime updates response time metrics
func (si *ServiceInstance) UpdateResponseTime(responseTime time.Duration, success bool) {
	si.mutex.Lock()
	defer si.mutex.Unlock()

	atomic.AddInt64(&si.TotalRequests, 1)
	
	if success {
		atomic.AddInt64(&si.SuccessRequests, 1)
	} else {
		atomic.AddInt64(&si.FailedRequests, 1)
	}
	
	si.LastResponseTime = responseTime
	si.LastUsed = time.Now()
	
	// Calculate moving average
	if si.AverageResponse == 0 {
		si.AverageResponse = responseTime
	} else {
		// Exponential moving average with alpha = 0.3
		si.AverageResponse = time.Duration(float64(si.AverageResponse)*0.7 + float64(responseTime)*0.3)
	}
	
	// Update health score
	si.updateHealthScore()
}

// updateHealthScore calculates an overall health score for the instance
func (si *ServiceInstance) updateHealthScore() {
	total := atomic.LoadInt64(&si.TotalRequests)
	if total == 0 {
		si.Score = 1.0
		return
	}
	
	success := atomic.LoadInt64(&si.SuccessRequests)
	failed := atomic.LoadInt64(&si.FailedRequests)
	
	// Success rate component (0-1)
	successRate := float64(success) / float64(total)
	
	// Response time component (0-1, inverted so faster = better)
	responseScore := 1.0
	if si.AverageResponse > 0 {
		// Normalize response time (assume 1s is baseline, scale accordingly)
		responseScore = math.Max(0.1, 1.0-float64(si.AverageResponse)/float64(time.Second))
	}
	
	// Load component (0-1, inverted so less load = better)
	loadScore := 1.0
	if si.ActiveConnections > 0 {
		// Normalize active connections (assume 100 connections is high load)
		loadScore = math.Max(0.1, 1.0-float64(si.ActiveConnections)/100.0)
	}
	
	// Error rate penalty
	errorRate := float64(failed) / float64(total)
	errorPenalty := 1.0 - errorRate
	
	// Weighted combination
	si.Score = (successRate*0.4 + responseScore*0.3 + loadScore*0.2 + errorPenalty*0.1)
	si.Score = math.Max(0.01, math.Min(1.0, si.Score))
}

// GetSuccessRate returns the success rate of the instance
func (si *ServiceInstance) GetSuccessRate() float64 {
	total := atomic.LoadInt64(&si.TotalRequests)
	if total == 0 {
		return 1.0
	}
	success := atomic.LoadInt64(&si.SuccessRequests)
	return float64(success) / float64(total)
}

// GetActiveConnections returns current active connections
func (si *ServiceInstance) GetActiveConnections() int64 {
	return atomic.LoadInt64(&si.ActiveConnections)
}

// 1. Round Robin Algorithm
type RoundRobinBalancer struct {
	currentIndex int64
	mutex        sync.Mutex
}

func NewRoundRobinBalancer() *RoundRobinBalancer {
	return &RoundRobinBalancer{currentIndex: 0}
}

func (rb *RoundRobinBalancer) Select(instances []*ServiceInstance) *ServiceInstance {
	if len(instances) == 0 {
		return nil
	}
	
	rb.mutex.Lock()
	index := rb.currentIndex % int64(len(instances))
	rb.currentIndex++
	rb.mutex.Unlock()
	
	return instances[index]
}

func (rb *RoundRobinBalancer) UpdateInstanceStats(instance *ServiceInstance, responseTime time.Duration, success bool) {
	instance.UpdateResponseTime(responseTime, success)
}

func (rb *RoundRobinBalancer) GetAlgorithmType() string {
	return "round_robin"
}

// 2. Weighted Round Robin Algorithm
type WeightedRoundRobinBalancer struct {
	currentWeights map[string]int
	mutex          sync.Mutex
}

func NewWeightedRoundRobinBalancer() *WeightedRoundRobinBalancer {
	return &WeightedRoundRobinBalancer{
		currentWeights: make(map[string]int),
	}
}

func (wrb *WeightedRoundRobinBalancer) Select(instances []*ServiceInstance) *ServiceInstance {
	if len(instances) == 0 {
		return nil
	}
	
	wrb.mutex.Lock()
	defer wrb.mutex.Unlock()
	
	var selected *ServiceInstance
	maxCurrentWeight := -1
	totalWeight := 0
	
	for _, instance := range instances {
		totalWeight += instance.Weight
		
		// Initialize current weight if not exists
		key := instance.Host
		if _, exists := wrb.currentWeights[key]; !exists {
			wrb.currentWeights[key] = 0
		}
		
		// Increase current weight
		wrb.currentWeights[key] += instance.Weight
		
		// Select instance with highest current weight
		if wrb.currentWeights[key] > maxCurrentWeight {
			maxCurrentWeight = wrb.currentWeights[key]
			selected = instance
		}
	}
	
	// Reduce selected instance's current weight
	if selected != nil {
		wrb.currentWeights[selected.Host] -= totalWeight
	}
	
	return selected
}

func (wrb *WeightedRoundRobinBalancer) UpdateInstanceStats(instance *ServiceInstance, responseTime time.Duration, success bool) {
	instance.UpdateResponseTime(responseTime, success)
}

func (wrb *WeightedRoundRobinBalancer) GetAlgorithmType() string {
	return "weighted_round_robin"
}

// 3. Least Connections Algorithm
type LeastConnectionsBalancer struct{}

func NewLeastConnectionsBalancer() *LeastConnectionsBalancer {
	return &LeastConnectionsBalancer{}
}

func (lcb *LeastConnectionsBalancer) Select(instances []*ServiceInstance) *ServiceInstance {
	if len(instances) == 0 {
		return nil
	}
	
	var selected *ServiceInstance
	minConnections := int64(math.MaxInt64)
	
	for _, instance := range instances {
		connections := instance.GetActiveConnections()
		if connections < minConnections {
			minConnections = connections
			selected = instance
		}
	}
	
	return selected
}

func (lcb *LeastConnectionsBalancer) UpdateInstanceStats(instance *ServiceInstance, responseTime time.Duration, success bool) {
	instance.UpdateResponseTime(responseTime, success)
}

func (lcb *LeastConnectionsBalancer) GetAlgorithmType() string {
	return "least_connections"
}

// 4. Least Response Time Algorithm
type LeastResponseTimeBalancer struct{}

func NewLeastResponseTimeBalancer() *LeastResponseTimeBalancer {
	return &LeastResponseTimeBalancer{}
}

func (lrtb *LeastResponseTimeBalancer) Select(instances []*ServiceInstance) *ServiceInstance {
	if len(instances) == 0 {
		return nil
	}
	
	var selected *ServiceInstance
	minResponseTime := time.Duration(math.MaxInt64)
	
	for _, instance := range instances {
		instance.mutex.RLock()
		responseTime := instance.AverageResponse
		instance.mutex.RUnlock()
		
		if responseTime == 0 {
			// No historical data, prefer this instance
			return instance
		}
		
		if responseTime < minResponseTime {
			minResponseTime = responseTime
			selected = instance
		}
	}
	
	return selected
}

func (lrtb *LeastResponseTimeBalancer) UpdateInstanceStats(instance *ServiceInstance, responseTime time.Duration, success bool) {
	instance.UpdateResponseTime(responseTime, success)
}

func (lrtb *LeastResponseTimeBalancer) GetAlgorithmType() string {
	return "least_response_time"
}

// 5. Health-aware Weighted Algorithm
type HealthAwareBalancer struct{}

func NewHealthAwareBalancer() *HealthAwareBalancer {
	return &HealthAwareBalancer{}
}

func (hab *HealthAwareBalancer) Select(instances []*ServiceInstance) *ServiceInstance {
	if len(instances) == 0 {
		return nil
	}
	
	// Calculate total weighted score
	totalScore := 0.0
	for _, instance := range instances {
		instance.mutex.RLock()
		score := instance.Score
		instance.mutex.RUnlock()
		totalScore += score * float64(instance.Weight)
	}
	
	if totalScore == 0 {
		// Fallback to round robin
		return instances[0]
	}
	
	// Weighted random selection based on health scores
	threshold := totalScore * float64(time.Now().UnixNano()%1000) / 1000.0
	currentSum := 0.0
	
	for _, instance := range instances {
		instance.mutex.RLock()
		score := instance.Score
		instance.mutex.RUnlock()
		currentSum += score * float64(instance.Weight)
		
		if currentSum >= threshold {
			return instance
		}
	}
	
	// Fallback to last instance
	return instances[len(instances)-1]
}

func (hab *HealthAwareBalancer) UpdateInstanceStats(instance *ServiceInstance, responseTime time.Duration, success bool) {
	instance.UpdateResponseTime(responseTime, success)
}

func (hab *HealthAwareBalancer) GetAlgorithmType() string {
	return "health_aware"
}

// 6. Consistent Hashing Algorithm
type ConsistentHashBalancer struct {
	hashRing map[uint32]*ServiceInstance
	keys     []uint32
	mutex    sync.RWMutex
}

func NewConsistentHashBalancer() *ConsistentHashBalancer {
	return &ConsistentHashBalancer{
		hashRing: make(map[uint32]*ServiceInstance),
		keys:     make([]uint32, 0),
	}
}

func (chb *ConsistentHashBalancer) Select(instances []*ServiceInstance) *ServiceInstance {
	if len(instances) == 0 {
		return nil
	}
	
	chb.updateHashRing(instances)
	
	chb.mutex.RLock()
	defer chb.mutex.RUnlock()
	
	if len(chb.keys) == 0 {
		return instances[0]
	}
	
	// Generate hash for current request (could be based on client IP, session, etc.)
	requestHash := chb.hash(fmt.Sprintf("%d", time.Now().UnixNano()))
	
	// Find the first instance whose hash is >= requestHash
	index := sort.Search(len(chb.keys), func(i int) bool {
		return chb.keys[i] >= requestHash
	})
	
	// Wrap around if we've gone past the end
	if index == len(chb.keys) {
		index = 0
	}
	
	return chb.hashRing[chb.keys[index]]
}

func (chb *ConsistentHashBalancer) updateHashRing(instances []*ServiceInstance) {
	chb.mutex.Lock()
	defer chb.mutex.Unlock()
	
	// Clear existing ring
	chb.hashRing = make(map[uint32]*ServiceInstance)
	chb.keys = make([]uint32, 0)
	
	// Add each instance multiple times to the ring (virtual nodes)
	virtualNodes := 150
	for _, instance := range instances {
		for i := 0; i < virtualNodes; i++ {
			key := fmt.Sprintf("%s-%d", instance.Host, i)
			hash := chb.hash(key)
			chb.hashRing[hash] = instance
			chb.keys = append(chb.keys, hash)
		}
	}
	
	// Sort keys for binary search
	sort.Slice(chb.keys, func(i, j int) bool {
		return chb.keys[i] < chb.keys[j]
	})
}

func (chb *ConsistentHashBalancer) hash(key string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(key))
	return h.Sum32()
}

func (chb *ConsistentHashBalancer) UpdateInstanceStats(instance *ServiceInstance, responseTime time.Duration, success bool) {
	instance.UpdateResponseTime(responseTime, success)
}

func (chb *ConsistentHashBalancer) GetAlgorithmType() string {
	return "consistent_hash"
}

// 7. Adaptive Performance-based Algorithm
type AdaptiveBalancer struct {
	performanceWindow time.Duration
	mutex             sync.RWMutex
}

func NewAdaptiveBalancer() *AdaptiveBalancer {
	return &AdaptiveBalancer{
		performanceWindow: 5 * time.Minute,
	}
}

func (ab *AdaptiveBalancer) Select(instances []*ServiceInstance) *ServiceInstance {
	if len(instances) == 0 {
		return nil
	}
	
	// Calculate performance scores for each instance
	var scoredInstances []*scoredInstance
	totalScore := 0.0
	
	for _, instance := range instances {
		score := ab.calculatePerformanceScore(instance)
		scoredInstances = append(scoredInstances, &scoredInstance{
			instance: instance,
			score:    score,
		})
		totalScore += score
	}
	
	if totalScore == 0 {
		return instances[0]
	}
	
	// Weighted random selection based on performance scores
	threshold := totalScore * float64(time.Now().UnixNano()%1000) / 1000.0
	currentSum := 0.0
	
	for _, scored := range scoredInstances {
		currentSum += scored.score
		if currentSum >= threshold {
			return scored.instance
		}
	}
	
	return instances[len(instances)-1]
}

type scoredInstance struct {
	instance *ServiceInstance
	score    float64
}

func (ab *AdaptiveBalancer) calculatePerformanceScore(instance *ServiceInstance) float64 {
	instance.mutex.RLock()
	defer instance.mutex.RUnlock()
	
	// Base score from health
	score := instance.Score
	
	// Adjust based on recent performance
	now := time.Now()
	timeSinceLastUse := now.Sub(instance.LastUsed)
	
	// Penalty for instances that haven't been used recently (they might be stale)
	if timeSinceLastUse > ab.performanceWindow {
		score *= 0.8
	}
	
	// Bonus for instances with low response time
	if instance.AverageResponse > 0 {
		responseBonus := 1.0 / (1.0 + float64(instance.AverageResponse)/float64(time.Second))
		score *= (1.0 + responseBonus*0.2)
	}
	
	// Penalty for high active connections
	connectionPenalty := 1.0 / (1.0 + float64(instance.ActiveConnections)*0.01)
	score *= connectionPenalty
	
	return math.Max(0.01, score)
}

func (ab *AdaptiveBalancer) UpdateInstanceStats(instance *ServiceInstance, responseTime time.Duration, success bool) {
	instance.UpdateResponseTime(responseTime, success)
}

func (ab *AdaptiveBalancer) GetAlgorithmType() string {
	return "adaptive"
}

// AlgorithmFactory creates load balancing algorithms
func CreateAlgorithm(algorithmType string) LoadBalancingAlgorithm {
	switch algorithmType {
	case "round_robin":
		return NewRoundRobinBalancer()
	case "weighted_round_robin":
		return NewWeightedRoundRobinBalancer()
	case "least_connections":
		return NewLeastConnectionsBalancer()
	case "least_response_time":
		return NewLeastResponseTimeBalancer()
	case "health_aware":
		return NewHealthAwareBalancer()
	case "consistent_hash":
		return NewConsistentHashBalancer()
	case "adaptive":
		return NewAdaptiveBalancer()
	default:
		return NewRoundRobinBalancer() // Default fallback
	}
}