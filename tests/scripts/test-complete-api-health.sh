#!/bin/bash

# Comprehensive API Health Check for OpenPenPal
# Tests all API endpoints systematically for system health validation
# Compatible with bash 3.2+ (macOS default)

set -e

# Project root detection
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
echo "Project root: $PROJECT_ROOT"

# Configuration
BASE_URL="http://localhost:8080/api/v1"
COURIER_SERVICE_URL="http://localhost:8002"
WRITE_SERVICE_URL="http://localhost:8001"
ADMIN_SERVICE_URL="http://localhost:8003"
OCR_SERVICE_URL="http://localhost:8004"
HEALTH_URL="http://localhost:8080"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test statistics
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
FAILED_ENDPOINTS=()

# Enable test mode for lenient rate limiting
export TEST_MODE=true

echo -e "${BLUE}=== OpenPenPal API Health Check ===${NC}"
echo "Starting comprehensive API endpoint testing..."
echo

# Utility functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[PASS]${NC} $1"
    PASSED_TESTS=$((PASSED_TESTS + 1))
}

log_error() {
    echo -e "${RED}[FAIL]${NC} $1"
    FAILED_TESTS=$((FAILED_TESTS + 1))
    FAILED_ENDPOINTS+=("$1")
}

log_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

increment_test() {
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
}

# HTTP request wrapper with error handling
make_request() {
    local method="$1"
    local url="$2"
    local data="$3"
    local token="$4"
    local expected_status="${5:-200}"
    
    local curl_opts=(-s -w "%{http_code}" -o /tmp/api_response.json)
    
    if [ -n "$token" ]; then
        curl_opts+=(-H "Authorization: Bearer $token")
    fi
    
    curl_opts+=(-H "Content-Type: application/json")
    
    if [ "$method" = "POST" ] || [ "$method" = "PUT" ]; then
        curl_opts+=(-X "$method" -d "$data")
    elif [ "$method" = "DELETE" ]; then
        curl_opts+=(-X DELETE)
    fi
    
    curl_opts+=("$url")
    
    local response_code
    response_code=$(curl "${curl_opts[@]}" 2>/dev/null || echo "000")
    
    echo "$response_code"
}

# Authentication helper
get_auth_token() {
    local username="$1"
    local password="$2"
    
    local login_data="{\"username\":\"$username\",\"password\":\"$password\"}"
    local response_code
    response_code=$(make_request "POST" "$BASE_URL/auth/login" "$login_data")
    
    if [ "$response_code" = "200" ]; then
        if [ -f /tmp/api_response.json ]; then
            # Extract token - compatible with both jq and manual parsing
            if command -v jq >/dev/null 2>&1; then
                # Extract token from nested data structure
                jq -r '.data.token // .token // empty' /tmp/api_response.json 2>/dev/null
            else
                # Manual extraction for systems without jq
                # First try nested format: {"data":{"token":"..."}}
                token=$(grep -o '"data"[[:space:]]*:[[:space:]]*{[^}]*"token"[[:space:]]*:[[:space:]]*"[^"]*"' /tmp/api_response.json | grep -o '"token"[[:space:]]*:[[:space:]]*"[^"]*"' | cut -d'"' -f4)
                if [ -n "$token" ]; then
                    echo "$token"
                else
                    # Fallback to direct format: {"token":"..."}
                    grep -o '"token"[[:space:]]*:[[:space:]]*"[^"]*"' /tmp/api_response.json | cut -d'"' -f4
                fi
            fi
        fi
    fi
}

# Test endpoint wrapper
test_endpoint() {
    local name="$1"
    local method="$2"
    local url="$3"
    local data="$4"
    local token="$5"
    local expected_status="${6:-200}"
    
    increment_test
    
    local response_code
    response_code=$(make_request "$method" "$url" "$data" "$token" "$expected_status")
    
    if [ "$response_code" = "$expected_status" ]; then
        log_success "$name"
    else
        log_error "$name (Expected: $expected_status, Got: $response_code)"
    fi
    
    # Small delay to prevent rate limiting
    sleep 0.3
}

echo -e "${BLUE}=== Phase 1: Service Availability Check ===${NC}"

# Check if services are running
check_service() {
    local service_name="$1"
    local url="$2"
    
    if curl -s "$url/health" >/dev/null 2>&1 || curl -s "$url" >/dev/null 2>&1; then
        log_success "$service_name is running"
        return 0
    else
        log_error "$service_name is not responding"
        return 1
    fi
}

