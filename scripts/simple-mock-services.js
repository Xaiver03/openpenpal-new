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
        'courier1': { password: 'courier123', role: 'courier', school: 'PKU', courier_level: 1, courier_info: { level: 1, permissions: ['PKA1**'] } },
        'courier2': { password: 'courier123', role: 'courier', school: 'PKU', courier_level: 2, courier_info: { level: 2, permissions: ['PKA*'] } },
        'courier3': { password: 'courier123', role: 'courier', school: 'PKU', courier_level: 3, courier_info: { level: 3, permissions: ['PK*'] } },
        'courier4': { password: 'courier123', role: 'courier', school: 'PKU', courier_level: 4, courier_info: { level: 4, permissions: ['**'] } }
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
            permissions: user.role === 'admin' ? ['ALL'] : ['read', 'write'],
            courier_level: user.courier_level || null,
            courierLevel: user.courier_level || null,
            courier_info: user.courier_info || null,
            courierInfo: user.courier_info || null
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

// v1 API 版本的认证路由 - 与网关路径匹配
writeApp.post('/api/v1/auth/login', (req, res) => {
    const { username, password } = req.body;
    
    // 简化的用户验证
    const users = {
        'alice': { password: 'secret', role: 'student', school: 'PKU' },
        'bob': { password: 'password123', role: 'student', school: 'THU' },
        'admin': { password: 'admin123', role: 'admin', school: 'ADMIN' },
        'courier1': { password: 'courier123', role: 'courier', school: 'PKU', courier_level: 1, courier_info: { level: 1, permissions: ['PKA1**'] } },
        'courier2': { password: 'courier123', role: 'courier', school: 'PKU', courier_level: 2, courier_info: { level: 2, permissions: ['PKA*'] } },
        'courier3': { password: 'courier123', role: 'courier', school: 'PKU', courier_level: 3, courier_info: { level: 3, permissions: ['PK*'] } },
        'courier4': { password: 'courier123', role: 'courier', school: 'PKU', courier_level: 4, courier_info: { level: 4, permissions: ['**'] } }
    };
    
    const user = users[username];
    if (!user || user.password !== password) {
        return res.status(401).json(errorResponse('用户名或密码错误', 401));
    }
    
    res.json(successResponse({
        success: true,
        data: {
            token: `mock-jwt-token-${username}-${Date.now()}`,
            user: {
                id: `user_${username}`,
                username,
                email: `${username}@example.com`,
                nickname: username === 'alice' ? '爱丽丝' : username === 'bob' ? '鲍勃' : username,
                role: user.role,
                school_code: user.school,
                school_name: user.school === 'PKU' ? '北京大学' : user.school === 'THU' ? '清华大学' : '系统管理',
                permissions: user.role === 'admin' ? ['ALL'] : ['read', 'write'],
                courier_level: user.courier_level || null,
                courierLevel: user.courier_level || null,
                courier_info: user.courier_info || null,
                courierInfo: user.courier_info || null
            }
        }
    }, '登录成功'));
});

writeApp.post('/api/v1/auth/register', (req, res) => {
    res.json(successResponse({ 
        success: true,
        data: { 
            id: 'new_user_' + Date.now() 
        } 
    }, '注册成功'));
});

