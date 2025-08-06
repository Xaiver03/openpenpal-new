#!/bin/bash

# OpenPenPal 浏览器管理器 - SOTA级别的跨平台浏览器检测和打开机制
# 设计原则：
# 1. 跨平台兼容性 (macOS, Linux, Windows/WSL)
# 2. 多浏览器支持和优先级检测
# 3. 优雅的错误处理和降级机制
# 4. 配置驱动的URL管理
# 5. 健壮的进程检测和错误恢复

set -e

# 导入通用工具
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/utils.sh"

# 浏览器配置 - 使用兼容的方式
get_browser_commands() {
    local browser="$1"
    case "$browser" in
        "chrome") echo "google-chrome chromium-browser chrome" ;;
        "firefox") echo "firefox firefox-esr" ;;
        "safari") echo "safari" ;;
        "edge") echo "microsoft-edge edge" ;;
        "opera") echo "opera" ;;
        "brave") echo "brave-browser brave" ;;
        *) echo "" ;;
    esac
}

get_browser_priority() {
    local browser="$1"
    case "$browser" in
        "chrome") echo "100" ;;
        "firefox") echo "90" ;;
        "safari") echo "85" ;;
        "edge") echo "80" ;;
        "brave") echo "75" ;;
        "opera") echo "70" ;;
        *) echo "0" ;;
    esac
}

get_platform_open_command() {
    local platform="$1"
    case "$platform" in
        "darwin") echo "open" ;;
        "linux") echo "xdg-open" ;;
        "msys") echo "start" ;;
        "cygwin") echo "cygstart" ;;
        *) echo "" ;;
    esac
}

# 检测操作系统平台
detect_platform() {
    local uname_s="$(uname -s)"
    case "$uname_s" in
        Darwin*) echo "darwin" ;;
        Linux*) echo "linux" ;;
        CYGWIN*) echo "cygwin" ;;
        MINGW*|MSYS*) echo "msys" ;;
        *) echo "unknown" ;;
    esac
}

# 检测桌面环境（用于Linux）
detect_desktop_environment() {
    if [ -n "$XDG_CURRENT_DESKTOP" ]; then
        echo "$XDG_CURRENT_DESKTOP" | tr '[:upper:]' '[:lower:]'
    elif [ -n "$DESKTOP_SESSION" ]; then
        echo "$DESKTOP_SESSION" | tr '[:upper:]' '[:lower:]'
    elif command -v gnome-session >/dev/null 2>&1; then
        echo "gnome"
    elif command -v kde-open >/dev/null 2>&1; then
        echo "kde"
    elif command -v xfce4-session >/dev/null 2>&1; then
        echo "xfce"
    else
        echo "unknown"
    fi
}

# 检查显示服务器是否可用
check_display_server() {
    local platform="$1"
    
    case "$platform" in
        "darwin")
            # macOS总是有桌面环境
            return 0
            ;;
        "linux")
            # 检查X11或Wayland
            if [ -n "$DISPLAY" ] || [ -n "$WAYLAND_DISPLAY" ]; then
                return 0
            else
                log_debug "没有检测到图形显示服务器 (DISPLAY=$DISPLAY, WAYLAND_DISPLAY=$WAYLAND_DISPLAY)"
                return 1
            fi
            ;;
        *)
            return 0
            ;;
    esac
}

# 检测可用的浏览器
detect_available_browsers() {
    local platform="$1"
    local available_browsers=()
    
    # 定义所有支持的浏览器
    local all_browsers="chrome firefox safari edge brave opera"
    
    # 对于macOS，使用bundle identifier检测
    if [ "$platform" = "darwin" ]; then
        # 检测Safari
        if [ -d "/Applications/Safari.app" ]; then
            available_browsers+=("safari")
        fi
        
        # 检测Chrome
        if [ -d "/Applications/Google Chrome.app" ] || command -v google-chrome >/dev/null 2>&1; then
            available_browsers+=("chrome")
        fi
        
        # 检测Firefox
        if [ -d "/Applications/Firefox.app" ] || command -v firefox >/dev/null 2>&1; then
            available_browsers+=("firefox")
        fi
        
        # 检测Edge
        if [ -d "/Applications/Microsoft Edge.app" ] || command -v microsoft-edge >/dev/null 2>&1; then
            available_browsers+=("edge")
        fi
        
        # 检测Brave
        if [ -d "/Applications/Brave Browser.app" ] || command -v brave-browser >/dev/null 2>&1; then
            available_browsers+=("brave")
        fi
        
        # 检测Opera
        if [ -d "/Applications/Opera.app" ] || command -v opera >/dev/null 2>&1; then
            available_browsers+=("opera")
        fi
    else
        # 对于Linux和其他平台，使用命令检测
        for browser in $all_browsers; do
            local commands
            commands="$(get_browser_commands "$browser")"
            for cmd in $commands; do
                if command -v "$cmd" >/dev/null 2>&1; then
                    available_browsers+=("$browser")
                    break
                fi
            done
        done
    fi
    
    # 按优先级排序
    local sorted_browsers=()
    for priority in 100 90 85 80 75 70; do
        for browser in "${available_browsers[@]}"; do
            local browser_priority
            browser_priority="$(get_browser_priority "$browser")"
            if [ "$browser_priority" = "$priority" ]; then
                sorted_browsers+=("$browser")
            fi
        done
    done
    
    printf '%s\n' "${sorted_browsers[@]}"
}

