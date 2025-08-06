const http = require('http');

/**
 * End-to-End Consistency Test
 * Tests the complete flow with SOTA fixes
 */
class E2EConsistencyTest {
  constructor() {
    this.baseUrl = 'http://localhost:8080';
    this.token = null;
    this.csrfToken = null;
  }

  async request(path, options = {}) {
    const url = new URL(path, this.baseUrl);
    
    return new Promise((resolve, reject) => {
      const req = http.request(url, {
        ...options,
        headers: {
          'Content-Type': 'application/json',
          'Accept': 'application/json',
          ...(this.token && { 'Authorization': `Bearer ${this.token}` }),
          ...(this.csrfToken && options.method !== 'GET' && { 'X-CSRF-Token': this.csrfToken }),
          ...options.headers
        }
      }, (res) => {
        let data = '';
        res.on('data', chunk => data += chunk);
        res.on('end', () => {
          try {
            const json = JSON.parse(data);
            resolve({ 
              status: res.statusCode, 
              data: json, 
              headers: res.headers,
              cookies: res.headers['set-cookie']
            });
          } catch (e) {
            resolve({ status: res.statusCode, data: data, headers: res.headers });
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

  async runTests() {
    console.log('🚀 Starting End-to-End Consistency Test\n');
    console.log('This test verifies:');
    console.log('1. API route aliases work correctly');
    console.log('2. Response transformation (snake_case → camelCase)');
    console.log('3. Authentication flow consistency');
    console.log('4. Model field consistency\n');

    const results = {
      passed: 0,
      failed: 0,
      issues: []
    };

    try {
      // Test 1: CSRF Token (Frontend route)
      console.log('📍 Test 1: CSRF Token Endpoint');
      const csrfResponse = await this.request('/api/auth/csrf');
      console.log(`   Status: ${csrfResponse.status}`);
      
      if (csrfResponse.status === 200 && csrfResponse.data.data?.csrfToken) {
        this.csrfToken = csrfResponse.data.data.csrfToken;
        console.log('   ✅ CSRF token obtained');
        results.passed++;
      } else {
        console.log('   ❌ Failed to get CSRF token');
        results.failed++;
        results.issues.push('CSRF token endpoint failed');
      }

      // Test 2: Login with transformation check
      console.log('\n📍 Test 2: Login with Field Transformation');
      const loginResponse = await this.request('/api/v1/auth/login', {
        method: 'POST',
        body: JSON.stringify({ username: 'alice', password: 'secret' })
      });

      if (loginResponse.status === 200) {
        const userData = loginResponse.data.data?.user;
        const responseStr = JSON.stringify(loginResponse.data);
        
        console.log('   ✅ Login successful');
        
        // Check field transformation
        const transformationChecks = [
          { field: 'createdAt', oldField: 'created_at' },
          { field: 'updatedAt', oldField: 'updated_at' },
          { field: 'isActive', oldField: 'is_active' },
          { field: 'schoolCode', oldField: 'school_code' },
          { field: 'lastLoginAt', oldField: 'last_login_at' }
        ];

        let transformationSuccess = true;
        for (const check of transformationChecks) {
          if (userData && check.field in userData && !responseStr.includes(check.oldField)) {
            console.log(`   ✅ ${check.oldField} → ${check.field}`);
          } else if (responseStr.includes(check.oldField)) {
            console.log(`   ❌ ${check.oldField} not transformed`);
            transformationSuccess = false;
          }
        }

        if (transformationSuccess) {
          console.log('   ✅ All fields correctly transformed to camelCase');
          results.passed++;
        } else {
          console.log('   ❌ Some fields not transformed');
          results.failed++;
          results.issues.push('Field transformation incomplete');
        }

        // Save token for subsequent requests
        if (loginResponse.data.data?.token) {
          this.token = loginResponse.data.data.token;
          console.log('   ✅ Token saved for authenticated requests');
          results.passed++;
        }
      } else {
        console.log('   ❌ Login failed');
        results.failed++;
        results.issues.push('Login endpoint failed');
      }

      // Test 3: Frontend expected routes (with aliases)
      console.log('\n📍 Test 3: Frontend Route Aliases');
      
      const aliasTests = [
        { name: 'Schools List', path: '/api/schools' },
        { name: 'Postcode Lookup', path: '/api/postcode/100080' },
        { name: 'Address Search', path: '/api/address/search?q=test' }
      ];

      for (const test of aliasTests) {
        const response = await this.request(test.path);
        if (response.status === 200) {
          console.log(`   ✅ ${test.name}: Working`);
          results.passed++;
        } else {
          console.log(`   ❌ ${test.name}: Failed (${response.status})`);
          results.failed++;
          results.issues.push(`${test.name} failed`);
        }
      }

      // Test 4: Authenticated endpoint with transformation
      console.log('\n📍 Test 4: Authenticated Endpoint');
      const meResponse = await this.request('/api/v1/users/me');
      
      if (meResponse.status === 200) {
        const userData = meResponse.data.data;
        if (userData && 'createdAt' in userData && !JSON.stringify(userData).includes('created_at')) {
          console.log('   ✅ User profile correctly transformed');
          results.passed++;
        } else {
          console.log('   ❌ User profile transformation issue');
          results.failed++;
          results.issues.push('User profile transformation failed');
        }
      } else {
        console.log('   ❌ Failed to get user profile');
        results.failed++;
        results.issues.push('User profile endpoint failed');
      }

      // Test 5: AI endpoint (public)
      console.log('\n📍 Test 5: AI Endpoint');
      const aiResponse = await this.request('/api/v1/ai/inspiration', {
        method: 'POST',
        body: JSON.stringify({ theme: '日常生活', count: 1 })
      });

      if (aiResponse.status === 200 && aiResponse.data.data?.inspirations) {
        console.log('   ✅ AI inspiration endpoint working');
        console.log(`   ✅ Generated ${aiResponse.data.data.inspirations.length} inspiration(s)`);
        results.passed++;
      } else {
        console.log('   ❌ AI endpoint failed');
        results.failed++;
        results.issues.push('AI endpoint failed');
      }

    } catch (error) {
      console.error('\n❌ Test error:', error.message);
      results.failed++;
      results.issues.push(`Test error: ${error.message}`);
    }

    // Summary
    console.log('\n' + '='.repeat(60));
    console.log('📊 End-to-End Test Summary');
    console.log('='.repeat(60));
    console.log(`✅ Passed: ${results.passed}`);
    console.log(`❌ Failed: ${results.failed}`);
    
    if (results.issues.length > 0) {
      console.log('\n🔍 Issues found:');
      results.issues.forEach(issue => console.log(`   - ${issue}`));
    }

    const successRate = (results.passed / (results.passed + results.failed) * 100).toFixed(1);
    console.log(`\n🎯 Success Rate: ${successRate}%`);

    if (successRate === '100.0') {
      console.log('\n✅ All consistency issues have been resolved!');
      console.log('The frontend and backend are now fully synchronized.');
    } else {
      console.log('\n⚠️  Some consistency issues remain.');
      console.log('Please check the issues above and fix them.');
    }

    // Save report
    const report = {
      timestamp: new Date().toISOString(),
      results,
      successRate: parseFloat(successRate)
    };

    require('fs').writeFileSync('e2e-consistency-report.json', JSON.stringify(report, null, 2));
    console.log('\n📄 Report saved to e2e-consistency-report.json');
  }
}

// Run the test
const test = new E2EConsistencyTest();
test.runTests().catch(console.error);