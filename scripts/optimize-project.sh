#!/bin/bash

# OpenPenPal项目优化脚本
# 用于清理项目、优化结构和提升性能

echo "🚀 开始OpenPenPal项目优化..."
echo "========================================"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 统计变量
CLEANED_FILES=0
SAVED_SPACE=0

# 函数：记录清理的文件大小
log_cleanup() {
    local file="$1"
    if [[ -f "$file" ]]; then
        local size=$(stat -f%z "$file" 2>/dev/null || echo 0)
        SAVED_SPACE=$((SAVED_SPACE + size))
        CLEANED_FILES=$((CLEANED_FILES + 1))
    fi
}

echo -e "${BLUE}📋 第1步: 项目体积分析${NC}"
echo "----------------------------------------"

# 分析当前项目大小
echo "🔍 分析项目结构..."
TOTAL_SIZE=$(du -sh . 2>/dev/null | awk '{print $1}')
NODE_MODULES_COUNT=$(find . -name "node_modules" -type d | wc -l | xargs)
LOG_FILES_COUNT=$(find . -name "*.log" | wc -l | xargs)

echo "📊 当前项目状态:"
echo "   • 总大小: $TOTAL_SIZE"
echo "   • node_modules目录: $NODE_MODULES_COUNT 个"
echo "   • 日志文件: $LOG_FILES_COUNT 个"
echo ""

echo -e "${YELLOW}🧹 第2步: 清理临时文件${NC}"
echo "----------------------------------------"

# 清理日志文件
echo "🗑️  清理日志文件..."
find . -name "*.log" -type f | while read file; do
    log_cleanup "$file"
    echo "   删除: $file"
    rm -f "$file"
done

# 清理PID文件
echo "🗑️  清理PID文件..."
find . -name "*.pid" -type f | while read file; do
    log_cleanup "$file"
    echo "   删除: $file"
    rm -f "$file"
done

# 清理临时文件
echo "🗑️  清理临时文件..."
find . -name "*.tmp" -o -name "*.temp" -o -name "*.bak" | while read file; do
    log_cleanup "$file"
    echo "   删除: $file"
    rm -f "$file"
done

# 清理tsbuildinfo文件
echo "🗑️  清理TypeScript构建缓存..."
find . -name "*.tsbuildinfo" | while read file; do
    log_cleanup "$file"
    echo "   删除: $file"
    rm -f "$file"
done

echo ""

echo -e "${BLUE}📁 第3步: 创建标准化目录结构${NC}"
echo "----------------------------------------"

# 创建统一的日志目录
if [[ ! -d "logs" ]]; then
    mkdir -p logs
    echo "✅ 创建 logs/ 目录"
fi

# 创建配置目录结构
if [[ ! -d "config/templates" ]]; then
    mkdir -p config/templates
    echo "✅ 创建 config/templates/ 目录"
fi

# 创建部署配置目录
if [[ ! -d "deploy" ]]; then
    mkdir -p deploy/{docker,k8s,monitoring}
    echo "✅ 创建部署配置目录结构"
fi

echo ""

echo -e "${GREEN}⚙️ 第4步: 创建配置文件模板${NC}"
echo "----------------------------------------"

# 创建环境变量模板
cat > .env.template << 'EOF'
# OpenPenPal环境配置模板
# 复制此文件为 .env.local 并填入实际值

# ===================
# 应用基础配置
# ===================
NODE_ENV=development
NEXT_PUBLIC_APP_NAME=OpenPenPal
NEXT_PUBLIC_APP_VERSION=2.1.0

# ===================
# API网关配置
# ===================
NEXT_PUBLIC_GATEWAY_URL=http://localhost:8000
NEXT_PUBLIC_API_URL=http://localhost:8000/api/v1
NEXT_PUBLIC_WS_URL=ws://localhost:8000/ws

