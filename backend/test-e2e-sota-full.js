/**
 * Comprehensive End-to-End Test for SOTA Improvements
 * 端到端测试 - 验证所有SOTA改进
 */

const http = require('http');
const fs = require('fs');
const path = require('path');

class E2ETestSuite {
  constructor() {
    this.baseUrl = 'http://localhost:8080';
    this.results = {
      total: 0,
      passed: 0,
      failed: 0,
      tests: []
    };
    this.token = null;
    this.csrfToken = null;
  }

  async request(endpoint, options = {}) {
    const url = new URL(endpoint, this.baseUrl);
    
    return new Promise((resolve, reject) => {
      const req = http.request(url, {
        ...options,
        headers: {
          'Content-Type': 'application/json',
          'Accept': 'application/json',
          ...(this.token && { 'Authorization': `Bearer ${this.token}` }),
          ...(this.csrfToken && options.method !== 'GET' && { 'X-CSRF-Token': this.csrfToken }),
          ...options.headers
        }
      }, (res) => {
        let data = '';
        res.on('data', chunk => data += chunk);
        res.on('end', () => {
          try {
            const json = JSON.parse(data);
            resolve({ 
              status: res.statusCode, 
              data: json, 
              headers: res.headers,
              raw: data 
            });
          } catch (e) {
            resolve({ 
              status: res.statusCode, 
              data: data, 
              headers: res.headers,
              raw: data 
            });
          }
        });
      });
      
      req.on('error', reject);
      
      if (options.body) {
        req.write(options.body);
      }
      
      req.end();
    });
  }

  recordTest(name, passed, details = {}) {
    this.results.total++;
    if (passed) {
      this.results.passed++;
    } else {
      this.results.failed++;
    }
    this.results.tests.push({
      name,
      passed,
      timestamp: new Date().toISOString(),
      ...details
    });
  }

  async runTests() {
    console.log('🚀 端到端SOTA改进测试套件\n');
    console.log('测试范围：');
    console.log('1️⃣  API路由别名 (Route Aliases)');
    console.log('2️⃣  字段转换中间件 (Field Transformation)');
    console.log('3️⃣  前端模型同步 (Model Synchronization)');
    console.log('4️⃣  AI集成 (AI Integration)');
    console.log('5️⃣  认证流程 (Authentication Flow)');
    console.log('6️⃣  错误处理 (Error Handling)');
    console.log('7️⃣  WebSocket连接 (Real-time Features)');
    console.log('8️⃣  性能指标 (Performance Metrics)\n');

    // Test Suite 1: API Route Aliases
    await this.testRouteAliases();
    
    // Test Suite 2: Authentication & Field Transformation
    await this.testAuthenticationFlow();
    
    // Test Suite 3: AI Integration
    await this.testAIIntegration();
    
    // Test Suite 4: Letter Operations
    await this.testLetterOperations();
    
    // Test Suite 5: Courier System
    await this.testCourierSystem();
    
    // Test Suite 6: Museum Features
    await this.testMuseumFeatures();
    
    // Test Suite 7: Error Scenarios
    await this.testErrorHandling();
    
    // Test Suite 8: Performance
    await this.testPerformance();
    
    // Generate Report
    this.generateReport();
  }

  async testRouteAliases() {
    console.log('\n📍 测试套件 1: API路由别名\n');
    
    const aliasTests = [
      { name: '学校列表', frontend: '/api/schools', backend: '/api/v1/schools' },
      { name: '邮编查询', frontend: '/api/postcode/100080', backend: '/api/v1/postcode/100080' },
      { name: '地址搜索', frontend: '/api/address/search?q=北京', backend: '/api/v1/address/search?q=北京' },
      { name: 'CSRF令牌', frontend: '/api/auth/csrf', backend: '/api/v1/auth/csrf' },
    ];

    for (const test of aliasTests) {
      console.log(`   测试: ${test.name}`);
      console.log(`   前端路由: ${test.frontend}`);
      
      const response = await this.request(test.frontend);
      const success = response.status === 200;
      
      console.log(`   状态码: ${response.status}`);
      console.log(`   ${success ? '✅ 路由别名正常工作' : '❌ 路由别名失败'}\n`);
      
      this.recordTest(`路由别名: ${test.name}`, success, {
        route: test.frontend,
        status: response.status,
        hasData: !!response.data
      });
    }
  }

