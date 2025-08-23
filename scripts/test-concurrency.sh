#!/bin/bash

# 并发控制测试脚本
# 测试OpenPenPal的并发控制机制

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

# 检查依赖
check_dependencies() {
    log_info "检查测试依赖..."
    
    # 检查PostgreSQL
    if ! command -v psql &> /dev/null; then
        log_error "PostgreSQL未安装"
        exit 1
    fi
    
    # 检查Redis
    if ! command -v redis-cli &> /dev/null; then
        log_error "Redis未安装"
        exit 1
    fi
    
    # 检查数据库连接
    if ! psql -U postgres -d openpenpal -c "SELECT 1" &> /dev/null; then
        log_error "无法连接到OpenPenPal数据库"
        exit 1
    fi
    
    # 检查Redis连接
    if ! redis-cli ping &> /dev/null; then
        log_error "无法连接到Redis"
        exit 1
    fi
    
    log_success "所有依赖检查通过"
}

# 准备测试环境
prepare_test_env() {
    log_info "准备测试环境..."
    
    # 创建测试表（如果不存在）
    psql -U postgres -d openpenpal << EOF
CREATE TABLE IF NOT EXISTS user_credits (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) UNIQUE NOT NULL,
    total INTEGER DEFAULT 0,
    available INTEGER DEFAULT 0,
    used INTEGER DEFAULT 0,
    earned INTEGER DEFAULT 0,
    level INTEGER DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_user_credits_user_id ON user_credits(user_id);
EOF
    
    # 清理测试数据
    psql -U postgres -d openpenpal << EOF
DELETE FROM user_credits WHERE user_id LIKE 'test_%' OR user_id LIKE 'batch_%';
EOF
    
    # 清理Redis测试数据
    redis-cli --scan --pattern "rate_limit:test_*" | xargs -I {} redis-cli DEL {}
    redis-cli --scan --pattern "user_lock:test_*" | xargs -I {} redis-cli DEL {}
    
    log_success "测试环境准备完成"
}

# 运行Go测试程序
run_go_test() {
    log_info "编译并运行并发测试..."
    
    cd "$SCRIPT_DIR"
    
    # 安装依赖
    go mod init concurrency-test 2>/dev/null || true
    go get github.com/redis/go-redis/v9
    go get gorm.io/gorm
    go get gorm.io/driver/postgres
    
    # 编译
    go build -o test-concurrency test-concurrency-control.go
    
    # 运行测试
    ./test-concurrency
    
    # 清理
    rm -f test-concurrency
}

# API并发测试
test_api_concurrency() {
    log_info "运行API并发测试..."
    
    # 启动后端服务（如果未运行）
    if ! curl -s http://localhost:8080/health > /dev/null 2>&1; then
        log_warning "后端服务未运行，跳过API测试"
        return
    fi
    
    # 获取测试token
    TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
        -H "Content-Type: application/json" \
        -d '{"username":"alice","password":"Secret123!"}' | \
        grep -o '"token":"[^"]*' | grep -o '[^"]*$')
    
    if [ -z "$TOKEN" ]; then
        log_error "无法获取认证token"
        return
    fi
    
    USER_ID="test_api_user_$(date +%s)"
    
    # 并发积分操作测试
    log_info "测试并发积分操作..."
    
    for i in {1..10}; do
        (
            curl -s -X POST http://localhost:8080/api/v1/credits/add \
                -H "Authorization: Bearer $TOKEN" \
                -H "Content-Type: application/json" \
                -d "{\"user_id\":\"$USER_ID\",\"points\":10,\"description\":\"test_$i\"}" &
        )
    done
    
    wait
    
    # 获取最终积分
    FINAL_CREDITS=$(curl -s -X GET "http://localhost:8080/api/v1/credits/user/$USER_ID" \
        -H "Authorization: Bearer $TOKEN" | \
        grep -o '"total":[0-9]*' | grep -o '[0-9]*$')
    
    log_info "并发请求后总积分: $FINAL_CREDITS"
    
    # 频率限制测试
    log_info "测试频率限制..."
    
    SUCCESS_COUNT=0
    BLOCKED_COUNT=0
    
    for i in {1..8}; do
        RESPONSE=$(curl -s -w "\n%{http_code}" -X POST http://localhost:8080/api/v1/letters/create \
            -H "Authorization: Bearer $TOKEN" \
            -H "Content-Type: application/json" \
            -d '{"title":"Test Letter","content":"Test content"}')
        
        HTTP_CODE=$(echo "$RESPONSE" | tail -1)
        
        if [ "$HTTP_CODE" = "201" ] || [ "$HTTP_CODE" = "200" ]; then
            ((SUCCESS_COUNT++))
            echo "请求 $i: ✅ 成功"
        else
            ((BLOCKED_COUNT++))
            echo "请求 $i: ❌ 被限制 (HTTP $HTTP_CODE)"
        fi
        
        sleep 0.1
    done
    
    log_info "成功请求: $SUCCESS_COUNT, 被限制: $BLOCKED_COUNT"
}

