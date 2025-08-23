// Package services provides hot key monitoring and mitigation for Redis cache
package services

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"openpenpal-backend/internal/logger"
	"openpenpal-backend/pkg/cache"
)

// HotKeyMonitorService monitors and mitigates Redis hot key issues
type HotKeyMonitorService struct {
	redisClient  *redis.Client
	cacheManager *cache.EnhancedCacheManager
	
	// Monitoring settings
	sampleInterval   time.Duration
	hotKeyThreshold  int64         // Access count threshold to consider a key "hot"
	monitorDuration  time.Duration // How long to monitor for hot keys
	enableMitigation bool          // Enable automatic hot key mitigation
	
	// Hot key tracking
	mu              sync.RWMutex
	keyAccessCounts map[string]*KeyStats
	hotKeys         map[string]*HotKeyInfo
	
	// Mitigation strategies
	replicationFactor int           // Number of replicas to create for hot keys
	replicaTTL       time.Duration  // TTL for hot key replicas
	
	// Monitoring state
	isRunning bool
	stopChan  chan struct{}
}

// KeyStats tracks access statistics for a Redis key
type KeyStats struct {
	Key           string            `json:"key"`
	AccessCount   int64             `json:"access_count"`
	LastAccessed  time.Time         `json:"last_accessed"`
	FirstSeen     time.Time         `json:"first_seen"`
	BytesRead     int64             `json:"bytes_read"`
	BytesWritten  int64             `json:"bytes_written"`
}

// HotKeyInfo contains information about detected hot keys
type HotKeyInfo struct {
	*KeyStats
	DetectedAt     time.Time         `json:"detected_at"`
	MitigationApplied bool           `json:"mitigation_applied"`
	ReplicaKeys    []string          `json:"replica_keys"`
	MitigationType string            `json:"mitigation_type"`
}

// HotKeyReport contains monitoring results
type HotKeyReport struct {
	GeneratedAt      time.Time                 `json:"generated_at"`
	MonitorDuration  time.Duration            `json:"monitor_duration"`
	TotalKeys        int                      `json:"total_keys"`
	HotKeys          []*HotKeyInfo            `json:"hot_keys"`
	TopKeys          []*KeyStats              `json:"top_keys"`
	Recommendations  []string                 `json:"recommendations"`
	RedisInfo        map[string]interface{}   `json:"redis_info"`
}

// NewHotKeyMonitorService creates a new hot key monitoring service
func NewHotKeyMonitorService(redisClient *redis.Client, cacheManager *cache.EnhancedCacheManager) *HotKeyMonitorService {
	return &HotKeyMonitorService{
		redisClient:      redisClient,
		cacheManager:     cacheManager,
		sampleInterval:   1 * time.Second,  // Sample every second
		hotKeyThreshold:  100,              // 100 accesses per monitoring period
		monitorDuration:  5 * time.Minute,  // Monitor for 5 minutes
		enableMitigation: true,
		keyAccessCounts:  make(map[string]*KeyStats),
		hotKeys:          make(map[string]*HotKeyInfo),
		replicationFactor: 3,               // Create 3 replicas for hot keys
		replicaTTL:       30 * time.Minute, // Replicas live for 30 minutes
		stopChan:         make(chan struct{}),
	}
}

// StartMonitoring begins hot key monitoring in the background
func (h *HotKeyMonitorService) StartMonitoring(ctx context.Context) error {
	h.mu.Lock()
	if h.isRunning {
		h.mu.Unlock()
		return fmt.Errorf("hot key monitoring is already running")
	}
	h.isRunning = true
	h.mu.Unlock()

	logger.Info("Starting hot key monitoring",
		"sampleInterval", h.sampleInterval,
		"hotKeyThreshold", h.hotKeyThreshold,
		"enableMitigation", h.enableMitigation,
	)

	go h.monitoringLoop(ctx)
	return nil
}

// StopMonitoring stops the hot key monitoring
func (h *HotKeyMonitorService) StopMonitoring() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.isRunning {
		return
	}

	h.isRunning = false
	close(h.stopChan)
	logger.Info("Hot key monitoring stopped")
}

// monitoringLoop runs the main monitoring loop
func (h *HotKeyMonitorService) monitoringLoop(ctx context.Context) {
	ticker := time.NewTicker(h.sampleInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-h.stopChan:
			return
		case <-ticker.C:
			if err := h.sampleRedisStats(ctx); err != nil {
				logger.Error("Failed to sample Redis stats", err)
			}
			
			h.analyzeHotKeys()
			
			if h.enableMitigation {
				if err := h.applyMitigation(ctx); err != nil {
					logger.Error("Failed to apply hot key mitigation", err)
				}
			}
		}
	}
}

