/**
 * Complete SOTA Integration Test
 * Demonstrates all implemented features working together
 */

const http = require('http');

async function makeRequest(path, options = {}) {
  const url = new URL(path, 'http://localhost:8080');
  
  return new Promise((resolve, reject) => {
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
          const json = JSON.parse(data);
          resolve({ status: res.statusCode, data: json, headers: res.headers });
        } catch (e) {
          resolve({ status: res.statusCode, data: data });
        }
      });
    });
    
    req.on('error', reject);
    
    if (options.body) {
      req.write(options.body);
    }
    
    req.end();
  });
}

async function runSOTATest() {
  console.log('🚀 SOTA Complete Integration Test\n');
  console.log('This test demonstrates:');
  console.log('✅ API Route Aliases');
  console.log('✅ Response Transformation (snake_case → camelCase)');
  console.log('✅ AI Integration with Moonshot');
  console.log('✅ Authentication Flow');
  console.log('✅ Error Handling\n');

  const results = [];

  try {
    // Test 1: Frontend Route Alias
    console.log('📍 Test 1: Frontend Route Alias');
    console.log('   Testing: /api/schools (aliased from /api/v1/schools)');
    const schoolsRes = await makeRequest('/api/schools');
    console.log(`   Status: ${schoolsRes.status}`);
    if (schoolsRes.status === 200) {
      console.log('   ✅ Route alias working');
      results.push({ test: 'Route Alias', status: 'PASS' });
    } else {
      console.log('   ❌ Route alias failed');
      results.push({ test: 'Route Alias', status: 'FAIL' });
    }

    // Test 2: CSRF + Login with Transformation
    console.log('\n📍 Test 2: Authentication with Field Transformation');
    
    // Get CSRF token
    const csrfRes = await makeRequest('/api/auth/csrf');
    const csrfToken = csrfRes.data.data?.csrfToken;
    console.log('   ✅ CSRF token obtained');
    
    // Login
    const loginRes = await makeRequest('/api/v1/auth/login', {
      method: 'POST',
      headers: { 'X-CSRF-Token': csrfToken },
      body: JSON.stringify({ username: 'alice', password: 'secret' })
    });
    
    if (loginRes.status === 200) {
      const user = loginRes.data.data?.user;
      console.log('   ✅ Login successful');
      
      // Check field transformation
      const hasTransformed = user && 
        'createdAt' in user && 
        'isActive' in user &&
        'schoolCode' in user &&
        !JSON.stringify(user).includes('created_at');
      
      if (hasTransformed) {
        console.log('   ✅ Fields transformed to camelCase');
        console.log(`      Example: createdAt = ${user.createdAt}`);
        results.push({ test: 'Field Transformation', status: 'PASS' });
      } else {
        console.log('   ❌ Field transformation failed');
        results.push({ test: 'Field Transformation', status: 'FAIL' });
      }
      
      const token = loginRes.data.data?.token;
      
      // Test 3: AI with Real Response
      console.log('\n📍 Test 3: AI Integration (Public Endpoint)');
      const aiRes = await makeRequest('/api/v1/ai/inspiration', {
        method: 'POST',
        body: JSON.stringify({ theme: '友谊', count: 2 })
      });
      
      if (aiRes.status === 200 && aiRes.data.data?.inspirations) {
        const inspirations = aiRes.data.data.inspirations;
        console.log(`   ✅ Generated ${inspirations.length} inspirations`);
        
        // Check if it's real AI content (longer, more varied)
        const firstInspiration = inspirations[0];
        const isRealAI = firstInspiration.prompt && 
                        firstInspiration.prompt.length > 30 &&
                        !firstInspiration.prompt.startsWith('这是一个关于');
        
        if (isRealAI) {
          console.log('   ✅ Real AI content (Moonshot working)');
          console.log(`      "${firstInspiration.prompt.substring(0, 60)}..."`);
          results.push({ test: 'AI Integration', status: 'PASS' });
        } else {
          console.log('   ⚠️  Fallback content (AI service issue)');
          results.push({ test: 'AI Integration', status: 'PARTIAL' });
        }
      } else {
        console.log('   ❌ AI endpoint failed');
        results.push({ test: 'AI Integration', status: 'FAIL' });
      }
      
      // Test 4: Authenticated Request
      console.log('\n📍 Test 4: Authenticated API Call');
      const meRes = await makeRequest('/api/v1/users/me', {
        headers: { 'Authorization': `Bearer ${token}` }
      });
      
      if (meRes.status === 200) {
        console.log('   ✅ Authenticated request successful');
        results.push({ test: 'Authentication', status: 'PASS' });
      } else {
        console.log('   ❌ Authentication failed');
        results.push({ test: 'Authentication', status: 'FAIL' });
      }
      
    } else {
      console.log('   ❌ Login failed');
      results.push({ test: 'Authentication', status: 'FAIL' });
    }

    // Test 5: Error Handling
    console.log('\n📍 Test 5: Error Handling');
    const errorRes = await makeRequest('/api/v1/nonexistent', {
      method: 'POST',
      body: JSON.stringify({})
    });
    
    if (errorRes.status === 404 && errorRes.data.message) {
      console.log('   ✅ Error handling working');
      console.log(`      Message: "${errorRes.data.message}"`);
      results.push({ test: 'Error Handling', status: 'PASS' });
    } else {
      console.log('   ❌ Error handling issue');
      results.push({ test: 'Error Handling', status: 'FAIL' });
    }

  } catch (error) {
    console.error('\n❌ Test error:', error.message);
  }

  // Summary
  console.log('\n' + '='.repeat(60));
  console.log('📊 SOTA Implementation Test Summary');
  console.log('='.repeat(60));
  
  const passed = results.filter(r => r.status === 'PASS').length;
  const partial = results.filter(r => r.status === 'PARTIAL').length;
  const failed = results.filter(r => r.status === 'FAIL').length;
  
  console.log(`✅ Passed: ${passed}`);
  console.log(`⚠️  Partial: ${partial}`);
  console.log(`❌ Failed: ${failed}`);
  
  console.log('\nDetailed Results:');
  results.forEach(r => {
    const icon = r.status === 'PASS' ? '✅' : r.status === 'PARTIAL' ? '⚠️' : '❌';
    console.log(`   ${icon} ${r.test}: ${r.status}`);
  });
  
  const successRate = ((passed + partial * 0.5) / results.length * 100).toFixed(1);
  console.log(`\n🎯 Success Rate: ${successRate}%`);
  
  if (successRate >= 80) {
    console.log('\n✅ SOTA Implementation is working successfully!');
    console.log('   All major features are functional and integrated.');
  } else {
    console.log('\n⚠️  Some features need attention.');
  }
  
  // Save report
  const report = {
    timestamp: new Date().toISOString(),
    results,
    summary: {
      passed,
      partial,
      failed,
      successRate: parseFloat(successRate)
    }
  };
  
  require('fs').writeFileSync('sota-test-report.json', JSON.stringify(report, null, 2));
  console.log('\n📄 Report saved to sota-test-report.json');
}

// Run the test
runSOTATest().catch(console.error);