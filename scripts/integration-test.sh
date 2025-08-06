#!/bin/bash

# OpenPenPal 完整集成测试脚本
# 测试开发模式（SQLite）和生产模式（PostgreSQL）

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

# 测试结果
TESTS_PASSED=0
TESTS_FAILED=0
TEST_RESULTS=""

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
    TESTS_PASSED=$((TESTS_PASSED + 1))
    TEST_RESULTS="${TEST_RESULTS}\n✅ $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
    TESTS_FAILED=$((TESTS_FAILED + 1))
    TEST_RESULTS="${TEST_RESULTS}\n❌ $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# 等待服务启动
wait_for_service() {
    local url=$1
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if curl -s -o /dev/null -w "%{http_code}" "$url" | grep -q "200\|404"; then
            return 0
        fi
        sleep 1
        attempt=$((attempt + 1))
    done
    
    return 1
}

# 测试 API
test_api() {
    local mode=$1
    local base_url="http://localhost:8080"
    
    log_info "测试 $mode 模式的 API..."
    
    # 1. 健康检查
    if curl -s "$base_url/health" | grep -q "healthy"; then
        log_success "$mode - 健康检查 API"
    else
        log_error "$mode - 健康检查 API"
    fi
    
    # 2. 登录测试
    local login_response=$(curl -s -X POST "$base_url/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d '{"username":"admin","password":"admin123"}' 2>/dev/null || echo "{}")
    
    if echo "$login_response" | grep -q "token"; then
        log_success "$mode - 认证 API"
        TOKEN=$(echo "$login_response" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
    else
        log_error "$mode - 认证 API"
        TOKEN=""
    fi
    
    # 3. 用户信息
    if [ -n "$TOKEN" ]; then
        if curl -s -H "Authorization: Bearer $TOKEN" "$base_url/api/v1/users/me" | grep -q "username"; then
            log_success "$mode - 用户信息 API"
        else
            log_error "$mode - 用户信息 API"
        fi
    fi
    
    # 4. 信件列表
    if curl -s "$base_url/api/v1/letters/public" | grep -q "letters"; then
        log_success "$mode - 信件列表 API"
    else
        log_error "$mode - 信件列表 API"
    fi
    
    # 5. 博物馆统计
    if curl -s "$base_url/api/v1/museum/stats" | grep -q "total"; then
        log_success "$mode - 博物馆统计 API"
    else
        log_error "$mode - 博物馆统计 API"
    fi
}

# 检查数据库
check_database() {
    local mode=$1
    local db_type=$2
    
    cd "$PROJECT_ROOT/backend"
    
    if [ "$db_type" = "postgres" ]; then
        if PGPASSWORD=openpenpal123 psql -h localhost -U openpenpal -d openpenpal -c "SELECT COUNT(*) FROM users;" 2>/dev/null | grep -q "[0-9]"; then
            log_success "$mode - PostgreSQL 数据库连接"
        else
            log_error "$mode - PostgreSQL 数据库连接"
        fi
    else
        if [ -f "./openpenpal.db" ]; then
            log_success "$mode - SQLite 数据库文件存在"
        else
            log_error "$mode - SQLite 数据库文件存在"
        fi
    fi
}

# 停止所有服务
cleanup() {
    log_info "清理服务..."
    cd "$PROJECT_ROOT"
    ./startup/stop-all.sh >/dev/null 2>&1 || true
    sleep 2
}

# 主测试函数
run_test() {
    local mode=$1
    local expected_db=$2
    
    echo ""
    echo "=============================="
    echo "测试 $mode 模式"
    echo "=============================="
    
    # 清理之前的服务
    cleanup
    
    # 启动服务
    log_info "启动 $mode 模式..."
    cd "$PROJECT_ROOT"
    
    if [ "$mode" = "production" ]; then
        # 生产模式会自动使用 PostgreSQL
        export DATABASE_TYPE="postgres"
        ./startup/quick-start.sh production >/dev/null 2>&1 &
    else
        # 开发模式使用 SQLite
        export DATABASE_TYPE="sqlite"
        export DATABASE_URL="./openpenpal.db"
        ./startup/quick-start.sh development >/dev/null 2>&1 &
    fi
    
    local start_pid=$!
    
    # 等待服务启动
    log_info "等待服务启动..."
    sleep 5  # 给服务更多时间启动
    
    # 检查后端进程
    if ps aux | grep -q "[o]penpenpal-backend"; then
        log_info "后端进程已启动"
    else
        log_error "后端进程未找到"
        # 显示最近的日志
        echo "后端日志:"
        tail -20 "$PROJECT_ROOT/logs/go-backend.log" 2>/dev/null || true
    fi
    
    if wait_for_service "http://localhost:8080/health"; then
        log_success "$mode - 服务启动成功"
    else
        log_error "$mode - 服务启动失败"
        cleanup
        return 1
    fi
    
    # 检查数据库类型
    check_database "$mode" "$expected_db"
    
    # 测试 API
    test_api "$mode"
    
    # 检查前端
    if wait_for_service "http://localhost:3000"; then
        if curl -s "http://localhost:3000" | grep -q "OpenPenPal\|React"; then
            log_success "$mode - 前端服务"
        else
            log_error "$mode - 前端服务"
        fi
    else
        log_error "$mode - 前端服务未启动"
    fi
    
    # 清理
    cleanup
}

# 主函数
main() {
    echo "OpenPenPal 完整集成测试"
    echo "========================"
    echo "时间: $(date)"
    echo ""
    
    # 预检查
    log_info "执行预检查..."
    
    # 检查 Node.js
    if command -v node >/dev/null 2>&1; then
        log_success "Node.js 已安装 ($(node --version))"
    else
        log_error "Node.js 未安装"
        exit 1
    fi
    
    # 检查 Go
    if command -v go >/dev/null 2>&1; then
        log_success "Go 已安装 ($(go version | awk '{print $3}'))"
    else
        log_error "Go 未安装"
        exit 1
    fi
    
    # 检查 PostgreSQL（仅警告）
    if command -v psql >/dev/null 2>&1; then
        if PGPASSWORD=openpenpal123 psql -h localhost -U openpenpal -d openpenpal -c "SELECT 1" >/dev/null 2>&1; then
            log_success "PostgreSQL 已运行"
        else
            log_warning "PostgreSQL 未运行，生产模式测试可能失败"
        fi
    else
        log_warning "PostgreSQL 未安装，生产模式测试可能失败"
    fi
    
    # 编译后端
    log_info "编译后端..."
    cd "$PROJECT_ROOT/backend"
    if go build -o openpenpal-backend; then
        log_success "后端编译成功"
    else
        log_error "后端编译失败"
        exit 1
    fi
    
    # 测试开发模式（SQLite）
    run_test "development" "sqlite"
    
    # 测试生产模式（PostgreSQL）
    if command -v psql >/dev/null 2>&1; then
        run_test "production" "postgres"
    else
        log_warning "跳过生产模式测试（PostgreSQL 未安装）"
    fi
    
    # 生成测试报告
    echo ""
    echo "=============================="
    echo "测试报告"
    echo "=============================="
    echo -e "测试结果：$TEST_RESULTS"
    echo ""
    echo "总计: $((TESTS_PASSED + TESTS_FAILED)) 个测试"
    echo -e "${GREEN}通过: $TESTS_PASSED${NC}"
    echo -e "${RED}失败: $TESTS_FAILED${NC}"
    echo ""
    
    if [ $TESTS_FAILED -eq 0 ]; then
        echo -e "${GREEN}✨ 所有测试通过！${NC}"
        exit 0
    else
        echo -e "${RED}❌ 有测试失败，请检查日志${NC}"
        exit 1
    fi
}

# 捕获中断信号
trap cleanup EXIT INT TERM

# 运行主函数
main "$@"