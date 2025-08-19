#!/bin/bash

# OpenPenPal 测试账号密码更新脚本
# 更新所有测试账号密码以符合安全要求

set -e

echo "======================================"
echo "OpenPenPal 测试账号密码更新工具"
echo "======================================"

# 检查是否提供了数据库连接信息
if [ -z "$DATABASE_URL" ]; then
    echo "⚠️  请设置环境变量 DATABASE_URL"
    echo "例如: export DATABASE_URL='postgres://username:password@localhost:5432/openpenpal'"
    exit 1
fi

echo "📝 正在更新测试账号密码..."
echo "   管理员: admin -> Admin123!"
echo "   用户: alice, bob -> Secret123!"
echo "   信使: courier_level[1-4] -> Secret123!"
echo "   其他测试账号 -> Secret123!"

# 执行 SQL 更新脚本
if psql "$DATABASE_URL" -f "$(dirname "$0")/update-test-account-passwords.sql"; then
    echo ""
    echo "✅ 密码更新成功!"
    echo ""
    echo "📋 新的测试账号信息:"
    echo "├── 管理员: admin / Admin123!"
    echo "├── 普通用户: alice, bob / Secret123!"
    echo "├── 信使账号: courier_level[1-4] / Secret123!"
    echo "└── 其他测试账号: Secret123!"
    echo ""
    echo "🔐 新密码安全特性:"
    echo "├── 长度: 9位字符"
    echo "├── 包含: 大写字母、小写字母、数字、符号"
    echo "└── 符合企业安全标准"
    echo ""
    echo "📖 详细信息请查看: docs/test-accounts/README.md"
    echo "======================================"
else
    echo ""
    echo "❌ 密码更新失败!"
    echo "请检查:"
    echo "├── 数据库连接是否正常"
    echo "├── DATABASE_URL 是否正确"
    echo "└── 用户是否有更新权限"
    echo "======================================"
    exit 1
fi