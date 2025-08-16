-- OpenPenPal Database Index Optimization Script (Minimal Safe Version)
-- Only adds verified missing indexes

-- ========================================
-- 1. ADD MISSING FOREIGN KEY INDEXES
-- ========================================

-- Envelope design relationship
CREATE INDEX IF NOT EXISTS idx_envelopes_design_id 
    ON envelopes(design_id) 
    WHERE design_id IS NOT NULL;

-- Museum item relationships
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
-- 2. ADD KEY COMPOSITE INDEXES
-- ========================================

-- Credit transaction history queries
CREATE INDEX IF NOT EXISTS idx_credit_transactions_user_created 
    ON credit_transactions(user_id, created_at DESC);

-- Credit expiration batch processing
CREATE INDEX IF NOT EXISTS idx_credit_expiration_batches_status 
    ON credit_expiration_batches(status) 
    WHERE status IN ('pending', 'processing');

-- ========================================
-- 3. UPDATE STATISTICS
-- ========================================

ANALYZE envelopes;
ANALYZE museum_items;
ANALYZE credit_transactions;
ANALYZE credit_expiration_batches;

-- ========================================
-- 4. SUMMARY
-- ========================================

SELECT 
    'Index Optimization Complete' as status,
    COUNT(*) as total_indexes
FROM pg_indexes 
WHERE schemaname = 'public';