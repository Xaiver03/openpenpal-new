-- Additional queries for complete index analysis

-- Query 1: Get actual statistics for index usage
SELECT 
    t.schemaname,
    t.tablename,
    indexname,
    idx_scan AS scans_count,
    idx_tup_read AS tuples_read,
    idx_tup_fetch AS tuples_fetched,
    pg_size_pretty(pg_relation_size(i.indexrelid)) AS index_size
FROM pg_stat_user_indexes i
JOIN pg_stat_user_tables t ON i.schemaname = t.schemaname AND i.tablename = t.tablename
WHERE i.schemaname = 'public'
    AND i.tablename IN (
        'users', 'letters', 'user_credits', 'credit_transactions',
        'couriers', 'courier_tasks', 'user_profiles', 'comments',
        'museum_items', 'museum_collections', 'envelopes',
        'letter_codes', 'status_logs', 'letter_photos',
        'letter_likes', 'letter_shares', 'notifications',
        'moderation_records', 'ai_matches', 'ai_replies'
    )
ORDER BY i.tablename, idx_scan DESC;

-- Query 2: Identify potential missing indexes based on sequential scans
SELECT 
    schemaname,
    tablename,
    seq_scan AS sequential_scans,
    seq_tup_read AS seq_tuples_read,
    idx_scan AS index_scans,
    n_live_tup AS total_rows,
    CASE 
        WHEN seq_scan > 0 
        THEN ROUND((seq_tup_read::numeric / seq_scan)::numeric, 2)
        ELSE 0
    END AS avg_rows_per_seq_scan,
    CASE 
        WHEN seq_scan > 0 AND idx_scan > 0
        THEN ROUND((seq_scan::numeric / (seq_scan + idx_scan) * 100)::numeric, 2)
        ELSE 0
    END AS seq_scan_percentage
FROM pg_stat_user_tables
WHERE schemaname = 'public'
    AND tablename IN (
        'users', 'letters', 'user_credits', 'credit_transactions',
        'couriers', 'courier_tasks', 'user_profiles', 'comments',
        'museum_items', 'museum_collections', 'envelopes',
        'letter_codes', 'status_logs', 'letter_photos',
        'letter_likes', 'letter_shares', 'notifications',
        'moderation_records', 'ai_matches', 'ai_replies'
    )
    AND seq_scan > 0
ORDER BY seq_scan DESC;

-- Query 3: Index bloat analysis (simplified)
WITH index_stats AS (
    SELECT
        schemaname,
        tablename,
        indexrelname,
        pg_relation_size(indexrelid) AS actual_size,
        pg_stat_get_live_tuples(indexrelid) AS num_rows
    FROM pg_stat_user_indexes
    WHERE schemaname = 'public'
        AND tablename IN (
            'users', 'letters', 'user_credits', 'credit_transactions',
            'couriers', 'courier_tasks', 'user_profiles', 'comments',
            'museum_items', 'museum_collections', 'envelopes',
            'letter_codes', 'status_logs', 'letter_photos',
            'letter_likes', 'letter_shares', 'notifications',
            'moderation_records', 'ai_matches', 'ai_replies'
        )
)
SELECT 
    schemaname,
    tablename,
    indexrelname,
    pg_size_pretty(actual_size) AS index_size,
    num_rows AS estimated_rows
FROM index_stats
WHERE actual_size > 8192  -- Only show indexes larger than 8KB
ORDER BY actual_size DESC;

-- Query 4: Summary recommendations
SELECT 'INDEX ANALYSIS SUMMARY' AS report_section, '' AS details
UNION ALL
SELECT '====================' AS report_section, '' AS details
UNION ALL
SELECT '', ''
UNION ALL
SELECT 'DUPLICATE INDEXES FOUND:', '' AS details
UNION ALL
SELECT '- courier_tasks: delivery_op_code has 2 duplicate indexes', '' AS details
UNION ALL
SELECT '- courier_tasks: pickup_op_code has 2 duplicate indexes', '' AS details
UNION ALL
SELECT '- couriers: managed_op_code_prefix has 2 duplicate indexes', '' AS details
UNION ALL
SELECT '', ''
UNION ALL
SELECT 'MISSING INDEXES (Foreign Keys without indexes):', '' AS details
UNION ALL
SELECT '- envelopes.design_id -> envelope_designs.id', '' AS details
UNION ALL
SELECT '- museum_items.approved_by -> users.id', '' AS details
UNION ALL
SELECT '- museum_items.source_id -> letters.id', '' AS details
UNION ALL
SELECT '- museum_items.submitted_by -> users.id', '' AS details
UNION ALL
SELECT '', ''
UNION ALL
SELECT 'WELL-INDEXED TABLES:', '' AS details
UNION ALL
SELECT '- letters: Comprehensive indexing with composite and partial indexes', '' AS details
UNION ALL
SELECT '- users: Good coverage for authentication and queries', '' AS details
UNION ALL
SELECT '- courier_tasks: Excellent indexing for routing and status queries', '' AS details
UNION ALL
SELECT '- letter_codes: Well-indexed for tracking and scanning', '' AS details
UNION ALL
SELECT '', ''
UNION ALL
SELECT 'RECOMMENDATIONS:', '' AS details
UNION ALL
SELECT '1. Remove duplicate indexes to save space and improve write performance', '' AS details
UNION ALL
SELECT '2. Add missing foreign key indexes for better join performance', '' AS details
UNION ALL
SELECT '3. Consider adding composite index on credit_transactions(user_id, created_at)', '' AS details
UNION ALL
SELECT '4. Monitor sequential scan patterns on less-used tables', '' AS details;