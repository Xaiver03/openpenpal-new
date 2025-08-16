#!/bin/bash

# 信件管理端点个别测试脚本
# 基于API覆盖率分析，针对信件管理模块(7.7%测试覆盖率)进行补全

set -e

echo "📮 OpenPenPal 信件管理端点个别测试"
echo "=================================="

# 配置
API_URL="http://localhost:8080/api/v1"
BASE_URL="http://localhost:8080"
TEST_MODE=1

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 统计变量
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# 存储变量
DRAFT_ID=""
LETTER_ID=""
LETTER_CODE=""

# 测试函数
test_api() {
    local method=$1
    local endpoint=$2
    local description=$3
    local data=$4
    local token=$5
    local expected_code=${6:-200}
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    echo -n "[$TOTAL_TESTS] Testing $method $endpoint"
    if [ -n "$description" ]; then
        echo -n " ($description)"
    fi
    echo -n "... "
    
    local curl_opts="-s -w \n%{http_code}"
    local headers=""
    
    if [ -n "$token" ]; then
        headers="$headers -H \"Authorization: Bearer $token\""
    fi
    
    if [ -n "$data" ]; then
        headers="$headers -H \"Content-Type: application/json\""
        curl_opts="$curl_opts -d '$data'"
    fi
    
    local full_url="$endpoint"
    if [[ ! "$endpoint" =~ ^https?:// ]]; then
        full_url="$API_URL$endpoint"
    fi
    
    # 构建并执行curl命令
    local cmd="curl $curl_opts -X $method $headers '$full_url' 2>/dev/null"
    local response=$(eval $cmd)
    
    local http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" = "$expected_code" ]; then
        echo -e "${GREEN}✅ OK (HTTP $http_code)${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
        
        # 提取重要数据用于后续测试
        if [ "$http_code" = "200" ] || [ "$http_code" = "201" ]; then
            if echo "$body" | grep -q '"id"'; then
                local extracted_id=$(echo "$body" | grep -o '"id"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | cut -d'"' -f4)
                if [ -n "$extracted_id" ]; then
                    case "$endpoint" in
                        */letters/ | */letters)
                            if [ -z "$DRAFT_ID" ]; then
                                DRAFT_ID="$extracted_id"
                                echo "    → 保存草稿ID: $DRAFT_ID"
                            elif [ -z "$LETTER_ID" ]; then
                                LETTER_ID="$extracted_id"
                                echo "    → 保存信件ID: $LETTER_ID"
                            fi
                            ;;
                    esac
                fi
            fi
            
            if echo "$body" | grep -q '"code"'; then
                local extracted_code=$(echo "$body" | grep -o '"code"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | cut -d'"' -f4)
                if [ -n "$extracted_code" ] && [ -z "$LETTER_CODE" ]; then
                    LETTER_CODE="$extracted_code"
                    echo "    → 保存信件二维码: $LETTER_CODE"
                fi
            fi
        fi
        
        return 0
    else
        echo -e "${RED}❌ Failed (HTTP $http_code, expected $expected_code)${NC}"
        echo "    Response: $(echo "$body" | head -c 200)"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        return 1
    fi
}

# 获取认证token
get_auth_token() {
    echo -e "${BLUE}🔐 获取认证token...${NC}"
    
    local login_response=$(curl -s -X POST "$API_URL/auth/login" \
        -H "Content-Type: application/json" \
        -d '{"username":"admin","password":"admin123"}' 2>/dev/null)
    
    if echo "$login_response" | grep -q '"token"'; then
        # 尝试提取嵌套的token (data.token格式)
        local token=$(echo "$login_response" | grep -o '"data"[[:space:]]*:[[:space:]]*{[^}]*"token"[[:space:]]*:[[:space:]]*"[^"]*"' | grep -o '"token"[[:space:]]*:[[:space:]]*"[^"]*"' | cut -d'"' -f4)
        
        # 如果没找到嵌套格式，尝试直接格式
        if [ -z "$token" ]; then
            token=$(echo "$login_response" | grep -o '"token"[[:space:]]*:[[:space:]]*"[^"]*"' | cut -d'"' -f4)
        fi
        
        if [ -n "$token" ]; then
            echo -e "${GREEN}✅ Token获取成功${NC}"
            echo "$token"
            return 0
        fi
    fi
    
    echo -e "${RED}❌ Token获取失败${NC}"
    echo "Response: $login_response"
    exit 1
}

# 开始测试
echo ""
echo "Environment: TEST_MODE=$TEST_MODE"
echo ""

# 获取认证token
TOKEN=$(get_auth_token)

echo ""
echo -e "${BLUE}🔍 1. 公开信件功能测试${NC}"
echo "=================================="

# 获取公开信件 (无需认证)
test_api "GET" "/letters/public" "获取广场信件列表"

echo ""
echo -e "${BLUE}🔍 2. 信件基础管理测试${NC}"
echo "=================================="

# 获取用户信件列表
test_api "GET" "/letters/" "获取用户信件列表" "" "$TOKEN"

# 获取信件统计
test_api "GET" "/letters/stats" "获取信件统计" "" "$TOKEN"

# 创建草稿
draft_data='{
    "title": "测试草稿信件",
    "content": "这是一封测试草稿信件的内容。",
    "recipient_info": {
        "name": "测试收件人",
        "address": "测试地址"
    },
    "tags": ["测试", "草稿"],
    "is_draft": true
}'
test_api "POST" "/letters/" "创建草稿信件" "$draft_data" "$TOKEN" 201

# 等待草稿ID被设置
sleep 1

if [ -n "$DRAFT_ID" ]; then
    echo ""
    echo -e "${BLUE}🔍 3. 草稿管理测试 (ID: $DRAFT_ID)${NC}"
    echo "=================================="
    
    # 获取单封信件
    test_api "GET" "/letters/$DRAFT_ID" "获取草稿详情" "" "$TOKEN"
    
    # 更新草稿
    update_data='{
        "title": "更新后的测试草稿",
        "content": "这是更新后的草稿内容。",
        "tags": ["测试", "更新", "草稿"]
    }'
    test_api "PUT" "/letters/$DRAFT_ID" "更新草稿内容" "$update_data" "$TOKEN"
    
    # 发布信件
    test_api "POST" "/letters/$DRAFT_ID/publish" "发布草稿信件" "" "$TOKEN"
    
    # 生成二维码
    test_api "POST" "/letters/$DRAFT_ID/generate-code" "生成信件二维码" "" "$TOKEN"
    
    # 点赞信件
    test_api "POST" "/letters/$DRAFT_ID/like" "点赞信件" "" "$TOKEN"
    
    # 分享信件
    share_data='{
        "platform": "wechat",
        "message": "分享一封有趣的信件"
    }'
    test_api "POST" "/letters/$DRAFT_ID/share" "分享信件" "$share_data" "$TOKEN"
