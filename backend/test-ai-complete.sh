#!/bin/bash

echo "🧪 Testing Complete AI Functionality"
echo "===================================="

# Test 1: Direct API call without auth
echo -e "\n1️⃣ Testing AI Inspiration API (no auth)..."
curl -s -X POST "http://localhost:8080/api/v1/ai/inspiration" \
  -H "Content-Type: application/json" \
  -d '{"theme":"日常生活","count":3}' | jq '.'

# Test 2: Login and test with auth
echo -e "\n2️⃣ Logging in as alice..."
LOGIN_RESPONSE=$(curl -s -X POST "http://localhost:8080/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","password":"secret"}')

TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.data.token')
echo "Token obtained: ${TOKEN:0:20}..."

# Test 3: Test AI with auth
echo -e "\n3️⃣ Testing AI Inspiration API (with auth)..."
curl -s -X POST "http://localhost:8080/api/v1/ai/inspiration" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"theme":"科技与未来","count":2}' | jq '.'

# Test 4: Test different AI endpoints
echo -e "\n4️⃣ Testing AI Personas API..."
curl -s -X GET "http://localhost:8080/api/v1/ai/personas" | jq '.'

echo -e "\n✅ AI Testing Complete!"