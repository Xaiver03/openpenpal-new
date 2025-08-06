#!/bin/bash

# 迁移到统一Mock服务脚本
# 停止旧的临时mock服务，启用新的完整mock服务系统

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}🔄 迁移到统一Mock服务系统${NC}"
echo "========================================"

# 1. 停止旧的临时服务
echo -e "${YELLOW}📱 停止旧的临时Mock服务...${NC}"

# 停止可能运行的mock-services.js
pkill -f "node mock-services.js" 2>/dev/null || true

# 停止start-integration.sh启动的服务
pkill -f "start-integration.sh" 2>/dev/null || true

# 清理临时文件
rm -f mock-services.js 2>/dev/null || true
rm -f logs/gateway.pid 2>/dev/null || true

echo -e "${GREEN}✅ 旧服务已停止${NC}"

# 2. 启动新的完整Mock服务
echo -e "${YELLOW}🚀 启动新的完整Mock服务系统...${NC}"

# 确保依赖已安装
if [ ! -d "apps/mock-services/node_modules" ]; then
    echo -e "${BLUE}📦 安装Mock服务依赖...${NC}"
    cd apps/mock-services
    npm install
    cd ../..
fi

# 启动新的Mock服务
./scripts/start-mock.sh &

# 等待服务启动
echo -e "${BLUE}⏳ 等待服务启动...${NC}"
sleep 5

# 3. 验证服务状态
echo -e "${YELLOW}🔍 验证服务状态...${NC}"

check_service() {
    local port=$1
    local service=$2
    
    if curl -s "http://localhost:$port/health" > /dev/null 2>&1; then
        echo -e "${GREEN}✅ $service (端口 $port) - 运行正常${NC}"
        return 0
    else
        echo -e "${RED}❌ $service (端口 $port) - 启动失败${NC}"
        return 1
    fi
}

# 检查各个服务
check_service 8000 "API Gateway"
check_service 8001 "Write Service" 
check_service 8002 "Courier Service"
check_service 8003 "Admin Service"
check_service 8004 "OCR Service"

# 4. 更新前端配置建议
echo ""
echo -e "${BLUE}📝 前端配置建议:${NC}"
echo "========================================"
echo "更新前端环境变量:"
echo "VITE_API_BASE_URL=http://localhost:8000/api"
echo ""
echo "或在前端代码中使用:"
echo "const API_BASE = 'http://localhost:8000/api'"
echo ""

# 5. 显示新服务的API文档
echo -e "${BLUE}📚 新Mock服务API文档:${NC}"
echo "========================================"
echo "认证API:"
echo "  POST /api/auth/login       - 用户登录"
echo "  POST /api/auth/register    - 用户注册"  
echo "  GET  /api/auth/me          - 获取当前用户"
echo ""
echo "写信服务API:"
echo "  GET  /api/write/letters    - 获取信件列表"
echo "  POST /api/write/letters    - 创建新信件"
echo "  GET  /api/write/letters/:id - 获取信件详情"
echo ""
echo "信使服务API:"
echo "  GET  /api/courier/tasks    - 获取可用任务"
echo "  POST /api/courier/tasks/:id/accept - 接受任务"
echo ""
echo "管理服务API:"
echo "  GET  /api/admin/users      - 获取用户列表"
echo "  GET  /api/admin/system/config - 获取系统配置"
echo ""

# 6. 测试建议
echo -e "${BLUE}🧪 测试建议:${NC}"
echo "========================================"
echo "运行集成测试:"
echo "  ./scripts/test-mock-integration.sh"
echo ""
echo "测试用户账号:"
echo "  alice/secret      - 学生用户"
echo "  admin/admin123    - 管理员用户"
echo "  courier1/courier123 - 信使用户"
echo ""

echo -e "${GREEN}🎉 迁移完成！${NC}"
echo "新的Mock服务系统已启动，具有更完整的功能和更好的维护性。"