// Test token expiry fix
const jwt = require('jsonwebtoken');

// 🔐 安全令牌生成 - 替代硬编码令牌
const { generateTestToken } = require('../../backend/scripts/test-token-generator');

console.log('=== Testing JWT Token Expiry Fix ===\n');

// Simulate what the backend should return
const backendResponse = {
  code: 0,
  message: "Login successful",
  data: {
    token: generateTestToken('USER', {}, '24h'), // 安全生成的24小时令牌
    expires_at: new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString(), // 24 hours from now
    user: {
      id: "test-user-123",
      username: "admin",
      role: "admin"
    }
  }
};

console.log('Backend response expires_at:', backendResponse.data.expires_at);
console.log('Backend response expiresAt:', backendResponse.data.expiresAt);

// Test the old logic (incorrect)
const oldExpiresAt = backendResponse.data.expiresAt || new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString();
console.log('\nOld logic result (should be fallback):', oldExpiresAt);

// Test the new logic (fixed)
const newExpiresAt = backendResponse.data.expires_at || backendResponse.data.expiresAt || new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString();
console.log('New logic result (should match backend):', newExpiresAt);

// Verify they're different
console.log('\nAre they the same?', oldExpiresAt === newExpiresAt);
console.log('Time difference (ms):', Math.abs(new Date(oldExpiresAt).getTime() - new Date(newExpiresAt).getTime()));

// Test token parsing
const token = backendResponse.data.token;
try {
  const decoded = jwt.decode(token);
  console.log('\nToken expires at (from JWT):', new Date(decoded.exp * 1000).toISOString());
  console.log('Backend expires_at field:', backendResponse.data.expires_at);
  console.log('Should they match?', Math.abs(decoded.exp * 1000 - new Date(backendResponse.data.expires_at).getTime()) < 1000);
} catch (e) {
  console.error('Token decode error:', e.message);
}

console.log('\n=== Fix Summary ===');
console.log('Issue: Frontend was looking for expiresAt but backend returns expires_at');
console.log('Solution: Check both expires_at and expiresAt in that order');
console.log('Result: Now uses actual backend expiry time instead of fallback');