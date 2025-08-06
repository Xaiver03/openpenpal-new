-- OpenPenPal 信使管理Mock数据
-- 包含四个层级信使管理页面的数据
-- Generated from React component mock data

-- 清理现有数据
DELETE FROM courier_stats WHERE id LIKE 'mock_%';
DELETE FROM couriers WHERE id LIKE 'mock_%';

-- 城市级信使数据 (Level 4 - 管理城市)
INSERT INTO couriers (
    id, username, email, level, status, points, task_count, completed_tasks, 
    subordinate_count, average_rating, join_date, last_active,
    city_name, city_code, school_count, coverage,
    contact_phone, contact_wechat, working_hours_start, working_hours_end, working_weekdays
) VALUES
-- 北京市总管理员
('mock_city_001', 'beijing_city_manager', 'beijing_manager@openpenpal.com', 4, 'active', 2500, 1200, 1150, 25, 4.9, '2023-06-01', '2024-01-24 07:45:00',
 '北京市', 'BEIJING_CITY', 25, '北京市高校全覆盖', '138****0001', 'beijing_courier_chief', '06:00', '22:00', '1,2,3,4,5,6,7'),

-- 上海市总管理员
('mock_city_002', 'shanghai_city_manager', 'shanghai_manager@openpenpal.com', 4, 'active', 2200, 980, 950, 20, 4.8, '2023-06-15', '2024-01-24 08:20:00',
 '上海市', 'SHANGHAI_CITY', 20, '上海市高校全覆盖', '138****0002', 'shanghai_courier_chief', '06:30', '21:30', '1,2,3,4,5,6,7'),

-- 广州市总管理员
('mock_city_003', 'guangzhou_city_manager', 'guangzhou_manager@openpenpal.com', 4, 'pending', 1800, 750, 720, 15, 4.7, '2023-07-01', '2024-01-23 19:30:00',
 '广州市', 'GUANGZHOU_CITY', 15, '广州市高校区域', '138****0003', 'guangzhou_courier_chief', NULL, NULL, NULL),

-- 深圳市总管理员
('mock_city_004', 'shenzhen_city_manager', 'shenzhen_manager@openpenpal.com', 4, 'frozen', 1500, 600, 570, 12, 4.6, '2023-07-15', '2024-01-15 16:20:00',
 '深圳市', 'SHENZHEN_CITY', 12, '深圳市高校区域', '138****0004', NULL, NULL, NULL, NULL);

-- 学校级信使数据 (Level 3 - 管理学校)
INSERT INTO couriers (
    id, username, email, level, status, points, task_count, completed_tasks, 
    subordinate_count, average_rating, join_date, last_active,
    school_name, school_code, zone_count, coverage,
    contact_phone, contact_wechat, working_hours_start, working_hours_end, working_weekdays
) VALUES
-- 北京大学校级信使
('mock_school_001', 'university_peking_manager', 'peking_manager@pku.edu.cn', 3, 'active', 1250, 456, 445, 8, 4.95, '2023-09-01', '2024-01-24 07:45:00',
 '北京大学', 'PKU_MAIN', 8, '燕园校区全区', '138****2001', 'pku_courier_head', '06:00', '22:00', '1,2,3,4,5,6,7'),

-- 清华大学校级信使  
('mock_school_002', 'university_tsinghua_manager', 'tsinghua_manager@thu.edu.cn', 3, 'active', 1180, 398, 390, 6, 4.88, '2023-09-05', '2024-01-24 08:20:00',
 '清华大学', 'THU_MAIN', 6, '紫荆校区全区', '138****2002', 'thu_courier_head', '06:30', '21:30', '1,2,3,4,5,6,7'),

-- 中国人民大学校级信使
('mock_school_003', 'university_renda_manager', 'renda_manager@ruc.edu.cn', 3, 'pending', 680, 234, 225, 4, 4.72, '2024-01-10', '2024-01-23 19:30:00',
 '中国人民大学', 'RUC_MAIN', 4, '明德楼区域', '138****2003', 'ruc_courier_head', NULL, NULL, NULL),

-- 北京师范大学校级信使
('mock_school_004', 'university_beishi_manager', 'beishi_manager@bnu.edu.cn', 3, 'frozen', 420, 189, 165, 3, 4.35, '2023-11-15', '2024-01-15 16:20:00',
 '北京师范大学', 'BNU_MAIN', 5, '北师大校区', '138****2004', NULL, NULL, NULL, NULL);

