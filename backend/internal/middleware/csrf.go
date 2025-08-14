package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	CSRFTokenLength   = 32
	CSRFCookieName    = "csrf-token"
	CSRFHeaderName    = "X-CSRF-Token"
	CSRFFormFieldName = "_csrf_token"
	CSRFTokenTTL      = 24 * time.Hour
)

// CSRFToken represents a CSRF token with metadata
type CSRFToken struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	IssuedAt  time.Time `json:"issued_at"`
}

// generateCSRFToken generates a cryptographically secure CSRF token
func generateCSRFToken() (string, error) {
	bytes := make([]byte, CSRFTokenLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// CSRFMiddleware provides CSRF protection for state-changing operations
func CSRFMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip CSRF protection for safe methods (GET, HEAD, OPTIONS)
		if c.Request.Method == http.MethodGet ||
			c.Request.Method == http.MethodHead ||
			c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		// Skip for certain endpoints that don't need CSRF protection
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api/v1/auth/csrf") ||
			strings.HasPrefix(path, "/health") ||
			strings.HasPrefix(path, "/ping") {
			c.Next()
			return
		}

		// Get CSRF token from various sources
		token := getCSRFTokenFromRequest(c)
		if token == "" {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "CSRF_TOKEN_MISSING",
				"message": "CSRF token is required for this operation",
				"code":    http.StatusForbidden,
			})
			c.Abort()
			return
		}

		// Validate CSRF token
		cookieToken, err := c.Cookie(CSRFCookieName)
		if err != nil || cookieToken == "" {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "CSRF_COOKIE_MISSING",
				"message": "CSRF cookie is missing or expired",
				"code":    http.StatusForbidden,
			})
			c.Abort()
			return
		}

		// Compare tokens (constant time comparison)
		if !secureCompare(token, cookieToken) {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "CSRF_TOKEN_INVALID",
				"message": "CSRF token validation failed",
				"code":    http.StatusForbidden,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// getCSRFTokenFromRequest extracts CSRF token from request
func getCSRFTokenFromRequest(c *gin.Context) string {
	// Try header first
	if token := c.GetHeader(CSRFHeaderName); token != "" {
		return token
	}

	// Try form data
	if token := c.PostForm(CSRFFormFieldName); token != "" {
		return token
	}

	// Try JSON body (if applicable)
	if c.ContentType() == "application/json" {
		if token := c.GetString("csrf_token"); token != "" {
			return token
		}
	}

	return ""
}

// secureCompare performs constant-time comparison to prevent timing attacks
func secureCompare(a, b string) bool {
	if len(a) != len(b) {
		return false
	}

	result := byte(0)
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}

	return result == 0
}

// CSRFHandler provides CSRF token generation endpoint
type CSRFHandler struct{}

func NewCSRFHandler() *CSRFHandler {
	return &CSRFHandler{}
}

// GetCSRFToken generates and returns a new CSRF token
func (h *CSRFHandler) GetCSRFToken(c *gin.Context) {
	// Generate new CSRF token
	token, err := generateCSRFToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "TOKEN_GENERATION_FAILED",
			"message": "Failed to generate CSRF token",
		})
		return
	}

	// Set secure cookie with proper settings for development
	expiresAt := time.Now().Add(CSRFTokenTTL)

	// 开发环境使用更宽松的SameSite设置
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		CSRFCookieName,
		token,
		int(CSRFTokenTTL.Seconds()),
		"/",
		"",    // Domain - let browser set automatically
		false, // Secure - set to true in production with HTTPS
		false, // HttpOnly - 设为false让前端可以读取CSRF cookie
	)

	// Return token info
	csrfToken := CSRFToken{
		Token:     token,
		ExpiresAt: expiresAt,
		IssuedAt:  time.Now(),
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "CSRF token generated successfully",
		"data": gin.H{
			"token":      csrfToken.Token,
			"expires_at": csrfToken.ExpiresAt.Unix(),
			"expires_in": int(CSRFTokenTTL.Seconds()),
		},
	})
}

// RefreshCSRFToken refreshes an existing CSRF token
func (h *CSRFHandler) RefreshCSRFToken(c *gin.Context) {
	// Generate new token
	h.GetCSRFToken(c)
}

// GetCSRFTokenHandler is a simple handler function for getting CSRF tokens
func GetCSRFTokenHandler(c *gin.Context) {
	handler := NewCSRFHandler()
	handler.GetCSRFToken(c)
}
