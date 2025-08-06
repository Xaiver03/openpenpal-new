#!/bin/bash

# OpenPenPal API 测试脚本
# 用于测试各个服务的 API 接口

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 基础 URL
BACKEND_URL="http://localhost:8080"
WRITE_SERVICE_URL="http://localhost:8001"
COURIER_SERVICE_URL="http://localhost:8002"
ADMIN_SERVICE_URL="http://localhost:8003"
OCR_SERVICE_URL="http://localhost:8004"

echo "🧪 OpenPenPal API 测试开始..."
echo "========================================"

# 测试后端主服务
echo -e "\n${YELLOW}测试后端主服务 (${BACKEND_URL})${NC}"
echo "----------------------------------------"

# 健康检查
echo -n "健康检查: "
if curl -s "${BACKEND_URL}/api/health" > /dev/null 2>&1; then
    echo -e "${GREEN}✓ 成功${NC}"
else
    echo -e "${RED}✗ 失败${NC}"
fi

# 测试写信服务
echo -e "\n${YELLOW}测试写信服务 (${WRITE_SERVICE_URL})${NC}"
echo "----------------------------------------"

echo -n "健康检查: "
if curl -s "${WRITE_SERVICE_URL}/health" > /dev/null 2>&1; then
    echo -e "${GREEN}✓ 成功${NC}"
else
    echo -e "${RED}✗ 失败${NC}"
fi

# 测试信使服务
echo -e "\n${YELLOW}测试信使服务 (${COURIER_SERVICE_URL})${NC}"
echo "----------------------------------------"

echo -n "健康检查: "
if curl -s "${COURIER_SERVICE_URL}/api/v1/health" > /dev/null 2>&1; then
    echo -e "${GREEN}✓ 成功${NC}"
else
    echo -e "${RED}✗ 失败${NC}"
fi

# 测试管理服务
echo -e "\n${YELLOW}测试管理服务 (${ADMIN_SERVICE_URL})${NC}"
echo "----------------------------------------"

echo -n "健康检查: "
if curl -s "${ADMIN_SERVICE_URL}/api/health" > /dev/null 2>&1; then
    echo -e "${GREEN}✓ 成功${NC}"
else
    echo -e "${RED}✗ 失败${NC}"
fi

# 测试 OCR 服务
echo -e "\n${YELLOW}测试 OCR 服务 (${OCR_SERVICE_URL})${NC}"
echo "----------------------------------------"

echo -n "健康检查: "
if curl -s "${OCR_SERVICE_URL}/health" > /dev/null 2>&1; then
    echo -e "${GREEN}✓ 成功${NC}"
else
    echo -e "${RED}✗ 失败${NC}"
fi

echo -e "\n========================================"
echo "🎯 API 测试完成!"
echo ""
echo "💡 提示："
echo "   - 如果某个服务测试失败，请确保该服务已启动"
echo "   - 可以使用 ./scripts/start.sh 启动所有服务"
echo "   - 查看日志: tail -f logs/*.log"