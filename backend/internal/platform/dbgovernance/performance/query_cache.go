// Package performance provides query result caching capabilities
package performance

import (
	"context"
	"crypto/md5"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"openpenpal-backend/internal/platform/dbgovernance/core"
)

// QueryCache provides intelligent query result caching
type QueryCache struct {
	config        *core.QueryPerformanceConfig
	cache         map[string]*CacheEntry
	patterns      map[string]*QueryPattern
	mu            sync.RWMutex
	maxSize       int64
	currentSize   int64
	
	// Statistics
	hits          int64
	misses        int64
	evictions     int64
}

// CacheEntry represents a cached query result
type CacheEntry struct {
	Key         string
	Query       string
	Result      interface{}
	CachedAt    time.Time
	ExpiresAt   time.Time
	AccessCount int64
	LastAccess  time.Time
	Size        int64
}

// NewQueryCache creates a new query cache
func NewQueryCache(config *core.QueryPerformanceConfig) *QueryCache {
	return &QueryCache{
		config:      config,
		cache:       make(map[string]*CacheEntry),
		patterns:    make(map[string]*QueryPattern),
		maxSize:     config.QueryCacheMaxSize,
	}
}

// Get retrieves a cached query result
func (qc *QueryCache) Get(ctx context.Context, query string, params []interface{}) (interface{}, bool) {
	key := qc.generateCacheKey(query, params)
	
	qc.mu.RLock()
	entry, exists := qc.cache[key]
	qc.mu.RUnlock()
	
	if !exists {
		qc.recordMiss()
		return nil, false
	}
	
	// Check if entry has expired
	if time.Now().After(entry.ExpiresAt) {
		qc.mu.Lock()
		delete(qc.cache, key)
		qc.currentSize -= entry.Size
		qc.mu.Unlock()
		qc.recordMiss()
		return nil, false
	}
	
	// Update access statistics
	qc.mu.Lock()
	entry.AccessCount++
	entry.LastAccess = time.Now()
	qc.mu.Unlock()
	
	qc.recordHit()
	return entry.Result, true
}

// Put stores a query result in the cache
func (qc *QueryCache) Put(ctx context.Context, query string, params []interface{}, result interface{}) error {
	if !qc.shouldCache(query) {
		return nil
	}
	
	key := qc.generateCacheKey(query, params)
	ttl := qc.getTTLForQuery(query)
	size := qc.estimateSize(result)
	
	entry := &CacheEntry{
		Key:         key,
		Query:       query,
		Result:      result,
		CachedAt:    time.Now(),
		ExpiresAt:   time.Now().Add(ttl),
		AccessCount: 0,
		LastAccess:  time.Now(),
		Size:        size,
	}
	
	qc.mu.Lock()
	defer qc.mu.Unlock()
	
	// Check if we need to evict entries
	if qc.currentSize+size > qc.maxSize {
		qc.evictEntries(size)
	}
	
	// Store the entry
	qc.cache[key] = entry
	qc.currentSize += size
	
	return nil
}

// EnableCachingForPattern enables caching for queries matching a pattern
func (qc *QueryCache) EnableCachingForPattern(pattern string, ttl time.Duration) error {
	qc.mu.Lock()
	defer qc.mu.Unlock()
	
	qc.patterns[pattern] = &QueryPattern{
		Pattern:  pattern,
		TTL:      ttl,
		Enabled:  true,
		LastUsed: time.Now(),
	}
	
	return nil
}

// InvalidatePattern invalidates all cached entries matching a pattern
func (qc *QueryCache) InvalidatePattern(pattern string) int {
	qc.mu.Lock()
	defer qc.mu.Unlock()
	
	count := 0
	for key, entry := range qc.cache {
		if qc.matchesPattern(entry.Query, pattern) {
			delete(qc.cache, key)
			qc.currentSize -= entry.Size
			count++
		}
	}
	
	return count
}

// GetStatistics returns cache statistics
func (qc *QueryCache) GetStatistics() *CacheStatistics {
	qc.mu.RLock()
	defer qc.mu.RUnlock()
	
	totalRequests := qc.hits + qc.misses
	hitRate := 0.0
	if totalRequests > 0 {
		hitRate = float64(qc.hits) / float64(totalRequests)
	}
	
	return &CacheStatistics{
		Entries:       len(qc.cache),
		CurrentSize:   qc.currentSize,
		MaxSize:       qc.maxSize,
		Hits:          qc.hits,
		Misses:        qc.misses,
		Evictions:     qc.evictions,
		HitRate:       hitRate,
		MemoryUsage:   float64(qc.currentSize) / float64(qc.maxSize) * 100,
	}
}

