const axios = require('axios');

const BACKEND_URL = 'http://localhost:8080';

async function testLogin(username, password) {
  console.log(`\nTesting: ${username} / ${password}`);
  console.log('-'.repeat(40));
  
  try {
    const response = await axios.post(`${BACKEND_URL}/api/v1/auth/login`, {
      username,
      password
    }, {
      validateStatus: () => true, // Accept all status codes
      timeout: 5000
    });
    
    console.log(`Status: ${response.status} ${response.statusText}`);
    
    if (response.status === 200) {
      console.log('✅ Login successful!');
      const data = response.data.data;
      console.log(`Token: ${data.token.substring(0, 50)}...`);
      console.log(`User: ${data.user.username} (${data.user.role})`);
    } else {
      console.log('❌ Login failed!');
      console.log(`Response: ${JSON.stringify(response.data)}`);
    }
    
    return response;
  } catch (error) {
    console.log('🔥 Request error:', error.message);
    return null;
  }
}

async function checkHealth() {
  try {
    const response = await axios.get(`${BACKEND_URL}/health`);
    console.log('Health check:', response.data);
  } catch (error) {
    console.log('Health check failed:', error.message);
  }
}

async function main() {
  console.log('Direct Backend Login Test');
  console.log('=' .repeat(60));
  
  // Check health first
  await checkHealth();
  
  // Test each user
  await testLogin('admin', 'admin123');
  await testLogin('alice', 'secret');
  await testLogin('courier_level1', 'secret');
  
  // Test with wrong password
  console.log('\n\nTesting with wrong password:');
  await testLogin('admin', 'wrongpassword');
}

main().catch(console.error);