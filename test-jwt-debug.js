// Debug JWT Authentication Issue

const API_URL = 'http://localhost:8080';

// Store cookies manually
let cookies = {};

// Helper to parse Set-Cookie headers
function parseCookies(setCookieHeaders) {
  if (!setCookieHeaders) return;
  const headers = Array.isArray(setCookieHeaders) ? setCookieHeaders : [setCookieHeaders];
  headers.forEach(header => {
    const [cookie] = header.split(';');
    const [name, value] = cookie.split('=');
    cookies[name] = value;
  });
}

// Helper to create Cookie header
function getCookieHeader() {
  return Object.entries(cookies)
    .map(([name, value]) => `${name}=${value}`)
    .join('; ');
}

// Helper function to get CSRF token
async function getCSRFToken() {
  const response = await fetch(`${API_URL}/api/v1/auth/csrf`, {
    headers: {
      'Cookie': getCookieHeader()
    }
  });
  parseCookies(response.headers.get('set-cookie'));
  const data = await response.json();
  return data.data.token;
}

// Helper function to login
async function login(username, password) {
  const csrfToken = await getCSRFToken();
  
  const response = await fetch(`${API_URL}/api/v1/auth/login`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'X-CSRF-Token': csrfToken,
      'Cookie': getCookieHeader()
    },
    body: JSON.stringify({ username, password })
  });
  
  parseCookies(response.headers.get('set-cookie'));
  
  if (!response.ok) {
    const error = await response.text();
    console.log('Login error:', error);
    throw new Error(`Login failed: ${response.status}`);
  }
  
  const result = await response.json();
  return { token: result.data.token, user: result.data.user, csrfToken };
}

// Test the JWT authentication
async function testJWTAuth() {
  try {
    // Login as courier1
    console.log('\n1. Logging in as courier1...');
    const auth = await login('courier1', 'password');
    console.log('Login successful!');
    console.log('User ID:', auth.user.id);
    console.log('User Role:', auth.user.role);
    console.log('JWT Token:', auth.token.substring(0, 50) + '...');
    
    // Test a simple authenticated endpoint
    console.log('\n2. Testing authenticated endpoint /api/v1/users/me...');
    const meResponse = await fetch(`${API_URL}/api/v1/users/me`, {
      headers: {
        'Authorization': `Bearer ${auth.token}`,
        'Cookie': getCookieHeader()
      }
    });
    
    if (meResponse.ok) {
      const userData = await meResponse.json();
      console.log('User data retrieved successfully!');
      console.log('User ID from /me endpoint:', userData.data.id);
    } else {
      console.log('Failed to get user data:', meResponse.status);
      console.log(await meResponse.text());
    }
    
    // Test courier info endpoint
    console.log('\n3. Testing /api/v1/courier/me endpoint...');
    const courierResponse = await fetch(`${API_URL}/api/v1/courier/me`, {
      headers: {
        'Authorization': `Bearer ${auth.token}`,
        'Cookie': getCookieHeader()
      }
    });
    
    if (courierResponse.ok) {
      const courierData = await courierResponse.json();
      console.log('Courier data retrieved successfully!');
      console.log(JSON.stringify(courierData, null, 2));
    } else {
      console.log('Failed to get courier data:', courierResponse.status);
      console.log(await courierResponse.text());
    }
    
    // Test the growth path endpoint
    console.log('\n4. Testing /api/v1/courier/growth/path endpoint...');
    const growthResponse = await fetch(`${API_URL}/api/v1/courier/growth/path`, {
      headers: {
        'Authorization': `Bearer ${auth.token}`,
        'Cookie': getCookieHeader()
      }
    });
    
    if (growthResponse.ok) {
      const growthData = await growthResponse.json();
      console.log('Growth path retrieved successfully!');
      console.log(JSON.stringify(growthData, null, 2));
    } else {
      console.log('Failed to get growth path:', growthResponse.status);
      console.log(await growthResponse.text());
    }
    
  } catch (error) {
    console.error('Test failed:', error);
  }
}

// Run the test
testJWTAuth();