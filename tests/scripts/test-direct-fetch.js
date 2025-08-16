// 直接测试 fetch 调用，模拟前端的行为

async function testDirectFetch() {
  console.log('🧪 Testing direct fetch from Node.js...\n');
  
  // 测试1：使用 Node.js 的 fetch（Next.js 使用的）
  console.log('📍 Test 1: Using Node.js fetch (like Next.js)');
  try {
    const response = await fetch('http://localhost:8000/api/v1/auth/login', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ username: 'admin', password: 'password' }),
    });
    
    console.log('   Status:', response.status);
    console.log('   Headers:', Object.fromEntries(response.headers.entries()));
    
    const text = await response.text();
    console.log('   Response length:', text.length);
    console.log('   First 200 chars:', text.substring(0, 200));
    console.log('   Last 200 chars:', text.substring(Math.max(0, text.length - 200)));
    
    // 检查是否有 BOM 或其他隐藏字符
    console.log('   First 10 bytes (hex):', Buffer.from(text.substring(0, 10)).toString('hex'));
    
    // 尝试解析
    try {
      const json = JSON.parse(text);
      console.log('   ✅ JSON parsed successfully');
    } catch (e) {
      console.log('   ❌ JSON parse error:', e.message);
      
      // 查找 JSON 开始位置
      const jsonStart = text.indexOf('{');
      console.log('   JSON starts at position:', jsonStart);
      if (jsonStart >= 0) {
        console.log('   Characters before JSON:', JSON.stringify(text.substring(0, jsonStart)));
      }
    }
    
  } catch (error) {
    console.error('   ❌ Fetch error:', error.message);
  }
  
  // 测试2：使用 axios（作为对比）
  console.log('\n📍 Test 2: Using axios (for comparison)');
  const axios = require('axios');
  try {
    const response = await axios.post('http://localhost:8000/api/v1/auth/login', {
      username: 'admin',
      password: 'password'
    }, {
      validateStatus: () => true
    });
    
    console.log('   Status:', response.status);
    console.log('   Data type:', typeof response.data);
    console.log('   Data:', response.data ? 'Has data' : 'No data');
    
  } catch (error) {
    console.error('   ❌ Axios error:', error.message);
  }
  
  // 测试3：检查代理环境变量
  console.log('\n📍 Test 3: Environment check');
  console.log('   HTTP_PROXY:', process.env.HTTP_PROXY || 'not set');
  console.log('   http_proxy:', process.env.http_proxy || 'not set');
  console.log('   NO_PROXY:', process.env.NO_PROXY || 'not set');
}

testDirectFetch();