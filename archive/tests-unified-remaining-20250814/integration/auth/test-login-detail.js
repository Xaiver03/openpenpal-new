const axios = require('axios');

const API_BASE_URL = 'http://localhost:8080';

async function testLoginDetail() {
  console.log('Testing login with alice/secret...\n');
  
  try {
    const response = await axios.post(`${API_BASE_URL}/api/v1/auth/login`, {
      username: 'alice',
      password: 'secret'
    }, {
      validateStatus: function (status) {
        return true; // Accept any status code
      }
    });
    
    console.log('Response Status:', response.status);
    console.log('Response Headers:', response.headers);
    console.log('Response Data:', JSON.stringify(response.data, null, 2));
    
    if (response.status === 200) {
      console.log('\n✅ Login successful!');
    } else {
      console.log('\n❌ Login failed!');
    }
    
  } catch (error) {
    console.error('Request Error:', error.message);
    if (error.response) {
      console.error('Error Response:', error.response.data);
    }
  }
}

testLoginDetail();