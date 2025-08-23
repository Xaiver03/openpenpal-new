#!/bin/bash

# Test SiliconFlow API Integration
# Following CLAUDE.md principles

echo "🔍 Testing SiliconFlow API Integration..."
echo "========================================"

# Check if API key is set
if [ -z "$SILICONFLOW_API_KEY" ]; then
    echo "⚠️  Warning: SILICONFLOW_API_KEY environment variable not set"
    echo "   Please set it using: export SILICONFLOW_API_KEY='your-api-key'"
    exit 1
fi

# Base URL
BASE_URL="https://api.siliconflow.cn/v1"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "\n${YELLOW}1. Testing Direct SiliconFlow API Connection${NC}"
echo "------------------------------------------------"

# Test direct API call
echo "Testing chat completions endpoint..."
RESPONSE=$(curl -s -X POST "$BASE_URL/chat/completions" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $SILICONFLOW_API_KEY" \
  -d '{
    "model": "Qwen/Qwen2.5-7B-Instruct",
    "messages": [
      {
        "role": "user",
        "content": "Say hello in Chinese"
      }
    ],
    "temperature": 0.7,
    "max_tokens": 50
  }')

# Check if response contains error
if echo "$RESPONSE" | grep -q '"error"'; then
    echo -e "${RED}❌ Direct API test failed${NC}"
    echo "Error response: $RESPONSE"
else
    echo -e "${GREEN}✅ Direct API test successful${NC}"
    echo "Response: $RESPONSE" | jq -r '.choices[0].message.content' 2>/dev/null || echo "$RESPONSE"
fi

echo -e "\n${YELLOW}2. Testing Through Backend API${NC}"
echo "------------------------------------------------"

# First, check if backend is running
if ! curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo -e "${RED}❌ Backend is not running on port 8080${NC}"
    echo "   Please start the backend first"
    exit 1
fi

echo "Backend is running, testing AI endpoints..."

# Test provider status
echo -e "\n${YELLOW}Testing provider status...${NC}"
PROVIDER_STATUS=$(curl -s -X GET "http://localhost:8080/api/ai/providers/status")
echo "Provider Status: $PROVIDER_STATUS" | jq '.' 2>/dev/null || echo "$PROVIDER_STATUS"

# Test text generation with SiliconFlow
echo -e "\n${YELLOW}Testing text generation with SiliconFlow...${NC}"
GENERATION_RESPONSE=$(curl -s -X POST "http://localhost:8080/api/ai/generate" \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "写一首关于春天的短诗",
    "preferred_provider": "siliconflow",
    "max_tokens": 100,
    "temperature": 0.8
  }')

if echo "$GENERATION_RESPONSE" | grep -q '"success":true' || echo "$GENERATION_RESPONSE" | grep -q '"data"'; then
    echo -e "${GREEN}✅ Text generation test successful${NC}"
    echo "$GENERATION_RESPONSE" | jq -r '.data.content' 2>/dev/null || echo "$GENERATION_RESPONSE"
else
    echo -e "${RED}❌ Text generation test failed${NC}"
    echo "Response: $GENERATION_RESPONSE"
fi

# Test chat functionality
echo -e "\n${YELLOW}Testing chat functionality...${NC}"
CHAT_RESPONSE=$(curl -s -X POST "http://localhost:8080/api/ai/chat" \
  -H "Content-Type: application/json" \
  -d '{
    "messages": [
      {"role": "system", "content": "You are a helpful assistant."},
      {"role": "user", "content": "What is OpenPenPal?"}
    ],
    "preferred_provider": "siliconflow",
    "max_tokens": 150
  }')

if echo "$CHAT_RESPONSE" | grep -q '"success":true' || echo "$CHAT_RESPONSE" | grep -q '"data"'; then
    echo -e "${GREEN}✅ Chat test successful${NC}"
    echo "$CHAT_RESPONSE" | jq -r '.data.content' 2>/dev/null || echo "$CHAT_RESPONSE"
else
    echo -e "${RED}❌ Chat test failed${NC}"
    echo "Response: $CHAT_RESPONSE"
fi

# Test summarization
echo -e "\n${YELLOW}Testing summarization...${NC}"
SUMMARY_RESPONSE=$(curl -s -X POST "http://localhost:8080/api/ai/summarize" \
  -H "Content-Type: application/json" \
  -d '{
    "text": "OpenPenPal是一个创新的手写信件平台，它将传统的书信文化与现代技术相结合。用户可以在平台上写信、寄信，通过四级信使系统进行配送。平台还提供AI辅助写作、信件博物馆等特色功能，让每一封信都成为值得珍藏的记忆。",
    "preferred_provider": "siliconflow"
  }')

if echo "$SUMMARY_RESPONSE" | grep -q '"success":true' || echo "$SUMMARY_RESPONSE" | grep -q '"data"'; then
    echo -e "${GREEN}✅ Summarization test successful${NC}"
    echo "$SUMMARY_RESPONSE" | jq -r '.data.content' 2>/dev/null || echo "$SUMMARY_RESPONSE"
else
    echo -e "${RED}❌ Summarization test failed${NC}"
    echo "Response: $SUMMARY_RESPONSE"
fi

# Test translation
echo -e "\n${YELLOW}Testing translation...${NC}"
TRANSLATION_RESPONSE=$(curl -s -X POST "http://localhost:8080/api/ai/translate" \
  -H "Content-Type: application/json" \
  -d '{
    "text": "Hello, welcome to OpenPenPal!",
    "target_language": "中文",
    "preferred_provider": "siliconflow"
  }')

if echo "$TRANSLATION_RESPONSE" | grep -q '"success":true' || echo "$TRANSLATION_RESPONSE" | grep -q '"data"'; then
    echo -e "${GREEN}✅ Translation test successful${NC}"
    echo "$TRANSLATION_RESPONSE" | jq -r '.data.content' 2>/dev/null || echo "$TRANSLATION_RESPONSE"
else
    echo -e "${RED}❌ Translation test failed${NC}"
    echo "Response: $TRANSLATION_RESPONSE"
fi

echo -e "\n${YELLOW}3. Testing Available Models${NC}"
echo "------------------------------------------------"

# List available models
echo "Checking available models on SiliconFlow..."
MODELS_RESPONSE=$(curl -s -X GET "$BASE_URL/models" \
  -H "Authorization: Bearer $SILICONFLOW_API_KEY")

if echo "$MODELS_RESPONSE" | grep -q '"data"'; then
    echo -e "${GREEN}Available models:${NC}"
    echo "$MODELS_RESPONSE" | jq -r '.data[].id' 2>/dev/null | head -10
else
    echo -e "${YELLOW}Could not retrieve model list${NC}"
fi

echo -e "\n========================================"
echo "🎯 SiliconFlow API Testing Complete"
echo "========================================"