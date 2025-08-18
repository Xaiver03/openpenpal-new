-- OP Code 层级系统测试数据（最终版）

-- 创建一些示例的已占用OP Code记录
-- 使用系统用户作为created_by
INSERT INTO signal_codes (code, school_code, area_code, point_code, description, code_type, created_by, is_public, is_active, created_at, updated_at) VALUES
-- 北京大学示例
('PK1A01', 'PK', '1A', '01', '东区A栋101室', 'dormitory', 'system', false, true, NOW(), NOW()),
('PK1A02', 'PK', '1A', '02', '东区A栋102室', 'dormitory', 'system', false, true, NOW(), NOW()),
('PK1B01', 'PK', '1B', '01', '东区B栋101室', 'dormitory', 'system', false, true, NOW(), NOW()),
('PK2A01', 'PK', '2A', '01', '西区A栋101室', 'dormitory', 'system', false, true, NOW(), NOW()),
('PK2A02', 'PK', '2A', '02', '西区A栋102室', 'dormitory', 'system', false, true, NOW(), NOW()),
('PK2B01', 'PK', '2B', '01', '西区B栋101室', 'dormitory', 'system', false, true, NOW(), NOW()),
-- 清华大学示例
('QHAA01', 'QH', 'AA', '01', '紫荆1号楼101室', 'dormitory', 'system', false, true, NOW(), NOW()),
('QHAA02', 'QH', 'AA', '02', '紫荆1号楼102室', 'dormitory', 'system', false, true, NOW(), NOW()),
('QHAB01', 'QH', 'AB', '01', '紫荆2号楼101室', 'dormitory', 'system', false, true, NOW(), NOW()),
('QHBA01', 'QH', 'BA', '01', 'W楼A座101室', 'dormitory', 'system', false, true, NOW(), NOW()),
-- 复旦大学示例
('FD1A01', 'FD', '1A', '01', '东区A楼101室', 'dormitory', 'system', false, true, NOW(), NOW()),
('FD1A02', 'FD', '1A', '02', '东区A楼102室', 'dormitory', 'system', false, true, NOW(), NOW()),
-- 中山大学示例
('ZS1A01', 'ZS', '1A', '01', '东校区A栋101室', 'dormitory', 'system', false, true, NOW(), NOW()),
('ZS1A02', 'ZS', '1A', '02', '东校区A栋102室', 'dormitory', 'system', false, true, NOW(), NOW())
ON CONFLICT (code) DO NOTHING;

-- 输出测试信息
SELECT '已占用的OP Code总数:' as info, COUNT(*) as count FROM signal_codes WHERE is_active = true;
SELECT '北京大学已占用:' as info, COUNT(*) as count FROM signal_codes WHERE school_code = 'PK' AND is_active = true;
SELECT '清华大学已占用:' as info, COUNT(*) as count FROM signal_codes WHERE school_code = 'QH' AND is_active = true;

-- 查看学校和片区数据
SELECT '活跃的学校:' as info, COUNT(*) as count FROM op_code_schools WHERE is_active = true;
SELECT '活跃的片区:' as info, COUNT(*) as count FROM op_code_areas WHERE is_active = true;

-- 显示部分学校列表
SELECT school_code, school_name, city FROM op_code_schools WHERE city IN ('北京', '上海', '广州', '深圳', '长沙') ORDER BY city, school_code LIMIT 20;