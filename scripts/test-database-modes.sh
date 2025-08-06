#!/bin/bash

# 简化的数据库模式测试脚本

set -e

# 获取脚本目录和项目根目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }

# 清理函数
cleanup() {
    pkill -f openpenpal-backend 2>/dev/null || true
    pkill -f "npm run dev" 2>/dev/null || true
    sleep 1
}

# 测试 SQLite 模式
test_sqlite() {
    echo ""
    echo "=============================="
    echo "测试 SQLite 模式"
    echo "=============================="
    
    cd "$PROJECT_ROOT/backend"
    
    # 设置环境变量
    export DATABASE_TYPE=sqlite
    export DATABASE_URL=./openpenpal.db
    export NODE_ENV=development
    
    log_info "使用 SQLite 启动后端..."
    ./openpenpal-backend &
    local pid=$!
    
    # 等待启动
    sleep 3
    
    # 测试健康检查
    if curl -s http://localhost:8080/health | grep -q "healthy"; then
        log_success "SQLite 模式 - 健康检查通过"
    else
        log_error "SQLite 模式 - 健康检查失败"
    fi
    
    # 测试认证
    local token=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
        -H "Content-Type: application/json" \
        -d '{"username":"admin","password":"admin123"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4)
    
    if [ -n "$token" ]; then
        log_success "SQLite 模式 - 认证测试通过"
    else
        log_error "SQLite 模式 - 认证测试失败"
    fi
    
    # 测试数据库文件
    if [ -f "./openpenpal.db" ]; then
        log_success "SQLite 模式 - 数据库文件存在"
    else
        log_error "SQLite 模式 - 数据库文件不存在"
    fi
    
    # 停止服务
    kill $pid 2>/dev/null || true
    sleep 1
}

# 测试 PostgreSQL 模式
test_postgres() {
    echo ""
    echo "=============================="
    echo "测试 PostgreSQL 模式"
    echo "=============================="
    
    # 检查 PostgreSQL
    if ! command -v psql >/dev/null 2>&1; then
        log_warning "PostgreSQL 未安装，跳过测试"
        return
    fi
    
    if ! PGPASSWORD=openpenpal123 psql -h localhost -U openpenpal -d openpenpal -c "SELECT 1" >/dev/null 2>&1; then
        log_warning "PostgreSQL 未运行或凭据不正确，跳过测试"
        return
    fi
    
    cd "$PROJECT_ROOT/backend"
    
    # 设置环境变量
    export DATABASE_TYPE=postgres
    export DB_HOST=localhost
    export DB_PORT=5432
    export DB_USER=openpenpal
    export DB_PASSWORD=openpenpal123
    export DB_NAME=openpenpal
    export NODE_ENV=production
    
    log_info "使用 PostgreSQL 启动后端..."
    ./openpenpal-backend &
    local pid=$!
    
    # 等待启动
    sleep 3
    
    # 测试健康检查
    if curl -s http://localhost:8080/health | grep -q "healthy"; then
        log_success "PostgreSQL 模式 - 健康检查通过"
    else
        log_error "PostgreSQL 模式 - 健康检查失败"
    fi
    
    # 测试认证
    local token=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
        -H "Content-Type: application/json" \
        -d '{"username":"admin","password":"admin123"}' | grep -o '"token":"[^"]*' | cut -d'"' -f4)
    
    if [ -n "$token" ]; then
        log_success "PostgreSQL 模式 - 认证测试通过"
    else
        log_error "PostgreSQL 模式 - 认证测试失败"
    fi
    
    # 测试数据库连接
    if PGPASSWORD=openpenpal123 psql -h localhost -U openpenpal -d openpenpal -c "SELECT COUNT(*) FROM users;" 2>/dev/null | grep -q "[0-9]"; then
        log_success "PostgreSQL 模式 - 数据库连接正常"
    else
        log_error "PostgreSQL 模式 - 数据库连接失败"
    fi
    
    # 停止服务
    kill $pid 2>/dev/null || true
    sleep 1
}

# 主函数
main() {
    echo "OpenPenPal 数据库模式测试"
    echo "========================"
    echo "时间: $(date)"
    
    # 清理环境
    cleanup
    
    # 编译后端
    log_info "编译后端..."
    cd "$PROJECT_ROOT/backend"
    if go build -o openpenpal-backend; then
        log_success "后端编译成功"
    else
        log_error "后端编译失败"
        exit 1
    fi
    
    # 测试 SQLite
    test_sqlite
    
    # 测试 PostgreSQL
    test_postgres
    
    # 清理
    cleanup
    
    echo ""
    log_success "测试完成！"
    echo ""
    echo "总结："
    echo "- SQLite 模式：适合开发环境，零配置"
    echo "- PostgreSQL 模式：适合生产环境，需要安装配置"
    echo ""
}

# 捕获中断信号
trap cleanup EXIT INT TERM

# 运行主函数
main "$@"