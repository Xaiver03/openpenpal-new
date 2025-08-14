package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
	
	"openpenpal-backend/internal/config"
	"openpenpal-backend/internal/middleware"
	"openpenpal-backend/internal/models"
	"openpenpal-backend/pkg/auth"

	"github.com/gin-gonic/gin"
)

// MockUser for testing without database
type MockUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	IsActive bool   `json:"is_active"`
}

var mockUsers = map[string]*MockUser{
	"admin": {
		ID:       "admin-user",
		Username: "admin",
		Email:    "admin@example.com",
		Role:     "super_admin",
		IsActive: true,
	},
	"user": {
		ID:       "test-user",
		Username: "user",
		Email:    "user@example.com", 
		Role:     "user",
		IsActive: true,
	},
}

var mockTokens = map[string]string{
	"admin": "",
	"user":  "",
}

func main() {
	// Load config
	cfg := &config.Config{
		Environment: "development",
		Host:        "localhost",
		Port:        "8080",
		JWTSecret:   "test-secret-for-middleware-testing",
		FrontendURL: "http://localhost:3000",
	}

	// Create router
	router := gin.New()

	// Apply all middleware in SOTA order
	router.Use(middleware.RequestIDMiddleware())
	router.Use(middleware.LoggerMiddleware())
	router.Use(middleware.RecoveryMiddleware())
	router.Use(middleware.MetricsMiddleware())
	router.Use(middleware.SecurityHeadersMiddleware())
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.RequestSizeLimitMiddleware(middleware.DefaultMaxRequestSize))
	router.Use(middleware.RateLimitMiddleware())
	router.Use(middleware.RequestTransformMiddleware())
	router.Use(middleware.ResponseTransformMiddleware())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":     "healthy",
			"service":    "middleware-test-server",
			"timestamp":  time.Now().Format(time.RFC3339),
			"middleware": "all active",
		})
	})

	// API routes
	v1 := router.Group("/api/v1")

	// Auth routes
	authGroup := v1.Group("/auth")
	authGroup.Use(middleware.AuthRateLimitMiddleware())
	{
		// CSRF token endpoint
		authGroup.GET("/csrf", func(c *gin.Context) {
			handler := middleware.NewCSRFHandler()
			handler.GetCSRFToken(c)
		})

		// Login endpoint
		authGroup.POST("/login", func(c *gin.Context) {
			var req struct {
				Username string `json:"username"`
				Password string `json:"password"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(400, gin.H{"error": "Invalid request"})
				return
			}

			// Mock authentication
			user, exists := mockUsers[req.Username]
			if !exists || req.Password != "password123" {
				c.JSON(401, gin.H{"error": "Invalid credentials"})
				return
			}

			// Generate token
			expiresAt := time.Now().Add(24 * time.Hour)
			token, err := auth.GenerateJWT(user.ID, models.UserRole(user.Role), cfg.JWTSecret, expiresAt)
			if err != nil {
				c.JSON(500, gin.H{"error": "Failed to generate token"})
				return
			}

			mockTokens[user.Username] = token

			c.JSON(200, gin.H{
				"success": true,
				"data": gin.H{
					"token": token,
					"user": gin.H{
						"id":         user.ID,
						"username":   user.Username,
						"email":      user.Email,
						"role":       user.Role,
						"is_active":  user.IsActive,
						"created_at": time.Now().Format(time.RFC3339),
					},
				},
			})
		})

		// Protected route - Get current user
		authGroup.GET("/me", mockAuthMiddleware(cfg), func(c *gin.Context) {
			userID, _ := c.Get("user_id")
			role, _ := c.Get("user_role")

			// Find user
			var user *MockUser
			for _, u := range mockUsers {
				if u.ID == userID {
					user = u
					break
				}
			}

			if user == nil {
				c.JSON(404, gin.H{"error": "User not found"})
				return
			}

			c.JSON(200, gin.H{
				"success": true,
				"data": gin.H{
					"id":         user.ID,
					"username":   user.Username,
					"email":      user.Email,
					"role":       role,
					"is_active":  user.IsActive,
					"created_at": time.Now().Format(time.RFC3339),
					"updated_at": time.Now().Format(time.RFC3339),
				},
			})
		})
	}

	// Admin routes
	adminGroup := v1.Group("/admin")
	adminGroup.Use(mockAuthMiddleware(cfg))
	adminGroup.Use(mockAdminMiddleware())
	{
		adminGroup.GET("/users", func(c *gin.Context) {
			users := []gin.H{}
			for _, user := range mockUsers {
				users = append(users, gin.H{
					"id":         user.ID,
					"username":   user.Username,
					"email":      user.Email,
					"role":       user.Role,
					"is_active":  user.IsActive,
					"created_at": time.Now().Format(time.RFC3339),
				})
			}

			c.JSON(200, gin.H{
				"success": true,
				"data": gin.H{
					"users": users,
					"total": len(users),
					"page":  1,
					"limit": 20,
				},
			})
		})

		adminGroup.GET("/dashboard/stats", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"success": true,
				"data": gin.H{
					"total_users":   len(mockUsers),
					"active_users":  2,
					"total_letters": 100,
					"letters_today": 5,
				},
			})
		})
	}

	// Test endpoints
	testGroup := v1.Group("/test")
	{
		// Test transformation
		testGroup.POST("/transform", mockAuthMiddleware(cfg), func(c *gin.Context) {
			var body map[string]interface{}
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(400, gin.H{"error": "Invalid JSON"})
				return
			}

			// Echo back the transformed data
			c.JSON(200, gin.H{
				"success": true,
				"received_data": body,
				"server_time": time.Now().Format(time.RFC3339),
			})
		})

		// Test rate limiting
		testGroup.GET("/rate-limit", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"success": true,
				"message": "Rate limit test successful",
				"timestamp": time.Now().Unix(),
			})
		})
	}

	// Start server
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	log.Printf("Starting middleware test server on %s", addr)
	log.Printf("Environment: %s", cfg.Environment)
	log.Printf("All middleware layers active")
	
	if err := router.Run(addr); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

// mockAuthMiddleware simulates auth middleware without database
func mockAuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Missing authorization token",
			})
			c.Abort()
			return
		}

		token, err := auth.ExtractTokenFromHeader(authHeader)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Invalid token format",
			})
			c.Abort()
			return
		}

		claims, err := auth.ValidateJWT(token, cfg.JWTSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// Find mock user
		var user *models.User
		for _, u := range mockUsers {
			if u.ID == claims.UserID {
				user = &models.User{
					ID:       u.ID,
					Username: u.Username,
					Email:    u.Email,
					Role:     models.UserRole(u.Role),
					IsActive: u.IsActive,
				}
				break
			}
		}

		if user == nil || !user.IsActive {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "User not found or inactive",
			})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("user_role", claims.Role)
		c.Set("user", user)
		c.Set("role", claims.Role) // Compatibility
		
		c.Next()
	}
}

// mockAdminMiddleware checks admin permissions
func mockAdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("user_role")
		if !exists {
			c.JSON(401, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		allowedRoles := []string{"admin", "super_admin", "platform_admin"}
		roleStr := role.(string)
		
		allowed := false
		for _, r := range allowedRoles {
			if roleStr == r {
				allowed = true
				break
			}
		}

		if !allowed {
			c.JSON(403, gin.H{
				"error": "Insufficient permissions",
				"required": "admin",
				"current": roleStr,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}