const { Pool } = require('pg');

// PostgreSQL 连接配置
const pool = new Pool({
  host: 'localhost',
  port: 5432,
  database: 'openpenpal',
  user: process.env.USER || 'postgres',
  password: process.env.DB_PASSWORD || '',
});

async function testQuery() {
  try {
    console.log('Testing database queries...\n');
    
    // Test the exact query used in Go code
    const username = 'alice';
    const result = await pool.query(
      "SELECT * FROM users WHERE username = $1 OR email = $1",
      [username]
    );
    
    console.log(`Query: SELECT * FROM users WHERE username = '${username}' OR email = '${username}'`);
    console.log(`Found ${result.rows.length} row(s)\n`);
    
    if (result.rows.length > 0) {
      const user = result.rows[0];
      console.log('User details:');
      console.log('- ID:', user.id);
      console.log('- Username:', user.username);
      console.log('- Email:', user.email);
      console.log('- Role:', user.role);
      console.log('- Active:', user.is_active);
      console.log('- Password Hash Length:', user.password_hash ? user.password_hash.length : 0);
      console.log('- Created:', user.created_at);
      console.log('- Updated:', user.updated_at);
    }
    
    // Also check admin for comparison
    console.log('\n' + '='.repeat(50) + '\n');
    
    const adminResult = await pool.query(
      "SELECT * FROM users WHERE username = $1 OR email = $1",
      ['admin']
    );
    
    if (adminResult.rows.length > 0) {
      const admin = adminResult.rows[0];
      console.log('Admin user for comparison:');
      console.log('- ID:', admin.id);
      console.log('- Username:', admin.username);
      console.log('- Email:', admin.email);
      console.log('- Role:', admin.role);
      console.log('- Active:', admin.is_active);
      console.log('- Password Hash Length:', admin.password_hash ? admin.password_hash.length : 0);
    }
    
  } catch (error) {
    console.error('Error:', error);
  } finally {
    await pool.end();
  }
}

testQuery();