#!/bin/bash

# OpenPenPal URL管理器 - SOTA级别的配置驱动URL管理系统
# 设计原则：
# 1. 配置驱动的URL管理
# 2. 模式特定的行为配置
# 3. 健康状态检查和URL验证
# 4. 智能的浏览器打开策略
# 5. 用户友好的URL展示

set -e

# 导入依赖
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/utils.sh"
source "$SCRIPT_DIR/browser-manager.sh"

# 配置文件路径
URL_CONFIG_FILE="$SCRIPT_DIR/url-config.json"

# 检查配置文件是否存在
check_config_file() {
    if [ ! -f "$URL_CONFIG_FILE" ]; then
        log_error "URL配置文件不存在: $URL_CONFIG_FILE"
        return 1
    fi
    
    if ! command -v jq >/dev/null 2>&1; then
        log_error "jq 未安装，无法解析JSON配置"
        return 1
    fi
    
    # 验证JSON格式
    if ! jq . "$URL_CONFIG_FILE" >/dev/null 2>&1; then
        log_error "URL配置文件格式无效"
        return 1
    fi
    
    return 0
}

# 获取URL配置
get_url_config() {
    local url_key="$1"
    local field="${2:-url}"
    
    if ! check_config_file; then
        return 1
    fi
    
    jq -r ".urls[\"$url_key\"].$field // empty" "$URL_CONFIG_FILE"
}

# 获取模式配置
get_mode_config() {
    local mode="$1"
    local field="$2"
    
    if ! check_config_file; then
        return 1
    fi
    
    local result
    result=$(jq -r ".modes[\"$mode\"].$field" "$URL_CONFIG_FILE" 2>/dev/null)
    if [ "$result" = "null" ] || [ -z "$result" ]; then
        echo ""
    else
        echo "$result"
    fi
}

# 获取浏览器配置
get_browser_config() {
    local field="$1"
    
    if ! check_config_file; then
        return 1
    fi
    
    jq -r ".browser.$field // empty" "$URL_CONFIG_FILE"
}

# 获取显示配置
get_display_config() {
    local field="$1"
    
    if ! check_config_file; then
        return 1
    fi
    
    jq -r ".display.$field // empty" "$URL_CONFIG_FILE"
}

# 获取所有URL键
get_all_url_keys() {
    if ! check_config_file; then
        return 1
    fi
    
    jq -r '.urls | keys[]' "$URL_CONFIG_FILE"
}

# 检查URL健康状态
check_url_health() {
    local url="$1"
    local health_path="$2"
    local timeout="${3:-10}"
    
    local health_url="$url$health_path"
    
    log_debug "检查健康状态: $health_url"
    
    # 使用curl检查健康状态，禁用代理对本地服务的影响
    if command -v curl >/dev/null 2>&1; then
        if curl -s --max-time "$timeout" --fail --noproxy "*" "$health_url" >/dev/null 2>&1; then
            return 0
        fi
    fi
    
    # 备用方案：检查端口是否开启
    local port
    port=$(echo "$url" | sed -n 's|.*:\([0-9]*\).*|\1|p')
    if [ -n "$port" ]; then
        if check_port_occupied "$port"; then
            return 0
        fi
    fi
    
    return 1
}

# 等待URL变为可用
wait_for_url() {
    local url_key="$1"
    local timeout="${2:-30}"
    
    local url health_path wait_timeout
    url="$(get_url_config "$url_key" "url")"
    health_path="$(get_url_config "$url_key" "health_check")"
    wait_timeout="$(get_url_config "$url_key" "wait_timeout")"
    
    if [ -z "$url" ]; then
        log_error "URL配置不存在: $url_key"
        return 1
    fi
    
    # 使用配置的超时时间或传入的超时时间
    local actual_timeout="${wait_timeout:-$timeout}"
    local elapsed=0
    local check_interval=2
    
    log_info "等待 $url_key 服务启动 (超时: ${actual_timeout}s)"
    
    while [ $elapsed -lt "$actual_timeout" ]; do
        if check_url_health "$url" "$health_path" 5; then
            log_success "✓ $url_key 服务已就绪"
            return 0
        fi
        
        sleep $check_interval
        elapsed=$((elapsed + check_interval))
        
        if [ $((elapsed % 10)) -eq 0 ]; then
            log_debug "等待 $url_key... (${elapsed}/${actual_timeout}s)"
        fi
    done
    
    log_warning "$url_key 服务启动超时"
    return 1
}

