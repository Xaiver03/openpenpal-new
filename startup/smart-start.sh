#!/bin/bash

# OpenPenPal 智能启动脚本
# 自动检测端口并启动前后端服务

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 默认端口配置
FRONTEND_PORTS=(3000 3001 3002 3003 3004)
BACKEND_PORTS=(8080 8081 8082 8083 8084)

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查端口是否可用
check_port() {
    local port=$1
    if nc -z localhost $port 2>/dev/null; then
        return 1  # 端口被占用
    else
        return 0  # 端口可用
    fi
}

# 找到可用端口
find_available_port() {
    local ports=("$@")
    for port in "${ports[@]}"; do
        if check_port $port; then
            echo $port
            return 0
        fi
    done
    
    # 如果默认端口都被占用，生成随机端口
    local random_port=$((8000 + RANDOM % 2000))
    if check_port $random_port; then
        echo $random_port
        return 0
    fi
    
    return 1
}

# 显示端口状态
show_port_status() {
    log_info "当前端口使用情况:"
    
    echo "前端端口:"
    for port in "${FRONTEND_PORTS[@]}"; do
        if check_port $port; then
            echo -e "  端口 $port: ${GREEN}✅ 可用${NC}"
        else
            echo -e "  端口 $port: ${RED}❌ 被占用${NC}"
        fi
    done
    
    echo "后端端口:"
    for port in "${BACKEND_PORTS[@]}"; do
        if check_port $port; then
            echo -e "  端口 $port: ${GREEN}✅ 可用${NC}"
        else
            echo -e "  端口 $port: ${RED}❌ 被占用${NC}"
        fi
    done
}

# 启动后端服务
start_backend() {
    local backend_port=$(find_available_port "${BACKEND_PORTS[@]}")
    
    if [ -z "$backend_port" ]; then
        log_error "无法找到可用的后端端口"
        return 1
    fi
    
    if [ $backend_port -eq 8080 ]; then
        log_success "后端使用首选端口: $backend_port"
    else
        log_warning "端口 8080 被占用，后端使用端口: $backend_port"
    fi
    
    cd backend
    
    # 设置端口环境变量
    export PORT=$backend_port
    
    # 启动后端服务
    log_info "正在启动后端服务..."
    if [ -f "openpenpal" ]; then
        ./openpenpal > ../backend.log 2>&1 &
        BACKEND_PID=$!
        echo $BACKEND_PID > ../backend.pid
    else
        log_error "后端可执行文件不存在，请先编译"
        return 1
    fi
    
    cd ..
    
    # 等待后端启动
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if curl -s http://localhost:$backend_port/health > /dev/null 2>&1; then
            log_success "后端服务启动成功: http://localhost:$backend_port"
            echo $backend_port > backend.port
            return 0
        fi
        
        if [ $attempt -eq 1 ]; then
            log_info "等待后端服务启动..."
        fi
        
        sleep 1
        attempt=$((attempt + 1))
    done
    
    log_error "后端服务启动超时"
    return 1
}

# 启动前端服务  
start_frontend() {
    local frontend_port=$(find_available_port "${FRONTEND_PORTS[@]}")
    
    if [ -z "$frontend_port" ]; then
        log_error "无法找到可用的前端端口"
        return 1
    fi
    
    if [ $frontend_port -eq 3000 ]; then
        log_success "前端使用首选端口: $frontend_port" 
    else
        log_warning "端口 3000 被占用，前端使用端口: $frontend_port"
    fi
    
    cd frontend
    
    # 启动前端服务
    log_info "正在启动前端服务..."
    npm run smart-dev > ../frontend.log 2>&1 &
    FRONTEND_PID=$!
    echo $FRONTEND_PID > ../frontend.pid
    
    cd ..
    
    # 等待前端启动
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if curl -s http://localhost:$frontend_port > /dev/null 2>&1; then
            log_success "前端服务启动成功: http://localhost:$frontend_port"
            echo $frontend_port > frontend.port
            return 0
        fi
        
        sleep 1
        attempt=$((attempt + 1))
    done
    
    log_error "前端服务启动超时"
    return 1
}

# 停止服务
stop_services() {
    log_info "正在停止服务..."
    
    # 停止前端
    if [ -f "frontend.pid" ]; then
        local frontend_pid=$(cat frontend.pid)
        if kill -0 $frontend_pid 2>/dev/null; then
            kill $frontend_pid
            log_success "前端服务已停止"
        fi
        rm -f frontend.pid frontend.port
    fi
    
    # 停止后端
    if [ -f "backend.pid" ]; then
        local backend_pid=$(cat backend.pid)
        if kill -0 $backend_pid 2>/dev/null; then
            kill $backend_pid
            log_success "后端服务已停止"
        fi
        rm -f backend.pid backend.port
    fi
}

# 显示帮助
show_help() {
    cat << EOF
OpenPenPal 智能启动脚本

用法:
  $0 [选项]

选项:
  start       启动前后端服务（默认）
  stop        停止所有服务
  restart     重启所有服务
  status      显示端口使用情况
  frontend    仅启动前端服务
  backend     仅启动后端服务
  --help, -h  显示此帮助信息

特性:
  - 自动检测可用端口
  - 智能端口切换
  - 服务健康检查
  - 优雅关闭处理
  - 详细的启动日志

默认端口:
  前端: ${FRONTEND_PORTS[*]}
  后端: ${BACKEND_PORTS[*]}
EOF
}

# 主函数
main() {
    case "${1:-start}" in
        "start")
            log_info "启动 OpenPenPal 开发环境..."
            
            # 检查依赖
            if ! command -v nc &> /dev/null; then
                log_error "需要安装 netcat (nc) 命令"
                exit 1
            fi
            
            # 停止现有服务
            stop_services
            
            # 启动服务
            if start_backend && start_frontend; then
                local backend_port=$(cat backend.port 2>/dev/null || echo "未知")
                local frontend_port=$(cat frontend.port 2>/dev/null || echo "未知")
                
                echo
                log_success "🎉 OpenPenPal 启动成功!"
                echo -e "${GREEN}前端访问地址: http://localhost:$frontend_port${NC}"
                echo -e "${GREEN}后端API地址: http://localhost:$backend_port${NC}"
                echo
                log_info "按 Ctrl+C 停止服务"
                
                # 等待中断信号
                trap stop_services EXIT INT TERM
                wait
            else
                log_error "服务启动失败"
                stop_services
                exit 1
            fi
            ;;
        "stop")
            stop_services
            ;;
        "restart")
            stop_services
            sleep 2
            $0 start
            ;;
        "status")
            show_port_status
            ;;
        "frontend")
            stop_services
            start_frontend
            ;;
        "backend")
            stop_services
            start_backend
            ;;
        "--help"|"-h")
            show_help
            ;;
        *)
            log_error "未知选项: $1"
            show_help
            exit 1
            ;;
    esac
}

# 运行主函数
main "$@"