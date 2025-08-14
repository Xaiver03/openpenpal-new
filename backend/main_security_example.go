package main

import (
	"fmt"
	"log"
	"time"

	"openpenpal-backend/internal/adapters"
	"openpenpal-backend/internal/config"
	"openpenpal-backend/internal/handlers"
	"openpenpal-backend/internal/middleware"
	"openpenpal-backend/internal/models"
	"openpenpal-backend/internal/routes"
	"openpenpal-backend/internal/security"
	"openpenpal-backend/internal/services"
	"openpenpal-backend/internal/websocket"

	"github.com/gin-gonic/gin"
)

// This is an example of how to integrate the security enhancements into main.go
// Replace the existing main.go content with this enhanced version

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Initialize security components FIRST
	securityConfig, err := security.InitializeSecurity(cfg)
	if err != nil {
		log.Fatal("Failed to initialize security:", err)
	}

	// Validate security configuration
	if err := security.ValidateSecurityConfig(cfg); err != nil {
		log.Fatal("Security configuration validation failed:", err)
	}

	// Update config with secure values
	if secureJWT := securityConfig.SecureConfig.Get("JWT_SECRET"); secureJWT != "" {
		cfg.JWTSecret = secureJWT
	}

	// Initialize database
	db, err := config.SetupDatabaseDirect(cfg)
	if err != nil {
		log.Fatal("Failed to setup database:", err)
	}

	// In development environment, initialize test data
	if cfg.Environment == "development" {
		if err := config.SeedData(db); err != nil {
			log.Printf("Warning: Failed to seed data: %v", err)
		} else {
			log.Printf("Test data seeded successfully")
		}
	}

	// Initialize services
	userService := services.NewUserService(db, cfg)
	letterService := services.NewLetterService(db, cfg)
	envelopeService := services.NewEnvelopeService(db)
	courierService := services.NewCourierService(db)
	museumService := services.NewMuseumService(db)
	
	// Initialize AI service with unified service
	unifiedAIService := services.NewUnifiedAIService(db, cfg)
	aiService := unifiedAIService // Use unified service as AI service

	letterCodeService := services.NewLetterCodeService(db)
	opCodeService := services.NewOPCodeService(db)
	privacyService := services.NewPrivacyService(db)
	adaptiveMatchingService := services.NewAdaptiveMatchingService(db)
	
	// Configuration service
	configService := services.NewConfigService(db)

	// Setup Gin router with appropriate mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	router := gin.New()

	// Apply security middleware FIRST
	security.ApplySecurityMiddleware(router, db, cfg, securityConfig)

	// Apply other global middleware
	router.Use(gin.Recovery())
	router.Use(middleware.Logger())

	// Health check endpoint (public, no auth)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":      "healthy",
			"environment": cfg.Environment,
			"timestamp":   time.Now().Unix(),
		})
	})

	// Public API routes
	public := router.Group("/api/v1")
	{
		// Apply general rate limiting to public routes
		public.Use(securityConfig.RateLimiter.GeneralLimiter())

		// Auth routes with special handling
		auth := public.Group("/auth")
		{
			// Apply stricter rate limiting for auth endpoints
			auth.Use(securityConfig.RateLimiter.AuthLimiter())

			// CSRF token endpoint (no CSRF check needed)
			auth.GET("/csrf", func(c *gin.Context) {
				token := securityConfig.CSRFProtection.GetToken(c)
				c.JSON(200, gin.H{
					"success": true,
					"token":   token,
				})
			})

			// Login and register endpoints WITH CSRF protection
			authWithCSRF := auth.Group("")
			authWithCSRF.Use(securityConfig.CSRFProtection.Middleware())
			
			authHandler := handlers.NewAuthHandler(userService, cfg, securityConfig.JWTManager)
			authWithCSRF.POST("/login", authHandler.LoginWithJWTManager)
			authWithCSRF.POST("/register", authHandler.RegisterWithValidation(securityConfig.InputValidator))
			authWithCSRF.POST("/refresh", middleware.RefreshTokenMiddleware(cfg, db, securityConfig.JWTManager), authHandler.RefreshToken)
		}

		// Other public endpoints
		public.GET("/schools", handlers.NewSchoolHandler().GetSchools)
	}

	// Protected API routes
	api := router.Group("/api/v1")
	{
		// Apply enhanced auth middleware
		api.Use(middleware.EnhancedAuthMiddleware(cfg, db, securityConfig.JWTManager))
		
		// Apply CSRF protection
		api.Use(securityConfig.CSRFProtection.Middleware())
		
		// Apply API rate limiting
		api.Use(securityConfig.RateLimiter.APILimiter())

		// User routes
		userHandler := handlers.NewUserHandler(userService)
		api.GET("/users/me", userHandler.GetCurrentUser)
		api.PUT("/users/me", userHandler.UpdateProfile)
		api.DELETE("/users/me", userHandler.DeleteAccount)

		// Letter routes
		letterHandler := handlers.NewLetterHandler(letterService, letterCodeService)
		api.POST("/letters", letterHandler.CreateLetter)
		api.GET("/letters", letterHandler.GetLetters)
		api.GET("/letters/:id", letterHandler.GetLetter)
		api.PUT("/letters/:id", letterHandler.UpdateLetter)
		api.DELETE("/letters/:id", letterHandler.DeleteLetter)

		// AI routes
		aiHandler := handlers.NewAIHandler(unifiedAIService)
		api.POST("/ai/match", aiHandler.MatchPenPal)
		api.POST("/ai/reply", aiHandler.GenerateReply)
		api.POST("/ai/inspiration", aiHandler.GetInspiration)
		api.GET("/ai/personas", aiHandler.GetPersonaList)

		// Privacy routes
		privacyHandler := handlers.NewPrivacyHandler(privacyService)
		api.GET("/privacy/settings", privacyHandler.GetSettings)
		api.PUT("/privacy/settings", privacyHandler.UpdateSettings)
		api.POST("/privacy/check", privacyHandler.CheckVisibility)

		// Upload routes with special rate limiting
		upload := api.Group("/upload")
		upload.Use(securityConfig.RateLimiter.UploadLimiter())
		uploadHandler := handlers.NewUploadHandler(cfg)
		upload.POST("/avatar", uploadHandler.UploadAvatar)
		upload.POST("/letter-image", uploadHandler.UploadLetterImage)
	}

	// Admin routes
	admin := router.Group("/api/v1/admin")
	{
		// Apply auth and role middleware
		admin.Use(middleware.EnhancedAuthMiddleware(cfg, db, securityConfig.JWTManager))
		admin.Use(middleware.RequireRole("admin", "super_admin"))
		admin.Use(securityConfig.CSRFProtection.Middleware())
		admin.Use(securityConfig.RateLimiter.APILimiter())

		// Security monitoring
		admin.GET("/security/metrics", func(c *gin.Context) {
			metrics := securityConfig.RateLimiter.GetMetrics()
			c.JSON(200, gin.H{
				"success": true,
				"metrics": metrics,
			})
		})

		admin.GET("/security/blocked", func(c *gin.Context) {
			metrics := securityConfig.RateLimiter.GetMetrics()
			c.JSON(200, gin.H{
				"success": true,
				"blocked": metrics["blocked_identifiers"],
			})
		})

		// Admin handlers
		adminHandler := handlers.NewAdminHandler(userService, letterService, configService)
		admin.GET("/stats", adminHandler.GetSystemStats)
		admin.GET("/users", adminHandler.GetUsers)
		admin.PUT("/users/:id/status", adminHandler.UpdateUserStatus)
		admin.GET("/config", adminHandler.GetConfig)
		admin.PUT("/config", adminHandler.UpdateConfig)
	}

	// WebSocket endpoint (with auth)
	wsHub := websocket.NewHub()
	go wsHub.Run()

	router.GET("/ws", middleware.EnhancedAuthMiddleware(cfg, db, securityConfig.JWTManager), func(c *gin.Context) {
		websocket.HandleWebSocket(wsHub, c)
	})

	// Security event reporting
	router.POST("/api/v1/security/csp-report", middleware.SecurityReportHandler())

	// Start server
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	log.Printf("🚀 Server starting on %s in %s mode", addr, cfg.Environment)
	log.Printf("🔒 Security features enabled:")
	log.Printf("  - JWT with rotation support")
	log.Printf("  - Enhanced CSRF protection")
	log.Printf("  - Adaptive rate limiting")
	log.Printf("  - Input validation")
	log.Printf("  - Security headers")

	if err := router.Run(addr); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

// Example of enhanced auth handler methods
type EnhancedAuthHandler struct {
	*handlers.AuthHandler
	jwtManager     *secrets.JWTManager
	inputValidator *validation.InputValidator
}

func (h *EnhancedAuthHandler) LoginWithJWTManager(c *gin.Context) {
	// Implementation would use the JWT manager for token generation
	// This is just an example structure
}

func (h *EnhancedAuthHandler) RegisterWithValidation(validator *validation.InputValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Username string `json:"username" validate:"required,safe_username"`
			Email    string `json:"email" validate:"required,safe_email"`
			Password string `json:"password" validate:"required,min=8,max=128"`
		}

		if err := validator.ValidateJSON(c, &req); err != nil {
			return // Validation error already sent
		}

		// Continue with registration logic
	}
}