// Package servicemesh provides zero-trust security gateway functionality
package servicemesh

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"openpenpal-backend/internal/config"

	"github.com/golang-jwt/jwt/v5"
)

// ZeroTrustGateway provides zero-trust security for service mesh
type ZeroTrustGateway struct {
	// Security policies
	policies map[string]*SecurityPolicy
	
	// Authentication providers
	authProviders map[string]AuthenticationProvider
	
	// Authorization engine
	authzEngine *AuthorizationEngine
	
	// Threat detection system
	threatDetector *ThreatDetector
	
	// Rate limiting
	rateLimiter *RateLimiter
	
	// Security audit logger
	auditLogger *SecurityAuditLogger
	
	// IP reputation manager
	ipReputation *IPReputationManager
	
	// Configuration
	config *ZeroTrustConfig
	
	// Mutex for thread safety
	mu sync.RWMutex
}

// SecurityPolicy defines security policies for services
type SecurityPolicy struct {
	ID                  string                 `json:"id"`
	ServiceID           string                 `json:"service_id"`
	Name                string                 `json:"name"`
	Enabled             bool                   `json:"enabled"`
	
	// Authentication requirements
	RequireAuthentication bool                  `json:"require_authentication"`
	AllowedAuthMethods    []string              `json:"allowed_auth_methods"`
	
	// Authorization rules
	AuthorizationRules    []*AuthorizationRule  `json:"authorization_rules"`
	
	// Rate limiting
	RateLimitRules        []*RateLimitRule      `json:"rate_limit_rules"`
	
	// IP restrictions
	AllowedIPs            []string              `json:"allowed_ips"`
	BlockedIPs            []string              `json:"blocked_ips"`
	
	// Threat detection
	ThreatDetectionRules  []*ThreatDetectionRule `json:"threat_detection_rules"`
	
	// Security headers
	SecurityHeaders       map[string]string      `json:"security_headers"`
	
	// Encryption requirements
	RequireEncryption     bool                   `json:"require_encryption"`
	MinTLSVersion         string                 `json:"min_tls_version"`
	
	// Metadata
	CreatedAt             time.Time              `json:"created_at"`
	UpdatedAt             time.Time              `json:"updated_at"`
	CreatedBy             string                 `json:"created_by"`
}

// AuthorizationRule defines authorization logic
type AuthorizationRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Condition   string                 `json:"condition"`
	Action      string                 `json:"action"` // allow, deny, challenge
	Priority    int                    `json:"priority"`
	Resources   []string               `json:"resources"`
	Methods     []string               `json:"methods"`
	Roles       []string               `json:"roles"`
	Permissions []string               `json:"permissions"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// RateLimitRule defines rate limiting logic
type RateLimitRule struct {
	ID            string        `json:"id"`
	Name          string        `json:"name"`
	Scope         string        `json:"scope"` // global, user, ip, service
	Limit         int           `json:"limit"`
	Window        time.Duration `json:"window"`
	BurstLimit    int           `json:"burst_limit"`
	Action        string        `json:"action"` // block, throttle, log
	Exemptions    []string      `json:"exemptions"`
}

// ThreatDetectionRule defines threat detection logic
type ThreatDetectionRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"` // sql_injection, xss, brute_force, anomaly
	Pattern     string                 `json:"pattern"`
	Threshold   float64                `json:"threshold"`
	Action      string                 `json:"action"` // block, log, alert, challenge
	Severity    string                 `json:"severity"`
	Enabled     bool                   `json:"enabled"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ZeroTrustConfig holds zero-trust gateway configuration
type ZeroTrustConfig struct {
	// Authentication settings
	JWTSecret                   string        `json:"jwt_secret"`
	JWTExpiration               time.Duration `json:"jwt_expiration"`
	RefreshTokenExpiration      time.Duration `json:"refresh_token_expiration"`
	
	// Rate limiting defaults
	DefaultRateLimit            int           `json:"default_rate_limit"`
	DefaultRateLimitWindow      time.Duration `json:"default_rate_limit_window"`
	
	// Threat detection settings
	EnableThreatDetection       bool          `json:"enable_threat_detection"`
	ThreatDetectionSensitivity  float64       `json:"threat_detection_sensitivity"`
	
	// IP reputation settings
	EnableIPReputation          bool          `json:"enable_ip_reputation"`
	IPReputationCacheSize       int           `json:"ip_reputation_cache_size"`
	IPReputationCacheTTL        time.Duration `json:"ip_reputation_cache_ttl"`
	
	// Security audit settings
	EnableSecurityAudit         bool          `json:"enable_security_audit"`
	AuditLogRetentionPeriod     time.Duration `json:"audit_log_retention_period"`
	
	// General settings
	EnableMutualTLS             bool          `json:"enable_mutual_tls"`
	RequireHTTPS                bool          `json:"require_https"`
	SessionTimeout              time.Duration `json:"session_timeout"`
	MaxConcurrentSessions       int           `json:"max_concurrent_sessions"`
}

// AuthenticationProvider defines interface for authentication providers
type AuthenticationProvider interface {
	Authenticate(ctx context.Context, request *AuthenticationRequest) (*AuthenticationResult, error)
	ValidateToken(ctx context.Context, token string) (*TokenValidationResult, error)
	RefreshToken(ctx context.Context, refreshToken string) (*TokenRefreshResult, error)
	Name() string
}

// AuthenticationRequest represents an authentication request
type AuthenticationRequest struct {
	Method      string                 `json:"method"`
	Credentials map[string]interface{} `json:"credentials"`
	Metadata    map[string]interface{} `json:"metadata"`
	ClientIP    string                 `json:"client_ip"`
	UserAgent   string                 `json:"user_agent"`
	Timestamp   time.Time              `json:"timestamp"`
}

// AuthenticationResult represents the result of authentication
type AuthenticationResult struct {
	Success      bool                   `json:"success"`
	UserID       string                 `json:"user_id"`
	Username     string                 `json:"username"`
	Roles        []string               `json:"roles"`
	Permissions  []string               `json:"permissions"`
	Token        string                 `json:"token"`
	RefreshToken string                 `json:"refresh_token"`
	ExpiresAt    time.Time              `json:"expires_at"`
	Metadata     map[string]interface{} `json:"metadata"`
	Error        string                 `json:"error"`
}

// TokenValidationResult represents token validation result
type TokenValidationResult struct {
	Valid       bool                   `json:"valid"`
	UserID      string                 `json:"user_id"`
	Username    string                 `json:"username"`
	Roles       []string               `json:"roles"`
	Permissions []string               `json:"permissions"`
	ExpiresAt   time.Time              `json:"expires_at"`
	Metadata    map[string]interface{} `json:"metadata"`
	Error       string                 `json:"error"`
}

// TokenRefreshResult represents token refresh result
type TokenRefreshResult struct {
	Success      bool      `json:"success"`
	NewToken     string    `json:"new_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	Error        string    `json:"error"`
}

