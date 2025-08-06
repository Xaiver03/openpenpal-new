-- OpenPenPal 数据库初始化脚本
-- 创建所有微服务需要的数据库表和初始数据

-- =====================================================
-- 主后端服务表结构
-- =====================================================

-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(36) PRIMARY KEY DEFAULT gen_random_uuid()::text,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(100),
    school_code VARCHAR(20),
    avatar_url TEXT,
    bio TEXT,
    role VARCHAR(20) DEFAULT 'user',
    is_courier BOOLEAN DEFAULT FALSE,
    status VARCHAR(20) DEFAULT 'active',
    last_login TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 学校信息表
CREATE TABLE IF NOT EXISTS schools (
    id SERIAL PRIMARY KEY,
    code VARCHAR(20) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    address TEXT,
    city VARCHAR(50),
    province VARCHAR(50),
    postal_code VARCHAR(10),
    contact_phone VARCHAR(20),
    contact_email VARCHAR(255),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- =====================================================
-- 写信服务表结构
-- =====================================================

-- 信件表
CREATE TABLE IF NOT EXISTS letters (
    id VARCHAR(36) PRIMARY KEY DEFAULT gen_random_uuid()::text,
    letter_code VARCHAR(20) UNIQUE NOT NULL,
    sender_id VARCHAR(36) NOT NULL,
    receiver_hint TEXT,
    title VARCHAR(200),
    content TEXT,
    letter_type VARCHAR(20) DEFAULT 'normal',
    status VARCHAR(20) DEFAULT 'draft',
    is_urgent BOOLEAN DEFAULT FALSE,
    is_anonymous BOOLEAN DEFAULT FALSE,
    qr_code TEXT,
    estimated_delivery_time TIMESTAMP,
    actual_delivery_time TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    FOREIGN KEY (sender_id) REFERENCES users(id) ON DELETE CASCADE
);

-- 信件历史记录表
CREATE TABLE IF NOT EXISTS letter_history (
    id SERIAL PRIMARY KEY,
    letter_id VARCHAR(36) NOT NULL,
    status VARCHAR(20) NOT NULL,
    location TEXT,
    note TEXT,
    courier_id VARCHAR(36),
    created_at TIMESTAMP DEFAULT NOW(),
    
    FOREIGN KEY (letter_id) REFERENCES letters(id) ON DELETE CASCADE,
    FOREIGN KEY (courier_id) REFERENCES users(id) ON DELETE SET NULL
);

-- 草稿表
CREATE TABLE IF NOT EXISTS drafts (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    title VARCHAR(200),
    content TEXT,
    receiver_hint TEXT,
    version INTEGER DEFAULT 1,
    is_auto_save BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- =====================================================
-- 信使服务表结构
-- =====================================================

-- 信使信息表
CREATE TABLE IF NOT EXISTS couriers (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(36) UNIQUE NOT NULL,
    application_id VARCHAR(20) UNIQUE,
    zone VARCHAR(100) NOT NULL,
    zone_code VARCHAR(20),
    phone VARCHAR(20) NOT NULL,
    id_card VARCHAR(20) NOT NULL,
    experience TEXT,
    status VARCHAR(20) DEFAULT 'pending',
    level INTEGER DEFAULT 1,
    points INTEGER DEFAULT 0,
    rating DECIMAL(3,2) DEFAULT 5.00,
    total_tasks INTEGER DEFAULT 0,
    completed_tasks INTEGER DEFAULT 0,
    failed_tasks INTEGER DEFAULT 0,
    current_tasks INTEGER DEFAULT 0,
    parent_id INTEGER,
    approved_by VARCHAR(36),
    approved_at TIMESTAMP,
    last_active TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (approved_by) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (parent_id) REFERENCES couriers(id) ON DELETE SET NULL
);

-- 信使等级表
CREATE TABLE IF NOT EXISTS courier_levels (
    id SERIAL PRIMARY KEY,
    level INTEGER UNIQUE NOT NULL,
    name VARCHAR(50) NOT NULL,
    description TEXT,
    min_points INTEGER NOT NULL,
    max_tasks_concurrent INTEGER DEFAULT 5,
    permissions JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);

-- 信使权限表
CREATE TABLE IF NOT EXISTS courier_permissions (
    id SERIAL PRIMARY KEY,
    courier_id INTEGER NOT NULL,
    permission VARCHAR(100) NOT NULL,
    scope VARCHAR(100),
    granted_by VARCHAR(36),
    granted_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP,
    
    FOREIGN KEY (courier_id) REFERENCES couriers(id) ON DELETE CASCADE,
    FOREIGN KEY (granted_by) REFERENCES users(id) ON DELETE SET NULL,
    UNIQUE(courier_id, permission, scope)
);

-- 任务表
CREATE TABLE IF NOT EXISTS courier_tasks (
    id SERIAL PRIMARY KEY,
    task_id VARCHAR(20) UNIQUE NOT NULL,
    letter_id VARCHAR(36) NOT NULL,
    courier_id INTEGER,
    status VARCHAR(20) DEFAULT 'available',
    priority VARCHAR(20) DEFAULT 'normal',
    pickup_location TEXT NOT NULL,
    delivery_location TEXT NOT NULL,
    estimated_distance DECIMAL(8,2),
    reward DECIMAL(8,2) DEFAULT 5.00,
    estimated_time INTEGER,
    accepted_at TIMESTAMP,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    FOREIGN KEY (letter_id) REFERENCES letters(id) ON DELETE CASCADE,
    FOREIGN KEY (courier_id) REFERENCES couriers(id) ON DELETE SET NULL
);

-- 扫码记录表
CREATE TABLE IF NOT EXISTS scan_records (
    id SERIAL PRIMARY KEY,
    task_id VARCHAR(20) NOT NULL,
    courier_id INTEGER NOT NULL,
    letter_id VARCHAR(36) NOT NULL,
    action VARCHAR(20) NOT NULL,
    location TEXT,
    latitude DECIMAL(10,8),
    longitude DECIMAL(11,8),
    note TEXT,
    photo_url TEXT,
    timestamp TIMESTAMP DEFAULT NOW(),
    
    FOREIGN KEY (courier_id) REFERENCES couriers(id) ON DELETE CASCADE,
    FOREIGN KEY (letter_id) REFERENCES letters(id) ON DELETE CASCADE
);

-- 信使成长记录表
CREATE TABLE IF NOT EXISTS courier_growth (
    id SERIAL PRIMARY KEY,
    courier_id INTEGER NOT NULL,
    action VARCHAR(50) NOT NULL,
    points INTEGER NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    
    FOREIGN KEY (courier_id) REFERENCES couriers(id) ON DELETE CASCADE
);

-- 信使徽章表
CREATE TABLE IF NOT EXISTS courier_badges (
    id SERIAL PRIMARY KEY,
    courier_id INTEGER NOT NULL,
    badge_type VARCHAR(50) NOT NULL,
    badge_name VARCHAR(100) NOT NULL,
    description TEXT,
    icon_url TEXT,
    earned_at TIMESTAMP DEFAULT NOW(),
    
    FOREIGN KEY (courier_id) REFERENCES couriers(id) ON DELETE CASCADE
);

-- =====================================================
-- 管理后台服务表结构
-- =====================================================

-- 管理员日志表
CREATE TABLE IF NOT EXISTS admin_logs (
    id SERIAL PRIMARY KEY,
    admin_id VARCHAR(36) NOT NULL,
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50),
    resource_id VARCHAR(100),
    details JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    
    FOREIGN KEY (admin_id) REFERENCES users(id) ON DELETE CASCADE
);

-- 权限表
CREATE TABLE IF NOT EXISTS permissions (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    resource VARCHAR(50) NOT NULL,
    action VARCHAR(50) NOT NULL,
    scope VARCHAR(50),
    created_at TIMESTAMP DEFAULT NOW()
);

-- 角色权限关联表
CREATE TABLE IF NOT EXISTS role_permissions (
    role VARCHAR(50) NOT NULL,
    permission_id INTEGER NOT NULL,
    granted_at TIMESTAMP DEFAULT NOW(),
    
    PRIMARY KEY (role, permission_id),
    FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
);

-- 系统配置表
CREATE TABLE IF NOT EXISTS system_config (
    id SERIAL PRIMARY KEY,
    config_key VARCHAR(100) UNIQUE NOT NULL,
    config_value TEXT NOT NULL,
    description TEXT,
    is_public BOOLEAN DEFAULT FALSE,
    is_encrypted BOOLEAN DEFAULT FALSE,
    category VARCHAR(50) DEFAULT 'general',
    validation_rules JSONB,
    updated_by VARCHAR(36),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    FOREIGN KEY (updated_by) REFERENCES users(id) ON DELETE SET NULL
);

-- =====================================================
-- OCR服务表结构
-- =====================================================

-- OCR任务表
CREATE TABLE IF NOT EXISTS ocr_tasks (
    id SERIAL PRIMARY KEY,
    task_id VARCHAR(50) UNIQUE NOT NULL,
    letter_id VARCHAR(36),
    image_path TEXT NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    ocr_engine VARCHAR(20) DEFAULT 'paddle',
    language VARCHAR(10) DEFAULT 'zh',
    is_handwriting BOOLEAN DEFAULT FALSE,
    confidence DECIMAL(4,3),
    recognized_text TEXT,
    processing_time DECIMAL(8,3),
    error_message TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    completed_at TIMESTAMP,
    
    FOREIGN KEY (letter_id) REFERENCES letters(id) ON DELETE SET NULL
);

-- =====================================================
-- 博物馆和内容管理表结构
-- =====================================================

-- 展览表
CREATE TABLE IF NOT EXISTS exhibitions (
    id SERIAL PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    theme VARCHAR(100),
    is_active BOOLEAN DEFAULT TRUE,
    letter_count INTEGER DEFAULT 0,
    created_by VARCHAR(36) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE
);

-- 展览信件关联表
CREATE TABLE IF NOT EXISTS exhibition_letters (
    exhibition_id INTEGER NOT NULL,
    letter_id VARCHAR(36) NOT NULL,
    display_order INTEGER,
    added_at TIMESTAMP DEFAULT NOW(),
    
    PRIMARY KEY (exhibition_id, letter_id),
    FOREIGN KEY (exhibition_id) REFERENCES exhibitions(id) ON DELETE CASCADE,
    FOREIGN KEY (letter_id) REFERENCES letters(id) ON DELETE CASCADE
);

-- 内容审核表
CREATE TABLE IF NOT EXISTS content_moderation (
    id SERIAL PRIMARY KEY,
    content_id VARCHAR(100) NOT NULL,
    content_type VARCHAR(20) NOT NULL,
    content_preview TEXT,
    author_id VARCHAR(36) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    moderator_id VARCHAR(36),
    reason TEXT,
    submitted_at TIMESTAMP DEFAULT NOW(),
    reviewed_at TIMESTAMP,
    
    FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (moderator_id) REFERENCES users(id) ON DELETE SET NULL
);

-- 敏感词表
CREATE TABLE IF NOT EXISTS sensitive_words (
    id SERIAL PRIMARY KEY,
    word VARCHAR(100) NOT NULL,
    type VARCHAR(20) NOT NULL,
    severity VARCHAR(20) NOT NULL,
    action VARCHAR(20) DEFAULT 'review',
    is_active BOOLEAN DEFAULT TRUE,
    created_by VARCHAR(36) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE
);

-- =====================================================
-- 索引创建
-- =====================================================

-- 用户相关索引
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_school_code ON users(school_code);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);

