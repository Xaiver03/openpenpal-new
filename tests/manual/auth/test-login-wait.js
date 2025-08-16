const axios = require('axios');

const API_BASE_URL = 'http://localhost:8080';

async function testLogin(username, password) {
  try {
    const response = await axios.post(`${API_BASE_URL}/api/v1/auth/login`, {
      username,
      password
    });
    
    console.log(`✅ ${username} login successful`);
    return true;
  } catch (error) {
    if (error.response?.status === 429) {
      console.log(`⏱️  ${username} - Rate limited (429)`);
      console.log(`   Retry-After: ${error.response.headers['retry-after']} seconds`);
    } else if (error.response?.status === 401) {
      console.log(`❌ ${username} - Invalid credentials (401)`);
    } else {
      console.log(`❌ ${username} - Error: ${error.message}`);
    }
    return false;
  }
}

async function main() {
  console.log('Testing login with delays to avoid rate limiting...\n');
  
  // Test with significant delays between requests
  console.log('1. Testing admin...');
  await testLogin('admin', 'admin123');
  
  console.log('\n2. Waiting 15 seconds before next attempt...');
  await new Promise(r => setTimeout(r, 15000));
  
  console.log('3. Testing alice...');
  await testLogin('alice', 'secret');
  
  console.log('\n4. Waiting 15 seconds before next attempt...');
  await new Promise(r => setTimeout(r, 15000));
  
  console.log('5. Testing courier_level1...');
  await testLogin('courier_level1', 'secret');
  
  console.log('\nDone!');
}

main().catch(console.error);