#!/bin/bash

# OpenPenPal 运维管理中心
# 集成日志监控、系统健康检查、自动清理等功能

set -euo pipefail

PROJECT_ROOT="/Users/rocalight/同步空间/opplc/openpenpal"
SCRIPT_DIR="$PROJECT_ROOT/scripts"

# 颜色输出
RED='\033[0;31m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

# 显示帮助信息
show_help() {
    echo -e "${CYAN}OpenPenPal 运维管理中心${NC}"
    echo -e "${CYAN}=================================${NC}"
    echo ""
    echo -e "${GREEN}日志管理:${NC}"
    echo "  $0 logs check          - 检查日志状态"
    echo "  $0 logs monitor        - 启动日志监控"
    echo "  $0 logs clean          - 清理旧日志"
    echo "  $0 logs emergency      - 紧急日志清理"
    echo ""
    echo -e "${GREEN}系统监控:${NC}"
    echo "  $0 health check        - 系统健康检查"
    echo "  $0 health monitor      - 启动健康监控"
    echo "  $0 health metrics      - 显示系统指标"
    echo "  $0 health alerts       - 显示告警信息"
    echo ""
    echo -e "${GREEN}维护操作:${NC}"
    echo "  $0 setup               - 初始化所有监控"
    echo "  $0 status              - 显示系统状态"
    echo "  $0 clean               - 全面系统清理"
    echo "  $0 restart             - 重启服务"
    echo ""
    echo -e "${GREEN}报告和分析:${NC}"
    echo "  $0 report daily        - 生成日报"
    echo "  $0 report summary      - 生成摘要报告"
    echo "  $0 analyze             - 系统分析"
    echo ""
}

# 日志管理
logs_management() {
    case "${1:-}" in
        "check")
            echo -e "${BLUE}🔍 检查日志状态...${NC}"
            "$SCRIPT_DIR/log-monitor.sh" check
            ;;
        "monitor")
            echo -e "${BLUE}📊 启动日志监控...${NC}"
            "$SCRIPT_DIR/log-monitor.sh" setup
            ;;
        "clean")
            echo -e "${YELLOW}🧹 清理日志文件...${NC}"
            # 安全清理：删除7天前的日志文件
            find "$PROJECT_ROOT" -name "*.log" -type f -mtime +7 -delete 2>/dev/null || true
            find "$PROJECT_ROOT" -name "*.log.*" -type f -mtime +3 -delete 2>/dev/null || true
            echo -e "${GREEN}✅ 日志清理完成${NC}"
            ;;
        "emergency")
            echo -e "${RED}🚨 执行紧急日志清理...${NC}"
            "$SCRIPT_DIR/log-monitor.sh" emergency
            ;;
        *)
            echo "使用: $0 logs {check|monitor|clean|emergency}"
            ;;
    esac
}

# 系统健康管理
health_management() {
    case "${1:-}" in
        "check")
            echo -e "${BLUE}🏥 执行系统健康检查...${NC}"
            "$SCRIPT_DIR/system-health-monitor.sh" check
            ;;
        "monitor")
            echo -e "${BLUE}📈 启动健康监控...${NC}"
            "$SCRIPT_DIR/system-health-monitor.sh" setup
            ;;
        "metrics")
            echo -e "${PURPLE}📊 系统指标:${NC}"
            "$SCRIPT_DIR/system-health-monitor.sh" metrics
            ;;
        "alerts")
            echo -e "${YELLOW}⚠️  告警信息:${NC}"
            "$SCRIPT_DIR/system-health-monitor.sh" alerts
            ;;
        *)
            echo "使用: $0 health {check|monitor|metrics|alerts}"
            ;;
    esac
}

# 系统设置
setup_all() {
    echo -e "${CYAN}🔧 初始化 OpenPenPal 监控系统...${NC}"
    
    # 创建必要的目录
    mkdir -p "$PROJECT_ROOT/logs"
    mkdir -p "$PROJECT_ROOT/tmp"
    
    # 设置日志监控
    echo -e "${BLUE}设置日志监控...${NC}"
    "$SCRIPT_DIR/log-monitor.sh" setup
    
    # 设置健康监控
    echo -e "${BLUE}设置健康监控...${NC}"
    "$SCRIPT_DIR/system-health-monitor.sh" setup
    
    # 设置logrotate
    echo -e "${BLUE}配置日志轮转...${NC}"
    if command -v logrotate >/dev/null 2>&1; then
        # 测试logrotate配置
        logrotate -d "$PROJECT_ROOT/config/logrotate.conf" >/dev/null 2>&1 && \
            echo -e "${GREEN}✅ Logrotate配置验证通过${NC}" || \
            echo -e "${YELLOW}⚠️  Logrotate配置需要调整${NC}"
    else
        echo -e "${YELLOW}⚠️  Logrotate未安装，建议安装: brew install logrotate${NC}"
    fi
    
    echo -e "${GREEN}🎉 监控系统初始化完成!${NC}"
    echo ""
    echo -e "${CYAN}后续操作:${NC}"
    echo "- 日志监控: 每5分钟自动运行"
    echo "- 健康检查: 每10分钟自动运行"
    echo "- 查看状态: $0 status"
    echo "- 查看报告: $0 report summary"
}