// AuthorizationEngine handles authorization decisions
type AuthorizationEngine struct {
	policies    map[string]*SecurityPolicy
	mu          sync.RWMutex
}

// ThreatDetector detects security threats
type ThreatDetector struct {
	rules           map[string]*ThreatDetectionRule
	anomalyDetector *AnomalyDetector
	patternMatcher  *PatternMatcher
	behaviorAI      *BehaviorAnalysisAI
	mu              sync.RWMutex
}

// RateLimiter implements rate limiting functionality
type RateLimiter struct {
	rules      map[string]*RateLimitRule
	counters   map[string]*RateLimitCounter
	mu         sync.RWMutex
}

// RateLimitCounter tracks rate limit counters
type RateLimitCounter struct {
	Limit       int           `json:"limit"`
	Remaining   int           `json:"remaining"`
	ResetTime   time.Time     `json:"reset_time"`
	Window      time.Duration `json:"window"`
	LastUpdate  time.Time     `json:"last_update"`
}

// SecurityAuditLogger logs security events
type SecurityAuditLogger struct {
	events   []SecurityEvent
	mu       sync.RWMutex
	maxEvents int
}

// SecurityEvent represents a security audit event
type SecurityEvent struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Severity    string                 `json:"severity"`
	UserID      string                 `json:"user_id"`
	ServiceID   string                 `json:"service_id"`
	ClientIP    string                 `json:"client_ip"`
	UserAgent   string                 `json:"user_agent"`
	Action      string                 `json:"action"`
	Resource    string                 `json:"resource"`
	Result      string                 `json:"result"`
	Message     string                 `json:"message"`
	Details     map[string]interface{} `json:"details"`
	Timestamp   time.Time              `json:"timestamp"`
}

// IPReputationManager manages IP reputation data
type IPReputationManager struct {
	reputation   map[string]*IPReputation
	cache        map[string]*CachedReputation
	cacheSize    int
	cacheTTL     time.Duration
	mu           sync.RWMutex
}

// IPReputation represents IP reputation data
type IPReputation struct {
	IP              string                 `json:"ip"`
	Score           float64                `json:"score"` // 0-1, where 1 is good
	Category        string                 `json:"category"`
	ThreatTypes     []string               `json:"threat_types"`
	FirstSeen       time.Time              `json:"first_seen"`
	LastSeen        time.Time              `json:"last_seen"`
	RequestCount    int64                  `json:"request_count"`
	ViolationCount  int64                  `json:"violation_count"`
	Metadata        map[string]interface{} `json:"metadata"`
	LastUpdated     time.Time              `json:"last_updated"`
}

// CachedReputation represents cached reputation data
type CachedReputation struct {
	Reputation *IPReputation `json:"reputation"`
	CachedAt   time.Time     `json:"cached_at"`
}

// BehaviorAnalysisAI provides AI-powered behavior analysis
type BehaviorAnalysisAI struct {
	userProfiles    map[string]*UserBehaviorProfile
	anomalyScores   map[string]float64
	learningRate    float64
	mu              sync.RWMutex
}

// UserBehaviorProfile represents user behavior patterns
type UserBehaviorProfile struct {
	UserID              string                 `json:"user_id"`
	TypicalHours        [24]float64            `json:"typical_hours"`
	TypicalDays         [7]float64             `json:"typical_days"`
	TypicalIPs          map[string]float64     `json:"typical_ips"`
	TypicalUserAgents   map[string]float64     `json:"typical_user_agents"`
	TypicalResources    map[string]float64     `json:"typical_resources"`
	RequestPatterns     []RequestPattern       `json:"request_patterns"`
	AnomalyScore        float64                `json:"anomaly_score"`
	LastUpdated         time.Time              `json:"last_updated"`
	Metadata            map[string]interface{} `json:"metadata"`
}

// RequestPattern represents a request pattern
type RequestPattern struct {
	Method        string    `json:"method"`
	Path          string    `json:"path"`
	Frequency     float64   `json:"frequency"`
	AvgInterval   float64   `json:"avg_interval"`
	LastSeen      time.Time `json:"last_seen"`
}

// PatternMatcher matches security threat patterns
type PatternMatcher struct {
	sqlInjectionPatterns []string
	xssPatterns          []string
	commandInjectionPatterns []string
	pathTraversalPatterns []string
}