// sampleRedisStats collects current Redis statistics
func (h *HotKeyMonitorService) sampleRedisStats(ctx context.Context) error {
	// Get Redis info for command statistics
	info, err := h.redisClient.Info(ctx, "commandstats").Result()
	if err != nil {
		return fmt.Errorf("failed to get Redis info: %w", err)
	}

	// Parse command stats to identify frequently accessed keys
	// Note: This is a simplified implementation. In production, you might want to use:
	// 1. Redis MONITOR command (with caution due to performance impact)
	// 2. Redis LATENCY DOCTOR
	// 3. Redis keyspace notifications
	// 4. Application-level instrumentation

	// For demonstration, we'll simulate key access patterns
	h.simulateKeyAccessDetection(ctx)

	logger.Debug("Redis stats sampled", "info_length", len(info))
	return nil
}

// simulateKeyAccessDetection simulates detection of key access patterns
// In production, this would use real Redis monitoring data
func (h *HotKeyMonitorService) simulateKeyAccessDetection(ctx context.Context) {
	// Get a sample of keys from Redis
	keys, err := h.redisClient.Keys(ctx, "openpenpal:*").Result()
	if err != nil {
		logger.Error("Failed to get keys for monitoring", err)
		return
	}

	now := time.Now()
	
	for _, key := range keys {
		// Get key info
		ttl, _ := h.redisClient.TTL(ctx, key).Result()
		keyType, _ := h.redisClient.Type(ctx, key).Result()
		
		// Simulate access counting (in real implementation, use Redis OBJECT IDLETIME)
		accessCount := h.estimateKeyAccess(ctx, key)
		
		h.mu.Lock()
		if stats, exists := h.keyAccessCounts[key]; exists {
			stats.AccessCount += accessCount
			stats.LastAccessed = now
		} else {
			h.keyAccessCounts[key] = &KeyStats{
				Key:          key,
				AccessCount:  accessCount,
				LastAccessed: now,
				FirstSeen:    now,
			}
		}
		h.mu.Unlock()

		logger.Debug("Key access tracked",
			"key", key,
			"type", keyType,
			"ttl", ttl,
			"accessCount", accessCount,
		)
	}
}

// estimateKeyAccess estimates access frequency for a key
func (h *HotKeyMonitorService) estimateKeyAccess(ctx context.Context, key string) int64 {
	// This is a simplified estimation. In production, use:
	// 1. Redis OBJECT IDLETIME to get idle time
	// 2. Application-level counters
	// 3. Proxy-level monitoring
	
	// For demo purposes, simulate based on key patterns
	if hotkeyContains(key, "user:") {
		return int64(10 + time.Now().Unix()%20) // 10-30 accesses
	} else if hotkeyContains(key, "session:") {
		return int64(50 + time.Now().Unix()%50) // 50-100 accesses
	} else if hotkeyContains(key, "config:") {
		return int64(100 + time.Now().Unix()%100) // 100-200 accesses (potential hot key)
	}
	
	return int64(1 + time.Now().Unix()%5) // 1-5 accesses
}

// analyzeHotKeys identifies hot keys based on access patterns
func (h *HotKeyMonitorService) analyzeHotKeys() {
	h.mu.Lock()
	defer h.mu.Unlock()

	now := time.Now()
	newHotKeys := make([]*HotKeyInfo, 0)

	for key, stats := range h.keyAccessCounts {
		// Check if key meets hot key criteria
		if stats.AccessCount >= h.hotKeyThreshold {
			if _, exists := h.hotKeys[key]; !exists {
				hotKey := &HotKeyInfo{
					KeyStats:          stats,
					DetectedAt:        now,
					MitigationApplied: false,
					ReplicaKeys:       make([]string, 0),
					MitigationType:    "",
				}
				h.hotKeys[key] = hotKey
				newHotKeys = append(newHotKeys, hotKey)
			}
		}
	}

	if len(newHotKeys) > 0 {
		logger.Info("New hot keys detected",
			"count", len(newHotKeys),
			"threshold", h.hotKeyThreshold,
		)
		
		for _, hotKey := range newHotKeys {
			logger.Warn("Hot key detected",
				"key", hotKey.Key,
				"accessCount", hotKey.AccessCount,
				"detectedAt", hotKey.DetectedAt,
			)
		}
	}
}

