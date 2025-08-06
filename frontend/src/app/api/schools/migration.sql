-- OpenPenPal Schools Master Data Migration
-- 创建学校主数据表和相关管理接口

-- 学校主数据表
CREATE TABLE IF NOT EXISTS schools (
    -- 主键和基本信息
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(20) NOT NULL UNIQUE, -- 6位学校编码，如 BJDX01
    name VARCHAR(100) NOT NULL, -- 学校简称，如 北京大学
    full_name VARCHAR(200) NOT NULL, -- 学校全称
    english_name VARCHAR(200), -- 英文名称
    
    -- 地理位置信息
    province VARCHAR(50) NOT NULL,
    city VARCHAR(50) NOT NULL,
    district VARCHAR(50),
    address TEXT,
    postal_code VARCHAR(10),
    coordinates POINT, -- 经纬度坐标，用于地图显示
    
    -- 学校属性
    school_type VARCHAR(50) NOT NULL, -- 综合类、理工类、师范类等
    school_level VARCHAR(50) DEFAULT '本科', -- 本科、专科、研究生院等
    is_985 BOOLEAN DEFAULT FALSE,
    is_211 BOOLEAN DEFAULT FALSE,
    is_double_first_class BOOLEAN DEFAULT FALSE, -- 双一流
    
    -- 机构信息
    established_year INTEGER, -- 建校年份
    website VARCHAR(255),
    official_email VARCHAR(100),
    phone VARCHAR(20),
    logo_url VARCHAR(500),
    
    -- 管理信息
    status VARCHAR(20) DEFAULT 'active' NOT NULL, -- active, inactive, pending
    verification_status VARCHAR(20) DEFAULT 'verified', -- verified, pending, rejected
    created_by VARCHAR(50), -- 创建者ID
    verified_by VARCHAR(50), -- 审核者ID
    verified_at TIMESTAMP WITH TIME ZONE,
    
    -- 统计信息
    student_count INTEGER DEFAULT 0,
    user_count INTEGER DEFAULT 0, -- 平台用户数
    letter_count INTEGER DEFAULT 0, -- 信件数
    courier_count INTEGER DEFAULT 0, -- 信使数
    
    -- 时间戳
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- 学校院系表
CREATE TABLE IF NOT EXISTS school_departments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    code VARCHAR(20) NOT NULL, -- 院系编码
    name VARCHAR(100) NOT NULL,
    full_name VARCHAR(200),
    type VARCHAR(50), -- 学院、系、部门等
    parent_id UUID REFERENCES school_departments(id) ON DELETE SET NULL,
    contact_email VARCHAR(100),
    contact_phone VARCHAR(20),
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- 复合唯一约束
    UNIQUE(school_id, code)
);

-- 学校管理员表
CREATE TABLE IF NOT EXISTS school_admins (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    user_id VARCHAR(50) NOT NULL, -- 关联用户ID
    role VARCHAR(50) DEFAULT 'admin', -- admin, coordinator, reviewer
    permissions TEXT[], -- 权限列表
    department_id UUID REFERENCES school_departments(id) ON DELETE SET NULL,
    is_active BOOLEAN DEFAULT TRUE,
    assigned_by VARCHAR(50), -- 指派者ID
    assigned_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- 复合唯一约束
    UNIQUE(school_id, user_id)
);

-- 学校配置表
CREATE TABLE IF NOT EXISTS school_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id UUID NOT NULL REFERENCES schools(id) ON DELETE CASCADE,
    setting_key VARCHAR(100) NOT NULL,
    setting_value TEXT,
    setting_type VARCHAR(20) DEFAULT 'string', -- string, number, boolean, json
    description VARCHAR(500),
    is_public BOOLEAN DEFAULT FALSE, -- 是否公开显示
    updated_by VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- 复合唯一约束
    UNIQUE(school_id, setting_key)
);

-- 学校申请审核表
CREATE TABLE IF NOT EXISTS school_applications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    applicant_id VARCHAR(50) NOT NULL, -- 申请人ID
    school_name VARCHAR(200) NOT NULL,
    school_full_name VARCHAR(200),
    province VARCHAR(50) NOT NULL,
    city VARCHAR(50) NOT NULL,
    school_type VARCHAR(50),
    website VARCHAR(255),
    contact_email VARCHAR(100),
    contact_phone VARCHAR(20),
    
    -- 申请材料
    description TEXT,
    supporting_documents TEXT[], -- 支持文档链接
    verification_documents TEXT[], -- 验证文档
    
    -- 审核状态
    status VARCHAR(20) DEFAULT 'pending', -- pending, approved, rejected, need_more_info
    reviewer_id VARCHAR(50),
    review_note TEXT,
    reviewed_at TIMESTAMP WITH TIME ZONE,
    
    -- 生成的学校数据
    generated_school_id UUID REFERENCES schools(id) ON DELETE SET NULL,
    generated_code VARCHAR(20),
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- ================================================================
-- 索引优化
-- ================================================================