# 显示单个URL信息
show_url_info() {
    local url_key="$1"
    local show_health="${2:-true}"
    
    local name url description health_path
    name="$(get_url_config "$url_key" "name")"
    url="$(get_url_config "$url_key" "url")"
    description="$(get_url_config "$url_key" "description")"
    health_path="$(get_url_config "$url_key" "health_check")"
    
    if [ -z "$url" ]; then
        log_error "URL配置不存在: $url_key"
        return 1
    fi
    
    local status_indicator=""
    if [ "$show_health" = "true" ]; then
        if check_url_health "$url" "$health_path" 3; then
            status_indicator="${GREEN}✓${NC}"
        else
            status_indicator="${RED}✗${NC}"
        fi
    fi
    
    printf "  %s ${CYAN}%s${NC} - %s\n" \
        "$status_indicator" \
        "$name" \
        "$url"
    
    if [ -n "$description" ] && [ "$(get_display_config "compact_mode")" != "true" ]; then
        printf "    %s\n" "$description"
    fi
}

# 展示所有URL信息
show_all_urls() {
    local mode="${1:-development}"
    local show_health="${2:-true}"
    
    log_info "🌐 可用服务"
    
    # 获取显示配置
    local show_all_urls
    show_all_urls="$(get_mode_config "$mode" "show_all_urls")"
    
    if [ "$show_all_urls" = "false" ]; then
        # 只显示会自动打开的URL
        local auto_open_urls
        auto_open_urls="$(get_mode_config "$mode" "auto_open_urls")"
        
        if [ -n "$auto_open_urls" ] && [ "$auto_open_urls" != "null" ]; then
            echo "$auto_open_urls" | jq -r '.[]' | while read -r url_key; do
                show_url_info "$url_key" "$show_health"
            done
        fi
    else
        # 显示所有URL
        while read -r url_key; do
            show_url_info "$url_key" "$show_health"
        done < <(get_all_url_keys)
    fi
    
    # 显示网络信息
    if [ "$(get_display_config "show_network_info")" = "true" ]; then
        echo ""
        log_info "📡 网络信息"
        if command -v ifconfig >/dev/null 2>&1; then
            local local_ip
            local_ip=$(ifconfig | grep -E 'inet (10\.|172\.(1[6-9]|2[0-9]|3[01])\.|192\.168\.)' | head -1 | awk '{print $2}')
            if [ -n "$local_ip" ]; then
                printf "  本地IP: %s\n" "$local_ip"
            fi
        fi
    fi
}

# 打开配置的URL
open_configured_urls() {
    local mode="${1:-development}"
    
    if ! check_config_file; then
        return 1
    fi
    
    # 获取要自动打开的URL列表
    local auto_open_urls
    auto_open_urls="$(get_mode_config "$mode" "auto_open_urls")"
    
    if [ -z "$auto_open_urls" ] || [ "$auto_open_urls" = "null" ] || [ "$auto_open_urls" = "[]" ]; then
        log_info "当前模式无需自动打开浏览器"
        return 0
    fi
    
    # 获取浏览器配置
    local preferred_browser fallback_enabled retry_count retry_delay browser_delay
    preferred_browser="$(get_browser_config "preferred")"
    fallback_enabled="$(get_browser_config "fallback_enabled")"
    retry_count="$(get_browser_config "retry_count")"
    retry_delay="$(get_browser_config "retry_delay")"
    browser_delay="$(get_mode_config "$mode" "browser_delay")"
    
    # 设置默认值
    preferred_browser="${preferred_browser:-auto}"
    fallback_enabled="${fallback_enabled:-true}"
    retry_count="${retry_count:-3}"
    retry_delay="${retry_delay:-2}"
    browser_delay="${browser_delay:-3}"
    
    log_info "🌐 准备打开浏览器 (延迟${browser_delay}秒)"
    sleep "$browser_delay"
    
    local success_count=0
    local total_count=0
    
    # 解析JSON数组并逐个处理
    echo "$auto_open_urls" | jq -r '.[]' | while read -r url_key; do
        local name url
        name="$(get_url_config "$url_key" "name")"
        url="$(get_url_config "$url_key" "url")"
        
        if [ -z "$url" ]; then
            log_warning "跳过无效的URL配置: $url_key"
            continue
        fi
        
        total_count=$((total_count + 1))
        
        log_info "打开 $name: $url"
        
        # 重试机制
        local attempt=1
        local opened=false
        
        while [ $attempt -le "$retry_count" ] && [ "$opened" = "false" ]; do
            if [ $attempt -gt 1 ]; then
                log_info "重试 $attempt/$retry_count: $name"
                sleep "$retry_delay"
            fi
            
            if open_url "$url" "$preferred_browser" "$fallback_enabled"; then
                opened=true
                success_count=$((success_count + 1))
                log_success "✓ 成功打开 $name"
            else
                log_warning "尝试 $attempt 失败: $name"
            fi
            
            attempt=$((attempt + 1))
        done
        
        if [ "$opened" = "false" ]; then
            log_error "无法打开 $name ($url)"
        fi
        
        # URL之间的延迟
        sleep 1
    done
    
    # 显示总结（这里有个问题，子shell中的变量不会传到外面）
    # 所以我们用不同的方法来统计
    local total_urls
    total_urls=$(echo "$auto_open_urls" | jq -r '. | length')
    log_info "浏览器打开完成，共 $total_urls 个链接"
}

