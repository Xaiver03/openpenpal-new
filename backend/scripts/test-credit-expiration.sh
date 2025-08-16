#!/bin/bash

# =====================================================
# Phase 4.1: Credit Expiration System Test Script
# =====================================================
# Description: Comprehensive testing for credit expiration functionality
# Author: Claude Code Assistant  
# Date: 2025-08-15

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BASE_URL="http://localhost:8080/api/v1"
ADMIN_URL="http://localhost:8080/api/v1/admin"

# Test credentials
ADMIN_USERNAME="admin"
ADMIN_PASSWORD="admin123"
TEST_USER_USERNAME="alice"
TEST_USER_PASSWORD="secret123"

# Global variables for tokens
ADMIN_TOKEN=""
USER_TOKEN=""

echo -e "${BLUE}=== Phase 4.1: Credit Expiration System Test Script ===${NC}"
echo "Testing comprehensive credit expiration functionality..."

# Function to make HTTP requests with error handling
make_request() {
    local method=$1
    local url=$2
    local data=$3
    local token=$4
    local description=$5

    echo -e "\n${YELLOW}Testing: $description${NC}"
    echo "Request: $method $url"
    
    if [ -n "$data" ]; then
        echo "Data: $data"
    fi

    local response
    local http_code
    local auth_header=""
    
    if [ -n "$token" ]; then
        auth_header="-H \"Authorization: Bearer $token\""
    fi

    if [ "$method" = "GET" ]; then
        response=$(curl -s -w "\n%{http_code}" $auth_header "$url" 2>/dev/null || echo -e "\n000")
    elif [ "$method" = "POST" ]; then
        response=$(curl -s -w "\n%{http_code}" -X POST $auth_header \
            -H "Content-Type: application/json" \
            -d "$data" "$url" 2>/dev/null || echo -e "\n000")
    elif [ "$method" = "PUT" ]; then
        response=$(curl -s -w "\n%{http_code}" -X PUT $auth_header \
            -H "Content-Type: application/json" \
            -d "$data" "$url" 2>/dev/null || echo -e "\n000")
    elif [ "$method" = "DELETE" ]; then
        response=$(curl -s -w "\n%{http_code}" -X DELETE $auth_header "$url" 2>/dev/null || echo -e "\n000")
    fi

    http_code=$(echo "$response" | tail -n 1)
    response_body=$(echo "$response" | sed '$d')

    if [ "$http_code" -ge 200 ] && [ "$http_code" -lt 300 ]; then
        echo -e "${GREEN}✓ Success ($http_code)${NC}"
        echo "$response_body" | jq '.' 2>/dev/null || echo "$response_body"
        return 0
    else
        echo -e "${RED}✗ Failed ($http_code)${NC}"
        echo "$response_body" | jq '.' 2>/dev/null || echo "$response_body"
        return 1
    fi
}

# Function to login and get token
login() {
    local username=$1
    local password=$2
    local role=$3

    echo -e "\n${BLUE}=== Logging in as $role ($username) ===${NC}"
    
    local response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "{\"username\":\"$username\",\"password\":\"$password\"}" \
        "$BASE_URL/auth/login" 2>/dev/null)
    
    local token=$(echo "$response" | jq -r '.token // .data.token // .access_token // empty' 2>/dev/null)
    
    if [ -n "$token" ] && [ "$token" != "null" ]; then
        echo -e "${GREEN}✓ Login successful${NC}"
        echo "$token"
    else
        echo -e "${RED}✗ Login failed${NC}"
        echo "Response: $response"
        exit 1
    fi
}

# Function to test basic credit operations
test_basic_credit_operations() {
    echo -e "\n${BLUE}=== Testing Basic Credit Operations ===${NC}"
    
    # Get user credit info
    make_request "GET" "$BASE_URL/credits/me" "" "$USER_TOKEN" "Get user credit info"
    
    # Get credit history
    make_request "GET" "$BASE_URL/credits/me/history" "" "$USER_TOKEN" "Get credit history"
    
    # Add some test credits for expiration testing
    make_request "POST" "$ADMIN_URL/credits/users/add-points" \
        '{"user_id":"user-alice","points":100,"description":"测试信件创建积分","reference":"test-letter-001"}' \
        "$ADMIN_TOKEN" "Add test letter creation credits"
        
    make_request "POST" "$ADMIN_URL/credits/users/add-points" \
        '{"user_id":"user-alice","points":50,"description":"公开信被点赞","reference":"test-like-001"}' \
        "$ADMIN_TOKEN" "Add test social interaction credits"
}

