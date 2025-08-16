const { Client } = require('pg');
const crypto = require('crypto');

// Database connection configuration
const config = {
  host: 'localhost',
  port: 5432,
  database: 'openpenpal',
  user: process.env.DATABASE_USER || process.env.USER || 'postgres',
  password: process.env.DATABASE_PASSWORD || 'password',
};

// Test courier users data (matching the Go seed data)
const courierUsers = [
  {
    id: 'courier-level1',
    username: 'courier_level1',
    email: 'courier1@openpenpal.com',
    password_hash: '$2a$10$N9qo8uLOickgx2ZMRZoMye.Lf/rSDRfYHYxX1dpIjJJNzTHFN1UTO', // secret
    nickname: '一级信使',
    role: 'courier_level1',
    school_code: 'PKU001',
    is_active: true
  },
  {
    id: 'courier-level2',
    username: 'courier_level2',
    email: 'courier2@openpenpal.com',
    password_hash: '$2a$10$N9qo8uLOickgx2ZMRZoMye.Lf/rSDRfYHYxX1dpIjJJNzTHFN1UTO', // secret
    nickname: '二级信使',
    role: 'courier_level2',
    school_code: 'PKU001',
    is_active: true
  },
  {
    id: 'courier-level3',
    username: 'courier_level3',
    email: 'courier3@openpenpal.com',
    password_hash: '$2a$10$N9qo8uLOickgx2ZMRZoMye.Lf/rSDRfYHYxX1dpIjJJNzTHFN1UTO', // secret
    nickname: '三级信使',
    role: 'courier_level3',
    school_code: 'PKU001',
    is_active: true
  },
  {
    id: 'courier-level4',
    username: 'courier_level4',
    email: 'courier4@openpenpal.com',
    password_hash: '$2a$10$N9qo8uLOickgx2ZMRZoMye.Lf/rSDRfYHYxX1dpIjJJNzTHFN1UTO', // secret
    nickname: '四级信使',
    role: 'courier_level4',
    school_code: 'PKU001',
    is_active: true
  }
];

async function seedCourierUsers() {
  const client = new Client(config);
  
  try {
    await client.connect();
    console.log('✅ Connected to PostgreSQL');
    
    // Check if users table exists
    const tableResult = await client.query(`
      SELECT EXISTS (
        SELECT FROM information_schema.tables 
        WHERE table_schema = 'public' 
        AND table_name = 'users'
      );
    `);
    
    if (!tableResult.rows[0].exists) {
      console.log('❌ Users table does not exist. Please run migrations first.');
      return;
    }
    
    // Seed each courier user
    for (const user of courierUsers) {
      try {
        // Check if user already exists
        const existingUser = await client.query(
          'SELECT id FROM users WHERE username = $1',
          [user.username]
        );
        
        if (existingUser.rows.length > 0) {
          console.log(`⏭️  User ${user.username} already exists, skipping...`);
          continue;
        }
        
        // Insert user
        await client.query(`
          INSERT INTO users (
            id, username, email, password_hash, nickname, 
            role, school_code, is_active, created_at, updated_at
          ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
        `, [
          user.id,
          user.username,
          user.email,
          user.password_hash,
          user.nickname,
          user.role,
          user.school_code,
          user.is_active
        ]);
        
        console.log(`✅ Created user: ${user.username} (${user.nickname})`);
        
      } catch (error) {
        console.error(`❌ Error creating user ${user.username}:`, error.message);
      }
    }
    
    // Verify all courier users
    const courierResult = await client.query(`
      SELECT username, email, nickname, role, school_code 
      FROM users 
      WHERE username LIKE 'courier_level%' 
      ORDER BY username
    `);
    
    console.log('\n📋 Final courier user list:');
    courierResult.rows.forEach((user, index) => {
      console.log(`${index + 1}. ${user.username} (${user.role}) - ${user.nickname}`);
    });
    
    // Also create the admin user if not exists
    const adminUser = {
      id: 'test-admin',
      username: 'admin',
      email: 'admin@penpal.com',
      password_hash: '$2a$10$dwSXE/fBcbAJVy0jMZHYI.vFjjUZFYRMPpeAzcgmHd.XqwfqgOrEW', // admin123
      nickname: '系统管理员',
      role: 'super_admin',
      school_code: 'SYSTEM',
      is_active: true
    };
    
    const existingAdmin = await client.query(
      'SELECT id FROM users WHERE username = $1',
      [adminUser.username]
    );
    
    if (existingAdmin.rows.length === 0) {
      await client.query(`
        INSERT INTO users (
          id, username, email, password_hash, nickname, 
          role, school_code, is_active, created_at, updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
      `, [
        adminUser.id,
        adminUser.username,
        adminUser.email,
        adminUser.password_hash,
        adminUser.nickname,
        adminUser.role,
        adminUser.school_code,
        adminUser.is_active
      ]);
      console.log(`\n✅ Created admin user: ${adminUser.username}`);
    }
    
  } catch (error) {
    console.error('❌ Database error:', error);
  } finally {
    await client.end();
    console.log('\n✅ Database seeding completed');
  }
}

// Run the seeding
console.log('🌱 Starting courier user seeding...\n');
seedCourierUsers();