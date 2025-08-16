// Test AI Frontend Components
// This script tests the AI functionality through the frontend API

const API_URL = 'http://localhost:8080/api/v1';

// Get auth token (replace with actual token after login)
const AUTH_TOKEN = 'YOUR_AUTH_TOKEN_HERE';

// Color codes for console output
const colors = {
  success: '\x1b[32m',
  error: '\x1b[31m',
  info: '\x1b[36m',
  warning: '\x1b[33m',
  reset: '\x1b[0m'
};

async function testAIEndpoint(endpoint, method = 'GET', body = null) {
  console.log(`\n${colors.info}Testing ${method} ${endpoint}...${colors.reset}`);
  
  try {
    const options = {
      method,
      headers: {
        'Authorization': `Bearer ${AUTH_TOKEN}`,
        'Content-Type': 'application/json'
      }
    };
    
    if (body && method !== 'GET') {
      options.body = JSON.stringify(body);
    }
    
    const response = await fetch(`${API_URL}${endpoint}`, options);
    const data = await response.json();
    
    if (response.ok) {
      console.log(`${colors.success}✓ Success (${response.status}):${colors.reset}`);
      console.log(JSON.stringify(data, null, 2));
      return { success: true, data };
    } else {
      console.log(`${colors.error}✗ Error (${response.status}):${colors.reset}`);
      console.log(JSON.stringify(data, null, 2));
      return { success: false, error: data };
    }
  } catch (error) {
    console.log(`${colors.error}✗ Network Error:${colors.reset}`, error.message);
    return { success: false, error: error.message };
  }
}

async function runTests() {
  console.log(`${colors.info}=== Testing AI Frontend Functionality ===${colors.reset}`);
  console.log(`${colors.warning}Note: Make sure to set AUTH_TOKEN first!${colors.reset}`);
  
  // Test 1: Get AI Personas
  console.log(`\n${colors.info}1. Testing AI Personas${colors.reset}`);
  await testAIEndpoint('/ai/personas');
  
  // Test 2: Get Daily Inspiration
  console.log(`\n${colors.info}2. Testing Daily Inspiration${colors.reset}`);
  await testAIEndpoint('/ai/daily-inspiration');
  
  // Test 3: Generate Writing Inspiration
  console.log(`\n${colors.info}3. Testing Writing Inspiration${colors.reset}`);
  await testAIEndpoint('/ai/inspiration', 'POST', {
    theme: '日常生活',
    count: 3
  });
  
  // Test 4: Get AI Stats
  console.log(`\n${colors.info}4. Testing AI Stats${colors.reset}`);
  await testAIEndpoint('/ai/stats');
  
  // Test 5: Test AI Reply (requires valid letter_id)
  console.log(`\n${colors.info}5. Testing AI Reply Generation${colors.reset}`);
  await testAIEndpoint('/ai/reply', 'POST', {
    letter_id: '24b6c37e-b2eb-4639-9bc8-8834cea914e2',
    persona: 'friend',
    delay_hours: 24
  });
  
  console.log(`\n${colors.info}=== Test Summary ===${colors.reset}`);
  console.log(`${colors.warning}To run these tests:${colors.reset}`);
  console.log('1. Start the backend: cd backend && go run main.go');
  console.log('2. Login to get a token (use test account: alice/secret)');
  console.log('3. Update AUTH_TOKEN in this script');
  console.log('4. Run: node test-ai-frontend.js');
}

// Check if backend is running
async function checkBackend() {
  try {
    const response = await fetch(`${API_URL}/health`);
    if (response.ok) {
      console.log(`${colors.success}✓ Backend is running${colors.reset}`);
      return true;
    }
  } catch (error) {
    console.log(`${colors.error}✗ Backend is not running at ${API_URL}${colors.reset}`);
    console.log(`${colors.warning}Please start the backend first: cd backend && go run main.go${colors.reset}`);
    return false;
  }
}

// Main execution
(async () => {
  const backendReady = await checkBackend();
  if (backendReady && AUTH_TOKEN !== 'YOUR_AUTH_TOKEN_HERE') {
    await runTests();
  } else if (AUTH_TOKEN === 'YOUR_AUTH_TOKEN_HERE') {
    console.log(`\n${colors.warning}Please set AUTH_TOKEN first!${colors.reset}`);
    console.log('1. Login with: curl -X POST http://localhost:8080/api/v1/auth/login -d \'{"username":"alice","password":"secret"}\'');
    console.log('2. Copy the token from the response');
    console.log('3. Update AUTH_TOKEN in this script');
  }
})();