#!/bin/bash

# OpenPenPal 依赖管理器 - SOTA级别的依赖版本统一和冲突解决工具
# 设计原则：
# 1. 自动化依赖检测和版本比较
# 2. 智能冲突解决和版本统一
# 3. 安全漏洞扫描和自动修复
# 4. 详细的变更日志和影响分析

set -euo pipefail

# 获取脚本目录和项目根目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# 加载服务框架
source "$SCRIPT_DIR/common/service-framework.sh"

# ==============================================================================
# 全局变量和配置
# ==============================================================================

readonly STRATEGY_FILE="$PROJECT_ROOT/package-lock-strategy.json"
readonly TEMP_DIR="$PROJECT_ROOT/.tmp/dependency-analysis"
readonly REPORT_DIR="$PROJECT_ROOT/reports/dependencies"
readonly CHANGELOG_FILE="$PROJECT_ROOT/DEPENDENCY_CHANGELOG.md"

# 创建必要的目录
mkdir -p "$TEMP_DIR" "$REPORT_DIR"

# ==============================================================================
# 依赖发现和分析
# ==============================================================================

# 发现所有package.json文件
discover_package_files() {
    log_info "🔍 发现项目中的package.json文件..."
    
    local package_files=()
    
    # 搜索所有package.json文件，排除node_modules
    while IFS= read -r -d '' file; do
        # 过滤掉node_modules中的文件
        if [[ "$file" != *"node_modules"* ]]; then
            package_files+=("$file")
        fi
    done < <(find "$PROJECT_ROOT" -name "package.json" -type f -print0)
    
    log_info "发现 ${#package_files[@]} 个package.json文件："
    for file in "${package_files[@]}"; do
        local relative_path
        # 兼容macOS的相对路径计算
        relative_path=$(echo "$file" | sed "s|^$PROJECT_ROOT/||")
        log_info "  - $relative_path"
    done
    
    # 保存到临时文件
    printf '%s\n' "${package_files[@]}" > "$TEMP_DIR/package-files.txt"
    
    echo "${#package_files[@]}"
}

