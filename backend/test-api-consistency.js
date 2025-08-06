const http = require('http');
const https = require('https');

class APIConsistencyTester {
  constructor(baseUrl = 'http://localhost:8080') {
    this.baseUrl = baseUrl;
    this.results = {
      passed: [],
      failed: [],
      warnings: []
    };
  }

  async request(path, options = {}) {
    const url = new URL(path, this.baseUrl);
    const isHttps = url.protocol === 'https:';
    const lib = isHttps ? https : http;
    
    return new Promise((resolve, reject) => {
      const req = lib.request(url, {
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

  async testEndpoint(name, path, options = {}) {
    console.log(`\n🧪 Testing: ${name}`);
    console.log(`   Path: ${path}`);
    
    try {
      const response = await this.request(path, options);
      console.log(`   Status: ${response.status}`);
      
      if (response.status >= 200 && response.status < 300) {
        console.log(`   ✅ Success`);
        
        // Check for field transformation
        if (response.data && typeof response.data === 'object') {
          const hasSnakeCase = JSON.stringify(response.data).includes('_');
          const hasCamelCase = /[a-z][A-Z]/.test(JSON.stringify(response.data));
          
          if (hasSnakeCase && !hasCamelCase) {
            console.log(`   ⚠️  Warning: Response contains snake_case fields`);
            this.results.warnings.push({ name, issue: 'snake_case in response' });
          }
        }
        
        this.results.passed.push(name);
        return response;
      } else {
        console.log(`   ❌ Failed with status ${response.status}`);
        this.results.failed.push({ name, status: response.status });
        return response;
      }
    } catch (error) {
      console.log(`   ❌ Error: ${error.message}`);
      this.results.failed.push({ name, error: error.message });
      return null;
    }
  }

  async runTests() {
    console.log('🚀 Starting API Consistency Tests\n');
    
    // Wait for backend to be ready
    await new Promise(resolve => setTimeout(resolve, 3000));
    
    // Test health endpoint
    await this.testEndpoint('Health Check', '/health');
    
    // Test API aliases (frontend expected routes)
    console.log('\n📍 Testing API Aliases (Frontend Expected Routes)');
    
    // Authentication endpoints
    await this.testEndpoint('CSRF Token', '/api/auth/csrf');
    
    await this.testEndpoint('Check Username', '/api/auth/check-username', {
      method: 'POST',
      body: JSON.stringify({ username: 'testuser' })
    });
    
    await this.testEndpoint('Check Email', '/api/auth/check-email', {
      method: 'POST',
      body: JSON.stringify({ email: 'test@example.com' })
    });
    
    // School endpoints
    await this.testEndpoint('List Schools', '/api/schools');
    await this.testEndpoint('Search Schools', '/api/schools?search=北京');
    
    await this.testEndpoint('Validate School', '/api/schools/validate', {
      method: 'POST',
      body: JSON.stringify({ code: 'PKU001' })
    });
    
    // Postcode endpoints
    await this.testEndpoint('Get Postcode', '/api/postcode/100080');
    await this.testEndpoint('Address Search', '/api/address/search?q=北京大学');
    
    // Admin permission endpoints
    await this.testEndpoint('Permissions Overview', '/api/admin/permissions?type=overview');
    await this.testEndpoint('Permissions Roles', '/api/admin/permissions?type=roles');
    await this.testEndpoint('Permissions Audit', '/api/admin/permissions/audit');
    
    // Error reporting
    await this.testEndpoint('Report Error', '/api/errors/report', {
      method: 'POST',
      body: JSON.stringify({
        error: 'Test error',
        stack: 'Test stack trace',
        context: { test: true }
      })
    });
    
    // Test actual backend routes
    console.log('\n📍 Testing Backend Routes (Direct API)');
    
    // Test login with field transformation
    const loginResponse = await this.testEndpoint('Login (Backend)', '/api/v1/auth/login', {
      method: 'POST',
      body: JSON.stringify({ username: 'alice', password: 'secret' })
    });
    
    if (loginResponse && loginResponse.data.success) {
      console.log('   🔍 Checking field transformation:');
      const responseStr = JSON.stringify(loginResponse.data);
      
      // Check if response has been transformed to camelCase
      if (responseStr.includes('created_at')) {
        console.log('   ⚠️  Response still contains snake_case (created_at)');
      } else if (responseStr.includes('createdAt')) {
        console.log('   ✅ Response correctly transformed to camelCase (createdAt)');
      }
    }
    
    // Summary
    console.log('\n📊 Test Summary');
    console.log('================');
    console.log(`✅ Passed: ${this.results.passed.length}`);
    console.log(`❌ Failed: ${this.results.failed.length}`);
    console.log(`⚠️  Warnings: ${this.results.warnings.length}`);
    
    if (this.results.failed.length > 0) {
      console.log('\nFailed tests:');
      this.results.failed.forEach(f => {
        console.log(`  - ${f.name}: ${f.status || f.error}`);
      });
    }
    
    if (this.results.warnings.length > 0) {
      console.log('\nWarnings:');
      this.results.warnings.forEach(w => {
        console.log(`  - ${w.name}: ${w.issue}`);
      });
    }
    
    // Save detailed report
    const report = {
      timestamp: new Date().toISOString(),
      summary: {
        total: this.results.passed.length + this.results.failed.length,
        passed: this.results.passed.length,
        failed: this.results.failed.length,
        warnings: this.results.warnings.length
      },
      results: this.results
    };
    
    require('fs').writeFileSync('api-consistency-test-report.json', JSON.stringify(report, null, 2));
    console.log('\n✅ Detailed report saved to api-consistency-test-report.json');
  }
}

// Run tests
const tester = new APIConsistencyTester();
tester.runTests().catch(console.error);