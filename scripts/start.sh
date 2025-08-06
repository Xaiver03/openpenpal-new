#!/bin/bash

# OpenPenPal 项目启动程序 (.command 文件)
# 可以直接双击运行

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 函数：打印带颜色的消息
print_message() {
    echo -e "${2}[$(date '+%H:%M:%S')] $1${NC}"
}

# 函数：检查端口是否被占用
check_port() {
    local port=$1
    if lsof -i :$port > /dev/null 2>&1; then
        return 0  # 端口被占用
    else
        return 1  # 端口空闲
    fi
}

# 函数：停止占用端口的进程
kill_port() {
    local port=$1
    print_message "检测到端口 $port 被占用，正在停止相关进程..." $YELLOW
    
    # 查找并终止占用端口的进程
    local pids=$(lsof -ti :$port)
    if [ ! -z "$pids" ]; then
        echo $pids | xargs kill -9 2>/dev/null
        sleep 2
        print_message "已停止端口 $port 上的进程" $GREEN
    fi
}

# 函数：检查Go环境
check_go() {
    # 设置Go环境变量
    export PATH="/usr/local/go/bin:$PATH"
    export GOROOT="/usr/local/go"
    export GOPATH="$HOME/go"
    
    if ! command -v go &> /dev/null; then
        print_message "❌ 错误：未安装 Go" $RED
        print_message "请先安装 Go (https://golang.org/dl/)" $RED
        return 1
    fi
    print_message "✅ Go 版本: $(go version | cut -d' ' -f3)" $GREEN
    return 0
}

# 函数：检查Node.js环境
check_node() {
    if ! command -v node &> /dev/null; then
        print_message "❌ 错误：未安装 Node.js" $RED
        print_message "请先安装 Node.js (https://nodejs.org)" $RED
        return 1
    fi
    
    if ! command -v npm &> /dev/null; then
        print_message "❌ 错误：未安装 npm" $RED
        return 1
    fi
    
    print_message "✅ Node.js 版本: $(node --version)" $GREEN
    print_message "✅ npm 版本: $(npm --version)" $GREEN
    return 0
}

# 函数：启动后端服务
start_backend() {
    print_message "🔧 准备启动后端服务..." $CYAN
    
    # 设置Go环境变量
    export PATH="/usr/local/go/bin:$PATH"
    export GOROOT="/usr/local/go"
    export GOPATH="$HOME/go"
    
    # 检查后端目录
    if [ ! -d "backend" ]; then
        print_message "❌ 错误：未找到 backend 目录" $RED
        return 1
    fi
    
    cd backend
    
    # 检查 go.mod
    if [ ! -f "go.mod" ]; then
        print_message "❌ 错误：backend 目录不是有效的 Go 项目" $RED
        cd ..
        return 1
    fi
    
    # 安装依赖
    print_message "📦 安装后端依赖..." $YELLOW
    go mod tidy
    
    # 检查端口
    BACKEND_PORT=8080
    if check_port $BACKEND_PORT; then
        kill_port $BACKEND_PORT
    fi
    
    # 启动后端服务
    print_message "🚀 启动后端服务 (端口 $BACKEND_PORT)..." $GREEN
    go run main.go &
    BACKEND_PID=$!
    
    # 等待后端启动
    sleep 3
    
    # 检查后端是否启动成功
    if ! kill -0 $BACKEND_PID 2>/dev/null; then
        print_message "❌ 后端服务启动失败" $RED
        cd ..
        return 1
    fi
    
    print_message "✅ 后端服务启动成功 (PID: $BACKEND_PID)" $GREEN
    cd ..
    return 0
}

# 函数：启动前端服务
start_frontend() {
    print_message "🎨 准备启动前端服务..." $CYAN
    
    # 检查前端目录
    if [ ! -d "frontend" ]; then
        print_message "❌ 错误：未找到 frontend 目录" $RED
        return 1
    fi
    
    cd frontend
    
    # 检查 package.json
    if [ ! -f "package.json" ]; then
        print_message "❌ 错误：frontend 目录不是有效的 Node.js 项目" $RED
        cd ..
        return 1
    fi
    
    # 安装依赖
    if [ ! -d "node_modules" ]; then
        print_message "📦 首次运行，安装前端依赖..." $YELLOW
        npm install
        if [ $? -ne 0 ]; then
            print_message "❌ 前端依赖安装失败" $RED
            cd ..
            return 1
        fi
        print_message "✅ 前端依赖安装完成" $GREEN
    fi
    
    # 选择前端端口
    FRONTEND_PORT=3000
    if check_port $FRONTEND_PORT; then
        kill_port $FRONTEND_PORT
        
        # 再次检查
        if check_port $FRONTEND_PORT; then
            # 寻找可用端口
            for p in 3001 3002 3003 3004 3005; do
                if ! check_port $p; then
                    FRONTEND_PORT=$p
                    print_message "自动选择前端端口 $FRONTEND_PORT" $GREEN
                    break
                fi
            done
            
            if check_port $FRONTEND_PORT; then
                print_message "❌ 无法找到可用的前端端口" $RED
                cd ..
                return 1
            fi
        fi
    fi
    
    print_message "✅ 前端使用端口: $FRONTEND_PORT" $GREEN
    
    # 启动前端服务
    print_message "🚀 启动前端服务 (端口 $FRONTEND_PORT)..." $GREEN
    npm run dev -- --port $FRONTEND_PORT &
    FRONTEND_PID=$!
    
    cd ..
    
    # 等待前端启动
    print_message "⏳ 等待前端服务启动..." $YELLOW
    sleep 5
    
    # 检查前端是否启动成功
    local max_attempts=10
    local attempt=0
    while [ $attempt -lt $max_attempts ]; do
        if curl -s "http://localhost:$FRONTEND_PORT" > /dev/null 2>&1; then
            print_message "✅ 前端服务启动成功 (PID: $FRONTEND_PID)" $GREEN
            return 0
        fi
        sleep 2
        attempt=$((attempt + 1))
        print_message "⏳ 等待前端服务响应... ($attempt/$max_attempts)" $YELLOW
    done
    
    print_message "❌ 前端服务启动超时" $RED
    return 1
}

