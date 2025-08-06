const axios = require('axios');

const API_BASE_URL = 'http://localhost:8080';

async function testSingleLogin(username, password) {
  console.log(`\n${'='.repeat(60)}`);
  console.log(`Testing: ${username} / ${password}`);
  console.log('='.repeat(60));
  
  try {
    console.log('1. Sending POST request to:', `${API_BASE_URL}/api/v1/auth/login`);
    console.log('2. Request body:', JSON.stringify({ username, password }));
    
    const startTime = Date.now();
    const response = await axios.post(`${API_BASE_URL}/api/v1/auth/login`, {
      username,
      password
    }, {
      validateStatus: () => true, // Accept all status codes
      timeout: 5000
    });
    const endTime = Date.now();
    
    console.log(`3. Response received in ${endTime - startTime}ms`);
    console.log('4. Status:', response.status);
    console.log('5. Status Text:', response.statusText);
    console.log('6. Response Data:', JSON.stringify(response.data, null, 2));
    
    if (response.status === 200) {
      console.log('\n✅ LOGIN SUCCESSFUL');
      if (response.data.data && response.data.data.token) {
        console.log('Token received (first 50 chars):', response.data.data.token.substring(0, 50) + '...');
      }
    } else {
      console.log('\n❌ LOGIN FAILED');
    }
    
  } catch (error) {
    console.error('\n🔥 REQUEST ERROR:', error.message);
    if (error.code) {
      console.error('Error code:', error.code);
    }
  }
}

async function main() {
  console.log('OpenPenPal Login Testing - Step by Step');
  console.log('Time:', new Date().toISOString());
  
  // Test each user one by one with delay
  await testSingleLogin('admin', 'admin123');
  await new Promise(r => setTimeout(r, 2000));
  
  await testSingleLogin('alice', 'secret');
  await new Promise(r => setTimeout(r, 2000));
  
  await testSingleLogin('courier_level1', 'secret');
}

main().catch(console.error);