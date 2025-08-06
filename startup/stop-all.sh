#!/bin/bash

# OpenPenPal 停止所有服务脚本
# 安全地停止所有运行中的服务

set -e

# 获取脚本目录和项目根目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# 导入工具函数和环境变量
source "$SCRIPT_DIR/utils.sh"
source "$SCRIPT_DIR/environment-vars.sh"

# 默认选项
FORCE=false
QUIET=false
CLEAN=false
TIMEOUT=10

# 显示帮助信息
show_help() {
    cat << EOF
OpenPenPal 停止所有服务脚本

用法: $0 [选项]

选项:
  --force        强制停止服务（使用 SIGKILL）
  --quiet        静默模式，不显示详细输出
  --clean        停止后清理临时文件和日志
  --timeout=N    等待服务停止的超时时间（秒，默认: 10）
  --help, -h     显示此帮助信息

示例:
  $0                    # 正常停止所有服务
  $0 --force            # 强制停止所有服务
  $0 --clean            # 停止服务并清理文件
  $0 --quiet --timeout=5  # 静默停止，5秒超时

EOF
}

# 解析命令行参数
parse_arguments() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --force)
                FORCE=true
                shift
                ;;
            --quiet)
                QUIET=true
                shift
                ;;
            --clean)
                CLEAN=true
                shift
                ;;
            --timeout=*)
                TIMEOUT="${1#*=}"
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
    
    # 静默模式设置
    if [ "$QUIET" = true ]; then
        export LOG_LEVEL="error"
    fi
}

# 定义服务端口列表
get_service_ports() {
    echo "3000 3001 8000 8080 8001 8002 8003 8004"
}

# 定义PID文件列表
get_pid_files() {
    if [ -d "$LOG_DIR" ]; then
        find "$LOG_DIR" -name "*.pid" -type f 2>/dev/null
    fi
}

# 停止单个服务（通过PID文件）
stop_service_by_pid() {
    local pid_file="$1"
    local service_name="$(basename "$pid_file" .pid)"
    
    if [ ! -f "$pid_file" ]; then
        log_debug "PID文件不存在: $pid_file"
        return 0
    fi
    
    local pid=$(cat "$pid_file" 2>/dev/null)
    
    if [ -z "$pid" ]; then
        log_debug "PID文件为空: $pid_file"
        rm -f "$pid_file"
        return 0
    fi
    
    # 检查进程是否存在
    if ! ps -p $pid >/dev/null 2>&1; then
        log_debug "进程不存在 (PID: $pid, 服务: $service_name)"
        rm -f "$pid_file"
        return 0
    fi
    
    log_info "停止服务: $service_name (PID: $pid)"
    
    # 尝试优雅停止
    if [ "$FORCE" = false ]; then
        kill -TERM $pid 2>/dev/null
        
        # 等待进程退出
        local count=0
        while [ $count -lt $TIMEOUT ] && ps -p $pid >/dev/null 2>&1; do
            sleep 1
            count=$((count + 1))
        done
        
        # 如果进程仍在运行，强制停止
        if ps -p $pid >/dev/null 2>&1; then
            log_warning "进程 $pid 未响应 TERM 信号，强制停止"
            kill -9 $pid 2>/dev/null
            sleep 1
        fi
    else
        # 直接强制停止
        kill -9 $pid 2>/dev/null
        sleep 1
    fi
    
    # 验证进程是否已停止
    if ps -p $pid >/dev/null 2>&1; then
        log_error "无法停止进程 $pid ($service_name)"
        return 1
    else
        log_success "服务已停止: $service_name"
        rm -f "$pid_file"
        return 0
    fi
}

