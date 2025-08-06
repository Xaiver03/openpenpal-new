#!/usr/bin/env node

/**
 * OpenPenPal API集成测试脚本
 * 测试前端API调用与后端Mock服务的对接情况
 */

// Use built-in fetch (Node.js 18+)

const API_BASE = 'http://localhost:8000'; // API网关地址
const COURIER_API_BASE = 'http://localhost:8002'; // 信使服务直接地址

// 测试用的认证token（模拟）
let authToken = null;

// 颜色输出
const colors = {
    green: '\x1b[32m',
    red: '\x1b[31m',
    yellow: '\x1b[33m',
    blue: '\x1b[34m',
    reset: '\x1b[0m'
};

function log(color, message) {
    console.log(`${colors[color]}${message}${colors.reset}`);
}

// API调用封装
async function apiCall(endpoint, options = {}) {
    const url = endpoint.startsWith('/courier') ? 
        `${COURIER_API_BASE}/api${endpoint}` : 
        `${API_BASE}/api${endpoint}`;
    
    const config = {
        headers: {
            'Content-Type': 'application/json',
            ...(authToken && { Authorization: `Bearer ${authToken}` }),
            ...options.headers,
        },
        ...options,
    };

    try {
        const response = await fetch(url, config);
        const result = await response.json();
        return { success: response.ok, status: response.status, data: result };
    } catch (error) {
        return { success: false, error: error.message };
    }
}

// 测试用例
const testCases = [
    {
        name: '认证系统测试',
        tests: [
            {
                name: '用户登录',
                endpoint: '/auth/login',
                method: 'POST',
                body: { username: 'courier1', password: 'courier123' },
                validate: (result) => result.data?.token ? 'token获取成功' : '登录失败'
            }
        ]
    },
    {
        name: '信使管理API测试',
        tests: [
            {
                name: '获取信使个人信息',
                endpoint: '/courier/me',
                method: 'GET',
                validate: (result) => result.data?.level ? `信使等级: ${result.data.level}` : '获取失败'
            },
            {
                name: '获取城市级统计',
                endpoint: '/courier/stats/city',
                method: 'GET',
                validate: (result) => result.data?.total_schools ? `管理学校数: ${result.data.total_schools}` : '获取失败'
            },
            {
                name: '获取学校级统计',
                endpoint: '/courier/stats/school',
                method: 'GET',
                validate: (result) => result.data?.total_zones ? `管理片区数: ${result.data.total_zones}` : '获取失败'
            },
            {
                name: '获取片区级统计',
                endpoint: '/courier/stats/zone',
                method: 'GET',
                validate: (result) => result.data?.total_buildings ? `管理楼栋数: ${result.data.total_buildings}` : '获取失败'
            },
            {
                name: '获取一级信使统计',
                endpoint: '/courier/first-level/stats',
                method: 'GET',
                validate: (result) => result.data?.totalBuildings ? `管理楼栋数: ${result.data.totalBuildings}` : '获取失败'
            },
            {
                name: '获取下级信使列表',
                endpoint: '/courier/subordinates',
                method: 'GET',
                validate: (result) => result.data?.couriers ? `下级信使数: ${result.data.couriers.length}` : '获取失败'
            },
            {
                name: '获取城市级信使列表',
                endpoint: '/courier/city/couriers',
                method: 'GET',
                validate: (result) => Array.isArray(result.data) ? `城市级信使数: ${result.data.length}` : '获取失败'
            },
            {
                name: '获取学校级信使列表',
                endpoint: '/courier/school/couriers',
                method: 'GET',
                validate: (result) => Array.isArray(result.data) ? `学校级信使数: ${result.data.length}` : '获取失败'
            },
            {
                name: '获取片区级信使列表',
                endpoint: '/courier/zone/couriers',
                method: 'GET',
                validate: (result) => Array.isArray(result.data) ? `片区级信使数: ${result.data.length}` : '获取失败'
            },
            {
                name: '获取一级信使列表',
                endpoint: '/courier/first-level/couriers',
                method: 'GET',
                validate: (result) => Array.isArray(result.data) ? `一级信使数: ${result.data.length}` : '获取失败'
            },
            {
                name: '获取积分排行榜',
                endpoint: '/courier/leaderboard/school',
                method: 'GET',
                validate: (result) => result.data?.leaderboard ? `排行榜条目数: ${result.data.leaderboard.length}` : '获取失败'
            }
        ]
    },
    {
        name: 'Postcode系统测试',
        tests: [
            {
                name: '获取学校列表',
                endpoint: '/v1/postcode/schools',
                method: 'GET',
                validate: (result) => result.data?.items ? `学校数量: ${result.data.items.length}` : '获取失败'
            },
            {
                name: 'Postcode查询',
                endpoint: '/v1/postcode/PKA101',
                method: 'GET',
                validate: (result) => result.data?.postcode ? `查询到: ${result.data.postcode}` : '查询失败'
            }
        ]
    },
    {
        name: '信件系统测试',
        tests: [
            {
                name: '获取公开信件',
                endpoint: '/v1/letters/public',
                method: 'GET',
                validate: (result) => result.data?.data ? `公开信件数: ${result.data.data.length}` : '获取失败'
            },
            {
                name: '获取信件统计',
                endpoint: '/letters/stats',
                method: 'GET',
                validate: (result) => result.data?.total_letters ? `总信件数: ${result.data.total_letters}` : '获取失败'
            }
        ]
    },
    {
        name: '用户系统测试',
        tests: [
            {
                name: '获取用户信息',
                endpoint: '/users/me',
                method: 'GET',
                validate: (result) => result.data?.username ? `用户: ${result.data.username}` : '获取失败'
            },
            {
                name: '获取用户统计',
                endpoint: '/users/me/stats',
                method: 'GET',
                validate: (result) => result.data?.letters_sent ? `发送信件数: ${result.data.letters_sent}` : '获取失败'
            }
        ]
    }
];

