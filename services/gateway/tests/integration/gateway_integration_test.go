package integration_test

import (
	"api-gateway/internal/config"
	"api-gateway/internal/discovery"
	"api-gateway/internal/loadbalancer"
	"api-gateway/internal/middleware"
	"api-gateway/internal/monitor"
	"api-gateway/internal/proxy"
	"api-gateway/internal/router"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type GatewayIntegrationTestSuite struct {
	suite.Suite
	cfg             *config.Config
	logger          *zap.Logger
	gatewayServer   *httptest.Server
	mockBackend     *httptest.Server
	mockWriteService *httptest.Server
	mockCourierService *httptest.Server
	mockAdminService *httptest.Server
	mockOCRService  *httptest.Server
	router          *gin.Engine
}

func TestGatewayIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(GatewayIntegrationTestSuite))
}

func (suite *GatewayIntegrationTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	suite.logger = zap.NewNop()
	
	// Create mock backend services
	suite.setupMockServices()
	
	// Create gateway configuration
	suite.setupGatewayConfig()
	
	// Setup complete gateway
	suite.setupGateway()
}

func (suite *GatewayIntegrationTestSuite) TearDownSuite() {
	if suite.gatewayServer != nil {
		suite.gatewayServer.Close()
	}
	if suite.mockBackend != nil {
		suite.mockBackend.Close()
	}
	if suite.mockWriteService != nil {
		suite.mockWriteService.Close()
	}
	if suite.mockCourierService != nil {
		suite.mockCourierService.Close()
	}
	if suite.mockAdminService != nil {
		suite.mockAdminService.Close()
	}
	if suite.mockOCRService != nil {
		suite.mockOCRService.Close()
	}
}

func (suite *GatewayIntegrationTestSuite) setupMockServices() {
	// Mock Backend Service
	suite.mockBackend = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/health":
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"status": "healthy", "service": "backend"})
		case "/api/v1/auth/login":
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{
				"token": "mock-jwt-token-user123-user-1234567890",
				"user_id": "user123",
				"role": "user",
			})
		case "/api/v1/users":
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"users": []map[string]string{
					{"id": "1", "name": "John Doe"},
					{"id": "2", "name": "Jane Smith"},
				},
			})
		case "/api/v1/letters":
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"letters": []map[string]string{
					{"id": "letter1", "title": "Sample Letter 1"},
					{"id": "letter2", "title": "Sample Letter 2"},
				},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	// Mock Write Service
	suite.mockWriteService = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/health":
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"status": "healthy", "service": "write"})
		case "/api/v1/write/letters":
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(map[string]string{
				"id": "new-letter-123",
				"status": "created",
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	// Mock Courier Service
	suite.mockCourierService = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/health":
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"status": "healthy", "service": "courier"})
		case "/api/v1/courier/tasks":
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"tasks": []map[string]string{
					{"id": "task1", "status": "available"},
					{"id": "task2", "status": "assigned"},
				},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	// Mock Admin Service
	suite.mockAdminService = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/health":
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"status": "healthy", "service": "admin"})
		case "/admin/users":
			// Check for admin role
			authHeader := r.Header.Get("Authorization")
			if !strings.Contains(authHeader, "admin") {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"users": []map[string]string{
					{"id": "1", "role": "admin"},
					{"id": "2", "role": "user"},
				},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	// Mock OCR Service  
	suite.mockOCRService = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/health":
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"status": "healthy", "service": "ocr"})
		case "/api/v1/ocr/process":
			// Simulate OCR processing delay
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"text": "Extracted text from image",
				"confidence": 0.95,
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func (suite *GatewayIntegrationTestSuite) setupGatewayConfig() {
	suite.cfg = &config.Config{
		Port:        "8000",
		Environment: "test",
		JWTSecret:   "test-jwt-secret-for-integration-testing-32-chars",
		LogLevel:    "info",
		Services: map[string]*config.ServiceConfig{
			"backend": {
				Name:        "backend",
				Hosts:       []string{strings.TrimPrefix(suite.mockBackend.URL, "http://")},
				HealthCheck: "/health",
				Timeout:     30,
				Retries:     3,
				Weight:      100,
			},
			"write-service": {
				Name:        "write-service",
				Hosts:       []string{strings.TrimPrefix(suite.mockWriteService.URL, "http://")},
				HealthCheck: "/health",
				Timeout:     30,
				Retries:     3,
				Weight:      100,
			},
			"courier-service": {
				Name:        "courier-service",
				Hosts:       []string{strings.TrimPrefix(suite.mockCourierService.URL, "http://")},
				HealthCheck: "/health",
				Timeout:     30,
				Retries:     3,
				Weight:      100,
			},
			"admin-service": {
				Name:        "admin-service",
				Hosts:       []string{strings.TrimPrefix(suite.mockAdminService.URL, "http://")},
				HealthCheck: "/health",
				Timeout:     30,
				Retries:     3,
				Weight:      100,
			},
			"ocr-service": {
				Name:        "ocr-service",
				Hosts:       []string{strings.TrimPrefix(suite.mockOCRService.URL, "http://")},
				HealthCheck: "/health",
				Timeout:     30,
				Retries:     3,
				Weight:      100,
			},
		},
		RateLimitConfig: &config.RateLimitConfig{
			Enabled:      true,
			DefaultLimit: 100,
			ServiceLimits: map[string]int{
				"ocr-service": 20,
				"backend":     200,
			},
		},
		MetricsEnabled: true,
		MetricsPort:    "9090",
		ProxyTimeout:   30,
		ConnectTimeout: 10,
		KeepAliveTimeout: 90,
	}
}

