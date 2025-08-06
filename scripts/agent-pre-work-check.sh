#!/bin/bash

# OpenPenPal Agent工作前检查脚本
# 确保Agent获取完整上下文后再开始工作

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔍 OpenPenPal Agent工作前检查${NC}"
echo "========================================"
echo ""

# 检查当前目录是否为项目根目录
if [ ! -f "README.md" ] || [ ! -d "docs" ]; then
    echo -e "${RED}❌ 错误: 请在项目根目录运行此脚本${NC}"
    exit 1
fi

# 1. 显示项目总览
echo -e "${YELLOW}📊 项目总览${NC}"
echo "----------------------------------------"
if [ -f "PROJECT_STATUS_DASHBOARD.yml" ]; then
    echo -e "${GREEN}✓ 项目状态仪表板可用${NC}"
    grep -A 5 "project:" PROJECT_STATUS_DASHBOARD.yml | sed 's/^/  /'
else
    echo -e "${RED}✗ 项目状态仪表板不存在${NC}"
fi
echo ""

# 2. 检查服务状态
echo -e "${YELLOW}🔧 服务状态概览${NC}"
echo "----------------------------------------"
if [ -f "PROJECT_STATUS_DASHBOARD.yml" ]; then
    echo "服务运行状态:"
    grep -A 20 "services:" PROJECT_STATUS_DASHBOARD.yml | grep -E "(status|health_check|completion)" | sed 's/^/  /'
else
    echo -e "${RED}✗ 无法获取服务状态${NC}"
fi
echo ""

# 3. 显示最新工作记录
echo -e "${YELLOW}📝 最新工作记录 (最近5条)${NC}"
echo "----------------------------------------"
if [ -f "LATEST_WORK_LOG.md" ]; then
    echo -e "${GREEN}✓ 工作日志可用${NC}"
    head -30 LATEST_WORK_LOG.md | tail -25
else
    echo -e "${RED}✗ 工作日志不存在${NC}"
fi
echo ""

# 4. 检查当前任务
echo -e "${YELLOW}📋 当前任务状态${NC}"
echo "----------------------------------------"
if [ -f "PROJECT_STATUS_DASHBOARD.yml" ]; then
    echo "进行中的任务:"
    grep -A 10 "in_progress:" PROJECT_STATUS_DASHBOARD.yml | sed 's/^/  /'
    echo ""
    echo "待处理的任务:"
    grep -A 10 "pending:" PROJECT_STATUS_DASHBOARD.yml | sed 's/^/  /'
else
    echo -e "${RED}✗ 无法获取任务状态${NC}"
fi
echo ""

# 5. 检查阻塞问题
echo -e "${YELLOW}⚠️  当前阻塞问题${NC}"
echo "----------------------------------------"
if [ -f "PROJECT_STATUS_DASHBOARD.yml" ]; then
    blockers=$(grep -A 5 "blockers:" PROJECT_STATUS_DASHBOARD.yml)
    if echo "$blockers" | grep -q "^\s*$" || echo "$blockers" | grep -q "blockers: \[\]"; then
        echo -e "${GREEN}✅ 当前无阻塞问题${NC}"
    else
        echo -e "${RED}❌ 发现阻塞问题:${NC}"
        echo "$blockers" | sed 's/^/  /'
    fi
else
    echo -e "${YELLOW}⚠️  无法检查阻塞问题${NC}"
fi
echo ""

# 6. 验证系统状态
echo -e "${YELLOW}🧪 系统验证状态${NC}"
echo "----------------------------------------"
if [ -f "SYSTEM_VERIFICATION_REPORT.md" ]; then
    echo -e "${GREEN}✓ 系统验证报告可用${NC}"
    echo "最新验证结果:"
    grep -E "(✅|❌|⚠️)" SYSTEM_VERIFICATION_REPORT.md | head -5 | sed 's/^/  /'
else
    echo -e "${YELLOW}⚠️  系统验证报告不存在${NC}"
fi
echo ""

# 7. 检查文档状态
echo -e "${YELLOW}📚 文档系统状态${NC}"
echo "----------------------------------------"
if [ -f "docs/README.md" ]; then
    echo -e "${GREEN}✓ 文档系统可用${NC}"
    if [ -f "scripts/check-doc-links.sh" ]; then
        echo "  正在检查文档链接..."
        ./scripts/check-doc-links.sh > /dev/null 2>&1
        if [ $? -eq 0 ]; then
            echo -e "  ${GREEN}✅ 所有文档链接有效${NC}"
        else
            echo -e "  ${RED}❌ 发现失效文档链接${NC}"
        fi
    fi
else
    echo -e "${RED}✗ 文档系统不可用${NC}"
fi
echo ""

# 8. 推荐的上下文文件
echo -e "${YELLOW}📖 推荐阅读的上下文文件${NC}"
echo "----------------------------------------"
echo "开始工作前，建议阅读以下文件:"
echo -e "  ${BLUE}1. README.md${NC} - 项目总览和快速启动"
echo -e "  ${BLUE}2. docs/README.md${NC} - 完整文档导航"
echo -e "  ${BLUE}3. docs/team-collaboration/context-management.md${NC} - 共享上下文"
echo -e "  ${BLUE}4. docs/team-collaboration/MULTI_AGENT_SYNC_SYSTEM.md${NC} - 协作机制"
echo -e "  ${BLUE}5. agent-tasks/AGENT-[YOUR-ID]-*.md${NC} - 你的具体任务"
echo -e "  ${BLUE}6. docs/api/README.md${NC} - API文档总览"
echo -e "  ${BLUE}7. docs/development/README.md${NC} - 开发规范"
echo ""

# 9. 快速命令提示
echo -e "${YELLOW}⚡ 快速命令${NC}"
echo "----------------------------------------"
echo "常用检查命令:"
echo -e "  ${BLUE}项目状态:${NC} cat PROJECT_STATUS_DASHBOARD.yml"
echo -e "  ${BLUE}工作日志:${NC} cat LATEST_WORK_LOG.md"
echo -e "  ${BLUE}服务状态:${NC} ./startup/check-status.sh"
echo -e "  ${BLUE}API测试:${NC} ./scripts/test-apis.sh"
echo -e "  ${BLUE}文档检查:${NC} ./scripts/check-doc-links.sh"
echo ""

# 10. 总结
echo "========================================"
echo -e "${GREEN}✅ 上下文检查完成！${NC}"
echo ""
echo -e "${BLUE}💡 下一步:${NC}"
echo "1. 仔细阅读相关文档"
echo "2. 确认没有任务冲突"
echo "3. 开始你的工作"
echo "4. 完成后运行: ./scripts/agent-post-work-sync.sh"
echo ""