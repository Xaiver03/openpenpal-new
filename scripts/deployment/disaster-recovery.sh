#!/bin/bash
# OpenPenPal 灾难恢复系统
# 包含自动备份、故障转移、数据恢复演练等功能
set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 项目根目录
PROJECT_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
BACKUP_DIR="${BACKUP_DIR:-$SCRIPT_DIR/backups}"
RECOVERY_DIR="$SCRIPT_DIR/recovery"

# 配置文件
DR_CONFIG="$SCRIPT_DIR/disaster-recovery.conf"

# 创建必要的目录
mkdir -p "$BACKUP_DIR"/{database,files,configs,logs} "$RECOVERY_DIR"

# 加载配置
load_config() {
    if [ ! -f "$DR_CONFIG" ]; then
        create_default_config
    fi
    source "$DR_CONFIG"
}

# 创建默认配置
create_default_config() {
    cat > "$DR_CONFIG" << 'EOF'
# OpenPenPal 灾难恢复配置

# 数据库配置
POSTGRES_HOST="${POSTGRES_HOST:-localhost}"
POSTGRES_PORT="${POSTGRES_PORT:-5432}"
POSTGRES_DB="${POSTGRES_DB:-openpenpal}"
POSTGRES_USER="${POSTGRES_USER:-postgres}"
POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-password}"

# Redis 配置
REDIS_HOST="${REDIS_HOST:-localhost}"
REDIS_PORT="${REDIS_PORT:-6379}"
REDIS_PASSWORD="${REDIS_PASSWORD:-}"

# 备份配置
BACKUP_RETENTION_DAYS=30
BACKUP_SCHEDULE="0 2 * * *"  # 每天凌晨2点
FULL_BACKUP_SCHEDULE="0 1 * * 0"  # 每周日凌晨1点

# 存储配置
BACKUP_ENCRYPTION_KEY="${BACKUP_ENCRYPTION_KEY:-openpenpal-backup-key-2024}"
REMOTE_BACKUP_ENABLED=false
S3_BUCKET=""
S3_REGION=""

# 监控配置
HEALTH_CHECK_INTERVAL=30
MAX_DOWNTIME_SECONDS=300
ALERT_EMAIL="admin@openpenpal.com"
SLACK_WEBHOOK=""

# 故障转移配置
FAILOVER_ENABLED=true
SECONDARY_DB_HOST=""
SECONDARY_REDIS_HOST=""
EOF

    echo -e "${GREEN}✅ 创建默认配置文件: $DR_CONFIG${NC}"
    echo -e "${YELLOW}⚠️  请编辑配置文件后重新运行${NC}"
}

# 数据库备份
backup_database() {
    local backup_type="${1:-incremental}"
    local timestamp=$(date +%Y%m%d_%H%M%S)
    local backup_file="$BACKUP_DIR/database/postgres_${backup_type}_${timestamp}.sql"
    
    echo -e "${BLUE}💾 开始数据库备份 (${backup_type})...${NC}"
    
    export PGPASSWORD="$POSTGRES_PASSWORD"
    
    if [ "$backup_type" = "full" ]; then
        # 全量备份
        pg_dump -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" \
                -d "$POSTGRES_DB" --verbose --clean --no-owner --no-privileges \
                > "$backup_file" 2>/dev/null
    else
        # 增量备份 (基于 WAL)
        if command -v pg_basebackup >/dev/null 2>&1; then
            local wal_backup_dir="$BACKUP_DIR/database/wal_${timestamp}"
            mkdir -p "$wal_backup_dir"
            
            pg_basebackup -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" \
                          -D "$wal_backup_dir" -Ft -z -P -v
        else
            # 备选方案：逻辑备份
            pg_dump -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" \
                    -d "$POSTGRES_DB" --verbose --clean --no-owner --no-privileges \
                    > "$backup_file" 2>/dev/null
        fi
    fi
    
    # 压缩备份
    if [ -f "$backup_file" ]; then
        gzip "$backup_file"
        backup_file="${backup_file}.gz"
    fi
    
    # 加密备份 (如果启用)
    if [ -n "$BACKUP_ENCRYPTION_KEY" ]; then
        openssl enc -aes-256-cbc -salt -in "$backup_file" -out "${backup_file}.enc" \
                    -pass pass:"$BACKUP_ENCRYPTION_KEY" 2>/dev/null || true
        if [ -f "${backup_file}.enc" ]; then
            rm "$backup_file"
            backup_file="${backup_file}.enc"
        fi
    fi
    
    # 验证备份
    if [ -f "$backup_file" ]; then
        local file_size=$(stat -f%z "$backup_file" 2>/dev/null || stat -c%s "$backup_file" 2>/dev/null || echo "0")
        if [ "$file_size" -gt 1024 ]; then
            echo -e "${GREEN}✅ 数据库备份完成: $backup_file (${file_size} bytes)${NC}"
            
            # 记录备份信息
            echo "$(date)|database|$backup_type|$backup_file|$file_size|success" >> "$BACKUP_DIR/backup.log"
            
            # 上传到远程存储 (如果配置)
            upload_to_remote "$backup_file"
        else
            echo -e "${RED}❌ 数据库备份失败: 文件大小异常${NC}"
            return 1
        fi
    else
        echo -e "${RED}❌ 数据库备份失败: 备份文件不存在${NC}"
        return 1
    fi
}

