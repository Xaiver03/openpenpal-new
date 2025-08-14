#!/usr/bin/env bash

# OpenPenPal 角色权限详细测试脚本
# 测试每个角色的具体权限是否正确工作

# 确保使用bash并检查关联数组支持
if ! declare -A test_array 2>/dev/null; then
    echo "错误: 此脚本需要bash 4.0或更高版本来支持关联数组"
    echo "当前shell: $0"
    echo "请使用: bash $0"
    exit 1
fi

echo "🎭 开始OpenPenPal角色权限详细测试..."
echo "============================================"

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
TOTAL_PERMISSION_TESTS=0
PASSED_PERMISSION_TESTS=0
FAILED_PERMISSION_TESTS=0

# 用户tokens - 使用关联数组前先声明
declare -A USER_TOKENS

# 权限到API端点的映射 - 使用关联数组前先声明
declare -A PERMISSION_ENDPOINTS
PERMISSION_ENDPOINTS["write_letter"]="POST:/letters"
PERMISSION_ENDPOINTS["read_letter"]="GET:/letters"
PERMISSION_ENDPOINTS["manage_profile"]="PUT:/users/me"
PERMISSION_ENDPOINTS["deliver_letter"]="POST:/courier/letters/TEST123/status"
PERMISSION_ENDPOINTS["scan_code"]="GET:/letters/read/TEST123"
PERMISSION_ENDPOINTS["view_tasks"]="GET:/courier/status"
PERMISSION_ENDPOINTS["manage_couriers"]="GET:/admin/courier/applications"
PERMISSION_ENDPOINTS["assign_tasks"]="POST:/admin/courier/1/approve"
PERMISSION_ENDPOINTS["view_reports"]="GET:/letters/stats"
PERMISSION_ENDPOINTS["manage_users"]="GET:/admin/users/1"
PERMISSION_ENDPOINTS["manage_school"]="GET:/admin/users/1"
PERMISSION_ENDPOINTS["view_analytics"]="GET:/letters/stats"
PERMISSION_ENDPOINTS["manage_system"]="GET:/admin/courier/applications"

# 角色权限映射（从models/user.go复制）
declare -A ROLE_HAS_PERMISSIONS
ROLE_HAS_PERMISSIONS["user"]="write_letter,read_letter,manage_profile"
ROLE_HAS_PERMISSIONS["courier"]="write_letter,read_letter,manage_profile,deliver_letter,scan_code,view_tasks"
ROLE_HAS_PERMISSIONS["senior_courier"]="write_letter,read_letter,manage_profile,deliver_letter,scan_code,view_tasks,view_reports"
ROLE_HAS_PERMISSIONS["courier_coordinator"]="write_letter,read_letter,manage_profile,deliver_letter,scan_code,view_tasks,manage_couriers,assign_tasks,view_reports"
ROLE_HAS_PERMISSIONS["school_admin"]="write_letter,read_letter,manage_profile,manage_users,manage_couriers,assign_tasks,view_reports,manage_school,view_analytics"
ROLE_HAS_PERMISSIONS["platform_admin"]="write_letter,read_letter,manage_profile,manage_users,manage_couriers,assign_tasks,view_reports,manage_school,view_analytics,manage_system"
ROLE_HAS_PERMISSIONS["super_admin"]="write_letter,read_letter,manage_profile,manage_users,manage_couriers,assign_tasks,view_reports,manage_school,view_analytics,manage_system,manage_platform,manage_admins,system_config"

# 函数：登录用户并获取token
login_and_get_token() {
    local username="$1"
    local password="$2"
    
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
        echo "$token"
    else
        echo ""
    fi
}