# Function to test expiration rules management
test_expiration_rules() {
    echo -e "\n${BLUE}=== Testing Expiration Rules Management ===${NC}"
    
    # Get all expiration rules
    make_request "GET" "$ADMIN_URL/credits/expiration/rules" "" "$ADMIN_TOKEN" "Get all expiration rules"
    
    # Create a new expiration rule for testing
    local test_rule_data='{
        "rule_name": "测试积分规则",
        "credit_type": "test_credit",
        "expiration_days": 30,
        "notify_days": 3,
        "is_active": true,
        "priority": 50,
        "description": "用于测试的短期过期规则"
    }'
    
    local rule_response=$(make_request "POST" "$ADMIN_URL/credits/expiration/rules" "$test_rule_data" "$ADMIN_TOKEN" "Create test expiration rule")
    local rule_id=$(echo "$rule_response" | jq -r '.id // .data.id // empty' 2>/dev/null)
    
    if [ -n "$rule_id" ] && [ "$rule_id" != "null" ]; then
        echo "Created test rule ID: $rule_id"
        
        # Update the rule
        local update_data='{"notify_days": 5, "description": "更新后的测试规则"}'
        make_request "PUT" "$ADMIN_URL/credits/expiration/rules/$rule_id" "$update_data" "$ADMIN_TOKEN" "Update expiration rule"
        
        # Delete the test rule
        make_request "DELETE" "$ADMIN_URL/credits/expiration/rules/$rule_id" "" "$ADMIN_TOKEN" "Delete test expiration rule"
    fi
}

# Function to test user expiration queries
test_user_expiration_queries() {
    echo -e "\n${BLUE}=== Testing User Expiration Queries ===${NC}"
    
    # Get user's expiring credits (30 days)
    make_request "GET" "$BASE_URL/credits/expiring?days=30" "" "$USER_TOKEN" "Get expiring credits in 30 days"
    
    # Get user's expiring credits (365 days) 
    make_request "GET" "$BASE_URL/credits/expiring?days=365" "" "$USER_TOKEN" "Get expiring credits in 365 days"
    
    # Get user's expiration history
    make_request "GET" "$BASE_URL/credits/expiration-history?page=1&limit=10" "" "$USER_TOKEN" "Get user expiration history"
}

# Function to test admin expiration management
test_admin_expiration_management() {
    echo -e "\n${BLUE}=== Testing Admin Expiration Management ===${NC}"
    
    # Get expiration statistics
    make_request "GET" "$ADMIN_URL/credits/expiration/statistics" "" "$ADMIN_TOKEN" "Get expiration statistics"
    
    # Send expiration warnings manually
    make_request "POST" "$ADMIN_URL/credits/expiration/warnings" "" "$ADMIN_TOKEN" "Send expiration warnings"
    
    # Process expired credits manually (this is safe in test environment)
    make_request "POST" "$ADMIN_URL/credits/expiration/process" "" "$ADMIN_TOKEN" "Process expired credits"
    
    # Get expiration batches
    make_request "GET" "$ADMIN_URL/credits/expiration/batches?page=1&limit=5" "" "$ADMIN_TOKEN" "Get expiration batches"
    
    # Get expiration logs
    make_request "GET" "$ADMIN_URL/credits/expiration/logs?page=1&limit=10" "" "$ADMIN_TOKEN" "Get expiration logs"
    
    # Get expiration notifications
    make_request "GET" "$ADMIN_URL/credits/expiration/notifications?page=1&limit=10" "" "$ADMIN_TOKEN" "Get expiration notifications"
}

# Function to test edge cases and error handling
test_edge_cases() {
    echo -e "\n${BLUE}=== Testing Edge Cases and Error Handling ===${NC}"
    
    # Test invalid days parameter
    make_request "GET" "$BASE_URL/credits/expiring?days=-1" "" "$USER_TOKEN" "Test invalid negative days"
    
    # Test very large days parameter
    make_request "GET" "$BASE_URL/credits/expiring?days=999999" "" "$USER_TOKEN" "Test very large days parameter"
    
    # Test invalid rule creation
    local invalid_rule='{
        "rule_name": "",
        "credit_type": "",
        "expiration_days": -1
    }'
    make_request "POST" "$ADMIN_URL/credits/expiration/rules" "$invalid_rule" "$ADMIN_TOKEN" "Test invalid rule creation"
    
    # Test non-existent rule update
    make_request "PUT" "$ADMIN_URL/credits/expiration/rules/non-existent-id" '{"notify_days": 10}' "$ADMIN_TOKEN" "Test non-existent rule update"
    
    # Test non-existent rule deletion
    make_request "DELETE" "$ADMIN_URL/credits/expiration/rules/non-existent-id" "" "$ADMIN_TOKEN" "Test non-existent rule deletion"
}