-- 片区级信使数据 (Level 2 - 管理片区)  
INSERT INTO couriers (
    id, username, email, level, status, points, task_count, completed_tasks,
    subordinate_count, average_rating, join_date, last_active,
    zone_name, zone_code, building_count, coverage_area,
    contact_phone, contact_wechat, working_hours_start, working_hours_end, working_weekdays
) VALUES
-- 东区片区信使
('mock_zone_001', 'zone_a_manager', 'zone_a@pku.edu.cn', 2, 'active', 580, 234, 225, 6, 4.9, '2023-12-01', '2024-01-24 08:30:00',
 '东区', 'SCHOOL_ZONE_A', 6, '宿舍区A1-A6', '138****3001', 'zone_a_manager', '07:00', '19:00', '1,2,3,4,5,6,7'),

-- 西区片区信使
('mock_zone_002', 'zone_b_manager', 'zone_b@pku.edu.cn', 2, 'active', 520, 198, 192, 4, 4.7, '2024-01-05', '2024-01-24 10:15:00',
 '西区', 'SCHOOL_ZONE_B', 4, '宿舍区B1-B4', '138****3002', NULL, '08:00', '20:00', '1,2,3,4,5,6'),

-- 南区片区信使
('mock_zone_003', 'zone_c_manager', 'zone_c@pku.edu.cn', 2, 'pending', 245, 89, 82, 3, 4.4, '2024-01-18', '2024-01-23 17:20:00',
 '南区', 'SCHOOL_ZONE_C', 5, '宿舍区C1-C5', '138****3003', 'zone_c_helper', NULL, NULL, NULL),

-- 北区片区信使
('mock_zone_004', 'zone_d_manager', 'zone_d@pku.edu.cn', 2, 'frozen', 180, 156, 135, 2, 3.8, '2023-11-20', '2024-01-18 14:45:00',
 '北区', 'SCHOOL_ZONE_D', 3, '宿舍区D1-D3', '138****3004', NULL, NULL, NULL, NULL);

-- 楼栋级信使数据 (Level 1 - 管理楼栋)
INSERT INTO couriers (
    id, username, email, level, status, points, task_count, completed_tasks,
    average_rating, join_date, last_active,
    building_name, building_code, floor_range, room_range,
    contact_phone, contact_wechat, working_hours_start, working_hours_end, working_weekdays
) VALUES
-- A栋楼栋信使
('mock_building_001', 'building_a_courier', 'building_a@pku.edu.cn', 1, 'active', 320, 156, 148, 4.9, '2024-01-10', '2024-01-24 09:15:00',
 'A栋', 'ZONE_A_001', '1-6层', '101-620', '138****4001', 'building_a_courier', '08:00', '18:00', '1,2,3,4,5,6'),

-- B栋楼栋信使
('mock_building_002', 'building_b_courier', 'building_b@pku.edu.cn', 1, 'active', 280, 134, 129, 4.6, '2024-01-15', '2024-01-24 11:30:00',
 'B栋', 'ZONE_A_002', '1-8层', '101-825', '138****4002', NULL, '09:00', '19:00', '1,2,3,4,5'),

-- C栋楼栋信使
('mock_building_003', 'building_c_courier', 'building_c@pku.edu.cn', 1, 'pending', 120, 45, 42, 4.3, '2024-01-20', '2024-01-23 15:45:00',
 'C栋', 'ZONE_A_003', '1-5层', NULL, '138****4003', 'building_c_helper', NULL, NULL, NULL),

-- D栋楼栋信使
('mock_building_004', 'building_d_courier', 'building_d@pku.edu.cn', 1, 'frozen', 95, 67, 58, 3.9, '2023-12-05', '2024-01-19 16:20:00',
 'D栋', 'ZONE_A_004', '1-7层', '101-715', '138****4004', NULL, NULL, NULL, NULL);

-- 统计数据
INSERT INTO courier_stats (
    id, level, stat_type, stat_key, value, updated_at
) VALUES
-- 城市级统计 (Level 4)
('mock_city_stats_001', 4, 'overview', 'total_schools', 25, NOW()),
('mock_city_stats_002', 4, 'overview', 'active_couriers', 38, NOW()),
('mock_city_stats_003', 4, 'overview', 'total_deliveries', 12678, NOW()),
('mock_city_stats_004', 4, 'overview', 'pending_tasks', 45, NOW()),
('mock_city_stats_005', 4, 'overview', 'average_rating', 4.9, NOW()),
('mock_city_stats_006', 4, 'overview', 'success_rate', 97.2, NOW()),

