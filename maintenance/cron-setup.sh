#!/bin/bash

# OpenPenPal 日志维护 Cron 设置
echo "设置日志维护定时任务..."

# 检查当前用户的crontab
echo "当前crontab内容："
crontab -l 2>/dev/null || echo "没有现有的crontab"

echo ""
echo "建议添加以下定时任务："
echo "# OpenPenPal 日志维护"
echo "*/30 * * * * /path/to/openpenpal/scripts/auto-log-cleanup.sh"
echo "0 2 * * * /path/to/openpenpal/scripts/log-management.sh"
echo "0 0 * * 0 /path/to/openpenpal/scripts/weekly-log-archive.sh"

echo ""
echo "要添加这些任务，请运行："
echo "crontab -e"
echo "然后添加上述行到文件中"
