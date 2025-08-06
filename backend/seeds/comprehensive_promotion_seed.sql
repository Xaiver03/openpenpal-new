-- 综合晋升系统种子数据 (SOTA Level)
-- 完整的真实可用数据库数据

-- 1. 确保有足够的等级要求数据
INSERT INTO courier_level_requirements (from_level, to_level, requirement_type, requirement_value, is_mandatory, description) VALUES
-- 1级到2级要求
(1, 2, 'min_deliveries', '{"value": 50}', true, '完成至少50次投递'),
(1, 2, 'min_success_rate', '{"value": 95}', true, '成功率达到95%以上'),
(1, 2, 'min_service_days', '{"value": 30}', true, '服务时间超过30天'),
(1, 2, 'max_complaints', '{"value": 5}', true, '投诉次数不超过5次'),

-- 2级到3级要求
(2, 3, 'min_deliveries', '{"value": 200}', true, '完成至少200次投递'),
(2, 3, 'min_success_rate', '{"value": 97}', true, '成功率达到97%以上'),
(2, 3, 'min_subordinates', '{"value": 5}', true, '管理至少5名下级信使'),
(2, 3, 'min_service_months', '{"value": 3}', true, '担任二级信使满3个月'),
(2, 3, 'performance_score', '{"value": 85}', true, '绩效评分达到85分以上'),

-- 3级到4级要求
(3, 4, 'min_deliveries', '{"value": 500}', true, '完成至少500次投递'),
(3, 4, 'min_success_rate', '{"value": 98}', true, '成功率达到98%以上'),
(3, 4, 'school_recommendation', '{"value": 1}', true, '获得学校管理部门推荐'),
(3, 4, 'platform_approval', '{"value": 1}', true, '通过平台高级审核'),
(3, 4, 'min_service_months', '{"value": 6}', true, '担任三级信使满6个月'),
(3, 4, 'leadership_score', '{"value": 90}', true, '领导力评分达到90分以上')

ON CONFLICT (id) DO NOTHING;

-- 2. 插入更多真实的晋升申请数据
INSERT INTO courier_upgrade_requests (id, courier_id, current_level, request_level, reason, evidence, status, reviewer_id, reviewer_comment, created_at, reviewed_at, expires_at) VALUES
-- 待审核申请
('req-001', 'e60c3e31-666b-40fa-8168-665726f26b12', 1, 2, '已达到晋升标准，申请成为二级信使。在过去3个月中表现优秀，希望承担更多责任。', 
 '{"deliveries": 72, "success_rate": 96.8, "service_days": 45, "complaints": 1, "performance_score": 88}', 
 'pending', NULL, NULL, NOW() - INTERVAL '2 days', NULL, NOW() + INTERVAL '28 days'),

('req-002', 'courier2-record', 2, 3, '申请晋升为三级信使。已成功管理8名下级信使，业绩突出。', 
 '{"deliveries": 245, "success_rate": 97.5, "subordinates": 8, "service_months": 4, "performance_score": 87}', 
 'pending', NULL, NULL, NOW() - INTERVAL '1 day', NULL, NOW() + INTERVAL '29 days'),

-- 已批准申请
('req-003', 'courier1-id', 1, 2, '新手期表现优秀，申请晋升', 
 '{"deliveries": 55, "success_rate": 96.5, "service_days": 35}', 
 'approved', 'admin-user-id', '表现优秀，同意晋升', NOW() - INTERVAL '7 days', NOW() - INTERVAL '5 days', NOW() + INTERVAL '23 days'),

-- 已拒绝申请
('req-004', 'courier-test-id', 1, 2, '申请晋升为二级信使', 
 '{"deliveries": 35, "success_rate": 92.5, "service_days": 20}', 
 'rejected', 'admin-user-id', '投递次数和服务天数不足，建议继续努力', NOW() - INTERVAL '10 days', NOW() - INTERVAL '8 days', NOW() + INTERVAL '20 days'),

-- 过期申请
('req-005', 'courier-old-id', 2, 3, '申请晋升为三级信使', 
 '{"deliveries": 180, "success_rate": 96.8, "subordinates": 3}', 
 'expired', NULL, NULL, NOW() - INTERVAL '35 days', NULL, NOW() - INTERVAL '5 days')

ON CONFLICT (id) DO NOTHING;

-- 3. 插入晋升历史记录
INSERT INTO courier_promotion_history (id, courier_id, from_level, to_level, promoted_at, promoted_by, reason, evidence) VALUES
('hist-001', 'e60c3e31-666b-40fa-8168-665726f26b12', 0, 1, NOW() - INTERVAL '6 months', 'system', '新用户注册自动设为一级信使', '{}'),
('hist-002', 'courier2-record', 0, 1, NOW() - INTERVAL '8 months', 'system', '新用户注册自动设为一级信使', '{}'),
('hist-003', 'courier2-record', 1, 2, NOW() - INTERVAL '4 months', 'admin-user-id', '表现优秀，晋升为二级信使', 
 '{"deliveries": 68, "success_rate": 97.2, "service_days": 42}'),
('hist-004', 'senior-courier-id', 0, 1, NOW() - INTERVAL '2 years', 'system', '新用户注册', '{}'),
('hist-005', 'senior-courier-id', 1, 2, NOW() - INTERVAL '20 months', 'admin-user-id', '优秀表现晋升', 
 '{"deliveries": 89, "success_rate": 98.1, "service_days": 65}'),
('hist-006', 'senior-courier-id', 2, 3, NOW() - INTERVAL '1 year', 'admin-user-id', '管理能力突出', 
 '{"deliveries": 267, "success_rate": 98.5, "subordinates": 12, "service_months": 8}')

