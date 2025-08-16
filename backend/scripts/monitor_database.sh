#!/bin/bash

# OpenPenPal Database Performance Monitor Runner
# =============================================

# Database connection parameters
DB_HOST="localhost"
DB_PORT="5432"
DB_NAME="openpenpal"
DB_USER="openpenpal_user"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Output directory
OUTPUT_DIR="monitoring_reports"
mkdir -p "$OUTPUT_DIR"

# Generate timestamp
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
REPORT_FILE="$OUTPUT_DIR/performance_report_$TIMESTAMP.txt"

# Function to print header
print_header() {
    echo -e "${BLUE}╔════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║       OpenPenPal Database Performance Monitor              ║${NC}"
    echo -e "${BLUE}╚════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "${GREEN}Report Time:${NC} $(date)"
    echo -e "${GREEN}Database:${NC} $DB_NAME @ $DB_HOST:$DB_PORT"
    echo ""
}

# Function to check prerequisites
check_prerequisites() {
    echo -n "Checking database connection... "
    if psql "host=$DB_HOST port=$DB_PORT user=$DB_USER dbname=$DB_NAME sslmode=disable" -c '\q' 2>/dev/null; then
        echo -e "${GREEN}✓ Connected${NC}"
    else
        echo -e "${RED}✗ Failed${NC}"
        echo "Cannot connect to database. Please check your connection settings."
        exit 1
    fi
    
    echo -n "Checking pg_stat_statements extension... "
    if psql "host=$DB_HOST port=$DB_PORT user=$DB_USER dbname=$DB_NAME sslmode=disable" -c "SELECT 1 FROM pg_extension WHERE extname = 'pg_stat_statements';" | grep -q 1; then
        echo -e "${GREEN}✓ Installed${NC}"
    else
        echo -e "${YELLOW}⚠ Not installed${NC}"
        echo "Some performance metrics will be unavailable."
        echo "To install: CREATE EXTENSION IF NOT EXISTS pg_stat_statements;"
    fi
}

# Function to run monitoring
run_monitoring() {
    echo ""
    echo -e "${BLUE}Running performance analysis...${NC}"
    echo ""
    
    # Run the monitoring script and save to file
    psql "host=$DB_HOST port=$DB_PORT user=$DB_USER dbname=$DB_NAME sslmode=disable" \
        -f performance_monitor.sql > "$REPORT_FILE" 2>&1
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ Performance report generated successfully!${NC}"
        echo -e "${GREEN}📄 Report saved to:${NC} $REPORT_FILE"
    else
        echo -e "${RED}❌ Error generating performance report${NC}"
        echo "Check $REPORT_FILE for error details"
        exit 1
    fi
}

# Function to show summary
show_summary() {
    echo ""
    echo -e "${BLUE}Quick Summary:${NC}"
    echo "=============="
    
    # Extract key metrics from report
    if [ -f "$REPORT_FILE" ]; then
        # Database size
        DB_SIZE=$(grep -A 2 "Database Size:" "$REPORT_FILE" | grep "total_size" -A 1 | tail -1 | awk -F'|' '{print $3}' | xargs)
        if [ ! -z "$DB_SIZE" ]; then
            echo -e "Database Size: ${YELLOW}$DB_SIZE${NC}"
        fi
        
        # Active connections
        CONNECTIONS=$(grep -A 3 "Connection Statistics:" "$REPORT_FILE" | grep "openpenpal" | awk -F'|' '{print $3 "/" $4}' | xargs)
        if [ ! -z "$CONNECTIONS" ]; then
            echo -e "Connections: ${YELLOW}$CONNECTIONS${NC}"
        fi
        
        # Cache hit ratio
        CACHE_HIT=$(grep -A 4 "Cache Hit Ratios:" "$REPORT_FILE" | grep "Database" | awk -F'|' '{print $5}' | xargs)
        if [ ! -z "$CACHE_HIT" ]; then
            echo -e "Cache Hit Ratio: ${YELLOW}${CACHE_HIT}%${NC}"
        fi
        
        # Tables needing vacuum
        VACUUM_NEEDED=$(grep -c "NEEDS VACUUM" "$REPORT_FILE" 2>/dev/null || echo "0")
        if [ "$VACUUM_NEEDED" -gt 0 ]; then
            echo -e "Tables Needing Vacuum: ${RED}$VACUUM_NEEDED${NC}"
        else
            echo -e "Tables Needing Vacuum: ${GREEN}0${NC}"
        fi
    fi
}