-- 学校表索引
CREATE INDEX IF NOT EXISTS idx_schools_code ON schools(code);
CREATE INDEX IF NOT EXISTS idx_schools_name ON schools(name);
CREATE INDEX IF NOT EXISTS idx_schools_province_city ON schools(province, city);
CREATE INDEX IF NOT EXISTS idx_schools_type ON schools(school_type);
CREATE INDEX IF NOT EXISTS idx_schools_status ON schools(status);
CREATE INDEX IF NOT EXISTS idx_schools_verification ON schools(verification_status);
CREATE INDEX IF NOT EXISTS idx_schools_level ON schools(school_level);
CREATE INDEX IF NOT EXISTS idx_schools_created_at ON schools(created_at DESC);

-- 地理位置索引（PostGIS如果可用）
CREATE INDEX IF NOT EXISTS idx_schools_coordinates ON schools USING gist(coordinates);

-- 复合索引
CREATE INDEX IF NOT EXISTS idx_schools_active_province ON schools(status, province) WHERE status = 'active';
CREATE INDEX IF NOT EXISTS idx_schools_type_level ON schools(school_type, school_level);

-- 全文搜索索引
CREATE INDEX IF NOT EXISTS idx_schools_name_search ON schools USING gin(to_tsvector('simple', name || ' ' || full_name || ' ' || english_name));

-- 院系表索引
CREATE INDEX IF NOT EXISTS idx_school_departments_school_id ON school_departments(school_id);
CREATE INDEX IF NOT EXISTS idx_school_departments_code ON school_departments(code);
CREATE INDEX IF NOT EXISTS idx_school_departments_parent ON school_departments(parent_id);
CREATE INDEX IF NOT EXISTS idx_school_departments_active ON school_departments(is_active);

-- 管理员表索引
CREATE INDEX IF NOT EXISTS idx_school_admins_school_id ON school_admins(school_id);
CREATE INDEX IF NOT EXISTS idx_school_admins_user_id ON school_admins(user_id);
CREATE INDEX IF NOT EXISTS idx_school_admins_role ON school_admins(role);
CREATE INDEX IF NOT EXISTS idx_school_admins_active ON school_admins(is_active);

-- 配置表索引
CREATE INDEX IF NOT EXISTS idx_school_settings_school_id ON school_settings(school_id);
CREATE INDEX IF NOT EXISTS idx_school_settings_key ON school_settings(setting_key);
CREATE INDEX IF NOT EXISTS idx_school_settings_public ON school_settings(is_public);

-- 申请表索引
CREATE INDEX IF NOT EXISTS idx_school_applications_applicant ON school_applications(applicant_id);
CREATE INDEX IF NOT EXISTS idx_school_applications_status ON school_applications(status);
CREATE INDEX IF NOT EXISTS idx_school_applications_reviewer ON school_applications(reviewer_id);
CREATE INDEX IF NOT EXISTS idx_school_applications_created ON school_applications(created_at DESC);

-- ================================================================
-- 约束和检查
-- ================================================================

-- 学校编码格式检查
ALTER TABLE schools ADD CONSTRAINT chk_schools_code_format 
CHECK (code ~ '^[A-Z0-9]{4,10}$');

-- 状态检查
ALTER TABLE schools ADD CONSTRAINT chk_schools_status 
CHECK (status IN ('active', 'inactive', 'pending', 'suspended'));

ALTER TABLE schools ADD CONSTRAINT chk_schools_verification_status 
CHECK (verification_status IN ('verified', 'pending', 'rejected', 'need_verification'));

-- 学校类型检查
ALTER TABLE schools ADD CONSTRAINT chk_schools_type 
CHECK (school_type IN ('综合类', '理工类', '文科类', '师范类', '农林类', '医药类', '艺术类', '体育类', '军事类', '民族类', '其他'));

-- 学校层次检查
ALTER TABLE schools ADD CONSTRAINT chk_schools_level 
CHECK (school_level IN ('本科', '专科', '研究生院', '技工学校', '中等专业学校', '其他'));

-- 申请状态检查
ALTER TABLE school_applications ADD CONSTRAINT chk_school_applications_status 
CHECK (status IN ('pending', 'approved', 'rejected', 'need_more_info', 'withdrawn'));

-- ================================================================
-- 触发器
-- ================================================================

