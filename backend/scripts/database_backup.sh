#!/bin/bash

# OpenPenPal Database Backup Strategy
# Comprehensive backup solution with rotation and monitoring
# =========================================================

# Configuration
DB_HOST="localhost"
DB_PORT="5432"
DB_NAME="openpenpal"
DB_USER="openpenpal_user"

# Backup configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKUP_DIR="$SCRIPT_DIR/../backups/openpenpal"
BACKUP_RETENTION_DAYS=30
DAILY_RETENTION=7
WEEKLY_RETENTION=4
MONTHLY_RETENTION=12

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Logging
LOG_FILE="$BACKUP_DIR/backup.log"

# Function to log messages
log_message() {
    local level=$1
    local message=$2
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo "[$timestamp] [$level] $message" | tee -a "$LOG_FILE"
}

# Function to create backup directories
setup_directories() {
    log_message "INFO" "Setting up backup directories..."
    
    mkdir -p "$BACKUP_DIR"/{daily,weekly,monthly,logs}
    
    if [ ! -w "$BACKUP_DIR" ]; then
        log_message "ERROR" "Cannot write to backup directory: $BACKUP_DIR"
        exit 1
    fi
    
    log_message "INFO" "Backup directories created successfully"
}

# Function to check database connection
check_database() {
    log_message "INFO" "Checking database connection..."
    
    if ! pg_isready -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" >/dev/null 2>&1; then
        log_message "ERROR" "Cannot connect to database"
        exit 1
    fi
    
    log_message "INFO" "Database connection verified"
}

# Function to get database size
get_database_size() {
    psql "host=$DB_HOST port=$DB_PORT user=$DB_USER dbname=$DB_NAME sslmode=disable" \
        -t -c "SELECT pg_size_pretty(pg_database_size('$DB_NAME'));" 2>/dev/null | xargs
}

# Function to create backup
create_backup() {
    local backup_type=$1
    local timestamp=$(date '+%Y%m%d_%H%M%S')
    local backup_file="$BACKUP_DIR/$backup_type/openpenpal_${backup_type}_${timestamp}.sql"
    local compressed_file="${backup_file}.gz"
    
    log_message "INFO" "Starting $backup_type backup..."
    log_message "INFO" "Database size: $(get_database_size)"
    
    # Create the backup
    if pg_dump \
        --host="$DB_HOST" \
        --port="$DB_PORT" \
        --username="$DB_USER" \
        --dbname="$DB_NAME" \
        --no-password \
        --format=custom \
        --verbose \
        --file="$backup_file" 2>>"$LOG_FILE"; then
        
        # Compress the backup
        if gzip "$backup_file"; then
            local backup_size=$(ls -lh "$compressed_file" | awk '{print $5}')
            log_message "INFO" "$backup_type backup completed successfully"
            log_message "INFO" "Backup file: $compressed_file ($backup_size)"
            
            # Verify backup integrity
            if gunzip -t "$compressed_file" 2>/dev/null; then
                log_message "INFO" "Backup integrity verified"
                echo "$compressed_file"
            else
                log_message "ERROR" "Backup integrity check failed"
                rm -f "$compressed_file"
                exit 1
            fi
        else
            log_message "ERROR" "Failed to compress backup"
            rm -f "$backup_file"
            exit 1
        fi
    else
        log_message "ERROR" "$backup_type backup failed"
        rm -f "$backup_file"
        exit 1
    fi
}

# Function to cleanup old backups
cleanup_backups() {
    local backup_type=$1
    local retention_count=$2
    local backup_dir="$BACKUP_DIR/$backup_type"
    
    log_message "INFO" "Cleaning up old $backup_type backups (keeping $retention_count)"
    
    # Remove old backups, keeping only the specified number
    ls -t "$backup_dir"/openpenpal_${backup_type}_*.sql.gz 2>/dev/null | \
    tail -n +$((retention_count + 1)) | \
    while read -r old_backup; do
        if [ -f "$old_backup" ]; then
            log_message "INFO" "Removing old backup: $(basename "$old_backup")"
            rm -f "$old_backup"
        fi
    done
}

