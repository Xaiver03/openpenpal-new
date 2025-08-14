#!/bin/bash

# Direct middleware test without proxy interference

# Unset proxy variables
unset http_proxy https_proxy HTTP_PROXY HTTPS_PROXY

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

API_BASE="http://127.0.0.1:8080/api/v1"

echo -e "${BLUE}=== Direct Middleware Test ===${NC}\n"

# Test 1: CORS
echo -e "${YELLOW}1. Testing CORS${NC}"
CORS_TEST=$(curl -s -o /dev/null -w "%{http_code}" -H "Origin: http://localhost:3000" -X OPTIONS "$API_BASE/auth/login")
if [ "$CORS_TEST" = "204" ]; then
    echo -e "${GREEN}✓ CORS working correctly${NC}"
else
    echo -e "${RED}✗ CORS failed (status: $CORS_TEST)${NC}"
fi

# Test 2: Health check
echo -e "\n${YELLOW}2. Testing Health Check${NC}"
HEALTH=$(curl -s "$API_BASE/../health" | grep -o '"status":"healthy"')
if [ -n "$HEALTH" ]; then
    echo -e "${GREEN}✓ Health check passed${NC}"
else
    echo -e "${RED}✗ Health check failed${NC}"
fi

# Test 3: Rate limiting
echo -e "\n${YELLOW}3. Testing Rate Limiting${NC}"
RATE_LIMITED=false
for i in {1..20}; do
    STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$API_BASE/test/rate-limit")
    if [ "$STATUS" = "429" ]; then
        RATE_LIMITED=true
        echo -e "${GREEN}✓ Rate limiting triggered after $i requests${NC}"
        break
    fi
done

if [ "$RATE_LIMITED" = false ]; then
    echo -e "${YELLOW}⚠ Rate limiting not triggered in 20 requests (may be in test mode)${NC}"
fi

# Test 4: CSRF Token
echo -e "\n${YELLOW}4. Testing CSRF Token${NC}"
CSRF_RESPONSE=$(curl -s "$API_BASE/auth/csrf")
CSRF_TOKEN=$(echo "$CSRF_RESPONSE" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
if [ -n "$CSRF_TOKEN" ]; then
    echo -e "${GREEN}✓ CSRF token generated: ${CSRF_TOKEN:0:10}...${NC}"
else
    echo -e "${RED}✗ CSRF token generation failed${NC}"
fi

# Test 5: Authentication
echo -e "\n${YELLOW}5. Testing Authentication${NC}"
LOGIN_RESPONSE=$(curl -s -X POST \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"password123"}' \
    "$API_BASE/auth/login")
    
TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
if [ -n "$TOKEN" ]; then
    echo -e "${GREEN}✓ Login successful, token received${NC}"
    
    # Test authenticated request
    ME_RESPONSE=$(curl -s -H "Authorization: Bearer $TOKEN" "$API_BASE/auth/me")
    if echo "$ME_RESPONSE" | grep -q '"username":"admin"'; then
        echo -e "${GREEN}✓ Authenticated request successful${NC}"
    else
        echo -e "${RED}✗ Authenticated request failed${NC}"
    fi
else
    echo -e "${RED}✗ Login failed${NC}"
fi

# Test 6: Response transformation
echo -e "\n${YELLOW}6. Testing Response Transformation${NC}"
if echo "$ME_RESPONSE" | grep -q '"createdAt"'; then
    echo -e "${GREEN}✓ Response transformed to camelCase${NC}"
elif echo "$ME_RESPONSE" | grep -q '"created_at"'; then
    echo -e "${RED}✗ Response still in snake_case${NC}"
else
    echo -e "${YELLOW}⚠ Could not verify transformation${NC}"
fi

# Test 7: Security headers
echo -e "\n${YELLOW}7. Testing Security Headers${NC}"
HEADERS=$(curl -s -I "$API_BASE/../health")
SECURITY_HEADERS=("X-Content-Type-Options" "X-Frame-Options" "X-XSS-Protection" "Content-Security-Policy")
ALL_PRESENT=true

for header in "${SECURITY_HEADERS[@]}"; do
    if echo "$HEADERS" | grep -qi "$header"; then
        echo -e "${GREEN}✓ $header present${NC}"
    else
        echo -e "${RED}✗ $header missing${NC}"
        ALL_PRESENT=false
    fi
done

echo -e "\n${BLUE}=== Test Complete ===${NC}"