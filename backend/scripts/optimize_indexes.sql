-- OpenPenPal Database Index Optimization Script
-- Based on analysis of high-frequency queries and business logic

-- ========================================
-- 1. User and Authentication Indexes
-- ========================================

-- User lookups by username/email (login scenarios)
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_users_phone ON users(phone) WHERE phone IS NOT NULL AND deleted_at IS NULL;

-- User profile lookups
CREATE INDEX IF NOT EXISTS idx_user_profiles_user_id ON user_profiles(user_id);
CREATE INDEX IF NOT EXISTS idx_user_profiles_school_id ON user_profiles(school_id) WHERE school_id IS NOT NULL;

-- ========================================
-- 2. Letter System Indexes
-- ========================================

-- Letter queries by sender/receiver
CREATE INDEX IF NOT EXISTS idx_letters_sender_status ON letters(sender_id, status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_letters_receiver_status ON letters(receiver_id, status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_letters_created_at ON letters(created_at DESC) WHERE deleted_at IS NULL;

-- Letter code lookups (frequent barcode scanning)
CREATE INDEX IF NOT EXISTS idx_letter_codes_code ON letter_codes(code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_letter_codes_letter_id ON letter_codes(letter_id);

-- Scan records for tracking
CREATE INDEX IF NOT EXISTS idx_scan_records_code_time ON scan_records(signal_code, scan_time DESC);
CREATE INDEX IF NOT EXISTS idx_scan_records_op_code ON scan_records(op_code) WHERE op_code IS NOT NULL;

-- ========================================
-- 3. Credit System Indexes
-- ========================================

-- User credit balance lookups
CREATE INDEX IF NOT EXISTS idx_user_credits_user_id ON user_credits(user_id);

-- Credit transaction history
CREATE INDEX IF NOT EXISTS idx_credit_transactions_user_created ON credit_transactions(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_credit_transactions_ref_type ON credit_transactions(reference_type, reference_id) WHERE reference_id IS NOT NULL;

-- Credit expiration queries
CREATE INDEX IF NOT EXISTS idx_credit_expiration_batches_user_status ON credit_expiration_batches(user_id, status) 
    WHERE status IN ('pending', 'processing');
CREATE INDEX IF NOT EXISTS idx_credit_expiration_batches_expire_at ON credit_expiration_batches(expire_at) 
    WHERE status = 'pending';

-- Credit transfer lookups
CREATE INDEX IF NOT EXISTS idx_credit_transfers_sender_created ON credit_transfers(sender_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_credit_transfers_receiver_created ON credit_transfers(receiver_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_credit_transfers_status ON credit_transfers(status) 
    WHERE status IN ('pending', 'processing');

-- ========================================
-- 4. Courier System Indexes
-- ========================================

-- Courier lookups by user and level
CREATE INDEX IF NOT EXISTS idx_couriers_user_level ON couriers(user_id, level);
CREATE INDEX IF NOT EXISTS idx_couriers_parent ON couriers(parent_id) WHERE parent_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_couriers_op_code ON couriers(op_code) WHERE op_code IS NOT NULL;

-- Task assignments and status
CREATE INDEX IF NOT EXISTS idx_tasks_assignee_status ON tasks(assignee_id, status) 
    WHERE status IN ('pending', 'accepted', 'in_progress');
CREATE INDEX IF NOT EXISTS idx_tasks_created_at ON tasks(created_at DESC);

-- Courier statistics
CREATE INDEX IF NOT EXISTS idx_courier_statistics_courier_period ON courier_statistics(courier_id, period_start DESC);

-- ========================================
-- 5. Museum System Indexes
-- ========================================

-- Museum entry queries
CREATE INDEX IF NOT EXISTS idx_museum_entries_user_created ON museum_entries(user_id, created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_museum_entries_status ON museum_entries(status) WHERE status = 'active';
CREATE INDEX IF NOT EXISTS idx_museum_entries_featured ON museum_entries(is_featured, created_at DESC) WHERE is_featured = true;

-- Museum interactions
CREATE INDEX IF NOT EXISTS idx_museum_interactions_entry_type ON museum_interactions(entry_id, interaction_type);
CREATE INDEX IF NOT EXISTS idx_museum_interactions_user_entry ON museum_interactions(user_id, entry_id);

-- ========================================
-- 6. Notification System Indexes
-- ========================================

-- Unread notifications
CREATE INDEX IF NOT EXISTS idx_notifications_user_read ON notifications(user_id, is_read, created_at DESC) 
    WHERE is_read = false;

-- Notification batches
CREATE INDEX IF NOT EXISTS idx_notification_batches_status_scheduled ON notification_batches(status, scheduled_at) 
    WHERE status IN ('pending', 'scheduled');

-- ========================================
-- 7. Social Features Indexes
-- ========================================

-- Comments by target
CREATE INDEX IF NOT EXISTS idx_comments_target ON comments(target_type, target_id, created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_comments_user_created ON comments(user_id, created_at DESC) WHERE deleted_at IS NULL;

-- User relationships (follows)
CREATE INDEX IF NOT EXISTS idx_user_relationships_follower ON user_relationships(follower_id, relationship_type) 
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_user_relationships_followed ON user_relationships(followed_id, relationship_type) 
    WHERE deleted_at IS NULL;

-- ========================================
-- 8. Analytics and Reporting Indexes
-- ========================================

-- Analytics metrics for dashboards
CREATE INDEX IF NOT EXISTS idx_analytics_metrics_type_period ON analytics_metrics(metric_type, period_start DESC);
CREATE INDEX IF NOT EXISTS idx_user_analytics_user_metric ON user_analytics(user_id, metric_type, created_at DESC);

-- System analytics
CREATE INDEX IF NOT EXISTS idx_system_analytics_category_created ON system_analytics(category, created_at DESC);

-- ========================================
-- 9. AI System Indexes
-- ========================================

-- AI matches
CREATE INDEX IF NOT EXISTS idx_ai_matches_user_status ON ai_matches(user_id, status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_ai_matches_matched_user ON ai_matches(matched_user_id, status);

-- AI curations
CREATE INDEX IF NOT EXISTS idx_ai_curations_target ON ai_curations(target_type, target_id) WHERE status = 'active';

-- ========================================
-- 10. Performance Views
-- ========================================

-- Create materialized view for user statistics
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_user_stats AS
SELECT 
    u.id as user_id,
    u.username,
    COUNT(DISTINCT l.id) FILTER (WHERE l.sender_id = u.id) as letters_sent,
    COUNT(DISTINCT l.id) FILTER (WHERE l.receiver_id = u.id) as letters_received,
    COALESCE(uc.balance, 0) as credit_balance,
    COUNT(DISTINCT mi.id) as museum_interactions,
    COUNT(DISTINCT f.id) as followers_count,
    COUNT(DISTINCT fg.id) as following_count
FROM users u
LEFT JOIN letters l ON (l.sender_id = u.id OR l.receiver_id = u.id) AND l.deleted_at IS NULL
LEFT JOIN user_credits uc ON uc.user_id = u.id
LEFT JOIN museum_interactions mi ON mi.user_id = u.id
LEFT JOIN user_relationships f ON f.followed_id = u.id AND f.deleted_at IS NULL
LEFT JOIN user_relationships fg ON fg.follower_id = u.id AND fg.deleted_at IS NULL
WHERE u.deleted_at IS NULL
GROUP BY u.id, u.username, uc.balance;

-- Index for the materialized view
CREATE UNIQUE INDEX IF NOT EXISTS idx_mv_user_stats_user_id ON mv_user_stats(user_id);

-- Refresh function for materialized view
CREATE OR REPLACE FUNCTION refresh_user_stats()
RETURNS void AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY mv_user_stats;
END;
$$ LANGUAGE plpgsql;

-- ========================================
-- 11. Cleanup and Statistics Update
-- ========================================

-- Update table statistics for query planner
ANALYZE users;
ANALYZE letters;
ANALYZE user_credits;
ANALYZE credit_transactions;
ANALYZE couriers;
ANALYZE tasks;
ANALYZE museum_entries;
ANALYZE notifications;

-- Show index usage statistics
SELECT 
    schemaname,
    tablename,
    indexname,
    idx_scan,
    idx_tup_read,
    idx_tup_fetch
FROM pg_stat_user_indexes
WHERE schemaname = 'public'
ORDER BY idx_scan DESC
LIMIT 20;