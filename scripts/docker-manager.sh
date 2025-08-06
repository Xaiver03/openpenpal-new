#!/bin/bash

# OpenPenPal Docker Manager Script
# 用于管理多环境Docker服务的统一脚本

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

# 显示帮助信息
show_help() {
    echo -e "${BLUE}OpenPenPal Docker Manager${NC}"
    echo -e "${CYAN}用于管理OpenPenPal项目的Docker服务${NC}"
    echo ""
    echo "用法: $0 <command> [options]"
    echo ""
    echo -e "${YELLOW}可用命令:${NC}"
    echo "  dev              启动开发环境基础设施"
    echo "  dev-full         启动完整开发环境(包括所有服务)"
    echo "  prod             启动生产环境微服务"
    echo "  stop             停止指定环境服务"
    echo "  restart          重启指定环境服务"
    echo "  logs             查看指定服务日志"
    echo "  status           查看服务状态"
    echo "  clean            清理Docker资源"
    echo "  build            构建服务镜像"
    echo "  health           检查服务健康状态"
    echo "  scale            扩展服务实例"
    echo ""
    echo -e "${YELLOW}环境选项:${NC}"
    echo "  --env=<dev|prod>     指定环境 (默认: dev)"
    echo "  --service=<name>     指定特定服务"
    echo "  --force              强制执行操作"
    echo "  --no-build           跳过构建步骤"
    echo ""
    echo -e "${YELLOW}示例:${NC}"
    echo "  $0 dev                          # 启动开发环境基础设施"
    echo "  $0 prod --no-build              # 启动生产环境(跳过构建)"
    echo "  $0 logs --env=dev --service=redis-dev   # 查看开发环境Redis日志"
    echo "  $0 scale --env=prod --service=gateway --replicas=3  # 扩展网关服务到3个实例"
    echo ""
}

# 检查Docker和Docker Compose是否安装
check_dependencies() {
    if ! command -v docker &> /dev/null; then
        echo -e "${RED}错误: Docker 未安装${NC}"
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        echo -e "${RED}错误: Docker Compose 未安装${NC}"
        exit 1
    fi
}

# 解析命令行参数
parse_arguments() {
    COMMAND=""
    ENVIRONMENT="dev"
    SERVICE=""
    FORCE=false
    NO_BUILD=false
    REPLICAS=""
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --env=*)
                ENVIRONMENT="${1#*=}"
                shift
                ;;
            --service=*)
                SERVICE="${1#*=}"
                shift
                ;;
            --replicas=*)
                REPLICAS="${1#*=}"
                shift
                ;;
            --force)
                FORCE=true
                shift
                ;;
            --no-build)
                NO_BUILD=true
                shift
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                if [[ -z "$COMMAND" ]]; then
                    COMMAND="$1"
                else
                    echo -e "${RED}错误: 未知参数 $1${NC}"
                    exit 1
                fi
                shift
                ;;
        esac
    done
}

# 获取Docker Compose文件路径
get_compose_file() {
    case $ENVIRONMENT in
        dev)
            echo "$PROJECT_ROOT/docker-compose.dev.yml"
            ;;
        prod)
            echo "$PROJECT_ROOT/docker-compose.microservices.yml"
            ;;
        *)
            echo -e "${RED}错误: 未知环境 $ENVIRONMENT${NC}"
            exit 1
            ;;
    esac
}

# 启动开发环境基础设施
start_dev() {
    echo -e "${GREEN}启动开发环境基础设施...${NC}"
    local compose_file=$(get_compose_file)
    
    cd "$PROJECT_ROOT"
    docker-compose -f "$compose_file" up -d postgres-dev redis-dev pgadmin-dev redis-commander-dev mailhog-dev
    
    echo -e "${GREEN}开发环境基础设施启动完成!${NC}"
    echo -e "${CYAN}访问地址:${NC}"
    echo -e "  PostgreSQL: localhost:5433"
    echo -e "  Redis: localhost:6380"  
    echo -e "  pgAdmin: http://localhost:5050"
    echo -e "  Redis Commander: http://localhost:8081"
    echo -e "  MailHog: http://localhost:8025"
}

