#!/bin/bash

# OpenPenPal 生产环境启动脚本
# 启动所有核心服务

set -e

# 获取脚本目录和项目根目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# 导入工具函数
source "$SCRIPT_DIR/utils.sh"
source "$SCRIPT_DIR/environment-vars.sh"

# 设置生产环境变量
export NODE_ENV="production"
export DATABASE_TYPE="postgres"
export LOG_LEVEL="info"
export DEBUG="false"

log_info "🚀 启动 OpenPenPal 生产环境"
log_info "================================"

# 创建必要目录
mkdir -p "$LOG_DIR"
mkdir -p "$UPLOAD_DIR"

# 检查数据库连接
check_database() {
    log_info "检查数据库连接..."
    
    # 检查PostgreSQL是否运行
    if command -v pg_isready &> /dev/null; then
        if pg_isready -h localhost -p 5432 &> /dev/null; then
            log_success "PostgreSQL 正在运行"
        else
            log_error "PostgreSQL 未运行，请先启动数据库"
            log_info "运行: docker-compose up -d postgres"
            exit 1
        fi
    else
        log_warning "无法检查PostgreSQL状态，假设已运行"
    fi
}

# 启动服务的通用函数
start_service() {
    local service_name="$1"
    local service_dir="$2"
    local port="$3"
    local start_command="$4"
    local build_command="$5"
    
    log_info "启动 $service_name (端口 $port)..."
    
    # 检查端口
    if lsof -i :$port &> /dev/null; then
        log_warning "$service_name 端口 $port 已被占用，跳过启动"
        return 0
    fi
    
    # 检查目录
    if [ ! -d "$service_dir" ]; then
        log_error "$service_name 目录不存在: $service_dir"
        return 1
    fi
    
    cd "$service_dir"
    
    # 如果有构建命令，先构建
    if [ -n "$build_command" ]; then
        log_info "构建 $service_name..."
        eval "$build_command" || {
            log_error "$service_name 构建失败"
            return 1
        }
    fi
    
    # 启动服务
    local log_file="$LOG_DIR/${service_name}.log"
    local pid_file="$LOG_DIR/${service_name}.pid"
    
    nohup $start_command > "$log_file" 2>&1 &
    local pid=$!
    echo $pid > "$pid_file"
    
    # 等待服务启动
    sleep 3
    
    # 检查服务是否启动成功
    if kill -0 $pid 2>/dev/null; then
        log_success "$service_name 启动成功 (PID: $pid)"
    else
        log_error "$service_name 启动失败，请查看日志: $log_file"
        return 1
    fi
    
    cd "$PROJECT_ROOT"
}

# 主函数
main() {
    # 1. 检查数据库
    check_database
    
    # 2. 启动主后端服务
    start_service "go-backend" \
        "$PROJECT_ROOT/backend" \
        "8080" \
        "./openpenpal-backend" \
        ""
    
    # 3. 启动网关服务
    start_service "gateway" \
        "$PROJECT_ROOT/services/gateway" \
        "8000" \
        "./bin/gateway" \
        "go build -o bin/gateway cmd/main.go"
    
    # 4. 启动写信服务
    if [ -d "$PROJECT_ROOT/services/write-service/venv" ]; then
        start_service "write-service" \
            "$PROJECT_ROOT/services/write-service" \
            "8001" \
            "venv/bin/python app/main.py" \
            ""
    else
        log_warning "写信服务虚拟环境未创建，跳过"
    fi
    
    # 5. 启动信使服务
    start_service "courier-service" \
        "$PROJECT_ROOT/services/courier-service" \
        "8002" \
        "./bin/courier-service" \
        "go build -o bin/courier-service cmd/main.go"
    
    # 6. 启动前端服务
    start_service "frontend" \
        "$PROJECT_ROOT/frontend" \
        "3000" \
        "npm run dev" \
        ""
    
    log_info ""
    log_success "🎉 所有服务启动完成！"
    log_info ""
    log_info "📊 服务状态："
    log_info "  • 主后端: http://localhost:8080/health"
    log_info "  • API网关: http://localhost:8000/health"
    log_info "  • 写信服务: http://localhost:8001/health"
    log_info "  • 信使服务: http://localhost:8002/health"
    log_info "  • 前端应用: http://localhost:3000"
    log_info ""
    log_info "🔑 测试账号："
    log_info "  • admin/admin123 - 管理员"
    log_info "  • alice/secret - 普通用户"
    log_info ""
    log_info "📋 管理命令："
    log_info "  • 查看状态: ./startup/check-status.sh"
    log_info "  • 查看日志: tail -f logs/*.log"
    log_info "  • 停止服务: ./startup/stop-all.sh"
}

# 执行主函数
main "$@"