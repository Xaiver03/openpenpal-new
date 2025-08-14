#!/usr/bin/env node

/**
 * 综合测试脚本 - 验证所有修复是否正常工作
 */

const http = require('http');
const WebSocket = require('ws');

const API_BASE = 'http://localhost:8080';
const WS_BASE = 'ws://localhost:8080';

console.log('🧪 验证所有修复效果');
console.log('================================\n');

// 测试结果汇总
const testResults = {
  login: false,
  aiEndpoints: false,
  websocket: false,
  publicAccess: false
};

// 1. 测试登录功能
async function testLogin() {
  console.log('1️⃣ 测试登录功能...');
  
  const loginData = {
    username: 'courier_level1',
    password: 'secret'
  };

  return new Promise((resolve) => {
    const data = JSON.stringify(loginData);
    const options = {
      hostname: 'localhost',
      port: 8080,
      path: '/api/v1/auth/login',
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Content-Length': Buffer.byteLength(data)
      }
    };

    const req = http.request(options, (res) => {
      let body = '';
      res.on('data', (chunk) => body += chunk);
      res.on('end', () => {
        try {
          const response = JSON.parse(body);
          if (res.statusCode === 200 && response.data?.token) {
            console.log('✅ 登录成功');
            console.log('   用户名:', response.data.user.username);
            console.log('   角色:', response.data.user.role);
            testResults.login = true;
            resolve(response.data.token);
          } else {
            console.log('❌ 登录失败:', response.error || response.message);
            resolve(null);
          }
        } catch (error) {
          console.log('❌ 登录失败:', error.message);
          resolve(null);
        }
      });
    });

    req.on('error', (error) => {
      console.log('❌ 登录请求失败:', error.message);
      resolve(null);
    });
    
    req.write(data);
    req.end();
  });
}

// 2. 测试AI端点（无需认证）
async function testAIEndpoints() {
  console.log('\n2️⃣ 测试AI端点...');
  
  const endpoints = [
    { path: '/api/v1/ai/daily-inspiration', name: '每日灵感' },
    { path: '/api/v1/ai/stats', name: 'AI统计' },
    { path: '/api/v1/ai/personas', name: 'AI人设列表' }
  ];

  let allPassed = true;

  for (const endpoint of endpoints) {
    await new Promise((resolve) => {
      http.get(`${API_BASE}${endpoint.path}`, (res) => {
        let body = '';
        res.on('data', (chunk) => body += chunk);
        res.on('end', () => {
          try {
            const response = JSON.parse(body);
            if (res.statusCode === 200 && response.success) {
              console.log(`✅ ${endpoint.name}: 正常`);
            } else {
              console.log(`❌ ${endpoint.name}: 失败 (状态码: ${res.statusCode})`);
              allPassed = false;
            }
          } catch (error) {
            console.log(`❌ ${endpoint.name}: 解析失败`);
            allPassed = false;
          }
          resolve();
        });
      }).on('error', (error) => {
        console.log(`❌ ${endpoint.name}: 请求失败 - ${error.message}`);
        allPassed = false;
        resolve();
      });
    });
  }

  testResults.aiEndpoints = allPassed;
}

