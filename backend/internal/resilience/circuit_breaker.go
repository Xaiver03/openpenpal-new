package resilience

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

// CircuitBreakerState represents the current state of the circuit breaker
type CircuitBreakerState int

const (
	// StateClosed - circuit is closed, requests are allowed
	StateClosed CircuitBreakerState = iota
	// StateOpen - circuit is open, requests are blocked
	StateOpen
	// StateHalfOpen - circuit is in testing mode, limited requests allowed
	StateHalfOpen
)

// String returns string representation of the state
func (s CircuitBreakerState) String() string {
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

var (
	// ErrCircuitBreakerOpen is returned when circuit breaker is open
	ErrCircuitBreakerOpen = errors.New("circuit breaker is open")
	// ErrCircuitBreakerTimeout is returned when request times out
	ErrCircuitBreakerTimeout = errors.New("circuit breaker timeout")
)

// CircuitBreakerConfig contains configuration for circuit breaker
type CircuitBreakerConfig struct {
	Name                string        // Name of the circuit breaker for logging
	MaxRequests         uint32        // Maximum requests allowed in half-open state
	Interval            time.Duration // Interval for clearing counters in closed state
	Timeout             time.Duration // Timeout for moving from open to half-open
	ReadyToTrip         func(counts Counts) bool // Function to determine if circuit should trip
	OnStateChange       func(name string, from, to CircuitBreakerState) // Callback for state changes
	ShouldTrip          func(counts Counts) bool // Custom trip condition
	IsSuccessful        func(err error) bool     // Function to determine if result is successful
}

// Counts holds the numbers of requests and their results
type Counts struct {
	Requests             uint32    // Total requests
	TotalSuccesses       uint32    // Total successful requests
	TotalFailures        uint32    // Total failed requests
	ConsecutiveSuccesses uint32    // Consecutive successful requests
	ConsecutiveFailures  uint32    // Consecutive failed requests
	LastSuccessTime      time.Time // Time of last successful request
	LastFailureTime      time.Time // Time of last failed request
}

// SuccessRate returns the success rate as a percentage
func (c Counts) SuccessRate() float64 {
	if c.Requests == 0 {
		return 0
	}
	return float64(c.TotalSuccesses) / float64(c.Requests) * 100
}

// FailureRate returns the failure rate as a percentage
func (c Counts) FailureRate() float64 {
	if c.Requests == 0 {
		return 0
	}
	return float64(c.TotalFailures) / float64(c.Requests) * 100
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	name         string
	maxRequests  uint32
	interval     time.Duration
	timeout      time.Duration
	readyToTrip  func(counts Counts) bool
	isSuccessful func(err error) bool
	onStateChange func(name string, from, to CircuitBreakerState)

	mutex      sync.Mutex
	state      CircuitBreakerState
	generation uint64
	counts     Counts
	expiry     time.Time
}

// NewCircuitBreaker creates a new circuit breaker with the given configuration
func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker {
	cb := &CircuitBreaker{
		name:        config.Name,
		maxRequests: config.MaxRequests,
		interval:    config.Interval,
		timeout:     config.Timeout,
		readyToTrip: config.ReadyToTrip,
		isSuccessful: config.IsSuccessful,
		onStateChange: config.OnStateChange,
	}

	// Set default values
	if cb.maxRequests == 0 {
		cb.maxRequests = 1
	}
	if cb.interval <= 0 {
		cb.interval = 60 * time.Second
	}
	if cb.timeout <= 0 {
		cb.timeout = 60 * time.Second
	}
	if cb.readyToTrip == nil {
		cb.readyToTrip = defaultReadyToTrip
	}
	if cb.isSuccessful == nil {
		cb.isSuccessful = defaultIsSuccessful
	}

	cb.toNewGeneration(time.Now())
	return cb
}

// Execute runs the given function with circuit breaker protection
func (cb *CircuitBreaker) Execute(fn func() (interface{}, error)) (interface{}, error) {
	generation, err := cb.beforeRequest()
	if err != nil {
		return nil, err
	}

	defer func() {
		if r := recover(); r != nil {
			cb.afterRequest(generation, false)
			panic(r)
		}
	}()

	result, err := fn()
	cb.afterRequest(generation, cb.isSuccessful(err))
	return result, err
}

// ExecuteWithTimeout runs the given function with timeout and circuit breaker protection
func (cb *CircuitBreaker) ExecuteWithTimeout(ctx context.Context, timeout time.Duration, fn func() (interface{}, error)) (interface{}, error) {
	generation, err := cb.beforeRequest()
	if err != nil {
		return nil, err
	}

	// Create timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	type result struct {
		data interface{}
		err  error
	}

	resultChan := make(chan result, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				cb.afterRequest(generation, false)
				resultChan <- result{nil, fmt.Errorf("panic: %v", r)}
			}
		}()

		data, err := fn()
		resultChan <- result{data, err}
	}()

	select {
	case res := <-resultChan:
		cb.afterRequest(generation, cb.isSuccessful(res.err))
		return res.data, res.err
	case <-timeoutCtx.Done():
		cb.afterRequest(generation, false)
		return nil, ErrCircuitBreakerTimeout
	}
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	now := time.Now()
	state, _ := cb.currentState(now)
	return state
}

