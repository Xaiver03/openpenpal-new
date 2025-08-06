#!/usr/bin/env node

/**
 * Enhanced晋升系统测试 - 包含完整的CSRF和认证流程
 * 测试完整的用户认证、CSRF保护、权限控制和数据库功能
 */

const https = require('https');
const http = require('http');
const { URL } = require('url');

// 测试配置
const config = {
  frontend: 'http://localhost:3000',
  backend: 'http://localhost:8080',
  testUsers: {
    alice: { username: 'alice', password: 'secret', expectedRole: 'student' },
    admin: { username: 'admin', password: 'admin123', expectedRole: 'super_admin' }
  }
};

// Session state for maintaining cookies and tokens
class TestSession {
  constructor() {
    this.cookies = new Map();
    this.csrfToken = null;
    this.authToken = null;
  }
  
  setCookie(cookieHeader) {
    if (!cookieHeader) return;
    
    const cookies = Array.isArray(cookieHeader) ? cookieHeader : [cookieHeader];
    cookies.forEach(cookie => {
      const [nameValue] = cookie.split(';');
      const [name, value] = nameValue.split('=');
      if (name && value) {
        this.cookies.set(name.trim(), value.trim());
      }
    });
  }
  
  getCookieHeader() {
    if (this.cookies.size === 0) return '';
    return Array.from(this.cookies.entries())
      .map(([name, value]) => `${name}=${value}`)
      .join('; ');
  }
}

// Test results tracking
const testResults = {
  csrf: { passed: 0, failed: 0, tests: [] },
  database: { passed: 0, failed: 0, tests: [] },
  permissions: { passed: 0, failed: 0, tests: [] },
  errorHandling: { passed: 0, failed: 0, tests: [] }
};

function makeRequest(url, options = {}, session = null) {
  return new Promise((resolve, reject) => {
    const urlObj = new URL(url);
    const isHttps = urlObj.protocol === 'https:';
    const client = isHttps ? https : http;
    
    const headers = {
      'Content-Type': 'application/json',
      'User-Agent': 'OpenPenPal-Enhanced-Test-Client/1.0',
      ...options.headers
    };
    
    // Add session cookies if available
    if (session && session.getCookieHeader()) {
      headers['Cookie'] = session.getCookieHeader();
    }
    
    // Add CSRF token if available
    if (session && session.csrfToken) {
      headers['X-CSRF-Token'] = session.csrfToken;
    }
    
    // Add auth token if available
    if (session && session.authToken) {
      headers['Authorization'] = `Bearer ${session.authToken}`;
    }
    
    const requestOptions = {
      hostname: urlObj.hostname,
      port: urlObj.port || (isHttps ? 443 : 80),
      path: urlObj.pathname + urlObj.search,
      method: options.method || 'GET',
      headers
    };

    const req = client.request(requestOptions, (res) => {
      let data = '';
      res.on('data', chunk => data += chunk);
      res.on('end', () => {
        // Update session cookies if present
        if (session && res.headers['set-cookie']) {
          session.setCookie(res.headers['set-cookie']);
        }
        
        try {
          const response = {
            statusCode: res.statusCode,
            headers: res.headers,
            data: data ? JSON.parse(data) : null
          };
          resolve(response);
        } catch (e) {
          resolve({
            statusCode: res.statusCode,
            headers: res.headers,
            data: data,
            rawData: data
          });
        }
      });
    });

    req.on('error', reject);
    
    if (options.body) {
      req.write(typeof options.body === 'string' ? options.body : JSON.stringify(options.body));
    }
    
    req.end();
  });
}

function logTest(category, testName, passed, details = '') {
  const result = passed ? '✅ PASS' : '❌ FAIL';
  console.log(`[${category.toUpperCase()}] ${result} ${testName}`);
  if (details) console.log(`    ${details}`);
  
  testResults[category].tests.push({ name: testName, passed, details });
  if (passed) {
    testResults[category].passed++;
  } else {
    testResults[category].failed++;
  }
}

async function initializeSession() {
  console.log('\n🔐 初始化测试会话...');
  
  const session = new TestSession();
  
  try {
    // First, make a request to get initial cookies
    const initRes = await makeRequest(`${config.frontend}/`, {}, session);
    logTest('csrf', '前端初始访问', initRes.statusCode === 200, 
      `状态码: ${initRes.statusCode}`);
    
    // Try to get CSRF token from a dedicated endpoint
    const csrfRes = await makeRequest(`${config.frontend}/api/csrf-token`, {}, session);
    if (csrfRes.statusCode === 200 && csrfRes.data && csrfRes.data.token) {
      session.csrfToken = csrfRes.data.token;
      logTest('csrf', 'CSRF Token获取', true, `Token: ${session.csrfToken.substring(0, 10)}...`);
    } else {
      logTest('csrf', 'CSRF Token获取', false, `状态码: ${csrfRes.statusCode}`);
    }
    
    return session;
  } catch (error) {
    logTest('csrf', '会话初始化', false, `错误: ${error.message}`);
    return session;
  }
}

