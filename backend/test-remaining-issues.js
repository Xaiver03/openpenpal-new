/**
 * 测试剩余问题并验证修复
 * Test remaining issues and verify fixes
 */

const http = require('http');
const fs = require('fs');

class IssueVerificationTest {
  constructor() {
    this.baseUrl = 'http://localhost:8080';
    this.issues = {
      letterCreation: { status: 'pending', description: '信件创建失败' },
      courierFieldTransform: { status: 'pending', description: '信使字段转换不完整' },
      hierarchyAccess: { status: 'pending', description: '层级信息访问失败' },
      validationError: { status: 'pending', description: '验证错误处理' }
    };
  }

  async request(path, options = {}) {
    const url = new URL(path, this.baseUrl);
    
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
            resolve({ status: res.statusCode, data: json, raw: data });
          } catch (e) {
            resolve({ status: res.statusCode, data: data, raw: data });
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
    console.log('🔍 剩余问题验证测试\n');
    console.log('测试目标：');
    console.log('1. 信件创建功能');
    console.log('2. 信使字段转换完整性');
    console.log('3. 层级API访问');
    console.log('4. 验证错误处理\n');

    // Get auth token first
    const authToken = await this.authenticate();
    if (!authToken) {
      console.error('❌ 认证失败，无法继续测试');
      return;
    }

    // Test each issue
    await this.testLetterCreation(authToken);
    await this.testCourierFieldTransform();
    await this.testHierarchyAccess();
    await this.testValidationErrorHandling(authToken);

    // Generate report
    this.generateReport();
  }

  async authenticate() {
    console.log('🔐 获取认证令牌...');
    
    // Get CSRF
    const csrfRes = await this.request('/api/auth/csrf');
    const csrfToken = csrfRes.data.data?.csrfToken;
    
    // Login
    const loginRes = await this.request('/api/v1/auth/login', {
      method: 'POST',
      headers: { 'X-CSRF-Token': csrfToken },
      body: JSON.stringify({ username: 'alice', password: 'secret' })
    });
    
    if (loginRes.status === 200 && loginRes.data.data?.token) {
      console.log('✅ 认证成功\n');
      return loginRes.data.data.token;
    }
    
    return null;
  }

  async testLetterCreation(token) {
    console.log('📝 测试1: 信件创建');
    
    // Test with minimal required fields
    const minimalLetter = {
      title: '测试信件',
      content: '这是一封测试信件的内容。'
    };
    
    console.log('   尝试1: 最小字段集');
    let res = await this.request('/api/v1/letters', {
      method: 'POST',
      headers: { 'Authorization': `Bearer ${token}` },
      body: JSON.stringify(minimalLetter)
    });
    
    if (res.status === 201) {
      console.log('   ✅ 最小字段创建成功');
      this.issues.letterCreation.status = 'fixed';
    } else {
      console.log(`   ❌ 失败 (${res.status}): ${res.data.message || res.data.error}`);
      
      // Try with all fields
      console.log('   尝试2: 完整字段集');
      const fullLetter = {
        title: '测试信件',
        content: '这是一封测试信件的内容。',
        style: 'warm',
        visibility: 'private',
        recipientOpCode: '',
        senderOpCode: ''
      };
      
      res = await this.request('/api/v1/letters', {
        method: 'POST',
        headers: { 'Authorization': `Bearer ${token}` },
        body: JSON.stringify(fullLetter)
      });
      
      if (res.status === 201) {
        console.log('   ✅ 完整字段创建成功');
        this.issues.letterCreation.status = 'fixed';
        this.issues.letterCreation.solution = '需要提供完整字段';
      } else {
        console.log(`   ❌ 仍然失败: ${res.data.message}`);
        this.issues.letterCreation.details = res.data;
      }
    }
    
    console.log('');
  }

  async testCourierFieldTransform() {
    console.log('🚴 测试2: 信使字段转换');
    
    // Login as courier
    const csrfRes = await this.request('/api/auth/csrf');
    const csrfToken = csrfRes.data.data?.csrfToken;
    
    const loginRes = await this.request('/api/v1/auth/login', {
      method: 'POST',
      headers: { 'X-CSRF-Token': csrfToken },
      body: JSON.stringify({ username: 'courier_level1', password: 'secret' })
    });
    
    if (loginRes.status !== 200) {
      console.log('   ❌ 信使登录失败');
      return;
    }
    
    const courierToken = loginRes.data.data?.token;
    
    // Get courier info
    const courierRes = await this.request('/api/v1/courier/me', {
      headers: { 'Authorization': `Bearer ${courierToken}` }
    });
    
    if (courierRes.status === 200) {
      const courier = courierRes.data.data;
      const rawData = courierRes.raw;
      
      // Check specific courier fields
      const courierFields = [
        { camel: 'managedOpCodePrefix', snake: 'managed_op_code_prefix' },
        { camel: 'hasPrinter', snake: 'has_printer' },
        { camel: 'weeklyHours', snake: 'weekly_hours' },
        { camel: 'maxDailyTasks', snake: 'max_daily_tasks' },
        { camel: 'transportMethod', snake: 'transport_method' }
      ];
      
      let allTransformed = true;
      for (const field of courierFields) {
        const hasCamel = courier && field.camel in courier;
        const hasSnake = rawData.includes(field.snake);
        
        if (hasCamel && !hasSnake) {
          console.log(`   ✅ ${field.snake} → ${field.camel}`);
        } else if (!hasCamel && courier) {
          console.log(`   ⚠️  ${field.camel} 字段缺失`);
          allTransformed = false;
        }
      }
      
      this.issues.courierFieldTransform.status = allTransformed ? 'fixed' : 'partial';
      this.issues.courierFieldTransform.details = {
        fieldsFound: Object.keys(courier || {}),
        expectedFields: courierFields.map(f => f.camel)
      };
    }
    
    console.log('');
  }

