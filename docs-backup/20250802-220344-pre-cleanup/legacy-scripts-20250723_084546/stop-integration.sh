#!/bin/bash

# OpenPenPal Frontend-Backend Integration Stop Script
# OpenPenPal前后端集成停止脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🛑 Stopping OpenPenPal Integration Services...${NC}"

# 停止服务的函数
stop_service() {
    local service_name=$1
    local pid_file=$2
    
    if [ -f "$pid_file" ]; then
        local pid=$(cat "$pid_file")
        if ps -p $pid > /dev/null 2>&1; then
            echo -e "${YELLOW}🛑 Stopping $service_name (PID: $pid)...${NC}"
            kill -TERM $pid 2>/dev/null || kill -9 $pid 2>/dev/null || true
            
            # 等待进程结束
            for i in {1..10}; do
                if ! ps -p $pid > /dev/null 2>&1; then
                    echo -e "${GREEN}✅ $service_name stopped${NC}"
                    break
                fi
                sleep 1
            done
            
            # 如果进程仍然存在，强制杀死
            if ps -p $pid > /dev/null 2>&1; then
                echo -e "${RED}⚠️  Force killing $service_name...${NC}"
                kill -9 $pid 2>/dev/null || true
            fi
        else
            echo -e "${YELLOW}⚠️  $service_name is not running${NC}"
        fi
        rm -f "$pid_file"
    else
        echo -e "${YELLOW}⚠️  $service_name PID file not found${NC}"
    fi
}

# 停止按端口运行的进程
stop_port() {
    local port=$1
    local service_name=$2
    
    local pids=$(lsof -ti:$port 2>/dev/null || true)
    if [ -n "$pids" ]; then
        echo -e "${YELLOW}🛑 Stopping $service_name on port $port...${NC}"
        echo $pids | xargs kill -TERM 2>/dev/null || true
        sleep 2
        
        # 检查是否还有进程在运行
        local remaining_pids=$(lsof -ti:$port 2>/dev/null || true)
        if [ -n "$remaining_pids" ]; then
            echo -e "${RED}⚠️  Force killing $service_name on port $port...${NC}"
            echo $remaining_pids | xargs kill -9 2>/dev/null || true
        fi
        echo -e "${GREEN}✅ $service_name on port $port stopped${NC}"
    else
        echo -e "${YELLOW}⚠️  No process running on port $port${NC}"
    fi
}

# 创建日志目录（如果不存在）
mkdir -p logs

# 停止各个服务
echo -e "${BLUE}📋 Stopping services...${NC}"

# 停止前端
stop_service "Frontend" "logs/frontend.pid"

# 停止API网关
stop_service "API Gateway" "logs/gateway.pid"

# 停止模拟服务
stop_service "Mock Services" "logs/mock-services.pid"

# 确保所有端口都被释放
echo -e "${BLUE}🔍 Checking for remaining processes on ports...${NC}"
stop_port 3000 "Frontend"
stop_port 8000 "API Gateway"
stop_port 8001 "Write Service"
stop_port 8002 "Courier Service"
stop_port 8003 "Admin Service"
stop_port 8004 "OCR Service"

# 清理临时文件
echo -e "${BLUE}🧹 Cleaning up temporary files...${NC}"
rm -f mock-services.js
rm -f logs/*.pid
rm -f frontend/package-gateway.json

# 显示清理结果
echo ""
echo -e "${GREEN}✅ All services stopped successfully!${NC}"
echo ""
echo -e "${BLUE}📋 Cleanup Summary:${NC}"
echo -e "   ✅ Frontend stopped"
echo -e "   ✅ API Gateway stopped"
echo -e "   ✅ Mock services stopped"
echo -e "   ✅ All ports freed"
echo -e "   ✅ Temporary files cleaned"
echo ""
echo -e "${BLUE}💡 To restart the integration:${NC}"
echo -e "   ./start-integration.sh"
echo ""