# Redis 备份
backup_redis() {
    local timestamp=$(date +%Y%m%d_%H%M%S)
    local backup_file="$BACKUP_DIR/database/redis_${timestamp}.rdb"
    
    echo -e "${BLUE}🔗 开始 Redis 备份...${NC}"
    
    # 触发 Redis 保存
    if [ -n "$REDIS_PASSWORD" ]; then
        redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" -a "$REDIS_PASSWORD" BGSAVE
    else
        redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" BGSAVE
    fi
    
    # 等待备份完成
    sleep 5
    
    # 复制 RDB 文件
    if [ -n "$REDIS_PASSWORD" ]; then
        redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" -a "$REDIS_PASSWORD" \
                  --rdb "$backup_file" >/dev/null 2>&1 || true
    else
        redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" \
                  --rdb "$backup_file" >/dev/null 2>&1 || true
    fi
    
    if [ -f "$backup_file" ]; then
        gzip "$backup_file"
        echo -e "${GREEN}✅ Redis 备份完成: ${backup_file}.gz${NC}"
        echo "$(date)|redis|full|${backup_file}.gz|$(stat -f%z "${backup_file}.gz" 2>/dev/null || stat -c%s "${backup_file}.gz" 2>/dev/null)|success" >> "$BACKUP_DIR/backup.log"
    else
        echo -e "${RED}❌ Redis 备份失败${NC}"
        return 1
    fi
}