# 获取浏览器的实际命令
get_browser_command() {
    local browser="$1"
    local platform="$2"
    
    case "$platform" in
        "darwin")
            case "$browser" in
                "safari") echo "open -a Safari" ;;
                "chrome") echo "open -a 'Google Chrome'" ;;
                "firefox") echo "open -a Firefox" ;;
                "edge") echo "open -a 'Microsoft Edge'" ;;
                "brave") echo "open -a 'Brave Browser'" ;;
                "opera") echo "open -a Opera" ;;
                *) echo "open" ;;
            esac
            ;;
        *)
            local commands
            commands="$(get_browser_commands "$browser")"
            for cmd in $commands; do
                if command -v "$cmd" >/dev/null 2>&1; then
                    echo "$cmd"
                    return 0
                fi
            done
            echo ""
            ;;
    esac
}

# 测试浏览器是否能够启动
test_browser_launch() {
    local browser_cmd="$1"
    local test_url="$2"
    local timeout="${3:-10}"
    
    log_debug "测试浏览器命令: $browser_cmd"
    
    # 使用超时控制和后台执行
    if command -v timeout >/dev/null 2>&1; then
        timeout "$timeout" $browser_cmd "$test_url" >/dev/null 2>&1 &
    else
        # macOS没有timeout命令，使用其他方法
        $browser_cmd "$test_url" >/dev/null 2>&1 &
    fi
    
    local browser_pid=$!
    
    # 等待短暂时间检查进程是否成功启动
    sleep 2
    
    if kill -0 "$browser_pid" 2>/dev/null; then
        log_debug "浏览器进程启动成功 (PID: $browser_pid)"
        return 0
    else
        log_debug "浏览器进程启动失败"
        return 1
    fi
}

# 尝试使用系统默认方式打开URL
try_system_open() {
    local url="$1"
    local platform="$2"
    
    local open_cmd
    open_cmd="$(get_platform_open_command "$platform")"
    if [ -n "$open_cmd" ] && command -v "$open_cmd" >/dev/null 2>&1; then
        log_debug "使用系统默认打开命令: $open_cmd"
        if $open_cmd "$url" >/dev/null 2>&1; then
            return 0
        fi
    fi
    
    return 1
}

