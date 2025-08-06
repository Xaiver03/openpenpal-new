#!/bin/bash

# OpenPenPal 服务模板 - 统一的服务启动模板
# 使用方法：
# 1. 复制此模板到具体服务目录
# 2. 设置 SERVICE_CONFIG 变量
# 3. 调用 start_service 函数

set -euo pipefail

# 获取脚本目录和项目根目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# 加载服务框架
source "$PROJECT_ROOT/scripts/common/service-framework.sh"

# ==============================================================================
# 服务配置模板 - 请根据具体服务修改
# ==============================================================================

# 默认服务配置 - 需要在具体服务中覆盖 (兼容老版本bash)
_init_default_service_config() {
    # 基本信息
    set_service_config "name" "sample-service"
    set_service_config "description" "示例服务"
    set_service_config "version" "1.0.0"
    
    # 启动配置
    set_service_config "script" "npm start"
    set_service_config "port" "3000"
    set_service_config "env_file" ".env"
    set_service_config "work_dir" "."
    
    # 健康检查
    set_service_config "health_path" "/health"
    set_service_config "health_timeout" "10"
    
    # 超时设置
    set_service_config "startup_timeout" "60"
    set_service_config "shutdown_timeout" "30"
    
    # 依赖管理
    set_service_config "package_file" "package.json"
    set_service_config "auto_install_deps" "true"
    set_service_config "node_version" "16"
    
    # 日志配置
    set_service_config "log_level" "info"
    set_service_config "log_file" ""
}

# ==============================================================================
# 配置管理
# ==============================================================================

# 初始化服务配置
init_service_config() {
    local config_file="${1:-service.json}"
    
    log_debug "初始化服务配置"
    
    # 加载默认配置
    _init_default_service_config
    
    # 加载配置文件 (如果存在)
    if [ -f "$config_file" ]; then
        load_config_file "$config_file"
    fi
    
    # 环境变量覆盖
    load_env_overrides
    
    log_debug "服务配置初始化完成"
}

# 加载JSON配置文件
load_config_file() {
    local config_file="$1"
    
    if ! command_exists jq; then
        log_warn "jq 未安装，跳过配置文件加载: $config_file"
        return 0
    fi
    
    log_info "加载配置文件: $config_file"
    
    # 读取JSON配置并设置
    while IFS='=' read -r key value; do
        if [ -n "$key" ] && [ -n "$value" ]; then
            set_service_config "$key" "$value"
            log_debug "配置加载: $key=$value"
        fi
    done < <(jq -r 'to_entries | .[] | "\(.key)=\(.value)"' "$config_file" 2>/dev/null || true)
}

# 加载环境变量覆盖
load_env_overrides() {
    local service_name
    service_name=$(get_service_config "name" | tr '[:lower:]' '[:upper:]' | tr '-' '_')
    
    # 定义所有可能的配置键 (替代关联数组键的迭代)
    local config_keys="name description version script port env_file work_dir health_path health_timeout startup_timeout shutdown_timeout package_file auto_install_deps node_version log_level log_file"
    
    # 支持 SERVICE_NAME_KEY 格式的环境变量覆盖
    for key in $config_keys; do
        local env_key="${service_name}_${key^^}"
        env_key=$(echo "$env_key" | tr '[:lower:]' '[:upper:]')
        
        if [ -n "${!env_key:-}" ]; then
            set_service_config "$key" "${!env_key}"
            log_debug "环境变量覆盖: $key=${!env_key}"
        fi
    done
}

# ==============================================================================
# 服务生命周期管理
# ==============================================================================

