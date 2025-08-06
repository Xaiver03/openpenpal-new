-- OpenPenPal Postcode 地址编码系统数据库表结构
-- 基于四级信使体系设计的分层地址管理

-- 1. 学校站点表
CREATE TABLE IF NOT EXISTS postcode_schools (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(2) NOT NULL UNIQUE, -- 2位学校编码 (如: PK)
    name VARCHAR(100) NOT NULL,
    full_name VARCHAR(200) NOT NULL,
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
    managed_by VARCHAR(100), -- 四级信使ID
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 2. 片区表
CREATE TABLE IF NOT EXISTS postcode_areas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_code VARCHAR(2) NOT NULL REFERENCES postcode_schools(code) ON DELETE CASCADE,
    code VARCHAR(1) NOT NULL, -- 1位片区编码 (如: 5)
    name VARCHAR(100) NOT NULL,
    description TEXT,
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
    managed_by VARCHAR(100), -- 三级信使ID
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(school_code, code)
);

-- 3. 楼栋表
CREATE TABLE IF NOT EXISTS postcode_buildings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_code VARCHAR(2) NOT NULL,
    area_code VARCHAR(1) NOT NULL,
    code VARCHAR(1) NOT NULL, -- 1位楼栋编码 (如: F)
    name VARCHAR(100) NOT NULL,
    type VARCHAR(20) DEFAULT 'dormitory' CHECK (type IN ('dormitory', 'teaching', 'office', 'other')),
    floors INTEGER,
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
    managed_by VARCHAR(100), -- 二级信使ID
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (school_code, area_code) REFERENCES postcode_areas(school_code, code) ON DELETE CASCADE,
    UNIQUE(school_code, area_code, code)
);

-- 4. 房间表
CREATE TABLE IF NOT EXISTS postcode_rooms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_code VARCHAR(2) NOT NULL,
    area_code VARCHAR(1) NOT NULL,
    building_code VARCHAR(1) NOT NULL,
    code VARCHAR(2) NOT NULL, -- 2位房间编码 (如: 3D)
    name VARCHAR(100) NOT NULL,
    type VARCHAR(20) DEFAULT 'dormitory' CHECK (type IN ('dormitory', 'classroom', 'office', 'other')),
    capacity INTEGER,
    floor INTEGER,
    full_postcode VARCHAR(6) GENERATED ALWAYS AS (school_code || area_code || building_code || code) STORED,
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
    managed_by VARCHAR(100), -- 一级信使ID
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (school_code, area_code, building_code) REFERENCES postcode_buildings(school_code, area_code, code) ON DELETE CASCADE,
    UNIQUE(school_code, area_code, building_code, code)
);

-- 5. 信使Postcode权限表
CREATE TABLE IF NOT EXISTS postcode_courier_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    courier_id VARCHAR(100) NOT NULL,
    level INTEGER NOT NULL CHECK (level IN (1, 2, 3, 4)),
    prefix_patterns TEXT[] NOT NULL, -- 权限前缀数组
    can_manage BOOLEAN DEFAULT FALSE,
    can_create BOOLEAN DEFAULT FALSE,
    can_review BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(courier_id)
);

-- 6. 地址反馈表
CREATE TABLE IF NOT EXISTS postcode_feedbacks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type VARCHAR(20) NOT NULL CHECK (type IN ('new_address', 'error_report', 'delivery_failed')),
    postcode VARCHAR(6),
    description TEXT NOT NULL,
    suggested_school_code VARCHAR(2),
    suggested_area_code VARCHAR(1),
    suggested_building_code VARCHAR(1),
    suggested_room_code VARCHAR(2),
    suggested_name VARCHAR(200),
    submitted_by VARCHAR(100) NOT NULL,
    submitter_type VARCHAR(20) DEFAULT 'user' CHECK (submitter_type IN ('user', 'courier')),
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
    reviewed_by VARCHAR(100),
    review_notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 7. Postcode使用统计表
