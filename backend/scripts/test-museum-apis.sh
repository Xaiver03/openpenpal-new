#!/bin/bash

# Museum API Test Script
# 博物馆API测试脚本

echo "🏛️ Museum API Endpoints Test"
echo "============================"

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# API基础URL
BASE_URL="http://localhost:8080/api/v1"
ADMIN_BASE_URL="http://localhost:8080/api/v1/admin"

# 测试账号
ADMIN_USER="admin"
ADMIN_PASS="Admin123!"
STUDENT_USER="alice"
STUDENT_PASS="Secret123!"

# 获取认证令牌的函数
get_token() {
    local username=$1
    local password=$2
    
    response=$(curl -s -X POST "$BASE_URL/auth/login" \
        -H "Content-Type: application/json" \
        -d "{\"username\":\"$username\",\"password\":\"$password\"}")
    
    token=$(echo $response | grep -o '"token":"[^"]*' | grep -o '[^"]*$' | head -1)
    
    if [ -z "$token" ]; then
        echo ""
    else
        echo "$token"
    fi
}

# 测试函数
test_endpoint() {
    local method=$1
    local url=$2
    local data=$3
    local token=$4
    local description=$5
    
    echo -e "\n${BLUE}Testing: $description${NC}"
    echo "Method: $method"
    echo "URL: $url"
    
    if [ -n "$data" ]; then
        echo "Data: $data"
    fi
    
    # 构建curl命令
    cmd="curl -s -X $method \"$url\""
    
    if [ -n "$token" ]; then
        cmd="$cmd -H \"Authorization: Bearer $token\""
    fi
    
    if [ -n "$data" ]; then
        cmd="$cmd -H \"Content-Type: application/json\" -d '$data'"
    fi
    
    # 执行请求
    response=$(eval $cmd)
    status_code=$(curl -s -o /dev/null -w "%{http_code}" -X $method "$url" \
        ${token:+-H "Authorization: Bearer $token"} \
        ${data:+-H "Content-Type: application/json" -d "$data"})
    
    if [ "$status_code" -ge 200 ] && [ "$status_code" -lt 300 ]; then
        echo -e "${GREEN}✅ Success (Status: $status_code)${NC}"
        echo "Response: $(echo $response | jq -C '.' 2>/dev/null || echo $response | head -c 200)"
    else
        echo -e "${RED}❌ Failed (Status: $status_code)${NC}"
        echo "Response: $(echo $response | jq -C '.' 2>/dev/null || echo $response)"
    fi
}

# 获取令牌
echo -e "${YELLOW}Getting authentication tokens...${NC}"
ADMIN_TOKEN=$(get_token "$ADMIN_USER" "$ADMIN_PASS")
STUDENT_TOKEN=$(get_token "$STUDENT_USER" "$STUDENT_PASS")

if [ -z "$ADMIN_TOKEN" ]; then
    echo -e "${RED}Failed to get admin token${NC}"
    exit 1
fi

if [ -z "$STUDENT_TOKEN" ]; then
    echo -e "${RED}Failed to get student token${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Authentication successful${NC}"

# 公开API测试
echo -e "\n${YELLOW}=== Testing Public Museum APIs ===${NC}"

test_endpoint "GET" "$BASE_URL/museum/entries" "" "" \
    "Get museum entries (public)"

test_endpoint "GET" "$BASE_URL/museum/exhibitions" "" "" \
    "Get exhibitions (public)"

test_endpoint "GET" "$BASE_URL/museum/stats" "" "" \
    "Get museum statistics (public)"

test_endpoint "GET" "$BASE_URL/museum/tags" "" "" \
    "Get museum tags (public)"

test_endpoint "GET" "$BASE_URL/museum/popular" "" "" \
    "Get popular entries (public)"

# 认证用户API测试
echo -e "\n${YELLOW}=== Testing Authenticated Museum APIs ===${NC}"

# 创建测试数据
LETTER_ID="test-letter-$(date +%s)"
SUBMIT_DATA=$(cat <<EOF
{
    "letter_id": "$LETTER_ID",
    "title": "Test Museum Submission",
    "author_name": "Test Author",
    "tags": ["test", "museum", "api"],
    "display_preference": "anonymous",
    "submission_reason": "Testing museum API"
}
EOF
)

test_endpoint "POST" "$BASE_URL/museum/submit" "$SUBMIT_DATA" "$STUDENT_TOKEN" \
    "Submit letter to museum"

test_endpoint "GET" "$BASE_URL/museum/my-submissions" "" "$STUDENT_TOKEN" \
    "Get my submissions"

# 互动测试
INTERACTION_DATA='{"type": "view"}'
test_endpoint "POST" "$BASE_URL/museum/entries/1/interact" "$INTERACTION_DATA" "$STUDENT_TOKEN" \
    "Record interaction (view)"

REACTION_DATA='{"reaction_type": "like", "comment": "Great letter!"}'
test_endpoint "POST" "$BASE_URL/museum/entries/1/react" "$REACTION_DATA" "$STUDENT_TOKEN" \
    "Add reaction"

# 搜索测试
SEARCH_DATA='{"query": "test", "tags": ["test"], "limit": 10}'
test_endpoint "POST" "$BASE_URL/museum/search" "$SEARCH_DATA" "$STUDENT_TOKEN" \
    "Search museum entries"

# 管理员API测试
echo -e "\n${YELLOW}=== Testing Admin Museum APIs ===${NC}"

test_endpoint "GET" "$ADMIN_BASE_URL/museum/entries/pending" "" "$ADMIN_TOKEN" \
    "Get pending entries (admin)"

# 创建展览
EXHIBITION_DATA=$(cat <<EOF
{
    "title": "Test Exhibition",
    "description": "A test exhibition for API testing",
    "theme_keywords": ["test", "api", "exhibition"],
    "curator_name": "Test Curator",
    "start_date": "2024-01-01T00:00:00Z"
}
EOF
)

test_endpoint "POST" "$ADMIN_BASE_URL/museum/exhibitions" "$EXHIBITION_DATA" "$ADMIN_TOKEN" \
    "Create exhibition (admin)"

test_endpoint "GET" "$ADMIN_BASE_URL/museum/analytics" "" "$ADMIN_TOKEN" \
    "Get museum analytics (admin)"

# 统计测试结果
echo -e "\n${YELLOW}=== Test Summary ===${NC}"
echo "Test completed at: $(date)"
echo ""
echo "Note: Some tests may fail if:"
echo "1. Database tables are not properly migrated"
echo "2. Test data does not exist"
echo "3. Services are not running"
echo ""
echo "Please check the responses above for details."