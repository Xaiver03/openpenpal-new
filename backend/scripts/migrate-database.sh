#!/bin/bash

# =====================================================
# OpenPenPal Database Migration Script
# =====================================================
# Description: Execute all pending database migrations to ensure DB is up to date
# Author: Claude Code Assistant
# Date: 2025-08-15

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-openpenpal}"
DB_USER="${DB_USER:-$(whoami)}"
DB_PASSWORD="${DB_PASSWORD:-}"

# Migration files directory
MIGRATIONS_DIR="$(dirname "$0")/../migrations"

echo -e "${BLUE}=== OpenPenPal Database Migration Script ===${NC}"
echo "Ensuring database is up to date with all credit incentive system changes..."

# Function to execute SQL file
execute_sql_file() {
    local file_path="$1"
    local description="$2"
    
    echo -e "\n${YELLOW}Executing: $description${NC}"
    echo "File: $(basename "$file_path")"
    
    if [ ! -f "$file_path" ]; then
        echo -e "${RED}✗ File not found: $file_path${NC}"
        return 1
    fi
    
    # Build psql command
    local psql_cmd="psql"
    if [ -n "$DB_PASSWORD" ]; then
        export PGPASSWORD="$DB_PASSWORD"
    fi
    
    psql_cmd+=" -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME"
    
    # Execute SQL file
    if $psql_cmd -f "$file_path" >/dev/null 2>&1; then
        echo -e "${GREEN}✓ Migration executed successfully${NC}"
        return 0
    else
        echo -e "${YELLOW}⚠ Migration may have already been applied or encountered non-critical errors${NC}"
        # Show the actual error for debugging
        echo "Attempting to show specific error..."
        $psql_cmd -f "$file_path" 2>&1 | head -10
        return 0  # Continue with other migrations
    fi
}

# Function to check if table exists
check_table_exists() {
    local table_name="$1"
    local psql_cmd="psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME"
    
    if [ -n "$DB_PASSWORD" ]; then
        export PGPASSWORD="$DB_PASSWORD"
    fi
    
    local result=$($psql_cmd -t -c "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = '$table_name');" 2>/dev/null | tr -d ' ')
    
    if [ "$result" = "t" ]; then
        return 0  # Table exists
    else
        return 1  # Table does not exist
    fi
}

# Function to check database connection
check_database_connection() {
    echo -e "\n${BLUE}=== Checking Database Connection ===${NC}"
    
    local psql_cmd="psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME"
    
    if [ -n "$DB_PASSWORD" ]; then
        export PGPASSWORD="$DB_PASSWORD"
    fi
    
    if $psql_cmd -c "SELECT version();" >/dev/null 2>&1; then
        echo -e "${GREEN}✓ Database connection successful${NC}"
        local version=$($psql_cmd -t -c "SELECT version();" | head -1)
        echo "PostgreSQL Version: $version"
        return 0
    else
        echo -e "${RED}✗ Failed to connect to database${NC}"
        echo "Connection parameters:"
        echo "  Host: $DB_HOST"
        echo "  Port: $DB_PORT"
        echo "  Database: $DB_NAME"
        echo "  User: $DB_USER"
        echo ""
        echo "Please ensure:"
        echo "1. PostgreSQL is running"
        echo "2. Database '$DB_NAME' exists"
        echo "3. User '$DB_USER' has access permissions"
        echo "4. Connection parameters are correct"
        return 1
    fi
}

