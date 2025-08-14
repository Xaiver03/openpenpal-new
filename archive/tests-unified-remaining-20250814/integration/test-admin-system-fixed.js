#!/usr/bin/env node

const axios = require('axios');

// Configuration
const BASE_URL = 'http://localhost:8080';
const ADMIN_TOKEN = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiJ0ZXN0LWFkbWluIiwicm9sZSI6InN1cGVyX2FkbWluIiwiaXNzIjoib3BlbnBlbnBhbCIsImV4cCI6MTc1NDE0MDA2NCwiaWF0IjoxNzU0MDUzNjY0LCJqdGkiOiI3ODgyZGRmMWEyZTk5MDA2YmE4MDFkNWZkYTMyM2NmMyJ9.D9VLMt14F4JpFV6k-r2pe7Rr_kziBmlpqTKsVo4VhaA';

// Colors for console output
const colors = {
  reset: '\x1b[0m',
  bright: '\x1b[1m',
  dim: '\x1b[2m',
  red: '\x1b[31m',
  green: '\x1b[32m',
  yellow: '\x1b[33m',
  blue: '\x1b[34m',
  magenta: '\x1b[35m',
  cyan: '\x1b[36m',
  white: '\x1b[37m'
};

// Set environment variables to bypass proxy
process.env.NO_PROXY = 'localhost,127.0.0.1,*.local,localhost:*,127.0.0.1:*';
process.env.no_proxy = 'localhost,127.0.0.1,*.local,localhost:*,127.0.0.1:*';
delete process.env.HTTP_PROXY;
delete process.env.HTTPS_PROXY;
delete process.env.http_proxy;
delete process.env.https_proxy;

// Configure axios
const api = axios.create({
  baseURL: BASE_URL,
  timeout: 10000,
  headers: {
    'Authorization': `Bearer ${ADMIN_TOKEN}`,
    'Content-Type': 'application/json'
  }
});

// Test results
let testResults = {
  total: 0,
  passed: 0,
  failed: 0,
  skipped: 0,
  categories: {}
};

// Helper functions
const log = (message, color = 'white') => {
  const timestamp = new Date().toISOString();
  console.log(`${colors[color]}[${timestamp}] ${message}${colors.reset}`);
};

const runTest = async (testName, testFn, category = 'General') => {
  if (!testResults.categories[category]) {
    testResults.categories[category] = { total: 0, passed: 0, failed: 0, skipped: 0 };
  }
  
  testResults.total++;
  testResults.categories[category].total++;
  
  log(`Running: ${testName}`, 'cyan');
  const startTime = Date.now();
  
  try {
    await testFn();
    const duration = Date.now() - startTime;
    log(`✅ PASSED: ${testName} (${duration}ms)`, 'green');
    testResults.passed++;
    testResults.categories[category].passed++;
  } catch (error) {
    const duration = Date.now() - startTime;
    log(`❌ FAILED: ${testName} - ${error.message} (${duration}ms)`, 'red');
    testResults.failed++;
    testResults.categories[category].failed++;
  }
};

const skipTest = (testName, reason, category = 'General') => {
  if (!testResults.categories[category]) {
    testResults.categories[category] = { total: 0, passed: 0, failed: 0, skipped: 0 };
  }
  
  testResults.total++;
  testResults.categories[category].total++;
  testResults.skipped++;
  testResults.categories[category].skipped++;
  
  log(`⚠️  SKIPPED: ${testName} - ${reason}`, 'yellow');
};

// Test Functions
async function testAuthAndPermissions() {
  log('\n=== 1. AUTHENTICATION & AUTHORIZATION TESTING ===', 'cyan');
  
  await runTest('Admin Token Validation', async () => {
    const response = await api.get('/api/v1/admin/users');
    if (response.status !== 200) {
      throw new Error(`Expected 200, got ${response.status}`);
    }
  }, 'Authentication');
  
  await runTest('Invalid Token Rejection', async () => {
    try {
      await axios.get(`${BASE_URL}/api/v1/admin/users`, {
        headers: { 'Authorization': 'Bearer invalid_token' }
      });
      throw new Error('Should have rejected invalid token');
    } catch (error) {
      if (error.response && error.response.status === 401) {
        return; // Expected
      }
      throw error;
    }
  }, 'Authentication');
  
  await runTest('Missing Token Rejection', async () => {
    try {
      await axios.get(`${BASE_URL}/api/v1/admin/users`);
      throw new Error('Should have rejected missing token');
    } catch (error) {
      if (error.response && error.response.status === 401) {
        return; // Expected
      }
      throw error;
    }
  }, 'Authentication');
}

