#!/bin/bash

# Museum Tables Migration Script
# 博物馆表迁移脚本

echo "🏛️ Starting Museum Tables Migration..."
echo "======================================="

# 设置颜色输出
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 切换到backend目录
cd "$(dirname "$0")/.." || exit 1

# 检查是否在backend目录
if [ ! -f "main.go" ]; then
    echo -e "${RED}Error: Not in backend directory${NC}"
    exit 1
fi

echo -e "${YELLOW}Running database migration...${NC}"

# 执行数据库迁移
go run main.go migrate

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Database migration completed successfully${NC}"
else
    echo -e "${RED}❌ Database migration failed${NC}"
    exit 1
fi

# 验证新表是否创建成功
echo -e "\n${YELLOW}Verifying new museum tables...${NC}"

# 检查PostgreSQL中的表
if command -v psql &> /dev/null; then
    echo -e "\n${YELLOW}Museum related tables in database:${NC}"
    psql -U $(whoami) -d openpenpal -c "\dt museum_*" 2>/dev/null || {
        echo -e "${YELLOW}Note: Could not verify tables via psql. Please check manually.${NC}"
    }
fi

echo -e "\n${GREEN}🎉 Museum tables migration completed!${NC}"
echo "======================================="
echo ""
echo "Next steps:"
echo "1. Verify tables in your database client"
echo "2. Test museum API endpoints"
echo "3. Check logs for any migration warnings"

# 显示最近的日志
echo -e "\n${YELLOW}Recent migration logs:${NC}"
tail -n 20 logs/app.log 2>/dev/null | grep -i "museum\|migration" || echo "No recent museum migration logs found"