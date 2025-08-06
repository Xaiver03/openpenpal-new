/**
 * OpenPenPal 模拟后端服务
 * 为前端开发提供模拟API响应
 */

const express = require('express');
const cors = require('cors');

console.log('🚀 Starting OpenPenPal Mock Services...');

// 写信服务 (8001)
const writeApp = express();
writeApp.use(cors());
writeApp.use(express.json());

// Plaza页面API - 解决网络错误
writeApp.get('/plaza/posts', (req, res) => {
    res.json({
        success: true,
        data: {
            posts: [
                {
                    id: '1',
                    title: '欢迎来到OpenPenPal Plaza!',
                    content: '这里是校园信件交流的主要广场，大家可以分享有趣的信件内容。',
                    author: '系统管理员',
                    author_avatar: '/images/system-admin.png',
                    created_at: '2025-01-22T10:00:00Z',
                    likes: 42,
                    comments: 8,
                    tags: ['欢迎', '系统公告']
                },
                {
                    id: '2', 
                    title: '今日最佳信件分享',
                    content: '看看大家都写了什么有趣的内容，温暖的文字总能触动人心。',
                    author: '张小明',
                    author_avatar: '/images/user-001.png',
                    created_at: '2025-01-22T09:30:00Z',
                    likes: 28,
                    comments: 12,
                    tags: ['分享', '优质内容']
                },
                {
                    id: '3',
                    title: '信使招募活动开始啦！',
                    content: '想要成为连接校园的纽带吗？加入我们的信使团队，体验不一样的校园生活。',
                    author: '信使管理员',
                    author_avatar: '/images/courier-admin.png',
                    created_at: '2025-01-22T08:45:00Z',
                    likes: 35,
                    comments: 15,
                    tags: ['招募', '信使', '活动']
                }
            ],
            total: 3,
            page: 1,
            per_page: 10
        }
    });
});

// 认证相关API
writeApp.post('/auth/login', (req, res) => {
    const { username, password } = req.body;
    
    // 模拟登录验证
    if (username && password) {
        res.json({
            success: true,
            data: {
                token: 'mock-jwt-token-' + Date.now(),
                user: {
                    id: 'test-user-1',
                    username: username,
                    email: 'test@example.com',
                    nickname: '测试用户',
                    role: 'user',
                    school_code: 'BJDX01',
                    school_name: '北京大学',
                    permissions: ['read', 'write', 'comment']
                }
            }
        });
    } else {
        res.status(400).json({
            success: false,
            message: '用户名和密码不能为空'
        });
    }
});

writeApp.post('/auth/register', (req, res) => {
    res.json({ 
        success: true, 
        message: '注册成功，请等待邮箱验证',
        data: {
            user_id: 'new-user-' + Date.now()
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
            role: 'user',
            school_code: 'BJDX01',
            school_name: '北京大学',
            avatar: '/images/default-avatar.png'
        }
    });
});

// 学校相关API
writeApp.get('/schools/search', (req, res) => {
    const { keyword = '', province = '' } = req.query;
    
    let schools = [
        { id: '1', code: 'BJDX01', name: '北京大学', province: '北京', city: '北京', type: 'university', status: 'active' },
        { id: '2', code: 'THU001', name: '清华大学', province: '北京', city: '北京', type: 'university', status: 'active' },
        { id: '3', code: 'BJFU02', name: '北京林业大学', province: '北京', city: '北京', type: 'university', status: 'active' },
        { id: '4', code: 'SJTU01', name: '上海交通大学', province: '上海', city: '上海', type: 'university', status: 'active' },
        { id: '5', code: 'ZJU001', name: '浙江大学', province: '浙江', city: '杭州', type: 'university', status: 'active' }
    ];
    
    // 简单搜索过滤
    if (keyword) {
        schools = schools.filter(s => s.name.includes(keyword) || s.code.includes(keyword));
    }
    if (province) {
        schools = schools.filter(s => s.province === province);
    }
    
    res.json({
        success: true,
        data: {
            items: schools,
            total: schools.length
        }
    });
});

writeApp.get('/schools/provinces', (req, res) => {
    res.json({
        success: true,
        data: ['北京', '上海', '广东', '江苏', '浙江', '山东', '四川', '湖北', '陕西', '湖南']
    });
});

