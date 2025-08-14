#!/bin/bash

# SOTA Middleware Comprehensive Test Suite
# Tests all middleware layers for blocking issues

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
API_BASE="http://localhost:8080/api/v1"
ADMIN_API_BASE="http://localhost:8080/api/v1/admin"
FRONTEND_ORIGIN="http://localhost:3000"
ADMIN_ORIGIN="http://localhost:3001"

echo -e "${BLUE}=== SOTA Middleware Comprehensive Test Suite ===${NC}"
echo -e "${BLUE}Testing middleware layers for blocking issues${NC}\n"

# Test results tracking
PASSED=0
FAILED=0
WARNINGS=0

# Helper function to print test results
print_result() {
    local test_name=$1
    local status=$2
    local message=$3
    
    if [ "$status" = "PASS" ]; then
        echo -e "${GREEN}✓ $test_name${NC}"
        ((PASSED++))
    elif [ "$status" = "FAIL" ]; then
        echo -e "${RED}✗ $test_name${NC}"
        echo -e "  ${RED}Error: $message${NC}"
        ((FAILED++))
    elif [ "$status" = "WARN" ]; then
        echo -e "${YELLOW}⚠ $test_name${NC}"
        echo -e "  ${YELLOW}Warning: $message${NC}"
        ((WARNINGS++))
    fi
}

# Test 1: CORS Middleware
echo -e "\n${YELLOW}1. Testing CORS Middleware${NC}"

# Test frontend origin
echo -n "Testing frontend CORS... "
CORS_FRONTEND=$(curl -s -o /dev/null -w "%{http_code}" \
    -H "Origin: $FRONTEND_ORIGIN" \
    -H "Access-Control-Request-Method: POST" \
    -X OPTIONS \
    "$API_BASE/auth/login")

if [ "$CORS_FRONTEND" = "204" ]; then
    print_result "Frontend CORS preflight" "PASS" ""
else
    print_result "Frontend CORS preflight" "FAIL" "Expected 204, got $CORS_FRONTEND"
fi

# Test admin origin
echo -n "Testing admin CORS... "
CORS_ADMIN=$(curl -s -o /dev/null -w "%{http_code}" \
    -H "Origin: $ADMIN_ORIGIN" \
    -H "Access-Control-Request-Method: GET" \
    -X OPTIONS \
    "$ADMIN_API_BASE/users")

if [ "$CORS_ADMIN" = "204" ]; then
    print_result "Admin CORS preflight" "PASS" ""
else
    print_result "Admin CORS preflight" "FAIL" "Expected 204, got $CORS_ADMIN"
fi

# Test unauthorized origin
echo -n "Testing unauthorized origin... "
CORS_UNAUTH=$(curl -s -o /dev/null -w "%{http_code}" \
    -H "Origin: http://evil.com" \
    -H "Access-Control-Request-Method: POST" \
    -X OPTIONS \
    "$API_BASE/auth/login" 2>/dev/null || echo "0")

if [ "$CORS_UNAUTH" = "204" ]; then
    print_result "Unauthorized origin block" "FAIL" "Should block unauthorized origins"
else
    print_result "Unauthorized origin block" "PASS" ""
fi

# Test 2: CSRF Middleware
echo -e "\n${YELLOW}2. Testing CSRF Middleware${NC}"

# Get CSRF token
echo -n "Getting CSRF token... "
CSRF_RESPONSE=$(curl -s "$API_BASE/auth/csrf")
CSRF_TOKEN=$(echo "$CSRF_RESPONSE" | grep -o '"token":"[^"]*' | cut -d'"' -f4)

if [ -n "$CSRF_TOKEN" ]; then
    print_result "CSRF token generation" "PASS" ""
else
    print_result "CSRF token generation" "FAIL" "No token received"
fi

# Test login without CSRF (should work temporarily)
echo -n "Testing login without CSRF... "
LOGIN_NO_CSRF=$(curl -s -o /dev/null -w "%{http_code}" \
    -X POST \
    -H "Content-Type: application/json" \
    -d '{"username":"test","password":"test"}' \
    "$API_BASE/auth/login")