# 验证URLs是否可访问
validate_all_urls() {
    local mode="${1:-development}"
    local timeout="${2:-5}"
    
    log_info "🔍 验证服务状态"
    
    local all_healthy=true
    
    while read -r url_key; do
        local name url health_path
        name="$(get_url_config "$url_key" "name")"
        url="$(get_url_config "$url_key" "url")"
        health_path="$(get_url_config "$url_key" "health_check")"
        
        if [ -z "$url" ]; then
            continue
        fi
        
        printf "  检查 %-15s " "$name"
        
        if check_url_health "$url" "$health_path" "$timeout"; then
            printf "${GREEN}✓ 正常${NC}\n"
        else
            printf "${RED}✗ 不可用${NC}\n"
            all_healthy=false
        fi
    done < <(get_all_url_keys)
    
    if [ "$all_healthy" = "true" ]; then
        log_success "所有服务状态正常"
        return 0
    else
        log_warning "部分服务不可用"
        return 1
    fi
}

# 显示配置信息
show_config_info() {
    local mode="${1:-development}"
    
    echo "====== URL配置信息 ======"
    echo "配置文件: $URL_CONFIG_FILE"
    echo "当前模式: $mode"
    echo ""
    
    # 显示模式配置
    echo "模式配置:"
    local auto_open_urls show_all_urls keep_running browser_delay
    auto_open_urls="$(get_mode_config "$mode" "auto_open_urls")"
    show_all_urls="$(get_mode_config "$mode" "show_all_urls")"
    keep_running="$(get_mode_config "$mode" "keep_running")"
    browser_delay="$(get_mode_config "$mode" "browser_delay")"
    
    echo "  自动打开URL: $auto_open_urls"
    echo "  显示所有URL: $show_all_urls"
    echo "  保持运行: ${keep_running:-未设置}"
    echo "  浏览器延迟: ${browser_delay}秒"
    echo ""
    
    # 显示浏览器配置
    echo "浏览器配置:"
    local preferred fallback_enabled retry_count retry_delay
    preferred="$(get_browser_config "preferred")"
    fallback_enabled="$(get_browser_config "fallback_enabled")"
    retry_count="$(get_browser_config "retry_count")"
    retry_delay="$(get_browser_config "retry_delay")"
    
    echo "  首选浏览器: $preferred"
    echo "  启用回退: $fallback_enabled"
    echo "  重试次数: $retry_count"
    echo "  重试延迟: ${retry_delay}秒"
    echo ""
    
    # 显示URL列表
    echo "配置的URL:"
    while read -r url_key; do
        local name url priority auto_open
        name="$(get_url_config "$url_key" "name")"
        url="$(get_url_config "$url_key" "url")"
        priority="$(get_url_config "$url_key" "priority")"
        auto_open="$(get_url_config "$url_key" "auto_open")"
        
        echo "  - $url_key: $name ($url) [优先级:$priority, 自动打开:$auto_open]"
    done < <(get_all_url_keys)
    
    echo "=========================="
}

# 命令行接口
main() {
    case "${1:-}" in
        "open")
            shift
            local mode="${1:-development}"
            open_configured_urls "$mode"
            ;;
        "show")
            shift
            local mode="${1:-development}"
            local show_health="${2:-true}"
            show_all_urls "$mode" "$show_health"
            ;;
        "validate")
            shift
            local mode="${1:-development}"
            local timeout="${2:-5}"
            validate_all_urls "$mode" "$timeout"
            ;;
        "wait")
            shift
            local url_key="$1"
            local timeout="${2:-30}"
            if [ -z "$url_key" ]; then
                log_error "usage: $0 wait <url_key> [timeout]"
                exit 1
            fi
            wait_for_url "$url_key" "$timeout"
            ;;
        "config")
            shift
            local mode="${1:-development}"
            show_config_info "$mode"
            ;;
        "health")
            shift
            local url_key="$1"
            if [ -z "$url_key" ]; then
                log_error "usage: $0 health <url_key>"
                exit 1
            fi
            
            local url health_path
            url="$(get_url_config "$url_key" "url")"
            health_path="$(get_url_config "$url_key" "health_check")"
            
            if check_url_health "$url" "$health_path" 10; then
                log_success "$url_key 健康状态正常"
                exit 0
            else
                log_error "$url_key 健康状态异常"
                exit 1
            fi
            ;;
        *)
            echo "usage: $0 {open|show|validate|wait|config|health} [args...]"
            echo ""
            echo "Commands:"
            echo "  open <mode>               - 打开配置的URLs"
            echo "  show <mode> [show_health] - 显示所有URLs"
            echo "  validate <mode> [timeout] - 验证所有URLs状态"
            echo "  wait <url_key> [timeout]  - 等待指定URL可用"
            echo "  config <mode>             - 显示配置信息"
            echo "  health <url_key>          - 检查指定URL健康状态"
            exit 1
            ;;
    esac
}

# 如果直接执行此脚本
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    main "$@"
fi