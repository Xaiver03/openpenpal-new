-- 修复 AI 配置相关的数据库表结构问题
-- 执行时间: 2025-08-22

-- 1. 备份现有的 ai_configs 数据
CREATE TABLE IF NOT EXISTS ai_configs_backup AS SELECT * FROM ai_configs;

-- 2. 重命名旧表为 ai_provider_configs
ALTER TABLE ai_configs RENAME TO ai_provider_configs;

-- 3. 创建新的 ai_configs 表，用于存储动态配置
CREATE TABLE ai_configs (
    id VARCHAR(36) PRIMARY KEY,
    config_type VARCHAR(50) NOT NULL,
    config_key VARCHAR(100) NOT NULL,
    config_value JSONB NOT NULL,
    category VARCHAR(50),
    is_active BOOLEAN DEFAULT true,
    priority INTEGER DEFAULT 0,
    version INTEGER DEFAULT 1,
    created_by VARCHAR(36),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_ai_configs_type ON ai_configs(config_type);
CREATE INDEX idx_ai_configs_key ON ai_configs(config_key);
CREATE INDEX idx_ai_configs_category ON ai_configs(category);
CREATE INDEX idx_ai_configs_active ON ai_configs(is_active);
CREATE UNIQUE INDEX idx_ai_configs_type_key ON ai_configs(config_type, config_key) WHERE is_active = true;

-- 4. 创建 ai_content_templates 表
CREATE TABLE IF NOT EXISTS ai_content_templates (
    id VARCHAR(36) PRIMARY KEY,
    template_type VARCHAR(50) NOT NULL,
    category VARCHAR(50),
    title VARCHAR(200) NOT NULL,
    content TEXT NOT NULL,
    tags TEXT[],
    metadata JSONB,
    usage_count INTEGER DEFAULT 0,
    rating DECIMAL(3,2) DEFAULT 0,
    quality_score INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    priority INTEGER DEFAULT 0,
    created_by VARCHAR(36),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_ai_templates_type ON ai_content_templates(template_type);
CREATE INDEX idx_ai_templates_category ON ai_content_templates(category);
CREATE INDEX idx_ai_templates_active ON ai_content_templates(is_active);
CREATE INDEX idx_ai_templates_priority ON ai_content_templates(priority DESC, rating DESC);

-- 5. 创建 ai_config_history 表
CREATE TABLE IF NOT EXISTS ai_config_history (
    id VARCHAR(36) PRIMARY KEY,
    config_id VARCHAR(36) NOT NULL,
    old_value JSONB,
    new_value JSONB,
    change_reason TEXT,
    changed_by VARCHAR(36),
    changed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_ai_config_history_config ON ai_config_history(config_id);
CREATE INDEX idx_ai_config_history_time ON ai_config_history(changed_at DESC);

-- 6. 为 credit_tasks 表添加缺失的索引以优化查询性能
CREATE INDEX IF NOT EXISTS idx_credit_tasks_status ON credit_tasks(status);
CREATE INDEX IF NOT EXISTS idx_credit_tasks_scheduled ON credit_tasks(scheduled_at);
CREATE INDEX IF NOT EXISTS idx_credit_tasks_priority_created ON credit_tasks(priority DESC, created_at ASC);

-- 7. 插入一些默认配置数据
INSERT INTO ai_configs (id, config_type, config_key, config_value, is_active) VALUES
(gen_random_uuid()::text, 'persona', 'poet', '{"name": "诗人", "description": "富有诗意的写作风格", "prompt": "你是一位充满诗意的写作者，善于用优美的语言表达情感。"}', true),
(gen_random_uuid()::text, 'persona', 'friend', '{"name": "朋友", "description": "温暖友好的交流风格", "prompt": "你是一位温暖的朋友，用亲切自然的语言与人交流。"}', true),
(gen_random_uuid()::text, 'persona', 'mentor', '{"name": "导师", "description": "智慧指导的交流风格", "prompt": "你是一位经验丰富的导师，提供深刻的见解和建议。"}', true),
(gen_random_uuid()::text, 'system_prompt', 'default', '{"prompt": "你是一个专业的信件写作助手，帮助用户创作优美的手写信件。", "temperature": 0.7, "max_tokens": 1000}', true),
(gen_random_uuid()::text, 'system_prompt', 'inspiration', '{"prompt": "你是一个创意灵感生成器，为用户提供写信的创意和灵感。", "temperature": 0.8, "max_tokens": 500}', true)
ON CONFLICT DO NOTHING;

-- 输出完成信息
SELECT '数据库表结构修复完成！' as message;