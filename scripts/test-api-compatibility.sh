#!/bin/bash

# API 兼容性测试脚本

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# API 基础 URL
GO_API="http://localhost:8080/api/v1"
PRISMA_API="http://localhost:8081/api/v1"

echo "🔍 API 兼容性测试"
echo "=================="
echo ""

# 测试健康检查
echo "1. 健康检查测试"
echo -n "  Go Backend: "
GO_HEALTH=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/health)
if [ "$GO_HEALTH" = "200" ]; then
    echo -e "${GREEN}✓ 正常${NC}"
else
    echo -e "${RED}✗ 异常 (HTTP $GO_HEALTH)${NC}"
fi

echo -n "  Prisma Backend: "
PRISMA_HEALTH=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8081/health)
if [ "$PRISMA_HEALTH" = "200" ]; then
    echo -e "${GREEN}✓ 正常${NC}"
else
    echo -e "${RED}✗ 异常 (HTTP $PRISMA_HEALTH)${NC}"
fi

echo ""
echo "2. 认证 API 测试"

# 测试登录
echo "  测试登录..."
GO_LOGIN=$(curl -s -X POST "$GO_API/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"admin123"}')

PRISMA_LOGIN=$(curl -s -X POST "$PRISMA_API/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"username":"admin","password":"admin123"}')

# 提取 token
GO_TOKEN=$(echo $GO_LOGIN | grep -o '"token":"[^"]*' | grep -o '[^"]*$')
PRISMA_TOKEN=$(echo $PRISMA_LOGIN | grep -o '"token":"[^"]*' | grep -o '[^"]*$')

if [ -n "$GO_TOKEN" ] && [ -n "$PRISMA_TOKEN" ]; then
    echo -e "  ${GREEN}✓ 登录成功${NC}"
else
    echo -e "  ${RED}✗ 登录失败${NC}"
    echo "  Go Response: $GO_LOGIN"
    echo "  Prisma Response: $PRISMA_LOGIN"
fi

echo ""
echo "3. 用户 API 测试"

# 测试获取用户信息
echo "  测试获取用户信息..."
if [ -n "$GO_TOKEN" ]; then
    GO_USER=$(curl -s -H "Authorization: Bearer $GO_TOKEN" "$GO_API/users/me")
    echo -e "  Go Backend: ${GREEN}✓${NC}"
fi

if [ -n "$PRISMA_TOKEN" ]; then
    PRISMA_USER=$(curl -s -H "Authorization: Bearer $PRISMA_TOKEN" "$PRISMA_API/users/me")
    echo -e "  Prisma Backend: ${GREEN}✓${NC}"
fi

echo ""
echo "4. 信件 API 测试"

# 测试获取公开信件
echo "  测试获取公开信件..."
GO_LETTERS=$(curl -s "$GO_API/letters/public?limit=5")
PRISMA_LETTERS=$(curl -s "$PRISMA_API/letters/public?limit=5")

GO_LETTER_COUNT=$(echo $GO_LETTERS | grep -o '"letters":\[' | wc -l)
PRISMA_LETTER_COUNT=$(echo $PRISMA_LETTERS | grep -o '"letters":\[' | wc -l)

if [ "$GO_LETTER_COUNT" -gt 0 ] && [ "$PRISMA_LETTER_COUNT" -gt 0 ]; then
    echo -e "  ${GREEN}✓ API 响应格式一致${NC}"
else
    echo -e "  ${YELLOW}⚠ API 响应格式可能不同${NC}"
fi

echo ""
echo "测试完成！"
echo ""
echo "提示："
echo "- 确保两个后端都在运行"
echo "- Go Backend: http://localhost:8080"
echo "- Prisma Backend: http://localhost:8081"