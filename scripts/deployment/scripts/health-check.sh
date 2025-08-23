#!/bin/bash
# 健康检查脚本 - 验证所有服务状态

set -euo pipefail

# 配置
ENV=${1:-production}
BASE_URL="http://localhost"
TIMEOUT=10
MAX_RETRIES=3

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# 计数器
total_checks=0
passed_checks=0
failed_checks=0

# 日志函数
log_success() {
    echo -e "${GREEN}✓${NC} $1"
    ((passed_checks++))
}

log_failure() {
    echo -e "${RED}✗${NC} $1"
    ((failed_checks++))
}

log_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

log_info() {
    echo -e "ℹ $1"
}

# HTTP 健康检查
check_http_endpoint() {
    local name=$1
    local url=$2
    local expected_status=${3:-200}
    
    ((total_checks++))
    
    log_info "检查 $name..."
    
    for retry in $(seq 1 $MAX_RETRIES); do
        local response=$(curl -s -o /dev/null -w "%{http_code}" \
            --connect-timeout $TIMEOUT \
            --max-time $TIMEOUT \
            "$url" 2>/dev/null || echo "000")
        
        if [ "$response" = "$expected_status" ]; then
            log_success "$name (HTTP $response)"
            return 0
        elif [ $retry -lt $MAX_RETRIES ]; then
            sleep 2
        fi
    done
    
    log_failure "$name (HTTP $response，期望 $expected_status)"
    return 1
}

# 数据库健康检查
check_database() {
    local name=$1
    local container=$2
    local db_name=$3
    local db_user=$4
    
    ((total_checks++))
    
    log_info "检查 $name..."
    
    if docker exec "$container" pg_isready -U "$db_user" -d "$db_name" &> /dev/null; then
        # 进一步检查连接
        if docker exec "$container" psql -U "$db_user" -d "$db_name" -c "SELECT 1" &> /dev/null; then
            log_success "$name 连接正常"
            return 0
        fi
    fi
    
    log_failure "$name 连接失败"
    return 1
}

# Redis 健康检查
check_redis() {
    local name=$1
    local container=$2
    
    ((total_checks++))
    
    log_info "检查 $name..."
    
    if docker exec "$container" redis-cli ping | grep -q "PONG" 2>/dev/null; then
        # 检查内存使用
        local memory_usage=$(docker exec "$container" redis-cli info memory | grep used_memory_human | cut -d: -f2 | tr -d '\r')
        log_success "$name (内存使用: $memory_usage)"
        return 0
    fi
    
    log_failure "$name 连接失败"
    return 1
}

# 容器资源检查
check_container_resources() {
    local container=$1
    
    if ! docker ps --format "table {{.Names}}" | grep -q "$container"; then
        return 1
    fi
    
    # 获取资源使用情况
    local stats=$(docker stats "$container" --no-stream --format "{{.CPUPerc}}\t{{.MemPerc}}" | head -1)
    local cpu=$(echo "$stats" | cut -f1)
    local mem=$(echo "$stats" | cut -f2)
    
    # 检查是否超过阈值
    local cpu_value=$(echo "$cpu" | sed 's/%//')
    local mem_value=$(echo "$mem" | sed 's/%//')
    
    if (( $(echo "$cpu_value > 90" | bc -l) )); then
        log_warning "$container CPU 使用率过高: $cpu"
    fi
    
    if (( $(echo "$mem_value > 90" | bc -l) )); then
        log_warning "$container 内存使用率过高: $mem"
    fi
}

# 端口检查
check_port() {
    local service=$1
    local port=$2
    
    if nc -z localhost "$port" 2>/dev/null; then
        return 0
    else
        return 1
    fi
}

# WebSocket 健康检查
check_websocket() {
    local name=$1
    local url=$2
    
    ((total_checks++))
    
    log_info "检查 $name WebSocket..."
    
    # 使用 curl 测试 WebSocket 升级
    local response=$(curl -s -o /dev/null -w "%{http_code}" \
        -H "Upgrade: websocket" \
        -H "Connection: Upgrade" \
        -H "Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==" \
        -H "Sec-WebSocket-Version: 13" \
        "$url" 2>/dev/null || echo "000")
    
    if [ "$response" = "101" ] || [ "$response" = "426" ]; then
        log_success "$name WebSocket 端点可用"
        return 0
    fi
    
    log_failure "$name WebSocket 端点不可用 (HTTP $response)"
    return 1
}

