/**
 * 统一数据库管理器单元测试
 * 测试多数据库支持、连接池管理、健康检查等功能
 */

package database

import (
	"testing"
	"time"

	"gorm.io/gorm/logger"
)

// TestDatabaseConfig 测试数据库配置
func TestDatabaseConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "有效的MySQL配置",
			config: &Config{
				Type:     MySQL,
				Host:     "localhost",
				Port:     3306,
				Database: "test_db",
				Username: "root",
				Password: "password",
			},
			wantErr: false,
		},
		{
			name: "有效的PostgreSQL配置",
			config: &Config{
				Type:     PostgreSQL,
				Host:     "localhost",
				Port:     5432,
				Database: "test_db",
				Username: "postgres",
				Password: "password",
				SSLMode:  "disable",
			},
			wantErr: false,
		},
		{
			name: "有效的SQLite配置",
			config: &Config{
				Type:     SQLite,
				Database: "test.db",
			},
			wantErr: false,
		},
		{
			name: "缺少数据库类型",
			config: &Config{
				Host:     "localhost",
				Database: "test_db",
			},
			wantErr: true,
		},
		{
			name: "MySQL缺少主机",
			config: &Config{
				Type:     MySQL,
				Database: "test_db",
				Username: "root",
			},
			wantErr: true,
		},
		{
			name: "缺少数据库名",
			config: &Config{
				Type:     MySQL,
				Host:     "localhost",
				Username: "root",
			},
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestDatabaseManager 测试数据库管理器
func TestDatabaseManager(t *testing.T) {
	manager := NewManager()
	
	// 测试添加配置
	config := &Config{
		Type:     SQLite,
		Database: ":memory:",
		LogLevel: logger.Silent,
	}
	
	err := manager.AddConfig("test", config)
	if err != nil {
		t.Fatalf("AddConfig failed: %v", err)
	}
	
	// 测试添加nil配置
	err = manager.AddConfig("nil", nil)
	if err == nil {
		t.Error("AddConfig should fail with nil config")
	}
	
	// 测试连接数据库
	db, err := manager.Connect("test")
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	if db == nil {
		t.Error("Connect should return a valid DB connection")
	}
	
	// 测试重复连接（应该返回相同连接）
	db2, err := manager.Connect("test")
	if err != nil {
		t.Fatalf("Second Connect failed: %v", err)
	}
	if db != db2 {
		t.Error("Should return the same connection for repeated Connect calls")
	}
	
	// 测试获取不存在的连接
	_, err = manager.GetConnection("nonexistent")
	if err == nil {
		t.Error("GetConnection should fail for non-existent connection")
	}
	
	// 测试关闭连接
	err = manager.CloseConnection("test")
	if err != nil {
		t.Errorf("CloseConnection failed: %v", err)
	}
	
	// 测试关闭不存在的连接
	err = manager.CloseConnection("nonexistent")
	if err == nil {
		t.Error("CloseConnection should fail for non-existent connection")
	}
}

// TestMultipleDatabases 测试多数据库管理
func TestMultipleDatabases(t *testing.T) {
	manager := NewManager()
	
	// 添加多个数据库配置
	configs := map[string]*Config{
		"main": {
			Type:     SQLite,
			Database: ":memory:",
			LogLevel: logger.Silent,
		},
		"analytics": {
			Type:     SQLite,
			Database: ":memory:",
			LogLevel: logger.Silent,
		},
		"cache": {
			Type:     SQLite,
			Database: ":memory:",
			LogLevel: logger.Silent,
		},
	}
	
	// 添加所有配置
	for name, config := range configs {
		err := manager.AddConfig(name, config)
		if err != nil {
			t.Fatalf("AddConfig(%s) failed: %v", name, err)
		}
	}
	
	// 连接所有数据库
	connections := make(map[string]bool)
	for name := range configs {
		db, err := manager.Connect(name)
		if err != nil {
			t.Fatalf("Connect(%s) failed: %v", name, err)
		}
		if db == nil {
			t.Errorf("Connect(%s) returned nil", name)
		}
		connections[name] = true
	}
	
	// 验证所有连接
	if len(connections) != len(configs) {
		t.Errorf("Expected %d connections, got %d", len(configs), len(connections))
	}
	
	// 获取统计信息
	stats := manager.GetStats()
	if len(stats) != len(configs) {
		t.Errorf("Expected %d stats entries, got %d", len(configs), len(stats))
	}
	
	// 关闭所有连接
	err := manager.CloseAll()
	if err != nil {
		t.Errorf("CloseAll failed: %v", err)
	}
}

// TestConnectionPoolConfig 测试连接池配置
func TestConnectionPoolConfig(t *testing.T) {
	manager := NewManager()
	
	config := &Config{
		Type:            SQLite,
		Database:        ":memory:",
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 30 * time.Minute,
		LogLevel:        logger.Silent,
	}
	
	err := manager.AddConfig("pool_test", config)
	if err != nil {
		t.Fatalf("AddConfig failed: %v", err)
	}
	
	db, err := manager.Connect("pool_test")
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	
	// 验证连接池设置
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("Failed to get sql.DB: %v", err)
	}
	
	stats := sqlDB.Stats()
	if stats.MaxOpenConnections != 10 {
		t.Errorf("MaxOpenConnections = %d, want 10", stats.MaxOpenConnections)
	}
}

