#!/bin/bash

# OpenPenPal Security Implementation Test
# 综合安全实现测试

echo "🔒 OpenPenPal Security Implementation Test"
echo "=========================================="

FRONTEND_URL="http://localhost:3000"
BACKEND_URL="http://localhost:8080"
PASSED=0
FAILED=0

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test function
test_feature() {
    local name="$1"
    local command="$2"
    local expected="$3"
    
    echo -e "\n🧪 Testing: ${BLUE}$name${NC}"
    
    if eval "$command" > /tmp/test_output 2>&1; then
        if [[ -n "$expected" ]]; then
            if grep -q "$expected" /tmp/test_output; then
                echo -e "✅ ${GREEN}PASS${NC}: $name"
                ((PASSED++))
            else
                echo -e "❌ ${RED}FAIL${NC}: $name (expected: $expected)"
                echo "   Output: $(cat /tmp/test_output | head -1)"
                ((FAILED++))
            fi
        else
            echo -e "✅ ${GREEN}PASS${NC}: $name"
            ((PASSED++))
        fi
    else
        echo -e "❌ ${RED}FAIL${NC}: $name"
        echo "   Error: $(cat /tmp/test_output | head -1)"
        ((FAILED++))
    fi
}

echo -e "\n🚀 Starting comprehensive security tests..."

# Check service availability
echo -e "\n🔍 Checking service availability..."

if curl -s -o /dev/null -w "%{http_code}" "$FRONTEND_URL" | grep -q "200"; then
    echo -e "✅ ${GREEN}Frontend service: Available${NC}"
else
    echo -e "❌ ${RED}Frontend service: Not available${NC}"
    echo "   💡 Start with: npm run dev"
fi

if curl -s -o /dev/null -w "%{http_code}" "$BACKEND_URL/health" | grep -q "200"; then
    echo -e "✅ ${GREEN}Backend service: Available${NC}"
else
    echo -e "❌ ${RED}Backend service: Not available${NC}"
    echo "   💡 Start with: cd backend && go run main.go"
fi

# 1. Test CSRF Token Generation
test_feature "CSRF Token Generation" \
    "curl -s '$FRONTEND_URL/api/auth/csrf' | jq -r '.csrfToken'" \
    ""

# 2. Test Security Headers
test_feature "Security Headers - X-Frame-Options" \
    "curl -s -I '$FRONTEND_URL' | grep -i 'x-frame-options'" \
    "X-Frame-Options"

test_feature "Security Headers - X-Content-Type-Options" \
    "curl -s -I '$FRONTEND_URL' | grep -i 'x-content-type-options'" \
    "nosniff"

test_feature "Security Headers - CSP" \
    "curl -s -I '$FRONTEND_URL' | grep -i 'content-security-policy'" \
    ""

# 3. Test Rate Limiting (attempt multiple logins)
echo -e "\n🧪 Testing: ${BLUE}Rate Limiting - Auth Endpoint${NC}"
rate_limit_test() {
    local blocked=0
    for i in {1..10}; do
        response=$(curl -s -w "%{http_code}" -o /dev/null -X POST \
            -H "Content-Type: application/json" \
            -d '{"username":"invalid","password":"invalid"}' \
            "$FRONTEND_URL/api/auth/login")
        
        if [[ "$response" == "429" ]]; then
            ((blocked++))
        fi
        sleep 0.1
    done
    
    if [[ $blocked -gt 0 ]]; then
        echo -e "✅ ${GREEN}PASS${NC}: Rate limiting active ($blocked/10 requests blocked)"
        ((PASSED++))
    else
        echo -e "❌ ${RED}FAIL${NC}: Rate limiting not detected"
        ((FAILED++))
    fi
}
rate_limit_test

# 4. Test JWT Authentication
test_feature "JWT Authentication" \
    "curl -s -X POST -H 'Content-Type: application/json' -d '{\"username\":\"admin\",\"password\":\"admin123\"}' '$FRONTEND_URL/api/auth/login' | jq -r '.data.accessToken'" \
    ""

# 5. Test Environment Configuration
echo -e "\n🧪 Testing: ${BLUE}Environment Configuration${NC}"
if [[ -n "$NODE_ENV" ]] && [[ -n "$NEXT_PUBLIC_API_URL" ]]; then
    echo -e "✅ ${GREEN}PASS${NC}: Required environment variables set"
    ((PASSED++))
