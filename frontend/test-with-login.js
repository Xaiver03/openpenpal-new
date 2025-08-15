#!/usr/bin/env node

/**
 * Test with Login Flow - Tests backend APIs using existing test users
 */

const axios = require('axios');
const fs = require('fs');

// Configuration
const BACKEND_URL = 'http://localhost:8080';
const API_VERSION = '/api/v1';

// Use the existing test users
const TEST_USERS = {
  alice: {
    username: 'alice_test',
    password: 'Test@123456',
    email: 'alice@test.com'
  },
  bob: {
    username: 'bob_test', 
    password: 'Test@123456',
    email: 'bob@test.com'
  }
};

let authData1 = null;
let authData2 = null;
let csrfToken = null;
let cookies = {};

// Axios instance
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

// Request interceptor
api.interceptors.request.use((config) => {
  const cookieString = getCookieString();
  if (cookieString) {
    config.headers['Cookie'] = cookieString;
  }
  if (csrfToken) {
    config.headers['X-CSRF-Token'] = csrfToken;
  }
  return config;
});

// Response interceptor
api.interceptors.response.use((response) => {
  const setCookieHeader = response.headers['set-cookie'];
  if (setCookieHeader) {
    setCookieHeader.forEach(cookieStr => {
      const cookie = parseCookies(cookieStr.split(';')[0]);
      Object.assign(cookies, cookie);
      
      // Extract CSRF token
      if (cookie['csrf-token']) {
        csrfToken = cookie['csrf-token'];
        console.log('   CSRF token updated from cookie');
      }
    });
  }
  return response;
});

// Test functions
async function loginUser(username, password) {
  console.log(`\n🔐 Logging in as ${username}...`);
  
  try {
    const response = await api.post('/auth/login', { username, password });
    const data = response.data;
    
    if (data.success || data.code === 0) {
      const userData = data.data || data;
      console.log(`✅ Login successful for ${username}`);
      console.log(`   User ID: ${userData.user?.id}`);
      console.log(`   Token: ${userData.token?.substring(0, 30)}...`);
      return {
        token: userData.token,
        user: userData.user
      };
    }
  } catch (error) {
    console.log(`❌ Login failed for ${username}:`, error.response?.data || error.message);
    return null;
  }
}

async function testFollowSystem() {
  console.log('\n🧪 Testing Follow System APIs...');
  
  if (!authData1 || !authData2) {
    console.log('❌ Missing auth data');
    return false;
  }

  // Set auth header for user 1
  api.defaults.headers['Authorization'] = `Bearer ${authData1.token}`;

  // Test 1: Follow user 2
  console.log('\n📍 Test 1: Alice follows Bob');
  try {
    const response = await api.post('/follow/users', {
      user_id: authData2.user.id,
      notification_enabled: true
    });
    console.log('✅ Follow action successful:', response.data);
  } catch (error) {
    console.log('❌ Follow action failed:', error.response?.data || error.message);
  }

  // Test 2: Get follow status
  console.log('\n📍 Test 2: Check follow status');
  try {
    const response = await api.get(`/follow/users/${authData2.user.id}/status`);
    console.log('✅ Follow status:', response.data);
  } catch (error) {
    console.log('❌ Failed to get follow status:', error.response?.data || error.message);
  }

  // Test 3: Get followers (switch to Bob's token)
  console.log('\n📍 Test 3: Get Bob\'s followers');
  api.defaults.headers['Authorization'] = `Bearer ${authData2.token}`;
  try {
    const response = await api.get(`/follow/followers`);
    console.log('✅ Followers list:', response.data);
  } catch (error) {
    console.log('❌ Failed to get followers:', error.response?.data || error.message);
  }

  // Test 4: Get following (switch back to Alice's token)
  console.log('\n📍 Test 4: Get Alice\'s following list');
  api.defaults.headers['Authorization'] = `Bearer ${authData1.token}`;
  try {
    const response = await api.get('/follow/following');
    console.log('✅ Following list:', response.data);
  } catch (error) {
    console.log('❌ Failed to get following:', error.response?.data || error.message);
  }

  // Test 5: Unfollow
  console.log('\n📍 Test 5: Alice unfollows Bob');
  try {
    const response = await api.delete(`/follow/users/${authData2.user.id}`);
    console.log('✅ Unfollow successful:', response.data);
  } catch (error) {
    console.log('❌ Unfollow failed:', error.response?.data || error.message);
  }

  return true;
}