# 启动完整开发环境
start_dev_full() {
    echo -e "${GREEN}启动完整开发环境...${NC}"
    local compose_file=$(get_compose_file)
    
    # 先启动基础设施
    start_dev
    
    echo -e "${YELLOW}等待基础设施就绪...${NC}"
    sleep 5
    
    # 启动所有微服务
    cd "$PROJECT_ROOT"
    if [[ "$NO_BUILD" == false ]]; then
        echo -e "${YELLOW}构建服务镜像...${NC}"
        docker-compose -f "$compose_file" build
    fi
    
    docker-compose -f "$compose_file" up -d
    
    echo -e "${GREEN}完整开发环境启动完成!${NC}"
    show_service_urls
}

# 启动生产环境
start_prod() {
    echo -e "${GREEN}启动生产环境微服务...${NC}"
    local compose_file=$(get_compose_file)
    
    cd "$PROJECT_ROOT"
    
    if [[ "$NO_BUILD" == false ]]; then
        echo -e "${YELLOW}构建生产镜像...${NC}"
        docker-compose -f "$compose_file" build
    fi
    
    # 先启动基础设施
    echo -e "${YELLOW}启动基础设施...${NC}"
    docker-compose -f "$compose_file" up -d postgres redis
    
    echo -e "${YELLOW}等待基础设施就绪...${NC}"
    sleep 10
    
    # 启动所有微服务
    echo -e "${YELLOW}启动微服务...${NC}"
    docker-compose -f "$compose_file" up -d
    
    echo -e "${GREEN}生产环境启动完成!${NC}"
    show_service_urls
}

# 显示服务访问地址
show_service_urls() {
    echo -e "${CYAN}服务访问地址:${NC}"
    case $ENVIRONMENT in
        dev)
            echo -e "  主前端: http://localhost:3000"
            echo -e "  管理后台: http://localhost:3001"
            echo -e "  API网关: http://localhost:8000"
            echo -e "  主后端: http://localhost:8080"
            echo -e "  写信服务: http://localhost:8001"
            echo -e "  信使服务: http://localhost:8002"
            echo -e "  管理服务: http://localhost:8003"
            echo -e "  OCR服务: http://localhost:8004"
            ;;
        prod)
            echo -e "  主入口: http://localhost:80"
            echo -e "  HTTPS: https://localhost:443"
            echo -e "  管理后台: http://localhost:3001"
            echo -e "  监控面板: http://localhost:3002"
            echo -e "  API网关: http://localhost:8000"
            ;;
    esac
}

# 停止服务
stop_services() {
    echo -e "${YELLOW}停止 $ENVIRONMENT 环境服务...${NC}"
    local compose_file=$(get_compose_file)
    
    cd "$PROJECT_ROOT"
    if [[ -n "$SERVICE" ]]; then
        docker-compose -f "$compose_file" stop "$SERVICE"
        echo -e "${GREEN}服务 $SERVICE 已停止${NC}"
    else
        docker-compose -f "$compose_file" down
        echo -e "${GREEN}所有服务已停止${NC}"
    fi
}

# 重启服务
restart_services() {
    echo -e "${YELLOW}重启 $ENVIRONMENT 环境服务...${NC}"
    local compose_file=$(get_compose_file)
    
    cd "$PROJECT_ROOT"
    if [[ -n "$SERVICE" ]]; then
        docker-compose -f "$compose_file" restart "$SERVICE"
        echo -e "${GREEN}服务 $SERVICE 已重启${NC}"
    else
        docker-compose -f "$compose_file" restart
        echo -e "${GREEN}所有服务已重启${NC}"
    fi
}

# 查看日志
show_logs() {
    local compose_file=$(get_compose_file)
    
    cd "$PROJECT_ROOT"
    if [[ -n "$SERVICE" ]]; then
        echo -e "${CYAN}查看服务 $SERVICE 的日志:${NC}"
        docker-compose -f "$compose_file" logs -f "$SERVICE"
    else
        echo -e "${CYAN}查看所有服务日志:${NC}"
        docker-compose -f "$compose_file" logs -f
    fi
}

# 查看服务状态
show_status() {
    local compose_file=$(get_compose_file)
    
    cd "$PROJECT_ROOT"
    echo -e "${CYAN}$ENVIRONMENT 环境服务状态:${NC}"
    docker-compose -f "$compose_file" ps
    
    echo -e "\n${CYAN}Docker 系统信息:${NC}"
    docker system df
}

