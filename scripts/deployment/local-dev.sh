#!/bin/bash

# 本地一键联调脚本
# 用于快速启动开发环境，包括依赖服务和应用

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
GRAY='\033[0;90m'
NC='\033[0m'

# 项目根目录
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
DEPLOYMENT_DIR="$PROJECT_ROOT/scripts/deployment"
TMP_DIR="$PROJECT_ROOT/.tmp"
LOG_DIR="$PROJECT_ROOT/logs"

# 服务端口配置
FRONTEND_PORT="${FRONTEND_PORT:-3000}"
BACKEND_PORT="${PORT:-8080}"
POSTGRES_PORT="${POSTGRES_PORT:-5432}"
REDIS_PORT="${REDIS_PORT:-6379}"
PROMETHEUS_PORT="${PROMETHEUS_PORT:-9090}"
GRAFANA_PORT="${GRAFANA_PORT:-3001}"

# PID文件
BACKEND_PID_FILE="$TMP_DIR/backend.pid"
FRONTEND_PID_FILE="$TMP_DIR/frontend.pid"

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

log_step() {
    echo -e "\n${BLUE}==>${NC} $1"
}

# 清理函数
cleanup() {
    log_step "清理环境"
    
    # 停止应用服务
    if [ -f "$BACKEND_PID_FILE" ]; then
        kill -TERM $(cat "$BACKEND_PID_FILE") 2>/dev/null || true
        rm -f "$BACKEND_PID_FILE"
    fi
    
    if [ -f "$FRONTEND_PID_FILE" ]; then
        kill -TERM $(cat "$FRONTEND_PID_FILE") 2>/dev/null || true
        rm -f "$FRONTEND_PID_FILE"
    fi
    
    # 停止Docker服务
    if command -v docker-compose &> /dev/null; then
        cd "$DEPLOYMENT_DIR" && docker-compose down -v || true
    fi
    
    log_success "环境清理完成"
}

# 检查端口占用
check_port() {
    local port=$1
    local service=$2
    
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
        log_error "端口 $port ($service) 已被占用"
        log_info "使用以下命令查看占用进程: lsof -i :$port"
        return 1
    fi
    return 0
}

# 检查所有端口
check_all_ports() {
    log_step "检查端口可用性"
    
    local all_clear=true
    
    check_port $FRONTEND_PORT "Frontend" || all_clear=false
    check_port $BACKEND_PORT "Backend" || all_clear=false
    check_port $POSTGRES_PORT "PostgreSQL" || all_clear=false
    check_port $REDIS_PORT "Redis" || all_clear=false
    check_port $PROMETHEUS_PORT "Prometheus" || all_clear=false
    check_port $GRAFANA_PORT "Grafana" || all_clear=false
    
    if [ "$all_clear" = false ]; then
        log_error "某些端口被占用，请先释放端口"
        exit 1
    fi
    
    log_success "所有端口检查通过"
}

# 校验环境变量
validate_env() {
    log_step "校验环境变量"
    
    if [ -x "$DEPLOYMENT_DIR/validate-env.js" ]; then
        node "$DEPLOYMENT_DIR/validate-env.js" || {
            log_error "环境变量校验失败"
            exit 1
        }
    else
        log_warning "环境变量校验脚本不存在，跳过校验"
    fi
}

# 创建必要目录
create_directories() {
    log_step "创建必要目录"
    
    mkdir -p "$TMP_DIR"
    mkdir -p "$LOG_DIR"
    mkdir -p "$PROJECT_ROOT/uploads"
    mkdir -p "$PROJECT_ROOT/backups"
    
    log_success "目录创建完成"
}

# 启动依赖服务
start_dependencies() {
    log_step "启动依赖服务"
    
    # 创建docker-compose文件（如果不存在）
    if [ ! -f "$DEPLOYMENT_DIR/docker-compose.yml" ]; then
        create_docker_compose
    fi
    
    # 启动Docker服务
    cd "$DEPLOYMENT_DIR"
    docker-compose up -d postgres redis
    
    # 等待服务就绪
    log_info "等待PostgreSQL启动..."
    local retries=30
    while [ $retries -gt 0 ]; do
        if docker-compose exec -T postgres pg_isready &>/dev/null; then
            break
        fi
        retries=$((retries - 1))
        sleep 1
    done
    
    if [ $retries -eq 0 ]; then
        log_error "PostgreSQL启动超时"
        exit 1
    fi
    
    log_info "等待Redis启动..."
    retries=30
    while [ $retries -gt 0 ]; do
        if docker-compose exec -T redis redis-cli ping &>/dev/null; then
            break
        fi
        retries=$((retries - 1))
        sleep 1
    done
    
    if [ $retries -eq 0 ]; then
        log_error "Redis启动超时"
        exit 1
    fi
    
    log_success "依赖服务启动成功"
}