# ===================
# 微服务地址配置
# ===================
NEXT_PUBLIC_WRITE_SERVICE_URL=http://localhost:8001
NEXT_PUBLIC_COURIER_SERVICE_URL=http://localhost:8002
NEXT_PUBLIC_ADMIN_SERVICE_URL=http://localhost:8003
NEXT_PUBLIC_OCR_SERVICE_URL=http://localhost:8004

# ===================
# 数据库配置
# ===================
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_NAME=openpenpal
DATABASE_USER=postgres
DATABASE_PASSWORD=your_password_here

# ===================
# Redis配置
# ===================
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=your_redis_password_here

# ===================
# JWT配置
# ===================
JWT_SECRET=your_super_secret_jwt_key_here
JWT_EXPIRATION=7d

# ===================
# 文件存储配置
# ===================
UPLOAD_MAX_SIZE=10MB
STATIC_FILES_PATH=./uploads

# ===================
# 外部服务配置
# ===================
# 邮件服务
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your_email@gmail.com
SMTP_PASS=your_email_password

# OSS存储(可选)
OSS_ENDPOINT=
OSS_ACCESS_KEY=
OSS_SECRET_KEY=
OSS_BUCKET=

# ===================
# 监控配置
# ===================
ENABLE_MONITORING=true
METRICS_PORT=9090
JAEGER_ENDPOINT=http://localhost:14268/api/traces
EOF

echo "✅ 创建 .env.template 文件"

# 创建Docker Compose生产环境配置
cat > docker-compose.production.yml << 'EOF'
version: '3.8'

services:
  # Nginx反向代理
  nginx:
    image: nginx:alpine
    container_name: openpenpal-nginx
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./config/nginx.prod.conf:/etc/nginx/nginx.conf
      - ./ssl:/etc/nginx/ssl
    depends_on:
      - frontend
      - api-gateway
    restart: unless-stopped
    networks:
      - openpenpal-network

  # 前端服务
  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile.prod
    container_name: openpenpal-frontend
    environment:
      - NODE_ENV=production
    expose:
      - "3000"
    restart: unless-stopped
    networks:
      - openpenpal-network

  # API网关
  api-gateway:
    build:
      context: ./services/gateway
      dockerfile: Dockerfile
    container_name: openpenpal-gateway
    environment:
      - NODE_ENV=production
      - DATABASE_URL=postgresql://postgres:${POSTGRES_PASSWORD}@postgres:5432/openpenpal
      - REDIS_URL=redis://redis:6379
    expose:
      - "8000"
    depends_on:
      - postgres
      - redis
    restart: unless-stopped
    networks:
      - openpenpal-network

  # 数据库
  postgres:
    image: postgres:15-alpine
    container_name: openpenpal-db
    environment:
      - POSTGRES_DB=openpenpal
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./config/init.sql:/docker-entrypoint-initdb.d/init.sql
    ports:
      - "5432:5432"
    restart: unless-stopped
    networks:
      - openpenpal-network

  # Redis缓存
  redis:
    image: redis:7-alpine
    container_name: openpenpal-redis
    command: redis-server --appendonly yes --requirepass ${REDIS_PASSWORD}
    volumes:
      - redis_data:/data
    ports:
      - "6379:6379"
    restart: unless-stopped
    networks:
      - openpenpal-network

  # 监控 - Prometheus
  prometheus:
    image: prom/prometheus:latest
    container_name: openpenpal-prometheus
    ports:
      - "9090:9090"
    volumes:
      - ./config/prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus_data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/etc/prometheus/console_libraries'
      - '--web.console.templates=/etc/prometheus/consoles'
    restart: unless-stopped
    networks:
      - openpenpal-network

  # 监控 - Grafana
  grafana:
    image: grafana/grafana:latest
    container_name: openpenpal-grafana
    ports:
      - "3001:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=${GRAFANA_PASSWORD}
    volumes:
      - grafana_data:/var/lib/grafana
      - ./config/grafana:/etc/grafana/provisioning
    restart: unless-stopped
    networks:
      - openpenpal-network

