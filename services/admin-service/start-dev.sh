#!/bin/bash

echo "🚀 启动 OpenPenPal Admin Service 开发环境..."

# 检查Docker是否运行
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker未运行，请先启动Docker"
    exit 1
fi

# 启动开发环境数据库和Redis
echo "📦 启动开发环境数据库和缓存..."
docker-compose -f docker-compose.dev.yml up -d

# 等待数据库就绪
echo "⏳ 等待数据库启动..."
sleep 10

# 检查数据库连接
docker-compose -f docker-compose.dev.yml exec -T postgres-dev pg_isready -U postgres -d openpenpal

if [ $? -eq 0 ]; then
    echo "✅ 数据库连接正常"
else
    echo "❌ 数据库连接失败"
    exit 1
fi

echo "🌟 开发环境启动完成！"
echo ""
echo "🔧 可用服务："
echo "  - PostgreSQL: localhost:5432 (postgres/postgres)"
echo "  - Redis: localhost:6379"
echo "  - PgAdmin: http://localhost:5050 (admin@openpenpal.com/admin123)"
echo "  - Redis Commander: http://localhost:8081"
echo ""
echo "💻 启动Spring Boot应用："
echo "  cd backend && ./mvnw spring-boot:run -Dspring-boot.run.profiles=dev"
echo ""
echo "🛑 停止环境："
echo "  docker-compose -f docker-compose.dev.yml down"