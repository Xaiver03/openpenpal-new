#!/bin/bash

# OpenPenPal 日志监控和告警系统
# 防止日志爆炸，实时监控日志文件大小和增长速度

set -euo pipefail

# 配置
PROJECT_ROOT="/Users/rocalight/同步空间/opplc/openpenpal"
BACKEND_LOG_DIR="$PROJECT_ROOT/backend"
FRONTEND_LOG_DIR="$PROJECT_ROOT/frontend"
GENERAL_LOG_DIR="$PROJECT_ROOT/logs"
ALERT_LOG="$PROJECT_ROOT/logs/log-monitor-alerts.log"
STATE_FILE="$PROJECT_ROOT/tmp/log-monitor.state"

# 阈值配置
MAX_FILE_SIZE_MB=500      # 单个文件最大500MB
MAX_GROWTH_RATE_MB=50     # 每分钟最大增长50MB
MAX_TOTAL_SIZE_GB=5       # 总日志大小最大5GB
CRITICAL_SIZE_GB=10       # 关键告警阈值10GB

# 颜色输出
RED='\033[0;31m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 确保目录存在
mkdir -p "$(dirname "$ALERT_LOG")"
mkdir -p "$(dirname "$STATE_FILE")"

# 日志函数
log_message() {
    local level="$1"
    local message="$2"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo "[$timestamp] [$level] $message" | tee -a "$ALERT_LOG"
}

