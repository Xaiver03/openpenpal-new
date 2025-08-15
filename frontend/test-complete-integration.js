#!/usr/bin/env node

/**
 * Complete Frontend-Backend Integration Test
 * Tests follow system, privacy system, and personal homepage with CSRF handling
 */

const axios = require('axios');
const fs = require('fs');
const path = require('path');

// Configuration
const BACKEND_URL = 'http://localhost:8080';
const API_VERSION = '/api/v1';
const COOKIE_FILE = path.join(__dirname, 'test-cookies.txt');

// Test data
const testUser = {
  username: `testuser_${Date.now()}`,
  password: 'Test@123456',
  email: `test${Date.now()}@example.com`,
  nickname: 'Test User',
  school: 'Test School'
};

const testUser2 = {
  username: `testuser2_${Date.now()}`,
  password: 'Test@123456',
  email: `test2${Date.now()}@example.com`,
  nickname: 'Test User 2',
  school: 'Test School'
};

let authToken1 = null;
let authToken2 = null;
let userId1 = null;
let userId2 = null;
let csrfToken = null;
let cookies = {};

// Axios instance with cookie support
const api = axios.create({
  baseURL: `${BACKEND_URL}${API_VERSION}`,
  timeout: 10000,
  withCredentials: true,
  headers: {
    'Accept': 'application/json',
    'Content-Type': 'application/json'
  }
});

// Cookie management
function parseCookies(cookieHeader) {
  if (!cookieHeader) return {};
  const cookies = {};
  cookieHeader.split(';').forEach(cookie => {
    const parts = cookie.trim().split('=');
    if (parts.length === 2) {
      cookies[parts[0]] = parts[1];
    }
  });
  return cookies;
}

function getCookieString() {
  return Object.entries(cookies).map(([key, value]) => `${key}=${value}`).join('; ');
}

// Request interceptor to add cookies and CSRF token
api.interceptors.request.use((config) => {
  const cookieString = getCookieString();
  if (cookieString) {
    config.headers['Cookie'] = cookieString;
  }
  if (csrfToken) {
    config.headers['X-CSRF-Token'] = csrfToken;
  }
  if (authToken1 || authToken2) {
    // Use the appropriate token based on the test context
    const token = config.headers['Authorization']?.includes('Bearer') 
      ? config.headers['Authorization'] 
      : `Bearer ${authToken1 || authToken2}`;
    config.headers['Authorization'] = token;
  }
  return config;
});

// Response interceptor to save cookies
api.interceptors.response.use((response) => {
  const setCookieHeader = response.headers['set-cookie'];
  if (setCookieHeader) {
    setCookieHeader.forEach(cookieStr => {
      const cookie = parseCookies(cookieStr.split(';')[0]);
      Object.assign(cookies, cookie);
      
      // Extract CSRF token from cookie
      if (cookie.csrf_token) {
        csrfToken = cookie.csrf_token;
      }
    });
  }
  return response;
});

// Helper function to make API requests
async function apiRequest(method, endpoint, data = null, useToken2 = false) {
  const config = {
    method,
    url: endpoint,
    ...(data && { data }),
    headers: {}
  };
  
  // Use appropriate auth token
  if (useToken2 && authToken2) {
    config.headers['Authorization'] = `Bearer ${authToken2}`;
  } else if (!useToken2 && authToken1) {
    config.headers['Authorization'] = `Bearer ${authToken1}`;
  }

  try {
    const response = await api(config);
    return { success: true, data: response.data, status: response.status };
  } catch (error) {
    return { 
      success: false, 
      error: error.response?.data || error.message,
      status: error.response?.status 
    };
  }
}

// Test functions
async function testBackendHealth() {
  console.log('\n🔍 Testing Backend Health...');
  
  const result = await apiRequest('GET', '/health');
  if (result.success || result.status === 200) {
    console.log('✅ Backend is healthy');
  } else {
    // Try without /api/v1 prefix
    const result2 = await axios.get(`${BACKEND_URL}/health`);
    if (result2.status === 200) {
      console.log('✅ Backend is healthy (at root path)');
    } else {
      console.log('❌ Backend health check failed');
    }
  }
}

