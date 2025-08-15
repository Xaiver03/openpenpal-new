package handlers

import (
	"net/http"
	"openpenpal-backend/internal/docs"

	"github.com/gin-gonic/gin"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
)

// DocsHandler handles API documentation endpoints
type DocsHandler struct{}

// NewDocsHandler creates a new documentation handler
func NewDocsHandler() *DocsHandler {
	return &DocsHandler{}
}

// RegisterSwaggerRoutes registers Swagger documentation routes
func (h *DocsHandler) RegisterSwaggerRoutes(router *gin.Engine) {
	// Swagger documentation endpoint
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	
	// API documentation redirect
	router.GET("/docs", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})
	
	// API schema endpoint
	router.GET("/api/schema", h.GetAPISchema)
	
	// OpenAPI JSON endpoint
	router.GET("/api/v1/openapi.json", h.GetOpenAPISpec)
	
	// Health check for documentation
	router.GET("/docs/health", h.DocsHealthCheck)
}

// GetAPISchema returns the OpenAPI schema
func (h *DocsHandler) GetAPISchema(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	c.Header("Access-Control-Allow-Origin", "*")
	
	// Return the Swagger spec
	c.JSON(http.StatusOK, gin.H{
		"openapi": "3.0.0",
		"info": gin.H{
			"title":       "OpenPenPal API",
			"version":     "1.0.0",
			"description": "Comprehensive API for OpenPenPal campus letter platform",
		},
		"servers": []gin.H{
			{
				"url":         "http://localhost:8080",
				"description": "Development server",
			},
			{
				"url":         "https://api.openpenpal.org",
				"description": "Production server",
			},
		},
	})
}

// GetOpenAPISpec returns the complete OpenAPI specification
func (h *DocsHandler) GetOpenAPISpec(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	c.Header("Access-Control-Allow-Origin", "*")
	
	// Get the Swagger spec from docs package
	spec := docs.SwaggerInfo
	
	c.JSON(http.StatusOK, gin.H{
		"openapi": "3.0.0",
		"info": gin.H{
			"title":       spec.Title,
			"version":     spec.Version,
			"description": spec.Description,
			"contact": gin.H{
				"name":  "OpenPenPal Support",
				"email": "support@openpenpal.org",
				"url":   "https://openpenpal.org/support",
			},
			"license": gin.H{
				"name": "MIT",
				"url":  "https://opensource.org/licenses/MIT",
			},
		},
		"servers": []gin.H{
			{
				"url":         "http://localhost:8080",
				"description": "Development server",
			},
			{
				"url":         "https://api.openpenpal.org",
				"description": "Production server",
			},
		},
		"tags": []gin.H{
			{
				"name":        "Authentication",
				"description": "User authentication and authorization endpoints",
			},
			{
				"name":        "Letters",
				"description": "Letter creation, management, and publishing endpoints",
			},
			{
				"name":        "Courier",
				"description": "Courier system and task management endpoints",
			},
			{
				"name":        "Museum",
				"description": "Museum entries and exhibition management endpoints",
			},
			{
				"name":        "AI",
				"description": "AI-powered features including writing assistance",
			},
			{
				"name":        "Scheduler",
				"description": "Task scheduling and automation endpoints",
			},
			{
				"name":        "Admin",
				"description": "Administrative functions and system management",
			},
		},
	})
}

// DocsHealthCheck provides health status for documentation service
func (h *DocsHandler) DocsHealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":      "healthy",
		"service":     "documentation",
		"swagger_ui":  "available",
		"openapi_spec": "available",
		"endpoints": gin.H{
			"swagger_ui":    "/swagger/index.html",
			"docs_redirect": "/docs",
			"schema":        "/api/schema",
			"openapi_json":  "/api/v1/openapi.json",
		},
	})
}

// GetAPIEndpoints returns a structured list of all available API endpoints
func (h *DocsHandler) GetAPIEndpoints(c *gin.Context) {
	endpoints := map[string]interface{}{
		"authentication": map[string]string{
			"POST /api/v1/auth/register":      "Register new user",
			"POST /api/v1/auth/login":         "User login",
			"POST /api/v1/auth/logout":        "User logout",
			"POST /api/v1/auth/refresh":       "Refresh JWT token",
			"GET  /api/v1/auth/me":            "Get current user info",
		},
		"letters": map[string]string{
			"GET    /api/v1/letters":           "Get user letters",
			"POST   /api/v1/letters":           "Create draft letter",
			"GET    /api/v1/letters/:id":       "Get letter by ID",
			"PUT    /api/v1/letters/:id":       "Update letter",
			"DELETE /api/v1/letters/:id":       "Delete letter",
			"POST   /api/v1/letters/:id/publish": "Publish letter",
			"POST   /api/v1/letters/:id/generate-code": "Generate letter code",
		},
		"courier": map[string]string{
			"POST /api/v1/courier/apply":       "Apply for courier position",
			"GET  /api/v1/courier/status":      "Get courier application status",
			"GET  /api/v1/courier/tasks":       "Get courier tasks",
			"POST /api/v1/courier/create":      "Create subordinate courier",
		},
		"museum": map[string]string{
			"GET  /api/v1/museum/entries":      "Get museum entries",
			"POST /api/v1/museum/items":        "Create museum item",
			"GET  /api/v1/museum/exhibitions":  "Get exhibitions",
			"POST /api/v1/museum/submit":       "Submit letter to museum",
		},
		"ai": map[string]string{
			"POST /api/v1/ai/inspiration":      "Get writing inspiration",
			"POST /api/v1/ai/match":            "Match penpal",
			"POST /api/v1/ai/reply":            "Generate AI reply",
			"GET  /api/v1/ai/personas":         "Get AI personas",
		},
		"scheduler": map[string]string{
			"GET  /api/v1/scheduler/tasks":     "Get scheduled tasks",
			"POST /api/v1/scheduler/tasks":     "Create scheduled task",
			"PUT  /api/v1/scheduler/tasks/:id/enable": "Enable task",
			"GET  /api/v1/scheduler/stats":     "Get scheduler statistics",
		},
		"admin": map[string]string{
			"GET /api/v1/admin/dashboard/stats": "Get dashboard statistics",
			"GET /api/v1/admin/users":          "Manage users",
			"GET /api/v1/admin/courier/applications": "Manage courier applications",
			"GET /api/v1/admin/museum/pending": "Review pending museum entries",
		},
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"endpoints": endpoints,
		"total_endpoints": countEndpoints(endpoints),
		"documentation": gin.H{
			"swagger_ui": "/swagger/index.html",
			"openapi_spec": "/api/v1/openapi.json",
		},
	})
}

// Helper function to count total endpoints
func countEndpoints(endpoints map[string]interface{}) int {
	total := 0
	for _, category := range endpoints {
		if categoryMap, ok := category.(map[string]string); ok {
			total += len(categoryMap)
		}
	}
	return total
}