CREATE TABLE IF NOT EXISTS postcode_stats (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    postcode VARCHAR(6) NOT NULL,
    delivery_count INTEGER DEFAULT 0,
    error_count INTEGER DEFAULT 0,
    last_used TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    popularity_score DECIMAL(5,2) DEFAULT 0.0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(postcode)
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_postcode_rooms_full_postcode ON postcode_rooms(full_postcode);
CREATE INDEX IF NOT EXISTS idx_postcode_rooms_managed_by ON postcode_rooms(managed_by);
CREATE INDEX IF NOT EXISTS idx_postcode_buildings_managed_by ON postcode_buildings(managed_by);
CREATE INDEX IF NOT EXISTS idx_postcode_areas_managed_by ON postcode_areas(managed_by);
CREATE INDEX IF NOT EXISTS idx_postcode_schools_managed_by ON postcode_schools(managed_by);
CREATE INDEX IF NOT EXISTS idx_postcode_courier_permissions_courier_id ON postcode_courier_permissions(courier_id);
CREATE INDEX IF NOT EXISTS idx_postcode_feedbacks_status ON postcode_feedbacks(status);
CREATE INDEX IF NOT EXISTS idx_postcode_feedbacks_submitted_by ON postcode_feedbacks(submitted_by);
CREATE INDEX IF NOT EXISTS idx_postcode_stats_postcode ON postcode_stats(postcode);
CREATE INDEX IF NOT EXISTS idx_postcode_stats_popularity_score ON postcode_stats(popularity_score DESC);

-- 创建全文搜索索引
CREATE INDEX IF NOT EXISTS idx_postcode_rooms_search ON postcode_rooms USING gin(to_tsvector('simple', name || ' ' || full_postcode));
CREATE INDEX IF NOT EXISTS idx_postcode_buildings_search ON postcode_buildings USING gin(to_tsvector('simple', name));
CREATE INDEX IF NOT EXISTS idx_postcode_schools_search ON postcode_schools USING gin(to_tsvector('simple', name || ' ' || full_name));

-- 插入初始测试数据
INSERT INTO postcode_schools (code, name, full_name, managed_by) VALUES
('PK', '北京大学', '北京大学', 'courier_level4_city'),
('QH', '清华大学', '清华大学', 'courier_level4_city')
ON CONFLICT (code) DO NOTHING;

INSERT INTO postcode_areas (school_code, code, name, managed_by) VALUES
('PK', '5', '第五片区', 'courier_level3_school'),
('PK', '3', '第三片区', 'courier_level3_school'),
('QH', '1', '第一片区', 'courier_level3_school')
ON CONFLICT (school_code, code) DO NOTHING;

INSERT INTO postcode_buildings (school_code, area_code, code, name, type, floors, managed_by) VALUES
('PK', '5', 'F', 'F栋宿舍', 'dormitory', 6, 'courier_level2_zone'),
('PK', '3', 'A', 'A栋教学楼', 'teaching', 5, 'courier_level2_zone'),
('QH', '1', 'C', 'C栋宿舍', 'dormitory', 8, 'courier_level2_zone')
ON CONFLICT (school_code, area_code, code) DO NOTHING;

INSERT INTO postcode_rooms (school_code, area_code, building_code, code, name, type, capacity, floor, managed_by) VALUES
('PK', '5', 'F', '3D', '3D宿舍', 'dormitory', 4, 3, 'courier_level1_basic'),
('PK', '5', 'F', '2A', '2A宿舍', 'dormitory', 4, 2, 'courier_level1_basic'),
('PK', '3', 'A', '1B', '1B教室', 'classroom', 50, 1, 'courier_level1_basic'),
('QH', '1', 'C', '2E', '2E宿舍', 'dormitory', 4, 2, 'courier_level1_basic')
ON CONFLICT (school_code, area_code, building_code, code) DO NOTHING;

-- 插入权限示例数据
INSERT INTO postcode_courier_permissions (courier_id, level, prefix_patterns, can_manage, can_create, can_review) VALUES
('courier_level4_city', 4, ARRAY['PK', 'QH'], true, true, true),
('courier_level3_school', 3, ARRAY['PK5', 'PK3'], true, true, true),
('courier_level2_zone', 2, ARRAY['PK5F'], true, true, false),
('courier_level1_basic', 1, ARRAY['PK5F3D'], false, false, false)
ON CONFLICT (courier_id) DO NOTHING;

-- 插入统计示例数据
INSERT INTO postcode_stats (postcode, delivery_count, error_count, popularity_score) VALUES
('PK5F3D', 25, 1, 95.5),
('PK5F2A', 18, 0, 88.2),
('PK3A1B', 12, 2, 75.0),
('QH1C2E', 8, 0, 82.1)
ON CONFLICT (postcode) DO NOTHING;

-- 创建更新时间触发器函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 为所有表创建更新时间触发器
DO $$
DECLARE
    table_name TEXT;
BEGIN
    FOR table_name IN 
        SELECT tablename FROM pg_tables 
        WHERE schemaname = 'public' 
        AND tablename LIKE 'postcode_%'
    LOOP
        EXECUTE format('
            CREATE TRIGGER update_%I_updated_at 
            BEFORE UPDATE ON %I 
            FOR EACH ROW EXECUTE函数 update_updated_at_column();
        ', table_name, table_name);
    END LOOP;
END;
$$;

COMMENT ON TABLE postcode_schools IS '学校站点表 - 2位编码管理';
COMMENT ON TABLE postcode_areas IS '片区表 - 1位编码管理';  
COMMENT ON TABLE postcode_buildings IS '楼栋表 - 1位编码管理';
COMMENT ON TABLE postcode_rooms IS '房间表 - 2位编码管理，6位完整编码';
COMMENT ON TABLE postcode_courier_permissions IS '信使Postcode权限表';
COMMENT ON TABLE postcode_feedbacks IS '地址反馈表';
COMMENT ON TABLE postcode_stats IS 'Postcode使用统计表';