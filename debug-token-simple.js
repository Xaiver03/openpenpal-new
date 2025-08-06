// 简单的token调试脚本 - 在浏览器控制台运行
console.log('=== Token 调试 ===');

// 1. 获取token
const token = localStorage.getItem('openpenpal_auth_token') || 
              document.cookie.split(';').find(c => c.trim().startsWith('openpenpal_auth_token='))?.split('=')[1];

console.log('Token存在:', !!token);
if (token) {
  console.log('Token (前50字符):', token.substring(0, 50) + '...');
  
  try {
    // 2. 解析token
    const parts = token.split('.');
    console.log('Token部分数量:', parts.length);
    
    if (parts.length === 3) {
      const header = JSON.parse(atob(parts[0]));
      const payload = JSON.parse(atob(parts[1]));
      
      console.log('Token Header:', header);
      console.log('Token Payload:', payload);
      
      // 3. 检查过期时间
      if (payload.exp) {
        const expDate = new Date(payload.exp * 1000);
        const now = new Date();
        console.log('过期时间:', expDate.toLocaleString());
        console.log('当前时间:', now.toLocaleString());
        console.log('是否过期:', now >= expDate);
        console.log('剩余时间:', Math.round((expDate - now) / 1000 / 60), '分钟');
      } else {
        console.log('❌ Token没有过期时间字段');
      }
      
      // 4. 检查必要字段
      console.log('用户ID:', payload.userId || payload.user_id || payload.sub);
      console.log('用户角色:', payload.role);
      console.log('签发时间:', payload.iat ? new Date(payload.iat * 1000).toLocaleString() : 'N/A');
    } else {
      console.log('❌ Token格式错误：不是标准JWT格式');
    }
  } catch (error) {
    console.log('❌ Token解析失败:', error.message);
  }
}

// 5. 测试API调用
if (token) {
  console.log('测试API调用...');
  fetch('/api/auth/me', {
    headers: { 'Authorization': `Bearer ${token}` }
  })
  .then(response => {
    console.log('API响应状态:', response.status);
    console.log('API响应头:', Object.fromEntries(response.headers.entries()));
    return response.json();
  })
  .then(data => {
    console.log('API响应数据:', data);
  })
  .catch(error => {
    console.log('API调用失败:', error);
  });
}

console.log('=== 调试完成 ===');