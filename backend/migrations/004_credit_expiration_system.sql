-- Migration: Phase 4.1 - Credit Expiration System
-- Description: Add credit expiration functionality to the credit system
-- Author: Claude Code Assistant
-- Date: 2025-08-15

-- =====================================================
-- Phase 4.1: 积分过期系统数据库迁移
-- =====================================================

-- 1. 扩展 credit_transactions 表以支持过期功能
ALTER TABLE credit_transactions 
ADD COLUMN expires_at TIMESTAMP,
ADD COLUMN expired_at TIMESTAMP,  
ADD COLUMN is_expired BOOLEAN DEFAULT false NOT NULL;

-- 为过期相关字段添加索引
CREATE INDEX idx_credit_transactions_expires_at ON credit_transactions (expires_at) WHERE expires_at IS NOT NULL;
CREATE INDEX idx_credit_transactions_expired_at ON credit_transactions (expired_at) WHERE expired_at IS NOT NULL; 
CREATE INDEX idx_credit_transactions_is_expired ON credit_transactions (is_expired);

-- 2. 创建积分过期规则表
CREATE TABLE IF NOT EXISTS credit_expiration_rules (
    id VARCHAR(36) PRIMARY KEY,
    rule_name VARCHAR(100) NOT NULL COMMENT '规则名称',
    credit_type VARCHAR(50) NOT NULL COMMENT '积分类型',
    expiration_days INTEGER NOT NULL COMMENT '过期天数',
    notify_days INTEGER DEFAULT 7 COMMENT '提前通知天数',
    is_active BOOLEAN DEFAULT true COMMENT '是否启用',
    priority INTEGER DEFAULT 0 COMMENT '优先级(数值越大优先级越高)',
    description TEXT COMMENT '规则描述',
    created_by VARCHAR(36) COMMENT '创建人',
    updated_by VARCHAR(36) COMMENT '更新人',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    -- 创建索引
    INDEX idx_expiration_rules_credit_type (credit_type),
    INDEX idx_expiration_rules_is_active (is_active),
    INDEX idx_expiration_rules_priority (priority),
    
    -- 确保同一积分类型的活跃规则唯一性
    UNIQUE KEY uk_expiration_rules_type_active (credit_type, is_active)
);

-- 3. 创建积分过期批次表
CREATE TABLE IF NOT EXISTS credit_expiration_batches (
    id VARCHAR(36) PRIMARY KEY,
    batch_date TIMESTAMP NOT NULL COMMENT '批次日期',
    total_credits INTEGER DEFAULT 0 COMMENT '过期积分总数',
    total_users INTEGER DEFAULT 0 COMMENT '影响用户数',
    total_transactions INTEGER DEFAULT 0 COMMENT '过期交易数',
    status ENUM('processing', 'completed', 'failed') DEFAULT 'processing' COMMENT '处理状态',
    error_message TEXT COMMENT '错误信息',
    started_at TIMESTAMP COMMENT '开始时间',
    completed_at TIMESTAMP COMMENT '完成时间',
    processed_by VARCHAR(50) DEFAULT 'system' COMMENT '处理人',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    -- 创建索引
    INDEX idx_expiration_batches_date (batch_date),
    INDEX idx_expiration_batches_status (status),
    INDEX idx_expiration_batches_created_at (created_at)
);

-- 4. 创建积分过期日志表
CREATE TABLE IF NOT EXISTS credit_expiration_logs (
    id VARCHAR(36) PRIMARY KEY,
    batch_id VARCHAR(36) NOT NULL COMMENT '批次ID',
    user_id VARCHAR(36) NOT NULL COMMENT '用户ID',
    transaction_id VARCHAR(36) NOT NULL COMMENT '交易ID',
    expired_credits INTEGER NOT NULL COMMENT '过期积分数',
    original_amount INTEGER NOT NULL COMMENT '原始积分数',
    expiration_reason VARCHAR(200) DEFAULT 'Reached expiration date' COMMENT '过期原因',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- 创建索引
    INDEX idx_expiration_logs_batch_id (batch_id),
    INDEX idx_expiration_logs_user_id (user_id),
    INDEX idx_expiration_logs_transaction_id (transaction_id),
    INDEX idx_expiration_logs_created_at (created_at),
    
    -- 外键约束
    FOREIGN KEY (batch_id) REFERENCES credit_expiration_batches(id) ON DELETE CASCADE
);

-- 5. 创建积分过期通知表
CREATE TABLE IF NOT EXISTS credit_expiration_notifications (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL COMMENT '用户ID',
    notification_type ENUM('warning', 'expired') NOT NULL COMMENT '通知类型',
    credits_to_expire INTEGER NOT NULL COMMENT '涉及积分数',
    expiration_date TIMESTAMP NOT NULL COMMENT '过期日期',
    notification_sent BOOLEAN DEFAULT false COMMENT '是否已发送',
    notification_time TIMESTAMP COMMENT '通知发送时间',
    notification_error TEXT COMMENT '通知发送错误',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    -- 创建索引
    INDEX idx_expiration_notifications_user_id (user_id),
    INDEX idx_expiration_notifications_type (notification_type),
    INDEX idx_expiration_notifications_date (expiration_date),
    INDEX idx_expiration_notifications_sent (notification_sent),
    INDEX idx_expiration_notifications_created_at (created_at),
    
    -- 确保同一用户同一过期日期的同类型通知唯一性
    UNIQUE KEY uk_expiration_notifications (user_id, notification_type, expiration_date)
);

