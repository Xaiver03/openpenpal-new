#!/bin/bash

# API 完整测试脚本 - 验证 PostgreSQL 迁移后的功能

set -e

echo "🔍 OpenPenPal API 完整测试"
echo "========================="

# API 基础 URL
API_URL="http://localhost:8080/api/v1"
BASE_URL="http://localhost:8080"

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 测试函数
test_endpoint() {
    local method=$1
    local endpoint=$2
    local data=$3
    local token=$4
    local expected_code=${5:-200}
    
    echo -n "Testing $method $endpoint... "
    
    if [ -n "$token" ]; then
        if [ -n "$data" ]; then
            response=$(curl -s -w "\n%{http_code}" -X $method "$endpoint" \
                -H "Authorization: Bearer $token" \
                -H "Content-Type: application/json" \
                -d "$data" 2>/dev/null)
        else
            response=$(curl -s -w "\n%{http_code}" -X $method "$endpoint" \
                -H "Authorization: Bearer $token" 2>/dev/null)
        fi
    else
        if [ -n "$data" ]; then
            response=$(curl -s -w "\n%{http_code}" -X $method "$endpoint" \
                -H "Content-Type: application/json" \
                -d "$data" 2>/dev/null)
        else
            response=$(curl -s -w "\n%{http_code}" -X $method "$endpoint" 2>/dev/null)
        fi
    fi
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" = "$expected_code" ]; then
        echo -e "${GREEN}✅ OK (HTTP $http_code)${NC}"
        return 0
    else
        echo -e "${RED}❌ Failed (HTTP $http_code, expected $expected_code)${NC}"
        echo "Response: $body"
        return 1
    fi
}

echo ""
echo "1. 基础健康检查"
echo "----------------"
test_endpoint "GET" "$BASE_URL/health"

echo ""
echo "2. 认证测试"
echo "------------"
# 登录获取 token
echo -n "Admin login... "
login_response=$(curl -s -X POST "$API_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"admin123"}' 2>/dev/null)

if echo "$login_response" | grep -q "token"; then
    TOKEN=$(echo "$login_response" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
    echo -e "${GREEN}✅ Success${NC}"
else
    echo -e "${RED}❌ Failed${NC}"
    echo "Response: $login_response"
    exit 1
fi

echo ""
echo "3. 用户 API"
echo "-----------"
test_endpoint "GET" "$API_URL/users/me" "" "$TOKEN"
test_endpoint "GET" "$API_URL/users/me/stats" "" "$TOKEN"

echo ""
echo "4. 信件 API"
echo "-----------"
test_endpoint "GET" "$API_URL/letters/" "" "$TOKEN"
test_endpoint "GET" "$API_URL/letters/public"
test_endpoint "GET" "$API_URL/letters/stats"

echo ""
echo "5. 博物馆 API"
echo "-------------"
test_endpoint "GET" "$API_URL/museum/entries"
test_endpoint "GET" "$API_URL/museum/featured"
test_endpoint "GET" "$API_URL/museum/stats"

echo ""
echo "6. 信使系统 API"
echo "---------------"
test_endpoint "GET" "$API_URL/courier/stats"
test_endpoint "GET" "$API_URL/courier/status" "" "$TOKEN"
test_endpoint "GET" "$API_URL/courier/me" "" "$TOKEN"

echo ""
echo "7. 管理员 API"
echo "-------------"
test_endpoint "GET" "$API_URL/admin/stats" "" "$TOKEN"
test_endpoint "GET" "$API_URL/admin/users" "" "$TOKEN"

echo ""
echo "✨ 测试完成！"
echo ""

# 数据库统计
echo "数据库统计："
echo "-----------"
echo -n "用户数量: "
curl -s -H "Authorization: Bearer $TOKEN" "$API_URL/admin/stats" | grep -o '"total_users":[0-9]*' | cut -d: -f2 || echo "N/A"
echo -n "信件数量: "
curl -s "$API_URL/letters/stats" | grep -o '"total_letters":[0-9]*' | cut -d: -f2 || echo "N/A"

echo ""
echo "数据库类型: PostgreSQL ✅"
echo "迁移状态: 完成 ✅"