#!/bin/bash

# =====================================================
# Phase 4.2: Credit Transfer System Test Script
# =====================================================
# Description: Comprehensive testing for credit transfer functionality
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
ALICE_USERNAME="alice"
ALICE_PASSWORD="secret123"
BOB_USERNAME="bob"
BOB_PASSWORD="secret123"

# Global variables for tokens and user IDs
ADMIN_TOKEN=""
ALICE_TOKEN=""
BOB_TOKEN=""
ALICE_USER_ID=""
BOB_USER_ID=""

echo -e "${BLUE}=== Phase 4.2: Credit Transfer System Test Script ===${NC}"
echo "Testing comprehensive credit transfer functionality..."

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
        echo "$response_body"
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
    local user_id=$(echo "$response" | jq -r '.user.id // .data.user.id // empty' 2>/dev/null)
    
    if [ -n "$token" ] && [ "$token" != "null" ]; then
        echo -e "${GREEN}✓ Login successful${NC}"
        echo "$token|$user_id"
    else
        echo -e "${RED}✗ Login failed${NC}"
        echo "Response: $response"
        exit 1
    fi
}

# Function to ensure test users have credits
ensure_user_credits() {
    echo -e "\n${BLUE}=== Ensuring test users have sufficient credits ===${NC}"
    
    # Give Alice 1000 credits for testing
    make_request "POST" "$ADMIN_URL/credits/users/add-points" \
        "{\"user_id\":\"$ALICE_USER_ID\",\"points\":1000,\"description\":\"测试转赠用积分\",\"reference\":\"test-transfer\"}" \
        "$ADMIN_TOKEN" "Add test credits to Alice"
        
    # Give Bob 500 credits for testing
    make_request "POST" "$ADMIN_URL/credits/users/add-points" \
        "{\"user_id\":\"$BOB_USER_ID\",\"points\":500,\"description\":\"测试转赠用积分\",\"reference\":\"test-transfer\"}" \
        "$ADMIN_TOKEN" "Add test credits to Bob"
}

# Function to test basic transfer operations
test_basic_transfer() {
    echo -e "\n${BLUE}=== Testing Basic Transfer Operations ===${NC}"
    
    # Alice checks her credits before transfer
    echo -e "\n${YELLOW}Alice checks her credits before transfer${NC}"
    make_request "GET" "$BASE_URL/credits/me" "" "$ALICE_TOKEN" "Alice checks her credits"
    
    # Alice creates a transfer to Bob
    local transfer_data='{
        "to_user_id": "'$BOB_USER_ID'",
        "amount": 100,
        "transfer_type": "direct",
        "message": "测试转赠100积分给Bob",
        "reference": "test-transfer-001"
    }'
    
    local transfer_response=$(make_request "POST" "$BASE_URL/credits/transfer" "$transfer_data" "$ALICE_TOKEN" "Alice creates transfer to Bob")
    local transfer_id=$(echo "$transfer_response" | jq -r '.data.transfer.id // empty' 2>/dev/null)
    
    if [ -n "$transfer_id" ] && [ "$transfer_id" != "null" ]; then
        echo "Created transfer ID: $transfer_id"
        
        # Alice checks her transfer list
        make_request "GET" "$BASE_URL/credits/transfers" "" "$ALICE_TOKEN" "Alice checks her transfers"
        
        # Bob checks his pending transfers
        make_request "GET" "$BASE_URL/credits/transfers/pending" "" "$BOB_TOKEN" "Bob checks pending transfers"
        
        # Bob accepts the transfer
        make_request "POST" "$BASE_URL/credits/transfers/$transfer_id/process" \
            '{"action":"accept"}' \
            "$BOB_TOKEN" "Bob accepts the transfer"
            
        # Both users check their credits after transfer
        make_request "GET" "$BASE_URL/credits/me" "" "$ALICE_TOKEN" "Alice checks credits after transfer"
        make_request "GET" "$BASE_URL/credits/me" "" "$BOB_TOKEN" "Bob checks credits after transfer"
    fi
}

