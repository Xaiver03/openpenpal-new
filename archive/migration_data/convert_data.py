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
