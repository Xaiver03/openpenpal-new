#!/bin/bash

# Diagnose database migration issues
# This script helps identify constraint naming conflicts between GORM and PostgreSQL

set -e

echo "=========================================="
echo "Database Migration Diagnostics"
echo "=========================================="
echo ""

# Database connection parameters
DB_NAME="${DATABASE_NAME:-openpenpal}"
DB_USER="${DATABASE_USER:-$(whoami)}"
DB_HOST="${DATABASE_HOST:-localhost}"
DB_PORT="${DATABASE_PORT:-5432}"

echo "Database Configuration:"
echo "----------------------"
echo "Database: $DB_NAME"
echo "User: $DB_USER"
echo "Host: $DB_HOST"
echo "Port: $DB_PORT"
echo ""

# Check if database exists
echo "1. Checking database connection..."
echo "--------------------------------"
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT version();" || {
    echo "ERROR: Cannot connect to database"
    exit 1
}

# Check users table structure
echo ""
echo "2. Users table structure:"
echo "------------------------"
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "\d users" || {
    echo "ERROR: Users table does not exist"
}

# Check all constraints on users table
echo ""
echo "3. All constraints on users table:"
echo "---------------------------------"
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
SELECT 
    con.conname as constraint_name,
    CASE con.contype
        WHEN 'p' THEN 'PRIMARY KEY'
        WHEN 'u' THEN 'UNIQUE'
        WHEN 'c' THEN 'CHECK'
        WHEN 'f' THEN 'FOREIGN KEY'
    END as constraint_type,
    pg_get_constraintdef(con.oid) as definition
FROM pg_constraint con
INNER JOIN pg_namespace nsp ON nsp.oid = con.connamespace
INNER JOIN pg_class cls ON cls.oid = con.conrelid
WHERE cls.relname = 'users'
AND nsp.nspname = 'public'
ORDER BY con.conname;
"

# Check indexes on users table
echo ""
echo "4. All indexes on users table:"
echo "-----------------------------"
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
SELECT 
    indexname,
    indexdef
FROM pg_indexes
WHERE tablename = 'users'
AND schemaname = 'public'
ORDER BY indexname;
"

# Check for specific constraint names GORM might be looking for
echo ""
echo "5. Checking for GORM expected constraint names:"
echo "----------------------------------------------"
for constraint in "uni_users_username" "uni_users_email" "unique_username" "unique_email"; do
    echo -n "Checking for '$constraint': "
    count=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -tAc "
    SELECT COUNT(*) FROM pg_constraint 
    WHERE conname = '$constraint' 
    AND conrelid = 'public.users'::regclass;
    ")
    if [ "$count" = "1" ]; then
        echo "FOUND"
    else
        echo "NOT FOUND"
    fi
done

# Check for foreign keys referencing users table
echo ""
echo "6. Foreign keys referencing users table:"
echo "---------------------------------------"
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
SELECT 
    conrelid::regclass AS table_name,
    conname AS constraint_name,
    pg_get_constraintdef(oid) AS definition
FROM pg_constraint
WHERE confrelid = 'public.users'::regclass
AND contype = 'f'
LIMIT 10;
"

# Get GORM migration SQL (dry run)
echo ""
echo "7. Testing GORM migration (dry run):"
echo "-----------------------------------"
echo "To see what GORM is trying to do, run the backend with:"
echo "  GORM_LOG_LEVEL=info go run main.go"
echo ""

# Provide recommendations
echo ""
echo "=========================================="
echo "Recommendations:"
echo "=========================================="
echo ""
echo "1. If you see 'constraint does not exist' errors:"
echo "   - GORM is looking for constraints with different names"
echo "   - Use the custom naming strategy in gorm_naming_strategy.go"
echo ""
echo "2. Quick fixes:"
echo "   a) Use SKIP_DB_MIGRATION=true to bypass migrations (temporary)"
echo "   b) Run ./scripts/fix-constraint-names.sh to rename constraints"
echo "   c) Update backend code to use custom naming strategy (permanent)"
echo ""
echo "3. To test the fix:"
echo "   cd backend && go run main.go"
echo ""

# Create a summary report
echo ""
echo "Creating diagnostic report..."
cat > /tmp/migration_diagnostics.txt << EOF
Migration Diagnostics Report
Generated: $(date)

Database: $DB_NAME
User: $DB_USER

Constraints Found:
$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -tAc "
SELECT conname FROM pg_constraint 
WHERE conrelid = 'public.users'::regclass
AND contype IN ('u', 'p')
ORDER BY conname;
")

Indexes Found:
$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -tAc "
SELECT indexname FROM pg_indexes
WHERE tablename = 'users'
AND schemaname = 'public';
")

GORM Expected Constraints:
- uni_users_username (for username unique constraint)
- uni_users_email (for email unique constraint)

Current Status:
- If constraints have different names, GORM migration will fail
- Solution: Use custom naming strategy or rename constraints
EOF

echo "Diagnostic report saved to: /tmp/migration_diagnostics.txt"
echo ""
echo "Diagnostics complete!"