# Function to test transfer rejection
test_transfer_rejection() {
    echo -e "\n${BLUE}=== Testing Transfer Rejection ===${NC}"
    
    # Alice creates another transfer to Bob
    local transfer_data='{
        "to_user_id": "'$BOB_USER_ID'",
        "amount": 50,
        "transfer_type": "gift",
        "message": "这是一个会被拒绝的转赠",
        "reference": "test-reject-001"
    }'
    
    local transfer_response=$(make_request "POST" "$BASE_URL/credits/transfer" "$transfer_data" "$ALICE_TOKEN" "Alice creates gift transfer to Bob")
    local transfer_id=$(echo "$transfer_response" | jq -r '.data.transfer.id // empty' 2>/dev/null)
    
    if [ -n "$transfer_id" ] && [ "$transfer_id" != "null" ]; then
        # Bob rejects the transfer
        make_request "POST" "$BASE_URL/credits/transfers/$transfer_id/process" \
            '{"action":"reject","reason":"测试拒绝转赠"}' \
            "$BOB_TOKEN" "Bob rejects the transfer"
            
        # Alice checks her credits (should be refunded minus fee)
        make_request "GET" "$BASE_URL/credits/me" "" "$ALICE_TOKEN" "Alice checks credits after rejection"
    fi
}

# Function to test transfer cancellation
test_transfer_cancellation() {
    echo -e "\n${BLUE}=== Testing Transfer Cancellation ===${NC}"
    
    # Alice creates a transfer and then cancels it
    local transfer_data='{
        "to_user_id": "'$BOB_USER_ID'",
        "amount": 75,
        "transfer_type": "reward",
        "message": "这个转赠会被取消",
        "reference": "test-cancel-001"
    }'
    
    local transfer_response=$(make_request "POST" "$BASE_URL/credits/transfer" "$transfer_data" "$ALICE_TOKEN" "Alice creates reward transfer")
    local transfer_id=$(echo "$transfer_response" | jq -r '.data.transfer.id // empty' 2>/dev/null)
    
    if [ -n "$transfer_id" ] && [ "$transfer_id" != "null" ]; then
        # Alice cancels the transfer
        make_request "DELETE" "$BASE_URL/credits/transfers/$transfer_id" "" "$ALICE_TOKEN" "Alice cancels the transfer"
        
        # Alice checks her credits (should be fully refunded including fee)
        make_request "GET" "$BASE_URL/credits/me" "" "$ALICE_TOKEN" "Alice checks credits after cancellation"
    fi
}

# Function to test transfer validation
test_transfer_validation() {
    echo -e "\n${BLUE}=== Testing Transfer Validation ===${NC}"
    
    # Test validation before creating transfer
    local validation_data='{
        "to_user_id": "'$BOB_USER_ID'",
        "amount": 200,
        "transfer_type": "direct",
        "message": "验证转赠可行性"
    }'
    
    make_request "POST" "$BASE_URL/credits/transfer/validate" "$validation_data" "$ALICE_TOKEN" "Validate transfer feasibility"
    
    # Test invalid transfer (amount too large)
    local invalid_data='{
        "to_user_id": "'$BOB_USER_ID'",
        "amount": 10000,
        "transfer_type": "direct",
        "message": "测试超额转赠"
    }'
    
    make_request "POST" "$BASE_URL/credits/transfer" "$invalid_data" "$ALICE_TOKEN" "Test transfer with insufficient balance"
    
    # Test invalid transfer (to self)
    local self_transfer='{
        "to_user_id": "'$ALICE_USER_ID'",
        "amount": 10,
        "transfer_type": "direct",
        "message": "测试自转"
    }'
    
    make_request "POST" "$BASE_URL/credits/transfer" "$self_transfer" "$ALICE_TOKEN" "Test self-transfer (should fail)"
}

# Function to test batch transfer
test_batch_transfer() {
    echo -e "\n${BLUE}=== Testing Batch Transfer ===${NC}"
    
    # Create a third user for batch transfer test
    local batch_data='{
        "to_user_ids": ["'$BOB_USER_ID'", "'$ADMIN_USER_ID'"],
        "amount": 10,
        "transfer_type": "gift",
        "message": "批量转赠测试",
        "reference": "test-batch-001"
    }'
    
    make_request "POST" "$BASE_URL/credits/transfer/batch" "$batch_data" "$ALICE_TOKEN" "Alice performs batch transfer"
}

