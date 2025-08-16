-- Analyze Foreign Key Constraints and Data Consistency
-- OpenPenPal Database

-- ========================================
-- 1. List all existing foreign key constraints
-- ========================================

SELECT 
    conname AS constraint_name,
    conrelid::regclass AS table_name,
    confrelid::regclass AS referenced_table,
    array_to_string(ARRAY(
        SELECT attname 
        FROM pg_attribute 
        WHERE attrelid = conrelid AND attnum = ANY(conkey)
    ), ', ') AS columns,
    array_to_string(ARRAY(
        SELECT attname 
        FROM pg_attribute 
        WHERE attrelid = confrelid AND attnum = ANY(confkey)
    ), ', ') AS referenced_columns,
    confdeltype AS delete_action,
    confupdtype AS update_action
FROM pg_constraint
WHERE contype = 'f'
ORDER BY conrelid::regclass::text, conname;

-- ========================================
-- 2. Find tables that should have foreign keys but don't
-- ========================================

WITH potential_fks AS (
    SELECT 
        table_name,
        column_name,
        CASE 
            WHEN column_name LIKE '%_id' AND column_name != 'id' THEN 
                regexp_replace(column_name, '_id$', 's')
            WHEN column_name = 'user_id' THEN 'users'
            WHEN column_name = 'school_id' THEN 'schools'
            WHEN column_name = 'letter_id' THEN 'letters'
            WHEN column_name = 'design_id' THEN 'envelope_designs'
            WHEN column_name = 'courier_id' THEN 'couriers'
            WHEN column_name = 'entry_id' THEN 'museum_entries'
            WHEN column_name = 'product_id' THEN 'products'
            WHEN column_name = 'order_id' THEN 'orders'
            ELSE NULL
        END AS likely_ref_table
    FROM information_schema.columns
    WHERE table_schema = 'public'
    AND (column_name LIKE '%_id' OR column_name IN ('parent_id', 'sender_id', 'receiver_id'))
    AND column_name != 'id'
)
SELECT 
    pf.table_name,
    pf.column_name,
    pf.likely_ref_table,
    CASE 
        WHEN c.constraint_name IS NULL THEN 'MISSING FK'
        ELSE 'Has FK'
    END AS fk_status
FROM potential_fks pf
LEFT JOIN (
    SELECT 
        tc.table_name,
        kcu.column_name,
        tc.constraint_name
    FROM information_schema.table_constraints tc
    JOIN information_schema.key_column_usage kcu 
        ON tc.constraint_name = kcu.constraint_name
    WHERE tc.constraint_type = 'FOREIGN KEY'
    AND tc.table_schema = 'public'
) c ON pf.table_name = c.table_name AND pf.column_name = c.column_name
WHERE pf.likely_ref_table IS NOT NULL
AND c.constraint_name IS NULL
ORDER BY pf.table_name, pf.column_name;

-- ========================================
-- 3. Check for orphaned records (data consistency)
-- ========================================

-- Check user_profiles referencing non-existent users
SELECT 'user_profiles' as table_name, COUNT(*) as orphaned_records
FROM user_profiles up
WHERE NOT EXISTS (SELECT 1 FROM users u WHERE u.id = up.user_id)
UNION ALL
-- Check letters referencing non-existent users
SELECT 'letters.sender_id' as table_name, COUNT(*) as orphaned_records
FROM letters l
WHERE l.sender_id IS NOT NULL 
AND NOT EXISTS (SELECT 1 FROM users u WHERE u.id = l.sender_id)
UNION ALL
SELECT 'letters.receiver_id' as table_name, COUNT(*) as orphaned_records
FROM letters l
WHERE l.receiver_id IS NOT NULL 
AND NOT EXISTS (SELECT 1 FROM users u WHERE u.id = l.receiver_id)
UNION ALL
-- Check credit_transactions referencing non-existent users
SELECT 'credit_transactions' as table_name, COUNT(*) as orphaned_records
FROM credit_transactions ct
WHERE NOT EXISTS (SELECT 1 FROM users u WHERE u.id = ct.user_id)
UNION ALL
-- Check museum_entries referencing non-existent users
SELECT 'museum_entries' as table_name, COUNT(*) as orphaned_records
FROM museum_entries me
WHERE NOT EXISTS (SELECT 1 FROM users u WHERE u.id = me.user_id)
UNION ALL
-- Check couriers referencing non-existent users
SELECT 'couriers' as table_name, COUNT(*) as orphaned_records
FROM couriers c
WHERE NOT EXISTS (SELECT 1 FROM users u WHERE u.id = c.user_id)
ORDER BY orphaned_records DESC;

-- ========================================
-- 4. Find columns that might need unique constraints
-- ========================================

WITH duplicate_check AS (
    SELECT 'users.username' as column_check, 
           COUNT(*) - COUNT(DISTINCT username) as duplicates
    FROM users WHERE username IS NOT NULL
    UNION ALL
    SELECT 'users.email' as column_check, 
           COUNT(*) - COUNT(DISTINCT email) as duplicates
    FROM users WHERE email IS NOT NULL
    UNION ALL
    SELECT 'letter_codes.code' as column_check,
           COUNT(*) - COUNT(DISTINCT code) as duplicates
    FROM letter_codes WHERE code IS NOT NULL
    UNION ALL
    SELECT 'schools.code' as column_check,
           COUNT(*) - COUNT(DISTINCT code) as duplicates
    FROM schools WHERE code IS NOT NULL
)
SELECT * FROM duplicate_check WHERE duplicates > 0;

-- ========================================
-- 5. Analyze cascade delete risks
-- ========================================

WITH fk_analysis AS (
    SELECT 
        conname AS constraint_name,
        conrelid::regclass AS table_name,
        confrelid::regclass AS referenced_table,
        CASE confdeltype
            WHEN 'a' THEN 'NO ACTION'
            WHEN 'r' THEN 'RESTRICT'
            WHEN 'c' THEN 'CASCADE'
            WHEN 'n' THEN 'SET NULL'
            WHEN 'd' THEN 'SET DEFAULT'
        END AS delete_action
    FROM pg_constraint
    WHERE contype = 'f'
)
SELECT 
    constraint_name,
    table_name,
    referenced_table,
    delete_action,
    CASE 
        WHEN delete_action = 'CASCADE' THEN 'HIGH RISK'
        WHEN delete_action = 'SET NULL' THEN 'MEDIUM RISK'
        ELSE 'LOW RISK'
    END as risk_level
FROM fk_analysis
ORDER BY risk_level DESC, table_name;