-- OpenPenPal 广场公开信件Mock数据
-- 从mock服务中提取的公开信件数据

-- 清理现有数据
DELETE FROM public_letters WHERE id LIKE 'mock_plaza_%';

-- 公开信件数据 - 广场展示
INSERT INTO public_letters (
    id, title, content, author_nickname, author_avatar, style,
    emotional_tone, tags, view_count, like_count, comment_count,
    is_featured, featured_reason, created_at, updated_at
) VALUES
-- 未来主题信件
('mock_plaza_001', '写给三年后的自己',
 '亲爱的未来的我，当你读到这封信的时候，希望你已经成为了更好的自己。还记得现在的我吗？那个在图书馆里挥汗如雨的学生，那个为了一道数学题而熬夜到凌晨的少年。我知道路还很长，但我相信，只要坚持下去，总会到达想要的地方。时间会证明一切，而我相信时间会站在努力的人这一边。愿未来的你，回望现在的我时，能够欣慰地说一句：谢谢你的坚持。',
 '匿名作者', '/images/user-001.png', 'future',
 'hopeful', 'future,growth,persistence,study', 1245, 89, 23,
 true, '深受读者喜爱的励志信件', '2024-01-20 10:00:00', '2024-01-20 10:00:00'),

-- 温暖主题信件  
('mock_plaza_002', '致正在迷茫的你',
 '如果你正在经历人生的低谷，请记住这只是暂时的。每个人都会有迷茫的时候，这是成长路上的必经之路。不要害怕迷茫，因为只有经历过黑暗，我们才能更珍惜光明。愿你在迷雾中找到前进的方向，愿你的心永远充满希望。记住，每一个明天都是新的开始，每一次呼吸都是生命的恩赐。你比自己想象的更坚强，比困难认为的更勇敢。',
 '温暖使者', '/images/user-002.png', 'warm',
 'comforting', 'comfort,hope,encouragement,strength', 2156, 167, 45,
 true, '温暖治愈，帮助很多读者走出困境', '2024-01-19 14:30:00', '2024-01-19 14:30:00'),

-- 故事主题信件
('mock_plaza_003', '一个关于友谊的故事',
 '我想和你分享一个关于友谊的故事，这个故事改变了我对友情的理解。那是一个秋天的下午，我坐在宿舍里感到孤独，突然收到了一个陌生人的来信。信里没有华丽的辞藻，只是简单地问候和分享日常生活。就这样，我们开始了笔友关系。虽然从未谋面，但通过文字，我们成为了最好的朋友。友谊不需要见面，不需要相同的背景，只需要一颗真诚的心。',
 '故事讲述者', '/images/user-003.png', 'story',
 'nostalgic', 'friendship,story,connection,letters', 892, 76, 18,
 false, NULL, '2024-01-18 09:15:00', '2024-01-18 09:15:00'),

-- 漂流主题信件
('mock_plaza_004', '漂流到远方的思念',
 '这封信将随风漂流到某个角落，希望能遇到同样思念远方的你。也许我们从未谋面，但在这个瞬间，我们的心是相通的。思念是一种奇妙的情感，它让距离变得不重要，让时间变得温柔。如果你收到这封信，请知道在某个地方，有一个人正在想念着远方的某个人，就像此刻的你我。愿所有的思念都能找到归宿，愿所有的等待都有意义。',
 '漂流者', '/images/user-004.png', 'drift',
 'melancholic', 'longing,distance,connection,drift', 567, 34, 12,
 false, NULL, '2024-01-17 16:45:00', '2024-01-17 16:45:00'),

-- 校园生活主题信件
('mock_plaza_005', '食堂里的小确幸',
 '今天食堂阿姨给我多打了一勺菜，瞬间觉得整个世界都亮了。大学生活中的小确幸，往往就藏在这些微不足道的细节里。一份意外的加菜，一个陌生同学的微笑，一本在图书馆偶然发现的好书，这些都是青春里最珍贵的记忆。愿每个人都能在平凡的日子里发现不平凡的美好。',
 '食堂常客', '/images/user-005.png', 'warm',
 'joyful', 'campus,happiness,daily,gratitude', 1034, 92, 27,
 false, NULL, '2024-01-16 12:20:00', '2024-01-16 12:20:00'),

-- 梦想主题信件
('mock_plaza_006', '关于梦想这件小事',
 '有人问我梦想是什么，我说梦想就是那个让你半夜想起来都会微笑的东西。它不一定要多么宏大，也不一定要多么现实，但它一定要是真实的。我的梦想是开一家书店，有猫咪，有咖啡香，有来来往往的读书人。也许在别人看来这很平凡，但对我来说，这就是我愿意为之努力一生的事情。',
 '书店梦想家', '/images/user-006.png', 'future',
 'inspired', 'dreams,bookstore,passion,simple', 743, 58, 15,
 false, NULL, '2024-01-15 20:30:00', '2024-01-15 20:30:00'),

