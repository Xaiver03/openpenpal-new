#!/bin/bash

# OpenPenPal Production Security Validation Script
# 生产环境安全验证脚本
#
# Usage: ./validate-production-security.sh <domain>
# Example: ./validate-production-security.sh https://openpenpal.com

set -e

DOMAIN=${1:-"http://localhost:3000"}
PASSED=0
FAILED=0
WARNINGS=0

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔒 OpenPenPal Production Security Validation${NC}"
echo -e "${BLUE}Domain: ${DOMAIN}${NC}"
echo "=============================================="

# Helper functions
pass() {
    echo -e "✅ ${GREEN}PASS${NC}: $1"
    ((PASSED++))
}

fail() {
    echo -e "❌ ${RED}FAIL${NC}: $1"
    echo -e "   ${RED}$2${NC}"
    ((FAILED++))
}

warn() {
    echo -e "⚠️  ${YELLOW}WARN${NC}: $1"
    echo -e "   ${YELLOW}$2${NC}"
    ((WARNINGS++))
}

info() {
    echo -e "ℹ️  ${BLUE}INFO${NC}: $1"
}

# Test function with timeout
test_endpoint() {
    local url="$1"
    local method="${2:-GET}"
    local data="$3"
    local timeout="${4:-10}"
    
    if [[ -n "$data" ]]; then
        curl -s -m "$timeout" -X "$method" -H "Content-Type: application/json" -d "$data" "$url" 2>/dev/null
    else
        curl -s -m "$timeout" -X "$method" "$url" 2>/dev/null
    fi
}

# Test headers with timeout
test_headers() {
    local url="$1"
    local timeout="${2:-10}"
    curl -s -I -m "$timeout" "$url" 2>/dev/null
}

echo -e "\n${PURPLE}🌐 Basic Connectivity Tests${NC}"
echo "================================"

# 1. Basic connectivity
info "Testing basic connectivity..."
if test_endpoint "$DOMAIN" > /dev/null; then
    pass "Domain is accessible"
else
    fail "Domain is not accessible" "Check DNS, SSL certificates, and server status"
    exit 1
fi

# 2. HTTPS enforcement (if domain uses https)
if [[ "$DOMAIN" == https* ]]; then
    info "Testing HTTPS enforcement..."
    HTTP_DOMAIN=$(echo "$DOMAIN" | sed 's/https/http/')
    HTTP_RESPONSE=$(curl -s -I -m 10 "$HTTP_DOMAIN" 2>/dev/null | head -1 || echo "")
    
    if echo "$HTTP_RESPONSE" | grep -q "301\|302"; then
        pass "HTTPS redirect is working"
    else
        warn "HTTPS redirect not detected" "HTTP requests should redirect to HTTPS"
    fi
fi

echo -e "\n${PURPLE}🛡️  Security Headers Tests${NC}"
echo "================================="

# 3. Security headers
info "Testing security headers..."
HEADERS=$(test_headers "$DOMAIN")

# X-Frame-Options
if echo "$HEADERS" | grep -qi "x-frame-options"; then
    FRAME_VALUE=$(echo "$HEADERS" | grep -i "x-frame-options" | cut -d: -f2 | tr -d ' \r\n')
    pass "X-Frame-Options: $FRAME_VALUE"
else
    fail "X-Frame-Options header missing" "This header prevents clickjacking attacks"
fi

# X-Content-Type-Options
if echo "$HEADERS" | grep -qi "x-content-type-options.*nosniff"; then
    pass "X-Content-Type-Options: nosniff"
else
    fail "X-Content-Type-Options: nosniff header missing" "This prevents MIME type sniffing attacks"
fi

# Content-Security-Policy
if echo "$HEADERS" | grep -qi "content-security-policy"; then
    pass "Content-Security-Policy header present"
    CSP_HEADER=$(echo "$HEADERS" | grep -i "content-security-policy" | cut -d: -f2-)
    if echo "$CSP_HEADER" | grep -q "default-src.*'self'"; then
        pass "CSP includes 'self' directive"
    else
        warn "CSP may be too permissive" "Review CSP policy for security"
    fi
else
    fail "Content-Security-Policy header missing" "CSP helps prevent XSS attacks"
fi

# Strict-Transport-Security (for HTTPS)
if [[ "$DOMAIN" == https* ]]; then
    if echo "$HEADERS" | grep -qi "strict-transport-security"; then
        HSTS_VALUE=$(echo "$HEADERS" | grep -i "strict-transport-security" | cut -d: -f2 | tr -d ' \r\n')
        pass "Strict-Transport-Security: $HSTS_VALUE"
        
        if echo "$HSTS_VALUE" | grep -q "preload"; then
            pass "HSTS preload directive present"
        else
            warn "HSTS preload directive missing" "Consider adding preload for stronger security"
        fi
    else
        fail "Strict-Transport-Security header missing" "HSTS prevents SSL stripping attacks"
    fi
