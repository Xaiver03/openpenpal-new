#!/bin/bash

# Health check script for OpenPenPal services
# Provides detailed health status for all services

set -e

# Color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Base URLs
BACKEND_URL="http://127.0.0.1:8080"
FRONTEND_URL="http://127.0.0.1:3000"

# Function to check service health
check_service() {
    local name=$1
    local url=$2
    local expected_status=${3:-200}
    
    echo -n "Checking $name... "
    
    # Disable proxy for local requests
    response=$(curl -s -o /dev/null -w "%{http_code}" "$url" --noproxy '*' --connect-timeout 5 2>/dev/null || echo "000")
    
    if [[ "$response" == "$expected_status" ]]; then
        echo -e "${GREEN}✓ Healthy${NC} (HTTP $response)"
        return 0
    elif [[ "$response" == "000" ]]; then
        echo -e "${RED}✗ Unreachable${NC}"
        return 1
    else
        echo -e "${YELLOW}⚠ Unhealthy${NC} (HTTP $response)"
        return 1
    fi
}

# Function to check detailed backend health
check_backend_health() {
    echo -e "\n${BLUE}Backend Health Details:${NC}"
    
    health_response=$(curl -s "$BACKEND_URL/health" --noproxy '*' 2>/dev/null || echo "{}")
    
    if [[ -n "$health_response" ]] && [[ "$health_response" != "{}" ]]; then
        echo "$health_response" | jq '.' 2>/dev/null || echo "$health_response"
    else
        echo -e "${RED}Unable to get health details${NC}"
    fi
}

# Function to check process status
check_process() {
    local process_name=$1
    local pid_file=$2
    
    echo -n "Checking process $process_name... "
    
    if [[ -f "$pid_file" ]]; then
        pid=$(cat "$pid_file")
        if ps -p "$pid" > /dev/null 2>&1; then
            echo -e "${GREEN}✓ Running${NC} (PID: $pid)"
            return 0
        else
            echo -e "${RED}✗ Dead${NC} (Stale PID: $pid)"
            return 1
        fi
    else
        # Try to find by process name
        pid=$(pgrep -f "$process_name" | head -1)
        if [[ -n "$pid" ]]; then
            echo -e "${GREEN}✓ Running${NC} (PID: $pid, no PID file)"
            return 0
        else
            echo -e "${RED}✗ Not running${NC}"
            return 1
        fi
    fi
}

# Function to check port status
check_port() {
    local port=$1
    local service=$2
    
    echo -n "Port $port ($service)... "
    
    if lsof -i ":$port" | grep -q LISTEN; then
        echo -e "${GREEN}✓ Listening${NC}"
        return 0
    else
        echo -e "${RED}✗ Not listening${NC}"
        return 1
    fi
}

# Main health check
echo -e "${BLUE}=== OpenPenPal Health Check ===${NC}"
echo "Time: $(date)"
echo

# Check services
echo -e "${YELLOW}Service Status:${NC}"
check_service "Backend API" "$BACKEND_URL/health"
check_service "Frontend" "$FRONTEND_URL"

# Check processes
echo -e "\n${YELLOW}Process Status:${NC}"
check_process "openpenpal-backend" "logs/go-backend.pid"
check_process "next-server" "logs/frontend.pid"

# Check ports
echo -e "\n${YELLOW}Port Status:${NC}"
check_port 8080 "Backend"
check_port 3000 "Frontend"

# Detailed backend health
if check_service "Backend" "$BACKEND_URL/health" > /dev/null 2>&1; then
    check_backend_health
fi

# Database check
echo -e "\n${YELLOW}Database Status:${NC}"
if [[ -f "backend/openpenpal.db" ]]; then
    db_size=$(ls -lh backend/openpenpal.db | awk '{print $5}')
    echo -e "Database file: ${GREEN}✓ Exists${NC} (Size: $db_size)"
    
    # Check if we can query the database
    user_count=$(sqlite3 backend/openpenpal.db "SELECT COUNT(*) FROM users;" 2>/dev/null || echo "Error")
    if [[ "$user_count" != "Error" ]]; then
        echo -e "User count: ${GREEN}$user_count users${NC}"
    else
        echo -e "Database query: ${RED}✗ Failed${NC}"
    fi
else
    echo -e "Database file: ${RED}✗ Not found${NC}"
fi

# Summary
echo -e "\n${BLUE}=== Summary ===${NC}"
all_healthy=true

if ! check_service "Backend" "$BACKEND_URL/health" > /dev/null 2>&1; then
    echo -e "Backend: ${RED}Needs attention${NC}"
    all_healthy=false
fi

if ! check_service "Frontend" "$FRONTEND_URL" > /dev/null 2>&1; then
    echo -e "Frontend: ${RED}Needs attention${NC}"
    all_healthy=false
fi

if $all_healthy; then
    echo -e "\n${GREEN}✅ All services are healthy!${NC}"
    exit 0
else
    echo -e "\n${RED}❌ Some services need attention${NC}"
    echo -e "\n${YELLOW}Troubleshooting tips:${NC}"
    echo "1. Check logs: tail -f logs/*.log"
    echo "2. Restart services: ./startup/quick-start.sh"
    echo "3. Check for port conflicts: lsof -i :8080 -i :3000"
    echo "4. Clear and restart: ./startup/stop-all.sh && ./startup/quick-start.sh"
    exit 1
fi