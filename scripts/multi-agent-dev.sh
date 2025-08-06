#!/bin/bash

# OpenPenPal Multi-Agent Development Script
# Agent #1 (队长) 创建的协同开发管理脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 项目根目录
PROJECT_ROOT="/Users/rocalight/同步空间/opplc/openpenpal"
cd "$PROJECT_ROOT"

echo -e "${CYAN}🧠 OpenPenPal Multi-Agent Development Manager${NC}"
echo -e "${CYAN}================================================${NC}"

# 显示当前Agent状态
show_agent_status() {
    echo -e "\n${BLUE}📊 Agent 开发状态检查${NC}"
    echo -e "${BLUE}========================${NC}"
    
    echo -e "${PURPLE}Agent #1 (队长)${NC}: 前端集成 + 整体协调 ✅"
    
    if [ -d "services/write-service" ] && [ -f "services/write-service/requirements.txt" ]; then
        echo -e "${GREEN}Agent #2${NC}: 写信服务 (Python FastAPI) ✅"
    else
        echo -e "${YELLOW}Agent #2${NC}: 写信服务 (Python FastAPI) 🔄 待开发"
    fi
    
    if [ -d "services/courier-service" ] && [ -f "services/courier-service/go.mod" ]; then
        echo -e "${GREEN}Agent #3${NC}: 信使服务 (Go Gin) ✅"
    else
        echo -e "${YELLOW}Agent #3${NC}: 信使服务 (Go Gin) 🔄 待开发"
    fi
    
    if [ -d "services/admin-service" ] && [ -f "services/admin-service/pom.xml" ]; then
        echo -e "${GREEN}Agent #4${NC}: 管理后台 (Spring Boot) ✅"
    else
        echo -e "${YELLOW}Agent #4${NC}: 管理后台 (Spring Boot) 🔄 待开发"
    fi
    
    if [ -d "services/ocr-service" ] && [ -f "services/ocr-service/requirements.txt" ]; then
        echo -e "${GREEN}Agent #5${NC}: OCR识别 (Python Flask) ✅"
    else
        echo -e "${YELLOW}Agent #5${NC}: OCR识别 (Python Flask) 🔄 待开发"
    fi
}

# 启动特定服务
start_service() {
    local service=$1
    
    case $service in
        "frontend")
            echo -e "${GREEN}🚀 启动前端服务...${NC}"
            cd frontend
            if [ ! -d "node_modules" ]; then
                echo -e "${YELLOW}📦 安装前端依赖...${NC}"
                npm install
            fi
            npm run dev &
            ;;
            
        "backend")
            echo -e "${GREEN}🚀 启动当前Go后端...${NC}"
            cd backend
            if [ ! -f "go.mod" ]; then
                echo -e "${RED}❌ 后端go.mod不存在${NC}"
                return 1
            fi
            go run main.go &
            ;;
            
        "write-service")
            echo -e "${GREEN}🚀 启动写信服务...${NC}"
            if [ -d "services/write-service" ]; then
                cd services/write-service
                if [ -f "requirements.txt" ]; then
                    python -m uvicorn app.main:app --host 0.0.0.0 --port 8001 --reload &
                else
                    echo -e "${RED}❌ 写信服务未初始化${NC}"
                fi
            else
                echo -e "${RED}❌ 写信服务目录不存在${NC}"
            fi
            ;;
            
        "courier-service")
            echo -e "${GREEN}🚀 启动信使服务...${NC}"
            if [ -d "services/courier-service" ]; then
                cd services/courier-service
                if [ -f "go.mod" ]; then
                    go run cmd/main.go &
                else
                    echo -e "${RED}❌ 信使服务未初始化${NC}"
                fi
            else
                echo -e "${RED}❌ 信使服务目录不存在${NC}"
            fi
            ;;
            
        "all")
            echo -e "${GREEN}🚀 启动所有可用服务...${NC}"
            start_service "frontend"
            start_service "backend" 
            start_service "write-service"
            start_service "courier-service"
            ;;
            
        *)
            echo -e "${RED}❌ 未知服务: $service${NC}"
            echo -e "${YELLOW}可用服务: frontend, backend, write-service, courier-service, all${NC}"
            ;;
    esac
}

# 停止所有服务
stop_all_services() {
    echo -e "${RED}🛑 停止所有服务...${NC}"
    
    # 停止Node.js进程
    pkill -f "next-server" 2>/dev/null || true
    pkill -f "npm run dev" 2>/dev/null || true
    
    # 停止Go进程
    pkill -f "main.go" 2>/dev/null || true
    pkill -f "gin-bin" 2>/dev/null || true
    
    # 停止Python进程
    pkill -f "uvicorn" 2>/dev/null || true
    pkill -f "flask" 2>/dev/null || true
    
    echo -e "${GREEN}✅ 所有服务已停止${NC}"
}

