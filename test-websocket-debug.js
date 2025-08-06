#!/usr/bin/env node

/**
 * WebSocket调试工具
 * 系统性检查WebSocket连接问题
 */

const https = require('https');
const http = require('http');
const crypto = require('crypto');

// 配置
const API_BASE = 'http://localhost:8080';
const WS_BASE = 'ws://localhost:8080';

console.log('🔍 WebSocket系统调试工具');
console.log('================================\n');

// 1. 登录获取token
async function login() {
  console.log('1️⃣ 登录获取Token...');
  
  const loginData = {
    username: 'courier_level1',
    password: 'secret'
  };

  return new Promise((resolve, reject) => {
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
          if (res.statusCode === 200 && response.data.token) {
            console.log('✅ 登录成功');
            console.log(`   Token: ${response.data.token.substring(0, 20)}...`);
            resolve(response.data.token);
          } else {
            console.log('❌ 登录失败:', response);
            reject(new Error('Login failed'));
          }
        } catch (error) {
          console.log('❌ 解析登录响应失败:', error.message);
          reject(error);
        }
      });
    });

    req.on('error', (error) => {
      console.log('❌ 登录请求失败:', error.message);
      reject(error);
    });

    req.write(data);
    req.end();
  });
}

// 2. 验证Token格式和内容
function analyzeToken(token) {
  console.log('\n2️⃣ 分析Token结构...');
  
  try {
    const parts = token.split('.');
    if (parts.length !== 3) {
      console.log('❌ Token格式错误，不是标准JWT格式');
      return false;
    }

    // 解析Header
    const header = JSON.parse(Buffer.from(parts[0], 'base64url').toString());
    console.log('📋 JWT Header:', JSON.stringify(header, null, 2));

    // 解析Payload
    const payload = JSON.parse(Buffer.from(parts[1], 'base64url').toString());
    console.log('📋 JWT Payload:', JSON.stringify(payload, null, 2));

    // 检查过期时间
    if (payload.exp) {
      const expiryDate = new Date(payload.exp * 1000);
      const now = new Date();
      console.log(`⏰ Token过期时间: ${expiryDate.toISOString()}`);
      console.log(`⏰ 当前时间: ${now.toISOString()}`);
      
      if (now >= expiryDate) {
        console.log('❌ Token已过期！');
        return false;
      } else {
        console.log('✅ Token有效期正常');
      }
    }

    console.log('✅ Token结构正常');
    return true;
  } catch (error) {
    console.log('❌ Token解析失败:', error.message);
    return false;
  }
}

