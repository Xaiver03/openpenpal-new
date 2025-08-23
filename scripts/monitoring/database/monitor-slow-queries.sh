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