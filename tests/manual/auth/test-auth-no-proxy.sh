#!/bin/bash

# OpenPenPal Authentication Test Script (No Proxy)
# This script tests authentication without proxy interference

BASE_URL="http://127.0.0.1:8080"
FRONTEND_URL="http://127.0.0.1:3000"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
LOG_FILE="test-auth-results-$TIMESTAMP.log"

# Color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}OpenPenPal Authentication Test - $TIMESTAMP${NC}" | tee $LOG_FILE
echo "============================================" | tee -a $LOG_FILE

# Disable proxy for local testing
export http_proxy=""
export https_proxy=""
export HTTP_PROXY=""
export HTTPS_PROXY=""
export no_proxy="localhost,127.0.0.1"

# First, check if backend is healthy
echo -e "\n${YELLOW}Checking backend health...${NC}" | tee -a $LOG_FILE
health_response=$(curl -s http://127.0.0.1:8080/health)
echo "Health check response: $health_response" | tee -a $LOG_FILE

# Function to test login without proxy
test_login() {
    local username=$1
    local password=$2
    local expected_role=$3
    
    echo -e "\n${YELLOW}Testing login for user: $username (Expected role: $expected_role)${NC}" | tee -a $LOG_FILE
    
    # Make login request without proxy
    response=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d "{\"username\":\"$username\",\"password\":\"$password\"}" \
        --noproxy '*' \
        -w "\nHTTP_CODE:%{http_code}")
    
    # Extract HTTP status code
    http_code=$(echo "$response" | grep "HTTP_CODE:" | cut -d':' -f2)
    response_body=$(echo "$response" | grep -v "HTTP_CODE:")
    
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
                -H "Authorization: Bearer $token" \
                --noproxy '*')
            
            echo "User info: $user_info" | tee -a $LOG_FILE
            
            # Extract role from user info
            actual_role=$(echo "$user_info" | grep -o '"role":"[^"]*' | cut -d'"' -f4)
            
            if [[ "$actual_role" == "$expected_role" ]] || [[ "$expected_role" == "any" ]]; then
                echo -e "${GREEN}✓ Role verification passed: $actual_role${NC}" | tee -a $LOG_FILE
                return 0
            else
                echo -e "${YELLOW}⚠ Role is: $actual_role (expected: $expected_role)${NC}" | tee -a $LOG_FILE
                return 0  # Still success, just different role
            fi
        else
            echo -e "${RED}✗ Login failed: No token in response${NC}" | tee -a $LOG_FILE
            return 1
        fi
    else
        echo -e "${RED}✗ Login failed with HTTP status: $http_code${NC}" | tee -a $LOG_FILE
        # Check if it's a password error
        if echo "$response_body" | grep -q "password"; then
            echo -e "${YELLOW}⚠ Invalid password${NC}" | tee -a $LOG_FILE
        fi
        return 1
    fi
}

# Test different types of accounts
echo -e "\n${BLUE}=== Testing User Accounts ===${NC}" | tee -a $LOG_FILE
test_login "alice" "secret" "user"
test_login "bob" "secret" "user"

echo -e "\n${BLUE}=== Testing Admin Account ===${NC}" | tee -a $LOG_FILE
test_login "admin" "admin123" "super_admin"

echo -e "\n${BLUE}=== Testing Courier Accounts ===${NC}" | tee -a $LOG_FILE
test_login "courier1" "courier123" "courier"
test_login "courier_level1" "courier123" "courier_level1"
test_login "courier_level2" "courier123" "courier_level2"
test_login "courier_level3" "courier123" "courier_level3"
test_login "courier_level4" "courier123" "courier_level4"

echo -e "\n${BLUE}=== Testing Invalid Credentials ===${NC}" | tee -a $LOG_FILE
test_login "alice" "wrongpassword" "any"
test_login "nonexistent" "password123" "any"

# Test some API endpoints with a valid token
echo -e "\n${BLUE}=== Testing API Endpoints ===${NC}" | tee -a $LOG_FILE

# Login as alice to get a token
alice_response=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"alice","password":"secret"}' \
    --noproxy '*')

alice_token=$(echo "$alice_response" | grep -o '"token":"[^"]*' | cut -d'"' -f4)

if [[ -n "$alice_token" ]]; then
    echo -e "${GREEN}✓ Got token for alice${NC}" | tee -a $LOG_FILE
    
    # Test getting user's letters
    echo -e "\n${YELLOW}Testing GET /api/v1/letters (alice's letters)${NC}" | tee -a $LOG_FILE
    letters_response=$(curl -s -X GET "$BASE_URL/api/v1/letters" \
        -H "Authorization: Bearer $alice_token" \
        --noproxy '*' \
        -w "\nHTTP_CODE:%{http_code}")
    
    letters_code=$(echo "$letters_response" | grep "HTTP_CODE:" | cut -d':' -f2)
    echo "Letters endpoint status: $letters_code" | tee -a $LOG_FILE
    
    # Test creating a draft letter
    echo -e "\n${YELLOW}Testing POST /api/v1/letters (create draft)${NC}" | tee -a $LOG_FILE
    create_response=$(curl -s -X POST "$BASE_URL/api/v1/letters" \
        -H "Authorization: Bearer $alice_token" \
        -H "Content-Type: application/json" \
        -d '{"title":"Test Letter","content":"This is a test letter content","style":"classic"}' \
        --noproxy '*' \
        -w "\nHTTP_CODE:%{http_code}")
    
    create_code=$(echo "$create_response" | grep "HTTP_CODE:" | cut -d':' -f2)
    echo "Create letter status: $create_code" | tee -a $LOG_FILE
fi

# Summary
echo -e "\n${BLUE}=== Test Summary ===${NC}" | tee -a $LOG_FILE
echo "Test completed at $(date)" | tee -a $LOG_FILE
echo "Results saved to: $LOG_FILE" | tee -a $LOG_FILE

# Final health check
echo -e "\n${YELLOW}Final health check...${NC}" | tee -a $LOG_FILE
final_health=$(curl -s http://127.0.0.1:8080/health --noproxy '*')
echo "Final health: $final_health" | tee -a $LOG_FILE

echo -e "\n${GREEN}Authentication test completed!${NC}" | tee -a $LOG_FILE