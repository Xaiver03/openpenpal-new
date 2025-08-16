#!/bin/bash
# 备份脚本 - 支持数据库、文件和配置备份

set -euo pipefail

# 配置
BACKUP_TYPE=${1:-scheduled}  # scheduled, pre-deployment, manual
BACKUP_BASE_DIR="/home/ubuntu/openpenpal/backups"
BACKUP_DIR="$BACKUP_BASE_DIR/$(date +%Y%m%d_%H%M%S)_$BACKUP_TYPE"
RETENTION_DAYS=7
TENCENT_COS_BUCKET=${TENCENT_COS_BUCKET:-""}
TENCENT_COS_REGION=${TENCENT_COS_REGION:-"ap-guangzhou"}

# 颜色输出
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# 日志函数
log() {
    echo -e "[$(date +'%Y-%m-%d %H:%M:%S')] $1"
}

success() {
    echo -e "${GREEN}✓${NC} $1"
}

warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

error() {
    echo -e "${RED}✗${NC} $1" >&2
}

# 创建备份目录
create_backup_directory() {
    log "创建备份目录: $BACKUP_DIR"
    mkdir -p "$BACKUP_DIR"/{database,files,config,logs}
}

# 备份数据库
backup_database() {
    log "开始备份数据库..."
    
    # PostgreSQL 主数据库
    if docker exec openpenpal-postgres pg_isready -U openpenpal &> /dev/null; then
        log "备份 PostgreSQL 数据库..."
        
        # 完整备份
        docker exec openpenpal-postgres pg_dump -U openpenpal \
            --clean --if-exists --create \
            --exclude-table-data='*_log' \
            --exclude-table-data='*_temp' \
            openpenpal | gzip -9 > "$BACKUP_DIR/database/openpenpal_full.sql.gz"
        
        # 仅结构备份
        docker exec openpenpal-postgres pg_dump -U openpenpal \
            --schema-only \
            openpenpal > "$BACKUP_DIR/database/openpenpal_schema.sql"
        
        # 备份角色和权限
        docker exec openpenpal-postgres pg_dumpall -U openpenpal \
            --roles-only > "$BACKUP_DIR/database/roles.sql"
        
        success "PostgreSQL 备份完成"
    else
        error "PostgreSQL 服务不可用"
    fi
    
    # Redis 数据备份
    if docker exec openpenpal-redis redis-cli ping &> /dev/null; then
        log "备份 Redis 数据..."
        
        # 触发 Redis 持久化
        docker exec openpenpal-redis redis-cli BGSAVE
        sleep 5
        
        # 复制 RDB 文件
        docker cp openpenpal-redis:/data/dump.rdb "$BACKUP_DIR/database/redis_dump.rdb"
        
        # 导出 Redis 键值对（用于调试）
        docker exec openpenpal-redis redis-cli --scan | while read key; do
            echo "SET $key \"$(docker exec openpenpal-redis redis-cli GET "$key")\""
        done > "$BACKUP_DIR/database/redis_keys.txt"
        
        success "Redis 备份完成"
    else
        warning "Redis 服务不可用，跳过备份"
    fi
}

# 备份文件
backup_files() {
    log "开始备份文件..."
    
    # 用户上传文件
    local upload_dirs=(
        "/home/ubuntu/openpenpal/data/uploads"
        "/home/ubuntu/openpenpal/data/letters"
        "/home/ubuntu/openpenpal/data/ocr"
        "/home/ubuntu/openpenpal/data/qrcodes"
    )
    
    for dir in "${upload_dirs[@]}"; do
        if [ -d "$dir" ]; then
            local dir_name=$(basename "$dir")
            log "备份 $dir_name 目录..."
            tar -czf "$BACKUP_DIR/files/${dir_name}.tar.gz" -C "$(dirname "$dir")" "$dir_name" 2>/dev/null || true
        fi
    done
    
    # 备份日志文件（最近7天）
    log "备份日志文件..."
    find /home/ubuntu/openpenpal/logs -name "*.log" -mtime -7 -type f | \
        tar -czf "$BACKUP_DIR/logs/recent_logs.tar.gz" -T - 2>/dev/null || true
    
    success "文件备份完成"
}

