package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"regexp"
	"sync"
	"time"
)

// SmartLogger 智能日志管理器，防止重复错误导致的日志膨胀
type SmartLogger struct {
	// 错误聚合器：基于错误模式进行聚合
	errorAggregator map[string]*ErrorEntry
	mu              sync.RWMutex
	
	// 配置
	config *SmartLoggerConfig
	
	// 统计
	stats *LoggerStats
	
	// 时间窗口清理
	lastCleanup time.Time
}

type SmartLoggerConfig struct {
	// 时间窗口：在此窗口内的相同错误会被聚合
	TimeWindow time.Duration `json:"time_window"`
	
	// 最大聚合数：单个错误模式最多聚合多少次
	MaxAggregation int `json:"max_aggregation"`
	
	// 详细阈值：错误重复次数超过此值时切换为简洁模式
	VerboseThreshold int `json:"verbose_threshold"`
	
	// 断路器阈值：错误重复次数超过此值时停止记录
	CircuitBreakerThreshold int `json:"circuit_breaker_threshold"`
	
	// 采样率：高频错误的采样率（1/N）
	SamplingRate int `json:"sampling_rate"`
	
	// 清理间隔：多久清理一次聚合器
	CleanupInterval time.Duration `json:"cleanup_interval"`
}

type ErrorEntry struct {
	// 首次出现时间
	FirstSeen time.Time `json:"first_seen"`
	
	// 最后出现时间
	LastSeen time.Time `json:"last_seen"`
	
	// 出现次数
	Count int `json:"count"`
	
	// 错误消息
	Message string `json:"message"`
	
	// 错误模式哈希
	PatternHash string `json:"pattern_hash"`
	
	// 状态
	Status ErrorStatus `json:"status"`
	
	// 详细信息（仅保留第一次和最后几次）
	Details []ErrorDetail `json:"details"`
}

type ErrorDetail struct {
	Timestamp time.Time              `json:"timestamp"`
	Context   map[string]interface{} `json:"context"`
	StackTrace string                `json:"stack_trace,omitempty"`
}

type ErrorStatus string

const (
	StatusActive    ErrorStatus = "active"     // 正常记录
	StatusSilenced  ErrorStatus = "silenced"   // 静默模式（只记录计数）
	StatusCircuit   ErrorStatus = "circuit"    // 断路器模式（停止记录）
)

type LoggerStats struct {
	TotalErrors       int64 `json:"total_errors"`
	AggregatedErrors  int64 `json:"aggregated_errors"`
	SilencedErrors    int64 `json:"silenced_errors"`
	CircuitedErrors   int64 `json:"circuited_errors"`
	LogReduction      float64 `json:"log_reduction_percent"`
}

// NewSmartLogger 创建智能日志管理器
func NewSmartLogger(config *SmartLoggerConfig) *SmartLogger {
	if config == nil {
		config = &SmartLoggerConfig{
			TimeWindow:              5 * time.Minute,
			MaxAggregation:          1000,
			VerboseThreshold:        10,
			CircuitBreakerThreshold: 100,
			SamplingRate:            10,
			CleanupInterval:         1 * time.Hour,
		}
	}
	
	return &SmartLogger{
		errorAggregator: make(map[string]*ErrorEntry),
		config:          config,
		stats:           &LoggerStats{},
		lastCleanup:     time.Now(),
	}
}

