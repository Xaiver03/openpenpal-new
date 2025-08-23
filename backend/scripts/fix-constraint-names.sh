#!/bin/bash

# Fix database constraint naming mismatch between GORM and actual database
# This script renames constraints to match GORM's expectations

set -e

echo "================================"
echo "Constraint Name Fix Script"
echo "================================"
echo ""

# Database connection parameters
DB_NAME="${DATABASE_NAME:-openpenpal}"
DB_USER="${DATABASE_USER:-$(whoami)}"
DB_HOST="${DATABASE_HOST:-localhost}"
DB_PORT="${DATABASE_PORT:-5432}"

echo "Database: $DB_NAME"
echo "User: $DB_USER"
echo "Host: $DB_HOST"
echo "Port: $DB_PORT"
echo ""

# Function to execute SQL
execute_sql() {
    local sql="$1"
    echo "Executing: $sql"
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "$sql" || {
        echo "Failed to execute SQL. Continuing..."
        return 1
    }
}

echo "Step 1: Checking current constraints on users table..."
echo "-----------------------------------------------"
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "\d users" | grep -i constraint || true

echo ""
echo "Step 2: Creating backup of constraints..."
echo "-----------------------------------------------"

# Backup constraint definitions
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
SELECT 
    con.conname as constraint_name,
    con.contype as constraint_type,
    pg_get_constraintdef(con.oid) as definition
FROM pg_constraint con
INNER JOIN pg_namespace nsp ON nsp.oid = con.connamespace
INNER JOIN pg_class cls ON cls.oid = con.conrelid
WHERE cls.relname = 'users'
AND nsp.nspname = 'public'
AND con.contype IN ('u', 'p')
ORDER BY con.conname;
" > /tmp/users_constraints_backup.txt

echo "Backup saved to: /tmp/users_constraints_backup.txt"

echo ""
echo "Step 3: Renaming constraints to match GORM expectations..."
echo "-----------------------------------------------"

# Check if the constraints exist before trying to rename
echo "Checking if unique_username exists..."
HAS_UNIQUE_USERNAME=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -tAc "
SELECT COUNT(*) FROM pg_constraint 
WHERE conname = 'unique_username' 
AND conrelid = 'public.users'::regclass;
")

if [ "$HAS_UNIQUE_USERNAME" = "1" ]; then
    echo "Renaming unique_username to uni_users_username..."
    execute_sql "ALTER TABLE users RENAME CONSTRAINT unique_username TO uni_users_username;"
else
    echo "unique_username constraint not found, checking for uni_users_username..."
    HAS_UNI_USERNAME=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -tAc "
    SELECT COUNT(*) FROM pg_constraint 
    WHERE conname = 'uni_users_username' 
    AND conrelid = 'public.users'::regclass;
    ")
    if [ "$HAS_UNI_USERNAME" = "1" ]; then
        echo "uni_users_username already exists, no action needed"
    else
        echo "Neither constraint exists, this might indicate a different issue"
    fi
fi

echo "Checking if unique_email exists..."
HAS_UNIQUE_EMAIL=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -tAc "
SELECT COUNT(*) FROM pg_constraint 
WHERE conname = 'unique_email' 
AND conrelid = 'public.users'::regclass;
")

if [ "$HAS_UNIQUE_EMAIL" = "1" ]; then
    echo "Renaming unique_email to uni_users_email..."
    execute_sql "ALTER TABLE users RENAME CONSTRAINT unique_email TO uni_users_email;"
else
    echo "unique_email constraint not found, checking for uni_users_email..."
    HAS_UNI_EMAIL=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -tAc "
    SELECT COUNT(*) FROM pg_constraint 
    WHERE conname = 'uni_users_email' 
    AND conrelid = 'public.users'::regclass;
    ")
    if [ "$HAS_UNI_EMAIL" = "1" ]; then
        echo "uni_users_email already exists, no action needed"
    else
        echo "Neither constraint exists, this might indicate a different issue"
    fi
fi

echo ""
echo "Step 4: Verifying constraints after rename..."
echo "-----------------------------------------------"
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
SELECT 
    con.conname as constraint_name,
    con.contype as constraint_type,
    pg_get_constraintdef(con.oid) as definition
FROM pg_constraint con
INNER JOIN pg_namespace nsp ON nsp.oid = con.connamespace
INNER JOIN pg_class cls ON cls.oid = con.conrelid
WHERE cls.relname = 'users'
AND nsp.nspname = 'public'
AND con.contype IN ('u', 'p')
ORDER BY con.conname;
"

echo ""
echo "Step 5: Creating rollback script..."
echo "-----------------------------------------------"

cat > /tmp/rollback_constraints.sh << 'EOF'
#!/bin/bash
# Rollback script to restore original constraint names

DB_NAME="${DATABASE_NAME:-openpenpal}"
DB_USER="${DATABASE_USER:-$(whoami)}"
DB_HOST="${DATABASE_HOST:-localhost}"
DB_PORT="${DATABASE_PORT:-5432}"

echo "Rolling back constraint names..."

# Check and rollback uni_users_username to unique_username
HAS_UNI_USERNAME=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -tAc "
SELECT COUNT(*) FROM pg_constraint 
WHERE conname = 'uni_users_username' 
AND conrelid = 'public.users'::regclass;
")

if [ "$HAS_UNI_USERNAME" = "1" ]; then
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "ALTER TABLE users RENAME CONSTRAINT uni_users_username TO unique_username;"
    echo "Rolled back uni_users_username to unique_username"
fi

# Check and rollback uni_users_email to unique_email
HAS_UNI_EMAIL=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -tAc "
SELECT COUNT(*) FROM pg_constraint 
WHERE conname = 'uni_users_email' 
AND conrelid = 'public.users'::regclass;
")

if [ "$HAS_UNI_EMAIL" = "1" ]; then
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "ALTER TABLE users RENAME CONSTRAINT uni_users_email TO unique_email;"
    echo "Rolled back uni_users_email to unique_email"
fi

echo "Rollback complete!"
EOF

chmod +x /tmp/rollback_constraints.sh
echo "Rollback script created at: /tmp/rollback_constraints.sh"

echo ""
echo "================================"
echo "Script completed!"
echo "================================"
echo ""
echo "If you need to rollback these changes, run:"
echo "  /tmp/rollback_constraints.sh"
echo ""
echo "Now try running the backend without SKIP_DB_MIGRATION=true"