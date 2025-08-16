-- Migration: Phase 4.1 - Credit Expiration System (PostgreSQL)
-- Description: Add credit expiration functionality with batch processing
-- Author: Claude Code Assistant
-- Date: 2025-08-15

-- =====================================================
-- Phase 4.1: 积分过期系统数据库迁移 (PostgreSQL版本)
-- =====================================================

-- 1. 更新积分交易表，添加过期相关字段（如果不存在）
ALTER TABLE credit_transactions 
ADD COLUMN IF NOT EXISTS expires_at TIMESTAMP;

ALTER TABLE credit_transactions 
ADD COLUMN IF NOT EXISTS expired_at TIMESTAMP;

ALTER TABLE credit_transactions 
ADD COLUMN IF NOT EXISTS is_expired BOOLEAN DEFAULT FALSE;

-- 创建索引（如果不存在）
CREATE INDEX IF NOT EXISTS idx_credit_transactions_expires_at ON credit_transactions(expires_at);
CREATE INDEX IF NOT EXISTS idx_credit_transactions_expired_at ON credit_transactions(expired_at);
CREATE INDEX IF NOT EXISTS idx_credit_transactions_is_expired ON credit_transactions(is_expired);

-- 2. 创建积分过期规则表
CREATE TABLE IF NOT EXISTS credit_expiration_rules (
    id VARCHAR(36) PRIMARY KEY,
    rule_name VARCHAR(100) NOT NULL,
    credit_type VARCHAR(50) NOT NULL,
    expiration_days INTEGER NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    notification_days_before INTEGER DEFAULT 7,
    auto_extend_days INTEGER DEFAULT 0,
    description TEXT,
    rule_conditions JSONB,
    created_by VARCHAR(36),
    updated_by VARCHAR(36),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- 检查约束
    CONSTRAINT chk_expiration_days_positive CHECK (expiration_days > 0),
    CONSTRAINT chk_notification_days_non_negative CHECK (notification_days_before >= 0),
    CONSTRAINT chk_auto_extend_days_non_negative CHECK (auto_extend_days >= 0)
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_expiration_rules_credit_type ON credit_expiration_rules(credit_type);
CREATE INDEX IF NOT EXISTS idx_expiration_rules_is_active ON credit_expiration_rules(is_active);
CREATE INDEX IF NOT EXISTS idx_expiration_rules_expiration_days ON credit_expiration_rules(expiration_days);

-- 3. 创建积分过期批次表
CREATE TABLE IF NOT EXISTS credit_expiration_batches (
    id VARCHAR(36) PRIMARY KEY,
    batch_date TIMESTAMP NOT NULL,
    total_credits INTEGER DEFAULT 0,
    total_users INTEGER DEFAULT 0,
    processed_credits INTEGER DEFAULT 0,
    expired_credits INTEGER DEFAULT 0,
    status VARCHAR(50) DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'failed', 'canceled')),
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    failure_reason TEXT,
    batch_metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- 检查约束
    CONSTRAINT chk_batch_total_credits_non_negative CHECK (total_credits >= 0),
    CONSTRAINT chk_batch_total_users_non_negative CHECK (total_users >= 0),
    CONSTRAINT chk_batch_processed_credits_non_negative CHECK (processed_credits >= 0),
    CONSTRAINT chk_batch_expired_credits_non_negative CHECK (expired_credits >= 0)
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_expiration_batches_batch_date ON credit_expiration_batches(batch_date);
CREATE INDEX IF NOT EXISTS idx_expiration_batches_status ON credit_expiration_batches(status);
CREATE INDEX IF NOT EXISTS idx_expiration_batches_started_at ON credit_expiration_batches(started_at);
CREATE INDEX IF NOT EXISTS idx_expiration_batches_completed_at ON credit_expiration_batches(completed_at);