-- 学校级统计 (Level 3)
('mock_school_stats_001', 3, 'overview', 'total_zones', 8, NOW()),
('mock_school_stats_002', 3, 'overview', 'active_couriers', 12, NOW()),
('mock_school_stats_003', 3, 'overview', 'total_deliveries', 2456, NOW()),
('mock_school_stats_004', 3, 'overview', 'pending_tasks', 15, NOW()),
('mock_school_stats_005', 3, 'overview', 'average_rating', 4.8, NOW()),
('mock_school_stats_006', 3, 'overview', 'coverage_rate', 94.5, NOW()),

-- 片区级统计 (Level 2)
('mock_zone_stats_001', 2, 'overview', 'total_buildings', 12, NOW()),
('mock_zone_stats_002', 2, 'overview', 'active_couriers', 18, NOW()),
('mock_zone_stats_003', 2, 'overview', 'total_deliveries', 892, NOW()),
('mock_zone_stats_004', 2, 'overview', 'pending_tasks', 5, NOW()),
('mock_zone_stats_005', 2, 'overview', 'average_rating', 4.7, NOW()),
('mock_zone_stats_006', 2, 'overview', 'success_rate', 96.3, NOW()),

-- 楼栋级统计 (Level 1)
('mock_building_stats_001', 1, 'overview', 'total_couriers', 15, NOW()),
('mock_building_stats_002', 1, 'overview', 'active_couriers', 12, NOW()),
('mock_building_stats_003', 1, 'overview', 'pending_tasks', 8, NOW()),
('mock_building_stats_004', 1, 'overview', 'completed_today', 23, NOW()),
('mock_building_stats_005', 1, 'overview', 'average_rating', 4.8, NOW()),
('mock_building_stats_006', 1, 'overview', 'delivery_success_rate', 95.6, NOW());

-- 用户认证数据
INSERT INTO users (
    id, username, email, password_hash, role, status, 
    school_code, school_name, courier_level, courier_info,
    created_at, updated_at
) VALUES
-- 测试用户账号
('mock_user_alice', 'alice', 'alice@pku.edu.cn', '$2b$10$hashedpassword1', 'student', 'active',
 'PKU', '北京大学', NULL, NULL, NOW(), NOW()),

('mock_user_bob', 'bob', 'bob@thu.edu.cn', '$2b$10$hashedpassword2', 'student', 'active', 
 'THU', '清华大学', NULL, NULL, NOW(), NOW()),

('mock_user_admin', 'admin', 'admin@system.com', '$2b$10$hashedpassword3', 'admin', 'active',
 'ADMIN', '系统管理', NULL, NULL, NOW(), NOW()),

-- 信使用户账号
('mock_user_courier1', 'courier1', 'courier1@pku.edu.cn', '$2b$10$hashedpassword4', 'courier', 'active',
 'PKU', '北京大学', 1, '{"level": 1, "permissions": ["PKA1**"]}', NOW(), NOW()),

('mock_user_courier2', 'courier2', 'courier2@pku.edu.cn', '$2b$10$hashedpassword5', 'courier', 'active',
 'PKU', '北京大学', 2, '{"level": 2, "permissions": ["PKA*"]}', NOW(), NOW()),

('mock_user_courier3', 'courier3', 'courier3@pku.edu.cn', '$2b$10$hashedpassword6', 'courier', 'active',
 'PKU', '北京大学', 3, '{"level": 3, "permissions": ["PK*"]}', NOW(), NOW()),

('mock_user_courier4', 'courier4', 'courier4@pku.edu.cn', '$2b$10$hashedpassword7', 'courier', 'active',
 'PKU', '北京大学', 4, '{"level": 4, "permissions": ["**"]}', NOW(), NOW());

-- 信件数据
INSERT INTO letters (
    id, title, content, sender_id, receiver_hint, status, 
    style, is_public, created_at, updated_at
) VALUES
-- 个人信件
('mock_letter_001', '给远方朋友的信', '这是一封测试信件...', 'mock_user_alice', '北京大学的朋友', 'pending', 
 'personal', false, '2024-01-15 10:30:00', '2024-01-15 10:30:00'),

