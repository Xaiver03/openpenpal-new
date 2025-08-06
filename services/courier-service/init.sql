-- 信使服务数据库初始化脚本

-- 创建数据库（如果不存在）
-- CREATE DATABASE IF NOT EXISTS openpenpal;

-- 使用数据库
\c openpenpal;

-- 创建扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "postgis";

-- 创建索引提升查询性能
-- 这些索引会在 GORM 自动迁移后创建

-- 信使表索引
-- CREATE INDEX IF NOT EXISTS idx_couriers_user_id ON couriers(user_id);
-- CREATE INDEX IF NOT EXISTS idx_couriers_zone ON couriers(zone);
-- CREATE INDEX IF NOT EXISTS idx_couriers_status ON couriers(status);
-- CREATE INDEX IF NOT EXISTS idx_couriers_rating ON couriers(rating);

-- 任务表索引
-- CREATE INDEX IF NOT EXISTS idx_tasks_letter_id ON tasks(letter_id);
-- CREATE INDEX IF NOT EXISTS idx_tasks_courier_id ON tasks(courier_id);
-- CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
-- CREATE INDEX IF NOT EXISTS idx_tasks_priority ON tasks(priority);
-- CREATE INDEX IF NOT EXISTS idx_tasks_created_at ON tasks(created_at);
-- CREATE INDEX IF NOT EXISTS idx_tasks_location ON tasks(pickup_lat, pickup_lng);

-- 扫码记录表索引
-- CREATE INDEX IF NOT EXISTS idx_scan_records_task_id ON scan_records(task_id);
-- CREATE INDEX IF NOT EXISTS idx_scan_records_courier_id ON scan_records(courier_id);
-- CREATE INDEX IF NOT EXISTS idx_scan_records_letter_id ON scan_records(letter_id);
-- CREATE INDEX IF NOT EXISTS idx_scan_records_timestamp ON scan_records(timestamp);

-- 创建存储过程
-- 计算两点之间距离的函数
CREATE OR REPLACE FUNCTION calculate_distance(
    lat1 DOUBLE PRECISION,
    lng1 DOUBLE PRECISION,
    lat2 DOUBLE PRECISION,
    lng2 DOUBLE PRECISION
) RETURNS DOUBLE PRECISION AS $$
DECLARE
    R CONSTANT DOUBLE PRECISION := 6371; -- 地球半径（公里）
    dLat DOUBLE PRECISION;
    dLng DOUBLE PRECISION;
    a DOUBLE PRECISION;
    c DOUBLE PRECISION;
BEGIN
    dLat := radians(lat2 - lat1);
    dLng := radians(lng2 - lng1);
    
    a := sin(dLat/2) * sin(dLat/2) + 
         cos(radians(lat1)) * cos(radians(lat2)) * 
         sin(dLng/2) * sin(dLng/2);
    
    c := 2 * atan2(sqrt(a), sqrt(1-a));
    
    RETURN R * c;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- 创建触发器函数：任务状态变更时记录历史
CREATE OR REPLACE FUNCTION task_status_change_trigger()
RETURNS TRIGGER AS $$
BEGIN
    -- 如果状态发生变化，记录到历史表
    IF OLD.status IS DISTINCT FROM NEW.status THEN
        INSERT INTO task_status_history (
            task_id,
            old_status,
            new_status,
            changed_at,
            changed_by
        ) VALUES (
            NEW.task_id,
            OLD.status,
            NEW.status,
            NOW(),
            NEW.courier_id
        );
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 创建任务状态历史表
CREATE TABLE IF NOT EXISTS task_status_history (
    id SERIAL PRIMARY KEY,
    task_id VARCHAR(50) NOT NULL,
    old_status VARCHAR(20),
    new_status VARCHAR(20) NOT NULL,
    changed_at TIMESTAMP DEFAULT NOW(),
    changed_by VARCHAR(50)
);

-- 创建统计视图
CREATE OR REPLACE VIEW courier_stats_view AS
SELECT 
    c.id,
    c.user_id,
    c.zone,
    c.status,
    c.rating,
    COUNT(t.id) as total_tasks,
    COUNT(CASE WHEN t.status = 'delivered' THEN 1 END) as completed_tasks,
    ROUND(
        (COUNT(CASE WHEN t.status = 'delivered' THEN 1 END)::DECIMAL / 
         NULLIF(COUNT(t.id), 0) * 100), 2
    ) as success_rate,
    COALESCE(SUM(CASE WHEN t.status = 'delivered' THEN t.reward END), 0) as total_earnings,
    COUNT(CASE WHEN t.created_at >= date_trunc('month', NOW()) THEN 1 END) as this_month_tasks
FROM couriers c
LEFT JOIN tasks t ON c.user_id = t.courier_id
GROUP BY c.id, c.user_id, c.zone, c.status, c.rating;

-- 创建任务队列统计表
CREATE TABLE IF NOT EXISTS queue_stats (
    id SERIAL PRIMARY KEY,
    queue_name VARCHAR(50) NOT NULL,
    action VARCHAR(20) NOT NULL,
    count INTEGER DEFAULT 1,
    last_updated TIMESTAMP DEFAULT NOW(),
    UNIQUE(queue_name, action)
);

-- 插入示例数据（开发环境）
-- 示例信使数据
INSERT INTO couriers (user_id, zone, phone, id_card, status, rating, experience, created_at, updated_at) 
VALUES 
    ('courier1', '北京大学', '13800138001', '110101199001011234', 'approved', 4.8, '有丰富的校园投递经验', NOW(), NOW()),
    ('courier2', '清华大学', '13800138002', '110101199001011235', 'approved', 4.9, '熟悉校园路线', NOW(), NOW()),
    ('courier3', '北京大学', '13800138003', '110101199001011236', 'pending', 5.0, '新申请信使', NOW(), NOW())
ON CONFLICT (user_id) DO NOTHING;

-- 提交事务
COMMIT;