-- 4. 创建积分过期详细日志表
CREATE TABLE IF NOT EXISTS credit_expiration_logs (
    id VARCHAR(36) PRIMARY KEY,
    batch_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    transaction_id VARCHAR(36) NOT NULL,
    original_amount INTEGER NOT NULL,
    expired_amount INTEGER NOT NULL,
    expiration_reason VARCHAR(255),
    expired_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    notification_sent BOOLEAN DEFAULT FALSE,
    notification_sent_at TIMESTAMP,
    recovery_deadline TIMESTAMP,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- 外键约束
    CONSTRAINT fk_expiration_logs_batch FOREIGN KEY (batch_id) REFERENCES credit_expiration_batches(id) ON DELETE CASCADE,
    CONSTRAINT fk_expiration_logs_transaction FOREIGN KEY (transaction_id) REFERENCES credit_transactions(id) ON DELETE CASCADE,
    
    -- 检查约束
    CONSTRAINT chk_expiration_original_amount_positive CHECK (original_amount > 0),
    CONSTRAINT chk_expiration_expired_amount_positive CHECK (expired_amount > 0)
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_expiration_logs_batch_id ON credit_expiration_logs(batch_id);
CREATE INDEX IF NOT EXISTS idx_expiration_logs_user_id ON credit_expiration_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_expiration_logs_transaction_id ON credit_expiration_logs(transaction_id);
CREATE INDEX IF NOT EXISTS idx_expiration_logs_expired_at ON credit_expiration_logs(expired_at);
CREATE INDEX IF NOT EXISTS idx_expiration_logs_notification_sent ON credit_expiration_logs(notification_sent);
CREATE INDEX IF NOT EXISTS idx_expiration_logs_recovery_deadline ON credit_expiration_logs(recovery_deadline);

-- 5. 创建积分过期通知表
CREATE TABLE IF NOT EXISTS credit_expiration_notifications (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    notification_type VARCHAR(50) NOT NULL CHECK (notification_type IN ('warning', 'expired', 'summary', 'recovery')),
    expiring_amount INTEGER DEFAULT 0,
    expiring_date TIMESTAMP,
    message TEXT,
    status VARCHAR(50) DEFAULT 'pending' CHECK (status IN ('pending', 'sent', 'failed', 'canceled')),
    sent_at TIMESTAMP,
    delivery_method VARCHAR(50) DEFAULT 'system',
    failure_reason TEXT,
    retry_count INTEGER DEFAULT 0,
    max_retries INTEGER DEFAULT 3,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- 检查约束
    CONSTRAINT chk_notification_expiring_amount_non_negative CHECK (expiring_amount >= 0),
    CONSTRAINT chk_notification_retry_count_non_negative CHECK (retry_count >= 0),
    CONSTRAINT chk_notification_max_retries_non_negative CHECK (max_retries >= 0)
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_expiration_notifications_user_id ON credit_expiration_notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_expiration_notifications_type ON credit_expiration_notifications(notification_type);
CREATE INDEX IF NOT EXISTS idx_expiration_notifications_status ON credit_expiration_notifications(status);
CREATE INDEX IF NOT EXISTS idx_expiration_notifications_expiring_date ON credit_expiration_notifications(expiring_date);
CREATE INDEX IF NOT EXISTS idx_expiration_notifications_sent_at ON credit_expiration_notifications(sent_at);
CREATE INDEX IF NOT EXISTS idx_expiration_notifications_retry_count ON credit_expiration_notifications(retry_count);

-- 6. 插入默认过期规则
INSERT INTO credit_expiration_rules (
    id, rule_name, credit_type, expiration_days, is_active, 
    notification_days_before, auto_extend_days, description, 
    rule_conditions, created_by, updated_by
) VALUES 
-- 任务积分规则
('rule-task-credits', '任务积分过期规则', 'task_reward', 90, TRUE, 7, 0, 
'用户完成任务获得的积分，90天后过期', 
'{"applies_to": ["task_completion", "daily_task", "weekly_task"]}', 
'system', 'system'),

-- 活动积分规则  
('rule-activity-credits', '活动积分过期规则', 'activity_reward', 180, TRUE, 14, 30, 
'用户参与活动获得的积分，180天后过期，可自动延期30天', 
'{"applies_to": ["event_participation", "seasonal_activity"]}', 
'system', 'system'),

-- 购买积分规则
('rule-purchased-credits', '购买积分过期规则', 'purchased', 365, TRUE, 30, 0, 
'用户购买的积分，365天后过期', 
'{"applies_to": ["credit_purchase", "package_purchase"]}', 
'system', 'system'),

-- 奖励积分规则
('rule-bonus-credits', '奖励积分过期规则', 'bonus_reward', 60, TRUE, 5, 0, 
'系统奖励积分，60天后过期', 
'{"applies_to": ["login_bonus", "referral_bonus", "achievement_bonus"]}', 
'system', 'system'),

-- 补偿积分规则
('rule-compensation-credits', '补偿积分过期规则', 'compensation', 120, TRUE, 10, 0, 
'系统补偿积分，120天后过期', 
'{"applies_to": ["system_compensation", "error_compensation"]}', 
'system', 'system')

ON CONFLICT (id) DO UPDATE SET
    rule_name = EXCLUDED.rule_name,
    expiration_days = EXCLUDED.expiration_days,
    notification_days_before = EXCLUDED.notification_days_before,
    auto_extend_days = EXCLUDED.auto_extend_days,
    description = EXCLUDED.description,
    rule_conditions = EXCLUDED.rule_conditions,
    updated_at = CURRENT_TIMESTAMP;

-- 7. 创建视图：即将过期的积分统计
CREATE OR REPLACE VIEW v_expiring_credits_summary AS
SELECT 
    u.id as user_id,
    u.username,
    u.email,
    COUNT(*) as expiring_transactions,
    SUM(ct.amount) as total_expiring_amount,
    MIN(ct.expires_at) as earliest_expiration,
    MAX(ct.expires_at) as latest_expiration,
    STRING_AGG(DISTINCT ct.transaction_type, ', ') as credit_types
FROM users u
JOIN credit_transactions ct ON u.id = ct.user_id
WHERE ct.expires_at IS NOT NULL 
  AND ct.expires_at <= CURRENT_TIMESTAMP + INTERVAL '30 days'
  AND ct.is_expired = FALSE
  AND ct.amount > 0
GROUP BY u.id, u.username, u.email
ORDER BY earliest_expiration ASC;

-- 8. 创建视图：过期批次统计
CREATE OR REPLACE VIEW v_expiration_batch_stats AS
SELECT 
    DATE(ceb.batch_date) as batch_date,
    COUNT(*) as total_batches,
    SUM(ceb.total_credits) as total_credits_processed,
    SUM(ceb.expired_credits) as total_credits_expired,
    SUM(ceb.total_users) as total_users_affected,
    ROUND(
        SUM(ceb.expired_credits) * 100.0 / NULLIF(SUM(ceb.total_credits), 0), 2
    ) as expiration_rate,
    COUNT(CASE WHEN ceb.status = 'completed' THEN 1 END) as completed_batches,
    COUNT(CASE WHEN ceb.status = 'failed' THEN 1 END) as failed_batches
FROM credit_expiration_batches ceb
GROUP BY DATE(ceb.batch_date)
ORDER BY batch_date DESC;

-- 添加表注释
COMMENT ON TABLE credit_expiration_rules IS '积分过期规则配置表 - Phase 4.1';
COMMENT ON TABLE credit_expiration_batches IS '积分过期批次处理表';
COMMENT ON TABLE credit_expiration_logs IS '积分过期详细日志表';
COMMENT ON TABLE credit_expiration_notifications IS '积分过期通知管理表';

-- Migration completed: Phase 4.1 Credit Expiration System
-- This migration adds comprehensive credit expiration functionality:
-- 1. Enhanced credit transactions with expiration tracking
-- 2. Flexible rule-based expiration system supporting 12+ credit types
-- 3. Batch processing for efficient expiration handling
-- 4. Detailed logging and audit trail for all expiration operations
-- 5. Multi-channel notification system with retry logic
-- 6. Statistical views for monitoring and reporting
-- 7. Default rules for common credit types with different expiration periods
-- 8. Support for automatic extension and recovery mechanisms