# Function to create HTML report
create_html_report() {
    HTML_FILE="$OUTPUT_DIR/performance_report_$TIMESTAMP.html"
    
    cat > "$HTML_FILE" << 'EOF'
<!DOCTYPE html>
<html>
<head>
    <title>OpenPenPal Database Performance Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background-color: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; background-color: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        h1 { color: #333; border-bottom: 2px solid #4CAF50; padding-bottom: 10px; }
        h2 { color: #555; margin-top: 30px; }
        pre { background-color: #f8f8f8; padding: 15px; border-radius: 4px; overflow-x: auto; }
        .metric { display: inline-block; margin: 10px; padding: 15px; background-color: #e3f2fd; border-radius: 4px; }
        .warning { background-color: #fff3cd; }
        .error { background-color: #f8d7da; }
        .success { background-color: #d4edda; }
    </style>
</head>
<body>
    <div class="container">
        <h1>OpenPenPal Database Performance Report</h1>
        <p>Generated: TIMESTAMP_PLACEHOLDER</p>
        <pre>
EOF
    
    # Add report content
    cat "$REPORT_FILE" >> "$HTML_FILE"
    
    cat >> "$HTML_FILE" << 'EOF'
        </pre>
    </div>
</body>
</html>
EOF
    
    # Replace timestamp
    sed -i.bak "s/TIMESTAMP_PLACEHOLDER/$(date)/" "$HTML_FILE" && rm "$HTML_FILE.bak"
    
    echo -e "${GREEN}📊 HTML report saved to:${NC} $HTML_FILE"
}

# Function to show usage
show_usage() {
    echo "Usage: $0 [options]"
    echo ""
    echo "Options:"
    echo "  -h, --help     Show this help message"
    echo "  -w, --watch    Run monitoring continuously (every 5 minutes)"
    echo "  -i, --interval Set watch interval in seconds (default: 300)"
    echo "  -o, --html     Generate HTML report in addition to text"
    echo ""
}

# Parse command line arguments
WATCH_MODE=false
INTERVAL=300
HTML_OUTPUT=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_usage
            exit 0
            ;;
        -w|--watch)
            WATCH_MODE=true
            shift
            ;;
        -i|--interval)
            INTERVAL="$2"
            shift 2
            ;;
        -o|--html)
            HTML_OUTPUT=true
            shift
            ;;
        *)
            echo "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Main execution
main() {
    print_header
    check_prerequisites
    
    if [ "$WATCH_MODE" = true ]; then
        echo -e "${YELLOW}Running in watch mode (interval: ${INTERVAL}s)${NC}"
        echo "Press Ctrl+C to stop"
        echo ""
        
        while true; do
            run_monitoring
            show_summary
            
            if [ "$HTML_OUTPUT" = true ]; then
                create_html_report
            fi
            
            echo ""
            echo -e "${BLUE}Next run in ${INTERVAL} seconds...${NC}"
            sleep "$INTERVAL"
            
            # Generate new timestamp for next run
            TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
            REPORT_FILE="$OUTPUT_DIR/performance_report_$TIMESTAMP.txt"
        done
    else
        run_monitoring
        show_summary
        
        if [ "$HTML_OUTPUT" = true ]; then
            create_html_report
        fi
    fi
    
    echo ""
    echo -e "${GREEN}✨ Monitoring complete!${NC}"
}

# Run main function
main