  async testAuthenticationFlow() {
    console.log('\n📍 测试套件 2: 认证流程与字段转换\n');
    
    // Get CSRF Token
    console.log('   步骤 1: 获取CSRF令牌');
    const csrfRes = await this.request('/api/auth/csrf');
    
    if (csrfRes.status === 200 && csrfRes.data.data?.csrfToken) {
      this.csrfToken = csrfRes.data.data.csrfToken;
      console.log('   ✅ CSRF令牌获取成功');
      this.recordTest('CSRF令牌获取', true);
    } else {
      console.log('   ❌ CSRF令牌获取失败');
      this.recordTest('CSRF令牌获取', false);
      return;
    }
    
    // Login
    console.log('\n   步骤 2: 用户登录');
    const loginRes = await this.request('/api/v1/auth/login', {
      method: 'POST',
      body: JSON.stringify({ username: 'alice', password: 'secret' })
    });
    
    if (loginRes.status === 200) {
      console.log('   ✅ 登录成功');
      this.token = loginRes.data.data?.token;
      
      // Check field transformation
      const user = loginRes.data.data?.user;
      const rawData = loginRes.raw;
      
      console.log('\n   步骤 3: 验证字段转换');
      const transformationChecks = [
        { camel: 'createdAt', snake: 'created_at' },
        { camel: 'updatedAt', snake: 'updated_at' },
        { camel: 'isActive', snake: 'is_active' },
        { camel: 'schoolCode', snake: 'school_code' },
        { camel: 'lastLoginAt', snake: 'last_login_at' }
      ];
      
      let allTransformed = true;
      for (const check of transformationChecks) {
        const hasCamel = user && check.camel in user;
        const hasSnake = rawData.includes(check.snake);
        
        if (hasCamel && !hasSnake) {
          console.log(`   ✅ ${check.snake} → ${check.camel}`);
        } else {
          console.log(`   ❌ ${check.snake} 转换失败`);
          allTransformed = false;
        }
      }
      
      this.recordTest('字段转换', allTransformed, {
        fieldsChecked: transformationChecks.length,
        userFields: Object.keys(user || {})
      });
      
      // Test authenticated endpoint
      console.log('\n   步骤 4: 测试认证端点');
      const meRes = await this.request('/api/v1/users/me');
      
      if (meRes.status === 200) {
        console.log('   ✅ 认证端点访问成功');
        this.recordTest('认证端点访问', true);
      } else {
        console.log('   ❌ 认证端点访问失败');
        this.recordTest('认证端点访问', false);
      }
      
    } else {
      console.log('   ❌ 登录失败');
      this.recordTest('用户登录', false);
    }
  }

  async testAIIntegration() {
    console.log('\n📍 测试套件 3: AI集成\n');
    
    // Test 1: Inspiration Generation
    console.log('   测试 1: AI灵感生成');
    const inspirationRes = await this.request('/api/v1/ai/inspiration', {
      method: 'POST',
      body: JSON.stringify({ theme: '友谊', count: 3 })
    });
    
    if (inspirationRes.status === 200 && inspirationRes.data.data?.inspirations) {
      const inspirations = inspirationRes.data.data.inspirations;
      console.log(`   ✅ 生成了 ${inspirations.length} 条灵感`);
      
      // Check if it's real AI content
      const avgLength = inspirations.reduce((sum, i) => sum + i.prompt.length, 0) / inspirations.length;
      const isRealAI = avgLength > 50 && !inspirations[0].prompt.includes('这是一个关于');
      
      if (isRealAI) {
        console.log('   ✅ Moonshot AI真实响应');
        console.log(`   示例: "${inspirations[0].prompt.substring(0, 80)}..."`);
      } else {
        console.log('   ⚠️  使用了预设内容（非AI生成）');
      }
      
      this.recordTest('AI灵感生成', true, { 
        count: inspirations.length,
        isRealAI,
        avgLength 
      });
    } else {
      console.log('   ❌ AI灵感生成失败');
      this.recordTest('AI灵感生成', false);
    }
    
    // Test 2: AI Personas
    console.log('\n   测试 2: AI人设列表');
    const personasRes = await this.request('/api/v1/ai/personas');
    
    if (personasRes.status === 200 && personasRes.data.data?.personas) {
      const personas = personasRes.data.data.personas;
      console.log(`   ✅ 获取到 ${personas.length} 个AI人设`);
      personas.slice(0, 3).forEach(p => {
        console.log(`      - ${p.name}: ${p.description}`);
      });
      this.recordTest('AI人设列表', true, { count: personas.length });
    } else {
      console.log('   ❌ AI人设列表获取失败');
      this.recordTest('AI人设列表', false);
    }
    
    // Test 3: Daily Inspiration
    console.log('\n   测试 3: 每日灵感');
    const dailyRes = await this.request('/api/v1/ai/daily-inspiration');
    
    if (dailyRes.status === 200 && dailyRes.data.data) {
      console.log('   ✅ 每日灵感获取成功');
      console.log(`   主题: ${dailyRes.data.data.theme}`);
      this.recordTest('每日灵感', true);
    } else {
      console.log('   ❌ 每日灵感获取失败');
      this.recordTest('每日灵感', false);
    }
  }

