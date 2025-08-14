# Security Enhancement Migration Guide

## Overview

This guide explains how to migrate from the current security implementation to the enhanced security system that addresses critical vulnerabilities identified in the security audit.

## Security Enhancements Implemented

### 1. JWT Secret Management (`internal/security/secrets/jwt_manager.go`)
- **Problem**: JWT secret hardcoded in environment variables
- **Solution**: Cryptographically secure JWT secret generation with rotation support
- **Features**:
  - Automatic secret generation
  - 90-day rotation cycle
  - Version-based token validation
  - Backward compatibility

### 2. Enhanced CSRF Protection (`internal/security/csrf/enhanced_csrf.go`)
- **Problem**: Login endpoint bypassed CSRF, weak implementation
- **Solution**: Double-submit cookie pattern with strict validation
- **Features**:
  - Removed login bypass vulnerability
  - Secure cookie settings
  - Constant-time comparison
  - Automatic token cleanup

### 3. Adaptive Rate Limiting (`internal/security/ratelimit/adaptive_limiter.go`)
- **Problem**: Fixed rate limits too permissive
- **Solution**: Behavior-based adaptive limiting
- **Features**:
  - Dynamic rate adjustment based on behavior
  - Automatic blocking for suspicious activity
  - Different limits for auth/API/upload endpoints
  - IP and user-based tracking

### 4. Input Validation (`internal/security/validation/input_validator.go`)
- **Problem**: Limited input sanitization, SQL injection risks
- **Solution**: Comprehensive validation middleware
- **Features**:
  - Field-specific validation rules
  - SQL injection pattern detection
  - XSS prevention
  - Length and format validation

### 5. Secure Configuration (`internal/security/env/secure_config.go`)
- **Problem**: Sensitive data in plaintext environment variables
- **Solution**: Encrypted configuration management
- **Features**:
  - AES-GCM encryption for sensitive values
  - Automatic key generation
  - Development vs production modes
  - Validation of required configs

### 6. Enhanced Security Headers (`internal/middleware/security_headers_enhanced.go`)
- **Problem**: Missing or weak security headers
- **Solution**: Comprehensive security headers with CSP
- **Features**:
  - Content Security Policy with nonces
  - HSTS in production
  - Permissions Policy
  - XSS/Clickjacking protection

## Migration Steps

### Step 1: Install Dependencies

```bash
cd backend
go get github.com/golang-jwt/jwt/v5
go get golang.org/x/time/rate
go get github.com/go-playground/validator/v10
```

### Step 2: Copy Security Modules

Copy all files from the security implementation:
- `internal/security/` directory
- `internal/middleware/auth_enhanced.go`
- `internal/middleware/rate_limiter_enhanced.go`
- `internal/middleware/security_headers_enhanced.go`

### Step 3: Update Configuration

1. Add to `.env` (development only):
```env
CONFIG_ENCRYPTION_KEY=<base64-encoded-32-byte-key>
```

2. Remove hardcoded JWT_SECRET from .env (will be auto-generated)

### Step 4: Update main.go

Replace the security initialization in main.go:

```go
// Old code
cfg, err := config.Load()
// ... database setup ...

// New code
cfg, err := config.Load()
securityConfig, err := security.InitializeSecurity(cfg)
if err != nil {
    log.Fatal("Failed to initialize security:", err)
}
// ... database setup ...
```

### Step 5: Update Middleware

Replace middleware usage:

```go
// Old
router.Use(middleware.CORS())
router.Use(middleware.SecurityHeaders())
api.Use(middleware.AuthMiddleware(cfg, db))
api.Use(middleware.RateLimiter())

// New
security.ApplySecurityMiddleware(router, db, cfg, securityConfig)
api.Use(middleware.EnhancedAuthMiddleware(cfg, db, securityConfig.JWTManager))
api.Use(securityConfig.CSRFProtection.Middleware())
api.Use(securityConfig.RateLimiter.APILimiter())
```

### Step 6: Update Auth Handlers

Update authentication handlers to use the JWT manager:

```go
// In auth handler
token, err := s.jwtManager.GenerateToken(claims)
```

### Step 7: Add Security Monitoring Endpoints

Add admin endpoints for security monitoring:

```go
admin.GET("/security/metrics", getSecurityMetrics(securityConfig))
admin.GET("/security/blocked", getBlockedIdentifiers(securityConfig))
```

## Configuration

### Development Environment

```yaml
# Permissive settings for development
RateLimit:
  General: 10 req/sec
  Auth: 1 req/sec
  
CSRF:
  SameSite: Lax
  Secure: false
  
Headers:
  CSP: Allows unsafe-eval for hot reload
```

### Production Environment

```yaml
# Strict settings for production
RateLimit:
  General: 1 req/sec
  Auth: 0.2 req/sec (1 per 5 seconds)
  
CSRF:
  SameSite: Strict
  Secure: true
  
Headers:
  CSP: Strict with nonces
  HSTS: Enabled with preload
```

## Testing

### 1. Test JWT Rotation

```bash
# Get initial token
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","password":"secret123"}'

# Token should work with new JWT manager
curl -X GET http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer <token>"
```

### 2. Test CSRF Protection

```bash
# Get CSRF token
curl -X GET http://localhost:8080/api/v1/auth/csrf

# Login should now require CSRF token
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "X-CSRF-Token: <token>" \
  -d '{"username":"alice","password":"secret123"}'
```

### 3. Test Rate Limiting

```bash
# Rapid requests should be blocked
for i in {1..10}; do
  curl -X POST http://localhost:8080/api/v1/auth/login \
    -d '{"username":"wrong","password":"wrong"}'
done
# Should see 429 Too Many Requests after threshold
```

### 4. Test Input Validation

```bash
# SQL injection attempt should be blocked
curl -X POST http://localhost:8080/api/v1/auth/register \
  -d '{"username":"admin'; DROP TABLE users;--","password":"test"}'
# Should see validation error
```

## Rollback Plan

If issues occur:

1. Keep old main.go as main_backup.go
2. Revert to original middleware
3. Set JWT_SECRET in .env manually
4. Remove security package imports

## Monitoring

Monitor these metrics after deployment:

1. **Failed authentication attempts** - Watch for spikes
2. **Rate limit hits** - Adjust limits if too restrictive
3. **CSRF failures** - May indicate integration issues
4. **CSP violations** - Check browser console for reports
5. **Blocked IPs** - Monitor for false positives

## Security Checklist

- [ ] Remove all hardcoded secrets
- [ ] Enable HTTPS in production
- [ ] Configure proper CORS origins
- [ ] Set up log monitoring
- [ ] Test all auth flows
- [ ] Verify rate limits work
- [ ] Check CSP doesn't break UI
- [ ] Document emergency procedures

## Support

For issues or questions about the security migration:
1. Check error logs for detailed messages
2. Review security metrics endpoint
3. Test with reduced rate limits first
4. Enable debug logging for troubleshooting