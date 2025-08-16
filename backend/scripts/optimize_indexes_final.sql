-- OpenPenPal Database Index Optimization Script (Final Version)
-- Verified against actual table structures

-- ========================================
-- 1. ADD MISSING FOREIGN KEY INDEXES
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
-- 2. ADD HIGH-VALUE COMPOSITE INDEXES
-- ========================================

-- Credit transaction history queries (user dashboard)
CREATE INDEX IF NOT EXISTS idx_credit_transactions_user_created_composite 
    ON credit_transactions(user_id, created_at DESC);

-- Active user listings (admin dashboard)
CREATE INDEX IF NOT EXISTS idx_users_active_created 
    ON users(is_active, created_at DESC) 
    WHERE deleted_at IS NULL AND is_active = true;

-- Credit expiration batch processing
CREATE INDEX IF NOT EXISTS idx_credit_expiration_batches_status_date 
    ON credit_expiration_batches(status, batch_date DESC) 
    WHERE status IN ('pending', 'processing');

-- Museum featured entries (homepage queries)
CREATE INDEX IF NOT EXISTS idx_museum_entries_featured_active 
    ON museum_entries(is_featured, status, created_at DESC) 
    WHERE is_featured = true AND status = 'active' AND deleted_at IS NULL;

-- Analytics metrics composite
CREATE INDEX IF NOT EXISTS idx_analytics_metrics_composite 
    ON analytics_metrics(metric_type, period_start DESC);

-- Performance metrics composite
CREATE INDEX IF NOT EXISTS idx_performance_metrics_composite 
    ON performance_metrics(metric_name, recorded_at DESC)
    WHERE value > 0;

-- ========================================
-- 3. VERIFY CRITICAL INDEXES EXIST
-- ========================================

-- Letter code lookups (barcode scanning)
CREATE INDEX IF NOT EXISTS idx_letter_codes_code 
    ON letter_codes(code) 
    WHERE deleted_at IS NULL;

-- Scan records for tracking
CREATE INDEX IF NOT EXISTS idx_scan_records_signal_code 
    ON scan_records(signal_code, scan_time DESC);

-- ========================================
-- 4. UPDATE STATISTICS
-- ========================================

ANALYZE envelopes;
ANALYZE museum_items;
ANALYZE credit_transactions;
ANALYZE credit_expiration_batches;
ANALYZE museum_entries;
ANALYZE analytics_metrics;
ANALYZE performance_metrics;
ANALYZE users;
ANALYZE letter_codes;
ANALYZE scan_records;

-- ========================================
-- 5. CREATE INDEX USAGE VIEW
-- ========================================

CREATE OR REPLACE VIEW v_index_usage_report AS
SELECT 
    s.schemaname,
    s.relname as table_name,
    s.indexrelname as index_name,
    s.idx_scan as index_scans,
    s.idx_tup_read as tuples_read,
    s.idx_tup_fetch as tuples_fetched,
    pg_size_pretty(pg_relation_size(s.indexrelid)) as index_size,
    CASE 
        WHEN s.idx_scan = 0 THEN 'UNUSED'
        WHEN s.idx_scan < 100 THEN 'LOW_USE'
        WHEN s.idx_scan < 1000 THEN 'MODERATE_USE'
        ELSE 'HIGH_USE'
    END as usage_category,
    i.indisunique as is_unique,
    i.indisprimary as is_primary
FROM pg_stat_user_indexes s
JOIN pg_index i ON i.indexrelid = s.indexrelid
WHERE s.schemaname = 'public'
ORDER BY s.idx_scan DESC;

-- ========================================
-- 6. SUMMARY REPORT
-- ========================================

SELECT 
    'Index Optimization Summary' as report,
    COUNT(*) as total_indexes,
    COUNT(*) FILTER (WHERE idx_scan = 0) as unused_indexes,
    COUNT(*) FILTER (WHERE idx_scan > 1000) as high_use_indexes,
    pg_size_pretty(SUM(pg_relation_size(indexrelid))) as total_index_size
FROM pg_stat_user_indexes
WHERE schemaname = 'public';