const { Pool } = require('pg');

// PostgreSQL 连接配置
const pool = new Pool({
  host: 'localhost',
  port: 5432,
  database: 'openpenpal',
  user: process.env.USER || 'postgres',
  password: process.env.DB_PASSWORD || '',
});

async function checkUserPasswords() {
  try {
    console.log('Checking user passwords in database...\n');
    
    // Query users table
    const result = await pool.query(`
      SELECT id, username, email, password_hash, role, is_active 
      FROM users 
      WHERE username IN ('admin', 'alice', 'courier_level1', 'courier_level2', 'courier_level3', 'courier_level4')
      ORDER BY username
    `);
    
    console.log(`Found ${result.rows.length} users:\n`);
    
    const knownHashes = {
      'secret': '$2a$10$N9qo8uLOickgx2ZMRZoMye.Lf/rSDRfYHYxX1dpIjJJNzTHFN1UTO',
      'admin123': '$2a$10$dwSXE/fBcbAJVy0jMZHYI.vFjjUZFYRMPpeAzcgmHd.XqwfqgOrEW'
    };
    
    for (const user of result.rows) {
      console.log(`Username: ${user.username}`);
      console.log(`Email: ${user.email || 'N/A'}`);
      console.log(`Role: ${user.role}`);
      console.log(`Active: ${user.is_active}`);
      console.log(`Password Hash: ${user.password_hash}`);
      
      // Check if it matches known hashes
      if (user.password_hash === knownHashes.secret) {
        console.log('✅ Password hash matches "secret"');
      } else if (user.password_hash === knownHashes.admin123) {
        console.log('✅ Password hash matches "admin123"');
      } else {
        console.log('❌ Password hash does not match known values');
      }
      
      console.log('-'.repeat(60) + '\n');
    }
    
  } catch (error) {
    console.error('Error checking passwords:', error);
  } finally {
    await pool.end();
  }
}

checkUserPasswords();