fi

# X-XSS-Protection
if echo "$HEADERS" | grep -qi "x-xss-protection"; then
    XSS_VALUE=$(echo "$HEADERS" | grep -i "x-xss-protection" | cut -d: -f2 | tr -d ' \r\n')
    pass "X-XSS-Protection: $XSS_VALUE"
else
    warn "X-XSS-Protection header missing" "While deprecated, still provides legacy browser protection"
fi

echo -e "\n${PURPLE}🔐 Authentication & CSRF Tests${NC}"
echo "======================================="

# 4. CSRF token generation
info "Testing CSRF token generation..."
CSRF_RESPONSE=$(test_endpoint "$DOMAIN/api/auth/csrf")
if echo "$CSRF_RESPONSE" | grep -q "csrfToken\|token"; then
    pass "CSRF token endpoint is working"
    
    # Extract token for further testing
    CSRF_TOKEN=$(echo "$CSRF_RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4 2>/dev/null || echo "")
    if [[ -n "$CSRF_TOKEN" && ${#CSRF_TOKEN} -ge 32 ]]; then
        pass "CSRF token is sufficiently long (${#CSRF_TOKEN} chars)"
    else
        warn "CSRF token may be too short" "Tokens should be at least 32 characters"
    fi
else
    fail "CSRF token endpoint not working" "Check /api/auth/csrf endpoint"
fi

echo -e "\n${PURPLE}⚡ Rate Limiting Tests${NC}"
echo "=========================="

# 5. Rate limiting
info "Testing rate limiting on auth endpoint..."
RATE_LIMIT_HITS=0
for i in {1..15}; do
    RESPONSE=$(curl -s -w "%{http_code}" -o /dev/null -m 5 \
        -X POST "$DOMAIN/api/auth/login" \
        -H "Content-Type: application/json" \
        -d '{"username":"testlimit","password":"testlimit"}' 2>/dev/null || echo "000")
    
    if [[ "$RESPONSE" == "429" ]]; then
        ((RATE_LIMIT_HITS++))
    fi
    
    # Small delay to avoid overwhelming
    sleep 0.1
done

if [[ $RATE_LIMIT_HITS -gt 0 ]]; then
    pass "Rate limiting is active ($RATE_LIMIT_HITS/15 requests blocked)"
else
    warn "Rate limiting not detected" "Consider implementing rate limiting for production"
fi

echo -e "\n${PURPLE}🔍 API Security Tests${NC}"
echo "======================"

# 6. Authentication endpoint
info "Testing authentication endpoint..."
AUTH_RESPONSE=$(test_endpoint "$DOMAIN/api/auth/login" "POST" '{"username":"nonexistent","password":"test"}')
if echo "$AUTH_RESPONSE" | grep -q "401\|400\|error\|unauthorized"; then
    pass "Authentication endpoint properly rejects invalid credentials"
else
    warn "Authentication endpoint response unclear" "Verify proper error handling"
fi

# 7. Protected routes (if accessible)
info "Testing protected route access..."
PROTECTED_RESPONSE=$(curl -s -w "%{http_code}" -o /dev/null -m 10 "$DOMAIN/api/users/me" 2>/dev/null || echo "000")
if [[ "$PROTECTED_RESPONSE" == "401" || "$PROTECTED_RESPONSE" == "403" ]]; then
    pass "Protected routes properly require authentication"
elif [[ "$PROTECTED_RESPONSE" == "404" ]]; then
    info "Protected route not found (may not be implemented)"
else
    warn "Protected route may be accessible without auth" "Verify authentication middleware"
fi

echo -e "\n${PURPLE}🗂️  Configuration Tests${NC}"
echo "========================="

# 8. Environment detection
info "Testing environment configuration..."
# Try to detect if we're in development mode
DEV_INDICATORS=$(test_endpoint "$DOMAIN" | grep -c "development\|dev-mode\|localhost" || echo "0")
if [[ "$DEV_INDICATORS" -gt 0 ]]; then
    warn "Development indicators detected" "Ensure production mode is enabled"
else
    pass "No development indicators found"
fi

# 9. Error handling
info "Testing error handling..."
ERROR_RESPONSE=$(test_endpoint "$DOMAIN/api/nonexistent-endpoint")
if echo "$ERROR_RESPONSE" | grep -qi "stack\|trace\|debug"; then
    fail "Error responses may leak sensitive information" "Ensure stack traces are disabled in production"
else
    pass "Error responses don't leak sensitive information"
fi

echo -e "\n${PURPLE}📊 SSL/TLS Tests (HTTPS only)${NC}"
echo "============================="

if [[ "$DOMAIN" == https* ]]; then
    # 10. SSL certificate
    info "Testing SSL certificate..."
    SSL_INFO=$(echo | openssl s_client -servername "$(echo $DOMAIN | sed 's|https://||' | sed 's|/.*||')" -connect "$(echo $DOMAIN | sed 's|https://||' | sed 's|/.*||'):443" 2>/dev/null | openssl x509 -noout -dates 2>/dev/null || echo "")
    
    if [[ -n "$SSL_INFO" ]]; then
        pass "SSL certificate is valid"
        
        # Check expiration
        EXPIRY=$(echo "$SSL_INFO" | grep "notAfter" | cut -d= -f2)
        if [[ -n "$EXPIRY" ]]; then
            EXPIRY_TIMESTAMP=$(date -d "$EXPIRY" +%s 2>/dev/null || echo "0")
            CURRENT_TIMESTAMP=$(date +%s)
            DAYS_UNTIL_EXPIRY=$(( (EXPIRY_TIMESTAMP - CURRENT_TIMESTAMP) / 86400 ))
            
            if [[ $DAYS_UNTIL_EXPIRY -gt 30 ]]; then
                pass "SSL certificate expires in $DAYS_UNTIL_EXPIRY days"
            elif [[ $DAYS_UNTIL_EXPIRY -gt 7 ]]; then
                warn "SSL certificate expires in $DAYS_UNTIL_EXPIRY days" "Consider renewing soon"
            else
                fail "SSL certificate expires in $DAYS_UNTIL_EXPIRY days" "Renew immediately"
            fi
        fi
    else
        fail "Unable to verify SSL certificate" "Check certificate configuration"
    fi
    
    # 11. TLS version
    info "Testing TLS version..."
    TLS_VERSION=$(echo | openssl s_client -servername "$(echo $DOMAIN | sed 's|https://||' | sed 's|/.*||')" -connect "$(echo $DOMAIN | sed 's|https://||' | sed 's|/.*||'):443" 2>/dev/null | grep "Protocol" | head -1)
    if echo "$TLS_VERSION" | grep -q "TLSv1.2\|TLSv1.3"; then
        pass "Secure TLS version in use: $TLS_VERSION"
    else
        warn "TLS version may be insecure: $TLS_VERSION" "Use TLS 1.2 or 1.3"
    fi
else
    info "Skipping SSL/TLS tests (HTTP domain)"
fi

# Results summary
echo -e "\n${BLUE}📋 Security Validation Summary${NC}"
echo "================================"
echo -e "✅ ${GREEN}Passed: $PASSED${NC}"
echo -e "⚠️  ${YELLOW}Warnings: $WARNINGS${NC}"
echo -e "❌ ${RED}Failed: $FAILED${NC}"
echo -e "📊 Total Tests: $((PASSED + WARNINGS + FAILED))"

# Security score calculation
TOTAL_TESTS=$((PASSED + WARNINGS + FAILED))
if [[ $TOTAL_TESTS -gt 0 ]]; then
    SECURITY_SCORE=$(( (PASSED * 100) / TOTAL_TESTS ))
    echo -e "🏆 Security Score: ${SECURITY_SCORE}%"
    
    if [[ $SECURITY_SCORE -ge 90 ]]; then
        echo -e "\n🎉 ${GREEN}Excellent! Your application has strong security posture.${NC}"
    elif [[ $SECURITY_SCORE -ge 75 ]]; then
        echo -e "\n👍 ${YELLOW}Good security posture. Address warnings for improvement.${NC}"
    elif [[ $SECURITY_SCORE -ge 60 ]]; then
        echo -e "\n⚠️  ${YELLOW}Moderate security. Several issues need attention.${NC}"
    else
        echo -e "\n🚨 ${RED}Security needs significant improvement before production use.${NC}"
    fi
fi

# Recommendations
echo -e "\n${PURPLE}💡 Next Steps${NC}"
echo "=============="

if [[ $FAILED -gt 0 ]]; then
    echo -e "${RED}High Priority:${NC}"
    echo "  • Address all failed security tests immediately"
    echo "  • Review and implement missing security headers"
    echo "  • Ensure proper authentication and authorization"
fi

if [[ $WARNINGS -gt 0 ]]; then
    echo -e "${YELLOW}Medium Priority:${NC}"
    echo "  • Review and address security warnings"
    echo "  • Consider implementing additional security measures"
    echo "  • Plan for security improvements in next release"
fi

echo -e "${GREEN}Ongoing:${NC}"
echo "  • Set up automated security monitoring"
echo "  • Schedule regular security assessments"
echo "  • Keep dependencies and certificates up to date"
echo "  • Review security logs regularly"

# Exit with appropriate code
if [[ $FAILED -gt 0 ]]; then
    exit 1
elif [[ $WARNINGS -gt 0 ]]; then
    exit 2
else
    exit 0
fi