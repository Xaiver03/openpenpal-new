#!/bin/bash

# Test all OpenPenPal startup modes
# This script tests each startup mode in sequence

set -e

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$SCRIPT_DIR"
STARTUP_DIR="$PROJECT_ROOT/startup"
LOG_DIR="$PROJECT_ROOT/logs"

# Ensure log directory exists
mkdir -p "$LOG_DIR"

# Test report file
REPORT_FILE="$LOG_DIR/startup-test-report-$(date +%Y%m%d-%H%M%S).md"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Initialize report
cat > "$REPORT_FILE" << EOF
# OpenPenPal Startup Modes Test Report
Date: $(date)

## Test Configuration
- Project Root: $PROJECT_ROOT
- Test Platform: $(uname -s)
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
    local url="http://localhost:$port/health"
    
    if curl -s -f -o /dev/null "$url" 2>/dev/null; then
        return 0
    else
        # Try root endpoint if health endpoint doesn't exist
        if curl -s -f -o /dev/null "http://localhost:$port" 2>/dev/null; then
            return 0
        else
            return 1
        fi
    fi
}

# Function to test a single mode
test_mode() {
    local mode=$1
    log_info "=========================================="
    log_info "Testing mode: $mode"
    log_info "=========================================="
    
    echo "" >> "$REPORT_FILE"
    echo "### Mode: $mode" >> "$REPORT_FILE"
    echo "Start Time: $(date +%H:%M:%S)" >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"
    
    # Stop any existing services
    log_info "Stopping existing services..."
    "$STARTUP_DIR/stop-all.sh" --quiet --force > /dev/null 2>&1 || true
    sleep 3
    
    # Start services in the mode
    log_info "Starting services in $mode mode..."
    local start_time=$(date +%s)
    
    if "$STARTUP_DIR/quick-start.sh" "$mode" --timeout=120 > "$LOG_DIR/startup-$mode.log" 2>&1; then
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))
        log_success "Services started successfully (took ${duration}s)"
        echo "**Status**: ✅ Started successfully" >> "$REPORT_FILE"
        echo "**Duration**: ${duration} seconds" >> "$REPORT_FILE"
    else
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))
        log_error "Failed to start services (took ${duration}s)"
        echo "**Status**: ❌ Failed to start" >> "$REPORT_FILE"
        echo "**Duration**: ${duration} seconds" >> "$REPORT_FILE"
        echo "" >> "$REPORT_FILE"
        echo "**Error Log:**" >> "$REPORT_FILE"
        echo '```' >> "$REPORT_FILE"
        tail -n 20 "$LOG_DIR/startup-$mode.log" >> "$REPORT_FILE"
        echo '```' >> "$REPORT_FILE"
        
        # Continue to next mode
        return
    fi
    
    # Wait for services to stabilize
    log_info "Waiting for services to stabilize..."
    sleep 10
    
    # Check service health based on mode
    echo "" >> "$REPORT_FILE"
    echo "**Service Health Check:**" >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"
    
    case $mode in
        "simple"|"demo"|"development")
            # Check Go backend
            if check_service_health 8080 "Go Backend"; then
                log_success "✓ Go Backend (8080) is healthy"
                echo "- ✅ Go Backend (port 8080): Healthy" >> "$REPORT_FILE"
            else
                log_error "✗ Go Backend (8080) is not responding"
                echo "- ❌ Go Backend (port 8080): Not responding" >> "$REPORT_FILE"
            fi
            
            # Check Frontend
            if check_service_health 3000 "Frontend"; then
                log_success "✓ Frontend (3000) is healthy"
                echo "- ✅ Frontend (port 3000): Healthy" >> "$REPORT_FILE"
            else
                log_error "✗ Frontend (3000) is not responding"
                echo "- ❌ Frontend (port 3000): Not responding" >> "$REPORT_FILE"
            fi
            ;;
            
        "production"|"complete")
            # Check Go backend
            if check_service_health 8080 "Go Backend"; then
                log_success "✓ Go Backend (8080) is healthy"
                echo "- ✅ Go Backend (port 8080): Healthy" >> "$REPORT_FILE"
            else
                log_error "✗ Go Backend (8080) is not responding"
                echo "- ❌ Go Backend (port 8080): Not responding" >> "$REPORT_FILE"
            fi
            
            # Check Frontend
            if check_service_health 3000 "Frontend"; then
                log_success "✓ Frontend (3000) is healthy"
                echo "- ✅ Frontend (port 3000): Healthy" >> "$REPORT_FILE"
            else
                log_error "✗ Frontend (3000) is not responding"
                echo "- ❌ Frontend (port 3000): Not responding" >> "$REPORT_FILE"
            fi
            
            # Check Gateway (if exists)
            if [ -f "$LOG_DIR/gateway.pid" ]; then
                if check_service_health 8000 "Gateway"; then
                    log_success "✓ Gateway (8000) is healthy"
                    echo "- ✅ Gateway (port 8000): Healthy" >> "$REPORT_FILE"
                else
                    log_error "✗ Gateway (8000) is not responding"
                    echo "- ❌ Gateway (port 8000): Not responding" >> "$REPORT_FILE"
                fi
            else
                log_warning "Gateway service not started"
                echo "- ⚠️ Gateway (port 8000): Not started" >> "$REPORT_FILE"
            fi
            
            # Check Write Service (if exists)
            if [ -f "$LOG_DIR/write-service.pid" ]; then
                if check_service_health 8001 "Write Service"; then
                    log_success "✓ Write Service (8001) is healthy"
                    echo "- ✅ Write Service (port 8001): Healthy" >> "$REPORT_FILE"
                else
                    log_error "✗ Write Service (8001) is not responding"
                    echo "- ❌ Write Service (port 8001): Not responding" >> "$REPORT_FILE"
                fi
            else
                log_warning "Write Service not started"
                echo "- ⚠️ Write Service (port 8001): Not started" >> "$REPORT_FILE"
            fi
            
            # Check Courier Service (if exists)
            if [ -f "$LOG_DIR/courier-service.pid" ]; then
                if check_service_health 8002 "Courier Service"; then
                    log_success "✓ Courier Service (8002) is healthy"
                    echo "- ✅ Courier Service (port 8002): Healthy" >> "$REPORT_FILE"
                else
                    log_error "✗ Courier Service (8002) is not responding"
                    echo "- ❌ Courier Service (port 8002): Not responding" >> "$REPORT_FILE"
                fi
            else
                log_warning "Courier Service not started"
                echo "- ⚠️ Courier Service (port 8002): Not started" >> "$REPORT_FILE"
            fi
            
            # Check Admin Service (Java - expected to fail)
            if [ -f "$LOG_DIR/admin-service.pid" ]; then
                if check_service_health 8003 "Admin Service"; then
                    log_success "✓ Admin Service (8003) is healthy"
                    echo "- ✅ Admin Service (port 8003): Healthy" >> "$REPORT_FILE"
                else
                    log_error "✗ Admin Service (8003) is not responding"
                    echo "- ❌ Admin Service (port 8003): Not responding" >> "$REPORT_FILE"
                fi
            else
                log_warning "Admin Service not started (expected - Java not installed)"
                echo "- ⚠️ Admin Service (port 8003): Not started (Java not installed)" >> "$REPORT_FILE"
            fi
            
            # Check OCR Service (if exists)
            if [ -f "$LOG_DIR/ocr-service.pid" ]; then
                if check_service_health 8004 "OCR Service"; then
                    log_success "✓ OCR Service (8004) is healthy"
                    echo "- ✅ OCR Service (port 8004): Healthy" >> "$REPORT_FILE"
                else
                    log_error "✗ OCR Service (8004) is not responding"
                    echo "- ❌ OCR Service (port 8004): Not responding" >> "$REPORT_FILE"
                fi
            else
                log_warning "OCR Service not started"
                echo "- ⚠️ OCR Service (port 8004): Not started" >> "$REPORT_FILE"
            fi
            
            # Check Admin Frontend (if exists)
            if [ -f "$LOG_DIR/admin-frontend.pid" ]; then
                if check_service_health 3001 "Admin Frontend"; then
                    log_success "✓ Admin Frontend (3001) is healthy"
                    echo "- ✅ Admin Frontend (port 3001): Healthy" >> "$REPORT_FILE"
                else
                    log_error "✗ Admin Frontend (3001) is not responding"
                    echo "- ❌ Admin Frontend (port 3001): Not responding" >> "$REPORT_FILE"
                fi
            else
                log_warning "Admin Frontend not started"
                echo "- ⚠️ Admin Frontend (port 3001): Not started" >> "$REPORT_FILE"
            fi
            ;;
            
        "mock")
            # Check Simple Mock Service
            if check_service_health 8000 "Simple Mock"; then
                log_success "✓ Simple Mock (8000) is healthy"
                echo "- ✅ Simple Mock (port 8000): Healthy" >> "$REPORT_FILE"
            else
                log_error "✗ Simple Mock (8000) is not responding"
                echo "- ❌ Simple Mock (port 8000): Not responding" >> "$REPORT_FILE"
            fi
            
            # Check Frontend
            if check_service_health 3000 "Frontend"; then
                log_success "✓ Frontend (3000) is healthy"
                echo "- ✅ Frontend (port 3000): Healthy" >> "$REPORT_FILE"
            else
                log_error "✗ Frontend (3000) is not responding"
                echo "- ❌ Frontend (port 3000): Not responding" >> "$REPORT_FILE"
            fi
            ;;
    esac
    
    # List running processes
    echo "" >> "$REPORT_FILE"
    echo "**Running Processes:**" >> "$REPORT_FILE"
    echo '```' >> "$REPORT_FILE"
    ls -la "$LOG_DIR"/*.pid 2>/dev/null | awk '{print $9}' | while read pidfile; do
        if [ -f "$pidfile" ]; then
            pid=$(cat "$pidfile")
            service=$(basename "$pidfile" .pid)
            if ps -p "$pid" > /dev/null 2>&1; then
                echo "- $service (PID: $pid)" >> "$REPORT_FILE"
            fi
        fi
    done
    echo '```' >> "$REPORT_FILE"
    
    # Stop services
    log_info "Stopping services..."
    "$STARTUP_DIR/stop-all.sh" --quiet --force > /dev/null 2>&1 || true
    
    echo "" >> "$REPORT_FILE"
    echo "End Time: $(date +%H:%M:%S)" >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"
    echo "---" >> "$REPORT_FILE"
    
    # Wait between tests
    sleep 5
}

# Main test execution
main() {
    log_info "Starting OpenPenPal startup modes test"
    log_info "Test report will be saved to: $REPORT_FILE"
    
    # Test all modes
    modes=("simple" "demo" "development" "production" "mock" "complete")
    
    for mode in "${modes[@]}"; do
        test_mode "$mode"
    done
    
    # Final summary
    echo "" >> "$REPORT_FILE"
    echo "## Summary" >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"
    echo "Test completed at: $(date)" >> "$REPORT_FILE"
    
    log_success "All tests completed!"
    log_info "Test report saved to: $REPORT_FILE"
    
    # Display the report
    echo ""
    echo "===== TEST REPORT ====="
    cat "$REPORT_FILE"
}

# Run the tests
main