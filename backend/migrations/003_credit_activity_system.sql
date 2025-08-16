-- Migration: Phase 3 - Credit Activity System
-- Description: Add comprehensive credit activity functionality
-- Author: Claude Code Assistant
-- Date: 2025-08-15

-- =====================================================
-- Phase 3: 积分活动系统数据库迁移
-- =====================================================

-- 1. 创建积分活动表
CREATE TABLE IF NOT EXISTS credit_activities (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(200) NOT NULL COMMENT '活动名称',
    activity_type ENUM('daily', 'weekly', 'monthly', 'seasonal', 'first_time', 'cumulative', 'time_limited') NOT NULL COMMENT '活动类型',
    status ENUM('draft', 'active', 'paused', 'completed', 'canceled') DEFAULT 'draft' COMMENT '活动状态',
    description TEXT COMMENT '活动描述',
    start_time TIMESTAMP NULL COMMENT '开始时间',
    end_time TIMESTAMP NULL COMMENT '结束时间',
    reward_type ENUM('fixed', 'percentage', 'tiered') NOT NULL COMMENT '奖励类型',
    reward_amount INTEGER NOT NULL COMMENT '奖励数量',
    max_participants INTEGER DEFAULT 0 COMMENT '最大参与人数，0表示无限制',
    current_participants INTEGER DEFAULT 0 COMMENT '当前参与人数',
    completion_requirement TEXT COMMENT '完成要求JSON',
    trigger_conditions TEXT COMMENT '触发条件JSON',
    target_audience_type ENUM('all', 'new_users', 'level', 'school', 'custom') DEFAULT 'all' COMMENT '目标用户类型',
    target_audience_criteria TEXT COMMENT '目标用户条件JSON',
    priority INTEGER DEFAULT 0 COMMENT '优先级',
    is_recurring BOOLEAN DEFAULT FALSE COMMENT '是否重复',
    recurring_pattern VARCHAR(100) COMMENT '重复模式',
    max_rewards_per_user INTEGER DEFAULT 1 COMMENT '每用户最大奖励次数',
    cooldown_hours INTEGER DEFAULT 0 COMMENT '冷却时间(小时)',
    created_by VARCHAR(36) COMMENT '创建人',
    updated_by VARCHAR(36) COMMENT '更新人',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    -- 创建索引
    INDEX idx_activities_type (activity_type),
    INDEX idx_activities_status (status),
    INDEX idx_activities_start_time (start_time),
    INDEX idx_activities_end_time (end_time),
    INDEX idx_activities_target_audience (target_audience_type),
    INDEX idx_activities_priority (priority),
    INDEX idx_activities_created_at (created_at),
    
    -- 检查约束
    CONSTRAINT chk_activity_reward_amount_positive CHECK (reward_amount > 0),
    CONSTRAINT chk_activity_max_participants_non_negative CHECK (max_participants >= 0),
    CONSTRAINT chk_activity_current_participants_non_negative CHECK (current_participants >= 0),
    CONSTRAINT chk_activity_priority_non_negative CHECK (priority >= 0),
    CONSTRAINT chk_activity_max_rewards_positive CHECK (max_rewards_per_user > 0),
    CONSTRAINT chk_activity_cooldown_non_negative CHECK (cooldown_hours >= 0),
    CONSTRAINT chk_activity_time_range CHECK (end_time IS NULL OR start_time IS NULL OR end_time > start_time)
);

-- 2. 创建积分活动参与记录表
CREATE TABLE IF NOT EXISTS credit_activity_participations (
    id VARCHAR(36) PRIMARY KEY,
    activity_id VARCHAR(36) NOT NULL COMMENT '活动ID',
    user_id VARCHAR(36) NOT NULL COMMENT '用户ID',
    participation_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '参与时间',
    completion_status ENUM('pending', 'completed', 'failed') DEFAULT 'pending' COMMENT '完成状态',
    completion_time TIMESTAMP NULL COMMENT '完成时间',
    progress_data TEXT COMMENT '进度数据JSON',
    reward_amount INTEGER DEFAULT 0 COMMENT '获得奖励数量',
    is_rewarded BOOLEAN DEFAULT FALSE COMMENT '是否已奖励',
    reward_time TIMESTAMP NULL COMMENT '奖励时间',
    failure_reason TEXT COMMENT '失败原因',
    extra_data TEXT COMMENT '额外数据JSON',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    -- 创建索引
    INDEX idx_participations_activity_id (activity_id),
    INDEX idx_participations_user_id (user_id),
    INDEX idx_participations_status (completion_status),
    INDEX idx_participations_participation_time (participation_time),
    INDEX idx_participations_completion_time (completion_time),
    INDEX idx_participations_is_rewarded (is_rewarded),
    INDEX idx_participations_activity_user (activity_id, user_id),
    
    -- 外键约束
    FOREIGN KEY (activity_id) REFERENCES credit_activities(id) ON DELETE CASCADE,
    
    -- 检查约束
    CONSTRAINT chk_participation_reward_amount_non_negative CHECK (reward_amount >= 0),
    
    -- 唯一性约束（防止重复参与，根据活动类型可能需要调整）
    INDEX idx_participations_unique_activity_user_time (activity_id, user_id, DATE(participation_time))
);

