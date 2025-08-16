const bcrypt = require('bcryptjs');
const { Client } = require('pg');

async function fixPassword() {
  const client = new Client({
    connectionString: process.env.DATABASE_URL || 'postgres://rocalight:@localhost:5432/openpenpal'
  });
  
  try {
    await client.connect();
    
    // 生成新的密码哈希
    const password = 'secret';
    const newHash = await bcrypt.hash(password, 10);
    console.log('生成新密码哈希:', newHash);
    
    // 更新courier_level3的密码
    await client.query('UPDATE users SET password_hash = $1 WHERE username = $2', [newHash, 'courier_level3']);
    console.log('✅ courier_level3密码已更新');
    
    // 验证密码
    const user = await client.query('SELECT password_hash FROM users WHERE username = $1', ['courier_level3']);
    const storedHash = user.rows[0].password_hash;
    const isValid = await bcrypt.compare(password, storedHash);
    console.log('密码验证结果:', isValid);
    
  } catch (error) {
    console.error('❌ 修复密码失败:', error.message);
  } finally {
    await client.end();
  }
}

fixPassword();