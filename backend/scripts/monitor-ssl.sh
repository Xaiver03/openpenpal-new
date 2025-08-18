#!/bin/bash

# SSL Connection Monitoring Script
# Monitors PostgreSQL SSL connections and certificate health

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
SSL_DIR=${SSL_DIR:-"/etc/ssl/postgresql"}

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

# Function to check SSL connection status
check_ssl_status() {
    print_info "Checking SSL connection status..."
    
    # SQL query to check SSL connections
    local sql_query="
    SELECT 
        datname as database,
        usename as username,
        client_addr,
        ssl,
        version as ssl_version,
        cipher as ssl_cipher,
        bits as ssl_bits
    FROM pg_stat_ssl
    JOIN pg_stat_activity USING(pid)
    WHERE pid != pg_backend_pid()
    ORDER BY datname, usename;
    "
    
    # Execute query
    if command -v psql > /dev/null; then
        psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "$sql_query" 2>/dev/null || {
            print_error "Failed to query SSL status"
            return 1
        }
    else
        print_error "psql not found"
        return 1
    fi
}

# Function to count SSL vs non-SSL connections
count_ssl_connections() {
    print_info "Counting SSL connections..."
    
    local sql_query="
    SELECT 
        CASE WHEN ssl THEN 'SSL' ELSE 'Non-SSL' END as connection_type,
        COUNT(*) as count
    FROM pg_stat_ssl
    JOIN pg_stat_activity USING(pid)
    WHERE pid != pg_backend_pid()
    GROUP BY ssl
    ORDER BY ssl DESC;
    "
    
    if command -v psql > /dev/null; then
        psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "$sql_query" 2>/dev/null || {
            print_error "Failed to count connections"
            return 1
        }
    fi
}

# Function to check certificate expiry
check_certificate_expiry() {
    print_info "Checking certificate expiry..."
    
    local cert_files=(
        "$SSL_DIR/ca-cert.pem"
        "$SSL_DIR/client-cert.pem"
        "$SSL_DIR/server-cert.pem"
        "./dev-ssl/ca-cert.pem"
        "./dev-ssl/client-cert.pem"
        "./dev-ssl/server-cert.pem"
    )
    
    for cert_file in "${cert_files[@]}"; do
        if [[ -f "$cert_file" ]]; then
            print_info "Checking $cert_file"
            
            # Get certificate dates
            local not_after=$(openssl x509 -in "$cert_file" -noout -enddate 2>/dev/null | cut -d= -f2)
            if [[ -n "$not_after" ]]; then
                local expiry_timestamp=$(date -d "$not_after" +%s 2>/dev/null || date -j -f "%b %d %T %Y %Z" "$not_after" +%s)
                local current_timestamp=$(date +%s)
                local days_until_expiry=$(( ($expiry_timestamp - $current_timestamp) / 86400 ))
                
                if [[ $days_until_expiry -lt 0 ]]; then
                    print_error "  Certificate EXPIRED!"
                elif [[ $days_until_expiry -lt 7 ]]; then
                    print_error "  Certificate expires in $days_until_expiry days!"
                elif [[ $days_until_expiry -lt 30 ]]; then
                    print_warning "  Certificate expires in $days_until_expiry days"
                else
                    print_success "  Certificate valid for $days_until_expiry days"
                fi
                
                # Show certificate details
                local subject=$(openssl x509 -in "$cert_file" -noout -subject | cut -d= -f2-)
                print_info "  Subject: $subject"
                print_info "  Expires: $not_after"
            fi
            echo ""
        fi
    done
}

# Function to test SSL performance
test_ssl_performance() {
    print_info "Testing SSL connection performance..."
    
    # Test connection time with SSL
    local start_time=$(date +%s%N)
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT 1" > /dev/null 2>&1
    local end_time=$(date +%s%N)
    local ssl_time=$(( ($end_time - $start_time) / 1000000 ))
    
    print_info "SSL connection time: ${ssl_time}ms"
    
    # Get SSL cipher and protocol info
    local ssl_info=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "
    SELECT 
        'Protocol: ' || version || ', Cipher: ' || cipher || ', Bits: ' || bits
    FROM pg_stat_ssl 
    WHERE pid = pg_backend_pid()
    " 2>/dev/null | xargs)
    
    if [[ -n "$ssl_info" ]]; then
        print_info "SSL Info: $ssl_info"
    fi
}

