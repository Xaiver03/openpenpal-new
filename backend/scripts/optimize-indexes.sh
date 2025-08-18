#!/bin/bash

# PostgreSQL Index Optimization Script
# Manages index creation, monitoring, and maintenance

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
ENVIRONMENT=${ENVIRONMENT:-"development"}

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

# Function to check prerequisites
check_prerequisites() {
    print_info "Checking prerequisites..."
    
    # Check PostgreSQL connection
    if psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT 1" > /dev/null 2>&1; then
        print_success "Database connection successful"
    else
        print_error "Cannot connect to database"
        exit 1
    fi
    
    # Check PostgreSQL version (need 11+ for covering indexes)
    local pg_version=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT version();" | grep -oE 'PostgreSQL [0-9]+' | cut -d' ' -f2)
    if [[ $pg_version -ge 11 ]]; then
        print_success "PostgreSQL version $pg_version (supports covering indexes)"
    else
        print_warning "PostgreSQL version $pg_version (covering indexes not supported)"
    fi
}

# Function to analyze current indexes
analyze_current_indexes() {
    print_info "Analyzing current indexes..."
    
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" << EOF
-- Show table sizes and index counts
SELECT 
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS total_size,
    pg_size_pretty(pg_relation_size(schemaname||'.'||tablename)) AS table_size,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename) - pg_relation_size(schemaname||'.'||tablename)) AS indexes_size,
    (SELECT count(*) FROM pg_indexes WHERE tablename = t.tablename) as index_count
FROM pg_tables t
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC
LIMIT 20;

-- Show existing indexes
\echo '\nExisting Indexes:'
SELECT 
    tablename,
    indexname,
    pg_size_pretty(pg_relation_size(indexrelid)) as index_size
FROM pg_stat_user_indexes
WHERE schemaname = 'public'
ORDER BY pg_relation_size(indexrelid) DESC;
EOF
}

# Function to check missing indexes
check_missing_indexes() {
    print_info "Checking for missing indexes..."
    
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" << 'EOF'
-- Find tables with sequential scans but no index scans
SELECT 
    schemaname,
    tablename,
    seq_scan,
    seq_tup_read,
    idx_scan,
    CASE 
        WHEN seq_scan > 0 THEN 
            ROUND(100.0 * idx_scan / (seq_scan + idx_scan), 2)
        ELSE 100
    END as index_usage_percent
FROM pg_stat_user_tables
WHERE schemaname = 'public'
    AND seq_scan > 100
ORDER BY seq_scan DESC
LIMIT 10;

-- Find columns that might need indexes (foreign keys without indexes)
\echo '\nForeign Keys without Indexes:'
SELECT
    tc.table_name,
    kcu.column_name,
    ccu.table_name AS foreign_table_name,
    ccu.column_name AS foreign_column_name
FROM information_schema.table_constraints AS tc
JOIN information_schema.key_column_usage AS kcu
    ON tc.constraint_name = kcu.constraint_name
JOIN information_schema.constraint_column_usage AS ccu
    ON ccu.constraint_name = tc.constraint_name
WHERE tc.constraint_type = 'FOREIGN KEY'
    AND NOT EXISTS (
        SELECT 1
        FROM pg_indexes
        WHERE tablename = tc.table_name
            AND indexdef LIKE '%' || kcu.column_name || '%'
    );
EOF
}

# Function to show unused indexes
show_unused_indexes() {
    print_info "Checking for unused indexes..."
    
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" << EOF
SELECT 
    schemaname || '.' || tablename as table,
    indexname,
    pg_size_pretty(pg_relation_size(indexrelid)) as index_size,
    idx_scan as scans
FROM pg_stat_user_indexes
WHERE schemaname = 'public'
    AND idx_scan = 0
    AND indexrelname NOT LIKE '%_pkey'
    AND pg_relation_size(indexrelid) > 100000 -- Only show indexes > 100KB
ORDER BY pg_relation_size(indexrelid) DESC;
EOF
}

# Function to show index usage statistics
show_index_usage() {
    print_info "Index usage statistics..."
    
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" << EOF
-- Most used indexes
\echo 'Top 10 Most Used Indexes:'
SELECT 
    schemaname || '.' || tablename as table,
    indexname,
    idx_scan as scans,
    pg_size_pretty(pg_relation_size(indexrelid)) as size,
    ROUND(100.0 * idx_scan / NULLIF(seq_scan + idx_scan, 0), 2) as scan_ratio
FROM pg_stat_user_indexes s
JOIN pg_stat_user_tables t USING (schemaname, tablename)
WHERE s.schemaname = 'public'
    AND idx_scan > 0
ORDER BY idx_scan DESC
LIMIT 10;

-- Index hit rate
\echo '\nIndex Hit Rate by Table:'
SELECT 
    tablename,
    100 * idx_scan / NULLIF(seq_scan + idx_scan, 0) AS index_hit_rate,
    n_tup_ins + n_tup_upd + n_tup_del as write_activity,
    pg_size_pretty(pg_relation_size(schemaname||'.'||tablename)) as table_size
FROM pg_stat_user_tables
WHERE schemaname = 'public'
    AND (n_tup_ins + n_tup_upd + n_tup_del) > 0
ORDER BY (n_tup_ins + n_tup_upd + n_tup_del) DESC
LIMIT 15;
EOF
}

# Function to create critical indexes
create_critical_indexes() {
    print_info "Creating critical indexes..."
    
    # Run the Go index optimizer
    cd /Users/rocalight/同步空间/opplc/openpenpal/backend
    go run cmd/tools/optimize-indexes/main.go --mode=create --verbose
}

