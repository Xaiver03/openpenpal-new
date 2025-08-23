package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"openpenpal-backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// EnhancedAuditService 增强的审计服务
type EnhancedAuditService struct {
	db            *gorm.DB
	redisClient   *redis.Client
	storageSvc    *AuditStorageService
	notificationSvc *NotificationService
	
	// 审计策略配置
	enableSensitiveDataMasking bool
	enableRealTimeAlerts       bool
	criticalEventThreshold     int
}

// NewEnhancedAuditService 创建增强的审计服务
func NewEnhancedAuditService(db *gorm.DB, redisClient *redis.Client) *EnhancedAuditService {
	// 创建存储服务
	storageConfig := DefaultAuditStorageConfig()
	storageSvc := NewAuditStorageService(db, redisClient, storageConfig)
	
	return &EnhancedAuditService{
		db:                        db,
		redisClient:              redisClient,
		storageSvc:               storageSvc,
		enableSensitiveDataMasking: true,
		enableRealTimeAlerts:      true,
		criticalEventThreshold:    10, // 10分钟内超过10次关键事件触发告警
	}
}

// SetNotificationService 设置通知服务
func (s *EnhancedAuditService) SetNotificationService(notificationSvc *NotificationService) {
	s.notificationSvc = notificationSvc
}

// LogEvent 记录审计事件（增强版）
func (s *EnhancedAuditService) LogEvent(ctx context.Context, entry *AuditEntry) error {
	// 生成追踪ID
	if entry.TraceID == "" {
		if traceID, ok := ctx.Value("trace_id").(string); ok {
			entry.TraceID = traceID
		} else {
			entry.TraceID = uuid.New().String()
		}
	}
	
	// 提取会话信息
	if ginCtx, ok := ctx.Value("gin_context").(*gin.Context); ok {
		if sessionID, exists := ginCtx.Get("session_id"); exists {
			entry.SessionID = fmt.Sprintf("%v", sessionID)
		}
	}
	
	// 敏感数据脱敏
	if s.enableSensitiveDataMasking {
		entry.Details = s.maskSensitiveData(entry.Details)
		entry.Changes = s.maskSensitiveData(entry.Changes)
	}
	
	// 转换为存储格式
	storageEntry := AuditLogEntry{
		AuditLog: models.AuditLog{
			ID:         entry.ID,
			UserID:     entry.UserID,
			Action:     string(entry.EventType),
			Resource:   entry.Resource,
			ResourceID: entry.ResourceID,
			IP:         entry.IP,
			UserAgent:  entry.UserAgent,
			Result:     entry.Result,
			Error:      entry.Error,
			Duration:   entry.Duration,
			CreatedAt:  entry.CreatedAt,
		},
		Level:     string(entry.Level),
		Username:  entry.Username,
		SessionID: entry.SessionID,
		TraceID:   entry.TraceID,
		Changes:   entry.Changes,
	}
	
	// 序列化详情
	if entry.Details != nil {
		detailsJSON, err := json.Marshal(entry.Details)
		if err != nil {
			return fmt.Errorf("failed to serialize audit details: %w", err)
		}
		storageEntry.Details = string(detailsJSON)
	}
	
	// 异步写入存储
	if err := s.storageSvc.WriteAuditLog(storageEntry); err != nil {
		return fmt.Errorf("failed to write audit log: %w", err)
	}
	
	// 处理关键事件
	if entry.Level == AuditLevelCritical {
		s.handleCriticalEvent(ctx, entry)
	}
	
	// 实时分析
	if s.enableRealTimeAlerts {
		s.analyzeEventPattern(ctx, entry)
	}
	
	return nil
}

