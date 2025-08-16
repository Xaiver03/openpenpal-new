const { Pool } = require('pg');
const crypto = require('crypto');

// 生成UUID
function generateUUID() {
  return crypto.randomUUID();
}

// 数据库连接配置
const pool = new Pool({
  host: 'localhost',
  port: 5432,
  database: 'openpenpal',
  user: process.env.USER || 'rocalight',
  password: 'password'
});

async function createMuseumEntries() {
  const client = await pool.connect();
  
  try {
    console.log('开始创建museum_entries数据...\n');
    
    // 获取已存在的信件
    const lettersResult = await client.query(`
      SELECT l.id, l.title, l.content, l.user_id 
      FROM letters l 
      WHERE l.visibility = 'public' 
      LIMIT 10
    `);
    
    if (lettersResult.rows.length === 0) {
      console.log('没有找到公开的信件');
      return;
    }
    
    console.log(`找到 ${lettersResult.rows.length} 封公开信件\n`);
    
    let successCount = 0;
    const categories = ['时光信件', '校园故事', '青春记忆', '成长感悟'];
    const tags = [
      ['时光信件', '毕业季', '青春', '梦想'],
      ['家书', '新生', '思念', '成长'],
      ['友谊', '青春', '回忆', '珍贵'],
      ['考研', '励志', '经验分享', '坚持'],
      ['爱情', '遗憾', '樱花', '青春'],
      ['支教', '公益', '感动', '责任']
    ];
    
    for (const [index, letter] of lettersResult.rows.entries()) {
      try {
        await client.query('BEGIN');
        
        const entryId = generateUUID();
        const selectedCategories = [categories[index % categories.length]];
        const selectedTags = tags[index % tags.length];
        const viewCount = Math.floor(Math.random() * 3000) + 500;
        const likeCount = Math.floor(viewCount * 0.1);
        
        const query = `
          INSERT INTO museum_entries (
            id, letter_id, display_title, 
            author_display_type, author_display_name,
            curator_type, curator_id,
            categories, tags, 
            status, moderation_status,
            view_count, like_count, bookmark_count, share_count,
            submitted_at, approved_at, 
            created_at, updated_at
          ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11,
            $12, $13, $14, $15, $16, $17, NOW(), NOW()
          )
        `;
        
        await client.query(query, [
          entryId,
          letter.id,
          letter.title,
          'anonymous',
          `匿名作者${index + 1}`,
          'user',
          letter.user_id,
          selectedCategories,
          selectedTags,
          'published',
          'approved',
          viewCount,
          likeCount,
          Math.floor(likeCount * 0.5),
          Math.floor(likeCount * 0.3),
          new Date(Date.now() - Math.random() * 30 * 24 * 60 * 60 * 1000), // 随机30天内
          new Date(Date.now() - Math.random() * 20 * 24 * 60 * 60 * 1000), // 随机20天内
        ]);
        
        console.log(`✓ 创建museum_entry: ${letter.title}`);
        
        await client.query('COMMIT');
        successCount++;
        
      } catch (error) {
        await client.query('ROLLBACK');
        console.error(`❌ 处理信件 "${letter.title}" 时出错:`, error.message);
      }
    }
    
    console.log(`\n✅ 成功创建 ${successCount} 条museum_entries！`);
    
    // 查询统计
    const stats = await client.query(`
      SELECT 
        COUNT(*) as total_entries,
        COUNT(DISTINCT letter_id) as unique_letters,
        SUM(view_count) as total_views,
        SUM(like_count) as total_likes
      FROM museum_entries 
      WHERE status = 'published'
    `);
    
    console.log('\n📊 博物馆统计:');
    console.log(`   总条目数: ${stats.rows[0].total_entries}`);
    console.log(`   信件数量: ${stats.rows[0].unique_letters}`);
    console.log(`   总浏览量: ${stats.rows[0].total_views || 0}`);
    console.log(`   总点赞数: ${stats.rows[0].total_likes || 0}`);
    
  } catch (error) {
    console.error('\n❌ 发生错误:', error.message);
  } finally {
    client.release();
    await pool.end();
  }
}

// 执行
createMuseumEntries()
  .then(() => {
    console.log('\n🎉 操作完成！');
    console.log('您现在可以访问 http://localhost:3000 查看馆藏信件了。');
    process.exit(0);
  })
  .catch((error) => {
    console.error('\n💥 失败:', error.message);
    process.exit(1);
  });