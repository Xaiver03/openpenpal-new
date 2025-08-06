#!/usr/bin/env node

const { Client } = require('pg');

// Database configuration
const dbConfig = {
  host: process.env.DATABASE_HOST || 'localhost',
  port: process.env.DATABASE_PORT || 5432,
  database: process.env.DATABASE_NAME || 'openpenpal',
  user: process.env.DATABASE_USER || 'rocalight',
  password: process.env.DATABASE_PASSWORD || 'password',
};

// The correct bcrypt hash for password "secret"
const correctPasswordHash = '$2a$10$N9qo8uLOickgx2ZMRZoMye.Lf/rSDRfYHYxX1dpIjJJNzTHFN1UTO';

// Users to update
const courierUsers = [
  'courier_level1',
  'courier_level2', 
  'courier_level3',
  'courier_level4'
];

async function updatePasswords() {
  const client = new Client(dbConfig);
  
  try {
    await client.connect();
    console.log('✅ Connected to PostgreSQL');
    
    for (const username of courierUsers) {
      const updateQuery = `
        UPDATE users 
        SET password_hash = $1
        WHERE username = $2
        RETURNING username, email, role
      `;
      
      const result = await client.query(updateQuery, [correctPasswordHash, username]);
      
      if (result.rowCount > 0) {
        console.log(`✅ Updated password for ${username}:`, result.rows[0]);
      } else {
        console.log(`⚠️  User ${username} not found`);
      }
    }
    
    // Verify the update
    console.log('\n🔐 Verifying password updates...');
    const verifyQuery = `
      SELECT username, password_hash = $1 as has_correct_password
      FROM users
      WHERE username = ANY($2)
    `;
    
    const verifyResult = await client.query(verifyQuery, [correctPasswordHash, courierUsers]);
    console.log('Password verification results:');
    verifyResult.rows.forEach(row => {
      console.log(`  ${row.username}: ${row.has_correct_password ? '✅ Correct' : '❌ Incorrect'}`);
    });
    
  } catch (err) {
    console.error('❌ Database error:', err.message);
  } finally {
    await client.end();
    console.log('\n🔌 Database connection closed');
  }
}

console.log('🔧 Updating courier user passwords to "secret"...\n');
updatePasswords();