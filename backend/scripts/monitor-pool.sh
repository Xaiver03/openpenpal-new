#!/bin/bash

# Connection Pool Monitoring Script
# Monitors PostgreSQL connection pool performance and health

set -e

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
DB_HOST=${DB_HOST:-"localhost"}
DB_PORT=${DB_PORT:-"5432"}
DB_USER=${DB_USER:-"postgres"}
DB_NAME=${DB_NAME:-"openpenpal"}
INTERVAL=${MONITOR_INTERVAL:-5}
LOG_FILE=${LOG_FILE:-"pool-monitor.log"}

# Function to print colored output
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to get connection statistics
get_connection_stats() {
    local stats=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "
    SELECT 
        COUNT(*) FILTER (WHERE state = 'active') as active,
        COUNT(*) FILTER (WHERE state = 'idle') as idle,
        COUNT(*) FILTER (WHERE state = 'idle in transaction') as idle_in_transaction,
        COUNT(*) FILTER (WHERE state = 'idle in transaction (aborted)') as idle_aborted,
        COUNT(*) FILTER (WHERE wait_event_type IS NOT NULL) as waiting,
        COUNT(*) as total
    FROM pg_stat_activity
    WHERE datname = '$DB_NAME'
    AND pid != pg_backend_pid()
    " 2>/dev/null)
    
    echo "$stats"
}

# Function to get connection pool metrics
get_pool_metrics() {
    local metrics=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "
    SELECT 
        numbackends as connections,
        xact_commit as commits,
        xact_rollback as rollbacks,
        blks_read as disk_reads,
        blks_hit as cache_hits,
        tup_returned as rows_returned,
        tup_fetched as rows_fetched,
        tup_inserted as rows_inserted,
        tup_updated as rows_updated,
        tup_deleted as rows_deleted,
        conflicts,
        deadlocks
    FROM pg_stat_database
    WHERE datname = '$DB_NAME'
    " 2>/dev/null)
    
    echo "$metrics"
}

# Function to get long-running queries
get_long_queries() {
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
    SELECT 
        pid,
        now() - pg_stat_activity.query_start AS duration,
        usename,
        state,
        LEFT(query, 60) as query_snippet
    FROM pg_stat_activity
    WHERE (now() - pg_stat_activity.query_start) > interval '1 minute'
    AND state != 'idle'
    AND datname = '$DB_NAME'
    ORDER BY duration DESC
    LIMIT 5;
    " 2>/dev/null || echo "No long-running queries"
}

