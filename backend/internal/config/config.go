package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server
	Port        string
	Host        string
	Environment string

	// Database
	DatabaseType string
	DatabaseURL  string
	DatabaseName string

	// PostgreSQL specific (optional)
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBSSLMode     string
	DBSSLCert     string // SSL证书路径
	DBSSLKey      string // SSL私钥路径
	DBSSLRootCert string // CA证书路径

	// App
	AppName    string
	AppVersion string
	BaseURL    string

	// Security
	JWTSecret  string
	BCryptCost int

	// Frontend
	FrontendURL string

	// QR Code
	QRCodeStorePath string

	// AI
	OpenAIAPIKey      string
	ClaudeAPIKey      string
	SiliconFlowAPIKey string
	MoonshotAPIKey    string
	AIProvider        string

	// Email/SMTP
	SMTPHost         string
	SMTPPort         int
	SMTPUsername     string
	SMTPPassword     string
	EmailFromAddress string
	EmailFromName    string
	EmailProvider    string
	EmailAPIKey      string

	// Service Mesh
	EtcdEndpoints   string
	ConsulEndpoint  string
	JWTExpiry       int
	
	// Performance
	HighTrafficMode bool // 高流量模式
}

func Load() (*Config, error) {
	// Load .env file if exists
	_ = godotenv.Load()

	config := &Config{
		// Server defaults
		Port:        getEnv("PORT", "8080"),
		Host:        getEnv("HOST", "0.0.0.0"),
		Environment: getEnv("ENVIRONMENT", "development"),

		// Database defaults - 必须使用PostgreSQL
		DatabaseType: getEnv("DATABASE_TYPE", "postgres"),
		DatabaseURL:  getEnv("DATABASE_URL", "postgres://rocalight:password@localhost:5432/openpenpal"),
		DatabaseName: getEnv("DATABASE_NAME", "openpenpal"),

		// PostgreSQL specific
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "rocalight"),
		DBPassword: getEnv("DB_PASSWORD", "password"),
		DBSSLMode:     getEnv("DB_SSLMODE", "disable"),
		DBSSLCert:     getEnv("DB_SSL_CERT", ""),
		DBSSLKey:      getEnv("DB_SSL_KEY", ""),
		DBSSLRootCert: getEnv("DB_SSL_ROOT_CERT", ""),

		// App defaults
		AppName:    getEnv("APP_NAME", "OpenPenPal"),
		AppVersion: getEnv("APP_VERSION", "1.0.0"),
		BaseURL:    getEnv("BASE_URL", "http://localhost:8080"),

		// Security
		JWTSecret:  getEnvOrPanic("JWT_SECRET"),
		BCryptCost: getEnvAsInt("BCRYPT_COST", 12),

		// Frontend
		FrontendURL: getEnv("FRONTEND_URL", "http://localhost:3000"),

		// QR Code
		QRCodeStorePath: getEnv("QR_CODE_STORE_PATH", "./uploads/qrcodes"),

		// AI
		OpenAIAPIKey:      getEnv("OPENAI_API_KEY", ""),
		ClaudeAPIKey:      getEnv("CLAUDE_API_KEY", ""),
		SiliconFlowAPIKey: getEnv("SILICONFLOW_API_KEY", ""),
		MoonshotAPIKey:    getEnv("MOONSHOT_API_KEY", ""),
		AIProvider:        getEnv("AI_PROVIDER", "moonshot"),

		// Email/SMTP
		SMTPHost:         getEnv("SMTP_HOST", ""),
		SMTPPort:         getEnvAsInt("SMTP_PORT", 587),
		SMTPUsername:     getEnv("SMTP_USERNAME", ""),
		SMTPPassword:     getEnv("SMTP_PASSWORD", ""),
		EmailFromAddress: getEnv("EMAIL_FROM_ADDRESS", "noreply@openpenpal.com"),
		EmailFromName:    getEnv("EMAIL_FROM_NAME", "OpenPenPal"),
		EmailProvider:    getEnv("EMAIL_PROVIDER", "smtp"),
		EmailAPIKey:      getEnv("EMAIL_API_KEY", ""),

		// Service Mesh
		EtcdEndpoints:  getEnv("ETCD_ENDPOINTS", "localhost:2379"),
		ConsulEndpoint: getEnv("CONSUL_ENDPOINT", "localhost:8500"),
		JWTExpiry:      getEnvAsInt("JWT_EXPIRY", 24),
		
		// Performance
		HighTrafficMode: getEnv("HIGH_TRAFFIC_MODE", "false") == "true",
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvOrPanic(key string) string {
	value := os.Getenv(key)
	if value == "" {
		if key == "JWT_SECRET" && os.Getenv("ENVIRONMENT") == "development" {
			return "dev-secret-key-do-not-use-in-production"
		}
		panic("Required environment variable " + key + " is not set")
	}
	return value
}
