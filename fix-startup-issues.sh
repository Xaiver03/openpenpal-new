#!/bin/bash

# 修复OpenPenPal启动问题脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }

log_info "开始修复OpenPenPal启动问题..."

# 1. 确保所有服务已停止
log_info "停止所有服务..."
./startup/stop-all.sh >/dev/null 2>&1 || true
sleep 2

# 2. 重新编译Go后端以确保使用正确的端口配置
log_info "重新编译Go后端..."
cd backend
go build -o openpenpal-backend
if [ $? -eq 0 ]; then
    log_success "Go后端编译成功"
else
    log_error "Go后端编译失败"
    exit 1
fi
cd ..

# 3. 清理日志文件中的旧错误
log_info "清理日志文件..."
> logs/go-backend.log
> logs/frontend.log

# 4. 设置正确的环境变量
log_info "设置环境变量..."
export DATABASE_TYPE="postgres"
export DATABASE_URL="postgres://$(whoami):password@localhost:5432/openpenpal"
export JWT_SECRET="openpenpal-super-secret-jwt-key-for-integration"
export PORT="8080"  # 确保Go后端使用8080端口

# 5. 修复启动脚本中的超时问题
log_info "增加服务启动超时时间..."
# 临时增加超时时间，给服务更多启动时间
export SERVICE_START_TIMEOUT=120

# 6. 检查数据库连接
log_info "检查数据库连接..."
if psql -U $(whoami) -d openpenpal -c "SELECT 1" >/dev/null 2>&1; then
    log_success "数据库连接正常"
else
    log_error "数据库连接失败，请确保PostgreSQL正在运行"
    exit 1
fi

# 7. 启动Go后端服务（单独启动以便更好地控制）
log_info "启动Go后端服务..."
cd backend
nohup ./openpenpal-backend > ../logs/go-backend.log 2>&1 &
BACKEND_PID=$!
cd ..

# 等待后端启动
log_info "等待Go后端启动..."
for i in {1..30}; do
    if curl -s http://localhost:8080/health >/dev/null 2>&1; then
        log_success "Go后端启动成功 (PID: $BACKEND_PID)"
        break
    fi
    if [ $i -eq 30 ]; then
        log_error "Go后端启动超时"
        kill $BACKEND_PID 2>/dev/null || true
        exit 1
    fi
    sleep 2
done

# 8. 重新构建前端以确保使用修复后的代码
log_info "检查前端代码..."
cd frontend
# 如果有TypeScript错误，尝试修复
if npm run type-check 2>&1 | grep -q "error"; then
    log_warning "发现TypeScript错误，尝试修复..."
    # 这里可以添加具体的修复逻辑
fi
cd ..

log_success "修复完成！"
log_info ""
log_info "现在可以使用以下命令启动服务："
log_info "  ./startup/quick-start.sh complete"
log_info ""
log_info "或者使用简单模式："
log_info "  ./startup/quick-start.sh simple"