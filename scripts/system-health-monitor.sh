#!/bin/bash

# OpenPenPal 系统健康监控脚本
# 综合监控系统资源、服务状态、数据库连接、日志健康度

set -euo pipefail

# 配置
PROJECT_ROOT="/Users/rocalight/同步空间/opplc/openpenpal"
HEALTH_LOG="$PROJECT_ROOT/logs/health-monitor.log"
ALERT_LOG="$PROJECT_ROOT/logs/health-alerts.log"
METRICS_FILE="$PROJECT_ROOT/tmp/health-metrics.json"

# 阈值配置
CPU_THRESHOLD=80           # CPU使用率告警阈值
MEMORY_THRESHOLD=85        # 内存使用率告警阈值
DISK_THRESHOLD=90          # 磁盘使用率告警阈值
LOG_SIZE_THRESHOLD_GB=2    # 日志大小告警阈值
DB_RESPONSE_THRESHOLD=5    # 数据库响应时间告警阈值(秒)

# 颜色输出
RED='\033[0;31m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m'

# 确保目录存在
mkdir -p "$(dirname "$HEALTH_LOG")"
mkdir -p "$(dirname "$METRICS_FILE")"

# 日志函数
log_message() {
    local level="$1"
    local message="$2"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo "[$timestamp] [$level] $message" | tee -a "$HEALTH_LOG"
}