# 提取所有依赖信息
extract_dependencies() {
    log_info "📦 分析依赖信息..."
    
    local analysis_file="$TEMP_DIR/dependency-analysis.json"
    
    # 初始化分析文件
    echo '{"projects": {}, "summary": {"total_dependencies": 0, "duplicates": {}, "conflicts": {}}}' > "$analysis_file"
    
    # 分析每个package.json
    while IFS= read -r package_file; do
        if [ ! -f "$package_file" ]; then
            continue
        fi
        
        local project_name
        project_name=$(basename "$(dirname "$package_file")")
        if [ "$project_name" = "." ]; then
            project_name="root"
        fi
        
        log_debug "分析项目: $project_name ($package_file)"
        
        # 提取依赖信息
        if command_exists jq; then
            local deps_info
            deps_info=$(jq -r '
                {
                    name: (.name // "unknown"),
                    version: (.version // "0.0.0"),
                    file: "'"$package_file"'",
                    dependencies: (.dependencies // {}),
                    devDependencies: (.devDependencies // {}),
                    peerDependencies: (.peerDependencies // {})
                }
            ' "$package_file" 2>/dev/null || echo '{}')
            
            # 更新分析文件
            jq --argjson project "$deps_info" --arg name "$project_name" '
                .projects[$name] = $project
            ' "$analysis_file" > "$analysis_file.tmp" && mv "$analysis_file.tmp" "$analysis_file"
        fi
    done < "$TEMP_DIR/package-files.txt"
    
    log_success "✓ 依赖分析完成"
    echo "$analysis_file"
}

# 识别依赖冲突
identify_conflicts() {
    local analysis_file="$1"
    
    log_info "⚔️ 识别依赖冲突..."
    
    local conflicts_file="$TEMP_DIR/conflicts.json"
    
    # 使用jq分析冲突
    jq '
        .projects as $projects |
        {} as $conflicts |
        # 收集所有依赖
        reduce ($projects | to_entries[]) as $proj ({};
            reduce (
                $proj.value.dependencies // {}, 
                $proj.value.devDependencies // {}
            | to_entries[]
            ) as $dep (.;
                .[$dep.key] += [{
                    project: $proj.key,
                    version: $dep.value,
                    type: (if $proj.value.dependencies[$dep.key] then "production" else "development" end)
                }]
            )
        ) |
        # 识别版本冲突
        with_entries(
            select(.value | length > 1) |
            select(.value | map(.version) | unique | length > 1) |
            {
                key: .key,
                value: {
                    package: .key,
                    conflictCount: (.value | map(.version) | unique | length),
                    versions: (.value | map(.version) | unique),
                    usages: .value,
                    severity: (
                        if (.value | map(.type) | unique | contains(["production"]))
                        then "high"
                        else "medium"
                        end
                    )
                }
            }
        )
    ' "$analysis_file" > "$conflicts_file"
    
    local conflict_count
    conflict_count=$(jq 'length' "$conflicts_file")
    
    if [ "$conflict_count" -gt 0 ]; then
        log_warn "发现 $conflict_count 个依赖冲突"
        
        # 显示冲突详情
        jq -r '
            to_entries[] |
            "  - \(.key): \(.value.versions | join(" vs ")) (\(.value.severity) severity)"
        ' "$conflicts_file"
    else
        log_success "✓ 未发现依赖冲突"
    fi
    
    echo "$conflicts_file"
}

# ==============================================================================
# 版本解析和策略应用
# ==============================================================================

# 加载策略配置
load_strategy() {
    if [ ! -f "$STRATEGY_FILE" ]; then
        log_error "策略文件不存在: $STRATEGY_FILE"
        return 1
    fi
    
    if ! jq . "$STRATEGY_FILE" >/dev/null 2>&1; then
        log_error "策略文件格式无效: $STRATEGY_FILE"
        return 1
    fi
    
    log_debug "策略配置已加载"
}

# 生成版本解决方案
generate_resolution() {
    local conflicts_file="$1"
    
    log_info "🔧 生成版本解决方案..."
    
    local resolution_file="$TEMP_DIR/resolution.json"
    
    # 基于策略生成解决方案
    jq --slurpfile strategy "$STRATEGY_FILE" '
        . as $conflicts |
        $strategy[0] as $strat |
        
        # 为每个冲突生成解决方案
        with_entries({
            key: .key,
            value: (.value + {
                recommendedVersion: (
                    # 首先检查策略文件中的预定义解决方案
                    if $strat.conflictResolution[.key]
                    then $strat.conflictResolution[.key].resolution
                    # 如果是共享依赖，使用策略版本
                    elif $strat.sharedDependencies.runtimeDependencies[.key]
                    then $strat.sharedDependencies.runtimeDependencies[.key].version
                    elif $strat.sharedDependencies.developmentDependencies[.key]
                    then $strat.sharedDependencies.developmentDependencies[.key].version
                    # 否则选择最新的稳定版本
                    else (.value.versions | max)
                    end
                ),
                strategy: (
                    if $strat.conflictResolution[.key]
                    then $strat.conflictResolution[.key].strategy
                    else "upgrade"
                    end
                ),
                reason: (
                    if $strat.conflictResolution[.key]
                    then $strat.conflictResolution[.key].reason
                    else "统一版本以避免冲突"
                    end
                ),
                affectedProjects: [.value.usages[].project],
                impact: (
                    if .value.severity == "high" then "major"
                    else "minor"
                    end
                )
            })
        })
    ' "$conflicts_file" > "$resolution_file"
    
    log_success "✓ 解决方案已生成"
    echo "$resolution_file"
}

# ==============================================================================
# 自动修复和版本更新
# ==============================================================================

# 应用版本解决方案
apply_resolution() {
    local resolution_file="$1"
    local dry_run="${2:-false}"
    
    if [ "$dry_run" = "true" ]; then
        log_info "🔍 DRY RUN: 预览版本更新..."
    else
        log_info "🔄 应用版本解决方案..."
    fi
    
    local update_count=0
    local backup_dir="$TEMP_DIR/backup-$(date +%Y%m%d-%H%M%S)"
    
    if [ "$dry_run" = "false" ]; then
        mkdir -p "$backup_dir"
    fi
    
    # 处理每个解决方案
    jq -r '
        to_entries[] |
        @json
    ' "$resolution_file" | while IFS= read -r resolution_json; do
        local package_name
        local recommended_version
        local affected_projects
        local strategy
        local reason
        
        package_name=$(echo "$resolution_json" | jq -r '.key')
        recommended_version=$(echo "$resolution_json" | jq -r '.value.recommendedVersion')
        strategy=$(echo "$resolution_json" | jq -r '.value.strategy')
        reason=$(echo "$resolution_json" | jq -r '.value.reason')
        
        log_info "处理依赖: $package_name -> $recommended_version ($strategy)"
        log_debug "原因: $reason"
        
        # 获取受影响的项目
        while IFS= read -r project; do
            local package_file
            
            # 找到对应的package.json文件
            if [ "$project" = "root" ]; then
                package_file="$PROJECT_ROOT/package.json"
            else
                package_file=$(find "$PROJECT_ROOT" -name "package.json" -path "*/$project/*" | head -1)
            fi
            
            if [ ! -f "$package_file" ]; then
                log_warn "找不到项目的package.json: $project"
                continue
            fi
            
            if [ "$dry_run" = "true" ]; then
                log_info "  [DRY RUN] 将更新 $project: $package_name -> $recommended_version"
            else
                # 备份原文件
                cp "$package_file" "$backup_dir/$(basename "$package_file")-$project"
                
                # 更新package.json
                update_package_json "$package_file" "$package_name" "$recommended_version"
                
                log_success "  ✓ 已更新 $project: $package_name -> $recommended_version"
                update_count=$((update_count + 1))
            fi
            
        done < <(echo "$resolution_json" | jq -r '.value.affectedProjects[]')
    done
    
    if [ "$dry_run" = "false" ]; then
        log_success "✓ 完成 $update_count 个版本更新"
        log_info "备份文件位置: $backup_dir"
    fi
}

# 更新package.json文件
update_package_json() {
    local package_file="$1"
    local package_name="$2"
    local new_version="$3"
    
    # 使用jq更新package.json
    jq --arg pkg "$package_name" --arg ver "$new_version" '
        if .dependencies[$pkg] then
            .dependencies[$pkg] = $ver
        elif .devDependencies[$pkg] then
            .devDependencies[$pkg] = $ver
        else
            .
        end
    ' "$package_file" > "$package_file.tmp" && mv "$package_file.tmp" "$package_file"
}

# ==============================================================================
# 安全扫描和漏洞修复
# ==============================================================================

# 执行安全审计
security_audit() {
    log_info "🔒 执行安全审计..."
    
    local audit_file="$TEMP_DIR/security-audit.json"
    local findings_file="$TEMP_DIR/security-findings.json"
    
    # 运行npm audit
    if command_exists npm; then
        log_info "运行 npm audit..."
        
        # 收集所有项目的审计结果
        echo '{"projects": {}}' > "$audit_file"
        
        while IFS= read -r package_file; do
            local project_dir
            project_dir=$(dirname "$package_file")
            local project_name
            project_name=$(basename "$project_dir")
            
            if [ "$project_name" = "." ]; then
                project_name="root"
            fi
            
            log_debug "审计项目: $project_name"
            
            if [ -d "$project_dir/node_modules" ]; then
                local audit_result
                audit_result=$(cd "$project_dir" && npm audit --json 2>/dev/null || echo '{"vulnerabilities": {}, "metadata": {"vulnerabilities": {"total": 0}}}')
                
                # 添加到总结果中
                jq --argjson audit "$audit_result" --arg name "$project_name" '
                    .projects[$name] = $audit
                ' "$audit_file" > "$audit_file.tmp" && mv "$audit_file.tmp" "$audit_file"
            fi
        done < "$TEMP_DIR/package-files.txt"
        
        # 汇总安全发现
        jq '
            .projects |
            to_entries |
            map(.value.vulnerabilities // {} | to_entries) |
            flatten |
            group_by(.key) |
            map({
                package: .[0].key,
                severity: (.[0].value.severity // "unknown"),
                vulnerabilities: length,
                projects: [.[].value.via // []] | flatten | unique,
                fixAvailable: (.[0].value.fixAvailable // false)
            }) |
            sort_by(.severity == "critical", .severity == "high", .severity == "moderate")
        ' "$audit_file" > "$findings_file"
        
        local vuln_count
        vuln_count=$(jq 'length' "$findings_file")
        
        if [ "$vuln_count" -gt 0 ]; then
            log_warn "发现 $vuln_count 个安全漏洞"
            
            # 显示严重漏洞
            jq -r '
                .[] |
                select(.severity == "critical" or .severity == "high") |
                "  - \(.package): \(.severity) (\(.vulnerabilities) issues)"
            ' "$findings_file"
        else
            log_success "✓ 未发现安全漏洞"
        fi
    fi
    
    echo "$findings_file"
}

# 自动修复安全漏洞
fix_vulnerabilities() {
    local findings_file="$1"
    local auto_fix="${2:-false}"
    
    log_info "🔧 处理安全漏洞..."
    
    if [ "$auto_fix" = "true" ]; then
        log_info "自动修复安全漏洞..."
        
        while IFS= read -r package_file; do
            local project_dir
            project_dir=$(dirname "$package_file")
            
            if [ -d "$project_dir/node_modules" ]; then
                log_info "修复项目: $(basename "$project_dir")"
                
                (cd "$project_dir" && npm audit fix --force) || log_warn "自动修复失败: $(basename "$project_dir")"
            fi
        done < "$TEMP_DIR/package-files.txt"
        
        log_success "✓ 自动修复完成"
    else
        log_info "生成手动修复建议..."
        
        # 生成修复建议
        local suggestions_file="$REPORT_DIR/security-fix-suggestions.md"
        
        cat > "$suggestions_file" << EOF
# 安全漏洞修复建议

生成时间: $(date '+%Y-%m-%d %H:%M:%S')

## 发现的漏洞

EOF
        
        jq -r '
            .[] |
            "### \(.package) (\(.severity))\n" +
            "- 漏洞数量: \(.vulnerabilities)\n" +
            "- 影响项目: \(.projects | join(", "))\n" +
            "- 自动修复: \(if .fixAvailable then "可用" else "不可用" end)\n" +
            "- 建议操作: " + (
                if .fixAvailable then "运行 `npm audit fix`"
                elif .severity == "critical" or .severity == "high" then "立即手动更新到安全版本"
                else "计划在下次维护窗口更新"
                end
            ) + "\n"
        ' "$findings_file" >> "$suggestions_file"
        
        log_success "✓ 修复建议已生成: $suggestions_file"
    fi
}

# ==============================================================================
# 报告生成
# ==============================================================================

# 生成依赖报告
generate_report() {
    local analysis_file="$1"
    local conflicts_file="$2"
    local resolution_file="$3"
    local findings_file="$4"
    
    log_info "📊 生成依赖管理报告..."
    
    local report_file="$REPORT_DIR/dependency-report-$(date +%Y%m%d-%H%M%S).md"
    
    cat > "$report_file" << EOF
# OpenPenPal 依赖管理报告

生成时间: $(date '+%Y-%m-%d %H:%M:%S')
分析工具: dependency-manager.sh v1.0.0

## 📋 概要

EOF
    
    # 项目概要
    local project_count
    local total_deps
    local conflict_count
    local vuln_count
    
    project_count=$(jq '.projects | length' "$analysis_file")
    total_deps=$(jq '[.projects[].dependencies, .projects[].devDependencies] | add | length' "$analysis_file" 2>/dev/null || echo "0")
    conflict_count=$(jq 'length' "$conflicts_file")
    vuln_count=$(jq 'length' "$findings_file" 2>/dev/null || echo "0")
    
    cat >> "$report_file" << EOF
- 项目数量: $project_count
- 总依赖数: $total_deps
- 版本冲突: $conflict_count
- 安全漏洞: $vuln_count

## 🏗️ 项目结构

EOF
    
    # 项目列表
    jq -r '
        .projects |
        to_entries[] |
        "- **\(.key)**: \(.value.name) v\(.value.version)"
    ' "$analysis_file" >> "$report_file"
    
    # 冲突详情
    if [ "$conflict_count" -gt 0 ]; then
        cat >> "$report_file" << EOF

## ⚔️ 版本冲突

EOF
        
        jq -r '
            to_entries[] |
            "### \(.key)\n" +
            "\n" +
            "- **严重程度**: \(.value.severity)\n" +
            "- **冲突版本**: \(.value.versions | join(", "))\n" +
            "- **使用项目**: \(.value.usages | map(.project) | join(", "))\n" +
            "\n" +
            "**推荐解决方案**:\n" +
            "```bash\n" +
            "# 统一使用版本 \(.value.recommendedVersion // "待定")\n" +
            "npm install \(.key)@\(.value.recommendedVersion // "latest")\n" +
            "```\n"
        ' "$resolution_file" >> "$report_file" 2>/dev/null || true
    fi
    
    # 安全问题
    if [ "$vuln_count" -gt 0 ]; then
        cat >> "$report_file" << EOF

## 🔒 安全漏洞

EOF
        
        jq -r '
            .[] |
            "### \(.package) (\(.severity))\n" +
            "\n" +
            "- **漏洞数量**: \(.vulnerabilities)\n" +
            "- **影响项目**: \(.projects | join(", "))\n" +
            "- **自动修复**: \(if .fixAvailable then "✅ 可用" else "❌ 需手动处理" end)\n" +
            "\n"
        ' "$findings_file" >> "$report_file" 2>/dev/null || true
    fi
    
    # 建议操作
    cat >> "$report_file" << EOF

## 🎯 建议操作

### 立即处理
- 修复所有高危和严重安全漏洞
- 解决生产环境依赖冲突

### 计划处理
- 统一开发依赖版本
- 更新过时的依赖包

### 监控
- 定期执行安全审计
- 监控新的漏洞披露

## 📈 趋势分析

*注：需要历史数据来生成趋势分析*

---

*报告由 OpenPenPal 依赖管理器自动生成*
EOF
    
    log_success "✓ 报告已生成: $report_file"
    echo "$report_file"
}

# ==============================================================================
# 命令行接口
# ==============================================================================

# 显示帮助信息
show_help() {
    cat << EOF
OpenPenPal 依赖管理器

使用方法:
  $0 [命令] [选项]

命令:
  analyze    分析依赖冲突和版本问题
  fix        自动修复依赖冲突
  audit      执行安全审计
  report     生成依赖管理报告
  sync       同步所有项目的依赖版本
  clean      清理所有node_modules
  help       显示帮助信息

选项:
  --dry-run        预览模式，不实际修改文件
  --auto-fix       自动修复安全漏洞
  --force          强制执行操作
  --output DIR     指定输出目录

示例:
  $0 analyze                    # 分析依赖问题
  $0 fix --dry-run             # 预览修复方案
  $0 audit --auto-fix          # 审计并自动修复
  $0 report                    # 生成完整报告
  $0 sync                      # 同步依赖版本

EOF
}

# 主函数
main() {
    local command="analyze"
    local dry_run="false"
    local auto_fix="false"
    local force="false"
    local output_dir="$REPORT_DIR"
    
    # 解析命令行参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            analyze|fix|audit|report|sync|clean|help)
                command="$1"
                shift
                ;;
            --dry-run)
                dry_run="true"
                shift
                ;;
            --auto-fix)
                auto_fix="true"
                shift
                ;;
            --force)
                force="true"
                shift
                ;;
            --output)
                output_dir="$2"
                shift 2
                ;;
            --help|-h)
                show_help
                exit 0
                ;;
            *)
                log_error "未知选项: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    # 初始化
    init_framework
    load_strategy
    
    # 执行命令
    case $command in
        analyze)
            log_info "🔍 开始依赖分析..."
            discover_package_files
            local analysis_file
            analysis_file=$(extract_dependencies)
            local conflicts_file
            conflicts_file=$(identify_conflicts "$analysis_file")
            log_success "✓ 分析完成"
            ;;
        fix)
            log_info "🔧 开始修复依赖冲突..."
            discover_package_files
            local analysis_file
            analysis_file=$(extract_dependencies)
            local conflicts_file
            conflicts_file=$(identify_conflicts "$analysis_file")
            local resolution_file
            resolution_file=$(generate_resolution "$conflicts_file")
            apply_resolution "$resolution_file" "$dry_run"
            log_success "✓ 修复完成"
            ;;
        audit)
            log_info "🔒 开始安全审计..."
            discover_package_files
            local findings_file
            findings_file=$(security_audit)
            fix_vulnerabilities "$findings_file" "$auto_fix"
            log_success "✓ 审计完成"
            ;;
        report)
            log_info "📊 生成完整报告..."
            discover_package_files
            local analysis_file
            analysis_file=$(extract_dependencies)
            local conflicts_file
            conflicts_file=$(identify_conflicts "$analysis_file")
            local resolution_file
            resolution_file=$(generate_resolution "$conflicts_file")
            local findings_file
            findings_file=$(security_audit)
            generate_report "$analysis_file" "$conflicts_file" "$resolution_file" "$findings_file"
            log_success "✓ 报告生成完成"
            ;;
        sync)
            log_info "🔄 同步依赖版本..."
            # 组合执行分析和修复
            discover_package_files
            local analysis_file
            analysis_file=$(extract_dependencies)
            local conflicts_file
            conflicts_file=$(identify_conflicts "$analysis_file")
            local resolution_file
            resolution_file=$(generate_resolution "$conflicts_file")
            apply_resolution "$resolution_file" "false"
            log_success "✓ 同步完成"
            ;;
        clean)
            log_info "🧹 清理依赖..."
            while IFS= read -r package_file; do
                local project_dir
                project_dir=$(dirname "$package_file")
                local node_modules="$project_dir/node_modules"
                
                if [ -d "$node_modules" ]; then
                    log_info "清理: $node_modules"
                    rm -rf "$node_modules"
                fi
            done < <(find "$PROJECT_ROOT" -name "package.json" -not -path "*/node_modules/*")
            log_success "✓ 清理完成"
            ;;
        help)
            show_help
            ;;
        *)
            log_error "未知命令: $command"
            show_help
            exit 1
            ;;
    esac
}

# 如果直接执行此脚本
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    main "$@"
fi