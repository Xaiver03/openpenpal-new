// Package pool provides connection health checking capabilities
package pool

import (
	"context"
	"database/sql"
	"log"
	"sync"
	"time"

	"openpenpal-backend/internal/platform/dbgovernance/core"
)

// HealthChecker monitors connection health and performs health checks
type HealthChecker struct {
	config          *core.ConnectionPoolConfig
	healthHistory   map[string]*ConnectionHealthHistory
	mu              sync.RWMutex
	checkInterval   time.Duration
	failureThreshold int
}

// ConnectionHealthHistory tracks health history for a connection
type ConnectionHealthHistory struct {
	ConnectionID     string
	HealthChecks     []HealthCheckResult
	ConsecutiveFailures int
	LastHealthy      time.Time
	TotalChecks      int
	SuccessfulChecks int
	AverageLatency   time.Duration
}

// HealthCheckResult represents the result of a health check
type HealthCheckResult struct {
	Timestamp time.Time
	Success   bool
	Latency   time.Duration
	Error     string
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(config *core.ConnectionPoolConfig) *HealthChecker {
	return &HealthChecker{
		config:           config,
		healthHistory:    make(map[string]*ConnectionHealthHistory),
		checkInterval:    config.HealthCheckInterval,
		failureThreshold: 3, // Default failure threshold
	}
}

// CheckConnection performs a health check on a single connection
func (hc *HealthChecker) CheckConnection(conn *PooledConnection) bool {
	if conn == nil || conn.DB == nil {
		return false
	}
	
	startTime := time.Now()
	
	// Perform ping test
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	err := conn.DB.PingContext(ctx)
	latency := time.Since(startTime)
	
	success := err == nil
	
	// Record health check result
	result := HealthCheckResult{
		Timestamp: startTime,
		Success:   success,
		Latency:   latency,
	}
	
	if err != nil {
		result.Error = err.Error()
	}
	
	hc.recordHealthCheck(conn.ID, result)
	
	// Update connection health status
	conn.mu.Lock()
	conn.Healthy = success
	if !success {
		conn.FailureCount++
	} else {
		conn.FailureCount = 0
	}
	conn.mu.Unlock()
	
	if !success {
		log.Printf("🏥 Connection %s health check failed: %v (latency: %v)", 
			conn.ID, err, latency)
	}
	
	return success
}

// GetConnectionHealth returns health information for a connection
func (hc *HealthChecker) GetConnectionHealth(conn *PooledConnection) *core.ConnectionHealth {
	hc.mu.RLock()
	history, exists := hc.healthHistory[conn.ID]
	hc.mu.RUnlock()
	
	status := "unknown"
	var latency time.Duration
	var errorCount int
	
	if exists && len(history.HealthChecks) > 0 {
		latestCheck := history.HealthChecks[len(history.HealthChecks)-1]
		latency = latestCheck.Latency
		errorCount = history.TotalChecks - history.SuccessfulChecks
		
		if history.ConsecutiveFailures == 0 {
			status = "healthy"
		} else if history.ConsecutiveFailures < hc.failureThreshold {
			status = "degraded"
		} else {
			status = "unhealthy"
		}
	}
	
	return &core.ConnectionHealth{
		DatabaseName: conn.ID,
		Status:       status,
		Latency:      latency,
		LastChecked:  time.Now(),
		ErrorCount:   errorCount,
		Metadata: map[string]interface{}{
			"consecutive_failures": history.ConsecutiveFailures,
			"success_rate":         hc.calculateSuccessRate(history),
			"average_latency":      history.AverageLatency,
		},
	}
}

// PerformAdvancedHealthCheck performs comprehensive health check
func (hc *HealthChecker) PerformAdvancedHealthCheck(conn *PooledConnection) *AdvancedHealthResult {
	result := &AdvancedHealthResult{
		ConnectionID: conn.ID,
		Timestamp:    time.Now(),
		Tests:        make(map[string]HealthTestResult),
	}
	
	// Test 1: Basic ping
	result.Tests["ping"] = hc.testPing(conn.DB)
	
	// Test 2: Simple query
	result.Tests["query"] = hc.testSimpleQuery(conn.DB)
	
	// Test 3: Transaction
	result.Tests["transaction"] = hc.testTransaction(conn.DB)
	
	// Test 4: Connection limits
	result.Tests["limits"] = hc.testConnectionLimits(conn.DB)
	
	// Calculate overall score
	result.OverallScore = hc.calculateOverallScore(result.Tests)
	result.Healthy = result.OverallScore >= 0.7
	
	return result
}

// MonitorConnectionHealth starts continuous health monitoring for a connection
func (hc *HealthChecker) MonitorConnectionHealth(conn *PooledConnection) {
	ticker := time.NewTicker(hc.checkInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			hc.CheckConnection(conn)
		}
	}
}

// GetHealthSummary returns a summary of all connection health
func (hc *HealthChecker) GetHealthSummary() *HealthSummary {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	
	summary := &HealthSummary{
		TotalConnections:    len(hc.healthHistory),
		HealthyConnections:  0,
		DegradedConnections: 0,
		UnhealthyConnections: 0,
		Timestamp:           time.Now(),
	}
	
	for _, history := range hc.healthHistory {
		if history.ConsecutiveFailures == 0 {
			summary.HealthyConnections++
		} else if history.ConsecutiveFailures < hc.failureThreshold {
			summary.DegradedConnections++
		} else {
			summary.UnhealthyConnections++
		}
	}
	
	// Calculate overall health percentage
	if summary.TotalConnections > 0 {
		summary.OverallHealthPercentage = float64(summary.HealthyConnections) / float64(summary.TotalConnections) * 100
	}
	
	return summary
}

