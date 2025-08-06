#!/bin/bash

# OpenPenPal日志管理设置脚本
# 设置日志轮转、清理和监控

echo "📝 设置OpenPenPal日志管理系统..."
echo "=================================="

# 颜色定义
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[0;33m'
NC='\033[0m'

echo -e "${BLUE}📁 第1步: 创建日志目录结构${NC}"
echo "----------------------------------------"

# 创建详细的日志目录结构
mkdir -p logs/{app,access,error,performance,archive}
echo "✅ 创建日志目录结构"

# 设置正确的权限
chmod 755 logs
chmod 755 logs/*
echo "✅ 设置目录权限"

echo ""

echo -e "${BLUE}⚙️ 第2步: 创建日志轮转配置${NC}"
echo "----------------------------------------"

# 创建logrotate配置
cat > config/logrotate.conf << 'EOF'
# OpenPenPal日志轮转配置

# 全局配置
compress
delaycompress
missingok
notifempty
create 644 root root

# 应用日志
/Users/rocalight/同步空间/opplc/openpenpal/logs/app/*.log {
    daily
    rotate 30
    size 100M
    postrotate
        # 重启应用以重新打开日志文件 (如果需要)
        # systemctl reload openpenpal || true
    endscript
}

# 访问日志
/Users/rocalight/同步空间/opplc/openpenpal/logs/access/*.log {
    daily
    rotate 90
    size 500M
    compress
    delaycompress
}

# 错误日志
/Users/rocalight/同步空间/opplc/openpenpal/logs/error/*.log {
    daily
    rotate 60
    size 50M
    compress
    delaycompress
    copytruncate
}

# 性能日志
/Users/rocalight/同步空间/opplc/openpenpal/logs/performance/*.log {
    weekly
    rotate 12
    size 200M
    compress
    delaycompress
}
EOF

echo "✅ 创建logrotate配置"

echo ""

echo -e "${BLUE}🔧 第3步: 创建日志管理工具${NC}"
echo "----------------------------------------"

# 创建日志清理脚本
cat > scripts/cleanup-logs.sh << 'EOF'
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
EOF

chmod +x scripts/cleanup-logs.sh
echo "✅ 创建日志清理脚本"

# 创建日志分析脚本
cat > scripts/analyze-logs.sh << 'EOF'
#!/bin/bash

# OpenPenPal日志分析脚本
# 分析日志文件并生成统计报告

LOGS_DIR="logs"
REPORT_FILE="logs/analysis_report_$(date +%Y%m%d_%H%M%S).txt"

echo "📊 开始日志分析..."
echo "=====================" > "$REPORT_FILE"
echo "OpenPenPal日志分析报告" >> "$REPORT_FILE"
echo "生成时间: $(date)" >> "$REPORT_FILE"
echo "=====================" >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"

# 分析错误日志
if [[ -f "$LOGS_DIR/error.log" ]]; then
    echo "🔍 分析错误日志..."
    echo "错误统计:" >> "$REPORT_FILE"
    echo "--------" >> "$REPORT_FILE"
    
    # 错误级别统计
    grep -i "error" "$LOGS_DIR/error.log" 2>/dev/null | wc -l | xargs -I {} echo "ERROR级别: {} 条" >> "$REPORT_FILE"
    grep -i "warn" "$LOGS_DIR"/*.log 2>/dev/null | wc -l | xargs -I {} echo "WARN级别: {} 条" >> "$REPORT_FILE"
    grep -i "fatal" "$LOGS_DIR"/*.log 2>/dev/null | wc -l | xargs -I {} echo "FATAL级别: {} 条" >> "$REPORT_FILE"
    
    echo "" >> "$REPORT_FILE"
    
    # 最频繁的错误
    echo "最频繁错误 (Top 5):" >> "$REPORT_FILE"
    grep -i "error" "$LOGS_DIR"/*.log 2>/dev/null | cut -d':' -f3- | sort | uniq -c | sort -nr | head -5 >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"
fi

# 分析访问日志
if [[ -f "$LOGS_DIR/access.log" ]]; then
    echo "🌐 分析访问日志..."
    echo "访问统计:" >> "$REPORT_FILE"
    echo "--------" >> "$REPORT_FILE"
    
    # 总请求数
    wc -l "$LOGS_DIR/access.log" 2>/dev/null | awk '{print "总请求数: " $1}' >> "$REPORT_FILE"
    
    # HTTP状态码统计
    echo "HTTP状态码分布:" >> "$REPORT_FILE"
    awk '{print $9}' "$LOGS_DIR/access.log" 2>/dev/null | sort | uniq -c | sort -nr >> "$REPORT_FILE"
    
    echo "" >> "$REPORT_FILE"
fi

# 分析性能日志
echo "⚡ 系统性能概览:" >> "$REPORT_FILE"
echo "-------------" >> "$REPORT_FILE"

# 磁盘使用情况
echo "日志目录磁盘使用:" >> "$REPORT_FILE"
du -sh "$LOGS_DIR"/* 2>/dev/null >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"

# 生成建议
echo "🎯 优化建议:" >> "$REPORT_FILE"
echo "----------" >> "$REPORT_FILE"

# 检查日志文件大小
large_files=$(find "$LOGS_DIR" -name "*.log" -size +100M 2>/dev/null)
if [[ -n "$large_files" ]]; then
    echo "• 以下日志文件过大，建议清理或轮转:" >> "$REPORT_FILE"
    echo "$large_files" >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"
fi

# 检查错误率
error_count=$(grep -i "error" "$LOGS_DIR"/*.log 2>/dev/null | wc -l)
if [[ $error_count -gt 100 ]]; then
    echo "• 错误日志较多($error_count条)，建议检查应用状态" >> "$REPORT_FILE"
fi

echo "✅ 日志分析完成，报告已保存到: $REPORT_FILE"
cat "$REPORT_FILE"
EOF

chmod +x scripts/analyze-logs.sh
echo "✅ 创建日志分析脚本"

echo ""

echo -e "${YELLOW}📋 第4步: 创建日志监控配置${NC}"
echo "----------------------------------------"

# 创建日志监控配置
cat > config/log-monitoring.yml << 'EOF'
# OpenPenPal日志监控配置
# 用于Prometheus + Grafana监控

# 日志收集器配置 (Promtail)
server:
  http_listen_port: 9080
  grpc_listen_port: 0

positions:
  filename: /tmp/positions.yaml

clients:
  - url: http://localhost:3100/loki/api/v1/push

scrape_configs:
  # 应用日志
  - job_name: openpenpal-app
    static_configs:
      - targets:
          - localhost
        labels:
          job: openpenpal
          component: app
          __path__: /Users/rocalight/同步空间/opplc/openpenpal/logs/app/*.log

  # 访问日志
  - job_name: openpenpal-access
    static_configs:
      - targets:
          - localhost
        labels:
          job: openpenpal
          component: access
          __path__: /Users/rocalight/同步空间/opplc/openpenpal/logs/access/*.log

  # 错误日志
  - job_name: openpenpal-error
    static_configs:
      - targets:
          - localhost
        labels:
          job: openpenpal
          component: error
          __path__: /Users/rocalight/同步空间/opplc/openpenpal/logs/error/*.log
    pipeline_stages:
      - match:
          selector: '{component="error"}'
          stages:
            - regex:
                expression: '(?P<timestamp>\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}) \[(?P<level>\w+)\] (?P<message>.*)'
            - timestamp:
                source: timestamp
                format: '2006-01-02 15:04:05'
EOF

echo "✅ 创建日志监控配置"

echo ""

echo -e "${GREEN}🎊 日志管理系统设置完成${NC}"
echo "=================================="

echo "📋 已创建的文件:"
echo "   • config/logrotate.conf - 日志轮转配置"
echo "   • config/log-monitoring.yml - 日志监控配置"
echo "   • scripts/cleanup-logs.sh - 日志清理脚本"
echo "   • scripts/analyze-logs.sh - 日志分析脚本"

echo ""
echo -e "${YELLOW}📋 使用说明:${NC}"
echo "1. 运行日志清理: ./scripts/cleanup-logs.sh"
echo "2. 分析日志统计: ./scripts/analyze-logs.sh"
echo "3. 设置定时任务: crontab -e"
echo "   添加: 0 2 * * * /path/to/cleanup-logs.sh"

echo ""
echo -e "${GREEN}✨ 日志管理系统配置完成！${NC}"