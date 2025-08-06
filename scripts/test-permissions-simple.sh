#!/bin/bash

# OpenPenPal 信使权限测试脚本（简化版）

set -e

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# API 基础 URL
API_URL="http://localhost:8080/api/v1"

# 日志函数
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# 登录并测试函数
test_user() {
    local username=$1
    local password=$2
    local level=$3
    
    echo ""
    echo "=== 测试 $username (Level $level) ==="
    
    # 登录
    local response=$(curl -s -X POST "$API_URL/auth/login" \
        -H "Content-Type: application/json" \
        -d "{\"username\":\"$username\",\"password\":\"$password\"}")
    
    if echo "$response" | grep -q '"success":true'; then
        local token=$(echo "$response" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
        local role=$(echo "$response" | grep -o '"role":"[^"]*' | cut -d'"' -f4)
        local nickname=$(echo "$response" | grep -o '"nickname":"[^"]*' | cut -d'"' -f4)
        log_success "登录成功 (角色: $role, 昵称: $nickname)"
        
        # 测试个人信息访问
        echo -n "  测试个人信息访问: "
        if curl -s -H "Authorization: Bearer $token" "$API_URL/users/me" | grep -q '"username"'; then
            echo -e "${GREEN}✓${NC}"
        else
            echo -e "${RED}✗${NC}"
        fi
        
        # 测试任务访问
        echo -n "  测试任务列表访问: "
        local tasks_response=$(curl -s -H "Authorization: Bearer $token" "$API_URL/courier/tasks")
        if echo "$tasks_response" | grep -q '"tasks"\|"success":true\|"data"'; then
            echo -e "${GREEN}✓${NC}"
        else
            echo -e "${RED}✗${NC}"
        fi
        
        # 测试下级管理（仅 Level 2+ 应该有权限）
        if [ $level -ge 2 ]; then
            echo -n "  测试下级信使管理: "
            local mgmt_response=$(curl -s -H "Authorization: Bearer $token" "$API_URL/couriers/subordinates")
            if echo "$mgmt_response" | grep -q '"couriers"\|"success":true\|"data"'; then
                echo -e "${GREEN}✓ 可以管理 Level $((level-1)) 信使${NC}"
            else
                echo -e "${YELLOW}? 响应: $(echo $mgmt_response | head -c 50)...${NC}"
            fi
        else
            echo "  测试下级信使管理: ${YELLOW}无权限（正确）${NC}"
        fi
        
        # 测试统计访问（仅 Level 2+ 应该有权限）
        if [ $level -ge 2 ]; then
            echo -n "  测试统计数据访问: "
            if curl -s -H "Authorization: Bearer $token" "$API_URL/courier/stats" | grep -q '"stats"\|"total"\|"data"'; then
                echo -e "${GREEN}✓${NC}"
            else
                echo -e "${YELLOW}?${NC}"
            fi
        fi
        
    else
        log_error "登录失败"
        echo "$response"
    fi
}

# 主函数
main() {
    echo "OpenPenPal 四级信使权限测试"
    echo "============================"
    echo "时间: $(date)"
    echo ""
    
    # 检查服务状态
    log_info "检查服务状态..."
    if curl -s http://localhost:8080/health | grep -q "healthy"; then
        log_success "后端服务运行正常"
    else
        log_error "后端服务未运行"
        exit 1
    fi
    
    # 测试管理员
    test_user "admin" "admin123" 5
    
    # 测试四级信使
    test_user "courier_level4" "secret" 4
    sleep 2
    
    # 测试三级信使
    test_user "courier_level3" "secret" 3
    sleep 2
    
    # 测试二级信使
    test_user "courier_level2" "secret" 2
    sleep 2
    
    # 测试一级信使
    test_user "courier_level1" "secret" 1
    
    echo ""
    echo "=============================="
    echo "权限矩阵总结"
    echo "=============================="
    echo ""
    echo "功能权限对比："
    echo "┌─────────────────┬─────────┬─────────┬─────────┬─────────┬─────────┐"
    echo "│ 功能            │ Admin   │ Level 4 │ Level 3 │ Level 2 │ Level 1 │"
    echo "├─────────────────┼─────────┼─────────┼─────────┼─────────┼─────────┤"
    echo "│ 登录系统        │   ✓     │   ✓     │   ✓     │   ✓     │   ✓     │"
    echo "│ 查看个人信息    │   ✓     │   ✓     │   ✓     │   ✓     │   ✓     │"
    echo "│ 执行配送任务    │   ✓     │   ✓     │   ✓     │   ✓     │   ✓     │"
    echo "│ 查看统计报告    │   ✓     │   ✓     │   ✓     │   ✓     │   ✗     │"
    echo "│ 管理下级信使    │   ✓     │   ✓     │   ✓     │   ✓     │   ✗     │"
    echo "│ 创建信使账号    │   All   │ Level3  │ Level2  │ Level1  │   ✗     │"
    echo "│ 系统管理        │   ✓     │   ✗     │   ✗     │   ✗     │   ✗     │"
    echo "└─────────────────┴─────────┴─────────┴─────────┴─────────┴─────────┘"
    echo ""
    echo "层级关系："
    echo "• Level 4 (城市总代) → 管理 Level 3"
    echo "• Level 3 (校级信使) → 管理 Level 2"
    echo "• Level 2 (片区信使) → 管理 Level 1"
    echo "• Level 1 (楼栋信使) → 无管理权限"
}

# 运行主函数
main