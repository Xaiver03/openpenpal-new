#!/bin/bash

# OpenPenPal Mock Services 集成测试脚本
# 验证整个 Mock 服务系统的完整性

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# API 基础地址
API_BASE="http://localhost:8000"

# 测试结果
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
    PASSED_TESTS=$((PASSED_TESTS + 1))
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
    FAILED_TESTS=$((FAILED_TESTS + 1))
}

run_test() {
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    echo -e "${YELLOW}[TEST]${NC} $1"
}

# 测试工具函数
http_get() {
    local url="$1"
    local auth_header="$2"
    
    if [ -n "$auth_header" ]; then
        curl -s -w "%{http_code}" -H "Authorization: Bearer $auth_header" "$url"
    else
        curl -s -w "%{http_code}" "$url"
    fi
}

http_post() {
    local url="$1"
    local data="$2"
    local auth_header="$3"
    
    local curl_cmd="curl -s -w '%{http_code}' -X POST -H 'Content-Type: application/json'"
    
    if [ -n "$auth_header" ]; then
        curl_cmd="$curl_cmd -H 'Authorization: Bearer $auth_header'"
    fi
    
    curl_cmd="$curl_cmd -d '$data' '$url'"
    eval $curl_cmd
}

extract_status_code() {
    echo "$1" | tail -c 4
}

extract_response_body() {
    echo "$1" | head -c -4
}

# 测试1: 健康检查
test_health_checks() {
    run_test "健康检查 - API Gateway"
    local response=$(http_get "$API_BASE/health")
    local status=$(extract_status_code "$response")
    
    if [ "$status" = "200" ]; then
        log_success "API Gateway 健康检查通过"
    else
        log_error "API Gateway 健康检查失败: HTTP $status"
    fi
    
    run_test "健康检查 - 写信服务"
    local response=$(http_get "http://localhost:8001/health")
    local status=$(extract_status_code "$response")
    
    if [ "$status" = "200" ]; then
        log_success "写信服务健康检查通过"
    else
        log_error "写信服务健康检查失败: HTTP $status"
    fi
}

