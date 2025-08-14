const axios = require('axios');

async function debugFrontendLogin() {
  console.log('🧪 Debugging Frontend Login...\n');
  
  try {
    // Step 1: Get CSRF token from frontend
    console.log('📍 Step 1: Getting CSRF token from frontend...');
    const csrfResponse = await axios.get('http://localhost:3000/api/auth/csrf', {
      headers: {
        'Accept': 'application/json',
        'User-Agent': 'test-script'
      }
    });
    
    const csrfToken = csrfResponse.data.data.token;
    const cookies = csrfResponse.headers['set-cookie'];
    
    console.log(`   ✅ CSRF token: ${csrfToken.substring(0, 16)}...`);
    console.log(`   🍪 Cookies: ${cookies ? 'Set' : 'Not set'}`);
    
    // Step 2: Prepare login request
    console.log('\n📍 Step 2: Preparing login request...');
    const loginData = {
      username: 'admin',
      password: 'password'
    };
    
    const headers = {
      'Content-Type': 'application/json',
      'X-CSRF-Token': csrfToken,
      'Accept': 'application/json',
      'User-Agent': 'test-script'
    };
    
    if (cookies) {
      headers['Cookie'] = cookies.join('; ');
    }
    
    console.log('   📤 Request headers:', Object.keys(headers));
    console.log('   📤 Request body:', loginData);
    
    // Step 3: Send login request
    console.log('\n📍 Step 3: Sending login request to frontend proxy...');
    const loginResponse = await axios.post('http://localhost:3000/api/auth/login', 
      loginData,
      { 
        headers,
        validateStatus: (status) => true // Don't throw on error status
      }
    );
    
    console.log(`   📥 Response status: ${loginResponse.status}`);
    console.log(`   📥 Response headers:`, Object.keys(loginResponse.headers));
    
    if (loginResponse.status === 500) {
      console.log('\n   ❌ Got 500 error. Response data:');
      console.log(JSON.stringify(loginResponse.data, null, 2));
      
      // Try to understand what's happening
      console.log('\n📍 Debugging: Testing direct gateway call...');
      const directResponse = await axios.post('http://localhost:8000/api/v1/auth/login',
        loginData,
        { 
          headers: { 'Content-Type': 'application/json' },
          validateStatus: (status) => true
        }
      );
      
      console.log(`   📥 Direct gateway response: ${directResponse.status}`);
      if (directResponse.status === 200) {
        console.log('   ✅ Gateway is working fine. Issue is in frontend proxy.');
      }
    } else if (loginResponse.status === 200) {
      console.log('\n   ✅ Login successful!');
      console.log(`   📧 User: ${loginResponse.data.data.user.username}`);
    }
    
  } catch (error) {
    console.error('\n❌ Unexpected error:', error.message);
    if (error.response) {
      console.log('Response data:', error.response.data);
    }
  }
}

// Also check environment variables
console.log('📋 Environment Check:');
console.log('   NODE_ENV:', process.env.NODE_ENV);
console.log('   NEXT_PUBLIC_API_URL:', process.env.NEXT_PUBLIC_API_URL);
console.log('   NEXT_PUBLIC_GATEWAY_URL:', process.env.NEXT_PUBLIC_GATEWAY_URL);
console.log('   NEXT_PUBLIC_BACKEND_URL:', process.env.NEXT_PUBLIC_BACKEND_URL);
console.log('');

debugFrontendLogin();