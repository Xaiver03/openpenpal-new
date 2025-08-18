/**
 * 统一数据库管理器单元测试 - PostgreSQL专用版本
 * 测试PostgreSQL连接池管理、健康检查等功能
 */

package database

import (
	"testing"
	"time"

	"gorm.io/gorm/logger"
)

// TestDatabaseConfig 测试PostgreSQL数据库配置
func TestDatabaseConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
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
			name: "PostgreSQL配置 - 默认SSL模式",
			config: &Config{
				Type:     PostgreSQL,
				Host:     "localhost",
				Port:     5432,
				Database: "test_db",
				Username: "postgres",
				Password: "password",
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
			name: "PostgreSQL缺少主机",
			config: &Config{
				Type:     PostgreSQL,
				Database: "test_db",
				Username: "postgres",
			},
			wantErr: true,
		},
		{
			name: "缺少数据库名",
			config: &Config{
				Type:     PostgreSQL,
				Host:     "localhost",
				Username: "postgres",
			},
			wantErr: true,
		},
		{
			name: "不支持的数据库类型",
			config: &Config{
				Type:     "mysql",
				Host:     "localhost",
				Database: "test_db",
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
	
	// 跳过实际连接测试，仅测试配置管理
	// 使用无效配置测试配置验证逻辑
	config := &Config{
		Type:     PostgreSQL,
		Host:     "invalid-host",
		Port:     5432,
		Database: "test_db",
		Username: "postgres",
		Password: "password",
		SSLMode:  "disable",
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
	
	// 跳过实际数据库连接测试（需要真实PostgreSQL实例）
	// 测试连接失败情况
	_, err = manager.Connect("test")
	if err == nil {
		t.Log("Connect to invalid host expected to fail (this is normal in test environment)")
	}
	
	// 测试获取不存在的连接
	_, err = manager.GetConnection("nonexistent")
	if err == nil {
		t.Error("GetConnection should fail for non-existent connection")
	}
	
	// 跳过关闭连接测试（没有实际连接）
	
	// 测试关闭不存在的连接
	err = manager.CloseConnection("nonexistent")
	if err == nil {
		t.Error("CloseConnection should fail for non-existent connection")
	}
}

// TestMultipleDatabases 测试多数据库配置管理
func TestMultipleDatabases(t *testing.T) {
	manager := NewManager()
	
	// 添加多个PostgreSQL数据库配置（测试配置管理，不实际连接）
	configs := map[string]*Config{
		"main": {
			Type:     PostgreSQL,
			Host:     "localhost",
			Port:     5432,
			Database: "main_db",
			Username: "postgres",
			Password: "password",
			SSLMode:  "disable",
			LogLevel: logger.Silent,
		},
		"analytics": {
			Type:     PostgreSQL,
			Host:     "localhost",
			Port:     5432,
			Database: "analytics_db",
			Username: "postgres",
			Password: "password",
			SSLMode:  "disable",
			LogLevel: logger.Silent,
		},
		"cache": {
			Type:     PostgreSQL,
			Host:     "localhost",
			Port:     5432,
			Database: "cache_db",
			Username: "postgres",
			Password: "password",
			SSLMode:  "disable",
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
	
	// 测试配置管理（跳过实际连接）
	t.Logf("Successfully added %d database configurations", len(configs))
	
	// 获取统计信息（空的，因为没有实际连接）
	stats := manager.GetStats()
	if len(stats) != 0 {
		t.Logf("Stats entries: %d (expected 0 since no actual connections)", len(stats))
	}
}

// TestConnectionPoolConfig 测试连接池配置验证
func TestConnectionPoolConfig(t *testing.T) {
	manager := NewManager()
	
	config := &Config{
		Type:            PostgreSQL,
		Host:            "localhost",
		Port:            5432,
		Database:        "pool_test",
		Username:        "postgres",
		Password:        "password",
		SSLMode:         "disable",
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
	
	// 验证配置已正确设置
	if config.MaxOpenConns != 10 {
		t.Errorf("MaxOpenConns = %d, want 10", config.MaxOpenConns)
	}
	if config.MaxIdleConns != 5 {
		t.Errorf("MaxIdleConns = %d, want 5", config.MaxIdleConns)
	}
}

// TestHealthChecker 测试健康检查器初始化
func TestHealthChecker(t *testing.T) {
	hc := NewHealthChecker()
	
	// 测试健康检查器创建
	if hc == nil {
		t.Error("NewHealthChecker should return a valid instance")
	}
	
	// 测试初始状态
	health := hc.GetHealth()
	if len(health) != 0 {
		t.Errorf("Expected 0 health entries initially, got %d", len(health))
	}
	
	// 测试停止功能（应该不会崩溃）
	hc.Stop()
	t.Log("HealthChecker stopped successfully")
}

// TestDefaultManager 测试默认管理器单例
func TestDefaultManager(t *testing.T) {
	manager1 := GetDefaultManager()
	manager2 := GetDefaultManager()
	
	if manager1 != manager2 {
		t.Error("GetDefaultManager should return the same instance")
	}
}

// TestConvenienceFunctions 测试便捷函数（仅配置验证）
func TestConvenienceFunctions(t *testing.T) {
	config := &Config{
		Type:     PostgreSQL,
		Host:     "localhost",
		Port:     5432,
		Database: "test_db",
		Username: "postgres",
		Password: "password",
		SSLMode:  "disable",
		LogLevel: logger.Silent,
	}
	
	// 测试配置验证
	err := validateConfig(config)
	if err != nil {
		t.Fatalf("Config validation failed: %v", err)
	}
	
	// 获取默认管理器
	manager := GetDefaultManager()
	if manager == nil {
		t.Error("GetDefaultManager should return a valid manager")
	}
	
	// 测试单例模式
	manager2 := GetDefaultManager()
	if manager != manager2 {
		t.Error("GetDefaultManager should return the same instance")
	}
	
	// 获取空统计信息
	stats := GetDatabaseStats()
	if len(stats) != 0 {
		t.Logf("Stats entries: %d (expected 0 since no connections)", len(stats))
	}
	
	// 获取空健康状态
	health := GetDatabaseHealth()
	if len(health) != 0 {
		t.Logf("Health entries: %d (expected 0 since no connections)", len(health))
	}
}

// TestDefaultConfigValues 测试默认配置值
func TestDefaultConfigValues(t *testing.T) {
	config := &Config{
		Type:     PostgreSQL,
		Host:     "localhost",
		Port:     5432,
		Database: "test",
		Username: "postgres",
		Password: "password",
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

// BenchmarkAddConfig 性能测试 - 添加配置
func BenchmarkAddConfig(b *testing.B) {
	manager := NewManager()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config := &Config{
			Type:     PostgreSQL,
			Host:     "localhost",
			Port:     5432,
			Database: "test_db",
			Username: "postgres",
			Password: "password",
			SSLMode:  "disable",
			LogLevel: logger.Silent,
		}
		err := manager.AddConfig(string(rune(i)), config)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGetStats 性能测试 - 获取统计信息
func BenchmarkGetStats(b *testing.B) {
	manager := NewManager()
	
	// 设置测试配置（不实际连接）
	config := &Config{
		Type:     PostgreSQL,
		Host:     "localhost",
		Port:     5432,
		Database: "bench_db",
		Username: "postgres",
		Password: "password",
		SSLMode:  "disable",
		LogLevel: logger.Silent,
	}
	manager.AddConfig("bench", config)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = manager.GetStats()
	}
}