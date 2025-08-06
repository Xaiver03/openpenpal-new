// Debug token expiration issue
const jwt = require('jsonwebtoken');

console.log('=== JWT Token Expiration Issue Debug ===\n');

// Test the same isExpired function as in the TokenManager
function isExpired(token) {
  try {
    const payload = JSON.parse(atob(token.split('.')[1]));
    const now = Date.now();
    const expTime = payload.exp * 1000;
    
    console.log('Current timestamp (ms):', now);
    console.log('Token exp timestamp (ms):', expTime);
    console.log('Difference (ms):', expTime - now);
    console.log('Difference (hours):', (expTime - now) / (1000 * 60 * 60));
    
    return now >= expTime;
  } catch (e) {
    console.error('Error parsing token:', e.message);
    return true;
  }
}

// Generate a token with 24-hour expiry like the backend does
const secret = 'dev-secret-key-do-not-use-in-production';
const now = Date.now();
const expiresAt = new Date(now + (24 * 60 * 60 * 1000)); // 24 hours from now

const payload = {
  user_id: 'test-user',
  role: 'user',
  iat: Math.floor(now / 1000),
  exp: Math.floor(expiresAt.getTime() / 1000)
};

console.log('Generating token with:');
console.log('- Issued at:', new Date(payload.iat * 1000).toISOString());
console.log('- Expires at:', new Date(payload.exp * 1000).toISOString());
console.log('- Duration (hours):', (payload.exp - payload.iat) / 3600);

const token = jwt.sign(payload, secret);
console.log('\nGenerated token:', token.substring(0, 50) + '...');

// Test expiration check
console.log('\n=== Testing isExpired function ===');
const expired = isExpired(token);
console.log('Is token expired?', expired);

// Simulate what happens when clock is slightly off
console.log('\n=== Testing with clock drift ===');
const testToken = jwt.sign({
  user_id: 'test-user',
  role: 'user',
  iat: Math.floor(now / 1000) - 1, // Issued 1 second ago
  exp: Math.floor(now / 1000) + 3600 // Expires in 1 hour
}, secret);

console.log('Test token (1 hour expiry):');
const testExpired = isExpired(testToken);
console.log('Is test token expired?', testExpired);

// Test with very short expiry
console.log('\n=== Testing with short expiry ===');
const shortToken = jwt.sign({
  user_id: 'test-user',
  role: 'user', 
  iat: Math.floor(now / 1000),
  exp: Math.floor(now / 1000) + 5 // Expires in 5 seconds
}, secret);

console.log('Short token (5 seconds expiry):');
console.log('Is short token expired?', isExpired(shortToken));

setTimeout(() => {
  console.log('After 6 seconds, is short token expired?', isExpired(shortToken));
}, 6000);