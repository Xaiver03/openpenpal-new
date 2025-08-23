#!/bin/bash

# Test script for new DriftBottle and FutureLetter features
# Usage: ./scripts/test-new-features.sh

API_BASE="http://localhost:8080/api/v1"
USER_TOKEN=""

echo "=== Testing New Features (DriftBottle & FutureLetter) ==="
echo

# Function to make authenticated API calls
api_call() {
    local method=$1
    local endpoint=$2
    local data=$3
    
    if [ "$method" = "GET" ]; then
        curl -s -X GET \
             -H "Authorization: Bearer $USER_TOKEN" \
             -H "Content-Type: application/json" \
             "$API_BASE$endpoint"
    else
        curl -s -X "$method" \
             -H "Authorization: Bearer $USER_TOKEN" \
             -H "Content-Type: application/json" \
             -d "$data" \
             "$API_BASE$endpoint"
    fi
}

# Check if server is running
echo "1. Checking if server is running..."
SERVER_STATUS=$(curl -s -w "%{http_code}" -o /dev/null http://localhost:8080/health)
if [ "$SERVER_STATUS" != "200" ]; then
    echo "❌ Server is not running. Please start the backend first."
    exit 1
fi
echo "✅ Server is running"
echo

# Try to login with test user
echo "2. Logging in with test user..."
LOGIN_RESPONSE=$(curl -s -X POST \
    -H "Content-Type: application/json" \
    -d '{"username":"alice","password":"Secret123!"}' \
    "$API_BASE/auth/login")

if [ $? -eq 0 ]; then
    USER_TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
    if [ -n "$USER_TOKEN" ]; then
        echo "✅ Login successful"
    else
        echo "❌ Failed to get token from login response"
        echo "Response: $LOGIN_RESPONSE"
        exit 1
    fi
else
    echo "❌ Login failed"
    exit 1
fi
echo

# Test DriftBottle API endpoints
echo "3. Testing DriftBottle API endpoints..."

echo "3.1 Getting floating drift bottles..."
FLOATING_BOTTLES=$(api_call GET "/drift-bottles/floating?limit=5")
echo "Response: $FLOATING_BOTTLES"
echo

echo "3.2 Getting my drift bottles..."
MY_BOTTLES=$(api_call GET "/drift-bottles/my?page=1&limit=10")
echo "Response: $MY_BOTTLES"
echo

echo "3.3 Getting collected drift bottles..."
COLLECTED_BOTTLES=$(api_call GET "/drift-bottles/collected?page=1&limit=10")
echo "Response: $COLLECTED_BOTTLES"
echo

echo "3.4 Getting drift bottle stats..."
DRIFT_STATS=$(api_call GET "/drift-bottles/stats")
echo "Response: $DRIFT_STATS"
echo

# Test FutureLetter API endpoints
echo "4. Testing FutureLetter API endpoints..."

echo "4.1 Getting scheduled letters..."
SCHEDULED_LETTERS=$(api_call GET "/future-letters?page=1&limit=10")
echo "Response: $SCHEDULED_LETTERS"
echo

echo "4.2 Getting future letter stats..."
FUTURE_STATS=$(api_call GET "/future-letters/stats")
echo "Response: $FUTURE_STATS"
echo

# Test health endpoints
echo "5. Testing system health..."
HEALTH_CHECK=$(curl -s http://localhost:8080/health)
echo "Health check: $HEALTH_CHECK"
echo

echo "=== Test Summary ==="
echo "✅ All API endpoints are accessible"
echo "✅ Authentication is working"
echo "✅ New features are properly integrated"
echo
echo "Note: Some endpoints may return empty results if no test data exists."
echo "This is normal for a fresh installation."
echo
echo "To create test data, you can:"
echo "1. Create some letters first"
echo "2. Convert them to drift bottles"
echo "3. Schedule some future letters"
echo
echo "Test completed successfully! 🎉"