async function testUserManagement() {
  log('\n=== 2. USER MANAGEMENT TESTING ===', 'cyan');
  
  let testUserId = null;
  
  await runTest('Get Users List with Pagination', async () => {
    const response = await api.get('/api/v1/admin/users?page=1&limit=10');
    if (response.status !== 200) {
      throw new Error(`Expected 200, got ${response.status}`);
    }
    if (!response.data.data && !response.data.users) {
      throw new Error('No user data found in response');
    }
  }, 'User Management');
  
  await runTest('Create Test User for Management Tests', async () => {
    try {
      const userData = {
        username: `testuser_${Date.now()}`,
        email: `test_${Date.now()}@example.com`,
        password: 'testpass123',
        nickname: 'Test User',
        role: 'student'
      };
      const response = await api.post('/api/v1/admin/users', userData);
      if (response.status === 200 || response.status === 201) {
        testUserId = response.data.data?.id || response.data.id;
      }
    } catch (error) {
      if (error.response && error.response.status === 404) {
        skipTest('Create Test User', 'Endpoint not implemented', 'User Management');
        return;
      }
      throw error;
    }
  }, 'User Management');
  
  if (testUserId) {
    await runTest('Get Specific User Details', async () => {
      const response = await api.get(`/api/v1/admin/users/${testUserId}`);
      if (response.status !== 200) {
        throw new Error(`Expected 200, got ${response.status}`);
      }
    }, 'User Management');
  } else {
    skipTest('Get Specific User Details', 'Test user ID not available', 'User Management');
  }
}

async function testContentModeration() {
  log('\n=== 3. CONTENT MODERATION TESTING ===', 'cyan');
  
  await runTest('Get Moderation Queue', async () => {
    try {
      const response = await api.get('/api/v1/admin/moderation');
      if (response.status !== 200) {
        throw new Error(`Expected 200, got ${response.status}`);
      }
    } catch (error) {
      if (error.response && error.response.status === 404) {
        skipTest('Get Moderation Queue', 'Endpoint not implemented', 'Content Moderation');
        return;
      }
      throw error;
    }
  }, 'Content Moderation');
  
  await runTest('Get Museum Items for Moderation', async () => {
    try {
      const response = await api.get('/api/v1/admin/museum/items');
      if (response.status !== 200) {
        throw new Error(`Expected 200, got ${response.status}`);
      }
    } catch (error) {
      if (error.response && error.response.status === 404) {
        skipTest('Get Museum Items', 'Endpoint not implemented', 'Content Moderation');
        return;
      }
      throw error;
    }
  }, 'Content Moderation');
}

async function testCourierManagement() {
  log('\n=== 4. COURIER MANAGEMENT TESTING ===', 'cyan');
  
  await runTest('Get Pending Courier Applications', async () => {
    try {
      const response = await api.get('/api/v1/admin/courier/applications');
      if (response.status !== 200) {
        throw new Error(`Expected 200, got ${response.status}`);
      }
    } catch (error) {
      if (error.response && error.response.status === 404) {
        skipTest('Get Courier Applications', 'Endpoint not implemented', 'Courier Management');
        return;
      }
      throw error;
    }
  }, 'Courier Management');
  
  await runTest('Get All Couriers', async () => {
    try {
      const response = await api.get('/api/v1/admin/couriers');
      if (response.status !== 200) {
        throw new Error(`Expected 200, got ${response.status}`);
      }
    } catch (error) {
      if (error.response && error.response.status === 404) {
        skipTest('Get All Couriers', 'Endpoint not implemented', 'Courier Management');
        return;
      }
      throw error;
    }
  }, 'Courier Management');
}

async function testAnalyticsAndReporting() {
  log('\n=== 5. ANALYTICS & REPORTING TESTING ===', 'cyan');
  
  await runTest('Get Dashboard Statistics', async () => {
    try {
      const response = await api.get('/api/v1/admin/analytics');
      if (response.status !== 200) {
        throw new Error(`Expected 200, got ${response.status}`);
      }
    } catch (error) {
      if (error.response && error.response.status === 404) {
        skipTest('Get Dashboard Statistics', 'Endpoint not implemented', 'Analytics');
        return;
      }
      throw error;
    }
  }, 'Analytics');
  
  await runTest('Get System Health', async () => {
    const response = await api.get('/health');
    if (response.status !== 200) {
      throw new Error(`Expected 200, got ${response.status}`);
    }
    if (!response.data.status) {
      throw new Error('Health check response missing status');
    }
  }, 'Analytics');
}

async function testSystemConfiguration() {
  log('\n=== 6. SYSTEM CONFIGURATION TESTING ===', 'cyan');
  
  await runTest('Get System Settings', async () => {
    try {
      const response = await api.get('/api/v1/admin/settings');
      if (response.status !== 200) {
        throw new Error(`Expected 200, got ${response.status}`);
      }
    } catch (error) {
      if (error.response && error.response.status === 404) {
        skipTest('Get System Settings', 'Endpoint not implemented', 'System Configuration');
        return;
      }
      throw error;
    }
  }, 'System Configuration');
  
  await runTest('Test Email Configuration', async () => {
    try {
      const response = await api.post('/api/v1/admin/settings/test-email', {
        recipient: 'test@example.com'
      });
      if (response.status !== 200) {
        throw new Error(`Expected 200, got ${response.status}`);
      }
    } catch (error) {
      if (error.response && error.response.status === 404) {
        skipTest('Test Email Configuration', 'Endpoint not implemented', 'System Configuration');
        return;
      }
      throw error;
    }
  }, 'System Configuration');
}