// Clear removes all entries from the cache
func (qc *QueryCache) Clear() {
	qc.mu.Lock()
	defer qc.mu.Unlock()
	
	qc.cache = make(map[string]*CacheEntry)
	qc.currentSize = 0
}

// Private methods

func (qc *QueryCache) generateCacheKey(query string, params []interface{}) string {
	// Create a unique key combining query and parameters
	keyData := fmt.Sprintf("%s:%v", query, params)
	hash := md5.Sum([]byte(keyData))
	return fmt.Sprintf("%x", hash)
}

func (qc *QueryCache) shouldCache(query string) bool {
	if !qc.config.EnableQueryCache {
		return false
	}
	
	// Check if query matches any caching patterns
	qc.mu.RLock()
	defer qc.mu.RUnlock()
	
	for pattern, patternConfig := range qc.patterns {
		if patternConfig.Enabled && qc.matchesPattern(query, pattern) {
			return true
		}
	}
	
	// Default caching rules
	queryLower := strings.ToLower(query)
	
	// Cache SELECT queries
	if strings.HasPrefix(queryLower, "select") {
		// Don't cache queries with certain functions
		if strings.Contains(queryLower, "now()") ||
		   strings.Contains(queryLower, "random()") ||
		   strings.Contains(queryLower, "uuid_generate") {
			return false
		}
		return true
	}
	
	return false
}

func (qc *QueryCache) getTTLForQuery(query string) time.Duration {
	// Check pattern-specific TTL
	qc.mu.RLock()
	for pattern, patternConfig := range qc.patterns {
		if qc.matchesPattern(query, pattern) {
			qc.mu.RUnlock()
			return patternConfig.TTL
		}
	}
	qc.mu.RUnlock()
	
	// Default TTL from config
	return qc.config.QueryCacheTTL
}

func (qc *QueryCache) matchesPattern(query, pattern string) bool {
	// Simple pattern matching - in practice, you'd use more sophisticated matching
	return strings.Contains(strings.ToLower(query), strings.ToLower(pattern))
}

func (qc *QueryCache) estimateSize(result interface{}) int64 {
	// Simplified size estimation
	// In practice, you'd implement more accurate size calculation
	return 1024 // 1KB default estimate
}

func (qc *QueryCache) evictEntries(neededSize int64) {
	// LRU eviction strategy
	type entryWithKey struct {
		key   string
		entry *CacheEntry
	}
	
	var entries []entryWithKey
	for key, entry := range qc.cache {
		entries = append(entries, entryWithKey{key: key, entry: entry})
	}
	
	// Sort by last access time (LRU first)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].entry.LastAccess.Before(entries[j].entry.LastAccess)
	})
	
	freedSize := int64(0)
	for _, item := range entries {
		if freedSize >= neededSize {
			break
		}
		
		delete(qc.cache, item.key)
		qc.currentSize -= item.entry.Size
		freedSize += item.entry.Size
		qc.evictions++
	}
}

func (qc *QueryCache) recordHit() {
	qc.mu.Lock()
	qc.hits++
	qc.mu.Unlock()
}

func (qc *QueryCache) recordMiss() {
	qc.mu.Lock()
	qc.misses++
	qc.mu.Unlock()
}

// Background cleanup

func (qc *QueryCache) StartCleanup() {
	go qc.cleanupLoop()
}

func (qc *QueryCache) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			qc.cleanupExpiredEntries()
		}
	}
}

func (qc *QueryCache) cleanupExpiredEntries() {
	qc.mu.Lock()
	defer qc.mu.Unlock()
	
	now := time.Now()
	for key, entry := range qc.cache {
		if now.After(entry.ExpiresAt) {
			delete(qc.cache, key)
			qc.currentSize -= entry.Size
		}
	}
}

// Data structures

// CacheStatistics represents cache performance statistics
type CacheStatistics struct {
	Entries     int     `json:"entries"`
	CurrentSize int64   `json:"current_size"`
	MaxSize     int64   `json:"max_size"`
	Hits        int64   `json:"hits"`
	Misses      int64   `json:"misses"`
	Evictions   int64   `json:"evictions"`
	HitRate     float64 `json:"hit_rate"`
	MemoryUsage float64 `json:"memory_usage"`
}