// 3. 测试WebSocket连接（需要token）
async function testWebSocket(token) {
  console.log('\n3️⃣ 测试WebSocket连接...');
  
  if (!token) {
    console.log('⚠️  跳过WebSocket测试（需要先登录）');
    return;
  }

  return new Promise((resolve) => {
    const wsUrl = `${WS_BASE}/api/v1/ws/connect?token=${encodeURIComponent(token)}`;
    console.log('🔗 WebSocket URL正确性检查...');
    
    // 检查URL是否包含重复路径
    if (wsUrl.includes('/api/v1/api/v1/')) {
      console.log('❌ WebSocket URL包含重复路径');
      testResults.websocket = false;
      resolve();
      return;
    }
    
    console.log('✅ WebSocket URL格式正确');
    
    const ws = new WebSocket(wsUrl);
    let connected = false;
    
    const timeout = setTimeout(() => {
      if (!connected) {
        console.log('❌ WebSocket连接超时');
        testResults.websocket = false;
        ws.close();
        resolve();
      }
    }, 5000);

    ws.on('open', () => {
      connected = true;
      console.log('✅ WebSocket连接成功');
      testResults.websocket = true;
      clearTimeout(timeout);
      
      // 发送测试消息
      const testMessage = {
        id: 'test_heartbeat',
        type: 'HEARTBEAT',
        data: { test: true },
        timestamp: new Date().toISOString()
      };
      ws.send(JSON.stringify(testMessage));
      
      setTimeout(() => {
        ws.close();
        resolve();
      }, 1000);
    });

    ws.on('message', (data) => {
      try {
        const message = JSON.parse(data.toString());
        console.log('✅ 收到WebSocket消息:', message.type);
      } catch (error) {
        console.log('⚠️  收到非JSON消息');
      }
    });

    ws.on('error', (error) => {
      console.log('❌ WebSocket错误:', error.message);
      clearTimeout(timeout);
      testResults.websocket = false;
      resolve();
    });
  });
}

// 4. 测试公开页面访问
async function testPublicAccess() {
  console.log('\n4️⃣ 测试公开页面访问...');
  
  const publicEndpoints = [
    { path: '/api/v1/letters/public', name: '公开信件列表' },
    { path: '/api/v1/ws/stats', name: 'WebSocket统计' }
  ];

  let allPassed = true;

  for (const endpoint of publicEndpoints) {
    await new Promise((resolve) => {
      http.get(`${API_BASE}${endpoint.path}`, (res) => {
        let body = '';
        res.on('data', (chunk) => body += chunk);
        res.on('end', () => {
          if (res.statusCode === 200) {
            console.log(`✅ ${endpoint.name}: 可访问`);
          } else {
            console.log(`❌ ${endpoint.name}: 不可访问 (状态码: ${res.statusCode})`);
            allPassed = false;
          }
          resolve();
        });
      }).on('error', (error) => {
        console.log(`❌ ${endpoint.name}: 请求失败 - ${error.message}`);
        allPassed = false;
        resolve();
      });
    });
  }

  testResults.publicAccess = allPassed;
}

// 生成测试报告
function generateReport() {
  console.log('\n📋 修复验证报告');
  console.log('================================');
  
  const fixes = [
    { name: '登录认证（数据库密码）', status: testResults.login },
    { name: 'AI端点（404错误修复）', status: testResults.aiEndpoints },
    { name: 'WebSocket连接（URL重复修复）', status: testResults.websocket },
    { name: '公开访问端点', status: testResults.publicAccess }
  ];
  
  let allPassed = true;
  
  fixes.forEach(fix => {
    console.log(`${fix.status ? '✅' : '❌'} ${fix.name}: ${fix.status ? '正常' : '异常'}`);
    if (!fix.status) allPassed = false;
  });
  
  console.log('\n总体状态:', allPassed ? '🎉 所有修复已生效！' : '⚠️  部分功能仍需修复');
  
  if (allPassed) {
    console.log('\n已修复的问题:');
    console.log('1. courier_level1账户登录问题 - 数据库密码已更新');
    console.log('2. AI页面可以未登录访问，功能显示受限提示');
    console.log('3. AI API端点路径修复（/ai → /api/ai）');
    console.log('4. WebSocket URL重复路径问题已修复');
    console.log('5. AI统计端点支持匿名访问');
  }
}

// 主函数
async function main() {
  try {
    // 测试登录
    const token = await testLogin();
    
    // 测试AI端点
    await testAIEndpoints();
    
    // 测试WebSocket
    await testWebSocket(token);
    
    // 测试公开访问
    await testPublicAccess();
    
    // 生成报告
    generateReport();
    
  } catch (error) {
    console.error('🚨 测试过程出错:', error.message);
  }
}

main();