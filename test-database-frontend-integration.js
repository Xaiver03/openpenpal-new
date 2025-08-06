#!/usr/bin/env node

/**
 * 数据库和前端集成测试
 * 测试整个系统的端到端交互：数据库 ↔ 后端 ↔ 前端
 */

const http = require('http');
const { execSync } = require('child_process');

const BASE_URL = 'http://localhost:8080';
const FRONTEND_URL = 'http://localhost:3000';

// 测试结果统计
const results = {
    passed: 0,
    failed: 0,
    tests: []
};

function log(message, type = 'info') {
    const colors = {
        info: '\x1b[36m',
        success: '\x1b[32m',
        error: '\x1b[31m',
        warning: '\x1b[33m'
    };
    const reset = '\x1b[0m';
    const timestamp = new Date().toISOString().substr(11, 8);
    console.log(`${colors[type]}[${timestamp}] ${message}${reset}`);
}

function makeRequest(url, options = {}) {
    return new Promise((resolve, reject) => {
        const urlObj = new URL(url);
        const requestOptions = {
            hostname: urlObj.hostname,
            port: urlObj.port,
            path: urlObj.pathname + urlObj.search,
            method: options.method || 'GET',
            headers: {
                'Content-Type': 'application/json',
                'User-Agent': 'Integration-Test/1.0',
                ...options.headers
            }
        };

        const req = http.request(requestOptions, (res) => {
            let data = '';
            res.on('data', chunk => data += chunk);
            res.on('end', () => {
                try {
                    const parsed = data ? JSON.parse(data) : {};
                    resolve({ 
                        status: res.statusCode, 
                        data: parsed,
                        headers: res.headers 
                    });
                } catch (e) {
                    resolve({ 
                        status: res.statusCode, 
                        data: data,
                        headers: res.headers 
                    });
                }
            });
        });

        req.on('error', reject);
        
        if (options.body) {
            req.write(JSON.stringify(options.body));
        }
        req.end();
    });
}

async function testStep(name, testFunc, critical = false) {
    try {
        log(`🧪 测试: ${name}...`);
        const result = await testFunc();
        
        if (result.success) {
            log(`✅ ${name} - 成功`, 'success');
            results.passed++;
            results.tests.push({ name, status: 'PASSED', details: result.message });
            return result.data;
        } else {
            log(`❌ ${name} - 失败: ${result.message}`, 'error');
            results.failed++;
            results.tests.push({ name, status: 'FAILED', details: result.message });
            
            if (critical) {
                throw new Error(`关键测试失败: ${name}`);
            }
            return null;
        }
    } catch (error) {
        log(`💥 ${name} - 异常: ${error.message}`, 'error');
        results.failed++;
        results.tests.push({ name, status: 'ERROR', details: error.message });
        
        if (critical) {
            throw error;
        }
        return null;
    }
}

// 1. 数据库连接测试
async function testDatabaseConnection() {
    try {
        const result = execSync('psql -h localhost -U rocalight -d openpenpal -c "SELECT COUNT(*) as user_count FROM users;"', { encoding: 'utf8' });
        const userCount = parseInt(result.match(/\d+/)[0]);
        
        return {
            success: userCount > 0,
            message: `数据库连接成功，用户数量: ${userCount}`,
            data: { userCount }
        };
    } catch (error) {
        return {
            success: false,
            message: `数据库连接失败: ${error.message}`
        };
    }
}

// 2. 后端API健康检查
async function testBackendHealth() {
    const response = await makeRequest(`${BASE_URL}/health`);
    
    if (response.status !== 200) {
        return {
            success: false,
            message: `后端健康检查失败: ${response.status}`
        };
    }
    
    const health = response.data;
    const dbHealthy = health.database === 'healthy';
    const wsHealthy = health.websocket === 'healthy';
    
    return {
        success: dbHealthy && wsHealthy,
        message: `后端健康状态 - 数据库: ${health.database}, WebSocket: ${health.websocket}`,
        data: health
    };
}

// 3. 前端服务检查
async function testFrontendHealth() {
    try {
        const response = await makeRequest(FRONTEND_URL);
        return {
            success: response.status === 200,
            message: `前端服务状态: ${response.status}`,
            data: { status: response.status }
        };
    } catch (error) {
        return {
            success: false,
            message: `前端服务无法访问: ${error.message}`
        };
    }
}

