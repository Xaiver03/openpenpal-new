#!/bin/bash

# PostgreSQL 连接测试脚本

set -e

echo "🔍 PostgreSQL 连接测试"
echo "===================="

# 切换到后端目录
cd backend

# 检查环境文件
if [ ! -f .env ]; then
    echo "⚠️  未找到 .env 文件"
    echo "请先配置环境变量："
    echo "  cp .env.production .env"
    echo "  然后编辑 .env 文件设置数据库连接信息"
    exit 1
fi

# 显示当前配置
echo "当前数据库配置:"
grep -E "DATABASE_TYPE|DB_HOST|DB_PORT|DB_NAME|DB_USER" .env | grep -v PASSWORD || true

# 运行测试
echo ""
echo "运行连接测试..."
go run cmd/test-db/main.go

echo ""
echo "✅ 测试完成！"