// LogError 智能错误记录
func (sl *SmartLogger) LogError(message string, context map[string]interface{}) {
	sl.mu.Lock()
	defer sl.mu.Unlock()
	
	// 生成错误模式哈希
	patternHash := sl.generatePatternHash(message, context)
	
	now := time.Now()
	sl.stats.TotalErrors++
	
	// 获取或创建错误条目
	entry, exists := sl.errorAggregator[patternHash]
	if !exists {
		entry = &ErrorEntry{
			FirstSeen:   now,
			LastSeen:    now,
			Count:       1,
			Message:     sl.sanitizeMessage(message),
			PatternHash: patternHash,
			Status:      StatusActive,
			Details:     []ErrorDetail{{Timestamp: now, Context: context}},
		}
		sl.errorAggregator[patternHash] = entry
		
		// 首次出现，正常记录
		log.Printf("[SMART_LOG] NEW_ERROR: %s", message)
		return
	}
	
	// 更新现有条目
	entry.LastSeen = now
	entry.Count++
	
	// 根据状态决定记录策略
	switch entry.Status {
	case StatusActive:
		sl.handleActiveError(entry, message, context)
	case StatusSilenced:
		sl.handleSilencedError(entry, message, context)
	case StatusCircuit:
		sl.stats.CircuitedErrors++
		// 断路器模式：完全不记录
		return
	}
	
	// 定期清理
	if now.Sub(sl.lastCleanup) > sl.config.CleanupInterval {
		sl.cleanup()
	}
}

// handleActiveError 处理活跃状态的错误
func (sl *SmartLogger) handleActiveError(entry *ErrorEntry, message string, context map[string]interface{}) {
	if entry.Count <= sl.config.VerboseThreshold {
		// 详细模式：正常记录
		log.Printf("[SMART_LOG] ERROR[%d]: %s", entry.Count, message)
		
		// 保留详细信息
		if len(entry.Details) < 5 {
			entry.Details = append(entry.Details, ErrorDetail{
				Timestamp: time.Now(),
				Context:   context,
			})
		}
	} else if entry.Count <= sl.config.CircuitBreakerThreshold {
		// 切换到静默模式
		entry.Status = StatusSilenced
		log.Printf("[SMART_LOG] SILENCED: Error pattern repeated %d times, switching to count mode: %s", 
			entry.Count, sl.sanitizeMessage(entry.Message))
		sl.stats.SilencedErrors++
	} else {
		// 切换到断路器模式
		entry.Status = StatusCircuit
		log.Printf("[SMART_LOG] CIRCUIT: Error pattern exceeded threshold (%d), stopping logs: %s", 
			entry.Count, sl.sanitizeMessage(entry.Message))
	}
}

// handleSilencedError 处理静默状态的错误
func (sl *SmartLogger) handleSilencedError(entry *ErrorEntry, message string, context map[string]interface{}) {
	sl.stats.AggregatedErrors++
	
	// 采样记录：每N次记录一次
	if entry.Count%sl.config.SamplingRate == 0 {
		log.Printf("[SMART_LOG] SAMPLED[%d]: %s", entry.Count, sl.sanitizeMessage(entry.Message))
	}
	
	// 检查是否需要切换到断路器模式
	if entry.Count > sl.config.CircuitBreakerThreshold {
		entry.Status = StatusCircuit
		log.Printf("[SMART_LOG] CIRCUIT: Error pattern exceeded threshold (%d): %s", 
			entry.Count, sl.sanitizeMessage(entry.Message))
	}
}

// generatePatternHash 生成错误模式哈希
func (sl *SmartLogger) generatePatternHash(message string, context map[string]interface{}) string {
	// 提取错误模式，忽略变化的部分（如时间戳、ID等）
	pattern := sl.extractErrorPattern(message)
	
	// 添加关键上下文
	contextKey := ""
	if context != nil {
		if userID, ok := context["user_id"]; ok {
			contextKey += fmt.Sprintf("user:%v,", userID)
		}
		if taskType, ok := context["task_type"]; ok {
			contextKey += fmt.Sprintf("type:%v,", taskType)
		}
	}
	
	hash := md5.Sum([]byte(pattern + contextKey))
	return hex.EncodeToString(hash[:])[:12] // 使用12位哈希
}

