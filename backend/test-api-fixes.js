/**
 * API修复验证测试
 */

const http = require('http');

async function testAPIFixes() {
  console.log('🔧 API修复验证测试\n');
  
  // Helper function
  async function apiRequest(path, options = {}) {
    return new Promise((resolve, reject) => {
      const url = new URL(path, 'http://localhost:8080');
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
            resolve({ status: res.statusCode, data: data, headers: res.headers });
          }
        });
      });
      
      req.on('error', reject);
      if (options.body) req.write(options.body);
      req.end();
    });
  }
  
  // Get auth token
  console.log('1️⃣  获取认证令牌...');
  const csrfRes = await apiRequest('/api/auth/csrf');
  const csrfToken = csrfRes.data.data?.csrfToken;
  
  const loginRes = await apiRequest('/api/v1/auth/login', {
    method: 'POST',
    headers: { 'X-CSRF-Token': csrfToken },
    body: JSON.stringify({ username: 'alice', password: 'secret' })
  });
  
  const token = loginRes.data.data?.token;
  console.log(token ? '✅ 登录成功\n' : '❌ 登录失败\n');
  
  // Test 1: Letter creation with correct endpoint
  console.log('2️⃣  测试信件创建（修正后）...');
  
  // Note: Remove trailing slash to avoid 307 redirect
  const letterRes = await apiRequest('/api/v1/letters', {
    method: 'POST',
    headers: { 'Authorization': `Bearer ${token}` },
    body: JSON.stringify({
      title: '测试信件',
      content: '这是测试内容',
      style: 'warm',
      visibility: 'private'
    })
  });
  
  console.log(`   状态码: ${letterRes.status}`);
  if (letterRes.status === 201 || letterRes.status === 200) {
    console.log('   ✅ 信件创建成功');
    if (letterRes.data.data) {
      const letter = letterRes.data.data;
      console.log(`   信件ID: ${letter.id}`);
      console.log(`   字段检查: ${letter.createdAt ? '✅ camelCase' : '❌ 未转换'}`);
    }
  } else {
    console.log(`   ❌ 创建失败: ${letterRes.data.message || letterRes.data.error || '未知错误'}`);
  }
  
  // Test 2: Courier field transformation
  console.log('\n3️⃣  测试信使字段转换...');
  
  // Login as courier
  const courierLoginRes = await apiRequest('/api/v1/auth/login', {
    method: 'POST',
    headers: { 'X-CSRF-Token': csrfToken },
    body: JSON.stringify({ username: 'courier_level1', password: 'secret' })
  });
  
  if (courierLoginRes.status === 200) {
    const courierToken = courierLoginRes.data.data?.token;
    
    // Get courier profile
    const profileRes = await apiRequest('/api/v1/courier/profile', {
      headers: { 'Authorization': `Bearer ${courierToken}` }
    });
    
    console.log(`   状态码: ${profileRes.status}`);
    if (profileRes.status === 200 && profileRes.data.data) {
      const courier = profileRes.data.data;
      console.log('   ✅ 信使信息获取成功');
      
      // Check transformed fields
      const transformedFields = [
        'createdAt',
        'updatedAt',
        'deletedAt',
        'userId'
      ];
      
      transformedFields.forEach(field => {
        if (field in courier) {
          console.log(`   ✅ ${field} 已转换`);
        }
      });
    }
  }
  
  // Test 3: Validation errors
  console.log('\n4️⃣  测试验证错误处理...');
  
  const invalidRes = await apiRequest('/api/v1/letters', {
    method: 'POST',
    headers: { 'Authorization': `Bearer ${token}` },
    body: JSON.stringify({}) // Empty body
  });
  
  console.log(`   状态码: ${invalidRes.status}`);
  if (invalidRes.status === 400 || invalidRes.status === 422) {
    console.log('   ✅ 正确返回验证错误');
    console.log(`   错误信息: ${invalidRes.data.message || invalidRes.data.error}`);
  } else {
    console.log('   ❌ 验证错误处理不正确');
  }
  
  // Test 4: Check available APIs
  console.log('\n5️⃣  检查可用的API端点...');
  
  const endpoints = [
    { path: '/api/v1/courier/me', name: '信使个人信息' },
    { path: '/api/v1/courier/tasks', name: '信使任务列表' },
    { path: '/api/v1/courier/hierarchy', name: '信使层级信息' },
    { path: '/api/v1/letters', name: '信件列表' },
    { path: '/api/v1/museum/entries', name: '博物馆条目' }
  ];
  
  for (const endpoint of endpoints) {
    const res = await apiRequest(endpoint.path, {
      headers: { 'Authorization': `Bearer ${token}` }
    });
    console.log(`   ${endpoint.name}: ${res.status === 200 ? '✅' : '❌'} (${res.status})`);
  }
  
  console.log('\n✨ 测试完成');
}

testAPIFixes().catch(console.error);