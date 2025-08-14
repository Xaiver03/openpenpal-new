package ratelimit

import (
	"sync"
	"time"
)

// AdaptiveLimiter provides dynamic rate limiting based on behavior patterns
type AdaptiveLimiter struct {
	mu              sync.RWMutex
	baseRate        float64
	baseBurst       int
	suspiciousRate  float64
	suspiciousBurst int
	strictRate      float64
	strictBurst     int
	
	// Track suspicious activity
	failureTracking map[string]*FailureRecord
	blockList       map[string]time.Time
	
	// Cleanup settings
	cleanupInterval time.Duration
	blockDuration   time.Duration
}

// FailureRecord tracks consecutive failures for adaptive limiting
type FailureRecord struct {
	Count        int
	LastFailure  time.Time
	FirstFailure time.Time
}

// SecurityLevel represents the current security posture
type SecurityLevel int

const (
	LevelNormal SecurityLevel = iota
	LevelSuspicious
	LevelStrict
	LevelBlocked
)

// NewAdaptiveLimiter creates a new adaptive rate limiter
func NewAdaptiveLimiter(baseRate float64, baseBurst int) *AdaptiveLimiter {
	al := &AdaptiveLimiter{
		baseRate:        baseRate,
		baseBurst:       baseBurst,
		suspiciousRate:  baseRate / 2,    // Half the rate when suspicious
		suspiciousBurst: baseBurst / 2,
		strictRate:      baseRate / 10,   // 10% of rate when under attack
		strictBurst:     1,
		failureTracking: make(map[string]*FailureRecord),
		blockList:       make(map[string]time.Time),
		cleanupInterval: 5 * time.Minute,
		blockDuration:   30 * time.Minute,
	}
	
	// Start cleanup routine
	go al.cleanupRoutine()
	
	return al
}

// RecordSuccess marks a successful request
func (al *AdaptiveLimiter) RecordSuccess(identifier string) {
	al.mu.Lock()
	defer al.mu.Unlock()
	
	// Reset failure count on success
	delete(al.failureTracking, identifier)
}

// RecordFailure marks a failed request (e.g., failed login)
func (al *AdaptiveLimiter) RecordFailure(identifier string) {
	al.mu.Lock()
	defer al.mu.Unlock()
	
	now := time.Now()
	
	record, exists := al.failureTracking[identifier]
	if !exists {
		record = &FailureRecord{
			FirstFailure: now,
		}
		al.failureTracking[identifier] = record
	}
	
	record.Count++
	record.LastFailure = now
	
	// Auto-block after threshold
	if record.Count >= 10 {
		al.blockList[identifier] = now.Add(al.blockDuration)
	}
}

// GetSecurityLevel determines the current security level for an identifier
func (al *AdaptiveLimiter) GetSecurityLevel(identifier string) SecurityLevel {
	al.mu.RLock()
	defer al.mu.RUnlock()
	
	// Check if blocked
	if blockUntil, blocked := al.blockList[identifier]; blocked {
		if time.Now().Before(blockUntil) {
			return LevelBlocked
		}
	}
	
	// Check failure record
	record, exists := al.failureTracking[identifier]
	if !exists {
		return LevelNormal
	}
	
	// Determine level based on failure pattern
	timeSinceFirst := time.Since(record.FirstFailure)
	
	// Rapid failures (more than 5 in 1 minute)
	if record.Count > 5 && timeSinceFirst < time.Minute {
		return LevelStrict
	}
	
	// Moderate failures (more than 3 in 5 minutes)
	if record.Count > 3 && timeSinceFirst < 5*time.Minute {
		return LevelSuspicious
	}
	
	return LevelNormal
}

// GetRateLimits returns the rate and burst limits for the given security level
func (al *AdaptiveLimiter) GetRateLimits(level SecurityLevel) (rate float64, burst int) {
	switch level {
	case LevelBlocked:
		return 0, 0
	case LevelStrict:
		return al.strictRate, al.strictBurst
	case LevelSuspicious:
		return al.suspiciousRate, al.suspiciousBurst
	default:
		return al.baseRate, al.baseBurst
	}
}

// IsBlocked checks if an identifier is currently blocked
func (al *AdaptiveLimiter) IsBlocked(identifier string) bool {
	al.mu.RLock()
	defer al.mu.RUnlock()
	
	if blockUntil, blocked := al.blockList[identifier]; blocked {
		return time.Now().Before(blockUntil)
	}
	return false
}

// Unblock manually removes an identifier from the block list
func (al *AdaptiveLimiter) Unblock(identifier string) {
	al.mu.Lock()
	defer al.mu.Unlock()
	
	delete(al.blockList, identifier)
	delete(al.failureTracking, identifier)
}

// cleanupRoutine periodically cleans up old records
func (al *AdaptiveLimiter) cleanupRoutine() {
	ticker := time.NewTicker(al.cleanupInterval)
	defer ticker.Stop()
	
	for range ticker.C {
		al.cleanup()
	}
}

// cleanup removes expired blocks and old failure records
func (al *AdaptiveLimiter) cleanup() {
	al.mu.Lock()
	defer al.mu.Unlock()
	
	now := time.Now()
	
	// Clean up expired blocks
	for identifier, blockUntil := range al.blockList {
		if now.After(blockUntil) {
			delete(al.blockList, identifier)
		}
	}
	
	// Clean up old failure records (older than 1 hour)
	for identifier, record := range al.failureTracking {
		if time.Since(record.LastFailure) > time.Hour {
			delete(al.failureTracking, identifier)
		}
	}
}

// GetBlockedIdentifiers returns currently blocked identifiers for monitoring
func (al *AdaptiveLimiter) GetBlockedIdentifiers() []string {
	al.mu.RLock()
	defer al.mu.RUnlock()
	
	blocked := make([]string, 0, len(al.blockList))
	now := time.Now()
	
	for identifier, blockUntil := range al.blockList {
		if now.Before(blockUntil) {
			blocked = append(blocked, identifier)
		}
	}
	
	return blocked
}

// GetMetrics returns current metrics for monitoring
func (al *AdaptiveLimiter) GetMetrics() map[string]interface{} {
	al.mu.RLock()
	defer al.mu.RUnlock()
	
	return map[string]interface{}{
		"blocked_count":    len(al.blockList),
		"tracking_count":   len(al.failureTracking),
		"base_rate":        al.baseRate,
		"suspicious_rate":  al.suspiciousRate,
		"strict_rate":      al.strictRate,
	}
}