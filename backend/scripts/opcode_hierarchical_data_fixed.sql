-- OP Code 层级系统测试数据（修复版）

-- 插入测试学校数据（使用正确的ID格式）
INSERT INTO op_code_schools (id, school_code, school_name, full_name, city, province, created_at, updated_at) VALUES
-- 北京
(gen_random_uuid()::text, 'PK', '北京大学', '北京大学', '北京', '北京市', NOW(), NOW()),
(gen_random_uuid()::text, 'QH', '清华大学', '清华大学', '北京', '北京市', NOW(), NOW()),
(gen_random_uuid()::text, 'BD', '北京交通大学', '北京交通大学', '北京', '北京市', NOW(), NOW()),
(gen_random_uuid()::text, 'BH', '北京航空航天大学', '北京航空航天大学', '北京', '北京市', NOW(), NOW()),
(gen_random_uuid()::text, 'BN', '北京师范大学', '北京师范大学', '北京', '北京市', NOW(), NOW()),
-- 上海
(gen_random_uuid()::text, 'FD', '复旦大学', '复旦大学', '上海', '上海市', NOW(), NOW()),
(gen_random_uuid()::text, 'SJ', '上海交通大学', '上海交通大学', '上海', '上海市', NOW(), NOW()),
(gen_random_uuid()::text, 'TJ', '同济大学', '同济大学', '上海', '上海市', NOW(), NOW()),
(gen_random_uuid()::text, 'HD', '华东师范大学', '华东师范大学', '上海', '上海市', NOW(), NOW()),
-- 广州
(gen_random_uuid()::text, 'ZS', '中山大学', '中山大学', '广州', '广东省', NOW(), NOW()),
(gen_random_uuid()::text, 'HG', '华南理工大学', '华南理工大学', '广州', '广东省', NOW(), NOW()),
(gen_random_uuid()::text, 'GW', '广东外语外贸大学', '广东外语外贸大学', '广州', '广东省', NOW(), NOW()),
-- 深圳
(gen_random_uuid()::text, 'SZ', '深圳大学', '深圳大学', '深圳', '广东省', NOW(), NOW()),
(gen_random_uuid()::text, 'ST', '南方科技大学', '南方科技大学', '深圳', '广东省', NOW(), NOW()),
-- 长沙
(gen_random_uuid()::text, 'CS', '中南大学', '中南大学', '长沙', '湖南省', NOW(), NOW()),
(gen_random_uuid()::text, 'HU', '湖南大学', '湖南大学', '长沙', '湖南省', NOW(), NOW()),
-- 武汉
(gen_random_uuid()::text, 'WH', '武汉大学', '武汉大学', '武汉', '湖北省', NOW(), NOW()),
(gen_random_uuid()::text, 'HZ', '华中科技大学', '华中科技大学', '武汉', '湖北省', NOW(), NOW()),
-- 成都
(gen_random_uuid()::text, 'SC', '四川大学', '四川大学', '成都', '四川省', NOW(), NOW()),
(gen_random_uuid()::text, 'CD', '电子科技大学', '电子科技大学', '成都', '四川省', NOW(), NOW()),
-- 南京
(gen_random_uuid()::text, 'NJ', '南京大学', '南京大学', '南京', '江苏省', NOW(), NOW()),
(gen_random_uuid()::text, 'DN', '东南大学', '东南大学', '南京', '江苏省', NOW(), NOW()),
-- 杭州
(gen_random_uuid()::text, 'ZJ', '浙江大学', '浙江大学', '杭州', '浙江省', NOW(), NOW()),
(gen_random_uuid()::text, 'HE', '杭州电子科技大学', '杭州电子科技大学', '杭州', '浙江省', NOW(), NOW()),
-- 西安
(gen_random_uuid()::text, 'XJ', '西安交通大学', '西安交通大学', '西安', '陕西省', NOW(), NOW()),
(gen_random_uuid()::text, 'XD', '西安电子科技大学', '西安电子科技大学', '西安', '陕西省', NOW(), NOW())
ON CONFLICT (school_code) DO NOTHING;

-- 检查op_code_areas表结构
-- 假设该表也需要字符串ID，如果不存在则创建
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'op_code_areas') THEN
        CREATE TABLE op_code_areas (
            id VARCHAR(36) PRIMARY KEY,
            school_code VARCHAR(2) NOT NULL,
            area_code VARCHAR(1) NOT NULL,
            area_name VARCHAR(100) NOT NULL,
            description TEXT,
            is_active BOOLEAN DEFAULT true,
            managed_by TEXT,
            parent_id TEXT,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            UNIQUE(school_code, area_code)
        );
    END IF;
