#!/bin/bash

# OpenPenPal 前后端集成测试脚本
# 全系统功能验证测试

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 服务配置
FRONTEND_URL="http://localhost:3000"
BACKEND_URL="http://localhost:8080"
GATEWAY_URL="http://localhost:8083"

# 测试配置
TEST_EMAIL="integration_test@penpal.com"
TEST_PASSWORD="test123456"
TEST_SCHOOL_CODE="TEST01"
TEST_NICKNAME="Integration Tester"

# 工具检查
check_dependencies() {
    echo -e "${BLUE}🔍 检查测试依赖...${NC}"
    
    if ! command -v curl &> /dev/null; then
        echo -e "${RED}❌ curl 未安装${NC}"
        exit 1
    fi
    
    if ! command -v jq &> /dev/null; then
        echo -e "${RED}❌ jq 未安装${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}✅ 依赖检查通过${NC}"
}

# 服务健康检查
check_service_health() {
    echo -e "${BLUE}🏥 检查服务健康状态...${NC}"
    
    services=(
        "$BACKEND_URL/health:主后端"
        "$FRONTEND_URL:前端服务"
        "$GATEWAY_URL/health:网关服务"
    )
    
    for service_info in "${services[@]}"; do
        IFS=':' read -r url name <<< "$service_info"
        
        if curl -s "$url" >/dev/null; then
            echo -e "${GREEN}✅ $name 运行正常${NC}"
        else
            echo -e "${RED}❌ $name 无法访问${NC}"
            return 1
        fi
    done
}

# 生成随机字符串
generate_random_string() {
    date +%s | sha256sum | base64 | head -c 8
}

# 注册新用户
register_user() {
    local email=$1
    local password=$2
    local school_code=$3
    local nickname=$4
    
    echo -e "${YELLOW}📝 注册用户: $email${NC}"
    
    response=$(curl -s -X POST "$BACKEND_URL/api/v1/auth/register" \
        -H "Content-Type: application/json" \
        -d "{
            \"username\": \"$email\",
            \"email\": \"$email\",
            \"password\": \"$password\",
            \"nickname\": \"$nickname\",
            \"school_code\": \"$school_code\"
        }")
    
    if echo "$response" | jq -e '.message == "User registered successfully"' >/dev/null; then
        echo -e "${GREEN}✅ 注册成功${NC}"
        echo "$response" | jq '.user.id'
        return 0
    else
        echo -e "${RED}❌ 注册失败: $(echo "$response" | jq -r '.error' 2>/dev/null || echo 'Unknown error')${NC}"
        return 1
    fi
}

# 用户登录
login_user() {
    local email=$1
    local password=$2
    
    echo -e "${YELLOW}🔑 用户登录: $email${NC}"
    
    response=$(curl -s -X POST "$BACKEND_URL/api/v1/auth/login" \
        -H "Content-Type: application/json" \
        -d "{
            \"username\": \"$email\",
            \"password\": \"$password\"
        }")
    
    if echo "$response" | jq -e '.message == "Login successful"' >/dev/null; then
        token=$(echo "$response" | jq -r '.data.token')
        user_id=$(echo "$response" | jq -r '.data.user.id')
        role=$(echo "$response" | jq -r '.data.user.role')
        
        echo -e "${GREEN}✅ 登录成功${NC}"
        echo -e "${BLUE}   用户ID: $user_id${NC}"
        echo -e "${BLUE}   角色: $role${NC}"
        echo -e "${BLUE}   Token: ${token:0:20}...${NC}"
        
        echo "$token"
        return 0
    else
        echo -e "${RED}❌ 登录失败${NC}"
        return 1
    fi
}

