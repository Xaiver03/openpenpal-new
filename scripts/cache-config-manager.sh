#!/bin/bash

# 权限缓存配置管理工具
# 使用: ./scripts/cache-config-manager.sh [action] [options]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 显示帮助信息
show_help() {
    cat << EOF
🔐 权限缓存配置管理工具

用法: $0 [action] [options]

Actions:
  show                          显示当前缓存配置
  check                         检查缓存健康状态
  update-security              应用安全缓存配置（推荐）
  update-performance           应用高性能缓存配置
  update-custom [config.json]  应用自定义配置文件
  validate                     验证配置文件格式
  backup                       备份当前配置
  restore [backup_file]        恢复配置

安全配置（推荐生产环境）:
  - 用户权限缓存: 5分钟
  - 角色权限缓存: 10分钟  
  - 菜单缓存: 30分钟
  - 会话缓存: 24小时

高性能配置（开发/测试环境）:
  - 用户权限缓存: 30分钟
  - 角色权限缓存: 1小时
  - 菜单缓存: 2小时
  - 会话缓存: 24小时

示例:
  $0 show                                    # 显示当前配置
  $0 check                                   # 检查缓存状态
  $0 update-security                         # 应用安全配置
  $0 update-custom custom-cache-config.json # 应用自定义配置
  $0 backup                                  # 备份当前配置

EOF
}

# 检查依赖
check_dependencies() {
    local missing_deps=()
    
    command -v jq >/dev/null 2>&1 || missing_deps+=("jq")
    command -v redis-cli >/dev/null 2>&1 || missing_deps+=("redis-cli")
    
    if [ ${#missing_deps[@]} -ne 0 ]; then
        log_error "Missing dependencies: ${missing_deps[*]}"
        log_info "Please install missing tools:"
        for dep in "${missing_deps[@]}"; do
            case $dep in
                jq) echo "  brew install jq (macOS) or apt-get install jq (Ubuntu)" ;;
                redis-cli) echo "  brew install redis (macOS) or apt-get install redis-tools (Ubuntu)" ;;
            esac
        done
        exit 1
    fi
}

# 检查Redis连接
check_redis() {
    local redis_host="${REDIS_HOST:-localhost}"
    local redis_port="${REDIS_PORT:-6379}"
    
    if ! redis-cli -h "$redis_host" -p "$redis_port" ping >/dev/null 2>&1; then
        log_warning "Redis is not accessible at $redis_host:$redis_port"
        log_info "Cache configuration will be applied when Redis is available"
        return 1
    fi
    
    return 0
}

# 显示当前缓存配置
show_config() {
    log_info "Current Cache Configuration:"
    echo
    
    # Go后端配置
    echo "🔧 Go Backend (Redis-based):"
    if check_redis; then
        redis-cli HGETALL openpenpal:cache:config 2>/dev/null | while read -r key; read -r value; do
            printf "  %-25s: %s\n" "$key" "$value"
        done || echo "  No configuration found in Redis"
    else
        echo "  Redis not available - using default values"
        echo "  user_permission_ttl     : 300 seconds (5 minutes)"
        echo "  role_permission_ttl     : 600 seconds (10 minutes)"
        echo "  menu_cache_ttl          : 1800 seconds (30 minutes)"
        echo "  session_ttl             : 86400 seconds (24 hours)"
    fi
    
    echo
    
    # Python服务配置
    echo "🐍 Python Service:"
    if [ -f "$PROJECT_ROOT/services/write-service/app/services/rbac_service.py" ]; then
        local user_perm=$(grep -o "user_permission_expire = [0-9]*" "$PROJECT_ROOT/services/write-service/app/services/rbac_service.py" | grep -o "[0-9]*")
        local role_perm=$(grep -o "role_permission_expire = [0-9]*" "$PROJECT_ROOT/services/write-service/app/services/rbac_service.py" | grep -o "[0-9]*")
        local menu_cache=$(grep -o "menu_cache_expire = [0-9]*" "$PROJECT_ROOT/services/write-service/app/services/rbac_service.py" | grep -o "[0-9]*")
        
        printf "  %-25s: %s seconds (%d minutes)\n" "user_permission_expire" "${user_perm:-300}" $((${user_perm:-300}/60))
        printf "  %-25s: %s seconds (%d minutes)\n" "role_permission_expire" "${role_perm:-600}" $((${role_perm:-600}/60))
        printf "  %-25s: %s seconds (%d minutes)\n" "menu_cache_expire" "${menu_cache:-1800}" $((${menu_cache:-1800}/60))
    else
        echo "  Configuration file not found"
    fi
}

