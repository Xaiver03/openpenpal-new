-- 检查用户密码情况
SELECT 
    id,
    username,
    email,
    role,
    CASE 
        WHEN password IS NOT NULL AND password != '' THEN 'SET' 
        ELSE 'NOT SET' 
    END as password_status,
    LENGTH(password) as password_length,
    is_active,
    created_at,
    updated_at
FROM users
WHERE username IN ('alice', 'admin', 'courier_level1', 'courier_level2', 'courier_level3', 'courier_level4')
ORDER BY username;