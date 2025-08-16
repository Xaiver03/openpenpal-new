-- Migration: Create OPCode system tables
-- This creates the comprehensive OPCode system for 6-digit location encoding
-- Format: AABBCC where AA=school, BB=area, CC=point

-- ==========================================
-- OPCode Schools Table (学校编码映射)
-- ==========================================
CREATE TABLE IF NOT EXISTS op_code_schools (
    id VARCHAR(36) PRIMARY KEY,
    school_code VARCHAR(2) UNIQUE NOT NULL,
    school_name VARCHAR(100) NOT NULL,
    full_name VARCHAR(200),
    city VARCHAR(50),
    province VARCHAR(50),
    is_active BOOLEAN DEFAULT TRUE,
    managed_by VARCHAR(36), -- 四级信使ID
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_school_code (school_code),
    INDEX idx_city (city),
    INDEX idx_province (province),
    INDEX idx_managed_by (managed_by)
);

-- ==========================================
-- OPCode Areas Table (片区编码映射)
-- ==========================================
CREATE TABLE IF NOT EXISTS op_code_areas (
    id VARCHAR(36) PRIMARY KEY,
    school_code VARCHAR(2) NOT NULL,
    area_code VARCHAR(2) NOT NULL,
    area_name VARCHAR(100) NOT NULL,
    description VARCHAR(200),
    is_active BOOLEAN DEFAULT TRUE,
    managed_by VARCHAR(36), -- 三级信使ID
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE KEY unique_school_area (school_code, area_code),
    INDEX idx_school_code (school_code),
    INDEX idx_area_code (area_code),
    INDEX idx_managed_by (managed_by),
    FOREIGN KEY (school_code) REFERENCES op_code_schools(school_code)
);

-- ==========================================
-- OPCodes Main Table (重用signal_codes表结构)
-- ==========================================
CREATE TABLE IF NOT EXISTS op_codes (
    id VARCHAR(36) PRIMARY KEY,
    code VARCHAR(6) UNIQUE NOT NULL, -- 完整6位编码，如: PK5F3D
    school_code VARCHAR(2) NOT NULL, -- 前2位: 学校代码
    area_code VARCHAR(2) NOT NULL,   -- 中2位: 片区/楼栋代码
    point_code VARCHAR(2) NOT NULL,  -- 后2位: 具体位置代码
    
    -- 类型和属性
    point_type VARCHAR(20) NOT NULL, -- 类型: dormitory/shop/box/club
    point_name VARCHAR(100),         -- 位置名称
    full_address VARCHAR(200),       -- 完整地址描述
    is_public BOOLEAN DEFAULT FALSE, -- 后两位是否公开
    is_active BOOLEAN DEFAULT TRUE,  -- 是否激活
    
    -- 绑定信息
    binding_type VARCHAR(20),                    -- 绑定类型: user/shop/public
    binding_id VARCHAR(36),                      -- 绑定对象ID
    binding_status VARCHAR(20) DEFAULT 'pending', -- 绑定状态: pending/approved/rejected
    
    -- 管理信息
    managed_by VARCHAR(36) NOT NULL,  -- 管理者ID (二级信使)
    approved_by VARCHAR(36),          -- 审核者ID
    approved_at TIMESTAMP,            -- 审核时间
    
    -- 使用统计
    usage_count INT DEFAULT 0,        -- 使用次数
    last_used_at TIMESTAMP,           -- 最后使用时间
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_code (code),
    INDEX idx_school_code (school_code),
    INDEX idx_area_code (area_code),
    INDEX idx_point_code (point_code),
    INDEX idx_point_type (point_type),
    INDEX idx_is_active (is_active),
    INDEX idx_is_public (is_public),
    INDEX idx_managed_by (managed_by),
    
    FOREIGN KEY (school_code) REFERENCES op_code_schools(school_code),
    FOREIGN KEY (school_code, area_code) REFERENCES op_code_areas(school_code, area_code)
);

-- ==========================================
-- OPCode Applications Table (申请记录)
-- ==========================================
CREATE TABLE IF NOT EXISTS op_code_applications (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    requested_code VARCHAR(6), -- 申请的完整编码
    school_code VARCHAR(2) NOT NULL,
    area_code VARCHAR(2) NOT NULL,
    point_type VARCHAR(20) NOT NULL,
    point_name VARCHAR(100),
    full_address VARCHAR(200),
    reason TEXT,
    evidence JSON, -- 证明材料JSON
    
    status VARCHAR(20) DEFAULT 'pending', -- pending/approved/rejected
    assigned_code VARCHAR(6), -- 最终分配的编码
    reviewer_id VARCHAR(36),
    review_comment TEXT,
    reviewed_at TIMESTAMP,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_user_id (user_id),
    INDEX idx_status (status),
    INDEX idx_school_code (school_code),
    INDEX idx_reviewer_id (reviewer_id),
    
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (school_code) REFERENCES op_code_schools(school_code)
);

