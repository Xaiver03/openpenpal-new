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
    content: `亲爱的未来的我：

当你读到这封信的时候，不知道你是否还记得写下这些文字时的心情。

现在的我，站在人生的十字路口，有些迷茫，有些不安。大学即将毕业，面临着许多选择：是继续深造还是步入社会？是留在熟悉的城市还是去远方闯荡？

我想告诉你，无论你现在身在何处，做着什么，我都希望你还保持着那份初心。希望你还记得那个夏天，我们在梧桐树下许下的诺言——要成为一个对社会有用的人，要让父母过上更好的生活，要找到真正热爱的事业。

不知道五年后的你，是否实现了这些愿望？是否找到了那个对的人？是否还会在深夜里仰望星空，思考人生的意义？

无论如何，请记得善待自己，也善待身边的人。生活或许不如意，但请保持微笑。

此致
五年前的自己
2020年6月15日`,
    description: '一封充满希望与憧憬的时光信件，记录了一个即将毕业的大学生对未来的期许。',
    tags: '时光信件,毕业季,青春,梦想'
  },
  {
    title: '第一次离家',
    content: `亲爱的爸爸妈妈：

今天是我来到大学的第三天，终于有时间给你们写信了。

宿舍的室友都很友好，有来自天南海北的同学。昨天我们一起去食堂吃饭，这里的饭菜虽然不如家里的好吃，但种类很多。我记得妈妈叮嘱我要好好吃饭，不要总是吃泡面，我都记在心里了。

学校很大，比我想象的还要大。第一天我差点迷路，还好有学长学姐帮忙带路。图书馆特别壮观，有好多层，里面的书多得数不清。我已经办好了图书证，打算这周末就去看看。

爸爸，您不用担心我的生活费，我会好好规划的。妈妈，我已经学会自己洗衣服了，虽然第一次洗得不太干净，但我会慢慢进步的。

想念家里的一切，想念妈妈做的红烧肉，想念和爸爸一起看新闻的时光。但我知道，这是成长必经的路。

等放假我就回家看你们。

爱你们的女儿
小芳`,
    description: '一封朴实而感人的家书，展现了大学新生初次离家的复杂心情。',
    tags: '家书,新生,思念,成长'
  },
  {
    title: '致我的挚友',
    content: `亲爱的小雨：

时间过得真快，转眼我们已经认识十年了。

还记得初中第一天，你主动跟坐在角落里的我打招呼，那一刻的温暖我至今难忘。从那时起，我们就成了无话不谈的好朋友。

一起走过的这些年，有太多美好的回忆：一起在图书馆熬夜复习，一起在操场上挥洒汗水，一起为了一道数学题争论不休，一起在深夜的宿舍里分享秘密...

现在我们在不同的城市读大学，见面的机会少了，但我知道我们的友谊不会因为距离而改变。每次看到有趣的事情，第一个想到的还是要分享给你。

谢谢你一直以来的陪伴和支持。在我迷茫的时候给我方向，在我失落的时候给我力量。

希望多年以后，我们还能像现在这样，做彼此最好的朋友。

永远爱你的
小月`,
    description: '真挚的友谊是青春最宝贵的财富，这封信完美诠释了这一点。',
    tags: '友谊,青春,回忆,珍贵'
  },
  {
    title: '考研路上的坚持',
    content: `未来的学弟学妹们：

当你们看到这封信的时候，或许正在经历我曾经历过的煎熬。

考研，是一条孤独而漫长的路。每天早上6点起床，晚上11点才离开图书馆，这样的日子持续了整整一年。有过崩溃，有过想要放弃，但最终还是坚持了下来。

我想告诉你们几点经验：
1. 制定合理的计划，但要留出调整的空间
2. 找到适合自己的学习方法，不要盲目跟风
3. 保持运动，身体是革命的本钱
4. 适当放松，不要把自己逼得太紧
5. 相信自己，你比想象中更强大

最重要的是，要记住你为什么出发。当你累了倦了，就想想最初的梦想。

考研不是唯一的出路，但如果你选择了这条路，就请全力以赴。无论结果如何，这段经历都会成为你人生中宝贵的财富。

加油，未来的研究生们！

一个上岸的学长`,
    description: '一封充满正能量的信，为正在考研路上奋斗的学子们带来鼓励和指引。',
    tags: '考研,励志,经验分享,坚持'
  },
  {
    title: '那年樱花树下',
    content: `致那个樱花树下的女孩：

不知道你是否还记得，三年前的春天，图书馆前的樱花树下，我们的第一次相遇。

你穿着白色的连衣裙，手里拿着一本《百年孤独》，阳光透过花瓣洒在你的发梢上。我鼓起勇气问你能否坐在旁边，你微笑着点了点头。

从那以后，我们常常在树下看书、聊天。你说你喜欢村上春树，我就把他的书都看了一遍。你说你想去看海，我们就一起规划了毕业旅行。

可是后来，我们因为一些误会渐行渐远。我想说却没说出口的话，成了心中永远的遗憾。

如今又是樱花盛开的季节，我又来到了这棵树下。花还是那样美，只是树下再也没有穿白裙子的你。

如果时光可以重来，我一定会勇敢地告诉你：我喜欢你。

祝你幸福。

一个错过的人`,
    description: '青春的遗憾总是让人唏嘘，但正是这些遗憾，让回忆变得更加珍贵。',
    tags: '爱情,遗憾,樱花,青春'
  },
  {
    title: '支教的那些日子',
    content: `亲爱的朋友们：

已经在这个小山村待了三个月了，想和你们分享一些这里的故事。

这里的孩子们真的太可爱了。虽然条件艰苦，但他们的眼睛里总是闪着光。每天早上，他们要走很远的山路来上学，却从来不迟到。上课的时候，那种求知的眼神让我特别感动。

记得有个叫小花的女孩，她的梦想是成为一名医生。她说要治好村里所有人的病。每次看到她认真做笔记的样子，我就觉得自己做的这一切都是值得的。

这里的生活确实不容易。没有网络，水电也不稳定。但是看着孩子们的笑脸，听着他们朗朗的读书声，所有的辛苦都烟消云散了。

我教他们知识，他们教会我什么是纯真和坚强。这段支教经历，将是我一生中最宝贵的回忆。

希望更多的人能够关注山区教育，让更多的孩子有机会走出大山，看看外面的世界。

爱你们的
小志愿者`,
    description: '支教不仅是知识的传递，更是心灵的交流。这封信让我们看到了教育的力量。',
    tags: '支教,公益,感动,责任'
  }
];

