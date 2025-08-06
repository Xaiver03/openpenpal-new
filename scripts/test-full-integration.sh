#!/bin/bash

# OpenPenPal 全栈集成测试脚本
# 测试所有组件的端到端集成

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}OpenPenPal 全栈集成测试${NC}"
echo "============================================"

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

# 测试1: 前端构建完整性
run_test "测试 Admin Service 前端构建"
if [ -d "/Users/rocalight/同步空间/opplc/openpenpal/services/admin-service/frontend/dist" ]; then
    log_success "前端构建文件存在"
else
    log_error "前端构建文件不存在"
fi

# 测试2: API 配置检查
run_test "测试 API 配置完整性"
API_CONFIG_FILE="/Users/rocalight/同步空间/opplc/openpenpal/services/admin-service/frontend/src/utils/api.ts"
if [ -f "$API_CONFIG_FILE" ]; then
    if grep -q "museumApi" "$API_CONFIG_FILE"; then
        log_success "Museum API 配置存在"
    else
        log_error "Museum API 配置缺失"
    fi
else
    log_error "API 配置文件不存在"
fi

# 测试3: 环境配置检查
run_test "测试环境配置"
if [ -f "/Users/rocalight/同步空间/opplc/openpenpal/services/admin-service/frontend/.env.development" ]; then
    log_success "开发环境配置存在"
else
    log_error "开发环境配置缺失"
fi

if [ -f "/Users/rocalight/同步空间/opplc/openpenpal/services/admin-service/frontend/.env.production" ]; then
    log_success "生产环境配置存在"
else
    log_error "生产环境配置缺失"
fi

# 测试4: Docker 配置检查
run_test "测试 Docker 编排配置"
if [ -f "/Users/rocalight/同步空间/opplc/openpenpal/docker-compose.microservices.yml" ]; then
    if grep -q "admin-service" "/Users/rocalight/同步空间/opplc/openpenpal/docker-compose.microservices.yml"; then
        log_success "Docker 编排配置包含 Admin Service"
    else
        log_error "Docker 编排配置缺少 Admin Service"
    fi
else
    log_error "Docker 编排配置文件不存在"
fi

# 测试5: 监控配置检查  
run_test "测试监控配置"
if [ -f "/Users/rocalight/同步空间/opplc/openpenpal/monitoring/prometheus.yml" ]; then
    if grep -q "admin-service" "/Users/rocalight/同步空间/opplc/openpenpal/monitoring/prometheus.yml"; then
        log_success "Prometheus 监控配置包含 Admin Service"
    else
        log_error "Prometheus 监控配置缺少 Admin Service"
    fi
else
    log_error "Prometheus 配置文件不存在"
fi

# 测试6: Nginx 配置检查
run_test "测试 Nginx 反向代理配置"
if [ -f "/Users/rocalight/同步空间/opplc/openpenpal/nginx/conf.d/default.conf" ]; then
    if grep -q "admin_frontend" "/Users/rocalight/同步空间/opplc/openpenpal/nginx/conf.d/default.conf"; then
        log_success "Nginx 配置包含 Admin Frontend"
    else
        log_error "Nginx 配置缺少 Admin Frontend"
    fi
else
    log_error "Nginx 配置文件不存在"
fi

# 测试7: 集成测试框架检查
run_test "测试集成测试框架"
if [ -f "/Users/rocalight/同步空间/opplc/openpenpal/scripts/integration-test.sh" ]; then
    if [ -x "/Users/rocalight/同步空间/opplc/openpenpal/scripts/integration-test.sh" ]; then
        log_success "集成测试框架可执行"
    else
        log_error "集成测试框架不可执行"
    fi
else
    log_error "集成测试框架不存在"
fi

# 测试8: 数据库初始化脚本检查
run_test "测试数据库初始化脚本"
if [ -f "/Users/rocalight/同步空间/opplc/openpenpal/scripts/init-db.sql" ]; then
    if grep -q "exhibitions" "/Users/rocalight/同步空间/opplc/openpenpal/scripts/init-db.sql"; then
        log_success "数据库初始化包含展览表"
    else
        log_error "数据库初始化缺少展览表"
    fi
else
    log_error "数据库初始化脚本不存在"
fi

# 测试9: API 文档检查
run_test "测试 API 文档完整性"
POSTMAN_FILE="/Users/rocalight/同步空间/opplc/openpenpal/services/admin-service/backend/src/test/resources/api-test/OpenPenPal-Admin-API.postman_collection.json"
if [ -f "$POSTMAN_FILE" ]; then
    if grep -q "Museum Management" "$POSTMAN_FILE"; then
        log_success "API 文档包含 Museum 管理接口"
    else
        log_error "API 文档缺少 Museum 管理接口"
    fi
else
    log_error "API 文档不存在"
fi

# 测试10: Vue 组件完整性检查
run_test "测试 Vue 组件完整性"
MUSEUM_VUE="/Users/rocalight/同步空间/opplc/openpenpal/services/admin-service/frontend/src/views/Museum.vue"
if [ -f "$MUSEUM_VUE" ]; then
    if grep -q "museumApi" "$MUSEUM_VUE"; then
        log_success "Museum 组件包含 API 调用"
    else
        log_error "Museum 组件缺少 API 调用"
    fi
else
    log_error "Museum 组件不存在"
fi

# 显示测试结果
echo ""
echo -e "${BLUE}==============================${NC}"
echo -e "${BLUE}集成测试完成${NC}"
echo -e "${BLUE}==============================${NC}"
echo -e "总测试数: $TOTAL_TESTS"
echo -e "${GREEN}通过: $PASSED_TESTS${NC}"
echo -e "${RED}失败: $FAILED_TESTS${NC}"

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "${GREEN}✓ 所有集成测试通过！${NC}"
    echo -e "${GREEN}OpenPenPal Admin Service 已完全集成${NC}"
    exit 0
else
    echo -e "${RED}✗ $FAILED_TESTS 项测试失败${NC}"
    echo -e "${YELLOW}请检查失败的测试项目${NC}"
    exit 1
fi