// GetCounts returns the current counts
func (cb *CircuitBreaker) GetCounts() Counts {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	return cb.counts
}

// GetName returns the name of the circuit breaker
func (cb *CircuitBreaker) GetName() string {
	return cb.name
}

// Reset resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.toNewGeneration(time.Now())
	cb.setState(StateClosed, time.Now())
}

// beforeRequest checks if request is allowed and updates internal state
func (cb *CircuitBreaker) beforeRequest() (uint64, error) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	now := time.Now()
	state, generation := cb.currentState(now)

	if state == StateOpen {
		return generation, ErrCircuitBreakerOpen
	} else if state == StateHalfOpen && cb.counts.Requests >= cb.maxRequests {
		return generation, ErrCircuitBreakerOpen
	}

	cb.counts.Requests++
	return generation, nil
}

// afterRequest updates the circuit breaker state after request completion
func (cb *CircuitBreaker) afterRequest(before uint64, success bool) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	now := time.Now()
	state, generation := cb.currentState(now)
	if generation != before {
		return
	}

	if success {
		cb.onSuccess(state, now)
	} else {
		cb.onFailure(state, now)
	}
}

// onSuccess handles successful request
func (cb *CircuitBreaker) onSuccess(state CircuitBreakerState, now time.Time) {
	cb.counts.TotalSuccesses++
	cb.counts.ConsecutiveSuccesses++
	cb.counts.ConsecutiveFailures = 0
	cb.counts.LastSuccessTime = now

	if state == StateHalfOpen && cb.counts.ConsecutiveSuccesses >= cb.maxRequests {
		cb.setState(StateClosed, now)
	}
}

// onFailure handles failed request
func (cb *CircuitBreaker) onFailure(state CircuitBreakerState, now time.Time) {
	cb.counts.TotalFailures++
	cb.counts.ConsecutiveFailures++
	cb.counts.ConsecutiveSuccesses = 0
	cb.counts.LastFailureTime = now

	if cb.readyToTrip(cb.counts) {
		cb.setState(StateOpen, now)
	}
}

// currentState returns the current state and generation
func (cb *CircuitBreaker) currentState(now time.Time) (CircuitBreakerState, uint64) {
	switch cb.state {
	case StateClosed:
		if !cb.expiry.IsZero() && cb.expiry.Before(now) {
			cb.toNewGeneration(now)
		}
	case StateOpen:
		if cb.expiry.Before(now) {
			cb.setState(StateHalfOpen, now)
		}
	}
	return cb.state, cb.generation
}

// setState changes the state and calls the callback
func (cb *CircuitBreaker) setState(state CircuitBreakerState, now time.Time) {
	if cb.state == state {
		return
	}

	prev := cb.state
	cb.state = state

	cb.toNewGeneration(now)

	if cb.onStateChange != nil {
		cb.onStateChange(cb.name, prev, state)
	}

	log.Printf("[CircuitBreaker:%s] State changed from %s to %s", cb.name, prev, state)
}

// toNewGeneration starts a new generation
func (cb *CircuitBreaker) toNewGeneration(now time.Time) {
	cb.generation++
	cb.counts = Counts{}

	var zero time.Time
	switch cb.state {
	case StateClosed:
		if cb.interval == 0 {
			cb.expiry = zero
		} else {
			cb.expiry = now.Add(cb.interval)
		}
	case StateOpen:
		cb.expiry = now.Add(cb.timeout)
	default: // StateHalfOpen
		cb.expiry = zero
	}
}

