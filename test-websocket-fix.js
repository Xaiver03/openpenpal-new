#!/usr/bin/env node

/**
 * WebSocket修复验证工具
 * 验证WebSocket连接问题是否已修复
 */

const WebSocket = require('ws');
const http = require('http');

const API_BASE = 'http://localhost:8080';
const WS_BASE = 'ws://localhost:8080';

console.log('🔧 验证WebSocket修复效果');
console.log('================================\n');

// 1. 获取token
async function getToken() {
  console.log('1️⃣ 获取认证token...');
  
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
            console.log('✅ 获取token成功');
            resolve(response.data.token);
          } else {
            reject(new Error('Login failed'));
          }
        } catch (error) {
          reject(error);
        }
      });
    });

    req.on('error', reject);
    req.write(data);
    req.end();
  });
}

// 2. 测试WebSocket连接稳定性
async function testWebSocketStability(token) {
  console.log('2️⃣ 测试WebSocket连接稳定性...');

  return new Promise((resolve) => {
    // 使用修复后的标准路径
    const wsUrl = `${WS_BASE}/api/v1/ws/connect?token=${encodeURIComponent(token)}`;
    console.log(`🔗 连接URL: ${wsUrl}`);

    const ws = new WebSocket(wsUrl);
    let messageCount = 0;
    let connectionStable = false;

    const testTimer = setTimeout(() => {
      console.log(`📊 测试结果: 收到 ${messageCount} 条消息`);
      if (messageCount >= 2) {
        console.log('✅ WebSocket连接稳定');
        connectionStable = true;
      } else {
        console.log('❌ WebSocket连接不稳定');
      }
      
      ws.close();
      resolve(connectionStable);
    }, 3000); // 测试3秒

    ws.on('open', () => {
      console.log('🔌 WebSocket连接已建立');
      
      // 发送心跳测试
      const heartbeat = {
        id: 'test_heartbeat',
        type: 'HEARTBEAT',
        data: { test: true },
        timestamp: new Date().toISOString()
      };
      ws.send(JSON.stringify(heartbeat));
    });

    ws.on('message', (data) => {
      try {
        const message = JSON.parse(data.toString());
        messageCount++;
        console.log(`📨 收到消息 #${messageCount}: ${message.type}`);
      } catch (error) {
        console.log(`📨 收到原始消息: ${data.toString()}`);
      }
    });

    ws.on('error', (error) => {
      console.log('❌ WebSocket错误:', error.message);
      clearTimeout(testTimer);
      resolve(false);
    });

    ws.on('close', (code, reason) => {
      console.log(`🔌 WebSocket连接关闭: ${code} - ${reason || 'No reason'}`);
    });
  });
}

// 3. 测试多连接并发
async function testConcurrentConnections(token) {
  console.log('\n3️⃣ 测试并发连接...');

  const promises = [];
  const connectionCount = 3;

  for (let i = 0; i < connectionCount; i++) {
    const promise = new Promise((resolve) => {
      const wsUrl = `${WS_BASE}/api/v1/ws/connect?token=${encodeURIComponent(token)}`;
      const ws = new WebSocket(wsUrl);
      let connected = false;

      const timeout = setTimeout(() => {
        if (!connected) {
          console.log(`❌ 连接 #${i + 1} 超时`);
          ws.close();
          resolve(false);
        }
      }, 2000);

      ws.on('open', () => {
        connected = true;
        console.log(`✅ 连接 #${i + 1} 成功`);
        clearTimeout(timeout);
        
        setTimeout(() => {
          ws.close();
          resolve(true);
        }, 1000);
      });

      ws.on('error', (error) => {
        console.log(`❌ 连接 #${i + 1} 错误:`, error.message);
        clearTimeout(timeout);
        resolve(false);
      });
    });

    promises.push(promise);
  }

  const results = await Promise.all(promises);
  const successCount = results.filter(r => r).length;
  
  console.log(`📊 并发测试结果: ${successCount}/${connectionCount} 连接成功`);
  return successCount === connectionCount;
}

// 4. 生成修复报告
function generateFixReport(results) {
  console.log('\n📋 WebSocket修复报告');
  console.log('================================');
  
  const { stability, concurrent } = results;
  
  if (stability && concurrent) {
    console.log('🎉 WebSocket修复成功！');
    console.log('✅ 连接稳定性: 正常');
    console.log('✅ 并发连接: 正常');
    console.log('\n修复内容:');
    console.log('- 统一WebSocket URL路径为 /api/v1/ws/connect');
    console.log('- 优化认证状态变化处理');
    console.log('- 添加防抖机制防止意外登出');
    console.log('- 改进错误处理逻辑');
  } else {
    console.log('⚠️  WebSocket可能仍有问题');
    if (!stability) {
      console.log('❌ 连接稳定性: 异常');
    }
    if (!concurrent) {
      console.log('❌ 并发连接: 异常');
    }
  }
}

// 主函数
async function main() {
  try {
    const token = await getToken();
    
    const results = {
      stability: await testWebSocketStability(token),
      concurrent: await testConcurrentConnections(token)
    };
    
    generateFixReport(results);
    
  } catch (error) {
    console.error('🚨 测试失败:', error.message);
  }
}

main();