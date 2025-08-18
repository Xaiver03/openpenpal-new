#!/bin/bash

# Simple OP Code API Test

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# API基础URL
BASE_URL="http://localhost:8080/api/v1"

# 使用现有的测试账号获取token（假设已有有效token）
# 或者临时禁用CSRF验证进行测试

echo -e "${YELLOW}测试公开API（不需要认证）...${NC}"

# 1. 测试学校搜索（如果有公开endpoint）
echo -e "\n${YELLOW}1. 测试学校搜索...${NC}"
curl -s -X GET "${BASE_URL}/schools?search=北京" | jq '.'

# 如果需要认证，可以使用之前获取的有效token
# TOKEN="your_valid_token_here"

# 或者使用cookie jar方式处理CSRF
echo -e "\n${YELLOW}使用cookie jar方式登录...${NC}"

# 创建cookie jar
COOKIE_JAR="/tmp/opcode_test_cookies.txt"

# 1. 获取CSRF token并保存cookie
echo -e "\n${YELLOW}1. 获取CSRF Token...${NC}"
curl -s -c "$COOKIE_JAR" -b "$COOKIE_JAR" \
  "${BASE_URL}/auth/csrf" | jq '.'

# 2. 登录
echo -e "\n${YELLOW}2. 登录...${NC}"
LOGIN_RESP=$(curl -s -c "$COOKIE_JAR" -b "$COOKIE_JAR" \
  -X POST "${BASE_URL}/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "alice",
    "password": "secret123"
  }')

echo "$LOGIN_RESP" | jq '.'

TOKEN=$(echo "$LOGIN_RESP" | jq -r '.data.token // empty')

if [ -n "$TOKEN" ]; then
  echo -e "${GREEN}登录成功！${NC}"
  
  # 3. 测试OP Code API
  echo -e "\n${YELLOW}3. 获取城市列表...${NC}"
  curl -s -b "$COOKIE_JAR" \
    -H "Authorization: Bearer $TOKEN" \
    "${BASE_URL}/opcode/cities" | jq '.'
  
  echo -e "\n${YELLOW}4. 搜索北京的学校...${NC}"
  curl -s -b "$COOKIE_JAR" \
    -H "Authorization: Bearer $TOKEN" \
    "${BASE_URL}/opcode/search/schools/by-city?city=北京&limit=5" | jq '.'
fi

# 清理
rm -f "$COOKIE_JAR"

echo -e "\n${GREEN}测试完成！${NC}"