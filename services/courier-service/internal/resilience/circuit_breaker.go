package resilience

import (
	"context"
	"courier-service/internal/errors"
	"courier-service/internal/logging"
	"sync"
	"time"
)

// CircuitState 熔断器状态
type CircuitState int

const (
	StateClosed CircuitState = iota
	StateOpen
	StateHalfOpen
)

func (s CircuitState) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateOpen:
		return "OPEN"
	case StateHalfOpen:
		return "HALF_OPEN"
	default:
		return "UNKNOWN"
	}
}

// CircuitBreakerConfig 熔断器配置
type CircuitBreakerConfig struct {
	Name             string        `json:"name"`
	MaxRequests      uint32        `json:"max_requests"`      // 半开状态最大请求数
	Interval         time.Duration `json:"interval"`          // 统计间隔
	Timeout          time.Duration `json:"timeout"`           // 开启状态超时时间
	FailureThreshold uint32        `json:"failure_threshold"` // 失败阈值
	SuccessThreshold uint32        `json:"success_threshold"` // 恢复阈值
	FailureRate      float64       `json:"failure_rate"`      // 失败率阈值
	MinRequestCount  uint32        `json:"min_request_count"` // 最小请求数
	OnStateChange    func(name string, from CircuitState, to CircuitState)
	IsFailure        func(err error) bool
}

// DefaultCircuitBreakerConfig 默认熔断器配置
func DefaultCircuitBreakerConfig(name string) CircuitBreakerConfig {
	return CircuitBreakerConfig{
		Name:             name,
		MaxRequests:      5,
		Interval:         60 * time.Second,
		Timeout:          30 * time.Second,
		FailureThreshold: 5,
		SuccessThreshold: 2,
		FailureRate:      0.5,
		MinRequestCount:  3,
		IsFailure: func(err error) bool {
			return err != nil && errors.IsRetryableError(err)
		},
	}
}

// DatabaseCircuitBreakerConfig 数据库熔断器配置
func DatabaseCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		Name:             "database",
		MaxRequests:      3,
		Interval:         30 * time.Second,
		Timeout:          60 * time.Second,
		FailureThreshold: 10,
		SuccessThreshold: 5,
		FailureRate:      0.6,
		MinRequestCount:  5,
		IsFailure: func(err error) bool {
			return errors.IsError(err, errors.CodeDatabaseError) ||
				errors.IsError(err, errors.CodeConnectionTimeout) ||
				errors.IsError(err, errors.CodeDeadlock)
		},
	}
}

// ExternalServiceCircuitBreakerConfig 外部服务熔断器配置
func ExternalServiceCircuitBreakerConfig(serviceName string) CircuitBreakerConfig {
	return CircuitBreakerConfig{
		Name:             serviceName,
		MaxRequests:      2,
		Interval:         45 * time.Second,
		Timeout:          90 * time.Second,
		FailureThreshold: 3,
		SuccessThreshold: 2,
		FailureRate:      0.4,
		MinRequestCount:  2,
		IsFailure: func(err error) bool {
			return errors.IsError(err, errors.CodeExternalServiceError) ||
				errors.IsError(err, errors.CodeServiceUnavailable) ||
				errors.IsError(err, errors.CodeConnectionTimeout)
		},
	}
}

// CircuitBreakerStats 熔断器统计信息
type CircuitBreakerStats struct {
	State                CircuitState `json:"state"`
	TotalRequests        uint64       `json:"total_requests"`
	FailureCount         uint64       `json:"failure_count"`
	SuccessCount         uint64       `json:"success_count"`
	ConsecutiveFailures  uint32       `json:"consecutive_failures"`
	ConsecutiveSuccesses uint32       `json:"consecutive_successes"`
	FailureRate          float64      `json:"failure_rate"`
	LastFailureTime      *time.Time   `json:"last_failure_time,omitempty"`
	LastSuccessTime      *time.Time   `json:"last_success_time,omitempty"`
	StateChangeTime      time.Time    `json:"state_change_time"`
	NextRetryTime        *time.Time   `json:"next_retry_time,omitempty"`
}

// CircuitBreaker 熔断器实现
type CircuitBreaker struct {
	config CircuitBreakerConfig
	logger logging.Logger
	mutex  sync.RWMutex

	// 状态
	state       CircuitState
	stateExpiry time.Time
	generation  uint64

	// 统计
	stats         CircuitBreakerStats
	requests      []requestRecord
	halfOpenCount uint32
}

// requestRecord 请求记录
type requestRecord struct {
	timestamp time.Time
	success   bool
}

// NewCircuitBreaker 创建熔断器
func NewCircuitBreaker(config CircuitBreakerConfig, logger logging.Logger) *CircuitBreaker {
	if logger == nil {
		logger = logging.GetDefaultLogger()
	}

	cb := &CircuitBreaker{
		config: config,
		logger: logger,
		state:  StateClosed,
		stats: CircuitBreakerStats{
			State:           StateClosed,
			StateChangeTime: time.Now(),
		},
		requests: make([]requestRecord, 0),
	}

	return cb
}

