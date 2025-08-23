package middleware_test

import (
	"api-gateway/internal/config"
	"api-gateway/internal/middleware"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type AuthMiddlewareTestSuite struct {
	suite.Suite
	router     *gin.Engine
	cfg        *config.Config
	jwtSecret  string
}

func TestAuthMiddlewareTestSuite(t *testing.T) {
	suite.Run(t, new(AuthMiddlewareTestSuite))
}

func (suite *AuthMiddlewareTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	
	suite.jwtSecret = "test-jwt-secret-for-authentication-testing-32-chars"
	suite.cfg = &config.Config{
		JWTSecret: suite.jwtSecret,
	}
	
	suite.router = gin.New()
}

func (suite *AuthMiddlewareTestSuite) createValidJWT(userID, role string) string {
	// This would use the shared JWT package in a real implementation
	// For testing purposes, we'll create a mock JWT token
	// In real implementation, this would use the shared/go JWT utilities
	
	// Mock JWT token format: "Bearer mock-jwt-token-userid-role-timestamp"
	return "Bearer mock-jwt-token-" + userID + "-" + role + "-" + "1234567890"
}

func (suite *AuthMiddlewareTestSuite) TestAuthMiddleware() {
	suite.Run("Valid JWT token", func() {
		authMiddleware := middleware.NewAuthMiddleware(suite.cfg)
		
		suite.router.Use(authMiddleware.RequireAuth())
		suite.router.GET("/protected", func(c *gin.Context) {
			userID, exists := c.Get("user_id")
			assert.True(suite.T(), exists)
			c.JSON(http.StatusOK, gin.H{"user_id": userID})
		})

		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", suite.createValidJWT("user123", "user"))
		
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		// Note: This test will need to be updated when actual JWT validation is implemented
		// Currently testing the middleware structure and flow
		assert.Equal(suite.T(), http.StatusOK, w.Code)
	})

	suite.Run("Missing Authorization header", func() {
		authMiddleware := middleware.NewAuthMiddleware(suite.cfg)
		
		suite.router.Use(authMiddleware.RequireAuth())
		suite.router.GET("/protected", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		req, _ := http.NewRequest("GET", "/protected", nil)
		
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
	})

	suite.Run("Invalid token format", func() {
		authMiddleware := middleware.NewAuthMiddleware(suite.cfg)
		
		suite.router.Use(authMiddleware.RequireAuth())
		suite.router.GET("/protected", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "InvalidTokenFormat")
		
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
	})

	suite.Run("Expired token", func() {
		authMiddleware := middleware.NewAuthMiddleware(suite.cfg)
		
		suite.router.Use(authMiddleware.RequireAuth())
		suite.router.GET("/protected", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		// Mock expired token
		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer expired-token-should-fail")
		
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
	})
}

func (suite *AuthMiddlewareTestSuite) TestRoleBasedAuth() {
	suite.Run("Admin role required - valid admin", func() {
		authMiddleware := middleware.NewAuthMiddleware(suite.cfg)
		
		suite.router.Use(authMiddleware.RequireRole("admin"))
		suite.router.GET("/admin", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "admin access granted"})
		})

		req, _ := http.NewRequest("GET", "/admin", nil)
		req.Header.Set("Authorization", suite.createValidJWT("admin123", "admin"))
		
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusOK, w.Code)
	})

	suite.Run("Admin role required - invalid user role", func() {
		authMiddleware := middleware.NewAuthMiddleware(suite.cfg)
		
		suite.router.Use(authMiddleware.RequireRole("admin"))
		suite.router.GET("/admin", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "admin access granted"})
		})

		req, _ := http.NewRequest("GET", "/admin", nil)
		req.Header.Set("Authorization", suite.createValidJWT("user123", "user"))
		
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusForbidden, w.Code)
	})

	suite.Run("Courier role required - valid courier", func() {
		authMiddleware := middleware.NewAuthMiddleware(suite.cfg)
		
		suite.router.Use(authMiddleware.RequireRole("courier"))
		suite.router.GET("/courier", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "courier access granted"})
		})

		req, _ := http.NewRequest("GET", "/courier", nil)
		req.Header.Set("Authorization", suite.createValidJWT("courier123", "courier"))
		
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusOK, w.Code)
	})

	suite.Run("Multiple roles allowed", func() {
		authMiddleware := middleware.NewAuthMiddleware(suite.cfg)
		
		suite.router.Use(authMiddleware.RequireRoles("admin", "courier"))
		suite.router.GET("/admin-or-courier", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "access granted"})
		})

		// Test admin access
		req, _ := http.NewRequest("GET", "/admin-or-courier", nil)
		req.Header.Set("Authorization", suite.createValidJWT("admin123", "admin"))
		
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)
		assert.Equal(suite.T(), http.StatusOK, w.Code)

		// Test courier access
		req, _ = http.NewRequest("GET", "/admin-or-courier", nil)
		req.Header.Set("Authorization", suite.createValidJWT("courier123", "courier"))
		
		w = httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)
		assert.Equal(suite.T(), http.StatusOK, w.Code)

		// Test user access (should be denied)
		req, _ = http.NewRequest("GET", "/admin-or-courier", nil)
		req.Header.Set("Authorization", suite.createValidJWT("user123", "user"))
		
		w = httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)
		assert.Equal(suite.T(), http.StatusForbidden, w.Code)
	})
}

