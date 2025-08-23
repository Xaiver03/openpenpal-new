-- Check API configurations in the database
-- Following CLAUDE.md principles

-- 1. Check all AI configurations
SELECT 
    id,
    provider,
    api_key,
    api_endpoint,
    model,
    is_active,
    daily_quota,
    used_quota,
    created_at,
    updated_at
FROM ai_configs
ORDER BY provider;

-- 2. Check active providers
SELECT 
    provider,
    model,
    is_active,
    CASE 
        WHEN api_key IS NULL OR api_key = '' THEN 'No API Key'
        WHEN LENGTH(api_key) < 10 THEN 'Invalid Key'
        ELSE 'Key Configured'
    END as key_status,
    daily_quota,
    used_quota,
    ROUND((used_quota::numeric / NULLIF(daily_quota, 0) * 100), 2) as usage_percentage
FROM ai_configs
WHERE is_active = true
ORDER BY provider;

-- 3. Check SiliconFlow specific configuration
SELECT 
    provider,
    api_key,
    api_endpoint,
    model,
    is_active,
    'SiliconFlow Status:' as status,
    CASE
        WHEN api_key IS NULL OR api_key = '' THEN 'Missing API Key - Need configuration'
        WHEN api_key LIKE 'sk-%' THEN 'Has API Key format'
        ELSE 'Has API Key (non-standard format)'
    END as api_key_status
FROM ai_configs
WHERE provider = 'siliconflow';

-- 4. Check if we need to insert SiliconFlow configuration
SELECT 
    'SiliconFlow provider exists:' as check_type,
    CASE 
        WHEN COUNT(*) > 0 THEN 'Yes'
        ELSE 'No - Need to insert'
    END as result
FROM ai_configs
WHERE provider = 'siliconflow';

-- 5. Summary of all providers
SELECT 
    COUNT(DISTINCT provider) as total_providers,
    COUNT(CASE WHEN is_active = true THEN 1 END) as active_providers,
    COUNT(CASE WHEN api_key IS NOT NULL AND api_key != '' THEN 1 END) as configured_providers,
    COUNT(CASE WHEN provider = 'siliconflow' THEN 1 END) as siliconflow_count
FROM ai_configs;