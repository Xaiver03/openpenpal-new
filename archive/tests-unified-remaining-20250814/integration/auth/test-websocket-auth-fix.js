// 测试WebSocket认证修复
const WebSocket = require('ws');

console.log('=== WebSocket认证修复测试 ===');

// 1. 先测试登录获取token
async function testLogin() {
  const loginData = {
    username: 'courier_level1',
    password: 'courier123'
  };

  try {
    console.log('🔐 正在登录用户:', loginData.username);
    
    const response = await fetch('http://localhost:8080/api/v1/auth/login', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(loginData),
    });

    const result = await response.json();
    console.log('✅ 登录响应状态:', response.status);
    console.log('✅ 登录结果:', result.code === 0 ? '成功' : '失败');
    
    if (result.code === 0) {
      const token = result.data.accessToken || result.data.token;
      console.log('✅ 获取到Token:', token ? token.substring(0, 20) + '...' : 'No token');
      return token;
    } else {
      console.log('❌ 登录失败:', result.message);
      return null;
    }
  } catch (error) {
    console.log('❌ 登录请求失败:', error.message);
    return null;
  }
}

// 2. 测试WebSocket连接
async function testWebSocketConnection(token) {
  if (!token) {
    console.log('❌ 无Token，跳过WebSocket测试');
    return;
  }

  console.log('\n📡 测试WebSocket连接...');
  
  const wsUrl = `ws://localhost:8080/api/v1/ws/connect?token=${encodeURIComponent(token)}`;
  console.log('📡 连接URL:', wsUrl.replace(token, 'TOKEN_HIDDEN'));

  return new Promise((resolve, reject) => {
    const ws = new WebSocket(wsUrl);
    let resolved = false;

    const cleanup = () => {
      if (!resolved) {
        resolved = true;
        ws.terminate();
      }
    };

    // 设置超时
    const timeout = setTimeout(() => {
      console.log('⏰ WebSocket连接超时');
      cleanup();
      resolve(false);
    }, 10000);

    ws.on('open', () => {
      console.log('✅ WebSocket连接成功！认证修复生效！');
      clearTimeout(timeout);
      
      // 发送心跳测试
      ws.send(JSON.stringify({ type: 'ping' }));
      
      setTimeout(() => {
        ws.close(1000, 'Test completed');
        cleanup();
        resolve(true);
      }, 2000);
    });

    ws.on('message', (data) => {
      try {
        const message = JSON.parse(data.toString());
        console.log('📨 收到消息:', message);
      } catch (e) {
        console.log('📨 收到原始消息:', data.toString());
      }
    });

    ws.on('error', (error) => {
      console.log('❌ WebSocket错误:', error.message);
      clearTimeout(timeout);
      cleanup();
      resolve(false);
    });

    ws.on('close', (code, reason) => {
      console.log('📡 WebSocket关闭:', code, reason.toString());
      clearTimeout(timeout);
      cleanup();
      if (!resolved) resolve(code === 1000);
    });
  });
}

// 3. 运行完整测试
async function runTest() {
  console.log('开始时间:', new Date().toLocaleString());
  
  const token = await testLogin();
  const wsSuccess = await testWebSocketConnection(token);
  
  console.log('\n=== 测试结果总结 ===');
  console.log('登录状态:', token ? '✅ 成功' : '❌ 失败');
  console.log('WebSocket状态:', wsSuccess ? '✅ 成功' : '❌ 失败');
  console.log('整体修复:', (token && wsSuccess) ? '✅ 完全成功' : '❌ 需要进一步修复');
  
  console.log('\n结束时间:', new Date().toLocaleString());
  process.exit(wsSuccess ? 0 : 1);
}

runTest();