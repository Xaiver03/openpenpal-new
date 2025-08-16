#!/bin/bash

# 测试文件查找助手
# 帮助找到重新组织后的测试文件位置

echo "🔍 OpenPenPal 测试文件查找助手"
echo "================================"

if [ $# -eq 0 ]; then
    echo "用法: $0 <测试文件名>"
    echo "示例: $0 test-ai-endpoints.sh"
    echo ""
    echo "可用的测试目录:"
    echo "📁 tests/manual/ai/      - AI相关测试"
    echo "📁 tests/manual/auth/    - 认证相关测试"
    echo "📁 tests/manual/courier/ - 信使系统测试"
    echo "📁 tests/manual/admin/   - 管理后台测试"
    echo "📁 tests/manual/security/ - 安全相关测试"
    echo "📁 tests/scripts/        - 其他测试脚本"
    echo "📁 test-kimi/            - 集成测试"
    exit 1
fi

SEARCH_FILE="$1"

echo "正在查找: $SEARCH_FILE"
echo ""

# 在新的测试目录中搜索
FOUND_FILES=$(find tests/ -name "*$SEARCH_FILE*" 2>/dev/null)

if [ -n "$FOUND_FILES" ]; then
    echo "✅ 找到匹配文件:"
    echo "$FOUND_FILES" | while read file; do
        echo "   📍 $file"
    done
else
    echo "❌ 未找到匹配的文件"
    echo ""
    echo "💡 建议:"
    echo "1. 检查文件名是否正确"
    echo "2. 文件可能在以下位置:"
    echo "   - tests/manual/     (按功能分类的手动测试)"
    echo "   - tests/scripts/    (通用测试脚本)"
    echo "   - test-kimi/        (集成测试)"
fi

echo ""
echo "📋 所有测试文件列表:"
echo "===================="
find tests/ -name "test-*" 2>/dev/null | head -20
if [ $(find tests/ -name "test-*" 2>/dev/null | wc -l) -gt 20 ]; then
    echo "   ... (还有更多文件)"
fi