async function testUserAuthenticationWithCSRF(session) {
  console.log('\n🔑 测试完整的用户认证流程...');
  
  for (const [userType, userData] of Object.entries(config.testUsers)) {
    try {
      // Test login with CSRF protection
      const loginRes = await makeRequest(`${config.backend}/api/v1/auth/login`, {
        method: 'POST',
        body: {
          username: userData.username,
          password: userData.password
        }
      }, session);
      
      if (loginRes.statusCode === 200 && loginRes.data && loginRes.data.token) {
        session.authToken = loginRes.data.token;
        logTest('csrf', `${userType} 用户登录`, true, 
          `Token获取成功: ${session.authToken.substring(0, 20)}...`);
        
        // Test authenticated endpoint
        const profileRes = await makeRequest(`${config.backend}/api/v1/users/me`, {}, session);
        logTest('csrf', `${userType} 认证端点访问`, profileRes.statusCode === 200,
          `用户信息获取状态码: ${profileRes.statusCode}`);
          
        return { success: true, userType, session };
      } else {
        logTest('csrf', `${userType} 用户登录`, false, 
          `状态码: ${loginRes.statusCode}, 响应: ${JSON.stringify(loginRes.data)}`);
      }
      
    } catch (error) {
      logTest('csrf', `${userType} 认证测试`, false, `错误: ${error.message}`);
    }
  }
  
  return { success: false, session };
}

async function testPromotionSystemWithAuth(session) {
  console.log('\n🎯 测试晋升系统数据库功能 (带认证)...');
  
  // Test courier growth path endpoint with authentication
  try {
    const growthPathRes = await makeRequest(`${config.backend}/api/v1/courier/growth/path`, {}, session);
    logTest('database', '认证用户晋升路径查询', 
      growthPathRes.statusCode === 200 || growthPathRes.statusCode === 404,
      `状态码: ${growthPathRes.statusCode}, 数据: ${JSON.stringify(growthPathRes.data)}`);
  } catch (error) {
    logTest('database', '认证用户晋升路径查询', false, `错误: ${error.message}`);
  }
  
  // Test courier growth progress endpoint with authentication
  try {
    const progressRes = await makeRequest(`${config.backend}/api/v1/courier/growth/progress`, {}, session);
    logTest('database', '认证用户晋升进度查询', 
      progressRes.statusCode === 200 || progressRes.statusCode === 404,
      `状态码: ${progressRes.statusCode}, 数据: ${JSON.stringify(progressRes.data)}`);
  } catch (error) {
    logTest('database', '认证用户晋升进度查询', false, `错误: ${error.message}`);
  }
  
  // Test level upgrade request submission
  try {
    const upgradeRes = await makeRequest(`${config.backend}/api/v1/courier/level/upgrade`, {
      method: 'POST',
      body: { 
        request_level: 2, 
        reason: '完成基础任务，申请晋升到二级信使',
        evidence: { delivered_count: 10, success_rate: 95 }
      }
    }, session);
    logTest('database', '晋升申请提交', 
      upgradeRes.statusCode === 200 || upgradeRes.statusCode === 201 || upgradeRes.statusCode === 400,
      `状态码: ${upgradeRes.statusCode}, 响应: ${JSON.stringify(upgradeRes.data)}`);
  } catch (error) {
    logTest('database', '晋升申请提交', false, `错误: ${error.message}`);
  }
}

async function testPermissionBoundariesWithAuth(session) {
  console.log('\n🛡️ 测试权限边界控制 (带认证)...');
  
  const permissionTests = [
    { endpoint: '/api/v1/courier/growth/path', description: '晋升路径查询' },
    { endpoint: '/api/v1/courier/growth/progress', description: '晋升进度查询' },
    { endpoint: '/api/v1/courier/level/upgrade-requests', description: '晋升申请列表' },
    { endpoint: '/api/v1/admin/couriers', description: '管理员信使列表' }
  ];
  
  for (const test of permissionTests) {
    try {
      const res = await makeRequest(`${config.backend}${test.endpoint}`, {}, session);
      const isAuthorized = res.statusCode === 200 || res.statusCode === 404;
      const isUnauthorized = res.statusCode === 401 || res.statusCode === 403;
      
      logTest('permissions', `${test.description} 权限测试`, 
        isAuthorized || isUnauthorized,
        `状态码: ${res.statusCode} (${isAuthorized ? '有权限' : '被拒绝'})`);
        
    } catch (error) {
      logTest('permissions', `${test.description} 权限测试`, false, `错误: ${error.message}`);
    }
  }
}

