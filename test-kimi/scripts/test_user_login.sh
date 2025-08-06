#!/bin/bash

# OpenPenPal 用户登录测试脚本
# 测试刚才注册的用户是否能正常登录

echo "🔐 开始OpenPenPal用户登录测试..."
echo "========================================"

# API基础URL
BASE_URL="http://localhost:8080"
LOGIN_URL="${BASE_URL}/api/v1/auth/login"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 统计变量
SUCCESS_COUNT=0
FAILURE_COUNT=0

# 函数：测试登录
test_login() {
    local username="$1"
    local password="$2"
    
    echo -e "${BLUE}🔐 测试登录: $username${NC}"
    
    # 构建JSON数据
    json_data=$(cat <<EOF
{
  "username": "$username",
  "password": "$password"
}
EOF
)
    
    # 发送HTTP请求
    response=$(curl -s -w "\n%{http_code}" \
        -X POST \
        -H "Content-Type: application/json" \
        -d "$json_data" \
        "$LOGIN_URL" 2>&1)
    
    # 分离响应体和状态码
    response_body=$(echo "$response" | sed '$d')
    http_code=$(echo "$response" | tail -n 1)
    
    # 检查结果
    if [[ "$http_code" == "200" ]]; then
        echo -e "   ✅ ${GREEN}登录成功${NC} (HTTP $http_code)"
        # 尝试提取token
        token=$(echo "$response_body" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
        if [[ -n "$token" ]]; then
            echo -e "   🔑 Token: ${token:0:20}..."
        fi
        SUCCESS_COUNT=$((SUCCESS_COUNT + 1))
    else
        echo -e "   ❌ ${RED}登录失败${NC} (HTTP $http_code)"
        echo -e "   📄 错误: $response_body"
        FAILURE_COUNT=$((FAILURE_COUNT + 1))
    fi
    
    echo ""
}

# 测试登录用户
declare -a test_logins=(
    "testuser02:password123"
    "testuser03:password123"
    "testuser04:password123"
    "testuser05:password123"
    "testuser06:password123"
)

echo "📋 测试配置:"
echo "   • 登录API: $LOGIN_URL"
echo "   • 测试用户数: ${#test_logins[@]}"
echo ""

echo -e "${YELLOW}🔐 开始登录测试...${NC}"
echo "========================================"

# 执行登录测试
for login_data in "${test_logins[@]}"; do
    IFS=':' read -r username password <<< "$login_data"
    test_login "$username" "$password"
done

# 测试错误登录
echo -e "${YELLOW}🚫 测试错误登录...${NC}"
echo "========================================"

echo -e "${BLUE}🔐 测试错误密码${NC}"
test_login "testuser02" "wrongpassword"

echo -e "${BLUE}🔐 测试不存在的用户${NC}"
test_login "nonexistent" "password123"

echo "========================================"
echo -e "${YELLOW}📊 登录测试结果统计${NC}"
echo "========================================"
echo -e "成功登录:   ${GREEN}$SUCCESS_COUNT${NC}"
echo -e "登录失败:   ${RED}$FAILURE_COUNT${NC}"

echo ""
echo -e "${YELLOW}🏁 登录测试完成！${NC}"

if [[ $SUCCESS_COUNT -gt 0 ]]; then
    echo -e "${GREEN}✅ 用户注册和登录功能正常工作！${NC}"
else
    echo -e "${RED}❌ 登录功能存在问题${NC}"
fi