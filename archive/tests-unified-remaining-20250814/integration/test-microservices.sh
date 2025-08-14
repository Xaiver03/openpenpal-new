#!/bin/bash

# OpenPenPal 微服务架构测试脚本
# 不包括Java Admin Service（需要本地安装Java）

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

PROJECT_ROOT="/Users/rocalight/同步空间/opplc/openpenpal"
LOG_DIR="$PROJECT_ROOT/logs"

# 日志函数
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# 创建日志目录
mkdir -p "$LOG_DIR"

# 停止所有服务
stop_all_services() {
    log_info "停止所有服务..."
    cd "$PROJECT_ROOT"
    ./startup/stop-all.sh > /dev/null 2>&1 || true
    sleep 2
}

# 启动服务并等待
start_service() {
    local service_name="$1"
    local service_dir="$2"
    local start_cmd="$3"
    local port="$4"
    local wait_time="${5:-10}"
    
    log_info "启动 $service_name..."
    
    cd "$service_dir"
    nohup $start_cmd > "$LOG_DIR/${service_name}.log" 2>&1 &
    local pid=$!
    echo $pid > "$LOG_DIR/${service_name}.pid"
    
    # 等待服务启动
    local count=0
    while [ $count -lt $wait_time ]; do
        if curl -s --max-time 2 --noproxy "*" "http://localhost:$port/health" > /dev/null 2>&1; then
            log_success "$service_name 启动成功 (PID: $pid, 端口: $port)"
            return 0
        fi
        sleep 1
        count=$((count + 1))
    done
    
    log_error "$service_name 启动失败或健康检查超时"
    return 1
}

# 主函数
main() {
    echo "========================================"
    echo "   OpenPenPal 微服务架构测试"
    echo "========================================"
    echo ""
    
    # 停止现有服务
    stop_all_services
    
    # 启动服务
    local success_count=0
    local total_services=5
    
    # 1. 主后端服务 (Go)
    log_info "=== 启动核心服务 ==="
    if start_service "go-backend" "$PROJECT_ROOT/backend" "./openpenpal" "8080" 15; then
        success_count=$((success_count + 1))
    fi
    
    # 2. API网关 (Go)
    log_info "=== 启动网关服务 ==="
    if start_service "gateway" "$PROJECT_ROOT/services/gateway" "./gateway" "8000" 10; then
        success_count=$((success_count + 1))
    fi
    
    # 3. Write服务 (Python)
    log_info "=== 启动Python微服务 ==="
    cd "$PROJECT_ROOT/services/write-service"
    source venv/bin/activate
    if start_service "write-service" "$PROJECT_ROOT/services/write-service" "python -m app.main" "8001" 15; then
        success_count=$((success_count + 1))
    fi
    
    # 4. Courier服务 (Go)
    if start_service "courier-service" "$PROJECT_ROOT/services/courier-service" "./main" "8002" 10; then
        success_count=$((success_count + 1))
    fi
    
    # 5. 前端 (Next.js)
    log_info "=== 启动前端服务 ==="
    if start_service "frontend" "$PROJECT_ROOT/frontend" "npm run dev" "3000" 20; then
        success_count=$((success_count + 1))
    fi
    
    # 测试结果
    echo ""
    echo "========================================"
    echo "   微服务启动测试结果"
    echo "========================================"
    echo ""
    echo "成功启动: $success_count/$total_services 个服务"
    
    if [ $success_count -eq $total_services ]; then
        log_success "🎉 所有微服务启动成功！"
        
        echo ""
        echo "可用服务："
        echo "• 主后端: http://localhost:8080"
        echo "• API网关: http://localhost:8000"  
        echo "• Write服务: http://localhost:8001"
        echo "• Courier服务: http://localhost:8002"
        echo "• 前端应用: http://localhost:3000"
        echo ""
        echo "测试API："
        echo "curl http://localhost:8080/health"
        echo "curl http://localhost:8000/health"
        echo "curl http://localhost:8001/health"
        echo "curl http://localhost:8002/health"
        echo ""
        
        return 0
    else
        log_warning "部分服务启动失败，请检查日志："
        echo "• 日志目录: $LOG_DIR"
        echo "• 查看特定服务: tail -f $LOG_DIR/[service-name].log"
        
        return 1
    fi
}

# 运行测试
main "$@"