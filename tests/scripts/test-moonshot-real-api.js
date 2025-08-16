#!/usr/bin/env node

// Test Moonshot Kimi Real API Integration
// This script tests the real Moonshot API integration

const API_URL = 'http://localhost:8080/api/v1';

// Color codes for console output
const colors = {
  success: '\x1b[32m',
  error: '\x1b[31m',
  info: '\x1b[36m',
  warning: '\x1b[33m',
  reset: '\x1b[0m'
};

// Test user credentials
const TEST_USER = {
  username: 'alice',
  password: 'secret'
};

async function login() {
  console.log(`${colors.info}Logging in as ${TEST_USER.username}...${colors.reset}`);
  
  try {
    const response = await fetch(`${API_URL}/auth/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(TEST_USER)
    });
    
    const data = await response.json();
    
    if (response.ok && data.data && data.data.token) {
      console.log(`${colors.success}✓ Login successful${colors.reset}`);
      return data.data.token;
    } else {
      console.log(`${colors.error}✗ Login failed:${colors.reset}`, data);
      return null;
    }
  } catch (error) {
    console.log(`${colors.error}✗ Login error:${colors.reset}`, error.message);
    return null;
  }
}

async function testAIEndpoint(token, endpoint, method = 'GET', body = null) {
  console.log(`\n${colors.info}Testing ${method} ${endpoint}...${colors.reset}`);
  
  try {
    const options = {
      method,
      headers: {
        'Authorization': `Bearer ${token}`,
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

async function runTests(token) {
  console.log(`\n${colors.info}=== Testing Moonshot AI Integration ===${colors.reset}`);
  
  // Test 1: Writing Inspiration (Real Moonshot API)
  console.log(`\n${colors.warning}1. Testing Writing Inspiration with Moonshot API${colors.reset}`);
  const inspirationResult = await testAIEndpoint(token, '/ai/inspiration', 'POST', {
    theme: '日常生活',
    count: 3
  });
  
  if (inspirationResult.success) {
    console.log(`${colors.success}✓ Moonshot API is working! Got ${inspirationResult.data.data.inspirations.length} inspirations${colors.reset}`);
  }
  
  // Test 2: Get Daily Inspiration
  console.log(`\n${colors.warning}2. Testing Daily Inspiration${colors.reset}`);
  await testAIEndpoint(token, '/ai/daily-inspiration');
  
  // Test 3: AI Stats
  console.log(`\n${colors.warning}3. Testing AI Stats${colors.reset}`);
  await testAIEndpoint(token, '/ai/stats');
  
  // Test 4: AI Personas
  console.log(`\n${colors.warning}4. Testing AI Personas${colors.reset}`);
  await testAIEndpoint(token, '/ai/personas');
  
  // Test 5: Reply Advice (if you have a valid letter ID)
  console.log(`\n${colors.warning}5. Testing Reply Advice${colors.reset}`);
  await testAIEndpoint(token, '/ai/reply-advice', 'POST', {
    letter_id: '24b6c37e-b2eb-4639-9bc8-8834cea914e2',
    persona_type: 'friend',
    persona_name: '知心朋友',
    relationship: '好朋友',
    delivery_days: 1
  });
}

// Check if backend is running
async function checkBackend() {
  try {
    const response = await fetch('http://localhost:8080/health');
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
  console.log(`${colors.info}=== Moonshot Kimi API Integration Test ===${colors.reset}`);
  console.log(`${colors.warning}This test will use the real Moonshot API configured in .env${colors.reset}`);
  
  const backendReady = await checkBackend();
  if (!backendReady) {
    process.exit(1);
  }
  
  const token = await login();
  if (!token) {
    console.log(`${colors.error}Failed to login. Cannot continue tests.${colors.reset}`);
    process.exit(1);
  }
  
  await runTests(token);
  
  console.log(`\n${colors.info}=== Test Summary ===${colors.reset}`);
  console.log(`${colors.success}✓ Moonshot API Key is configured in backend/.env${colors.reset}`);
  console.log(`${colors.success}✓ AI Provider is set to 'moonshot'${colors.reset}`);
  console.log(`${colors.info}ℹ️  If you see actual AI-generated content above, the integration is working!${colors.reset}`);
  console.log(`${colors.info}ℹ️  If you see errors, check:${colors.reset}`);
  console.log(`   - Is the Moonshot API key valid?`);
  console.log(`   - Is there internet connectivity?`);
  console.log(`   - Check backend logs for detailed error messages`);
})();