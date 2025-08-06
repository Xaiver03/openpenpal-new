-- OpenPenPal 博物馆信件Mock数据
-- 从多个组件中提取的博物馆相关数据

-- 清理现有数据
DELETE FROM museum_letters WHERE id LIKE 'mock_%';
DELETE FROM museum_exhibitions WHERE id LIKE 'mock_%';

-- 博物馆展览数据
INSERT INTO museum_exhibitions (
    id, title, description, status, letter_count, 
    featured_image, curator_notes, created_at, updated_at
) VALUES
('mock_exhibition_winter', '冬日温暖信件展', '收录冬季主题的温暖信件，传递人间温暖', 'active', 15,
 '/images/exhibitions/winter-warmth.jpg', '这个展览展示了在寒冷冬日中人们互相传递的温暖话语', 
 '2024-01-15 00:00:00', '2024-01-15 00:00:00'),

('mock_exhibition_friendship', '友谊永恒主题展', '关于友谊的珍贵信件集合', 'active', 12,
 '/images/exhibitions/friendship.jpg', '友谊是人生中最珍贵的财富之一，这些信件见证了真挚的友情',
 '2024-01-10 00:00:00', '2024-01-10 00:00:00'),

('mock_exhibition_future', '致未来的自己', '时间胶囊式的未来信件展', 'active', 8,
 '/images/exhibitions/future-self.jpg', '每个人都有对未来的憧憬，这些信件记录了青春的梦想与希望',
 '2024-01-05 00:00:00', '2024-01-05 00:00:00'),

('mock_exhibition_campus', '校园时光记忆展', '大学生活的美好回忆', 'planning', 0,
 '/images/exhibitions/campus-life.jpg', '即将开放的校园主题展览，征集中...',
 '2024-01-25 00:00:00', '2024-01-25 00:00:00');

-- 博物馆信件数据
INSERT INTO museum_letters (
    id, title, content, original_author, exhibition_id, category,
    date_written, date_archived, curator_notes, tags, featured,
    emotional_score, historical_value, created_at, updated_at
) VALUES
-- 冬日温暖主题信件
('mock_museum_001', '雪夜里的温暖', 
 '今夜雪花纷飞，我坐在宿舍的窗前，想起了家乡的炉火。虽然身在异乡，但同学们的关怀让我感到无比温暖。就像这封信一样，希望能把温暖传递给更多需要的人。愿每个在寒夜中的人都能感受到人间的温情。',
 '匿名学生', 'mock_exhibition_winter', 'warmth', 
 '2023-12-20', '2024-01-15', '这封信体现了冬日中人与人之间的温暖关怀', 'winter,warmth,care,dormitory', true,
 9.2, 8.5, NOW(), NOW()),

('mock_museum_002', '热茶与友情',
 '室友为我泡的这杯热茶，比任何昂贵的礼物都要珍贵。在这个寒冷的冬日里，有人愿意为你递上一杯热茶，这就是最朴素也最真挚的关爱。希望我也能成为别人寒冬中的一杯热茶。',
 '宿舍103的小王', 'mock_exhibition_winter', 'friendship',
 '2023-12-15', '2024-01-15', '简单的日常生活中蕴含的深厚友情', 'friendship,tea,roommate,simple', false,
 8.8, 7.9, NOW(), NOW()),

('mock_museum_003', '图书馆里的约定',
 '我们约定每个周末都在图书馆的老位置见面，一起学习，一起奋斗。这个冬天因为有了你的陪伴而不再孤单。知识的温度和友情的温度交织在一起，成为了我大学生活中最美好的回忆。',
 '图书馆常客', 'mock_exhibition_winter', 'study',
 '2023-12-10', '2024-01-15', '学习伙伴之间的深厚友谊', 'library,study,partnership,memory', true,
 8.5, 8.1, NOW(), NOW()),

-- 友谊主题信件
('mock_museum_004', '室友的生日惊喜',
 '为了给室友准备生日惊喜，我们偷偷策划了一个月。看到她感动得哭了的样子，我知道所有的努力都是值得的。友情不在于昂贵的礼物，而在于那份用心和真诚。这份友谊将是我一生的财富。',
 '宿舍策划团', 'mock_exhibition_friendship', 'friendship',
 '2023-11-25', '2024-01-10', '朋友间精心准备的惊喜体现了友谊的珍贵', 'birthday,surprise,roommate,treasure', true,
 9.5, 8.8, NOW(), NOW()),

('mock_museum_005', '深夜的谈心',
 '那个失眠的夜晚，我们聊到了天亮。从学业压力到人生理想，从家庭琐事到未来规划。原来每个人都有自己的困惑和迷茫，但有朋友愿意倾听就已经足够。感谢那个陪我到天亮的你。',
 '夜谈者', 'mock_exhibition_friendship', 'support',
 '2023-11-20', '2024-01-10', '深夜谈心展现了友谊中的倾听与支持', 'night,talk,support,understanding', false,
 8.9, 8.3, NOW(), NOW()),

('mock_museum_006', '食堂里的默契',
 '每天中午12点，我们总是不约而同地出现在食堂的那张桌子旁。不需要提前约定，这就是朋友之间的默契。一起吃饭、一起抱怨菜品、一起笑到肚子疼，这些平凡的时光就是青春最美的样子。',
 '食堂老友', 'mock_exhibition_friendship', 'daily',
 '2023-11-15', '2024-01-10', '日常生活中的友谊默契', 'cafeteria,routine,understanding,youth', false,
 8.7, 7.6, NOW(), NOW()),