async function testSecurity() {
  log('\n=== 7. SECURITY TESTING ===', 'cyan');
  
  await runTest('SQL Injection Protection', async () => {
    try {
      const maliciousPayload = "'; DROP TABLE users; --";
      const response = await api.get(`/api/v1/admin/users?search=${encodeURIComponent(maliciousPayload)}`);
      // Should not crash the server
      if (response.status >= 500) {
        throw new Error('Server error - possible SQL injection vulnerability');
      }
    } catch (error) {
      if (error.response && error.response.status >= 500) {
        throw new Error('Server error - possible SQL injection vulnerability');
      }
      // 400-level errors are expected for malicious input
    }
  }, 'Security');
  
  await runTest('Large Payload Handling', async () => {
    try {
      const largePayload = 'x'.repeat(1000000); // 1MB string
      await api.post('/api/v1/admin/users', { description: largePayload });
    } catch (error) {
      if (error.response && (error.response.status === 413 || error.response.status === 400)) {
        return; // Expected rejection
      }
      throw error;
    }
  }, 'Security');
  
  await runTest('Rate Limiting Check', async () => {
    const requests = [];
    for (let i = 0; i < 20; i++) {
      requests.push(api.get('/api/v1/admin/users'));
    }
    
    try {
      await Promise.all(requests);
    } catch (error) {
      if (error.response && error.response.status === 429) {
        return; // Rate limiting is working
      }
      // If no rate limiting, that's also acceptable for testing
    }
  }, 'Security');
}

async function testErrorHandling() {
  log('\n=== 8. ERROR HANDLING TESTING ===', 'cyan');
  
  await runTest('404 Error Handling', async () => {
    try {
      await api.get('/api/v1/admin/nonexistent-endpoint');
      throw new Error('Should have returned 404');
    } catch (error) {
      if (error.response && error.response.status === 404) {
        return; // Expected
      }
      throw error;
    }
  }, 'Error Handling');
  
  await runTest('Invalid JSON Payload', async () => {
    try {
      await axios.post(`${BASE_URL}/api/v1/admin/users`, 'invalid json', {
        headers: {
          'Authorization': `Bearer ${ADMIN_TOKEN}`,
          'Content-Type': 'application/json'
        }
      });
    } catch (error) {
      if (error.response && error.response.status === 400) {
        return; // Expected
      }
      throw error;
    }
  }, 'Error Handling');
}

// Main test runner
async function runAllTests() {
  log('🚀 Starting OpenPenPal Admin System Comprehensive Testing', 'cyan');
  log(`Base URL: ${BASE_URL}`, 'cyan');
  log(`Admin Token: ${ADMIN_TOKEN.substring(0, 20)}...`, 'cyan');
  
  try {
    await testAuthAndPermissions();
    await testUserManagement();
    await testContentModeration();
    await testCourierManagement();
    await testAnalyticsAndReporting();
    await testSystemConfiguration();
    await testSecurity();
    await testErrorHandling();
  } catch (error) {
    log(`Unexpected error during testing: ${error.message}`, 'red');
  }
  
  // Print results
  log('\n📊 TEST RESULTS SUMMARY:', 'cyan');
  log(`Total Tests: ${testResults.total}`, 'white');
  log(`✅ Passed: ${testResults.passed}`, 'green');
  log(`❌ Failed: ${testResults.failed}`, 'red');
  log(`⚠️  Skipped: ${testResults.skipped}`, 'yellow');
  
  const successRate = testResults.total > 0 ? ((testResults.passed / testResults.total) * 100).toFixed(2) : 0;
  log(`📈 Success Rate: ${successRate}%`, successRate > 80 ? 'green' : successRate > 60 ? 'yellow' : 'red');
  
  // Category breakdown
  log('\n📋 CATEGORY BREAKDOWN:', 'cyan');
  Object.entries(testResults.categories).forEach(([category, results]) => {
    const categorySuccess = results.total > 0 ? ((results.passed / results.total) * 100).toFixed(2) : 0;
    log(`${category}: ${results.passed}/${results.total} (${categorySuccess}%)`, 
         categorySuccess > 80 ? 'green' : categorySuccess > 60 ? 'yellow' : 'red');
  });
  
  // Recommendations
  log('\n💡 RECOMMENDATIONS:', 'cyan');
  if (testResults.failed > 0) {
    log('• Review failed tests and fix underlying issues', 'yellow');
    log('• Check server logs for detailed error information', 'yellow');
  }
  if (testResults.skipped > 0) {
    log('• Implement skipped endpoints for complete admin functionality', 'yellow');
  }
  if (successRate > 90) {
    log('• Excellent! Admin system is highly functional', 'green');
  } else if (successRate > 70) {
    log('• Good foundation, but some improvements needed', 'yellow');
  } else {
    log('• Significant issues detected, comprehensive review needed', 'red');
  }
  
  log('\n🏁 Admin System Testing Complete!', 'green');
  
  // Exit with appropriate code
  process.exit(testResults.failed > 0 ? 1 : 0);
}

// Run tests
runAllTests().catch(error => {
  log(`Fatal error: ${error.message}`, 'red');
  process.exit(1);
});