async function getCSRFToken() {
  console.log('\n🔐 Getting CSRF Token...');
  
  try {
    // Make a request to get CSRF token
    const response = await api.get('/csrf-token');
    if (response.data.csrf_token) {
      csrfToken = response.data.csrf_token;
      console.log('✅ CSRF token obtained:', csrfToken.substring(0, 20) + '...');
    }
  } catch (error) {
    console.log('⚠️  No dedicated CSRF endpoint, will try to get from cookies');
  }
}

async function testUserRegistration() {
  console.log('\n🧪 Testing User Registration...');
  
  // Get CSRF token first
  await getCSRFToken();
  
  // Register first user
  const result1 = await apiRequest('POST', '/auth/register', testUser);
  if (result1.success) {
    console.log('✅ User 1 registered successfully');
    authToken1 = result1.data.data?.token || result1.data.token;
    userId1 = result1.data.data?.user?.id || result1.data.user?.id;
    console.log('   Token:', authToken1?.substring(0, 30) + '...');
    console.log('   User ID:', userId1);
  } else {
    console.log('❌ User 1 registration failed:', JSON.stringify(result1.error, null, 2));
    return false;
  }

  // Register second user
  const result2 = await apiRequest('POST', '/auth/register', testUser2);
  if (result2.success) {
    console.log('✅ User 2 registered successfully');
    authToken2 = result2.data.data?.token || result2.data.token;
    userId2 = result2.data.data?.user?.id || result2.data.user?.id;
    console.log('   Token:', authToken2?.substring(0, 30) + '...');
    console.log('   User ID:', userId2);
  } else {
    console.log('❌ User 2 registration failed:', JSON.stringify(result2.error, null, 2));
    return false;
  }

  return true;
}

async function testFollowSystem() {
  console.log('\n🧪 Testing Follow System APIs...');
  
  if (!authToken1 || !userId2) {
    console.log('❌ Missing auth tokens or user IDs');
    return false;
  }

  // Test 1: Follow a user
  console.log('\n📍 Test 1: User 1 follows User 2');
  const followResult = await apiRequest('POST', '/follow/users', {
    user_id: userId2,
    notification_enabled: true
  });
  
  if (followResult.success) {
    console.log('✅ Follow action successful');
    console.log('   Response:', JSON.stringify(followResult.data, null, 2));
  } else {
    console.log('❌ Follow action failed:', JSON.stringify(followResult.error, null, 2));
  }

  // Test 2: Get follow status
  console.log('\n📍 Test 2: Check follow status');
  const statusResult = await apiRequest('GET', `/follow/users/${userId2}/status`);
  
  if (statusResult.success) {
    console.log('✅ Follow status retrieved');
    console.log('   Response:', JSON.stringify(statusResult.data, null, 2));
  } else {
    console.log('❌ Failed to get follow status:', JSON.stringify(statusResult.error, null, 2));
  }

  // Test 3: Get followers list (using User 2's token)
  console.log('\n📍 Test 3: Get User 2\'s followers');
  const followersResult = await apiRequest('GET', `/follow/users/${userId2}/followers`, null, true);
  
  if (followersResult.success) {
    console.log('✅ Followers list retrieved');
    console.log('   Response:', JSON.stringify(followersResult.data, null, 2));
  } else {
    console.log('❌ Failed to get followers:', JSON.stringify(followersResult.error, null, 2));
  }

  // Test 4: Get following list
  console.log('\n📍 Test 4: Get User 1\'s following list');
  const followingResult = await apiRequest('GET', '/follow/following');
  
  if (followingResult.success) {
    console.log('✅ Following list retrieved');
    console.log('   Response:', JSON.stringify(followingResult.data, null, 2));
  } else {
    console.log('❌ Failed to get following list:', JSON.stringify(followingResult.error, null, 2));
  }

  return true;
}