// 博物馆模块API
writeApp.get('/museum/posts', (req, res) => {
    const { era = 'all', type = 'all', page = 1, limit = 10 } = req.query;
    
    let posts = [
        {
            id: 'museum-1',
            title: '写给十年后的自己',
            content: '亲爱的未来的我，希望你还记得现在这个青涩但满怀梦想的自己...',
            author: '匿名用户',
            category: 'future',
            era: '2024',
            tags: ['未来信', '青春', '梦想'],
            likes: 156,
            views: 2340,
            created_at: '2024-11-15T14:30:00Z',
            featured: true
        },
        {
            id: 'museum-2', 
            title: '致远方的朋友',
            content: '虽然我们相隔千里，但这份友谊跨越了时空的距离...',
            author: '校园诗人',
            category: 'drift',
            era: '2024',
            tags: ['漂流信', '友谊', '温暖'],
            likes: 89,
            views: 1520,
            created_at: '2024-10-28T09:15:00Z',
            featured: false
        },
        {
            id: 'museum-3',
            title: '感谢那个帮助我的陌生人',
            content: '在那个下雨的傍晚，是你为我撑起了一把伞...',
            author: '感恩的心',
            category: 'warm',
            era: '2024',
            tags: ['温暖信', '感恩', '善良'],
            likes: 234,
            views: 3210,
            created_at: '2024-09-20T16:45:00Z',
            featured: true
        }
    ];
    
    // 简单过滤
    if (era !== 'all') {
        posts = posts.filter(p => p.era === era);
    }
    if (type !== 'all') {
        posts = posts.filter(p => p.category === type);
    }
    
    const start = (page - 1) * limit;
    const paginatedPosts = posts.slice(start, start + limit);
    
    res.json({
        success: true,
        data: {
            posts: paginatedPosts,
            total: posts.length,
            page: parseInt(page),
            per_page: parseInt(limit),
            featured: posts.filter(p => p.featured)
        }
    });
});

writeApp.get('/museum/posts/:id', (req, res) => {
    const { id } = req.params;
    
    // 模拟详细信件内容
    const post = {
        id: id,
        title: '写给十年后的自己',
        content: `亲爱的未来的我：
        
当你读到这封信的时候，我希望你已经实现了那些青春年少时的梦想。

还记得现在的我吗？那个在图书馆里挥汗如雨的学生，那个为了一道数学题而熬夜到凌晨的少年。我知道路还很长，但我相信，只要坚持下去，总会到达想要的地方。

十年后的你，是否还会想起这个秋天？想起那些和室友一起刷夜的日子，想起食堂里的那碗热腾腾的面条，想起第一次收到心仪女孩回信时的激动心情？

无论你变成了什么样子，我希望你还能保持这份初心，还能记得那个曾经为了梦想而拼搏的自己。

此致
敬礼！

过去的你
2024年11月15日`,
        author: '时光旅行者',
        author_avatar: '/images/time-traveler.png',
        category: 'future',
        era: '2024',
        tags: ['未来信', '青春', '梦想', '成长'],
        likes: 156,
        views: 2340,
        comments: [
            {
                id: 'comment-1',
                author: '同路人',
                content: '看哭了，这就是我想对未来的自己说的话',
                created_at: '2024-11-16T10:20:00Z'
            },
            {
                id: 'comment-2',
                author: '梦想家',
                content: '十年后的我们会在哪里呢？希望都能成为更好的自己',
                created_at: '2024-11-16T15:30:00Z'
            }
        ],
        created_at: '2024-11-15T14:30:00Z',
        featured: true,
        exhibition_info: {
            theme: '青春记忆',
            curator: '系统管理员',
            exhibition_date: '2024-11-01'
        }
    };
    
    res.json({
        success: true,
        data: post
    });
});

writeApp.post('/museum/submit', (req, res) => {
    const { title, content, author, tags, category } = req.body;
    
    res.json({
        success: true,
        data: {
            submission_id: 'sub-' + Date.now(),
            status: 'pending',
            estimated_review_time: '1-3 个工作日'
        },
        message: '投稿已提交，感谢您的分享！我们会尽快审核并通知您结果。'
    });
});

writeApp.get('/museum/featured', (req, res) => {
    res.json({
        success: true,
        data: {
            current_theme: {
                id: 'theme-winter',
                title: '冬日温情',
                description: '分享那些在寒冷冬日里温暖人心的故事',
                cover_image: '/images/winter-theme.jpg',
                start_date: '2024-12-01',
                end_date: '2024-12-31'
            },
            hot_posts: [
                { id: 'museum-1', title: '写给十年后的自己', views: 2340 },
                { id: 'museum-3', title: '感谢那个帮助我的陌生人', views: 3210 }
            ]
        }
    });
});