-- 自动更新时间戳触发器
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 为所有表添加更新时间触发器
CREATE TRIGGER update_schools_updated_at 
    BEFORE UPDATE ON schools 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_school_departments_updated_at 
    BEFORE UPDATE ON school_departments 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_school_admins_updated_at 
    BEFORE UPDATE ON school_admins 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_school_settings_updated_at 
    BEFORE UPDATE ON school_settings 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_school_applications_updated_at 
    BEFORE UPDATE ON school_applications 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 学校统计更新触发器
CREATE OR REPLACE FUNCTION update_school_stats()
RETURNS TRIGGER AS $$
BEGIN
    -- 更新用户统计
    IF TG_OP = 'INSERT' THEN
        UPDATE schools SET user_count = user_count + 1 
        WHERE code = NEW.school_code;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE schools SET user_count = user_count - 1 
        WHERE code = OLD.school_code AND user_count > 0;
    ELSIF TG_OP = 'UPDATE' AND OLD.school_code != NEW.school_code THEN
        UPDATE schools SET user_count = user_count - 1 
        WHERE code = OLD.school_code AND user_count > 0;
        UPDATE schools SET user_count = user_count + 1 
        WHERE code = NEW.school_code;
    END IF;
    
    RETURN COALESCE(NEW, OLD);
END;
$$ language 'plpgsql';

-- 注意：这个触发器需要在users表上创建，这里只是示例
-- CREATE TRIGGER update_school_user_stats
--     AFTER INSERT OR UPDATE OR DELETE ON users
--     FOR EACH ROW EXECUTE FUNCTION update_school_stats();

-- ================================================================
-- 视图
-- ================================================================

-- 学校统计视图
CREATE OR REPLACE VIEW school_statistics AS
SELECT 
    s.id,
    s.code,
    s.name,
    s.province,
    s.city,
    s.school_type,
    s.school_level,
    s.user_count,
    s.letter_count,
    s.courier_count,
    COUNT(DISTINCT sd.id) as department_count,
    COUNT(DISTINCT sa.id) as admin_count,
    s.created_at,
    s.updated_at
FROM schools s
LEFT JOIN school_departments sd ON s.id = sd.school_id AND sd.is_active = true
LEFT JOIN school_admins sa ON s.id = sa.school_id AND sa.is_active = true
WHERE s.deleted_at IS NULL
GROUP BY s.id, s.code, s.name, s.province, s.city, s.school_type, s.school_level, 
         s.user_count, s.letter_count, s.courier_count, s.created_at, s.updated_at;

-- 学校审核待办视图
CREATE OR REPLACE VIEW pending_school_reviews AS
SELECT 
    sa.id,
    sa.school_name,
    sa.province,
    sa.city,
    sa.school_type,
    sa.contact_email,
    sa.applicant_id,
    sa.description,
    sa.status,
    sa.created_at,
    EXTRACT(DAY FROM CURRENT_TIMESTAMP - sa.created_at) as pending_days
FROM school_applications sa
WHERE sa.status = 'pending'
ORDER BY sa.created_at ASC;

-- ================================================================
-- 初始化数据
-- ================================================================

-- 插入基础学校数据
INSERT INTO schools (
    code, name, full_name, english_name, province, city, 
    school_type, school_level, is_985, is_211, is_double_first_class,
    established_year, website, status, verification_status
) VALUES 
-- 北京地区
('BJDX01', '北京大学', '北京大学', 'Peking University', '北京', '北京', 
 '综合类', '本科', true, true, true, 1898, 'https://www.pku.edu.cn', 'active', 'verified'),
 
('QHDX01', '清华大学', '清华大学', 'Tsinghua University', '北京', '北京', 
 '理工类', '本科', true, true, true, 1911, 'https://www.tsinghua.edu.cn', 'active', 'verified'),
 
('BJLG01', '北京理工大学', '北京理工大学', 'Beijing Institute of Technology', '北京', '北京', 
 '理工类', '本科', false, true, true, 1940, 'https://www.bit.edu.cn', 'active', 'verified'),
 
('BJHG01', '北京航空航天大学', '北京航空航天大学', 'Beihang University', '北京', '北京', 
 '理工类', '本科', false, true, true, 1952, 'https://www.buaa.edu.cn', 'active', 'verified'),

-- 上海地区
('FDDX01', '复旦大学', '复旦大学', 'Fudan University', '上海', '上海', 
 '综合类', '本科', true, true, true, 1905, 'https://www.fudan.edu.cn', 'active', 'verified'),
 
('JDDX01', '上海交通大学', '上海交通大学', 'Shanghai Jiao Tong University', '上海', '上海', 
 '理工类', '本科', true, true, true, 1896, 'https://www.sjtu.edu.cn', 'active', 'verified'),

