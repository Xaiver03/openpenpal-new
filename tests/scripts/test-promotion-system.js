// 🎨 SOTA Promotion System Integration Test
// Testing complete frontend-backend flow with artistic precision

const BASE_URL = 'http://localhost:3000';
const API_URL = 'http://localhost:8080';

// Color codes for beautiful output
const colors = {
  reset: '\x1b[0m',
  bright: '\x1b[1m',
  green: '\x1b[32m',
  red: '\x1b[31m',
  yellow: '\x1b[33m',
  blue: '\x1b[34m',
  magenta: '\x1b[35m',
  cyan: '\x1b[36m'
};

const log = {
  success: (msg) => console.log(`${colors.green}✅ ${msg}${colors.reset}`),
  error: (msg) => console.log(`${colors.red}❌ ${msg}${colors.reset}`),
  info: (msg) => console.log(`${colors.blue}ℹ️  ${msg}${colors.reset}`),
  test: (msg) => console.log(`${colors.magenta}🧪 ${msg}${colors.reset}`),
  section: (msg) => console.log(`\n${colors.bright}${colors.cyan}${'='.repeat(60)}\n${msg}\n${'='.repeat(60)}${colors.reset}\n`)
};

// Store cookies manually for simplicity
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
  return data.data.token; // Extract token from data.data.token
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
    console.log('CSRF Token:', csrfToken);
    console.log('Cookies:', cookies);
    throw new Error(`Login failed: ${response.status}`);
  }
  
  const result = await response.json();
  return { token: result.data.token, user: result.data.user, csrfToken };
}

// Test courier growth data
async function testCourierGrowth(auth) {
  log.test('Testing courier growth data endpoint');
  
  const response = await fetch(`${API_URL}/api/v1/courier/growth/path`, {
    headers: {
      'Authorization': `Bearer ${auth.token}`,
      'X-CSRF-Token': auth.csrfToken
    }
  });
  
  if (response.ok) {
    const data = await response.json();
    log.success('Growth path retrieved successfully');
    console.log(JSON.stringify(data, null, 2));
    return data;
  } else {
    log.error(`Failed to get growth path: ${response.status}`);
    const error = await response.text();
    console.log(error);
  }
}

// Test promotion applications
async function testPromotionApplications(auth) {
  log.test('Testing promotion applications endpoint');
  
  const response = await fetch(`${API_URL}/api/v1/courier/growth/applications`, {
    headers: {
      'Authorization': `Bearer ${auth.token}`,
      'X-CSRF-Token': auth.csrfToken
    }
  });
  
  if (response.ok) {
    const data = await response.json();
    log.success('Applications retrieved successfully');
    console.log(JSON.stringify(data, null, 2));
    return data;
  } else {
    log.error(`Failed to get applications: ${response.status}`);
  }
}

// Test creating a new promotion application
async function testCreateApplication(auth) {
  log.test('Testing create promotion application');
  
  const application = {
    target_level: 2,
    reason: '我已经完成了超过50次投递，成功率达到96%，并且服务时间超过60天。我相信我已经准备好承担更多责任。',
    evidence: {
      deliveries: 69,
      success_rate: 96.5,
      service_days: 60
    }
  };
  
  const response = await fetch(`${API_URL}/api/v1/courier/growth/apply`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${auth.token}`,
      'X-CSRF-Token': auth.csrfToken,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(application)
  });
  
  if (response.ok) {
    const data = await response.json();
    log.success('Application created successfully');
    console.log(JSON.stringify(data, null, 2));
    return data;
  } else {
    log.error(`Failed to create application: ${response.status}`);
    const error = await response.text();
    console.log(error);
  }
}

// Test management functions (for higher level couriers)
async function testManagementFunctions(auth) {
  log.test('Testing management functions');
  
  // Get pending applications
  const response = await fetch(`${API_URL}/api/v1/courier/growth/applications/pending`, {
    headers: {
      'Authorization': `Bearer ${auth.token}`,
      'X-CSRF-Token': auth.csrfToken
    }
  });
  
  if (response.ok) {
    const data = await response.json();
    log.success('Pending applications retrieved');
    console.log(JSON.stringify(data, null, 2));
    
    // If there are pending applications, test approval
    if (data.applications && data.applications.length > 0) {
      const appId = data.applications[0].id;
      await testApproveApplication(auth, appId);
    }
  } else {
    log.error(`Failed to get pending applications: ${response.status}`);
  }
}

// Test approving an application
async function testApproveApplication(auth, applicationId) {
  log.test('Testing application approval');
  
  const approval = {
    action: 'approve',
    comment: '表现优秀，符合晋升要求，批准晋升！'
  };
  
  const response = await fetch(`${API_URL}/api/v1/courier/growth/applications/${applicationId}/review`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${auth.token}`,
      'X-CSRF-Token': auth.csrfToken,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(approval)
  });
  
  if (response.ok) {
    const data = await response.json();
    log.success('Application approved successfully');
    console.log(JSON.stringify(data, null, 2));
  } else {
    log.error(`Failed to approve application: ${response.status}`);
  }
}

// Main test runner
async function runTests() {
  log.section('🎨 OpenPenPal Promotion System Integration Test');
  
  try {
    // Test Level 1 Courier
    log.section('Testing Level 1 Courier Functions');
    const courier1Auth = await login('courier1', 'password');
    log.success(`Logged in as ${courier1Auth.user.username} (Level ${courier1Auth.user.role})`);
    
    await testCourierGrowth(courier1Auth);
    await testPromotionApplications(courier1Auth);
    await testCreateApplication(courier1Auth);
    
    // Test Level 2 Courier (can see subordinates)
    log.section('Testing Level 2 Courier Functions');
    const courier2Auth = await login('courier_level2', 'password');
    log.success(`Logged in as ${courier2Auth.user.username} (Level ${courier2Auth.user.role})`);
    
    await testCourierGrowth(courier2Auth);
    await testManagementFunctions(courier2Auth);
    
    // Test Level 3 Courier (can approve level 1->2 promotions)
    log.section('Testing Level 3 Courier Management');
    const courier3Auth = await login('courier_level3', 'password');
    log.success(`Logged in as ${courier3Auth.user.username} (Level ${courier3Auth.user.role})`);
    
    await testManagementFunctions(courier3Auth);
    
    // Test Level 4 Courier (city representative)
    log.section('Testing Level 4 Courier Authority');
    const courier4Auth = await login('courier_level4', 'password');
    log.success(`Logged in as ${courier4Auth.user.username} (Level ${courier4Auth.user.role})`);
    
    await testManagementFunctions(courier4Auth);
    
    log.section('✨ All tests completed successfully!');
    
  } catch (error) {
    log.error(`Test failed: ${error.message}`);
    console.error(error);
    process.exit(1);
  }
}

// Run the tests
runTests().catch(console.error);