# Function to generate SSL report
generate_ssl_report() {
    local report_file="${1:-ssl-report-$(date +%Y%m%d-%H%M%S).txt}"
    
    print_info "Generating SSL report to $report_file"
    
    {
        echo "PostgreSQL SSL Status Report"
        echo "Generated: $(date)"
        echo "=================================="
        echo ""
        
        echo "Environment:"
        echo "  Host: $DB_HOST"
        echo "  Port: $DB_PORT"
        echo "  Database: $DB_NAME"
        echo "  SSL Mode: ${DB_SSLMODE:-unknown}"
        echo ""
        
        echo "Connection Statistics:"
        count_ssl_connections
        echo ""
        
        echo "Active SSL Connections:"
        check_ssl_status
        echo ""
        
        echo "Certificate Status:"
        check_certificate_expiry
        echo ""
        
        echo "Performance Metrics:"
        test_ssl_performance
        
    } > "$report_file" 2>&1
    
    print_success "Report saved to $report_file"
}

# Function to monitor SSL in real-time
monitor_realtime() {
    print_info "Starting real-time SSL monitoring (Press Ctrl+C to stop)..."
    
    while true; do
        clear
        echo "=== PostgreSQL SSL Monitor ==="
        echo "Time: $(date)"
        echo ""
        
        count_ssl_connections
        echo ""
        
        check_ssl_status
        
        sleep 5
    done
}

# Function to check SSL requirements
check_ssl_requirements() {
    print_info "Checking SSL requirements..."
    
    # Check if PostgreSQL supports SSL
    local ssl_compiled=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SHOW ssl_compiled" 2>/dev/null | xargs)
    if [[ "$ssl_compiled" == "on" ]]; then
        print_success "PostgreSQL compiled with SSL support"
    else
        print_error "PostgreSQL not compiled with SSL support"
    fi
    
    # Check if SSL is enabled
    local ssl_enabled=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SHOW ssl" 2>/dev/null | xargs)
    if [[ "$ssl_enabled" == "on" ]]; then
        print_success "SSL is enabled on the server"
    else
        print_warning "SSL is not enabled on the server"
    fi
    
    # Check SSL library
    local ssl_library=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SHOW ssl_library" 2>/dev/null | xargs)
    if [[ -n "$ssl_library" ]]; then
        print_info "SSL Library: $ssl_library"
    fi
}

# Function to show alerts
check_ssl_alerts() {
    print_info "Checking for SSL alerts..."
    
    local has_alerts=false
    
    # Check for non-SSL connections in production
    if [[ "$ENVIRONMENT" == "production" ]]; then
        local non_ssl_count=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "
        SELECT COUNT(*) 
        FROM pg_stat_ssl 
        JOIN pg_stat_activity USING(pid) 
        WHERE NOT ssl AND pid != pg_backend_pid()
        " 2>/dev/null | xargs)
        
        if [[ -n "$non_ssl_count" ]] && [[ "$non_ssl_count" -gt 0 ]]; then
            print_warning "ALERT: $non_ssl_count non-SSL connections detected in production!"
            has_alerts=true
        fi
    fi
    
    # Check for expiring certificates
    local cert_files=("$SSL_DIR/ca-cert.pem" "$SSL_DIR/client-cert.pem")
    for cert_file in "${cert_files[@]}"; do
        if [[ -f "$cert_file" ]]; then
            local not_after=$(openssl x509 -in "$cert_file" -noout -enddate 2>/dev/null | cut -d= -f2)
            if [[ -n "$not_after" ]]; then
                local expiry_timestamp=$(date -d "$not_after" +%s 2>/dev/null || date -j -f "%b %d %T %Y %Z" "$not_after" +%s)
                local current_timestamp=$(date +%s)
                local days_until_expiry=$(( ($expiry_timestamp - $current_timestamp) / 86400 ))
                
                if [[ $days_until_expiry -lt 30 ]]; then
                    print_warning "ALERT: Certificate $cert_file expires in $days_until_expiry days!"
                    has_alerts=true
                fi
            fi
        fi
    done
    
    if [[ "$has_alerts" == "false" ]]; then
        print_success "No SSL alerts found"
    fi
}

# Main menu
show_usage() {
    echo "PostgreSQL SSL Monitoring Script"
    echo ""
    echo "Usage: $0 [command]"
    echo ""
    echo "Commands:"
    echo "  status        Show current SSL connection status"
    echo "  count         Count SSL vs non-SSL connections"
    echo "  expiry        Check certificate expiry dates"
    echo "  performance   Test SSL connection performance"
    echo "  requirements  Check SSL requirements and configuration"
    echo "  alerts        Check for SSL-related alerts"
    echo "  report        Generate comprehensive SSL report"
    echo "  monitor       Real-time SSL monitoring"
    echo "  help          Show this help message"
}

# Main script logic
case "${1:-help}" in
    status)
        check_ssl_status
        ;;
    count)
        count_ssl_connections
        ;;
    expiry)
        check_certificate_expiry
        ;;
    performance)
        test_ssl_performance
        ;;
    requirements)
        check_ssl_requirements
        ;;
    alerts)
        check_ssl_alerts
        ;;
    report)
        generate_ssl_report "$2"
        ;;
    monitor)
        monitor_realtime
        ;;
    help|*)
        show_usage
        ;;
esac