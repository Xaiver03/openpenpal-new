-- OpenPenPal Essential Data Seeding Script
-- ULTRATHINK Complete Database Integrity & Data Injection
-- Generated: 2025-08-16

-- =============================================================================
-- 1. COURIER SYSTEM DATA INJECTION
-- =============================================================================

-- Insert Courier Profiles for existing courier users
INSERT INTO couriers (id, user_id, name, contact, school, zone, managed_op_code_prefix, has_printer, self_intro, can_mentor, weekly_hours, max_daily_tasks, level, status, rating, experience_points, total_deliveries, successful_deliveries, current_capacity, max_capacity, created_at, updated_at) VALUES
-- L4 City Coordinator (Highest Level)
('courier-l4-001', 
 (SELECT id FROM users WHERE username = 'courier_level4'), 
 'L4-北京总代-张明', 
 '+86-138-0000-1001', 
 '北京大学', 
 'BEIJING', 
 'PK', 
 true, 
 '北京地区四级信使总代，负责整个北京高校信使网络的协调管理，具有丰富的物流管理经验。', 
 'senior', 
 40, 
 50,
 4,
 'active',
 4.9,
 2500,
 1200,
 1180,
 5,
 50,
 NOW(),
 NOW()),

-- L3 School Manager  
('courier-l3-001', 
 (SELECT id FROM users WHERE username = 'courier_level3'), 
 'L3-北大校级-李华', 
 '+86-138-0000-1002', 
 '北京大学', 
 'BJDX', 
 'PK5F', 
 true, 
 '北京大学校级信使，负责校内各区域的信件分发协调，熟悉校园地理位置。', 
 'yes', 
 30, 
 40,
 3,
 'active',
 4.8,
 800,
 450,
 440,
 8,
 40,
 NOW(),
 NOW()),

-- L2 Area Manager
('courier-l2-001', 
 (SELECT id FROM users WHERE username = 'courier_level2'), 
 'L2-5号楼区域-王芳', 
 '+86-138-0000-1003', 
 '北京大学', 
 'BJDX-5F', 
 'PK5F', 
 false, 
 '负责5号楼及周边区域的信件投递，对该区域非常熟悉，投递效率高。', 
 'yes', 
 20, 
 25,
 2,
 'active',
 4.7,
 300,
 180,
 175,
 12,
 25,
 NOW(),
 NOW()),

-- L1 Building Courier
('courier-l1-001', 
 (SELECT id FROM users WHERE username = 'courier_level1'), 
 'L1-宿舍楼-赵强', 
 '+86-138-0000-1004', 
 '北京大学', 
 'BJDX-5F-3D', 
 'PK5F3D', 
 false, 
 '负责宿舍楼的直接投递，认真负责，深受同学们信赖。', 
 'no', 
 15, 
 20,
 1,
 'active',
 4.6,
 150,
 85,
 82,
 8,
 20,
 NOW(),
 NOW())

ON CONFLICT (user_id) DO NOTHING;

-- =============================================================================
-- 2. CREDIT SYSTEM CONFIGURATION
-- =============================================================================

-- Insert Credit Task Rules for all major task types
INSERT INTO credit_task_rules (id, task_type, points, daily_limit, weekly_limit, is_active, auto_execute, description, constraints, created_at, updated_at) VALUES
-- Letter Related Tasks
('rule-letter-001', 'letter_created', 10, 5, 20, true, true, '创建信件奖励', '{"min_content_length": 50}', NOW(), NOW()),
('rule-letter-002', 'letter_generated', 5, 10, 40, true, true, '生成信件编号奖励', '{}', NOW(), NOW()),
('rule-letter-003', 'letter_delivered', 15, 8, 30, true, true, '信件成功送达奖励', '{}', NOW(), NOW()),
('rule-letter-004', 'letter_read', 8, 15, 50, true, true, '信件被阅读奖励', '{}', NOW(), NOW()),
('rule-letter-005', 'receive_letter', 12, 10, 35, true, true, '收到信件奖励', '{}', NOW(), NOW()),
('rule-letter-006', 'public_like', 20, 3, 10, true, true, '公开信被点赞奖励', '{"min_likes": 1}', NOW(), NOW()),

