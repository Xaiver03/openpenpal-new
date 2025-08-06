-- 用户表 (扩展现有用户表的管理字段)
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(100) NOT NULL UNIQUE,
    password VARCHAR(60) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'user',
    school_code VARCHAR(20),
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    last_login TIMESTAMP WITH TIME ZONE,
    failed_login_attempts INTEGER DEFAULT 0,
    locked_until TIMESTAMP WITH TIME ZONE,
    avatar_url VARCHAR(255),
    bio VARCHAR(500),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    version BIGINT DEFAULT 0
);

-- 用户权限表
CREATE TABLE IF NOT EXISTS user_permissions (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    permission VARCHAR(100) NOT NULL,
    PRIMARY KEY (user_id, permission)
);

-- 权限表
CREATE TABLE IF NOT EXISTS permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description VARCHAR(500),
    resource VARCHAR(50) NOT NULL,
    action VARCHAR(50) NOT NULL,
    permission_type VARCHAR(20) NOT NULL DEFAULT 'RESOURCE',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    version BIGINT DEFAULT 0
);

-- 角色权限关联表
CREATE TABLE IF NOT EXISTS role_permissions (
    role VARCHAR(50) NOT NULL,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    scope VARCHAR(100),
    PRIMARY KEY (role, permission_id)
);

-- 管理员操作日志表
CREATE TABLE IF NOT EXISTS admin_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_id UUID NOT NULL REFERENCES users(id),
    action VARCHAR(100) NOT NULL,
    target_type VARCHAR(50) NOT NULL,
    target_id VARCHAR(100),
    details JSONB,
    ip_address INET,
    user_agent VARCHAR(500),
    result VARCHAR(20) NOT NULL DEFAULT 'SUCCESS',
    error_message VARCHAR(1000),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    version BIGINT DEFAULT 0
);

-- 系统配置表
CREATE TABLE IF NOT EXISTS system_config (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    config_key VARCHAR(100) NOT NULL UNIQUE,
    config_value JSONB NOT NULL,
    description VARCHAR(500),
    updated_by UUID REFERENCES users(id),
    config_type VARCHAR(20) NOT NULL DEFAULT 'SYSTEM',
    is_public BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    version BIGINT DEFAULT 0
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_school_code ON users(school_code);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);

CREATE INDEX IF NOT EXISTS idx_permissions_name ON permissions(name);
CREATE INDEX IF NOT EXISTS idx_permissions_resource_action ON permissions(resource, action);

CREATE INDEX IF NOT EXISTS idx_role_permissions_role ON role_permissions(role);
CREATE INDEX IF NOT EXISTS idx_role_permissions_permission_id ON role_permissions(permission_id);

CREATE INDEX IF NOT EXISTS idx_admin_logs_admin_id ON admin_logs(admin_id);
CREATE INDEX IF NOT EXISTS idx_admin_logs_action ON admin_logs(action);
CREATE INDEX IF NOT EXISTS idx_admin_logs_target_type ON admin_logs(target_type);
CREATE INDEX IF NOT EXISTS idx_admin_logs_created_at ON admin_logs(created_at);

-- 插入基础权限数据
INSERT INTO permissions (name, description, resource, action, permission_type) VALUES
-- 用户管理权限
('user.read', '查看用户信息', 'user', 'read', 'RESOURCE'),
('user.write', '编辑用户信息', 'user', 'write', 'RESOURCE'),
('user.create', '创建用户', 'user', 'create', 'RESOURCE'),
('user.delete', '删除用户', 'user', 'delete', 'RESOURCE'),
('user.role.manage', '管理用户角色', 'user', 'role.manage', 'RESOURCE'),

-- 信件管理权限
('letter.read', '查看信件信息', 'letter', 'read', 'RESOURCE'),
('letter.write', '编辑信件信息', 'letter', 'write', 'RESOURCE'),
('letter.delete', '删除信件', 'letter', 'delete', 'RESOURCE'),
('letter.status.manage', '管理信件状态', 'letter', 'status.manage', 'RESOURCE'),

-- 信使管理权限
('courier.read', '查看信使信息', 'courier', 'read', 'RESOURCE'),
('courier.write', '编辑信使信息', 'courier', 'write', 'RESOURCE'),
('courier.approve', '审核信使申请', 'courier', 'approve', 'RESOURCE'),
('courier.task.assign', '分配信使任务', 'courier', 'task.assign', 'RESOURCE'),

-- 统计查看权限
('stats.read', '查看统计数据', 'stats', 'read', 'RESOURCE'),
('stats.export', '导出统计数据', 'stats', 'export', 'RESOURCE'),

-- 配置管理权限
('config.read', '查看系统配置', 'config', 'read', 'RESOURCE'),
('config.write', '修改系统配置', 'config', 'write', 'RESOURCE'),

-- 系统权限
('system.admin', '系统管理员权限', 'system', 'admin', 'SYSTEM'),
('system.maintenance', '系统维护权限', 'system', 'maintenance', 'SYSTEM')
ON CONFLICT (name) DO NOTHING;

-- 插入角色权限关联
INSERT INTO role_permissions (role, permission_id) 
SELECT 'super_admin', id FROM permissions 
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role, permission_id) 
SELECT 'platform_admin', id FROM permissions 
WHERE name IN (
    'user.read', 'user.write', 'user.create', 'user.role.manage',
    'letter.read', 'letter.write', 'letter.status.manage',
    'courier.read', 'courier.write', 'courier.approve', 'courier.task.assign',
    'stats.read', 'config.read'
)
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role, permission_id) 
SELECT 'school_admin', id FROM permissions 
WHERE name IN (
    'user.read', 'user.write', 'letter.read',
    'courier.read', 'courier.write', 'stats.read'
)
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions (role, permission_id) 
SELECT 'courier_manager', id FROM permissions 
WHERE name IN (
    'courier.read', 'courier.write', 'courier.task.assign',
    'letter.read', 'stats.read'
)
ON CONFLICT DO NOTHING;

-- 插入基础系统配置
INSERT INTO system_config (config_key, config_value, description, config_type, is_public) VALUES
('max_letters_per_user_per_day', '10', '用户每日最大写信数量', 'USER', true),
('delivery_timeout_hours', '48', '投递超时时间(小时)', 'SYSTEM', true),
('auto_assign_couriers', 'true', '自动分配信使', 'SYSTEM', false),
('enable_anonymous_letters', 'true', '允许匿名信件', 'FEATURE', true),
('maintenance_mode', 'false', '维护模式', 'SYSTEM', true),
('registration_enabled', 'true', '允许用户注册', 'USER', true),
('courier_application_enabled', 'true', '允许信使申请', 'FEATURE', true)
ON CONFLICT (config_key) DO NOTHING;