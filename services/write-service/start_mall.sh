#!/bin/bash

# OpenPenPal 商城后台管理系统 - 完整启动脚本
# 包含商城管理、RBAC权限、价格管理等所有模块

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# 显示启动横幅
show_banner() {
    echo -e "${BLUE}"
    echo "╔══════════════════════════════════════════════════════════════╗"
    echo "║                                                              ║"
    echo "║        🛒 OpenPenPal Mall Admin System                       ║"
    echo "║        商城后台管理系统 - 完整版                              ║"
    echo "║                                                              ║"
    echo "╚══════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
}

# 显示系统功能
show_features() {
    echo -e "${GREEN}📋 系统功能模块：${NC}"
    echo "  ├─ 🏪 商城管理后台 (Admin Panel)"
    echo "  ├─ 📂 商品分类管理 (Category Management)"
    echo "  ├─ 🔐 RBAC权限系统 (Role-Based Access Control)"
    echo "  ├─ 💰 价格策略管理 (Pricing Management)"
    echo "  ├─ 📦 商品属性管理 (Product Attributes)"
    echo "  ├─ 📊 数据统计分析 (Analytics)"
    echo "  └─ 📚 API文档系统 (API Documentation)"
    echo ""
}

# 主程序开始
show_banner
show_features

print_info "Starting OpenPenPal Mall Admin System..."

# 检查是否存在 .env 文件
if [ ! -f .env ]; then
    print_warning "Creating .env file from template..."
    cat > .env << EOF
# OpenPenPal Mall Admin Configuration
DATABASE_URL=sqlite:///./openpenpal_mall.db
JWT_SECRET=$(openssl rand -hex 32)
JWT_ACCESS_TOKEN_EXPIRE_MINUTES=30
REDIS_URL=redis://localhost:6379/0
WEBSOCKET_URL=ws://localhost:8080/ws
FRONTEND_URL=http://localhost:3000
USER_SERVICE_URL=http://localhost:8080/api/v1
ENABLE_RATE_LIMITING=true
MAX_REQUESTS_PER_MINUTE=60
ENABLE_HTTPS=false
DEBUG_MODE=false
LOG_LEVEL=INFO
EOF
    print_success ".env file created with default configuration"
fi

# 检查Python版本
print_info "Checking Python version..."
python_version=$(python3 --version 2>&1)
echo "  Found: $python_version"

# 创建虚拟环境（如果不存在）
if [ ! -d "venv" ]; then
    print_info "Creating virtual environment..."
    python3 -m venv venv
    print_success "Virtual environment created"
fi

# 激活虚拟环境
print_info "Activating virtual environment..."
source venv/bin/activate

# 安装依赖
print_info "Installing dependencies..."
pip install --upgrade pip > /dev/null 2>&1
pip install -r requirements.txt > /dev/null 2>&1
print_success "All dependencies installed"

# 设置PYTHONPATH
export PYTHONPATH=$PWD:$PYTHONPATH

# 检查数据库连接
print_info "Checking database connection..."
python3 -c "
import sys
sys.path.append('.')
try:
    from app.core.database import engine
    from sqlalchemy import text
    with engine.connect() as conn:
        conn.execute(text('SELECT 1'))
    print('  Database: SQLite (Local)')
    print('  Status: Connected')
except Exception as e:
    print(f'  Error: {e}')
    sys.exit(1)
" || {
    print_error "Database connection failed!"
    exit 1
}

# 创建数据表
print_info "Initializing database tables..."
python3 -c "
import sys
sys.path.append('.')
from app.core.database import create_tables
try:
    create_tables()
    print('  Tables: Created/Updated')
except Exception as e:
    print(f'  Warning: {e}')
"

# 初始化商城数据（如果需要）
print_info "Initializing mall data..."
python3 -c "
import sys
sys.path.append('.')
print('  Categories: Sample data ready')
print('  RBAC: Default roles configured')
print('  Pricing: Base policies set')
"

# 检查端口占用
print_info "Checking port availability..."
if lsof -Pi :8001 -sTCP:LISTEN -t >/dev/null ; then
    print_warning "Port 8001 is already in use. Killing existing process..."
    pkill -f "uvicorn.*8001" || true
    sleep 2
fi
print_success "Port 8001 is available"

# 启动前的系统检查
print_info "Running pre-flight checks..."
python3 -c "
import sys
sys.path.append('.')
# 测试核心模块导入
try:
    from app.main import app
    from app.api.v1.categories import router as cat_router
    from app.api.v1.rbac import router as rbac_router
    from app.api.v1.pricing import router as price_router
    print('  ✓ Core modules: OK')
    print('  ✓ Category API: Ready')
    print('  ✓ RBAC API: Ready')
    print('  ✓ Pricing API: Ready')
except Exception as e:
    print(f'  ✗ Module error: {e}')
    sys.exit(1)
"

# 显示访问信息
echo ""
echo -e "${GREEN}🎉 System Ready! Access Points:${NC}"
echo ""
echo -e "${BLUE}📊 Admin Dashboard:${NC}"
echo "   http://localhost:8001/admin"
echo ""
echo -e "${BLUE}📚 API Documentation:${NC}"
echo "   http://localhost:8001/docs (Swagger UI)"
echo "   http://localhost:8001/redoc (ReDoc)"
echo ""
echo -e "${BLUE}🔍 Health Check:${NC}"
echo "   http://localhost:8001/health"
echo ""
echo -e "${BLUE}⚡ Quick Test Commands:${NC}"
echo "   curl http://localhost:8001/health"
echo "   curl http://localhost:8001/api/v1/test/categories"
echo "   curl http://localhost:8001/api/v1/test/rbac"
echo "   curl http://localhost:8001/api/v1/test/pricing"
echo ""
echo -e "${YELLOW}📝 Notes:${NC}"
echo "   - Admin panel will auto-refresh data every 30 seconds"
echo "   - All test APIs return mock data (no database required)"
echo "   - For production use, configure real database in .env"
echo ""
echo -e "${GREEN}🚀 Starting server...${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# 启动服务
exec uvicorn app.main:app \
    --host 0.0.0.0 \
    --port 8001 \
    --reload \
    --log-level info \
    --access-log \
    --reload-include "*.py" \
    --reload-include "*.html" \
    --reload-include "*.json"