// Default implementations

// defaultReadyToTrip is the default function to determine if circuit should trip
func defaultReadyToTrip(counts Counts) bool {
	return counts.Requests >= 5 && counts.FailureRate() >= 60
}

// defaultIsSuccessful is the default function to determine if result is successful
func defaultIsSuccessful(err error) bool {
	return err == nil
}

// CircuitBreakerManager manages multiple circuit breakers
type CircuitBreakerManager struct {
	breakers map[string]*CircuitBreaker
	mutex    sync.RWMutex
}

// NewCircuitBreakerManager creates a new circuit breaker manager
func NewCircuitBreakerManager() *CircuitBreakerManager {
	return &CircuitBreakerManager{
		breakers: make(map[string]*CircuitBreaker),
	}
}

// GetCircuitBreaker returns a circuit breaker by name, creating it if it doesn't exist
func (cbm *CircuitBreakerManager) GetCircuitBreaker(name string, config CircuitBreakerConfig) *CircuitBreaker {
	cbm.mutex.RLock()
	if cb, exists := cbm.breakers[name]; exists {
		cbm.mutex.RUnlock()
		return cb
	}
	cbm.mutex.RUnlock()

	cbm.mutex.Lock()
	defer cbm.mutex.Unlock()

	// Double-check after acquiring write lock
	if cb, exists := cbm.breakers[name]; exists {
		return cb
	}

	config.Name = name
	cb := NewCircuitBreaker(config)
	cbm.breakers[name] = cb
	return cb
}

// GetAllBreakers returns all circuit breakers
func (cbm *CircuitBreakerManager) GetAllBreakers() map[string]*CircuitBreaker {
	cbm.mutex.RLock()
	defer cbm.mutex.RUnlock()

	result := make(map[string]*CircuitBreaker)
	for name, cb := range cbm.breakers {
		result[name] = cb
	}
	return result
}

// GetStats returns statistics for all circuit breakers
func (cbm *CircuitBreakerManager) GetStats() map[string]interface{} {
	cbm.mutex.RLock()
	defer cbm.mutex.RUnlock()

	stats := make(map[string]interface{})
	for name, cb := range cbm.breakers {
		counts := cb.GetCounts()
		stats[name] = map[string]interface{}{
			"state":                  cb.GetState().String(),
			"total_requests":         counts.Requests,
			"total_successes":        counts.TotalSuccesses,
			"total_failures":         counts.TotalFailures,
			"consecutive_successes":  counts.ConsecutiveSuccesses,
			"consecutive_failures":   counts.ConsecutiveFailures,
			"success_rate":          counts.SuccessRate(),
			"failure_rate":          counts.FailureRate(),
			"last_success_time":     counts.LastSuccessTime,
			"last_failure_time":     counts.LastFailureTime,
		}
	}
	return stats
}

// Default circuit breaker manager instance
var DefaultCircuitBreakerManager = NewCircuitBreakerManager()

// Helper functions for common use cases

// ExecuteWithCircuitBreaker executes a function with circuit breaker protection
func ExecuteWithCircuitBreaker(name string, fn func() (interface{}, error)) (interface{}, error) {
	cb := DefaultCircuitBreakerManager.GetCircuitBreaker(name, CircuitBreakerConfig{})
	return cb.Execute(fn)
}

// ExecuteHTTPWithCircuitBreaker executes an HTTP request with circuit breaker protection
func ExecuteHTTPWithCircuitBreaker(serviceName string, fn func() (interface{}, error)) (interface{}, error) {
	config := CircuitBreakerConfig{
		MaxRequests: 3,
		Interval:    30 * time.Second,
		Timeout:     60 * time.Second,
		ReadyToTrip: func(counts Counts) bool {
			return counts.Requests >= 10 && counts.FailureRate() >= 50
		},
		OnStateChange: func(name string, from, to CircuitBreakerState) {
			log.Printf("[CircuitBreaker:%s] HTTP service state changed: %s -> %s", name, from, to)
		},
	}
	
	cb := DefaultCircuitBreakerManager.GetCircuitBreaker("http_"+serviceName, config)
	return cb.Execute(fn)
}