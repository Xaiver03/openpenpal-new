#!/bin/bash

# OpenPenPal Mock数据导入脚本
# 将所有Mock数据导入PostgreSQL数据库

set -e  # 遇到错误立即退出

# 配置数据库连接参数
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-openpenpal}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-password}"

# 颜色输出函数
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echoColor() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# 检查数据库连接
checkDatabaseConnection() {
    echoColor $BLUE "🔍 检查数据库连接..."
    
    export PGPASSWORD=$DB_PASSWORD
    if ! psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -c "SELECT 1;" &> /dev/null; then
        echoColor $RED "❌ 无法连接到PostgreSQL数据库"
        echoColor $RED "   请确保数据库服务正在运行，并且连接参数正确"
        exit 1
    fi
    
    echoColor $GREEN "✅ 数据库连接成功"
}

# 检查目标数据库是否存在
checkTargetDatabase() {
    echoColor $BLUE "🔍 检查目标数据库..."
    
    if ! psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -tc "SELECT 1 FROM pg_database WHERE datname = '$DB_NAME'" | grep -q 1; then
        echoColor $YELLOW "⚠️  数据库 '$DB_NAME' 不存在，正在创建..."
        psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -c "CREATE DATABASE $DB_NAME;"
        echoColor $GREEN "✅ 数据库 '$DB_NAME' 创建成功"
    else
        echoColor $GREEN "✅ 数据库 '$DB_NAME' 已存在"
    fi
}

# 执行SQL文件
executeSQLFile() {
    local sql_file=$1
    local description=$2
    
    if [ ! -f "$sql_file" ]; then
        echoColor $RED "❌ SQL文件不存在: $sql_file"
        return 1
    fi
    
    echoColor $BLUE "📝 执行 $description..."
    echoColor $BLUE "   文件: $sql_file"
    
    if psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f "$sql_file"; then
        echoColor $GREEN "✅ $description 完成"
    else
        echoColor $RED "❌ $description 失败"
        return 1
    fi
}

# 显示帮助信息
showHelp() {
    echo "OpenPenPal Mock数据导入脚本"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -h, --help              显示帮助信息"
    echo "  --host HOST             数据库主机 (默认: localhost)"
    echo "  --port PORT             数据库端口 (默认: 5432)"
    echo "  --database DATABASE     数据库名称 (默认: openpenpal)"
    echo "  --user USER             数据库用户 (默认: postgres)"
    echo "  --password PASSWORD     数据库密码 (默认: password)"
    echo "  --skip-postcode         跳过Postcode数据导入"
    echo "  --skip-courier          跳过信使管理数据导入"
    echo "  --skip-museum           跳过博物馆数据导入"
    echo "  --skip-plaza            跳过广场数据导入"
    echo "  --dry-run               仅显示将要执行的操作，不实际执行"
    echo ""
    echo "环境变量:"
    echo "  DB_HOST                 数据库主机"
    echo "  DB_PORT                 数据库端口"
    echo "  DB_NAME                 数据库名称"
    echo "  DB_USER                 数据库用户"
    echo "  DB_PASSWORD             数据库密码"
    echo ""
    echo "示例:"
    echo "  $0                      使用默认配置导入所有数据"
    echo "  $0 --host mydb.com --user myuser --password mypass"
    echo "  $0 --skip-postcode      跳过Postcode数据导入"
}

# 解析命令行参数
SKIP_POSTCODE=false
SKIP_COURIER=false
SKIP_MUSEUM=false
SKIP_PLAZA=false
DRY_RUN=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            showHelp
            exit 0
            ;;
        --host)
            DB_HOST="$2"
            shift 2
            ;;
        --port)
            DB_PORT="$2"
            shift 2
            ;;
        --database)
            DB_NAME="$2"
            shift 2
            ;;
        --user)
            DB_USER="$2"
            shift 2
            ;;
        --password)
            DB_PASSWORD="$2"
            shift 2
            ;;
        --skip-postcode)
            SKIP_POSTCODE=true
            shift
            ;;
        --skip-courier)
            SKIP_COURIER=true
            shift
            ;;
        --skip-museum)
            SKIP_MUSEUM=true
            shift
            ;;
        --skip-plaza)
            SKIP_PLAZA=true
            shift
            ;;
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        *)
            echoColor $RED "❌ 未知选项: $1"
            showHelp
            exit 1
            ;;
    esac
done

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DATA_DIR="$SCRIPT_DIR/../services/database/seed_data"