log_info() {
    log_message "INFO" "$1"
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_warn() {
    log_message "WARN" "$1"
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    log_message "ERROR" "$1"
    echo -e "${RED}[ERROR]${NC} $1"
}

log_success() {
    log_message "SUCCESS" "$1"
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

# 获取文件大小（MB）
get_file_size_mb() {
    local file="$1"
    if [[ -f "$file" ]]; then
        local size_bytes=$(stat -f%z "$file" 2>/dev/null || stat -c%s "$file" 2>/dev/null || echo "0")
        echo $((size_bytes / 1024 / 1024))
    else
        echo "0"
    fi
}

# 获取目录总大小（GB）
get_dir_size_gb() {
    local dir="$1"
    if [[ -d "$dir" ]]; then
        local size_bytes=$(du -sb "$dir" 2>/dev/null | cut -f1 || echo "0")
        echo "scale=2; $size_bytes / 1024 / 1024 / 1024" | bc -l
    else
        echo "0"
    fi
}

# 读取上次状态
read_previous_state() {
    if [[ -f "$STATE_FILE" ]]; then
        source "$STATE_FILE"
    fi
}

# 保存当前状态
save_current_state() {
    cat > "$STATE_FILE" << EOF
# Log Monitor State - $(date)
LAST_CHECK_TIME=$(date +%s)
PREVIOUS_BACKEND_SIZE=$CURRENT_BACKEND_SIZE
PREVIOUS_FRONTEND_SIZE=$CURRENT_FRONTEND_SIZE
PREVIOUS_TOTAL_SIZE=$CURRENT_TOTAL_SIZE
EOF
}

# 检查单个文件
check_file() {
    local file="$1"
    local size_mb=$(get_file_size_mb "$file")
    
    if [[ $size_mb -gt $MAX_FILE_SIZE_MB ]]; then
        log_error "🚨 CRITICAL: File size exceeded - $file ($size_mb MB > $MAX_FILE_SIZE_MB MB)"
        
        # 自动处理：截断文件
        if [[ -w "$file" ]]; then
            log_warn "Auto-truncating file: $file"
            echo "# File truncated by log-monitor at $(date)" > "$file"
            log_success "File truncated successfully: $file"
        fi
        
        return 1
    elif [[ $size_mb -gt $((MAX_FILE_SIZE_MB * 80 / 100)) ]]; then
        log_warn "⚠️  File size warning - $file ($size_mb MB)"
    fi
    
    return 0
}

# 检查增长速度
check_growth_rate() {
    local current_size="$1"
    local previous_size="$2"
    local time_diff="$3"
    local location="$4"
    
    if [[ -n "$previous_size" && -n "$time_diff" && $time_diff -gt 0 ]]; then
        local growth=$((current_size - previous_size))
        local growth_rate=$((growth * 60 / time_diff))  # MB per minute
        
        if [[ $growth_rate -gt $MAX_GROWTH_RATE_MB ]]; then
            log_error "🚨 CRITICAL: High log growth rate in $location: $growth_rate MB/min (threshold: $MAX_GROWTH_RATE_MB MB/min)"
            return 1
        elif [[ $growth_rate -gt $((MAX_GROWTH_RATE_MB * 70 / 100)) ]]; then
            log_warn "⚠️  High log growth rate in $location: $growth_rate MB/min"
        fi
    fi
    
    return 0
}

# 主检查函数
main_check() {
    log_info "🔍 Starting log monitoring check..."
    
    # 读取上次状态
    read_previous_state
    local current_time=$(date +%s)
    local time_diff=0
    
    if [[ -n "${LAST_CHECK_TIME:-}" ]]; then
        time_diff=$((current_time - LAST_CHECK_TIME))
    fi
    
    local alert_count=0
    local total_files=0
    
    # 检查Backend日志
    log_info "Checking backend logs..."
    CURRENT_BACKEND_SIZE=0
    if [[ -d "$BACKEND_LOG_DIR" ]]; then
        while IFS= read -r -d '' file; do
            if [[ -f "$file" && "$file" =~ \.log$ ]]; then
                ((total_files++))
                local size_mb=$(get_file_size_mb "$file")
                CURRENT_BACKEND_SIZE=$((CURRENT_BACKEND_SIZE + size_mb))
                
                if ! check_file "$file"; then
                    ((alert_count++))
                fi
            fi
        done < <(find "$BACKEND_LOG_DIR" -name "*.log" -type f -print0 2>/dev/null)
    fi
    
    # 检查Backend增长速度
    if ! check_growth_rate "$CURRENT_BACKEND_SIZE" "${PREVIOUS_BACKEND_SIZE:-0}" "$time_diff" "Backend"; then
        ((alert_count++))
    fi
    
    # 检查Frontend日志
    log_info "Checking frontend logs..."
    CURRENT_FRONTEND_SIZE=0
    if [[ -d "$FRONTEND_LOG_DIR" ]]; then
        while IFS= read -r -d '' file; do
            if [[ -f "$file" && "$file" =~ \.log$ ]]; then
                ((total_files++))
                local size_mb=$(get_file_size_mb "$file")
                CURRENT_FRONTEND_SIZE=$((CURRENT_FRONTEND_SIZE + size_mb))
                
                if ! check_file "$file"; then
                    ((alert_count++))
                fi
            fi
        done < <(find "$FRONTEND_LOG_DIR" -name "*.log" -type f -print0 2>/dev/null)
    fi
    
    # 检查总大小
    CURRENT_TOTAL_SIZE=$((CURRENT_BACKEND_SIZE + CURRENT_FRONTEND_SIZE))
    local total_size_gb=$(echo "scale=2; $CURRENT_TOTAL_SIZE / 1024" | bc -l)
    
    if (( $(echo "$total_size_gb > $CRITICAL_SIZE_GB" | bc -l) )); then
        log_error "🚨 CRITICAL: Total log size exceeded critical threshold: ${total_size_gb}GB > ${CRITICAL_SIZE_GB}GB"
        ((alert_count++))
        
        # 自动清理最老的日志文件
        log_warn "Auto-cleaning old log files..."
        find "$BACKEND_LOG_DIR" "$FRONTEND_LOG_DIR" -name "*.log" -type f -mtime +1 -size +100M -delete 2>/dev/null || true
        
    elif (( $(echo "$total_size_gb > $MAX_TOTAL_SIZE_GB" | bc -l) )); then
        log_warn "⚠️  Total log size warning: ${total_size_gb}GB > ${MAX_TOTAL_SIZE_GB}GB"
    fi
    
    # 统计报告
    log_info "📊 Monitoring Summary:"
    log_info "  - Total files checked: $total_files"
    log_info "  - Backend logs: ${CURRENT_BACKEND_SIZE}MB"
    log_info "  - Frontend logs: ${CURRENT_FRONTEND_SIZE}MB"
    log_info "  - Total size: ${total_size_gb}GB"
    log_info "  - Alerts generated: $alert_count"
    
    if [[ $alert_count -eq 0 ]]; then
        log_success "✅ All log files are within normal limits"
    else
        log_error "❌ Found $alert_count log issues requiring attention"
    fi
    
    # 保存状态
    save_current_state
    
    return $alert_count
}

# 应急清理函数
emergency_cleanup() {
    log_error "🚨 EMERGENCY: Performing emergency log cleanup..."
    
    # 清空超大日志文件
    find "$BACKEND_LOG_DIR" "$FRONTEND_LOG_DIR" -name "*.log" -type f -size +1G -exec sh -c 'echo "# Emergency cleanup at $(date)" > "$1"' _ {} \;
    
    # 删除旧的日志文件
    find "$BACKEND_LOG_DIR" "$FRONTEND_LOG_DIR" -name "*.log" -type f -mtime +7 -delete 2>/dev/null || true
    
    # 压缩中等大小的日志文件
    find "$BACKEND_LOG_DIR" "$FRONTEND_LOG_DIR" -name "*.log" -type f -size +100M -size -1G -exec gzip {} \; 2>/dev/null || true
    
    log_success "Emergency cleanup completed"
}

# 设置cron定时任务
setup_monitoring() {
    log_info "Setting up log monitoring..."
    
    # 创建cron任务
    local cron_entry="*/5 * * * * $PROJECT_ROOT/scripts/log-monitor.sh check >/dev/null 2>&1"
    
    # 检查是否已存在
    if ! crontab -l 2>/dev/null | grep -q "log-monitor.sh"; then
        (crontab -l 2>/dev/null; echo "$cron_entry") | crontab -
        log_success "Log monitoring cron job installed (runs every 5 minutes)"
    else
        log_info "Log monitoring cron job already exists"
    fi
    
    # 创建启动时监控
    if [[ ! -f "$PROJECT_ROOT/startup/log-monitor-startup.sh" ]]; then
        cat > "$PROJECT_ROOT/startup/log-monitor-startup.sh" << 'EOF'
#!/bin/bash
# 启动时检查日志
cd "$(dirname "$0")/.."
./scripts/log-monitor.sh check
EOF
        chmod +x "$PROJECT_ROOT/startup/log-monitor-startup.sh"
        log_success "Startup log monitoring script created"
    fi
}

# 命令行界面
case "${1:-check}" in
    "check")
        main_check
        exit_code=$?
        if [[ $exit_code -gt 5 ]]; then
            log_error "Too many alerts, considering emergency cleanup..."
            emergency_cleanup
        fi
        exit $exit_code
        ;;
    "setup")
        setup_monitoring
        ;;
    "emergency")
        emergency_cleanup
        ;;
    "status")
        if [[ -f "$STATE_FILE" ]]; then
            echo "Last monitoring state:"
            cat "$STATE_FILE"
        else
            echo "No previous monitoring state found"
        fi
        ;;
    "clean")
        log_info "Cleaning up log monitoring state..."
        rm -f "$STATE_FILE"
        log_success "State cleaned"
        ;;
    *)
        echo "Usage: $0 {check|setup|emergency|status|clean}"
        echo "  check     - Run log monitoring check"
        echo "  setup     - Install monitoring cron job"
        echo "  emergency - Perform emergency cleanup"
        echo "  status    - Show monitoring status"
        echo "  clean     - Clean monitoring state"
        exit 1
        ;;
esac