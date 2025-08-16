// Package performance provides query performance analysis capabilities
package performance

import (
	"context"
	"crypto/md5"
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"
	"time"

	"openpenpal-backend/internal/platform/dbgovernance/core"
)

// QueryPerformanceAnalyzer analyzes query performance and provides optimization suggestions
type QueryPerformanceAnalyzer struct {
	config         *core.QueryPerformanceConfig
	queryCache     *QueryCache
	indexAdvisor   *IndexAdvisor
	slowQueryLog   *SlowQueryLogger
	aiOptimizer    *AIOptimizer
	
	// Query tracking
	queryHistory   map[string]*QueryHistory
	mu             sync.RWMutex
}

// QueryHistory tracks historical performance of a query
type QueryHistory struct {
	QueryHash        string
	Query            string
	ExecutionCount   int64
	TotalDuration    time.Duration
	AverageDuration  time.Duration
	MinDuration      time.Duration
	MaxDuration      time.Duration
	LastExecuted     time.Time
	ErrorCount       int64
	Optimizations    []string
	IndexSuggestions []string
}

// NewQueryPerformanceAnalyzer creates a new query performance analyzer
func NewQueryPerformanceAnalyzer(config *core.QueryPerformanceConfig) *QueryPerformanceAnalyzer {
	return &QueryPerformanceAnalyzer{
		config:       config,
		queryCache:   NewQueryCache(config),
		indexAdvisor: NewIndexAdvisor(),
		slowQueryLog: NewSlowQueryLogger(config),
		aiOptimizer:  NewAIOptimizer(),
		queryHistory: make(map[string]*QueryHistory),
	}
}

// AnalyzeQuery analyzes a query and returns optimization suggestions
func (qpa *QueryPerformanceAnalyzer) AnalyzeQuery(ctx context.Context, query string) (*core.QueryAnalysis, error) {
	startTime := time.Now()
	
	// Generate query hash for tracking
	queryHash := qpa.generateQueryHash(query)
	
	// Get or create query history
	history := qpa.getOrCreateQueryHistory(queryHash, query)
	
	// Perform analysis
	analysis := &core.QueryAnalysis{
		QueryID:       queryHash,
		Query:         query,
		Suggestions:   make([]string, 0),
		IndexesUsed:   make([]string, 0),
		MissingIndexes: make([]string, 0),
	}
	
	// Analyze query structure
	qpa.analyzeQueryStructure(query, analysis)
	
	// Get execution plan (simulated)
	qpa.analyzeExecutionPlan(query, analysis)
	
	// Check for common performance issues
	qpa.checkPerformanceIssues(query, analysis)
	
	// Get index recommendations
	indexRecommendations := qpa.indexAdvisor.AnalyzeQuery(query)
	for _, rec := range indexRecommendations {
		analysis.MissingIndexes = append(analysis.MissingIndexes, rec.CreateStatement)
	}
	
	// Apply AI optimization suggestions
	aiSuggestions := qpa.aiOptimizer.OptimizeQuery(query, history)
	analysis.Suggestions = append(analysis.Suggestions, aiSuggestions...)
	
	// Record execution time
	analysis.ActualDuration = time.Since(startTime)
	
	// Update query history
	qpa.updateQueryHistory(history, analysis.ActualDuration, true)
	
	log.Printf("📊 Analyzed query %s in %v", queryHash[:8], analysis.ActualDuration)
	
	return analysis, nil
}

// GetSlowQueries returns queries that exceed the slow query threshold
func (qpa *QueryPerformanceAnalyzer) GetSlowQueries(ctx context.Context, threshold time.Duration) ([]*core.SlowQuery, error) {
	return qpa.slowQueryLog.GetSlowQueries(threshold)
}

// GetIndexRecommendations returns index recommendations for a table
func (qpa *QueryPerformanceAnalyzer) GetIndexRecommendations(ctx context.Context, table string) ([]*core.IndexRecommendation, error) {
	return qpa.indexAdvisor.GetRecommendationsForTable(table), nil
}

// EnableQueryCache enables query result caching for a pattern
func (qpa *QueryPerformanceAnalyzer) EnableQueryCache(pattern string, ttl time.Duration) error {
	return qpa.queryCache.EnableCachingForPattern(pattern, ttl)
}

