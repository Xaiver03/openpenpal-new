package middleware

import (
	"context"
	"fmt"
	"net/http"
	"openpenpal-backend/internal/resilience"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// CircuitBreakerMiddleware provides circuit breaker protection for HTTP handlers
type CircuitBreakerMiddleware struct {
	manager *resilience.CircuitBreakerManager
}

// NewCircuitBreakerMiddleware creates a new circuit breaker middleware
func NewCircuitBreakerMiddleware() *CircuitBreakerMiddleware {
	return &CircuitBreakerMiddleware{
		manager: resilience.NewCircuitBreakerManager(),
	}
}

// ServiceCircuitBreaker creates middleware for protecting service calls
func (cbm *CircuitBreakerMiddleware) ServiceCircuitBreaker(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create circuit breaker configuration based on service type
		config := cbm.getServiceConfig(serviceName)
		
		cb := cbm.manager.GetCircuitBreaker(serviceName, config)
		
		// Check if circuit breaker allows the request
		state := cb.GetState()
		if state == resilience.StateOpen {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"success": false,
				"code":    5003,
				"message": fmt.Sprintf("Service %s is temporarily unavailable", serviceName),
				"error":   "Circuit breaker is open",
				"retry_after": 60, // seconds
			})
			c.Abort()
			return
		}
		
		// Store circuit breaker in context for handler use
		c.Set("circuit_breaker", cb)
		c.Set("service_name", serviceName)
		
		c.Next()
	}
}

// DatabaseCircuitBreaker creates middleware for protecting database operations
func (cbm *CircuitBreakerMiddleware) DatabaseCircuitBreaker() gin.HandlerFunc {
	config := resilience.CircuitBreakerConfig{
		MaxRequests: 5,
		Interval:    30 * time.Second,
		Timeout:     120 * time.Second,
		ReadyToTrip: func(counts resilience.Counts) bool {
			// Trip if we have >= 10 requests and failure rate >= 70%
			return counts.Requests >= 10 && counts.FailureRate() >= 70
		},
		OnStateChange: func(name string, from, to resilience.CircuitBreakerState) {
			logStateChange("database", from, to)
		},
		IsSuccessful: func(err error) bool {
			// Consider database timeouts and connection errors as failures
			if err == nil {
				return true
			}
			errStr := err.Error()
			return !strings.Contains(errStr, "timeout") && 
				   !strings.Contains(errStr, "connection") &&
				   !strings.Contains(errStr, "database")
		},
	}

	cb := cbm.manager.GetCircuitBreaker("database", config)

	return func(c *gin.Context) {
		state := cb.GetState()
		if state == resilience.StateOpen {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"success": false,
				"code":    5004,
				"message": "Database service is temporarily unavailable",
				"error":   "Database circuit breaker is open",
				"retry_after": 120,
			})
			c.Abort()
			return
		}

		c.Set("db_circuit_breaker", cb)
		c.Next()
	}
}

// ExternalAPICircuitBreaker creates middleware for protecting external API calls
func (cbm *CircuitBreakerMiddleware) ExternalAPICircuitBreaker(apiName string) gin.HandlerFunc {
	config := resilience.CircuitBreakerConfig{
		MaxRequests: 3,
		Interval:    60 * time.Second,
		Timeout:     300 * time.Second, // 5 minutes
		ReadyToTrip: func(counts resilience.Counts) bool {
			// More sensitive for external APIs
			return counts.Requests >= 5 && counts.FailureRate() >= 60
		},
		OnStateChange: func(name string, from, to resilience.CircuitBreakerState) {
			logStateChange(fmt.Sprintf("external_api_%s", apiName), from, to)
		},
	}

	cb := cbm.manager.GetCircuitBreaker(fmt.Sprintf("external_api_%s", apiName), config)

	return func(c *gin.Context) {
		state := cb.GetState()
		if state == resilience.StateOpen {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"success": false,
				"code":    5005,
				"message": fmt.Sprintf("External API %s is temporarily unavailable", apiName),
				"error":   "External API circuit breaker is open",
				"retry_after": 300,
			})
			c.Abort()
			return
		}

		c.Set("external_api_circuit_breaker", cb)
		c.Set("external_api_name", apiName)
		c.Next()
	}
}

// AIServiceCircuitBreaker creates specific protection for AI service calls
func (cbm *CircuitBreakerMiddleware) AIServiceCircuitBreaker() gin.HandlerFunc {
	config := resilience.CircuitBreakerConfig{
		MaxRequests: 2, // Conservative for AI services
		Interval:    45 * time.Second,
		Timeout:     180 * time.Second,
		ReadyToTrip: func(counts resilience.Counts) bool {
			// Trip faster for AI services due to potential cost
			return counts.Requests >= 3 && counts.FailureRate() >= 50
		},
		OnStateChange: func(name string, from, to resilience.CircuitBreakerState) {
			logStateChange("ai_service", from, to)
		},
		IsSuccessful: func(err error) bool {
			if err == nil {
				return true
			}
			// Consider rate limit errors as temporary failures
			errStr := strings.ToLower(err.Error())
			return !strings.Contains(errStr, "rate limit") &&
				   !strings.Contains(errStr, "quota") &&
				   !strings.Contains(errStr, "timeout")
		},
	}

	cb := cbm.manager.GetCircuitBreaker("ai_service", config)

	return func(c *gin.Context) {
		state := cb.GetState()
		if state == resilience.StateOpen {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"success": false,
				"code":    5006,
				"message": "AI service is temporarily unavailable",
				"error":   "AI service circuit breaker is open",
				"suggestion": "Please try again later or use non-AI features",
				"retry_after": 180,
			})
			c.Abort()
			return
		}

		c.Set("ai_circuit_breaker", cb)
		c.Next()
	}
}