# 停止单个端口上的服务
stop_service_by_port() {
    local port="$1"
    local service_name="$2"
    
    if check_port_available $port; then
        log_debug "端口 $port 没有服务运行"
        return 0
    fi
    
    local pids=$(lsof -ti :$port 2>/dev/null)
    
    if [ -z "$pids" ]; then
        log_debug "端口 $port 没有找到进程"
        return 0
    fi
    
    log_info "停止端口 $port 上的服务${service_name:+ ($service_name)}"
    
    for pid in $pids; do
        log_debug "停止进程 $pid (端口 $port)"
        
        # 尝试优雅停止
        if [ "$FORCE" = false ]; then
            kill -TERM $pid 2>/dev/null
        else
            kill -9 $pid 2>/dev/null
        fi
    done
    
    # 等待进程退出
    if [ "$FORCE" = false ]; then
        local count=0
        while [ $count -lt $TIMEOUT ]; do
            if check_port_available $port; then
                break
            fi
            sleep 1
            count=$((count + 1))
        done
        
        # 如果端口仍被占用，强制停止
        if check_port_occupied $port; then
            log_warning "端口 $port 仍被占用，强制停止"
            kill_port $port "KILL"
        fi
    fi
    
    # 验证端口是否已释放
    if check_port_occupied $port; then
        log_error "无法释放端口 $port"
        return 1
    else
        log_success "端口 $port 已释放"
        return 0
    fi
}

# 停止所有已知的服务进程
stop_known_processes() {
    log_info "停止已知的服务进程..."
    
    # 定义进程名称模式
    local process_patterns=(
        "mock-services.js"
        "simple-mock-services.js"
        "node.*src/index.js"
        "npm.*run.*dev"
        "vite"
        "webpack-dev-server"
    )
    
    for pattern in "${process_patterns[@]}"; do
        local pids=$(pgrep -f "$pattern" 2>/dev/null || true)
        
        if [ -n "$pids" ]; then
            log_info "发现匹配进程: $pattern"
            for pid in $pids; do
                log_debug "停止进程: $pid ($pattern)"
                
                if [ "$FORCE" = true ]; then
                    kill -9 $pid 2>/dev/null || true
                else
                    kill -TERM $pid 2>/dev/null || true
                fi
            done
        fi
    done
    
    # 如果不是强制模式，等待进程退出
    if [ "$FORCE" = false ]; then
        sleep 2
        
        # 检查是否还有残留进程
        for pattern in "${process_patterns[@]}"; do
            local remaining_pids=$(pgrep -f "$pattern" 2>/dev/null || true)
            
            if [ -n "$remaining_pids" ]; then
                log_warning "强制停止残留进程: $pattern"
                for pid in $remaining_pids; do
                    kill -9 $pid 2>/dev/null || true
                done
            fi
        done
    fi
}

# 通过PID文件停止服务
stop_services_by_pid() {
    log_info "通过PID文件停止服务..."
    
    local pid_files=$(get_pid_files)
    local stopped_count=0
    local failed_count=0
    
    if [ -z "$pid_files" ]; then
        log_debug "没有找到PID文件"
        return 0
    fi
    
    for pid_file in $pid_files; do
        if stop_service_by_pid "$pid_file"; then
            stopped_count=$((stopped_count + 1))
        else
            failed_count=$((failed_count + 1))
        fi
    done
    
    if [ $stopped_count -gt 0 ]; then
        log_success "通过PID文件停止了 $stopped_count 个服务"
    fi
    
    if [ $failed_count -gt 0 ]; then
        log_warning "有 $failed_count 个服务停止失败"
    fi
}

# 通过端口停止服务
stop_services_by_port() {
    log_info "通过端口停止服务..."
    
    local ports=($(get_service_ports))
    local stopped_count=0
    local failed_count=0
    
    # 定义端口和服务名称的映射函数
    get_service_name() {
        local port="$1"
        case $port in
            3000) echo "前端应用" ;;
            3001) echo "管理后台" ;;
            8000) echo "API网关/Mock服务" ;;
            8080) echo "主后端服务" ;;
            8001) echo "写信服务" ;;
            8002) echo "信使服务" ;;
            8003) echo "管理服务" ;;
            8004) echo "OCR服务" ;;
            *) echo "未知服务" ;;
        esac
    }
    
    for port in "${ports[@]}"; do
        local service_name="$(get_service_name "$port")"
        
        if stop_service_by_port "$port" "$service_name"; then
            if check_port_occupied $port; then
                failed_count=$((failed_count + 1))
            else
                stopped_count=$((stopped_count + 1))
            fi
        else
            failed_count=$((failed_count + 1))
        fi
    done
    
    if [ $stopped_count -gt 0 ]; then
        log_success "通过端口停止了 $stopped_count 个服务"
    fi
    
    if [ $failed_count -gt 0 ]; then
        log_warning "有 $failed_count 个服务停止失败"
    fi
}

