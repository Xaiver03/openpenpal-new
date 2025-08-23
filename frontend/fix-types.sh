#!/bin/bash

# 修复TypeScript API响应类型错误
echo "🔧 修复 TypeScript API 响应类型错误..."

# 修复payment组件
echo "修复payment-gateway.tsx..."
sed -i '' 's/response\.data/((response as any)?.data?.data || (response as any)?.data)/g' src/components/payment/payment-gateway.tsx

# 修复delivery-guide相关页面
echo "修复delivery-guide页面..."
find src/app/\(main\)/delivery-guide -name "*.tsx" -exec sed -i '' 's/response\.data/((response as any)?.data?.data || (response as any)?.data)/g' {} \;

# 修复orders页面
echo "修复orders页面..."
find src/app/\(main\)/orders -name "*.tsx" -exec sed -i '' 's/response\.data/((response as any)?.data?.data || (response as any)?.data)/g' {} \;

# 修复admin barcode页面
echo "修复admin barcode页面..."
find src/app/admin/barcodes -name "*.tsx" -exec sed -i '' 's/logsResponse\.data/((logsResponse as any)?.data?.data || (logsResponse as any)?.data)/g' {} \;
find src/app/admin/barcodes -name "*.tsx" -exec sed -i '' 's/response\.data/((response as any)?.data?.data || (response as any)?.data)/g' {} \;

echo "✅ 类型修复完成"