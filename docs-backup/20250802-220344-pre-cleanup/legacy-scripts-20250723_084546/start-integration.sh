#!/bin/bash

# OpenPenPal Frontend-Backend Integration Startup Script
# OpenPenPal前后端集成启动脚本

set -e

echo "🚀 Starting OpenPenPal Frontend-Backend Integration..."

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 检查Node.js是否已安装
check_node() {
    if ! command -v node &> /dev/null; then
        echo -e "${RED}❌ Node.js is not installed. Please install Node.js 16+ first.${NC}"
        exit 1
    fi
    
    NODE_VERSION=$(node --version | cut -d'v' -f2)
    echo -e "${GREEN}✅ Node.js version: $NODE_VERSION${NC}"
}

# 检查端口是否被占用
check_port() {
    local port=$1
    local service=$2
    
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
        echo -e "${YELLOW}⚠️  Port $port is already in use. Killing existing process...${NC}"
        lsof -ti:$port | xargs kill -9 2>/dev/null || true
        sleep 2
    fi
    echo -e "${GREEN}✅ Port $port is available for $service${NC}"
}

# 安装网关依赖
install_gateway_deps() {
    echo -e "${BLUE}📦 Installing API Gateway dependencies...${NC}"
    
    if [ ! -f "frontend/package-gateway.json" ]; then
        cp frontend/gateway-package.json frontend/package-gateway.json
    fi
    
    cd frontend
    npm install --prefix . --package-lock-only=false \
        express \
        http-proxy-middleware \
        cors \
        helmet \
        express-rate-limit \
        jsonwebtoken \
        ws \
        dotenv \
        nodemon
    cd ..
    
    echo -e "${GREEN}✅ Gateway dependencies installed${NC}"
}

# 启动API网关
start_gateway() {
    echo -e "${BLUE}🌐 Starting API Gateway on port 8000...${NC}"
    
    export GATEWAY_PORT=8000
    export JWT_SECRET="openpenpal-super-secret-jwt-key-for-integration"
    export FRONTEND_URL="http://localhost:3000"
    export WRITE_SERVICE_URL="http://localhost:8001"
    export COURIER_SERVICE_URL="http://localhost:8002"
    export ADMIN_SERVICE_URL="http://localhost:8003"
    export OCR_SERVICE_URL="http://localhost:8004"
    
    cd frontend
    nohup node api-gateway.config.js > ../logs/gateway.log 2>&1 &
    echo $! > ../logs/gateway.pid
    cd ..
    
    echo -e "${GREEN}✅ API Gateway started (PID: $(cat logs/gateway.pid))${NC}"
}

# 启动模拟后端服务
start_mock_services() {
    echo -e "${BLUE}🔧 Starting mock backend services...${NC}"
    
    # 创建模拟服务脚本
    cat > mock-services.js << 'EOF'
const express = require('express');
const cors = require('cors');

// 写信服务 (8001)
const writeApp = express();
writeApp.use(cors());
writeApp.use(express.json());

writeApp.post('/auth/login', (req, res) => {
    res.json({
        success: true,
        data: {
            token: 'mock-jwt-token',
            user: {
                id: 'test-user-1',
                username: 'testuser',
                email: 'test@example.com',
                nickname: '测试用户',
                role: 'user',
                school_code: 'BJDX01',
                school_name: '北京大学',
                permissions: ['read', 'write']
            }
        }
    });
});

writeApp.post('/auth/register', (req, res) => {
    res.json({ success: true, message: '注册成功' });
});

writeApp.get('/auth/me', (req, res) => {
    res.json({
        success: true,
        data: {
            id: 'test-user-1',
            username: 'testuser',
            email: 'test@example.com',
            nickname: '测试用户',
            role: 'user',
            school_code: 'BJDX01',
            school_name: '北京大学'
        }
    });
});

writeApp.get('/schools/search', (req, res) => {
    res.json({
        success: true,
        data: {
            items: [
                {
                    id: '1',
                    code: 'BJDX01',
                    name: '北京大学',
                    province: '北京',
                    city: '北京',
                    type: 'university',
                    status: 'active'
                },
                {
                    id: '2',
                    code: 'THU001',
                    name: '清华大学',
                    province: '北京',
                    city: '北京',
                    type: 'university',
                    status: 'active'
                }
            ],
            total: 2
        }
    });
});

writeApp.get('/schools/provinces', (req, res) => {
    res.json({
        success: true,
        data: ['北京', '上海', '广东', '江苏', '浙江']
    });
});

writeApp.listen(8001, () => console.log('📝 Write Service running on port 8001'));

// 信使服务 (8002)
const courierApp = express();
courierApp.use(cors());
courierApp.use(express.json());

courierApp.get('/courier/info', (req, res) => {
    res.json({
        success: true,
        data: {
            id: 'courier-1',
            level: 1,
            region: '北京大学',
            total_points: 150,
            completed_tasks: 5
        }
    });
});

courierApp.listen(8002, () => console.log('🏃‍♂️ Courier Service running on port 8002'));

// 管理服务 (8003)
const adminApp = express();
adminApp.use(cors());
adminApp.use(express.json());

adminApp.get('/api/admin/dashboard', (req, res) => {
    res.json({
        success: true,
        data: {
            users: { total: 100, active: 85 },
            letters: { total: 500, sent_today: 25 },
            couriers: { total: 20, active: 15 }
        }
    });
});

adminApp.listen(8003, () => console.log('👨‍💼 Admin Service running on port 8003'));

// OCR服务 (8004)
const ocrApp = express();
ocrApp.use(cors());
ocrApp.use(express.json());

ocrApp.post('/ocr/process', (req, res) => {
    res.json({
        success: true,
        data: {
            text: '这是OCR识别的文字内容',
            confidence: 0.95
        }
    });
});

ocrApp.listen(8004, () => console.log('🔍 OCR Service running on port 8004'));
EOF

    # 启动模拟服务
    nohup node mock-services.js > logs/mock-services.log 2>&1 &
    echo $! > logs/mock-services.pid
    
    echo -e "${GREEN}✅ Mock services started (PID: $(cat logs/mock-services.pid))${NC}"
}

