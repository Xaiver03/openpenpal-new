package config

import (
	"fmt"
	"runtime"
	"time"
)

// PoolConfig 数据库连接池配置
type PoolConfig struct {
	// 基础配置
	MaxOpenConns    int           `json:"max_open_conns" yaml:"max_open_conns"`       // 最大打开连接数
	MaxIdleConns    int           `json:"max_idle_conns" yaml:"max_idle_conns"`       // 最大空闲连接数
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime" yaml:"conn_max_lifetime"` // 连接最大生命周期
	ConnMaxIdleTime time.Duration `json:"conn_max_idle_time" yaml:"conn_max_idle_time"` // 连接最大空闲时间

	// 高级配置
	MinIdleConns        int           `json:"min_idle_conns" yaml:"min_idle_conns"`               // 最小空闲连接数
	HealthCheckInterval time.Duration `json:"health_check_interval" yaml:"health_check_interval"` // 健康检查间隔
	ConnectionTimeout   time.Duration `json:"connection_timeout" yaml:"connection_timeout"`       // 连接超时
	RetryAttempts       int           `json:"retry_attempts" yaml:"retry_attempts"`               // 重试次数
	RetryDelay          time.Duration `json:"retry_delay" yaml:"retry_delay"`                     // 重试延迟

	// 环境配置
	Environment string `json:"environment" yaml:"environment"`
}

// PoolPreset 预设连接池配置
type PoolPreset string

const (
	PoolPresetDevelopment PoolPreset = "development"
	PoolPresetTesting     PoolPreset = "testing"
	PoolPresetStaging     PoolPreset = "staging"
	PoolPresetProduction  PoolPreset = "production"
	PoolPresetHighTraffic PoolPreset = "high_traffic"
	PoolPresetLowLatency  PoolPreset = "low_latency"
	PoolPresetBatch       PoolPreset = "batch_processing"
)

// GetPoolPreset 获取预设配置
func GetPoolPreset(preset PoolPreset) *PoolConfig {
	switch preset {
	case PoolPresetDevelopment:
		return &PoolConfig{
			MaxOpenConns:        10,
			MaxIdleConns:        5,
			MinIdleConns:        1,
			ConnMaxLifetime:     time.Hour,
			ConnMaxIdleTime:     15 * time.Minute,
			HealthCheckInterval: 5 * time.Minute,
			ConnectionTimeout:   5 * time.Second,
			RetryAttempts:       3,
			RetryDelay:          100 * time.Millisecond,
			Environment:         "development",
		}

	case PoolPresetTesting:
		return &PoolConfig{
			MaxOpenConns:        5,
			MaxIdleConns:        2,
			MinIdleConns:        0,
			ConnMaxLifetime:     30 * time.Minute,
			ConnMaxIdleTime:     5 * time.Minute,
			HealthCheckInterval: 10 * time.Minute,
			ConnectionTimeout:   3 * time.Second,
			RetryAttempts:       2,
			RetryDelay:          50 * time.Millisecond,
			Environment:         "testing",
		}

	case PoolPresetStaging:
		return &PoolConfig{
			MaxOpenConns:        50,
			MaxIdleConns:        20,
			MinIdleConns:        5,
			ConnMaxLifetime:     30 * time.Minute,
			ConnMaxIdleTime:     10 * time.Minute,
			HealthCheckInterval: 2 * time.Minute,
			ConnectionTimeout:   10 * time.Second,
			RetryAttempts:       3,
			RetryDelay:          200 * time.Millisecond,
			Environment:         "staging",
		}

	case PoolPresetProduction:
		return &PoolConfig{
			MaxOpenConns:        100,
			MaxIdleConns:        30,
			MinIdleConns:        10,
			ConnMaxLifetime:     15 * time.Minute,
			ConnMaxIdleTime:     5 * time.Minute,
			HealthCheckInterval: 1 * time.Minute,
			ConnectionTimeout:   15 * time.Second,
			RetryAttempts:       5,
			RetryDelay:          500 * time.Millisecond,
			Environment:         "production",
		}

	case PoolPresetHighTraffic:
		return &PoolConfig{
			MaxOpenConns:        200,
			MaxIdleConns:        50,
			MinIdleConns:        20,
			ConnMaxLifetime:     10 * time.Minute,
			ConnMaxIdleTime:     3 * time.Minute,
			HealthCheckInterval: 30 * time.Second,
			ConnectionTimeout:   20 * time.Second,
			RetryAttempts:       5,
			RetryDelay:          1 * time.Second,
			Environment:         "high_traffic",
		}

	case PoolPresetLowLatency:
		return &PoolConfig{
			MaxOpenConns:        150,
			MaxIdleConns:        100,
			MinIdleConns:        50,
			ConnMaxLifetime:     5 * time.Minute,
			ConnMaxIdleTime:     2 * time.Minute,
			HealthCheckInterval: 20 * time.Second,
			ConnectionTimeout:   5 * time.Second,
			RetryAttempts:       2,
			RetryDelay:          50 * time.Millisecond,
			Environment:         "low_latency",
		}

	case PoolPresetBatch:
		return &PoolConfig{
			MaxOpenConns:        20,
			MaxIdleConns:        5,
			MinIdleConns:        2,
			ConnMaxLifetime:     2 * time.Hour,
			ConnMaxIdleTime:     30 * time.Minute,
			HealthCheckInterval: 10 * time.Minute,
			ConnectionTimeout:   30 * time.Second,
			RetryAttempts:       10,
			RetryDelay:          2 * time.Second,
			Environment:         "batch_processing",
		}

	default:
		return GetPoolPreset(PoolPresetDevelopment)
	}
}

