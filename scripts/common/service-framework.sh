#!/bin/bash

# OpenPenPal 服务框架 - SOTA级别的统一启动脚本框架
# 设计原则：
# 1. 单一职责：每个函数只负责一个特定功能
# 2. 可配置性：通过配置文件驱动行为
# 3. 可扩展性：支持插件式功能扩展
# 4. 错误处理：完善的错误检测和恢复机制
# 5. 日志记录：结构化日志输出

set -euo pipefail

# ==============================================================================
# 全局变量和配置
# ==============================================================================

# 框架版本
readonly FRAMEWORK_VERSION="1.0.0"

# 颜色常量 - 统一的终端颜色方案
readonly COLOR_RED='\033[0;31m'
readonly COLOR_GREEN='\033[0;32m'
readonly COLOR_YELLOW='\033[1;33m'
readonly COLOR_BLUE='\033[0;34m'
readonly COLOR_PURPLE='\033[0;35m'
readonly COLOR_CYAN='\033[0;36m'
readonly COLOR_WHITE='\033[1;37m'
readonly COLOR_NC='\033[0m'

# 日志级别
readonly LOG_LEVEL_DEBUG=0
readonly LOG_LEVEL_INFO=1
readonly LOG_LEVEL_WARN=2
readonly LOG_LEVEL_ERROR=3

# 默认配置
DEFAULT_LOG_LEVEL=${LOG_LEVEL_INFO}
DEFAULT_TIMEOUT=60
DEFAULT_RETRY_COUNT=3
DEFAULT_RETRY_DELAY=2

# 当前日志级别
CURRENT_LOG_LEVEL=${DEFAULT_LOG_LEVEL}

# ==============================================================================
# 日志系统 - 结构化日志输出
# ==============================================================================

# 设置日志级别
set_log_level() {
    local level="$1"
    case "$level" in
        "debug"|"DEBUG") CURRENT_LOG_LEVEL=${LOG_LEVEL_DEBUG} ;;
        "info"|"INFO") CURRENT_LOG_LEVEL=${LOG_LEVEL_INFO} ;;
        "warn"|"WARN") CURRENT_LOG_LEVEL=${LOG_LEVEL_WARN} ;;
        "error"|"ERROR") CURRENT_LOG_LEVEL=${LOG_LEVEL_ERROR} ;;
        *) 
            log_error "无效的日志级别: $level"
            return 1
            ;;
    esac
}

# 通用日志函数
_log() {
    local level="$1"
    local level_num="$2"
    local color="$3"
    local message="$4"
    
    if [ "$level_num" -ge "$CURRENT_LOG_LEVEL" ]; then
        local timestamp
        timestamp=$(date '+%H:%M:%S')
        printf "${color}[%s]${COLOR_NC} [%s] %s\n" "$level" "$timestamp" "$message" >&2
    fi
}

# 日志函数
log_debug() { _log "DEBUG" "$LOG_LEVEL_DEBUG" "$COLOR_PURPLE" "$1"; }
log_info() { _log "INFO" "$LOG_LEVEL_INFO" "$COLOR_BLUE" "$1"; }
log_warn() { _log "WARN" "$LOG_LEVEL_WARN" "$COLOR_YELLOW" "$1"; }
log_error() { _log "ERROR" "$LOG_LEVEL_ERROR" "$COLOR_RED" "$1"; }
log_success() { _log "SUCCESS" "$LOG_LEVEL_INFO" "$COLOR_GREEN" "$1"; }

# 带时间戳的详细日志
log_step() {
    local step="$1"
    local message="$2"
    printf "${COLOR_CYAN}[STEP %s]${COLOR_NC} [%s] %s\n" "$step" "$(date '+%H:%M:%S')" "$message" >&2
}

# ==============================================================================
# 系统检测和验证
# ==============================================================================

# 检查命令是否存在
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# 检查端口是否被占用
is_port_occupied() {
    local port="$1"
    if command_exists lsof; then
        lsof -ti :"$port" >/dev/null 2>&1
    elif command_exists netstat; then
        netstat -tuln 2>/dev/null | grep -q ":${port} "
    elif command_exists ss; then
        ss -tuln 2>/dev/null | grep -q ":${port} "
    else
        log_warn "无法检查端口状态：缺少 lsof, netstat 或 ss 命令"
        return 1
    fi
}

# 等待端口变为可用
wait_for_port() {
    local host="${1:-localhost}"
    local port="$2"
    local timeout="${3:-$DEFAULT_TIMEOUT}"
    local service_name="${4:-service}"
    
    log_info "等待 $service_name 在 $host:$port 启动..."
    
    local elapsed=0
    local interval=2
    
    while [ $elapsed -lt "$timeout" ]; do
        if is_port_occupied "$port"; then
            log_success "✓ $service_name 已在 $host:$port 启动"
            return 0
        fi
        
        sleep $interval
        elapsed=$((elapsed + interval))
        
        if [ $((elapsed % 10)) -eq 0 ]; then
            log_debug "等待中... (${elapsed}/${timeout}s)"
        fi
    done
    
    log_error "✗ $service_name 启动超时 (${timeout}s)"
    return 1
}