# 构建镜像
build_images() {
    local compose_file=$(get_compose_file)
    
    cd "$PROJECT_ROOT"
    if [[ -n "$SERVICE" ]]; then
        echo -e "${YELLOW}构建服务 $SERVICE 的镜像...${NC}"
        docker-compose -f "$compose_file" build "$SERVICE"
    else
        echo -e "${YELLOW}构建所有服务镜像...${NC}"
        docker-compose -f "$compose_file" build
    fi
    echo -e "${GREEN}镜像构建完成!${NC}"
}

# 检查健康状态
check_health() {
    local compose_file=$(get_compose_file)
    
    cd "$PROJECT_ROOT"
    echo -e "${CYAN}检查服务健康状态:${NC}"
    
    # 获取所有服务状态
    services=$(docker-compose -f "$compose_file" ps --services)
    
    for service in $services; do
        container_id=$(docker-compose -f "$compose_file" ps -q "$service" 2>/dev/null)
        if [[ -n "$container_id" ]]; then
            health=$(docker inspect --format='{{.State.Health.Status}}' "$container_id" 2>/dev/null || echo "no-healthcheck")
            status=$(docker inspect --format='{{.State.Status}}' "$container_id")
            
            case $health in
                healthy)
                    echo -e "  $service: ${GREEN}健康 ✓${NC} (状态: $status)"
                    ;;
                unhealthy)
                    echo -e "  $service: ${RED}不健康 ✗${NC} (状态: $status)"
                    ;;
                starting)
                    echo -e "  $service: ${YELLOW}启动中...${NC} (状态: $status)"
                    ;;
                no-healthcheck)
                    echo -e "  $service: ${CYAN}无健康检查${NC} (状态: $status)"
                    ;;
                *)
                    echo -e "  $service: ${PURPLE}未知状态${NC} (状态: $status)"
                    ;;
            esac
        else
            echo -e "  $service: ${RED}未运行 ✗${NC}"
        fi
    done
}

# 扩展服务
scale_service() {
    if [[ -z "$SERVICE" ]] || [[ -z "$REPLICAS" ]]; then
        echo -e "${RED}错误: 扩展服务需要指定 --service 和 --replicas 参数${NC}"
        exit 1
    fi
    
    local compose_file=$(get_compose_file)
    
    cd "$PROJECT_ROOT"
    echo -e "${YELLOW}扩展服务 $SERVICE 到 $REPLICAS 个实例...${NC}"
    docker-compose -f "$compose_file" up -d --scale "$SERVICE=$REPLICAS" "$SERVICE"
    echo -e "${GREEN}服务扩展完成!${NC}"
}

# 清理资源
clean_resources() {
    echo -e "${YELLOW}清理Docker资源...${NC}"
    
    if [[ "$FORCE" == true ]]; then
        echo -e "${RED}强制清理所有资源...${NC}"
        docker system prune -af --volumes
    else
        echo -e "${CYAN}清理未使用的资源...${NC}"
        docker system prune -f
    fi
    
    echo -e "${GREEN}清理完成!${NC}"
}

# 主函数
main() {
    check_dependencies
    parse_arguments "$@"
    
    if [[ -z "$COMMAND" ]]; then
        show_help
        exit 1
    fi
    
    echo -e "${BLUE}OpenPenPal Docker Manager${NC}"
    echo -e "${PURPLE}环境: $ENVIRONMENT${NC}"
    [[ -n "$SERVICE" ]] && echo -e "${PURPLE}服务: $SERVICE${NC}"
    echo ""
    
    case $COMMAND in
        dev)
            start_dev
            ;;
        dev-full)
            start_dev_full
            ;;
        prod)
            start_prod
            ;;
        stop)
            stop_services
            ;;
        restart)
            restart_services
            ;;
        logs)
            show_logs
            ;;
        status)
            show_status
            ;;
        build)
            build_images
            ;;
        health)
            check_health
            ;;
        scale)
            scale_service
            ;;
        clean)
            clean_resources
            ;;
        *)
            echo -e "${RED}错误: 未知命令 $COMMAND${NC}"
            show_help
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"