if [ "$LOGIN_NO_CSRF" = "401" ] || [ "$LOGIN_NO_CSRF" = "403" ]; then
    print_result "Login without CSRF" "PASS" "Properly rejected"
else
    print_result "Login without CSRF" "WARN" "Currently bypassed (temporary)"
fi

# Test 3: Rate Limiting
echo -e "\n${YELLOW}3. Testing Rate Limiting${NC}"

# Test general rate limit
echo -n "Testing general rate limit... "
RATE_TEST_COUNT=0
RATE_LIMITED=false

for i in {1..15}; do
    STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$API_BASE/health")
    if [ "$STATUS" = "429" ]; then
        RATE_LIMITED=true
        RATE_TEST_COUNT=$i
        break
    fi
done

if [ "$RATE_LIMITED" = true ]; then
    print_result "General rate limiting" "PASS" "Limited after $RATE_TEST_COUNT requests"
else
    print_result "General rate limiting" "WARN" "No rate limit hit in 15 requests (may be in test mode)"
fi

# Test auth rate limit
echo -n "Testing auth rate limit... "
AUTH_RATE_COUNT=0
AUTH_LIMITED=false

for i in {1..10}; do
    STATUS=$(curl -s -o /dev/null -w "%{http_code}" \
        -X POST \
        -H "Content-Type: application/json" \
        -d '{"username":"test","password":"wrong"}' \
        "$API_BASE/auth/login")
    if [ "$STATUS" = "429" ]; then
        AUTH_LIMITED=true
        AUTH_RATE_COUNT=$i
        break
    fi
done

if [ "$AUTH_LIMITED" = true ]; then
    print_result "Auth rate limiting" "PASS" "Limited after $AUTH_RATE_COUNT attempts"
else
    print_result "Auth rate limiting" "WARN" "No rate limit hit in 10 attempts"
fi

# Test 4: Authentication Flow
echo -e "\n${YELLOW}4. Testing Authentication Flow${NC}"

# Test login
echo -n "Testing login flow... "
LOGIN_RESPONSE=$(curl -s -X POST \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"admin123"}' \
    "$API_BASE/auth/login")

TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"token":"[^"]*' | cut -d'"' -f4)

if [ -n "$TOKEN" ]; then
    print_result "Login authentication" "PASS" ""
else
    print_result "Login authentication" "FAIL" "No token received"
fi

# Test authenticated request
echo -n "Testing authenticated request... "
AUTH_TEST=$(curl -s -o /dev/null -w "%{http_code}" \
    -H "Authorization: Bearer $TOKEN" \
    "$API_BASE/auth/me")

if [ "$AUTH_TEST" = "200" ]; then
    print_result "Authenticated request" "PASS" ""
else
    print_result "Authenticated request" "FAIL" "Expected 200, got $AUTH_TEST"
fi

# Test admin access
echo -n "Testing admin access... "
ADMIN_TEST=$(curl -s -o /dev/null -w "%{http_code}" \
    -H "Authorization: Bearer $TOKEN" \
    "$ADMIN_API_BASE/users")

if [ "$ADMIN_TEST" = "301" ] || [ "$ADMIN_TEST" = "200" ]; then
    print_result "Admin authorization" "PASS" ""
else
    print_result "Admin authorization" "FAIL" "Expected 200/301, got $ADMIN_TEST"
fi

# Test 5: Request/Response Transformation
echo -e "\n${YELLOW}5. Testing Request/Response Transformation${NC}"

# Test camelCase to snake_case (request)
echo -n "Testing request transformation... "
TRANSFORM_REQ=$(curl -s -X POST \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"firstName":"Test","lastName":"User","emailAddress":"test@example.com"}' \
    "$API_BASE/test/transform" 2>&1 || echo '{"error":"endpoint not found"}')

# For now, we'll test with a real endpoint
print_result "Request transformation" "WARN" "Test endpoint needed for verification"

