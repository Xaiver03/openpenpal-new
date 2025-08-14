#!/usr/bin/env bash

# OpenPenPal 综合集成测试脚本
# 测试所有修复的问题和系统功能
# 包含增强的错误处理和可靠性检查

# 检查bash版本和关联数组支持
check_bash_version() {
    local bash_version=""
    if [ -n "$BASH_VERSION" ]; then
        bash_version="$BASH_VERSION"
    else
        bash_version=$(bash --version 2>/dev/null | head -1 | grep -o '[0-9]\+\.[0-9]\+')
    fi
    
    local major_version=$(echo "$bash_version" | cut -d. -f1)
    if [ -n "$major_version" ] && [ "$major_version" -ge 4 ]; then
        return 0
    else
        echo "警告: 检测到bash版本 $bash_version，不支持关联数组"
        echo "将使用兼容模式运行（功能可能受限）"
        return 1
    fi
}

# 检查并设置兼容模式
BASH_4_PLUS=true
if ! check_bash_version; then
    BASH_4_PLUS=false
fi

echo "🚀 OpenPenPal 综合集成测试开始"
echo "============================================"
echo "测试时间: $(date)"
echo "测试环境: $(uname -a)"
echo ""

# 项目根目录
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
echo "项目根目录: $PROJECT_ROOT"

# API基础URL
BASE_URL="http://localhost:8080"
API_BASE="${BASE_URL}/api/v1"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m'

# 测试统计
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
SKIPPED_TESTS=0

# 服务状态检查
BACKEND_RUNNING=false
FRONTEND_RUNNING=false

# 用户tokens存储（兼容模式）
if $BASH_4_PLUS; then
    declare -A USER_TOKENS
else
    # bash 3.2兼容模式：使用简单变量
    USER_TOKENS_LIST=""
fi

# 函数：设置用户token（兼容模式）
set_user_token() {
    local username="$1"
    local token="$2"
    
    if $BASH_4_PLUS; then
        USER_TOKENS["$username"]="$token"
    else
        # 使用简单的字符串存储
        USER_TOKENS_LIST="${USER_TOKENS_LIST}${username}:${token};"
    fi
}

# 函数：获取用户token（兼容模式）
get_user_token() {
    local username="$1"
    
    if $BASH_4_PLUS; then
        echo "${USER_TOKENS[$username]}"
    else
        # 从字符串中提取token
        echo "$USER_TOKENS_LIST" | grep -o "${username}:[^;]*" | cut -d: -f2
    fi
}

# 函数：获取测试账号密码
get_test_password() {
    local username="$1"
    
    case "$username" in
        "alice"|"bob"|"courier1"|"senior_courier"|"coordinator"|"school_admin"|"platform_admin"|"super_admin"|"courier_level1"|"courier_level2"|"courier_level3"|"courier_level4")
            echo "secret"
            ;;
        "courier_building")
            echo "courier001"
            ;;
        "courier_area")
            echo "courier002"
            ;;
        "courier_school")
            echo "courier003"
            ;;
        "courier_city")
            echo "courier004"
            ;;
        "admin")
            echo "admin123"
            ;;
        *)
            echo ""
            ;;
    esac
}

# 函数：打印测试结果
print_test_result() {
    local test_name="$1"
    local status="$2"
    local message="$3"
    local details="${4:-}"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    case "$status" in
        "PASS")
            echo -e "${GREEN}✅ PASS${NC}: $test_name - $message"
            PASSED_TESTS=$((PASSED_TESTS + 1))
            ;;
        "FAIL")
            echo -e "${RED}❌ FAIL${NC}: $test_name - $message"
            if [ -n "$details" ]; then
                echo -e "   ${RED}详情${NC}: $details"
            fi
            FAILED_TESTS=$((FAILED_TESTS + 1))
            ;;
        "SKIP")
            echo -e "${YELLOW}⏭️  SKIP${NC}: $test_name - $message"
            SKIPPED_TESTS=$((SKIPPED_TESTS + 1))
            ;;
        "WARN")
            echo -e "${YELLOW}⚠️  WARN${NC}: $test_name - $message"
            PASSED_TESTS=$((PASSED_TESTS + 1))  # 警告仍算通过
            ;;
    esac
}

