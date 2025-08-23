#!/bin/bash
# 安全扫描脚本 - 全面的安全检查工具
# 针对微服务架构的安全扫描，包括代码、依赖、镜像、配置等
set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 项目根目录
PROJECT_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
SCAN_RESULTS_DIR="$PROJECT_ROOT/.security-scan-results"

# 创建结果目录
mkdir -p "$SCAN_RESULTS_DIR"

echo -e "${BLUE}🔒 开始全面安全扫描...${NC}"

# 1. 代码静态安全分析
echo -e "${YELLOW}📝 1/8 代码静态安全分析...${NC}"

# Go 代码安全扫描
if command -v gosec >/dev/null 2>&1; then
    echo "   扫描 Go 代码安全问题..."
    cd "$PROJECT_ROOT/backend"
    gosec -fmt json -out "$SCAN_RESULTS_DIR/gosec-report.json" ./... || true
    gosec ./... > "$SCAN_RESULTS_DIR/gosec-report.txt" || true
else
    echo "   ⚠️  gosec 未安装，跳过 Go 安全扫描"
    echo "   安装: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"
fi

# TypeScript/JavaScript 安全扫描
if command -v eslint >/dev/null 2>&1; then
    echo "   扫描前端代码安全问题..."
    cd "$PROJECT_ROOT/frontend"
    npx eslint . --ext .ts,.tsx,.js,.jsx --format json --output-file "$SCAN_RESULTS_DIR/eslint-security.json" || true
fi

# 2. 依赖漏洞扫描
echo -e "${YELLOW}📦 2/8 依赖漏洞扫描...${NC}"

# Go 依赖漏洞扫描
if command -v govulncheck >/dev/null 2>&1; then
    echo "   扫描 Go 依赖漏洞..."
    cd "$PROJECT_ROOT/backend"
    govulncheck -json ./... > "$SCAN_RESULTS_DIR/go-vulns.json" 2>/dev/null || true
fi

# Node.js 依赖漏洞扫描
echo "   扫描 Node.js 依赖漏洞..."
cd "$PROJECT_ROOT/frontend"
npm audit --json > "$SCAN_RESULTS_DIR/npm-audit.json" 2>/dev/null || true
npm audit > "$SCAN_RESULTS_DIR/npm-audit.txt" 2>/dev/null || true

# Python 依赖漏洞扫描
for service in write-service ocr-service; do
    if [ -d "$PROJECT_ROOT/services/$service" ]; then
        echo "   扫描 $service Python 依赖漏洞..."
        cd "$PROJECT_ROOT/services/$service"
        if command -v safety >/dev/null 2>&1; then
            safety check --json > "$SCAN_RESULTS_DIR/$service-safety.json" 2>/dev/null || true
        fi
        if command -v pip-audit >/dev/null 2>&1; then
            pip-audit --format=json > "$SCAN_RESULTS_DIR/$service-pip-audit.json" 2>/dev/null || true
        fi
    fi
done

# 3. 容器镜像安全扫描
echo -e "${YELLOW}🐳 3/8 容器镜像安全扫描...${NC}"

if command -v docker >/dev/null 2>&1; then
    # 构建测试镜像
    cd "$PROJECT_ROOT"
    
    # Frontend 镜像扫描
    if docker build -t openpenpal-frontend:security-scan -f frontend/Dockerfile frontend/ >/dev/null 2>&1; then
        echo "   扫描前端镜像..."
        if command -v trivy >/dev/null 2>&1; then
            trivy image --format json --output "$SCAN_RESULTS_DIR/frontend-image-scan.json" openpenpal-frontend:security-scan 2>/dev/null || true
        fi
    fi
    
    # Backend 镜像扫描
    if docker build -t openpenpal-backend:security-scan -f backend/Dockerfile backend/ >/dev/null 2>&1; then
        echo "   扫描后端镜像..."
        if command -v trivy >/dev/null 2>&1; then
            trivy image --format json --output "$SCAN_RESULTS_DIR/backend-image-scan.json" openpenpal-backend:security-scan 2>/dev/null || true
        fi
    fi
fi

# 4. 秘钥和敏感信息扫描
echo -e "${YELLOW}🔑 4/8 秘钥和敏感信息扫描...${NC}"

if command -v truffleHog >/dev/null 2>&1; then
    echo "   扫描代码中的秘钥..."
    cd "$PROJECT_ROOT"
    truffleHog --json filesystem . > "$SCAN_RESULTS_DIR/secrets-scan.json" 2>/dev/null || true