# 检查缓存健康状态
check_cache_health() {
    log_info "Checking cache health status..."
    echo
    
    if ! check_redis; then
        log_error "Redis connection failed"
        return 1
    fi
    
    # Redis信息
    echo "📊 Redis Information:"
    redis-cli INFO memory | grep -E "(used_memory_human|used_memory_peak_human|maxmemory_human)" | while IFS=: read -r key value; do
        printf "  %-25s: %s\n" "$key" "$value"
    done
    
    echo
    
    # 缓存键统计
    echo "🗄️ Cache Key Statistics:"
    local patterns=("openpenpal:user:*:permissions" "openpenpal:user:*:menu:*" "openpenpal:session:*" "openpenpal:role:*")
    
    for pattern in "${patterns[@]}"; do
        local count=$(redis-cli KEYS "$pattern" 2>/dev/null | wc -l)
        local category=$(echo "$pattern" | sed 's/openpenpal://;s/:.*$//')
        printf "  %-15s: %d keys\n" "$category" "$count"
    done
    
    echo
    
    # TTL分析
    echo "⏰ TTL Analysis (sample keys):"
    for pattern in "${patterns[@]}"; do
        local sample_key=$(redis-cli KEYS "$pattern" 2>/dev/null | head -1)
        if [ -n "$sample_key" ]; then
            local ttl=$(redis-cli TTL "$sample_key" 2>/dev/null)
            local category=$(echo "$pattern" | sed 's/openpenpal://;s/:.*$//')
            if [ "$ttl" -gt 0 ]; then
                printf "  %-15s: %d seconds remaining\n" "$category" "$ttl"
            elif [ "$ttl" -eq -1 ]; then
                printf "  %-15s: no expiration set\n" "$category"
            fi
        fi
    done
}

# 应用安全配置
update_security_config() {
    log_info "Applying secure cache configuration..."
    
    local config='{
        "user_permission_ttl": 300,
        "role_permission_ttl": 600,
        "menu_cache_ttl": 1800,
        "session_ttl": 86400,
        "description": "Security-focused configuration with shorter TTLs"
    }'
    
    apply_config "$config"
    log_success "Security configuration applied successfully"
}

# 应用高性能配置
update_performance_config() {
    log_info "Applying high-performance cache configuration..."
    
    local config='{
        "user_permission_ttl": 1800,
        "role_permission_ttl": 3600,
        "menu_cache_ttl": 7200,
        "session_ttl": 86400,
        "description": "Performance-focused configuration with longer TTLs"
    }'
    
    apply_config "$config"
    log_success "Performance configuration applied successfully"
}

# 应用配置
apply_config() {
    local config="$1"
    
    # 更新Redis配置
    if check_redis; then
        echo "$config" | jq -r 'to_entries[] | "\(.key) \(.value)"' | while read -r key value; do
            if [ "$key" != "description" ]; then
                redis-cli HSET openpenpal:cache:config "$key" "$value" >/dev/null
            fi
        done
        
        # 设置配置过期时间（7天）
        redis-cli EXPIRE openpenpal:cache:config 604800 >/dev/null
        
        log_success "Redis configuration updated"
    fi
    
    # 更新Python服务配置
    local python_config_file="$PROJECT_ROOT/services/write-service/app/services/rbac_service.py"
    if [ -f "$python_config_file" ]; then
        local user_perm=$(echo "$config" | jq -r '.user_permission_ttl')
        local role_perm=$(echo "$config" | jq -r '.role_permission_ttl')
        local menu_cache=$(echo "$config" | jq -r '.menu_cache_ttl')
        
        # 备份原文件
        cp "$python_config_file" "$python_config_file.backup.$(date +%Y%m%d_%H%M%S)"
        
        # 更新配置值
        sed -i.tmp "s/user_permission_expire = [0-9]*/user_permission_expire = $user_perm/" "$python_config_file"
        sed -i.tmp "s/role_permission_expire = [0-9]*/role_permission_expire = $role_perm/" "$python_config_file"
        sed -i.tmp "s/menu_cache_expire = [0-9]*/menu_cache_expire = $menu_cache/" "$python_config_file"
        
        rm -f "$python_config_file.tmp"
        
        log_success "Python service configuration updated"
    fi
}