// 公共信件API - Plaza页面使用
writeApp.get('/letters/public', (req, res) => {
    const { style, sort_by = 'created_at', sort_order = 'desc', limit = 20 } = req.query;
    
    let letters = [
        {
            id: 'letter-plaza-1',
            title: '写给三年后的自己',
            content: '亲爱的未来的我，当你读到这封信的时候，希望你已经成为了更好的自己。还记得现在的我吗？那个在图书馆里挥汗如雨的学生，那个为了一道数学题而熬夜到凌晨的少年。我知道路还很长，但我相信，只要坚持下去，总会到达想要的地方...',
            user: { nickname: '匿名作者', avatar: '/images/user-001.png' },
            style: 'future',
            created_at: '2024-01-20T10:00:00Z',
            likes: 156,
            views: 892,
            comments: 23,
            is_public: true
        },
        {
            id: 'letter-plaza-2',
            title: '致正在迷茫的你',
            content: '如果你正在经历人生的低谷，请记住这只是暂时的。每个人都会有迷茫的时候，这是成长路上的必经之路。不要害怕迷茫，因为只有经历过黑暗，我们才能更珍惜光明。愿你在迷雾中找到前进的方向，愿你的心永远充满希望...',
            user: { nickname: '温暖使者', avatar: '/images/user-002.png' },
            style: 'warm',
            created_at: '2024-01-19T15:30:00Z',
            likes: 234,
            views: 1247,
            comments: 45,
            is_public: true
        },
        {
            id: 'letter-plaza-3',
            title: '一个关于友谊的故事',
            content: '我想和你分享一个关于友谊的故事，这个故事改变了我对友情的理解。那是大学的第一年，我遇到了我的室友小李。起初我们并不熟悉，甚至有些小摩擦，但后来发生的事情让我明白，真正的友谊能够跨越一切障碍...',
            user: { nickname: '故事讲述者', avatar: '/images/user-003.png' },
            style: 'story',
            created_at: '2024-01-18T20:15:00Z',
            likes: 189,
            views: 756,
            comments: 31,
            is_public: true
        },
        {
            id: 'letter-plaza-4',
            title: '漂流到远方的思念',
            content: '这封信将随风漂流到某个角落，希望能遇到同样思念远方的你。也许我们素不相识，但我们都有过思念的经历。思念是一种神奇的情感，它能让相距千里的人心灵相通，让时光倒流回到最美好的时光...',
            user: { nickname: '漂流者', avatar: '/images/user-004.png' },
            style: 'drift',
            created_at: '2024-01-17T12:45:00Z',
            likes: 167,
            views: 623,
            comments: 18,
            is_public: true
        },
        {
            id: 'letter-plaza-5',
            title: '关于成长的思考',
            content: '成长是什么？成长是学会接受不完美的自己，是在失败中汲取经验，是在困难面前不退缩。每一次的跌倒都是为了更好地站起来，每一次的眼泪都是为了更深刻的理解生活。愿我们都能在成长的路上，成为更好的自己...',
            user: { nickname: '思考者', avatar: '/images/user-005.png' },
            style: 'warm',
            created_at: '2024-01-16T09:20:00Z',
            likes: 203,
            views: 934,
            comments: 28,
            is_public: true
        }
    ];
    
    // 按风格筛选
    if (style && style !== 'all') {
        letters = letters.filter(letter => letter.style === style);
    }
    
    // 排序
    letters.sort((a, b) => {
        switch (sort_by) {
            case 'likes':
                return sort_order === 'desc' ? b.likes - a.likes : a.likes - b.likes;
            case 'views':
                return sort_order === 'desc' ? b.views - a.views : a.views - b.views;
            case 'created_at':
            default:
                const dateA = new Date(a.created_at).getTime();
                const dateB = new Date(b.created_at).getTime();
                return sort_order === 'desc' ? dateB - dateA : dateA - dateB;
        }
    });
    
    // 分页
    const limitNum = parseInt(limit) || 20;
    const paginatedLetters = letters.slice(0, limitNum);
    
    res.json({
        success: true,
        data: paginatedLetters,
        meta: {
            total: letters.length,
            limit: limitNum,
            sort_by,
            sort_order,
            style: style || 'all'
        }
    });
});