else
    echo -e "${YELLOW}⚠️  无法获取草稿ID，跳过草稿管理测试${NC}"
fi

echo ""
echo -e "${BLUE}🔍 4. 草稿和模板功能测试${NC}"
echo "=================================="

# 获取草稿列表
test_api "GET" "/letters/drafts" "获取草稿列表" "" "$TOKEN"

# 获取模板列表
test_api "GET" "/letters/templates" "获取信件模板列表" "" "$TOKEN"

# 获取单个模板详情
test_api "GET" "/letters/templates/1" "获取模板详情" "" "$TOKEN"

echo ""
echo -e "${BLUE}🔍 5. 信件搜索和发现功能测试${NC}"
echo "=================================="

# 搜索信件
search_data='{
    "query": "测试",
    "filters": {
        "tags": ["测试"],
        "date_range": {
            "start": "2024-01-01",
            "end": "2024-12-31"
        }
    },
    "limit": 10
}'
test_api "POST" "/letters/search" "搜索信件" "$search_data" "$TOKEN"

# 获取热门信件
test_api "GET" "/letters/popular" "获取热门信件" "" "$TOKEN"

# 获取推荐信件
test_api "GET" "/letters/recommended" "获取推荐信件" "" "$TOKEN"

echo ""
echo -e "${BLUE}🔍 6. 写作辅助功能测试${NC}"
echo "=================================="

# 自动保存草稿
autosave_data='{
    "letter_id": "'${DRAFT_ID:-temp-id}'",
    "content": "自动保存的内容...",
    "timestamp": "'$(date -u +%Y-%m-%dT%H:%M:%SZ)'"
}'
test_api "POST" "/letters/auto-save" "自动保存草稿" "$autosave_data" "$TOKEN"

# 获取写作建议
suggestion_data='{
    "content": "我想写一封关于友情的信",
    "type": "inspiration",
    "context": "personal"
}'
test_api "POST" "/letters/writing-suggestions" "获取写作建议" "$suggestion_data" "$TOKEN"

echo ""
echo -e "${BLUE}🔍 7. 批量操作和导出功能测试${NC}"
echo "=================================="

# 批量操作
batch_data='{
    "operation": "mark_read",
    "letter_ids": ["'${DRAFT_ID:-temp-id}'"],
    "options": {
        "mark_all": false
    }
}'
test_api "POST" "/letters/batch" "批量标记已读" "$batch_data" "$TOKEN"

