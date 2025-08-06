#!/bin/bash

# OpenPenPal 集成启动脚本 (传统模式)
# 基于原 start-integration.sh 的完整功能，但使用新的工具函数

set -e

# 获取脚本目录和项目根目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# 导入工具函数和环境变量
source "$SCRIPT_DIR/utils.sh"
source "$SCRIPT_DIR/environment-vars.sh"

# 默认选项
AUTO_OPEN=true
SKIP_DEPS=false

# 显示帮助信息
show_help() {
    cat << EOF
OpenPenPal 集成启动脚本 (传统模式)

用法: $0 [选项]

选项:
  --no-open      不自动打开浏览器
  --skip-deps    跳过依赖检查
  --help, -h     显示此帮助信息

示例:
  $0                    # 完整启动流程
  $0 --no-open          # 启动但不打开浏览器
  $0 --skip-deps        # 跳过依赖检查

说明:
  这是传统的集成启动模式，保持与原 start-integration.sh 的兼容性。
  推荐使用新的启动系统: ./startup/quick-start.sh

EOF
}

# 解析命令行参数
parse_arguments() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --no-open)
                AUTO_OPEN=false
                shift
                ;;
            --skip-deps)
                SKIP_DEPS=true
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

# 检查和安装依赖
check_dependencies() {
    if [ "$SKIP_DEPS" = true ]; then
        log_info "跳过依赖检查"
        return 0
    fi
    
    log_info "检查前端依赖..."
    cd "$PROJECT_ROOT/frontend"
    
    if [ ! -d "node_modules" ] || [ ! -f "node_modules/.installed" ]; then
        log_step "首次运行，安装前端依赖包..."
        npm install
        if [ $? -ne 0 ]; then
            log_error "前端依赖安装失败"
            exit 1
        fi
        touch node_modules/.installed
        log_success "前端依赖安装完成"
    else
        log_success "前端依赖已存在"
    fi
    
    cd "$PROJECT_ROOT"
}

# 创建传统Mock服务脚本
create_integration_mock_script() {
    local script_file="$PROJECT_ROOT/temp-integration-mock.js"
    
    cat > "$script_file" << 'EOF'
const express = require('express');
const cors = require('cors');

// 写信服务 (8001)
const writeApp = express();
writeApp.use(cors());
writeApp.use(express.json());

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

writeApp.listen(8001, () => console.log('📝 Write Service running on port 8001'));

// 信使服务 (8002)
const courierApp = express();
courierApp.use(cors());
courierApp.use(express.json());

courierApp.get('/courier/info', (req, res) => {
    res.json({
        success: true,
        data: { id: 'courier-1', level: 1, region: '北京大学', total_points: 150, completed_tasks: 5 }
    });
});

courierApp.listen(8002, () => console.log('🏃‍♂️ Courier Service running on port 8002'));

// 管理服务 (8003)
const adminApp = express();
adminApp.use(cors());
adminApp.use(express.json());

adminApp.get('/api/admin/dashboard', (req, res) => {
    res.json({
        success: true,
        data: { users: { total: 100, active: 85 }, letters: { total: 500, sent_today: 25 }, couriers: { total: 20, active: 15 } }
    });
});

adminApp.listen(8003, () => console.log('👨‍💼 Admin Service running on port 8003'));

// OCR服务 (8004)
const ocrApp = express();
ocrApp.use(cors());
ocrApp.use(express.json());

ocrApp.post('/ocr/process', (req, res) => {
    res.json({ success: true, data: { text: '这是OCR识别的文字内容', confidence: 0.95 } });
});

ocrApp.listen(8004, () => console.log('🔍 OCR Service running on port 8004'));
EOF

    echo "$script_file"
}

