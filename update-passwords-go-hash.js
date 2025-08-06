const { Pool } = require('pg');

// PostgreSQL 连接配置
const pool = new Pool({
  host: 'localhost',
  port: 5432,
  database: 'openpenpal',
  user: process.env.USER || 'postgres',
  password: process.env.DB_PASSWORD || '',
});

async function updatePasswords() {
  try {
    console.log('Updating passwords with Go-generated hashes...\n');
    
    // Go-generated bcrypt hashes
    const secretHash = '$2a$10$KuNOKKOmFExYEe/BYHOQWOtuwywR3mHeOeBm7On0ZAozMWVqcmoU.';
    const admin123Hash = '$2a$10$cH8Xq3cHw.nxkHBtepdYBekdP/85F1cn1LMBqii7tjB.VSmjInf/i';
    
    // Update admin password
    const adminResult = await pool.query(
      'UPDATE users SET password_hash = $1 WHERE username = $2',
      [admin123Hash, 'admin']
    );
    console.log(`Updated admin: ${adminResult.rowCount} row(s) affected`);
    
    // Update users with "secret" password
    const usersToUpdate = ['alice', 'courier_level1', 'courier_level2', 'courier_level3', 'courier_level4'];
    
    for (const username of usersToUpdate) {
      const result = await pool.query(
        'UPDATE users SET password_hash = $1 WHERE username = $2',
        [secretHash, username]
      );
      console.log(`Updated ${username}: ${result.rowCount} row(s) affected`);
    }
    
    console.log('\n✅ All passwords updated with Go-generated hashes!');
    
  } catch (error) {
    console.error('Error updating passwords:', error);
  } finally {
    await pool.end();
  }
}

updatePasswords();