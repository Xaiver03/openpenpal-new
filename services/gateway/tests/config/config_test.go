package config_test

import (
	"api-gateway/internal/config"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ConfigTestSuite struct {
	suite.Suite
}

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}

func (suite *ConfigTestSuite) SetupTest() {
	// Clean environment before each test
	os.Unsetenv("PORT")
	os.Unsetenv("ENVIRONMENT") 
	os.Unsetenv("JWT_SECRET")
	os.Unsetenv("REDIS_URL")
	os.Unsetenv("DATABASE_URL")
	os.Unsetenv("LOG_LEVEL")
	os.Unsetenv("METRICS_ENABLED")
	os.Unsetenv("METRICS_PORT")
	os.Unsetenv("PROXY_TIMEOUT")
	os.Unsetenv("CONNECT_TIMEOUT")
	os.Unsetenv("KEEPALIVE_TIMEOUT")
}

func (suite *ConfigTestSuite) TestLoadConfig() {
	suite.Run("Load config with default values", func() {
		cfg := config.LoadConfig()
		
		assert.NotNil(suite.T(), cfg)
		assert.Equal(suite.T(), "8000", cfg.Port)
		assert.Equal(suite.T(), "development", cfg.Environment)
		assert.Equal(suite.T(), "info", cfg.LogLevel)
		assert.NotEmpty(suite.T(), cfg.JWTSecret)
		assert.NotNil(suite.T(), cfg.Services)
		assert.True(suite.T(), cfg.MetricsEnabled)
		assert.Equal(suite.T(), "9090", cfg.MetricsPort)
	})

	suite.Run("Load config with environment variables", func() {
		os.Setenv("PORT", "9000")
		os.Setenv("ENVIRONMENT", "production")
		os.Setenv("JWT_SECRET", "test-secret-123")
		os.Setenv("LOG_LEVEL", "debug")
		os.Setenv("REDIS_URL", "redis://localhost:6380")
		os.Setenv("DATABASE_URL", "postgres://test:test@localhost:5433/testdb")
		os.Setenv("METRICS_ENABLED", "false")
		os.Setenv("METRICS_PORT", "9091")
		os.Setenv("PROXY_TIMEOUT", "45")
		os.Setenv("CONNECT_TIMEOUT", "15")
		os.Setenv("KEEPALIVE_TIMEOUT", "120")

		cfg := config.LoadConfig()
		
		assert.Equal(suite.T(), "9000", cfg.Port)
		assert.Equal(suite.T(), "production", cfg.Environment)
		assert.Equal(suite.T(), "test-secret-123", cfg.JWTSecret)
		assert.Equal(suite.T(), "debug", cfg.LogLevel)
		assert.Equal(suite.T(), "redis://localhost:6380", cfg.RedisURL)
		assert.Equal(suite.T(), "postgres://test:test@localhost:5433/testdb", cfg.DatabaseURL)
		assert.False(suite.T(), cfg.MetricsEnabled)
		assert.Equal(suite.T(), "9091", cfg.MetricsPort)
		assert.Equal(suite.T(), 45, cfg.ProxyTimeout)
		assert.Equal(suite.T(), 15, cfg.ConnectTimeout)
		assert.Equal(suite.T(), 120, cfg.KeepAliveTimeout)
	})
}

func (suite *ConfigTestSuite) TestConfigValidation() {
	suite.Run("Valid configuration", func() {
		cfg := &config.Config{
			Port:        "8000",
			Environment: "development",
			JWTSecret:   "valid-secret-at-least-32-chars-long",
			Services:    make(map[string]*config.ServiceConfig),
		}

		err := cfg.Validate()
		assert.NoError(suite.T(), err)
	})

	suite.Run("Invalid port", func() {
		cfg := &config.Config{
			Port:        "invalid-port",
			Environment: "development",
			JWTSecret:   "valid-secret-at-least-32-chars-long",
		}

		err := cfg.Validate()
		assert.Error(suite.T(), err)
		assert.Contains(suite.T(), err.Error(), "invalid port")
	})

	suite.Run("Missing JWT secret", func() {
		cfg := &config.Config{
			Port:        "8000",
			Environment: "development",
			JWTSecret:   "",
		}

		err := cfg.Validate()
		assert.Error(suite.T(), err)
		assert.Contains(suite.T(), err.Error(), "JWT secret")
	})

	suite.Run("JWT secret too short", func() {
		cfg := &config.Config{
			Port:        "8000",
			Environment: "development",
			JWTSecret:   "short",
		}

		err := cfg.Validate()
		assert.Error(suite.T(), err)
		assert.Contains(suite.T(), err.Error(), "JWT secret must be at least")
	})
}

