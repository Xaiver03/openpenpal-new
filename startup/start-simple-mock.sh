#!/bin/bash

# OpenPenPal 简化Mock服务启动脚本
# 基于原 simple-start.js 和 simple-gateway.js 的功能

set -e

# 获取脚本目录和项目根目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# 导入工具函数和环境变量
source "$SCRIPT_DIR/utils.sh"
source "$SCRIPT_DIR/environment-vars.sh"

# 默认选项
AUTO_OPEN=false
SIMPLE_MODE=false

# 显示帮助信息
show_help() {
    cat << EOF
OpenPenPal 简化Mock服务启动脚本

用法: $0 [选项]

选项:
  --auto-open    自动打开浏览器
  --simple       使用简化模式（仅基础功能）
  --help, -h     显示此帮助信息

示例:
  $0                    # 启动完整简化Mock服务
  $0 --auto-open        # 启动并打开浏览器
  $0 --simple           # 启动最简版本

EOF
}

# 解析命令行参数
parse_arguments() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --auto-open)
                AUTO_OPEN=true
                shift
                ;;
            --simple)
                SIMPLE_MODE=true
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

# 创建简化Mock服务脚本
create_simple_mock_script() {
    local script_file="$PROJECT_ROOT/temp-simple-mock.js"
    
    cat > "$script_file" << 'EOF'
const express = require('express');
const cors = require('cors');

console.log('🚀 Starting OpenPenPal Simple Mock Services...');

// Write Service (8001)
const writeApp = express();
writeApp.use(cors());
writeApp.use(express.json());

// Plaza API
writeApp.get('/plaza/posts', (req, res) => {
    res.json({
        success: true,
        data: {
            posts: [
                {
                    id: '1',
                    title: '欢迎来到OpenPenPal Plaza!',
                    content: '这里是校园信件交流的主要广场',
                    author: '系统管理员',
                    created_at: '2025-01-22T10:00:00Z',
                    likes: 42
                },
                {
                    id: '2', 
                    title: '今日最佳信件分享',
                    content: '看看大家都写了什么有趣的内容',
                    author: '用户001',
                    created_at: '2025-01-22T09:30:00Z',
                    likes: 28
                }
            ],
            total: 2
        }
    });
});

// Auth API
writeApp.post('/auth/login', (req, res) => {
    res.json({
        success: true,
        data: {
            token: 'mock-jwt-token-' + Date.now(),
            user: {
                id: 'test-user-1',
                username: req.body.username || 'testuser',
                email: 'test@example.com',
                nickname: '测试用户',
                role: 'user',
                school_code: 'BJDX01',
                school_name: '北京大学',
                permissions: ['read', 'write']
            }
        }
    });
});

writeApp.post('/auth/register', (req, res) => {
    res.json({ success: true, message: '注册成功' });
});

writeApp.get('/auth/me', (req, res) => {
    res.json({
        success: true,
        data: {
            id: 'test-user-1',
            username: 'testuser',
            email: 'test@example.com',
            nickname: '测试用户',
            role: 'user',
            school_code: 'BJDX01',
            school_name: '北京大学'
        }
    });
});

// Schools API
writeApp.get('/schools/search', (req, res) => {
    res.json({
        success: true,
        data: {
            items: [
                { id: '1', code: 'BJDX01', name: '北京大学', province: '北京', city: '北京', type: 'university', status: 'active' },
                { id: '2', code: 'THU001', name: '清华大学', province: '北京', city: '北京', type: 'university', status: 'active' },
                { id: '3', code: 'BJFU02', name: '北京林业大学', province: '北京', city: '北京', type: 'university', status: 'active' }
            ],
            total: 3
        }
    });
});

writeApp.get('/schools/provinces', (req, res) => {
    res.json({
        success: true,
        data: ['北京', '上海', '广东', '江苏', '浙江', '山东', '四川', '湖北']
    });
});

writeApp.listen(8001, () => {
    console.log('📝 Write Service running on port 8001');
});

// Simple API Gateway (8000)
const { createProxyMiddleware } = require('http-proxy-middleware');
const gatewayApp = express();
gatewayApp.use(cors({
    origin: ['http://localhost:3000', 'http://localhost:3001'],
    credentials: true
}));

// Health check
gatewayApp.get('/api/v1/health', (req, res) => {
    res.json({ status: 'healthy', timestamp: new Date().toISOString(), services: ['write-service'] });
});

gatewayApp.get('/health', (req, res) => {
    res.json({ status: 'healthy', timestamp: new Date().toISOString() });
});

// Proxy to write service
gatewayApp.use('/api/v1', createProxyMiddleware({
    target: 'http://localhost:8001',
    changeOrigin: true,
    pathRewrite: { '^/api/v1': '' },
    onError: (err, req, res) => {
        console.error('Proxy error:', err.message);
        res.status(502).json({ error: 'Gateway error', message: err.message });
    },
    onProxyReq: (proxyReq, req) => {
        console.log(`Proxying ${req.method} ${req.url} to ${proxyReq.path}`);
    }
}));

gatewayApp.listen(8000, () => {
    console.log('🚪 API Gateway running on port 8000');
    console.log('Route: /api/v1/* → http://localhost:8001/*');
});

console.log('✅ All simple services started successfully!');
console.log('📱 Frontend URL: http://localhost:3000');
console.log('🚪 API Gateway: http://localhost:8000');
console.log('🏥 Health Check: http://localhost:8000/health');
EOF

    echo "$script_file"
}

