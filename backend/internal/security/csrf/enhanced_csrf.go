package csrf

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// EnhancedCSRF provides improved CSRF protection with double-submit cookie pattern
type EnhancedCSRF struct {
	mu           sync.RWMutex
	tokenStore   map[string]*TokenInfo
	cookieName   string
	headerName   string
	fieldName    string
	secure       bool
	sameSite     http.SameSite
	maxAge       time.Duration
	cleanupTime  time.Duration
}

// TokenInfo stores CSRF token metadata
type TokenInfo struct {
	Token     string
	CreatedAt time.Time
	UsedCount int
}

// Config for enhanced CSRF protection
type Config struct {
	CookieName  string
	HeaderName  string
	FieldName   string
	Secure      bool
	SameSite    http.SameSite
	MaxAge      time.Duration
	CleanupTime time.Duration
}

// DefaultConfig returns secure default configuration
func DefaultConfig() *Config {
	return &Config{
		CookieName:  "csrf_token",
		HeaderName:  "X-CSRF-Token",
		FieldName:   "csrf_token",
		Secure:      true,
		SameSite:    http.SameSiteStrictMode,
		MaxAge:      24 * time.Hour,
		CleanupTime: 1 * time.Hour,
	}
}

// NewEnhancedCSRF creates a new enhanced CSRF protection instance
func NewEnhancedCSRF(config *Config) *EnhancedCSRF {
	if config == nil {
		config = DefaultConfig()
	}

	csrf := &EnhancedCSRF{
		tokenStore:  make(map[string]*TokenInfo),
		cookieName:  config.CookieName,
		headerName:  config.HeaderName,
		fieldName:   config.FieldName,
		secure:      config.Secure,
		sameSite:    config.SameSite,
		maxAge:      config.MaxAge,
		cleanupTime: config.CleanupTime,
	}

	// Start cleanup routine
	go csrf.cleanupRoutine()

	return csrf
}

// Middleware returns the Gin middleware function
func (c *EnhancedCSRF) Middleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Generate token for safe methods
		if c.isSafeMethod(ctx.Request.Method) {
			token := c.generateOrGetToken(ctx)
			ctx.Set("csrf_token", token)
			ctx.Next()
			return
		}

		// Validate token for state-changing methods
		if !c.validateRequest(ctx) {
			ctx.JSON(http.StatusForbidden, gin.H{
				"error": "CSRF token validation failed",
				"code":  "CSRF_VALIDATION_FAILED",
			})
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}

// generateOrGetToken generates a new token or retrieves existing one
func (c *EnhancedCSRF) generateOrGetToken(ctx *gin.Context) string {
	// Check for existing valid token in cookie
	if cookie, err := ctx.Cookie(c.cookieName); err == nil && cookie != "" {
		c.mu.RLock()
		if info, exists := c.tokenStore[cookie]; exists {
			c.mu.RUnlock()
			if time.Since(info.CreatedAt) < c.maxAge {
				return cookie
			}
		} else {
			c.mu.RUnlock()
		}
	}

	// Generate new token
	token := c.generateToken()
	
	// Store token info
	c.mu.Lock()
	c.tokenStore[token] = &TokenInfo{
		Token:     token,
		CreatedAt: time.Now(),
		UsedCount: 0,
	}
	c.mu.Unlock()

	// Set cookie with security flags
	isSecure := c.secure
	if strings.Contains(ctx.Request.Host, "localhost") {
		isSecure = false // Allow non-HTTPS in local development
	}

	ctx.SetSameSite(c.sameSite)
	ctx.SetCookie(
		c.cookieName,
		token,
		int(c.maxAge.Seconds()),
		"/",
		"",
		isSecure,
		true, // HttpOnly
	)

	return token
}

// generateToken creates a cryptographically secure token
func (c *EnhancedCSRF) generateToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic("failed to generate CSRF token: " + err.Error())
	}
	return base64.URLEncoding.EncodeToString(b)
}