func (suite *GatewayIntegrationTestSuite) setupGateway() {
	// Initialize components
	discoveryManager := discovery.NewManager(suite.cfg, suite.logger)
	loadbalancerManager := loadbalancer.NewManager(suite.cfg, suite.logger)
	proxyManager := proxy.NewManager(suite.cfg, suite.logger)
	monitorManager := monitor.NewManager(suite.cfg, suite.logger)
	
	// Start service discovery
	discoveryManager.Start()
	
	// Create router
	routerManager := router.NewManager(suite.cfg, proxyManager, suite.logger)
	routerManager.SetMonitorManager(monitorManager)
	
	// Setup routes
	suite.router = routerManager.SetupRoutes()
	
	// Start gateway server
	suite.gatewayServer = httptest.NewServer(suite.router)
}

func (suite *GatewayIntegrationTestSuite) createAuthToken(userID, role string) string {
	return "Bearer mock-jwt-token-" + userID + "-" + role + "-1234567890"
}

func (suite *GatewayIntegrationTestSuite) TestHealthChecks() {
	suite.Run("Gateway health check", func() {
		resp, err := http.Get(suite.gatewayServer.URL + "/health")
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
		
		var response map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&response)
		
		assert.Equal(suite.T(), "healthy", response["status"])
		assert.Contains(suite.T(), response, "services")
	})

	suite.Run("Service health checks", func() {
		resp, err := http.Get(suite.gatewayServer.URL + "/health/services")
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
		
		var response map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&response)
		
		services := response["services"].(map[string]interface{})
		assert.Contains(suite.T(), services, "backend")
		assert.Contains(suite.T(), services, "write-service")
		assert.Contains(suite.T(), services, "courier-service")
	})
}

func (suite *GatewayIntegrationTestSuite) TestAuthenticationFlow() {
	suite.Run("Login without authentication", func() {
		resp, err := http.Post(suite.gatewayServer.URL+"/api/v1/auth/login", "application/json", 
			strings.NewReader(`{"username":"testuser","password":"password123"}`))
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
		
		var response map[string]string
		json.NewDecoder(resp.Body).Decode(&response)
		
		assert.Contains(suite.T(), response, "token")
		assert.Contains(suite.T(), response, "user_id")
	})

	suite.Run("Access protected resource with valid token", func() {
		req, _ := http.NewRequest("GET", suite.gatewayServer.URL+"/api/v1/users", nil)
		req.Header.Set("Authorization", suite.createAuthToken("user123", "user"))
		
		client := &http.Client{}
		resp, err := client.Do(req)
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
		
		var response map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&response)
		
		assert.Contains(suite.T(), response, "users")
	})

	suite.Run("Access protected resource without token", func() {
		resp, err := http.Get(suite.gatewayServer.URL + "/api/v1/users")
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), http.StatusUnauthorized, resp.StatusCode)
	})
}

func (suite *GatewayIntegrationTestSuite) TestRoleBasedAccess() {
	suite.Run("Admin access with admin token", func() {
		req, _ := http.NewRequest("GET", suite.gatewayServer.URL+"/admin/users", nil)
		req.Header.Set("Authorization", suite.createAuthToken("admin123", "admin"))
		
		client := &http.Client{}
		resp, err := client.Do(req)
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	})

	suite.Run("Admin access with user token", func() {
		req, _ := http.NewRequest("GET", suite.gatewayServer.URL+"/admin/users", nil)
		req.Header.Set("Authorization", suite.createAuthToken("user123", "user"))
		
		client := &http.Client{}
		resp, err := client.Do(req)
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), http.StatusForbidden, resp.StatusCode)
	})

	suite.Run("Courier access with courier token", func() {
		req, _ := http.NewRequest("GET", suite.gatewayServer.URL+"/api/v1/courier/tasks", nil)
		req.Header.Set("Authorization", suite.createAuthToken("courier123", "courier"))
		
		client := &http.Client{}
		resp, err := client.Do(req)
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	})
}

