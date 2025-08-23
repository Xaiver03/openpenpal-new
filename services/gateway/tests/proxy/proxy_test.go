package proxy_test

import (
	"api-gateway/internal/config"
	"api-gateway/internal/proxy"
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

type ProxyTestSuite struct {
	suite.Suite
	cfg           *config.Config
	logger        *zap.Logger
	proxyManager  *proxy.Manager
	mockBackend   *httptest.Server
	mockBackend2  *httptest.Server
}

func TestProxyTestSuite(t *testing.T) {
	suite.Run(t, new(ProxyTestSuite))
}

func (suite *ProxyTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	
	// Create mock backend servers
	suite.mockBackend = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/health":
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
		case "/api/v1/users":
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"users": []map[string]string{
					{"id": "1", "name": "John"},
					{"id": "2", "name": "Jane"},
				},
			})
		case "/api/v1/users/123":
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"id": "123", "name": "Test User"})
		case "/slow":
			time.Sleep(2 * time.Second)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("slow response"))
		case "/error":
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("internal server error"))
		case "/echo":
			// Echo back the request info
			reqInfo := map[string]interface{}{
				"method":  r.Method,
				"path":    r.URL.Path,
				"query":   r.URL.RawQuery,
				"headers": r.Header,
			}
			json.NewEncoder(w).Encode(reqInfo)
		default:
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("not found"))
		}
	}))

	suite.mockBackend2 = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"server": "backend2", "path": r.URL.Path})
	}))

	// Setup configuration
	suite.cfg = &config.Config{
		Services: map[string]*config.ServiceConfig{
			"backend": {
				Name:        "backend",
				Hosts:       []string{strings.TrimPrefix(suite.mockBackend.URL, "http://")},
				HealthCheck: "/health",
				Timeout:     30,
				Retries:     3,
				Weight:      100,
			},
			"backend2": {
				Name:        "backend2",
				Hosts:       []string{strings.TrimPrefix(suite.mockBackend2.URL, "http://")},
				HealthCheck: "/health",
				Timeout:     30,
				Retries:     3,
				Weight:      100,
			},
		},
		ProxyTimeout:     30,
		ConnectTimeout:   10,
		KeepAliveTimeout: 90,
	}

	suite.logger = zap.NewNop()
	suite.proxyManager = proxy.NewManager(suite.cfg, suite.logger)
}

func (suite *ProxyTestSuite) TearDownTest() {
	if suite.mockBackend != nil {
		suite.mockBackend.Close()
	}
	if suite.mockBackend2 != nil {
		suite.mockBackend2.Close()
	}
}

func (suite *ProxyTestSuite) TestBasicProxying() {
	suite.Run("Successful proxy request", func() {
		router := gin.New()
		router.Any("/api/v1/*path", suite.proxyManager.ProxyHandler("backend"))

		req, _ := http.NewRequest("GET", "/api/v1/users", nil)
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(suite.T(), err)
		assert.Contains(suite.T(), response, "users")
	})

	suite.Run("Proxy with path parameters", func() {
		router := gin.New()
		router.Any("/api/v1/*path", suite.proxyManager.ProxyHandler("backend"))

		req, _ := http.NewRequest("GET", "/api/v1/users/123", nil)
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusOK, w.Code)
		
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), "123", response["id"])
		assert.Equal(suite.T(), "Test User", response["name"])
	})

	suite.Run("Proxy POST request with body", func() {
		router := gin.New()
		router.Any("/api/v1/*path", suite.proxyManager.ProxyHandler("backend"))

		requestBody := `{"name": "New User", "email": "test@example.com"}`
		req, _ := http.NewRequest("POST", "/api/v1/echo", strings.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), "POST", response["method"])
		assert.Equal(suite.T(), "/echo", response["path"])
	})
}

func (suite *ProxyTestSuite) TestErrorHandling() {
	suite.Run("Service not found", func() {
		router := gin.New()
		router.Any("/api/v1/*path", suite.proxyManager.ProxyHandler("nonexistent"))

		req, _ := http.NewRequest("GET", "/api/v1/users", nil)
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusServiceUnavailable, w.Code)
	})

	suite.Run("Backend returns error", func() {
		router := gin.New()
		router.Any("/api/v1/*path", suite.proxyManager.ProxyHandler("backend"))

		req, _ := http.NewRequest("GET", "/api/v1/error", nil)
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)
	})

	suite.Run("Backend not found endpoint", func() {
		router := gin.New()
		router.Any("/api/v1/*path", suite.proxyManager.ProxyHandler("backend"))

		req, _ := http.NewRequest("GET", "/api/v1/nonexistent", nil)
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusNotFound, w.Code)
	})
}