// CalculateOptimalPoolSize 根据系统资源计算最优连接池大小
func CalculateOptimalPoolSize() *PoolConfig {
	// 获取CPU核心数
	numCPU := runtime.NumCPU()
	
	// 基础计算公式：
	// MaxOpenConns = CPU核心数 * 4 (每个核心4个连接)
	// MaxIdleConns = MaxOpenConns * 0.3 (30%的连接保持空闲)
	// MinIdleConns = MaxIdleConns * 0.3 (最少保持30%的空闲连接)
	
	maxOpen := numCPU * 4
	if maxOpen < 10 {
		maxOpen = 10 // 最小10个连接
	}
	if maxOpen > 200 {
		maxOpen = 200 // 最大200个连接
	}
	
	maxIdle := int(float64(maxOpen) * 0.3)
	if maxIdle < 5 {
		maxIdle = 5
	}
	
	minIdle := int(float64(maxIdle) * 0.3)
	if minIdle < 2 {
		minIdle = 2
	}
	
	return &PoolConfig{
		MaxOpenConns:        maxOpen,
		MaxIdleConns:        maxIdle,
		MinIdleConns:        minIdle,
		ConnMaxLifetime:     15 * time.Minute,
		ConnMaxIdleTime:     5 * time.Minute,
		HealthCheckInterval: 1 * time.Minute,
		ConnectionTimeout:   10 * time.Second,
		RetryAttempts:       3,
		RetryDelay:          200 * time.Millisecond,
		Environment:         "auto_calculated",
	}
}

// Validate 验证连接池配置
func (p *PoolConfig) Validate() error {
	if p.MaxOpenConns <= 0 {
		return fmt.Errorf("MaxOpenConns must be greater than 0")
	}
	
	if p.MaxIdleConns <= 0 {
		return fmt.Errorf("MaxIdleConns must be greater than 0")
	}
	
	if p.MaxIdleConns > p.MaxOpenConns {
		return fmt.Errorf("MaxIdleConns cannot be greater than MaxOpenConns")
	}
	
	if p.MinIdleConns < 0 {
		return fmt.Errorf("MinIdleConns cannot be negative")
	}
	
	if p.MinIdleConns > p.MaxIdleConns {
		return fmt.Errorf("MinIdleConns cannot be greater than MaxIdleConns")
	}
	
	if p.ConnMaxLifetime <= 0 {
		return fmt.Errorf("ConnMaxLifetime must be greater than 0")
	}
	
	if p.ConnMaxIdleTime <= 0 {
		return fmt.Errorf("ConnMaxIdleTime must be greater than 0")
	}
	
	if p.ConnectionTimeout <= 0 {
		return fmt.Errorf("ConnectionTimeout must be greater than 0")
	}
	
	return nil
}

