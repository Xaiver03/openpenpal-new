#!/bin/bash

# OpenPenPal SQLite to PostgreSQL Migration Script
# ================================================

set -e  # Exit on any error

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
BACKUP_DIR="$PROJECT_ROOT/migration_backup/$(date +%Y%m%d_%H%M%S)"
MIGRATION_DIR="$PROJECT_ROOT/migration_data"

# Database configuration
SQLITE_DB="$PROJECT_ROOT/openpenpal.db"
BACKEND_SQLITE_DB="$PROJECT_ROOT/backend/openpenpal.db"
PG_HOST="localhost"
PG_PORT="5432"
PG_USER="openpenpal_user"
PG_DATABASE="openpenpal"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Logging function
log() {
    echo -e "${BLUE}[$(date '+%Y-%m-%d %H:%M:%S')]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Create directories
setup_directories() {
    log "Setting up migration directories..."
    mkdir -p "$BACKUP_DIR"
    mkdir -p "$MIGRATION_DIR"
    mkdir -p "$PROJECT_ROOT/logs"
    success "Migration directories created"
}

# Find the main SQLite database
find_sqlite_database() {
    log "Finding main SQLite database..."
    
    # Priority order: backend db, main db, largest db
    local candidates=(
        "$BACKEND_SQLITE_DB"
        "$SQLITE_DB"
        "$PROJECT_ROOT/openpenpal_sota.db"
        "$PROJECT_ROOT/opcode_test.db"
    )
    
    for db in "${candidates[@]}"; do
        if [[ -f "$db" && -s "$db" ]]; then
            # Check if database has tables
            local table_count=$(sqlite3 "$db" ".tables" 2>/dev/null | wc -w || echo "0")
            if [[ $table_count -gt 10 ]]; then
                SQLITE_DB="$db"
                log "Using SQLite database: $SQLITE_DB ($table_count tables)"
                return 0
            fi
        fi
    done
    
    error "No suitable SQLite database found!"
    return 1
}

# Backup current SQLite database
backup_sqlite() {
    log "Backing up SQLite databases..."
    
    # Copy all database files
    find "$PROJECT_ROOT" -name "*.db" -type f -exec cp {} "$BACKUP_DIR/" \;
    
    # Create SQL dump of main database
    if [[ -f "$SQLITE_DB" ]]; then
        sqlite3 "$SQLITE_DB" .dump > "$BACKUP_DIR/openpenpal_dump.sql"
        success "SQLite database backed up to: $BACKUP_DIR"
    else
        error "Main SQLite database not found: $SQLITE_DB"
        return 1
    fi
}

# Analyze SQLite schema
analyze_sqlite_schema() {
    log "Analyzing SQLite schema..."
    
    # Get table list
    sqlite3 "$SQLITE_DB" ".tables" > "$MIGRATION_DIR/tables.txt"
    
    # Get schema for each table
    sqlite3 "$SQLITE_DB" ".schema" > "$MIGRATION_DIR/schema.sql"
    
    # Get row counts
    echo "-- Row counts for each table" > "$MIGRATION_DIR/row_counts.sql"
    sqlite3 "$SQLITE_DB" ".tables" | tr ' ' '\n' | while read table; do
        if [[ -n "$table" ]]; then
            count=$(sqlite3 "$SQLITE_DB" "SELECT COUNT(*) FROM $table;" 2>/dev/null || echo "0")
            echo "-- $table: $count rows" >> "$MIGRATION_DIR/row_counts.sql"
        fi
    done
    
    local table_count=$(wc -w < "$MIGRATION_DIR/tables.txt")
    success "Schema analysis complete: $table_count tables found"
}

# Export SQLite data to PostgreSQL-compatible format
export_sqlite_data() {
    log "Exporting SQLite data to PostgreSQL format..."
    
    # Create conversion script
    cat > "$MIGRATION_DIR/convert_data.py" << 'EOF'
#!/usr/bin/env python3
import sqlite3
import psycopg2
import sys
import json
from datetime import datetime

def convert_sqlite_to_postgresql(sqlite_db, pg_config):
    """Convert SQLite database to PostgreSQL"""
    
    # Connect to SQLite
    sqlite_conn = sqlite3.connect(sqlite_db)
    sqlite_conn.row_factory = sqlite3.Row
    sqlite_cursor = sqlite_conn.cursor()
    
    # Connect to PostgreSQL
    pg_conn = psycopg2.connect(**pg_config)
    pg_cursor = pg_conn.cursor()
    
    # Get list of tables
    sqlite_cursor.execute("SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'")
    tables = [row[0] for row in sqlite_cursor.fetchall()]
    
    print(f"Found {len(tables)} tables to migrate")
    
    migrated_count = 0
    for table in tables:
        try:
            # Get table schema
            sqlite_cursor.execute(f"PRAGMA table_info({table})")
            columns = sqlite_cursor.fetchall()
            
            if not columns:
                print(f"Skipping empty table: {table}")
                continue
                
            # Get column names
            column_names = [col[1] for col in columns]
            
            # Check if table exists in PostgreSQL
            pg_cursor.execute("""
                SELECT EXISTS (
                    SELECT FROM information_schema.tables 
                    WHERE table_schema = 'public' 
                    AND table_name = %s
                )
            """, (table,))
            
            if not pg_cursor.fetchone()[0]:
                print(f"Table {table} does not exist in PostgreSQL, skipping...")
                continue
            
            # Get data from SQLite
            sqlite_cursor.execute(f"SELECT * FROM {table}")
            rows = sqlite_cursor.fetchall()
            
            if not rows:
                print(f"Table {table} is empty, skipping...")
                continue
            
            # Clear existing data in PostgreSQL table
            pg_cursor.execute(f"TRUNCATE TABLE {table} CASCADE")
            
            # Insert data into PostgreSQL
            placeholders = ','.join(['%s'] * len(column_names))
            insert_query = f"INSERT INTO {table} ({','.join(column_names)}) VALUES ({placeholders})"
            
            for row in rows:
                try:
                    # Convert row to list and handle special types
                    row_data = []
                    for i, value in enumerate(row):
                        if value is None:
                            row_data.append(None)
                        elif isinstance(value, str) and value.startswith('{') and value.endswith('}'):
                            # Might be JSON
                            try:
                                json.loads(value)
                                row_data.append(value)
                            except:
                                row_data.append(value)
                        else:
                            row_data.append(value)
                    
                    pg_cursor.execute(insert_query, row_data)
                except Exception as e:
                    print(f"Error inserting row into {table}: {e}")
                    continue
            
            pg_conn.commit()
            migrated_count += 1
            print(f"✓ Migrated table: {table} ({len(rows)} rows)")
            
        except Exception as e:
            print(f"✗ Error migrating table {table}: {e}")
            pg_conn.rollback()
            continue
    
    # Close connections
    sqlite_conn.close()
    pg_conn.close()
    
    print(f"\nMigration complete: {migrated_count}/{len(tables)} tables migrated")
    return migrated_count

if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("Usage: python3 convert_data.py <sqlite_db_path>")
        sys.exit(1)
    
    sqlite_db = sys.argv[1]
    pg_config = {
        'host': 'localhost',
        'port': 5432,
        'user': 'openpenpal_user',
        'password': '',
        'database': 'openpenpal'
    }
    
    try:
        convert_sqlite_to_postgresql(sqlite_db, pg_config)
    except Exception as e:
        print(f"Migration failed: {e}")
        sys.exit(1)
EOF

    chmod +x "$MIGRATION_DIR/convert_data.py"
    success "Data conversion script created"
}

# Check PostgreSQL connection
check_postgresql() {
    log "Checking PostgreSQL connection..."
    
    if pg_isready -h "$PG_HOST" -p "$PG_PORT" -U "$PG_USER" -d "$PG_DATABASE" >/dev/null 2>&1; then
        success "PostgreSQL connection verified"
        
        # Check if tables exist
        local table_count=$(psql -h "$PG_HOST" -p "$PG_PORT" -U "$PG_USER" -d "$PG_DATABASE" -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';" 2>/dev/null | xargs)
        log "PostgreSQL has $table_count tables"
        
        if [[ $table_count -gt 100 ]]; then
            success "PostgreSQL database appears to be properly initialized"
            return 0
        else
            warning "PostgreSQL database may need schema initialization"
            return 1
        fi
    else
        error "Cannot connect to PostgreSQL database"
        return 1
    fi
}

# Create schema diff report
create_schema_diff() {
    log "Creating schema difference report..."
    
    # SQLite tables
    sqlite3 "$SQLITE_DB" ".tables" | tr ' ' '\n' | sort > "$MIGRATION_DIR/sqlite_tables.txt"
    
    # PostgreSQL tables
    psql -h "$PG_HOST" -p "$PG_PORT" -U "$PG_USER" -d "$PG_DATABASE" -t -c "SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' ORDER BY table_name;" | grep -v '^$' | xargs -n1 echo > "$MIGRATION_DIR/postgresql_tables.txt"
    
    # Create diff report
    {
        echo "# Schema Difference Report"
        echo "Generated: $(date)"
        echo ""
        echo "## SQLite Tables ($(wc -l < "$MIGRATION_DIR/sqlite_tables.txt"))"
        cat "$MIGRATION_DIR/sqlite_tables.txt"
        echo ""
        echo "## PostgreSQL Tables ($(wc -l < "$MIGRATION_DIR/postgresql_tables.txt"))"
        cat "$MIGRATION_DIR/postgresql_tables.txt"
        echo ""
        echo "## Tables only in SQLite:"
        comm -23 "$MIGRATION_DIR/sqlite_tables.txt" "$MIGRATION_DIR/postgresql_tables.txt"
        echo ""
        echo "## Tables only in PostgreSQL:"
        comm -13 "$MIGRATION_DIR/sqlite_tables.txt" "$MIGRATION_DIR/postgresql_tables.txt"
    } > "$MIGRATION_DIR/schema_diff_report.md"
    
    success "Schema difference report created: $MIGRATION_DIR/schema_diff_report.md"
}

# Main migration function
main() {
    log "🚀 Starting OpenPenPal SQLite to PostgreSQL Migration"
    log "======================================================="
    
    # Setup
    setup_directories
    
    # Find and analyze SQLite database
    if ! find_sqlite_database; then
        exit 1
    fi
    
    # Backup current data
    if ! backup_sqlite; then
        exit 1
    fi
    
    # Analyze schema
    analyze_sqlite_schema
    
    # Check PostgreSQL
    if ! check_postgresql; then
        warning "PostgreSQL may need to be initialized first"
        log "Please ensure PostgreSQL is running and has the correct schema"
    fi
    
    # Create comparison reports
    create_schema_diff
    
    # Create migration script
    export_sqlite_data
    
    log ""
    log "🎯 Migration preparation complete!"
    log "============================================"
    log "Next steps:"
    log "1. Review schema diff: $MIGRATION_DIR/schema_diff_report.md"
    log "2. Ensure PostgreSQL has all required tables"
    log "3. Run data migration: python3 $MIGRATION_DIR/convert_data.py $SQLITE_DB"
    log "4. Verify data integrity"
    log ""
    log "Backup location: $BACKUP_DIR"
    log "Migration files: $MIGRATION_DIR"
}

# Run main function
main "$@"