#!/bin/bash

# OpenPenPal Complete Flow Test
# Tests authentication and main user flows

BASE_URL="http://127.0.0.1:8080"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
LOG_FILE="test-complete-flow-$TIMESTAMP.log"

# Color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Disable proxy
export http_proxy=""
export https_proxy=""

echo -e "${BLUE}OpenPenPal Complete Flow Test - $TIMESTAMP${NC}" | tee $LOG_FILE
echo "===========================================" | tee -a $LOG_FILE

# Helper function for API calls
api_call() {
    local method=$1
    local endpoint=$2
    local data=$3
    local token=$4
    
    local auth_header=""
    if [[ -n "$token" ]]; then
        auth_header="-H \"Authorization: Bearer $token\""
    fi
    
    if [[ "$method" == "GET" ]]; then
        eval "curl -s -X GET \"$BASE_URL$endpoint\" $auth_header --noproxy '*'"
    else
        eval "curl -s -X $method \"$BASE_URL$endpoint\" \
            -H \"Content-Type: application/json\" \
            $auth_header \
            -d '$data' \
            --noproxy '*'"
    fi
}

# Test 1: Admin Login and Profile
echo -e "\n${BLUE}=== Test 1: Admin Authentication ===${NC}" | tee -a $LOG_FILE

echo -e "${YELLOW}1.1 Admin Login${NC}" | tee -a $LOG_FILE
admin_response=$(api_call POST "/api/v1/auth/login" '{"username":"admin","password":"admin123"}')
echo "Response: $admin_response" | tee -a $LOG_FILE

admin_token=$(echo "$admin_response" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
if [[ -n "$admin_token" ]]; then
    echo -e "${GREEN}✓ Admin login successful${NC}" | tee -a $LOG_FILE
    
    echo -e "\n${YELLOW}1.2 Get Admin Profile${NC}" | tee -a $LOG_FILE
    profile_response=$(api_call GET "/api/v1/users/me" "" "$admin_token")
    echo "Profile: $profile_response" | tee -a $LOG_FILE
    
    echo -e "\n${YELLOW}1.3 Get Dashboard Stats${NC}" | tee -a $LOG_FILE
    stats_response=$(api_call GET "/api/v1/admin/dashboard/stats" "" "$admin_token")
    echo "Stats: $stats_response" | tee -a $LOG_FILE
else
    echo -e "${RED}✗ Admin login failed${NC}" | tee -a $LOG_FILE
fi

# Test 2: Regular User Registration
echo -e "\n${BLUE}=== Test 2: User Registration ===${NC}" | tee -a $LOG_FILE

test_user="testuser_$(date +%s)"
echo -e "${YELLOW}2.1 Register New User: $test_user${NC}" | tee -a $LOG_FILE
register_data="{
    \"username\":\"$test_user\",
    \"password\":\"test123456\",
    \"email\":\"$test_user@test.com\",
    \"school_code\":\"BJDX\"
}"
register_response=$(api_call POST "/api/v1/auth/register" "$register_data")
echo "Register Response: $register_response" | tee -a $LOG_FILE

# Test 3: Letter Creation Flow
echo -e "\n${BLUE}=== Test 3: Letter Creation Flow ===${NC}" | tee -a $LOG_FILE

if [[ -n "$admin_token" ]]; then
    echo -e "${YELLOW}3.1 Create Draft Letter${NC}" | tee -a $LOG_FILE
    letter_data='{
        "title":"Integration Test Letter",
        "content":"This is a test letter created during integration testing.",
        "style":"classic"
    }'
    create_response=$(api_call POST "/api/v1/letters" "$letter_data" "$admin_token")
    echo "Create Response: $create_response" | tee -a $LOG_FILE
    
    letter_id=$(echo "$create_response" | grep -o '"id":"[^"]*' | head -1 | cut -d'"' -f4)
    if [[ -n "$letter_id" ]]; then
        echo -e "${GREEN}✓ Letter created with ID: $letter_id${NC}" | tee -a $LOG_FILE
        
        echo -e "\n${YELLOW}3.2 Generate Letter Code${NC}" | tee -a $LOG_FILE
        generate_response=$(api_call POST "/api/v1/letters/$letter_id/generate-code" "" "$admin_token")
        echo "Generate Response: $generate_response" | tee -a $LOG_FILE
        
        echo -e "\n${YELLOW}3.3 Get Letter List${NC}" | tee -a $LOG_FILE
        list_response=$(api_call GET "/api/v1/letters" "" "$admin_token")
        echo "List Response: $list_response" | tee -a $LOG_FILE | head -5
    fi
fi

# Test 4: Check Database State
echo -e "\n${BLUE}=== Test 4: Database Verification ===${NC}" | tee -a $LOG_FILE
echo -e "${YELLOW}Checking database state...${NC}" | tee -a $LOG_FILE

# Create SQL script to check data
cat > check-db-state.sql << EOF
-- Check users
SELECT COUNT(*) as user_count FROM users;
SELECT username, role, is_active FROM users ORDER BY created_at DESC LIMIT 5;

-- Check letters
SELECT COUNT(*) as letter_count FROM letters;
SELECT id, title, status, created_at FROM letters ORDER BY created_at DESC LIMIT 5;

-- Check letter codes
SELECT COUNT(*) as code_count FROM letter_codes;
EOF

echo "SQL verification script created: check-db-state.sql" | tee -a $LOG_FILE

# Test 5: Frontend Accessibility
echo -e "\n${BLUE}=== Test 5: Frontend Check ===${NC}" | tee -a $LOG_FILE
frontend_status=$(curl -s -o /dev/null -w "%{http_code}" http://127.0.0.1:3000 --noproxy '*')
if [[ $frontend_status -eq 200 ]]; then
    echo -e "${GREEN}✓ Frontend is accessible (HTTP $frontend_status)${NC}" | tee -a $LOG_FILE
else
    echo -e "${RED}✗ Frontend not accessible (HTTP $frontend_status)${NC}" | tee -a $LOG_FILE
fi

# Summary
echo -e "\n${BLUE}=== Test Summary ===${NC}" | tee -a $LOG_FILE
echo "Test completed at $(date)" | tee -a $LOG_FILE
echo "Results saved to: $LOG_FILE" | tee -a $LOG_FILE

# Final recommendations
echo -e "\n${YELLOW}=== Recommendations ===${NC}" | tee -a $LOG_FILE
echo "1. Update test account passwords in the database or documentation" | tee -a $LOG_FILE
echo "2. Add /api/v1/auth/me endpoint for consistency" | tee -a $LOG_FILE
echo "3. Improve error messages for authentication failures" | tee -a $LOG_FILE
echo "4. Add rate limiting for login attempts" | tee -a $LOG_FILE

echo -e "\n${GREEN}Complete flow test finished!${NC}" | tee -a $LOG_FILE