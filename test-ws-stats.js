const fetch = require('node-fetch');

async function testWebSocketStats() {
  // Test with a valid token
  const token = process.env.TEST_TOKEN || 'your-jwt-token-here';
  
  console.log('Testing WebSocket stats endpoint...');
  
  try {
    const response = await fetch('http://localhost:8080/api/v1/ws/stats', {
      headers: {
        'Authorization': `Bearer ${token}`
      }
    });
    
    console.log('Response status:', response.status);
    console.log('Response headers:', response.headers.raw());
    
    const text = await response.text();
    console.log('Response body:', text);
    
    if (response.ok) {
      try {
        const data = JSON.parse(text);
        console.log('Parsed data:', JSON.stringify(data, null, 2));
      } catch (e) {
        console.log('Failed to parse JSON:', e.message);
      }
    }
  } catch (error) {
    console.error('Request failed:', error);
  }
}

// Get token from login first
async function getToken() {
  try {
    const loginResponse = await fetch('http://localhost:8080/api/v1/auth/login', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        username: 'alice',
        password: 'secret'
      })
    });
    
    if (loginResponse.ok) {
      const loginData = await loginResponse.json();
      return loginData.data.token;
    }
  } catch (error) {
    console.error('Login failed:', error);
  }
  return null;
}

async function main() {
  const token = await getToken();
  if (token) {
    console.log('Got token:', token.substring(0, 20) + '...');
    process.env.TEST_TOKEN = token;
    await testWebSocketStats();
  } else {
    console.error('Failed to get token');
  }
}

main();