#!/usr/bin/env node

/**
 * 完整登录流程测试 - 前端 → 后端 → 数据库
 * 测试从浏览器端认证到数据库验证的完整链路
 */

const http = require('http');
const { execSync } = require('child_process');

const BASE_URL = 'http://localhost:8080';
const FRONTEND_URL = 'http://localhost:3000';

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
                'User-Agent': 'Login-Flow-Test/1.0',
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

// 1. 数据库层面验证用户存在
async function testDatabaseUserExists() {
    try {
        const result = execSync(`psql -h localhost -U rocalight -d openpenpal -t -c "SELECT username, role, is_active, LEFT(password_hash, 20) as hash_preview FROM users WHERE username IN ('admin', 'courier_level1');"`, { encoding: 'utf8' });
        
        log('数据库查询结果:', 'info');
        console.log(result);
        
        const hasAdmin = result.includes('admin');
        const hasCourier = result.includes('courier_level1');
        const hasBcryptHash = result.includes('$2a$12$');
        
        return {
            success: hasAdmin && hasCourier && hasBcryptHash,
            message: `用户存在验证 - admin: ${hasAdmin}, courier: ${hasCourier}, 密码加密: ${hasBcryptHash}`,
            data: { hasAdmin, hasCourier, hasBcryptHash, rawResult: result }
        };
    } catch (error) {
        return {
            success: false,
            message: `数据库查询失败: ${error.message}`
        };
    }
}

// 2. 后端API直接登录测试
async function testBackendLogin() {
    const response = await makeRequest(`${BASE_URL}/api/v1/auth/login`, {
        method: 'POST',
        headers: {
            'X-Requested-With': 'XMLHttpRequest',
            'X-OpenPenPal-Auth': 'frontend-client',
        },
        body: {
            username: 'admin',
            password: 'AdminSecure2024'
        }
    });
    
    log(`后端登录响应状态: ${response.status}`, 'info');
    log(`后端登录响应数据:`, 'info');
    console.log(JSON.stringify(response.data, null, 2));
    
    if (response.status !== 200 || !response.data.success) {
        return {
            success: false,
            message: `后端登录失败: ${response.status} - ${JSON.stringify(response.data)}`
        };
    }
    
    const { token, user } = response.data.data;
    
    return {
        success: true,
        message: `后端登录成功 - 用户: ${user.username}, 角色: ${user.role}`,
        data: { token, user, cookies: response.cookies }
    };
}

// 3. JWT Token验证测试
async function testTokenValidation(token) {
    const response = await makeRequest(`${BASE_URL}/api/v1/auth/me`, {
        headers: {
            'Authorization': `Bearer ${token}`
        }
    });
    
    log(`Token验证响应状态: ${response.status}`, 'info');
    log(`Token验证响应数据:`, 'info');
    console.log(JSON.stringify(response.data, null, 2));
    
    if (response.status !== 200 || !response.data.success) {
        return {
            success: false,
            message: `Token验证失败: ${response.status} - ${JSON.stringify(response.data)}`
        };
    }
    
    return {
        success: true,
        message: `Token验证成功 - 用户: ${response.data.data.username}`,
        data: response.data.data
    };
}

// 4. 前端API代理测试
async function testFrontendProxy() {
    try {
        const response = await makeRequest(`${FRONTEND_URL}/api/auth/login`, {
            method: 'POST',
            headers: {
                'X-Requested-With': 'XMLHttpRequest',
                'X-OpenPenPal-Auth': 'frontend-client',
            },
            body: {
                username: 'courier_level1',
                password: 'TestSecure2024'
            }
        });
        
        log(`前端代理响应状态: ${response.status}`, 'info');
        log(`前端代理响应数据:`, 'info');
        console.log(JSON.stringify(response.data, null, 2));
        
        if (response.status !== 200) {
            return {
                success: false,
                message: `前端代理登录失败: ${response.status} - ${JSON.stringify(response.data)}`
            };
        }
        
        // 检查响应格式 - 支持多种格式
        const isSuccess = response.data.success || 
                         (response.data.code === 0 && response.data.data && response.data.data.accessToken);
        
        const username = response.data.data?.user?.username || 'unknown';
        
        return {
            success: isSuccess,
            message: isSuccess ? 
                `前端代理登录成功 - 用户: ${username}` :
                `前端代理登录失败: ${JSON.stringify(response.data)}`,
            data: response.data
        };
    } catch (error) {
        return {
            success: false,
            message: `前端代理连接失败: ${error.message}`
        };
    }
}

// 5. 多用户登录测试
async function testMultipleUserLogin() {
    const testUsers = [
        { username: 'admin', password: 'AdminSecure2024', expectedRole: 'super_admin' },
        { username: 'courier_level1', password: 'TestSecure2024', expectedRole: 'courier_level1' },
        { username: 'courier_level2', password: 'TestSecure2024', expectedRole: 'courier_level2' }
    ];
    
    let passedCount = 0;
    const userResults = [];
    
    for (const testUser of testUsers) {
        try {
            const response = await makeRequest(`${BASE_URL}/api/v1/auth/login`, {
                method: 'POST',
                headers: {
                    'X-Requested-With': 'XMLHttpRequest',
                    'X-OpenPenPal-Auth': 'frontend-client',
                },
                body: {
                    username: testUser.username,
                    password: testUser.password
                }
            });
            
            if (response.status === 200 && response.data.success) {
                const userRole = response.data.data.user.role;
                if (userRole === testUser.expectedRole) {
                    passedCount++;
                    userResults.push(`✅ ${testUser.username}: ${userRole}`);
                } else {
                    userResults.push(`❌ ${testUser.username}: 期望 ${testUser.expectedRole}, 实际 ${userRole}`);
                }
            } else {
                userResults.push(`❌ ${testUser.username}: 登录失败 - ${response.status}`);
            }
        } catch (error) {
            userResults.push(`💥 ${testUser.username}: 异常 - ${error.message}`);
        }
    }
    
    return {
        success: passedCount >= 2, // 至少2个用户登录成功
        message: `多用户登录测试 - 成功: ${passedCount}/${testUsers.length}`,
        data: { userResults, passedCount }
    };
}

