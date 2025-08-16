package security

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"openpenpal-backend/internal/config"
	"openpenpal-backend/internal/middleware"
	"openpenpal-backend/internal/security/csrf"
	"openpenpal-backend/internal/security/env"
	"openpenpal-backend/internal/security/secrets"
	"openpenpal-backend/internal/security/validation"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SecurityConfig holds all security-related configurations
type SecurityConfig struct {
	JWTManager      *secrets.JWTManager
	CSRFProtection  *csrf.EnhancedCSRF
	RateLimiter     *middleware.EnhancedRateLimiter
	InputValidator  *validation.InputValidator
	SecureConfig    *env.SecureConfig
	SecurityHeaders *middleware.SecurityConfig
	Environment     string
}

// InitializeSecurity sets up all security components
func InitializeSecurity(cfg *config.Config) (*SecurityConfig, error) {
	environment := cfg.Environment
	if environment == "" {
		environment = "development"
	}

	log.Printf("🔒 Initializing security for environment: %s", environment)

	// 1. Initialize secure configuration
	secureConfig, err := env.NewSecureConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize secure config: %w", err)
	}

	// Generate secure defaults if needed
	if environment == "development" {
		defaults := env.GenerateSecureDefaults()
		for key, value := range defaults {
			if secureConfig.Get(key) == "" {
				if err := secureConfig.Set(key, value); err != nil {
					log.Printf("Warning: failed to set default for %s: %v", key, err)
				}
			}
		}
	}

	// Validate required configuration
	requiredKeys := []string{"JWT_SECRET"}
	if err := secureConfig.ValidateRequired(requiredKeys); err != nil {
		return nil, fmt.Errorf("missing required configuration: %w", err)
	}

	// 2. Initialize JWT Manager
	jwtManager := secrets.NewJWTManager(90) // 90 days rotation
	if err := jwtManager.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize JWT manager: %w", err)
	}

	// Override config JWT secret with managed one
	if jwtSecret, err := jwtManager.GetSecretString(); err == nil {
		cfg.JWTSecret = jwtSecret
	}

	// 3. Initialize CSRF Protection
	csrfConfig := &csrf.Config{
		CookieName:  "csrf_token",
		HeaderName:  "X-CSRF-Token",
		FieldName:   "csrf_token",
		Secure:      environment == "production",
		SameSite:    getSameSite(environment),
		MaxAge:      24 * 60 * 60, // 24 hours
		CleanupTime: 60 * 60,      // 1 hour
	}
	csrfProtection := csrf.NewEnhancedCSRF(csrfConfig)

	// 4. Initialize Rate Limiter
	rateLimitConfig := getRateLimitConfig(environment)
	rateLimiter := middleware.NewEnhancedRateLimiter(rateLimitConfig)

	// 5. Initialize Input Validator
	inputValidator := validation.NewInputValidator()

	// 6. Configure Security Headers
	securityHeaders := middleware.NewSecurityConfig()

	// Log security initialization summary
	log.Println("✅ Security initialization completed:")
	log.Printf("  - JWT Manager: Initialized with %d-day rotation", 90)
	log.Printf("  - CSRF Protection: %s mode", environment)
	log.Printf("  - Rate Limiting: Adaptive limiting enabled")
	log.Printf("  - Input Validation: Comprehensive validation active")
	log.Printf("  - Security Headers: %s configuration", environment)

	return &SecurityConfig{
		JWTManager:      jwtManager,
		CSRFProtection:  csrfProtection,
		RateLimiter:     rateLimiter,
		InputValidator:  inputValidator,
		SecureConfig:    secureConfig,
		SecurityHeaders: securityHeaders,
		Environment:     environment,
	}, nil
}

// ApplySecurityMiddleware applies all security middleware to the router
func ApplySecurityMiddleware(router *gin.Engine, db *gorm.DB, cfg *config.Config, security *SecurityConfig) {
	// 1. Security Headers (first middleware)
	router.Use(middleware.SecurityHeadersMiddleware())

	// 2. Request ID (for tracking)
	router.Use(middleware.RequestIDMiddleware())

	// 3. Input Validation
	router.Use(security.InputValidator.Middleware())

	// 4. CORS (if needed)
	if corsMiddleware := getCORSMiddleware(security.Environment); corsMiddleware != nil {
		router.Use(corsMiddleware)
	}

	// 5. Apply route-specific middleware
	applyRouteMiddleware(router, db, cfg, security)
}