// Execute 执行操作
func (cb *CircuitBreaker) Execute(_ context.Context, operation func() error) error {
	generation, err := cb.beforeRequest()
	if err != nil {
		return err
	}

	defer func() {
		cb.afterRequest(generation, err)
	}()

	// 执行操作
	err = operation()
	return err
}

// ExecuteWithFallback 执行操作（带降级）
func (cb *CircuitBreaker) ExecuteWithFallback(
	ctx context.Context,
	operation func() error,
	fallback func() error,
) error {
	err := cb.Execute(ctx, operation)

	// 如果熔断器开启，执行降级操作
	if errors.IsError(err, errors.CodeCircuitBreakerOpen) && fallback != nil {
		cb.logger.Info("Circuit breaker open, executing fallback",
			"circuit", cb.config.Name,
		)
		return fallback()
	}

	return err
}

// beforeRequest 请求前检查
func (cb *CircuitBreaker) beforeRequest() (uint64, error) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	now := time.Now()

	switch cb.state {
	case StateClosed:
		// 关闭状态，允许所有请求
		return cb.generation, nil

	case StateOpen:
		// 开启状态，检查是否可以转为半开状态
		if cb.stateExpiry.Before(now) {
			cb.setState(StateHalfOpen, now)
			return cb.generation, nil
		}

		// 仍在开启状态，拒绝请求
		return cb.generation, errors.NewCircuitBreakerError(cb.config.Name)

	case StateHalfOpen:
		// 半开状态，限制请求数
		if cb.halfOpenCount >= cb.config.MaxRequests {
			return cb.generation, errors.NewCircuitBreakerError(cb.config.Name)
		}

		cb.halfOpenCount++
		return cb.generation, nil

	default:
		return cb.generation, errors.NewCircuitBreakerError(cb.config.Name)
	}
}

// afterRequest 请求后处理
func (cb *CircuitBreaker) afterRequest(generation uint64, err error) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	now := time.Now()

	// 检查是否是同一代的请求
	if generation != cb.generation {
		return
	}

	// 判断请求是否失败
	isFailure := cb.config.IsFailure(err)

	// 记录请求
	cb.recordRequest(now, !isFailure)

	// 更新统计
	cb.updateStats(now, !isFailure)

	// 根据状态和结果决定状态转换
	switch cb.state {
	case StateClosed:
		if isFailure {
			cb.onFailure(now)
		} else {
			cb.onSuccess()
		}

	case StateHalfOpen:
		if isFailure {
			cb.setState(StateOpen, now)
		} else {
			cb.onSuccess()
			if cb.stats.ConsecutiveSuccesses >= cb.config.SuccessThreshold {
				cb.setState(StateClosed, now)
			}
		}
	}
}

// recordRequest 记录请求
func (cb *CircuitBreaker) recordRequest(timestamp time.Time, success bool) {
	// 清理过期的记录
	cutoff := timestamp.Add(-cb.config.Interval)
	i := 0
	for i < len(cb.requests) && cb.requests[i].timestamp.Before(cutoff) {
		i++
	}
	cb.requests = cb.requests[i:]

	// 添加新记录
	cb.requests = append(cb.requests, requestRecord{
		timestamp: timestamp,
		success:   success,
	})
}

// updateStats 更新统计信息
func (cb *CircuitBreaker) updateStats(timestamp time.Time, success bool) {
	cb.stats.TotalRequests++

	if success {
		cb.stats.SuccessCount++
		cb.stats.ConsecutiveSuccesses++
		cb.stats.ConsecutiveFailures = 0
		now := timestamp
		cb.stats.LastSuccessTime = &now
	} else {
		cb.stats.FailureCount++
		cb.stats.ConsecutiveFailures++
		cb.stats.ConsecutiveSuccesses = 0
		now := timestamp
		cb.stats.LastFailureTime = &now
	}

	// 计算失败率
	if len(cb.requests) > 0 {
		failures := 0
		for _, req := range cb.requests {
			if !req.success {
				failures++
			}
		}
		cb.stats.FailureRate = float64(failures) / float64(len(cb.requests))
	}
}

// onFailure 失败处理
func (cb *CircuitBreaker) onFailure(timestamp time.Time) {
	if cb.shouldTrip() {
		cb.setState(StateOpen, timestamp)
	}
}

// onSuccess 成功处理
func (cb *CircuitBreaker) onSuccess() {
	// 在关闭状态下的成功不需要特殊处理
}

