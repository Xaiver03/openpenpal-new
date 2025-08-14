package middleware

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"openpenpal-backend/internal/security/ratelimit"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// EnhancedRateLimiter provides adaptive rate limiting based on security levels
type EnhancedRateLimiter struct {
	limiters      map[string]*rateLimiterInfo
	mu            sync.RWMutex
	adaptive      *ratelimit.AdaptiveLimiter
	cleanupTicker *time.Ticker
}

type rateLimiterInfo struct {
	limiter      *rate.Limiter
	lastAccess   time.Time
	failureCount int
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	// General endpoints
	GeneralRate  float64
	GeneralBurst int
	
	// Authentication endpoints
	AuthRate  float64
	AuthBurst int
	
	// API endpoints
	APIRate  float64
	APIBurst int
	
	// Upload endpoints
	UploadRate  float64
	UploadBurst int
}

// DefaultRateLimitConfig returns secure default rate limits
func DefaultRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		GeneralRate:  2,    // 2 requests per second
		GeneralBurst: 10,   // burst of 10
		AuthRate:     0.5,  // 1 request per 2 seconds
		AuthBurst:    3,    // burst of 3
		APIRate:      5,    // 5 requests per second
		APIBurst:     20,   // burst of 20
		UploadRate:   0.2,  // 1 request per 5 seconds
		UploadBurst:  2,    // burst of 2
	}
}

// NewEnhancedRateLimiter creates a new enhanced rate limiter
func NewEnhancedRateLimiter(config *RateLimitConfig) *EnhancedRateLimiter {
	if config == nil {
		config = DefaultRateLimitConfig()
	}

	limiter := &EnhancedRateLimiter{
		limiters:      make(map[string]*rateLimiterInfo),
		adaptive:      ratelimit.NewAdaptiveLimiter(config.GeneralRate, config.GeneralBurst),
		cleanupTicker: time.NewTicker(5 * time.Minute),
	}

	// Start cleanup routine
	go limiter.cleanupRoutine()

	return limiter
}

// GeneralLimiter returns middleware for general endpoints
func (r *EnhancedRateLimiter) GeneralLimiter() gin.HandlerFunc {
	return r.createLimiter("general", 2, 10)
}

// AuthLimiter returns middleware for authentication endpoints
func (r *EnhancedRateLimiter) AuthLimiter() gin.HandlerFunc {
	return r.createLimiter("auth", 0.5, 3)
}

// APILimiter returns middleware for API endpoints
func (r *EnhancedRateLimiter) APILimiter() gin.HandlerFunc {
	return r.createLimiter("api", 5, 20)
}

// UploadLimiter returns middleware for upload endpoints
func (r *EnhancedRateLimiter) UploadLimiter() gin.HandlerFunc {
	return r.createLimiter("upload", 0.2, 2)
}

// createLimiter creates a rate limiting middleware with specified limits
func (r *EnhancedRateLimiter) createLimiter(limitType string, defaultRate float64, defaultBurst int) gin.HandlerFunc {
	return func(c *gin.Context) {
		identifier := r.getIdentifier(c)
		
		// Check if blocked
		if r.adaptive.IsBlocked(identifier) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error":   "Too many failed attempts. Please try again later.",
				"code":    "RATE_LIMIT_BLOCKED",
			})
			c.Abort()
			return
		}

		// Get security level and appropriate rate limits
		level := r.adaptive.GetSecurityLevel(identifier)
		rateLimit, burst := r.adaptive.GetRateLimits(level)
		
		// Use default limits if adaptive returns zero (normal level)
		if rateLimit == 0 {
			rateLimit = defaultRate
			burst = defaultBurst
		}

		// Get or create limiter for this identifier
		limiter := r.getLimiter(identifier, rateLimit, burst)

		// Check rate limit
		if !limiter.Allow() {
			// Record failure for adaptive limiting
			if limitType == "auth" {
				r.adaptive.RecordFailure(identifier)
			}

			c.Header("X-RateLimit-Limit", fmt.Sprintf("%.0f", rateLimit))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Second).Unix()))
			c.Header("Retry-After", "1")

			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error":   "Rate limit exceeded. Please slow down.",
				"code":    "RATE_LIMIT_EXCEEDED",
				"retry_after": 1,
			})
			c.Abort()
			return
		}

		// Add rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%.0f", rateLimit))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", limiter.Burst()))

		// Store rate limit info in context
		c.Set("rate_limit_identifier", identifier)
		c.Set("rate_limit_type", limitType)

		c.Next()

		// Record success for auth endpoints
		if limitType == "auth" && c.Writer.Status() < 400 {
			r.adaptive.RecordSuccess(identifier)
		}
	}
}

