const { Pool } = require('pg');

// PostgreSQL 连接配置
const pool = new Pool({
  host: 'localhost',
  port: 5432,
  database: 'openpenpal',
  user: process.env.USER || 'postgres',
  password: process.env.DB_PASSWORD || '',
});

async function fixUserPasswords() {
  try {
    console.log('Fixing user passwords in database...\n');
    
    // Correct bcrypt hash for "secret"
    const secretHash = '$2a$10$N9qo8uLOickgx2ZMRZoMye.Lf/rSDRfYHYxX1dpIjJJNzTHFN1UTO';
    
    // Update users with "secret" password
    const usersToUpdate = ['alice', 'courier_level1', 'courier_level2', 'courier_level3', 'courier_level4'];
    
    for (const username of usersToUpdate) {
      const result = await pool.query(
        'UPDATE users SET password_hash = $1 WHERE username = $2',
        [secretHash, username]
      );
      console.log(`Updated ${username}: ${result.rowCount} row(s) affected`);
    }
    
    console.log('\n✅ Password hashes updated successfully!');
    console.log('\nVerifying updates...\n');
    
    // Verify the updates
    const verifyResult = await pool.query(`
      SELECT username, password_hash 
      FROM users 
      WHERE username IN ('alice', 'courier_level1', 'courier_level2', 'courier_level3', 'courier_level4')
      ORDER BY username
    `);
    
    for (const user of verifyResult.rows) {
      const isCorrect = user.password_hash === secretHash;
      console.log(`${user.username}: ${isCorrect ? '✅' : '❌'} ${isCorrect ? 'Correct' : 'Incorrect'}`);
    }
    
  } catch (error) {
    console.error('Error fixing passwords:', error);
  } finally {
    await pool.end();
  }
}

fixUserPasswords();