// extractErrorPattern 提取错误模式（移除变化的部分）
func (sl *SmartLogger) extractErrorPattern(message string) string {
	// 移除UUID、时间戳、数字ID等变化部分
	pattern := message
	
	// 移除UUID模式
	pattern = regexp.MustCompile(`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`).ReplaceAllString(pattern, "[UUID]")
	
	// 移除时间戳
	pattern = regexp.MustCompile(`\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}`).ReplaceAllString(pattern, "[TIMESTAMP]")
	
	// 移除数字ID
	pattern = regexp.MustCompile(`\bid:\s*\d+`).ReplaceAllString(pattern, "id:[ID]")
	
	// 移除路径中的具体值
	pattern = regexp.MustCompile(`/\d+/`).ReplaceAllString(pattern, "/[ID]/")
	
	return pattern
}

// sanitizeMessage 清理消息中的敏感信息
func (sl *SmartLogger) sanitizeMessage(message string) string {
	// 只保留前200个字符避免日志过长
	if len(message) > 200 {
		return message[:200] + "..."
	}
	return message
}

// cleanup 清理过期的错误条目
func (sl *SmartLogger) cleanup() {
	now := time.Now()
	cutoff := now.Add(-sl.config.TimeWindow * 2) // 保留2个窗口的数据
	
	for hash, entry := range sl.errorAggregator {
		if entry.LastSeen.Before(cutoff) {
			delete(sl.errorAggregator, hash)
		}
	}
	
	sl.lastCleanup = now
}

// GetStats 获取统计信息
func (sl *SmartLogger) GetStats() *LoggerStats {
	sl.mu.RLock()
	defer sl.mu.RUnlock()
	
	stats := *sl.stats
	if stats.TotalErrors > 0 {
		saved := stats.AggregatedErrors + stats.CircuitedErrors
		stats.LogReduction = float64(saved) / float64(stats.TotalErrors) * 100
	}
	
	return &stats
}

// GetErrorSummary 获取错误摘要
func (sl *SmartLogger) GetErrorSummary() map[string]*ErrorEntry {
	sl.mu.RLock()
	defer sl.mu.RUnlock()
	
	summary := make(map[string]*ErrorEntry)
	for hash, entry := range sl.errorAggregator {
		// 只返回当前时间窗口内的错误
		if time.Since(entry.LastSeen) <= sl.config.TimeWindow {
			summary[hash] = entry
		}
	}
	
	return summary
}

// LogInfo 记录信息日志（简单透传）
func (sl *SmartLogger) LogInfo(message string) {
	log.Printf("[INFO] %s", message)
}

// LogWarning 记录警告日志
func (sl *SmartLogger) LogWarning(message string, context map[string]interface{}) {
	// 警告也可以应用聚合策略，但阈值更高
	patternHash := sl.generatePatternHash(message, context)
	
	sl.mu.Lock()
	entry, exists := sl.errorAggregator[patternHash]
	if !exists || entry.Count < 5 {
		log.Printf("[WARNING] %s", message)
	} else if entry.Count%20 == 0 {
		log.Printf("[WARNING] [SAMPLED:%d] %s", entry.Count, sl.sanitizeMessage(message))
	}
	sl.mu.Unlock()
	
	sl.LogError("WARNING: "+message, context)
}

// PrintSummary 打印错误摘要
func (sl *SmartLogger) PrintSummary() {
	sl.mu.RLock()
	defer sl.mu.RUnlock()
	
	log.Println("=== SMART LOGGER SUMMARY ===")
	log.Printf("Total Errors: %d", sl.stats.TotalErrors)
	log.Printf("Log Reduction: %.1f%%", sl.GetStats().LogReduction)
	
	for hash, entry := range sl.errorAggregator {
		if entry.Count > 1 {
			log.Printf("[%s] %s: %d times (Status: %s, First: %s, Last: %s)",
				hash[:8], 
				sl.sanitizeMessage(entry.Message),
				entry.Count,
				entry.Status,
				entry.FirstSeen.Format("15:04:05"),
				entry.LastSeen.Format("15:04:05"))
		}
	}
	log.Println("=== END SUMMARY ===")
}

