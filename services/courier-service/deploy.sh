#!/bin/bash

# OpenPenPal 信使服务部署脚本

set -e

# 颜色定义
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

# 检查Docker是否安装
check_docker() {
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed. Please install Docker first."
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
        log_error "Docker Compose is not installed. Please install Docker Compose first."
        exit 1
    fi
}

# 检查环境变量文件
check_env() {
    if [ ! -f ".env" ]; then
        if [ -f ".env.example" ]; then
            log_warning ".env file not found. Copying from .env.example"
            cp .env.example .env
            log_warning "Please edit .env file with your configuration"
        else
            log_error ".env file not found and .env.example doesn't exist"
            exit 1
        fi
    fi
}

# 构建服务
build_service() {
    log_info "Building courier service..."
    docker-compose build courier-service
    log_success "Service built successfully"
}

# 启动服务
start_services() {
    log_info "Starting all services..."
    docker-compose up -d
    
    # 等待服务启动
    log_info "Waiting for services to be ready..."
    sleep 10
    
    # 检查服务健康状态
    check_health
}

# 停止服务
stop_services() {
    log_info "Stopping all services..."
    docker-compose down
    log_success "Services stopped"
}

# 重启服务
restart_services() {
    log_info "Restarting all services..."
    docker-compose restart
    log_success "Services restarted"
}

# 查看日志
view_logs() {
    if [ -n "$2" ]; then
        docker-compose logs -f "$2"
    else
        docker-compose logs -f
    fi
}

# 检查健康状态
check_health() {
    log_info "Checking service health..."
    
    services=("postgres" "redis" "courier-service")
    for service in "${services[@]}"; do
        if docker-compose ps "$service" | grep -q "Up"; then
            log_success "$service is running"
        else
            log_error "$service is not running"
        fi
    done
    
    # 检查API健康状态
    sleep 5
    if curl -f http://localhost:8002/health &> /dev/null; then
        log_success "Courier service API is healthy"
    else
        log_warning "Courier service API is not responding"
    fi
}

# 清理资源
cleanup() {
    log_info "Cleaning up..."
    docker-compose down -v
    docker system prune -f
    log_success "Cleanup completed"
}

# 数据库迁移
migrate_db() {
    log_info "Running database migrations..."
    docker-compose exec courier-service ./courier-service -migrate
    log_success "Database migration completed"
}

# 查看服务状态
status() {
    log_info "Service status:"
    docker-compose ps
    
    log_info "Resource usage:"
    docker stats --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.NetIO}}"
}

# 备份数据
backup() {
    backup_dir="backups/$(date +%Y%m%d_%H%M%S)"
    mkdir -p "$backup_dir"
    
    log_info "Creating database backup..."
    docker-compose exec postgres pg_dump -U postgres openpenpal > "$backup_dir/database.sql"
    
    log_info "Creating Redis backup..."
    docker-compose exec redis redis-cli BGSAVE
    docker cp "$(docker-compose ps -q redis):/data/dump.rdb" "$backup_dir/redis.rdb"
    
    log_success "Backup created in $backup_dir"
}

# 恢复数据
restore() {
    if [ -z "$2" ]; then
        log_error "Please specify backup directory"
        exit 1
    fi
    
    backup_dir="$2"
    if [ ! -d "$backup_dir" ]; then
        log_error "Backup directory not found: $backup_dir"
        exit 1
    fi
    
    log_info "Restoring database..."
    docker-compose exec -T postgres psql -U postgres openpenpal < "$backup_dir/database.sql"
    
    log_info "Restoring Redis..."
    docker cp "$backup_dir/redis.rdb" "$(docker-compose ps -q redis):/data/dump.rdb"
    docker-compose restart redis
    
    log_success "Restore completed"
}

# 显示帮助信息
show_help() {
    echo "OpenPenPal Courier Service Deployment Script"
    echo ""
    echo "Usage: $0 [COMMAND]"
    echo ""
    echo "Commands:"
    echo "  build         Build the service"
    echo "  start         Start all services"
    echo "  stop          Stop all services"
    echo "  restart       Restart all services"
    echo "  logs [service] View logs (optionally for specific service)"
    echo "  status        Show service status"
    echo "  health        Check service health"
    echo "  migrate       Run database migrations"
    echo "  backup        Create backup"
    echo "  restore [dir] Restore from backup"
    echo "  cleanup       Clean up all resources"
    echo "  help          Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 start                 # Start all services"
    echo "  $0 logs courier-service  # View courier service logs"
    echo "  $0 backup               # Create backup"
    echo "  $0 restore backups/20231120_143000  # Restore from backup"
}

# 主逻辑
main() {
    case "${1:-help}" in
        "build")
            check_docker
            check_env
            build_service
            ;;
        "start")
            check_docker
            check_env
            start_services
            ;;
        "stop")
            stop_services
            ;;
        "restart")
            restart_services
            ;;
        "logs")
            view_logs "$@"
            ;;
        "status")
            status
            ;;
        "health")
            check_health
            ;;
        "migrate")
            migrate_db
            ;;
        "backup")
            backup
            ;;
        "restore")
            restore "$@"
            ;;
        "cleanup")
            cleanup
            ;;
        "help"|*)
            show_help
            ;;
    esac
}

# 执行主函数
main "$@"