#!/bin/bash

# Script to fix ManagedOPCodePrefix for existing couriers
# This ensures Level 4 couriers can manage all OP codes

echo "Fixing ManagedOPCodePrefix for existing couriers..."

# Database connection details
DB_NAME="${DATABASE_NAME:-openpenpal}"
DB_HOST="${DATABASE_HOST:-localhost}"
DB_PORT="${DATABASE_PORT:-5432}"
DB_USER="${DATABASE_USER:-$(whoami)}"

# SQL to update courier managed prefixes based on their level
SQL_SCRIPT=$(cat <<'EOF'
-- Update Level 4 couriers to have empty prefix (manage all)
UPDATE couriers c
SET managed_op_code_prefix = ''
FROM users u
WHERE c.user_id = u.id 
AND u.role = 'courier_level4';

-- Update Level 3 couriers to manage school prefix (first 2 chars)
UPDATE couriers c
SET managed_op_code_prefix = 
  CASE 
    WHEN c.school = '北京大学' THEN 'PK'
    WHEN c.school = '清华大学' THEN 'QH'
    WHEN c.school = '北京交通大学' THEN 'BD'
    ELSE UPPER(SUBSTRING(c.school, 1, 2))
  END
FROM users u
WHERE c.user_id = u.id 
AND u.role = 'courier_level3'
AND (c.managed_op_code_prefix IS NULL OR c.managed_op_code_prefix = '');

-- Update Level 2 couriers to manage area prefix (first 4 chars)
UPDATE couriers c
SET managed_op_code_prefix = 
  CASE 
    WHEN c.zone LIKE 'BJDX%' THEN SUBSTRING(c.zone, 1, 4)
    WHEN c.zone LIKE 'PK%' THEN SUBSTRING(c.zone, 1, 4)
    ELSE UPPER(SUBSTRING(c.zone, 1, 4))
  END
FROM users u
WHERE c.user_id = u.id 
AND u.role = 'courier_level2'
AND (c.managed_op_code_prefix IS NULL OR c.managed_op_code_prefix = '');

-- Update Level 1 couriers to manage building prefix (full 6 chars if available)
UPDATE couriers c
SET managed_op_code_prefix = 
  CASE 
    WHEN LENGTH(c.zone) >= 6 THEN UPPER(SUBSTRING(c.zone, 1, 6))
    ELSE UPPER(c.zone)
  END
FROM users u
WHERE c.user_id = u.id 
AND u.role = 'courier_level1'
AND (c.managed_op_code_prefix IS NULL OR c.managed_op_code_prefix = '');

-- Show the updated couriers
SELECT c.id, u.username, u.role, c.level, c.school, c.zone, c.managed_op_code_prefix
FROM couriers c
JOIN users u ON c.user_id = u.id
WHERE u.role LIKE 'courier%'
ORDER BY c.level DESC;
EOF
)

# Execute the SQL
echo "$SQL_SCRIPT" | psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME"

echo "Done! Courier managed prefixes have been updated."