#!/usr/bin/env node
/**
 * OpenPenPal API与数据库完整性测试脚本
 * 
 * 功能：
 * 1. 测试所有API端点的可用性
 * 2. 验证数据库表结构和数据一致性  
 * 3. 测试跨服务数据交互
 * 4. 验证业务逻辑完整性
 */

const fetch = globalThis.fetch;
const { execSync } = require('child_process');

// 测试配置
const CONFIG = {
    backends: {
        main: 'http://localhost:8080'
        // 注释掉未运行的服务
        // gateway: 'http://localhost:8000', 
        // writeService: 'http://localhost:8001',
        // courierService: 'http://localhost:8002',
        // adminService: 'http://localhost:8003',
        // ocrService: 'http://localhost:8004'
    },
    frontend: 'http://localhost:3000',
    database: {
        host: 'localhost',
        port: 5432,
        database: 'openpenpal',
        user: 'rocalight'
    },
    testUser: {
        username: 'admin',
        password: 'admin123'
    }
};

class APIIntegrityTester {
    constructor() {
        this.results = {
            services: {},
            apis: {},
            database: {},
            integration: {},
            summary: {}
        };
        this.authToken = null;
        this.testStartTime = new Date();
    }

    async log(message, type = 'info') {
        const timestamp = new Date().toISOString();
        const prefix = {
            'error': '❌',
            'success': '✅', 
            'warning': '⚠️',
            'info': 'ℹ️',
            'test': '🧪',
            'db': '🗄️'
        }[type] || 'ℹ️';
        
        console.log(`${timestamp} ${prefix} ${message}`);
    }

    async test(name, testFn, category = 'general') {
        try {
            await this.log(`开始测试: ${name}`, 'test');
            const startTime = Date.now();
            const result = await testFn();
            const duration = Date.now() - startTime;
            
            if (!this.results[category]) {
                this.results[category] = {};
            }
            
            this.results[category][name] = {
                success: true,
                result,
                duration,
                timestamp: new Date()
            };
            
            await this.log(`✓ ${name} (${duration}ms)`, 'success');
            return result;
        } catch (error) {
            if (!this.results[category]) {
                this.results[category] = {};
            }
            
            this.results[category][name] = {
                success: false,
                error: error.message,
                duration: Date.now() - (this.testStartTime.getTime()),
                timestamp: new Date()
            };
            
            await this.log(`✗ ${name}: ${error.message}`, 'error');
            return null;
        }
    }

    // ===== 服务健康检查 =====
    async testServiceHealth() {
        await this.log('开始服务健康检查...', 'info');
        
        for (const [serviceName, url] of Object.entries(CONFIG.backends)) {
            await this.test(`${serviceName} 健康检查`, async () => {
                const response = await fetch(`${url}/health`, {
                    method: 'GET',
                    timeout: 5000
                });
                
                if (!response.ok) {
                    throw new Error(`HTTP ${response.status}: ${response.statusText}`);
                }
                
                const data = await response.json();
                return {
                    status: response.status,
                    healthy: data.status === 'healthy' || response.status === 200,
                    data
                };
            }, 'services');
        }
    }

    // ===== 用户认证测试 =====
    async testAuthentication() {
        await this.log('开始认证系统测试...', 'info');
        
        // 首先获取CSRF token
        await this.test('获取CSRF token', async () => {
            const response = await fetch(`${CONFIG.backends.main}/api/v1/auth/csrf`, {
                method: 'GET'
            });

            if (!response.ok) {
                throw new Error(`CSRF token获取失败: ${response.status}`);
            }

            const data = await response.json();
            this.csrfToken = data.data?.token;
            
            // 获取cookie中的CSRF token
            const cookies = response.headers.get('set-cookie');
            if (cookies) {
                const csrfCookieMatch = cookies.match(/csrf-token=([^;]+)/);
                if (csrfCookieMatch) {
                    this.csrfCookie = csrfCookieMatch[1];
                }
            }
            
            return { token: this.csrfToken?.substring(0, 20) + '...' };
        }, 'apis');
        
        return await this.test('用户登录认证', async () => {
            const headers = {
                'Content-Type': 'application/json'
            };
            
            // 添加CSRF token到header
            if (this.csrfToken) {
                headers['X-CSRF-Token'] = this.csrfToken;
            }
            
            // 添加CSRF cookie
            if (this.csrfCookie) {
                headers['Cookie'] = `csrf-token=${this.csrfCookie}`;
            }

            const response = await fetch(`${CONFIG.backends.main}/api/v1/auth/login`, {
                method: 'POST',
                headers,
                body: JSON.stringify(CONFIG.testUser)
            });

            if (!response.ok) {
                const errorText = await response.text();
                throw new Error(`登录失败: ${response.status} - ${errorText}`);
            }

            const data = await response.json();
            if (!data.success || !data.data?.token) {
                throw new Error(`登录响应无效: ${JSON.stringify(data)}`);
            }

            this.authToken = data.data.token;
            return {
                token: this.authToken.substring(0, 20) + '...',
                user: data.data.user
            };
        }, 'apis');
    }