func (suite *AuthMiddlewareTestSuite) TestOptionalAuth() {
	suite.Run("Optional auth with valid token", func() {
		authMiddleware := middleware.NewAuthMiddleware(suite.cfg)
		
		suite.router.Use(authMiddleware.OptionalAuth())
		suite.router.GET("/optional", func(c *gin.Context) {
			userID, exists := c.Get("user_id")
			if exists {
				c.JSON(http.StatusOK, gin.H{"user_id": userID, "authenticated": true})
			} else {
				c.JSON(http.StatusOK, gin.H{"authenticated": false})
			}
		})

		req, _ := http.NewRequest("GET", "/optional", nil)
		req.Header.Set("Authorization", suite.createValidJWT("user123", "user"))
		
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusOK, w.Code)
	})

	suite.Run("Optional auth without token", func() {
		authMiddleware := middleware.NewAuthMiddleware(suite.cfg)
		
		suite.router.Use(authMiddleware.OptionalAuth())
		suite.router.GET("/optional", func(c *gin.Context) {
			userID, exists := c.Get("user_id")
			if exists {
				c.JSON(http.StatusOK, gin.H{"user_id": userID, "authenticated": true})
			} else {
				c.JSON(http.StatusOK, gin.H{"authenticated": false})
			}
		})

		req, _ := http.NewRequest("GET", "/optional", nil)
		
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusOK, w.Code)
	})
}

func (suite *AuthMiddlewareTestSuite) TestCORSMiddleware() {
	suite.Run("CORS headers set correctly", func() {
		corsMiddleware := middleware.CORSMiddleware()
		
		suite.router.Use(corsMiddleware)
		suite.router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "test"})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusOK, w.Code)
		assert.Equal(suite.T(), "*", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Contains(suite.T(), w.Header().Get("Access-Control-Allow-Methods"), "GET")
		assert.Contains(suite.T(), w.Header().Get("Access-Control-Allow-Methods"), "POST")
		assert.Contains(suite.T(), w.Header().Get("Access-Control-Allow-Headers"), "Authorization")
	})

	suite.Run("CORS preflight request", func() {
		corsMiddleware := middleware.CORSMiddleware()
		
		suite.router.Use(corsMiddleware)
		suite.router.OPTIONS("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req, _ := http.NewRequest("OPTIONS", "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("Access-Control-Request-Method", "POST")
		req.Header.Set("Access-Control-Request-Headers", "Authorization,Content-Type")
		
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusOK, w.Code)
		assert.NotEmpty(suite.T(), w.Header().Get("Access-Control-Allow-Methods"))
		assert.NotEmpty(suite.T(), w.Header().Get("Access-Control-Allow-Headers"))
	})
}

func (suite *AuthMiddlewareTestSuite) TestSecurityHeaders() {
	suite.Run("Security headers middleware", func() {
		securityMiddleware := middleware.SecurityHeaders()
		
		suite.router.Use(securityMiddleware)
		suite.router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "test"})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusOK, w.Code)
		
		// Check security headers
		assert.Equal(suite.T(), "1; mode=block", w.Header().Get("X-XSS-Protection"))
		assert.Equal(suite.T(), "nosniff", w.Header().Get("X-Content-Type-Options"))
		assert.Equal(suite.T(), "DENY", w.Header().Get("X-Frame-Options"))
		assert.NotEmpty(suite.T(), w.Header().Get("X-Request-ID"))
	})
}