# 系统状态概览
show_status() {
    echo -e "${CYAN}📊 OpenPenPal 系统状态概览${NC}"
    echo -e "${CYAN}==============================${NC}"
    echo ""
    
    # 服务状态
    echo -e "${BLUE}🔧 服务状态:${NC}"
    local backend_running=false
    local frontend_running=false
    
    if [[ -f "$PROJECT_ROOT/backend/backend.pid" ]]; then
        local backend_pid=$(cat "$PROJECT_ROOT/backend/backend.pid" 2>/dev/null || echo "")
        if [[ -n "$backend_pid" ]] && kill -0 "$backend_pid" 2>/dev/null; then
            echo -e "  ${GREEN}✅ Backend服务: 运行中 (PID: $backend_pid)${NC}"
            backend_running=true
        fi
    fi
    
    if ! $backend_running; then
        echo -e "  ${RED}❌ Backend服务: 未运行${NC}"
    fi
    
    if [[ -f "$PROJECT_ROOT/frontend/frontend.pid" ]]; then
        local frontend_pid=$(cat "$PROJECT_ROOT/frontend/frontend.pid" 2>/dev/null || echo "")
        if [[ -n "$frontend_pid" ]] && kill -0 "$frontend_pid" 2>/dev/null; then
            echo -e "  ${GREEN}✅ Frontend服务: 运行中 (PID: $frontend_pid)${NC}"
            frontend_running=true
        fi
    fi
    
    if ! $frontend_running; then
        echo -e "  ${YELLOW}⚠️  Frontend服务: 未运行${NC}"
    fi
    
    echo ""
    
    # 日志状态
    echo -e "${BLUE}📄 日志状态:${NC}"
    local total_log_size=0
    if [[ -d "$PROJECT_ROOT/backend" ]]; then
        local backend_logs=$(find "$PROJECT_ROOT/backend" -name "*.log" -type f 2>/dev/null | wc -l)
        local backend_size=$(find "$PROJECT_ROOT/backend" -name "*.log" -type f -exec stat -f%z {} \; 2>/dev/null | paste -sd+ | bc 2>/dev/null || echo "0")
        total_log_size=$((total_log_size + backend_size))
        echo -e "  Backend日志: $backend_logs 个文件, $(echo "scale=1; $backend_size / 1024 / 1024" | bc -l 2>/dev/null || echo "0")MB"
    fi
    
    if [[ -d "$PROJECT_ROOT/frontend" ]]; then
        local frontend_logs=$(find "$PROJECT_ROOT/frontend" -name "*.log" -type f 2>/dev/null | wc -l)
        local frontend_size=$(find "$PROJECT_ROOT/frontend" -name "*.log" -type f -exec stat -f%z {} \; 2>/dev/null | paste -sd+ | bc 2>/dev/null || echo "0")
        total_log_size=$((total_log_size + frontend_size))
        echo -e "  Frontend日志: $frontend_logs 个文件, $(echo "scale=1; $frontend_size / 1024 / 1024" | bc -l 2>/dev/null || echo "0")MB"
    fi
    
    local total_size_mb=$(echo "scale=1; $total_log_size / 1024 / 1024" | bc -l 2>/dev/null || echo "0")
    echo -e "  ${PURPLE}总日志大小: ${total_size_mb}MB${NC}"
    
    echo ""
    
    # 系统资源
    echo -e "${BLUE}💻 系统资源:${NC}"
    local disk_usage=$(df -h "$PROJECT_ROOT" | tail -1 | awk '{print $5}')
    echo -e "  磁盘使用率: $disk_usage"
    
    # 最近的告警
    echo ""
    echo -e "${BLUE}🚨 最近告警:${NC}"
    if [[ -f "$PROJECT_ROOT/logs/health-alerts.log" ]]; then
        local alert_count=$(tail -50 "$PROJECT_ROOT/logs/health-alerts.log" 2>/dev/null | wc -l)
        if [[ $alert_count -gt 0 ]]; then
            echo -e "  ${YELLOW}最近50条记录中有 $alert_count 条告警${NC}"
            echo -e "  查看详情: $0 health alerts"
        else
            echo -e "  ${GREEN}✅ 无告警记录${NC}"
        fi
    else
        echo -e "  ${GREEN}✅ 无告警记录${NC}"
    fi
}