// applyRouteMiddleware applies middleware to specific route groups
func applyRouteMiddleware(router *gin.Engine, db *gorm.DB, cfg *config.Config, security *SecurityConfig) {
	// Public routes (no auth required)
	public := router.Group("/api/v1")
	{
		// Rate limiting for public endpoints
		public.Use(security.RateLimiter.GeneralLimiter())

		// Auth routes with stricter rate limiting
		auth := public.Group("/auth")
		auth.Use(security.RateLimiter.AuthLimiter())
		
		// CSRF token endpoint (exempt from CSRF check)
		auth.GET("/csrf", gin.HandlerFunc(func(c *gin.Context) {
			token := security.CSRFProtection.GetToken(c)
			c.JSON(200, gin.H{
				"success": true,
				"token":   token,
			})
		}))

		// Login/Register with CSRF protection
		authWithCSRF := auth.Group("")
		authWithCSRF.Use(security.CSRFProtection.Middleware())
		// Auth handlers will be added here
	}

	// Protected routes (auth required)
	protected := router.Group("/api/v1")
	{
		// Apply auth middleware
		protected.Use(middleware.AuthMiddleware(cfg, db))
		
		// Apply CSRF protection
		protected.Use(security.CSRFProtection.Middleware())
		
		// Apply general rate limiting
		protected.Use(security.RateLimiter.APILimiter())
		
		// Protected handlers will be added here
	}

	// Admin routes (admin auth required)
	admin := router.Group("/api/v1/admin")
	{
		admin.Use(middleware.AuthMiddleware(cfg, db))
		admin.Use(middleware.RoleMiddleware("admin"))
		admin.Use(security.CSRFProtection.Middleware())
		admin.Use(security.RateLimiter.APILimiter())
		
		// Security monitoring endpoints
		admin.GET("/security/metrics", getSecurityMetrics(security))
		admin.GET("/security/blocked", getBlockedIdentifiers(security))
	}

	// Upload routes with special rate limiting
	upload := router.Group("/api/v1/upload")
	{
		upload.Use(middleware.AuthMiddleware(cfg, db))
		upload.Use(security.CSRFProtection.Middleware())
		upload.Use(security.RateLimiter.UploadLimiter())
		// Upload handlers will be added here
	}

	// Health check (no middleware)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	// CSP violation reports
	router.POST("/api/v1/security/csp-report", middleware.CSPViolationHandler())
}

// getRateLimitConfig returns rate limit configuration based on environment
func getRateLimitConfig(environment string) *middleware.RateLimitConfig {
	if environment == "production" {
		return &middleware.RateLimitConfig{
			GeneralRate:  1,    // 1 req/sec
			GeneralBurst: 5,    // burst of 5
			AuthRate:     0.2,  // 1 req per 5 seconds
			AuthBurst:    2,    // burst of 2
			APIRate:      3,    // 3 req/sec
			APIBurst:     10,   // burst of 10
			UploadRate:   0.1,  // 1 req per 10 seconds
			UploadBurst:  1,    // burst of 1
		}
	}

	// Development/test environment - more permissive
	return middleware.DefaultRateLimitConfig()
}

// getSameSite returns appropriate SameSite setting
func getSameSite(environment string) http.SameSite {
	if environment == "production" {
		return http.SameSiteStrictMode
	}
	return http.SameSiteLaxMode
}

// getCORSMiddleware returns CORS middleware if needed
func getCORSMiddleware(environment string) gin.HandlerFunc {
	// Implement CORS middleware based on your needs
	// For now, returning nil to use existing CORS setup
	return nil
}

// getSecurityMetrics returns a handler for security metrics
func getSecurityMetrics(security *SecurityConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		metrics := security.RateLimiter.GetMetrics()
		c.JSON(200, gin.H{
			"success": true,
			"metrics": metrics,
		})
	}
}

// getBlockedIdentifiers returns a handler for blocked identifiers
func getBlockedIdentifiers(security *SecurityConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// This would get blocked IPs/users from the rate limiter
		metrics := security.RateLimiter.GetMetrics()
		blocked := metrics["blocked_identifiers"]
		
		c.JSON(200, gin.H{
			"success": true,
			"blocked": blocked,
		})
	}
}

// ValidateSecurityConfig performs security configuration validation
func ValidateSecurityConfig(cfg *config.Config) error {
	// Check for insecure defaults
	insecureValues := map[string][]string{
		"JWT_SECRET": {
			"change-this-to-a-secure-secret-key",
			"dev-secret-key-do-not-use-in-production",
			"secret",
			"password",
		},
		"DATABASE_PASSWORD": {
			"password",
			"admin",
			"root",
			"123456",
		},
	}

	for key, insecureList := range insecureValues {
		value := os.Getenv(key)
		for _, insecure := range insecureList {
			if value == insecure {
				return fmt.Errorf("insecure value detected for %s", key)
			}
		}
	}

	return nil
}