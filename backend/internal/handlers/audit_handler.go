package handlers

import (
	"net/http"
	"strconv"
	"time"

	"openpenpal-backend/internal/middleware"
	"openpenpal-backend/internal/services"

	"github.com/gin-gonic/gin"
)

// AuditHandler 审计日志处理器
type AuditHandler struct {
	auditService *services.AuditService
}

// NewAuditHandler 创建审计处理器
func NewAuditHandler(auditService *services.AuditService) *AuditHandler {
	return &AuditHandler{
		auditService: auditService,
	}
}

// GetAuditLogs 获取审计日志
func (h *AuditHandler) GetAuditLogs(c *gin.Context) {
	// 验证管理员权限
	_, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Unauthorized",
		})
		return
	}

	role, _ := middleware.GetUserRole(c)
	if role != "admin" && role != "super_admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "Insufficient permissions",
		})
		return
	}

	// 解析查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	eventType := c.Query("event_type")
	userFilter := c.Query("user_id")
	
	// 时间范围
	var startTime, endTime *time.Time
	if start := c.Query("start_time"); start != "" {
		t, err := time.Parse(time.RFC3339, start)
		if err == nil {
			startTime = &t
		}
	}
	if end := c.Query("end_time"); end != "" {
		t, err := time.Parse(time.RFC3339, end)
		if err == nil {
			endTime = &t
		}
	}

	// 构建过滤器
	filters := map[string]interface{}{}
	if eventType != "" {
		filters["action"] = eventType
	}
	if userFilter != "" {
		filters["user_id"] = userFilter
	}
	if startTime != nil {
		filters["start_time"] = *startTime
	}
	if endTime != nil {
		filters["end_time"] = *endTime
	}

	// 查询审计日志
	entries, total, err := h.auditService.QueryAuditLogs(c.Request.Context(), filters, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to query audit logs",
			"error":   err.Error(),
		})
		return
	}
	
	// 转换为日志格式（兼容前端期望的格式）
	logs := make([]gin.H, len(entries))
	for i, entry := range entries {
		logs[i] = gin.H{
			"id":          entry.ID,
			"user_id":     entry.UserID,
			"username":    entry.Username,
			"action":      string(entry.EventType),
			"resource":    entry.Resource,
			"resource_id": entry.ResourceID,
			"details":     entry.Details,
			"ip":          entry.IP,
			"user_agent":  entry.UserAgent,
			"result":      entry.Result,
			"error":       entry.Error,
			"duration":    entry.Duration,
			"created_at":  entry.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"logs":  logs,
			"total": total,
			"page":  page,
			"limit": limit,
		},
	})
}

// GetAuditStats 获取审计统计
func (h *AuditHandler) GetAuditStats(c *gin.Context) {
	// 验证管理员权限
	role, _ := middleware.GetUserRole(c)
	if role != "admin" && role != "super_admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "Insufficient permissions",
		})
		return
	}

	// 时间范围
	period := c.DefaultQuery("period", "7d")
	stats, err := h.auditService.GetAuditStats(period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get audit stats",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": stats,
	})
}

// ExportAuditLogs 导出审计日志
func (h *AuditHandler) ExportAuditLogs(c *gin.Context) {
	// 验证管理员权限
	role, _ := middleware.GetUserRole(c)
	if role != "admin" && role != "super_admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "Insufficient permissions",
		})
		return
	}

	// 解析参数
	format := c.DefaultQuery("format", "csv")
	eventType := c.Query("event_type")
	
	// 时间范围
	var startTime, endTime *time.Time
	if start := c.Query("start_time"); start != "" {
		t, err := time.Parse(time.RFC3339, start)
		if err == nil {
			startTime = &t
		}
	}
	if end := c.Query("end_time"); end != "" {
		t, err := time.Parse(time.RFC3339, end)
		if err == nil {
			endTime = &t
		}
	}

	// 导出审计日志
	data, contentType, err := h.auditService.ExportAuditLogs(
		eventType,
		startTime,
		endTime,
		format,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to export audit logs",
			"error":   err.Error(),
		})
		return
	}

	// 设置响应头
	filename := "audit_logs_" + time.Now().Format("20060102_150405") + "." + format
	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, contentType, data)
}