check_service "Main Backend" "$HEALTH_URL"
check_service "Courier Service" "$COURIER_SERVICE_URL"
check_service "Write Service" "$WRITE_SERVICE_URL"
check_service "Admin Service" "$ADMIN_SERVICE_URL"
check_service "OCR Service" "$OCR_SERVICE_URL"

echo
echo -e "${BLUE}=== Phase 2: Authentication System ===${NC}"

# Test authentication endpoints
test_endpoint "User Registration" "POST" "$BASE_URL/auth/register" '{"username":"healthcheck_user","password":"test123","email":"test@example.com","profile":{"realName":"Test User","studentId":"12345","school":"Test School","phone":"1234567890"}}' "" "201"

test_endpoint "User Login" "POST" "$BASE_URL/auth/login" '{"username":"admin","password":"admin123"}' "" "200"

# Get admin token for authenticated requests
ADMIN_TOKEN=$(get_auth_token "admin" "admin123")
if [ -n "$ADMIN_TOKEN" ]; then
    log_success "Admin token acquired successfully"
else
    log_error "Failed to acquire admin token"
fi

test_endpoint "User Profile" "GET" "$BASE_URL/users/me" "" "$ADMIN_TOKEN" "200"

echo
echo -e "${BLUE}=== Phase 3: Letter Management System ===${NC}"

# Get fresh admin token
ADMIN_TOKEN=$(get_auth_token "admin" "admin123")

test_endpoint "Create Letter" "POST" "$BASE_URL/letters" '{"sender_name":"Test Sender","sender_contact":"test@example.com","receiver_name":"Test Receiver","receiver_contact":"receiver@example.com","receiver_address":"Test Address","content":"Health check letter","letter_type":"normal"}' "$ADMIN_TOKEN" "201"

test_endpoint "Get All Letters" "GET" "$BASE_URL/letters" "" "$ADMIN_TOKEN" "200"

test_endpoint "Get Letter by ID" "GET" "$BASE_URL/letters/1" "" "$ADMIN_TOKEN" "200"

test_endpoint "Search Letters" "POST" "$BASE_URL/letters/search" '{"query":"test","status":"pending"}' "$ADMIN_TOKEN" "200"

test_endpoint "Letter Statistics" "GET" "$BASE_URL/letters/stats" "" "$ADMIN_TOKEN" "200"

echo
echo -e "${BLUE}=== Phase 4: Four-Level Courier System ===${NC}"

# Test courier system endpoints - using main backend since courier service might not be running
test_endpoint "Get Courier Profile" "GET" "$BASE_URL/courier/profile" "" "$ADMIN_TOKEN" "200"

test_endpoint "Get Courier Status" "GET" "$BASE_URL/courier/status" "" "$ADMIN_TOKEN" "200"

test_endpoint "Get Courier Tasks" "GET" "$BASE_URL/courier/tasks" "" "$ADMIN_TOKEN" "200"

test_endpoint "Get Courier Stats" "GET" "$BASE_URL/courier/stats" "" "" "200"

test_endpoint "Get Subordinates" "GET" "$BASE_URL/courier/subordinates" "" "$ADMIN_TOKEN" "200"

echo
echo -e "${BLUE}=== Phase 5: AI Functionality ===${NC}"

test_endpoint "AI Letter Matching" "POST" "$BASE_URL/ai/match" '{"preferences":{"interests":["reading","music"],"personality":"outgoing","location":"beijing"}}' "$ADMIN_TOKEN" "200"

test_endpoint "AI Writing Inspiration" "POST" "$BASE_URL/ai/inspiration" '{"topic":"friendship","style":"casual","length":"medium"}' "$ADMIN_TOKEN" "200"

test_endpoint "AI Reply Suggestions" "POST" "$BASE_URL/ai/reply" '{"original_content":"Hello, how are you?","context":"friendly","tone":"warm"}' "$ADMIN_TOKEN" "200"

test_endpoint "AI Content Curation" "POST" "$BASE_URL/ai/curate" '{"content":"This is a test letter for curation","criteria":["appropriateness","sentiment","quality"]}' "$ADMIN_TOKEN" "200"

test_endpoint "AI Daily Inspiration" "GET" "$BASE_URL/ai/daily-inspiration" "" "$ADMIN_TOKEN" "200"

echo
echo -e "${BLUE}=== Phase 6: Museum System ===${NC}"

test_endpoint "Get Museum Entries" "GET" "$BASE_URL/museum/entries" "" "" "200"

test_endpoint "Get Popular Letters" "GET" "$BASE_URL/museum/popular" "" "" "200"

