-- 🎨 OpenPenPal Courier Promotion System Migration (Fixed)
-- 修复版本 - 解决类型不匹配问题
-- Created: 2025-08-02
-- Version: 2.0.1

-- ============================================
-- 0. 创建迁移历史表（如果不存在）
-- ============================================

CREATE TABLE IF NOT EXISTS schema_migrations (
    version VARCHAR(50) PRIMARY KEY,
    description TEXT,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ============================================
-- 1. 修复 couriers 表 - 添加层级关系
-- ============================================

-- 添加 parent_id 字段（使用正确的类型）
ALTER TABLE couriers 
ADD COLUMN IF NOT EXISTS parent_id VARCHAR(36),
ADD COLUMN IF NOT EXISTS promoted_at TIMESTAMP,
ADD COLUMN IF NOT EXISTS promotion_count INTEGER DEFAULT 0;

-- 创建索引以优化层级查询
CREATE INDEX IF NOT EXISTS idx_couriers_parent_id ON couriers(parent_id);
CREATE INDEX IF NOT EXISTS idx_couriers_level_status ON couriers(level, status);

-- 添加外键约束（自引用）
DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'fk_couriers_parent'
    ) THEN
        ALTER TABLE couriers 
        ADD CONSTRAINT fk_couriers_parent 
        FOREIGN KEY (parent_id) REFERENCES couriers(id) 
        ON DELETE SET NULL;
    END IF;
END $$;

-- ============================================
-- 2. 创建晋升申请表（如果不存在）
-- ============================================

