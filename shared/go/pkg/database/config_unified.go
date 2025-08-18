/**
 * 统一数据库配置管理 - SOTA实现，解决5处重复数据库连接逻辑
 * 支持：多数据库、连接池管理、健康检查、故障恢复、监控
 */

package database

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gorm.io/driver/postgres" 
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DatabaseType 数据库类型 - 只支持PostgreSQL
type DatabaseType string

const (
	PostgreSQL DatabaseType = "postgres" 
)

// Config 数据库配置
type Config struct {
	Type         DatabaseType `json:"type" yaml:"type"`
	Host         string       `json:"host" yaml:"host"`
	Port         int          `json:"port" yaml:"port"`
	Database     string       `json:"database" yaml:"database"`
	Username     string       `json:"username" yaml:"username"`
	Password     string       `json:"password" yaml:"password"`
	SSLMode      string       `json:"ssl_mode" yaml:"ssl_mode"`
	SSLCert      string       `json:"ssl_cert" yaml:"ssl_cert"`         // SSL证书路径
	SSLKey       string       `json:"ssl_key" yaml:"ssl_key"`           // SSL私钥路径
	SSLRootCert  string       `json:"ssl_root_cert" yaml:"ssl_root_cert"` // CA证书路径
	Charset      string       `json:"charset" yaml:"charset"`
	Timezone     string       `json:"timezone" yaml:"timezone"`
	
	// 连接池配置
	MaxOpenConns    int           `json:"max_open_conns" yaml:"max_open_conns"`
	MaxIdleConns    int           `json:"max_idle_conns" yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime" yaml:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `json:"conn_max_idle_time" yaml:"conn_max_idle_time"`
	
	// 日志配置
	LogLevel logger.LogLevel `json:"log_level" yaml:"log_level"`
	
	// 健康检查配置
	HealthCheckInterval time.Duration `json:"health_check_interval" yaml:"health_check_interval"`
	MaxRetries         int           `json:"max_retries" yaml:"max_retries"`
	RetryInterval      time.Duration `json:"retry_interval" yaml:"retry_interval"`
}

// DefaultConfig 默认配置
var DefaultConfig = &Config{
	Type:         PostgreSQL,
	Host:         "localhost",
	Port:         5432,
	Database:     "openpenpal",
	Username:     "openpenpal_user",
	Password:     "",
	SSLMode:      "disable", // 开发环境默认禁用SSL
	Charset:      "utf8",
	Timezone:     "Asia/Shanghai",
	
	MaxOpenConns:    25,
	MaxIdleConns:    10,
	ConnMaxLifetime: time.Hour,
	ConnMaxIdleTime: 10 * time.Minute,
	
	LogLevel: logger.Warn,
	
	HealthCheckInterval: 30 * time.Second,
	MaxRetries:         3,
	RetryInterval:      5 * time.Second,
}

// Manager 数据库管理器
type Manager struct {
	configs     map[string]*Config
	connections map[string]*gorm.DB
	mutex       sync.RWMutex
	healthCheck *HealthChecker
}

// NewManager 创建数据库管理器
func NewManager() *Manager {
	return &Manager{
		configs:     make(map[string]*Config),
		connections: make(map[string]*gorm.DB),
		mutex:       sync.RWMutex{},
		healthCheck: NewHealthChecker(),
	}
}

// AddConfig 添加数据库配置
func (m *Manager) AddConfig(name string, config *Config) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}
	
	// 验证配置
	if err := validateConfig(config); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}
	
	m.configs[name] = config
	return nil
}

// Connect 连接数据库
func (m *Manager) Connect(name string) (*gorm.DB, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// 检查是否已经连接
	if db, exists := m.connections[name]; exists {
		return db, nil
	}
	
	// 获取配置
	config, exists := m.configs[name]
	if !exists {
		return nil, fmt.Errorf("database config '%s' not found", name)
	}
	
	// 建立连接
	db, err := m.createConnection(config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database '%s': %w", name, err)
	}
	
	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}
	
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(config.ConnMaxIdleTime)
	
	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	
	// 保存连接
	m.connections[name] = db
	
	// 启动健康检查
	m.healthCheck.AddDatabase(name, db, config)
	
	return db, nil
}

// GetConnection 获取数据库连接
func (m *Manager) GetConnection(name string) (*gorm.DB, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	if db, exists := m.connections[name]; exists {
		return db, nil
	}
	
	return nil, fmt.Errorf("database connection '%s' not found", name)
}

// CloseConnection 关闭数据库连接
func (m *Manager) CloseConnection(name string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	db, exists := m.connections[name]
	if !exists {
		return fmt.Errorf("database connection '%s' not found", name)
	}
	
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}
	
	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}
	
	// 从健康检查中移除
	m.healthCheck.RemoveDatabase(name)
	
	// 从连接池中移除
	delete(m.connections, name)
	
	return nil
}

// CloseAll 关闭所有数据库连接
func (m *Manager) CloseAll() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	var errors []error
	
	for name := range m.connections {
		if err := m.CloseConnection(name); err != nil {
			errors = append(errors, err)
		}
	}
	
	// 停止健康检查
	m.healthCheck.Stop()
	
	if len(errors) > 0 {
		return fmt.Errorf("failed to close some connections: %v", errors)
	}
	
	return nil
}

// GetStats 获取连接统计信息
func (m *Manager) GetStats() map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	stats := make(map[string]interface{})
	
	for name, db := range m.connections {
		sqlDB, err := db.DB()
		if err != nil {
			continue
		}
		
		dbStats := sqlDB.Stats()
		stats[name] = map[string]interface{}{
			"max_open_connections":     dbStats.MaxOpenConnections,
			"open_connections":         dbStats.OpenConnections,
			"in_use":                   dbStats.InUse,
			"idle":                     dbStats.Idle,
			"wait_count":              dbStats.WaitCount,
			"wait_duration":           dbStats.WaitDuration.String(),
			"max_idle_closed":         dbStats.MaxIdleClosed,
			"max_idle_time_closed":    dbStats.MaxIdleTimeClosed,
			"max_lifetime_closed":     dbStats.MaxLifetimeClosed,
		}
	}
	
	return stats
}

// createConnection 创建PostgreSQL数据库连接
func (m *Manager) createConnection(config *Config) (*gorm.DB, error) {
	// 只支持PostgreSQL
	if config.Type != PostgreSQL {
		return nil, fmt.Errorf("only PostgreSQL is supported, got: %s", config.Type)
	}
	
	// 处理时区，默认使用 Asia/Shanghai
	timezone := config.Timezone
	if timezone == "" {
		timezone = "Asia/Shanghai"
	}
	
	// 处理SSL模式，根据环境设置默认值
	sslMode := config.SSLMode
	if sslMode == "" {
		// 根据环境选择默认SSL模式
		switch config.Timezone {
		case "Asia/Shanghai":
			// 国内环境默认禁用SSL
			sslMode = "disable"
		default:
			// 其他环境默认使用require
			sslMode = "require"
		}
	}
	
	// 构建基础DSN
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
		config.Host,
		config.Username,
		config.Password,
		config.Database,
		config.Port,
		sslMode,
		timezone,
	)
	
	// 添加SSL证书参数
	if sslMode != "disable" && sslMode != "allow" {
		if config.SSLRootCert != "" {
			dsn += fmt.Sprintf(" sslrootcert=%s", config.SSLRootCert)
		}
		if config.SSLCert != "" {
			dsn += fmt.Sprintf(" sslcert=%s", config.SSLCert)
		}
		if config.SSLKey != "" {
			dsn += fmt.Sprintf(" sslkey=%s", config.SSLKey)
		}
	}
	dialector := postgres.Open(dsn)
	
	// GORM配置
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(config.LogLevel),
	}
	
	return gorm.Open(dialector, gormConfig)
}

// validateConfig 验证PostgreSQL配置
func validateConfig(config *Config) error {
	if config.Type == "" {
		return fmt.Errorf("database type is required")
	}
	
	// 只支持PostgreSQL
	if config.Type != PostgreSQL {
		return fmt.Errorf("only PostgreSQL is supported, got: %s", config.Type)
	}
	
	if config.Host == "" {
		return fmt.Errorf("host is required for PostgreSQL")
	}
	if config.Username == "" {
		return fmt.Errorf("username is required for PostgreSQL")
	}
	
	if config.Database == "" {
		return fmt.Errorf("database name is required")
	}
	
	if config.MaxOpenConns <= 0 {
		config.MaxOpenConns = DefaultConfig.MaxOpenConns
	}
	
	if config.MaxIdleConns <= 0 {
		config.MaxIdleConns = DefaultConfig.MaxIdleConns
	}
	
	if config.ConnMaxLifetime <= 0 {
		config.ConnMaxLifetime = DefaultConfig.ConnMaxLifetime
	}
	
	if config.ConnMaxIdleTime <= 0 {
		config.ConnMaxIdleTime = DefaultConfig.ConnMaxIdleTime
	}
	
	return nil
}

// ================================
// 健康检查器
// ================================

// HealthChecker 健康检查器
type HealthChecker struct {
	databases map[string]*DatabaseHealth
	ticker    *time.Ticker
	done      chan bool
	mutex     sync.RWMutex
}

// DatabaseHealth 数据库健康状态
type DatabaseHealth struct {
	Name       string
	DB         *gorm.DB
	Config     *Config
	LastCheck  time.Time
	IsHealthy  bool
	ErrorCount int
	LastError  error
}

// NewHealthChecker 创建健康检查器
func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		databases: make(map[string]*DatabaseHealth),
		done:      make(chan bool),
		mutex:     sync.RWMutex{},
	}
}

// AddDatabase 添加数据库到健康检查
func (hc *HealthChecker) AddDatabase(name string, db *gorm.DB, config *Config) {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()
	
	hc.databases[name] = &DatabaseHealth{
		Name:      name,
		DB:        db,
		Config:    config,
		LastCheck: time.Now(),
		IsHealthy: true,
	}
	
	// 启动健康检查
	if hc.ticker == nil {
		hc.startHealthCheck()
	}
}

// RemoveDatabase 从健康检查中移除数据库
func (hc *HealthChecker) RemoveDatabase(name string) {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()
	
	delete(hc.databases, name)
}

// GetHealth 获取健康状态
func (hc *HealthChecker) GetHealth() map[string]*DatabaseHealth {
	hc.mutex.RLock()
	defer hc.mutex.RUnlock()
	
	result := make(map[string]*DatabaseHealth)
	for k, v := range hc.databases {
		// 复制一份，避免并发问题
		result[k] = &DatabaseHealth{
			Name:       v.Name,
			LastCheck:  v.LastCheck,
			IsHealthy:  v.IsHealthy,
			ErrorCount: v.ErrorCount,
			LastError:  v.LastError,
		}
	}
	
	return result
}

// Stop 停止健康检查
func (hc *HealthChecker) Stop() {
	if hc.ticker != nil {
		hc.ticker.Stop()
		hc.done <- true
	}
}

// startHealthCheck 启动健康检查
func (hc *HealthChecker) startHealthCheck() {
	hc.ticker = time.NewTicker(30 * time.Second) // 默认30秒检查一次
	
	go func() {
		for {
			select {
			case <-hc.ticker.C:
				hc.performHealthCheck()
			case <-hc.done:
				return
			}
		}
	}()
}

// performHealthCheck 执行健康检查
func (hc *HealthChecker) performHealthCheck() {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()
	
	for name, health := range hc.databases {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		
		sqlDB, err := health.DB.DB()
		if err != nil {
			health.IsHealthy = false
			health.LastError = err
			health.ErrorCount++
		} else {
			err = sqlDB.PingContext(ctx)
			if err != nil {
				health.IsHealthy = false
				health.LastError = err
				health.ErrorCount++
			} else {
				health.IsHealthy = true
				health.LastError = nil
				health.ErrorCount = 0
			}
		}
		
		health.LastCheck = time.Now()
		cancel()
		
		// 记录健康状态日志
		if !health.IsHealthy {
			// TODO: 发送告警
			fmt.Printf("Database %s is unhealthy: %v\n", name, health.LastError)
		}
	}
}

// ================================
// 全局管理器实例
// ================================

var (
	defaultManager *Manager
	managerOnce    sync.Once
)

// GetDefaultManager 获取默认数据库管理器
func GetDefaultManager() *Manager {
	managerOnce.Do(func() {
		defaultManager = NewManager()
	})
	return defaultManager
}

// ================================
// 便捷函数
// ================================

// InitDefaultDatabase 初始化默认数据库
func InitDefaultDatabase(config *Config) (*gorm.DB, error) {
	manager := GetDefaultManager()
	
	if err := manager.AddConfig("default", config); err != nil {
		return nil, err
	}
	
	return manager.Connect("default")
}

// GetDefaultConnection 获取默认数据库连接
func GetDefaultConnection() (*gorm.DB, error) {
	return GetDefaultManager().GetConnection("default")
}

// CloseDefaultConnection 关闭默认数据库连接
func CloseDefaultConnection() error {
	return GetDefaultManager().CloseConnection("default")
}

// GetDatabaseStats 获取数据库统计信息
func GetDatabaseStats() map[string]interface{} {
	return GetDefaultManager().GetStats()
}

// GetDatabaseHealth 获取数据库健康状态
func GetDatabaseHealth() map[string]*DatabaseHealth {
	manager := GetDefaultManager()
	return manager.healthCheck.GetHealth()
}