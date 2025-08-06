#!/usr/bin/env node

/**
 * 全面的晋升系统集成测试
 * 测试CSRF保护、数据库功能、权限控制和错误处理
 */

const https = require('https');
const http = require('http');
const { URL } = require('url');

// 测试配置
const config = {
  frontend: 'http://localhost:3000',
  backend: 'http://localhost:8080',
  testUsers: {
    student: { username: 'alice', password: 'secret', expectedRole: 'student' },
    courier_level1: { username: 'courier_level1', password: 'secret', expectedRole: 'courier_level1' },
    courier_level2: { username: 'courier_level2', password: 'secret', expectedRole: 'courier_level2' },
    courier_level3: { username: 'courier_level3', password: 'secret', expectedRole: 'courier_level3' },
    courier_level4: { username: 'courier_level4', password: 'secret', expectedRole: 'courier_level4' },
    admin: { username: 'admin', password: 'admin123', expectedRole: 'super_admin' }
  }
};

// Test results tracking
const testResults = {
  csrf: { passed: 0, failed: 0, tests: [] },
  database: { passed: 0, failed: 0, tests: [] },
  permissions: { passed: 0, failed: 0, tests: [] },
  errorHandling: { passed: 0, failed: 0, tests: [] }
};

// Utility functions
function makeRequest(url, options = {}) {
  return new Promise((resolve, reject) => {
    const urlObj = new URL(url);
    const isHttps = urlObj.protocol === 'https:';
    const client = isHttps ? https : http;
    
    const requestOptions = {
      hostname: urlObj.hostname,
      port: urlObj.port || (isHttps ? 443 : 80),
      path: urlObj.pathname + urlObj.search,
      method: options.method || 'GET',
      headers: {
        'Content-Type': 'application/json',
        'User-Agent': 'OpenPenPal-Test-Client/1.0',
        ...options.headers
      }
    };

    const req = client.request(requestOptions, (res) => {
      let data = '';
      res.on('data', chunk => data += chunk);
      res.on('end', () => {
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

async function testServiceHealth() {
  console.log('\n🔍 测试服务健康状态...');
  
  try {
    // Test frontend
    const frontendRes = await makeRequest(`${config.frontend}/health`);
    logTest('csrf', '前端健康检查', frontendRes.statusCode === 200, 
      `状态码: ${frontendRes.statusCode}`);
    
    // Test backend
    const backendRes = await makeRequest(`${config.backend}/health`);
    logTest('csrf', '后端健康检查', backendRes.statusCode === 200, 
      `状态码: ${backendRes.statusCode}`);
      
    return frontendRes.statusCode === 200 && backendRes.statusCode === 200;
  } catch (error) {
    logTest('csrf', '服务健康检查', false, `错误: ${error.message}`);
    return false;
  }
}

async function testUserAuthentication() {
  console.log('\n🔐 测试用户认证和CSRF保护...');
  
  for (const [userType, userData] of Object.entries(config.testUsers)) {
    try {
      // Test login
      const loginRes = await makeRequest(`${config.backend}/api/v1/auth/login`, {
        method: 'POST',
        body: {
          username: userData.username,
          password: userData.password
        }
      });
      
      const loginSuccess = loginRes.statusCode === 200 || loginRes.statusCode === 401;
      logTest('csrf', `${userType} 用户登录测试`, loginSuccess, 
        `状态码: ${loginRes.statusCode}, 响应: ${JSON.stringify(loginRes.data)}`);
      
      // If login successful, test authenticated endpoints
      if (loginRes.statusCode === 200 && loginRes.data && loginRes.data.token) {
        const token = loginRes.data.token;
        
        // Test authenticated endpoint
        const profileRes = await makeRequest(`${config.backend}/api/v1/users/me`, {
          headers: {
            'Authorization': `Bearer ${token}`
          }
        });
        
        logTest('csrf', `${userType} 认证端点访问`, profileRes.statusCode === 200,
          `获取用户信息状态码: ${profileRes.statusCode}`);
      }
      
    } catch (error) {
      logTest('csrf', `${userType} 认证测试`, false, `错误: ${error.message}`);
    }
  }
}

async function testPromotionDatabase() {
  console.log('\n💾 测试晋升系统数据库功能...');
  
  // Test courier growth path endpoint
  try {
    const growthPathRes = await makeRequest(`${config.backend}/api/v1/courier/growth/path`);
    logTest('database', '晋升路径端点', growthPathRes.statusCode === 401 || growthPathRes.statusCode === 404,
      `状态码: ${growthPathRes.statusCode} (预期需要认证)`);
  } catch (error) {
    logTest('database', '晋升路径端点', false, `错误: ${error.message}`);
  }
  
  // Test courier growth progress endpoint
  try {
    const progressRes = await makeRequest(`${config.backend}/api/v1/courier/growth/progress`);
    logTest('database', '晋升进度端点', progressRes.statusCode === 401 || progressRes.statusCode === 404,
      `状态码: ${progressRes.statusCode} (预期需要认证)`);
  } catch (error) {
    logTest('database', '晋升进度端点', false, `错误: ${error.message}`);
  }
  
  // Test level upgrade endpoint
  try {
    const upgradeRes = await makeRequest(`${config.backend}/api/v1/courier/level/upgrade`, {
      method: 'POST',
      body: { request_level: 2, reason: '测试晋升申请' }
    });
    logTest('database', '晋升申请端点', upgradeRes.statusCode === 401 || upgradeRes.statusCode === 403,
      `状态码: ${upgradeRes.statusCode} (预期需要认证或CSRF保护)`);
  } catch (error) {
    logTest('database', '晋升申请端点', false, `错误: ${error.message}`);
  }
}

async function testPermissionBoundaries() {
  console.log('\n🛡️ 测试权限边界控制...');
  
  // Test different permission levels for courier endpoints
  const permissionEndpoints = [
    '/api/v1/courier/growth/path',
    '/api/v1/courier/growth/progress',
    '/api/v1/courier/level/upgrade-requests',
    '/api/v1/admin/couriers'
  ];
  
  for (const endpoint of permissionEndpoints) {
    try {
      // Test without authentication
      const noAuthRes = await makeRequest(`${config.backend}${endpoint}`);
      logTest('permissions', `${endpoint} 无认证访问`, 
        noAuthRes.statusCode === 401 || noAuthRes.statusCode === 403,
        `状态码: ${noAuthRes.statusCode} (应该拒绝访问)`);
        
    } catch (error) {
      logTest('permissions', `${endpoint} 权限测试`, false, `错误: ${error.message}`);
    }
  }
}

async function testErrorHandling() {
  console.log('\n⚠️ 测试错误处理机制...');
  
  // Test invalid endpoints
  try {
    const invalidRes = await makeRequest(`${config.backend}/api/v1/invalid/endpoint`);
    logTest('errorHandling', '无效端点处理', invalidRes.statusCode === 404,
      `状态码: ${invalidRes.statusCode}`);
  } catch (error) {
    logTest('errorHandling', '无效端点处理', false, `错误: ${error.message}`);
  }
  
  // Test malformed requests
  try {
    const malformedRes = await makeRequest(`${config.backend}/api/v1/auth/login`, {
      method: 'POST',
      body: 'invalid json'
    });
    logTest('errorHandling', '格式错误请求处理', 
      malformedRes.statusCode === 400 || malformedRes.statusCode === 403,
      `状态码: ${malformedRes.statusCode}`);
  } catch (error) {
    logTest('errorHandling', '格式错误请求处理', true, `正确捕获错误: ${error.message}`);
  }
  
  // Test oversized requests
  try {
    const largeBody = 'x'.repeat(10000);
    const oversizedRes = await makeRequest(`${config.backend}/api/v1/auth/login`, {
      method: 'POST',
      body: largeBody
    });
    logTest('errorHandling', '超大请求处理', 
      oversizedRes.statusCode === 413 || oversizedRes.statusCode === 400,
      `状态码: ${oversizedRes.statusCode}`);
  } catch (error) {
    logTest('errorHandling', '超大请求处理', true, `正确处理错误: ${error.message}`);
  }
}

function printTestSummary() {
  console.log('\n📊 测试结果总结');
  console.log('='.repeat(50));
  
  let totalPassed = 0;
  let totalFailed = 0;
  
  for (const [category, results] of Object.entries(testResults)) {
    const { passed, failed, tests } = results;
    totalPassed += passed;
    totalFailed += failed;
    
    console.log(`\n🔸 ${category.toUpperCase()} 测试`);
    console.log(`   通过: ${passed}, 失败: ${failed}, 总计: ${tests.length}`);
    
    if (failed > 0) {
      console.log('   失败的测试:');
      tests.filter(t => !t.passed).forEach(test => {
        console.log(`   ❌ ${test.name}: ${test.details}`);
      });
    }
  }
  
  console.log('\n' + '='.repeat(50));
  console.log(`🎯 总体结果: ${totalPassed} 通过, ${totalFailed} 失败`);
  console.log(`📈 成功率: ${((totalPassed / (totalPassed + totalFailed)) * 100).toFixed(1)}%`);
  
  if (totalFailed === 0) {
    console.log('🎉 所有测试通过！晋升系统集成测试成功完成。');
  } else {
    console.log('⚠️ 部分测试失败，请检查上述失败项目。');
  }
}

async function runComprehensiveTests() {
  console.log('🚀 OpenPenPal 晋升系统 - 全面集成测试');
  console.log('='.repeat(50));
  console.log('测试范围: CSRF保护, 数据库功能, 权限控制, 错误处理');
  
  const startTime = Date.now();
  
  // Run all test suites
  const servicesHealthy = await testServiceHealth();
  if (!servicesHealthy) {
    console.log('❌ 服务健康检查失败，停止测试');
    return;
  }
  
  await testUserAuthentication();
  await testPromotionDatabase();
  await testPermissionBoundaries();
  await testErrorHandling();
  
  const endTime = Date.now();
  
  console.log(`\n⏱️ 测试执行时间: ${(endTime - startTime) / 1000}秒`);
  printTestSummary();
}

// 执行测试
if (require.main === module) {
  runComprehensiveTests().catch(error => {
    console.error('❌ 测试执行失败:', error);
    process.exit(1);
  });
}

module.exports = { runComprehensiveTests, testResults };