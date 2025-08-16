#!/usr/bin/env node
/**
 * 简化版API测试脚本 - 绕过CSRF直接测试核心功能
 */

const fetch = globalThis.fetch;

const CONFIG = {
    backend: 'http://localhost:8080',
    testUsers: [
        { username: 'alice', password: 'password' },
        { username: 'admin', password: 'password' },
        { username: 'user1', password: 'password' }
    ]
};

class SimpleAPITester {
    constructor() {
        this.results = [];
        this.authToken = null;
    }

    async log(message, type = 'info') {
        const timestamp = new Date().toISOString();
        const prefix = {
            'error': '❌',
            'success': '✅', 
            'warning': '⚠️',
            'info': 'ℹ️'
        }[type] || 'ℹ️';
        
        console.log(`${timestamp} ${prefix} ${message}`);
    }

    async test(name, testFn) {
        try {
            await this.log(`测试: ${name}`);
            const result = await testFn();
            this.results.push({ name, success: true, result });
            await this.log(`✓ ${name}`, 'success');
            return result;
        } catch (error) {
            this.results.push({ name, success: false, error: error.message });
            await this.log(`✗ ${name}: ${error.message}`, 'error');
            return null;
        }
    }

    // 测试健康检查
    async testHealth() {
        return await this.test('后端健康检查', async () => {
            const response = await fetch(`${CONFIG.backend}/health`);
            if (!response.ok) {
                throw new Error(`Health check failed: ${response.status}`);
            }
            const data = await response.json();
            return data;
        });
    }