# 检查端口占用
check_ports() {
    echo -e "\n${BLUE}🔍 端口占用检查${NC}"
    echo -e "${BLUE}=================${NC}"
    
    local ports=(3000 8080 8001 8002 8003 8004 5432 6379)
    
    for port in "${ports[@]}"; do
        if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
            local process=$(lsof -Pi :$port -sTCP:LISTEN -t | head -1)
            local process_name=$(ps -p $process -o comm= 2>/dev/null || echo "未知")
            echo -e "${YELLOW}端口 $port${NC}: 被占用 (PID: $process, 进程: $process_name)"
        else
            echo -e "${GREEN}端口 $port${NC}: 可用"
        fi
    done
}

# 初始化Agent开发环境
init_agent_env() {
    local agent_id=$1
    
    case $agent_id in
        "2")
            echo -e "${GREEN}🔧 初始化Agent #2 写信服务环境...${NC}"
            mkdir -p services/write-service/app/{models,schemas,api,core,utils}
            cd services/write-service
            
            # 创建Python虚拟环境
            if [ ! -d "venv" ]; then
                python3 -m venv venv
                source venv/bin/activate
                pip install fastapi uvicorn sqlalchemy psycopg2-binary pydantic
                pip freeze > requirements.txt
                deactivate
            fi
            
            echo -e "${GREEN}✅ Agent #2 环境初始化完成${NC}"
            echo -e "${YELLOW}📋 任务卡片: agent-tasks/AGENT-2-WRITE-SERVICE.md${NC}"
            ;;
            
        "3")
            echo -e "${GREEN}🔧 初始化Agent #3 信使服务环境...${NC}"
            mkdir -p services/courier-service/{cmd,internal/{config,models,handlers,services,middleware,utils}}
            cd services/courier-service
            
            # 初始化Go模块
            if [ ! -f "go.mod" ]; then
                go mod init courier-service
                go get github.com/gin-gonic/gin
                go get gorm.io/gorm
                go get gorm.io/driver/postgres
                go get github.com/golang-jwt/jwt/v4
                go get github.com/go-redis/redis/v8
            fi
            
            echo -e "${GREEN}✅ Agent #3 环境初始化完成${NC}"
            echo -e "${YELLOW}📋 任务卡片: agent-tasks/AGENT-3-COURIER-SERVICE.md${NC}"
            ;;
            
        *)
            echo -e "${RED}❌ 未知Agent ID: $agent_id${NC}"
            echo -e "${YELLOW}可用Agent: 2 (写信服务), 3 (信使服务)${NC}"
            ;;
    esac
}

# 运行集成测试
run_integration_tests() {
    echo -e "${BLUE}🧪 运行集成测试...${NC}"
    
    # 检查服务是否运行
    if ! curl -s http://localhost:3000 >/dev/null; then
        echo -e "${RED}❌ 前端服务未运行${NC}"
        return 1
    fi
    
    if ! curl -s http://localhost:8080/health >/dev/null; then
        echo -e "${RED}❌ 后端服务未运行${NC}"
        return 1
    fi
    
    echo -e "${GREEN}✅ 基础服务检查通过${NC}"
    
    # 可以添加更多集成测试逻辑
}

# 显示开发者指南
show_dev_guide() {
    echo -e "\n${CYAN}📚 Multi-Agent 开发指南${NC}"
    echo -e "${CYAN}========================${NC}"
    echo -e "${YELLOW}1. 检查开发状态:${NC} ./multi-agent-dev.sh status"
    echo -e "${YELLOW}2. 初始化Agent环境:${NC} ./multi-agent-dev.sh init <agent_id>"
    echo -e "${YELLOW}3. 启动服务:${NC} ./multi-agent-dev.sh start <service_name>"
    echo -e "${YELLOW}4. 停止所有服务:${NC} ./multi-agent-dev.sh stop"
    echo -e "${YELLOW}5. 检查端口:${NC} ./multi-agent-dev.sh ports"
    echo -e "${YELLOW}6. 运行测试:${NC} ./multi-agent-dev.sh test"
    echo -e "${YELLOW}7. 显示此帮助:${NC} ./multi-agent-dev.sh help"
    
    echo -e "\n${CYAN}📋 Agent任务分配:${NC}"
    echo -e "${PURPLE}Agent #2${NC}: 写信服务 (Python FastAPI)"
    echo -e "${PURPLE}Agent #3${NC}: 信使服务 (Go Gin)"  
    echo -e "${PURPLE}Agent #4${NC}: 管理后台 (Spring Boot)"
    echo -e "${PURPLE}Agent #5${NC}: OCR服务 (Python Flask)"
    
    echo -e "\n${CYAN}🔗 相关文档:${NC}"
    echo -e "• 协同规范: MULTI_AGENT_COORDINATION.md"
    echo -e "• API规范: docs/api/UNIFIED_API_SPECIFICATION.md"
    echo -e "• 任务卡片: agent-tasks/"
}

# 主函数
main() {
    case ${1:-help} in
        "status")
            show_agent_status
            ;;
        "start")
            start_service ${2:-all}
            ;;
        "stop")
            stop_all_services
            ;;
        "ports")
            check_ports
            ;;
        "init")
            init_agent_env $2
            ;;
        "test")
            run_integration_tests
            ;;
        "help"|*)
            show_dev_guide
            ;;
    esac
}

# 设置脚本执行权限
chmod +x "$0"

# 执行主函数
main "$@"