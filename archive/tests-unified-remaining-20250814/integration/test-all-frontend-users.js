const axios = require('axios');

const FRONTEND_URL = 'http://localhost:3001';

// Test users
const TEST_USERS = [
  { username: 'admin', password: 'admin123', expectedRole: 'super_admin' },
  { username: 'alice', password: 'secret', expectedRole: 'user' },
  { username: 'courier_level1', password: 'secret', expectedRole: 'courier_level1' },
  { username: 'courier_level2', password: 'secret', expectedRole: 'courier_level2' },
  { username: 'courier_level3', password: 'secret', expectedRole: 'courier_level3' },
  { username: 'courier_level4', password: 'secret', expectedRole: 'courier_level4' }
];

async function loginUser(username, password) {
  try {
    // Get CSRF token
    const csrfResponse = await axios.get(`${FRONTEND_URL}/api/auth/csrf`, {
      withCredentials: true
    });
    const csrfToken = csrfResponse.data.data?.token;
    
    // Login
    const loginResponse = await axios.post(
      `${FRONTEND_URL}/api/auth/login`,
      { username, password },
      {
        headers: { 'X-CSRF-Token': csrfToken },
        withCredentials: true
      }
    );
    
    return { success: true, data: loginResponse.data };
  } catch (error) {
    console.log(`   Login error status: ${error.response?.status}`);
    console.log(`   Login error data:`, error.response?.data);
    return { success: false, error: error.response?.data || error.message };
  }
}

async function testFrontendLogins() {
  console.log('Frontend Login Test - All Users');
  console.log('=' .repeat(60));
  console.log(`URL: ${FRONTEND_URL}`);
  console.log(`Time: ${new Date().toLocaleString()}`);
  console.log();
  
  let successCount = 0;
  let failCount = 0;
  
  for (const user of TEST_USERS) {
    console.log(`\nTesting: ${user.username}`);
    console.log('-'.repeat(40));
    
    const result = await loginUser(user.username, user.password);
    
    if (result.success) {
      const userData = result.data.data?.user;
      const tokenPresent = !!result.data.data?.accessToken;
      const roleMatch = userData?.role === user.expectedRole;
      
      if (roleMatch && tokenPresent) {
        console.log(`✅ SUCCESS - Login successful`);
        console.log(`   Username: ${userData.username}`);
        console.log(`   Role: ${userData.role}`);
        console.log(`   Email: ${userData.email || 'N/A'}`);
        console.log(`   Token: ${tokenPresent ? 'Present' : 'Missing'}`);
        successCount++;
      } else {
        console.log(`⚠️  PARTIAL - Login returned but with issues`);
        console.log(`   Role match: ${roleMatch ? 'Yes' : 'No'} (expected: ${user.expectedRole})`);
        console.log(`   Token: ${tokenPresent ? 'Present' : 'Missing'}`);
        failCount++;
      }
    } else {
      console.log(`❌ FAILED - ${JSON.stringify(result.error)}`);
      failCount++;
    }
    
    // Wait between requests to avoid rate limiting
    await new Promise(r => setTimeout(r, 1000));
  }
  
  console.log('\n' + '=' .repeat(60));
  console.log('Summary:');
  console.log(`Total: ${TEST_USERS.length}`);
  console.log(`✅ Success: ${successCount}`);
  console.log(`❌ Failed: ${failCount}`);
  console.log(`Success Rate: ${((successCount / TEST_USERS.length) * 100).toFixed(1)}%`);
  
  if (successCount === TEST_USERS.length) {
    console.log('\n🎉 All users can login successfully!');
  } else {
    console.log('\n⚠️  Some users failed to login');
  }
}

testFrontendLogins().catch(console.error);