// getIdentifier returns a unique identifier for rate limiting
func (r *EnhancedRateLimiter) getIdentifier(c *gin.Context) string {
	// Try to use user ID if authenticated
	if userID, exists := c.Get("user_id"); exists {
		return fmt.Sprintf("user:%v", userID)
	}

	// Use IP address as fallback
	ip := c.ClientIP()
	
	// Handle X-Forwarded-For header
	if forwardedFor := c.GetHeader("X-Forwarded-For"); forwardedFor != "" {
		// Take the first IP in the chain
		ips := strings.Split(forwardedFor, ",")
		if len(ips) > 0 {
			ip = strings.TrimSpace(ips[0])
		}
	}

	// Handle X-Real-IP header
	if realIP := c.GetHeader("X-Real-IP"); realIP != "" {
		ip = realIP
	}

	return fmt.Sprintf("ip:%s", ip)
}

// getLimiter gets or creates a rate limiter for the identifier
func (r *EnhancedRateLimiter) getLimiter(identifier string, rateLimit float64, burst int) *rate.Limiter {
	r.mu.RLock()
	info, exists := r.limiters[identifier]
	r.mu.RUnlock()

	if exists {
		// Update last access time
		r.mu.Lock()
		info.lastAccess = time.Now()
		r.mu.Unlock()
		
		// Check if we need to adjust the rate
		currentLimit := info.limiter.Limit()
		if currentLimit != rate.Limit(rateLimit) {
			info.limiter.SetLimit(rate.Limit(rateLimit))
			info.limiter.SetBurst(burst)
		}
		
		return info.limiter
	}

	// Create new limiter
	r.mu.Lock()
	limiter := rate.NewLimiter(rate.Limit(rateLimit), burst)
	r.limiters[identifier] = &rateLimiterInfo{
		limiter:    limiter,
		lastAccess: time.Now(),
	}
	r.mu.Unlock()

	return limiter
}

// cleanupRoutine removes inactive limiters
func (r *EnhancedRateLimiter) cleanupRoutine() {
	for range r.cleanupTicker.C {
		r.cleanup()
	}
}

// cleanup removes limiters that haven't been used recently
func (r *EnhancedRateLimiter) cleanup() {
	r.mu.Lock()
	defer r.mu.Unlock()

	cutoff := time.Now().Add(-30 * time.Minute)
	
	for identifier, info := range r.limiters {
		if info.lastAccess.Before(cutoff) {
			delete(r.limiters, identifier)
			log.Printf("Cleaned up rate limiter for: %s", identifier)
		}
	}
}

// RecordAuthFailure records an authentication failure for adaptive limiting
func (r *EnhancedRateLimiter) RecordAuthFailure(identifier string) {
	r.adaptive.RecordFailure(identifier)
}

// RecordAuthSuccess records an authentication success for adaptive limiting
func (r *EnhancedRateLimiter) RecordAuthSuccess(identifier string) {
	r.adaptive.RecordSuccess(identifier)
}

// GetMetrics returns current rate limiting metrics
func (r *EnhancedRateLimiter) GetMetrics() map[string]interface{} {
	r.mu.RLock()
	limiterCount := len(r.limiters)
	r.mu.RUnlock()

	adaptiveMetrics := r.adaptive.GetMetrics()
	blocked := r.adaptive.GetBlockedIdentifiers()

	return map[string]interface{}{
		"active_limiters":    limiterCount,
		"blocked_identifiers": blocked,
		"adaptive_metrics":   adaptiveMetrics,
	}
}

// Stop gracefully stops the rate limiter
func (r *EnhancedRateLimiter) Stop() {
	r.cleanupTicker.Stop()
}