#!/bin/bash

# 更新所有 API 文件的导入，使用统一的 API 客户端
# 这将启用自动的 snake_case/camelCase 转换

API_DIR="src/lib/api"

echo "🔄 更新 API 文件导入..."

# 需要更新的文件列表（排除已经更新的和特殊文件）
FILES=(
  "ai.ts"
  "batch-management.ts"
  "barcode-binding.ts"
  "comment.ts"
  "courier.ts"
  "courier-growth.ts"
  "credit-limits.ts"
  "credit-shop.ts"
  "follow.ts"
  "moderation.ts"
  "museum.ts"
  "museum-fixed.ts"
  "ocr.ts"
  "operation-log.ts"
  "privacy.ts"
  "qr-scan.ts"
  "scheduler.ts"
  "shop.ts"
  "user.ts"
)

# 更新每个文件
for file in "${FILES[@]}"; do
  filepath="$API_DIR/$file"
  if [ -f "$filepath" ]; then
    echo "📝 更新 $file..."
    
    # 替换 apiClient 导入
    # 从 '../api-client' 或 '@/lib/api-client' 改为从 index 导入
    sed -i '' "s|import { apiClient.*} from '\.\./api-client'|import { apiClient } from './'|g" "$filepath"
    sed -i '' "s|import { apiClient.*} from '@/lib/api-client'|import { apiClient } from '@/lib/api'|g" "$filepath"
    
    # 处理只导入 apiClient 的情况
    sed -i '' "s|import { apiClient } from '\.\./api-client'|import { apiClient } from './'|g" "$filepath"
    sed -i '' "s|import { apiClient } from '@/lib/api-client'|import { apiClient } from '@/lib/api'|g" "$filepath"
  fi
done

echo "✅ API 导入更新完成！"
echo ""
echo "📌 注意事项："
echo "1. 现在所有 API 调用都会自动进行 snake_case/camelCase 转换"
echo "2. 如果某些 API 不需要转换，可以使用 rawApiClient"
echo "3. 请运行 npm run type-check 验证类型"