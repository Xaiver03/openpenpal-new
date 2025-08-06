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