writeApp.get('/api/v1/auth/me', (req, res) => {
    res.json(successResponse({
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

// 公开信件接口 - 用于广场页面
writeApp.get('/letters/public', (req, res) => {
    const { limit = 20, sort_by = 'created_at', sort_order = 'desc', style } = req.query;
    
    const mockPublicLetters = [
        {
            id: 'plaza-letter-1',
            title: '写给三年后的自己',
            content: '亲爱的未来的我，当你读到这封信的时候，希望你已经成为了更好的自己。还记得现在的我吗？那个在图书馆里挥汗如雨的学生，那个为了一道数学题而熬夜到凌晨的少年。我知道路还很长，但我相信，只要坚持下去，总会到达想要的地方...',
            user: { nickname: '匿名作者', avatar: '/images/user-001.png' },
            style: 'future',
            created_at: '2024-01-20T10:00:00Z',
            is_public: true
        },
        {
            id: 'plaza-letter-2',
            title: '致正在迷茫的你',
            content: '如果你正在经历人生的低谷，请记住这只是暂时的。每个人都会有迷茫的时候，这是成长路上的必经之路。不要害怕迷茫，因为只有经历过黑暗，我们才能更珍惜光明。愿你在迷雾中找到前进的方向，愿你的心永远充满希望...',
            user: { nickname: '温暖使者', avatar: '/images/user-002.png' },
            style: 'warm',
            created_at: '2024-01-19T14:30:00Z',
            is_public: true
        },
        {
            id: 'plaza-letter-3',
            title: '一个关于友谊的故事',
            content: '我想和你分享一个关于友谊的故事，这个故事改变了我对友情的理解。那是一个秋天的下午，我坐在宿舍里感到孤独，突然收到了一个陌生人的来信...',
            user: { nickname: '故事讲述者', avatar: '/images/user-003.png' },
            style: 'story',
            created_at: '2024-01-18T09:15:00Z',
            is_public: true
        },
        {
            id: 'plaza-letter-4',
            title: '漂流到远方的思念',
            content: '这封信将随风漂流到某个角落，希望能遇到同样思念远方的你。也许我们从未谋面，但在这个瞬间，我们的心是相通的...',
            user: { nickname: '漂流者', avatar: '/images/user-004.png' },
            style: 'drift',
            created_at: '2024-01-17T16:45:00Z',
            is_public: true
        }
    ];
    
    // 根据样式过滤
    let filteredLetters = mockPublicLetters;
    if (style && style !== 'all') {
        filteredLetters = mockPublicLetters.filter(letter => letter.style === style);
    }
    
    // 排序
    filteredLetters.sort((a, b) => {
        if (sort_order === 'desc') {
            return new Date(b.created_at) - new Date(a.created_at);
        } else {
            return new Date(a.created_at) - new Date(b.created_at);
        }
    });
    
    // 限制数量
    const limitNum = parseInt(limit, 10) || 20;
    const limitedLetters = filteredLetters.slice(0, limitNum);
    
    res.json(successResponse({
        data: limitedLetters,
        total: filteredLetters.length,
        limit: limitNum
    }, '获取公开信件成功'));
});

// v1 API 版本的路由 - 与网关路径匹配
writeApp.get('/api/v1/letters/public', (req, res) => {
    const { limit = 20, sort_by = 'created_at', sort_order = 'desc', style } = req.query;
    
    const mockPublicLetters = [
        {
            id: 'plaza-letter-1',
            title: '写给三年后的自己',
            content: '亲爱的未来的我，当你读到这封信的时候，希望你已经成为了更好的自己。还记得现在的我吗？那个在图书馆里挥汗如雨的学生，那个为了一道数学题而熬夜到凌晨的少年。我知道路还很长，但我相信，只要坚持下去，总会到达想要的地方...',
            user: { nickname: '匿名作者', avatar: '/images/user-001.png' },
            style: 'future',
            created_at: '2024-01-20T10:00:00Z',
            is_public: true
        },
        {
            id: 'plaza-letter-2',
            title: '致正在迷茫的你',
            content: '如果你正在经历人生的低谷，请记住这只是暂时的。每个人都会有迷茫的时候，这是成长路上的必经之路。不要害怕迷茫，因为只有经历过黑暗，我们才能更珍惜光明。愿你在迷雾中找到前进的方向，愿你的心永远充满希望...',
            user: { nickname: '温暖使者', avatar: '/images/user-002.png' },
            style: 'warm',
            created_at: '2024-01-19T14:30:00Z',
            is_public: true
        },
        {
            id: 'plaza-letter-3',
            title: '一个关于友谊的故事',
            content: '我想和你分享一个关于友谊的故事，这个故事改变了我对友情的理解。那是一个秋天的下午，我坐在宿舍里感到孤独，突然收到了一个陌生人的来信...',
            user: { nickname: '故事讲述者', avatar: '/images/user-003.png' },
            style: 'story',
            created_at: '2024-01-18T09:15:00Z',
            is_public: true
        },
        {
            id: 'plaza-letter-4',
            title: '漂流到远方的思念',
            content: '这封信将随风漂流到某个角落，希望能遇到同样思念远方的你。也许我们从未谋面，但在这个瞬间，我们的心是相通的...',
            user: { nickname: '漂流者', avatar: '/images/user-004.png' },
            style: 'drift',
            created_at: '2024-01-17T16:45:00Z',
            is_public: true
        }
    ];
    
    // 根据样式过滤
    let filteredLetters = mockPublicLetters;
    if (style && style !== 'all') {
        filteredLetters = mockPublicLetters.filter(letter => letter.style === style);
    }
    
    // 排序
    filteredLetters.sort((a, b) => {
        if (sort_order === 'desc') {
            return new Date(b.created_at) - new Date(a.created_at);
        } else {
            return new Date(a.created_at) - new Date(b.created_at);
        }
    });
    
    // 限制数量
    const limitNum = parseInt(limit, 10) || 20;
    const limitedLetters = filteredLetters.slice(0, limitNum);
    
    res.json(successResponse({
        data: limitedLetters,
        total: filteredLetters.length,
        limit: limitNum
    }, '获取公开信件成功'));
});

// 健康检查
// Postcode 编码系统相关接口
// 先定义具体路径的路由，再定义通配符路由

// 3. 学校管理接口
writeApp.get('/api/v1/postcode/schools', (req, res) => {
    const mockSchools = [
        {
            id: 'school-pk-001',
            code: 'PK',
            name: '北京大学',
            full_name: '北京大学',
            status: 'active',
            managed_by: 'courier_level4_001',
            created_at: '2024-01-01T00:00:00Z',
            updated_at: '2024-01-01T00:00:00Z'
        },
        {
            id: 'school-th-001',
            code: 'TH',
            name: '清华大学',
            full_name: '清华大学',
            status: 'active',
            managed_by: 'courier_level4_002',
            created_at: '2024-01-01T00:00:00Z',
            updated_at: '2024-01-01T00:00:00Z'
        }
    ];
    
    res.json(successResponse({
        items: mockSchools,
        total: mockSchools.length,
        page: 0,
        limit: 20
    }, '获取学校列表成功'));
});

writeApp.post('/api/v1/postcode/schools', (req, res) => {
    const { code, name, full_name } = req.body;
    
    const newSchool = {
        id: 'school-' + Date.now(),
        code,
        name,
        full_name,
        status: 'active',
        managed_by: 'current_courier',
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString()
    };
    
    res.json(successResponse(newSchool, '学校创建成功'));
});

// 4. 片区管理接口
writeApp.get('/api/v1/postcode/schools/:schoolCode/areas', (req, res) => {
    const { schoolCode } = req.params;
    
    const mockAreas = [
        {
            id: 'area-' + schoolCode + '-a',
            school_code: schoolCode,
            code: 'A',
            name: '东区',
            description: '东区片区',
            status: 'active',
            managed_by: 'courier_level3_001',
            created_at: '2024-01-01T00:00:00Z',
            updated_at: '2024-01-01T00:00:00Z'
        },
        {
            id: 'area-' + schoolCode + '-b',
            school_code: schoolCode,
            code: 'B',
            name: '西区',
            description: '西区片区',
            status: 'active',
            managed_by: 'courier_level3_002',
            created_at: '2024-01-01T00:00:00Z',
            updated_at: '2024-01-01T00:00:00Z'
        }
    ];
    
    res.json(successResponse({
        items: mockAreas,
        total: mockAreas.length
    }, '获取片区列表成功'));
});

writeApp.post('/api/v1/postcode/schools/:schoolCode/areas', (req, res) => {
    const { schoolCode } = req.params;
    const { code, name, description } = req.body;
    
    const newArea = {
        id: 'area-' + Date.now(),
        school_code: schoolCode,
        code,
        name,
        description,
        status: 'active',
        managed_by: 'current_courier',
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString()
    };
    
    res.json(successResponse(newArea, '片区创建成功'));
});

// 5. 楼栋管理接口
writeApp.get('/api/v1/postcode/schools/:schoolCode/areas/:areaCode/buildings', (req, res) => {
    const { schoolCode, areaCode } = req.params;
    
    const mockBuildings = [
        {
            id: 'building-' + schoolCode + areaCode + '1',
            school_code: schoolCode,
            area_code: areaCode,
            code: '1',
            name: '1栋',
            type: 'dormitory',
            floors: 6,
            status: 'active',
            managed_by: 'courier_level2_001',
            created_at: '2024-01-01T00:00:00Z',
            updated_at: '2024-01-01T00:00:00Z'
        },
        {
            id: 'building-' + schoolCode + areaCode + '2',
            school_code: schoolCode,
            area_code: areaCode,
            code: '2',
            name: '2栋',
            type: 'dormitory',
            floors: 8,
            status: 'active',
            managed_by: 'courier_level2_002',
            created_at: '2024-01-01T00:00:00Z',
            updated_at: '2024-01-01T00:00:00Z'
        }
    ];
    
    res.json(successResponse({
        items: mockBuildings,
        total: mockBuildings.length
    }, '获取楼栋列表成功'));
});

writeApp.post('/api/v1/postcode/schools/:schoolCode/areas/:areaCode/buildings', (req, res) => {
    const { schoolCode, areaCode } = req.params;
    const { code, name, type, floors } = req.body;
    
    const newBuilding = {
        id: 'building-' + Date.now(),
        school_code: schoolCode,
        area_code: areaCode,
        code,
        name,
        type,
        floors,
        status: 'active',
        managed_by: 'current_courier',
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString()
    };
    
    res.json(successResponse(newBuilding, '楼栋创建成功'));
});

// 6. 房间管理接口
writeApp.get('/api/v1/postcode/schools/:schoolCode/areas/:areaCode/buildings/:buildingCode/rooms', (req, res) => {
    const { schoolCode, areaCode, buildingCode } = req.params;
    
    const mockRooms = [];
    for (let i = 1; i <= 20; i++) {
        const roomCode = i.toString().padStart(2, '0');
        const postcode = `${schoolCode}${areaCode}${buildingCode}${roomCode}`;
        
        mockRooms.push({
            id: 'room-' + postcode,
            school_code: schoolCode,
            area_code: areaCode,
            building_code: buildingCode,
            code: roomCode,
            name: `${buildingCode}${roomCode}`,
            type: 'dormitory',
            capacity: 4,
            floor: Math.ceil(i / 10),
            full_postcode: postcode,
            status: 'active',
            managed_by: 'courier_level1_001',
            created_at: '2024-01-01T00:00:00Z',
            updated_at: '2024-01-01T00:00:00Z'
        });
    }
    
    res.json(successResponse({
        items: mockRooms,
        total: mockRooms.length
    }, '获取房间列表成功'));
});

writeApp.post('/api/v1/postcode/schools/:schoolCode/areas/:areaCode/buildings/:buildingCode/rooms', (req, res) => {
    const { schoolCode, areaCode, buildingCode } = req.params;
    const { code, name, type, capacity, floor } = req.body;
    
    const postcode = `${schoolCode}${areaCode}${buildingCode}${code}`;
    
    const newRoom = {
        id: 'room-' + Date.now(),
        school_code: schoolCode,
        area_code: areaCode,
        building_code: buildingCode,
        code,
        name,
        type,
        capacity,
        floor,
        full_postcode: postcode,
        status: 'active',
        managed_by: 'current_courier',
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString()
    };
    
    res.json(successResponse(newRoom, '房间创建成功'));
});

// 7. 权限管理接口
writeApp.get('/api/v1/postcode/permissions/:courierId', (req, res) => {
    const { courierId } = req.params;
    
    const mockPermission = {
        id: 'permission-' + courierId,
        courier_id: courierId,
        level: 4,
        prefix_patterns: ['PK**', 'TH**'],
        can_manage: true,
        can_create: true,
        can_review: true,
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z'
    };
    
    res.json(successResponse(mockPermission, '获取权限成功'));
});

// 8. 反馈管理接口
writeApp.get('/api/v1/postcode/feedbacks', (req, res) => {
    const mockFeedbacks = [
        {
            id: 'feedback-001',
            type: 'new_address',
            postcode: 'PKA301',
            description: '新增宿舍楼3栋301室',
            suggested_school_code: 'PK',
            suggested_area_code: 'A',
            suggested_building_code: '3',
            suggested_room_code: '01',
            suggested_name: '3栋301室',
            submitted_by: 'user_001',
            submitter_type: 'user',
            status: 'pending',
            created_at: '2024-01-20T10:00:00Z',
            updated_at: '2024-01-20T10:00:00Z'
        }
    ];
    
    res.json(successResponse({
        items: mockFeedbacks,
        total: mockFeedbacks.length,
        page: 0,
        limit: 20
    }, '获取反馈列表成功'));
});

writeApp.post('/api/v1/postcode/feedbacks', (req, res) => {
    const { type, postcode, description, suggested_school_code, suggested_area_code, suggested_building_code, suggested_room_code, suggested_name } = req.body;
    
    const newFeedback = {
        id: 'feedback-' + Date.now(),
        type,
        postcode,
        description,
        suggested_school_code,
        suggested_area_code,
        suggested_building_code,
        suggested_room_code,
        suggested_name,
        submitted_by: 'current_user',
        submitter_type: 'user',
        status: 'pending',
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString()
    };
    
    res.json(successResponse(newFeedback, '反馈提交成功'));
});

// 9. 统计接口
writeApp.get('/api/v1/postcode/stats/popular', (req, res) => {
    const { limit = 10 } = req.query;
    
    const mockStats = [
        {
            postcode: 'PKA101',
            delivery_count: 150,
            error_count: 2,
            popularity_score: 95.5,
            last_used: '2024-01-20T15:30:00Z'
        },
        {
            postcode: 'PKA102',
            delivery_count: 132,
            error_count: 1,
            popularity_score: 92.3,
            last_used: '2024-01-20T14:20:00Z'
        },
        {
            postcode: 'THB201',
            delivery_count: 89,
            error_count: 0,
            popularity_score: 88.7,
            last_used: '2024-01-20T13:15:00Z'
        }
    ].slice(0, parseInt(limit));
    
    res.json(successResponse({
        items: mockStats,
        total: mockStats.length
    }, '获取热门地址统计成功'));
});

// 10. 工具接口
writeApp.post('/api/v1/postcode/validate', (req, res) => {
    const { codes } = req.body;
    
    if (!Array.isArray(codes)) {
        return res.status(400).json(errorResponse('codes必须为数组', 400));
    }
    
    const results = codes.map(code => ({
        code,
        is_valid: code.length === 6 && /^[A-Z0-9]{6}$/.test(code),
        exists: ['PKA101', 'PKA102', 'THB201'].includes(code),
        errors: code.length !== 6 ? ['编码长度必须为6位'] : []
    }));
    
    const valid = results.filter(r => r.is_valid).length;
    const invalid = results.length - valid;
    
    res.json(successResponse({
        total: results.length,
        valid,
        invalid,
        results
    }, '批量验证完成'));
});

// 1. Postcode查询接口 - 放在最后，避免与具体路径冲突
writeApp.get('/api/v1/postcode/:code', (req, res) => {
    const { code } = req.params;
    
    // 模拟解析Postcode
    if (code.length !== 6) {
        return res.status(400).json(errorResponse('Postcode必须为6位', 400));
    }
    
    const school = code.substring(0, 2);
    const area = code.substring(2, 3);
    const building = code.substring(3, 4);
    const room = code.substring(4, 6);
    
    // 模拟数据库查询结果
    const mockResult = {
        postcode: code,
        exists: true,
        hierarchy: {
            school: {
                code: school,
                name: school === 'PK' ? '北京大学' : school === 'TH' ? '清华大学' : '示例大学',
                full_name: school === 'PK' ? '北京大学' : school === 'TH' ? '清华大学' : '示例大学完整名称'
            },
            area: {
                code: area,
                name: area === 'A' ? '东区' : area === 'B' ? '西区' : area === 'C' ? '南区' : '北区',
                description: '校园片区'
            },
            building: {
                code: building,
                name: `${building}栋`,
                type: 'dormitory',
                floors: 6
            },
            room: {
                code: room,
                name: room,
                type: 'dormitory',
                capacity: 4,
                full_postcode: code
            }
        }
    };
    
    res.json(successResponse(mockResult, 'Postcode查询成功'));
});

// 2. 地址搜索接口
writeApp.get('/api/v1/address/search', (req, res) => {
    const { query, limit = 10 } = req.query;
    
    if (!query) {
        return res.status(400).json(errorResponse('搜索关键词不能为空', 400));
    }
    
    // 模拟搜索结果
    const mockResults = [
        {
            postcode: 'PKA101',
            fullAddress: '北京大学东区A栋101室',
            hierarchy: {
                school: { code: 'PK', name: '北京大学', full_name: '北京大学' },
                area: { code: 'A', name: '东区', description: '东区片区' },
                building: { code: '1', name: '1栋', type: 'dormitory', floors: 6 },
                room: { code: '01', name: '101', type: 'dormitory', capacity: 4, full_postcode: 'PKA101' }
            },
            matchScore: 0.95
        },
        {
            postcode: 'PKA102',
            fullAddress: '北京大学东区A栋102室',
            hierarchy: {
                school: { code: 'PK', name: '北京大学', full_name: '北京大学' },
                area: { code: 'A', name: '东区', description: '东区片区' },
                building: { code: '1', name: '1栋', type: 'dormitory', floors: 6 },
                room: { code: '02', name: '102', type: 'dormitory', capacity: 4, full_postcode: 'PKA102' }
            },
            matchScore: 0.90
        },
        {
            postcode: 'THB201',
            fullAddress: '清华大学西区B栋201室',
            hierarchy: {
                school: { code: 'TH', name: '清华大学', full_name: '清华大学' },
                area: { code: 'B', name: '西区', description: '西区片区' },
                building: { code: '2', name: '2栋', type: 'dormitory', floors: 8 },
                room: { code: '01', name: '201', type: 'dormitory', capacity: 2, full_postcode: 'THB201' }
            },
            matchScore: 0.85
        }
    ].filter(item => 
        item.fullAddress.toLowerCase().includes(query.toLowerCase()) ||
        item.postcode.toLowerCase().includes(query.toLowerCase())
    ).slice(0, parseInt(limit));
    
    res.json(successResponse({
        results: mockResults,
        total: mockResults.length,
        query: query
    }, '地址搜索成功'));
});

// 用户相关API补充
writeApp.get('/api/users/me', (req, res) => {
    const mockUserProfile = {
        id: 'user_123',
        username: 'testuser',
        nickname: '测试用户',
        email: 'test@example.com',
        role: 'student',
        school_code: 'PKU',
        school_name: '北京大学',
        avatar: '/images/avatar/default.png',
        bio: '热爱写信和阅读的大学生',
        address: '北京大学东区A栋101',
        created_at: '2024-01-01T00:00:00Z'
    };
    res.json(successResponse(mockUserProfile, '获取用户信息成功'));
});

writeApp.put('/api/users/me', (req, res) => {
    const { nickname, avatar, bio, address } = req.body;
    
    const updatedProfile = {
        nickname: nickname || '测试用户',
        avatar: avatar || '/images/avatar/default.png',
        bio: bio || '热爱写信和阅读的大学生',
        address: address || '北京大学东区A栋101',
        updated_at: new Date().toISOString()
    };
    
    res.json(successResponse(updatedProfile, '用户信息更新成功'));
});

writeApp.get('/api/users/me/stats', (req, res) => {
    const mockUserStats = {
        letters_sent: 12,
        letters_received: 8,
        letters_read: 25,
        museum_contributions: 3,
        total_points: 320,
        member_since: '2024-01-01',
        favorite_styles: ['warm', 'story', 'future']
    };
    res.json(successResponse(mockUserStats, '获取用户统计成功'));
});

writeApp.post('/api/users/me/change-password', (req, res) => {
    const { old_password, new_password } = req.body;
    
    // 简单验证模拟
    if (!old_password || !new_password) {
        return res.status(400).json(errorResponse('密码不能为空', 400));
    }
    
    res.json(successResponse({}, '密码修改成功'));
});

writeApp.get('/api/letters/stats', (req, res) => {
    const mockLetterStats = {
        total_letters: 1247,
        public_letters: 156,
        monthly_letters: 89,
        popular_styles: [
            { style: 'warm', count: 45 },
            { style: 'future', count: 38 },
            { style: 'story', count: 32 },
            { style: 'drift', count: 28 }
        ],
        recent_activity: {
            last_24h: 12,
            last_week: 67,
            last_month: 234
        }
    };
    res.json(successResponse(mockLetterStats, '获取信件统计成功'));
});

writeApp.get('/api/letters/read/:code', (req, res) => {
    const { code } = req.params;
    
    const mockLetter = {
        id: 'letter_' + code,
        code: code,
        title: '一封来自远方的信',
        content: '你好，这是一封测试信件的内容。希望你能喜欢这个简单而温暖的问候。',
        style: 'warm',
        sender_nickname: '匿名作者',
        created_at: '2024-01-20T14:30:00Z',
        read_count: 23,
        like_count: 8,
        is_public: false
    };
    
    res.json(successResponse(mockLetter, '信件读取成功'));
});

writeApp.post('/api/letters/read/:code/mark-read', (req, res) => {
    const { code } = req.params;
    
    res.json(successResponse({
        letter_code: code,
        marked_at: new Date().toISOString(),
        read_count: 24
    }, '信件标记已读成功'));
});

writeApp.post('/api/letters/:letterId/generate-code', (req, res) => {
    const { letterId } = req.params;
    
    const letterCode = 'LC' + Date.now().toString().slice(-8);
    
    res.json(successResponse({
        letter_code: letterCode,
        qr_code_url: `/api/qr/${letterCode}`,
        read_url: `/read/${letterCode}`
    }, '信件编码生成成功'));
});

// 博物馆API补充
writeApp.post('/api/museum/contribute', (req, res) => {
    // 处理FormData上传
    const contributionId = 'contribution_' + Date.now();
    
    res.json(successResponse({
        id: contributionId,
        status: 'pending_review',
        submitted_at: new Date().toISOString(),
        estimated_review_time: '3-5工作日'
    }, '博物馆贡献提交成功'));
});

writeApp.post('/api/museum/contribute/letter', (req, res) => {
    const { letter_id } = req.body;
    
    if (!letter_id) {
        return res.status(400).json(errorResponse('信件ID不能为空', 400));
    }
    
    res.json(successResponse({
        id: 'contribution_' + Date.now(),
        letter_id,
        status: 'pending_review',
        submitted_at: new Date().toISOString()
    }, '信件贡献提交成功'));
});

// 信使相关API
writeApp.post('/api/courier/letters/:code/status', (req, res) => {
    const { code } = req.params;
    const { status, location, note } = req.body;
    
    res.json(successResponse({
        letter_code: code,
        new_status: status,
        location,
        note,
        updated_at: new Date().toISOString(),
        updated_by: 'courier_123'
    }, '信件状态更新成功'));
});

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

// 信使管理API - 补充缺失的接口
// 获取信使个人信息
courierApp.get('/api/courier/me', (req, res) => {
    const mockCourierInfo = {
        id: 'courier_123',
        level: 2,
        region: '东区',
        school: '北京大学',
        zone: 'SCHOOL_ZONE_A',
        total_points: 580,
        completed_tasks: 225,
        parent_id: 'courier_456'
    };
    res.json(successResponse(mockCourierInfo, '获取信使信息成功'));
});

// 获取信使统计信息
courierApp.get('/api/courier/stats', (req, res) => {
    const mockStats = {
        total_zones: 8,
        active_couriers: 12,
        total_tasks: 234,
        completed_tasks: 225,
        total_points: 580,
        level_progress: {
            current_level: 2,
            next_level: 3,
            progress_percentage: 75.5,
            points_needed: 420
        }
    };
    res.json(successResponse(mockStats, '获取统计信息成功'));
});

// 获取城市级统计 (四级信使专用)
courierApp.get('/api/courier/stats/city', (req, res) => {
    const mockCityStats = {
        total_schools: 25,
        active_couriers: 38,
        total_deliveries: 12678,
        pending_tasks: 45,
        average_rating: 4.9,
        success_rate: 97.2
    };
    res.json(successResponse(mockCityStats, '获取城市统计成功'));
});

// 获取学校级统计 (三级信使专用)  
courierApp.get('/api/courier/stats/school', (req, res) => {
    const mockSchoolStats = {
        total_zones: 8,
        active_couriers: 12,
        total_deliveries: 2456,
        pending_tasks: 15,
        average_rating: 4.8,
        coverage_rate: 94.5
    };
    res.json(successResponse(mockSchoolStats, '获取学校统计成功'));
});

// 获取片区级统计 (二级信使专用)
courierApp.get('/api/courier/stats/zone', (req, res) => {
    const mockZoneStats = {
        total_buildings: 12,
        active_couriers: 18,
        total_deliveries: 892,
        pending_tasks: 5,
        average_rating: 4.7,
        success_rate: 96.3
    };
    res.json(successResponse(mockZoneStats, '获取片区统计成功'));
});

// 获取一级信使统计信息
courierApp.get('/api/courier/first-level/stats', (req, res) => {
    const mockFirstLevelStats = {
        totalBuildings: 12,
        activeCouriers: 15,
        totalDeliveries: 456,
        pendingTasks: 8,
        averageRating: 4.8,
        completionRate: 95.6
    };
    res.json(successResponse(mockFirstLevelStats, '获取一级信使统计成功'));
});

// 获取下级信使列表
courierApp.get('/api/courier/subordinates', (req, res) => {
    const mockSubordinates = [
        {
            id: 'sub_001',
            name: 'building_a_courier',
            level: 1,
            region: 'A栋',
            school: '北京大学',
            zone: 'ZONE_A_001',
            total_points: 320,
            completed_tasks: 148,
            status: 'active'
        },
        {
            id: 'sub_002',
            name: 'building_b_courier',
            level: 1,
            region: 'B栋',
            school: '北京大学',
            zone: 'ZONE_A_002',
            total_points: 280,
            completed_tasks: 129,
            status: 'active'
        }
    ];
    res.json(successResponse({ couriers: mockSubordinates }, '获取下级信使列表成功'));
});

// 获取一级信使列表
courierApp.get('/api/courier/first-level/couriers', (req, res) => {
    const mockFirstLevelCouriers = [
        {
            id: 'courier_001',
            username: 'building_a_courier',
            buildingName: 'A栋',
            buildingCode: 'ZONE_A_001',
            floorRange: '1-6层',
            roomRange: '101-620',
            level: 1,
            status: 'active',
            points: 320,
            taskCount: 156,
            completedTasks: 148,
            averageRating: 4.9,
            joinDate: '2024-01-10',
            lastActive: '2024-01-24T09:15:00Z',
            contactInfo: {
                phone: '138****1234',
                wechat: 'building_a_courier'
            },
            workingHours: {
                start: '08:00',
                end: '18:00',
                weekdays: [1, 2, 3, 4, 5, 6]
            }
        },
        {
            id: 'courier_002',
            username: 'building_b_courier',
            buildingName: 'B栋',
            buildingCode: 'ZONE_A_002',
            floorRange: '1-8层',
            roomRange: '101-825',
            level: 1,
            status: 'active',
            points: 280,
            taskCount: 134,
            completedTasks: 129,
            averageRating: 4.6,
            joinDate: '2024-01-15',
            lastActive: '2024-01-24T11:30:00Z',
            contactInfo: {
                phone: '159****5678'
            },
            workingHours: {
                start: '09:00',
                end: '19:00',
                weekdays: [1, 2, 3, 4, 5]
            }
        }
    ];
    res.json(successResponse(mockFirstLevelCouriers, '获取一级信使列表成功'));
});

// 获取积分排行榜
courierApp.get('/api/courier/leaderboard/:scope', (req, res) => {
    const { scope } = req.params;
    const mockLeaderboard = [
        {
            id: 'leader_001',
            name: 'university_peking_manager',
            level: 3,
            total_points: 1250,
            rank: 1,
            school: '北京大学',
            zone: '全校'
        },
        {
            id: 'leader_002',
            name: 'university_tsinghua_manager',
            level: 3,
            total_points: 1180,
            rank: 2,
            school: '清华大学',
            zone: '全校'
        },
        {
            id: 'leader_003',
            name: 'zone_a_manager',
            level: 2,
            total_points: 580,
            rank: 3,
            school: '北京大学',
            zone: '东区'
        }
    ];
    res.json(successResponse({ leaderboard: mockLeaderboard }, `获取${scope}排行榜成功`));
});

// 获取积分历史
courierApp.get('/api/courier/points-history', (req, res) => {
    const mockHistory = [
        {
            id: 'history_001',
            points: 50,
            action: '完成配送任务',
            created_at: '2024-01-24T10:30:00Z',
            task_id: 'task_456'
        },
        {
            id: 'history_002',
            points: 20,
            action: '用户好评',
            created_at: '2024-01-24T09:15:00Z',
            task_id: 'task_455'
        },
        {
            id: 'history_003',
            points: 30,
            action: '准时送达',
            created_at: '2024-01-23T16:45:00Z',
            task_id: 'task_454'
        }
    ];
    res.json(successResponse({ history: mockHistory }, '获取积分历史成功'));
});

// 接受信使任务
courierApp.put('/api/courier/tasks/:taskId/accept', (req, res) => {
    const { taskId } = req.params;
    const { estimated_time, note } = req.body;
    
    res.json(successResponse({
        task_id: taskId,
        status: 'accepted',
        estimated_time,
        note,
        accepted_at: new Date().toISOString()
    }, '任务接受成功'));
});

// 高级信使管理API - 对应前端管理页面需求

// 获取城市级信使列表 (四级信使管理页面使用)
courierApp.get('/api/courier/city/couriers', (req, res) => {
    const mockCityCouriers = [
        {
            id: '1',
            username: 'university_peking_manager',
            schoolName: '北京大学',
            schoolCode: 'PKU_MAIN',
            zoneCount: 8,
            coverage: '燕园校区全区',
            level: 3,
            status: 'active',
            points: 1250,
            taskCount: 456,
            completedTasks: 445,
            subordinateCount: 8,
            averageRating: 4.95,
            joinDate: '2023-09-01',
            lastActive: '2024-01-24T07:45:00Z',
            contactInfo: {
                phone: '138****0001',
                wechat: 'pku_courier_head'
            },
            workingHours: {
                start: '06:00',
                end: '22:00',
                weekdays: [1, 2, 3, 4, 5, 6, 7]
            }
        },
        {
            id: '2',
            username: 'university_tsinghua_manager',
            schoolName: '清华大学',
            schoolCode: 'THU_MAIN',
            zoneCount: 6,
            coverage: '紫荆校区全区',
            level: 3,
            status: 'active',
            points: 1180,
            taskCount: 398,
            completedTasks: 390,
            subordinateCount: 6,
            averageRating: 4.88,
            joinDate: '2023-09-05',
            lastActive: '2024-01-24T08:20:00Z'
        }
    ];
    res.json(successResponse(mockCityCouriers, '获取城市级信使列表成功'));
});

// 获取学校级信使列表 (三级信使管理页面使用)
courierApp.get('/api/courier/school/couriers', (req, res) => {
    const mockSchoolCouriers = [
        {
            id: '1',
            username: 'zone_a_manager',
            zoneName: '东区',
            zoneCode: 'SCHOOL_ZONE_A',
            buildingCount: 6,
            coverageArea: '宿舍区A1-A6',
            level: 2,
            status: 'active',
            points: 580,
            taskCount: 234,
            completedTasks: 225,
            subordinateCount: 6,
            averageRating: 4.9,
            joinDate: '2023-12-01',
            lastActive: '2024-01-24T08:30:00Z'
        },
        {
            id: '2',
            username: 'zone_b_manager',
            zoneName: '西区',
            zoneCode: 'SCHOOL_ZONE_B',
            buildingCount: 4,
            coverageArea: '宿舍区B1-B4',
            level: 2,
            status: 'active',
            points: 520,
            taskCount: 198,
            completedTasks: 192,
            subordinateCount: 4,
            averageRating: 4.7,
            joinDate: '2024-01-05',
            lastActive: '2024-01-24T10:15:00Z'
        }
    ];
    res.json(successResponse(mockSchoolCouriers, '获取学校级信使列表成功'));
});

// 获取片区级信使列表 (二级信使管理页面使用)
courierApp.get('/api/courier/zone/couriers', (req, res) => {
    const mockZoneCouriers = [
        {
            id: '1',
            username: 'building_a_courier',
            buildingName: 'A栋',
            buildingCode: 'ZONE_A_001',
            floorRange: '1-6层',
            roomRange: '101-620',
            level: 1,
            status: 'active',
            points: 320,
            taskCount: 156,
            completedTasks: 148,
            averageRating: 4.9,
            joinDate: '2024-01-10',
            lastActive: '2024-01-24T09:15:00Z'
        },
        {
            id: '2',
            username: 'building_b_courier',
            buildingName: 'B栋',
            buildingCode: 'ZONE_A_002',
            floorRange: '1-8层',
            roomRange: '101-825',
            level: 1,
            status: 'active',
            points: 280,
            taskCount: 134,
            completedTasks: 129,
            averageRating: 4.6,
            joinDate: '2024-01-15',
            lastActive: '2024-01-24T11:30:00Z'
        }
    ];
    res.json(successResponse(mockZoneCouriers, '获取片区级信使列表成功'));
});

// 创建下级信使 API - 四级信使创建三级信使，三级创建二级，二级创建一级
courierApp.post('/api/courier/create', (req, res) => {
    const { username, email, level, region, school, zone, building } = req.body;
    
    // 模拟权限检查 - 根据请求头的token获取当前用户级别
    const authHeader = req.headers.authorization;
    if (!authHeader || !authHeader.startsWith('Bearer ')) {
        return res.status(401).json(errorResponse('需要认证Token', 401));
    }
    
    // 从token中解析当前用户信息（简化版）
    const token = authHeader.replace('Bearer ', '');
    let currentUserLevel = 4; // 默认给最高权限做测试
    
    // 根据token判断用户级别
    if (token.includes('courier1')) currentUserLevel = 1;
    else if (token.includes('courier2')) currentUserLevel = 2; 
    else if (token.includes('courier3')) currentUserLevel = 3;
    else if (token.includes('courier4')) currentUserLevel = 4;
    
    // 权限验证：只能创建比自己低一级的信使
    if (level >= currentUserLevel) {
        return res.status(403).json(errorResponse('无权限创建同级或更高级别的信使', 403));
    }
    
    // 级别验证
    if (level < 1 || level > 4) {
        return res.status(400).json(errorResponse('信使级别必须在1-4之间', 400));
    }
    
    // 创建新信使
    const newCourier = {
        id: `courier_${Date.now()}`,
        username,
        email,
        level,
        region: region || '默认区域',
        school: school || '默认学校',
        zone: zone || null,
        building: building || null,
        status: 'pending', // 新创建的信使状态为待审核
        total_points: 0,
        completed_tasks: 0,
        parent_id: `current_user_${currentUserLevel}`,
        created_at: new Date().toISOString(),
        created_by: `level_${currentUserLevel}_user`
    };
    
    res.json(successResponse(newCourier, '信使创建成功，等待审核'));
});

// 获取可创建的信使级别 - 用于前端显示可选级别
courierApp.get('/api/courier/creatable-levels', (req, res) => {
    const authHeader = req.headers.authorization;
    if (!authHeader || !authHeader.startsWith('Bearer ')) {
        return res.status(401).json(errorResponse('需要认证Token', 401));
    }
    
    const token = authHeader.replace('Bearer ', '');
    let currentUserLevel = 4;
    
    if (token.includes('courier1')) currentUserLevel = 1;
    else if (token.includes('courier2')) currentUserLevel = 2;
    else if (token.includes('courier3')) currentUserLevel = 3; 
    else if (token.includes('courier4')) currentUserLevel = 4;
    
    // 返回可创建的级别（比自己低一级）
    const creatableLevels = [];
    if (currentUserLevel > 1) {
        const targetLevel = currentUserLevel - 1;
        const levelNames = {
            1: '楼栋级信使',
            2: '片区级信使', 
            3: '学校级信使',
            4: '城市级信使'
        };
        
        creatableLevels.push({
            level: targetLevel,
            name: levelNames[targetLevel],
            description: `管理${targetLevel === 1 ? '楼栋' : targetLevel === 2 ? '片区' : targetLevel === 3 ? '学校' : '城市'}范围内的信件投递`
        });
    }
    
    res.json(successResponse({ 
        current_level: currentUserLevel,
        creatable_levels: creatableLevels,
        can_create: creatableLevels.length > 0
    }, '获取可创建级别成功'));
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

// 管理员API补充 - 用户管理相关
// 获取可任命角色列表
adminApp.get('/api/admin/appointable-roles', (req, res) => {
    const mockRoles = ['courier_level_1', 'courier_level_2', 'courier_level_3', 'courier_level_4', 'moderator'];
    res.json(successResponse({ roles: mockRoles }, '获取可任命角色成功'));
});

// 任命用户
adminApp.post('/api/admin/appoint', (req, res) => {
    const { user_id, new_role, reason } = req.body;
    
    const appointmentRecord = {
        id: 'appointment_' + Date.now(),
        user_id,
        old_role: 'student',
        new_role,
        reason,
        appointed_by: 'admin',
        appointed_at: new Date().toISOString(),
        status: 'approved'
    };
    
    res.json(successResponse(appointmentRecord, '用户任命成功'));
});

// 获取任命记录
adminApp.get('/api/admin/appointment-records', (req, res) => {
    const mockRecords = [
        {
            id: 'appointment_001',
            user_id: 'user_alice',
            old_role: 'student',
            new_role: 'courier_level_1',
            reason: '表现优秀，积极配送',
            appointed_by: 'admin',
            appointed_at: '2024-01-20T10:30:00Z',
            status: 'approved'
        },
        {
            id: 'appointment_002',
            user_id: 'user_bob',
            old_role: 'courier_level_1',
            new_role: 'courier_level_2',
            reason: '管理能力强，负责区域表现优秀',
            appointed_by: 'admin',
            appointed_at: '2024-01-22T14:15:00Z',
            status: 'approved'
        }
    ];
    
    res.json(successResponse({ 
        records: mockRecords, 
        total: mockRecords.length 
    }, '获取任命记录成功'));
});

// 获取信使候选用户
adminApp.get('/api/admin/courier-candidates', (req, res) => {
    const mockCandidates = [
        {
            id: 'candidate_001',
            username: 'student_zhang',
            nickname: '张同学',
            email: 'zhang@pku.edu.cn',
            role: 'student',
            school_code: 'PKU',
            created_at: '2024-01-10T00:00:00Z',
            status: 'active'
        },
        {
            id: 'candidate_002',
            username: 'student_li',
            nickname: '李同学',
            email: 'li@thu.edu.cn',
            role: 'student',
            school_code: 'THU',
            created_at: '2024-01-12T00:00:00Z',
            status: 'active'
        }
    ];
    
    res.json(successResponse({ candidates: mockCandidates }, '获取候选用户成功'));
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
gatewayApp.use('/api/v1/auth', proxy(8001));
gatewayApp.use('/api/v1/letters', proxy(8001));
gatewayApp.use('/api/v1/postcode', proxy(8001));  // Postcode路由
gatewayApp.use('/api/v1/address', proxy(8001));   // 地址搜索路由
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

// WebSocket支持
const http = require('http');
const server = http.createServer(gatewayApp);

// 简化版WebSocket处理 - 模拟WebSocket升级请求
server.on('upgrade', (request, socket, head) => {
    console.log('📡 WebSocket升级请求:', request.url);
    
    // 解析URL和参数
    const url = require('url');
    const parsedUrl = url.parse(request.url, true);
    const { token } = parsedUrl.query;
    
    // 简单的WebSocket握手响应
    const responseKey = 'mock-websocket-accept-key';
    const responseHeaders = [
        'HTTP/1.1 101 Switching Protocols',
        'Upgrade: websocket',
        'Connection: Upgrade',
        `Sec-WebSocket-Accept: ${responseKey}`,
        'Sec-WebSocket-Version: 13',
        '',
        ''
    ].join('\r\n');
    
    socket.write(responseHeaders);
    
    // 模拟WebSocket消息
    socket.on('data', (buffer) => {
        console.log('📨 收到WebSocket消息');
        // 发送ping消息保持连接
        socket.write(Buffer.from('{"type":"ping","timestamp":"' + new Date().toISOString() + '"}'));
    });
    
    // 定期发送心跳
    const heartbeat = setInterval(() => {
        if (!socket.destroyed) {
            socket.write(Buffer.from('{"type":"heartbeat","timestamp":"' + new Date().toISOString() + '"}'));
        } else {
            clearInterval(heartbeat);
        }
    }, 30000);
    
    socket.on('close', () => {
        console.log('📡 WebSocket连接关闭');
        clearInterval(heartbeat);
    });
    
    socket.on('error', (err) => {
        console.log('📡 WebSocket错误:', err.message);
        clearInterval(heartbeat);
    });
});

server.listen(8000, () => {
    console.log('✅ API网关已启动: http://localhost:8000');
    console.log('✅ WebSocket服务已启动: ws://localhost:8000');
    console.log('');
    console.log('🎉 OpenPenPal 简化版 Mock 服务全部启动完成！');
    console.log('');
    console.log('📋 服务列表:');
    console.log('   • API网关: http://localhost:8000');
    console.log('   • WebSocket: ws://localhost:8000/ws');
    console.log('   • 写信服务: http://localhost:8001');
    console.log('   • 信使服务: http://localhost:8002');
    console.log('   • 管理服务: http://localhost:8003');
    console.log('   • OCR服务: http://localhost:8004');
    console.log('');
    console.log('🔑 测试账号:');
    console.log('   • alice/secret - 学生用户');
    console.log('   • admin/admin123 - 管理员');
    console.log('   • courier1/courier123 - 信使 (1级)');
    console.log('   • courier2/courier123 - 信使 (2级)');
    console.log('   • courier3/courier123 - 信使 (3级)');
    console.log('   • courier4/courier123 - 信使 (4级,可创建下级)');
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