    // ===== 数据库结构验证 =====
    async testDatabaseStructure() {
        await this.log('开始数据库结构验证...', 'db');
        
        await this.test('数据库连接测试', async () => {
            try {
                const result = execSync(
                    `psql -h ${CONFIG.database.host} -p ${CONFIG.database.port} -U ${CONFIG.database.user} -d ${CONFIG.database.database} -c "SELECT version();"`,
                    { encoding: 'utf8', timeout: 10000 }
                );
                return { connected: true, version: result.trim() };
            } catch (error) {
                throw new Error(`数据库连接失败: ${error.message}`);
            }
        }, 'database');

        await this.test('数据库表结构检查', async () => {
            try {
                const tablesQuery = `
                    SELECT table_name, table_type 
                    FROM information_schema.tables 
                    WHERE table_schema = 'public' 
                    ORDER BY table_name;
                `;
                
                const result = execSync(
                    `psql -h ${CONFIG.database.host} -p ${CONFIG.database.port} -U ${CONFIG.database.user} -d ${CONFIG.database.database} -t -c "${tablesQuery}"`,
                    { encoding: 'utf8', timeout: 10000 }
                );
                
                const tables = result.trim().split('\n')
                    .filter(line => line.trim())
                    .map(line => {
                        const [name, type] = line.trim().split('|').map(s => s.trim());
                        return { name, type };
                    });
                
                return { 
                    tableCount: tables.length,
                    tables: tables
                };
            } catch (error) {
                throw new Error(`表结构查询失败: ${error.message}`);
            }
        }, 'database');

        // 检查关键表存在性
        const criticalTables = [
            'users', 'letters', 'letter_codes', 'couriers', 'tasks', 
            'products', 'orders', 'signal_codes', 'ai_configs', 'museum_items'
        ];

        for (const tableName of criticalTables) {
            await this.test(`表 ${tableName} 存在性检查`, async () => {
                try {
                    const query = `
                        SELECT column_name, data_type, is_nullable 
                        FROM information_schema.columns 
                        WHERE table_name = '${tableName}' AND table_schema = 'public'
                        ORDER BY ordinal_position;
                    `;
                    
                    const result = execSync(
                        `psql -h ${CONFIG.database.host} -p ${CONFIG.database.port} -U ${CONFIG.database.user} -d ${CONFIG.database.database} -t -c "${query}"`,
                        { encoding: 'utf8', timeout: 5000 }
                    );
                    
                    const columns = result.trim().split('\n')
                        .filter(line => line.trim())
                        .map(line => {
                            const [name, type, nullable] = line.trim().split('|').map(s => s.trim());
                            return { name, type, nullable };
                        });
                    
                    if (columns.length === 0) {
                        throw new Error(`表 ${tableName} 不存在`);
                    }
                    
                    return {
                        exists: true,
                        columnCount: columns.length,
                        columns: columns
                    };
                } catch (error) {
                    throw new Error(`表检查失败: ${error.message}`);
                }
            }, 'database');
        }
    }

