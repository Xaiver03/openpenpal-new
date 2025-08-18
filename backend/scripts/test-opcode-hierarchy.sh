#!/bin/bash

# Test OP Code Hierarchical APIs

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# API基础URL
BASE_URL="http://localhost:8080/api/v1"

# 登录获取token
echo -e "${YELLOW}1. 登录获取认证令牌...${NC}"
LOGIN_RESPONSE=$(curl -s -X POST "${BASE_URL}/auth/login" \
  -H "Content-Type: application/json" \
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
  -H "Authorization: Bearer $TOKEN" | jq '.'

# 3. 按城市搜索学校（北京）
echo -e "\n${YELLOW}3. 搜索北京的学校...${NC}"
curl -s -X GET "${BASE_URL}/opcode/search/schools/by-city?city=北京&limit=10" \
  -H "Authorization: Bearer $TOKEN" | jq '.'

# 4. 获取北京大学(PK)的片区
echo -e "\n${YELLOW}4. 获取北京大学的片区列表...${NC}"
curl -s -X GET "${BASE_URL}/opcode/districts/PK" \
  -H "Authorization: Bearer $TOKEN" | jq '.'

# 5. 获取北京大学东区的楼栋
echo -e "\n${YELLOW}5. 获取北京大学东区(1)的楼栋列表...${NC}"
curl -s -X GET "${BASE_URL}/opcode/buildings/PK/1" \
  -H "Authorization: Bearer $TOKEN" | jq '.'

# 6. 获取投递点列表（PK1A前缀）
echo -e "\n${YELLOW}6. 获取PK1A开头的投递点列表...${NC}"
curl -s -X GET "${BASE_URL}/opcode/delivery-points/PK1A" \
  -H "Authorization: Bearer $TOKEN" | jq '.'

# 7. 申请一个OP Code
echo -e "\n${YELLOW}7. 申请OP Code PK1A03...${NC}"
curl -s -X POST "${BASE_URL}/opcode/apply" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "code": "PK1A03",
    "type": "dormitory",
    "description": "北京大学 东区 A栋 103室"
  }' | jq '.'

# 8. 验证OP Code
echo -e "\n${YELLOW}8. 验证OP Code PK1A01...${NC}"
curl -s -X GET "${BASE_URL}/opcode/validate?code=PK1A01" \
  -H "Authorization: Bearer $TOKEN" | jq '.'

# 9. 搜索已存在的OP Code
echo -e "\n${YELLOW}9. 搜索学校代码为PK的OP Code...${NC}"
curl -s -X GET "${BASE_URL}/opcode/search?school_code=PK&page=1&page_size=5" \
  -H "Authorization: Bearer $TOKEN" | jq '.'

echo -e "\n${GREEN}测试完成！${NC}"