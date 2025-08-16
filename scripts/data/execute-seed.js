const { Client } = require('pg');
const fs = require('fs');

async function executeSeedData() {
  const client = new Client({
    connectionString: process.env.DATABASE_URL || 'postgres://rocalight:@localhost:5432/openpenpal'
  });
  
  try {
    await client.connect();
    console.log('📡 连接数据库成功');
    
    const seedSQL = fs.readFileSync('backend/seeds/comprehensive_promotion_seed.sql', 'utf8');
    
    // 分割SQL语句并执行
    const statements = seedSQL.split(';').filter(stmt => stmt.trim());
    
    for (const statement of statements) {
      if (statement.trim()) {
        try {
          const result = await client.query(statement);
          if (result.rows && result.rows.length > 0) {
            console.log('✅', result.rows[0]);
          }
        } catch (error) {
          if (!error.message.includes('already exists') && !error.message.includes('duplicate key')) {
            console.log('⚠️', error.message.substring(0, 100));
          }
        }
      }
    }
    
    console.log('🎉 种子数据执行完成');
    
  } catch (error) {
    console.error('❌ 执行失败:', error.message);
  } finally {
    await client.end();
  }
}

executeSeedData();