-- 3. 创建积分活动奖励记录表
CREATE TABLE IF NOT EXISTS credit_activity_rewards (
    id VARCHAR(36) PRIMARY KEY,
    activity_id VARCHAR(36) NOT NULL COMMENT '活动ID',
    user_id VARCHAR(36) NOT NULL COMMENT '用户ID',
    participation_id VARCHAR(36) COMMENT '参与记录ID',
    reward_amount INTEGER NOT NULL COMMENT '奖励数量',
    reward_type VARCHAR(50) NOT NULL COMMENT '奖励类型',
    transaction_id VARCHAR(36) COMMENT '关联的积分交易ID',
    awarded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '奖励时间',
    status ENUM('pending', 'completed', 'failed', 'canceled') DEFAULT 'pending' COMMENT '奖励状态',
    failure_reason TEXT COMMENT '失败原因',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    -- 创建索引
    INDEX idx_rewards_activity_id (activity_id),
    INDEX idx_rewards_user_id (user_id),
    INDEX idx_rewards_participation_id (participation_id),
    INDEX idx_rewards_transaction_id (transaction_id),
    INDEX idx_rewards_status (status),
    INDEX idx_rewards_awarded_at (awarded_at),
    INDEX idx_rewards_activity_user (activity_id, user_id),
    
    -- 外键约束
    FOREIGN KEY (activity_id) REFERENCES credit_activities(id) ON DELETE CASCADE,
    FOREIGN KEY (participation_id) REFERENCES credit_activity_participations(id) ON DELETE SET NULL,
    
    -- 检查约束
    CONSTRAINT chk_reward_amount_positive CHECK (reward_amount > 0)
);

-- 4. 创建积分活动规则表
CREATE TABLE IF NOT EXISTS credit_activity_rules (
    id VARCHAR(36) PRIMARY KEY,
    activity_id VARCHAR(36) NOT NULL COMMENT '活动ID',
    rule_type ENUM('trigger', 'completion', 'reward', 'restriction') NOT NULL COMMENT '规则类型',
    rule_name VARCHAR(100) NOT NULL COMMENT '规则名称',
    rule_conditions TEXT NOT NULL COMMENT '规则条件JSON',
    rule_actions TEXT COMMENT '规则动作JSON',
    priority INTEGER DEFAULT 0 COMMENT '规则优先级',
    is_active BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    -- 创建索引
    INDEX idx_activity_rules_activity_id (activity_id),
    INDEX idx_activity_rules_type (rule_type),
    INDEX idx_activity_rules_priority (priority),
    INDEX idx_activity_rules_is_active (is_active),
    
    -- 外键约束
    FOREIGN KEY (activity_id) REFERENCES credit_activities(id) ON DELETE CASCADE
);

-- 5. 创建积分活动模板表
CREATE TABLE IF NOT EXISTS credit_activity_templates (
    id VARCHAR(36) PRIMARY KEY,
    template_name VARCHAR(200) NOT NULL COMMENT '模板名称',
    template_type VARCHAR(50) NOT NULL COMMENT '模板类型',
    description TEXT COMMENT '模板描述',
    template_data TEXT NOT NULL COMMENT '模板数据JSON',
    category VARCHAR(100) COMMENT '模板分类',
    tags TEXT COMMENT '标签JSON数组',
    usage_count INTEGER DEFAULT 0 COMMENT '使用次数',
    is_public BOOLEAN DEFAULT TRUE COMMENT '是否公开',
    created_by VARCHAR(36) COMMENT '创建人',
    updated_by VARCHAR(36) COMMENT '更新人',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    -- 创建索引
    INDEX idx_activity_templates_type (template_type),
    INDEX idx_activity_templates_category (category),
    INDEX idx_activity_templates_is_public (is_public),
    INDEX idx_activity_templates_usage_count (usage_count),
    INDEX idx_activity_templates_created_at (created_at)
);