# 启动简化Mock服务
start_simple_mock() {
    log_info "启动简化Mock服务..."
    
    # 检查端口
    if ! check_port_available 8000; then
        log_error "端口 8000 被占用"
        return 1
    fi
    
    if ! check_port_available 8001; then
        log_error "端口 8001 被占用"
        return 1
    fi
    
    # 创建脚本
    local script_file=$(create_simple_mock_script)
    local log_file="$LOG_DIR/simple-mock.log"
    local pid_file="$LOG_DIR/simple-mock.pid"
    
    # 启动服务
    cd "$PROJECT_ROOT"
    nohup node "$script_file" > "$log_file" 2>&1 &
    local pid=$!
    echo $pid > "$pid_file"
    
    # 等待服务启动
    if wait_for_service 8000 "简化Mock服务" 30; then
        log_success "简化Mock服务启动成功 (PID: $pid)"
        
        # 验证写信服务
        if wait_for_service 8001 "写信服务" 10; then
            log_success "写信服务启动成功"
        else
            log_warning "写信服务启动可能有问题"
        fi
        
        return 0
    else
        log_error "简化Mock服务启动失败"
        return 1
    fi
}

# 打开浏览器
open_browser() {
    if [ "$AUTO_OPEN" = true ]; then
        log_info "正在打开浏览器..."
        sleep 2
        
        if command_exists open; then
            open "$FRONTEND_URL"
        elif command_exists xdg-open; then
            xdg-open "$FRONTEND_URL"
        else
            log_info "无法自动打开浏览器，请手动访问: $FRONTEND_URL"
        fi
    fi
}

# 显示启动结果
show_result() {
    log_info ""
    log_success "🎉 简化Mock服务启动完成！"
    log_info "=========================="
    log_info "📱 前端地址: $FRONTEND_URL"
    log_info "🚪 API网关: $API_BASE_URL"
    log_info "🏥 健康检查: $API_BASE_URL/health"
    log_info "📝 写信服务: http://localhost:8001"
    log_info ""
    log_info "🔑 测试账号:"
    log_info "  • testuser/secret - 测试用户"
    log_info ""
    log_info "💡 常用命令:"
    log_info "  • 查看状态: ./startup/check-status.sh"
    log_info "  • 查看日志: tail -f logs/simple-mock.log"
    log_info "  • 停止服务: ./startup/stop-all.sh"
    log_info ""
}

# 清理临时文件
cleanup() {
    rm -f "$PROJECT_ROOT/temp-simple-mock.js"
}

# 主函数
main() {
    # 解析参数
    parse_arguments "$@"
    
    # 显示启动信息
    log_info "🚀 OpenPenPal 简化Mock服务启动器"
    if [ "$SIMPLE_MODE" = true ]; then
        log_info "模式: 简化模式"
    else
        log_info "模式: 标准模式"
    fi
    log_info ""
    
    # 检查端口占用
    log_info "检查端口占用..."
    if ! check_port_available 8000 || ! check_port_available 8001; then
        log_warning "检测到端口占用，正在清理..."
        "$SCRIPT_DIR/stop-all.sh" --quiet || true
        sleep 2
    fi
    
    # 创建必需目录
    create_directories
    
    # 启动服务
    if start_simple_mock; then
        show_result
        open_browser
        
        # 设置退出处理
        trap cleanup EXIT
        
        log_info "按 Ctrl+C 停止服务"
        
        # 保持运行
        while true; do
            sleep 10
            # 检查服务是否还在运行
            if ! check_port_occupied 8000; then
                log_warning "检测到服务停止"
                break
            fi
        done
    else
        log_error "服务启动失败"
        cleanup
        exit 1
    fi
}

# 执行主函数
main "$@"