// Optimize 根据监控数据优化配置
func (p *PoolConfig) Optimize(metrics *PoolMetrics) {
	// 如果等待连接的请求过多，增加最大连接数
	if metrics.WaitCount > int64(p.MaxOpenConns)*2 {
		p.MaxOpenConns = int(float64(p.MaxOpenConns) * 1.5)
		if p.MaxOpenConns > 300 {
			p.MaxOpenConns = 300
		}
	}
	
	// 如果空闲连接使用率低，减少最大空闲连接数
	idleRatio := float64(metrics.IdleConnections) / float64(p.MaxIdleConns)
	if idleRatio < 0.2 && p.MaxIdleConns > 10 {
		p.MaxIdleConns = int(float64(p.MaxIdleConns) * 0.8)
	}
	
	// 如果连接频繁创建和销毁，增加连接生命周期
	if metrics.ConnectionChurn > 0.5 {
		p.ConnMaxLifetime = time.Duration(float64(p.ConnMaxLifetime) * 1.5)
		if p.ConnMaxLifetime > 2*time.Hour {
			p.ConnMaxLifetime = 2 * time.Hour
		}
	}
	
	// 确保配置仍然有效
	if p.MaxIdleConns > p.MaxOpenConns {
		p.MaxIdleConns = int(float64(p.MaxOpenConns) * 0.3)
	}
}

// PoolMetrics 连接池监控指标
type PoolMetrics struct {
	OpenConnections   int           `json:"open_connections"`
	IdleConnections   int           `json:"idle_connections"`
	InUseConnections  int           `json:"in_use_connections"`
	WaitCount         int64         `json:"wait_count"`
	WaitDuration      time.Duration `json:"wait_duration"`
	MaxIdleClosed     int64         `json:"max_idle_closed"`
	MaxIdleTimeClosed int64         `json:"max_idle_time_closed"`
	MaxLifetimeClosed int64         `json:"max_lifetime_closed"`
	ConnectionChurn   float64       `json:"connection_churn"` // 连接创建/销毁率
}

// PoolManager 连接池管理器
type PoolManager struct {
	config  *PoolConfig
	metrics *PoolMetrics
}

// NewPoolManager 创建连接池管理器
func NewPoolManager(config *PoolConfig) *PoolManager {
	return &PoolManager{
		config:  config,
		metrics: &PoolMetrics{},
	}
}

