-- OpenPenPal Postcode系统测试数据
-- 用于开发和测试环境的稳定数据

-- 清理现有数据（按依赖关系倒序删除）
DELETE FROM postcode_stats;
DELETE FROM postcode_feedbacks;
DELETE FROM postcode_courier_permissions;
DELETE FROM postcode_rooms;
DELETE FROM postcode_buildings;
DELETE FROM postcode_areas;
DELETE FROM postcode_schools;

-- 1. 插入学校数据
INSERT INTO postcode_schools (id, code, name, full_name, status, managed_by, created_at, updated_at) VALUES
('550e8400-e29b-41d4-a716-446655440001', 'PK', '北京大学', '北京大学', 'active', 'courier_level4_001', NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440002', 'TH', '清华大学', '清华大学', 'active', 'courier_level4_002', NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440003', 'BJ', '北京师范大学', '北京师范大学', 'active', 'courier_level4_003', NOW(), NOW()),
('550e8400-e29b-41d4-a716-446655440004', 'RD', '中国人民大学', '中国人民大学', 'active', 'courier_level4_004', NOW(), NOW());

-- 2. 插入片区数据
-- 北京大学片区
INSERT INTO postcode_areas (id, school_code, code, name, description, status, managed_by, created_at, updated_at) VALUES
('660e8400-e29b-41d4-a716-446655440001', 'PK', 'A', '东区', '北京大学东区片区，主要为学生宿舍区', 'active', 'courier_level3_001', NOW(), NOW()),
('660e8400-e29b-41d4-a716-446655440002', 'PK', 'B', '西区', '北京大学西区片区，包含教学楼和办公区', 'active', 'courier_level3_002', NOW(), NOW()),
('660e8400-e29b-41d4-a716-446655440003', 'PK', 'C', '南区', '北京大学南区片区，研究生宿舍和实验室', 'active', 'courier_level3_003', NOW(), NOW());

-- 清华大学片区
INSERT INTO postcode_areas (id, school_code, code, name, description, status, managed_by, created_at, updated_at) VALUES
('660e8400-e29b-41d4-a716-446655440004', 'TH', 'A', '紫荆区', '清华大学紫荆学生公寓区', 'active', 'courier_level3_004', NOW(), NOW()),
('660e8400-e29b-41d4-a716-446655440005', 'TH', 'B', '主楼区', '清华大学主楼教学区', 'active', 'courier_level3_005', NOW(), NOW()),
('660e8400-e29b-41d4-a716-446655440006', 'TH', 'C', '东区', '清华大学东区宿舍区', 'active', 'courier_level3_006', NOW(), NOW());

-- 3. 插入楼栋数据
-- 北京大学东区楼栋
INSERT INTO postcode_buildings (id, school_code, area_code, code, name, type, floors, status, managed_by, created_at, updated_at) VALUES
('770e8400-e29b-41d4-a716-446655440001', 'PK', 'A', '1', '1号楼', 'dormitory', 6, 'active', 'courier_level2_001', NOW(), NOW()),
('770e8400-e29b-41d4-a716-446655440002', 'PK', 'A', '2', '2号楼', 'dormitory', 6, 'active', 'courier_level2_002', NOW(), NOW()),
('770e8400-e29b-41d4-a716-446655440003', 'PK', 'A', '3', '3号楼', 'dormitory', 8, 'active', 'courier_level2_003', NOW(), NOW());

-- 北京大学西区楼栋
INSERT INTO postcode_buildings (id, school_code, area_code, code, name, type, floors, status, managed_by, created_at, updated_at) VALUES
('770e8400-e29b-41d4-a716-446655440004', 'PK', 'B', '1', '教学楼1号', 'teaching', 5, 'active', 'courier_level2_004', NOW(), NOW()),
('770e8400-e29b-41d4-a716-446655440005', 'PK', 'B', '2', '教学楼2号', 'teaching', 4, 'active', 'courier_level2_005', NOW(), NOW());

-- 清华大学紫荆区楼栋
INSERT INTO postcode_buildings (id, school_code, area_code, code, name, type, floors, status, managed_by, created_at, updated_at) VALUES
('770e8400-e29b-41d4-a716-446655440006', 'TH', 'A', '1', '紫荆1号楼', 'dormitory', 12, 'active', 'courier_level2_006', NOW(), NOW()),
('770e8400-e29b-41d4-a716-446655440007', 'TH', 'A', '2', '紫荆2号楼', 'dormitory', 12, 'active', 'courier_level2_007', NOW(), NOW());

