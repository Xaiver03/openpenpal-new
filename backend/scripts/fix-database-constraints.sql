-- Fix database constraints for GORM migration
-- 这个脚本修复数据库约束名称不匹配的问题

-- 1. 检查现有约束
SELECT conname 
FROM pg_constraint 
WHERE conrelid = 'users'::regclass 
AND contype = 'u';

-- 2. 如果存在旧的约束名，重命名为GORM期望的名称
DO $$
BEGIN
    -- 检查是否存在 unique_username 约束
    IF EXISTS (
        SELECT 1 FROM pg_constraint 
        WHERE conname = 'unique_username' 
        AND conrelid = 'users'::regclass
    ) THEN
        -- 如果GORM寻找 uni_users_username，创建一个别名或重命名
        -- 但实际上我们应该让GORM使用正确的名称
        RAISE NOTICE 'Constraint unique_username already exists';
    END IF;
    
    -- 检查username列是否有唯一索引
    IF NOT EXISTS (
        SELECT 1 
        FROM pg_indexes 
        WHERE tablename = 'users' 
        AND indexname = 'unique_username'
    ) THEN
        CREATE UNIQUE INDEX unique_username ON users(username);
        RAISE NOTICE 'Created unique index on username';
    END IF;
END $$;

-- 3. 修复其他可能的约束问题
-- 检查 credit_shop_categories 表的外键约束
SELECT conname, contype, confrelid::regclass 
FROM pg_constraint 
WHERE conrelid = 'credit_shop_categories'::regclass
AND conname LIKE '%parent%';