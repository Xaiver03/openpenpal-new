const { Client } = require('pg');

// Database connection configuration
const config = {
  host: 'localhost',
  port: 5432,
  database: 'openpenpal',
  user: process.env.DATABASE_USER || process.env.USER || 'postgres',
  password: process.env.DATABASE_PASSWORD || 'password',
};

console.log('Testing PostgreSQL connection with config:', {
  ...config,
  password: '***' // Hide password in logs
});

async function testConnection() {
  const client = new Client(config);
  
  try {
    // Connect to database
    await client.connect();
    console.log('✅ Successfully connected to PostgreSQL');
    
    // Test database name
    const dbResult = await client.query('SELECT current_database()');
    console.log(`📊 Connected to database: ${dbResult.rows[0].current_database}`);
    
    // Check if users table exists
    const tableResult = await client.query(`
      SELECT EXISTS (
        SELECT FROM information_schema.tables 
        WHERE table_schema = 'public' 
        AND table_name = 'users'
      );
    `);
    
    if (!tableResult.rows[0].exists) {
      console.log('⚠️  Users table does not exist. Please run migrations first.');
      return;
    }
    
    console.log('✅ Users table exists');
    
    // Count total users
    const countResult = await client.query('SELECT COUNT(*) FROM users');
    console.log(`👥 Total users in database: ${countResult.rows[0].count}`);
    
    // Check for courier_level1 user specifically
    const courierResult = await client.query(`
      SELECT id, username, email, nickname, role, school_code, is_active 
      FROM users 
      WHERE username = 'courier_level1'
    `);
    
    if (courierResult.rows.length > 0) {
      console.log('\n✅ Found courier_level1 user:');
      console.log(JSON.stringify(courierResult.rows[0], null, 2));
    } else {
      console.log('\n❌ courier_level1 user not found in database');
    }
    
    // List all test courier users
    const courierListResult = await client.query(`
      SELECT username, email, nickname, role, school_code 
      FROM users 
      WHERE username LIKE 'courier%' 
      ORDER BY username
    `);
    
    if (courierListResult.rows.length > 0) {
      console.log('\n📋 All courier test users:');
      courierListResult.rows.forEach((user, index) => {
        console.log(`${index + 1}. ${user.username} (${user.role}) - ${user.nickname}`);
      });
    }
    
    // Check password hash for courier_level1
    const passwordResult = await client.query(`
      SELECT username, password_hash 
      FROM users 
      WHERE username = 'courier_level1'
    `);
    
    if (passwordResult.rows.length > 0) {
      console.log('\n🔐 Password hash verification:');
      console.log(`Username: ${passwordResult.rows[0].username}`);
      console.log(`Has password hash: ${passwordResult.rows[0].password_hash ? 'Yes' : 'No'}`);
      // The expected hash for "secret" is: $2a$10$N9qo8uLOickgx2ZMRZoMye.Lf/rSDRfYHYxX1dpIjJJNzTHFN1UTO
      const expectedHash = '$2a$10$N9qo8uLOickgx2ZMRZoMye.Lf/rSDRfYHYxX1dpIjJJNzTHFN1UTO';
      if (passwordResult.rows[0].password_hash === expectedHash) {
        console.log('✅ Password hash matches expected value for "secret"');
      } else {
        console.log('⚠️  Password hash does not match expected value');
      }
    }
    
  } catch (error) {
    console.error('❌ Database connection error:', error.message);
    console.error('Error details:', error);
  } finally {
    await client.end();
    console.log('\n🔌 Database connection closed');
  }
}

// Run the test
testConnection();