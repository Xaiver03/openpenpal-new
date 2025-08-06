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
