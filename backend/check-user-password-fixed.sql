-- 检查用户密码情况
SELECT 
    id,
    username,
    email,
    role,
    CASE 
        WHEN password_hash IS NOT NULL AND password_hash != '' THEN 'SET' 
        ELSE 'NOT SET' 
    END as password_status,
    LENGTH(password_hash) as password_length,
    SUBSTRING(password_hash, 1, 10) as password_prefix,
    is_active,
    created_at
FROM users
WHERE username IN ('alice', 'admin', 'courier_level1', 'courier_level2', 'courier_level3', 'courier_level4')
   OR email = 'alice@example.com'
ORDER BY username;