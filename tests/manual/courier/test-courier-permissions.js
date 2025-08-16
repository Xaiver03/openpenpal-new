const axios = require('axios');
const fs = require('fs');

const API_URL = 'http://localhost:8000/api/v1';
const FRONTEND_URL = 'http://localhost:3000';

// Test accounts for each courier level
const courierAccounts = [
  { username: 'courier_level1', password: 'password', level: 1, expectedRole: 'courier' },
  { username: 'courier_level2', password: 'password', level: 2, expectedRole: 'senior_courier' },
  { username: 'courier_level3', password: 'password', level: 3, expectedRole: 'coordinator' },
  { username: 'courier_level4', password: 'password', level: 4, expectedRole: 'coordinator' }
];

// Permission test endpoints
const permissionTests = [
  // Basic courier info (all levels)
  { endpoint: '/courier/info', method: 'GET', minLevel: 1, description: 'Get courier info' },
  { endpoint: '/courier/stats', method: 'GET', minLevel: 1, description: 'Get courier stats' },
  
  // Task management (L2+)
  { endpoint: '/courier/tasks', method: 'GET', minLevel: 2, description: 'Get tasks' },
  { endpoint: '/courier/hierarchy/subordinates', method: 'GET', minLevel: 2, description: 'Get subordinates' },
  
  // School management (L3+)
  { endpoint: '/courier/hierarchy/level/2', method: 'GET', minLevel: 3, description: 'Get L2 couriers' },
  { endpoint: '/courier/hierarchy/create', method: 'POST', minLevel: 3, description: 'Create subordinate', 
    body: { username: 'test_subordinate', level: 2 } },
  
  // City management (L4)
  { endpoint: '/courier/hierarchy/level/3', method: 'GET', minLevel: 4, description: 'Get L3 couriers' },
  { endpoint: '/courier/hierarchy/city/overview', method: 'GET', minLevel: 4, description: 'City overview' }
];

// Frontend route tests
const frontendRoutes = [
  { path: '/courier', minLevel: 1, description: 'Basic courier dashboard' },
  { path: '/courier/tasks', minLevel: 1, description: 'Task list' },
  { path: '/courier/building-manage', minLevel: 2, description: 'Building management' },
  { path: '/courier/zone-manage', minLevel: 3, description: 'Zone management' },
  { path: '/courier/school-manage', minLevel: 3, description: 'School management' },
  { path: '/courier/city-manage', minLevel: 4, description: 'City management' }
];

async function loginCourier(account) {
  try {
    const response = await axios.post(`${API_URL}/auth/login`, {
      username: account.username,
      password: account.password
    });
    
    return {
      success: true,
      token: response.data.data.token,
      user: response.data.data.user
    };
  } catch (error) {
    return {
      success: false,
      error: error.response?.data?.message || error.message
    };
  }
}

async function testAPIPermission(test, token, courierLevel) {
  try {
    const config = {
      headers: { Authorization: `Bearer ${token}` }
    };
    
    if (test.body) {
      config.headers['Content-Type'] = 'application/json';
    }
    
    const response = await axios({
      method: test.method,
      url: `${API_URL}${test.endpoint}`,
      data: test.body,
      ...config
    });
    
    const shouldHaveAccess = courierLevel >= test.minLevel;
    const hasAccess = response.status < 400;
    
    return {
      endpoint: test.endpoint,
      description: test.description,
      minLevel: test.minLevel,
      courierLevel,
      shouldHaveAccess,
      hasAccess,
      passed: shouldHaveAccess === hasAccess,
      status: response.status,
      data: response.data
    };
  } catch (error) {
    const shouldHaveAccess = courierLevel >= test.minLevel;
    const hasAccess = false;
    
    return {
      endpoint: test.endpoint,
      description: test.description,
      minLevel: test.minLevel,
      courierLevel,
      shouldHaveAccess,
      hasAccess,
      passed: shouldHaveAccess === hasAccess,
      status: error.response?.status || 0,
      error: error.response?.data?.message || error.message
    };
  }
}

async function testFrontendRoute(route, token, courierLevel) {
  try {
    // Get CSRF token first
    const csrfResponse = await axios.get(`${FRONTEND_URL}/api/auth/csrf`, {
      headers: { 
        Cookie: `token=${token}`,
        Accept: 'application/json'
      }
    });
    
    const response = await axios.get(`${FRONTEND_URL}${route.path}`, {
      headers: {
        Cookie: `token=${token}; csrf_token=${csrfResponse.data.csrfToken}`,
        'X-CSRF-Token': csrfResponse.data.csrfToken,
        Accept: 'text/html'
      },
      maxRedirects: 0,
      validateStatus: (status) => status < 500
    });
    
    const shouldHaveAccess = courierLevel >= route.minLevel;
    const hasAccess = response.status === 200;
    
    return {
      path: route.path,
      description: route.description,
      minLevel: route.minLevel,
      courierLevel,
      shouldHaveAccess,
      hasAccess,
      passed: shouldHaveAccess === hasAccess,
      status: response.status
    };
  } catch (error) {
    const shouldHaveAccess = courierLevel >= route.minLevel;
    const hasAccess = false;
    
    return {
      path: route.path,
      description: route.description,
      minLevel: route.minLevel,
      courierLevel,
      shouldHaveAccess,
      hasAccess,
      passed: shouldHaveAccess === hasAccess,
      status: error.response?.status || 0,
      error: error.message
    };
  }
}

