#!/bin/bash

# OpenPenPal 依赖安装脚本
# 自动检查和安装所有项目依赖

set -e

# 获取脚本目录和项目根目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# 导入工具函数和环境变量
source "$SCRIPT_DIR/utils.sh"
source "$SCRIPT_DIR/environment-vars.sh"

# 默认选项
FORCE=false
SKIP_FRONTEND=false
SKIP_MOCK=false
SKIP_ADMIN=false
CLEANUP=false

# 显示帮助信息
show_help() {
    cat << EOF
OpenPenPal 依赖安装脚本

用法: $0 [选项]

选项:
  --force            强制重新安装所有依赖
  --skip-frontend    跳过前端依赖安装
  --skip-mock        跳过Mock服务依赖安装
  --skip-admin       跳过管理后台依赖安装
  --cleanup          安装前清理 node_modules
  --help, -h         显示此帮助信息

示例:
  $0                    # 检查并安装缺失的依赖
  $0 --force            # 强制重新安装所有依赖
  $0 --cleanup --force  # 清理后重新安装
  $0 --skip-admin       # 跳过管理后台依赖

EOF
}

# 解析命令行参数
parse_arguments() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --force)
                FORCE=true
                shift
                ;;
            --skip-frontend)
                SKIP_FRONTEND=true
                shift
                ;;
            --skip-mock)
                SKIP_MOCK=true
                shift
                ;;
            --skip-admin)
                SKIP_ADMIN=true
                shift
                ;;
            --cleanup)
                CLEANUP=true
                shift
                ;;
            --help|-h)
                show_help
                exit 0
                ;;
            *)
                log_error "未知参数: $1"
                show_help
                exit 1
                ;;
        esac
    done
}

# 检查系统要求
check_system_requirements() {
    log_info "检查系统要求..."
    
    # 检查 Node.js
    if ! command_exists node; then
        log_error "Node.js 未安装"
        log_info "请访问 https://nodejs.org/ 下载并安装 Node.js 18+ 版本"
        exit 1
    fi
    
    local node_version=$(node --version | sed 's/v//')
    if version_lt "$node_version" "18.0.0"; then
        log_error "Node.js 版本过低: $node_version (需要 18.0.0+)"
        log_info "请升级 Node.js 到 18+ 版本"
        exit 1
    fi
    
    # 检查 npm
    if ! command_exists npm; then
        log_error "npm 未安装"
        exit 1
    fi
    
    local npm_version=$(npm --version)
    if version_lt "$npm_version" "8.0.0"; then
        log_warning "npm 版本较低: $npm_version (建议 8.0.0+)"
    fi
    
    log_success "✓ Node.js v$node_version"
    log_success "✓ npm v$npm_version"
    
    # 检查可用内存
    local memory=$(get_memory_usage)
    log_info "可用内存: $memory"
    
    # 检查磁盘空间
    if ! check_disk_space "$PROJECT_ROOT" 2; then
        log_warning "磁盘空间可能不足，建议至少保留 2GB 空间"
    fi
}

# 清理依赖
cleanup_dependencies() {
    if [ "$CLEANUP" = false ]; then
        return 0
    fi
    
    log_info "清理现有依赖..."
    
    local dirs_to_clean=(
        "$PROJECT_ROOT/frontend/node_modules"
        "$PROJECT_ROOT/apps/mock-services/node_modules"
        "$PROJECT_ROOT/services/admin-service/frontend/node_modules"
    )
    
    for dir in "${dirs_to_clean[@]}"; do
        if [ -d "$dir" ]; then
            log_debug "清理目录: $dir"
            rm -rf "$dir"
        fi
    done
    
    # 清理 package-lock.json 文件
    local lockfiles=(
        "$PROJECT_ROOT/frontend/package-lock.json"
        "$PROJECT_ROOT/apps/mock-services/package-lock.json"
        "$PROJECT_ROOT/services/admin-service/frontend/package-lock.json"
    )
    
    for lockfile in "${lockfiles[@]}"; do
        if [ -f "$lockfile" ]; then
            log_debug "清理锁文件: $lockfile"
            rm -f "$lockfile"
        fi
    done
    
    log_success "依赖清理完成"
}