# Function to test database migration
test_database_migration() {
    echo -e "\n${BLUE}=== Testing Database Migration Verification ===${NC}"
    
    # This test verifies that the database migration was applied correctly
    # by checking if we can query the new tables and fields
    
    # Check if expiration rules exist (should return the default rules)
    local rules_response=$(make_request "GET" "$ADMIN_URL/credits/expiration/rules" "" "$ADMIN_TOKEN" "Verify expiration rules table")
    
    if echo "$rules_response" | jq -e '.data | length > 0' >/dev/null 2>&1 || \
       echo "$rules_response" | jq -e '. | length > 0' >/dev/null 2>&1; then
        echo -e "${GREEN}✓ Database migration appears successful - expiration rules found${NC}"
    else
        echo -e "${RED}✗ Database migration may have failed - no expiration rules found${NC}"
    fi
}

# Function to run credit expiration system test
test_credit_expiration_flow() {
    echo -e "\n${BLUE}=== Testing Credit Expiration Complete Flow ===${NC}"
    
    # Step 1: Add credits with different types
    echo "Adding test credits for expiration testing..."
    
    make_request "POST" "$ADMIN_URL/credits/users/add-points" \
        '{"user_id":"user-alice","points":200,"description":"参与写作挑战并完成投稿","reference":"test-writing-challenge"}' \
        "$ADMIN_TOKEN" "Add writing challenge credits"
        
    make_request "POST" "$ADMIN_URL/credits/users/add-points" \
        '{"user_id":"user-alice","points":30,"description":"使用AI笔友并留下评价","reference":"test-ai-interaction"}' \
        "$ADMIN_TOKEN" "Add AI interaction credits"
    
    # Step 2: Check if expiration was applied
    echo "Checking if expiration dates were automatically applied..."
    make_request "GET" "$BASE_URL/credits/me/history?limit=5" "" "$USER_TOKEN" "Check recent credit history for expiration dates"
    
    # Step 3: Test expiration queries
    echo "Testing expiration queries..."
    make_request "GET" "$BASE_URL/credits/expiring?days=180" "" "$USER_TOKEN" "Check credits expiring in 6 months"
    make_request "GET" "$BASE_URL/credits/expiring?days=60" "" "$USER_TOKEN" "Check credits expiring in 2 months"
}

# Main execution
main() {
    echo -e "${BLUE}Starting Credit Expiration System Tests...${NC}"
    
    # Check if server is running
    if ! curl -s "$BASE_URL/health" >/dev/null 2>&1; then
        echo -e "${RED}✗ Server is not running at $BASE_URL${NC}"
        echo "Please start the server first: go run main.go"
        exit 1
    fi
    
    # Login
    ADMIN_TOKEN=$(login "$ADMIN_USERNAME" "$ADMIN_PASSWORD" "admin")
    USER_TOKEN=$(login "$TEST_USER_USERNAME" "$TEST_USER_PASSWORD" "user")
    
    # Run tests
    test_database_migration
    test_basic_credit_operations
    test_expiration_rules
    test_user_expiration_queries  
    test_admin_expiration_management
    test_credit_expiration_flow
    test_edge_cases
    
    echo -e "\n${GREEN}=== All Credit Expiration System Tests Completed ===${NC}"
    echo -e "${BLUE}Test Summary:${NC}"
    echo "• Database migration verification: Completed"
    echo "• Basic credit operations: Completed"
    echo "• Expiration rules management: Completed"
    echo "• User expiration queries: Completed"
    echo "• Admin expiration management: Completed"
    echo "• Complete expiration flow: Completed"
    echo "• Edge cases and error handling: Completed"
    
    echo -e "\n${YELLOW}Next Steps:${NC}"
    echo "1. Review server logs for any expiration processing messages"
    echo "2. Check database tables for expiration data:"
    echo "   - credit_expiration_rules"
    echo "   - credit_expiration_batches"  
    echo "   - credit_expiration_logs"
    echo "   - credit_expiration_notifications"
    echo "3. Set up automated expiration processing (cron job or scheduler)"
    echo "4. Monitor expiration statistics in admin dashboard"
}

# Run the script
main "$@"