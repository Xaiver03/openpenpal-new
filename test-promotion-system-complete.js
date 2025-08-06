// 完整的晋升系统测试 - SOTA级别验证
const API_URL = 'http://localhost:8080';

// Store cookies manually
let cookies = {};

// Helper to parse Set-Cookie headers
function parseCookies(setCookieHeaders) {
  if (!setCookieHeaders) return;
  const headers = Array.isArray(setCookieHeaders) ? setCookieHeaders : [setCookieHeaders];
  headers.forEach(header => {
    const [cookie] = header.split(';');
    const [name, value] = cookie.split('=');
    cookies[name] = value;
  });
}

// Helper to create Cookie header
function getCookieHeader() {
  return Object.entries(cookies)
    .map(([name, value]) => `${name}=${value}`)
    .join('; ');
}

// Helper function to get CSRF token
async function getCSRFToken() {
  const response = await fetch(`${API_URL}/api/v1/auth/csrf`, {
    headers: {
      'Cookie': getCookieHeader()
    }
  });
  parseCookies(response.headers.get('set-cookie'));
  const data = await response.json();
  return data.data.token;
}

// Helper function to login
async function login(username, password) {
  const csrfToken = await getCSRFToken();
  
  const response = await fetch(`${API_URL}/api/v1/auth/login`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'X-CSRF-Token': csrfToken,
      'Cookie': getCookieHeader()
    },
    body: JSON.stringify({ username, password })
  });
  
  parseCookies(response.headers.get('set-cookie'));
  
  if (!response.ok) {
    const error = await response.text();
    console.log('Login error:', error);
    throw new Error(`Login failed: ${response.status}`);
  }
  
  const result = await response.json();
  return { token: result.data.token, user: result.data.user, csrfToken };
}

