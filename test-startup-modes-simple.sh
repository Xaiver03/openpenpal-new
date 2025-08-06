#!/bin/bash

# Simple test for OpenPenPal startup modes
# This script tests each startup mode with shorter timeouts

set -e

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$SCRIPT_DIR"
STARTUP_DIR="$PROJECT_ROOT/startup"
LOG_DIR="$PROJECT_ROOT/logs"

# Ensure log directory exists
mkdir -p "$LOG_DIR"

# Test report file
REPORT_FILE="$LOG_DIR/startup-test-simple-$(date +%Y%m%d-%H%M%S).md"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Initialize report
cat > "$REPORT_FILE" << EOF
# OpenPenPal Startup Modes Test Report (Simple)
Date: $(date)

## System Information
- Platform: $(uname -s)
- Node Version: $(node --version)
- Go Version: $(go version 2>/dev/null || echo "Go not installed")
- Python Version: $(python3 --version 2>/dev/null || echo "Python3 not installed")
- Java Version: $(java -version 2>&1 | head -n 1 || echo "Java not installed")

## Test Results

EOF

# Function to log
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Function to check service health
check_service_health() {
    local port=$1
    local service_name=$2
    local url="http://localhost:$port"
    
    # Try multiple endpoints
    for path in "/health" "/" ""; do
        if curl -s -f -o /dev/null --max-time 5 "$url$path" 2>/dev/null; then
            return 0
        fi
    done
    return 1
}

# Function to check if port is occupied
check_port_occupied() {
    local port=$1
    lsof -i :$port > /dev/null 2>&1
}

# Function to test a single mode with timeout
test_mode_with_timeout() {
    local mode=$1
    local timeout=${2:-60}
    
    log_info "=========================================="
    log_info "Testing mode: $mode (timeout: ${timeout}s)"
    log_info "=========================================="
    
    echo "" >> "$REPORT_FILE"
    echo "### Mode: $mode" >> "$REPORT_FILE"
    echo "Start Time: $(date +%H:%M:%S)" >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"
    
    # Stop any existing services
    log_info "Stopping existing services..."
    "$STARTUP_DIR/stop-all.sh" --quiet --force > /dev/null 2>&1 || true
    sleep 2
    
    # Start services in the mode with timeout
    log_info "Starting services in $mode mode..."
    local start_time=$(date +%s)
    
    # Use timeout command to prevent hanging
    local startup_log="$LOG_DIR/startup-$mode.log"
    if timeout ${timeout}s "$STARTUP_DIR/quick-start.sh" "$mode" --timeout=30 > "$startup_log" 2>&1; then
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))
        log_success "Services started successfully (took ${duration}s)"
        echo "**Status**: ✅ Started successfully" >> "$REPORT_FILE"
        echo "**Duration**: ${duration} seconds" >> "$REPORT_FILE"
        
        # Wait briefly for services to settle
        sleep 5
        
        # Check services
        check_services_for_mode "$mode"
        
    else
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))
        log_error "Failed to start services or timed out (took ${duration}s)"
        echo "**Status**: ❌ Failed to start or timed out" >> "$REPORT_FILE"
        echo "**Duration**: ${duration} seconds" >> "$REPORT_FILE"
        echo "" >> "$REPORT_FILE"
        echo "**Error Log (last 20 lines):**" >> "$REPORT_FILE"
        echo '```' >> "$REPORT_FILE"
        tail -n 20 "$startup_log" >> "$REPORT_FILE" 2>/dev/null || echo "No log file found" >> "$REPORT_FILE"
        echo '```' >> "$REPORT_FILE"
    fi
    
    # Stop services
    log_info "Stopping services..."
    "$STARTUP_DIR/stop-all.sh" --quiet --force > /dev/null 2>&1 || true
    
    echo "" >> "$REPORT_FILE"
    echo "End Time: $(date +%H:%M:%S)" >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"
    echo "---" >> "$REPORT_FILE"
    
    # Wait between tests
    sleep 3
}

