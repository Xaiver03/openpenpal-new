const axios = require('axios');

async function testBackendResponse() {
  console.log('Testing Backend Response Format');
  console.log('=' .repeat(60));
  
  try {
    const response = await axios.post('http://localhost:8080/api/v1/auth/login', {
      username: 'courier_level1',
      password: 'secret'
    });
    
    console.log('\n✅ Login successful!');
    console.log('\nFull Response Data:');
    console.log(JSON.stringify(response.data, null, 2));
    
    console.log('\nKey fields:');
    console.log('- response.data.success:', response.data.success);
    console.log('- response.data.data:', typeof response.data.data);
    console.log('- response.data.data.token:', response.data.data.token ? 'Present' : 'Missing');
    console.log('- response.data.data.user:', response.data.data.user ? 'Present' : 'Missing');
    console.log('- response.data.message:', response.data.message);
    
  } catch (error) {
    console.log('\n❌ Login failed!');
    console.log('Status:', error.response?.status);
    console.log('Response data:', JSON.stringify(error.response?.data, null, 2));
  }
}

testBackendResponse();