#!/bin/bash

# Test Moonshot API directly
# This script tests the Moonshot API without going through the backend

echo "Testing Moonshot API directly..."
echo "================================"

# API Key from backend/.env
API_KEY="sk-wQUdnk3pwUdEkAGQl85krQcmAE36eLl8DuKj1hLZZrjzuxvV"

# Test 1: Basic connectivity test
echo -e "\n1. Testing basic connectivity to Moonshot API..."
curl -s -w "\nHTTP Status: %{http_code}\n" \
  https://api.moonshot.cn/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $API_KEY" \
  -d '{
    "model": "moonshot-v1-8k",
    "messages": [
      {"role": "system", "content": "You are a helpful assistant."},
      {"role": "user", "content": "Say hello in Chinese."}
    ],
    "temperature": 0.7
  }' | jq .

echo -e "\n================================"

# Test 2: Writing inspiration format test
echo -e "\n2. Testing writing inspiration with JSON format..."
curl -s \
  https://api.moonshot.cn/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $API_KEY" \
  -d '{
    "model": "moonshot-v1-8k",
    "messages": [
      {
        "role": "system",
        "content": "你是OpenPenPal的AI助手，在这个温暖的数字书信平台上，帮助用户进行笔友匹配、生成回信、提供写作灵感和策展信件。请用温暖、友好、富有人文情怀的语气回应。"
      },
      {
        "role": "user",
        "content": "请生成1个写信灵感提示：\n\n主题：日常生活\n风格：温暖友好\n标签：\n\n每个灵感应该：\n1. 提供一个具体的写作切入点\n2. 激发情感共鸣\n3. 适合手写信的形式\n4. 50-100字的描述\n\n返回JSON格式：\n{\n  \"inspirations\": [\n    {\n      \"theme\": \"主题\",\n      \"prompt\": \"写作提示\",\n      \"style\": \"风格\",\n      \"tags\": [\"标签1\", \"标签2\"]\n    }\n  ]\n}"
      }
    ],
    "temperature": 0.7,
    "max_tokens": 500
  }' | jq .

echo -e "\n================================"
echo "Test complete!"
echo ""
echo "If you see valid JSON responses above, the Moonshot API is working correctly."
echo "If you see errors, check:"
echo "1. Is the API key valid?"
echo "2. Can you reach api.moonshot.cn from your network?"
echo "3. Do you have sufficient API quota?"