async function insertMuseumData() {
  const client = await pool.connect();
  
  try {
    await client.query('BEGIN');
    
    console.log('开始创建馆藏信件数据...\n');
    
    // 获取一个管理员用户作为审批人
    const adminResult = await client.query(
      "SELECT id FROM users WHERE role = 'admin' OR role = 'super_admin' LIMIT 1"
    );
    
    let adminUserId;
    if (adminResult.rows.length > 0) {
      adminUserId = adminResult.rows[0].id;
    } else {
      // 如果没有管理员，使用第一个用户
      const userResult = await client.query("SELECT id FROM users LIMIT 1");
      adminUserId = userResult.rows[0]?.id;
    }
    
    console.log(`使用管理员ID: ${adminUserId}\n`);
    
    // 先创建一些信件
    const letterIds = [];
    for (const [index, letterData] of letterContents.entries()) {
      const letterId = generateUUID();
      letterIds.push(letterId);
      
      // 插入信件
      const letterQuery = `
        INSERT INTO letters (
          id, user_id, title, content, status, is_public,
          letter_type, envelope_id, letter_code_id,
          created_at, updated_at
        ) VALUES (
          $1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW()
        )
      `;
      
      await client.query(letterQuery, [
        letterId,
        adminUserId,
        letterData.title,
        letterData.content,
        'sent',
        true,
        'standard',
        null,
        null
      ]);
      
      console.log(`✓ 创建信件: ${letterData.title}`);
      
      // 将信件添加到博物馆
      const museumItemId = generateUUID();
      const museumQuery = `
        INSERT INTO museum_items (
          id, source_type, source_id, title, description,
          tags, status, submitted_by, approved_by, approved_at,
          view_count, like_count, share_count, origin_op_code,
          created_at, updated_at
        ) VALUES (
          $1, $2, $3, $4, $5,
          $6, $7, $8, $9, $10,
          $11, $12, $13, $14,
          NOW(), NOW()
        )
      `;
      
      const viewCount = Math.floor(Math.random() * 3000) + 500;
      const likeCount = Math.floor(viewCount * (Math.random() * 0.1 + 0.05));
      const shareCount = Math.floor(likeCount * (Math.random() * 0.3 + 0.1));
      
      await client.query(museumQuery, [
        museumItemId,
        'letter',
        letterId,
        letterData.title,
        letterData.description,
        letterData.tags,
        'approved',
        adminUserId,
        adminUserId,
        new Date(),
        viewCount,
        likeCount,
        shareCount,
        null
      ]);
      
      console.log(`✓ 添加到博物馆: ${letterData.title}`);
      
      // 创建一个museum_entry
      const entryId = generateUUID();
      const entryQuery = `
        INSERT INTO museum_entries (
          id, letter_id, display_title, author_display_type,
          author_display_name, curator_type, curator_id,
          categories, tags, status, moderation_status,
          view_count, like_count, bookmark_count, share_count,
          submitted_at, approved_at, created_at, updated_at
        ) VALUES (
          $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11,
          $12, $13, $14, $15, $16, $17, NOW(), NOW()
        )
      `;
      
      const categories = ['时光信件', '校园故事', '青春记忆', '成长感悟'];
      const selectedCategories = [categories[index % categories.length]];
      
      await client.query(entryQuery, [
        entryId,
        letterId,
        letterData.title,
        'anonymous',
        `匿名用户${index + 1}`,
        'admin',
        adminUserId,
        selectedCategories,
        letterData.tags.split(','),
        'published',
        'approved',
        viewCount,
        likeCount,
        Math.floor(likeCount * 0.5),
        shareCount,
        new Date(),
        new Date()
      ]);
      
      console.log(`✓ 创建museum_entry: ${letterData.title}\n`);
    }
    
    // 创建一些展览
    console.log('\n开始创建展览...\n');
    
    const exhibitions = [
      {
        title: '时光邮局 - 写给未来的信',
        description: '收集了来自不同年份的时光信件，每一封都承载着写信人对未来的期许和想象。',
        tags: ['时光信件', '未来', '梦想']
      },
      {
        title: '青春记忆 - 校园爱情故事',
        description: '那些关于青春、关于爱情的美好与遗憾，都在这些信件中静静诉说。',
        tags: ['爱情', '青春', '校园']
      },
      {
        title: '成长足迹 - 从学生到社会人',
        description: '记录了大学生们在成长道路上的点点滴滴，从迷茫到坚定，从稚嫩到成熟。',
        tags: ['成长', '实习', '毕业']
      }
    ];
    
    for (const exhibition of exhibitions) {
      const exhibitionId = generateUUID();
      const exhibitionQuery = `
        INSERT INTO museum_exhibitions (
          id, title, description, start_date, end_date,
          status, is_featured, view_count, entry_count,
          created_at, updated_at
        ) VALUES (
          $1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW()
        )
      `;
      
      await client.query(exhibitionQuery, [
        exhibitionId,
        exhibition.title,
        exhibition.description,
        new Date('2024-01-01'),
        new Date('2024-12-31'),
        'active',
        true,
        Math.floor(Math.random() * 5000) + 1000,
        Math.floor(letterIds.length / 3)
      ]);
      
      console.log(`✓ 创建展览: ${exhibition.title}`);
      
      // 将部分信件添加到展览中
      const numEntries = Math.floor(letterIds.length / 3);
      for (let i = 0; i < numEntries; i++) {
        const exhibitionEntryQuery = `
          INSERT INTO museum_exhibition_entries (
            id, exhibition_id, entry_id, display_order,
            created_at, updated_at
          ) VALUES (
            $1, $2, $3, $4, NOW(), NOW()
          )
        `;
        
        // 获取对应的museum_entry
        const entryResult = await client.query(
          "SELECT id FROM museum_entries WHERE letter_id = $1 LIMIT 1",
          [letterIds[i]]
        );
        
        if (entryResult.rows.length > 0) {
          await client.query(exhibitionEntryQuery, [
            generateUUID(),
            exhibitionId,
            entryResult.rows[0].id,
            i + 1
          ]);
        }
      }
    }
    
    await client.query('COMMIT');
    console.log('\n✅ 所有数据插入成功！');
    
    // 查询统计
    const statsQuery = `
      SELECT 
        (SELECT COUNT(*) FROM letters WHERE id IN (${letterIds.map((_, i) => `$${i + 1}`).join(',')})) as letter_count,
        (SELECT COUNT(*) FROM museum_items WHERE source_type = 'letter') as museum_items_count,
        (SELECT COUNT(*) FROM museum_entries) as museum_entries_count,
        (SELECT COUNT(*) FROM museum_exhibitions WHERE status = 'active') as exhibition_count
    `;
    
    const stats = await client.query(statsQuery, letterIds);
    console.log('\n📊 数据统计:');
    console.log(`   信件数量: ${stats.rows[0].letter_count}`);
    console.log(`   博物馆物品: ${stats.rows[0].museum_items_count}`);
    console.log(`   博物馆条目: ${stats.rows[0].museum_entries_count}`);
    console.log(`   活跃展览: ${stats.rows[0].exhibition_count}`);
    
  } catch (error) {
    await client.query('ROLLBACK');
    console.error('\n❌ 插入数据时发生错误:', error);
    throw error;
  } finally {
    client.release();
    await pool.end();
  }
}

// 执行插入
insertMuseumData()
  .then(() => {
    console.log('\n🎉 馆藏信件Mock数据创建完成！');
    console.log('您现在可以访问 http://localhost:3000 查看馆藏信件了。');
    process.exit(0);
  })
  .catch((error) => {
    console.error('\n💥 创建失败:', error.message);
    process.exit(1);
  });