-- 信件相关索引
CREATE INDEX IF NOT EXISTS idx_letters_sender_id ON letters(sender_id);
CREATE INDEX IF NOT EXISTS idx_letters_status ON letters(status);
CREATE INDEX IF NOT EXISTS idx_letters_letter_code ON letters(letter_code);
CREATE INDEX IF NOT EXISTS idx_letters_created_at ON letters(created_at DESC);

-- 信使相关索引
CREATE INDEX IF NOT EXISTS idx_couriers_user_id ON couriers(user_id);
CREATE INDEX IF NOT EXISTS idx_couriers_status ON couriers(status);
CREATE INDEX IF NOT EXISTS idx_couriers_zone_code ON couriers(zone_code);
CREATE INDEX IF NOT EXISTS idx_couriers_level ON couriers(level);

-- 任务相关索引
CREATE INDEX IF NOT EXISTS idx_courier_tasks_letter_id ON courier_tasks(letter_id);
CREATE INDEX IF NOT EXISTS idx_courier_tasks_courier_id ON courier_tasks(courier_id);
CREATE INDEX IF NOT EXISTS idx_courier_tasks_status ON courier_tasks(status);
CREATE INDEX IF NOT EXISTS idx_courier_tasks_priority ON courier_tasks(priority);

-- =====================================================
-- 初始数据插入
-- =====================================================

