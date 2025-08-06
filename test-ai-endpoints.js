#!/usr/bin/env node

const axios = require('axios');

// Test configuration
const BASE_URL = 'http://localhost:8080';
const endpoints = [
  { method: 'GET', path: '/api/ai/daily-inspiration', name: 'Daily Inspiration (alias)' },
  { method: 'GET', path: '/api/v1/ai/daily-inspiration', name: 'Daily Inspiration (direct)' },
  { method: 'GET', path: '/api/ai/stats', name: 'AI Stats (alias)' },
  { method: 'GET', path: '/api/v1/ai/stats', name: 'AI Stats (direct)' },
  { method: 'POST', path: '/api/ai/inspiration', name: 'Writing Inspiration (alias)', data: { theme: "日常感悟", count: 3 } },
  { method: 'POST', path: '/api/v1/ai/inspiration', name: 'Writing Inspiration (direct)', data: { theme: "友情", count: 2 } }
];

// Test function
async function testEndpoint(endpoint) {
  try {
    console.log(`\n🧪 Testing: ${endpoint.name}`);
    console.log(`📍 ${endpoint.method} ${BASE_URL}${endpoint.path}`);
    
    const config = {
      method: endpoint.method,
      url: `${BASE_URL}${endpoint.path}`,
      headers: {
        'Content-Type': 'application/json'
      }
    };
    
    if (endpoint.data) {
      config.data = endpoint.data;
      console.log(`📦 Request body:`, JSON.stringify(endpoint.data, null, 2));
    }
    
    const response = await axios(config);
    
    console.log(`✅ Status: ${response.status}`);
    console.log(`📊 Response:`, JSON.stringify(response.data, null, 2));
    
    return { success: true, endpoint: endpoint.name };
  } catch (error) {
    console.error(`❌ Error: ${error.response ? error.response.status : error.message}`);
    if (error.response && error.response.data) {
      console.error(`📊 Error Response:`, JSON.stringify(error.response.data, null, 2));
    }
    return { success: false, endpoint: endpoint.name, error: error.message };
  }
}

// Main test runner
async function runTests() {
  console.log('🚀 Starting AI endpoint tests...\n');
  
  const results = [];
  for (const endpoint of endpoints) {
    const result = await testEndpoint(endpoint);
    results.push(result);
  }
  
  // Summary
  console.log('\n\n📈 Test Summary:');
  console.log('================');
  const successful = results.filter(r => r.success).length;
  const failed = results.filter(r => !r.success).length;
  
  console.log(`✅ Successful: ${successful}`);
  console.log(`❌ Failed: ${failed}`);
  
  if (failed > 0) {
    console.log('\nFailed endpoints:');
    results.filter(r => !r.success).forEach(r => {
      console.log(`  - ${r.endpoint}: ${r.error}`);
    });
  }
  
  console.log('\n✨ Test complete!');
}

// Run tests
runTests().catch(console.error);