# 数据库迁移
run_migrations() {
    log_step "执行数据库迁移"
    
    if [ -x "$DEPLOYMENT_DIR/db-migrate.sh" ]; then
        "$DEPLOYMENT_DIR/db-migrate.sh" up || {
            log_error "数据库迁移失败"
            exit 1
        }
    else
        log_warning "数据库迁移脚本不存在，跳过迁移"
    fi
}

# 加载种子数据
load_seed_data() {
    log_step "加载种子数据"
    
    if [ -x "$DEPLOYMENT_DIR/db-migrate.sh" ]; then
        "$DEPLOYMENT_DIR/db-migrate.sh" seed development || {
            log_warning "种子数据加载失败"
        }
    fi
}

# 启动后端服务
start_backend() {
    log_step "启动后端服务"
    
    cd "$PROJECT_ROOT/backend"
    
    # 检查是否有air（热重载）
    if command -v air &> /dev/null; then
        log_info "使用air启动后端（热重载模式）"
        AIR_TMP_DIR="$TMP_DIR" air -c .air.toml > "$LOG_DIR/backend.log" 2>&1 &
        echo $! > "$BACKEND_PID_FILE"
    else
        log_info "使用go run启动后端"
        go run main.go > "$LOG_DIR/backend.log" 2>&1 &
        echo $! > "$BACKEND_PID_FILE"
    fi
    
    # 等待后端启动
    log_info "等待后端服务启动..."
    local retries=30
    while [ $retries -gt 0 ]; do
        if curl -s "http://localhost:$BACKEND_PORT/healthz" &>/dev/null; then
            break
        fi
        retries=$((retries - 1))
        sleep 1
    done
    
    if [ $retries -eq 0 ]; then
        log_error "后端启动超时"
        tail -n 20 "$LOG_DIR/backend.log"
        exit 1
    fi
    
    log_success "后端服务启动成功 (端口: $BACKEND_PORT)"
}

# 启动前端服务
start_frontend() {
    log_step "启动前端服务"
    
    cd "$PROJECT_ROOT/frontend"
    
    # 安装依赖（如果需要）
    if [ ! -d "node_modules" ]; then
        log_info "安装前端依赖..."
        npm install
    fi
    
    # 启动前端
    npm run dev > "$LOG_DIR/frontend.log" 2>&1 &
    echo $! > "$FRONTEND_PID_FILE"
    
    # 等待前端启动
    log_info "等待前端服务启动..."
    local retries=30
    while [ $retries -gt 0 ]; do
        if curl -s "http://localhost:$FRONTEND_PORT" &>/dev/null; then
            break
        fi
        retries=$((retries - 1))
        sleep 1
    done
    
    if [ $retries -eq 0 ]; then
        log_error "前端启动超时"
        tail -n 20 "$LOG_DIR/frontend.log"
        exit 1
    fi
    
    log_success "前端服务启动成功 (端口: $FRONTEND_PORT)"
}

# 启动监控服务
start_monitoring() {
    log_step "启动监控服务（可选）"
    
    if [ "$PROMETHEUS_ENABLED" = "true" ] || [ "$GRAFANA_ENABLED" = "true" ]; then
        cd "$DEPLOYMENT_DIR"
        
        if [ "$PROMETHEUS_ENABLED" = "true" ]; then
            docker-compose up -d prometheus
            log_success "Prometheus启动成功 (端口: $PROMETHEUS_PORT)"
        fi
        
        if [ "$GRAFANA_ENABLED" = "true" ]; then
            docker-compose up -d grafana
            log_success "Grafana启动成功 (端口: $GRAFANA_PORT)"
        fi
    else
        log_info "监控服务未启用"
    fi
}

# 运行测试
run_tests() {
    log_step "运行冒烟测试"
    
    # 运行API测试
    if [ -x "$PROJECT_ROOT/scripts/test-apis.sh" ]; then
        "$PROJECT_ROOT/scripts/test-apis.sh" || {
            log_warning "API测试失败"
        }
    fi
    
    # 运行前端测试
    cd "$PROJECT_ROOT/frontend"
    npm run test:ci || {
        log_warning "前端测试失败"
    }
    
    log_success "测试完成"
}