# 创建传统网关脚本
create_integration_gateway_script() {
    local script_file="$PROJECT_ROOT/temp-integration-gateway.js"
    
    cat > "$script_file" << 'EOF'
const express = require('express');
const { createProxyMiddleware } = require('http-proxy-middleware');
const cors = require('cors');

const app = express();
const PORT = process.env.GATEWAY_PORT || 8000;

// 中间件设置
app.use(cors({
    origin: ['http://localhost:3000', 'http://localhost:3001'],
    credentials: true
}));

app.use(express.json());

// 健康检查
app.get('/api/v1/health', (req, res) => {
    res.json({
        status: 'healthy',
        timestamp: new Date().toISOString(),
        services: {
            gateway: 'running',
            write_service: 'running',
            courier_service: 'running',
            admin_service: 'running',
            ocr_service: 'running'
        }
    });
});

// 路由代理配置
const services = [
    { path: '/api/v1/auth', target: 'http://localhost:8001', pathRewrite: { '^/api/v1/auth': '/auth' } },
    { path: '/api/v1/schools', target: 'http://localhost:8001', pathRewrite: { '^/api/v1/schools': '/schools' } },
    { path: '/api/v1/letters', target: 'http://localhost:8001', pathRewrite: { '^/api/v1/letters': '/letters' } },
    { path: '/api/v1/courier', target: 'http://localhost:8002', pathRewrite: { '^/api/v1/courier': '/courier' } },
    { path: '/api/v1/admin', target: 'http://localhost:8003', pathRewrite: { '^/api/v1/admin': '/api/admin' } },
    { path: '/api/v1/ocr', target: 'http://localhost:8004', pathRewrite: { '^/api/v1/ocr': '/ocr' } }
];

// 配置代理中间件
services.forEach(service => {
    app.use(service.path, createProxyMiddleware({
        target: service.target,
        changeOrigin: true,
        pathRewrite: service.pathRewrite,
        onError: (err, req, res) => {
            console.error(`Proxy error for ${service.path}:`, err.message);
            res.status(502).json({ 
                error: 'Service Unavailable', 
                message: `${service.path} service is not responding`,
                timestamp: new Date().toISOString()
            });
        },
        onProxyReq: (proxyReq, req) => {
            console.log(`[GATEWAY] ${req.method} ${req.url} → ${service.target}${proxyReq.path}`);
        }
    }));
});

// 启动网关
app.listen(PORT, () => {
    console.log(`🚪 API Gateway running on port ${PORT}`);
    console.log('Routes configured:');
    services.forEach(service => {
        console.log(`  ${service.path} → ${service.target}`);
    });
});
EOF

    echo "$script_file"
}

# 启动Mock服务
start_mock_services() {
    log_info "启动模拟后端服务..."
    
    # 检查端口
    local ports=(8001 8002 8003 8004)
    for port in "${ports[@]}"; do
        if ! check_port_available $port; then
            log_error "端口 $port 被占用"
            return 1
        fi
    done
    
    # 创建并启动Mock服务
    local script_file=$(create_integration_mock_script)
    local log_file="$LOG_DIR/integration-mock.log"
    local pid_file="$LOG_DIR/integration-mock.pid"
    
    nohup node "$script_file" > "$log_file" 2>&1 &
    local pid=$!
    echo $pid > "$pid_file"
    
    # 等待服务启动
    sleep 3
    
    if wait_for_service 8001 "写信服务" 10; then
        log_success "模拟服务启动成功 (PID: $pid)"
        return 0
    else
        log_error "模拟服务启动失败"
        return 1
    fi
}

# 启动API网关
start_api_gateway() {
    log_info "启动 API 网关..."
    
    if ! check_port_available 8000; then
        log_error "端口 8000 被占用"
        return 1
    fi
    
    # 创建并启动网关
    local script_file=$(create_integration_gateway_script)
    local log_file="$LOG_DIR/integration-gateway.log"
    local pid_file="$LOG_DIR/integration-gateway.pid"
    
    nohup node "$script_file" > "$log_file" 2>&1 &
    local pid=$!
    echo $pid > "$pid_file"
    
    sleep 3
    
    # 检查网关是否成功启动
    if ps -p $pid > /dev/null 2>&1 && wait_for_service 8000 "API网关" 10; then
        log_success "API网关启动成功 (PID: $pid)"
        return 0
    else
        log_error "API网关启动失败"
        if [ -f "$log_file" ]; then
            log_info "查看错误日志："
            tail -n 10 "$log_file"
        fi
        return 1
    fi
}

# 启动前端
start_frontend() {
    log_info "启动前端应用..."
    
    if ! check_port_available 3000; then
        log_error "端口 3000 被占用"
        return 1
    fi
    
    cd "$PROJECT_ROOT/frontend"
    
    local log_file="$LOG_DIR/integration-frontend.log"
    local pid_file="$LOG_DIR/integration-frontend.pid"
    
    nohup npm run dev > "$log_file" 2>&1 &
    local pid=$!
    echo $pid > "$pid_file"
    
    cd "$PROJECT_ROOT"
    
    # 等待前端启动
    if wait_for_service 3000 "前端应用" 30; then
        log_success "前端应用启动成功 (PID: $pid)"
        return 0
    else
        log_error "前端应用启动失败"
        return 1
    fi
}