# Function to test transfer statistics
test_transfer_statistics() {
    echo -e "\n${BLUE}=== Testing Transfer Statistics ===${NC}"
    
    # Get user transfer statistics
    make_request "GET" "$BASE_URL/credits/transfers/stats" "" "$ALICE_TOKEN" "Get Alice's transfer statistics"
    make_request "GET" "$BASE_URL/credits/transfers/stats" "" "$BOB_TOKEN" "Get Bob's transfer statistics"
}

# Function to test admin transfer management
test_admin_transfer_management() {
    echo -e "\n${BLUE}=== Testing Admin Transfer Management ===${NC}"
    
    # Admin gets all transfers
    make_request "GET" "$ADMIN_URL/credits/transfers/all?page=1&limit=10" "" "$ADMIN_TOKEN" "Admin gets all transfers"
    
    # Admin gets transfer statistics
    make_request "GET" "$ADMIN_URL/credits/transfers/statistics" "" "$ADMIN_TOKEN" "Admin gets transfer statistics"
    
    # Admin processes expired transfers
    make_request "POST" "$ADMIN_URL/credits/transfers/process-expired" "" "$ADMIN_TOKEN" "Admin processes expired transfers"
}

# Function to test edge cases
test_edge_cases() {
    echo -e "\n${BLUE}=== Testing Edge Cases ===${NC}"
    
    # Test transfer with empty message
    local empty_message='{
        "to_user_id": "'$BOB_USER_ID'",
        "amount": 5,
        "transfer_type": "direct",
        "message": ""
    }'
    
    make_request "POST" "$BASE_URL/credits/transfer" "$empty_message" "$ALICE_TOKEN" "Test transfer with empty message"
    
    # Test transfer with very long message
    local long_message='{
        "to_user_id": "'$BOB_USER_ID'",
        "amount": 5,
        "transfer_type": "direct",
        "message": "'$(printf '%.0s测试' {1..200})'"
    }'
    
    make_request "POST" "$BASE_URL/credits/transfer" "$long_message" "$ALICE_TOKEN" "Test transfer with very long message"
    
    # Test invalid transfer type
    local invalid_type='{
        "to_user_id": "'$BOB_USER_ID'",
        "amount": 5,
        "transfer_type": "invalid_type",
        "message": "测试无效类型"
    }'
    
    make_request "POST" "$BASE_URL/credits/transfer" "$invalid_type" "$ALICE_TOKEN" "Test invalid transfer type"
    
    # Test zero amount transfer
    local zero_amount='{
        "to_user_id": "'$BOB_USER_ID'",
        "amount": 0,
        "transfer_type": "direct",
        "message": "测试零积分转赠"
    }'
    
    make_request "POST" "$BASE_URL/credits/transfer" "$zero_amount" "$ALICE_TOKEN" "Test zero amount transfer"
    
    # Test negative amount transfer
    local negative_amount='{
        "to_user_id": "'$BOB_USER_ID'",
        "amount": -10,
        "transfer_type": "direct",
        "message": "测试负数积分转赠"
    }'
    
    make_request "POST" "$BASE_URL/credits/transfer" "$negative_amount" "$ALICE_TOKEN" "Test negative amount transfer"
}

# Function to test transfer permissions
test_transfer_permissions() {
    echo -e "\n${BLUE}=== Testing Transfer Permissions ===${NC}"
    
    # Create a transfer from Alice to Bob
    local transfer_data='{
        "to_user_id": "'$BOB_USER_ID'",
        "amount": 20,
        "transfer_type": "direct",
        "message": "测试权限的转赠"
    }'
    
    local transfer_response=$(make_request "POST" "$BASE_URL/credits/transfer" "$transfer_data" "$ALICE_TOKEN" "Alice creates transfer for permission test")
    local transfer_id=$(echo "$transfer_response" | jq -r '.data.transfer.id // empty' 2>/dev/null)
    
    if [ -n "$transfer_id" ] && [ "$transfer_id" != "null" ]; then
        # Alice tries to process her own transfer (should fail)
        make_request "POST" "$BASE_URL/credits/transfers/$transfer_id/process" \
            '{"action":"accept"}' \
            "$ALICE_TOKEN" "Alice tries to accept her own transfer (should fail)"
            
        # Bob tries to cancel Alice's transfer (should fail)
        make_request "DELETE" "$BASE_URL/credits/transfers/$transfer_id" "" "$BOB_TOKEN" "Bob tries to cancel Alice's transfer (should fail)"
        
        # Admin cancels the transfer
        make_request "DELETE" "$ADMIN_URL/credits/transfers/$transfer_id/cancel" \
            '{"reason":"管理员测试取消"}' \
            "$ADMIN_TOKEN" "Admin cancels the transfer"
    fi
}

