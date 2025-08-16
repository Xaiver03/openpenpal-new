#!/usr/bin/env node

/**
 * Test Authentication Fixes
 * 测试认证修复是否生效
 */

const TEST_USER = {
  username: 'admin',
  password: 'admin123'
};

const BASE_URL = 'http://localhost:8080';
const FRONTEND_URL = 'http://localhost:3001';

async function testBackendLogin() {
  console.log('🧪 Testing backend login...');
  
  try {
    const response = await fetch(`${BASE_URL}/api/v1/auth/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(TEST_USER),
    });
    
    const result = await response.json();
    console.log('✅ Backend login response:', response.status);
    
    if (response.ok && result.data?.token) {
      console.log('✅ JWT token received');
      console.log('✅ Expires at:', result.data.expires_at);
      return result.data.token;
    } else {
      console.log('❌ Backend login failed:', result);
      return null;
    }
  } catch (error) {
    console.log('❌ Backend login error:', error.message);
    return null;
  }
}

async function testFrontendAuthAPI(token) {
  console.log('🧪 Testing frontend auth API...');
  
  try {
    const response = await fetch(`${FRONTEND_URL}/api/auth/me`, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
    });
    
    const result = await response.json();
    console.log('✅ Frontend auth API response:', response.status);
    
    if (response.ok && result.data) {
      console.log('✅ User data received:', result.data.username);
      return true;
    } else {
      console.log('❌ Frontend auth API failed:', result);
      return false;
    }
  } catch (error) {
    console.log('❌ Frontend auth API error:', error.message);
    return false;
  }
}

async function testWebSocket(token) {
  console.log('🧪 Testing WebSocket connection...');
  
  return new Promise((resolve) => {
    try {
      const ws = new WebSocket(`ws://localhost:8080/api/v1/ws/connect?token=${token}`);
      
      ws.onopen = () => {
        console.log('✅ WebSocket connected successfully');
        ws.close();
        resolve(true);
      };
      
      ws.onerror = (error) => {
        console.log('❌ WebSocket connection failed:', error.message);
        resolve(false);
      };
      
      ws.onclose = (event) => {
        if (event.code !== 1000) {
          console.log('❌ WebSocket closed with error:', event.code, event.reason);
          resolve(false);
        }
      };
      
      // Timeout after 5 seconds
      setTimeout(() => {
        if (ws.readyState === WebSocket.CONNECTING) {
          console.log('❌ WebSocket connection timeout');
          ws.close();
          resolve(false);
        }
      }, 5000);
      
    } catch (error) {
      console.log('❌ WebSocket test error:', error.message);
      resolve(false);
    }
  });
}

async function testRateLimiting() {
  console.log('🧪 Testing rate limiting...');
  
  try {
    const promises = [];
    for (let i = 0; i < 5; i++) {
      promises.push(fetch(`${BASE_URL}/health`));
    }
    
    const responses = await Promise.all(promises);
    const successCount = responses.filter(r => r.ok).length;
    const rateLimitedCount = responses.filter(r => r.status === 429).length;
    
    console.log(`✅ Successful requests: ${successCount}/5`);
    console.log(`ℹ️ Rate limited requests: ${rateLimitedCount}/5`);
    
    return successCount >= 3; // At least 3 out of 5 should succeed
  } catch (error) {
    console.log('❌ Rate limiting test error:', error.message);
    return false;
  }
}

async function runAllTests() {
  console.log('🚀 Starting authentication fixes test suite...\n');
  
  // Test 1: Backend login
  const token = await testBackendLogin();
  if (!token) {
    console.log('❌ Cannot continue tests without valid token');
    return;
  }
  
  console.log('');
  
  // Test 2: Frontend auth API
  const authApiWorking = await testFrontendAuthAPI(token);
  
  console.log('');
  
  // Test 3: Rate limiting
  const rateLimitingOk = await testRateLimiting();
  
  console.log('');
  
  // Test 4: WebSocket (requires WebSocket support in Node.js)
  console.log('🧪 WebSocket test skipped (requires browser environment)');
  
  console.log('\n📊 Test Results Summary:');
  console.log(`✅ Backend Login: ${token ? 'PASS' : 'FAIL'}`);
  console.log(`✅ Frontend Auth API: ${authApiWorking ? 'PASS' : 'FAIL'}`);
  console.log(`✅ Rate Limiting: ${rateLimitingOk ? 'PASS' : 'FAIL'}`);
  console.log(`⏭️ WebSocket: MANUAL TEST REQUIRED`);
  
  const allPassed = token && authApiWorking && rateLimitingOk;
  console.log(`\n🎯 Overall Status: ${allPassed ? '✅ ALL CRITICAL TESTS PASSED' : '❌ SOME TESTS FAILED'}`);
  
  if (allPassed) {
    console.log('\n🎉 Authentication fixes appear to be working correctly!');
    console.log('💡 Please test the WebSocket connection manually in the browser.');
  } else {
    console.log('\n🔧 Some issues remain. Check the backend logs and ensure services are running.');
  }
}

// Check if fetch is available (Node.js 18+)
if (typeof fetch === 'undefined') {
  console.log('❌ This test requires Node.js 18+ or a fetch polyfill');
  console.log('💡 Please run: npm install node-fetch');
  process.exit(1);
}

// Run tests
runAllTests().catch(console.error);