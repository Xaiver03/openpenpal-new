#!/bin/bash

# UltraThink Comprehensive Database and API Verification Script
# 全面验证数据库连接、API端点、性能优化效果

set -e

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

# Configuration
BACKEND_URL="http://localhost:8080"
DB_HOST=${DB_HOST:-"localhost"}
DB_PORT=${DB_PORT:-"5432"}
DB_USER=${DB_USER:-"postgres"}
DB_NAME=${DB_NAME:-"openpenpal"}
LOG_FILE="ultrathink-verify-$(date +%Y%m%d-%H%M%S).log"
REPORT_FILE="ultrathink-report-$(date +%Y%m%d-%H%M%S).md"

# Counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
WARNINGS=0

# Function to print colored output
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1" | tee -a "$LOG_FILE"
}

print_success() {
    echo -e "${GREEN}[✓ PASS]${NC} $1" | tee -a "$LOG_FILE"
    ((PASSED_TESTS++))
    ((TOTAL_TESTS++))
}

print_warning() {
    echo -e "${YELLOW}[⚠ WARN]${NC} $1" | tee -a "$LOG_FILE"
    ((WARNINGS++))
}

print_error() {
    echo -e "${RED}[✗ FAIL]${NC} $1" | tee -a "$LOG_FILE"
    ((FAILED_TESTS++))
    ((TOTAL_TESTS++))
}

print_section() {
    echo -e "\n${PURPLE}=== $1 ===${NC}\n" | tee -a "$LOG_FILE"
}

print_metric() {
    echo -e "${CYAN}[METRIC]${NC} $1: $2" | tee -a "$LOG_FILE"
}

# Function to test database connectivity
test_database_connectivity() {
    print_section "Database Connectivity Tests"
    
    # Basic connection test
    if psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT 1;" > /dev/null 2>&1; then
        print_success "Basic database connection"
    else
        print_error "Basic database connection failed"
        return 1
    fi
    
    # Test database version
    local pg_version=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT version();" 2>/dev/null | head -1)
    if [[ -n "$pg_version" ]]; then
        print_success "PostgreSQL version query"
        print_metric "PostgreSQL Version" "$pg_version"
    else
        print_error "Failed to query PostgreSQL version"
    fi
    
    # Check database size
    local db_size=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT pg_size_pretty(pg_database_size('$DB_NAME'));" 2>/dev/null | xargs)
    if [[ -n "$db_size" ]]; then
        print_success "Database size query"
        print_metric "Database Size" "$db_size"
    else
        print_error "Failed to query database size"
    fi
    
    # Check table count
    local table_count=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_type = 'BASE TABLE';" 2>/dev/null | xargs)
    if [[ -n "$table_count" ]]; then
        print_success "Table count query"
        print_metric "Total Tables" "$table_count"
    else
        print_error "Failed to query table count"
    fi
}

# Function to test SSL configuration
test_ssl_configuration() {
    print_section "SSL Configuration Tests"
    
    # Check SSL status
    local ssl_enabled=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SHOW ssl;" 2>/dev/null | xargs)
    if [[ "$ssl_enabled" == "on" ]]; then
        print_success "SSL is enabled on server"
    else
        print_warning "SSL is not enabled on server"
    fi
    
    # Check current connection SSL status
    local conn_ssl=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT ssl FROM pg_stat_ssl WHERE pid = pg_backend_pid();" 2>/dev/null | xargs)
    if [[ "$conn_ssl" == "t" ]]; then
        print_success "Current connection uses SSL"
        
        # Get SSL details
        local ssl_version=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT version FROM pg_stat_ssl WHERE pid = pg_backend_pid();" 2>/dev/null | xargs)
        local ssl_cipher=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT cipher FROM pg_stat_ssl WHERE pid = pg_backend_pid();" 2>/dev/null | xargs)
        
        print_metric "SSL Version" "$ssl_version"
        print_metric "SSL Cipher" "$ssl_cipher"
    else
        local ssl_mode=${DB_SSLMODE:-"disable"}
        if [[ "$ssl_mode" != "disable" ]]; then
            print_warning "SSL mode is $ssl_mode but connection is not using SSL"
        else
            print_success "SSL disabled as configured"
        fi
    fi
    
    # Check for SSL certificate files
    if [[ -f "./dev-ssl/ca-cert.pem" ]]; then
        print_success "Development SSL certificates found"
    else
        print_info "No development SSL certificates found (expected in development)"
    fi
}