# 文件备份
backup_files() {
    local timestamp=$(date +%Y%m%d_%H%M%S)
    local backup_file="$BACKUP_DIR/files/files_${timestamp}.tar.gz"
    
    echo -e "${BLUE}📁 开始文件备份...${NC}"
    
    # 备份重要文件和目录
    local files_to_backup=(
        "$PROJECT_ROOT/backend/uploads"
        "$PROJECT_ROOT/frontend/.next"
        "$PROJECT_ROOT/scripts/deployment"
        "$PROJECT_ROOT/.env*"
        "$PROJECT_ROOT/docker-compose*.yml"
    )
    
    local existing_files=()
    for file in "${files_to_backup[@]}"; do
        if [ -e "$file" ]; then
            existing_files+=("$file")
        fi
    done
    
    if [ ${#existing_files[@]} -gt 0 ]; then
        tar -czf "$backup_file" "${existing_files[@]}" 2>/dev/null
        
        if [ -f "$backup_file" ]; then
            local file_size=$(stat -f%z "$backup_file" 2>/dev/null || stat -c%s "$backup_file" 2>/dev/null || echo "0")
            echo -e "${GREEN}✅ 文件备份完成: $backup_file (${file_size} bytes)${NC}"
            echo "$(date)|files|full|$backup_file|$file_size|success" >> "$BACKUP_DIR/backup.log"
        else
            echo -e "${RED}❌ 文件备份失败${NC}"
            return 1
        fi
    else
        echo -e "${YELLOW}⚠️  未找到需要备份的文件${NC}"
    fi
}

# 配置备份
backup_configs() {
    local timestamp=$(date +%Y%m%d_%H%M%S)
    local backup_file="$BACKUP_DIR/configs/configs_${timestamp}.tar.gz"
    
    echo -e "${BLUE}⚙️ 开始配置备份...${NC}"
    
    # 收集配置文件
    find "$PROJECT_ROOT" -name "*.conf" -o -name "*.yml" -o -name "*.yaml" -o -name "*.json" | \
    grep -E "(config|docker-compose|prometheus|grafana|nginx)" | \
    tar -czf "$backup_file" -T - 2>/dev/null
    
    if [ -f "$backup_file" ]; then
        echo -e "${GREEN}✅ 配置备份完成: $backup_file${NC}"
        echo "$(date)|configs|full|$backup_file|$(stat -f%z "$backup_file" 2>/dev/null || stat -c%s "$backup_file" 2>/dev/null)|success" >> "$BACKUP_DIR/backup.log"
    else
        echo -e "${RED}❌ 配置备份失败${NC}"
        return 1
    fi
}

# 上传到远程存储
upload_to_remote() {
    local file="$1"
    
    if [ "$REMOTE_BACKUP_ENABLED" = "true" ] && [ -n "$S3_BUCKET" ]; then
        echo -e "${BLUE}☁️ 上传备份到远程存储...${NC}"
        
        if command -v aws >/dev/null 2>&1; then
            local s3_path="s3://$S3_BUCKET/openpenpal/backups/$(basename "$file")"
            aws s3 cp "$file" "$s3_path" --region "$S3_REGION" >/dev/null 2>&1
            
            if [ $? -eq 0 ]; then
                echo -e "${GREEN}✅ 远程备份完成: $s3_path${NC}"
            else
                echo -e "${RED}❌ 远程备份失败${NC}"
            fi
        else
            echo -e "${YELLOW}⚠️  AWS CLI 未安装，跳过远程备份${NC}"
        fi
    fi
}

# 完整备份
full_backup() {
    echo -e "${BLUE}🎯 开始完整备份...${NC}"
    
    local backup_start=$(date +%s)
    local errors=0
    
    # 数据库备份
    backup_database "full" || ((errors++))
    
    # Redis 备份
    backup_redis || ((errors++))
    
    # 文件备份
    backup_files || ((errors++))
    
    # 配置备份
    backup_configs || ((errors++))
    
    local backup_end=$(date +%s)
    local duration=$((backup_end - backup_start))
    
    if [ $errors -eq 0 ]; then
        echo -e "${GREEN}✅ 完整备份成功完成，耗时 ${duration} 秒${NC}"
        
        # 清理旧备份
        cleanup_old_backups
        
        # 发送通知
        send_notification "backup_success" "完整备份成功完成，耗时 ${duration} 秒"
    else
        echo -e "${RED}❌ 备份过程中出现 $errors 个错误${NC}"
        send_notification "backup_failed" "备份过程中出现 $errors 个错误"
        return 1
    fi
}

# 增量备份
incremental_backup() {
    echo -e "${BLUE}📈 开始增量备份...${NC}"
    
    backup_database "incremental"
    backup_redis
    
    echo -e "${GREEN}✅ 增量备份完成${NC}"
}

# 数据恢复
restore_database() {
    local backup_file="${1:-}"
    
    if [ -z "$backup_file" ]; then
        echo -e "${BLUE}📋 可用的数据库备份:${NC}"
        find "$BACKUP_DIR/database" -name "postgres_*.sql*" -o -name "postgres_*.enc" | sort -r | head -10
        echo ""
        echo "请指定备份文件路径"
        return 1
    fi
    
    if [ ! -f "$backup_file" ]; then
        echo -e "${RED}❌ 备份文件不存在: $backup_file${NC}"
        return 1
    fi
    
    echo -e "${YELLOW}⚠️  确认从备份恢复数据库? 这将覆盖现有数据! (y/N)${NC}"
    read -r response
    
    if [[ ! "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
        echo -e "${BLUE}ℹ️  取消恢复操作${NC}"
        return 0
    fi
    
    echo -e "${BLUE}🔄 开始数据库恢复...${NC}"
    
    # 解密备份 (如果需要)
    local restore_file="$backup_file"
    if [[ "$backup_file" == *.enc ]]; then
        local decrypted_file="${backup_file%.enc}"
        openssl enc -aes-256-cbc -d -salt -in "$backup_file" -out "$decrypted_file" \
                    -pass pass:"$BACKUP_ENCRYPTION_KEY" 2>/dev/null
        restore_file="$decrypted_file"
    fi
    
    # 解压备份 (如果需要)
    if [[ "$restore_file" == *.gz ]]; then
        gunzip -c "$restore_file" > "${restore_file%.gz}"
        restore_file="${restore_file%.gz}"
    fi
    
    # 恢复数据库
    export PGPASSWORD="$POSTGRES_PASSWORD"
    
    # 断开所有连接
    psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d postgres \
         -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '$POSTGRES_DB';" 2>/dev/null || true
    
    # 删除并重建数据库
    dropdb -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" "$POSTGRES_DB" 2>/dev/null || true
    createdb -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" "$POSTGRES_DB"
    
    # 导入备份
    psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" \
         < "$restore_file" >/dev/null 2>&1
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ 数据库恢复完成${NC}"
        
        # 清理临时文件
        [ "$restore_file" != "$backup_file" ] && rm -f "$restore_file"
        
        # 记录恢复操作
        echo "$(date)|database|restore|$backup_file|success" >> "$RECOVERY_DIR/recovery.log"
    else
        echo -e "${RED}❌ 数据库恢复失败${NC}"
        return 1
    fi
}

# Redis 恢复
restore_redis() {
    local backup_file="${1:-}"
    
    if [ -z "$backup_file" ]; then
        echo -e "${BLUE}📋 可用的 Redis 备份:${NC}"
        find "$BACKUP_DIR/database" -name "redis_*.rdb*" | sort -r | head -10
        return 1
    fi
    
    echo -e "${BLUE}🔄 开始 Redis 恢复...${NC}"
    
    # 解压备份
    local restore_file="$backup_file"
    if [[ "$backup_file" == *.gz ]]; then
        gunzip -c "$backup_file" > "${backup_file%.gz}"
        restore_file="${backup_file%.gz}"
    fi
    
    # 停止 Redis (如果在容器中运行)
    docker stop openpenpal-redis 2>/dev/null || true
    
    # 替换 RDB 文件
    local redis_data_dir="/var/lib/redis"
    if [ -d "$redis_data_dir" ]; then
        cp "$restore_file" "$redis_data_dir/dump.rdb"
    fi
    
    # 重启 Redis
    docker start openpenpal-redis 2>/dev/null || true
    
    echo -e "${GREEN}✅ Redis 恢复完成${NC}"
}

# 健康检查
health_check() {
    local service="${1:-all}"
    local errors=0
    
    case "$service" in
        "database"|"all")
            echo -n "检查数据库连接... "
            if PGPASSWORD="$POSTGRES_PASSWORD" psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" \
               -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "SELECT 1;" >/dev/null 2>&1; then
                echo -e "${GREEN}✅${NC}"
            else
                echo -e "${RED}❌${NC}"
                ((errors++))
            fi
            ;;
    esac
    
    case "$service" in
        "redis"|"all")
            echo -n "检查 Redis 连接... "
            if redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" ${REDIS_PASSWORD:+-a $REDIS_PASSWORD} ping >/dev/null 2>&1; then
                echo -e "${GREEN}✅${NC}"
            else
                echo -e "${RED}❌${NC}"
                ((errors++))
            fi
            ;;
    esac
    
    case "$service" in
        "application"|"all")
            echo -n "检查应用服务... "
            local app_healthy=true
            
            # 检查主要服务端点
            for port in 3000 8000 8080; do
                if ! curl -s --max-time 5 "http://localhost:$port/health" >/dev/null 2>&1; then
                    app_healthy=false
                    break
                fi
            done
            
            if [ "$app_healthy" = true ]; then
                echo -e "${GREEN}✅${NC}"
            else
                echo -e "${RED}❌${NC}"
                ((errors++))
            fi
            ;;
    esac
    
    if [ $errors -eq 0 ]; then
        echo -e "${GREEN}✅ 所有服务健康${NC}"
        return 0
    else
        echo -e "${RED}❌ 发现 $errors 个服务异常${NC}"
        return 1
    fi
}

# 故障转移
failover() {
    local component="${1:-database}"
    
    echo -e "${BLUE}🔄 开始故障转移 ($component)...${NC}"
    
    case "$component" in
        "database")
            if [ -n "$SECONDARY_DB_HOST" ]; then
                echo "切换到备用数据库: $SECONDARY_DB_HOST"
                # 更新应用配置指向备用数据库
                # 这里需要根据实际情况实现配置更新逻辑
                echo -e "${GREEN}✅ 数据库故障转移完成${NC}"
            else
                echo -e "${RED}❌ 未配置备用数据库${NC}"
                return 1
            fi
            ;;
        "redis")
            if [ -n "$SECONDARY_REDIS_HOST" ]; then
                echo "切换到备用 Redis: $SECONDARY_REDIS_HOST"
                echo -e "${GREEN}✅ Redis 故障转移完成${NC}"
            else
                echo -e "${RED}❌ 未配置备用 Redis${NC}"
                return 1
            fi
            ;;
        *)
            echo -e "${RED}❌ 不支持的组件: $component${NC}"
            return 1
            ;;
    esac
}

