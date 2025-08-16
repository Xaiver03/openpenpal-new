-- Migration: Phase 4.2 - Credit Transfer System
-- Description: Add credit transfer functionality to enable users to transfer credits to each other
-- Author: Claude Code Assistant
-- Date: 2025-08-15

-- =====================================================
-- Phase 4.2: 积分转赠系统数据库迁移
-- =====================================================

-- 1. 创建积分转赠记录表
CREATE TABLE IF NOT EXISTS credit_transfers (
    id VARCHAR(36) PRIMARY KEY,
    from_user_id VARCHAR(36) NOT NULL COMMENT '转出用户ID',
    to_user_id VARCHAR(36) NOT NULL COMMENT '转入用户ID', 
    amount INTEGER NOT NULL COMMENT '转赠积分数量',
    transfer_type ENUM('direct', 'gift', 'reward') NOT NULL COMMENT '转赠类型',
    status ENUM('pending', 'processed', 'canceled', 'expired', 'rejected') DEFAULT 'pending' COMMENT '转赠状态',
    message TEXT COMMENT '转赠留言',
    processed_at TIMESTAMP NULL COMMENT '处理时间',
    expires_at TIMESTAMP NOT NULL COMMENT '过期时间',
    fee INTEGER DEFAULT 0 COMMENT '转赠手续费',
    reference VARCHAR(100) COMMENT '关联引用',
    metadata JSON COMMENT '额外元数据',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    -- 创建索引
    INDEX idx_transfers_from_user (from_user_id),
    INDEX idx_transfers_to_user (to_user_id),
    INDEX idx_transfers_status (status),
    INDEX idx_transfers_processed_at (processed_at),
    INDEX idx_transfers_expires_at (expires_at),
    INDEX idx_transfers_created_at (created_at),
    INDEX idx_transfers_transfer_type (transfer_type),
    
    -- 外键约束（如果users表存在）
    FOREIGN KEY (from_user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (to_user_id) REFERENCES users(id) ON DELETE CASCADE,
    
    -- 检查约束
    CONSTRAINT chk_transfer_amount_positive CHECK (amount > 0),
    CONSTRAINT chk_transfer_fee_non_negative CHECK (fee >= 0),
    CONSTRAINT chk_transfer_different_users CHECK (from_user_id != to_user_id),
    CONSTRAINT chk_transfer_expires_after_creation CHECK (expires_at > created_at)
);

-- 2. 创建积分转赠规则表
CREATE TABLE IF NOT EXISTS credit_transfer_rules (
    id VARCHAR(36) PRIMARY KEY,
    rule_name VARCHAR(100) NOT NULL COMMENT '规则名称',
    min_amount INTEGER DEFAULT 1 COMMENT '最小转赠数量',
    max_amount INTEGER DEFAULT 1000 COMMENT '最大转赠数量',
    daily_limit INTEGER DEFAULT 500 COMMENT '每日转赠限制',
    monthly_limit INTEGER DEFAULT 5000 COMMENT '每月转赠限制',
    fee_rate DECIMAL(5,4) DEFAULT 0 COMMENT '手续费率 (0-1)',
    min_fee INTEGER DEFAULT 0 COMMENT '最小手续费',
    max_fee INTEGER DEFAULT 100 COMMENT '最大手续费',
    expiration_hours INTEGER DEFAULT 72 COMMENT '转赠过期小时数',
    require_confirmation BOOLEAN DEFAULT TRUE COMMENT '是否需要确认',
    allow_self_transfer BOOLEAN DEFAULT FALSE COMMENT '是否允许自转',
    restricted_user_levels JSON COMMENT '受限用户等级',
    allowed_transfer_types JSON COMMENT '允许的转赠类型',
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    priority INTEGER DEFAULT 0 COMMENT '优先级',
    applicable_user_roles JSON COMMENT '适用用户角色',
    description TEXT COMMENT '规则描述',
    created_by VARCHAR(36) COMMENT '创建人',
    updated_by VARCHAR(36) COMMENT '更新人',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    -- 创建索引
    INDEX idx_transfer_rules_is_active (is_active),
    INDEX idx_transfer_rules_priority (priority),
    INDEX idx_transfer_rules_created_at (created_at),
    
    -- 检查约束
    CONSTRAINT chk_rule_min_max_amount CHECK (min_amount <= max_amount),
    CONSTRAINT chk_rule_daily_monthly_limit CHECK (daily_limit <= monthly_limit),
    CONSTRAINT chk_rule_fee_rate CHECK (fee_rate >= 0 AND fee_rate <= 1),
    CONSTRAINT chk_rule_min_max_fee CHECK (min_fee <= max_fee),
    CONSTRAINT chk_rule_expiration_hours CHECK (expiration_hours > 0)
);

-- 3. 创建积分转赠限制记录表
CREATE TABLE IF NOT EXISTS credit_transfer_limits (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL COMMENT '用户ID',
    date DATE NOT NULL COMMENT '日期',
    daily_used INTEGER DEFAULT 0 COMMENT '当日已使用',
    monthly_used INTEGER DEFAULT 0 COMMENT '当月已使用',
    daily_count INTEGER DEFAULT 0 COMMENT '当日转赠次数',
    monthly_count INTEGER DEFAULT 0 COMMENT '当月转赠次数',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    -- 创建索引
    INDEX idx_transfer_limits_user_id (user_id),
    INDEX idx_transfer_limits_date (date),
    INDEX idx_transfer_limits_user_date (user_id, date),
    
    -- 外键约束
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    
    -- 唯一性约束
    UNIQUE KEY uk_transfer_limits_user_date (user_id, date),
    
    -- 检查约束
    CONSTRAINT chk_limit_daily_used_non_negative CHECK (daily_used >= 0),
    CONSTRAINT chk_limit_monthly_used_non_negative CHECK (monthly_used >= 0),
    CONSTRAINT chk_limit_daily_count_non_negative CHECK (daily_count >= 0),
    CONSTRAINT chk_limit_monthly_count_non_negative CHECK (monthly_count >= 0)
);

-- 4. 创建积分转赠通知表
CREATE TABLE IF NOT EXISTS credit_transfer_notifications (
    id VARCHAR(36) PRIMARY KEY,
    transfer_id VARCHAR(36) NOT NULL COMMENT '转赠ID',
    user_id VARCHAR(36) NOT NULL COMMENT '接收用户ID',
    notification_type ENUM('transfer_sent', 'transfer_received', 'transfer_expired', 'transfer_canceled') NOT NULL COMMENT '通知类型',
    title VARCHAR(200) NOT NULL COMMENT '通知标题',
    content TEXT NOT NULL COMMENT '通知内容',
    is_read BOOLEAN DEFAULT FALSE COMMENT '是否已读',
    read_at TIMESTAMP NULL COMMENT '阅读时间',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    -- 创建索引
    INDEX idx_transfer_notifications_transfer_id (transfer_id),
    INDEX idx_transfer_notifications_user_id (user_id),
    INDEX idx_transfer_notifications_type (notification_type),
    INDEX idx_transfer_notifications_is_read (is_read),
    INDEX idx_transfer_notifications_created_at (created_at),
    
    -- 外键约束
    FOREIGN KEY (transfer_id) REFERENCES credit_transfers(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- 5. 插入默认转赠规则
INSERT INTO credit_transfer_rules (
    id, rule_name, min_amount, max_amount, daily_limit, monthly_limit, 
    fee_rate, min_fee, max_fee, expiration_hours, require_confirmation, 
    allow_self_transfer, restricted_user_levels, allowed_transfer_types, 
    is_active, priority, applicable_user_roles, description, created_by, updated_by
) VALUES 
-- 普通用户转赠规则
('transfer-rule-001', '普通用户转赠规则', 1, 500, 200, 2000, 0.02, 1, 20, 72, TRUE, FALSE, '[]', '["direct", "gift"]', TRUE, 100, '["student", "user"]', '普通用户的积分转赠规则，2%手续费', 'system', 'system'),

-- VIP用户转赠规则
('transfer-rule-002', 'VIP用户转赠规则', 1, 1000, 500, 5000, 0.01, 0, 10, 72, TRUE, FALSE, '[]', '["direct", "gift", "reward"]', TRUE, 200, '["vip", "premium"]', 'VIP用户的积分转赠规则，1%手续费', 'system', 'system'),

-- 信使转赠规则
('transfer-rule-003', '信使转赠规则', 1, 2000, 1000, 10000, 0.005, 0, 5, 72, TRUE, FALSE, '[]', '["direct", "gift", "reward"]', TRUE, 300, '["courier_level_1", "courier_level_2", "courier_level_3", "courier_level_4"]', '信使的积分转赠规则，0.5%手续费', 'system', 'system'),

-- 管理员转赠规则
('transfer-rule-004', '管理员转赠规则', 1, 10000, 5000, 50000, 0, 0, 0, 168, FALSE, FALSE, '[]', '["direct", "gift", "reward"]', TRUE, 400, '["admin", "platform_admin", "super_admin"]', '管理员的积分转赠规则，无手续费', 'system', 'system'),

-- 默认规则（最低优先级）
('transfer-rule-default', '默认转赠规则', 1, 100, 50, 500, 0.05, 1, 10, 24, TRUE, FALSE, '[]', '["direct"]', TRUE, 1, '[]', '系统默认的积分转赠规则，5%手续费', 'system', 'system')

ON DUPLICATE KEY UPDATE
    rule_name = VALUES(rule_name),
    min_amount = VALUES(min_amount),
    max_amount = VALUES(max_amount),
    daily_limit = VALUES(daily_limit),
    monthly_limit = VALUES(monthly_limit),
    fee_rate = VALUES(fee_rate),
    min_fee = VALUES(min_fee),
    max_fee = VALUES(max_fee),
    description = VALUES(description),
    updated_at = CURRENT_TIMESTAMP;

-- 6. 创建视图：用户转赠统计摘要
CREATE OR REPLACE VIEW v_user_transfer_summary AS
SELECT 
    u.id as user_id,
    u.username,
    -- 发送统计
    COALESCE(sent.sent_count, 0) as sent_count,
    COALESCE(sent.sent_amount, 0) as sent_amount,
    COALESCE(sent.sent_fees, 0) as sent_fees,
    -- 接收统计
    COALESCE(received.received_count, 0) as received_count,
    COALESCE(received.received_amount, 0) as received_amount,
    -- 最后转赠时间
    GREATEST(COALESCE(sent.last_sent, '1970-01-01'), COALESCE(received.last_received, '1970-01-01')) as last_transfer_time
FROM users u
LEFT JOIN (
    SELECT 
        from_user_id,
        COUNT(*) as sent_count,
        SUM(amount) as sent_amount,
        SUM(fee) as sent_fees,
        MAX(created_at) as last_sent
    FROM credit_transfers 
    WHERE status = 'processed'
    GROUP BY from_user_id
) sent ON u.id = sent.from_user_id
LEFT JOIN (
    SELECT 
        to_user_id,
        COUNT(*) as received_count,
        SUM(amount) as received_amount,
        MAX(created_at) as last_received
    FROM credit_transfers 
    WHERE status = 'processed'
    GROUP BY to_user_id
) received ON u.id = received.to_user_id;

-- 7. 创建视图：转赠状态统计
CREATE OR REPLACE VIEW v_transfer_status_stats AS
SELECT 
    DATE(created_at) as transfer_date,
    transfer_type,
    status,
    COUNT(*) as transfer_count,
    SUM(amount) as total_amount,
    SUM(fee) as total_fees,
    AVG(amount) as avg_amount
FROM credit_transfers
GROUP BY DATE(created_at), transfer_type, status
ORDER BY transfer_date DESC, transfer_type, status;

-- 8. 创建视图：每日转赠趋势
CREATE OR REPLACE VIEW v_daily_transfer_trend AS
SELECT 
    DATE(created_at) as transfer_date,
    COUNT(*) as total_transfers,
    COUNT(DISTINCT from_user_id) as unique_senders,
    COUNT(DISTINCT to_user_id) as unique_recipients,
    SUM(amount) as total_amount,
    SUM(fee) as total_fees,
    AVG(amount) as avg_amount,
    SUM(CASE WHEN status = 'processed' THEN 1 ELSE 0 END) as processed_count,
    SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END) as pending_count,
    SUM(CASE WHEN status = 'canceled' THEN 1 ELSE 0 END) as canceled_count,
    SUM(CASE WHEN status = 'expired' THEN 1 ELSE 0 END) as expired_count,
    SUM(CASE WHEN status = 'rejected' THEN 1 ELSE 0 END) as rejected_count
FROM credit_transfers
GROUP BY DATE(created_at)
ORDER BY transfer_date DESC;

-- 9. 添加触发器：自动清理过期转赠通知
DELIMITER $$

CREATE TRIGGER tr_cleanup_old_transfer_notifications
AFTER UPDATE ON credit_transfers
FOR EACH ROW
BEGIN
    -- 当转赠状态变为非pending时，清理相关通知
    IF OLD.status = 'pending' AND NEW.status != 'pending' THEN
        DELETE FROM credit_transfer_notifications 
        WHERE transfer_id = NEW.id 
        AND notification_type IN ('transfer_sent', 'transfer_received')
        AND created_at < DATE_SUB(NOW(), INTERVAL 30 DAY);
    END IF;
END$$

DELIMITER ;

-- 10. 添加表注释
ALTER TABLE credit_transfers COMMENT = '积分转赠记录表 - Phase 4.2';
ALTER TABLE credit_transfer_rules COMMENT = '积分转赠规则配置表';
ALTER TABLE credit_transfer_limits COMMENT = '用户转赠限制记录表';
ALTER TABLE credit_transfer_notifications COMMENT = '转赠通知记录表';

-- 11. 创建定时任务（需要MySQL事件调度器）
-- 注意：这需要启用事件调度器 SET GLOBAL event_scheduler = ON;

-- 每小时处理过期转赠
/*
CREATE EVENT IF NOT EXISTS ev_process_expired_transfers
ON SCHEDULE EVERY 1 HOUR
STARTS CURRENT_TIMESTAMP
DO
BEGIN
    -- 标记过期的转赠
    UPDATE credit_transfers 
    SET status = 'expired', 
        processed_at = NOW(),
        updated_at = NOW()
    WHERE status = 'pending' 
    AND expires_at < NOW();
    
    -- 记录处理日志
    INSERT INTO system_logs (level, message, created_at) 
    VALUES ('INFO', CONCAT('Processed expired credit transfers: ', ROW_COUNT(), ' records'), NOW());
END;
*/

-- 每天清理旧的通知记录（保留30天）
/*
CREATE EVENT IF NOT EXISTS ev_cleanup_old_transfer_notifications
ON SCHEDULE EVERY 1 DAY
STARTS CURRENT_TIMESTAMP + INTERVAL 1 HOUR
DO
BEGIN
    DELETE FROM credit_transfer_notifications 
    WHERE created_at < DATE_SUB(NOW(), INTERVAL 30 DAY);
    
    INSERT INTO system_logs (level, message, created_at) 
    VALUES ('INFO', CONCAT('Cleaned up old transfer notifications: ', ROW_COUNT(), ' records'), NOW());
END;
*/

-- 12. 权限设置（如果需要）
-- GRANT SELECT, INSERT, UPDATE, DELETE ON credit_transfers TO 'app_user'@'%';
-- GRANT SELECT, INSERT, UPDATE ON credit_transfer_rules TO 'app_user'@'%';
-- GRANT SELECT, INSERT, UPDATE, DELETE ON credit_transfer_limits TO 'app_user'@'%';
-- GRANT SELECT, INSERT, UPDATE, DELETE ON credit_transfer_notifications TO 'app_user'@'%';

-- Migration completed: Phase 4.2 Credit Transfer System
-- This migration adds comprehensive credit transfer functionality:
-- 1. Credit transfer records with full status lifecycle
-- 2. Flexible rule system for different user types
-- 3. Daily/monthly transfer limits and tracking
-- 4. Notification system for transfer events
-- 5. Pre-configured rules for different user roles
-- 6. Statistical views for monitoring and reporting
-- 7. Automated cleanup and maintenance triggers
-- 8. Performance optimized with proper indexes and constraints
-- 9. Security measures with validation and foreign key constraints