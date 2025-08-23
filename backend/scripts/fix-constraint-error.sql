-- Fix constraint error
-- 检查并创建缺失的约束

-- 首先检查约束是否已存在
DO $$
BEGIN
    -- 如果约束不存在，尝试创建它
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint 
        WHERE conname = 'uni_users_username' 
        AND conrelid = 'users'::regclass
    ) THEN
        -- 如果需要创建约束，这里可以添加
        -- ALTER TABLE users ADD CONSTRAINT uni_users_username UNIQUE(username);
        RAISE NOTICE 'Constraint uni_users_username does not exist, which is expected';
    END IF;
END $$;

-- 确保username列有唯一索引
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username ON users(username);