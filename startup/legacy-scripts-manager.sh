#!/bin/bash

# OpenPenPal 遗留脚本管理器
# 管理和整合根目录中的启动脚本

set -e

# 获取脚本目录和项目根目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# 导入工具函数
source "$SCRIPT_DIR/utils.sh"

# 选项
ACTION=""
BACKUP=true

# 显示帮助信息
show_help() {
    cat << EOF
OpenPenPal 遗留脚本管理器

用法: $0 <action> [选项]

操作:
  migrate        迁移遗留脚本到 startup 目录
  backup         备份遗留脚本
  clean          清理根目录中的遗留脚本
  restore        恢复备份的脚本
  status         显示遗留脚本状态

选项:
  --no-backup    不创建备份
  --help, -h     显示此帮助信息

示例:
  $0 migrate           # 迁移所有遗留脚本
  $0 clean            # 清理根目录遗留脚本
  $0 status           # 查看状态

EOF
}

# 解析命令行参数
parse_arguments() {
    if [ $# -eq 0 ]; then
        show_help
        exit 1
    fi
    
    ACTION="$1"
    shift
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --no-backup)
                BACKUP=false
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

# 定义遗留文件列表
get_legacy_files() {
    local files=(
        "simple-start.js:legacy"
        "simple-gateway.js:legacy"
        "simple-mock-services.js:keep"
        "start-integration.sh:legacy"
        "stop-integration.sh:legacy"
        "test-permissions.sh:move"
        "verify-permissions.sh:move"
    )
    echo "${files[@]}"
}

# 创建备份
create_backup() {
    if [ "$BACKUP" = false ]; then
        return 0
    fi
    
    log_info "创建遗留脚本备份..."
    
    local backup_dir="$PROJECT_ROOT/backup/legacy-scripts-$(date +%Y%m%d_%H%M%S)"
    mkdir -p "$backup_dir"
    
    local files=($(get_legacy_files))
    
    for file_info in "${files[@]}"; do
        local file_name="${file_info%%:*}"
        local file_path="$PROJECT_ROOT/$file_name"
        
        if [ -f "$file_path" ]; then
            cp "$file_path" "$backup_dir/"
            log_debug "备份: $file_name"
        fi
    done
    
    log_success "备份创建完成: $backup_dir"
    echo "$backup_dir" > "$PROJECT_ROOT/.last_backup"
}

# 迁移遗留脚本
migrate_legacy_scripts() {
    log_info "迁移遗留脚本到 startup 目录..."
    
    create_backup
    
    # 移动测试脚本到 startup/tests/
    local test_dir="$SCRIPT_DIR/tests"
    mkdir -p "$test_dir"
    
    if [ -f "$PROJECT_ROOT/test-permissions.sh" ]; then
        mv "$PROJECT_ROOT/test-permissions.sh" "$test_dir/"
        log_success "移动: test-permissions.sh → startup/tests/"
    fi
    
    if [ -f "$PROJECT_ROOT/verify-permissions.sh" ]; then
        mv "$PROJECT_ROOT/verify-permissions.sh" "$test_dir/"
        log_success "移动: verify-permissions.sh → startup/tests/"
    fi
    
    # 创建遗留脚本的兼容性包装器
    create_compatibility_wrappers
    
    log_success "遗留脚本迁移完成"
}

# 创建兼容性包装器
create_compatibility_wrappers() {
    log_info "创建兼容性包装器..."
    
    # 为 simple-start.js 创建包装器
    if [ -f "$PROJECT_ROOT/simple-start.js" ]; then
        cat > "$PROJECT_ROOT/simple-start.sh" << 'EOF'
#!/bin/bash
# 兼容性包装器 - 重定向到新的启动系统
echo "⚠️ simple-start.js 已被替换为新的启动系统"
echo "正在启动简化Mock服务..."
exec "$(dirname "$0")/startup/start-simple-mock.sh" "$@"
EOF
        chmod +x "$PROJECT_ROOT/simple-start.sh"
        log_success "创建包装器: simple-start.sh"
    fi
    
    # 为 start-integration.sh 创建包装器
    if [ -f "$PROJECT_ROOT/start-integration.sh" ]; then
        cat > "$PROJECT_ROOT/start-integration-wrapper.sh" << 'EOF'
#!/bin/bash
# 兼容性包装器 - 重定向到新的启动系统
echo "⚠️ start-integration.sh 已被替换为新的启动系统"
echo "正在启动集成环境..."
exec "$(dirname "$0")/startup/start-integration.sh" "$@"
EOF
        chmod +x "$PROJECT_ROOT/start-integration-wrapper.sh"
        log_success "创建包装器: start-integration-wrapper.sh"
    fi
    
    # 为 stop-integration.sh 创建包装器
    if [ -f "$PROJECT_ROOT/stop-integration.sh" ]; then
        cat > "$PROJECT_ROOT/stop-integration-wrapper.sh" << 'EOF'
#!/bin/bash
# 兼容性包装器 - 重定向到新的启动系统
echo "⚠️ stop-integration.sh 已被替换为新的启动系统"
echo "正在停止所有服务..."
exec "$(dirname "$0")/startup/stop-all.sh" "$@"
EOF
        chmod +x "$PROJECT_ROOT/stop-integration-wrapper.sh"
        log_success "创建包装器: stop-integration-wrapper.sh"
    fi
}

# 清理遗留脚本
clean_legacy_scripts() {
    log_info "清理根目录中的遗留脚本..."
    
    create_backup
    
    local files_to_remove=(
        "simple-start.js"
        "simple-gateway.js"
        "start-integration.sh"
        "stop-integration.sh"
    )
    
    for file in "${files_to_remove[@]}"; do
        if [ -f "$PROJECT_ROOT/$file" ]; then
            rm "$PROJECT_ROOT/$file"
            log_success "删除: $file"
        fi
    done
    
    log_success "遗留脚本清理完成"
}

# 显示状态
show_status() {
    log_info "遗留脚本状态检查"
    log_info "=================="
    
    local files=($(get_legacy_files))
    local total_files=0
    local existing_files=0
    
    for file_info in "${files[@]}"; do
        local file_name="${file_info%%:*}"
        local action="${file_info##*:}"
        local file_path="$PROJECT_ROOT/$file_name"
        
        total_files=$((total_files + 1))
        
        if [ -f "$file_path" ]; then
            existing_files=$((existing_files + 1))
            local file_size=$(ls -lh "$file_path" | awk '{print $5}')
            
            case $action in
                legacy)
                    log_warning "⚠️  $file_name ($file_size) - 遗留文件，建议迁移"
                    ;;
                keep)
                    log_info "📄 $file_name ($file_size) - 保留使用"
                    ;;
                move)
                    log_info "📁 $file_name ($file_size) - 建议移动到 startup/tests/"
                    ;;
            esac
        else
            log_debug "❌ $file_name - 不存在"
        fi
    done
    
    log_info ""
    log_info "统计:"
    log_info "  总计: $total_files 个文件"
    log_info "  存在: $existing_files 个文件"
    log_info "  缺失: $((total_files - existing_files)) 个文件"
    
    # 检查 startup 目录状态
    log_info ""
    log_info "Startup 目录状态:"
    local startup_scripts=(
        "quick-start.sh"
        "stop-all.sh"
        "check-status.sh"
        "install-deps.sh"
        "start-simple-mock.sh"
        "start-integration.sh"
    )
    
    for script in "${startup_scripts[@]}"; do
        if [ -f "$SCRIPT_DIR/$script" ]; then
            log_success "✅ $script"
        else
            log_error "❌ $script"
        fi
    done
}

