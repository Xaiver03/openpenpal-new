#!/bin/bash

# Start backend without running migrations
# 启动后端但跳过数据库迁移

echo "Starting backend service without migrations..."

# Set environment variables
export DB_TYPE=postgres
export DATABASE_TYPE=postgres
export DATABASE_URL="postgresql://openpenpal_user:password@localhost:5432/openpenpal"
export SKIP_DB_MIGRATION=true

# Change to script directory
cd "$(dirname "$0")"

# Build and run
echo "Building backend..."
go build -o openpenpal-backend main.go

if [ $? -eq 0 ]; then
    echo "Starting server on port 8080..."
    ./openpenpal-backend serve
else
    echo "Build failed!"
    exit 1
fi