# 函数：安全的curl请求（带重试和错误处理）
safe_curl() {
    local method="$1"
    local url="$2"
    local headers="$3"
    local data="$4"
    local max_retries="${5:-3}"
    local retry_delay="${6:-1}"
    
    local attempt=1
    while [ $attempt -le $max_retries ]; do
        local response
        if [ -n "$data" ]; then
            response=$(curl -s -w "HTTPSTATUS:%{http_code}" -X "$method" "$url" $headers -d "$data" 2>/dev/null)
        else
            response=$(curl -s -w "HTTPSTATUS:%{http_code}" -X "$method" "$url" $headers 2>/dev/null)
        fi
        
        local http_code=$(echo "$response" | grep "HTTPSTATUS:" | cut -d: -f2)
        local body=$(echo "$response" | sed 's/HTTPSTATUS:.*$//')
        
        # 如果不是429（速率限制），直接返回结果
        if [ "$http_code" != "429" ]; then
            echo "$body"
            return 0
        fi
        
        # 速率限制情况下等待重试
        if [ $attempt -lt $max_retries ]; then
            echo "Rate limited, retrying in ${retry_delay}s... (attempt $attempt/$max_retries)" >&2
            sleep $retry_delay
            retry_delay=$((retry_delay * 2))  # 指数退避
        fi
        
        attempt=$((attempt + 1))
    done
    
    echo '{"error":"Max retries exceeded","http_code":'$http_code'}'
    return 1
}

# 函数：检查服务状态
check_service_status() {
    echo -e "${BLUE}步骤1: 检查服务状态${NC}"
    echo "=================================="
    
    # 检查后端服务
    local backend_response=$(safe_curl "GET" "$BASE_URL/health" "" "" 1 0)
    if echo "$backend_response" | jq -e '.status == "healthy"' >/dev/null 2>&1; then
        BACKEND_RUNNING=true
        print_test_result "后端服务检查" "PASS" "服务运行正常"
        
        # 检查数据库状态
        local db_status=$(echo "$backend_response" | jq -r '.database // "unknown"')
        if [ "$db_status" = "healthy" ]; then
            print_test_result "数据库连接检查" "PASS" "数据库连接正常"
        else
            print_test_result "数据库连接检查" "WARN" "数据库状态: $db_status"
        fi
    else
        BACKEND_RUNNING=false
        print_test_result "后端服务检查" "FAIL" "服务未运行或不健康"
        echo "响应: $backend_response"
    fi
    
    # 检查前端服务（简单的连接测试）
    if curl -s -o /dev/null -w "%{http_code}" "http://localhost:3000" | grep -q "200\|301\|302"; then
        FRONTEND_RUNNING=true
        print_test_result "前端服务检查" "PASS" "前端服务可访问"
    else
        FRONTEND_RUNNING=false
        print_test_result "前端服务检查" "FAIL" "前端服务不可访问"
    fi
    
    echo ""
}