# 清理临时文件
cleanup_files() {
    if [ "$CLEAN" = false ]; then
        return 0
    fi
    
    log_info "清理临时文件..."
    
    # 清理PID文件
    if [ -d "$LOG_DIR" ]; then
        rm -f "$LOG_DIR"/*.pid
        log_debug "已清理PID文件"
    fi
    
    # 清理临时Mock服务文件
    rm -f "$PROJECT_ROOT/mock-services.js"
    log_debug "已清理临时Mock服务文件"
    
    # 清理日志文件（可选）
    if confirm "是否清理日志文件？"; then
        if [ -d "$LOG_DIR" ]; then
            rm -f "$LOG_DIR"/*.log
            log_debug "已清理日志文件"
        fi
    fi
    
    # 清理缓存目录
    local cache_dirs=(
        "$PROJECT_ROOT/tmp"
        "$PROJECT_ROOT/cache"
        "$PROJECT_ROOT/.cache"
    )
    
    for cache_dir in "${cache_dirs[@]}"; do
        if [ -d "$cache_dir" ]; then
            rm -rf "$cache_dir"
            log_debug "已清理缓存目录: $cache_dir"
        fi
    done
    
    log_success "临时文件清理完成"
}

# 验证所有服务已停止
verify_all_stopped() {
    log_info "验证所有服务已停止..."
    
    local ports=($(get_service_ports))
    local running_services=()
    
    for port in "${ports[@]}"; do
        if check_port_occupied $port; then
            running_services+=("$port")
        fi
    done
    
    if [ ${#running_services[@]} -eq 0 ]; then
        log_success "所有服务已成功停止"
        return 0
    else
        log_warning "以下端口仍有服务运行: ${running_services[*]}"
        
        if [ "$FORCE" = false ] && confirm "是否强制停止残留服务？"; then
            for port in "${running_services[@]}"; do
                kill_port "$port" "KILL"
            done
            
            sleep 1
            verify_all_stopped
        fi
        
        return 1
    fi
}

# 显示停止结果
show_result() {
    log_info ""
    log_info "🛑 OpenPenPal 服务停止完成"
    log_info "=========================="
    
    # 显示端口状态
    local ports=($(get_service_ports))
    local running_count=0
    
    for port in "${ports[@]}"; do
        if check_port_occupied $port; then
            log_warning "端口 $port: 仍有服务运行"
            running_count=$((running_count + 1))
        else
            log_success "端口 $port: 已释放"
        fi
    done
    
    log_info ""
    
    if [ $running_count -eq 0 ]; then
        log_success "✅ 所有服务已成功停止"
        log_info "现在可以安全地重新启动服务"
    else
        log_warning "⚠️ 有 $running_count 个服务仍在运行"
        log_info "您可以使用 --force 选项强制停止所有服务"
    fi
    
    log_info ""
    log_info "💡 常用命令:"
    log_info "  • 重新启动: ./startup/quick-start.sh"
    log_info "  • 检查状态: ./startup/check-status.sh"
    log_info "  • 强制停止: ./startup/stop-all.sh --force"
    
    if [ "$CLEAN" = true ]; then
        log_info "  • 已清理临时文件和缓存"
    fi
}

# 主函数
main() {
    # 解析参数
    parse_arguments "$@"
    
    # 显示停止信息
    if [ "$QUIET" = false ]; then
        log_info "🛑 OpenPenPal 服务停止器"
        log_info "强制模式: $FORCE"
        log_info "清理文件: $CLEAN"
        log_info "超时时间: ${TIMEOUT}秒"
        log_info ""
    fi
    
    # 停止流程
    stop_services_by_pid
    stop_services_by_port
    stop_known_processes
    
    # 最终验证
    verify_all_stopped
    
    # 清理文件
    cleanup_files
    
    # 显示结果
    if [ "$QUIET" = false ]; then
        show_result
    fi
}

# 执行主函数
main "$@"