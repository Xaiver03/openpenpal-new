-- Query 1: List all indexes on key tables with their definitions
SELECT 
    schemaname,
    tablename,
    indexname,
    indexdef,
    tablespace
FROM pg_indexes 
WHERE schemaname = 'public' 
    AND tablename IN (
        'users', 'letters', 'user_credits', 'credit_transactions',
        'couriers', 'courier_tasks', 'user_profiles', 'comments',
        'museum_items', 'museum_collections', 'envelopes',
        'letter_codes', 'status_logs', 'letter_photos',
        'letter_likes', 'letter_shares', 'notifications',
        'moderation_records', 'ai_matches', 'ai_replies'
    )
ORDER BY tablename, indexname;

-- Query 2: Detailed index statistics
SELECT 
    schemaname,
    tablename,
    indexrelname AS index_name,
    idx_scan AS index_scans,
    idx_tup_read AS tuples_read,
    idx_tup_fetch AS tuples_fetched,
    pg_size_pretty(pg_relation_size(indexrelid)) AS index_size
FROM pg_stat_user_indexes
WHERE schemaname = 'public'
ORDER BY tablename, indexrelname;

-- Query 3: Find duplicate or redundant indexes (indexes with same columns)
WITH index_info AS (
    SELECT 
        n.nspname AS schema_name,
        t.relname AS table_name,
        i.relname AS index_name,
        array_agg(a.attname ORDER BY a.attnum) AS column_names,
        pg_get_indexdef(i.oid) AS index_definition
    FROM pg_index ix
    JOIN pg_class t ON t.oid = ix.indrelid
    JOIN pg_class i ON i.oid = ix.indexrelid
    JOIN pg_namespace n ON n.oid = t.relnamespace
    JOIN pg_attribute a ON a.attrelid = t.oid AND a.attnum = ANY(ix.indkey)
    WHERE n.nspname = 'public'
    GROUP BY n.nspname, t.relname, i.relname, i.oid
)
SELECT 
    table_name,
    array_agg(index_name) AS duplicate_indexes,
    column_names
FROM index_info
GROUP BY table_name, column_names
HAVING count(*) > 1
ORDER BY table_name;

-- Query 4: Find composite indexes
SELECT 
    schemaname,
    tablename,
    indexname,
    indexdef
FROM pg_indexes
WHERE schemaname = 'public'
    AND indexdef LIKE '%,%' -- Contains comma, indicating multiple columns
ORDER BY tablename, indexname;

-- Query 5: Find partial indexes
SELECT 
    schemaname,
    tablename,
    indexname,
    indexdef
FROM pg_indexes
WHERE schemaname = 'public'
    AND indexdef LIKE '%WHERE%'
ORDER BY tablename, indexname;

-- Query 6: Find unused indexes (never scanned)
SELECT 
    schemaname,
    tablename,
    indexrelname AS index_name,
    idx_scan AS times_used,
    pg_size_pretty(pg_relation_size(indexrelid)) AS index_size,
    indexrelid::regclass AS index
FROM pg_stat_user_indexes
WHERE schemaname = 'public'
    AND idx_scan = 0
    AND indexrelname NOT LIKE '%_pkey'  -- Exclude primary keys
ORDER BY pg_relation_size(indexrelid) DESC;

-- Query 7: Missing indexes based on foreign key relationships
SELECT 
    tc.table_name,
    kcu.column_name,
    ccu.table_name AS foreign_table_name,
    ccu.column_name AS foreign_column_name,
    'CREATE INDEX idx_' || tc.table_name || '_' || kcu.column_name || 
    ' ON ' || tc.table_schema || '.' || tc.table_name || 
    ' (' || kcu.column_name || ');' AS suggested_index
FROM information_schema.table_constraints AS tc
JOIN information_schema.key_column_usage AS kcu
    ON tc.constraint_name = kcu.constraint_name
    AND tc.table_schema = kcu.table_schema
JOIN information_schema.constraint_column_usage AS ccu
    ON ccu.constraint_name = tc.constraint_name
    AND ccu.table_schema = tc.table_schema
WHERE tc.constraint_type = 'FOREIGN KEY'
    AND tc.table_schema = 'public'
    AND NOT EXISTS (
        SELECT 1 
        FROM pg_indexes pi
        WHERE pi.schemaname = tc.table_schema
            AND pi.tablename = tc.table_name
            AND pi.indexdef LIKE '%(' || kcu.column_name || ')%'
    )
ORDER BY tc.table_name, kcu.column_name;

-- Query 8: Table sizes and row counts for context
SELECT 
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS total_size,
    pg_size_pretty(pg_relation_size(schemaname||'.'||tablename)) AS table_size,
    n_live_tup AS row_count,
    n_dead_tup AS dead_rows
FROM pg_stat_user_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC
LIMIT 20;

-- Query 9: Index bloat analysis
WITH btree_index_atts AS (
    SELECT 
        nspname,
        indexclass.relname AS index_name,
        indexclass.reltuples,
        indexclass.relpages,
        tableclass.relname AS tablename,
        (
            SELECT COUNT(*)
            FROM pg_index
            WHERE pg_index.indexrelid = indexclass.oid
        ) AS number_of_columns,
        idx_scan,
        idx_tup_read,
        idx_tup_fetch
    FROM pg_index
    JOIN pg_class AS indexclass ON pg_index.indexrelid = indexclass.oid
    JOIN pg_class AS tableclass ON pg_index.indrelid = tableclass.oid
    JOIN pg_namespace ON pg_namespace.oid = indexclass.relnamespace
    JOIN pg_stat_user_indexes ON indexclass.oid = pg_stat_user_indexes.indexrelid
    WHERE indexclass.relkind = 'i'
        AND nspname = 'public'
)
SELECT 
    nspname AS schema_name,
    tablename,
    index_name,
    pg_size_pretty(pg_relation_size(quote_ident(nspname)||'.'||quote_ident(index_name))) AS index_size,
    CASE 
        WHEN relpages > 0 
        THEN round(100.0 * (relpages - (reltuples * 16 / 8192)) / relpages, 2)
        ELSE 0
    END AS bloat_percentage
FROM btree_index_atts
WHERE relpages > 10
ORDER BY bloat_percentage DESC;

-- Query 10: Frequently accessed tables without proper indexing
SELECT 
    schemaname,
    tablename,
    seq_scan,
    seq_tup_read,
    idx_scan,
    n_live_tup AS total_rows,
    CASE 
        WHEN seq_scan > 0 
        THEN round((seq_tup_read::numeric / seq_scan), 2)
        ELSE 0
    END AS avg_rows_per_seq_scan
FROM pg_stat_user_tables
WHERE schemaname = 'public'
    AND seq_scan > 1000
    AND n_live_tup > 1000
ORDER BY seq_scan DESC
LIMIT 20;