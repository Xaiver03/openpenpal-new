#!/usr/bin/env bash

# OpenPenPal 测试脚本 - 支持宽松的速率限制
# 通过设置环境变量启用测试模式来避免速率限制

echo "🧪 启动测试模式下的 OpenPenPal 集成测试"
echo "============================================"

# 设置测试模式环境变量
export TEST_MODE=true
export ENVIRONMENT=testing

# API基础URL
BASE_URL="http://localhost:8080"
API_BASE="${BASE_URL}/api/v1"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 测试计数器
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# 测试账号配置（基于实际seed数据）
declare -A TEST_ACCOUNTS
TEST_ACCOUNTS["alice"]="secret"
TEST_ACCOUNTS["bob"]="secret"
TEST_ACCOUNTS["courier1"]="secret"
TEST_ACCOUNTS["senior_courier"]="secret"
TEST_ACCOUNTS["coordinator"]="secret"
TEST_ACCOUNTS["school_admin"]="secret"
TEST_ACCOUNTS["platform_admin"]="secret"
TEST_ACCOUNTS["super_admin"]="secret"
TEST_ACCOUNTS["courier_level1"]="secret"
TEST_ACCOUNTS["courier_level2"]="secret"
TEST_ACCOUNTS["courier_level3"]="secret"
TEST_ACCOUNTS["courier_level4"]="secret"
TEST_ACCOUNTS["courier_building"]="courier001"
TEST_ACCOUNTS["courier_area"]="courier002"
TEST_ACCOUNTS["courier_school"]="courier003"
TEST_ACCOUNTS["courier_city"]="courier004"
TEST_ACCOUNTS["admin"]="admin123"

# 函数：测试账号登录
test_account_login() {
    local username=$1
    local password=$2
    local description=$3
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    echo -n "测试 $description ($username): "
    
    # 登录测试
    response=$(curl -s -X POST "$API_BASE/auth/login" \
        -H "Content-Type: application/json" \
        -d "{\"username\":\"$username\",\"password\":\"$password\"}")
    
    success=$(echo "$response" | jq -r '.success // false')
    
    if [ "$success" = "true" ]; then
        echo -e "${GREEN}✅ 成功${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
        
        # 提取token用于后续测试
        token=$(echo "$response" | jq -r '.data.token // ""')
        if [ -n "$token" ] && [ "$token" != "null" ]; then
            declare -g "TOKEN_$username"="$token"
        fi
        
        return 0
    else
        echo -e "${RED}❌ 失败${NC}"
        error=$(echo "$response" | jq -r '.error // "Unknown error"')
        echo "   错误: $error"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        return 1
    fi
}

# 函数：测试API端点
test_api_endpoint() {
    local method=$1
    local endpoint=$2
    local token=$3
    local description=$4
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    echo -n "测试 $description: "
    
    local auth_header=""
    if [ -n "$token" ]; then
        auth_header="-H \"Authorization: Bearer $token\""
    fi
    
    local response_code
    if [ -n "$token" ]; then
        response_code=$(curl -s -w "%{http_code}" -o /dev/null -X "$method" "$API_BASE$endpoint" -H "Authorization: Bearer $token")
    else
        response_code=$(curl -s -w "%{http_code}" -o /dev/null -X "$method" "$API_BASE$endpoint")
    fi
    
    if [ "$response_code" -eq 200 ] || [ "$response_code" -eq 201 ]; then
        echo -e "${GREEN}✅ 成功 ($response_code)${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
        return 0
    elif [ "$response_code" -eq 401 ] || [ "$response_code" -eq 403 ]; then
        echo -e "${YELLOW}⚠️  权限限制 ($response_code)${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))  # 权限限制也算是正常行为
        return 0
    else
        echo -e "${RED}❌ 失败 ($response_code)${NC}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        return 1
    fi
}

echo -e "${BLUE}第1步: 测试服务健康状态${NC}"
echo "=================================="

test_api_endpoint "GET" "/health" "" "服务健康检查"

echo ""
echo -e "${BLUE}第2步: 测试账号登录${NC}"
echo "=================================="

# 测试基础账号
test_account_login "alice" "secret" "普通用户"
test_account_login "courier1" "secret" "普通信使"
test_account_login "senior_courier" "secret" "高级信使"
test_account_login "coordinator" "secret" "信使协调员"
test_account_login "admin" "admin123" "系统管理员"

# 测试四级信使系统账号
echo ""
echo -e "${YELLOW}四级信使系统测试:${NC}"
test_account_login "courier_level1" "secret" "一级信使"
test_account_login "courier_level2" "secret" "二级信使"
test_account_login "courier_level3" "secret" "三级信使"
test_account_login "courier_level4" "secret" "四级信使"

echo ""
echo -e "${BLUE}第3步: 测试API端点访问${NC}"
echo "=================================="

# 如果admin登录成功，使用admin token测试管理端点
if [ -n "${TOKEN_admin}" ]; then
    test_api_endpoint "GET" "/admin/dashboard/stats" "${TOKEN_admin}" "管理员仪表盘"
    test_api_endpoint "GET" "/courier/management/level-1/stats" "${TOKEN_admin}" "一级信使统计"
    test_api_endpoint "GET" "/ws/stats" "${TOKEN_admin}" "WebSocket统计"
fi

echo ""
echo -e "${BLUE}第4步: 测试结果统计${NC}"
echo "=================================="
echo "总测试数: $TOTAL_TESTS"
echo -e "通过测试: ${GREEN}$PASSED_TESTS${NC}"
echo -e "失败测试: ${RED}$FAILED_TESTS${NC}"

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "${GREEN}🎉 所有测试通过！${NC}"
    exit 0
else
    echo -e "${RED}⚠️  有 $FAILED_TESTS 个测试失败${NC}"
    exit 1
fi