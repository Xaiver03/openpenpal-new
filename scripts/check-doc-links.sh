#!/bin/bash

# OpenPenPal 增强版文档链接检查脚本
# 用于检查文档中的链接有效性和内容同步性

set -e

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 统计变量
total_links=0
valid_links=0
invalid_links=0
external_links=0
anchor_links=0
port_issues=0
file_ref_issues=0

echo -e "${BLUE}🔍 OpenPenPal 文档链接检查${NC}"
echo "========================================"
echo ""

# 获取文档根目录
if [ -n "$1" ]; then
    DOC_ROOT="$1"
else
    DOC_ROOT="docs"
fi

# 检查目录是否存在
if [ ! -d "$DOC_ROOT" ]; then
    echo -e "${RED}错误: 目录 $DOC_ROOT 不存在${NC}"
    exit 1
fi

# 临时文件存储无效链接
INVALID_LINKS_FILE=$(mktemp)

# 查找所有 markdown 文件
find "$DOC_ROOT" -name "*.md" -type f | sort | while read -r file; do
    echo -e "${YELLOW}检查文件: $file${NC}"
    
    # 提取所有链接
    grep -oE '\[([^]]+)\]\(([^)]+)\)' "$file" | while read -r link; do
        # 提取链接文本和目标
        link_text=$(echo "$link" | sed -E 's/\[([^]]+)\].*/\1/')
        link_target=$(echo "$link" | sed -E 's/.*\(([^)]+)\).*/\1/')
        
        ((total_links++))
        
        # 跳过锚点链接
        if [[ "$link_target" == "#"* ]]; then
            echo -e "  ${BLUE}⚓ 锚点链接: $link_target${NC}"
            ((anchor_links++))
            continue
        fi
        
        # 检查外部链接
        if [[ "$link_target" == http* ]] || [[ "$link_target" == https* ]]; then
            echo -e "  ${BLUE}🌐 外部链接: $link_target${NC}"
            ((external_links++))
            continue
        fi
        
        # 检查相对路径
        if [[ "$link_target" == "./"* ]] || [[ "$link_target" == "../"* ]]; then
            # 计算实际路径
            file_dir=$(dirname "$file")
            actual_path=$(cd "$file_dir" && realpath "$link_target" 2>/dev/null)
            
            if [ -e "$actual_path" ]; then
                echo -e "  ${GREEN}✓ $link_text → $link_target${NC}"
                ((valid_links++))
            else
                echo -e "  ${RED}✗ $link_text → $link_target (文件不存在)${NC}"
                echo "$file: $link_text → $link_target" >> "$INVALID_LINKS_FILE"
                ((invalid_links++))
            fi
        else
            # 绝对路径或其他情况
            if [ -e "$link_target" ]; then
                echo -e "  ${GREEN}✓ $link_text → $link_target${NC}"
                ((valid_links++))
            else
                echo -e "  ${RED}✗ $link_text → $link_target (文件不存在)${NC}"
                echo "$file: $link_text → $link_target" >> "$INVALID_LINKS_FILE"
                ((invalid_links++))
            fi
        fi
    done
    echo ""
done

# 额外检查：端口信息一致性
echo -e "${CYAN}🔌 检查端口信息一致性...${NC}"
echo "========================================"

# 检查常用端口在代码和文档中的一致性
for port in 3000 8000 8001 8002 8003 8004; do
    code_has_port=$(grep -r ":$port\|PORT.*$port" --include="*.js" --include="*.ts" --include="*.go" --include="*.py" . 2>/dev/null | wc -l)
    doc_has_port=$(grep -r "$port" --include="*.md" . 2>/dev/null | wc -l)
    
    if [ "$code_has_port" -gt 0 ] && [ "$doc_has_port" -eq 0 ]; then
        echo -e "${YELLOW}⚠️ 端口 $port 在代码中使用但文档中未提及${NC}"
        ((port_issues++))
    elif [ "$code_has_port" -eq 0 ] && [ "$doc_has_port" -gt 0 ]; then
        echo -e "${YELLOW}⚠️ 端口 $port 在文档中提及但代码中未使用${NC}"
        ((port_issues++))
    elif [ "$code_has_port" -gt 0 ] && [ "$doc_has_port" -gt 0 ]; then
        echo -e "${GREEN}✓ 端口 $port 在代码和文档中一致${NC}"
    fi
done

# 额外检查：文件引用
echo ""
echo -e "${CYAN}📁 检查脚本文件引用...${NC}"
echo "========================================"

grep -r "\.sh\|\.command" --include="*.md" . 2>/dev/null | while IFS: read file content; do
    # 提取脚本文件名
    scripts=$(echo "$content" | grep -oE '[^[:space:]`"'"'"']*\.(sh|command)' | head -5)
    echo "$scripts" | while read script; do
        if [ -n "$script" ] && [[ ! "$script" =~ ^http ]]; then
            found=false
            # 在多个位置查找脚本
            for prefix in "" "./" "./scripts/" "./startup/" "../"; do
                if [ -f "${prefix}${script}" ]; then
                    found=true
                    break
                fi
            done
            
            if [ "$found" = false ]; then
                echo -e "${YELLOW}⚠️ 文件引用可能无效: $script (在 $file)${NC}"
                ((file_ref_issues++))
            fi
        fi
    done
done

# 显示统计结果
echo ""
echo "========================================"
echo -e "${BLUE}📊 完整检查结果统计${NC}"
echo "========================================"
echo -e "总链接数: ${total_links}"
echo -e "有效链接: ${GREEN}${valid_links}${NC}"
echo -e "无效链接: ${RED}${invalid_links}${NC}"
echo -e "外部链接: ${BLUE}${external_links}${NC}"
echo -e "锚点链接: ${CYAN}${anchor_links}${NC}"
echo -e "端口问题: ${YELLOW}${port_issues}${NC}"
echo -e "文件引用问题: ${YELLOW}${file_ref_issues}${NC}"
echo ""

# 总问题数
total_issues=$((invalid_links + port_issues + file_ref_issues))

# 如果有无效链接，显示汇总
if [ $invalid_links -gt 0 ]; then
    echo -e "${RED}❌ 发现 $invalid_links 个无效链接:${NC}"
    echo "----------------------------------------"
    cat "$INVALID_LINKS_FILE"
    echo ""
fi

# 显示建议
if [ $total_issues -gt 0 ]; then
    echo -e "${YELLOW}💡 修复建议:${NC}"
    echo "----------------------------------------"
    [ $invalid_links -gt 0 ] && echo "🔗 修复无效链接：检查文件路径，更新已移动的文件引用"
    [ $port_issues -gt 0 ] && echo "🔌 统一端口信息：确保代码和文档中的端口信息一致"
    [ $file_ref_issues -gt 0 ] && echo "📁 检查文件引用：确认引用的脚本文件确实存在"
    echo ""
    echo -e "${CYAN}🔧 自动修复命令:${NC}"
    echo "   git status                    # 查看文件变更"
    echo "   find . -name '*.md' -exec grep -l 'broken_link' {} \;"
    echo "   ./scripts/fix-doc-links.sh   # 运行自动修复脚本"
else
    echo -e "${GREEN}🎉 所有检查都通过了！文档质量很好！${NC}"
fi

# 清理临时文件
rm -f "$INVALID_LINKS_FILE"

# 返回状态码
exit $total_issues