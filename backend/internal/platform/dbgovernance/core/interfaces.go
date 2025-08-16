// Package core provides core interfaces for the database governance system
package core

import (
	"context"
	"database/sql"
	"time"
)

// DatabaseType represents the type of database
type DatabaseType string

const (
	DatabaseTypePostgreSQL DatabaseType = "postgresql"
	DatabaseTypeMySQL      DatabaseType = "mysql"
	DatabaseTypeSQLite     DatabaseType = "sqlite"
)

// ConnectionPoolManager manages database connection pools
type ConnectionPoolManager interface {
	// GetConnection gets a database connection from the pool
	GetConnection(ctx context.Context, dbName string) (*sql.DB, error)
	
	// ReleaseConnection releases a connection back to the pool
	ReleaseConnection(conn *sql.DB) error
	
	// GetPoolStats returns statistics about the connection pool
	GetPoolStats(dbName string) (*PoolStats, error)
	
	// ResizePool dynamically resizes the connection pool
	ResizePool(dbName string, minSize, maxSize int) error
	
	// HealthCheck checks the health of all connections
	HealthCheck(ctx context.Context) (map[string]*ConnectionHealth, error)
}

// QueryAnalyzer analyzes query performance
type QueryAnalyzer interface {
	// AnalyzeQuery analyzes a query and returns optimization suggestions
	AnalyzeQuery(ctx context.Context, query string) (*QueryAnalysis, error)
	
	// GetSlowQueries returns slow queries from the log
	GetSlowQueries(ctx context.Context, threshold time.Duration) ([]*SlowQuery, error)
	
	// GetIndexRecommendations returns index recommendations
	GetIndexRecommendations(ctx context.Context, table string) ([]*IndexRecommendation, error)
	
	// EnableQueryCache enables query result caching
	EnableQueryCache(pattern string, ttl time.Duration) error
}

// MigrationManager manages database migrations
type MigrationManager interface {
	// ApplyMigration applies a migration
	ApplyMigration(ctx context.Context, migration *Migration) error
	
	// RollbackMigration rollbacks a migration
	RollbackMigration(ctx context.Context, version string) error
	
	// GetMigrationHistory returns migration history
	GetMigrationHistory(ctx context.Context) ([]*MigrationHistory, error)
	
	// PlanMigration plans a zero-downtime migration
	PlanMigration(ctx context.Context, migration *Migration) (*MigrationPlan, error)
}

// SecurityManager manages database security
type SecurityManager interface {
	// EncryptData encrypts sensitive data
	EncryptData(data []byte) ([]byte, error)
	
	// DecryptData decrypts encrypted data
	DecryptData(encrypted []byte) ([]byte, error)
	
	// ApplyRowLevelSecurity applies row-level security policies
	ApplyRowLevelSecurity(ctx context.Context, table string, policy *SecurityPolicy) error
	
	// AuditLog logs database operations for audit
	AuditLog(ctx context.Context, operation *AuditOperation) error
	
	// DetectPII detects and masks PII in query results
	DetectPII(ctx context.Context, data interface{}) (*PIIDetectionResult, error)
}

// BackupManager manages database backups
type BackupManager interface {
	// CreateBackup creates a database backup
	CreateBackup(ctx context.Context, config *BackupRequest) (*BackupResult, error)
	
	// RestoreBackup restores from a backup
	RestoreBackup(ctx context.Context, backupID string, targetTime *time.Time) error
	
	// ListBackups lists available backups
	ListBackups(ctx context.Context) ([]*BackupInfo, error)
	
	// ScheduleBackup schedules automated backups
	ScheduleBackup(config *BackupSchedule) error
	
	// TestRecovery tests disaster recovery procedures
	TestRecovery(ctx context.Context) (*RecoveryTestResult, error)
}

// MonitoringManager manages database monitoring
type MonitoringManager interface {
	// CollectMetrics collects database metrics
	CollectMetrics(ctx context.Context) (*DatabaseMetrics, error)
	
	// DetectAnomalies detects anomalies in database behavior
	DetectAnomalies(ctx context.Context) ([]*Anomaly, error)
	
	// CreateAlert creates a monitoring alert
	CreateAlert(alert *Alert) error
	
	// GetDashboard returns monitoring dashboard data
	GetDashboard(ctx context.Context) (*DashboardData, error)
}

// Data Models

// PoolStats represents connection pool statistics
type PoolStats struct {
	ActiveConnections   int     `json:"active_connections"`
	IdleConnections     int     `json:"idle_connections"`
	TotalConnections    int     `json:"total_connections"`
	WaitingRequests     int     `json:"waiting_requests"`
	AverageWaitTime     float64 `json:"average_wait_time"`
	ConnectionsCreated  int64   `json:"connections_created"`
	ConnectionsClosed   int64   `json:"connections_closed"`
	FailedConnections   int64   `json:"failed_connections"`
}

