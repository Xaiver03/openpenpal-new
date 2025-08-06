#!/usr/bin/env node

/**
 * 完整CSRF + JWT认证流程测试
 * 验证前端 → 后端的完整安全认证链路
 */

const http = require('http');

const FRONTEND_URL = 'http://localhost:3000';
const BACKEND_URL = 'http://localhost:8080';

// 测试结果
const results = {
    passed: 0,
    failed: 0,
    details: []
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
                        headers: res.headers,
                        cookies: res.headers['set-cookie'] || []
                    });
                } catch (e) {
                    resolve({ 
                        status: res.statusCode, 
                        data: data,
                        headers: res.headers,
                        cookies: res.headers['set-cookie'] || []
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

async function testStep(name, testFunc) {
    try {
        log(`🧪 测试: ${name}...`);
        const result = await testFunc();
        
        if (result.success) {
            log(`✅ ${name} - 通过`, 'success');
            results.passed++;
            results.details.push(`✅ ${name}: ${result.message}`);
            return result.data;
        } else {
            log(`❌ ${name} - 失败: ${result.message}`, 'error');
            results.failed++;
            results.details.push(`❌ ${name}: ${result.message}`);
            return null;
        }
    } catch (error) {
        log(`💥 ${name} - 异常: ${error.message}`, 'error');
        results.failed++;
        results.details.push(`💥 ${name}: ${error.message}`);
        return null;
    }
}

// 1. 测试后端CSRF Token生成
async function testBackendCSRFGeneration() {
    const response = await makeRequest(`${BACKEND_URL}/api/v1/auth/csrf`);
    
    if (response.status !== 200 || !response.data.success) {
        return {
            success: false,
            message: `后端CSRF生成失败: ${response.status} - ${JSON.stringify(response.data)}`
        };
    }
    
    const token = response.data.data.token;
    const hasCSRFCookie = response.cookies.some(cookie => cookie.includes('csrf-token='));
    
    return {
        success: token && hasCSRFCookie,
        message: `后端CSRF生成成功 - Token: ${token.substring(0, 16)}..., Cookie: ${hasCSRFCookie}`,
        data: { token, cookies: response.cookies }
    };
}

// 2. 测试前端CSRF代理
async function testFrontendCSRFProxy() {
    const response = await makeRequest(`${FRONTEND_URL}/api/auth/csrf`);
    
    if (response.status !== 200) {
        return {
            success: false,
            message: `前端CSRF代理失败: ${response.status}`
        };
    }
    
    // 前端响应格式不同
    const token = response.data.data?.token;
    const hasToken = token && token.length > 30;
    
    return {
        success: hasToken,
        message: `前端CSRF代理成功 - Token: ${token ? token.substring(0, 16) + '...' : 'null'}`,
        data: { token }
    };
}

// 3. 测试完整CSRF认证流程（后端直连）
async function testBackendCSRFAuth() {
    // Step 1: 获取CSRF Token
    const csrfResponse = await makeRequest(`${BACKEND_URL}/api/v1/auth/csrf`);
    
    if (csrfResponse.status !== 200 || !csrfResponse.data.success) {
        return {
            success: false,
            message: 'CSRF Token获取失败'
        };
    }
    
    const csrfToken = csrfResponse.data.data.token;
    const csrfCookie = csrfResponse.cookies.find(cookie => cookie.includes('csrf-token='));
    
    if (!csrfCookie) {
        return {
            success: false,
            message: 'CSRF Cookie未设置'
        };
    }
    
    // Step 2: 使用CSRF Token进行登录
    const loginResponse = await makeRequest(`${BACKEND_URL}/api/v1/auth/login`, {
        method: 'POST',
        headers: {
            'X-CSRF-Token': csrfToken,
            'Cookie': csrfCookie.split(';')[0] // 只取cookie值部分
        },
        body: {
            username: 'admin',
            password: 'AdminSecure2024'
        }
    });
    
    log(`登录响应状态: ${loginResponse.status}`, 'info');
    log(`登录响应数据: ${JSON.stringify(loginResponse.data).substring(0, 200)}...`, 'info');
    
    if (loginResponse.status !== 200 || !loginResponse.data.success) {
        return {
            success: false,
            message: `CSRF认证登录失败: ${JSON.stringify(loginResponse.data)}`
        };
    }
    
    const jwtToken = loginResponse.data.data.token;
    const user = loginResponse.data.data.user;
    
    return {
        success: true,
        message: `CSRF + JWT认证成功 - 用户: ${user.username}, 角色: ${user.role}`,
        data: { jwtToken, user }
    };
}

// 4. 测试前端CSRF认证流程
async function testFrontendCSRFAuth() {
    try {
        // Step 1: 通过前端代理获取CSRF Token
        const csrfResponse = await makeRequest(`${FRONTEND_URL}/api/auth/csrf`);
        
        if (csrfResponse.status !== 200) {
            return {
                success: false,
                message: `前端CSRF获取失败: ${csrfResponse.status}`
            };
        }
        
        const csrfToken = csrfResponse.data.data?.token;
        if (!csrfToken) {
            return {
                success: false,
                message: '前端CSRF Token为空'
            };
        }
        
        // Step 2: 使用前端代理进行登录
        const loginResponse = await makeRequest(`${FRONTEND_URL}/api/auth/login`, {
            method: 'POST',
            headers: {
                'X-CSRF-Token': csrfToken
            },
            body: {
                username: 'courier_level1',
                password: 'TestSecure2024'
            }
        });
        
        log(`前端登录响应状态: ${loginResponse.status}`, 'info');
        
        if (loginResponse.status !== 200) {
            return {
                success: false,
                message: `前端CSRF登录失败: ${loginResponse.status}`
            };
        }
        
        // 前端返回的格式可能不同
        const isSuccess = loginResponse.data.code === 0 || loginResponse.data.success;
        const userData = loginResponse.data.data;
        
        return {
            success: isSuccess && userData,
            message: isSuccess ? 
                `前端CSRF认证成功 - 用户: ${userData?.user?.username}` :
                `前端CSRF认证失败: ${JSON.stringify(loginResponse.data)}`,
            data: userData
        };
    } catch (error) {
        return {
            success: false,
            message: `前端CSRF认证异常: ${error.message}`
        };
    }
}

// 5. 测试JWT Token验证
async function testJWTValidation(jwtToken) {
    if (!jwtToken) {
        return {
            success: false,
            message: 'JWT Token为空，跳过验证'
        };
    }
    
    const response = await makeRequest(`${BACKEND_URL}/api/v1/auth/me`, {
        headers: {
            'Authorization': `Bearer ${jwtToken}`
        }
    });
    
    if (response.status !== 200 || !response.data.success) {
        return {
            success: false,
            message: `JWT验证失败: ${response.status}`
        };
    }
    
    return {
        success: true,
        message: `JWT验证成功 - 用户: ${response.data.data.username}`,
        data: response.data.data
    };
}

// 6. 测试错误CSRF Token被拒绝
async function testInvalidCSRFRejection() {
    const response = await makeRequest(`${BACKEND_URL}/api/v1/auth/login`, {
        method: 'POST',
        headers: {
            'X-CSRF-Token': 'invalid-token-should-be-rejected'
        },
        body: {
            username: 'admin',
            password: 'AdminSecure2024'
        }
    });
    
    // 应该被拒绝
    const shouldBeRejected = response.status === 403 || response.status === 401;
    
    return {
        success: shouldBeRejected,
        message: `错误CSRF Token ${shouldBeRejected ? '正确被拒绝' : '未被拒绝'}: ${response.status}`,
        data: { status: response.status }
    };
}

// 主测试流程
async function runCompleteCSRFJWTTest() {
    log('🚀 开始完整CSRF + JWT认证测试', 'info');
    log('=' * 100);
    
    try {
        // 1. 后端CSRF测试
        const backendCSRFData = await testStep('1. 后端CSRF Token生成', testBackendCSRFGeneration);
        
        // 2. 前端CSRF代理测试
        await testStep('2. 前端CSRF代理', testFrontendCSRFProxy);
        
        // 3. 后端完整认证测试  
        const backendAuthData = await testStep('3. 后端CSRF + JWT认证', testBackendCSRFAuth);
        
        // 4. 前端完整认证测试
        await testStep('4. 前端CSRF认证流程', testFrontendCSRFAuth);
        
        // 5. JWT Token验证
        if (backendAuthData?.jwtToken) {
            await testStep('5. JWT Token验证', () => testJWTValidation(backendAuthData.jwtToken));
        }
        
        // 6. 安全性测试 - 错误CSRF被拒绝
        await testStep('6. 错误CSRF Token拒绝', testInvalidCSRFRejection);
        
        // 汇总结果
        log('\\n' + '=' * 100);
        log('📊 CSRF + JWT认证测试结果', 'info');
        log(`✅ 通过: ${results.passed}`, 'success');
        log(`❌ 失败: ${results.failed}`, results.failed > 0 ? 'error' : 'success');
        
        const successRate = Math.round((results.passed / (results.passed + results.failed)) * 100);
        log(`📈 成功率: ${successRate}%`);
        
        log('\\n📋 详细结果:', 'info');
        results.details.forEach((detail, index) => {
            console.log(`${index + 1}. ${detail}`);
        });
        
        // 最终安全评估
        if (results.failed === 0) {
            log('\\n🎉 所有CSRF + JWT认证测试通过！', 'success');
            log('✅ CSRF Token生成和验证正常', 'success');
            log('✅ JWT Token生成和验证正常', 'success');
            log('✅ 前端 ↔ 后端认证链路安全', 'success');
            log('✅ 错误CSRF Token正确被拒绝', 'success');
            log('✅ 双重安全认证机制完全正常', 'success');
        } else if (results.failed <= 1) {
            log(`\\n⚠️ 大部分安全机制正常，${results.failed}个问题需要修复`, 'warning');
        } else {
            log(`\\n❌ 发现${results.failed}个安全问题，需要重点修复`, 'error');
        }
        
    } catch (error) {
        log(`\\n💥 认证测试异常终止: ${error.message}`, 'error');
    }
}

// 运行测试
runCompleteCSRFJWTTest().catch(console.error);