# 函数：测试权限
test_permission() {
    local role="$1"
    local permission="$2"
    local should_have="$3"  # true/false
    local token="$4"
    
    TOTAL_PERMISSION_TESTS=$((TOTAL_PERMISSION_TESTS + 1))
    
    # 获取对应的API端点
    local endpoint_info="${PERMISSION_ENDPOINTS[$permission]}"
    if [[ -z "$endpoint_info" ]]; then
        echo -e "   ⚠️ ${YELLOW}权限 $permission 没有对应的测试端点${NC}"
        return
    fi
    
    IFS=':' read -r method endpoint <<< "$endpoint_info"
    
    echo -e "${CYAN}🧪 测试权限: $permission (角色: $role)${NC}"
    
    local curl_args=("-s" "-w" "\n%{http_code}" "-X" "$method")
    
    if [[ -n "$token" ]]; then
        curl_args+=("-H" "Authorization: Bearer $token")
    fi
    
    # 为某些端点添加测试数据
    case "$endpoint" in
        "POST:/letters")
            curl_args+=("-H" "Content-Type: application/json" "-d" '{"content":"test","recipient_id":"123"}')
            ;;
        "POST:/courier/letters/TEST123/status")
            curl_args+=("-H" "Content-Type: application/json" "-d" '{"status":"delivered"}')
            ;;
        "PUT:/users/me")
            curl_args+=("-H" "Content-Type: application/json" "-d" '{"nickname":"test"}')
            ;;
        "POST:/admin/courier/1/approve")
            curl_args+=("-H" "Content-Type: application/json" "-d" '{}')
            ;;
    esac
    
    curl_args+=("${API_BASE}${endpoint}")
    
    response=$(curl "${curl_args[@]}" 2>&1)
    response_body=$(echo "$response" | sed '$d')
    http_code=$(echo "$response" | tail -n 1)
    
    # 判断结果
    local test_passed=false
    
    if [[ "$should_have" == "true" ]]; then
        # 应该有权限 - 不应该返回403
        if [[ "$http_code" != "403" ]]; then
            test_passed=true
        fi
    else
        # 不应该有权限 - 应该返回403或401
        if [[ "$http_code" == "403" || "$http_code" == "401" ]]; then
            test_passed=true
        fi
    fi
    
    if [[ "$test_passed" == "true" ]]; then
        echo -e "   ✅ ${GREEN}权限测试通过${NC} (HTTP $http_code)"
        PASSED_PERMISSION_TESTS=$((PASSED_PERMISSION_TESTS + 1))
    else
        echo -e "   ❌ ${RED}权限测试失败${NC} (HTTP $http_code)"
        echo -e "   📋 预期: 角色 $role ${should_have} 有权限 $permission"
        FAILED_PERMISSION_TESTS=$((FAILED_PERMISSION_TESTS + 1))
    fi
    
    echo ""
}

# 函数：测试角色的所有权限
test_role_permissions() {
    local role="$1"
    local token="$2"
    
    echo -e "${BLUE}🎭 测试角色: $role${NC}"
    echo "----------------------------------------"
    
    # 获取该角色应该拥有的权限
    local role_permissions="${ROLE_HAS_PERMISSIONS[$role]}"
    IFS=',' read -ra HAS_PERMS <<< "$role_permissions"
    
    # 所有可能的权限
    local all_permissions=(
        "write_letter" "read_letter" "manage_profile"
        "deliver_letter" "scan_code" "view_tasks"
        "manage_couriers" "assign_tasks" "view_reports"
        "manage_users" "manage_school" "view_analytics"
        "manage_system"
    )
    
    # 测试每个权限
    for permission in "${all_permissions[@]}"; do
        local should_have="false"
        
        # 检查该角色是否应该有这个权限
        for has_perm in "${HAS_PERMS[@]}"; do
            if [[ "$has_perm" == "$permission" ]]; then
                should_have="true"
                break
            fi
        done
        
        test_permission "$role" "$permission" "$should_have" "$token"
    done
    
    echo ""
}

echo "📋 权限测试配置:"
echo "   • 测试目标: 验证每个角色的权限控制是否正确"
echo "   • 测试方法: 检查API端点的访问权限"
echo "   • 权限映射: 基于后端models/user.go中的定义"
echo ""

echo -e "${YELLOW}📝 第1步: 准备测试用户${NC}"
echo "=========================================="

# 尝试登录之前创建的测试用户
declare -a test_roles=("user" "courier" "senior_courier" "courier_coordinator" "school_admin" "platform_admin" "super_admin")