# 安装单个项目的依赖
install_project_deps() {
    local project_name="$1"
    local project_path="$2"
    local skip_flag="$3"
    
    if [ "$skip_flag" = true ]; then
        log_info "跳过 $project_name 依赖安装"
        return 0
    fi
    
    if [ ! -d "$project_path" ]; then
        log_warning "$project_name 目录不存在: $project_path"
        return 0
    fi
    
    if [ ! -f "$project_path/package.json" ]; then
        log_warning "$project_name 没有 package.json: $project_path"
        return 0
    fi
    
    log_info "安装 $project_name 依赖..."
    
    # 检查是否需要安装
    local needs_install=false
    
    if [ "$FORCE" = true ]; then
        needs_install=true
        log_debug "强制安装模式"
    elif [ ! -d "$project_path/node_modules" ]; then
        needs_install=true
        log_debug "node_modules 不存在"
    elif [ "$project_path/package.json" -nt "$project_path/node_modules" ]; then
        needs_install=true
        log_debug "package.json 更新"
    elif [ -f "$project_path/package-lock.json" ] && [ "$project_path/package-lock.json" -nt "$project_path/node_modules" ]; then
        needs_install=true
        log_debug "package-lock.json 更新"
    fi
    
    if [ "$needs_install" = false ]; then
        log_success "$project_name 依赖已是最新"
        return 0
    fi
    
    # 进入项目目录
    cd "$project_path"
    
    # 清理缓存（如果需要）
    if [ "$FORCE" = true ] || [ "$CLEANUP" = true ]; then
        log_debug "清理 npm 缓存"
        npm cache clean --force >/dev/null 2>&1 || true
    fi
    
    # 安装依赖
    log_step "正在安装 $project_name 依赖..."
    
    # 使用更详细的日志
    if [ "$VERBOSE" = true ]; then
        npm install
    else
        if ! npm install --silent; then
            log_error "$project_name 依赖安装失败"
            log_info "尝试查看错误日志："
            if [ -f "npm-debug.log" ]; then
                tail -n 20 npm-debug.log
            fi
            return 1
        fi
    fi
    
    # 验证安装
    if [ ! -d "node_modules" ]; then
        log_error "$project_name 依赖安装失败 - node_modules 不存在"
        return 1
    fi
    
    # 统计依赖数量
    local dep_count=$(find node_modules -maxdepth 1 -type d | wc -l)
    dep_count=$((dep_count - 1))  # 排除 node_modules 本身
    
    log_success "$project_name 依赖安装完成 ($dep_count 个包)"
    
    # 返回项目根目录
    cd "$PROJECT_ROOT"
    
    return 0
}

# 检查依赖漏洞
check_security() {
    log_info "检查安全漏洞..."
    
    local projects=(
        "frontend:$PROJECT_ROOT/frontend"
        "mock-services:$PROJECT_ROOT/apps/mock-services"
    )
    
    if [ -d "$PROJECT_ROOT/services/admin-service/frontend" ]; then
        projects+=("admin-frontend:$PROJECT_ROOT/services/admin-service/frontend")
    fi
    
    local total_vulnerabilities=0
    
    for project_info in "${projects[@]}"; do
        local project_name=$(echo "$project_info" | cut -d':' -f1)
        local project_path=$(echo "$project_info" | cut -d':' -f2)
        
        if [ ! -d "$project_path/node_modules" ]; then
            continue
        fi
        
        cd "$project_path"
        
        log_debug "检查 $project_name 安全漏洞..."
        
        # 运行安全审计
        local audit_output=$(npm audit --audit-level=moderate --json 2>/dev/null || echo '{"vulnerabilities":{}}')
        local vuln_count=$(echo "$audit_output" | jq -r '.metadata.vulnerabilities.total // 0' 2>/dev/null || echo "0")
        
        if [ "$vuln_count" -gt 0 ]; then
            log_warning "$project_name 发现 $vuln_count 个安全漏洞"
            total_vulnerabilities=$((total_vulnerabilities + vuln_count))
            
            if [ "$VERBOSE" = true ]; then
                npm audit --audit-level=moderate
            fi
        else
            log_success "$project_name 安全检查通过"
        fi
        
        cd "$PROJECT_ROOT"
    done
    
    if [ $total_vulnerabilities -gt 0 ]; then
        log_warning "总计发现 $total_vulnerabilities 个安全漏洞"
        log_info "运行 'npm audit fix' 尝试自动修复"
    else
        log_success "所有项目安全检查通过"
    fi
}

# 验证安装结果
verify_installation() {
    log_info "验证安装结果..."
    
    local projects=(
        "前端:$PROJECT_ROOT/frontend:$SKIP_FRONTEND"
        "Mock服务:$PROJECT_ROOT/apps/mock-services:$SKIP_MOCK"
    )
    
    if [ -d "$PROJECT_ROOT/services/admin-service/frontend" ]; then
        projects+=("管理后台:$PROJECT_ROOT/services/admin-service/frontend:$SKIP_ADMIN")
    fi
    
    local all_success=true
    
    for project_info in "${projects[@]}"; do
        local project_name=$(echo "$project_info" | cut -d':' -f1)
        local project_path=$(echo "$project_info" | cut -d':' -f2)
        local skip_flag=$(echo "$project_info" | cut -d':' -f3)
        
        if [ "$skip_flag" = true ]; then
            continue
        fi
        
        if [ ! -d "$project_path/node_modules" ]; then
            log_error "$project_name 依赖安装失败 - node_modules 不存在"
            all_success=false
            continue
        fi
        
        # 检查关键依赖
        cd "$project_path"
        
        if [ "$project_name" = "前端" ] || [ "$project_name" = "管理后台" ]; then
            # 检查 Vue.js 相关依赖
            if [ ! -d "node_modules/vue" ]; then
                log_error "$project_name Vue.js 依赖缺失"
                all_success=false
            fi
        elif [ "$project_name" = "Mock服务" ]; then
            # 检查 Express.js 相关依赖
            if [ ! -d "node_modules/express" ]; then
                log_error "$project_name Express.js 依赖缺失"
                all_success=false
            fi
        fi
        
        if [ "$all_success" = true ]; then
            log_success "$project_name 依赖验证通过"
        fi
        
        cd "$PROJECT_ROOT"
    done
    
    return $([ "$all_success" = true ] && echo 0 || echo 1)
}