// Private methods

func (hc *HealthChecker) recordHealthCheck(connID string, result HealthCheckResult) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	
	history, exists := hc.healthHistory[connID]
	if !exists {
		history = &ConnectionHealthHistory{
			ConnectionID: connID,
			HealthChecks: make([]HealthCheckResult, 0),
		}
		hc.healthHistory[connID] = history
	}
	
	// Add new health check result
	history.HealthChecks = append(history.HealthChecks, result)
	history.TotalChecks++
	
	if result.Success {
		history.SuccessfulChecks++
		history.ConsecutiveFailures = 0
		history.LastHealthy = result.Timestamp
	} else {
		history.ConsecutiveFailures++
	}
	
	// Update average latency
	totalLatency := time.Duration(0)
	for _, check := range history.HealthChecks {
		totalLatency += check.Latency
	}
	history.AverageLatency = totalLatency / time.Duration(len(history.HealthChecks))
	
	// Keep only recent health checks (last 100)
	if len(history.HealthChecks) > 100 {
		history.HealthChecks = history.HealthChecks[len(history.HealthChecks)-100:]
	}
}

func (hc *HealthChecker) calculateSuccessRate(history *ConnectionHealthHistory) float64 {
	if history == nil || history.TotalChecks == 0 {
		return 0
	}
	return float64(history.SuccessfulChecks) / float64(history.TotalChecks)
}

func (hc *HealthChecker) testPing(db *sql.DB) HealthTestResult {
	startTime := time.Now()
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	err := db.PingContext(ctx)
	latency := time.Since(startTime)
	
	return HealthTestResult{
		Name:    "ping",
		Success: err == nil,
		Latency: latency,
		Error:   getErrorString(err),
		Score:   getTestScore(err == nil, latency, 100*time.Millisecond),
	}
}

func (hc *HealthChecker) testSimpleQuery(db *sql.DB) HealthTestResult {
	startTime := time.Now()
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	_, err := db.QueryContext(ctx, "SELECT 1")
	latency := time.Since(startTime)
	
	return HealthTestResult{
		Name:    "simple_query",
		Success: err == nil,
		Latency: latency,
		Error:   getErrorString(err),
		Score:   getTestScore(err == nil, latency, 200*time.Millisecond),
	}
}

func (hc *HealthChecker) testTransaction(db *sql.DB) HealthTestResult {
	startTime := time.Now()
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return HealthTestResult{
			Name:    "transaction",
			Success: false,
			Latency: time.Since(startTime),
			Error:   err.Error(),
			Score:   0,
		}
	}
	
	err = tx.Rollback()
	latency := time.Since(startTime)
	
	return HealthTestResult{
		Name:    "transaction",
		Success: err == nil,
		Latency: latency,
		Error:   getErrorString(err),
		Score:   getTestScore(err == nil, latency, 100*time.Millisecond),
	}
}

func (hc *HealthChecker) testConnectionLimits(db *sql.DB) HealthTestResult {
	// This is a simplified test - in practice, you'd check current connection count
	// against database limits
	
	stats := db.Stats()
	
	success := stats.OpenConnections < stats.MaxOpenConnections
	score := 1.0
	if !success {
		score = 0.5
	}
	
	return HealthTestResult{
		Name:    "connection_limits",
		Success: success,
		Latency: 0,
		Error:   "",
		Score:   score,
	}
}

func (hc *HealthChecker) calculateOverallScore(tests map[string]HealthTestResult) float64 {
	if len(tests) == 0 {
		return 0
	}
	
	totalScore := 0.0
	for _, test := range tests {
		totalScore += test.Score
	}
	
	return totalScore / float64(len(tests))
}

// Helper functions

func getErrorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func getTestScore(success bool, latency, threshold time.Duration) float64 {
	if !success {
		return 0
	}
	
	if latency <= threshold {
		return 1.0
	}
	
	// Degraded score based on latency
	ratio := float64(threshold) / float64(latency)
	if ratio < 0.1 {
		return 0.1
	}
	return ratio
}

// Data structures

// AdvancedHealthResult represents comprehensive health check results
type AdvancedHealthResult struct {
	ConnectionID   string                        `json:"connection_id"`
	Timestamp      time.Time                     `json:"timestamp"`
	Healthy        bool                          `json:"healthy"`
	OverallScore   float64                       `json:"overall_score"`
	Tests          map[string]HealthTestResult   `json:"tests"`
}

// HealthTestResult represents the result of a specific health test
type HealthTestResult struct {
	Name    string        `json:"name"`
	Success bool          `json:"success"`
	Latency time.Duration `json:"latency"`
	Error   string        `json:"error,omitempty"`
	Score   float64       `json:"score"`
}

// HealthSummary provides an overview of connection health
type HealthSummary struct {
	TotalConnections         int       `json:"total_connections"`
	HealthyConnections       int       `json:"healthy_connections"`
	DegradedConnections      int       `json:"degraded_connections"`
	UnhealthyConnections     int       `json:"unhealthy_connections"`
	OverallHealthPercentage  float64   `json:"overall_health_percentage"`
	Timestamp                time.Time `json:"timestamp"`
}