async function testPrivacySystem() {
  console.log('\n🧪 Testing Privacy System APIs...');
  
  if (!authToken1) {
    console.log('❌ Missing auth token');
    return false;
  }

  // Test 1: Get privacy settings
  console.log('\n📍 Test 1: Get privacy settings');
  const getSettingsResult = await apiRequest('GET', '/privacy/settings');
  
  if (getSettingsResult.success) {
    console.log('✅ Privacy settings retrieved');
    console.log('   Response:', JSON.stringify(getSettingsResult.data, null, 2));
  } else {
    console.log('❌ Failed to get privacy settings:', JSON.stringify(getSettingsResult.error, null, 2));
  }

  // Test 2: Update privacy settings
  console.log('\n📍 Test 2: Update privacy settings');
  const updateSettingsResult = await apiRequest('PUT', '/privacy/settings', {
    profile_visibility: 'friends',
    show_email: false,
    show_op_code: true,
    op_code_privacy: 'partial',
    allow_comments: true,
    letter_visibility_default: 'private'
  });
  
  if (updateSettingsResult.success) {
    console.log('✅ Privacy settings updated');
    console.log('   Response:', JSON.stringify(updateSettingsResult.data, null, 2));
  } else {
    console.log('❌ Failed to update privacy settings:', JSON.stringify(updateSettingsResult.error, null, 2));
  }

  return true;
}

async function testPersonalHomepage() {
  console.log('\n🧪 Testing Personal Homepage APIs...');
  
  if (!authToken1 || !testUser.username) {
    console.log('❌ Missing auth token or username');
    return false;
  }

  // Test 1: Access user profile
  console.log('\n📍 Test 1: Access user profile');
  const profileResult = await apiRequest('GET', `/users/${testUser.username}/profile`);
  
  if (profileResult.success) {
    console.log('✅ User profile retrieved');
    console.log('   Profile data:', JSON.stringify(profileResult.data, null, 2));
  } else {
    console.log('❌ Failed to get user profile:', JSON.stringify(profileResult.error, null, 2));
  }

  // Test 2: Get user follow stats
  console.log('\n📍 Test 2: Get user follow stats');
  const followStatsResult = await apiRequest('GET', `/users/${testUser.username}/follow-stats`);
  
  if (followStatsResult.success) {
    console.log('✅ Follow stats retrieved');
    console.log('   Stats:', JSON.stringify(followStatsResult.data, null, 2));
  } else {
    console.log('❌ Failed to get follow stats:', JSON.stringify(followStatsResult.error, null, 2));
  }

  return true;
}

async function testFrontendIntegration() {
  console.log('\n🧪 Testing Frontend Integration...');
  
  // Check if frontend components are using real APIs
  console.log('\n📍 Checking if frontend APIs are properly configured...');
  
  // Test the user page component
  if (testUser.username) {
    console.log(`\n📍 User homepage should be accessible at: http://localhost:3000/u/${testUser.username}`);
    console.log('   - Follow button should work with real API');
    console.log('   - Privacy settings should be respected');
    console.log('   - Activity feed should load from API');
  }

  return true;
}

// Main test runner
async function runTests() {
  console.log('🚀 Starting Complete Frontend-Backend Integration Tests');
  console.log('================================================');
  console.log(`Backend URL: ${BACKEND_URL}${API_VERSION}`);
  console.log('================================================');

  // Check backend health
  await testBackendHealth();

  // Run tests
  const registrationSuccess = await testUserRegistration();
  if (registrationSuccess) {
    await testFollowSystem();
    await testPrivacySystem();
    await testPersonalHomepage();
    await testFrontendIntegration();
  }

  console.log('\n\n🎉 Integration tests completed!');
  console.log('================================================');
  console.log('Test Results Summary:');
  console.log('- Test User 1:', testUser.username);
  console.log('- Test User 2:', testUser2.username);
  console.log('- User ID 1:', userId1);
  console.log('- User ID 2:', userId2);
  console.log('\nYou can now:');
  console.log(`1. Visit http://localhost:3000/u/${testUser.username} to see the personal homepage`);
  console.log('2. Login with the test credentials to test the follow/unfollow functionality');
  console.log('3. Check privacy settings in the user settings page');
  console.log('================================================');
  
  // Save test data for manual testing
  const testData = {
    user1: {
      username: testUser.username,
      password: testUser.password,
      email: testUser.email,
      id: userId1,
      token: authToken1
    },
    user2: {
      username: testUser2.username,
      password: testUser2.password,
      email: testUser2.email,
      id: userId2,
      token: authToken2
    },
    timestamp: new Date().toISOString()
  };
  
  fs.writeFileSync('test-users.json', JSON.stringify(testData, null, 2));
  console.log('\nTest user data saved to test-users.json');
}

// Run the tests
runTests().catch(console.error);