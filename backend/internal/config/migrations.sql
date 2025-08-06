-- OpenPenPal 数据库优化 SQL
-- 基于 SOTA 原则的性能优化

-- 1. 优化索引策略
-- 信件表性能优化
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_letters_status_created 
    ON letters(status, created_at DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_letters_code_hash 
    ON letters USING hash(code);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_letters_user_status 
    ON letters(user_id, status);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_letters_school_status 
    ON letters(school_code, status, created_at DESC);

-- 用户表性能优化
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_role_school 
    ON users(role, school_code);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_email_hash 
    ON users USING hash(email);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_username_hash 
    ON users USING hash(username);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_active_role 
    ON users(is_active, role) WHERE is_active = true;

-- 信使任务表性能优化
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_courier_tasks_status_priority 
    ON courier_tasks(status, priority, created_at DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_courier_tasks_courier_status 
    ON courier_tasks(courier_id, status);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_courier_tasks_location 
    ON courier_tasks USING gist(target_location);

-- 状态日志表优化
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_status_logs_letter_created 
    ON status_logs(letter_id, created_at DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_status_logs_status_created 
    ON status_logs(status, created_at DESC);

-- 2. 分区策略（按时间分区）
-- 为大表实现分区以提高查询性能

-- 信件表按月分区
-- 注意：这需要在表创建时设置，这里提供参考结构
/*
-- 创建分区主表（如果重新设计）
CREATE TABLE letters_partitioned (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    title VARCHAR(200) NOT NULL,
    content TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    style VARCHAR(20) NOT NULL DEFAULT 'classic',
    code VARCHAR(20) UNIQUE,
    school_code VARCHAR(20) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
) PARTITION BY RANGE (created_at);

-- 创建分区
CREATE TABLE letters_2024_q1 PARTITION OF letters_partitioned 
    FOR VALUES FROM ('2024-01-01') TO ('2024-04-01');
CREATE TABLE letters_2024_q2 PARTITION OF letters_partitioned 
    FOR VALUES FROM ('2024-04-01') TO ('2024-07-01');
CREATE TABLE letters_2024_q3 PARTITION OF letters_partitioned 
    FOR VALUES FROM ('2024-07-01') TO ('2024-10-01');
CREATE TABLE letters_2024_q4 PARTITION OF letters_partitioned 
    FOR VALUES FROM ('2024-10-01') TO ('2025-01-01');
*/

-- 3. 物化视图（预计算常用查询）
-- 用户统计物化视图
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_user_stats AS
SELECT 
    u.id,
    u.username,
    u.nickname,
    u.role,
    u.school_code,
    COUNT(l.id) FILTER (WHERE l.status = 'draft') as draft_count,
    COUNT(l.id) FILTER (WHERE l.status = 'generated') as generated_count,
    COUNT(l.id) FILTER (WHERE l.status = 'delivered') as delivered_count,
    COUNT(l.id) as total_letters,
    MAX(l.created_at) as last_letter_date,
    u.created_at as user_created_at
FROM users u
LEFT JOIN letters l ON u.id = l.user_id AND l.deleted_at IS NULL
WHERE u.deleted_at IS NULL
GROUP BY u.id, u.username, u.nickname, u.role, u.school_code, u.created_at;

-- 为物化视图创建索引
CREATE UNIQUE INDEX IF NOT EXISTS idx_mv_user_stats_id ON mv_user_stats(id);
CREATE INDEX IF NOT EXISTS idx_mv_user_stats_school_role ON mv_user_stats(school_code, role);

-- 信使任务统计物化视图
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_courier_stats AS
SELECT 
    c.id,
    c.user_id,
    u.nickname,
    u.school_code,
    COUNT(ct.id) FILTER (WHERE ct.status = 'pending') as pending_tasks,
    COUNT(ct.id) FILTER (WHERE ct.status = 'in_progress') as active_tasks,
    COUNT(ct.id) FILTER (WHERE ct.status = 'completed') as completed_tasks,
    COUNT(ct.id) as total_tasks,
    AVG(ct.completion_time) FILTER (WHERE ct.status = 'completed') as avg_completion_time,
    SUM(ct.reward_points) FILTER (WHERE ct.status = 'completed') as total_points,
    MAX(ct.updated_at) as last_activity
FROM couriers c
INNER JOIN users u ON c.user_id = u.id
LEFT JOIN courier_tasks ct ON c.id = ct.courier_id
WHERE c.deleted_at IS NULL AND u.deleted_at IS NULL
GROUP BY c.id, c.user_id, u.nickname, u.school_code;

-- 为物化视图创建索引
CREATE UNIQUE INDEX IF NOT EXISTS idx_mv_courier_stats_id ON mv_courier_stats(id);
CREATE INDEX IF NOT EXISTS idx_mv_courier_stats_school ON mv_courier_stats(school_code);
CREATE INDEX IF NOT EXISTS idx_mv_courier_stats_performance ON mv_courier_stats(total_points DESC, avg_completion_time ASC);

-- 4. 全文搜索优化
-- 为信件内容添加全文搜索索引
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_letters_fulltext_search 
    ON letters USING gin(to_tsvector('simple', title || ' ' || content));

-- 为用户信息添加搜索索引
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_fulltext_search 
    ON users USING gin(to_tsvector('simple', username || ' ' || nickname));

-- 5. 约束和触发器优化
-- 自动更新 updated_at 字段的触发器
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 为主要表添加更新时间触发器
DROP TRIGGER IF EXISTS update_letters_updated_at ON letters;
CREATE TRIGGER update_letters_updated_at 
    BEFORE UPDATE ON letters 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_users_updated_at ON users;
CREATE TRIGGER update_users_updated_at 
    BEFORE UPDATE ON users 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_couriers_updated_at ON couriers;
CREATE TRIGGER update_couriers_updated_at 
    BEFORE UPDATE ON couriers 
    FOR EACH ROW EXECUTE FUNCTION update_couriers_updated_at_column();

-- 6. 数据完整性约束
-- 添加业务逻辑约束
ALTER TABLE letters 
ADD CONSTRAINT chk_letters_status 
CHECK (status IN ('draft', 'generated', 'collected', 'in_transit', 'delivered', 'failed'));

ALTER TABLE users 
ADD CONSTRAINT chk_users_role 
CHECK (role IN ('user', 'courier', 'senior_courier', 'courier_coordinator', 'school_admin', 'platform_admin', 'super_admin'));

-- 确保学校代码格式正确
ALTER TABLE users 
ADD CONSTRAINT chk_users_school_code 
CHECK (school_code ~ '^[A-Z0-9]{4,10}$');

-- 7. 性能优化设置
-- 调整 PostgreSQL 配置建议（需要在 postgresql.conf 中设置）
/*
-- 连接和内存设置
max_connections = 200
shared_buffers = 256MB
effective_cache_size = 1GB
work_mem = 4MB
maintenance_work_mem = 64MB

-- 检查点和 WAL 设置
checkpoint_completion_target = 0.9
wal_buffers = 16MB
default_statistics_target = 100

-- 查询规划器设置
random_page_cost = 1.1
effective_io_concurrency = 200
*/

-- 8. 定期维护任务（建议通过 cron 执行）
-- 刷新物化视图的函数
CREATE OR REPLACE FUNCTION refresh_materialized_views()
RETURNS void AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY mv_user_stats;
    REFRESH MATERIALIZED VIEW CONCURRENTLY mv_courier_stats;
END;
$$ LANGUAGE plpgsql;

-- 清理软删除数据的函数（30天后永久删除）
CREATE OR REPLACE FUNCTION cleanup_soft_deleted_data()
RETURNS void AS $$
BEGIN
    DELETE FROM letters WHERE deleted_at < CURRENT_TIMESTAMP - INTERVAL '30 days';
    DELETE FROM users WHERE deleted_at < CURRENT_TIMESTAMP - INTERVAL '30 days';
    DELETE FROM couriers WHERE deleted_at < CURRENT_TIMESTAMP - INTERVAL '30 days';
END;
$$ LANGUAGE plpgsql;

-- 9. 查询性能分析辅助函数
-- 创建慢查询日志分析视图
CREATE OR REPLACE VIEW slow_queries AS
SELECT 
    query,
    calls,
    total_time,
    total_time/calls as avg_time,
    rows,
    100.0 * shared_blks_hit / nullif(shared_blks_hit + shared_blks_read, 0) AS hit_percent
FROM pg_stat_statements 
ORDER BY total_time DESC;

-- 创建表大小分析视图
CREATE OR REPLACE VIEW table_sizes AS
SELECT 
    schemaname,
    tablename,
    attname,
    n_distinct,
    correlation,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size
FROM pg_stats 
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;

-- 10. 数据库监控查询
-- 锁等待监控
CREATE OR REPLACE VIEW lock_monitoring AS
SELECT 
    blocked_locks.pid AS blocked_pid,
    blocked_activity.usename AS blocked_user,
    blocking_locks.pid AS blocking_pid,
    blocking_activity.usename AS blocking_user,
    blocked_activity.query AS blocked_statement,
    blocking_activity.query AS current_statement_in_blocking_process
FROM pg_catalog.pg_locks blocked_locks
JOIN pg_catalog.pg_stat_activity blocked_activity ON blocked_activity.pid = blocked_locks.pid
JOIN pg_catalog.pg_locks blocking_locks ON blocking_locks.locktype = blocked_locks.locktype
JOIN pg_catalog.pg_stat_activity blocking_activity ON blocking_activity.pid = blocking_locks.pid
WHERE NOT blocked_locks.granted;

-- 连接数监控
CREATE OR REPLACE VIEW connection_monitoring AS
SELECT 
    count(*) as total_connections,
    count(*) FILTER (WHERE state = 'active') as active_connections,
    count(*) FILTER (WHERE state = 'idle') as idle_connections,
    count(*) FILTER (WHERE state = 'idle in transaction') as idle_in_transaction
FROM pg_stat_activity;

COMMENT ON MATERIALIZED VIEW mv_user_stats IS 'OpenPenPal 用户统计物化视图 - 每小时刷新';
COMMENT ON MATERIALIZED VIEW mv_courier_stats IS 'OpenPenPal 信使统计物化视图 - 每小时刷新';
COMMENT ON FUNCTION refresh_materialized_views() IS '刷新所有物化视图 - 建议每小时执行';
COMMENT ON FUNCTION cleanup_soft_deleted_data() IS '清理软删除数据 - 建议每日执行';

-- 11. 系统配置表
-- 用于存储动态系统配置项
CREATE TABLE IF NOT EXISTS system_settings (
    id VARCHAR(36) PRIMARY KEY,
    key VARCHAR(100) UNIQUE NOT NULL,
    value TEXT,
    category VARCHAR(50),
    data_type VARCHAR(20) DEFAULT 'string',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 为系统配置表创建索引
CREATE INDEX IF NOT EXISTS idx_system_settings_key ON system_settings USING hash(key);
CREATE INDEX IF NOT EXISTS idx_system_settings_category ON system_settings(category);

-- 创建更新触发器以自动更新 updated_at
CREATE OR REPLACE FUNCTION update_system_settings_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_system_settings_updated_at
    BEFORE UPDATE ON system_settings
    FOR EACH ROW
    EXECUTE FUNCTION update_system_settings_updated_at();

COMMENT ON TABLE system_settings IS 'OpenPenPal 系统配置表 - 存储动态配置项';
COMMENT ON COLUMN system_settings.key IS '配置键，唯一标识';
COMMENT ON COLUMN system_settings.value IS '配置值，存储为文本';
COMMENT ON COLUMN system_settings.category IS '配置分类：general, email, letter, user, courier, security, notification';
COMMENT ON COLUMN system_settings.data_type IS '数据类型：string, number, boolean, json';