#!/usr/bin/env node

const axios = require('axios');

async function testFrontendAPI() {
    console.log('🧪 Testing Frontend AI API...\n');
    
    try {
        // First login to get a token
        console.log('🔐 Logging in via backend...');
        const loginResponse = await axios.post('http://localhost:8080/api/v1/auth/login', {
            username: 'admin',
            password: 'admin123'
        });
        
        const token = loginResponse.data.data.token;
        console.log('✅ Login successful!\n');
        
        // Test backend API directly
        console.log('🎯 Testing backend /api/v1/ai/inspiration directly...');
        const backendResponse = await axios.post(
            'http://localhost:8080/api/v1/ai/inspiration',
            {
                theme: "日常生活",
                style: "温暖友好", 
                tags: ["日常", "生活"],
                count: 1
            },
            {
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json'
                }
            }
        );
        
        console.log('✅ Backend response:', JSON.stringify(backendResponse.data, null, 2));
        
        // Now test via frontend proxy (if it exists)
        console.log('\n🎯 Testing frontend /api/ai/inspiration...');
        try {
            const frontendResponse = await axios.post(
                'http://localhost:3000/api/ai/inspiration',
                {
                    theme: "日常生活",
                    style: "温暖友好",
                    tags: ["日常", "生活"], 
                    count: 1
                },
                {
                    headers: {
                        'Cookie': `openpenpal_auth_token=${token}`,
                        'Content-Type': 'application/json'
                    }
                }
            );
            
            console.log('✅ Frontend proxy response:', JSON.stringify(frontendResponse.data, null, 2));
        } catch (frontendError) {
            console.log('❌ Frontend proxy error (expected if no proxy route exists)');
            console.log('   This is OK - the frontend should call backend directly');
        }
        
    } catch (error) {
        console.error('\n❌ Test failed!');
        
        if (error.response) {
            console.error('📥 Error status:', error.response.status);
            console.error('📥 Error data:', JSON.stringify(error.response.data, null, 2));
        } else if (error.request) {
            console.error('🚫 No response received');
        } else {
            console.error('🚨 Error:', error.message);
        }
    }
}

testFrontendAPI();