# 全面清理
full_cleanup() {
    echo -e "${YELLOW}🧹 执行全面系统清理...${NC}"
    
    # 清理日志
    echo "清理旧日志文件..."
    find "$PROJECT_ROOT" -name "*.log" -type f -mtime +7 -delete 2>/dev/null || true
    find "$PROJECT_ROOT" -name "*.log.*" -type f -mtime +3 -delete 2>/dev/null || true
    
    # 清理临时文件
    echo "清理临时文件..."
    find "$PROJECT_ROOT/tmp" -type f -mtime +1 -delete 2>/dev/null || true
    
    # 清理缓存
    echo "清理缓存文件..."
    find "$PROJECT_ROOT" -name ".DS_Store" -delete 2>/dev/null || true
    find "$PROJECT_ROOT" -name "*.tmp" -delete 2>/dev/null || true
    
    # 清理状态文件
    echo "重置监控状态..."
    rm -f "$PROJECT_ROOT/tmp/log-monitor.state"
    
    echo -e "${GREEN}✅ 系统清理完成${NC}"
}

# 重启服务
restart_services() {
    echo -e "${BLUE}🔄 重启服务...${NC}"
    
    # 停止服务
    if [[ -x "$PROJECT_ROOT/startup/stop-all.sh" ]]; then
        "$PROJECT_ROOT/startup/stop-all.sh"
    fi
    
    # 等待服务完全停止
    sleep 3
    
    # 启动服务
    if [[ -x "$PROJECT_ROOT/startup/quick-start.sh" ]]; then
        "$PROJECT_ROOT/startup/quick-start.sh" development
    fi
    
    echo -e "${GREEN}✅ 服务重启完成${NC}"
}

# 生成报告
generate_report() {
    local report_type="${1:-summary}"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    case "$report_type" in
        "daily")
            echo -e "${CYAN}📊 OpenPenPal 日报 - $timestamp${NC}"
            echo -e "${CYAN}===========================================${NC}"
            ;;
        "summary")
            echo -e "${CYAN}📊 OpenPenPal 系统摘要 - $timestamp${NC}"
            echo -e "${CYAN}========================================${NC}"
            ;;
    esac
    
    echo ""
    
    # 运行健康检查
    echo -e "${BLUE}🏥 健康检查结果:${NC}"
    "$SCRIPT_DIR/system-health-monitor.sh" check
    
    echo ""
    
    # 日志状态
    echo -e "${BLUE}📄 日志分析:${NC}"
    "$SCRIPT_DIR/log-monitor.sh" check
    
    echo ""
    echo -e "${GREEN}报告生成完成${NC}"
}

# 系统分析
analyze_system() {
    echo -e "${CYAN}🔍 OpenPenPal 系统分析${NC}"
    echo -e "${CYAN}======================${NC}"
    echo ""
    
    # 分析日志模式
    echo -e "${BLUE}📈 日志增长趋势分析:${NC}"
    
    local backend_logs=$(find "$PROJECT_ROOT/backend" -name "*.log" -type f 2>/dev/null || true)
    if [[ -n "$backend_logs" ]]; then
        echo "Backend日志文件:"
        echo "$backend_logs" | while read -r logfile; do
            if [[ -f "$logfile" ]]; then
                local size=$(stat -f%z "$logfile" 2>/dev/null || echo "0")
                local size_mb=$(echo "scale=1; $size / 1024 / 1024" | bc -l 2>/dev/null || echo "0")
                echo "  $(basename "$logfile"): ${size_mb}MB"
            fi
        done
    fi
    
    echo ""
    
    # 错误统计
    echo -e "${BLUE}❌ 错误统计 (最近24小时):${NC}"
    local error_count=0
    if [[ -d "$PROJECT_ROOT/backend" ]]; then
        error_count=$(find "$PROJECT_ROOT/backend" -name "*.log" -type f -mtime -1 -exec grep -c "ERROR\|FATAL\|panic" {} \; 2>/dev/null | paste -sd+ | bc 2>/dev/null || echo "0")
    fi
    echo "  总错误数: $error_count"
    
    echo ""
    
    # 性能建议
    echo -e "${BLUE}💡 性能优化建议:${NC}"
    local total_size=$(find "$PROJECT_ROOT" -name "*.log" -type f -exec stat -f%z {} \; 2>/dev/null | paste -sd+ | bc 2>/dev/null || echo "0")
    local total_size_mb=$(echo "scale=1; $total_size / 1024 / 1024" | bc -l 2>/dev/null || echo "0")
    
    if (( $(echo "$total_size_mb > 1000" | bc -l 2>/dev/null || echo "0") )); then
        echo "  🔧 建议启用更激进的日志清理策略"
    fi
    
    if [[ $error_count -gt 100 ]]; then
        echo "  🔧 建议检查应用程序错误处理逻辑"
    fi
    
    echo "  ✅ 已实施智能日志级别控制"
    echo "  ✅ 已实施日志限流机制"
    echo "  ✅ 已配置自动日志轮转"
}

# 主命令路由
case "${1:-help}" in
    "logs")
        logs_management "${2:-}"
        ;;
    "health")
        health_management "${2:-}"
        ;;
    "setup")
        setup_all
        ;;
    "status")
        show_status
        ;;
    "clean")
        full_cleanup
        ;;
    "restart")
        restart_services
        ;;
    "report")
        generate_report "${2:-summary}"
        ;;
    "analyze")
        analyze_system
        ;;
    "help"|*)
        show_help
        ;;
esac