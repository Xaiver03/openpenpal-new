#!/bin/bash

# OpenPenPal Agent工作完成后同步脚本
# 确保Agent及时同步工作成果和关键信息

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 参数检查
if [ $# -lt 3 ]; then
    echo -e "${RED}用法: $0 <Agent-ID> <服务名称> <工作摘要>${NC}"
    echo ""
    echo "示例:"
    echo "  $0 \"Agent-1\" \"frontend\" \"完成用户界面优化\""
    echo "  $0 \"Agent-2\" \"write-service\" \"新增批量操作功能\""
    echo ""
    exit 1
fi

AGENT_ID="$1"
SERVICE_NAME="$2"
WORK_SUMMARY="$3"
TIMESTAMP=$(date '+%Y-%m-%d %H:%M')
TIMESTAMP_FILE=$(date '+%Y%m%d_%H%M')

echo -e "${BLUE}📤 OpenPenPal Agent工作同步${NC}"
echo "========================================"
echo -e "Agent: ${YELLOW}${AGENT_ID}${NC}"
echo -e "服务: ${YELLOW}${SERVICE_NAME}${NC}"
echo -e "时间: ${YELLOW}${TIMESTAMP}${NC}"
echo ""

# 检查当前目录
if [ ! -f "README.md" ] || [ ! -d "docs" ]; then
    echo -e "${RED}❌ 错误: 请在项目根目录运行此脚本${NC}"
    exit 1
fi

# 创建工作日志目录
mkdir -p work-logs

# 1. 更新最新工作日志
echo -e "${YELLOW}📝 更新工作日志...${NC}"
{
    echo ""
    echo "### ${AGENT_ID} (${SERVICE_NAME}) - ${TIMESTAMP}"
    echo "- ✅ ${WORK_SUMMARY}"
    
    # 获取最近的Git提交
    if git rev-parse --git-dir > /dev/null 2>&1; then
        echo "- 📋 最近提交:"
        git log --oneline -3 | sed 's/^/  - /'
    fi
    
    # 检查是否有配置或API变更
    if git diff --name-only HEAD~1 HEAD 2>/dev/null | grep -E "\.(yml|yaml|json|env)$" > /dev/null; then
        echo "- ⚠️  包含配置文件变更，请其他Agent检查兼容性"
    fi
    
    if git diff --name-only HEAD~1 HEAD 2>/dev/null | grep -E "api|route|controller" > /dev/null; then
        echo "- ⚠️  可能包含API变更，请其他Agent检查接口兼容性"
    fi
    
} >> LATEST_WORK_LOG.md

echo -e "${GREEN}✓ 工作日志已更新${NC}"

# 2. 创建详细工作摘要
echo -e "${YELLOW}📋 创建工作摘要...${NC}"
cat > "work-logs/WORK_SUMMARY_${AGENT_ID}_${TIMESTAMP_FILE}.md" << EOF
# ${AGENT_ID} 工作摘要

**Agent ID**: ${AGENT_ID}
**服务**: ${SERVICE_NAME}
**工作时间**: ${TIMESTAMP}
**任务**: ${WORK_SUMMARY}

## 🔧 主要变更
EOF

# 自动检测变更文件
if git rev-parse --git-dir > /dev/null 2>&1; then
    echo "检测到以下文件变更:" >> "work-logs/WORK_SUMMARY_${AGENT_ID}_${TIMESTAMP_FILE}.md"
    git diff --name-only HEAD~1 HEAD 2>/dev/null | while read file; do
        if [ -f "$file" ]; then
            echo "- 📄 修改: \`$file\`" >> "work-logs/WORK_SUMMARY_${AGENT_ID}_${TIMESTAMP_FILE}.md"
        fi
    done
    
    echo "" >> "work-logs/WORK_SUMMARY_${AGENT_ID}_${TIMESTAMP_FILE}.md"
    echo "## 📊 Git提交记录" >> "work-logs/WORK_SUMMARY_${AGENT_ID}_${TIMESTAMP_FILE}.md"
    git log --oneline -5 | sed 's/^/- /' >> "work-logs/WORK_SUMMARY_${AGENT_ID}_${TIMESTAMP_FILE}.md"
fi

cat >> "work-logs/WORK_SUMMARY_${AGENT_ID}_${TIMESTAMP_FILE}.md" << EOF

## 📊 影响评估
- **影响的服务**: ${SERVICE_NAME}
- **API兼容性**: 需要其他Agent验证
- **数据库变更**: 请检查
- **配置依赖**: 请检查

## 🧪 测试状态
- [ ] 单元测试: 请运行测试验证
- [ ] 集成测试: 请运行集成测试
- [ ] API测试: 运行 \`./scripts/test-apis.sh\`

## 🔗 相关文件
- 详细变更请查看Git提交记录
- 文档更新请查看docs目录变更

---
**生成时间**: ${TIMESTAMP}
**同步脚本**: agent-post-work-sync.sh
EOF

echo -e "${GREEN}✓ 工作摘要已创建: work-logs/WORK_SUMMARY_${AGENT_ID}_${TIMESTAMP_FILE}.md${NC}"

# 3. 更新项目状态仪表板
echo -e "${YELLOW}📊 更新项目状态仪表板...${NC}"
if [ -f "PROJECT_STATUS_DASHBOARD.yml" ]; then
    # 备份原文件
    cp PROJECT_STATUS_DASHBOARD.yml PROJECT_STATUS_DASHBOARD.yml.bak
    
    # 更新时间戳和操作者
    sed -i.tmp "s/last_updated: .*/last_updated: \"$(date -u '+%Y-%m-%dT%H:%M:%SZ')\"/" PROJECT_STATUS_DASHBOARD.yml
    sed -i.tmp "s/updated_by: .*/updated_by: \"${AGENT_ID}\"/" PROJECT_STATUS_DASHBOARD.yml
    
    # 更新服务的最后修改时间
    if grep -q "${SERVICE_NAME}:" PROJECT_STATUS_DASHBOARD.yml; then
        sed -i.tmp "/${SERVICE_NAME}:/,/last_modified:/ s/last_modified: .*/last_modified: \"$(date -u '+%Y-%m-%dT%H:%M:%SZ')\"/" PROJECT_STATUS_DASHBOARD.yml
    fi
    
    # 清理临时文件
    rm -f PROJECT_STATUS_DASHBOARD.yml.tmp
    
    echo -e "${GREEN}✓ 项目状态仪表板已更新${NC}"
else
    echo -e "${YELLOW}⚠️  项目状态仪表板不存在，跳过更新${NC}"
fi

# 4. 运行兼容性检查
echo -e "${YELLOW}🔍 运行兼容性检查...${NC}"
if [ -f "scripts/test-apis.sh" ]; then
    echo "正在测试API接口..."
    if ./scripts/test-apis.sh > /tmp/api-test-results.log 2>&1; then
        echo -e "${GREEN}✅ API测试通过${NC}"
    else
        echo -e "${RED}❌ API测试失败，请检查兼容性${NC}"
        echo "详细结果:"
        tail -10 /tmp/api-test-results.log | sed 's/^/  /'
    fi
else
    echo -e "${YELLOW}⚠️  API测试脚本不存在，跳过测试${NC}"
fi

# 5. 检查文档一致性
echo -e "${YELLOW}📚 检查文档一致性...${NC}"
if [ -f "scripts/check-doc-links.sh" ]; then
    if ./scripts/check-doc-links.sh > /dev/null 2>&1; then
        echo -e "${GREEN}✅ 文档链接检查通过${NC}"
    else
        echo -e "${RED}❌ 发现失效文档链接${NC}"
        echo "请运行: ./scripts/check-doc-links.sh 查看详情"
    fi
else
    echo -e "${YELLOW}⚠️  文档链接检查脚本不存在${NC}"
fi

# 6. 提醒更新相关文档
echo -e "${YELLOW}📖 文档更新提醒${NC}"
echo "----------------------------------------"
echo "请确保以下文档已更新:"
echo -e "  ${BLUE}1.${NC} agent-tasks/AGENT-[YOUR-ID]-*.md - 你的任务状态"
echo -e "  ${BLUE}2.${NC} docs/team-collaboration/context-management.md - 共享上下文"

# 检查是否有API或架构变更
if echo "$WORK_SUMMARY" | grep -iE "(api|接口|架构|数据库)" > /dev/null; then
    echo -e "  ${BLUE}3.${NC} docs/api/ - API文档 (如有API变更)"
    echo -e "  ${BLUE}4.${NC} docs/architecture/ - 架构文档 (如有架构变更)"
fi

echo ""

# 7. 生成给其他Agent的通知
echo -e "${YELLOW}📢 生成协作通知...${NC}"
cat > "work-logs/NOTIFICATION_${AGENT_ID}_${TIMESTAMP_FILE}.md" << EOF
# 🔔 Agent协作通知

**来自**: ${AGENT_ID}
**时间**: ${TIMESTAMP}
**服务**: ${SERVICE_NAME}

## 📝 工作摘要
${WORK_SUMMARY}

## ⚠️  需要其他Agent注意
EOF

# 基于变更类型生成具体通知
if echo "$WORK_SUMMARY" | grep -iE "(api|接口)" > /dev/null; then
    echo "- 🔌 **API变更**: 请检查你的服务是否需要适配新的接口" >> "work-logs/NOTIFICATION_${AGENT_ID}_${TIMESTAMP_FILE}.md"
fi

if echo "$WORK_SUMMARY" | grep -iE "(数据库|schema|migration)" > /dev/null; then
    echo "- 🗄️ **数据库变更**: 请检查是否需要运行数据迁移" >> "work-logs/NOTIFICATION_${AGENT_ID}_${TIMESTAMP_FILE}.md"
fi

if echo "$WORK_SUMMARY" | grep -iE "(配置|config|env)" > /dev/null; then
    echo "- ⚙️ **配置变更**: 请检查是否需要更新配置文件" >> "work-logs/NOTIFICATION_${AGENT_ID}_${TIMESTAMP_FILE}.md"
fi

if echo "$WORK_SUMMARY" | grep -iE "(前端|ui|界面)" > /dev/null; then
    echo "- 🎨 **前端变更**: 后端Agent请注意接口兼容性" >> "work-logs/NOTIFICATION_${AGENT_ID}_${TIMESTAMP_FILE}.md"
fi

cat >> "work-logs/NOTIFICATION_${AGENT_ID}_${TIMESTAMP_FILE}.md" << EOF

## 🔗 相关资源
- 详细工作摘要: work-logs/WORK_SUMMARY_${AGENT_ID}_${TIMESTAMP_FILE}.md
- 最新工作日志: LATEST_WORK_LOG.md
- 项目状态: PROJECT_STATUS_DASHBOARD.yml

## 🧪 建议检查
1. 运行 \`./scripts/test-apis.sh\` 验证API兼容性
2. 运行 \`./startup/check-status.sh\` 检查服务状态
3. 查看最新的工作日志了解整体进展

---
**自动生成于**: ${TIMESTAMP}
EOF

echo -e "${GREEN}✓ 协作通知已生成: work-logs/NOTIFICATION_${AGENT_ID}_${TIMESTAMP_FILE}.md${NC}"

# 8. 总结
echo ""
echo "========================================"
echo -e "${GREEN}✅ 工作同步完成！${NC}"
echo ""
echo -e "${BLUE}📋 已完成的同步任务:${NC}"
echo "  ✓ 更新工作日志"
echo "  ✓ 创建详细工作摘要"
echo "  ✓ 更新项目状态仪表板"
echo "  ✓ 运行兼容性检查"
echo "  ✓ 生成协作通知"
echo ""
echo -e "${BLUE}📝 还需要手动完成:${NC}"
echo "  - 更新你的Agent任务文档"
echo "  - 更新相关的API或架构文档"
echo "  - 如有重大变更，通知团队"
echo ""
echo -e "${YELLOW}💡 提醒:${NC} 其他Agent可以通过以下方式了解你的工作:"
echo "  - 查看 LATEST_WORK_LOG.md"
echo "  - 查看 work-logs/WORK_SUMMARY_${AGENT_ID}_${TIMESTAMP_FILE}.md"
echo "  - 查看 work-logs/NOTIFICATION_${AGENT_ID}_${TIMESTAMP_FILE}.md"
echo ""