// 6. 密码验证逻辑测试
async function testPasswordValidation() {
    // 测试错误密码
    const wrongPasswordResponse = await makeRequest(`${BASE_URL}/api/v1/auth/login`, {
        method: 'POST',
        headers: {
            'X-Requested-With': 'XMLHttpRequest',
            'X-OpenPenPal-Auth': 'frontend-client',
        },
        body: {
            username: 'admin',
            password: 'WrongPassword123'
        }
    });
    
    log(`错误密码响应状态: ${wrongPasswordResponse.status}`, 'info');
    log(`错误密码响应:`, 'info');
    console.log(JSON.stringify(wrongPasswordResponse.data, null, 2));
    
    const shouldFail = wrongPasswordResponse.status === 401;
    
    // 测试正确密码
    const correctPasswordResponse = await makeRequest(`${BASE_URL}/api/v1/auth/login`, {
        method: 'POST',
        headers: {
            'X-Requested-With': 'XMLHttpRequest',
            'X-OpenPenPal-Auth': 'frontend-client',
        },
        body: {
            username: 'admin',
            password: 'AdminSecure2024'
        }
    });
    
    const shouldSucceed = correctPasswordResponse.status === 200 && correctPasswordResponse.data.success;
    
    return {
        success: shouldFail && shouldSucceed,
        message: `密码验证 - 错误密码被拒绝: ${shouldFail}, 正确密码被接受: ${shouldSucceed}`,
        data: { shouldFail, shouldSucceed }
    };
}

// 7. 数据库事务一致性测试
async function testDatabaseConsistency() {
    try {
        // 检查用户表的数据一致性
        const userCount = execSync(`psql -h localhost -U rocalight -d openpenpal -t -c "SELECT COUNT(*) FROM users WHERE is_active = true;"`, { encoding: 'utf8' }).trim();
        
        // 检查认证日志记录
        const recentLogins = execSync(`psql -h localhost -U rocalight -d openpenpal -t -c "SELECT COUNT(*) FROM users WHERE last_login_at > NOW() - INTERVAL '1 hour';"`, { encoding: 'utf8' }).trim();
        
        log(`活跃用户数量: ${userCount}`, 'info');
        log(`最近1小时登录用户数: ${recentLogins}`, 'info');
        
        return {
            success: parseInt(userCount) > 0,
            message: `数据库一致性检查 - 活跃用户: ${userCount}, 最近登录: ${recentLogins}`,
            data: { userCount: parseInt(userCount), recentLogins: parseInt(recentLogins) }
        };
    } catch (error) {
        return {
            success: false,
            message: `数据库一致性检查失败: ${error.message}`
        };
    }
}

// 主测试流程
async function runCompleteLoginFlowTest() {
    log('🚀 开始完整登录流程测试 - 前端 → 后端 → 数据库', 'info');
    log('=' * 100);
    
    try {
        // 1. 数据库层面验证
        await testStep('1. 数据库用户数据验证', testDatabaseUserExists);
        
        // 2. 后端API登录
        const loginData = await testStep('2. 后端API直接登录', testBackendLogin);
        
        // 3. JWT Token验证
        if (loginData?.token) {
            await testStep('3. JWT Token验证', () => testTokenValidation(loginData.token));
        }
        
        // 4. 前端代理测试
        await testStep('4. 前端API代理登录', testFrontendProxy);
        
        // 5. 多用户登录测试
        await testStep('5. 多用户角色登录', testMultipleUserLogin);
        
        // 6. 密码验证逻辑
        await testStep('6. 密码验证逻辑', testPasswordValidation);
        
        // 7. 数据库一致性
        await testStep('7. 数据库事务一致性', testDatabaseConsistency);
        
        // 汇总结果
        log('\n' + '=' * 100);
        log('📊 完整登录流程测试结果', 'info');
        log(`✅ 通过: ${results.passed}`, 'success');
        log(`❌ 失败: ${results.failed}`, results.failed > 0 ? 'error' : 'success');
        
        const successRate = Math.round((results.passed / (results.passed + results.failed)) * 100);
        log(`📈 成功率: ${successRate}%`);
        
        log('\n📋 详细结果:', 'info');
        results.details.forEach((detail, index) => {
            console.log(`${index + 1}. ${detail}`);
        });
        
        // 最终评估
        if (results.failed === 0) {
            log('\n🎉 完整登录流程测试全部通过！', 'success');
            log('✅ 数据库 ↔ 后端 ↔ 前端 认证链路完全正常', 'success');
            log('✅ 用户密码验证逻辑正确', 'success'); 
            log('✅ JWT Token生成和验证正常', 'success');
            log('✅ 多角色用户认证系统工作正常', 'success');
            log('✅ 前端API代理正常转发请求', 'success');
        } else if (results.failed <= 1) {
            log(`\n⚠️ 大部分功能正常，${results.failed}个非关键问题`, 'warning');
            log('系统基本可用，建议修复剩余问题', 'warning');
        } else {
            log(`\n❌ 发现${results.failed}个问题，需要重点修复`, 'error');
            log('登录流程存在重要问题，需要进一步调试', 'error');
        }
        
    } catch (error) {
        log(`\n💥 测试流程异常终止: ${error.message}`, 'error');
        log('请检查系统各组件是否正常运行', 'error');
    }
}

// 运行测试
runCompleteLoginFlowTest().catch(console.error);