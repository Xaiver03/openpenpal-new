#!/bin/bash

# OpenPenPal Mock Services 启动脚本
# 提供多种启动模式和配置选项

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 项目根目录
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MOCK_SERVICES_DIR="$PROJECT_ROOT/apps/mock-services"

# 默认配置
DEFAULT_MODE="all"
DEFAULT_LOG_LEVEL="info"
DEFAULT_ENV="development"

# 显示帮助信息
show_help() {
    echo -e "${BLUE}OpenPenPal Mock Services 启动脚本${NC}"
    echo -e "${CYAN}统一管理所有微服务的 Mock 实例${NC}"
    echo ""
    echo "用法: $0 [模式] [选项]"
    echo ""
    echo -e "${YELLOW}启动模式:${NC}"
    echo "  all              启动所有服务 (默认)"
    echo "  gateway          只启动 API Gateway"
    echo "  write            只启动写信服务"
    echo "  courier          只启动信使服务"
    echo "  admin            只启动管理服务"
    echo "  backend          只启动主后端服务"
    echo "  ocr              只启动 OCR 服务"
    echo "  custom           自定义服务组合"
    echo ""
    echo -e "${YELLOW}选项:${NC}"
    echo "  --env=<env>      设置环境 (development|production, 默认: development)"
    echo "  --log=<level>    设置日志级别 (debug|info|warn|error, 默认: info)"
    echo "  --port=<number>  指定端口 (仅单服务模式)"
    echo "  --watch          启用文件监听模式 (开发环境)"
    echo "  --install        安装依赖包"
    echo "  --test           运行测试"
    echo "  --stop           停止所有运行中的服务"
    echo "  --status         查看服务状态"
    echo ""
    echo -e "${YELLOW}示例:${NC}"
    echo "  $0                              # 启动所有服务"
    echo "  $0 gateway                      # 只启动 API Gateway"
    echo "  $0 write --watch                # 启动写信服务并监听文件变化"
    echo "  $0 all --env=production         # 生产模式启动所有服务"
    echo "  $0 custom --services=gateway,write  # 自定义启动 Gateway 和写信服务"
    echo "  $0 --install                    # 安装依赖"
    echo "  $0 --test                       # 运行测试"
    echo "  $0 --stop                       # 停止所有服务"
    echo ""
}

