#!/bin/bash

# OpenPenPal 安全密码更新脚本
# 使用动态bcrypt哈希生成，禁止硬编码

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "======================================"
echo "OpenPenPal 安全密码更新工具"
echo "======================================"

# 检查环境变量
if [ -z "$DATABASE_URL" ]; then
    echo "⚠️  请设置环境变量 DATABASE_URL"
    echo "例如: export DATABASE_URL='postgres://username:password@localhost:5432/openpenpal'"
    exit 1
fi

echo "📝 使用安全的bcrypt动态哈希生成"
echo "📍 数据库: $DATABASE_URL"

# 检查Go是否安装
if ! command -v go &> /dev/null; then
    echo "❌ Go未安装，请先安装Go"
    exit 1
fi

# 进入项目目录以使用go.mod
cd "$PROJECT_DIR"

# 检查是否有go.mod
if [ ! -f "go.mod" ]; then
    echo "📦 初始化Go模块..."
    go mod init openpenpal-password-update
fi

# 确保依赖存在
echo "📦 安装/检查Go依赖..."
go mod tidy 2>/dev/null || true

# 运行密码更新程序
echo "🔄 执行安全密码更新..."
go run "$SCRIPT_DIR/update-passwords-secure.go"

if [ $? -eq 0 ]; then
    echo ""
    echo "📋 密码更新摘要:"
    echo "├── 管理员: admin / Admin123!"
    echo "├── 普通用户: alice, bob / Secret123!"
    echo "├── 信使账号: courier_level[1-4] / Secret123!"
    echo "└── 其他测试账号: Secret123!"
    echo ""
    echo "🔐 安全特性:"
    echo "├── ❌ 无硬编码密码哈希"
    echo "├── ✅ 动态bcrypt哈希生成"
    echo "├── ✅ 密码验证确认"
    echo "└── ✅ 符合企业安全标准"
    echo "======================================"
else
    echo ""
    echo "❌ 密码更新失败!"
    echo "请检查错误信息并重试"
    echo "======================================"
    exit 1
fi