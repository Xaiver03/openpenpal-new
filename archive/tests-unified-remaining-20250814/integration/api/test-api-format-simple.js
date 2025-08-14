const axios = require('axios');

const API_URL = 'http://localhost:8080/api/v1';

async function testApiFormat() {
    console.log('🔍 Testing API response format\n');
    
    try {
        // 1. Test health endpoint
        console.log('1. Testing health endpoint...');
        const healthRes = await axios.get(`${API_URL}/health`);
        console.log('Health response:', JSON.stringify(healthRes.data, null, 2));
        
        // 2. Test login with wrong credentials to see error format
        console.log('\n2. Testing login error response format...');
        try {
            await axios.post(`${API_URL}/auth/login`, {
                username: 'test',
                password: 'test'
            });
        } catch (error) {
            if (error.response) {
                console.log('Login error response format:', JSON.stringify(error.response.data, null, 2));
                
                // Check if response contains snake_case or camelCase
                const responseStr = JSON.stringify(error.response.data);
                const hasSnakeCase = responseStr.includes('_');
                const hasCamelCase = /[a-z][A-Z]/.test(responseStr);
                
                console.log('\nField format analysis:');
                console.log('- Contains snake_case fields:', hasSnakeCase);
                console.log('- Contains camelCase fields:', hasCamelCase);
            }
        }
        
        // 3. Test register endpoint to see validation response
        console.log('\n3. Testing register validation response format...');
        try {
            await axios.post(`${API_URL}/auth/register`, {
                username: 'a',  // Too short
                password: '123' // Too short
            });
        } catch (error) {
            if (error.response) {
                console.log('Register validation response:', JSON.stringify(error.response.data, null, 2));
            }
        }
        
    } catch (error) {
        console.error('Unexpected error:', error.message);
    }
}

testApiFormat();