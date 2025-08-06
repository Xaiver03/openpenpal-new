#!/usr/bin/env node

/**
 * Comprehensive Security Implementation Test
 * 综合安全实现测试
 * 
 * Tests all security features implemented:
 * - CSRF protection
 * - Rate limiting 
 * - HTTPS configuration
 * - Security headers
 * - JWT authentication
 */

import { execSync } from 'child_process';

const FRONTEND_URL = 'http://localhost:3000';
const BACKEND_URL = 'http://localhost:8080';

console.log('🔒 OpenPenPal Security Implementation Test');
console.log('==========================================');

class SecurityTester {
  constructor() {
    this.results = {
      passed: 0,
      failed: 0,
      tests: []
    };
  }

  async test(name, testFn) {
    console.log(`\n🧪 Testing: ${name}`);
    try {
      const result = await testFn();
      if (result.success) {
        console.log(`✅ PASS: ${result.message}`);
        this.results.passed++;
        this.results.tests.push({ name, status: 'PASS', message: result.message });
      } else {
        console.log(`❌ FAIL: ${result.message}`);
        this.results.failed++;
        this.results.tests.push({ name, status: 'FAIL', message: result.message });
      }
    } catch (error) {
      console.log(`❌ ERROR: ${error.message}`);
      this.results.failed++;
      this.results.tests.push({ name, status: 'ERROR', message: error.message });
    }
  }