# 函数：清理函数
cleanup() {
    print_message "🛑 正在停止服务..." $YELLOW
    
    if [ ! -z "$BACKEND_PID" ]; then
        kill $BACKEND_PID 2>/dev/null
        print_message "✅ 后端服务已停止" $GREEN
    fi
    
    if [ ! -z "$FRONTEND_PID" ]; then
        kill $FRONTEND_PID 2>/dev/null
        print_message "✅ 前端服务已停止" $GREEN
    fi
    
    # 清理可能残留的进程
    kill_port 8080 2>/dev/null
    kill_port 3000 2>/dev/null
    kill_port 3001 2>/dev/null
    
    print_message "🛑 所有服务已停止" $YELLOW
}

# 设置信号处理
trap cleanup EXIT INT TERM

# 清屏
clear

print_message "🌟 OpenPenPal 项目启动程序" $BLUE
print_message "====================================" $BLUE
print_message "当前目录: $SCRIPT_DIR" $BLUE
print_message "====================================" $BLUE

# 检查项目结构
if [ ! -d "frontend" ] || [ ! -d "backend" ]; then
    print_message "❌ 错误：项目结构不完整" $RED
    print_message "请确保当前目录包含 frontend 和 backend 文件夹" $RED
    read -p "按任意键退出..."
    exit 1
fi

# 检查环境
print_message "🔍 检查开发环境..." $CYAN

if ! check_go; then
    read -p "按任意键退出..."
    exit 1
fi

if ! check_node; then
    read -p "按任意键退出..."
    exit 1
fi

print_message "✅ 开发环境检查完成" $GREEN
echo ""

# 启动服务
print_message "🚀 开始启动 OpenPenPal 服务..." $BLUE
print_message "====================================" $BLUE

# 启动后端
if ! start_backend; then
    print_message "❌ 后端服务启动失败，退出程序" $RED
    read -p "按任意键退出..."
    exit 1
fi

echo ""

# 启动前端
if ! start_frontend; then
    print_message "❌ 前端服务启动失败，退出程序" $RED
    read -p "按任意键退出..."
    exit 1
fi

echo ""
print_message "🎉 OpenPenPal 服务启动成功！" $GREEN
print_message "====================================" $BLUE
print_message "🌐 访问地址:" $GREEN
print_message "   📱 前端应用: http://localhost:$FRONTEND_PORT" $CYAN
print_message "   🔧 后端API: http://localhost:8080" $CYAN
print_message "   📚 API文档: http://localhost:8080/swagger/index.html" $CYAN
print_message "====================================" $BLUE
print_message "👤 测试账号:" $YELLOW
print_message "   🔑 超级管理员: super_admin / secret" $YELLOW
print_message "   👨‍💼 平台管理员: platform_admin / secret" $YELLOW
print_message "   🚀 信使账号: courier1 / secret" $YELLOW
print_message "   👤 普通用户: alice / secret" $YELLOW
print_message "====================================" $BLUE
print_message "💡 使用提示:" $YELLOW
print_message "   • 浏览器将自动打开前端应用" $YELLOW
print_message "   • 按 Ctrl+C 可停止所有服务" $YELLOW
print_message "   • 关闭此窗口也会停止所有服务" $YELLOW
print_message "====================================" $BLUE
echo ""

# 自动打开浏览器
print_message "🌐 正在打开浏览器..." $CYAN
sleep 2
open "http://localhost:$FRONTEND_PORT"

# 保持运行
print_message "✅ 所有服务正在运行中..." $GREEN
print_message "按 Ctrl+C 停止服务" $YELLOW
echo ""

# 等待用户中断
while true; do
    sleep 1
done