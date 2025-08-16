const axios = require('axios');

async function testAuthEndpoints() {
  console.log('🧪 Testing Auth Endpoints...\n');
  
  // Test 1: Direct backend health check
  console.log('📍 Test 1: Backend Health Check');
  try {
    const health = await axios.get('http://localhost:8080/health');
    console.log('   ✅ Backend is healthy:', health.data.status);
  } catch (error) {
    console.log('   ❌ Backend health check failed:', error.message);
  }
  
  // Test 2: Backend auth endpoints
  console.log('\n📍 Test 2: Backend Auth Endpoints');
  const endpoints = [
    { url: 'http://localhost:8080/auth/csrf', desc: 'Backend CSRF (no /api/v1)' },
    { url: 'http://localhost:8080/api/auth/csrf', desc: 'Backend CSRF (/api)' },
    { url: 'http://localhost:8080/api/v1/auth/csrf', desc: 'Backend CSRF (full path)' }
  ];
  
  for (const endpoint of endpoints) {
    try {
      const response = await axios.get(endpoint.url);
      console.log(`   ✅ ${endpoint.desc}: ${response.status}`);
    } catch (error) {
      console.log(`   ❌ ${endpoint.desc}: ${error.response?.status || error.message}`);
    }
  }
  
  // Test 3: Gateway endpoints
  console.log('\n📍 Test 3: Gateway Auth Endpoints');
  try {
    const gatewayCSRF = await axios.get('http://localhost:8000/api/v1/auth/csrf');
    console.log(`   ✅ Gateway CSRF: ${gatewayCSRF.status}`);
  } catch (error) {
    console.log(`   ❌ Gateway CSRF: ${error.response?.status || error.message}`);
  }
  
  // Test 4: Frontend API proxy
  console.log('\n📍 Test 4: Frontend API Proxy');
  try {
    const frontendCSRF = await axios.get('http://localhost:3000/api/auth/csrf');
    console.log(`   ✅ Frontend CSRF: ${frontendCSRF.status}`);
    console.log(`   📝 CSRF Token: ${frontendCSRF.data.data.token.substring(0, 16)}...`);
  } catch (error) {
    console.log(`   ❌ Frontend CSRF: ${error.response?.status || error.message}`);
  }
  
  // Test 5: Try a complete login flow
  console.log('\n📍 Test 5: Complete Login Flow');
  try {
    // Get CSRF from frontend
    const csrfResponse = await axios.get('http://localhost:3000/api/auth/csrf');
    const csrfToken = csrfResponse.data.data.token;
    const cookies = csrfResponse.headers['set-cookie'];
    
    console.log('   ✅ Got CSRF token from frontend');
    
    // Try login
    const loginResponse = await axios.post('http://localhost:3000/api/auth/login', {
      username: 'admin',
      password: 'password'
    }, {
      headers: {
        'X-CSRF-Token': csrfToken,
        'Cookie': cookies ? cookies.join('; ') : ''
      }
    });
    
    console.log('   ✅ Login successful!');
    console.log(`   📧 User: ${loginResponse.data.data.user.username}`);
    
  } catch (error) {
    console.log('   ❌ Login failed:', error.response?.status, error.response?.data?.message || error.message);
    if (error.response?.data) {
      console.log('   📝 Error details:', JSON.stringify(error.response.data, null, 2));
    }
  }
  
  // Test 6: Check if gateway is running
  console.log('\n📍 Test 6: Service Status');
  const { exec } = require('child_process');
  
  exec('lsof -i :8000 | grep LISTEN', (error, stdout) => {
    if (stdout) {
      console.log('   ✅ Gateway is running on port 8000');
    } else {
      console.log('   ❌ Gateway is NOT running on port 8000');
    }
    
    exec('lsof -i :8080 | grep LISTEN', (error, stdout) => {
      if (stdout) {
        console.log('   ✅ Backend is running on port 8080');
      } else {
        console.log('   ❌ Backend is NOT running on port 8080');
      }
    });
  });
}

testAuthEndpoints();