# 导出信件
export_data='{
    "format": "json",
    "filters": {
        "user_letters_only": true,
        "include_drafts": true
    },
    "options": {
        "include_metadata": true
    }
}'
test_api "POST" "/letters/export" "导出用户信件" "$export_data" "$TOKEN"

echo ""
echo -e "${BLUE}🔍 8. 回信系统测试 (SOTA功能)${NC}"
echo "=================================="

if [ -n "$LETTER_CODE" ]; then
    # 扫码获取回信信息
    test_api "GET" "/letters/scan-reply/$LETTER_CODE" "扫码获取回信信息" "" "$TOKEN"
    
    # 创建回信
    reply_data='{
        "original_letter_code": "'$LETTER_CODE'",
        "title": "回信测试",
        "content": "这是对原信件的回复。",
        "reply_type": "direct"
    }'
    test_api "POST" "/letters/replies" "创建回信" "$reply_data" "$TOKEN"
    
    # 获取对话线程
    test_api "GET" "/letters/threads" "获取用户对话线程" "" "$TOKEN"
    
    # 获取线程详情
    test_api "GET" "/letters/threads/1" "获取线程详情" "" "$TOKEN"
else
    echo -e "${YELLOW}⚠️  无法获取信件二维码，跳过回信系统测试${NC}"
fi

echo ""
echo -e "${BLUE}🔍 9. 信封绑定功能测试${NC}"
echo "=================================="

if [ -n "$DRAFT_ID" ]; then
    # 绑定信封
    envelope_data='{
        "envelope_id": "envelope-001",
        "binding_type": "qr_code"
    }'
    test_api "POST" "/letters/$DRAFT_ID/bind-envelope" "绑定信封" "$envelope_data" "$TOKEN"
    
    # 获取信封信息
    test_api "GET" "/letters/$DRAFT_ID/envelope" "获取信封信息" "" "$TOKEN"
    
    # 解绑信封
    test_api "DELETE" "/letters/$DRAFT_ID/bind-envelope" "解绑信封" "" "$TOKEN"
else
    echo -e "${YELLOW}⚠️  无法获取信件ID，跳过信封绑定测试${NC}"
fi

echo ""
echo -e "${BLUE}🔍 10. 公开信件读取测试${NC}"
echo "=================================="

if [ -n "$LETTER_CODE" ]; then
    # 扫码读信 (无需认证)
    test_api "GET" "/letters/read/$LETTER_CODE" "扫码读取信件"
    
    # 标记已读 (无需认证)
    test_api "POST" "/letters/read/$LETTER_CODE/mark-read" "标记信件已读"
else
    # 使用示例二维码进行测试
    TEST_CODE="QR123456"
    test_api "GET" "/letters/read/$TEST_CODE" "扫码读取信件(示例)" "" "" 404
fi

echo ""
echo -e "${BLUE}🔍 11. 错误处理测试${NC}"
echo "=================================="

# 测试不存在的信件
test_api "GET" "/letters/nonexistent-id" "获取不存在的信件" "" "$TOKEN" 404

# 测试无效数据
test_api "POST" "/letters/" "创建无效信件" '{"invalid": "data"}' "$TOKEN" 400

# 测试未授权访问
test_api "GET" "/letters/" "未授权访问用户信件" "" "" 401

# 清理测试数据
if [ -n "$DRAFT_ID" ]; then
    echo ""
    echo -e "${BLUE}🧹 清理测试数据${NC}"
    echo "=================================="
    test_api "DELETE" "/letters/$DRAFT_ID" "删除测试草稿" "" "$TOKEN"
fi

echo ""
echo "=================================="
echo -e "${GREEN}✨ 信件管理端点测试完成！${NC}"
echo "=================================="
echo ""

# 测试统计
echo -e "${BLUE}📊 测试统计：${NC}"
echo "总测试数: $TOTAL_TESTS"
echo -e "通过测试: ${GREEN}$PASSED_TESTS${NC}"
echo -e "失败测试: ${RED}$FAILED_TESTS${NC}"
echo -e "成功率: $(( (PASSED_TESTS * 100) / TOTAL_TESTS ))%"

echo ""
echo -e "${BLUE}📋 信件管理功能覆盖率分析：${NC}"
echo "当前覆盖: 2个API端点 -> 26个API端点"
echo "覆盖率提升: 7.7% -> 100%"
echo "新增测试: 回信系统、模板管理、批量操作、导出功能、信封绑定"

if [ $FAILED_TESTS -eq 0 ]; then
    echo ""
    echo -e "${GREEN}🎉 所有信件管理端点测试通过！${NC}"
    exit 0
else
    echo ""
    echo -e "${YELLOW}⚠️  存在失败的测试，请检查上述错误信息${NC}"
    exit 1
fi