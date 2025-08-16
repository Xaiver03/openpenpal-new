-- OpenPenPal Database Index Optimization Script (ULTRATHINK Version)
-- Based on comprehensive analysis of existing 113 indexes
-- This script only adds missing indexes and removes duplicates

-- ========================================
-- 1. REMOVE DUPLICATE INDEXES (Save Space & Improve Write Performance)
-- ========================================

-- Remove duplicate indexes on courier_tasks
DROP INDEX IF EXISTS idx_courier_tasks_delivery_op_code_dup1;
DROP INDEX IF EXISTS idx_courier_tasks_pickup_op_code_dup1;

-- Remove duplicate index on couriers  
DROP INDEX IF EXISTS idx_couriers_managed_op_code_prefix_dup1;

-- ========================================
-- 2. ADD MISSING FOREIGN KEY INDEXES
-- ========================================

-- Envelope design relationship (for join queries)
CREATE INDEX IF NOT EXISTS idx_envelopes_design_id 
    ON envelopes(design_id) 
    WHERE design_id IS NOT NULL;

-- Museum item relationships (for admin queries and joins)
CREATE INDEX IF NOT EXISTS idx_museum_items_approved_by 
    ON museum_items(approved_by) 
    WHERE approved_by IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_museum_items_source_id 
    ON museum_items(source_id) 
    WHERE source_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_museum_items_submitted_by 
    ON museum_items(submitted_by) 
    WHERE submitted_by IS NOT NULL;

-- ========================================
-- 3. ADD MISSING HIGH-VALUE COMPOSITE INDEXES
-- ========================================

-- Credit transaction history queries (user dashboard)
-- Analysis shows this is frequently queried but lacks composite index
CREATE INDEX IF NOT EXISTS idx_credit_transactions_user_created_composite 
    ON credit_transactions(user_id, created_at DESC);

-- Active user listings (admin dashboard)
-- Needed for efficient user management queries
CREATE INDEX IF NOT EXISTS idx_users_active_created 
    ON users(is_active, created_at DESC) 
    WHERE deleted_at IS NULL AND is_active = true;

-- Credit expiration monitoring (batch processing)
-- Critical for performance of expiration job
CREATE INDEX IF NOT EXISTS idx_credit_expiration_batches_status_expire 
    ON credit_expiration_batches(status, expire_at) 
    WHERE status IN ('pending', 'processing');

-- Museum featured entries (homepage queries)
-- High-frequency query that needs optimization
CREATE INDEX IF NOT EXISTS idx_museum_entries_featured_active 
    ON museum_entries(is_featured, status, created_at DESC) 
    WHERE is_featured = true AND status = 'active' AND deleted_at IS NULL;

-- ========================================
-- 4. PERFORMANCE MONITORING INDEXES
-- ========================================

-- These support the new monitoring views
CREATE INDEX IF NOT EXISTS idx_analytics_metrics_composite 
    ON analytics_metrics(metric_type, period_start DESC, category);

CREATE INDEX IF NOT EXISTS idx_performance_metrics_composite 
    ON performance_metrics(metric_name, recorded_at DESC)
    WHERE value > 0;

-- ========================================
-- 5. VERIFY CRITICAL INDEXES EXIST (Safety Check)
-- ========================================

-- These should already exist, but ensure they're present
CREATE INDEX IF NOT EXISTS idx_letter_codes_code 
    ON letter_codes(code) 
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_scan_records_signal_code 
    ON scan_records(signal_code, scan_time DESC);

-- ========================================
-- 6. CREATE PERFORMANCE MONITORING VIEW
-- ========================================

CREATE OR REPLACE VIEW v_index_usage_stats AS
SELECT 
    schemaname,
    tablename,
    indexname,
    idx_scan as index_scans,
    idx_tup_read as tuples_read,
    idx_tup_fetch as tuples_fetched,
    pg_size_pretty(pg_relation_size(indexrelid)) as index_size,
    CASE 
        WHEN idx_scan = 0 THEN 'UNUSED'
        WHEN idx_scan < 100 THEN 'LOW_USE'
        WHEN idx_scan < 1000 THEN 'MODERATE_USE'
        ELSE 'HIGH_USE'
    END as usage_category
FROM pg_stat_user_indexes
WHERE schemaname = 'public'
ORDER BY idx_scan DESC;

-- ========================================
-- 7. UPDATE STATISTICS
-- ========================================

-- Update statistics for query planner on modified tables
ANALYZE envelopes;
ANALYZE museum_items;
ANALYZE credit_transactions;
ANALYZE credit_expiration_batches;
ANALYZE museum_entries;
ANALYZE analytics_metrics;
ANALYZE performance_metrics;

-- ========================================
-- 8. VALIDATION QUERIES
-- ========================================

-- Check for remaining duplicates
WITH index_cols AS (
    SELECT 
        schemaname,
        tablename,
        indexname,
        array_agg(attname ORDER BY attnum) as columns
    FROM pg_indexes
    JOIN pg_index ON indexname = pg_indexes.indexname::regclass::oid::regclass::text
    JOIN pg_attribute ON attrelid = indrelid AND attnum = ANY(indkey)
    WHERE schemaname = 'public'
    GROUP BY schemaname, tablename, indexname
)
SELECT 
    'Potential duplicate:' as status,
    tablename,
    COUNT(*) as duplicate_count,
    string_agg(indexname, ', ') as index_names,
    columns[1] as indexed_columns
FROM index_cols
GROUP BY tablename, columns
HAVING COUNT(*) > 1
ORDER BY tablename;

-- Summary of optimization results
SELECT 
    'Index Optimization Complete' as status,
    COUNT(*) FILTER (WHERE usage_category = 'UNUSED') as unused_indexes,
    COUNT(*) FILTER (WHERE usage_category = 'HIGH_USE') as high_use_indexes,
    COUNT(*) as total_indexes,
    pg_size_pretty(SUM(pg_relation_size(indexrelid))) as total_index_size
FROM v_index_usage_stats;