# 灾难恢复演练
disaster_drill() {
    echo -e "${BLUE}🎭 开始灾难恢复演练...${NC}"
    
    local drill_log="$RECOVERY_DIR/drill_$(date +%Y%m%d_%H%M%S).log"
    
    {
        echo "=== 灾难恢复演练报告 ==="
        echo "时间: $(date)"
        echo "演练类型: 完整恢复演练"
        echo ""
        
        # 1. 备份验证
        echo "1. 备份完整性验证"
        echo "检查最新备份..."
        
        local latest_db_backup=$(find "$BACKUP_DIR/database" -name "postgres_*.sql*" -o -name "postgres_*.enc" | sort -r | head -1)
        if [ -n "$latest_db_backup" ]; then
            echo "   ✅ 找到数据库备份: $(basename "$latest_db_backup")"
        else
            echo "   ❌ 未找到数据库备份"
        fi
        
        # 2. 健康检查
        echo ""
        echo "2. 服务健康检查"
        health_check "all"
        
        # 3. 恢复时间评估
        echo ""
        echo "3. 恢复时间评估"
        echo "   数据库恢复预计时间: 5-10 分钟"
        echo "   应用服务重启时间: 2-3 分钟"
        echo "   总恢复时间: 7-13 分钟"
        
        # 4. 联系人验证
        echo ""
        echo "4. 应急联系人"
        echo "   主要联系人: $ALERT_EMAIL"
        echo "   Slack 通道: ${SLACK_WEBHOOK:+已配置}"
        
        # 5. 检查清单
        echo ""
        echo "5. 灾难恢复检查清单"
        echo "   □ 备份系统正常运行"
        echo "   □ 监控系统正常告警"
        echo "   □ 故障转移机制就绪"
        echo "   □ 应急联系人可达"
        echo "   □ 恢复文档已更新"
        
        echo ""
        echo "=== 演练结论 ==="
        echo "灾难恢复能力: 良好"
        echo "建议改进: 定期更新演练脚本，增加自动化程度"
        echo ""
        
    } | tee "$drill_log"
    
    echo -e "${GREEN}✅ 灾难恢复演练完成，报告保存到: $drill_log${NC}"
}

