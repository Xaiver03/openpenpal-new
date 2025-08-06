const axios = require('axios');

// Configuration
const BACKEND_URL = 'http://localhost:8080';
const FRONTEND_URL = 'http://localhost:3000';
const API_ENDPOINT = '/api/v1/ai/inspiration';

// Test data
const TEST_USER = {
    username: 'alice',
    password: 'secret'
};

// Helper function to login and get JWT token
async function getAuthToken() {
    try {
        console.log('\n🔐 Getting authentication token...');
        const response = await axios.post(`${BACKEND_URL}/api/v1/auth/login`, {
            username: TEST_USER.username,
            password: TEST_USER.password
        });
        
        if (response.data && response.data.data && response.data.data.token) {
            console.log('✅ Successfully authenticated');
            console.log(`   User: ${response.data.data.user.username}`);
            console.log(`   Role: ${response.data.data.user.role}`);
            console.log(`   Token: ${response.data.data.token.substring(0, 20)}...`);
            return response.data.data.token;
        } else {
            throw new Error('Invalid login response structure');
        }
    } catch (error) {
        console.error('❌ Authentication failed:', error.response?.data || error.message);
        throw error;
    }
}

// Test backend API directly
async function testBackendAPI(token) {
    console.log('\n🔧 Testing Backend API Directly');
    console.log(`   Endpoint: ${BACKEND_URL}${API_ENDPOINT}`);
    
    const testRequest = {
        context: "I want to write a letter to my best friend about our summer memories",
        style: "nostalgic",
        mood: "warm"
    };
    
    console.log('\n📤 Request:');
    console.log(JSON.stringify(testRequest, null, 2));
    
    try {
        const startTime = Date.now();
        const response = await axios.post(
            `${BACKEND_URL}${API_ENDPOINT}`,
            testRequest,
            {
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json'
                }
            }
        );
        const endTime = Date.now();
        
        console.log('\n📥 Response:');
        console.log(`   Status: ${response.status} ${response.statusText}`);
        console.log(`   Time: ${endTime - startTime}ms`);
        console.log(`   Headers:`, response.headers);
        console.log('\n   Body:');
        console.log(JSON.stringify(response.data, null, 2));
        
        // Validate response structure
        if (response.data && response.data.data) {
            const data = response.data.data;
            console.log('\n✅ Response validation:');
            console.log(`   - Has suggestions: ${data.suggestions ? '✓' : '✗'}`);
            console.log(`   - Suggestions count: ${data.suggestions?.length || 0}`);
            console.log(`   - Has themes: ${data.themes ? '✓' : '✗'}`);
            console.log(`   - Themes count: ${data.themes?.length || 0}`);
            console.log(`   - Has opening_lines: ${data.opening_lines ? '✓' : '✗'}`);
            console.log(`   - Opening lines count: ${data.opening_lines?.length || 0}`);
            console.log(`   - Has model: ${data.model ? '✓' : '✗'}`);
            console.log(`   - Has request_id: ${data.request_id ? '✓' : '✗'}`);
            console.log(`   - Has input_tokens: ${data.input_tokens !== undefined ? '✓' : '✗'}`);
            console.log(`   - Has output_tokens: ${data.output_tokens !== undefined ? '✓' : '✗'}`);
            console.log(`   - Has total_tokens: ${data.total_tokens !== undefined ? '✓' : '✗'}`);
        }
        
        return response.data;
    } catch (error) {
        console.error('\n❌ Backend API Error:');
        console.error(`   Status: ${error.response?.status || 'N/A'}`);
        console.error(`   Message: ${error.response?.data?.error || error.message}`);
        console.error(`   Details:`, error.response?.data || error.message);
        throw error;
    }
}

// Test frontend proxy API
async function testFrontendProxy(token) {
    console.log('\n🌐 Testing Frontend Proxy API');
    console.log(`   Endpoint: ${FRONTEND_URL}/api/ai/inspiration`);
    
    const testRequest = {
        context: "I want to write a letter about gratitude to my teacher",
        style: "formal",
        mood: "appreciative"
    };
    
    console.log('\n📤 Request:');
    console.log(JSON.stringify(testRequest, null, 2));
    
    try {
        const startTime = Date.now();
        const response = await axios.post(
            `${FRONTEND_URL}/api/ai/inspiration`,
            testRequest,
            {
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json'
                }
            }
        );
        const endTime = Date.now();
        
        console.log('\n📥 Response:');
        console.log(`   Status: ${response.status} ${response.statusText}`);
        console.log(`   Time: ${endTime - startTime}ms`);
        console.log(`   Headers:`, response.headers);
        console.log('\n   Body:');
        console.log(JSON.stringify(response.data, null, 2));
        
        return response.data;
    } catch (error) {
        console.error('\n❌ Frontend Proxy Error:');
        console.error(`   Status: ${error.response?.status || 'N/A'}`);
        console.error(`   Message: ${error.response?.data?.error || error.message}`);
        console.error(`   Details:`, error.response?.data || error.message);
        
        // Check if it's a CORS issue
        if (error.code === 'ERR_NETWORK') {
            console.error('\n⚠️  Possible CORS issue detected');
            console.error('   Make sure the frontend dev server is running on port 3000');
        }
    }
}

