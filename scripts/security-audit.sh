#!/bin/bash

# OpenPenPal安全审计脚本
# 执行全面的安全检查和漏洞扫描

echo "🔐 开始OpenPenPal安全审计..."
echo "=========================="

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m'

# 统计变量
TOTAL_CHECKS=0
PASSED_CHECKS=0
FAILED_CHECKS=0
WARNINGS=0

# 函数：执行安全检查
run_check() {
    local check_name="$1"
    local check_command="$2"
    local expected_result="$3"
    
    TOTAL_CHECKS=$((TOTAL_CHECKS + 1))
    
    echo -e "${BLUE}🔍 检查: $check_name${NC}"
    
    if eval "$check_command"; then
        if [[ "$expected_result" == "success" ]]; then
            echo -e "   ✅ ${GREEN}通过${NC}"
            PASSED_CHECKS=$((PASSED_CHECKS + 1))
        else
            echo -e "   ⚠️ ${YELLOW}警告: 检查通过但可能存在问题${NC}"
            WARNINGS=$((WARNINGS + 1))
        fi
    else
        if [[ "$expected_result" == "fail" ]]; then
            echo -e "   ✅ ${GREEN}通过 (预期失败)${NC}"
            PASSED_CHECKS=$((PASSED_CHECKS + 1))
        else
            echo -e "   ❌ ${RED}失败${NC}"
            FAILED_CHECKS=$((FAILED_CHECKS + 1))
        fi
    fi
}

echo -e "${PURPLE}📋 第1步: 依赖安全检查${NC}"
echo "----------------------------------------"

# 检查Node.js依赖漏洞
if command -v npm &> /dev/null; then
    echo "🔍 检查Node.js依赖安全..."
    
    # 检查前端依赖
    if [[ -f "frontend/package.json" ]]; then
        cd frontend
        run_check "前端依赖漏洞扫描" "npm audit --audit-level=moderate" "success"
        cd ..
    fi
    
    # 检查根目录依赖
    if [[ -f "package.json" ]]; then
        run_check "根目录依赖漏洞扫描" "npm audit --audit-level=moderate" "success"
    fi
else
    echo "⚠️ npm未安装，跳过Node.js依赖检查"
fi

echo ""

echo -e "${PURPLE}📋 第2步: 文件权限检查${NC}"
echo "----------------------------------------"

# 检查敏感文件权限
echo "🔍 检查敏感文件权限..."

# 检查环境配置文件
for file in .env .env.local .env.production; do
    if [[ -f "$file" ]]; then
        perms=$(stat -f "%A" "$file" 2>/dev/null || stat -c "%a" "$file" 2>/dev/null)
        if [[ "$perms" =~ ^6[0-4][0-4]$ ]]; then
            run_check "环境文件权限($file)" "true" "success"
        else
            run_check "环境文件权限($file)" "false" "success"
            echo "   📋 建议设置: chmod 600 $file"
        fi
    fi
done

# 检查脚本文件权限
for script in scripts/*.sh; do
    if [[ -f "$script" ]]; then
        perms=$(stat -f "%A" "$script" 2>/dev/null || stat -c "%a" "$script" 2>/dev/null)
        if [[ "$perms" =~ ^75[0-5]$ ]]; then
            run_check "脚本文件权限($script)" "true" "success"
        else
            run_check "脚本文件权限($script)" "false" "success"
            echo "   📋 建议设置: chmod 750 $script"
        fi
    fi
done

echo ""

echo -e "${PURPLE}📋 第3步: 配置安全检查${NC}"
echo "----------------------------------------"

# 检查Docker配置安全
echo "🔍 检查Docker配置安全..."

# 检查Dockerfile安全
for dockerfile in Dockerfile*/Dockerfile*; do
    if [[ -f "$dockerfile" ]]; then
        # 检查是否使用非root用户
        if grep -q "USER" "$dockerfile"; then
            run_check "Dockerfile非root用户($dockerfile)" "true" "success"
        else
            run_check "Dockerfile非root用户($dockerfile)" "false" "success"
            echo "   📋 建议: 添加非root用户配置"
        fi
        
        # 检查是否有健康检查
        if grep -q "HEALTHCHECK" "$dockerfile"; then
            run_check "Dockerfile健康检查($dockerfile)" "true" "success"
        else
            run_check "Dockerfile健康检查($dockerfile)" "false" "success"
            echo "   📋 建议: 添加健康检查配置"
        fi
    fi
done