CREATE TABLE IF NOT EXISTS courier_upgrade_requests (
    id VARCHAR(36) PRIMARY KEY DEFAULT gen_random_uuid()::text,
    courier_id VARCHAR(36) NOT NULL,
    current_level INTEGER NOT NULL,
    request_level INTEGER NOT NULL,
    reason TEXT NOT NULL,
    evidence JSONB DEFAULT '{}',
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected', 'cancelled')),
    reviewer_id VARCHAR(36),
    reviewer_comment TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    reviewed_at TIMESTAMP,
    expires_at TIMESTAMP DEFAULT (CURRENT_TIMESTAMP + INTERVAL '30 days'),
    CONSTRAINT fk_upgrade_courier FOREIGN KEY (courier_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_upgrade_reviewer FOREIGN KEY (reviewer_id) REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT chk_level_upgrade CHECK (request_level > current_level AND request_level <= 4)
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_upgrade_requests_courier ON courier_upgrade_requests(courier_id);
CREATE INDEX IF NOT EXISTS idx_upgrade_requests_status ON courier_upgrade_requests(status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_upgrade_requests_reviewer ON courier_upgrade_requests(reviewer_id) WHERE reviewer_id IS NOT NULL;

-- ============================================
-- 3. 创建晋升历史表（如果不存在）
-- ============================================

CREATE TABLE IF NOT EXISTS courier_promotion_history (
    id VARCHAR(36) PRIMARY KEY DEFAULT gen_random_uuid()::text,
    courier_id VARCHAR(36) NOT NULL,
    from_level INTEGER NOT NULL,
    to_level INTEGER NOT NULL,
    promoted_by VARCHAR(36),
    promotion_type VARCHAR(20) DEFAULT 'regular' CHECK (promotion_type IN ('regular', 'exceptional', 'system')),
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_history_courier FOREIGN KEY (courier_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_history_promoter FOREIGN KEY (promoted_by) REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_promotion_history_courier ON courier_promotion_history(courier_id, created_at DESC);

-- ============================================
-- 4. 创建晋升要求配置表（如果不存在）
-- ============================================

CREATE TABLE IF NOT EXISTS courier_level_requirements (
    id SERIAL PRIMARY KEY,
    from_level INTEGER NOT NULL,
    to_level INTEGER NOT NULL,
    requirement_type VARCHAR(50) NOT NULL,
    requirement_value JSONB NOT NULL,
    is_mandatory BOOLEAN DEFAULT true,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(from_level, to_level, requirement_type)
);

-- 插入默认晋升要求（忽略已存在的）
INSERT INTO courier_level_requirements (from_level, to_level, requirement_type, requirement_value, description) VALUES
(1, 2, 'min_deliveries', '{"value": 50}', '完成至少50次投递'),
(1, 2, 'min_success_rate', '{"value": 95}', '成功率达到95%以上'),
(1, 2, 'min_service_days', '{"value": 30}', '服务时间超过30天'),
(2, 3, 'min_deliveries', '{"value": 200}', '完成至少200次投递'),
(2, 3, 'min_success_rate', '{"value": 97}', '成功率达到97%以上'),
(2, 3, 'min_subordinates', '{"value": 3}', '培养至少3名下级信使'),
(3, 4, 'min_deliveries', '{"value": 500}', '完成至少500次投递'),
(3, 4, 'min_success_rate', '{"value": 98}', '成功率达到98%以上'),
(3, 4, 'min_subordinates', '{"value": 10}', '管理至少10名下级信使')
ON CONFLICT DO NOTHING;

-- ============================================
-- 5. 创建晋升统计视图（修复类型问题）
-- ============================================

CREATE OR REPLACE VIEW v_courier_promotion_stats AS
SELECT 
    c.id,
    u.username,
    u.nickname,
    c.level,
    c.status,
    c.task_count,
    c.points,
    COALESCE(sub.subordinate_count, 0) as subordinate_count,
    COALESCE(req.pending_requests, 0) as pending_upgrade_requests,
    c.created_at as courier_since,
    EXTRACT(DAY FROM (CURRENT_TIMESTAMP - c.created_at))::INTEGER as service_days
FROM couriers c
JOIN users u ON c.user_id = u.id
LEFT JOIN (
    SELECT parent_id, COUNT(*) as subordinate_count
    FROM couriers
    WHERE parent_id IS NOT NULL
    GROUP BY parent_id
) sub ON c.id = sub.parent_id
LEFT JOIN (
    SELECT reviewer_id, COUNT(*) as pending_requests
    FROM courier_upgrade_requests
    WHERE status = 'pending'
    GROUP BY reviewer_id
) req ON c.user_id = req.reviewer_id;

-- ============================================
-- 6. 创建触发器函数（如果不存在）
-- ============================================

CREATE OR REPLACE FUNCTION log_courier_promotion() 
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.level < NEW.level THEN
        INSERT INTO courier_promotion_history (
            courier_id, from_level, to_level, promotion_type
        ) VALUES (
            NEW.user_id, OLD.level, NEW.level, 'system'
        );
        
        NEW.promoted_at = CURRENT_TIMESTAMP;
        NEW.promotion_count = COALESCE(OLD.promotion_count, 0) + 1;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 创建触发器
DROP TRIGGER IF EXISTS trg_courier_promotion ON couriers;
CREATE TRIGGER trg_courier_promotion
BEFORE UPDATE OF level ON couriers
FOR EACH ROW
EXECUTE FUNCTION log_courier_promotion();

-- ============================================
-- 7. 数据初始化 - 为现有用户创建信使记录
-- ============================================

-- 为信使用户创建courier记录（如果不存在）
DO $$
DECLARE
    v_user_record RECORD;
    v_courier_id VARCHAR(36);
BEGIN
    FOR v_user_record IN 
        SELECT u.id, u.username, u.nickname, u.email, u.role
        FROM users u
        WHERE u.role LIKE 'courier%'
        AND NOT EXISTS (
            SELECT 1 FROM couriers c WHERE c.user_id = u.id
        )
    LOOP
        v_courier_id := gen_random_uuid()::text;
        
        INSERT INTO couriers (
            id, user_id, name, contact, school, zone, 
            level, status, created_at, task_count, points
        ) VALUES (
            v_courier_id,
            v_user_record.id,
            COALESCE(v_user_record.nickname, v_user_record.username),
            v_user_record.email,
            'OpenPenPal University',
            CASE 
                WHEN v_user_record.role = 'courier_level4' THEN 'CITY'
                WHEN v_user_record.role = 'courier_level3' THEN 'SCHOOL'
                WHEN v_user_record.role = 'courier_level2' THEN 'DISTRICT'
                ELSE 'BUILDING'
            END,
            CASE 
                WHEN v_user_record.role = 'courier_level1' THEN 1
                WHEN v_user_record.role = 'courier_level2' THEN 2
                WHEN v_user_record.role = 'courier_level3' THEN 3
                WHEN v_user_record.role = 'courier_level4' THEN 4
                ELSE 1
            END,
            'active',
            CURRENT_TIMESTAMP - INTERVAL '60 days',
            FLOOR(RANDOM() * 100 + 10)::INTEGER,  -- 随机10-110个任务
            FLOOR(RANDOM() * 1000 + 100)::INTEGER  -- 随机100-1100积分
        );
        
        RAISE NOTICE 'Created courier record for user: %', v_user_record.username;
    END LOOP;
END $$;

-- 建立示例层级关系
DO $$
DECLARE
    v_level4_id VARCHAR(36);
    v_level3_id VARCHAR(36);
    v_level2_id VARCHAR(36);
BEGIN
    -- 获取一个4级信使作为顶级
    SELECT c.id INTO v_level4_id
    FROM couriers c
    WHERE c.level = 4
    LIMIT 1;
    
    -- 如果有4级信使，建立层级
    IF v_level4_id IS NOT NULL THEN
        -- 将所有3级信使设为4级信使的下级
        UPDATE couriers 
        SET parent_id = v_level4_id
        WHERE level = 3;
        
        -- 获取一个3级信使
        SELECT c.id INTO v_level3_id
        FROM couriers c
        WHERE c.level = 3
        LIMIT 1;
        
        IF v_level3_id IS NOT NULL THEN
            -- 将部分2级信使设为3级信使的下级
            UPDATE couriers 
            SET parent_id = v_level3_id
            WHERE level = 2
            AND id IN (
                SELECT id FROM couriers 
                WHERE level = 2 
                LIMIT 2
            );
        END IF;
        
        -- 获取一个2级信使
        SELECT c.id INTO v_level2_id
        FROM couriers c
        WHERE c.level = 2
        LIMIT 1;
        
        IF v_level2_id IS NOT NULL THEN
            -- 将部分1级信使设为2级信使的下级
            UPDATE couriers 
            SET parent_id = v_level2_id
            WHERE level = 1
            AND id IN (
                SELECT id FROM couriers 
                WHERE level = 1 
                LIMIT 3
            );
        END IF;
    END IF;
END $$;

-- ============================================
-- 8. 创建示例晋升申请数据
-- ============================================

INSERT INTO courier_upgrade_requests (
    courier_id, current_level, request_level, reason, evidence, status
)
SELECT 
    u.id,
    1,
    2,
    '已完成新手期任务，申请晋升为二级信使',
    jsonb_build_object(
        'deliveries', 55,
        'success_rate', 96.5,
        'service_days', 65
    ),
    'pending'
FROM users u
JOIN couriers c ON u.id = c.user_id
WHERE c.level = 1
LIMIT 2
ON CONFLICT DO NOTHING;

-- ============================================
-- 9. 添加注释
-- ============================================

COMMENT ON TABLE courier_upgrade_requests IS '信使晋升申请表 - 记录所有晋升申请及审核状态';
COMMENT ON TABLE courier_promotion_history IS '信使晋升历史表 - 追踪所有晋升记录';
COMMENT ON TABLE courier_level_requirements IS '晋升要求配置表 - 定义各级别晋升所需条件';
COMMENT ON VIEW v_courier_promotion_stats IS '信使晋升统计视图 - 提供综合统计数据';

-- ============================================
-- 迁移完成标记
-- ============================================

INSERT INTO schema_migrations (version, description, applied_at)
VALUES ('002_fixed', 'Courier Promotion System (Fixed)', CURRENT_TIMESTAMP)
ON CONFLICT DO NOTHING;

-- 输出完成信息
DO $$
BEGIN
    RAISE NOTICE '🎉 晋升系统数据库迁移完成！';
    RAISE NOTICE '✅ 创建了晋升申请表 courier_upgrade_requests';
    RAISE NOTICE '✅ 创建了晋升历史表 courier_promotion_history';
    RAISE NOTICE '✅ 创建了晋升要求表 courier_level_requirements';
    RAISE NOTICE '✅ 创建了统计视图 v_courier_promotion_stats';
    RAISE NOTICE '✅ 为现有信使用户创建了courier记录';
    RAISE NOTICE '✅ 建立了示例层级关系';
    RAISE NOTICE '✅ 创建了示例晋升申请';
END $$;