#!/bin/bash

# 数据库迁移脚本
# 支持迁移、回滚、种子数据等操作

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 项目根目录
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
BACKEND_DIR="$PROJECT_ROOT/backend"
MIGRATIONS_DIR="$BACKEND_DIR/migrations"
SEEDS_DIR="$BACKEND_DIR/seeds"

# 默认配置
DB_TYPE="${DB_TYPE:-postgres}"
DATABASE_URL="${DATABASE_URL:-}"
MIGRATION_TABLE="${MIGRATION_TABLE:-schema_migrations}"

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

# 检查环境
check_environment() {
    log_info "检查环境配置..."
    
    # 检查数据库URL
    if [ -z "$DATABASE_URL" ]; then
        # 尝试从.env文件加载
        if [ -f "$PROJECT_ROOT/.env" ]; then
            source "$PROJECT_ROOT/.env"
        fi
        
        if [ -z "$DATABASE_URL" ]; then
            log_error "DATABASE_URL 未设置"
            exit 1
        fi
    fi
    
    # 检查psql命令
    if ! command -v psql &> /dev/null; then
        log_error "psql 命令未找到，请安装 PostgreSQL 客户端"
        exit 1
    fi
    
    # 测试数据库连接
    if ! psql "$DATABASE_URL" -c "SELECT 1" &> /dev/null; then
        log_error "无法连接到数据库"
        exit 1
    fi
    
    log_success "环境检查通过"
}

# 创建迁移表
create_migration_table() {
    log_info "创建迁移记录表..."
    
    psql "$DATABASE_URL" <<EOF
CREATE TABLE IF NOT EXISTS $MIGRATION_TABLE (
    id SERIAL PRIMARY KEY,
    version VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    execution_time_ms INTEGER,
    checksum VARCHAR(64)
);

CREATE INDEX IF NOT EXISTS idx_migration_version ON $MIGRATION_TABLE(version);
EOF
    
    log_success "迁移记录表已创建"
}

# 获取已应用的迁移
get_applied_migrations() {
    psql "$DATABASE_URL" -t -A -c "SELECT version FROM $MIGRATION_TABLE ORDER BY version" 2>/dev/null || echo ""
}

# 计算文件校验和
calculate_checksum() {
    local file=$1
    if command -v sha256sum &> /dev/null; then
        sha256sum "$file" | cut -d' ' -f1
    else
        shasum -a 256 "$file" | cut -d' ' -f1
    fi
}

# 运行单个迁移
run_migration() {
    local migration_file=$1
    local version=$(basename "$migration_file" .sql | cut -d'_' -f1)
    local description=$(basename "$migration_file" .sql | cut -d'_' -f2-)
    local checksum=$(calculate_checksum "$migration_file")
    
    log_info "运行迁移: $version - $description"
    
    # 开始计时
    local start_time=$(date +%s%3N)
    
    # 在事务中执行迁移
    psql "$DATABASE_URL" <<EOF
BEGIN;

-- 执行迁移脚本
\i $migration_file

-- 记录迁移
INSERT INTO $MIGRATION_TABLE (version, description, checksum, execution_time_ms)
VALUES ('$version', '$description', '$checksum', 0);

COMMIT;
EOF
    
    # 计算执行时间
    local end_time=$(date +%s%3N)
    local execution_time=$((end_time - start_time))
    
    # 更新执行时间
    psql "$DATABASE_URL" -c "UPDATE $MIGRATION_TABLE SET execution_time_ms = $execution_time WHERE version = '$version'"
    
    log_success "迁移完成: $version (耗时: ${execution_time}ms)"
}

