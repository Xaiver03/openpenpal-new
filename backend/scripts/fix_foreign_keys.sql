-- OpenPenPal Database Foreign Key Fix Script
-- Based on ULTRATHINK analysis of database structure
-- ========================================

-- 1. Fix orphaned records first (data consistency)
-- ========================================

-- Fix orphaned credit_transactions
DELETE FROM credit_transactions ct
WHERE NOT EXISTS (SELECT 1 FROM users u WHERE u.id = ct.user_id);

-- Fix orphaned museum_entries
DELETE FROM museum_entries me
WHERE me.user_id IS NOT NULL 
AND NOT EXISTS (SELECT 1 FROM users u WHERE u.id = me.user_id);

-- Fix orphaned letters (use user_id not sender_id)
DELETE FROM letters l
WHERE l.user_id IS NOT NULL 
AND NOT EXISTS (SELECT 1 FROM users u WHERE u.id = l.user_id);

-- Fix orphaned courier tasks
DELETE FROM courier_tasks ct
WHERE ct.courier_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM users u WHERE u.id = ct.courier_id);

-- Fix orphaned user_profiles
DELETE FROM user_profiles up
WHERE NOT EXISTS (SELECT 1 FROM users u WHERE u.id = up.user_id);

-- Fix orphaned credit_activity_participations
DELETE FROM credit_activity_participations cap
WHERE cap.user_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM users u WHERE u.id = cap.user_id);

-- Fix orphaned credit_transfers
DELETE FROM credit_transfers ct
WHERE ct.from_user_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM users u WHERE u.id = ct.from_user_id);

DELETE FROM credit_transfers ct
WHERE ct.to_user_id IS NOT NULL
AND NOT EXISTS (SELECT 1 FROM users u WHERE u.id = ct.to_user_id);

-- ========================================
-- 2. Add missing critical foreign keys
-- ========================================

-- AI related tables - add foreign keys to users
ALTER TABLE ai_inspirations 
ADD CONSTRAINT fk_ai_inspirations_user_id 
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL;

ALTER TABLE ai_reply_advices
ADD CONSTRAINT fk_ai_reply_advices_user_id
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL;

ALTER TABLE ai_reply_advices
ADD CONSTRAINT fk_ai_reply_advices_letter_id
FOREIGN KEY (letter_id) REFERENCES letters(id) ON DELETE CASCADE;

ALTER TABLE ai_usage_logs
ADD CONSTRAINT fk_ai_usage_logs_user_id
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL;

-- Museum related tables
ALTER TABLE museum_entries
ADD CONSTRAINT fk_museum_entries_user_id
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL;

ALTER TABLE museum_entries
ADD CONSTRAINT fk_museum_entries_letter_id
FOREIGN KEY (letter_id) REFERENCES letters(id) ON DELETE CASCADE;

ALTER TABLE museum_interactions
ADD CONSTRAINT fk_museum_interactions_user_id
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE museum_reactions
ADD CONSTRAINT fk_museum_reactions_user_id
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- Credit system tables
ALTER TABLE credit_transactions
ADD CONSTRAINT fk_credit_transactions_user_id
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE credit_transfers
ADD CONSTRAINT fk_credit_transfers_from_user_id
FOREIGN KEY (from_user_id) REFERENCES users(id) ON DELETE RESTRICT;

ALTER TABLE credit_transfers
ADD CONSTRAINT fk_credit_transfers_to_user_id
FOREIGN KEY (to_user_id) REFERENCES users(id) ON DELETE RESTRICT;

ALTER TABLE credit_transfer_limits
ADD CONSTRAINT fk_credit_transfer_limits_user_id
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE credit_transfer_notifications
ADD CONSTRAINT fk_credit_transfer_notifications_user_id
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE credit_expiration_logs
ADD CONSTRAINT fk_credit_expiration_logs_user_id
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE credit_expiration_notifications
ADD CONSTRAINT fk_credit_expiration_notifications_user_id
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- Envelope related tables
ALTER TABLE envelopes
ADD CONSTRAINT fk_envelopes_user_id
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL;

ALTER TABLE envelope_orders
ADD CONSTRAINT fk_envelope_orders_user_id
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE envelope_votes
ADD CONSTRAINT fk_envelope_votes_user_id
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- Notification tables
ALTER TABLE notifications
ADD CONSTRAINT fk_notifications_user_id
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE notification_preferences
ADD CONSTRAINT fk_notification_preferences_user_id
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- School admin tables
ALTER TABLE school_admins
ADD CONSTRAINT fk_school_admins_user_id
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- User credit tables
ALTER TABLE user_credits
ADD CONSTRAINT fk_user_credits_user_id
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE user_daily_usages
ADD CONSTRAINT fk_user_daily_usages_user_id
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE user_analytics
ADD CONSTRAINT fk_user_analytics_user_id
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- ========================================
-- 3. Fix high-risk CASCADE deletes (change to RESTRICT/SET NULL)
-- ========================================

