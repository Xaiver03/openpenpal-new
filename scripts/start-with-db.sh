#!/bin/bash

# 带数据库启动脚本
# 优先使用PostgreSQL数据库，如果不可用则降级到mock服务

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m'

echoColor() {
    echo -e "${1}${2}${NC}"
}

# 脚本目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# 数据库配置
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_NAME=${DB_NAME:-openpenpal}
DB_USER=${DB_USER:-postgres}
DB_PASSWORD=${DB_PASSWORD:-password}

# 检查PostgreSQL是否可用
checkPostgreSQL() {
    echoColor $BLUE "🔍 检查PostgreSQL数据库..."
    
    if ! command -v psql &> /dev/null; then
        echoColor $YELLOW "⚠️  PostgreSQL客户端未安装，将使用mock服务"
        return 1
    fi
    
    export PGPASSWORD=$DB_PASSWORD
    if ! psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "SELECT 1;" &> /dev/null; then
        echoColor $YELLOW "⚠️  无法连接到PostgreSQL，将使用mock服务"
        return 1
    fi
    
    # 检查postcode表是否存在
    local table_count=$(psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "
        SELECT COUNT(*) FROM information_schema.tables 
        WHERE table_schema = 'public' AND table_name LIKE 'postcode_%';
    " 2>/dev/null | xargs || echo "0")
    
    if [[ "$table_count" -lt 5 ]]; then
        echoColor $YELLOW "⚠️  Postcode表不完整，正在初始化数据库..."
        if [[ -f "$SCRIPT_DIR/init-postcode-db.sh" ]]; then
            "$SCRIPT_DIR/init-postcode-db.sh"
        else
            echoColor $RED "❌ 找不到数据库初始化脚本"
            return 1
        fi
    fi
    
    echoColor $GREEN "✅ PostgreSQL数据库已就绪"
    return 0
}

# 启动服务
startServices() {
    local use_database=$1
    
    echoColor $PURPLE "🚀 启动OpenPenPal服务..."
    
    if [[ "$use_database" == "true" ]]; then
        echoColor $GREEN "📊 使用PostgreSQL数据库模式"
        
        # 设置数据库环境变量
        export DATABASE_URL="postgresql://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME"
        export USE_DATABASE=true
        export DB_HOST=$DB_HOST
        export DB_PORT=$DB_PORT
        export DB_NAME=$DB_NAME
        export DB_USER=$DB_USER
        export DB_PASSWORD=$DB_PASSWORD
        
        # 启动真实服务
        echoColor $BLUE "启动写信服务 (数据库模式)..."
        cd "$PROJECT_ROOT/services/write-service"
        python -m uvicorn app.main:app --host 0.0.0.0 --port 8001 --reload &
        WRITE_SERVICE_PID=$!
        
        # 启动其他服务（这里简化，实际应该启动完整的微服务）
        echoColor $BLUE "启动API网关..."
        cd "$PROJECT_ROOT/services/gateway"
        # go run main.go & # 需要Go环境
        # GATEWAY_PID=$!
        
    else
        echoColor $YELLOW "🎭 使用Mock服务模式"
        
        # 启动mock服务
        cd "$PROJECT_ROOT"
        node scripts/simple-mock-services.js &
        MOCK_SERVICE_PID=$!
    fi
    
    # 启动前端
    echoColor $BLUE "启动前端服务..."
    cd "$PROJECT_ROOT/frontend"
    npm run dev &
    FRONTEND_PID=$!
    
    # 等待服务启动
    echoColor $BLUE "⏳ 等待服务启动完成..."
    sleep 5
    
    # 显示服务状态
    showServiceStatus $use_database
}

# 显示服务状态
showServiceStatus() {
    local use_database=$1
    
    echoColor $PURPLE "\n📋 服务状态概览"
    echoColor $PURPLE "=" * 40
    
    # 检查服务健康状态
    local services=(
        "前端服务|http://localhost:3000/api/health"
        "API服务|http://localhost:8001/health"
        "网关服务|http://localhost:8000/health"
    )
    
    for service_info in "${services[@]}"; do
        IFS='|' read -r name url <<< "$service_info"
        
        if curl -s "$url" &> /dev/null; then
            echoColor $GREEN "✅ $name: 运行正常"
        else
            echoColor $YELLOW "⚠️  $name: 启动中或不可用"
        fi
    done
    
    echo ""
    if [[ "$use_database" == "true" ]]; then
        echoColor $GREEN "🎯 数据库模式已启动"
        echoColor $BLUE "   • 数据持久化: PostgreSQL"
        echoColor $BLUE "   • 测试数据: 已初始化"
    else
        echoColor $YELLOW "🎭 Mock模式已启动"
        echoColor $BLUE "   • 数据存储: 内存"
        echoColor $BLUE "   • 测试数据: 运行时生成"
    fi
    
    echo ""
    echoColor $BLUE "🔑 测试账号:"
    echoColor $BLUE "   • courier1/courier123 - 一级信使"
    echoColor $BLUE "   • courier2/courier123 - 二级信使"
    echoColor $BLUE "   • courier3/courier123 - 三级信使"
    echoColor $BLUE "   • courier4/courier123 - 四级信使"
    
    echo ""
    echoColor $BLUE "🌐 访问地址:"
    echoColor $BLUE "   • 前端应用: http://localhost:3000"
    echoColor $BLUE "   • API文档: http://localhost:8001/docs"
    echoColor $BLUE "   • 健康检查: http://localhost:8000/health"
    
    echo ""
    echoColor $PURPLE "按 Ctrl+C 停止所有服务"
}

# 清理函数
cleanup() {
    echoColor $YELLOW "\n🛑 正在停止服务..."
    
    # 杀死所有子进程
    if [[ -n "$WRITE_SERVICE_PID" ]]; then
        kill $WRITE_SERVICE_PID 2>/dev/null || true
    fi
    if [[ -n "$GATEWAY_PID" ]]; then
        kill $GATEWAY_PID 2>/dev/null || true
    fi
    if [[ -n "$MOCK_SERVICE_PID" ]]; then
        kill $MOCK_SERVICE_PID 2>/dev/null || true
    fi
    if [[ -n "$FRONTEND_PID" ]]; then
        kill $FRONTEND_PID 2>/dev/null || true
    fi
    
    # 清理可能的其他进程
    pkill -f "simple-mock-services" 2>/dev/null || true
    pkill -f "npm run dev" 2>/dev/null || true
    pkill -f "uvicorn" 2>/dev/null || true
    
    echoColor $GREEN "✅ 所有服务已停止"
    exit 0
}

# 设置信号处理
trap cleanup SIGINT SIGTERM

# 显示使用帮助
showUsage() {
    echoColor $BLUE "OpenPenPal 启动脚本"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -h, --help        显示此帮助信息"
    echo "  --mock-only       强制使用Mock服务模式"
    echo "  --db-only         强制使用数据库模式（如果数据库不可用则失败）"
    echo "  --init-db         初始化数据库后启动"
    echo "  --test            启动后运行集成测试"
    echo ""
    echo "环境变量:"
    echo "  DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME"
    echo ""
    echo "示例:"
    echo "  $0                    # 自动检测数据库"
    echo "  $0 --mock-only        # 仅使用Mock服务"
    echo "  $0 --init-db --test   # 初始化数据库并测试"
}

# 解析命令行参数
MOCK_ONLY=false
DB_ONLY=false
INIT_DB=false
RUN_TEST=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            showUsage
            exit 0
            ;;
        --mock-only)
            MOCK_ONLY=true
            shift
            ;;
        --db-only)
            DB_ONLY=true
            shift
            ;;
        --init-db)
            INIT_DB=true
            shift
            ;;
        --test)
            RUN_TEST=true
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
    echoColor $PURPLE "🎉 OpenPenPal Postcode系统启动器"
    echoColor $PURPLE "=" * 50
    
    # 初始化数据库（如果需要）
    if [[ "$INIT_DB" == "true" ]]; then
        echoColor $BLUE "🔧 初始化数据库..."
        "$SCRIPT_DIR/init-postcode-db.sh"
    fi
    
    # 决定使用哪种模式
    local use_database=false
    
    if [[ "$MOCK_ONLY" == "true" ]]; then
        echoColor $YELLOW "🎭 强制使用Mock服务模式"
        use_database=false
    elif [[ "$DB_ONLY" == "true" ]]; then
        echoColor $BLUE "📊 强制使用数据库模式"
        if checkPostgreSQL; then
            use_database=true
        else
            echoColor $RED "❌ 数据库模式要求失败"
            exit 1
        fi
    else
        # 自动检测
        if checkPostgreSQL; then
            use_database=true
        else
            use_database=false
        fi
    fi
    
    # 启动服务
    startServices $use_database
    
    # 运行测试（如果需要）
    if [[ "$RUN_TEST" == "true" ]]; then
        echoColor $BLUE "🧪 运行集成测试..."
        sleep 3  # 等待服务完全启动
        python3 "$SCRIPT_DIR/test-postcode-db.py" || true
    fi
    
    # 等待用户中断
    while true; do
        sleep 1
    done
}

# 执行主函数
main "$@"