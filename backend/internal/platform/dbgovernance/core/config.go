// Package core provides configuration for the database governance system
package core

import (
	"time"
)

// DBGovernanceConfig represents the database governance configuration
type DBGovernanceConfig struct {
	// Connection Pool Configuration
	ConnectionPool ConnectionPoolConfig `json:"connection_pool"`
	
	// Query Performance Configuration
	QueryPerformance QueryPerformanceConfig `json:"query_performance"`
	
	// Migration Configuration
	Migration MigrationConfig `json:"migration"`
	
	// Security Configuration
	Security SecurityConfig `json:"security"`
	
	// Backup Configuration
	Backup BackupConfig `json:"backup"`
	
	// Monitoring Configuration
	Monitoring MonitoringConfig `json:"monitoring"`
	
	// Database Configurations
	Databases map[string]DatabaseConfig `json:"databases"`
}

// ConnectionPoolConfig represents connection pool configuration
type ConnectionPoolConfig struct {
	DefaultMinSize        int           `json:"default_min_size"`
	DefaultMaxSize        int           `json:"default_max_size"`
	ConnectionTimeout     time.Duration `json:"connection_timeout"`
	IdleTimeout           time.Duration `json:"idle_timeout"`
	MaxLifetime           time.Duration `json:"max_lifetime"`
	HealthCheckInterval   time.Duration `json:"health_check_interval"`
	EnableAdaptiveSizing  bool          `json:"enable_adaptive_sizing"`
	AdaptiveCheckInterval time.Duration `json:"adaptive_check_interval"`
}

// QueryPerformanceConfig represents query performance configuration
type QueryPerformanceConfig struct {
	SlowQueryThreshold    time.Duration `json:"slow_query_threshold"`
	EnableQueryCache      bool          `json:"enable_query_cache"`
	QueryCacheTTL         time.Duration `json:"query_cache_ttl"`
	QueryCacheMaxSize     int64         `json:"query_cache_max_size"`
	EnableIndexAdvisor    bool          `json:"enable_index_advisor"`
	AnalysisInterval      time.Duration `json:"analysis_interval"`
	MaxSlowQueries        int           `json:"max_slow_queries"`
	EnableAutoOptimization bool         `json:"enable_auto_optimization"`
}

// MigrationConfig represents migration configuration
type MigrationConfig struct {
	MigrationsPath       string        `json:"migrations_path"`
	EnableVersionControl bool          `json:"enable_version_control"`
	LockTimeout          time.Duration `json:"lock_timeout"`
	ValidationMode       string        `json:"validation_mode"`
	EnableDryRun         bool          `json:"enable_dry_run"`
	BackupBeforeMigration bool         `json:"backup_before_migration"`
	ParallelMigrations   int           `json:"parallel_migrations"`
}

// SecurityConfig represents security configuration
type SecurityConfig struct {
	EncryptionAlgorithm   string        `json:"encryption_algorithm"`
	EncryptionKey         string        `json:"encryption_key"`
	EnableAuditLog        bool          `json:"enable_audit_log"`
	AuditLogRetention     time.Duration `json:"audit_log_retention"`
	EnablePIIDetection    bool          `json:"enable_pii_detection"`
	PIIDetectionRules     []string      `json:"pii_detection_rules"`
	EnableRowLevelSecurity bool         `json:"enable_row_level_security"`
	MaxAuditLogSize       int64         `json:"max_audit_log_size"`
}

// BackupConfig represents backup configuration
type BackupConfig struct {
	DefaultStoragePath    string        `json:"default_storage_path"`
	DefaultRetentionDays  int           `json:"default_retention_days"`
	EnableCompression     bool          `json:"enable_compression"`
	CompressionLevel      int           `json:"compression_level"`
	EnableEncryption      bool          `json:"enable_encryption"`
	ParallelBackups       int           `json:"parallel_backups"`
	BackupTimeout         time.Duration `json:"backup_timeout"`
	EnableIncrementalBackup bool        `json:"enable_incremental_backup"`
	GeoRedundantRegions   []string      `json:"geo_redundant_regions"`
}

// MonitoringConfig represents monitoring configuration
type MonitoringConfig struct {
	MetricsInterval       time.Duration          `json:"metrics_interval"`
	EnableAnomalyDetection bool                  `json:"enable_anomaly_detection"`
	AnomalyThreshold      float64                `json:"anomaly_threshold"`
	AlertChannels         []string               `json:"alert_channels"`
	DashboardRefreshRate  time.Duration          `json:"dashboard_refresh_rate"`
	MetricsRetention      time.Duration          `json:"metrics_retention"`
	PrometheusEndpoint    string                 `json:"prometheus_endpoint"`
	GrafanaURL            string                 `json:"grafana_url"`
	CustomMetrics         map[string]MetricConfig `json:"custom_metrics"`
}