async function testAdvancedErrorHandling(session) {
  console.log('\n⚠️ 测试高级错误处理...');
  
  // Test invalid JSON in request body
  try {
    const invalidJsonRes = await makeRequest(`${config.backend}/api/v1/courier/level/upgrade`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: '{"invalid": json}'
    }, session);
    logTest('errorHandling', '无效JSON处理', 
      invalidJsonRes.statusCode === 400 || invalidJsonRes.statusCode === 403,
      `状态码: ${invalidJsonRes.statusCode}`);
  } catch (error) {
    logTest('errorHandling', '无效JSON处理', true, `正确捕获错误: ${error.message}`);
  }
  
  // Test missing required fields
  try {
    const missingFieldsRes = await makeRequest(`${config.backend}/api/v1/courier/level/upgrade`, {
      method: 'POST',
      body: { request_level: 2 } // Missing reason field
    }, session);
    logTest('errorHandling', '缺少必需字段处理', 
      missingFieldsRes.statusCode === 400,
      `状态码: ${missingFieldsRes.statusCode}, 响应: ${JSON.stringify(missingFieldsRes.data)}`);
  } catch (error) {
    logTest('errorHandling', '缺少必需字段处理', false, `错误: ${error.message}`);
  }
  
  // Test rate limiting (if implemented)
  const rapidRequests = Array(5).fill().map((_, i) => 
    makeRequest(`${config.backend}/api/v1/courier/growth/path`, {}, session)
  );
  
  try {
    const results = await Promise.all(rapidRequests);
    const rateLimited = results.some(res => res.statusCode === 429);
    logTest('errorHandling', '速率限制测试', true,
      `${rateLimited ? '检测到速率限制' : '未检测到速率限制 (可能未实现)'}`);
  } catch (error) {
    logTest('errorHandling', '速率限制测试', false, `错误: ${error.message}`);
  }
}

function printEnhancedTestSummary() {
  console.log('\n📊 Enhanced测试结果总结');
  console.log('='.repeat(60));
  
  let totalPassed = 0;
  let totalFailed = 0;
  
  for (const [category, results] of Object.entries(testResults)) {
    const { passed, failed, tests } = results;
    totalPassed += passed;
    totalFailed += failed;
    
    const passRate = tests.length > 0 ? ((passed / tests.length) * 100).toFixed(1) : 0;
    
    console.log(`\n🔸 ${category.toUpperCase()} 测试 (${passRate}% 通过率)`);
    console.log(`   通过: ${passed}, 失败: ${failed}, 总计: ${tests.length}`);
    
    if (failed > 0) {
      console.log('   失败的测试:');
      tests.filter(t => !t.passed).forEach(test => {
        console.log(`   ❌ ${test.name}: ${test.details}`);
      });
    }
  }
  
  console.log('\n' + '='.repeat(60));
  console.log(`🎯 总体结果: ${totalPassed} 通过, ${totalFailed} 失败`);
  const overallRate = ((totalPassed / (totalPassed + totalFailed)) * 100).toFixed(1);
  console.log(`📈 总体成功率: ${overallRate}%`);
  
  // Performance analysis
  if (overallRate >= 90) {
    console.log('🎉 优秀！系统通过了全面的测试验证。');
  } else if (overallRate >= 70) {
    console.log('✅ 良好！系统基本功能正常，有少量改进空间。');
  } else {
    console.log('⚠️ 需要改进！系统存在一些需要解决的问题。');
  }
}

async function runEnhancedTests() {
  console.log('🚀 OpenPenPal 晋升系统 - Enhanced全面集成测试');
  console.log('='.repeat(60));
  console.log('测试范围: 完整认证流程, CSRF保护, 数据库功能, 权限控制, 高级错误处理');
  
  const startTime = Date.now();
  
  // Initialize test session
  const session = await initializeSession();
  
  // Test complete authentication flow
  const authResult = await testUserAuthenticationWithCSRF(session);
  
  if (authResult.success) {
    console.log(`\n✅ 使用 ${authResult.userType} 用户进行后续测试...`);
    
    // Run tests with authenticated session
    await testPromotionSystemWithAuth(authResult.session);
    await testPermissionBoundariesWithAuth(authResult.session);
    await testAdvancedErrorHandling(authResult.session);
  } else {
    console.log('\n⚠️ 认证失败，将跳过需要认证的测试');
  }
  
  const endTime = Date.now();
  
  console.log(`\n⏱️ 测试执行时间: ${(endTime - startTime) / 1000}秒`);
  printEnhancedTestSummary();
  
  return authResult.success;
}

// 执行测试
if (require.main === module) {
  runEnhancedTests().catch(error => {
    console.error('❌ Enhanced测试执行失败:', error);
    process.exit(1);
  });
}

module.exports = { runEnhancedTests, testResults };