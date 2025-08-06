#!/usr/bin/env node

// 简单的Moonshot API测试脚本
const https = require('https');

const API_KEY = 'sk-wQUdnk3pwUdEkAGQl85krQcmAE36eLl8DuKj1hLZZrjzuxvV';
const API_URL = 'https://api.moonshot.cn/v1/chat/completions';

const testData = {
  model: 'moonshot-v1-8k',
  messages: [
    {
      role: 'system',
      content: '你是OpenPenPal的AI助手，在这个温暖的数字书信平台上，帮助用户进行笔友匹配、生成回信、提供写作灵感和策展信件。请用温暖、友好、富有人文情怀的语气回应。'
    },
    {
      role: 'user',
      content: '请为"秋日校园"这个主题提供一个温暖风格的写作灵感，包含友情和回忆的元素。'
    }
  ],
  temperature: 0.7,
  max_tokens: 500
};

const requestData = JSON.stringify(testData);

const options = {
  hostname: 'api.moonshot.cn',
  port: 443,
  path: '/v1/chat/completions',
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${API_KEY}`,
    'Content-Length': Buffer.byteLength(requestData)
  }
};

console.log('🌙 [Moonshot API Test] 开始测试Moonshot Kimi API连接...');
console.log('🌙 [Moonshot API Test] API端点:', API_URL);
console.log('🌙 [Moonshot API Test] 模型:', testData.model);

const req = https.request(options, (res) => {
  console.log(`🌙 [Moonshot API Test] 状态码: ${res.statusCode}`);
  console.log(`🌙 [Moonshot API Test] 响应头:`, res.headers);

  let data = '';

  res.on('data', (chunk) => {
    data += chunk;
  });

  res.on('end', () => {
    try {
      const response = JSON.parse(data);
      
      if (res.statusCode === 200) {
        console.log('✅ [Moonshot API Test] API调用成功!');
        console.log('🌙 [Response] 完整响应:', JSON.stringify(response, null, 2));
        
        if (response.choices && response.choices[0] && response.choices[0].message) {
          console.log('💬 [AI回复内容]:');
          console.log(response.choices[0].message.content);
        }
        
        if (response.usage) {
          console.log('📊 [Token使用情况]:');
          console.log(`   - 输入tokens: ${response.usage.prompt_tokens}`);
          console.log(`   - 输出tokens: ${response.usage.completion_tokens}`);
          console.log(`   - 总tokens: ${response.usage.total_tokens}`);
        }
      } else {
        console.log('❌ [Moonshot API Test] API调用失败');
        console.log('错误响应:', JSON.stringify(response, null, 2));
      }
    } catch (error) {
      console.log('❌ [Moonshot API Test] 解析响应失败:', error.message);
      console.log('原始响应:', data);
    }
  });
});

req.on('error', (error) => {
  console.log('❌ [Moonshot API Test] 请求失败:', error.message);
});

// 发送请求
req.write(requestData);
req.end();

console.log('🚀 [Moonshot API Test] 请求已发送，等待响应...');