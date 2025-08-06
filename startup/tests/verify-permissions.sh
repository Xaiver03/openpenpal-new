#!/bin/bash

echo "=== OpenPenPal 权限验证测试 ==="
echo "测试时间: $(date)"
echo "========================================="

# 测试账号和预期权限
declare -A test_accounts=(
    ["alice"]="1:用户:read,write"
    ["courier1"]="2:一级信使:read,write,deliver_letter,scan_code,view_tasks"
    ["senior_courier"]="3:二级信使:read,write,deliver_letter,scan_code,view_tasks,manage_couriers,assign_tasks"
    ["coordinator"]="4:三级信使:read,write,deliver_letter,scan_code,view_tasks,manage_couriers,assign_tasks,view_reports,coordinate_school"
    ["school_admin"]="5:学校管理员:read,write,deliver_letter,scan_code,view_tasks,manage_couriers,assign_tasks,view_reports,manage_users,manage_school,view_analytics"
    ["platform_admin"]="6:平台管理员:read,write,deliver_letter,scan_code,view_tasks,manage_couriers,assign_tasks,view_reports,manage_users,manage_school,view_analytics,manage_platform,cross_school_management"
    ["super_admin"]="7:超级管理员:read,write,deliver_letter,scan_code,view_tasks,manage_couriers,assign_tasks,view_reports,manage_users,manage_school,view_analytics,manage_platform,cross_school_management,system_config,audit_submissions,handle_reports"
)

# 测试API端点和所需权限
declare -A test_apis=(
    ["GET:/api/letters"]="read:信件查看"
    ["POST:/api/letters"]="write:信件创建"
    ["GET:/api/users"]="manage_users:用户管理"
    ["GET:/api/system/config"]="system_config:系统配置"
    ["GET:8002/api/courier/tasks"]="view_tasks:信使任务查看"
    ["GET:8002/api/courier/manage"]="manage_couriers:信使管理"
    ["GET:8002/api/courier/rankings"]="view_tasks:积分排行榜"
)

# 禁用代理
unset http_proxy
unset https_proxy

test_api() {
    local method="$1"
    local url="$2"
    local token="$3"
    local expected_status="$4"
    
    response=$(curl -s -o /dev/null -w "%{http_code}" -X "$method" -H "Authorization: Bearer $token" "$url")
    
    if [ "$response" = "$expected_status" ]; then
        echo "  ✅ $method $url → $response"
        return 0
    else
        echo "  ❌ $method $url → $response (期望: $expected_status)"
        return 1
    fi
}

# 开始测试每个账号
for username in "${!test_accounts[@]}"; do
    IFS=":" read -r level role_name permissions <<< "${test_accounts[$username]}"
    
    echo
    echo "🔍 测试账号: $username ($role_name - 权限等级$level)"
    echo "--------------------------------"
    
    # 登录获取token
    login_response=$(curl -s -X POST http://localhost:8001/auth/login \
        -H "Content-Type: application/json" \
        -d "{\"username\":\"$username\",\"password\":\"secret\"}")
    
    if echo "$login_response" | grep -q "success.*true"; then
        token=$(echo "$login_response" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
        echo "  🟢 登录成功"
        echo "  📝 Token: ${token:0:20}..."
        
        # 解析用户权限
        user_permissions=$(echo "$login_response" | grep -o '"permissions":\[[^]]*\]' | sed 's/"permissions":\[//; s/\]//; s/"//g')
        echo "  🔑 权限列表: $user_permissions"
        
        # 测试基础API权限
        echo "  📊 API权限测试:"
        
        # 测试信件读取权限
        if echo "$user_permissions" | grep -q "read"; then
            test_api "GET" "http://localhost:8001/api/letters" "$token" "200"
        else
            test_api "GET" "http://localhost:8001/api/letters" "$token" "403"
        fi
        
        # 测试信件创建权限
        if echo "$user_permissions" | grep -q "write"; then
            test_api "POST" "http://localhost:8001/api/letters" "$token" "200"
        else
            test_api "POST" "http://localhost:8001/api/letters" "$token" "403"
        fi
        
        # 测试用户管理权限
        if echo "$user_permissions" | grep -q "manage_users"; then
            test_api "GET" "http://localhost:8001/api/users" "$token" "200"
        else
            test_api "GET" "http://localhost:8001/api/users" "$token" "403"
        fi
        
        # 测试系统配置权限
        if echo "$user_permissions" | grep -q "system_config"; then
            test_api "GET" "http://localhost:8001/api/system/config" "$token" "200"
        else
            test_api "GET" "http://localhost:8001/api/system/config" "$token" "403"
        fi
        
        # 测试信使权限（如果有的话）
        if echo "$user_permissions" | grep -q "view_tasks"; then
            test_api "GET" "http://localhost:8002/api/courier/tasks" "$token" "200"
            test_api "GET" "http://localhost:8002/api/courier/rankings" "$token" "200"
        else
            test_api "GET" "http://localhost:8002/api/courier/tasks" "$token" "403"
            test_api "GET" "http://localhost:8002/api/courier/rankings" "$token" "403"
        fi
        
        # 测试信使管理权限（如果有的话）
        if echo "$user_permissions" | grep -q "manage_couriers"; then
            test_api "GET" "http://localhost:8002/api/courier/manage" "$token" "200"
        else
            test_api "GET" "http://localhost:8002/api/courier/manage" "$token" "403"
        fi
        
    else
        echo "  ❌ 登录失败"
        echo "  📄 响应: $login_response"
    fi
done

echo
echo "========================================="
echo "✅ 权限验证测试完成"
echo "测试说明:"
echo "  ✅ = 权限验证正确"
echo "  ❌ = 权限验证异常"
echo "  🟢 = 功能正常"
echo "  ❌ = 功能异常"