// 3. 测试普通API认证
async function testAPIAuth(token) {
  console.log('\n3️⃣ 测试普通API认证...');

  return new Promise((resolve) => {
    const options = {
      hostname: 'localhost',
      port: 8080,
      path: '/api/v1/users/me',
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${token}`
      }
    };

    const req = http.request(options, (res) => {
      let body = '';
      res.on('data', (chunk) => body += chunk);
      res.on('end', () => {
        if (res.statusCode === 200) {
          console.log('✅ 普通API认证成功');
          resolve(true);
        } else {
          console.log(`❌ 普通API认证失败 (${res.statusCode}):`, body);
          resolve(false);
        }
      });
    });

    req.on('error', (error) => {
      console.log('❌ API请求失败:', error.message);
      resolve(false);
    });

    req.end();
  });
}

// 4. 测试WebSocket连接
async function testWebSocketConnection(token) {
  console.log('\n4️⃣ 测试WebSocket连接...');

  const WebSocket = require('ws');
  
  return new Promise((resolve) => {
    const wsUrl = `${WS_BASE}/api/v1/ws/connect?token=${encodeURIComponent(token)}`;
    console.log(`🔗 连接URL: ${wsUrl}`);

    const ws = new WebSocket(wsUrl);
    let connectionSuccess = false;

    ws.on('open', () => {
      console.log('✅ WebSocket连接成功');
      connectionSuccess = true;
      
      // 发送测试消息
      const testMessage = {
        id: 'test_' + Date.now(),
        type: 'HEARTBEAT',
        data: { test: true },
        timestamp: new Date().toISOString()
      };
      
      ws.send(JSON.stringify(testMessage));
    });

    ws.on('message', (data) => {
      try {
        const message = JSON.parse(data.toString());
        console.log('📨 收到消息:', message);
      } catch (error) {
        console.log('📨 收到原始消息:', data.toString());
      }
    });

    ws.on('error', (error) => {
      console.log('❌ WebSocket连接错误:', error.message);
    });

    ws.on('close', (code, reason) => {
      console.log(`🔌 WebSocket连接关闭: ${code} - ${reason}`);
      resolve(connectionSuccess);
    });

    // 5秒后关闭测试连接
    setTimeout(() => {
      if (ws.readyState === WebSocket.OPEN) {
        ws.close();
      }
      resolve(connectionSuccess);
    }, 5000);
  });
}

// 5. 测试WebSocket认证端点直接调用
async function testWebSocketAuthEndpoint(token) {
  console.log('\n5️⃣ 测试WebSocket认证端点...');

  return new Promise((resolve) => {
    const options = {
      hostname: 'localhost',
      port: 8080,
      path: `/api/v1/ws/connect?token=${encodeURIComponent(token)}`,
      method: 'GET',
      headers: {
        'Upgrade': 'websocket',
        'Connection': 'Upgrade',
        'Sec-WebSocket-Key': Buffer.from('test-key-12345678901234567890').toString('base64'),
        'Sec-WebSocket-Version': '13'
      }
    };

    const req = http.request(options, (res) => {
      console.log(`📊 认证端点响应状态: ${res.statusCode}`);
      console.log('📋 响应头:', JSON.stringify(res.headers, null, 2));
      
      let body = '';
      res.on('data', (chunk) => body += chunk);
      res.on('end', () => {
        if (body) {
          console.log('📄 响应体:', body);
        }
        resolve(res.statusCode === 101 || res.statusCode === 200);
      });
    });

    req.on('error', (error) => {
      console.log('❌ 认证端点请求失败:', error.message);
      resolve(false);
    });

    req.end();
  });
}

// 6. 生成诊断报告
function generateDiagnosticReport(results) {
  console.log('\n📊 诊断报告');
  console.log('================================');
  
  const { login, tokenValid, apiAuth, wsConnection, wsAuth } = results;
  
  if (!login) {
    console.log('🚨 严重问题：用户登录失败');
    console.log('   建议：检查用户凭据和数据库连接');
    return;
  }
  
  if (!tokenValid) {
    console.log('🚨 严重问题：Token格式或内容无效');
    console.log('   建议：检查JWT生成逻辑和密钥配置');
    return;
  }
  
  if (!apiAuth) {
    console.log('🚨 严重问题：普通API认证失败');
    console.log('   建议：检查认证中间件和JWT验证逻辑');
    return;
  }
  
  if (!wsConnection) {
    console.log('🚨 WebSocket连接问题');
    if (!wsAuth) {
      console.log('   原因：WebSocket认证端点失败');
      console.log('   建议：检查WebSocket认证中间件的JWT验证');
    } else {
      console.log('   可能原因：WebSocket升级协议问题');
    }
    return;
  }
  
  console.log('✅ 所有测试通过！WebSocket应该正常工作');
}

// 主函数
async function main() {
  try {
    // 安装websocket依赖（如果没有的话）
    try {
      require('ws');
    } catch (error) {
      console.log('⚠️  需要安装ws依赖: npm install ws');
      console.log('   跳过WebSocket连接测试...\n');
    }

    const results = {};
    
    // 1. 登录
    try {
      const token = await login();
      results.login = true;
      
      // 2. 分析Token
      results.tokenValid = analyzeToken(token);
      
      if (results.tokenValid) {
        // 3. 测试API认证
        results.apiAuth = await testAPIAuth(token);
        
        // 4. 测试WebSocket连接
        try {
          results.wsConnection = await testWebSocketConnection(token);
        } catch (error) {
          console.log('⚠️  跳过WebSocket连接测试:', error.message);
          results.wsConnection = false;
        }
        
        // 5. 测试WebSocket认证端点
        results.wsAuth = await testWebSocketAuthEndpoint(token);
      }
      
    } catch (error) {
      results.login = false;
    }
    
    // 6. 生成报告
    generateDiagnosticReport(results);
    
  } catch (error) {
    console.error('🚨 调试工具执行失败:', error);
  }
}

// 运行调试
main();