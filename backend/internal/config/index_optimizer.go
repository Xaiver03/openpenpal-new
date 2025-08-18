package config

import (
	"fmt"
	"log"
	"strings"
	"time"

	"gorm.io/gorm"
)

// IndexOptimizer PostgreSQL索引优化器
type IndexOptimizer struct {
	db           *gorm.DB
	dryRun       bool
	verbose      bool
	indexHistory []IndexOperation
}

// IndexOperation 索引操作记录
type IndexOperation struct {
	TableName   string
	IndexName   string
	Operation   string // CREATE, DROP, REINDEX
	SQL         string
	ExecutedAt  time.Time
	Duration    time.Duration
	Success     bool
	Error       error
}

// IndexDefinition 索引定义
type IndexDefinition struct {
	TableName   string
	IndexName   string
	Columns     []string
	Unique      bool
	Concurrent  bool
	Partial     string // WHERE条件
	Include     []string // INCLUDE列(覆盖索引)
	Type        string // btree, hash, gin, gist
	Comment     string
}

// NewIndexOptimizer 创建索引优化器
func NewIndexOptimizer(db *gorm.DB, dryRun bool, verbose bool) *IndexOptimizer {
	return &IndexOptimizer{
		db:           db,
		dryRun:       dryRun,
		verbose:      verbose,
		indexHistory: make([]IndexOperation, 0),
	}
}

// GetCriticalIndexes 获取关键索引定义
func (io *IndexOptimizer) GetCriticalIndexes() []IndexDefinition {
	return []IndexDefinition{
		// === Users表索引 ===
		{
			TableName:  "users",
			IndexName:  "idx_users_school_role_active",
			Columns:    []string{"school_code", "role", "is_active"},
			Partial:    "is_active = true",
			Comment:    "学校用户角色查询优化",
		},
		{
			TableName:  "users",
			IndexName:  "idx_users_created_at_desc",
			Columns:    []string{"created_at DESC"},
			Comment:    "用户列表时间排序优化",
		},
		
		// === Letters表索引 ===
		{
			TableName:  "letters",
			IndexName:  "idx_letters_user_status_created",
			Columns:    []string{"user_id", "status", "created_at DESC"},
			Include:    []string{"title", "style"},
			Comment:    "用户信件列表查询优化(覆盖索引)",
		},
		{
			TableName:  "letters",
			IndexName:  "idx_letters_recipient_status",
			Columns:    []string{"recipient_op_code", "status"},
			Partial:    "status IN ('published', 'delivered')",
			Comment:    "收件人信件查询优化",
		},
		{
			TableName:  "letters",
			IndexName:  "idx_letters_deleted_at",
			Columns:    []string{"deleted_at"},
			Partial:    "deleted_at IS NULL",
			Comment:    "软删除过滤优化",
		},
		
		// === Letter Codes表索引 ===
		{
			TableName:  "letter_codes",
			IndexName:  "idx_letter_codes_code_status",
			Columns:    []string{"code", "status"},
			Include:    []string{"letter_id", "recipient_code"},
			Comment:    "条码查询优化",
		},
		{
			TableName:  "letter_codes",
			IndexName:  "idx_letter_codes_status_created",
			Columns:    []string{"status", "created_at DESC"},
			Partial:    "status != 'delivered'",
			Comment:    "未投递条码查询优化",
		},
		
		// === Courier Tasks表索引 ===
		{
			TableName:  "courier_tasks",
			IndexName:  "idx_courier_tasks_courier_status",
			Columns:    []string{"courier_id", "status", "priority DESC", "created_at"},
			Comment:    "信使任务列表优化",
		},
		{
			TableName:  "courier_tasks",
			IndexName:  "idx_courier_tasks_pickup_delivery",
			Columns:    []string{"pickup_op_code", "delivery_op_code"},
			Partial:    "status NOT IN ('completed', 'cancelled')",
			Comment:    "活跃任务地理查询优化",
		},
		{
			TableName:  "courier_tasks",
			IndexName:  "idx_courier_tasks_assigned_at",
			Columns:    []string{"assigned_at DESC"},
			Partial:    "assigned_at IS NOT NULL",
			Comment:    "已分配任务时间排序",
		},
		
		// === OPCode/SignalCode表索引 ===
		{
			TableName:  "signal_codes",
			IndexName:  "idx_signal_codes_prefix_lookup",
			Columns:    []string{"code"},
			Type:       "btree",
			Comment:    "OP Code前缀查询优化(支持LIKE 'PK%')",
		},
		{
			TableName:  "signal_codes",
			IndexName:  "idx_signal_codes_school_area",
			Columns:    []string{"school_code", "area_code", "status"},
			Comment:    "学校区域码查询优化",
		},
		
		// === Museum表索引 ===
		{
			TableName:  "museum_items",
			IndexName:  "idx_museum_items_featured",
			Columns:    []string{"is_featured", "view_count DESC", "created_at DESC"},
			Partial:    "is_featured = true",
			Comment:    "精选展品查询优化",
		},
		{
			TableName:  "museum_items",
			IndexName:  "idx_museum_items_user_public",
			Columns:    []string{"user_id", "is_public", "created_at DESC"},
			Comment:    "用户公开展品查询",
		},
		
		// === Notifications表索引 ===
		{
			TableName:  "notifications",
			IndexName:  "idx_notifications_user_unread",
			Columns:    []string{"user_id", "is_read", "created_at DESC"},
			Partial:    "is_read = false",
			Comment:    "未读通知查询优化",
		},
		{
			TableName:  "notifications",
			IndexName:  "idx_notifications_type_created",
			Columns:    []string{"type", "created_at DESC"},
			Comment:    "通知类型筛选优化",
		},
		
		// === Analytics表索引 ===
		{
			TableName:  "analytics_metrics",
			IndexName:  "idx_analytics_time_type",
			Columns:    []string{"recorded_at DESC", "metric_type"},
			Comment:    "时序数据查询优化",
		},
		{
			TableName:  "user_analytics",
			IndexName:  "idx_user_analytics_period",
			Columns:    []string{"user_id", "period_start DESC", "period_type"},
			Comment:    "用户统计周期查询",
		},
		
		// === Credit System表索引 ===
		{
			TableName:  "credit_transactions",
			IndexName:  "idx_credit_trans_user_time",
			Columns:    []string{"user_id", "created_at DESC"},
			Include:    []string{"amount", "type", "balance_after"},
			Comment:    "用户积分流水查询优化",
		},
		{
			TableName:  "credit_activities",
			IndexName:  "idx_credit_activities_active",
			Columns:    []string{"status", "start_time", "end_time"},
			Partial:    "status = 'active'",
			Comment:    "活跃积分活动查询",
		},
	}
}

