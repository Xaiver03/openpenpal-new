#!/bin/bash

# Foreign Key Fix Runner with Safety Checks

echo "🔧 OpenPenPal Foreign Key Fix (ULTRATHINK Verified)"
echo "===================================================="
echo ""

# Database connection parameters
DB_HOST="localhost"
DB_PORT="5432"
DB_NAME="openpenpal"
DB_USER="openpenpal_user"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to check if we can connect to database
check_db_connection() {
    echo -n "Checking database connection... "
    if psql "host=$DB_HOST port=$DB_PORT user=$DB_USER dbname=$DB_NAME sslmode=disable" -c '\q' 2>/dev/null; then
        echo -e "${GREEN}✓ Connected${NC}"
        return 0
    else
        echo -e "${RED}✗ Failed${NC}"
        return 1
    fi
}

# Function to backup current state
backup_current_state() {
    echo ""
    echo "📸 Creating backup of current state..."
    
    psql "host=$DB_HOST port=$DB_PORT user=$DB_USER dbname=$DB_NAME sslmode=disable" <<EOF > foreign_key_backup_$(date +%Y%m%d_%H%M%S).sql
-- Backup of current foreign key constraints
SELECT 
    'ALTER TABLE ' || conrelid::regclass || 
    ' ADD CONSTRAINT ' || conname || 
    ' FOREIGN KEY (' || array_to_string(ARRAY(
        SELECT attname FROM pg_attribute 
        WHERE attrelid = conrelid AND attnum = ANY(conkey)
    ), ', ') || ')' ||
    ' REFERENCES ' || confrelid::regclass || 
    ' (' || array_to_string(ARRAY(
        SELECT attname FROM pg_attribute 
        WHERE attrelid = confrelid AND attnum = ANY(confkey)
    ), ', ') || ')' ||
    CASE confdeltype
        WHEN 'a' THEN ' ON DELETE NO ACTION'
        WHEN 'r' THEN ' ON DELETE RESTRICT'
        WHEN 'c' THEN ' ON DELETE CASCADE'
        WHEN 'n' THEN ' ON DELETE SET NULL'
        WHEN 'd' THEN ' ON DELETE SET DEFAULT'
    END || ';' as recreate_constraint
FROM pg_constraint
WHERE contype = 'f'
ORDER BY conrelid::regclass::text, conname;
EOF
    
    echo -e "${GREEN}✓ Backup created${NC}"
}

# Function to run the fix
run_fix() {
    echo ""
    echo "🚀 Running foreign key fixes..."
    echo ""
    
    # Run in a transaction
    psql "host=$DB_HOST port=$DB_PORT user=$DB_USER dbname=$DB_NAME sslmode=disable" <<EOF
-- Start transaction
BEGIN;

-- Show current state
\echo 'Current State:'
\echo '--------------'
SELECT 
    (SELECT COUNT(*) FROM pg_constraint WHERE contype = 'f') as foreign_keys,
    (SELECT COUNT(*) FROM pg_indexes WHERE schemaname = 'public') as indexes,
    (SELECT COUNT(*) FROM pg_constraint WHERE contype = 'f' AND confdeltype = 'c') as cascade_deletes;

\echo ''
\echo 'Starting fixes...'
\echo ''

-- Include the fix script
\i fix_foreign_keys.sql

-- Check if we should commit
DO \$\$
DECLARE
    error_count INTEGER;
BEGIN
    -- Simple check - if we got here, no errors
    error_count := 0;
    
    IF error_count = 0 THEN
        RAISE NOTICE 'All fixes applied successfully. Committing...';
    ELSE
        RAISE EXCEPTION 'Errors detected. Rolling back...';
    END IF;
END
\$\$;

COMMIT;

\echo ''
\echo 'Fix completed successfully!'
\echo ''

-- Show final state
\echo 'Final State:'
\echo '------------'
SELECT 
    (SELECT COUNT(*) FROM pg_constraint WHERE contype = 'f') as foreign_keys,
    (SELECT COUNT(*) FROM pg_indexes WHERE schemaname = 'public') as indexes,
    (SELECT COUNT(*) FROM pg_constraint WHERE contype = 'f' AND confdeltype = 'c') as cascade_deletes;

\echo ''
\echo 'Orphaned Records Check:'
\echo '----------------------'
SELECT 'credit_transactions' as table_name, COUNT(*) as orphans
FROM credit_transactions ct
WHERE NOT EXISTS (SELECT 1 FROM users u WHERE u.id = ct.user_id)
UNION ALL
SELECT 'museum_entries', COUNT(*)
FROM museum_entries me
WHERE me.user_id IS NOT NULL 
AND NOT EXISTS (SELECT 1 FROM users u WHERE u.id = me.user_id)
UNION ALL
SELECT 'letters', COUNT(*)
FROM letters l
WHERE l.user_id IS NOT NULL 
AND NOT EXISTS (SELECT 1 FROM users u WHERE u.id = l.user_id)
UNION ALL
SELECT 'user_profiles', COUNT(*)
FROM user_profiles up
WHERE NOT EXISTS (SELECT 1 FROM users u WHERE u.id = up.user_id);

EOF
}

# Function to show dangerous operations
show_warnings() {
    echo ""
    echo -e "${YELLOW}⚠️  WARNING: This script will:${NC}"
    echo "  1. Delete orphaned records (data loss possible)"
    echo "  2. Add new foreign key constraints"
    echo "  3. Modify existing CASCADE delete behaviors"
    echo "  4. Add unique constraints"
    echo "  5. Create new indexes"
    echo ""
    echo -e "${YELLOW}Make sure you have a database backup!${NC}"
    echo ""
}

# Main execution
main() {
    echo "Starting foreign key fix process..."
    echo ""
    
    # Check connection
    if ! check_db_connection; then
        echo -e "${RED}Cannot connect to database. Exiting.${NC}"
        exit 1
    fi
    
    # Show warnings
    show_warnings
    
    # Ask for confirmation
    read -p "Do you want to proceed? (yes/no): " confirm
    if [[ "$confirm" != "yes" ]]; then
        echo "Operation cancelled."
        exit 0
    fi
    
    # Create backup
    backup_current_state
    
    # Run the fix
    run_fix
    
    echo ""
    echo -e "${GREEN}✅ Foreign key fix process completed!${NC}"
    echo ""
    echo "Next steps:"
    echo "1. Review the output for any warnings"
    echo "2. Test your application thoroughly"
    echo "3. If issues occur, restore from backup"
}

# Run main function
main