async function testPrivacySystem() {
  console.log('\n🧪 Testing Privacy System APIs...');
  
  if (!authData1) {
    console.log('❌ Missing auth data');
    return false;
  }

  // Use Alice's token
  api.defaults.headers['Authorization'] = `Bearer ${authData1.token}`;

  // Test 1: Get privacy settings
  console.log('\n📍 Test 1: Get privacy settings');
  try {
    const response = await api.get('/privacy/settings');
    console.log('✅ Privacy settings:', response.data);
  } catch (error) {
    console.log('❌ Failed to get privacy settings:', error.response?.data || error.message);
  }

  // Test 2: Update privacy settings
  console.log('\n📍 Test 2: Update privacy settings');
  try {
    const response = await api.put('/privacy/settings', {
      profile_visibility: 'friends',
      show_email: false,
      show_op_code: true,
      op_code_privacy: 'partial',
      allow_comments: true,
      letter_visibility_default: 'private'
    });
    console.log('✅ Privacy settings updated:', response.data);
  } catch (error) {
    console.log('❌ Failed to update privacy settings:', error.response?.data || error.message);
  }

  // Test 3: Block user
  console.log('\n📍 Test 3: Block a user');
  try {
    const response = await api.post('/privacy/block', {
      user_id: authData2.user.id
    });
    console.log('✅ User blocked:', response.data);
  } catch (error) {
    console.log('❌ Failed to block user:', error.response?.data || error.message);
  }

  // Test 4: Get blocked users
  console.log('\n📍 Test 4: Get blocked users');
  try {
    const response = await api.get('/privacy/blocked');
    console.log('✅ Blocked users:', response.data);
  } catch (error) {
    console.log('❌ Failed to get blocked users:', error.response?.data || error.message);
  }

  // Test 5: Unblock user
  console.log('\n📍 Test 5: Unblock user');
  try {
    const response = await api.delete(`/privacy/block/${authData2.user.id}`);
    console.log('✅ User unblocked:', response.data);
  } catch (error) {
    console.log('❌ Failed to unblock user:', error.response?.data || error.message);
  }

  return true;
}

async function testPersonalHomepage() {
  console.log('\n🧪 Testing Personal Homepage APIs...');
  
  if (!authData1) {
    console.log('❌ Missing auth data');
    return false;
  }

  // Test 1: Get user profile
  console.log('\n📍 Test 1: Get Alice\'s profile');
  try {
    const response = await api.get(`/users/${authData1.user.username}/profile`);
    console.log('✅ User profile:', response.data);
  } catch (error) {
    console.log('❌ Failed to get profile:', error.response?.data || error.message);
  }

  // Test 2: Get follow stats
  console.log('\n📍 Test 2: Get follow stats');
  try {
    const response = await api.get(`/users/${authData1.user.username}/follow-stats`);
    console.log('✅ Follow stats:', response.data);
  } catch (error) {
    console.log('❌ Failed to get follow stats:', error.response?.data || error.message);
  }

  return true;
}

async function testFrontendAccess() {
  console.log('\n🌐 Testing Frontend Access...');
  
  if (!authData1 || !authData2) {
    console.log('❌ Missing auth data');
    return false;
  }

  console.log('\n📍 You can now test the frontend with these URLs:');
  console.log(`   1. Alice's profile: http://localhost:3000/u/${authData1.user.username}`);
  console.log(`   2. Bob's profile: http://localhost:3000/u/${authData2.user.username}`);
  console.log('\n📍 Login credentials:');
  console.log(`   - Alice: ${TEST_USERS.alice.username} / ${TEST_USERS.alice.password}`);
  console.log(`   - Bob: ${TEST_USERS.bob.username} / ${TEST_USERS.bob.password}`);

  return true;
}

// Main test runner
async function runTests() {
  console.log('🚀 Starting Backend API Tests with Login Flow');
  console.log('================================================');
  console.log(`Backend URL: ${BACKEND_URL}${API_VERSION}`);
  console.log('================================================');

  // Login both test users
  authData1 = await loginUser(TEST_USERS.alice.username, TEST_USERS.alice.password);
  authData2 = await loginUser(TEST_USERS.bob.username, TEST_USERS.bob.password);

  if (authData1 && authData2) {
    // Run API tests
    await testFollowSystem();
    await testPrivacySystem();
    await testPersonalHomepage();
    await testFrontendAccess();

    // Save session data
    const sessionData = {
      alice: {
        ...TEST_USERS.alice,
        id: authData1.user.id,
        token: authData1.token
      },
      bob: {
        ...TEST_USERS.bob,
        id: authData2.user.id,
        token: authData2.token
      },
      csrfToken: csrfToken,
      timestamp: new Date().toISOString()
    };
    
    fs.writeFileSync('test-session.json', JSON.stringify(sessionData, null, 2));
    console.log('\n✅ Test session data saved to test-session.json');
  } else {
    console.log('\n❌ Failed to login test users. Make sure the backend has been seeded with test data.');
  }

  console.log('\n🎉 Tests completed!');
}

// Run the tests
runTests().catch(console.error);