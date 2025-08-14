package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// EnhancedCSRFMiddleware provides SOTA CSRF protection with proper exclusions
func EnhancedCSRFMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip CSRF for safe methods
		if c.Request.Method == http.MethodGet ||
			c.Request.Method == http.MethodHead ||
			c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		// Define endpoints that don't need CSRF protection
		csrfExemptPaths := []string{
			"/health",
			"/ping",
			"/api/v1/auth/csrf",     // Getting CSRF token
			"/api/v1/auth/register", // Initial registration needs special handling
		}

		// Check if current path is exempt
		path := c.Request.URL.Path
		isExempt := false
		for _, exemptPath := range csrfExemptPaths {
			if strings.HasPrefix(path, exemptPath) {
				isExempt = true
				break
			}
		}

		if isExempt {
			c.Next()
			return
		}

		// Special handling for login - allow first login without CSRF for session establishment
		if path == "/api/v1/auth/login" {
			// Check if this is an initial login (no existing session)
			if _, err := c.Cookie("session-established"); err != nil {
				// First login - establish session
				c.SetCookie(
					"session-established",
					"true",
					3600, // 1 hour
					"/",
					"",
					false,
					true, // HttpOnly
				)
				// Generate and set CSRF token for future requests
				if token, err := generateCSRFToken(); err == nil {
					setCSRFCookie(c, token)
				}
				c.Next()
				return
			}
			// Subsequent logins require CSRF
		}

		// Validate CSRF token
		token := getCSRFTokenFromRequest(c)
		if token == "" {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "CSRF_TOKEN_MISSING",
				"message": "CSRF token is required for this operation",
			})
			c.Abort()
			return
		}

		// Get cookie token
		cookieToken, err := c.Cookie(CSRFCookieName)
		if err != nil || cookieToken == "" {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "CSRF_COOKIE_MISSING",
				"message": "CSRF cookie is missing. Please refresh and try again.",
			})
			c.Abort()
			return
		}

		// Validate tokens match
		if !secureCompare(token, cookieToken) {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "CSRF_TOKEN_INVALID",
				"message": "CSRF token validation failed",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// setCSRFCookie sets the CSRF cookie with proper settings
func setCSRFCookie(c *gin.Context, token string) {
	isDev := strings.Contains(c.Request.Host, "localhost") ||
		strings.Contains(c.Request.Host, "127.0.0.1")

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		CSRFCookieName,
		token,
		int(CSRFTokenTTL.Seconds()),
		"/",
		"",
		!isDev, // Secure only in production
		false,  // Not HttpOnly so JS can read it
	)
}

// SOTACSRFProtection provides a complete CSRF protection setup
type SOTACSRFProtection struct {
	exemptPaths []string
	tokenTTL    time.Duration
}

// NewSOTACSRFProtection creates a new CSRF protection instance
func NewSOTACSRFProtection() *SOTACSRFProtection {
	return &SOTACSRFProtection{
		exemptPaths: []string{
			"/health",
			"/ping",
			"/api/v1/auth/csrf",
		},
		tokenTTL: 24 * time.Hour,
	}
}

// Middleware returns the CSRF protection middleware
func (csrf *SOTACSRFProtection) Middleware() gin.HandlerFunc {
	return EnhancedCSRFMiddleware()
}

// TokenHandler handles CSRF token generation requests
func (csrf *SOTACSRFProtection) TokenHandler(c *gin.Context) {
	token, err := generateCSRFToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "TOKEN_GENERATION_FAILED",
		})
		return
	}

	// Set cookie
	setCSRFCookie(c, token)

	// Return token info
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"token":      token,
			"expires_in": int(csrf.tokenTTL.Seconds()),
		},
	})
}
