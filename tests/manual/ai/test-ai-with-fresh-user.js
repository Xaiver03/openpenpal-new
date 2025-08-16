#!/usr/bin/env node

const axios = require('axios');

// Test AI inspiration with a fresh user (alice)
async function testInspirationWithFreshUser() {
    console.log('🧪 Testing OpenPenPal Inspiration API with fresh user...\n');
    
    // Login with alice account
    try {
        console.log('🔐 Logging in with alice account...');
        const loginResponse = await axios.post('http://localhost:8080/api/v1/auth/login', {
            username: 'alice',
            password: 'secret'
        });
        
        const token = loginResponse.data.data.token;
        console.log('✅ Login successful! Token:', token.substring(0, 20) + '...\n');
        
        // Test inspiration API
        console.log('🎨 Testing /api/v1/ai/inspiration endpoint...');
        
        const requestData = {
            theme: '日常生活',
            style: '温暖友好',
            tags: ['日常', '生活'],
            count: 1
        };
        
        console.log('📤 Request body:', JSON.stringify(requestData, null, 2));
        
        const response = await axios.post('http://localhost:8080/api/v1/ai/inspiration', requestData, {
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json'
            }
        });
        
        console.log('\n✅ API call successful!');
        console.log('📥 Response status:', response.status);
        console.log('📥 Response data:', JSON.stringify(response.data, null, 2));
        
        // Parse the inspiration if it's in the data
        if (response.data.data && response.data.data.inspirations) {
            console.log('\n🎨 Generated Inspirations:');
            response.data.data.inspirations.forEach((insp, idx) => {
                console.log(`\n${idx + 1}. ${insp.theme}`);
                console.log(`   📝 ${insp.prompt}`);
                console.log(`   🎯 Style: ${insp.style}`);
                console.log(`   🏷️ Tags: ${insp.tags.join(', ')}`);
            });
        }
        
    } catch (error) {
        console.error('\n❌ Test failed!');
        if (error.response) {
            console.log('📥 Error status:', error.response.status);
            console.log('📥 Error data:', JSON.stringify(error.response.data, null, 2));
        } else {
            console.error('Error:', error.message);
        }
    }
}

// Direct API test to verify SiliconFlow is working
async function testDirectSiliconFlowCall() {
    console.log('\n\n🔧 Testing direct SiliconFlow API call...\n');
    
    const https = require('https');
    const API_KEY = 'sk-agfoqrwfruszwriilkyktckovqcsneieadupydihlunynlek';
    
    const requestData = {
        model: 'Qwen/Qwen2.5-7B-Instruct',
        messages: [
            {
                role: "system",
                content: "你是一个创意写作助手"
            },
            {
                role: "user",
                content: "给我一个写信的灵感，主题是日常生活"
            }
        ],
        temperature: 0.7,
        max_tokens: 500
    };
    
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
    
    return new Promise((resolve) => {
        const req = https.request(options, (res) => {
            let data = '';
            
            res.on('data', (chunk) => {
                data += chunk;
            });
            
            res.on('end', () => {
                try {
                    const response = JSON.parse(data);
                    if (res.statusCode === 200) {
                        console.log('✅ Direct SiliconFlow API call successful!');
                        console.log('📥 Response:', response.choices[0].message.content.substring(0, 100) + '...');
                    } else {
                        console.log('❌ Direct API call failed!');
                        console.log('Status:', res.statusCode);
                        console.log('Response:', data);
                    }
                } catch (error) {
                    console.error('Failed to parse response:', error.message);
                }
                resolve();
            });
        });
        
        req.on('error', (error) => {
            console.error('❌ Request error:', error.message);
            resolve();
        });
        
        req.write(requestBody);
        req.end();
    });
}

// Run both tests
async function runTests() {
    // First test with fresh user
    await testInspirationWithFreshUser();
    
    // Then test direct API to confirm SiliconFlow is working
    await testDirectSiliconFlowCall();
}

runTests();