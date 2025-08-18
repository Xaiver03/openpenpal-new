#!/bin/bash

# Full Integration Test Script
# Tests complete frontend-backend-database interaction with optimized configurations

set -e

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
BACKEND_URL="http://localhost:8080"
FRONTEND_URL="http://localhost:3000"
DB_NAME="openpenpal"
TEST_USER="alice"
TEST_PASS="secret123"
LOG_FILE="integration-test-$(date +%Y%m%d-%H%M%S).log"

# Function to print colored output
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1" | tee -a "$LOG_FILE"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1" | tee -a "$LOG_FILE"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1" | tee -a "$LOG_FILE"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1" | tee -a "$LOG_FILE"
}

# Function to check service status
check_service() {
    local service=$1
    local url=$2
    
    if curl -s -o /dev/null -w "%{http_code}" "$url" | grep -q "200\|404"; then
        print_success "$service is running at $url"
        return 0
    else
        print_error "$service is not responding at $url"
        return 1
    fi
}

# Function to test database connection
test_database() {
    print_info "Testing database connection..."
    
    if psql -h localhost -U postgres -d "$DB_NAME" -c "SELECT 1;" > /dev/null 2>&1; then
        print_success "Database connection successful"
        
        # Check connection pool status
        local conn_count=$(psql -h localhost -U postgres -d "$DB_NAME" -t -c "SELECT count(*) FROM pg_stat_activity WHERE datname = '$DB_NAME';" 2>/dev/null | xargs)
        print_info "Current database connections: $conn_count"
        
        # Check SSL status
        local ssl_status=$(psql -h localhost -U postgres -d "$DB_NAME" -t -c "SHOW ssl;" 2>/dev/null | xargs)
        print_info "SSL status: $ssl_status"
        
        return 0
    else
        print_error "Database connection failed"
        return 1
    fi
}

