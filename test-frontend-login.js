const axios = require('axios');

async function testFrontendLogin() {
  console.log('🧪 Testing frontend login flow...\n');
  
  try {
    // Step 1: Get CSRF token from frontend API
    console.log('📍 Step 1: Getting CSRF token from frontend...');
    const csrfResponse = await axios.get('http://localhost:3000/api/auth/csrf');
    const csrfToken = csrfResponse.data.data.token;
    const cookies = csrfResponse.headers['set-cookie'];
    
    console.log(`   ✅ CSRF token obtained: ${csrfToken.substring(0, 16)}...`);
    console.log(`   🍪 Cookie: ${cookies ? 'Set' : 'Not set'}`);
    
    // Step 2: Login via frontend API
    console.log('\n📍 Step 2: Logging in as courier_level1...');
    const loginResponse = await axios.post('http://localhost:3000/api/auth/login', {
      username: 'courier_level1',
      password: 'password'
    }, {
      headers: {
        'X-CSRF-Token': csrfToken,
        'Cookie': cookies ? cookies.join('; ') : ''
      }
    });
    
    console.log(`   ✅ Login successful!`);
    console.log(`   📧 User: ${loginResponse.data.data.user.username}`);
    console.log(`   👤 Role: ${loginResponse.data.data.user.role}`);
    console.log(`   🔑 Access Token: ${loginResponse.data.data.accessToken.substring(0, 20)}...`);
    
    // Step 3: Test authenticated request
    console.log('\n📍 Step 3: Testing authenticated request...');
    const meResponse = await axios.get('http://localhost:3000/api/auth/me', {
      headers: {
        'Authorization': `Bearer ${loginResponse.data.data.accessToken}`
      }
    });
    
    console.log(`   ✅ Auth verified!`);
    console.log(`   📧 Current user: ${meResponse.data.data.username}`);
    
  } catch (error) {
    console.error('❌ Error:', error.response?.status, error.response?.data || error.message);
    if (error.response?.data) {
      console.log('Response data:', JSON.stringify(error.response.data, null, 2));
    }
  }
}

// Also test backend directly
async function testBackendLogin() {
  console.log('\n\n🧪 Testing backend login directly...\n');
  
  try {
    // Step 1: Get CSRF from backend via gateway
    console.log('📍 Testing auth endpoints via gateway (port 8000)...');
    const gatewayCSRF = await axios.get('http://localhost:8000/api/v1/auth/csrf')
      .catch(e => console.log(`   ❌ Gateway CSRF: ${e.response?.status || e.message}`));
    
    // Step 2: Try backend directly
    console.log('\n📍 Testing auth endpoints on backend directly (port 8080)...');
    const backendCSRF = await axios.get('http://localhost:8080/api/v1/auth/csrf')
      .catch(e => console.log(`   ❌ Backend CSRF: ${e.response?.status || e.message}`));
    
    // Step 3: Check actual backend routes
    console.log('\n📍 Checking backend health...');
    const health = await axios.get('http://localhost:8080/health');
    console.log(`   ✅ Backend health: ${health.data.status}`);
    
  } catch (error) {
    console.error('❌ Backend test error:', error.message);
  }
}

// Run tests
(async () => {
  await testFrontendLogin();
  await testBackendLogin();
})();