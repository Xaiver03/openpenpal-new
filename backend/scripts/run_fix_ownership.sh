#!/bin/bash
# Script to fix database ownership issues

echo "🔧 Fixing database ownership issues..."
echo "This script will change ownership of all database objects to 'openpenpal_user'"
echo ""

# Database connection parameters
DB_HOST="localhost"
DB_PORT="5432"
DB_NAME="openpenpal"
DB_SUPERUSER="rocalight"  # Use the superuser who owns the tables

# Run the fix ownership script
echo "Running ownership fix script..."
psql "host=$DB_HOST port=$DB_PORT user=$DB_SUPERUSER dbname=$DB_NAME sslmode=disable" -f fix_ownership.sql

# Verify the changes
echo ""
echo "Verifying ownership changes..."
psql "host=$DB_HOST port=$DB_PORT user=openpenpal_user dbname=$DB_NAME sslmode=disable" -c "
SELECT 
    'Tables: ' || tableowner as ownership,
    COUNT(*) as count
FROM pg_tables 
WHERE schemaname = 'public'
GROUP BY tableowner;"

echo ""
echo "✅ Ownership fix completed!"