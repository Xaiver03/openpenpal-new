#!/usr/bin/env node

const fetch = require('node-fetch');

async function testAIMoonshot() {
    console.log('🚀 Testing AI Moonshot API...\n');

    const baseUrl = 'http://localhost:8080';
    
    // First, login to get token
    console.log('1. Logging in as admin...');
    const loginRes = await fetch(`${baseUrl}/api/v1/auth/login`, {
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
    console.log('✅ Login successful\n');

    // Test daily inspiration (GET)
    console.log('2. Testing daily inspiration endpoint...');
    const dailyRes = await fetch(`${baseUrl}/api/v1/ai/daily-inspiration`, {
        headers: { 
            'Authorization': `Bearer ${token}`,
            'Accept': 'application/json'
        }
    });
    
    const dailyData = await dailyRes.json();
    console.log('Daily inspiration response:', JSON.stringify(dailyData, null, 2));
    console.log('\n');

    // Test inspiration generation (POST)
    console.log('3. Testing AI inspiration generation with Moonshot...');
    const inspirationRes = await fetch(`${baseUrl}/api/v1/ai/inspiration`, {
        method: 'POST',
        headers: { 
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            theme: '友谊',
            count: 3,
            style: '温暖'
        })
    });
    
    const inspirationData = await inspirationRes.json();
    console.log('AI inspiration response:', JSON.stringify(inspirationData, null, 2));
    
    if (inspirationData.success && inspirationData.data.inspirations) {
        console.log(`\n✅ Generated ${inspirationData.data.inspirations.length} inspirations`);
        inspirationData.data.inspirations.forEach((insp, i) => {
            console.log(`\nInspiration ${i + 1}:`);
            console.log(`  Theme: ${insp.theme}`);
            console.log(`  Prompt: ${insp.prompt}`);
            console.log(`  Style: ${insp.style}`);
            console.log(`  Tags: ${insp.tags?.join(', ') || 'none'}`);
        });
    }

    // Test AI stats
    console.log('\n4. Testing AI stats endpoint...');
    const statsRes = await fetch(`${baseUrl}/api/v1/ai/stats`, {
        headers: { 
            'Authorization': `Bearer ${token}`,
            'Accept': 'application/json'
        }
    });
    
    const statsData = await statsRes.json();
    console.log('AI stats response:', JSON.stringify(statsData, null, 2));
}

testAIMoonshot().catch(console.error);