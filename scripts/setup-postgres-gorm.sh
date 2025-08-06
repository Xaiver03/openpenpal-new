#!/bin/bash

# GORM + PostgreSQL 设置脚本

set -e

echo "🚀 GORM + PostgreSQL 设置开始"
echo "============================"

# 检查 Docker
if ! command -v docker &> /dev/null; then
    echo "❌ Docker 未安装"
    echo "请先安装 Docker 或使用本地 PostgreSQL"
    exit 1
fi

# 启动 PostgreSQL
echo "📦 启动 PostgreSQL..."
docker run --name openpenpal-postgres \
  -e POSTGRES_USER=openpenpal \
  -e POSTGRES_PASSWORD=openpenpal123 \
  -e POSTGRES_DB=openpenpal \
  -p 5432:5432 \
  -d postgres:15-alpine 2>/dev/null || {
    echo "PostgreSQL 容器已存在，尝试启动..."
    docker start openpenpal-postgres
}

# 等待 PostgreSQL 启动
echo "⏳ 等待 PostgreSQL 启动..."
sleep 5

# 测试连接
echo "🔍 测试数据库连接..."
docker exec openpenpal-postgres pg_isready -U openpenpal || {
    echo "❌ 无法连接到 PostgreSQL"
    exit 1
}

echo "✅ PostgreSQL 已准备就绪"

# 创建环境变量文件
echo "📝 创建环境变量文件..."
cat > backend/.env.postgres << EOF
# PostgreSQL 配置
DATABASE_TYPE=postgres
DATABASE_URL=postgresql://openpenpal:openpenpal123@localhost:5432/openpenpal

# 或使用分离的配置
DB_HOST=localhost
DB_USER=openpenpal
DB_PASSWORD=openpenpal123
DB_NAME=openpenpal
DB_PORT=5432
EOF

echo "✅ 环境配置完成"

# 提示下一步
echo ""
echo "✨ PostgreSQL 设置完成！"
echo ""
echo "下一步："
echo "1. 更新 backend/go.mod 添加 PostgreSQL 驱动："
echo "   go get gorm.io/driver/postgres"
echo ""
echo "2. 运行应用使用 PostgreSQL："
echo "   cd backend"
echo "   cp .env.postgres .env"
echo "   go run main.go"
echo ""
echo "3. 查看数据库："
echo "   docker exec -it openpenpal-postgres psql -U openpenpal"
echo ""
echo "4. 停止 PostgreSQL："
echo "   docker stop openpenpal-postgres"
echo ""