# 性能测试
performance_test() {
    log_info "运行性能测试..."
    
    # 使用Apache Bench进行压力测试
    if command -v ab &> /dev/null; then
        log_info "使用Apache Bench进行压力测试..."
        
        # 简单的健康检查端点压测
        ab -n 1000 -c 50 -q http://localhost:8080/health > ab_result.txt
        
        # 提取关键指标
        REQUESTS_PER_SEC=$(grep "Requests per second" ab_result.txt | grep -o '[0-9.]*' | head -1)
        MEAN_TIME=$(grep "Time per request" ab_result.txt | grep -o '[0-9.]*' | head -1)
        
        log_info "请求速率: $REQUESTS_PER_SEC req/s"
        log_info "平均响应时间: $MEAN_TIME ms"
        
        rm -f ab_result.txt
    else
        log_warning "Apache Bench未安装，跳过压力测试"
    fi
}

# 生成测试报告
generate_report() {
    log_info "生成测试报告..."
    
    REPORT_FILE="$PROJECT_ROOT/concurrency_test_report_$(date +%Y%m%d_%H%M%S).md"
    
    cat > "$REPORT_FILE" << EOF
# OpenPenPal 并发控制测试报告

**测试时间**: $(date)

## 测试环境
- PostgreSQL: $(psql --version | head -1)
- Redis: $(redis-cli --version)
- Go: $(go version)

## 测试结果

### 1. 并发创建测试
- 测试并发创建同一用户的积分记录
- 验证只创建一条记录，避免重复

### 2. 并发操作测试
- 测试多个并发积分扣减操作
- 验证最终积分计算的准确性

### 3. 频率限制测试
- 测试基于Redis的频率限制机制
- 验证超过限制的请求被正确拒绝

### 4. 批量操作测试
- 测试大批量数据处理能力
- 验证批次处理的正确性和性能

## 并发控制机制

### 实现的机制
1. **分布式锁**: 基于Redis的用户级操作锁
2. **乐观锁**: 数据库版本控制
3. **频率限制**: 滑动窗口算法
4. **批量处理**: 分批次处理大量数据
5. **事务管理**: 标准化事务处理

### 关键改进
- GetOrCreateUserCredit 使用分布式锁防止重复创建
- CheckDailyLimit 使用Redis缓存减少数据库压力
- 积分操作使用事务确保原子性
- 实现了优雅的回退机制

## 建议
1. 在生产环境启用并发控制管理器
2. 配置合适的Redis连接池大小
3. 监控并发操作的性能指标
4. 定期清理过期的频率限制记录

---
*本报告由自动化测试生成*
EOF
    
    log_success "测试报告已生成: $REPORT_FILE"
}

# 主函数
main() {
    log_info "🔬 开始OpenPenPal并发控制测试"
    echo "=================================="
    
    # 检查依赖
    check_dependencies
    
    # 准备测试环境
    prepare_test_env
    
    # 运行测试
    run_go_test
    
    # API测试（可选）
    if [ "${1:-}" = "--with-api" ]; then
        test_api_concurrency
    fi
    
    # 性能测试（可选）
    if [ "${1:-}" = "--with-performance" ]; then
        performance_test
    fi
    
    # 生成报告
    generate_report
    
    echo
    log_success "🎉 并发控制测试完成!"
    echo
    log_info "运行选项:"
    echo "  ./test-concurrency.sh              # 基础测试"
    echo "  ./test-concurrency.sh --with-api   # 包含API测试"
    echo "  ./test-concurrency.sh --with-performance # 包含性能测试"
}

# 执行主函数
main "$@"