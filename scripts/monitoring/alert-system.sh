#!/bin/bash

# OpenPenPal Automated Monitoring Alert System
# This script performs comprehensive monitoring and sends alerts when issues are detected

set -e

# Configuration
ALERT_LOG_FILE="../logs/monitoring-alerts.log"
DB_NAME="${DB_NAME:-openpenpal}"
DB_USER="${DB_USER:-$(whoami)}"
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"

# Alert thresholds
CONNECTION_THRESHOLD=25        # Alert if total connections > 25
UTILIZATION_THRESHOLD=80.0    # Alert if pool utilization > 80%
LONG_QUERY_THRESHOLD=30       # Alert if queries running > 30 seconds
IDLE_CONNECTION_THRESHOLD=10  # Alert if idle connections > 10 for 5+ minutes
WAITING_THREAD_THRESHOLD=1    # Alert if threads waiting for connections

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Logging function
log_alert() {
    local level=$1
    local message=$2
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo "[$timestamp] [$level] $message" >> "$ALERT_LOG_FILE"
    echo -e "${RED}[$level] $message${NC}"
}

log_info() {
    local message=$1
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo "[$timestamp] [INFO] $message" >> "$ALERT_LOG_FILE"
    echo -e "${GREEN}[INFO] $message${NC}"
}

# Database query function
execute_query() {
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "$1" 2>/dev/null || echo "Query failed"
}

# Check database connections
check_database_connections() {
    log_info "Checking database connections..."
    
    local total_connections=$(execute_query "SELECT COUNT(*) FROM pg_stat_activity WHERE datname = '$DB_NAME';")
    local active_connections=$(execute_query "SELECT COUNT(*) FROM pg_stat_activity WHERE datname = '$DB_NAME' AND state = 'active';")
    local idle_connections=$(execute_query "SELECT COUNT(*) FROM pg_stat_activity WHERE datname = '$DB_NAME' AND state = 'idle';")
    
    # Check total connections
    if [ "$total_connections" -gt "$CONNECTION_THRESHOLD" ]; then
        log_alert "CRITICAL" "High number of database connections: $total_connections (threshold: $CONNECTION_THRESHOLD)"
    fi
    
    # Check for long-running idle connections
    local old_idle=$(execute_query "SELECT COUNT(*) FROM pg_stat_activity WHERE datname = '$DB_NAME' AND state = 'idle' AND now() - state_change > interval '5 minutes';")
    if [ "$old_idle" -gt "$IDLE_CONNECTION_THRESHOLD" ]; then
        log_alert "WARNING" "Many idle connections older than 5 minutes: $old_idle (threshold: $IDLE_CONNECTION_THRESHOLD)"
    fi
    
    log_info "Database connections OK: Total=$total_connections, Active=$active_connections, Idle=$idle_connections"
}

# Check for long-running queries
check_long_running_queries() {
    log_info "Checking for long-running queries..."
    
    local long_queries=$(execute_query "SELECT COUNT(*) FROM pg_stat_activity WHERE datname = '$DB_NAME' AND state != 'idle' AND now() - query_start > interval '$LONG_QUERY_THRESHOLD seconds';")
    
    if [ "$long_queries" -gt 0 ]; then
        log_alert "WARNING" "Found $long_queries queries running longer than $LONG_QUERY_THRESHOLD seconds"
        
        # Get details of long-running queries
        execute_query "
        SELECT 
            pid,
            now() - query_start as duration,
            LEFT(query, 100) as query_preview
        FROM pg_stat_activity
        WHERE datname = '$DB_NAME'
            AND state != 'idle'
            AND now() - query_start > interval '$LONG_QUERY_THRESHOLD seconds'
        ORDER BY duration DESC;" >> "$ALERT_LOG_FILE"
    else
        log_info "No long-running queries detected"
    fi
}

