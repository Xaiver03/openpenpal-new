const axios = require('axios');

async function testCourierLogin() {
  const testAccounts = [
    { username: 'courier_level1', password: 'password' },
    { username: 'courier_level1', password: 'secret' },
    { username: 'courier1', password: 'password' },
    { username: 'admin', password: 'password' },
    { username: 'user1', password: 'password' }
  ];

  console.log('🧪 Testing login for courier accounts...\n');

  for (const account of testAccounts) {
    try {
      console.log(`📍 Testing: ${account.username} / ${account.password}`);
      
      // Step 1: Get CSRF token
      const csrfResponse = await axios.get('http://localhost:8080/api/v1/auth/csrf');
      const csrfToken = csrfResponse.data.data.token;
      const cookies = csrfResponse.headers['set-cookie'];
      
      console.log(`   ✅ CSRF token obtained`);
      
      // Step 2: Login
      const loginResponse = await axios.post('http://localhost:8080/api/v1/auth/login', {
        username: account.username,
        password: account.password
      }, {
        headers: {
          'X-CSRF-Token': csrfToken,
          'Cookie': cookies ? cookies.join('; ') : ''
        }
      });
      
      console.log(`   ✅ Login successful!`);
      console.log(`   📧 Email: ${loginResponse.data.data.user.email}`);
      console.log(`   👤 Role: ${loginResponse.data.data.user.role}`);
      console.log(`   🏫 School: ${loginResponse.data.data.user.school_code || 'N/A'}`);
      console.log(`   🔑 Token: ${loginResponse.data.data.token.substring(0, 20)}...`);
      console.log('   ---');
      
    } catch (error) {
      console.log(`   ❌ Login failed: ${error.response?.data?.message || error.message}`);
      console.log('   ---');
    }
  }
}

// Also test direct database query
async function checkDatabasePasswords() {
  console.log('\n📊 Checking database password hashes...\n');
  
  const { Client } = require('pg');
  const client = new Client({
    connectionString: 'postgres://rocalight:password@localhost:5432/openpenpal'
  });
  
  try {
    await client.connect();
    
    const result = await client.query(`
      SELECT username, email, password_hash, role, is_active 
      FROM users 
      WHERE username IN ('courier_level1', 'courier1', 'admin', 'user1', 'alice')
      ORDER BY username
    `);
    
    console.log('Found users in database:');
    result.rows.forEach(row => {
      console.log(`- ${row.username} (${row.email})`);
      console.log(`  Role: ${row.role}, Active: ${row.is_active}`);
      console.log(`  Hash: ${row.password_hash.substring(0, 20)}...`);
    });
    
    // Test bcrypt directly
    const bcrypt = require('bcrypt');
    console.log('\n🔐 Testing password verification...\n');
    
    const passwordsToTest = ['password', 'secret', 'admin123'];
    const hashToTest = result.rows.find(r => r.username === 'courier_level1')?.password_hash;
    
    if (hashToTest) {
      for (const pwd of passwordsToTest) {
        const match = await bcrypt.compare(pwd, hashToTest);
        console.log(`courier_level1 + "${pwd}": ${match ? '✅ MATCH' : '❌ NO MATCH'}`);
      }
    }
    
  } catch (error) {
    console.error('Database error:', error.message);
  } finally {
    await client.end();
  }
}

// Run tests
(async () => {
  await checkDatabasePasswords();
  console.log('\n');
  await testCourierLogin();
})();