// 4. 用户认证集成测试
async function testUserAuthIntegration() {
    // 登录获取token
    const loginResponse = await makeRequest(`${BASE_URL}/api/v1/auth/login`, {
        method: 'POST',
        body: {
            username: 'admin',
            password: 'AdminSecure2024'
        }
    });
    
    if (loginResponse.status !== 200 || !loginResponse.data.success) {
        return {
            success: false,
            message: `用户登录失败: ${JSON.stringify(loginResponse.data)}`
        };
    }
    
    const token = loginResponse.data.data.token;
    const user = loginResponse.data.data.user;
    
    // 验证用户信息
    const meResponse = await makeRequest(`${BASE_URL}/api/v1/auth/me`, {
        headers: {
            'Authorization': `Bearer ${token}`
        }
    });
    
    if (meResponse.status !== 200) {
        return {
            success: false,
            message: `获取用户信息失败: ${meResponse.status}`
        };
    }
    
    return {
        success: true,
        message: `用户认证集成成功 - 用户: ${user.username}, 角色: ${user.role}`,
        data: { token, user }
    };
}

// 5. 数据库用户查询测试
async function testDatabaseUserQuery() {
    try {
        const result = execSync(`psql -h localhost -U rocalight -d openpenpal -c "SELECT username, role, is_active FROM users WHERE username IN ('admin', 'courier_level1') ORDER BY username;"`, { encoding: 'utf8' });
        
        const hasAdmin = result.includes('admin');
        const hasCourier = result.includes('courier_level1');
        
        return {
            success: hasAdmin && hasCourier,
            message: `数据库用户查询成功 - admin: ${hasAdmin}, courier: ${hasCourier}`,
            data: { result }
        };
    } catch (error) {
        return {
            success: false,
            message: `数据库用户查询失败: ${error.message}`
        };
    }
}

// 6. WebSocket连接测试
async function testWebSocketConnection(token) {
    return new Promise((resolve) => {
        try {
            log('正在测试WebSocket连接...');
            const WebSocket = require('ws');
            const wsUrl = `ws://localhost:8080/api/v1/ws/connect?token=${encodeURIComponent(token)}`;
            const ws = new WebSocket(wsUrl);
            
            const timeout = setTimeout(() => {
                ws.close();
                resolve({
                    success: false,
                    message: 'WebSocket连接超时'
                });
            }, 5000);
            
            ws.on('open', () => {
                clearTimeout(timeout);
                ws.close();
                resolve({
                    success: true,
                    message: 'WebSocket连接成功',
                    data: { connected: true }
                });
            });
            
            ws.on('error', (error) => {
                clearTimeout(timeout);
                resolve({
                    success: false,
                    message: `WebSocket连接失败: ${error.message}`
                });
            });
        } catch (error) {
            resolve({
                success: false,
                message: `WebSocket测试异常: ${error.message}`
            });
        }
    });
}

// 7. AI功能集成测试
async function testAIIntegration() {
    const response = await makeRequest(`${BASE_URL}/api/v1/ai/personas`);
    
    if (response.status !== 200) {
        return {
            success: false,
            message: `AI功能测试失败: ${response.status}`
        };
    }
    
    const personas = response.data.data;
    const hasPersonas = Array.isArray(personas) && personas.length > 0;
    
    return {
        success: hasPersonas,
        message: `AI功能集成成功 - 角色数量: ${personas?.length || 0}`,
        data: { personas }
    };
}

// 8. 前端API代理测试
async function testFrontendAPIProxy() {
    try {
        // 测试前端API代理是否正常工作
        const response = await makeRequest(`${FRONTEND_URL}/api/health`);
        
        return {
            success: response.status === 200,
            message: `前端API代理${response.status === 200 ? '正常' : '异常'}: ${response.status}`,
            data: response.data
        };
    } catch (error) {
        return {
            success: false,
            message: `前端API代理测试失败: ${error.message}`
        };
    }
}