# Check HikariCP pool health (for admin service)
check_hikaricp_health() {
    log_info "Checking HikariCP health..."
    
    # Try to get health status from admin service actuator endpoint
    local health_response=$(curl -s http://localhost:8003/api/admin/actuator/health/hikaricp 2>/dev/null || echo "")
    
    if [ -n "$health_response" ]; then
        # Parse health response to check for issues
        if echo "$health_response" | grep -q '"status":"DOWN"'; then
            log_alert "CRITICAL" "HikariCP pool health check failed"
            echo "$health_response" >> "$ALERT_LOG_FILE"
        elif echo "$health_response" | grep -q '"threads_waiting":[1-9]'; then
            log_alert "WARNING" "Threads waiting for database connections detected"
            echo "$health_response" >> "$ALERT_LOG_FILE"
        else
            log_info "HikariCP pool health OK"
        fi
    else
        log_alert "WARNING" "Cannot reach admin service health endpoint"
    fi
}

# Check system resources
check_system_resources() {
    log_info "Checking system resources..."
    
    # Check memory usage
    local memory_usage=$(free | grep Mem | awk '{printf "%.1f", $3/$2 * 100.0}')
    if (( $(echo "$memory_usage > 90.0" | bc -l) )); then
        log_alert "CRITICAL" "High memory usage: ${memory_usage}%"
    elif (( $(echo "$memory_usage > 80.0" | bc -l) )); then
        log_alert "WARNING" "Memory usage: ${memory_usage}%"
    fi
    
    # Check disk usage
    local disk_usage=$(df / | tail -1 | awk '{print $5}' | sed 's/%//')
    if [ "$disk_usage" -gt 90 ]; then
        log_alert "CRITICAL" "High disk usage: ${disk_usage}%"
    elif [ "$disk_usage" -gt 80 ]; then
        log_alert "WARNING" "Disk usage: ${disk_usage}%"
    fi
    
    log_info "System resources OK: Memory=${memory_usage}%, Disk=${disk_usage}%"
}

# Check service availability
check_service_availability() {
    log_info "Checking service availability..."
    
    local services=(
        "localhost:3000:Frontend"
        "localhost:8080:Backend"
        "localhost:8001:WriteService"
        "localhost:8002:CourierService"
        "localhost:8003:AdminService"
        "localhost:8004:OCRService"
    )
    
    for service in "${services[@]}"; do
        IFS=':' read -r host port name <<< "$service"
        if nc -z "$host" "$port" 2>/dev/null; then
            log_info "$name service OK ($host:$port)"
        else
            log_alert "CRITICAL" "$name service DOWN ($host:$port)"
        fi
    done
}

# Check log file sizes
check_log_sizes() {
    log_info "Checking log file sizes..."
    
    local log_dir="../logs"
    local max_size_mb=500
    
    find "$log_dir" -name "*.log" -type f | while read -r logfile; do
        local size_mb=$(du -m "$logfile" 2>/dev/null | cut -f1)
        if [ "$size_mb" -gt "$max_size_mb" ]; then
            log_alert "WARNING" "Large log file detected: $logfile (${size_mb}MB)"
        fi
    done
}

# Generate health report
generate_health_report() {
    local report_file="../logs/health-report-$(date +%Y%m%d-%H%M%S).txt"
    
    {
        echo "OpenPenPal System Health Report"
        echo "Generated: $(date)"
        echo "=================================="
        echo
        
        echo "Database Connection Summary:"
        execute_query "SELECT state, COUNT(*) as count FROM pg_stat_activity WHERE datname = '$DB_NAME' GROUP BY state;"
        echo
        
        echo "Service Status:"
        ../startup/check-status.sh
        echo
        
        echo "Recent Alerts (last 24 hours):"
        tail -100 "$ALERT_LOG_FILE" | grep "$(date -d '1 day ago' '+%Y-%m-%d')\|$(date '+%Y-%m-%d')" || echo "No alerts"
        
    } > "$report_file"
    
    log_info "Health report generated: $report_file"
}

# Main monitoring loop
main() {
    echo -e "${BLUE}=== OpenPenPal Monitoring Alert System ===${NC}"
    echo -e "${BLUE}Started at: $(date)${NC}\n"
    
    # Create logs directory if it doesn't exist
    mkdir -p ../logs
    
    # Run all checks
    check_database_connections
    check_long_running_queries
    check_hikaricp_health
    check_system_resources
    check_service_availability
    check_log_sizes
    
    # Generate report
    generate_health_report
    
    echo -e "\n${BLUE}Monitoring completed at: $(date)${NC}"
    echo -e "${BLUE}Alert log: $ALERT_LOG_FILE${NC}"
}

# Handle script arguments
case "${1:-monitor}" in
    "monitor")
        main
        ;;
    "continuous")
        echo "Starting continuous monitoring (every 5 minutes)..."
        while true; do
            main
            sleep 300  # 5 minutes
        done
        ;;
    "report")
        generate_health_report
        ;;
    *)
        echo "Usage: $0 [monitor|continuous|report]"
        echo "  monitor     - Run monitoring once (default)"
        echo "  continuous  - Run continuous monitoring"
        echo "  report      - Generate health report only"
        exit 1
        ;;
esac