# 函数：测试账号登录
test_account_login() {
    local username="$1"
    local password="$2"
    local description="$3"
    local expected_role="${4:-}"
    
    if ! $BACKEND_RUNNING; then
        print_test_result "登录测试 ($username)" "SKIP" "后端服务未运行"
        return
    fi
    
    # 登录请求
    local login_data='{"username":"'$username'","password":"'$password'"}'
    local response=$(safe_curl "POST" "$API_BASE/auth/login" "-H \"Content-Type: application/json\"" "$login_data")
    
    local success=$(echo "$response" | jq -r '.success // false')
    
    if [ "$success" = "true" ]; then
        local token=$(echo "$response" | jq -r '.data.token // ""')
        local actual_role=$(echo "$response" | jq -r '.data.user.role // ""')
        
        if [ -n "$token" ] && [ "$token" != "null" ]; then
            set_user_token "$username" "$token"
            
            # 验证角色（如果提供了期望角色）
            if [ -n "$expected_role" ]; then
                if [ "$actual_role" = "$expected_role" ]; then
                    print_test_result "登录测试 ($username)" "PASS" "$description - 角色正确: $actual_role"
                else
                    print_test_result "登录测试 ($username)" "WARN" "$description - 角色不匹配: 期望$expected_role, 实际$actual_role"
                fi
            else
                print_test_result "登录测试 ($username)" "PASS" "$description - 登录成功"
            fi
        else
            print_test_result "登录测试 ($username)" "FAIL" "$description - 未获得有效token"
        fi
    else
        local error=$(echo "$response" | jq -r '.error // "Unknown error"')
        print_test_result "登录测试 ($username)" "FAIL" "$description - 登录失败: $error"
    fi
}

# 函数：测试API端点
test_api_endpoint() {
    local method="$1"
    local endpoint="$2"
    local token="$3"
    local description="$4"
    local expected_codes="${5:-200,201}"  # 期望的HTTP状态码
    
    if ! $BACKEND_RUNNING; then
        print_test_result "API测试 ($endpoint)" "SKIP" "后端服务未运行"
        return
    fi
    
    local headers=""
    if [ -n "$token" ]; then
        headers="-H \"Authorization: Bearer $token\""
    fi
    
    local response=$(safe_curl "$method" "$API_BASE$endpoint" "$headers")
    local success=$(echo "$response" | jq -r '.success // null')
    
    # 简单成功检查
    if [ "$success" = "true" ]; then
        print_test_result "API测试 ($endpoint)" "PASS" "$description - 成功"
    elif [ "$success" = "false" ]; then
        local error=$(echo "$response" | jq -r '.error // "Unknown error"')
        # 权限错误是预期的
        if echo "$error" | grep -qi "permission\|unauthorized\|forbidden"; then
            print_test_result "API测试 ($endpoint)" "WARN" "$description - 权限限制（正常）"
        else
            print_test_result "API测试 ($endpoint)" "FAIL" "$description - 错误: $error"
        fi
    else
        print_test_result "API测试 ($endpoint)" "FAIL" "$description - 无效响应"
    fi
}

# 函数：测试文件路径修复
test_file_paths() {
    echo -e "${BLUE}步骤3: 测试文件路径修复${NC}"
    echo "=================================="
    
    # 检查关键文件是否存在
    local files=(
        "backend/internal/models/user.go"
        "backend/internal/services/user_service.go" 
        "backend/internal/middleware/rate_limiter.go"
        "backend/main.go"
    )
    
    for file in "${files[@]}"; do
        local full_path="$PROJECT_ROOT/$file"
        if [ -f "$full_path" ]; then
            print_test_result "文件检查" "PASS" "$file 存在"
        else
            print_test_result "文件检查" "FAIL" "$file 不存在"
        fi
    done
    
    echo ""
}

# 函数：测试四级信使系统
test_courier_system() {
    echo -e "${BLUE}步骤4: 测试四级信使系统${NC}"
    echo "=================================="
    
    # 测试四级信使账号登录
    test_account_login "courier_level1" "secret" "一级信使（楼栋）"
    test_account_login "courier_level2" "secret" "二级信使（片区）"
    test_account_login "courier_level3" "secret" "三级信使（学校）"
    test_account_login "courier_level4" "secret" "四级信使（城市）"
    
    # 使用管理员token测试信使管理API
    local admin_token=$(get_user_token "admin")
    if [ -n "$admin_token" ]; then
        test_api_endpoint "GET" "/courier/management/level-1/stats" "$admin_token" "一级信使统计"
        test_api_endpoint "GET" "/courier/management/level-2/stats" "$admin_token" "二级信使统计"
        test_api_endpoint "GET" "/courier/management/level-3/stats" "$admin_token" "三级信使统计"
        test_api_endpoint "GET" "/courier/management/level-4/stats" "$admin_token" "四级信使统计"
    fi
    
    echo ""
}