('mock_letter_002', '关于大学生活', '分享一些大学生活的感悟...', 'mock_user_bob', '清华大学的同学', 'delivered',
 'personal', false, '2024-01-16 14:20:00', '2024-01-16 14:20:00'),

-- 公开信件 - 广场展示
('mock_letter_plaza_001', '写给三年后的自己', 
 '亲爱的未来的我，当你读到这封信的时候，希望你已经成为了更好的自己。还记得现在的我吗？那个在图书馆里挥汗如雨的学生，那个为了一道数学题而熬夜到凌晨的少年。我知道路还很长，但我相信，只要坚持下去，总会到达想要的地方...',
 'mock_user_alice', NULL, 'public', 'future', true, '2024-01-20 10:00:00', '2024-01-20 10:00:00'),

('mock_letter_plaza_002', '致正在迷茫的你',
 '如果你正在经历人生的低谷，请记住这只是暂时的。每个人都会有迷茫的时候，这是成长路上的必经之路。不要害怕迷茫，因为只有经历过黑暗，我们才能更珍惜光明。愿你在迷雾中找到前进的方向，愿你的心永远充满希望...',
 'mock_user_bob', NULL, 'public', 'warm', true, '2024-01-19 14:30:00', '2024-01-19 14:30:00'),

('mock_letter_plaza_003', '一个关于友谊的故事',
 '我想和你分享一个关于友谊的故事，这个故事改变了我对友情的理解。那是一个秋天的下午，我坐在宿舍里感到孤独，突然收到了一个陌生人的来信...',
 'mock_user_alice', NULL, 'public', 'story', true, '2024-01-18 09:15:00', '2024-01-18 09:15:00'),

('mock_letter_plaza_004', '漂流到远方的思念',
 '这封信将随风漂流到某个角落，希望能遇到同样思念远方的你。也许我们从未谋面，但在这个瞬间，我们的心是相通的...',
 'mock_user_bob', NULL, 'public', 'drift', true, '2024-01-17 16:45:00', '2024-01-17 16:45:00');

-- 配送任务数据
INSERT INTO delivery_tasks (
    id, letter_id, courier_id, pickup_location, delivery_location,
    status, reward, estimated_time, created_at, updated_at
) VALUES
('mock_task_001', 'mock_letter_001', NULL, '北京大学邮局', '清华大学邮局', 
 'available', 15.00, 120, '2024-01-16 09:00:00', '2024-01-16 09:00:00'),

('mock_task_002', 'mock_letter_002', 'mock_user_courier1', '清华大学邮局', '北京大学邮局',
 'assigned', 12.00, 90, '2024-01-17 08:30:00', '2024-01-17 08:30:00');

-- 系统配置数据
INSERT INTO system_config (
    config_key, config_value, description, updated_at
) VALUES
('max_letter_length', '2000', '信件最大长度限制', NOW()),
('delivery_timeout', '72', '配送超时时间(小时)', NOW()),
('auto_match_enabled', 'true', '自动匹配功能开关', NOW()),
('maintenance_mode', 'false', '维护模式开关', NOW());

-- 博物馆展览数据
INSERT INTO museum_exhibitions (
    id, title, description, status, letter_count, created_at, updated_at
) VALUES
('mock_exhibition_001', '冬日温暖信件展', '收录冬季主题的温暖信件', 'active', 15, 
 '2024-01-15 00:00:00', '2024-01-15 00:00:00');

-- 添加索引以提高查询性能
CREATE INDEX IF NOT EXISTS idx_couriers_level ON couriers(level);
CREATE INDEX IF NOT EXISTS idx_couriers_status ON couriers(status);
CREATE INDEX IF NOT EXISTS idx_courier_stats_level ON courier_stats(level);
CREATE INDEX IF NOT EXISTS idx_letters_status ON letters(status);
CREATE INDEX IF NOT EXISTS idx_letters_public ON letters(is_public);
CREATE INDEX IF NOT EXISTS idx_delivery_tasks_status ON delivery_tasks(status);

-- 数据完整性验证
SELECT 
    'Mock数据导入完成' as message,
    (SELECT COUNT(*) FROM couriers WHERE id LIKE 'mock_%') as mock_couriers_count,
    (SELECT COUNT(*) FROM courier_stats WHERE id LIKE 'mock_%') as mock_stats_count,
    (SELECT COUNT(*) FROM users WHERE id LIKE 'mock_%') as mock_users_count,
    (SELECT COUNT(*) FROM letters WHERE id LIKE 'mock_%') as mock_letters_count;