# Function to check connection pool health
check_pool_health() {
    local stats=($1)
    local active=${stats[0]}
    local idle=${stats[1]}
    local total=${stats[5]}
    
    local issues=()
    
    # Check for high active connections
    if [[ $active -gt 80 ]]; then
        issues+=("High active connections: $active")
    fi
    
    # Check for connection leak (too many idle)
    if [[ $idle -gt 50 ]]; then
        issues+=("Possible connection leak: $idle idle connections")
    fi
    
    # Check for transaction issues
    local idle_in_tx=$((${stats[2]} + ${stats[3]}))
    if [[ $idle_in_tx -gt 5 ]]; then
        issues+=("Idle transactions detected: $idle_in_tx")
    fi
    
    # Check for waiting connections
    local waiting=${stats[4]}
    if [[ $waiting -gt 10 ]]; then
        issues+=("High number of waiting connections: $waiting")
    fi
    
    if [[ ${#issues[@]} -gt 0 ]]; then
        print_warning "Pool health issues detected:"
        for issue in "${issues[@]}"; do
            echo "  - $issue"
        done
        return 1
    else
        print_success "Connection pool healthy"
        return 0
    fi
}

# Function to generate recommendations
generate_recommendations() {
    local stats=($1)
    local active=${stats[0]}
    local idle=${stats[1]}
    local total=${stats[5]}
    
    echo ""
    echo "=== Recommendations ==="
    
    # Connection pool size recommendation
    local optimal_max=$((active * 2))
    if [[ $optimal_max -lt 20 ]]; then
        optimal_max=20
    fi
    echo "Recommended MaxOpenConns: $optimal_max"
    
    # Idle connection recommendation
    local optimal_idle=$((optimal_max / 3))
    echo "Recommended MaxIdleConns: $optimal_idle"
    
    # Lifetime recommendation based on activity
    if [[ $active -gt 50 ]]; then
        echo "Recommended ConnMaxLifetime: 5 minutes (high traffic)"
    elif [[ $active -gt 20 ]]; then
        echo "Recommended ConnMaxLifetime: 15 minutes (medium traffic)"
    else
        echo "Recommended ConnMaxLifetime: 30 minutes (low traffic)"
    fi
}

# Function to monitor in real-time
monitor_realtime() {
    print_info "Starting real-time connection pool monitoring (Ctrl+C to stop)"
    print_info "Interval: ${INTERVAL}s"
    echo ""
    
    local iteration=0
    while true; do
        clear
        echo "=== PostgreSQL Connection Pool Monitor ==="
        echo "Time: $(date)"
        echo "Database: $DB_NAME@$DB_HOST:$DB_PORT"
        echo ""
        
        # Get connection statistics
        local stats=$(get_connection_stats)
        local stats_array=($stats)
        
        echo "Connection Statistics:"
        echo "  Active:              ${stats_array[0]}"
        echo "  Idle:                ${stats_array[1]}"
        echo "  Idle in Transaction: ${stats_array[2]}"
        echo "  Idle (Aborted):      ${stats_array[3]}"
        echo "  Waiting:             ${stats_array[4]}"
        echo "  Total:               ${stats_array[5]}"
        echo ""
        
        # Check health
        check_pool_health "$stats"
        
        # Get pool metrics every 5 iterations
        if [[ $((iteration % 5)) -eq 0 ]]; then
            echo ""
            echo "Database Metrics:"
            local metrics=$(get_pool_metrics)
            local metrics_array=($metrics)
            echo "  Commits:    ${metrics_array[1]}"
            echo "  Rollbacks:  ${metrics_array[2]}"
            echo "  Cache Hit:  $(awk "BEGIN {printf \"%.2f%%\", ${metrics_array[4]}*100/(${metrics_array[3]}+${metrics_array[4]}+0.001)}")"
            echo "  Deadlocks:  ${metrics_array[11]}"
        fi
        
        # Show long queries every 10 iterations
        if [[ $((iteration % 10)) -eq 0 ]]; then
            echo ""
            echo "Long Running Queries:"
            get_long_queries
        fi
        
        # Generate recommendations every 20 iterations
        if [[ $((iteration % 20)) -eq 0 ]]; then
            generate_recommendations "$stats"
        fi
        
        # Log to file
        echo "[$(date)] Active: ${stats_array[0]}, Idle: ${stats_array[1]}, Total: ${stats_array[5]}" >> "$LOG_FILE"
        
        ((iteration++))
        sleep "$INTERVAL"
    done
}

# Function to analyze historical data
analyze_historical() {
    if [[ ! -f "$LOG_FILE" ]]; then
        print_error "No historical data found in $LOG_FILE"
        return 1
    fi
    
    print_info "Analyzing historical connection pool data..."
    
    # Extract and analyze patterns
    echo ""
    echo "=== Historical Analysis ==="
    
    # Peak connections
    echo "Peak Connections:"
    grep -o "Total: [0-9]*" "$LOG_FILE" | awk -F': ' '{print $2}' | sort -nr | head -5 | while read count; do
        echo "  $count connections"
    done
    
    # Average connections
    local avg_total=$(grep -o "Total: [0-9]*" "$LOG_FILE" | awk -F': ' '{sum+=$2; count++} END {print int(sum/count)}')
    echo ""
    echo "Average Total Connections: $avg_total"
    
    # Active connection patterns
    local avg_active=$(grep -o "Active: [0-9]*" "$LOG_FILE" | awk -F': ' '{sum+=$2; count++} END {print int(sum/count)}')
    echo "Average Active Connections: $avg_active"
    
    # Time-based analysis
    echo ""
    echo "Hourly Patterns:"
    grep -E "^\[[0-9]{4}-[0-9]{2}-[0-9]{2} [0-9]{2}:" "$LOG_FILE" | \
        awk -F'[ :]' '{print $2":00"}' | sort | uniq -c | sort -k2
}

# Function to export metrics
export_metrics() {
    local output_file="${1:-pool-metrics-$(date +%Y%m%d-%H%M%S).csv}"
    
    print_info "Exporting connection pool metrics to $output_file"
    
    echo "timestamp,active,idle,idle_in_transaction,waiting,total" > "$output_file"
    
    # Collect metrics for 1 minute
    local end_time=$(($(date +%s) + 60))
    while [[ $(date +%s) -lt $end_time ]]; do
        local stats=$(get_connection_stats)
        local stats_array=($stats)
        echo "$(date '+%Y-%m-%d %H:%M:%S'),${stats_array[0]},${stats_array[1]},${stats_array[2]},${stats_array[4]},${stats_array[5]}" >> "$output_file"
        sleep 1
    done
    
    print_success "Metrics exported to $output_file"
}

# Function to test connection pool
test_pool() {
    print_info "Testing connection pool performance..."
    
    # Create test function
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
    CREATE OR REPLACE FUNCTION pool_test() RETURNS void AS \$\$
    BEGIN
        PERFORM pg_sleep(0.1);
    END;
    \$\$ LANGUAGE plpgsql;
    " 2>/dev/null
    
    # Run concurrent connections
    print_info "Running 50 concurrent connections..."
    for i in {1..50}; do
        psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT pool_test();" &
    done
    
    # Monitor during test
    sleep 2
    local stats=$(get_connection_stats)
    local stats_array=($stats)
    
    echo "During test:"
    echo "  Active connections: ${stats_array[0]}"
    echo "  Total connections: ${stats_array[5]}"
    
    # Wait for completion
    wait
    
    # Check after test
    sleep 2
    stats=$(get_connection_stats)
    stats_array=($stats)
    
    echo "After test:"
    echo "  Active connections: ${stats_array[0]}"
    echo "  Total connections: ${stats_array[5]}"
    
    # Cleanup
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "DROP FUNCTION IF EXISTS pool_test();" 2>/dev/null
    
    print_success "Connection pool test completed"
}

# Main menu
show_usage() {
    echo "PostgreSQL Connection Pool Monitor"
    echo ""
    echo "Usage: $0 [command]"
    echo ""
    echo "Commands:"
    echo "  monitor     Real-time connection pool monitoring"
    echo "  analyze     Analyze historical connection data"
    echo "  export      Export metrics to CSV"
    echo "  test        Test connection pool performance"
    echo "  help        Show this help message"
    echo ""
    echo "Environment variables:"
    echo "  DB_HOST     Database host (default: localhost)"
    echo "  DB_PORT     Database port (default: 5432)"
    echo "  DB_USER     Database user (default: postgres)"
    echo "  DB_NAME     Database name (default: openpenpal)"
}

# Main script logic
case "${1:-monitor}" in
    monitor)
        monitor_realtime
        ;;
    analyze)
        analyze_historical
        ;;
    export)
        export_metrics "$2"
        ;;
    test)
        test_pool
        ;;
    help|*)
        show_usage
        ;;
esac