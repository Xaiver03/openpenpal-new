#!/bin/bash

# OCR服务部署脚本
# 用法: ./deploy.sh [dev|prod] [action]

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

# 显示帮助信息
show_help() {
    echo "OCR服务部署脚本"
    echo ""
    echo "用法: $0 [ENVIRONMENT] [ACTION]"
    echo ""
    echo "环境:"
    echo "  dev     开发环境 (默认)"
    echo "  prod    生产环境"
    echo ""
    echo "操作:"
    echo "  up      启动服务 (默认)"
    echo "  down    停止服务"
    echo "  restart 重启服务"
    echo "  build   重新构建镜像"
    echo "  logs    查看日志"
    echo "  status  查看服务状态"
    echo "  clean   清理资源"
    echo ""
    echo "示例:"
    echo "  $0 dev up       # 启动开发环境"
    echo "  $0 prod build   # 构建生产环境镜像"
    echo "  $0 logs         # 查看日志"
}

# 环境和操作参数
ENVIRONMENT=${1:-dev}
ACTION=${2:-up}

# 检查Docker是否安装
check_docker() {
    if ! command -v docker &> /dev/null; then
        log_error "Docker未安装，请先安装Docker"
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose未安装，请先安装Docker Compose"
        exit 1
    fi
}

# 创建必要的目录
create_directories() {
    log_info "创建必要的目录..."
    mkdir -p uploads models logs
    log_success "目录创建完成"
}

# 检查环境变量
check_env_vars() {
    if [ "$ENVIRONMENT" = "prod" ]; then
        if [ -z "$JWT_SECRET" ]; then
            log_warning "生产环境建议设置JWT_SECRET环境变量"
            export JWT_SECRET="default-production-secret-please-change"
        fi
    fi
}

# 选择docker-compose文件
get_compose_file() {
    if [ "$ENVIRONMENT" = "prod" ]; then
        echo "docker-compose.yml"
    else
        echo "docker-compose.dev.yml"
    fi
}

# 启动服务
start_services() {
    local compose_file=$(get_compose_file)
    log_info "启动$ENVIRONMENT环境服务..."
    
    docker-compose -f $compose_file up -d
    
    log_success "服务启动完成"
    log_info "等待服务就绪..."
    sleep 10
    
    # 检查服务健康状态
    check_health
}

# 停止服务
stop_services() {
    local compose_file=$(get_compose_file)
    log_info "停止$ENVIRONMENT环境服务..."
    
    docker-compose -f $compose_file down
    
    log_success "服务停止完成"
}

# 重启服务
restart_services() {
    log_info "重启$ENVIRONMENT环境服务..."
    stop_services
    start_services
}

# 构建镜像
build_images() {
    local compose_file=$(get_compose_file)
    log_info "构建$ENVIRONMENT环境镜像..."
    
    docker-compose -f $compose_file build --no-cache
    
    log_success "镜像构建完成"
}

# 查看日志
show_logs() {
    local compose_file=$(get_compose_file)
    log_info "显示$ENVIRONMENT环境日志..."
    
    docker-compose -f $compose_file logs -f --tail=100
}

# 查看服务状态
show_status() {
    local compose_file=$(get_compose_file)
    log_info "$ENVIRONMENT环境服务状态:"
    
    docker-compose -f $compose_file ps
    
    echo ""
    log_info "Docker镜像:"
    docker images | grep -E "(ocr|redis)" || echo "未找到相关镜像"
    
    echo ""
    log_info "Docker网络:"
    docker network ls | grep ocr || echo "未找到OCR网络"
}

# 检查服务健康状态
check_health() {
    log_info "检查服务健康状态..."
    
    # 检查OCR服务
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if curl -s http://localhost:8004/health > /dev/null; then
            log_success "OCR服务健康检查通过"
            break
        fi
        
        if [ $attempt -eq $max_attempts ]; then
            log_error "OCR服务健康检查失败"
            return 1
        fi
        
        log_info "等待OCR服务启动... ($attempt/$max_attempts)"
        sleep 2
        ((attempt++))
    done
    
    # 检查Redis连接
    if docker exec -it ocr-redis redis-cli ping > /dev/null 2>&1; then
        log_success "Redis连接正常"
    else
        log_warning "Redis连接检查失败"
    fi
    
    # 显示服务信息
    echo ""
    log_success "服务部署完成！"
    echo ""
    echo "访问地址:"
    echo "  OCR API: http://localhost:8004"
    echo "  健康检查: http://localhost:8004/health"
    if [ "$ENVIRONMENT" = "dev" ]; then
        echo "  Redis管理: http://localhost:8081"
    fi
    echo ""
}

# 清理资源
clean_resources() {
    local compose_file=$(get_compose_file)
    log_warning "这将删除所有相关的Docker资源，是否继续? (y/N)"
    read -r response
    
    if [[ "$response" =~ ^[Yy]$ ]]; then
        log_info "清理$ENVIRONMENT环境资源..."
        
        # 停止并删除容器
        docker-compose -f $compose_file down -v --remove-orphans
        
        # 删除镜像
        docker images | grep ocr | awk '{print $3}' | xargs -r docker rmi -f
        
        # 删除网络
        docker network ls | grep ocr | awk '{print $1}' | xargs -r docker network rm
        
        log_success "资源清理完成"
    else
        log_info "取消清理操作"
    fi
}

# 主函数
main() {
    # 检查参数
    if [ "$1" = "-h" ] || [ "$1" = "--help" ]; then
        show_help
        exit 0
    fi
    
    # 验证环境参数
    if [ "$ENVIRONMENT" != "dev" ] && [ "$ENVIRONMENT" != "prod" ]; then
        log_error "无效的环境参数: $ENVIRONMENT"
        show_help
        exit 1
    fi
    
    log_info "OCR服务部署 - 环境: $ENVIRONMENT, 操作: $ACTION"
    
    # 检查依赖
    check_docker
    create_directories
    check_env_vars
    
    # 执行操作
    case $ACTION in
        up|start)
            start_services
            ;;
        down|stop)
            stop_services
            ;;
        restart)
            restart_services
            ;;
        build)
            build_images
            ;;
        logs)
            show_logs
            ;;
        status)
            show_status
            ;;
        clean)
            clean_resources
            ;;
        *)
            log_error "无效的操作: $ACTION"
            show_help
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"