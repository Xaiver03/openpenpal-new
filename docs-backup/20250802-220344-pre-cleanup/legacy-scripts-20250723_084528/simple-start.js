const express = require('express');
const cors = require('cors');

console.log('🚀 Starting OpenPenPal Mock Services...');

// Write Service (8001)
const writeApp = express();
writeApp.use(cors());
writeApp.use(express.json());

writeApp.get('/plaza/posts', (req, res) => {
    res.json({
        success: true,
        data: {
            posts: [
                {
                    id: '1',
                    title: '欢迎来到OpenPenPal Plaza\!',
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
                school_name: '北京大学'
            }
        }
    });
});

writeApp.get('/auth/me', (req, res) => {
    res.json({
        success: true,
        data: {
            id: 'test-user-1',
            username: 'testuser',
            email: 'test@example.com',
            nickname: '测试用户',
            role: 'user'
        }
    });
});

writeApp.listen(8001, () => {
    console.log('📝 Write Service running on port 8001');
});

// Simple API Gateway (8000)
const { createProxyMiddleware } = require('http-proxy-middleware');
const gatewayApp = express();
gatewayApp.use(cors());

// Health check
gatewayApp.get('/api/v1/health', (req, res) => {
    res.json({ status: 'healthy', timestamp: new Date().toISOString() });
});

// Proxy to write service
gatewayApp.use('/api/v1', createProxyMiddleware({
    target: 'http://localhost:8001',
    changeOrigin: true,
    pathRewrite: { '^/api/v1': '' }
}));

gatewayApp.listen(8000, () => {
    console.log('🚪 API Gateway running on port 8000');
});

console.log('✅ All services started successfully\!');
EOF < /dev/null