    // 测试用户登录（尝试多个用户）
    async testLogin() {
        for (const user of CONFIG.testUsers) {
            const result = await this.test(`登录测试 - ${user.username}`, async () => {
                // 首先尝试直接登录（可能绕过CSRF）
                const response = await fetch(`${CONFIG.backend}/api/v1/auth/login`, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify(user)
                });

                if (response.status === 403) {
                    // CSRF问题，尝试获取token后再登录
                    const csrfResponse = await fetch(`${CONFIG.backend}/api/v1/auth/csrf`);
                    if (!csrfResponse.ok) {
                        throw new Error('无法获取CSRF token');
                    }
                    
                    const csrfData = await csrfResponse.json();
                    const csrfToken = csrfData.data.token;
                    
                    // 重新尝试登录
                    const loginResponse = await fetch(`${CONFIG.backend}/api/v1/auth/login`, {
                        method: 'POST',
                        headers: {
                            'Content-Type': 'application/json',
                            'X-CSRF-Token': csrfToken
                        },
                        body: JSON.stringify(user)
                    });
                    
                    if (!loginResponse.ok) {
                        const errorText = await loginResponse.text();
                        throw new Error(`登录失败: ${loginResponse.status} - ${errorText}`);
                    }
                    
                    const loginData = await loginResponse.json();
                    if (loginData.success && loginData.data?.token) {
                        this.authToken = loginData.data.token;
                        return { 
                            user: user.username, 
                            token: this.authToken.substring(0, 20) + '...',
                            userData: loginData.data.user
                        };
                    }
                } else if (response.ok) {
                    const data = await response.json();
                    if (data.success && data.data?.token) {
                        this.authToken = data.data.token;
                        return { 
                            user: user.username, 
                            token: this.authToken.substring(0, 20) + '...',
                            userData: data.data.user
                        };
                    }
                }

                throw new Error(`登录失败: 响应不包含有效token`);
            });

            if (result) {
                await this.log(`成功登录用户: ${user.username}`, 'success');
                break; // 成功登录一个用户就停止
            }
        }
    }

    // 测试认证API
    async testAuthenticatedAPIs() {
        if (!this.authToken) {
            await this.log('跳过认证API测试 - 没有有效token', 'warning');
            return;
        }

        const headers = {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${this.authToken}`
        };

        // 测试用户信息
        await this.test('获取当前用户信息', async () => {
            const response = await fetch(`${CONFIG.backend}/api/v1/auth/me`, {
                headers
            });
            
            if (!response.ok) {
                throw new Error(`用户信息获取失败: ${response.status}`);
            }
            
            const data = await response.json();
            return data.data || data;
        });

        // 测试信件API  
        await this.test('获取信件列表', async () => {
            const response = await fetch(`${CONFIG.backend}/api/v1/letters`, {
                headers
            });
            
            if (!response.ok) {
                throw new Error(`信件列表获取失败: ${response.status}`);
            }
            
            const data = await response.json();
            return {
                letters: data.data?.letters || data.letters || [],
                total: data.data?.total || data.total || 0
            };
        });

        // 测试商品API
        await this.test('获取商品列表', async () => {
            const response = await fetch(`${CONFIG.backend}/api/v1/shop/products`, {
                headers
            });
            
            if (!response.ok) {
                throw new Error(`商品列表获取失败: ${response.status}`);
            }
            
            const data = await response.json();
            return {
                products: data.data?.products || data.products || [],
                total: data.data?.total || data.total || 0
            };
        });

        // 测试AI API
        await this.test('AI灵感生成', async () => {
            const response = await fetch(`${CONFIG.backend}/api/v1/ai/inspiration`, {
                method: 'POST',
                headers,
                body: JSON.stringify({
                    theme: '校园生活',
                    count: 2
                })
            });
            
            if (!response.ok) {
                throw new Error(`AI灵感生成失败: ${response.status}`);
            }
            
            const data = await response.json();
            return {
                inspirations: data.data?.inspirations || [],
                message: data.message
            };
        });
    }

    // 测试数据库数据一致性
    async testDataConsistency() {
        const { execSync } = require('child_process');
        
        await this.test('数据库用户数量检查', async () => {
            try {
                const result = execSync(
                    'psql -h localhost -p 5432 -U rocalight -d openpenpal -t -c "SELECT COUNT(*) FROM users;"',
                    { encoding: 'utf8', timeout: 5000 }
                );
                const userCount = parseInt(result.trim());
                return { userCount, hasUsers: userCount > 0 };
            } catch (error) {
                throw new Error(`数据库查询失败: ${error.message}`);
            }
        });

        await this.test('数据库商品数量检查', async () => {
            try {
                const result = execSync(
                    'psql -h localhost -p 5432 -U rocalight -d openpenpal -t -c "SELECT COUNT(*) FROM products;"',
                    { encoding: 'utf8', timeout: 5000 }
                );
                const productCount = parseInt(result.trim());
                return { productCount, hasProducts: productCount > 0 };
            } catch (error) {
                throw new Error(`数据库查询失败: ${error.message}`);
            }
        });
    }

    // 运行所有测试
    async runAllTests() {
        await this.log('🚀 开始OpenPenPal简化API测试');
        
        try {
            // 1. 健康检查
            await this.testHealth();
            
            // 2. 用户登录
            await this.testLogin();
            
            // 3. 认证API测试
            await this.testAuthenticatedAPIs();
            
            // 4. 数据一致性检查
            await this.testDataConsistency();
            
        } catch (error) {
            await this.log(`测试过程中发生错误: ${error.message}`, 'error');
        }

        // 生成报告
        await this.generateReport();
    }

    async generateReport() {
        await this.log('\n📊 ===== 测试报告 =====');
        
        const total = this.results.length;
        const successful = this.results.filter(r => r.success).length;
        const successRate = total > 0 ? ((successful / total) * 100).toFixed(1) : 0;
        
        await this.log(`📈 总测试数: ${total}`);
        await this.log(`✅ 成功: ${successful}`, 'success');
        await this.log(`❌ 失败: ${total - successful}`, total === successful ? 'info' : 'error');
        await this.log(`📊 成功率: ${successRate}%`, successRate >= 80 ? 'success' : 'warning');
        
        // 显示失败的测试
        const failures = this.results.filter(r => !r.success);
        if (failures.length > 0) {
            await this.log('\n❌ 失败的测试:', 'error');
            for (const failure of failures) {
                await this.log(`  • ${failure.name}: ${failure.error}`, 'error');
            }
        }

        // 关键发现
        await this.log('\n🔍 关键发现:');
        const healthTest = this.results.find(r => r.name === '后端健康检查');
        if (healthTest?.success) {
            await this.log('✅ 后端服务运行正常', 'success');
        }

        const loginTests = this.results.filter(r => r.name.includes('登录测试'));
        const successfulLogins = loginTests.filter(r => r.success);
        if (successfulLogins.length > 0) {
            await this.log(`✅ 用户认证正常 (${successfulLogins.length}/${loginTests.length})`, 'success');
        } else if (loginTests.length > 0) {
            await this.log('❌ 用户认证存在问题', 'error');
        }

        await this.log('\n✨ 测试完成！', 'success');
        
        // 保存结果
        require('fs').writeFileSync(
            'simple-api-test-report.json',
            JSON.stringify({
                timestamp: new Date().toISOString(),
                summary: { total, successful, failed: total - successful, successRate },
                results: this.results
            }, null, 2)
        );
        
        await this.log('📄 详细报告已保存到: simple-api-test-report.json');
    }
}

// 执行测试
async function main() {
    const tester = new SimpleAPITester();
    await tester.runAllTests();
}

if (require.main === module) {
    main().catch(console.error);
}

module.exports = SimpleAPITester;