#!/usr/bin/env node

// Debug AI Service - Check if Moonshot is really being called

const API_URL = 'http://localhost:8080/api/v1';

async function testAI() {
  console.log('1. Logging in...');
  
  const loginRes = await fetch(`${API_URL}/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username: 'alice', password: 'secret' })
  });
  
  const loginData = await loginRes.json();
  if (!loginData.data?.token) {
    console.error('Login failed:', loginData);
    return;
  }
  
  const token = loginData.data.token;
  console.log('✓ Logged in successfully\n');
  
  console.log('2. Testing AI Inspiration endpoint...');
  const inspirationRes = await fetch(`${API_URL}/ai/inspiration`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${token}`
    },
    body: JSON.stringify({
      theme: '测试主题',
      count: 1
    })
  });
  
  const inspirationData = await inspirationRes.json();
  console.log('Response:', JSON.stringify(inspirationData, null, 2));
  
  // Check if it's fallback or real
  if (inspirationData.message?.includes('fallback')) {
    console.log('\n⚠️  Using FALLBACK data - Moonshot API not called');
    console.log('Possible reasons:');
    console.log('- Daily usage limit exceeded');
    console.log('- AI service error');
    console.log('- API key issue');
  } else {
    console.log('\n✅ Using REAL Moonshot API data');
  }
  
  // Test usage stats
  console.log('\n3. Checking AI usage stats...');
  const statsRes = await fetch(`${API_URL}/ai/stats`, {
    headers: { 'Authorization': `Bearer ${token}` }
  });
  
  const statsData = await statsRes.json();
  console.log('Usage stats:', JSON.stringify(statsData, null, 2));
}

testAI().catch(console.error);