// TestHealthChecker 测试健康检查器
func TestHealthChecker(t *testing.T) {
	hc := NewHealthChecker()
	
	// 创建测试数据库管理器
	manager := NewManager()
	config := &Config{
		Type:                SQLite,
		Database:            ":memory:",
		LogLevel:            logger.Silent,
		HealthCheckInterval: 100 * time.Millisecond,
	}
	
	err := manager.AddConfig("health_test", config)
	if err != nil {
		t.Fatalf("AddConfig failed: %v", err)
	}
	
	_, err = manager.Connect("health_test")
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	
	// 等待健康检查执行
	time.Sleep(200 * time.Millisecond)
	
	// 获取健康状态
	health := manager.healthCheck.GetHealth()
	if len(health) != 1 {
		t.Errorf("Expected 1 health entry, got %d", len(health))
	}
	
	if healthInfo, exists := health["health_test"]; exists {
		if !healthInfo.IsHealthy {
			t.Error("Database should be healthy")
		}
		if healthInfo.ErrorCount != 0 {
			t.Errorf("ErrorCount = %d, want 0", healthInfo.ErrorCount)
		}
	} else {
		t.Error("Health entry not found")
	}
	
	// 停止健康检查
	hc.Stop()
}

// TestDefaultManager 测试默认管理器单例
func TestDefaultManager(t *testing.T) {
	manager1 := GetDefaultManager()
	manager2 := GetDefaultManager()
	
	if manager1 != manager2 {
		t.Error("GetDefaultManager should return the same instance")
	}
}

// TestConvenienceFunctions 测试便捷函数
func TestConvenienceFunctions(t *testing.T) {
	config := &Config{
		Type:     SQLite,
		Database: ":memory:",
		LogLevel: logger.Silent,
	}
	
	// 初始化默认数据库
	db, err := InitDefaultDatabase(config)
	if err != nil {
		t.Fatalf("InitDefaultDatabase failed: %v", err)
	}
	if db == nil {
		t.Error("InitDefaultDatabase should return a valid connection")
	}
	
	// 获取默认连接
	db2, err := GetDefaultConnection()
	if err != nil {
		t.Fatalf("GetDefaultConnection failed: %v", err)
	}
	if db != db2 {
		t.Error("Should return the same default connection")
	}
	
	// 获取统计信息
	stats := GetDatabaseStats()
	if len(stats) == 0 {
		t.Error("GetDatabaseStats should return stats")
	}
	
	// 获取健康状态
	health := GetDatabaseHealth()
	if len(health) == 0 {
		t.Error("GetDatabaseHealth should return health info")
	}
	
	// 关闭默认连接
	err = CloseDefaultConnection()
	if err != nil {
		t.Errorf("CloseDefaultConnection failed: %v", err)
	}
}

// TestDefaultConfigValues 测试默认配置值
func TestDefaultConfigValues(t *testing.T) {
	config := &Config{
		Type:     MySQL,
		Host:     "localhost",
		Database: "test",
		Username: "root",
		// 不设置连接池参数，应该使用默认值
	}
	
	err := validateConfig(config)
	if err != nil {
		t.Fatalf("validateConfig failed: %v", err)
	}
	
	// 验证默认值已设置
	if config.MaxOpenConns != DefaultConfig.MaxOpenConns {
		t.Errorf("MaxOpenConns = %d, want %d", config.MaxOpenConns, DefaultConfig.MaxOpenConns)
	}
	if config.MaxIdleConns != DefaultConfig.MaxIdleConns {
		t.Errorf("MaxIdleConns = %d, want %d", config.MaxIdleConns, DefaultConfig.MaxIdleConns)
	}
	if config.ConnMaxLifetime != DefaultConfig.ConnMaxLifetime {
		t.Errorf("ConnMaxLifetime = %v, want %v", config.ConnMaxLifetime, DefaultConfig.ConnMaxLifetime)
	}
}

// BenchmarkConnect 性能测试 - 连接数据库
func BenchmarkConnect(b *testing.B) {
	manager := NewManager()
	
	// 预先添加配置
	for i := 0; i < b.N; i++ {
		config := &Config{
			Type:     SQLite,
			Database: ":memory:",
			LogLevel: logger.Silent,
		}
		manager.AddConfig(string(rune(i)), config)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := manager.Connect(string(rune(i)))
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGetStats 性能测试 - 获取统计信息
func BenchmarkGetStats(b *testing.B) {
	manager := NewManager()
	
	// 设置测试连接
	config := &Config{
		Type:     SQLite,
		Database: ":memory:",
		LogLevel: logger.Silent,
	}
	manager.AddConfig("bench", config)
	manager.Connect("bench")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = manager.GetStats()
	}
}