// LogUserAction 记录用户操作（增强版）
func (s *EnhancedAuditService) LogUserAction(ctx context.Context, userID string, eventType AuditEventType, resource string, resourceID string, details map[string]interface{}) error {
	// 获取用户信息
	var username string
	var user models.User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err == nil {
		username = user.Username
	}
	
	// 从上下文获取请求信息
	var ip, userAgent string
	if ginCtx, ok := ctx.Value("gin_context").(*gin.Context); ok {
		ip = ginCtx.ClientIP()
		userAgent = ginCtx.Request.UserAgent()
	}
	
	// 计算操作耗时
	startTime := time.Now()
	if start, ok := ctx.Value("start_time").(time.Time); ok {
		duration := time.Since(start).Seconds()
		details["duration_ms"] = int(duration * 1000)
	}
	
	entry := &AuditEntry{
		UserID:     userID,
		Username:   username,
		EventType:  eventType,
		Level:      AuditLevelInfo,
		Resource:   resource,
		ResourceID: resourceID,
		Action:     string(eventType),
		Result:     "success",
		Details:    details,
		IP:         ip,
		UserAgent:  userAgent,
		Duration:   time.Since(startTime).Seconds(),
	}
	
	return s.LogEvent(ctx, entry)
}

// LogSecurityEvent 记录安全事件（增强版）
func (s *EnhancedAuditService) LogSecurityEvent(ctx context.Context, eventType AuditEventType, level AuditLevel, details map[string]interface{}) error {
	// 增加安全事件的额外信息
	details["security_event"] = true
	details["timestamp"] = time.Now().Unix()
	
	// 获取请求指纹
	if ginCtx, ok := ctx.Value("gin_context").(*gin.Context); ok {
		details["request_method"] = ginCtx.Request.Method
		details["request_path"] = ginCtx.Request.URL.Path
		details["request_headers"] = s.getSecurityHeaders(ginCtx)
	}
	
	// 从上下文获取信息
	var ip, userAgent, userID, username string
	if ginCtx, ok := ctx.Value("gin_context").(*gin.Context); ok {
		ip = ginCtx.ClientIP()
		userAgent = ginCtx.Request.UserAgent()
		if user, exists := ginCtx.Get("user"); exists {
			if u, ok := user.(*models.User); ok {
				userID = u.ID
				username = u.Username
			}
		}
	}
	
	entry := &AuditEntry{
		UserID:     userID,
		Username:   username,
		EventType:  eventType,
		Level:      level,
		Resource:   "security",
		ResourceID: "",
		Action:     string(eventType),
		Result:     "logged",
		Details:    details,
		IP:         ip,
		UserAgent:  userAgent,
		CreatedAt:  time.Now(),
	}
	
	return s.LogEvent(ctx, entry)
}

// LogDataChange 记录数据变更
func (s *EnhancedAuditService) LogDataChange(ctx context.Context, userID string, resource string, resourceID string, oldData, newData interface{}) error {
	// 计算变更内容
	changes := s.calculateChanges(oldData, newData)
	
	details := map[string]interface{}{
		"operation": "update",
		"changes_count": len(changes),
	}
	
	entry := &AuditEntry{
		UserID:     userID,
		EventType:  AuditEventType(fmt.Sprintf("%s_update", resource)),
		Level:      AuditLevelInfo,
		Resource:   resource,
		ResourceID: resourceID,
		Action:     "update",
		Result:     "success",
		Details:    details,
		Changes:    changes,
		CreatedAt:  time.Now(),
	}
	
	// 从上下文获取请求信息
	if ginCtx, ok := ctx.Value("gin_context").(*gin.Context); ok {
		entry.IP = ginCtx.ClientIP()
		entry.UserAgent = ginCtx.Request.UserAgent()
	}
	
	return s.LogEvent(ctx, entry)
}

// maskSensitiveData 脱敏敏感数据
func (s *EnhancedAuditService) maskSensitiveData(data map[string]interface{}) map[string]interface{} {
	if data == nil {
		return nil
	}
	
	masked := make(map[string]interface{})
	sensitiveFields := []string{
		"password", "token", "secret", "api_key", "private_key",
		"credit_card", "ssn", "id_card", "phone", "email",
	}
	
	for key, value := range data {
		isSensitive := false
		for _, field := range sensitiveFields {
			if key == field || auditContains(key, field) {
				isSensitive = true
				break
			}
		}
		
		if isSensitive {
			masked[key] = "***MASKED***"
		} else {
			// 递归处理嵌套的map
			if nestedMap, ok := value.(map[string]interface{}); ok {
				masked[key] = s.maskSensitiveData(nestedMap)
			} else {
				masked[key] = value
			}
		}
	}
	
	return masked
}