-- Courier Related Tasks  
('rule-courier-001', 'courier_first', 50, 1, 1, true, true, '信使首次任务奖励', '{}', NOW(), NOW()),
('rule-courier-002', 'courier_delivery', 25, 20, 100, true, true, '信使投递任务奖励', '{}', NOW(), NOW()),

-- Museum Related Tasks
('rule-museum-001', 'museum_submit', 30, 3, 10, true, true, '博物馆投稿奖励', '{"min_content_quality": 70}', NOW(), NOW()),
('rule-museum-002', 'museum_approved', 100, 2, 5, true, true, '博物馆审核通过奖励', '{}', NOW(), NOW()),
('rule-museum-003', 'museum_liked', 15, 5, 20, true, true, '博物馆点赞奖励', '{}', NOW(), NOW()),

-- System Tasks
('rule-system-001', 'opcode_approval', 200, 1, 2, true, false, 'OP Code审核奖励', '{"admin_only": true}', NOW(), NOW()),
('rule-system-002', 'community_badge', 150, 1, 3, true, false, '社区徽章奖励', '{}', NOW(), NOW()),
('rule-system-003', 'admin_reward', 500, 1, 1, true, false, '管理员特殊奖励', '{"admin_only": true}', NOW(), NOW()),

-- AI & Writing Tasks
('rule-ai-001', 'writing_challenge', 40, 2, 8, true, true, '写作挑战奖励', '{"min_words": 100}', NOW(), NOW()),
('rule-ai-002', 'ai_interaction', 5, 10, 30, true, true, 'AI互动奖励', '{}', NOW(), NOW())

ON CONFLICT (task_type) DO NOTHING;

-- =============================================================================
-- 3. SYSTEM CONFIGURATION DATA
-- =============================================================================