# 恢复备份
restore_backup() {
    if [ ! -f "$PROJECT_ROOT/.last_backup" ]; then
        log_error "没有找到备份信息"
        exit 1
    fi
    
    local backup_dir=$(cat "$PROJECT_ROOT/.last_backup")
    
    if [ ! -d "$backup_dir" ]; then
        log_error "备份目录不存在: $backup_dir"
        exit 1
    fi
    
    log_info "恢复备份: $backup_dir"
    
    for file in "$backup_dir"/*; do
        if [ -f "$file" ]; then
            local filename=$(basename "$file")
            cp "$file" "$PROJECT_ROOT/"
            log_success "恢复: $filename"
        fi
    done
    
    log_success "备份恢复完成"
}

# 创建启动脚本索引
create_startup_index() {
    log_info "创建启动脚本索引..."
    
    local index_file="$PROJECT_ROOT/STARTUP_SCRIPTS.md"
    
    cat > "$index_file" << 'EOF'
# OpenPenPal 启动脚本索引

本文档列出了所有可用的启动脚本和工具。

## 主要启动脚本

### 新统一启动系统 (推荐)

| 脚本 | 描述 | 用法 |
|------|------|------|
| `启动 OpenPenPal 集成.command` | macOS主启动器，提供多种选择 | 双击运行或 `./启动\ OpenPenPal\ 集成.command` |
| `startup/openpenpal-launcher.command` | 图形化启动菜单 | 双击运行或 `./startup/openpenpal-launcher.command` |
| `startup/quick-start.sh` | 一键启动脚本 | `./startup/quick-start.sh [模式]` |

### 启动模式

- **development**: 开发模式，完整微服务环境
- **production**: 生产模式，包含管理后台
- **simple**: 简化模式，最小服务集
- **demo**: 演示模式，自动打开浏览器

### 管理工具

| 脚本 | 描述 | 用法 |
|------|------|------|
| `startup/stop-all.sh` | 停止所有服务 | `./startup/stop-all.sh [--force]` |
| `startup/check-status.sh` | 检查服务状态 | `./startup/check-status.sh [--detailed]` |
| `startup/install-deps.sh` | 安装项目依赖 | `./startup/install-deps.sh [--force]` |

### 专用启动脚本

| 脚本 | 描述 | 用法 |
|------|------|------|
| `startup/start-simple-mock.sh` | 简化Mock服务 | `./startup/start-simple-mock.sh` |
| `startup/start-integration.sh` | 传统集成模式 | `./startup/start-integration.sh` |

## 快速开始

### 首次运行
```bash
# 1. 安装依赖
./startup/install-deps.sh

# 2. 启动演示模式
./startup/quick-start.sh demo --auto-open
```

### 开发环境
```bash
# 启动开发模式
./startup/quick-start.sh development --auto-open
```

### 检查状态
```bash
# 查看服务状态
./startup/check-status.sh

# 持续监控
./startup/check-status.sh --continuous
```

### 停止服务
```bash
# 正常停止
./startup/stop-all.sh

# 强制停止
./startup/stop-all.sh --force
```

## 配置文件

- `startup/startup-config.json`: 服务配置
- `startup/environment-vars.sh`: 环境变量
- `startup/utils.sh`: 工具函数

## 日志文件

所有日志文件位于 `logs/` 目录：
- `logs/frontend.log`: 前端日志
- `logs/simple-mock.log`: 简化Mock服务日志
- `logs/*.pid`: 进程ID文件

## 兼容性说明

原有的启动脚本已被新系统替代：
- `simple-start.js` → `startup/start-simple-mock.sh`
- `start-integration.sh` → `startup/start-integration.sh`
- `stop-integration.sh` → `startup/stop-all.sh`

如需使用原脚本，仍可通过兼容性包装器访问。

## 故障排查

1. **端口被占用**: 使用 `./startup/stop-all.sh --force` 清理
2. **依赖问题**: 运行 `./startup/install-deps.sh --force --cleanup`
3. **服务启动失败**: 查看 `logs/*.log` 文件
4. **权限问题**: 运行 `chmod +x startup/*.sh`

## 技术支持

- 查看日志: `tail -f logs/*.log`
- 检查进程: `./startup/check-status.sh --detailed`
- 完整重启: `./startup/stop-all.sh && ./startup/quick-start.sh`
EOF

    log_success "启动脚本索引创建完成: $index_file"
}

# 主函数
main() {
    # 解析参数
    parse_arguments "$@"
    
    # 进入项目根目录
    cd "$PROJECT_ROOT"
    
    case $ACTION in
        migrate)
            migrate_legacy_scripts
            create_startup_index
            ;;
        backup)
            create_backup
            ;;
        clean)
            clean_legacy_scripts
            ;;
        restore)
            restore_backup
            ;;
        status)
            show_status
            ;;
        *)
            log_error "未知操作: $ACTION"
            show_help
            exit 1
            ;;
    esac
    
    log_success "操作完成: $ACTION"
}

# 执行主函数
main "$@"