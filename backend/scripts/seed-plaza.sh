#!/bin/bash

# 种子数据脚本 - 为写作广场添加测试数据

cd "$(dirname "$0")/.."

echo "🌱 开始为写作广场添加测试数据..."
echo ""

# 检查数据库连接
if ! go run scripts/seed-plaza-data.go 2>&1 | grep -q "pq:"; then
    echo "✅ 数据库连接正常"
else
    echo "❌ 数据库连接失败，请确保 PostgreSQL 正在运行"
    exit 1
fi

# 运行种子脚本
go run scripts/seed-plaza-data.go

echo ""
echo "📝 提示："
echo "1. 请确保后端服务正在运行 (端口 8080)"
echo "2. 请确保前端服务正在运行 (端口 3000)"
echo "3. 访问 http://localhost:3000/plaza 查看写作广场"
echo ""
echo "如果看不到数据，请检查："
echo "- Network 标签中的 API 请求是否成功"
echo "- Console 中是否有错误信息"