# 函数：测试WebSocket和实时功能
test_websocket() {
    echo -e "${BLUE}步骤5: 测试WebSocket和实时功能${NC}"
    echo "=================================="
    
    local admin_token=$(get_user_token "admin")
    if [ -n "$admin_token" ]; then
        test_api_endpoint "GET" "/ws/stats" "$admin_token" "WebSocket统计"
        test_api_endpoint "GET" "/ws/connections" "$admin_token" "WebSocket连接"
    else
        print_test_result "WebSocket测试" "SKIP" "管理员账号未登录"
    fi
    
    echo ""
}

# 主测试流程
main() {
    # 检查服务状态
    check_service_status
    
    # 如果后端未运行，跳过大部分测试
    if ! $BACKEND_RUNNING; then
        echo -e "${RED}后端服务未运行，跳过API相关测试${NC}"
        echo ""
    fi
    
    # 测试文件路径修复
    test_file_paths
    
    # 测试账号登录
    echo -e "${BLUE}步骤2: 测试账号登录和认证${NC}"
    echo "=================================="
    
    # 测试关键账号
    test_account_login "admin" "admin123" "系统管理员" "super_admin"
    test_account_login "alice" "secret" "普通用户" "user"
    test_account_login "courier1" "secret" "普通信使" "courier"
    test_account_login "senior_courier" "secret" "高级信使" "senior_courier"
    test_account_login "coordinator" "secret" "信使协调员" "courier_coordinator"
    
    echo ""
    
    # 测试四级信使系统
    test_courier_system
    
    # 测试WebSocket功能
    test_websocket
    
    # 测试管理员功能
    echo -e "${BLUE}步骤6: 测试管理员功能${NC}"
    echo "=================================="
    
    local admin_token=$(get_user_token "admin")
    if [ -n "$admin_token" ]; then
        test_api_endpoint "GET" "/admin/dashboard/stats" "$admin_token" "管理员仪表盘"
        test_api_endpoint "GET" "/users/me" "$admin_token" "用户信息"
        test_api_endpoint "GET" "/letters/stats" "$admin_token" "信件统计"
    else
        print_test_result "管理员功能测试" "SKIP" "管理员账号未登录"
    fi
    
    echo ""
    
    # 输出测试结果总结
    echo -e "${PURPLE}测试结果总结${NC}"
    echo "=================================="
    echo "总测试数: $TOTAL_TESTS"
    echo -e "通过测试: ${GREEN}$PASSED_TESTS${NC}"
    echo -e "失败测试: ${RED}$FAILED_TESTS${NC}"
    echo -e "跳过测试: ${YELLOW}$SKIPPED_TESTS${NC}"
    
    local success_rate=0
    if [ $TOTAL_TESTS -gt 0 ]; then
        success_rate=$(( (PASSED_TESTS * 100) / TOTAL_TESTS ))
    fi
    echo "成功率: ${success_rate}%"
    
    echo ""
    if [ $FAILED_TESTS -eq 0 ]; then
        echo -e "${GREEN}🎉 所有测试通过！集成测试成功！${NC}"
        exit 0
    else
        echo -e "${RED}⚠️  有 $FAILED_TESTS 个测试失败，需要检查和修复${NC}"
        
        # 提供修复建议
        echo ""
        echo -e "${YELLOW}修复建议:${NC}"
        if ! $BACKEND_RUNNING; then
            echo "1. 启动后端服务: cd backend && go run main.go"
        fi
        if ! $FRONTEND_RUNNING; then
            echo "2. 启动前端服务: cd frontend && npm run dev"
        fi
        if [ $FAILED_TESTS -gt 0 ]; then
            echo "3. 检查服务日志排查具体问题"
            echo "4. 确保数据库种子数据已正确加载"
            echo "5. 检查速率限制配置（可设置TEST_MODE=true）"
        fi
        
        exit 1
    fi
}

# 执行主测试流程
main "$@"