-- 4. 插入房间数据
-- 北京大学东区1号楼房间（PKA1XX）
INSERT INTO postcode_rooms (id, school_code, area_code, building_code, code, name, type, capacity, floor, full_postcode, status, managed_by, created_at, updated_at) VALUES
-- 1楼
('880e8400-e29b-41d4-a716-446655440001', 'PK', 'A', '1', '01', '101室', 'dormitory', 4, 1, 'PKA101', 'active', 'courier_level1_001', NOW(), NOW()),
('880e8400-e29b-41d4-a716-446655440002', 'PK', 'A', '1', '02', '102室', 'dormitory', 4, 1, 'PKA102', 'active', 'courier_level1_001', NOW(), NOW()),
('880e8400-e29b-41d4-a716-446655440003', 'PK', 'A', '1', '03', '103室', 'dormitory', 4, 1, 'PKA103', 'active', 'courier_level1_001', NOW(), NOW()),
('880e8400-e29b-41d4-a716-446655440004', 'PK', 'A', '1', '04', '104室', 'dormitory', 4, 1, 'PKA104', 'active', 'courier_level1_001', NOW(), NOW()),
('880e8400-e29b-41d4-a716-446655440005', 'PK', 'A', '1', '05', '105室', 'dormitory', 4, 1, 'PKA105', 'active', 'courier_level1_001', NOW(), NOW()),
-- 2楼
('880e8400-e29b-41d4-a716-446655440006', 'PK', 'A', '1', '06', '201室', 'dormitory', 4, 2, 'PKA106', 'active', 'courier_level1_001', NOW(), NOW()),
('880e8400-e29b-41d4-a716-446655440007', 'PK', 'A', '1', '07', '202室', 'dormitory', 4, 2, 'PKA107', 'active', 'courier_level1_001', NOW(), NOW()),
('880e8400-e29b-41d4-a716-446655440008', 'PK', 'A', '1', '08', '203室', 'dormitory', 4, 2, 'PKA108', 'active', 'courier_level1_001', NOW(), NOW()),
('880e8400-e29b-41d4-a716-446655440009', 'PK', 'A', '1', '09', '204室', 'dormitory', 4, 2, 'PKA109', 'active', 'courier_level1_001', NOW(), NOW()),
('880e8400-e29b-41d4-a716-446655440010', 'PK', 'A', '1', '10', '205室', 'dormitory', 4, 2, 'PKA110', 'active', 'courier_level1_001', NOW(), NOW());

-- 北京大学东区2号楼房间（PKA2XX）
INSERT INTO postcode_rooms (id, school_code, area_code, building_code, code, name, type, capacity, floor, full_postcode, status, managed_by, created_at, updated_at) VALUES
-- 1楼
('880e8400-e29b-41d4-a716-446655440011', 'PK', 'A', '2', '01', '101室', 'dormitory', 4, 1, 'PKA201', 'active', 'courier_level1_002', NOW(), NOW()),
('880e8400-e29b-41d4-a716-446655440012', 'PK', 'A', '2', '02', '102室', 'dormitory', 4, 1, 'PKA202', 'active', 'courier_level1_002', NOW(), NOW()),
('880e8400-e29b-41d4-a716-446655440013', 'PK', 'A', '2', '03', '103室', 'dormitory', 4, 1, 'PKA203', 'active', 'courier_level1_002', NOW(), NOW()),
('880e8400-e29b-41d4-a716-446655440014', 'PK', 'A', '2', '04', '104室', 'dormitory', 4, 1, 'PKA204', 'active', 'courier_level1_002', NOW(), NOW()),
('880e8400-e29b-41d4-a716-446655440015', 'PK', 'A', '2', '05', '105室', 'dormitory', 4, 1, 'PKA205', 'active', 'courier_level1_002', NOW(), NOW());

-- 清华大学紫荆区房间（THA1XX）
INSERT INTO postcode_rooms (id, school_code, area_code, building_code, code, name, type, capacity, floor, full_postcode, status, managed_by, created_at, updated_at) VALUES
-- 1楼
('880e8400-e29b-41d4-a716-446655440016', 'TH', 'A', '1', '01', '101室', 'dormitory', 2, 1, 'THA101', 'active', 'courier_level1_003', NOW(), NOW()),
('880e8400-e29b-41d4-a716-446655440017', 'TH', 'A', '1', '02', '102室', 'dormitory', 2, 1, 'THA102', 'active', 'courier_level1_003', NOW(), NOW()),
('880e8400-e29b-41d4-a716-446655440018', 'TH', 'A', '1', '03', '103室', 'dormitory', 2, 1, 'THA103', 'active', 'courier_level1_003', NOW(), NOW()),
-- 2楼
('880e8400-e29b-41d4-a716-446655440019', 'TH', 'A', '1', '04', '201室', 'dormitory', 2, 2, 'THA104', 'active', 'courier_level1_003', NOW(), NOW()),
('880e8400-e29b-41d4-a716-446655440020', 'TH', 'A', '1', '05', '202室', 'dormitory', 2, 2, 'THA105', 'active', 'courier_level1_003', NOW(), NOW());

