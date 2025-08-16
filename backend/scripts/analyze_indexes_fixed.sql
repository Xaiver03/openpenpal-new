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

-- Query 2: Detailed index statistics (fixing column name issue)
SELECT 
    s.schemaname,
    s.tablename,
    s.indexrelname AS index_name,
    s.idx_scan AS index_scans,
    s.idx_tup_read AS tuples_read,
    s.idx_tup_fetch AS tuples_fetched,
    pg_size_pretty(pg_relation_size(s.indexrelid)) AS index_size
FROM pg_stat_user_indexes s
WHERE s.schemaname = 'public'
    AND s.tablename IN (
        'users', 'letters', 'user_credits', 'credit_transactions',
        'couriers', 'courier_tasks', 'user_profiles', 'comments',
        'museum_items', 'museum_collections', 'envelopes',
        'letter_codes', 'status_logs', 'letter_photos',
        'letter_likes', 'letter_shares', 'notifications',
        'moderation_records', 'ai_matches', 'ai_replies'
    )
ORDER BY s.tablename, s.indexrelname;

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
        AND t.relname IN (
            'users', 'letters', 'user_credits', 'credit_transactions',
            'couriers', 'courier_tasks', 'user_profiles', 'comments',
            'museum_items', 'museum_collections', 'envelopes',
            'letter_codes', 'status_logs', 'letter_photos',
            'letter_likes', 'letter_shares', 'notifications',
            'moderation_records', 'ai_matches', 'ai_replies'
        )
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

-- Query 4: Find composite indexes on key tables
SELECT 
    schemaname,
    tablename,
    indexname,
    indexdef
FROM pg_indexes
WHERE schemaname = 'public'
    AND indexdef LIKE '%,%' -- Contains comma, indicating multiple columns
    AND tablename IN (
        'users', 'letters', 'user_credits', 'credit_transactions',
        'couriers', 'courier_tasks', 'user_profiles', 'comments',
        'museum_items', 'museum_collections', 'envelopes',
        'letter_codes', 'status_logs', 'letter_photos',
        'letter_likes', 'letter_shares', 'notifications',
        'moderation_records', 'ai_matches', 'ai_replies'
    )
ORDER BY tablename, indexname;

-- Query 5: Find partial indexes on key tables
SELECT 
    schemaname,
    tablename,
    indexname,
    indexdef
FROM pg_indexes
WHERE schemaname = 'public'
    AND indexdef LIKE '%WHERE%'
    AND tablename IN (
        'users', 'letters', 'user_credits', 'credit_transactions',
        'couriers', 'courier_tasks', 'user_profiles', 'comments',
        'museum_items', 'museum_collections', 'envelopes',
        'letter_codes', 'status_logs', 'letter_photos',
        'letter_likes', 'letter_shares', 'notifications',
        'moderation_records', 'ai_matches', 'ai_replies'
    )
ORDER BY tablename, indexname;

-- Query 6: Find unused indexes on key tables (never scanned)
SELECT 
    s.schemaname,
    s.tablename,
    s.indexrelname AS index_name,
    s.idx_scan AS times_used,
    pg_size_pretty(pg_relation_size(s.indexrelid)) AS index_size,
    s.indexrelid::regclass AS index
FROM pg_stat_user_indexes s
WHERE s.schemaname = 'public'
    AND s.idx_scan = 0
    AND s.indexrelname NOT LIKE '%_pkey'  -- Exclude primary keys
    AND s.tablename IN (
        'users', 'letters', 'user_credits', 'credit_transactions',
        'couriers', 'courier_tasks', 'user_profiles', 'comments',
        'museum_items', 'museum_collections', 'envelopes',
        'letter_codes', 'status_logs', 'letter_photos',
        'letter_likes', 'letter_shares', 'notifications',
        'moderation_records', 'ai_matches', 'ai_replies'
    )
ORDER BY pg_relation_size(s.indexrelid) DESC;

