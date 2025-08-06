#!/bin/bash
# OpenPenPal 测试套件启动器
# 一键运行所有测试

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 项目根目录
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEST_DIR="$PROJECT_ROOT"

# 打印标题
print_title() {
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}🎯 $1${NC}"
    echo -e "${BLUE}========================================${NC}"
}

# 打印成功
print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

# 打印错误
print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# 打印警告
print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

# 检查环境
check_environment() {
    print_title "环境检查"
    
    # 检查Node.js
    if command -v node > /dev/null 2>&1; then
        NODE_VERSION=$(node --version)
        print_success "Node.js: $NODE_VERSION"
    else
        print_error "Node.js 未安装"
        exit 1
    fi
    
    # 检查Go
    if command -v go > /dev/null 2>&1; then
        GO_VERSION=$(go version)
        print_success "Go: $GO_VERSION"
    else
        print_error "Go 未安装"
        exit 1
    fi
    
    # 检查Docker
    if command -v docker > /dev/null 2>&1; then
        DOCKER_VERSION=$(docker --version)
        print_success "Docker: $DOCKER_VERSION"
    else
        print_warning "Docker 未安装，跳过容器测试"
    fi
    
    # 检查服务
    print_title "服务状态检查"
    
    # 检查前端
    if curl -s http://localhost:3000 > /dev/null 2>&1; then
        print_success "前端服务运行正常 (http://localhost:3000)"
    else
        print_error "前端服务未运行"
    fi
    
    # 检查后端
    if curl -s http://localhost:8080/health > /dev/null 2>&1; then
        print_success "后端服务运行正常 (http://localhost:8080)"
    else
        print_warning "后端服务未运行，尝试启动..."
    fi
}

# 运行测试套件
run_test_suite() {
    local test_type=$1
    local script_name=$2
    
    print_title "运行 $test_type 测试"
    
    if [[ -f "$TEST_DIR/scripts/$script_name" ]]; then
        chmod +x "$TEST_DIR/scripts/$script_name"
        if bash "$TEST_DIR/scripts/$script_name"; then
            print_success "$test_type 测试通过"
            return 0
        else
            print_error "$test_type 测试失败"
            return 1
        fi
    else
        print_warning "$test_type 测试脚本不存在: $script_name"
        return 1
    fi
}

# 生成测试报告
generate_report() {
    print_title "生成测试报告"
    
    REPORT_FILE="$TEST_DIR/reports/test_report_$(date +%Y%m%d_%H%M%S).md"
    
    cat << EOF > "$REPORT_FILE"
# OpenPenPal 测试报告

## 测试执行摘要
- **执行时间**: $(date '+%Y-%m-%d %H:%M:%S')
- **环境**: $(uname -a)
- **Node.js**: $(node --version)
- **Go**: $(go version)

## 测试结果
- **总体状态**: $1
- **测试套件**: $2
- **通过**: $3
- **失败**: $4

## 服务状态
- 前端: $(curl -s http://localhost:3000 > /dev/null 2>&1 && echo "运行中" || echo "停止")
- 后端: $(curl -s http://localhost:8080/health > /dev/null 2>&1 && echo "运行中" || echo "停止")

## 后续建议
1. 修复失败的测试
2. 更新测试用例
3. 优化性能
4. 安全扫描

EOF

    print_success "测试报告已生成: $REPORT_FILE"
}

# 主函数
main() {
    print_title "OpenPenPal 测试套件"
    echo "开始执行全面测试..."
    
    # 检查环境
    check_environment
    
    # 初始化计数器
    TOTAL_TESTS=0
    PASSED_TESTS=0
    FAILED_TESTS=0
    
    # 定义测试套件
    declare -a TEST_SUITES=(
        "PRD符合度:prd_compliance_test.sh"
        "集成测试:integration_test.sh"
        "权限测试:test_admin_permissions.sh"
        "角色测试:test_role_permissions.sh"
        "预约测试:appointment_test.sh"
    )
    
    # 运行所有测试
    for test_suite in "${TEST_SUITES[@]}"; do
        IFS=':' read -r test_name script_name <<< "$test_suite"
        TOTAL_TESTS=$((TOTAL_TESTS + 1))
        
        if run_test_suite "$test_name" "$script_name"; then
            PASSED_TESTS=$((PASSED_TESTS + 1))
        else
            FAILED_TESTS=$((FAILED_TESTS + 1))
        fi
    done
    
    # 生成最终报告
    if [[ $FAILED_TESTS -eq 0 ]]; then
        STATUS="全部通过 ✅"
    else
        STATUS="部分失败 ⚠️"
    fi
    
    generate_report "$STATUS" "$TOTAL_TESTS" "$PASSED_TESTS" "$FAILED_TESTS"
    
    # 打印总结
    print_title "测试完成总结"
    echo -e "${BLUE}总测试数: $TOTAL_TESTS${NC}"
    echo -e "${GREEN}通过: $PASSED_TESTS${NC}"
    echo -e "${RED}失败: $FAILED_TESTS${NC}"
    
    if [[ $FAILED_TESTS -eq 0 ]]; then
        print_success "🎉 所有测试通过！系统质量良好"
    else
        print_error "⚠️  有 $FAILED_TESTS 个测试失败，请检查并修复"
    fi
}

# 命令行参数处理
case "${1:-all}" in
    "env"|"environment")
        check_environment
        ;;
    "unit")
        run_test_suite "单元测试" "unit_tests.sh"
        ;;
    "integration")
        run_test_suite "集成测试" "integration_test.sh"
        ;;
    "compliance")
        run_test_suite "PRD符合度" "prd_compliance_test.sh"
        ;;
    "security")
        run_test_suite "安全测试" "security_test.sh"
        ;;
    "all"|"")
        main
        ;;
    *)
        echo "使用方法: $0 [env|unit|integration|compliance|security|all]"
        exit 1
        ;;
esac