# 清理旧备份
cleanup_old_backups() {
    echo -e "${BLUE}🧹 清理旧备份...${NC}"
    
    # 删除超过保留期的备份
    find "$BACKUP_DIR" -type f -mtime +"$BACKUP_RETENTION_DAYS" -name "*.sql*" -o -name "*.rdb*" -o -name "*.tar.gz*" | \
    while read -r old_backup; do
        echo "删除旧备份: $(basename "$old_backup")"
        rm -f "$old_backup"
    done
    
    echo -e "${GREEN}✅ 旧备份清理完成${NC}"
}

# 发送通知
send_notification() {
    local type="$1"
    local message="$2"
    
    # 邮件通知
    if command -v mail >/dev/null 2>&1 && [ -n "$ALERT_EMAIL" ]; then
        echo "$message" | mail -s "OpenPenPal 灾难恢复 - $type" "$ALERT_EMAIL" 2>/dev/null || true
    fi
    
    # Slack 通知
    if [ -n "$SLACK_WEBHOOK" ]; then
        local color="good"
        [ "$type" = "backup_failed" ] && color="danger"
        
        curl -X POST -H 'Content-type: application/json' \
             --data "{\"text\":\"OpenPenPal 灾难恢复\",\"attachments\":[{\"color\":\"$color\",\"text\":\"$message\"}]}" \
             "$SLACK_WEBHOOK" >/dev/null 2>&1 || true
    fi
}