# 日志函数
log_info() {
    echo -e "${CYAN}[INFO]${NC} $1"
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

# 检查依赖
check_dependencies() {
    log_info "检查依赖..."
    
    if ! command -v node &> /dev/null; then
        log_error "Node.js 未安装，请先安装 Node.js 18+"
        exit 1
    fi
    
    local node_version=$(node --version | cut -d'v' -f2 | cut -d'.' -f1)
    if [ "$node_version" -lt 18 ]; then
        log_error "Node.js 版本过低，需要 18+，当前版本: $(node --version)"
        exit 1
    fi
    
    if ! command -v npm &> /dev/null; then
        log_error "npm 未安装"
        exit 1
    fi
    
    log_success "依赖检查通过"
}

# 安装依赖包
install_dependencies() {
    log_info "安装 Mock Services 依赖..."
    
    if [ ! -d "$MOCK_SERVICES_DIR" ]; then
        log_error "Mock Services 目录不存在: $MOCK_SERVICES_DIR"
        exit 1
    fi
    
    cd "$MOCK_SERVICES_DIR"
    
    if [ ! -f "package.json" ]; then
        log_error "package.json 不存在"
        exit 1
    fi
    
    npm install
    
    log_success "依赖安装完成"
}

# 检查服务状态
check_service_status() {
    log_info "检查服务状态..."
    
    local services=("8000:gateway" "8080:main-backend" "8001:write-service" "8002:courier-service" "8003:admin-service" "8004:ocr-service")
    local running_services=0
    
    echo ""
    echo -e "${BLUE}端口占用情况:${NC}"
    echo "----------------------------------------"
    
    for service in "${services[@]}"; do
        local port=$(echo $service | cut -d':' -f1)
        local name=$(echo $service | cut -d':' -f2)
        
        if lsof -i :$port >/dev/null 2>&1; then
            local pid=$(lsof -t -i :$port | head -n1)
            echo -e "  ${GREEN}✓${NC} $name (端口 $port) - PID: $pid"
            running_services=$((running_services + 1))
        else
            echo -e "  ${RED}✗${NC} $name (端口 $port) - 未运行"
        fi
    done
    
    echo "----------------------------------------"
    echo -e "运行中的服务: ${GREEN}$running_services${NC}/6"
    echo ""
}

# 停止所有服务
stop_all_services() {
    log_info "停止所有 Mock 服务..."
    
    local ports=(8000 8080 8001 8002 8003 8004)
    local stopped_count=0
    
    for port in "${ports[@]}"; do
        if lsof -i :$port >/dev/null 2>&1; then
            local pids=$(lsof -t -i :$port)
            for pid in $pids; do
                if kill -TERM $pid 2>/dev/null; then
                    log_info "停止端口 $port 上的进程 (PID: $pid)"
                    stopped_count=$((stopped_count + 1))
                fi
            done
        fi
    done
    
    # 等待进程优雅退出
    sleep 2
    
    # 强制杀死仍在运行的进程
    for port in "${ports[@]}"; do
        if lsof -i :$port >/dev/null 2>&1; then
            local pids=$(lsof -t -i :$port)
            for pid in $pids; do
                if kill -KILL $pid 2>/dev/null; then
                    log_warning "强制停止端口 $port 上的进程 (PID: $pid)"
                fi
            done
        fi
    done
    
    if [ $stopped_count -gt 0 ]; then
        log_success "已停止 $stopped_count 个服务进程"
    else
        log_info "没有运行中的服务需要停止"
    fi
}

# 运行测试
run_tests() {
    log_info "运行 Mock Services 测试..."
    
    cd "$MOCK_SERVICES_DIR"
    
    if [ ! -f "package.json" ]; then
        log_error "package.json 不存在，请先安装依赖"
        exit 1
    fi
    
    # 检查是否有测试脚本
    if npm run | grep -q "test"; then
        npm test
    else
        log_warning "未找到测试脚本，跳过测试"
    fi
    
    # 运行权限测试
    if npm run | grep -q "test:permissions"; then
        npm run test:permissions
    fi
}

# 启动服务
start_services() {
    local mode="$1"
    local options="$2"
    
    log_info "启动模式: $mode"
    
    cd "$MOCK_SERVICES_DIR"
    
    # 构建启动命令
    local start_cmd="node src/index.js"
    
    # 解析选项
    local services=""
    local port=""
    local log_level="$DEFAULT_LOG_LEVEL"
    local watch_mode=false
    
    # 解析传入的选项
    IFS=' ' read -ra OPTS <<< "$options"
    for opt in "${OPTS[@]}"; do
        case $opt in
            --env=*)
                local env="${opt#*=}"
                export NODE_ENV="$env"
                ;;
            --log=*)
                log_level="${opt#*=}"
                ;;
            --port=*)
                port="${opt#*=}"
                ;;
            --watch)
                watch_mode=true
                ;;
            --services=*)
                services="${opt#*=}"
                ;;
        esac
    done
    
    # 根据模式设置服务
    case $mode in
        all)
            log_info "启动所有 Mock 服务"
            ;;
        gateway)
            start_cmd="$start_cmd --service gateway"
            ;;
        write)
            start_cmd="$start_cmd --service write-service"
            ;;
        courier)
            start_cmd="$start_cmd --service courier-service"
            ;;
        admin)
            start_cmd="$start_cmd --service admin-service"
            ;;
        backend)
            start_cmd="$start_cmd --service main-backend"
            ;;
        ocr)
            start_cmd="$start_cmd --service ocr-service"
            ;;
        custom)
            if [ -n "$services" ]; then
                IFS=',' read -ra SERVICE_ARRAY <<< "$services"
                for service in "${SERVICE_ARRAY[@]}"; do
                    start_cmd="$start_cmd --service $service"
                done
            else
                log_error "自定义模式需要指定 --services 参数"
                exit 1
            fi
            ;;
        *)
            log_error "未知启动模式: $mode"
            exit 1
            ;;
    esac
    
    # 添加其他选项
    if [ -n "$port" ]; then
        start_cmd="$start_cmd --port $port"
    fi
    
    if [ -n "$log_level" ]; then
        start_cmd="$start_cmd --log-level $log_level"
    fi
    
    log_info "执行命令: $start_cmd"
    
    # 启动服务
    if [ "$watch_mode" = true ]; then
        log_info "启用文件监听模式..."
        if command -v nodemon &> /dev/null; then
            nodemon --exec "$start_cmd"
        else
            log_warning "nodemon 未安装，使用 node --watch"
            node --watch $start_cmd
        fi
    else
        $start_cmd
    fi
}

# 解析命令行参数
parse_arguments() {
    MODE="$DEFAULT_MODE"
    OPTIONS=""
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --install)
                install_dependencies
                exit 0
                ;;
            --test)
                run_tests
                exit 0
                ;;
            --stop)
                stop_all_services
                exit 0
                ;;
            --status)
                check_service_status
                exit 0
                ;;
            --help|-h)
                show_help
                exit 0
                ;;
            all|gateway|write|courier|admin|backend|ocr|custom)
                MODE="$1"
                shift
                ;;
            --*)
                OPTIONS="$OPTIONS $1"
                shift
                ;;
            *)
                log_error "未知参数: $1"
                show_help
                exit 1
                ;;
        esac
    done
}

# 主函数
main() {
    echo -e "${BLUE}OpenPenPal Mock Services 启动器${NC}"
    echo -e "${PURPLE}========================================${NC}"
    echo ""
    
    # 检查依赖
    check_dependencies
    
    # 检查 Mock Services 目录是否存在
    if [ ! -d "$MOCK_SERVICES_DIR" ]; then
        log_error "Mock Services 目录不存在: $MOCK_SERVICES_DIR"
        log_info "请先运行: $0 --install"
        exit 1
    fi
    
    # 检查是否已安装依赖
    if [ ! -d "$MOCK_SERVICES_DIR/node_modules" ]; then
        log_warning "依赖未安装，正在自动安装..."
        install_dependencies
    fi
    
    # 解析参数
    parse_arguments "$@"
    
    # 启动服务
    start_services "$MODE" "$OPTIONS"
}

# 捕获退出信号
trap 'echo -e "\n${YELLOW}正在停止服务...${NC}"; exit 0' INT TERM

# 执行主函数
main "$@"