# 备份配置
backup_configuration() {
    log "开始备份配置..."
    
    # 环境配置文件
    cp /home/ubuntu/openpenpal/.env.* "$BACKUP_DIR/config/" 2>/dev/null || true
    
    # Docker 配置
    cp -r /home/ubuntu/openpenpal/deploy/*.yml "$BACKUP_DIR/config/" 2>/dev/null || true
    
    # Nginx 配置
    if [ -f /etc/nginx/sites-available/openpenpal ]; then
        cp /etc/nginx/sites-available/openpenpal "$BACKUP_DIR/config/nginx.conf"
    fi
    
    # SSL 证书信息（不备份私钥）
    if [ -d /etc/letsencrypt/live/openpenpal.com ]; then
        ls -la /etc/letsencrypt/live/openpenpal.com/ > "$BACKUP_DIR/config/ssl_info.txt"
    fi
    
    # 容器配置
    docker ps -a --format "table {{.Names}}\t{{.Image}}\t{{.Status}}" > "$BACKUP_DIR/config/containers.txt"
    docker images --format "table {{.Repository}}:{{.Tag}}\t{{.ID}}\t{{.Size}}" > "$BACKUP_DIR/config/images.txt"
    
    # 系统信息
    {
        echo "=== 系统信息 ==="
        uname -a
        echo
        echo "=== 磁盘使用 ==="
        df -h
        echo
        echo "=== 内存使用 ==="
        free -h
        echo
        echo "=== Docker 信息 ==="
        docker info
    } > "$BACKUP_DIR/config/system_info.txt"
    
    success "配置备份完成"
}

# 创建备份元数据
create_backup_metadata() {
    log "创建备份元数据..."
    
    cat > "$BACKUP_DIR/backup_metadata.json" << EOF
{
    "backup_id": "$(basename "$BACKUP_DIR")",
    "backup_type": "$BACKUP_TYPE",
    "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
    "version": "$(cd /home/ubuntu/openpenpal && git rev-parse HEAD 2>/dev/null || echo 'unknown')",
    "size": "$(du -sh "$BACKUP_DIR" | cut -f1)",
    "components": {
        "database": $([ -d "$BACKUP_DIR/database" ] && echo "true" || echo "false"),
        "files": $([ -d "$BACKUP_DIR/files" ] && echo "true" || echo "false"),
        "config": $([ -d "$BACKUP_DIR/config" ] && echo "true" || echo "false"),
        "logs": $([ -d "$BACKUP_DIR/logs" ] && echo "true" || echo "false")
    }
}
EOF
}

# 压缩备份
compress_backup() {
    log "压缩备份文件..."
    
    cd "$BACKUP_BASE_DIR"
    local backup_name=$(basename "$BACKUP_DIR")
    tar -czf "${backup_name}.tar.gz" "$backup_name"
    
    # 计算校验和
    sha256sum "${backup_name}.tar.gz" > "${backup_name}.tar.gz.sha256"
    
    # 删除未压缩的目录
    rm -rf "$backup_name"
    
    success "备份已压缩: ${backup_name}.tar.gz"
}

# 上传到腾讯云 COS
upload_to_cos() {
    if [ -z "$TENCENT_COS_BUCKET" ]; then
        warning "未配置腾讯云 COS，跳过远程备份"
        return
    fi
    
    log "上传备份到腾讯云 COS..."
    
    local backup_file="${BACKUP_DIR}.tar.gz"
    local remote_path="openpenpal-backups/$(date +%Y/%m)/$(basename "$backup_file")"
    
    # 使用 coscli 上传
    if command -v coscli &> /dev/null; then
        coscli cp "$backup_file" "cos://${TENCENT_COS_BUCKET}/${remote_path}" \
            --region "$TENCENT_COS_REGION" \
            --storage-class STANDARD_IA
        
        success "备份已上传到 COS: $remote_path"
    else
        warning "coscli 未安装，无法上传到 COS"
    fi
}

# 清理旧备份
cleanup_old_backups() {
    log "清理超过 ${RETENTION_DAYS} 天的旧备份..."
    
    # 本地清理
    find "$BACKUP_BASE_DIR" -name "*.tar.gz" -type f -mtime +$RETENTION_DAYS -delete
    find "$BACKUP_BASE_DIR" -name "*.sha256" -type f -mtime +$RETENTION_DAYS -delete
    
    # COS 清理（如果配置了）
    if [ -n "$TENCENT_COS_BUCKET" ] && command -v coscli &> /dev/null; then
        log "清理 COS 上的旧备份..."
        # 这里需要更复杂的逻辑来清理 COS 上的旧文件
        # coscli ls "cos://${TENCENT_COS_BUCKET}/openpenpal-backups/" | ...
    fi
    
    success "旧备份清理完成"
}

# 验证备份
verify_backup() {
    log "验证备份完整性..."
    
    local backup_file="${BACKUP_DIR}.tar.gz"
    
    # 验证校验和
    if sha256sum -c "${backup_file}.sha256" &> /dev/null; then
        success "备份文件校验通过"
    else
        error "备份文件校验失败！"
        return 1
    fi
    
    # 测试解压
    if tar -tzf "$backup_file" > /dev/null 2>&1; then
        success "备份文件可正常解压"
    else
        error "备份文件损坏！"
        return 1
    fi
}

# 发送通知
send_notification() {
    local status=$1
    local message=$2
    
    # 记录到系统日志
    logger -t "openpenpal-backup" "$status: $message"
    
    # 如果配置了 webhook，发送通知
    if [ -n "${BACKUP_NOTIFICATION_WEBHOOK:-}" ]; then
        curl -X POST "$BACKUP_NOTIFICATION_WEBHOOK" \
            -H "Content-Type: application/json" \
            -d "{
                \"backup_type\": \"$BACKUP_TYPE\",
                \"status\": \"$status\",
                \"message\": \"$message\",
                \"timestamp\": \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\",
                \"backup_size\": \"$(du -sh "${BACKUP_DIR}.tar.gz" 2>/dev/null | cut -f1 || echo "unknown")\"
            }" 2>/dev/null || true
    fi
}

# 主函数
main() {
    log "========================================="
    log "OpenPenPal 备份任务"
    log "类型: $BACKUP_TYPE"
    log "时间: $(date '+%Y-%m-%d %H:%M:%S')"
    log "========================================="
    
    # 检查依赖
    for cmd in docker tar gzip sha256sum; do
        if ! command -v "$cmd" &> /dev/null; then
            error "$cmd 命令未找到"
            exit 1
        fi
    done
    
    # 执行备份步骤
    create_backup_directory
    
    # 根据备份类型执行不同的备份策略
    case "$BACKUP_TYPE" in
        "pre-deployment")
            backup_database
            backup_configuration
            ;;
        "scheduled"|"manual")
            backup_database
            backup_files
            backup_configuration
            ;;
        *)
            error "未知的备份类型: $BACKUP_TYPE"
            exit 1
            ;;
    esac
    
    create_backup_metadata
    compress_backup
    
    # 验证备份
    if verify_backup; then
        upload_to_cos
        cleanup_old_backups
        send_notification "success" "备份完成: $(basename "${BACKUP_DIR}.tar.gz")"
        success "备份任务完成！"
    else
        send_notification "failed" "备份验证失败"
        error "备份任务失败！"
        exit 1
    fi
}

# 执行主函数
main "$@"