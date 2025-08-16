#!/bin/bash

# OpenPenPal 日志管理脚本
# 用于清理膨胀的日志文件并防止未来的日志膨胀

set -e

PROJECT_ROOT="/Users/rocalight/同步空间/opplc/openpenpal"
LOGS_DIR="$PROJECT_ROOT/logs"
ARCHIVE_DIR="$LOGS_DIR/archive"
CURRENT_DATE=$(date +%Y%m%d_%H%M%S)

echo "🧹 OpenPenPal 日志管理工具"
echo "=========================="

# 创建归档目录
mkdir -p "$ARCHIVE_DIR/$CURRENT_DATE"

# 检查日志大小
echo ""
echo "📊 当前日志状态："
find "$PROJECT_ROOT" -name "*.log" -type f -exec ls -lh {} \; | sort -rh -k5 | head -10

echo ""
echo "📈 总日志目录大小："
du -sh "$LOGS_DIR"

# 备份并清理大日志文件
echo ""
echo "🗂️  处理大型日志文件..."

# 处理超大的backend.log
BACKEND_LOG="$LOGS_DIR/backend.log"
if [ -f "$BACKEND_LOG" ]; then
    BACKEND_SIZE=$(du -m "$BACKEND_LOG" | cut -f1)
    echo "Backend log 大小: ${BACKEND_SIZE}MB"
    
    if [ "$BACKEND_SIZE" -gt 100 ]; then
        echo "备份并压缩 backend.log..."
        
        # 保留最后1000行作为当前日志
        tail -1000 "$BACKEND_LOG" > "$BACKEND_LOG.tmp"
        
        # 压缩并归档原文件
        gzip -c "$BACKEND_LOG" > "$ARCHIVE_DIR/$CURRENT_DATE/backend-${CURRENT_DATE}.log.gz"
        
        # 替换为精简版本
        mv "$BACKEND_LOG.tmp" "$BACKEND_LOG"
        
        echo "✅ Backend log 已清理：${BACKEND_SIZE}MB -> $(du -m "$BACKEND_LOG" | cut -f1)MB"
    fi
fi

# 清理其他大型日志文件
find "$PROJECT_ROOT" -name "*.log" -type f -size +10M | while read -r logfile; do
    echo "处理大文件: $logfile"
    
    # 获取相对路径和文件名 (macOS compatible)
    REL_PATH=$(python3 -c "import os; print(os.path.relpath('$logfile', '$PROJECT_ROOT'))")
    FILENAME=$(basename "$logfile")
    
    # 保留最后500行
    tail -500 "$logfile" > "$logfile.tmp"
    
    # 压缩归档
    gzip -c "$logfile" > "$ARCHIVE_DIR/$CURRENT_DATE/${FILENAME%.*}-${CURRENT_DATE}.log.gz"
    
    # 替换
    mv "$logfile.tmp" "$logfile"
    
    echo "✅ 已清理: $REL_PATH"
done

# 创建日志轮转配置
echo ""
echo "⚙️  创建日志轮转配置..."

cat > "$PROJECT_ROOT/logrotate.conf" << 'EOF'
# OpenPenPal 日志轮转配置
# 使用方法: logrotate -f logrotate.conf

/Users/rocalight/同步空间/opplc/openpenpal/logs/*.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
    create 644 rocalight staff
    maxsize 50M
    postrotate
        # 重启服务以重新打开日志文件（如果需要）
        # killall -HUP openpenpal-backend || true
    endscript
}

/Users/rocalight/同步空间/opplc/openpenpal/backend/*.log {
    daily
    rotate 5
    compress
    delaycompress
    missingok
    notifempty
    create 644 rocalight staff
    maxsize 10M
}

/Users/rocalight/同步空间/opplc/openpenpal/frontend/*.log {
    daily
    rotate 5
    compress
    delaycompress
    missingok
    notifempty
    create 644 rocalight staff
    maxsize 10M
}
EOF

# 创建自动清理脚本
cat > "$PROJECT_ROOT/scripts/auto-log-cleanup.sh" << 'EOF'
#!/bin/bash

# 自动日志清理脚本 - 每小时运行
PROJECT_ROOT="/Users/rocalight/同步空间/opplc/openpenpal"

# 清理超过100MB的日志文件
find "$PROJECT_ROOT" -name "*.log" -type f -size +100M -exec truncate -s 0 {} \;

# 清理7天前的归档文件
find "$PROJECT_ROOT/logs/archive" -type f -mtime +7 -delete

# 清理空目录
find "$PROJECT_ROOT/logs/archive" -type d -empty -delete
EOF

chmod +x "$PROJECT_ROOT/scripts/auto-log-cleanup.sh"

echo ""
echo "🔍 分析日志问题..."

# 分析重复错误
if [ -f "$BACKEND_LOG" ]; then
    echo "最常见的错误模式："
    grep -o "Task [a-f0-9\-]* failed" "$BACKEND_LOG" | sort | uniq -c | sort -nr | head -5
    
    echo ""
    echo "重复失败的任务："
    grep "failed to generate AI reply: letter not found" "$BACKEND_LOG" | tail -5
fi

echo ""
echo "📋 建议操作："
echo "1. 检查 AI 回复任务调度器是否有死循环"
echo "2. 修复 'letter not found' 错误的根本原因"
echo "3. 添加任务失败的最大重试限制"
echo "4. 设置 cron 任务定期清理日志："
echo "   */30 * * * * $PROJECT_ROOT/scripts/auto-log-cleanup.sh"

echo ""
echo "🎯 日志轮转命令："
echo "   logrotate -f $PROJECT_ROOT/logrotate.conf"

echo ""
echo "✅ 日志管理完成！"

# 显示清理后的状态
echo ""
echo "📊 清理后的日志状态："
find "$PROJECT_ROOT" -name "*.log" -type f -exec ls -lh {} \; | sort -rh -k5 | head -5

echo ""
echo "💾 释放的磁盘空间："
du -sh "$LOGS_DIR"