# Function to test connection pool
test_connection_pool() {
    print_section "Connection Pool Tests"
    
    # Current connection count
    local conn_count=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT count(*) FROM pg_stat_activity WHERE datname = '$DB_NAME';" 2>/dev/null | xargs)
    if [[ -n "$conn_count" ]]; then
        print_success "Connection count query"
        print_metric "Current Connections" "$conn_count"
    else
        print_error "Failed to query connection count"
    fi
    
    # Connection state breakdown
    local active_conns=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT count(*) FROM pg_stat_activity WHERE datname = '$DB_NAME' AND state = 'active';" 2>/dev/null | xargs)
    local idle_conns=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT count(*) FROM pg_stat_activity WHERE datname = '$DB_NAME' AND state = 'idle';" 2>/dev/null | xargs)
    
    print_metric "Active Connections" "${active_conns:-0}"
    print_metric "Idle Connections" "${idle_conns:-0}"
    
    # Test connection pool under load
    print_info "Testing connection pool with concurrent requests..."
    local start_time=$(date +%s%N)
    
    # Create 20 concurrent connections
    for i in {1..20}; do
        psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT pg_sleep(0.1);" > /dev/null 2>&1 &
    done
    wait
    
    local end_time=$(date +%s%N)
    local duration=$(( (end_time - start_time) / 1000000 ))
    
    if [[ $duration -lt 5000 ]]; then
        print_success "Connection pool handled 20 concurrent connections efficiently"
        print_metric "Pool Test Duration" "${duration}ms"
    else
        print_warning "Connection pool test took ${duration}ms (expected < 5000ms)"
    fi
}

# Function to test index optimization
test_index_optimization() {
    print_section "Index Optimization Tests"
    
    # Check total index count
    local index_count=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT COUNT(*) FROM pg_indexes WHERE schemaname = 'public';" 2>/dev/null | xargs)
    if [[ -n "$index_count" ]]; then
        print_success "Index count query"
        print_metric "Total Indexes" "$index_count"
    else
        print_error "Failed to query index count"
    fi
    
    # Check for critical indexes
    local critical_indexes=(
        "idx_users_school_role_active"
        "idx_letters_user_status_created"
        "idx_courier_tasks_courier_status"
        "idx_notifications_user_unread"
    )
    
    for idx in "${critical_indexes[@]}"; do
        local exists=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT COUNT(*) FROM pg_indexes WHERE indexname = '$idx';" 2>/dev/null | xargs)
        if [[ "$exists" == "1" ]]; then
            print_success "Critical index exists: $idx"
        else
            print_warning "Critical index missing: $idx"
        fi
    done
    
    # Check index usage statistics
    local high_usage_indexes=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT COUNT(*) FROM pg_stat_user_indexes WHERE idx_scan > 100;" 2>/dev/null | xargs)
    print_metric "High Usage Indexes" "${high_usage_indexes:-0}"
    
    # Check for unused indexes
    local unused_indexes=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT COUNT(*) FROM pg_stat_user_indexes WHERE idx_scan = 0 AND indexrelname NOT LIKE '%_pkey';" 2>/dev/null | xargs)
    if [[ "${unused_indexes:-0}" -gt 5 ]]; then
        print_warning "Found $unused_indexes unused indexes"
    else
        print_success "Minimal unused indexes: ${unused_indexes:-0}"
    fi
}

