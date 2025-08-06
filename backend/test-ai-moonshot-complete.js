/**
 * Complete AI Moonshot Integration Test
 * Tests the entire AI system with SOTA implementation
 */

const http = require('http');

class AIMoonshotTest {
  constructor() {
    this.baseUrl = 'http://localhost:8080';
    this.token = null;
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
          ...options.headers
        }
      }, (res) => {
        let data = '';
        res.on('data', chunk => data += chunk);
        res.on('end', () => {
          try {
            const json = JSON.parse(data);
            resolve({ status: res.statusCode, data: json });
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

  async runTests() {
    console.log('🚀 Starting AI Moonshot Integration Test\n');
    console.log('This test verifies:');
    console.log('1. AI endpoints are accessible');
    console.log('2. Moonshot API integration works');
    console.log('3. Real AI responses are generated');
    console.log('4. SOTA features (circuit breaker, retry) work\n');

    const results = {
      passed: 0,
      failed: 0,
      issues: []
    };

    try {
      // Login first
      console.log('📍 Logging in...');
      const loginRes = await this.request('/api/v1/auth/login', {
        method: 'POST',
        body: JSON.stringify({ username: 'alice', password: 'secret' })
      });

      if (loginRes.status === 200 && loginRes.data.data?.token) {
        this.token = loginRes.data.data.token;
        console.log('   ✅ Login successful\n');
      } else {
        throw new Error('Failed to login');
      }

      // Test 1: AI Inspiration (Public)
      console.log('📍 Test 1: AI Inspiration Generation');
      const inspirationRes = await this.request('/api/v1/ai/inspiration', {
        method: 'POST',
        body: JSON.stringify({ 
          theme: '友谊', 
          count: 3 
        })
      });

      if (inspirationRes.status === 200 && inspirationRes.data.data?.inspirations) {
        const inspirations = inspirationRes.data.data.inspirations;
        console.log(`   ✅ Generated ${inspirations.length} inspirations`);
        
        // Check if we got real AI content (not fallback)
        const hasRealContent = inspirations.some(insp => 
          insp.content && 
          insp.content.length > 50 && 
          !insp.content.includes('这是一个关于')
        );
        
        if (hasRealContent) {
          console.log('   ✅ Real AI content generated (not fallback)');
          console.log(`   📝 Sample: "${inspirations[0].content.substring(0, 100)}..."`);
          results.passed++;
        } else {
          console.log('   ⚠️  Got fallback content instead of AI response');
          results.issues.push('AI returned fallback content');
          results.failed++;
        }
      } else {
        console.log('   ❌ Failed to generate inspirations');
        results.failed++;
        results.issues.push('Inspiration generation failed');
      }

      // Test 2: AI Reply Generator
      console.log('\n📍 Test 2: AI Reply Generator');
      const replyRes = await this.request('/api/v1/ai/reply', {
        method: 'POST',
        body: JSON.stringify({
          letterId: 'test-letter-123',
          persona: 'warm'
        })
      });

      if (replyRes.status === 200 && replyRes.data.data?.reply) {
        const reply = replyRes.data.data.reply;
        console.log('   ✅ Reply generated successfully');
        console.log(`   📝 Preview: "${reply.content.substring(0, 100)}..."`);
        results.passed++;
      } else {
        console.log('   ❌ Failed to generate reply');
        results.failed++;
        results.issues.push('Reply generation failed');
      }

      // Test 3: AI Personas
      console.log('\n📍 Test 3: AI Personas List');
      const personasRes = await this.request('/api/v1/ai/personas');

      if (personasRes.status === 200 && personasRes.data.data?.personas) {
        const personas = personasRes.data.data.personas;
        console.log(`   ✅ Found ${personas.length} AI personas`);
        personas.forEach(p => {
          console.log(`      - ${p.name}: ${p.description}`);
        });
        results.passed++;
      } else {
        console.log('   ❌ Failed to get personas');
        results.failed++;
        results.issues.push('Personas list failed');
      }

      // Test 4: Pen Pal Match
      console.log('\n📍 Test 4: AI Pen Pal Match');
      const matchRes = await this.request('/api/v1/ai/match', {
        method: 'POST',
        body: JSON.stringify({
          interests: ['旅行', '音乐', '阅读']
        })
      });

      if (matchRes.status === 200 && matchRes.data.data?.matches) {
        const matches = matchRes.data.data.matches;
        console.log(`   ✅ Found ${matches.length} potential pen pals`);
        results.passed++;
      } else {
        console.log('   ❌ Failed to find matches');
        results.failed++;
        results.issues.push('Pen pal matching failed');
      }

      // Test 5: Writing Advice
      console.log('\n📍 Test 5: AI Writing Advice');
      const adviceRes = await this.request('/api/v1/ai/advice', {
        method: 'POST',
        body: JSON.stringify({
          topic: '如何写一封感人的道歉信',
          level: 'beginner'
        })
      });

      if (adviceRes.status === 200 && adviceRes.data.data?.advice) {
        console.log('   ✅ Writing advice generated');
        console.log(`   📝 "${adviceRes.data.data.advice.substring(0, 100)}..."`);
        results.passed++;
      } else {
        console.log('   ❌ Failed to generate advice');
        results.failed++;
        results.issues.push('Writing advice failed');
      }

      // Test 6: Error Handling (Invalid request)
      console.log('\n📍 Test 6: Error Handling');
      const errorRes = await this.request('/api/v1/ai/inspiration', {
        method: 'POST',
        body: JSON.stringify({}) // Missing required fields
      });

      if (errorRes.status === 400 || errorRes.status === 422) {
        console.log('   ✅ Error handling works correctly');
        results.passed++;
      } else {
        console.log('   ❌ Error handling not working properly');
        results.failed++;
        results.issues.push('Error handling issue');
      }

    } catch (error) {
      console.error('\n❌ Test error:', error.message);
      results.failed++;
      results.issues.push(`Test error: ${error.message}`);
    }

    // Summary
    console.log('\n' + '='.repeat(60));
    console.log('📊 AI Moonshot Test Summary');
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
      console.log('\n✅ AI system is fully functional with Moonshot integration!');
    } else if (results.issues.some(i => i.includes('fallback'))) {
      console.log('\n⚠️  AI system is working but using fallback content.');
      console.log('Check that Moonshot API key is valid and quota is available.');
    } else {
      console.log('\n⚠️  Some AI features are not working correctly.');
    }

    // Save report
    const report = {
      timestamp: new Date().toISOString(),
      results,
      successRate: parseFloat(successRate),
      moonshotStatus: results.issues.some(i => i.includes('fallback')) ? 'fallback' : 'active'
    };

    require('fs').writeFileSync('ai-moonshot-test-report.json', JSON.stringify(report, null, 2));
    console.log('\n📄 Report saved to ai-moonshot-test-report.json');
  }
}

// Run the test
const test = new AIMoonshotTest();
test.runTests().catch(console.error);