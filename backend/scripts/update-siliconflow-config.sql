-- Update SiliconFlow configuration to indicate it needs proper API key
-- Following CLAUDE.md principles

-- Update SiliconFlow with placeholder indicating configuration needed
UPDATE ai_configs
SET 
    api_key = 'PLEASE_CONFIGURE_YOUR_SILICONFLOW_API_KEY',
    api_endpoint = 'https://api.siliconflow.cn/v1/chat/completions',
    model = 'Qwen/Qwen2.5-7B-Instruct',
    is_active = false,  -- Disable until proper key is configured
    daily_quota = 100000,  -- SiliconFlow typically has higher quotas
    updated_at = NOW()
WHERE provider = 'siliconflow';

-- Add other common AI providers if they don't exist
INSERT INTO ai_configs (id, provider, api_key, api_endpoint, model, is_active, daily_quota, used_quota)
VALUES 
    (gen_random_uuid(), 'openai', 'PLEASE_CONFIGURE_YOUR_OPENAI_API_KEY', 'https://api.openai.com/v1/chat/completions', 'gpt-3.5-turbo', false, 50000, 0),
    (gen_random_uuid(), 'claude', 'PLEASE_CONFIGURE_YOUR_CLAUDE_API_KEY', 'https://api.anthropic.com/v1/messages', 'claude-3-sonnet-20240229', false, 50000, 0),
    (gen_random_uuid(), 'local', 'NO_KEY_REQUIRED', 'http://localhost:8001', 'local-model', true, 999999, 0)
ON CONFLICT (provider) DO NOTHING;

-- Show updated configurations
SELECT 
    provider,
    CASE 
        WHEN api_key LIKE 'PLEASE_CONFIGURE%' THEN '⚠️  Configuration Required'
        WHEN api_key = 'NO_KEY_REQUIRED' THEN '✅ No Key Required'
        WHEN LENGTH(api_key) > 10 THEN '✅ Configured'
        ELSE '❌ Invalid Key'
    END as status,
    model,
    is_active,
    daily_quota,
    api_endpoint
FROM ai_configs
ORDER BY 
    CASE 
        WHEN is_active = true THEN 0
        ELSE 1
    END,
    provider;