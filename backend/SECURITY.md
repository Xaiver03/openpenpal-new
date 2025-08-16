# OpenPenPal Security Implementation Guide

## Overview

This document outlines the comprehensive security measures implemented in the OpenPenPal backend system, including Content Security Policy (CSP), security headers, threat detection, and input validation.

## Security Features

### 1. Content Security Policy (CSP)

#### Dynamic CSP Configuration
- **Development Environment**: Relaxed policies to support hot reload and development tools
- **Production Environment**: Strict policies with nonce-based script execution
- **CSP Reporting**: Comprehensive violation reporting and monitoring

#### CSP Directives Implemented
```
default-src 'self'
script-src 'self' 'nonce-{random}' https://cdn.jsdelivr.net
style-src 'self' 'nonce-{random}' https://fonts.googleapis.com
font-src 'self' https://fonts.gstatic.com data:
img-src 'self' data: https:
connect-src 'self' wss://your-domain.com
media-src 'self'
object-src 'none'
frame-ancestors 'none'
base-uri 'self'
form-action 'self'
upgrade-insecure-requests
block-all-mixed-content
require-trusted-types-for 'script'
```

### 2. Security Headers

#### Comprehensive Header Protection
- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `X-XSS-Protection: 0` (disabled in favor of CSP)
- `Referrer-Policy: strict-origin-when-cross-origin`
- `Strict-Transport-Security: max-age=63072000; includeSubDomains; preload`
- `Cross-Origin-Embedder-Policy: require-corp`
- `Cross-Origin-Opener-Policy: same-origin`
- `Cross-Origin-Resource-Policy: same-origin`

#### Modern Permissions Policy
Restricts access to sensitive browser APIs:
```
geolocation=()
microphone=()
camera=()
usb=()
bluetooth=()
payment=()
interest-cohort=()  // Blocks Google FLoC
browsing-topics=()  // Blocks Topics API
```

### 3. Threat Detection

#### Real-time Attack Pattern Recognition
- **SQL Injection Detection**: Monitors for common SQL injection patterns
- **XSS Attempt Detection**: Identifies cross-site scripting attempts
- **Directory Traversal Detection**: Prevents path traversal attacks
- **Suspicious User Agent Detection**: Flags potentially malicious bots

#### Threat Response
- Automatic logging of security events
- IP-based threat flagging
- Security header warnings
- Configurable response actions

### 4. Input Validation and Sanitization

#### Multi-Layer Validation
- **Request Header Validation**: User-Agent, Content-Type verification
- **Query Parameter Validation**: Length limits, character restrictions
- **Path Parameter Validation**: UUID format verification, injection prevention
- **Request Body Validation**: JSON schema validation, content sanitization

#### Content Security Service Integration
- **Sensitive Word Filtering**: Real-time content moderation
- **XSS Prevention**: HTML sanitization using bluemonday
- **Content Scoring**: AI-powered risk assessment
- **Automated Actions**: Block, review, or approve based on risk level

### 5. Rate Limiting

#### Multi-Tier Rate Limiting
- **IP-based Limiting**: Global request limits per IP address
- **User-based Limiting**: Authenticated user request limits
- **Operation-specific Limiting**: Targeted limits for sensitive operations
- **Adaptive Limiting**: Dynamic adjustment based on threat level

#### Rate Limit Configuration
```go
// General API access: 10 requests/second, burst 100
generalLimiter = NewIPRateLimiter(rate.Every(100*time.Millisecond), 100)

// Authentication: 6 requests/minute, burst 20
authLimiter = NewIPRateLimiter(rate.Every(10*time.Second), 20)

// Sensitive operations: Custom limits per operation
sensitiveWordCreate = rate.NewLimiter(rate.Every(time.Second), 10)
batchOperations = rate.NewLimiter(rate.Every(time.Minute), 5)
```

### 6. Secure Session Management

#### JWT Security
- **Secure Token Generation**: Cryptographically secure random tokens
- **Token Rotation**: Automatic refresh token mechanism
- **Expiry Management**: Short-lived access tokens with longer refresh tokens
- **CSRF Protection**: Additional CSRF tokens for state-changing operations

#### Session Configuration
- `HttpOnly` cookies to prevent XSS access
- `Secure` flag for HTTPS-only transmission
- `SameSite=Strict` for CSRF protection
- Configurable session timeouts

### 7. Sensitive Data Protection

