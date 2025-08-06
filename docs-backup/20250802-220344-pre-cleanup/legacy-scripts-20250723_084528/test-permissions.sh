#!/bin/bash

echo "=== OpenPenPal 权限验证测试 ==="
echo "测试时间: $(date)"
echo "========================================="

# 禁用代理
unset http_proxy
unset https_proxy

test_user() {
    local username="$1"
    local description="$2"
    
    echo
    echo "🔍 测试账号: $username ($description)"
    echo "--------------------------------"
    
    # 登录获取token
    login_response=$(curl -s -X POST http://localhost:8001/auth/login \
        -H "Content-Type: application/json" \
        -d "{\"username\":\"$username\",\"password\":\"secret\"}")
    
    if echo "$login_response" | grep -q "success.*true"; then
        token=$(echo "$login_response" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
        echo "  🟢 登录成功"
        
        # 显示用户信息
        role=$(echo "$login_response" | grep -o '"role":"[^"]*"' | cut -d'"' -f4)
        echo "  👤 角色: $role"
        
        # 解析权限
        permissions=$(echo "$login_response" | sed 's/.*"permissions":\[\([^]]*\)\].*/\1/' | sed 's/"//g')
        echo "  🔑 权限: $permissions"
        
        # 测试基础API
        echo "  📊 API权限测试:"
        
        # 信件读取
        response=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $token" http://localhost:8001/api/letters)
        echo "    信件查看: $response $([ "$response" = "200" ] && echo "✅" || echo "❌")"
        
        # 信件创建
        response=$(curl -s -o /dev/null -w "%{http_code}" -X POST -H "Authorization: Bearer $token" -H "Content-Type: application/json" -d '{"title":"测试信件"}' http://localhost:8001/api/letters)
        echo "    信件创建: $response $([ "$response" = "200" ] && echo "✅" || echo "❌")"
        
        # 用户管理
        response=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $token" http://localhost:8001/api/users)
        echo "    用户管理: $response $([ "$response" = "200" ] && echo "✅" || [ "$response" = "403" ] && echo "🚫" || echo "❌")"
        
        # 系统配置
        response=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $token" http://localhost:8001/api/system/config)
        echo "    系统配置: $response $([ "$response" = "200" ] && echo "✅" || [ "$response" = "403" ] && echo "🚫" || echo "❌")"
        
        # 信使任务
        response=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $token" http://localhost:8002/api/courier/tasks)
        echo "    信使任务: $response $([ "$response" = "200" ] && echo "✅" || [ "$response" = "403" ] && echo "🚫" || echo "❌")"
        
        # 信使管理
        response=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $token" http://localhost:8002/api/courier/manage)
        echo "    信使管理: $response $([ "$response" = "200" ] && echo "✅" || [ "$response" = "403" ] && echo "🚫" || echo "❌")"
        
    else
        echo "  ❌ 登录失败: $login_response"
    fi
}

# 测试各个账号
test_user "alice" "普通用户"
test_user "courier1" "一级信使" 
test_user "senior_courier" "二级信使"
test_user "coordinator" "三级信使"
test_user "school_admin" "学校管理员"
test_user "platform_admin" "平台管理员"
test_user "super_admin" "超级管理员"

echo
echo "========================================="
echo "✅ 权限验证完成"
echo "说明: ✅=有权限且成功 🚫=无权限(403) ❌=错误"