-- OpenPenPal Database Performance Monitor
-- Comprehensive monitoring script for PostgreSQL
-- ========================================

-- 1. Current Active Queries
-- ========================================
\echo '📊 Active Queries (Running > 1 second):'
\echo '======================================='

SELECT 
    pid,
    now() - pg_stat_activity.query_start AS duration,
    usename,
    application_name,
    state,
    substring(query, 1, 100) AS query_preview
FROM pg_stat_activity
WHERE (now() - pg_stat_activity.query_start) > interval '1 second'
AND state != 'idle'
ORDER BY duration DESC;

-- 2. Top 10 Slowest Queries (requires pg_stat_statements)
-- ========================================
\echo ''
\echo '🐌 Top 10 Slowest Queries:'
\echo '=========================='

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'pg_stat_statements') THEN
        EXECUTE '
        SELECT 
            substring(query, 1, 60) AS query_preview,
            round(total_time::numeric, 2) AS total_time_ms,
            calls,
            round(mean_time::numeric, 2) AS mean_time_ms,
            round((100 * total_time / sum(total_time) OVER ())::numeric, 2) AS percentage_cpu
        FROM pg_stat_statements
        WHERE query NOT LIKE ''%pg_stat%''
        ORDER BY total_time DESC
        LIMIT 10';
    ELSE
        RAISE NOTICE 'pg_stat_statements extension not available. Install with: CREATE EXTENSION pg_stat_statements;';
    END IF;
END $$;

-- 3. Table Size and Bloat Analysis
-- ========================================
\echo ''
\echo '💾 Table Sizes and Bloat:'
\echo '========================'

WITH table_stats AS (
    SELECT 
        schemaname,
        relname AS tablename,
        pg_size_pretty(pg_total_relation_size(schemaname||'.'||relname)) AS total_size,
        pg_size_pretty(pg_relation_size(schemaname||'.'||relname)) AS table_size,
        pg_size_pretty(pg_total_relation_size(schemaname||'.'||relname) - pg_relation_size(schemaname||'.'||relname)) AS indexes_size,
        n_live_tup AS live_rows,
        n_dead_tup AS dead_rows,
        round(100.0 * n_dead_tup / NULLIF(n_live_tup + n_dead_tup, 0), 2) AS dead_percentage
    FROM pg_stat_user_tables
    WHERE schemaname = 'public'
)
SELECT * FROM table_stats
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC
LIMIT 20;

-- 4. Index Usage Statistics
-- ========================================
\echo ''
\echo '📈 Index Usage Statistics:'
\echo '========================='

WITH index_usage AS (
    SELECT 
        schemaname,
        relname AS tablename,
        indexrelname,
        idx_scan,
        idx_tup_read,
        idx_tup_fetch,
        pg_size_pretty(pg_relation_size(indexrelid)) AS index_size,
        CASE 
            WHEN idx_scan = 0 THEN 'UNUSED'
            WHEN idx_scan < 100 THEN 'RARELY USED'
            WHEN idx_scan < 1000 THEN 'OCCASIONALLY USED'
            ELSE 'FREQUENTLY USED'
        END AS usage_category
    FROM pg_stat_user_indexes
    WHERE schemaname = 'public'
)
SELECT 
    schemaname,
    tablename,
    indexrelname,
    idx_scan,
    idx_tup_read,
    idx_tup_fetch,
    index_size,
    usage_category
FROM index_usage
ORDER BY 
    CASE usage_category
        WHEN 'UNUSED' THEN 1
        WHEN 'RARELY USED' THEN 2
        WHEN 'OCCASIONALLY USED' THEN 3
        ELSE 4
    END,
    idx_scan DESC
LIMIT 30;

-- 5. Cache Hit Ratio
-- ========================================
\echo ''
\echo '🎯 Cache Hit Ratios:'
\echo '===================='

SELECT 
    'Database' AS cache_level,
    sum(heap_blks_read) AS disk_reads,
    sum(heap_blks_hit) AS cache_hits,
    CASE 
        WHEN sum(heap_blks_read) + sum(heap_blks_hit) = 0 THEN 0
        ELSE round(100.0 * sum(heap_blks_hit) / (sum(heap_blks_read) + sum(heap_blks_hit)), 2)
    END AS cache_hit_ratio
FROM pg_statio_user_tables
UNION ALL
SELECT 
    'Indexes' AS cache_level,
    sum(idx_blks_read) AS disk_reads,
    sum(idx_blks_hit) AS cache_hits,
    CASE 
        WHEN sum(idx_blks_read) + sum(idx_blks_hit) = 0 THEN 0
        ELSE round(100.0 * sum(idx_blks_hit) / (sum(idx_blks_read) + sum(idx_blks_hit)), 2)
    END AS cache_hit_ratio