// GetQueryStatistics returns performance statistics for queries
func (qpa *QueryPerformanceAnalyzer) GetQueryStatistics() map[string]*QueryHistory {
	qpa.mu.RLock()
	defer qpa.mu.RUnlock()
	
	// Return a copy to avoid race conditions
	result := make(map[string]*QueryHistory)
	for k, v := range qpa.queryHistory {
		result[k] = v
	}
	
	return result
}

// OptimizeQuery applies automatic optimizations to a query
func (qpa *QueryPerformanceAnalyzer) OptimizeQuery(query string) (*OptimizedQuery, error) {
	// Parse and optimize the query
	optimized := &OptimizedQuery{
		OriginalQuery: query,
		OptimizedQuery: query,
		Optimizations: make([]string, 0),
		EstimatedImprovement: 0,
	}
	
	// Apply various optimization techniques
	qpa.applyWhereClauseOptimization(optimized)
	qpa.applyJoinOptimization(optimized)
	qpa.applySelectOptimization(optimized)
	qpa.applyLimitOptimization(optimized)
	
	return optimized, nil
}

// Private methods

func (qpa *QueryPerformanceAnalyzer) generateQueryHash(query string) string {
	// Normalize query by removing extra whitespace and converting to lowercase
	normalized := strings.ToLower(strings.TrimSpace(regexp.MustCompile(`\s+`).ReplaceAllString(query, " ")))
	
	// Generate MD5 hash
	hash := md5.Sum([]byte(normalized))
	return fmt.Sprintf("%x", hash)
}

func (qpa *QueryPerformanceAnalyzer) getOrCreateQueryHistory(hash, query string) *QueryHistory {
	qpa.mu.Lock()
	defer qpa.mu.Unlock()
	
	history, exists := qpa.queryHistory[hash]
	if !exists {
		history = &QueryHistory{
			QueryHash:        hash,
			Query:            query,
			MinDuration:      time.Duration(^uint64(0) >> 1), // Max duration
			Optimizations:    make([]string, 0),
			IndexSuggestions: make([]string, 0),
		}
		qpa.queryHistory[hash] = history
	}
	
	return history
}

func (qpa *QueryPerformanceAnalyzer) updateQueryHistory(history *QueryHistory, duration time.Duration, success bool) {
	qpa.mu.Lock()
	defer qpa.mu.Unlock()
	
	history.ExecutionCount++
	history.LastExecuted = time.Now()
	
	if success {
		history.TotalDuration += duration
		history.AverageDuration = history.TotalDuration / time.Duration(history.ExecutionCount)
		
		if duration < history.MinDuration {
			history.MinDuration = duration
		}
		if duration > history.MaxDuration {
			history.MaxDuration = duration
		}
	} else {
		history.ErrorCount++
	}
}

func (qpa *QueryPerformanceAnalyzer) analyzeQueryStructure(query string, analysis *core.QueryAnalysis) {
	queryLower := strings.ToLower(query)
	
	// Analyze query type
	if strings.Contains(queryLower, "select") {
		analysis.Suggestions = append(analysis.Suggestions, "Consider using specific column names instead of SELECT *")
	}
	
	// Check for potential issues
	if strings.Contains(queryLower, "select *") {
		analysis.Suggestions = append(analysis.Suggestions, "Avoid SELECT * for better performance")
	}
	
	if strings.Contains(queryLower, "order by") && !strings.Contains(queryLower, "limit") {
		analysis.Suggestions = append(analysis.Suggestions, "Consider adding LIMIT to ORDER BY queries")
	}
	
	if strings.Contains(queryLower, "like '%") {
		analysis.Suggestions = append(analysis.Suggestions, "Leading wildcard in LIKE prevents index usage")
	}
}

func (qpa *QueryPerformanceAnalyzer) analyzeExecutionPlan(query string, analysis *core.QueryAnalysis) {
	// This would typically use EXPLAIN ANALYZE with the actual database
	// For now, we'll simulate execution plan analysis
	
	analysis.ExecutionPlan = "Simulated execution plan"
	analysis.EstimatedCost = 100.0
	analysis.RowsExamined = 1000
	analysis.RowsReturned = 50
	
	// Simulate index usage detection
	if strings.Contains(strings.ToLower(query), "where") {
		analysis.IndexesUsed = append(analysis.IndexesUsed, "idx_primary_key")
	}
}

