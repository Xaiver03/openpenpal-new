#!/bin/bash

# OpenPenPal监控系统启动脚本

echo "📊 启动OpenPenPal监控系统..."

# 检查Docker
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker未运行，请启动Docker"
    exit 1
fi

# 创建网络
docker network create openpenpal-monitoring 2>/dev/null || true

# 启动监控服务
echo "🏗️  启动监控服务..."
docker-compose -f docker-compose.monitoring.yml up -d

# 等待服务启动
echo "⏳ 等待服务启动..."
sleep 10

# 健康检查
echo "🔍 执行健康检查..."
services=("prometheus" "grafana" "alertmanager" "node-exporter")
for service in "${services[@]}"; do
    if docker-compose -f docker-compose.monitoring.yml ps -q $service > /dev/null 2>&1; then
        echo "✅ $service 运行正常"
    else
        echo "❌ $service 启动失败"
    fi
done

echo ""
echo "🎉 监控系统启动完成！"
echo "📋 访问地址:"
echo "   • Prometheus: http://localhost:9090"
echo "   • Grafana: http://localhost:3001 (admin/admin123)"
echo "   • AlertManager: http://localhost:9093"
echo "   • Node Exporter: http://localhost:9100"
