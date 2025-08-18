#!/bin/bash

# OpenPenPal AI API全面测试脚本
# 测试所有AI相关的API端点

set -e

BASE_URL="http://localhost:8080"
TEST_LOG="/tmp/ai-api-test.log"

echo "🧠 OpenPenPal AI API接口测试"
echo "================================"
echo "开始时间: $(date)"
echo ""

# 清理之前的日志
> "$TEST_LOG"

log() {
    echo "$(date '+%H:%M:%S') - $1" | tee -a "$TEST_LOG"
}

test_endpoint() {
    local method="$1"
    local endpoint="$2"
    local data="$3"
    local expected_status="$4"
    local auth_header="$5"
    
    log "测试 $method $endpoint"
    
    local curl_cmd="curl -s -w '%{http_code}' -X $method '$BASE_URL$endpoint'"
    
    if [[ -n "$data" ]]; then
        curl_cmd="$curl_cmd -H 'Content-Type: application/json' -d '$data'"
    fi
    
    if [[ -n "$auth_header" ]]; then
        curl_cmd="$curl_cmd -H 'Authorization: $auth_header'"
    fi
    
    local response=$(eval "$curl_cmd")
    local status_code="${response: -3}"
    local body="${response%???}"
    
    if [[ "$status_code" == "$expected_status" ]]; then
        log "  ✅ 状态码: $status_code"
        if [[ -n "$body" && "$body" != "null" ]]; then
            log "  📄 响应预览: $(echo "$body" | head -c 100)..."
        fi
    else
        log "  ❌ 状态码: $status_code (期望: $expected_status)"
        log "  📄 响应: $body"
    fi
    
    echo "$status_code|$body" >> "$TEST_LOG"
}

