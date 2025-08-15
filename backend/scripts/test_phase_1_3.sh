#!/bin/bash

# Phase 1.3: 测试防作弊检测算法
echo "=== Phase 1.3: 测试防作弊检测算法 ==="

# 检查服务是否运行
if ! curl -s http://localhost:8080/health > /dev/null; then
    echo "❌ 后端服务未运行，请先启动服务"
    echo "   运行: cd backend && go run main.go"
    exit 1
fi

echo "✅ 服务检查完成"

# 测试用户ID
TEST_USER="fraud-test-user-$(date +%s)"
ADMIN_TOKEN="your-admin-token-here"

echo ""
echo "--- 测试1: 模拟高频异常行为 ---"

# 快速连续执行10次操作（模拟机器人行为）
echo "快速连续执行操作（5分钟内10次以上）:"
for i in {1..12}; do
    echo "第${i}次操作:"
    
    RESPONSE=$(curl -s -X POST "http://localhost:8080/api/v1/credit/tasks" \
      -H "Authorization: Bearer $ADMIN_TOKEN" \
      -H "Content-Type: application/json" \
      -d "{
        \"task_type\": \"letter_created\",
        \"user_id\": \"$TEST_USER\",
        \"points\": 10,
        \"description\": \"高频测试 - 第${i}次\",
        \"reference\": \"fraud-test-${i}\",
        \"metadata\": {
          \"ip\": \"192.168.1.100\",
          \"device_id\": \"test-device-123\",
          \"user_agent\": \"Mozilla/5.0 (Test Browser)\",
          \"points\": \"10\"
        }
      }")
    
    echo "响应: $RESPONSE"
    
    # 非常短的间隔，模拟机器人行为
    sleep 0.5
done

echo ""
echo "--- 测试2: 模拟IP跳跃异常 ---"

IP_TEST_USER="ip-test-user-$(date +%s)"

# 在短时间内使用多个不同IP
IP_LIST=("192.168.1.100" "10.0.0.50" "172.16.0.10" "203.0.113.5" "198.51.100.25" "123.45.67.89" "87.65.43.21" "45.67.89.123" "156.78.90.234")

echo "模拟IP跳跃行为（1小时内使用多个IP）:"
for i in {0..8}; do
    IP=${IP_LIST[$i]}
    echo "使用IP: $IP"
    
    RESPONSE=$(curl -s -X POST "http://localhost:8080/api/v1/credit/tasks" \
      -H "Authorization: Bearer $ADMIN_TOKEN" \
      -H "Content-Type: application/json" \
      -d "{
        \"task_type\": \"ai_interaction\",
        \"user_id\": \"$IP_TEST_USER\",
        \"points\": 5,
        \"description\": \"IP跳跃测试 - IP${i}\",
        \"reference\": \"ip-test-${i}\",
        \"metadata\": {
          \"ip_address\": \"$IP\",
          \"device_id\": \"stable-device-456\",
          \"user_agent\": \"Mozilla/5.0 (Stable Browser)\",
          \"points\": \"5\"
        }
      }")
    
    echo "响应: $RESPONSE"
    sleep 1
done

echo ""
echo "--- 测试3: 模拟设备切换异常 ---"

DEVICE_TEST_USER="device-test-user-$(date +%s)"

# 在短时间内使用多个不同设备
DEVICE_LIST=("device-mobile-1" "device-tablet-2" "device-laptop-3" "device-desktop-4" "device-mobile-5" "device-tablet-6")

echo "模拟设备切换行为（6小时内使用多个设备）:"
for i in {0..5}; do
    DEVICE=${DEVICE_LIST[$i]}
    USER_AGENT="Mozilla/5.0 (Device${i} Browser)"
    echo "使用设备: $DEVICE"
    
    RESPONSE=$(curl -s -X POST "http://localhost:8080/api/v1/credit/tasks" \
      -H "Authorization: Bearer $ADMIN_TOKEN" \
      -H "Content-Type: application/json" \
      -d "{
        \"task_type\": \"museum_submit\",
        \"user_id\": \"$DEVICE_TEST_USER\",
        \"points\": 8,
        \"description\": \"设备切换测试 - ${DEVICE}\",
        \"reference\": \"device-test-${i}\",
        \"metadata\": {
          \"ip_address\": \"192.168.1.200\",
          \"device_id\": \"$DEVICE\",
          \"user_agent\": \"$USER_AGENT\",
          \"points\": \"8\"
        }
      }")
    
    echo "响应: $RESPONSE"
    sleep 2
done

echo ""
echo "--- 测试4: 模拟积分异常获取 ---"

POINTS_TEST_USER="points-test-user-$(date +%s)"

