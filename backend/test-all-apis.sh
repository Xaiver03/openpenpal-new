#!/bin/bash

# OpenPenPal API Test Script
# Tests all major API endpoints

echo "========================================="
echo "OpenPenPal API Test Suite"
echo "========================================="
echo ""

BASE_URL="http://localhost:8080"

# Color codes for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test function
test_api() {
    local test_name="$1"
    local method="$2"
    local endpoint="$3"
    local data="$4"
    local headers="$5"
    
    echo -n "Testing $test_name... "
    
    if [ -z "$headers" ]; then
        response=$(curl -s -w "\n%{http_code}" -X "$method" "$BASE_URL$endpoint" \
            -H "Content-Type: application/json" \
            -d "$data" 2>/dev/null)
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" "$BASE_URL$endpoint" \
            -H "Content-Type: application/json" \
            $headers \
            -d "$data" 2>/dev/null)
    fi
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    if [[ "$http_code" =~ ^2[0-9][0-9]$ ]]; then
        echo -e "${GREEN}✓ PASS${NC} (HTTP $http_code)"
        echo "Response: $body" | head -c 100
        echo ""
    else
        echo -e "${RED}✗ FAIL${NC} (HTTP $http_code)"
        echo "Response: $body" | head -c 100
        echo ""
    fi
    
    echo ""
}

echo "1. Testing Authentication APIs"
echo "------------------------------"

# Test admin login
test_api "Admin Login" "POST" "/api/v1/auth/login" \
    '{"username": "admin", "password": "admin123"}'

# Extract token from response (simple grep)
ADMIN_TOKEN=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username": "admin", "password": "admin123"}' | grep -o '"token":"[^"]*' | grep -o '[^"]*$')

# Test user login
test_api "User Login (Alice)" "POST" "/api/v1/auth/login" \
    '{"username": "alice", "password": "secret"}'

USER_TOKEN=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username": "alice", "password": "secret"}' | grep -o '"token":"[^"]*' | grep -o '[^"]*$')

# Test courier login
test_api "Courier Login" "POST" "/api/v1/auth/login" \
    '{"username": "courier_level4", "password": "secret"}'

COURIER_TOKEN=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username": "courier_level4", "password": "secret"}' | grep -o '"token":"[^"]*' | grep -o '[^"]*$')

echo "2. Testing User APIs"
echo "--------------------"

test_api "Get Current User (Admin)" "GET" "/api/v1/users/me" "" \
    "-H \"Authorization: Bearer $ADMIN_TOKEN\""

test_api "Get Current User (Alice)" "GET" "/api/v1/users/me" "" \
    "-H \"Authorization: Bearer $USER_TOKEN\""

echo "3. Testing Letter Museum APIs"
echo "------------------------------"

# Test letter templates
test_api "Get Letter Templates" "GET" "/api/v1/museum/templates" "" \
    "-H \"Authorization: Bearer $USER_TOKEN\""

# Test museum statistics
test_api "Get Museum Statistics" "GET" "/api/v1/museum/statistics" "" \
    "-H \"Authorization: Bearer $USER_TOKEN\""

# Test featured exhibitions
test_api "Get Featured Exhibitions" "GET" "/api/v1/museum/exhibitions/featured" "" \
    "-H \"Authorization: Bearer $USER_TOKEN\""

echo "4. Testing Letter APIs"
echo "----------------------"

# Create a test letter
test_api "Create Letter" "POST" "/api/v1/letters" \
    '{"recipient_id": "bob", "subject": "Test Letter", "content": "This is a test letter from the API test suite."}' \
    "-H \"Authorization: Bearer $USER_TOKEN\""

# Get user's letters
test_api "Get My Letters" "GET" "/api/v1/letters" "" \
    "-H \"Authorization: Bearer $USER_TOKEN\""

echo "5. Testing Courier APIs"
echo "-----------------------"

test_api "Get Courier Tasks" "GET" "/api/v1/courier/tasks" "" \
    "-H \"Authorization: Bearer $COURIER_TOKEN\""

test_api "Get Courier Statistics" "GET" "/api/v1/courier/statistics" "" \
    "-H \"Authorization: Bearer $COURIER_TOKEN\""

echo "6. Testing Admin APIs"
echo "---------------------"

test_api "Get All Users (Admin)" "GET" "/api/v1/admin/users" "" \
    "-H \"Authorization: Bearer $ADMIN_TOKEN\""

echo "7. Testing Envelope APIs"
echo "------------------------"

# Create an envelope
test_api "Create Envelope" "POST" "/api/v1/envelopes" \
    '{"name": "Test Envelope", "description": "API Test Envelope"}' \
    "-H \"Authorization: Bearer $USER_TOKEN\""

# Get user's envelopes
test_api "Get My Envelopes" "GET" "/api/v1/envelopes" "" \
    "-H \"Authorization: Bearer $USER_TOKEN\""

echo "8. Testing Credit System APIs"
echo "-----------------------------"

test_api "Get My Credits" "GET" "/api/v1/credits/balance" "" \
    "-H \"Authorization: Bearer $USER_TOKEN\""

echo "9. Testing Notification APIs"
echo "----------------------------"

test_api "Get My Notifications" "GET" "/api/v1/notifications" "" \
    "-H \"Authorization: Bearer $USER_TOKEN\""

echo "========================================="
echo "API Test Suite Completed"
echo "========================================="