#!/bin/bash

# OpenPenPal 服务状态检查脚本
# 检查所有服务的运行状态和健康状况

set -e

# 获取脚本目录和项目根目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# 导入工具函数和环境变量
source "$SCRIPT_DIR/utils.sh"
source "$SCRIPT_DIR/environment-vars.sh"

# 确保本地服务不使用代理
export NO_PROXY="localhost,127.0.0.1,*.local"
export no_proxy="localhost,127.0.0.1,*.local"

# 默认选项
QUIET=false
DETAILED=false
CONTINUOUS=false
INTERVAL=10

# 显示帮助信息
show_help() {
    cat << EOF
OpenPenPal 服务状态检查脚本

用法: $0 [选项]

选项:
  --quiet        静默模式，只显示错误
  --detailed     显示详细信息
  --continuous   持续监控模式
  --interval=N   持续监控间隔（秒，默认: 10）
  --help, -h     显示此帮助信息

示例:
  $0                    # 检查所有服务状态
  $0 --detailed         # 显示详细信息
  $0 --continuous       # 持续监控
  $0 --quiet            # 静默模式

EOF
}

# 解析命令行参数
parse_arguments() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --quiet)
                QUIET=true
                export LOG_LEVEL="error"
                shift
                ;;
            --detailed)
                DETAILED=true
                export LOG_LEVEL="debug"
                shift
                ;;
            --continuous)
                CONTINUOUS=true
                shift
                ;;
            --interval=*)
                INTERVAL="${1#*=}"
                shift
                ;;
            --help|-h)
                show_help
                exit 0
                ;;
            *)
                log_error "未知参数: $1"
                show_help
                exit 1
                ;;
        esac
    done
}

# 检查单个服务状态
check_service_status() {
    local service_name="$1"
    local port="$2"
    local pid_file="$3"
    local health_url="$4"
    
    local status="unknown"
    local pid=""
    local response_code=""
    local response_time=""
    
    # 检查PID文件
    if [ -n "$pid_file" ] && [ -f "$pid_file" ]; then
        pid=$(cat "$pid_file" 2>/dev/null)
        if [ -n "$pid" ] && ps -p $pid >/dev/null 2>&1; then
            status="running"
        else
            # PID文件存在但进程不存在
            rm -f "$pid_file"
            status="stopped"
        fi
    fi
    
    # 检查端口
    if [ -n "$port" ]; then
        if check_port_occupied $port; then
            if [ "$status" = "unknown" ]; then
                status="running"
                # 获取端口对应的PID
                pid=$(get_pid_by_port $port)
            fi
        else
            if [ "$status" = "running" ]; then
                status="port_mismatch"
            else
                status="stopped"
            fi
        fi
    fi
    
    # 健康检查
    if [ -n "$health_url" ] && [ "$status" = "running" ]; then
        # macOS兼容的时间计算（秒级精度）
        local start_time=$(date +%s)
        # 调试输出
        if [ "$DETAILED" = true ]; then
            echo "DEBUG: Testing URL: $health_url" >&2
        fi
        response_code=$(curl -s -o /dev/null -w "%{http_code}" --noproxy localhost,127.0.0.1 --connect-timeout 5 --max-time 10 "$health_url" 2>/dev/null || echo "000")
        local end_time=$(date +%s)
        response_time=$((end_time - start_time))
        
        if [ "$response_code" = "200" ]; then
            status="healthy"
        elif [ "$response_code" != "000" ]; then
            status="unhealthy"
        else
            status="unreachable"
        fi
    fi
    
    # 返回结果
    echo "$status|$pid|$response_code|$response_time"
}

# 显示服务状态
display_service_status() {
    local service_name="$1"
    local port="$2"
    local status_info="$3"
    
    local status=$(echo "$status_info" | cut -d'|' -f1)
    local pid=$(echo "$status_info" | cut -d'|' -f2)
    local response_code=$(echo "$status_info" | cut -d'|' -f3)
    local response_time=$(echo "$status_info" | cut -d'|' -f4)
    
    local status_icon=""
    local status_color=""
    local status_text=""
    
    case $status in
        healthy)
            status_icon="✅"
            status_color="$GREEN"
            status_text="健康"
            ;;
        running)
            status_icon="🟢"
            status_color="$GREEN"
            status_text="运行中"
            ;;
        unhealthy)
            status_icon="⚠️"
            status_color="$YELLOW"
            status_text="不健康"
            ;;
        unreachable)
            status_icon="🔴"
            status_color="$RED"
            status_text="无法访问"
            ;;
        stopped)
            status_icon="⭕"
            status_color="$RED"
            status_text="已停止"
            ;;
        port_mismatch)
            status_icon="⚠️"
            status_color="$YELLOW"
            status_text="端口不匹配"
            ;;
        *)
            status_icon="❓"
            status_color="$PURPLE"
            status_text="未知"
            ;;
    esac
    
    if [ "$QUIET" = false ]; then
        printf "%-20s %s %-12s" "$service_name" "$status_icon" "$status_text"
        
        if [ -n "$port" ]; then
            printf " (端口: %s)" "$port"
        fi
        
        if [ -n "$pid" ]; then
            printf " [PID: %s]" "$pid"
        fi
        
        if [ "$DETAILED" = true ]; then
            if [ -n "$response_code" ] && [ "$response_code" != "" ]; then
                printf " HTTP: %s" "$response_code"
            fi
            if [ -n "$response_time" ] && [ "$response_time" != "" ]; then
                printf " (%sms)" "$response_time"
            fi
        fi
        
        echo ""
    fi
    
    # 返回状态代码
    case $status in
        healthy|running) return 0 ;;
        *) return 1 ;;
    esac
}

