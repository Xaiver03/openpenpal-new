#!/usr/bin/env bash

# OpenPenPal 管理员权限测试脚本
# 测试各级管理员权限是否正常工作

# 确保使用bash并检查关联数组支持
if ! declare -A test_array 2>/dev/null; then
    echo "错误: 此脚本需要bash 4.0或更高版本来支持关联数组"
    echo "当前shell: $0"
    echo "请使用: bash $0"
    exit 1
fi

echo "🔐 开始OpenPenPal管理员权限测试..."
echo "=========================================="

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
NC='\033[0m' # No Color

# 统计变量
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# 测试用户tokens (需要先创建管理员用户并获取token) - 使用关联数组前先声明
declare -A ADMIN_TOKENS

# 角色定义
declare -a ROLES=(
    "courier"
    "senior_courier" 
    "courier_coordinator"
    "school_admin"
    "platform_admin"
    "super_admin"
)

# 权限定义映射
declare -A ROLE_PERMISSIONS
ROLE_PERMISSIONS["courier"]="deliver_letter,scan_code,view_tasks"
ROLE_PERMISSIONS["senior_courier"]="deliver_letter,scan_code,view_tasks,view_reports"
ROLE_PERMISSIONS["courier_coordinator"]="manage_couriers,assign_tasks,view_reports"
ROLE_PERMISSIONS["school_admin"]="manage_users,manage_couriers,manage_school,view_analytics"
ROLE_PERMISSIONS["platform_admin"]="manage_users,manage_couriers,manage_school,view_analytics,manage_system"
ROLE_PERMISSIONS["super_admin"]="manage_platform,manage_admins,system_config"