// GetCircuitBreakerStats returns statistics for all circuit breakers
func (cbm *CircuitBreakerMiddleware) GetCircuitBreakerStats() gin.HandlerFunc {
	return func(c *gin.Context) {
		stats := cbm.manager.GetStats()
		
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"circuit_breakers": stats,
				"total_breakers":   len(stats),
				"timestamp":        time.Now(),
			},
		})
	}
}

// HealthCheck endpoint for circuit breaker status
func (cbm *CircuitBreakerMiddleware) HealthCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		breakers := cbm.manager.GetAllBreakers()
		
		openBreakers := make([]string, 0)
		halfOpenBreakers := make([]string, 0)
		
		for name, cb := range breakers {
			state := cb.GetState()
			switch state {
			case resilience.StateOpen:
				openBreakers = append(openBreakers, name)
			case resilience.StateHalfOpen:
				halfOpenBreakers = append(halfOpenBreakers, name)
			}
		}
		
		healthy := len(openBreakers) == 0
		status := http.StatusOK
		if !healthy {
			status = http.StatusServiceUnavailable
		}
		
		c.JSON(status, gin.H{
			"healthy":            healthy,
			"total_breakers":     len(breakers),
			"open_breakers":      openBreakers,
			"half_open_breakers": halfOpenBreakers,
			"timestamp":          time.Now(),
		})
	}
}

// Helper functions

// getServiceConfig returns configuration based on service type
func (cbm *CircuitBreakerMiddleware) getServiceConfig(serviceName string) resilience.CircuitBreakerConfig {
	baseConfig := resilience.CircuitBreakerConfig{
		MaxRequests: 5,
		Interval:    60 * time.Second,
		Timeout:     120 * time.Second,
		OnStateChange: func(name string, from, to resilience.CircuitBreakerState) {
			logStateChange(name, from, to)
		},
	}

	// Customize based on service type
	switch {
	case strings.Contains(serviceName, "courier"):
		baseConfig.ReadyToTrip = func(counts resilience.Counts) bool {
			return counts.Requests >= 8 && counts.FailureRate() >= 65
		}
	case strings.Contains(serviceName, "museum"):
		baseConfig.ReadyToTrip = func(counts resilience.Counts) bool {
			return counts.Requests >= 10 && counts.FailureRate() >= 70
		}
	case strings.Contains(serviceName, "notification"):
		baseConfig.MaxRequests = 3
		baseConfig.ReadyToTrip = func(counts resilience.Counts) bool {
			return counts.Requests >= 5 && counts.FailureRate() >= 50
		}
	default:
		baseConfig.ReadyToTrip = func(counts resilience.Counts) bool {
			return counts.Requests >= 8 && counts.FailureRate() >= 60
		}
	}

	return baseConfig
}

// logStateChange logs circuit breaker state changes
func logStateChange(serviceName string, from, to resilience.CircuitBreakerState) {
	// In production, this should integrate with your logging system
	fmt.Printf("[CircuitBreaker] %s: %s -> %s at %s\n", 
		serviceName, from.String(), to.String(), time.Now().Format(time.RFC3339))
}

// Helper functions for handlers to use circuit breakers

// ExecuteWithCircuitBreaker executes a function with circuit breaker protection
func ExecuteWithCircuitBreaker(c *gin.Context, fn func() (interface{}, error)) (interface{}, error) {
	if cb, exists := c.Get("circuit_breaker"); exists {
		if circuitBreaker, ok := cb.(*resilience.CircuitBreaker); ok {
			return circuitBreaker.Execute(fn)
		}
	}
	// Fallback to direct execution if no circuit breaker
	return fn()
}

// ExecuteWithTimeout executes a function with timeout and circuit breaker protection
func ExecuteWithTimeout(c *gin.Context, timeout time.Duration, fn func() (interface{}, error)) (interface{}, error) {
	if cb, exists := c.Get("circuit_breaker"); exists {
		if circuitBreaker, ok := cb.(*resilience.CircuitBreaker); ok {
			return circuitBreaker.ExecuteWithTimeout(c.Request.Context(), timeout, fn)
		}
	}
	// Fallback with context timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
	defer cancel()
	
	type result struct {
		data interface{}
		err  error
	}
	
	resultChan := make(chan result, 1)
	go func() {
		data, err := fn()
		resultChan <- result{data, err}
	}()
	
	select {
	case res := <-resultChan:
		return res.data, res.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Default instance for easy use
var DefaultCircuitBreakerMiddleware = NewCircuitBreakerMiddleware()