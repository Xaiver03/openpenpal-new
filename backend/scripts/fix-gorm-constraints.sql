-- 修复 GORM 约束名称问题
-- GORM 期望 uni_users_username 但数据库中是 unique_username

-- 1. 检查当前约束
SELECT conname, contype 
FROM pg_constraint 
WHERE conrelid = 'users'::regclass 
AND contype = 'u';

-- 2. 创建 GORM 期望的约束名称（如果不存在）
DO $$
BEGIN
    -- 如果 unique_username 存在但 uni_users_username 不存在，创建一个指向相同索引的约束
    IF EXISTS (
        SELECT 1 FROM pg_constraint 
        WHERE conname = 'unique_username' 
        AND conrelid = 'users'::regclass
    ) AND NOT EXISTS (
        SELECT 1 FROM pg_constraint 
        WHERE conname = 'uni_users_username' 
        AND conrelid = 'users'::regclass
    ) THEN
        -- 重命名约束
        ALTER TABLE users RENAME CONSTRAINT unique_username TO uni_users_username;
        RAISE NOTICE 'Renamed constraint unique_username to uni_users_username';
    END IF;

    -- 同样处理 email 约束
    IF EXISTS (
        SELECT 1 FROM pg_constraint 
        WHERE conname = 'unique_email' 
        AND conrelid = 'users'::regclass
    ) AND NOT EXISTS (
        SELECT 1 FROM pg_constraint 
        WHERE conname = 'uni_users_email' 
        AND conrelid = 'users'::regclass
    ) THEN
        -- 重命名约束
        ALTER TABLE users RENAME CONSTRAINT unique_email TO uni_users_email;
        RAISE NOTICE 'Renamed constraint unique_email to uni_users_email';
    END IF;
END $$;

-- 3. 再次检查约束
SELECT conname, contype 
FROM pg_constraint 
WHERE conrelid = 'users'::regclass 
AND contype = 'u';