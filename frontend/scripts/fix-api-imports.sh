#!/bin/bash

# 修复 apiClient 导入路径问题

echo "🔄 修复 apiClient 导入路径..."

# 修复从 @/lib/api 导入的情况
echo "📝 修复 @/lib/api 导入..."
find src/ -name "*.ts" -o -name "*.tsx" | xargs sed -i '' 's|import { apiClient } from '\''@/lib/api'\''|import { apiClient } from '\''@/lib/api-client'\''|g'

# 修复从 ./ 导入的情况（在 lib/api 目录内）
echo "📝 修复相对路径导入..."
find src/lib/api -name "*.ts" | xargs sed -i '' 's|import { apiClient } from '\''\.\/'\''|import { apiClient } from '\''../api-client-enhanced'\''|g'

# 修复 admin/credit-shop/page.tsx 的特殊情况
echo "📝 修复 admin/credit-shop/page.tsx..."
sed -i '' 's|import { apiClient } from '\''@/lib/api'\''|import { apiClient } from '\''@/lib/api-client-enhanced'\''|g' src/app/admin/credit-shop/page.tsx

echo "✅ 导入路径修复完成！"