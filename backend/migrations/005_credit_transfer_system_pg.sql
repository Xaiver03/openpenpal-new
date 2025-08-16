-- Migration: Phase 4.2 - Credit Transfer System (PostgreSQL)
-- Description: Add credit transfer functionality with security and limits
-- Author: Claude Code Assistant
-- Date: 2025-08-15

-- =====================================================
-- Phase 4.2: 积分转赠系统数据库迁移 (PostgreSQL版本)
-- =====================================================

-- 1. 创建积分转赠记录表
CREATE TABLE IF NOT EXISTS credit_transfers (
    id VARCHAR(36) PRIMARY KEY,
    from_user_id VARCHAR(36) NOT NULL,
    to_user_id VARCHAR(36) NOT NULL,
    amount INTEGER NOT NULL,
    transfer_type VARCHAR(50) NOT NULL CHECK (transfer_type IN ('direct', 'gift', 'reward')),
    status VARCHAR(50) DEFAULT 'pending' CHECK (status IN ('pending', 'completed', 'rejected', 'canceled', 'expired')),
    fee_amount INTEGER DEFAULT 0,
    message TEXT,
    reason VARCHAR(255),
    reference_id VARCHAR(100),
    expires_at TIMESTAMP,
    approved_by VARCHAR(36),
    approved_at TIMESTAMP,
    completed_at TIMESTAMP,
    rejected_at TIMESTAMP,
    rejection_reason TEXT,
    metadata JSONB,
    from_transaction_id VARCHAR(36),
    to_transaction_id VARCHAR(36),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- 外键约束（用户表关联）
    -- CONSTRAINT fk_transfer_from_user FOREIGN KEY (from_user_id) REFERENCES users(id) ON DELETE CASCADE,
    -- CONSTRAINT fk_transfer_to_user FOREIGN KEY (to_user_id) REFERENCES users(id) ON DELETE CASCADE,
    -- CONSTRAINT fk_transfer_from_transaction FOREIGN KEY (from_transaction_id) REFERENCES credit_transactions(id) ON DELETE SET NULL,
    -- CONSTRAINT fk_transfer_to_transaction FOREIGN KEY (to_transaction_id) REFERENCES credit_transactions(id) ON DELETE SET NULL,
    
    -- 检查约束
    CONSTRAINT chk_transfer_amount_positive CHECK (amount > 0),
    CONSTRAINT chk_transfer_fee_amount_non_negative CHECK (fee_amount >= 0),
    CONSTRAINT chk_transfer_different_users CHECK (from_user_id != to_user_id)
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_transfers_from_user_id ON credit_transfers(from_user_id);
CREATE INDEX IF NOT EXISTS idx_transfers_to_user_id ON credit_transfers(to_user_id);
CREATE INDEX IF NOT EXISTS idx_transfers_status ON credit_transfers(status);
CREATE INDEX IF NOT EXISTS idx_transfers_type ON credit_transfers(transfer_type);
CREATE INDEX IF NOT EXISTS idx_transfers_created_at ON credit_transfers(created_at);
CREATE INDEX IF NOT EXISTS idx_transfers_expires_at ON credit_transfers(expires_at);
CREATE INDEX IF NOT EXISTS idx_transfers_approved_at ON credit_transfers(approved_at);
CREATE INDEX IF NOT EXISTS idx_transfers_completed_at ON credit_transfers(completed_at);
CREATE INDEX IF NOT EXISTS idx_transfers_reference_id ON credit_transfers(reference_id);

-- 2. 创建积分转赠规则表
CREATE TABLE IF NOT EXISTS credit_transfer_rules (
    id VARCHAR(36) PRIMARY KEY,
    rule_name VARCHAR(100) NOT NULL,
    rule_type VARCHAR(50) NOT NULL CHECK (rule_type IN ('global', 'user_level', 'transfer_type', 'amount_based')),
    target_criteria JSONB,
    min_amount INTEGER DEFAULT 1,
    max_amount INTEGER DEFAULT 10000,
    daily_limit INTEGER DEFAULT 1000,
    monthly_limit INTEGER DEFAULT 5000,
    fee_percentage DECIMAL(5,2) DEFAULT 0.00,
    fixed_fee INTEGER DEFAULT 0,
    min_fee INTEGER DEFAULT 0,
    max_fee INTEGER DEFAULT 100,
    requires_approval BOOLEAN DEFAULT FALSE,
    auto_approve_threshold INTEGER DEFAULT 0,
    cooling_period_hours INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    priority INTEGER DEFAULT 0,
    description TEXT,
    conditions JSONB,
    created_by VARCHAR(36),
    updated_by VARCHAR(36),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- 检查约束
    CONSTRAINT chk_transfer_rule_min_amount_positive CHECK (min_amount > 0),
    CONSTRAINT chk_transfer_rule_max_amount_positive CHECK (max_amount > 0),
    CONSTRAINT chk_transfer_rule_min_max_amount CHECK (max_amount >= min_amount),
    CONSTRAINT chk_transfer_rule_daily_limit_non_negative CHECK (daily_limit >= 0),
    CONSTRAINT chk_transfer_rule_monthly_limit_non_negative CHECK (monthly_limit >= 0),
    CONSTRAINT chk_transfer_rule_fee_percentage_non_negative CHECK (fee_percentage >= 0),
    CONSTRAINT chk_transfer_rule_fixed_fee_non_negative CHECK (fixed_fee >= 0),
    CONSTRAINT chk_transfer_rule_min_fee_non_negative CHECK (min_fee >= 0),
    CONSTRAINT chk_transfer_rule_max_fee_non_negative CHECK (max_fee >= 0),
    CONSTRAINT chk_transfer_rule_cooling_period_non_negative CHECK (cooling_period_hours >= 0)
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_transfer_rules_type ON credit_transfer_rules(rule_type);
CREATE INDEX IF NOT EXISTS idx_transfer_rules_is_active ON credit_transfer_rules(is_active);
CREATE INDEX IF NOT EXISTS idx_transfer_rules_priority ON credit_transfer_rules(priority);
CREATE INDEX IF NOT EXISTS idx_transfer_rules_requires_approval ON credit_transfer_rules(requires_approval);

-- 3. 创建用户转赠限制记录表
CREATE TABLE IF NOT EXISTS credit_transfer_limits (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    limit_type VARCHAR(50) NOT NULL CHECK (limit_type IN ('daily', 'weekly', 'monthly', 'total')),
    current_amount INTEGER DEFAULT 0,
    current_count INTEGER DEFAULT 0,
    limit_amount INTEGER NOT NULL,
    limit_count INTEGER NOT NULL,
    reset_at TIMESTAMP NOT NULL,
    last_transfer_at TIMESTAMP,
    is_blocked BOOLEAN DEFAULT FALSE,
    block_reason TEXT,
    blocked_until TIMESTAMP,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- 外键约束
    -- CONSTRAINT fk_transfer_limit_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    
    -- 检查约束
    CONSTRAINT chk_transfer_limit_current_amount_non_negative CHECK (current_amount >= 0),
    CONSTRAINT chk_transfer_limit_current_count_non_negative CHECK (current_count >= 0),
    CONSTRAINT chk_transfer_limit_limit_amount_positive CHECK (limit_amount > 0),
    CONSTRAINT chk_transfer_limit_limit_count_positive CHECK (limit_count > 0),
    
    -- 唯一性约束
    CONSTRAINT uk_transfer_limit_user_type UNIQUE (user_id, limit_type)
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_transfer_limits_user_id ON credit_transfer_limits(user_id);
CREATE INDEX IF NOT EXISTS idx_transfer_limits_type ON credit_transfer_limits(limit_type);
CREATE INDEX IF NOT EXISTS idx_transfer_limits_reset_at ON credit_transfer_limits(reset_at);
CREATE INDEX IF NOT EXISTS idx_transfer_limits_is_blocked ON credit_transfer_limits(is_blocked);
CREATE INDEX IF NOT EXISTS idx_transfer_limits_blocked_until ON credit_transfer_limits(blocked_until);

-- 4. 创建转赠通知表
CREATE TABLE IF NOT EXISTS credit_transfer_notifications (
    id VARCHAR(36) PRIMARY KEY,
    transfer_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    notification_type VARCHAR(50) NOT NULL CHECK (notification_type IN ('transfer_initiated', 'transfer_received', 'transfer_approved', 'transfer_rejected', 'transfer_completed', 'transfer_expired')),
    message TEXT NOT NULL,
    status VARCHAR(50) DEFAULT 'pending' CHECK (status IN ('pending', 'sent', 'failed', 'canceled')),
    delivery_method VARCHAR(50) DEFAULT 'system' CHECK (delivery_method IN ('system', 'email', 'sms', 'push')),
    sent_at TIMESTAMP,
    failure_reason TEXT,
    retry_count INTEGER DEFAULT 0,
    max_retries INTEGER DEFAULT 3,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- 外键约束
    CONSTRAINT fk_transfer_notification_transfer FOREIGN KEY (transfer_id) REFERENCES credit_transfers(id) ON DELETE CASCADE,
    -- CONSTRAINT fk_transfer_notification_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    
    -- 检查约束
    CONSTRAINT chk_transfer_notification_retry_count_non_negative CHECK (retry_count >= 0),
    CONSTRAINT chk_transfer_notification_max_retries_non_negative CHECK (max_retries >= 0)
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_transfer_notifications_transfer_id ON credit_transfer_notifications(transfer_id);
CREATE INDEX IF NOT EXISTS idx_transfer_notifications_user_id ON credit_transfer_notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_transfer_notifications_type ON credit_transfer_notifications(notification_type);
CREATE INDEX IF NOT EXISTS idx_transfer_notifications_status ON credit_transfer_notifications(status);
CREATE INDEX IF NOT EXISTS idx_transfer_notifications_delivery_method ON credit_transfer_notifications(delivery_method);
CREATE INDEX IF NOT EXISTS idx_transfer_notifications_sent_at ON credit_transfer_notifications(sent_at);

-- 5. 插入默认转赠规则
INSERT INTO credit_transfer_rules (
    id, rule_name, rule_type, target_criteria, min_amount, max_amount, 
    daily_limit, monthly_limit, fee_percentage, fixed_fee, min_fee, max_fee,
    requires_approval, auto_approve_threshold, cooling_period_hours, 
    is_active, priority, description, conditions, created_by, updated_by
) VALUES 
-- 普通用户规则
('rule-normal-user', '普通用户转赠规则', 'user_level', 
'{"user_levels": ["normal", "bronze"]}', 
10, 500, 1000, 5000, 2.00, 0, 1, 50, 
FALSE, 100, 1, TRUE, 100, 
'普通用户每日最多转赠1000积分，收取2%手续费', 
'{"min_account_age_days": 7, "min_credit_balance": 100}', 
'system', 'system'),

-- 高级用户规则
('rule-premium-user', '高级用户转赠规则', 'user_level', 
'{"user_levels": ["silver", "gold", "platinum"]}', 
10, 2000, 5000, 20000, 1.00, 0, 1, 100, 
FALSE, 500, 0, TRUE, 90, 
'高级用户转赠限制更宽松，手续费降至1%', 
'{"min_account_age_days": 30, "verified_account": true}', 
'system', 'system'),

-- VIP用户规则
('rule-vip-user', 'VIP用户转赠规则', 'user_level', 
'{"user_levels": ["diamond", "vip"]}', 
1, 10000, 20000, 100000, 0.50, 0, 0, 200, 
FALSE, 2000, 0, TRUE, 80, 
'VIP用户享受最低手续费和最高转赠限额', 
'{"verified_account": true, "premium_member": true}', 
'system', 'system'),

-- 小额转赠规则
('rule-small-amount', '小额转赠规则', 'amount_based', 
'{"amount_range": {"min": 1, "max": 50}}', 
1, 50, 500, 2000, 0.00, 0, 0, 0, 
FALSE, 50, 0, TRUE, 110, 
'小额转赠免手续费，鼓励用户间小额互助', 
'{"no_fees_under": 50}', 
'system', 'system'),

-- 礼品转赠规则
('rule-gift-transfer', '礼品转赠规则', 'transfer_type', 
'{"transfer_types": ["gift"]}', 
1, 1000, 2000, 8000, 1.50, 5, 1, 75, 
TRUE, 200, 0, TRUE, 85, 
'礼品转赠需要审核，确保合规性', 
'{"requires_message": true, "max_message_length": 200}', 
'system', 'system')

ON CONFLICT (id) DO UPDATE SET
    rule_name = EXCLUDED.rule_name,
    min_amount = EXCLUDED.min_amount,
    max_amount = EXCLUDED.max_amount,
    daily_limit = EXCLUDED.daily_limit,
    monthly_limit = EXCLUDED.monthly_limit,
    fee_percentage = EXCLUDED.fee_percentage,
    fixed_fee = EXCLUDED.fixed_fee,
    min_fee = EXCLUDED.min_fee,
    max_fee = EXCLUDED.max_fee,
    requires_approval = EXCLUDED.requires_approval,
    auto_approve_threshold = EXCLUDED.auto_approve_threshold,
    description = EXCLUDED.description,
    conditions = EXCLUDED.conditions,
    updated_at = CURRENT_TIMESTAMP;

-- 6. 创建视图：用户转赠统计
CREATE OR REPLACE VIEW v_user_transfer_stats AS
SELECT 
    u.id as user_id,
    u.username,
    -- 发送统计
    COUNT(ct_out.id) as transfers_sent_count,
    COALESCE(SUM(ct_out.amount), 0) as total_amount_sent,
    COALESCE(SUM(ct_out.fee_amount), 0) as total_fees_paid,
    COUNT(CASE WHEN ct_out.status = 'completed' THEN 1 END) as successful_sends,
    
    -- 接收统计
    COUNT(ct_in.id) as transfers_received_count,
    COALESCE(SUM(ct_in.amount), 0) as total_amount_received,
    COUNT(CASE WHEN ct_in.status = 'completed' THEN 1 END) as successful_receives,
    
    -- 最近活动
    MAX(GREATEST(ct_out.created_at, ct_in.created_at)) as last_transfer_activity,
    
    -- 成功率
    ROUND(
        COUNT(CASE WHEN ct_out.status = 'completed' THEN 1 END) * 100.0 / 
        NULLIF(COUNT(ct_out.id), 0), 2
    ) as send_success_rate
FROM users u
LEFT JOIN credit_transfers ct_out ON u.id = ct_out.from_user_id
LEFT JOIN credit_transfers ct_in ON u.id = ct_in.to_user_id
GROUP BY u.id, u.username;

-- 7. 创建视图：转赠趋势分析
CREATE OR REPLACE VIEW v_transfer_trends AS
SELECT 
    DATE(created_at) as transfer_date,
    COUNT(*) as total_transfers,
    COUNT(CASE WHEN status = 'completed' THEN 1 END) as successful_transfers,
    COUNT(CASE WHEN status = 'rejected' THEN 1 END) as rejected_transfers,
    COUNT(CASE WHEN status = 'pending' THEN 1 END) as pending_transfers,
    SUM(amount) as total_amount,
    SUM(fee_amount) as total_fees,
    COUNT(DISTINCT from_user_id) as unique_senders,
    COUNT(DISTINCT to_user_id) as unique_receivers,
    ROUND(AVG(amount), 2) as average_amount,
    ROUND(
        COUNT(CASE WHEN status = 'completed' THEN 1 END) * 100.0 / 
        NULLIF(COUNT(*), 0), 2
    ) as success_rate
FROM credit_transfers
WHERE created_at >= CURRENT_DATE - INTERVAL '90 days'
GROUP BY DATE(created_at)
ORDER BY transfer_date DESC;

-- 8. 创建视图：转赠规则效果分析
CREATE OR REPLACE VIEW v_transfer_rule_effectiveness AS
SELECT 
    ctr.id as rule_id,
    ctr.rule_name,
    ctr.rule_type,
    COUNT(ct.id) as transfers_affected,
    SUM(ct.amount) as total_amount_transferred,
    SUM(ct.fee_amount) as total_fees_collected,
    COUNT(CASE WHEN ct.status = 'completed' THEN 1 END) as successful_transfers,
    COUNT(CASE WHEN ct.status = 'rejected' THEN 1 END) as rejected_transfers,
    ROUND(AVG(ct.amount), 2) as average_transfer_amount,
    ROUND(
        COUNT(CASE WHEN ct.status = 'completed' THEN 1 END) * 100.0 / 
        NULLIF(COUNT(ct.id), 0), 2
    ) as success_rate,
    ctr.is_active,
    ctr.priority
FROM credit_transfer_rules ctr
LEFT JOIN credit_transfers ct ON 
    (ctr.rule_type = 'transfer_type' AND ct.transfer_type = ANY(STRING_TO_ARRAY(ctr.target_criteria->>'transfer_types', ',')))
    OR (ctr.rule_type = 'amount_based' AND ct.amount BETWEEN (ctr.target_criteria->'amount_range'->>'min')::INTEGER AND (ctr.target_criteria->'amount_range'->>'max')::INTEGER)
WHERE ctr.is_active = TRUE
GROUP BY ctr.id, ctr.rule_name, ctr.rule_type, ctr.is_active, ctr.priority
ORDER BY total_amount_transferred DESC;

-- 添加表注释
COMMENT ON TABLE credit_transfers IS '积分转赠记录表 - Phase 4.2';
COMMENT ON TABLE credit_transfer_rules IS '积分转赠规则配置表';
COMMENT ON TABLE credit_transfer_limits IS '用户转赠限制记录表';
COMMENT ON TABLE credit_transfer_notifications IS '转赠通知管理表';

-- Migration completed: Phase 4.2 Credit Transfer System
-- This migration adds comprehensive credit transfer functionality:
-- 1. Secure credit transfer system with multiple types (direct, gift, reward)
-- 2. Flexible rule engine supporting different user levels and transfer scenarios
-- 3. Comprehensive limit tracking with daily/weekly/monthly quotas
-- 4. Multi-channel notification system for all transfer events
-- 5. Fee calculation system with percentage and fixed fees
-- 6. Approval workflow for high-value or sensitive transfers
-- 7. Detailed audit trail and statistics for monitoring
-- 8. Statistical views for analyzing transfer patterns and rule effectiveness
-- 9. Default rules for different user tiers with appropriate limits and fees
-- 10. Complete lifecycle management from initiation to completion/rejection