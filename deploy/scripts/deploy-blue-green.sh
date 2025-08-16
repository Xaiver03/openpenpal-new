#!/bin/bash
# 蓝绿部署脚本 - 实现零停机部署

set -euo pipefail

# 配置
ENV=${1:-production}
DEPLOY_DIR="/home/ubuntu/openpenpal"
BACKUP_DIR="$DEPLOY_DIR/backups/$(date +%Y%m%d_%H%M%S)"
HEALTH_CHECK_RETRIES=10
HEALTH_CHECK_INTERVAL=6

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

# 日志函数
log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# 检查依赖
check_dependencies() {
    log "检查系统依赖..."
    local deps=("docker" "docker-compose" "jq" "curl")
    
    for dep in "${deps[@]}"; do
        if ! command -v "$dep" &> /dev/null; then
            error "$dep 未安装"
            exit 1
        fi
    done
    
    # 检查 Docker 服务
    if ! docker info &> /dev/null; then
        error "Docker 服务未运行"
        exit 1
    fi
    
    success "所有依赖检查通过"
}

# 加载环境变量
load_environment() {
    log "加载环境变量..."
    
    if [ ! -f "$DEPLOY_DIR/.env.$ENV" ]; then
        error "环境配置文件不存在: $DEPLOY_DIR/.env.$ENV"
        exit 1
    fi
    
    source "$DEPLOY_DIR/.env.$ENV"
    
    # 验证必需的环境变量
    required_vars=("DOCKER_REGISTRY" "DOCKER_NAMESPACE" "IMAGE_TAG" "JWT_SECRET" "POSTGRES_PASSWORD")
    for var in "${required_vars[@]}"; do
        if [ -z "${!var:-}" ]; then
            error "缺少必需的环境变量: $var"
            exit 1
        fi
    done
    
    success "环境变量加载完成"
}

# 创建备份
create_backup() {
    log "创建部署前备份..."
    
    mkdir -p "$BACKUP_DIR"
    
    # 备份当前运行的容器信息
    docker-compose -f "$DEPLOY_DIR/deploy/docker-compose.$ENV.yml" \
        -p "${PROJECT_NAME}_blue" \
        ps -q | xargs -I {} docker inspect {} > "$BACKUP_DIR/containers.json"
    
    # 备份数据库
    if docker exec "${PROJECT_NAME}-postgres" pg_isready &> /dev/null; then
        log "备份数据库..."
        docker exec "${PROJECT_NAME}-postgres" \
            pg_dump -U openpenpal openpenpal | gzip > "$BACKUP_DIR/database.sql.gz"
    fi
    
    # 备份配置文件
    cp "$DEPLOY_DIR/.env.$ENV" "$BACKUP_DIR/"
    
    # 记录当前镜像版本
    docker images --format "table {{.Repository}}:{{.Tag}}\t{{.ID}}" \
        | grep "$DOCKER_NAMESPACE" > "$BACKUP_DIR/images.txt"
    
    success "备份完成: $BACKUP_DIR"
}

# 拉取新镜像
pull_images() {
    log "拉取最新镜像..."
    
    # 登录镜像仓库
    echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin "$DOCKER_REGISTRY"
    
    # 使用 docker-compose 拉取所有镜像
    docker-compose -f "$DEPLOY_DIR/deploy/docker-compose.$ENV.yml" pull
    
    success "镜像拉取完成"
}

# 启动绿环境
start_green_environment() {
    log "启动绿环境..."
    
    # 使用不同的项目名启动绿环境
    docker-compose -f "$DEPLOY_DIR/deploy/docker-compose.$ENV.yml" \
        -p "${PROJECT_NAME}_green" \
        up -d
    
    # 等待容器启动
    sleep 10
    
    success "绿环境已启动"
}

# 健康检查
health_check() {
    local env_name=$1
    log "执行健康检查: $env_name 环境..."
    
    local services=(
        "frontend:3000:/"
        "api-gateway:8000:/health"
        "backend:8080:/health"
        "write-service:8001:/health"
        "courier-service:8002:/health"
        "admin-service:8003:/actuator/health"
        "ocr-service:8004:/health"
    )
    
    local all_healthy=true
    
    for retry in $(seq 1 $HEALTH_CHECK_RETRIES); do
        all_healthy=true
        log "健康检查尝试 $retry/$HEALTH_CHECK_RETRIES"
        
        for service_info in "${services[@]}"; do
            IFS=':' read -r service port path <<< "$service_info"
            local container_name="${PROJECT_NAME}_${env_name}_${service}_1"
            
            # 获取容器 IP
            local container_ip=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' "$container_name" 2>/dev/null)
            
            if [ -z "$container_ip" ]; then
                warning "$service 容器未找到或未运行"
                all_healthy=false
                continue
            fi
            
            # 检查健康状态
            if curl -sf "http://${container_ip}:${port}${path}" > /dev/null; then
                success "$service 健康检查通过"
            else
                warning "$service 健康检查失败"
                all_healthy=false
            fi
        done
        
        if [ "$all_healthy" = true ]; then
            success "所有服务健康检查通过"
            return 0
        fi
        
        if [ $retry -lt $HEALTH_CHECK_RETRIES ]; then
            log "等待 ${HEALTH_CHECK_INTERVAL} 秒后重试..."
            sleep $HEALTH_CHECK_INTERVAL
        fi
    done
    
    error "健康检查失败"
    return 1
}

