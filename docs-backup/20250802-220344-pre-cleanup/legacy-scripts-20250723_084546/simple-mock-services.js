/**
 * OpenPenPal 简化版 Mock 服务
 * 基于原有start-integration.sh中的动态生成逻辑
 * 提供基础的API mock功能，便于快速开发和测试
 */

const express = require('express');
const cors = require('cors');

console.log('🚀 启动 OpenPenPal 简化版 Mock 服务...');

// 写信服务 (8001)
const writeApp = express();
writeApp.use(cors());
writeApp.use(express.json());

// 通用响应格式
const successResponse = (data, message = '操作成功') => ({
    success: true,
    code: 0,
    message,
    data,
    timestamp: new Date().toISOString()
});

const errorResponse = (message, code = 500) => ({
    success: false,
    code,
    message,
    timestamp: new Date().toISOString()
});

// 认证相关接口
writeApp.post('/auth/login', (req, res) => {
    const { username, password } = req.body;
    
    // 简化的用户验证
    const users = {
        'alice': { password: 'secret', role: 'student', school: 'PKU' },
        'bob': { password: 'password123', role: 'student', school: 'THU' },
        'admin': { password: 'admin123', role: 'admin', school: 'ADMIN' },
        'courier1': { password: 'courier123', role: 'courier', school: 'PKU' }
    };
    
    const user = users[username];
    if (!user || user.password !== password) {
        return res.status(401).json(errorResponse('用户名或密码错误', 401));
    }
    
    res.json(successResponse({
        token: `mock-jwt-token-${username}-${Date.now()}`,
        user: {
            id: `user_${username}`,
            username,
            email: `${username}@example.com`,
            nickname: username === 'alice' ? '爱丽丝' : username === 'bob' ? '鲍勃' : username,
            role: user.role,
            school_code: user.school,
            school_name: user.school === 'PKU' ? '北京大学' : user.school === 'THU' ? '清华大学' : '系统管理',
            permissions: user.role === 'admin' ? ['ALL'] : ['read', 'write']
        }
    }, '登录成功'));
});

writeApp.post('/auth/register', (req, res) => {
    res.json(successResponse({ id: 'new_user_' + Date.now() }, '注册成功'));
});

writeApp.get('/auth/me', (req, res) => {
    res.json(successResponse({
        id: 'test-user-1',
        username: 'testuser',
        email: 'test@example.com',
        nickname: '测试用户',
        role: 'user',
        school_code: 'BJDX01',
        school_name: '北京大学'
    }));
});

// 写信相关接口
writeApp.get('/api/letters', (req, res) => {
    const mockLetters = [
        {
            id: 'letter_1',
            title: '给远方朋友的信',
            content: '这是一封测试信件...',
            sender: '爱丽丝',
            receiver_hint: '北京大学的朋友',
            status: 'pending',
            created_at: '2024-01-15T10:30:00Z'
        },
        {
            id: 'letter_2', 
            title: '关于大学生活',
            content: '分享一些大学生活的感悟...',
            sender: '鲍勃',
            receiver_hint: '清华大学的同学',
            status: 'delivered',
            created_at: '2024-01-16T14:20:00Z'
        }
    ];
    
    res.json(successResponse({
        items: mockLetters,
        total: mockLetters.length,
        page: 0,
        pageSize: 20
    }));
});

writeApp.post('/api/letters', (req, res) => {
    const { title, content, receiver_hint } = req.body;
    
    const newLetter = {
        id: 'letter_' + Date.now(),
        title,
        content,
        receiver_hint,
        sender: '当前用户',
        status: 'pending',
        created_at: new Date().toISOString()
    };
    
    res.json(successResponse(newLetter, '信件创建成功'));
});

// 健康检查
writeApp.get('/health', (req, res) => {
    res.json({ status: 'healthy', service: 'write-service', timestamp: new Date().toISOString() });
});

// 启动写信服务
writeApp.listen(8001, () => {
    console.log('✅ 写信服务已启动: http://localhost:8001');
});

// 信使服务 (8002)
const courierApp = express();
courierApp.use(cors());
courierApp.use(express.json());

// 任务相关接口
courierApp.get('/api/tasks', (req, res) => {
    const mockTasks = [
        {
            id: 'task_1',
            letter_id: 'letter_1',
            pickup_location: '北京大学邮局',
            delivery_location: '清华大学邮局',
            status: 'available',
            reward: 15.00,
            estimated_time: 120,
            created_at: '2024-01-16T09:00:00Z'
        },
        {
            id: 'task_2',
            letter_id: 'letter_2', 
            pickup_location: '清华大学邮局',
            delivery_location: '北京大学邮局',
            status: 'assigned',
            reward: 12.00,
            estimated_time: 90,
            created_at: '2024-01-17T08:30:00Z'
        }
    ];
    
    res.json(successResponse({
        items: mockTasks,
        total: mockTasks.length
    }));
});

courierApp.post('/api/courier/apply', (req, res) => {
    res.json(successResponse({ application_id: 'app_' + Date.now() }, '信使申请已提交'));
});

courierApp.get('/health', (req, res) => {
    res.json({ status: 'healthy', service: 'courier-service', timestamp: new Date().toISOString() });
});

courierApp.listen(8002, () => {
    console.log('✅ 信使服务已启动: http://localhost:8002');
});

// 管理服务 (8003)
const adminApp = express();
adminApp.use(cors());
adminApp.use(express.json());

// 管理员认证
adminApp.post('/api/admin/auth/login', (req, res) => {
    const { username, password } = req.body;
    
    if (username === 'admin' && password === 'admin123') {
        res.json(successResponse({
            token: `admin-token-${Date.now()}`,
            user: {
                id: 'admin_001',
                username: 'admin',
                role: 'super_admin',
                permissions: ['ALL']
            }
        }, '管理员登录成功'));
    } else {
        res.status(401).json(errorResponse('管理员认证失败', 401));
    }
});