-- 5. 插入信使权限数据
INSERT INTO postcode_courier_permissions (id, courier_id, level, prefix_patterns, can_manage, can_create, can_review, created_at, updated_at) VALUES
('990e8400-e29b-41d4-a716-446655440001', 'courier1', 1, ARRAY['PKA1**'], true, false, false, NOW(), NOW()),
('990e8400-e29b-41d4-a716-446655440002', 'courier2', 2, ARRAY['PKA*'], true, true, false, NOW(), NOW()),
('990e8400-e29b-41d4-a716-446655440003', 'courier3', 3, ARRAY['PK*'], true, true, true, NOW(), NOW()),
('990e8400-e29b-41d4-a716-446655440004', 'courier4', 4, ARRAY['**'], true, true, true, NOW(), NOW()),
('990e8400-e29b-41d4-a716-446655440005', 'courier_th_1', 1, ARRAY['THA1**'], true, false, false, NOW(), NOW()),
('990e8400-e29b-41d4-a716-446655440006', 'courier_th_2', 2, ARRAY['THA*'], true, true, false, NOW(), NOW()),
('990e8400-e29b-41d4-a716-446655440007', 'courier_th_3', 3, ARRAY['TH*'], true, true, true, NOW(), NOW());

-- 6. 插入反馈数据
INSERT INTO postcode_feedbacks (id, type, postcode, description, suggested_school_code, suggested_area_code, suggested_building_code, suggested_room_code, suggested_name, submitted_by, submitter_type, status, created_at, updated_at) VALUES
('aa0e8400-e29b-41d4-a716-446655440001', 'new_address', 'PKA301', '新增宿舍楼3栋301室', 'PK', 'A', '3', '01', '3栋301室', 'user_001', 'user', 'pending', NOW(), NOW()),
('aa0e8400-e29b-41d4-a716-446655440002', 'error_report', 'PKA102', '门牌号码不匹配，实际是102A室', 'PK', 'A', '1', '02', '102A室', 'courier1', 'courier', 'approved', NOW(), NOW()),
('aa0e8400-e29b-41d4-a716-446655440003', 'delivery_failed', 'THA201', '投递失败，房间已搬迁', 'TH', 'A', '1', '04', '该房间已空置', 'user_002', 'user', 'pending', NOW(), NOW());

-- 7. 插入使用统计数据
INSERT INTO postcode_stats (id, postcode, delivery_count, error_count, last_used, popularity_score, created_at, updated_at) VALUES
('bb0e8400-e29b-41d4-a716-446655440001', 'PKA101', 150, 2, NOW() - INTERVAL '2 hours', 95.50, NOW(), NOW()),
('bb0e8400-e29b-41d4-a716-446655440002', 'PKA102', 132, 1, NOW() - INTERVAL '3 hours', 92.30, NOW(), NOW()),
('bb0e8400-e29b-41d4-a716-446655440003', 'PKA103', 89, 0, NOW() - INTERVAL '1 hour', 88.70, NOW(), NOW()),
('bb0e8400-e29b-41d4-a716-446655440004', 'PKA201', 76, 3, NOW() - INTERVAL '4 hours', 85.20, NOW(), NOW()),
('bb0e8400-e29b-41d4-a716-446655440005', 'PKA202', 65, 1, NOW() - INTERVAL '5 hours', 82.10, NOW(), NOW()),
('bb0e8400-e29b-41d4-a716-446655440006', 'THA101', 58, 0, NOW() - INTERVAL '1 hour', 90.30, NOW(), NOW()),
('bb0e8400-e29b-41d4-a716-446655440007', 'THA102', 45, 2, NOW() - INTERVAL '6 hours', 78.50, NOW(), NOW()),
('bb0e8400-e29b-41d4-a716-446655440008', 'THA103', 32, 1, NOW() - INTERVAL '8 hours', 75.20, NOW(), NOW());

-- 验证数据插入结果
SELECT 
    'Schools' as table_name, 
    COUNT(*) as record_count 
FROM postcode_schools
UNION ALL
SELECT 
    'Areas' as table_name, 
    COUNT(*) as record_count 
FROM postcode_areas
UNION ALL
SELECT 
    'Buildings' as table_name, 
    COUNT(*) as record_count 
FROM postcode_buildings
UNION ALL
SELECT 
    'Rooms' as table_name, 
    COUNT(*) as record_count 
FROM postcode_rooms
UNION ALL
SELECT 
    'Permissions' as table_name, 
    COUNT(*) as record_count 
FROM postcode_courier_permissions
UNION ALL
SELECT 
    'Feedbacks' as table_name, 
    COUNT(*) as record_count 
FROM postcode_feedbacks
UNION ALL
SELECT 
    'Stats' as table_name, 
    COUNT(*) as record_count 
FROM postcode_stats;

-- 显示一些示例查询结果
SELECT 
    r.full_postcode,
    s.name as school_name,
    a.name as area_name,
    b.name as building_name,
    r.name as room_name,
    r.type,
    r.capacity
FROM postcode_rooms r
JOIN postcode_buildings b ON b.school_code = r.school_code AND b.area_code = r.area_code AND b.code = r.building_code
JOIN postcode_areas a ON a.school_code = r.school_code AND a.code = r.area_code
JOIN postcode_schools s ON s.code = r.school_code
WHERE r.full_postcode IN ('PKA101', 'PKA102', 'THA101', 'THA102')
ORDER BY r.full_postcode;