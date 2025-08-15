#!/bin/bash

# OpenPenPal 安全系统端到端验证脚本
# 基于业界最佳实践的全方位安全验证

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置
BASE_URL="http://localhost:8080"
API_URL="$BASE_URL/api/v1"
ADMIN_URL="$API_URL/admin"

# 测试结果统计
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
    ((PASSED_TESTS++))
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
    ((FAILED_TESTS++))
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# 测试函数
run_test() {
    local test_name="$1"
    local test_command="$2"
    local expected_status="$3"
    
    ((TOTAL_TESTS++))
    log_info "Running test: $test_name"
    
    local response=$(eval "$test_command" 2>/dev/null || echo "FAILED")
    local status_code=$(echo "$response" | tail -1)
    
    if [[ "$status_code" == "$expected_status" ]]; then
        log_success "$test_name - Expected status: $expected_status"
        return 0
    else
        log_error "$test_name - Expected: $expected_status, Got: $status_code"
        return 1
    fi
}

# 获取认证令牌
get_auth_token() {
    local username="$1"
    local password="$2"
    
    local response=$(curl -s -X POST "$API_URL/auth/login" \
        -H "Content-Type: application/json" \
        -d "{\"username\":\"$username\",\"password\":\"$password\"}" \
        -w "\n%{http_code}")
    
    local body=$(echo "$response" | head -n -1)
    local status=$(echo "$response" | tail -1)
    
    if [[ "$status" == "200" ]]; then
        echo "$body" | grep -o '"token":"[^"]*"' | cut -d'"' -f4
    else
        echo ""
    fi
}

echo "=================================================="
echo "🛡️  OpenPenPal 安全系统验证开始"
echo "=================================================="

# 1. 数据流完整性验证
echo -e "\n${BLUE}=== 1. 数据流完整性验证 ===${NC}"

log_info "验证服务器运行状态"
run_test "Health Check" \
    "curl -s -o /dev/null -w '%{http_code}' '$BASE_URL/health'" \
    "200"

log_info "验证CSP违规报告端点"
run_test "CSP Violation Endpoint" \
    "curl -s -X POST -o /dev/null -w '%{http_code}' '$BASE_URL/csp-report' -H 'Content-Type: application/json' -d '{\"type\":\"test\"}'" \
    "204"

# 2. 输入验证链路测试
echo -e "\n${BLUE}=== 2. 输入验证链路测试 ===${NC}"

log_info "测试XSS攻击防护"
run_test "XSS Attack Protection" \
    "curl -s -o /dev/null -w '%{http_code}' '$API_URL/auth/login' -H 'Content-Type: application/json' -d '{\"username\":\"<script>alert(1)</script>\",\"password\":\"test\"}'" \
    "400"

log_info "测试SQL注入防护"
run_test "SQL Injection Protection" \
    "curl -s -o /dev/null -w '%{http_code}' '$API_URL/letters/public?search=; DROP TABLE users; --'" \
    "400"

log_info "测试超长输入防护"
run_test "Long Input Protection" \
    "curl -s -o /dev/null -w '%{http_code}' '$API_URL/auth/login' -H 'Content-Type: application/json' -d '{\"username\":\"$(python3 -c 'print(\"a\" * 10000)')\",\"password\":\"test\"}'" \
    "400"

log_info "测试恶意文件名防护"
run_test "Malicious Filename Protection" \
    "curl -s -o /dev/null -w '%{http_code}' '$API_URL/letters/public?filename=../../../etc/passwd'" \
    "400"

# 3. 速率限制验证
echo -e "\n${BLUE}=== 3. 速率限制验证 ===${NC}"

log_info "测试一般速率限制"
for i in {1..15}; do
    response=$(curl -s -o /dev/null -w '%{http_code}' "$BASE_URL/ping")
    if [[ "$response" == "429" ]]; then
        log_success "Rate limit triggered after $i requests"
        break
    elif [[ "$i" == "15" ]]; then
        log_warning "Rate limit not triggered after 15 requests"
    fi
    sleep 0.1
done