volumes:
  postgres_data:
  redis_data:
  prometheus_data:
  grafana_data:

networks:
  openpenpal-network:
    driver: bridge
EOF

echo "✅ 创建 docker-compose.production.yml 文件"

echo ""

echo -e "${BLUE}📊 第5步: 项目分析和依赖优化建议${NC}"
echo "----------------------------------------"

# 分析package.json文件数量
PACKAGE_JSON_COUNT=$(find . -name "package.json" | wc -l | xargs)
echo "📦 发现 $PACKAGE_JSON_COUNT 个 package.json 文件"

# 检查是否可以使用pnpm workspace
echo "💡 建议实施以下优化:"
echo "   1. 使用 pnpm workspace 替代多个 node_modules"
echo "   2. 实施 monorepo 架构统一依赖管理"
echo "   3. 添加 Dockerfile.prod 进行多阶段构建优化"
echo "   4. 配置 nginx 进行静态资源压缩和缓存"

echo ""

echo -e "${GREEN}✅ 第6步: 创建优化的启动脚本${NC}"
echo "----------------------------------------"

# 创建优化的启动脚本
cat > scripts/start-optimized.sh << 'EOF'
#!/bin/bash

# OpenPenPal优化版启动脚本
# 包含性能监控和健康检查

echo "🚀 启动OpenPenPal优化版..."

# 检查环境配置
if [[ ! -f ".env.local" ]]; then
    echo "⚠️  警告: .env.local 文件不存在"
    echo "📋 请复制 .env.template 为 .env.local 并填入配置"
    exit 1
fi

# 检查Docker是否运行
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker未运行，请启动Docker"
    exit 1
fi

# 启动监控服务
echo "📊 启动监控服务..."
docker-compose -f docker-compose.production.yml up -d prometheus grafana

# 启动核心服务
echo "🏗️  启动核心服务..."
docker-compose -f docker-compose.production.yml up -d postgres redis

# 等待数据库启动
echo "⏳ 等待数据库启动..."
sleep 10

# 启动应用服务
echo "🚀 启动应用服务..."
docker-compose -f docker-compose.production.yml up -d

# 健康检查
echo "🔍 执行健康检查..."
sleep 5

# 检查服务状态
services=("nginx" "frontend" "api-gateway" "postgres" "redis")
for service in "${services[@]}"; do
    if docker-compose -f docker-compose.production.yml ps -q $service > /dev/null 2>&1; then
        echo "✅ $service 服务运行正常"
    else
        echo "❌ $service 服务启动失败"
    fi
done

echo ""
echo "🎉 OpenPenPal启动完成！"
echo "📋 访问地址:"
echo "   • 前端: http://localhost"
echo "   • API文档: http://localhost/api/docs"
echo "   • 监控面板: http://localhost:3001 (admin/admin)"
echo "   • 指标监控: http://localhost:9090"
EOF

chmod +x scripts/start-optimized.sh
echo "✅ 创建优化启动脚本"

echo ""

echo -e "${GREEN}🎊 优化完成统计${NC}"
echo "========================================"

# 计算节省的空间
if [[ $SAVED_SPACE -gt 0 ]]; then
    SAVED_MB=$((SAVED_SPACE / 1024 / 1024))
    echo -e "清理文件数: ${GREEN}${CLEANED_FILES}${NC}"
    echo -e "节省空间: ${GREEN}${SAVED_MB}MB${NC}"
else
    echo -e "清理文件数: ${GREEN}${CLEANED_FILES}${NC}"
fi

echo ""
echo -e "${YELLOW}📋 下一步建议:${NC}"
echo "1. 复制 .env.template 为 .env.local 并填入配置"
echo "2. 运行 ./scripts/start-optimized.sh 启动优化版本"
echo "3. 访问 http://localhost:3001 查看监控面板"
echo "4. 考虑实施 pnpm workspace 进一步优化依赖管理"

echo ""
echo -e "${GREEN}✨ OpenPenPal项目优化第一阶段完成！${NC}"