// DatabaseConfig represents individual database configuration
type DatabaseConfig struct {
	Type              DatabaseType      `json:"type"`
	ConnectionString  string            `json:"connection_string"`
	MaxConnections    int               `json:"max_connections"`
	MinConnections    int               `json:"min_connections"`
	ReadReplicas      []string          `json:"read_replicas"`
	WriteEndpoint     string            `json:"write_endpoint"`
	EnableSharding    bool              `json:"enable_sharding"`
	ShardKey          string            `json:"shard_key"`
	CustomSettings    map[string]string `json:"custom_settings"`
}

// MetricConfig represents custom metric configuration
type MetricConfig struct {
	Query       string        `json:"query"`
	Interval    time.Duration `json:"interval"`
	Threshold   float64       `json:"threshold"`
	AlertLevel  string        `json:"alert_level"`
}

// DefaultDBGovernanceConfig returns the default configuration
func DefaultDBGovernanceConfig() *DBGovernanceConfig {
	return &DBGovernanceConfig{
		ConnectionPool: ConnectionPoolConfig{
			DefaultMinSize:        5,
			DefaultMaxSize:        100,
			ConnectionTimeout:     30 * time.Second,
			IdleTimeout:           30 * time.Minute,
			MaxLifetime:           1 * time.Hour,
			HealthCheckInterval:   1 * time.Minute,
			EnableAdaptiveSizing:  true,
			AdaptiveCheckInterval: 5 * time.Minute,
		},
		QueryPerformance: QueryPerformanceConfig{
			SlowQueryThreshold:     1 * time.Second,
			EnableQueryCache:       true,
			QueryCacheTTL:          5 * time.Minute,
			QueryCacheMaxSize:      1024 * 1024 * 100, // 100MB
			EnableIndexAdvisor:     true,
			AnalysisInterval:       10 * time.Minute,
			MaxSlowQueries:         1000,
			EnableAutoOptimization: false,
		},
		Migration: MigrationConfig{
			MigrationsPath:        "./migrations",
			EnableVersionControl:  true,
			LockTimeout:           5 * time.Minute,
			ValidationMode:        "strict",
			EnableDryRun:          true,
			BackupBeforeMigration: true,
			ParallelMigrations:    1,
		},
		Security: SecurityConfig{
			EncryptionAlgorithm:    "AES-256-GCM",
			EnableAuditLog:         true,
			AuditLogRetention:      90 * 24 * time.Hour, // 90 days
			EnablePIIDetection:     true,
			PIIDetectionRules:      []string{"email", "phone", "ssn", "credit_card"},
			EnableRowLevelSecurity: true,
			MaxAuditLogSize:        1024 * 1024 * 1024, // 1GB
		},
		Backup: BackupConfig{
			DefaultStoragePath:      "./backups",
			DefaultRetentionDays:    30,
			EnableCompression:       true,
			CompressionLevel:        6,
			EnableEncryption:        true,
			ParallelBackups:         2,
			BackupTimeout:           1 * time.Hour,
			EnableIncrementalBackup: true,
			GeoRedundantRegions:     []string{"us-east-1", "us-west-2"},
		},
		Monitoring: MonitoringConfig{
			MetricsInterval:        10 * time.Second,
			EnableAnomalyDetection: true,
			AnomalyThreshold:       2.5,
			AlertChannels:          []string{"email", "slack"},
			DashboardRefreshRate:   30 * time.Second,
			MetricsRetention:       30 * 24 * time.Hour, // 30 days
			PrometheusEndpoint:     "http://localhost:9090",
			GrafanaURL:             "http://localhost:3000",
			CustomMetrics:          make(map[string]MetricConfig),
		},
		Databases: make(map[string]DatabaseConfig),
	}
}

// LoadFromFile loads configuration from a file
func LoadFromFile(path string) (*DBGovernanceConfig, error) {
	// Implementation would load from JSON/YAML file
	// For now, return default config
	return DefaultDBGovernanceConfig(), nil
}

// Validate validates the configuration
func (c *DBGovernanceConfig) Validate() error {
	// Add validation logic here
	return nil
}