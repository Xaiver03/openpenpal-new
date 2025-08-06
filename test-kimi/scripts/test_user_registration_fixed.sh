#!/bin/bash

# OpenPenPal 用户注册测试脚本 (修正版)
# 测试10个用户账号的注册功能

echo "🚀 开始OpenPenPal用户注册测试(修正版)..."
echo "========================================"

# API基础URL
BASE_URL="http://localhost:8080"
API_URL="${BASE_URL}/api/v1/auth/register"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 统计变量
SUCCESS_COUNT=0
FAILURE_COUNT=0
TOTAL_TESTS=10

# 日志文件
LOG_FILE="registration_test_fixed_$(date +%Y%m%d_%H%M%S).log"

# 函数：记录日志
log_message() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') - $1" >> "$LOG_FILE"
}

# 函数：发送注册请求
register_user() {
    local username="$1"
    local email="$2"
    local password="$3"
    local nickname="$4"
    local school_code="$5"
    
    echo -e "${BLUE}📝 测试用户: $username${NC}"
    
    # 构建JSON数据
    json_data=$(cat <<EOF
{
  "username": "$username",
  "email": "$email", 
  "password": "$password",
  "nickname": "$nickname",
  "school_code": "$school_code"
}
EOF
)
    
    # 记录请求
    log_message "发送注册请求: $username ($email)"
    log_message "请求数据: $json_data"
    
    # 发送HTTP请求
    response=$(curl -s -w "\n%{http_code}" \
        -X POST \
        -H "Content-Type: application/json" \
        -d "$json_data" \
        "$API_URL" 2>&1)
    
    # 分离响应体和状态码
    response_body=$(echo "$response" | head -n -1)
    http_code=$(echo "$response" | tail -n 1)
    
    # 记录响应
    log_message "HTTP状态码: $http_code"
    log_message "响应内容: $response_body"
    
    # 检查结果
    if [[ "$http_code" == "200" ]] || [[ "$http_code" == "201" ]]; then
        echo -e "   ✅ ${GREEN}注册成功${NC} (HTTP $http_code)"
        echo -e "   📄 响应: $(echo "$response_body" | jq -r '.message // .' 2>/dev/null || echo "$response_body")"
        SUCCESS_COUNT=$((SUCCESS_COUNT + 1))
        log_message "注册成功: $username"
    else
        echo -e "   ❌ ${RED}注册失败${NC} (HTTP $http_code)"
        echo -e "   📄 错误: $(echo "$response_body" | jq -r '.error // .' 2>/dev/null || echo "$response_body")"
        FAILURE_COUNT=$((FAILURE_COUNT + 1))
        log_message "注册失败: $username - HTTP $http_code"
    fi
    
    echo ""
}

# 函数：测试服务连接
test_connection() {
    echo -e "${YELLOW}🔗 测试服务连接...${NC}"
    
    health_response=$(curl -s -w "%{http_code}" "$BASE_URL/health" 2>/dev/null)
    if [[ "$health_response" == *"200" ]]; then
        echo -e "   ✅ ${GREEN}后端服务连接正常${NC}"
        log_message "后端服务连接测试成功"
        return 0
    else
        echo -e "   ❌ ${RED}后端服务连接失败${NC}"
        echo -e "   💡 请确保后端服务在 $BASE_URL 运行"
        log_message "后端服务连接测试失败"
        return 1
    fi
}

# 测试数据数组 (修正格式)
declare -a test_users=(
    "testuser02:test02@tsinghua.edu.cn:password123:李小红:THU001"
    "testuser03:test03@pku.edu.cn:password123:王小强:PKU001"
    "testuser04:test04@bjfu.edu.cn:password123:赵小美:BJFU02"
    "testuser05:test05@ruc.edu.cn:password123:钱小刚:RUC001"
    "testuser06:test06@buaa.edu.cn:password123:孙小华:BUAA01"
    "testuser07:test07@bnu.edu.cn:password123:周小丽:BNU001"
    "testuser08:test08@bjtu.edu.cn:password123:吴小东:BJTU01"
    "testuser09:test09@bit.edu.cn:password123:郑小西:BIT001"
    "testuser10:test10@cau.edu.cn:password123:冯小南:CAU001"
    "testuser11:test11@bupt.edu.cn:password123:陈小北:BUPT01"
)

