#!/usr/bin/env node

/**
 * Direct Backend API Integration Test
 * Tests the follow system and privacy system APIs directly against the backend
 */

const axios = require('axios');

// Configuration
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
    url: `${BACKEND_URL}${API_VERSION}${endpoint}`,
    headers: {
      'Content-Type': 'application/json',
      ...(token && { 'Authorization': `Bearer ${token}` })
    },
    ...(data && { data })
  };

  try {
    const response = await axios(config);
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
  if (result.success) {
    console.log('✅ Backend is healthy:', result.data);
  } else {
    console.log('❌ Backend health check failed:', result.error);
  }
}

async function testUserRegistration() {
  console.log('\n🧪 Testing User Registration...');
  
  // Register first user
  const result1 = await apiRequest('POST', '/auth/register', testUser);
  if (result1.success) {
    console.log('✅ User 1 registered successfully');
    authToken1 = result1.data.data?.token || result1.data.token;
    userId1 = result1.data.data?.user?.id || result1.data.user?.id;
    console.log('   Token:', authToken1?.substring(0, 30) + '...');
    console.log('   User ID:', userId1);
  } else {
    console.log('❌ User 1 registration failed:', result1.error);
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
  const followResult = await apiRequest('POST', '/follow/users', {
    user_id: userId2,
    notification_enabled: true
  }, authToken1);
  
  if (followResult.success) {
    console.log('✅ Follow action successful');
    console.log('   Response:', JSON.stringify(followResult.data, null, 2));
  } else {
    console.log('❌ Follow action failed:', followResult.error);
    console.log('   Status:', followResult.status);
  }

  // Test 2: Get follow status
  console.log('\n📍 Test 2: Check follow status');
  const statusResult = await apiRequest('GET', `/follow/users/${userId2}/status`, null, authToken1);
  
  if (statusResult.success) {
    console.log('✅ Follow status retrieved');
    console.log('   Response:', JSON.stringify(statusResult.data, null, 2));
  } else {
    console.log('❌ Failed to get follow status:', statusResult.error);
    console.log('   Status:', statusResult.status);
  }

  // Test 3: Get followers list
  console.log('\n📍 Test 3: Get User 2\'s followers');
  const followersResult = await apiRequest('GET', `/follow/users/${userId2}/followers`, null, authToken2);
  
  if (followersResult.success) {
    console.log('✅ Followers list retrieved');
    console.log('   Response:', JSON.stringify(followersResult.data, null, 2));
  } else {
    console.log('❌ Failed to get followers:', followersResult.error);
    console.log('   Status:', followersResult.status);
  }

  // Test 4: Get following list
  console.log('\n📍 Test 4: Get User 1\'s following list');
  const followingResult = await apiRequest('GET', '/follow/following', null, authToken1);
  
  if (followingResult.success) {
    console.log('✅ Following list retrieved');
    console.log('   Response:', JSON.stringify(followingResult.data, null, 2));
  } else {
    console.log('❌ Failed to get following list:', followingResult.error);
    console.log('   Status:', followingResult.status);
  }

  // Test 5: Get follow statistics
  console.log('\n📍 Test 5: Get follow statistics');
  const statsResult = await apiRequest('GET', '/me/follow-stats', null, authToken1);
  
  if (statsResult.success) {
    console.log('✅ Follow stats retrieved');
    console.log('   Response:', JSON.stringify(statsResult.data, null, 2));
  } else {
    console.log('❌ Failed to get follow stats:', statsResult.error);
    console.log('   Status:', statsResult.status);
  }

  // Test 6: Unfollow user
  console.log('\n📍 Test 6: User 1 unfollows User 2');
  const unfollowResult = await apiRequest('DELETE', `/follow/users/${userId2}`, null, authToken1);
  
  if (unfollowResult.success) {
    console.log('✅ Unfollow action successful');
    console.log('   Response:', JSON.stringify(unfollowResult.data, null, 2));
  } else {
    console.log('❌ Unfollow action failed:', unfollowResult.error);
    console.log('   Status:', unfollowResult.status);
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
  const getSettingsResult = await apiRequest('GET', '/privacy/settings', null, authToken1);
  
  if (getSettingsResult.success) {
    console.log('✅ Privacy settings retrieved');
    console.log('   Response:', JSON.stringify(getSettingsResult.data, null, 2));
  } else {
    console.log('❌ Failed to get privacy settings:', getSettingsResult.error);
    console.log('   Status:', getSettingsResult.status);
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
  }, authToken1);
  
  if (updateSettingsResult.success) {
    console.log('✅ Privacy settings updated');
    console.log('   Response:', JSON.stringify(updateSettingsResult.data, null, 2));
  } else {
    console.log('❌ Failed to update privacy settings:', updateSettingsResult.error);
    console.log('   Status:', updateSettingsResult.status);
  }

  // Test 3: Check privacy permission
  console.log('\n📍 Test 3: Check privacy permission');
  const checkPrivacyResult = await apiRequest('GET', `/privacy/check/${userId2}?action=view_profile`, null, authToken1);
  
  if (checkPrivacyResult.success) {
    console.log('✅ Privacy check result');
    console.log('   Response:', JSON.stringify(checkPrivacyResult.data, null, 2));
  } else {
    console.log('❌ Failed to check privacy:', checkPrivacyResult.error);
    console.log('   Status:', checkPrivacyResult.status);
  }

  // Test 4: Block a user
  console.log('\n📍 Test 4: Block a user');
  const blockResult = await apiRequest('POST', '/privacy/block', {
    user_id: userId2
  }, authToken1);
  
  if (blockResult.success) {
    console.log('✅ User blocked successfully');
    console.log('   Response:', JSON.stringify(blockResult.data, null, 2));
  } else {
    console.log('❌ Failed to block user:', blockResult.error);
    console.log('   Status:', blockResult.status);
  }

  // Test 5: Get blocked users list
  console.log('\n📍 Test 5: Get blocked users list');
  const blockedListResult = await apiRequest('GET', '/privacy/blocked', null, authToken1);
  
  if (blockedListResult.success) {
    console.log('✅ Blocked users list retrieved');
    console.log('   Response:', JSON.stringify(blockedListResult.data, null, 2));
  } else {
    console.log('❌ Failed to get blocked users:', blockedListResult.error);
    console.log('   Status:', blockedListResult.status);
  }

  // Test 6: Unblock user
  console.log('\n📍 Test 6: Unblock user');
  const unblockResult = await apiRequest('DELETE', `/privacy/block/${userId2}`, null, authToken1);
  
  if (unblockResult.success) {
    console.log('✅ User unblocked successfully');
    console.log('   Response:', JSON.stringify(unblockResult.data, null, 2));
  } else {
    console.log('❌ Failed to unblock user:', unblockResult.error);
    console.log('   Status:', unblockResult.status);
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
  const profileResult = await apiRequest('GET', `/users/${testUser.username}/profile`, null, authToken1);
  
  if (profileResult.success) {
    console.log('✅ User profile retrieved');
    console.log('   Response:', JSON.stringify(profileResult.data, null, 2));
  } else {
    console.log('❌ Failed to get user profile:', profileResult.error);
    console.log('   Status:', profileResult.status);
  }

  // Test 2: Get user follow stats
  console.log('\n📍 Test 2: Get user follow stats');
  const followStatsResult = await apiRequest('GET', `/users/${testUser.username}/follow-stats`, null, authToken1);
  
  if (followStatsResult.success) {
    console.log('✅ Follow stats retrieved');
    console.log('   Response:', JSON.stringify(followStatsResult.data, null, 2));
  } else {
    console.log('❌ Failed to get follow stats:', followStatsResult.error);
    console.log('   Status:', followStatsResult.status);
  }

  return true;
}

// Main test runner
async function runTests() {
  console.log('🚀 Starting Backend API Integration Tests');
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
  }

  console.log('\n\n🎉 Backend API tests completed!');
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