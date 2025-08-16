#!/bin/bash

# 简化信件管理测试脚本
echo "📮 信件管理端点测试"
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

# 测试函数
test_count=0
pass_count=0

test_endpoint() {
    local method=$1
    local endpoint=$2
    local description=$3
    local data=$4
    test_count=$((test_count + 1))
    
    echo -n "[$test_count] $description... "
    
    if [ -n "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X "$method" -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" -d "$data" "$API_URL$endpoint")
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" -H "Authorization: Bearer $TOKEN" "$API_URL$endpoint")
    fi
    
    http_code=$(echo "$response" | tail -n1)
    
    if [ "$http_code" = "200" ] || [ "$http_code" = "201" ]; then
        echo "✅ OK (HTTP $http_code)"
        pass_count=$((pass_count + 1))
        
        # 提取ID用于后续测试
        body=$(echo "$response" | sed '$d')
        if [ "$endpoint" = "/letters/" ] && [ "$method" = "POST" ] && [ -z "$DRAFT_ID" ]; then
            DRAFT_ID=$(echo "$body" | grep -o '"id"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | cut -d'"' -f4)
            if [ -n "$DRAFT_ID" ]; then
                echo "    → 保存草稿ID: $DRAFT_ID"
            fi
        fi
    else
        echo "❌ Failed (HTTP $http_code)"
    fi
}

# 测试公开信件功能
echo ""
echo "测试公开信件功能："
curl -s "$API_URL/letters/public" > /dev/null && echo "[✓] 公开信件列表 - OK" || echo "[✗] 公开信件列表 - Failed"

# 测试信件基础管理
echo ""
echo "测试信件基础管理："
test_endpoint "GET" "/letters/" "获取用户信件列表"
test_endpoint "GET" "/letters/stats" "获取信件统计"
test_endpoint "GET" "/letters/drafts" "获取草稿列表"
test_endpoint "GET" "/letters/templates" "获取模板列表"

# 创建测试草稿
echo ""
echo "测试草稿创建和管理："
draft_data='{"title":"测试草稿","content":"测试内容","is_draft":true}'
test_endpoint "POST" "/letters/" "创建草稿" "$draft_data"

# 如果成功创建草稿，继续测试
if [ -n "$DRAFT_ID" ]; then
    test_endpoint "GET" "/letters/$DRAFT_ID" "获取草稿详情"
    
    update_data='{"title":"更新的草稿","content":"更新的内容"}'
    test_endpoint "PUT" "/letters/$DRAFT_ID" "更新草稿" "$update_data"
    
    test_endpoint "POST" "/letters/$DRAFT_ID/generate-code" "生成二维码"
    test_endpoint "POST" "/letters/$DRAFT_ID/publish" "发布信件"
    
    # 清理
    test_endpoint "DELETE" "/letters/$DRAFT_ID" "删除测试草稿"
fi

# 测试搜索功能
echo ""
echo "测试搜索和发现功能："
search_data='{"query":"测试","limit":5}'
test_endpoint "POST" "/letters/search" "搜索信件" "$search_data"
test_endpoint "GET" "/letters/popular" "获取热门信件"
test_endpoint "GET" "/letters/recommended" "获取推荐信件"

# 测试写作辅助
echo ""
echo "测试写作辅助功能："
suggestion_data='{"content":"写信测试","type":"inspiration"}'
test_endpoint "POST" "/letters/writing-suggestions" "获取写作建议" "$suggestion_data"

autosave_data='{"content":"自动保存测试"}'
test_endpoint "POST" "/letters/auto-save" "自动保存" "$autosave_data"

echo ""
echo "测试完成："
echo "总测试: $test_count"
echo "通过: $pass_count"
echo "成功率: $(( (pass_count * 100) / test_count ))%"