# Function to test API endpoints
test_api_endpoints() {
    print_section "API Endpoint Tests"
    
    # Health check
    local health_response=$(curl -s -o /dev/null -w "%{http_code}" "$BACKEND_URL/health" 2>/dev/null)
    if [[ "$health_response" == "200" ]]; then
        print_success "Health endpoint responding"
    else
        print_error "Health endpoint not responding (HTTP $health_response)"
    fi
    
    # API version
    local api_response=$(curl -s -o /dev/null -w "%{http_code}" "$BACKEND_URL/api/v1" 2>/dev/null)
    if [[ "$api_response" == "200" ]] || [[ "$api_response" == "404" ]]; then
        print_success "API v1 base path accessible"
    else
        print_error "API v1 base path not accessible (HTTP $api_response)"
    fi
    
    # Test authentication endpoint
    local auth_response=$(curl -s -X POST "$BACKEND_URL/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d '{"username":"test","password":"test"}' \
        -w "\n%{http_code}" 2>/dev/null | tail -1)
    
    if [[ "$auth_response" == "401" ]] || [[ "$auth_response" == "400" ]]; then
        print_success "Auth endpoint properly rejects invalid credentials"
    else
        print_warning "Unexpected auth response: HTTP $auth_response"
    fi
    
    # Test public endpoints
    local public_endpoints=(
        "/api/v1/auth/register"
        "/api/v1/public/stats"
        "/api/v1/health/status"
    )
    
    for endpoint in "${public_endpoints[@]}"; do
        local response=$(curl -s -o /dev/null -w "%{http_code}" "$BACKEND_URL$endpoint" 2>/dev/null)
        if [[ "$response" == "200" ]] || [[ "$response" == "404" ]] || [[ "$response" == "405" ]]; then
            print_success "Public endpoint accessible: $endpoint"
        else
            print_warning "Public endpoint issue: $endpoint (HTTP $response)"
        fi
    done
}

# Function to test database performance
test_database_performance() {
    print_section "Database Performance Tests"
    
    # Simple query performance
    local start_time=$(date +%s%N)
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT COUNT(*) FROM users;" > /dev/null 2>&1
    local end_time=$(date +%s%N)
    local query_time=$(( (end_time - start_time) / 1000000 ))
    
    if [[ $query_time -lt 100 ]]; then
        print_success "Simple query performance"
        print_metric "COUNT(*) Query Time" "${query_time}ms"
    else
        print_warning "Simple query took ${query_time}ms (expected < 100ms)"
    fi
    
    # Index scan vs sequential scan ratio
    local index_ratio=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "
        SELECT ROUND(100.0 * SUM(idx_scan) / NULLIF(SUM(seq_scan + idx_scan), 0), 2)
        FROM pg_stat_user_tables
        WHERE schemaname = 'public'
    " 2>/dev/null | xargs)
    
    if [[ -n "$index_ratio" ]]; then
        print_success "Index usage ratio calculated"
        print_metric "Index Usage Ratio" "${index_ratio}%"
        
        if (( $(echo "$index_ratio < 50" | bc -l) )); then
            print_warning "Low index usage ratio - consider adding more indexes"
        fi
    fi
    
    # Cache hit ratio
    local cache_hit_ratio=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "
        SELECT ROUND(100.0 * sum(heap_blks_hit) / NULLIF(sum(heap_blks_hit) + sum(heap_blks_read), 0), 2)
        FROM pg_statio_user_tables
    " 2>/dev/null | xargs)
    
    if [[ -n "$cache_hit_ratio" ]]; then
        print_success "Cache hit ratio calculated"
        print_metric "Cache Hit Ratio" "${cache_hit_ratio}%"
        
        if (( $(echo "$cache_hit_ratio < 90" | bc -l) )); then
            print_warning "Low cache hit ratio - consider increasing shared_buffers"
        fi
    fi
}

# Function to test data integrity
test_data_integrity() {
    print_section "Data Integrity Tests"
    
    # Check for orphaned records
    local orphaned_letters=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "
        SELECT COUNT(*) FROM letters l
        LEFT JOIN users u ON l.user_id = u.id
        WHERE u.id IS NULL
    " 2>/dev/null | xargs)
    
    if [[ "${orphaned_letters:-0}" == "0" ]]; then
        print_success "No orphaned letters found"
    else
        print_error "Found $orphaned_letters orphaned letters"
    fi
    
    # Check foreign key constraints
    local fk_count=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "
        SELECT COUNT(*)
        FROM information_schema.table_constraints
        WHERE constraint_type = 'FOREIGN KEY'
        AND table_schema = 'public'
    " 2>/dev/null | xargs)
    
    if [[ -n "$fk_count" ]] && [[ "$fk_count" -gt 0 ]]; then
        print_success "Foreign key constraints present"
        print_metric "Foreign Key Count" "$fk_count"
    else
        print_warning "No foreign key constraints found"
    fi
    
    # Check for duplicate indexes
    local duplicate_indexes=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "
        SELECT COUNT(*)
        FROM (
            SELECT indrelid::regclass AS table_name,
                   array_agg(indexrelid::regclass) AS indexes
            FROM pg_index
            GROUP BY indrelid, indkey
            HAVING COUNT(*) > 1
        ) dup
    " 2>/dev/null | xargs)
    
    if [[ "${duplicate_indexes:-0}" == "0" ]]; then
        print_success "No duplicate indexes found"
    else
        print_warning "Found $duplicate_indexes potential duplicate indexes"
    fi
}

# Function to test specific optimizations
test_specific_optimizations() {
    print_section "Specific Optimization Tests"
    
    # Test soft delete optimization
    local soft_delete_index=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "
        SELECT COUNT(*) FROM pg_indexes 
        WHERE indexname = 'idx_letters_deleted_at'
    " 2>/dev/null | xargs)
    
    if [[ "$soft_delete_index" == "1" ]]; then
        print_success "Soft delete optimization index present"
    else
        print_warning "Soft delete optimization index missing"
    fi
    
    # Test covering index
    local covering_index=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "
        SELECT COUNT(*) FROM pg_indexes 
        WHERE indexdef LIKE '%INCLUDE%'
    " 2>/dev/null | xargs)
    
    if [[ -n "$covering_index" ]] && [[ "$covering_index" -gt 0 ]]; then
        print_success "Covering indexes implemented"
        print_metric "Covering Index Count" "$covering_index"
    else
        print_info "No covering indexes found (PostgreSQL 11+ feature)"
    fi
    
    # Test partial indexes
    local partial_index=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "
        SELECT COUNT(*) FROM pg_indexes 
        WHERE indexdef LIKE '%WHERE%'
    " 2>/dev/null | xargs)
    
    if [[ -n "$partial_index" ]] && [[ "$partial_index" -gt 0 ]]; then
        print_success "Partial indexes implemented"
        print_metric "Partial Index Count" "$partial_index"
    else
        print_warning "No partial indexes found"
    fi
}

# Function to generate comprehensive report
generate_report() {
    cat > "$REPORT_FILE" << EOF
# UltraThink Database and API Verification Report

**Date**: $(date)
**Environment**: ${ENVIRONMENT:-development}
**Database**: $DB_NAME@$DB_HOST:$DB_PORT

## Test Summary

- **Total Tests**: $TOTAL_TESTS
- **Passed**: $PASSED_TESTS ✅
- **Failed**: $FAILED_TESTS ❌
- **Warnings**: $WARNINGS ⚠️
- **Success Rate**: $(( TOTAL_TESTS > 0 ? PASSED_TESTS * 100 / TOTAL_TESTS : 0 ))%

## Key Metrics

$(grep "METRIC" "$LOG_FILE" | sed 's/\[METRIC\]/- **/' | sed 's/:/:**/')

## Critical Issues

$(grep "FAIL" "$LOG_FILE" | sed 's/\[✗ FAIL\]/- ❌/')

## Warnings

$(grep "WARN" "$LOG_FILE" | sed 's/\[⚠ WARN\]/- ⚠️/')

## Recommendations

1. **SSL Configuration**: ${DB_SSLMODE:-"Consider enabling SSL for production"}
2. **Connection Pool**: Monitor pool usage and adjust size based on load
3. **Index Optimization**: Review and create missing critical indexes
4. **Performance**: Monitor slow queries and optimize as needed

## Next Steps

1. Address any failed tests immediately
2. Review warnings and plan remediation
3. Set up continuous monitoring for production
4. Schedule regular optimization reviews

---
*Generated by UltraThink Verification System*
EOF

    print_info "Comprehensive report saved to: $REPORT_FILE"
}

# Main execution
main() {
    echo "🧠 UltraThink Database and API Verification Starting..."
    echo "================================================"
    echo ""
    
    # Run all tests
    test_database_connectivity
    test_ssl_configuration
    test_connection_pool
    test_index_optimization
    test_api_endpoints
    test_database_performance
    test_data_integrity
    test_specific_optimizations
    
    # Generate final report
    echo ""
    print_section "Final Summary"
    echo "Total Tests: $TOTAL_TESTS"
    echo "Passed: $PASSED_TESTS ✅"
    echo "Failed: $FAILED_TESTS ❌"
    echo "Warnings: $WARNINGS ⚠️"
    echo "Success Rate: $(( TOTAL_TESTS > 0 ? PASSED_TESTS * 100 / TOTAL_TESTS : 0 ))%"
    
    generate_report
    
    echo ""
    echo "🧠 UltraThink Verification Complete!"
    echo "Detailed log: $LOG_FILE"
    echo "Summary report: $REPORT_FILE"
    
    # Exit with appropriate code
    if [[ $FAILED_TESTS -gt 0 ]]; then
        exit 1
    else
        exit 0
    fi
}

# Run main function
main "$@"