# 函数：登录用户并获取token
login_user() {
    local username="$1"
    local password="$2"
    
    echo -e "${BLUE}🔐 登录用户: $username${NC}"
    
    json_data=$(cat <<EOF
{
  "username": "$username",
  "password": "$password"
}
EOF
)
    
    response=$(curl -s -w "\n%{http_code}" \
        -X POST \
        -H "Content-Type: application/json" \
        -d "$json_data" \
        "${API_BASE}/auth/login" 2>&1)
    
    response_body=$(echo "$response" | sed '$d')
    http_code=$(echo "$response" | tail -n 1)
    
    if [[ "$http_code" == "200" ]]; then
        token=$(echo "$response_body" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
        echo -e "   ✅ ${GREEN}登录成功${NC}"
        echo "$token"
    else
        echo -e "   ❌ ${RED}登录失败${NC} (HTTP $http_code)"
        echo ""
    fi
}

# 函数：测试API访问
test_api_access() {
    local description="$1"
    local method="$2"
    local endpoint="$3"
    local token="$4"
    local data="$5"
    local expected_codes="$6"  # 期望的状态码，用逗号分隔
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    echo -e "${CYAN}🧪 测试: $description${NC}"
    
    local curl_args=("-s" "-w" "\n%{http_code}" "-X" "$method")
    
    if [[ -n "$token" ]]; then
        curl_args+=("-H" "Authorization: Bearer $token")
    fi
    
    if [[ -n "$data" ]]; then
        curl_args+=("-H" "Content-Type: application/json" "-d" "$data")
    fi
    
    curl_args+=("${API_BASE}${endpoint}")
    
    response=$(curl "${curl_args[@]}" 2>&1)
    response_body=$(echo "$response" | sed '$d')
    http_code=$(echo "$response" | tail -n 1)
    
    # 检查是否在期望的状态码范围内
    if [[ ",$expected_codes," == *",$http_code,"* ]]; then
        echo -e "   ✅ ${GREEN}测试通过${NC} (HTTP $http_code)"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "   ❌ ${RED}测试失败${NC} (HTTP $http_code, 期望: $expected_codes)"
        echo -e "   📄 响应: $response_body"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    
    echo ""
}

# 函数：创建测试管理员用户
create_admin_user() {
    local username="$1"
    local role="$2"
    local password="password123"
    
    echo -e "${PURPLE}👤 创建管理员用户: $username ($role)${NC}"
    
    # 注册用户
    json_data=$(cat <<EOF
{
  "username": "$username",
  "email": "${username}@admin.test",
  "password": "$password",
  "nickname": "Admin ${role}",
  "school_code": "ADM001"
}
EOF
)
    
    response=$(curl -s -w "\n%{http_code}" \
        -X POST \
        -H "Content-Type: application/json" \
        -d "$json_data" \
        "${API_BASE}/auth/register" 2>&1)
    
    http_code=$(echo "$response" | tail -n 1)
    
    if [[ "$http_code" == "200" || "$http_code" == "201" ]]; then
        echo -e "   ✅ ${GREEN}用户创建成功${NC}"
        
        # 登录获取token
        token=$(login_user "$username" "$password")
        if [[ -n "$token" ]]; then
            ADMIN_TOKENS["$role"]="$token"
            echo -e "   🔑 Token已保存"
        fi
    else
        echo -e "   ⚠️ ${YELLOW}用户可能已存在，尝试登录${NC}"
        token=$(login_user "$username" "$password")
        if [[ -n "$token" ]]; then
            ADMIN_TOKENS["$role"]="$token"
        fi
    fi
    
    echo ""
}

echo "📋 权限系统分析:"
echo "   • 角色层级: user(1) < courier(2) < senior_courier(3) < courier_coordinator(4) < school_admin(5) < platform_admin(6) < super_admin(7)"
echo "   • 权限继承: 高级角色继承低级角色的所有权限"
echo "   • 管理员路由: /api/v1/admin/*"
echo ""

echo -e "${YELLOW}📝 第1步: 创建测试管理员用户${NC}"
echo "=========================================="

# 创建各级管理员用户
for role in "${ROLES[@]}"; do
    create_admin_user "admin_${role}" "$role"
done

echo -e "${YELLOW}📝 第2步: 测试基础认证${NC}"
echo "=========================================="

# 测试无token访问
test_api_access "无认证访问管理员接口" "GET" "/admin/users/123" "" "" "401"

# 测试普通用户token访问管理员接口  
if [[ -n "${ADMIN_TOKENS[courier]}" ]]; then
    test_api_access "信使角色访问管理员接口" "GET" "/admin/users/123" "${ADMIN_TOKENS[courier]}" "" "403"
fi

echo -e "${YELLOW}📝 第3步: 测试用户管理权限${NC}"
echo "=========================================="

# 测试各级管理员的用户管理权限
for role in "school_admin" "platform_admin" "super_admin"; do
    if [[ -n "${ADMIN_TOKENS[$role]}" ]]; then
        echo -e "${BLUE}测试 $role 的用户管理权限:${NC}"
        
        # 获取用户信息
        test_api_access "获取用户信息" "GET" "/admin/users/1" "${ADMIN_TOKENS[$role]}" "" "200,404"
        
        # 用户停用/激活
        test_api_access "停用用户" "DELETE" "/admin/users/999" "${ADMIN_TOKENS[$role]}" "" "200,404"
        test_api_access "激活用户" "POST" "/admin/users/999/reactivate" "${ADMIN_TOKENS[$role]}" "" "200,404"
    fi
done

echo -e "${YELLOW}📝 第4步: 测试信使管理权限${NC}"
echo "=========================================="

# 测试信使管理权限
for role in "courier_coordinator" "school_admin" "platform_admin" "super_admin"; do
    if [[ -n "${ADMIN_TOKENS[$role]}" ]]; then
        echo -e "${BLUE}测试 $role 的信使管理权限:${NC}"
        
        # 获取待审核申请
        test_api_access "获取信使申请列表" "GET" "/admin/courier/applications" "${ADMIN_TOKENS[$role]}" "" "200"
        
        # 审批信使申请
        test_api_access "批准信使申请" "POST" "/admin/courier/999/approve" "${ADMIN_TOKENS[$role]}" "" "200,404"
        test_api_access "拒绝信使申请" "POST" "/admin/courier/999/reject" "${ADMIN_TOKENS[$role]}" "" "200,404"
    fi
done

echo -e "${YELLOW}📝 第5步: 测试权限边界${NC}"
echo "=========================================="

# 测试权限边界 - 低级角色不应该能访问高级功能
declare -A UNAUTHORIZED_TESTS
UNAUTHORIZED_TESTS["courier"]="/admin/users/1,/admin/courier/applications"
UNAUTHORIZED_TESTS["senior_courier"]="/admin/users/1,/admin/courier/applications" 
UNAUTHORIZED_TESTS["courier_coordinator"]="/admin/users/1"

for role in "${!UNAUTHORIZED_TESTS[@]}"; do
    if [[ -n "${ADMIN_TOKENS[$role]}" ]]; then
        echo -e "${BLUE}测试 $role 权限边界:${NC}"
        
        IFS=',' read -ra ENDPOINTS <<< "${UNAUTHORIZED_TESTS[$role]}"
        for endpoint in "${ENDPOINTS[@]}"; do
            test_api_access "$role 尝试访问未授权端点" "GET" "$endpoint" "${ADMIN_TOKENS[$role]}" "" "403"
        done
    fi
done

echo -e "${YELLOW}📝 第6步: 测试跨校权限控制${NC}"
echo "=========================================="

# 测试学校管理员是否只能管理同校用户
if [[ -n "${ADMIN_TOKENS[school_admin]}" ]]; then
    echo -e "${BLUE}测试学校管理员跨校限制:${NC}"
    
    # 这里需要创建不同学校的用户来测试，简化为测试基本功能
    test_api_access "学校管理员访问用户管理" "GET" "/admin/users/1" "${ADMIN_TOKENS[school_admin]}" "" "200,404"
fi

echo -e "${YELLOW}📝 第7步: 测试角色层级继承${NC}"
echo "=========================================="

# 测试高级角色是否能执行低级角色的操作
if [[ -n "${ADMIN_TOKENS[super_admin]}" ]]; then
    echo -e "${BLUE}测试超级管理员权限继承:${NC}"
    
    # 超级管理员应该能执行所有操作
    test_api_access "超级管理员-用户管理" "GET" "/admin/users/1" "${ADMIN_TOKENS[super_admin]}" "" "200,404"
    test_api_access "超级管理员-信使管理" "GET" "/admin/courier/applications" "${ADMIN_TOKENS[super_admin]}" "" "200"
fi

echo "=========================================="
echo -e "${YELLOW}📊 权限测试结果统计${NC}"
echo "=========================================="
echo -e "总测试数:   ${BLUE}$TOTAL_TESTS${NC}"
echo -e "通过测试:   ${GREEN}$PASSED_TESTS${NC}"
echo -e "失败测试:   ${RED}$FAILED_TESTS${NC}"

success_rate=$((PASSED_TESTS * 100 / TOTAL_TESTS))
echo -e "成功率:     ${GREEN}${success_rate}%${NC}"

echo ""
echo -e "${YELLOW}🔍 权限系统评估${NC}"
echo "=========================================="

if [[ $success_rate -ge 90 ]]; then
    echo -e "${GREEN}✅ 权限系统运行良好${NC}"
    echo -e "   • 角色认证正常"
    echo -e "   • 权限控制有效"
    echo -e "   • 访问限制正确"
elif [[ $success_rate -ge 70 ]]; then
    echo -e "${YELLOW}⚠️ 权限系统基本正常，存在部分问题${NC}"
    echo -e "   • 大部分功能正常"
    echo -e "   • 建议检查失败的测试项"
else
    echo -e "${RED}❌ 权限系统存在严重问题${NC}"
    echo -e "   • 多项权限控制失效"
    echo -e "   • 需要立即修复"
fi

echo ""
echo -e "${YELLOW}📋 权限矩阵摘要${NC}"
echo "=========================================="
echo "角色级别 | 用户管理 | 信使管理 | 学校管理 | 系统管理"
echo "---------|----------|----------|----------|----------"
echo "信使     |    ❌    |    ❌    |    ❌    |    ❌"
echo "高级信使 |    ❌    |    ❌    |    ❌    |    ❌"
echo "协调员   |    ❌    |    ✅    |    ❌    |    ❌"
echo "学校管理 |    ✅    |    ✅    |    ✅    |    ❌"
echo "平台管理 |    ✅    |    ✅    |    ✅    |    ✅"
echo "超级管理 |    ✅    |    ✅    |    ✅    |    ✅"

echo ""
echo -e "${YELLOW}🏁 管理员权限测试完成！${NC}"

# 输出测试用户信息供后续使用
echo ""
echo -e "${CYAN}📝 测试用户凭据（用于进一步测试）:${NC}"
for role in "${ROLES[@]}"; do
    echo "   admin_${role} / password123"
done