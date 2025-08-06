-- Check AI configuration
SELECT 
    id,
    provider,
    api_endpoint,
    model,
    CASE 
        WHEN api_key IS NOT NULL AND api_key != '' THEN 'SET' 
        ELSE 'NOT SET' 
    END as api_key_status,
    is_active,
    priority,
    daily_quota,
    used_quota,
    quota_reset_at,
    created_at,
    updated_at
FROM ai_configs
ORDER BY priority DESC, created_at DESC;