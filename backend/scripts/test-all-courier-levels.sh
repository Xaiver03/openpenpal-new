#!/bin/bash

# Comprehensive test script for all courier levels
echo "Testing All Courier Levels - OP Code Management"
echo "==============================================="

# Configuration
API_BASE="${API_BASE:-http://localhost:8080}"

# Test accounts and passwords
declare -A COURIERS=(
    ["courier_level1"]="Secret123!"
    ["courier_level2"]="Secret123!"
    ["courier_level3"]="Secret123!"  
    ["courier_level4"]="Secret123!"
)

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print test results
print_result() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✓ $2${NC}"
    else
        echo -e "${RED}✗ $2${NC}"
        echo "Response: $3"
    fi
}

print_section() {
    echo -e "\n${BLUE}=== $1 ===${NC}"
}

# Function to test each courier level
test_courier_level() {
    local username=$1
    local password=$2
    local expected_level=$3
    
    print_section "Testing $username (Level $expected_level)"
    
    # Step 1: Login
    echo -e "\n${YELLOW}Step 1: Login as $username${NC}"
    LOGIN_RESPONSE=$(curl -s -X POST "$API_BASE/api/auth/login" \
        -H "Content-Type: application/json" \
        -d "{\"username\":\"$username\",\"password\":\"$password\"}")
    
    TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.data.token // empty' 2>/dev/null)
    
    if [ -z "$TOKEN" ]; then
        echo -e "${RED}Failed to login as $username${NC}"
        echo "Response: $LOGIN_RESPONSE"
        return 1
    fi
    
    echo -e "${GREEN}✓ Successfully logged in as $username${NC}"
    
    # Step 2: Get user profile to check courier info
    echo -e "\n${YELLOW}Step 2: Getting user profile${NC}"
    PROFILE_RESPONSE=$(curl -s -X GET "$API_BASE/api/auth/me" \
        -H "Authorization: Bearer $TOKEN")
    
    echo "Profile Response: $PROFILE_RESPONSE"
    
    # Step 3: Test OP Code managed list
    echo -e "\n${YELLOW}Step 3: Testing GetManagedOPCodes${NC}"
    MANAGED_RESPONSE=$(curl -s -X GET "$API_BASE/api/v1/courier/opcode/managed" \
        -H "Authorization: Bearer $TOKEN")
    
    echo "Managed OP Codes Response: $MANAGED_RESPONSE"
    
    # Step 4: Test OP Code applications list
    echo -e "\n${YELLOW}Step 4: Testing GetApplications${NC}"
    APPS_RESPONSE=$(curl -s -X GET "$API_BASE/api/v1/courier/opcode/applications" \
        -H "Authorization: Bearer $TOKEN")
    
    echo "Applications Response: $APPS_RESPONSE"
    
    # Step 5: Test create permission (L2+ should be able to create)
    if [ $expected_level -ge 2 ]; then
        echo -e "\n${YELLOW}Step 5: Testing CreateOPCode (Level $expected_level should have permission)${NC}"
        TEST_CODE=""
        case $expected_level in
            2) TEST_CODE="PK5F99" ;;  # L2 can create within their area
            3) TEST_CODE="PK9Z99" ;;  # L3 can create within their school
            4) TEST_CODE="QH9Z99" ;;  # L4 can create any code
        esac
        
        CREATE_RESPONSE=$(curl -s -X POST "$API_BASE/api/v1/courier/opcode/create" \
            -H "Authorization: Bearer $TOKEN" \
            -H "Content-Type: application/json" \
            -d "{
                \"code\": \"$TEST_CODE\",
                \"type\": \"box\",
                \"name\": \"Test Box L$expected_level\",
                \"description\": \"Test creation by L$expected_level courier\",
                \"is_public\": true
            }")
        
        echo "Create Response: $CREATE_RESPONSE"
        SUCCESS=$(echo "$CREATE_RESPONSE" | jq -r '.success // false' 2>/dev/null)
        if [ "$SUCCESS" = "true" ]; then
            print_result 0 "L$expected_level courier can create OP codes"
        else
            print_result 1 "L$expected_level courier failed to create OP codes" "$CREATE_RESPONSE"
        fi
    else
        echo -e "\n${YELLOW}Step 5: L1 couriers should not have create permission${NC}"
        CREATE_RESPONSE=$(curl -s -X POST "$API_BASE/api/v1/courier/opcode/create" \
            -H "Authorization: Bearer $TOKEN" \
            -H "Content-Type: application/json" \
            -d '{
                "code": "PK5F98",
                "type": "box",
                "name": "Test Box L1",
                "description": "Test creation by L1 courier",
                "is_public": true
            }')
        
        echo "Create Response: $CREATE_RESPONSE"
        SUCCESS=$(echo "$CREATE_RESPONSE" | jq -r '.success // false' 2>/dev/null)
        if [ "$SUCCESS" = "false" ]; then
            print_result 0 "L1 courier correctly denied create permission"
        else
            print_result 1 "L1 courier should not have create permission" "$CREATE_RESPONSE"
        fi
    fi
    
    echo -e "\n${BLUE}--- End of $username test ---${NC}\n"
}

# Main test execution
echo "Starting comprehensive courier level testing..."
echo "Testing against: $API_BASE"
echo ""

# Test each courier level
test_courier_level "courier_level1" "${COURIERS[courier_level1]}" 1
test_courier_level "courier_level2" "${COURIERS[courier_level2]}" 2  
test_courier_level "courier_level3" "${COURIERS[courier_level3]}" 3
test_courier_level "courier_level4" "${COURIERS[courier_level4]}" 4

# Summary
print_section "Test Summary"
echo "Expected behavior:"
echo "- L1: Can view tasks, no create permission"
echo "- L2: Can view and create OP codes in their area (PK5F**)"
echo "- L3: Can view and create OP codes in their school (PK****)"  
echo "- L4: Can view and create any OP codes globally"
echo ""
echo "If any tests failed, check:"
echo "1. Backend service is running on $API_BASE"
echo "2. Database has correct courier levels and prefixes"
echo "3. OP Code permission logic in courier_opcode_handler.go"