// 用户管理
adminApp.get('/api/admin/users', (req, res) => {
    const mockUsers = [
        {
            id: 'user_001',
            username: 'alice',
            email: 'alice@pku.edu.cn',
            role: 'student',
            school: '北京大学',
            status: 'active',
            created_at: '2024-01-10T00:00:00Z'
        },
        {
            id: 'user_002',
            username: 'bob', 
            email: 'bob@tsinghua.edu.cn',
            role: 'student',
            school: '清华大学',
            status: 'active',
            created_at: '2024-01-12T00:00:00Z'
        }
    ];
    
    res.json(successResponse({
        items: mockUsers,
        pagination: {
            page: 0,
            limit: 20,
            total: mockUsers.length
        }
    }));
});

// 系统配置
adminApp.get('/api/admin/system/config', (req, res) => {
    res.json(successResponse({
        max_letter_length: 2000,
        delivery_timeout: 72,
        auto_match_enabled: true,
        maintenance_mode: false
    }));
});

// 博物馆管理
adminApp.get('/api/admin/museum/exhibitions', (req, res) => {
    const mockExhibitions = [
        {
            id: 'exhibition_001',
            title: '冬日温暖信件展',
            description: '收录冬季主题的温暖信件',
            status: 'active',
            letter_count: 15,
            created_at: '2024-01-15T00:00:00Z'
        }
    ];
    
    res.json(successResponse({
        items: mockExhibitions,
        pagination: {
            page: 0,
            limit: 20, 
            total: mockExhibitions.length
        }
    }));
});

adminApp.get('/health', (req, res) => {
    res.json({ status: 'healthy', service: 'admin-service', timestamp: new Date().toISOString() });
});

adminApp.listen(8003, () => {
    console.log('✅ 管理服务已启动: http://localhost:8003');
});

// OCR服务 (8004)
const ocrApp = express();
ocrApp.use(cors());
ocrApp.use(express.json());

ocrApp.get('/api/ocr/models', (req, res) => {
    const models = [
        {
            id: 'general',
            name: '通用文字识别',
            description: '适用于各种类型的文字识别',
            accuracy: 0.95
        },
        {
            id: 'handwriting',
            name: '手写文字识别', 
            description: '专门用于手写文字的识别',
            accuracy: 0.88
        }
    ];
    
    res.json(successResponse(models));
});

ocrApp.post('/api/ocr/process', (req, res) => {
    const { image_url } = req.body;
    
    // 模拟OCR处理
    setTimeout(() => {
        res.json(successResponse({
            id: 'ocr_' + Date.now(),
            text: '这是识别出的文字内容示例',
            confidence: 0.92,
            processing_time: 1.5
        }, 'OCR处理完成'));
    }, 1500);
});

ocrApp.get('/health', (req, res) => {
    res.json({ status: 'healthy', service: 'ocr-service', timestamp: new Date().toISOString() });
});

ocrApp.listen(8004, () => {
    console.log('✅ OCR服务已启动: http://localhost:8004');
});

// API网关 (8000)
const gatewayApp = express();
gatewayApp.use(cors());
gatewayApp.use(express.json());

// 代理到各个服务
const proxy = (targetPort) => (req, res) => {
    const http = require('http');
    const url = require('url');
    
    const options = {
        hostname: 'localhost',
        port: targetPort,
        path: req.url,
        method: req.method,
        headers: req.headers
    };
    
    const proxyReq = http.request(options, (proxyRes) => {
        res.writeHead(proxyRes.statusCode, proxyRes.headers);
        proxyRes.pipe(res);
    });
    
    proxyReq.on('error', (err) => {
        res.status(500).json(errorResponse('服务不可用'));
    });
    
    if (req.body) {
        proxyReq.write(JSON.stringify(req.body));
    }
    proxyReq.end();
};

// 路由配置
gatewayApp.use('/api/auth', proxy(8001));
gatewayApp.use('/api/write', proxy(8001));
gatewayApp.use('/api/courier', proxy(8002));
gatewayApp.use('/api/admin', proxy(8003));
gatewayApp.use('/api/ocr', proxy(8004));

// 健康检查
gatewayApp.get('/health', (req, res) => {
    res.json({ 
        status: 'healthy', 
        service: 'api-gateway',
        services: {
            'write-service': 'http://localhost:8001',
            'courier-service': 'http://localhost:8002', 
            'admin-service': 'http://localhost:8003',
            'ocr-service': 'http://localhost:8004'
        },
        timestamp: new Date().toISOString() 
    });
});

gatewayApp.listen(8000, () => {
    console.log('✅ API网关已启动: http://localhost:8000');
    console.log('');
    console.log('🎉 OpenPenPal 简化版 Mock 服务全部启动完成！');
    console.log('');
    console.log('📋 服务列表:');
    console.log('   • API网关: http://localhost:8000');
    console.log('   • 写信服务: http://localhost:8001');
    console.log('   • 信使服务: http://localhost:8002');
    console.log('   • 管理服务: http://localhost:8003');
    console.log('   • OCR服务: http://localhost:8004');
    console.log('');
    console.log('🔑 测试账号:');
    console.log('   • alice/secret - 学生用户');
    console.log('   • admin/admin123 - 管理员');
    console.log('   • courier1/courier123 - 信使');
    console.log('');
    console.log('📖 API 文档: http://localhost:8000/health');
});

// 优雅关闭
process.on('SIGTERM', () => {
    console.log('🛑 正在关闭 Mock 服务...');
    process.exit(0);
});

process.on('SIGINT', () => {
    console.log('🛑 正在关闭 Mock 服务...');
    process.exit(0);
});