#!/bin/bash

# Test script for Level 4 courier OP Code permissions
# This script tests that Level 4 couriers have full OP Code management capabilities

echo "Testing Level 4 Courier OP Code Permissions"
echo "==========================================="

# Configuration
API_BASE="${API_BASE:-http://localhost:8080}"
L4_USERNAME="${L4_USERNAME:-courier_level4}"
L4_PASSWORD="${L4_PASSWORD:-Secret123!}"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
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

# Step 1: Login as Level 4 courier
echo -e "\n${YELLOW}Step 1: Login as Level 4 courier${NC}"
LOGIN_RESPONSE=$(curl -s -X POST "$API_BASE/api/auth/login" \
    -H "Content-Type: application/json" \
    -d "{\"username\":\"$L4_USERNAME\",\"password\":\"$L4_PASSWORD\"}")

TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"token":"[^"]*' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
    echo -e "${RED}Failed to login as Level 4 courier${NC}"
    echo "Response: $LOGIN_RESPONSE"
    exit 1
fi

echo -e "${GREEN}✓ Successfully logged in as Level 4 courier${NC}"

# Step 2: Get managed OP codes (should see all codes or empty list)
echo -e "\n${YELLOW}Step 2: Testing GetManagedOPCodes endpoint${NC}"
MANAGED_RESPONSE=$(curl -s -X GET "$API_BASE/api/v1/courier/opcode/managed" \
    -H "Authorization: Bearer $TOKEN")

echo "Response: $MANAGED_RESPONSE"
SUCCESS=$(echo "$MANAGED_RESPONSE" | grep -o '"success":true')
if [ -n "$SUCCESS" ]; then
    print_result 0 "GetManagedOPCodes successful"
else
    print_result 1 "GetManagedOPCodes failed" "$MANAGED_RESPONSE"
fi

# Step 3: Get pending applications (should see all applications)
echo -e "\n${YELLOW}Step 3: Testing GetApplications endpoint${NC}"
APPS_RESPONSE=$(curl -s -X GET "$API_BASE/api/v1/courier/opcode/applications" \
    -H "Authorization: Bearer $TOKEN")

echo "Response: $APPS_RESPONSE"
SUCCESS=$(echo "$APPS_RESPONSE" | grep -o '"success":true')
if [ -n "$SUCCESS" ]; then
    print_result 0 "GetApplications successful"
else
    print_result 1 "GetApplications failed" "$APPS_RESPONSE"
fi

# Step 4: Create a new OP Code (Level 4 should be able to create any code)
echo -e "\n${YELLOW}Step 4: Testing CreateOPCode endpoint${NC}"
CREATE_RESPONSE=$(curl -s -X POST "$API_BASE/api/v1/courier/opcode/create" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
        "code": "QH3A01",
        "type": "dormitory",
        "name": "清华大学3号楼A区01室",
        "description": "Test OP Code created by L4 courier",
        "is_public": true
    }')

echo "Response: $CREATE_RESPONSE"
SUCCESS=$(echo "$CREATE_RESPONSE" | grep -o '"success":true')
if [ -n "$SUCCESS" ]; then
    print_result 0 "CreateOPCode successful"
    CREATED_ID=$(echo "$CREATE_RESPONSE" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
else
    print_result 1 "CreateOPCode failed" "$CREATE_RESPONSE"
fi

# Step 5: Update an OP Code (if creation was successful)
if [ -n "$CREATED_ID" ]; then
    echo -e "\n${YELLOW}Step 5: Testing UpdateOPCode endpoint${NC}"
    UPDATE_RESPONSE=$(curl -s -X PUT "$API_BASE/api/v1/courier/opcode/$CREATED_ID" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
            "name": "清华大学3号楼A区01室 (Updated)",
            "description": "Updated by L4 courier test",
            "is_public": false,
            "is_active": true
        }')

    echo "Response: $UPDATE_RESPONSE"
    SUCCESS=$(echo "$UPDATE_RESPONSE" | grep -o '"success":true')
    if [ -n "$SUCCESS" ]; then
        print_result 0 "UpdateOPCode successful"
    else
        print_result 1 "UpdateOPCode failed" "$UPDATE_RESPONSE"
    fi
fi

# Step 6: Test creating codes for different schools (L4 should manage all)
echo -e "\n${YELLOW}Step 6: Testing cross-school OP Code creation${NC}"
SCHOOLS=("PK" "QH" "BD")
for school in "${SCHOOLS[@]}"; do
    CODE="${school}9Z99"
    CREATE_RESPONSE=$(curl -s -X POST "$API_BASE/api/v1/courier/opcode/create" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "{
            \"code\": \"$CODE\",
            \"type\": \"box\",
            \"name\": \"Test Box for $school\",
            \"description\": \"Cross-school test by L4\",
            \"is_public\": true
        }")

    SUCCESS=$(echo "$CREATE_RESPONSE" | grep -o '"success":true')
    if [ -n "$SUCCESS" ]; then
        print_result 0 "Created OP Code for school $school"
    else
        print_result 1 "Failed to create OP Code for school $school" "$CREATE_RESPONSE"
    fi
done

# Summary
echo -e "\n${YELLOW}Test Summary:${NC}"
echo "Level 4 courier should have full OP Code management permissions."
echo "If any tests failed, check the backend logs for details."
echo ""
echo "Common issues:"
echo "1. ManagedOPCodePrefix not set correctly - run fix-courier-managed-prefix.sh"
echo "2. Permission logic errors - check courier_opcode_handler.go"
echo "3. Database issues - check if OP Code tables exist"