# 检查docker-compose配置
for compose_file in docker-compose*.yml; do
    if [[ -f "$compose_file" ]]; then
        # 检查是否暴露了不必要的端口
        exposed_ports=$(grep -c "ports:" "$compose_file" 2>/dev/null || echo 0)
        if [[ $exposed_ports -lt 5 ]]; then
            run_check "Docker端口暴露($compose_file)" "true" "success"
        else
            run_check "Docker端口暴露($compose_file)" "false" "success"
            echo "   📋 建议: 减少暴露的端口数量"
        fi
        
        # 检查是否使用了secrets
        if grep -q "secrets:" "$compose_file"; then
            run_check "Docker Secrets使用($compose_file)" "true" "success"
        else
            run_check "Docker Secrets使用($compose_file)" "false" "success"
            echo "   📋 建议: 使用Docker Secrets管理敏感数据"
        fi
    fi
done

echo ""

echo -e "${PURPLE}📋 第4步: 代码安全扫描${NC}"
echo "----------------------------------------"

echo "🔍 代码安全模式检查..."

# 检查硬编码密码
echo "🔑 检查硬编码凭据..."
hardcoded_patterns=(
    "password.*=.*['\"][^'\"]{8,}['\"]"
    "api[_-]?key.*=.*['\"][^'\"]{16,}['\"]"
    "secret.*=.*['\"][^'\"]{16,}['\"]"
    "token.*=.*['\"][^'\"]{20,}['\"]"
)

for pattern in "${hardcoded_patterns[@]}"; do
    matches=$(grep -r -i -E "$pattern" --include="*.js" --include="*.ts" --include="*.go" --include="*.java" --include="*.py" . 2>/dev/null | grep -v node_modules | grep -v .git || true)
    if [[ -n "$matches" ]]; then
        run_check "硬编码凭据检查" "false" "success"
        echo "   📋 发现可能的硬编码凭据:"
        echo "$matches" | head -3
        break
    else
        run_check "硬编码凭据检查" "true" "success"
    fi
done

# 检查SQL注入风险
echo "💉 检查SQL注入风险..."
sql_injection_patterns=(
    "query.*\+.*req\."
    "SELECT.*\+.*input"
    "UPDATE.*\+.*params"
    "DELETE.*\+.*body"
)

sql_risk_found=false
for pattern in "${sql_injection_patterns[@]}"; do
    matches=$(grep -r -i -E "$pattern" --include="*.js" --include="*.ts" --include="*.go" --include="*.java" --include="*.py" . 2>/dev/null | grep -v node_modules | grep -v .git || true)
    if [[ -n "$matches" ]]; then
        sql_risk_found=true
        break
    fi
done

if [[ "$sql_risk_found" == "true" ]]; then
    run_check "SQL注入风险检查" "false" "success"
    echo "   📋 建议: 使用参数化查询"
else
    run_check "SQL注入风险检查" "true" "success"
fi

# 检查XSS风险
echo "🔗 检查XSS风险..."
xss_patterns=(
    "innerHTML.*=.*req\."
    "document\.write.*req\."
    "eval.*req\."
    "dangerouslySetInnerHTML"
)

xss_risk_found=false
for pattern in "${xss_patterns[@]}"; do
    matches=$(grep -r -i -E "$pattern" --include="*.js" --include="*.ts" --include="*.jsx" --include="*.tsx" . 2>/dev/null | grep -v node_modules | grep -v .git || true)
    if [[ -n "$matches" ]]; then
        xss_risk_found=true
        break
    fi
done

if [[ "$xss_risk_found" == "true" ]]; then
    run_check "XSS风险检查" "false" "success"
    echo "   📋 建议: 使用安全的DOM操作方法"
else
    run_check "XSS风险检查" "true" "success"
fi

echo ""

echo -e "${PURPLE}📋 第5步: 网络安全检查${NC}"
echo "----------------------------------------"

echo "🌐 检查网络安全配置..."

# 检查HTTPS配置
if [[ -f "config/nginx.prod.conf" ]]; then
    if grep -q "ssl_certificate" "config/nginx.prod.conf"; then
        run_check "HTTPS配置检查" "true" "success"
    else
        run_check "HTTPS配置检查" "false" "success"
        echo "   📋 建议: 配置SSL证书"
    fi
    
    # 检查安全头
    security_headers=("X-Frame-Options" "X-XSS-Protection" "X-Content-Type-Options" "Strict-Transport-Security")
    for header in "${security_headers[@]}"; do
        if grep -q "$header" "config/nginx.prod.conf"; then
            run_check "安全头配置($header)" "true" "success"
        else
            run_check "安全头配置($header)" "false" "success"
            echo "   📋 建议: 添加 $header 安全头"
        fi
    done