# 生成依赖报告
generate_report() {
    log_info "生成依赖报告..."
    
    local report_file="$PROJECT_ROOT/dependency-report.txt"
    
    {
        echo "OpenPenPal 依赖安装报告"
        echo "======================="
        echo "生成时间: $(date)"
        echo "Node.js: $(node --version)"
        echo "npm: $(npm --version)"
        echo ""
        
        echo "项目依赖状态:"
        echo "============"
        
        local projects=(
            "frontend:前端应用"
            "apps/mock-services:Mock服务"
        )
        
        if [ -d "$PROJECT_ROOT/services/admin-service/frontend" ]; then
            projects+=("services/admin-service/frontend:管理后台")
        fi
        
        for project_info in "${projects[@]}"; do
            local project_path=$(echo "$project_info" | cut -d':' -f1)
            local project_name=$(echo "$project_info" | cut -d':' -f2)
            
            echo ""
            echo "$project_name ($project_path):"
            
            if [ -f "$PROJECT_ROOT/$project_path/package.json" ]; then
                local package_name=$(jq -r '.name // "unknown"' "$PROJECT_ROOT/$project_path/package.json" 2>/dev/null || echo "unknown")
                local package_version=$(jq -r '.version // "unknown"' "$PROJECT_ROOT/$project_path/package.json" 2>/dev/null || echo "unknown")
                echo "  包名: $package_name"
                echo "  版本: $package_version"
            fi
            
            if [ -d "$PROJECT_ROOT/$project_path/node_modules" ]; then
                local dep_count=$(find "$PROJECT_ROOT/$project_path/node_modules" -maxdepth 1 -type d | wc -l)
                dep_count=$((dep_count - 1))
                echo "  依赖数量: $dep_count"
                echo "  状态: ✓ 已安装"
            else
                echo "  状态: ✗ 未安装"
            fi
        done
        
    } > "$report_file"
    
    log_success "依赖报告已生成: $report_file"
}

# 主函数
main() {
    # 解析参数
    parse_arguments "$@"
    
    # 显示启动信息
    log_info "📦 OpenPenPal 依赖安装器"
    log_info "========================"
    
    # 进入项目根目录
    cd "$PROJECT_ROOT"
    
    # 检查系统要求
    check_system_requirements
    
    # 清理依赖
    cleanup_dependencies
    
    # 安装各项目依赖
    log_info ""
    log_info "开始安装项目依赖..."
    
    local install_failed=false
    
    # 安装前端依赖
    if ! install_project_deps "前端应用" "$PROJECT_ROOT/frontend" "$SKIP_FRONTEND"; then
        install_failed=true
    fi
    
    # 安装Mock服务依赖
    if ! install_project_deps "Mock服务" "$PROJECT_ROOT/apps/mock-services" "$SKIP_MOCK"; then
        install_failed=true
    fi
    
    # 安装管理后台依赖（如果存在）
    if [ -d "$PROJECT_ROOT/services/admin-service/frontend" ]; then
        if ! install_project_deps "管理后台" "$PROJECT_ROOT/services/admin-service/frontend" "$SKIP_ADMIN"; then
            install_failed=true
        fi
    fi
    
    if [ "$install_failed" = true ]; then
        log_error "部分依赖安装失败"
        exit 1
    fi
    
    # 验证安装
    if ! verify_installation; then
        log_error "依赖验证失败"
        exit 1
    fi
    
    # 安全检查
    if command_exists jq; then
        check_security
    else
        log_warning "jq 未安装，跳过安全检查"
    fi
    
    # 生成报告
    if command_exists jq; then
        generate_report
    fi
    
    # 显示结果
    log_info ""
    log_success "🎉 所有依赖安装完成！"
    log_info ""
    log_info "💡 下一步："
    log_info "  • 启动服务: ./startup/quick-start.sh"
    log_info "  • 检查状态: ./startup/check-status.sh"
    log_info "  • 查看指南: ./startup/README.md"
    log_info ""
}

# 执行主函数
main "$@"