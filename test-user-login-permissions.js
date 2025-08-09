const axios = require('axios');

const API_URL = 'http://localhost:8080/api/v1';

// Test users from the database
const testUsers = [
  { username: 'admin', password: 'admin123', expectedRole: 'super_admin' },
  { username: 'alice', password: 'secret', expectedRole: 'user' },
  { username: 'courier_level1', password: 'secret', expectedRole: 'courier_level1' },
  { username: 'courier_level2', password: 'secret', expectedRole: 'courier_level2' },
  { username: 'courier_level3', password: 'secret', expectedRole: 'courier_level3' },
  { username: 'courier_level4', password: 'secret', expectedRole: 'courier_level4' }
];

// Permission test endpoints
const permissionTests = {
  'user': [
    { method: 'GET', path: '/user/profile', shouldPass: true },
    { method: 'GET', path: '/admin/users', shouldPass: false },
    { method: 'GET', path: '/courier/tasks', shouldPass: false }
  ],
  'courier_level1': [
    { method: 'GET', path: '/user/profile', shouldPass: true },
    { method: 'GET', path: '/courier/tasks', shouldPass: true },
    { method: 'POST', path: '/courier/scan', shouldPass: true },
    { method: 'GET', path: '/admin/users', shouldPass: false }
  ],
  'super_admin': [
    { method: 'GET', path: '/user/profile', shouldPass: true },
    { method: 'GET', path: '/courier/tasks', shouldPass: true },
    { method: 'GET', path: '/admin/users', shouldPass: true },
    { method: 'GET', path: '/admin/system/config', shouldPass: true }
  ]
};

async function testUserLoginAndPermissions() {
  console.log('🔍 Testing User Login and Permission Flow\n');
  console.log('='.repeat(80));

  for (const user of testUsers) {
    console.log(`\n📤 Testing user: ${user.username} (Expected role: ${user.expectedRole})`);
    console.log('-'.repeat(50));

    try {
      // Step 1: Get CSRF token
      console.log('1️⃣  Getting CSRF token...');
      const csrfResponse = await axios.get(`${API_URL}/auth/csrf`);
      const csrfToken = csrfResponse.data.data.token;
      console.log(`   ✅ CSRF token received: ${csrfToken.substring(0, 20)}...`);

      // Step 2: Login
      console.log('2️⃣  Attempting login...');
      const loginResponse = await axios.post(`${API_URL}/auth/login`, {
        username: user.username,
        password: user.password
      }, {
        headers: {
          'X-CSRF-Token': csrfToken
        }
      });

      const { token, user: userData } = loginResponse.data.data;
      console.log(`   ✅ Login successful!`);
      console.log(`   • User ID: ${userData.id}`);
      console.log(`   • Role: ${userData.role}`);
      console.log(`   • Email: ${userData.email}`);
      console.log(`   • Is Active: ${userData.is_active}`);
      
      if (userData.courierInfo) {
        console.log(`   • Courier Level: ${userData.courierInfo.level}`);
        console.log(`   • Zone: ${userData.courierInfo.zoneCode}`);
        console.log(`   • Points: ${userData.courierInfo.points}`);
      }

      // Step 3: Test permissions
      console.log('\n3️⃣  Testing permissions...');
      const tests = permissionTests[userData.role] || permissionTests['user'];
      
      for (const test of tests) {
        try {
          const response = await axios({
            method: test.method,
            url: `${API_URL}${test.path}`,
            headers: {
              'Authorization': `Bearer ${token}`
            }
          });
          
          if (test.shouldPass) {
            console.log(`   ✅ ${test.method} ${test.path} - Access granted (Expected)`);
          } else {
            console.log(`   ❌ ${test.method} ${test.path} - Access granted (Should have been denied!)`);
          }
        } catch (error) {
          if (error.response?.status === 403 || error.response?.status === 401) {
            if (!test.shouldPass) {
              console.log(`   ✅ ${test.method} ${test.path} - Access denied (Expected)`);
            } else {
              console.log(`   ❌ ${test.method} ${test.path} - Access denied (Should have been granted!)`);
            }
          } else {
            console.log(`   ⚠️  ${test.method} ${test.path} - Error: ${error.response?.status || error.message}`);
          }
        }
      }

      // Step 4: Check token expiry
      console.log('\n4️⃣  Checking token expiry...');
      try {
        const expiryResponse = await axios.get(`${API_URL}/auth/check-token-expiry`, {
          headers: {
            'Authorization': `Bearer ${token}`
          }
        });
        const expiryData = expiryResponse.data.data;
        console.log(`   • Token expires at: ${new Date(expiryData.expires_at).toLocaleString()}`);
        console.log(`   • Remaining time: ${Math.floor(expiryData.remaining_time / 3600)} hours`);
      } catch (error) {
        console.log(`   ❌ Failed to check token expiry: ${error.message}`);
      }

    } catch (error) {
      console.log(`   ❌ Login failed: ${error.response?.data?.message || error.message}`);
    }
  }

  console.log('\n' + '='.repeat(80));
  console.log('✅ User Login and Permission Flow Test Complete\n');
}

// Run the test
testUserLoginAndPermissions().catch(console.error);