-- OP Code 层级系统测试数据
-- 创建学校数据（如果不存在）

-- 确保 op_code_schools 表存在
CREATE TABLE IF NOT EXISTS op_code_schools (
    id SERIAL PRIMARY KEY,
    school_code VARCHAR(2) UNIQUE NOT NULL,
    school_name VARCHAR(100) NOT NULL,
    full_name VARCHAR(200),
    city VARCHAR(50) NOT NULL,
    province VARCHAR(50) NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 确保 op_code_areas 表存在
CREATE TABLE IF NOT EXISTS op_code_areas (
    id SERIAL PRIMARY KEY,
    school_code VARCHAR(2) NOT NULL,
    area_code VARCHAR(1) NOT NULL,
    area_name VARCHAR(100) NOT NULL,
    description TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(school_code, area_code)
);

-- 插入测试学校数据
INSERT INTO op_code_schools (school_code, school_name, full_name, city, province) VALUES
-- 北京
('PK', '北京大学', '北京大学', '北京', '北京市'),
('QH', '清华大学', '清华大学', '北京', '北京市'),
('BD', '北京交通大学', '北京交通大学', '北京', '北京市'),
('BH', '北京航空航天大学', '北京航空航天大学', '北京', '北京市'),
('BN', '北京师范大学', '北京师范大学', '北京', '北京市'),
-- 上海
('FD', '复旦大学', '复旦大学', '上海', '上海市'),
('SJ', '上海交通大学', '上海交通大学', '上海', '上海市'),
('TJ', '同济大学', '同济大学', '上海', '上海市'),
('HN', '华东师范大学', '华东师范大学', '上海', '上海市'),
-- 广州
('ZS', '中山大学', '中山大学', '广州', '广东省'),
('HG', '华南理工大学', '华南理工大学', '广州', '广东省'),
('GD', '广东外语外贸大学', '广东外语外贸大学', '广州', '广东省'),
-- 深圳
('SZ', '深圳大学', '深圳大学', '深圳', '广东省'),
('ST', '南方科技大学', '南方科技大学', '深圳', '广东省'),
-- 长沙
('ZN', '中南大学', '中南大学', '长沙', '湖南省'),
('HN', '湖南大学', '湖南大学', '长沙', '湖南省'),
-- 武汉
('WH', '武汉大学', '武汉大学', '武汉', '湖北省'),
('HZ', '华中科技大学', '华中科技大学', '武汉', '湖北省'),
-- 成都
('SC', '四川大学', '四川大学', '成都', '四川省'),
('CD', '电子科技大学', '电子科技大学', '成都', '四川省'),
-- 南京
('NJ', '南京大学', '南京大学', '南京', '江苏省'),
('DN', '东南大学', '东南大学', '南京', '江苏省'),
-- 杭州
('ZJ', '浙江大学', '浙江大学', '杭州', '浙江省'),
('HZ', '杭州电子科技大学', '杭州电子科技大学', '杭州', '浙江省'),
-- 西安
('XJ', '西安交通大学', '西安交通大学', '西安', '陕西省'),
('XD', '西安电子科技大学', '西安电子科技大学', '西安', '陕西省')
ON CONFLICT (school_code) DO NOTHING;

-- 为北京大学(PK)插入片区数据
INSERT INTO op_code_areas (school_code, area_code, area_name, description) VALUES
('PK', '1', '东区', '宿舍楼1-5栋，包含本科生宿舍'),
('PK', '2', '西区', '宿舍楼6-10栋，包含研究生宿舍'),
('PK', '3', '南区', '宿舍楼11-15栋，混合宿舍区'),
('PK', '4', '北区', '宿舍楼16-20栋，博士生宿舍'),
('PK', '5', '中心区', '教学楼、图书馆、行政楼')
ON CONFLICT (school_code, area_code) DO NOTHING;

-- 为清华大学(QH)插入片区数据
INSERT INTO op_code_areas (school_code, area_code, area_name, description) VALUES
('QH', 'A', '紫荆园区', '紫荆学生公寓1-23号楼'),
('QH', 'B', '西北区', 'W楼群，研究生宿舍'),
('QH', 'C', '东区', '东区学生公寓'),
('QH', 'D', '南区', '南区学生公寓'),
('QH', 'E', '教学区', '主楼、教学楼群')
ON CONFLICT (school_code, area_code) DO NOTHING;

-- 为其他学校插入基本片区数据
INSERT INTO op_code_areas (school_code, area_code, area_name, description) VALUES
-- 复旦大学
('FD', '1', '本部东区', '东区宿舍群'),
('FD', '2', '本部西区', '西区宿舍群'),
('FD', '3', '南区', '南区新宿舍'),
-- 中山大学
('ZS', '1', '东校区', '东校区学生宿舍'),
('ZS', '2', '南校区', '南校区学生宿舍'),
('ZS', '3', '北校区', '北校区学生宿舍')
ON CONFLICT (school_code, area_code) DO NOTHING;

-- 创建一些示例的已占用OP Code记录
INSERT INTO signal_codes (code, school_code, area_code, point_code, description, code_type, is_public, is_active, binding_type, created_at, updated_at) VALUES
-- 北京大学示例
('PK1A01', 'PK', '1A', '01', '东区A栋101室', 'dormitory', false, true, 'user', NOW(), NOW()),
('PK1A02', 'PK', '1A', '02', '东区A栋102室', 'dormitory', false, true, 'user', NOW(), NOW()),
('PK1B01', 'PK', '1B', '01', '东区B栋101室', 'dormitory', false, true, 'user', NOW(), NOW()),
('PK2A01', 'PK', '2A', '01', '西区A栋101室', 'dormitory', false, true, 'user', NOW(), NOW()),
-- 清华大学示例
('QHAA01', 'QH', 'AA', '01', '紫荆1号楼101室', 'dormitory', false, true, 'user', NOW(), NOW()),
('QHAA02', 'QH', 'AA', '02', '紫荆1号楼102室', 'dormitory', false, true, 'user', NOW(), NOW())
ON CONFLICT (code) DO NOTHING;

-- 输出测试信息
SELECT '已插入学校数据:' as info, COUNT(*) as count FROM op_code_schools WHERE is_active = true;
SELECT '已插入片区数据:' as info, COUNT(*) as count FROM op_code_areas WHERE is_active = true;
SELECT '已占用的OP Code:' as info, COUNT(*) as count FROM signal_codes WHERE is_active = true;