-- 插入默认学校数据
INSERT INTO schools (code, name, city, province) VALUES 
('PKU', '北京大学', '北京', '北京市'),
('THU', '清华大学', '北京', '北京市'),
('FDU', '复旦大学', '上海', '上海市'),
('SJTU', '上海交通大学', '上海', '上海市'),
('ZJU', '浙江大学', '杭州', '浙江省')
ON CONFLICT (code) DO NOTHING;

-- 插入信使等级数据
INSERT INTO courier_levels (level, name, description, min_points, max_tasks_concurrent) VALUES 
(1, 'LevelOne', '一级信使 - 校园配送', 0, 3),
(2, 'LevelTwo', '二级信使 - 跨校配送', 500, 5),
(3, 'LevelThree', '三级信使 - 区域管理', 2000, 8),
(4, 'LevelFour', '四级信使 - 城市协调', 5000, 12),
(5, 'LevelFive', '五级信使 - 全国管理', 10000, 20)
ON CONFLICT (level) DO NOTHING;

-- 插入权限数据
INSERT INTO permissions (name, description, resource, action) VALUES 
('user.read', '查看用户信息', 'user', 'read'),
('user.write', '编辑用户信息', 'user', 'write'),
('user.delete', '删除用户', 'user', 'delete'),
('letter.read', '查看信件', 'letter', 'read'),
('letter.write', '编辑信件', 'letter', 'write'),
('letter.manage', '管理信件状态', 'letter', 'manage'),
('courier.read', '查看信使信息', 'courier', 'read'),
('courier.write', '编辑信使信息', 'courier', 'write'),
('courier.approve', '审核信使申请', 'courier', 'approve'),
('courier.assign', '分配信使任务', 'courier', 'assign'),
('museum.read', '查看博物馆内容', 'museum', 'read'),
('museum.write', '编辑博物馆内容', 'museum', 'write'),
('museum.moderate', '审核博物馆内容', 'museum', 'moderate'),
('stats.read', '查看统计数据', 'stats', 'read'),
('config.read', '查看系统配置', 'config', 'read'),
('config.write', '编辑系统配置', 'config', 'write')
ON CONFLICT (name) DO NOTHING;

