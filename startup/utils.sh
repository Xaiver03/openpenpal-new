#!/bin/bash

# OpenPenPal 启动脚本工具函数
# 提供日志、端口检查、版本比较等通用功能

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 日志级别
LOG_LEVEL=${LOG_LEVEL:-"info"}

# 日志函数
log_debug() {
    if [ "$LOG_LEVEL" = "debug" ] || [ "$DEBUG" = "true" ]; then
        echo -e "${PURPLE}[DEBUG]${NC} [$(date '+%H:%M:%S')] $1" >&2
    fi
}

log_info() {
    if [ "$LOG_LEVEL" != "error" ] && [ "$LOG_LEVEL" != "warn" ]; then
        echo -e "${BLUE}[INFO]${NC} [$(date '+%H:%M:%S')] $1"
    fi
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} [$(date '+%H:%M:%S')] $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} [$(date '+%H:%M:%S')] $1" >&2
}

log_error() {
    echo -e "${RED}[ERROR]${NC} [$(date '+%H:%M:%S')] $1" >&2
}

log_step() {
    echo -e "${CYAN}[STEP]${NC} [$(date '+%H:%M:%S')] $1"
}

# 运行命令（支持dry-run模式）
run_command() {
    local cmd="$1"
    local description="$2"
    
    if [ -n "$description" ]; then
        log_debug "执行: $description"
    fi
    
    if [ "$DRY_RUN" = true ]; then
        log_info "[DRY RUN] $cmd"
        return 0
    fi
    
    log_debug "命令: $cmd"
    
    if [ "$VERBOSE" = true ]; then
        eval "$cmd"
    else
        eval "$cmd" >/dev/null 2>&1
    fi
    
    local exit_code=$?
    if [ $exit_code -ne 0 ]; then
        log_error "命令执行失败: $cmd (退出码: $exit_code)"
        return $exit_code
    fi
    
    return 0
}

# 检查端口是否可用（未被占用）
check_port_available() {
    local port="$1"
    
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
        return 1  # 端口被占用
    else
        return 0  # 端口可用
    fi
}

# 检查端口是否被占用
check_port_occupied() {
    local port="$1"
    
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
        return 0  # 端口被占用
    else
        return 1  # 端口未被占用
    fi
}

# 等待端口可用
wait_for_port() {
    local port="$1"
    local timeout="${2:-30}"
    local interval="${3:-1}"
    
    local count=0
    while [ $count -lt $timeout ]; do
        if check_port_occupied $port; then
            return 0
        fi
        sleep $interval
        count=$((count + interval))
    done
    
    return 1
}

# 等待服务启动
wait_for_service() {
    local port="$1"
    local service_name="$2"
    local timeout="${3:-30}"
    
    log_debug "等待 $service_name 在端口 $port 启动..."
    
    if wait_for_port $port $timeout; then
        # 额外验证HTTP服务
        if curl -s "http://localhost:$port" >/dev/null 2>&1 || 
           curl -s "http://localhost:$port/health" >/dev/null 2>&1; then
            return 0
        fi
        
        # 如果HTTP检查失败，再等待一下
        sleep 2
        if curl -s "http://localhost:$port" >/dev/null 2>&1 || 
           curl -s "http://localhost:$port/health" >/dev/null 2>&1; then
            return 0
        fi
    fi
    
    return 1
}

# 停止端口上的进程
kill_port() {
    local port="$1"
    local signal="${2:-TERM}"
    
    local pids=$(lsof -ti :$port 2>/dev/null)
    
    if [ -n "$pids" ]; then
        log_debug "停止端口 $port 上的进程: $pids"
        
        for pid in $pids; do
            if [ "$signal" = "KILL" ]; then
                kill -9 $pid 2>/dev/null
            else
                kill -TERM $pid 2>/dev/null
            fi
        done
        
        # 等待进程退出
        sleep 2
        
        # 如果还有进程，强制杀死
        local remaining_pids=$(lsof -ti :$port 2>/dev/null)
        if [ -n "$remaining_pids" ]; then
            for pid in $remaining_pids; do
                kill -9 $pid 2>/dev/null
            done
        fi
        
        return 0
    else
        log_debug "端口 $port 没有进程运行"
        return 1
    fi
}

# 版本比较函数
version_gte() {
    printf '%s\n%s\n' "$2" "$1" | sort -V -C
}

version_lt() {
    ! version_gte "$1" "$2"
}

# 检查命令是否存在
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# 检查文件是否存在且可执行
is_executable() {
    [ -f "$1" ] && [ -x "$1" ]
}

# 获取进程PID（通过端口）
get_pid_by_port() {
    local port="$1"
    lsof -ti :$port 2>/dev/null | head -1
}

# 获取进程PID（通过PID文件）
get_pid_by_file() {
    local pid_file="$1"
    
    if [ -f "$pid_file" ]; then
        local pid=$(cat "$pid_file" 2>/dev/null)
        if [ -n "$pid" ] && ps -p $pid >/dev/null 2>&1; then
            echo "$pid"
            return 0
        else
            # PID文件存在但进程不存在，清理文件
            rm -f "$pid_file"
        fi
    fi
    
    return 1
}

# 检查服务是否运行
is_service_running() {
    local service_name="$1"
    local port="$2"
    local pid_file="$3"
    
    # 优先检查PID文件
    if [ -n "$pid_file" ]; then
        local pid=$(get_pid_by_file "$pid_file")
        if [ -n "$pid" ]; then
            return 0
        fi
    fi
    
    # 检查端口
    if [ -n "$port" ]; then
        if check_port_occupied $port; then
            return 0
        fi
    fi
    
    return 1
}

