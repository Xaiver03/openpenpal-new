/**
 * 测试四级信使权限
 * 验证四级信使能否创建三级信使
 */

const http = require('http');

async function makeRequest(path, options = {}) {
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
          resolve({ status: res.statusCode, data: data });
        }
      });
    });
    
    req.on('error', reject);
    if (options.body) req.write(options.body);
    req.end();
  });
}

async function testL4CourierPermission() {
  console.log('🚴 测试四级信使权限\n');
  
  try {
    // Step 1: Get CSRF token
    console.log('1️⃣  获取CSRF令牌...');
    const csrfRes = await makeRequest('/api/auth/csrf');
    const csrfToken = csrfRes.data.data?.csrfToken;
    console.log('   ✅ CSRF令牌获取成功');
    
    // Step 2: Login as L4 courier
    console.log('\n2️⃣  四级信使登录...');
    const loginRes = await makeRequest('/api/v1/auth/login', {
      method: 'POST',
      headers: { 'X-CSRF-Token': csrfToken },
      body: JSON.stringify({ 
        username: 'courier_level4', 
        password: 'secret' 
      })
    });
    
    if (loginRes.status !== 200) {
      console.log('   ❌ 登录失败:', loginRes.data.message || loginRes.status);
      return;
    }
    
    const token = loginRes.data.data?.token;
    const user = loginRes.data.data?.user;
    console.log('   ✅ 登录成功');
    console.log(`   用户角色: ${user?.role}`);
    console.log(`   用户ID: ${user?.id}`);
    
    // Step 3: Get courier info
    console.log('\n3️⃣  获取信使信息...');
    const courierRes = await makeRequest('/api/v1/courier/me', {
      headers: { 'Authorization': `Bearer ${token}` }
    });
    
    if (courierRes.status === 200 && courierRes.data.data) {
      const courier = courierRes.data.data;
      console.log('   ✅ 信使信息获取成功');
      console.log(`   信使级别: L${courier.level}`);
      console.log(`   管理区域: ${courier.zone || '未分配'}`);
      console.log(`   管理OP码前缀: ${courier.managedOpCodePrefix || '未分配'}`);
    } else {
      console.log('   ❌ 获取信使信息失败');
    }
    
    // Step 4: Test creating L3 courier (should succeed)
    console.log('\n4️⃣  测试创建三级信使（应该成功）...');
    const createL3Res = await makeRequest('/api/v1/courier/create', {
      method: 'POST',
      headers: { 
        'Authorization': `Bearer ${token}`,
        'X-CSRF-Token': csrfToken
      },
      body: JSON.stringify({
        username: `test_l3_${Date.now()}`,
        email: `test_l3_${Date.now()}@test.com`,
        name: '测试三级信使',
        level: 3,
        zone: 'BJDX',
        school: 'Beijing University',
        contact: '13800138003',
        managedOpCodePrefix: 'BD'
      })
    });
    
    console.log(`   响应状态: ${createL3Res.status}`);
    if (createL3Res.status === 200 || createL3Res.status === 201) {
      console.log('   ✅ 成功创建三级信使');
      const newCourier = createL3Res.data.data;
      console.log(`   新信使用户名: ${newCourier?.username}`);
      console.log(`   新信使级别: L${newCourier?.courier?.level || 3}`);
    } else {
      console.log('   ❌ 创建失败:', createL3Res.data.message || createL3Res.data.error);
    }
    
    // Step 5: Test creating L4 courier (should fail)
    console.log('\n5️⃣  测试创建四级信使（应该失败）...');
    const createL4Res = await makeRequest('/api/v1/courier/create', {
      method: 'POST',
      headers: { 
        'Authorization': `Bearer ${token}`,
        'X-CSRF-Token': csrfToken
      },
      body: JSON.stringify({
        username: `test_l4_${Date.now()}`,
        email: `test_l4_${Date.now()}@test.com`,
        name: '测试四级信使',
        level: 4,
        zone: 'SHANGHAI',
        school: 'Shanghai',
        contact: '13800138004',
        managedOpCodePrefix: 'SH'
      })
    });
    
    console.log(`   响应状态: ${createL4Res.status}`);
    if (createL4Res.status === 400 || createL4Res.status === 403) {
      console.log('   ✅ 正确阻止创建同级信使');
      console.log(`   错误信息: ${createL4Res.data.message || createL4Res.data.error}`);
    } else if (createL4Res.status === 200 || createL4Res.status === 201) {
      console.log('   ❌ 错误地允许创建同级信使');
    }
    
    // Step 6: Check permissions
    console.log('\n6️⃣  检查权限信息...');
    const permissionsRes = await makeRequest('/api/v1/courier/permissions', {
      headers: { 'Authorization': `Bearer ${token}` }
    });
    
    if (permissionsRes.status === 200 && permissionsRes.data.data) {
      const permissions = permissionsRes.data.data;
      console.log('   权限列表:');
      if (Array.isArray(permissions)) {
        permissions.forEach(p => console.log(`   - ${p}`));
      } else if (permissions.permissions) {
        permissions.permissions.forEach(p => console.log(`   - ${p}`));
      }
    }
    
  } catch (error) {
    console.error('\n❌ 测试错误:', error.message);
  }
  
  // Summary
  console.log('\n' + '='.repeat(50));
  console.log('📊 测试总结');
  console.log('='.repeat(50));
  console.log('\n如果四级信使能够成功创建三级信使，但不能创建四级信使，');
  console.log('那么权限系统工作正常。\n');
  console.log('前端页面应该也能正常使用"创建下级信使"功能了。');
}

// Run the test
testL4CourierPermission().catch(console.error);