package middleware

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// SecurityEvent 安全事件
type SecurityEvent struct {
	Type        string                 `json:"type"`
	Severity    string                 `json:"severity"`
	Source      string                 `json:"source"`
	UserAgent   string                 `json:"user_agent"`
	IPAddress   string                 `json:"ip_address"`
	URL         string                 `json:"url"`
	Method      string                 `json:"method"`
	Details     map[string]interface{} `json:"details"`
	Timestamp   time.Time              `json:"timestamp"`
	RequestID   string                 `json:"request_id"`
}

// SecurityMonitor 安全监控器
type SecurityMonitor struct {
	events    []SecurityEvent
	mu        sync.RWMutex
	maxEvents int
}

// globalSecurityMonitor 全局安全监控实例
var globalSecurityMonitor = &SecurityMonitor{
	events:    make([]SecurityEvent, 0),
	maxEvents: 1000, // 最多保存1000个事件
}

// RecordSecurityEvent 记录安全事件
func RecordSecurityEvent(c *gin.Context, eventType, severity string, details map[string]interface{}) {
	event := SecurityEvent{
		Type:      eventType,
		Severity:  severity,
		Source:    "middleware",
		UserAgent: c.GetHeader("User-Agent"),
		IPAddress: c.ClientIP(),
		URL:       c.Request.URL.String(),
		Method:    c.Request.Method,
		Details:   details,
		Timestamp: time.Now(),
		RequestID: c.GetString("request_id"),
	}
	
	globalSecurityMonitor.AddEvent(event)
	
	// 根据严重性级别决定是否立即记录日志
	switch severity {
	case "critical", "high":
		logSecurityEvent(event)
	case "medium":
		if eventType == "rate_limit_exceeded" || eventType == "xss_attempt" {
			logSecurityEvent(event)
		}
	}
}

// AddEvent 添加安全事件
func (sm *SecurityMonitor) AddEvent(event SecurityEvent) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	// 如果超过最大容量，移除最老的事件
	if len(sm.events) >= sm.maxEvents {
		sm.events = sm.events[1:]
	}
	
	sm.events = append(sm.events, event)
}

// GetRecentEvents 获取最近的安全事件
func (sm *SecurityMonitor) GetRecentEvents(limit int) []SecurityEvent {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	if limit <= 0 || limit > len(sm.events) {
		limit = len(sm.events)
	}
	
	// 返回最近的事件（倒序）
	events := make([]SecurityEvent, limit)
	for i := 0; i < limit; i++ {
		events[i] = sm.events[len(sm.events)-1-i]
	}
	
	return events
}

// GetEventsByType 根据类型获取事件
func (sm *SecurityMonitor) GetEventsByType(eventType string, limit int) []SecurityEvent {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	var filtered []SecurityEvent
	for i := len(sm.events) - 1; i >= 0 && len(filtered) < limit; i-- {
		if sm.events[i].Type == eventType {
			filtered = append(filtered, sm.events[i])
		}
	}
	
	return filtered
}

// GetStats 获取安全统计信息
func (sm *SecurityMonitor) GetStats() map[string]interface{} {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	stats := map[string]interface{}{
		"total_events": len(sm.events),
		"event_types":  make(map[string]int),
		"severity":     make(map[string]int),
		"recent_24h":   0,
	}
	
	now := time.Now()
	for _, event := range sm.events {
		// 统计事件类型
		stats["event_types"].(map[string]int)[event.Type]++
		
		// 统计严重性级别
		stats["severity"].(map[string]int)[event.Severity]++
		
		// 统计最近24小时的事件
		if now.Sub(event.Timestamp) <= 24*time.Hour {
			stats["recent_24h"] = stats["recent_24h"].(int) + 1
		}
	}
	
	return stats
}

// logSecurityEvent 记录安全事件到日志
func logSecurityEvent(event SecurityEvent) {
	eventJSON, _ := json.Marshal(event)
	log.Printf("[SECURITY_EVENT] %s", string(eventJSON))
}

// RequestMonitoringMiddleware 请求监控中间件
func RequestMonitoringMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录请求开始
		start := time.Now()
		
		// 继续处理请求
		c.Next()
		
		// 检查响应状态码
		statusCode := c.Writer.Status()
		duration := time.Since(start)
		
		// 记录可疑的活动
		if statusCode == 429 {
			// 速率限制触发
			RecordSecurityEvent(c, "rate_limit_exceeded", "medium", map[string]interface{}{
				"status_code": statusCode,
				"duration_ms": duration.Milliseconds(),
			})
		} else if statusCode >= 400 && statusCode < 500 {
			// 客户端错误
			severity := "low"
			if statusCode == 401 || statusCode == 403 {
				severity = "medium"
			}
			
			RecordSecurityEvent(c, "client_error", severity, map[string]interface{}{
				"status_code": statusCode,
				"duration_ms": duration.Milliseconds(),
			})
		} else if statusCode >= 500 {
			// 服务器错误
			RecordSecurityEvent(c, "server_error", "high", map[string]interface{}{
				"status_code": statusCode,
				"duration_ms": duration.Milliseconds(),
			})
		}
		
		// 检查异常慢的请求
		if duration > 10*time.Second {
			RecordSecurityEvent(c, "slow_request", "medium", map[string]interface{}{
				"duration_ms": duration.Milliseconds(),
				"status_code": statusCode,
			})
		}
	}
}

// GetSecurityMonitor 获取全局安全监控实例
func GetSecurityMonitor() *SecurityMonitor {
	return globalSecurityMonitor
}

// SecurityMonitoringMiddleware 安全监控中间件 - 专门用于威胁检测和监控
func SecurityMonitoringMiddleware() gin.HandlerFunc {
	// 可疑User-Agent模式
	suspiciousUserAgents := []string{
		"sqlmap", "nikto", "nmap", "masscan", "zap",
		"burp", "dirbuster", "dirb", "gobuster",
		"python-requests", "curl", "wget",
	}
	
	// 可疑路径模式
	suspiciousPaths := []string{
		"/.env", "/config", "/admin", "/phpmyadmin",
		"/wp-admin", "/api/v1/users", "/debug",
		"/../", "/./", "%2e%2e", "%2f",
		"<script", "javascript:", "data:",
	}
	
	return func(c *gin.Context) {
		userAgent := c.GetHeader("User-Agent")
		path := c.Request.URL.Path
		
		// 检测可疑User-Agent
		for _, suspicious := range suspiciousUserAgents {
			if userAgent == suspicious {
				RecordSecurityEvent(c, "suspicious_user_agent", "medium", map[string]interface{}{
					"user_agent": userAgent,
					"pattern":    suspicious,
				})
				break
			}
		}
		
		// 检测可疑路径
		for _, suspicious := range suspiciousPaths {
			if path == suspicious {
				RecordSecurityEvent(c, "suspicious_path", "medium", map[string]interface{}{
					"path":    path,
					"pattern": suspicious,
				})
				break
			}
		}
		
		// 检测SQL注入尝试
		query := c.Request.URL.RawQuery
		if query != "" {
			sqlPatterns := []string{
				"union", "select", "insert", "update", "delete",
				"drop", "exec", "script", "'", "\"", ";",
				"--", "/*", "*/", "xp_", "sp_",
			}
			
			for _, pattern := range sqlPatterns {
				if query == pattern {
					RecordSecurityEvent(c, "sql_injection_attempt", "high", map[string]interface{}{
						"query":   query,
						"pattern": pattern,
					})
					break
				}
			}
		}
		
		c.Next()
	}
}