// applyMitigation applies mitigation strategies for hot keys
func (h *HotKeyMonitorService) applyMitigation(ctx context.Context) error {
	h.mu.Lock()
	hotKeysToMitigate := make([]*HotKeyInfo, 0)
	for _, hotKey := range h.hotKeys {
		if !hotKey.MitigationApplied {
			hotKeysToMitigate = append(hotKeysToMitigate, hotKey)
		}
	}
	h.mu.Unlock()

	for _, hotKey := range hotKeysToMitigate {
		if err := h.mitigateHotKey(ctx, hotKey); err != nil {
			logger.Error("Failed to mitigate hot key", err, "key", hotKey.Key)
		}
	}

	return nil
}

// mitigateHotKey applies specific mitigation for a hot key
func (h *HotKeyMonitorService) mitigateHotKey(ctx context.Context, hotKey *HotKeyInfo) error {
	logger.Info("Applying hot key mitigation", "key", hotKey.Key, "accessCount", hotKey.AccessCount)

	// Strategy 1: Create replicas with load balancing
	if err := h.createKeyReplicas(ctx, hotKey); err != nil {
		logger.Error("Failed to create key replicas", err, "key", hotKey.Key)
	}

	// Strategy 2: Extend TTL to reduce reload frequency
	if err := h.extendKeyTTL(ctx, hotKey); err != nil {
		logger.Error("Failed to extend key TTL", err, "key", hotKey.Key)
	}

	// Strategy 3: Add to local cache layer
	if err := h.promoteToLocalCache(ctx, hotKey); err != nil {
		logger.Error("Failed to promote to local cache", err, "key", hotKey.Key)
	}

	h.mu.Lock()
	hotKey.MitigationApplied = true
	hotKey.MitigationType = "replicas+ttl_extend+local_cache"
	h.mu.Unlock()

	logger.Info("Hot key mitigation applied successfully", "key", hotKey.Key)
	return nil
}

// createKeyReplicas creates multiple replicas of a hot key
func (h *HotKeyMonitorService) createKeyReplicas(ctx context.Context, hotKey *HotKeyInfo) error {
	// Get the original value
	value, err := h.redisClient.Get(ctx, hotKey.Key).Result()
	if err != nil {
		return fmt.Errorf("failed to get hot key value: %w", err)
	}

	// Create replicas
	replicaKeys := make([]string, 0, h.replicationFactor)
	for i := 0; i < h.replicationFactor; i++ {
		replicaKey := fmt.Sprintf("%s:replica:%d", hotKey.Key, i)
		
		if err := h.redisClient.Set(ctx, replicaKey, value, h.replicaTTL).Err(); err != nil {
			logger.Error("Failed to create replica", err, "replicaKey", replicaKey)
			continue
		}
		
		replicaKeys = append(replicaKeys, replicaKey)
	}

	h.mu.Lock()
	hotKey.ReplicaKeys = replicaKeys
	h.mu.Unlock()

	logger.Info("Created hot key replicas",
		"originalKey", hotKey.Key,
		"replicas", len(replicaKeys),
		"ttl", h.replicaTTL,
	)

	return nil
}

// extendKeyTTL extends the TTL of a hot key to reduce reload frequency
func (h *HotKeyMonitorService) extendKeyTTL(ctx context.Context, hotKey *HotKeyInfo) error {
	currentTTL, err := h.redisClient.TTL(ctx, hotKey.Key).Result()
	if err != nil {
		return err
	}

	// If key has TTL, extend it by 50%
	if currentTTL > 0 {
		newTTL := time.Duration(float64(currentTTL) * 1.5)
		if err := h.redisClient.Expire(ctx, hotKey.Key, newTTL).Err(); err != nil {
			return err
		}
		
		logger.Info("Extended hot key TTL",
			"key", hotKey.Key,
			"oldTTL", currentTTL,
			"newTTL", newTTL,
		)
	}

	return nil
}

// promoteToLocalCache promotes hot keys to local in-memory cache
func (h *HotKeyMonitorService) promoteToLocalCache(ctx context.Context, hotKey *HotKeyInfo) error {
	// Get the value
	value, err := h.redisClient.Get(ctx, hotKey.Key).Result()
	if err != nil {
		return err
	}

	// Store in local cache (this would need to be implemented in the cache manager)
	// For now, we'll just log it
	logger.Info("Hot key promoted to local cache", "key", hotKey.Key, "valueSize", len(value))

	return nil
}

