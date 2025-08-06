#!/bin/bash

# 快速修复response包导入问题的脚本

echo "🔧 修复response包导入问题..."

backend_dir="/Users/rocalight/同步空间/opplc/openpenpal/backend"

# 需要修复的文件列表
files=(
    "internal/handlers/letter_handler.go"
    "internal/handlers/courier_handler.go" 
    "internal/handlers/user_handler.go"
    "internal/handlers/letter_handler_envelope.go"
    "internal/handlers/envelope_handler.go"
    "internal/handlers/credit_handler.go"
)

for file in "${files[@]}"; do
    file_path="$backend_dir/$file"
    if [ -f "$file_path" ]; then
        echo "处理文件: $file"
        
        # 备份原文件
        cp "$file_path" "$file_path.backup"
        
        # 移除problematic import
        sed -i '' '/shared\/pkg\/response/d' "$file_path"
        
        # 确保utils import存在
        if ! grep -q "openpenpal-backend/internal/utils" "$file_path"; then
            # 在其他import之后添加utils import
            sed -i '' '/import (/a\
	"openpenpal-backend/internal/utils"
' "$file_path"
        fi
        
        echo "✅ $file 处理完成"
    else
        echo "⚠️  文件不存在: $file_path"
    fi
done

echo "🎉 修复完成！"
echo "💡 如果出现问题，可以用 .backup 文件恢复"