# 启动前端
start_frontend() {
    echo -e "${BLUE}🎨 Starting Frontend on port 3000...${NC}"
    
    cd frontend
    
    # 检查是否需要安装依赖
    if [ ! -d "node_modules" ]; then
        echo -e "${YELLOW}📦 Installing frontend dependencies...${NC}"
        npm install
    fi
    
    nohup npm run dev > ../logs/frontend.log 2>&1 &
    echo $! > ../logs/frontend.pid
    cd ..
    
    echo -e "${GREEN}✅ Frontend started (PID: $(cat logs/frontend.pid))${NC}"
}

# 等待服务启动
wait_for_services() {
    echo -e "${BLUE}⏳ Waiting for services to start...${NC}"
    
    # 等待网关启动
    for i in {1..30}; do
        if curl -s http://localhost:8000/api/v1/health > /dev/null 2>&1; then
            echo -e "${GREEN}✅ API Gateway is ready${NC}"
            break
        fi
        echo -e "${YELLOW}⏳ Waiting for API Gateway... ($i/30)${NC}"
        sleep 2
    done
    
    # 等待前端启动
    for i in {1..30}; do
        if curl -s http://localhost:3000 > /dev/null 2>&1; then
            echo -e "${GREEN}✅ Frontend is ready${NC}"
            break
        fi
        echo -e "${YELLOW}⏳ Waiting for Frontend... ($i/30)${NC}"
        sleep 2
    done
}

# 显示状态
show_status() {
    echo ""
    echo -e "${GREEN}🎉 OpenPenPal Frontend-Backend Integration is running!${NC}"
    echo ""
    echo -e "${BLUE}📋 Service URLs:${NC}"
    echo -e "   🌐 Frontend:     http://localhost:3000"
    echo -e "   🚪 API Gateway:  http://localhost:8000"
    echo -e "   📝 Write Service: http://localhost:8001"
    echo -e "   🏃‍♂️ Courier Service: http://localhost:8002"  
    echo -e "   👨‍💼 Admin Service: http://localhost:8003"
    echo -e "   🔍 OCR Service:   http://localhost:8004"
    echo ""
    echo -e "${BLUE}📊 Test Accounts:${NC}"
    echo -e "   Username: testuser"
    echo -e "   Password: secret"
    echo ""
    echo -e "${BLUE}🔧 Management Commands:${NC}"
    echo -e "   Stop all:    ./stop-integration.sh"
    echo -e "   View logs:   tail -f logs/*.log"
    echo -e "   Check status: curl http://localhost:8000/api/v1/health"
    echo ""
}

# 主函数
main() {
    # 创建日志目录
    mkdir -p logs
    
    echo -e "${BLUE}🔍 Checking system requirements...${NC}"
    check_node
    
    echo -e "${BLUE}🔧 Checking ports availability...${NC}"
    check_port 3000 "Frontend"
    check_port 8000 "API Gateway"
    check_port 8001 "Write Service"
    check_port 8002 "Courier Service"
    check_port 8003 "Admin Service"
    check_port 8004 "OCR Service"
    
    # 安装依赖
    install_gateway_deps
    
    # 启动服务
    start_mock_services
    sleep 3
    
    start_gateway
    sleep 3
    
    start_frontend
    
    # 等待服务就绪
    wait_for_services
    
    # 显示状态
    show_status
}

# 执行主函数
main

echo -e "${GREEN}✅ Integration startup completed! Press Ctrl+C to stop all services.${NC}"

# 等待中断信号
trap 'echo -e "\n${YELLOW}🛑 Shutting down services...${NC}"; ./stop-integration.sh; exit 0' INT

# 保持脚本运行
while true; do
    sleep 10
done