elif command -v gitleaks >/dev/null 2>&1; then
    echo "   扫描代码中的秘钥..."
    cd "$PROJECT_ROOT"
    gitleaks detect --source . --report-format json --report-path "$SCAN_RESULTS_DIR/gitleaks-report.json" || true
fi

# 检查常见的敏感文件
echo "   检查敏感文件..."
{
    echo "=== 潜在敏感文件 ==="
    find "$PROJECT_ROOT" -type f \( -name "*.pem" -o -name "*.key" -o -name ".env*" -o -name "*secret*" -o -name "*password*" \) 2>/dev/null || true
    echo ""
    echo "=== .env 文件内容检查 ==="
    find "$PROJECT_ROOT" -name ".env*" -exec echo "文件: {}" \; -exec grep -H "PASSWORD\|SECRET\|KEY\|TOKEN" {} \; 2>/dev/null || true
} > "$SCAN_RESULTS_DIR/sensitive-files.txt"

# 5. 网络安全配置检查
echo -e "${YELLOW}🌐 5/8 网络安全配置检查...${NC}"

{
    echo "=== Docker Compose 安全配置检查 ==="
    echo "检查端口暴露..."
    find "$PROJECT_ROOT" -name "docker-compose*.yml" -exec echo "文件: {}" \; -exec grep -n "ports:" {} \; 2>/dev/null || true
    echo ""
    echo "检查网络配置..."
    find "$PROJECT_ROOT" -name "docker-compose*.yml" -exec echo "文件: {}" \; -exec grep -A 5 "networks:" {} \; 2>/dev/null || true
    echo ""
    echo "检查环境变量暴露..."
    find "$PROJECT_ROOT" -name "docker-compose*.yml" -exec echo "文件: {}" \; -exec grep -A 10 "environment:" {} \; 2>/dev/null || true
} > "$SCAN_RESULTS_DIR/network-security.txt"

# 6. 配置文件安全检查
echo -e "${YELLOW}⚙️ 6/8 配置文件安全检查...${NC}"

{
    echo "=== Nginx 配置安全检查 ==="
    find "$PROJECT_ROOT" -name "*.conf" -o -name "nginx.conf" | while read conf_file; do
        echo "检查文件: $conf_file"
        if grep -q "ssl_protocols" "$conf_file"; then
            echo "  ✅ SSL 协议已配置"
        else
            echo "  ⚠️  未找到 SSL 协议配置"
        fi
        if grep -q "add_header.*X-Frame-Options" "$conf_file"; then
            echo "  ✅ X-Frame-Options 头已配置"
        else
            echo "  ⚠️  未配置 X-Frame-Options 安全头"
        fi
    done
    
    echo ""
    echo "=== 数据库配置安全检查 ==="
    find "$PROJECT_ROOT" -name "*.yml" -o -name "*.yaml" -o -name "*.json" | xargs grep -l "database\|postgres\|mysql" 2>/dev/null | while read db_config; do
        echo "检查文件: $db_config"
        if grep -q "ssl.*true\|sslmode.*require" "$db_config"; then
            echo "  ✅ 数据库 SSL 已启用"
        else
            echo "  ⚠️  数据库 SSL 可能未启用"
        fi
    done
} > "$SCAN_RESULTS_DIR/config-security.txt"

# 7. API 安全检查
echo -e "${YELLOW}🔌 7/8 API 安全检查...${NC}"

{
    echo "=== API 安全配置检查 ==="
    echo "检查 CORS 配置..."
    find "$PROJECT_ROOT" -name "*.go" -o -name "*.py" -o -name "*.js" -o -name "*.ts" | xargs grep -l "CORS\|cors" 2>/dev/null | head -10
    echo ""
    echo "检查认证中间件..."
    find "$PROJECT_ROOT" -name "*.go" -o -name "*.py" -o -name "*.js" -o -name "*.ts" | xargs grep -l "JWT\|auth\|token" 2>/dev/null | head -10
    echo ""
    echo "检查输入验证..."
    find "$PROJECT_ROOT" -name "*.go" -o -name "*.py" -o -name "*.js" -o -name "*.ts" | xargs grep -l "validate\|sanitize" 2>/dev/null | head -10
} > "$SCAN_RESULTS_DIR/api-security.txt"

# 8. 生成安全报告
echo -e "${YELLOW}📊 8/8 生成安全报告...${NC}"

cat > "$SCAN_RESULTS_DIR/security-summary.md" << 'EOF'
# OpenPenPal 安全扫描报告

## 扫描时间
EOF

echo "扫描时间: $(date)" >> "$SCAN_RESULTS_DIR/security-summary.md"
echo "" >> "$SCAN_RESULTS_DIR/security-summary.md"

