#!/bin/bash

# 修复所有使用 user.token 的地方，改为使用 apiClient

echo "🔄 修复 user.token 使用问题..."

# 查找所有使用 user.token 的文件
FILES=$(grep -r "user\.token" src/ --include="*.tsx" --include="*.ts" -l 2>/dev/null)

if [ -z "$FILES" ]; then
  echo "✅ 没有找到使用 user.token 的文件"
  exit 0
fi

echo "📋 找到以下文件使用了 user.token:"
echo "$FILES"
echo ""

# 对于 credit-shop 页面，替换所有 fetch 调用
CREDIT_SHOP="src/app/admin/credit-shop/page.tsx"
if [ -f "$CREDIT_SHOP" ]; then
  echo "📝 更新 $CREDIT_SHOP..."
  
  # 替换所有带 Authorization header 的 fetch 调用
  sed -i '' 's/await fetch(\(.*\), {[^}]*headers:[^}]*Authorization.*Bearer.*user\.token.*}[^}]*})/await apiClient.get(\1)/g' "$CREDIT_SHOP"
  sed -i '' 's/await fetch(\(.*\), {[^}]*method:[[:space:]]*["'\'']*POST["'\'']*.*headers:[^}]*Authorization.*Bearer.*user\.token.*}[^}]*})/await apiClient.post(\1)/g' "$CREDIT_SHOP"
  sed -i '' 's/await fetch(\(.*\), {[^}]*method:[[:space:]]*["'\'']*PUT["'\'']*.*headers:[^}]*Authorization.*Bearer.*user\.token.*}[^}]*})/await apiClient.put(\1)/g' "$CREDIT_SHOP"
  sed -i '' 's/await fetch(\(.*\), {[^}]*method:[[:space:]]*["'\'']*DELETE["'\'']*.*headers:[^}]*Authorization.*Bearer.*user\.token.*}[^}]*})/await apiClient.delete(\1)/g' "$CREDIT_SHOP"
  
  # 替换 response.ok 为 response.success
  sed -i '' 's/response\.ok/response.success/g' "$CREDIT_SHOP"
  
  # 替换 await response.json() 为 response.data
  sed -i '' 's/await response\.json()/response.data/g' "$CREDIT_SHOP"
fi

echo ""
echo "📌 注意事项："
echo "1. apiClient 会自动处理认证 token"
echo "2. 不再需要手动设置 Authorization header"
echo "3. 响应格式为 { success, data, message }"
echo ""
echo "✅ 修复完成！请手动检查以下更复杂的情况："
grep -n "user\.token" src/ --include="*.tsx" --include="*.ts" -r 2>/dev/null || echo "没有剩余的 user.token 使用"