#!/bin/bash

# Phase 1.2: 测试每日/每周限制控制
echo "=== Phase 1.2: 测试每日/每周限制控制 ==="

# 检查服务是否运行
if ! curl -s http://localhost:8080/health > /dev/null; then
    echo "❌ 后端服务未运行，请先启动服务"
    echo "   运行: cd backend && go run main.go"
    exit 1
fi

# 检查Redis是否可用
if ! curl -s http://localhost:8080/api/v1/health/redis > /dev/null; then
    echo "⚠️ Redis不可用，将使用降级模式"
fi

echo "✅ 服务检查完成"

# 测试用户ID
TEST_USER="test-limit-user-$(date +%s)"
ADMIN_TOKEN="your-admin-token-here"

echo ""
echo "--- 测试1: 获取当前限制规则 ---"

# 获取限制规则
curl -s -X GET "http://localhost:8080/api/v1/credit/limits/rules" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" | jq '.' || echo "❌ 获取规则失败"

echo ""
echo "--- 测试2: 创建测试限制规则 ---"

# 创建每日限制规则
DAILY_RULE_RESPONSE=$(curl -s -X POST "http://localhost:8080/api/v1/credit/limits/rules" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "action_type": "test_action",
    "limit_type": "count",
    "limit_period": "daily",
    "max_count": 3,
    "max_points": 0,
    "enabled": true
  }')

echo "每日规则创建结果: $DAILY_RULE_RESPONSE"

# 创建每周积分限制规则
WEEKLY_RULE_RESPONSE=$(curl -s -X POST "http://localhost:8080/api/v1/credit/limits/rules" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "action_type": "test_action",
    "limit_type": "points",
    "limit_period": "weekly",
    "max_count": 0,
    "max_points": 100,
    "enabled": true
  }')

echo "每周规则创建结果: $WEEKLY_RULE_RESPONSE"

echo ""
echo "--- 测试3: 模拟用户行为 - 每日限制 ---"

# 模拟用户执行行为（应该前3次成功，第4次失败）
for i in {1..5}; do
    echo "第${i}次尝试:"
    
    # 模拟创建积分任务（这会触发限制检查）
    RESPONSE=$(curl -s -X POST "http://localhost:8080/api/v1/credit/tasks" \
      -H "Authorization: Bearer $ADMIN_TOKEN" \
      -H "Content-Type: application/json" \
      -d "{
        \"task_type\": \"test_action\",
        \"user_id\": \"$TEST_USER\",
        \"points\": 10,
        \"description\": \"测试每日限制 - 第${i}次\",
        \"reference\": \"test-ref-${i}\",
        \"metadata\": {
          \"ip\": \"192.168.1.100\",
          \"device_id\": \"test-device-123\"
        }
      }")
    
    echo "响应: $RESPONSE"
    echo ""
    
    # 稍等一下再继续
    sleep 1
done

echo ""
echo "--- 测试4: 检查用户限制状态 ---"

# 检查用户当前限制状态
USER_STATUS=$(curl -s -X GET "http://localhost:8080/api/v1/credit/limits/user/$TEST_USER/status" \
  -H "Authorization: Bearer $ADMIN_TOKEN")

echo "用户限制状态: $USER_STATUS"

echo ""
echo "--- 测试5: 模拟每周积分累积 ---"

WEEKLY_TEST_USER="weekly-test-user-$(date +%s)"

# 模拟一周内的积分累积（每天15积分，应该在第7天达到限制）
for day in {1..8}; do
    echo "第${day}天尝试获得15积分:"
    
    RESPONSE=$(curl -s -X POST "http://localhost:8080/api/v1/credit/tasks" \
      -H "Authorization: Bearer $ADMIN_TOKEN" \
      -H "Content-Type: application/json" \
      -d "{
        \"task_type\": \"test_action\",
        \"user_id\": \"$WEEKLY_TEST_USER\",
        \"points\": 15,
        \"description\": \"测试每周限制 - 第${day}天\",
        \"reference\": \"weekly-test-${day}\",
        \"metadata\": {
          \"day\": \"$day\",
          \"ip\": \"192.168.1.101\"
        }
      }")
    
    echo "响应: $RESPONSE"
    echo ""
    
    sleep 1
done

echo ""
echo "--- 测试6: 检查限制统计 ---"

# 获取限制统计信息
STATS=$(curl -s -X GET "http://localhost:8080/api/v1/credit/limits/stats" \
  -H "Authorization: Bearer $ADMIN_TOKEN")

echo "限制统计: $STATS"

echo ""
echo "=== Phase 1.2 测试完成 ==="
echo ""
echo "📊 测试结果总结:"
echo "✅ 每日行为次数限制测试"
echo "✅ 每周积分总量限制测试"
echo "✅ 限制规则配置测试"
echo "✅ 用户状态查询测试"
echo "✅ 统计信息获取测试"
echo ""
echo "🔍 请检查上述响应确认功能正常工作"
echo "   - 前3次每日请求应该成功"
echo "   - 第4次每日请求应该被限制"
echo "   - 每周积分在达到100分后应该被限制"