func (suite *ProxyTestSuite) TestRequestModification() {
	suite.Run("Headers are properly forwarded", func() {
		router := gin.New()
		router.Any("/api/v1/*path", suite.proxyManager.ProxyHandler("backend"))

		req, _ := http.NewRequest("GET", "/api/v1/echo", nil)
		req.Header.Set("Authorization", "Bearer test-token")
		req.Header.Set("X-User-ID", "123")
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(suite.T(), err)
		
		headers := response["headers"].(map[string]interface{})
		assert.Contains(suite.T(), headers, "Authorization")
		assert.Contains(suite.T(), headers, "X-User-Id")
	})

	suite.Run("Query parameters are preserved", func() {
		router := gin.New()
		router.Any("/api/v1/*path", suite.proxyManager.ProxyHandler("backend"))

		req, _ := http.NewRequest("GET", "/api/v1/echo?page=1&limit=10&sort=name", nil)
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(suite.T(), err)
		
		query := response["query"].(string)
		assert.Contains(suite.T(), query, "page=1")
		assert.Contains(suite.T(), query, "limit=10")
		assert.Contains(suite.T(), query, "sort=name")
	})
}

func (suite *ProxyTestSuite) TestTimeoutHandling() {
	suite.Run("Request timeout", func() {
		// Configure a shorter timeout for this test
		shortTimeoutConfig := &config.Config{
			Services: map[string]*config.ServiceConfig{
				"backend": {
					Name:        "backend",
					Hosts:       []string{strings.TrimPrefix(suite.mockBackend.URL, "http://")},
					HealthCheck: "/health",
					Timeout:     1, // 1 second timeout
					Retries:     1,
					Weight:      100,
				},
			},
			ProxyTimeout: 1,
		}

		shortTimeoutManager := proxy.NewManager(shortTimeoutConfig, suite.logger)
		
		router := gin.New()
		router.Any("/api/v1/*path", shortTimeoutManager.ProxyHandler("backend"))

		req, _ := http.NewRequest("GET", "/api/v1/slow", nil)
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)

		// Should timeout and return 504 Gateway Timeout
		assert.Equal(suite.T(), http.StatusGatewayTimeout, w.Code)
	})
}

func (suite *ProxyTestSuite) TestPathRewriting() {
	suite.Run("Path rewriting for different services", func() {
		router := gin.New()
		
		// Setup path rewriting for different services
		router.Any("/backend/*path", suite.proxyManager.ProxyHandler("backend"))
		router.Any("/backend2/*path", suite.proxyManager.ProxyHandler("backend2"))

		// Test backend routing
		req, _ := http.NewRequest("GET", "/backend/api/v1/echo", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(suite.T(), http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), "/echo", response["path"]) // Path should be rewritten

		// Test backend2 routing
		req, _ = http.NewRequest("GET", "/backend2/test", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(suite.T(), http.StatusOK, w.Code)
		
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), "backend2", response["server"])
	})
}

func (suite *ProxyTestSuite) TestLoadBalancing() {
	suite.Run("Load balancing between multiple instances", func() {
		// Create multiple backend instances
		backend3 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(map[string]string{"server": "backend3"})
		}))
		defer backend3.Close()

		// Configure service with multiple hosts
		multiHostConfig := &config.Config{
			Services: map[string]*config.ServiceConfig{
				"multi-backend": {
					Name: "multi-backend",
					Hosts: []string{
						strings.TrimPrefix(suite.mockBackend.URL, "http://"),
						strings.TrimPrefix(suite.mockBackend2.URL, "http://"),
						strings.TrimPrefix(backend3.URL, "http://"),
					},
					HealthCheck: "/health",
					Timeout:     30,
					Retries:     3,
					Weight:      100,
				},
			},
		}

		multiHostManager := proxy.NewManager(multiHostConfig, suite.logger)
		
		router := gin.New()
		router.Any("/multi/*path", multiHostManager.ProxyHandler("multi-backend"))

		// Make multiple requests and check load balancing
		serverCounts := make(map[string]int)
		
		for i := 0; i < 10; i++ {
			req, _ := http.NewRequest("GET", "/multi/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			if w.Code == http.StatusOK {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				if server, exists := response["server"]; exists {
					serverCounts[server.(string)]++
				}
			}
		}

		// Should distribute requests across backends
		assert.Greater(suite.T(), len(serverCounts), 0)
	})
}

