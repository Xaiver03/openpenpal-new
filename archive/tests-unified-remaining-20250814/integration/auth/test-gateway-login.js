const axios = require('axios');

// 注意：使用网关端口 8000 而不是直接访问后端 8080
const GATEWAY_URL = 'http://localhost:8000';
const BACKEND_URL = 'http://localhost:8080';

async function testLogin(baseUrl, username, password) {
  try {
    console.log(`\nTesting ${username} via ${baseUrl}`);
    const response = await axios.post(`${baseUrl}/api/v1/auth/login`, {
      username,
      password
    }, {
      timeout: 10000
    });
    
    console.log(`✅ Login successful!`);
    console.log(`   Token: ${response.data.data.token.substring(0, 50)}...`);
    console.log(`   User: ${response.data.data.user.username} (${response.data.data.user.role})`);
    return true;
  } catch (error) {
    console.log(`❌ Login failed!`);
    console.log(`   Status: ${error.response?.status}`);
    console.log(`   Error: ${error.response?.data?.error || error.message}`);
    return false;
  }
}

async function main() {
  console.log('Testing Login via Gateway vs Direct Backend');
  console.log('=' .repeat(60));
  
  // Test users
  const users = [
    { username: 'admin', password: 'admin123' },
    { username: 'alice', password: 'secret' },
    { username: 'courier_level1', password: 'secret' }
  ];
  
  // Test via Gateway (正确的方式)
  console.log('\n1. Testing via GATEWAY (port 8000) - This is the correct way:');
  console.log('-'.repeat(60));
  for (const user of users) {
    await testLogin(GATEWAY_URL, user.username, user.password);
    await new Promise(r => setTimeout(r, 1000));
  }
  
  // Test via Backend directly (可能有问题)
  console.log('\n\n2. Testing DIRECTLY to Backend (port 8080) - May have issues:');
  console.log('-'.repeat(60));
  for (const user of users) {
    await testLogin(BACKEND_URL, user.username, user.password);
    await new Promise(r => setTimeout(r, 1000));
  }
}

main().catch(console.error);