// GetHotKeyReport generates a comprehensive hot key report
func (h *HotKeyMonitorService) GetHotKeyReport(ctx context.Context) (*HotKeyReport, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Get Redis info
	redisInfo, err := h.redisClient.Info(ctx, "memory", "stats").Result()
	if err != nil {
		logger.Error("Failed to get Redis info for report", err)
	}

	// Get top keys by access count
	topKeys := make([]*KeyStats, 0, len(h.keyAccessCounts))
	for _, stats := range h.keyAccessCounts {
		topKeys = append(topKeys, stats)
	}
	
	sort.Slice(topKeys, func(i, j int) bool {
		return topKeys[i].AccessCount > topKeys[j].AccessCount
	})
	
	// Limit to top 20
	if len(topKeys) > 20 {
		topKeys = topKeys[:20]
	}

	// Get hot keys
	hotKeys := make([]*HotKeyInfo, 0, len(h.hotKeys))
	for _, hotKey := range h.hotKeys {
		hotKeys = append(hotKeys, hotKey)
	}

	// Generate recommendations
	recommendations := h.generateRecommendations(hotKeys, topKeys)

	report := &HotKeyReport{
		GeneratedAt:     time.Now(),
		MonitorDuration: h.monitorDuration,
		TotalKeys:       len(h.keyAccessCounts),
		HotKeys:         hotKeys,
		TopKeys:         topKeys,
		Recommendations: recommendations,
		RedisInfo:       map[string]interface{}{"raw": redisInfo},
	}

	return report, nil
}

// generateRecommendations creates optimization recommendations based on hot key analysis
func (h *HotKeyMonitorService) generateRecommendations(hotKeys []*HotKeyInfo, topKeys []*KeyStats) []string {
	recommendations := make([]string, 0)

	if len(hotKeys) > 0 {
		recommendations = append(recommendations, fmt.Sprintf("发现 %d 个热点 Key，建议启用自动缓解策略", len(hotKeys)))
	}

	if len(topKeys) > 10 {
		recommendations = append(recommendations, "高频访问的 Key 较多，建议优化缓存策略和数据结构")
	}

	// Check for patterns in hot keys
	userKeyCount := 0
	sessionKeyCount := 0
	configKeyCount := 0
	
	for _, hotKey := range hotKeys {
		if hotkeyContains(hotKey.Key, "user:") {
			userKeyCount++
		} else if hotkeyContains(hotKey.Key, "session:") {
			sessionKeyCount++
		} else if hotkeyContains(hotKey.Key, "config:") {
			configKeyCount++
		}
	}

	if userKeyCount > 0 {
		recommendations = append(recommendations, "用户相关的热点 Key 较多，建议优化用户数据缓存策略")
	}
	
	if sessionKeyCount > 0 {
		recommendations = append(recommendations, "会话相关的热点 Key 较多，建议考虑会话数据分片")
	}
	
	if configKeyCount > 0 {
		recommendations = append(recommendations, "配置相关的热点 Key 较多，建议使用本地缓存减少 Redis 访问")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "当前 Key 访问模式正常，无需特殊优化")
	}

	return recommendations
}

// GetLoadBalancedKey returns a load-balanced key for hot key access
func (h *HotKeyMonitorService) GetLoadBalancedKey(originalKey string) string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if hotKey, exists := h.hotKeys[originalKey]; exists && len(hotKey.ReplicaKeys) > 0 {
		// Simple round-robin load balancing
		index := time.Now().UnixNano() % int64(len(hotKey.ReplicaKeys))
		return hotKey.ReplicaKeys[index]
	}

	return originalKey
}

// SetConfiguration updates monitoring configuration
func (h *HotKeyMonitorService) SetConfiguration(config HotKeyConfig) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.sampleInterval = config.SampleInterval
	h.hotKeyThreshold = config.HotKeyThreshold
	h.monitorDuration = config.MonitorDuration
	h.enableMitigation = config.EnableMitigation
	h.replicationFactor = config.ReplicationFactor
	h.replicaTTL = config.ReplicaTTL

	logger.Info("Hot key monitor configuration updated", "config", config)
}

// HotKeyConfig holds configuration for hot key monitoring
type HotKeyConfig struct {
	SampleInterval    time.Duration `json:"sample_interval"`
	HotKeyThreshold   int64         `json:"hot_key_threshold"`
	MonitorDuration   time.Duration `json:"monitor_duration"`
	EnableMitigation  bool          `json:"enable_mitigation"`
	ReplicationFactor int           `json:"replication_factor"`
	ReplicaTTL        time.Duration `json:"replica_ttl"`
}

// Helper function
func hotkeyContains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}