  async testHierarchyAccess() {
    console.log('🏗️ 测试3: 层级信息访问');
    
    // Test with different courier levels
    const courierLevels = [
      { username: 'courier_level1', level: 1 },
      { username: 'courier_level2', level: 2 },
      { username: 'courier_level3', level: 3 },
      { username: 'courier_level4', level: 4 }
    ];
    
    let anySuccess = false;
    
    for (const courierInfo of courierLevels) {
      console.log(`   测试 L${courierInfo.level} 信使...`);
      
      // Login
      const csrfRes = await this.request('/api/auth/csrf');
      const csrfToken = csrfRes.data.data?.csrfToken;
      
      const loginRes = await this.request('/api/v1/auth/login', {
        method: 'POST',
        headers: { 'X-CSRF-Token': csrfToken },
        body: JSON.stringify({ 
          username: courierInfo.username, 
          password: 'secret' 
        })
      });
      
      if (loginRes.status === 200) {
        const token = loginRes.data.data?.token;
        
        // Try hierarchy endpoint
        const hierarchyRes = await this.request('/api/v1/courier/hierarchy/me', {
          headers: { 'Authorization': `Bearer ${token}` }
        });
        
        if (hierarchyRes.status === 200) {
          console.log(`   ✅ L${courierInfo.level} 可以访问层级信息`);
          anySuccess = true;
        } else {
          console.log(`   ❌ L${courierInfo.level} 访问失败 (${hierarchyRes.status})`);
        }
      }
    }
    
    this.issues.hierarchyAccess.status = anySuccess ? 'partial' : 'failed';
    this.issues.hierarchyAccess.note = anySuccess ? 
      '部分信使级别可以访问' : '所有级别都无法访问';
    
    console.log('');
  }

  async testValidationErrorHandling(token) {
    console.log('⚠️  测试4: 验证错误处理');
    
    const invalidRequests = [
      {
        name: '空请求体',
        endpoint: '/api/v1/letters',
        method: 'POST',
        body: {}
      },
      {
        name: '无效字段类型',
        endpoint: '/api/v1/letters',
        method: 'POST',
        body: { title: 123, content: true }
      },
      {
        name: '超长内容',
        endpoint: '/api/v1/letters',
        method: 'POST',
        body: { 
          title: 'A'.repeat(1000), 
          content: 'B'.repeat(10000) 
        }
      }
    ];
    
    let properErrorHandling = 0;
    
    for (const test of invalidRequests) {
      console.log(`   测试: ${test.name}`);
      const res = await this.request(test.endpoint, {
        method: test.method,
        headers: { 'Authorization': `Bearer ${token}` },
        body: JSON.stringify(test.body)
      });
      
      // Check if we get proper validation error (400 or 422)
      if (res.status === 400 || res.status === 422) {
        console.log(`   ✅ 返回验证错误 (${res.status})`);
        if (res.data.message || res.data.error || res.data.errors) {
          console.log(`      错误信息: ${res.data.message || res.data.error}`);
          properErrorHandling++;
        }
      } else {
        console.log(`   ❌ 未返回预期的验证错误 (${res.status})`);
      }
    }
    
    this.issues.validationError.status = 
      properErrorHandling === invalidRequests.length ? 'fixed' : 
      properErrorHandling > 0 ? 'partial' : 'failed';
    
    console.log('');
  }

  generateReport() {
    console.log('=' + '='.repeat(60));
    console.log('📊 问题修复验证报告');
    console.log('=' + '='.repeat(60));
    
    let fixed = 0;
    let partial = 0;
    let failed = 0;
    
    for (const [key, issue] of Object.entries(this.issues)) {
      const icon = 
        issue.status === 'fixed' ? '✅' :
        issue.status === 'partial' ? '⚠️' :
        issue.status === 'failed' ? '❌' : '⏳';
      
      console.log(`${icon} ${issue.description}: ${issue.status.toUpperCase()}`);
      
      if (issue.solution) {
        console.log(`   解决方案: ${issue.solution}`);
      }
      if (issue.note) {
        console.log(`   备注: ${issue.note}`);
      }
      
      if (issue.status === 'fixed') fixed++;
      else if (issue.status === 'partial') partial++;
      else if (issue.status === 'failed') failed++;
    }
    
    console.log('\n总结:');
    console.log(`✅ 已修复: ${fixed}`);
    console.log(`⚠️  部分修复: ${partial}`);
    console.log(`❌ 未修复: ${failed}`);
    
    // Save detailed report
    const report = {
      timestamp: new Date().toISOString(),
      issues: this.issues,
      summary: { fixed, partial, failed }
    };
    
    fs.writeFileSync('issue-verification-report.json', JSON.stringify(report, null, 2));
    console.log('\n📄 详细报告已保存至: issue-verification-report.json');
    
    // Recommendations
    console.log('\n建议:');
    if (this.issues.letterCreation.status !== 'fixed') {
      console.log('1. 检查信件创建的验证规则，可能需要调整必填字段');
    }
    if (this.issues.courierFieldTransform.status !== 'fixed') {
      console.log('2. 确保信使模型的所有字段都包含在转换中间件中');
    }
    if (this.issues.hierarchyAccess.status !== 'fixed') {
      console.log('3. 检查层级API的权限设置，可能需要调整访问控制');
    }
    if (this.issues.validationError.status !== 'fixed') {
      console.log('4. 统一验证错误的响应格式和状态码');
    }
  }
}

// Run the test
const test = new IssueVerificationTest();
test.runTests().catch(console.error);