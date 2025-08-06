#!/bin/bash

# OpenPenPal日志清理脚本
# 自动清理过期日志并归档重要日志

LOGS_DIR="logs"
ARCHIVE_DIR="logs/archive"
DAYS_TO_KEEP=30
ARCHIVE_DAYS=90

echo "🧹 开始日志清理..."

# 创建归档目录
mkdir -p "$ARCHIVE_DIR"

# 统计变量
CLEANED_FILES=0
ARCHIVED_FILES=0

# 清理过期的普通日志
echo "🗑️  清理 ${DAYS_TO_KEEP} 天前的日志文件..."
find "$LOGS_DIR" -name "*.log" -type f -mtime +${DAYS_TO_KEEP} | while read file; do
    if [[ "$file" != *"/archive/"* ]]; then
        echo "   删除: $file"
        rm -f "$file"
        CLEANED_FILES=$((CLEANED_FILES + 1))
    fi
done

# 归档重要日志
echo "📦 归档重要日志文件..."
find "$LOGS_DIR" -name "error*.log" -o -name "crash*.log" -o -name "security*.log" | while read file; do
    if [[ "$file" != *"/archive/"* ]]; then
        filename=$(basename "$file")
        timestamp=$(date +%Y%m%d_%H%M%S)
        archived_name="${timestamp}_${filename}"
        
        gzip -c "$file" > "$ARCHIVE_DIR/$archived_name.gz"
        echo "   归档: $file -> $archived_name.gz"
        ARCHIVED_FILES=$((ARCHIVED_FILES + 1))
    fi
done

# 清理过期归档
echo "🗑️  清理 ${ARCHIVE_DAYS} 天前的归档文件..."
find "$ARCHIVE_DIR" -name "*.gz" -type f -mtime +${ARCHIVE_DAYS} -delete

echo "✅ 日志清理完成"
echo "   清理文件: $CLEANED_FILES 个"
echo "   归档文件: $ARCHIVED_FILES 个"