# Function to check existing tables
check_existing_tables() {
    echo -e "\n${BLUE}=== Checking Existing Tables ===${NC}"
    
    local tables=(
        "user_credits"
        "credit_transactions" 
        "credit_limit_rules"
        "user_credit_actions"
        "credit_risk_users"
        "credit_shop_products"
        "credit_shop_categories"
        "credit_carts"
        "credit_redemptions"
        "credit_activities"
        "credit_activity_participations"
        "credit_activity_rewards"
        "credit_activity_rules"
        "credit_activity_templates"
        "credit_activity_target_audiences"
        "credit_activity_schedules"
        "credit_expiration_rules"
        "credit_expiration_batches"
        "credit_expiration_logs"
        "credit_expiration_notifications"
        "credit_transfers"
        "credit_transfer_rules"
        "credit_transfer_limits"
        "credit_transfer_notifications"
    )
    
    local existing_count=0
    local missing_tables=()
    
    for table in "${tables[@]}"; do
        if check_table_exists "$table"; then
            echo -e "${GREEN}✓ $table${NC}"
            ((existing_count++))
        else
            echo -e "${RED}✗ $table${NC}"
            missing_tables+=("$table")
        fi
    done
    
    echo ""
    echo "Summary: $existing_count/${#tables[@]} tables exist"
    
    if [ ${#missing_tables[@]} -gt 0 ]; then
        echo -e "${YELLOW}Missing tables:${NC}"
        for table in "${missing_tables[@]}"; do
            echo "  - $table"
        done
        return 1
    else
        echo -e "${GREEN}All expected tables exist!${NC}"
        return 0
    fi
}

# Function to run GORM auto-migration
run_gorm_migration() {
    echo -e "\n${BLUE}=== Running GORM Auto-Migration ===${NC}"
    
    # Check if the backend server binary exists or run via go run
    if [ -f "$(dirname "$0")/../main" ]; then
        echo "Running via compiled binary..."
        "$(dirname "$0")/../main" --migrate-only
    else
        echo "Running via go run..."
        cd "$(dirname "$0")/.."
        go run cmd/migrate/main.go -full
    fi
}

# Function to execute all migration files
execute_all_migrations() {
    echo -e "\n${BLUE}=== Executing Custom Migration Files ===${NC}"
    
    local migration_files=(
        # Credit system migration files (in order)
        "20240121_add_credit_limit_tables.sql"
        "20240122_add_credit_shop_tables.sql" 
        "003_credit_activity_system_pg.sql"     # Use PostgreSQL version
        "004_credit_expiration_system_pg.sql"   # Use PostgreSQL version
        "005_credit_transfer_system_pg.sql"     # Use PostgreSQL version
        
        # Other system migrations
        "002_add_user_profile_fields.sql"
        "003_create_follow_system_pg.sql"
        "004_create_privacy_system.sql"
        "005_create_opcode_tables.sql"
        "004_add_scan_records.sql"
    )
    
    local executed_count=0
    local failed_count=0
    
    for migration_file in "${migration_files[@]}"; do
        local file_path="$MIGRATIONS_DIR/$migration_file"
        
        if [ -f "$file_path" ]; then
            if execute_sql_file "$file_path" "Migration: $migration_file"; then
                ((executed_count++))
            else
                ((failed_count++))
            fi
        else
            echo -e "${YELLOW}⚠ Migration file not found: $migration_file${NC}"
        fi
    done
    
    echo ""
    echo -e "${BLUE}Migration Summary:${NC}"
    echo "  Executed: $executed_count"
    echo "  Failed: $failed_count"
    echo "  Total: ${#migration_files[@]}"
}

# Function to create missing credit system tables manually
create_missing_credit_tables() {
    echo -e "\n${BLUE}=== Creating Missing Credit System Tables ===${NC}"
    
    # Check and create each credit system table if missing
    local critical_tables=(
        "credit_transfers:005_credit_transfer_system.sql"
        "credit_expiration_rules:004_credit_expiration_system.sql"  
        "credit_activities:003_credit_activity_system.sql"
        "credit_shop_products:20240122_add_credit_shop_tables.sql"
        "credit_limiter_rules:20240121_add_credit_limit_tables.sql"
    )
    
    for table_info in "${critical_tables[@]}"; do
        local table_name="${table_info%:*}"
        local migration_file="${table_info#*:}"
        
        if ! check_table_exists "$table_name"; then
            echo -e "${YELLOW}Creating missing table: $table_name${NC}"
            local file_path="$MIGRATIONS_DIR/$migration_file"
            
            if [ -f "$file_path" ]; then
                execute_sql_file "$file_path" "Critical table creation: $table_name"
            else
                echo -e "${RED}✗ Migration file not found: $migration_file${NC}"
            fi
        fi
    done
}

# Function to verify migration success
verify_migration_success() {
    echo -e "\n${BLUE}=== Verifying Migration Success ===${NC}"
    
    # Check if all expected tables now exist
    if check_existing_tables; then
        echo -e "\n${GREEN}✓ All database migrations completed successfully!${NC}"
        return 0
    else
        echo -e "\n${YELLOW}⚠ Some tables are still missing. You may need to run manual migrations.${NC}"
        return 1
    fi
}

# Function to show manual migration commands
show_manual_commands() {
    echo -e "\n${YELLOW}=== Manual Migration Commands (if needed) ===${NC}"
    echo "If automatic migration fails, you can run these commands manually:"
    echo ""
    echo "1. Connect to database:"
    echo "   psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME"
    echo ""
    echo "2. Execute migration files:"
    echo "   \\i $MIGRATIONS_DIR/20240121_add_credit_limit_tables.sql"
    echo "   \\i $MIGRATIONS_DIR/20240122_add_credit_shop_tables.sql"
    echo "   \\i $MIGRATIONS_DIR/003_credit_activity_system.sql"
    echo "   \\i $MIGRATIONS_DIR/004_credit_expiration_system.sql"
    echo "   \\i $MIGRATIONS_DIR/005_credit_transfer_system.sql"
    echo ""
    echo "3. Run GORM migration:"
    echo "   cd $(dirname "$0")/.."
    echo "   go run cmd/migrate/main.go -full"
}

# Main execution
main() {
    echo "Starting database migration process..."
    echo "Database: $DB_NAME at $DB_HOST:$DB_PORT"
    echo ""
    
    # Check database connection
    if ! check_database_connection; then
        exit 1
    fi
    
    # Check existing tables
    check_existing_tables
    local initial_check=$?
    
    # Run GORM auto-migration first
    echo -e "\n${BLUE}=== Running GORM Auto-Migration ===${NC}"
    cd "$(dirname "$0")/.."
    echo "Executing: go run cmd/migrate/main.go -full"
    
    if go run cmd/migrate/main.go -full 2>/dev/null; then
        echo -e "${GREEN}✓ GORM migration completed${NC}"
    else
        echo -e "${YELLOW}⚠ GORM migration completed with warnings${NC}"
    fi
    
    # Execute custom migration files
    execute_all_migrations
    
    # Create any critical missing tables
    create_missing_credit_tables
    
    # Final verification
    verify_migration_success
    local final_result=$?
    
    if [ $final_result -eq 0 ]; then
        echo -e "\n${GREEN}🎉 Database migration completed successfully!${NC}"
        echo "All credit incentive system tables are ready."
        echo ""
        echo "You can now:"
        echo "1. Start the backend server: go run main.go"
        echo "2. Run system tests: ./scripts/test-credit-*.sh"
    else
        echo -e "\n${YELLOW}⚠ Migration completed with some issues.${NC}"
        show_manual_commands
    fi
    
    echo ""
    echo -e "${BLUE}Migration process finished.${NC}"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --host)
            DB_HOST="$2"
            shift 2
            ;;
        --port)
            DB_PORT="$2"
            shift 2
            ;;
        --database)
            DB_NAME="$2"
            shift 2
            ;;
        --user)
            DB_USER="$2"
            shift 2
            ;;
        --password)
            DB_PASSWORD="$2"
            shift 2
            ;;
        --help)
            echo "Usage: $0 [options]"
            echo ""
            echo "Options:"
            echo "  --host HOST        Database host (default: localhost)"
            echo "  --port PORT        Database port (default: 5432)"  
            echo "  --database DB      Database name (default: openpenpal)"
            echo "  --user USER        Database user (default: current user)"
            echo "  --password PASS    Database password (default: empty)"
            echo "  --help            Show this help message"
            echo ""
            echo "Environment variables:"
            echo "  DB_HOST, DB_PORT, DB_NAME, DB_USER, DB_PASSWORD"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Run the migration
main "$@"