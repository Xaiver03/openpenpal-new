#!/usr/bin/env node

/**
 * 认证修复测试脚本
 * Test script for authentication fixes
 */

const axios = require('axios');

const FRONTEND_URL = 'http://localhost:3000';
const API_URL = 'http://localhost:8080/api/v1';

// 测试用户凭据
const TEST_CREDENTIALS = {
  username: 'admin',
  password: 'admin123'
};

async function testLogin() {
  console.log('🔐 测试登录流程...');
  
  try {
    // 1. 获取CSRF token
    console.log('1. 获取CSRF token...');
    const csrfResponse = await axios.get(`${FRONTEND_URL}/api/auth/csrf`);
    console.log('✅ CSRF token获取成功');
    
    // 2. 登录
    console.log('2. 执行登录...');
    const loginResponse = await axios.post(`${FRONTEND_URL}/api/auth/login`, TEST_CREDENTIALS);
    
    if (loginResponse.data.success) {
      console.log('✅ 登录成功');
      console.log('Token:', loginResponse.data.data.token?.substring(0, 50) + '...');
      return loginResponse.data.data.token;
    } else {
      console.log('❌ 登录失败:', loginResponse.data.message);
      return null;
    }
  } catch (error) {
    console.log('❌ 登录过程出错:', error.message);
    return null;
  }
}

async function testAuthMe(token) {
  console.log('👤 测试获取用户信息...');
  
  try {
    const response = await axios.get(`${FRONTEND_URL}/api/auth/me`, {
      headers: { 'Authorization': `Bearer ${token}` }
    });
    
    if (response.data.success) {
      console.log('✅ 获取用户信息成功');
      console.log('用户:', response.data.data.username, '角色:', response.data.data.role);
      console.log('缓存头:', response.headers['x-cache'] || 'No cache header');
      return true;
    } else {
      console.log('❌ 获取用户信息失败:', response.data.message);
      return false;
    }
  } catch (error) {
    console.log('❌ 获取用户信息出错:', error.message);
    return false;
  }
}

async function testMiddlewareProtection() {
  console.log('🛡️ 测试中间件路由保护...');
  
  const protectedRoutes = ['/ai', '/write', '/courier', '/admin'];
  
  for (const route of protectedRoutes) {
    try {
      console.log(`测试路由: ${route}`);
      const response = await axios.get(`${FRONTEND_URL}${route}`, {
        maxRedirects: 0,
        validateStatus: (status) => status < 400 || status === 302
      });
      
      if (response.status === 302) {
        const location = response.headers.location;
        if (location && location.includes('/login')) {
          console.log(`✅ ${route} 正确重定向到登录页`);
        } else {
          console.log(`⚠️ ${route} 重定向到: ${location}`);
        }
      } else if (response.status === 200) {
        console.log(`⚠️ ${route} 允许未认证访问`);
      }
    } catch (error) {
      if (error.response && error.response.status === 302) {
        console.log(`✅ ${route} 正确重定向到登录页`);
      } else {
        console.log(`❌ ${route} 测试出错:`, error.message);
      }
    }
  }
}

async function testAuthenticatedAccess(token) {
  console.log('🔓 测试认证后的路由访问...');
  
  const protectedRoutes = ['/ai', '/write'];
  
  for (const route of protectedRoutes) {
    try {
      console.log(`测试已认证访问: ${route}`);
      const response = await axios.get(`${FRONTEND_URL}${route}`, {
        headers: { 
          'Authorization': `Bearer ${token}`,
          'Cookie': `openpenpal_auth_token=${token}`
        },
        maxRedirects: 0,
        validateStatus: (status) => status < 400 || status === 302
      });
      
      if (response.status === 200) {
        console.log(`✅ ${route} 认证访问成功`);
      } else if (response.status === 302) {
        console.log(`❌ ${route} 仍然重定向到: ${response.headers.location}`);
      }
    } catch (error) {
      if (error.response) {
        console.log(`❌ ${route} 访问失败 (${error.response.status}):`, error.response.headers.location || error.message);
      } else {
        console.log(`❌ ${route} 访问出错:`, error.message);
      }
    }
  }
}

async function testCachePerformance(token) {
  console.log('⚡ 测试缓存性能...');
  
  const times = [];
  const iterations = 10;
  
  for (let i = 0; i < iterations; i++) {
    const start = Date.now();
    try {
      await axios.get(`${FRONTEND_URL}/api/auth/me`, {
        headers: { 'Authorization': `Bearer ${token}` }
      });
      times.push(Date.now() - start);
    } catch (error) {
      console.log(`请求 ${i + 1} 失败:`, error.message);
    }
  }
  
  if (times.length > 0) {
    const avg = times.reduce((a, b) => a + b, 0) / times.length;
    const min = Math.min(...times);
    const max = Math.max(...times);
    
    console.log(`✅ 缓存性能测试完成:`);
    console.log(`  平均响应时间: ${avg.toFixed(2)}ms`);
    console.log(`  最快响应: ${min}ms`);
    console.log(`  最慢响应: ${max}ms`);
    console.log(`  成功请求: ${times.length}/${iterations}`);
  }
}

async function runAllTests() {
  console.log('🚀 开始认证修复测试...\n');
  
  // 1. 测试中间件保护（未认证）
  await testMiddlewareProtection();
  console.log('');
  
  // 2. 测试登录
  const token = await testLogin();
  console.log('');
  
  if (!token) {
    console.log('❌ 无法获取token，终止测试');
    return;
  }
  
  // 3. 测试用户信息获取
  const authMeSuccess = await testAuthMe(token);
  console.log('');
  
  if (!authMeSuccess) {
    console.log('❌ 无法获取用户信息，终止测试');
    return;
  }
  
  // 4. 测试认证后的路由访问
  await testAuthenticatedAccess(token);
  console.log('');
  
  // 5. 测试缓存性能
  await testCachePerformance(token);
  console.log('');
  
  console.log('✅ 所有测试完成！');
  console.log('\n💡 调试建议:');
  console.log('1. 打开浏览器开发者工具查看认证状态');
  console.log('2. 在控制台运行 AuthStateFixer.generateDiagnosticReport() 查看详细状态');
  console.log('3. 使用右下角的 🔧 Auth Debug 按钮打开调试面板');
  console.log('4. 如果仍有问题，运行 AuthStateFixer.autoFix() 自动修复');
}

// 运行测试
runAllTests().catch(console.error);