-- 6. 创建积分活动目标用户表
CREATE TABLE IF NOT EXISTS credit_activity_target_audiences (
    id VARCHAR(36) PRIMARY KEY,
    activity_id VARCHAR(36) NOT NULL COMMENT '活动ID',
    target_type ENUM('user_level', 'school', 'region', 'user_group', 'custom') NOT NULL COMMENT '目标类型',
    target_value VARCHAR(100) NOT NULL COMMENT '目标值',
    inclusion_rule ENUM('include', 'exclude') DEFAULT 'include' COMMENT '包含规则',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- 创建索引
    INDEX idx_target_audiences_activity_id (activity_id),
    INDEX idx_target_audiences_type (target_type),
    INDEX idx_target_audiences_value (target_value),
    INDEX idx_target_audiences_rule (inclusion_rule),
    
    -- 外键约束
    FOREIGN KEY (activity_id) REFERENCES credit_activities(id) ON DELETE CASCADE,
    
    -- 唯一性约束
    UNIQUE KEY uk_target_audiences (activity_id, target_type, target_value)
);

-- 7. 创建积分活动调度表
CREATE TABLE IF NOT EXISTS credit_activity_schedules (
    id VARCHAR(36) PRIMARY KEY,
    activity_id VARCHAR(36) NOT NULL COMMENT '活动ID',
    schedule_type ENUM('immediate', 'delayed', 'recurring', 'cron') NOT NULL COMMENT '调度类型',
    schedule_time TIMESTAMP NULL COMMENT '计划执行时间',
    cron_expression VARCHAR(100) COMMENT 'Cron表达式',
    status ENUM('pending', 'running', 'completed', 'failed', 'canceled') DEFAULT 'pending' COMMENT '调度状态',
    last_execution_time TIMESTAMP NULL COMMENT '最后执行时间',
    next_execution_time TIMESTAMP NULL COMMENT '下次执行时间',
    execution_count INTEGER DEFAULT 0 COMMENT '执行次数',
    max_executions INTEGER DEFAULT 0 COMMENT '最大执行次数，0表示无限制',
    retry_count INTEGER DEFAULT 0 COMMENT '重试次数',
    max_retries INTEGER DEFAULT 3 COMMENT '最大重试次数',
    failure_reason TEXT COMMENT '失败原因',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    -- 创建索引
    INDEX idx_activity_schedules_activity_id (activity_id),
    INDEX idx_activity_schedules_type (schedule_type),
    INDEX idx_activity_schedules_status (status),
    INDEX idx_activity_schedules_schedule_time (schedule_time),
    INDEX idx_activity_schedules_next_execution (next_execution_time),
    INDEX idx_activity_schedules_last_execution (last_execution_time),
    
    -- 外键约束
    FOREIGN KEY (activity_id) REFERENCES credit_activities(id) ON DELETE CASCADE,
    
    -- 检查约束
    CONSTRAINT chk_schedule_execution_count_non_negative CHECK (execution_count >= 0),
    CONSTRAINT chk_schedule_max_executions_non_negative CHECK (max_executions >= 0),
    CONSTRAINT chk_schedule_retry_count_non_negative CHECK (retry_count >= 0),
    CONSTRAINT chk_schedule_max_retries_non_negative CHECK (max_retries >= 0)
);

-- 8. 插入默认活动模板
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

ON DUPLICATE KEY UPDATE
    template_name = VALUES(template_name),
    description = VALUES(description),
    template_data = VALUES(template_data),
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
            ROUND((ca.current_participants / ca.max_participants) * 100, 2)
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
ALTER TABLE credit_activities COMMENT = '积分活动主表 - Phase 3';
ALTER TABLE credit_activity_participations COMMENT = '积分活动参与记录表';
ALTER TABLE credit_activity_rewards COMMENT = '积分活动奖励记录表';
ALTER TABLE credit_activity_rules COMMENT = '积分活动规则配置表';
ALTER TABLE credit_activity_templates COMMENT = '积分活动模板表';
ALTER TABLE credit_activity_target_audiences COMMENT = '积分活动目标用户表';
ALTER TABLE credit_activity_schedules COMMENT = '积分活动调度表';

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