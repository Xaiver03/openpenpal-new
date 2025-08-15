package resilience

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestCircuitBreakerStates tests all circuit breaker state transitions
func TestCircuitBreakerStates(t *testing.T) {
	config := CircuitBreakerConfig{
		Name:        "test-breaker",
		MaxRequests: 3,
		Interval:    100 * time.Millisecond,
		Timeout:     500 * time.Millisecond,
		ReadyToTrip: func(counts Counts) bool {
			return counts.Requests >= 3 && counts.FailureRate() >= 60
		},
	}

	cb := NewCircuitBreaker(config)

	// Test 1: Initial state should be closed
	if cb.GetState() != StateClosed {
		t.Errorf("Initial state should be closed, got %v", cb.GetState())
	}

	// Test 2: Execute successful requests
	for i := 0; i < 2; i++ {
		result, err := cb.Execute(func() (interface{}, error) {
			return "success", nil
		})
		if err != nil {
			t.Errorf("Expected successful execution, got error: %v", err)
		}
		if result != "success" {
			t.Errorf("Expected 'success', got %v", result)
		}
	}

	// State should still be closed
	if cb.GetState() != StateClosed {
		t.Errorf("State should remain closed after successful requests, got %v", cb.GetState())
	}

	// Test 3: Execute failing requests to trip the breaker
	for i := 0; i < 3; i++ {
		_, err := cb.Execute(func() (interface{}, error) {
			return nil, errors.New("test error")
		})
		if err == nil || err.Error() != "test error" {
			t.Errorf("Expected 'test error', got %v", err)
		}
	}

	// Circuit should now be open
	if cb.GetState() != StateOpen {
		t.Errorf("Circuit should be open after failures, got %v", cb.GetState())
	}

	// Test 4: Requests should fail immediately when open
	_, err := cb.Execute(func() (interface{}, error) {
		return "should not execute", nil
	})
	if err != ErrCircuitBreakerOpen {
		t.Errorf("Expected ErrCircuitBreakerOpen, got %v", err)
	}

	// Test 5: Wait for timeout to transition to half-open
	time.Sleep(600 * time.Millisecond)
	
	if cb.GetState() != StateHalfOpen {
		t.Errorf("Circuit should be half-open after timeout, got %v", cb.GetState())
	}

	// Test 6: Successful request in half-open should close the circuit
	result, err := cb.Execute(func() (interface{}, error) {
		return "recovery", nil
	})
	if err != nil {
		t.Errorf("Expected successful execution in half-open, got error: %v", err)
	}
	if result != "recovery" {
		t.Errorf("Expected 'recovery', got %v", result)
	}

	// More successful requests to fully close
	for i := 0; i < 2; i++ {
		cb.Execute(func() (interface{}, error) {
			return "success", nil
		})
	}

	if cb.GetState() != StateClosed {
		t.Errorf("Circuit should be closed after successful recovery, got %v", cb.GetState())
	}
}

// TestCircuitBreakerConcurrency tests concurrent access to circuit breaker
func TestCircuitBreakerConcurrency(t *testing.T) {
	config := CircuitBreakerConfig{
		Name:        "concurrent-test",
		MaxRequests: 5,
		Interval:    100 * time.Millisecond,
		Timeout:     200 * time.Millisecond,
		ReadyToTrip: func(counts Counts) bool {
			return counts.ConsecutiveFailures >= 5
		},
	}

	cb := NewCircuitBreaker(config)
	
	var wg sync.WaitGroup
	var successCount int32
	var failureCount int32
	var openCount int32

	// Run 100 concurrent requests
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			
			// Mix of success and failure
			shouldFail := index%3 == 0
			
			_, err := cb.Execute(func() (interface{}, error) {
				if shouldFail {
					return nil, errors.New("concurrent error")
				}
				return "success", nil
			})
			
			if err == nil {
				atomic.AddInt32(&successCount, 1)
			} else if err == ErrCircuitBreakerOpen {
				atomic.AddInt32(&openCount, 1)
			} else {
				atomic.AddInt32(&failureCount, 1)
			}
		}(i)
	}

	wg.Wait()

	t.Logf("Results: Success=%d, Failure=%d, Open=%d", 
		successCount, failureCount, openCount)

	// Verify counts are reasonable
	total := successCount + failureCount + openCount
	if total != 100 {
		t.Errorf("Total count mismatch: expected 100, got %d", total)
	}
}

