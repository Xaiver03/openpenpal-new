#!/usr/bin/env node

const axios = require('axios');

// 🔐 安全令牌生成 - 替代硬编码令牌
const { generateTestToken } = require('../../backend/scripts/test-token-generator');

// 测试配置
const API_BASE_URL = 'http://localhost:8080/api/v1';
const FRONTEND_API_URL = 'http://localhost:3000/api';
const TEST_TOKEN = generateTestToken('ADMIN', {}, '2h');

async function measurePerformance(name, testFunc) {
  console.log(`\n=== Testing: ${name} ===`);
  
  const times = [];
  const iterations = 100;
  let errors = 0;
  
  // 预热
  for (let i = 0; i < 5; i++) {
    try {
      await testFunc();
    } catch (error) {
      // 忽略预热错误
    }
  }
  
  // 正式测试
  for (let i = 0; i < iterations; i++) {
    const start = Date.now();
    try {
      await testFunc();
      times.push(Date.now() - start);
    } catch (error) {
      errors++;
      console.error(`Error in iteration ${i + 1}:`, error.message);
    }
  }
  
  if (times.length > 0) {
    const avg = times.reduce((a, b) => a + b, 0) / times.length;
    const min = Math.min(...times);
    const max = Math.max(...times);
    const p95 = times.sort((a, b) => a - b)[Math.floor(times.length * 0.95)];
    
    console.log(`Average: ${avg.toFixed(2)}ms`);
    console.log(`Min: ${min}ms`);
    console.log(`Max: ${max}ms`);
    console.log(`P95: ${p95}ms`);
    console.log(`Success rate: ${((times.length / iterations) * 100).toFixed(2)}%`);
  } else {
    console.log('All requests failed!');
  }
}

async function testBackendAuth() {
  const response = await axios.get(`${API_BASE_URL}/users/me`, {
    headers: {
      'Authorization': `Bearer ${TEST_TOKEN}`
    }
  });
  return response.data;
}

async function testFrontendAuth() {
  const response = await axios.get(`${FRONTEND_API_URL}/auth/me`, {
    headers: {
      'Authorization': `Bearer ${TEST_TOKEN}`
    }
  });
  return response.data;
}

async function testBackendPublicEndpoint() {
  const response = await axios.get(`${API_BASE_URL}/health`);
  return response.data;
}

async function testBackendWithHeavyAuth() {
  const promises = [];
  for (let i = 0; i < 10; i++) {
    promises.push(axios.get(`${API_BASE_URL}/users/me`, {
      headers: {
        'Authorization': `Bearer ${TEST_TOKEN}`
      }
    }));
  }
  await Promise.all(promises);
}

async function main() {
  console.log('🚀 Starting Middleware Performance Tests...\n');
  
  console.log('Configuration:');
  console.log(`Backend URL: ${API_BASE_URL}`);
  console.log(`Frontend URL: ${FRONTEND_API_URL}`);
  console.log(`Test iterations: 100 per test`);
  
  // 测试后端认证中间件
  await measurePerformance('Backend Auth Middleware (/users/me)', testBackendAuth);
  
  // 测试前端认证中间件（带缓存）
  await measurePerformance('Frontend Auth Middleware (/auth/me)', testFrontendAuth);
  
  // 测试无认证的公开端点
  await measurePerformance('Public Endpoint (/health)', testBackendPublicEndpoint);
  
  // 测试并发认证请求
  await measurePerformance('Concurrent Auth Requests (10x parallel)', testBackendWithHeavyAuth);
  
  console.log('\n✅ Performance tests completed!');
  
  // 测试缓存命中率
  console.log('\n=== Cache Hit Rate Test ===');
  console.log('Making 50 identical requests to test cache...');
  
  const cacheStart = Date.now();
  for (let i = 0; i < 50; i++) {
    try {
      const response = await axios.get(`${FRONTEND_API_URL}/auth/me`, {
        headers: {
          'Authorization': `Bearer ${TEST_TOKEN}`
        }
      });
      if (i === 49) {
        console.log('Cache Hit Rate:', response.headers['x-cache-hit-rate'] || 'Not available');
        console.log('Last request cached:', response.headers['x-auth-cached'] || 'Not available');
      }
    } catch (error) {
      // Ignore errors
    }
  }
  console.log(`Total time for 50 requests: ${Date.now() - cacheStart}ms`);
}

// 运行测试
main().catch(console.error);