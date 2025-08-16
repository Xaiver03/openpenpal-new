const { Pool } = require('pg');
const bcrypt = require('bcryptjs');

// PostgreSQL 连接配置
const pool = new Pool({
  host: 'localhost',
  port: 5432,
  database: 'openpenpal',
  user: process.env.USER || 'postgres',
  password: process.env.DB_PASSWORD || '',
});

async function debugLogin() {
  try {
    console.log('Debugging login issue...\n');
    
    // 1. Check alice user in database
    const userResult = await pool.query(
      "SELECT id, username, email, password_hash, role, is_active FROM users WHERE username = 'alice'"
    );
    
    if (userResult.rows.length === 0) {
      console.log('❌ User alice not found in database');
      return;
    }
    
    const user = userResult.rows[0];
    console.log('User found:');
    console.log('- ID:', user.id);
    console.log('- Username:', user.username);
    console.log('- Email:', user.email);
    console.log('- Role:', user.role);
    console.log('- Active:', user.is_active);
    console.log('- Password Hash:', user.password_hash);
    console.log('\n');
    
    // 2. Test bcrypt comparison
    const testPassword = 'secret';
    const knownGoodHash = '$2a$10$N9qo8uLOickgx2ZMRZoMye.Lf/rSDRfYHYxX1dpIjJJNzTHFN1UTO';
    
    console.log('Testing bcrypt comparison:');
    console.log('- Test password:', testPassword);
    console.log('- Known good hash for "secret":', knownGoodHash);
    
    // Test with known good hash
    try {
      const isValidKnown = await bcrypt.compare(testPassword, knownGoodHash);
      console.log('- Comparison with known hash:', isValidKnown ? '✅ Valid' : '❌ Invalid');
    } catch (err) {
      console.log('- Error comparing with known hash:', err.message);
    }
    
    // Test with database hash
    try {
      const isValidDb = await bcrypt.compare(testPassword, user.password_hash);
      console.log('- Comparison with DB hash:', isValidDb ? '✅ Valid' : '❌ Invalid');
    } catch (err) {
      console.log('- Error comparing with DB hash:', err.message);
    }
    
    // 3. Check if hashes match
    if (user.password_hash === knownGoodHash) {
      console.log('\n✅ Database hash matches known good hash');
    } else {
      console.log('\n❌ Database hash does NOT match known good hash');
      console.log('This is the problem!');
    }
    
  } catch (error) {
    console.error('Error:', error);
  } finally {
    await pool.end();
  }
}

debugLogin();