async function runTests() {
  console.log('🚀 Starting Courier Permission Tests\n');
  
  const results = {
    login: [],
    api: [],
    frontend: [],
    summary: {
      totalTests: 0,
      passed: 0,
      failed: 0
    }
  };
  
  // Test login for each courier level
  console.log('📝 Testing Login for Each Courier Level:');
  console.log('─'.repeat(60));
  
  for (const account of courierAccounts) {
    console.log(`\nTesting ${account.username} (Level ${account.level})...`);
    const loginResult = await loginCourier(account);
    
    if (loginResult.success) {
      console.log(`✅ Login successful`);
      console.log(`   Role: ${loginResult.user.role}`);
      console.log(`   Courier Level: ${loginResult.user.courier?.level || 'N/A'}`);
      
      // Store token for permission tests
      account.token = loginResult.token;
      account.actualRole = loginResult.user.role;
      
      results.login.push({
        username: account.username,
        level: account.level,
        success: true,
        expectedRole: account.expectedRole,
        actualRole: loginResult.user.role,
        passed: loginResult.user.role === account.expectedRole
      });
    } else {
      console.log(`❌ Login failed: ${loginResult.error}`);
      results.login.push({
        username: account.username,
        level: account.level,
        success: false,
        error: loginResult.error,
        passed: false
      });
    }
  }
  
  // Test API permissions
  console.log('\n\n📝 Testing API Permissions:');
  console.log('─'.repeat(60));
  
  for (const account of courierAccounts) {
    if (!account.token) continue;
    
    console.log(`\nTesting permissions for ${account.username} (Level ${account.level}):`);
    
    for (const test of permissionTests) {
      const result = await testAPIPermission(test, account.token, account.level);
      results.api.push(result);
      
      const icon = result.passed ? '✅' : '❌';
      const accessText = result.hasAccess ? 'allowed' : 'denied';
      console.log(`${icon} ${test.description}: ${accessText} (${result.shouldHaveAccess ? 'expected' : 'unexpected'})`);
    }
  }
  
  // Test frontend routes
  console.log('\n\n📝 Testing Frontend Route Access:');
  console.log('─'.repeat(60));
  
  for (const account of courierAccounts) {
    if (!account.token) continue;
    
    console.log(`\nTesting routes for ${account.username} (Level ${account.level}):`);
    
    for (const route of frontendRoutes) {
      const result = await testFrontendRoute(route, account.token, account.level);
      results.frontend.push(result);
      
      const icon = result.passed ? '✅' : '❌';
      const accessText = result.hasAccess ? 'accessible' : 'blocked';
      console.log(`${icon} ${route.path}: ${accessText} (${result.shouldHaveAccess ? 'expected' : 'unexpected'})`);
    }
  }
  
  // Calculate summary
  results.summary.totalTests = results.login.length + results.api.length + results.frontend.length;
  results.summary.passed = [...results.login, ...results.api, ...results.frontend].filter(r => r.passed).length;
  results.summary.failed = results.summary.totalTests - results.summary.passed;
  
  // Print summary
  console.log('\n\n📊 Test Summary:');
  console.log('─'.repeat(60));
  console.log(`Total Tests: ${results.summary.totalTests}`);
  console.log(`Passed: ${results.summary.passed} (${(results.summary.passed / results.summary.totalTests * 100).toFixed(1)}%)`);
  console.log(`Failed: ${results.summary.failed} (${(results.summary.failed / results.summary.totalTests * 100).toFixed(1)}%)`);
  
  // Print failed tests
  if (results.summary.failed > 0) {
    console.log('\n\n❌ Failed Tests:');
    console.log('─'.repeat(60));
    
    const failedTests = [
      ...results.login.filter(r => !r.passed).map(r => ({ type: 'Login', ...r })),
      ...results.api.filter(r => !r.passed).map(r => ({ type: 'API', ...r })),
      ...results.frontend.filter(r => !r.passed).map(r => ({ type: 'Frontend', ...r }))
    ];
    
    failedTests.forEach(test => {
      console.log(`\n${test.type} Test Failed:`);
      console.log(JSON.stringify(test, null, 2));
    });
  }
  
  // Save results
  fs.writeFileSync('courier-permission-test-results.json', JSON.stringify(results, null, 2));
  console.log('\n\n📄 Full results saved to courier-permission-test-results.json');
  
  // Specific check for L4 courier management access
  const l4Account = courierAccounts.find(a => a.level === 4);
  const l4ManagementRoute = results.frontend.find(r => 
    r.path === '/courier/city-manage' && r.courierLevel === 4
  );
  
  if (l4ManagementRoute && !l4ManagementRoute.hasAccess) {
    console.log('\n\n⚠️  ISSUE DETECTED: Level 4 courier cannot access city management interface!');
    console.log('This matches the reported issue. Investigating further...');
  }
}

// Run the tests
runTests().catch(console.error);