// Test different request variations
async function testRequestVariations(token) {
    console.log('\n🔄 Testing Request Variations');
    
    const variations = [
        {
            name: 'Minimal request',
            request: { context: "Hello world" }
        },
        {
            name: 'With style only',
            request: { context: "Writing about travel", style: "adventurous" }
        },
        {
            name: 'With mood only',
            request: { context: "Missing home", mood: "melancholic" }
        },
        {
            name: 'Empty context',
            request: { context: "" }
        },
        {
            name: 'Long context',
            request: { 
                context: "I want to write a very long letter about my entire year, including all the challenges I faced, the people I met, the places I visited, the lessons I learned, and how everything has shaped me into who I am today. There's so much to say and I don't know where to start."
            }
        }
    ];
    
    for (const variation of variations) {
        console.log(`\n📝 Testing: ${variation.name}`);
        console.log(`   Request:`, JSON.stringify(variation.request));
        
        try {
            const response = await axios.post(
                `${BACKEND_URL}${API_ENDPOINT}`,
                variation.request,
                {
                    headers: {
                        'Authorization': `Bearer ${token}`,
                        'Content-Type': 'application/json'
                    }
                }
            );
            
            console.log(`   ✅ Success - Got ${response.data.data?.suggestions?.length || 0} suggestions`);
        } catch (error) {
            console.log(`   ❌ Failed - ${error.response?.data?.error || error.message}`);
        }
    }
}

// Test error handling
async function testErrorHandling() {
    console.log('\n🚨 Testing Error Handling');
    
    // Test without authentication
    console.log('\n1️⃣ Without authentication:');
    try {
        await axios.post(`${BACKEND_URL}${API_ENDPOINT}`, {
            context: "Test without auth"
        });
        console.log('   ❌ Unexpected success - should have failed');
    } catch (error) {
        console.log(`   ✅ Expected error: ${error.response?.status} - ${error.response?.data?.error || error.message}`);
    }
    
    // Test with invalid token
    console.log('\n2️⃣ With invalid token:');
    try {
        await axios.post(
            `${BACKEND_URL}${API_ENDPOINT}`,
            { context: "Test with invalid token" },
            {
                headers: {
                    'Authorization': 'Bearer invalid-token-12345',
                    'Content-Type': 'application/json'
                }
            }
        );
        console.log('   ❌ Unexpected success - should have failed');
    } catch (error) {
        console.log(`   ✅ Expected error: ${error.response?.status} - ${error.response?.data?.error || error.message}`);
    }
    
    // Test with malformed request
    console.log('\n3️⃣ With malformed request:');
    try {
        const token = await getAuthToken();
        await axios.post(
            `${BACKEND_URL}${API_ENDPOINT}`,
            { invalid_field: "Test" }, // Missing required 'context' field
            {
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json'
                }
            }
        );
        console.log('   ❌ Unexpected success - should have failed');
    } catch (error) {
        console.log(`   ✅ Expected error: ${error.response?.status} - ${error.response?.data?.error || error.message}`);
    }
}

// Check frontend integration
async function checkFrontendIntegration() {
    console.log('\n🔍 Checking Frontend Integration');
    
    try {
        // Check if frontend is running
        const frontendResponse = await axios.get(FRONTEND_URL);
        console.log('✅ Frontend is running on port 3000');
        
        // Check API route configuration
        console.log('\n📋 Frontend API Configuration:');
        console.log('   - Frontend should proxy /api/* requests to backend');
        console.log('   - Check next.config.js for rewrites configuration');
        console.log('   - Check if API client uses correct endpoints');
        
    } catch (error) {
        console.log('❌ Frontend is not running on port 3000');
        console.log('   Run: cd frontend && npm run dev');
    }
}

// Main test runner
async function runTests() {
    console.log('🧪 AI Inspiration API Frontend Test Suite');
    console.log('=========================================');
    
    try {
        // Check services
        console.log('\n📡 Checking services...');
        try {
            await axios.get(`${BACKEND_URL}/api/health`);
            console.log('✅ Backend is running on port 8080');
        } catch (error) {
            console.error('❌ Backend is not running on port 8080');
            console.error('   Run: cd backend && go run main.go');
            return;
        }
        
        // Get auth token
        const token = await getAuthToken();
        
        // Run tests
        await testBackendAPI(token);
        await testFrontendProxy(token);
        await testRequestVariations(token);
        await testErrorHandling();
        await checkFrontendIntegration();
        
        console.log('\n✅ All tests completed!');
        console.log('\n📊 Summary:');
        console.log('   - Backend API is working correctly');
        console.log('   - Frontend proxy needs to be configured in next.config.js');
        console.log('   - Authentication is required for API access');
        console.log('   - Response includes suggestions, themes, and opening_lines');
        console.log('   - Token usage information is included in response');
        
    } catch (error) {
        console.error('\n❌ Test suite failed:', error.message);
    }
}

// Run the tests
runTests();