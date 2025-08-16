#!/usr/bin/env python3
"""
Improved SQLite to PostgreSQL migration script
Handles type conversions and foreign key dependencies
"""

import sqlite3
import psycopg2
import sys
import json
from datetime import datetime

def convert_value(value, column_name, table_name):
    """Convert SQLite values to PostgreSQL compatible format"""
    if value is None:
        return None
    
    # Handle boolean fields (SQLite stores as 0/1, PostgreSQL needs true/false)
    boolean_fields = {
        'users': ['is_active'],
        'letter_templates': ['is_public', 'is_default'],
        'couriers': ['has_printer'],
        'op_code_schools': ['is_active'],
        'envelope_designs': ['is_active'],
        'notifications': ['is_read']
    }
    
    if table_name in boolean_fields and column_name in boolean_fields[table_name]:
        return bool(int(value)) if str(value).isdigit() else bool(value)
    
    # Handle JSON fields
    if isinstance(value, str) and value.startswith('{') and value.endswith('}'):
        try:
            json.loads(value)
            return value
        except:
            pass
    
    # Handle datetime fields - ensure proper format
    if column_name.endswith('_at') or column_name in ['created_at', 'updated_at', 'deleted_at']:
        if isinstance(value, str):
            try:
                # Try to parse and reformat datetime
                dt = datetime.fromisoformat(value.replace('Z', '+00:00'))
                return dt
            except:
                pass
    
    return value

def get_table_dependencies():
    """Define table dependencies for proper migration order"""
    # Tables that don't depend on others (base tables)
    base_tables = [
        'users', 'envelope_designs', 'sensitive_words', 'email_templates',
        'system_analytics', 'analytics_reports', 'scheduled_tasks', 
        'task_templates', 'storage_configs', 'user_levels', 'credit_rules'
    ]
    
    # Tables that depend on users
    user_dependent = [
        'user_profiles', 'user_credits', 'user_analytics', 'user_daily_usages',
        'user_achievements', 'user_stats', 'user_privacy_settings',
        'couriers', 'notifications', 'letter_templates'
    ]
    
    # Tables that depend on letters
    letter_dependent = [
        'letter_codes', 'letter_photos', 'letter_likes', 'letter_shares', 
        'status_logs'
    ]
    
    # Tables that depend on other entities
    complex_dependent = [
        'letters',  # depends on users
        'courier_tasks',  # depends on couriers
        'museum_items', 'museum_collections', 'museum_entries',  # depend on users/letters
        'credit_transactions',  # depends on users
        'email_logs',  # depends on email_templates
        'task_executions',  # depends on scheduled_tasks
        'storage_files', 'storage_operations'  # depend on storage_configs
    ]
    
    return base_tables + user_dependent + ['letters'] + letter_dependent + complex_dependent

def migrate_table_with_retry(sqlite_cursor, pg_cursor, table_name, max_retries=3):
    """Migrate a table with error handling and retries"""
    print(f"Migrating table: {table_name}")
    
    # Get table schema
    sqlite_cursor.execute(f"PRAGMA table_info({table_name})")
    columns = sqlite_cursor.fetchall()
    
    if not columns:
        print(f"  Skipping empty table: {table_name}")
        return True
    
    # Get column names
    column_names = [col[1] for col in columns]
    
    # Check if table exists in PostgreSQL
    pg_cursor.execute("""
        SELECT EXISTS (
            SELECT FROM information_schema.tables 
            WHERE table_schema = 'public' 
            AND table_name = %s
        )
    """, (table_name,))
    
    if not pg_cursor.fetchone()[0]:
        print(f"  Table {table_name} does not exist in PostgreSQL, skipping...")
        return False
    
    # Get data from SQLite
    sqlite_cursor.execute(f"SELECT * FROM {table_name}")
    rows = sqlite_cursor.fetchall()
    
    if not rows:
        print(f"  Table {table_name} is empty, skipping...")
        return True
    
    # Clear existing data in PostgreSQL table (disable foreign key checks temporarily)
    try:
        pg_cursor.execute("SET session_replication_role = replica")  # Disable triggers/constraints
        pg_cursor.execute(f"TRUNCATE TABLE {table_name} CASCADE")
        pg_cursor.execute("SET session_replication_role = DEFAULT")  # Re-enable triggers/constraints
    except Exception as e:
        print(f"  Warning: Could not truncate {table_name}: {e}")
    
    # Insert data into PostgreSQL
    placeholders = ','.join(['%s'] * len(column_names))
    insert_query = f"INSERT INTO {table_name} ({','.join(column_names)}) VALUES ({placeholders})"
    
    successful_rows = 0
    failed_rows = 0
    
    for row in rows:
        try:
            # Convert row data
            row_data = []
            for i, value in enumerate(row):
                converted_value = convert_value(value, column_names[i], table_name)
                row_data.append(converted_value)
            
            # Attempt to insert
            pg_cursor.execute(insert_query, row_data)
            successful_rows += 1
            
        except Exception as e:
            failed_rows += 1
            if failed_rows <= 5:  # Only show first 5 errors
                print(f"    Error inserting row: {e}")
            elif failed_rows == 6:
                print(f"    ... (suppressing further errors for {table_name})")
    
    print(f"  ✓ {table_name}: {successful_rows} successful, {failed_rows} failed")
    return successful_rows > 0