# Function to test authentication
test_authentication() {
    print_info "Testing authentication flow..."
    
    # Login request
    local response=$(curl -s -X POST "$BACKEND_URL/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d "{\"username\":\"$TEST_USER\",\"password\":\"$TEST_PASS\"}" \
        -w "\n%{http_code}")
    
    local http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | head -n-1)
    
    if [[ "$http_code" == "200" ]]; then
        print_success "Authentication successful"
        
        # Extract token
        TOKEN=$(echo "$body" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
        if [[ -n "$TOKEN" ]]; then
            print_info "JWT token obtained"
            return 0
        else
            print_error "No token in response"
            return 1
        fi
    else
        print_error "Authentication failed with HTTP $http_code"
        echo "$body" >> "$LOG_FILE"
        return 1
    fi
}

# Function to test API endpoints
test_api_endpoints() {
    print_info "Testing API endpoints..."
    
    # Test user profile
    local response=$(curl -s -X GET "$BACKEND_URL/api/v1/users/profile" \
        -H "Authorization: Bearer $TOKEN" \
        -w "\n%{http_code}")
    
    local http_code=$(echo "$response" | tail -n1)
    if [[ "$http_code" == "200" ]]; then
        print_success "User profile endpoint working"
    else
        print_error "User profile endpoint failed with HTTP $http_code"
    fi
    
    # Test letters endpoint
    response=$(curl -s -X GET "$BACKEND_URL/api/v1/letters" \
        -H "Authorization: Bearer $TOKEN" \
        -w "\n%{http_code}")
    
    http_code=$(echo "$response" | tail -n1)
    if [[ "$http_code" == "200" ]]; then
        print_success "Letters endpoint working"
    else
        print_error "Letters endpoint failed with HTTP $http_code"
    fi
    
    # Test credits endpoint
    response=$(curl -s -X GET "$BACKEND_URL/api/v1/credits/balance" \
        -H "Authorization: Bearer $TOKEN" \
        -w "\n%{http_code}")
    
    http_code=$(echo "$response" | tail -n1)
    if [[ "$http_code" == "200" ]]; then
        print_success "Credits endpoint working"
    else
        print_error "Credits endpoint failed with HTTP $http_code"
    fi
}

# Function to test connection pool under load
test_connection_pool() {
    print_info "Testing connection pool performance..."
    
    # Start monitoring
    ./scripts/monitor-pool.sh monitor > pool-monitor.log 2>&1 &
    MONITOR_PID=$!
    
    # Generate concurrent requests
    print_info "Generating 50 concurrent requests..."
    for i in {1..50}; do
        curl -s -X GET "$BACKEND_URL/api/v1/letters" \
            -H "Authorization: Bearer $TOKEN" > /dev/null 2>&1 &
    done
    
    # Wait for requests to complete
    wait
    
    # Check pool metrics
    sleep 2
    local active_conns=$(psql -h localhost -U postgres -d "$DB_NAME" -t -c "
        SELECT COUNT(*) FILTER (WHERE state = 'active') 
        FROM pg_stat_activity 
        WHERE datname = '$DB_NAME';" 2>/dev/null | xargs)
    
    print_info "Active connections after load test: $active_conns"
    
    # Stop monitoring
    kill $MONITOR_PID 2>/dev/null || true
    
    # Analyze results
    if [[ -f pool-monitor.log ]]; then
        local max_conns=$(grep "Total:" pool-monitor.log | awk -F': ' '{print $2}' | sort -nr | head -1)
        print_info "Peak connections during test: $max_conns"
    fi
}

# Function to test SSL configuration
test_ssl_configuration() {
    print_info "Testing SSL configuration..."
    
    # Check if SSL files exist
    if [[ -f "./dev-ssl/ca-cert.pem" ]]; then
        print_info "Development SSL certificates found"
    else
        print_warning "No development SSL certificates found"
        print_info "Generating self-signed certificates..."
        ./scripts/setup-ssl.sh generate-dev
    fi
    
    # Test SSL connection
    local ssl_mode=$(echo $DB_SSLMODE)
    print_info "Current SSL mode: ${ssl_mode:-disable}"
    
    # Verify SSL certificate if in verify mode
    if [[ "$ssl_mode" == "verify-ca" ]] || [[ "$ssl_mode" == "verify-full" ]]; then
        ./scripts/setup-ssl.sh verify
    fi
}

# Function to test data consistency
test_data_consistency() {
    print_info "Testing data consistency..."
    
    # Create a test letter
    local letter_data='{
        "title": "Integration Test Letter",
        "content": "This is a test letter created during integration testing.",
        "recipient_op_code": "PK5F01",
        "style": "classic"
    }'
    
    local response=$(curl -s -X POST "$BACKEND_URL/api/v1/letters" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "$letter_data" \
        -w "\n%{http_code}")
    
    local http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | head -n-1)
    
    if [[ "$http_code" == "201" ]] || [[ "$http_code" == "200" ]]; then
        print_success "Letter created successfully"
        
        # Extract letter ID
        local letter_id=$(echo "$body" | grep -o '"id":"[^"]*' | cut -d'"' -f4)
        if [[ -n "$letter_id" ]]; then
            print_info "Created letter ID: $letter_id"
            
            # Verify in database
            local db_check=$(psql -h localhost -U postgres -d "$DB_NAME" -t -c "
                SELECT COUNT(*) FROM letters WHERE id = '$letter_id';" 2>/dev/null | xargs)
            
            if [[ "$db_check" == "1" ]]; then
                print_success "Letter verified in database"
            else
                print_error "Letter not found in database"
            fi
        fi
    else
        print_error "Letter creation failed with HTTP $http_code"
    fi
}

# Function to generate test report
generate_report() {
    print_info "Generating test report..."
    
    cat > "integration-test-report.md" << EOF
# Integration Test Report

**Date**: $(date)
**Environment**: ${ENVIRONMENT:-development}

## Test Results

### Service Status
- Backend: $(check_service "Backend" "$BACKEND_URL/health" && echo "✅ Running" || echo "❌ Not running")
- Frontend: $(check_service "Frontend" "$FRONTEND_URL" && echo "✅ Running" || echo "❌ Not running")
- Database: $(test_database > /dev/null 2>&1 && echo "✅ Connected" || echo "❌ Not connected")

### Configuration
- SSL Mode: ${DB_SSLMODE:-disable}
- High Traffic Mode: ${HIGH_TRAFFIC_MODE:-false}
- Connection Pool: Optimized for ${ENVIRONMENT:-development}

### Test Summary
$(grep -c SUCCESS "$LOG_FILE" 2>/dev/null || echo 0) tests passed
$(grep -c ERROR "$LOG_FILE" 2>/dev/null || echo 0) tests failed
$(grep -c WARNING "$LOG_FILE" 2>/dev/null || echo 0) warnings

### Connection Pool Metrics
$(tail -20 pool-monitor.log 2>/dev/null || echo "No metrics available")

### Recommendations
$(./scripts/monitor-pool.sh analyze | grep -A 10 "Recommendations" 2>/dev/null || echo "Run full analysis for recommendations")

## Log File
Full logs available at: $LOG_FILE
EOF
    
    print_success "Test report generated: integration-test-report.md"
}

# Main test execution
main() {
    print_info "Starting full integration test..."
    print_info "Environment: ${ENVIRONMENT:-development}"
    echo ""
    
    # Check prerequisites
    print_info "Checking prerequisites..."
    check_service "Backend" "$BACKEND_URL/health" || exit 1
    test_database || exit 1
    echo ""
    
    # Run tests
    test_authentication || exit 1
    echo ""
    
    test_api_endpoints
    echo ""
    
    test_ssl_configuration
    echo ""
    
    test_connection_pool
    echo ""
    
    test_data_consistency
    echo ""
    
    # Generate report
    generate_report
    
    print_success "Integration test completed!"
    print_info "Check integration-test-report.md for detailed results"
}

# Run main function
main "$@"