# Function to perform dry run
dry_run_optimization() {
    print_info "Performing dry run of index optimization..."
    
    cd /Users/rocalight/同步空间/opplc/openpenpal/backend
    go run cmd/tools/optimize-indexes/main.go --mode=create --dry-run --verbose
}

# Function to analyze query performance
analyze_query_performance() {
    print_info "Analyzing slow queries..."
    
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" << EOF
-- Enable pg_stat_statements if available
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

-- Show slowest queries
SELECT 
    SUBSTRING(query, 1, 60) as query_snippet,
    calls,
    ROUND(total_exec_time::numeric, 2) as total_ms,
    ROUND(mean_exec_time::numeric, 2) as mean_ms,
    ROUND(stddev_exec_time::numeric, 2) as stddev_ms,
    rows
FROM pg_stat_statements
WHERE query NOT LIKE '%pg_stat_statements%'
ORDER BY mean_exec_time DESC
LIMIT 10;
EOF
}

# Function to maintenance tasks
perform_maintenance() {
    print_info "Performing index maintenance..."
    
    # Reindex bloated indexes
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" << EOF
-- Find bloated indexes
WITH index_bloat AS (
    SELECT
        schemaname,
        tablename,
        indexname,
        pg_relation_size(indexrelid) as index_size,
        CASE WHEN pg_relation_size(indexrelid) > 0
            THEN (100 * (pg_relation_size(indexrelid) - pg_relation_size(indexrelid::regclass))) / pg_relation_size(indexrelid)
            ELSE 0
        END as bloat_ratio
    FROM pg_stat_user_indexes
    WHERE schemaname = 'public'
)
SELECT 
    schemaname || '.' || tablename as table,
    indexname,
    pg_size_pretty(index_size) as size,
    bloat_ratio
FROM index_bloat
WHERE index_size > 1000000 -- Only indexes > 1MB
ORDER BY bloat_ratio DESC
LIMIT 10;

-- Update table statistics
\echo '\nUpdating table statistics...'
ANALYZE;
EOF
    
    print_success "Maintenance completed"
}

# Function to generate optimization report
generate_report() {
    local report_file="index-optimization-report-$(date +%Y%m%d-%H%M%S).md"
    
    print_info "Generating optimization report..."
    
    {
        echo "# PostgreSQL Index Optimization Report"
        echo "**Date**: $(date)"
        echo "**Database**: $DB_NAME"
        echo "**Environment**: $ENVIRONMENT"
        echo ""
        
        echo "## Current Index Statistics"
        psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "
            SELECT COUNT(*) || ' total indexes' FROM pg_indexes WHERE schemaname = 'public';
        "
        
        echo "## Table Sizes"
        psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
            SELECT 
                tablename,
                pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS total_size,
                (SELECT count(*) FROM pg_indexes WHERE tablename = t.tablename) as indexes
            FROM pg_tables t
            WHERE schemaname = 'public'
            ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC
            LIMIT 10;
        "
        
        echo "## Recommendations"
        echo "1. Review unused indexes for potential removal"
        echo "2. Consider adding indexes for foreign keys without indexes"
        echo "3. Monitor slow queries and add appropriate indexes"
        echo "4. Schedule regular REINDEX for bloated indexes"
        
    } > "$report_file"
    
    print_success "Report saved to $report_file"
}

# Function to monitor real-time index usage
monitor_realtime() {
    print_info "Starting real-time index monitoring (Ctrl+C to stop)..."
    
    while true; do
        clear
        echo "=== Index Usage Monitor - $(date) ==="
        
        psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
            SELECT 
                indexrelname,
                idx_scan,
                idx_tup_read,
                idx_tup_fetch
            FROM pg_stat_user_indexes
            WHERE schemaname = 'public'
                AND idx_scan > 0
            ORDER BY idx_scan DESC
            LIMIT 10;
        "
        
        sleep 5
    done
}

# Main menu
show_menu() {
    echo "PostgreSQL Index Optimizer"
    echo ""
    echo "1. Analyze current indexes"
    echo "2. Check missing indexes"
    echo "3. Show unused indexes"
    echo "4. Show index usage statistics"
    echo "5. Dry run optimization"
    echo "6. Create critical indexes"
    echo "7. Analyze query performance"
    echo "8. Perform maintenance"
    echo "9. Generate report"
    echo "10. Monitor real-time usage"
    echo "0. Exit"
    echo ""
    read -p "Select option: " choice
    
    case $choice in
        1) analyze_current_indexes ;;
        2) check_missing_indexes ;;
        3) show_unused_indexes ;;
        4) show_index_usage ;;
        5) dry_run_optimization ;;
        6) create_critical_indexes ;;
        7) analyze_query_performance ;;
        8) perform_maintenance ;;
        9) generate_report ;;
        10) monitor_realtime ;;
        0) exit 0 ;;
        *) print_error "Invalid option"; show_menu ;;
    esac
    
    echo ""
    read -p "Press Enter to continue..."
    show_menu
}

# Command line interface
case "${1:-menu}" in
    analyze)
        check_prerequisites
        analyze_current_indexes
        ;;
    missing)
        check_prerequisites
        check_missing_indexes
        ;;
    unused)
        check_prerequisites
        show_unused_indexes
        ;;
    usage)
        check_prerequisites
        show_index_usage
        ;;
    create)
        check_prerequisites
        create_critical_indexes
        ;;
    dryrun)
        check_prerequisites
        dry_run_optimization
        ;;
    performance)
        check_prerequisites
        analyze_query_performance
        ;;
    maintenance)
        check_prerequisites
        perform_maintenance
        ;;
    report)
        check_prerequisites
        generate_report
        ;;
    monitor)
        check_prerequisites
        monitor_realtime
        ;;
    menu|*)
        check_prerequisites
        show_menu
        ;;
esac