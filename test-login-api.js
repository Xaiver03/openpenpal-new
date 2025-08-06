const axios = require('axios');

const API_BASE_URL = 'http://localhost:8080';

async function testLogin(username, password) {
  console.log(`\nTesting login for user: ${username}`);
  console.log('=' .repeat(50));
  
  try {
    const response = await axios.post(`${API_BASE_URL}/api/v1/auth/login`, {
      username,
      password
    });
    
    console.log('✅ Login successful!');
    console.log('Status:', response.status);
    console.log('Response:', JSON.stringify(response.data, null, 2));
    
    if (response.data.data && response.data.data.access_token) {
      console.log('\n🔑 Access Token:', response.data.data.access_token);
      console.log('👤 User Info:', JSON.stringify(response.data.data.user, null, 2));
    }
    
    return response.data;
  } catch (error) {
    console.error('❌ Login failed!');
    console.error('Status:', error.response?.status);
    console.error('Error:', error.response?.data || error.message);
    return null;
  }
}

async function main() {
  console.log('Testing OpenPenPal Login API');
  console.log('=' .repeat(50));
  
  // Test cases
  const testCases = [
    { username: 'admin', password: 'admin123' },
    { username: 'alice', password: 'secret' },
    { username: 'courier_level1', password: 'secret' },
    { username: 'courier_level2', password: 'secret' },
    { username: 'courier_level3', password: 'secret' },
    { username: 'courier_level4', password: 'secret' },
    // Test with wrong password
    { username: 'admin', password: 'wrongpassword' },
  ];
  
  for (const testCase of testCases) {
    await testLogin(testCase.username, testCase.password);
    await new Promise(resolve => setTimeout(resolve, 1000)); // Wait 1 second between tests
  }
}

main().catch(console.error);