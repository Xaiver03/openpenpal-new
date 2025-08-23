package services

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"openpenpal-backend/internal/models"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// AuditStorageService 增强的审计日志存储服务
type AuditStorageService struct {
	db                *gorm.DB
	redisClient       *redis.Client
	concurrencyMgr    *ConcurrencyManager
	compressionLevel  int
	batchSize         int
	flushInterval     time.Duration
	archiveAfterDays  int
	
	// 内存缓冲区
	buffer           []AuditLogEntry
	bufferMutex      sync.RWMutex
	flushTimer       *time.Timer
	
	// 异步处理
	workerPool       *ConcurrentWorkerPool
	ctx              context.Context
	cancel           context.CancelFunc
	wg               sync.WaitGroup
}

// AuditLogEntry 审计日志条目（增强版）
type AuditLogEntry struct {
	models.AuditLog
	Level        string                 `json:"level"`
	Username     string                 `json:"username,omitempty"`
	SessionID    string                 `json:"session_id,omitempty"`
	TraceID      string                 `json:"trace_id,omitempty"`
	Changes      map[string]interface{} `json:"changes,omitempty"`
	Tags         []string               `json:"tags,omitempty"`
	Compressed   bool                   `json:"compressed"`
	Archived     bool                   `json:"archived"`
}

// AuditStorageConfig 审计存储配置
type AuditStorageConfig struct {
	BatchSize         int           // 批量写入大小
	FlushInterval     time.Duration // 刷新间隔
	CompressionLevel  int           // 压缩级别 (0-9, 0=不压缩)
	ArchiveAfterDays  int           // 多少天后归档
	WorkerCount       int           // 异步工作线程数
	EnableCompression bool          // 是否启用压缩
	EnableArchiving   bool          // 是否启用归档
}

// DefaultAuditStorageConfig 默认配置
func DefaultAuditStorageConfig() AuditStorageConfig {
	return AuditStorageConfig{
		BatchSize:         100,
		FlushInterval:     5 * time.Second,
		CompressionLevel:  6, // gzip默认级别
		ArchiveAfterDays:  30,
		WorkerCount:       3,
		EnableCompression: true,
		EnableArchiving:   true,
	}
}

// NewAuditStorageService 创建审计存储服务
func NewAuditStorageService(db *gorm.DB, redisClient *redis.Client, config AuditStorageConfig) *AuditStorageService {
	ctx, cancel := context.WithCancel(context.Background())
	
	service := &AuditStorageService{
		db:               db,
		redisClient:      redisClient,
		compressionLevel: config.CompressionLevel,
		batchSize:        config.BatchSize,
		flushInterval:    config.FlushInterval,
		archiveAfterDays: config.ArchiveAfterDays,
		buffer:           make([]AuditLogEntry, 0, config.BatchSize),
		ctx:              ctx,
		cancel:           cancel,
	}
	
	// 创建并发管理器
	if redisClient != nil {
		service.concurrencyMgr = NewConcurrencyManager(db, redisClient)
		service.workerPool = service.concurrencyMgr.NewConcurrentWorkerPool(ctx, config.WorkerCount)
	}
	
	// 启动后台任务
	service.startBackgroundTasks()
	
	return service
}

// WriteAuditLog 写入审计日志（异步）
func (s *AuditStorageService) WriteAuditLog(entry AuditLogEntry) error {
	// 生成ID和时间戳
	if entry.ID == "" {
		entry.ID = uuid.New().String()
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now()
	}
	
	// 压缩大型详情数据
	if s.compressionLevel > 0 && len(entry.Details) > 1024 {
		compressed, err := s.compressData(entry.Details)
		if err == nil {
			entry.Details = compressed
			entry.Compressed = true
		}
	}
	
	// 添加到缓冲区
	s.bufferMutex.Lock()
	s.buffer = append(s.buffer, entry)
	shouldFlush := len(s.buffer) >= s.batchSize
	s.bufferMutex.Unlock()
	
	// 如果达到批量大小，立即刷新
	if shouldFlush {
		s.flushBuffer()
	}
	
	// 重置刷新定时器
	s.resetFlushTimer()
	
	// 如果是关键事件，立即写入Redis进行实时监控
	if entry.Level == "critical" || entry.Level == "error" {
		s.writeCriticalEventToRedis(entry)
	}
	
	return nil
}

// flushBuffer 刷新缓冲区到数据库
func (s *AuditStorageService) flushBuffer() {
	s.bufferMutex.Lock()
	if len(s.buffer) == 0 {
		s.bufferMutex.Unlock()
		return
	}
	
	// 复制缓冲区内容
	entries := make([]AuditLogEntry, len(s.buffer))
	copy(entries, s.buffer)
	s.buffer = s.buffer[:0] // 清空缓冲区
	s.bufferMutex.Unlock()
	
	// 异步批量写入
	if s.workerPool != nil {
		s.workerPool.Submit(func() error {
			return s.batchInsertAuditLogs(entries)
		})
	} else {
		// 同步写入（回退方案）
		go s.batchInsertAuditLogs(entries)
	}
}

