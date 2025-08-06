#!/bin/bash

# 测试不同数据库配置的启动脚本

set -e

# 获取脚本目录和项目根目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 显示信息
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# 测试数据库连接
test_database_connection() {
    local db_type=$1
    
    echo ""
    echo "=============================="
    echo "测试 $db_type 数据库配置"
    echo "=============================="
    
    cd "$PROJECT_ROOT/backend"
    
    # 设置环境变量
    if [ "$db_type" = "postgres" ]; then
        export DATABASE_TYPE="postgres"
        log_info "使用 PostgreSQL 配置"
        
        # 检查 PostgreSQL 是否运行
        if command -v psql >/dev/null 2>&1; then
            if PGPASSWORD=openpenpal123 psql -h localhost -U openpenpal -d openpenpal -c "SELECT 1" >/dev/null 2>&1; then
                log_success "PostgreSQL 连接正常"
            else
                log_error "PostgreSQL 连接失败"
                log_info "请确保 PostgreSQL 正在运行并且配置正确"
                return 1
            fi
        fi
        
        # 复制 PostgreSQL 配置
        cp .env.production .env
    else
        export DATABASE_TYPE="sqlite"
        log_info "使用 SQLite 配置"
        
        # 复制 SQLite 配置
        cp .env.development .env
    fi
    
    # 运行数据库测试
    log_info "运行数据库连接测试..."
    if go run cmd/test-db/main.go; then
        log_success "$db_type 数据库测试通过！"
        return 0
    else
        log_error "$db_type 数据库测试失败！"
        return 1
    fi
}

# 编译后端
compile_backend() {
    log_info "编译后端..."
    cd "$PROJECT_ROOT/backend"
    
    if go build -o openpenpal-backend; then
        log_success "后端编译成功"
        return 0
    else
        log_error "后端编译失败"
        return 1
    fi
}

# 测试启动脚本
test_startup_script() {
    local db_type=$1
    
    echo ""
    echo "=============================="
    echo "测试启动脚本 - $db_type 模式"
    echo "=============================="
    
    # 设置数据库类型
    export DATABASE_TYPE="$db_type"
    
    cd "$PROJECT_ROOT"
    
    # 使用 dry-run 模式测试
    log_info "测试启动脚本（dry-run 模式）..."
    if ./startup/quick-start.sh --dry-run development; then
        log_success "启动脚本配置正确"
    else
        log_error "启动脚本配置错误"
    fi
}

# 主函数
main() {
    echo "OpenPenPal 数据库启动测试"
    echo "=========================="
    echo ""
    
    # 1. 编译后端
    if ! compile_backend; then
        exit 1
    fi
    
    # 2. 测试 SQLite
    test_database_connection "sqlite"
    
    # 3. 测试 PostgreSQL
    test_database_connection "postgres"
    
    # 4. 测试启动脚本
    test_startup_script "sqlite"
    test_startup_script "postgres"
    
    echo ""
    echo "=============================="
    echo "测试总结"
    echo "=============================="
    log_info "所有测试完成"
    log_info "您现在可以使用以下方式启动："
    echo ""
    echo "  # SQLite 模式（开发）"
    echo "  export DATABASE_TYPE=sqlite"
    echo "  ./startup/quick-start.sh development"
    echo ""
    echo "  # PostgreSQL 模式（生产）"
    echo "  export DATABASE_TYPE=postgres"
    echo "  ./startup/quick-start.sh production"
    echo ""
}

# 运行主函数
main "$@"