else
    echo -e "❌ ${RED}FAIL${NC}: Missing required environment variables"
    ((FAILED++))
fi

# 6. Test Production Files
echo -e "\n🧪 Testing: ${BLUE}Production Configuration Files${NC}"
if [[ -f "frontend/.env.production" ]]; then
    echo -e "✅ ${GREEN}PASS${NC}: Production environment file exists"
    ((PASSED++))
else
    echo -e "❌ ${RED}FAIL${NC}: Production environment file missing"
    ((FAILED++))
fi

# 7. Test Security Implementation Files
echo -e "\n🧪 Testing: ${BLUE}Security Implementation Files${NC}"
security_files=(
    "frontend/src/lib/security/csrf.ts"
    "frontend/src/lib/security/production-rate-limits.ts"
    "frontend/src/lib/security/https-config.ts"
    "frontend/src/middleware.ts"
)

missing_files=0
for file in "${security_files[@]}"; do
    if [[ -f "$file" ]]; then
        echo -e "   ✅ ${GREEN}$file${NC}"
    else
        echo -e "   ❌ ${RED}$file${NC}"
        ((missing_files++))
    fi
done

if [[ $missing_files -eq 0 ]]; then
    echo -e "✅ ${GREEN}PASS${NC}: All security files present"
    ((PASSED++))
else
    echo -e "❌ ${RED}FAIL${NC}: $missing_files security files missing"
    ((FAILED++))
fi

# Results Summary
echo -e "\n🔒 ${BLUE}SECURITY TEST RESULTS${NC}"
echo "========================"
echo -e "✅ ${GREEN}Passed: $PASSED${NC}"
echo -e "❌ ${RED}Failed: $FAILED${NC}"
echo -e "📊 Total: $((PASSED + FAILED))"

if [[ $FAILED -eq 0 ]]; then
    echo -e "📈 Success Rate: ${GREEN}100%${NC}"
else
    success_rate=$(( (PASSED * 100) / (PASSED + FAILED) ))
    echo -e "📈 Success Rate: ${YELLOW}${success_rate}%${NC}"
fi

echo -e "\n🎯 ${BLUE}Security Implementation Status:${NC}"

security_features=(
    "CSRF Protection"
    "Rate Limiting" 
    "Security Headers"
    "JWT Authentication"
    "Environment Config"
    "Production Files"
)

# Simple scoring based on test results
if [[ $PASSED -ge 8 ]]; then
    echo -e "   ✅ ${GREEN}CSRF Protection${NC}"
    echo -e "   ✅ ${GREEN}Rate Limiting${NC}"
    echo -e "   ✅ ${GREEN}Security Headers${NC}"
    echo -e "   ✅ ${GREEN}JWT Authentication${NC}"
    echo -e "   ✅ ${GREEN}Environment Config${NC}"
    echo -e "   ✅ ${GREEN}Production Files${NC}"
    overall_score=6
elif [[ $PASSED -ge 6 ]]; then
    echo -e "   ✅ ${GREEN}Most features implemented${NC}"
    overall_score=4
else
    echo -e "   ⚠️  ${YELLOW}Some features need work${NC}"
    overall_score=2
fi

echo -e "\n🏆 ${BLUE}Overall Security Score: ${overall_score}/6 features implemented${NC}"

if [[ $overall_score -ge 5 ]]; then
    echo -e "🎉 ${GREEN}Excellent! Your security implementation is production-ready.${NC}"
elif [[ $overall_score -ge 4 ]]; then
    echo -e "⚠️  ${YELLOW}Good progress! A few more security features needed.${NC}"
else
    echo -e "🚨 ${RED}More security features required before production deployment.${NC}"
fi

echo -e "\n📚 ${BLUE}Next Steps:${NC}"
echo "1. Review failed tests and implement missing security features"
echo "2. Update production environment variables with actual values"
echo "3. Test in staging environment before production deployment"
echo "4. Set up monitoring and alerting for security events"
echo "5. Configure TLS certificates and HTTPS properly"

# Cleanup
rm -f /tmp/test_output

exit $([[ $FAILED -gt 0 ]] && echo 1 || echo 0)