// TestCircuitBreakerTimeout tests timeout functionality
func TestCircuitBreakerTimeout(t *testing.T) {
	config := CircuitBreakerConfig{
		Name:        "timeout-test",
		MaxRequests: 3,
		Interval:    100 * time.Millisecond,
		Timeout:     200 * time.Millisecond,
	}

	cb := NewCircuitBreaker(config)
	ctx := context.Background()

	// Test timeout
	start := time.Now()
	_, err := cb.ExecuteWithTimeout(ctx, 100*time.Millisecond, func() (interface{}, error) {
		time.Sleep(200 * time.Millisecond)
		return "should timeout", nil
	})

	duration := time.Since(start)
	if err != ErrCircuitBreakerTimeout {
		t.Errorf("Expected timeout error, got %v", err)
	}

	// Duration should be around 100ms
	if duration < 90*time.Millisecond || duration > 150*time.Millisecond {
		t.Errorf("Timeout duration unexpected: %v", duration)
	}
}

// TestCircuitBreakerCustomReadyToTrip tests custom trip conditions
func TestCircuitBreakerCustomReadyToTrip(t *testing.T) {
	tripCount := 0
	config := CircuitBreakerConfig{
		Name:        "custom-trip-test",
		MaxRequests: 5,
		ReadyToTrip: func(counts Counts) bool {
			tripCount++
			// Trip after 2 consecutive failures
			return counts.ConsecutiveFailures >= 2
		},
	}

	cb := NewCircuitBreaker(config)

	// First failure
	cb.Execute(func() (interface{}, error) {
		return nil, errors.New("error 1")
	})

	if cb.GetState() != StateClosed {
		t.Errorf("Should still be closed after 1 failure")
	}

	// Second consecutive failure should trip
	cb.Execute(func() (interface{}, error) {
		return nil, errors.New("error 2")
	})

	if cb.GetState() != StateOpen {
		t.Errorf("Should be open after 2 consecutive failures")
	}

	// Verify ReadyToTrip was called
	if tripCount < 2 {
		t.Errorf("ReadyToTrip should have been called at least twice, was called %d times", tripCount)
	}
}

// TestCircuitBreakerReset tests manual reset functionality
func TestCircuitBreakerReset(t *testing.T) {
	config := CircuitBreakerConfig{
		Name:        "reset-test",
		MaxRequests: 1,
		ReadyToTrip: func(counts Counts) bool {
			return counts.Requests >= 1 && counts.FailureRate() >= 100
		},
	}

	cb := NewCircuitBreaker(config)

	// Cause failure to open circuit
	cb.Execute(func() (interface{}, error) {
		return nil, errors.New("fail")
	})

	if cb.GetState() != StateOpen {
		t.Errorf("Circuit should be open")
	}

	// Reset the circuit
	cb.Reset()

	if cb.GetState() != StateClosed {
		t.Errorf("Circuit should be closed after reset")
	}

	// Verify counts are reset
	counts := cb.GetCounts()
	if counts.Requests != 0 || counts.TotalFailures != 0 {
		t.Errorf("Counts should be reset: %+v", counts)
	}
}

// TestCircuitBreakerStateChangeCallback tests state change notifications
func TestCircuitBreakerStateChangeCallback(t *testing.T) {
	var stateChanges []string
	var mu sync.Mutex

	config := CircuitBreakerConfig{
		Name:        "callback-test",
		MaxRequests: 1,
		Timeout:     100 * time.Millisecond,
		ReadyToTrip: func(counts Counts) bool {
			return counts.Requests >= 1 && counts.FailureRate() >= 100
		},
		OnStateChange: func(name string, from, to CircuitBreakerState) {
			mu.Lock()
			stateChanges = append(stateChanges, fmt.Sprintf("%s->%s", from, to))
			mu.Unlock()
		},
	}

	cb := NewCircuitBreaker(config)

	// Cause state change: Closed -> Open
	cb.Execute(func() (interface{}, error) {
		return nil, errors.New("fail")
	})

	// Wait for timeout: Open -> HalfOpen
	time.Sleep(150 * time.Millisecond)
	cb.GetState() // Trigger state check

	// Recover: HalfOpen -> Closed
	cb.Execute(func() (interface{}, error) {
		return "success", nil
	})

	mu.Lock()
	defer mu.Unlock()

	if len(stateChanges) < 2 {
		t.Errorf("Expected at least 2 state changes, got %d: %v", len(stateChanges), stateChanges)
	}

	// Verify first state change was Closed -> Open
	if len(stateChanges) > 0 && stateChanges[0] != "CLOSED->OPEN" {
		t.Errorf("First state change should be CLOSED->OPEN, got %s", stateChanges[0])
	}
}

