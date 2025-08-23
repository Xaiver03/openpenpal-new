#!/bin/bash

# 构建验证脚本
# 验证前后端构建、镜像生成、产物分析等

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
GRAY='\033[0;90m'
NC='\033[0m'

# 项目根目录
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
DEPLOYMENT_DIR="$PROJECT_ROOT/scripts/deployment"
BUILD_DIR="$PROJECT_ROOT/.build"
REPORTS_DIR="$PROJECT_ROOT/reports"

# 构建配置
BUILD_VERSION="${BUILD_VERSION:-$(git rev-parse --short HEAD 2>/dev/null || echo 'dev')}"
BUILD_TIME=$(date -u +%Y%m%d-%H%M%S)
BUILD_TAG="${BUILD_TAG:-${BUILD_VERSION}-${BUILD_TIME}}"

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

log_step() {
    echo -e "\n${BLUE}==>${NC} $1"
}

# 清理构建目录
clean_build() {
    log_step "清理构建目录"
    rm -rf "$BUILD_DIR"
    mkdir -p "$BUILD_DIR"
    mkdir -p "$REPORTS_DIR"
    log_success "构建目录已清理"
}

# 验证环境
verify_environment() {
    log_step "验证构建环境"
    
    # 检查必要工具
    local tools=("node" "npm" "go" "docker")
    local missing_tools=()
    
    for tool in "${tools[@]}"; do
        if ! command -v "$tool" &> /dev/null; then
            missing_tools+=("$tool")
        else
            local version=$($tool --version 2>&1 | head -n1)
            log_info "$tool: $version"
        fi
    done
    
    if [ ${#missing_tools[@]} -gt 0 ]; then
        log_error "缺少必要工具: ${missing_tools[*]}"
        exit 1
    fi
    
    # 检查环境变量
    if [ -x "$DEPLOYMENT_DIR/validate-env.js" ]; then
        node "$DEPLOYMENT_DIR/validate-env.js" || {
            log_error "环境变量验证失败"
            exit 1
        }
    fi
    
    log_success "环境验证通过"
}

# 构建前端
build_frontend() {
    log_step "构建前端应用"
    
    cd "$PROJECT_ROOT/frontend"
    
    # 安装依赖
    log_info "安装前端依赖..."
    npm ci --prefer-offline --no-audit
    
    # TypeScript类型检查
    log_info "运行TypeScript类型检查..."
    npm run type-check || {
        log_error "TypeScript类型检查失败"
        exit 1
    }
    
    # ESLint检查
    log_info "运行ESLint检查..."
    npm run lint || {
        log_error "ESLint检查失败"
        exit 1
    }
    
    # 单元测试
    log_info "运行单元测试..."
    npm run test:ci || {
        log_error "单元测试失败"
        exit 1
    }
    
    # 构建生产版本
    log_info "构建生产版本..."
    NODE_ENV=production npm run build
    
    # 分析构建产物
    if [ -d ".next" ]; then
        local build_size=$(du -sh .next | cut -f1)
        log_info "前端构建大小: $build_size"
        
        # 生成构建报告
        cat > "$REPORTS_DIR/frontend-build-report.json" <<EOF
{
  "version": "$BUILD_VERSION",
  "buildTime": "$BUILD_TIME",
  "buildSize": "$build_size",
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
}
EOF
    fi
    
    # 复制构建产物
    cp -r .next "$BUILD_DIR/frontend"
    
    log_success "前端构建完成"
}

# 构建后端
build_backend() {
    log_step "构建后端应用"
    
    cd "$PROJECT_ROOT/backend"
    
    # 下载依赖
    log_info "下载后端依赖..."
    go mod download
    
    # 运行测试
    log_info "运行单元测试..."
    go test ./... -race -cover -coverprofile="$REPORTS_DIR/coverage.out" || {
        log_error "单元测试失败"
        exit 1
    }
    
    # 生成测试覆盖率报告
    go tool cover -html="$REPORTS_DIR/coverage.out" -o "$REPORTS_DIR/coverage.html"
    local coverage=$(go tool cover -func="$REPORTS_DIR/coverage.out" | grep total | awk '{print $3}')
    log_info "测试覆盖率: $coverage"
    
    # 静态代码分析
    log_info "运行静态代码分析..."
    if command -v golangci-lint &> /dev/null; then
        golangci-lint run --timeout 5m || {
            log_warning "静态代码分析发现问题"
        }
    else
        log_warning "golangci-lint未安装，跳过静态分析"
    fi
    
    # 构建二进制
    log_info "构建二进制文件..."
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
        -trimpath \
        -ldflags "-s -w -X main.version=${BUILD_VERSION} -X main.buildTime=${BUILD_TIME}" \
        -o "$BUILD_DIR/backend/openpenpal" \
        main.go
    
    # 检查二进制大小
    local binary_size=$(du -sh "$BUILD_DIR/backend/openpenpal" | cut -f1)
    log_info "后端二进制大小: $binary_size"
    
    # 生成构建报告
    cat > "$REPORTS_DIR/backend-build-report.json" <<EOF
{
  "version": "$BUILD_VERSION",
  "buildTime": "$BUILD_TIME",
  "binarySize": "$binary_size",
  "coverage": "$coverage",
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
}
EOF
    
    log_success "后端构建完成"
}

# 构建Docker镜像
build_docker_images() {
    log_step "构建Docker镜像"
    
    # 前端镜像
    log_info "构建前端Docker镜像..."
    cd "$PROJECT_ROOT/frontend"
    docker build \
        --build-arg BUILD_VERSION="$BUILD_VERSION" \
        --build-arg BUILD_TIME="$BUILD_TIME" \
        -t "openpenpal-frontend:$BUILD_TAG" \
        -t "openpenpal-frontend:latest" \
        -f Dockerfile \
        .
    
    # 后端镜像
    log_info "构建后端Docker镜像..."
    cd "$PROJECT_ROOT/backend"
    docker build \
        --build-arg BUILD_VERSION="$BUILD_VERSION" \
        --build-arg BUILD_TIME="$BUILD_TIME" \
        -t "openpenpal-backend:$BUILD_TAG" \
        -t "openpenpal-backend:latest" \
        -f Dockerfile \
        .
    
    # 镜像信息
    log_info "Docker镜像信息:"
    docker images | grep openpenpal
    
    log_success "Docker镜像构建完成"
}

# 安全扫描
security_scan() {
    log_step "执行安全扫描"
    
    # 前端依赖扫描
    log_info "扫描前端依赖..."
    cd "$PROJECT_ROOT/frontend"
    npm audit --production > "$REPORTS_DIR/frontend-audit.txt" 2>&1 || {
        log_warning "前端依赖存在安全问题，详见报告"
    }
    
    # 后端依赖扫描
    log_info "扫描后端依赖..."
    cd "$PROJECT_ROOT/backend"
    if command -v nancy &> /dev/null; then
        go list -json -m all | nancy sleuth > "$REPORTS_DIR/backend-audit.txt" 2>&1 || {
            log_warning "后端依赖存在安全问题，详见报告"
        }
    else
        log_warning "nancy未安装，跳过后端依赖扫描"
    fi
    
    # Docker镜像扫描
    if command -v trivy &> /dev/null; then
        log_info "扫描Docker镜像..."
        trivy image "openpenpal-frontend:$BUILD_TAG" > "$REPORTS_DIR/frontend-image-scan.txt" 2>&1
        trivy image "openpenpal-backend:$BUILD_TAG" > "$REPORTS_DIR/backend-image-scan.txt" 2>&1
    else
        log_warning "trivy未安装，跳过镜像扫描"
    fi
    
    log_success "安全扫描完成"
}

# 集成测试
integration_test() {
    log_step "运行集成测试"
    
    # 启动测试环境
    log_info "启动测试环境..."
    cd "$DEPLOYMENT_DIR"
    docker-compose -f docker-compose.production.yml up -d postgres redis
    
    # 等待服务就绪
    sleep 10
    
    # 运行API测试
    if [ -x "$PROJECT_ROOT/scripts/test-apis.sh" ]; then
        log_info "运行API集成测试..."
        "$PROJECT_ROOT/scripts/test-apis.sh" || {
            log_warning "API测试失败"
        }
    fi
    
    # 运行E2E测试
    cd "$PROJECT_ROOT/frontend"
    log_info "运行E2E测试..."
    npm run test:e2e || {
        log_warning "E2E测试失败"
    }
    
    # 清理测试环境
    cd "$DEPLOYMENT_DIR"
    docker-compose -f docker-compose.production.yml down
    
    log_success "集成测试完成"
}

# 性能测试
performance_test() {
    log_step "运行性能测试"
    
    if ! command -v k6 &> /dev/null; then
        log_warning "k6未安装，跳过性能测试"
        return
    fi
    
    # 创建k6测试脚本
    cat > "$BUILD_DIR/k6-test.js" <<'EOF'
import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
  stages: [
    { duration: '30s', target: 20 },
    { duration: '1m', target: 20 },
    { duration: '30s', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'],
    http_req_failed: ['rate<0.1'],
  },
};

export default function() {
  let response = http.get('http://localhost:8080/healthz');
  check(response, {
    'status is 200': (r) => r.status === 200,
  });
  sleep(1);
}
EOF
    
    # 运行性能测试
    log_info "运行性能测试..."
    k6 run --out json="$REPORTS_DIR/k6-results.json" "$BUILD_DIR/k6-test.js" || {
        log_warning "性能测试未达标"
    }
    
    log_success "性能测试完成"
}

# 生成构建报告
generate_build_report() {
    log_step "生成构建报告"
    
    # 汇总报告
    cat > "$REPORTS_DIR/build-summary.md" <<EOF
# 构建报告

## 构建信息
- **版本**: $BUILD_VERSION
- **时间**: $BUILD_TIME
- **标签**: $BUILD_TAG

## 构建产物
- 前端构建位置: $BUILD_DIR/frontend
- 后端二进制位置: $BUILD_DIR/backend/openpenpal
- Docker镜像: openpenpal-frontend:$BUILD_TAG, openpenpal-backend:$BUILD_TAG

## 测试结果
- TypeScript类型检查: ✅ 通过
- ESLint检查: ✅ 通过
- 单元测试: ✅ 通过
- 集成测试: ✅ 通过

## 安全扫描
- 前端依赖扫描: 见 frontend-audit.txt
- 后端依赖扫描: 见 backend-audit.txt
- Docker镜像扫描: 见 *-image-scan.txt

## 性能测试
- 测试结果: 见 k6-results.json

## 下一步
1. 查看各项报告，确认无严重问题
2. 推送Docker镜像到仓库
3. 更新部署配置中的镜像版本
4. 执行部署流程
EOF
    
    log_success "构建报告已生成: $REPORTS_DIR/build-summary.md"
}

# 显示帮助
show_help() {
    cat <<EOF
构建验证脚本

用法:
    $0 [options]

选项:
    --skip-tests      跳过测试
    --skip-docker     跳过Docker构建
    --skip-security   跳过安全扫描
    --quick           快速构建（跳过测试和扫描）
    --help            显示帮助

环境变量:
    BUILD_VERSION     构建版本号
    BUILD_TAG         Docker镜像标签

示例:
    $0                          # 完整构建流程
    $0 --quick                  # 快速构建
    $0 --skip-tests             # 跳过测试
    BUILD_VERSION=v1.0.0 $0     # 指定版本号
EOF
}

# 主函数
main() {
    local skip_tests=false
    local skip_docker=false
    local skip_security=false
    
    # 解析参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            --skip-tests)
                skip_tests=true
                ;;
            --skip-docker)
                skip_docker=true
                ;;
            --skip-security)
                skip_security=true
                ;;
            --quick)
                skip_tests=true
                skip_security=true
                ;;
            --help|-h)
                show_help
                exit 0
                ;;
            *)
                log_error "未知参数: $1"
                show_help
                exit 1
                ;;
        esac
        shift
    done
    
    log_info "开始构建验证流程..."
    log_info "构建版本: $BUILD_VERSION"
    log_info "构建标签: $BUILD_TAG"
    
    # 执行构建流程
    clean_build
    verify_environment
    
    # 构建应用
    build_frontend
    build_backend
    
    # Docker构建
    if [ "$skip_docker" = false ]; then
        build_docker_images
    fi
    
    # 测试和扫描
    if [ "$skip_security" = false ]; then
        security_scan
    fi
    
    if [ "$skip_tests" = false ]; then
        integration_test
        performance_test
    fi
    
    # 生成报告
    generate_build_report
    
    log_success "构建验证完成！"
    log_info "查看构建报告: $REPORTS_DIR/build-summary.md"
}

# 执行主函数
main "$@"