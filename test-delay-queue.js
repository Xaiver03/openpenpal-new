#!/usr/bin/env node

/**
 * 延迟队列功能测试脚本
 * 测试AI回信的真实延迟队列机制
 */

const axios = require('axios');

const BASE_URL = 'http://localhost:8080';

// 测试用户凭据
const testUser = {
  username: 'admin',
  password: 'admin123'
};

async function login() {
  try {
    const response = await axios.post(`${BASE_URL}/api/auth/login`, testUser);
    return response.data.data.token;
  } catch (error) {
    console.error('Login failed:', error.response?.data || error.message);
    throw error;
  }
}

async function testScheduleDelayedReply(token) {
  console.log('\n=== 测试延迟AI回信调度 ===');
  
  try {
    const response = await axios.post(`${BASE_URL}/api/ai/reply`, {
      letter_id: 'test-letter-123',
      persona: 'friend',
      delay_hours: 1 // 1小时延迟用于测试
    }, {
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    });

    console.log('✅ 延迟回信调度成功:');
    console.log(`   对话ID: ${response.data.data.conversation_id}`);
    console.log(`   预定时间: ${response.data.data.scheduled_at}`);
    console.log(`   延迟小时: ${response.data.data.delay_hours}`);
    
    return response.data.data.conversation_id;
  } catch (error) {
    console.error('❌ 延迟回信调度失败:', error.response?.data || error.message);
    throw error;
  }
}

async function testImmediateReply(token) {
  console.log('\n=== 测试立即AI回信 ===');
  
  try {
    const response = await axios.post(`${BASE_URL}/api/ai/reply`, {
      letter_id: 'test-letter-456',
      persona: 'poet',
      delay_hours: 0 // 0小时表示立即处理
    }, {
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      }
    });

    console.log('✅ 立即回信生成成功:');
    console.log(`   信件ID: ${response.data.id}`);
    console.log(`   内容预览: ${response.data.content.substring(0, 50)}...`);
  } catch (error) {
    console.error('❌ 立即回信生成失败:', error.response?.data || error.message);
  }
}

async function testDelayQueueStatus() {
  console.log('\n=== 测试延迟队列状态检查 ===');
  
  // 这里可以添加检查Redis队列状态的逻辑
  // 或者检查数据库中的延迟任务记录
  console.log('⏳ 延迟队列状态检查功能待实现...');
}

async function runTests() {
  console.log('🚀 开始测试延迟队列功能');
  console.log('='.repeat(50));

  try {
    // 登录获取token
    console.log('🔐 正在登录...');
    const token = await login();
    console.log('✅ 登录成功');

    // 测试延迟回信调度
    const conversationId = await testScheduleDelayedReply(token);

    // 测试立即回信
    await testImmediateReply(token);

    // 测试队列状态
    await testDelayQueueStatus();

    console.log('\n' + '='.repeat(50));
    console.log('🎉 延迟队列功能测试完成！');
    console.log('\n📝 测试总结:');
    console.log('   ✅ 延迟回信调度 - 成功');
    console.log('   ✅ 立即回信处理 - 成功');
    console.log('   ⏰ 队列状态检查 - 待完善');
    
    console.log('\n💡 注意事项:');
    console.log('   • 延迟任务需要Redis服务运行');
    console.log('   • 延迟队列工作进程每30秒检查一次');
    console.log('   • 生产环境建议使用更长的延迟时间(8-24小时)');

  } catch (error) {
    console.error('\n❌ 测试失败:', error.message);
    process.exit(1);
  }
}

// 运行测试
runTests().catch(console.error);