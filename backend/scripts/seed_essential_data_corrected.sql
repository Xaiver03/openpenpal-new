-- OpenPenPal Essential Data Seeding Script (CORRECTED)
-- ULTRATHINK Complete Database Integrity & Data Injection
-- Schema-matched version: 2025-08-16

-- =============================================================================
-- 1. COURIER SYSTEM DATA INJECTION
-- =============================================================================

-- Insert Courier Profiles for existing courier users (using actual schema)
INSERT INTO couriers (id, user_id, name, contact, school, zone, managed_op_code_prefix, has_printer, self_intro, can_mentor, weekly_hours, max_daily_tasks, created_at, updated_at) VALUES
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
 NOW(),
 NOW())

ON CONFLICT (user_id) DO NOTHING;

-- =============================================================================
-- 2. AI SYSTEM CONFIGURATION (using actual schema)
-- =============================================================================

-- Insert AI Configurations using actual table structure
INSERT INTO ai_configs (id, config_type, config_key, config_value, category, is_active, priority, version, provider, api_key, api_endpoint, model, temperature, max_tokens, daily_quota, used_quota, created_at, updated_at) VALUES
('ai-config-001', 'provider', 'moonshot_primary', '{"enabled": true, "priority": 1}', 'ai_provider', true, 1, 1, 'moonshot', 'demo-key-placeholder', 'https://api.moonshot.cn/v1', 'moonshot-v1-8k', 0.7, 2000, 10000, 0, NOW(), NOW()),
('ai-config-002', 'provider', 'openai_secondary', '{"enabled": false, "priority": 2}', 'ai_provider', false, 2, 1, 'openai', 'demo-key-placeholder', 'https://api.openai.com/v1', 'gpt-3.5-turbo', 0.8, 1500, 5000, 0, NOW(), NOW()),
('ai-config-003', 'provider', 'local_dev', '{"enabled": true, "dev_mode": true}', 'ai_provider', true, 3, 1, 'local', 'local-dev', 'http://localhost:8001', 'development-mode', 0.6, 1000, 99999, 0, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- =============================================================================
-- 3. CREDIT SHOP CONFIGURATION (using UUID)
-- =============================================================================

-- Credit Shop Configuration using proper UUID format
INSERT INTO credit_shop_configs (id, key, value, description, category, is_editable, created_at, updated_at) VALUES
('550e8400-e29b-41d4-a716-446655440001'::uuid, 'min_credits_for_shop', '100', '积分商城最低积分要求', 'access', true, NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440002'::uuid, 'max_daily_redemptions', '5', '每日最大兑换次数', 'limits', true, NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440003'::uuid, 'featured_products_count', '6', '首页推荐商品数量', 'display', true, NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440004'::uuid, 'shipping_fee_threshold', '500', '免运费积分门槛', 'shipping', true, NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440005'::uuid, 'return_window_days', '7', '退换货窗口期（天）', 'policy', true, NOW(), NOW())
ON CONFLICT (key) DO NOTHING;

-- =============================================================================
-- 4. SAMPLE LETTERS FOR DEMONSTRATION (using proper JSONB casting)
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
 '{"ai_generated": false, "word_count": 87, "writing_time_minutes": 15}'::jsonb,
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
 '{"ai_generated": false, "word_count": 82, "reply_to": "letter-demo-001", "response_time_hours": 18}'::jsonb,
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
 '{"ai_generated": false, "word_count": 95, "category": "campus_life", "season": "spring"}'::jsonb,
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
-- 6. USER PROFILES ENHANCEMENT (using actual schema)
-- =============================================================================

-- Add user profiles for better user experience
INSERT INTO user_profiles (user_id, real_name, phone, address, bio, preferences, created_at, updated_at) VALUES
((SELECT id FROM users WHERE username = 'alice'),
 '爱丽丝',
 '+86-138-1234-5678',
 '北京大学5号楼303室',
 '喜欢文学和摄影的大二学生，相信文字的力量能够连接心灵。',
 '{"interests": ["literature", "photography", "music"], "writing_style": "elegant", "preferred_topics": ["literature", "campus_life", "philosophy"]}'::json,
 NOW(),
 NOW()),

((SELECT id FROM users WHERE username = 'bob'),
 '小明',
 '+86-138-8765-4321',
 '北京大学6号楼201室',
 '计算机系学生，喜欢技术但也热爱传统文化，希望通过书信交流学习。',
 '{"interests": ["technology", "traditional_culture", "reading"], "writing_style": "modern", "preferred_topics": ["technology", "learning", "culture"]}'::json,
 NOW(),
 NOW())

ON CONFLICT (user_id) DO NOTHING;

-- =============================================================================
-- 7. SUMMARY REPORT
-- =============================================================================

-- Generate summary of data injection
SELECT 'CORRECTED DATA INJECTION SUMMARY:' as report;
SELECT COUNT(*) as courier_profiles_created FROM couriers WHERE id LIKE 'courier-%';
SELECT COUNT(*) as credit_rules_created FROM credit_task_rules;
SELECT COUNT(*) as sample_letters_created FROM letters WHERE id LIKE 'letter-demo-%';
SELECT COUNT(*) as letter_codes_created FROM letter_codes WHERE id LIKE 'code-demo-%';
SELECT COUNT(*) as user_profiles_created FROM user_profiles WHERE user_id IN (SELECT id FROM users WHERE username IN ('alice', 'bob'));
SELECT COUNT(*) as ai_configs_created FROM ai_configs WHERE id LIKE 'ai-config-%';
SELECT COUNT(*) as shop_configs_created FROM credit_shop_configs WHERE key LIKE '%_for_shop' OR key LIKE '%_redemptions' OR key LIKE '%_count' OR key LIKE '%_threshold' OR key LIKE '%_days';

SELECT 'CORRECTED ESSENTIAL DATA INJECTION COMPLETED SUCCESSFULLY!' as status;