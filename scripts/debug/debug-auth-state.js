// 在浏览器控制台运行此脚本来调试认证状态
console.log('=== 认证状态调试 ===');

// 1. 检查localStorage中的token
const token = localStorage.getItem('auth_token') || localStorage.getItem('token') || localStorage.getItem('access_token');
console.log('LocalStorage Token:', token ? token.substring(0, 50) + '...' : 'NOT FOUND');

// 2. 检查sessionStorage中的token
const sessionToken = sessionStorage.getItem('auth_token') || sessionStorage.getItem('token');
console.log('SessionStorage Token:', sessionToken ? sessionToken.substring(0, 50) + '...' : 'NOT FOUND');

// 3. 检查cookies
document.cookie.split(';').forEach(cookie => {
  const [name, value] = cookie.trim().split('=');
  if (name.includes('token') || name.includes('auth') || name.includes('session')) {
    console.log(`Cookie ${name}:`, value ? value.substring(0, 50) + '...' : 'EMPTY');
  }
});

// 4. 解析JWT token（如果存在）
function parseJWT(token) {
  try {
    const base64Url = token.split('.')[1];
    const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
    const jsonPayload = decodeURIComponent(window.atob(base64).split('').map(function(c) {
      return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2);
    }).join(''));
    return JSON.parse(jsonPayload);
  } catch (e) {
    return null;
  }
}

if (token) {
  const payload = parseJWT(token);
  if (payload) {
    console.log('Token Payload:', payload);
    console.log('Token Expires:', new Date(payload.exp * 1000));
    console.log('Current Time:', new Date());
    console.log('Token Expired?', payload.exp * 1000 < Date.now());
  }
}

// 5. 测试 /api/auth/me 接口
if (token) {
  fetch('/api/auth/me', {
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    }
  })
  .then(response => {
    console.log('Auth API Status:', response.status);
    console.log('Auth API Headers:', Object.fromEntries(response.headers.entries()));
    return response.json();
  })
  .then(data => {
    console.log('Auth API Response:', data);
  })
  .catch(err => {
    console.error('Auth API Error:', err);
  });
}

// 6. 检查认证状态同步服务
if (window.authSyncService) {
  console.log('Auth Sync Service State:', window.authSyncService.getState());
} else {
  console.log('Auth Sync Service: NOT FOUND');
}

// 7. 检查Zustand store状态
if (window.__ZUSTAND_DEVTOOLS_GLOBAL_HOOK__) {
  console.log('Zustand stores available');
}

console.log('=== 调试完成 ===');