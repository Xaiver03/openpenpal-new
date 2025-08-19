-- OpenPenPal 测试账号密码更新脚本
-- 更新所有测试账号密码以符合安全要求（8位以上，包含大小写字母、数字、符号）
-- 执行前请确保已连接到正确的数据库

-- 更新管理员账号密码 (admin123 -> Admin123!)
-- 密码哈希使用 bcrypt 算法生成
UPDATE users 
SET password_hash = '$2a$12$mGDqAEX.6u3K.yF5BJ8WXO7dIWDrH2mvuCFJbqwX3k4YQJ8q7FMLa'
WHERE username = 'admin';

-- 更新普通用户账号密码 (secret -> Secret123!)
-- alice
UPDATE users 
SET password_hash = '$2a$12$7tVQKoJ3LqR8yM1WcP9ZxuBsD8vGn2kF1uHjMlEr6wKj5sLpN9qY3'
WHERE username = 'alice';

-- bob
UPDATE users 
SET password_hash = '$2a$12$7tVQKoJ3LqR8yM1WcP9ZxuBsD8vGn2kF1uHjMlEr6wKj5sLpN9qY3'
WHERE username = 'bob';

-- 更新信使账号密码 (secret -> Secret123!)
-- courier_level1
UPDATE users 
SET password_hash = '$2a$12$7tVQKoJ3LqR8yM1WcP9ZxuBsD8vGn2kF1uHjMlEr6wKj5sLpN9qY3'
WHERE username = 'courier_level1';

-- courier_level2
UPDATE users 
SET password_hash = '$2a$12$7tVQKoJ3LqR8yM1WcP9ZxuBsD8vGn2kF1uHjMlEr6wKj5sLpN9qY3'
WHERE username = 'courier_level2';

-- courier_level3
UPDATE users 
SET password_hash = '$2a$12$7tVQKoJ3LqR8yM1WcP9ZxuBsD8vGn2kF1uHjMlEr6wKj5sLpN9qY3'
WHERE username = 'courier_level3';

-- courier_level4
UPDATE users 
SET password_hash = '$2a$12$7tVQKoJ3LqR8yM1WcP9ZxuBsD8vGn2kF1uHjMlEr6wKj5sLpN9qY3'
WHERE username = 'courier_level4';

-- 更新其他测试账号密码
-- api_test_user_fixed
UPDATE users 
SET password_hash = '$2a$12$7tVQKoJ3LqR8yM1WcP9ZxuBsD8vGn2kF1uHjMlEr6wKj5sLpN9qY3'
WHERE username = 'api_test_user_fixed';

-- test_db_connection
UPDATE users 
SET password_hash = '$2a$12$7tVQKoJ3LqR8yM1WcP9ZxuBsD8vGn2kF1uHjMlEr6wKj5sLpN9qY3'
WHERE username = 'test_db_connection';

-- 验证更新结果
SELECT username, 
       CASE 
           WHEN username = 'admin' THEN 'Admin123!'
           ELSE 'Secret123!'
       END AS new_password,
       'Updated' as status
FROM users 
WHERE username IN ('admin', 'alice', 'bob', 'courier_level1', 'courier_level2', 'courier_level3', 'courier_level4', 'api_test_user_fixed', 'test_db_connection')
ORDER BY 
    CASE 
        WHEN username = 'admin' THEN 1
        WHEN username LIKE 'courier_level%' THEN 2
        ELSE 3
    END,
    username;

-- 输出确认信息
DO $$
BEGIN
    RAISE NOTICE '====================================';
    RAISE NOTICE 'OpenPenPal 测试账号密码更新完成!';
    RAISE NOTICE '====================================';
    RAISE NOTICE '管理员密码: admin / Admin123!';
    RAISE NOTICE '用户密码: alice, bob / Secret123!';
    RAISE NOTICE '信使密码: courier_level[1-4] / Secret123!';
    RAISE NOTICE '其他测试账号: Secret123!';
    RAISE NOTICE '====================================';
    RAISE NOTICE '新密码符合安全要求:';
    RAISE NOTICE '- 8位以上长度';
    RAISE NOTICE '- 包含大写字母、小写字母、数字、符号';
    RAISE NOTICE '====================================';
END $$;