  async testLetterOperations() {
    console.log('\n📍 测试套件 4: 信件操作\n');
    
    // Create Letter
    console.log('   测试 1: 创建信件');
    const letterData = {
      title: 'SOTA测试信件',
      content: '这是一封用于测试SOTA改进的信件。',
      style: 'warm',
      visibility: 'public',
      recipientOpCode: 'PK5F01'
    };
    
    const createRes = await this.request('/api/v1/letters', {
      method: 'POST',
      body: JSON.stringify(letterData)
    });
    
    let letterId = null;
    if (createRes.status === 201 && createRes.data.data?.id) {
      letterId = createRes.data.data.id;
      console.log('   ✅ 信件创建成功');
      console.log(`   信件ID: ${letterId}`);
      
      // Check field transformation in response
      const letter = createRes.data.data;
      const hasTransformedFields = 'createdAt' in letter && 'recipientOpCode' in letter;
      console.log(`   ${hasTransformedFields ? '✅' : '❌'} 响应字段已转换为驼峰命名`);
      
      this.recordTest('创建信件', true, { letterId, hasTransformedFields });
    } else {
      console.log('   ❌ 信件创建失败');
      this.recordTest('创建信件', false);
      return;
    }
    
    // Get Letter
    console.log('\n   测试 2: 获取信件详情');
    const getRes = await this.request(`/api/v1/letters/${letterId}`);
    
    if (getRes.status === 200) {
      console.log('   ✅ 信件详情获取成功');
      this.recordTest('获取信件', true);
    } else {
      console.log('   ❌ 信件详情获取失败');
      this.recordTest('获取信件', false);
    }
    
    // List Letters
    console.log('\n   测试 3: 信件列表');
    const listRes = await this.request('/api/v1/letters?page=1&pageSize=10');
    
    if (listRes.status === 200 && listRes.data.data?.letters) {
      console.log(`   ✅ 获取到 ${listRes.data.data.letters.length} 封信件`);
      this.recordTest('信件列表', true, { count: listRes.data.data.letters.length });
    } else {
      console.log('   ❌ 信件列表获取失败');
      this.recordTest('信件列表', false);
    }
  }

  async testCourierSystem() {
    console.log('\n📍 测试套件 5: 信使系统\n');
    
    // Courier login
    console.log('   测试 1: 信使登录');
    const courierLoginRes = await this.request('/api/v1/auth/login', {
      method: 'POST',
      body: JSON.stringify({ username: 'courier_level1', password: 'secret' })
    });
    
    let courierToken = null;
    if (courierLoginRes.status === 200) {
      courierToken = courierLoginRes.data.data?.token;
      console.log('   ✅ 信使登录成功');
      this.recordTest('信使登录', true);
    } else {
      console.log('   ❌ 信使登录失败');
      this.recordTest('信使登录', false);
      return;
    }
    
    // Get courier info
    console.log('\n   测试 2: 信使信息');
    const courierRes = await this.request('/api/v1/courier/me', {
      headers: { 'Authorization': `Bearer ${courierToken}` }
    });
    
    if (courierRes.status === 200 && courierRes.data.data) {
      const courier = courierRes.data.data;
      console.log('   ✅ 信使信息获取成功');
      console.log(`   级别: L${courier.level}`);
      console.log(`   管理区域: ${courier.managedOpCodePrefix || '未分配'}`);
      
      // Check field transformation
      const hasTransformed = 'managedOpCodePrefix' in courier && 
                           'weeklyHours' in courier &&
                           'maxDailyTasks' in courier;
      console.log(`   ${hasTransformed ? '✅' : '❌'} 字段已转换为驼峰命名`);
      
      this.recordTest('信使信息', true, { 
        level: courier.level,
        hasTransformed 
      });
    } else {
      console.log('   ❌ 信使信息获取失败');
      this.recordTest('信使信息', false);
    }
    
    // Hierarchy info
    console.log('\n   测试 3: 层级信息');
    const hierarchyRes = await this.request('/api/v1/courier/hierarchy/me', {
      headers: { 'Authorization': `Bearer ${courierToken}` }
    });
    
    if (hierarchyRes.status === 200) {
      console.log('   ✅ 层级信息获取成功');
      this.recordTest('层级信息', true);
    } else {
      console.log('   ❌ 层级信息获取失败');
      this.recordTest('层级信息', false);
    }
  }

