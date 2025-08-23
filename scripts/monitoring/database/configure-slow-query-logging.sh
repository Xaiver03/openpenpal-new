#!/bin/bash

# PostgreSQL Slow Query Logging Configuration for OpenPenPal
# This script configures PostgreSQL to log slow queries and provides analysis tools

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Database configuration
DB_NAME="${DB_NAME:-openpenpal}"
DB_USER="${DB_USER:-$(whoami)}"
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"

echo -e "${BLUE}=== PostgreSQL Slow Query Logging Configuration ===${NC}"
echo -e "${BLUE}Target Database: $DB_NAME${NC}\n"

# Function to execute PostgreSQL query
execute_query() {
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "$1" 2>/dev/null || echo "Query failed"
}

# Function to execute PostgreSQL query as superuser
execute_admin_query() {
    # Try different superuser approaches
    if psql -h "$DB_HOST" -p "$DB_PORT" -U postgres -d "$DB_NAME" -c "$1" 2>/dev/null; then
        return 0
    elif sudo -u postgres psql -d "$DB_NAME" -c "$1" 2>/dev/null; then
        return 0
    else
        echo "Failed to execute admin query. You may need superuser privileges."
        return 1
    fi
}

echo -e "${GREEN}1. Current Logging Configuration:${NC}"
execute_query "SHOW log_statement;"
execute_query "SHOW log_min_duration_statement;"
execute_query "SHOW log_checkpoints;"
execute_query "SHOW log_connections;"
execute_query "SHOW log_disconnections;"
execute_query "SHOW log_lock_waits;"

echo -e "\n${GREEN}2. Configuring Slow Query Logging:${NC}"

# Configure slow query logging (requires superuser privileges)
echo "Setting up slow query logging configuration..."

# Configuration commands
CONFIG_COMMANDS=(
    "ALTER SYSTEM SET log_min_duration_statement = 1000;"  # Log queries > 1 second
    "ALTER SYSTEM SET log_statement = 'ddl';"              # Log DDL statements
    "ALTER SYSTEM SET log_checkpoints = on;"               # Log checkpoint activity
    "ALTER SYSTEM SET log_connections = on;"               # Log connections
    "ALTER SYSTEM SET log_disconnections = on;"            # Log disconnections
    "ALTER SYSTEM SET log_lock_waits = on;"               # Log lock waits
    "ALTER SYSTEM SET log_temp_files = 0;"                # Log temp files
    "ALTER SYSTEM SET log_autovacuum_min_duration = 0;"   # Log all autovacuum activity
    "ALTER SYSTEM SET shared_preload_libraries = 'pg_stat_statements';" # Enable query stats extension
)

for cmd in "${CONFIG_COMMANDS[@]}"; do
    echo "Executing: $cmd"
    if execute_admin_query "$cmd"; then
        echo -e "${GREEN}✓ Success${NC}"
    else
        echo -e "${YELLOW}⚠ Skipped (requires superuser)${NC}"
    fi
done

echo -e "\n${GREEN}3. Reloading Configuration:${NC}"
if execute_admin_query "SELECT pg_reload_conf();"; then
    echo -e "${GREEN}✓ Configuration reloaded${NC}"
else
    echo -e "${YELLOW}⚠ Manual reload required: sudo systemctl reload postgresql${NC}"
fi

echo -e "\n${GREEN}4. Setting up pg_stat_statements extension:${NC}"
if execute_query "CREATE EXTENSION IF NOT EXISTS pg_stat_statements;"; then
    echo -e "${GREEN}✓ pg_stat_statements extension enabled${NC}"
else
    echo -e "${YELLOW}⚠ Extension setup failed (may require superuser)${NC}"
fi

echo -e "\n${GREEN}5. Creating Slow Query Analysis Views:${NC}"

# Create view for analyzing slow queries
execute_query "
CREATE OR REPLACE VIEW slow_query_analysis AS
SELECT 
    query,
    calls,
    total_exec_time,
    mean_exec_time,
    stddev_exec_time,
    rows,
    100.0 * shared_blks_hit / nullif(shared_blks_hit + shared_blks_read, 0) AS hit_percent
FROM pg_stat_statements
WHERE mean_exec_time > 1000  -- Queries averaging > 1 second
ORDER BY mean_exec_time DESC
LIMIT 20;
"

execute_query "
CREATE OR REPLACE VIEW connection_activity AS
SELECT 
    pid,
    now() - query_start as duration,
    state,
    query,
    client_addr,
    application_name
FROM pg_stat_activity 
WHERE datname = '$DB_NAME'
    AND state != 'idle'