// shouldTrip 判断是否应该跳闸
func (cb *CircuitBreaker) shouldTrip() bool {
	// 检查最小请求数
	if uint32(len(cb.requests)) < cb.config.MinRequestCount {
		return false
	}

	// 检查连续失败次数
	if cb.stats.ConsecutiveFailures >= cb.config.FailureThreshold {
		return true
	}

	// 检查失败率
	if cb.stats.FailureRate >= cb.config.FailureRate {
		return true
	}

	return false
}

// setState 设置状态
func (cb *CircuitBreaker) setState(state CircuitState, timestamp time.Time) {
	if cb.state == state {
		return
	}

	oldState := cb.state
	cb.state = state
	cb.generation++
	cb.stats.State = state
	cb.stats.StateChangeTime = timestamp

	// 设置状态过期时间
	switch state {
	case StateOpen:
		cb.stateExpiry = timestamp.Add(cb.config.Timeout)
		next := cb.stateExpiry
		cb.stats.NextRetryTime = &next

	case StateHalfOpen:
		cb.halfOpenCount = 0
		cb.stats.NextRetryTime = nil

	case StateClosed:
		cb.stateExpiry = time.Time{}
		cb.stats.NextRetryTime = nil
	}

	// 记录状态变化
	cb.logger.Info("Circuit breaker state changed",
		"circuit", cb.config.Name,
		"from", oldState,
		"to", state,
		"failure_rate", cb.stats.FailureRate,
		"consecutive_failures", cb.stats.ConsecutiveFailures,
	)

	// 调用状态变化回调
	if cb.config.OnStateChange != nil {
		go cb.config.OnStateChange(cb.config.Name, oldState, state)
	}
}

// GetStats 获取统计信息
func (cb *CircuitBreaker) GetStats() CircuitBreakerStats {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	// 创建副本
	stats := cb.stats
	return stats
}

// GetName 获取熔断器名称
func (cb *CircuitBreaker) GetName() string {
	return cb.config.Name
}

// GetState 获取当前状态
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.state
}

// Reset 重置熔断器
func (cb *CircuitBreaker) Reset() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.setState(StateClosed, time.Now())
	cb.requests = make([]requestRecord, 0)
	cb.stats = CircuitBreakerStats{
		State:           StateClosed,
		StateChangeTime: time.Now(),
	}

	cb.logger.Info("Circuit breaker reset", "circuit", cb.config.Name)
}

// 全局熔断器管理器
type CircuitBreakerManager struct {
	breakers map[string]*CircuitBreaker
	logger   logging.Logger
	mutex    sync.RWMutex
}

var globalManager *CircuitBreakerManager
var managerOnce sync.Once

// GetGlobalManager 获取全局管理器
func GetGlobalManager() *CircuitBreakerManager {
	managerOnce.Do(func() {
		globalManager = &CircuitBreakerManager{
			breakers: make(map[string]*CircuitBreaker),
			logger:   logging.GetDefaultLogger(),
		}
	})
	return globalManager
}

// GetOrCreate 获取或创建熔断器
func (cbm *CircuitBreakerManager) GetOrCreate(name string, config CircuitBreakerConfig) *CircuitBreaker {
	cbm.mutex.Lock()
	defer cbm.mutex.Unlock()

	if cb, exists := cbm.breakers[name]; exists {
		return cb
	}

	cb := NewCircuitBreaker(config, cbm.logger)
	cbm.breakers[name] = cb
	return cb
}

// GetBreaker 获取熔断器
func (cbm *CircuitBreakerManager) GetBreaker(name string) *CircuitBreaker {
	cbm.mutex.RLock()
	defer cbm.mutex.RUnlock()

	return cbm.breakers[name]
}

// GetAllStats 获取所有熔断器统计
func (cbm *CircuitBreakerManager) GetAllStats() map[string]CircuitBreakerStats {
	cbm.mutex.RLock()
	defer cbm.mutex.RUnlock()

	stats := make(map[string]CircuitBreakerStats)
	for name, cb := range cbm.breakers {
		stats[name] = cb.GetStats()
	}

	return stats
}

// 便捷函数

// ExecuteWithCircuitBreaker 使用熔断器执行操作
func ExecuteWithCircuitBreaker(
	ctx context.Context,
	name string,
	config CircuitBreakerConfig,
	operation func() error,
) error {
	manager := GetGlobalManager()
	cb := manager.GetOrCreate(name, config)
	return cb.Execute(ctx, operation)
}

// ExecuteDatabase 执行数据库操作
func ExecuteDatabase(ctx context.Context, operation func() error) error {
	config := DatabaseCircuitBreakerConfig()
	return ExecuteWithCircuitBreaker(ctx, "database", config, operation)
}

// ExecuteExternalService 执行外部服务调用
func ExecuteExternalService(ctx context.Context, serviceName string, operation func() error) error {
	config := ExternalServiceCircuitBreakerConfig(serviceName)
	return ExecuteWithCircuitBreaker(ctx, serviceName, config, operation)
}
