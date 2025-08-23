#!/bin/bash

# Script to fix courier levels in the database
# The level field should match the role suffix

echo "Fixing courier levels in database..."

# Database connection details
DB_NAME="${DATABASE_NAME:-openpenpal}"
DB_HOST="${DATABASE_HOST:-localhost}"
DB_PORT="${DATABASE_PORT:-5432}"
DB_USER="${DATABASE_USER:-$(whoami)}"

# SQL to update courier levels based on their role
SQL_SCRIPT=$(cat <<'EOF'
-- Update L4 couriers
UPDATE couriers c
SET level = 4
FROM users u
WHERE c.user_id = u.id 
AND u.role = 'courier_level4';

-- Update L3 couriers  
UPDATE couriers c
SET level = 3
FROM users u
WHERE c.user_id = u.id 
AND u.role = 'courier_level3';

-- Update L2 couriers
UPDATE couriers c
SET level = 2
FROM users u
WHERE c.user_id = u.id 
AND u.role = 'courier_level2';

-- Update L1 couriers
UPDATE couriers c
SET level = 1
FROM users u
WHERE c.user_id = u.id 
AND u.role = 'courier_level1';

-- Show the updated couriers
SELECT u.username, u.role, c.level, c.school, c.zone, c.managed_op_code_prefix
FROM couriers c
JOIN users u ON c.user_id = u.id
WHERE u.role LIKE 'courier%'
ORDER BY c.level DESC;
EOF
)

# Execute the SQL
echo "$SQL_SCRIPT" | psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME"

echo "Done! Courier levels have been updated to match their roles."