# 预启动检查
pre_start_checks() {
    local service_name
    local node_version
    local work_dir
    
    service_name=$(get_service_config "name")
    node_version=$(get_service_config "node_version")
    work_dir=$(get_service_config "work_dir")
    
    log_step "1" "执行 $service_name 预启动检查"
    
    # 检查工作目录
    if [ ! -d "$work_dir" ]; then
        log_error "工作目录不存在: $work_dir"
        return 1
    fi
    
    cd "$work_dir"
    
    # 检查Node.js版本
    if ! check_node_version "$node_version"; then
        return 1
    fi
    
    # 检查npm环境
    if ! check_npm_environment; then
        return 1
    fi
    
    # 依赖检查和安装
    local package_file
    local auto_install
    package_file=$(get_service_config "package_file")
    auto_install=$(get_service_config "auto_install_deps")
    
    if [ "$auto_install" = "true" ]; then
        if ! ensure_dependencies "$package_file"; then
            return 1
        fi
    fi
    
    log_success "✓ 预启动检查完成"
    return 0
}

# 启动服务
start_service() {
    local service_name
    local port
    local script
    local env_file
    local startup_timeout
    local health_path
    local log_level
    
    service_name=$(get_service_config "name")
    port=$(get_service_config "port")
    script=$(get_service_config "script")
    env_file=$(get_service_config "env_file")
    startup_timeout=$(get_service_config "startup_timeout")
    health_path=$(get_service_config "health_path")
    log_level=$(get_service_config "log_level")
    
    log_step "2" "启动 $service_name 服务"
    
    # 设置日志级别
    set_log_level "$log_level"
    
    # 检查端口冲突
    if is_port_occupied "$port"; then
        log_warn "$service_name 端口 $port 已被占用"
        
        if confirm "是否停止现有进程？" "y"; then
            kill_port_gracefully "$port" "$service_name"
        else
            log_error "端口冲突，启动中止"
            return 1
        fi
    fi
    
    # 加载环境变量
    setup_environment "$env_file"
    
    # 启动服务进程
    log_info "执行启动命令: $script"
    
    local log_file
    log_file=$(get_service_config "log_file")
    
    if [ -n "$log_file" ]; then
        # 输出到日志文件
        nohup bash -c "$script" >"$log_file" 2>&1 &
    else
        # 静默启动
        nohup bash -c "$script" >/dev/null 2>&1 &
    fi
    
    local service_pid=$!
    
    # 等待服务启动
    log_info "等待服务启动..."
    if wait_for_port "localhost" "$port" "$startup_timeout" "$service_name"; then
        # 健康检查
        local health_url="http://localhost:${port}${health_path}"
        local health_timeout
        health_timeout=$(get_service_config "health_timeout")
        
        if check_service_health "$health_url" "$health_timeout" "$service_name"; then
            log_success "✓ $service_name 启动成功 (PID: $service_pid, 端口: $port)"
            
            # 显示服务信息
            show_service_info
            return 0
        else
            log_warn "$service_name 已启动，但健康检查失败"
            log_info "请手动验证服务状态: http://localhost:$port"
            return 0
        fi
    else
        log_error "✗ $service_name 启动失败或超时"
        return 1
    fi
}

# 停止服务
stop_service() {
    local service_name
    local port
    local shutdown_timeout
    
    service_name=$(get_service_config "name")
    port=$(get_service_config "port")
    shutdown_timeout=$(get_service_config "shutdown_timeout")
    
    log_step "3" "停止 $service_name 服务"
    
    if kill_port_gracefully "$port" "$service_name" "$shutdown_timeout"; then
        log_success "✓ $service_name 已停止"
        return 0
    else
        log_error "✗ $service_name 停止失败"
        return 1
    fi
}

# 重启服务
restart_service() {
    log_info "重启服务..."
    
    if stop_service; then
        sleep 2
        start_service
    else
        log_error "服务停止失败，重启中止"
        return 1
    fi
}

# ==============================================================================
# 信息显示
# ==============================================================================