// TestCircuitBreakerManager tests the circuit breaker manager
func TestCircuitBreakerManager(t *testing.T) {
	manager := NewCircuitBreakerManager()

	// Test getting circuit breaker
	config1 := CircuitBreakerConfig{
		MaxRequests: 5,
	}
	cb1 := manager.GetCircuitBreaker("service1", config1)
	if cb1 == nil {
		t.Error("Should create circuit breaker")
	}

	// Test getting same circuit breaker
	cb2 := manager.GetCircuitBreaker("service1", config1)
	if cb1 != cb2 {
		t.Error("Should return same circuit breaker instance")
	}

	// Test getting different circuit breaker
	cb3 := manager.GetCircuitBreaker("service2", config1)
	if cb3 == cb1 {
		t.Error("Should return different circuit breaker for different service")
	}

	// Test GetAllBreakers
	breakers := manager.GetAllBreakers()
	if len(breakers) != 2 {
		t.Errorf("Expected 2 breakers, got %d", len(breakers))
	}

	// Test GetStats
	stats := manager.GetStats()
	if len(stats) != 2 {
		t.Errorf("Expected stats for 2 breakers, got %d", len(stats))
	}

	// Verify stats structure
	for name, stat := range stats {
		statMap, ok := stat.(map[string]interface{})
		if !ok {
			t.Errorf("Stats for %s should be a map", name)
			continue
		}
		
		// Check required fields
		requiredFields := []string{"state", "total_requests", "success_rate", "failure_rate"}
		for _, field := range requiredFields {
			if _, exists := statMap[field]; !exists {
				t.Errorf("Stats for %s missing field: %s", name, field)
			}
		}
	}
}

// TestCounts tests the Counts methods
func TestCounts(t *testing.T) {
	tests := []struct {
		name          string
		counts        Counts
		expectedSR    float64
		expectedFR    float64
	}{
		{
			name:       "No requests",
			counts:     Counts{Requests: 0},
			expectedSR: 0,
			expectedFR: 0,
		},
		{
			name: "All success",
			counts: Counts{
				Requests:       10,
				TotalSuccesses: 10,
				TotalFailures:  0,
			},
			expectedSR: 100,
			expectedFR: 0,
		},
		{
			name: "All failures",
			counts: Counts{
				Requests:       10,
				TotalSuccesses: 0,
				TotalFailures:  10,
			},
			expectedSR: 0,
			expectedFR: 100,
		},
		{
			name: "Mixed",
			counts: Counts{
				Requests:       10,
				TotalSuccesses: 7,
				TotalFailures:  3,
			},
			expectedSR: 70,
			expectedFR: 30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sr := tt.counts.SuccessRate()
			if sr != tt.expectedSR {
				t.Errorf("SuccessRate() = %v, want %v", sr, tt.expectedSR)
			}

			fr := tt.counts.FailureRate()
			if fr != tt.expectedFR {
				t.Errorf("FailureRate() = %v, want %v", fr, tt.expectedFR)
			}
		})
	}
}

// BenchmarkCircuitBreaker benchmarks circuit breaker performance
func BenchmarkCircuitBreaker(b *testing.B) {
	config := CircuitBreakerConfig{
		Name:        "bench-test",
		MaxRequests: 10,
	}
	cb := NewCircuitBreaker(config)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cb.Execute(func() (interface{}, error) {
				return "success", nil
			})
		}
	})
}

// TestEdgeCases tests various edge cases
func TestEdgeCases(t *testing.T) {
	t.Run("Panic recovery", func(t *testing.T) {
		config := CircuitBreakerConfig{Name: "panic-test"}
		cb := NewCircuitBreaker(config)

		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic to propagate")
			}
		}()

		cb.Execute(func() (interface{}, error) {
			panic("test panic")
		})
	})

	t.Run("Nil function", func(t *testing.T) {
		config := CircuitBreakerConfig{Name: "nil-test"}
		cb := NewCircuitBreaker(config)

		// This should handle gracefully
		result, err := cb.Execute(nil)
		if result != nil || err == nil {
			t.Error("Expected error for nil function")
		}
	})

	t.Run("Custom IsSuccessful", func(t *testing.T) {
		customErr := errors.New("acceptable error")
		config := CircuitBreakerConfig{
			Name: "custom-success-test",
			IsSuccessful: func(err error) bool {
				// Consider our custom error as success
				return err == nil || err == customErr
			},
		}
		cb := NewCircuitBreaker(config)

		// This should be considered successful
		_, err := cb.Execute(func() (interface{}, error) {
			return nil, customErr
		})

		if err != customErr {
			t.Errorf("Expected customErr, got %v", err)
		}

		counts := cb.GetCounts()
		if counts.TotalSuccesses != 1 || counts.TotalFailures != 0 {
			t.Errorf("Custom error should be counted as success: %+v", counts)
		}
	})
}