// 9. 信使等级权限测试
async function testCourierPermissions() {
    const courierTests = [
        { username: 'courier_level1', expectedRole: 'courier_level1' },
        { username: 'courier_level2', expectedRole: 'courier_level2' },
        { username: 'courier_level3', expectedRole: 'courier_level3' },
        { username: 'courier_level4', expectedRole: 'courier_level4' }
    ];
    
    let passedCount = 0;
    const results = [];
    
    for (const test of courierTests) {
        try {
            const loginResponse = await makeRequest(`${BASE_URL}/api/v1/auth/login`, {
                method: 'POST',
                body: {
                    username: test.username,
                    password: 'TestSecure2024'
                }
            });
            
            if (loginResponse.status === 200 && loginResponse.data.success) {
                const userRole = loginResponse.data.data.user.role;
                if (userRole === test.expectedRole) {
                    passedCount++;
                    results.push(`✅ ${test.username}: ${userRole}`);
                } else {
                    results.push(`❌ ${test.username}: 期望 ${test.expectedRole}, 实际 ${userRole}`);
                }
            } else {
                results.push(`❌ ${test.username}: 登录失败`);
            }
        } catch (error) {
            results.push(`❌ ${test.username}: 异常 ${error.message}`);
        }
    }
    
    return {
        success: passedCount === courierTests.length,
        message: `信使权限测试 - 通过: ${passedCount}/${courierTests.length}`,
        data: { results }
    };
}

// 10. 数据持久化测试
async function testDataPersistence() {
    try {
        // 检查用户密码是否正确存储
        const result = execSync(`psql -h localhost -U rocalight -d openpenpal -c "SELECT username, LEFT(password_hash, 10) as hash_preview, role FROM users WHERE username = 'admin';"`, { encoding: 'utf8' });
        
        const hasAdmin = result.includes('admin');
        const hasBcryptHash = result.includes('$2a$12$');
        
        return {
            success: hasAdmin && hasBcryptHash,
            message: `数据持久化正常 - 用户存在: ${hasAdmin}, 密码加密: ${hasBcryptHash}`,
            data: { result }
        };
    } catch (error) {
        return {
            success: false,
            message: `数据持久化测试失败: ${error.message}`
        };
    }
}

// 主测试流程
async function runIntegrationTest() {
    log('🚀 开始数据库和前端集成测试', 'info');
    log('=' * 100);
    
    try {
        // 1. 基础设施测试
        await testStep('数据库连接', testDatabaseConnection, true);
        await testStep('后端健康检查', testBackendHealth, true);
        await testStep('前端服务检查', testFrontendHealth);
        
        // 2. 认证集成测试
        const authData = await testStep('用户认证集成', testUserAuthIntegration, true);
        
        // 3. 数据库集成测试
        await testStep('数据库用户查询', testDatabaseUserQuery, true);
        await testStep('数据持久化', testDataPersistence);
        
        // 4. WebSocket集成测试
        if (authData?.token) {
            await testStep('WebSocket连接', () => testWebSocketConnection(authData.token));
        }
        
        // 5. 功能集成测试
        await testStep('AI功能集成', testAIIntegration);
        await testStep('信使权限系统', testCourierPermissions);
        
        // 6. 前端集成测试
        await testStep('前端API代理', testFrontendAPIProxy);
        
        // 汇总结果
        log('\n' + '=' * 100);
        log('📊 集成测试结果汇总', 'info');
        log(`✅ 成功: ${results.passed}`, 'success');
        log(`❌ 失败: ${results.failed}`, results.failed > 0 ? 'error' : 'success');
        
        const successRate = Math.round((results.passed / (results.passed + results.failed)) * 100);
        log(`📈 成功率: ${successRate}%`);
        
        log('\n📋 详细结果:', 'info');
        results.tests.forEach((test, index) => {
            const icon = test.status === 'PASSED' ? '✅' : test.status === 'FAILED' ? '❌' : '💥';
            const color = test.status === 'PASSED' ? 'success' : 'error';
            log(`${index + 1}. ${icon} ${test.name} - ${test.status}`, color);
            if (test.details) log(`   ${test.details}`);
        });
        
        // 系统状态评估
        if (results.failed === 0) {
            log('\n🎉 所有集成测试通过！', 'success');
            log('✅ 数据库 ↔ 后端 ↔ 前端 全链路正常工作', 'success');
            log('✅ 认证系统完整运行', 'success');
            log('✅ WebSocket实时通信正常', 'success');
            log('✅ 四级信使权限系统正常', 'success');
            log('✅ AI功能集成正常', 'success');
        } else if (results.failed <= 2) {
            log(`\n⚠️ 大部分功能正常，${results.failed}个非关键问题需要修复`, 'warning');
        } else {
            log(`\n❌ 发现${results.failed}个问题，需要重点关注`, 'error');
        }
        
    } catch (error) {
        log(`\n💥 集成测试异常终止: ${error.message}`, 'error');
        log('请检查系统各组件是否正常运行', 'error');
    }
}

// 运行集成测试
runIntegrationTest().catch(console.error);