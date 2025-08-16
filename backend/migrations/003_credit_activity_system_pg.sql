-- Migration: Phase 3 - Credit Activity System (PostgreSQL)
-- Description: Add comprehensive credit activity functionality
-- Author: Claude Code Assistant  
-- Date: 2025-08-15

-- =====================================================
-- Phase 3: 积分活动系统数据库迁移 (PostgreSQL版本)
-- =====================================================

-- 1. 创建积分活动表
CREATE TABLE IF NOT EXISTS credit_activities (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    activity_type VARCHAR(50) NOT NULL CHECK (activity_type IN ('daily', 'weekly', 'monthly', 'seasonal', 'first_time', 'cumulative', 'time_limited')),
    status VARCHAR(50) DEFAULT 'draft' CHECK (status IN ('draft', 'active', 'paused', 'completed', 'canceled')),
    description TEXT,
    start_time TIMESTAMP,
    end_time TIMESTAMP,
    reward_type VARCHAR(50) NOT NULL CHECK (reward_type IN ('fixed', 'percentage', 'tiered')),
    reward_amount INTEGER NOT NULL,
    max_participants INTEGER DEFAULT 0,
    current_participants INTEGER DEFAULT 0,
    completion_requirement JSONB,
    trigger_conditions JSONB,
    target_audience_type VARCHAR(50) DEFAULT 'all' CHECK (target_audience_type IN ('all', 'new_users', 'level', 'school', 'custom')),
    target_audience_criteria JSONB,
    priority INTEGER DEFAULT 0,
    is_recurring BOOLEAN DEFAULT FALSE,
    recurring_pattern VARCHAR(100),
    max_rewards_per_user INTEGER DEFAULT 1,
    cooldown_hours INTEGER DEFAULT 0,
    created_by VARCHAR(36),
    updated_by VARCHAR(36),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- 约束检查
    CONSTRAINT chk_activity_reward_amount_positive CHECK (reward_amount > 0),
    CONSTRAINT chk_activity_max_participants_non_negative CHECK (max_participants >= 0),
    CONSTRAINT chk_activity_current_participants_non_negative CHECK (current_participants >= 0),
    CONSTRAINT chk_activity_priority_non_negative CHECK (priority >= 0),
    CONSTRAINT chk_activity_max_rewards_positive CHECK (max_rewards_per_user > 0),
    CONSTRAINT chk_activity_cooldown_non_negative CHECK (cooldown_hours >= 0),
    CONSTRAINT chk_activity_time_range CHECK (end_time IS NULL OR start_time IS NULL OR end_time > start_time)
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_activities_type ON credit_activities(activity_type);
CREATE INDEX IF NOT EXISTS idx_activities_status ON credit_activities(status);
CREATE INDEX IF NOT EXISTS idx_activities_start_time ON credit_activities(start_time);
CREATE INDEX IF NOT EXISTS idx_activities_end_time ON credit_activities(end_time);
CREATE INDEX IF NOT EXISTS idx_activities_target_audience ON credit_activities(target_audience_type);
CREATE INDEX IF NOT EXISTS idx_activities_priority ON credit_activities(priority);
CREATE INDEX IF NOT EXISTS idx_activities_created_at ON credit_activities(created_at);

-- 2. 创建积分活动参与记录表
CREATE TABLE IF NOT EXISTS credit_activity_participations (
    id VARCHAR(36) PRIMARY KEY,
    activity_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    participation_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completion_status VARCHAR(50) DEFAULT 'pending' CHECK (completion_status IN ('pending', 'completed', 'failed')),
    completion_time TIMESTAMP,
    progress_data JSONB,
    reward_amount INTEGER DEFAULT 0,
    is_rewarded BOOLEAN DEFAULT FALSE,
    reward_time TIMESTAMP,
    failure_reason TEXT,
    extra_data JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- 外键约束
    CONSTRAINT fk_participations_activity FOREIGN KEY (activity_id) REFERENCES credit_activities(id) ON DELETE CASCADE,
    
    -- 检查约束
    CONSTRAINT chk_participation_reward_amount_non_negative CHECK (reward_amount >= 0)
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_participations_activity_id ON credit_activity_participations(activity_id);
CREATE INDEX IF NOT EXISTS idx_participations_user_id ON credit_activity_participations(user_id);
CREATE INDEX IF NOT EXISTS idx_participations_status ON credit_activity_participations(completion_status);
CREATE INDEX IF NOT EXISTS idx_participations_participation_time ON credit_activity_participations(participation_time);
CREATE INDEX IF NOT EXISTS idx_participations_completion_time ON credit_activity_participations(completion_time);
CREATE INDEX IF NOT EXISTS idx_participations_is_rewarded ON credit_activity_participations(is_rewarded);
CREATE INDEX IF NOT EXISTS idx_participations_activity_user ON credit_activity_participations(activity_id, user_id);
CREATE INDEX IF NOT EXISTS idx_participations_unique_activity_user_time ON credit_activity_participations(activity_id, user_id, DATE(participation_time));

-- 3. 创建积分活动奖励记录表
CREATE TABLE IF NOT EXISTS credit_activity_rewards (
    id VARCHAR(36) PRIMARY KEY,
    activity_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    participation_id VARCHAR(36),
    reward_amount INTEGER NOT NULL,
    reward_type VARCHAR(50) NOT NULL,
    transaction_id VARCHAR(36),
    awarded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(50) DEFAULT 'pending' CHECK (status IN ('pending', 'completed', 'failed', 'canceled')),
    failure_reason TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- 外键约束
    CONSTRAINT fk_rewards_activity FOREIGN KEY (activity_id) REFERENCES credit_activities(id) ON DELETE CASCADE,
    CONSTRAINT fk_rewards_participation FOREIGN KEY (participation_id) REFERENCES credit_activity_participations(id) ON DELETE SET NULL,
    
    -- 检查约束
    CONSTRAINT chk_reward_amount_positive CHECK (reward_amount > 0)
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_rewards_activity_id ON credit_activity_rewards(activity_id);
CREATE INDEX IF NOT EXISTS idx_rewards_user_id ON credit_activity_rewards(user_id);
CREATE INDEX IF NOT EXISTS idx_rewards_participation_id ON credit_activity_rewards(participation_id);
CREATE INDEX IF NOT EXISTS idx_rewards_transaction_id ON credit_activity_rewards(transaction_id);
CREATE INDEX IF NOT EXISTS idx_rewards_status ON credit_activity_rewards(status);
CREATE INDEX IF NOT EXISTS idx_rewards_awarded_at ON credit_activity_rewards(awarded_at);
CREATE INDEX IF NOT EXISTS idx_rewards_activity_user ON credit_activity_rewards(activity_id, user_id);

-- 4. 创建积分活动规则表
CREATE TABLE IF NOT EXISTS credit_activity_rules (
    id VARCHAR(36) PRIMARY KEY,
    activity_id VARCHAR(36) NOT NULL,
    rule_type VARCHAR(50) NOT NULL CHECK (rule_type IN ('trigger', 'completion', 'reward', 'restriction')),
    rule_name VARCHAR(100) NOT NULL,
    rule_conditions JSONB NOT NULL,
    rule_actions JSONB,
    priority INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- 外键约束
    CONSTRAINT fk_activity_rules_activity FOREIGN KEY (activity_id) REFERENCES credit_activities(id) ON DELETE CASCADE
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_activity_rules_activity_id ON credit_activity_rules(activity_id);
CREATE INDEX IF NOT EXISTS idx_activity_rules_type ON credit_activity_rules(rule_type);
CREATE INDEX IF NOT EXISTS idx_activity_rules_priority ON credit_activity_rules(priority);
CREATE INDEX IF NOT EXISTS idx_activity_rules_is_active ON credit_activity_rules(is_active);

-- 5. 创建积分活动模板表
CREATE TABLE IF NOT EXISTS credit_activity_templates (
    id VARCHAR(36) PRIMARY KEY,
    template_name VARCHAR(200) NOT NULL,
    template_type VARCHAR(50) NOT NULL,
    description TEXT,
    template_data JSONB NOT NULL,
    category VARCHAR(100),
    tags JSONB,
    usage_count INTEGER DEFAULT 0,
    is_public BOOLEAN DEFAULT TRUE,
    created_by VARCHAR(36),
    updated_by VARCHAR(36),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_activity_templates_type ON credit_activity_templates(template_type);
CREATE INDEX IF NOT EXISTS idx_activity_templates_category ON credit_activity_templates(category);
CREATE INDEX IF NOT EXISTS idx_activity_templates_is_public ON credit_activity_templates(is_public);
CREATE INDEX IF NOT EXISTS idx_activity_templates_usage_count ON credit_activity_templates(usage_count);
CREATE INDEX IF NOT EXISTS idx_activity_templates_created_at ON credit_activity_templates(created_at);

-- 6. 创建积分活动目标用户表
CREATE TABLE IF NOT EXISTS credit_activity_target_audiences (
    id VARCHAR(36) PRIMARY KEY,
    activity_id VARCHAR(36) NOT NULL,
    target_type VARCHAR(50) NOT NULL CHECK (target_type IN ('user_level', 'school', 'region', 'user_group', 'custom')),
    target_value VARCHAR(100) NOT NULL,
    inclusion_rule VARCHAR(50) DEFAULT 'include' CHECK (inclusion_rule IN ('include', 'exclude')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- 外键约束
    CONSTRAINT fk_target_audiences_activity FOREIGN KEY (activity_id) REFERENCES credit_activities(id) ON DELETE CASCADE,
    
    -- 唯一性约束
    CONSTRAINT uk_target_audiences UNIQUE (activity_id, target_type, target_value)
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_target_audiences_activity_id ON credit_activity_target_audiences(activity_id);
CREATE INDEX IF NOT EXISTS idx_target_audiences_type ON credit_activity_target_audiences(target_type);
CREATE INDEX IF NOT EXISTS idx_target_audiences_value ON credit_activity_target_audiences(target_value);
CREATE INDEX IF NOT EXISTS idx_target_audiences_rule ON credit_activity_target_audiences(inclusion_rule);

-- 7. 创建积分活动调度表
CREATE TABLE IF NOT EXISTS credit_activity_schedules (
    id VARCHAR(36) PRIMARY KEY,
    activity_id VARCHAR(36) NOT NULL,
    schedule_type VARCHAR(50) NOT NULL CHECK (schedule_type IN ('immediate', 'delayed', 'recurring', 'cron')),
    schedule_time TIMESTAMP,
    cron_expression VARCHAR(100),
    status VARCHAR(50) DEFAULT 'pending' CHECK (status IN ('pending', 'running', 'completed', 'failed', 'canceled')),
    last_execution_time TIMESTAMP,
    next_execution_time TIMESTAMP,
    execution_count INTEGER DEFAULT 0,
    max_executions INTEGER DEFAULT 0,
    retry_count INTEGER DEFAULT 0,
    max_retries INTEGER DEFAULT 3,
    failure_reason TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- 外键约束
    CONSTRAINT fk_activity_schedules_activity FOREIGN KEY (activity_id) REFERENCES credit_activities(id) ON DELETE CASCADE,
    
    -- 检查约束
    CONSTRAINT chk_schedule_execution_count_non_negative CHECK (execution_count >= 0),
    CONSTRAINT chk_schedule_max_executions_non_negative CHECK (max_executions >= 0),
    CONSTRAINT chk_schedule_retry_count_non_negative CHECK (retry_count >= 0),
    CONSTRAINT chk_schedule_max_retries_non_negative CHECK (max_retries >= 0)
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_activity_schedules_activity_id ON credit_activity_schedules(activity_id);
CREATE INDEX IF NOT EXISTS idx_activity_schedules_type ON credit_activity_schedules(schedule_type);
CREATE INDEX IF NOT EXISTS idx_activity_schedules_status ON credit_activity_schedules(status);
CREATE INDEX IF NOT EXISTS idx_activity_schedules_schedule_time ON credit_activity_schedules(schedule_time);
CREATE INDEX IF NOT EXISTS idx_activity_schedules_next_execution ON credit_activity_schedules(next_execution_time);
CREATE INDEX IF NOT EXISTS idx_activity_schedules_last_execution ON credit_activity_schedules(last_execution_time);

-- 8. 插入默认活动模板 (使用 INSERT ... ON CONFLICT 替代 ON DUPLICATE KEY UPDATE)
INSERT INTO credit_activity_templates (
    id, template_name, template_type, description, template_data, category, tags, 
    is_public, created_by, updated_by
) VALUES 
-- 每日登录活动模板
('template-daily-login', '每日登录奖励', 'daily', '用户每日登录获得积分奖励', 
'{"activity_type":"daily","reward_type":"fixed","reward_amount":10,"trigger_conditions":{"action":"user_login"},"max_rewards_per_user":1,"cooldown_hours":24}', 
'engagement', '["登录","每日","基础"]', TRUE, 'system', 'system'),

-- 写信活动模板
('template-letter-writing', '写信奖励活动', 'time_limited', '用户写信获得积分奖励', 
'{"activity_type":"time_limited","reward_type":"fixed","reward_amount":50,"trigger_conditions":{"action":"letter_created"},"max_rewards_per_user":5}', 
'content', '["写信","内容创作"]', TRUE, 'system', 'system'),

-- 新用户注册奖励
('template-new-user', '新用户注册奖励', 'first_time', '新用户注册完成后获得积分', 
'{"activity_type":"first_time","reward_type":"fixed","reward_amount":100,"target_audience_type":"new_users","trigger_conditions":{"action":"user_registered"},"max_rewards_per_user":1}', 
'onboarding', '["新用户","注册","欢迎"]', TRUE, 'system', 'system'),

-- 月度活跃用户奖励
('template-monthly-active', '月度活跃用户奖励', 'monthly', '月度活跃用户获得额外积分', 
'{"activity_type":"monthly","reward_type":"tiered","reward_amount":200,"completion_requirement":{"min_actions":20},"trigger_conditions":{"action":"monthly_summary"}}', 
'retention', '["月度","活跃","留存"]', TRUE, 'system', 'system'),

-- 社交互动奖励
('template-social-interaction', '社交互动奖励', 'cumulative', '用户互动行为累计奖励', 
'{"activity_type":"cumulative","reward_type":"percentage","reward_amount":5,"trigger_conditions":{"action":"social_interaction"},"completion_requirement":{"cumulative_target":100}}', 
'social', '["社交","互动","点赞"]', TRUE, 'system', 'system')

ON CONFLICT (id) DO UPDATE SET
    template_name = EXCLUDED.template_name,
    description = EXCLUDED.description,
    template_data = EXCLUDED.template_data,
    updated_at = CURRENT_TIMESTAMP;

-- 9. 创建视图：活动统计汇总
CREATE OR REPLACE VIEW v_credit_activity_stats AS
SELECT 
    ca.id as activity_id,
    ca.name as activity_name,
    ca.activity_type,
    ca.status,
    ca.current_participants,
    ca.max_participants,
    COALESCE(participation_stats.total_participations, 0) as total_participations,
    COALESCE(participation_stats.completed_participations, 0) as completed_participations,
    COALESCE(participation_stats.pending_participations, 0) as pending_participations,
    COALESCE(reward_stats.total_rewards_amount, 0) as total_rewards_amount,
    COALESCE(reward_stats.total_rewards_count, 0) as total_rewards_count,
    CASE 
        WHEN ca.max_participants > 0 THEN 
            ROUND((ca.current_participants::numeric / ca.max_participants) * 100, 2)
        ELSE NULL 
    END as participation_rate,
    ca.created_at,
    ca.start_time,
    ca.end_time
FROM credit_activities ca
LEFT JOIN (
    SELECT 
        activity_id,
        COUNT(*) as total_participations,
        SUM(CASE WHEN completion_status = 'completed' THEN 1 ELSE 0 END) as completed_participations,
        SUM(CASE WHEN completion_status = 'pending' THEN 1 ELSE 0 END) as pending_participations
    FROM credit_activity_participations 
    GROUP BY activity_id
) participation_stats ON ca.id = participation_stats.activity_id
LEFT JOIN (
    SELECT 
        activity_id,
        SUM(reward_amount) as total_rewards_amount,
        COUNT(*) as total_rewards_count
    FROM credit_activity_rewards 
    WHERE status = 'completed'
    GROUP BY activity_id
) reward_stats ON ca.id = reward_stats.activity_id;

-- 10. 创建视图：用户活动参与汇总
CREATE OR REPLACE VIEW v_user_activity_summary AS
SELECT 
    u.id as user_id,
    u.username,
    COUNT(DISTINCT cap.activity_id) as activities_participated,
    COUNT(cap.id) as total_participations,
    SUM(CASE WHEN cap.completion_status = 'completed' THEN 1 ELSE 0 END) as completed_participations,
    COALESCE(SUM(car.reward_amount), 0) as total_rewards_earned,
    MAX(cap.participation_time) as last_participation_time,
    ROUND(
        SUM(CASE WHEN cap.completion_status = 'completed' THEN 1 ELSE 0 END) * 100.0 / 
        NULLIF(COUNT(cap.id), 0), 2
    ) as completion_rate
FROM users u
LEFT JOIN credit_activity_participations cap ON u.id = cap.user_id
LEFT JOIN credit_activity_rewards car ON cap.id = car.participation_id AND car.status = 'completed'
GROUP BY u.id, u.username;

-- 11. 添加表注释
COMMENT ON TABLE credit_activities IS '积分活动主表 - Phase 3';
COMMENT ON TABLE credit_activity_participations IS '积分活动参与记录表';
COMMENT ON TABLE credit_activity_rewards IS '积分活动奖励记录表';
COMMENT ON TABLE credit_activity_rules IS '积分活动规则配置表';
COMMENT ON TABLE credit_activity_templates IS '积分活动模板表';
COMMENT ON TABLE credit_activity_target_audiences IS '积分活动目标用户表';
COMMENT ON TABLE credit_activity_schedules IS '积分活动调度表';

-- Migration completed: Phase 3 Credit Activity System
-- This migration adds comprehensive credit activity functionality:
-- 1. Main activity management with flexible types and statuses
-- 2. User participation tracking with progress monitoring
-- 3. Reward distribution system with multiple reward types
-- 4. Flexible rule engine for activity logic
-- 5. Template system for quick activity creation
-- 6. Target audience management for precise user targeting
-- 7. Advanced scheduling system with cron support
-- 8. Statistical views for monitoring and analysis
-- 9. Performance optimized with proper indexes
-- 10. Pre-configured templates for common activity types