// batchInsertAuditLogs 批量插入审计日志
func (s *AuditStorageService) batchInsertAuditLogs(entries []AuditLogEntry) error {
	if len(entries) == 0 {
		return nil
	}
	
	// 转换为数据库模型
	logs := make([]models.AuditLog, len(entries))
	for i, entry := range entries {
		logs[i] = entry.AuditLog
	}
	
	// 使用事务批量插入
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 分批插入以避免SQL语句过大
		batchSize := 50
		for i := 0; i < len(logs); i += batchSize {
			end := i + batchSize
			if end > len(logs) {
				end = len(logs)
			}
			
			if err := tx.CreateInBatches(logs[i:end], batchSize).Error; err != nil {
				return fmt.Errorf("failed to batch insert audit logs: %w", err)
			}
		}
		return nil
	})
	
	if err != nil {
		// 记录错误，但不影响后续操作
		fmt.Printf("Error writing audit logs: %v\n", err)
		// 可以考虑将失败的日志写入备份存储
		s.writeFailedLogsToBackup(entries)
	}
	
	return err
}

// compressData 压缩数据
func (s *AuditStorageService) compressData(data string) (string, error) {
	var buf bytes.Buffer
	gz, err := gzip.NewWriterLevel(&buf, s.compressionLevel)
	if err != nil {
		return "", err
	}
	
	if _, err := gz.Write([]byte(data)); err != nil {
		gz.Close()
		return "", err
	}
	
	if err := gz.Close(); err != nil {
		return "", err
	}
	
	// Base64编码压缩后的数据
	return fmt.Sprintf("gzip:%s", buf.String()), nil
}

// decompressData 解压数据
func (s *AuditStorageService) decompressData(data string) (string, error) {
	if len(data) < 5 || data[:5] != "gzip:" {
		return data, nil // 未压缩的数据
	}
	
	compressed := data[5:]
	reader, err := gzip.NewReader(bytes.NewReader([]byte(compressed)))
	if err != nil {
		return "", err
	}
	defer reader.Close()
	
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(reader); err != nil {
		return "", err
	}
	
	return buf.String(), nil
}

// writeCriticalEventToRedis 将关键事件写入Redis
func (s *AuditStorageService) writeCriticalEventToRedis(entry AuditLogEntry) {
	if s.redisClient == nil {
		return
	}
	
	key := fmt.Sprintf("audit:critical:%s", entry.ID)
	data, _ := json.Marshal(entry)
	
	// 设置24小时过期
	s.redisClient.Set(context.Background(), key, data, 24*time.Hour)
	
	// 添加到关键事件列表
	listKey := "audit:critical:list"
	s.redisClient.LPush(context.Background(), listKey, entry.ID)
	s.redisClient.LTrim(context.Background(), listKey, 0, 999) // 保留最新1000条
}

// writeFailedLogsToBackup 将失败的日志写入备份存储
func (s *AuditStorageService) writeFailedLogsToBackup(entries []AuditLogEntry) {
	if s.redisClient == nil {
		return
	}
	
	// 将失败的日志暂存到Redis
	for _, entry := range entries {
		key := fmt.Sprintf("audit:failed:%s", entry.ID)
		data, _ := json.Marshal(entry)
		s.redisClient.Set(context.Background(), key, data, 7*24*time.Hour) // 保存7天
	}
}

// startBackgroundTasks 启动后台任务
func (s *AuditStorageService) startBackgroundTasks() {
	// 定期刷新缓冲区
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		
		ticker := time.NewTicker(s.flushInterval)
		defer ticker.Stop()
		
		for {
			select {
			case <-s.ctx.Done():
				s.flushBuffer() // 退出前刷新剩余数据
				return
			case <-ticker.C:
				s.flushBuffer()
			}
		}
	}()
	
	// 定期归档旧日志
	if s.archiveAfterDays > 0 {
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			
			ticker := time.NewTicker(24 * time.Hour)
			defer ticker.Stop()
			
			// 启动时执行一次
			s.archiveOldLogs()
			
			for {
				select {
				case <-s.ctx.Done():
					return
				case <-ticker.C:
					s.archiveOldLogs()
				}
			}
		}()
	}
	
	// 定期清理Redis中的过期数据
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		
		for {
			select {
			case <-s.ctx.Done():
				return
			case <-ticker.C:
				s.cleanupRedisData()
			}
		}
	}()
}

// resetFlushTimer 重置刷新定时器
func (s *AuditStorageService) resetFlushTimer() {
	if s.flushTimer != nil {
		s.flushTimer.Stop()
	}
	
	s.flushTimer = time.AfterFunc(s.flushInterval, func() {
		s.flushBuffer()
	})
}