-- 致未来主题信件
('mock_museum_007', '二十年后的自己',
 '亲爱的四十岁的我，现在的我20岁，正在为期末考试焦虑，为找实习发愁，为感情困惑。我不知道二十年后的你是否还记得现在的这些烦恼，但我希望你能记住现在这份纯真和执着。愿你成为了更好的自己。',
 '20岁的梦想家', 'mock_exhibition_future', 'future',
 '2023-12-01', '2024-01-05', '年轻人对未来自己的期待与祝福', 'future,growth,dreams,twenty', true,
 9.1, 9.2, NOW(), NOW()),

('mock_museum_008', '毕业十年后',
 '十年后再读这封信的我，希望你还记得现在坐在宿舍里写这封信的我。希望你已经实现了成为一名教师的梦想，希望你还记得那个说要改变世界的青涩少年。无论走到哪里，都不要忘记初心。',
 '师范生小李', 'mock_exhibition_future', 'career',
 '2023-11-30', '2024-01-05', '对职业理想的坚持与期待', 'career,teacher,dreams,persistence', false,
 8.6, 8.7, NOW(), NOW()),

('mock_museum_009', '给未来的妈妈',
 '未来的我，如果你已经为人母，请记住现在单身的你也曾经很快乐。不要因为生活的琐碎而忘记了自己的梦想，不要因为责任的重担而丢失了少女的心境。愿你平衡好所有的角色，活出自己的精彩。',
 '独立女孩', 'mock_exhibition_future', 'life',
 '2023-11-28', '2024-01-05', '女性对未来人生角色的思考', 'motherhood,independence,balance,women', true,
 9.0, 8.9, NOW(), NOW()),

-- 校园生活主题信件 (为即将开放的展览准备)
('mock_museum_010', '初入校园的那一天',
 '还记得拖着行李箱走进宿舍的那个下午吗？一切都是那么新鲜，室友的第一次见面，食堂的第一顿饭，第一次熄灯后的卧谈会。虽然当时有些紧张和不安，但现在回想起来，那就是青春最美好的开始。',
 '大一新生', 'mock_exhibition_campus', 'beginning',
 '2023-09-01', '2024-01-25', '大学生活的美好开始', 'freshman,beginning,dormitory,nervousness', false,
 8.4, 8.0, NOW(), NOW()),

('mock_museum_011', '期末周的疯狂',
 '图书馆、咖啡、熬夜、复习资料满天飞。期末周的疯狂是每个大学生都经历过的"战争"。虽然压力山大，但和同学们一起奋战的感觉很棒。那些一起熬过的夜晚，成为了最深刻的回忆。',
 '考试战士', 'mock_exhibition_campus', 'study',
 '2023-12-25', '2024-01-25', '期末考试期间的学习生活', 'exam,library,study,pressure', false,
 7.8, 7.5, NOW(), NOW()),

('mock_museum_012', '社团活动的收获',
 '加入话剧社是我大学里做得最正确的决定。从紧张的试镜到第一次登台，从背台词到谢幕的掌声，每一个环节都让我成长。社团不仅给了我展示的舞台，更给了我一群志同道合的朋友。',
 '话剧社成员', 'mock_exhibition_campus', 'activity',
 '2023-10-15', '2024-01-25', '大学社团活动的美好体验', 'club,drama,growth,friendship', true,
 8.8, 8.2, NOW(), NOW());

-- 创建统计视图
CREATE OR REPLACE VIEW museum_exhibition_stats AS
SELECT 
    e.id,
    e.title,
    e.status,
    COUNT(ml.id) as actual_letter_count,
    COUNT(CASE WHEN ml.featured = true THEN 1 END) as featured_count,
    AVG(ml.emotional_score) as avg_emotional_score,
    AVG(ml.historical_value) as avg_historical_value,
    MIN(ml.date_written) as earliest_letter,
    MAX(ml.date_written) as latest_letter
FROM museum_exhibitions e
LEFT JOIN museum_letters ml ON e.id = ml.exhibition_id
WHERE e.id LIKE 'mock_%'
GROUP BY e.id, e.title, e.status;

-- 添加索引以提高查询性能
CREATE INDEX IF NOT EXISTS idx_museum_letters_exhibition ON museum_letters(exhibition_id);
CREATE INDEX IF NOT EXISTS idx_museum_letters_category ON museum_letters(category);
CREATE INDEX IF NOT EXISTS idx_museum_letters_featured ON museum_letters(featured);
CREATE INDEX IF NOT EXISTS idx_museum_letters_date_written ON museum_letters(date_written);
CREATE INDEX IF NOT EXISTS idx_museum_exhibitions_status ON museum_exhibitions(status);

-- 数据完整性验证
SELECT 
    'Museum Mock数据导入完成' as message,
    (SELECT COUNT(*) FROM museum_exhibitions WHERE id LIKE 'mock_%') as mock_exhibitions_count,
    (SELECT COUNT(*) FROM museum_letters WHERE id LIKE 'mock_%') as mock_letters_count,
    (SELECT COUNT(*) FROM museum_letters WHERE featured = true AND id LIKE 'mock_%') as featured_letters_count;

-- 展示统计信息
SELECT * FROM museum_exhibition_stats;