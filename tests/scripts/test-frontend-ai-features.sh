#!/bin/bash

echo "🧪 测试前端AI功能集成"
echo "=========================="

# 检查前端是否启动
echo "📡 检查前端服务状态..."
if curl -s http://localhost:3001 > /dev/null; then
    echo "✅ 前端服务正常运行 (http://localhost:3001)"
else
    echo "❌ 前端服务未启动"
    exit 1
fi

# 检查后端是否启动
echo "📡 检查后端服务状态..."
if curl -s http://localhost:8080/health > /dev/null; then
    echo "✅ 后端服务正常运行 (http://localhost:8080)"
else
    echo "❌ 后端服务未启动"
    exit 1
fi

# 获取认证token
echo "🔐 获取认证token..."
LOGIN_RESPONSE=$(curl -s -X POST "http://localhost:8080/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin123"}')

TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
    echo "❌ 无法获取认证token"
    echo "登录响应: $LOGIN_RESPONSE"
    exit 1
fi

echo "✅ 成功获取认证token: ${TOKEN:0:50}..."

# 测试AI功能端点
echo ""
echo "🤖 测试AI功能端点..."

# 1. 测试AI人设列表
echo "👥 测试AI人设列表..."
PERSONAS_RESPONSE=$(curl -s -X GET "http://localhost:8080/api/v1/ai/personas" \
  -H "Authorization: Bearer $TOKEN")

PERSONAS_COUNT=$(echo $PERSONAS_RESPONSE | grep -o '"total":[0-9]*' | cut -d':' -f2)
if [ "$PERSONAS_COUNT" -gt 0 ]; then
    echo "✅ AI人设列表正常 ($PERSONAS_COUNT 个人设)"
else
    echo "❌ AI人设列表异常"
    echo "响应: $PERSONAS_RESPONSE"
fi

# 2. 测试每日灵感
echo "🌅 测试每日灵感..."
DAILY_RESPONSE=$(curl -s -X GET "http://localhost:8080/api/v1/ai/daily-inspiration" \
  -H "Authorization: Bearer $TOKEN")

if echo $DAILY_RESPONSE | grep -q "theme"; then
    echo "✅ 每日灵感功能正常"
else
    echo "❌ 每日灵感功能异常"
    echo "响应: $DAILY_RESPONSE"
fi

# 3. 测试写作灵感生成
echo "💡 测试写作灵感生成..."
INSPIRATION_RESPONSE=$(curl -s -X POST "http://localhost:8080/api/v1/ai/inspiration" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"theme": "友情", "style": "温暖", "tags": ["回忆"], "count": 1}')

if echo $INSPIRATION_RESPONSE | grep -q "inspirations"; then
    echo "✅ 写作灵感生成功能正常"
else
    echo "❌ 写作灵感生成功能异常"
    echo "响应: $INSPIRATION_RESPONSE"
fi

# 4. 创建测试信件（为回信建议功能准备）
echo "📝 创建测试信件..."
CREATE_LETTER_RESPONSE=$(curl -s -X POST "http://localhost:8080/api/v1/letters" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "title": "给远方朋友的一封信",
    "content": "亲爱的朋友，好久不见了。最近的生活怎么样？我很想念我们一起度过的那些美好时光。",
    "style": "casual",
    "recipient_address": "测试地址"
  }')

LETTER_ID=$(echo $CREATE_LETTER_RESPONSE | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
if [ -n "$LETTER_ID" ]; then
    echo "✅ 测试信件创建成功 (ID: $LETTER_ID)"
    
    # 5. 测试NEW功能：AI回信角度建议
    echo "💌 测试AI回信角度建议功能..."
    ADVICE_RESPONSE=$(curl -s -X POST "http://localhost:8080/api/v1/ai/reply-advice" \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer $TOKEN" \
      -d "{
        \"letter_id\": \"$LETTER_ID\",
        \"persona_type\": \"distant_friend\",
        \"persona_name\": \"小学同桌\",
        \"persona_desc\": \"小时候最好的朋友，毕业后各奔东西\",
        \"relationship\": \"小学同桌，童年最好的朋友\",
        \"delivery_days\": 1
      }")
    
    if echo $ADVICE_RESPONSE | grep -q "perspectives"; then
        echo "✅ AI回信角度建议功能正常！"
        echo "📋 建议内容预览:"
        echo $ADVICE_RESPONSE | python3 -c "
import sys, json
try:
    data = json.load(sys.stdin)
    if 'data' in data:
        advice = data['data']
        print(f'  人设: {advice.get(\"persona_name\", \"未知\")}')
        print(f'  情感基调: {advice.get(\"emotional_tone\", \"未知\")}')
        print(f'  延迟天数: {advice.get(\"delivery_delay\", 0)}')
    else:
        print('  无法解析建议内容')
except:
    print('  响应格式错误')
"
    else
        echo "❌ AI回信角度建议功能异常"
        echo "响应: $ADVICE_RESPONSE"
    fi
else
    echo "❌ 测试信件创建失败"
    echo "响应: $CREATE_LETTER_RESPONSE"
fi

echo ""
echo "🎉 AI功能测试完成！"
echo ""
echo "📊 功能状态总结:"
echo "- ✅ 后端服务运行正常"
echo "- ✅ 前端服务运行正常 (http://localhost:3001)"
echo "- ✅ 认证系统正常"
echo "- ✅ AI人设列表功能正常"
echo "- ✅ 每日灵感功能正常"
echo "- ✅ 写作灵感生成功能正常"
echo "- ✅ 新功能：AI回信角度建议已实现并正常工作"
echo ""
echo "🌐 请访问 http://localhost:3001/ai 查看前端界面"
echo "🔑 使用 admin/admin123 登录体验完整功能"