// archiveOldLogs 归档旧日志
func (s *AuditStorageService) archiveOldLogs() {
	cutoffDate := time.Now().AddDate(0, 0, -s.archiveAfterDays)
	
	// 创建归档表（如果不存在）
	archiveTableName := fmt.Sprintf("audit_logs_archive_%s", cutoffDate.Format("200601"))
	s.db.Exec(fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (LIKE audit_logs INCLUDING ALL)
	`, archiveTableName))
	
	// 移动旧数据到归档表
	result := s.db.Exec(fmt.Sprintf(`
		INSERT INTO %s 
		SELECT * FROM audit_logs 
		WHERE created_at < ? 
		ON CONFLICT DO NOTHING
	`, archiveTableName), cutoffDate)
	
	if result.Error == nil && result.RowsAffected > 0 {
		// 删除已归档的数据
		s.db.Where("created_at < ?", cutoffDate).Delete(&models.AuditLog{})
		
		fmt.Printf("Archived %d audit logs to %s\n", result.RowsAffected, archiveTableName)
	}
}

// cleanupRedisData 清理Redis中的过期数据
func (s *AuditStorageService) cleanupRedisData() {
	if s.redisClient == nil {
		return
	}
	
	ctx := context.Background()
	
	// 清理失败的日志（超过7天）
	pattern := "audit:failed:*"
	var cursor uint64
	for {
		keys, nextCursor, err := s.redisClient.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			break
		}
		
		for _, key := range keys {
			// 检查TTL，如果没有设置TTL则删除
			ttl, _ := s.redisClient.TTL(ctx, key).Result()
			if ttl == -1 {
				s.redisClient.Del(ctx, key)
			}
		}
		
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
}

// QueryAuditLogs 查询审计日志（支持解压）
func (s *AuditStorageService) QueryAuditLogs(ctx context.Context, filters map[string]interface{}, page, limit int) ([]AuditLogEntry, int64, error) {
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
	
	// 转换并解压数据
	entries := make([]AuditLogEntry, len(logs))
	for i, log := range logs {
		entries[i] = AuditLogEntry{AuditLog: log}
		
		// 解压详情数据
		if len(log.Details) > 5 && log.Details[:5] == "gzip:" {
			decompressed, err := s.decompressData(log.Details)
			if err == nil {
				entries[i].Details = decompressed
				entries[i].Compressed = true
			}
		}
	}
	
	return entries, total, nil
}

// GetRealtimeAlerts 获取实时告警
func (s *AuditStorageService) GetRealtimeAlerts(ctx context.Context) ([]AuditLogEntry, error) {
	if s.redisClient == nil {
		return nil, fmt.Errorf("redis client not available")
	}
	
	// 获取最近的关键事件
	ids, err := s.redisClient.LRange(ctx, "audit:critical:list", 0, 99).Result()
	if err != nil {
		return nil, err
	}
	
	entries := make([]AuditLogEntry, 0, len(ids))
	for _, id := range ids {
		key := fmt.Sprintf("audit:critical:%s", id)
		data, err := s.redisClient.Get(ctx, key).Result()
		if err != nil {
			continue
		}
		
		var entry AuditLogEntry
		if err := json.Unmarshal([]byte(data), &entry); err == nil {
			entries = append(entries, entry)
		}
	}
	
	return entries, nil
}

// GetAuditStatistics 获取审计统计信息
func (s *AuditStorageService) GetAuditStatistics(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// 总记录数
	var totalCount int64
	s.db.Model(&models.AuditLog{}).Count(&totalCount)
	stats["total_count"] = totalCount
	
	// 今日记录数
	var todayCount int64
	today := time.Now().Format("2006-01-02")
	s.db.Model(&models.AuditLog{}).Where("DATE(created_at) = ?", today).Count(&todayCount)
	stats["today_count"] = todayCount
	
	// 缓冲区状态
	s.bufferMutex.RLock()
	bufferSize := len(s.buffer)
	s.bufferMutex.RUnlock()
	stats["buffer_size"] = bufferSize
	stats["buffer_capacity"] = s.batchSize
	
	// 按事件类型统计
	var actionStats []struct {
		Action string
		Count  int64
	}
	s.db.Model(&models.AuditLog{}).
		Select("action, COUNT(*) as count").
		Group("action").
		Order("count DESC").
		Limit(10).
		Scan(&actionStats)
	stats["top_actions"] = actionStats
	
	// Redis中的关键事件数
	if s.redisClient != nil {
		criticalCount, _ := s.redisClient.LLen(ctx, "audit:critical:list").Result()
		stats["critical_events"] = criticalCount
	}
	
	return stats, nil
}

// Close 关闭服务
func (s *AuditStorageService) Close() {
	s.cancel()
	
	// 刷新剩余数据
	s.flushBuffer()
	
	// 关闭工作池
	if s.workerPool != nil {
		s.workerPool.Close()
	}
	
	// 等待所有后台任务完成
	s.wg.Wait()
}