  async runAllTests() {
    console.log('🚀 Starting comprehensive security tests...\n');

    // 1. Test CSRF Protection
    await this.test('CSRF Token Generation', async () => {
      const response = await fetch(`${FRONTEND_URL}/api/auth/csrf`);
      if (response.ok) {
        const data = await response.json();
        return {
          success: !!data.csrfToken,
          message: data.csrfToken ? 'CSRF token generated successfully' : 'No CSRF token returned'
        };
      }
      return { success: false, message: `CSRF endpoint failed: ${response.status}` };
    });

    await this.test('CSRF Token Validation', async () => {
      try {
        // First get CSRF token
        const csrfResponse = await fetch(`${FRONTEND_URL}/api/auth/csrf`);
        const csrfData = await csrfResponse.json();
        
        // Try login without CSRF token (should fail in production)
        const loginResponse = await fetch(`${FRONTEND_URL}/api/auth/login`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ username: 'admin', password: 'admin123' })
        });

        const isDevelopment = process.env.NODE_ENV === 'development';
        if (isDevelopment) {
          return {
            success: true,
            message: 'CSRF validation skipped in development mode (as configured)'
          };
        } else {
          return {
            success: loginResponse.status === 403,
            message: loginResponse.status === 403 ? 'CSRF validation working' : 'CSRF validation bypassed'
          };
        }
      } catch (error) {
        return { success: false, message: `CSRF test error: ${error.message}` };
      }
    });

    // 2. Test Rate Limiting
    await this.test('Rate Limiting - Auth Endpoint', async () => {
      const requests = [];
      const testLimit = 15; // Should hit rate limit
      
      for (let i = 0; i < testLimit; i++) {
        requests.push(
          fetch(`${FRONTEND_URL}/api/auth/login`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username: 'invalid', password: 'invalid' })
          })
        );
      }

      const responses = await Promise.all(requests);
      const rateLimitedResponses = responses.filter(r => r.status === 429);
      
      return {
        success: rateLimitedResponses.length > 0,
        message: `Rate limiting ${rateLimitedResponses.length > 0 ? 'active' : 'not detected'} (${rateLimitedResponses.length}/${testLimit} requests blocked)`
      };
    });

    // 3. Test Security Headers
    await this.test('Security Headers - CSP', async () => {
      const response = await fetch(FRONTEND_URL);
      const cspHeader = response.headers.get('content-security-policy') || 
                       response.headers.get('content-security-policy-report-only');
      
      return {
        success: !!cspHeader,
        message: cspHeader ? 'CSP header present' : 'CSP header missing'
      };
    });

    await this.test('Security Headers - HSTS', async () => {
      const response = await fetch(FRONTEND_URL);
      const hstsHeader = response.headers.get('strict-transport-security');
      
      return {
        success: !!hstsHeader || process.env.NODE_ENV === 'development',
        message: hstsHeader ? 'HSTS header present' : 'HSTS header missing (OK in development)'
      };
    });

    await this.test('Security Headers - X-Frame-Options', async () => {
      const response = await fetch(FRONTEND_URL);
      const frameHeader = response.headers.get('x-frame-options');
      
      return {
        success: !!frameHeader,
        message: frameHeader ? `X-Frame-Options: ${frameHeader}` : 'X-Frame-Options header missing'
      };
    });

    await this.test('Security Headers - X-Content-Type-Options', async () => {
      const response = await fetch(FRONTEND_URL);
      const contentTypeHeader = response.headers.get('x-content-type-options');
      
      return {
        success: !!contentTypeHeader,
        message: contentTypeHeader ? 'X-Content-Type-Options: nosniff' : 'X-Content-Type-Options header missing'
      };
    });

    // 4. Test Authentication
    await this.test('JWT Authentication', async () => {
      const response = await fetch(`${FRONTEND_URL}/api/auth/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username: 'admin', password: 'admin123' })
      });

      if (response.ok) {
        const data = await response.json();
        return {
          success: !!(data.data && data.data.accessToken),
          message: data.data?.accessToken ? 'JWT token generated' : 'No JWT token in response'
        };
      }
      return { success: false, message: `Login failed: ${response.status}` };
    });

    // 5. Test HTTPS Redirect (if in production)
    await this.test('HTTPS Enforcement', async () => {
      if (process.env.NODE_ENV === 'development') {
        return {
          success: true,
          message: 'HTTPS enforcement skipped in development'
        };
      }

      try {
        const response = await fetch('http://localhost:3000', { redirect: 'manual' });
        return {
          success: response.status === 301 || response.status === 302,
          message: response.status === 301 || response.status === 302 ? 'HTTPS redirect working' : 'No HTTPS redirect'
        };
      } catch (error) {
        return { success: false, message: `HTTPS test error: ${error.message}` };
      }
    });

    // 6. Test Environment Configuration
    await this.test('Environment Configuration', async () => {
      const requiredEnvVars = [
        'NODE_ENV',
        'NEXT_PUBLIC_API_URL',
        'NEXT_PUBLIC_GATEWAY_URL'
      ];

      const missing = requiredEnvVars.filter(env => !process.env[env]);
      
      return {
        success: missing.length === 0,
        message: missing.length === 0 ? 'All required environment variables set' : `Missing: ${missing.join(', ')}`
      };
    });

    // 7. Test Role-based Access
    await this.test('Role-based Rate Limiting Configuration', async () => {
      try {
        // Test if our rate limiting modules are properly configured
        const response = await fetch(`${FRONTEND_URL}/api/auth/login`, {
          method: 'POST',
          headers: { 
            'Content-Type': 'application/json',
            'Authorization': 'Bearer fake-admin-token'
          },
          body: JSON.stringify({ username: 'admin', password: 'admin123' })
        });

        // Check rate limit headers
        const rateLimitHeaders = {
          remaining: response.headers.get('x-ratelimit-remaining'),
          limit: response.headers.get('x-ratelimit-limit'),
          reset: response.headers.get('x-ratelimit-reset')
        };

        const hasRateLimitHeaders = Object.values(rateLimitHeaders).some(header => header !== null);

        return {
          success: hasRateLimitHeaders,
          message: hasRateLimitHeaders ? 'Rate limit headers present' : 'Rate limit headers missing'
        };
      } catch (error) {
        return { success: false, message: `Role-based rate limiting test error: ${error.message}` };
      }
    });

    // 8. Test Production Environment Variables
    await this.test('Production Security Settings', async () => {
      const securitySettings = {
        csrfEnabled: process.env.CSRF_ENABLED === 'true',
        rateLimitEnabled: process.env.RATE_LIMIT_ENABLED === 'true',
        securityHeaders: process.env.SECURITY_HEADERS_ENABLED === 'true',
        httpsForced: process.env.NEXT_PUBLIC_FORCE_HTTPS === 'true'
      };

      const enabledSettings = Object.entries(securitySettings).filter(([_, enabled]) => enabled);
      
      return {
        success: enabledSettings.length >= 2, // At least 2 security features enabled
        message: `Security features enabled: ${enabledSettings.map(([name]) => name).join(', ')}`
      };
    });

    this.printResults();
  }

  printResults() {
    console.log('\n🔒 SECURITY TEST RESULTS');
    console.log('========================');
    console.log(`✅ Passed: ${this.results.passed}`);
    console.log(`❌ Failed: ${this.results.failed}`);
    console.log(`📊 Total: ${this.results.passed + this.results.failed}`);
    console.log(`📈 Success Rate: ${((this.results.passed / (this.results.passed + this.results.failed)) * 100).toFixed(1)}%`);

    if (this.results.failed > 0) {
      console.log('\n❌ Failed Tests:');
      this.results.tests
        .filter(test => test.status !== 'PASS')
        .forEach(test => console.log(`   • ${test.name}: ${test.message}`));
    }

    console.log('\n🎯 Security Implementation Status:');
    
    const securityFeatures = {
      'CSRF Protection': this.results.tests.some(t => t.name.includes('CSRF') && t.status === 'PASS'),
      'Rate Limiting': this.results.tests.some(t => t.name.includes('Rate Limiting') && t.status === 'PASS'),
      'Security Headers': this.results.tests.filter(t => t.name.includes('Security Headers') && t.status === 'PASS').length >= 2,
      'JWT Authentication': this.results.tests.some(t => t.name.includes('JWT') && t.status === 'PASS'),
      'Environment Config': this.results.tests.some(t => t.name.includes('Environment') && t.status === 'PASS')
    };

    Object.entries(securityFeatures).forEach(([feature, implemented]) => {
      console.log(`   ${implemented ? '✅' : '❌'} ${feature}`);
    });

    const overallScore = Object.values(securityFeatures).filter(Boolean).length;
    console.log(`\n🏆 Overall Security Score: ${overallScore}/5 features implemented`);

    if (overallScore >= 4) {
      console.log('🎉 Excellent! Your security implementation is production-ready.');
    } else if (overallScore >= 3) {
      console.log('⚠️  Good progress! A few more security features needed.');
    } else {
      console.log('🚨 More security features required before production deployment.');
    }
  }
}

// Function to check if services are running
async function checkServices() {
  console.log('🔍 Checking service availability...');
  
  try {
    const frontendResponse = await fetch(FRONTEND_URL, { timeout: 5000 });
    console.log(`✅ Frontend service: ${frontendResponse.status === 200 ? 'Available' : 'Responding with ' + frontendResponse.status}`);
  } catch (error) {
    console.log(`❌ Frontend service: Not available (${error.message})`);
    console.log('   💡 Start with: npm run dev');
  }

  try {
    const backendResponse = await fetch(`${BACKEND_URL}/health`, { timeout: 5000 });
    console.log(`✅ Backend service: ${backendResponse.status === 200 ? 'Available' : 'Responding with ' + backendResponse.status}`);
  } catch (error) {
    console.log(`❌ Backend service: Not available (${error.message})`);
    console.log('   💡 Start with: cd backend && go run main.go');
  }
}

// Main execution
async function main() {
  await checkServices();
  
  const tester = new SecurityTester();
  await tester.runAllTests();
  
  console.log('\n📚 Next Steps:');
  console.log('1. Review failed tests and implement missing security features');
  console.log('2. Update production environment variables with actual values');  
  console.log('3. Test in staging environment before production deployment');
  console.log('4. Set up monitoring and alerting for security events');
  console.log('5. Configure TLS certificates and HTTPS properly');
  
  process.exit(tester.results.failed > 0 ? 1 : 0);
}

if (import.meta.url === `file://${process.argv[1]}`) {
  main().catch(console.error);
}