#!/bin/bash

echo "Starting backend with Moonshot AI configuration..."

# Export environment variables
export MOONSHOT_API_KEY="sk-wQUdnk3pwUdEkAGQl85krQcmAE36eLl8DuKj1hLZZrjzuxvV"
export AI_PROVIDER="moonshot"
export DATABASE_URL="postgres://rocalight:password@localhost:5432/openpenpal"
export DB_TYPE="postgres"
export JWT_SECRET="KY6QtIecDZocllQSYoqyTkYx8AuKDkpA7RfondzVB2Y="

echo "Environment variables set:"
echo "  AI_PROVIDER=$AI_PROVIDER"
echo "  MOONSHOT_API_KEY=sk-wQU...uxvV (masked)"

# Kill any existing process
echo "Stopping any existing backend process..."
pkill -f "openpenpal" 2>/dev/null || true
sleep 2

# Start the backend
echo "Starting backend on port 8080..."
cd /Users/rocalight/同步空间/opplc/openpenpal/backend
./openpenpal