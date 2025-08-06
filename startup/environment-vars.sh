#!/bin/bash

# OpenPenPal 环境变量配置
# 所有启动脚本共享的环境变量设置

# 项目信息
export PROJECT_NAME="OpenPenPal"
export PROJECT_VERSION="1.0.0"
export PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# 代理设置 - 确保本地服务不使用代理
export NO_PROXY="localhost,127.0.0.1,*.local"
export no_proxy="localhost,127.0.0.1,*.local"

# 服务端口配置
export FRONTEND_PORT=3000
export ADMIN_FRONTEND_PORT=3001
export GATEWAY_PORT=8000
export BACKEND_PORT=8080
export WRITE_SERVICE_PORT=8001
export COURIER_SERVICE_PORT=8002
export ADMIN_SERVICE_PORT=8003
export OCR_SERVICE_PORT=8004

# API 地址配置
export API_BASE_URL="http://localhost:${GATEWAY_PORT}"
export FRONTEND_URL="http://localhost:${FRONTEND_PORT}"
export ADMIN_FRONTEND_URL="http://localhost:${ADMIN_FRONTEND_PORT}"

# 服务间通信地址
export WRITE_SERVICE_URL="http://localhost:${WRITE_SERVICE_PORT}"
export COURIER_SERVICE_URL="http://localhost:${COURIER_SERVICE_PORT}"
export ADMIN_SERVICE_URL="http://localhost:${ADMIN_SERVICE_PORT}"
export OCR_SERVICE_URL="http://localhost:${OCR_SERVICE_PORT}"

# 安全配置
export JWT_SECRET="openpenpal-super-secret-jwt-key-for-integration"
export CORS_ORIGIN="http://localhost:${FRONTEND_PORT},http://localhost:${ADMIN_FRONTEND_PORT}"
export SESSION_SECRET="openpenpal-session-secret-key"

# 数据库配置
# 检查是否已经设置了数据库类型，如果没有则使用默认值
if [ -z "$DATABASE_TYPE" ]; then
    export DATABASE_TYPE="${DATABASE_TYPE:-sqlite}"
fi

# 根据数据库类型设置相应的配置
if [ "$DATABASE_TYPE" = "postgres" ] || [ "$DATABASE_TYPE" = "postgresql" ]; then
    # PostgreSQL 配置
    export DATABASE_URL="${DATABASE_URL:-}"
    export DATABASE_NAME="${DATABASE_NAME:-openpenpal}"
    export DB_HOST="${DB_HOST:-localhost}"
    export DB_PORT="${DB_PORT:-5432}"
    export DB_USER="${DB_USER:-openpenpal}"
    export DB_PASSWORD="${DB_PASSWORD:-openpenpal123}"
    export DB_SSLMODE="${DB_SSLMODE:-disable}"
else
    # SQLite 配置（默认）
    export DATABASE_URL="${DATABASE_URL:-./openpenpal.db}"
    export DATABASE_NAME="${DATABASE_NAME:-openpenpal}"
fi

# Redis配置（Mock环境）
export REDIS_HOST="localhost"
export REDIS_PORT="6379"
export REDIS_PASSWORD=""

# 日志配置
export LOG_LEVEL="info"
export LOG_FORMAT="combined"
export LOG_DIR="${PROJECT_ROOT}/logs"
export DEBUG="false"

# 开发环境特定配置
if [ "$NODE_ENV" = "development" ]; then
    export DEBUG="true"
    export LOG_LEVEL="debug"
    export WEBPACK_DEV_SERVER_HOST="0.0.0.0"
    export WEBPACK_DEV_SERVER_PORT="3000"
fi

# 生产环境特定配置
if [ "$NODE_ENV" = "production" ]; then
    export DEBUG="false"
    export LOG_LEVEL="warn"
    export ENABLE_COMPRESSION="true"
    export ENABLE_SECURITY_HEADERS="true"
fi

# 测试环境特定配置
if [ "$NODE_ENV" = "test" ]; then
    export DEBUG="true"
    export LOG_LEVEL="debug"
    export API_BASE_URL="http://localhost:8900"
    export FRONTEND_PORT=3900
    export GATEWAY_PORT=8900
fi

# 文件上传配置
export UPLOAD_MAX_SIZE="10MB"
export UPLOAD_ALLOWED_TYPES="jpg,jpeg,png,gif,pdf,doc,docx"
export UPLOAD_DIR="${PROJECT_ROOT}/uploads"

# 邮件配置（Mock环境）
export MAIL_HOST="smtp.mock.com"
export MAIL_PORT="587"
export MAIL_USER="noreply@openpenpal.com"
export MAIL_PASSWORD="mock-password"
export MAIL_FROM="OpenPenPal <noreply@openpenpal.com>"

# OCR服务配置
export OCR_PROVIDER="mock"
export OCR_API_KEY="mock-api-key"
export OCR_MAX_IMAGE_SIZE="5MB"

# 缓存配置
export CACHE_TYPE="memory"
export CACHE_TTL="3600"
export CACHE_MAX_SIZE="100MB"

# 监控配置
export MONITORING_ENABLED="true"
export HEALTH_CHECK_INTERVAL="30"
export PERFORMANCE_METRICS="true"
export ERROR_TRACKING="true"

# 安全配置
export RATE_LIMIT_WINDOW="15"
export RATE_LIMIT_MAX="100"
export BCRYPT_ROUNDS="12"
export PASSWORD_MIN_LENGTH="6"

# 业务配置
export MAX_LETTER_LENGTH="2000"
export MAX_LETTERS_PER_DAY="10"
export DELIVERY_TIMEOUT_HOURS="72"
export AUTO_MATCH_ENABLED="true"

# 创建必需的目录
create_directories() {
    mkdir -p "${LOG_DIR}"
    mkdir -p "${UPLOAD_DIR}"
    mkdir -p "${PROJECT_ROOT}/tmp"
    mkdir -p "${PROJECT_ROOT}/cache"
}

# 验证环境变量
validate_environment() {
    local required_vars=(
        "PROJECT_ROOT"
        "API_BASE_URL"
        "FRONTEND_URL"
        "JWT_SECRET"
    )
    
    for var in "${required_vars[@]}"; do
        if [ -z "${!var}" ]; then
            echo "错误: 必需的环境变量 $var 未设置"
            return 1
        fi
    done
    
    return 0
}

# 显示环境配置
show_environment() {
    echo "OpenPenPal 环境配置:"
    echo "===================="
    echo "项目名称: $PROJECT_NAME"
    echo "项目版本: $PROJECT_VERSION"
    echo "项目根目录: $PROJECT_ROOT"
    echo "运行环境: ${NODE_ENV:-development}"
    echo "前端地址: $FRONTEND_URL"
    echo "API地址: $API_BASE_URL"
    echo "日志级别: $LOG_LEVEL"
    echo "调试模式: $DEBUG"
    echo "===================="
}

# 导出函数
export -f create_directories
export -f validate_environment
export -f show_environment

# 如果直接运行此脚本，显示环境配置
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    show_environment
fi