// NewZeroTrustGateway creates a new zero-trust security gateway
func NewZeroTrustGateway(cfg *config.Config) *ZeroTrustGateway {
	config := &ZeroTrustConfig{
		JWTSecret:                   cfg.JWTSecret,
		JWTExpiration:               time.Duration(cfg.JWTExpiry) * time.Hour,
		RefreshTokenExpiration:      24 * time.Hour,
		DefaultRateLimit:            1000,
		DefaultRateLimitWindow:      time.Hour,
		EnableThreatDetection:       true,
		ThreatDetectionSensitivity:  0.7,
		EnableIPReputation:          true,
		IPReputationCacheSize:       10000,
		IPReputationCacheTTL:        time.Hour,
		EnableSecurityAudit:         true,
		AuditLogRetentionPeriod:     30 * 24 * time.Hour,
		EnableMutualTLS:             false,
		RequireHTTPS:                cfg.Environment == "production",
		SessionTimeout:              8 * time.Hour,
		MaxConcurrentSessions:       5,
	}

	gateway := &ZeroTrustGateway{
		policies:       make(map[string]*SecurityPolicy),
		authProviders:  make(map[string]AuthenticationProvider),
		authzEngine:    NewAuthorizationEngine(),
		threatDetector: NewThreatDetector(),
		rateLimiter:    NewRateLimiter(),
		auditLogger:    NewSecurityAuditLogger(1000),
		ipReputation:   NewIPReputationManager(config.IPReputationCacheSize, config.IPReputationCacheTTL),
		config:         config,
	}

	// Register built-in authentication providers
	gateway.registerAuthProviders()

	// Initialize default security policies
	gateway.initializeDefaultPolicies()

	return gateway
}

// NewAuthorizationEngine creates a new authorization engine
func NewAuthorizationEngine() *AuthorizationEngine {
	return &AuthorizationEngine{
		policies: make(map[string]*SecurityPolicy),
	}
}

// NewThreatDetector creates a new threat detector
func NewThreatDetector() *ThreatDetector {
	return &ThreatDetector{
		rules:           make(map[string]*ThreatDetectionRule),
		anomalyDetector: &AnomalyDetector{},
		patternMatcher:  NewPatternMatcher(),
		behaviorAI:      NewBehaviorAnalysisAI(),
	}
}

// NewPatternMatcher creates a new pattern matcher
func NewPatternMatcher() *PatternMatcher {
	return &PatternMatcher{
		sqlInjectionPatterns: []string{
			"UNION.*SELECT",
			"DROP.*TABLE",
			"INSERT.*INTO",
			"'.*OR.*'.*'",
			"--.*",
			"/\\*.*\\*/",
		},
		xssPatterns: []string{
			"<script.*>",
			"javascript:",
			"onload=",
			"onerror=",
			"onclick=",
			"<iframe.*>",
		},
		commandInjectionPatterns: []string{
			";.*rm.*",
			"\\|.*cat.*",
			"&&.*",
			"`.*`",
			"$\\(.*\\)",
		},
		pathTraversalPatterns: []string{
			"\\.\\./",
			"\\.\\.\\\\",
			"/etc/passwd",
			"/proc/",
			"..%2F",
		},
	}
}

// NewBehaviorAnalysisAI creates a new behavior analysis AI
func NewBehaviorAnalysisAI() *BehaviorAnalysisAI {
	return &BehaviorAnalysisAI{
		userProfiles:  make(map[string]*UserBehaviorProfile),
		anomalyScores: make(map[string]float64),
		learningRate:  0.01,
	}
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		rules:    make(map[string]*RateLimitRule),
		counters: make(map[string]*RateLimitCounter),
	}
}

// NewSecurityAuditLogger creates a new security audit logger
func NewSecurityAuditLogger(maxEvents int) *SecurityAuditLogger {
	return &SecurityAuditLogger{
		events:    make([]SecurityEvent, 0),
		maxEvents: maxEvents,
	}
}

// NewIPReputationManager creates a new IP reputation manager
func NewIPReputationManager(cacheSize int, cacheTTL time.Duration) *IPReputationManager {
	return &IPReputationManager{
		reputation: make(map[string]*IPReputation),
		cache:      make(map[string]*CachedReputation),
		cacheSize:  cacheSize,
		cacheTTL:   cacheTTL,
	}
}

// registerAuthProviders registers built-in authentication providers
func (ztg *ZeroTrustGateway) registerAuthProviders() {
	ztg.authProviders["jwt"] = &JWTAuthenticationProvider{
		secret:    []byte(ztg.config.JWTSecret),
		expiresIn: ztg.config.JWTExpiration,
	}
}

// initializeDefaultPolicies initializes default security policies
func (ztg *ZeroTrustGateway) initializeDefaultPolicies() {
	// Create default policy for all services
	defaultPolicy := &SecurityPolicy{
		ID:                    "default",
		ServiceID:             "*",
		Name:                  "Default Security Policy",
		Enabled:               true,
		RequireAuthentication: true,
		AllowedAuthMethods:    []string{"jwt"},
		AuthorizationRules: []*AuthorizationRule{
			{
				ID:        "default_auth",
				Name:      "Default Authorization",
				Condition: "authenticated",
				Action:    "allow",
				Priority:  1000,
			},
		},
		RateLimitRules: []*RateLimitRule{
			{
				ID:         "default_rate_limit",
				Name:       "Default Rate Limit",
				Scope:      "user",
				Limit:      ztg.config.DefaultRateLimit,
				Window:     ztg.config.DefaultRateLimitWindow,
				BurstLimit: ztg.config.DefaultRateLimit / 10,
				Action:     "throttle",
			},
		},
		ThreatDetectionRules: []*ThreatDetectionRule{
			{
				ID:        "sql_injection",
				Name:      "SQL Injection Detection",
				Type:      "sql_injection",
				Threshold: 0.8,
				Action:    "block",
				Severity:  "high",
				Enabled:   true,
			},
			{
				ID:        "xss_detection",
				Name:      "XSS Detection",
				Type:      "xss",
				Threshold: 0.8,
				Action:    "block",
				Severity:  "high",
				Enabled:   true,
			},
		},
		SecurityHeaders: map[string]string{
			"X-Content-Type-Options": "nosniff",
			"X-Frame-Options":        "DENY",
			"X-XSS-Protection":       "1; mode=block",
			"Strict-Transport-Security": "max-age=31536000; includeSubDomains",
		},
		RequireEncryption: ztg.config.RequireHTTPS,
		MinTLSVersion:     "1.2",
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		CreatedBy:         "system",
	}

	ztg.policies["default"] = defaultPolicy
	ztg.authzEngine.policies["default"] = defaultPolicy
}