# 显示服务状态
show_status() {
    log_step "服务状态"
    
    echo -e "\n${GREEN}服务已启动:${NC}"
    echo -e "  前端: ${BLUE}http://localhost:$FRONTEND_PORT${NC}"
    echo -e "  后端: ${BLUE}http://localhost:$BACKEND_PORT${NC}"
    echo -e "  健康检查: ${BLUE}http://localhost:$BACKEND_PORT/healthz${NC}"
    
    if [ "$PROMETHEUS_ENABLED" = "true" ]; then
        echo -e "  Prometheus: ${BLUE}http://localhost:$PROMETHEUS_PORT${NC}"
    fi
    
    if [ "$GRAFANA_ENABLED" = "true" ]; then
        echo -e "  Grafana: ${BLUE}http://localhost:$GRAFANA_PORT${NC}"
    fi
    
    echo -e "\n${YELLOW}日志文件:${NC}"
    echo -e "  前端: $LOG_DIR/frontend.log"
    echo -e "  后端: $LOG_DIR/backend.log"
    
    echo -e "\n${YELLOW}测试账号:${NC}"
    echo -e "  管理员: admin/Admin123!"
    echo -e "  用户: alice/Secret123!"
    
    echo -e "\n${GRAY}停止服务: $0 stop${NC}"
}

# 创建docker-compose文件
create_docker_compose() {
    cat > "$DEPLOYMENT_DIR/docker-compose.yml" <<EOF
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    container_name: openpenpal-postgres
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: openpenpal
    ports:
      - "$POSTGRES_PORT:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    container_name: openpenpal-redis
    ports:
      - "$REDIS_PORT:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  prometheus:
    image: prom/prometheus:latest
    container_name: openpenpal-prometheus
    ports:
      - "$PROMETHEUS_PORT:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus_data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
    profiles:
      - monitoring

  grafana:
    image: grafana/grafana:latest
    container_name: openpenpal-grafana
    ports:
      - "$GRAFANA_PORT:3000"
    environment:
      - GF_SECURITY_ADMIN_USER=admin
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - grafana_data:/var/lib/grafana
    profiles:
      - monitoring

volumes:
  postgres_data:
  prometheus_data:
  grafana_data:
EOF
}

# 停止服务
stop_services() {
    log_step "停止所有服务"
    cleanup
}

# 显示帮助
show_help() {
    cat <<EOF
本地开发环境一键启动脚本

用法:
    $0 [command] [options]

命令:
    start    启动所有服务（默认）
    stop     停止所有服务
    restart  重启所有服务
    status   显示服务状态
    test     运行测试
    help     显示帮助

选项:
    --skip-tests     跳过测试
    --skip-seed      跳过种子数据
    --monitoring     启用监控服务

环境变量:
    FRONTEND_PORT     前端端口 (默认: 3000)
    BACKEND_PORT      后端端口 (默认: 8080)
    POSTGRES_PORT     PostgreSQL端口 (默认: 5432)
    REDIS_PORT        Redis端口 (默认: 6379)
    PROMETHEUS_PORT   Prometheus端口 (默认: 9090)
    GRAFANA_PORT      Grafana端口 (默认: 3001)

示例:
    $0                      # 启动所有服务
    $0 --skip-tests         # 启动服务但跳过测试
    $0 --monitoring         # 启动服务和监控
    $0 stop                 # 停止所有服务
EOF
}

# 主函数
main() {
    local command=${1:-start}
    local skip_tests=false
    local skip_seed=false
    
    # 解析参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            start|stop|restart|status|test|help)
                command=$1
                ;;
            --skip-tests)
                skip_tests=true
                ;;
            --skip-seed)
                skip_seed=true
                ;;
            --monitoring)
                export PROMETHEUS_ENABLED=true
                export GRAFANA_ENABLED=true
                ;;
            *)
                log_error "未知参数: $1"
                show_help
                exit 1
                ;;
        esac
        shift
    done
    
    case $command in
        start)
            log_info "启动本地开发环境..."
            
            # 设置清理trap
            trap cleanup EXIT
            
            # 执行启动流程
            check_all_ports
            validate_env
            create_directories
            start_dependencies
            run_migrations
            
            if [ "$skip_seed" = false ]; then
                load_seed_data
            fi
            
            start_backend
            start_frontend
            start_monitoring
            
            if [ "$skip_tests" = false ]; then
                run_tests
            fi
            
            show_status
            
            # 保持运行
            log_info "按 Ctrl+C 停止服务..."
            wait
            ;;
            
        stop)
            stop_services
            ;;
            
        restart)
            stop_services
            sleep 2
            exec "$0" start "${@:2}"
            ;;
            
        status)
            show_status
            ;;
            
        test)
            run_tests
            ;;
            
        help|--help|-h)
            show_help
            ;;
            
        *)
            log_error "未知命令: $command"
            show_help
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"