// GetTextSearchIndexes 获取全文搜索索引
func (io *IndexOptimizer) GetTextSearchIndexes() []IndexDefinition {
	return []IndexDefinition{
		{
			TableName:  "letters",
			IndexName:  "idx_letters_fulltext",
			Columns:    []string{"to_tsvector('simple', title || ' ' || content)"},
			Type:       "gin",
			Comment:    "信件全文搜索索引",
		},
		{
			TableName:  "museum_items",
			IndexName:  "idx_museum_items_fulltext",
			Columns:    []string{"to_tsvector('simple', title || ' ' || description)"},
			Type:       "gin",
			Comment:    "展品全文搜索索引",
		},
		{
			TableName:  "comments",
			IndexName:  "idx_comments_fulltext",
			Columns:    []string{"to_tsvector('simple', content)"},
			Type:       "gin",
			Comment:    "评论全文搜索索引",
		},
	}
}

// CreateIndex 创建单个索引
func (io *IndexOptimizer) CreateIndex(def IndexDefinition) error {
	startTime := time.Now()
	
	// 构建CREATE INDEX语句
	sql := io.buildCreateIndexSQL(def)
	
	operation := IndexOperation{
		TableName:  def.TableName,
		IndexName:  def.IndexName,
		Operation:  "CREATE",
		SQL:        sql,
		ExecutedAt: startTime,
	}
	
	if io.verbose {
		log.Printf("Creating index: %s on %s", def.IndexName, def.TableName)
		log.Printf("SQL: %s", sql)
	}
	
	if !io.dryRun {
		// 检查索引是否已存在
		exists, err := io.indexExists(def.TableName, def.IndexName)
		if err != nil {
			operation.Error = err
			io.indexHistory = append(io.indexHistory, operation)
			return fmt.Errorf("failed to check index existence: %w", err)
		}
		
		if exists {
			if io.verbose {
				log.Printf("Index %s already exists, skipping", def.IndexName)
			}
			return nil
		}
		
		// 执行创建索引
		err = io.db.Exec(sql).Error
		operation.Duration = time.Since(startTime)
		operation.Success = err == nil
		operation.Error = err
		
		if err != nil {
			io.indexHistory = append(io.indexHistory, operation)
			return fmt.Errorf("failed to create index %s: %w", def.IndexName, err)
		}
		
		if io.verbose {
			log.Printf("Index %s created successfully in %v", def.IndexName, operation.Duration)
		}
		
		// 添加注释
		if def.Comment != "" {
			commentSQL := fmt.Sprintf("COMMENT ON INDEX %s IS '%s'", def.IndexName, def.Comment)
			io.db.Exec(commentSQL)
		}
	}
	
	operation.Success = true
	operation.Duration = time.Since(startTime)
	io.indexHistory = append(io.indexHistory, operation)
	
	return nil
}