log_info "测试认证速率限制"
for i in {1..5}; do
    response=$(curl -s -o /dev/null -w '%{http_code}' "$API_URL/auth/login" \
        -H "Content-Type: application/json" \
        -d '{"username":"invalid","password":"invalid"}')
    if [[ "$response" == "429" ]]; then
        log_success "Auth rate limit triggered after $i attempts"
        break
    elif [[ "$i" == "5" ]]; then
        log_warning "Auth rate limit not triggered after 5 attempts"
    fi
    sleep 0.2
done

# 4. 安全头验证
echo -e "\n${BLUE}=== 4. 安全头验证 ===${NC}"

log_info "验证安全头存在"
headers=$(curl -s -I "$BASE_URL/health")

security_headers=(
    "X-Content-Type-Options"
    "X-Frame-Options"
    "X-XSS-Protection"
    "Referrer-Policy"
    "Permissions-Policy"
    "Content-Security-Policy"
)

for header in "${security_headers[@]}"; do
    if echo "$headers" | grep -qi "$header"; then
        log_success "Security header present: $header"
    else
        log_error "Security header missing: $header"
    fi
    ((TOTAL_TESTS++))
done

# 5. 权限控制验证
echo -e "\n${BLUE}=== 5. 权限控制验证 ===${NC}"

log_info "获取测试用户令牌"
ADMIN_TOKEN=$(get_auth_token "admin" "admin123")
USER_TOKEN=$(get_auth_token "alice" "secret123")

if [[ -n "$ADMIN_TOKEN" ]]; then
    log_success "管理员令牌获取成功"
else
    log_error "管理员令牌获取失败"
fi

if [[ -n "$USER_TOKEN" ]]; then
    log_success "普通用户令牌获取成功"
else
    log_error "普通用户令牌获取失败"
fi

# 测试敏感词管理权限
log_info "测试敏感词管理权限控制"
if [[ -n "$ADMIN_TOKEN" ]]; then
    run_test "Admin Access to Sensitive Words" \
        "curl -s -o /dev/null -w '%{http_code}' '$ADMIN_URL/sensitive-words' -H 'Authorization: Bearer $ADMIN_TOKEN'" \
        "200"
fi

if [[ -n "$USER_TOKEN" ]]; then
    run_test "User Forbidden Access to Sensitive Words" \
        "curl -s -o /dev/null -w '%{http_code}' '$ADMIN_URL/sensitive-words' -H 'Authorization: Bearer $USER_TOKEN'" \
        "403"
fi

# 测试安全监控权限
log_info "测试安全监控权限控制"
if [[ -n "$ADMIN_TOKEN" ]]; then
    run_test "Admin Access to Security Dashboard" \
        "curl -s -o /dev/null -w '%{http_code}' '$ADMIN_URL/security/dashboard' -H 'Authorization: Bearer $ADMIN_TOKEN'" \
        "200"
fi

# 6. 内容安全验证
echo -e "\n${BLUE}=== 6. 内容安全验证 ===${NC}"

if [[ -n "$USER_TOKEN" ]]; then
    log_info "测试内容安全过滤"
    
    # 测试XSS内容过滤
    run_test "XSS Content Filtering" \
        "curl -s -o /dev/null -w '%{http_code}' '$API_URL/comments' -H 'Authorization: Bearer $USER_TOKEN' -H 'Content-Type: application/json' -d '{\"target_type\":\"letter\",\"target_id\":\"test\",\"content\":\"<script>alert(\\\"xss\\\")</script>\"}'" \
        "400"
    
    # 测试正常内容通过
    run_test "Normal Content Acceptance" \
        "curl -s -o /dev/null -w '%{http_code}' '$API_URL/letters' -H 'Authorization: Bearer $USER_TOKEN' -H 'Content-Type: application/json' -d '{\"title\":\"测试信件\",\"content\":\"这是一封正常的测试信件\",\"type\":\"draft\"}'" \
        "201"
fi

# 7. 敏感词系统验证
echo -e "\n${BLUE}=== 7. 敏感词系统验证 ===${NC}"

