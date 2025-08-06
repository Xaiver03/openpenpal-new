#!/bin/bash

# OpenPenPal 数据库服务管理器
# 自动启动和管理 PostgreSQL 和 Redis 服务

# 导入工具函数
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/utils.sh" 2>/dev/null || true

# 服务配置
POSTGRES_SERVICE_NAME="postgresql"
REDIS_SERVICE_NAME="redis"
DB_NAME="openpenpal"
DB_USER="${DB_USER:-$(whoami)}"
DB_PASSWORD="${DB_PASSWORD:-password}"
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
REDIS_PORT="${REDIS_PORT:-6379}"

# 检测操作系统
detect_os() {
    if [[ "$OSTYPE" == "darwin"* ]]; then
        echo "macos"
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        if command -v systemctl &> /dev/null; then
            echo "linux-systemd"
        else
            echo "linux-other"
        fi
    else
        echo "unknown"
    fi
}

# 检查服务是否运行
is_service_running() {
    local service="$1"
    local port="$2"
    
    # 检查端口是否被占用
    if lsof -i :$port >/dev/null 2>&1; then
        return 0
    fi
    
    # 检查服务状态
    local os=$(detect_os)
    case $os in
        macos)
            brew services list | grep "$service" | grep "started" >/dev/null 2>&1
            ;;
        linux-systemd)
            systemctl is-active --quiet "$service"
            ;;
        *)
            pgrep "$service" >/dev/null 2>&1
            ;;
    esac
}

# 启动服务
start_service() {
    local service="$1"
    local os=$(detect_os)
    
    log_info "启动 $service 服务..."
    
    case $os in
        macos)
            if command -v brew &> /dev/null; then
                brew services start "$service"
            else
                log_error "Homebrew 未安装，无法启动 $service"
                return 1
            fi
            ;;
        linux-systemd)
            sudo systemctl start "$service"
            sudo systemctl enable "$service"
            ;;
        linux-other)
            case $service in
                postgresql)
                    sudo service postgresql start 2>/dev/null || \
                    sudo /etc/init.d/postgresql start 2>/dev/null || \
                    pg_ctl start -D /var/lib/postgresql/data 2>/dev/null
                    ;;
                redis)
                    sudo service redis-server start 2>/dev/null || \
                    sudo /etc/init.d/redis-server start 2>/dev/null || \
                    redis-server --daemonize yes 2>/dev/null
                    ;;
            esac
            ;;
        *)
            log_error "不支持的操作系统，请手动启动 $service"
            return 1
            ;;
    esac
}

# 检查并启动 PostgreSQL
ensure_postgresql() {
    log_step "检查 PostgreSQL 服务..."
    
    # 检查 PostgreSQL 是否安装
    if ! command -v psql &> /dev/null && ! command -v postgres &> /dev/null; then
        log_error "PostgreSQL 未安装"
        log_info "安装方法："
        log_info "  macOS: brew install postgresql"
        log_info "  Ubuntu: sudo apt-get install postgresql postgresql-contrib"
        log_info "  CentOS: sudo yum install postgresql-server postgresql-contrib"
        return 1
    fi
    
    # 检查服务是否运行
    if is_service_running "$POSTGRES_SERVICE_NAME" "$DB_PORT"; then
        log_success "PostgreSQL 服务已运行"
    else
        log_warning "PostgreSQL 服务未运行，正在启动..."
        if start_service "$POSTGRES_SERVICE_NAME"; then
            sleep 3  # 等待服务启动
            if is_service_running "$POSTGRES_SERVICE_NAME" "$DB_PORT"; then
                log_success "PostgreSQL 服务启动成功"
            else
                log_error "PostgreSQL 服务启动失败"
                return 1
            fi
        else
            log_error "无法启动 PostgreSQL 服务"
            return 1
        fi
    fi
    
    # 检查数据库连接
    if test_postgresql_connection; then
        log_success "PostgreSQL 连接正常"
        return 0
    else
        log_error "PostgreSQL 连接失败"
        return 1
    fi
}

# 检查并启动 Redis
ensure_redis() {
    log_step "检查 Redis 服务..."
    
    # 检查 Redis 是否安装
    if ! command -v redis-cli &> /dev/null && ! command -v redis-server &> /dev/null; then
        log_warning "Redis 未安装（可选服务）"
        log_info "安装方法："
        log_info "  macOS: brew install redis"
        log_info "  Ubuntu: sudo apt-get install redis-server"
        log_info "  CentOS: sudo yum install redis"
        return 0  # Redis 是可选的，不阻止启动
    fi
    
    # 检查服务是否运行
    if is_service_running "$REDIS_SERVICE_NAME" "$REDIS_PORT"; then
        log_success "Redis 服务已运行"
    else
        log_warning "Redis 服务未运行，正在启动..."
        if start_service "$REDIS_SERVICE_NAME"; then
            sleep 2  # 等待服务启动
            if is_service_running "$REDIS_SERVICE_NAME" "$REDIS_PORT"; then
                log_success "Redis 服务启动成功"
            else
                log_warning "Redis 服务启动失败（不影响核心功能）"
            fi
        else
            log_warning "无法启动 Redis 服务（不影响核心功能）"
        fi
    fi
    
    # 测试 Redis 连接
    if test_redis_connection; then
        log_success "Redis 连接正常"
    else
        log_warning "Redis 连接失败（不影响核心功能）"
    fi
    
    return 0
}

