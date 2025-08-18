#!/bin/bash

# 修复 credit-limits.ts 中的类型断言问题

FILE="src/lib/api/credit-limits.ts"

echo "🔄 修复 credit-limits.ts 类型断言..."

# 备份原文件
cp "$FILE" "$FILE.bak"

# 修复 batchUpdateRules
sed -i '' 's/export async function batchUpdateRules(data: {/export async function batchUpdateRules(data: {/' "$FILE"
sed -i '' '/export async function batchUpdateRules/,/^}$/ {
  s/const response = await apiClient.put.*$/const response = await apiClient.put<{ message: string; updated_count: number; errors?: string[] }>('"'"'\/admin\/credits\/limit-rules\/batch'"'"', data)/
  s/return response.data$/return response.data as { message: string; updated_count: number; errors?: string[] }/
}' "$FILE"

# 修复 getRiskUsers
sed -i '' '/export async function getRiskUsers/,/^}$/ {
  s/const response = await apiClient.get(url)$/const response = await apiClient.get<{ users: CreditRiskUser[]; total: number; page: number; limit: number }>(url)/
  s/return response.data$/return response.data as { users: CreditRiskUser[]; total: number; page: number; limit: number }/
}' "$FILE"

# 修复其他简单的返回类型
sed -i '' 's/return response.data$/return response.data as any/g' "$FILE"

echo "✅ 类型断言修复完成！"
echo "📌 注意：使用了 'as any' 作为临时解决方案，建议后续定义更精确的类型"