get_auth_token() {
    log "获取认证令牌..."
    
    # Step 1: Get CSRF token
    log "  1. 获取CSRF令牌..."
    local csrf_response=$(curl -s -X GET "$BASE_URL/api/v1/auth/csrf")
    local csrf_token=$(echo "$csrf_response" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
    
    if [[ -z "$csrf_token" ]]; then
        log "  ❌ CSRF令牌获取失败: $csrf_response"
        echo ""
        return
    fi
    
    log "  ✅ CSRF令牌获取成功"
    
    # Step 2: Login with CSRF protection
    log "  2. 登录认证..."
    local login_response=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -H "X-CSRF-Token: $csrf_token" \
        -d '{"username":"admin","password":"admin123"}')
    
    if echo "$login_response" | grep -q "token"; then
        local token=$(echo "$login_response" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
        log "  ✅ 认证成功"
        echo "$token"
    else
        log "  ❌ 认证失败: $login_response"
        echo ""
    fi
}

log "🔐 第一阶段：认证测试"
echo "============================="

AUTH_TOKEN=$(get_auth_token)
AUTH_HEADER=""
if [[ -n "$AUTH_TOKEN" ]]; then
    AUTH_HEADER="Bearer $AUTH_TOKEN"
fi

log ""
log "🌐 第二阶段：公开AI API测试"
echo "============================="

# 测试公开的AI提供商状态
test_endpoint "GET" "/api/ai/providers/status" "" "200"

log ""
log "🔒 第三阶段：认证AI API测试"
echo "============================="

if [[ -n "$AUTH_TOKEN" ]]; then
    # 文本生成测试
    test_endpoint "POST" "/api/ai/generate" \
        '{"prompt":"写一首关于春天的短诗","max_tokens":100,"temperature":0.7}' \
        "200" "$AUTH_HEADER"
    
    # 聊天测试
    test_endpoint "POST" "/api/ai/chat" \
        '{"messages":[{"role":"user","content":"你好，今天天气怎么样？"}],"max_tokens":50}' \
        "200" "$AUTH_HEADER"
    
    # 文本总结测试
    test_endpoint "POST" "/api/ai/summarize" \
        '{"text":"这是一段很长的文本，需要被总结。春天来了，万物复苏，花开鸟鸣，大地一片生机勃勃的景象。"}' \
        "200" "$AUTH_HEADER"
    
    # 翻译测试
    test_endpoint "POST" "/api/ai/translate" \
        '{"text":"Hello, how are you?","target_language":"zh"}' \
        "200" "$AUTH_HEADER"
    
    # 情感分析测试
    test_endpoint "POST" "/api/ai/sentiment" \
        '{"text":"今天心情很好，阳光明媚！"}' \
        "200" "$AUTH_HEADER"
    
    # 内容审核测试
    test_endpoint "POST" "/api/ai/moderate" \
        '{"text":"这是一段正常的文本内容"}' \
        "200" "$AUTH_HEADER"
    
    # 信件写作辅助测试
    test_endpoint "POST" "/api/ai/letter/assist" \
        '{"topic":"友谊","tone":"温暖","length":"short"}' \
        "200" "$AUTH_HEADER"
    
    # 使用统计测试
    test_endpoint "GET" "/api/ai/usage/stats" "" "200" "$AUTH_HEADER"
else
    log "❌ 跳过认证API测试 - 无法获取认证令牌"
fi

log ""
log "🔧 第四阶段：管理员AI API测试"
echo "================================="

if [[ -n "$AUTH_TOKEN" ]]; then
    # 提供商管理
    test_endpoint "POST" "/api/admin/ai/providers/reload" "" "200" "$AUTH_HEADER"
    
    # AI配置获取
    test_endpoint "GET" "/api/admin/ai/config" "" "200" "$AUTH_HEADER"
    
    # 内容模板获取
    test_endpoint "GET" "/api/admin/ai/templates" "" "200" "$AUTH_HEADER"
    
    # AI监控数据
    test_endpoint "GET" "/api/admin/ai/monitoring" "" "200" "$AUTH_HEADER"
    
    # AI分析数据
    test_endpoint "GET" "/api/admin/ai/analytics" "" "200" "$AUTH_HEADER"
    
    # AI操作日志
    test_endpoint "GET" "/api/admin/ai/logs" "" "200" "$AUTH_HEADER"
    
    # AI提供商测试
    test_endpoint "POST" "/api/admin/ai/test-provider" \
        '{"provider":"local","test_text":"测试连接"}' \
        "200" "$AUTH_HEADER"
else
    log "❌ 跳过管理员API测试 - 无法获取认证令牌"
fi

log ""
log "📊 第五阶段：AI系统信息测试"
echo "================================="

# 测试基础系统信息
test_endpoint "GET" "/health" "" "200"
test_endpoint "GET" "/ping" "" "200"

log ""
log "🎯 第六阶段：错误处理测试"
echo "=========================="

if [[ -n "$AUTH_TOKEN" ]]; then
    # 测试无效数据
    test_endpoint "POST" "/api/ai/generate" \
        '{"prompt":"","max_tokens":-1}' \
        "400" "$AUTH_HEADER"
    
    # 测试不存在的端点
    test_endpoint "GET" "/api/ai/nonexistent" "" "404" "$AUTH_HEADER"
    
    # 测试无效JSON
    test_endpoint "POST" "/api/ai/chat" \
        '{"invalid":"json""}' \
        "400" "$AUTH_HEADER"
fi

# 测试未认证访问需要认证的端点
test_endpoint "GET" "/api/admin/ai/config" "" "401"
test_endpoint "POST" "/api/ai/generate" '{"prompt":"test"}' "401"

log ""
log "📈 测试结果统计"
echo "================"

total_tests=$(grep -c "测试 " "$TEST_LOG")
success_tests=$(grep -c "✅ 状态码" "$TEST_LOG")
failed_tests=$(grep -c "❌ 状态码" "$TEST_LOG")

log "总测试数: $total_tests"
log "成功测试: $success_tests"
log "失败测试: $failed_tests"
log "成功率: $(( success_tests * 100 / total_tests ))%"

log ""
log "🔍 失败测试详情:"
echo "================"

if [[ $failed_tests -gt 0 ]]; then
    grep -A1 "❌ 状态码" "$TEST_LOG" | while read line; do
        if [[ "$line" =~ ❌ ]]; then
            log "$line"
        fi
    done
else
    log "✅ 所有测试均通过！"
fi

log ""
log "📋 AI系统可用性评估:"
echo "==================="

if [[ $success_tests -gt $(( total_tests * 8 / 10 )) ]]; then
    log "🎉 AI系统整体可用性良好 (成功率 >= 80%)"
elif [[ $success_tests -gt $(( total_tests * 6 / 10 )) ]]; then
    log "⚠️  AI系统部分功能可用 (成功率 60-80%)"
else
    log "🚨 AI系统存在重大问题 (成功率 < 60%)"
fi

log ""
log "测试完成时间: $(date)"
log "详细日志: $TEST_LOG"

echo ""
echo "🎯 AI API测试完成！"
echo "📄 完整日志: $TEST_LOG"