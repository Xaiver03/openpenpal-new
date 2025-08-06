#!/bin/bash

echo "Simple AI Test"
echo "=============="

# Login
echo -e "\n1. Login..."
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","password":"secret"}' | jq -r '.data.token')

echo "Token: ${TOKEN:0:30}..."

# Test AI
echo -e "\n2. Test AI inspiration..."
curl -X POST http://localhost:8080/api/v1/ai/inspiration \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"theme":"test","count":1}' \
  -w "\n\nHTTP Status: %{http_code}\n"

echo -e "\n3. Check if response contains 'fallback'..."
RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/ai/inspiration \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"theme":"test","count":1}')

if echo "$RESPONSE" | grep -q "fallback"; then
  echo "❌ Still using fallback - Moonshot API not working"
else
  echo "✅ Real AI response - Moonshot API is working!"
fi

echo -e "\n==============\nIMPORTANT: Check the terminal where backend is running for error logs!"