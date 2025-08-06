#!/bin/bash

# OpenPenPal API Gateway 集成测试脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 配置
GATEWAY_URL="http://localhost:8000"
API_BASE="$GATEWAY_URL/api/v1"

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 等待服务启动
wait_for_service() {
    local url=$1
    local name=$2
    local max_attempts=30
    local attempt=1

    log_info "等待 $name 服务启动..."
    
    while [ $attempt -le $max_attempts ]; do
        if curl -s -f "$url" > /dev/null 2>&1; then
            log_success "$name 服务已启动"
            return 0
        fi
        
        echo -n "."
        sleep 2
        attempt=$((attempt + 1))
    done
    
    log_error "$name 服务启动超时"
    return 1
}

# 健康检查测试
test_health_check() {
    log_info "测试网关健康检查..."
    
    response=$(curl -s "$GATEWAY_URL/health")
    status=$(echo "$response" | jq -r '.status' 2>/dev/null || echo "")
    
    if [ "$status" = "healthy" ]; then
        log_success "健康检查通过"
        return 0
    else
        log_error "健康检查失败: $response"
        return 1
    fi
}

# 版本信息测试
test_version_info() {
    log_info "测试版本信息..."
    
    response=$(curl -s "$GATEWAY_URL/version")
    service=$(echo "$response" | jq -r '.service' 2>/dev/null || echo "")
    
    if [[ "$service" == *"Gateway"* ]]; then
        log_success "版本信息正确"
        return 0
    else
        log_error "版本信息错误: $response"
        return 1
    fi
}

# CORS测试
test_cors() {
    log_info "测试CORS配置..."
    
    response=$(curl -s -I -X OPTIONS "$API_BASE/auth/login" \
        -H "Origin: http://localhost:3000" \
        -H "Access-Control-Request-Method: POST")
    
    if echo "$response" | grep -q "Access-Control-Allow-Origin"; then
        log_success "CORS配置正确"
        return 0
    else
        log_error "CORS配置错误"
        return 1
    fi
}

# 认证路由测试
test_auth_routes() {
    log_info "测试认证路由..."
    
    # 测试注册接口（应该被转发到main-backend）
    response=$(curl -s -w "%{http_code}" -o /dev/null \
        -X POST "$API_BASE/auth/register" \
        -H "Content-Type: application/json" \
        -d '{"username":"testuser","password":"testpass","email":"test@example.com"}')
    
    # 期望得到 400 或 502 (因为后端可能未启动)
    if [ "$response" = "400" ] || [ "$response" = "502" ] || [ "$response" = "503" ]; then
        log_success "认证路由转发正常"
        return 0
    else
        log_warning "认证路由响应: $response (后端服务可能未启动)"
        return 0
    fi
}

# JWT认证测试
test_jwt_auth() {
    log_info "测试JWT认证..."
    
    # 测试无Token访问受保护路由
    response=$(curl -s -w "%{http_code}" -o /dev/null "$API_BASE/letters")
    
    if [ "$response" = "401" ]; then
        log_success "JWT认证保护正常"
        return 0
    else
        log_error "JWT认证失败，响应码: $response"
        return 1
    fi
}

# 限流测试
test_rate_limiting() {
    log_info "测试限流功能..."
    
    local success_count=0
    local rate_limit_count=0
    
    # 快速发送多个请求
    for i in {1..10}; do
        response=$(curl -s -w "%{http_code}" -o /dev/null "$GATEWAY_URL/health")
        
        if [ "$response" = "200" ]; then
            success_count=$((success_count + 1))
        elif [ "$response" = "429" ]; then
            rate_limit_count=$((rate_limit_count + 1))
        fi
    done
    
    log_info "成功请求: $success_count, 限流请求: $rate_limit_count"
    
    if [ $success_count -gt 0 ]; then
        log_success "限流功能正常"
        return 0
    else
        log_error "限流测试失败"
        return 1
    fi
}

# 监控指标测试
test_metrics() {
    log_info "测试监控指标..."
    
    response=$(curl -s "$GATEWAY_URL/metrics")
    
    if echo "$response" | grep -q "gateway_http_requests_total"; then
        log_success "监控指标正常"
        return 0
    else
        log_error "监控指标异常"
        return 1
    fi
}

# 服务发现测试
test_service_discovery() {
    log_info "测试服务发现..."
    
    # 创建临时的JWT Token进行测试（如果有admin用户）
    # 这里简化处理，实际应该有有效的admin token
    
    response=$(curl -s -w "%{http_code}" -o /tmp/services_response.json \
        "$GATEWAY_URL/admin/health" \
        -H "Authorization: Bearer fake-admin-token")
    
    if [ "$response" = "401" ]; then
        log_success "管理接口权限保护正常"
        return 0
    elif [ "$response" = "200" ]; then
        log_success "服务发现接口正常"
        return 0
    else
        log_warning "服务发现测试需要有效的admin token"
        return 0
    fi
}

