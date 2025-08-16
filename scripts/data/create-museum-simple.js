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

// Mock信件数据
const letterContents = [
  {
    title: '致未来的自己',
    content: `亲爱的未来的我：\n\n当你读到这封信的时候，不知道你是否还记得写下这些文字时的心情。\n\n现在的我，站在人生的十字路口，有些迷茫，有些不安。大学即将毕业，面临着许多选择。\n\n我想告诉你，无论你现在身在何处，做着什么，我都希望你还保持着那份初心。\n\n无论如何，请记得善待自己，也善待身边的人。生活或许不如意，但请保持微笑。\n\n此致\n五年前的自己`,
    description: '一封充满希望与憧憬的时光信件',
    tags: '时光信件,毕业季'
  },
  {
    title: '第一次离家',
    content: `亲爱的爸爸妈妈：\n\n今天是我来到大学的第三天，终于有时间给你们写信了。\n\n宿舍的室友都很友好，有来自天南海北的同学。学校很大，比我想象的还要大。\n\n想念家里的一切，想念妈妈做的红烧肉，想念和爸爸一起看新闻的时光。\n\n等放假我就回家看你们。\n\n爱你们的女儿`,
    description: '一封朴实而感人的家书',
    tags: '家书,新生'
  },
  {
    title: '考研路上的坚持',
    content: `未来的学弟学妹们：\n\n当你们看到这封信的时候，或许正在经历我曾经历过的煎熬。\n\n考研，是一条孤独而漫长的路。但请相信自己，你比想象中更强大。\n\n加油，未来的研究生们！\n\n一个上岸的学长`,
    description: '一封充满正能量的鼓励信',
    tags: '考研,励志'
  }
];

async function insertMuseumData() {
  const client = await pool.connect();
  
  try {
    console.log('开始创建馆藏信件数据...\n');
    
    // 获取一个用户ID
    const userResult = await client.query("SELECT id FROM users LIMIT 1");
    if (userResult.rows.length === 0) {
      throw new Error('数据库中没有用户，请先创建用户');
    }
    const userId = userResult.rows[0].id;
    console.log(`使用用户ID: ${userId}\n`);
    
    let successCount = 0;
    
    // 逐个插入数据，避免事务回滚影响所有数据
    for (const letterData of letterContents) {
      try {
        await client.query('BEGIN');
        
        const letterId = generateUUID();
        
        // 插入信件
        const letterQuery = `
          INSERT INTO letters (
            id, user_id, author_id, title, content, 
            style, status, visibility,
            created_at, updated_at
          ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW()
          )
        `;
        
        await client.query(letterQuery, [
          letterId,
          userId,
          userId,
          letterData.title,
          letterData.content,
          'classic',
          'sent',
          'public'
        ]);
        
        console.log(`✓ 创建信件: ${letterData.title}`);
        
        // 添加到博物馆
        const museumItemId = generateUUID();
        const museumQuery = `
          INSERT INTO museum_items (
            id, source_type, source_id, title, description,
            tags, status, submitted_by, approved_by, approved_at,
            view_count, like_count, share_count,
            created_at, updated_at
          ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
            $11, $12, $13, NOW(), NOW()
          )
        `;
        
        const viewCount = Math.floor(Math.random() * 2000) + 100;
        const likeCount = Math.floor(viewCount * 0.1);
        
        await client.query(museumQuery, [
          museumItemId,
          'letter',
          letterId,
          letterData.title,
          letterData.description,
          letterData.tags,
          'approved',
          userId,
          userId,
          new Date(),
          viewCount,
          likeCount,
          Math.floor(likeCount * 0.3)
        ]);
        
        console.log(`✓ 添加到博物馆: ${letterData.title}\n`);
        
        await client.query('COMMIT');
        successCount++;
        
      } catch (error) {
        await client.query('ROLLBACK');
        console.error(`❌ 处理 "${letterData.title}" 时出错:`, error.message);
      }
    }
    
    console.log(`\n✅ 成功创建 ${successCount}/${letterContents.length} 条数据！`);
    
    // 查询统计
    const stats = await client.query(`
      SELECT 
        (SELECT COUNT(*) FROM museum_items WHERE source_type = 'letter' AND status = 'approved') as count
    `);
    
    console.log(`\n📊 当前博物馆中共有 ${stats.rows[0].count} 封信件`);
    
  } catch (error) {
    console.error('\n❌ 发生错误:', error.message);
  } finally {
    client.release();
    await pool.end();
  }
}

// 执行插入
insertMuseumData()
  .then(() => {
    console.log('\n🎉 操作完成！');
    console.log('您现在可以访问 http://localhost:3000 查看馆藏信件了。');
    process.exit(0);
  })
  .catch((error) => {
    console.error('\n💥 失败:', error.message);
    process.exit(1);
  });