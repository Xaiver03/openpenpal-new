#!/bin/bash

# OpenPenPal 强制端口清理脚本
# 强制释放所有项目相关端口

set -e

# 获取脚本目录和项目根目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# 导入工具函数
source "$SCRIPT_DIR/utils.sh"

# 定义需要清理的端口
PORTS=(3000 3001 8000 8001 8002 8003 8004 8080)

# 显示帮助信息
show_help() {
    cat << EOF
OpenPenPal 强制端口清理脚本

用法: $0 [选项]

选项:
  --quiet        静默模式
  --help, -h     显示此帮助信息

说明:
  强制释放 OpenPenPal 项目使用的所有端口

EOF
}

# 解析参数
parse_arguments() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --quiet)
                export LOG_LEVEL="error"
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

# 强制清理单个端口
force_cleanup_port() {
    local port="$1"
    
    log_debug "检查端口 $port..."
    
    # 查找占用端口的进程
    local pids=$(lsof -ti :$port 2>/dev/null || true)
    
    if [ -n "$pids" ]; then
        log_info "清理端口 $port (PID: $pids)"
        
        # 先尝试 TERM 信号
        echo $pids | xargs kill -TERM 2>/dev/null || true
        sleep 1
        
        # 检查是否还有进程
        local remaining_pids=$(lsof -ti :$port 2>/dev/null || true)
        if [ -n "$remaining_pids" ]; then
            log_warning "强制终止端口 $port 上的进程"
            echo $remaining_pids | xargs kill -9 2>/dev/null || true
            sleep 1
        fi
        
        # 最终验证
        if lsof -ti :$port >/dev/null 2>&1; then
            log_error "无法释放端口 $port"
            return 1
        else
            log_success "端口 $port 已释放"
            return 0
        fi
    else
        log_debug "端口 $port 空闲"
        return 0
    fi
}

# 清理所有Node.js相关进程
cleanup_node_processes() {
    log_info "清理 Node.js 相关进程..."
    
    # 查找可能的 OpenPenPal 相关进程
    local process_patterns=(
        "mock-services"
        "simple-mock"
        "gateway"
        "npm.*run.*dev"
        "vite"
        "next"
    )
    
    for pattern in "${process_patterns[@]}"; do
        local pids=$(pgrep -f "$pattern" 2>/dev/null || true)
        
        if [ -n "$pids" ]; then
            log_info "清理进程模式: $pattern"
            for pid in $pids; do
                log_debug "终止进程: $pid"
                kill -TERM $pid 2>/dev/null || true
            done
        fi
    done
    
    # 等待进程退出
    sleep 2
    
    # 强制清理残留进程
    for pattern in "${process_patterns[@]}"; do
        local remaining_pids=$(pgrep -f "$pattern" 2>/dev/null || true)
        
        if [ -n "$remaining_pids" ]; then
            log_warning "强制清理残留进程: $pattern"
            for pid in $remaining_pids; do
                kill -9 $pid 2>/dev/null || true
            done
        fi
    done
}

# 清理PID文件
cleanup_pid_files() {
    log_info "清理PID文件..."
    
    if [ -d "$PROJECT_ROOT/logs" ]; then
        rm -f "$PROJECT_ROOT/logs"/*.pid
        log_debug "已清理PID文件"
    fi
}

# 主函数
main() {
    # 解析参数
    parse_arguments "$@"
    
    log_info "🧹 OpenPenPal 强制端口清理"
    log_info "========================="
    
    # 清理Node.js进程
    cleanup_node_processes
    
    # 清理端口
    log_info "清理项目端口..."
    local failed_ports=()
    
    for port in "${PORTS[@]}"; do
        if ! force_cleanup_port "$port"; then
            failed_ports+=("$port")
        fi
    done
    
    # 清理PID文件
    cleanup_pid_files
    
    # 显示结果
    log_info ""
    if [ ${#failed_ports[@]} -eq 0 ]; then
        log_success "✅ 所有端口清理完成"
    else
        log_warning "⚠️ 部分端口清理失败: ${failed_ports[*]}"
        log_info "您可能需要手动检查这些端口"
    fi
    
    log_info ""
    log_info "💡 现在可以安全启动服务："
    log_info "  ./startup/quick-start.sh demo --auto-open"
}

# 执行主函数
main "$@"