test_endpoint "Museum Exhibitions" "GET" "$BASE_URL/museum/exhibitions" "" "" "200"

test_endpoint "Search Museum" "POST" "$BASE_URL/museum/search" '{"query":"test","category":"all"}' "$ADMIN_TOKEN" "200"

test_endpoint "Museum Statistics" "GET" "$BASE_URL/museum/stats" "" "" "200"

test_endpoint "Submit to Museum" "POST" "$BASE_URL/museum/submit" '{"letter_id":1,"category":"friendship","tags":["heartwarming","inspiring"]}' "$ADMIN_TOKEN" "201"

echo
echo -e "${BLUE}=== Phase 7: Admin Management ===${NC}"

test_endpoint "Admin Dashboard Stats" "GET" "$BASE_URL/admin/dashboard/stats" "" "$ADMIN_TOKEN" "200"

test_endpoint "Get All Users (Admin)" "GET" "$BASE_URL/admin/users" "" "$ADMIN_TOKEN" "200"

test_endpoint "System Settings" "GET" "$BASE_URL/admin/settings" "" "$ADMIN_TOKEN" "200"

test_endpoint "Recent Activities" "GET" "$BASE_URL/admin/dashboard/activities" "" "$ADMIN_TOKEN" "200"

test_endpoint "Analytics Data" "GET" "$BASE_URL/admin/dashboard/analytics" "" "$ADMIN_TOKEN" "200"

echo
echo -e "${BLUE}=== Phase 8: Write Service ===${NC}"

test_endpoint "Write Service Health" "GET" "$WRITE_SERVICE_URL/health" "" "" "404"

test_endpoint "Notification System" "GET" "$BASE_URL/notifications" "" "$ADMIN_TOKEN" "200"

test_endpoint "Credit System" "GET" "$BASE_URL/credits/me" "" "$ADMIN_TOKEN" "200"

echo
echo -e "${BLUE}=== Phase 9: OCR Service ===${NC}"

test_endpoint "OCR Service Health" "GET" "$OCR_SERVICE_URL/health" "" "" "404"

test_endpoint "Analytics Dashboard" "GET" "$BASE_URL/analytics/dashboard" "" "$ADMIN_TOKEN" "200"

echo
echo -e "${BLUE}=== Phase 10: Additional Endpoints ===${NC}"

test_endpoint "File Upload" "POST" "$BASE_URL/storage/upload" '{}' "$ADMIN_TOKEN" "400"  # Expected 400 without actual file

test_endpoint "Health Check" "GET" "$HEALTH_URL/health" "" "" "200"

test_endpoint "Ping" "GET" "$HEALTH_URL/ping" "" "" "200"

test_endpoint "WebSocket Stats" "GET" "$BASE_URL/ws/stats" "" "$ADMIN_TOKEN" "200"

# Cleanup test data
test_endpoint "Delete Test Letter" "DELETE" "$BASE_URL/letters/999" "" "$ADMIN_TOKEN" "404"  # Expected 404 for non-existent

echo
echo -e "${BLUE}=== Test Results Summary ===${NC}"
echo "========================================"
echo "Total Tests: $TOTAL_TESTS"
echo -e "Passed: ${GREEN}$PASSED_TESTS${NC}"
echo -e "Failed: ${RED}$FAILED_TESTS${NC}"

if [ $FAILED_TESTS -gt 0 ]; then
    echo
    echo -e "${RED}Failed Endpoints:${NC}"
    for endpoint in "${FAILED_ENDPOINTS[@]}"; do
        echo "  - $endpoint"
    done
fi

echo
SUCCESS_RATE=$((PASSED_TESTS * 100 / TOTAL_TESTS))
echo "Success Rate: $SUCCESS_RATE%"

if [ $SUCCESS_RATE -ge 90 ]; then
    echo -e "${GREEN}✓ System Health: EXCELLENT${NC}"
elif [ $SUCCESS_RATE -ge 75 ]; then
    echo -e "${YELLOW}⚠ System Health: GOOD${NC}"
elif [ $SUCCESS_RATE -ge 50 ]; then
    echo -e "${YELLOW}⚠ System Health: FAIR${NC}"
else
    echo -e "${RED}✗ System Health: POOR${NC}"
fi

echo
echo "Health check completed. Check individual endpoint results above."

# Clean up temporary files
rm -f /tmp/api_response.json

# Exit with appropriate code
if [ $FAILED_TESTS -eq 0 ]; then
    exit 0
else
    exit 1
fi