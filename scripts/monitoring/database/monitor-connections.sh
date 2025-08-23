#!/bin/bash

# Database Connection Monitoring Script for OpenPenPal
# This script monitors PostgreSQL connections and identifies potential issues

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Database configuration
DB_NAME="${DB_NAME:-openpenpal}"
DB_USER="${DB_USER:-$(whoami)}"
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"

echo -e "${BLUE}=== OpenPenPal Database Connection Monitor ===${NC}"
echo -e "${BLUE}Database: $DB_NAME | Time: $(date '+%Y-%m-%d %H:%M:%S')${NC}\n"

# Function to execute PostgreSQL query
execute_query() {
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "$1" 2>/dev/null || echo "Query failed"
}

# 1. Connection Summary
echo -e "${GREEN}1. Connection Summary:${NC}"
execute_query "
SELECT 
    state,
    COUNT(*) as count,
    MAX(now() - state_change) as max_duration
FROM pg_stat_activity 
WHERE datname = '$DB_NAME'
GROUP BY state
ORDER BY count DESC;"

echo -e "\n${GREEN}2. Connection Details by Application:${NC}"
execute_query "
SELECT 
    application_name,
    COUNT(*) as connections,
    COUNT(CASE WHEN state = 'active' THEN 1 END) as active,
    COUNT(CASE WHEN state = 'idle' THEN 1 END) as idle,
    COUNT(CASE WHEN state = 'idle in transaction' THEN 1 END) as idle_in_transaction
FROM pg_stat_activity 
WHERE datname = '$DB_NAME'
GROUP BY application_name
ORDER BY connections DESC;"

echo -e "\n${GREEN}3. Long Running Queries (> 1 minute):${NC}"
execute_query "
SELECT 
    pid,
    now() - query_start AS duration,
    state,
    LEFT(query, 80) as query_preview
FROM pg_stat_activity
WHERE datname = '$DB_NAME'
    AND state != 'idle'
    AND now() - query_start > interval '1 minute'
ORDER BY duration DESC
LIMIT 10;"

echo -e "\n${GREEN}4. Idle Connections (> 5 minutes):${NC}"
execute_query "
SELECT 
    pid,
    application_name,
    client_addr,
    now() - state_change as idle_duration
FROM pg_stat_activity
WHERE datname = '$DB_NAME'
    AND state = 'idle'
    AND now() - state_change > interval '5 minutes'
ORDER BY idle_duration DESC
LIMIT 10;"

echo -e "\n${GREEN}5. Lock Analysis:${NC}"
execute_query "
SELECT 
    blocked_locks.pid AS blocked_pid,
    blocked_activity.usename AS blocked_user,
    blocking_locks.pid AS blocking_pid,
    blocking_activity.usename AS blocking_user,
    blocked_activity.query AS blocked_statement,
    blocking_activity.query AS blocking_statement
FROM pg_catalog.pg_locks blocked_locks
JOIN pg_catalog.pg_stat_activity blocked_activity ON blocked_activity.pid = blocked_locks.pid
JOIN pg_catalog.pg_locks blocking_locks ON blocking_locks.locktype = blocked_locks.locktype
    AND blocking_locks.DATABASE IS NOT DISTINCT FROM blocked_locks.DATABASE
    AND blocking_locks.relation IS NOT DISTINCT FROM blocked_locks.relation
    AND blocking_locks.page IS NOT DISTINCT FROM blocked_locks.page
    AND blocking_locks.tuple IS NOT DISTINCT FROM blocked_locks.tuple
    AND blocking_locks.virtualxid IS NOT DISTINCT FROM blocked_locks.virtualxid
    AND blocking_locks.transactionid IS NOT DISTINCT FROM blocked_locks.transactionid
    AND blocking_locks.classid IS NOT DISTINCT FROM blocked_locks.classid
    AND blocking_locks.objid IS NOT DISTINCT FROM blocked_locks.objid
    AND blocking_locks.objsubid IS NOT DISTINCT FROM blocked_locks.objsubid
    AND blocking_locks.pid != blocked_locks.pid
JOIN pg_catalog.pg_stat_activity blocking_activity ON blocking_activity.pid = blocking_locks.pid
WHERE NOT blocked_locks.granted
LIMIT 10;"

# 6. Connection Pool Recommendations
echo -e "\n${GREEN}6. Connection Pool Analysis:${NC}"
TOTAL_CONNECTIONS=$(execute_query "SELECT COUNT(*) FROM pg_stat_activity WHERE datname = '$DB_NAME';")
MAX_CONNECTIONS=$(execute_query "SHOW max_connections;")
USAGE_PERCENT=$(awk "BEGIN {printf \"%.1f\", ($TOTAL_CONNECTIONS / $MAX_CONNECTIONS) * 100}")

echo "Total Connections: $TOTAL_CONNECTIONS"
echo "Max Connections: $MAX_CONNECTIONS"
echo "Usage: ${USAGE_PERCENT}%"

if (( $(echo "$USAGE_PERCENT > 80" | bc -l) )); then
    echo -e "${RED}WARNING: Connection usage is above 80%!${NC}"
elif (( $(echo "$USAGE_PERCENT > 60" | bc -l) )); then
    echo -e "${YELLOW}CAUTION: Connection usage is above 60%${NC}"
else
    echo -e "${GREEN}Connection usage is healthy${NC}"
fi

# 7. Create monitoring view if it doesn't exist
echo -e "\n${GREEN}7. Creating/Updating Monitoring View:${NC}"
execute_query "
CREATE OR REPLACE VIEW connection_stats AS
SELECT 
    datname,
    state,
    application_name,
    COUNT(*) as count,
    MAX(now() - state_change) as max_idle_time,
    AVG(EXTRACT(EPOCH FROM (now() - state_change))) as avg_idle_seconds
FROM pg_stat_activity
WHERE datname IS NOT NULL
GROUP BY datname, state, application_name
ORDER BY datname, count DESC;"

echo -e "${GREEN}View 'connection_stats' created/updated successfully${NC}"

# 8. Generate recommendations
echo -e "\n${BLUE}=== Recommendations ===${NC}"

# Check for idle connections
IDLE_COUNT=$(execute_query "SELECT COUNT(*) FROM pg_stat_activity WHERE datname = '$DB_NAME' AND state = 'idle' AND now() - state_change > interval '5 minutes';")
if [ "$IDLE_COUNT" -gt 5 ]; then
    echo -e "${YELLOW}• Found $IDLE_COUNT idle connections > 5 minutes. Consider reducing idle_timeout in HikariCP${NC}"
fi

# Check for long running queries
LONG_QUERIES=$(execute_query "SELECT COUNT(*) FROM pg_stat_activity WHERE datname = '$DB_NAME' AND state != 'idle' AND now() - query_start > interval '5 minutes';")
if [ "$LONG_QUERIES" -gt 0 ]; then
    echo -e "${RED}• Found $LONG_QUERIES long-running queries. Investigate and add appropriate indexes${NC}"
fi

# Check connection pool size
if [ "$TOTAL_CONNECTIONS" -lt 10 ]; then
    echo -e "${GREEN}• Connection pool size appears adequate${NC}"
elif [ "$TOTAL_CONNECTIONS" -lt 30 ]; then
    echo -e "${YELLOW}• Consider monitoring connection pool size during peak load${NC}"
else
    echo -e "${RED}• High number of connections detected. Review pool configuration${NC}"
fi

echo -e "\n${BLUE}Monitor completed at $(date '+%Y-%m-%d %H:%M:%S')${NC}"