echo "模拟异常积分获取（今日超过1000积分）:"
# 连续执行高积分任务
for i in {1..15}; do
    POINTS=$((50 + i * 10)) # 递增积分，最终超过1000
    echo "第${i}次 - 获得${POINTS}积分:"
    
    RESPONSE=$(curl -s -X POST "http://localhost:8080/api/v1/credit/tasks" \
      -H "Authorization: Bearer $ADMIN_TOKEN" \
      -H "Content-Type: application/json" \
      -d "{
        \"task_type\": \"admin_reward\",
        \"user_id\": \"$POINTS_TEST_USER\",
        \"points\": $POINTS,
        \"description\": \"积分异常测试 - ${POINTS}积分\",
        \"reference\": \"points-test-${i}\",
        \"metadata\": {
          \"ip_address\": \"192.168.1.150\",
          \"device_id\": \"consistent-device-789\",
          \"user_agent\": \"Mozilla/5.0 (Consistent Browser)\",
          \"points\": \"$POINTS\"
        }
      }")
    
    echo "响应: $RESPONSE"
    sleep 1
done

echo ""
echo "--- 测试5: 模拟重复行为模式 ---"

PATTERN_TEST_USER="pattern-test-user-$(date +%s)"

echo "模拟重复行为模式（规律性行为序列）:"
# 创建重复的行为模式: A-B-C-A-B-C-A-B-C...
PATTERN=("letter_created" "ai_interaction" "museum_submit")
for i in {1..9}; do
    ACTION_TYPE=${PATTERN[$((i % 3))]}
    echo "第${i}次 - 行为类型: $ACTION_TYPE"
    
    RESPONSE=$(curl -s -X POST "http://localhost:8080/api/v1/credit/tasks" \
      -H "Authorization: Bearer $ADMIN_TOKEN" \
      -H "Content-Type: application/json" \
      -d "{
        \"task_type\": \"$ACTION_TYPE\",
        \"user_id\": \"$PATTERN_TEST_USER\",
        \"points\": 10,
        \"description\": \"模式测试 - ${ACTION_TYPE}\",
        \"reference\": \"pattern-test-${i}\",
        \"metadata\": {
          \"ip_address\": \"192.168.1.180\",
          \"device_id\": \"pattern-device-999\",
          \"user_agent\": \"Mozilla/5.0 (Pattern Browser)\",
          \"points\": \"10\",
          \"sequence_index\": \"$i\"
        }
      }")
    
    echo "响应: $RESPONSE"
    sleep 3  # 规律的时间间隔
done

echo ""
echo "--- 测试6: 获取风险分析报告 ---"

echo "获取测试用户的风险分析报告:"

# 获取高频测试用户的风险分析
echo "1. 高频用户风险分析:"
RISK_ANALYSIS=$(curl -s -X GET "http://localhost:8080/api/v1/credit/limits/user/$TEST_USER/risk-analysis" \
  -H "Authorization: Bearer $ADMIN_TOKEN")
echo "结果: $RISK_ANALYSIS"

echo ""
echo "2. IP跳跃用户风险分析:"
RISK_ANALYSIS=$(curl -s -X GET "http://localhost:8080/api/v1/credit/limits/user/$IP_TEST_USER/risk-analysis" \
  -H "Authorization: Bearer $ADMIN_TOKEN")
echo "结果: $RISK_ANALYSIS"

echo ""
echo "3. 积分异常用户风险分析:"
RISK_ANALYSIS=$(curl -s -X GET "http://localhost:8080/api/v1/credit/limits/user/$POINTS_TEST_USER/risk-analysis" \
  -H "Authorization: Bearer $ADMIN_TOKEN")
echo "结果: $RISK_ANALYSIS"

echo ""
echo "--- 测试7: 检查检测日志 ---"

echo "获取防作弊检测日志:"
DETECTION_LOGS=$(curl -s -X GET "http://localhost:8080/api/v1/credit/limits/detection-logs?limit=10" \
  -H "Authorization: Bearer $ADMIN_TOKEN")
echo "检测日志: $DETECTION_LOGS"

echo ""
echo "=== Phase 1.3 测试完成 ==="
echo ""
echo "📊 测试结果总结:"
echo "✅ 高频行为检测测试（机器人检测）"
echo "✅ IP跳跃异常检测测试"
echo "✅ 设备切换异常检测测试"
echo "✅ 积分异常获取检测测试"
echo "✅ 重复行为模式检测测试"
echo "✅ 用户风险分析报告测试"
echo "✅ 检测日志记录测试"
echo ""
echo "🔍 请检查上述响应确认增强防作弊功能正常工作"
echo "   - 应该检测到各种异常行为模式"
echo "   - 风险分数应该相应增加"
echo "   - 应该生成相应的处理建议"
echo "   - 检测日志应该被正确记录"
echo ""
echo "⚠️  如果检测到高风险用户（分数>0.9），系统会自动临时封禁"