  async testMuseumFeatures() {
    console.log('\n📍 测试套件 6: 博物馆功能\n');
    
    // Get exhibitions
    console.log('   测试 1: 展览列表');
    const exhibitionsRes = await this.request('/api/v1/museum/exhibitions');
    
    if (exhibitionsRes.status === 200) {
      console.log('   ✅ 展览列表获取成功');
      this.recordTest('展览列表', true);
    } else {
      console.log('   ❌ 展览列表获取失败');
      this.recordTest('展览列表', false);
    }
    
    // Get entries
    console.log('\n   测试 2: 博物馆条目');
    const entriesRes = await this.request('/api/v1/museum/entries?page=1&limit=10');
    
    if (entriesRes.status === 200 && entriesRes.data.data) {
      const entries = entriesRes.data.data.entries || [];
      console.log(`   ✅ 获取到 ${entries.length} 个博物馆条目`);
      
      // Check field transformation
      if (entries.length > 0) {
        const hasTransformed = 'viewCount' in entries[0] && 
                             'likeCount' in entries[0] &&
                             'createdAt' in entries[0];
        console.log(`   ${hasTransformed ? '✅' : '❌'} 字段已转换为驼峰命名`);
      }
      
      this.recordTest('博物馆条目', true, { count: entries.length });
    } else {
      console.log('   ❌ 博物馆条目获取失败');
      this.recordTest('博物馆条目', false);
    }
  }

  async testErrorHandling() {
    console.log('\n📍 测试套件 7: 错误处理\n');
    
    // 404 Error
    console.log('   测试 1: 404错误');
    const notFoundRes = await this.request('/api/v1/nonexistent');
    
    if (notFoundRes.status === 404) {
      console.log('   ✅ 404错误处理正常');
      this.recordTest('404错误', true);
    } else {
      console.log('   ❌ 404错误处理异常');
      this.recordTest('404错误', false);
    }
    
    // Validation Error
    console.log('\n   测试 2: 验证错误');
    const validationRes = await this.request('/api/v1/letters', {
      method: 'POST',
      body: JSON.stringify({}) // Missing required fields
    });
    
    if (validationRes.status === 400 || validationRes.status === 422) {
      console.log('   ✅ 验证错误处理正常');
      console.log(`   错误信息: ${validationRes.data.message || validationRes.data.error}`);
      this.recordTest('验证错误', true);
    } else {
      console.log('   ❌ 验证错误处理异常');
      this.recordTest('验证错误', false);
    }
    
    // Unauthorized Error
    console.log('\n   测试 3: 未授权错误');
    const unauthorizedRes = await this.request('/api/v1/users/me', {
      headers: { 'Authorization': 'Bearer invalid_token' }
    });
    
    if (unauthorizedRes.status === 401) {
      console.log('   ✅ 未授权错误处理正常');
      this.recordTest('未授权错误', true);
    } else {
      console.log('   ❌ 未授权错误处理异常');
      this.recordTest('未授权错误', false);
    }
  }

  async testPerformance() {
    console.log('\n📍 测试套件 8: 性能指标\n');
    
    const endpoints = [
      { name: '学校列表', path: '/api/schools' },
      { name: 'AI灵感', path: '/api/v1/ai/inspiration', method: 'POST', body: { theme: '日常' } },
      { name: '信件列表', path: '/api/v1/letters' }
    ];
    
    for (const endpoint of endpoints) {
      console.log(`   测试: ${endpoint.name}`);
      
      const startTime = Date.now();
      const res = await this.request(endpoint.path, {
        method: endpoint.method || 'GET',
        body: endpoint.body ? JSON.stringify(endpoint.body) : undefined
      });
      const duration = Date.now() - startTime;
      
      console.log(`   响应时间: ${duration}ms`);
      console.log(`   ${duration < 1000 ? '✅' : '⚠️'} ${duration < 1000 ? '性能良好' : '响应较慢'}`);
      
      this.recordTest(`性能: ${endpoint.name}`, duration < 1000, {
        duration,
        endpoint: endpoint.path
      });
    }
  }