# 执行迁移
migrate_up() {
    check_environment
    create_migration_table
    
    # 创建迁移目录
    mkdir -p "$MIGRATIONS_DIR"
    
    # 获取已应用的迁移
    local applied_migrations=$(get_applied_migrations)
    
    # 查找待执行的迁移
    local pending_count=0
    for migration_file in "$MIGRATIONS_DIR"/*.sql; do
        if [ -f "$migration_file" ]; then
            local version=$(basename "$migration_file" .sql | cut -d'_' -f1)
            
            # 检查是否已应用
            if ! echo "$applied_migrations" | grep -q "^$version$"; then
                run_migration "$migration_file"
                ((pending_count++))
            fi
        fi
    done
    
    if [ $pending_count -eq 0 ]; then
        log_info "没有待执行的迁移"
    else
        log_success "成功执行 $pending_count 个迁移"
    fi
}

# 回滚迁移
migrate_down() {
    local steps=${1:-1}
    
    check_environment
    
    log_info "回滚最近 $steps 个迁移..."
    
    # 获取最近的迁移
    local recent_migrations=$(psql "$DATABASE_URL" -t -A -c "SELECT version FROM $MIGRATION_TABLE ORDER BY version DESC LIMIT $steps")
    
    for version in $recent_migrations; do
        local down_file="$MIGRATIONS_DIR/${version}_down.sql"
        
        if [ -f "$down_file" ]; then
            log_info "回滚迁移: $version"
            
            psql "$DATABASE_URL" <<EOF
BEGIN;

-- 执行回滚脚本
\i $down_file

-- 删除迁移记录
DELETE FROM $MIGRATION_TABLE WHERE version = '$version';

COMMIT;
EOF
            
            log_success "回滚完成: $version"
        else
            log_warning "回滚脚本不存在: $down_file"
        fi
    done
}

# 生成迁移文件
generate_migration() {
    local name=$1
    
    if [ -z "$name" ]; then
        log_error "请提供迁移名称"
        exit 1
    fi
    
    # 生成时间戳
    local timestamp=$(date +%Y%m%d%H%M%S)
    local up_file="$MIGRATIONS_DIR/${timestamp}_${name}.sql"
    local down_file="$MIGRATIONS_DIR/${timestamp}_${name}_down.sql"
    
    # 创建迁移目录
    mkdir -p "$MIGRATIONS_DIR"
    
    # 创建迁移文件
    cat > "$up_file" <<EOF
-- Migration: $name
-- Created at: $(date)

-- 在此处添加迁移SQL

EOF
    
    cat > "$down_file" <<EOF
-- Rollback: $name
-- Created at: $(date)

-- 在此处添加回滚SQL

EOF
    
    log_success "迁移文件已创建:"
    echo "  UP:   $up_file"
    echo "  DOWN: $down_file"
}

# 执行种子数据
run_seeds() {
    local env=${1:-development}
    
    check_environment
    
    log_info "执行种子数据 (环境: $env)..."
    
    # 创建种子目录
    mkdir -p "$SEEDS_DIR"
    
    # 查找种子文件
    local seed_file="$SEEDS_DIR/seed_${env}.sql"
    
    if [ -f "$seed_file" ]; then
        log_info "执行种子文件: $seed_file"
        psql "$DATABASE_URL" < "$seed_file"
        log_success "种子数据执行完成"
    else
        log_warning "种子文件不存在: $seed_file"
    fi
}

# 数据库状态
migration_status() {
    check_environment
    
    log_info "数据库迁移状态:"
    echo ""
    
    # 显示已应用的迁移
    psql "$DATABASE_URL" <<EOF
SELECT 
    version,
    description,
    applied_at,
    execution_time_ms || 'ms' as execution_time
FROM $MIGRATION_TABLE
ORDER BY version DESC
LIMIT 10;
EOF
    
    # 统计信息
    echo ""
    log_info "统计信息:"
    psql "$DATABASE_URL" -t <<EOF
SELECT 
    'Total migrations: ' || COUNT(*) || ', ' ||
    'Total time: ' || SUM(execution_time_ms) || 'ms'
FROM $MIGRATION_TABLE;
EOF
}

# 数据库备份
backup_database() {
    local backup_name=${1:-"backup_$(date +%Y%m%d_%H%M%S)"}
    local backup_dir="$PROJECT_ROOT/backups"
    local backup_file="$backup_dir/${backup_name}.sql"
    
    mkdir -p "$backup_dir"
    
    log_info "备份数据库到: $backup_file"
    
    # 解析数据库URL
    local db_name=$(echo "$DATABASE_URL" | sed -n 's/.*\/\([^?]*\).*/\1/p')
    
    # 执行备份
    pg_dump "$DATABASE_URL" > "$backup_file"
    
    # 压缩备份
    gzip "$backup_file"
    
    log_success "备份完成: ${backup_file}.gz"
}

# 恢复数据库
restore_database() {
    local backup_file=$1
    
    if [ -z "$backup_file" ]; then
        log_error "请提供备份文件路径"
        exit 1
    fi
    
    if [ ! -f "$backup_file" ]; then
        log_error "备份文件不存在: $backup_file"
        exit 1
    fi
    
    log_warning "此操作将覆盖现有数据库，是否继续？[y/N]"
    read -r confirm
    
    if [ "$confirm" != "y" ] && [ "$confirm" != "Y" ]; then
        log_info "操作已取消"
        exit 0
    fi
    
    log_info "恢复数据库..."
    
    # 如果是gz文件，先解压
    if [[ "$backup_file" == *.gz ]]; then
        gunzip -c "$backup_file" | psql "$DATABASE_URL"
    else
        psql "$DATABASE_URL" < "$backup_file"
    fi
    
    log_success "数据库恢复完成"
}

# 显示帮助
show_help() {
    cat <<EOF
数据库迁移管理工具

用法:
    $0 <command> [options]

命令:
    up              执行所有待处理的迁移
    down [steps]    回滚指定数量的迁移 (默认: 1)
    generate <name> 生成新的迁移文件
    status          显示迁移状态
    seed [env]      执行种子数据 (默认: development)
    backup [name]   备份数据库
    restore <file>  恢复数据库
    help            显示帮助信息

环境变量:
    DATABASE_URL    数据库连接URL (必需)
    DB_TYPE         数据库类型 (默认: postgres)
    MIGRATION_TABLE 迁移记录表名 (默认: schema_migrations)

示例:
    $0 up                    # 执行所有迁移
    $0 down 2                # 回滚最近2个迁移
    $0 generate add_users    # 生成add_users迁移
    $0 seed production       # 执行生产环境种子数据
    $0 backup                # 备份数据库
EOF
}

# 主函数
main() {
    local command=${1:-help}
    
    case $command in
        up)
            migrate_up
            ;;
        down)
            migrate_down "${2:-1}"
            ;;
        generate)
            generate_migration "$2"
            ;;
        status)
            migration_status
            ;;
        seed)
            run_seeds "${2:-development}"
            ;;
        backup)
            backup_database "$2"
            ;;
        restore)
            restore_database "$2"
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