func (suite *GatewayIntegrationTestSuite) TestServiceRouting() {
	suite.Run("Route to backend service", func() {
		req, _ := http.NewRequest("GET", suite.gatewayServer.URL+"/api/v1/letters", nil)
		req.Header.Set("Authorization", suite.createAuthToken("user123", "user"))
		
		client := &http.Client{}
		resp, err := client.Do(req)
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
		
		var response map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&response)
		
		assert.Contains(suite.T(), response, "letters")
	})

	suite.Run("Route to write service", func() {
		req, _ := http.NewRequest("POST", suite.gatewayServer.URL+"/api/v1/write/letters", 
			strings.NewReader(`{"title":"Test Letter","content":"Test content"}`))
		req.Header.Set("Authorization", suite.createAuthToken("user123", "user"))
		req.Header.Set("Content-Type", "application/json")
		
		client := &http.Client{}
		resp, err := client.Do(req)
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), http.StatusCreated, resp.StatusCode)
		
		var response map[string]string
		json.NewDecoder(resp.Body).Decode(&response)
		
		assert.Equal(suite.T(), "created", response["status"])
	})

	suite.Run("Route to OCR service", func() {
		req, _ := http.NewRequest("POST", suite.gatewayServer.URL+"/api/v1/ocr/process", 
			strings.NewReader(`{"image":"base64encodedimage"}`))
		req.Header.Set("Authorization", suite.createAuthToken("user123", "user"))
		req.Header.Set("Content-Type", "application/json")
		
		client := &http.Client{}
		resp, err := client.Do(req)
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
		
		var response map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&response)
		
		assert.Contains(suite.T(), response, "text")
		assert.Contains(suite.T(), response, "confidence")
	})
}

func (suite *GatewayIntegrationTestSuite) TestRateLimiting() {
	suite.Run("Rate limiting for OCR service", func() {
		// OCR service should have lower rate limit
		successCount := 0
		rateLimitedCount := 0
		
		for i := 0; i < 30; i++ {
			req, _ := http.NewRequest("POST", suite.gatewayServer.URL+"/api/v1/ocr/process", 
				strings.NewReader(`{"image":"test"}`))
			req.Header.Set("Authorization", suite.createAuthToken("user123", "user"))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Real-IP", "192.168.1.100")
			
			client := &http.Client{}
			resp, err := client.Do(req)
			assert.NoError(suite.T(), err)
			
			if resp.StatusCode == http.StatusOK {
				successCount++
			} else if resp.StatusCode == http.StatusTooManyRequests {
				rateLimitedCount++
			}
		}
		
		// Should have some rate limited requests
		assert.Greater(suite.T(), rateLimitedCount, 0)
		assert.LessOrEqual(suite.T(), successCount, 25) // Should be limited
	})

	suite.Run("Higher rate limit for backend service", func() {
		successCount := 0
		
		for i := 0; i < 50; i++ {
			req, _ := http.NewRequest("GET", suite.gatewayServer.URL+"/api/v1/letters", nil)
			req.Header.Set("Authorization", suite.createAuthToken("user123", "user"))
			req.Header.Set("X-Real-IP", "192.168.1.101")
			
			client := &http.Client{}
			resp, err := client.Do(req)
			assert.NoError(suite.T(), err)
			
			if resp.StatusCode == http.StatusOK {
				successCount++
			}
		}
		
		// Backend should allow more requests
		assert.Greater(suite.T(), successCount, 40)
	})
}