# Function to check services for a mode
check_services_for_mode() {
    local mode=$1
    
    echo "" >> "$REPORT_FILE"
    echo "**Service Health Check:**" >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"
    
    case $mode in
        "simple"|"demo"|"development")
            # Check Go backend and Frontend
            check_and_report_service 8080 "Go Backend"
            check_and_report_service 3000 "Frontend"
            ;;
            
        "production"|"complete")
            # Check core services
            check_and_report_service 8080 "Go Backend"
            check_and_report_service 3000 "Frontend"
            
            # Check optional microservices
            if check_port_occupied 8000; then
                check_and_report_service 8000 "Gateway"
            else
                echo "- ⚠️ Gateway (port 8000): Not started" >> "$REPORT_FILE"
            fi
            
            if check_port_occupied 8001; then
                check_and_report_service 8001 "Write Service"
            else
                echo "- ⚠️ Write Service (port 8001): Not started" >> "$REPORT_FILE"
            fi
            
            if check_port_occupied 8002; then
                check_and_report_service 8002 "Courier Service"
            else
                echo "- ⚠️ Courier Service (port 8002): Not started" >> "$REPORT_FILE"
            fi
            
            if check_port_occupied 8003; then
                check_and_report_service 8003 "Admin Service"
            else
                echo "- ⚠️ Admin Service (port 8003): Not started (Java required)" >> "$REPORT_FILE"
            fi
            
            if check_port_occupied 8004; then
                check_and_report_service 8004 "OCR Service"
            else
                echo "- ⚠️ OCR Service (port 8004): Not started" >> "$REPORT_FILE"
            fi
            
            if check_port_occupied 3001; then
                check_and_report_service 3001 "Admin Frontend"
            else
                echo "- ⚠️ Admin Frontend (port 3001): Not started" >> "$REPORT_FILE"
            fi
            ;;
            
        "mock")
            # Check Mock service and Frontend
            check_and_report_service 8000 "Simple Mock"
            check_and_report_service 3000 "Frontend"
            ;;
    esac
    
    # List actually running processes
    echo "" >> "$REPORT_FILE"
    echo "**Actually Running Processes:**" >> "$REPORT_FILE"
    echo '```' >> "$REPORT_FILE"
    
    # Check each port
    for port in 3000 3001 8000 8080 8001 8002 8003 8004; do
        if check_port_occupied $port; then
            local process_info=$(lsof -i :$port 2>/dev/null | tail -n +2 | head -n 1 | awk '{print $1, $2}')
            echo "Port $port: $process_info" >> "$REPORT_FILE"
        fi
    done
    
    echo '```' >> "$REPORT_FILE"
}

# Function to check and report a single service
check_and_report_service() {
    local port=$1
    local service_name=$2
    
    if check_service_health "$port" "$service_name"; then
        log_success "✓ $service_name ($port) is healthy"
        echo "- ✅ $service_name (port $port): Healthy" >> "$REPORT_FILE"
    else
        log_error "✗ $service_name ($port) is not responding"
        echo "- ❌ $service_name (port $port): Not responding" >> "$REPORT_FILE"
    fi
}

# Function to get timeout for mode
get_mode_timeout() {
    local mode=$1
    case $mode in
        "simple"|"demo"|"development"|"mock") echo 60 ;;
        "production"|"complete") echo 120 ;;
        *) echo 60 ;;
    esac
}

# Main test execution
main() {
    log_info "Starting OpenPenPal startup modes test (simple version)"
    log_info "Test report will be saved to: $REPORT_FILE"
    
    # Test all modes
    for mode in "simple" "demo" "development" "mock" "production" "complete"; do
        local timeout=$(get_mode_timeout "$mode")
        test_mode_with_timeout "$mode" "$timeout"
    done
    
    # Final summary
    echo "" >> "$REPORT_FILE"
    echo "## Summary" >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"
    echo "Test completed at: $(date)" >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"
    echo "### Key Findings" >> "$REPORT_FILE"
    echo "- Simple modes (simple, demo, development, mock) should start quickly" >> "$REPORT_FILE"
    echo "- Complex modes (production, complete) may take longer and some services may fail" >> "$REPORT_FILE"
    echo "- Admin Service (port 8003) is expected to fail if Java is not installed" >> "$REPORT_FILE"
    echo "- Python-based services (Write, OCR) may fail if Python virtual environments are not set up" >> "$REPORT_FILE"
    
    log_success "All tests completed!"
    log_info "Test report saved to: $REPORT_FILE"
    
    # Display the report
    echo ""
    echo "===== TEST REPORT ====="
    cat "$REPORT_FILE"
}

# Run the tests
main