// buildCreateIndexSQL 构建CREATE INDEX SQL语句
func (io *IndexOptimizer) buildCreateIndexSQL(def IndexDefinition) string {
	var parts []string
	
	// CREATE [UNIQUE] INDEX [CONCURRENTLY]
	parts = append(parts, "CREATE")
	if def.Unique {
		parts = append(parts, "UNIQUE")
	}
	parts = append(parts, "INDEX")
	if def.Concurrent {
		parts = append(parts, "CONCURRENTLY")
	}
	
	// IF NOT EXISTS index_name
	parts = append(parts, fmt.Sprintf("IF NOT EXISTS %s", def.IndexName))
	
	// ON table_name
	parts = append(parts, fmt.Sprintf("ON %s", def.TableName))
	
	// USING method
	if def.Type != "" && def.Type != "btree" {
		parts = append(parts, fmt.Sprintf("USING %s", def.Type))
	}
	
	// (columns)
	columnList := strings.Join(def.Columns, ", ")
	parts = append(parts, fmt.Sprintf("(%s)", columnList))
	
	// INCLUDE (columns) - PostgreSQL 11+
	if len(def.Include) > 0 {
		includeList := strings.Join(def.Include, ", ")
		parts = append(parts, fmt.Sprintf("INCLUDE (%s)", includeList))
	}
	
	// WHERE condition
	if def.Partial != "" {
		parts = append(parts, fmt.Sprintf("WHERE %s", def.Partial))
	}
	
	return strings.Join(parts, " ")
}

// indexExists 检查索引是否存在
func (io *IndexOptimizer) indexExists(tableName, indexName string) (bool, error) {
	var count int64
	err := io.db.Raw(`
		SELECT COUNT(*) 
		FROM pg_indexes 
		WHERE tablename = ? AND indexname = ?
	`, tableName, indexName).Scan(&count).Error
	
	return count > 0, err
}

// OptimizeAll 执行所有优化
func (io *IndexOptimizer) OptimizeAll() error {
	log.Println("Starting PostgreSQL index optimization...")
	
	// 阶段1: 创建关键索引
	log.Println("Phase 1: Creating critical indexes...")
	criticalIndexes := io.GetCriticalIndexes()
	for _, idx := range criticalIndexes {
		if err := io.CreateIndex(idx); err != nil {
			log.Printf("Warning: %v", err)
			// 继续处理其他索引
		}
	}
	
	// 阶段2: 创建全文搜索索引
	log.Println("Phase 2: Creating text search indexes...")
	textIndexes := io.GetTextSearchIndexes()
	for _, idx := range textIndexes {
		if err := io.CreateIndex(idx); err != nil {
			log.Printf("Warning: %v", err)
		}
	}
	
	// 阶段3: 分析表统计信息
	if !io.dryRun {
		log.Println("Phase 3: Analyzing table statistics...")
		if err := io.analyzeAllTables(); err != nil {
			log.Printf("Warning: Failed to analyze tables: %v", err)
		}
	}
	
	// 生成报告
	io.generateReport()
	
	return nil
}