echo "📋 测试配置:"
echo "   • API地址: $API_URL"
echo "   • 测试用户数: $TOTAL_TESTS"
echo "   • 日志文件: $LOG_FILE"
echo ""

# 检查服务连接
if ! test_connection; then
    echo -e "${RED}❌ 无法连接到后端服务，测试终止${NC}"
    exit 1
fi

echo ""
echo -e "${YELLOW}🧪 开始用户注册测试...${NC}"
echo "========================================"

# 执行注册测试
for i in "${!test_users[@]}"; do
    user_data="${test_users[$i]}"
    IFS=':' read -r username email password nickname school_code <<< "$user_data"
    
    echo -e "${BLUE}测试 $((i+1))/$TOTAL_TESTS${NC}"
    register_user "$username" "$email" "$password" "$nickname" "$school_code"
    
    # 添加延迟避免请求过快
    sleep 0.5
done

echo "========================================"
echo -e "${YELLOW}📊 测试结果统计${NC}"
echo "========================================"
echo -e "总测试数:   $TOTAL_TESTS"
echo -e "成功注册:   ${GREEN}$SUCCESS_COUNT${NC}"
echo -e "注册失败:   ${RED}$FAILURE_COUNT${NC}"

# 计算成功率
if [[ $TOTAL_TESTS -gt 0 ]]; then
    success_rate=$(( SUCCESS_COUNT * 100 / TOTAL_TESTS ))
    echo -e "成功率:     ${GREEN}${success_rate}%${NC}"
fi

echo ""
echo "📄 详细日志已保存到: $LOG_FILE"

# 测试额外功能
echo ""
echo -e "${YELLOW}🔍 附加测试...${NC}"
echo "========================================"

# 测试重复用户名注册
echo -e "${BLUE}📝 测试重复用户名注册${NC}"
register_user "testuser02" "duplicate@test.com" "password123" "重复测试" "TEST01"

# 测试重复邮箱
echo -e "${BLUE}📝 测试重复邮箱注册${NC}"
register_user "duplicate_email" "test02@tsinghua.edu.cn" "password123" "重复邮箱测试" "TEST02"

# 测试无效数据
echo -e "${BLUE}📝 测试无效数据注册 (用户名过短)${NC}"
register_user "ab" "invalid@test.com" "password123" "无效数据测试" "INVALID"

# 测试无效数据
echo -e "${BLUE}📝 测试无效数据注册 (密码过短)${NC}"
register_user "invalid_pass" "invalid2@test.com" "123" "无效密码测试" "INVALID"

# 测试无效数据
echo -e "${BLUE}📝 测试无效数据注册 (学校代码不是6位)${NC}"
register_user "invalid_school" "invalid3@test.com" "password123" "无效学校测试" "ABC"

echo ""
echo -e "${YELLOW}🏁 所有测试完成！${NC}"

# 显示一些成功注册的用户信息
if [[ $SUCCESS_COUNT -gt 0 ]]; then
    echo ""
    echo -e "${YELLOW}📋 验证数据库中的注册用户...${NC}"
    echo "   可以通过以下用户名和密码登录测试:"
    echo "   • 用户名: testuser02, 密码: password123"
    echo "   • 用户名: testuser03, 密码: password123"
    echo "   • 等等..."
fi

# 根据结果返回适当的退出码
if [[ $FAILURE_COUNT -eq 0 ]]; then
    echo -e "${GREEN}✅ 所有测试通过！${NC}"
    log_message "所有注册测试完成 - 成功率100%"
    exit 0
else
    if [[ $SUCCESS_COUNT -gt 0 ]]; then
        echo -e "${YELLOW}⚠️  部分测试通过，但存在一些失败用例${NC}"
        log_message "注册测试完成 - 成功 $SUCCESS_COUNT 个，失败 $FAILURE_COUNT 个"
        exit 0
    else
        echo -e "${RED}❌ 所有测试都失败了${NC}"
        log_message "注册测试完成 - 所有测试失败"
        exit 1
    fi
fi