# 切换流量
switch_traffic() {
    log "切换流量到绿环境..."
    
    # 更新 Nginx 配置指向绿环境
    cat > /tmp/nginx-green.conf << EOF
upstream frontend {
    server ${PROJECT_NAME}_green_frontend_1:3000;
}

upstream api {
    server ${PROJECT_NAME}_green_api-gateway_1:8000;
}

upstream admin {
    server ${PROJECT_NAME}_green_admin-frontend_1:3001;
}
EOF
    
    # 应用新配置
    docker cp /tmp/nginx-green.conf "${PROJECT_NAME}-nginx:/etc/nginx/conf.d/upstream.conf"
    docker exec "${PROJECT_NAME}-nginx" nginx -s reload
    
    success "流量已切换到绿环境"
}

# 停止蓝环境
stop_blue_environment() {
    log "停止蓝环境..."
    
    # 给一些时间让现有请求完成
    sleep 5
    
    # 停止蓝环境
    docker-compose -f "$DEPLOY_DIR/deploy/docker-compose.$ENV.yml" \
        -p "${PROJECT_NAME}_blue" \
        down
    
    success "蓝环境已停止"
}

# 更新环境标签
update_environment_labels() {
    log "更新环境标签..."
    
    # 将绿环境重命名为蓝环境
    local green_containers=$(docker ps -q -f "label=com.docker.compose.project=${PROJECT_NAME}_green")
    
    for container_id in $green_containers; do
        local container_name=$(docker inspect -f '{{.Name}}' "$container_id" | sed 's/\///')
        local new_name=$(echo "$container_name" | sed "s/_green_/_blue_/")
        docker rename "$container_name" "$new_name" 2>/dev/null || true
    done
    
    success "环境标签已更新"
}

# 清理资源
cleanup() {
    log "清理资源..."
    
    # 清理未使用的镜像
    docker image prune -a --force --filter "until=72h"
    
    # 清理未使用的卷
    docker volume prune -f
    
    # 保留最近10个备份
    if [ -d "$DEPLOY_DIR/backups" ]; then
        ls -t "$DEPLOY_DIR/backups" | tail -n +11 | xargs -I {} rm -rf "$DEPLOY_DIR/backups/{}"
    fi
    
    success "资源清理完成"
}

# 发送通知
send_notification() {
    local status=$1
    local message=$2
    
    log "发送部署通知..."
    
    # 发送到监控系统
    if [ -n "${MONITORING_WEBHOOK:-}" ]; then
        curl -X POST "$MONITORING_WEBHOOK" \
            -H "Content-Type: application/json" \
            -d "{
                \"deployment\": \"$ENV\",
                \"status\": \"$status\",
                \"message\": \"$message\",
                \"timestamp\": \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\",
                \"version\": \"$IMAGE_TAG\"
            }"
    fi
}

# 回滚函数
rollback() {
    error "部署失败，执行回滚..."
    
    # 停止绿环境
    docker-compose -f "$DEPLOY_DIR/deploy/docker-compose.$ENV.yml" \
        -p "${PROJECT_NAME}_green" \
        down
    
    # 如果蓝环境已停止，重新启动
    if ! docker ps | grep -q "${PROJECT_NAME}_blue"; then
        warning "重新启动蓝环境..."
        docker-compose -f "$DEPLOY_DIR/deploy/docker-compose.$ENV.yml" \
            -p "${PROJECT_NAME}_blue" \
            up -d
    fi
    
    send_notification "failed" "部署失败并已回滚"
    exit 1
}

# 主函数
main() {
    log "开始蓝绿部署: 环境=$ENV"
    
    # 设置错误处理
    trap rollback ERR
    
    # 执行部署步骤
    check_dependencies
    load_environment
    create_backup
    pull_images
    start_green_environment
    
    # 健康检查
    if ! health_check "green"; then
        error "绿环境健康检查失败"
        rollback
    fi
    
    # 切换流量
    switch_traffic
    
    # 验证切换后的服务
    sleep 5
    if ! curl -sf "http://localhost/health" > /dev/null; then
        error "流量切换后服务不可用"
        rollback
    fi
    
    # 停止旧环境
    stop_blue_environment
    update_environment_labels
    
    # 清理
    cleanup
    
    # 发送成功通知
    send_notification "success" "部署成功完成"
    
    success "蓝绿部署完成！版本: $IMAGE_TAG"
}

# 执行主函数
main "$@"