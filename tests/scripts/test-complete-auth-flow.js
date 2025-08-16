#!/usr/bin/env node

/**
 * 完整的认证流程测试 - 彻底排查所有认证问题
 * 测试所有环节：CSRF -> 注册 -> 登录 -> 令牌验证 -> 刷新令牌
 */

const http = require('http');

const BASE_URL = 'http://localhost:8080';

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
                'User-Agent': 'AuthTest/1.0',
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
        log(`🧪 测试步骤: ${name}...`);
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

// 测试步骤1: 后端健康检查
async function testHealthCheck() {
    const response = await makeRequest(`${BASE_URL}/health`);
    return {
        success: response.status === 200,
        message: `Status: ${response.status}`,
        data: response.data
    };
}

// 测试步骤2: 获取CSRF令牌
async function testGetCSRFToken() {
    const response = await makeRequest(`${BASE_URL}/api/v1/auth/csrf`);
    
    if (response.status !== 200) {
        return {
            success: false,
            message: `CSRF端点返回 ${response.status}: ${JSON.stringify(response.data)}`
        };
    }
    
    if (!response.data.success || !response.data.data?.token) {
        return {
            success: false,
            message: `CSRF响应格式错误: ${JSON.stringify(response.data)}`
        };
    }
    
    return {
        success: true,
        message: `CSRF令牌获取成功: ${response.data.data.token.substring(0, 16)}...`,
        data: {
            token: response.data.data.token,
            expires_at: response.data.data.expires_at,
            cookie: response.headers['set-cookie']
        }
    };
}

// 测试步骤3: 用户注册
async function testUserRegistration(csrfToken) {
    const username = `testuser_${Date.now()}`;
    const email = `${username}@test.com`;
    const password = 'TestPassword2024!';
    
    const response = await makeRequest(`${BASE_URL}/api/v1/auth/register`, {
        method: 'POST',
        headers: {
            'X-CSRF-Token': csrfToken
        },
        body: {
            username: username,
            email: email,
            password: password,
            nickname: 'Test User',
            school_code: 'TEST001'
        }
    });
    
    if (response.status !== 201) {
        return {
            success: false,
            message: `注册失败 ${response.status}: ${JSON.stringify(response.data)}`
        };
    }
    
    return {
        success: true,
        message: `用户注册成功: ${username}`,
        data: { username, email, password }
    };
}

// 测试步骤4: 用户登录（使用已知用户）
async function testUserLogin() {
    const credentials = {
        username: 'admin',
        password: 'AdminSecure2024!'
    };
    
    const response = await makeRequest(`${BASE_URL}/api/v1/auth/login`, {
        method: 'POST',
        body: credentials
    });
    
    if (response.status !== 200) {
        return {
            success: false,
            message: `登录失败 ${response.status}: ${JSON.stringify(response.data)}`
        };
    }
    
    if (!response.data.success || !response.data.data?.token) {
        return {
            success: false,
            message: `登录响应格式错误: ${JSON.stringify(response.data)}`
        };
    }
    
    return {
        success: true,
        message: `用户登录成功: ${credentials.username}`,
        data: {
            token: response.data.data.token,
            user: response.data.data.user
        }
    };
}

// 测试步骤5: 验证JWT令牌
async function testTokenValidation(token) {
    const response = await makeRequest(`${BASE_URL}/api/v1/auth/me`, {
        headers: {
            'Authorization': `Bearer ${token}`
        }
    });
    
    if (response.status !== 200) {
        return {
            success: false,
            message: `令牌验证失败 ${response.status}: ${JSON.stringify(response.data)}`
        };
    }
    
    return {
        success: true,
        message: `令牌验证成功: ${response.data.data?.username}`,
        data: response.data.data
    };
}

// 测试步骤6: 检查令牌过期时间
async function testTokenExpiry(token) {
    const response = await makeRequest(`${BASE_URL}/api/v1/auth/check-expiry`, {
        headers: {
            'Authorization': `Bearer ${token}`
        }
    });
    
    if (response.status !== 200) {
        return {
            success: false,
            message: `令牌过期检查失败 ${response.status}: ${JSON.stringify(response.data)}`
        };
    }
    
    return {
        success: true,
        message: `令牌过期检查成功，剩余时间: ${Math.round(response.data.data.remaining_time)}秒`,
        data: response.data.data
    };
}

// 测试步骤7: 刷新令牌
async function testTokenRefresh(token) {
    const response = await makeRequest(`${BASE_URL}/api/v1/auth/refresh`, {
        method: 'POST',
        headers: {
            'Authorization': `Bearer ${token}`
        }
    });
    
    if (response.status !== 200) {
        return {
            success: false,
            message: `令牌刷新失败 ${response.status}: ${JSON.stringify(response.data)}`
        };
    }
    
    return {
        success: true,
        message: `令牌刷新成功`,
        data: {
            token: response.data.data.token,
            expires_at: response.data.data.expires_at
        }
    };
}

// 主测试流程
async function runCompleteAuthTest() {
    log('🚀 开始完整认证流程测试', 'info');
    log('=' * 80);
    
    try {
        // 1. 健康检查
        await testStep('后端健康检查', testHealthCheck, true);
        
        // 2. 获取CSRF令牌
        const csrfData = await testStep('获取CSRF令牌', testGetCSRFToken, true);
        
        // 3. 测试用户登录（跳过注册，使用已知用户）
        const loginData = await testStep('用户登录', testUserLogin, true);
        
        // 4. 验证JWT令牌
        await testStep('JWT令牌验证', () => testTokenValidation(loginData.token), true);
        
        // 5. 检查令牌过期时间
        await testStep('令牌过期检查', () => testTokenExpiry(loginData.token));
        
        // 6. 刷新令牌
        const refreshData = await testStep('令牌刷新', () => testTokenRefresh(loginData.token));
        
        // 7. 验证新令牌
        if (refreshData?.token) {
            await testStep('新令牌验证', () => testTokenValidation(refreshData.token));
        }
        
        log('\n' + '=' * 80);
        log('📊 测试结果汇总', 'info');
        log(`✅ 成功: ${results.passed}`, 'success');
        log(`❌ 失败: ${results.failed}`, results.failed > 0 ? 'error' : 'success');
        log(`📈 成功率: ${Math.round((results.passed / (results.passed + results.failed)) * 100)}%`);
        
        log('\n📋 详细结果:', 'info');
        results.tests.forEach((test, index) => {
            const icon = test.status === 'PASSED' ? '✅' : test.status === 'FAILED' ? '❌' : '💥';
            const color = test.status === 'PASSED' ? 'success' : 'error';
            log(`${index + 1}. ${icon} ${test.name} - ${test.status}`, color);
            if (test.details) log(`   ${test.details}`);
        });
        
        if (results.failed === 0) {
            log('\n🎉 所有认证测试通过！系统认证流程正常工作。', 'success');
        } else {
            log(`\n⚠️  ${results.failed} 个测试失败。需要修复相关问题。`, 'warning');
        }
        
    } catch (error) {
        log(`\n💥 测试流程异常终止: ${error.message}`, 'error');
        log('请检查后端服务是否正常运行。', 'error');
    }
}

// 运行测试
runCompleteAuthTest().catch(console.error);