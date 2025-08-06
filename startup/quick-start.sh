#!/bin/bash

# OpenPenPal 快速启动脚本
# 一键启动所有必需的服务

set -e

# 获取脚本目录和项目根目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# 导入环境变量和工具函数
source "$SCRIPT_DIR/environment-vars.sh"
source "$SCRIPT_DIR/utils.sh"
source "$SCRIPT_DIR/browser-manager.sh"
source "$SCRIPT_DIR/url-manager.sh"
source "$SCRIPT_DIR/database-manager.sh"

# 确保本地服务不使用代理 - 包含所有可能的端口
export NO_PROXY="localhost,127.0.0.1,*.local,localhost:*,127.0.0.1:*"
export no_proxy="localhost,127.0.0.1,*.local,localhost:*,127.0.0.1:*"
# 如果有 HTTP_PROXY 或 HTTPS_PROXY，临时禁用它们对本地的影响
export HTTP_PROXY_BACKUP="$HTTP_PROXY"
export HTTPS_PROXY_BACKUP="$HTTPS_PROXY"
unset HTTP_PROXY
unset HTTPS_PROXY

# 默认配置
DEFAULT_MODE="development"
DEFAULT_TIMEOUT=60
VERBOSE=false
DRY_RUN=false
AUTO_OPEN=false

# 显示帮助信息
show_help() {
    cat << EOF
OpenPenPal 快速启动脚本

用法: $0 [选项] [模式]

模式:
  development    开发模式 (默认) - 使用Go后端
  production     生产模式 - 使用Go后端
  simple         简化模式 - 使用Go后端
  demo           演示模式 - 使用Go后端
  complete       完整模式 - 使用Go后端
  mock           Mock模式 - 使用Mock服务

选项:
  --timeout=N    服务启动超时时间 (秒，默认: 60)
  --verbose      显示详细输出
  --dry-run      仅显示将要执行的操作，不实际执行
  --auto-open    自动打开浏览器
  --no-deps      跳过依赖安装检查
  --help, -h     显示此帮助信息

示例:
  $0                          # 默认开发模式
  $0 demo --auto-open         # 演示模式并自动打开浏览器
  $0 production --verbose     # 生产模式，详细输出
  $0 simple --dry-run         # 简化模式，预览操作

EOF
}

# 解析命令行参数
parse_arguments() {
    MODE="$DEFAULT_MODE"
    TIMEOUT="$DEFAULT_TIMEOUT"
    SKIP_DEPS=false
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            development|production|simple|demo|complete|mock)
                MODE="$1"
                shift
                ;;
            --timeout=*)
                TIMEOUT="${1#*=}"
                shift
                ;;
            --verbose)
                VERBOSE=true
                export DEBUG="true"
                export LOG_LEVEL="debug"
                shift
                ;;
            --dry-run)
                DRY_RUN=true
                shift
                ;;
            --auto-open)
                AUTO_OPEN=true
                shift
                ;;
            --no-deps)
                SKIP_DEPS=true
                shift
                ;;
            --help|-h)
                show_help
                exit 0
                ;;
            *)
                log_error "未知参数: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    # 设置环境变量
    export NODE_ENV="$MODE"
    
    # 根据模式设置特定配置
    case $MODE in
        demo)
            AUTO_OPEN=true
            ;;
        production)
            export DEBUG="false"
            export LOG_LEVEL="warn"
            # 生产模式使用 PostgreSQL
            export DATABASE_TYPE="postgres"
            log_info "生产模式：使用 PostgreSQL 数据库"
            # 检查配置文件中的全局自动打开设置
            if command -v jq &> /dev/null; then
                local global_auto_open=$(jq -r '.browser.autoOpen // false' "$SCRIPT_DIR/startup-config.json" 2>/dev/null)
                if [ "$global_auto_open" = "true" ] && [ "$AUTO_OPEN" != "true" ]; then
                    AUTO_OPEN=true
                fi
            fi
            ;;
        complete)
            AUTO_OPEN=true
            ;;
        development|simple)
            # 如果配置文件中启用了全局自动打开，则启用
            if command -v jq &> /dev/null; then
                local global_auto_open=$(jq -r '.browser.autoOpen // false' "$SCRIPT_DIR/startup-config.json" 2>/dev/null)
                if [ "$global_auto_open" = "true" ] && [ "$AUTO_OPEN" != "true" ]; then
                    AUTO_OPEN=true
                fi
            fi
            ;;
    esac
}

# 检查系统要求
check_requirements() {
    log_info "检查系统要求..."
    
    # 检查 Node.js
    if ! command -v node &> /dev/null; then
        log_error "Node.js 未安装。请安装 Node.js 18+ 版本。"
        log_info "下载地址: https://nodejs.org/"
        exit 1
    fi
    
    local node_version=$(node --version | sed 's/v//')
    local required_version="18.0.0"
    
    if ! version_gte "$node_version" "$required_version"; then
        log_error "Node.js 版本过低。当前版本: v$node_version，要求: v$required_version+"
        exit 1
    fi
    
    # 检查 npm
    if ! command -v npm &> /dev/null; then
        log_error "npm 未安装。"
        exit 1
    fi
    
    log_success "✓ Node.js $(node --version)"
    log_success "✓ npm $(npm --version)"
}