# 格式化时间（秒转换为可读格式）
format_duration() {
    local seconds="$1"
    
    if [ $seconds -lt 60 ]; then
        echo "${seconds}秒"
    elif [ $seconds -lt 3600 ]; then
        echo "$((seconds / 60))分$((seconds % 60))秒"
    else
        echo "$((seconds / 3600))小时$(((seconds % 3600) / 60))分$((seconds % 60))秒"
    fi
}

# 获取系统信息
get_system_info() {
    local os_name=""
    local os_version=""
    
    if [ "$(uname)" = "Darwin" ]; then
        os_name="macOS"
        os_version=$(sw_vers -productVersion 2>/dev/null)
    elif [ -f /etc/os-release ]; then
        os_name=$(grep '^NAME=' /etc/os-release | cut -d'"' -f2)
        os_version=$(grep '^VERSION=' /etc/os-release | cut -d'"' -f2)
    else
        os_name=$(uname -s)
        os_version=$(uname -r)
    fi
    
    echo "$os_name $os_version"
}

# 获取内存使用情况
get_memory_usage() {
    if [ "$(uname)" = "Darwin" ]; then
        # macOS
        local total_memory=$(sysctl -n hw.memsize)
        local total_gb=$((total_memory / 1024 / 1024 / 1024))
        echo "${total_gb}GB"
    elif [ -f /proc/meminfo ]; then
        # Linux
        local total_kb=$(grep '^MemTotal:' /proc/meminfo | awk '{print $2}')
        local total_gb=$((total_kb / 1024 / 1024))
        echo "${total_gb}GB"
    else
        echo "Unknown"
    fi
}

# 检查磁盘空间
check_disk_space() {
    local path="${1:-$PWD}"
    local required_gb="${2:-1}"
    
    local available_kb=$(df "$path" | awk 'NR==2 {print $4}')
    local available_gb=$((available_kb / 1024 / 1024))
    
    if [ $available_gb -lt $required_gb ]; then
        log_warning "磁盘空间不足。可用: ${available_gb}GB，需要: ${required_gb}GB"
        return 1
    fi
    
    return 0
}

# 创建备份
create_backup() {
    local source="$1"
    local backup_dir="${2:-./backups}"
    local timestamp=$(date +%Y%m%d_%H%M%S)
    
    if [ ! -e "$source" ]; then
        log_error "备份源不存在: $source"
        return 1
    fi
    
    mkdir -p "$backup_dir"
    
    local backup_name="$(basename "$source")_$timestamp"
    local backup_path="$backup_dir/$backup_name"
    
    if [ -d "$source" ]; then
        cp -r "$source" "$backup_path"
    else
        cp "$source" "$backup_path"
    fi
    
    log_success "备份创建成功: $backup_path"
    echo "$backup_path"
}

# 显示进度条
show_progress() {
    local current="$1"
    local total="$2"
    local width="${3:-50}"
    local title="${4:-Progress}"
    
    local percentage=$((current * 100 / total))
    local completed=$((current * width / total))
    local remaining=$((width - completed))
    
    local bar=""
    for ((i=0; i<completed; i++)); do
        bar+="█"
    done
    for ((i=0; i<remaining; i++)); do
        bar+="░"
    done
    
    printf "\r${BLUE}%s:${NC} [%s] %d%% (%d/%d)" "$title" "$bar" "$percentage" "$current" "$total"
    
    if [ $current -eq $total ]; then
        echo ""
    fi
}

# 确认操作
confirm() {
    local message="$1"
    local default="${2:-n}"
    
    local prompt="$message"
    if [ "$default" = "y" ]; then
        prompt="$prompt [Y/n]: "
    else
        prompt="$prompt [y/N]: "
    fi
    
    read -p "$prompt" response
    
    if [ -z "$response" ]; then
        response="$default"
    fi
    
    case "$response" in
        [yY]|[yY][eE][sS])
            return 0
            ;;
        *)
            return 1
            ;;
    esac
}

# 生成随机字符串
generate_random_string() {
    local length="${1:-32}"
    local chars="ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
    
    if command_exists openssl; then
        openssl rand -base64 $length | tr -d '\n'
    elif command_exists head && [ -c /dev/urandom ]; then
        head -c $length /dev/urandom | base64 | tr -d '\n'
    else
        # 备用方法
        local result=""
        for ((i=0; i<length; i++)); do
            result+="${chars:$((RANDOM % ${#chars})):1}"
        done
        echo "$result"
    fi
}

# 清理函数
cleanup_on_exit() {
    log_debug "执行清理操作..."
    
    # 如果有自定义清理函数，执行它
    if declare -f custom_cleanup >/dev/null; then
        custom_cleanup
    fi
}

# 设置退出处理
setup_exit_handler() {
    trap cleanup_on_exit EXIT INT TERM
}

# 检查网络连接
check_internet() {
    local test_hosts=("8.8.8.8" "1.1.1.1" "114.114.114.114")
    
    for host in "${test_hosts[@]}"; do
        if ping -c 1 -W 3 "$host" >/dev/null 2>&1; then
            return 0
        fi
    done
    
    return 1
}

# 导出所有函数
export -f log_debug log_info log_success log_warning log_error log_step
export -f run_command check_port_available check_port_occupied wait_for_port wait_for_service
export -f kill_port version_gte version_lt command_exists is_executable
export -f get_pid_by_port get_pid_by_file is_service_running format_duration
export -f get_system_info get_memory_usage check_disk_space create_backup
export -f show_progress confirm generate_random_string
export -f cleanup_on_exit setup_exit_handler check_internet