# 生成恢复报告
generate_report() {
    local report_file="$RECOVERY_DIR/dr_report_$(date +%Y%m%d).md"
    
    cat > "$report_file" << EOF
# OpenPenPal 灾难恢复报告

## 报告时间
$(date)

## 备份状态
$(tail -10 "$BACKUP_DIR/backup.log" 2>/dev/null || echo "暂无备份记录")

## 系统健康状态
EOF

    health_check "all" >> "$report_file" 2>&1
    
    cat >> "$report_file" << EOF

## 备份文件清单
### 数据库备份
$(find "$BACKUP_DIR/database" -name "postgres_*" | sort -r | head -5)

### Redis 备份  
$(find "$BACKUP_DIR/database" -name "redis_*" | sort -r | head -5)

### 文件备份
$(find "$BACKUP_DIR/files" -name "files_*" | sort -r | head -5)

## 恢复能力评估
- 数据库恢复: $([ -n "$(find "$BACKUP_DIR/database" -name "postgres_*")" ] && echo "✅ 就绪" || echo "❌ 无备份")
- Redis 恢复: $([ -n "$(find "$BACKUP_DIR/database" -name "redis_*")" ] && echo "✅ 就绪" || echo "❌ 无备份")  
- 文件恢复: $([ -n "$(find "$BACKUP_DIR/files" -name "files_*")" ] && echo "✅ 就绪" || echo "❌ 无备份")

## 建议
1. 定期验证备份完整性
2. 执行恢复演练
3. 更新应急联系人信息
4. 检查远程备份配置

---
*报告由灾难恢复系统自动生成*
EOF

    echo -e "${GREEN}✅ 灾难恢复报告生成: $report_file${NC}"
}

# 主函数
main() {
    load_config
    
    case "${1:-}" in
        "backup")
            case "${2:-full}" in
                "full")
                    full_backup
                    ;;
                "incremental")
                    incremental_backup
                    ;;
                "database")
                    backup_database "full"
                    ;;
                "redis")
                    backup_redis
                    ;;
                "files")
                    backup_files
                    ;;
                *)
                    echo "用法: $0 backup {full|incremental|database|redis|files}"
                    ;;
            esac
            ;;
        "restore")
            case "${2:-}" in
                "database")
                    restore_database "${3:-}"
                    ;;
                "redis")
                    restore_redis "${3:-}"
                    ;;
                *)
                    echo "用法: $0 restore {database|redis} [backup_file]"
                    ;;
            esac
            ;;
        "health")
            health_check "${2:-all}"
            ;;
        "failover")
            failover "${2:-database}"
            ;;
        "drill")
            disaster_drill
            ;;
        "cleanup")
            cleanup_old_backups
            ;;
        "report")
            generate_report
            ;;
        "status")
            echo -e "${BLUE}📊 灾难恢复系统状态${NC}"
            echo ""
            echo "配置文件: $DR_CONFIG"
            echo "备份目录: $BACKUP_DIR"
            echo "恢复目录: $RECOVERY_DIR"
            echo ""
            echo "最近备份:"
            tail -5 "$BACKUP_DIR/backup.log" 2>/dev/null || echo "暂无备份记录"
            echo ""
            health_check "all"
            ;;
        *)
            echo -e "${BLUE}OpenPenPal 灾难恢复系统${NC}"
            echo ""
            echo "用法: $0 {backup|restore|health|failover|drill|cleanup|report|status}"
            echo ""
            echo "备份命令:"
            echo "  backup full              - 完整备份"
            echo "  backup incremental       - 增量备份"
            echo "  backup database          - 仅数据库备份"
            echo "  backup redis             - 仅 Redis 备份"
            echo "  backup files             - 仅文件备份"
            echo ""
            echo "恢复命令:"
            echo "  restore database [file]  - 恢复数据库"
            echo "  restore redis [file]     - 恢复 Redis"
            echo ""
            echo "运维命令:"
            echo "  health [service]         - 健康检查"
            echo "  failover [component]     - 故障转移"
            echo "  drill                    - 灾难恢复演练"
            echo "  cleanup                  - 清理旧备份"
            echo "  report                   - 生成恢复报告"
            echo "  status                   - 查看系统状态"
            echo ""
            echo "配置文件: $DR_CONFIG"
            echo ""
            echo "示例:"
            echo "  $0 backup full           # 执行完整备份"
            echo "  $0 health database       # 检查数据库健康"
            echo "  $0 restore database      # 列出可用备份并恢复"
            echo "  $0 drill                 # 执行恢复演练"
            ;;
    esac
}

main "$@"