# 打开URL的主函数
open_url() {
    local url="$1"
    local preferred_browser="${2:-auto}"
    local fallback_enabled="${3:-true}"
    
    if [ -z "$url" ]; then
        log_error "URL不能为空"
        return 1
    fi
    
    local platform
    platform="$(detect_platform)"
    
    log_debug "检测到平台: $platform"
    
    # 检查显示服务器
    if ! check_display_server "$platform"; then
        log_warning "未检测到图形显示服务器，无法打开浏览器"
        if [ "$fallback_enabled" = true ]; then
            log_info "请手动访问: $url"
        fi
        return 1
    fi
    
    # 获取可用浏览器列表
    local available_browsers=()
    while IFS= read -r browser; do
        [ -n "$browser" ] && available_browsers+=("$browser")
    done < <(detect_available_browsers "$platform")
    
    if [ ${#available_browsers[@]} -eq 0 ]; then
        log_warning "未检测到可用的浏览器"
        if [ "$fallback_enabled" = true ]; then
            log_info "请手动访问: $url"
        fi
        return 1
    fi
    
    log_debug "可用浏览器: ${available_browsers[*]}"
    
    # 确定要使用的浏览器
    local target_browsers=()
    if [ "$preferred_browser" = "auto" ]; then
        target_browsers=("${available_browsers[@]}")
    else
        # 检查指定的浏览器是否可用
        for browser in "${available_browsers[@]}"; do
            if [ "$browser" = "$preferred_browser" ]; then
                target_browsers=("$preferred_browser")
                break
            fi
        done
        
        if [ ${#target_browsers[@]} -eq 0 ]; then
            log_warning "指定的浏览器 '$preferred_browser' 不可用，使用自动检测"
            target_browsers=("${available_browsers[@]}")
        fi
    fi
    
    # 尝试打开浏览器
    for browser in "${target_browsers[@]}"; do
        local browser_cmd
        browser_cmd="$(get_browser_command "$browser" "$platform")"
        
        if [ -n "$browser_cmd" ]; then
            log_info "尝试使用 $browser 打开 $url"
            
            if test_browser_launch "$browser_cmd" "$url"; then
                log_success "✓ 成功使用 $browser 打开浏览器"
                return 0
            else
                log_warning "使用 $browser 打开浏览器失败"
            fi
        fi
    done
    
    # 尝试系统默认方式
    if [ "$fallback_enabled" = true ]; then
        log_info "尝试使用系统默认方式打开URL"
        if try_system_open "$url" "$platform"; then
            log_success "✓ 使用系统默认方式打开浏览器成功"
            return 0
        fi
    fi
    
    # 所有方法都失败
    log_error "无法打开浏览器"
    if [ "$fallback_enabled" = true ]; then
        log_info "请手动访问: $url"
    fi
    return 1
}

# 批量打开多个URL
open_multiple_urls() {
    local urls=("$@")
    local success_count=0
    local total_count=${#urls[@]}
    
    if [ $total_count -eq 0 ]; then
        log_error "没有提供URL"
        return 1
    fi
    
    log_info "准备打开 $total_count 个URL"
    
    for url in "${urls[@]}"; do
        if open_url "$url" "auto" true; then
            ((success_count++))
        fi
        # 在URL之间添加短暂延迟，避免浏览器启动冲突
        sleep 1
    done
    
    log_info "成功打开 $success_count/$total_count 个URL"
    
    if [ $success_count -eq $total_count ]; then
        return 0
    else
        return 1
    fi
}

# 显示浏览器环境信息
show_browser_info() {
    local platform
    platform="$(detect_platform)"
    
    echo "====== 浏览器环境信息 ======"
    echo "平台: $platform"
    
    if [ "$platform" = "linux" ]; then
        echo "桌面环境: $(detect_desktop_environment)"
        echo "显示服务器: DISPLAY=$DISPLAY, WAYLAND_DISPLAY=$WAYLAND_DISPLAY"
    fi
    
    echo "可用浏览器:"
    local available_browsers=()
    while IFS= read -r browser; do
        [ -n "$browser" ] && available_browsers+=("$browser")
    done < <(detect_available_browsers "$platform")
    
    if [ ${#available_browsers[@]} -eq 0 ]; then
        echo "  无"
    else
        for browser in "${available_browsers[@]}"; do
            local cmd priority
            cmd="$(get_browser_command "$browser" "$platform")"
            priority="$(get_browser_priority "$browser")"
            echo "  - $browser (优先级: $priority, 命令: $cmd)"
        done
    fi
    
    echo "=========================="
}

# 命令行接口
main() {
    case "${1:-}" in
        "open")
            shift
            if [ $# -eq 0 ]; then
                log_error "usage: $0 open <url> [browser] [fallback]"
                exit 1
            fi
            open_url "$@"
            ;;
        "open-multiple")
            shift
            open_multiple_urls "$@"
            ;;
        "info")
            show_browser_info
            ;;
        "test")
            shift
            local test_url="${1:-http://localhost:3000}"
            log_info "测试打开: $test_url"
            open_url "$test_url"
            ;;
        *)
            echo "usage: $0 {open|open-multiple|info|test} [args...]"
            echo ""
            echo "Commands:"
            echo "  open <url> [browser] [fallback]  - 打开指定URL"
            echo "  open-multiple <url1> <url2> ... - 批量打开多个URL"
            echo "  info                            - 显示浏览器环境信息"
            echo "  test [url]                      - 测试打开URL（默认localhost:3000）"
            exit 1
            ;;
    esac
}

# 如果直接执行此脚本
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    main "$@"
fi