# Function to send backup notification
send_notification() {
    local status=$1
    local backup_file=$2
    local message="OpenPenPal Database Backup: $status"
    
    if [ "$status" = "SUCCESS" ]; then
        message="$message\nBackup file: $(basename "$backup_file")\nSize: $(ls -lh "$backup_file" 2>/dev/null | awk '{print $5}')\nTime: $(date)"
    else
        message="$message\nError details in log: $LOG_FILE\nTime: $(date)"
    fi
    
    # You can implement email notification here
    log_message "INFO" "Notification: $message"
}

# Function to create daily backup
daily_backup() {
    log_message "INFO" "=== Starting Daily Backup ==="
    
    local backup_file=$(create_backup "daily")
    if [ $? -eq 0 ]; then
        cleanup_backups "daily" "$DAILY_RETENTION"
        send_notification "SUCCESS" "$backup_file"
        log_message "INFO" "=== Daily Backup Completed ==="
    else
        send_notification "FAILED" ""
        log_message "ERROR" "=== Daily Backup Failed ==="
        exit 1
    fi
}

# Function to create weekly backup
weekly_backup() {
    log_message "INFO" "=== Starting Weekly Backup ==="
    
    local backup_file=$(create_backup "weekly")
    if [ $? -eq 0 ]; then
        cleanup_backups "weekly" "$WEEKLY_RETENTION"
        send_notification "SUCCESS" "$backup_file"
        log_message "INFO" "=== Weekly Backup Completed ==="
    else
        send_notification "FAILED" ""
        log_message "ERROR" "=== Weekly Backup Failed ==="
        exit 1
    fi
}

# Function to create monthly backup
monthly_backup() {
    log_message "INFO" "=== Starting Monthly Backup ==="
    
    local backup_file=$(create_backup "monthly")
    if [ $? -eq 0 ]; then
        cleanup_backups "monthly" "$MONTHLY_RETENTION"
        send_notification "SUCCESS" "$backup_file"
        log_message "INFO" "=== Monthly Backup Completed ==="
    else
        send_notification "FAILED" ""
        log_message "ERROR" "=== Monthly Backup Failed ==="
        exit 1
    fi
}

# Function to restore from backup
restore_backup() {
    local backup_file=$1
    
    if [ ! -f "$backup_file" ]; then
        log_message "ERROR" "Backup file not found: $backup_file"
        exit 1
    fi
    
    log_message "WARNING" "Starting database restore from: $(basename "$backup_file")"
    read -p "This will overwrite the current database. Are you sure? (yes/no): " confirm
    
    if [ "$confirm" != "yes" ]; then
        log_message "INFO" "Restore cancelled by user"
        exit 0
    fi
    
    # Decompress if needed
    if [[ "$backup_file" == *.gz ]]; then
        local temp_file="${backup_file%.gz}"
        gunzip -c "$backup_file" > "$temp_file"
        backup_file="$temp_file"
    fi
    
    # Drop and recreate database
    dropdb -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" "$DB_NAME" 2>/dev/null
    createdb -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" "$DB_NAME"
    
    # Restore from backup
    if pg_restore \
        --host="$DB_HOST" \
        --port="$DB_PORT" \
        --username="$DB_USER" \
        --dbname="$DB_NAME" \
        --no-password \
        --verbose \
        "$backup_file" 2>>"$LOG_FILE"; then
        
        log_message "INFO" "Database restore completed successfully"
        
        # Clean up temp file
        if [[ "$1" == *.gz ]]; then
            rm -f "$backup_file"
        fi
    else
        log_message "ERROR" "Database restore failed"
        exit 1
    fi
}

