#!/bin/bash

# OpenPenPal 文档快速查找工具
# 帮助用户快速找到需要的文档

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
BOLD='\033[1m'
NC='\033[0m' # No Color

# 使用说明
show_usage() {
    echo -e "${BLUE}📚 OpenPenPal 文档查找工具${NC}"
    echo "========================================"
    echo ""
    echo "用法: $0 [关键词]"
    echo ""
    echo "示例:"
    echo "  $0 api          # 查找API相关文档"
    echo "  $0 启动         # 查找启动相关文档"  
    echo "  $0 测试         # 查找测试相关文档"
    echo "  $0 agent        # 查找Agent相关文档"
    echo ""
    echo "或者运行: $0 menu  # 显示分类菜单"
    echo ""
}

# 显示分类菜单
show_menu() {
    echo -e "${BLUE}📚 OpenPenPal 文档分类菜单${NC}"
    echo "========================================"
    echo ""
    echo -e "${YELLOW}🎯 最常用文档${NC}"
    echo "  1. 项目总览      - README.md"
    echo "  2. 快速启动      - QUICK_START.md"  
    echo "  3. 统一文档中心   - UNIFIED_DOC_CENTER.md"
    echo "  4. 项目状态      - PROJECT_STATUS_DASHBOARD.yml"
    echo "  5. 测试账号      - docs/getting-started/test-accounts.md"
    echo ""
    echo -e "${YELLOW}🛠️ 开发文档${NC}"
    echo "  6. API文档       - docs/api/README.md"
    echo "  7. 开发规范      - docs/development/README.md"
    echo "  8. 系统架构      - docs/architecture/README.md"
    echo "  9. 故障排查      - docs/troubleshooting/"
    echo ""
    echo -e "${YELLOW}👥 协作文档${NC}"
    echo "  10. Agent协作    - docs/team-collaboration/MULTI_AGENT_SYNC_SYSTEM.md"
    echo "  11. Agent任务    - agent-tasks/README.md"
    echo "  12. 共享上下文    - docs/team-collaboration/context-management.md"
    echo ""
    read -p "请选择文档编号 (1-12): " choice
    
    case $choice in
        1) cat README.md ;;
        2) cat QUICK_START.md ;;
        3) cat UNIFIED_DOC_CENTER.md ;;
        4) cat PROJECT_STATUS_DASHBOARD.yml ;;
        5) cat docs/getting-started/test-accounts.md ;;
        6) cat docs/api/README.md ;;
        7) cat docs/development/README.md ;;
        8) cat docs/architecture/README.md ;;
        9) ls -la docs/troubleshooting/ && echo -e "\n选择具体文件查看" ;;
        10) cat docs/team-collaboration/MULTI_AGENT_SYNC_SYSTEM.md ;;
        11) cat agent-tasks/README.md ;;
        12) cat docs/team-collaboration/context-management.md ;;
        *) echo -e "${RED}无效选择${NC}" ;;
    esac
}

# 搜索文档
search_docs() {
    local keyword="$1"
    echo -e "${BLUE}🔍 搜索关键词: \"$keyword\"${NC}"
    echo "========================================"
    echo ""
    
    # 在文档中搜索
    echo -e "${YELLOW}📝 在文档内容中搜索:${NC}"
    find . -name "*.md" -type f | head -20 | while read file; do
        if grep -l -i "$keyword" "$file" 2>/dev/null; then
            echo -e "  ${GREEN}✓${NC} $file"
            # 显示匹配的行
            grep -n -i "$keyword" "$file" | head -2 | sed 's/^/    /'
            echo ""
        fi
    done
    
    echo ""
    echo -e "${YELLOW}📂 在文件名中搜索:${NC}"
    find . -name "*$keyword*" -type f | head -10 | while read file; do
        echo -e "  ${GREEN}✓${NC} $file"
    done
    
    echo ""
    echo -e "${YELLOW}📁 在目录名中搜索:${NC}"
    find . -name "*$keyword*" -type d | head -10 | while read dir; do
        echo -e "  ${GREEN}✓${NC} $dir/"
    done
}

# 智能建议
suggest_docs() {
    local keyword="$1"
    echo ""
    echo -e "${YELLOW}💡 相关建议:${NC}"
    
    case "$keyword" in
        *api*|*API*|*接口*)
            echo "  → 查看 docs/api/README.md (API总览)"
            echo "  → 查看 docs/api/unified-specification.md (API规范)"
            echo "  → 运行 ./scripts/test-apis.sh (测试API)"
            ;;
        *启动*|*start*|*launch*)
            echo "  → 查看 QUICK_START.md (快速启动)"
            echo "  → 运行 ./startup/quick-start.sh demo (启动项目)"
            echo "  → 查看 startup/ 目录 (启动脚本)"
            ;;
        *测试*|*test*|*账号*|*account*)
            echo "  → 查看 docs/getting-started/test-accounts.md (测试账号)"
            echo "  → 查看 test-kimi/ 目录 (测试套件)"
            echo "  → 运行 ./scripts/test-apis.sh (API测试)"
            ;;
        *agent*|*Agent*|*协作*)
            echo "  → 查看 agent-tasks/README.md (Agent任务)"
            echo "  → 查看 docs/team-collaboration/ (协作文档)"
            echo "  → 运行 ./scripts/agent-pre-work-check.sh (工作检查)"
            ;;
        *问题*|*错误*|*故障*|*troubl*)
            echo "  → 查看 docs/troubleshooting/ (故障排查)"
            echo "  → 运行 ./startup/check-status.sh (检查状态)"
            echo "  → 运行 ./startup/force-cleanup.sh (重置环境)"
            ;;
        *部署*|*deploy*|*运维*)
            echo "  → 查看 docs/deployment/README.md (部署指南)"
            echo "  → 查看 docs/operations/ (运维文档)"
            echo "  → 查看 docker-compose.yml (容器配置)"
            ;;
        *)
            echo "  → 查看 UNIFIED_DOC_CENTER.md (统一文档中心)"
            echo "  → 查看 docs/README.md (完整文档导航)"
            echo "  → 运行 $0 menu (显示分类菜单)"
            ;;
    esac
}

# 主逻辑
main() {
    # 检查是否在项目根目录
    if [ ! -f "README.md" ] || [ ! -d "docs" ]; then
        echo -e "${RED}❌ 错误: 请在项目根目录运行此脚本${NC}"
        exit 1
    fi
    
    # 处理参数
    if [ $# -eq 0 ]; then
        show_usage
        exit 0
    fi
    
    case "$1" in
        -h|--help|help)
            show_usage
            ;;
        menu)
            show_menu
            ;;
        *)
            search_docs "$1"
            suggest_docs "$1"
            ;;
    esac
}

# 运行主函数
main "$@"