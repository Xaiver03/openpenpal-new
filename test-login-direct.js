#!/usr/bin/env node

const { Client } = require('pg');
const bcrypt = require('bcryptjs');

// Database configuration
const dbConfig = {
  host: process.env.DATABASE_HOST || 'localhost',
  port: process.env.DATABASE_PORT || 5432,
  database: process.env.DATABASE_NAME || 'openpenpal',
  user: process.env.DATABASE_USER || 'rocalight',
  password: process.env.DATABASE_PASSWORD || 'password',
};

async function testLogin() {
  const client = new Client(dbConfig);
  
  try {
    await client.connect();
    console.log('✅ Connected to PostgreSQL\n');
    
    // Query for courier_level1 user
    const userQuery = `
      SELECT id, username, email, password_hash, role, is_active
      FROM users
      WHERE username = $1
    `;
    
    const userResult = await client.query(userQuery, ['courier_level1']);
    
    if (userResult.rows.length === 0) {
      console.log('❌ User courier_level1 not found');
      return;
    }
    
    const user = userResult.rows[0];
    console.log('👤 Found user:');
    console.log('  ID:', user.id);
    console.log('  Username:', user.username);
    console.log('  Email:', user.email);
    console.log('  Role:', user.role);
    console.log('  Active:', user.is_active);
    console.log('  Password hash:', user.password_hash ? user.password_hash.substring(0, 20) + '...' : 'NULL');
    
    // Test password verification
    console.log('\n🔐 Testing password verification...');
    const testPassword = 'secret';
    
    try {
      const isValid = await bcrypt.compare(testPassword, user.password_hash);
      console.log(`  Password "${testPassword}" is ${isValid ? '✅ VALID' : '❌ INVALID'}`);
      
      if (!isValid) {
        // Generate a new hash for comparison
        const newHash = await bcrypt.hash(testPassword, 10);
        console.log('\n  Expected hash format:', newHash.substring(0, 20) + '...');
        console.log('  Actual hash format:  ', user.password_hash.substring(0, 20) + '...');
        
        // Update with correct hash
        console.log('\n🔧 Updating password hash...');
        const updateResult = await client.query(
          'UPDATE users SET password_hash = $1 WHERE username = $2',
          [newHash, 'courier_level1']
        );
        console.log('  Update result:', updateResult.rowCount, 'row(s) updated');
      }
    } catch (err) {
      console.error('  Error during password verification:', err.message);
    }
    
  } catch (err) {
    console.error('❌ Database error:', err.message);
  } finally {
    await client.end();
    console.log('\n🔌 Database connection closed');
  }
}

console.log('🧪 Testing login for courier_level1...\n');
testLogin();