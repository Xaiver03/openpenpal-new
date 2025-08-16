const { Client } = require('pg');

async function fixPromotionHistoryTable() {
  const client = new Client({
    connectionString: process.env.DATABASE_URL || 'postgres://rocalight:@localhost:5432/openpenpal'
  });
  
  try {
    await client.connect();
    
    // 检查courier_promotion_history表结构
    const columns = await client.query(`
      SELECT column_name, data_type, is_nullable 
      FROM information_schema.columns 
      WHERE table_name = 'courier_promotion_history'
      ORDER BY ordinal_position;
    `);
    
    console.log('📋 courier_promotion_history表结构:');
    columns.rows.forEach(col => {
      console.log(`  - ${col.column_name}: ${col.data_type} (${col.is_nullable})`);
    });
    
    // 检查是否需要添加字段
    const hasReason = columns.rows.some(col => col.column_name === 'reason');
    const hasEvidence = columns.rows.some(col => col.column_name === 'evidence');
    
    if (!hasReason || !hasEvidence) {
      console.log('\n需要添加缺失的字段...');
      
      if (!hasReason) {
        await client.query('ALTER TABLE courier_promotion_history ADD COLUMN reason TEXT');
        console.log('✅ 添加reason字段成功');
      }
      
      if (!hasEvidence) {
        await client.query('ALTER TABLE courier_promotion_history ADD COLUMN evidence JSONB');
        console.log('✅ 添加evidence字段成功');
      }
    } else {
      console.log('✅ 表结构完整');
    }
    
  } catch (error) {
    console.error('❌ 修复失败:', error.message);
  } finally {
    await client.end();
  }
}

fixPromotionHistoryTable();