# Main execution
main() {
    echo -e "${BLUE}Starting Credit Transfer System Tests...${NC}"
    
    # Check if server is running
    if ! curl -s "$BASE_URL/health" >/dev/null 2>&1; then
        echo -e "${RED}✗ Server is not running at $BASE_URL${NC}"
        echo "Please start the server first: go run main.go"
        exit 1
    fi
    
    # Login and get tokens
    local admin_result=$(login "$ADMIN_USERNAME" "$ADMIN_PASSWORD" "admin")
    ADMIN_TOKEN=$(echo "$admin_result" | cut -d'|' -f1)
    ADMIN_USER_ID=$(echo "$admin_result" | cut -d'|' -f2)
    
    local alice_result=$(login "$ALICE_USERNAME" "$ALICE_PASSWORD" "alice")
    ALICE_TOKEN=$(echo "$alice_result" | cut -d'|' -f1)
    ALICE_USER_ID=$(echo "$alice_result" | cut -d'|' -f2)
    
    # Try to login as Bob, create account if needed
    local bob_result=$(login "$BOB_USERNAME" "$BOB_PASSWORD" "bob" 2>/dev/null || echo "")
    if [ -z "$bob_result" ]; then
        echo -e "${YELLOW}Creating Bob's account...${NC}"
        # Register Bob
        curl -s -X POST \
            -H "Content-Type: application/json" \
            -d '{"username":"bob","password":"secret123","email":"bob@example.com"}' \
            "$BASE_URL/auth/register" >/dev/null 2>&1
        
        bob_result=$(login "$BOB_USERNAME" "$BOB_PASSWORD" "bob")
    fi
    
    BOB_TOKEN=$(echo "$bob_result" | cut -d'|' -f1)
    BOB_USER_ID=$(echo "$bob_result" | cut -d'|' -f2)
    
    echo "User IDs - Admin: $ADMIN_USER_ID, Alice: $ALICE_USER_ID, Bob: $BOB_USER_ID"
    
    # Ensure users have credits
    ensure_user_credits
    
    # Run tests
    test_basic_transfer
    test_transfer_rejection
    test_transfer_cancellation
    test_transfer_validation
    test_batch_transfer
    test_transfer_statistics
    test_admin_transfer_management
    test_transfer_permissions
    test_edge_cases
    
    echo -e "\n${GREEN}=== All Credit Transfer System Tests Completed ===${NC}"
    echo -e "${BLUE}Test Summary:${NC}"
    echo "• Basic transfer operations: Completed"
    echo "• Transfer rejection flow: Completed"
    echo "• Transfer cancellation: Completed"
    echo "• Transfer validation: Completed"
    echo "• Batch transfer: Completed"
    echo "• Transfer statistics: Completed"
    echo "• Admin management: Completed"
    echo "• Permission checks: Completed"
    echo "• Edge cases: Completed"
    
    echo -e "\n${YELLOW}Next Steps:${NC}"
    echo "1. Review server logs for any transfer processing messages"
    echo "2. Check database tables for transfer data:"
    echo "   - credit_transfers"
    echo "   - credit_transfer_rules"
    echo "   - credit_transfer_limits"
    echo "   - credit_transfer_notifications"
    echo "3. Test transfer expiration by waiting or manually updating expires_at"
    echo "4. Monitor transfer statistics in admin dashboard"
    echo "5. Configure transfer rules for different user roles"
}

# Run the script
main "$@"