#!/bin/bash

# Authentication Test Script for OpenPenPal
# Updated with correct test account names from database

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
    response_body=$(echo "$response" | sed '$d')
    
    echo "HTTP Status: $http_code" | tee -a $LOG_FILE
    echo "Response: $response_body" | tee -a $LOG_FILE
    
    # Check if login was successful
    if [[ $http_code -eq 200 ]]; then
        # Extract token from response
        token=$(echo "$response_body" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
        
        if [[ -n "$token" ]]; then
            echo -e "${GREEN}✓ Login successful! Token received${NC}" | tee -a $LOG_FILE
            
            # Verify user info
            user_info=$(curl -s -X GET "$BASE_URL/api/v1/auth/me" \
                -H "Authorization: Bearer $token")
            
            echo "User info: $user_info" | tee -a $LOG_FILE
            
            # Extract role from user info
            actual_role=$(echo "$user_info" | grep -o '"role":"[^"]*' | cut -d'"' -f4)
            
            if [[ "$actual_role" == "$expected_role" ]] || [[ "$expected_role" == "any" ]]; then
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

# Test all accounts based on database records
echo -e "\n${YELLOW}=== Testing Admin Accounts ===${NC}" | tee -a $LOG_FILE
test_login "admin" "admin123" "super_admin"
test_login "platform_admin" "admin123" "platform_admin"
test_login "school_admin" "admin123" "school_admin"

echo -e "\n${YELLOW}=== Testing Courier Accounts (Four Levels) ===${NC}" | tee -a $LOG_FILE
test_login "courier_level4" "courier123" "courier_level4"
test_login "courier_level3" "courier123" "courier_level3"
test_login "courier_level2" "courier123" "courier_level2"
test_login "courier_level1" "courier123" "courier_level1"

echo -e "\n${YELLOW}=== Testing Legacy Courier Accounts ===${NC}" | tee -a $LOG_FILE
test_login "courier_city" "courier123" "platform_admin"
test_login "courier_school" "courier123" "courier_coordinator"
test_login "courier_area" "courier123" "senior_courier"
test_login "courier_building" "courier123" "courier"

echo -e "\n${YELLOW}=== Testing Regular User Accounts ===${NC}" | tee -a $LOG_FILE
test_login "alice" "secret" "user"
test_login "bob" "secret" "user"

echo -e "\n${YELLOW}=== Testing Other Role Accounts ===${NC}" | tee -a $LOG_FILE
test_login "courier1" "courier123" "courier"
test_login "senior_courier" "senior123" "senior_courier"
test_login "coordinator" "coord123" "courier_coordinator"

echo -e "\n${YELLOW}=== Testing Invalid Credentials ===${NC}" | tee -a $LOG_FILE
test_login "admin" "wrongpassword" "any"
test_login "nonexistent" "password123" "any"

# Summary
echo -e "\n${YELLOW}=== Test Summary ===${NC}" | tee -a $LOG_FILE
echo "Test completed at $(date)" | tee -a $LOG_FILE
echo "Results saved to: $LOG_FILE" | tee -a $LOG_FILE

# Frontend Button Test
echo -e "\n${YELLOW}=== Testing Frontend Button Interactions ===${NC}" | tee -a $LOG_FILE
echo "Testing if frontend login page is accessible..." | tee -a $LOG_FILE

# Check login page
login_page_status=$(curl -s -o /dev/null -w "%{http_code}" "$FRONTEND_URL/login")
echo "Login page status: $login_page_status" | tee -a $LOG_FILE

if [[ $login_page_status -eq 200 ]]; then
    echo -e "${GREEN}✓ Login page is accessible${NC}" | tee -a $LOG_FILE
else
    echo -e "${RED}✗ Login page not accessible${NC}" | tee -a $LOG_FILE
fi

# Check register page
register_page_status=$(curl -s -o /dev/null -w "%{http_code}" "$FRONTEND_URL/register")
echo "Register page status: $register_page_status" | tee -a $LOG_FILE

echo -e "\n${GREEN}Authentication test completed!${NC}" | tee -a $LOG_FILE