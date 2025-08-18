#!/bin/bash

# Test OP Code Complete Data Flow

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# API基础URL
BASE_URL="http://localhost:8080/api/v1"

echo -e "${YELLOW}OP Code层级选择系统数据流测试${NC}"
echo -e "${YELLOW}================================${NC}\n"

# 1. 测试城市列表API（模拟前端加载城市）
echo -e "${YELLOW}1. 测试获取城市列表...${NC}"
curl -s -X GET "${BASE_URL}/opcode/cities" \
  -H "Authorization: Bearer test" | jq '.'

# 2. 测试按城市搜索学校（模拟用户选择北京）
echo -e "\n${YELLOW}2. 测试搜索北京的学校...${NC}"
curl -s -X GET "${BASE_URL}/opcode/search/schools/by-city?city=北京&limit=5" \
  -H "Authorization: Bearer test" | jq '.'

# 3. 测试获取学校片区（模拟用户选择北京大学PK）
echo -e "\n${YELLOW}3. 测试获取北京大学(PK)的片区...${NC}"
curl -s -X GET "${BASE_URL}/opcode/districts/PK" \
  -H "Authorization: Bearer test" | jq '.'

# 4. 测试获取楼栋列表（模拟用户选择东区1）
echo -e "\n${YELLOW}4. 测试获取东区(1)的楼栋...${NC}"
curl -s -X GET "${BASE_URL}/opcode/buildings/PK/1" \
  -H "Authorization: Bearer test" | jq '.'

# 5. 测试获取投递点（模拟用户选择A栋）
echo -e "\n${YELLOW}5. 测试获取投递点列表(PK1A)...${NC}"
curl -s -X GET "${BASE_URL}/opcode/delivery-points/PK1A" \
  -H "Authorization: Bearer test" | jq '.'

# 6. 验证数据库中的数据
echo -e "\n${YELLOW}6. 验证数据库数据...${NC}"
echo -e "${GREEN}检查op_code_schools表:${NC}"
psql -h localhost -d openpenpal -U $(whoami) -c "SELECT school_code, school_name, city FROM op_code_schools WHERE school_code IN ('PK', 'QH', 'FD') LIMIT 5;"

echo -e "\n${GREEN}检查op_code_areas表:${NC}"
psql -h localhost -d openpenpal -U $(whoami) -c "SELECT school_code, area_code, area_name FROM op_code_areas WHERE school_code = 'PK';"

echo -e "\n${GREEN}检查signal_codes表中的OP Code:${NC}"
psql -h localhost -d openpenpal -U $(whoami) -c "SELECT code, description, is_active FROM signal_codes WHERE code LIKE 'PK%' LIMIT 5;"

echo -e "\n${GREEN}测试完成！${NC}"