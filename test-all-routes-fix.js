#!/usr/bin/env node

const fetch = require('node-fetch');

async function testAllRoutes() {
    console.log('🔍 Testing all routes through Gateway...\n');

    const gatewayUrl = 'http://localhost:8000';
    
    // Test public endpoints (no auth needed)
    const publicEndpoints = [
        '/api/v1/letters/public',
        '/api/v1/letters/popular',
        '/api/v1/letters/recommended',
        '/api/v1/letters/templates',
        '/api/v1/museum/entries',
        '/api/v1/museum/exhibitions',
        '/api/v1/museum/stats',
        '/api/v1/ai/daily-inspiration',
        '/api/v1/ai/stats'
    ];
    
    console.log('1. Testing public endpoints (no auth)...');
    for (const endpoint of publicEndpoints) {
        try {
            const res = await fetch(`${gatewayUrl}${endpoint}`, {
                headers: { 'Accept': 'application/json' }
            });
            console.log(`${endpoint}: ${res.status} ${res.status === 404 ? '❌' : '✅'}`);
        } catch (error) {
            console.log(`${endpoint}: ❌ Error - ${error.message}`);
        }
    }
    
    // Login to test authenticated endpoints
    console.log('\n2. Logging in...');
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
    console.log('✅ Login successful\n');
    
    // Test authenticated endpoints
    console.log('3. Testing authenticated endpoints...');
    const authEndpoints = [
        { method: 'POST', path: '/api/v1/ai/inspiration', body: { theme: '友谊', count: 1 } },
        { method: 'GET', path: '/api/v1/letters' },
        { method: 'GET', path: '/api/v1/courier/status' }
    ];
    
    for (const endpoint of authEndpoints) {
        try {
            const options = {
                method: endpoint.method,
                headers: { 
                    'Authorization': `Bearer ${token}`,
                    'Accept': 'application/json',
                    'Content-Type': 'application/json'
                }
            };
            
            if (endpoint.body) {
                options.body = JSON.stringify(endpoint.body);
            }
            
            const res = await fetch(`${gatewayUrl}${endpoint.path}`, options);
            console.log(`${endpoint.method} ${endpoint.path}: ${res.status} ${res.status === 404 ? '❌' : '✅'}`);
        } catch (error) {
            console.log(`${endpoint.method} ${endpoint.path}: ❌ Error - ${error.message}`);
        }
    }
    
    console.log('\n✅ Route testing complete!');
}

testAllRoutes().catch(console.error);