ON CONFLICT (id) DO NOTHING;

-- 4. 创建积分系统数据
INSERT INTO courier_points (courier_id, total_points, available_points, used_points, earned_points, updated_at) VALUES
('e60c3e31-666b-40fa-8168-665726f26b12', 1450, 1200, 250, 1450, NOW()),
('courier2-record', 2800, 2300, 500, 2800, NOW()),
('courier1-id', 950, 850, 100, 950, NOW()),
('senior-courier-id', 5600, 4200, 1400, 5600, NOW())
ON CONFLICT (courier_id) DO UPDATE SET
total_points = EXCLUDED.total_points,
available_points = EXCLUDED.available_points,
used_points = EXCLUDED.used_points,
earned_points = EXCLUDED.earned_points,
updated_at = EXCLUDED.updated_at;

-- 5. 创建徽章系统数据
INSERT INTO courier_badges (id, code, name, description, icon, points_reward, rarity, conditions, is_active, created_at) VALUES
('badge-001', 'FIRST_DELIVERY', '首次投递', '完成第一次投递任务', '📦', 50, 'common', '{"min_deliveries": 1}', true, NOW()),
('badge-002', 'CENTURY_COURIER', '百次投递', '完成100次投递任务', '💯', 200, 'rare', '{"min_deliveries": 100}', true, NOW()),
('badge-003', 'PERFECT_WEEK', '完美一周', '连续7天保持100%成功率', '⭐', 300, 'epic', '{"consecutive_perfect_days": 7}', true, NOW()),
('badge-004', 'TEAM_LEADER', '团队领袖', '成功管理10名下级信使', '👑', 500, 'legendary', '{"min_subordinates": 10}', true, NOW()),
('badge-005', 'SPEED_DEMON', '速度之王', '平均投递时间少于30分钟', '⚡', 250, 'epic', '{"max_avg_delivery_time": 30}', true, NOW())
ON CONFLICT (id) DO NOTHING;

-- 6. 分配一些徽章给信使
INSERT INTO courier_badge_earneds (courier_id, badge_id, earned_at, reason, reference) VALUES
('e60c3e31-666b-40fa-8168-665726f26b12', 'badge-001', NOW() - INTERVAL '3 months', '完成首次投递', 'delivery-001'),
('e60c3e31-666b-40fa-8168-665726f26b12', 'badge-002', NOW() - INTERVAL '1 month', '完成100次投递', 'delivery-100'),
('courier2-record', 'badge-001', NOW() - INTERVAL '8 months', '完成首次投递', 'delivery-002'),
('courier2-record', 'badge-002', NOW() - INTERVAL '6 months', '完成100次投递', 'delivery-150'),
('courier2-record', 'badge-003', NOW() - INTERVAL '2 months', '连续完美表现', 'week-perfect-1'),
('senior-courier-id', 'badge-001', NOW() - INTERVAL '2 years', '完成首次投递', 'delivery-ancient'),
('senior-courier-id', 'badge-002', NOW() - INTERVAL '20 months', '完成100次投递', 'delivery-veteran'),
('senior-courier-id', 'badge-003', NOW() - INTERVAL '18 months', '连续完美表现', 'week-perfect-senior'),
('senior-courier-id', 'badge-004', NOW() - INTERVAL '1 year', '优秀管理表现', 'management-excellence'),
('senior-courier-id', 'badge-005', NOW() - INTERVAL '6 months', '投递速度突出', 'speed-record-1')
ON CONFLICT (courier_id, badge_id) DO NOTHING;

-- 7. 创建积分交易记录
INSERT INTO courier_points_transactions (id, courier_id, type, amount, description, reference, created_at) VALUES
('trans-001', 'e60c3e31-666b-40fa-8168-665726f26b12', 'earn', 50, '完成投递任务', 'delivery-task-001', NOW() - INTERVAL '1 day'),
('trans-002', 'e60c3e31-666b-40fa-8168-665726f26b12', 'earn', 200, '获得百次投递徽章', 'badge-002', NOW() - INTERVAL '1 month'),
('trans-003', 'e60c3e31-666b-40fa-8168-665726f26b12', 'use', 100, '兑换优惠券', 'voucher-001', NOW() - INTERVAL '2 weeks'),
('trans-004', 'courier2-record', 'earn', 75, '完成紧急任务', 'urgent-task-001', NOW() - INTERVAL '3 days'),
('trans-005', 'courier2-record', 'earn', 300, '获得完美一周徽章', 'badge-003', NOW() - INTERVAL '2 months'),
('trans-006', 'senior-courier-id', 'earn', 500, '获得团队领袖徽章', 'badge-004', NOW() - INTERVAL '1 year'),
('trans-007', 'senior-courier-id', 'use', 200, '兑换培训课程', 'training-advanced', NOW() - INTERVAL '3 months')
ON CONFLICT (id) DO NOTHING;

-- 8. 更新统计信息
REFRESH MATERIALIZED VIEW v_courier_promotion_stats;

-- 9. 显示插入结果
SELECT 'Seed data inserted successfully' as status;
SELECT COUNT(*) as upgrade_requests FROM courier_upgrade_requests;
SELECT COUNT(*) as promotion_history FROM courier_promotion_history;
SELECT COUNT(*) as level_requirements FROM courier_level_requirements;
SELECT COUNT(*) as badges FROM courier_badges;
SELECT COUNT(*) as badge_earneds FROM courier_badge_earneds;
SELECT COUNT(*) as points_transactions FROM courier_points_transactions;