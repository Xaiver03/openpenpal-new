#!/bin/bash
# 统一代码检查脚本

set -e

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}🔍 OpenPenPal 代码质量检查${NC}"

# 错误计数
ERROR_COUNT=0

# 前端代码检查
echo -e "\n${YELLOW}📱 前端代码检查${NC}"
if [ -d "frontend" ]; then
    cd frontend
    
    echo "运行 ESLint..."
    if npm run lint; then
        echo -e "${GREEN}✅ ESLint 检查通过${NC}"
    else
        echo -e "${RED}❌ ESLint 检查失败${NC}"
        ((ERROR_COUNT++))
    fi
    
    echo "运行 TypeScript 类型检查..."
    if npm run type-check; then
        echo -e "${GREEN}✅ TypeScript 检查通过${NC}"
    else
        echo -e "${RED}❌ TypeScript 检查失败${NC}"
        ((ERROR_COUNT++))
    fi
    
    cd ..
else
    echo -e "${YELLOW}⚠️ 前端目录不存在，跳过检查${NC}"
fi

# Go代码检查
echo -e "\n${YELLOW}🐹 Go代码检查${NC}"

# 检查 golangci-lint 是否安装
if ! command -v golangci-lint &> /dev/null; then
    echo "安装 golangci-lint..."
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.55.2
fi

# 后端主服务
if [ -d "backend" ]; then
    echo "检查后端主服务..."
    cd backend
    if golangci-lint run; then
        echo -e "${GREEN}✅ 后端主服务检查通过${NC}"
    else
        echo -e "${RED}❌ 后端主服务检查失败${NC}"
        ((ERROR_COUNT++))
    fi
    cd ..
fi

# Courier服务
if [ -d "services/courier-service" ]; then
    echo "检查 Courier 服务..."
    cd services/courier-service
    if golangci-lint run; then
        echo -e "${GREEN}✅ Courier服务检查通过${NC}"
    else
        echo -e "${RED}❌ Courier服务检查失败${NC}"
        ((ERROR_COUNT++))
    fi
    cd ../..
fi

# Python代码检查
echo -e "\n${YELLOW}🐍 Python代码检查${NC}"

# 检查Python工具是否安装
check_python_tool() {
    if ! python -m $1 --version &> /dev/null; then
        echo "安装 $1..."
        pip install $1
    fi
}

check_python_tool "black"
check_python_tool "isort"
check_python_tool "flake8"
check_python_tool "mypy"
check_python_tool "bandit"

# Write服务
if [ -d "services/write-service" ]; then
    echo "检查 Write 服务..."
    cd services/write-service
    
    # Black格式检查
    if python -m black --check .; then
        echo -e "${GREEN}✅ Black 格式检查通过${NC}"
    else
        echo -e "${RED}❌ Black 格式检查失败${NC}"
        ((ERROR_COUNT++))
    fi
    
    # isort导入排序检查
    if python -m isort --check-only .; then
        echo -e "${GREEN}✅ isort 检查通过${NC}"
    else
        echo -e "${RED}❌ isort 检查失败${NC}"
        ((ERROR_COUNT++))
    fi
    
    # flake8代码质量检查
    if python -m flake8; then
        echo -e "${GREEN}✅ flake8 检查通过${NC}"
    else
        echo -e "${RED}❌ flake8 检查失败${NC}"
        ((ERROR_COUNT++))
    fi
    
    # MyPy类型检查
    if python -m mypy .; then
        echo -e "${GREEN}✅ MyPy 类型检查通过${NC}"
    else
        echo -e "${RED}❌ MyPy 类型检查失败${NC}"
        ((ERROR_COUNT++))
    fi
    
    # Bandit安全检查
    if python -m bandit -r .; then
        echo -e "${GREEN}✅ Bandit 安全检查通过${NC}"
    else
        echo -e "${RED}❌ Bandit 安全检查失败${NC}"
        ((ERROR_COUNT++))
    fi
    
    cd ../..
fi

# OCR服务
if [ -d "services/ocr-service" ]; then
    echo "检查 OCR 服务..."
    cd services/ocr-service
    
    # 运行相同的Python检查
    python -m black --check . && \
    python -m isort --check-only . && \
    python -m flake8 && \
    python -m mypy . && \
    python -m bandit -r .
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ OCR服务检查通过${NC}"
    else
        echo -e "${RED}❌ OCR服务检查失败${NC}"
        ((ERROR_COUNT++))
    fi
    
    cd ../..
fi

# Java代码检查
echo -e "\n${YELLOW}☕ Java代码检查${NC}"
if [ -d "services/admin-service" ]; then
    echo "检查 Admin 服务..."
    cd services/admin-service
    
    # Maven检查
    if ./mvnw spotless:check; then
        echo -e "${GREEN}✅ Java代码格式检查通过${NC}"
    else
        echo -e "${RED}❌ Java代码格式检查失败${NC}"
        ((ERROR_COUNT++))
    fi
    
    # SpotBugs检查
    if ./mvnw spotbugs:check; then
        echo -e "${GREEN}✅ SpotBugs检查通过${NC}"
    else
        echo -e "${RED}❌ SpotBugs检查失败${NC}"
        ((ERROR_COUNT++))
    fi
    
    cd ../..
fi

# 通用文件检查
echo -e "\n${YELLOW}📄 通用文件检查${NC}"

# 检查是否有TODO或FIXME
echo "检查待办事项..."
TODO_COUNT=$(grep -r "TODO\|FIXME\|XXX" --include="*.js" --include="*.ts" --include="*.tsx" --include="*.go" --include="*.py" --include="*.java" . | wc -l)
if [ $TODO_COUNT -gt 0 ]; then
    echo -e "${YELLOW}⚠️ 发现 $TODO_COUNT 个待办事项${NC}"
    grep -r "TODO\|FIXME\|XXX" --include="*.js" --include="*.ts" --include="*.tsx" --include="*.go" --include="*.py" --include="*.java" . | head -10
    if [ $TODO_COUNT -gt 10 ]; then
        echo "... 还有 $((TODO_COUNT - 10)) 个"
    fi
fi

# 检查文件编码
echo "检查文件编码..."
if find . -name "*.js" -o -name "*.ts" -o -name "*.tsx" -o -name "*.go" -o -name "*.py" -o -name "*.java" | xargs file | grep -v "UTF-8\|ASCII"; then
    echo -e "${RED}❌ 发现非UTF-8编码文件${NC}"
    ((ERROR_COUNT++))
else
    echo -e "${GREEN}✅ 文件编码检查通过${NC}"
fi

# 检查行尾符
echo "检查行尾符..."
if find . -name "*.js" -o -name "*.ts" -o -name "*.tsx" -o -name "*.go" -o -name "*.py" -o -name "*.java" | xargs grep -l $'\r'; then
    echo -e "${RED}❌ 发现Windows行尾符(CRLF)${NC}"
    ((ERROR_COUNT++))
else
    echo -e "${GREEN}✅ 行尾符检查通过${NC}"
fi

# 总结
echo -e "\n${BLUE}📊 检查总结${NC}"
if [ $ERROR_COUNT -eq 0 ]; then
    echo -e "${GREEN}🎉 所有代码质量检查通过！${NC}"
    exit 0
else
    echo -e "${RED}❌ 发现 $ERROR_COUNT 个问题需要修复${NC}"
    echo -e "${YELLOW}💡 提示：运行 'make format' 可以自动修复格式问题${NC}"
    exit 1
fi