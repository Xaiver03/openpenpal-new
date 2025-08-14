#!/usr/bin/env node

// Test to reset user's AI usage and try again

const API_URL = 'http://localhost:8080/api/v1';

async function resetAndTest() {
  console.log('Testing AI with fresh usage limits...\n');
  
  // Create a new user to ensure fresh limits
  const timestamp = Date.now();
  const newUser = {
    username: `testuser_${timestamp}`,
    password: 'testpass123',
    email: `test${timestamp}@example.com`,
    nickname: `Test User ${timestamp}`,
    school_code: 'PK5F01'  // 6 character OP code
  };
  
  console.log('1. Creating new test user...');
  const registerRes = await fetch(`${API_URL}/auth/register`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(newUser)
  });
  
  const registerData = await registerRes.json();
  if (!registerRes.ok) {
    console.error('Registration failed:', registerData);
    return;
  }
  console.log('✓ User created:', newUser.username);
  
  console.log('\n2. Logging in...');
  const loginRes = await fetch(`${API_URL}/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      username: newUser.username,
      password: newUser.password
    })
  });
  
  const loginData = await loginRes.json();
  if (!loginData.data?.token) {
    console.error('Login failed:', loginData);
    return;
  }
  
  const token = loginData.data.token;
  console.log('✓ Logged in successfully');
  
  console.log('\n3. Testing AI Inspiration with fresh limits...');
  const inspirationRes = await fetch(`${API_URL}/ai/inspiration`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    },
    body: JSON.stringify({
      theme: '未来科技',
      count: 1
    })
  });
  
  const inspirationData = await inspirationRes.json();
  console.log('Response:', JSON.stringify(inspirationData, null, 2));
  
  // Check if it's fallback or real
  if (inspirationData.message?.includes('fallback')) {
    console.log('\n⚠️  Still using FALLBACK - Issue is NOT usage limits');
    console.log('The AI service is failing to call Moonshot API');
  } else {
    console.log('\n✅ SUCCESS! Real Moonshot API response received');
  }
}

resetAndTest().catch(console.error);