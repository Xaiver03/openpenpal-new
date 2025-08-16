// Package performance provides slow query logging and analysis
package performance

import (
	"strings"
	"sync"
	"time"

	"openpenpal-backend/internal/platform/dbgovernance/core"
)

// SlowQueryLogger tracks and logs slow queries
type SlowQueryLogger struct {
	config      *core.QueryPerformanceConfig
	slowQueries []*core.SlowQuery
	mu          sync.RWMutex
	maxEntries  int
}

// NewSlowQueryLogger creates a new slow query logger
func NewSlowQueryLogger(config *core.QueryPerformanceConfig) *SlowQueryLogger {
	return &SlowQueryLogger{
		config:     config,
		maxEntries: config.MaxSlowQueries,
	}
}

// GetSlowQueries returns slow queries that exceed the threshold
func (sql *SlowQueryLogger) GetSlowQueries(threshold time.Duration) ([]*core.SlowQuery, error) {
	sql.mu.RLock()
	defer sql.mu.RUnlock()
	
	var result []*core.SlowQuery
	for _, query := range sql.slowQueries {
		if query.Duration >= threshold {
			result = append(result, query)
		}
	}
	
	return result, nil
}

// AIOptimizer provides AI-powered query optimization
type AIOptimizer struct {
	optimizationRules map[string]func(string) []string
}

// NewAIOptimizer creates a new AI optimizer
func NewAIOptimizer() *AIOptimizer {
	optimizer := &AIOptimizer{
		optimizationRules: make(map[string]func(string) []string),
	}
	optimizer.initializeRules()
	return optimizer
}

// OptimizeQuery provides AI-powered optimization suggestions
func (ai *AIOptimizer) OptimizeQuery(query string, history *QueryHistory) []string {
	suggestions := make([]string, 0)
	
	// Apply optimization rules
	for ruleName, rule := range ai.optimizationRules {
		ruleSuggestions := rule(query)
		suggestions = append(suggestions, ruleSuggestions...)
		_ = ruleName // Avoid unused variable
	}
	
	return suggestions
}

func (ai *AIOptimizer) initializeRules() {
	ai.optimizationRules["avoid_select_star"] = func(query string) []string {
		if strings.Contains(strings.ToLower(query), "select *") {
			return []string{"Consider selecting specific columns instead of SELECT *"}
		}
		return nil
	}
	
	ai.optimizationRules["add_limit"] = func(query string) []string {
		lower := strings.ToLower(query)
		if strings.Contains(lower, "order by") && !strings.Contains(lower, "limit") {
			return []string{"Consider adding LIMIT clause to ORDER BY queries"}
		}
		return nil
	}
}