#!/bin/bash

# Postcode数据库初始化脚本
# 用于在PostgreSQL中创建表结构并插入测试数据

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 数据库连接配置
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_NAME=${DB_NAME:-openpenpal}
DB_USER=${DB_USER:-postgres}
DB_PASSWORD=${DB_PASSWORD:-password}

# 脚本目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
SQL_DIR="$PROJECT_ROOT/services/database"

echoColor() {
    echo -e "${1}${2}${NC}"
}

# 检查PostgreSQL连接
checkDatabaseConnection() {
    echoColor $BLUE "🔍 检查数据库连接..."
    
    if ! command -v psql &> /dev/null; then
        echoColor $RED "❌ PostgreSQL客户端未安装"
        echoColor $YELLOW "请先安装PostgreSQL: brew install postgresql"
        exit 1
    fi
    
    # 测试连接
    export PGPASSWORD=$DB_PASSWORD
    if ! psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -c "SELECT 1;" &> /dev/null; then
        echoColor $RED "❌ 无法连接到PostgreSQL数据库"
        echoColor $YELLOW "请确认数据库服务已启动，连接信息正确："
        echoColor $YELLOW "  Host: $DB_HOST:$DB_PORT"
        echoColor $YELLOW "  User: $DB_USER"
        echoColor $YELLOW "  Database: $DB_NAME"
        exit 1
    fi
    
    echoColor $GREEN "✅ 数据库连接成功"
}

# 创建数据库（如果不存在）
createDatabase() {
    echoColor $BLUE "🏗️  创建数据库..."
    
    export PGPASSWORD=$DB_PASSWORD
    if ! psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -tc "SELECT 1 FROM pg_database WHERE datname = '$DB_NAME'" | grep -q 1; then
        echoColor $YELLOW "数据库 $DB_NAME 不存在，正在创建..."
        psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -c "CREATE DATABASE $DB_NAME;"
        echoColor $GREEN "✅ 数据库 $DB_NAME 创建成功"
    else
        echoColor $GREEN "✅ 数据库 $DB_NAME 已存在"
    fi
}

# 创建表结构
createTables() {
    echoColor $BLUE "📋 创建Postcode表结构..."
    
    export PGPASSWORD=$DB_PASSWORD
    if [[ -f "$SQL_DIR/migrations/001_create_postcode_tables.sql" ]]; then
        psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f "$SQL_DIR/migrations/001_create_postcode_tables.sql"
        echoColor $GREEN "✅ 表结构创建成功"
    else
        echoColor $RED "❌ 找不到表结构文件: $SQL_DIR/migrations/001_create_postcode_tables.sql"
        exit 1
    fi
}

# 插入测试数据
insertTestData() {
    echoColor $BLUE "🎯 插入Postcode测试数据..."
    
    export PGPASSWORD=$DB_PASSWORD
    if [[ -f "$SQL_DIR/seed_data/postcode_test_data.sql" ]]; then
        psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f "$SQL_DIR/seed_data/postcode_test_data.sql"
        echoColor $GREEN "✅ 测试数据插入成功"
    else
        echoColor $RED "❌ 找不到测试数据文件: $SQL_DIR/seed_data/postcode_test_data.sql"
        exit 1
    fi
}

# 验证数据
verifyData() {
    echoColor $BLUE "🔍 验证数据完整性..."
    
    export PGPASSWORD=$DB_PASSWORD
    local result=$(psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "
        SELECT 
            COUNT(*) as total_rooms,
            COUNT(DISTINCT school_code) as schools,
            COUNT(DISTINCT school_code || area_code) as areas,
            COUNT(DISTINCT school_code || area_code || building_code) as buildings
        FROM postcode_rooms;
    ")
    
    if [[ -n "$result" ]]; then
        echoColor $GREEN "✅ 数据验证结果:"
        echo "$result" | while read line; do
            echoColor $GREEN "   $line"
        done
        
        # 显示一些示例postcode
        echoColor $BLUE "📝 示例Postcode编码:"
        psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "
            SELECT 
                '  ' || full_postcode || ' - ' || s.name || a.name || b.name || r.name
            FROM postcode_rooms r
            JOIN postcode_buildings b ON b.school_code = r.school_code AND b.area_code = r.area_code AND b.code = r.building_code
            JOIN postcode_areas a ON a.school_code = r.school_code AND a.code = r.area_code  
            JOIN postcode_schools s ON s.code = r.school_code
            ORDER BY r.full_postcode
            LIMIT 5;
        " | while read line; do
            echoColor $GREEN "$line"
        done
    else
        echoColor $RED "❌ 数据验证失败"
        exit 1
    fi
}

# 显示使用帮助
showUsage() {
    echoColor $BLUE "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -h, --help     显示此帮助信息"
    echo "  --host HOST    数据库主机 (默认: localhost)"
    echo "  --port PORT    数据库端口 (默认: 5432)"
    echo "  --user USER    数据库用户 (默认: postgres)"
    echo "  --password PWD 数据库密码 (默认: password)"
    echo "  --database DB  数据库名称 (默认: openpenpal)"
    echo "  --tables-only  仅创建表结构，不插入数据"
    echo "  --data-only    仅插入数据，不创建表结构"
    echo ""
    echo "环境变量:"
    echo "  DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME"
    echo ""
    echo "示例:"
    echo "  $0                           # 使用默认配置"
    echo "  $0 --host mydb.com --user admin  # 指定主机和用户"
    echo "  DB_NAME=testdb $0            # 使用环境变量"
}

# 解析命令行参数
TABLES_ONLY=false
DATA_ONLY=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            showUsage
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
        --user)
            DB_USER="$2"
            shift 2
            ;;
        --password)
            DB_PASSWORD="$2"
            shift 2
            ;;
        --database)
            DB_NAME="$2"
            shift 2
            ;;
        --tables-only)
            TABLES_ONLY=true
            shift
            ;;
        --data-only)
            DATA_ONLY=true
            shift
            ;;
        *)
            echoColor $RED "未知选项: $1"
            showUsage
            exit 1
            ;;
    esac
done

# 主执行流程
main() {
    echoColor $BLUE "🚀 开始初始化Postcode数据库..."
    echoColor $BLUE "数据库配置: $DB_USER@$DB_HOST:$DB_PORT/$DB_NAME"
    echo ""
    
    checkDatabaseConnection
    createDatabase
    
    if [[ "$DATA_ONLY" != "true" ]]; then
        createTables
    fi
    
    if [[ "$TABLES_ONLY" != "true" ]]; then
        insertTestData
        verifyData
    fi
    
    echo ""
    echoColor $GREEN "🎉 Postcode数据库初始化完成！"
    echoColor $YELLOW "💡 现在可以启动应用并使用以下测试账号:"
    echoColor $YELLOW "   • courier1/courier123 - 一级信使 (PKA1**)"
    echoColor $YELLOW "   • courier2/courier123 - 二级信使 (PKA*)"
    echoColor $YELLOW "   • courier3/courier123 - 三级信使 (PK*)"
    echoColor $YELLOW "   • courier4/courier123 - 四级信使 (**)"
    echo ""
}

# 执行主函数
main "$@"