// analyzeAllTables 分析所有表的统计信息
func (io *IndexOptimizer) analyzeAllTables() error {
	tables := []string{
		"users", "letters", "letter_codes", "courier_tasks",
		"signal_codes", "museum_items", "notifications",
		"analytics_metrics", "credit_transactions",
	}
	
	for _, table := range tables {
		if io.verbose {
			log.Printf("Analyzing table: %s", table)
		}
		if err := io.db.Exec(fmt.Sprintf("ANALYZE %s", table)).Error; err != nil {
			return fmt.Errorf("failed to analyze %s: %w", table, err)
		}
	}
	
	return nil
}

// generateReport 生成优化报告
func (io *IndexOptimizer) generateReport() {
	log.Println("\n=== Index Optimization Report ===")
	
	successCount := 0
	failureCount := 0
	totalDuration := time.Duration(0)
	
	for _, op := range io.indexHistory {
		if op.Success {
			successCount++
		} else {
			failureCount++
		}
		totalDuration += op.Duration
	}
	
	log.Printf("Total operations: %d", len(io.indexHistory))
	log.Printf("Successful: %d", successCount)
	log.Printf("Failed: %d", failureCount)
	log.Printf("Total duration: %v", totalDuration)
	
	if failureCount > 0 {
		log.Println("\nFailed operations:")
		for _, op := range io.indexHistory {
			if !op.Success {
				log.Printf("- %s on %s: %v", op.IndexName, op.TableName, op.Error)
			}
		}
	}
	
	if io.dryRun {
		log.Println("\n[DRY RUN] No changes were made to the database")
	}
}

// GetIndexStats 获取索引统计信息
func (io *IndexOptimizer) GetIndexStats() ([]IndexStats, error) {
	var stats []IndexStats
	
	query := `
		SELECT 
			schemaname,
			tablename,
			indexname,
			pg_size_pretty(pg_relation_size(schemaname||'.'||indexname)) as index_size,
			idx_scan as scan_count,
			idx_tup_read as tuples_read,
			idx_tup_fetch as tuples_fetched
		FROM pg_stat_user_indexes
		WHERE schemaname = 'public'
		ORDER BY pg_relation_size(schemaname||'.'||indexname) DESC
	`
	
	err := io.db.Raw(query).Scan(&stats).Error
	return stats, err
}

// IndexStats 索引统计信息
type IndexStats struct {
	SchemaName    string `json:"schema_name"`
	TableName     string `json:"table_name"`
	IndexName     string `json:"index_name"`
	IndexSize     string `json:"index_size"`
	ScanCount     int64  `json:"scan_count"`
	TuplesRead    int64  `json:"tuples_read"`
	TuplesFetched int64  `json:"tuples_fetched"`
}

// GetUnusedIndexes 获取未使用的索引
func (io *IndexOptimizer) GetUnusedIndexes(days int) ([]string, error) {
	var unusedIndexes []string
	
	query := `
		SELECT 
			schemaname || '.' || indexname as full_index_name
		FROM pg_stat_user_indexes
		WHERE schemaname = 'public'
			AND idx_scan = 0
			AND indexrelname NOT LIKE '%_pkey'
			AND pg_relation_size(indexrelid) > 1000000 -- 只考虑大于1MB的索引
	`
	
	err := io.db.Raw(query).Pluck("full_index_name", &unusedIndexes).Error
	return unusedIndexes, err
}

// ReindexTable 重建表索引
func (io *IndexOptimizer) ReindexTable(tableName string) error {
	if io.verbose {
		log.Printf("Reindexing table: %s", tableName)
	}
	
	sql := fmt.Sprintf("REINDEX TABLE CONCURRENTLY %s", tableName)
	
	if !io.dryRun {
		return io.db.Exec(sql).Error
	}
	
	return nil
}

// DropIndex 删除索引
func (io *IndexOptimizer) DropIndex(indexName string, concurrent bool) error {
	sql := "DROP INDEX"
	if concurrent {
		sql += " CONCURRENTLY"
	}
	sql += fmt.Sprintf(" IF EXISTS %s", indexName)
	
	if io.verbose {
		log.Printf("Dropping index: %s", indexName)
		log.Printf("SQL: %s", sql)
	}
	
	if !io.dryRun {
		return io.db.Exec(sql).Error
	}
	
	return nil
}