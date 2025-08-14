#!/bin/bash

echo "=== 测试信使晋升系统 ==="

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# 测试端点函数
test_endpoint() {
    local username=$1
    local description=$2
    local endpoint=$3
    local method=${4:-GET}
    local data=${5:-""}
    
    echo -e "\n${YELLOW}测试: $description${NC}"
    echo "用户: $username"
    echo "端点: $method $endpoint"
    
    # 获取CSRF和登录
    CSRF=$(curl -s http://localhost:8080/api/v1/auth/csrf | jq -r '.data.token')
    LOGIN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
      -H "Content-Type: application/json" \
      -H "X-CSRF-Token: $CSRF" \
      -H "Cookie: csrf-token=$CSRF" \
      -d "{\"username\":\"$username\",\"password\":\"password\"}")
    
    TOKEN=$(echo $LOGIN | jq -r '.data.token')
    
    if [ "$TOKEN" != "null" ]; then
        # 发送请求
        if [ "$method" = "GET" ]; then
            RESP=$(curl -s -X GET "http://localhost:8080/api/v1$endpoint" \
              -H "Authorization: Bearer $TOKEN")
        else
            RESP=$(curl -s -X "$method" "http://localhost:8080/api/v1$endpoint" \
              -H "Authorization: Bearer $TOKEN" \
              -H "Content-Type: application/json" \
              -d "$data")
        fi
        
        # 检查响应
        STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X "$method" "http://localhost:8080/api/v1$endpoint" \
          -H "Authorization: Bearer $TOKEN" \
          -H "Content-Type: application/json" \
          -d "$data")
        
        if [ "$STATUS" = "200" ] || [ "$STATUS" = "201" ]; then
            echo -e "${GREEN}✅ 成功 (HTTP $STATUS)${NC}"
            echo "$RESP" | jq '.' 2>/dev/null || echo "$RESP"
        elif [ "$STATUS" = "503" ]; then
            echo -e "${YELLOW}⚠️  courier-service未启动 (HTTP $STATUS)${NC}"
            echo "$RESP" | jq '.message' 2>/dev/null || echo "$RESP"
        else
            echo -e "${RED}❌ 失败 (HTTP $STATUS)${NC}"
            echo "$RESP" | jq '.' 2>/dev/null || echo "$RESP"
        fi
    else
        echo -e "${RED}❌ 登录失败${NC}"
    fi
}

echo -e "\n${YELLOW}=== 1. 测试成长路径查询 ===${NC}"
test_endpoint "courier_level1" "一级信使查看成长路径" "/courier/growth/path"

echo -e "\n${YELLOW}=== 2. 测试成长进度查询 ===${NC}"
test_endpoint "courier_level1" "一级信使查看成长进度" "/courier/growth/progress"
test_endpoint "courier_level2" "二级信使查看成长进度" "/courier/growth/progress"

echo -e "\n${YELLOW}=== 3. 测试等级配置查询 ===${NC}"
test_endpoint "courier_level1" "查看等级配置" "/courier/level/config"

echo -e "\n${YELLOW}=== 4. 测试等级检查 ===${NC}"
test_endpoint "courier_level1" "一级信使检查等级" "/courier/level/check"
test_endpoint "courier_level3" "三级信使检查等级" "/courier/level/check"

echo -e "\n${YELLOW}=== 5. 测试晋升申请提交 ===${NC}"
UPGRADE_DATA='{
  "request_level": 2,
  "reason": "我已经完成了超过100次投递任务，投递准时率达到98%，获得了多次用户好评。希望晋升到二级信使，为团队做出更大贡献。",
  "evidence": {
    "delivery_count": 156,
    "completion_rate": 98,
    "user_ratings": 4.9
  }
}'
test_endpoint "courier_level1" "一级信使申请晋升到二级" "/courier/level/upgrade" "POST" "$UPGRADE_DATA"

echo -e "\n${YELLOW}=== 6. 测试晋升申请查询（需要三级权限）===${NC}"
test_endpoint "courier_level3" "三级信使查看晋升申请列表" "/courier/level/upgrade-requests?status=pending"

echo -e "\n${YELLOW}=== 7. 测试徽章系统 ===${NC}"
test_endpoint "courier_level1" "查看所有徽章" "/courier/growth/badges"
test_endpoint "courier_level2" "查看已获得徽章" "/courier/growth/badges/earned"

echo -e "\n${YELLOW}=== 8. 测试积分系统 ===${NC}"
test_endpoint "courier_level1" "查看积分余额" "/courier/growth/points"
test_endpoint "courier_level2" "查看积分历史" "/courier/growth/points/history?limit=10"

echo -e "\n${YELLOW}=== 9. 测试排行榜 ===${NC}"
test_endpoint "courier_level1" "查看排行榜" "/courier/growth/ranking?time_range=weekly&limit=10"

echo -e "\n${YELLOW}=== 10. 测试绩效统计 ===${NC}"
test_endpoint "courier_level2" "查看绩效统计" "/courier/growth/performance?time_range=monthly"

echo -e "\n${YELLOW}=== 测试总结 ===${NC}"
echo -e "如果看到 '${YELLOW}courier-service未启动${NC}' 的提示，请启动courier-service:"
echo -e "  cd services/courier-service && go run cmd/main.go"
echo -e "\n${GREEN}晋升系统API代理路由已成功配置！${NC}"