#!/bin/bash

# SQLite 到 PostgreSQL 数据迁移脚本

set -e

echo "🚀 SQLite 到 PostgreSQL 数据迁移"
echo "================================"

# 检查参数
if [ $# -eq 0 ]; then
    SQLITE_FILE="backend/openpenpal.db"
else
    SQLITE_FILE="$1"
fi

# 检查 SQLite 文件
if [ ! -f "$SQLITE_FILE" ]; then
    echo "❌ SQLite 文件不存在: $SQLITE_FILE"
    exit 1
fi

echo "源数据库: $SQLITE_FILE"

# 切换到后端目录
cd backend

# 检查环境配置
if [ ! -f .env ]; then
    echo "❌ 未找到 .env 文件"
    echo "请先运行: cp .env.production .env"
    exit 1
fi

# 检查数据库类型
DB_TYPE=$(grep DATABASE_TYPE .env | cut -d '=' -f2)
if [ "$DB_TYPE" != "postgres" ]; then
    echo "❌ 请设置 DATABASE_TYPE=postgres"
    exit 1
fi

# 备份当前 PostgreSQL 数据（如果需要）
echo ""
echo "⚠️  警告: 此操作将迁移数据到 PostgreSQL"
echo "建议先备份现有的 PostgreSQL 数据"
read -p "是否继续？(y/N) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "操作已取消"
    exit 0
fi

# 运行迁移
echo ""
echo "开始迁移..."
go run cmd/migrate-data/main.go "../$SQLITE_FILE"

echo ""
echo "✨ 迁移完成！"
echo ""
echo "下一步："
echo "1. 验证数据: psql -U openpenpal -d openpenpal"
echo "2. 启动应用: go run main.go"
echo "3. 测试功能是否正常"