// AuthenticateRequest authenticates an incoming request
func (ztg *ZeroTrustGateway) AuthenticateRequest(ctx context.Context, request *SecurityRequest) (*SecurityResponse, error) {
	response := &SecurityResponse{
		Allowed:   false,
		ServiceID: request.ServiceID,
		Timestamp: time.Now(),
		Details:   make(map[string]interface{}),
	}

	// Get security policy for the service
	policy := ztg.getSecurityPolicy(request.ServiceID)
	if policy == nil {
		policy = ztg.policies["default"]
	}

	// Check IP reputation
	if ztg.config.EnableIPReputation {
		if !ztg.checkIPReputation(request.ClientIP) {
			response.Reason = "IP reputation check failed"
			ztg.logSecurityEvent("ip_reputation_blocked", "high", request)
			return response, nil
		}
	}

	// Check rate limiting
	if !ztg.checkRateLimit(request, policy) {
		response.Reason = "Rate limit exceeded"
		ztg.logSecurityEvent("rate_limit_exceeded", "medium", request)
		return response, nil
	}

	// Perform threat detection
	if ztg.config.EnableThreatDetection {
		if threat := ztg.detectThreats(request, policy); threat != nil {
			response.Reason = fmt.Sprintf("Threat detected: %s", threat.Type)
			response.ThreatInfo = threat
			ztg.logSecurityEvent("threat_detected", threat.Severity, request)
			return response, nil
		}
	}

	// Authenticate if required
	if policy.RequireAuthentication {
		authResult, err := ztg.authenticate(ctx, request, policy)
		if err != nil || !authResult.Success {
			response.Reason = "Authentication failed"
			if authResult != nil {
				response.Details["auth_error"] = authResult.Error
			}
			ztg.logSecurityEvent("authentication_failed", "medium", request)
			return response, nil
		}

		// Store authentication info in response
		response.UserID = authResult.UserID
		response.Username = authResult.Username
		response.Roles = authResult.Roles
		response.Permissions = authResult.Permissions
		response.Details["auth_result"] = authResult
	}

	// Authorize the request
	authzResult := ztg.authorize(request, policy, response)
	if !authzResult.Allowed {
		response.Reason = authzResult.Reason
		ztg.logSecurityEvent("authorization_failed", "medium", request)
		return response, nil
	}

	// Update behavior analysis
	if response.UserID != "" {
		ztg.threatDetector.behaviorAI.UpdateBehavior(response.UserID, request)
	}

	// Request is allowed
	response.Allowed = true
	response.SecurityHeaders = policy.SecurityHeaders
	
	ztg.logSecurityEvent("request_allowed", "info", request)

	return response, nil
}

// GetAuditLogger returns the security audit logger instance
func (ztg *ZeroTrustGateway) GetAuditLogger() *SecurityAuditLogger {
	return ztg.auditLogger
}

