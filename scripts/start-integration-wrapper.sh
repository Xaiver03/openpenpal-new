#!/bin/bash
# 兼容性包装器 - 重定向到新的启动系统
echo "⚠️ start-integration.sh 已被替换为新的启动系统"
echo "正在启动集成环境..."
exec "$(dirname "$0")/startup/start-integration.sh" "$@"