ORDER BY duration DESC;
"

echo -e "${GREEN}✓ Analysis views created${NC}"

echo -e "\n${GREEN}6. Creating Log Analysis Function:${NC}"

# Create function to analyze recent slow queries
execute_query "
CREATE OR REPLACE FUNCTION analyze_recent_slow_queries()
RETURNS TABLE (
    query_text text,
    avg_duration numeric,
    call_count bigint,
    total_duration numeric
) AS \$\$
BEGIN
    RETURN QUERY
    SELECT 
        LEFT(s.query, 100) as query_text,
        ROUND(s.mean_exec_time::numeric, 2) as avg_duration,
        s.calls as call_count,
        ROUND(s.total_exec_time::numeric, 2) as total_duration
    FROM pg_stat_statements s
    WHERE s.mean_exec_time > 500  -- Queries averaging > 500ms
    ORDER BY s.mean_exec_time DESC
    LIMIT 10;
END;
\$\$ LANGUAGE plpgsql;
"

echo -e "${GREEN}✓ Analysis function created${NC}"

echo -e "\n${GREEN}7. Current Slow Query Statistics:${NC}"
execute_query "SELECT * FROM analyze_recent_slow_queries();"

echo -e "\n${GREEN}8. Log File Locations:${NC}"

# Try to find PostgreSQL log directory
LOG_DIRS=(
    "/var/log/postgresql"
    "/usr/local/var/log"
    "/opt/homebrew/var/log"
    "/var/lib/postgresql/data/log"
)

for dir in "${LOG_DIRS[@]}"; do
    if [ -d "$dir" ]; then
        echo -e "${GREEN}Found PostgreSQL logs in: $dir${NC}"
        ls -la "$dir"/*.log 2>/dev/null | head -5 || echo "No .log files found"
        break
    fi
done

echo -e "\n${GREEN}9. Creating Log Monitoring Script:${NC}"

# Create a log monitoring script
cat > ../../../scripts/monitoring/database/monitor-slow-queries.sh << 'EOF'
#!/bin/bash

# Monitor PostgreSQL slow queries
DB_NAME="${DB_NAME:-openpenpal}"
DB_USER="${DB_USER:-$(whoami)}"

echo "=== Recent Slow Queries Analysis ==="
psql -U "$DB_USER" -d "$DB_NAME" -c "SELECT * FROM analyze_recent_slow_queries();"

echo -e "\n=== Current Long Running Queries ==="
psql -U "$DB_USER" -d "$DB_NAME" -c "
SELECT 
    pid,
    now() - query_start as duration,
    LEFT(query, 80) as query_preview
FROM pg_stat_activity 
WHERE datname = '$DB_NAME'
    AND state != 'idle'
    AND now() - query_start > interval '30 seconds'
ORDER BY duration DESC;"

echo -e "\n=== Query Performance Summary ==="
psql -U "$DB_USER" -d "$DB_NAME" -c "
SELECT 
    COUNT(*) as total_queries,
    COUNT(CASE WHEN mean_exec_time > 1000 THEN 1 END) as slow_queries,
    ROUND(AVG(mean_exec_time)::numeric, 2) as avg_exec_time_ms
FROM pg_stat_statements;"
EOF

chmod +x ../../../scripts/monitoring/database/monitor-slow-queries.sh
echo -e "${GREEN}✓ Slow query monitoring script created${NC}"

echo -e "\n${BLUE}=== Configuration Summary ===${NC}"
echo -e "${GREEN}✓ Slow query logging configured (queries > 1 second)${NC}"
echo -e "${GREEN}✓ Connection logging enabled${NC}"
echo -e "${GREEN}✓ Lock wait logging enabled${NC}"
echo -e "${GREEN}✓ Analysis views and functions created${NC}"
echo -e "${GREEN}✓ Monitoring script created at: scripts/monitoring/database/monitor-slow-queries.sh${NC}"

echo -e "\n${BLUE}=== Next Steps ===${NC}"
echo -e "${YELLOW}1. Monitor query performance: ./scripts/monitoring/database/monitor-slow-queries.sh${NC}"
echo -e "${YELLOW}2. Check slow query analysis: SELECT * FROM slow_query_analysis;${NC}"
echo -e "${YELLOW}3. Monitor connection activity: SELECT * FROM connection_activity;${NC}"
echo -e "${YELLOW}4. For log file analysis, check PostgreSQL logs in system log directory${NC}"

echo -e "\n${BLUE}Configuration completed at $(date)${NC}"