// SecurityRequest represents a security validation request
type SecurityRequest struct {
	ServiceID     string                 `json:"service_id"`
	Method        string                 `json:"method"`
	Path          string                 `json:"path"`
	Headers       map[string]string      `json:"headers"`
	Body          string                 `json:"body"`
	ClientIP      string                 `json:"client_ip"`
	UserAgent     string                 `json:"user_agent"`
	Timestamp     time.Time              `json:"timestamp"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// SecurityResponse represents a security validation response
type SecurityResponse struct {
	Allowed         bool                   `json:"allowed"`
	Reason          string                 `json:"reason"`
	ServiceID       string                 `json:"service_id"`
	UserID          string                 `json:"user_id"`
	Username        string                 `json:"username"`
	Roles           []string               `json:"roles"`
	Permissions     []string               `json:"permissions"`
	SecurityHeaders map[string]string      `json:"security_headers"`
	ThreatInfo      *ThreatInfo           `json:"threat_info"`
	RateLimitInfo   *RateLimitInfo        `json:"rate_limit_info"`
	Timestamp       time.Time              `json:"timestamp"`
	Details         map[string]interface{} `json:"details"`
}

// ThreatInfo represents detected threat information
type ThreatInfo struct {
	Type        string                 `json:"type"`
	Severity    string                 `json:"severity"`
	Confidence  float64                `json:"confidence"`
	Description string                 `json:"description"`
	Patterns    []string               `json:"patterns"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// RateLimitInfo represents rate limiting information
type RateLimitInfo struct {
	Limit     int       `json:"limit"`
	Remaining int       `json:"remaining"`
	ResetTime time.Time `json:"reset_time"`
	Window    string    `json:"window"`
}

// AuthorizationResult represents authorization result
type AuthorizationResult struct {
	Allowed     bool                   `json:"allowed"`
	Reason      string                 `json:"reason"`
	MatchedRule *AuthorizationRule     `json:"matched_rule"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// getSecurityPolicy gets the security policy for a service
func (ztg *ZeroTrustGateway) getSecurityPolicy(serviceID string) *SecurityPolicy {
	ztg.mu.RLock()
	defer ztg.mu.RUnlock()

	// Try exact match first
	if policy, exists := ztg.policies[serviceID]; exists && policy.Enabled {
		return policy
	}

	// Return default policy
	return ztg.policies["default"]
}

// checkIPReputation checks IP reputation
func (ztg *ZeroTrustGateway) checkIPReputation(clientIP string) bool {
	reputation := ztg.ipReputation.GetReputation(clientIP)
	if reputation == nil {
		// Unknown IP, allow but start tracking
		ztg.ipReputation.InitializeIP(clientIP)
		return true
	}

	// Block IPs with low reputation score
	return reputation.Score > 0.3
}

// checkRateLimit checks rate limiting
func (ztg *ZeroTrustGateway) checkRateLimit(request *SecurityRequest, policy *SecurityPolicy) bool {
	for _, rule := range policy.RateLimitRules {
		if !ztg.rateLimiter.CheckLimit(request, rule) {
			return false
		}
	}
	return true
}

// detectThreats performs threat detection
func (ztg *ZeroTrustGateway) detectThreats(request *SecurityRequest, policy *SecurityPolicy) *ThreatInfo {
	for _, rule := range policy.ThreatDetectionRules {
		if !rule.Enabled {
			continue
		}

		if threat := ztg.threatDetector.DetectThreat(request, rule); threat != nil {
			return threat
		}
	}
	return nil
}

// authenticate performs authentication
func (ztg *ZeroTrustGateway) authenticate(ctx context.Context, request *SecurityRequest, policy *SecurityPolicy) (*AuthenticationResult, error) {
	// Extract token from Authorization header
	authHeader := request.Headers["Authorization"]
	if authHeader == "" {
		return &AuthenticationResult{
			Success: false,
			Error:   "Missing Authorization header",
		}, nil
	}

	// Extract JWT token
	tokenParts := strings.SplitN(authHeader, " ", 2)
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		return &AuthenticationResult{
			Success: false,
			Error:   "Invalid Authorization header format",
		}, nil
	}

	token := tokenParts[1]

	// Validate token using appropriate provider
	for _, method := range policy.AllowedAuthMethods {
		if provider, exists := ztg.authProviders[method]; exists {
			result, err := provider.ValidateToken(ctx, token)
			if err == nil && result.Valid {
				return &AuthenticationResult{
					Success:     true,
					UserID:      result.UserID,
					Username:    result.Username,
					Roles:       result.Roles,
					Permissions: result.Permissions,
					ExpiresAt:   result.ExpiresAt,
					Metadata:    result.Metadata,
				}, nil
			}
		}
	}

	return &AuthenticationResult{
		Success: false,
		Error:   "Token validation failed",
	}, nil
}

// authorize performs authorization
func (ztg *ZeroTrustGateway) authorize(request *SecurityRequest, policy *SecurityPolicy, authInfo *SecurityResponse) *AuthorizationResult {
	return ztg.authzEngine.Authorize(request, policy, authInfo)
}

// logSecurityEvent logs a security event
func (ztg *ZeroTrustGateway) logSecurityEvent(eventType, severity string, request *SecurityRequest) {
	if !ztg.config.EnableSecurityAudit {
		return
	}

	event := SecurityEvent{
		ID:        generateEventID(),
		Type:      eventType,
		Severity:  severity,
		ServiceID: request.ServiceID,
		ClientIP:  request.ClientIP,
		UserAgent: request.UserAgent,
		Action:    request.Method,
		Resource:  request.Path,
		Timestamp: time.Now(),
		Details: map[string]interface{}{
			"headers": request.Headers,
			"body":    request.Body,
		},
	}

	ztg.auditLogger.LogEvent(event)
}

// === Authentication Provider Implementations ===

// JWTAuthenticationProvider implements JWT-based authentication
type JWTAuthenticationProvider struct {
	secret    []byte
	expiresIn time.Duration
}

func (j *JWTAuthenticationProvider) Name() string { return "jwt" }

func (j *JWTAuthenticationProvider) Authenticate(ctx context.Context, request *AuthenticationRequest) (*AuthenticationResult, error) {
	// This would typically validate username/password and return a JWT
	// For now, return a placeholder implementation
	return &AuthenticationResult{
		Success: false,
		Error:   "Not implemented",
	}, nil
}

func (j *JWTAuthenticationProvider) ValidateToken(ctx context.Context, tokenString string) (*TokenValidationResult, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secret, nil
	})

	if err != nil {
		return &TokenValidationResult{
			Valid: false,
			Error: err.Error(),
		}, nil
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		result := &TokenValidationResult{
			Valid:    true,
			UserID:   getStringClaim(claims, "user_id"),
			Username: getStringClaim(claims, "username"),
			Metadata: make(map[string]interface{}),
		}

		// Extract expiration
		if exp, ok := claims["exp"].(float64); ok {
			result.ExpiresAt = time.Unix(int64(exp), 0)
		}

		// Extract roles
		if roles, ok := claims["roles"].([]interface{}); ok {
			for _, role := range roles {
				if roleStr, ok := role.(string); ok {
					result.Roles = append(result.Roles, roleStr)
				}
			}
		}

		// Extract permissions
		if permissions, ok := claims["permissions"].([]interface{}); ok {
			for _, perm := range permissions {
				if permStr, ok := perm.(string); ok {
					result.Permissions = append(result.Permissions, permStr)
				}
			}
		}

		return result, nil
	}

	return &TokenValidationResult{
		Valid: false,
		Error: "Invalid token",
	}, nil
}

func (j *JWTAuthenticationProvider) RefreshToken(ctx context.Context, refreshToken string) (*TokenRefreshResult, error) {
	// Implement token refresh logic
	return &TokenRefreshResult{
		Success: false,
		Error:   "Not implemented",
	}, nil
}

// === Authorization Engine Methods ===

// Authorize performs authorization check
func (ae *AuthorizationEngine) Authorize(request *SecurityRequest, policy *SecurityPolicy, authInfo *SecurityResponse) *AuthorizationResult {
	ae.mu.RLock()
	defer ae.mu.RUnlock()

	// Sort rules by priority
	rules := make([]*AuthorizationRule, len(policy.AuthorizationRules))
	copy(rules, policy.AuthorizationRules)
	
	// Simple bubble sort by priority
	for i := 0; i < len(rules); i++ {
		for j := i + 1; j < len(rules); j++ {
			if rules[i].Priority > rules[j].Priority {
				rules[i], rules[j] = rules[j], rules[i]
			}
		}
	}

	// Evaluate rules
	for _, rule := range rules {
		if ae.evaluateRule(rule, request, authInfo) {
			switch rule.Action {
			case "allow":
				return &AuthorizationResult{
					Allowed:     true,
					MatchedRule: rule,
				}
			case "deny":
				return &AuthorizationResult{
					Allowed:     false,
					Reason:      fmt.Sprintf("Access denied by rule: %s", rule.Name),
					MatchedRule: rule,
				}
			case "challenge":
				return &AuthorizationResult{
					Allowed:     false,
					Reason:      "Additional authentication required",
					MatchedRule: rule,
				}
			}
		}
	}

	// Default deny
	return &AuthorizationResult{
		Allowed: false,
		Reason:  "No matching authorization rule",
	}
}

// evaluateRule evaluates an authorization rule
func (ae *AuthorizationEngine) evaluateRule(rule *AuthorizationRule, request *SecurityRequest, authInfo *SecurityResponse) bool {
	// Check condition
	switch rule.Condition {
	case "authenticated":
		if authInfo.UserID == "" {
			return false
		}
	case "anonymous":
		if authInfo.UserID != "" {
			return false
		}
	case "always":
		// Always matches
	default:
		// Unknown condition
		return false
	}

	// Check resources
	if len(rule.Resources) > 0 {
		resourceMatch := false
		for _, resource := range rule.Resources {
			if matchesPattern(request.Path, resource) {
				resourceMatch = true
				break
			}
		}
		if !resourceMatch {
			return false
		}
	}

	// Check methods
	if len(rule.Methods) > 0 {
		methodMatch := false
		for _, method := range rule.Methods {
			if strings.EqualFold(request.Method, method) {
				methodMatch = true
				break
			}
		}
		if !methodMatch {
			return false
		}
	}

	// Check roles
	if len(rule.Roles) > 0 {
		roleMatch := false
		for _, requiredRole := range rule.Roles {
			for _, userRole := range authInfo.Roles {
				if strings.EqualFold(requiredRole, userRole) {
					roleMatch = true
					break
				}
			}
			if roleMatch {
				break
			}
		}
		if !roleMatch {
			return false
		}
	}

	// Check permissions
	if len(rule.Permissions) > 0 {
		permMatch := false
		for _, requiredPerm := range rule.Permissions {
			for _, userPerm := range authInfo.Permissions {
				if strings.EqualFold(requiredPerm, userPerm) {
					permMatch = true
					break
				}
			}
			if permMatch {
				break
			}
		}
		if !permMatch {
			return false
		}
	}

	return true
}

// === Threat Detector Methods ===

// DetectThreat detects threats based on rules
func (td *ThreatDetector) DetectThreat(request *SecurityRequest, rule *ThreatDetectionRule) *ThreatInfo {
	switch rule.Type {
	case "sql_injection":
		return td.detectSQLInjection(request, rule)
	case "xss":
		return td.detectXSS(request, rule)
	case "command_injection":
		return td.detectCommandInjection(request, rule)
	case "path_traversal":
		return td.detectPathTraversal(request, rule)
	case "anomaly":
		return td.detectAnomaly(request, rule)
	case "brute_force":
		return td.detectBruteForce(request, rule)
	}
	return nil
}

// detectSQLInjection detects SQL injection attempts
func (td *ThreatDetector) detectSQLInjection(request *SecurityRequest, rule *ThreatDetectionRule) *ThreatInfo {
	content := strings.ToLower(request.Path + " " + request.Body)
	
	matchedPatterns := []string{}
	for _, pattern := range td.patternMatcher.sqlInjectionPatterns {
		if matches, _ := matchPattern(content, pattern); matches {
			matchedPatterns = append(matchedPatterns, pattern)
		}
	}

	if len(matchedPatterns) > 0 {
		confidence := float64(len(matchedPatterns)) / float64(len(td.patternMatcher.sqlInjectionPatterns))
		if confidence >= rule.Threshold {
			return &ThreatInfo{
				Type:        "sql_injection",
				Severity:    rule.Severity,
				Confidence:  confidence,
				Description: "Potential SQL injection attempt detected",
				Patterns:    matchedPatterns,
			}
		}
	}

	return nil
}

// detectXSS detects XSS attempts
func (td *ThreatDetector) detectXSS(request *SecurityRequest, rule *ThreatDetectionRule) *ThreatInfo {
	content := strings.ToLower(request.Path + " " + request.Body)
	
	matchedPatterns := []string{}
	for _, pattern := range td.patternMatcher.xssPatterns {
		if matches, _ := matchPattern(content, pattern); matches {
			matchedPatterns = append(matchedPatterns, pattern)
		}
	}

	if len(matchedPatterns) > 0 {
		confidence := float64(len(matchedPatterns)) / float64(len(td.patternMatcher.xssPatterns))
		if confidence >= rule.Threshold {
			return &ThreatInfo{
				Type:        "xss",
				Severity:    rule.Severity,
				Confidence:  confidence,
				Description: "Potential XSS attempt detected",
				Patterns:    matchedPatterns,
			}
		}
	}

	return nil
}

// detectCommandInjection detects command injection attempts
func (td *ThreatDetector) detectCommandInjection(request *SecurityRequest, rule *ThreatDetectionRule) *ThreatInfo {
	content := request.Path + " " + request.Body
	
	matchedPatterns := []string{}
	for _, pattern := range td.patternMatcher.commandInjectionPatterns {
		if matches, _ := matchPattern(content, pattern); matches {
			matchedPatterns = append(matchedPatterns, pattern)
		}
	}

	if len(matchedPatterns) > 0 {
		confidence := float64(len(matchedPatterns)) / float64(len(td.patternMatcher.commandInjectionPatterns))
		if confidence >= rule.Threshold {
			return &ThreatInfo{
				Type:        "command_injection",
				Severity:    rule.Severity,
				Confidence:  confidence,
				Description: "Potential command injection attempt detected",
				Patterns:    matchedPatterns,
			}
		}
	}

	return nil
}

// detectPathTraversal detects path traversal attempts
func (td *ThreatDetector) detectPathTraversal(request *SecurityRequest, rule *ThreatDetectionRule) *ThreatInfo {
	content := request.Path
	
	matchedPatterns := []string{}
	for _, pattern := range td.patternMatcher.pathTraversalPatterns {
		if matches, _ := matchPattern(content, pattern); matches {
			matchedPatterns = append(matchedPatterns, pattern)
		}
	}

	if len(matchedPatterns) > 0 {
		return &ThreatInfo{
			Type:        "path_traversal",
			Severity:    rule.Severity,
			Confidence:  1.0,
			Description: "Potential path traversal attempt detected",
			Patterns:    matchedPatterns,
		}
	}

	return nil
}

// detectAnomaly detects behavioral anomalies
func (td *ThreatDetector) detectAnomaly(request *SecurityRequest, rule *ThreatDetectionRule) *ThreatInfo {
	// Use AI behavior analysis
	if userID := request.Headers["X-User-ID"]; userID != "" {
		anomalyScore := td.behaviorAI.CalculateAnomalyScore(userID, request)
		if anomalyScore >= rule.Threshold {
			return &ThreatInfo{
				Type:        "anomaly",
				Severity:    rule.Severity,
				Confidence:  anomalyScore,
				Description: "Behavioral anomaly detected",
				Metadata: map[string]interface{}{
					"anomaly_score": anomalyScore,
				},
			}
		}
	}

	return nil
}

// detectBruteForce detects brute force attempts
func (td *ThreatDetector) detectBruteForce(request *SecurityRequest, rule *ThreatDetectionRule) *ThreatInfo {
	// Simple brute force detection based on failed attempts
	// This would typically track failed login attempts per IP
	
	if strings.Contains(request.Path, "/login") || strings.Contains(request.Path, "/auth") {
		// In a real implementation, this would check a cache of recent failed attempts
		// For now, return nil (no brute force detected)
	}

	return nil
}

// === Rate Limiter Methods ===

// CheckLimit checks if a request is within rate limits
func (rl *RateLimiter) CheckLimit(request *SecurityRequest, rule *RateLimitRule) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	key := rl.generateKey(request, rule)
	counter, exists := rl.counters[key]
	
	now := time.Now()

	if !exists || now.After(counter.ResetTime) {
		// Initialize or reset counter
		counter = &RateLimitCounter{
			Limit:      rule.Limit,
			Remaining:  rule.Limit - 1,
			ResetTime:  now.Add(rule.Window),
			Window:     rule.Window,
			LastUpdate: now,
		}
		rl.counters[key] = counter
		return true
	}

	if counter.Remaining <= 0 {
		// Rate limit exceeded
		return false
	}

	// Decrement counter
	counter.Remaining--
	counter.LastUpdate = now
	
	return true
}

// generateKey generates a rate limit key based on scope
func (rl *RateLimiter) generateKey(request *SecurityRequest, rule *RateLimitRule) string {
	switch rule.Scope {
	case "global":
		return "global"
	case "ip":
		return "ip:" + request.ClientIP
	case "user":
		if userID := request.Headers["X-User-ID"]; userID != "" {
			return "user:" + userID
		}
		return "ip:" + request.ClientIP
	case "service":
		return "service:" + request.ServiceID
	default:
		return "default"
	}
}

// === Behavior Analysis AI Methods ===

// UpdateBehavior updates user behavior patterns
func (bai *BehaviorAnalysisAI) UpdateBehavior(userID string, request *SecurityRequest) {
	bai.mu.Lock()
	defer bai.mu.Unlock()

	profile, exists := bai.userProfiles[userID]
	if !exists {
		profile = &UserBehaviorProfile{
			UserID:            userID,
			TypicalIPs:        make(map[string]float64),
			TypicalUserAgents: make(map[string]float64),
			TypicalResources:  make(map[string]float64),
			RequestPatterns:   make([]RequestPattern, 0),
			LastUpdated:       time.Now(),
		}
		bai.userProfiles[userID] = profile
	}

	// Update hourly patterns
	hour := request.Timestamp.Hour()
	profile.TypicalHours[hour] += bai.learningRate

	// Update daily patterns
	weekday := int(request.Timestamp.Weekday())
	profile.TypicalDays[weekday] += bai.learningRate

	// Update IP patterns
	if _, exists := profile.TypicalIPs[request.ClientIP]; !exists {
		profile.TypicalIPs[request.ClientIP] = 0
	}
	profile.TypicalIPs[request.ClientIP] += bai.learningRate

	// Update user agent patterns
	if _, exists := profile.TypicalUserAgents[request.UserAgent]; !exists {
		profile.TypicalUserAgents[request.UserAgent] = 0
	}
	profile.TypicalUserAgents[request.UserAgent] += bai.learningRate

	// Update resource patterns
	if _, exists := profile.TypicalResources[request.Path]; !exists {
		profile.TypicalResources[request.Path] = 0
	}
	profile.TypicalResources[request.Path] += bai.learningRate

	profile.LastUpdated = time.Now()
}

// CalculateAnomalyScore calculates anomaly score for a request
func (bai *BehaviorAnalysisAI) CalculateAnomalyScore(userID string, request *SecurityRequest) float64 {
	bai.mu.RLock()
	defer bai.mu.RUnlock()

	profile, exists := bai.userProfiles[userID]
	if !exists {
		return 0.5 // Neutral score for unknown users
	}

	score := 0.0

	// Check time-based anomalies
	hour := request.Timestamp.Hour()
	weekday := int(request.Timestamp.Weekday())
	
	hourlyScore := 1.0 - profile.TypicalHours[hour]
	dailyScore := 1.0 - profile.TypicalDays[weekday]
	
	score += 0.2 * hourlyScore
	score += 0.2 * dailyScore

	// Check IP anomalies
	ipScore := 1.0
	if freq, exists := profile.TypicalIPs[request.ClientIP]; exists {
		ipScore = 1.0 - freq
	}
	score += 0.3 * ipScore

	// Check user agent anomalies
	uaScore := 1.0
	if freq, exists := profile.TypicalUserAgents[request.UserAgent]; exists {
		uaScore = 1.0 - freq
	}
	score += 0.2 * uaScore

	// Check resource anomalies
	resourceScore := 1.0
	if freq, exists := profile.TypicalResources[request.Path]; exists {
		resourceScore = 1.0 - freq
	}
	score += 0.1 * resourceScore

	// Clamp score to [0, 1]
	if score < 0 {
		score = 0
	}
	if score > 1 {
		score = 1
	}

	return score
}

// === IP Reputation Manager Methods ===

// GetReputation gets IP reputation
func (irm *IPReputationManager) GetReputation(ip string) *IPReputation {
	irm.mu.RLock()
	defer irm.mu.RUnlock()

	// Check cache first
	if cached, exists := irm.cache[ip]; exists {
		if time.Since(cached.CachedAt) < irm.cacheTTL {
			return cached.Reputation
		}
	}

	// Check main reputation store
	if reputation, exists := irm.reputation[ip]; exists {
		return reputation
	}

	return nil
}

// InitializeIP initializes reputation tracking for an IP
func (irm *IPReputationManager) InitializeIP(ip string) {
	irm.mu.Lock()
	defer irm.mu.Unlock()

	if _, exists := irm.reputation[ip]; !exists {
		irm.reputation[ip] = &IPReputation{
			IP:           ip,
			Score:        0.8, // Start with good reputation
			Category:     "unknown",
			FirstSeen:    time.Now(),
			LastSeen:     time.Now(),
			RequestCount: 1,
			LastUpdated:  time.Now(),
		}
	}
}

// UpdateReputation updates IP reputation based on behavior
func (irm *IPReputationManager) UpdateReputation(ip string, positive bool) {
	irm.mu.Lock()
	defer irm.mu.Unlock()

	reputation, exists := irm.reputation[ip]
	if !exists {
		irm.InitializeIP(ip)
		reputation = irm.reputation[ip]
	}

	reputation.RequestCount++
	reputation.LastSeen = time.Now()

	if positive {
		// Gradually improve reputation
		reputation.Score += 0.01 * (1.0 - reputation.Score)
	} else {
		// Degrade reputation
		reputation.ViolationCount++
		reputation.Score *= 0.9
	}

	// Clamp score to [0, 1]
	if reputation.Score < 0 {
		reputation.Score = 0
	}
	if reputation.Score > 1 {
		reputation.Score = 1
	}

	reputation.LastUpdated = time.Now()

	// Update cache
	irm.cache[ip] = &CachedReputation{
		Reputation: reputation,
		CachedAt:   time.Now(),
	}
}

// === Security Audit Logger Methods ===

// LogEvent logs a security event
func (sal *SecurityAuditLogger) LogEvent(event SecurityEvent) {
	sal.mu.Lock()
	defer sal.mu.Unlock()

	sal.events = append(sal.events, event)

	// Keep only recent events
	if len(sal.events) > sal.maxEvents {
		sal.events = sal.events[1:]
	}

	// Log to system log as well
	log.Printf("🔒 Security Event: %s - %s (Service: %s, IP: %s)", 
		event.Severity, event.Type, event.ServiceID, event.ClientIP)
}

// GetEvents returns recent security events
func (sal *SecurityAuditLogger) GetEvents(limit int) []SecurityEvent {
	sal.mu.RLock()
	defer sal.mu.RUnlock()

	if limit <= 0 || limit > len(sal.events) {
		limit = len(sal.events)
	}

	result := make([]SecurityEvent, limit)
	startIndex := len(sal.events) - limit
	
	for i := 0; i < limit; i++ {
		result[i] = sal.events[startIndex+i]
	}

	return result
}

// === Helper Functions ===

// generateEventID generates a unique event ID
func generateEventID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// getStringClaim safely gets a string claim from JWT
func getStringClaim(claims jwt.MapClaims, key string) string {
	if value, ok := claims[key].(string); ok {
		return value
	}
	return ""
}

// matchesPattern checks if a string matches a pattern (simplified)
func matchesPattern(text, pattern string) bool {
	// Simple wildcard matching
	if pattern == "*" {
		return true
	}
	
	// Exact match for now
	return strings.Contains(text, pattern)
}

// matchPattern matches a regex pattern (simplified)
func matchPattern(text, pattern string) (bool, error) {
	// Simple string contains for now
	// In a real implementation, this would use regex
	return strings.Contains(text, strings.ToLower(pattern)), nil
}

// isPrivateIP checks if an IP is private
func isPrivateIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	// Check for private IP ranges
	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
	}

	for _, cidr := range privateRanges {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if network.Contains(ip) {
			return true
		}
	}

	return false
}

// hashString creates a SHA256 hash of a string
func hashString(input string) string {
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}