// 写信相关API
writeApp.post('/letters', (req, res) => {
    res.json({
        success: true,
        data: {
            letter_id: 'letter-' + Date.now(),
            code: 'LP' + Math.random().toString(36).substr(2, 8).toUpperCase(),
            status: 'pending'
        },
        message: '信件创建成功，正在等待信使分配'
    });
});

writeApp.get('/letters/my', (req, res) => {
    res.json({
        success: true,
        data: {
            items: [
                {
                    id: 'letter-1',
                    code: 'LP12345678',
                    title: '给室友的感谢信',
                    recipient: '李小红',
                    status: 'delivered',
                    created_at: '2025-01-20T10:00:00Z',
                    delivered_at: '2025-01-21T14:30:00Z'
                }
            ],
            total: 1
        }
    });
});

writeApp.listen(8001, () => {
    console.log('📝 Write Service running on port 8001');
});

// 信使服务 (8002)
const courierApp = express();
courierApp.use(cors());
courierApp.use(express.json());

courierApp.get('/courier/info', (req, res) => {
    res.json({
        success: true,
        data: {
            id: 'courier-1',
            level: 2,
            region: '北京大学',
            total_points: 1250,
            completed_tasks: 35,
            success_rate: 96.5,
            current_tasks: 5,
            rank: 15
        }
    });
});

courierApp.get('/courier/tasks', (req, res) => {
    res.json({
        success: true,
        data: {
            items: [
                {
                    id: 'task-1',
                    letter_code: 'LP12345678',
                    priority: 'high',
                    pickup_address: '北京大学宿舍1号楼',
                    delivery_address: '北京大学宿舍5号楼',
                    status: 'assigned',
                    estimated_time: 30,
                    assigned_at: '2025-01-22T09:00:00Z'
                }
            ],
            total: 1
        }
    });
});

courierApp.post('/courier/tasks/:id/accept', (req, res) => {
    res.json({
        success: true,
        message: '任务已接受'
    });
});

courierApp.listen(8002, () => {
    console.log('🏃‍♂️ Courier Service running on port 8002');
});

// 管理服务 (8003)
const adminApp = express();
adminApp.use(cors());
adminApp.use(express.json());

adminApp.get('/api/admin/dashboard', (req, res) => {
    res.json({
        success: true,
        data: {
            users: { total: 1250, active: 1100, new_today: 25 },
            letters: { total: 5500, sent_today: 180, delivered_today: 165 },
            couriers: { total: 45, active: 38, busy: 25 },
            schools: { total: 125, active: 120 }
        }
    });
});

adminApp.get('/api/admin/statistics', (req, res) => {
    res.json({
        success: true,
        data: {
            daily_letters: [120, 135, 150, 180, 165, 190, 210],
            delivery_rate: [95.2, 96.1, 94.8, 96.5, 97.2, 95.9, 96.8],
            courier_performance: [85, 90, 88, 92, 89, 94, 91]
        }
    });
});

adminApp.listen(8003, () => {
    console.log('👨‍💼 Admin Service running on port 8003');
});

// OCR服务 (8004)
const ocrApp = express();
ocrApp.use(cors());
ocrApp.use(express.json());

ocrApp.post('/ocr/process', (req, res) => {
    res.json({
        success: true,
        data: {
            text: '这是OCR识别的文字内容：亲爱的朋友，感谢你的来信...',
            confidence: 0.95,
            language: 'zh-CN',
            processing_time: 1.2
        }
    });
});

ocrApp.post('/ocr/scan', (req, res) => {
    res.json({
        success: true,
        data: {
            scan_id: 'scan-' + Date.now(),
            letter_code: 'LP' + Math.random().toString(36).substr(2, 8).toUpperCase(),
            status: 'recognized'
        }
    });
});

ocrApp.listen(8004, () => {
    console.log('🔍 OCR Service running on port 8004');
});

console.log('✅ All OpenPenPal mock services started successfully!');
console.log('📊 Service URLs:');
console.log('   📝 Write Service: http://localhost:8001');
console.log('   🏃‍♂️ Courier Service: http://localhost:8002');
console.log('   👨‍💼 Admin Service: http://localhost:8003');
console.log('   🔍 OCR Service: http://localhost:8004');