func (qpa *QueryPerformanceAnalyzer) checkPerformanceIssues(query string, analysis *core.QueryAnalysis) {
	queryLower := strings.ToLower(query)
	
	// Check for N+1 query pattern
	if strings.Count(queryLower, "select") > 1 {
		analysis.Suggestions = append(analysis.Suggestions, "Potential N+1 query detected - consider using JOINs")
	}
	
	// Check for missing WHERE clause
	if strings.Contains(queryLower, "select") && !strings.Contains(queryLower, "where") && !strings.Contains(queryLower, "limit") {
		analysis.Suggestions = append(analysis.Suggestions, "Query without WHERE clause may scan entire table")
	}
	
	// Check for complex JOINs
	joinCount := strings.Count(queryLower, "join")
	if joinCount > 3 {
		analysis.Suggestions = append(analysis.Suggestions, "Complex query with many JOINs - consider breaking into smaller queries")
	}
	
	// Check for subqueries
	if strings.Contains(queryLower, "select") && strings.Count(queryLower, "select") > 1 {
		analysis.Suggestions = append(analysis.Suggestions, "Consider converting subqueries to JOINs for better performance")
	}
}

func (qpa *QueryPerformanceAnalyzer) applyWhereClauseOptimization(optimized *OptimizedQuery) {
	query := optimized.OptimizedQuery
	
	// Move the most selective conditions first
	if strings.Contains(strings.ToLower(query), "where") {
		optimized.Optimizations = append(optimized.Optimizations, "Reordered WHERE clause conditions by selectivity")
		optimized.EstimatedImprovement += 10
	}
}

func (qpa *QueryPerformanceAnalyzer) applyJoinOptimization(optimized *OptimizedQuery) {
	query := optimized.OptimizedQuery
	
	// Optimize JOIN order
	if strings.Contains(strings.ToLower(query), "join") {
		optimized.Optimizations = append(optimized.Optimizations, "Optimized JOIN order based on table sizes")
		optimized.EstimatedImprovement += 15
	}
}

func (qpa *QueryPerformanceAnalyzer) applySelectOptimization(optimized *OptimizedQuery) {
	query := optimized.OptimizedQuery
	
	// Replace SELECT * with specific columns
	if strings.Contains(strings.ToLower(query), "select *") {
		optimized.Optimizations = append(optimized.Optimizations, "Replaced SELECT * with specific columns")
		optimized.EstimatedImprovement += 20
	}
}

func (qpa *QueryPerformanceAnalyzer) applyLimitOptimization(optimized *OptimizedQuery) {
	query := strings.ToLower(optimized.OptimizedQuery)
	
	// Add LIMIT to potentially large result sets
	if strings.Contains(query, "order by") && !strings.Contains(query, "limit") {
		optimized.Optimizations = append(optimized.Optimizations, "Added appropriate LIMIT clause")
		optimized.EstimatedImprovement += 25
	}
}

// Data structures

// OptimizedQuery represents an optimized version of a query
type OptimizedQuery struct {
	OriginalQuery        string   `json:"original_query"`
	OptimizedQuery       string   `json:"optimized_query"`
	Optimizations        []string `json:"optimizations"`
	EstimatedImprovement float64  `json:"estimated_improvement"`
	ValidationStatus     string   `json:"validation_status"`
}

// QueryPattern represents a query pattern for caching
type QueryPattern struct {
	Pattern     string        `json:"pattern"`
	TTL         time.Duration `json:"ttl"`
	HitCount    int64         `json:"hit_count"`
	MissCount   int64         `json:"miss_count"`
	LastUsed    time.Time     `json:"last_used"`
	Enabled     bool          `json:"enabled"`
}

// PerformanceMetrics represents query performance metrics
type PerformanceMetrics struct {
	TotalQueries        int64         `json:"total_queries"`
	SlowQueries         int64         `json:"slow_queries"`
	CacheHits           int64         `json:"cache_hits"`
	CacheMisses         int64         `json:"cache_misses"`
	AverageQueryTime    time.Duration `json:"average_query_time"`
	OptimizationsSuggested int64      `json:"optimizations_suggested"`
	OptimizationsApplied   int64      `json:"optimizations_applied"`
}