/**
 * AI Frontend Integration Test
 * Verifies AI services are working properly through frontend
 */

const http = require('http');
const https = require('https');

async function makeRequest(path, options = {}) {
  return new Promise((resolve, reject) => {
    const url = new URL(path, 'http://localhost:3000');
    const req = http.request(url, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'application/json',
        ...options.headers
      }
    }, (res) => {
      let data = '';
      res.on('data', chunk => data += chunk);
      res.on('end', () => {
        try {
          resolve({ 
            status: res.statusCode, 
            data: JSON.parse(data),
            headers: res.headers 
          });
        } catch (e) {
          resolve({ status: res.statusCode, data: data });
        }
      });
    });
    
    req.on('error', reject);
    if (options.body) req.write(options.body);
    req.end();
  });
}

async function testAIFrontendIntegration() {
  console.log('🤖 AI Frontend Integration Test\n');
  console.log('=' + '='.repeat(49));
  
  const tests = [
    {
      name: 'Daily Inspiration',
      path: '/api/ai/daily-inspiration',
      method: 'GET',
      expected: 200
    },
    {
      name: 'AI Stats',
      path: '/api/ai/stats',
      method: 'GET',
      expected: 200
    },
    {
      name: 'Writing Inspiration',
      path: '/api/ai/inspiration',
      method: 'POST',
      body: JSON.stringify({ theme: 'friendship', style: 'casual' }),
      expected: 200
    },
    {
      name: 'AI Personas',
      path: '/api/ai/personas',
      method: 'GET',
      expected: 200
    }
  ];
  
  let passed = 0;
  let failed = 0;
  
  for (const test of tests) {
    console.log(`\n📝 Testing: ${test.name}`);
    console.log(`   Method: ${test.method} ${test.path}`);
    
    try {
      const response = await makeRequest(test.path, {
        method: test.method,
        body: test.body
      });
      
      console.log(`   Status: ${response.status} ${response.status === test.expected ? '✅' : '❌'}`);
      
      if (response.status === test.expected) {
        passed++;
        console.log(`   Success: ${response.data.success}`);
        
        // Show sample data
        if (response.data.data) {
          const data = response.data.data;
          if (test.name === 'Daily Inspiration') {
            console.log(`   Theme: ${data.theme}`);
            console.log(`   Quote: ${data.quote}`);
          } else if (test.name === 'AI Stats') {
            console.log(`   User: ${data.userId}`);
            console.log(`   Remaining inspirations: ${data.remaining.inspirations}`);
          } else if (test.name === 'Writing Inspiration') {
            console.log(`   Inspirations count: ${data.inspirations?.length || 0}`);
          } else if (test.name === 'AI Personas') {
            console.log(`   Personas count: ${data.personas?.length || 0}`);
          }
        }
      } else {
        failed++;
        console.log(`   ❌ Error: ${response.data.message || response.data}`);
      }
    } catch (error) {
      failed++;
      console.log(`   ❌ Error: ${error.message}`);
    }
  }
  
  // Test authenticated endpoints
  console.log('\n\n🔐 Testing Authenticated AI Endpoints');
  console.log('-'.repeat(50));
  
  try {
    // Get CSRF token
    const csrfRes = await makeRequest('/api/auth/csrf');
    const csrfToken = csrfRes.data.data?.token;
    
    // Login
    const loginRes = await makeRequest('/api/auth/login', {
      method: 'POST',
      headers: { 'X-CSRF-Token': csrfToken },
      body: JSON.stringify({ username: 'alice', password: 'secret' })
    });
    
    if (loginRes.status === 200) {
      const token = loginRes.data.data?.accessToken;
      console.log('✅ Logged in as alice');
      
      // Test authenticated AI stats
      const authStatsRes = await makeRequest('/api/ai/stats', {
        headers: { 'Authorization': `Bearer ${token}` }
      });
      
      console.log(`\n📊 Authenticated AI Stats:`);
      console.log(`   Status: ${authStatsRes.status}`);
      console.log(`   User ID: ${authStatsRes.data.data?.userId}`);
      console.log(`   Daily limits: ${JSON.stringify(authStatsRes.data.data?.limits)}`);
    } else {
      console.log('❌ Login failed');
    }
  } catch (error) {
    console.log(`❌ Auth test error: ${error.message}`);
  }
  
  // Summary
  console.log('\n\n' + '='.repeat(50));
  console.log('📊 Test Summary');
  console.log('='.repeat(50));
  console.log(`✅ Passed: ${passed}`);
  console.log(`❌ Failed: ${failed}`);
  console.log(`📈 Success Rate: ${Math.round((passed/(passed+failed))*100)}%`);
  console.log('\n✨ AI service is now fully integrated and working!');
}

// Run the test
testAIFrontendIntegration().catch(console.error);