# 主执行函数
main() {
    echoColor $GREEN "🚀 OpenPenPal Mock数据导入开始"
    echoColor $BLUE "📋 配置信息:"
    echoColor $BLUE "   数据库主机: $DB_HOST"
    echoColor $BLUE "   数据库端口: $DB_PORT"
    echoColor $BLUE "   数据库名称: $DB_NAME"
    echoColor $BLUE "   数据库用户: $DB_USER"
    echoColor $BLUE "   数据目录: $DATA_DIR"
    echo ""
    
    if [ "$DRY_RUN" = true ]; then
        echoColor $YELLOW "🔍 DRY RUN模式 - 仅显示将要执行的操作"
        echo ""
    fi
    
    # 检查数据目录
    if [ ! -d "$DATA_DIR" ]; then
        echoColor $RED "❌ 数据目录不存在: $DATA_DIR"
        exit 1
    fi
    
    if [ "$DRY_RUN" = false ]; then
        checkDatabaseConnection
        checkTargetDatabase
    fi
    
    # 导入数据文件列表
    declare -A DATA_FILES
    DATA_FILES["Postcode系统数据"]="$DATA_DIR/postcode_test_data.sql"
    DATA_FILES["信使管理数据"]="$DATA_DIR/courier_management_data.sql"
    DATA_FILES["博物馆信件数据"]="$DATA_DIR/museum_letters_data.sql"
    DATA_FILES["广场公开信件数据"]="$DATA_DIR/plaza_public_letters_data.sql"
    
    # 检查跳过选项
    if [ "$SKIP_POSTCODE" = true ]; then
        unset DATA_FILES["Postcode系统数据"]
        echoColor $YELLOW "⏭️  跳过Postcode数据导入"
    fi
    
    if [ "$SKIP_COURIER" = true ]; then
        unset DATA_FILES["信使管理数据"]
        echoColor $YELLOW "⏭️  跳过信使管理数据导入"
    fi
    
    if [ "$SKIP_MUSEUM" = true ]; then
        unset DATA_FILES["博物馆信件数据"]
        echoColor $YELLOW "⏭️  跳过博物馆数据导入"
    fi
    
    if [ "$SKIP_PLAZA" = true ]; then
        unset DATA_FILES["广场公开信件数据"]
        echoColor $YELLOW "⏭️  跳过广场数据导入"
    fi
    
    # 执行数据导入
    local success_count=0
    local total_count=${#DATA_FILES[@]}
    
    for description in "${!DATA_FILES[@]}"; do
        local sql_file="${DATA_FILES[$description]}"
        
        if [ "$DRY_RUN" = true ]; then
            echoColor $BLUE "📝 将要执行: $description"
            echoColor $BLUE "   文件: $sql_file"
            if [ -f "$sql_file" ]; then
                echoColor $GREEN "   ✅ 文件存在"
            else
                echoColor $RED "   ❌ 文件不存在"
            fi
            echo ""
        else
            if executeSQLFile "$sql_file" "$description"; then
                ((success_count++))
            fi
        fi
    done
    
    echo ""
    if [ "$DRY_RUN" = true ]; then
        echoColor $BLUE "🔍 DRY RUN完成，共检查 $total_count 个数据文件"
    else
        if [ $success_count -eq $total_count ]; then
            echoColor $GREEN "🎉 所有Mock数据导入完成！"
            echoColor $GREEN "   成功导入: $success_count/$total_count"
        else
            echoColor $YELLOW "⚠️  部分数据导入完成"
            echoColor $YELLOW "   成功导入: $success_count/$total_count"
        fi
        
        # 显示数据统计
        echoColor $BLUE "📊 数据统计:"
        psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
            SELECT 
                'Mock数据总览' as category,
                'couriers' as table_name,
                COUNT(*) as count 
            FROM couriers WHERE id LIKE 'mock_%'
            UNION ALL
            SELECT 'Mock数据总览', 'users', COUNT(*) FROM users WHERE id LIKE 'mock_%'
            UNION ALL  
            SELECT 'Mock数据总览', 'letters', COUNT(*) FROM letters WHERE id LIKE 'mock_%'
            UNION ALL
            SELECT 'Mock数据总览', 'postcode_schools', COUNT(*) FROM postcode_schools WHERE id LIKE 'mock_%' OR id LIKE '550e8400%'
            ORDER BY table_name;
        " 2>/dev/null || echoColor $YELLOW "⚠️  无法获取数据统计（表可能不存在）"
    fi
}

# 清理函数
cleanup() {
    echoColor $BLUE "🧹 清理临时文件..."
    unset PGPASSWORD
}

# 设置清理trap
trap cleanup EXIT

# 执行主函数
main "$@"