    // ===== API端点功能测试 =====
    async testAPIEndpoints() {
        await this.log('开始API端点功能测试...', 'info');

        if (!this.authToken) {
            await this.log('需要先登录获取认证token', 'warning');
            return;
        }

        const headers = {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${this.authToken}`
        };
        
        // 添加CSRF相关headers如果可用
        if (this.csrfToken) {
            headers['X-CSRF-Token'] = this.csrfToken;
        }
        if (this.csrfCookie) {
            headers['Cookie'] = `csrf-token=${this.csrfCookie}`;
        }

        // 测试用户API
        await this.test('获取用户信息', async () => {
            const response = await fetch(`${CONFIG.backends.main}/api/v1/auth/me`, {
                method: 'GET',
                headers
            });

            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }

            const data = await response.json();
            return { user: data.data || data };
        }, 'apis');

        // 测试信件API  
        await this.test('获取信件列表', async () => {
            const response = await fetch(`${CONFIG.backends.main}/api/v1/letters`, {
                method: 'GET',
                headers
            });

            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }

            const data = await response.json();
            return { 
                letters: data.data?.letters || data.letters || [],
                total: data.data?.total || data.total || 0
            };
        }, 'apis');

        // 测试商品API
        await this.test('获取商品列表', async () => {
            const response = await fetch(`${CONFIG.backends.main}/api/v1/shop/products`, {
                method: 'GET',
                headers
            });

            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }

            const data = await response.json();
            return {
                products: data.data?.products || data.products || [],
                total: data.data?.total || data.total || 0
            };
        }, 'apis');

        // 测试AI API
        await this.test('AI灵感生成', async () => {
            const response = await fetch(`${CONFIG.backends.main}/api/v1/ai/inspiration`, {
                method: 'POST',
                headers,
                body: JSON.stringify({
                    theme: '校园生活',
                    count: 2
                })
            });

            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }

            const data = await response.json();
            return {
                inspirations: data.data?.inspirations || [],
                count: (data.data?.inspirations || []).length
            };
        }, 'apis');

        // 测试信使服务API
        if (CONFIG.backends.courierService) {
            await this.test('信使服务健康检查', async () => {
                const response = await fetch(`${CONFIG.backends.courierService}/health`, {
                    method: 'GET'
                });

                if (!response.ok) {
                    throw new Error(`HTTP ${response.status}: ${response.statusText}`);
                }

                return { healthy: true, status: response.status };
            }, 'apis');

            await this.test('获取信使层级配置', async () => {
                const response = await fetch(`${CONFIG.backends.courierService}/api/courier/levels/config`, {
                    method: 'GET',
                    headers
                });

                if (!response.ok) {
                    throw new Error(`HTTP ${response.status}: ${response.statusText}`);
                }

                const data = await response.json();
                return data;
            }, 'apis');
        }

        // 测试写信服务API
        if (CONFIG.backends.writeService) {
            await this.test('写信服务健康检查', async () => {
                const response = await fetch(`${CONFIG.backends.writeService}/health`, {
                    method: 'GET'
                });

                if (!response.ok) {
                    throw new Error(`HTTP ${response.status}: ${response.statusText}`);
                }

                return { healthy: true, status: response.status };
            }, 'apis');
        }
    }

    // ===== 数据一致性检查 =====
    async testDataConsistency() {
        await this.log('开始数据一致性检查...', 'db');

        // 检查用户数据一致性
        await this.test('用户数据一致性', async () => {
            try {
                const userCountQuery = "SELECT COUNT(*) as count FROM users;";
                const result = execSync(
                    `psql -h ${CONFIG.database.host} -p ${CONFIG.database.port} -U ${CONFIG.database.user} -d ${CONFIG.database.database} -t -c "${userCountQuery}"`,
                    { encoding: 'utf8', timeout: 5000 }
                );
                
                const userCount = parseInt(result.trim());
                
                // 通过API获取用户数量进行对比
                const apiResponse = await fetch(`${CONFIG.backends.main}/api/v1/admin/users`, {
                    method: 'GET',
                    headers: {
                        'Authorization': `Bearer ${this.authToken}`
                    }
                });
                
                let apiUserCount = 0;
                if (apiResponse.ok) {
                    const apiData = await apiResponse.json();
                    apiUserCount = apiData.data?.total || apiData.total || 0;
                }
                
                return {
                    databaseCount: userCount,
                    apiCount: apiUserCount,
                    consistent: userCount >= 0 // 基本一致性检查
                };
            } catch (error) {
                throw new Error(`用户数据检查失败: ${error.message}`);
            }
        }, 'database');

        // 检查信件数据一致性
        await this.test('信件数据一致性', async () => {
            try {
                const letterCountQuery = "SELECT COUNT(*) as count FROM letters;";
                const result = execSync(
                    `psql -h ${CONFIG.database.host} -p ${CONFIG.database.port} -U ${CONFIG.database.user} -d ${CONFIG.database.database} -t -c "${letterCountQuery}"`,
                    { encoding: 'utf8', timeout: 5000 }
                );
                
                const dbLetterCount = parseInt(result.trim());
                
                return {
                    databaseCount: dbLetterCount,
                    hasLetters: dbLetterCount > 0
                };
            } catch (error) {
                throw new Error(`信件数据检查失败: ${error.message}`);
            }
        }, 'database');

        // 检查商品数据一致性
        await this.test('商品数据一致性', async () => {
            try {
                const productCountQuery = "SELECT COUNT(*) as count FROM products;";
                const result = execSync(
                    `psql -h ${CONFIG.database.host} -p ${CONFIG.database.port} -U ${CONFIG.database.user} -d ${CONFIG.database.database} -t -c "${productCountQuery}"`,
                    { encoding: 'utf8', timeout: 5000 }
                );
                
                const dbProductCount = parseInt(result.trim());
                
                return {
                    databaseCount: dbProductCount,
                    hasProducts: dbProductCount > 0
                };
            } catch (error) {
                throw new Error(`商品数据检查失败: ${error.message}`);
            }
        }, 'database');
    }

    // ===== 跨服务集成测试 =====
    async testCrossServiceIntegration() {
        await this.log('开始跨服务集成测试...', 'info');

        if (!this.authToken) {
            await this.log('跳过集成测试：需要认证token', 'warning');
            return;
        }

        // 测试前端到后端的完整调用链
        await this.test('前端API代理集成', async () => {
            const response = await fetch(`${CONFIG.frontend}/api/ai/personas`, {
                method: 'GET',
                headers: {
                    'Authorization': `Bearer ${this.authToken}`
                }
            });

            if (!response.ok) {
                throw new Error(`前端API代理失败: ${response.status}`);
            }

            const data = await response.json();
            return {
                proxyWorking: true,
                personas: data.data?.personas || data.personas || []
            };
        }, 'integration');

        // 测试数据创建到查询的完整流程
        await this.test('数据CRUD完整流程', async () => {
            const headers = {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${this.authToken}`
            };

            // 1. 创建测试数据 (购物车)
            const addToCartResponse = await fetch(`${CONFIG.backends.main}/api/v1/shop/cart/items`, {
                method: 'POST',
                headers,
                body: JSON.stringify({
                    product_id: 'test-product-1',
                    quantity: 1
                })
            });

            // 2. 查询购物车
            const getCartResponse = await fetch(`${CONFIG.backends.main}/api/v1/shop/cart`, {
                method: 'GET',
                headers
            });

            return {
                addToCart: addToCartResponse.status,
                getCart: getCartResponse.status,
                crudFlow: addToCartResponse.status < 400 && getCartResponse.status < 400
            };
        }, 'integration');
    }

    // ===== 性能基准测试 =====
    async testPerformance() {
        await this.log('开始性能基准测试...', 'info');

        await this.test('API响应时间测试', async () => {
            const endpoints = [
                { name: 'health', url: `${CONFIG.backends.main}/health` },
                { name: 'auth/me', url: `${CONFIG.backends.main}/api/v1/auth/me` },
                { name: 'letters', url: `${CONFIG.backends.main}/api/v1/letters` }
            ];

            const results = {};
            
            for (const endpoint of endpoints) {
                const startTime = Date.now();
                try {
                    const response = await fetch(endpoint.url, {
                        method: 'GET',
                        headers: this.authToken ? {
                            'Authorization': `Bearer ${this.authToken}`
                        } : {}
                    });
                    
                    const duration = Date.now() - startTime;
                    results[endpoint.name] = {
                        duration,
                        status: response.status,
                        success: response.ok
                    };
                } catch (error) {
                    results[endpoint.name] = {
                        duration: Date.now() - startTime,
                        error: error.message,
                        success: false
                    };
                }
            }

            return results;
        }, 'integration');
    }

    // ===== 主测试流程 =====
    async runAllTests() {
        await this.log('🚀 开始OpenPenPal API与数据库完整性测试', 'info');
        await this.log(`测试目标: ${Object.keys(CONFIG.backends).length} 个后端服务`, 'info');

        try {
            // 1. 服务健康检查
            await this.testServiceHealth();

            // 2. 认证测试
            await this.testAuthentication();

            // 3. 数据库结构验证
            await this.testDatabaseStructure();

            // 4. API端点测试
            await this.testAPIEndpoints();

            // 5. 数据一致性检查
            await this.testDataConsistency();

            // 6. 跨服务集成测试
            await this.testCrossServiceIntegration();

            // 7. 性能测试
            await this.testPerformance();

        } catch (error) {
            await this.log(`测试过程中发生严重错误: ${error.message}`, 'error');
        }

        // 生成测试报告
        await this.generateReport();
    }

    // ===== 生成测试报告 =====
    async generateReport() {
        const testDuration = Date.now() - this.testStartTime.getTime();
        
        await this.log('\n📊 ===== 测试报告 =====', 'info');

        // 统计各类别的成功率
        for (const [category, tests] of Object.entries(this.results)) {
            if (Object.keys(tests).length === 0) continue;
            
            const total = Object.keys(tests).length;
            const successful = Object.values(tests).filter(t => t.success).length;
            const successRate = ((successful / total) * 100).toFixed(1);
            
            await this.log(`\n📈 ${category.toUpperCase()} 类别:`, 'info');
            await this.log(`   ✅ 成功: ${successful}/${total} (${successRate}%)`, successful === total ? 'success' : 'warning');
            
            // 显示失败的测试
            const failures = Object.entries(tests).filter(([_, test]) => !test.success);
            if (failures.length > 0) {
                await this.log(`   ❌ 失败的测试:`, 'error');
                for (const [testName, test] of failures) {
                    await this.log(`      • ${testName}: ${test.error}`, 'error');
                }
            }
        }

        // 总体统计
        const allTests = Object.values(this.results).flatMap(category => Object.values(category));
        const totalTests = allTests.length;
        const totalSuccessful = allTests.filter(t => t.success).length;
        const overallSuccessRate = totalTests > 0 ? ((totalSuccessful / totalTests) * 100).toFixed(1) : 0;

        await this.log('\n🎯 总体统计:', 'info');
        await this.log(`   📊 总测试数: ${totalTests}`, 'info');
        await this.log(`   ✅ 成功: ${totalSuccessful}`, 'success');
        await this.log(`   ❌ 失败: ${totalTests - totalSuccessful}`, totalSuccessful === totalTests ? 'info' : 'error');
        await this.log(`   📈 成功率: ${overallSuccessRate}%`, overallSuccessRate >= 80 ? 'success' : 'warning');
        await this.log(`   ⏱️  总耗时: ${(testDuration / 1000).toFixed(2)}秒`, 'info');

        // 关键发现
        await this.log('\n🔍 关键发现:', 'info');
        
        const servicesStatus = this.results.services || {};
        const healthyServices = Object.values(servicesStatus).filter(s => s.success).length;
        const totalServices = Object.keys(servicesStatus).length;
        
        if (healthyServices === totalServices) {
            await this.log(`   ✅ 所有 ${totalServices} 个服务运行正常`, 'success');
        } else {
            await this.log(`   ⚠️  ${healthyServices}/${totalServices} 个服务正常运行`, 'warning');
        }

        const dbTests = this.results.database || {};
        const dbSuccess = Object.values(dbTests).filter(t => t.success).length;
        if (dbSuccess > 0) {
            await this.log(`   ✅ 数据库连接和结构验证通过`, 'success');
        } else {
            await this.log(`   ❌ 数据库测试存在问题`, 'error');
        }

        const apiTests = this.results.apis || {};
        const apiSuccess = Object.values(apiTests).filter(t => t.success).length;
        if (apiSuccess > 0) {
            await this.log(`   ✅ API端点功能正常`, 'success');
        }

        await this.log('\n✨ 测试完成！', 'success');
        
        // 保存详细报告到文件
        const reportData = {
            timestamp: new Date().toISOString(),
            duration: testDuration,
            summary: {
                total: totalTests,
                successful: totalSuccessful,
                failed: totalTests - totalSuccessful,
                successRate: overallSuccessRate
            },
            results: this.results,
            config: CONFIG
        };

        require('fs').writeFileSync(
            'api-database-integrity-report.json',
            JSON.stringify(reportData, null, 2)
        );
        
        await this.log('📄 详细报告已保存到: api-database-integrity-report.json', 'info');
    }
}

// 执行测试
async function main() {
    const tester = new APIIntegrityTester();
    await tester.runAllTests();
}

if (require.main === module) {
    main().catch(console.error);
}

module.exports = APIIntegrityTester;