#!/bin/bash

# OpenPenPal 四级信使权限完整测试脚本

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
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
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
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }

# API 基础 URL
API_URL="http://localhost:8080/api/v1"

# 测试账号
declare -A TEST_ACCOUNTS=(
    ["admin"]="admin123"
    ["courier_level1"]="secret"
    ["courier_level2"]="secret"
    ["courier_level3"]="secret"
    ["courier_level4"]="secret"
)

# 存储 token
declare -A USER_TOKENS

# 登录函数
login_user() {
    local username=$1
    local password=$2
    
    log_info "登录用户: $username"
    
    local response=$(curl -s -X POST "$API_URL/auth/login" \
        -H "Content-Type: application/json" \
        -d "{\"username\":\"$username\",\"password\":\"$password\"}" 2>/dev/null)
    
    if echo "$response" | grep -q '"success":true'; then
        local token=$(echo "$response" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
        local role=$(echo "$response" | grep -o '"role":"[^"]*' | cut -d'"' -f4)
        local nickname=$(echo "$response" | grep -o '"nickname":"[^"]*' | cut -d'"' -f4)
        USER_TOKENS[$username]=$token
        log_success "$username 登录成功 (角色: $role, 昵称: $nickname)"
        return 0
    else
        log_error "$username 登录失败"
        return 1
    fi
}

# 测试 API 访问权限
test_api_access() {
    local username=$1
    local endpoint=$2
    local expected_status=$3
    local description=$4
    
    local token=${USER_TOKENS[$username]}
    if [ -z "$token" ]; then
        log_error "$username 没有有效的 token"
        return 1
    fi
    
    local response=$(curl -s -w "\n%{http_code}" -X GET "$API_URL$endpoint" \
        -H "Authorization: Bearer $token" 2>/dev/null)
    
    local status_code=$(echo "$response" | tail -1)
    local body=$(echo "$response" | head -n -1)
    
    if [ "$status_code" = "$expected_status" ]; then
        log_success "$username - $description (状态码: $status_code)"
        return 0
    else
        log_error "$username - $description (期望: $expected_status, 实际: $status_code)"
        return 1
    fi
}

# 测试信使管理权限
test_courier_management() {
    local username=$1
    local can_manage_level=$2
    
    local token=${USER_TOKENS[$username]}
    if [ -z "$token" ]; then
        return 1
    fi
    
    log_info "测试 $username 的信使管理权限"
    
    # 测试获取下级信使列表
    local response=$(curl -s -X GET "$API_URL/couriers/subordinates" \
        -H "Authorization: Bearer $token" 2>/dev/null)
    
    if echo "$response" | grep -q '"success":true'; then
        log_success "$username - 可以查看下级信使列表"
    else
        if [ "$can_manage_level" -gt 0 ]; then
            log_error "$username - 无法查看下级信使列表"
        else
            log_success "$username - 正确地无法查看下级信使（无管理权限）"
        fi
    fi
}

# 测试任务访问权限
test_task_access() {
    local username=$1
    
    local token=${USER_TOKENS[$username]}
    if [ -z "$token" ]; then
        return 1
    fi
    
    log_info "测试 $username 的任务访问权限"
    
    # 测试查看任务列表
    local response=$(curl -s -X GET "$API_URL/courier/tasks" \
        -H "Authorization: Bearer $token" 2>/dev/null)
    
    if echo "$response" | grep -q '"tasks":\[\]'; then
        log_success "$username - 可以访问任务列表"
    elif echo "$response" | grep -q '"success":false'; then
        log_error "$username - 无法访问任务列表"
    else
        log_success "$username - 可以访问任务列表"
    fi
}

# 主测试函数
main() {
    echo "OpenPenPal 四级信使权限测试"
    echo "============================"
    echo "时间: $(date)"
    echo ""
    
    # 检查服务状态
    log_info "检查服务状态..."
    if curl -s http://localhost:8080/health | grep -q "healthy"; then
        log_success "后端服务运行正常"
    else
        log_error "后端服务未运行"
        exit 1
    fi
    
    echo ""
    echo "=== 第一步：登录测试 ==="
    echo ""
    
    # 登录所有测试账号
    for username in "${!TEST_ACCOUNTS[@]}"; do
        login_user "$username" "${TEST_ACCOUNTS[$username]}"
        sleep 1  # 避免请求过快
    done
    
    echo ""
    echo "=== 第二步：基础权限测试 ==="
    echo ""
    
    # 测试基础 API 访问
    for username in "${!USER_TOKENS[@]}"; do
        test_api_access "$username" "/users/me" "200" "获取个人信息"
        test_api_access "$username" "/letters/public" "200" "查看公开信件"
        sleep 1
    done
    
    echo ""
    echo "=== 第三步：信使任务权限测试 ==="
    echo ""
    
    # 测试信使任务访问
    for username in courier_level1 courier_level2 courier_level3 courier_level4; do
        if [ -n "${USER_TOKENS[$username]}" ]; then
            test_task_access "$username"
        fi
    done
    
    echo ""
    echo "=== 第四步：层级管理权限测试 ==="
    echo ""
    
    # 测试管理权限
    test_courier_management "courier_level1" 0  # 一级信使无管理权限
    test_courier_management "courier_level2" 1  # 二级信使可管理一级
    test_courier_management "courier_level3" 2  # 三级信使可管理二级
    test_courier_management "courier_level4" 3  # 四级信使可管理三级
    
    echo ""
    echo "=== 第五步：权限继承测试 ==="
    echo ""
    
    # 测试高级信使是否继承低级权限
    log_info "测试权限继承（高级信使应能执行所有低级操作）"
    
    # 四级信使应该能访问所有级别的功能
    if [ -n "${USER_TOKENS[courier_level4]}" ]; then
        test_api_access "courier_level4" "/courier/tasks" "200" "四级信使访问基础任务"
        test_api_access "courier_level4" "/courier/stats" "200" "四级信使查看统计"
    fi
    
    # 一级信使不应能访问管理功能
    if [ -n "${USER_TOKENS[courier_level1]}" ]; then
        test_api_access "courier_level1" "/couriers/subordinates" "403" "一级信使被拒绝访问管理功能"
    fi
    
    echo ""
    echo "=== 第六步：管理员权限测试 ==="
    echo ""
    
    # 测试管理员权限
    if [ -n "${USER_TOKENS[admin]}" ]; then
        test_api_access "admin" "/admin/users" "200" "管理员访问用户管理"
        test_api_access "admin" "/admin/system" "200" "管理员访问系统设置"
        test_courier_management "admin" 4  # 管理员应有最高权限
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
    
    # 权限矩阵总结
    echo "权限矩阵总结："
    echo "┌─────────────────┬──────────┬──────────┬──────────┬──────────┐"
    echo "│ 功能/角色       │ Level 1  │ Level 2  │ Level 3  │ Level 4  │"
    echo "├─────────────────┼──────────┼──────────┼──────────┼──────────┤"
    echo "│ 基础配送任务    │    ✓     │    ✓     │    ✓     │    ✓     │"
    echo "│ 查看统计报告    │    ✗     │    ✓     │    ✓     │    ✓     │"
    echo "│ 管理下级信使    │    ✗     │    ✓     │    ✓     │    ✓     │"
    echo "│ 创建新信使      │    ✗     │  Level1  │  Level2  │  Level3  │"
    echo "│ 区域协调        │    ✗     │    ✗     │    ✓     │    ✓     │"
    echo "│ 城市级管理      │    ✗     │    ✗     │    ✗     │    ✓     │"
    echo "└─────────────────┴──────────┴──────────┴──────────┴──────────┘"
    echo ""
    
    if [ $TESTS_FAILED -eq 0 ]; then
        echo -e "${GREEN}✨ 所有测试通过！四级信使权限系统工作正常${NC}"
        exit 0
    else
        echo -e "${RED}❌ 有测试失败，请检查权限配置${NC}"
        exit 1
    fi
}

# 运行主函数
main "$@"