# 检查所有服务
check_all_services() {
    if [ "$QUIET" = false ]; then
        log_info "🔍 OpenPenPal 服务状态检查"
        log_info "=========================="
        echo ""
    fi
    
    local total_services=0
    local healthy_services=0
    
    # 定义服务列表
    local services=(
        "前端应用:$FRONTEND_PORT:$LOG_DIR/frontend.pid:http://localhost:$FRONTEND_PORT"
        "管理后台:$ADMIN_FRONTEND_PORT:$LOG_DIR/admin-frontend.pid:http://localhost:$ADMIN_FRONTEND_PORT"
        "API网关:$GATEWAY_PORT:$LOG_DIR/gateway.pid:http://localhost:$GATEWAY_PORT/health"
        "主后端:$BACKEND_PORT:$LOG_DIR/backend.pid:http://localhost:$BACKEND_PORT/health"
        "写信服务:$WRITE_SERVICE_PORT:$LOG_DIR/write-service.pid:http://localhost:$WRITE_SERVICE_PORT/health"
        "信使服务:$COURIER_SERVICE_PORT:$LOG_DIR/courier-service.pid:http://localhost:$COURIER_SERVICE_PORT/health"
        "管理服务:$ADMIN_SERVICE_PORT:$LOG_DIR/admin-service.pid:http://localhost:$ADMIN_SERVICE_PORT/health"
        "OCR服务:$OCR_SERVICE_PORT:$LOG_DIR/ocr-service.pid:http://localhost:$OCR_SERVICE_PORT/health"
        "简化Mock:$GATEWAY_PORT:$LOG_DIR/simple-mock.pid:http://localhost:$GATEWAY_PORT/health"
    )
    
    for service_info in "${services[@]}"; do
        local service_name=$(echo "$service_info" | cut -d':' -f1)
        local port=$(echo "$service_info" | cut -d':' -f2)
        local pid_file=$(echo "$service_info" | cut -d':' -f3)
        local health_url=$(echo "$service_info" | cut -d':' -f4-)
        
        # 跳过不存在的PID文件对应的服务（说明该服务未启动）
        if [ ! -f "$pid_file" ] && ! check_port_occupied $port; then
            continue
        fi
        
        total_services=$((total_services + 1))
        
        local status_info=$(check_service_status "$service_name" "$port" "$pid_file" "$health_url")
        
        if display_service_status "$service_name" "$port" "$status_info"; then
            healthy_services=$((healthy_services + 1))
        fi
    done
    
    if [ "$QUIET" = false ]; then
        echo ""
        log_info "状态总结:"
        log_info "========="
        
        if [ $total_services -eq 0 ]; then
            log_warning "没有检测到运行中的服务"
        else
            if [ $healthy_services -eq $total_services ]; then
                log_success "所有服务运行正常 ($healthy_services/$total_services)"
            else
                log_warning "部分服务存在问题 ($healthy_services/$total_services)"
            fi
        fi
        
        echo ""
        log_info "💡 常用命令:"
        log_info "  • 启动服务: ./startup/quick-start.sh"
        log_info "  • 停止服务: ./startup/stop-all.sh"
        log_info "  • 查看日志: tail -f logs/*.log"
        echo ""
    fi
    
    # 返回健康状态
    if [ $total_services -eq 0 ]; then
        return 2  # 没有服务运行
    elif [ $healthy_services -eq $total_services ]; then
        return 0  # 所有服务健康
    else
        return 1  # 部分服务有问题
    fi
}

# 持续监控模式
continuous_monitoring() {
    log_info "🔄 启动持续监控模式 (间隔: ${INTERVAL}秒)"
    log_info "按 Ctrl+C 停止监控"
    echo ""
    
    # 设置退出处理
    trap 'log_info "停止监控"; exit 0' INT TERM
    
    while true; do
        clear
        echo "$(date '+%Y-%m-%d %H:%M:%S') - OpenPenPal 服务监控"
        echo "=============================================="
        echo ""
        
        check_all_services
        
        echo ""
        echo "下次检查: $(date -d "+${INTERVAL} seconds" '+%H:%M:%S')"
        echo "按 Ctrl+C 停止监控"
        
        sleep $INTERVAL
    done
}

# 主函数
main() {
    # 解析参数
    parse_arguments "$@"
    
    # 进入项目根目录
    cd "$PROJECT_ROOT"
    
    if [ "$CONTINUOUS" = true ]; then
        continuous_monitoring
    else
        check_all_services
        exit $?
    fi
}

# 执行主函数
main "$@"