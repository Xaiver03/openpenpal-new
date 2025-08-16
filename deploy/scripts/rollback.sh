#!/bin/bash
# 回滚脚本 - 快速恢复到之前的版本

set -euo pipefail

# 配置
ROLLBACK_TARGET=${1:-latest}  # latest, 特定版本号, 或备份ID
ENV=${2:-production}
DEPLOY_DIR="/home/ubuntu/openpenpal"
BACKUP_DIR="$DEPLOY_DIR/backups"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
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

# 确认回滚
confirm_rollback() {
    echo -e "${YELLOW}警告: 即将执行回滚操作！${NC}"
    echo "目标: $ROLLBACK_TARGET"
    echo "环境: $ENV"
    echo
    read -p "确认要继续吗？(yes/no): " -r
    echo
    
    if [[ ! $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
        log "回滚操作已取消"
        exit 0
    fi
}

# 查找备份
find_backup() {
    local target=$1
    
    if [ "$target" = "latest" ]; then
        # 查找最新的备份
        local latest_backup=$(ls -t "$BACKUP_DIR"/*.tar.gz 2>/dev/null | head -1)
        if [ -z "$latest_backup" ]; then
            error "未找到任何备份文件"
            exit 1
        fi
        echo "$latest_backup"
    elif [ -f "$BACKUP_DIR/${target}.tar.gz" ]; then
        # 直接指定的备份文件
        echo "$BACKUP_DIR/${target}.tar.gz"
    elif [ -f "$target" ]; then
        # 完整路径
        echo "$target"
    else
        error "未找到备份: $target"
        exit 1
    fi
}

# 验证备份文件
verify_backup_file() {
    local backup_file=$1
    
    log "验证备份文件..."
    
    # 检查校验和
    if [ -f "${backup_file}.sha256" ]; then
        if sha256sum -c "${backup_file}.sha256" &> /dev/null; then
            success "备份文件校验通过"
        else
            error "备份文件校验失败"
            exit 1
        fi
    else
        warning "未找到校验和文件，跳过验证"
    fi
    
    # 测试能否解压
    if tar -tzf "$backup_file" > /dev/null 2>&1; then
        success "备份文件完整性检查通过"
    else
        error "备份文件损坏"
        exit 1
    fi
}

# 停止当前服务
stop_current_services() {
    log "停止当前运行的服务..."
    
    # 保存当前容器状态
    docker ps -a --format "json" > "$DEPLOY_DIR/rollback_container_state.json"
    
    # 停止所有服务
    docker-compose -f "$DEPLOY_DIR/deploy/docker-compose.$ENV.yml" down || true
    
    success "当前服务已停止"
}

# 提取备份
extract_backup() {
    local backup_file=$1
    local temp_dir="/tmp/openpenpal_rollback_$(date +%s)"
    
    log "解压备份文件..."
    
    mkdir -p "$temp_dir"
    tar -xzf "$backup_file" -C "$temp_dir"
    
    # 查找备份内容
    local backup_content=$(find "$temp_dir" -maxdepth 1 -type d | grep -v "^$temp_dir$" | head -1)
    
    if [ -z "$backup_content" ]; then
        error "备份文件格式错误"
        exit 1
    fi
    
    echo "$backup_content"
}

# 恢复数据库
restore_database() {
    local backup_dir=$1
    
    log "恢复数据库..."
    
    # 启动数据库容器
    docker-compose -f "$DEPLOY_DIR/deploy/docker-compose.$ENV.yml" up -d postgres redis
    
    # 等待数据库就绪
    sleep 10
    
    # 恢复 PostgreSQL
    if [ -f "$backup_dir/database/openpenpal_full.sql.gz" ]; then
        log "恢复 PostgreSQL 数据..."
        
        # 先删除现有数据库
        docker exec openpenpal-postgres psql -U postgres -c "DROP DATABASE IF EXISTS openpenpal;"
        docker exec openpenpal-postgres psql -U postgres -c "CREATE DATABASE openpenpal;"
        
        # 恢复数据
        gunzip -c "$backup_dir/database/openpenpal_full.sql.gz" | \
            docker exec -i openpenpal-postgres psql -U openpenpal openpenpal
        
        success "PostgreSQL 数据恢复完成"
    fi
    
    # 恢复 Redis
    if [ -f "$backup_dir/database/redis_dump.rdb" ]; then
        log "恢复 Redis 数据..."
        
        # 停止 Redis 保存
        docker exec openpenpal-redis redis-cli CONFIG SET save ""
        
        # 复制备份文件
        docker cp "$backup_dir/database/redis_dump.rdb" openpenpal-redis:/data/dump.rdb
        
        # 重启 Redis 加载数据
        docker restart openpenpal-redis
        
        success "Redis 数据恢复完成"
    fi
}

# 恢复文件
restore_files() {
    local backup_dir=$1
    
    log "恢复文件..."
    
    # 恢复上传文件
    local file_types=("uploads" "letters" "ocr" "qrcodes")
    
    for file_type in "${file_types[@]}"; do
        if [ -f "$backup_dir/files/${file_type}.tar.gz" ]; then
            log "恢复 $file_type 文件..."
            
            # 备份当前文件
            if [ -d "$DEPLOY_DIR/data/$file_type" ]; then
                mv "$DEPLOY_DIR/data/$file_type" "$DEPLOY_DIR/data/${file_type}.rollback"
            fi
            
            # 解压恢复
            tar -xzf "$backup_dir/files/${file_type}.tar.gz" -C "$DEPLOY_DIR/data/"
            
            success "$file_type 文件恢复完成"
        fi
    done
}

# 恢复配置
restore_configuration() {
    local backup_dir=$1
    
    log "恢复配置文件..."
    
    # 备份当前配置
    mkdir -p "$DEPLOY_DIR/config_rollback"
    cp "$DEPLOY_DIR"/.env.* "$DEPLOY_DIR/config_rollback/" 2>/dev/null || true
    
    # 恢复配置
    if [ -d "$backup_dir/config" ]; then
        cp "$backup_dir/config"/.env.* "$DEPLOY_DIR/" 2>/dev/null || true
        
        # 恢复 Docker Compose 配置
        if ls "$backup_dir/config"/*.yml &> /dev/null; then
            cp "$backup_dir/config"/*.yml "$DEPLOY_DIR/deploy/"
        fi
        
        success "配置文件恢复完成"
    fi
}

# 从备份元数据获取版本
get_version_from_backup() {
    local backup_dir=$1
    
    if [ -f "$backup_dir/backup_metadata.json" ]; then
        local version=$(jq -r '.version' "$backup_dir/backup_metadata.json" 2>/dev/null)
        if [ -n "$version" ] && [ "$version" != "null" ]; then
            echo "$version"
        else
            echo "unknown"
        fi
    else
        echo "unknown"
    fi
}

# 启动服务
start_services() {
    local version=$1
    
    log "启动服务..."
    
    # 如果有特定版本，更新镜像标签
    if [ "$version" != "unknown" ]; then
        export IMAGE_TAG="$version"
    fi
    
    # 启动所有服务
    docker-compose -f "$DEPLOY_DIR/deploy/docker-compose.$ENV.yml" up -d
    
    # 等待服务启动
    sleep 30
    
    success "服务启动完成"
}

# 验证回滚
verify_rollback() {
    log "验证回滚结果..."
    
    # 执行健康检查
    if "$DEPLOY_DIR/scripts/health-check.sh"; then
        success "回滚验证通过"
        return 0
    else
        error "回滚验证失败"
        return 1
    fi
}

# 清理临时文件
cleanup() {
    log "清理临时文件..."
    
    # 清理解压的临时文件
    rm -rf /tmp/openpenpal_rollback_*
    
    # 如果回滚成功，删除回滚备份
    if [ -d "$DEPLOY_DIR/data/uploads.rollback" ]; then
        read -p "是否删除回滚前的备份文件？(yes/no): " -r
        if [[ $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
            rm -rf "$DEPLOY_DIR/data"/*.rollback
            rm -rf "$DEPLOY_DIR/config_rollback"
        fi
    fi
}

# 记录回滚操作
log_rollback() {
    local backup_file=$1
    local status=$2
    
    cat >> "$DEPLOY_DIR/rollback.log" << EOF
=====================================
时间: $(date '+%Y-%m-%d %H:%M:%S')
备份文件: $backup_file
状态: $status
操作者: ${SUDO_USER:-$USER}
=====================================

EOF
}

# 主函数
main() {
    log "========================================="
    log "OpenPenPal 回滚操作"
    log "目标: $ROLLBACK_TARGET"
    log "环境: $ENV"
    log "========================================="
    
    # 确认回滚
    if [ "${FORCE_ROLLBACK:-false}" != "true" ]; then
        confirm_rollback
    fi
    
    # 查找并验证备份
    local backup_file=$(find_backup "$ROLLBACK_TARGET")
    log "使用备份文件: $backup_file"
    
    verify_backup_file "$backup_file"
    
    # 提取备份
    local backup_content=$(extract_backup "$backup_file")
    log "备份内容位置: $backup_content"
    
    # 获取版本信息
    local version=$(get_version_from_backup "$backup_content")
    log "回滚到版本: $version"
    
    # 执行回滚步骤
    stop_current_services
    
    # 根据备份内容执行不同的恢复策略
    if [ -d "$backup_content/database" ]; then
        restore_database "$backup_content"
    fi
    
    if [ -d "$backup_content/files" ]; then
        restore_files "$backup_content"
    fi
    
    if [ -d "$backup_content/config" ]; then
        restore_configuration "$backup_content"
    fi
    
    # 启动服务
    start_services "$version"
    
    # 验证回滚
    if verify_rollback; then
        success "回滚成功完成！"
        log_rollback "$backup_file" "成功"
        
        # 发送成功通知
        if [ -n "${ROLLBACK_NOTIFICATION_WEBHOOK:-}" ]; then
            curl -X POST "$ROLLBACK_NOTIFICATION_WEBHOOK" \
                -H "Content-Type: application/json" \
                -d "{
                    \"event\": \"rollback_success\",
                    \"backup\": \"$(basename "$backup_file")\",
                    \"version\": \"$version\",
                    \"timestamp\": \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\"
                }"
        fi
    else
        error "回滚失败！"
        log_rollback "$backup_file" "失败"
        
        # 尝试恢复到回滚前状态
        warning "尝试恢复到回滚前的状态..."
        # 这里可以添加恢复逻辑
        
        exit 1
    fi
    
    # 清理
    cleanup
}

# 执行主函数
main "$@"