-- 6. 插入默认过期规则
INSERT INTO credit_expiration_rules (
    id, rule_name, credit_type, expiration_days, notify_days, 
    is_active, priority, description, created_by
) VALUES 
-- 信件相关积分规则
('exp-rule-001', '信件创建积分', 'letter_create', 365, 7, true, 100, '写信和生成编号获得的积分，1年后过期', 'system'),
('exp-rule-002', '信件送达积分', 'letter_delivery', 180, 7, true, 90, '信件送达和阅读获得的积分，6个月后过期', 'system'),
('exp-rule-003', '收信积分', 'letter_receive', 90, 3, true, 80, '收到信件和回信获得的积分，3个月后过期', 'system'),

-- 社交互动积分规则
('exp-rule-004', '社交互动积分', 'social_interaction', 30, 3, true, 70, '点赞等社交互动积分，30天后过期', 'system'),

-- 活动相关积分规则
('exp-rule-005', '写作挑战积分', 'writing_challenge', 180, 7, true, 85, '参与写作挑战获得的积分，6个月后过期', 'system'),
('exp-rule-006', 'AI互动积分', 'ai_interaction', 60, 3, true, 60, 'AI笔友互动积分，60天后过期', 'system'),

-- 信使相关积分规则  
('exp-rule-007', '信使活动积分', 'courier_activity', 365, 14, true, 95, '信使任务和送达积分，1年后过期', 'system'),

-- 博物馆相关积分规则
('exp-rule-008', '博物馆活动积分', 'museum_activity', 730, 14, true, 110, '博物馆投稿和审核积分，2年后过期', 'system'),

-- OP Code相关积分规则
('exp-rule-009', 'OP码申请积分', 'opcode_activity', 365, 7, true, 75, 'OP码申请审核积分，1年后过期', 'system'),

-- 社区贡献积分规则
('exp-rule-010', '社区贡献积分', 'community_badge', 1095, 30, true, 120, '社区贡献徽章积分，3年后过期', 'system'),

-- 管理员奖励积分规则
('exp-rule-011', '管理员奖励积分', 'admin_reward', 730, 14, true, 105, '管理员手动奖励积分，2年后过期', 'system'),

-- 商业活动积分规则
('exp-rule-012', '商业活动积分', 'commerce_activity', 180, 7, true, 65, '购买和绑定信封积分，6个月后过期', 'system'),

-- 默认规则（最低优先级）
('exp-rule-default', '默认积分规则', 'default', 365, 7, true, 1, '未匹配其他规则的积分默认过期时间', 'system')

ON DUPLICATE KEY UPDATE
    rule_name = VALUES(rule_name),
    expiration_days = VALUES(expiration_days),
    notify_days = VALUES(notify_days),
    description = VALUES(description),
    updated_at = CURRENT_TIMESTAMP;

-- 7. 创建视图：即将过期的积分统计
CREATE OR REPLACE VIEW v_expiring_credits_summary AS
SELECT 
    u.username,
    u.id as user_id,
    COUNT(*) as expiring_transactions,
    SUM(ct.amount) as total_expiring_credits,
    MIN(ct.expires_at) as earliest_expiration,
    MAX(ct.expires_at) as latest_expiration
FROM credit_transactions ct
JOIN users u ON ct.user_id = u.id
WHERE ct.expires_at IS NOT NULL 
    AND ct.expires_at <= DATE_ADD(CURRENT_TIMESTAMP, INTERVAL 30 DAY)
    AND ct.is_expired = false
    AND ct.amount > 0
GROUP BY u.id, u.username
ORDER BY earliest_expiration ASC;

-- 8. 创建视图：过期积分统计
CREATE OR REPLACE VIEW v_expired_credits_summary AS
SELECT 
    DATE(cel.created_at) as expiry_date,
    COUNT(DISTINCT cel.user_id) as affected_users,
    COUNT(*) as expired_transactions,
    SUM(cel.expired_credits) as total_expired_credits,
    ceb.status as batch_status
FROM credit_expiration_logs cel
JOIN credit_expiration_batches ceb ON cel.batch_id = ceb.id
GROUP BY DATE(cel.created_at), ceb.status
ORDER BY expiry_date DESC;

-- 9. 添加注释
ALTER TABLE credit_transactions COMMENT = '积分交易记录表 - Phase 4.1扩展支持过期功能';
ALTER TABLE credit_expiration_rules COMMENT = '积分过期规则配置表';
ALTER TABLE credit_expiration_batches COMMENT = '积分过期批次处理记录表';
ALTER TABLE credit_expiration_logs COMMENT = '积分过期详细日志表';
ALTER TABLE credit_expiration_notifications COMMENT = '积分过期通知记录表';

-- 10. 权限设置（如果需要）
-- GRANT SELECT, INSERT, UPDATE ON credit_expiration_rules TO 'app_user'@'%';
-- GRANT SELECT, INSERT, UPDATE ON credit_expiration_batches TO 'app_user'@'%';
-- GRANT SELECT, INSERT ON credit_expiration_logs TO 'app_user'@'%';
-- GRANT SELECT, INSERT, UPDATE ON credit_expiration_notifications TO 'app_user'@'%';

-- Migration completed: Phase 4.1 Credit Expiration System
-- This migration adds comprehensive credit expiration functionality:
-- 1. Extended credit_transactions with expiration fields
-- 2. Credit expiration rules for different credit types  
-- 3. Batch processing for efficient expiration handling
-- 4. Detailed logging for audit trails
-- 5. Notification system for user warnings
-- 6. Pre-configured rules for all credit types
-- 7. Performance optimized with proper indexes
-- 8. Statistical views for monitoring and reporting