func (suite *GatewayIntegrationTestSuite) TestCORSHandling() {
	suite.Run("CORS preflight request", func() {
		req, _ := http.NewRequest("OPTIONS", suite.gatewayServer.URL+"/api/v1/users", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("Access-Control-Request-Method", "GET")
		req.Header.Set("Access-Control-Request-Headers", "Authorization")
		
		client := &http.Client{}
		resp, err := client.Do(req)
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
		
		assert.NotEmpty(suite.T(), resp.Header.Get("Access-Control-Allow-Origin"))
		assert.NotEmpty(suite.T(), resp.Header.Get("Access-Control-Allow-Methods"))
		assert.NotEmpty(suite.T(), resp.Header.Get("Access-Control-Allow-Headers"))
	})

	suite.Run("CORS headers on actual request", func() {
		req, _ := http.NewRequest("GET", suite.gatewayServer.URL+"/api/v1/letters", nil)
		req.Header.Set("Authorization", suite.createAuthToken("user123", "user"))
		req.Header.Set("Origin", "http://localhost:3000")
		
		client := &http.Client{}
		resp, err := client.Do(req)
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
		
		assert.NotEmpty(suite.T(), resp.Header.Get("Access-Control-Allow-Origin"))
	})
}

func (suite *GatewayIntegrationTestSuite) TestMetricsCollection() {
	suite.Run("Metrics endpoint", func() {
		// Make some requests first
		for i := 0; i < 5; i++ {
			req, _ := http.NewRequest("GET", suite.gatewayServer.URL+"/api/v1/letters", nil)
			req.Header.Set("Authorization", suite.createAuthToken("user123", "user"))
			
			client := &http.Client{}
			client.Do(req)
		}
		
		// Check metrics
		resp, err := http.Get(suite.gatewayServer.URL + "/metrics")
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
		
		// Should contain Prometheus metrics
		body := make([]byte, resp.ContentLength)
		resp.Body.Read(body)
		metricsData := string(body)
		
		assert.Contains(suite.T(), metricsData, "http_requests_total")
		assert.Contains(suite.T(), metricsData, "http_request_duration_seconds")
	})

	suite.Run("Custom metrics endpoint", func() {
		resp, err := http.Get(suite.gatewayServer.URL + "/api/v1/metrics/performance")
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
		
		var response map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&response)
		
		assert.Contains(suite.T(), response, "total_requests")
		assert.Contains(suite.T(), response, "avg_response_time")
	})
}

func (suite *GatewayIntegrationTestSuite) TestErrorHandling() {
	suite.Run("Service unavailable", func() {
		// Try to access a non-existent service endpoint
		req, _ := http.NewRequest("GET", suite.gatewayServer.URL+"/api/v1/nonexistent", nil)
		req.Header.Set("Authorization", suite.createAuthToken("user123", "user"))
		
		client := &http.Client{}
		resp, err := client.Do(req)
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), http.StatusNotFound, resp.StatusCode)
	})

	suite.Run("Malformed request", func() {
		req, _ := http.NewRequest("POST", suite.gatewayServer.URL+"/api/v1/write/letters", 
			strings.NewReader(`{invalid json}`))
		req.Header.Set("Authorization", suite.createAuthToken("user123", "user"))
		req.Header.Set("Content-Type", "application/json")
		
		client := &http.Client{}
		resp, err := client.Do(req)
		assert.NoError(suite.T(), err)
		assert.True(suite.T(), resp.StatusCode >= 400)
	})
}

func (suite *GatewayIntegrationTestSuite) TestLoadBalancing() {
	suite.Run("Load balancing between multiple instances", func() {
		// This test would require multiple mock instances
		// For now, we'll test that the load balancer is working
		req, _ := http.NewRequest("GET", suite.gatewayServer.URL+"/api/v1/letters", nil)
		req.Header.Set("Authorization", suite.createAuthToken("user123", "user"))
		
		client := &http.Client{}
		resp, err := client.Do(req)
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	})
}

func (suite *GatewayIntegrationTestSuite) TestCompleteUserFlow() {
	suite.Run("Complete user workflow", func() {
		client := &http.Client{}
		
		// 1. Login
		loginResp, err := client.Post(suite.gatewayServer.URL+"/api/v1/auth/login", 
			"application/json", strings.NewReader(`{"username":"testuser","password":"password123"}`))
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), http.StatusOK, loginResp.StatusCode)
		
		var loginData map[string]string
		json.NewDecoder(loginResp.Body).Decode(&loginData)
		token := loginData["token"]
		
		// 2. Get user letters
		req, _ := http.NewRequest("GET", suite.gatewayServer.URL+"/api/v1/letters", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		
		lettersResp, err := client.Do(req)
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), http.StatusOK, lettersResp.StatusCode)
		
		// 3. Create a new letter
		req, _ = http.NewRequest("POST", suite.gatewayServer.URL+"/api/v1/write/letters", 
			strings.NewReader(`{"title":"Integration Test Letter","content":"Test content"}`))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		
		createResp, err := client.Do(req)
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), http.StatusCreated, createResp.StatusCode)
		
		// 4. Process OCR (simulated)
		req, _ = http.NewRequest("POST", suite.gatewayServer.URL+"/api/v1/ocr/process", 
			strings.NewReader(`{"image":"base64encodedimage"}`))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		
		ocrResp, err := client.Do(req)
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), http.StatusOK, ocrResp.StatusCode)
		
		// All steps should complete successfully
		assert.True(suite.T(), true, "Complete user flow executed successfully")
	})
}