FROM pg_statio_user_indexes;

-- 6. Lock Analysis
-- ========================================
\echo ''
\echo '🔒 Current Locks:'
\echo '================'

SELECT 
    locktype,
    relation::regclass AS table_name,
    mode,
    granted,
    pid,
    pg_blocking_pids(pid) AS blocked_by
FROM pg_locks
WHERE relation IS NOT NULL
AND relation::regclass::text LIKE '%public%'
ORDER BY granted, relation;

-- 7. Connection Statistics
-- ========================================
\echo ''
\echo '🔌 Connection Statistics:'
\echo '========================'

SELECT 
    datname,
    numbackends AS active_connections,
    (SELECT setting::int FROM pg_settings WHERE name = 'max_connections') AS max_connections,
    round(100.0 * numbackends / (SELECT setting::int FROM pg_settings WHERE name = 'max_connections'), 2) AS connection_percentage
FROM pg_stat_database
WHERE datname = current_database();

-- 8. Table Access Patterns
-- ========================================
\echo ''
\echo '📊 Table Access Patterns (Top 20):'
\echo '=================================='

SELECT 
    schemaname,
    relname AS tablename,
    seq_scan,
    seq_tup_read,
    idx_scan,
    idx_tup_fetch,
    n_tup_ins AS inserts,
    n_tup_upd AS updates,
    n_tup_del AS deletes,
    CASE 
        WHEN seq_scan + idx_scan = 0 THEN 0
        ELSE round(100.0 * idx_scan / (seq_scan + idx_scan), 2)
    END AS index_usage_percent
FROM pg_stat_user_tables
WHERE schemaname = 'public'
ORDER BY seq_scan + idx_scan DESC
LIMIT 20;

-- 9. Autovacuum Status
-- ========================================
\echo ''
\echo '🧹 Autovacuum Status:'
\echo '===================='

SELECT 
    schemaname,
    relname AS tablename,
    last_vacuum,
    last_autovacuum,
    vacuum_count,
    autovacuum_count,
    last_analyze,
    last_autoanalyze,
    n_dead_tup AS dead_tuples,
    CASE 
        WHEN last_autovacuum IS NULL THEN 'NEVER'
        WHEN now() - last_autovacuum > interval '7 days' THEN 'NEEDS VACUUM'
        ELSE 'OK'
    END AS vacuum_status
FROM pg_stat_user_tables
WHERE schemaname = 'public'
AND n_dead_tup > 1000
ORDER BY n_dead_tup DESC
LIMIT 20;

-- 10. Database Growth Trend
-- ========================================
\echo ''
\echo '📈 Database Size:'
\echo '================'

SELECT 
    current_database() AS database,
    pg_size_pretty(pg_database_size(current_database())) AS total_size,
    (SELECT count(*) FROM pg_stat_user_tables WHERE schemaname = 'public') AS table_count,
    (SELECT count(*) FROM pg_stat_user_indexes WHERE schemaname = 'public') AS index_count,
    (SELECT sum(n_live_tup) FROM pg_stat_user_tables WHERE schemaname = 'public') AS total_rows;

-- 11. Performance Recommendations
-- ========================================
\echo ''
\echo '💡 Performance Recommendations:'
\echo '==============================='

-- Check for missing indexes on foreign keys
WITH fk_indexes AS (
    SELECT 
        c.conrelid::regclass AS table_name,
        array_to_string(array_agg(a.attname), ', ') AS fk_columns,
        c.confrelid::regclass AS referenced_table
    FROM pg_constraint c
    JOIN pg_attribute a ON a.attrelid = c.conrelid AND a.attnum = ANY(c.conkey)
    WHERE c.contype = 'f'
    GROUP BY c.conrelid, c.confrelid, c.conkey
)
SELECT 
    'Missing index on foreign key' AS recommendation,
    table_name,
    fk_columns,
    'CREATE INDEX idx_' || table_name || '_' || replace(fk_columns, ', ', '_') || ' ON ' || table_name || '(' || fk_columns || ');' AS suggested_action
FROM fk_indexes f
WHERE NOT EXISTS (
    SELECT 1 
    FROM pg_index i
    JOIN pg_attribute a ON a.attrelid = i.indrelid AND a.attnum = ANY(i.indkey)
    WHERE i.indrelid = f.table_name::regclass
    GROUP BY i.indexrelid
    HAVING array_to_string(array_agg(a.attname), ', ') = f.fk_columns
)
LIMIT 10;