-- Query 7: Missing indexes based on foreign key relationships for key tables
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
    AND tc.table_name IN (
        'users', 'letters', 'user_credits', 'credit_transactions',
        'couriers', 'courier_tasks', 'user_profiles', 'comments',
        'museum_items', 'museum_collections', 'envelopes',
        'letter_codes', 'status_logs', 'letter_photos',
        'letter_likes', 'letter_shares', 'notifications',
        'moderation_records', 'ai_matches', 'ai_replies'
    )
    AND NOT EXISTS (
        SELECT 1 
        FROM pg_indexes pi
        WHERE pi.schemaname = tc.table_schema
            AND pi.tablename = tc.table_name
            AND pi.indexdef LIKE '%(' || kcu.column_name || ')%'
    )
ORDER BY tc.table_name, kcu.column_name;

-- Query 8: Table sizes and row counts for key tables
SELECT 
    s.schemaname,
    s.tablename,
    pg_size_pretty(pg_total_relation_size(s.schemaname||'.'||s.tablename)) AS total_size,
    pg_size_pretty(pg_relation_size(s.schemaname||'.'||s.tablename)) AS table_size,
    s.n_live_tup AS row_count,
    s.n_dead_tup AS dead_rows
FROM pg_stat_user_tables s
WHERE s.schemaname = 'public'
    AND s.tablename IN (
        'users', 'letters', 'user_credits', 'credit_transactions',
        'couriers', 'courier_tasks', 'user_profiles', 'comments',
        'museum_items', 'museum_collections', 'envelopes',
        'letter_codes', 'status_logs', 'letter_photos',
        'letter_likes', 'letter_shares', 'notifications',
        'moderation_records', 'ai_matches', 'ai_replies'
    )
ORDER BY pg_total_relation_size(s.schemaname||'.'||s.tablename) DESC;

-- Query 9: Frequently accessed tables without proper indexing
SELECT 
    s.schemaname,
    s.tablename,
    s.seq_scan,
    s.seq_tup_read,
    s.idx_scan,
    s.n_live_tup AS total_rows,
    CASE 
        WHEN s.seq_scan > 0 
        THEN (s.seq_tup_read::numeric / s.seq_scan)::bigint
        ELSE 0
    END AS avg_rows_per_seq_scan
FROM pg_stat_user_tables s
WHERE s.schemaname = 'public'
    AND s.seq_scan > 100
    AND s.n_live_tup > 100
    AND s.tablename IN (
        'users', 'letters', 'user_credits', 'credit_transactions',
        'couriers', 'courier_tasks', 'user_profiles', 'comments',
        'museum_items', 'museum_collections', 'envelopes',
        'letter_codes', 'status_logs', 'letter_photos',
        'letter_likes', 'letter_shares', 'notifications',
        'moderation_records', 'ai_matches', 'ai_replies'
    )
ORDER BY s.seq_scan DESC;

-- Query 10: Analyze common query patterns based on models
SELECT 
    'Common Query Patterns Analysis:' AS analysis,
    'Based on the models and typical application usage' AS description;

SELECT 
    'users table' AS table_name,
    'Common queries: by username, email, school_code, role' AS query_pattern,
    'Existing: username, email, school_code, op_code, role+school_code+is_active' AS existing_indexes,
    'Recommendation: Add index on (is_active, created_at) for user listings' AS recommendation;

SELECT 
    'letters table' AS table_name,
    'Common queries: by user_id+status, recipient_op_code, sender_op_code' AS query_pattern,
    'Existing: Multiple composite indexes covering various patterns' AS existing_indexes,
    'Well-indexed table with good coverage' AS recommendation;

SELECT 
    'user_credits table' AS table_name,
    'Common queries: by user_id' AS query_pattern,
    'Existing: unique index on user_id' AS existing_indexes,
    'Well-indexed for primary use case' AS recommendation;

SELECT 
    'credit_transactions table' AS table_name,
    'Common queries: by user_id, expires_at, is_expired' AS query_pattern,
    'Existing: user_id, expires_at, is_expired' AS existing_indexes,
    'Consider composite index on (user_id, created_at) for transaction history' AS recommendation;