-- Change dangerous cascades to safer options
-- Keep CASCADE only for genuine child records

-- Change letter-related cascades to be more conservative
ALTER TABLE letter_codes DROP CONSTRAINT IF EXISTS fk_letters_code;
ALTER TABLE letter_codes
ADD CONSTRAINT fk_letters_code
FOREIGN KEY (letter_id) REFERENCES letters(id) ON DELETE RESTRICT;

-- Change user deletion cascades for important data
ALTER TABLE credit_transactions DROP CONSTRAINT IF EXISTS fk_credit_transactions_user_id;
ALTER TABLE credit_transactions
ADD CONSTRAINT fk_credit_transactions_user_id
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT;

-- ========================================
-- 4. Add unique constraints where needed
-- ========================================

-- Ensure unique usernames and emails
ALTER TABLE users ADD CONSTRAINT unique_username UNIQUE (username);
ALTER TABLE users ADD CONSTRAINT unique_email UNIQUE (email);

-- Ensure unique letter codes
ALTER TABLE letter_codes ADD CONSTRAINT unique_letter_code UNIQUE (code);

-- Ensure unique signal codes
ALTER TABLE signal_codes ADD CONSTRAINT unique_signal_code UNIQUE (code);

-- ========================================
-- 5. Create indexes for foreign key columns
-- ========================================

-- Index all foreign key columns that don't have indexes
CREATE INDEX IF NOT EXISTS idx_ai_inspirations_user_id ON ai_inspirations(user_id);
CREATE INDEX IF NOT EXISTS idx_ai_reply_advices_user_id ON ai_reply_advices(user_id);
CREATE INDEX IF NOT EXISTS idx_ai_reply_advices_letter_id ON ai_reply_advices(letter_id);
CREATE INDEX IF NOT EXISTS idx_ai_usage_logs_user_id ON ai_usage_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_museum_entries_user_id ON museum_entries(user_id);
CREATE INDEX IF NOT EXISTS idx_museum_entries_letter_id ON museum_entries(letter_id);
CREATE INDEX IF NOT EXISTS idx_museum_interactions_user_id ON museum_interactions(user_id);
CREATE INDEX IF NOT EXISTS idx_museum_reactions_user_id ON museum_reactions(user_id);
CREATE INDEX IF NOT EXISTS idx_credit_transfers_from_user_id ON credit_transfers(from_user_id);
CREATE INDEX IF NOT EXISTS idx_credit_transfers_to_user_id ON credit_transfers(to_user_id);
CREATE INDEX IF NOT EXISTS idx_credit_transfer_limits_user_id ON credit_transfer_limits(user_id);
CREATE INDEX IF NOT EXISTS idx_credit_transfer_notifications_user_id ON credit_transfer_notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_credit_expiration_logs_user_id ON credit_expiration_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_credit_expiration_notifications_user_id ON credit_expiration_notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_envelopes_user_id ON envelopes(user_id);
CREATE INDEX IF NOT EXISTS idx_envelope_orders_user_id ON envelope_orders(user_id);
CREATE INDEX IF NOT EXISTS idx_envelope_votes_user_id ON envelope_votes(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_notification_preferences_user_id ON notification_preferences(user_id);
CREATE INDEX IF NOT EXISTS idx_school_admins_user_id ON school_admins(user_id);
CREATE INDEX IF NOT EXISTS idx_user_credits_user_id ON user_credits(user_id);
CREATE INDEX IF NOT EXISTS idx_user_daily_usages_user_id ON user_daily_usages(user_id);
CREATE INDEX IF NOT EXISTS idx_user_analytics_user_id ON user_analytics(user_id);

-- ========================================
-- 6. Summary report
-- ========================================

\echo 'Foreign Key Fix Summary:'
\echo '========================'

-- Count foreign keys after fixes
SELECT COUNT(*) as total_foreign_keys
FROM pg_constraint
WHERE contype = 'f';

-- Count indexes
SELECT COUNT(*) as total_indexes
FROM pg_indexes
WHERE schemaname = 'public';

-- High risk cascades remaining
SELECT COUNT(*) as high_risk_cascades
FROM pg_constraint
WHERE contype = 'f' AND confdeltype = 'c';