# 获取用户信息
get_user_profile() {
    local token=$1
    
    echo -e "${YELLOW}👤 获取用户资料${NC}"
    
    response=$(curl -s -X GET "$BACKEND_URL/api/v1/users/profile" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $token")
    
    echo "$response" | jq '.'
}

# 创建信件
create_letter() {
    local token=$1
    local title=$2
    local content=$3
    
    echo -e "${YELLOW}✉️ 创建信件: $title${NC}"
    
    response=$(curl -s -X POST "$BACKEND_URL/api/v1/letters" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $token" \
        -d "{
            \"title\": \"$title\",
            \"content\": \"$content\",
            \"style\": \"classic\",
            \"anonymous\": false
        }")
    
    if echo "$response" | jq -e '.message == "Letter created successfully"' >/dev/null; then
        letter_id=$(echo "$response" | jq -r '.data.id')
        letter_code=$(echo "$response" | jq -r '.data.code')
        
        echo -e "${GREEN}✅ 信件创建成功${NC}"
        echo -e "${BLUE}   信件ID: $letter_id${NC}"
        echo -e "${BLUE}   信件代码: $letter_code${NC}"
        
        echo "$letter_code"
        return 0
    else
        echo -e "${RED}❌ 信件创建失败${NC}"
        echo "$response"
        return 1
    fi
}

# 获取信件列表
get_letters() {
    local token=$1
    
    echo -e "${YELLOW}📋 获取信件列表${NC}"
    
    response=$(curl -s -X GET "$BACKEND_URL/api/v1/letters" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $token")
    
    count=$(echo "$response" | jq '.data | length')
    echo -e "${GREEN}✅ 获取到 $count 封信件${NC}"
    echo "$response" | jq '.data[0:3]'
}

# 生成二维码
generate_qr_code() {
    local token=$1
    local letter_id=$2
    
    echo -e "${YELLOW}📱 生成二维码${NC}"
    
    response=$(curl -s -X POST "$BACKEND_URL/api/v1/codes/generate" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $token" \
        -d "{
            \"letter_id\": \"$letter_id\",
            \"expires_in\": 3600
        }")
    
    if echo "$response" | jq -e '.success' >/dev/null; then
        qr_code=$(echo "$response" | jq -r '.data.qr_code')
        echo -e "${GREEN}✅ 二维码生成成功${NC}"
        echo -e "${BLUE}   二维码: $qr_code${NC}"
        return 0
    else
        echo -e "${RED}❌ 二维码生成失败${NC}"
        return 1
    fi
}

# 测试前端页面可访问性
test_frontend_pages() {
    echo -e "${BLUE}🌐 测试前端页面${NC}"
    
    pages=(
        "/:首页"
        "/login:登录页"
        "/register:注册页"
        "/write:写信页"
        "/mailbox:收件箱"
        "/profile:个人资料"
        "/courier/scan:扫码页"
    )
    
    for page_info in "${pages[@]}"; do
        IFS=':' read -r path name <<< "$page_info"
        
        response=$(curl -s -o /dev/null -w "%{http_code}" "$FRONTEND_URL$path")
        
        if [ "$response" == "200" ]; then
            echo -e "${GREEN}✅ $name 可访问${NC}"
        else
            echo -e "${RED}❌ $name 访问失败 ($response)${NC}"
        fi
    done
}

# 测试API端点
test_api_endpoints() {
    echo -e "${BLUE}🔌 测试API端点${NC}"
    
    endpoints=(
        "GET:/api/v1/health:健康检查"
        "GET:/api/v1/letters:信件列表"
        "POST:/api/v1/auth/login:登录"
        "POST:/api/v1/auth/register:注册"
    )
    
    for endpoint_info in "${endpoints[@]}"; do
        IFS=':' read -r method path name <<< "$endpoint_info"
        
        if [ "$method" == "GET" ]; then
            response=$(curl -s -o /dev/null -w "%{http_code}" "$BACKEND_URL$path")
        else
            response=$(curl -s -o /dev/null -w "%{http_code}" -X $method "$BACKEND_URL$path" \
                -H "Content-Type: application/json" -d '{}')
        fi
        
        if [[ "$response" =~ ^(200|201|400|401)$ ]]; then
            echo -e "${GREEN}✅ $name ($method $path)${NC}"
        else
            echo -e "${RED}❌ $name ($method $path) - $response${NC}"
        fi
    done
}

# 测试权限系统
test_permission_system() {
    echo -e "${BLUE}🔐 测试权限系统${NC}"
    
    # 测试用户注册后角色
    test_email="permission_test_$(generate_random_string)@penpal.com"
    
    user_id=$(register_user "$test_email" "test123" "PERM01" "Permission Tester")
    if [ $? -eq 0 ]; then
        token=$(login_user "$test_email" "test123")
        if [ $? -eq 0 ]; then
            role=$(get_user_profile "$token" | jq -r '.data.role')
            if [ "$role" == "user" ]; then
                echo -e "${GREEN}✅ 权限系统正确: 新用户角色为 '$role'${NC}"
            else
                echo -e "${RED}❌ 权限系统错误: 期望 'user', 实际 '$role'${NC}"
            fi
        fi
    fi
}

# 测试数据一致性
test_data_consistency() {
    echo -e "${BLUE}🔄 测试数据一致性${NC}"
    
    # 创建测试用户和信件，验证数据同步
    test_email="consistency_$(generate_random_string)@penpal.com"
    
    user_id=$(register_user "$test_email" "test123" "CONS01" "Consistency Tester")
    if [ $? -eq 0 ]; then
        token=$(login_user "$test_email" "test123")
        if [ $? -eq 0 ]; then
            
            # 创建信件
            letter_code=$(create_letter "$token" "一致性测试信件" "这是测试数据一致性的信件内容")
            if [ $? -eq 0 ]; then
                
                # 验证信件出现在用户信件列表中
                letters=$(get_letters "$token")
                if echo "$letters" | jq -e ".data[] | select(.code == \"$letter_code\")" >/dev/null; then
                    echo -e "${GREEN}✅ 数据一致性验证通过${NC}"
                else
                    echo -e "${RED}❌ 数据一致性验证失败${NC}"
                fi
            fi
        fi
    fi
}

# 压力测试
stress_test() {
    echo -e "${BLUE}⚡ 执行简单压力测试${NC}"
    
    # 并发注册测试
    for i in {1..3}; do
        (
            email="stress_$i$(generate_random_string)@penpal.com"
            register_user "$email" "stress123" "STRS01" "Stress Test $i"
        ) &
    done
    
    wait
    echo -e "${GREEN}✅ 压力测试完成${NC}"
}

# 生成测试报告
generate_report() {
    local timestamp=$(date +"%Y%m%d_%H%M%S")
    local report_file="test-kimi/reports/integration_test_${timestamp}.json"
    
    mkdir -p test-kimi/reports
    
    cat > "$report_file" << EOF
{
  "test_type": "integration_test",
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "environment": {
    "frontend_url": "$FRONTEND_URL",
    "backend_url": "$BACKEND_URL",
    "gateway_url": "$GATEWAY_URL"
  },
  "services": {
    "frontend": "$(curl -s -o /dev/null -w "%{http_code}" "$FRONTEND_URL" || echo "down")",
    "backend": "$(curl -s -o /dev/null -w "%{http_code}" "$BACKEND_URL/health" || echo "down")",
    "gateway": "$(curl -s -o /dev/null -w "%{http_code}" "$GATEWAY_URL/health" || echo "down")"
  },
  "tests": {
    "user_registration": "completed",
    "user_login": "completed", 
    "letter_creation": "completed",
    "page_accessibility": "completed",
    "api_endpoints": "completed",
    "permission_system": "completed",
    "data_consistency": "completed"
  }
}
EOF
    
    echo -e "${GREEN}📊 测试报告已生成: $report_file${NC}"
}

# 主测试流程
main() {
    echo -e "${YELLOW}🚀 OpenPenPal 集成测试开始${NC}"
    echo "================================="
    
    check_dependencies
    check_service_health
    
    echo ""
    echo -e "${BLUE}📋 开始功能测试...${NC}"
    
    test_frontend_pages
    test_api_endpoints
    test_permission_system
    test_data_consistency
    
    # 可选压力测试
    if [[ "$1" == "--stress" ]]; then
        stress_test
    fi
    
    echo "================================="
    echo -e "${GREEN}🎉 集成测试完成${NC}"
    
    generate_report
}

# 错误处理
trap 'echo -e "${RED}❌ 测试中断${NC}"; exit 1' ERR

# 执行主函数
main "$@"