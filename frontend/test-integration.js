#!/usr/bin/env node

/**
 * Frontend-Backend Integration Test Script
 * Tests the follow system and privacy system APIs through the frontend
 */

const axios = require('axios');

// Configuration
const FRONTEND_URL = 'http://localhost:3000';
const BACKEND_URL = 'http://localhost:8080';
const API_VERSION = '/api/v1';

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

// Helper function to make API requests
async function apiRequest(method, endpoint, data = null, token = null) {
  const config = {
    method,
    url: `${FRONTEND_URL}${endpoint}`,
    headers: {
      'Content-Type': 'application/json',
      ...(token && { 'Authorization': `Bearer ${token}` })
    },
    ...(data && { data })
  };

  try {
    const response = await axios(config);
    return { success: true, data: response.data };
  } catch (error) {
    return { 
      success: false, 
      error: error.response?.data || error.message,
      status: error.response?.status 
    };
  }
}

// Test functions
async function testUserRegistration() {
  console.log('\n🧪 Testing User Registration...');
  
  // Register first user
  const result1 = await apiRequest('POST', '/api/auth/register', testUser);
  if (result1.success) {
    console.log('✅ User 1 registered successfully');
    authToken1 = result1.data.data?.token || result1.data.token;
    userId1 = result1.data.data?.user?.id || result1.data.user?.id;
  } else {
    console.log('❌ User 1 registration failed:', result1.error);
    return false;
  }

  // Register second user
  const result2 = await apiRequest('POST', '/api/auth/register', testUser2);
  if (result2.success) {
    console.log('✅ User 2 registered successfully');
    authToken2 = result2.data.data?.token || result2.data.token;
    userId2 = result2.data.data?.user?.id || result2.data.user?.id;
  } else {
    console.log('❌ User 2 registration failed:', result2.error);
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
  const followResult = await apiRequest('POST', '/api/v1/follow/users', {
    user_id: userId2,
    notification_enabled: true
  }, authToken1);
  
  if (followResult.success) {
    console.log('✅ Follow action successful:', followResult.data);
  } else {
    console.log('❌ Follow action failed:', followResult.error);
  }

  // Test 2: Get follow status
  console.log('\n📍 Test 2: Check follow status');
  const statusResult = await apiRequest('GET', `/api/v1/follow/users/${userId2}/status`, null, authToken1);
  
  if (statusResult.success) {
    console.log('✅ Follow status retrieved:', statusResult.data);
  } else {
    console.log('❌ Failed to get follow status:', statusResult.error);
  }

  // Test 3: Get followers list
  console.log('\n📍 Test 3: Get User 2\'s followers');
  const followersResult = await apiRequest('GET', `/api/v1/follow/users/${userId2}/followers`, null, authToken2);
  
  if (followersResult.success) {
    console.log('✅ Followers list retrieved:', followersResult.data);
  } else {
    console.log('❌ Failed to get followers:', followersResult.error);
  }

  // Test 4: Get following list
  console.log('\n📍 Test 4: Get User 1\'s following list');
  const followingResult = await apiRequest('GET', '/api/v1/follow/following', null, authToken1);
  
  if (followingResult.success) {
    console.log('✅ Following list retrieved:', followingResult.data);
  } else {
    console.log('❌ Failed to get following list:', followingResult.error);
  }

  // Test 5: Get follow statistics
  console.log('\n📍 Test 5: Get follow statistics');
  const statsResult = await apiRequest('GET', '/api/v1/me/follow-stats', null, authToken1);
  
  if (statsResult.success) {
    console.log('✅ Follow stats retrieved:', statsResult.data);
  } else {
    console.log('❌ Failed to get follow stats:', statsResult.error);
  }

  // Test 6: Get user suggestions
  console.log('\n📍 Test 6: Get user suggestions');
  const suggestionsResult = await apiRequest('GET', '/api/v1/follow/suggestions?limit=5', null, authToken1);
  
  if (suggestionsResult.success) {
    console.log('✅ User suggestions retrieved:', suggestionsResult.data);
  } else {
    console.log('❌ Failed to get suggestions:', suggestionsResult.error);
  }

  // Test 7: Unfollow user
  console.log('\n📍 Test 7: User 1 unfollows User 2');
  const unfollowResult = await apiRequest('DELETE', `/api/v1/follow/users/${userId2}`, null, authToken1);
  
  if (unfollowResult.success) {
    console.log('✅ Unfollow action successful:', unfollowResult.data);
  } else {
    console.log('❌ Unfollow action failed:', unfollowResult.error);
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
  const getSettingsResult = await apiRequest('GET', '/api/v1/privacy/settings', null, authToken1);
  
  if (getSettingsResult.success) {
    console.log('✅ Privacy settings retrieved:', getSettingsResult.data);
  } else {
    console.log('❌ Failed to get privacy settings:', getSettingsResult.error);
  }

  // Test 2: Update privacy settings
  console.log('\n📍 Test 2: Update privacy settings');
  const updateSettingsResult = await apiRequest('PUT', '/api/v1/privacy/settings', {
    profile_visibility: 'friends',
    show_email: false,
    show_op_code: true,
    op_code_privacy: 'partial',
    allow_comments: true,
    letter_visibility_default: 'private'
  }, authToken1);
  
  if (updateSettingsResult.success) {
    console.log('✅ Privacy settings updated:', updateSettingsResult.data);
  } else {
    console.log('❌ Failed to update privacy settings:', updateSettingsResult.error);
  }

  // Test 3: Check privacy permission
  console.log('\n📍 Test 3: Check privacy permission');
  const checkPrivacyResult = await apiRequest('GET', `/api/v1/privacy/check/${userId2}?action=view_profile`, null, authToken1);
  
  if (checkPrivacyResult.success) {
    console.log('✅ Privacy check result:', checkPrivacyResult.data);
  } else {
    console.log('❌ Failed to check privacy:', checkPrivacyResult.error);
  }

  // Test 4: Block a user
  console.log('\n📍 Test 4: Block a user');
  const blockResult = await apiRequest('POST', '/api/v1/privacy/block', {
    user_id: userId2
  }, authToken1);
  
  if (blockResult.success) {
    console.log('✅ User blocked successfully:', blockResult.data);
  } else {
    console.log('❌ Failed to block user:', blockResult.error);
  }

  // Test 5: Get blocked users list
  console.log('\n📍 Test 5: Get blocked users list');
  const blockedListResult = await apiRequest('GET', '/api/v1/privacy/blocked', null, authToken1);
  
  if (blockedListResult.success) {
    console.log('✅ Blocked users list retrieved:', blockedListResult.data);
  } else {
    console.log('❌ Failed to get blocked users:', blockedListResult.error);
  }

  // Test 6: Unblock user
  console.log('\n📍 Test 6: Unblock user');
  const unblockResult = await apiRequest('DELETE', `/api/v1/privacy/block/${userId2}`, null, authToken1);
  
  if (unblockResult.success) {
    console.log('✅ User unblocked successfully:', unblockResult.data);
  } else {
    console.log('❌ Failed to unblock user:', unblockResult.error);
  }

  return true;
}

async function testPersonalHomepage() {
  console.log('\n🧪 Testing Personal Homepage...');
  
  if (!authToken1 || !testUser.username) {
    console.log('❌ Missing auth token or username');
    return false;
  }

  // Test 1: Access user profile page
  console.log('\n📍 Test 1: Access user profile API');
  const profileResult = await apiRequest('GET', `/api/users/${testUser.username}/profile`, null, authToken1);
  
  if (profileResult.success) {
    console.log('✅ User profile retrieved:', profileResult.data);
  } else {
    console.log('❌ Failed to get user profile:', profileResult.error);
  }

  // Test 2: Get user follow stats
  console.log('\n📍 Test 2: Get user follow stats');
  const followStatsResult = await apiRequest('GET', `/api/users/${testUser.username}/follow-stats`, null, authToken1);
  
  if (followStatsResult.success) {
    console.log('✅ Follow stats retrieved:', followStatsResult.data);
  } else {
    console.log('❌ Failed to get follow stats:', followStatsResult.error);
  }

  // Test 3: Get user's public letters
  console.log('\n📍 Test 3: Get user\'s public letters');
  const lettersResult = await apiRequest('GET', `/api/users/${testUser.username}/letters?public=true`, null, authToken1);
  
  if (lettersResult.success) {
    console.log('✅ Public letters retrieved:', lettersResult.data);
  } else {
    console.log('❌ Failed to get public letters:', lettersResult.error);
  }

  return true;
}

async function checkFrontendComponents() {
  console.log('\n🧪 Checking Frontend Components...');
  
  // Check if frontend is using real API endpoints
  console.log('\n📍 Checking API configuration...');
  
  try {
    // Test if the frontend proxy is working
    const proxyTest = await apiRequest('GET', '/api/health');
    if (proxyTest.success || proxyTest.status === 200) {
      console.log('✅ Frontend API proxy is working');
    } else {
      console.log('❌ Frontend API proxy might not be configured correctly');
    }
  } catch (error) {
    console.log('❌ Error checking frontend proxy:', error.message);
  }

  return true;
}

// Main test runner
async function runTests() {
  console.log('🚀 Starting Frontend-Backend Integration Tests');
  console.log('================================================');
  console.log(`Frontend URL: ${FRONTEND_URL}`);
  console.log(`Backend URL: ${BACKEND_URL}`);
  console.log('================================================');

  // Check if servers are running
  console.log('\n🔍 Checking server status...');
  
  try {
    await axios.get(`${FRONTEND_URL}`);
    console.log('✅ Frontend server is running');
  } catch (error) {
    console.log('❌ Frontend server is not accessible');
    return;
  }

  try {
    await axios.get(`${BACKEND_URL}/health`);
    console.log('✅ Backend server is running');
  } catch (error) {
    console.log('❌ Backend server is not accessible');
    return;
  }

  // Run tests
  await testUserRegistration();
  await testFollowSystem();
  await testPrivacySystem();
  await testPersonalHomepage();
  await checkFrontendComponents();

  console.log('\n\n🎉 Integration tests completed!');
  console.log('================================================');
  console.log('Summary:');
  console.log('- Test User 1:', testUser.username);
  console.log('- Test User 2:', testUser2.username);
  console.log('- User ID 1:', userId1);
  console.log('- User ID 2:', userId2);
  console.log('================================================');
}

// Run the tests
runTests().catch(console.error);