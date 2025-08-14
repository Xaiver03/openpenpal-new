#!/usr/bin/env node

const axios = require('axios');

async function testInspirationAPI() {
    console.log('🧪 Testing OpenPenPal Inspiration API...\n');
    
    // First, let's login to get a token
    try {
        console.log('🔐 Logging in...');
        const loginResponse = await axios.post('http://localhost:8080/api/v1/auth/login', {
            username: 'admin',
            password: 'admin123'
        });
        
        const token = loginResponse.data.data.token;
        console.log('✅ Login successful! Token:', token.substring(0, 20) + '...\n');
        
        // Now test the inspiration API
        console.log('🎨 Testing /api/v1/ai/inspiration endpoint...');
        
        const inspirationRequest = {
            theme: "日常生活",
            style: "温暖友好",
            tags: ["日常", "生活"],
            count: 1
        };
        
        console.log('📤 Request body:', JSON.stringify(inspirationRequest, null, 2));
        
        const response = await axios.post('http://localhost:8080/api/v1/ai/inspiration', 
            inspirationRequest,
            {
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json'
                }
            }
        );
        
        console.log('\n✅ Response received!');
        console.log('📥 Status:', response.status);
        console.log('📥 Response data:', JSON.stringify(response.data, null, 2));
        
    } catch (error) {
        console.error('\n❌ API call failed!');
        
        if (error.response) {
            console.error('📥 Error status:', error.response.status);
            console.error('📥 Error data:', JSON.stringify(error.response.data, null, 2));
        } else if (error.request) {
            console.error('🚫 No response received from server');
            console.error('Is the backend running on port 8080?');
        } else {
            console.error('🚨 Error:', error.message);
        }
    }
}

// Wait a bit for backend to start
console.log('⏳ Waiting 3 seconds for backend to start...\n');
setTimeout(() => {
    testInspirationAPI();
}, 3000);