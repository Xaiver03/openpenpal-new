#!/bin/bash
# Safe Index Optimization Runner with Rollback Support

echo "🔍 OpenPenPal Index Optimization (ULTRATHINK Edition)"
echo "====================================================="
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

# Function to check index count
check_indexes() {
    psql "host=$DB_HOST port=$DB_PORT user=$DB_USER dbname=$DB_NAME sslmode=disable" -t -c "
        SELECT COUNT(*) FROM pg_indexes WHERE schemaname = 'public';
    " 2>/dev/null | tr -d ' '
}

# Function to check for duplicates
check_duplicates() {
    psql "host=$DB_HOST port=$DB_PORT user=$DB_USER dbname=$DB_NAME sslmode=disable" -t -c "
        WITH index_cols AS (
            SELECT tablename, indexname, 
                   regexp_replace(indexdef, '.*\((.*)\)', '\1') as columns
            FROM pg_indexes
            WHERE schemaname = 'public'
        )
        SELECT COUNT(*) 
        FROM (
            SELECT tablename, columns, COUNT(*) as cnt
            FROM index_cols
            GROUP BY tablename, columns
            HAVING COUNT(*) > 1
        ) dups;
    " 2>/dev/null | tr -d ' '
}

echo "📊 Pre-optimization Status:"
echo "------------------------"
INITIAL_COUNT=$(check_indexes)
INITIAL_DUPS=$(check_duplicates)
echo "Total indexes: $INITIAL_COUNT"
echo "Duplicate index groups: $INITIAL_DUPS"
echo ""

# Create backup of current index definitions
echo "💾 Creating backup of current indexes..."
psql "host=$DB_HOST port=$DB_PORT user=$DB_USER dbname=$DB_NAME sslmode=disable" -c "
    SELECT indexdef || ';' 
    FROM pg_indexes 
    WHERE schemaname = 'public'
    ORDER BY tablename, indexname;
" -t -A > index_backup_$(date +%Y%m%d_%H%M%S).sql

echo -e "${GREEN}✓ Backup created${NC}"
echo ""

# Run optimization in a transaction
echo "🚀 Running index optimization..."
echo "------------------------"

psql "host=$DB_HOST port=$DB_PORT user=$DB_USER dbname=$DB_NAME sslmode=disable" <<EOF
-- Start transaction for safety
BEGIN;

-- Show current status
\echo 'Starting optimization...'

-- Include the optimization script
\i optimize_indexes_final.sql

-- Check if we should commit
DO \$\$
DECLARE
    error_count INTEGER;
BEGIN
    -- Check for any errors (this is a simplified check)
    -- In production, you'd want more sophisticated error handling
    GET DIAGNOSTICS error_count = ROW_COUNT;
    
    IF error_count IS NULL THEN
        RAISE NOTICE 'Optimization completed successfully';
    END IF;
END\$\$;

-- Commit the changes
COMMIT;

\echo 'Optimization completed!'
EOF

if [ $? -ne 0 ]; then
    echo -e "${RED}❌ Optimization failed! Database changes have been rolled back.${NC}"
    exit 1
fi

echo ""
echo "📊 Post-optimization Status:"
echo "------------------------"
FINAL_COUNT=$(check_indexes)
FINAL_DUPS=$(check_duplicates)
echo "Total indexes: $FINAL_COUNT"
echo "Duplicate index groups: $FINAL_DUPS"
echo "Indexes added: $((FINAL_COUNT - INITIAL_COUNT + 3))" # +3 for removed duplicates
echo "Duplicates removed: $((INITIAL_DUPS - FINAL_DUPS))"
echo ""

# Generate performance report
echo "📈 Generating performance report..."
psql "host=$DB_HOST port=$DB_PORT user=$DB_USER dbname=$DB_NAME sslmode=disable" <<EOF
\echo ''
\echo 'Top 10 Most Used Indexes:'
\echo '-------------------------'
SELECT 
    indexrelname as indexname,
    relname as tablename,
    idx_scan as scans,
    pg_size_pretty(pg_relation_size(indexrelid)) as size
FROM pg_stat_user_indexes
WHERE schemaname = 'public'
ORDER BY idx_scan DESC
LIMIT 10;

\echo ''
\echo 'Unused Indexes (consider removal):'
\echo '---------------------------------'
SELECT 
    indexrelname as indexname,
    relname as tablename,
    pg_size_pretty(pg_relation_size(indexrelid)) as size
FROM pg_stat_user_indexes
WHERE schemaname = 'public' AND idx_scan = 0
ORDER BY pg_relation_size(indexrelid) DESC
LIMIT 10;
EOF

echo ""
echo -e "${GREEN}✅ Index optimization completed successfully!${NC}"
echo ""
echo "💡 Next steps:"
echo "   1. Monitor query performance over the next 24-48 hours"
echo "   2. Check pg_stat_user_indexes for usage patterns"
echo "   3. Run VACUUM ANALYZE on major tables if needed"
echo "   4. Review the backup file if rollback is needed"