# 显示服务信息
show_service_info() {
    local service_name
    local description
    local version
    local port
    local health_path
    
    service_name=$(get_service_config "name")
    description=$(get_service_config "description")
    version=$(get_service_config "version")
    port=$(get_service_config "port")
    health_path=$(get_service_config "health_path")
    
    echo ""
    echo "======================================"
    log_success "🎉 $service_name 服务启动完成"
    echo "======================================"
    echo ""
    echo "服务信息:"
    echo "  名称: $service_name"
    echo "  描述: $description"
    echo "  版本: $version"
    echo "  端口: $port"
    echo ""
    echo "访问地址:"
    echo "  主页: http://localhost:$port"
    echo "  健康检查: http://localhost:$port$health_path"
    echo ""
    echo "管理命令:"
    echo "  停止服务: $0 stop"
    echo "  重启服务: $0 restart"
    echo "  查看状态: $0 status"
    echo ""
    echo "======================================"
}

# 显示服务状态
show_service_status() {
    local service_name
    local port
    local health_path
    
    service_name=$(get_service_config "name")
    port=$(get_service_config "port")
    health_path=$(get_service_config "health_path")
    
    echo "======================================"
    echo "$service_name 服务状态"
    echo "======================================"
    
    # 检查端口状态
    if is_port_occupied "$port"; then
        local pid
        pid=$(get_port_pid "$port")
        log_success "✓ 服务运行中 (PID: $pid, 端口: $port)"
        
        # 健康检查
        local health_url="http://localhost:${port}${health_path}"
        if check_service_health "$health_url" 5 "$service_name"; then
            log_success "✓ 健康检查通过"
        else
            log_warn "⚠ 健康检查失败"
        fi
    else
        log_error "✗ 服务未运行"
    fi
    
    echo "======================================"
}

# ==============================================================================
# 命令行接口
# ==============================================================================

# 显示帮助信息
show_help() {
    local service_name
    service_name=$(get_service_config "name" "service")
    
    cat << EOF
$service_name 服务管理脚本

使用方法:
  $0 [命令] [选项]

命令:
  start     启动服务
  stop      停止服务  
  restart   重启服务
  status    查看服务状态
  config    显示配置信息
  help      显示帮助信息

选项:
  --config FILE     指定配置文件
  --log-level LEVEL 设置日志级别 (debug|info|warn|error)
  --port PORT       指定端口
  --env FILE        指定环境变量文件

示例:
  $0 start                    # 启动服务
  $0 start --port 3001        # 指定端口启动
  $0 stop                     # 停止服务
  $0 status                   # 查看状态
  $0 --config custom.json start  # 使用自定义配置启动

EOF
}

# 显示配置信息
show_config() {
    echo "======================================"
    echo "服务配置信息"
    echo "======================================"
    
    # 定义所有配置键 (替代关联数组键的迭代)
    local config_keys="name description version script port env_file work_dir health_path health_timeout startup_timeout shutdown_timeout package_file auto_install_deps node_version log_level log_file"
    
    for key in $config_keys; do
        local value
        value=$(get_service_config "$key")
        printf "  %-20s: %s\n" "$key" "$value"
    done
    
    echo "======================================"
}

# 主函数
main() {
    local command="start"
    local config_file=""
    
    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            start|stop|restart|status|config|help)
                command="$1"
                shift
                ;;
            --config)
                config_file="$2"
                shift 2
                ;;
            --log-level)
                set_log_level "$2"
                shift 2
                ;;
            --port)
                set_service_config "port" "$2"
                shift 2
                ;;
            --env)
                set_service_config "env_file" "$2"
                shift 2
                ;;
            --help|-h)
                command="help"
                shift
                ;;
            *)
                log_error "未知选项: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    # 初始化框架和配置
    init_framework "$config_file"
    init_service_config "$config_file"
    
    # 执行命令
    case $command in
        start)
            if pre_start_checks && start_service; then
                exit 0
            else
                exit 1
            fi
            ;;
        stop)
            if stop_service; then
                exit 0
            else
                exit 1
            fi
            ;;
        restart)
            if restart_service; then
                exit 0
            else
                exit 1
            fi
            ;;
        status)
            show_service_status
            ;;
        config)
            show_config
            ;;
        help)
            show_help
            ;;
        *)
            log_error "未知命令: $command"
            show_help
            exit 1
            ;;
    esac
}

# 如果直接执行此脚本，调用主函数
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    main "$@"
fi