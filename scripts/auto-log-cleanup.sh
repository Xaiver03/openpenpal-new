#!/bin/bash

# 自动日志清理脚本 - 每小时运行
PROJECT_ROOT="/Users/rocalight/同步空间/opplc/openpenpal"

# 清理超过100MB的日志文件
find "$PROJECT_ROOT" -name "*.log" -type f -size +100M -exec truncate -s 0 {} \;

# 清理7天前的归档文件
find "$PROJECT_ROOT/logs/archive" -type f -mtime +7 -delete

# 清理空目录
find "$PROJECT_ROOT/logs/archive" -type d -empty -delete