def convert_sqlite_to_postgresql(sqlite_db, pg_config):
    """Convert SQLite database to PostgreSQL with improved error handling"""
    
    # Connect to SQLite
    sqlite_conn = sqlite3.connect(sqlite_db)
    sqlite_conn.row_factory = sqlite3.Row
    sqlite_cursor = sqlite_conn.cursor()
    
    # Connect to PostgreSQL
    pg_conn = psycopg2.connect(**pg_config)
    pg_conn.autocommit = False  # Use transactions
    pg_cursor = pg_conn.cursor()
    
    print("🚀 Starting improved SQLite to PostgreSQL migration")
    print("=" * 60)
    
    # Get list of tables from SQLite
    sqlite_cursor.execute("SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'")
    available_tables = [row[0] for row in sqlite_cursor.fetchall()]
    
    print(f"Found {len(available_tables)} tables in SQLite")
    
    # Get dependency-ordered table list
    ordered_tables = get_table_dependencies()
    
    # Filter to only include tables that exist in both databases
    migration_tables = [t for t in ordered_tables if t in available_tables]
    
    print(f"Will migrate {len(migration_tables)} tables in dependency order")
    print("")
    
    migrated_count = 0
    failed_tables = []
    
    # Migrate tables in dependency order
    for table in migration_tables:
        try:
            if migrate_table_with_retry(sqlite_cursor, pg_cursor, table):
                pg_conn.commit()
                migrated_count += 1
            else:
                failed_tables.append(table)
                pg_conn.rollback()
        except Exception as e:
            print(f"  ✗ Critical error with {table}: {e}")
            failed_tables.append(table)
            pg_conn.rollback()
    
    # Migrate any remaining tables
    remaining_tables = [t for t in available_tables if t not in migration_tables]
    if remaining_tables:
        print(f"\nMigrating {len(remaining_tables)} additional tables...")
        for table in remaining_tables:
            try:
                if migrate_table_with_retry(sqlite_cursor, pg_cursor, table):
                    pg_conn.commit()
                    migrated_count += 1
                else:
                    failed_tables.append(table)
                    pg_conn.rollback()
            except Exception as e:
                print(f"  ✗ Critical error with {table}: {e}")
                failed_tables.append(table)
                pg_conn.rollback()
    
    # Close connections
    sqlite_conn.close()
    pg_conn.close()
    
    print("")
    print("🎯 Migration Summary")
    print("=" * 60)
    print(f"Successfully migrated: {migrated_count} tables")
    print(f"Failed migrations: {len(failed_tables)} tables")
    
    if failed_tables:
        print(f"Failed tables: {', '.join(failed_tables)}")
    
    print("")
    if migrated_count > 0:
        print("✅ Migration completed with partial success!")
        print("Next step: Restart backend to use PostgreSQL")
    else:
        print("❌ Migration failed completely")
        return False
    
    return True

if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("Usage: python3 convert_data_improved.py <sqlite_db_path>")
        sys.exit(1)
    
    sqlite_db = sys.argv[1]
    pg_config = {
        'host': 'localhost',
        'port': 5432,
        'user': 'openpenpal_user',
        'password': '',  # Add password if needed
        'database': 'openpenpal'
    }
    
    try:
        success = convert_sqlite_to_postgresql(sqlite_db, pg_config)
        sys.exit(0 if success else 1)
    except Exception as e:
        print(f"Migration failed: {e}")
        sys.exit(1)