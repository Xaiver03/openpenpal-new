const http = require('http');

// Test backend health
const testBackendHealth = () => {
  const options = {
    hostname: '127.0.0.1',
    port: 8080,
    path: '/health',
    method: 'GET'
  };

  const req = http.request(options, (res) => {
    console.log(`Health check status: ${res.statusCode}`);
    
    let data = '';
    res.on('data', (chunk) => {
      data += chunk;
    });
    
    res.on('end', () => {
      console.log('Response:', data);
    });
  });

  req.on('error', (error) => {
    console.error('Error:', error);
  });

  req.end();
};

// Test login endpoint
const testLogin = (username, password) => {
  const postData = JSON.stringify({
    username: username,
    password: password
  });

  const options = {
    hostname: '127.0.0.1',
    port: 8080,
    path: '/api/v1/auth/login',
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Content-Length': Buffer.byteLength(postData)
    }
  };

  const req = http.request(options, (res) => {
    console.log(`\nLogin test for ${username}: ${res.statusCode}`);
    
    let data = '';
    res.on('data', (chunk) => {
      data += chunk;
    });
    
    res.on('end', () => {
      console.log('Response:', data);
      if (res.statusCode === 200) {
        try {
          const parsed = JSON.parse(data);
          if (parsed.data && parsed.data.token) {
            console.log('✓ Login successful! Token received');
            console.log('User role:', parsed.data.user.role);
          }
        } catch (e) {
          console.log('Failed to parse response');
        }
      }
    });
  });

  req.on('error', (error) => {
    console.error('Error:', error);
  });

  req.write(postData);
  req.end();
};

// Run tests
console.log('Testing backend connectivity...');
testBackendHealth();

setTimeout(() => {
  console.log('\nTesting authentication endpoints...');
  testLogin('alice', 'secret');
  testLogin('admin', 'admin123');
  testLogin('courier1', 'courier123');
}, 1000);