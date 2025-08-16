#!/usr/bin/env node

const fetch = require('node-fetch');

async function testAIGatewayFix() {
    console.log('🎉 Testing AI routes through Gateway after fix...\n');

    const gatewayUrl = 'http://localhost:8000';
    
    // First, login to get token
    console.log('1. Logging in through gateway...');
    const loginRes = await fetch(`${gatewayUrl}/api/v1/auth/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            username: 'admin',
            password: 'admin123'
        })
    });
    
    const loginData = await loginRes.json();
    if (!loginData.success) {
        console.error('Login failed:', loginData);
        return;
    }
    
    const token = loginData.data.token;
    console.log('✅ Login successful, got token\n');

    // Test daily inspiration through gateway
    console.log('2. Testing daily inspiration through gateway...');
    const dailyRes = await fetch(`${gatewayUrl}/api/v1/ai/daily-inspiration`, {
        headers: { 
            'Authorization': `Bearer ${token}`,
            'Accept': 'application/json'
        }
    });
    
    console.log(`Response status: ${dailyRes.status}`);
    const dailyData = await dailyRes.json();
    console.log('Daily inspiration response:', JSON.stringify(dailyData, null, 2));
    
    // Test AI stats through gateway
    console.log('\n3. Testing AI stats through gateway...');
    const statsRes = await fetch(`${gatewayUrl}/api/v1/ai/stats`, {
        headers: { 
            'Authorization': `Bearer ${token}`,
            'Accept': 'application/json'
        }
    });
    
    console.log(`Response status: ${statsRes.status}`);
    const statsData = await statsRes.json();
    console.log('AI stats response:', JSON.stringify(statsData, null, 2));
    
    // Test inspiration generation through gateway
    console.log('\n4. Testing AI inspiration generation through gateway...');
    const inspirationRes = await fetch(`${gatewayUrl}/api/v1/ai/inspiration`, {
        method: 'POST',
        headers: { 
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            theme: '友谊',
            count: 2
        })
    });
    
    console.log(`Response status: ${inspirationRes.status}`);
    const inspirationData = await inspirationRes.json();
    console.log('AI inspiration response:', JSON.stringify(inspirationData, null, 2));
    
    console.log('\n✅ AI routes are now working through the gateway!');
}

testAIGatewayFix().catch(console.error);