-- 插入角色权限关联
INSERT INTO role_permissions (role, permission_id) 
SELECT 'super_admin', id FROM permissions
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role, permission_id)
SELECT 'school_admin', id FROM permissions WHERE resource IN ('user', 'letter', 'courier', 'stats')
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role, permission_id)
SELECT 'courier_manager', id FROM permissions WHERE resource = 'courier'
ON CONFLICT DO NOTHING;

-- 插入系统配置
INSERT INTO system_config (config_key, config_value, description, is_public, category) VALUES 
('app.name', 'OpenPenPal', '应用名称', true, 'general'),
('app.version', '2.1.0', '应用版本', true, 'general'),
('courier.max_tasks', '5', '信使最大并发任务数', false, 'courier'),
('letter.max_size', '10485760', '信件最大文件大小(字节)', false, 'letter'),
('ocr.default_engine', 'paddle', '默认OCR引擎', false, 'ocr'),
('museum.auto_approve', 'false', '博物馆内容自动审核', false, 'museum'),
('notification.email_enabled', 'true', '邮件通知开关', false, 'notification')
ON CONFLICT (config_key) DO NOTHING;

-- =====================================================
-- 触发器和函数
-- =====================================================

-- 更新 updated_at 字段的函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 为需要的表创建更新时间触发器
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_letters_updated_at ON letters;
CREATE TRIGGER update_letters_updated_at BEFORE UPDATE ON letters 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_couriers_updated_at ON couriers;
CREATE TRIGGER update_couriers_updated_at BEFORE UPDATE ON couriers 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_system_config_updated_at ON system_config;
CREATE TRIGGER update_system_config_updated_at BEFORE UPDATE ON system_config 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- =====================================================
-- 完成消息
-- =====================================================

DO $$
BEGIN
    RAISE NOTICE '====================================';
    RAISE NOTICE 'OpenPenPal 数据库初始化完成!';
    RAISE NOTICE '====================================';
    RAISE NOTICE '已创建表: users, schools, letters, couriers, tasks, etc.';
    RAISE NOTICE '已创建索引和触发器';
    RAISE NOTICE '已插入初始数据';
    RAISE NOTICE '====================================';
END $$;