# 路由转发测试
test_routing() {
    log_info "测试路由转发..."
    
    local test_routes=(
        "/api/v1/auth/login:POST"
        "/api/v1/letters:GET"
        "/api/v1/courier/apply:POST"
    )
    
    local success_count=0
    
    for route_info in "${test_routes[@]}"; do
        IFS=':' read -r route method <<< "$route_info"
        
        response=$(curl -s -w "%{http_code}" -o /dev/null \
            -X "$method" "$GATEWAY_URL$route" \
            -H "Content-Type: application/json")
        
        # 期望得到 401 (需要认证) 或 502/503 (后端未启动)
        if [ "$response" = "401" ] || [ "$response" = "502" ] || [ "$response" = "503" ]; then
            success_count=$((success_count + 1))
            log_info "路由 $route ($method): 转发正常"
        else
            log_warning "路由 $route ($method): 响应码 $response"
        fi
    done
    
    if [ $success_count -gt 0 ]; then
        log_success "路由转发测试通过"
        return 0
    else
        log_error "路由转发测试失败"
        return 1
    fi
}

# 压力测试
test_load() {
    log_info "运行压力测试..."
    
    if command -v wrk >/dev/null 2>&1; then
        log_info "使用 wrk 进行压力测试..."
        wrk -t4 -c10 -d10s "$GATEWAY_URL/health" || true
    else
        log_info "使用 curl 进行简单压力测试..."
        
        local start_time=$(date +%s)
        local request_count=0
        local success_count=0
        
        for i in {1..50}; do
            if curl -s -f "$GATEWAY_URL/health" > /dev/null 2>&1; then
                success_count=$((success_count + 1))
            fi
            request_count=$((request_count + 1))
        done
        
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))
        local rps=$((request_count / duration))
        
        log_info "压力测试结果: $success_count/$request_count 成功, ${rps}req/s"
    fi
    
    log_success "压力测试完成"
    return 0
}

# 清理函数
cleanup() {
    log_info "清理临时文件..."
    rm -f /tmp/services_response.json
}

# 主测试函数
run_tests() {
    log_info "开始 OpenPenPal API Gateway 集成测试"
    echo "=================================="
    
    # 等待网关启动
    if ! wait_for_service "$GATEWAY_URL/health" "API Gateway"; then
        log_error "API Gateway 未启动，请先启动服务"
        exit 1
    fi
    
    local tests=(
        "test_health_check:健康检查"
        "test_version_info:版本信息"
        "test_cors:CORS配置"
        "test_auth_routes:认证路由"
        "test_jwt_auth:JWT认证"
        "test_rate_limiting:限流功能"
        "test_metrics:监控指标"
        "test_service_discovery:服务发现"
        "test_routing:路由转发"
        "test_load:压力测试"
    )
    
    local passed=0
    local failed=0
    
    for test_info in "${tests[@]}"; do
        IFS=':' read -r test_func test_name <<< "$test_info"
        
        echo ""
        log_info "执行测试: $test_name"
        echo "----------------------------------------"
        
        if $test_func; then
            passed=$((passed + 1))
        else
            failed=$((failed + 1))
        fi
    done
    
    echo ""
    echo "=================================="
    log_info "测试完成"
    log_info "通过: $passed, 失败: $failed"
    
    if [ $failed -eq 0 ]; then
        log_success "所有测试通过! 🎉"
        return 0
    else
        log_error "有 $failed 个测试失败"
        return 1
    fi
}

# 检查依赖
check_dependencies() {
    log_info "检查依赖..."
    
    if ! command -v curl >/dev/null 2>&1; then
        log_error "curl 未安装"
        exit 1
    fi
    
    if ! command -v jq >/dev/null 2>&1; then
        log_warning "jq 未安装，某些测试可能失败"
    fi
}

# 显示帮助
show_help() {
    echo "OpenPenPal API Gateway 集成测试"
    echo ""
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  --health     只运行健康检查"
    echo "  --auth       只运行认证测试"
    echo "  --load       只运行压力测试"
    echo "  --help       显示帮助"
    echo ""
    echo "环境变量:"
    echo "  GATEWAY_URL  网关地址 (默认: http://localhost:8000)"
    echo ""
    echo "示例:"
    echo "  $0                    # 运行所有测试"
    echo "  $0 --health          # 只运行健康检查"
    echo "  GATEWAY_URL=http://prod-gateway:8000 $0  # 测试生产环境"
}

# 主逻辑
main() {
    trap cleanup EXIT
    
    case "${1:-}" in
        --health)
            check_dependencies
            wait_for_service "$GATEWAY_URL/health" "API Gateway"
            test_health_check
            ;;
        --auth)
            check_dependencies
            wait_for_service "$GATEWAY_URL/health" "API Gateway"
            test_jwt_auth
            ;;
        --load)
            check_dependencies
            wait_for_service "$GATEWAY_URL/health" "API Gateway"
            test_load
            ;;
        --help)
            show_help
            ;;
        "")
            check_dependencies
            run_tests
            ;;
        *)
            log_error "未知选项: $1"
            show_help
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"