fi

# 检查CORS配置
cors_configs=$(find . -name "*.js" -o -name "*.ts" -o -name "*.go" -o -name "*.java" | xargs grep -l -i "cors" 2>/dev/null | grep -v node_modules || true)
if [[ -n "$cors_configs" ]]; then
    run_check "CORS配置检查" "true" "success"
else
    run_check "CORS配置检查" "false" "success"
    echo "   📋 建议: 配置适当的CORS策略"
fi

echo ""

echo -e "${PURPLE}📋 第6步: 数据库安全检查${NC}"
echo "----------------------------------------"

echo "🗄️ 检查数据库安全配置..."

# 检查数据库连接配置
db_configs=(
    ".env*"
    "config/*.yml"
    "config/*.yaml"
    "*.properties"
)

for config_pattern in "${db_configs[@]}"; do
    for config_file in $config_pattern; do
        if [[ -f "$config_file" ]]; then
            # 检查是否使用了默认密码
            if grep -i -E "(password.*=.*(admin|root|password|123456|))" "$config_file" 2>/dev/null; then
                run_check "数据库弱密码检查($config_file)" "false" "success"
                echo "   📋 建议: 使用强密码"
            else
                run_check "数据库弱密码检查($config_file)" "true" "success"
            fi
            
            # 检查是否启用了SSL
            if grep -i -E "(ssl.*=.*true|sslmode.*=.*require)" "$config_file" 2>/dev/null; then
                run_check "数据库SSL配置($config_file)" "true" "success"
            else
                run_check "数据库SSL配置($config_file)" "false" "success"
                echo "   📋 建议: 启用数据库SSL连接"
            fi
        fi
    done
done

echo ""

echo -e "${GREEN}🎊 安全审计完成${NC}"
echo "=================="

# 计算得分
if [[ $TOTAL_CHECKS -gt 0 ]]; then
    security_score=$((PASSED_CHECKS * 100 / TOTAL_CHECKS))
else
    security_score=0
fi

echo "📊 安全审计统计:"
echo "   总检查项: $TOTAL_CHECKS"
echo -e "   通过检查: ${GREEN}$PASSED_CHECKS${NC}"
echo -e "   失败检查: ${RED}$FAILED_CHECKS${NC}"
echo -e "   警告数量: ${YELLOW}$WARNINGS${NC}"
echo -e "   安全得分: ${GREEN}${security_score}%${NC}"

echo ""
echo -e "${BLUE}🎯 安全等级评估:${NC}"
if [[ $security_score -ge 90 ]]; then
    echo -e "   ${GREEN}✅ 优秀 (≥90%)${NC} - 安全配置良好"
elif [[ $security_score -ge 80 ]]; then
    echo -e "   ${GREEN}✅ 良好 (80-89%)${NC} - 大部分安全措施到位"
elif [[ $security_score -ge 70 ]]; then
    echo -e "   ${YELLOW}⚠️ 一般 (70-79%)${NC} - 需要改进安全配置"
elif [[ $security_score -ge 60 ]]; then
    echo -e "   ${YELLOW}⚠️ 较差 (60-69%)${NC} - 存在较多安全风险"
else
    echo -e "   ${RED}❌ 危险 (<60%)${NC} - 安全风险严重"
fi

echo ""
echo -e "${YELLOW}📋 安全改进建议:${NC}"
echo "1. 定期运行安全审计"
echo "2. 更新依赖包到最新版本"
echo "3. 实施代码安全扫描"
echo "4. 配置Web应用防火墙(WAF)"
echo "5. 实施API限流和防刷机制"
echo "6. 定期进行渗透测试"

# 生成安全报告
report_file="logs/security_audit_$(date +%Y%m%d_%H%M%S).txt"
cat > "$report_file" << EOF
OpenPenPal安全审计报告
=====================

审计时间: $(date)
总检查项: $TOTAL_CHECKS
通过检查: $PASSED_CHECKS
失败检查: $FAILED_CHECKS
警告数量: $WARNINGS
安全得分: ${security_score}%

详细报告已保存到日志文件。
建议定期执行安全审计以确保系统安全。
EOF

echo ""
echo -e "${GREEN}✨ 安全审计报告已保存到: $report_file${NC}"