#### Data Classification
- **Public Data**: Publicly accessible information
- **Personal Data**: User-specific information requiring authentication
- **Sensitive Data**: Administrative data requiring elevated permissions
- **Secret Data**: Passwords, tokens, and cryptographic keys

#### Protection Measures
- **Encryption at Rest**: Database encryption for sensitive fields
- **Encryption in Transit**: TLS 1.3 for all communications
- **Key Management**: Secure key storage and rotation
- **Access Logging**: Comprehensive audit trails

### 8. Administrative Security

#### Role-Based Access Control (RBAC)
```
Super Admin > Platform Admin > L4 Courier > L3 Courier > L2 Courier > L1 Courier > User
```

#### Sensitive Word Management
- **L4 Courier and Platform Admin**: Full management capabilities
- **Input Validation**: Comprehensive word validation and sanitization
- **Rate Limiting**: Strict limits on sensitive word operations
- **Audit Logging**: Complete audit trail for all changes

## Security Configuration

### Environment Variables
See `.env.security.example` for comprehensive configuration options.

### Critical Security Settings
```bash
# Production Security (Required)
ENVIRONMENT=production
ENABLE_HSTS=true
ENABLE_CSP_REPORTING=true
SECURITY_MONITORING_ENABLED=true
THREAT_DETECTION_ENABLED=true

# Content Security
CONTENT_SECURITY_ENABLED=true
XSS_PROTECTION_LEVEL=strict
SENSITIVE_WORD_AUTO_UPDATE=true

# Rate Limiting
RATE_LIMIT_ENABLED=true
RATE_LIMIT_GENERAL=100
RATE_LIMIT_AUTH=10
```

## Security Monitoring

### Log Types
1. **Security Events**: Authentication, authorization, access control
2. **Threat Detection**: Attack attempts, suspicious behavior
3. **CSP Violations**: Content Security Policy violations
4. **Rate Limit Events**: Rate limiting triggers and blocks
5. **Admin Actions**: Sensitive administrative operations

### Monitoring Integration
- Structured logging with security event classification
- Real-time threat detection alerts
- Performance impact monitoring
- Compliance audit trail generation

## Security Testing

### Automated Security Tests
- CSP policy validation
- Security header verification
- Input validation testing
- Rate limiting functionality
- Authentication flow security

### Manual Security Verification
1. CSP violation reporting functionality
2. Threat detection accuracy
3. Rate limiting effectiveness
4. Input sanitization completeness
5. Session security implementation

## Security Incident Response

### Incident Classification
- **Low**: Minor policy violations, rate limit triggers
- **Medium**: Suspicious patterns, failed authentication attempts
- **High**: Active attack patterns, data access attempts
- **Critical**: Successful breaches, system compromise

### Response Procedures
1. **Immediate**: Automatic blocking and logging
2. **Short-term**: Security team notification and investigation
3. **Medium-term**: Incident analysis and mitigation
4. **Long-term**: Security policy updates and improvements

## Compliance and Standards

### Security Standards Compliance
- **OWASP Top 10**: Protection against all major web application risks
- **NIST Cybersecurity Framework**: Implementation of security controls
- **ISO 27001**: Information security management practices
- **GDPR**: Privacy and data protection compliance

### Regular Security Updates
- Dependency vulnerability scanning
- Security patch management
- Security configuration reviews
- Penetration testing and security audits

## Security Best Practices

### For Developers
1. Always validate and sanitize user input
2. Use parameterized queries to prevent SQL injection
3. Implement proper error handling without information disclosure
4. Follow the principle of least privilege
5. Keep dependencies updated and scan for vulnerabilities

### For Administrators
1. Regularly review security logs and events
2. Monitor for unusual patterns or activities
3. Keep security configurations up to date
4. Implement proper backup and recovery procedures
5. Conduct regular security assessments

### For Operations
1. Monitor system performance and security metrics
2. Implement proper incident response procedures
3. Maintain security documentation and procedures
4. Conduct regular security training and awareness
5. Coordinate with security teams on threat intelligence

## Support and Contacts

For security-related questions or incident reporting:
- Security Team: security@your-domain.com
- Emergency Contact: +1-XXX-XXX-XXXX
- Documentation: https://docs.your-domain.com/security

---

**Last Updated**: 2024-08-15
**Document Version**: 1.0
**Review Schedule**: Quarterly