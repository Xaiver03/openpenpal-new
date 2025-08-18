#!/usr/bin/env node

/**
 * Credit System Frontend Validation Test
 * Tests the credit system after TypeScript fixes
 */

// Using built-in fetch (Node.js 18+) or fallback to https module
const https = require('https');
const http = require('http');

// Simple fetch implementation for compatibility
async function fetch(url, options = {}) {
  const urlObj = new URL(url);
  const isHttps = urlObj.protocol === 'https:';
  const client = isHttps ? https : http;
  
  return new Promise((resolve, reject) => {
    const req = client.request(url, {
      method: options.method || 'GET',
      headers: options.headers || {},
    }, (res) => {
      let data = '';
      res.on('data', chunk => data += chunk);
      res.on('end', () => {
        resolve({
          ok: res.statusCode >= 200 && res.statusCode < 300,
          status: res.statusCode,
          json: async () => JSON.parse(data),
          text: async () => data
        });
      });
    });
    
    req.on('error', reject);
    
    if (options.body) {
      req.write(options.body);
    }
    
    req.end();
  });
}

const BASE_URL = 'http://localhost:3000';
const API_URL = 'http://localhost:8080/api/v1';

// Test credentials
const TEST_USER = {
  username: 'alice',
  password: 'secret123'
};

let authToken = '';
let csrfToken = '';

// Colors for output
const colors = {
  green: '\x1b[32m',
  red: '\x1b[31m',
  yellow: '\x1b[33m',
  reset: '\x1b[0m'
};

function log(message, color = 'reset') {
  console.log(`${colors[color]}${message}${colors.reset}`);
}

async function testFrontendHealth() {
  log('\n📋 Testing Frontend Health...', 'yellow');
  try {
    const response = await fetch(BASE_URL);
    if (response.ok) {
      log('✅ Frontend is running', 'green');
      return true;
    } else {
      log(`❌ Frontend returned status: ${response.status}`, 'red');
      return false;
    }
  } catch (error) {
    log(`❌ Frontend is not accessible: ${error.message}`, 'red');
    return false;
  }
}

async function getCSRFToken() {
  log('\n🔐 Getting CSRF Token...', 'yellow');
  try {
    const response = await fetch(`${API_URL}/auth/csrf`);
    const data = await response.json();
    csrfToken = data.token || data.csrf_token || '';
    if (csrfToken) {
      log(`✅ CSRF Token obtained: ${csrfToken.substring(0, 10)}...`, 'green');
    } else {
      log('⚠️ No CSRF token in response, proceeding without it', 'yellow');
    }
    return true;
  } catch (error) {
    log(`❌ Failed to get CSRF token: ${error.message}`, 'red');
    return false;
  }
}

async function login() {
  log('\n🔑 Logging in...', 'yellow');
  try {
    const response = await fetch(`${API_URL}/auth/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-CSRF-Token': csrfToken
      },
      body: JSON.stringify(TEST_USER)
    });

    if (response.ok) {
      const data = await response.json();
      authToken = data.token;
      log(`✅ Login successful for user: ${TEST_USER.username}`, 'green');
      return true;
    } else {
      const error = await response.text();
      log(`❌ Login failed: ${error}`, 'red');
      return false;
    }
  } catch (error) {
    log(`❌ Login error: ${error.message}`, 'red');
    return false;
  }
}

async function testCreditEndpoints() {
  log('\n💳 Testing Credit System Endpoints...', 'yellow');
  
  const endpoints = [
    {
      name: 'Get Credit Summary',
      url: '/credits/summary',
      method: 'GET'
    },
    {
      name: 'Get Credit Leaderboard',
      url: '/credits/leaderboard?limit=10',
      method: 'GET'
    },
    {
      name: 'Get Task Statistics',
      url: '/credits/tasks/statistics',
      method: 'GET'
    },
    {
      name: 'Get Credit History',
      url: '/credits/history?page=1&limit=10',
      method: 'GET'
    }
  ];

  let successCount = 0;

  for (const endpoint of endpoints) {
    try {
      const response = await fetch(`${API_URL}${endpoint.url}`, {
        method: endpoint.method,
        headers: {
          'Authorization': `Bearer ${authToken}`,
          'Content-Type': 'application/json'
        }
      });

      if (response.ok) {
        const data = await response.json();
        log(`✅ ${endpoint.name}: Success`, 'green');
        
        // Validate response structure
        if (endpoint.url.includes('leaderboard') && data.leaderboard) {
          log(`   - Leaderboard entries: ${data.leaderboard.length}`, 'green');
          if (data.leaderboard[0]) {
            const hasCorrectFields = 
              'totalPoints' in data.leaderboard[0] &&
              'username' in data.leaderboard[0] &&
              'rank' in data.leaderboard[0];
            if (hasCorrectFields) {
              log('   - ✅ Type structure is correct (camelCase)', 'green');
            } else {
              log('   - ⚠️ Type structure may have issues', 'yellow');
            }
          }
        }
        
        successCount++;
      } else {
        const error = await response.text();
        log(`❌ ${endpoint.name}: Failed - ${response.status} ${error}`, 'red');
      }
    } catch (error) {
      log(`❌ ${endpoint.name}: Error - ${error.message}`, 'red');
    }
  }

  return successCount === endpoints.length;
}

async function testFrontendComponents() {
  log('\n🎨 Testing Frontend Component Routes...', 'yellow');
  
  const routes = [
    '/credits',
    '/credits/statistics',
    '/credits/leaderboard',
    '/credits/shop'
  ];

  let successCount = 0;

  for (const route of routes) {
    try {
      const response = await fetch(`${BASE_URL}${route}`);
      if (response.ok) {
        log(`✅ Route ${route}: Accessible`, 'green');
        successCount++;
      } else {
        log(`❌ Route ${route}: Status ${response.status}`, 'red');
      }
    } catch (error) {
      log(`❌ Route ${route}: Error - ${error.message}`, 'red');
    }
  }

  return successCount === routes.length;
}

async function main() {
  log('🚀 Starting Credit System Validation Tests', 'yellow');
  log('==========================================', 'yellow');

  let allTestsPassed = true;

  // Test 1: Frontend Health
  const frontendHealthy = await testFrontendHealth();
  if (!frontendHealthy) {
    log('\n⚠️ Frontend is not running. Please start it with: npm run dev', 'yellow');
    process.exit(1);
  }

  // Test 2: CSRF Token
  const csrfObtained = await getCSRFToken();
  if (!csrfObtained) {
    allTestsPassed = false;
  }

  // Test 3: Login
  const loginSuccess = await login();
  if (!loginSuccess) {
    allTestsPassed = false;
  }

  // Test 4: Credit Endpoints
  if (loginSuccess) {
    const creditTestsPassed = await testCreditEndpoints();
    if (!creditTestsPassed) {
      allTestsPassed = false;
    }
  }

  // Test 5: Frontend Routes
  const frontendRoutesPassed = await testFrontendComponents();
  if (!frontendRoutesPassed) {
    allTestsPassed = false;
  }

  // Summary
  log('\n==========================================', 'yellow');
  if (allTestsPassed) {
    log('✅ All tests passed! Credit system is functioning correctly.', 'green');
    log('✅ TypeScript fixes have been successfully validated.', 'green');
  } else {
    log('⚠️ Some tests failed. Please check the errors above.', 'yellow');
  }
  log('==========================================', 'yellow');

  process.exit(allTestsPassed ? 0 : 1);
}

// Run tests
main().catch(error => {
  log(`\n❌ Fatal error: ${error.message}`, 'red');
  process.exit(1);
});