-- ==========================================
-- OPCode Permissions Table (权限表)
-- ==========================================
CREATE TABLE IF NOT EXISTS op_code_permissions (
    id VARCHAR(36) PRIMARY KEY,
    courier_id VARCHAR(36) NOT NULL,
    courier_level INT NOT NULL,
    code_prefix VARCHAR(6) NOT NULL, -- 管理的编码前缀
    permission VARCHAR(20) NOT NULL, -- view/assign/approve
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_courier_id (courier_id),
    INDEX idx_code_prefix (code_prefix),
    INDEX idx_courier_level (courier_level),
    
    FOREIGN KEY (courier_id) REFERENCES couriers(id)
);

-- ==========================================
-- Insert Initial Data
-- ==========================================

-- 添加一些示例学校数据（将通过CSV导入更多数据）
INSERT IGNORE INTO op_code_schools (id, school_code, school_name, full_name, city, province, managed_by) VALUES
('school_pku', 'PK', '北京大学', '北京大学', '北京', '北京市', 'admin'),
('school_tsinghua', 'QH', '清华大学', '清华大学', '北京', '北京市', 'admin'),
('school_ruc', 'RU', '中国人民大学', '中国人民大学', '北京', '北京市', 'admin'),
('school_bnu', 'BN', '北京师范大学', '北京师范大学', '北京', '北京市', 'admin');

-- 添加一些示例片区数据
INSERT IGNORE INTO op_code_areas (id, school_code, area_code, area_name, description, managed_by) VALUES
('area_pk_5f', 'PK', '5F', '5号楼', '学生宿舍5号楼', 'admin'),
('area_pk_3d', 'PK', '3D', '3号食堂', '学生食堂3号', 'admin'),
('area_pk_2g', 'PK', '2G', '2号门', '校园2号门岗', 'admin'),
('area_qh_1a', 'QH', '1A', '1号楼A区', '教学楼1号楼A区', 'admin'),
('area_qh_4b', 'QH', '4B', '4号楼B区', '学生宿舍4号楼B区', 'admin');

-- 添加一些示例OPCode数据
INSERT IGNORE INTO op_codes (id, code, school_code, area_code, point_code, point_type, point_name, full_address, is_public, managed_by) VALUES
('opcode_pk5f3d', 'PK5F3D', 'PK', '5F', '3D', 'dormitory', '5号楼303室', '北京大学5号楼303室', FALSE, 'admin'),
('opcode_pk3d01', 'PK3D01', 'PK', '3D', '01', 'shop', '3号食堂1号窗口', '北京大学3号食堂1号窗口', TRUE, 'admin'),
('opcode_pk2g01', 'PK2G01', 'PK', '2G', '01', 'box', '2号门投递箱', '北京大学2号门投递箱', TRUE, 'admin'),
('opcode_qh1a01', 'QH1A01', 'QH', '1A', '01', 'club', '1号楼A区活动室', '清华大学1号楼A区活动室', TRUE, 'admin'),
('opcode_qh4b12', 'QH4B12', 'QH', '4B', '12', 'dormitory', '4号楼B区402室', '清华大学4号楼B区402室', FALSE, 'admin');

-- ==========================================
-- Views for Easy Querying
-- ==========================================

-- 创建视图：完整的OPCode信息
CREATE OR REPLACE VIEW v_opcode_full AS
SELECT 
    oc.id,
    oc.code,
    oc.school_code,
    os.school_name,
    os.city,
    os.province,
    oc.area_code,
    oa.area_name,
    oc.point_code,
    oc.point_type,
    oc.point_name,
    oc.full_address,
    oc.is_public,
    oc.is_active,
    oc.usage_count,
    oc.created_at,
    oc.updated_at
FROM op_codes oc
LEFT JOIN op_code_schools os ON oc.school_code = os.school_code
LEFT JOIN op_code_areas oa ON oc.school_code = oa.school_code AND oc.area_code = oa.area_code
WHERE oc.is_active = TRUE;

-- 创建视图：学校统计
CREATE OR REPLACE VIEW v_school_opcode_stats AS
SELECT 
    os.school_code,
    os.school_name,
    os.city,
    os.province,
    COUNT(oc.id) as total_opcodes,
    COUNT(CASE WHEN oc.is_active = TRUE THEN 1 END) as active_opcodes,
    COUNT(CASE WHEN oc.is_public = TRUE THEN 1 END) as public_opcodes,
    COUNT(CASE WHEN oc.point_type = 'dormitory' THEN 1 END) as dormitory_count,
    COUNT(CASE WHEN oc.point_type = 'shop' THEN 1 END) as shop_count,
    COUNT(CASE WHEN oc.point_type = 'box' THEN 1 END) as box_count,
    COUNT(CASE WHEN oc.point_type = 'club' THEN 1 END) as club_count
FROM op_code_schools os
LEFT JOIN op_codes oc ON os.school_code = oc.school_code
GROUP BY os.school_code, os.school_name, os.city, os.province;

-- ==========================================
-- Migration Complete
-- ==========================================