log_info() {
    log_message "INFO" "$1"
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_warn() {
    log_message "WARN" "$1"
    echo -e "${YELLOW}[WARN]${NC} $1"
    echo "[$timestamp] [WARN] $1" >> "$ALERT_LOG"
}

log_error() {
    log_message "ERROR" "$1"
    echo -e "${RED}[ERROR]${NC} $1"
    echo "[$timestamp] [ERROR] $1" >> "$ALERT_LOG"
}

log_success() {
    log_message "SUCCESS" "$1"
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_metric() {
    log_message "METRIC" "$1"
    echo -e "${PURPLE}[METRIC]${NC} $1"
}

# 系统资源检查
check_system_resources() {
    log_info "🖥️  Checking system resources..."
    
    local alerts=0
    
    # CPU检查
    local cpu_usage
    if command -v top >/dev/null 2>&1; then
        cpu_usage=$(top -l 1 | grep "CPU usage" | awk '{print $3}' | sed 's/%//')
    else
        cpu_usage=0
    fi
    
    log_metric "CPU Usage: ${cpu_usage}%"
    if (( $(echo "$cpu_usage > $CPU_THRESHOLD" | bc -l 2>/dev/null || echo "0") )); then
        log_error "🚨 High CPU usage: ${cpu_usage}% > ${CPU_THRESHOLD}%"
        ((alerts++))
    fi
    
    # 内存检查
    local memory_usage
    if command -v vm_stat >/dev/null 2>&1; then
        # macOS
        local pages_free=$(vm_stat | grep "Pages free" | awk '{print $3}' | sed 's/\.//')
        local pages_total=$(vm_stat | grep -E "(Pages free|Pages active|Pages inactive|Pages speculative|Pages throttled|Pages wired down)" | awk '{print $3}' | sed 's/\.//' | paste -sd+ | bc)
        memory_usage=$(echo "scale=1; 100 - ($pages_free * 100 / $pages_total)" | bc -l)
    else
        # Linux
        memory_usage=$(free | grep Mem | awk '{printf "%.1f", $3/$2 * 100.0}')
    fi
    
    log_metric "Memory Usage: ${memory_usage}%"
    if (( $(echo "$memory_usage > $MEMORY_THRESHOLD" | bc -l 2>/dev/null || echo "0") )); then
        log_error "🚨 High memory usage: ${memory_usage}% > ${MEMORY_THRESHOLD}%"
        ((alerts++))
    fi
    
    # 磁盘检查
    local disk_usage=$(df -h "$PROJECT_ROOT" | tail -1 | awk '{print $5}' | sed 's/%//')
    log_metric "Disk Usage: ${disk_usage}%"
    if [[ $disk_usage -gt $DISK_THRESHOLD ]]; then
        log_error "🚨 High disk usage: ${disk_usage}% > ${DISK_THRESHOLD}%"
        ((alerts++))
    fi
    
    return $alerts
}

# 服务状态检查
check_services() {
    log_info "🔧 Checking service status..."
    
    local alerts=0
    
    # 检查后端服务
    local backend_pid=""
    if [[ -f "$PROJECT_ROOT/backend/backend.pid" ]]; then
        backend_pid=$(cat "$PROJECT_ROOT/backend/backend.pid" 2>/dev/null || echo "")
    fi
    
    if [[ -n "$backend_pid" ]] && kill -0 "$backend_pid" 2>/dev/null; then
        log_success "✅ Backend service running (PID: $backend_pid)"
        
        # 检查端口监听
        if lsof -i :8080 >/dev/null 2>&1; then
            log_success "✅ Backend port 8080 is listening"
        else
            log_error "❌ Backend port 8080 not listening"
            ((alerts++))
        fi
    else
        log_error "❌ Backend service not running"
        ((alerts++))
    fi
    
    # 检查前端服务
    local frontend_pid=""
    if [[ -f "$PROJECT_ROOT/frontend/frontend.pid" ]]; then
        frontend_pid=$(cat "$PROJECT_ROOT/frontend/frontend.pid" 2>/dev/null || echo "")
    fi
    
    if [[ -n "$frontend_pid" ]] && kill -0 "$frontend_pid" 2>/dev/null; then
        log_success "✅ Frontend service running (PID: $frontend_pid)"
        
        # 检查端口监听
        if lsof -i :3000 >/dev/null 2>&1; then
            log_success "✅ Frontend port 3000 is listening"
        else
            log_warn "⚠️  Frontend port 3000 not listening"
        fi
    else
        log_warn "⚠️  Frontend service not running"
    fi
    
    return $alerts
}

# 数据库连接检查
check_database() {
    log_info "🗄️  Checking database connection..."
    
    local alerts=0
    local start_time=$(date +%s.%N)
    
    # 检查PostgreSQL连接
    if command -v psql >/dev/null 2>&1; then
        local db_url="${DATABASE_URL:-postgres://openpenpal_user:password@localhost:5432/openpenpal}"
        
        if psql "$db_url" -c "SELECT 1;" >/dev/null 2>&1; then
            local end_time=$(date +%s.%N)
            local response_time=$(echo "$end_time - $start_time" | bc -l)
            
            log_metric "Database response time: ${response_time}s"
            log_success "✅ Database connection successful"
            
            if (( $(echo "$response_time > $DB_RESPONSE_THRESHOLD" | bc -l) )); then
                log_warn "⚠️  Slow database response: ${response_time}s > ${DB_RESPONSE_THRESHOLD}s"
            fi
        else
            log_error "❌ Database connection failed"
            ((alerts++))
        fi
    else
        log_warn "⚠️  psql not available, skipping database check"
    fi
    
    return $alerts
}

# 日志健康度检查
check_log_health() {
    log_info "📄 Checking log health..."
    
    local alerts=0
    
    # 检查日志总大小
    local total_size=0
    if [[ -d "$PROJECT_ROOT/backend" ]]; then
        local backend_size=$(find "$PROJECT_ROOT/backend" -name "*.log" -type f -exec stat -f%z {} \; 2>/dev/null | paste -sd+ | bc || echo "0")
        total_size=$((total_size + backend_size))
    fi
    
    if [[ -d "$PROJECT_ROOT/frontend" ]]; then
        local frontend_size=$(find "$PROJECT_ROOT/frontend" -name "*.log" -type f -exec stat -f%z {} \; 2>/dev/null | paste -sd+ | bc || echo "0")
        total_size=$((total_size + frontend_size))
    fi
    
    local total_size_gb=$(echo "scale=2; $total_size / 1024 / 1024 / 1024" | bc -l)
    log_metric "Total log size: ${total_size_gb}GB"
    
    if (( $(echo "$total_size_gb > $LOG_SIZE_THRESHOLD_GB" | bc -l) )); then
        log_error "🚨 Large log files detected: ${total_size_gb}GB > ${LOG_SIZE_THRESHOLD_GB}GB"
        ((alerts++))
        
        # 自动触发日志清理
        log_warn "Triggering automatic log cleanup..."
        if [[ -x "$PROJECT_ROOT/scripts/log-monitor.sh" ]]; then
            "$PROJECT_ROOT/scripts/log-monitor.sh" emergency
        fi
    fi
    
    # 检查错误日志模式
    local error_count=0
    if [[ -d "$PROJECT_ROOT/backend" ]]; then
        error_count=$(find "$PROJECT_ROOT/backend" -name "*.log" -type f -mtime -1 -exec grep -c "ERROR\|FATAL\|panic" {} \; 2>/dev/null | paste -sd+ | bc || echo "0")
    fi
    
    log_metric "Recent errors: $error_count"
    if [[ $error_count -gt 100 ]]; then
        log_error "🚨 High error rate: $error_count errors in last 24h"
        ((alerts++))
    elif [[ $error_count -gt 50 ]]; then
        log_warn "⚠️  Elevated error rate: $error_count errors in last 24h"
    fi
    
    return $alerts
}

# 网络连接检查
check_network() {
    log_info "🌐 Checking network connectivity..."
    
    local alerts=0
    
    # 检查本地健康检查端点
    if curl -s --max-time 10 http://localhost:8080/health >/dev/null 2>&1; then
        log_success "✅ Backend health endpoint accessible"
    else
        log_error "❌ Backend health endpoint unreachable"
        ((alerts++))
    fi
    
    if curl -s --max-time 10 http://localhost:3000 >/dev/null 2>&1; then
        log_success "✅ Frontend accessible"
    else
        log_warn "⚠️  Frontend unreachable"
    fi
    
    # 检查外部API连接（如果配置了）
    if [[ -n "${MOONSHOT_API_KEY:-}" ]]; then
        if curl -s --max-time 10 https://api.moonshot.cn/v1/models >/dev/null 2>&1; then
            log_success "✅ Moonshot API accessible"
        else
            log_warn "⚠️  Moonshot API unreachable"
        fi
    fi
    
    return $alerts
}

# 生成健康度指标JSON
generate_metrics() {
    local total_alerts="$1"
    local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    
    cat > "$METRICS_FILE" << EOF
{
  "timestamp": "$timestamp",
  "status": "$([ $total_alerts -eq 0 ] && echo "healthy" || echo "unhealthy")",
  "total_alerts": $total_alerts,
  "checks": {
    "system_resources": {
      "cpu_usage": "${cpu_usage:-0}",
      "memory_usage": "${memory_usage:-0}",
      "disk_usage": "${disk_usage:-0}"
    },
    "services": {
      "backend_running": $([ -n "${backend_pid:-}" ] && echo "true" || echo "false"),
      "frontend_running": $([ -n "${frontend_pid:-}" ] && echo "true" || echo "false")
    },
    "database": {
      "connected": $([ $db_alerts -eq 0 ] && echo "true" || echo "false"),
      "response_time": "${response_time:-0}"
    },
    "logs": {
      "total_size_gb": "${total_size_gb:-0}",
      "error_count": "${error_count:-0}"
    }
  }
}
EOF
    
    log_info "📊 Metrics saved to: $METRICS_FILE"
}

# 主健康检查函数
main_health_check() {
    log_info "🏥 Starting comprehensive health check..."
    
    local total_alerts=0
    
    # 系统资源检查
    check_system_resources
    total_alerts=$((total_alerts + $?))
    
    # 服务状态检查
    check_services
    total_alerts=$((total_alerts + $?))
    
    # 数据库检查
    check_database
    local db_alerts=$?
    total_alerts=$((total_alerts + db_alerts))
    
    # 日志健康度检查
    check_log_health
    total_alerts=$((total_alerts + $?))
    
    # 网络连接检查
    check_network
    total_alerts=$((total_alerts + $?))
    
    # 生成指标
    generate_metrics "$total_alerts"
    
    # 总结报告
    log_info "📋 Health Check Summary:"
    log_info "  - Total alerts: $total_alerts"
    log_info "  - Status: $([ $total_alerts -eq 0 ] && echo "HEALTHY ✅" || echo "NEEDS ATTENTION ⚠️")"
    
    if [[ $total_alerts -eq 0 ]]; then
        log_success "🎉 All systems are healthy!"
    else
        log_error "⚠️  Found $total_alerts issues requiring attention"
        
        # 如果告警过多，触发自动修复
        if [[ $total_alerts -gt 5 ]]; then
            log_warn "🔧 Too many alerts, triggering auto-recovery..."
            auto_recovery
        fi
    fi
    
    return $total_alerts
}

# 自动恢复函数
auto_recovery() {
    log_warn "🔧 Starting automatic recovery procedures..."
    
    # 清理日志
    if [[ -x "$PROJECT_ROOT/scripts/log-monitor.sh" ]]; then
        "$PROJECT_ROOT/scripts/log-monitor.sh" emergency
    fi
    
    # 重启服务（如果需要）
    # 这里可以添加服务重启逻辑
    
    log_success "🔧 Auto-recovery completed"
}

# 设置定时监控
setup_health_monitoring() {
    log_info "Setting up health monitoring..."
    
    # 创建cron任务
    local cron_entry="*/10 * * * * $PROJECT_ROOT/scripts/system-health-monitor.sh check >/dev/null 2>&1"
    
    if ! crontab -l 2>/dev/null | grep -q "system-health-monitor.sh"; then
        (crontab -l 2>/dev/null; echo "$cron_entry") | crontab -
        log_success "Health monitoring cron job installed (runs every 10 minutes)"
    else
        log_info "Health monitoring cron job already exists"
    fi
}

# 命令行界面
case "${1:-check}" in
    "check")
        main_health_check
        exit $?
        ;;
    "setup")
        setup_health_monitoring
        ;;
    "metrics")
        if [[ -f "$METRICS_FILE" ]]; then
            cat "$METRICS_FILE" | jq . 2>/dev/null || cat "$METRICS_FILE"
        else
            echo "No metrics file found"
        fi
        ;;
    "alerts")
        if [[ -f "$ALERT_LOG" ]]; then
            tail -20 "$ALERT_LOG"
        else
            echo "No alerts found"
        fi
        ;;
    *)
        echo "Usage: $0 {check|setup|metrics|alerts}"
        echo "  check   - Run comprehensive health check"
        echo "  setup   - Install monitoring cron job"
        echo "  metrics - Show latest metrics"
        echo "  alerts  - Show recent alerts"
        exit 1
        ;;
esac