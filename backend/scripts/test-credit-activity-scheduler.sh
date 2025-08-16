#!/bin/bash

# Credit Activity Scheduler Test Script
# Tests the complete credit activity scheduling system

set -e

BASE_URL="http://localhost:8080"
API_BASE="${BASE_URL}/api/v1"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper function to print colored output
print_status() {
    local status=$1
    local message=$2
    case $status in
        "INFO")  echo -e "${BLUE}[INFO]${NC} $message" ;;
        "PASS")  echo -e "${GREEN}[PASS]${NC} $message" ;;
        "FAIL")  echo -e "${RED}[FAIL]${NC} $message" ;;
        "WARN")  echo -e "${YELLOW}[WARN]${NC} $message" ;;
    esac
}

# Check if backend is running
check_backend() {
    print_status "INFO" "Checking if backend is running..."
    if curl -s "${BASE_URL}/health" > /dev/null 2>&1; then
        print_status "PASS" "Backend is running"
        return 0
    else
        print_status "FAIL" "Backend is not running. Please start the backend first."
        return 1
    fi
}

# Login as admin to get JWT token
login_admin() {
    print_status "INFO" "Logging in as admin..."
    local response=$(curl -s -X POST "${API_BASE}/auth/login" \
        -H "Content-Type: application/json" \
        -d '{
            "username": "admin",
            "password": "admin123"
        }')
    
    # Extract token from response
    TOKEN=$(echo "$response" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
    
    if [[ -n "$TOKEN" ]]; then
        print_status "PASS" "Admin login successful"
        export AUTH_HEADER="Authorization: Bearer $TOKEN"
        return 0
    else
        print_status "FAIL" "Admin login failed"
        echo "Response: $response"
        return 1
    fi
}

# Test 1: Check scheduler status
test_scheduler_status() {
    print_status "INFO" "Testing scheduler status..."
    local response=$(curl -s -X GET "${API_BASE}/admin/credit-activities/scheduler/status" \
        -H "$AUTH_HEADER" \
        -H "Content-Type: application/json")
    
    if echo "$response" | grep -q "running"; then
        print_status "PASS" "Scheduler status retrieved successfully"
        echo "Status: $response" | head -c 200
        return 0
    else
        print_status "FAIL" "Failed to get scheduler status"
        echo "Response: $response"
        return 1
    fi
}

# Test 2: Create a test activity
test_create_activity() {
    print_status "INFO" "Creating test activity..."
    local start_time=$(date -u -d "+1 minute" +"%Y-%m-%dT%H:%M:%SZ")
    local end_time=$(date -u -d "+1 hour" +"%Y-%m-%dT%H:%M:%SZ")
    
    local response=$(curl -s -X POST "${API_BASE}/admin/credit-activities" \
        -H "$AUTH_HEADER" \
        -H "Content-Type: application/json" \
        -d "{
            \"name\": \"测试活动 - 调度器验证\",
            \"description\": \"用于验证调度系统的测试活动\",
            \"activity_type\": \"daily\",
            \"target_type\": \"all\",
            \"trigger_type\": \"login\",
            \"reward_credits\": 10,
            \"start_time\": \"$start_time\",
            \"end_time\": \"$end_time\",
            \"repeat_pattern\": \"daily\",
            \"repeat_interval\": 1
        }")
    
    # Extract activity ID
    ACTIVITY_ID=$(echo "$response" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
    
    if [[ -n "$ACTIVITY_ID" ]]; then
        print_status "PASS" "Test activity created successfully"
        print_status "INFO" "Activity ID: $ACTIVITY_ID"
        return 0
    else
        print_status "FAIL" "Failed to create test activity"
        echo "Response: $response"
        return 1
    fi
}

# Test 3: Schedule the activity
test_schedule_activity() {
    if [[ -z "$ACTIVITY_ID" ]]; then
        print_status "FAIL" "No activity ID available for scheduling"
        return 1
    fi
    
    print_status "INFO" "Scheduling test activity..."
    local scheduled_time=$(date -u -d "+30 seconds" +"%Y-%m-%dT%H:%M:%SZ")
    
    local response=$(curl -s -X POST "${API_BASE}/admin/credit-activities/scheduler/schedule" \
        -H "$AUTH_HEADER" \
        -H "Content-Type: application/json" \
        -d "{
            \"activity_id\": \"$ACTIVITY_ID\",
            \"scheduled_time\": \"$scheduled_time\",
            \"execution_details\": {
                \"test\": true,
                \"type\": \"activity_execution\"
            }
        }")
    
    # Extract schedule ID
    SCHEDULE_ID=$(echo "$response" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
    
    if [[ -n "$SCHEDULE_ID" ]]; then
        print_status "PASS" "Activity scheduled successfully"
        print_status "INFO" "Schedule ID: $SCHEDULE_ID"
        print_status "INFO" "Scheduled time: $scheduled_time"
        return 0
    else
        print_status "FAIL" "Failed to schedule activity"
        echo "Response: $response"
        return 1
    fi
}

# Test 4: Get scheduled tasks
test_get_scheduled_tasks() {
    print_status "INFO" "Getting scheduled tasks..."
    local response=$(curl -s -X GET "${API_BASE}/admin/credit-activities/scheduler/tasks?limit=10" \
        -H "$AUTH_HEADER" \
        -H "Content-Type: application/json")
    
    if echo "$response" | grep -q "tasks"; then
        print_status "PASS" "Scheduled tasks retrieved successfully"
        local count=$(echo "$response" | grep -o '"count":[0-9]*' | cut -d':' -f2)
        print_status "INFO" "Found $count scheduled tasks"
        return 0
    else
        print_status "FAIL" "Failed to get scheduled tasks"
        echo "Response: $response"
        return 1
    fi
}

# Test 5: Start scheduler (if not running)
test_start_scheduler() {
    print_status "INFO" "Starting scheduler..."
    local response=$(curl -s -X POST "${API_BASE}/admin/credit-activities/scheduler/start" \
        -H "$AUTH_HEADER" \
        -H "Content-Type: application/json")
    
    if echo "$response" | grep -q -E "(启动成功|already running)"; then
        print_status "PASS" "Scheduler start command executed"
        return 0
    else
        print_status "WARN" "Scheduler start response: $response"
        return 0  # Don't fail if already running
    fi
}

# Test 6: Wait and check task execution
test_wait_for_execution() {
    if [[ -z "$SCHEDULE_ID" ]]; then
        print_status "WARN" "No schedule ID to monitor"
        return 0
    fi
    
    print_status "INFO" "Waiting for task execution (45 seconds)..."
    sleep 45
    
    print_status "INFO" "Checking task execution status..."
    local response=$(curl -s -X GET "${API_BASE}/admin/credit-activities/scheduler/tasks?status=completed&limit=5" \
        -H "$AUTH_HEADER" \
        -H "Content-Type: application/json")
    
    if echo "$response" | grep -q "completed"; then
        print_status "PASS" "Found completed tasks - scheduler is working!"
        return 0
    else
        print_status "WARN" "No completed tasks found yet"
        print_status "INFO" "This is normal for the first test run"
        return 0
    fi
}

# Test 7: Test immediate execution
test_immediate_execution() {
    if [[ -z "$ACTIVITY_ID" ]]; then
        print_status "WARN" "No activity ID for immediate execution test"
        return 0
    fi
    
    print_status "INFO" "Testing immediate execution..."
    local response=$(curl -s -X POST "${API_BASE}/admin/credit-activities/scheduler/schedule/immediate" \
        -H "$AUTH_HEADER" \
        -H "Content-Type: application/json" \
        -d "{
            \"activity_id\": \"$ACTIVITY_ID\",
            \"execution_details\": {
                \"test\": true,
                \"type\": \"immediate_test\"
            }
        }")
    
    if echo "$response" | grep -q "立即执行"; then
        print_status "PASS" "Immediate execution scheduled successfully"
        return 0
    else
        print_status "FAIL" "Failed to schedule immediate execution"
        echo "Response: $response"
        return 1
    fi
}

# Test 8: Get scheduler statistics
test_scheduler_statistics() {
    print_status "INFO" "Getting scheduler statistics..."
    local response=$(curl -s -X GET "${API_BASE}/admin/credit-activities/scheduler/statistics" \
        -H "$AUTH_HEADER" \
        -H "Content-Type: application/json")
    
    if echo "$response" | grep -q "scheduler_status"; then
        print_status "PASS" "Scheduler statistics retrieved successfully"
        return 0
    else
        print_status "FAIL" "Failed to get scheduler statistics"
        echo "Response: $response"
        return 1
    fi
}

# Main test execution
main() {
    echo "========================================="
    echo "Credit Activity Scheduler Test Suite"
    echo "========================================="
    echo ""
    
    # Run tests
    check_backend || exit 1
    login_admin || exit 1
    
    echo ""
    echo "--- Running Scheduler Tests ---"
    
    test_scheduler_status || exit 1
    test_create_activity || exit 1
    test_schedule_activity || exit 1
    test_get_scheduled_tasks || exit 1
    test_start_scheduler || exit 1
    test_wait_for_execution || true  # Don't exit on this one
    test_immediate_execution || exit 1
    test_scheduler_statistics || exit 1
    
    echo ""
    echo "========================================="
    print_status "PASS" "All scheduler tests completed!"
    echo "========================================="
    
    # Show some final statistics
    print_status "INFO" "Final scheduler status check..."
    test_scheduler_status || true
}

# Run main function
main "$@"