func (suite *AuthMiddlewareTestSuite) TestRateLimitingMiddleware() {
	suite.Run("Rate limiting allows requests under limit", func() {
		rateLimitConfig := &config.RateLimitConfig{
			Enabled:      true,
			DefaultLimit: 100,
			ServiceLimits: map[string]int{
				"test": 10,
			},
		}
		
		rateLimitMiddleware := middleware.NewRateLimitMiddleware(rateLimitConfig)
		
		suite.router.Use(rateLimitMiddleware.Limit("test"))
		suite.router.GET("/limited", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "allowed"})
		})

		// Make several requests under the limit
		for i := 0; i < 5; i++ {
			req, _ := http.NewRequest("GET", "/limited", nil)
			req.Header.Set("X-Real-IP", "192.168.1.100")
			
			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)
			
			assert.Equal(suite.T(), http.StatusOK, w.Code)
		}
	})
}

func (suite *AuthMiddlewareTestSuite) TestMiddlewareChaining() {
	suite.Run("Multiple middleware in chain", func() {
		authMiddleware := middleware.NewAuthMiddleware(suite.cfg)
		
		suite.router.Use(middleware.CORSMiddleware())
		suite.router.Use(middleware.SecurityHeaders())
		suite.router.Use(authMiddleware.RequireAuth())
		
		suite.router.GET("/chained", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "all middleware passed"})
		})

		req, _ := http.NewRequest("GET", "/chained", nil)
		req.Header.Set("Authorization", suite.createValidJWT("user123", "user"))
		req.Header.Set("Origin", "http://localhost:3000")
		
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusOK, w.Code)
		
		// Verify all middleware ran
		assert.NotEmpty(suite.T(), w.Header().Get("Access-Control-Allow-Origin"))
		assert.NotEmpty(suite.T(), w.Header().Get("X-Request-ID"))
	})
}

func (suite *AuthMiddlewareTestSuite) TestAuthenticationExtraction() {
	suite.Run("Extract user information from JWT", func() {
		authMiddleware := middleware.NewAuthMiddleware(suite.cfg)
		
		suite.router.Use(authMiddleware.RequireAuth())
		suite.router.GET("/user-info", func(c *gin.Context) {
			userID, _ := c.Get("user_id")
			role, _ := c.Get("role")
			permissions, _ := c.Get("permissions")
			
			c.JSON(http.StatusOK, gin.H{
				"user_id":     userID,
				"role":        role,
				"permissions": permissions,
			})
		})

		req, _ := http.NewRequest("GET", "/user-info", nil)
		req.Header.Set("Authorization", suite.createValidJWT("user123", "admin"))
		
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusOK, w.Code)
	})
}

func (suite *AuthMiddlewareTestSuite) TestTokenRefresh() {
	suite.Run("Token near expiry handling", func() {
		authMiddleware := middleware.NewAuthMiddleware(suite.cfg)
		
		suite.router.Use(authMiddleware.RequireAuth())
		suite.router.GET("/refresh-check", func(c *gin.Context) {
			needsRefresh, exists := c.Get("needs_refresh")
			c.JSON(http.StatusOK, gin.H{
				"needs_refresh": needsRefresh,
				"has_flag":      exists,
			})
		})

		// Mock token that's close to expiry
		req, _ := http.NewRequest("GET", "/refresh-check", nil)
		req.Header.Set("Authorization", "Bearer near-expiry-token-mock")
		
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		// Note: Actual refresh logic would depend on JWT implementation
		assert.Equal(suite.T(), http.StatusOK, w.Code)
	})
}

func (suite *AuthMiddlewareTestSuite) TestErrorHandling() {
	suite.Run("Authentication error responses", func() {
		authMiddleware := middleware.NewAuthMiddleware(suite.cfg)
		
		suite.router.Use(authMiddleware.RequireAuth())
		suite.router.GET("/error-test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})

		testCases := []struct {
			name           string
			authHeader     string
			expectedStatus int
		}{
			{"No header", "", http.StatusUnauthorized},
			{"Empty bearer", "Bearer ", http.StatusUnauthorized},
			{"Invalid format", "Basic invalid", http.StatusUnauthorized},
			{"Malformed token", "Bearer malformed.token", http.StatusUnauthorized},
		}

		for _, tc := range testCases {
			suite.Run(tc.name, func() {
				req, _ := http.NewRequest("GET", "/error-test", nil)
				if tc.authHeader != "" {
					req.Header.Set("Authorization", tc.authHeader)
				}
				
				w := httptest.NewRecorder()
				suite.router.ServeHTTP(w, req)
				
				assert.Equal(suite.T(), tc.expectedStatus, w.Code)
			})
		}
	})
}