// 执行测试
async function runTests() {
    log('blue', '🚀 开始OpenPenPal API集成测试\n');
    
    let totalTests = 0;
    let passedTests = 0;
    let failedTests = 0;

    for (const category of testCases) {
        log('yellow', `📋 ${category.name}`);
        
        for (const test of category.tests) {
            totalTests++;
            const options = {
                method: test.method || 'GET',
                ...(test.body && { body: JSON.stringify(test.body) })
            };
            
            const result = await apiCall(test.endpoint, options);
            
            if (result.success) {
                const validation = test.validate ? test.validate(result) : '成功';
                log('green', `  ✅ ${test.name}: ${validation}`);
                passedTests++;
                
                // 如果是登录测试，保存token
                if (test.name === '用户登录' && result.data?.data?.token) {
                    authToken = result.data.data.token;
                }
            } else {
                log('red', `  ❌ ${test.name}: ${result.error || `HTTP ${result.status}`}`);
                failedTests++;
            }
        }
        console.log();
    }
    
    // 测试总结
    log('blue', '📊 测试总结:');
    log('green', `  通过: ${passedTests}/${totalTests}`);
    if (failedTests > 0) {
        log('red', `  失败: ${failedTests}/${totalTests}`);
    }
    
    const passRate = ((passedTests / totalTests) * 100).toFixed(1);
    log('yellow', `  通过率: ${passRate}%`);
    
    if (passRate >= 80) {
        log('green', '\n🎉 API集成测试基本通过！');
    } else if (passRate >= 60) {
        log('yellow', '\n⚠️  API集成测试部分通过，需要改进');
    } else {
        log('red', '\n❌ API集成测试失败较多，需要修复');
    }
}

// 检查服务状态
async function checkServices() {
    log('blue', '🔍 检查服务状态...\n');
    
    const services = [
        { name: 'API网关', url: `${API_BASE}/health` },
        { name: '写信服务', url: 'http://localhost:8001/health' },
        { name: '信使服务', url: 'http://localhost:8002/health' },
        { name: '管理服务', url: 'http://localhost:8003/health' }
    ];
    
    let allServicesUp = true;
    
    for (const service of services) {
        try {
            const response = await fetch(service.url);
            const result = await response.json();
            
            if (response.ok && result.status === 'healthy') {
                log('green', `  ✅ ${service.name}: 运行正常`);
            } else {
                log('red', `  ❌ ${service.name}: 状态异常`);
                allServicesUp = false;
            }
        } catch (error) {
            log('red', `  ❌ ${service.name}: 无法连接`);
            allServicesUp = false;
        }
    }
    
    console.log();
    
    if (!allServicesUp) {
        log('yellow', '⚠️  部分服务未启动，请先启动Mock服务:');
        log('blue', '   node scripts/simple-mock-services.js');
        console.log();
        return false;
    }
    
    return true;
}

// 主函数
async function main() {
    const servicesOk = await checkServices();
    
    if (servicesOk) {
        await runTests();
    }
}

// 错误处理
process.on('unhandledRejection', (error) => {
    log('red', `\n❌ 未处理的错误: ${error.message}`);
    process.exit(1);
});

// 执行
main().catch(console.error);