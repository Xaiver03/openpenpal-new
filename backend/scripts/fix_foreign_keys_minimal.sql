-- OpenPenPal Minimal Foreign Key Fix
-- Conservative approach - only critical fixes
-- ========================================

-- 1. Add missing unique constraints (safe operation)
-- ========================================

-- These won't fail if duplicates exist, just skip
DO $$
BEGIN
    -- Try to add unique constraints
    BEGIN
        ALTER TABLE users ADD CONSTRAINT unique_username UNIQUE (username);
    EXCEPTION WHEN others THEN
        RAISE NOTICE 'unique_username constraint already exists or has duplicates';
    END;
    
    BEGIN
        ALTER TABLE users ADD CONSTRAINT unique_email UNIQUE (email);
    EXCEPTION WHEN others THEN
        RAISE NOTICE 'unique_email constraint already exists or has duplicates';
    END;
    
    BEGIN
        ALTER TABLE letter_codes ADD CONSTRAINT unique_letter_code UNIQUE (code);
    EXCEPTION WHEN others THEN
        RAISE NOTICE 'unique_letter_code constraint already exists or has duplicates';
    END;
END $$;

-- 2. Add missing foreign keys for credit system (critical for data integrity)
-- ========================================

-- Only add if they don't exist
DO $$
BEGIN
    -- Credit transactions to users
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint 
        WHERE conname = 'fk_credit_transactions_user_id'
    ) THEN
        ALTER TABLE credit_transactions
        ADD CONSTRAINT fk_credit_transactions_user_id
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT;
    END IF;
    
    -- Credit transfers
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint 
        WHERE conname = 'fk_credit_transfers_from_user_id'
    ) THEN
        ALTER TABLE credit_transfers
        ADD CONSTRAINT fk_credit_transfers_from_user_id
        FOREIGN KEY (from_user_id) REFERENCES users(id) ON DELETE RESTRICT;
    END IF;
    
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint 
        WHERE conname = 'fk_credit_transfers_to_user_id'
    ) THEN
        ALTER TABLE credit_transfers
        ADD CONSTRAINT fk_credit_transfers_to_user_id
        FOREIGN KEY (to_user_id) REFERENCES users(id) ON DELETE RESTRICT;
    END IF;
    
    -- User credits
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint 
        WHERE conname = 'fk_user_credits_user_id'
    ) THEN
        ALTER TABLE user_credits
        ADD CONSTRAINT fk_user_credits_user_id
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
    END IF;
END $$;

-- 3. Add indexes for foreign keys (improves performance)
-- ========================================

-- Credit system indexes
CREATE INDEX IF NOT EXISTS idx_credit_transactions_user_id ON credit_transactions(user_id);
CREATE INDEX IF NOT EXISTS idx_credit_transfers_from_user_id ON credit_transfers(from_user_id);
CREATE INDEX IF NOT EXISTS idx_credit_transfers_to_user_id ON credit_transfers(to_user_id);
CREATE INDEX IF NOT EXISTS idx_user_credits_user_id ON user_credits(user_id);

-- Museum system indexes
CREATE INDEX IF NOT EXISTS idx_museum_entries_user_id ON museum_entries(user_id);
CREATE INDEX IF NOT EXISTS idx_museum_entries_letter_id ON museum_entries(letter_id);

-- Notification system indexes
CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);

-- 4. Data consistency report (no changes, just analysis)
-- ========================================

\echo ''
\echo 'Data Consistency Report:'
\echo '========================'

-- Check for orphaned records
WITH orphan_check AS (
    SELECT 'credit_transactions' as table_name, COUNT(*) as orphans
    FROM credit_transactions ct
    WHERE NOT EXISTS (SELECT 1 FROM users u WHERE u.id = ct.user_id)
    UNION ALL
    SELECT 'museum_entries', COUNT(*)
    FROM museum_entries me
    WHERE me.user_id IS NOT NULL 
    AND NOT EXISTS (SELECT 1 FROM users u WHERE u.id = me.user_id)
    UNION ALL
    SELECT 'user_profiles', COUNT(*)
    FROM user_profiles up
    WHERE NOT EXISTS (SELECT 1 FROM users u WHERE u.id = up.user_id)
    UNION ALL
    SELECT 'letters', COUNT(*)
    FROM letters l
    WHERE l.user_id IS NOT NULL 
    AND NOT EXISTS (SELECT 1 FROM users u WHERE u.id = l.user_id)
)
SELECT * FROM orphan_check WHERE orphans > 0;

-- Summary stats
\echo ''
\echo 'Database Statistics:'
\echo '-------------------'
SELECT 
    (SELECT COUNT(*) FROM pg_constraint WHERE contype = 'f') as total_foreign_keys,
    (SELECT COUNT(*) FROM pg_indexes WHERE schemaname = 'public') as total_indexes,
    (SELECT COUNT(*) FROM pg_constraint WHERE contype = 'f' AND confdeltype = 'c') as cascade_deletes;

-- High risk cascade analysis
\echo ''
\echo 'High Risk CASCADE Deletes:'
\echo '-------------------------'
SELECT 
    conname AS constraint_name,
    conrelid::regclass AS table_name,
    confrelid::regclass AS referenced_table
FROM pg_constraint
WHERE contype = 'f' AND confdeltype = 'c'
AND conrelid::regclass::text IN (
    'credit_transactions', 'credit_transfers', 'user_credits',
    'letters', 'letter_codes', 'museum_entries'
)
ORDER BY table_name;