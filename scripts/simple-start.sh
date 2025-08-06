#!/bin/bash
# 兼容性包装器 - 重定向到新的启动系统
echo "⚠️ simple-start.js 已被替换为新的启动系统"
echo "正在启动简化Mock服务..."
exec "$(dirname "$0")/startup/start-simple-mock.sh" "$@"