func (suite *ProxyTestSuite) TestHealthChecking() {
	suite.Run("Health check functionality", func() {
		healthChecker := suite.proxyManager.GetHealthChecker()
		
		// Check healthy service
		isHealthy := healthChecker.IsServiceHealthy("backend")
		assert.True(suite.T(), isHealthy)
		
		// Force health check
		healthChecker.CheckServiceHealth("backend")
		
		// Service should still be healthy
		isHealthy = healthChecker.IsServiceHealthy("backend")
		assert.True(suite.T(), isHealthy)
	})

	suite.Run("Unhealthy service handling", func() {
		// Create a service that will fail health checks
		unhealthyBackend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/health" {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer unhealthyBackend.Close()

		unhealthyConfig := &config.Config{
			Services: map[string]*config.ServiceConfig{
				"unhealthy": {
					Name:        "unhealthy",
					Hosts:       []string{strings.TrimPrefix(unhealthyBackend.URL, "http://")},
					HealthCheck: "/health",
					Timeout:     30,
					Retries:     3,
					Weight:      100,
				},
			},
		}

		unhealthyManager := proxy.NewManager(unhealthyConfig, suite.logger)
		healthChecker := unhealthyManager.GetHealthChecker()
		
		// Force health check
		healthChecker.CheckServiceHealth("unhealthy")
		
		// Service should be marked as unhealthy
		isHealthy := healthChecker.IsServiceHealthy("unhealthy")
		assert.False(suite.T(), isHealthy)
	})
}

func (suite *ProxyTestSuite) TestMetricsCollection() {
	suite.Run("Request metrics collection", func() {
		router := gin.New()
		
		// Add metrics middleware
		router.Use(suite.proxyManager.MetricsMiddleware())
		router.Any("/api/v1/*path", suite.proxyManager.ProxyHandler("backend"))

		// Make several requests
		for i := 0; i < 5; i++ {
			req, _ := http.NewRequest("GET", "/api/v1/users", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			assert.Equal(suite.T(), http.StatusOK, w.Code)
		}

		// Check metrics
		metrics := suite.proxyManager.GetMetrics()
		assert.NotNil(suite.T(), metrics)
		assert.Greater(suite.T(), metrics.RequestCount, uint64(0))
	})
}

func (suite *ProxyTestSuite) TestRetryMechanism() {
	suite.Run("Retry on failure", func() {
		// Create a backend that fails first few times then succeeds
		attemptCount := 0
		retryBackend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attemptCount++
			if attemptCount < 3 {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"message": "success after retry",
				"attempt": attemptCount,
			})
		}))
		defer retryBackend.Close()

		retryConfig := &config.Config{
			Services: map[string]*config.ServiceConfig{
				"retry-backend": {
					Name:        "retry-backend",
					Hosts:       []string{strings.TrimPrefix(retryBackend.URL, "http://")},
					HealthCheck: "/health",
					Timeout:     30,
					Retries:     3,
					Weight:      100,
				},
			},
		}

		retryManager := proxy.NewManager(retryConfig, suite.logger)
		
		router := gin.New()
		router.Any("/retry/*path", retryManager.ProxyHandler("retry-backend"))

		req, _ := http.NewRequest("GET", "/retry/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), "success after retry", response["message"])
		assert.Equal(suite.T(), float64(3), response["attempt"])
	})
}

func (suite *ProxyTestSuite) TestCircuitBreaker() {
	suite.Run("Circuit breaker functionality", func() {
		// Create a backend that always fails
		failingBackend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer failingBackend.Close()

		circuitBreakerConfig := &config.Config{
			Services: map[string]*config.ServiceConfig{
				"failing-backend": {
					Name:        "failing-backend",
					Hosts:       []string{strings.TrimPrefix(failingBackend.URL, "http://")},
					HealthCheck: "/health",
					Timeout:     30,
					Retries:     1,
					Weight:      100,
				},
			},
		}

		cbManager := proxy.NewManager(circuitBreakerConfig, suite.logger)
		
		router := gin.New()
		router.Any("/failing/*path", cbManager.ProxyHandler("failing-backend"))

		// Make multiple requests to trigger circuit breaker
		for i := 0; i < 10; i++ {
			req, _ := http.NewRequest("GET", "/failing/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			// Should get either 500 (from backend) or 503 (circuit open)
			assert.True(suite.T(), w.Code == http.StatusInternalServerError || w.Code == http.StatusServiceUnavailable)
		}

		// Circuit breaker should eventually open and return 503
		circuitBreakerStatus := cbManager.GetCircuitBreakerStatus("failing-backend")
		assert.NotNil(suite.T(), circuitBreakerStatus)
	})
}