const axios = require('axios');

const BASE_URL = 'http://localhost:8080';

async function testApiFormat() {
    console.log('🔍 Testing API response format after middleware change\n');
    
    try {
        // 1. Test health endpoint
        console.log('1. Testing health endpoint...');
        const healthRes = await axios.get(`${BASE_URL}/health`);
        console.log('Health response:', JSON.stringify(healthRes.data, null, 2));
        
        // 2. Test login endpoint format
        console.log('\n2. Testing login API response format...');
        try {
            const loginRes = await axios.post(`${BASE_URL}/api/v1/auth/login`, {
                username: 'nonexistent',
                password: 'wrongpass'
            });
        } catch (error) {
            if (error.response) {
                console.log('Login error response:', JSON.stringify(error.response.data, null, 2));
                analyzeFormat(error.response.data, 'Login error');
            }
        }
        
        // 3. Test register validation
        console.log('\n3. Testing register API response format...');
        try {
            const registerRes = await axios.post(`${BASE_URL}/api/v1/auth/register`, {
                username: 'te',  // Too short
                password: '123', // Too short
                email: 'invalid' // Invalid email
            });
        } catch (error) {
            if (error.response) {
                console.log('Register validation response:', JSON.stringify(error.response.data, null, 2));
                analyzeFormat(error.response.data, 'Register validation');
            }
        }
        
        // 4. Test a successful endpoint if possible
        console.log('\n4. Testing successful response format with test account...');
        try {
            // Try with the test account from CLAUDE.md
            const loginRes = await axios.post(`${BASE_URL}/api/v1/auth/login`, {
                username: 'admin',
                password: 'admin123'
            });
            console.log('Login success response:', JSON.stringify(loginRes.data, null, 2));
            analyzeFormat(loginRes.data, 'Login success');
            
            // If login succeeds, test an authenticated endpoint
            if (loginRes.data.data && loginRes.data.data.token) {
                const token = loginRes.data.data.token;
                const userRes = await axios.get(`${BASE_URL}/api/v1/users/me`, {
                    headers: { Authorization: `Bearer ${token}` }
                });
                console.log('\n5. User info response:', JSON.stringify(userRes.data, null, 2));
                analyzeFormat(userRes.data, 'User info');
            }
        } catch (error) {
            console.log('Test account not available:', error.response?.data?.error || error.message);
        }
        
    } catch (error) {
        console.error('Unexpected error:', error.message);
    }
}

function analyzeFormat(data, context) {
    const jsonStr = JSON.stringify(data);
    const snakeCaseFields = jsonStr.match(/\"[a-z]+(_[a-z]+)+\":/g) || [];
    const camelCaseFields = jsonStr.match(/\"[a-z]+[A-Z][a-zA-Z]*\":/g) || [];
    
    console.log(`\n${context} field format analysis:`);
    console.log(`- Snake_case fields found: ${snakeCaseFields.length}`, snakeCaseFields.slice(0, 5).join(', '));
    console.log(`- CamelCase fields found: ${camelCaseFields.length}`, camelCaseFields.slice(0, 5).join(', '));
    console.log(`- Format: ${snakeCaseFields.length > 0 ? 'snake_case' : 'camelCase'} dominant\n`);
}

testApiFormat();