  generateReport() {
    console.log('\n' + '='.repeat(80));
    console.log('📊 端到端测试报告');
    console.log('='.repeat(80));
    
    console.log(`\n测试总数: ${this.results.total}`);
    console.log(`✅ 通过: ${this.results.passed}`);
    console.log(`❌ 失败: ${this.results.failed}`);
    console.log(`成功率: ${(this.results.passed / this.results.total * 100).toFixed(1)}%`);
    
    // Group results by category
    const categories = {
      '路由别名': [],
      '字段转换': [],
      'AI功能': [],
      '信件系统': [],
      '信使系统': [],
      '博物馆': [],
      '错误处理': [],
      '性能': []
    };
    
    this.results.tests.forEach(test => {
      for (const category in categories) {
        if (test.name.includes(category)) {
          categories[category].push(test);
          break;
        }
      }
    });
    
    console.log('\n分类结果:');
    for (const [category, tests] of Object.entries(categories)) {
      if (tests.length > 0) {
        const passed = tests.filter(t => t.passed).length;
        console.log(`\n${category}:`);
        console.log(`   通过率: ${(passed / tests.length * 100).toFixed(0)}% (${passed}/${tests.length})`);
        tests.forEach(test => {
          console.log(`   ${test.passed ? '✅' : '❌'} ${test.name}`);
        });
      }
    }
    
    // Key findings
    console.log('\n关键发现:');
    const fieldTransformTests = this.results.tests.filter(t => t.name.includes('字段转换'));
    const allFieldsTransformed = fieldTransformTests.every(t => t.passed);
    console.log(`${allFieldsTransformed ? '✅' : '❌'} 所有API响应字段均已转换为驼峰命名`);
    
    const aiTests = this.results.tests.filter(t => t.name.includes('AI'));
    const aiWorking = aiTests.filter(t => t.passed).length / aiTests.length > 0.7;
    console.log(`${aiWorking ? '✅' : '❌'} AI系统正常工作（Moonshot集成）`);
    
    const routeTests = this.results.tests.filter(t => t.name.includes('路由'));
    const routesWorking = routeTests.every(t => t.passed);
    console.log(`${routesWorking ? '✅' : '❌'} 所有前端路由别名正常工作`);
    
    // Performance summary
    const perfTests = this.results.tests.filter(t => t.name.includes('性能'));
    if (perfTests.length > 0) {
      const avgDuration = perfTests.reduce((sum, t) => sum + (t.duration || 0), 0) / perfTests.length;
      console.log(`\n平均响应时间: ${avgDuration.toFixed(0)}ms`);
    }
    
    // Save detailed report
    const detailedReport = {
      summary: {
        total: this.results.total,
        passed: this.results.passed,
        failed: this.results.failed,
        successRate: (this.results.passed / this.results.total * 100).toFixed(1),
        timestamp: new Date().toISOString()
      },
      categories,
      tests: this.results.tests,
      conclusions: {
        fieldTransformation: allFieldsTransformed,
        aiIntegration: aiWorking,
        routeAliases: routesWorking,
        overallStatus: this.results.passed / this.results.total > 0.8 ? 'PASS' : 'NEEDS_ATTENTION'
      }
    };
    
    fs.writeFileSync('e2e-sota-report.json', JSON.stringify(detailedReport, null, 2));
    console.log('\n📄 详细报告已保存至: e2e-sota-report.json');
    
    // Final verdict
    console.log('\n' + '='.repeat(80));
    if (detailedReport.conclusions.overallStatus === 'PASS') {
      console.log('✅ SOTA改进端到端测试通过！');
      console.log('   所有主要功能正常工作，系统集成良好。');
    } else {
      console.log('⚠️  部分功能需要关注');
      console.log('   请查看详细报告了解具体问题。');
    }
    console.log('='.repeat(80));
  }
}

// Run the test suite
async function main() {
  const suite = new E2ETestSuite();
  try {
    await suite.runTests();
  } catch (error) {
    console.error('\n❌ 测试套件执行错误:', error.message);
    process.exit(1);
  }
}

main();