-- 感恩主题信件
('mock_plaza_007', '给帮助过我的陌生人',
 '那个雨天为我撑伞的学长，那个在我迷路时为我指路的清洁阿姨，那个在我生病时给我买药的室友，还有那些我不知道名字但给过我帮助的陌生人们，谢谢你们。你们的善意让这个世界变得温暖，让我相信人间值得。我会把这份善意传递下去，成为别人生命中的那束光。',
 '感恩的心', '/images/user-007.png', 'warm',
 'grateful', 'gratitude,kindness,strangers,warmth', 1567, 134, 31,
 true, '传递正能量，让读者感受到人间温情', '2024-01-14 15:45:00', '2024-01-14 15:45:00'),

-- 成长主题信件
('mock_plaza_008', '那些让我成长的错误',
 '感谢那些我犯过的错误，它们教会了我比成功更多的东西。第一次考试失利让我学会了努力，第一次被拒绝让我学会了坚强，第一次与朋友争吵让我学会了宽容。错误不可怕，可怕的是不从错误中学习。每一个错误都是成长路上的垫脚石，让我们站得更高，看得更远。',
 '成长中的我', '/images/user-008.png', 'story',
 'reflective', 'growth,mistakes,learning,maturity', 823, 67, 19,
 false, NULL, '2024-01-13 18:20:00', '2024-01-13 18:20:00'),

-- 青春主题信件
('mock_plaza_009', '青春就是现在',
 '青春不是年龄，而是心境。不是粉面红唇，而是深沉的意志、恢宏的想象、炽热的感情。青春是生命深处的一股清泉。现在的我们正值青春，拥有无限的可能和机会。不要等到失去了才懂得珍惜，青春就是现在，就是此刻，就是你正在阅读这封信的这一瞬间。',
 '青春记录者', '/images/user-009.png', 'future',
 'energetic', 'youth,present,energy,possibility', 1456, 118, 38,
 true, '激励年轻人珍惜当下的美好文字', '2024-01-12 11:30:00', '2024-01-12 11:30:00'),

-- 思考主题信件
('mock_plaza_010', '关于孤独的思考',
 '孤独不是寂寞，孤独是一种选择。在这个喧嚣的世界里，能够独处是一种能力。孤独让我们学会与自己对话，学会倾听内心的声音，学会在宁静中思考人生。不要害怕孤独，享受孤独，在孤独中找到真正的自己。孤独是成长的必经之路，也是智慧的源泉。',
 '独处者', '/images/user-010.png', 'drift',
 'contemplative', 'solitude,self-reflection,wisdom,growth', 689, 45, 13,
 false, NULL, '2024-01-11 22:15:00', '2024-01-11 22:15:00');

-- 创建公开信件分类统计视图
CREATE OR REPLACE VIEW plaza_letters_stats AS
SELECT 
    style,
    COUNT(*) as letter_count,
    AVG(view_count) as avg_views,
    AVG(like_count) as avg_likes,
    AVG(comment_count) as avg_comments,
    COUNT(CASE WHEN is_featured = true THEN 1 END) as featured_count
FROM public_letters 
WHERE id LIKE 'mock_plaza_%'
GROUP BY style;

-- 创建情感色调统计视图
CREATE OR REPLACE VIEW plaza_emotional_stats AS
SELECT 
    emotional_tone,
    COUNT(*) as letter_count,
    AVG(view_count) as avg_views,
    AVG(like_count) as avg_likes
FROM public_letters 
WHERE id LIKE 'mock_plaza_%'
GROUP BY emotional_tone;

-- 添加索引以提高查询性能
CREATE INDEX IF NOT EXISTS idx_public_letters_style ON public_letters(style);
CREATE INDEX IF NOT EXISTS idx_public_letters_emotional_tone ON public_letters(emotional_tone);
CREATE INDEX IF NOT EXISTS idx_public_letters_featured ON public_letters(is_featured);
CREATE INDEX IF NOT EXISTS idx_public_letters_created_at ON public_letters(created_at);
CREATE INDEX IF NOT EXISTS idx_public_letters_view_count ON public_letters(view_count);
CREATE INDEX IF NOT EXISTS idx_public_letters_like_count ON public_letters(like_count);

-- 创建标签索引（如果使用PostgreSQL数组类型）
-- CREATE INDEX IF NOT EXISTS idx_public_letters_tags ON public_letters USING GIN(tags);

-- 数据完整性验证
SELECT 
    'Plaza Mock数据导入完成' as message,
    (SELECT COUNT(*) FROM public_letters WHERE id LIKE 'mock_plaza_%') as mock_letters_count,
    (SELECT COUNT(*) FROM public_letters WHERE is_featured = true AND id LIKE 'mock_plaza_%') as featured_count,
    (SELECT SUM(view_count) FROM public_letters WHERE id LIKE 'mock_plaza_%') as total_views,
    (SELECT SUM(like_count) FROM public_letters WHERE id LIKE 'mock_plaza_%') as total_likes;

-- 展示分类统计
SELECT '按风格统计' as stat_type, style as category, letter_count, avg_views, avg_likes 
FROM plaza_letters_stats
UNION ALL
SELECT '按情感统计' as stat_type, emotional_tone as category, letter_count, avg_views, avg_likes 
FROM plaza_emotional_stats
ORDER BY stat_type, letter_count DESC;