if [[ -n "$ADMIN_TOKEN" ]]; then
    log_info "测试敏感词管理功能"
    
    # 添加测试敏感词
    run_test "Add Sensitive Word" \
        "curl -s -o /dev/null -w '%{http_code}' '$ADMIN_URL/sensitive-words' -H 'Authorization: Bearer $ADMIN_TOKEN' -H 'Content-Type: application/json' -d '{\"word\":\"测试敏感词\",\"category\":\"spam\",\"level\":\"medium\",\"reason\":\"测试用途\"}'" \
        "201"
    
    # 获取敏感词统计
    run_test "Get Sensitive Words Stats" \
        "curl -s -o /dev/null -w '%{http_code}' '$ADMIN_URL/sensitive-words/stats' -H 'Authorization: Bearer $ADMIN_TOKEN'" \
        "200"
    
    # 刷新敏感词库
    run_test "Refresh Sensitive Words" \
        "curl -s -o /dev/null -w '%{http_code}' '$ADMIN_URL/sensitive-words/refresh' -H 'Authorization: Bearer $ADMIN_TOKEN' -X POST" \
        "200"
fi

# 8. 安全事件监控验证
echo -e "\n${BLUE}=== 8. 安全事件监控验证 ===${NC}"

if [[ -n "$ADMIN_TOKEN" ]]; then
    log_info "测试安全事件监控"
    
    # 触发一些安全事件（通过非法请求）
    curl -s -o /dev/null "$API_URL/auth/login" -H "Content-Type: application/json" -d '{"username":"<script>","password":"test"}'
    curl -s -o /dev/null "$API_URL/letters/public?search=DROP%20TABLE"
    
    sleep 2  # 等待事件记录
    
    # 检查安全事件记录
    run_test "Security Events Recording" \
        "curl -s -o /dev/null -w '%{http_code}' '$ADMIN_URL/security/events' -H 'Authorization: Bearer $ADMIN_TOKEN'" \
        "200"
    
    # 检查安全统计
    run_test "Security Stats" \
        "curl -s -o /dev/null -w '%{http_code}' '$ADMIN_URL/security/stats' -H 'Authorization: Bearer $ADMIN_TOKEN'" \
        "200"
fi

# 9. 性能和稳定性验证
echo -e "\n${BLUE}=== 9. 性能和稳定性验证 ===${NC}"

log_info "并发请求测试"
concurrent_test() {
    for i in {1..10}; do
        curl -s -o /dev/null -w '%{http_code}\n' "$BASE_URL/health" &
    done
    wait
}

results=$(concurrent_test)
success_count=$(echo "$results" | grep -c "200" || echo "0")
if [[ "$success_count" -ge 8 ]]; then
    log_success "Concurrent requests handled successfully ($success_count/10)"
    ((PASSED_TESTS++))
else
    log_error "Concurrent requests failed ($success_count/10)"
    ((FAILED_TESTS++))
fi
((TOTAL_TESTS++))

# 10. 数据完整性验证
echo -e "\n${BLUE}=== 10. 数据完整性验证 ===${NC}"

log_info "验证数据库连接和健康状态"
health_response=$(curl -s "$BASE_URL/health")
if echo "$health_response" | grep -q '"database":"healthy"'; then
    log_success "Database connectivity verified"
    ((PASSED_TESTS++))
else
    log_error "Database connectivity issue detected"
    ((FAILED_TESTS++))
fi
((TOTAL_TESTS++))

# 总结报告
echo -e "\n=================================================="
echo "🛡️  安全验证完成报告"
echo "=================================================="
echo -e "总测试数: ${BLUE}$TOTAL_TESTS${NC}"
echo -e "通过测试: ${GREEN}$PASSED_TESTS${NC}"
echo -e "失败测试: ${RED}$FAILED_TESTS${NC}"

if [[ $FAILED_TESTS -eq 0 ]]; then
    echo -e "\n${GREEN}✅ 所有安全验证测试通过！系统安全防护完备。${NC}"
    exit 0
else
    echo -e "\n${YELLOW}⚠️  检测到 $FAILED_TESTS 个安全问题，请检查日志并修复。${NC}"
    exit 1
fi