func (suite *ConfigTestSuite) TestEnvironmentMethods() {
	tests := []struct {
		environment string
		isDev       bool
		isProd      bool
		isTest      bool
	}{
		{"development", true, false, false},
		{"production", false, true, false},
		{"test", false, false, true},
		{"testing", false, false, true},
		{"staging", false, false, false},
		{"", true, false, false}, // Default to development
	}

	for _, tt := range tests {
		suite.Run("Environment: "+tt.environment, func() {
			cfg := &config.Config{Environment: tt.environment}
			
			assert.Equal(suite.T(), tt.isDev, cfg.IsDevelopment())
			assert.Equal(suite.T(), tt.isProd, cfg.IsProduction())
			assert.Equal(suite.T(), tt.isTest, cfg.IsTest())
		})
	}
}

func (suite *ConfigTestSuite) TestServiceConfig() {
	suite.Run("Valid service configuration", func() {
		serviceConfig := &config.ServiceConfig{
			Name:        "test-service",
			Hosts:       []string{"localhost:8080", "localhost:8081"},
			HealthCheck: "/health",
			Timeout:     30,
			Retries:     3,
			Weight:      100,
		}

		err := serviceConfig.Validate()
		assert.NoError(suite.T(), err)
	})

	suite.Run("Invalid service config - missing name", func() {
		serviceConfig := &config.ServiceConfig{
			Hosts:       []string{"localhost:8080"},
			HealthCheck: "/health",
			Timeout:     30,
		}

		err := serviceConfig.Validate()
		assert.Error(suite.T(), err)
		assert.Contains(suite.T(), err.Error(), "service name")
	})

	suite.Run("Invalid service config - no hosts", func() {
		serviceConfig := &config.ServiceConfig{
			Name:        "test-service",
			Hosts:       []string{},
			HealthCheck: "/health",
			Timeout:     30,
		}

		err := serviceConfig.Validate()
		assert.Error(suite.T(), err)
		assert.Contains(suite.T(), err.Error(), "at least one host")
	})

	suite.Run("Invalid service config - invalid timeout", func() {
		serviceConfig := &config.ServiceConfig{
			Name:        "test-service",
			Hosts:       []string{"localhost:8080"},
			HealthCheck: "/health",
			Timeout:     0,
		}

		err := serviceConfig.Validate()
		assert.Error(suite.T(), err)
		assert.Contains(suite.T(), err.Error(), "timeout must be positive")
	})
}

func (suite *ConfigTestSuite) TestRateLimitConfig() {
	suite.Run("Default rate limit configuration", func() {
		cfg := config.LoadConfig()
		
		assert.NotNil(suite.T(), cfg.RateLimitConfig)
		assert.True(suite.T(), cfg.RateLimitConfig.Enabled)
		assert.Greater(suite.T(), cfg.RateLimitConfig.DefaultLimit, 0)
		assert.NotNil(suite.T(), cfg.RateLimitConfig.ServiceLimits)
	})

	suite.Run("Rate limit per service", func() {
		cfg := config.LoadConfig()
		
		// Test default service limits
		if ocr, exists := cfg.RateLimitConfig.ServiceLimits["ocr"]; exists {
			assert.Equal(suite.T(), 20, ocr) // OCR should have lower limit
		}
		
		if backend, exists := cfg.RateLimitConfig.ServiceLimits["backend"]; exists {
			assert.Greater(suite.T(), backend, 50) // Backend should have higher limit
		}
	})
}

func (suite *ConfigTestSuite) TestGetServiceConfig() {
	suite.Run("Get existing service config", func() {
		cfg := config.LoadConfig()
		
		serviceConfig := cfg.GetServiceConfig("backend")
		if serviceConfig != nil {
			assert.Equal(suite.T(), "backend", serviceConfig.Name)
			assert.NotEmpty(suite.T(), serviceConfig.Hosts)
			assert.NotEmpty(suite.T(), serviceConfig.HealthCheck)
		}
	})

	suite.Run("Get non-existent service config", func() {
		cfg := config.LoadConfig()
		
		serviceConfig := cfg.GetServiceConfig("non-existent-service")
		assert.Nil(suite.T(), serviceConfig)
	})
}

