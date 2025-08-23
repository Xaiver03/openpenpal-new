#!/bin/bash

# Quick Fix Script for Common Compilation Errors
# 快速修复常见编译错误脚本

echo "🔧 Fixing common compilation errors..."
echo "====================================="

# 设置颜色输出
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 切换到backend目录
cd "$(dirname "$0")/.." || exit 1

echo -e "${YELLOW}Fixing int vs int64 type mismatches...${NC}"

# Fix courier_dashboard_handler.go int to int64
if [ -f "internal/handlers/courier_dashboard_handler.go" ]; then
    # Replace var declarations from int to int64
    sed -i.bak 's/var todayTasks int/var todayTasks int64/g' internal/handlers/courier_dashboard_handler.go
    sed -i.bak 's/var completedTasks int/var completedTasks int64/g' internal/handlers/courier_dashboard_handler.go
    sed -i.bak 's/var pendingTasks int/var pendingTasks int64/g' internal/handlers/courier_dashboard_handler.go
    sed -i.bak 's/var teamMembers int/var teamMembers int64/g' internal/handlers/courier_dashboard_handler.go
    sed -i.bak 's/var count int/var count int64/g' internal/handlers/courier_dashboard_handler.go
    
    echo -e "${GREEN}✅ Fixed int64 type issues in courier_dashboard_handler.go${NC}"
fi

echo -e "${YELLOW}Attempting to compile...${NC}"

# Try to compile
go build -o /tmp/test-build main.go 2>&1 | head -20

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Compilation successful!${NC}"
    rm /tmp/test-build
else
    echo -e "${RED}❌ Still have compilation errors${NC}"
    echo ""
    echo "Remaining errors need manual intervention:"
    echo "1. Missing fields in CourierTask model (Type, DeliveryAddress)"
    echo "2. Missing fields in Courier model (CompletedTasks)"
    echo "3. Missing method in GinResponse (Forbidden)"
    echo ""
    echo "These require checking the model definitions and updating accordingly."
fi

# Clean up backup files
find . -name "*.bak" -delete

echo ""
echo "Next steps:"
echo "1. Manually fix remaining model/method issues"
echo "2. Run: go build main.go"
echo "3. If successful, run: ./main migrate"