-- AI Configuration
INSERT INTO ai_configs (id, provider, model_name, api_key, base_url, max_tokens, temperature, is_active, config_data, created_at, updated_at) VALUES
('ai-config-001', 'moonshot', 'moonshot-v1-8k', 'demo-key-placeholder', 'https://api.moonshot.cn/v1', 2000, 0.7, true, '{"timeout": 30, "retry_count": 3}', NOW(), NOW()),
('ai-config-002', 'openai', 'gpt-3.5-turbo', 'demo-key-placeholder', 'https://api.openai.com/v1', 1500, 0.8, false, '{"timeout": 30, "retry_count": 3}', NOW(), NOW()),
('ai-config-003', 'local', 'development-mode', 'local-dev', 'http://localhost:8001', 1000, 0.6, true, '{"dev_mode": true}', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- Credit Shop Configuration
INSERT INTO credit_shop_configs (id, key, value, description, category, is_editable, created_at, updated_at) VALUES
('shop-config-001', 'min_credits_for_shop', '100', '积分商城最低积分要求', 'access', true, NOW(), NOW()),
('shop-config-002', 'max_daily_redemptions', '5', '每日最大兑换次数', 'limits', true, NOW(), NOW()),
('shop-config-003', 'featured_products_count', '6', '首页推荐商品数量', 'display', true, NOW(), NOW()),
('shop-config-004', 'shipping_fee_threshold', '500', '免运费积分门槛', 'shipping', true, NOW(), NOW()),
('shop-config-005', 'return_window_days', '7', '退换货窗口期（天）', 'policy', true, NOW(), NOW())
ON CONFLICT (key) DO NOTHING;

-- Storage Configuration  
INSERT INTO storage_configs (id, storage_type, base_path, max_file_size, allowed_extensions, is_active, config_json, created_at, updated_at) VALUES
('storage-001', 'local', '/uploads/letters', 10485760, 'jpg,jpeg,png,pdf', true, '{"backup_enabled": true, "compression": true}', NOW(), NOW()),
('storage-002', 'local', '/uploads/qr_codes', 1048576, 'png,svg', true, '{"auto_cleanup": true}', NOW(), NOW()),
('storage-003', 'local', '/uploads/avatars', 5242880, 'jpg,jpeg,png', true, '{"resize_enabled": true, "max_dimension": 500}', NOW(), NOW())
ON CONFLICT (storage_type, base_path) DO NOTHING;

-- =============================================================================
-- 4. SAMPLE LETTERS FOR DEMONSTRATION
-- =============================================================================

-- Create sample letters for testing and demonstration
INSERT INTO letters (id, user_id, author_id, recipient_id, title, content, author_name, style, status, visibility, type, recipient_op_code, sender_op_code, metadata, created_at, updated_at) VALUES
-- Letter from Alice to OP Code destination
('letter-demo-001', 
 (SELECT id FROM users WHERE username = 'alice'),
 (SELECT id FROM users WHERE username = 'alice'),
 (SELECT id FROM users WHERE username = 'bob'),
 '致远方朋友的第一封信',
 '亲爱的朋友，这是我在OpenPenPal平台写下的第一封信。希望通过这种传统而美好的方式，我们能够建立深厚的友谊。手写信件有着独特的温度，每一个字都承载着真诚的情感。期待收到你的回信！',
 '爱丽丝',
 'classic',
 'delivered',
 'private',
 'original',
 'PK5F3D',
 'PK5F01',
 '{"ai_generated": false, "word_count": 87, "writing_time_minutes": 15}',
 NOW() - INTERVAL '2 days',
 NOW() - INTERVAL '1 day'),

-- Letter from Bob replying to Alice
('letter-demo-002', 
 (SELECT id FROM users WHERE username = 'bob'),
 (SELECT id FROM users WHERE username = 'bob'),
 (SELECT id FROM users WHERE username = 'alice'),
 'Re: 致远方朋友的第一封信',
 '亲爱的爱丽丝，收到你的信件让我感到非常惊喜！在这个数字化的时代，能够收到手写信件确实是一种特别的体验。你的字里行间透露出的真诚深深打动了我。我也希望我们能够成为好朋友，通过书信分享彼此的生活和想法。',
 '小明',
 'modern',
 'read',
 'private',
 'reply',
 'PK5F01',
 'PK5F3D',
 '{"ai_generated": false, "word_count": 82, "reply_to": "letter-demo-001", "response_time_hours": 18}',
 NOW() - INTERVAL '1 day',
 NOW()),

-- Public letter for museum/showcase
('letter-demo-003', 
 (SELECT id FROM users WHERE username = 'alice'),
 (SELECT id FROM users WHERE username = 'alice'),
 NULL,
 '春日校园随想',
 '春天的校园格外美丽，樱花盛开，柳絮飞舞。走在石径小路上，听着鸟儿的啁啾声，心情不由得开朗起来。学习之余，这样的美景让人忘却了疲劳，仿佛整个世界都充满了希望和活力。希望能与更多朋友分享这份美好的心情。',
 '爱丽丝',
 'elegant',
 'approved',
 'public',
 'original',
 'PUBLIC',
 'PK5F01',
 '{"ai_generated": false, "word_count": 95, "category": "campus_life", "season": "spring", "public_approved_at": "' || (NOW() - INTERVAL '3 hours')::text || '"}',
 NOW() - INTERVAL '3 days',
 NOW() - INTERVAL '3 hours')

ON CONFLICT (id) DO NOTHING;

-- =============================================================================
-- 5. LETTER CODES FOR SAMPLE LETTERS
-- =============================================================================

-- Generate letter codes for the sample letters
INSERT INTO letter_codes (id, letter_id, code, status, recipient_code, created_at, updated_at) VALUES
('code-demo-001', 'letter-demo-001', 'LTR241608001', 'delivered', 'PK5F3D', NOW() - INTERVAL '2 days', NOW() - INTERVAL '1 day'),
('code-demo-002', 'letter-demo-002', 'LTR241608002', 'delivered', 'PK5F01', NOW() - INTERVAL '1 day', NOW()),
('code-demo-003', 'letter-demo-003', 'LTR241608003', 'bound', 'PUBLIC', NOW() - INTERVAL '3 days', NOW() - INTERVAL '3 hours')
ON CONFLICT (letter_id) DO NOTHING;

-- =============================================================================
-- 6. MUSEUM SAMPLE DATA
-- =============================================================================

-- Add sample museum entry for the public letter
INSERT INTO museum_items (id, title, content, author_name, original_date, category, status, letter_id, origin_op_code, view_count, like_count, is_featured, metadata, created_at, updated_at) VALUES
('museum-demo-001',
 '春日校园随想',
 '春天的校园格外美丽，樱花盛开，柳絮飞舞。走在石径小路上，听着鸟儿的啁啾声，心情不由得开朗起来。学习之余，这样的美景让人忘却了疲劳，仿佛整个世界都充满了希望和活力。希望能与更多朋友分享这份美好的心情。',
 '爱丽丝',
 NOW() - INTERVAL '3 days',
 'campus_life',
 'published',
 'letter-demo-003',
 'PK5F01',
 23,
 5,
 true,
 '{"tags": ["春天", "校园", "美景", "心情"], "emotion": "joyful", "style_analysis": "elegant_prose"}',
 NOW() - INTERVAL '3 hours',
 NOW())
ON CONFLICT (letter_id) DO NOTHING;

-- =============================================================================
-- 7. USER PROFILES ENHANCEMENT
-- =============================================================================

-- Add user profiles for better user experience
INSERT INTO user_profiles (id, user_id, display_name, bio, avatar_url, school_year, major, interests, personality_traits, letter_writing_style, preferred_topics, contact_preferences, privacy_settings, created_at, updated_at) VALUES
('profile-alice',
 (SELECT id FROM users WHERE username = 'alice'),
 '爱丽丝 Alice',
 '喜欢文学和摄影的大二学生，相信文字的力量能够连接心灵。',
 '/uploads/avatars/alice.jpg',
 2,
 '中文系',
 'literature,photography,music,nature',
 'thoughtful,creative,empathetic',
 'elegant',
 'literature,campus_life,philosophy,art',
 'letters_only',
 'friends_only',
 NOW(),
 NOW()),

('profile-bob',
 (SELECT id FROM users WHERE username = 'bob'),
 '小明 Bob',
 '计算机系学生，喜欢技术但也热爱传统文化，希望通过书信交流学习。',
 '/uploads/avatars/bob.jpg',
 3,
 '计算机科学',
 'technology,traditional_culture,reading,sports',
 'logical,curious,friendly',
 'modern',
 'technology,learning,culture,sports',
 'all_methods',
 'public',
 NOW(),
 NOW())

ON CONFLICT (user_id) DO NOTHING;

-- =============================================================================
-- 8. CREDIT INITIAL ALLOCATION
-- =============================================================================

-- Give initial credits to test users
INSERT INTO user_credits (id, user_id, credit_type, amount, source, description, expires_at, created_at, updated_at) VALUES
('credit-alice-001', (SELECT id FROM users WHERE username = 'alice'), 'letter_writing', 100, 'initial_allocation', '新用户注册奖励', NOW() + INTERVAL '1 year', NOW(), NOW()),
('credit-alice-002', (SELECT id FROM users WHERE username = 'alice'), 'social_activity', 50, 'initial_allocation', '新用户注册奖励', NOW() + INTERVAL '1 year', NOW(), NOW()),
('credit-bob-001', (SELECT id FROM users WHERE username = 'bob'), 'letter_writing', 100, 'initial_allocation', '新用户注册奖励', NOW() + INTERVAL '1 year', NOW(), NOW()),
('credit-bob-002', (SELECT id FROM users WHERE username = 'bob'), 'social_activity', 50, 'initial_allocation', '新用户注册奖励', NOW() + INTERVAL '1 year', NOW(), NOW()),
('credit-admin-001', (SELECT id FROM users WHERE username = 'admin'), 'administrative', 10000, 'admin_allocation', '管理员初始积分', NOW() + INTERVAL '10 years', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- =============================================================================
-- SUMMARY REPORT
-- =============================================================================

-- Generate summary of data injection
SELECT 'DATA INJECTION SUMMARY:' as report;
SELECT COUNT(*) as courier_profiles_created FROM couriers;
SELECT COUNT(*) as credit_rules_created FROM credit_task_rules;
SELECT COUNT(*) as sample_letters_created FROM letters;
SELECT COUNT(*) as letter_codes_created FROM letter_codes;
SELECT COUNT(*) as museum_items_created FROM museum_items;
SELECT COUNT(*) as user_profiles_created FROM user_profiles;
SELECT COUNT(*) as ai_configs_created FROM ai_configs;
SELECT COUNT(*) as shop_configs_created FROM credit_shop_configs;
SELECT COUNT(*) as storage_configs_created FROM storage_configs;

SELECT 'ESSENTIAL DATA INJECTION COMPLETED SUCCESSFULLY!' as status;