// 测试晋升申请提交
async function testPromotionApplication(auth) {
  console.log('\n🧪 测试晋升申请提交...');
  
  const applicationData = {
    request_level: 2,
    reason: '已达到所有晋升要求，申请成为二级信使。在过去的3个月中表现优秀，完成了69次投递任务，成功率达到96.8%，希望承担更多责任，为团队做出更大贡献。',
    evidence: {
      deliveries: 69,
      success_rate: 96.8,
      service_days: 45,
      performance_score: 88,
      complaints: 1,
      feedback_rating: 4.7
    }
  };
  
  try {
    const response = await fetch(`${API_URL}/api/v1/courier/growth/apply`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${auth.token}`,
        'Cookie': getCookieHeader()
      },
      body: JSON.stringify(applicationData)
    });

    if (response.ok) {
      const result = await response.json();
      console.log('✅ 晋升申请提交成功！');
      console.log(JSON.stringify(result, null, 2));
      return result.data;
    } else {
      const error = await response.text();
      console.log('❌ 晋升申请提交失败:', response.status);
      console.log(error);
      return null;
    }
  } catch (error) {
    console.log('❌ 晋升申请网络错误:', error.message);
    return null;
  }
}

// 测试获取申请列表
async function testGetApplications(auth) {
  console.log('\n🧪 测试获取申请列表...');
  
  try {
    const response = await fetch(`${API_URL}/api/v1/courier/growth/applications?status=pending&limit=10`, {
      headers: {
        'Authorization': `Bearer ${auth.token}`,
        'Cookie': getCookieHeader()
      }
    });

    if (response.ok) {
      const result = await response.json();
      console.log('✅ 申请列表获取成功！');
      console.log(JSON.stringify(result, null, 2));
      return result.data.requests;
    } else {
      const error = await response.text();
      console.log('❌ 获取申请列表失败:', response.status);
      console.log(error);
      return [];
    }
  } catch (error) {
    console.log('❌ 获取申请列表网络错误:', error.message);
    return [];
  }
}

// 测试处理申请
async function testProcessApplication(auth, requestId) {
  console.log('\n🧪 测试处理申请...');
  
  const processData = {
    action: 'approve',
    comment: '表现优秀，各项指标均已达标，同意晋升为二级信使。希望继续努力，为团队做出更大贡献。'
  };
  
  try {
    const response = await fetch(`${API_URL}/api/v1/courier/growth/applications/${requestId}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${auth.token}`,
        'Cookie': getCookieHeader()
      },
      body: JSON.stringify(processData)
    });

    if (response.ok) {
      const result = await response.json();
      console.log('✅ 申请处理成功！');
      console.log(JSON.stringify(result, null, 2));
      return result.data;
    } else {
      const error = await response.text();
      console.log('❌ 申请处理失败:', response.status);
      console.log(error);
      return null;
    }
  } catch (error) {
    console.log('❌ 申请处理网络错误:', error.message);
    return null;
  }
}

// 完整的晋升系统测试流程
async function runCompletePromotionTest() {
  console.log(`
  ============================================================
  🎨 OpenPenPal 完整晋升系统测试 (SOTA级别)
  ============================================================
  `);

  try {
    // 1. 测试Level 1用户提交申请
    console.log('\n【第一步】Level 1 信使提交晋升申请');
    console.log('============================================================');
    
    const courier1Auth = await login('courier1', 'password');
    console.log('✅ courier1 登录成功');
    
    // 获取当前成长路径
    const pathResponse = await fetch(`${API_URL}/api/v1/courier/growth/path`, {
      headers: {
        'Authorization': `Bearer ${courier1Auth.token}`,
        'Cookie': getCookieHeader()
      }
    });
    
    if (pathResponse.ok) {
      const pathData = await pathResponse.json();
      console.log('📊 当前成长路径:');
      console.log(`   - 当前等级: ${pathData.data.current_level} (${pathData.data.current_name})`);
      console.log(`   - 晋升进度: ${pathData.data.paths[0]?.completion_rate?.toFixed(2)}%`);
      console.log(`   - 可否晋升: ${pathData.data.paths[0]?.can_upgrade ? '是' : '否'}`);
    }
    
    // 提交晋升申请
    const newApplication = await testPromotionApplication(courier1Auth);
    
    // 2. 测试Level 3管理员处理申请
    console.log('\n【第二步】Level 3 信使审核申请');
    console.log('============================================================');
    
    const courier3Auth = await login('courier_level3', 'secret');
    console.log('✅ courier_level3 登录成功 (管理员权限)');
    
    // 获取待处理申请列表
    const pendingApplications = await testGetApplications(courier3Auth);
    
    if (pendingApplications && pendingApplications.length > 0) {
      const latestRequest = pendingApplications[0];
      console.log(`📋 找到待处理申请: ${latestRequest.id}`);
      
      // 处理申请
      await testProcessApplication(courier3Auth, latestRequest.id);
    } else {
      console.log('📋 没有待处理的申请');
    }
    
    // 3. 验证晋升结果
    console.log('\n【第三步】验证晋升结果');
    console.log('============================================================');
    
    // 重新获取courier1的成长路径，验证是否晋升成功
    const updatedPathResponse = await fetch(`${API_URL}/api/v1/courier/growth/path`, {
      headers: {
        'Authorization': `Bearer ${courier1Auth.token}`,
        'Cookie': getCookieHeader()
      }
    });
    
    if (updatedPathResponse.ok) {
      const updatedPathData = await updatedPathResponse.json();
      console.log('📊 晋升后成长路径:');
      console.log(`   - 当前等级: ${updatedPathData.data.current_level} (${updatedPathData.data.current_name})`);
      console.log(`   - 下一等级: ${updatedPathData.data.paths[0]?.target_level || '已达最高级'} (${updatedPathData.data.paths[0]?.target_name || '无'})`);
    }
    
    // 4. 测试数据库完整性
    console.log('\n【第四步】验证数据库完整性');
    console.log('============================================================');
    
    // 获取所有申请记录，验证数据持久化
    const allApplications = await fetch(`${API_URL}/api/v1/courier/growth/applications?limit=50`, {
      headers: {
        'Authorization': `Bearer ${courier3Auth.token}`,
        'Cookie': getCookieHeader()
      }
    });
    
    if (allApplications.ok) {
      const allData = await allApplications.json();
      console.log(`📊 数据库完整性验证:`);
      console.log(`   - 总申请数: ${allData.data.total}`);
      console.log(`   - 当前页申请数: ${allData.data.requests.length}`);
      
      const statusCounts = {};
      allData.data.requests.forEach(req => {
        statusCounts[req.status] = (statusCounts[req.status] || 0) + 1;
      });
      
      console.log(`   - 申请状态分布:`, statusCounts);
    }
    
    console.log('\n🎉 完整晋升系统测试完成！');
    console.log('============================================================');
    
  } catch (error) {
    console.error('❌ 测试过程中发生错误:', error);
  }
}

// 运行完整测试
runCompletePromotionTest();