# 测试 PostgreSQL 连接
test_postgresql_connection() {
    local max_attempts=5
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d postgres -c "SELECT 1;" >/dev/null 2>&1; then
            return 0
        fi
        
        log_debug "PostgreSQL 连接测试失败 (尝试 $attempt/$max_attempts)"
        sleep 2
        ((attempt++))
    done
    
    return 1
}

# 测试 Redis 连接
test_redis_connection() {
    if command -v redis-cli &> /dev/null; then
        redis-cli -h localhost -p "$REDIS_PORT" ping >/dev/null 2>&1
    else
        return 1
    fi
}

# 创建数据库（如果不存在）
ensure_database() {
    log_step "检查数据库 '$DB_NAME'..."
    
    if psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -lqt | cut -d \| -f 1 | grep -qw "$DB_NAME"; then
        log_success "数据库 '$DB_NAME' 已存在"
    else
        log_info "创建数据库 '$DB_NAME'..."
        if createdb -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" "$DB_NAME" 2>/dev/null; then
            log_success "数据库 '$DB_NAME' 创建成功"
        else
            log_error "数据库 '$DB_NAME' 创建失败"
            return 1
        fi
    fi
    
    return 0
}

# 设置环境变量
setup_database_environment() {
    log_step "设置数据库环境变量..."
    
    export DATABASE_TYPE="postgres"
    export DATABASE_URL="postgresql://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME"
    export DATABASE_NAME="$DB_NAME"
    export DB_HOST="$DB_HOST"
    export DB_PORT="$DB_PORT"
    export DB_USER="$DB_USER"
    export DB_PASSWORD="$DB_PASSWORD"
    export DB_SSLMODE="disable"
    
    # Redis 环境变量
    export REDIS_HOST="localhost"
    export REDIS_PORT="$REDIS_PORT"
    export REDIS_PASSWORD=""
    
    log_success "数据库环境变量设置完成"
    log_debug "DATABASE_URL: postgres://$DB_USER:***@$DB_HOST:$DB_PORT/$DB_NAME"
}

# 显示数据库状态
show_database_status() {
    echo ""
    echo "==================== 数据库状态 ===================="
    
    # PostgreSQL 状态
    if is_service_running "$POSTGRES_SERVICE_NAME" "$DB_PORT"; then
        echo "✅ PostgreSQL: 运行中 (端口: $DB_PORT)"
        if test_postgresql_connection; then
            echo "✅ PostgreSQL 连接: 正常"
        else
            echo "❌ PostgreSQL 连接: 失败"
        fi
    else
        echo "❌ PostgreSQL: 未运行"
    fi
    
    # Redis 状态
    if is_service_running "$REDIS_SERVICE_NAME" "$REDIS_PORT"; then
        echo "✅ Redis: 运行中 (端口: $REDIS_PORT)"
        if test_redis_connection; then
            echo "✅ Redis 连接: 正常"
        else
            echo "⚠️  Redis 连接: 失败"
        fi
    else
        echo "⚠️  Redis: 未运行（可选）"
    fi
    
    echo "=================================================="
    echo ""
}

# 主函数
main() {
    local action="${1:-start}"
    
    case $action in
        start|ensure)
            log_info "🚀 启动数据库服务..."
            
            # 启动 PostgreSQL
            if ensure_postgresql; then
                if ensure_database; then
                    setup_database_environment
                else
                    log_error "数据库初始化失败"
                    return 1
                fi
            else
                log_error "PostgreSQL 启动失败"
                return 1
            fi
            
            # 启动 Redis（可选）
            ensure_redis
            
            # 显示状态
            show_database_status
            
            log_success "数据库服务准备完成！"
            ;;
        status)
            show_database_status
            ;;
        test)
            log_info "测试数据库连接..."
            if test_postgresql_connection; then
                log_success "PostgreSQL 连接正常"
            else
                log_error "PostgreSQL 连接失败"
            fi
            
            if test_redis_connection; then
                log_success "Redis 连接正常"
            else
                log_warning "Redis 连接失败"
            fi
            ;;
        *)
            echo "用法: $0 [start|ensure|status|test]"
            echo "  start/ensure  - 启动并确保数据库服务运行"
            echo "  status        - 显示数据库服务状态"
            echo "  test          - 测试数据库连接"
            ;;
    esac
}

# 如果直接运行此脚本
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    main "$@"
fi