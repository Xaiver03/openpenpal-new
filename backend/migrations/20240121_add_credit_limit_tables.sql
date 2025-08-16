-- Migration: Add Credit Limit and Anti-Fraud Tables
-- Created: 2024-01-21
-- Purpose: Add tables for credit rate limiting and fraud detection

-- 积分限制规则表
CREATE TABLE IF NOT EXISTS credit_limit_rules (
    id VARCHAR(36) PRIMARY KEY,
    action_type VARCHAR(50) NOT NULL,
    limit_type VARCHAR(20) NOT NULL CHECK (limit_type IN ('count', 'points', 'combined')),
    limit_period VARCHAR(20) NOT NULL CHECK (limit_period IN ('daily', 'weekly', 'monthly')),
    max_count INTEGER NOT NULL DEFAULT 0,
    max_points INTEGER DEFAULT 0,
    enabled BOOLEAN DEFAULT true,
    priority INTEGER DEFAULT 100,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 添加索引
CREATE INDEX IF NOT EXISTS idx_credit_limit_rules_action_enabled 
ON credit_limit_rules(action_type, enabled);

CREATE INDEX IF NOT EXISTS idx_credit_limit_rules_priority 
ON credit_limit_rules(priority, enabled);

-- 用户积分行为记录表
CREATE TABLE IF NOT EXISTS user_credit_actions (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    action_type VARCHAR(50) NOT NULL,
    points INTEGER NOT NULL,
    ip_address VARCHAR(45),
    device_id VARCHAR(100),
    user_agent TEXT,
    reference VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 添加索引
CREATE INDEX IF NOT EXISTS idx_user_credit_actions_user_action_time 
ON user_credit_actions(user_id, action_type, created_at);

CREATE INDEX IF NOT EXISTS idx_user_credit_actions_user_time 
ON user_credit_actions(user_id, created_at);

CREATE INDEX IF NOT EXISTS idx_user_credit_actions_reference 
ON user_credit_actions(reference);

CREATE INDEX IF NOT EXISTS idx_user_credit_actions_ip 
ON user_credit_actions(ip_address, created_at);

-- 风险用户表
CREATE TABLE IF NOT EXISTS credit_risk_users (
    user_id VARCHAR(36) PRIMARY KEY,
    risk_score DECIMAL(5,2) DEFAULT 0.00,
    risk_level VARCHAR(20) DEFAULT 'low' CHECK (risk_level IN ('low', 'medium', 'high', 'blocked')),
    blocked_until TIMESTAMP,
    reason TEXT,
    notes TEXT,
    last_alert_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 添加索引
CREATE INDEX IF NOT EXISTS idx_credit_risk_users_level_score 
ON credit_risk_users(risk_level, risk_score);

CREATE INDEX IF NOT EXISTS idx_credit_risk_users_blocked 
ON credit_risk_users(blocked_until) WHERE blocked_until IS NOT NULL;

-- 插入默认限制规则（基于FSD规格）
INSERT INTO credit_limit_rules (id, action_type, limit_type, limit_period, max_count, max_points, description, priority) VALUES
-- 写信相关限制
('rule-001', 'letter_created', 'count', 'daily', 3, 30, '每日最多创建3封信', 10),
('rule-002', 'letter_reply', 'count', 'daily', 5, 25, '每日最多被回信5次', 10),
('rule-003', 'public_letter_like', 'points', 'daily', 0, 20, '每日公开信点赞积分上限20', 20),

-- AI相关限制
('rule-004', 'ai_interaction', 'count', 'daily', 10, 30, '每日AI互动限制10次', 15),

-- 信使相关限制
('rule-005', 'courier_delivery', 'count', 'daily', 50, 250, '每日信使投递限制50次', 5),
('rule-006', 'courier_first_task', 'count', 'monthly', 1, 20, '每月首次任务奖励限制1次', 1),

-- 博物馆相关限制
('rule-007', 'museum_submit', 'count', 'weekly', 5, 75, '每周博物馆投稿限制5次', 25),
('rule-008', 'museum_liked', 'count', 'daily', 20, 40, '每日博物馆点赞限制20次', 30),

-- 社区奖励限制
('rule-009', 'community_badge', 'count', 'monthly', 3, 150, '每月社区徽章限制3次', 5),
('rule-010', 'admin_reward', 'points', 'daily', 0, 500, '每日管理员奖励积分上限500', 1);

-- 创建触发器自动更新 updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 为规则表添加触发器
DROP TRIGGER IF EXISTS update_credit_limit_rules_updated_at ON credit_limit_rules;
CREATE TRIGGER update_credit_limit_rules_updated_at
    BEFORE UPDATE ON credit_limit_rules
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- 为风险用户表添加触发器
DROP TRIGGER IF EXISTS update_credit_risk_users_updated_at ON credit_risk_users;
CREATE TRIGGER update_credit_risk_users_updated_at
    BEFORE UPDATE ON credit_risk_users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- 添加外键约束（如果用户表存在）
DO $$
BEGIN
    -- 检查users表是否存在
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'users') THEN
        -- 为user_credit_actions添加外键
        IF NOT EXISTS (
            SELECT constraint_name 
            FROM information_schema.table_constraints 
            WHERE table_name = 'user_credit_actions' 
            AND constraint_name = 'fk_user_credit_actions_user_id'
        ) THEN
            ALTER TABLE user_credit_actions 
            ADD CONSTRAINT fk_user_credit_actions_user_id 
            FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
        END IF;
        
        -- 为credit_risk_users添加外键
        IF NOT EXISTS (
            SELECT constraint_name 
            FROM information_schema.table_constraints 
            WHERE table_name = 'credit_risk_users' 
            AND constraint_name = 'fk_credit_risk_users_user_id'
        ) THEN
            ALTER TABLE credit_risk_users 
            ADD CONSTRAINT fk_credit_risk_users_user_id 
            FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
        END IF;
    END IF;
END $$;

-- Phase 1.3: 防作弊检测日志表
CREATE TABLE IF NOT EXISTS fraud_detection_logs (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    action_type VARCHAR(50) NOT NULL,
    risk_score DECIMAL(3,2) NOT NULL DEFAULT 0.00,
    is_anomalous BOOLEAN NOT NULL DEFAULT FALSE,
    detected_patterns TEXT, -- JSON array of detected patterns
    evidence TEXT,          -- JSON object containing evidence data
    recommendations TEXT,   -- JSON array of recommendations
    alert_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 为检测日志表创建索引
CREATE INDEX IF NOT EXISTS idx_fraud_detection_logs_user_id ON fraud_detection_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_fraud_detection_logs_created_at ON fraud_detection_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_fraud_detection_logs_risk_score ON fraud_detection_logs(risk_score);
CREATE INDEX IF NOT EXISTS idx_fraud_detection_logs_anomalous ON fraud_detection_logs(is_anomalous);
CREATE INDEX IF NOT EXISTS idx_fraud_detection_logs_user_time ON fraud_detection_logs(user_id, created_at);

-- 为检测日志表添加外键（如果用户表存在）
DO $$
BEGIN
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'users') THEN
        IF NOT EXISTS (
            SELECT constraint_name 
            FROM information_schema.table_constraints 
            WHERE table_name = 'fraud_detection_logs' 
            AND constraint_name = 'fk_fraud_detection_logs_user_id'
        ) THEN
            ALTER TABLE fraud_detection_logs 
            ADD CONSTRAINT fk_fraud_detection_logs_user_id 
            FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
        END IF;
    END IF;
END $$;