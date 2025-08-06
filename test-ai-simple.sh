#!/bin/bash

# 简化AI功能测试脚本
echo "🤖 AI功能端点测试"
echo "=================="

API_URL="http://localhost:8080/api/v1"

# 获取token
echo "获取认证token..."
LOGIN_RESPONSE=$(curl -s -X POST "$API_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"admin123"}')

TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"data"[[:space:]]*:[[:space:]]*{[^}]*"token"[[:space:]]*:[[:space:]]*"[^"]*"' | grep -o '"token"[[:space:]]*:[[:space:]]*"[^"]*"' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
    TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"token"[[:space:]]*:[[:space:]]*"[^"]*"' | cut -d'"' -f4)
fi

if [ -z "$TOKEN" ]; then
    echo "❌ 无法获取token"
    exit 1
fi

echo "✅ Token获取成功"

# 测试AI端点
test_count=0
pass_count=0

test_endpoint() {
    local endpoint=$1
    local description=$2
    test_count=$((test_count + 1))
    
    echo -n "[$test_count] $description... "
    
    response=$(curl -s -w "\n%{http_code}" -H "Authorization: Bearer $TOKEN" "$API_URL$endpoint")
    http_code=$(echo "$response" | tail -n1)
    
    if [ "$http_code" = "200" ]; then
        echo "✅ OK"
        pass_count=$((pass_count + 1))
    else
        echo "❌ Failed (HTTP $http_code)"
    fi
}

# 测试AI基础功能
echo ""
echo "测试AI基础功能："
test_endpoint "/ai/personas" "获取AI人设列表"
test_endpoint "/ai/stats" "获取AI使用统计"
test_endpoint "/ai/daily-inspiration" "获取每日灵感"

# 测试AI管理功能
echo ""
echo "测试AI管理功能："
test_endpoint "/admin/ai/config" "获取AI配置"
test_endpoint "/admin/ai/monitoring" "获取AI监控数据"
test_endpoint "/admin/ai/analytics" "获取AI分析数据"
test_endpoint "/admin/ai/logs" "获取AI操作日志"

echo ""
echo "测试完成："
echo "总测试: $test_count"
echo "通过: $pass_count"
echo "成功率: $(( (pass_count * 100) / test_count ))%"