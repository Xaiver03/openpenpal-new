// Debug JWT Secret Mismatch
const jwt = require('jsonwebtoken');

// Test both possible JWT secrets
const BACKEND_SECRET_NEW = 'KY6QtIecDZocllQSYoqyTkYx8AuKDkpA7RfondzVB2Y=';
const BACKEND_SECRET_OLD = 'your-production-secret-key-here';
const BACKEND_SECRET_FALLBACK = 'dev-secret-key-do-not-use-in-production';
const FRONTEND_SECRET = 'KY6QtIecDZocllQSYoqyTkYx8AuKDkpA7RfondzVB2Y=';

console.log('🔍 JWT Secret Mismatch Debugging\n');

// Get a token from the backend
async function getBackendToken() {
  try {
    const response = await fetch('http://localhost:8080/api/v1/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username: 'admin', password: 'admin123' })
    });
    
    const result = await response.json();
    if (result.data?.token) {
      return result.data.token;
    }
    throw new Error('No token received');
  } catch (error) {
    console.error('❌ Failed to get backend token:', error.message);
    return null;
  }
}

function testTokenWithSecret(token, secret, secretName) {
  try {
    const decoded = jwt.verify(token, secret);
    console.log(`✅ ${secretName}: Token valid`);
    console.log(`   User ID: ${decoded.userId || decoded.user_id || decoded.UserID}`);
    console.log(`   Role: ${decoded.role || decoded.Role}`);
    console.log(`   Expires: ${new Date(decoded.exp * 1000).toISOString()}`);
    return true;
  } catch (error) {
    console.log(`❌ ${secretName}: ${error.message}`);
    return false;
  }
}

async function debugJWTMismatch() {
  console.log('🚀 Getting token from backend...');
  const token = await getBackendToken();
  
  if (!token) {
    console.log('❌ Cannot continue without token');
    return;
  }
  
  console.log(`\n📍 Token received (first 50 chars): ${token.substring(0, 50)}...`);
  
  // Try to decode without verification to see payload
  try {
    const decoded = jwt.decode(token, { complete: true });
    console.log('\n📋 Token header:', JSON.stringify(decoded.header, null, 2));
    console.log('📋 Token payload:', JSON.stringify(decoded.payload, null, 2));
  } catch (error) {
    console.log('❌ Failed to decode token:', error.message);
  }
  
  console.log('\n🧪 Testing token with different secrets:');
  
  // Test with all possible secrets
  const secrets = [
    { name: 'Updated Backend Secret (.env.production)', secret: BACKEND_SECRET_NEW },
    { name: 'Old Backend Secret', secret: BACKEND_SECRET_OLD },
    { name: 'Fallback Backend Secret', secret: BACKEND_SECRET_FALLBACK },
    { name: 'Frontend Secret (.env.local)', secret: FRONTEND_SECRET }
  ];
  
  let validSecrets = [];
  for (const { name, secret } of secrets) {
    if (testTokenWithSecret(token, secret, name)) {
      validSecrets.push(name);
    }
  }
  
  console.log('\n📊 Results:');
  if (validSecrets.length === 0) {
    console.log('❌ Token is not valid with any of the tested secrets');
    console.log('💡 The backend might be using a different secret or signing method');
  } else {
    console.log(`✅ Token is valid with: ${validSecrets.join(', ')}`);
    if (validSecrets.includes('Frontend Secret (.env.local)')) {
      console.log('🎉 JWT secrets are properly aligned!');
    } else {
      console.log('⚠️ Backend and frontend are using different JWT secrets');
    }
  }
}

debugJWTMismatch().catch(console.error);