# 检查服务健康状态
check_service_health() {
    local url="$1"
    local timeout="${2:-10}"
    local service_name="${3:-service}"
    
    log_debug "检查 $service_name 健康状态: $url"
    
    if command_exists curl; then
        if curl -sf --max-time "$timeout" "$url" >/dev/null 2>&1; then
            log_debug "✓ $service_name 健康检查通过"
            return 0
        fi
    elif command_exists wget; then
        if wget --timeout="$timeout" --tries=1 -q -O /dev/null "$url" 2>/dev/null; then
            log_debug "✓ $service_name 健康检查通过"
            return 0
        fi
    else
        log_warn "无法执行健康检查：缺少 curl 或 wget"
        return 1
    fi
    
    log_debug "✗ $service_name 健康检查失败"
    return 1
}

# ==============================================================================
# 进程管理
# ==============================================================================

# 获取占用端口的进程ID
get_port_pid() {
    local port="$1"
    if command_exists lsof; then
        lsof -ti :"$port" 2>/dev/null | head -1
    else
        log_warn "无法获取端口PID：缺少 lsof 命令"
        return 1
    fi
}

# 优雅停止端口上的进程
kill_port_gracefully() {
    local port="$1"
    local service_name="${2:-service}"
    local timeout="${3:-10}"
    
    local pid
    pid=$(get_port_pid "$port")
    
    if [ -z "$pid" ]; then
        log_debug "$service_name (端口 $port) 未运行"
        return 0
    fi
    
    log_info "优雅停止 $service_name (PID: $pid, 端口: $port)"
    
    # 发送TERM信号
    if kill -TERM "$pid" 2>/dev/null; then
        # 等待进程退出
        local elapsed=0
        local interval=1
        
        while [ $elapsed -lt "$timeout" ]; do
            if ! kill -0 "$pid" 2>/dev/null; then
                log_success "✓ $service_name 已优雅停止"
                return 0
            fi
            sleep $interval
            elapsed=$((elapsed + interval))
        done
        
        # 强制终止
        log_warn "$service_name 未在超时时间内停止，强制终止"
        if kill -KILL "$pid" 2>/dev/null; then
            log_success "✓ $service_name 已强制停止"
            return 0
        fi
    fi
    
    log_error "✗ 无法停止 $service_name"
    return 1
}