// validateRequest validates CSRF token in the request
func (c *EnhancedCSRF) validateRequest(ctx *gin.Context) bool {
	// Get token from cookie (double-submit cookie pattern)
	cookieToken, err := ctx.Cookie(c.cookieName)
	if err != nil || cookieToken == "" {
		return false
	}

	// Get token from request (header, form, or JSON)
	requestToken := c.extractRequestToken(ctx)
	if requestToken == "" {
		return false
	}

	// Constant-time comparison
	if subtle.ConstantTimeCompare([]byte(cookieToken), []byte(requestToken)) != 1 {
		return false
	}

	// Validate token exists and is not expired
	c.mu.RLock()
	info, exists := c.tokenStore[cookieToken]
	c.mu.RUnlock()

	if !exists {
		return false
	}

	if time.Since(info.CreatedAt) > c.maxAge {
		// Token expired
		c.mu.Lock()
		delete(c.tokenStore, cookieToken)
		c.mu.Unlock()
		return false
	}

	// Update usage count
	c.mu.Lock()
	info.UsedCount++
	c.mu.Unlock()

	return true
}

// extractRequestToken extracts CSRF token from various sources
func (c *EnhancedCSRF) extractRequestToken(ctx *gin.Context) string {
	// 1. Check header
	if token := ctx.GetHeader(c.headerName); token != "" {
		return token
	}

	// 2. Check form data
	if token := ctx.PostForm(c.fieldName); token != "" {
		return token
	}

	// 3. Check JSON body
	var jsonData map[string]interface{}
	if ctx.ContentType() == "application/json" {
		if err := ctx.ShouldBindJSON(&jsonData); err == nil {
			if token, ok := jsonData[c.fieldName].(string); ok {
				// Re-bind body for subsequent handlers
				ctx.Request.Body = nil
				return token
			}
		}
	}

	return ""
}

// isSafeMethod checks if the HTTP method is safe (doesn't change state)
func (c *EnhancedCSRF) isSafeMethod(method string) bool {
	return method == http.MethodGet || 
		   method == http.MethodHead || 
		   method == http.MethodOptions || 
		   method == http.MethodTrace
}

// cleanupRoutine periodically removes expired tokens
func (c *EnhancedCSRF) cleanupRoutine() {
	ticker := time.NewTicker(c.cleanupTime)
	defer ticker.Stop()

	for range ticker.C {
		c.cleanup()
	}
}

// cleanup removes expired tokens from memory
func (c *EnhancedCSRF) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for token, info := range c.tokenStore {
		if now.Sub(info.CreatedAt) > c.maxAge {
			delete(c.tokenStore, token)
		}
	}
}

// GetToken returns the current CSRF token for a context
func (c *EnhancedCSRF) GetToken(ctx *gin.Context) string {
	if token, exists := ctx.Get("csrf_token"); exists {
		return token.(string)
	}
	return c.generateOrGetToken(ctx)
}

// ExemptPaths allows certain paths to bypass CSRF protection
type ExemptPaths struct {
	paths map[string]bool
	mu    sync.RWMutex
}

// NewExemptPaths creates a new exempt paths manager
func NewExemptPaths(paths ...string) *ExemptPaths {
	ep := &ExemptPaths{
		paths: make(map[string]bool),
	}
	for _, path := range paths {
		ep.paths[path] = true
	}
	return ep
}

// IsExempt checks if a path is exempt from CSRF protection
func (ep *ExemptPaths) IsExempt(path string) bool {
	ep.mu.RLock()
	defer ep.mu.RUnlock()
	return ep.paths[path]
}

// MiddlewareWithExemptions returns middleware with path exemptions
func (c *EnhancedCSRF) MiddlewareWithExemptions(exempt *ExemptPaths) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Skip CSRF for exempt paths
		if exempt != nil && exempt.IsExempt(ctx.Request.URL.Path) {
			ctx.Next()
			return
		}

		// Use regular middleware
		c.Middleware()(ctx)
	}
}