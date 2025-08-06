// JWT debugging script
const jwt = require('jsonwebtoken');

// Test JWT with known secret
const testPayload = {
  user_id: 'test-user-123',
  role: 'user',
  iat: Math.floor(Date.now() / 1000),
  exp: Math.floor(Date.now() / 1000) + (24 * 60 * 60) // 24 hours
};

const secret = 'dev-secret-key-do-not-use-in-production';

try {
  // Generate a test token
  const token = jwt.sign(testPayload, secret);
  console.log('Generated token:', token);
  console.log('\nToken length:', token.length);
  
  // Decode without verification to see payload
  const decoded = jwt.decode(token, { complete: true });
  console.log('\nDecoded header:', JSON.stringify(decoded.header, null, 2));
  console.log('\nDecoded payload:', JSON.stringify(decoded.payload, null, 2));
  
  // Calculate expiry time
  const now = new Date();
  const expiry = new Date(decoded.payload.exp * 1000);
  const timeDiff = expiry - now;
  const hoursDiff = timeDiff / (1000 * 60 * 60);
  
  console.log('\nTime analysis:');
  console.log('Current time:', now.toISOString());
  console.log('Expiry time:', expiry.toISOString());
  console.log('Hours until expiry:', hoursDiff.toFixed(2));
  
  // Verify the token
  const verified = jwt.verify(token, secret);
  console.log('\nVerification successful:', !!verified);
  
} catch (error) {
  console.error('JWT Error:', error.message);
}