cat >> "$SCAN_RESULTS_DIR/security-summary.md" << 'EOF'
## 扫描结果汇总

### 🔍 代码安全分析
- Go 代码安全: 查看 `gosec-report.txt`
- 前端代码安全: 查看 `eslint-security.json`

### 📦 依赖漏洞扫描  
- Go 依赖: 查看 `go-vulns.json`
- Node.js 依赖: 查看 `npm-audit.txt`
- Python 依赖: 查看 `*-safety.json`, `*-pip-audit.json`

### 🐳 容器镜像安全
- 前端镜像: 查看 `frontend-image-scan.json`  
- 后端镜像: 查看 `backend-image-scan.json`

### 🔑 秘钥和敏感信息
- 秘钥扫描: 查看 `secrets-scan.json` 或 `gitleaks-report.json`
- 敏感文件: 查看 `sensitive-files.txt`

### 🌐 网络和配置安全
- 网络配置: 查看 `network-security.txt`
- 配置文件: 查看 `config-security.txt` 
- API 安全: 查看 `api-security.txt`

## 推荐的安全工具安装

```bash
# Go 安全工具
go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
go install golang.org/x/vuln/cmd/govulncheck@latest

# 容器安全扫描
# Trivy (推荐)
brew install trivy
# 或 curl -sfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh | sh -s -- -b /usr/local/bin

# 秘钥扫描
# TruffleHog
pip install truffleHog
# 或 GitLeaks  
brew install gitleaks

# Python 安全工具
pip install safety pip-audit
```

## 安全最佳实践建议

### 1. 代码安全
- [ ] 定期运行静态代码安全分析
- [ ] 实施代码审查流程  
- [ ] 使用安全的编码标准

### 2. 依赖管理
- [ ] 定期更新依赖版本
- [ ] 使用依赖锁定文件
- [ ] 监控已知漏洞

### 3. 容器安全
- [ ] 使用最小化的基础镜像
- [ ] 以非 root 用户运行
- [ ] 定期扫描镜像漏洞

### 4. 网络安全  
- [ ] 启用 HTTPS/TLS
- [ ] 配置适当的 CORS 策略
- [ ] 使用防火墙和网络分段

### 5. 认证和授权
- [ ] 实施强密码策略
- [ ] 使用多因素认证
- [ ] 定期轮换密钥和令牌

### 6. 监控和日志
- [ ] 启用安全事件日志
- [ ] 设置异常行为监控
- [ ] 定期审查访问日志
EOF

echo -e "${GREEN}✅ 安全扫描完成！${NC}"
echo -e "${BLUE}📊 扫描结果保存在: $SCAN_RESULTS_DIR${NC}"
echo -e "${BLUE}📋 查看汇总报告: cat $SCAN_RESULTS_DIR/security-summary.md${NC}"

# 检查高危问题
echo -e "${YELLOW}⚠️  正在检查高危安全问题...${NC}"

HIGH_RISK_FOUND=false

# 检查是否有高危漏洞
if [ -f "$SCAN_RESULTS_DIR/npm-audit.json" ]; then
    HIGH_COUNT=$(jq -r '.metadata.vulnerabilities.high // 0' "$SCAN_RESULTS_DIR/npm-audit.json" 2>/dev/null || echo "0")
    CRITICAL_COUNT=$(jq -r '.metadata.vulnerabilities.critical // 0' "$SCAN_RESULTS_DIR/npm-audit.json" 2>/dev/null || echo "0")
    if [ "$HIGH_COUNT" -gt 0 ] || [ "$CRITICAL_COUNT" -gt 0 ]; then
        echo -e "${RED}❌ 发现高危/严重漏洞: 高危 $HIGH_COUNT 个, 严重 $CRITICAL_COUNT 个${NC}"
        HIGH_RISK_FOUND=true
    fi
fi

# 检查敏感信息泄露
if [ -f "$SCAN_RESULTS_DIR/sensitive-files.txt" ] && [ -s "$SCAN_RESULTS_DIR/sensitive-files.txt" ]; then
    echo -e "${RED}❌ 发现潜在的敏感文件，请检查: $SCAN_RESULTS_DIR/sensitive-files.txt${NC}"
    HIGH_RISK_FOUND=true
fi

if [ "$HIGH_RISK_FOUND" = false ]; then
    echo -e "${GREEN}✅ 未发现高危安全问题${NC}"
fi

echo ""
echo -e "${BLUE}🔧 建议定期运行此脚本，特别是在以下情况：${NC}"
echo "   • 添加新依赖时"
echo "   • 发布新版本前"  
echo "   • 每周安全检查"
echo "   • CI/CD 流程中"