// GetRecommendedConfig 根据使用场景获取推荐配置
func GetRecommendedConfig(scenario string) *PoolConfig {
	switch scenario {
	case "web_api":
		// Web API: 需要快速响应，中等并发
		return &PoolConfig{
			MaxOpenConns:        50,
			MaxIdleConns:        20,
			MinIdleConns:        5,
			ConnMaxLifetime:     30 * time.Minute,
			ConnMaxIdleTime:     10 * time.Minute,
			HealthCheckInterval: 2 * time.Minute,
			ConnectionTimeout:   5 * time.Second,
			RetryAttempts:       3,
			RetryDelay:          100 * time.Millisecond,
		}
		
	case "microservice":
		// 微服务: 高并发，短连接
		return &PoolConfig{
			MaxOpenConns:        100,
			MaxIdleConns:        30,
			MinIdleConns:        10,
			ConnMaxLifetime:     15 * time.Minute,
			ConnMaxIdleTime:     5 * time.Minute,
			HealthCheckInterval: 1 * time.Minute,
			ConnectionTimeout:   3 * time.Second,
			RetryAttempts:       5,
			RetryDelay:          50 * time.Millisecond,
		}
		
	case "analytics":
		// 分析系统: 长查询，低并发
		return &PoolConfig{
			MaxOpenConns:        20,
			MaxIdleConns:        10,
			MinIdleConns:        2,
			ConnMaxLifetime:     2 * time.Hour,
			ConnMaxIdleTime:     30 * time.Minute,
			HealthCheckInterval: 10 * time.Minute,
			ConnectionTimeout:   60 * time.Second,
			RetryAttempts:       2,
			RetryDelay:          1 * time.Second,
		}
		
	case "realtime":
		// 实时系统: 超低延迟，高并发
		return &PoolConfig{
			MaxOpenConns:        200,
			MaxIdleConns:        100,
			MinIdleConns:        50,
			ConnMaxLifetime:     5 * time.Minute,
			ConnMaxIdleTime:     2 * time.Minute,
			HealthCheckInterval: 30 * time.Second,
			ConnectionTimeout:   1 * time.Second,
			RetryAttempts:       1,
			RetryDelay:          10 * time.Millisecond,
		}
		
	default:
		return GetPoolPreset(PoolPresetProduction)
	}
}

// MonitoringRecommendations 监控建议
type MonitoringRecommendations struct {
	MetricName  string `json:"metric_name"`
	CurrentValue interface{} `json:"current_value"`
	Threshold   interface{} `json:"threshold"`
	Recommendation string `json:"recommendation"`
	Severity    string `json:"severity"` // low, medium, high
}

// AnalyzePoolPerformance 分析连接池性能并给出建议
func AnalyzePoolPerformance(config *PoolConfig, metrics *PoolMetrics) []MonitoringRecommendations {
	var recommendations []MonitoringRecommendations
	
	// 分析连接使用率
	usageRatio := float64(metrics.InUseConnections) / float64(config.MaxOpenConns)
	if usageRatio > 0.8 {
		recommendations = append(recommendations, MonitoringRecommendations{
			MetricName:    "Connection Usage",
			CurrentValue:  fmt.Sprintf("%.2f%%", usageRatio*100),
			Threshold:     "80%",
			Recommendation: fmt.Sprintf("High connection usage. Consider increasing MaxOpenConns from %d to %d", 
				config.MaxOpenConns, int(float64(config.MaxOpenConns)*1.5)),
			Severity:      "high",
		})
	}
	
	// 分析空闲连接
	if metrics.IdleConnections < config.MinIdleConns {
		recommendations = append(recommendations, MonitoringRecommendations{
			MetricName:    "Idle Connections",
			CurrentValue:  metrics.IdleConnections,
			Threshold:     config.MinIdleConns,
			Recommendation: "Idle connections below minimum. Connection pool may be under pressure",
			Severity:      "medium",
		})
	}
	
	// 分析等待时间
	if metrics.WaitDuration > 100*time.Millisecond {
		recommendations = append(recommendations, MonitoringRecommendations{
			MetricName:    "Wait Duration",
			CurrentValue:  metrics.WaitDuration,
			Threshold:     100 * time.Millisecond,
			Recommendation: "High wait times detected. Increase MaxOpenConns or optimize queries",
			Severity:      "high",
		})
	}
	
	// 分析连接流失率
	totalClosed := metrics.MaxIdleClosed + metrics.MaxIdleTimeClosed + metrics.MaxLifetimeClosed
	if totalClosed > int64(config.MaxOpenConns) {
		recommendations = append(recommendations, MonitoringRecommendations{
			MetricName:    "Connection Churn",
			CurrentValue:  totalClosed,
			Threshold:     config.MaxOpenConns,
			Recommendation: "High connection churn. Consider increasing ConnMaxLifetime and ConnMaxIdleTime",
			Severity:      "medium",
		})
	}
	
	return recommendations
}