-- 其他重点城市
('ZJDX01', '浙江大学', '浙江大学', 'Zhejiang University', '浙江', '杭州', 
 '综合类', '本科', true, true, true, 1897, 'https://www.zju.edu.cn', 'active', 'verified'),
 
('NJDX01', '南京大学', '南京大学', 'Nanjing University', '江苏', '南京', 
 '综合类', '本科', true, true, true, 1902, 'https://www.nju.edu.cn', 'active', 'verified'),
 
('HZDX01', '华中科技大学', '华中科技大学', 'Huazhong University of Science and Technology', '湖北', '武汉', 
 '理工类', '本科', true, true, true, 1952, 'https://www.hust.edu.cn', 'active', 'verified'),
 
('XADX01', '西安交通大学', '西安交通大学', 'Xi\'an Jiaotong University', '陕西', '西安', 
 '理工类', '本科', true, true, true, 1896, 'https://www.xjtu.edu.cn', 'active', 'verified'),
 
('SCDX01', '四川大学', '四川大学', 'Sichuan University', '四川', '成都', 
 '综合类', '本科', true, true, true, 1896, 'https://www.scu.edu.cn', 'active', 'verified'),
 
('ZSDX01', '中山大学', '中山大学', 'Sun Yat-sen University', '广东', '广州', 
 '综合类', '本科', true, true, true, 1924, 'https://www.sysu.edu.cn', 'active', 'verified'),
 
('HNDX01', '湖南大学', '湖南大学', 'Hunan University', '湖南', '长沙', 
 '综合类', '本科', false, true, true, 976, 'https://www.hnu.edu.cn', 'active', 'verified'),
 
('DLDX01', '大连理工大学', '大连理工大学', 'Dalian University of Technology', '辽宁', '大连', 
 '理工类', '本科', false, true, true, 1949, 'https://www.dlut.edu.cn', 'active', 'verified'),
 
('TJDX01', '天津大学', '天津大学', 'Tianjin University', '天津', '天津', 
 '理工类', '本科', false, true, true, 1895, 'https://www.tju.edu.cn', 'active', 'verified')

ON CONFLICT (code) DO NOTHING;

-- 插入示例院系数据
INSERT INTO school_departments (school_id, code, name, full_name, type) 
SELECT s.id, 'CS', '计算机学院', '计算机科学与技术学院', '学院'
FROM schools s WHERE s.code IN ('BJDX01', 'QHDX01', 'FDDX01', 'JDDX01')
ON CONFLICT (school_id, code) DO NOTHING;

INSERT INTO school_departments (school_id, code, name, full_name, type) 
SELECT s.id, 'EE', '电子工程学院', '电子与信息工程学院', '学院'
FROM schools s WHERE s.code IN ('QHDX01', 'BJHG01', 'JDDX01', 'XADX01')
ON CONFLICT (school_id, code) DO NOTHING;

-- 插入默认学校配置
INSERT INTO school_settings (school_id, setting_key, setting_value, setting_type, description, is_public)
SELECT s.id, 'allow_public_letters', 'true', 'boolean', '是否允许公开信件', true
FROM schools s WHERE s.status = 'active'
ON CONFLICT (school_id, setting_key) DO NOTHING;

INSERT INTO school_settings (school_id, setting_key, setting_value, setting_type, description, is_public)
SELECT s.id, 'max_letters_per_day', '10', 'number', '每日最大信件数量', false
FROM schools s WHERE s.status = 'active'
ON CONFLICT (school_id, setting_key) DO NOTHING;

-- ================================================================
-- 权限和安全
-- ================================================================

-- 创建只读用户（可选）
-- CREATE ROLE school_readonly;
-- GRANT SELECT ON schools, school_departments, school_statistics TO school_readonly;

-- 创建学校管理员角色（可选）
-- CREATE ROLE school_admin;
-- GRANT SELECT, INSERT, UPDATE ON schools, school_departments, school_admins, school_settings TO school_admin;
-- GRANT SELECT ON school_applications TO school_admin;

-- 添加注释
COMMENT ON TABLE schools IS '学校主数据表 - 存储所有合作学校的基本信息';
COMMENT ON TABLE school_departments IS '学校院系表 - 存储学校的院系结构';
COMMENT ON TABLE school_admins IS '学校管理员表 - 管理学校级别的管理员权限';
COMMENT ON TABLE school_settings IS '学校配置表 - 存储学校特定的配置选项';
COMMENT ON TABLE school_applications IS '学校申请审核表 - 处理新学校的加入申请';
COMMENT ON VIEW school_statistics IS '学校统计视图 - 提供学校的统计信息概览';
COMMENT ON VIEW pending_school_reviews IS '待审核学校视图 - 显示需要审核的学校申请';