func (suite *ConfigTestSuite) TestConfigDefaults() {
	suite.Run("Default microservice configurations", func() {
		cfg := config.LoadConfig()
		
		// Should have default services configured
		expectedServices := []string{"backend", "write-service", "courier-service", "admin-service", "ocr-service"}
		
		for _, serviceName := range expectedServices {
			serviceConfig := cfg.GetServiceConfig(serviceName)
			if serviceConfig != nil {
				assert.Equal(suite.T(), serviceName, serviceConfig.Name)
				assert.NotEmpty(suite.T(), serviceConfig.Hosts)
				assert.NotEmpty(suite.T(), serviceConfig.HealthCheck)
				assert.Greater(suite.T(), serviceConfig.Timeout, 0)
				assert.GreaterOrEqual(suite.T(), serviceConfig.Retries, 0)
				assert.Greater(suite.T(), serviceConfig.Weight, 0)
			}
		}
	})
}

func (suite *ConfigTestSuite) TestConfigSerialization() {
	suite.Run("Config to JSON", func() {
		cfg := &config.Config{
			Port:        "8000",
			Environment: "test",
			JWTSecret:   "test-secret-for-serialization-testing",
			Services:    make(map[string]*config.ServiceConfig),
		}

		jsonBytes, err := cfg.ToJSON()
		assert.NoError(suite.T(), err)
		assert.NotEmpty(suite.T(), jsonBytes)
		
		// JWT secret should be masked in JSON output
		jsonStr := string(jsonBytes)
		assert.NotContains(suite.T(), jsonStr, "test-secret-for-serialization-testing")
		assert.Contains(suite.T(), jsonStr, "****")
	})

	suite.Run("Config from JSON", func() {
		jsonData := `{
			"port": "9000",
			"environment": "test",
			"jwt_secret": "test-secret-from-json-data-import",
			"log_level": "debug",
			"services": {},
			"metrics_enabled": true,
			"metrics_port": "9090"
		}`

		cfg, err := config.FromJSON([]byte(jsonData))
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), "9000", cfg.Port)
		assert.Equal(suite.T(), "test", cfg.Environment)
		assert.Equal(suite.T(), "test-secret-from-json-data-import", cfg.JWTSecret)
		assert.Equal(suite.T(), "debug", cfg.LogLevel)
		assert.True(suite.T(), cfg.MetricsEnabled)
		assert.Equal(suite.T(), "9090", cfg.MetricsPort)
	})
}

func (suite *ConfigTestSuite) TestConfigUpdate() {
	suite.Run("Update service configuration", func() {
		cfg := config.LoadConfig()
		
		newServiceConfig := &config.ServiceConfig{
			Name:        "new-service",
			Hosts:       []string{"localhost:9999"},
			HealthCheck: "/health",
			Timeout:     45,
			Retries:     2,
			Weight:      50,
		}

		cfg.AddService("new-service", newServiceConfig)
		
		retrievedConfig := cfg.GetServiceConfig("new-service")
		assert.NotNil(suite.T(), retrievedConfig)
		assert.Equal(suite.T(), newServiceConfig.Name, retrievedConfig.Name)
		assert.Equal(suite.T(), newServiceConfig.Hosts, retrievedConfig.Hosts)
		assert.Equal(suite.T(), newServiceConfig.Timeout, retrievedConfig.Timeout)
	})

	suite.Run("Remove service configuration", func() {
		cfg := config.LoadConfig()
		
		// Add a service first
		cfg.AddService("temp-service", &config.ServiceConfig{
			Name:  "temp-service",
			Hosts: []string{"localhost:8888"},
		})
		
		// Verify it exists
		assert.NotNil(suite.T(), cfg.GetServiceConfig("temp-service"))
		
		// Remove it
		cfg.RemoveService("temp-service")
		
		// Verify it's gone
		assert.Nil(suite.T(), cfg.GetServiceConfig("temp-service"))
	})
}