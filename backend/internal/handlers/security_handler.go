package handlers

import (
	"net/http"
	"strconv"

	"openpenpal-backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

// SecurityHandler 安全处理器
type SecurityHandler struct {
	monitor *middleware.SecurityMonitor
}

// NewSecurityHandler 创建安全处理器
func NewSecurityHandler() *SecurityHandler {
	return &SecurityHandler{
		monitor: middleware.GetSecurityMonitor(),
	}
}

// GetSecurityEvents 获取安全事件
func (h *SecurityHandler) GetSecurityEvents(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}
	
	eventType := c.Query("type")
	
	var events []middleware.SecurityEvent
	if eventType != "" {
		events = h.monitor.GetEventsByType(eventType, limit)
	} else {
		events = h.monitor.GetRecentEvents(limit)
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"events": events,
			"total":  len(events),
		},
	})
}

// GetSecurityStats 获取安全统计
func (h *SecurityHandler) GetSecurityStats(c *gin.Context) {
	stats := h.monitor.GetStats()
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// GetSecurityDashboard 获取安全仪表板数据
func (h *SecurityHandler) GetSecurityDashboard(c *gin.Context) {
	stats := h.monitor.GetStats()
	recentEvents := h.monitor.GetRecentEvents(10)
	
	// 获取威胁类型分布
	threatTypes := h.monitor.GetEventsByType("xss_attempt", 5)
	rateLimitEvents := h.monitor.GetEventsByType("rate_limit_exceeded", 5)
	
	dashboard := gin.H{
		"overview": stats,
		"recent_events": recentEvents,
		"threat_analysis": gin.H{
			"xss_attempts":     len(threatTypes),
			"rate_limit_hits":  len(rateLimitEvents),
			"total_threats":    stats["recent_24h"],
		},
		"security_status": gin.H{
			"level": getSecurityLevel(stats),
			"recommendations": getSecurityRecommendations(stats),
		},
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    dashboard,
	})
}

// getSecurityLevel 根据统计数据确定安全级别
func getSecurityLevel(stats map[string]interface{}) string {
	recent24h := stats["recent_24h"].(int)
	severity := stats["severity"].(map[string]int)
	
	criticalEvents := severity["critical"]
	highEvents := severity["high"]
	
	if criticalEvents > 0 || highEvents > 10 {
		return "critical"
	} else if highEvents > 0 || recent24h > 100 {
		return "high"
	} else if recent24h > 50 {
		return "medium"
	}
	
	return "low"
}

// getSecurityRecommendations 根据统计数据生成安全建议
func getSecurityRecommendations(stats map[string]interface{}) []string {
	var recommendations []string
	
	eventTypes := stats["event_types"].(map[string]int)
	severity := stats["severity"].(map[string]int)
	
	if eventTypes["rate_limit_exceeded"] > 20 {
		recommendations = append(recommendations, "考虑调整速率限制配置，可能存在DDoS攻击")
	}
	
	if eventTypes["xss_attempt"] > 0 {
		recommendations = append(recommendations, "检测到XSS攻击尝试，建议加强输入验证")
	}
	
	if eventTypes["sql_injection_attempt"] > 0 {
		recommendations = append(recommendations, "检测到SQL注入尝试，建议审查数据库查询")
	}
	
	if severity["critical"] > 0 {
		recommendations = append(recommendations, "存在严重安全事件，需要立即处理")
	}
	
	if eventTypes["suspicious_user_agent"] > 10 {
		recommendations = append(recommendations, "检测到大量可疑User-Agent，可能存在扫描器活动")
	}
	
	if len(recommendations) == 0 {
		recommendations = append(recommendations, "系统安全状态良好，继续保持监控")
	}
	
	return recommendations
}

// RecordCustomSecurityEvent 记录自定义安全事件
func (h *SecurityHandler) RecordCustomSecurityEvent(c *gin.Context) {
	var req struct {
		Type     string                 `json:"type" binding:"required"`
		Severity string                 `json:"severity" binding:"required"`
		Details  map[string]interface{} `json:"details"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}
	
	// 验证严重性级别
	validSeverities := map[string]bool{
		"low": true, "medium": true, "high": true, "critical": true,
	}
	
	if !validSeverities[req.Severity] {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid severity level. Must be one of: low, medium, high, critical",
		})
		return
	}
	
	// 记录安全事件
	middleware.RecordSecurityEvent(c, req.Type, req.Severity, req.Details)
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Security event recorded successfully",
	})
}