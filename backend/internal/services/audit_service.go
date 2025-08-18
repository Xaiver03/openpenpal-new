package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"openpenpal-backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuditService 审计日志服务
type AuditService struct {
	db *gorm.DB
}

// AuditEventType 审计事件类型
type AuditEventType string

const (
	// 用户相关事件
	AuditEventUserLogin         AuditEventType = "user_login"
	AuditEventUserLogout        AuditEventType = "user_logout"
	AuditEventUserRegister      AuditEventType = "user_register"
	AuditEventUserUpdate        AuditEventType = "user_update"
	AuditEventUserDelete        AuditEventType = "user_delete"
	AuditEventPasswordChange    AuditEventType = "password_change"
	AuditEventPasswordReset     AuditEventType = "password_reset"
	
	// 权限相关事件
	AuditEventRoleChange        AuditEventType = "role_change"
	AuditEventPermissionGrant   AuditEventType = "permission_grant"
	AuditEventPermissionRevoke  AuditEventType = "permission_revoke"
	
	// 内容相关事件
	AuditEventLetterCreate      AuditEventType = "letter_create"
	AuditEventLetterUpdate      AuditEventType = "letter_update"
	AuditEventLetterDelete      AuditEventType = "letter_delete"
	AuditEventLetterPublish     AuditEventType = "letter_publish"
	AuditEventCommentCreate     AuditEventType = "comment_create"
	AuditEventCommentDelete     AuditEventType = "comment_delete"
	AuditEventCommentModerate   AuditEventType = "comment_moderate"
	
	// 安全相关事件
	AuditEventSensitiveWordAdd  AuditEventType = "sensitive_word_add"
	AuditEventSensitiveWordDel  AuditEventType = "sensitive_word_delete"
	AuditEventSecurityViolation AuditEventType = "security_violation"
	AuditEventAccessDenied      AuditEventType = "access_denied"
	AuditEventSuspiciousActivity AuditEventType = "suspicious_activity"
	
	// 系统相关事件
	AuditEventSystemConfig      AuditEventType = "system_config"
	AuditEventBackup            AuditEventType = "system_backup"
	AuditEventRestore           AuditEventType = "system_restore"
	AuditEventIntegrityCheck    AuditEventType = "integrity_check"
	AuditEventDataRepair        AuditEventType = "data_repair"
	
	// 批量操作事件
	AuditEventBatchOperation    AuditEventType = "batch_operation"
	AuditEventDataExport        AuditEventType = "data_export"
	AuditEventDataImport        AuditEventType = "data_import"
)

// AuditLevel 审计级别
type AuditLevel string

const (
	AuditLevelInfo     AuditLevel = "info"
	AuditLevelWarning  AuditLevel = "warning"
	AuditLevelError    AuditLevel = "error"
	AuditLevelCritical AuditLevel = "critical"
)