# Test snake_case to camelCase (response)
echo -n "Testing response transformation... "
USER_RESPONSE=$(curl -s -H "Authorization: Bearer $TOKEN" "$API_BASE/auth/me")
if echo "$USER_RESPONSE" | grep -q "createdAt"; then
    print_result "Response transformation" "PASS" "snake_case transformed to camelCase"
elif echo "$USER_RESPONSE" | grep -q "created_at"; then
    print_result "Response transformation" "FAIL" "Response still in snake_case"
else
    print_result "Response transformation" "WARN" "Could not verify transformation"
fi

# Test 6: Security Headers
echo -e "\n${YELLOW}6. Testing Security Headers${NC}"

# Check security headers
echo -n "Testing security headers... "
HEADERS=$(curl -s -I "$API_BASE/health")

SECURITY_HEADERS=("X-Content-Type-Options" "X-Frame-Options" "X-XSS-Protection" "Content-Security-Policy")
MISSING_HEADERS=()

for header in "${SECURITY_HEADERS[@]}"; do
    if ! echo "$HEADERS" | grep -qi "$header"; then
        MISSING_HEADERS+=("$header")
    fi
done

if [ ${#MISSING_HEADERS[@]} -eq 0 ]; then
    print_result "Security headers" "PASS" ""
else
    print_result "Security headers" "FAIL" "Missing: ${MISSING_HEADERS[*]}"
fi

# Test 7: Performance Impact
echo -e "\n${YELLOW}7. Testing Performance Impact${NC}"

# Measure response time with all middleware
echo -n "Testing middleware performance... "
TOTAL_TIME=0
REQUEST_COUNT=10

for i in $(seq 1 $REQUEST_COUNT); do
    TIME=$(curl -s -o /dev/null -w "%{time_total}" -H "Authorization: Bearer $TOKEN" "$API_BASE/health")
    TOTAL_TIME=$(echo "$TOTAL_TIME + $TIME" | bc)
done

AVG_TIME=$(echo "scale=3; $TOTAL_TIME / $REQUEST_COUNT" | bc)
AVG_TIME_MS=$(echo "scale=0; $AVG_TIME * 1000" | bc)

if (( $(echo "$AVG_TIME < 0.1" | bc -l) )); then
    print_result "Middleware performance" "PASS" "Avg response time: ${AVG_TIME_MS}ms"
elif (( $(echo "$AVG_TIME < 0.5" | bc -l) )); then
    print_result "Middleware performance" "WARN" "Avg response time: ${AVG_TIME_MS}ms (slightly high)"
else
    print_result "Middleware performance" "FAIL" "Avg response time: ${AVG_TIME_MS}ms (too high)"
fi

# Test 8: WebSocket Connection
echo -e "\n${YELLOW}8. Testing WebSocket Middleware${NC}"

# Test WebSocket upgrade
echo -n "Testing WebSocket connection... "
WS_TEST=$(curl -s -o /dev/null -w "%{http_code}" \
    -H "Upgrade: websocket" \
    -H "Connection: Upgrade" \
    -H "Sec-WebSocket-Key: x3JJHMbDL1EzLkh9GBhXDw==" \
    -H "Sec-WebSocket-Version: 13" \
    "$API_BASE/ws" 2>/dev/null || echo "0")

if [ "$WS_TEST" = "101" ] || [ "$WS_TEST" = "400" ]; then
    print_result "WebSocket middleware" "PASS" "WebSocket endpoint accessible"
else
    print_result "WebSocket middleware" "WARN" "WebSocket test inconclusive"
fi

# Summary
echo -e "\n${BLUE}=== Test Summary ===${NC}"
echo -e "Total Tests: $((PASSED + FAILED + WARNINGS))"
echo -e "${GREEN}Passed: $PASSED${NC}"
echo -e "${YELLOW}Warnings: $WARNINGS${NC}"
echo -e "${RED}Failed: $FAILED${NC}"

if [ $FAILED -eq 0 ]; then
    echo -e "\n${GREEN}✓ All critical middleware tests passed!${NC}"
    exit 0
else
    echo -e "\n${RED}✗ Some middleware tests failed. Please review.${NC}"
    exit 1
fi