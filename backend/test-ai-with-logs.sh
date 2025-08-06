#!/bin/bash

# Start backend with full logging
echo "Starting backend with AI logging..."
cd backend
go run main.go 2>&1 | grep -E "(Moonshot|AI|moonshot|GetInspiration|callMoonshot)" > ai-debug.log &
BACKEND_PID=$!

# Wait for backend to start
echo "Waiting for backend to start..."
sleep 5

# Test AI endpoint
echo -e "\n\nTesting AI endpoint..."
curl -X POST http://localhost:8080/api/v1/ai/inspiration \
  -H "Content-Type: application/json" \
  -d '{"theme": "友谊", "count": 1}' | jq .

# Wait a bit to capture all logs
sleep 2

# Show the logs
echo -e "\n\nAI Debug Logs:"
cat ai-debug.log

# Clean up
kill $BACKEND_PID 2>/dev/null