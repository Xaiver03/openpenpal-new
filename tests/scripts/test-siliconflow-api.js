#!/usr/bin/env node

const https = require('https');

// 测试直接调用SiliconFlow API
async function testSiliconFlowAPI() {
    console.log('🔧 Testing SiliconFlow API directly...\n');
    
    const API_KEY = 'sk-agfoqrwfruszwriilkyktckovqcsneieadupydihlunynlek';
    const API_ENDPOINT = 'https://api.siliconflow.cn/v1/chat/completions';
    const MODEL = 'Qwen/Qwen2.5-7B-Instruct';
    
    const requestData = {
        model: MODEL,
        messages: [
            {
                role: "system",
                content: "你是OpenPenPal的AI助手，帮助用户进行笔友匹配、生成回信、提供写作灵感和策展信件。请用温暖、友好的语气回应。"
            },
            {
                role: "user",
                content: `请生成1个写信灵感提示：

主题：日常生活
风格：温暖友好
标签：日常, 生活

每个灵感应该：
1. 提供一个具体的写作切入点
2. 激发情感共鸣
3. 适合手写信的形式
4. 50-100字的描述

返回JSON格式：
{
  "inspirations": [
    {
      "theme": "主题",
      "prompt": "写作提示",
      "style": "风格",
      "tags": ["标签1", "标签2"]
    }
  ]
}`
            }
        ],
        temperature: 0.7,
        max_tokens: 1000
    };
    
    console.log('📤 Request details:');
    console.log(`   API Endpoint: ${API_ENDPOINT}`);
    console.log(`   Model: ${MODEL}`);
    console.log(`   API Key: ${API_KEY.substring(0, 6)}...${API_KEY.substring(API_KEY.length - 4)}`);
    console.log(`   Request body size: ${JSON.stringify(requestData).length} bytes\n`);
    
    const requestBody = JSON.stringify(requestData);
    
    const options = {
        hostname: 'api.siliconflow.cn',
        path: '/v1/chat/completions',
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${API_KEY}`,
            'Content-Length': Buffer.byteLength(requestBody)
        }
    };
    
    return new Promise((resolve, reject) => {
        const req = https.request(options, (res) => {
            let data = '';
            
            res.on('data', (chunk) => {
                data += chunk;
            });
            
            res.on('end', () => {
                console.log('✅ Response received!\n');
                console.log('📥 Response status:', res.statusCode);
                console.log('📥 Response headers:', JSON.stringify(res.headers, null, 2));
                
                try {
                    const response = JSON.parse(data);
                    console.log('\n📥 Response data:');
                    console.log(JSON.stringify(response, null, 2));
                    
                    if (response.choices && response.choices[0]) {
                        const content = response.choices[0].message.content;
                        console.log('\n🎨 Generated content:');
                        console.log(content);
                        
                        // Try to parse the JSON content
                        try {
                            const parsed = JSON.parse(content);
                            console.log('\n✅ Successfully parsed JSON response:');
                            console.log(JSON.stringify(parsed, null, 2));
                        } catch (parseError) {
                            console.log('\n⚠️ Failed to parse content as JSON:', parseError.message);
                        }
                    }
                    
                    if (response.usage) {
                        console.log('\n📊 Token usage:', response.usage);
                    }
                    
                    resolve(response);
                } catch (error) {
                    console.error('❌ Failed to parse response:', error.message);
                    console.error('Raw response:', data);
                    reject(error);
                }
            });
        });
        
        req.on('error', (error) => {
            console.error('❌ Request error:', error.message);
            reject(error);
        });
        
        req.write(requestBody);
        req.end();
    });
}

// 测试通过后端API调用
async function testBackendAPI(token) {
    console.log('\n\n🔧 Testing Backend AI Inspiration API...\n');
    
    const requestData = {
        theme: '思念',
        style: 'romantic'
    };
    
    console.log('📤 Request details:');
    console.log(`   Backend endpoint: http://localhost:8080/api/v1/ai/inspiration`);
    console.log(`   Request data:`, requestData);
    console.log(`   Token: ${token ? token.substring(0, 20) + '...' : 'Not provided'}\n`);
    
    if (!token) {
        console.log('⚠️  No JWT token provided. Skipping backend test.');
        console.log('To test backend API:');
        console.log('1. Login to the application');
        console.log('2. Get the JWT token from localStorage or browser cookies');
        console.log('3. Run: TEST_JWT_TOKEN="your-token-here" node test-siliconflow-api.js\n');
        return;
    }
    
    const http = require('http');
    const requestBody = JSON.stringify(requestData);
    
    const options = {
        hostname: 'localhost',
        port: 8080,
        path: '/api/v1/ai/inspiration',
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${token}`,
            'Content-Length': Buffer.byteLength(requestBody)
        }
    };
    
    return new Promise((resolve, reject) => {
        const req = http.request(options, (res) => {
            let data = '';
            
            res.on('data', (chunk) => {
                data += chunk;
            });
            
            res.on('end', () => {
                console.log('📥 Response status:', res.statusCode);
                
                try {
                    const response = JSON.parse(data);
                    
                    if (res.statusCode === 200) {
                        console.log('✅ Backend API Call Successful!\n');
                        console.log('📝 Response:');
                        console.log(JSON.stringify(response, null, 2));
                    } else {
                        console.log('❌ Backend API Call Failed!\n');
                        console.log('Error Response:', JSON.stringify(response, null, 2));
                    }
                    
                    resolve(response);
                } catch (error) {
                    console.log('❌ Failed to parse response:', error.message);
                    console.log('Raw response:', data);
                    reject(error);
                }
            });
        });
        
        req.on('error', (error) => {
            console.error('❌ Request Error:', error.message);
            reject(error);
        });
        
        req.write(requestBody);
        req.end();
    });
}

// 运行测试
async function runTests() {
    console.log('=== SiliconFlow API Test Tool ===\n');
    console.log('This tool tests both direct SiliconFlow API calls and backend integration.\n');
    
    try {
        // 先测试直接调用SiliconFlow API
        await testSiliconFlowAPI();
        
        // 再测试通过后端调用
        const token = process.env.TEST_JWT_TOKEN;
        await testBackendAPI(token);
        
    } catch (error) {
        console.error('Test failed:', error);
    }
}

// Run the test
runTests();