# 验证所有服务
verify_all_services() {
    log_info "验证服务状态..."
    
    local services=(
        "3000:前端应用"
        "8000:API网关"
        "8001:写信服务"
        "8002:信使服务"
        "8003:管理服务"
        "8004:OCR服务"
    )
    
    local all_healthy=true
    
    for service_info in "${services[@]}"; do
        local port="${service_info%%:*}"
        local name="${service_info##*:}"
        
        if check_port_occupied $port; then
            log_success "$name (端口 $port): 运行中"
        else
            log_error "$name (端口 $port): 启动失败"
            all_healthy=false
        fi
    done
    
    if [ "$all_healthy" = true ]; then
        log_success "所有服务启动成功！"
        return 0
    else
        log_warning "部分服务启动失败，请检查日志"
        return 1
    fi
}

# 打开浏览器
open_browser() {
    if [ "$AUTO_OPEN" = true ]; then
        log_info "正在自动打开浏览器..."
        sleep 2
        
        if command_exists open; then
            open "$FRONTEND_URL"
            sleep 1
            open "$API_BASE_URL/api/v1/health"
        elif command_exists xdg-open; then
            xdg-open "$FRONTEND_URL"
            sleep 1
            xdg-open "$API_BASE_URL/api/v1/health"
        else
            log_info "无法自动打开浏览器，请手动访问:"
            log_info "  前端: $FRONTEND_URL"
            log_info "  健康检查: $API_BASE_URL/api/v1/health"
        fi
    fi
}

# 显示启动结果
show_result() {
    log_info ""
    log_success "🌐 OpenPenPal 集成环境已启动！"
    log_info "========================================"
    log_info "📱 前端主页: $FRONTEND_URL"
    log_info "🚪 API网关: $API_BASE_URL"
    log_info "🏥 健康检查: $API_BASE_URL/api/v1/health"
    log_info "========================================"
    log_info "👤 快速测试账号:"
    log_info "   用户名: testuser"
    log_info "   密码: secret"
    log_info "   学校: 北京大学 (BJDX01)"
    log_info "========================================"
    log_info "💡 使用提示:"
    log_info "   • 请在浏览器中打开上述链接"
    log_info "   • 按 Ctrl+C 可停止所有服务"
    log_info "   • 查看日志: tail -f logs/*.log"
    log_info "========================================"
}

# 清理临时文件
cleanup() {
    rm -f "$PROJECT_ROOT/temp-integration-mock.js"
    rm -f "$PROJECT_ROOT/temp-integration-gateway.js"
}

# 退出处理
cleanup_on_exit() {
    log_info "正在停止所有服务..."
    "$SCRIPT_DIR/stop-all.sh" --quiet || true
    cleanup
    log_success "服务已停止"
}

# 主函数
main() {
    # 解析参数
    parse_arguments "$@"
    
    # 显示启动信息
    log_info "🚀 OpenPenPal 前后端集成启动程序 (传统模式)"
    log_info "================================================="
    log_info "当前目录: $PROJECT_ROOT"
    log_info ""
    
    # 检查运行环境
    if ! command_exists node; then
        log_error "Node.js 未安装，请先安装 Node.js 18+"
        exit 1
    fi
    
    if ! command_exists npm; then
        log_error "npm 未安装"
        exit 1
    fi
    
    log_success "✅ Node.js $(node --version)"
    log_success "✅ npm $(npm --version)"
    
    # 清理可能存在的服务
    log_info "清理可能运行的服务..."
    "$SCRIPT_DIR/stop-all.sh" --quiet || true
    sleep 2
    
    # 创建必需目录
    create_directories
    
    # 检查依赖
    check_dependencies
    
    # 启动服务
    log_info "启动集成服务..."
    
    if start_mock_services && start_api_gateway && start_frontend; then
        if verify_all_services; then
            show_result
            open_browser
            
            # 设置退出处理
            trap cleanup_on_exit INT TERM EXIT
            
            log_info "环境运行中，请在浏览器中体验 OpenPenPal"
            log_info "按 Ctrl+C 停止所有服务"
            
            # 保持脚本运行
            while true; do
                sleep 10
                # 检查关键服务是否还在运行
                if ! check_port_occupied 3000 && ! check_port_occupied 8000; then
                    log_warning "检测到服务异常停止"
                    break
                fi
            done
        else
            log_error "服务验证失败"
            cleanup
            exit 1
        fi
    else
        log_error "服务启动失败"
        cleanup
        exit 1
    fi
}

# 执行主函数
main "$@"