package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config 网关配置结构
type Config struct {
	// 服务器配置
	Port        string
	Environment string
	LogLevel    string

	// 认证配置
	JWTSecret string

	// 服务发现配置
	Services map[string]*ServiceConfig

	// Redis配置
	RedisURL string

	// 数据库配置
	DatabaseURL string

	// 限流配置
	RateLimitConfig *RateLimitConfig

	// 监控配置
	MetricsEnabled bool
	MetricsPort    string

	// 超时配置
	ProxyTimeout     int // 秒
	ConnectTimeout   int // 秒
	KeepAliveTimeout int // 秒
}

// ServiceConfig 微服务配置
type ServiceConfig struct {
	Name        string   `json:"name"`
	Hosts       []string `json:"hosts"`        // 支持多实例
	HealthCheck string   `json:"health_check"` // 健康检查路径
	Timeout     int      `json:"timeout"`      // 超时时间(秒)
	Retries     int      `json:"retries"`      // 重试次数
	Weight      int      `json:"weight"`       // 权重(负载均衡)
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	Enabled         bool
	DefaultLimit    int // 每分钟请求数
	BurstSize       int // 突发请求数
	WindowSize      int // 时间窗口(秒)
	CleanupInterval int // 清理间隔(秒)
}

// Load 加载配置
func Load() *Config {
	// 加载 .env 文件
	godotenv.Load()

	return &Config{
		Port:        getEnv("PORT", "8000"),
		Environment: getEnv("ENVIRONMENT", "development"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		JWTSecret:   getEnv("JWT_SECRET", "your-super-secret-jwt-key"),
		RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379/0"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://rocalight:password@localhost:5432/openpenpal?sslmode=disable"),

		Services: loadServiceConfigs(),

		RateLimitConfig: &RateLimitConfig{
			Enabled:         getBoolEnv("RATE_LIMIT_ENABLED", true),
			DefaultLimit:    getIntEnv("RATE_LIMIT_DEFAULT", 100),
			BurstSize:       getIntEnv("RATE_LIMIT_BURST", 10),
			WindowSize:      getIntEnv("RATE_LIMIT_WINDOW", 60),
			CleanupInterval: getIntEnv("RATE_LIMIT_CLEANUP", 300),
		},

		MetricsEnabled:   getBoolEnv("METRICS_ENABLED", true),
		MetricsPort:      getEnv("METRICS_PORT", "9000"),
		ProxyTimeout:     getIntEnv("PROXY_TIMEOUT", 30),
		ConnectTimeout:   getIntEnv("CONNECT_TIMEOUT", 5),
		KeepAliveTimeout: getIntEnv("KEEPALIVE_TIMEOUT", 30),
	}
}

// loadServiceConfigs 加载微服务配置
func loadServiceConfigs() map[string]*ServiceConfig {
	services := make(map[string]*ServiceConfig)

	// 主后端服务 (原有的Go后端)
	services["main-backend"] = &ServiceConfig{
		Name:        "main-backend",
		Hosts:       getHostList("MAIN_BACKEND_HOSTS", "http://localhost:8080"),
		HealthCheck: "/health",
		Timeout:     getIntEnv("MAIN_BACKEND_TIMEOUT", 30),
		Retries:     getIntEnv("MAIN_BACKEND_RETRIES", 3),
		Weight:      getIntEnv("MAIN_BACKEND_WEIGHT", 10),
	}

	// 写信服务
	services["write-service"] = &ServiceConfig{
		Name:        "write-service",
		Hosts:       getHostList("WRITE_SERVICE_HOSTS", "http://localhost:8001"),
		HealthCheck: "/health",
		Timeout:     getIntEnv("WRITE_SERVICE_TIMEOUT", 30),
		Retries:     getIntEnv("WRITE_SERVICE_RETRIES", 3),
		Weight:      getIntEnv("WRITE_SERVICE_WEIGHT", 10),
	}

	// 信使服务
	services["courier-service"] = &ServiceConfig{
		Name:        "courier-service",
		Hosts:       getHostList("COURIER_SERVICE_HOSTS", "http://localhost:8002"),
		HealthCheck: "/health",
		Timeout:     getIntEnv("COURIER_SERVICE_TIMEOUT", 30),
		Retries:     getIntEnv("COURIER_SERVICE_RETRIES", 3),
		Weight:      getIntEnv("COURIER_SERVICE_WEIGHT", 10),
	}

	// 管理服务
	services["admin-service"] = &ServiceConfig{
		Name:        "admin-service",
		Hosts:       getHostList("ADMIN_SERVICE_HOSTS", "http://localhost:8003"),
		HealthCheck: "/health",
		Timeout:     getIntEnv("ADMIN_SERVICE_TIMEOUT", 30),
		Retries:     getIntEnv("ADMIN_SERVICE_RETRIES", 3),
		Weight:      getIntEnv("ADMIN_SERVICE_WEIGHT", 10),
	}

	// OCR服务
	services["ocr-service"] = &ServiceConfig{
		Name:        "ocr-service",
		Hosts:       getHostList("OCR_SERVICE_HOSTS", "http://localhost:8004"),
		HealthCheck: "/health",
		Timeout:     getIntEnv("OCR_SERVICE_TIMEOUT", 60), // OCR处理时间较长
		Retries:     getIntEnv("OCR_SERVICE_RETRIES", 2),
		Weight:      getIntEnv("OCR_SERVICE_WEIGHT", 5),
	}

	return services
}

// getEnv 获取环境变量
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getIntEnv 获取整数环境变量
func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getBoolEnv 获取布尔环境变量
func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// getHostList 获取主机列表
func getHostList(key, defaultValue string) []string {
	value := getEnv(key, defaultValue)
	// 支持逗号分隔的多个主机
	hosts := strings.Split(value, ",")
	for i, host := range hosts {
		hosts[i] = strings.TrimSpace(host)
	}
	return hosts
}

// GetServiceConfig 获取指定服务的配置
func (c *Config) GetServiceConfig(serviceName string) *ServiceConfig {
	return c.Services[serviceName]
}

// IsProduction 是否为生产环境
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// IsDevelopment 是否为开发环境
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}