# 测试2: 用户认证
test_authentication() {
    run_test "用户登录 - 学生用户"
    local login_data='{"username":"alice","password":"secret"}'
    local response=$(http_post "$API_BASE/api/auth/login" "$login_data")
    local status=$(extract_status_code "$response")
    local body=$(extract_response_body "$response")
    
    if [ "$status" = "200" ]; then
        # 提取 token
        USER_TOKEN=$(echo "$body" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
        if [ -n "$USER_TOKEN" ]; then
            log_success "学生用户登录成功，获得 token"
        else
            log_error "登录成功但未获得 token"
        fi
    else
        log_error "学生用户登录失败: HTTP $status"
    fi
    
    run_test "用户登录 - 管理员用户"
    local admin_login_data='{"username":"admin","password":"admin123"}'
    local response=$(http_post "$API_BASE/api/auth/login" "$admin_login_data")
    local status=$(extract_status_code "$response")
    local body=$(extract_response_body "$response")
    
    if [ "$status" = "200" ]; then
        ADMIN_TOKEN=$(echo "$body" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
        if [ -n "$ADMIN_TOKEN" ]; then
            log_success "管理员登录成功，获得 token"
        else
            log_error "管理员登录成功但未获得 token"
        fi
    else
        log_error "管理员登录失败: HTTP $status"
    fi
}

# 测试3: 写信服务
test_write_service() {
    if [ -z "$USER_TOKEN" ]; then
        log_error "跳过写信服务测试 - 缺少用户 token"
        return
    fi
    
    run_test "获取信件列表"
    local response=$(http_get "$API_BASE/api/write/letters" "$USER_TOKEN")
    local status=$(extract_status_code "$response")
    
    if [ "$status" = "200" ]; then
        log_success "获取信件列表成功"
    else
        log_error "获取信件列表失败: HTTP $status"
    fi
    
    run_test "创建新信件"
    local letter_data='{"title":"测试信件","content":"这是一封测试信件的内容","receiverHint":"测试收件人"}'
    local response=$(http_post "$API_BASE/api/write/letters" "$letter_data" "$USER_TOKEN")
    local status=$(extract_status_code "$response")
    local body=$(extract_response_body "$response")
    
    if [ "$status" = "200" ]; then
        # 提取信件 ID
        LETTER_ID=$(echo "$body" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
        log_success "创建信件成功，ID: $LETTER_ID"
    else
        log_error "创建信件失败: HTTP $status"
    fi
    
    run_test "获取信件统计"
    local response=$(http_get "$API_BASE/api/write/letters/stats" "$USER_TOKEN")
    local status=$(extract_status_code "$response")
    
    if [ "$status" = "200" ]; then
        log_success "获取信件统计成功"
    else
        log_error "获取信件统计失败: HTTP $status"
    fi
}

# 测试4: 信使服务
test_courier_service() {
    if [ -z "$USER_TOKEN" ]; then
        log_error "跳过信使服务测试 - 缺少用户 token"
        return
    fi
    
    run_test "获取可用任务"
    local response=$(http_get "$API_BASE/api/courier/tasks" "$USER_TOKEN")
    local status=$(extract_status_code "$response")
    
    if [ "$status" = "200" ]; then
        log_success "获取可用任务成功"
    else
        log_error "获取可用任务失败: HTTP $status"
    fi
    
    run_test "申请成为信使"
    local courier_data='{"zone":"北京大学","phone":"13800138888","idCard":"110101199001011234","experience":"测试申请"}'
    local response=$(http_post "$API_BASE/api/courier/courier/apply" "$courier_data" "$USER_TOKEN")
    local status=$(extract_status_code "$response")
    
    if [ "$status" = "200" ] || [ "$status" = "409" ]; then
        log_success "信使申请处理正常"
    else
        log_error "信使申请失败: HTTP $status"
    fi
}

# 测试5: 管理服务
test_admin_service() {
    if [ -z "$ADMIN_TOKEN" ]; then
        log_error "跳过管理服务测试 - 缺少管理员 token"
        return
    fi
    
    run_test "获取用户列表"
    local response=$(http_get "$API_BASE/api/admin/users" "$ADMIN_TOKEN")
    local status=$(extract_status_code "$response")
    
    if [ "$status" = "200" ]; then
        log_success "获取用户列表成功"
    else
        log_error "获取用户列表失败: HTTP $status"
    fi
    
    run_test "获取系统配置"
    local response=$(http_get "$API_BASE/api/admin/system/config" "$ADMIN_TOKEN")
    local status=$(extract_status_code "$response")
    
    if [ "$status" = "200" ]; then
        log_success "获取系统配置成功"
    else
        log_error "获取系统配置失败: HTTP $status"
    fi
    
    run_test "获取博物馆展览"
    local response=$(http_get "$API_BASE/api/admin/museum/exhibitions" "$ADMIN_TOKEN")
    local status=$(extract_status_code "$response")
    
    if [ "$status" = "200" ]; then
        log_success "获取博物馆展览成功"
    else
        log_error "获取博物馆展览失败: HTTP $status"
    fi
}

# 测试6: 权限验证
test_permission_control() {
    if [ -z "$USER_TOKEN" ]; then
        log_error "跳过权限验证测试 - 缺少用户 token"
        return
    fi
    
    run_test "普通用户访问管理接口应被拒绝"
    local response=$(http_get "$API_BASE/api/admin/users" "$USER_TOKEN")
    local status=$(extract_status_code "$response")
    
    if [ "$status" = "403" ]; then
        log_success "权限控制正常 - 普通用户被拒绝访问管理接口"
    else
        log_error "权限控制异常 - 普通用户可以访问管理接口: HTTP $status"
    fi
    
    run_test "无 token 访问受保护接口应被拒绝"
    local response=$(http_get "$API_BASE/api/write/letters")
    local status=$(extract_status_code "$response")
    
    if [ "$status" = "401" ]; then
        log_success "认证控制正常 - 无 token 访问被拒绝"
    else
        log_error "认证控制异常 - 无 token 可以访问受保护接口: HTTP $status"
    fi
}

# 测试7: OCR 服务
test_ocr_service() {
    if [ -z "$USER_TOKEN" ]; then
        log_error "跳过 OCR 服务测试 - 缺少用户 token"
        return
    fi
    
    run_test "获取 OCR 模型列表"
    local response=$(http_get "$API_BASE/api/ocr/models" "$USER_TOKEN")
    local status=$(extract_status_code "$response")
    
    if [ "$status" = "200" ]; then
        log_success "获取 OCR 模型列表成功"
    else
        log_error "获取 OCR 模型列表失败: HTTP $status"
    fi
}

# 主测试函数
run_integration_tests() {
    echo -e "${BLUE}OpenPenPal Mock Services 集成测试${NC}"
    echo "============================================"
    
    # 检查服务是否运行
    log_info "检查服务状态..."
    if ! curl -s http://localhost:8000/health > /dev/null; then
        log_error "API Gateway 未运行，请先启动 Mock 服务"
        echo "运行: ./scripts/start-mock.sh"
        exit 1
    fi
    
    log_info "开始集成测试..."
    echo ""
    
    # 运行所有测试
    test_health_checks
    test_authentication
    test_write_service
    test_courier_service
    test_admin_service
    test_permission_control
    test_ocr_service
    
    # 显示测试结果
    echo ""
    echo "============================================"
    echo -e "${BLUE}集成测试完成${NC}"
    echo "============================================"
    echo -e "总测试数: $TOTAL_TESTS"
    echo -e "${GREEN}通过: $PASSED_TESTS${NC}"
    echo -e "${RED}失败: $FAILED_TESTS${NC}"
    
    if [ $FAILED_TESTS -eq 0 ]; then
        echo -e "${GREEN}✅ 所有集成测试通过！${NC}"
        echo -e "${GREEN}🎉 Mock 服务系统运行正常${NC}"
        exit 0
    else
        echo -e "${RED}❌ $FAILED_TESTS 项测试失败${NC}"
        echo -e "${YELLOW}请检查失败的测试项目${NC}"
        exit 1
    fi
}

# 显示帮助信息
show_help() {
    echo "OpenPenPal Mock Services 集成测试"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  --help, -h    显示帮助信息"
    echo "  --verbose     显示详细输出"
    echo ""
    echo "注意: 运行测试前请确保 Mock 服务已启动"
    echo "启动命令: ./scripts/start-mock.sh"
}

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        --help|-h)
            show_help
            exit 0
            ;;
        --verbose)
            set -x
            shift
            ;;
        *)
            echo "未知参数: $1"
            show_help
            exit 1
            ;;
    esac
done

# 运行测试
run_integration_tests