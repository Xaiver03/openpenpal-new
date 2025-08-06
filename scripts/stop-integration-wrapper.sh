#!/bin/bash
# 兼容性包装器 - 重定向到新的启动系统
echo "⚠️ stop-integration.sh 已被替换为新的启动系统"
echo "正在停止所有服务..."
exec "$(dirname "$0")/startup/stop-all.sh" "$@"
