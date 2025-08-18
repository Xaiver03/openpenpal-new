#!/bin/bash

# Test OP Code Hierarchical APIs v2

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# API基础URL
BASE_URL="http://localhost:8080/api/v1"

# 先获取CSRF Token
echo -e "${YELLOW}0. 获取CSRF Token...${NC}"
CSRF_RESPONSE=$(curl -s "${BASE_URL}/auth/csrf")
CSRF_TOKEN=$(echo $CSRF_RESPONSE | jq -r '.data.token // empty')

if [ -z "$CSRF_TOKEN" ]; then
  echo -e "${RED}获取CSRF Token失败！${NC}"
  echo $CSRF_RESPONSE | jq '.'
  exit 1
fi

echo -e "${GREEN}获取CSRF Token成功！${NC}"

# 登录获取token
echo -e "\n${YELLOW}1. 登录获取认证令牌...${NC}"
LOGIN_RESPONSE=$(curl -s -X POST "${BASE_URL}/auth/login" \
  -H "Content-Type: application/json" \
  -H "X-CSRF-Token: $CSRF_TOKEN" \
  -d '{
    "username": "alice",
    "password": "secret123"
  }')

TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.data.token // empty')

if [ -z "$TOKEN" ]; then
  echo -e "${RED}登录失败！${NC}"
  echo $LOGIN_RESPONSE | jq '.'
  exit 1
fi

echo -e "${GREEN}登录成功！获取到TOKEN${NC}"

# 2. 获取城市列表
echo -e "\n${YELLOW}2. 获取城市列表...${NC}"
curl -s -X GET "${BASE_URL}/opcode/cities" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-CSRF-Token: $CSRF_TOKEN" | jq '.'

# 3. 按城市搜索学校（北京）- 使用public API（如果需要）
echo -e "\n${YELLOW}3. 搜索北京的学校（使用alias API）...${NC}"
curl -s -X GET "${BASE_URL}/opcode/search/schools/by-city?city=北京&limit=10" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-CSRF-Token: $CSRF_TOKEN" | jq '.'

# 4. 获取北京大学(PK)的片区
echo -e "\n${YELLOW}4. 获取北京大学的片区列表...${NC}"
curl -s -X GET "${BASE_URL}/opcode/districts/PK" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-CSRF-Token: $CSRF_TOKEN" | jq '.'

# 5. 获取北京大学东区的楼栋
echo -e "\n${YELLOW}5. 获取北京大学东区(1)的楼栋列表...${NC}"
curl -s -X GET "${BASE_URL}/opcode/buildings/PK/1" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-CSRF-Token: $CSRF_TOKEN" | jq '.'

# 6. 获取投递点列表（PK1A前缀）
echo -e "\n${YELLOW}6. 获取PK1A开头的投递点列表...${NC}"
curl -s -X GET "${BASE_URL}/opcode/delivery-points/PK1A" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-CSRF-Token: $CSRF_TOKEN" | jq '.'

# 7. 申请一个OP Code
echo -e "\n${YELLOW}7. 申请OP Code PK1A03...${NC}"
curl -s -X POST "${BASE_URL}/opcode/apply" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-CSRF-Token: $CSRF_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "code": "PK1A03",
    "type": "dormitory",
    "description": "北京大学 东区 A栋 103室"
  }' | jq '.'

# 8. 验证OP Code
echo -e "\n${YELLOW}8. 验证OP Code PK1A01...${NC}"
curl -s -X GET "${BASE_URL}/opcode/validate?code=PK1A01" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-CSRF-Token: $CSRF_TOKEN" | jq '.'

# 9. 测试搜索片区API
echo -e "\n${YELLOW}9. 测试搜索片区API (SearchAreas)...${NC}"
curl -s -X GET "${BASE_URL}/opcode/search/areas?schoolCode=PK" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-CSRF-Token: $CSRF_TOKEN" | jq '.'

# 10. 测试搜索楼栋API
echo -e "\n${YELLOW}10. 测试搜索楼栋API (SearchBuildings)...${NC}"
curl -s -X GET "${BASE_URL}/opcode/search/buildings?schoolCode=PK&areaCode=1" \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-CSRF-Token: $CSRF_TOKEN" | jq '.'

echo -e "\n${GREEN}测试完成！${NC}"