END$$;

-- 为北京大学(PK)插入片区数据
INSERT INTO op_code_areas (id, school_code, area_code, area_name, description, created_at, updated_at) VALUES
(gen_random_uuid()::text, 'PK', '1', '东区', '宿舍楼1-5栋，包含本科生宿舍', NOW(), NOW()),
(gen_random_uuid()::text, 'PK', '2', '西区', '宿舍楼6-10栋，包含研究生宿舍', NOW(), NOW()),
(gen_random_uuid()::text, 'PK', '3', '南区', '宿舍楼11-15栋，混合宿舍区', NOW(), NOW()),
(gen_random_uuid()::text, 'PK', '4', '北区', '宿舍楼16-20栋，博士生宿舍', NOW(), NOW()),
(gen_random_uuid()::text, 'PK', '5', '中心区', '教学楼、图书馆、行政楼', NOW(), NOW())
ON CONFLICT (school_code, area_code) DO NOTHING;

-- 为清华大学(QH)插入片区数据
INSERT INTO op_code_areas (id, school_code, area_code, area_name, description, created_at, updated_at) VALUES
(gen_random_uuid()::text, 'QH', 'A', '紫荆园区', '紫荆学生公寓1-23号楼', NOW(), NOW()),
(gen_random_uuid()::text, 'QH', 'B', '西北区', 'W楼群，研究生宿舍', NOW(), NOW()),
(gen_random_uuid()::text, 'QH', 'C', '东区', '东区学生公寓', NOW(), NOW()),
(gen_random_uuid()::text, 'QH', 'D', '南区', '南区学生公寓', NOW(), NOW()),
(gen_random_uuid()::text, 'QH', 'E', '教学区', '主楼、教学楼群', NOW(), NOW())
ON CONFLICT (school_code, area_code) DO NOTHING;

-- 为其他学校插入基本片区数据
INSERT INTO op_code_areas (id, school_code, area_code, area_name, description, created_at, updated_at) VALUES
-- 复旦大学
(gen_random_uuid()::text, 'FD', '1', '本部东区', '东区宿舍群', NOW(), NOW()),
(gen_random_uuid()::text, 'FD', '2', '本部西区', '西区宿舍群', NOW(), NOW()),
(gen_random_uuid()::text, 'FD', '3', '南区', '南区新宿舍', NOW(), NOW()),
-- 中山大学
(gen_random_uuid()::text, 'ZS', '1', '东校区', '东校区学生宿舍', NOW(), NOW()),
(gen_random_uuid()::text, 'ZS', '2', '南校区', '南校区学生宿舍', NOW(), NOW()),
(gen_random_uuid()::text, 'ZS', '3', '北校区', '北校区学生宿舍', NOW(), NOW())
ON CONFLICT (school_code, area_code) DO NOTHING;

-- 创建一些示例的已占用OP Code记录
-- 首先检查signal_codes表结构
-- 只插入必要的字段
INSERT INTO signal_codes (code, school_code, area_code, point_code, description, code_type, is_public, is_active, created_at, updated_at) VALUES
-- 北京大学示例
('PK1A01', 'PK', '1A', '01', '东区A栋101室', 'dormitory', false, true, NOW(), NOW()),
('PK1A02', 'PK', '1A', '02', '东区A栋102室', 'dormitory', false, true, NOW(), NOW()),
('PK1B01', 'PK', '1B', '01', '东区B栋101室', 'dormitory', false, true, NOW(), NOW()),
('PK2A01', 'PK', '2A', '01', '西区A栋101室', 'dormitory', false, true, NOW(), NOW()),
-- 清华大学示例
('QHAA01', 'QH', 'AA', '01', '紫荆1号楼101室', 'dormitory', false, true, NOW(), NOW()),
('QHAA02', 'QH', 'AA', '02', '紫荆1号楼102室', 'dormitory', false, true, NOW(), NOW())
ON CONFLICT (code) DO NOTHING;

-- 输出测试信息
SELECT '新插入的学校数据:' as info, COUNT(*) as count FROM op_code_schools WHERE created_at > NOW() - INTERVAL '1 minute';
SELECT '新插入的片区数据:' as info, COUNT(*) as count FROM op_code_areas WHERE created_at > NOW() - INTERVAL '1 minute';
SELECT '新插入的OP Code:' as info, COUNT(*) as count FROM signal_codes WHERE created_at > NOW() - INTERVAL '1 minute';