# 服务依赖检查
check_service_dependencies() {
    log_info "检查服务依赖关系..."
    
    # 检查网络连接
    local services=("backend" "write-service" "courier-service" "admin-service" "ocr-service")
    
    for service in "${services[@]}"; do
        local container="openpenpal-$service"
        
        # 检查是否能连接到数据库
        if docker exec "$container" nc -zv openpenpal-postgres 5432 &> /dev/null; then
            log_success "$service → PostgreSQL 连接正常"
        else
            log_warning "$service → PostgreSQL 连接失败"
        fi
        
        # 检查是否能连接到 Redis
        if docker exec "$container" nc -zv openpenpal-redis 6379 &> /dev/null; then
            log_success "$service → Redis 连接正常"
        else
            log_warning "$service → Redis 连接失败"
        fi
    done
}

# 业务功能检查
check_business_functions() {
    log_info "\n执行业务功能检查..."
    
    # 检查用户注册
    local register_response=$(curl -s -X POST "$BASE_URL:8000/api/v1/auth/register" \
        -H "Content-Type: application/json" \
        -d '{"email":"test@example.com","password":"test123","name":"Test User"}' \
        -w "\n%{http_code}" 2>/dev/null | tail -1)
    
    if [ "$register_response" = "200" ] || [ "$register_response" = "409" ]; then
        log_success "用户注册接口正常"
    else
        log_failure "用户注册接口异常 (HTTP $register_response)"
    fi
    
    # 检查健康检查端点
    if curl -sf "$BASE_URL:8000/health" > /dev/null; then
        log_success "API 网关健康检查正常"
    else
        log_failure "API 网关健康检查失败"
    fi
}

# 主函数
main() {
    echo "========================================="
    echo "OpenPenPal 健康检查"
    echo "环境: $ENV"
    echo "时间: $(date '+%Y-%m-%d %H:%M:%S')"
    echo "========================================="
    echo
    
    # 基础服务检查
    echo "【基础服务检查】"
    check_database "PostgreSQL" "openpenpal-postgres" "openpenpal" "openpenpal"
    check_redis "Redis" "openpenpal-redis"
    echo
    
    # HTTP 服务检查
    echo "【HTTP 服务检查】"
    check_http_endpoint "前端应用" "$BASE_URL:3000" "200"
    check_http_endpoint "管理后台" "$BASE_URL:3001" "200"
    check_http_endpoint "API 网关" "$BASE_URL:8000/health" "200"
    check_http_endpoint "主后端服务" "$BASE_URL:8080/health" "200"
    check_http_endpoint "写信服务" "$BASE_URL:8001/health" "200"
    check_http_endpoint "信使服务" "$BASE_URL:8002/health" "200"
    check_http_endpoint "管理服务" "$BASE_URL:8003/actuator/health" "200"
    check_http_endpoint "OCR 服务" "$BASE_URL:8004/health" "200"
    echo
    
    # WebSocket 检查
    echo "【WebSocket 检查】"
    check_websocket "API 网关" "$BASE_URL:8000/ws"
    echo
    
    # 端口检查
    echo "【端口可用性检查】"
    local ports=("80:Nginx" "443:Nginx-SSL" "5432:PostgreSQL" "6379:Redis" "9090:Prometheus" "3002:Grafana")
    
    for port_info in "${ports[@]}"; do
        IFS=':' read -r port name <<< "$port_info"
        ((total_checks++))
        
        if check_port "$name" "$port"; then
            log_success "$name 端口 $port 已开放"
        else
            log_failure "$name 端口 $port 未开放"
        fi
    done
    echo
    
    # 容器资源检查
    echo "【容器资源使用情况】"
    local containers=(
        "openpenpal-frontend"
        "openpenpal-gateway"
        "openpenpal-backend"
        "openpenpal-write-service"
        "openpenpal-courier-service"
        "openpenpal-admin-service"
        "openpenpal-ocr-service"
        "openpenpal-postgres"
        "openpenpal-redis"
    )
    
    for container in "${containers[@]}"; do
        check_container_resources "$container"
    done
    echo
    
    # 服务依赖检查
    check_service_dependencies
    echo
    
    # 业务功能检查
    check_business_functions
    echo
    
    # 总结
    echo "========================================="
    echo "检查完成"
    echo "总检查项: $total_checks"
    echo -e "${GREEN}通过: $passed_checks${NC}"
    echo -e "${RED}失败: $failed_checks${NC}"
    
    if [ $failed_checks -eq 0 ]; then
        echo -e "\n${GREEN}✓ 所有健康检查通过！${NC}"
        exit 0
    else
        echo -e "\n${RED}✗ 健康检查失败！${NC}"
        exit 1
    fi
}

# 执行主函数
main "$@"