# 应用自定义配置
update_custom_config() {
    local config_file="$1"
    
    if [ ! -f "$config_file" ]; then
        log_error "Configuration file not found: $config_file"
        exit 1
    fi
    
    if ! jq empty "$config_file" 2>/dev/null; then
        log_error "Invalid JSON in configuration file: $config_file"
        exit 1
    fi
    
    log_info "Applying custom configuration from: $config_file"
    local config=$(cat "$config_file")
    apply_config "$config"
    log_success "Custom configuration applied successfully"
}

# 验证配置文件
validate_config() {
    local config_file="$1"
    
    if [ ! -f "$config_file" ]; then
        log_error "Configuration file not found: $config_file"
        exit 1
    fi
    
    log_info "Validating configuration file: $config_file"
    
    if ! jq empty "$config_file" 2>/dev/null; then
        log_error "Invalid JSON format"
        exit 1
    fi
    
    # 验证必需字段
    local required_fields=("user_permission_ttl" "role_permission_ttl" "menu_cache_ttl" "session_ttl")
    local config=$(cat "$config_file")
    
    for field in "${required_fields[@]}"; do
        if ! echo "$config" | jq -e ".$field" >/dev/null 2>&1; then
            log_error "Missing required field: $field"
            exit 1
        fi
        
        local value=$(echo "$config" | jq -r ".$field")
        if ! [[ "$value" =~ ^[0-9]+$ ]] || [ "$value" -le 0 ]; then
            log_error "Invalid value for $field: $value (must be positive integer)"
            exit 1
        fi
    done
    
    log_success "Configuration file is valid"
    
    # 显示配置内容
    echo
    echo "Configuration preview:"
    echo "$config" | jq .
}

# 备份当前配置
backup_config() {
    local backup_dir="$PROJECT_ROOT/backup/cache-configs"
    local timestamp=$(date +%Y%m%d_%H%M%S)
    local backup_file="$backup_dir/cache-config-$timestamp.json"
    
    mkdir -p "$backup_dir"
    
    log_info "Creating configuration backup..."
    
    # 收集当前配置
    local current_config="{}"
    
    if check_redis; then
        current_config=$(redis-cli HGETALL openpenpal:cache:config 2>/dev/null | \
            awk 'NR%2{key=$0; next} {print "\"" key "\": " $0 ","}' | \
            sed '$s/,$//' | \
            sed '1s/^/{/' | \
            sed '$s/$/}/')
    fi
    
    # 如果Redis没有配置，使用默认值
    if [ "$current_config" = "{}" ]; then
        current_config='{
            "user_permission_ttl": 300,
            "role_permission_ttl": 600,
            "menu_cache_ttl": 1800,
            "session_ttl": 86400,
            "source": "default_values",
            "backup_timestamp": "'$(date -Iseconds)'"
        }'
    else
        current_config=$(echo "$current_config" | jq '. + {"backup_timestamp": "'$(date -Iseconds)'"}')
    fi
    
    echo "$current_config" | jq . > "$backup_file"
    
    log_success "Configuration backed up to: $backup_file"
}

# 恢复配置
restore_config() {
    local backup_file="$1"
    
    if [ ! -f "$backup_file" ]; then
        log_error "Backup file not found: $backup_file"
        exit 1
    fi
    
    log_info "Restoring configuration from: $backup_file"
    
    # 验证备份文件
    if ! jq empty "$backup_file" 2>/dev/null; then
        log_error "Invalid backup file format"
        exit 1
    fi
    
    local config=$(cat "$backup_file")
    apply_config "$config"
    
    log_success "Configuration restored successfully"
}

# 主函数
main() {
    case "${1:-}" in
        show)
            show_config
            ;;
        check)
            check_cache_health
            ;;
        update-security)
            check_dependencies
            update_security_config
            ;;
        update-performance)
            check_dependencies
            update_performance_config
            ;;
        update-custom)
            if [ -z "$2" ]; then
                log_error "Configuration file required for update-custom"
                show_help
                exit 1
            fi
            check_dependencies
            update_custom_config "$2"
            ;;
        validate)
            if [ -z "$2" ]; then
                log_error "Configuration file required for validate"
                show_help
                exit 1
            fi
            validate_config "$2"
            ;;
        backup)
            check_dependencies
            backup_config
            ;;
        restore)
            if [ -z "$2" ]; then
                log_error "Backup file required for restore"
                show_help
                exit 1
            fi
            check_dependencies
            restore_config "$2"
            ;;
        help|--help|-h)
            show_help
            ;;
        "")
            log_info "OpenPenPal Cache Configuration Manager"
            echo
            show_help
            ;;
        *)
            log_error "Unknown action: $1"
            show_help
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"