// AuditEntry 审计条目
type AuditEntry struct {
	ID          string                 `json:"id"`
	UserID      string                 `json:"user_id"`
	Username    string                 `json:"username"`
	EventType   AuditEventType         `json:"event_type"`
	Level       AuditLevel             `json:"level"`
	Resource    string                 `json:"resource"`
	ResourceID  string                 `json:"resource_id"`
	Action      string                 `json:"action"`
	Result      string                 `json:"result"`
	Details     map[string]interface{} `json:"details"`
	Changes     map[string]interface{} `json:"changes,omitempty"`
	IP          string                 `json:"ip"`
	UserAgent   string                 `json:"user_agent"`
	SessionID   string                 `json:"session_id,omitempty"`
	TraceID     string                 `json:"trace_id,omitempty"`
	Duration    float64                `json:"duration,omitempty"`
	Error       string                 `json:"error,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
}

// NewAuditService 创建审计服务
func NewAuditService(db *gorm.DB) *AuditService {
	return &AuditService{
		db: db,
	}
}

// LogEvent 记录审计事件
func (s *AuditService) LogEvent(ctx context.Context, entry *AuditEntry) error {
	if entry.ID == "" {
		entry.ID = uuid.New().String()
	}
	
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now()
	}

	// 将审计条目转换为数据库模型
	auditLog := &models.AuditLog{
		ID:         entry.ID,
		UserID:     entry.UserID,
		Action:     string(entry.EventType),
		Resource:   entry.Resource,
		ResourceID: entry.ResourceID,
		IP:         entry.IP,
		UserAgent:  entry.UserAgent,
		CreatedAt:  entry.CreatedAt,
	}

	// 序列化详情
	if entry.Details != nil {
		detailsJSON, err := json.Marshal(entry.Details)
		if err != nil {
			return fmt.Errorf("failed to serialize audit details: %w", err)
		}
		auditLog.Details = string(detailsJSON)
	}

	// 保存到数据库
	if err := s.db.WithContext(ctx).Create(auditLog).Error; err != nil {
		return fmt.Errorf("failed to save audit log: %w", err)
	}

	// 如果是关键事件，可能需要额外的处理（如发送告警）
	if entry.Level == AuditLevelCritical {
		s.handleCriticalEvent(ctx, entry)
	}

	return nil
}

// LogUserAction 记录用户操作
func (s *AuditService) LogUserAction(ctx context.Context, userID string, eventType AuditEventType, resource string, resourceID string, details map[string]interface{}) error {
	// 从上下文获取请求信息
	var ip, userAgent string
	if ginCtx, ok := ctx.Value("gin_context").(*gin.Context); ok {
		ip = ginCtx.ClientIP()
		userAgent = ginCtx.Request.UserAgent()
	}

	entry := &AuditEntry{
		UserID:     userID,
		EventType:  eventType,
		Level:      AuditLevelInfo,
		Resource:   resource,
		ResourceID: resourceID,
		Action:     string(eventType),
		Result:     "success",
		Details:    details,
		IP:         ip,
		UserAgent:  userAgent,
	}

	return s.LogEvent(ctx, entry)
}

// LogSecurityEvent 记录安全事件
func (s *AuditService) LogSecurityEvent(ctx context.Context, eventType AuditEventType, level AuditLevel, details map[string]interface{}) error {
	// 从上下文获取请求信息
	var ip, userAgent, userID string
	if ginCtx, ok := ctx.Value("gin_context").(*gin.Context); ok {
		ip = ginCtx.ClientIP()
		userAgent = ginCtx.Request.UserAgent()
		if user, exists := ginCtx.Get("user"); exists {
			if u, ok := user.(*models.User); ok {
				userID = u.ID
			}
		}
	}

	entry := &AuditEntry{
		UserID:     userID,
		EventType:  eventType,
		Level:      level,
		Resource:   "security",
		ResourceID: "",
		Action:     string(eventType),
		Result:     "logged",
		Details:    details,
		IP:         ip,
		UserAgent:  userAgent,
	}

	return s.LogEvent(ctx, entry)
}

// LogDataChange 记录数据变更
func (s *AuditService) LogDataChange(ctx context.Context, userID string, resource string, resourceID string, oldData, newData interface{}) error {
	// 计算变更
	changes := s.calculateChanges(oldData, newData)
	
	details := map[string]interface{}{
		"resource":    resource,
		"resource_id": resourceID,
		"changes":     changes,
	}

	return s.LogUserAction(ctx, userID, AuditEventType(resource+"_update"), resource, resourceID, details)
}

// QueryAuditLogs 查询审计日志
func (s *AuditService) QueryAuditLogs(ctx context.Context, filters map[string]interface{}, page, limit int) ([]AuditEntry, int64, error) {
	query := s.db.Model(&models.AuditLog{})

	// 应用过滤器
	if userID, ok := filters["user_id"].(string); ok && userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	if action, ok := filters["action"].(string); ok && action != "" {
		query = query.Where("action = ?", action)
	}

	if resource, ok := filters["resource"].(string); ok && resource != "" {
		query = query.Where("resource = ?", resource)
	}

	if resourceID, ok := filters["resource_id"].(string); ok && resourceID != "" {
		query = query.Where("resource_id = ?", resourceID)
	}

	// 时间范围过滤
	if startTime, ok := filters["start_time"].(time.Time); ok && !startTime.IsZero() {
		query = query.Where("created_at >= ?", startTime)
	}

	if endTime, ok := filters["end_time"].(time.Time); ok && !endTime.IsZero() {
		query = query.Where("created_at <= ?", endTime)
	}

	// 计算总数
	var total int64
	query.Count(&total)

	// 分页查询
	offset := (page - 1) * limit
	var logs []models.AuditLog
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&logs).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to query audit logs: %w", err)
	}

	// 转换为审计条目
	entries := make([]AuditEntry, len(logs))
	for i, log := range logs {
		entries[i] = s.convertToAuditEntry(log)
	}

	return entries, total, nil
}

// GetUserActivityReport 获取用户活动报告
func (s *AuditService) GetUserActivityReport(ctx context.Context, userID string, startTime, endTime time.Time) (map[string]interface{}, error) {
	report := make(map[string]interface{})

	// 统计各类操作次数
	var stats []struct {
		Action string
		Count  int64
	}

	s.db.Model(&models.AuditLog{}).
		Select("action, COUNT(*) as count").
		Where("user_id = ? AND created_at BETWEEN ? AND ?", userID, startTime, endTime).
		Group("action").
		Scan(&stats)

	actionCounts := make(map[string]int64)
	for _, stat := range stats {
		actionCounts[stat.Action] = stat.Count
	}

	report["action_counts"] = actionCounts

	// 获取最近的活动
	var recentLogs []models.AuditLog
	s.db.Where("user_id = ? AND created_at BETWEEN ? AND ?", userID, startTime, endTime).
		Order("created_at DESC").
		Limit(10).
		Find(&recentLogs)

	recentActivities := make([]AuditEntry, len(recentLogs))
	for i, log := range recentLogs {
		recentActivities[i] = s.convertToAuditEntry(log)
	}

	report["recent_activities"] = recentActivities

	// 统计活跃时段
	var hourlyStats []struct {
		Hour  int
		Count int64
	}

	s.db.Model(&models.AuditLog{}).
		Select("EXTRACT(HOUR FROM created_at) as hour, COUNT(*) as count").
		Where("user_id = ? AND created_at BETWEEN ? AND ?", userID, startTime, endTime).
		Group("hour").
		Scan(&hourlyStats)

	hourlyActivity := make(map[int]int64)
	for _, stat := range hourlyStats {
		hourlyActivity[stat.Hour] = stat.Count
	}

	report["hourly_activity"] = hourlyActivity

	return report, nil
}

// GetSecurityEventReport 获取安全事件报告
func (s *AuditService) GetSecurityEventReport(ctx context.Context, startTime, endTime time.Time) (map[string]interface{}, error) {
	report := make(map[string]interface{})

	// 安全事件类型
	securityEvents := []string{
		string(AuditEventSecurityViolation),
		string(AuditEventAccessDenied),
		string(AuditEventSuspiciousActivity),
	}

	// 统计安全事件
	var securityStats []struct {
		Action string
		Count  int64
	}

	s.db.Model(&models.AuditLog{}).
		Select("action, COUNT(*) as count").
		Where("action IN ? AND created_at BETWEEN ? AND ?", securityEvents, startTime, endTime).
		Group("action").
		Scan(&securityStats)

	eventCounts := make(map[string]int64)
	for _, stat := range securityStats {
		eventCounts[stat.Action] = stat.Count
	}

	report["security_event_counts"] = eventCounts

	// 获取高风险事件
	var criticalEvents []models.AuditLog
	s.db.Where("action IN ? AND created_at BETWEEN ? AND ?", securityEvents, startTime, endTime).
		Order("created_at DESC").
		Limit(20).
		Find(&criticalEvents)

	criticalEntries := make([]AuditEntry, len(criticalEvents))
	for i, log := range criticalEvents {
		criticalEntries[i] = s.convertToAuditEntry(log)
	}

	report["critical_events"] = criticalEntries

	// 统计IP地址
	var ipStats []struct {
		IP    string
		Count int64
	}

	s.db.Model(&models.AuditLog{}).
		Select("ip, COUNT(*) as count").
		Where("action IN ? AND created_at BETWEEN ? AND ?", securityEvents, startTime, endTime).
		Group("ip").
		Order("count DESC").
		Limit(10).
		Scan(&ipStats)

	report["top_ips"] = ipStats

	return report, nil
}

// CleanupOldLogs 清理旧日志
func (s *AuditService) CleanupOldLogs(ctx context.Context, retentionDays int) (int64, error) {
	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)
	
	result := s.db.WithContext(ctx).
		Where("created_at < ?", cutoffDate).
		Delete(&models.AuditLog{})
	
	if result.Error != nil {
		return 0, fmt.Errorf("failed to cleanup old logs: %w", result.Error)
	}

	return result.RowsAffected, nil
}

// calculateChanges 计算数据变更
func (s *AuditService) calculateChanges(oldData, newData interface{}) map[string]interface{} {
	changes := make(map[string]interface{})
	
	// 将数据转换为map进行比较
	oldMap := s.structToMap(oldData)
	newMap := s.structToMap(newData)
	
	// 比较字段变化
	for key, newValue := range newMap {
		if oldValue, exists := oldMap[key]; exists {
			if oldValue != newValue {
				changes[key] = map[string]interface{}{
					"old": oldValue,
					"new": newValue,
				}
			}
		} else {
			changes[key] = map[string]interface{}{
				"old": nil,
				"new": newValue,
			}
		}
	}
	
	// 检查删除的字段
	for key, oldValue := range oldMap {
		if _, exists := newMap[key]; !exists {
			changes[key] = map[string]interface{}{
				"old": oldValue,
				"new": nil,
			}
		}
	}
	
	return changes
}

// structToMap 将结构体转换为map
func (s *AuditService) structToMap(data interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	
	// 简化实现，实际应该使用反射或JSON序列化
	jsonData, err := json.Marshal(data)
	if err != nil {
		return result
	}
	
	json.Unmarshal(jsonData, &result)
	return result
}

// convertToAuditEntry 将数据库模型转换为审计条目
func (s *AuditService) convertToAuditEntry(log models.AuditLog) AuditEntry {
	entry := AuditEntry{
		ID:         log.ID,
		UserID:     log.UserID,
		EventType:  AuditEventType(log.Action),
		Level:      AuditLevelInfo,
		Resource:   log.Resource,
		ResourceID: log.ResourceID,
		Action:     log.Action,
		Result:     "success",
		IP:         log.IP,
		UserAgent:  log.UserAgent,
		CreatedAt:  log.CreatedAt,
	}

	// 解析详情
	if log.Details != "" {
		var details map[string]interface{}
		if err := json.Unmarshal([]byte(log.Details), &details); err == nil {
			entry.Details = details
		}
	}

	// 从详情中提取其他字段
	if entry.Details != nil {
		if username, ok := entry.Details["username"].(string); ok {
			entry.Username = username
		}
		if level, ok := entry.Details["level"].(string); ok {
			entry.Level = AuditLevel(level)
		}
		if result, ok := entry.Details["result"].(string); ok {
			entry.Result = result
		}
		if errorMsg, ok := entry.Details["error"].(string); ok {
			entry.Error = errorMsg
		}
	}

	return entry
}

// handleCriticalEvent 处理关键事件
func (s *AuditService) handleCriticalEvent(ctx context.Context, entry *AuditEntry) {
	// 这里可以实现：
	// 1. 发送告警邮件
	// 2. 记录到特殊的告警表
	// 3. 触发自动响应机制
	// 4. 通知管理员
	
	// 暂时只记录日志
	fmt.Printf("CRITICAL SECURITY EVENT: %+v\n", entry)
}

// GetAuditStats 获取审计统计
func (s *AuditService) GetAuditStats(period string) (map[string]interface{}, error) {
	// 根据时间段计算开始时间
	var startTime time.Time
	now := time.Now()
	
	switch period {
	case "1d":
		startTime = now.Add(-24 * time.Hour)
	case "7d":
		startTime = now.Add(-7 * 24 * time.Hour)
	case "30d":
		startTime = now.Add(-30 * 24 * time.Hour)
	default:
		startTime = now.Add(-7 * 24 * time.Hour) // 默认7天
	}

	// 统计各类事件数量
	var eventStats []struct {
		Action string `json:"action"`
		Count  int64  `json:"count"`
	}
	if err := s.db.Model(&models.AuditLog{}).
		Select("action, COUNT(*) as count").
		Where("created_at >= ?", startTime).
		Group("action").
		Find(&eventStats).Error; err != nil {
		return nil, fmt.Errorf("failed to get event stats: %w", err)
	}

	// 统计活跃用户
	var activeUsers int64
	if err := s.db.Model(&models.AuditLog{}).
		Where("created_at >= ?", startTime).
		Distinct("user_id").
		Count(&activeUsers).Error; err != nil {
		return nil, fmt.Errorf("failed to count active users: %w", err)
	}

	// 统计总事件数
	var totalEvents int64
	if err := s.db.Model(&models.AuditLog{}).
		Where("created_at >= ?", startTime).
		Count(&totalEvents).Error; err != nil {
		return nil, fmt.Errorf("failed to count total events: %w", err)
	}

	stats := map[string]interface{}{
		"period":        period,
		"start_time":    startTime,
		"end_time":      now,
		"total_events":  totalEvents,
		"active_users":  activeUsers,
		"event_stats":   eventStats,
	}

	return stats, nil
}

// ExportAuditLogs 导出审计日志 (修正版)
func (s *AuditService) ExportAuditLogs(eventType string, startTime, endTime *time.Time, format string) ([]byte, string, error) {
	filters := map[string]interface{}{}
	if eventType != "" {
		filters["action"] = eventType
	}
	if startTime != nil {
		filters["start_time"] = *startTime
	}
	if endTime != nil {
		filters["end_time"] = *endTime
	}
	
	// 查询日志
	entries, _, err := s.QueryAuditLogs(context.Background(), filters, 1, 10000) // 最多导出10000条
	if err != nil {
		return nil, "", err
	}

	var data []byte
	var contentType string
	
	switch format {
	case "json":
		data, err = json.MarshalIndent(entries, "", "  ")
		contentType = "application/json"
	case "csv":
		data, err = s.exportToCSV(entries)
		contentType = "text/csv"
	default:
		return nil, "", fmt.Errorf("unsupported export format: %s", format)
	}
	
	if err != nil {
		return nil, "", err
	}
	
	return data, contentType, nil
}


// exportToCSV 导出为CSV格式
func (s *AuditService) exportToCSV(entries []AuditEntry) ([]byte, error) {
	// 简化的CSV导出实现
	csv := "ID,UserID,Username,EventType,Level,Resource,ResourceID,Action,Result,IP,UserAgent,CreatedAt\n"
	
	for _, entry := range entries {
		csv += fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s\n",
			entry.ID,
			entry.UserID,
			entry.Username,
			entry.EventType,
			entry.Level,
			entry.Resource,
			entry.ResourceID,
			entry.Action,
			entry.Result,
			entry.IP,
			entry.UserAgent,
			entry.CreatedAt.Format(time.RFC3339),
		)
	}
	
	return []byte(csv), nil
}