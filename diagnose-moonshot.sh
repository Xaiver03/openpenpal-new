#!/bin/bash

echo "=== Moonshot API Diagnostics ==="
echo "================================"

# Check if backend is running
echo -e "\n1. Checking backend health..."
curl -s http://localhost:8080/api/v1/health | jq . || echo "Backend not running!"

# Get auth token
echo -e "\n2. Getting auth token..."
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","password":"secret"}' | jq -r '.data.token')

if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
  echo "Failed to get auth token!"
  exit 1
fi

echo "Token obtained: ${TOKEN:0:20}..."

# Test AI inspiration endpoint
echo -e "\n3. Testing AI inspiration endpoint..."
echo "Request:"
echo '{"theme":"日常生活","count":1}'

echo -e "\nResponse:"
curl -s -X POST http://localhost:8080/api/v1/ai/inspiration \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"theme":"日常生活","count":1}' | jq .

echo -e "\n================================"
echo "Check the backend console for detailed error logs!"
echo "Look for lines starting with:"
echo "  ❌ [AIHandler] GetInspirationWithLimit error:"
echo "  🌙 [Moonshot] ..."
echo "  ❌ [Moonshot] ..."