# Function to list backups
list_backups() {
    echo -e "${BLUE}Available Backups:${NC}"
    echo "=================="
    
    for backup_type in daily weekly monthly; do
        echo -e "\n${YELLOW}$backup_type backups:${NC}"
        ls -lh "$BACKUP_DIR/$backup_type"/openpenpal_${backup_type}_*.sql.gz 2>/dev/null | \
        awk '{print $9, "(" $5 ")", $6, $7, $8}' | \
        while read -r line; do
            if [ ! -z "$line" ]; then
                echo "  $line"
            fi
        done
    done
}

# Function to show backup status
backup_status() {
    echo -e "${BLUE}Backup Status Report:${NC}"
    echo "===================="
    echo "Database: $DB_NAME @ $DB_HOST:$DB_PORT"
    echo "Database Size: $(get_database_size)"
    echo "Backup Directory: $BACKUP_DIR"
    echo ""
    
    for backup_type in daily weekly monthly; do
        local count=$(ls "$BACKUP_DIR/$backup_type"/openpenpal_${backup_type}_*.sql.gz 2>/dev/null | wc -l)
        local latest=""
        if [ $count -gt 0 ]; then
            latest=$(ls -t "$BACKUP_DIR/$backup_type"/openpenpal_${backup_type}_*.sql.gz 2>/dev/null | head -n1)
            latest=" (Latest: $(basename "$latest"))"
        fi
        echo "$backup_type: $count backups$latest"
    done
    
    echo ""
    echo "Disk Usage:"
    du -sh "$BACKUP_DIR"/* 2>/dev/null | while read -r size dir; do
        echo "  $(basename "$dir"): $size"
    done
}

# Function to show usage
show_usage() {
    echo "OpenPenPal Database Backup Tool"
    echo "==============================="
    echo ""
    echo "Usage: $0 [command] [options]"
    echo ""
    echo "Commands:"
    echo "  daily              Create daily backup"
    echo "  weekly             Create weekly backup"
    echo "  monthly            Create monthly backup"
    echo "  restore <file>     Restore from backup file"
    echo "  list               List all available backups"
    echo "  status             Show backup status"
    echo "  setup              Setup backup directories and cron jobs"
    echo ""
    echo "Examples:"
    echo "  $0 daily                                    # Create daily backup"
    echo "  $0 restore /path/to/backup.sql.gz         # Restore from backup"
    echo "  $0 list                                    # List all backups"
    echo ""
}

# Function to setup cron jobs
setup_cron() {
    log_message "INFO" "Setting up automated backup schedule..."
    
    local script_path=$(realpath "$0")
    local cron_daily="0 2 * * * $script_path daily >> $LOG_FILE 2>&1"
    local cron_weekly="0 3 * * 0 $script_path weekly >> $LOG_FILE 2>&1"
    local cron_monthly="0 4 1 * * $script_path monthly >> $LOG_FILE 2>&1"
    
    # Add cron jobs (you may need to run this as root or with appropriate permissions)
    (crontab -l 2>/dev/null | grep -v "$script_path"; echo "$cron_daily"; echo "$cron_weekly"; echo "$cron_monthly") | crontab -
    
    log_message "INFO" "Cron jobs setup completed:"
    log_message "INFO" "  Daily: 2:00 AM every day"
    log_message "INFO" "  Weekly: 3:00 AM every Sunday"
    log_message "INFO" "  Monthly: 4:00 AM on 1st of every month"
}

# Main execution
main() {
    case "${1:-}" in
        daily)
            setup_directories
            check_database
            daily_backup
            ;;
        weekly)
            setup_directories
            check_database
            weekly_backup
            ;;
        monthly)
            setup_directories
            check_database
            monthly_backup
            ;;
        restore)
            if [ -z "$2" ]; then
                echo "Error: Please specify backup file to restore"
                show_usage
                exit 1
            fi
            check_database
            restore_backup "$2"
            ;;
        list)
            list_backups
            ;;
        status)
            backup_status
            ;;
        setup)
            setup_directories
            setup_cron
            log_message "INFO" "Backup system setup completed"
            ;;
        *)
            show_usage
            ;;
    esac
}

# Run main function
main "$@"