# 批量停止多个端口的服务
kill_ports() {
    local ports=("$@")
    local failed_ports=()
    
    for port in "${ports[@]}"; do
        if ! kill_port_gracefully "$port" "service-${port}"; then
            failed_ports+=("$port")
        fi
    done
    
    if [ ${#failed_ports[@]} -gt 0 ]; then
        log_error "以下端口的服务停止失败: ${failed_ports[*]}"
        return 1
    fi
    
    return 0
}

# ==============================================================================
# 环境检测和设置
# ==============================================================================

# 检测Node.js版本
check_node_version() {
    local required_version="${1:-16}"
    
    if ! command_exists node; then
        log_error "Node.js 未安装"
        return 1
    fi
    
    local current_version
    current_version=$(node --version | sed 's/v//' | cut -d. -f1)
    
    if [ "$current_version" -lt "$required_version" ]; then
        log_error "Node.js 版本过低：当前 v${current_version}.x，要求 v${required_version}.x+"
        return 1
    fi
    
    log_success "✓ Node.js v${current_version}.x (满足要求)"
    return 0
}

# 检测npm版本和配置
check_npm_environment() {
    if ! command_exists npm; then
        log_error "npm 未安装"
        return 1
    fi
    
    local npm_version
    npm_version=$(npm --version)
    log_success "✓ npm v${npm_version}"
    
    # 检查npm配置
    local registry
    registry=$(npm config get registry 2>/dev/null || echo "")
    if [ -n "$registry" ]; then
        log_debug "npm registry: $registry"
    fi
    
    return 0
}

# 设置环境变量
setup_environment() {
    local env_file="${1:-.env}"
    
    if [ -f "$env_file" ]; then
        log_info "加载环境变量: $env_file"
        # 安全地加载环境变量
        while IFS= read -r line; do
            # 跳过注释和空行
            if [[ "$line" =~ ^[[:space:]]*# ]] || [[ -z "$line" ]]; then
                continue
            fi
            
            # 验证格式
            if [[ "$line" =~ ^[A-Za-z_][A-Za-z0-9_]*=.*$ ]]; then
                export "$line"
                log_debug "设置环境变量: ${line%%=*}"
            else
                log_warn "跳过无效的环境变量行: $line"
            fi
        done < "$env_file"
    else
        log_debug "环境变量文件不存在: $env_file"
    fi
}

# ==============================================================================
# 依赖管理
# ==============================================================================

# 检查并安装依赖
ensure_dependencies() {
    local package_file="${1:-package.json}"
    local force_install="${2:-false}"
    
    if [ ! -f "$package_file" ]; then
        log_warn "package.json 不存在: $package_file"
        return 0
    fi
    
    local package_dir
    package_dir=$(dirname "$package_file")
    local node_modules_dir="$package_dir/node_modules"
    
    # 检查是否需要安装依赖
    if [ "$force_install" = "true" ] || [ ! -d "$node_modules_dir" ]; then
        log_info "安装依赖: $package_file"
        
        if ! (cd "$package_dir" && npm install); then
            log_error "依赖安装失败: $package_file"
            return 1
        fi
        
        log_success "✓ 依赖安装完成: $package_file"
    else
        log_debug "依赖已存在: $node_modules_dir"
    fi
    
    return 0
}

# 清理依赖
clean_dependencies() {
    local package_file="${1:-package.json}"
    
    local package_dir
    package_dir=$(dirname "$package_file")
    local node_modules_dir="$package_dir/node_modules"
    
    if [ -d "$node_modules_dir" ]; then
        log_info "清理依赖: $node_modules_dir"
        rm -rf "$node_modules_dir"
        log_success "✓ 依赖清理完成"
    else
        log_debug "依赖目录不存在: $node_modules_dir"
    fi
}

# ==============================================================================
# 服务启动器
# ==============================================================================

# 服务配置存储文件 (替代关联数组以支持老版本bash)
SERVICE_CONFIG_FILE="/tmp/openpenpal-service-config-$$"

# 初始化配置存储
_init_service_config_storage() {
    if [ ! -f "$SERVICE_CONFIG_FILE" ]; then
        touch "$SERVICE_CONFIG_FILE"
    fi
}

# 设置服务配置
set_service_config() {
    local key="$1"
    local value="$2"
    
    _init_service_config_storage
    
    # 删除现有的key（如果存在）
    if [ -f "$SERVICE_CONFIG_FILE" ]; then
        grep -v "^${key}=" "$SERVICE_CONFIG_FILE" > "${SERVICE_CONFIG_FILE}.tmp" 2>/dev/null || true
        mv "${SERVICE_CONFIG_FILE}.tmp" "$SERVICE_CONFIG_FILE"
    fi
    
    # 添加新的键值对
    echo "${key}=${value}" >> "$SERVICE_CONFIG_FILE"
}

# 获取服务配置
get_service_config() {
    local key="$1"
    local default="${2:-}"
    
    _init_service_config_storage
    
    # 从文件中查找配置值
    if [ -f "$SERVICE_CONFIG_FILE" ]; then
        local value
        value=$(grep "^${key}=" "$SERVICE_CONFIG_FILE" 2>/dev/null | cut -d'=' -f2- | tail -1)
        if [ -n "$value" ]; then
            echo "$value"
        else
            echo "$default"
        fi
    else
        echo "$default"
    fi
}

# 启动Node.js服务
start_node_service() {
    local config_prefix="$1"
    
    local name
    local script
    local port
    local env_file
    local health_path
    local timeout
    
    name=$(get_service_config "${config_prefix}.name" "service")
    script=$(get_service_config "${config_prefix}.script" "")
    port=$(get_service_config "${config_prefix}.port" "")
    env_file=$(get_service_config "${config_prefix}.env_file" ".env")
    health_path=$(get_service_config "${config_prefix}.health_path" "/health")
    timeout=$(get_service_config "${config_prefix}.timeout" "$DEFAULT_TIMEOUT")
    
    if [ -z "$script" ] || [ -z "$port" ]; then
        log_error "服务配置不完整: $config_prefix (需要 script 和 port)"
        return 1
    fi
    
    log_step "1" "启动 $name 服务"
    
    # 检查端口冲突
    if is_port_occupied "$port"; then
        log_warn "$name 端口 $port 已被占用，尝试停止现有进程"
        kill_port_gracefully "$port" "$name"
    fi
    
    # 设置环境变量
    setup_environment "$env_file"
    
    # 启动服务
    log_info "执行启动脚本: $script"
    nohup bash -c "$script" >/dev/null 2>&1 &
    local service_pid=$!
    
    # 等待服务启动
    if wait_for_port "localhost" "$port" "$timeout" "$name"; then
        # 健康检查
        local health_url="http://localhost:${port}${health_path}"
        if check_service_health "$health_url" 10 "$name"; then
            log_success "✓ $name 服务启动成功 (PID: $service_pid, 端口: $port)"
            return 0
        else
            log_warn "$name 服务已启动，但健康检查失败"
            return 0  # 仍然认为启动成功，可能健康检查接口不存在
        fi
    else
        log_error "✗ $name 服务启动失败"
        return 1
    fi
}

# ==============================================================================
# 实用工具函数
# ==============================================================================

# 显示进度条
show_progress() {
    local message="$1"
    local duration="${2:-5}"
    
    local progress=0
    local total=50
    
    while [ $progress -le $total ]; do
        local percent=$((progress * 100 / total))
        local filled=$((progress * 2))
        local empty=$((100 - filled))
        
        printf "\r${COLOR_BLUE}%s${COLOR_NC} [" "$message"
        printf "%${filled}s" | tr ' ' '='
        printf "%${empty}s" | tr ' ' '-'
        printf "] %d%%" "$percent"
        
        if [ $progress -eq $total ]; then
            printf "\n"
            break
        fi
        
        sleep "$((duration / total))"
        progress=$((progress + 1))
    done
}

# 确认提示
confirm() {
    local message="$1"
    local default="${2:-n}"
    
    while true; do
        if [ "$default" = "y" ]; then
            printf "%s [Y/n]: " "$message"
        else
            printf "%s [y/N]: " "$message"
        fi
        
        read -r response
        response=${response:-$default}
        
        case "$response" in
            [yY]|[yY][eE][sS]) return 0 ;;
            [nN]|[nN][oO]) return 1 ;;
            *) echo "请输入 y 或 n" ;;
        esac
    done
}

# 重试执行函数
retry_command() {
    local max_attempts="${1:-$DEFAULT_RETRY_COUNT}"
    local delay="${2:-$DEFAULT_RETRY_DELAY}"
    local command_name="$3"
    shift 3
    
    local attempt=1
    
    while [ $attempt -le "$max_attempts" ]; do
        if "$@"; then
            if [ $attempt -gt 1 ]; then
                log_success "✓ $command_name 重试成功 (第 $attempt 次尝试)"
            fi
            return 0
        fi
        
        if [ $attempt -lt "$max_attempts" ]; then
            log_warn "$command_name 失败，${delay}秒后重试 (第 $attempt/$max_attempts 次)"
            sleep "$delay"
        fi
        
        attempt=$((attempt + 1))
    done
    
    log_error "✗ $command_name 重试 $max_attempts 次后仍然失败"
    return 1
}

# ==============================================================================
# 框架初始化和清理
# ==============================================================================

# 初始化框架
init_framework() {
    local config_file="${1:-}"
    
    log_info "初始化 OpenPenPal 服务框架 v${FRAMEWORK_VERSION}"
    
    # 加载配置文件
    if [ -n "$config_file" ] && [ -f "$config_file" ]; then
        log_info "加载配置文件: $config_file"
        # 这里可以扩展为加载JSON或YAML配置
    fi
    
    # 设置信号处理
    trap cleanup_framework INT TERM
    
    log_success "✓ 框架初始化完成"
}

# 清理框架资源
cleanup_framework() {
    log_info "清理框架资源..."
    
    # 清理临时配置文件
    if [ -f "$SERVICE_CONFIG_FILE" ]; then
        rm -f "$SERVICE_CONFIG_FILE" "${SERVICE_CONFIG_FILE}.tmp" 2>/dev/null || true
    fi
    
    # 这里可以添加其他清理逻辑
    # 例如：停止后台进程等
    
    log_info "框架已退出"
    exit 0
}

# ==============================================================================
# 导出函数
# ==============================================================================

# 导出所有公共函数
export -f set_log_level log_debug log_info log_warn log_error log_success log_step
export -f command_exists is_port_occupied wait_for_port check_service_health
export -f get_port_pid kill_port_gracefully kill_ports
export -f check_node_version check_npm_environment setup_environment
export -f ensure_dependencies clean_dependencies
export -f set_service_config get_service_config start_node_service
export -f show_progress confirm retry_command
export -f init_framework cleanup_framework

# 导出常量
export FRAMEWORK_VERSION
export COLOR_RED COLOR_GREEN COLOR_YELLOW COLOR_BLUE COLOR_PURPLE COLOR_CYAN COLOR_WHITE COLOR_NC
export LOG_LEVEL_DEBUG LOG_LEVEL_INFO LOG_LEVEL_WARN LOG_LEVEL_ERROR
export DEFAULT_LOG_LEVEL DEFAULT_TIMEOUT DEFAULT_RETRY_COUNT DEFAULT_RETRY_DELAY

log_debug "服务框架已加载 (v${FRAMEWORK_VERSION})"