for role in "${test_roles[@]}"; do
    if [[ "$role" == "user" ]]; then
        # 使用之前注册的普通用户
        token=$(login_and_get_token "testuser02" "password123")
    else
        # 使用管理员测试脚本创建的用户
        token=$(login_and_get_token "admin_${role}" "password123")
    fi
    
    if [[ -n "$token" ]]; then
        USER_TOKENS["$role"]="$token"
        echo -e "   ✅ ${GREEN}$role 用户登录成功${NC}"
    else
        echo -e "   ❌ ${RED}$role 用户登录失败${NC}"
    fi
done

echo ""

echo -e "${YELLOW}📝 第2步: 执行权限测试${NC}"
echo "=========================================="

# 为每个角色执行权限测试
for role in "${test_roles[@]}"; do
    if [[ -n "${USER_TOKENS[$role]}" ]]; then
        test_role_permissions "$role" "${USER_TOKENS[$role]}"
    else
        echo -e "${RED}⚠️ 跳过 $role 角色测试（用户未登录）${NC}"
        echo ""
    fi
done

echo "=========================================="
echo -e "${YELLOW}📊 权限测试结果统计${NC}"
echo "=========================================="
echo -e "总权限测试:  ${BLUE}$TOTAL_PERMISSION_TESTS${NC}"
echo -e "通过测试:    ${GREEN}$PASSED_PERMISSION_TESTS${NC}"
echo -e "失败测试:    ${RED}$FAILED_PERMISSION_TESTS${NC}"

if [[ $TOTAL_PERMISSION_TESTS -gt 0 ]]; then
    permission_success_rate=$((PASSED_PERMISSION_TESTS * 100 / TOTAL_PERMISSION_TESTS))
    echo -e "权限准确率:  ${GREEN}${permission_success_rate}%${NC}"
else
    permission_success_rate=0
    echo -e "权限准确率:  ${RED}无法计算${NC}"
fi

echo ""
echo -e "${YELLOW}🔍 权限系统分析${NC}"
echo "=========================================="

if [[ $permission_success_rate -ge 95 ]]; then
    echo -e "${GREEN}✅ 权限系统精确运行${NC}"
    echo -e "   • 所有角色权限控制准确"
    echo -e "   • 权限边界清晰"
    echo -e "   • 系统安全性良好"
elif [[ $permission_success_rate -ge 80 ]]; then
    echo -e "${YELLOW}⚠️ 权限系统基本准确${NC}"
    echo -e "   • 大部分权限控制正确"
    echo -e "   • 存在少量权限配置问题"
    echo -e "   • 建议检查失败项目"
elif [[ $permission_success_rate -ge 60 ]]; then
    echo -e "${YELLOW}⚠️ 权限系统存在问题${NC}"
    echo -e "   • 多项权限配置不正确"
    echo -e "   • 可能存在安全风险"
    echo -e "   • 需要修复权限控制"
else
    echo -e "${RED}❌ 权限系统严重故障${NC}"
    echo -e "   • 权限控制大量失效"
    echo -e "   • 存在严重安全风险"
    echo -e "   • 需要立即修复"
fi

echo ""
echo -e "${CYAN}📋 角色权限矩阵验证${NC}"
echo "=========================================="
echo "基础权限 (所有角色都应该有):"
echo "   • write_letter (写信)"
echo "   • read_letter (读信)"  
echo "   • manage_profile (管理个人资料)"
echo ""
echo "信使权限 (courier及以上):"
echo "   • deliver_letter (配送信件)"
echo "   • scan_code (扫描二维码)"
echo "   • view_tasks (查看任务)"
echo ""
echo "协调员权限 (courier_coordinator及以上):"
echo "   • manage_couriers (管理信使)"
echo "   • assign_tasks (分配任务)"
echo "   • view_reports (查看报告)"
echo ""
echo "管理员权限 (school_admin及以上):"
echo "   • manage_users (管理用户)"
echo "   • manage_school (管理学校)"
echo "   • view_analytics (查看分析)"

echo ""
echo -e "${YELLOW}🏁 角色权限详细测试完成！${NC}"