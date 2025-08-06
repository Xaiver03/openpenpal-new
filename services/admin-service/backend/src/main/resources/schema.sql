-- Admin Service PostgreSQL Schema
-- 确保基础表存在

-- 创建管理员日志表（如果不存在）
CREATE TABLE IF NOT EXISTS admin_logs (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(36),
    action VARCHAR(100) NOT NULL,
    resource VARCHAR(100),
    resource_id VARCHAR(36),
    details TEXT,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建系统配置表（如果不存在）
CREATE TABLE IF NOT EXISTS system_configs (
    id SERIAL PRIMARY KEY,
    key VARCHAR(100) UNIQUE NOT NULL,
    value TEXT,
    description TEXT,
    category VARCHAR(50),
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_by VARCHAR(36)
);

-- 创建管理会话表（如果不存在）
CREATE TABLE IF NOT EXISTS admin_sessions (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    token_hash VARCHAR(255) NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_accessed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引以提高查询性能
CREATE INDEX IF NOT EXISTS idx_admin_logs_user_id ON admin_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_admin_logs_created_at ON admin_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_admin_sessions_user_id ON admin_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_admin_sessions_expires_at ON admin_sessions(expires_at);

-- 插入默认配置（如果不存在）
INSERT INTO system_configs (key, value, description, category)
VALUES 
    ('admin.max_login_attempts', '5', '最大登录尝试次数', 'security'),
    ('admin.session_timeout', '30', '会话超时时间（分钟）', 'security'),
    ('admin.enable_2fa', 'false', '是否启用双因素认证', 'security')
ON CONFLICT (key) DO NOTHING;