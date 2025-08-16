// 完整的认证状态调试和测试脚本
console.log('=== 认证状态完整测试 ===');

// 1. 检查Token状态
const token = localStorage.getItem('openpenpal_auth_token') || 
              document.cookie.split(';').find(c => c.trim().startsWith('openpenpal_auth_token='))?.split('=')[1];

console.log('1. Token检查:');
console.log('   存在:', !!token);

if (token) {
  console.log('   长度:', token.length);
  console.log('   前缀:', token.substring(0, 20) + '...');
  
  try {
    const parts = token.split('.');
    if (parts.length === 3) {
      const payload = JSON.parse(atob(parts[1]));
      console.log('   用户ID:', payload.userId);
      console.log('   角色:', payload.role);
      console.log('   过期时间:', new Date(payload.exp * 1000).toLocaleString());
      console.log('   是否过期:', Date.now() >= payload.exp * 1000);
    }
  } catch (e) {
    console.log('   ❌ Token解析失败:', e.message);
  }
}

// 2. 检查用户状态
const userCookie = document.cookie.split(';').find(c => c.trim().startsWith('openpenpal_user='));
const userLocal = localStorage.getItem('openpenpal_user');

console.log('\n2. 用户数据检查:');
console.log('   Cookie中存在:', !!userCookie);
console.log('   LocalStorage中存在:', !!userLocal);

// 3. 检查认证服务状态（如果可用）
if (typeof window.AuthStateFixer !== 'undefined') {
  console.log('\n3. 认证状态诊断:');
  const report = window.AuthStateFixer.generateDiagnosticReport();
  console.log(report);
}

// 4. 测试AI页面访问
console.log('\n4. 测试路由访问:');
console.log('   当前路径:', window.location.pathname);

if (window.location.pathname !== '/ai') {
  console.log('   准备测试AI页面访问...');
  // 不立即跳转，先检查状态
  setTimeout(() => {
    if (confirm('是否跳转到AI页面测试认证？')) {
      window.location.href = '/ai';
    }
  }, 2000);
}

// 5. WebSocket连接状态
console.log('\n5. WebSocket状态:');
// 检查WebSocket相关的错误日志
const wsErrors = console.error.calls?.filter(call => 
  call.arguments.some(arg => 
    typeof arg === 'string' && arg.includes('WebSocket')
  )
) || [];
console.log('   最近WebSocket错误数:', wsErrors.length);

console.log('\n=== 测试完成 ===');
console.log('如果看到持续的WebSocket错误但不影响页面功能，说明修复生效');