// calculateChanges 计算数据变更
func (s *EnhancedAuditService) calculateChanges(oldData, newData interface{}) map[string]interface{} {
	changes := make(map[string]interface{})
	
	// 简化实现，实际应该使用反射或专门的diff库
	oldJSON, _ := json.Marshal(oldData)
	newJSON, _ := json.Marshal(newData)
	
	var oldMap, newMap map[string]interface{}
	json.Unmarshal(oldJSON, &oldMap)
	json.Unmarshal(newJSON, &newMap)
	
	for key, newValue := range newMap {
		if oldValue, exists := oldMap[key]; exists {
			if fmt.Sprintf("%v", oldValue) != fmt.Sprintf("%v", newValue) {
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

// getSecurityHeaders 获取安全相关的请求头
func (s *EnhancedAuditService) getSecurityHeaders(ctx *gin.Context) map[string]string {
	headers := make(map[string]string)
	securityHeaders := []string{
		"X-Forwarded-For",
		"X-Real-IP",
		"X-Request-ID",
		"Referer",
		"Origin",
		"Authorization",
	}
	
	for _, header := range securityHeaders {
		if value := ctx.GetHeader(header); value != "" {
			if header == "Authorization" {
				// 脱敏授权头
				headers[header] = "Bearer ***"
			} else {
				headers[header] = value
			}
		}
	}
	
	return headers
}

// handleCriticalEvent 处理关键事件
func (s *EnhancedAuditService) handleCriticalEvent(ctx context.Context, entry *AuditEntry) {
	// 发送实时告警
	if s.notificationSvc != nil {
		// TODO: Implement SendAlert method or use NotifyUser
		// s.notificationSvc.SendAlert(ctx, "critical_audit_event", map[string]interface{}{
		// 	"event_type": entry.EventType,
		// 	"user_id":    entry.UserID,
		// 	"resource":   entry.Resource,
		// 	"details":    entry.Details,
		// 	"timestamp":  entry.CreatedAt,
		// })
	}
	
	// 检查是否需要触发紧急响应
	s.checkEmergencyResponse(ctx, entry)
}

// analyzeEventPattern 分析事件模式
func (s *EnhancedAuditService) analyzeEventPattern(ctx context.Context, entry *AuditEntry) {
	if s.redisClient == nil {
		return
	}
	
	// 使用滑动窗口统计事件频率
	key := fmt.Sprintf("audit:pattern:%s:%s", entry.UserID, entry.EventType)
	now := time.Now()
	
	// 添加当前事件
	s.redisClient.ZAdd(ctx, key, redis.Z{
		Score:  float64(now.Unix()),
		Member: entry.ID,
	})
	
	// 清理10分钟前的数据
	cutoff := now.Add(-10 * time.Minute).Unix()
	s.redisClient.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", cutoff))
	
	// 统计10分钟内的事件数
	count, _ := s.redisClient.ZCard(ctx, key).Result()
	
	// 检查是否超过阈值
	if int(count) > s.criticalEventThreshold {
		// 触发告警
		s.triggerPatternAlert(ctx, entry, int(count))
	}
	
	// 设置过期时间
	s.redisClient.Expire(ctx, key, 1*time.Hour)
}

// checkEmergencyResponse 检查是否需要紧急响应
func (s *EnhancedAuditService) checkEmergencyResponse(ctx context.Context, entry *AuditEntry) {
	// 定义需要紧急响应的事件
	emergencyEvents := []AuditEventType{
		AuditEventSecurityViolation,
		AuditEventSuspiciousActivity,
		AuditEventDataExport, // 大量数据导出
	}
	
	for _, eventType := range emergencyEvents {
		if entry.EventType == eventType {
			// 可以在这里实现紧急响应逻辑
			// 例如：锁定账户、通知管理员、启动调查流程等
			fmt.Printf("EMERGENCY: %s event detected for user %s\n", eventType, entry.UserID)
			break
		}
	}
}

// triggerPatternAlert 触发模式告警
func (s *EnhancedAuditService) triggerPatternAlert(ctx context.Context, entry *AuditEntry, count int) {
	alertDetails := map[string]interface{}{
		"alert_type": "high_frequency_event",
		"user_id":    entry.UserID,
		"event_type": entry.EventType,
		"count":      count,
		"threshold":  s.criticalEventThreshold,
		"window":     "10_minutes",
		"timestamp":  time.Now(),
	}
	
	// 记录告警事件
	alertEntry := &AuditEntry{
		UserID:     "system",
		EventType:  AuditEventType("pattern_alert"),
		Level:      AuditLevelWarning,
		Resource:   "audit_pattern",
		ResourceID: entry.UserID,
		Action:     "alert_triggered",
		Result:     "notified",
		Details:    alertDetails,
		CreatedAt:  time.Now(),
	}
	
	// 避免递归，直接写入存储
	s.storageSvc.WriteAuditLog(AuditLogEntry{
		AuditLog: models.AuditLog{
			ID:         uuid.New().String(),
			UserID:     alertEntry.UserID,
			Action:     string(alertEntry.EventType),
			Resource:   alertEntry.Resource,
			ResourceID: alertEntry.ResourceID,
			Result:     alertEntry.Result,
			CreatedAt:  alertEntry.CreatedAt,
		},
		Level: string(alertEntry.Level),
	})
	
	// 发送通知
	if s.notificationSvc != nil {
		// TODO: Implement SendAlert method or use NotifyUser
		// s.notificationSvc.SendAlert(ctx, "pattern_alert", alertDetails)
	}
}

// QueryAuditLogs 查询审计日志
func (s *EnhancedAuditService) QueryAuditLogs(ctx context.Context, filters map[string]interface{}, page, limit int) ([]AuditEntry, int64, error) {
	// 使用存储服务查询
	logs, total, err := s.storageSvc.QueryAuditLogs(ctx, filters, page, limit)
	if err != nil {
		return nil, 0, err
	}
	
	// 转换为审计条目
	entries := make([]AuditEntry, len(logs))
	for i, log := range logs {
		// 解析详情
		var details map[string]interface{}
		if log.Details != "" {
			json.Unmarshal([]byte(log.Details), &details)
		}
		
		entries[i] = AuditEntry{
			ID:         log.ID,
			UserID:     log.UserID,
			Username:   log.Username,
			EventType:  AuditEventType(log.Action),
			Level:      AuditLevel(log.Level),
			Resource:   log.Resource,
			ResourceID: log.ResourceID,
			Action:     log.Action,
			Result:     log.Result,
			Details:    details,
			IP:         log.IP,
			UserAgent:  log.UserAgent,
			SessionID:  log.SessionID,
			TraceID:    log.TraceID,
			Duration:   log.Duration,
			Error:      log.Error,
			CreatedAt:  log.CreatedAt,
		}
	}
	
	return entries, total, nil
}

// GetRealtimeAlerts 获取实时告警
func (s *EnhancedAuditService) GetRealtimeAlerts(ctx context.Context) ([]AuditEntry, error) {
	logs, err := s.storageSvc.GetRealtimeAlerts(ctx)
	if err != nil {
		return nil, err
	}
	
	entries := make([]AuditEntry, len(logs))
	for i, log := range logs {
		var details map[string]interface{}
		if log.Details != "" {
			json.Unmarshal([]byte(log.Details), &details)
		}
		
		entries[i] = AuditEntry{
			ID:         log.ID,
			UserID:     log.UserID,
			EventType:  AuditEventType(log.Action),
			Level:      AuditLevel(log.Level),
			Resource:   log.Resource,
			ResourceID: log.ResourceID,
			Details:    details,
			CreatedAt:  log.CreatedAt,
		}
	}
	
	return entries, nil
}

// GetAuditStatistics 获取审计统计
func (s *EnhancedAuditService) GetAuditStatistics(ctx context.Context) (map[string]interface{}, error) {
	return s.storageSvc.GetAuditStatistics(ctx)
}

// Close 关闭服务
func (s *EnhancedAuditService) Close() {
	if s.storageSvc != nil {
		s.storageSvc.Close()
	}
}

// auditContains 辅助函数
func auditContains(s, substr string) bool {
	return len(s) >= len(substr) && auditFindSubstring(s, substr)
}

func auditFindSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}