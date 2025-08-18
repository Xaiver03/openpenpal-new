#!/bin/bash

# OP Code API Test with proper CSRF handling

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# API基础URL
BASE_URL="http://localhost:8080/api/v1"
COOKIE_JAR="/tmp/opcode_test_cookies.txt"

# 清理旧的cookie文件
rm -f "$COOKIE_JAR"

# 1. 获取CSRF Token
echo -e "${YELLOW}1. 获取CSRF Token...${NC}"
CSRF_RESPONSE=$(curl -s -c "$COOKIE_JAR" -b "$COOKIE_JAR" "${BASE_URL}/auth/csrf")
echo "$CSRF_RESPONSE" | jq '.'

CSRF_TOKEN=$(echo "$CSRF_RESPONSE" | jq -r '.data.token // empty')

if [ -z "$CSRF_TOKEN" ]; then
  echo -e "${RED}获取CSRF Token失败！${NC}"
  exit 1
fi

# 2. 登录（使用CSRF Token）
echo -e "\n${YELLOW}2. 登录...${NC}"
LOGIN_RESP=$(curl -s -c "$COOKIE_JAR" -b "$COOKIE_JAR" \
  -X POST "${BASE_URL}/auth/login" \
  -H "Content-Type: application/json" \
  -H "X-CSRF-Token: $CSRF_TOKEN" \
  -d '{
    "username": "alice",
    "password": "secret123"
  }')

echo "$LOGIN_RESP" | jq '.'

TOKEN=$(echo "$LOGIN_RESP" | jq -r '.data.token // empty')

if [ -z "$TOKEN" ]; then
  echo -e "${RED}登录失败！${NC}"
  exit 1
fi

echo -e "${GREEN}登录成功！${NC}"

# 3. 测试OP Code层级API
echo -e "\n${YELLOW}3. 获取城市列表...${NC}"
curl -s -b "$COOKIE_JAR" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-CSRF-Token: $CSRF_TOKEN" \
  "${BASE_URL}/opcode/cities" | jq '.'

echo -e "\n${YELLOW}4. 搜索北京的学校...${NC}"
curl -s -b "$COOKIE_JAR" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-CSRF-Token: $CSRF_TOKEN" \
  "${BASE_URL}/opcode/search/schools/by-city?city=北京&limit=5" | jq '.'

echo -e "\n${YELLOW}5. 获取北京大学(PK)的片区...${NC}"
curl -s -b "$COOKIE_JAR" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-CSRF-Token: $CSRF_TOKEN" \
  "${BASE_URL}/opcode/districts/PK" | jq '.'

echo -e "\n${YELLOW}6. 获取北京大学东区(1)的楼栋...${NC}"
curl -s -b "$COOKIE_JAR" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-CSRF-Token: $CSRF_TOKEN" \
  "${BASE_URL}/opcode/buildings/PK/1" | jq '.'

echo -e "\n${YELLOW}7. 获取投递点列表（PK1A前缀）...${NC}"
curl -s -b "$COOKIE_JAR" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-CSRF-Token: $CSRF_TOKEN" \
  "${BASE_URL}/opcode/delivery-points/PK1A" | jq '.'

echo -e "\n${YELLOW}8. 申请OP Code PK1A05...${NC}"
curl -s -b "$COOKIE_JAR" \
  -X POST "${BASE_URL}/opcode/apply" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-CSRF-Token: $CSRF_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "code": "PK1A05",
    "type": "dormitory",
    "description": "北京大学 东区 A栋 105室"
  }' | jq '.'

# 清理
rm -f "$COOKIE_JAR"

echo -e "\n${GREEN}测试完成！${NC}"