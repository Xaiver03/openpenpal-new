const { Pool } = require('pg');

// 数据库连接配置
const pool = new Pool({
  host: 'localhost',
  port: 5432,
  database: 'openpenpal',
  user: process.env.USER || 'rocalight',
  password: 'password'
});

// Mock数据
const museumEntries = [
  {
    letter_id: null,
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
    author: '小明',
    author_info: '2020届毕业生，计算机系',
    submission_date: new Date('2020-06-15'),
    display_date: '2020年夏',
    location: '北京大学',
    tags: ['时光信件', '毕业季', '青春', '梦想'],
    category: 'time_capsule',
    status: 'published',
    view_count: 1523,
    like_count: 89,
    is_featured: true,
    curator_notes: '这是一封充满希望与憧憬的时光信件，记录了一个即将毕业的大学生对未来的期许。',
    metadata: {
      original_format: 'handwritten',
      paper_type: '信纸',
      preservation_status: 'excellent',
      historical_context: '2020年疫情期间的毕业季'
    }
  },
  {
    letter_id: null,
    title: '第一次离家',
    content: `亲爱的爸爸妈妈：

今天是我来到大学的第三天，终于有时间给你们写信了。

宿舍的室友都很友好，有来自天南海北的同学。昨天我们一起去食堂吃饭，这里的饭菜虽然不如家里的好吃，但种类很多。我记得妈妈叮嘱我要好好吃饭，不要总是吃泡面，我都记在心里了。

学校很大，比我想象的还要大。第一天我差点迷路，还好有学长学姐帮忙带路。图书馆特别壮观，有好多层，里面的书多得数不清。我已经办好了图书证，打算这周末就去看看。

爸爸，您不用担心我的生活费，我会好好规划的。妈妈，我已经学会自己洗衣服了，虽然第一次洗得不太干净，但我会慢慢进步的。

想念家里的一切，想念妈妈做的红烧肉，想念和爸爸一起看新闻的时光。但我知道，这是成长必经的路。

等放假我就回家看你们。

爱你们的女儿
小芳
2019年9月5日`,
    author: '李芳',
    author_info: '2019级新生，文学院',
    submission_date: new Date('2019-09-05'),
    display_date: '2019年秋',
    location: '清华大学',
    tags: ['家书', '新生', '思念', '成长'],
    category: 'family',
    status: 'published',
    view_count: 2341,
    like_count: 156,
    is_featured: true,
    curator_notes: '一封朴实而感人的家书，展现了大学新生初次离家的复杂心情。',
    metadata: {
      original_format: 'handwritten',
      paper_type: '普通信纸',
      emotional_tone: 'nostalgic',
      recipient: '父母'
    }
  },
  {
    letter_id: null,
    title: '致我的挚友',
    content: `亲爱的小雨：

时间过得真快，转眼我们已经认识十年了。

还记得初中第一天，你主动跟坐在角落里的我打招呼，那一刻的温暖我至今难忘。从那时起，我们就成了无话不谈的好朋友。

一起走过的这些年，有太多美好的回忆：一起在图书馆熬夜复习，一起在操场上挥洒汗水，一起为了一道数学题争论不休，一起在深夜的宿舍里分享秘密...

现在我们在不同的城市读大学，见面的机会少了，但我知道我们的友谊不会因为距离而改变。每次看到有趣的事情，第一个想到的还是要分享给你。

谢谢你一直以来的陪伴和支持。在我迷茫的时候给我方向，在我失落的时候给我力量。

希望多年以后，我们还能像现在这样，做彼此最好的朋友。

永远爱你的
小月
2021年3月20日`,
    author: '王月',
    author_info: '2018级学生，新闻系',
    submission_date: new Date('2021-03-20'),
    display_date: '2021年春',
    location: '复旦大学',
    tags: ['友谊', '青春', '回忆', '珍贵'],
    category: 'friendship',
    status: 'published',
    view_count: 1876,
    like_count: 124,
    is_featured: false,
    curator_notes: '真挚的友谊是青春最宝贵的财富，这封信完美诠释了这一点。',
    metadata: {
      writing_style: 'casual',
      mood: 'grateful',
      special_occasion: '友谊纪念日'
    }
  },
  {
    letter_id: null,
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

一个上岸的学长
2022年4月10日`,
    author: '张强',
    author_info: '2022届毕业生，成功考取北大研究生',
    submission_date: new Date('2022-04-10'),
    display_date: '2022年春',
    location: '北京理工大学',
    tags: ['考研', '励志', '经验分享', '坚持'],
    category: 'inspiration',
    status: 'published',
    view_count: 3456,
    like_count: 234,
    is_featured: true,
    curator_notes: '一封充满正能量的信，为正在考研路上奋斗的学子们带来鼓励和指引。',
    metadata: {
      target_audience: '考研学生',
      writing_purpose: '经验分享与鼓励',
      achievement: '成功考取北京大学'
    }
  },
  {
    letter_id: null,
    title: '那年樱花树下',
    content: `致那个樱花树下的女孩：

不知道你是否还记得，三年前的春天，图书馆前的樱花树下，我们的第一次相遇。

你穿着白色的连衣裙，手里拿着一本《百年孤独》，阳光透过花瓣洒在你的发梢上。我鼓起勇气问你能否坐在旁边，你微笑着点了点头。

从那以后，我们常常在树下看书、聊天。你说你喜欢村上春树，我就把他的书都看了一遍。你说你想去看海，我们就一起规划了毕业旅行。

可是后来，我们因为一些误会渐行渐远。我想说却没说出口的话，成了心中永远的遗憾。

如今又是樱花盛开的季节，我又来到了这棵树下。花还是那样美，只是树下再也没有穿白裙子的你。

如果时光可以重来，我一定会勇敢地告诉你：我喜欢你。

祝你幸福。

一个错过的人
2023年4月1日`,
    author: '匿名',
    author_info: '2020级学生',
    submission_date: new Date('2023-04-01'),
    display_date: '2023年春',
    location: '武汉大学',
    tags: ['爱情', '遗憾', '樱花', '青春'],
    category: 'love',
    status: 'published',
    view_count: 4521,
    like_count: 367,
    is_featured: true,
    curator_notes: '青春的遗憾总是让人唏嘘，但正是这些遗憾，让回忆变得更加珍贵。',
    metadata: {
      anonymous: true,
      season: '樱花季',
      emotional_impact: 'high',
      writing_style: 'lyrical'
    }
  },
  {
    letter_id: null,
    title: '实习的第一天',
    content: `亲爱的日记：

今天是我实习的第一天，心情既紧张又兴奋。

早上7点就起床了，特意穿上了新买的职业装。到公司的时候还早了半个小时，在楼下的咖啡店坐了一会儿，看着来来往往的上班族，突然意识到自己也要成为他们中的一员了。

导师人很好，耐心地给我介绍了公司的情况和我接下来要做的工作。虽然现在还有很多不懂的地方，但我相信通过努力一定能够胜任。

中午和几个同事一起吃饭，他们都很友善，还给了我很多建议。有个姐姐告诉我，实习期间最重要的是多学多问，不要怕犯错。

下班的时候，走在熙熙攘攘的街道上，感觉自己真的长大了。从学生到职场人的转变，虽然充满挑战，但我已经准备好了。

明天要早点睡，养足精神迎接新的一天！

加油，自己！
2023年7月3日`,
    author: '陈晓',
    author_info: '2024届学生，市场营销专业',
    submission_date: new Date('2023-07-03'),
    display_date: '2023年夏',
    location: '上海交通大学',
    tags: ['实习', '成长', '职场', '新开始'],
    category: 'growth',
    status: 'published',
    view_count: 1234,
    like_count: 78,
    is_featured: false,
    curator_notes: '从校园到职场的转变是人生重要的里程碑，这封信记录了这个特殊时刻的心情。',
    metadata: {
      milestone: '第一份实习',
      company_type: '互联网公司',
      diary_style: true
    }
  },
  {
    letter_id: null,
    title: '支教的那些日子',
    content: `亲爱的朋友们：

已经在这个小山村待了三个月了，想和你们分享一些这里的故事。

这里的孩子们真的太可爱了。虽然条件艰苦，但他们的眼睛里总是闪着光。每天早上，他们要走很远的山路来上学，却从来不迟到。上课的时候，那种求知的眼神让我特别感动。

记得有个叫小花的女孩，她的梦想是成为一名医生。她说要治好村里所有人的病。每次看到她认真做笔记的样子，我就觉得自己做的这一切都是值得的。

这里的生活确实不容易。没有网络，水电也不稳定。但是看着孩子们的笑脸，听着他们朗朗的读书声，所有的辛苦都烟消云散了。

我教他们知识，他们教会我什么是纯真和坚强。这段支教经历，将是我一生中最宝贵的回忆。

希望更多的人能够关注山区教育，让更多的孩子有机会走出大山，看看外面的世界。

爱你们的
小志愿者
2022年11月15日`,
    author: '林悦',
    author_info: '2021级学生，教育学院',
    submission_date: new Date('2022-11-15'),
    display_date: '2022年冬',
    location: '北京师范大学',
    tags: ['支教', '公益', '感动', '责任'],
    category: 'volunteer',
    status: 'published',
    view_count: 2897,
    like_count: 189,
    is_featured: true,
    curator_notes: '支教不仅是知识的传递，更是心灵的交流。这封信让我们看到了教育的力量。',
    metadata: {
      location_detail: '云南山区',
      duration: '6个月',
      impact: '帮助30多名学生',
      volunteer_program: '大学生西部支教计划'
    }
  },
  {
    letter_id: null,
    title: '毕业典礼上想说的话',
    content: `亲爱的母校：

四年时光匆匆而过，今天终于要说再见了。

还记得第一次踏进校门时的懵懂，第一次熬夜准备考试的紧张，第一次上台演讲的忐忑，第一次拿到奖学金的喜悦...太多的第一次，都发生在这里。

感谢每一位老师的谆谆教诲。是您们不仅传授了知识，更是教会了我们如何做人做事。特别是我的导师，您的人格魅力和学术精神将永远激励着我。

感谢亲爱的同学们。我们一起哭过笑过，一起奋斗过迷茫过。那些深夜在宿舍的卧谈，那些一起刷夜的日子，都是最珍贵的记忆。

感谢美丽的校园。梧桐大道、图书馆、体育场...每一个角落都有我们的足迹。春天的樱花，夏天的梧桐，秋天的银杏，冬天的雪景，四季更替见证了我们的成长。

今天我们毕业了，但这不是结束，而是新的开始。无论走到哪里，我们都会记得，我们是这里的学生，我们要为母校争光。

再见了，我的大学！
谢谢你，给了我最美好的四年！

2023届全体毕业生
2023年6月28日`,
    author: '学生会',
    author_info: '2023届毕业生代表',
    submission_date: new Date('2023-06-28'),
    display_date: '2023年夏',
    location: '中国人民大学',
    tags: ['毕业', '感恩', '告别', '青春'],
    category: 'graduation',
    status: 'published',
    view_count: 5678,
    like_count: 456,
    is_featured: true,
    curator_notes: '每一个毕业季都有相似的不舍，但每一份不舍都是独一无二的青春记忆。',
    metadata: {
      event: '2023届毕业典礼',
      collective_letter: true,
      signatures: 1200,
      ceremony_date: '2023-06-28'
    }
  }
];

async function insertMuseumData() {
  const client = await pool.connect();
  
  try {
    await client.query('BEGIN');
    
    console.log('开始插入馆藏信件数据...');
    
    for (const entry of museumEntries) {
      const query = `
        INSERT INTO museum_entries (
          letter_id, title, content, author, author_info,
          submission_date, display_date, location, tags, category,
          status, view_count, like_count, is_featured, curator_notes,
          metadata, created_at, updated_at
        ) VALUES (
          $1, $2, $3, $4, $5,
          $6, $7, $8, $9, $10,
          $11, $12, $13, $14, $15,
          $16, NOW(), NOW()
        )
      `;
      
      const values = [
        entry.letter_id,
        entry.title,
        entry.content,
        entry.author,
        entry.author_info,
        entry.submission_date,
        entry.display_date,
        entry.location,
        JSON.stringify(entry.tags),
        entry.category,
        entry.status,
        entry.view_count,
        entry.like_count,
        entry.is_featured,
        entry.curator_notes,
        JSON.stringify(entry.metadata)
      ];
      
      await client.query(query, values);
      console.log(`✓ 插入信件: ${entry.title}`);
    }
    
    // 创建一些展览
    const exhibitions = [
      {
        title: '时光邮局 - 写给未来的信',
        description: '收集了来自不同年份的时光信件，每一封都承载着写信人对未来的期许和想象。',
        start_date: new Date('2024-01-01'),
        end_date: new Date('2024-12-31'),
        status: 'active',
        curator: '博物馆策展团队',
        tags: JSON.stringify(['时光信件', '未来', '梦想'])
      },
      {
        title: '青春记忆 - 校园爱情故事',
        description: '那些关于青春、关于爱情的美好与遗憾，都在这些信件中静静诉说。',
        start_date: new Date('2024-02-14'),
        end_date: new Date('2024-05-20'),
        status: 'active',
        curator: '博物馆策展团队',
        tags: JSON.stringify(['爱情', '青春', '校园'])
      },
      {
        title: '成长足迹 - 从学生到社会人',
        description: '记录了大学生们在成长道路上的点点滴滴，从迷茫到坚定，从稚嫩到成熟。',
        start_date: new Date('2024-06-01'),
        end_date: new Date('2024-08-31'),
        status: 'active',
        curator: '博物馆策展团队',
        tags: JSON.stringify(['成长', '实习', '毕业'])
      }
    ];
    
    console.log('\n开始创建展览...');
    
    for (const exhibition of exhibitions) {
      const query = `
        INSERT INTO museum_exhibitions (
          title, description, start_date, end_date,
          status, curator, tags, created_at, updated_at
        ) VALUES (
          $1, $2, $3, $4, $5, $6, $7, NOW(), NOW()
        )
      `;
      
      const values = [
        exhibition.title,
        exhibition.description,
        exhibition.start_date,
        exhibition.end_date,
        exhibition.status,
        exhibition.curator,
        exhibition.tags
      ];
      
      await client.query(query, values);
      console.log(`✓ 创建展览: ${exhibition.title}`);
    }
    
    await client.query('COMMIT');
    console.log('\n✅ 所有数据插入成功！');
    
  } catch (error) {
    await client.query('ROLLBACK');
    console.error('❌ 插入数据时发生错误:', error);
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
    process.exit(0);
  })
  .catch((error) => {
    console.error('\n💥 创建失败:', error.message);
    process.exit(1);
  });