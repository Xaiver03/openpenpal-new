/**
 * 核心SOTA功能快速测试
 */

const http = require('http');

async function quickTest() {
  console.log('🚀 SOTA核心功能快速测试\n');
  
  const tests = {
    '路由别名': 0,
    '字段转换': 0,
    'AI功能': 0,
    '认证流程': 0
  };
  
  // Test 1: Route Alias
  console.log('1️⃣  测试路由别名...');
  try {
    const res = await fetch('http://localhost:8080/api/schools');
    if (res.ok) {
      console.log('   ✅ /api/schools → /api/v1/schools');
      tests['路由别名']++;
    }
  } catch (e) {
    console.log('   ❌ 路由别名失败');
  }
  
  // Test 2: Field Transformation
  console.log('\n2️⃣  测试字段转换...');
  try {
    // Get CSRF
    const csrfRes = await fetch('http://localhost:8080/api/auth/csrf');
    const csrfData = await csrfRes.json();
    const csrfToken = csrfData.data?.csrfToken;
    
    // Login
    const loginRes = await fetch('http://localhost:8080/api/v1/auth/login', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-CSRF-Token': csrfToken
      },
      body: JSON.stringify({ username: 'alice', password: 'secret' })
    });
    
    const loginData = await loginRes.json();
    const user = loginData.data?.user;
    
    if (user && 'createdAt' in user && 'isActive' in user) {
      console.log('   ✅ created_at → createdAt');
      console.log('   ✅ is_active → isActive');
      tests['字段转换']++;
    }
    
    if (loginData.data?.token) {
      tests['认证流程']++;
      console.log('   ✅ 登录成功，获得JWT令牌');
    }
  } catch (e) {
    console.log('   ❌ 字段转换测试失败:', e.message);
  }
  
  // Test 3: AI
  console.log('\n3️⃣  测试AI功能...');
  try {
    const aiRes = await fetch('http://localhost:8080/api/v1/ai/inspiration', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ theme: '友谊', count: 1 })
    });
    
    const aiData = await aiRes.json();
    if (aiData.data?.inspirations?.length > 0) {
      console.log('   ✅ AI灵感生成成功');
      const inspiration = aiData.data.inspirations[0];
      console.log(`   📝 "${inspiration.prompt.substring(0, 50)}..."`);
      tests['AI功能']++;
    }
  } catch (e) {
    console.log('   ❌ AI测试失败');
  }
  
  // Summary
  console.log('\n' + '='.repeat(50));
  console.log('📊 测试总结');
  console.log('='.repeat(50));
  
  const total = Object.values(tests).reduce((a, b) => a + b, 0);
  const maxScore = Object.keys(tests).length;
  
  for (const [category, score] of Object.entries(tests)) {
    console.log(`${score > 0 ? '✅' : '❌'} ${category}: ${score > 0 ? '通过' : '失败'}`);
  }
  
  console.log(`\n总分: ${total}/${maxScore} (${(total/maxScore*100).toFixed(0)}%)`);
  
  if (total === maxScore) {
    console.log('\n🎉 所有SOTA核心功能测试通过！');
  } else {
    console.log('\n⚠️  部分功能需要检查');
  }
}

quickTest().catch(console.error);