# 检查和安装依赖
install_dependencies() {
    if [ "$SKIP_DEPS" = true ]; then
        log_info "跳过依赖检查"
        return 0
    fi
    
    log_info "检查项目依赖..."
    
    # 检查前端依赖
    if [ ! -d "$PROJECT_ROOT/frontend/node_modules" ]; then
        log_info "安装前端依赖..."
        cd "$PROJECT_ROOT/frontend"
        run_command "npm install"
        cd "$PROJECT_ROOT"
    fi
    
    # 检查管理后台依赖
    if [ -d "$PROJECT_ROOT/services/admin-service/frontend" ] && [ ! -d "$PROJECT_ROOT/services/admin-service/frontend/node_modules" ]; then
        log_info "安装管理后台依赖..."
        cd "$PROJECT_ROOT/services/admin-service/frontend"
        run_command "npm install"
        cd "$PROJECT_ROOT"
    fi
    
    # 检查Mock服务依赖
    if [ ! -d "$PROJECT_ROOT/apps/mock-services/node_modules" ]; then
        log_info "安装Mock服务依赖..."
        cd "$PROJECT_ROOT/apps/mock-services"
        run_command "npm install"
        cd "$PROJECT_ROOT"
    fi
    
    log_success "依赖检查完成"
}

# 准备环境
prepare_environment() {
    log_info "准备启动环境..."
    
    # 创建必需目录
    create_directories
    
    # 验证环境变量
    if ! validate_environment; then
        log_error "环境变量验证失败"
        exit 1
    fi
    
    # 启动数据库服务（针对生产模式和完整模式）
    if [[ "$MODE" == "production" || "$MODE" == "complete" ]]; then
        log_info "🗄️ 启动数据库服务..."
        if ensure_postgresql; then
            if ensure_database; then
                setup_database_environment
                log_success "数据库服务准备完成"
            else
                log_error "数据库初始化失败"
                exit 1
            fi
        else
            log_error "PostgreSQL 启动失败，无法继续启动微服务"
            log_info "💡 提示：请确保 PostgreSQL 已正确安装"
            log_info "   macOS: brew install postgresql"
            log_info "   Ubuntu: sudo apt-get install postgresql postgresql-contrib"
            exit 1
        fi
        
        # 启动 Redis（可选，不阻止启动）
        ensure_redis
    fi
    
    # 清理旧的PID文件
    rm -f "$LOG_DIR"/*.pid
    
    # 停止可能运行的服务
    log_info "清理可能运行的服务..."
    "$SCRIPT_DIR/stop-all.sh" --quiet --force || true
    
    # 等待端口释放
    sleep 3
    
    # 生产模式下启动基础设施服务
    if [ "$MODE" = "production" ]; then
        start_infrastructure_services
    fi
    
    log_success "环境准备完成"
}

# 启动基础设施服务（数据库、Redis等）
start_infrastructure_services() {
    log_info "启动基础设施服务..."
    
    # 检查Docker是否安装
    local docker_cmd=""
    
    # 尝试多个可能的Docker位置
    if command -v docker &> /dev/null; then
        docker_cmd="docker"
    elif [ -x "/Applications/Docker.app/Contents/Resources/bin/docker" ]; then
        docker_cmd="/Applications/Docker.app/Contents/Resources/bin/docker"
    elif [ -x "$HOME/.docker/bin/docker" ]; then
        docker_cmd="$HOME/.docker/bin/docker"
    else
        log_warning "未找到Docker命令，跳过基础设施服务启动"
        log_info "请确保Docker Desktop已启动，或手动启动PostgreSQL和Redis服务"
        return 0
    fi
    
    # 检查Docker是否正在运行
    if ! $docker_cmd info &> /dev/null; then
        log_warning "Docker未运行，请先启动Docker Desktop"
        return 0
    fi
    
    # 检查docker-compose文件
    if [ -f "$PROJECT_ROOT/docker-compose.yml" ]; then
        log_info "使用docker-compose启动基础设施服务..."
        
        cd "$PROJECT_ROOT"
        
        # 检查docker-compose命令
        local compose_cmd=""
        if command -v docker-compose &> /dev/null; then
            compose_cmd="docker-compose"
        elif $docker_cmd compose version &> /dev/null; then
            compose_cmd="$docker_cmd compose"
        else
            log_warning "未找到docker-compose命令"
            return 0
        fi
        
        # 只启动数据库和缓存服务
        if $compose_cmd up -d postgres 2>/dev/null; then
            log_success "PostgreSQL启动成功"
        else
            log_warning "PostgreSQL启动失败，请检查是否已在运行"
        fi
        
        # 如果有Redis服务定义，也启动它
        if $compose_cmd ps 2>/dev/null | grep -q redis; then
            if $compose_cmd up -d redis 2>/dev/null; then
                log_success "Redis启动成功"
            else
                log_warning "Redis启动失败，请检查是否已在运行"
            fi
        fi
        
        # 等待数据库就绪
        log_info "等待数据库就绪..."
        sleep 5
        
        cd "$PROJECT_ROOT"
    else
        log_info "未找到docker-compose.yml，请手动启动数据库服务"
    fi
}

# 读取服务配置
load_service_config() {
    local config_file="$SCRIPT_DIR/startup-config.json"
    
    if [ ! -f "$config_file" ]; then
        log_error "配置文件不存在: $config_file"
        exit 1
    fi
    
    # 使用 jq 解析配置，如果没有 jq 则使用简化解析
    if command -v jq &> /dev/null; then
        # 使用 jq 解析
        SERVICES=$(jq -r ".modes.${MODE}.services[]" "$config_file" 2>/dev/null | tr '\n' ' ')
        log_debug "jq解析结果: '$SERVICES'"
    else
        # 简化解析（仅支持基本配置）
        case $MODE in
            development)
                SERVICES="go-backend frontend"
                ;;
            production)
                SERVICES="go-backend real-gateway real-write-service real-courier-service real-admin-service real-ocr-service frontend admin-frontend"
                ;;
            simple)
                SERVICES="go-backend frontend"
                ;;
            demo)
                SERVICES="go-backend frontend"
                ;;
            complete)
                SERVICES="go-backend real-gateway real-write-service real-courier-service real-admin-service real-ocr-service frontend admin-frontend"
                ;;
            mock)
                SERVICES="simple-mock frontend"
                ;;
            *)
                SERVICES="go-backend frontend"
                ;;
        esac
    fi
    
    if [ -z "$SERVICES" ]; then
        log_error "无法解析服务配置或配置为空"
        exit 1
    fi
    
    log_debug "启动模式: $MODE"
    log_debug "服务列表: $SERVICES"
}

# 启动单个服务
start_service() {
    local service_name="$1"
    
    log_info "启动服务: $service_name"
    
    case $service_name in
        frontend)
            start_frontend_service
            ;;
        admin-frontend)
            start_admin_frontend_service
            ;;
        go-backend)
            start_go_backend_service
            ;;
        gateway|backend|main-backend|write-service|courier-service|admin-service|ocr-service)
            start_mock_service "$service_name"
            ;;
        simple-mock)
            start_simple_mock_service
            ;;
        real-gateway)
            start_real_gateway_service
            ;;
        real-write-service)
            start_real_write_service
            ;;
        real-courier-service)
            start_real_courier_service
            ;;
        real-admin-service)
            start_real_admin_service
            ;;
        real-ocr-service)
            start_real_ocr_service
            ;;
        *)
            log_error "未知服务: $service_name"
            return 1
            ;;
    esac
}

# 启动前端服务
start_frontend_service() {
    local port=$FRONTEND_PORT
    local service_dir="$PROJECT_ROOT/frontend"
    local log_file="$LOG_DIR/frontend.log"
    local pid_file="$LOG_DIR/frontend.pid"
    
    # 清理可能存在的旧PID文件
    if [ -f "$pid_file" ]; then
        local old_pid=$(cat "$pid_file" 2>/dev/null)
        if [ -n "$old_pid" ]; then
            if ! ps -p "$old_pid" > /dev/null 2>&1; then
                log_debug "清理陈旧的PID文件 (PID: $old_pid)"
                rm -f "$pid_file"
            else
                log_warning "前端服务可能已在运行 (PID: $old_pid)"
                return 1
            fi
        else
            rm -f "$pid_file"
        fi
    fi
    
    if ! check_port_available $port; then
        log_error "端口 $port (前端) 被占用"
        return 1
    fi
    
    cd "$service_dir"
    
    if [ "$DRY_RUN" = true ]; then
        log_info "[DRY RUN] 将启动前端服务: npm run dev"
        return 0
    fi
    
    # 启动前端开发服务器
    nohup npm run dev > "$log_file" 2>&1 &
    local pid=$!
    echo $pid > "$pid_file"
    
    cd "$PROJECT_ROOT"
    
    # 等待服务启动
    if wait_for_service $port "前端服务" $TIMEOUT; then
        log_success "前端服务启动成功 (PID: $pid, 端口: $port)"
        return 0
    else
        log_error "前端服务启动失败"
        # 清理失败的PID文件
        rm -f "$pid_file"
        return 1
    fi
}

# 启动管理后台服务
start_admin_frontend_service() {
    local port=$ADMIN_FRONTEND_PORT
    local service_dir="$PROJECT_ROOT/services/admin-service/frontend"
    local log_file="$LOG_DIR/admin-frontend.log"
    local pid_file="$LOG_DIR/admin-frontend.pid"
    
    if [ ! -d "$service_dir" ]; then
        log_warning "管理后台目录不存在，跳过启动"
        return 0
    fi
    
    if ! check_port_available $port; then
        log_error "端口 $port (管理后台) 被占用"
        return 1
    fi
    
    cd "$service_dir"
    
    if [ "$DRY_RUN" = true ]; then
        log_info "[DRY RUN] 将启动管理后台: npm run dev"
        return 0
    fi
    
    # 启动管理后台
    nohup npm run dev > "$log_file" 2>&1 &
    local pid=$!
    echo $pid > "$pid_file"
    
    cd "$PROJECT_ROOT"
    
    # 等待服务启动
    if wait_for_service $port "管理后台" $TIMEOUT; then
        log_success "管理后台启动成功 (PID: $pid, 端口: $port)"
        return 0
    else
        log_error "管理后台启动失败"
        return 1
    fi
}

# 启动Go后端服务
start_go_backend_service() {
    local port=$BACKEND_PORT
    local service_dir="$PROJECT_ROOT/backend"
    local log_file="$LOG_DIR/go-backend.log"
    local pid_file="$LOG_DIR/go-backend.pid"
    
    # 清理可能存在的旧PID文件
    if [ -f "$pid_file" ]; then
        local old_pid=$(cat "$pid_file" 2>/dev/null)
        if [ -n "$old_pid" ]; then
            if ! ps -p "$old_pid" > /dev/null 2>&1; then
                log_debug "清理陈旧的PID文件 (PID: $old_pid)"
                rm -f "$pid_file"
            else
                log_warning "Go后端可能已在运行 (PID: $old_pid)"
                return 1
            fi
        else
            rm -f "$pid_file"
        fi
    fi
    
    if [ ! -d "$service_dir" ]; then
        log_error "Go后端目录不存在: $service_dir"
        return 1
    fi
    
    if [ ! -f "$service_dir/openpenpal" ] && [ ! -f "$service_dir/openpenpal-backend" ]; then
        log_error "Go后端可执行文件不存在"
        log_info "请先编译Go后端: cd $service_dir && go build -o openpenpal"
        return 1
    fi
    
    # 检查实际的可执行文件名
    local backend_binary="openpenpal"
    if [ -f "$service_dir/openpenpal-backend" ]; then
        backend_binary="openpenpal-backend"
    fi
    
    if ! check_port_available $port; then
        log_error "端口 $port (Go后端) 被占用"
        return 1
    fi
    
    if [ "$DRY_RUN" = true ]; then
        log_info "[DRY RUN] 将启动Go后端: ./openpenpal-backend"
        return 0
    fi
    
    cd "$service_dir"
    
    # 启动Go后端
    nohup ./$backend_binary > "$log_file" 2>&1 &
    local pid=$!
    echo $pid > "$pid_file"
    
    cd "$PROJECT_ROOT"
    
    # 等待服务启动
    if wait_for_service $port "Go后端" $TIMEOUT; then
        log_success "Go后端启动成功 (PID: $pid, 端口: $port)"
        log_info "  • 数据库: PostgreSQL (rocalight@localhost:5432/openpenpal)"
        log_info "  • WebSocket: ws://localhost:$port/api/v1/ws/connect"
        log_info "  • 健康检查: http://localhost:$port/health"
        return 0
    else
        log_error "Go后端启动失败"
        # 清理失败的PID文件
        rm -f "$pid_file"
        return 1
    fi
}

# 启动Mock服务
start_mock_service() {
    local service_name="$1"
    local service_dir="$PROJECT_ROOT/apps/mock-services"
    local log_file="$LOG_DIR/mock-${service_name}.log"
    local pid_file="$LOG_DIR/mock-${service_name}.pid"
    
    # 获取服务端口
    local port
    case $service_name in
        gateway) port=$GATEWAY_PORT ;;
        backend) port=$BACKEND_PORT ;;
        main-backend) port=$BACKEND_PORT ;;
        write-service) port=$WRITE_SERVICE_PORT ;;
        courier-service) port=$COURIER_SERVICE_PORT ;;
        admin-service) port=$ADMIN_SERVICE_PORT ;;
        ocr-service) port=$OCR_SERVICE_PORT ;;
        *) 
            log_error "未知的mock服务: $service_name"
            return 1
            ;;
    esac
    
    if ! check_port_available $port; then
        log_error "端口 $port ($service_name) 被占用"
        return 1
    fi
    
    cd "$service_dir"
    
    if [ "$DRY_RUN" = true ]; then
        log_info "[DRY RUN] 将启动 $service_name: node src/index.js --service $service_name"
        return 0
    fi
    
    # 启动Mock服务
    nohup node src/index.js --service "$service_name" > "$log_file" 2>&1 &
    local pid=$!
    echo $pid > "$pid_file"
    
    cd "$PROJECT_ROOT"
    
    # 等待服务启动
    if wait_for_service $port "$service_name" $TIMEOUT; then
        log_success "$service_name 启动成功 (PID: $pid, 端口: $port)"
        return 0
    else
        log_error "$service_name 启动失败"
        return 1
    fi
}

# 启动简化Mock服务
start_simple_mock_service() {
    local port=$GATEWAY_PORT
    local log_file="$LOG_DIR/simple-mock.log"
    local pid_file="$LOG_DIR/simple-mock.pid"
    
    if ! check_port_available $port; then
        log_error "端口 $port (简化Mock服务) 被占用"
        return 1
    fi
    
    if [ "$DRY_RUN" = true ]; then
        log_info "[DRY RUN] 将启动简化Mock服务: node scripts/simple-mock-services.js"
        return 0
    fi
    
    # 启动简化Mock服务
    cd "$PROJECT_ROOT"
    nohup node scripts/simple-mock-services.js > "$log_file" 2>&1 &
    local pid=$!
    echo $pid > "$pid_file"
    
    # 等待服务启动
    if wait_for_service $port "简化Mock服务" $TIMEOUT; then
        log_success "简化Mock服务启动成功 (PID: $pid, 端口: $port)"
        return 0
    else
        log_error "简化Mock服务启动失败"
        return 1
    fi
}

# 启动真实网关服务
start_real_gateway_service() {
    local port=8000
    local service_dir="$PROJECT_ROOT/services/gateway"
    local log_file="$LOG_DIR/gateway.log"
    local pid_file="$LOG_DIR/gateway.pid"
    
    if [ ! -d "$service_dir" ]; then
        log_error "网关服务目录不存在: $service_dir"
        return 1
    fi
    
    if ! check_port_available $port; then
        log_error "端口 $port (网关) 被占用"
        return 1
    fi
    
    cd "$service_dir"
    
    # 检查Go环境
    if ! command -v go &> /dev/null; then
        log_error "Go未安装，无法编译网关服务"
        return 1
    fi
    
    # 检查是否需要编译
    if [ ! -f "bin/gateway" ]; then
        log_info "编译网关服务..."
        if [ -f "go.mod" ]; then
            go mod tidy
        fi
        if ! go build -o bin/gateway cmd/main.go; then
            log_error "网关服务编译失败"
            return 1
        fi
        chmod +x bin/gateway
    fi
    
    if [ "$DRY_RUN" = true ]; then
        log_info "[DRY RUN] 将启动网关服务: ./bin/gateway"
        return 0
    fi
    
    # 启动网关服务
    nohup ./bin/gateway > "$log_file" 2>&1 &
    local pid=$!
    echo $pid > "$pid_file"
    
    cd "$PROJECT_ROOT"
    
    if wait_for_service $port "网关服务" $TIMEOUT; then
        log_success "网关服务启动成功 (PID: $pid, 端口: $port)"
        return 0
    else
        log_error "网关服务启动失败"
        rm -f "$pid_file"
        return 1
    fi
}

# 启动真实写信服务
start_real_write_service() {
    local port=8001
    local service_dir="$PROJECT_ROOT/services/write-service"
    local log_file="$LOG_DIR/write-service.log"
    local pid_file="$LOG_DIR/write-service.pid"
    
    if [ ! -d "$service_dir" ]; then
        log_error "写信服务目录不存在: $service_dir"
        return 1
    fi
    
    if ! check_port_available $port; then
        log_error "端口 $port (写信服务) 被占用"
        return 1
    fi
    
    cd "$service_dir"
    
    # 检查Python虚拟环境
    if [ ! -d "venv" ]; then
        log_info "创建Python虚拟环境..."
        python3 -m venv venv
        source venv/bin/activate
        pip install -r requirements.txt
    else
        source venv/bin/activate
    fi
    
    if [ "$DRY_RUN" = true ]; then
        log_info "[DRY RUN] 将启动写信服务: python app/main.py"
        return 0
    fi
    
    # 启动写信服务 (设置正确的Python路径和数据库环境变量)
    DATABASE_URL="$DATABASE_URL" DATABASE_TYPE="$DATABASE_TYPE" DB_USER="$DB_USER" DB_PASSWORD="$DB_PASSWORD" \
    DB_HOST="$DB_HOST" DB_PORT="$DB_PORT" DB_NAME="$DATABASE_NAME" \
    nohup python -m app.main > "$log_file" 2>&1 &
    local pid=$!
    echo $pid > "$pid_file"
    
    cd "$PROJECT_ROOT"
    
    if wait_for_service $port "写信服务" $TIMEOUT; then
        log_success "写信服务启动成功 (PID: $pid, 端口: $port)"
        return 0
    else
        log_error "写信服务启动失败"
        rm -f "$pid_file"
        return 1
    fi
}

# 启动真实信使服务
start_real_courier_service() {
    local port=8002
    local service_dir="$PROJECT_ROOT/services/courier-service"
    local log_file="$LOG_DIR/courier-service.log"
    local pid_file="$LOG_DIR/courier-service.pid"
    
    if [ ! -d "$service_dir" ]; then
        log_error "信使服务目录不存在: $service_dir"
        return 1
    fi
    
    if ! check_port_available $port; then
        log_error "端口 $port (信使服务) 被占用"
        return 1
    fi
    
    cd "$service_dir"
    
    # 检查是否需要编译
    if [ ! -f "bin/courier-service" ]; then
        log_info "编译信使服务..."
        if ! go build -o bin/courier-service cmd/main.go; then
            log_error "信使服务编译失败"
            return 1
        fi
    fi
    
    if [ "$DRY_RUN" = true ]; then
        log_info "[DRY RUN] 将启动信使服务: ./bin/courier-service"
        return 0
    fi
    
    # 启动信使服务
    nohup ./bin/courier-service > "$log_file" 2>&1 &
    local pid=$!
    echo $pid > "$pid_file"
    
    cd "$PROJECT_ROOT"
    
    if wait_for_service $port "信使服务" $TIMEOUT; then
        log_success "信使服务启动成功 (PID: $pid, 端口: $port)"
        return 0
    else
        log_error "信使服务启动失败"
        rm -f "$pid_file"
        return 1
    fi
}

# 启动真实管理服务
start_real_admin_service() {
    local port=8003
    local service_dir="$PROJECT_ROOT/services/admin-service/backend"
    local log_file="$LOG_DIR/admin-service.log"
    local pid_file="$LOG_DIR/admin-service.pid"
    
    if [ ! -d "$service_dir" ]; then
        log_error "管理服务目录不存在: $service_dir"
        return 1
    fi
    
    if ! check_port_available $port; then
        log_error "端口 $port (管理服务) 被占用"
        return 1
    fi
    
    cd "$service_dir"
    
    if [ "$DRY_RUN" = true ]; then
        log_info "[DRY RUN] 将启动管理服务: ./mvnw spring-boot:run"
        return 0
    fi
    
    # 检查Java是否可用
    if ! command -v java &> /dev/null; then
        log_error "Java未安装或不可用，跳过管理服务启动"
        log_info "请安装Java 11+: brew install openjdk@11"
        return 1
    fi
    
    # 启动Spring Boot服务
    nohup ./mvnw spring-boot:run > "$log_file" 2>&1 &
    local pid=$!
    echo $pid > "$pid_file"
    
    cd "$PROJECT_ROOT"
    
    if wait_for_service $port "管理服务" $TIMEOUT; then
        log_success "管理服务启动成功 (PID: $pid, 端口: $port)"
        return 0
    else
        log_error "管理服务启动失败"
        rm -f "$pid_file"
        return 1
    fi
}

# 启动真实OCR服务
start_real_ocr_service() {
    local port=8004
    local service_dir="$PROJECT_ROOT/services/ocr-service"
    local log_file="$LOG_DIR/ocr-service.log"
    local pid_file="$LOG_DIR/ocr-service.pid"
    
    if [ ! -d "$service_dir" ]; then
        log_error "OCR服务目录不存在: $service_dir"
        return 1
    fi
    
    if ! check_port_available $port; then
        log_error "端口 $port (OCR服务) 被占用"
        return 1
    fi
    
    cd "$service_dir"
    
    # 检查Python虚拟环境
    if [ ! -d "venv" ]; then
        log_info "创建Python虚拟环境..."
        python3 -m venv venv
        source venv/bin/activate
        pip install -r requirements.txt
    else
        source venv/bin/activate
    fi
    
    if [ "$DRY_RUN" = true ]; then
        log_info "[DRY RUN] 将启动OCR服务: python app/main.py"
        return 0
    fi
    
    # 启动OCR服务 (使用gunicorn避免Flask开发服务器问题)
    export FLASK_ENV=production
    export PYTHONPATH="$service_dir:$PYTHONPATH"
    nohup gunicorn -w 1 -b 0.0.0.0:$port app.main:app --timeout 120 > "$log_file" 2>&1 &
    local pid=$!
    echo $pid > "$pid_file"
    
    cd "$PROJECT_ROOT"
    
    if wait_for_service $port "OCR服务" $TIMEOUT; then
        log_success "OCR服务启动成功 (PID: $pid, 端口: $port)"
        return 0
    else
        log_error "OCR服务启动失败"
        rm -f "$pid_file"
        return 1
    fi
}

# 启动所有服务
start_all_services() {
    log_info "启动所有服务..."
    
    local failed_services=()
    
    for service in $SERVICES; do
        if start_service "$service"; then
            log_debug "$service 启动成功"
        else
            log_error "$service 启动失败"
            failed_services+=("$service")
        fi
        
        # 服务间启动间隔
        sleep 2
    done
    
    if [ ${#failed_services[@]} -eq 0 ]; then
        log_success "所有服务启动成功！"
        return 0
    else
        log_error "以下服务启动失败: ${failed_services[*]}"
        return 1
    fi
}

# 验证服务状态
verify_services() {
    log_info "验证服务状态..."
    
    # 运行状态检查脚本 (暂时禁用有问题的脚本)
    if false && [ -f "$SCRIPT_DIR/check-status.sh" ]; then
        "$SCRIPT_DIR/check-status.sh" --quiet
    else
        # 简单的端口检查
        local all_healthy=true
        
        for service in $SERVICES; do
            local port
            case $service in
                frontend) port=$FRONTEND_PORT ;;
                admin-frontend) port=$ADMIN_FRONTEND_PORT ;;
                gateway|simple-mock) port=$GATEWAY_PORT ;;
                backend|main-backend) port=$BACKEND_PORT ;;
                write-service) port=$WRITE_SERVICE_PORT ;;
                courier-service) port=$COURIER_SERVICE_PORT ;;
                admin-service) port=$ADMIN_SERVICE_PORT ;;
                ocr-service) port=$OCR_SERVICE_PORT ;;
            esac
            
            if [ -n "$port" ]; then
                if check_port_occupied $port; then
                    log_success "✓ $service (端口 $port)"
                else
                    log_error "$service (端口 $port) 未响应"
                    all_healthy=false
                fi
            fi
        done
        
        if [ "$all_healthy" = true ]; then
            log_success "所有服务运行正常"
        else
            log_warning "部分服务存在问题"
        fi
    fi
}

# 打开浏览器
open_browser() {
    log_debug "AUTO_OPEN=$AUTO_OPEN, DRY_RUN=$DRY_RUN, MODE=$MODE"
    
    if [ "$DRY_RUN" = true ]; then
        log_debug "跳过浏览器打开 (DRY_RUN模式)"
        return 0
    fi
    
    # 使用新的URL管理器处理浏览器打开
    if [ "$AUTO_OPEN" = true ]; then
        log_info "🌐 使用SOTA浏览器管理系统打开应用..."
        
        # 等待服务完全启动
        local browser_delay
        browser_delay="$(get_mode_config "$MODE" "browser_delay")"
        if [ -n "$browser_delay" ] && [ "$browser_delay" != "null" ]; then
            log_info "等待服务稳定 (${browser_delay}秒)..."
            sleep "$browser_delay"
        else
            sleep 3
        fi
        
        # 使用URL管理器打开配置的URLs
        if ! open_configured_urls "$MODE"; then
            log_warning "SOTA浏览器管理器失败，尝试回退到传统方式"
            
            # 回退到传统方式
            log_info "打开前端应用: $FRONTEND_URL"
            if open_url "$FRONTEND_URL" "auto" true; then
                log_success "✓ 成功打开前端应用"
            else
                log_info "请手动访问: $FRONTEND_URL"
            fi
            
            # 如果是演示模式，也打开管理后台
            if [ "$MODE" = "demo" ] && echo "$SERVICES" | grep -q "admin-frontend"; then
                sleep 1
                log_info "打开管理后台: $ADMIN_FRONTEND_URL"
                if ! open_url "$ADMIN_FRONTEND_URL" "auto" true; then
                    log_info "请手动访问管理后台: $ADMIN_FRONTEND_URL"
                fi
            fi
        fi
    else
        log_debug "跳过自动打开浏览器 (AUTO_OPEN=$AUTO_OPEN)"
    fi
}

# 显示启动结果
show_result() {
    echo ""
    echo "🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉"
    log_info "🎉 OpenPenPal 启动完成！"
    log_info "✨ 所有服务正在运行中..."
    echo "🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉🎉"
    echo ""
    
    # 使用新的URL管理器显示服务信息
    local show_health
    show_health="$(get_display_config "show_health_status")"
    if [ "$show_health" = "null" ] || [ -z "$show_health" ]; then
        show_health="true"
    fi
    
    show_all_urls "$MODE" "$show_health"
    
    echo ""
    log_info "🔑 测试账号:"
    log_info "  • alice/secret - 学生用户"
    log_info "  • admin/admin123 - 管理员"
    log_info "  • courier_level1/secret - 一级信使"
    log_info "  • courier_level2/secret - 二级信使"
    log_info "  • courier_level3/secret - 三级信使"
    log_info "  • courier_level4/secret - 四级信使"
    echo ""
    log_info "📋 常用命令:"
    log_info "  • 查看状态: ./startup/check-status.sh"
    log_info "  • 验证服务: ./startup/url-manager.sh validate $MODE"
    log_info "  • 查看日志: tail -f logs/*.log"
    log_info "  • 停止服务: ./startup/stop-all.sh"
    echo ""
    
    # 显示模式特定的提示信息
    show_mode_tips
}

# 显示模式特定的提示信息
show_mode_tips() {
    case "$MODE" in
        "demo")
            log_info "🎬 演示模式提示:"
            log_info "  • 已预装测试数据"
            log_info "  • 适合展示和体验功能"
            log_info "  • 重启后数据会重置"
            log_info "  • 自动打开主页面和管理后台"
            echo ""
            ;;
        "development")
            log_info "🚀 开发模式提示:"
            log_info "  • 使用简化Mock服务，启动快速稳定"
            log_info "  • 前端支持热重载，修改代码自动刷新"
            log_info "  • 适合日常开发和调试"
            log_info "  • 自动打开主页面"
            echo ""
            ;;
        "simple")
            log_info "⚡ 简化模式提示:"
            log_info "  • 最小化启动配置，快速体验"
            log_info "  • 只启动核心服务"
            log_info "  • 自动打开主页面"
            echo ""
            ;;
        "complete")
            log_info "🏗️ 完整模式提示:"
            log_info "  • 启动所有微服务，完整功能体验"
            log_info "  • 包含OCR服务、写信服务、信使服务等"
            log_info "  • 适合功能测试和集成测试"
            log_info "  • 自动打开主页面"
            echo ""
            ;;
        "production")
            log_info "🏭 生产模式提示:"
            log_info "  • 优化的生产配置"
            log_info "  • 日志级别为warn，性能优先"
            log_info "  • 适合部署到生产环境"  
            log_info "  • 不会自动打开浏览器"
            echo ""
            ;;
    esac
}

# SOTA运行时管理 - 基于配置的智能运行决策
manage_service_runtime() {
    log_debug "开始运行时管理，模式: $MODE"
    
    # 从配置获取是否需要保持运行
    local keep_running
    keep_running="$(get_mode_config "$MODE" "keep_running")"
    
    if [ "$keep_running" = "null" ] || [ -z "$keep_running" ]; then
        # 配置不存在时的回退逻辑
        log_debug "配置中未找到keep_running设置，使用默认逻辑"
        case "$MODE" in
            "production")
                keep_running="false"
                ;;
            *)
                keep_running="true"
                ;;
        esac
    fi
    
    log_debug "配置决定是否保持运行: $keep_running"
    
    if [ "$keep_running" = "true" ]; then
        log_info "🔄 服务将保持运行"
        log_info "💡 按 Ctrl+C 停止所有服务"
        
        # 设置优雅退出处理
        setup_graceful_shutdown
        
        # 智能保持运行 - 定期检查服务状态
        keep_services_running
    else
        log_info "✅ 启动完成，服务在后台运行"
        log_info "💡 使用 './startup/stop-all.sh' 停止服务"
        
        # Production模式直接退出，不hang
        if [ "$MODE" = "production" ]; then
            log_info "🏭 生产模式启动完成，脚本退出"
        fi
    fi
}

# 设置优雅退出处理
setup_graceful_shutdown() {
    trap 'handle_shutdown' INT TERM
    log_debug "已设置优雅退出处理"
}

# 退出处理函数
handle_shutdown() {
    echo ""
    log_info "📤 收到退出信号，正在优雅停止服务..."
    
    # 显示停止进度
    show_progress "停止服务中" 3 &
    local progress_pid=$!
    
    # 停止服务
    if "$SCRIPT_DIR/stop-all.sh" --quiet; then
        kill $progress_pid 2>/dev/null || true
        log_success "✅ 所有服务已停止"
    else
        kill $progress_pid 2>/dev/null || true
        log_error "❌ 停止服务时出现错误"
    fi
    
    exit 0
}

# 验证当前模式下实际运行的服务
validate_running_services() {
    local mode="$1"
    local timeout="${2:-5}"
    
    log_debug "验证 $mode 模式下的服务状态"
    
    local all_healthy=true
    local services_to_check=""
    
    # 根据模式确定要检查的服务
    # 格式: service_name:port:health_path:required
    # required: true表示必须运行，false表示可选
    case $mode in
        production)
            # 生产模式：检查所有服务
            services_to_check="go-backend:8080:/health:true frontend:3000:/health:true"
            
            # 检查真实服务（如果启动了的话）
            if [ -f "$LOG_DIR/gateway.pid" ]; then
                services_to_check="$services_to_check real-gateway:8000:/health:true"
            fi
            if [ -f "$LOG_DIR/write-service.pid" ]; then
                services_to_check="$services_to_check real-write-service:8001:/health:true"
            fi
            if [ -f "$LOG_DIR/courier-service.pid" ]; then
                services_to_check="$services_to_check real-courier-service:8002:/health:true"
            fi
            if [ -f "$LOG_DIR/admin-service.pid" ]; then
                services_to_check="$services_to_check real-admin-service:8003:/health:false"
            fi
            if [ -f "$LOG_DIR/ocr-service.pid" ]; then
                services_to_check="$services_to_check real-ocr-service:8004:/health:false"
            fi
            if [ -f "$LOG_DIR/admin-frontend.pid" ]; then
                services_to_check="$services_to_check admin-frontend:3001:/health:false"
            fi
            ;;
        development|simple|demo|complete)
            # 其他模式：只检查后端和前端
            services_to_check="go-backend:8080:/health:true frontend:3000:/health:true"
            ;;
        mock)
            # Mock模式：检查mock服务和前端
            services_to_check="simple-mock:8000:/health:true frontend:3000:/health:true"
            ;;
        *)
            services_to_check="go-backend:8080:/health:true frontend:3000:/health:true"
            ;;
    esac
    
    # 检查每个服务
    for service_info in $services_to_check; do
        IFS=':' read -r service_name port health_path required <<< "$service_info"
        local url="http://localhost:$port"
        
        # 默认required为true
        if [ -z "$required" ]; then
            required="true"
        fi
        
        if check_url_health "$url" "$health_path" "$timeout"; then
            log_debug "✓ $service_name ($port) 健康"
        else
            if [ "$required" = "true" ]; then
                log_debug "✗ $service_name ($port) 不可用 [必需服务]"
                all_healthy=false
            else
                log_debug "✗ $service_name ($port) 不可用 [可选服务]"
                # 可选服务不影响整体健康状态
            fi
        fi
    done
    
    if [ "$all_healthy" = true ]; then
        return 0
    else
        return 1
    fi
}

# 智能保持运行 - 定期检查服务状态
keep_services_running() {
    local check_interval=30
    local last_health_check=0
    local consecutive_failures=0
    local max_failures=3
    
    log_debug "开始智能运行循环，检查间隔: ${check_interval}秒"
    
    while true; do
        sleep 10
        
        # 定期健康检查
        local current_time=$(date +%s)
        if [ $((current_time - last_health_check)) -ge $check_interval ]; then
            log_debug "执行定期健康检查"
            
            # 只检查当前模式实际启动的服务
            if validate_running_services "$MODE" 5; then
                consecutive_failures=0
                log_debug "✓ 健康检查通过"
            else
                consecutive_failures=$((consecutive_failures + 1))
                log_warning "健康检查失败 ($consecutive_failures/$max_failures)"
                
                if [ $consecutive_failures -ge $max_failures ]; then
                    log_error "连续健康检查失败，可能存在问题"
                    log_info "建议检查服务状态: ./startup/check-status.sh"
                    consecutive_failures=0  # 重置计数器，避免频繁报告
                fi
            fi
            
            last_health_check=$current_time
        fi
    done
}

# 主函数
main() {
    # 解析参数
    parse_arguments "$@"
    
    # 显示启动信息
    log_info "🚀 OpenPenPal 快速启动器"
    log_info "模式: $MODE"
    if [ "$VERBOSE" = true ]; then
        show_environment
    fi
    
    # 加载服务配置（干运行模式也需要）
    load_service_config
    
    # 预运行检查
    if [ "$DRY_RUN" = true ]; then
        log_info "========== DRY RUN 模式 =========="
        log_info "将要执行的操作:"
        log_info "1. 检查系统要求"
        log_info "2. 安装依赖 (如果需要)"
        log_info "3. 准备环境"
        log_info "4. 启动服务: $SERVICES"
        log_info "5. 验证服务状态"
        if [ "$AUTO_OPEN" = true ]; then
            log_info "6. 打开浏览器"
        fi
        log_info "================================="
        return 0
    fi
    
    # 执行启动流程
    check_requirements
    install_dependencies
    prepare_environment
    
    if start_all_services; then
        verify_services
        open_browser
        show_result
        
        # 使用配置驱动的运行管理
        manage_service_runtime
    else
        log_error "启动失败，请检查日志文件"
        exit 1
    fi
}

# 执行主函数
main "$@"