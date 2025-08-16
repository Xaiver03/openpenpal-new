#!/bin/bash

# Authentication Test Script for OpenPenPal
# This script tests all user accounts login functionality

BASE_URL="http://localhost:8080"
FRONTEND_URL="http://localhost:3000"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
LOG_FILE="test-auth-results-$TIMESTAMP.log"

# Color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "OpenPenPal Authentication Test - $TIMESTAMP" | tee $LOG_FILE
echo "============================================" | tee -a $LOG_FILE

# Function to test login
test_login() {
    local username=$1
    local password=$2
    local expected_role=$3
    
    echo -e "\n${YELLOW}Testing login for user: $username (Expected role: $expected_role)${NC}" | tee -a $LOG_FILE
    
    # Make login request
    response=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d "{\"username\":\"$username\",\"password\":\"$password\"}" \
        -w "\n%{http_code}")
    
    # Extract HTTP status code
    http_code=$(echo "$response" | tail -n 1)
    response_body=$(echo "$response" | head -n -1)
    
    echo "HTTP Status: $http_code" | tee -a $LOG_FILE
    echo "Response: $response_body" | tee -a $LOG_FILE
    
    # Check if login was successful
    if [[ $http_code -eq 200 ]]; then
        # Extract token from response
        token=$(echo "$response_body" | grep -o '"token":"[^"]*' | grep -o '[^"]*$')
        
        if [[ -n "$token" ]]; then
            echo -e "${GREEN}✓ Login successful! Token received${NC}" | tee -a $LOG_FILE
            
            # Verify user info
            user_info=$(curl -s -X GET "$BASE_URL/api/v1/auth/me" \
                -H "Authorization: Bearer $token")
            
            echo "User info: $user_info" | tee -a $LOG_FILE
            
            # Extract role from user info
            actual_role=$(echo "$user_info" | grep -o '"role":"[^"]*' | grep -o '[^"]*$')
            
            if [[ "$actual_role" == "$expected_role" ]]; then
                echo -e "${GREEN}✓ Role verification passed: $actual_role${NC}" | tee -a $LOG_FILE
                return 0
            else
                echo -e "${RED}✗ Role mismatch! Expected: $expected_role, Got: $actual_role${NC}" | tee -a $LOG_FILE
                return 1
            fi
        else
            echo -e "${RED}✗ Login failed: No token in response${NC}" | tee -a $LOG_FILE
            return 1
        fi
    else
        echo -e "${RED}✗ Login failed with HTTP status: $http_code${NC}" | tee -a $LOG_FILE
        return 1
    fi
}

# Function to test frontend access
test_frontend_access() {
    local username=$1
    local token=$2
    local expected_redirect=$3
    
    echo -e "\n${YELLOW}Testing frontend access for: $username${NC}" | tee -a $LOG_FILE
    
    # Check if frontend is accessible
    frontend_status=$(curl -s -o /dev/null -w "%{http_code}" "$FRONTEND_URL")
    
    if [[ $frontend_status -eq 200 ]]; then
        echo -e "${GREEN}✓ Frontend is accessible${NC}" | tee -a $LOG_FILE
        echo "Expected redirect after login: $expected_redirect" | tee -a $LOG_FILE
    else
        echo -e "${RED}✗ Frontend not accessible (HTTP $frontend_status)${NC}" | tee -a $LOG_FILE
    fi
}

# Test all accounts
echo -e "\n${YELLOW}=== Testing Standard Accounts ===${NC}" | tee -a $LOG_FILE

# Admin account
test_login "admin" "admin123" "admin"

# Courier accounts (four levels)
echo -e "\n${YELLOW}=== Testing Courier Accounts ===${NC}" | tee -a $LOG_FILE
test_login "courier_level4_city" "courier123" "courier"
test_login "courier_level3_school" "courier123" "courier"
test_login "courier_level2_zone" "courier123" "courier"
test_login "courier_level1_building" "courier123" "courier"

# Regular user account
echo -e "\n${YELLOW}=== Testing Regular User Account ===${NC}" | tee -a $LOG_FILE
test_login "test_user" "user123" "user"

# Test with wrong credentials
echo -e "\n${YELLOW}=== Testing Invalid Credentials ===${NC}" | tee -a $LOG_FILE
test_login "admin" "wrongpassword" "admin"
test_login "nonexistent" "password123" "user"

# Summary
echo -e "\n${YELLOW}=== Test Summary ===${NC}" | tee -a $LOG_FILE
echo "Test completed at $(date)" | tee -a $LOG_FILE
echo "Results saved to: $LOG_FILE" | tee -a $LOG_FILE

# Check database for user records
echo -e "\n${YELLOW}=== Database Verification ===${NC}" | tee -a $LOG_FILE
echo "Checking user records in database..." | tee -a $LOG_FILE

# Create SQL verification script
cat > verify-users.sql << EOF
SELECT username, role, is_active, created_at 
FROM users 
WHERE username IN ('admin', 'courier_level4_city', 'courier_level3_school', 
                   'courier_level2_zone', 'courier_level1_building', 'test_user')
ORDER BY username;
EOF

echo "SQL query saved to verify-users.sql" | tee -a $LOG_FILE

echo -e "\n${GREEN}Authentication test completed!${NC}" | tee -a $LOG_FILE