#!/bin/bash
# 启动时检查日志
cd "$(dirname "$0")/.."
./scripts/log-monitor.sh check