// ConnectionHealth represents health status of a connection
type ConnectionHealth struct {
	DatabaseName string        `json:"database_name"`
	Status       string        `json:"status"`
	Latency      time.Duration `json:"latency"`
	LastChecked  time.Time     `json:"last_checked"`
	ErrorCount   int           `json:"error_count"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// QueryAnalysis represents query analysis results
type QueryAnalysis struct {
	QueryID          string                 `json:"query_id"`
	Query            string                 `json:"query"`
	ExecutionPlan    string                 `json:"execution_plan"`
	EstimatedCost    float64                `json:"estimated_cost"`
	ActualDuration   time.Duration          `json:"actual_duration"`
	RowsExamined     int64                  `json:"rows_examined"`
	RowsReturned     int64                  `json:"rows_returned"`
	Suggestions      []string               `json:"suggestions"`
	IndexesUsed      []string               `json:"indexes_used"`
	MissingIndexes   []string               `json:"missing_indexes"`
}

// SlowQuery represents a slow query log entry
type SlowQuery struct {
	QueryID      string        `json:"query_id"`
	Query        string        `json:"query"`
	Duration     time.Duration `json:"duration"`
	Timestamp    time.Time     `json:"timestamp"`
	Database     string        `json:"database"`
	User         string        `json:"user"`
	RowsExamined int64         `json:"rows_examined"`
	RowsReturned int64         `json:"rows_returned"`
}

// IndexRecommendation represents an index recommendation
type IndexRecommendation struct {
	Table            string   `json:"table"`
	Columns          []string `json:"columns"`
	Type             string   `json:"type"`
	EstimatedBenefit float64  `json:"estimated_benefit"`
	Reason           string   `json:"reason"`
	CreateStatement  string   `json:"create_statement"`
}

// Migration represents a database migration
type Migration struct {
	Version     string    `json:"version"`
	Name        string    `json:"name"`
	UpScript    string    `json:"up_script"`
	DownScript  string    `json:"down_script"`
	Checksum    string    `json:"checksum"`
	CreatedAt   time.Time `json:"created_at"`
}

// MigrationHistory represents migration history entry
type MigrationHistory struct {
	Version      string        `json:"version"`
	Name         string        `json:"name"`
	AppliedAt    time.Time     `json:"applied_at"`
	Duration     time.Duration `json:"duration"`
	Success      bool          `json:"success"`
	ErrorMessage string        `json:"error_message,omitempty"`
}

// MigrationPlan represents a migration execution plan
type MigrationPlan struct {
	Steps           []*MigrationStep `json:"steps"`
	EstimatedTime   time.Duration    `json:"estimated_time"`
	RollbackPlan    string           `json:"rollback_plan"`
	RiskAssessment  string           `json:"risk_assessment"`
	RequiresDowntime bool            `json:"requires_downtime"`
}

// MigrationStep represents a single migration step
type MigrationStep struct {
	Order       int           `json:"order"`
	Name        string        `json:"name"`
	SQL         string        `json:"sql"`
	Duration    time.Duration `json:"duration"`
	Reversible  bool          `json:"reversible"`
}

// SecurityPolicy represents a row-level security policy
type SecurityPolicy struct {
	Name        string `json:"name"`
	Table       string `json:"table"`
	Operation   string `json:"operation"`
	Expression  string `json:"expression"`
	Roles       []string `json:"roles"`
	Enabled     bool   `json:"enabled"`
}

// AuditOperation represents an auditable database operation
type AuditOperation struct {
	ID          string                 `json:"id"`
	Timestamp   time.Time              `json:"timestamp"`
	User        string                 `json:"user"`
	Operation   string                 `json:"operation"`
	Table       string                 `json:"table"`
	Query       string                 `json:"query"`
	RowsAffected int64                 `json:"rows_affected"`
	Duration    time.Duration          `json:"duration"`
	Success     bool                   `json:"success"`
	ErrorMessage string                `json:"error_message,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// PIIDetectionResult represents PII detection results
type PIIDetectionResult struct {
	HasPII       bool                    `json:"has_pii"`
	PIIFields    map[string]PIIFieldInfo `json:"pii_fields"`
	MaskedData   interface{}             `json:"masked_data"`
	Confidence   float64                 `json:"confidence"`
}

// PIIFieldInfo represents information about a PII field
type PIIFieldInfo struct {
	FieldName  string  `json:"field_name"`
	PIIType    string  `json:"pii_type"`
	Confidence float64 `json:"confidence"`
	MaskMethod string  `json:"mask_method"`
}

// BackupRequest represents backup configuration request
type BackupRequest struct {
	DatabaseName     string            `json:"database_name"`
	BackupType       string            `json:"backup_type"`
	Compression      bool              `json:"compression"`
	Encryption       bool              `json:"encryption"`
	StorageLocation  string            `json:"storage_location"`
	RetentionDays    int               `json:"retention_days"`
	Metadata         map[string]string `json:"metadata"`
}

// BackupResult represents backup operation result
type BackupResult struct {
	BackupID         string        `json:"backup_id"`
	StartTime        time.Time     `json:"start_time"`
	EndTime          time.Time     `json:"end_time"`
	Size             int64         `json:"size"`
	CompressedSize   int64         `json:"compressed_size"`
	Status           string        `json:"status"`
	StorageLocation  string        `json:"storage_location"`
	Checksum         string        `json:"checksum"`
}

// BackupInfo represents backup information
type BackupInfo struct {
	BackupID         string    `json:"backup_id"`
	DatabaseName     string    `json:"database_name"`
	BackupType       string    `json:"backup_type"`
	CreatedAt        time.Time `json:"created_at"`
	Size             int64     `json:"size"`
	Status           string    `json:"status"`
	StorageLocation  string    `json:"storage_location"`
	ExpiresAt        time.Time `json:"expires_at"`
}

// BackupSchedule represents backup scheduling configuration
type BackupSchedule struct {
	ScheduleID       string            `json:"schedule_id"`
	DatabaseName     string            `json:"database_name"`
	CronExpression   string            `json:"cron_expression"`
	BackupConfig     *BackupConfig     `json:"backup_config"`
	Enabled          bool              `json:"enabled"`
	LastRun          *time.Time        `json:"last_run,omitempty"`
	NextRun          *time.Time        `json:"next_run,omitempty"`
}

// RecoveryTestResult represents disaster recovery test results
type RecoveryTestResult struct {
	TestID           string        `json:"test_id"`
	StartTime        time.Time     `json:"start_time"`
	EndTime          time.Time     `json:"end_time"`
	RecoveryTime     time.Duration `json:"recovery_time"`
	DataIntegrity    bool          `json:"data_integrity"`
	Success          bool          `json:"success"`
	Issues           []string      `json:"issues"`
}

// DatabaseMetrics represents database performance metrics
type DatabaseMetrics struct {
	Timestamp        time.Time              `json:"timestamp"`
	ConnectionsActive int                   `json:"connections_active"`
	ConnectionsIdle   int                   `json:"connections_idle"`
	QueriesPerSecond  float64               `json:"queries_per_second"`
	AverageQueryTime  float64               `json:"average_query_time"`
	CacheHitRatio     float64               `json:"cache_hit_ratio"`
	DiskUsage         int64                 `json:"disk_usage"`
	CPUUsage          float64               `json:"cpu_usage"`
	MemoryUsage       float64               `json:"memory_usage"`
	ReplicationLag    time.Duration         `json:"replication_lag"`
	CustomMetrics     map[string]float64    `json:"custom_metrics"`
}

// Anomaly represents a detected anomaly
type Anomaly struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"`
	Severity     string                 `json:"severity"`
	Description  string                 `json:"description"`
	DetectedAt   time.Time              `json:"detected_at"`
	MetricName   string                 `json:"metric_name"`
	ExpectedValue float64               `json:"expected_value"`
	ActualValue   float64               `json:"actual_value"`
	Confidence    float64               `json:"confidence"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// Alert represents a monitoring alert
type Alert struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Type         string                 `json:"type"`
	Condition    string                 `json:"condition"`
	Threshold    float64                `json:"threshold"`
	Severity     string                 `json:"severity"`
	Enabled      bool                   `json:"enabled"`
	Actions      []string               `json:"actions"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// DashboardData represents monitoring dashboard data
type DashboardData struct {
	Overview         *DatabaseOverview      `json:"overview"`
	Metrics          *DatabaseMetrics       `json:"metrics"`
	RecentQueries    []*SlowQuery          `json:"recent_queries"`
	ActiveAlerts     []*Alert              `json:"active_alerts"`
	ConnectionPools  map[string]*PoolStats `json:"connection_pools"`
	BackupStatus     []*BackupInfo         `json:"backup_status"`
}

// DatabaseOverview represents database overview information
type DatabaseOverview struct {
	TotalDatabases   int      `json:"total_databases"`
	TotalTables      int      `json:"total_tables"`
	TotalSize        int64    `json:"total_size"`
	Uptime           time.Duration `json:"uptime"`
	Version          string   `json:"version"`
	HealthStatus     string   `json:"health_status"`
}